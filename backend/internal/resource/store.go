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
	"fmt"

	"github.com/asgardeo/thunder/internal/system/database/provider"
)

// resourceStoreInterface defines the interface for resource store operations.
type resourceStoreInterface interface {
	// Resource Server operations
	CreateResourceServer(id string, rs ResourceServer) error
	GetResourceServer(id string) (ResourceServer, error)
	GetResourceServerList(limit, offset int) ([]ResourceServer, error)
	GetResourceServerListCount() (int, error)
	UpdateResourceServer(id string, rs ResourceServer) error
	DeleteResourceServer(id string) error
	IsResourceServerExist(id string) (bool, error)
	CheckResourceServerNameExists(name string) (bool, error)
	CheckResourceServerNameExistsExcludingID(name, excludeID string) (bool, error)
	CheckResourceServerIdentifierExists(identifier string) (bool, error)
	CheckResourceServerIdentifierExistsExcludingID(identifier, excludeID string) (bool, error)
	CheckResourceServerHasDependencies(id string) (bool, error)
	CheckResourceServerExistAndGetInternalID(uuid string) (int, error)

	// Resource operations
	CreateResource(uuid string, resourceServerInternalID int, parentInternalID *int, res Resource) error
	GetResource(id, resourceServerID string) (Resource, error)
	GetResourceList(resourceServerID string, limit, offset int) ([]Resource, error)
	GetResourceListByParent(resourceServerID string, parentID *string, limit, offset int) ([]Resource, error)
	GetResourceListCount(resourceServerID string) (int, error)
	GetResourceListCountByParent(resourceServerID string, parentID *string) (int, error)
	UpdateResource(id, resourceServerID string, res Resource) error
	DeleteResource(id, resourceServerID string) error
	IsResourceExist(id, resourceServerID string) (bool, error)
	CheckResourceHandleExistsUnderParent(
		resourceServerID, handle string, parentID *string,
	) (bool, error)
	CheckResourceHasDependencies(id string) (bool, error)
	CheckCircularDependency(resourceID, newParentID string) (bool, error)
	CheckResourceExistAndGetInternalID(uuid, resourceServerUUID string) (int, error)

	// Action operations
	CreateAction(uuid string, resourceServerInternalID int, resourceInternalID *int, action Action) error
	GetAction(id, resourceServerID string, resourceID *string) (Action, error)
	GetActionListAtResourceServer(resourceServerID string, limit, offset int) ([]Action, error)
	GetActionListAtResource(resourceServerID, resourceID string, limit, offset int) ([]Action, error)
	GetActionListCountAtResourceServer(resourceServerID string) (int, error)
	GetActionListCountAtResource(resourceServerID, resourceID string) (int, error)
	UpdateAction(id, resourceServerID string, resourceID *string, action Action) error
	DeleteAction(id, resourceServerID string, resourceID *string) error
	IsActionExist(id, resourceServerID string, resourceID *string) (bool, error)
	CheckActionHandleExists(resourceServerID string, resourceID *string, handle string) (bool, error)
}

// resourceStore is the default implementation of resourceStoreInterface.
type resourceStore struct {
	dbProvider provider.DBProviderInterface
}

// newResourceStore creates a new instance of resourceStore.
func newResourceStore() resourceStoreInterface {
	return &resourceStore{
		dbProvider: provider.GetDBProvider(),
	}
}

// Helper methods

// getIdentityDBClient retrieves the identity database client.
func (s *resourceStore) getIdentityDBClient() (provider.DBClientInterface, error) {
	dbClient, err := s.dbProvider.GetConfigDBClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get identity DB client: %w", err)
	}
	return dbClient, nil
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

// buildResourceServerFromResultRow builds a ResourceServer from a database result row.
func buildResourceServerFromResultRow(row map[string]interface{}) (ResourceServer, error) {
	rs := ResourceServer{}

	if id, ok := row["resource_server_id"].(string); ok {
		rs.ID = id
	} else {
		return rs, fmt.Errorf("resource_server_id field is missing or invalid")
	}

	if ouID, ok := row["ou_id"].(string); ok {
		rs.OrganizationUnitID = ouID
	} else {
		return rs, fmt.Errorf("ou_id field is missing or invalid")
	}

	if name, ok := row["name"].(string); ok {
		rs.Name = name
	} else {
		return rs, fmt.Errorf("name field is missing or invalid")
	}

	if desc, ok := row["description"].(string); ok {
		rs.Description = desc
	}

	if identifier, ok := row["identifier"].(string); ok {
		rs.Identifier = identifier
	}

	return rs, nil
}

