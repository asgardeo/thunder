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

// Package store provides the implementation for flow context persistence operations.
package store

import (
	"github.com/asgardeo/thunder/internal/system/database/model"
)

var (
	// QueryCreateFlowContext is the query to create a new flow context.
	QueryCreateFlowContext = model.DBQuery{
		ID: "FLQ-FLOW_CTX-01",
		Query: "INSERT INTO FLOW_CONTEXT (FLOW_ID, APP_ID, CURRENT_NODE_ID, " +
			"CURRENT_ACTION_ID, GRAPH_ID, RUNTIME_DATA) VALUES ($1, $2, $3, $4, $5, $6)",
	}

	// QueryUpdateFlowContext is the query to update a flow context.
	QueryUpdateFlowContext = model.DBQuery{
		ID: "FLQ-FLOW_CTX-03",
		Query: "UPDATE FLOW_CONTEXT SET CURRENT_NODE_ID = $2, CURRENT_ACTION_ID = $3, " +
			"RUNTIME_DATA = $4, UPDATED_AT = CURRENT_TIMESTAMP WHERE FLOW_ID = $1",
	}

	// QueryDeleteFlowContext is the query to delete a flow context.
	QueryDeleteFlowContext = model.DBQuery{
		ID:    "FLQ-FLOW_CTX-04",
		Query: "DELETE FROM FLOW_CONTEXT WHERE FLOW_ID = $1",
	}

	// QueryCreateFlowUserData is the query to create flow user data.
	QueryCreateFlowUserData = model.DBQuery{
		ID: "FLQ-FLOW_USER-01",
		Query: "INSERT INTO FLOW_USER_DATA (FLOW_ID, IS_AUTHENTICATED, USER_ID, " +
			"USER_INPUTS, USER_ATTRIBUTES) VALUES ($1, $2, $3, $4, $5)",
	}

	// QueryUpdateFlowUserData is the query to update flow user data.
	QueryUpdateFlowUserData = model.DBQuery{
		ID: "FLQ-FLOW_USER-03",
		Query: "UPDATE FLOW_USER_DATA SET IS_AUTHENTICATED = $2, USER_ID = $3, " +
			"USER_INPUTS = $4, USER_ATTRIBUTES = $5, " +
			"UPDATED_AT = CURRENT_TIMESTAMP WHERE FLOW_ID = $1",
	}

	// QueryDeleteFlowUserData is the query to delete flow user data.
	QueryDeleteFlowUserData = model.DBQuery{
		ID:    "FLQ-FLOW_USER-04",
		Query: "DELETE FROM FLOW_USER_DATA WHERE FLOW_ID = $1",
	}

	// QueryGetFlowContextWithUserData is the query to get flow context with user data in a single query.
	QueryGetFlowContextWithUserData = model.DBQuery{
		ID: "FLQ-FLOW_CTX-05",
		Query: `SELECT 
			fc.FLOW_ID, fc.APP_ID, fc.CURRENT_NODE_ID, fc.CURRENT_ACTION_ID, 
			fc.GRAPH_ID, fc.RUNTIME_DATA, fc.CREATED_AT, fc.UPDATED_AT,
			fud.IS_AUTHENTICATED, fud.USER_ID, fud.USER_INPUTS, 
			fud.USER_ATTRIBUTES
		FROM FLOW_CONTEXT fc
		LEFT JOIN FLOW_USER_DATA fud ON fc.FLOW_ID = fud.FLOW_ID
		WHERE fc.FLOW_ID = $1`,
	}
)
