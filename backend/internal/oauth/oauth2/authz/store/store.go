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

// Package store provides functionality for handling authorization code persistence and retrieval.
package store

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/asgardeo/thunder/internal/oauth/oauth2/authz/constants"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/authz/model"
	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/log"
)

const loggerComponentName = "AuthorizationCodeStore"

// AuthorizationCodeStoreInterface defines the interface for managing authorization codes.
type AuthorizationCodeStoreInterface interface {
	InsertAuthorizationCode(authzCode model.AuthorizationCode) error
	GetAuthorizationCode(clientID, authCode string) (model.AuthorizationCode, error)
	DeactivateAuthorizationCode(authzCode model.AuthorizationCode) error
	RevokeAuthorizationCode(authzCode model.AuthorizationCode) error
	ExpireAuthorizationCode(authzCode model.AuthorizationCode) error
}

// AuthorizationCodeStore implements the AuthorizationCodeStoreInterface for managing authorization codes.
type AuthorizationCodeStore struct {
	DBProvider provider.DBProviderInterface
}

// NewAuthorizationCodeStore creates a new instance of AuthorizationCodeStore.
func NewAuthorizationCodeStore() AuthorizationCodeStoreInterface {
	return &AuthorizationCodeStore{
		DBProvider: provider.GetDBProvider(),
	}
}

// InsertAuthorizationCode inserts a new authorization code into the database.
func (acs *AuthorizationCodeStore) InsertAuthorizationCode(authzCode model.AuthorizationCode) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	dbClient, err := acs.DBProvider.GetDBClient("runtime")
	if err != nil {
		logger.Error("Failed to get database client", log.Error(err))
		return err
	}

	tx, err := dbClient.BeginTx()
	if err != nil {
		logger.Error("Failed to begin transaction", log.Error(err))
		return errors.New("failed to begin transaction: " + err.Error())
	}

	// Insert authorization code.
	_, err = tx.Exec(constants.QueryInsertAuthorizationCode.Query, authzCode.CodeID, authzCode.Code,
		authzCode.ClientID, authzCode.RedirectURI, authzCode.AuthorizedUserID, authzCode.TimeCreated,
		authzCode.ExpiryTime, authzCode.State, authzCode.CodeChallenge, authzCode.CodeChallengeMethod)
	if err != nil {
		logger.Error("Failed to insert authorization code", log.Error(err))
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			logger.Error("Failed to rollback transaction", log.Error(rollbackErr))
			err = errors.Join(err, errors.New("failed to rollback transaction: "+rollbackErr.Error()))
		}
		return errors.New("failed to insert authorization code: " + err.Error())
	}

	// Insert auth code scopes.
	_, err = tx.Exec(constants.QueryInsertAuthorizationCodeScopes.Query, authzCode.CodeID,
		authzCode.Scopes)
	if err != nil {
		logger.Error("Failed to insert authorization code scopes", log.Error(err))
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			logger.Error("Failed to rollback transaction", log.Error(rollbackErr))
			err = errors.Join(err, errors.New("failed to rollback transaction: "+rollbackErr.Error()))
		}
		return errors.New("failed to insert authorization code scopes: " + err.Error())
	}

	// Commit the transaction.
	if err = tx.Commit(); err != nil {
		logger.Error("Failed to commit transaction", log.Error(err))
		return errors.New("failed to commit transaction: " + err.Error())
	}

	return nil
}

