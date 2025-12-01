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

package authz

import dbmodel "github.com/asgardeo/thunder/internal/system/database/model"

// queryInsertAuthorizationCode is the query to insert a new authorization code into the database.
var queryInsertAuthorizationCode = dbmodel.DBQuery{
	ID: "AZQ-ACS-01",
	Query: "INSERT INTO AUTHORIZATION_CODE (CODE_ID, AUTHORIZATION_CODE, CLIENT_ID, STATE, AUTHZ_DATA, " +
		"TIME_CREATED, EXPIRY_TIME, DEPLOYMENT_ID) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
}

// queryGetAuthorizationCode is the query to retrieve an authorization code by client ID and code.
var queryGetAuthorizationCode = dbmodel.DBQuery{
	ID: "AZQ-ACS-02",
	Query: "SELECT CODE_ID, AUTHORIZATION_CODE, CLIENT_ID, STATE, AUTHZ_DATA, TIME_CREATED, " +
		"EXPIRY_TIME FROM AUTHORIZATION_CODE WHERE CLIENT_ID = $1 AND AUTHORIZATION_CODE = $2 AND DEPLOYMENT_ID = $3",
}

// queryUpdateAuthorizationCodeState is the query to update the state of an authorization code.
var queryUpdateAuthorizationCodeState = dbmodel.DBQuery{
	ID:    "AZQ-ACS-03",
	Query: "UPDATE AUTHORIZATION_CODE SET STATE = $1 WHERE CODE_ID = $2 AND DEPLOYMENT_ID = $3",
}
