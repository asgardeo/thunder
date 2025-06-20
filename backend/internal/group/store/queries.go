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

// Package store provides the implementation for group persistence operations.
package store

import (
	dbmodel "github.com/asgardeo/thunder/internal/system/database/model"
)

var (
	// QueryGetGroupList is the query to get all root groups.
	QueryGetGroupList = dbmodel.DBQuery{
		ID:    "GRQ-GROUP_MGT-01",
		Query: `SELECT GROUP_ID, OU_ID, NAME, DESCRIPTION, PATH FROM "GROUP" WHERE OU_ID IS NOT NULL AND PARENT_ID IS NULL`,
	}

	// QueryCreateGroup is the query to create a new group.
	QueryCreateGroup = dbmodel.DBQuery{
		ID:    "GRQ-GROUP_MGT-02",
		Query: `INSERT INTO "GROUP" (GROUP_ID, PARENT_ID, OU_ID, NAME, DESCRIPTION, PATH) VALUES ($1, $2, $3, $4, $5, $6)`,
	}

	// QueryGetGroupByID is the query to get a group by id.
	QueryGetGroupByID = dbmodel.DBQuery{
		ID:    "GRQ-GROUP_MGT-03",
		Query: `SELECT GROUP_ID, PARENT_ID, OU_ID, NAME, DESCRIPTION, PATH FROM "GROUP" WHERE GROUP_ID = $1`,
	}

	// QueryUpdateGroup is the query to update a group.
	QueryUpdateGroup = dbmodel.DBQuery{
		ID:    "GRQ-GROUP_MGT-04",
		Query: `UPDATE "GROUP" SET PARENT_ID = $2, OU_ID = $3, NAME = $4, DESCRIPTION = $5, PATH = $6 WHERE GROUP_ID = $1`,
	}

	// QueryDeleteGroup is the query to delete a group.
	QueryDeleteGroup = dbmodel.DBQuery{
		ID:    "GRQ-GROUP_MGT-05",
		Query: `DELETE FROM "GROUP" WHERE GROUP_ID = $1`,
	}

	// QueryGetChildGroups is the query to get child groups of a group.
	QueryGetChildGroups = dbmodel.DBQuery{
		ID:    "GRQ-GROUP_MGT-06",
		Query: `SELECT GROUP_ID FROM "GROUP" WHERE PARENT_ID = $1`,
	}

	// QueryGetGroupUsers is the query to get users in a group.
	QueryGetGroupUsers = dbmodel.DBQuery{
		ID:    "GRQ-GROUP_MGT-07",
		Query: `SELECT USER_ID FROM GROUP_USER_REFERENCE WHERE GROUP_ID = $1`,
	}

	// QueryDeleteGroupUsers is the query to delete all users from a group.
	QueryDeleteGroupUsers = dbmodel.DBQuery{
		ID:    "GRQ-GROUP_MGT-08",
		Query: `DELETE FROM GROUP_USER_REFERENCE WHERE GROUP_ID = $1`,
	}

	// QueryAddUserToGroup is the query to add a user to a group.
	QueryAddUserToGroup = dbmodel.DBQuery{
		ID:    "GRQ-GROUP_MGT-09",
		Query: `INSERT INTO GROUP_USER_REFERENCE (GROUP_ID, USER_ID) VALUES ($1, $2)`,
	}

	// QueryCheckGroupNameConflict is the query to check if a group name conflicts under the same parent.
	QueryCheckGroupNameConflict = dbmodel.DBQuery{
		ID:    "GRQ-GROUP_MGT-10",
		Query: `SELECT COUNT(*) as count FROM "GROUP" WHERE NAME = $1 AND PARENT_ID = $2`,
	}

	// QueryCheckGroupNameConflictForUpdate is the query to check name conflict during update.
	QueryCheckGroupNameConflictForUpdate = dbmodel.DBQuery{
		ID:    "GRQ-GROUP_MGT-11",
		Query: `SELECT COUNT(*) as count FROM "GROUP" WHERE NAME = $1 AND PARENT_ID = $2 AND GROUP_ID != $3`,
	}

	// QueryCheckGroupNameConflictInOU is the query to check if a group name conflicts within an organization unit.
	QueryCheckGroupNameConflictInOU = dbmodel.DBQuery{
		ID:    "GRQ-GROUP_MGT-12",
		Query: `SELECT COUNT(*) as count FROM "GROUP" WHERE NAME = $1 AND OU_ID = $2 AND PARENT_ID IS NULL`,
	}

	// QueryCheckGroupNameConflictInOUForUpdate is the query to check name conflict in OU during update.
	QueryCheckGroupNameConflictInOUForUpdate = dbmodel.DBQuery{
		ID: "GRQ-GROUP_MGT-13",
		Query: `SELECT COUNT(*) as count FROM "GROUP" WHERE NAME = $1 AND OU_ID = $2 AND ` +
			`PARENT_ID IS NULL AND GROUP_ID != $3`,
	}
)
