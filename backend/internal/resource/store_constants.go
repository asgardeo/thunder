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

package resource

import (
	dbmodel "github.com/asgardeo/thunder/internal/system/database/model"
)

// Resource Server Queries
var (
	// queryCreateResourceServer creates a new resource server.
	queryCreateResourceServer = dbmodel.DBQuery{
		ID: "RSQ-RES_MGT-01",
		Query: `INSERT INTO RESOURCE_SERVER 
			(RESOURCE_SERVER_ID, OU_ID, NAME, DESCRIPTION, IDENTIFIER, PROPERTIES) 
			VALUES ($1, $2, $3, $4, $5, $6)`,
	}

	// queryGetResourceServerByID retrieves a resource server by ID.
	queryGetResourceServerByID = dbmodel.DBQuery{
		ID: "RSQ-RES_MGT-02",
		Query: `SELECT RESOURCE_SERVER_ID, OU_ID, NAME, DESCRIPTION, IDENTIFIER, PROPERTIES 
			FROM RESOURCE_SERVER 
			WHERE RESOURCE_SERVER_ID = $1`,
	}

	// queryGetResourceServerList retrieves a list of resource servers with pagination.
	queryGetResourceServerList = dbmodel.DBQuery{
		ID: "RSQ-RES_MGT-03",
		Query: `SELECT RESOURCE_SERVER_ID, OU_ID, NAME, DESCRIPTION, IDENTIFIER, PROPERTIES 
			FROM RESOURCE_SERVER 
			ORDER BY CREATED_AT DESC 
			LIMIT $1 OFFSET $2`,
	}

	// queryGetResourceServerListCount retrieves the total count of resource servers.
	queryGetResourceServerListCount = dbmodel.DBQuery{
		ID:    "RSQ-RES_MGT-04",
		Query: `SELECT COUNT(*) as total FROM RESOURCE_SERVER`,
	}

	// queryUpdateResourceServer updates a resource server.
	queryUpdateResourceServer = dbmodel.DBQuery{
		ID: "RSQ-RES_MGT-05",
		Query: `UPDATE RESOURCE_SERVER 
			SET OU_ID = $1, NAME = $2, DESCRIPTION = $3, IDENTIFIER = $4, PROPERTIES = $5 
			WHERE RESOURCE_SERVER_ID = $6`,
	}

	// queryDeleteResourceServer deletes a resource server.
	queryDeleteResourceServer = dbmodel.DBQuery{
		ID:    "RSQ-RES_MGT-06",
		Query: `DELETE FROM RESOURCE_SERVER WHERE RESOURCE_SERVER_ID = $1`,
	}

	// queryCheckResourceServerExists checks if a resource server exists by ID.
	queryCheckResourceServerExists = dbmodel.DBQuery{
		ID:    "RSQ-RES_MGT-07",
		Query: `SELECT COUNT(*) as count FROM RESOURCE_SERVER WHERE RESOURCE_SERVER_ID = $1`,
	}

	// queryCheckResourceServerNameExists checks if a resource server name already exists.
	queryCheckResourceServerNameExists = dbmodel.DBQuery{
		ID:    "RSQ-RES_MGT-08",
		Query: `SELECT COUNT(*) as count FROM RESOURCE_SERVER WHERE NAME = $1`,
	}

	// queryCheckResourceServerNameExistsExcludingID checks name exists excluding a specific ID.
	queryCheckResourceServerNameExistsExcludingID = dbmodel.DBQuery{
		ID:    "RSQ-RES_MGT-09",
		Query: `SELECT COUNT(*) as count FROM RESOURCE_SERVER WHERE NAME = $1 AND RESOURCE_SERVER_ID != $2`,
	}

	// queryCheckResourceServerIdentifierExists checks if a resource server identifier already exists.
	queryCheckResourceServerIdentifierExists = dbmodel.DBQuery{
		ID:    "RSQ-RES_MGT-10",
		Query: `SELECT COUNT(*) as count FROM RESOURCE_SERVER WHERE IDENTIFIER = $1`,
	}

	// queryCheckResourceServerIdentifierExistsExcludingID checks identifier exists excluding a specific ID.
	queryCheckResourceServerIdentifierExistsExcludingID = dbmodel.DBQuery{
		ID:    "RSQ-RES_MGT-11",
		Query: `SELECT COUNT(*) as count FROM RESOURCE_SERVER WHERE IDENTIFIER = $1 AND RESOURCE_SERVER_ID != $2`,
	}

	// queryCheckResourceServerHasDependencies checks if resource server has resources or actions.
	queryCheckResourceServerHasDependencies = dbmodel.DBQuery{
		ID: "RSQ-RES_MGT-12",
		Query: `SELECT COUNT(*) as count FROM (
			SELECT 1 FROM RESOURCE r 
				JOIN RESOURCE_SERVER rs ON r.RESOURCE_SERVER_ID = rs.ID 
				WHERE rs.RESOURCE_SERVER_ID = $1
			UNION ALL
			SELECT 1 FROM ACTION a 
				JOIN RESOURCE_SERVER rs ON a.RESOURCE_SERVER_ID = rs.ID 
				WHERE rs.RESOURCE_SERVER_ID = $1 AND a.RESOURCE_ID IS NULL
		) as dependencies`,
	}

	// queryGetResourceServerInternalID retrieves the internal ID of a resource server by UUID.
	queryGetResourceServerInternalID = dbmodel.DBQuery{
		ID:    "RSQ-RES_MGT-13",
		Query: `SELECT ID FROM RESOURCE_SERVER WHERE RESOURCE_SERVER_ID = $1`,
	}
)

