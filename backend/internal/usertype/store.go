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

package usertype

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/log"
)

// userTypeStoreInterface defines the interface for user type store operations.
type userTypeStoreInterface interface {
	GetUserTypeListCount() (int, error)
	GetUserTypeList(limit, offset int) ([]UserTypeListItem, error)
	CreateUserType(userType UserType) error
	GetUserTypeByID(schemaID string) (UserType, error)
	GetUserTypeByName(name string) (UserType, error)
	UpdateUserTypeByID(schemaID string, userType UserType) error
	DeleteUserTypeByID(schemaID string) error
}

// userTypeStore is the default implementation of userTypeStoreInterface.
type userTypeStore struct {
	dbProvider   provider.DBProviderInterface
	deploymentID string
}

// newUserTypeStore creates a new instance of userTypeStore.
func newUserTypeStore() userTypeStoreInterface {
	return &userTypeStore{
		dbProvider:   provider.GetDBProvider(),
		deploymentID: config.GetThunderRuntime().Config.Server.Identifier,
	}
}

// GetUserTypeListCount retrieves the total count of user types.
func (s *userTypeStore) GetUserTypeListCount() (int, error) {
	dbClient, err := s.dbProvider.GetConfigDBClient()
	if err != nil {
		return 0, fmt.Errorf("failed to get database client: %w", err)
	}

	countResults, err := dbClient.Query(queryGetUserTypeCount, s.deploymentID)
	if err != nil {
		return 0, fmt.Errorf("failed to execute count query: %w", err)
	}

	var totalCount int
	if len(countResults) > 0 {
		if count, ok := countResults[0]["total"].(int64); ok {
			totalCount = int(count)
		} else {
			return 0, fmt.Errorf("failed to parse count result")
		}
	}

	return totalCount, nil
}