// buildResourceFromResultRow builds a Resource from a database result row.
func buildResourceFromResultRow(row map[string]interface{}) (Resource, error) {
	res := Resource{}

	if id, ok := row["resource_id"].(string); ok {
		res.ID = id
	} else {
		return res, fmt.Errorf("resource_id field is missing or invalid")
	}

	if name, ok := row["name"].(string); ok {
		res.Name = name
	} else {
		return res, fmt.Errorf("name field is missing or invalid")
	}

	if handle, ok := row["handle"].(string); ok {
		res.Handle = handle
	} else {
		return res, fmt.Errorf("handle field is missing or invalid")
	}

	if desc, ok := row["description"].(string); ok {
		res.Description = desc
	}

	// PROPERTIES column exists in DB but not mapped to model (store as empty JSON)

	if parentID, ok := row["parent_resource_id"].(string); ok && parentID != "" {
		res.Parent = &parentID
	}

	return res, nil
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

	// PROPERTIES column exists in DB but not mapped to model (store as empty JSON)

	if resourceID, ok := row["resource_id"].(string); ok && resourceID != "" {
		action.ResourceID = &resourceID
	}

	return action, nil
}

// Resource Server Store Methods

// CreateResourceServer creates a new resource server in the database.
func (s *resourceStore) CreateResourceServer(id string, rs ResourceServer) error {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return err
	}

	// Convert empty identifier to NULL for database uniqueness constraint
	var identifier interface{}
	if rs.Identifier == "" {
		identifier = nil
	} else {
		identifier = rs.Identifier
	}

	_, err = dbClient.Execute(
		queryCreateResourceServer,
		id,
		rs.OrganizationUnitID,
		rs.Name,
		rs.Description,
		identifier,
		"{}", // PROPERTIES as empty JSON, as of now. Future: can be extended to accept properties.
	)
	if err != nil {
		return fmt.Errorf("failed to create resource server: %w", err)
	}

	return nil
}

// GetResourceServer retrieves a resource server by ID.
func (s *resourceStore) GetResourceServer(id string) (ResourceServer, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return ResourceServer{}, err
	}

	results, err := dbClient.Query(queryGetResourceServerByID, id)
	if err != nil {
		return ResourceServer{}, fmt.Errorf("failed to get resource server: %w", err)
	}

	if len(results) == 0 {
		return ResourceServer{}, ErrResourceServerNotFound
	}

	return buildResourceServerFromResultRow(results[0])
}

// GetResourceServerList retrieves a list of resource servers with pagination.
func (s *resourceStore) GetResourceServerList(limit, offset int) ([]ResourceServer, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return nil, err
	}

	results, err := dbClient.Query(queryGetResourceServerList, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource server list: %w", err)
	}

	resourceServers := make([]ResourceServer, 0, len(results))
	for _, row := range results {
		rs, err := buildResourceServerFromResultRow(row)
		if err != nil {
			return nil, fmt.Errorf("failed to build resource server: %w", err)
		}
		resourceServers = append(resourceServers, rs)
	}

	return resourceServers, nil
}

// GetResourceServerListCount retrieves the total count of resource servers.
func (s *resourceStore) GetResourceServerListCount() (int, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return 0, err
	}

	results, err := dbClient.Query(queryGetResourceServerListCount)
	if err != nil {
		return 0, fmt.Errorf("failed to get resource server count: %w", err)
	}

	return parseCountResult(results)
}

// UpdateResourceServer updates a resource server.
func (s *resourceStore) UpdateResourceServer(id string, rs ResourceServer) error {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return err
	}

	// Convert empty identifier to NULL for database uniqueness constraint
	var identifier interface{}
	if rs.Identifier == "" {
		identifier = nil
	} else {
		identifier = rs.Identifier
	}

	_, err = dbClient.Execute(
		queryUpdateResourceServer,
		rs.OrganizationUnitID,
		rs.Name,
		rs.Description,
		identifier,
		"{}", // PROPERTIES as empty JSON
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to update resource server: %w", err)
	}

	return nil
}