// GetAuthorizationCode retrieves an authorization code by client Id and authorization code.
func (acs *AuthorizationCodeStore) GetAuthorizationCode(clientID, authCode string) (model.AuthorizationCode, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	dbClient, err := acs.DBProvider.GetDBClient("runtime")
	if err != nil {
		logger.Error("Failed to get database client", log.Error(err))
		return model.AuthorizationCode{}, err
	}

	results, err := dbClient.Query(constants.QueryGetAuthorizationCode, clientID, authCode)
	if err != nil {
		return model.AuthorizationCode{}, fmt.Errorf("error while retrieving authorization code: %w", err)
	}
	if len(results) == 0 {
		return model.AuthorizationCode{}, constants.ErrAuthorizationCodeNotFound
	}
	row := results[0]

	codeID := row["code_id"].(string)
	if codeID == "" {
		return model.AuthorizationCode{}, constants.ErrAuthorizationCodeNotFound
	}

	// Handle time_created field.
	timeCreated, err := parseTimeField(row["time_created"], "time_created", logger)
	if err != nil {
		return model.AuthorizationCode{}, err
	}

	// Handle expiry_time field.
	expiryTime, err := parseTimeField(row["expiry_time"], "expiry_time", logger)
	if err != nil {
		return model.AuthorizationCode{}, err
	}

	// Extract PKCE fields
	codeChallenge := ""
	if val, ok := row["code_challenge"]; ok && val != nil {
		codeChallenge = val.(string)
	}
	codeChallengeMethod := ""
	if val, ok := row["code_challenge_method"]; ok && val != nil {
		codeChallengeMethod = val.(string)
	}

	// Retrieve authorized scopes for the authorization code.
	scopeResults, err := dbClient.Query(constants.QueryGetAuthorizationCodeScopes, codeID)
	if err != nil {
		return model.AuthorizationCode{}, fmt.Errorf("error while retrieving authorized scopes: %w", err)
	}
	scopes := ""
	if len(scopeResults) > 0 {
		scopes = scopeResults[0]["scope"].(string)
	}

	return model.AuthorizationCode{
		CodeID:              codeID,
		Code:                row["authorization_code"].(string),
		ClientID:            clientID,
		RedirectURI:         row["callback_url"].(string),
		AuthorizedUserID:    row["authz_user"].(string),
		TimeCreated:         timeCreated,
		ExpiryTime:          expiryTime,
		Scopes:              scopes,
		State:               row["state"].(string),
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
	}, nil
}

// DeactivateAuthorizationCode deactivates an authorization code.
func (acs *AuthorizationCodeStore) DeactivateAuthorizationCode(authzCode model.AuthorizationCode) error {
	return acs.updateAuthorizationCodeState(authzCode, constants.AuthCodeStateInactive)
}

// RevokeAuthorizationCode revokes an authorization code.
func (acs *AuthorizationCodeStore) RevokeAuthorizationCode(authzCode model.AuthorizationCode) error {
	return acs.updateAuthorizationCodeState(authzCode, constants.AuthCodeStateRevoked)
}

// ExpireAuthorizationCode expires an authorization code.
func (acs *AuthorizationCodeStore) ExpireAuthorizationCode(authzCode model.AuthorizationCode) error {
	return acs.updateAuthorizationCodeState(authzCode, constants.AuthCodeStateExpired)
}

// updateAuthorizationCodeState updates the state of an authorization code.
func (acs *AuthorizationCodeStore) updateAuthorizationCodeState(authzCode model.AuthorizationCode,
	newState string) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	dbClient, err := acs.DBProvider.GetDBClient("runtime")
	if err != nil {
		logger.Error("Failed to get database client", log.Error(err))
		return err
	}

	_, err = dbClient.Execute(constants.QueryUpdateAuthorizationCodeState, newState, authzCode.CodeID)
	return err
}

// Helper function to parse a time field from the database.
func parseTimeField(field interface{}, fieldName string, logger *log.Logger) (time.Time, error) {
	const customTimeFormat = "2006-01-02 15:04:05.999999999"

	switch v := field.(type) {
	case string:
		trimmedTime := trimTimeString(v)
		parsedTime, err := time.Parse(customTimeFormat, trimmedTime)
		if err != nil {
			logger.Error("Error parsing time field", log.String("field", fieldName), log.Error(err))
			return time.Time{}, fmt.Errorf("error parsing %s: %w", fieldName, err)
		}
		return parsedTime, nil
	case time.Time:
		return v, nil
	default:
		logger.Error("Unexpected type for time field", log.String("field", fieldName), log.Any("value", v))
		return time.Time{}, fmt.Errorf("unexpected type for %s", fieldName)
	}
}

// Helper function to trim a time string.
func trimTimeString(timeStr string) string {
	// Split the string into parts by spaces and retain only the first two parts.
	parts := strings.SplitN(timeStr, " ", 3)
	if len(parts) >= 2 {
		return parts[0] + " " + parts[1]
	}
	return timeStr
}
