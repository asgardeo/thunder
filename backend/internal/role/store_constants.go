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

package role

import (
	"fmt"
	"strings"

	dbmodel "github.com/asgardeo/thunder/internal/system/database/model"
)

var (
	// The table name "ROLE" is quoted to handle reserved keywords in SQL.
	// Hence, all queries involving the "ROLE" table use quoted identifiers.
	// queryCreateRole creates a new role.
	queryCreateRole = dbmodel.DBQuery{
		ID: "RLQ-ROLE_MGT-01",
		Query: `INSERT INTO "ROLE" (ROLE_ID, OU_ID, NAME, DESCRIPTION, DEPLOYMENT_ID) 
			VALUES ($1, $2, $3, $4, $5) RETURNING ID`,
	}

	// queryGetRoleByID retrieves a role by ID.
	queryGetRoleByID = dbmodel.DBQuery{
		ID:    "RLQ-ROLE_MGT-02",
		Query: `SELECT ID, ROLE_ID, OU_ID, NAME, DESCRIPTION FROM "ROLE" WHERE ROLE_ID = $1 AND DEPLOYMENT_ID = $2`,
	}

	// queryGetRoleList retrieves a list of roles with pagination.
	queryGetRoleList = dbmodel.DBQuery{
		ID: "RLQ-ROLE_MGT-03",
		Query: `SELECT ROLE_ID, OU_ID, NAME, DESCRIPTION FROM "ROLE" ` +
			`WHERE DEPLOYMENT_ID = $3 ORDER BY CREATED_AT DESC LIMIT $1 OFFSET $2`,
	}

	// queryGetRoleListCount retrieves the total count of roles.
	queryGetRoleListCount = dbmodel.DBQuery{
		ID:    "RLQ-ROLE_MGT-04",
		Query: `SELECT COUNT(*) as total FROM "ROLE" WHERE DEPLOYMENT_ID = $1`,
	}

	// queryUpdateRole updates a role.
	queryUpdateRole = dbmodel.DBQuery{
		ID:    "RLQ-ROLE_MGT-05",
		Query: `UPDATE "ROLE" SET OU_ID = $1, NAME = $2, DESCRIPTION = $3 WHERE ID = $4 AND DEPLOYMENT_ID = $5`,
	}

	// queryDeleteRole deletes a role.
	queryDeleteRole = dbmodel.DBQuery{
		ID:    "RLQ-ROLE_MGT-06",
		Query: `DELETE FROM "ROLE" WHERE ID = $1 AND DEPLOYMENT_ID = $2`,
	}

	// queryCreateRolePermission creates a new role permission.
	queryCreateRolePermission = dbmodel.DBQuery{
		ID: "RLQ-ROLE_MGT-07",
		Query: `INSERT INTO ROLE_PERMISSION (ROLE_INTERNAL_ID, RESOURCE_SERVER_ID, PERMISSION, ` +
			`DEPLOYMENT_ID) VALUES ($1, $2, $3, $4)`,
	}

	// queryGetRolePermissions retrieves all permissions for a role.
	queryGetRolePermissions = dbmodel.DBQuery{
		ID: "RLQ-ROLE_MGT-08",
		Query: `SELECT RESOURCE_SERVER_ID, PERMISSION FROM ROLE_PERMISSION WHERE ` +
			`ROLE_INTERNAL_ID = $1 AND DEPLOYMENT_ID = $2 ORDER BY CREATED_AT`,
	}

	// queryDeleteRolePermissions deletes all permissions for a role.
	queryDeleteRolePermissions = dbmodel.DBQuery{
		ID:    "RLQ-ROLE_MGT-09",
		Query: `DELETE FROM ROLE_PERMISSION WHERE ROLE_INTERNAL_ID = $1 AND DEPLOYMENT_ID = $2`,
	}

	// queryCreateRoleAssignment creates a new role assignment.
	queryCreateRoleAssignment = dbmodel.DBQuery{
		ID: "RLQ-ROLE_MGT-10",
		Query: `INSERT INTO ROLE_ASSIGNMENT (ROLE_INTERNAL_ID, ASSIGNEE_TYPE, ASSIGNEE_ID, DEPLOYMENT_ID) 
			VALUES ($1, $2, $3, $4)`,
	}

	// queryGetRoleAssignments retrieves all assignments for a role with pagination.
	queryGetRoleAssignments = dbmodel.DBQuery{
		ID: "RLQ-ROLE_MGT-11",
		Query: `SELECT ASSIGNEE_ID, ASSIGNEE_TYPE FROM ROLE_ASSIGNMENT
			WHERE ROLE_INTERNAL_ID = $1 AND DEPLOYMENT_ID = $4 ORDER BY CREATED_AT LIMIT $2 OFFSET $3`,
	}

	// queryGetRoleAssignmentsCount retrieves the total count of assignments for a role.
	queryGetRoleAssignmentsCount = dbmodel.DBQuery{
		ID:    "RLQ-ROLE_MGT-12",
		Query: `SELECT COUNT(*) as total FROM ROLE_ASSIGNMENT WHERE ROLE_INTERNAL_ID = $1 AND DEPLOYMENT_ID = $2`,
	}

	// queryDeleteRoleAssignmentsByIDs deletes specific assignments for a role.
	queryDeleteRoleAssignmentsByIDs = dbmodel.DBQuery{
		ID: "RLQ-ROLE_MGT-13",
		Query: `DELETE FROM ROLE_ASSIGNMENT ` +
			`WHERE ROLE_INTERNAL_ID = $1 AND ASSIGNEE_TYPE = $2 AND ASSIGNEE_ID = $3 AND DEPLOYMENT_ID = $4`,
	}

	// queryCheckRoleNameExists checks if a role name already exists for a given organization unit.
	queryCheckRoleNameExists = dbmodel.DBQuery{
		ID:    "RLQ-ROLE_MGT-14",
		Query: `SELECT COUNT(*) as count FROM "ROLE" WHERE OU_ID = $1 AND NAME = $2 AND DEPLOYMENT_ID = $3`,
	}

	// queryCheckRoleNameExistsExcludingID checks if a role name exists for an OU excluding a specific role ID.
	queryCheckRoleNameExistsExcludingID = dbmodel.DBQuery{
		ID: "RLQ-ROLE_MGT-15",
		Query: `SELECT COUNT(*) as count FROM "ROLE" 
			WHERE OU_ID = $1 AND NAME = $2 AND ROLE_ID != $3 AND DEPLOYMENT_ID = $4`,
	}

	// queryCheckRoleExists checks if a role exists by its ID and returns the internal ID.
	queryCheckRoleExists = dbmodel.DBQuery{
		ID:    "RLQ-ROLE_MGT-16",
		Query: `SELECT ID FROM "ROLE" WHERE ROLE_ID = $1 AND DEPLOYMENT_ID = $2`,
	}
)

