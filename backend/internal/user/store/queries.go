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
	"fmt"
	"strings"

	"github.com/asgardeo/thunder/internal/system/database/model"
	"github.com/asgardeo/thunder/internal/system/database/utils"
)

const (
	// AttributesColumn represents the ATTRIBUTES column name in the database.
	AttributesColumn = "ATTRIBUTES"
)

var (
	// QueryGetUserCount is the query to get total count of users.
	QueryGetUserCount = model.DBQuery{
		ID:    "ASQ-USER_MGT-01",
		Query: "SELECT COUNT(*) as total FROM \"USER\"",
	}
	// QueryGetUserList is the query to get a list of users.
	QueryGetUserList = model.DBQuery{
		ID:    "ASQ-USER_MGT-02",
		Query: "SELECT USER_ID, OU_ID, TYPE, ATTRIBUTES FROM \"USER\" ORDER BY USER_ID LIMIT $1 OFFSET $2",
	}
	// QueryCreateUser is the query to create a new user.
	QueryCreateUser = model.DBQuery{
		ID:    "ASQ-USER_MGT-03",
		Query: "INSERT INTO \"USER\" (USER_ID, OU_ID, TYPE, ATTRIBUTES, CREDENTIALS) VALUES ($1, $2, $3, $4, $5)",
	}
	// QueryGetUserByUserID is the query to get a user by user ID.
	QueryGetUserByUserID = model.DBQuery{
		ID:    "ASQ-USER_MGT-04",
		Query: "SELECT USER_ID, OU_ID, TYPE, ATTRIBUTES FROM \"USER\" WHERE USER_ID = $1",
	}
	// QueryUpdateUserByUserID is the query to update a user by user ID.
	QueryUpdateUserByUserID = model.DBQuery{
		ID:    "ASQ-USER_MGT-05",
		Query: "UPDATE \"USER\" SET OU_ID = $2, TYPE = $3, ATTRIBUTES = $4 WHERE USER_ID = $1;",
	}
	// QueryDeleteUserByUserID is the query to delete a user by user ID.
	QueryDeleteUserByUserID = model.DBQuery{
		ID:    "ASQ-USER_MGT-06",
		Query: "DELETE FROM \"USER\" WHERE USER_ID = $1",
	}
	// QueryValidateUserWithCredentials is the query to validate the user with the give credentials.
	QueryValidateUserWithCredentials = model.DBQuery{
		ID:    "ASQ-USER_MGT-07",
		Query: "SELECT USER_ID, OU_ID, TYPE, ATTRIBUTES, CREDENTIALS FROM \"USER\" WHERE USER_ID = $1",
	}
	// QueryGetGroupCountForUser is the query to get the count of groups for a given user.
	QueryGetGroupCountForUser = model.DBQuery{
		ID:    "ASQ-USER_MGT-12",
		Query: `SELECT COUNT(*) AS total FROM GROUP_MEMBER_REFERENCE WHERE MEMBER_ID = $1 AND MEMBER_TYPE = 'USER'`,
	}
	// QueryGetGroupsForUser is the query to get groups for a given user with pagination.
	QueryGetGroupsForUser = model.DBQuery{
		ID: "ASQ-USER_MGT-13",
		Query: `SELECT G.GROUP_ID, G.OU_ID, G.NAME FROM GROUP_MEMBER_REFERENCE GMR ` +
			`INNER JOIN "GROUP" G ON GMR.GROUP_ID = G.GROUP_ID ` +
			`WHERE GMR.MEMBER_ID = $1 AND GMR.MEMBER_TYPE = 'USER' ` +
			`ORDER BY G.NAME LIMIT $2 OFFSET $3`,
	}
)

// buildIdentifyQuery constructs a query to identify a user based on the provided filters.
func buildIdentifyQuery(filters map[string]interface{}) (model.DBQuery, []interface{}, error) {
	baseQuery := "SELECT USER_ID FROM \"USER\" WHERE 1=1"
	queryID := "ASQ-USER_MGT-08"
	columnName := AttributesColumn
	return utils.BuildFilterQuery(queryID, baseQuery, columnName, filters)
}

// buildBulkUserExistsQuery constructs a query to check which user IDs exist from a list.
func buildBulkUserExistsQuery(userIDs []string) (model.DBQuery, []interface{}, error) {
	if len(userIDs) == 0 {
		return model.DBQuery{}, nil, fmt.Errorf("userIDs list cannot be empty")
	}
	// Build placeholders for IN clause
	args := make([]interface{}, len(userIDs))

	postgresPlaceholders := make([]string, len(userIDs))
	sqlitePlaceholders := make([]string, len(userIDs))

	for i, userID := range userIDs {
		postgresPlaceholders[i] = fmt.Sprintf("$%d", i+1)
		sqlitePlaceholders[i] = "?"
		args[i] = userID
	}

	baseQuery := "SELECT USER_ID FROM \"USER\" WHERE USER_ID IN (%s)"
	postgresQuery := fmt.Sprintf(baseQuery, strings.Join(postgresPlaceholders, ","))
	sqliteQuery := fmt.Sprintf(baseQuery, strings.Join(sqlitePlaceholders, ","))

	query := model.DBQuery{
		ID:            "ASQ-USER_MGT-09",
		Query:         postgresQuery,
		PostgresQuery: postgresQuery,
		SQLiteQuery:   sqliteQuery,
	}

	return query, args, nil
}

// buildUserListQuery constructs a query to get users with optional filtering.
func buildUserListQuery(filters map[string]interface{}, limit, offset int) (model.DBQuery, []interface{}, error) {
	baseQuery := "SELECT USER_ID, OU_ID, TYPE, ATTRIBUTES FROM \"USER\""
	queryID := "ASQ-USER_MGT-10"
	columnName := AttributesColumn

	// Build the filter condition if filters are provided
	if len(filters) > 0 {
		filterQuery, filterArgs, err := utils.BuildFilterQuery(queryID, baseQuery+" WHERE 1=1", columnName, filters)
		if err != nil {
			return model.DBQuery{}, nil, err
		}

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
	return QueryGetUserList, []interface{}{limit, offset}, nil
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
func buildUserCountQuery(filters map[string]interface{}) (model.DBQuery, []interface{}, error) {
	baseQuery := "SELECT COUNT(*) as total FROM \"USER\""
	queryID := "ASQ-USER_MGT-11"
	columnName := AttributesColumn

	if len(filters) > 0 {
		return utils.BuildFilterQuery(queryID, baseQuery+" WHERE 1=1", columnName, filters)
	}

	return QueryGetUserCount, []interface{}{}, nil
}
