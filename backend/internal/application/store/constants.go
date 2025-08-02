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
	// QueryCreateApplication is the query to create a new application with basic details.
	QueryCreateApplication = dbmodel.DBQuery{
		ID: "ASQ-APP_MGT-01",
		Query: "INSERT INTO SP_APP (APP_ID, APP_NAME, DESCRIPTION, AUTH_FLOW_GRAPH_ID, REGISTRATION_FLOW_GRAPH_ID, " +
			"IS_REGISTRATION_FLOW_ENABLED, APP_JSON) VALUES ($1, $2, $3, $4, $5, $6, $7)",
	}
	// QueryCreateOAuthApplication is the query to create a new OAuth application.
	QueryCreateOAuthApplication = dbmodel.DBQuery{
		ID: "ASQ-APP_MGT-02",
		Query: "INSERT INTO IDN_OAUTH_CONSUMER_APPS (APP_ID, CONSUMER_KEY, CONSUMER_SECRET, CALLBACK_URIS, " +
			"GRANT_TYPES, RESPONSE_TYPES, TOKEN_ENDPOINT_AUTH_METHODS) VALUES ($1, $2, $3, $4, $5, $6, $7)",
	}
	// QueryGetApplicationByAppID is the query to retrieve application details by app ID.
	QueryGetApplicationByAppID = dbmodel.DBQuery{
		ID: "ASQ-APP_MGT-03",
		Query: "SELECT sp.APP_ID, sp.APP_NAME, sp.DESCRIPTION, sp.AUTH_FLOW_GRAPH_ID, " +
			"sp.REGISTRATION_FLOW_GRAPH_ID, sp.IS_REGISTRATION_FLOW_ENABLED, sp.APP_JSON, " +
			"oauth.CONSUMER_KEY, oauth.CONSUMER_SECRET, oauth.CALLBACK_URIS, oauth.GRANT_TYPES, " +
			"oauth.RESPONSE_TYPES, oauth.TOKEN_ENDPOINT_AUTH_METHODS " +
			"FROM SP_APP sp JOIN IDN_OAUTH_CONSUMER_APPS oauth ON sp.APP_ID = oauth.APP_ID WHERE sp.APP_ID = $1",
	}
	// QueryGetOAuthApplicationByClientID is the query to retrieve oauth application details by client ID.
	QueryGetOAuthApplicationByClientID = dbmodel.DBQuery{
		ID: "ASQ-APP_MGT-04",
		Query: "SELECT APP_ID, CONSUMER_KEY, CONSUMER_SECRET, CALLBACK_URIS, GRANT_TYPES, RESPONSE_TYPES, " +
			"TOKEN_ENDPOINT_AUTH_METHODS FROM IDN_OAUTH_CONSUMER_APPS WHERE CONSUMER_KEY = $1",
	}
	// QueryGetApplicationList is the query to list all the applications.
	QueryGetApplicationList = dbmodel.DBQuery{
		ID: "ASQ-APP_MGT-05",
		Query: "SELECT sp.APP_ID, sp.APP_NAME, sp.DESCRIPTION, sp.AUTH_FLOW_GRAPH_ID, " +
			"sp.REGISTRATION_FLOW_GRAPH_ID, sp.IS_REGISTRATION_FLOW_ENABLED, " +
			"oauth.CONSUMER_KEY FROM SP_APP sp JOIN IDN_OAUTH_CONSUMER_APPS oauth ON sp.APP_ID = oauth.APP_ID",
	}
	// QueryUpdateApplicationByAppID is the query to update application details by app ID.
	QueryUpdateApplicationByAppID = dbmodel.DBQuery{
		ID: "ASQ-APP_MGT-06",
		Query: "UPDATE SP_APP SET APP_NAME=$2, DESCRIPTION=$3, AUTH_FLOW_GRAPH_ID=$4, " +
			"REGISTRATION_FLOW_GRAPH_ID=$5, IS_REGISTRATION_FLOW_ENABLED=$6, APP_JSON=$7 " +
			"WHERE APP_ID = $1",
	}
	// QueryUpdateOAuthApplicationByAppID is the query to update OAuth application details by app ID.
	QueryUpdateOAuthApplicationByAppID = dbmodel.DBQuery{
		ID: "ASQ-APP_MGT-07",
		Query: "UPDATE IDN_OAUTH_CONSUMER_APPS SET CONSUMER_KEY=$2, CONSUMER_SECRET=$3, CALLBACK_URIS=$4, " +
			"GRANT_TYPES=$5, RESPONSE_TYPES=$6, TOKEN_ENDPOINT_AUTH_METHODS=$7 WHERE APP_ID=$1",
	}
	// QueryDeleteApplicationByAppID is the query to delete an application by app ID.
	QueryDeleteApplicationByAppID = dbmodel.DBQuery{
		ID:    "ASQ-APP_MGT-08",
		Query: "DELETE FROM SP_APP WHERE APP_ID = $1",
	}
)
