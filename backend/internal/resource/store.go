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
	"encoding/json"
	"fmt"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/database/provider"
)

// resourceStoreInterface defines the interface for resource store operations.
type resourceStoreInterface interface {
	// Resource Server operations
	CreateResourceServer(id string, rs ResourceServer) error
	GetResourceServer(id string) (int, ResourceServer, error)
	GetResourceServerList(limit, offset int) ([]ResourceServer, error)
	GetResourceServerListCount() (int, error)
	UpdateResourceServer(id string, rs ResourceServer) error
	DeleteResourceServer(id string) error
	CheckResourceServerNameExists(name string) (bool, error)
	CheckResourceServerIdentifierExists(identifier string) (bool, error)
	CheckResourceServerHasDependencies(resServerInternalID int) (bool, error)

	// Resource operations
	CreateResource(uuid string, resServerInternalID int, parentInternalID *int, res Resource) error
	GetResource(id string, resServerInternalID int) (int, Resource, error)
	GetResourceList(resServerInternalID int, limit, offset int) ([]Resource, error)
	GetResourceListByParent(resServerInternalID int, parentInternalID *int, limit, offset int) ([]Resource, error)
	GetResourceListCount(resServerInternalID int) (int, error)
	GetResourceListCountByParent(resServerInternalID int, parentInternalID *int) (int, error)
	UpdateResource(id string, resServerInternalID int, res Resource) error
	DeleteResource(id string, resServerInternalID int) error
	CheckResourceHandleExists(resServerInternalID int, handle string, parentInternalID *int) (bool, error)
	CheckResourceHasDependencies(resInternalID int) (bool, error)
	CheckCircularDependency(resourceID, newParentID string) (bool, error)

	// Action operations
	CreateAction(uuid string, resServerInternalID int, resInternalID *int, action Action) error
	GetAction(id string, resServerInternalID int, resInternalID *int) (Action, error)
	GetActionList(resServerInternalID int, resInternalID *int, limit, offset int) ([]Action, error)
	GetActionListCount(resServerInternalID int, resInternalID *int) (int, error)
	UpdateAction(id string, resServerInternalID int, resInternalID *int, action Action) error
	DeleteAction(id string, resServerInternalID int, resInternalID *int) error
	IsActionExist(id string, resServerInternalID int, resInternalID *int) (bool, error)
	CheckActionHandleExists(resServerInternalID int, resInternalID *int, handle string) (bool, error)
	ValidatePermissions(resServerInternalID int, permissions []string) ([]string, error)
}

// resourceStore is the default implementation of resourceStoreInterface.
type resourceStore struct {
	dbProvider   provider.DBProviderInterface
	deploymentID string
}

// resourceServerProperties represents the JSON structure of PROPERTIES column.
type resourceServerProperties struct {
	Delimiter string `json:"delimiter"`
}

// newResourceStore creates a new instance of resourceStore.
func newResourceStore() resourceStoreInterface {
	return &resourceStore{
		dbProvider:   provider.GetDBProvider(),
		deploymentID: config.GetThunderRuntime().Config.Server.Identifier,
	}
}

// CreateResourceServer creates a new resource server in the database.
func (s *resourceStore) CreateResourceServer(id string, rs ResourceServer) error {
	return s.withDBClient(func(dbClient provider.DBClientInterface) error {
		_, err := dbClient.Execute(
			queryCreateResourceServer,
			id,
			rs.OrganizationUnitID,
			rs.Name,
			rs.Description,
			resolveIdentifier(rs.Identifier),
			buildPropertiesJSON(rs),
			s.deploymentID,
		)
		if err != nil {
			return fmt.Errorf("failed to create resource server: %w", err)
		}

		return nil
	})
}

// GetResourceServer retrieves a resource server and internal ID by UUID.
func (s *resourceStore) GetResourceServer(id string) (int, ResourceServer, error) {
	var rs ResourceServer
	var internalID int
	err := s.withDBClient(func(dbClient provider.DBClientInterface) error {
		results, err := dbClient.Query(queryGetResourceServerByID, id, s.deploymentID)
		if err != nil {
			return fmt.Errorf("failed to get resource server: %w", err)
		}

		if len(results) == 0 {
			return errResourceServerNotFound
		}

		internalID, rs, err = buildResourceServerFromResultRow(results[0])
		return err
	})
	return internalID, rs, err
}