// DeleteResourceServer deletes a resource server.
func (s *resourceStore) DeleteResourceServer(id string) error {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return err
	}

	_, err = dbClient.Execute(queryDeleteResourceServer, id)
	if err != nil {
		return fmt.Errorf("failed to delete resource server: %w", err)
	}

	return nil
}

// IsResourceServerExist checks if a resource server exists.
func (s *resourceStore) IsResourceServerExist(id string) (bool, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return false, err
	}

	results, err := dbClient.Query(queryCheckResourceServerExists, id)
	if err != nil {
		return false, fmt.Errorf("failed to check resource server existence: %w", err)
	}

	return parseBoolFromCount(results)
}

// CheckResourceServerNameExists checks if a resource server name exists.
func (s *resourceStore) CheckResourceServerNameExists(name string) (bool, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return false, err
	}

	results, err := dbClient.Query(queryCheckResourceServerNameExists, name)
	if err != nil {
		return false, fmt.Errorf("failed to check resource server name: %w", err)
	}

	return parseBoolFromCount(results)
}

// CheckResourceServerNameExistsExcludingID checks if name exists excluding a specific ID.
func (s *resourceStore) CheckResourceServerNameExistsExcludingID(name, excludeID string) (bool, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return false, err
	}

	results, err := dbClient.Query(queryCheckResourceServerNameExistsExcludingID, name, excludeID)
	if err != nil {
		return false, fmt.Errorf("failed to check resource server name: %w", err)
	}

	return parseBoolFromCount(results)
}

// CheckResourceServerIdentifierExists checks if a resource server identifier exists.
func (s *resourceStore) CheckResourceServerIdentifierExists(identifier string) (bool, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return false, err
	}

	results, err := dbClient.Query(queryCheckResourceServerIdentifierExists, identifier)
	if err != nil {
		return false, fmt.Errorf("failed to check resource server identifier: %w", err)
	}

	return parseBoolFromCount(results)
}

// CheckResourceServerIdentifierExistsExcludingID checks if identifier exists excluding a specific ID.
func (s *resourceStore) CheckResourceServerIdentifierExistsExcludingID(identifier, excludeID string) (bool, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return false, err
	}

	results, err := dbClient.Query(queryCheckResourceServerIdentifierExistsExcludingID, identifier, excludeID)
	if err != nil {
		return false, fmt.Errorf("failed to check resource server identifier: %w", err)
	}

	return parseBoolFromCount(results)
}

// CheckResourceServerHasDependencies checks if resource server has dependencies.
func (s *resourceStore) CheckResourceServerHasDependencies(id string) (bool, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return false, err
	}

	results, err := dbClient.Query(queryCheckResourceServerHasDependencies, id)
	if err != nil {
		return false, fmt.Errorf("failed to check dependencies: %w", err)
	}

	return parseBoolFromCount(results)
}

// CheckResourceServerExistAndGetInternalID checks if resource server exists and returns its internal ID.
func (s *resourceStore) CheckResourceServerExistAndGetInternalID(uuid string) (int, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return 0, err
	}

	results, err := dbClient.Query(queryGetResourceServerInternalID, uuid)
	if err != nil {
		return 0, fmt.Errorf("failed to check resource server: %w", err)
	}

	if len(results) == 0 {
		return 0, ErrResourceServerNotFound
	}

	// Handle different integer types returned by database drivers
	var internalID int
	switch v := results[0]["id"].(type) {
	case int:
		internalID = v
	case int64:
		internalID = int(v)
	case float64:
		internalID = int(v)
	default:
		return 0, fmt.Errorf("unexpected internal ID type: %T", v)
	}

	return internalID, nil
}

// Resource Store Methods

// CreateResource creates a new resource.
func (s *resourceStore) CreateResource(
	uuid string,
	resourceServerInternalID int,
	parentInternalID *int,
	res Resource,
) error {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return err
	}

	_, err = dbClient.Execute(
		queryCreateResource,
		uuid,                     // $1: RESOURCE_ID (UUID)
		resourceServerInternalID, // $2: RESOURCE_SERVER_ID (int FK)
		res.Name,                 // $3: NAME
		res.Handle,               // $4: HANDLE
		res.Description,          // $5: DESCRIPTION
		"{}",                     // $6: PROPERTIES (empty JSON)
		parentInternalID,         // $7: PARENT_RESOURCE_ID (int FK or NULL)
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	return nil
}

