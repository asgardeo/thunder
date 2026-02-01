/*
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package authz

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/database/provider"
)

const (
	columnNameCodeID               = "code_id"
	columnNameAuthorizationCode    = "authorization_code"
	columnNameClientID             = "client_id"
	columnNameState                = "state"
	columnNameAuthZData            = "authz_data"
	columnNameTimeCreated          = "time_created"
	columnNameExpiryTime           = "expiry_time"
	jsonDataKeyRedirectURI         = "redirect_uri"
	jsonDataKeyAuthorizedUserID    = "authorized_user_id"
	jsonDataKeyScopes              = "scopes"
	jsonDataKeyCodeChallenge       = "code_challenge"
	jsonDataKeyCodeChallengeMethod = "code_challenge_method"
	jsonDataKeyResource            = "resource"
	jsonDataKeyUserAttributes      = "user_attributes"
)

// AuthorizationCodeStoreInterface defines the interface for managing authorization codes.
type AuthorizationCodeStoreInterface interface {
	InsertAuthorizationCode(authzCode AuthorizationCode) error
	GetAuthorizationCode(clientID, authCode string) (*AuthorizationCode, error)
	DeactivateAuthorizationCode(authzCode AuthorizationCode) error
	RevokeAuthorizationCode(authzCode AuthorizationCode) error
	ExpireAuthorizationCode(authzCode AuthorizationCode) error
}

// authorizationCodeStore implements the AuthorizationCodeStoreInterface for managing authorization codes.
type authorizationCodeStore struct {
	dbProvider   provider.DBProviderInterface
	deploymentID string
}

// newAuthorizationCodeStore creates a new instance of authorizationCodeStore with injected dependencies.
func newAuthorizationCodeStore() AuthorizationCodeStoreInterface {
	return &authorizationCodeStore{
		dbProvider:   provider.GetDBProvider(),
		deploymentID: config.GetThunderRuntime().Config.Server.Identifier,
	}
}

// InsertAuthorizationCode inserts a new authorization code into the database.
func (acs *authorizationCodeStore) InsertAuthorizationCode(authzCode AuthorizationCode) error {
	dbClient, err := acs.dbProvider.GetRuntimeDBClient()
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	jsonDataBytes, err := acs.getJSONDataBytes(authzCode)
	if err != nil {
		return err
	}

	_, err = dbClient.Execute(queryInsertAuthorizationCode, authzCode.CodeID, authzCode.Code,
		authzCode.ClientID, authzCode.State, jsonDataBytes, authzCode.TimeCreated, authzCode.ExpiryTime,
		acs.deploymentID)
	if err != nil {
		return fmt.Errorf("error inserting authorization code: %w", err)
	}

	return nil
}

// GetAuthorizationCode retrieves an authorization code by client Id and authorization code.
func (acs *authorizationCodeStore) GetAuthorizationCode(clientID, authCode string) (*AuthorizationCode, error) {
	dbClient, err := acs.dbProvider.GetRuntimeDBClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(queryGetAuthorizationCode, clientID, authCode, acs.deploymentID)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving authorization code: %w", err)
	}
	if len(results) == 0 {
		return nil, ErrAuthorizationCodeNotFound
	}
	row := results[0]

	return buildAuthorizationCodeFromResultRow(row)
}

// DeactivateAuthorizationCode deactivates an authorization code.
func (acs *authorizationCodeStore) DeactivateAuthorizationCode(authzCode AuthorizationCode) error {
	return acs.updateAuthorizationCodeState(authzCode, AuthCodeStateInactive)
}

// RevokeAuthorizationCode revokes an authorization code.
func (acs *authorizationCodeStore) RevokeAuthorizationCode(authzCode AuthorizationCode) error {
	return acs.updateAuthorizationCodeState(authzCode, AuthCodeStateRevoked)
}

// ExpireAuthorizationCode expires an authorization code.
func (acs *authorizationCodeStore) ExpireAuthorizationCode(authzCode AuthorizationCode) error {
	return acs.updateAuthorizationCodeState(authzCode, AuthCodeStateExpired)
}

// updateAuthorizationCodeState updates the state of an authorization code.
func (acs *authorizationCodeStore) updateAuthorizationCodeState(authzCode AuthorizationCode,
	newState string) error {
	dbClient, err := acs.dbProvider.GetRuntimeDBClient()
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	_, err = dbClient.Execute(queryUpdateAuthorizationCodeState, newState, authzCode.CodeID, acs.deploymentID)
	return err
}

// getJSONDataBytes prepares the JSON data bytes for the authorization code.
func (acs *authorizationCodeStore) getJSONDataBytes(authzCode AuthorizationCode) ([]byte, error) {
	jsonData := map[string]interface{}{
		jsonDataKeyRedirectURI:         authzCode.RedirectURI,
		jsonDataKeyAuthorizedUserID:    authzCode.AuthorizedUserID,
		jsonDataKeyScopes:              authzCode.Scopes,
		jsonDataKeyCodeChallenge:       authzCode.CodeChallenge,
		jsonDataKeyCodeChallengeMethod: authzCode.CodeChallengeMethod,
		jsonDataKeyResource:            authzCode.Resource,
	}

	// Include user attributes if present
	if len(authzCode.UserAttributes) > 0 {
		jsonData[jsonDataKeyUserAttributes] = authzCode.UserAttributes
	}

	jsonDataBytes, err := json.Marshal(jsonData)
	if err != nil {
		return nil, fmt.Errorf("error marshaling authz data to JSON: %w", err)
	}
	return jsonDataBytes, nil
}

// buildAuthorizationCodeFromResultRow builds an AuthorizationCode from a database result row.
func buildAuthorizationCodeFromResultRow(row map[string]interface{}) (*AuthorizationCode, error) {
	codeID, ok := row[columnNameCodeID].(string)
	if !ok {
		return nil, errors.New("code ID is of unexpected type")
	}
	if codeID == "" {
		return nil, ErrAuthorizationCodeNotFound
	}

	authorizationCode, ok := row[columnNameAuthorizationCode].(string)
	if !ok {
		return nil, errors.New("authorization code is of unexpected type")
	}
	if authorizationCode == "" {
		return nil, errors.New("authorization code is empty")
	}

	clientID, ok := row[columnNameClientID].(string)
	if !ok {
		return nil, errors.New("client ID is of unexpected type")
	}
	if clientID == "" {
		return nil, errors.New("client ID is empty")
	}

	state, ok := row[columnNameState].(string)
	if !ok {
		return nil, errors.New("state is of unexpected type")
	}
	if state == "" {
		return nil, errors.New("state is empty")
	}

	timeCreated, err := parseTimeField(row[columnNameTimeCreated], columnNameTimeCreated)
	if err != nil {
		return nil, err
	}
	expiryTime, err := parseTimeField(row[columnNameExpiryTime], columnNameExpiryTime)
	if err != nil {
		return nil, err
	}

	authzCode := AuthorizationCode{
		CodeID:      codeID,
		Code:        authorizationCode,
		ClientID:    clientID,
		State:       state,
		TimeCreated: timeCreated,
		ExpiryTime:  expiryTime,
	}

	return appendAuthzDataJSON(row, &authzCode)
}

// appendAuthzDataJSON parses and appends authz_data JSON fields to the AuthorizationCode struct.
func appendAuthzDataJSON(row map[string]interface{}, authzCode *AuthorizationCode) (*AuthorizationCode, error) {
	var dataJSON string
	if val, ok := row[columnNameAuthZData].(string); ok && val != "" {
		dataJSON = val
	} else if val, ok := row[columnNameAuthZData].([]byte); ok && len(val) > 0 {
		dataJSON = string(val)
	} else {
		return nil, errors.New("authz_data is missing or of unexpected type")
	}
	if dataJSON == "" || dataJSON == "{}" {
		return nil, errors.New("authz_data is empty")
	}

	var authzData map[string]interface{}
	if err := json.Unmarshal([]byte(dataJSON), &authzData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal authz_data JSON: %w", err)
	}

	if redirectURI, ok := authzData[jsonDataKeyRedirectURI].(string); ok {
		authzCode.RedirectURI = redirectURI
	}
	if authorizedUserID, ok := authzData[jsonDataKeyAuthorizedUserID].(string); ok {
		authzCode.AuthorizedUserID = authorizedUserID
	}
	if scopes, ok := authzData[jsonDataKeyScopes].(string); ok {
		authzCode.Scopes = scopes
	}
	if codeChallenge, ok := authzData[jsonDataKeyCodeChallenge].(string); ok {
		authzCode.CodeChallenge = codeChallenge
	}
	if codeChallengeMethod, ok := authzData[jsonDataKeyCodeChallengeMethod].(string); ok {
		authzCode.CodeChallengeMethod = codeChallengeMethod
	}
	if resource, ok := authzData[jsonDataKeyResource].(string); ok {
		authzCode.Resource = resource
	}

	if userAttrs, ok := authzData[jsonDataKeyUserAttributes].(map[string]interface{}); ok && userAttrs != nil {
		authzCode.UserAttributes = userAttrs
	}

	return authzCode, nil
}
