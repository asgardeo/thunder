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

package user

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/asgardeo/thunder/internal/system/config"
	dbmodel "github.com/asgardeo/thunder/internal/system/database/model"
	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/log"
)

// userStoreInterface defines the interface for user store operations.
type userStoreInterface interface {
	GetUserListCount(ctx context.Context, filters map[string]interface{}) (int, error)
	GetUserList(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]User, error)
	CreateUser(ctx context.Context, user User, credentials Credentials) error
	GetUser(ctx context.Context, id string) (User, error)
	GetGroupCountForUser(ctx context.Context, userID string) (int, error)
	GetUserGroups(ctx context.Context, userID string, limit, offset int) ([]UserGroup, error)
	UpdateUser(ctx context.Context, user *User) error
	UpdateUserCredentials(ctx context.Context, userID string, credentials Credentials) error
	DeleteUser(ctx context.Context, id string) error
	IdentifyUser(ctx context.Context, filters map[string]interface{}) (*string, error)
	GetCredentials(ctx context.Context, id string) (User, Credentials, error)
	ValidateUserIDs(ctx context.Context, userIDs []string) ([]string, error)
}

// userStore is the default implementation of userStoreInterface.
//
// indexedAttributes: Set of attribute keys that are indexed; immutable after initialization.
type userStore struct {
	deploymentID      string
	indexedAttributes map[string]bool
	dbProvider        provider.DBProviderInterface
}

// newUserStore creates a new instance of userStore.
func newUserStore() (userStoreInterface, error) {
	runtime := config.GetThunderRuntime()

	indexedAttributesFromConfig := runtime.Config.User.IndexedAttributes

	if err := validateIndexedAttributesConfig(indexedAttributesFromConfig); err != nil {
		return nil, fmt.Errorf("indexed attributes configuration validation failed: %w", err)
	}
	indexedAttributes := make(map[string]bool, len(indexedAttributesFromConfig))
	for _, attr := range indexedAttributesFromConfig {
		indexedAttributes[attr] = true
	}

	return &userStore{
		deploymentID:      runtime.Config.Server.Identifier,
		indexedAttributes: indexedAttributes,
		dbProvider:        provider.GetDBProvider(),
	}, nil
}