// GetResource retrieves a resource by ID.
func (s *resourceStore) GetResource(id, resourceServerID string) (Resource, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return Resource{}, err
	}

	results, err := dbClient.Query(queryGetResourceByID, id, resourceServerID)
	if err != nil {
		return Resource{}, fmt.Errorf("failed to get resource: %w", err)
	}

	if len(results) == 0 {
		return Resource{}, ErrResourceNotFound
	}

	return buildResourceFromResultRow(results[0])
}

// GetResourceList retrieves all resources for a resource server.
func (s *resourceStore) GetResourceList(resourceServerID string, limit, offset int) ([]Resource, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return nil, err
	}

	results, err := dbClient.Query(queryGetResourceList, resourceServerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource list: %w", err)
	}

	resources := make([]Resource, 0, len(results))
	for _, row := range results {
		res, err := buildResourceFromResultRow(row)
		if err != nil {
			return nil, fmt.Errorf("failed to build resource: %w", err)
		}
		resources = append(resources, res)
	}

	return resources, nil
}

// GetResourceListByParent retrieves resources filtered by parent.
func (s *resourceStore) GetResourceListByParent(
	resourceServerID string, parentID *string, limit, offset int,
) ([]Resource, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	// Treat empty string parent ID same as nil (top-level resources)
	if parentID == nil || *parentID == "" {
		results, err = dbClient.Query(
			queryGetResourceListByNullParent, resourceServerID, limit, offset,
		)
	} else {
		results, err = dbClient.Query(
			queryGetResourceListByParent, resourceServerID, *parentID, limit, offset,
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get resource list by parent: %w", err)
	}

	resources := make([]Resource, 0, len(results))
	for _, row := range results {
		res, err := buildResourceFromResultRow(row)
		if err != nil {
			return nil, fmt.Errorf("failed to build resource: %w", err)
		}
		resources = append(resources, res)
	}

	return resources, nil
}

// GetResourceListCount retrieves the count of all resources.
func (s *resourceStore) GetResourceListCount(resourceServerID string) (int, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return 0, err
	}

	results, err := dbClient.Query(queryGetResourceListCount, resourceServerID)
	if err != nil {
		return 0, fmt.Errorf("failed to get resource count: %w", err)
	}

	return parseCountResult(results)
}

// GetResourceListCountByParent retrieves count of resources by parent.
func (s *resourceStore) GetResourceListCountByParent(resourceServerID string, parentID *string) (int, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return 0, err
	}

	var results []map[string]interface{}
	// Treat empty string parent ID same as nil (top-level resources)
	if parentID == nil || *parentID == "" {
		results, err = dbClient.Query(queryGetResourceListCountByNullParent, resourceServerID)
	} else {
		results, err = dbClient.Query(queryGetResourceListCountByParent, resourceServerID, *parentID)
	}

	if err != nil {
		return 0, fmt.Errorf("failed to get resource count by parent: %w", err)
	}

	return parseCountResult(results)
}

// UpdateResource updates a resource.
func (s *resourceStore) UpdateResource(id, resourceServerID string, res Resource) error {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return err
	}

	_, err = dbClient.Execute(
		queryUpdateResource,
		res.Name,         // $1: NAME
		res.Description,  // $2: DESCRIPTION
		"{}",             // $3: PROPERTIES (empty JSON). Future: can be extended to accept properties.
		id,               // $4: RESOURCE_ID
		resourceServerID, // $5: RESOURCE_SERVER_ID
	)
	if err != nil {
		return fmt.Errorf("failed to update resource: %w", err)
	}

	return nil
}

// DeleteResource deletes a resource.
func (s *resourceStore) DeleteResource(id, resourceServerID string) error {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return err
	}

	_, err = dbClient.Execute(queryDeleteResource, id, resourceServerID)
	if err != nil {
		return fmt.Errorf("failed to delete resource: %w", err)
	}

	return nil
}

// IsResourceExist checks if a resource exists.
func (s *resourceStore) IsResourceExist(id, resourceServerID string) (bool, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return false, err
	}

	results, err := dbClient.Query(queryCheckResourceExists, id, resourceServerID)
	if err != nil {
		return false, fmt.Errorf("failed to check resource existence: %w", err)
	}

	return parseBoolFromCount(results)
}

