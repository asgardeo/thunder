/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildGetGroupListCountByOUIDsQuery(t *testing.T) {
	deploymentID := "dep1"
	testCases := []struct {
		name           string
		ouIDs          []string
		expectedPG     string
		expectedSQLite string
		expectedArgs   []interface{}
	}{
		{
			name:           "Empty list",
			ouIDs:          []string{},
			expectedPG:     "SELECT 0 WHERE 1=0",
			expectedSQLite: "SELECT 0 WHERE 1=0",
			expectedArgs:   []interface{}{},
		},
		{
			name:           "Single item",
			ouIDs:          []string{"ou1"},
			expectedPG:     `SELECT COUNT(*) as total FROM "GROUP" WHERE OU_ID IN ($1) AND DEPLOYMENT_ID = $2`,
			expectedSQLite: `SELECT COUNT(*) as total FROM "GROUP" WHERE OU_ID IN (?) AND DEPLOYMENT_ID = ?`,
			expectedArgs:   []interface{}{"ou1", deploymentID},
		},
		{
			name:           "Multiple items",
			ouIDs:          []string{"ou1", "ou2", "ou3"},
			expectedPG:     `SELECT COUNT(*) as total FROM "GROUP" WHERE OU_ID IN ($1,$2,$3) AND DEPLOYMENT_ID = $4`,
			expectedSQLite: `SELECT COUNT(*) as total FROM "GROUP" WHERE OU_ID IN (?,?,?) AND DEPLOYMENT_ID = ?`,
			expectedArgs:   []interface{}{"ou1", "ou2", "ou3", deploymentID},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, args := buildGetGroupsCountByOUIDsQuery(tc.ouIDs, deploymentID)
			require.Equal(t, tc.expectedPG, result.PostgresQuery)
			require.Equal(t, tc.expectedSQLite, result.SQLiteQuery)
			require.Equal(t, tc.expectedArgs, args)
		})
	}
}

func TestBuildGetGroupListByOUIDsQuery(t *testing.T) {
	limit, offset, deploymentID := 10, 5, "dep1"
	testCases := []struct {
		name           string
		ouIDs          []string
		expectedPG     string
		expectedSQLite string
		expectedArgs   []interface{}
	}{
		{
			name:           "Empty list",
			ouIDs:          []string{},
			expectedPG:     `SELECT ID, OU_ID, NAME, DESCRIPTION FROM "GROUP" WHERE 1=0`,
			expectedSQLite: `SELECT ID, OU_ID, NAME, DESCRIPTION FROM "GROUP" WHERE 1=0`,
			expectedArgs:   []interface{}{},
		},
		{
			name:  "Single item",
			ouIDs: []string{"ou1"},
			expectedPG: `SELECT ID, OU_ID, NAME, DESCRIPTION FROM "GROUP" ` +
				`WHERE OU_ID IN ($1) AND DEPLOYMENT_ID = $2 ORDER BY NAME LIMIT $3 OFFSET $4`,
			expectedSQLite: `SELECT ID, OU_ID, NAME, DESCRIPTION FROM "GROUP" ` +
				`WHERE OU_ID IN (?) AND DEPLOYMENT_ID = ? ORDER BY NAME LIMIT ? OFFSET ?`,
			expectedArgs: []interface{}{"ou1", deploymentID, limit, offset},
		},
		{
			name:  "Multiple items",
			ouIDs: []string{"ou1", "ou2", "ou3"},
			expectedPG: `SELECT ID, OU_ID, NAME, DESCRIPTION FROM "GROUP" ` +
				`WHERE OU_ID IN ($1,$2,$3) AND DEPLOYMENT_ID = $4 ORDER BY NAME LIMIT $5 OFFSET $6`,
			expectedSQLite: `SELECT ID, OU_ID, NAME, DESCRIPTION FROM "GROUP" ` +
				`WHERE OU_ID IN (?,?,?) AND DEPLOYMENT_ID = ? ORDER BY NAME LIMIT ? OFFSET ?`,
			expectedArgs: []interface{}{"ou1", "ou2", "ou3", deploymentID, limit, offset},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, args := buildGetGroupsByOUIDsQuery(tc.ouIDs, limit, offset, deploymentID)
			require.Equal(t, tc.expectedPG, result.PostgresQuery)
			require.Equal(t, tc.expectedSQLite, result.SQLiteQuery)
			require.Equal(t, tc.expectedArgs, args)
		})
	}
}
