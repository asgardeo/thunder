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

package constants

import dbmodel "github.com/asgardeo/thunder/internal/system/database/model"

// QueryInsertAuthorizationCode is the query to insert a new authorization code into the database.
var QueryInsertAuthorizationCode = dbmodel.DBQuery{
	ID: "AZQ-00001",
	Query: "INSERT INTO IDN_OAUTH2_AUTHZ_CODE (CODE_ID, AUTHORIZATION_CODE, CONSUMER_KEY, " +
		"CALLBACK_URL, AUTHZ_USER, TIME_CREATED, EXPIRY_TIME, STATE, CODE_CHALLENGE, CODE_CHALLENGE_METHOD)" +
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)",
}

// QueryInsertAuthorizationCodeScopes is the query to insert scopes for an authorization code.
var QueryInsertAuthorizationCodeScopes = dbmodel.DBQuery{
	ID:    "AZQ-00002",
	Query: "INSERT INTO IDN_OAUTH2_AUTHZ_CODE_SCOPE (CODE_ID, SCOPE) VALUES ($1, $2)",
}

// QueryGetAuthorizationCode is the query to retrieve an authorization code by client ID and code.
var QueryGetAuthorizationCode = dbmodel.DBQuery{
	ID: "AZQ-00003",
	Query: "SELECT CODE_ID, AUTHORIZATION_CODE, CALLBACK_URL, AUTHZ_USER, TIME_CREATED, " +
		"EXPIRY_TIME, STATE, CODE_CHALLENGE, CODE_CHALLENGE_METHOD FROM IDN_OAUTH2_AUTHZ_CODE WHERE " +
		"CONSUMER_KEY = $1 AND AUTHORIZATION_CODE = $2",
}

// QueryUpdateAuthorizationCodeState is the query to update the state of an authorization code.
var QueryUpdateAuthorizationCodeState = dbmodel.DBQuery{
	ID:    "AZQ-00004",
	Query: "UPDATE IDN_OAUTH2_AUTHZ_CODE SET STATE = $1 WHERE CODE_ID = $2",
}

// QueryGetAuthorizationCodeScopes is the query to retrieve scopes for an authorization code.
var QueryGetAuthorizationCodeScopes = dbmodel.DBQuery{
	ID:    "AZQ-00005",
	Query: "SELECT SCOPE FROM IDN_OAUTH2_AUTHZ_CODE_SCOPE WHERE CODE_ID = $1",
}