// CheckResourceHandleExistsUnderParent checks if resource handle exists under parent.
func (s *resourceStore) CheckResourceHandleExistsUnderParent(
	resourceServerID, handle string, parentID *string,
) (bool, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return false, err
	}

	var results []map[string]interface{}
	if parentID == nil {
		results, err = dbClient.Query(
			queryCheckResourceHandleExistsUnderNullParent, resourceServerID, handle,
		)
	} else {
		results, err = dbClient.Query(
			queryCheckResourceHandleExistsUnderParent, resourceServerID, handle, *parentID,
		)
	}

	if err != nil {
		return false, fmt.Errorf("failed to check resource handle: %w", err)
	}

	return parseBoolFromCount(results)
}

// CheckResourceHasDependencies checks if resource has dependencies.
func (s *resourceStore) CheckResourceHasDependencies(id string) (bool, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return false, err
	}

	results, err := dbClient.Query(queryCheckResourceHasDependencies, id)
	if err != nil {
		return false, fmt.Errorf("failed to check dependencies: %w", err)
	}

	return parseBoolFromCount(results)
}

// CheckCircularDependency checks if setting a parent would create circular dependency.
func (s *resourceStore) CheckCircularDependency(resourceID, newParentID string) (bool, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return false, err
	}

	results, err := dbClient.Query(queryCheckCircularDependency, newParentID, resourceID)
	if err != nil {
		return false, fmt.Errorf("failed to check circular dependency: %w", err)
	}

	return parseBoolFromCount(results)
}

// CheckResourceExistAndGetInternalID checks if resource exists and returns its internal ID.
func (s *resourceStore) CheckResourceExistAndGetInternalID(uuid, resourceServerUUID string) (int, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return 0, err
	}

	results, err := dbClient.Query(queryGetResourceInternalID, uuid, resourceServerUUID)
	if err != nil {
		return 0, fmt.Errorf("failed to check resource: %w", err)
	}

	if len(results) == 0 {
		return 0, ErrResourceNotFound
	}

	// Handle different integer types returned by database drivers
	var internalID int
	switch v := results[0]["id"].(type) {
	case int:
		internalID = v
	case int64:
		internalID = int(v)
	case float64:
		internalID = int(v)
	default:
		return 0, fmt.Errorf("unexpected internal ID type: %T", v)
	}

	return internalID, nil
}

// Action Store Methods

// CreateAction creates a new action.
func (s *resourceStore) CreateAction(
	uuid string,
	resourceServerInternalID int,
	resourceInternalID *int,
	action Action,
) error {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return err
	}

	_, err = dbClient.Execute(
		queryCreateAction,
		uuid,                     // $1: ACTION_ID (UUID)
		resourceServerInternalID, // $2: RESOURCE_SERVER_ID (int FK)
		resourceInternalID,       // $3: RESOURCE_ID (int FK or NULL)
		action.Name,              // $4: NAME
		action.Handle,            // $5: handle
		action.Description,       // $6: DESCRIPTION
		"{}",                     // $7: PROPERTIES (empty JSON)
	)
	if err != nil {
		return fmt.Errorf("failed to create action: %w", err)
	}

	return nil
}

// GetAction retrieves an action by ID.
func (s *resourceStore) GetAction(id, resourceServerID string, resourceID *string) (Action, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return Action{}, err
	}

	var results []map[string]interface{}
	if resourceID == nil {
		results, err = dbClient.Query(queryGetActionByID, id, resourceServerID)
	} else {
		results, err = dbClient.Query(queryGetActionByIDAtResource, id, resourceServerID, *resourceID)
	}

	if err != nil {
		return Action{}, fmt.Errorf("failed to get action: %w", err)
	}

	if len(results) == 0 {
		return Action{}, ErrActionNotFound
	}

	return buildActionFromResultRow(results[0])
}

// GetActionListAtResourceServer retrieves actions at resource server level.
func (s *resourceStore) GetActionListAtResourceServer(resourceServerID string, limit, offset int) ([]Action, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return nil, err
	}

	results, err := dbClient.Query(queryGetActionListAtResourceServer, resourceServerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get action list: %w", err)
	}

	actions := make([]Action, 0, len(results))
	for _, row := range results {
		action, err := buildActionFromResultRow(row)
		if err != nil {
			return nil, fmt.Errorf("failed to build action: %w", err)
		}
		actions = append(actions, action)
	}

	return actions, nil
}