// GetUserListCount retrieves the total count of users.
func (us *userStore) GetUserListCount(ctx context.Context, filters map[string]interface{}) (int, error) {
	dbClient, err := us.dbProvider.GetUserDBClient()
	if err != nil {
		return 0, fmt.Errorf("failed to get database client: %w", err)
	}

	countQuery, args, err := buildUserCountQuery(filters, us.deploymentID)
	if err != nil {
		return 0, fmt.Errorf("failed to build count query: %w", err)
	}

	countResults, err := dbClient.QueryContext(ctx, countQuery, args...)
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
func (us *userStore) GetUserList(ctx context.Context, limit, offset int,
	filters map[string]interface{}) ([]User, error) {
	dbClient, err := us.dbProvider.GetUserDBClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	listQuery, args, err := buildUserListQuery(filters, limit, offset, us.deploymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to build list query: %w", err)
	}

	results, err := dbClient.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute paginated query: %w", err)
	}

	users := make([]User, 0)

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
func (us *userStore) CreateUser(ctx context.Context, user User, credentials Credentials) error {
	dbClient, err := us.dbProvider.GetUserDBClient()
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	// Convert attributes to JSON string
	attributes, err := json.Marshal(user.Attributes)
	if err != nil {
		return ErrBadAttributesInRequest
	}

	// Convert credentials map to JSON string
	var credentialsJSON string
	if len(credentials) == 0 {
		credentialsJSON = "{}"
	} else {
		credentialsBytes, err := json.Marshal(credentials)
		if err != nil {
			return ErrBadAttributesInRequest
		}
		credentialsJSON = string(credentialsBytes)
	}

	// Insert user using context-aware method
	_, err = dbClient.ExecuteContext(
		ctx,
		QueryCreateUser,
		user.ID,
		user.OrganizationUnit,
		user.Type,
		string(attributes),
		credentialsJSON,
		us.deploymentID,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Sync indexed attributes using context
	if err := us.syncIndexedAttributes(ctx, dbClient, user.ID, user.Attributes); err != nil {
		return fmt.Errorf("failed to sync indexed attributes: %w", err)
	}

	return nil
}

// GetUser retrieves a user by ID.
func (us *userStore) GetUser(ctx context.Context, id string) (User, error) {
	dbClient, err := us.dbProvider.GetUserDBClient()
	if err != nil {
		return User{}, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.QueryContext(ctx, QueryGetUserByUserID, id, us.deploymentID)
	if err != nil {
		return User{}, fmt.Errorf("failed to execute query: %w", err)
	}

	if len(results) == 0 {
		return User{}, ErrUserNotFound
	}

	if len(results) != 1 {
		return User{}, fmt.Errorf("unexpected number of results: %d", len(results))
	}

	row := results[0]

	user, err := buildUserFromResultRow(row)
	if err != nil {
		return User{}, fmt.Errorf("failed to build user from result row: %w", err)
	}
	return user, nil
}

// UpdateUser updates the user in the database.
func (us *userStore) UpdateUser(ctx context.Context, user *User) error {
	dbClient, err := us.dbProvider.GetUserDBClient()
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	attributes, err := json.Marshal(user.Attributes)
	if err != nil {
		return fmt.Errorf("failed to marshal attributes: %w", err)
	}

	// Update user
	rowsAffected, err := dbClient.ExecuteContext(
		ctx,
		QueryUpdateUserByUserID, user.ID, user.OrganizationUnit, user.Type, string(attributes), us.deploymentID)
	if err != nil {
		return fmt.Errorf("failed to execute update user query: %w", err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	// Delete existing indexed attributes
	_, err = dbClient.ExecuteContext(
		ctx,
		QueryDeleteIndexedAttributesByUser,
		user.ID,
		us.deploymentID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete indexed attributes: %w", err)
	}

	// Sync new indexed attributes
	if err := us.syncIndexedAttributes(ctx, dbClient, user.ID, user.Attributes); err != nil {
		return fmt.Errorf("failed to sync indexed attributes: %w", err)
	}

	return nil
}

// UpdateUserCredentials updates the credentials of a user.
func (us *userStore) UpdateUserCredentials(ctx context.Context, userID string, credentials Credentials) error {
	dbClient, err := us.dbProvider.GetUserDBClient()
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	credentialsJSON, err := json.Marshal(credentials)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	rowsAffected, err := dbClient.ExecuteContext(ctx, QueryUpdateUserCredentialsByUserID,
		userID, string(credentialsJSON), us.deploymentID)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// DeleteUser deletes the user from the database.
func (us *userStore) DeleteUser(ctx context.Context, id string) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserStore"))

	dbClient, err := us.dbProvider.GetUserDBClient()
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	rowsAffected, err := dbClient.ExecuteContext(ctx, QueryDeleteUserByUserID, id, us.deploymentID)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	if rowsAffected == 0 {
		logger.Debug("user not found with id: " + id)
	}

	return nil
}

// IdentifyUser identifies a user with the given filters.
func (us *userStore) IdentifyUser(ctx context.Context, filters map[string]interface{}) (*string, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserStore"))

	dbClient, err := us.dbProvider.GetUserDBClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	// Categorize filters into indexed and non-indexed
	indexedFilters := make(map[string]interface{})
	nonIndexedFilters := make(map[string]interface{})

	for key, value := range filters {
		if us.isAttributeIndexed(key) {
			indexedFilters[key] = value
		} else {
			nonIndexedFilters[key] = value
		}
	}

	// Determine which query strategy to use based on indexed attribute coverage
	var identifyUserQuery dbmodel.DBQuery
	var args []interface{}

	if len(indexedFilters) == len(filters) && len(indexedFilters) > 0 {
		// Case 1: All filters are indexed - use indexed attributes table only
		identifyUserQuery, args, err = buildIdentifyQueryFromIndexedAttributes(filters, us.deploymentID)
		if err != nil {
			return nil, fmt.Errorf("failed to build indexed query: %w", err)
		}
	} else if len(indexedFilters) > 0 {
		// Case 2: Partial indexed - use hybrid approach (indexed + JSON)
		identifyUserQuery, args, err = buildIdentifyQueryHybrid(indexedFilters, nonIndexedFilters, us.deploymentID)
		if err != nil {
			return nil, fmt.Errorf("failed to build hybrid query: %w", err)
		}
	} else {
		// Case 3: No indexed filters - fallback to JSON query
		identifyUserQuery, args, err = buildIdentifyQuery(filters, us.deploymentID)
		if err != nil {
			return nil, fmt.Errorf("failed to build identify query: %w", err)
		}
	}

	results, err := dbClient.QueryContext(ctx, identifyUserQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	if len(results) == 0 {
		if logger.IsDebugEnabled() {
			maskedFilters := maskMapValues(filters)
			logger.Debug("User not found with the provided filters", log.Any("filters", maskedFilters))
		}
		return nil, ErrUserNotFound
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

// GetCredentials retrieves the credentials for a user.
func (us *userStore) GetCredentials(ctx context.Context, id string) (User, Credentials, error) {
	dbClient, err := us.dbProvider.GetUserDBClient()
	if err != nil {
		return User{}, nil, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.QueryContext(ctx, QueryValidateUserWithCredentials, id, us.deploymentID)
	if err != nil {
		return User{}, nil, fmt.Errorf("failed to execute query: %w", err)
	}

	if len(results) == 0 {
		return User{}, nil, ErrUserNotFound
	}

	if len(results) != 1 {
		return User{}, nil, fmt.Errorf("unexpected number of results: %d", len(results))
	}

	row := results[0]

	user, err := buildUserFromResultRow(row)
	if err != nil {
		return User{}, nil, fmt.Errorf("failed to build user from result row: %w", err)
	}

	// build the UserDTO with credentials.
	var credentialsJSON string
	switch v := row["credentials"].(type) {
	case string:
		credentialsJSON = v
	case []byte:
		credentialsJSON = string(v)
	default:
		return User{}, nil, fmt.Errorf("failed to parse credentials as string")
	}

	var credentials Credentials
	if err := json.Unmarshal([]byte(credentialsJSON), &credentials); err != nil {
		return User{}, nil, fmt.Errorf("failed to unmarshal credentials: %w", err)
	}

	return user, credentials, nil
}

// ValidateUserIDs checks if all provided user IDs exist.
func (us *userStore) ValidateUserIDs(ctx context.Context, userIDs []string) ([]string, error) {
	if len(userIDs) == 0 {
		return []string{}, nil
	}

	dbClient, err := us.dbProvider.GetUserDBClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	query, args, err := buildBulkUserExistsQuery(userIDs, us.deploymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to build bulk user exists query: %w", err)
	}

	results, err := dbClient.QueryContext(ctx, query, args...)
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
func (us *userStore) GetGroupCountForUser(ctx context.Context, userID string) (int, error) {
	dbClient, err := us.dbProvider.GetUserDBClient()
	if err != nil {
		return 0, fmt.Errorf("failed to get database client: %w", err)
	}

	countResults, err := dbClient.QueryContext(ctx, QueryGetGroupCountForUser, userID, us.deploymentID)
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
func (us *userStore) GetUserGroups(ctx context.Context, userID string, limit, offset int) ([]UserGroup, error) {
	dbClient, err := us.dbProvider.GetUserDBClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.QueryContext(ctx, QueryGetGroupsForUser, userID, limit, offset, us.deploymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get groups for user: %w", err)
	}

	groups := make([]UserGroup, 0, len(results))
	for _, row := range results {
		group, err := buildGroupFromResultRow(row)
		if err != nil {
			return nil, fmt.Errorf("failed to build group from result row: %w", err)
		}

		groups = append(groups, group)
	}

	return groups, nil
}

func buildUserFromResultRow(row map[string]interface{}) (User, error) {
	userID, ok := row["user_id"].(string)
	if !ok {
		return User{}, fmt.Errorf("failed to parse user_id as string")
	}

	orgID, ok := row["ou_id"].(string)
	if !ok {
		return User{}, fmt.Errorf("failed to parse org_id as string")
	}

	userType, ok := row["type"].(string)
	if !ok {
		return User{}, fmt.Errorf("failed to parse type as string")
	}

	var attributes string
	switch v := row["attributes"].(type) {
	case string:
		attributes = v
	case []byte:
		attributes = string(v) // Convert byte slice to string
	default:
		return User{}, fmt.Errorf("failed to parse attributes as string")
	}

	user := User{
		ID:               userID,
		OrganizationUnit: orgID,
		Type:             userType,
	}

	// Unmarshal JSON attributes
	if err := json.Unmarshal([]byte(attributes), &user.Attributes); err != nil {
		return User{}, fmt.Errorf("failed to unmarshal attributes")
	}

	return user, nil
}

// buildGroupFromResultRow constructs a UserGroup from a database result row.
func buildGroupFromResultRow(row map[string]interface{}) (UserGroup, error) {
	groupID, ok := row["group_id"].(string)
	if !ok {
		return UserGroup{}, fmt.Errorf("failed to parse group_id as string")
	}

	name, ok := row["name"].(string)
	if !ok {
		return UserGroup{}, fmt.Errorf("failed to parse name as string")
	}

	ouID, ok := row["ou_id"].(string)
	if !ok {
		return UserGroup{}, fmt.Errorf("failed to parse ou_id as string")
	}

	group := UserGroup{
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

// syncIndexedAttributes synchronizes indexed attributes to the
// USER_INDEXED_ATTRIBUTES table using context-aware transaction execution.
func (us *userStore) syncIndexedAttributes(
	ctx context.Context, dbClient provider.DBClientInterface, userID string, attributes json.RawMessage) error {
	query, args, err := us.prepareIndexedAttributesQuery(userID, attributes)
	if err != nil {
		return err
	}
	if query == nil {
		return nil
	}

	// Execute batch insert using context-aware method
	_, err = dbClient.ExecuteContext(ctx, *query, args...)
	if err != nil {
		return fmt.Errorf("failed to batch insert indexed attributes (query ID: %s): %w",
			QueryBatchInsertIndexedAttributes.ID, err)
	}

	return nil
}

// prepareIndexedAttributesQuery matches attributes with configuration and builds the batch insert query.
func (us *userStore) prepareIndexedAttributesQuery(
	userID string, attributes json.RawMessage,
) (*dbmodel.DBQuery, []interface{}, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserStore"))

	if len(attributes) == 0 {
		return nil, nil, nil
	}

	// Parse user attributes
	var userAttrs map[string]interface{}
	if err := json.Unmarshal(attributes, &userAttrs); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal user attributes: %w", err)
	}

	// Collect indexed attributes to insert
	type indexedAttr struct {
		name  string
		value string
	}
	toInsert := make([]indexedAttr, 0, len(userAttrs))

	for attrName, attrValue := range userAttrs {
		// Check if this attribute should be indexed
		if !us.isAttributeIndexed(attrName) {
			continue
		}

		// Convert attribute value to string for storage
		var valueStr string
		switch v := attrValue.(type) {
		case string:
			valueStr = v
		case float64, int, int64, bool:
			valueStr = fmt.Sprintf("%v", v)
		default:
			// Skip complex types (objects, arrays)
			logger.Warn("Skipping indexing complex attribute; only primitive types are indexed",
				log.String("attribute", attrName))
			continue
		}

		toInsert = append(toInsert, indexedAttr{name: attrName, value: valueStr})
	}

	// Return early if no indexed attributes to insert
	if len(toInsert) == 0 {
		return nil, nil, nil
	}

	// Build batch INSERT query using the pre-defined query constant
	valuePlaceholders := make([]string, 0, len(toInsert))
	args := make([]interface{}, 0, len(toInsert)*4)
	paramIndex := 1

	for _, attr := range toInsert {
		valuePlaceholders = append(valuePlaceholders,
			fmt.Sprintf("($%d, $%d, $%d, $%d)", paramIndex, paramIndex+1, paramIndex+2, paramIndex+3))
		args = append(args, userID, attr.name, attr.value, us.deploymentID)
		paramIndex += 4
	}

	// Construct the complete query with dynamic VALUES placeholders
	queryStr := QueryBatchInsertIndexedAttributes.Query + strings.Join(valuePlaceholders, ", ")
	query := &dbmodel.DBQuery{
		ID:    QueryBatchInsertIndexedAttributes.ID,
		Query: queryStr,
	}

	return query, args, nil
}

// isAttributeIndexed checks if a given attribute is configured to be indexed.
// This is an O(1) operation.
func (us *userStore) isAttributeIndexed(attributeName string) bool {
	return us.indexedAttributes[attributeName]
}

// validateIndexedAttributesConfig validates that the current configuration
// matches the previously stored configuration in the database.
func validateIndexedAttributesConfig(configuredAttrs []string) error {
	// Validate that indexed attributes count is less than the maximum allowed
	if len(configuredAttrs) > MaxIndexedAttributesCount {
		return fmt.Errorf("indexed attributes count (%d) must not exceed %d",
			len(configuredAttrs), MaxIndexedAttributesCount)
	}
	return nil
}
