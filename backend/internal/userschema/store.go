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

package userschema

import (
	"encoding/json"
	"fmt"

	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/log"
)

// userSchemaStoreInterface defines the interface for user schema store operations.
type userSchemaStoreInterface interface {
	GetUserSchemaListCount() (int, error)
	GetUserSchemaList(limit, offset int) ([]UserSchemaListItem, error)
	CreateUserSchema(userSchema UserSchema) error
	GetUserSchemaByID(schemaID string) (UserSchema, error)
	GetUserSchemaByName(name string) (UserSchema, error)
	UpdateUserSchemaByID(schemaID string, userSchema UserSchema) error
	DeleteUserSchemaByID(schemaID string) error
}

// userSchemaStore is the default implementation of userSchemaStoreInterface.
type userSchemaStore struct {
	dbProvider provider.DBProviderInterface
}

// newUserSchemaStore creates a new instance of userSchemaStore.
func newUserSchemaStore() userSchemaStoreInterface {
	return &userSchemaStore{
		dbProvider: provider.GetDBProvider(),
	}
}

// GetUserSchemaListCount retrieves the total count of user schemas.
func (s *userSchemaStore) GetUserSchemaListCount() (int, error) {
	dbClient, err := s.dbProvider.GetDBClient("identity")
	if err != nil {
		return 0, fmt.Errorf("failed to get database client: %w", err)
	}

	countResults, err := dbClient.Query(queryGetUserSchemaCount)
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

// GetUserSchemaList retrieves a list of user schemas with pagination.
func (s *userSchemaStore) GetUserSchemaList(limit, offset int) ([]UserSchemaListItem, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserSchemaPersistence"))

	dbClient, err := s.dbProvider.GetDBClient("identity")
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(queryGetUserSchemaList, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	userSchemas := make([]UserSchemaListItem, 0, len(results))
	for _, row := range results {
		userSchema, err := parseUserSchemaListItemFromRow(row)
		if err != nil {
			logger.Error("Failed to parse user schema list item from row", log.Error(err))
			continue
		}
		userSchemas = append(userSchemas, userSchema)
	}

	return userSchemas, nil
}

// CreateUserSchema creates a new user schema.
func (s *userSchemaStore) CreateUserSchema(userSchema UserSchema) error {
	dbClient, err := s.dbProvider.GetDBClient("identity")
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	_, err = dbClient.Query(queryCreateUserSchema, userSchema.ID, userSchema.Name, string(userSchema.Schema))
	if err != nil {
		return fmt.Errorf("failed to create user schema: %w", err)
	}

	return nil
}

// GetUserSchemaByID retrieves a user schema by its ID.
func (s *userSchemaStore) GetUserSchemaByID(schemaID string) (UserSchema, error) {
	dbClient, err := s.dbProvider.GetDBClient("identity")
	if err != nil {
		return UserSchema{}, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(queryGetUserSchemaByID, schemaID)
	if err != nil {
		return UserSchema{}, fmt.Errorf("failed to execute query: %w", err)
	}

	if len(results) == 0 {
		return UserSchema{}, ErrUserSchemaNotFound
	}

	return parseUserSchemaFromRow(results[0])
}

// GetUserSchemaByName retrieves a user schema by its name.
func (s *userSchemaStore) GetUserSchemaByName(name string) (UserSchema, error) {
	dbClient, err := s.dbProvider.GetDBClient("identity")
	if err != nil {
		return UserSchema{}, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(queryGetUserSchemaByName, name)
	if err != nil {
		return UserSchema{}, fmt.Errorf("failed to execute query: %w", err)
	}

	if len(results) == 0 {
		return UserSchema{}, ErrUserSchemaNotFound
	}

	return parseUserSchemaFromRow(results[0])
}

// UpdateUserSchemaByID updates a user schema by its ID.
func (s *userSchemaStore) UpdateUserSchemaByID(schemaID string, userSchema UserSchema) error {
	dbClient, err := s.dbProvider.GetDBClient("identity")
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	_, err = dbClient.Query(queryUpdateUserSchemaByID, userSchema.Name, string(userSchema.Schema), schemaID)
	if err != nil {
		return fmt.Errorf("failed to update user schema: %w", err)
	}

	return nil
}

// DeleteUserSchemaByID deletes a user schema by its ID.
func (s *userSchemaStore) DeleteUserSchemaByID(schemaID string) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserSchemaPersistence"))

	dbClient, err := s.dbProvider.GetDBClient("identity")
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	rowsAffected, err := dbClient.Execute(queryDeleteUserSchemaByID, schemaID)
	if err != nil {
		return fmt.Errorf("failed to delete user schema: %w", err)
	}

	if rowsAffected == 0 {
		logger.Debug("user not found with id: " + schemaID)
	}

	return nil
}

// parseUserSchemaFromRow parses a user schema from a database row.
func parseUserSchemaFromRow(row map[string]interface{}) (UserSchema, error) {
	schemaID, ok := row["schema_id"].(string)
	if !ok {
		return UserSchema{}, fmt.Errorf("failed to parse schema_id as string")
	}

	name, ok := row["name"].(string)
	if !ok {
		return UserSchema{}, fmt.Errorf("failed to parse name as string")
	}

	var schemaDef string
	switch v := row["schema_def"].(type) {
	case string:
		schemaDef = v
	case []byte:
		schemaDef = string(v) // Convert byte slice to string
	default:
		return UserSchema{}, fmt.Errorf("failed to parse schema_def as string")
	}

	userSchema := UserSchema{
		ID:     schemaID,
		Name:   name,
		Schema: json.RawMessage(schemaDef),
	}

	return userSchema, nil
}

// parseUserSchemaListItemFromRow parses a simplified user schema list item from a database row.
func parseUserSchemaListItemFromRow(row map[string]interface{}) (UserSchemaListItem, error) {
	schemaID, ok := row["schema_id"].(string)
	if !ok {
		return UserSchemaListItem{}, fmt.Errorf("failed to parse schema_id as string")
	}

	name, ok := row["name"].(string)
	if !ok {
		return UserSchemaListItem{}, fmt.Errorf("failed to parse name as string")
	}

	userSchemaListItem := UserSchemaListItem{
		ID:   schemaID,
		Name: name,
	}

	return userSchemaListItem, nil
}
