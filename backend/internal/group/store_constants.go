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

package group

import (
	"fmt"
	"strings"

	dbmodel "github.com/asgardeo/thunder/internal/system/database/model"
)

var (
	// QueryGetGroupListCount is the query to get total count of groups.
	QueryGetGroupListCount = dbmodel.DBQuery{
		ID:    "GRQ-GROUP_MGT-01",
		Query: `SELECT COUNT(*) as total FROM "GROUP" WHERE DEPLOYMENT_ID = $1`,
	}

	// QueryGetGroupList is the query to get groups with pagination.
	QueryGetGroupList = dbmodel.DBQuery{
		ID: "GRQ-GROUP_MGT-02",
		Query: `SELECT GROUP_ID, OU_ID, NAME, DESCRIPTION FROM "GROUP" ` +
			`WHERE DEPLOYMENT_ID = $3 ORDER BY NAME LIMIT $1 OFFSET $2`,
	}

	// QueryCreateGroup is the query to create a new group.
	QueryCreateGroup = dbmodel.DBQuery{
		ID:    "GRQ-GROUP_MGT-03",
		Query: `INSERT INTO "GROUP" (GROUP_ID, OU_ID, NAME, DESCRIPTION, DEPLOYMENT_ID) VALUES ($1, $2, $3, $4, $5)`,
	}

	// QueryGetGroupByID is the query to get a group by id.
	QueryGetGroupByID = dbmodel.DBQuery{
		ID:    "GRQ-GROUP_MGT-04",
		Query: `SELECT GROUP_ID, OU_ID, NAME, DESCRIPTION FROM "GROUP" WHERE GROUP_ID = $1 AND DEPLOYMENT_ID = $2`,
	}

	// QueryGetGroupMembers is the query to get members assigned to a group.
	QueryGetGroupMembers = dbmodel.DBQuery{
		ID: "GRQ-GROUP_MGT-05",
		Query: `SELECT MEMBER_ID, MEMBER_TYPE FROM GROUP_MEMBER_REFERENCE WHERE GROUP_ID = $1 AND DEPLOYMENT_ID = $4 ` +
			`ORDER BY MEMBER_TYPE, MEMBER_ID LIMIT $2 OFFSET $3`,
	}

	// QueryGetGroupMemberCount is the query to get total count of members in a group.
	QueryGetGroupMemberCount = dbmodel.DBQuery{
		ID:    "GRQ-GROUP_MGT-06",
		Query: `SELECT COUNT(*) as total FROM GROUP_MEMBER_REFERENCE WHERE GROUP_ID = $1 AND DEPLOYMENT_ID = $2`,
	}

	// QueryUpdateGroup is the query to update a group.
	QueryUpdateGroup = dbmodel.DBQuery{
		ID:    "GRQ-GROUP_MGT-07",
		Query: `UPDATE "GROUP" SET OU_ID = $2, NAME = $3, DESCRIPTION = $4 WHERE GROUP_ID = $1 AND DEPLOYMENT_ID = $5`,
	}

	// QueryDeleteGroup is the query to delete a group.
	QueryDeleteGroup = dbmodel.DBQuery{
		ID:    "GRQ-GROUP_MGT-08",
		Query: `DELETE FROM "GROUP" WHERE GROUP_ID = $1 AND DEPLOYMENT_ID = $2`,
	}

	// QueryDeleteGroupMembers is the query to delete all members assigned to a group.
	QueryDeleteGroupMembers = dbmodel.DBQuery{
		ID:    "GRQ-GROUP_MGT-09",
		Query: `DELETE FROM GROUP_MEMBER_REFERENCE WHERE GROUP_ID = $1 AND DEPLOYMENT_ID = $2`,
	}

	// QueryAddMemberToGroup is the query to assign member to a group.
	QueryAddMemberToGroup = dbmodel.DBQuery{
		ID: "GRQ-GROUP_MGT-10",
		Query: `INSERT INTO GROUP_MEMBER_REFERENCE (GROUP_ID, MEMBER_TYPE, MEMBER_ID, DEPLOYMENT_ID) ` +
			`VALUES ($1, $2, $3, $4) ON CONFLICT (GROUP_ID, MEMBER_TYPE, MEMBER_ID, DEPLOYMENT_ID) DO NOTHING`,
	}

	// QueryCheckGroupNameConflict is the query to check if a group name conflicts within the same organization unit.
	QueryCheckGroupNameConflict = dbmodel.DBQuery{
		ID:    "GRQ-GROUP_MGT-11",
		Query: `SELECT COUNT(*) as count FROM "GROUP" WHERE NAME = $1 AND OU_ID = $2 AND DEPLOYMENT_ID = $3`,
	}

	// QueryCheckGroupNameConflictForUpdate is the query to check name conflict during update.
	QueryCheckGroupNameConflictForUpdate = dbmodel.DBQuery{
		ID: "GRQ-GROUP_MGT-12",
		Query: `SELECT COUNT(*) as count FROM "GROUP" ` +
			`WHERE NAME = $1 AND OU_ID = $2 AND GROUP_ID != $3 AND DEPLOYMENT_ID = $4`,
	}

	// QueryGetGroupsByOrganizationUnitCount is the query to get total count of groups by organization unit.
	QueryGetGroupsByOrganizationUnitCount = dbmodel.DBQuery{
		ID:    "GRQ-GROUP_MGT-13",
		Query: `SELECT COUNT(*) as total FROM "GROUP" WHERE OU_ID = $1 AND DEPLOYMENT_ID = $2`,
	}

	// QueryGetGroupsByOrganizationUnit is the query to get groups by organization unit with pagination.
	QueryGetGroupsByOrganizationUnit = dbmodel.DBQuery{
		ID: "GRQ-GROUP_MGT-14",
		Query: `SELECT GROUP_ID, OU_ID, NAME, DESCRIPTION FROM "GROUP" ` +
			`WHERE OU_ID = $1 AND DEPLOYMENT_ID = $4 ORDER BY NAME LIMIT $2 OFFSET $3`,
	}

	// QueryDeleteGroupMember is the query to delete a specific member from a group.
	QueryDeleteGroupMember = dbmodel.DBQuery{
		ID: "GRQ-GROUP_MGT-16",
		Query: `DELETE FROM GROUP_MEMBER_REFERENCE ` +
			`WHERE GROUP_ID = $1 AND MEMBER_TYPE = $2 AND MEMBER_ID = $3 AND DEPLOYMENT_ID = $4`,
	}
)

