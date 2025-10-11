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

// Package store provides the implementation for user persistence operations.
package store

import (
	"encoding/json"
	"fmt"

	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/user/constants"
	"github.com/asgardeo/thunder/internal/user/model"
)

// GetUserListCount retrieves the total count of users.
func GetUserListCount(filters map[string]interface{}) (int, error) {
	dbClient, err := provider.GetDBProvider().GetDBClient("identity")
	if err != nil {
		return 0, fmt.Errorf("failed to get database client: %w", err)
	}

	countQuery, args, err := buildUserCountQuery(filters)
	if err != nil {
		return 0, fmt.Errorf("failed to build count query: %w", err)
	}

	countResults, err := dbClient.Query(countQuery, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to execute count query: %w", err)
	}

	var totalCount int
	if len(countResults) > 0 {
		if count, ok := countResults[0]["total"].(int64); ok {
			totalCount = int(count)
		} else {
			return 0, fmt.Errorf("unexpected type for total: %T", countResults[0]["total"])
		}
	}

	return totalCount, nil
}

// GetUserList retrieves a list of users from the database.
func GetUserList(limit, offset int, filters map[string]interface{}) ([]model.User, error) {
	dbClient, err := provider.GetDBProvider().GetDBClient("identity")
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	listQuery, args, err := buildUserListQuery(filters, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to build list query: %w", err)
	}

	results, err := dbClient.Query(listQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute paginated query: %w", err)
	}

	users := make([]model.User, 0)

	for _, row := range results {
		user, err := buildUserFromResultRow(row)
		if err != nil {
			return nil, fmt.Errorf("failed to build user from result row: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

// CreateUser handles the user creation in the database.
func CreateUser(user model.User, credentials []model.Credential) error {
	dbClient, err := provider.GetDBProvider().GetDBClient("identity")
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	// Convert attributes to JSON string
	attributes, err := json.Marshal(user.Attributes)
	if err != nil {
		return constants.ErrBadAttributesInRequest
	}

	// Convert credentials array to JSON string
	var credentialsJSON string
	if len(credentials) == 0 {
		credentialsJSON = "[]"
	} else {
		credentialsBytes, err := json.Marshal(credentials)
		if err != nil {
			return constants.ErrBadAttributesInRequest
		}
		credentialsJSON = string(credentialsBytes)
	}

	_, err = dbClient.Execute(
		QueryCreateUser,
		user.ID,
		user.OrganizationUnit,
		user.Type,
		string(attributes),
		credentialsJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}

// GetUser retrieves a specific user by its ID from the database.
func GetUser(id string) (model.User, error) {
	dbClient, err := provider.GetDBProvider().GetDBClient("identity")
	if err != nil {
		return model.User{}, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(QueryGetUserByUserID, id)
	if err != nil {
		return model.User{}, fmt.Errorf("failed to execute query: %w", err)
	}

	if len(results) == 0 {
		return model.User{}, constants.ErrUserNotFound
	}

	if len(results) != 1 {
		return model.User{}, fmt.Errorf("unexpected number of results: %d", len(results))
	}

	row := results[0]

	user, err := buildUserFromResultRow(row)
	if err != nil {
		return model.User{}, fmt.Errorf("failed to build user from result row: %w", err)
	}
	return user, nil
}

// UpdateUser updates the user in the database.
func UpdateUser(user *model.User) error {
	dbClient, err := provider.GetDBProvider().GetDBClient("identity")
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	// Convert attributes to JSON string
	attributes, err := json.Marshal(user.Attributes)
	if err != nil {
		return constants.ErrBadAttributesInRequest
	}

	rowsAffected, err := dbClient.Execute(
		QueryUpdateUserByUserID, user.ID, user.OrganizationUnit, user.Type, string(attributes))
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	if rowsAffected == 0 {
		return constants.ErrUserNotFound
	}

	return nil
}

// DeleteUser deletes the user from the database.
func DeleteUser(id string) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserStore"))

	dbClient, err := provider.GetDBProvider().GetDBClient("identity")
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	rowsAffected, err := dbClient.Execute(QueryDeleteUserByUserID, id)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	if rowsAffected == 0 {
		logger.Debug("user not found with id: " + id)
	}

	return nil
}

// IdentifyUser identifies a user with the given filters.
func IdentifyUser(filters map[string]interface{}) (*string, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserStore"))

	dbClient, err := provider.GetDBProvider().GetDBClient("identity")
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	identifyUserQuery, args, err := buildIdentifyQuery(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to build identify query: %w", err)
	}

	results, err := dbClient.Query(identifyUserQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	if len(results) == 0 {
		if logger.IsDebugEnabled() {
			maskedFilters := maskMapValues(filters)
			logger.Debug("User not found with the provided filters", log.Any("filters", maskedFilters))
		}
		return nil, constants.ErrUserNotFound
	}

	if len(results) != 1 {
		if logger.IsDebugEnabled() {
			maskedFilters := maskMapValues(filters)
			logger.Debug(
				"Unexpected number of results for the provided filters",
				log.Any("filters", maskedFilters),
				log.Int("result_count", len(results)),
			)
		}
		return nil, fmt.Errorf("unexpected number of results: %d", len(results))
	}

	row := results[0]
	userID, ok := row["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to parse user_id as string")
	}

	return &userID, nil
}

// VerifyUser validate the user specified user using the given credentials from the database.
func VerifyUser(id string) (model.User, []model.Credential, error) {
	dbClient, err := provider.GetDBProvider().GetDBClient("identity")
	if err != nil {
		return model.User{}, []model.Credential{}, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(QueryValidateUserWithCredentials, id)
	if err != nil {
		return model.User{}, []model.Credential{}, fmt.Errorf("failed to execute query: %w", err)
	}

	if len(results) == 0 {
		return model.User{}, []model.Credential{}, constants.ErrUserNotFound
	}

	if len(results) != 1 {
		return model.User{}, []model.Credential{}, fmt.Errorf("unexpected number of results: %d", len(results))
	}

	row := results[0]

	user, err := buildUserFromResultRow(row)
	if err != nil {
		return model.User{}, []model.Credential{}, fmt.Errorf("failed to build user from result row: %w", err)
	}

	// build the UserDTO with credentials.
	var credentialsJSON string
	switch v := row["credentials"].(type) {
	case string:
		credentialsJSON = v
	case []byte:
		credentialsJSON = string(v)
	default:
		return model.User{}, []model.Credential{}, fmt.Errorf("failed to parse credentials as string")
	}

	var credentials []model.Credential
	if err := json.Unmarshal([]byte(credentialsJSON), &credentials); err != nil {
		return model.User{}, []model.Credential{}, fmt.Errorf("failed to unmarshal credentials: %w", err)
	}

	return user, credentials, nil
}

// ValidateUserIDs checks if all provided user IDs exist.
func ValidateUserIDs(userIDs []string) ([]string, error) {
	if len(userIDs) == 0 {
		return []string{}, nil
	}

	dbClient, err := provider.GetDBProvider().GetDBClient("identity")
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	query, args, err := buildBulkUserExistsQuery(userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to build bulk user exists query: %w", err)
	}

	results, err := dbClient.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	existingUserIDs := make(map[string]bool)
	for _, row := range results {
		if userID, ok := row["user_id"].(string); ok {
			existingUserIDs[userID] = true
		}
	}

	var invalidUserIDs []string
	for _, userID := range userIDs {
		if !existingUserIDs[userID] {
			invalidUserIDs = append(invalidUserIDs, userID)
		}
	}

	return invalidUserIDs, nil
}

// GetGroupCountForUser retrieves the total count of groups a user belongs to.
func GetGroupCountForUser(userID string) (int, error) {
	dbClient, err := provider.GetDBProvider().GetDBClient("identity")
	if err != nil {
		return 0, fmt.Errorf("failed to get database client: %w", err)
	}

	countResults, err := dbClient.Query(QueryGetGroupCountForUser, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get group count for user: %w", err)
	}

	if len(countResults) == 0 {
		return 0, nil
	}

	if count, ok := countResults[0]["total"].(int64); ok {
		return int(count), nil
	}
	return 0, fmt.Errorf("unexpected type for total: %T", countResults[0]["total"])
}

// GetUserGroups retrieves groups that a user belongs to with pagination.
func GetUserGroups(userID string, limit, offset int) ([]model.UserGroup, error) {
	dbClient, err := provider.GetDBProvider().GetDBClient("identity")
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(QueryGetGroupsForUser, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get groups for user: %w", err)
	}

	groups := make([]model.UserGroup, 0, len(results))
	for _, row := range results {
		group, err := buildGroupFromResultRow(row)
		if err != nil {
			return nil, fmt.Errorf("failed to build group from result row: %w", err)
		}

		groups = append(groups, group)
	}

	return groups, nil
}

func buildUserFromResultRow(row map[string]interface{}) (model.User, error) {
	userID, ok := row["user_id"].(string)
	if !ok {
		return model.User{}, fmt.Errorf("failed to parse user_id as string")
	}

	orgID, ok := row["ou_id"].(string)
	if !ok {
		return model.User{}, fmt.Errorf("failed to parse org_id as string")
	}

	userType, ok := row["type"].(string)
	if !ok {
		return model.User{}, fmt.Errorf("failed to parse type as string")
	}

	var attributes string
	switch v := row["attributes"].(type) {
	case string:
		attributes = v
	case []byte:
		attributes = string(v) // Convert byte slice to string
	default:
		return model.User{}, fmt.Errorf("failed to parse attributes as string")
	}

	user := model.User{
		ID:               userID,
		OrganizationUnit: orgID,
		Type:             userType,
	}

	// Unmarshal JSON attributes
	if err := json.Unmarshal([]byte(attributes), &user.Attributes); err != nil {
		return model.User{}, fmt.Errorf("failed to unmarshal attributes")
	}

	return user, nil
}

// buildGroupFromResultRow constructs a model.UserGroup from a database result row.
func buildGroupFromResultRow(row map[string]interface{}) (model.UserGroup, error) {
	groupID, ok := row["group_id"].(string)
	if !ok {
		return model.UserGroup{}, fmt.Errorf("failed to parse group_id as string")
	}

	name, ok := row["name"].(string)
	if !ok {
		return model.UserGroup{}, fmt.Errorf("failed to parse name as string")
	}

	ouID, ok := row["ou_id"].(string)
	if !ok {
		return model.UserGroup{}, fmt.Errorf("failed to parse ou_id as string")
	}

	group := model.UserGroup{
		ID:                 groupID,
		Name:               name,
		OrganizationUnitID: ouID,
	}

	return group, nil
}

// maskMapValues masks the values in a map to prevent sensitive data from being logged.
func maskMapValues(input map[string]interface{}) map[string]interface{} {
	masked := make(map[string]interface{})
	for key, value := range input {
		if strValue, ok := value.(string); ok {
			masked[key] = log.MaskString(strValue)
		} else {
			masked[key] = "***"
		}
	}
	return masked
}
