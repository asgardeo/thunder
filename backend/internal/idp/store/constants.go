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

package store

import "github.com/asgardeo/thunder/internal/system/database/model"

var (
	// QueryCreateIdentityProvider is the query to create a new IdP.
	QueryCreateIdentityProvider = model.DBQuery{
		ID:    "IPQ-IDP_MGT-01",
		Query: "INSERT INTO IDP (IDP_ID, NAME, DESCRIPTION) VALUES ($1, $2, $3)",
	}
	// QueryGetIdentityProviderByID is the query to get a IdP by IdP ID.
	QueryGetIdentityProviderByID = model.DBQuery{
		ID:    "IPQ-IDP_MGT-02",
		Query: "SELECT IDP_ID, NAME, DESCRIPTION FROM IDP WHERE IDP_ID = $1",
	}
	// QueryGetIdentityProviderList is the query to get a list of IdPs.
	QueryGetIdentityProviderList = model.DBQuery{
		ID:    "IPQ-IDP_MGT-03",
		Query: "SELECT IDP_ID, NAME, DESCRIPTION FROM IDP",
	}
	// QueryUpdateIdentityProviderByID is the query to update a IdP by IdP ID.
	QueryUpdateIdentityProviderByID = model.DBQuery{
		ID:    "IPQ-IDP_MGT-04",
		Query: "UPDATE IDP SET NAME = $2, DESCRIPTION = $3 WHERE IDP_ID = $1",
	}
	// QueryDeleteIdentityProviderByID is the query to delete a IdP by IdP ID.
	QueryDeleteIdentityProviderByID = model.DBQuery{
		ID:    "IPQ-IDP_MGT-05",
		Query: "DELETE FROM IDP WHERE IDP_ID = $1",
	}
	// QueryGetIdentityProviderByName is the query to get a IdP by IdP name.
	QueryGetIdentityProviderByName = model.DBQuery{
		ID:    "IPQ-IDP_MGT-06",
		Query: "SELECT IDP_ID, NAME, DESCRIPTION FROM IDP WHERE NAME = $1",
	}
	// QueryInsertIDPProperties is the query to insert properties for a specific IdP.
	QueryInsertIDPProperties = model.DBQuery{
		ID:    "IPQ-IDP_MGT-07",
		Query: "INSERT INTO IDP_PROPERTY (IDP_ID, PROPERTY_NAME, PROPERTY_VALUE, IS_SECRET, IS_ENCRYPTED) VALUES %s",
	}
	// QueryGetIDPProperties is the query to get properties for a specific IdP.
	QueryGetIDPProperties = model.DBQuery{
		ID:    "IPQ-IDP_MGT-08",
		Query: "SELECT PROPERTY_NAME, PROPERTY_VALUE, IS_SECRET, IS_ENCRYPTED FROM IDP_PROPERTY WHERE IDP_ID = $1",
	}
	// QueryDeleteIDPProperties is the query to delete all properties for a specific IdP.
	QueryDeleteIDPProperties = model.DBQuery{
		ID:    "IPQ-IDP_MGT-10",
		Query: "DELETE FROM IDP_PROPERTY WHERE IDP_ID = $1",
	}
)
