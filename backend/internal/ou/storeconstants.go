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

package ou

import dbmodel "github.com/asgardeo/thunder/internal/system/database/model"

var (
	// queryGetRootOrganizationUnitListCount is the query to get total count of organization units.
	queryGetRootOrganizationUnitListCount = dbmodel.DBQuery{
		ID:    "OUQ-OU_MGT-01",
		Query: `SELECT COUNT(*) as total FROM ORGANIZATION_UNIT WHERE PARENT_ID IS NULL`,
	}

	// queryGetRootOrganizationUnitList is the query to get organization units with pagination.
	queryGetRootOrganizationUnitList = dbmodel.DBQuery{
		ID: "OUQ-OU_MGT-02",
		Query: `SELECT OU_ID, HANDLE, NAME, DESCRIPTION, PARENT_ID FROM ORGANIZATION_UNIT WHERE PARENT_ID IS NULL ` +
			`ORDER BY NAME LIMIT $1 OFFSET $2`,
	}

	// queryCreateOrganizationUnit is the query to create a new organization unit.
	queryCreateOrganizationUnit = dbmodel.DBQuery{
		ID:    "OUQ-OU_MGT-03",
		Query: `INSERT INTO ORGANIZATION_UNIT (OU_ID, PARENT_ID, HANDLE, NAME, DESCRIPTION) VALUES ($1, $2, $3, $4, $5)`,
	}

	// queryGetOrganizationUnitByID is the query to get an organization unit by id.
	queryGetOrganizationUnitByID = dbmodel.DBQuery{
		ID:    "OUQ-OU_MGT-04",
		Query: `SELECT OU_ID, PARENT_ID, HANDLE, NAME, DESCRIPTION FROM ORGANIZATION_UNIT WHERE OU_ID = $1`,
	}

	// queryGetRootOrganizationUnitByHandle is the query to get a root organization unit by handle.
	queryGetRootOrganizationUnitByHandle = dbmodel.DBQuery{
		ID: "OUQ-OU_MGT-05",
		Query: `SELECT OU_ID, PARENT_ID, HANDLE, NAME, DESCRIPTION FROM ORGANIZATION_UNIT ` +
			`WHERE HANDLE = $1 AND PARENT_ID IS NULL`,
	}

	// queryGetOrganizationUnitByHandle is the query to get an organization unit by handle and parent.
	queryGetOrganizationUnitByHandle = dbmodel.DBQuery{
		ID: "OUQ-OU_MGT-06",
		Query: `SELECT OU_ID, PARENT_ID, HANDLE, NAME, DESCRIPTION FROM ORGANIZATION_UNIT ` +
			`WHERE HANDLE = $1 AND PARENT_ID = $2`,
	}

	// queryCheckOrganizationUnitExists is the query to check if an organization unit exists.
	queryCheckOrganizationUnitExists = dbmodel.DBQuery{
		ID:    "OUQ-OU_MGT-07",
		Query: `SELECT COUNT(*) as count FROM ORGANIZATION_UNIT WHERE OU_ID = $1`,
	}

	// queryUpdateOrganizationUnit is the query to update an organization unit.
	queryUpdateOrganizationUnit = dbmodel.DBQuery{
		ID:    "OUQ-OU_MGT-08",
		Query: `UPDATE ORGANIZATION_UNIT SET PARENT_ID = $2, HANDLE = $3, NAME = $4, DESCRIPTION = $5 WHERE OU_ID = $1`,
	}

	// queryDeleteOrganizationUnit is the query to delete an organization unit.
	queryDeleteOrganizationUnit = dbmodel.DBQuery{
		ID:    "OUQ-OU_MGT-09",
		Query: `DELETE FROM ORGANIZATION_UNIT WHERE OU_ID = $1`,
	}

	// queryGetOrganizationUnitChildrenCount is the query to get total count of child organization units.
	queryGetOrganizationUnitChildrenCount = dbmodel.DBQuery{
		ID:    "OUQ-OU_MGT-10",
		Query: `SELECT COUNT(*) as total FROM ORGANIZATION_UNIT WHERE PARENT_ID = $1`,
	}

	// queryGetOrganizationUnitChildrenList is the query to get child organization units with pagination.
	queryGetOrganizationUnitChildrenList = dbmodel.DBQuery{
		ID: "OUQ-OU_MGT-11",
		Query: `SELECT OU_ID, HANDLE, NAME, DESCRIPTION FROM ORGANIZATION_UNIT WHERE PARENT_ID = $1 ` +
			`ORDER BY NAME LIMIT $2 OFFSET $3`,
	}

	// queryGetOrganizationUnitUsersCount is the query to get total count of users in an organization unit.
	queryGetOrganizationUnitUsersCount = dbmodel.DBQuery{
		ID:    "OUQ-OU_MGT-12",
		Query: `SELECT COUNT(*) as total FROM "USER" WHERE OU_ID = $1`,
	}

	// queryGetOrganizationUnitUsersList is the query to get users in an organization unit with pagination.
	queryGetOrganizationUnitUsersList = dbmodel.DBQuery{
		ID:    "OUQ-OU_MGT-13",
		Query: `SELECT USER_ID FROM "USER" WHERE OU_ID = $1 ORDER BY USER_ID LIMIT $2 OFFSET $3`,
	}

	// queryGetOrganizationUnitGroupsCount is the query to get total count of groups in an organization unit.
	queryGetOrganizationUnitGroupsCount = dbmodel.DBQuery{
		ID:    "OUQ-OU_MGT-14",
		Query: `SELECT COUNT(*) as total FROM "GROUP" WHERE OU_ID = $1`,
	}

	// queryGetOrganizationUnitGroupsList is the query to get groups in an organization unit with pagination.
	queryGetOrganizationUnitGroupsList = dbmodel.DBQuery{
		ID:    "OUQ-OU_MGT-15",
		Query: `SELECT GROUP_ID, NAME FROM "GROUP" WHERE OU_ID = $1 ORDER BY NAME LIMIT $2 OFFSET $3`,
	}

	// queryCheckOrganizationUnitNameConflict is the query to check if an organization
	// unit name conflicts under the same parent.
	queryCheckOrganizationUnitNameConflict = dbmodel.DBQuery{
		ID:    "OUQ-OU_MGT-16",
		Query: `SELECT COUNT(*) as count FROM ORGANIZATION_UNIT WHERE NAME = $1 AND PARENT_ID = $2`,
	}

	// queryCheckOrganizationUnitNameConflictRoot is the query to check if an organization
	// unit name conflicts at root level.
	queryCheckOrganizationUnitNameConflictRoot = dbmodel.DBQuery{
		ID:    "OUQ-OU_MGT-17",
		Query: `SELECT COUNT(*) as count FROM ORGANIZATION_UNIT WHERE NAME = $1 AND PARENT_ID IS NULL`,
	}

	// queryCheckOrganizationUnitHandleConflict is the query to check if an organization
	// unit handle conflicts under the same parent.
	queryCheckOrganizationUnitHandleConflict = dbmodel.DBQuery{
		ID:    "OUQ-OU_MGT-18",
		Query: `SELECT COUNT(*) as count FROM ORGANIZATION_UNIT WHERE HANDLE = $1 AND PARENT_ID = $2`,
	}

	// queryCheckOrganizationUnitHandleConflictRoot is the query to check if an organization
	// unit handle conflicts at root level.
	queryCheckOrganizationUnitHandleConflictRoot = dbmodel.DBQuery{
		ID:    "OUQ-OU_MGT-19",
		Query: `SELECT COUNT(*) as count FROM ORGANIZATION_UNIT WHERE HANDLE = $1 AND PARENT_ID IS NULL`,
	}

	// queryCheckOrganizationUnitHasUsersOrGroups is the query to check if an organization unit has users or groups.
	queryCheckOrganizationUnitHasUsersOrGroups = dbmodel.DBQuery{
		ID: "OUQ-OU_MGT-20",
		Query: `SELECT 
					(SELECT COUNT(*) FROM ORGANIZATION_UNIT WHERE PARENT_ID = $1) +
					(SELECT COUNT(*) FROM "USER" WHERE OU_ID = $1) + 
					(SELECT COUNT(*) FROM "GROUP" WHERE OU_ID = $1) as count`,
	}
)
