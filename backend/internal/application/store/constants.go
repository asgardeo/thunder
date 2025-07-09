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
	// QueryGetApplicationByClientID is the query to retrieve application details by client ID.
	QueryGetApplicationByClientID = dbmodel.DBQuery{
		ID: "ASQ-APP_MGT-00",
		Query: "SELECT CONSUMER_KEY, CONSUMER_SECRET, CALLBACK_URIS, GRANT_TYPES " +
			"FROM IDN_OAUTH_CONSUMER_APPS WHERE CONSUMER_KEY = $1",
	}

	// QueryCreateApplication is the query to create a new application with basic details.
	QueryCreateApplication = dbmodel.DBQuery{
		ID: "ASQ-APP_MGT-01",
		Query: "INSERT INTO SP_APP (APP_ID, ORG_ID, APP_NAME, DESCRIPTION, AUTH_FLOW_GRAPH_ID, REGISTRATION_FLOW_GRAPH_ID) " +
			"VALUES ($1, $2, $3, $4, $5, $6)",
	}
	// QueryCreateOAuthApplication is the query to create a new OAuth application.
	QueryCreateOAuthApplication = dbmodel.DBQuery{
		ID: "ASQ-APP_MGT-02",
		Query: "INSERT INTO IDN_OAUTH_CONSUMER_APPS (APP_ID, CONSUMER_KEY, CONSUMER_SECRET, CALLBACK_URIS, " +
			"GRANT_TYPES) VALUES ($1, $2, $3, $4, $5)",
	}
	// QueryGetApplicationByAppID is the query to retrieve application details by app ID.
	QueryGetApplicationByAppID = dbmodel.DBQuery{
		ID: "ASQ-APP_MGT-03",
		Query: "SELECT sp.APP_ID, sp.ORG_ID, sp.APP_NAME, sp.DESCRIPTION, sp.AUTH_FLOW_GRAPH_ID, " +
			"sp.REGISTRATION_FLOW_GRAPH_ID, oauth.CONSUMER_KEY, oauth.CALLBACK_URIS, oauth.GRANT_TYPES " +
			"FROM SP_APP sp JOIN IDN_OAUTH_CONSUMER_APPS oauth ON sp.APP_ID = oauth.APP_ID " +
			"WHERE sp.APP_ID = $1",
	}
	// QueryGetApplicationList is the query to list all the applications.
	QueryGetApplicationList = dbmodel.DBQuery{
		ID: "ASQ-APP_MGT-04",
		Query: "SELECT sp.APP_ID, sp.ORG_ID, sp.APP_NAME, sp.DESCRIPTION, sp.AUTH_FLOW_GRAPH_ID, " +
			"sp.REGISTRATION_FLOW_GRAPH_ID, oauth.CONSUMER_KEY, oauth.CALLBACK_URIS, oauth.GRANT_TYPES " +
			"FROM SP_APP sp JOIN IDN_OAUTH_CONSUMER_APPS oauth ON sp.APP_ID = oauth.APP_ID",
	}
	// QueryUpdateApplicationByAppID is the query to update application details by app ID.
	QueryUpdateApplicationByAppID = dbmodel.DBQuery{
		ID: "ASQ-APP_MGT-05",
		Query: "UPDATE SP_APP SET ORG_ID=$2, APP_NAME=$3, DESCRIPTION=$4, AUTH_FLOW_GRAPH_ID=$5, " +
			"REGISTRATION_FLOW_GRAPH_ID=$6 WHERE APP_ID = $1;",
	}
	// QueryUpdateOAuthApplicationByAppID is the query to update OAuth application details by app ID.
	QueryUpdateOAuthApplicationByAppID = dbmodel.DBQuery{
		ID: "ASQ-APP_MGT-05",
		Query: "UPDATE IDN_OAUTH_CONSUMER_APPS " +
			"SET CONSUMER_KEY=$2, CONSUMER_SECRET=$3, CALLBACK_URIS=$4, GRANT_TYPES=$5 WHERE APP_ID=$1",
	}
	// QueryDeleteApplicationByAppID is the query to delete an application by app ID.
	QueryDeleteApplicationByAppID = dbmodel.DBQuery{
		ID:    "ASQ-APP_MGT-06",
		Query: "DELETE FROM SP_APP WHERE APP_ID = $1",
	}
)
