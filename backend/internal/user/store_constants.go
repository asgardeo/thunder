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
	"fmt"
	"sort"
	"strings"

	"github.com/asgardeo/thunder/internal/system/database/model"
	"github.com/asgardeo/thunder/internal/system/database/utils"
)

const (
	// AttributesColumn represents the ATTRIBUTES column name in the database.
	AttributesColumn = "ATTRIBUTES"

	// MaxIndexedAttributesCount is the maximum number of indexed attributes allowed.
	MaxIndexedAttributesCount = 20
)

var (
	// QueryGetUserCount is the query to get total count of users.
	QueryGetUserCount = model.DBQuery{
		ID:    "ASQ-USER_MGT-01",
		Query: "SELECT COUNT(*) as total FROM \"USER\" WHERE DEPLOYMENT_ID = $1",
	}
	// QueryGetUserList is the query to get a list of users.
	QueryGetUserList = model.DBQuery{
		ID: "ASQ-USER_MGT-02",
		Query: "SELECT USER_ID, OU_ID, TYPE, ATTRIBUTES FROM \"USER\" " +
			"WHERE DEPLOYMENT_ID = $3 ORDER BY USER_ID LIMIT $1 OFFSET $2",
	}
	// QueryCreateUser is the query to create a new user.
	QueryCreateUser = model.DBQuery{
		ID: "ASQ-USER_MGT-03",
		Query: "INSERT INTO \"USER\" (USER_ID, OU_ID, TYPE, ATTRIBUTES, CREDENTIALS, DEPLOYMENT_ID) " +
			"VALUES ($1, $2, $3, $4, $5, $6)",
	}
	// QueryGetUserByUserID is the query to get a user by user ID.
	QueryGetUserByUserID = model.DBQuery{
		ID:    "ASQ-USER_MGT-04",
		Query: "SELECT USER_ID, OU_ID, TYPE, ATTRIBUTES FROM \"USER\" WHERE USER_ID = $1 AND DEPLOYMENT_ID = $2",
	}
	// QueryUpdateUserByUserID is the query to update a user by user ID.
	QueryUpdateUserByUserID = model.DBQuery{
		ID:    "ASQ-USER_MGT-05",
		Query: "UPDATE \"USER\" SET OU_ID = $2, TYPE = $3, ATTRIBUTES = $4 WHERE USER_ID = $1 AND DEPLOYMENT_ID = $5;",
	}
	// QueryUpdateUserCredentialsByUserID is the query to update user credentials by user ID.
	QueryUpdateUserCredentialsByUserID = model.DBQuery{
		ID:    "ASQ-USER_MGT-14",
		Query: "UPDATE \"USER\" SET CREDENTIALS = $2 WHERE USER_ID = $1 AND DEPLOYMENT_ID = $3",
	}
	// QueryDeleteUserByUserID is the query to delete a user by user ID.
	QueryDeleteUserByUserID = model.DBQuery{
		ID:    "ASQ-USER_MGT-06",
		Query: "DELETE FROM \"USER\" WHERE USER_ID = $1 AND DEPLOYMENT_ID = $2",
	}
	// QueryValidateUserWithCredentials is the query to validate the user with the give credentials.
	QueryValidateUserWithCredentials = model.DBQuery{
		ID: "ASQ-USER_MGT-07",
		Query: "SELECT USER_ID, OU_ID, TYPE, ATTRIBUTES, CREDENTIALS FROM \"USER\" " +
			"WHERE USER_ID = $1 AND DEPLOYMENT_ID = $2",
	}
	// QueryGetGroupCountForUser is the query to get the count of groups for a given user.
	QueryGetGroupCountForUser = model.DBQuery{
		ID: "ASQ-USER_MGT-12",
		Query: `SELECT COUNT(*) AS total FROM GROUP_MEMBER_REFERENCE ` +
			`WHERE MEMBER_ID = $1 AND MEMBER_TYPE = 'user' AND DEPLOYMENT_ID = $2`,
	}
	// QueryGetGroupsForUser is the query to get groups for a given user with pagination.
	QueryGetGroupsForUser = model.DBQuery{
		ID: "ASQ-USER_MGT-13",
		Query: `SELECT G.GROUP_ID, G.OU_ID, G.NAME FROM GROUP_MEMBER_REFERENCE GMR ` +
			`INNER JOIN "GROUP" G ON GMR.GROUP_ID = G.GROUP_ID AND GMR.DEPLOYMENT_ID = $4 AND G.DEPLOYMENT_ID = $4 ` +
			`WHERE GMR.MEMBER_ID = $1 AND GMR.MEMBER_TYPE = 'user' AND GMR.DEPLOYMENT_ID = $4 ` +
			`ORDER BY G.NAME LIMIT $2 OFFSET $3`,
	}
	// QueryBatchInsertIndexedAttributes is the base query for batch inserting indexed attributes.
	// The complete query is built dynamically by appending VALUES placeholders based on the number of attributes.
	QueryBatchInsertIndexedAttributes = model.DBQuery{
		ID: "ASQ-USER_MGT-18",
		Query: "INSERT INTO USER_INDEXED_ATTRIBUTES " +
			"(USER_ID, ATTRIBUTE_NAME, ATTRIBUTE_VALUE, DEPLOYMENT_ID) VALUES ",
	}
	// QueryDeleteIndexedAttributesByUser is the query to delete all indexed attributes for a user.
	QueryDeleteIndexedAttributesByUser = model.DBQuery{
		ID:    "ASQ-USER_MGT-15",
		Query: "DELETE FROM USER_INDEXED_ATTRIBUTES WHERE USER_ID = $1 AND DEPLOYMENT_ID = $2",
	}
)

