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

// Package store defines the database queries for user schema management operations.
package store

import "github.com/asgardeo/thunder/internal/system/database/model"

var (
	// QueryGetUserSchemaCount retrieves the total count of user schemas.
	QueryGetUserSchemaCount = model.DBQuery{
		ID:    "ASQ-USER_SCHEMA-001",
		Query: `SELECT COUNT(*) AS total FROM USER_SCHEMAS`,
	}

	// QueryGetUserSchemaList retrieves a paginated list of user schemas.
	QueryGetUserSchemaList = model.DBQuery{
		ID:    "ASQ-USER_SCHEMA-002",
		Query: `SELECT SCHEMA_ID, NAME FROM USER_SCHEMAS ORDER BY NAME LIMIT $1 OFFSET $2`,
	}

	// QueryCreateUserSchema creates a new user schema.
	QueryCreateUserSchema = model.DBQuery{
		ID:    "ASQ-USER_SCHEMA-003",
		Query: `INSERT INTO USER_SCHEMAS (SCHEMA_ID, NAME, SCHEMA_DEF) VALUES ($1, $2, $3)`,
	}

	// QueryGetUserSchemaByID retrieves a user schema by its ID.
	QueryGetUserSchemaByID = model.DBQuery{
		ID:    "ASQ-USER_SCHEMA-004",
		Query: `SELECT SCHEMA_ID, NAME, SCHEMA_DEF FROM USER_SCHEMAS WHERE SCHEMA_ID = $1`,
	}

	// QueryGetUserSchemaByName retrieves a user schema by its name.
	QueryGetUserSchemaByName = model.DBQuery{
		ID:    "ASQ-USER_SCHEMA-005",
		Query: `SELECT SCHEMA_ID, NAME, SCHEMA_DEF FROM USER_SCHEMAS WHERE NAME = $1`,
	}

	// QueryUpdateUserSchemaByID updates a user schema by its ID.
	QueryUpdateUserSchemaByID = model.DBQuery{
		ID:    "ASQ-USER_SCHEMA-006",
		Query: `UPDATE USER_SCHEMAS SET NAME = $1, SCHEMA_DEF = $2 WHERE SCHEMA_ID = $3`,
	}

	// QueryDeleteUserSchemaByID deletes a user schema by its ID.
	QueryDeleteUserSchemaByID = model.DBQuery{
		ID:    "ASQ-USER_SCHEMA-007",
		Query: `DELETE FROM USER_SCHEMAS WHERE SCHEMA_ID = $1`,
	}
)
