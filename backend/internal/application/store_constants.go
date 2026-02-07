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

package application

import dbmodel "github.com/asgardeo/thunder/internal/system/database/model"

var (
	// QueryCreateApplication is the query to create a new application with basic details.
	QueryCreateApplication = dbmodel.DBQuery{
		ID: "ASQ-APP_MGT-01",
		Query: "INSERT INTO SP_APP (APP_ID, APP_NAME, DESCRIPTION, AUTH_FLOW_ID, " +
			"REGISTRATION_FLOW_ID, IS_REGISTRATION_FLOW_ENABLED, THEME_ID, LAYOUT_ID, APP_JSON, DEPLOYMENT_ID) " +
			"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)",
	}
	// QueryCreateOAuthApplication is the query to create a new OAuth application.
	QueryCreateOAuthApplication = dbmodel.DBQuery{
		ID: "ASQ-APP_MGT-02",
		Query: "INSERT INTO IDN_OAUTH_CONSUMER_APPS (APP_ID, CONSUMER_KEY, CONSUMER_SECRET, OAUTH_CONFIG_JSON, " +
			"DEPLOYMENT_ID) VALUES ($1, $2, $3, $4, $5)",
	}
	// QueryGetApplicationByAppID is the query to retrieve application details by app ID.
	QueryGetApplicationByAppID = dbmodel.DBQuery{
		ID: "ASQ-APP_MGT-03",
		Query: "SELECT sp.APP_ID, sp.APP_NAME, sp.DESCRIPTION, sp.AUTH_FLOW_ID, " +
			"sp.REGISTRATION_FLOW_ID, sp.IS_REGISTRATION_FLOW_ENABLED, sp.THEME_ID, sp.LAYOUT_ID, sp.APP_JSON, " +
			"oauth.CONSUMER_KEY, oauth.CONSUMER_SECRET, oauth.OAUTH_CONFIG_JSON " +
			"FROM SP_APP sp LEFT JOIN IDN_OAUTH_CONSUMER_APPS oauth " +
			"ON sp.APP_ID = oauth.APP_ID AND sp.DEPLOYMENT_ID = $2 AND oauth.DEPLOYMENT_ID = $2 " +
			"WHERE sp.APP_ID = $1 AND sp.DEPLOYMENT_ID = $2",
	}
	// QueryGetApplicationByName is the query to retrieve application details by name.
	QueryGetApplicationByName = dbmodel.DBQuery{
		ID: "ASQ-APP_MGT-04",
		Query: "SELECT sp.APP_ID, sp.APP_NAME, sp.DESCRIPTION, sp.AUTH_FLOW_ID, " +
			"sp.REGISTRATION_FLOW_ID, sp.IS_REGISTRATION_FLOW_ENABLED, sp.THEME_ID, sp.LAYOUT_ID, sp.APP_JSON, " +
			"oauth.CONSUMER_KEY, oauth.CONSUMER_SECRET, oauth.OAUTH_CONFIG_JSON " +
			"FROM SP_APP sp LEFT JOIN IDN_OAUTH_CONSUMER_APPS oauth " +
			"ON sp.APP_ID = oauth.APP_ID AND sp.DEPLOYMENT_ID = $2 AND oauth.DEPLOYMENT_ID = $2 " +
			"WHERE sp.APP_NAME = $1 AND sp.DEPLOYMENT_ID = $2",
	}
	// QueryGetOAuthApplicationByClientID is the query to retrieve oauth application details by client ID.
	QueryGetOAuthApplicationByClientID = dbmodel.DBQuery{
		ID: "ASQ-APP_MGT-05",
		Query: "SELECT APP_ID, CONSUMER_KEY, CONSUMER_SECRET, OAUTH_CONFIG_JSON FROM IDN_OAUTH_CONSUMER_APPS " +
			"WHERE CONSUMER_KEY = $1 AND DEPLOYMENT_ID = $2",
	}
	// QueryGetApplicationList is the query to list all the applications.
	QueryGetApplicationList = dbmodel.DBQuery{
		ID: "ASQ-APP_MGT-06",
		Query: "SELECT sp.APP_ID, sp.APP_NAME, sp.DESCRIPTION, sp.AUTH_FLOW_ID, " +
			"sp.REGISTRATION_FLOW_ID, sp.IS_REGISTRATION_FLOW_ENABLED, sp.THEME_ID, sp.LAYOUT_ID, sp.APP_JSON, " +
			"oauth.CONSUMER_KEY FROM SP_APP sp LEFT JOIN IDN_OAUTH_CONSUMER_APPS oauth ON sp.APP_ID = oauth.APP_ID " +
			"AND sp.DEPLOYMENT_ID = $1 AND oauth.DEPLOYMENT_ID = $1 WHERE sp.DEPLOYMENT_ID = $1",
	}
	// QueryUpdateApplicationByAppID is the query to update application details by app ID.
	QueryUpdateApplicationByAppID = dbmodel.DBQuery{
		ID: "ASQ-APP_MGT-07",
		Query: "UPDATE SP_APP SET APP_NAME=$2, DESCRIPTION=$3, AUTH_FLOW_ID=$4, " +
			"REGISTRATION_FLOW_ID=$5, IS_REGISTRATION_FLOW_ENABLED=$6, THEME_ID=$7, LAYOUT_ID=$8, APP_JSON=$9 " +
			"WHERE APP_ID = $1 AND DEPLOYMENT_ID = $10",
	}
	// QueryUpdateOAuthApplicationByAppID is the query to update OAuth application details by app ID.
	QueryUpdateOAuthApplicationByAppID = dbmodel.DBQuery{
		ID: "ASQ-APP_MGT-08",
		Query: "UPDATE IDN_OAUTH_CONSUMER_APPS SET CONSUMER_KEY=$2, CONSUMER_SECRET=$3, OAUTH_CONFIG_JSON=$4 " +
			"WHERE APP_ID=$1 AND DEPLOYMENT_ID=$5",
	}
	// QueryDeleteApplicationByAppID is the query to delete an application by app ID.
	QueryDeleteApplicationByAppID = dbmodel.DBQuery{
		ID:    "ASQ-APP_MGT-09",
		Query: "DELETE FROM SP_APP WHERE APP_ID = $1 AND DEPLOYMENT_ID = $2",
	}
	// QueryGetApplicationCount is the query to get the total count of applications.
	QueryGetApplicationCount = dbmodel.DBQuery{
		ID:    "ASQ-APP_MGT-10",
		Query: "SELECT COUNT(*) as total FROM SP_APP WHERE DEPLOYMENT_ID = $1",
	}
	// QueryDeleteOAuthApplicationByClientID is the query to delete an OAuth application by client ID.
	QueryDeleteOAuthApplicationByClientID = dbmodel.DBQuery{
		ID:    "ASQ-APP_MGT-11",
		Query: "DELETE FROM IDN_OAUTH_CONSUMER_APPS WHERE CONSUMER_KEY = $1 AND DEPLOYMENT_ID = $2",
	}
)
