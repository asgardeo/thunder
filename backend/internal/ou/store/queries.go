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

// Package store provides the implementation for organization unit persistence operations.
package store

import (
	"fmt"
	"strings"

	dbmodel "github.com/asgardeo/thunder/internal/system/database/model"
)

var (
	// QueryGetOrganizationUnitListCount is the query to get total count of organization units.
	QueryGetOrganizationUnitListCount = dbmodel.DBQuery{
		ID:    "OUQ-OU_MGT-00",
		Query: `SELECT COUNT(*) as total FROM ORGANIZATION_UNIT`,
	}

	// QueryGetOrganizationUnitList is the query to get organization units with pagination.
	QueryGetOrganizationUnitList = dbmodel.DBQuery{
		ID:    "OUQ-OU_MGT-01",
		Query: `SELECT OU_ID, HANDLE, NAME, DESCRIPTION, PARENT_ID FROM ORGANIZATION_UNIT ORDER BY NAME LIMIT $1 OFFSET $2`,
	}

	// QueryCreateOrganizationUnit is the query to create a new organization unit.
	QueryCreateOrganizationUnit = dbmodel.DBQuery{
		ID:    "OUQ-OU_MGT-03",
		Query: `INSERT INTO ORGANIZATION_UNIT (OU_ID, PARENT_ID, HANDLE, NAME, DESCRIPTION) VALUES ($1, $2, $3, $4, $5)`,
	}

	// QueryGetOrganizationUnitByID is the query to get an organization unit by id.
	QueryGetOrganizationUnitByID = dbmodel.DBQuery{
		ID:    "OUQ-OU_MGT-04",
		Query: `SELECT OU_ID, PARENT_ID, HANDLE, NAME, DESCRIPTION FROM ORGANIZATION_UNIT WHERE OU_ID = $1`,
	}

	// QueryUpdateOrganizationUnit is the query to update an organization unit.
	QueryUpdateOrganizationUnit = dbmodel.DBQuery{
		ID:    "OUQ-OU_MGT-05",
		Query: `UPDATE ORGANIZATION_UNIT SET PARENT_ID = $2, HANDLE = $3, NAME = $4, DESCRIPTION = $5 WHERE OU_ID = $1`,
	}

	// QueryDeleteOrganizationUnit is the query to delete an organization unit.
	QueryDeleteOrganizationUnit = dbmodel.DBQuery{
		ID:    "OUQ-OU_MGT-06",
		Query: `DELETE FROM ORGANIZATION_UNIT WHERE OU_ID = $1`,
	}

	// QueryGetSubOrganizationUnits is the query to get sub organization units of an organization unit.
	QueryGetSubOrganizationUnits = dbmodel.DBQuery{
		ID:    "OUQ-OU_MGT-07",
		Query: `SELECT OU_ID FROM ORGANIZATION_UNIT WHERE PARENT_ID = $1`,
	}

	// QueryGetOrganizationUnitUsers is the query to get users in an organization unit.
	QueryGetOrganizationUnitUsers = dbmodel.DBQuery{
		ID:    "OUQ-OU_MGT-08",
		Query: `SELECT USER_ID FROM "USER" WHERE OU_ID = $1`,
	}

	// QueryGetOrganizationUnitGroups is the query to get groups in an organization unit.
	QueryGetOrganizationUnitGroups = dbmodel.DBQuery{
		ID:    "OUQ-OU_MGT-09",
		Query: `SELECT GROUP_ID FROM "GROUP" WHERE OU_ID = $1`,
	}

	// QueryCheckOrganizationUnitHandleConflict is the query to check if an organization
	// unit handle conflicts under the same parent.
	QueryCheckOrganizationUnitHandleConflict = dbmodel.DBQuery{
		ID:    "OUQ-OU_MGT-10",
		Query: `SELECT COUNT(*) as count FROM ORGANIZATION_UNIT WHERE HANDLE = $1 AND PARENT_ID = $2`,
	}

	// QueryCheckOrganizationUnitHandleConflictRoot is the query to check if an organization
	// unit handle conflicts at root level.
	QueryCheckOrganizationUnitHandleConflictRoot = dbmodel.DBQuery{
		ID:    "OUQ-OU_MGT-11",
		Query: `SELECT COUNT(*) as count FROM ORGANIZATION_UNIT WHERE HANDLE = $1 AND PARENT_ID IS NULL`,
	}

	// QueryCheckOrganizationUnitNameConflict is the query to check if an organization
	// unit name conflicts under the same parent.
	QueryCheckOrganizationUnitNameConflict = dbmodel.DBQuery{
		ID:    "OUQ-OU_MGT-12",
		Query: `SELECT COUNT(*) as count FROM ORGANIZATION_UNIT WHERE NAME = $1 AND PARENT_ID = $2`,
	}

	// QueryCheckOrganizationUnitNameConflictRoot is the query to check if an organization
	// unit name conflicts at root level.
	QueryCheckOrganizationUnitNameConflictRoot = dbmodel.DBQuery{
		ID:    "OUQ-OU_MGT-13",
		Query: `SELECT COUNT(*) as count FROM ORGANIZATION_UNIT WHERE NAME = $1 AND PARENT_ID IS NULL`,
	}

	// QueryCheckOrganizationUnitHasUsersOrGroups is the query to check if an organization unit has users or groups.
	QueryCheckOrganizationUnitHasUsersOrGroups = dbmodel.DBQuery{
		ID: "OUQ-OU_MGT-14",
		Query: `SELECT 
					(SELECT COUNT(*) FROM ORGANIZATION_UNIT WHERE PARENT_ID = $1) +
					(SELECT COUNT(*) FROM "USER" WHERE OU_ID = $1) + 
					(SELECT COUNT(*) FROM "GROUP" WHERE OU_ID = $1) as count`,
	}
)

// buildSubOrganizationUnitsQuery constructs a query to get sub organization units for multiple parent IDs.
func buildSubOrganizationUnitsQuery(parentIDs []string) (dbmodel.DBQuery, []interface{}, error) {
	if len(parentIDs) == 0 {
		return dbmodel.DBQuery{}, nil, fmt.Errorf("parentIDs list cannot be empty")
	}

	args := make([]interface{}, len(parentIDs))

	postgresPlaceholders := make([]string, len(parentIDs))
	sqlitePlaceholders := make([]string, len(parentIDs))

	for i, parentID := range parentIDs {
		postgresPlaceholders[i] = fmt.Sprintf("$%d", i+1)
		sqlitePlaceholders[i] = "?"
		args[i] = parentID
	}

	baseQuery := "SELECT OU_ID, PARENT_ID FROM ORGANIZATION_UNIT WHERE PARENT_ID IN (%s)"
	postgresQuery := fmt.Sprintf(baseQuery, strings.Join(postgresPlaceholders, ","))
	sqliteQuery := fmt.Sprintf(baseQuery, strings.Join(sqlitePlaceholders, ","))

	query := dbmodel.DBQuery{
		ID:            "OUQ-OU_MGT-02",
		Query:         postgresQuery,
		PostgresQuery: postgresQuery,
		SQLiteQuery:   sqliteQuery,
	}

	return query, args, nil
}
