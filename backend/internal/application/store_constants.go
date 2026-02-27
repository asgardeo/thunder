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
	// queryCreateApplication is the query to create a new application with basic details.
	queryCreateApplication = dbmodel.DBQuery{
		ID: "ASQ-APP_MGT-01",
		Query: "INSERT INTO APPLICATION (APP_ID, APP_NAME, DESCRIPTION, AUTH_FLOW_ID, " +
			"REGISTRATION_FLOW_ID, IS_REGISTRATION_FLOW_ENABLED, THEME_ID, LAYOUT_ID, APP_JSON, DEPLOYMENT_ID) " +
			"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)",
	}
	// queryCreateOAuthApplication is the query to create a new OAuth application.
	queryCreateOAuthApplication = dbmodel.DBQuery{
		ID: "ASQ-APP_MGT-02",
		Query: "INSERT INTO APP_OAUTH_INBOUND_CONFIG (APP_ID, CLIENT_ID, CLIENT_SECRET, OAUTH_CONFIG_JSON, " +
			"DEPLOYMENT_ID) VALUES ($1, $2, $3, $4, $5)",
	}
	// queryGetApplicationByAppID is the query to retrieve application details by app ID.
	queryGetApplicationByAppID = dbmodel.DBQuery{
		ID: "ASQ-APP_MGT-03",
		Query: "SELECT app.APP_ID, app.APP_NAME, app.DESCRIPTION, app.AUTH_FLOW_ID, " +
			"app.REGISTRATION_FLOW_ID, app.IS_REGISTRATION_FLOW_ENABLED, app.THEME_ID, app.LAYOUT_ID, app.APP_JSON, " +
			"oauth.CLIENT_ID, oauth.CLIENT_SECRET, oauth.OAUTH_CONFIG_JSON " +
			"FROM APPLICATION app LEFT JOIN APP_OAUTH_INBOUND_CONFIG oauth " +
			"ON app.APP_ID = oauth.APP_ID AND app.DEPLOYMENT_ID = $2 AND oauth.DEPLOYMENT_ID = $2 " +
			"WHERE app.APP_ID = $1 AND app.DEPLOYMENT_ID = $2",
	}
	// queryGetApplicationByName is the query to retrieve application details by name.
	queryGetApplicationByName = dbmodel.DBQuery{
		ID: "ASQ-APP_MGT-04",
		Query: "SELECT app.APP_ID, app.APP_NAME, app.DESCRIPTION, app.AUTH_FLOW_ID, " +
			"app.REGISTRATION_FLOW_ID, app.IS_REGISTRATION_FLOW_ENABLED, app.THEME_ID, app.LAYOUT_ID, app.APP_JSON, " +
			"oauth.CLIENT_ID, oauth.CLIENT_SECRET, oauth.OAUTH_CONFIG_JSON " +
			"FROM APPLICATION app LEFT JOIN APP_OAUTH_INBOUND_CONFIG oauth " +
			"ON app.APP_ID = oauth.APP_ID AND app.DEPLOYMENT_ID = $2 AND oauth.DEPLOYMENT_ID = $2 " +
			"WHERE app.APP_NAME = $1 AND app.DEPLOYMENT_ID = $2",
	}
	// queryGetOAuthApplicationByClientID is the query to retrieve oauth application details by client ID.
	queryGetOAuthApplicationByClientID = dbmodel.DBQuery{
		ID: "ASQ-APP_MGT-05",
		Query: "SELECT APP_ID, CLIENT_ID, CLIENT_SECRET, OAUTH_CONFIG_JSON FROM APP_OAUTH_INBOUND_CONFIG " +
			"WHERE CLIENT_ID = $1 AND DEPLOYMENT_ID = $2",
	}
	// queryGetApplicationList is the query to list all the applications.
	queryGetApplicationList = dbmodel.DBQuery{
		ID: "ASQ-APP_MGT-06",
		Query: "SELECT app.APP_ID, app.APP_NAME, app.DESCRIPTION, app.AUTH_FLOW_ID, " +
			"app.REGISTRATION_FLOW_ID, app.IS_REGISTRATION_FLOW_ENABLED, app.THEME_ID, app.LAYOUT_ID, app.APP_JSON, " +
			"oauth.CLIENT_ID FROM APPLICATION app " +
			"LEFT JOIN APP_OAUTH_INBOUND_CONFIG oauth ON app.APP_ID = oauth.APP_ID " +
			"AND app.DEPLOYMENT_ID = $1 AND oauth.DEPLOYMENT_ID = $1 WHERE app.DEPLOYMENT_ID = $1",
	}
	// queryUpdateApplicationByAppID is the query to update application details by app ID.
	queryUpdateApplicationByAppID = dbmodel.DBQuery{
		ID: "ASQ-APP_MGT-07",
		Query: "UPDATE APPLICATION SET APP_NAME=$2, DESCRIPTION=$3, AUTH_FLOW_ID=$4, " +
			"REGISTRATION_FLOW_ID=$5, IS_REGISTRATION_FLOW_ENABLED=$6, THEME_ID=$7, LAYOUT_ID=$8, APP_JSON=$9 " +
			"WHERE APP_ID = $1 AND DEPLOYMENT_ID = $10",
	}
	// queryUpdateOAuthApplicationByAppID is the query to update OAuth application details by app ID.
	queryUpdateOAuthApplicationByAppID = dbmodel.DBQuery{
		ID: "ASQ-APP_MGT-08",
		Query: "UPDATE APP_OAUTH_INBOUND_CONFIG SET CLIENT_ID=$2, CLIENT_SECRET=$3, OAUTH_CONFIG_JSON=$4 " +
			"WHERE APP_ID=$1 AND DEPLOYMENT_ID=$5",
	}
	// queryDeleteApplicationByAppID is the query to delete an application by app ID.
	queryDeleteApplicationByAppID = dbmodel.DBQuery{
		ID:    "ASQ-APP_MGT-09",
		Query: "DELETE FROM APPLICATION WHERE APP_ID = $1 AND DEPLOYMENT_ID = $2",
	}
	// queryGetApplicationCount is the query to get the total count of applications.
	queryGetApplicationCount = dbmodel.DBQuery{
		ID:    "ASQ-APP_MGT-10",
		Query: "SELECT COUNT(*) as total FROM APPLICATION WHERE DEPLOYMENT_ID = $1",
	}
	// queryDeleteOAuthApplicationByClientID is the query to delete an OAuth application by client ID.
	queryDeleteOAuthApplicationByClientID = dbmodel.DBQuery{
		ID:    "ASQ-APP_MGT-11",
		Query: "DELETE FROM APP_OAUTH_INBOUND_CONFIG WHERE CLIENT_ID = $1 AND DEPLOYMENT_ID = $2",
	}
	// queryCheckApplicationExistsByID is the query to check if an application exists by app ID.
	queryCheckApplicationExistsByID = dbmodel.DBQuery{
		ID:    "ASQ-APP_MGT-12",
		Query: "SELECT COUNT(*) as count FROM APPLICATION WHERE APP_ID = $1 AND DEPLOYMENT_ID = $2",
	}
	// queryCheckApplicationExistsByName is the query to check if an application exists by name.
	queryCheckApplicationExistsByName = dbmodel.DBQuery{
		ID:    "ASQ-APP_MGT-13",
		Query: "SELECT COUNT(*) as count FROM APPLICATION WHERE APP_NAME = $1 AND DEPLOYMENT_ID = $2",
	}
)