// buildIdentifyQuery constructs a query to identify a user based on the provided filters.
func buildIdentifyQuery(filters map[string]interface{}, deploymentID string) (model.DBQuery, []interface{}, error) {
	baseQuery := "SELECT USER_ID FROM \"USER\" WHERE 1=1"
	queryID := "ASQ-USER_MGT-08"
	columnName := AttributesColumn
	filterQuery, args, err := utils.BuildFilterQuery(queryID, baseQuery, columnName, filters)
	if err != nil {
		return model.DBQuery{}, nil, err
	}
	filterQuery, args = utils.AppendDeploymentIDToFilterQuery(filterQuery, args, deploymentID)
	return filterQuery, args, nil
}

// buildBulkUserExistsQuery constructs a query to check which user IDs exist from a list.
func buildBulkUserExistsQuery(userIDs []string, deploymentID string) (model.DBQuery, []interface{}, error) {
	if len(userIDs) == 0 {
		return model.DBQuery{}, nil, fmt.Errorf("userIDs list cannot be empty")
	}
	// Build placeholders for IN clause
	args := make([]interface{}, len(userIDs)+1)
	args[0] = deploymentID // DEPLOYMENT_ID will be filled by caller

	postgresPlaceholders := make([]string, len(userIDs))
	sqlitePlaceholders := make([]string, len(userIDs))

	for i, userID := range userIDs {
		postgresPlaceholders[i] = fmt.Sprintf("$%d", i+2)
		sqlitePlaceholders[i] = "?"
		args[i+1] = userID
	}

	baseQuery := "SELECT USER_ID FROM \"USER\" WHERE DEPLOYMENT_ID = %s AND USER_ID IN (%s)"
	postgresQuery := fmt.Sprintf(baseQuery, "$1", strings.Join(postgresPlaceholders, ","))
	sqliteQuery := fmt.Sprintf(baseQuery, "?", strings.Join(sqlitePlaceholders, ","))

	query := model.DBQuery{
		ID:            "ASQ-USER_MGT-09",
		Query:         postgresQuery,
		PostgresQuery: postgresQuery,
		SQLiteQuery:   sqliteQuery,
	}

	return query, args, nil
}

// buildUserListQuery constructs a query to get users with optional filtering.
func buildUserListQuery(
	filters map[string]interface{}, limit, offset int, deploymentID string,
) (model.DBQuery, []interface{}, error) {
	baseQuery := "SELECT USER_ID, OU_ID, TYPE, ATTRIBUTES FROM \"USER\""
	queryID := "ASQ-USER_MGT-10"
	columnName := AttributesColumn

	// Build the filter condition if filters are provided
	if len(filters) > 0 {
		filterQuery, filterArgs, err := utils.BuildFilterQuery(queryID, baseQuery+" WHERE 1=1", columnName, filters)
		if err != nil {
			return model.DBQuery{}, nil, err
		}
		filterQuery, filterArgs = utils.AppendDeploymentIDToFilterQuery(filterQuery, filterArgs, deploymentID)

		// Build PostgreSQL query
		postgresQuery, err := buildPaginatedQuery(filterQuery.PostgresQuery, len(filterArgs), "$")
		if err != nil {
			return model.DBQuery{}, nil, err
		}

		// Build SQLite query
		sqliteQuery, err := buildPaginatedQuery(filterQuery.SQLiteQuery, len(filterArgs), "?")
		if err != nil {
			return model.DBQuery{}, nil, err
		}

		filterArgs = append(filterArgs, limit, offset)
		return model.DBQuery{
			ID:            queryID,
			Query:         postgresQuery,
			PostgresQuery: postgresQuery,
			SQLiteQuery:   sqliteQuery,
		}, filterArgs, nil
	}

	// No filters, use the original query
	return QueryGetUserList, []interface{}{limit, offset, deploymentID}, nil
}

