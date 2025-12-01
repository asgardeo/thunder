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

package branding

import dbmodel "github.com/asgardeo/thunder/internal/system/database/model"

var (
	// queryCreateBranding creates a new branding configuration.
	queryCreateBranding = dbmodel.DBQuery{
		ID:    "BRQ-BRANDING_MGT-01",
		Query: "INSERT INTO BRANDING (BRANDING_ID, DISPLAY_NAME, PREFERENCES, DEPLOYMENT_ID) VALUES ($1, $2, $3, $4)",
	}

	// queryGetBrandingByID retrieves a branding configuration by ID.
	queryGetBrandingByID = dbmodel.DBQuery{
		ID:    "BRQ-BRANDING_MGT-02",
		Query: "SELECT BRANDING_ID, DISPLAY_NAME, PREFERENCES FROM BRANDING WHERE BRANDING_ID = $1 AND DEPLOYMENT_ID = $2",
	}

	// queryGetBrandingList retrieves a list of branding configurations with pagination.
	queryGetBrandingList = dbmodel.DBQuery{
		ID: "BRQ-BRANDING_MGT-03",
		Query: "SELECT BRANDING_ID, DISPLAY_NAME FROM BRANDING " +
			"WHERE DEPLOYMENT_ID = $3 ORDER BY CREATED_AT DESC LIMIT $1 OFFSET $2",
	}

	// queryGetBrandingListCount retrieves the total count of branding configurations.
	queryGetBrandingListCount = dbmodel.DBQuery{
		ID:    "BRQ-BRANDING_MGT-04",
		Query: "SELECT COUNT(*) as total FROM BRANDING WHERE DEPLOYMENT_ID = $1",
	}

	// queryUpdateBranding updates a branding configuration.
	queryUpdateBranding = dbmodel.DBQuery{
		ID: "BRQ-BRANDING_MGT-05",
		PostgresQuery: "UPDATE BRANDING SET DISPLAY_NAME = $1, PREFERENCES = $2, " +
			"UPDATED_AT = NOW() WHERE BRANDING_ID = $3 AND DEPLOYMENT_ID = $4",
		SQLiteQuery: "UPDATE BRANDING SET DISPLAY_NAME = $1, PREFERENCES = $2, " +
			"UPDATED_AT = datetime('now') WHERE BRANDING_ID = $3 AND DEPLOYMENT_ID = $4",
		Query: "UPDATE BRANDING SET DISPLAY_NAME = $1, PREFERENCES = $2, " +
			"UPDATED_AT = datetime('now') WHERE BRANDING_ID = $3 AND DEPLOYMENT_ID = $4",
	}

	// queryDeleteBranding deletes a branding configuration.
	queryDeleteBranding = dbmodel.DBQuery{
		ID:    "BRQ-BRANDING_MGT-06",
		Query: "DELETE FROM BRANDING WHERE BRANDING_ID = $1 AND DEPLOYMENT_ID = $2",
	}

	// queryCheckBrandingExists checks if a branding configuration exists by its ID.
	queryCheckBrandingExists = dbmodel.DBQuery{
		ID:    "BRQ-BRANDING_MGT-07",
		Query: "SELECT COUNT(*) as count FROM BRANDING WHERE BRANDING_ID = $1 AND DEPLOYMENT_ID = $2",
	}

	// queryGetApplicationsCountByBrandingID retrieves the count of applications using a branding configuration.
	queryGetApplicationsCountByBrandingID = dbmodel.DBQuery{
		ID:    "BRQ-BRANDING_MGT-08",
		Query: "SELECT COUNT(*) as count FROM SP_APP WHERE BRANDING_ID = $1 AND DEPLOYMENT_ID = $2",
	}
)
