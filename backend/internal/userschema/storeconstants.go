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

import dbmodel "github.com/asgardeo/thunder/internal/system/database/model"

var (
	// queryGetUserSchemaCount retrieves the total count of user schemas.
	queryGetUserSchemaCount = dbmodel.DBQuery{
		ID:    "ASQ-USER_SCHEMA-001",
		Query: `SELECT COUNT(*) AS total FROM USER_SCHEMAS`,
	}

	// queryGetUserSchemaList retrieves a paginated list of user schemas.
	queryGetUserSchemaList = dbmodel.DBQuery{
		ID:    "ASQ-USER_SCHEMA-002",
		Query: `SELECT SCHEMA_ID, NAME FROM USER_SCHEMAS ORDER BY NAME LIMIT $1 OFFSET $2`,
	}

	// queryCreateUserSchema creates a new user schema.
	queryCreateUserSchema = dbmodel.DBQuery{
		ID:    "ASQ-USER_SCHEMA-003",
		Query: `INSERT INTO USER_SCHEMAS (SCHEMA_ID, NAME, SCHEMA_DEF) VALUES ($1, $2, $3)`,
	}

	// queryGetUserSchemaByID retrieves a user schema by its ID.
	queryGetUserSchemaByID = dbmodel.DBQuery{
		ID:    "ASQ-USER_SCHEMA-004",
		Query: `SELECT SCHEMA_ID, NAME, SCHEMA_DEF FROM USER_SCHEMAS WHERE SCHEMA_ID = $1`,
	}

	// queryGetUserSchemaByName retrieves a user schema by its name.
	queryGetUserSchemaByName = dbmodel.DBQuery{
		ID:    "ASQ-USER_SCHEMA-005",
		Query: `SELECT SCHEMA_ID, NAME, SCHEMA_DEF FROM USER_SCHEMAS WHERE NAME = $1`,
	}

	// queryUpdateUserSchemaByID updates a user schema by its ID.
	queryUpdateUserSchemaByID = dbmodel.DBQuery{
		ID:    "ASQ-USER_SCHEMA-006",
		Query: `UPDATE USER_SCHEMAS SET NAME = $1, SCHEMA_DEF = $2 WHERE SCHEMA_ID = $3`,
	}

	// queryDeleteUserSchemaByID deletes a user schema by its ID.
	queryDeleteUserSchemaByID = dbmodel.DBQuery{
		ID:    "ASQ-USER_SCHEMA-007",
		Query: `DELETE FROM USER_SCHEMAS WHERE SCHEMA_ID = $1`,
	}
)