// buildBulkGroupExistsQuery constructs a query to check which group IDs exist from a list.
func buildBulkGroupExistsQuery(groupIDs []string, deploymentID string) (dbmodel.DBQuery, []interface{}, error) {
	if len(groupIDs) == 0 {
		return dbmodel.DBQuery{}, nil, fmt.Errorf("groupIDs list cannot be empty")
	}
	args := make([]interface{}, len(groupIDs)+1)
	args[0] = deploymentID

	postgresPlaceholders := make([]string, len(groupIDs))
	sqlitePlaceholders := make([]string, len(groupIDs))

	for i, groupID := range groupIDs {
		postgresPlaceholders[i] = fmt.Sprintf("$%d", i+2)
		sqlitePlaceholders[i] = "?"
		args[i+1] = groupID
	}

	baseQuery := "SELECT GROUP_ID FROM \"GROUP\" WHERE DEPLOYMENT_ID = %s AND GROUP_ID IN (%s)"
	postgresQuery := fmt.Sprintf(baseQuery, "$1", strings.Join(postgresPlaceholders, ","))
	sqliteQuery := fmt.Sprintf(baseQuery, "?", strings.Join(sqlitePlaceholders, ","))

	query := dbmodel.DBQuery{
		ID:            "GRQ-GROUP_MGT-15",
		Query:         postgresQuery,
		PostgresQuery: postgresQuery,
		SQLiteQuery:   sqliteQuery,
	}

	return query, args, nil
}
