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

package userschema

import (
	"fmt"
	"strings"

	dbmodel "github.com/asgardeo/thunder/internal/system/database/model"
)

var (
	// queryGetUserSchemaCount retrieves the total count of user schemas.
	queryGetUserSchemaCount = dbmodel.DBQuery{
		ID:    "ASQ-USER_SCHEMA-001",
		Query: `SELECT COUNT(*) AS total FROM USER_SCHEMAS WHERE DEPLOYMENT_ID = $1`,
	}

	// queryGetUserSchemaList retrieves a paginated list of user schemas.
	queryGetUserSchemaList = dbmodel.DBQuery{
		ID: "ASQ-USER_SCHEMA-002",
		Query: `SELECT ID, NAME, OU_ID, ALLOW_SELF_REGISTRATION FROM USER_SCHEMAS ` +
			`WHERE DEPLOYMENT_ID = $3 ORDER BY NAME LIMIT $1 OFFSET $2`,
	}

	// queryCreateUserSchema creates a new user schema.
	queryCreateUserSchema = dbmodel.DBQuery{
		ID: "ASQ-USER_SCHEMA-003",
		Query: `INSERT INTO USER_SCHEMAS (ID, NAME, OU_ID, ALLOW_SELF_REGISTRATION, SCHEMA_DEF, DEPLOYMENT_ID)
			VALUES ($1, $2, $3, $4, $5, $6)`,
	}

	// queryGetUserSchemaByID retrieves a user schema by its ID.
	queryGetUserSchemaByID = dbmodel.DBQuery{
		ID: "ASQ-USER_SCHEMA-004",
		Query: `SELECT ID, NAME, OU_ID, ALLOW_SELF_REGISTRATION, SCHEMA_DEF FROM USER_SCHEMAS ` +
			`WHERE ID = $1 AND DEPLOYMENT_ID = $2`,
	}

	// queryGetUserSchemaByName retrieves a user schema by its name.
	queryGetUserSchemaByName = dbmodel.DBQuery{
		ID: "ASQ-USER_SCHEMA-005",
		Query: `SELECT ID, NAME, OU_ID, ALLOW_SELF_REGISTRATION, SCHEMA_DEF FROM USER_SCHEMAS ` +
			`WHERE NAME = $1 AND DEPLOYMENT_ID = $2`,
	}

	// queryUpdateUserSchemaByID updates a user schema by its ID.
	queryUpdateUserSchemaByID = dbmodel.DBQuery{
		ID: "ASQ-USER_SCHEMA-006",
		Query: `UPDATE USER_SCHEMAS
			SET NAME = $1, OU_ID = $2, ALLOW_SELF_REGISTRATION = $3, SCHEMA_DEF = $4
			WHERE ID = $5 AND DEPLOYMENT_ID = $6`,
	}

	// queryDeleteUserSchemaByID deletes a user schema by its ID.
	queryDeleteUserSchemaByID = dbmodel.DBQuery{
		ID:    "ASQ-USER_SCHEMA-007",
		Query: `DELETE FROM USER_SCHEMAS WHERE ID = $1 AND DEPLOYMENT_ID = $2`,
	}
)