// GetActionListAtResource retrieves actions at resource level.
func (s *resourceStore) GetActionListAtResource(
	resourceServerID, resourceID string, limit, offset int,
) ([]Action, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return nil, err
	}

	results, err := dbClient.Query(
		queryGetActionListAtResource, resourceServerID, resourceID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get action list: %w", err)
	}

	actions := make([]Action, 0, len(results))
	for _, row := range results {
		action, err := buildActionFromResultRow(row)
		if err != nil {
			return nil, fmt.Errorf("failed to build action: %w", err)
		}
		actions = append(actions, action)
	}

	return actions, nil
}

// GetActionListCountAtResourceServer retrieves count of actions at resource server level.
func (s *resourceStore) GetActionListCountAtResourceServer(resourceServerID string) (int, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return 0, err
	}

	results, err := dbClient.Query(queryGetActionListCountAtResourceServer, resourceServerID)
	if err != nil {
		return 0, fmt.Errorf("failed to get action count: %w", err)
	}

	return parseCountResult(results)
}

// GetActionListCountAtResource retrieves count of actions at resource level.
func (s *resourceStore) GetActionListCountAtResource(resourceServerID, resourceID string) (int, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return 0, err
	}

	results, err := dbClient.Query(queryGetActionListCountAtResource, resourceServerID, resourceID)
	if err != nil {
		return 0, fmt.Errorf("failed to get action count: %w", err)
	}

	return parseCountResult(results)
}

// UpdateAction updates an action.
func (s *resourceStore) UpdateAction(id, resourceServerID string, resourceID *string, action Action) error {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return err
	}

	if resourceID == nil {
		// Update action at resource server level
		_, err = dbClient.Execute(
			queryUpdateAction,
			action.Name,        // $1: NAME
			action.Description, // $2: DESCRIPTION
			"{}",               // $3: PROPERTIES (empty JSON). Future: can be extended to accept properties.
			id,                 // $4: ACTION_ID
			resourceServerID,   // $5: RESOURCE_SERVER_ID
		)
	} else {
		// Update action at resource level
		_, err = dbClient.Execute(
			queryUpdateActionAtResource,
			action.Name,        // $1: NAME
			action.Description, // $2: DESCRIPTION
			"{}",               // $3: PROPERTIES (empty JSON). Future: can be extended to accept properties.
			id,                 // $4: ACTION_ID
			resourceServerID,   // $5: RESOURCE_SERVER_ID
			*resourceID,        // $6: RESOURCE_ID
		)
	}

	if err != nil {
		return fmt.Errorf("failed to update action: %w", err)
	}

	return nil
}

// DeleteAction deletes an action.
func (s *resourceStore) DeleteAction(id, resourceServerID string, resourceID *string) error {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return err
	}

	if resourceID == nil {
		// Delete action at resource server level
		_, err = dbClient.Execute(
			queryDeleteAction,
			id,               // $1: ACTION_ID
			resourceServerID, // $2: RESOURCE_SERVER_ID
		)
	} else {
		// Delete action at resource level
		_, err = dbClient.Execute(
			queryDeleteActionAtResource,
			id,               // $1: ACTION_ID
			resourceServerID, // $2: RESOURCE_SERVER_ID
			*resourceID,      // $3: RESOURCE_ID
		)
	}

	if err != nil {
		return fmt.Errorf("failed to delete action: %w", err)
	}

	return nil
}

// IsActionExist checks if an action exists.
func (s *resourceStore) IsActionExist(id, resourceServerID string, resourceID *string) (bool, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return false, err
	}

	var results []map[string]interface{}
	if resourceID == nil {
		results, err = dbClient.Query(queryCheckActionExists, id, resourceServerID)
	} else {
		results, err = dbClient.Query(queryCheckActionExistsAtResource, id, resourceServerID, *resourceID)
	}

	if err != nil {
		return false, fmt.Errorf("failed to check action existence: %w", err)
	}

	return parseBoolFromCount(results)
}

// CheckActionHandleExists checks if action handle exists.
func (s *resourceStore) CheckActionHandleExists(
	resourceServerID string, resourceID *string, handle string,
) (bool, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return false, err
	}

	var results []map[string]interface{}
	if resourceID == nil {
		results, err = dbClient.Query(queryCheckActionHandleExistsAtResourceServer, resourceServerID, handle)
	} else {
		results, err = dbClient.Query(queryCheckActionHandleExistsAtResource, resourceServerID, *resourceID, handle)
	}

	if err != nil {
		return false, fmt.Errorf("failed to check action handle: %w", err)
	}

	return parseBoolFromCount(results)
}
