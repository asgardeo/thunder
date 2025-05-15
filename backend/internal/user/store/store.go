/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
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

package store

import (
	"encoding/json"
	"fmt"
	"github.com/asgardeo/thunder/internal/system/database/client"
	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/user/model"
)

func CreateUser(user model.User) error {

	logger := log.GetLogger().With(log.String(log.LOGGER_KEY_COMPONENT_NAME, "UserPersistence"))

	dbClient, err := provider.NewDBProvider().GetDBClient("identity")
	if err != nil {
		logger.Error("Failed to get database client", log.Error(err))
		return fmt.Errorf("failed to get database client: %w", err)
	}
	defer func(dbc client.DBClientInterface) {
		err := dbc.Close()
		if err != nil {
			logger.Error("Failed to close database client", log.Error(err))
			err = fmt.Errorf("failed to close database client: %w", err)
		}
	}(dbClient)

	// Convert attributes to JSON string
	attributes, err := json.Marshal(user.Attributes)
	if err != nil {
		logger.Error("Failed to marshal attributes", log.Error(err))
		return fmt.Errorf("failed to marshal attributes: %w", err)
	}

	_, err = dbClient.ExecuteQuery(QueryCreateUser, user.Id, user.OrgId, user.Type, string(attributes))
	if err != nil {
		logger.Error("Failed to execute query", log.Error(err))
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}

func GetUserList() ([]model.User, error) {

	logger := log.GetLogger().With(log.String(log.LOGGER_KEY_COMPONENT_NAME, "UserPersistence"))

	dbClient, err := provider.NewDBProvider().GetDBClient("identity")
	if err != nil {
		logger.Error("Failed to get database client", log.Error(err))
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}
	defer func(dbc client.DBClientInterface) {
		err := dbc.Close()
		if err != nil {
			logger.Error("Failed to close database client", log.Error(err))
			err = fmt.Errorf("failed to close database client: %w", err)
		}
	}(dbClient)

	results, err := dbClient.ExecuteQuery(QueryGetUserList)
	if err != nil {
		logger.Error("Failed to execute query", log.Error(err))
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	users := make([]model.User, 0)

	for _, row := range results {
		user, err := buildUserFromResultRow(row)
		if err != nil {
			logger.Error("failed to build user from result row", log.Error(err))
			return nil, fmt.Errorf("failed to build user from result row: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

func GetUser(id string) (model.User, error) {

	logger := log.GetLogger().With(log.String(log.LOGGER_KEY_COMPONENT_NAME, "UserStore"))

	dbClient, err := provider.NewDBProvider().GetDBClient("identity")
	if err != nil {
		logger.Error("Failed to get database client", log.Error(err))
		return model.User{}, fmt.Errorf("failed to get database client: %w", err)
	}
	defer dbClient.Close()

	results, err := dbClient.ExecuteQuery(QueryGetUserByUserId, id)
	if err != nil {
		logger.Error("Failed to execute query", log.Error(err))
		return model.User{}, fmt.Errorf("failed to execute query: %w", err)
	}

	if len(results) == 0 {
		logger.Error("user not found with id: " + id)
		return model.User{}, fmt.Errorf("user not found")
	}

	if len(results) != 1 {
		logger.Error("unexpected number of results")
		return model.User{}, fmt.Errorf("unexpected number of results: %d", len(results))
	}

	row := results[0]

	user, err := buildUserFromResultRow(row)
	if err != nil {
		logger.Error("failed to build user from result row")
		return model.User{}, fmt.Errorf("failed to build user from result row: %w", err)
	}
	return user, nil
}

func UpdateUser(user *model.User) error {

	logger := log.GetLogger().With(log.String(log.LOGGER_KEY_COMPONENT_NAME, "UserStore"))

	dbClient, err := provider.NewDBProvider().GetDBClient("identity")
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}
	defer dbClient.Close()

	tx, err := dbClient.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Convert attributes to JSON string
	attributes, err := json.Marshal(user.Attributes)
	if err != nil {
		logger.Error("Failed to marshal attributes", log.Error(err))
		return fmt.Errorf("failed to marshal attributes: %w", err)
	}

	_, err = tx.Exec(QueryUpdateUserByUserId.Query, user.Id, user.OrgId, user.Type, string(attributes))
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create SP user: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func DeleteUser(id string) error {

	logger := log.GetLogger().With(log.String(log.LOGGER_KEY_COMPONENT_NAME, "UserStore"))

	dbClient, err := provider.NewDBProvider().GetDBClient("identity")
	if err != nil {
		logger.Error("Failed to get database client", log.Error(err))
		return fmt.Errorf("failed to get database client: %w", err)
	}
	defer dbClient.Close()

	_, err = dbClient.ExecuteQuery(QueryDeleteUserByUserId, id)
	if err != nil {
		logger.Error("Failed to execute query", log.Error(err))
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}

func buildUserFromResultRow(row map[string]interface{}) (model.User, error) {

	logger := log.GetLogger().With(log.String(log.LOGGER_KEY_COMPONENT_NAME, "UserStore"))

	userId, ok := row["user_id"].(string)
	if !ok {
		logger.Error("failed to parse user_id as string")
		return model.User{}, fmt.Errorf("failed to parse user_id as string")
	}

	orgId, ok := row["org_id"].(string)
	if !ok {
		logger.Error("failed to parse org_id as string")
		return model.User{}, fmt.Errorf("failed to parse org_id as string")
	}

	userType, ok := row["type"].(string)
	if !ok {
		logger.Error("failed to parse type as string")
		return model.User{}, fmt.Errorf("failed to parse type as string")
	}

	var attributes string
	switch v := row["attributes"].(type) {
	case string:
		attributes = v
	case []byte:
		attributes = string(v) // Convert byte slice to string
	default:
		logger.Error("failed to parse attributes", log.Any("raw_value", row["attributes"]), log.String("type", fmt.Sprintf("%T", row["attributes"])))
		return model.User{}, fmt.Errorf("failed to parse attributes as string")
	}

	user := model.User{
		Id:    userId,
		OrgId: orgId,
		Type:  userType,
	}

	// Unmarshal JSON attributes
	if err := json.Unmarshal([]byte(attributes), &user.Attributes); err != nil {
		logger.Error("Failed to unmarshal attributes")
		return model.User{}, fmt.Errorf("failed to unmarshal attributes")
	}

	return user, nil
}