// buildPaginatedQuery constructs a paginated query string with ORDER BY, LIMIT, and OFFSET clauses.
func buildPaginatedQuery(baseQuery string, paramCount int, placeholder string) (string, error) {
	switch placeholder {
	case "?":
		return fmt.Sprintf("%s ORDER BY USER_ID LIMIT %s OFFSET %s",
			baseQuery, placeholder, placeholder), nil
	case "$":
		limitPlaceholder := fmt.Sprintf("%s%d", placeholder, paramCount+1)
		offsetPlaceholder := fmt.Sprintf("%s%d", placeholder, paramCount+2)
		return fmt.Sprintf("%s ORDER BY USER_ID LIMIT %s OFFSET %s",
			baseQuery, limitPlaceholder, offsetPlaceholder), nil
	}
	return "", fmt.Errorf("unsupported placeholder: %s", placeholder)
}

// buildUserCountQuery constructs a query to count users with optional filtering.
func buildUserCountQuery(filters map[string]interface{}, deploymentID string) (model.DBQuery, []interface{}, error) {
	baseQuery := "SELECT COUNT(*) as total FROM \"USER\""
	queryID := "ASQ-USER_MGT-11"
	columnName := AttributesColumn

	if len(filters) > 0 {
		filterQuery, args, err := utils.BuildFilterQuery(queryID, baseQuery+" WHERE 1=1", columnName, filters)
		if err != nil {
			return model.DBQuery{}, nil, err
		}
		filterQuery, args = utils.AppendDeploymentIDToFilterQuery(filterQuery, args, deploymentID)
		return filterQuery, args, nil
	}

	return QueryGetUserCount, []interface{}{deploymentID}, nil
}

