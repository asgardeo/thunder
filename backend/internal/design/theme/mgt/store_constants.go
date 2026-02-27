/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

package thememgt

import dbmodel "github.com/asgardeo/thunder/internal/system/database/model"

var (
	// queryCreateTheme creates a new theme configuration.
	queryCreateTheme = dbmodel.DBQuery{
		ID: "THQ-THEME_MGT-01",
		Query: "INSERT INTO THEME (THEME_ID, DISPLAY_NAME, DESCRIPTION, THEME, DEPLOYMENT_ID) " +
			"VALUES ($1, $2, $3, $4, $5)",
	}

	// queryGetThemeByID retrieves a theme configuration by ID.
	queryGetThemeByID = dbmodel.DBQuery{
		ID: "THQ-THEME_MGT-02",
		Query: "SELECT THEME_ID, DISPLAY_NAME, DESCRIPTION, THEME, CREATED_AT, UPDATED_AT FROM THEME " +
			"WHERE THEME_ID = $1 AND DEPLOYMENT_ID = $2",
	}

	// queryGetThemeList retrieves a list of theme configurations with pagination.
	queryGetThemeList = dbmodel.DBQuery{
		ID: "THQ-THEME_MGT-03",
		Query: "SELECT THEME_ID, DISPLAY_NAME, DESCRIPTION, CREATED_AT, UPDATED_AT FROM THEME " +
			"WHERE DEPLOYMENT_ID = $3 ORDER BY CREATED_AT DESC LIMIT $1 OFFSET $2",
	}

	// queryGetThemeListCount retrieves the total count of theme configurations.
	queryGetThemeListCount = dbmodel.DBQuery{
		ID:    "THQ-THEME_MGT-04",
		Query: "SELECT COUNT(*) as total FROM THEME WHERE DEPLOYMENT_ID = $1",
	}

	// queryUpdateTheme updates a theme configuration.
	queryUpdateTheme = dbmodel.DBQuery{
		ID: "THQ-THEME_MGT-05",
		PostgresQuery: "UPDATE THEME SET DISPLAY_NAME = $1, DESCRIPTION = $2, THEME = $3, " +
			"UPDATED_AT = NOW() WHERE THEME_ID = $4 AND DEPLOYMENT_ID = $5",
		SQLiteQuery: "UPDATE THEME SET DISPLAY_NAME = $1, DESCRIPTION = $2, THEME = $3, " +
			"UPDATED_AT = datetime('now') WHERE THEME_ID = $4 AND DEPLOYMENT_ID = $5",
		Query: "UPDATE THEME SET DISPLAY_NAME = $1, DESCRIPTION = $2, THEME = $3, " +
			"UPDATED_AT = datetime('now') WHERE THEME_ID = $4 AND DEPLOYMENT_ID = $5",
	}

	// queryDeleteTheme deletes a theme configuration.
	queryDeleteTheme = dbmodel.DBQuery{
		ID:    "THQ-THEME_MGT-06",
		Query: "DELETE FROM THEME WHERE THEME_ID = $1 AND DEPLOYMENT_ID = $2",
	}

	// queryCheckThemeExists checks if a theme exists.
	queryCheckThemeExists = dbmodel.DBQuery{
		ID:    "THQ-THEME_MGT-07",
		Query: "SELECT COUNT(*) as total FROM THEME WHERE THEME_ID = $1 AND DEPLOYMENT_ID = $2",
	}

	// queryGetApplicationsCountByThemeID retrieves the count of applications using a theme.
	queryGetApplicationsCountByThemeID = dbmodel.DBQuery{
		ID:    "THQ-THEME_MGT-08",
		Query: "SELECT COUNT(*) as total FROM APPLICATION WHERE THEME_ID = $1 AND DEPLOYMENT_ID = $2",
	}
)
