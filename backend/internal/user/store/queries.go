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

// Package store provides the implementation for user persistence operations.
package store

import (
	"fmt"
	"strings"

	"github.com/asgardeo/thunder/internal/system/database/model"
	"github.com/asgardeo/thunder/internal/system/database/utils"
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
)

// buildIdentifyQuery constructs a query to identify a user based on the provided filters.
func buildIdentifyQuery(filters map[string]interface{}) (model.DBQuery, []interface{}, error) {
	baseQuery := "SELECT USER_ID FROM \"USER\" WHERE 1=1"
	queryID := "ASQ-USER_MGT-08"
	columnName := "ATTRIBUTES"
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
