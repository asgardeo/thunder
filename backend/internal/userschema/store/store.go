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

// Package store provides the implementation for user schema persistence operations.
package store

import (
	"encoding/json"
	"fmt"

	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/userschema/constants"
	"github.com/asgardeo/thunder/internal/userschema/model"
)

// GetUserSchemaListCount retrieves the total count of user schemas.
func GetUserSchemaListCount() (int, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserSchemaPersistence"))

	dbClient, err := provider.NewDBProvider().GetDBClient("identity")
	if err != nil {
		return 0, fmt.Errorf("failed to get database client: %w", err)
	}
	defer func() {
		if closeErr := dbClient.Close(); closeErr != nil {
			logger.Error("Failed to close database client", log.Error(closeErr))
		}
	}()

	countResults, err := dbClient.Query(QueryGetUserSchemaCount)
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
func GetUserSchemaList(limit, offset int) ([]model.UserSchemaListItem, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserSchemaPersistence"))

	dbClient, err := provider.NewDBProvider().GetDBClient("identity")
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}
	defer func() {
		if closeErr := dbClient.Close(); closeErr != nil {
			logger.Error("Failed to close database client", log.Error(closeErr))
		}
	}()

	results, err := dbClient.Query(QueryGetUserSchemaList, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	userSchemas := make([]model.UserSchemaListItem, 0, len(results))
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
func CreateUserSchema(userSchema model.UserSchema) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserSchemaPersistence"))

	dbClient, err := provider.NewDBProvider().GetDBClient("identity")
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}
	defer func() {
		if closeErr := dbClient.Close(); closeErr != nil {
			logger.Error("Failed to close database client", log.Error(closeErr))
		}
	}()

	_, err = dbClient.Query(QueryCreateUserSchema, userSchema.ID, userSchema.Name, string(userSchema.Schema))
	if err != nil {
		return fmt.Errorf("failed to create user schema: %w", err)
	}

	return nil
}

// GetUserSchemaByID retrieves a user schema by its ID.
func GetUserSchemaByID(schemaID string) (model.UserSchema, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserSchemaPersistence"))

	dbClient, err := provider.NewDBProvider().GetDBClient("identity")
	if err != nil {
		return model.UserSchema{}, fmt.Errorf("failed to get database client: %w", err)
	}
	defer func() {
		if closeErr := dbClient.Close(); closeErr != nil {
			logger.Error("Failed to close database client", log.Error(closeErr))
		}
	}()

	results, err := dbClient.Query(QueryGetUserSchemaByID, schemaID)
	if err != nil {
		return model.UserSchema{}, fmt.Errorf("failed to execute query: %w", err)
	}

	if len(results) == 0 {
		return model.UserSchema{}, constants.ErrUserSchemaNotFound
	}

	return parseUserSchemaFromRow(results[0])
}

// GetUserSchemaByName retrieves a user schema by its name.
func GetUserSchemaByName(name string) (model.UserSchema, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserSchemaPersistence"))

	dbClient, err := provider.NewDBProvider().GetDBClient("identity")
	if err != nil {
		return model.UserSchema{}, fmt.Errorf("failed to get database client: %w", err)
	}
	defer func() {
		if closeErr := dbClient.Close(); closeErr != nil {
			logger.Error("Failed to close database client", log.Error(closeErr))
		}
	}()

	results, err := dbClient.Query(QueryGetUserSchemaByName, name)
	if err != nil {
		return model.UserSchema{}, fmt.Errorf("failed to execute query: %w", err)
	}

	if len(results) == 0 {
		return model.UserSchema{}, constants.ErrUserSchemaNotFound
	}

	return parseUserSchemaFromRow(results[0])
}

// UpdateUserSchemaByID updates a user schema by its ID.
func UpdateUserSchemaByID(schemaID string, userSchema model.UserSchema) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserSchemaPersistence"))

	dbClient, err := provider.NewDBProvider().GetDBClient("identity")
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}
	defer func() {
		if closeErr := dbClient.Close(); closeErr != nil {
			logger.Error("Failed to close database client", log.Error(closeErr))
		}
	}()

	_, err = dbClient.Query(QueryUpdateUserSchemaByID, userSchema.Name, string(userSchema.Schema), schemaID)
	if err != nil {
		return fmt.Errorf("failed to update user schema: %w", err)
	}

	return nil
}

// DeleteUserSchemaByID deletes a user schema by its ID.
func DeleteUserSchemaByID(schemaID string) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserSchemaPersistence"))

	dbClient, err := provider.NewDBProvider().GetDBClient("identity")
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}
	defer func() {
		if closeErr := dbClient.Close(); closeErr != nil {
			logger.Error("Failed to close database client", log.Error(closeErr))
		}
	}()

	rowsAffected, err := dbClient.Execute(QueryDeleteUserSchemaByID, schemaID)
	if err != nil {
		return fmt.Errorf("failed to delete user schema: %w", err)
	}

	if rowsAffected == 0 {
		logger.Debug("user not found with id: " + schemaID)
	}

	return nil
}

// parseUserSchemaFromRow parses a user schema from a database row.
func parseUserSchemaFromRow(row map[string]interface{}) (model.UserSchema, error) {
	schemaID, ok := row["schema_id"].(string)
	if !ok {
		return model.UserSchema{}, fmt.Errorf("failed to parse schema_id as string")
	}

	name, ok := row["name"].(string)
	if !ok {
		return model.UserSchema{}, fmt.Errorf("failed to parse name as string")
	}

	var schemaDef string
	switch v := row["schema_def"].(type) {
	case string:
		schemaDef = v
	case []byte:
		schemaDef = string(v) // Convert byte slice to string
	default:
		return model.UserSchema{}, fmt.Errorf("failed to parse schema_def as string")
	}

	userSchema := model.UserSchema{
		ID:     schemaID,
		Name:   name,
		Schema: json.RawMessage(schemaDef),
	}

	return userSchema, nil
}

// parseUserSchemaListItemFromRow parses a simplified user schema list item from a database row.
func parseUserSchemaListItemFromRow(row map[string]interface{}) (model.UserSchemaListItem, error) {
	schemaID, ok := row["schema_id"].(string)
	if !ok {
		return model.UserSchemaListItem{}, fmt.Errorf("failed to parse schema_id as string")
	}

	name, ok := row["name"].(string)
	if !ok {
		return model.UserSchemaListItem{}, fmt.Errorf("failed to parse name as string")
	}

	userSchemaListItem := model.UserSchemaListItem{
		ID:   schemaID,
		Name: name,
	}

	return userSchemaListItem, nil
}