// buildGetUserSchemaListByOUIDsQuery dynamically builds a query to retrieve user schemas
// filtered by a list of OU IDs with pagination.
// For PostgreSQL: WHERE OU_ID IN ($1, $2, ...) AND DEPLOYMENT_ID = $N ORDER BY NAME LIMIT $N+1 OFFSET $N+2
// For SQLite: WHERE OU_ID IN (?, ?, ...) AND DEPLOYMENT_ID = ? ORDER BY NAME LIMIT ? OFFSET ?
func buildGetUserSchemaListByOUIDsQuery(ouIDs []string) dbmodel.DBQuery {
	n := len(ouIDs)

	if n == 0 {
		return dbmodel.DBQuery{
			ID: "ASQ-USER_SCHEMA-008",
			PostgresQuery: `SELECT SCHEMA_ID, NAME, OU_ID, ALLOW_SELF_REGISTRATION FROM USER_SCHEMAS ` +
				`WHERE 1=0 AND DEPLOYMENT_ID = $1 ORDER BY NAME LIMIT $2 OFFSET $3`,
			SQLiteQuery: `SELECT SCHEMA_ID, NAME, OU_ID, ALLOW_SELF_REGISTRATION FROM USER_SCHEMAS ` +
				`WHERE 1=0 AND DEPLOYMENT_ID = ? ORDER BY NAME LIMIT ? OFFSET ?`,
		}
	}

	// Build PostgreSQL placeholders: $1, $2, ..., $N
	pgPlaceholders := make([]string, n)
	for i := range ouIDs {
		pgPlaceholders[i] = fmt.Sprintf("$%d", i+1)
	}
	pgInClause := strings.Join(pgPlaceholders, ", ")
	pgDeploymentID := fmt.Sprintf("$%d", n+1)
	pgLimit := fmt.Sprintf("$%d", n+2)
	pgOffset := fmt.Sprintf("$%d", n+3)

	// Build SQLite placeholders: ?, ?, ...
	sqlitePlaceholders := make([]string, n)
	for i := range ouIDs {
		sqlitePlaceholders[i] = "?"
	}
	sqliteInClause := strings.Join(sqlitePlaceholders, ", ")

	return dbmodel.DBQuery{
		ID: "ASQ-USER_SCHEMA-008",
		PostgresQuery: `SELECT SCHEMA_ID, NAME, OU_ID, ALLOW_SELF_REGISTRATION FROM USER_SCHEMAS ` +
			`WHERE OU_ID IN (` + pgInClause + `) AND DEPLOYMENT_ID = ` + pgDeploymentID +
			` ORDER BY NAME LIMIT ` + pgLimit + ` OFFSET ` + pgOffset,
		SQLiteQuery: `SELECT SCHEMA_ID, NAME, OU_ID, ALLOW_SELF_REGISTRATION FROM USER_SCHEMAS ` +
			`WHERE OU_ID IN (` + sqliteInClause + `) AND DEPLOYMENT_ID = ? ORDER BY NAME LIMIT ? OFFSET ?`,
	}
}

// buildGetUserSchemaCountByOUIDsQuery dynamically builds a query to count user schemas
// filtered by a list of OU IDs.
// For PostgreSQL: WHERE OU_ID IN ($1, $2, ...) AND DEPLOYMENT_ID = $N
// For SQLite: WHERE OU_ID IN (?, ?, ...) AND DEPLOYMENT_ID = ?
func buildGetUserSchemaCountByOUIDsQuery(ouIDs []string) dbmodel.DBQuery {
	n := len(ouIDs)

	if n == 0 {
		return dbmodel.DBQuery{
			ID: "ASQ-USER_SCHEMA-009",
			PostgresQuery: `SELECT COUNT(*) AS total FROM USER_SCHEMAS ` +
				`WHERE 1=0 AND DEPLOYMENT_ID = $1`,
			SQLiteQuery: `SELECT COUNT(*) AS total FROM USER_SCHEMAS ` +
				`WHERE 1=0 AND DEPLOYMENT_ID = ?`,
		}
	}

	// Build PostgreSQL placeholders: $1, $2, ..., $N
	pgPlaceholders := make([]string, n)
	for i := range ouIDs {
		pgPlaceholders[i] = fmt.Sprintf("$%d", i+1)
	}
	pgInClause := strings.Join(pgPlaceholders, ", ")
	pgDeploymentID := fmt.Sprintf("$%d", n+1)

	// Build SQLite placeholders: ?, ?, ...
	sqlitePlaceholders := make([]string, n)
	for i := range ouIDs {
		sqlitePlaceholders[i] = "?"
	}
	sqliteInClause := strings.Join(sqlitePlaceholders, ", ")

	return dbmodel.DBQuery{
		ID: "ASQ-USER_SCHEMA-009",
		PostgresQuery: `SELECT COUNT(*) AS total FROM USER_SCHEMAS ` +
			`WHERE OU_ID IN (` + pgInClause + `) AND DEPLOYMENT_ID = ` + pgDeploymentID,
		SQLiteQuery: `SELECT COUNT(*) AS total FROM USER_SCHEMAS ` +
			`WHERE OU_ID IN (` + sqliteInClause + `) AND DEPLOYMENT_ID = ?`,
	}
}