// Resource Queries
var (
	// queryCreateResource creates a new resource.
	queryCreateResource = dbmodel.DBQuery{
		ID: "RSQ-RES_MGT-14",
		Query: `INSERT INTO RESOURCE
		        (RESOURCE_ID, RESOURCE_SERVER_ID, NAME, HANDLE, DESCRIPTION, PROPERTIES, PARENT_RESOURCE_ID)
		        VALUES ($1, $2, $3, $4, $5, $6, $7)`,
	}

	// queryGetResourceByID retrieves a resource by ID.
	queryGetResourceByID = dbmodel.DBQuery{
		ID: "RSQ-RES_MGT-15",
		Query: `SELECT r.RESOURCE_ID, rs.RESOURCE_SERVER_ID, r.NAME, r.HANDLE, r.DESCRIPTION, 
				r.PROPERTIES, pr.RESOURCE_ID as PARENT_RESOURCE_ID
			FROM RESOURCE r
			JOIN RESOURCE_SERVER rs ON r.RESOURCE_SERVER_ID = rs.ID
			LEFT JOIN RESOURCE pr ON r.PARENT_RESOURCE_ID = pr.ID
			WHERE r.RESOURCE_ID = $1 AND rs.RESOURCE_SERVER_ID = $2`,
	}

	// queryGetResourceList retrieves a list of resources with pagination.
	queryGetResourceList = dbmodel.DBQuery{
		ID: "RSQ-RES_MGT-16",
		Query: `SELECT r.RESOURCE_ID, rs.RESOURCE_SERVER_ID, r.NAME, r.HANDLE, r.DESCRIPTION, 
				r.PROPERTIES, pr.RESOURCE_ID as PARENT_RESOURCE_ID
			FROM RESOURCE r
			JOIN RESOURCE_SERVER rs ON r.RESOURCE_SERVER_ID = rs.ID
			LEFT JOIN RESOURCE pr ON r.PARENT_RESOURCE_ID = pr.ID
			WHERE rs.RESOURCE_SERVER_ID = $1
			ORDER BY r.CREATED_AT DESC LIMIT $2 OFFSET $3`,
	}

	// queryGetResourceListByParent retrieves resources by parent ID with pagination.
	queryGetResourceListByParent = dbmodel.DBQuery{
		ID: "RSQ-RES_MGT-17",
		Query: `SELECT r.RESOURCE_ID, rs.RESOURCE_SERVER_ID, r.NAME, r.HANDLE, r.DESCRIPTION, 
				r.PROPERTIES, pr.RESOURCE_ID as PARENT_RESOURCE_ID
			FROM RESOURCE r
			JOIN RESOURCE_SERVER rs ON r.RESOURCE_SERVER_ID = rs.ID
			LEFT JOIN RESOURCE pr ON r.PARENT_RESOURCE_ID = pr.ID
			JOIN RESOURCE parent ON r.PARENT_RESOURCE_ID = parent.ID
			WHERE rs.RESOURCE_SERVER_ID = $1 AND parent.RESOURCE_ID = $2
			ORDER BY r.CREATED_AT DESC LIMIT $3 OFFSET $4`,
	}

	// queryGetResourceListByNullParent retrieves top-level resources with pagination.
	queryGetResourceListByNullParent = dbmodel.DBQuery{
		ID: "RSQ-RES_MGT-18",
		Query: `SELECT r.RESOURCE_ID, rs.RESOURCE_SERVER_ID, r.NAME, r.HANDLE, r.DESCRIPTION, 
				r.PROPERTIES, pr.RESOURCE_ID as PARENT_RESOURCE_ID
			FROM RESOURCE r
			JOIN RESOURCE_SERVER rs ON r.RESOURCE_SERVER_ID = rs.ID
		        LEFT JOIN RESOURCE pr ON r.PARENT_RESOURCE_ID = pr.ID
		        WHERE rs.RESOURCE_SERVER_ID = $1 AND r.PARENT_RESOURCE_ID IS NULL
		        ORDER BY r.CREATED_AT DESC LIMIT $2 OFFSET $3`,
	}

	// queryGetResourceListCount retrieves the total count of resources.
	queryGetResourceListCount = dbmodel.DBQuery{
		ID: "RSQ-RES_MGT-19",
		Query: `SELECT COUNT(*) as total
		        FROM RESOURCE r
		        JOIN RESOURCE_SERVER rs ON r.RESOURCE_SERVER_ID = rs.ID
		        WHERE rs.RESOURCE_SERVER_ID = $1`,
	}

	// queryGetResourceListCountByParent retrieves count of resources by parent.
	queryGetResourceListCountByParent = dbmodel.DBQuery{
		ID: "RSQ-RES_MGT-20",
		Query: `SELECT COUNT(*) as total
		        FROM RESOURCE r
		        JOIN RESOURCE_SERVER rs ON r.RESOURCE_SERVER_ID = rs.ID
		        JOIN RESOURCE parent ON r.PARENT_RESOURCE_ID = parent.ID
		        WHERE rs.RESOURCE_SERVER_ID = $1 AND parent.RESOURCE_ID = $2`,
	}

	// queryGetResourceListCountByNullParent retrieves count of top-level resources.
	queryGetResourceListCountByNullParent = dbmodel.DBQuery{
		ID: "RSQ-RES_MGT-21",
		Query: `SELECT COUNT(*) as total
		        FROM RESOURCE r
		        JOIN RESOURCE_SERVER rs ON r.RESOURCE_SERVER_ID = rs.ID
		        WHERE rs.RESOURCE_SERVER_ID = $1 AND r.PARENT_RESOURCE_ID IS NULL`,
	}

	// queryUpdateResource updates a resource.
	queryUpdateResource = dbmodel.DBQuery{
		ID: "RSQ-RES_MGT-22",
		Query: `UPDATE RESOURCE
		        SET NAME = $1, 
				    DESCRIPTION = $2, 
		            PROPERTIES = $3 
		        WHERE ID = (
		            SELECT r.ID FROM RESOURCE r 
		            JOIN RESOURCE_SERVER rs ON r.RESOURCE_SERVER_ID = rs.ID 
		            WHERE r.RESOURCE_ID = $4 AND rs.RESOURCE_SERVER_ID = $5
		        )`,
	}

	// queryDeleteResource deletes a resource.
	queryDeleteResource = dbmodel.DBQuery{
		ID: "RSQ-RES_MGT-23",
		Query: `DELETE FROM RESOURCE
		        WHERE ID = (
		            SELECT r.ID FROM RESOURCE r
		            JOIN RESOURCE_SERVER rs ON r.RESOURCE_SERVER_ID = rs.ID
		            WHERE r.RESOURCE_ID = $1 AND rs.RESOURCE_SERVER_ID = $2
		        )`,
	}

	// queryCheckResourceExists checks if a resource exists by ID.
	queryCheckResourceExists = dbmodel.DBQuery{
		ID: "RSQ-RES_MGT-24",
		Query: `SELECT COUNT(*) as count
		        FROM RESOURCE r
		        JOIN RESOURCE_SERVER rs ON r.RESOURCE_SERVER_ID = rs.ID
		        WHERE r.RESOURCE_ID = $1 AND rs.RESOURCE_SERVER_ID = $2`,
	}

	// queryCheckResourceHandleExistsUnderParent checks if resource handle exists under same parent.
	queryCheckResourceHandleExistsUnderParent = dbmodel.DBQuery{
		ID: "RSQ-RES_MGT-25",
		Query: `SELECT COUNT(*) as count
		        FROM RESOURCE r
		        JOIN RESOURCE_SERVER rs ON r.RESOURCE_SERVER_ID = rs.ID
		        JOIN RESOURCE parent ON r.PARENT_RESOURCE_ID = parent.ID
		        WHERE rs.RESOURCE_SERVER_ID = $1 AND r.HANDLE = $2 AND parent.RESOURCE_ID = $3`,
	}

	// queryCheckResourceHandleExistsUnderNullParent checks if resource handle exists at top level.
	queryCheckResourceHandleExistsUnderNullParent = dbmodel.DBQuery{
		ID: "RSQ-RES_MGT-26",
		Query: `SELECT COUNT(*) as count
		        FROM RESOURCE r
		        JOIN RESOURCE_SERVER rs ON r.RESOURCE_SERVER_ID = rs.ID
		        WHERE rs.RESOURCE_SERVER_ID = $1 AND r.HANDLE = $2 AND r.PARENT_RESOURCE_ID IS NULL`,
	}

	// queryCheckResourceHasDependencies checks if resource has sub-resources or actions.
	queryCheckResourceHasDependencies = dbmodel.DBQuery{
		ID: "RSQ-RES_MGT-29",
		Query: `SELECT COUNT(*) as count FROM (
			SELECT 1 FROM RESOURCE child 
				JOIN RESOURCE parent ON child.PARENT_RESOURCE_ID = parent.ID 
				WHERE parent.RESOURCE_ID = $1
			UNION ALL
			SELECT 1 FROM ACTION a 
				JOIN RESOURCE r ON a.RESOURCE_ID = r.ID 
				WHERE r.RESOURCE_ID = $1
		) as dependencies`,
	}

	// queryCheckCircularDependency checks if setting a parent would create a circular dependency.
	// It traverses UP the parent chain from newParentID to check if resourceID appears as an ancestor.
	queryCheckCircularDependency = dbmodel.DBQuery{
		ID: "RSQ-RES_MGT-30",
		Query: `WITH RECURSIVE parent_hierarchy AS (
			SELECT ID, RESOURCE_ID, PARENT_RESOURCE_ID FROM RESOURCE WHERE RESOURCE_ID = $1
			UNION ALL
			SELECT r.ID, r.RESOURCE_ID, r.PARENT_RESOURCE_ID
			FROM RESOURCE r
			INNER JOIN parent_hierarchy ph ON ph.PARENT_RESOURCE_ID = r.ID
		)
		SELECT COUNT(*) as count FROM parent_hierarchy WHERE RESOURCE_ID = $2`,
	}

	// queryGetResourceInternalID retrieves the internal ID of a resource by UUID.
	queryGetResourceInternalID = dbmodel.DBQuery{
		ID: "RSQ-RES_MGT-31",
		Query: `SELECT r.ID FROM RESOURCE r
		        JOIN RESOURCE_SERVER rs ON r.RESOURCE_SERVER_ID = rs.ID
		        WHERE r.RESOURCE_ID = $1 AND rs.RESOURCE_SERVER_ID = $2`,
	}
)

