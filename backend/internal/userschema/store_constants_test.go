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

package userschema

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asgardeo/thunder/internal/system/database/model"
)

type buildQueryFunc func([]string) model.DBQuery

func runBuildUserSchemaQueryTests(t *testing.T, cases []struct {
	name       string
	ouIDs      []string
	wantPG     string
	wantSQLite string
}, fn buildQueryFunc, expectedID string) {
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			query := fn(tc.ouIDs)
			assert.Equal(t, expectedID, query.ID)
			assert.Equal(t, tc.wantPG, query.PostgresQuery)
			assert.Equal(t, tc.wantSQLite, query.SQLiteQuery)
		})
	}
}

func TestBuildGetUserSchemaListByOUIDsQuery(t *testing.T) {
	testCases := []struct {
		name       string
		ouIDs      []string
		wantPG     string
		wantSQLite string
	}{
		{
			name:  "Empty OUIDs",
			ouIDs: []string{},
			wantPG: `SELECT SCHEMA_ID, NAME, OU_ID, ALLOW_SELF_REGISTRATION FROM USER_SCHEMAS ` +
				`WHERE 1=0 AND DEPLOYMENT_ID = $1 ORDER BY NAME LIMIT $2 OFFSET $3`,
			wantSQLite: `SELECT SCHEMA_ID, NAME, OU_ID, ALLOW_SELF_REGISTRATION FROM USER_SCHEMAS ` +
				`WHERE 1=0 AND DEPLOYMENT_ID = ? ORDER BY NAME LIMIT ? OFFSET ?`,
		},
		{
			name:  "Single OUID",
			ouIDs: []string{"ou-1"},
			wantPG: `SELECT SCHEMA_ID, NAME, OU_ID, ALLOW_SELF_REGISTRATION FROM USER_SCHEMAS ` +
				`WHERE OU_ID IN ($1) AND DEPLOYMENT_ID = $2 ORDER BY NAME LIMIT $3 OFFSET $4`,
			wantSQLite: `SELECT SCHEMA_ID, NAME, OU_ID, ALLOW_SELF_REGISTRATION FROM USER_SCHEMAS ` +
				`WHERE OU_ID IN (?) AND DEPLOYMENT_ID = ? ORDER BY NAME LIMIT ? OFFSET ?`,
		},
		{
			name:  "Multiple OUIDs",
			ouIDs: []string{"ou-1", "ou-2", "ou-3"},
			wantPG: `SELECT SCHEMA_ID, NAME, OU_ID, ALLOW_SELF_REGISTRATION FROM USER_SCHEMAS ` +
				`WHERE OU_ID IN ($1, $2, $3) AND DEPLOYMENT_ID = $4 ORDER BY NAME LIMIT $5 OFFSET $6`,
			wantSQLite: `SELECT SCHEMA_ID, NAME, OU_ID, ALLOW_SELF_REGISTRATION FROM USER_SCHEMAS ` +
				`WHERE OU_ID IN (?, ?, ?) AND DEPLOYMENT_ID = ? ORDER BY NAME LIMIT ? OFFSET ?`,
		},
	}
	runBuildUserSchemaQueryTests(t, testCases, buildGetUserSchemaListByOUIDsQuery, "ASQ-USER_SCHEMA-008")
}

func TestBuildGetUserSchemaCountByOUIDsQuery(t *testing.T) {
	testCases := []struct {
		name       string
		ouIDs      []string
		wantPG     string
		wantSQLite string
	}{
		{
			name:  "Empty OUIDs",
			ouIDs: []string{},
			wantPG: `SELECT COUNT(*) AS total FROM USER_SCHEMAS ` +
				`WHERE 1=0 AND DEPLOYMENT_ID = $1`,
			wantSQLite: `SELECT COUNT(*) AS total FROM USER_SCHEMAS ` +
				`WHERE 1=0 AND DEPLOYMENT_ID = ?`,
		},
		{
			name:  "Single OUID",
			ouIDs: []string{"ou-1"},
			wantPG: `SELECT COUNT(*) AS total FROM USER_SCHEMAS ` +
				`WHERE OU_ID IN ($1) AND DEPLOYMENT_ID = $2`,
			wantSQLite: `SELECT COUNT(*) AS total FROM USER_SCHEMAS ` +
				`WHERE OU_ID IN (?) AND DEPLOYMENT_ID = ?`,
		},
		{
			name:  "Multiple OUIDs",
			ouIDs: []string{"ou-1", "ou-2", "ou-3"},
			wantPG: `SELECT COUNT(*) AS total FROM USER_SCHEMAS ` +
				`WHERE OU_ID IN ($1, $2, $3) AND DEPLOYMENT_ID = $4`,
			wantSQLite: `SELECT COUNT(*) AS total FROM USER_SCHEMAS ` +
				`WHERE OU_ID IN (?, ?, ?) AND DEPLOYMENT_ID = ?`,
		},
	}
	runBuildUserSchemaQueryTests(t, testCases, buildGetUserSchemaCountByOUIDsQuery, "ASQ-USER_SCHEMA-009")
}
