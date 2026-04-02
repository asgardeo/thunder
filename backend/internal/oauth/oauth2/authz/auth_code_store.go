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
	"context"
	"encoding/json"
	"errors"
	"fmt"

	oauth2model "github.com/asgardeo/thunder/internal/oauth/oauth2/model"
	oauth2utils "github.com/asgardeo/thunder/internal/oauth/oauth2/utils"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/transaction"
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
	jsonDataKeyAttributeCacheID    = "attribute_cache_id"
	jsonDataKeyClaimsRequest       = "claims_request"
	jsonDataKeyClaimsLocales       = "claims_locales"
	jsonDataKeyNonce               = "nonce"
)

// AuthorizationCodeStoreInterface defines the interface for managing authorization codes.
type AuthorizationCodeStoreInterface interface {
	InsertAuthorizationCode(ctx context.Context, authzCode AuthorizationCode) error
	ConsumeAuthorizationCode(ctx context.Context, clientID, authCode string) (*AuthorizationCode, error)
	GetAuthorizationCode(ctx context.Context, clientID, authCode string) (*AuthorizationCode, error)
}

// authorizationCodeStore implements the AuthorizationCodeStoreInterface for managing authorization codes.
type authorizationCodeStore struct {
	dbProvider    provider.DBProviderInterface
	deploymentID  string
	transactioner transaction.Transactioner
}

// newAuthorizationCodeStore creates a new instance of authorizationCodeStore with injected dependencies.
func newAuthorizationCodeStore() (AuthorizationCodeStoreInterface, error) {
	dbProvider := provider.GetDBProvider()
	transactioner, err := dbProvider.GetRuntimeDBTransactioner()
	if err != nil {
		return nil, fmt.Errorf("failed to get runtime DB transactioner for authorization code store: %w", err)
	}
	return &authorizationCodeStore{
		dbProvider:    dbProvider,
		deploymentID:  config.GetThunderRuntime().Config.Server.Identifier,
		transactioner: transactioner,
	}, nil
}

// InsertAuthorizationCode inserts a new authorization code into the database.
func (acs *authorizationCodeStore) InsertAuthorizationCode(
	ctx context.Context, authzCode AuthorizationCode) error {
	dbClient, err := acs.dbProvider.GetRuntimeDBClient()
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	jsonDataBytes, err := acs.getJSONDataBytes(authzCode)
	if err != nil {
		return err
	}

	_, err = dbClient.ExecuteContext(ctx, queryInsertAuthorizationCode, authzCode.CodeID, authzCode.Code,
		authzCode.ClientID, authzCode.State, jsonDataBytes, authzCode.TimeCreated, authzCode.ExpiryTime,
		acs.deploymentID)
	if err != nil {
		return fmt.Errorf("error inserting authorization code: %w", err)
	}

	return nil
}

// ConsumeAuthorizationCode transitions an ACTIVE authorization code to INACTIVE and returns the record.
// Returns errAuthorizationCodeNotFound if the code does not exist or is not in ACTIVE state.
func (acs *authorizationCodeStore) ConsumeAuthorizationCode(
	ctx context.Context, clientID, authCode string,
) (*AuthorizationCode, error) {
	dbClient, err := acs.dbProvider.GetRuntimeDBClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	var consumed *AuthorizationCode
	err = acs.transactioner.Transact(ctx, func(ctx context.Context) error {
		rowsAffected, err := dbClient.ExecuteContext(ctx, queryConsumeAuthorizationCode,
			AuthCodeStateInactive, clientID, authCode, AuthCodeStateActive, acs.deploymentID)
		if err != nil {
			return fmt.Errorf("error consuming authorization code: %w", err)
		}
		if rowsAffected == 0 {
			var lookupResults []map[string]interface{}
			lookupResults, err = dbClient.QueryContext(ctx, queryGetAuthorizationCode,
				clientID, authCode, acs.deploymentID)
			if err != nil {
				return fmt.Errorf("error checking authorization code existence: %w", err)
			}
			if len(lookupResults) > 0 {
				return fmt.Errorf("authorization code already consumed")
			}
			return errAuthorizationCodeNotFound
		}

		results, err := dbClient.QueryContext(ctx, queryGetAuthorizationCode, clientID, authCode, acs.deploymentID)
		if err != nil {
			return fmt.Errorf("error retrieving authorization code after consume: %w", err)
		}
		if len(results) == 0 {
			return fmt.Errorf("authorization code was updated but could not be reloaded")
		}

		consumed, err = buildAuthorizationCodeFromResultRow(results[0])
		return err
	})
	if err != nil {
		return nil, err
	}
	return consumed, nil
}

// GetAuthorizationCode retrieves an authorization code by client Id and authorization code.
func (acs *authorizationCodeStore) GetAuthorizationCode(
	ctx context.Context, clientID, authCode string,
) (*AuthorizationCode, error) {
	dbClient, err := acs.dbProvider.GetRuntimeDBClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.QueryContext(ctx, queryGetAuthorizationCode, clientID, authCode, acs.deploymentID)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving authorization code: %w", err)
	}
	if len(results) == 0 {
		return nil, errAuthorizationCodeNotFound
	}
	row := results[0]

	return buildAuthorizationCodeFromResultRow(row)
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
		jsonDataKeyClaimsLocales:       authzCode.ClaimsLocales,
		jsonDataKeyNonce:               authzCode.Nonce,
	}

	// Include user attributes if present
	if len(authzCode.AttributeCacheID) > 0 {
		jsonData[jsonDataKeyAttributeCacheID] = authzCode.AttributeCacheID
	}

	// Include claims request if present
	if authzCode.ClaimsRequest != nil {
		jsonData[jsonDataKeyClaimsRequest] = authzCode.ClaimsRequest
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
		return nil, errAuthorizationCodeNotFound
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
	if claimsLocales, ok := authzData[jsonDataKeyClaimsLocales].(string); ok {
		authzCode.ClaimsLocales = claimsLocales
	}
	if nonce, ok := authzData[jsonDataKeyNonce].(string); ok {
		authzCode.Nonce = nonce
	}
	if attributeCacheID, ok := authzData[jsonDataKeyAttributeCacheID].(string); ok {
		authzCode.AttributeCacheID = attributeCacheID
	}

	if claimsData, ok := authzData[jsonDataKeyClaimsRequest]; ok && claimsData != nil {
		claimsRequest, err := parseClaimsRequestFromJSON(claimsData)
		if err != nil {
			return nil, fmt.Errorf("failed to parse claims_request from authorization code: %w", err)
		}
		authzCode.ClaimsRequest = claimsRequest
	}

	return authzCode, nil
}

// parseClaimsRequestFromJSON parses a ClaimsRequest from JSON data stored in the database.
func parseClaimsRequestFromJSON(data interface{}) (*oauth2model.ClaimsRequest, error) {
	if data == nil {
		return nil, nil
	}

	// Marshal the data to JSON string
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal claims_request: %w", err)
	}

	return oauth2utils.ParseClaimsRequest(string(jsonBytes))
}