// Action Queries
var (
	// queryCreateAction creates a new action.
	queryCreateAction = dbmodel.DBQuery{
		ID: "RSQ-RES_MGT-32",
		Query: `INSERT INTO ACTION (ACTION_ID, RESOURCE_SERVER_ID, RESOURCE_ID, NAME, HANDLE, DESCRIPTION, PROPERTIES)
		        VALUES ($1, $2, $3, $4, $5, $6, $7)`,
	}

	// queryGetActionByID retrieves an action by ID (resource server level).
	queryGetActionByID = dbmodel.DBQuery{
		ID: "RSQ-RES_MGT-33",
		Query: `SELECT a.ACTION_ID, rs.RESOURCE_SERVER_ID, r.RESOURCE_ID, a.NAME, a.HANDLE, a.DESCRIPTION, a.PROPERTIES
		        FROM ACTION a
		        JOIN RESOURCE_SERVER rs ON a.RESOURCE_SERVER_ID = rs.ID
		        LEFT JOIN RESOURCE r ON a.RESOURCE_ID = r.ID
		        WHERE a.ACTION_ID = $1 AND rs.RESOURCE_SERVER_ID = $2 AND a.RESOURCE_ID IS NULL`,
	}

	// queryGetActionByIDAtResource retrieves an action by ID (resource level).
	queryGetActionByIDAtResource = dbmodel.DBQuery{
		ID: "RSQ-RES_MGT-34",
		Query: `SELECT a.ACTION_ID, rs.RESOURCE_SERVER_ID, r.RESOURCE_ID, a.NAME, a.HANDLE, a.DESCRIPTION, a.PROPERTIES
		        FROM ACTION a
		        JOIN RESOURCE_SERVER rs ON a.RESOURCE_SERVER_ID = rs.ID
		        JOIN RESOURCE r ON a.RESOURCE_ID = r.ID
		        WHERE a.ACTION_ID = $1 AND rs.RESOURCE_SERVER_ID = $2 AND r.RESOURCE_ID = $3`,
	}

	// queryGetActionListAtResourceServer retrieves actions at resource server level with pagination.
	queryGetActionListAtResourceServer = dbmodel.DBQuery{
		ID: "RSQ-RES_MGT-35",
		Query: `SELECT a.ACTION_ID, rs.RESOURCE_SERVER_ID, r.RESOURCE_ID, a.NAME, a.HANDLE, a.DESCRIPTION, a.PROPERTIES
		        FROM ACTION a
		        JOIN RESOURCE_SERVER rs ON a.RESOURCE_SERVER_ID = rs.ID
		        LEFT JOIN RESOURCE r ON a.RESOURCE_ID = r.ID
		        WHERE rs.RESOURCE_SERVER_ID = $1 AND a.RESOURCE_ID IS NULL
		        ORDER BY a.CREATED_AT DESC LIMIT $2 OFFSET $3`,
	}

	// queryGetActionListAtResource retrieves actions at resource level with pagination.
	queryGetActionListAtResource = dbmodel.DBQuery{
		ID: "RSQ-RES_MGT-36",
		Query: `SELECT a.ACTION_ID, rs.RESOURCE_SERVER_ID, r.RESOURCE_ID, a.NAME, a.HANDLE, a.DESCRIPTION, a.PROPERTIES
		        FROM ACTION a
		        JOIN RESOURCE_SERVER rs ON a.RESOURCE_SERVER_ID = rs.ID
		        JOIN RESOURCE r ON a.RESOURCE_ID = r.ID
		        WHERE rs.RESOURCE_SERVER_ID = $1 AND r.RESOURCE_ID = $2
		        ORDER BY a.CREATED_AT DESC LIMIT $3 OFFSET $4`,
	}

	// queryGetActionListCountAtResourceServer retrieves count of actions at resource server level.
	queryGetActionListCountAtResourceServer = dbmodel.DBQuery{
		ID: "RSQ-RES_MGT-37",
		Query: `SELECT COUNT(*) as total
		        FROM ACTION a
		        JOIN RESOURCE_SERVER rs ON a.RESOURCE_SERVER_ID = rs.ID
		        WHERE rs.RESOURCE_SERVER_ID = $1 AND a.RESOURCE_ID IS NULL`,
	}

	// queryGetActionListCountAtResource retrieves count of actions at resource level.
	queryGetActionListCountAtResource = dbmodel.DBQuery{
		ID: "RSQ-RES_MGT-38",
		Query: `SELECT COUNT(*) as total
		        FROM ACTION a
		        JOIN RESOURCE_SERVER rs ON a.RESOURCE_SERVER_ID = rs.ID
		        JOIN RESOURCE r ON a.RESOURCE_ID = r.ID
		        WHERE rs.RESOURCE_SERVER_ID = $1 AND r.RESOURCE_ID = $2`,
	}

	// queryUpdateAction updates an action at resource server level.
	queryUpdateAction = dbmodel.DBQuery{
		ID: "RSQ-RES_MGT-39",
		Query: `UPDATE ACTION
		        SET NAME = $1, DESCRIPTION = $2, PROPERTIES = $3
		        WHERE ID = (
		            SELECT a.ID FROM ACTION a
		            JOIN RESOURCE_SERVER rs ON a.RESOURCE_SERVER_ID = rs.ID
		            WHERE a.ACTION_ID = $4 AND rs.RESOURCE_SERVER_ID = $5
		              AND a.RESOURCE_ID IS NULL
		        )`,
	}

	// queryUpdateActionAtResource updates an action at resource level.
	queryUpdateActionAtResource = dbmodel.DBQuery{
		ID: "RSQ-RES_MGT-40",
		Query: `UPDATE ACTION
		        SET NAME = $1, DESCRIPTION = $2, PROPERTIES = $3
		        WHERE ID = (
		            SELECT a.ID FROM ACTION a
		            JOIN RESOURCE_SERVER rs ON a.RESOURCE_SERVER_ID = rs.ID
		            JOIN RESOURCE r ON a.RESOURCE_ID = r.ID
		            WHERE a.ACTION_ID = $4 AND rs.RESOURCE_SERVER_ID = $5
		              AND r.RESOURCE_ID = $6
		        )`,
	}

	// queryDeleteAction deletes an action at resource server level.
	queryDeleteAction = dbmodel.DBQuery{
		ID: "RSQ-RES_MGT-41",
		Query: `DELETE FROM ACTION
		        WHERE ID = (
		            SELECT a.ID FROM ACTION a
		            JOIN RESOURCE_SERVER rs ON a.RESOURCE_SERVER_ID = rs.ID
		            WHERE a.ACTION_ID = $1 AND rs.RESOURCE_SERVER_ID = $2
		              AND a.RESOURCE_ID IS NULL
		        )`,
	}

	// queryDeleteActionAtResource deletes an action at resource level.
	queryDeleteActionAtResource = dbmodel.DBQuery{
		ID: "RSQ-RES_MGT-42",
		Query: `DELETE FROM ACTION
		        WHERE ID = (
		            SELECT a.ID FROM ACTION a
		            JOIN RESOURCE_SERVER rs ON a.RESOURCE_SERVER_ID = rs.ID
		            JOIN RESOURCE r ON a.RESOURCE_ID = r.ID
		            WHERE a.ACTION_ID = $1 AND rs.RESOURCE_SERVER_ID = $2
		              AND r.RESOURCE_ID = $3
		        )`,
	}

	// queryCheckActionExists checks if an action exists by ID at resource server level.
	queryCheckActionExists = dbmodel.DBQuery{
		ID: "RSQ-RES_MGT-43",
		Query: `SELECT COUNT(*) as count
		        FROM ACTION a
		        JOIN RESOURCE_SERVER rs ON a.RESOURCE_SERVER_ID = rs.ID
		        WHERE a.ACTION_ID = $1 AND rs.RESOURCE_SERVER_ID = $2 AND a.RESOURCE_ID IS NULL`,
	}

	// queryCheckActionExistsAtResource checks if an action exists by ID at resource level.
	queryCheckActionExistsAtResource = dbmodel.DBQuery{
		ID: "RSQ-RES_MGT-44",
		Query: `SELECT COUNT(*) as count
		        FROM ACTION a
		        JOIN RESOURCE_SERVER rs ON a.RESOURCE_SERVER_ID = rs.ID
		        JOIN RESOURCE r ON a.RESOURCE_ID = r.ID
		        WHERE a.ACTION_ID = $1 AND rs.RESOURCE_SERVER_ID = $2 AND r.RESOURCE_ID = $3`,
	}

	// queryCheckActionHandleExistsAtResourceServer checks if action handle exists at resource server level.
	queryCheckActionHandleExistsAtResourceServer = dbmodel.DBQuery{
		ID: "RSQ-RES_MGT-45",
		Query: `SELECT COUNT(*) as count
		        FROM ACTION a
		        JOIN RESOURCE_SERVER rs ON a.RESOURCE_SERVER_ID = rs.ID
		        WHERE rs.RESOURCE_SERVER_ID = $1 AND a.HANDLE = $2 AND a.RESOURCE_ID IS NULL`,
	}

	// queryCheckActionHandleExistsAtResource checks if action handle exists at resource level.
	queryCheckActionHandleExistsAtResource = dbmodel.DBQuery{
		ID: "RSQ-RES_MGT-46",
		Query: `SELECT COUNT(*) as count
		        FROM ACTION a
		        JOIN RESOURCE_SERVER rs ON a.RESOURCE_SERVER_ID = rs.ID
		        JOIN RESOURCE r ON a.RESOURCE_ID = r.ID
		        WHERE rs.RESOURCE_SERVER_ID = $1 AND r.RESOURCE_ID = $2 AND a.HANDLE = $3`,
	}
)