// GetResourceServerList retrieves a list of resource servers with pagination.
func (s *resourceStore) GetResourceServerList(limit, offset int) ([]ResourceServer, error) {
	var resourceServers []ResourceServer
	err := s.withDBClient(func(dbClient provider.DBClientInterface) error {
		results, err := dbClient.Query(queryGetResourceServerList, limit, offset, s.deploymentID)
		if err != nil {
			return fmt.Errorf("failed to get resource server list: %w", err)
		}

		resourceServers = make([]ResourceServer, 0, len(results))
		for _, row := range results {
			_, rs, err := buildResourceServerFromResultRow(row)
			if err != nil {
				return fmt.Errorf("failed to build resource server: %w", err)
			}
			resourceServers = append(resourceServers, rs)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return resourceServers, nil
}

// GetResourceServerListCount retrieves the total count of resource servers.
func (s *resourceStore) GetResourceServerListCount() (int, error) {
	var count int
	err := s.withDBClient(func(dbClient provider.DBClientInterface) error {
		results, err := dbClient.Query(queryGetResourceServerListCount, s.deploymentID)
		if err != nil {
			return fmt.Errorf("failed to get resource server count: %w", err)
		}

		count, err = parseCountResult(results)
		return err
	})
	return count, err
}

// UpdateResourceServer updates a resource server.
func (s *resourceStore) UpdateResourceServer(id string, rs ResourceServer) error {
	return s.withDBClient(func(dbClient provider.DBClientInterface) error {
		_, err := dbClient.Execute(
			queryUpdateResourceServer,
			rs.OrganizationUnitID,
			rs.Name,
			rs.Description,
			resolveIdentifier(rs.Identifier),
			buildPropertiesJSON(rs),
			id,
			s.deploymentID,
		)
		if err != nil {
			return fmt.Errorf("failed to update resource server: %w", err)
		}

		return nil
	})
}

// DeleteResourceServer deletes a resource server.
func (s *resourceStore) DeleteResourceServer(id string) error {
	return s.withDBClient(func(dbClient provider.DBClientInterface) error {
		_, err := dbClient.Execute(queryDeleteResourceServer, id, s.deploymentID)
		if err != nil {
			return fmt.Errorf("failed to delete resource server: %w", err)
		}

		return nil
	})
}

// CheckResourceServerNameExists checks if a resource server name exists.
func (s *resourceStore) CheckResourceServerNameExists(name string) (bool, error) {
	var exists bool
	err := s.withDBClient(func(dbClient provider.DBClientInterface) error {
		results, err := dbClient.Query(queryCheckResourceServerNameExists, name, s.deploymentID)
		if err != nil {
			return fmt.Errorf("failed to check resource server name: %w", err)
		}

		exists, err = parseBoolFromCount(results)
		return err
	})
	return exists, err
}

// CheckResourceServerIdentifierExists checks if a resource server identifier exists.
func (s *resourceStore) CheckResourceServerIdentifierExists(identifier string) (bool, error) {
	var exists bool
	err := s.withDBClient(func(dbClient provider.DBClientInterface) error {
		results, err := dbClient.Query(queryCheckResourceServerIdentifierExists, identifier, s.deploymentID)
		if err != nil {
			return fmt.Errorf("failed to check resource server identifier: %w", err)
		}

		exists, err = parseBoolFromCount(results)
		return err
	})
	return exists, err
}

// CheckResourceServerHasDependencies checks if resource server has dependencies.
func (s *resourceStore) CheckResourceServerHasDependencies(resServerInternalID int) (bool, error) {
	var hasDeps bool
	err := s.withDBClient(func(dbClient provider.DBClientInterface) error {
		results, err := dbClient.Query(queryCheckResourceServerHasDependencies, resServerInternalID, s.deploymentID)
		if err != nil {
			return fmt.Errorf("failed to check dependencies: %w", err)
		}

		hasDeps, err = parseBoolFromCount(results)
		return err
	})
	return hasDeps, err
}

// Resource Store Methods

// CreateResource creates a new resource.
func (s *resourceStore) CreateResource(
	uuid string,
	resServerInternalID int,
	parentInternalID *int,
	res Resource,
) error {
	return s.withDBClient(func(dbClient provider.DBClientInterface) error {
		_, err := dbClient.Execute(
			queryCreateResource,
			uuid,                // $1: RESOURCE_ID (UUID)
			resServerInternalID, // $2: RESOURCE_SERVER_ID (int FK)
			res.Name,            // $3: NAME
			res.Handle,          // $4: HANDLE
			res.Description,     // $5: DESCRIPTION
			res.Permission,      // $6: PERMISSION
			"{}",                // $7: PROPERTIES (empty JSON).
			parentInternalID,    // $8: PARENT_RESOURCE_ID (int FK or NULL)
			s.deploymentID,      // $9: DEPLOYMENT_ID
		)
		if err != nil {
			return fmt.Errorf("failed to create resource: %w", err)
		}

		return nil
	})
}

// GetResource retrieves a resource and internal ID by UUID.
func (s *resourceStore) GetResource(id string, resServerInternalID int) (int, Resource, error) {
	var res Resource
	var internalID int
	err := s.withDBClient(func(dbClient provider.DBClientInterface) error {
		results, err := dbClient.Query(queryGetResourceByID, id, resServerInternalID, s.deploymentID)
		if err != nil {
			return fmt.Errorf("failed to get resource: %w", err)
		}

		if len(results) == 0 {
			return errResourceNotFound
		}

		internalID, res, err = buildResourceFromResultRow(results[0])
		return err
	})
	return internalID, res, err
}

// GetResourceList retrieves all resources for a resource server.
func (s *resourceStore) GetResourceList(resServerInternalID int, limit, offset int) ([]Resource, error) {
	var resources []Resource
	err := s.withDBClient(func(dbClient provider.DBClientInterface) error {
		results, err := dbClient.Query(queryGetResourceList, resServerInternalID, limit, offset, s.deploymentID)
		if err != nil {
			return fmt.Errorf("failed to get resource list: %w", err)
		}

		resources = make([]Resource, 0, len(results))
		for _, row := range results {
			_, res, err := buildResourceFromResultRow(row)
			if err != nil {
				return fmt.Errorf("failed to build resource: %w", err)
			}
			resources = append(resources, res)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return resources, nil
}

// GetResourceListByParent retrieves resources filtered by parent.
func (s *resourceStore) GetResourceListByParent(
	resServerInternalID int, parentInternalID *int, limit, offset int,
) ([]Resource, error) {
	var resources []Resource
	err := s.withDBClient(func(dbClient provider.DBClientInterface) error {
		var results []map[string]interface{}
		var err error
		// Treat nil parent ID as top-level resources
		if parentInternalID == nil {
			results, err = dbClient.Query(
				queryGetResourceListByNullParent, resServerInternalID, limit, offset, s.deploymentID,
			)
		} else {
			results, err = dbClient.Query(
				queryGetResourceListByParent, resServerInternalID, *parentInternalID, limit, offset, s.deploymentID,
			)
		}

		if err != nil {
			return fmt.Errorf("failed to get resource list by parent: %w", err)
		}

		resources = make([]Resource, 0, len(results))
		for _, row := range results {
			_, res, err := buildResourceFromResultRow(row)
			if err != nil {
				return fmt.Errorf("failed to build resource: %w", err)
			}
			resources = append(resources, res)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return resources, nil
}

// GetResourceListCount retrieves the count of all resources.
func (s *resourceStore) GetResourceListCount(resServerInternalID int) (int, error) {
	var count int
	err := s.withDBClient(func(dbClient provider.DBClientInterface) error {
		results, err := dbClient.Query(queryGetResourceListCount, resServerInternalID, s.deploymentID)
		if err != nil {
			return fmt.Errorf("failed to get resource count: %w", err)
		}

		count, err = parseCountResult(results)
		return err
	})
	return count, err
}

// GetResourceListCountByParent retrieves count of resources by parent.
func (s *resourceStore) GetResourceListCountByParent(resServerInternalID int, parentInternalID *int) (int, error) {
	var count int
	err := s.withDBClient(func(dbClient provider.DBClientInterface) error {
		var results []map[string]interface{}
		var err error
		// Treat nil parent ID as top-level resources
		if parentInternalID == nil {
			results, err = dbClient.Query(queryGetResourceListCountByNullParent, resServerInternalID, s.deploymentID)
		} else {
			results, err = dbClient.Query(
				queryGetResourceListCountByParent, resServerInternalID, *parentInternalID, s.deploymentID)
		}

		if err != nil {
			return fmt.Errorf("failed to get resource count by parent: %w", err)
		}

		count, err = parseCountResult(results)
		return err
	})
	return count, err
}

// UpdateResource updates a resource.
func (s *resourceStore) UpdateResource(id string, resServerInternalID int, res Resource) error {
	return s.withDBClient(func(dbClient provider.DBClientInterface) error {
		_, err := dbClient.Execute(
			queryUpdateResource,
			res.Name,            // $1: NAME
			res.Description,     // $2: DESCRIPTION
			"{}",                // $3: PROPERTIES (empty JSON).
			id,                  // $4: RESOURCE_ID
			resServerInternalID, // $5: RESOURCE_SERVER_ID (internal ID)
			s.deploymentID,      // $6: DEPLOYMENT_ID
		)
		if err != nil {
			return fmt.Errorf("failed to update resource: %w", err)
		}

		return nil
	})
}

// DeleteResource deletes a resource.
func (s *resourceStore) DeleteResource(id string, resServerInternalID int) error {
	return s.withDBClient(func(dbClient provider.DBClientInterface) error {
		_, err := dbClient.Execute(queryDeleteResource, id, resServerInternalID, s.deploymentID)
		if err != nil {
			return fmt.Errorf("failed to delete resource: %w", err)
		}

		return nil
	})
}

// CheckResourceHandleExistsUnderParent checks if resource handle exists under parent.
func (s *resourceStore) CheckResourceHandleExists(
	resServerInternalID int, handle string, parentInternalID *int,
) (bool, error) {
	var exists bool
	err := s.withDBClient(func(dbClient provider.DBClientInterface) error {
		var results []map[string]interface{}
		var err error
		if parentInternalID == nil {
			results, err = dbClient.Query(
				queryCheckResourceHandleExistsUnderNullParent, resServerInternalID, handle, s.deploymentID,
			)
		} else {
			results, err = dbClient.Query(
				queryCheckResourceHandleExistsUnderParent, resServerInternalID, handle, *parentInternalID,
				s.deploymentID,
			)
		}

		if err != nil {
			return fmt.Errorf("failed to check resource handle: %w", err)
		}

		exists, err = parseBoolFromCount(results)
		return err
	})
	return exists, err
}

// CheckResourceHasDependencies checks if resource has dependencies.
func (s *resourceStore) CheckResourceHasDependencies(resInternalID int) (bool, error) {
	var hasDeps bool
	err := s.withDBClient(func(dbClient provider.DBClientInterface) error {
		results, err := dbClient.Query(queryCheckResourceHasDependencies, resInternalID, s.deploymentID)
		if err != nil {
			return fmt.Errorf("failed to check dependencies: %w", err)
		}

		hasDeps, err = parseBoolFromCount(results)
		return err
	})
	return hasDeps, err
}

// CheckCircularDependency checks if setting a parent would create circular dependency.
func (s *resourceStore) CheckCircularDependency(resourceID, newParentID string) (bool, error) {
	var hasCircular bool
	err := s.withDBClient(func(dbClient provider.DBClientInterface) error {
		results, err := dbClient.Query(queryCheckCircularDependency, newParentID, resourceID, s.deploymentID)
		if err != nil {
			return fmt.Errorf("failed to check circular dependency: %w", err)
		}

		hasCircular, err = parseBoolFromCount(results)
		return err
	})
	return hasCircular, err
}

// Action Store Methods

// CreateAction creates a new action.
func (s *resourceStore) CreateAction(
	uuid string,
	resServerInternalID int,
	resInternalID *int,
	action Action,
) error {
	return s.withDBClient(func(dbClient provider.DBClientInterface) error {
		_, err := dbClient.Execute(
			queryCreateAction,
			uuid,                // $1: ACTION_ID (UUID)
			resServerInternalID, // $2: RESOURCE_SERVER_ID (int FK)
			resInternalID,       // $3: RESOURCE_ID (int FK or NULL)
			action.Name,         // $4: NAME
			action.Handle,       // $5: handle
			action.Description,  // $6: DESCRIPTION
			action.Permission,   // $7: PERMISSION
			"{}",                // $8: PROPERTIES (empty JSON).
			s.deploymentID,      // $9: DEPLOYMENT_ID
		)
		if err != nil {
			return fmt.Errorf("failed to create action: %w", err)
		}

		return nil
	})
}

// GetAction retrieves an action and internal ID by UUID.
// If resInternalID is nil, retrieves action at resource server level.
// If resInternalID is provided, retrieves action at resource level.
func (s *resourceStore) GetAction(id string, resServerInternalID int, resInternalID *int) (Action, error) {
	var action Action
	err := s.withDBClient(func(dbClient provider.DBClientInterface) error {
		// Single unified query handles both resource server and resource level via nullable parameter
		results, err := dbClient.Query(queryGetActionByID, id, resServerInternalID, resInternalID, s.deploymentID)
		if err != nil {
			return fmt.Errorf("failed to get action: %w", err)
		}

		if len(results) == 0 {
			return errActionNotFound
		}

		action, err = buildActionFromResultRow(results[0])
		return err
	})
	return action, err
}

// GetActionList retrieves actions with pagination.
func (s *resourceStore) GetActionList(
	resServerInternalID int, resInternalID *int, limit, offset int,
) ([]Action, error) {
	var actions []Action
	err := s.withDBClient(func(dbClient provider.DBClientInterface) error {
		results, err := dbClient.Query(queryGetActionList, resServerInternalID, resInternalID, limit, offset,
			s.deploymentID)
		if err != nil {
			return fmt.Errorf("failed to get action list: %w", err)
		}

		actions = make([]Action, 0, len(results))
		for _, row := range results {
			action, err := buildActionFromResultRow(row)
			if err != nil {
				return fmt.Errorf("failed to build action: %w", err)
			}
			actions = append(actions, action)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return actions, nil
}

// GetActionListCount retrieves count of actions.
func (s *resourceStore) GetActionListCount(resServerInternalID int, resInternalID *int) (int, error) {
	var count int
	err := s.withDBClient(func(dbClient provider.DBClientInterface) error {
		results, err := dbClient.Query(queryGetActionListCount, resServerInternalID, resInternalID, s.deploymentID)
		if err != nil {
			return fmt.Errorf("failed to get action count: %w", err)
		}

		count, err = parseCountResult(results)
		return err
	})
	return count, err
}

// UpdateAction updates an action.
func (s *resourceStore) UpdateAction(id string, resServerInternalID int, resInternalID *int, action Action) error {
	return s.withDBClient(func(dbClient provider.DBClientInterface) error {
		// Single unified query handles both levels via nullable parameter
		_, err := dbClient.Execute(
			queryUpdateAction,
			action.Name,         // $1: NAME
			action.Description,  // $2: DESCRIPTION
			"{}",                // $3: PROPERTIES (empty JSON).
			id,                  // $4: ACTION_ID
			resServerInternalID, // $5: RESOURCE_SERVER_ID (internal ID)
			resInternalID,       // $6: RESOURCE_ID (internal ID or NULL)
			s.deploymentID,      // $7: DEPLOYMENT_ID
		)
		if err != nil {
			return fmt.Errorf("failed to update action: %w", err)
		}

		return nil
	})
}

// DeleteAction deletes an action.
func (s *resourceStore) DeleteAction(id string, resServerInternalID int, resInternalID *int) error {
	return s.withDBClient(func(dbClient provider.DBClientInterface) error {
		_, err := dbClient.Execute(
			queryDeleteAction,
			id,                  // $1: ACTION_ID
			resServerInternalID, // $2: RESOURCE_SERVER_ID (internal ID)
			resInternalID,       // $3: RESOURCE_ID (internal ID or NULL)
			s.deploymentID,      // $4: DEPLOYMENT_ID
		)
		if err != nil {
			return fmt.Errorf("failed to delete action: %w", err)
		}

		return nil
	})
}

// IsActionExist checks if an action exists.
func (s *resourceStore) IsActionExist(id string, resServerInternalID int, resInternalID *int) (bool, error) {
	var exists bool
	err := s.withDBClient(func(dbClient provider.DBClientInterface) error {
		results, err := dbClient.Query(queryCheckActionExists, id, resServerInternalID, resInternalID, s.deploymentID)
		if err != nil {
			return fmt.Errorf("failed to check action existence: %w", err)
		}

		exists, err = parseBoolFromCount(results)
		return err
	})
	return exists, err
}

// CheckActionHandleExists checks if action handle exists.
func (s *resourceStore) CheckActionHandleExists(
	resServerInternalID int, resInternalID *int, handle string,
) (bool, error) {
	var exists bool
	err := s.withDBClient(func(dbClient provider.DBClientInterface) error {
		results, err := dbClient.Query(
			queryCheckActionHandleExists, resServerInternalID, resInternalID, handle, s.deploymentID,
		)
		if err != nil {
			return fmt.Errorf("failed to check action handle: %w", err)
		}

		exists, err = parseBoolFromCount(results)
		return err
	})
	return exists, err
}

// ValidatePermissions validates that permissions exist for a given resource server.
// Returns array of invalid permissions (empty if all are valid).
func (s *resourceStore) ValidatePermissions(resServerInternalID int, permissions []string) ([]string, error) {
	// Early return for empty input
	if len(permissions) == 0 {
		return []string{}, nil
	}

	var invalidPermissions []string

	err := s.withDBClient(func(dbClient provider.DBClientInterface) error {
		// Convert permissions to JSON array for json_each()
		permissionsJSON, jsonErr := json.Marshal(permissions)
		if jsonErr != nil {
			return fmt.Errorf("failed to marshal permissions to JSON: %w", jsonErr)
		}

		// Query directly returns invalid permissions
		results, err := dbClient.Query(
			queryValidatePermissions,
			resServerInternalID,
			s.deploymentID,
			string(permissionsJSON),
		)
		if err != nil {
			return fmt.Errorf("failed to validate permissions: %w", err)
		}

		// Simply collect the invalid permissions returned by the query
		for _, row := range results {
			perm, ok := row["permission"].(string)
			if !ok {
				return fmt.Errorf("permission field is missing or invalid in query result")
			}
			invalidPermissions = append(invalidPermissions, perm)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return invalidPermissions, nil
}

// Helper methods

// getConfigDBClient retrieves the identity database client.
func (s *resourceStore) getConfigDBClient() (provider.DBClientInterface, error) {
	dbClient, err := s.dbProvider.GetConfigDBClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get identity DB client: %w", err)
	}
	return dbClient, nil
}

// withDBClient executes a function with a DB client, handling client retrieval errors.
func (s *resourceStore) withDBClient(fn func(provider.DBClientInterface) error) error {
	dbClient, err := s.getConfigDBClient()
	if err != nil {
		return err
	}
	return fn(dbClient)
}

// resolveIdentifier converts empty identifier to nil for database storage.
func resolveIdentifier(identifier string) interface{} {
	if identifier == "" {
		return nil
	}
	return identifier
}

// parseCountResult parses a count result from database query.
func parseCountResult(results []map[string]interface{}) (int, error) {
	if len(results) == 0 {
		return 0, fmt.Errorf("no count result returned")
	}

	countVal, ok := results[0]["total"]
	if !ok {
		countVal, ok = results[0]["count"]
		if !ok {
			return 0, fmt.Errorf("count field not found in result")
		}
	}

	switch v := countVal.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("unexpected count type: %T", countVal)
	}
}

// parseBoolFromCount parses a boolean from a count result.
func parseBoolFromCount(results []map[string]interface{}) (bool, error) {
	count, err := parseCountResult(results)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// resolveInternalID extracts and converts the internal ID from a database result row.
// Handles different integer types returned by various database drivers.
func resolveInternalID(row map[string]interface{}) (int, error) {
	switch v := row["id"].(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("unexpected internal ID type: %T", v)
	}
}

// resolveProperties extracts and sets the properties from the PROPERTIES column.
func resolveProperties(row map[string]interface{}, rs *ResourceServer) {
	if propsVal, ok := row["properties"]; ok && propsVal != nil {
		var props resourceServerProperties
		var propsBytes []byte

		switch v := propsVal.(type) {
		case string:
			propsBytes = []byte(v)
		case []byte:
			propsBytes = v
		}

		if len(propsBytes) > 0 {
			if err := json.Unmarshal(propsBytes, &props); err == nil {
				rs.Delimiter = props.Delimiter
			}
		}
	}
}

// buildPropertiesJSON builds the PROPERTIES JSON for a ResourceServer.
func buildPropertiesJSON(rs ResourceServer) interface{} {
	properties := resourceServerProperties{Delimiter: rs.Delimiter}
	if propsJSON, err := json.Marshal(properties); err == nil {
		return propsJSON
	}
	return json.RawMessage("{}")
}

// buildResourceServerFromResultRow builds a ResourceServer from a database result row.
func buildResourceServerFromResultRow(row map[string]interface{}) (int, ResourceServer, error) {
	rs := ResourceServer{}

	internalID, err := resolveInternalID(row)
	if err != nil {
		return 0, rs, err
	}
	if id, ok := row["resource_server_id"].(string); ok {
		rs.ID = id
	} else {
		return 0, rs, fmt.Errorf("resource_server_id field is missing or invalid")
	}

	if ouID, ok := row["ou_id"].(string); ok {
		rs.OrganizationUnitID = ouID
	} else {
		return 0, rs, fmt.Errorf("ou_id field is missing or invalid")
	}

	if name, ok := row["name"].(string); ok {
		rs.Name = name
	} else {
		return 0, rs, fmt.Errorf("name field is missing or invalid")
	}

	if desc, ok := row["description"].(string); ok {
		rs.Description = desc
	}

	if identifier, ok := row["identifier"].(string); ok {
		rs.Identifier = identifier
	}

	resolveProperties(row, &rs)

	return internalID, rs, nil
}

// buildResourceFromResultRow builds a Resource from a database result row.
func buildResourceFromResultRow(row map[string]interface{}) (int, Resource, error) {
	res := Resource{}
	internalID, err := resolveInternalID(row)
	if err != nil {
		return 0, res, err
	}
	if id, ok := row["resource_id"].(string); ok {
		res.ID = id
	} else {
		return 0, res, fmt.Errorf("resource_id field is missing or invalid")
	}

	if name, ok := row["name"].(string); ok {
		res.Name = name
	} else {
		return 0, res, fmt.Errorf("name field is missing or invalid")
	}

	if handle, ok := row["handle"].(string); ok {
		res.Handle = handle
	} else {
		return 0, res, fmt.Errorf("handle field is missing or invalid")
	}

	if desc, ok := row["description"].(string); ok {
		res.Description = desc
	}

	if permission, ok := row["permission"].(string); ok {
		res.Permission = permission
	}

	// PROPERTIES column exists in DB but not mapped to model (store as empty JSON)

	if parentID, ok := row["parent_resource_id"].(string); ok && parentID != "" {
		res.Parent = &parentID
	}

	return internalID, res, nil
}

// buildActionFromResultRow builds an Action from a database result row.
func buildActionFromResultRow(row map[string]interface{}) (Action, error) {
	action := Action{}

	if id, ok := row["action_id"].(string); ok {
		action.ID = id
	} else {
		return action, fmt.Errorf("action_id field is missing or invalid")
	}

	if name, ok := row["name"].(string); ok {
		action.Name = name
	} else {
		return action, fmt.Errorf("name field is missing or invalid")
	}

	if handle, ok := row["handle"].(string); ok {
		action.Handle = handle
	} else {
		return action, fmt.Errorf("handle field is missing or invalid")
	}

	if desc, ok := row["description"].(string); ok {
		action.Description = desc
	}

	if permission, ok := row["permission"].(string); ok {
		action.Permission = permission
	}

	// PROPERTIES column exists in DB but not mapped to model (store as empty JSON)

	return action, nil
}
