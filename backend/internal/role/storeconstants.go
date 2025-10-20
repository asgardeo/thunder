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
	dbmodel "github.com/asgardeo/thunder/internal/system/database/model"
)

var (
	// queryCreateRole creates a new role.
	queryCreateRole = dbmodel.DBQuery{
		ID:    "RLQ-ROLE_MGT-01",
		Query: `INSERT INTO "ROLE" (ROLE_ID, OU_ID, NAME, DESCRIPTION) VALUES ($1, $2, $3, $4)`,
	}

	// queryGetRoleByID retrieves a role by ID.
	queryGetRoleByID = dbmodel.DBQuery{
		ID:    "RLQ-ROLE_MGT-02",
		Query: `SELECT ROLE_ID, OU_ID, NAME, DESCRIPTION FROM "ROLE" WHERE ROLE_ID = $1`,
	}

	// queryGetRoleList retrieves a list of roles with pagination.
	queryGetRoleList = dbmodel.DBQuery{
		ID:    "RLQ-ROLE_MGT-03",
		Query: `SELECT ROLE_ID, OU_ID, NAME, DESCRIPTION FROM "ROLE" ORDER BY CREATED_AT DESC LIMIT $1 OFFSET $2`,
	}

	// queryGetRoleListCount retrieves the total count of roles.
	queryGetRoleListCount = dbmodel.DBQuery{
		ID:    "RLQ-ROLE_MGT-04",
		Query: `SELECT COUNT(*) as total FROM "ROLE"`,
	}

	// queryUpdateRole updates a role.
	queryUpdateRole = dbmodel.DBQuery{
		ID:    "RLQ-ROLE_MGT-05",
		Query: `UPDATE "ROLE" SET OU_ID = $1, NAME = $2, DESCRIPTION = $3 WHERE ROLE_ID = $4`,
	}

	// queryDeleteRole deletes a role.
	queryDeleteRole = dbmodel.DBQuery{
		ID:    "RLQ-ROLE_MGT-06",
		Query: `DELETE FROM "ROLE" WHERE ROLE_ID = $1`,
	}

	// queryCreateRolePermission creates a new role permission.
	queryCreateRolePermission = dbmodel.DBQuery{
		ID:    "RLQ-ROLE_MGT-09",
		Query: `INSERT INTO ROLE_PERMISSION (ROLE_ID, PERMISSION) VALUES ($1, $2)`,
	}

	// queryGetRolePermissions retrieves all permissions for a role.
	queryGetRolePermissions = dbmodel.DBQuery{
		ID:    "RLQ-ROLE_MGT-10",
		Query: `SELECT PERMISSION FROM ROLE_PERMISSION WHERE ROLE_ID = $1 ORDER BY CREATED_AT`,
	}

	// queryDeleteRolePermissions deletes all permissions for a role.
	queryDeleteRolePermissions = dbmodel.DBQuery{
		ID:    "RLQ-ROLE_MGT-11",
		Query: `DELETE FROM ROLE_PERMISSION WHERE ROLE_ID = $1`,
	}

	// queryCreateRoleAssignment creates a new role assignment.
	queryCreateRoleAssignment = dbmodel.DBQuery{
		ID:    "RLQ-ROLE_MGT-12",
		Query: `INSERT INTO ROLE_ASSIGNMENT (ROLE_ID, ASSIGNEE_TYPE, ASSIGNEE_ID) VALUES ($1, $2, $3)`,
	}

	// queryGetRoleAssignments retrieves all assignments for a role with pagination.
	queryGetRoleAssignments = dbmodel.DBQuery{
		ID: "RLQ-ROLE_MGT-13",
		Query: `SELECT ASSIGNEE_ID, ASSIGNEE_TYPE FROM ROLE_ASSIGNMENT
			WHERE ROLE_ID = $1 ORDER BY CREATED_AT LIMIT $2 OFFSET $3`,
	}

	// queryGetRoleAssignmentsCount retrieves the total count of assignments for a role.
	queryGetRoleAssignmentsCount = dbmodel.DBQuery{
		ID:    "RLQ-ROLE_MGT-14",
		Query: `SELECT COUNT(*) as total FROM ROLE_ASSIGNMENT WHERE ROLE_ID = $1`,
	}

	// queryDeleteRoleAssignments deletes all assignments for a role.
	queryDeleteRoleAssignments = dbmodel.DBQuery{
		ID:    "RLQ-ROLE_MGT-15",
		Query: `DELETE FROM ROLE_ASSIGNMENT WHERE ROLE_ID = $1`,
	}

	// queryDeleteRoleAssignmentsByIDs deletes specific assignments for a role.
	queryDeleteRoleAssignmentsByIDs = dbmodel.DBQuery{
		ID:    "RLQ-ROLE_MGT-16",
		Query: `DELETE FROM ROLE_ASSIGNMENT WHERE ROLE_ID = $1 AND ASSIGNEE_TYPE = $2 AND ASSIGNEE_ID = $3`,
	}

	// queryGetAllRoleAssignments retrieves all assignments for a role without pagination.
	queryGetAllRoleAssignments = dbmodel.DBQuery{
		ID:    "RLQ-ROLE_MGT-17",
		Query: `SELECT ASSIGNEE_ID, ASSIGNEE_TYPE FROM ROLE_ASSIGNMENT WHERE ROLE_ID = $1 ORDER BY CREATED_AT`,
	}
)
