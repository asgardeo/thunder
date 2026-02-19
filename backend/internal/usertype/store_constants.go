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

package usertype

import dbmodel "github.com/asgardeo/thunder/internal/system/database/model"

var (
	// queryGetUserTypeCount retrieves the total count of user types.
	queryGetUserTypeCount = dbmodel.DBQuery{
		ID:    "ASQ-USER_TYPE-001",
		Query: `SELECT COUNT(*) AS total FROM USER_TYPES WHERE DEPLOYMENT_ID = $1`,
	}

	// queryGetUserTypeList retrieves a paginated list of user types.
	queryGetUserTypeList = dbmodel.DBQuery{
		ID: "ASQ-USER_TYPE-002",
		Query: `SELECT SCHEMA_ID, NAME, OU_ID, ALLOW_SELF_REGISTRATION FROM USER_TYPES ` +
			`WHERE DEPLOYMENT_ID = $3 ORDER BY NAME LIMIT $1 OFFSET $2`,
	}

	// queryCreateUserType creates a new user type.
	queryCreateUserType = dbmodel.DBQuery{
		ID: "ASQ-USER_TYPE-003",
		Query: `INSERT INTO USER_TYPES (SCHEMA_ID, NAME, OU_ID, ALLOW_SELF_REGISTRATION, SCHEMA_DEF, DEPLOYMENT_ID)
			VALUES ($1, $2, $3, $4, $5, $6)`,
	}

	// queryGetUserTypeByID retrieves a user type by its ID.
	queryGetUserTypeByID = dbmodel.DBQuery{
		ID: "ASQ-USER_TYPE-004",
		Query: `SELECT SCHEMA_ID, NAME, OU_ID, ALLOW_SELF_REGISTRATION, SCHEMA_DEF FROM USER_TYPES ` +
			`WHERE SCHEMA_ID = $1 AND DEPLOYMENT_ID = $2`,
	}

	// queryGetUserTypeByName retrieves a user type by its name.
	queryGetUserTypeByName = dbmodel.DBQuery{
		ID: "ASQ-USER_TYPE-005",
		Query: `SELECT SCHEMA_ID, NAME, OU_ID, ALLOW_SELF_REGISTRATION, SCHEMA_DEF FROM USER_TYPES ` +
			`WHERE NAME = $1 AND DEPLOYMENT_ID = $2`,
	}

	// queryUpdateUserTypeByID updates a user type by its ID.
	queryUpdateUserTypeByID = dbmodel.DBQuery{
		ID: "ASQ-USER_TYPE-006",
		Query: `UPDATE USER_TYPES
			SET NAME = $1, OU_ID = $2, ALLOW_SELF_REGISTRATION = $3, SCHEMA_DEF = $4
			WHERE SCHEMA_ID = $5 AND DEPLOYMENT_ID = $6`,
	}

	// queryDeleteUserTypeByID deletes a user type by its ID.
	queryDeleteUserTypeByID = dbmodel.DBQuery{
		ID:    "ASQ-USER_TYPE-007",
		Query: `DELETE FROM USER_TYPES WHERE SCHEMA_ID = $1 AND DEPLOYMENT_ID = $2`,
	}
)