// buildIdentifyQueryFromIndexedAttributes constructs a query to identify
// a user using only indexed attributes.
func buildIdentifyQueryFromIndexedAttributes(
	filters map[string]interface{}, deploymentID string) (model.DBQuery, []interface{}, error) {
	if len(filters) == 0 {
		return model.DBQuery{}, nil, fmt.Errorf("filters cannot be empty")
	}

	// Sort keys for deterministic query generation
	keys := make([]string, 0, len(filters))
	for key := range filters {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Build query using self-joins for multiple attributes
	baseQuery := `SELECT DISTINCT ia1.USER_ID FROM USER_INDEXED_ATTRIBUTES ia1`
	whereConditions := []string{}
	args := []interface{}{}
	paramIndex := 1

	// First attribute
	whereConditions = append(whereConditions, fmt.Sprintf("ia1.ATTRIBUTE_NAME = $%d AND ia1.ATTRIBUTE_VALUE = $%d",
		paramIndex, paramIndex+1))
	args = append(args, keys[0], fmt.Sprintf("%v", filters[keys[0]]))
	paramIndex += 2

	// Additional attributes via JOINs
	for i := 1; i < len(keys); i++ {
		alias := fmt.Sprintf("ia%d", i+1)
		baseQuery += fmt.Sprintf(
			" INNER JOIN USER_INDEXED_ATTRIBUTES %s ON ia1.USER_ID = %s.USER_ID "+
				"AND ia1.DEPLOYMENT_ID = %s.DEPLOYMENT_ID",
			alias, alias, alias)
		whereConditions = append(whereConditions, fmt.Sprintf("%s.ATTRIBUTE_NAME = $%d AND %s.ATTRIBUTE_VALUE = $%d",
			alias, paramIndex, alias, paramIndex+1))
		args = append(args, keys[i], fmt.Sprintf("%v", filters[keys[i]]))
		paramIndex += 2
	}

	queryString := baseQuery + " WHERE " + strings.Join(whereConditions, " AND ")

	// Add DEPLOYMENT_ID filter once at the end
	queryString += fmt.Sprintf(" AND ia1.DEPLOYMENT_ID = $%d", paramIndex)
	args = append(args, deploymentID)

	query := model.DBQuery{
		ID:    "ASQ-USER_MGT-16",
		Query: queryString,
	}

	return query, args, nil
}

// buildIdentifyQueryHybrid constructs a query using indexed attributes
// for initial filtering, then JSON attributes for remaining filters.
func buildIdentifyQueryHybrid(
	indexedFilters, nonIndexedFilters map[string]interface{},
	deploymentID string) (model.DBQuery, []interface{}, error) {
	if len(indexedFilters) == 0 {
		return model.DBQuery{}, nil, fmt.Errorf("indexed filters cannot be empty for hybrid query")
	}

	// Build query with JOINs on USER table (eliminates IN clause and potential limits)
	postgresQuery := `SELECT DISTINCT u.USER_ID FROM "USER" u`
	postgresQuery += ` INNER JOIN USER_INDEXED_ATTRIBUTES ia1 ON u.USER_ID = ia1.USER_ID ` +
		`AND u.DEPLOYMENT_ID = ia1.DEPLOYMENT_ID`

	sqliteQuery := `SELECT DISTINCT u.USER_ID FROM "USER" u`
	sqliteQuery += ` INNER JOIN USER_INDEXED_ATTRIBUTES ia1 ON u.USER_ID = ia1.USER_ID ` +
		`AND u.DEPLOYMENT_ID = ia1.DEPLOYMENT_ID`

	// Sort indexed keys for deterministic query generation
	indexedKeys := make([]string, 0, len(indexedFilters))
	for key := range indexedFilters {
		indexedKeys = append(indexedKeys, key)
	}
	sort.Strings(indexedKeys)

	whereConditions := []string{}
	args := []interface{}{}
	paramIndex := 1

	// First indexed attribute (no need to check ia1.DEPLOYMENT_ID since JOIN already ensures
	// it matches u.DEPLOYMENT_ID)
	whereConditions = append(whereConditions, fmt.Sprintf("ia1.ATTRIBUTE_NAME = $%d AND ia1.ATTRIBUTE_VALUE = $%d",
		paramIndex, paramIndex+1))
	args = append(args, indexedKeys[0], fmt.Sprintf("%v", indexedFilters[indexedKeys[0]]))
	paramIndex += 2

	// Additional indexed attributes via JOINs
	for i := 1; i < len(indexedKeys); i++ {
		alias := fmt.Sprintf("ia%d", i+1)
		postgresQuery += fmt.Sprintf(
			" INNER JOIN USER_INDEXED_ATTRIBUTES %s ON u.USER_ID = %s.USER_ID "+
				"AND u.DEPLOYMENT_ID = %s.DEPLOYMENT_ID",
			alias, alias, alias)
		sqliteQuery += fmt.Sprintf(
			" INNER JOIN USER_INDEXED_ATTRIBUTES %s ON u.USER_ID = %s.USER_ID "+
				"AND u.DEPLOYMENT_ID = %s.DEPLOYMENT_ID",
			alias, alias, alias)
		whereConditions = append(whereConditions, fmt.Sprintf("%s.ATTRIBUTE_NAME = $%d AND %s.ATTRIBUTE_VALUE = $%d",
			alias, paramIndex, alias, paramIndex+1))
		args = append(args, indexedKeys[i], fmt.Sprintf("%v", indexedFilters[indexedKeys[i]]))
		paramIndex += 2
	}

	postgresQuery += " WHERE " + strings.Join(whereConditions, " AND ")
	sqliteQuery += " WHERE " + strings.Join(whereConditions, " AND ")

	// Now add non-indexed filters on the ATTRIBUTES JSON column
	// Build JSON filter conditions directly with proper parameter numbering
	nonIndexedKeys := make([]string, 0, len(nonIndexedFilters))
	for key := range nonIndexedFilters {
		if err := utils.ValidateKey(key); err != nil {
			return model.DBQuery{}, nil, fmt.Errorf("invalid non-indexed filter key: %w", err)
		}
		nonIndexedKeys = append(nonIndexedKeys, key)
	}
	sort.Strings(nonIndexedKeys)

	for _, key := range nonIndexedKeys {
		// Build PostgreSQL JSON condition
		postgresCondition := utils.BuildPostgresJSONCondition("u."+AttributesColumn, key, paramIndex)
		postgresQuery += postgresCondition

		// Build SQLite JSON condition
		sqliteCondition := utils.BuildSQLiteJSONCondition("u."+AttributesColumn, key)
		sqliteQuery += sqliteCondition

		args = append(args, nonIndexedFilters[key])
		paramIndex++
	}

	// Add DEPLOYMENT_ID filter
	postgresQuery += fmt.Sprintf(" AND u.DEPLOYMENT_ID = $%d", paramIndex)
	sqliteQuery += " AND u.DEPLOYMENT_ID = ?"
	args = append(args, deploymentID)

	query := model.DBQuery{
		ID:            "ASQ-USER_MGT-17",
		Query:         postgresQuery,
		PostgresQuery: postgresQuery,
		SQLiteQuery:   sqliteQuery,
	}

	return query, args, nil
}