// buildAuthorizedPermissionsQuery constructs a database-specific query to retrieve authorized permissions
// for a user and/or groups from their assigned roles.
// It builds separate queries for PostgreSQL and SQLite to handle array parameters correctly.
func buildAuthorizedPermissionsQuery(
	userID string,
	groupIDs []string,
	requestedPermissions []string,
	deploymentID string,
) (dbmodel.DBQuery, []interface{}) {
	// Base query structure
	baseQuery := `SELECT DISTINCT rp.PERMISSION
		FROM ROLE_PERMISSION rp
		INNER JOIN ROLE_ASSIGNMENT ra ON rp.ROLE_INTERNAL_ID = ra.ROLE_INTERNAL_ID ` +
		`AND rp.DEPLOYMENT_ID = $1 AND ra.DEPLOYMENT_ID = $1
		WHERE rp.DEPLOYMENT_ID = $1 AND `

	var postgresWhere []string
	var sqliteWhere []string

	// Pre-allocate args slice with estimated capacity
	argsCapacity := 1 + len(groupIDs) + len(requestedPermissions) // +1 for DEPLOYMENT_ID
	if userID != "" {
		argsCapacity++
	}
	args := make([]interface{}, 0, argsCapacity)
	args = append(args, deploymentID)
	paramIndex := 2 // Start from $2 since $1 is DEPLOYMENT_ID

	// Build user condition if userID is provided
	if userID != "" {
		postgresWhere = append(postgresWhere,
			fmt.Sprintf("(ra.ASSIGNEE_TYPE = 'user' AND ra.ASSIGNEE_ID = $%d)", paramIndex))
		sqliteWhere = append(sqliteWhere,
			"(ra.ASSIGNEE_TYPE = 'user' AND ra.ASSIGNEE_ID = ?)")
		args = append(args, userID)
		paramIndex++
	}

	// Build group condition if groupIDs are provided
	if len(groupIDs) > 0 {
		groupPlaceholdersPostgres := make([]string, len(groupIDs))
		groupPlaceholdersSqlite := make([]string, len(groupIDs))

		for i, groupID := range groupIDs {
			groupPlaceholdersPostgres[i] = fmt.Sprintf("$%d", paramIndex+i)
			groupPlaceholdersSqlite[i] = "?"
			args = append(args, groupID)
		}

		postgresWhere = append(postgresWhere,
			fmt.Sprintf("(ra.ASSIGNEE_TYPE = 'group' AND ra.ASSIGNEE_ID IN (%s))",
				strings.Join(groupPlaceholdersPostgres, ",")))
		sqliteWhere = append(sqliteWhere,
			fmt.Sprintf("(ra.ASSIGNEE_TYPE = 'group' AND ra.ASSIGNEE_ID IN (%s))",
				strings.Join(groupPlaceholdersSqlite, ",")))
		paramIndex += len(groupIDs)
	}

	// Build permission condition
	permPlaceholdersPostgres := make([]string, len(requestedPermissions))
	permPlaceholdersSqlite := make([]string, len(requestedPermissions))

	for i, perm := range requestedPermissions {
		permPlaceholdersPostgres[i] = fmt.Sprintf("$%d", paramIndex+i)
		permPlaceholdersSqlite[i] = "?"
		args = append(args, perm)
	}

	// Construct PostgreSQL query
	postgresQuery := baseQuery +
		"(" + strings.Join(postgresWhere, " OR ") + ") AND " +
		fmt.Sprintf("rp.PERMISSION IN (%s)", strings.Join(permPlaceholdersPostgres, ",")) +
		" ORDER BY rp.PERMISSION"

	// Construct SQLite query
	sqliteQuery := baseQuery +
		"(" + strings.Join(sqliteWhere, " OR ") + ") AND " +
		fmt.Sprintf("rp.PERMISSION IN (%s)", strings.Join(permPlaceholdersSqlite, ",")) +
		" ORDER BY rp.PERMISSION"

	query := dbmodel.DBQuery{
		ID:            "RLQ-ROLE_MGT-20",
		Query:         postgresQuery,
		PostgresQuery: postgresQuery,
		SQLiteQuery:   sqliteQuery,
	}

	return query, args
}
