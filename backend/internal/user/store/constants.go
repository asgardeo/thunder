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

package store

import dbmodel "github.com/asgardeo/thunder/internal/system/database/model"

var (
	QueryCreateUser = dbmodel.DBQuery{
		Id:    "ASQ-USER_MGT-01",
		Query: "INSERT INTO \"USER\" (USER_ID, ORG_ID, TYPE, ATTRIBUTES) VALUES ($1, $2, $3, $4)",
	}
	QueryGetUserByUserId = dbmodel.DBQuery{
		Id:    "ASQ-USER_MGT-02",
		Query: "SELECT USER_ID, ORG_ID, TYPE, ATTRIBUTES FROM \"USER\" WHERE USER_ID = $1",
	}
	QueryGetUserList = dbmodel.DBQuery{
		Id:    "ASQ-USER_MGT-03",
		Query: "SELECT USER_ID, ORG_ID, TYPE, ATTRIBUTES FROM \"USER\"",
	}
	QueryUpdateUserByUserId = dbmodel.DBQuery{
		Id:    "ASQ-USER_MGT-04",
		Query: "UPDATE \"USER\" SET ORG_ID = $2, TYPE = $3, ATTRIBUTES = $4 WHERE USER_ID = $1;",
	}
	QueryDeleteUserByUserId = dbmodel.DBQuery{
		Id:    "ASQ-USER_MGT-05",
		Query: "DELETE FROM \"USER\" WHERE USER_ID = $1",
	}
)