// GetUserTypeList retrieves a list of user types with pagination.
func (s *userTypeStore) GetUserTypeList(limit, offset int) ([]UserTypeListItem, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserTypePersistence"))

	dbClient, err := s.dbProvider.GetConfigDBClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(queryGetUserTypeList, limit, offset, s.deploymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	userTypes := make([]UserTypeListItem, 0, len(results))
	for _, row := range results {
		userType, err := parseUserTypeListItemFromRow(row)
		if err != nil {
			logger.Error("Failed to parse user type list item from row", log.Error(err))
			continue
		}
		userTypes = append(userTypes, userType)
	}

	return userTypes, nil
}

// CreateUserType creates a new user type.
func (s *userTypeStore) CreateUserType(userType UserType) error {
	dbClient, err := s.dbProvider.GetConfigDBClient()
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	_, err = dbClient.Query(
		queryCreateUserType,
		userType.ID,
		userType.Name,
		userType.OrganizationUnitID,
		userType.AllowSelfRegistration,
		string(userType.Schema),
		s.deploymentID,
	)
	if err != nil {
		return fmt.Errorf("failed to create user type: %w", err)
	}

	return nil
}

// GetUserTypeByID retrieves a user type by its ID.
func (s *userTypeStore) GetUserTypeByID(schemaID string) (UserType, error) {
	dbClient, err := s.dbProvider.GetConfigDBClient()
	if err != nil {
		return UserType{}, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(queryGetUserTypeByID, schemaID, s.deploymentID)
	if err != nil {
		return UserType{}, fmt.Errorf("failed to execute query: %w", err)
	}

	if len(results) == 0 {
		return UserType{}, ErrUserTypeNotFound
	}

	return parseUserTypeFromRow(results[0])
}

// GetUserTypeByName retrieves a user type by its name.
func (s *userTypeStore) GetUserTypeByName(name string) (UserType, error) {
	dbClient, err := s.dbProvider.GetConfigDBClient()
	if err != nil {
		return UserType{}, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(queryGetUserTypeByName, name, s.deploymentID)
	if err != nil {
		return UserType{}, fmt.Errorf("failed to execute query: %w", err)
	}

	if len(results) == 0 {
		return UserType{}, ErrUserTypeNotFound
	}

	return parseUserTypeFromRow(results[0])
}

// UpdateUserTypeByID updates a user type by its ID.
func (s *userTypeStore) UpdateUserTypeByID(schemaID string, userType UserType) error {
	dbClient, err := s.dbProvider.GetConfigDBClient()
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	_, err = dbClient.Query(
		queryUpdateUserTypeByID,
		userType.Name,
		userType.OrganizationUnitID,
		userType.AllowSelfRegistration,
		string(userType.Schema),
		schemaID,
		s.deploymentID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user type: %w", err)
	}

	return nil
}

// DeleteUserTypeByID deletes a user type by its ID.
func (s *userTypeStore) DeleteUserTypeByID(schemaID string) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserTypePersistence"))

	dbClient, err := s.dbProvider.GetConfigDBClient()
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	rowsAffected, err := dbClient.Execute(queryDeleteUserTypeByID, schemaID, s.deploymentID)
	if err != nil {
		return fmt.Errorf("failed to delete user type: %w", err)
	}

	if rowsAffected == 0 {
		logger.Debug("user not found with id: " + schemaID)
	}

	return nil
}

// parseUserTypeFromRow parses a user type from a database row.
func parseUserTypeFromRow(row map[string]interface{}) (UserType, error) {
	schemaID, ok := row["schema_id"].(string)
	if !ok {
		return UserType{}, fmt.Errorf("failed to parse schema_id as string")
	}

	name, ok := row["name"].(string)
	if !ok {
		return UserType{}, fmt.Errorf("failed to parse name as string")
	}

	organizationUnitID, ok := row["ou_id"].(string)
	if !ok {
		return UserType{}, fmt.Errorf("failed to parse ou_id as string")
	}

	allowSelfRegistration, err := parseBool(row["allow_self_registration"], "allow_self_registration")
	if err != nil {
		return UserType{}, err
	}

	var schemaDef string
	switch v := row["schema_def"].(type) {
	case string:
		schemaDef = v
	case []byte:
		schemaDef = string(v) // Convert byte slice to string
	default:
		return UserType{}, fmt.Errorf("failed to parse schema_def as string")
	}

	userType := UserType{
		ID:                    schemaID,
		Name:                  name,
		OrganizationUnitID:    organizationUnitID,
		AllowSelfRegistration: allowSelfRegistration,
		Schema:                json.RawMessage(schemaDef),
	}

	return userType, nil
}

// parseUserTypeListItemFromRow parses a simplified user type list item from a database row.
func parseUserTypeListItemFromRow(row map[string]interface{}) (UserTypeListItem, error) {
	schemaID, ok := row["schema_id"].(string)
	if !ok {
		return UserTypeListItem{}, fmt.Errorf("failed to parse schema_id as string")
	}

	name, ok := row["name"].(string)
	if !ok {
		return UserTypeListItem{}, fmt.Errorf("failed to parse name as string")
	}

	organizationUnitID, ok := row["ou_id"].(string)
	if !ok {
		return UserTypeListItem{}, fmt.Errorf("failed to parse ou_id as string")
	}

	allowSelfRegistration, err := parseBool(row["allow_self_registration"], "allow_self_registration")
	if err != nil {
		return UserTypeListItem{}, err
	}

	userTypeListItem := UserTypeListItem{
		ID:                    schemaID,
		Name:                  name,
		OrganizationUnitID:    organizationUnitID,
		AllowSelfRegistration: allowSelfRegistration,
	}

	return userTypeListItem, nil
}

func parseBool(value interface{}, fieldName string) (bool, error) {
	switch v := value.(type) {
	case nil:
		return false, fmt.Errorf("required boolean field '%s' is nil", fieldName)
	case bool:
		return v, nil
	case int64:
		return v != 0, nil
	case float64:
		return v != 0, nil
	case string:
		return strings.EqualFold(v, "true") || v == "1", nil
	case []byte:
		strVal := string(v)
		return strings.EqualFold(strVal, "true") || strVal == "1", nil
	default:
		return false, fmt.Errorf("failed to parse %s as bool", fieldName)
	}
}
