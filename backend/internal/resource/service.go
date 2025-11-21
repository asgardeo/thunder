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

// Package resource implements the resource management service.
package resource

import (
	"errors"
	"fmt"

	oupkg "github.com/asgardeo/thunder/internal/ou"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
)

const loggerComponentName = "ResourceMgtService"

// ResourceServiceInterface defines the interface for the resource service.
type ResourceServiceInterface interface {
	// Resource Server operations
	CreateResourceServer(rs ResourceServer) (*ResourceServer, *serviceerror.ServiceError)
	GetResourceServer(id string) (*ResourceServer, *serviceerror.ServiceError)
	GetResourceServerList(limit, offset int) (*ResourceServerList, *serviceerror.ServiceError)
	UpdateResourceServer(id string, rs ResourceServer) (*ResourceServer, *serviceerror.ServiceError)
	DeleteResourceServer(id string) *serviceerror.ServiceError

	// Resource operations
	CreateResource(resourceServerID string, res Resource) (*Resource, *serviceerror.ServiceError)
	GetResource(resourceServerID, id string) (*Resource, *serviceerror.ServiceError)
	GetResourceList(
		resourceServerID string, parentID *string, limit, offset int,
	) (*ResourceList, *serviceerror.ServiceError)
	UpdateResource(resourceServerID, id string, res Resource) (*Resource, *serviceerror.ServiceError)
	DeleteResource(resourceServerID, id string) *serviceerror.ServiceError

	// Action operations
	CreateActionAtResourceServer(resourceServerID string, action Action) (*Action, *serviceerror.ServiceError)
	CreateActionAtResource(resourceServerID, resourceID string, action Action) (*Action, *serviceerror.ServiceError)
	GetActionAtResourceServer(resourceServerID, id string) (*Action, *serviceerror.ServiceError)
	GetActionAtResource(resourceServerID, resourceID, id string) (*Action, *serviceerror.ServiceError)
	GetActionListAtResourceServer(
		resourceServerID string, limit, offset int,
	) (*ActionList, *serviceerror.ServiceError)
	GetActionListAtResource(
		resourceServerID, resourceID string, limit, offset int,
	) (*ActionList, *serviceerror.ServiceError)
	UpdateActionAtResourceServer(resourceServerID, id string, action Action) (*Action, *serviceerror.ServiceError)
	UpdateActionAtResource(resourceServerID, resourceID, id string, action Action) (*Action, *serviceerror.ServiceError)
	DeleteActionAtResourceServer(resourceServerID, id string) *serviceerror.ServiceError
	DeleteActionAtResource(resourceServerID, resourceID, id string) *serviceerror.ServiceError
}

// resourceService is the default implementation of ResourceServiceInterface.
type resourceService struct {
	resourceStore resourceStoreInterface
	ouService     oupkg.OrganizationUnitServiceInterface
}

// newResourceService creates a new instance of ResourceService.
func newResourceService(
	resourceStore resourceStoreInterface,
	ouService oupkg.OrganizationUnitServiceInterface,
) ResourceServiceInterface {
	return &resourceService{
		resourceStore: resourceStore,
		ouService:     ouService,
	}
}

// Resource Server Methods

// CreateResourceServer creates a new resource server.
func (rs *resourceService) CreateResourceServer(
	resourceServer ResourceServer,
) (*ResourceServer, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Creating resource server", log.String("name", resourceServer.Name))

	if err := rs.validateResourceServerCreate(resourceServer); err != nil {
		return nil, err
	}

	// Validate organization unit exists
	_, svcErr := rs.ouService.GetOrganizationUnit(resourceServer.OrganizationUnitID)
	if svcErr != nil {
		if svcErr.Code == oupkg.ErrorOrganizationUnitNotFound.Code {
			logger.Debug("Organization unit not found", log.String("ouID", resourceServer.OrganizationUnitID))
			return nil, &ErrorOrganizationUnitNotFound
		}
		logger.Error("Failed to validate organization unit", log.String("error", svcErr.Error))
		return nil, &serviceerror.InternalServerError
	}

	// Check name uniqueness
	nameExists, err := rs.resourceStore.CheckResourceServerNameExists(resourceServer.Name)
	if err != nil {
		logger.Error("Failed to check resource server name", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if nameExists {
		logger.Debug("Resource server name already exists", log.String("name", resourceServer.Name))
		return nil, &ErrorNameConflict
	}

	// Check identifier uniqueness (if provided)
	if resourceServer.Identifier != "" {
		identifierExists, err := rs.resourceStore.CheckResourceServerIdentifierExists(resourceServer.Identifier)
		if err != nil {
			logger.Error("Failed to check resource server identifier", log.Error(err))
			return nil, &serviceerror.InternalServerError
		}
		if identifierExists {
			logger.Debug("Resource server identifier already exists", log.String("identifier", resourceServer.Identifier))
			return nil, &ErrorIdentifierConflict
		}
	}

	id := utils.GenerateUUIDv7()
	if err := rs.resourceStore.CreateResourceServer(id, resourceServer); err != nil {
		logger.Error("Failed to create resource server", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	createdRS := &ResourceServer{
		ID:                 id,
		Name:               resourceServer.Name,
		Description:        resourceServer.Description,
		Identifier:         resourceServer.Identifier,
		OrganizationUnitID: resourceServer.OrganizationUnitID,
	}

	logger.Debug("Successfully created resource server", log.String("id", id))
	return createdRS, nil
}

// GetResourceServer retrieves a resource server by ID.
func (rs *resourceService) GetResourceServer(id string) (*ResourceServer, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if id == "" {
		return nil, &ErrorMissingID
	}

	resourceServer, err := rs.resourceStore.GetResourceServer(id)
	if err != nil {
		if errors.Is(err, ErrResourceServerNotFound) {
			logger.Debug("Resource server not found", log.String("id", id))
			return nil, &ErrorResourceServerNotFound
		}
		logger.Error("Failed to get resource server", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	return &resourceServer, nil
}

// GetResourceServerList retrieves a paginated list of resource servers.
func (rs *resourceService) GetResourceServerList(limit, offset int) (*ResourceServerList, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}

	totalCount, err := rs.resourceStore.GetResourceServerListCount()
	if err != nil {
		logger.Error("Failed to get resource server count", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	resourceServers, err := rs.resourceStore.GetResourceServerList(limit, offset)
	if err != nil {
		logger.Error("Failed to list resource servers", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	response := &ResourceServerList{
		TotalResults:    totalCount,
		ResourceServers: resourceServers,
		StartIndex:      offset + 1,
		Count:           len(resourceServers),
		Links:           buildPaginationLinks("/resource-servers", limit, offset, totalCount),
	}

	return response, nil
}

// UpdateResourceServer updates a resource server.
func (rs *resourceService) UpdateResourceServer(
	id string, resourceServer ResourceServer,
) (*ResourceServer, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if id == "" {
		return nil, &ErrorMissingID
	}

	if err := rs.validateResourceServerUpdate(resourceServer); err != nil {
		return nil, err
	}

	exists, err := rs.resourceStore.IsResourceServerExist(id)
	if err != nil {
		logger.Error("Failed to check resource server existence", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if !exists {
		logger.Debug("Resource server not found", log.String("id", id))
		return nil, &ErrorResourceServerNotFound
	}

	// Validate organization unit
	_, svcErr := rs.ouService.GetOrganizationUnit(resourceServer.OrganizationUnitID)
	if svcErr != nil {
		if svcErr.Code == oupkg.ErrorOrganizationUnitNotFound.Code {
			return nil, &ErrorOrganizationUnitNotFound
		}
		return nil, &serviceerror.InternalServerError
	}

	// Check name uniqueness excluding current ID
	nameExists, err := rs.resourceStore.CheckResourceServerNameExistsExcludingID(
		resourceServer.Name, id,
	)
	if err != nil {
		logger.Error("Failed to check resource server name", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if nameExists {
		return nil, &ErrorNameConflict
	}

	// Check identifier uniqueness excluding current ID (if provided)
	if resourceServer.Identifier != "" {
		identifierExists, err := rs.resourceStore.CheckResourceServerIdentifierExistsExcludingID(
			resourceServer.Identifier, id,
		)
		if err != nil {
			logger.Error("Failed to check resource server identifier", log.Error(err))
			return nil, &serviceerror.InternalServerError
		}
		if identifierExists {
			logger.Debug("Resource server identifier already exists", log.String("identifier", resourceServer.Identifier))
			return nil, &ErrorIdentifierConflict
		}
	}

	if err := rs.resourceStore.UpdateResourceServer(id, resourceServer); err != nil {
		logger.Error("Failed to update resource server", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	updatedRS := &ResourceServer{
		ID:                 id,
		Name:               resourceServer.Name,
		Description:        resourceServer.Description,
		Identifier:         resourceServer.Identifier,
		OrganizationUnitID: resourceServer.OrganizationUnitID,
	}

	return updatedRS, nil
}

// DeleteResourceServer deletes a resource server.
func (rs *resourceService) DeleteResourceServer(id string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if id == "" {
		return &ErrorMissingID
	}

	exists, err := rs.resourceStore.IsResourceServerExist(id)
	if err != nil {
		logger.Error("Failed to check resource server existence", log.Error(err))
		return &serviceerror.InternalServerError
	}
	if !exists {
		return nil // Idempotent delete
	}

	// Check for dependencies
	hasDeps, err := rs.resourceStore.CheckResourceServerHasDependencies(id)
	if err != nil {
		logger.Error("Failed to check dependencies", log.Error(err))
		return &serviceerror.InternalServerError
	}
	if hasDeps {
		return &ErrorCannotDelete
	}

	if err := rs.resourceStore.DeleteResourceServer(id); err != nil {
		logger.Error("Failed to delete resource server", log.Error(err))
		return &serviceerror.InternalServerError
	}

	return nil
}

// Resource Methods

// CreateResource creates a new resource.
func (rs *resourceService) CreateResource(
	resourceServerID string, resource Resource,
) (*Resource, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	// Validate resource server exists and get internal ID
	resourceServerInternalID, err := rs.resourceStore.CheckResourceServerExistAndGetInternalID(resourceServerID)
	if err != nil {
		if errors.Is(err, ErrResourceServerNotFound) {
			return nil, &ErrorResourceServerNotFound
		}
		logger.Error("Failed to check resource server existence", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	if err := rs.validateResourceCreate(resource); err != nil {
		return nil, err
	}

	// Validate parent if specified and get internal ID
	var parentInternalID *int
	if resource.Parent != nil {
		parentID, err := rs.resourceStore.CheckResourceExistAndGetInternalID(*resource.Parent, resourceServerID)
		if err != nil {
			if errors.Is(err, ErrResourceNotFound) {
				return nil, &ErrorParentResourceNotFound
			}
			logger.Error("Failed to check parent resource", log.Error(err))
			return nil, &serviceerror.InternalServerError
		}
		parentInternalID = &parentID
	}

	// Check handle uniqueness under parent
	handleExists, err := rs.resourceStore.CheckResourceHandleExistsUnderParent(
		resourceServerID, resource.Handle, resource.Parent,
	)
	if err != nil {
		logger.Error("Failed to check resource handle", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if handleExists {
		return nil, &ErrorHandleConflict
	}

	id := utils.GenerateUUIDv7()
	if err := rs.resourceStore.CreateResource(id, resourceServerInternalID, parentInternalID, resource); err != nil {
		logger.Error("Failed to create resource", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	createdResource := &Resource{
		ID:          id,
		Name:        resource.Name,
		Handle:      resource.Handle,
		Description: resource.Description,
		Parent:      resource.Parent,
	}

	return createdResource, nil
}

// GetResource retrieves a resource by ID.
func (rs *resourceService) GetResource(resourceServerID, id string) (*Resource, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if id == "" || resourceServerID == "" {
		return nil, &ErrorMissingID
	}

	// Validate resource server exists
	rsExists, err := rs.resourceStore.IsResourceServerExist(resourceServerID)
	if err != nil {
		logger.Error("Failed to check resource server", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if !rsExists {
		return nil, &ErrorResourceServerNotFound
	}

	resource, err := rs.resourceStore.GetResource(id, resourceServerID)
	if err != nil {
		if errors.Is(err, ErrResourceNotFound) {
			return nil, &ErrorResourceNotFound
		}
		logger.Error("Failed to get resource", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	return &resource, nil
}

// GetResourceList retrieves a paginated list of resources.
func (rs *resourceService) GetResourceList(
	resourceServerID string, parentID *string, limit, offset int,
) (*ResourceList, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}
	if resourceServerID == "" {
		return nil, &ErrorMissingID
	}
	// Validate resource server exists
	exists, err := rs.resourceStore.IsResourceServerExist(resourceServerID)
	if err != nil {
		logger.Error("Failed to check resource server", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if !exists {
		return nil, &ErrorResourceServerNotFound
	}

	var totalCount int
	var resources []Resource

	// Handle different parent filtering modes
	if parentID == nil || *parentID == "" {
		totalCount, err = rs.resourceStore.GetResourceListCountByParent(resourceServerID, parentID)
		if err != nil {
			logger.Error("Failed to get top-level resource count", log.Error(err))
			return nil, &serviceerror.InternalServerError
		}
		resources, err = rs.resourceStore.GetResourceListByParent(resourceServerID, parentID, limit, offset)
	} else {
		// ParentID specified - validate and filter by that parent
		var exists bool
		exists, err = rs.resourceStore.IsResourceExist(*parentID, resourceServerID)
		if err != nil {
			logger.Error("Failed to check resource existence", log.Error(err))
			return nil, &serviceerror.InternalServerError
		}
		if !exists {
			return nil, &ErrorResourceNotFound
		}

		totalCount, err = rs.resourceStore.GetResourceListCountByParent(resourceServerID, parentID)
		if err != nil {
			logger.Error("Failed to get resource count by parent", log.Error(err))
			return nil, &serviceerror.InternalServerError
		}
		resources, err = rs.resourceStore.GetResourceListByParent(resourceServerID, parentID, limit, offset)
	}

	if err != nil {
		logger.Error("Failed to list resources", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	baseURL := fmt.Sprintf("/resource-servers/%s/resources", resourceServerID)
	response := &ResourceList{
		TotalResults: totalCount,
		Resources:    resources,
		StartIndex:   offset + 1,
		Count:        len(resources),
		Links:        buildPaginationLinks(baseURL, limit, offset, totalCount),
	}

	return response, nil
}

// UpdateResource updates a resource.
func (rs *resourceService) UpdateResource(
	resourceServerID, id string, resource Resource,
) (*Resource, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if id == "" || resourceServerID == "" {
		return nil, &ErrorMissingID
	}

	// Validate resource server exists
	rsExists, err := rs.resourceStore.IsResourceServerExist(resourceServerID)
	if err != nil {
		logger.Error("Failed to check resource server", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if !rsExists {
		return nil, &ErrorResourceServerNotFound
	}

	// Validate resource exists
	exists, err := rs.resourceStore.IsResourceExist(id, resourceServerID)
	if err != nil {
		logger.Error("Failed to check resource existence", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if !exists {
		return nil, &ErrorResourceNotFound
	}

	// Get current resource to preserve immutable handle and parent
	currentResource, err := rs.resourceStore.GetResource(id, resourceServerID)
	if err != nil {
		logger.Error("Failed to get current resource", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	// Update only mutable fields (name and description)
	// Note: handle and parent are immutable and preserved from current resource
	updateResource := Resource{
		Name:        resource.Name,          // Mutable
		Handle:      currentResource.Handle, // Immutable - preserve
		Description: resource.Description,
		Parent:      currentResource.Parent, // Immutable - preserve
	}

	if err := rs.resourceStore.UpdateResource(id, resourceServerID, updateResource); err != nil {
		logger.Error("Failed to update resource", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	updatedResource := &Resource{
		ID:          id,
		Name:        updateResource.Name,
		Handle:      updateResource.Handle,
		Description: updateResource.Description,
		Parent:      updateResource.Parent,
	}

	return updatedResource, nil
}

// DeleteResource deletes a resource.
func (rs *resourceService) DeleteResource(resourceServerID, id string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if id == "" || resourceServerID == "" {
		return &ErrorMissingID
	}

	// Validate resource server exists
	rsExists, err := rs.resourceStore.IsResourceServerExist(resourceServerID)
	if err != nil {
		logger.Error("Failed to check resource server", log.Error(err))
		return &serviceerror.InternalServerError
	}
	if !rsExists {
		return nil // Idempotent delete
	}

	// Check resource exists
	exists, err := rs.resourceStore.IsResourceExist(id, resourceServerID)
	if err != nil {
		logger.Error("Failed to check resource existence", log.Error(err))
		return &serviceerror.InternalServerError
	}
	if !exists {
		return nil // Idempotent delete
	}

	// Check for dependencies
	hasDeps, err := rs.resourceStore.CheckResourceHasDependencies(id)
	if err != nil {
		logger.Error("Failed to check dependencies", log.Error(err))
		return &serviceerror.InternalServerError
	}
	if hasDeps {
		return &ErrorCannotDelete
	}

	if err := rs.resourceStore.DeleteResource(id, resourceServerID); err != nil {
		logger.Error("Failed to delete resource", log.Error(err))
		return &serviceerror.InternalServerError
	}

	return nil
}

// Action Methods

// CreateActionAtResourceServer creates an action at resource server level.
func (rs *resourceService) CreateActionAtResourceServer(
	resourceServerID string, action Action,
) (*Action, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	// Validate resource server exists and get internal ID
	resourceServerInternalID, err := rs.resourceStore.CheckResourceServerExistAndGetInternalID(resourceServerID)
	if err != nil {
		if errors.Is(err, ErrResourceServerNotFound) {
			return nil, &ErrorResourceServerNotFound
		}
		logger.Error("Failed to check resource server", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	if err := rs.validateActionCreate(action); err != nil {
		return nil, err
	}

	// Check handle uniqueness at resource server level
	handleExists, err := rs.resourceStore.CheckActionHandleExists(resourceServerID, nil, action.Handle)
	if err != nil {
		logger.Error("Failed to check action handle", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if handleExists {
		return nil, &ErrorHandleConflict
	}

	id := utils.GenerateUUIDv7()
	if err := rs.resourceStore.CreateAction(id, resourceServerInternalID, nil, action); err != nil {
		logger.Error("Failed to create action", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	createdAction := &Action{
		ID:          id,
		Name:        action.Name,
		Handle:      action.Handle,
		Description: action.Description,
		ResourceID:  nil,
	}

	return createdAction, nil
}

// CreateActionAtResource creates an action at resource level.
func (rs *resourceService) CreateActionAtResource(
	resourceServerID, resourceID string, action Action,
) (*Action, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	// Validate resource server exists and get internal ID
	resourceServerInternalID, err := rs.resourceStore.CheckResourceServerExistAndGetInternalID(resourceServerID)
	if err != nil {
		if errors.Is(err, ErrResourceServerNotFound) {
			return nil, &ErrorResourceServerNotFound
		}
		logger.Error("Failed to check resource server", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	// Validate resource exists and get internal ID
	resourceInternalID, err := rs.resourceStore.CheckResourceExistAndGetInternalID(resourceID, resourceServerID)
	if err != nil {
		if errors.Is(err, ErrResourceNotFound) {
			return nil, &ErrorResourceNotFound
		}
		logger.Error("Failed to check resource", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	if err := rs.validateActionCreate(action); err != nil {
		return nil, err
	}

	// Check handle uniqueness at resource level
	handleExists, err := rs.resourceStore.CheckActionHandleExists(resourceServerID, &resourceID, action.Handle)
	if err != nil {
		logger.Error("Failed to check action handle", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if handleExists {
		return nil, &ErrorHandleConflict
	}

	id := utils.GenerateUUIDv7()
	if err := rs.resourceStore.CreateAction(id, resourceServerInternalID, &resourceInternalID, action); err != nil {
		logger.Error("Failed to create action", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	createdAction := &Action{
		ID:          id,
		Name:        action.Name,
		Handle:      action.Handle,
		Description: action.Description,
		ResourceID:  &resourceID,
	}

	return createdAction, nil
}

// GetActionAtResourceServer retrieves an action at resource server level.
func (rs *resourceService) GetActionAtResourceServer(
	resourceServerID, id string,
) (*Action, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if id == "" || resourceServerID == "" {
		return nil, &ErrorMissingID
	}

	// Validate resource server exists
	rsExists, err := rs.resourceStore.IsResourceServerExist(resourceServerID)
	if err != nil {
		logger.Error("Failed to check resource server", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if !rsExists {
		return nil, &ErrorResourceServerNotFound
	}

	action, err := rs.resourceStore.GetAction(id, resourceServerID, nil)
	if err != nil {
		if errors.Is(err, ErrActionNotFound) {
			return nil, &ErrorActionNotFound
		}
		logger.Error("Failed to get action", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	return &action, nil
}

// GetActionAtResource retrieves an action at resource level.
func (rs *resourceService) GetActionAtResource(
	resourceServerID, resourceID, id string,
) (*Action, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if id == "" || resourceServerID == "" || resourceID == "" {
		return nil, &ErrorMissingID
	}

	// Validate resource server exists
	rsExists, err := rs.resourceStore.IsResourceServerExist(resourceServerID)
	if err != nil {
		logger.Error("Failed to check resource server", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if !rsExists {
		return nil, &ErrorResourceServerNotFound
	}

	// Validate resource exists
	resExists, err := rs.resourceStore.IsResourceExist(resourceID, resourceServerID)
	if err != nil {
		logger.Error("Failed to check resource", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if !resExists {
		return nil, &ErrorResourceNotFound
	}

	action, err := rs.resourceStore.GetAction(id, resourceServerID, &resourceID)
	if err != nil {
		if errors.Is(err, ErrActionNotFound) {
			return nil, &ErrorActionNotFound
		}
		logger.Error("Failed to get action", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	return &action, nil
}

// GetActionListAtResourceServer retrieves actions at resource server level.
func (rs *resourceService) GetActionListAtResourceServer(
	resourceServerID string, limit, offset int,
) (*ActionList, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}

	if resourceServerID == "" {
		return nil, &ErrorMissingID
	}

	// Validate resource server exists
	exists, err := rs.resourceStore.IsResourceServerExist(resourceServerID)
	if err != nil {
		logger.Error("Failed to check resource server", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if !exists {
		return nil, &ErrorResourceServerNotFound
	}

	totalCount, err := rs.resourceStore.GetActionListCountAtResourceServer(resourceServerID)
	if err != nil {
		logger.Error("Failed to get action count", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	actions, err := rs.resourceStore.GetActionListAtResourceServer(resourceServerID, limit, offset)
	if err != nil {
		logger.Error("Failed to list actions", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	baseURL := fmt.Sprintf("/resource-servers/%s/actions", resourceServerID)
	response := &ActionList{
		TotalResults: totalCount,
		Actions:      actions,
		StartIndex:   offset + 1,
		Count:        len(actions),
		Links:        buildPaginationLinks(baseURL, limit, offset, totalCount),
	}

	return response, nil
}

// GetActionListAtResource retrieves actions at resource level.
func (rs *resourceService) GetActionListAtResource(
	resourceServerID, resourceID string, limit, offset int,
) (*ActionList, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}

	if resourceServerID == "" || resourceID == "" {
		return nil, &ErrorMissingID
	}

	// Validate resource server exists
	rsExists, err := rs.resourceStore.IsResourceServerExist(resourceServerID)
	if err != nil {
		logger.Error("Failed to check resource server", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if !rsExists {
		return nil, &ErrorResourceServerNotFound
	}

	// Validate resource exists
	resExists, err := rs.resourceStore.IsResourceExist(resourceID, resourceServerID)
	if err != nil {
		logger.Error("Failed to check resource", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if !resExists {
		return nil, &ErrorResourceNotFound
	}

	totalCount, err := rs.resourceStore.GetActionListCountAtResource(resourceServerID, resourceID)
	if err != nil {
		logger.Error("Failed to get action count", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	actions, err := rs.resourceStore.GetActionListAtResource(resourceServerID, resourceID, limit, offset)
	if err != nil {
		logger.Error("Failed to list actions", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	baseURL := fmt.Sprintf("/resource-servers/%s/resources/%s/actions", resourceServerID, resourceID)
	response := &ActionList{
		TotalResults: totalCount,
		Actions:      actions,
		StartIndex:   offset + 1,
		Count:        len(actions),
		Links:        buildPaginationLinks(baseURL, limit, offset, totalCount),
	}

	return response, nil
}

// UpdateActionAtResourceServer updates an action at resource server level (description only - name is immutable).
func (rs *resourceService) UpdateActionAtResourceServer(
	resourceServerID, id string, action Action,
) (*Action, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if id == "" || resourceServerID == "" {
		return nil, &ErrorMissingID
	}

	// Validate resource server exists
	rsExists, err := rs.resourceStore.IsResourceServerExist(resourceServerID)
	if err != nil {
		logger.Error("Failed to check resource server", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if !rsExists {
		return nil, &ErrorResourceServerNotFound
	}

	// Get current action at resource server level
	currentAction, err := rs.resourceStore.GetAction(id, resourceServerID, nil)
	if err != nil {
		if errors.Is(err, ErrActionNotFound) {
			return nil, &ErrorActionNotFound
		}
		logger.Error("Failed to get action", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	// Update only name and description (handle is immutable)
	updateAction := Action{
		Name:        action.Name,
		Handle:      currentAction.Handle, // Immutable - preserve
		Description: action.Description,
		ResourceID:  nil,
	}

	if err := rs.resourceStore.UpdateAction(id, resourceServerID, nil, updateAction); err != nil {
		logger.Error("Failed to update action", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	updatedAction := &Action{
		ID:          id,
		Name:        updateAction.Name,
		Handle:      updateAction.Handle,
		Description: updateAction.Description,
		ResourceID:  nil,
	}

	return updatedAction, nil
}

// UpdateActionAtResource updates an action at resource level (description only - name is immutable).
func (rs *resourceService) UpdateActionAtResource(
	resourceServerID, resourceID, id string, action Action,
) (*Action, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if id == "" || resourceServerID == "" || resourceID == "" {
		return nil, &ErrorMissingID
	}

	// Validate resource server exists
	rsExists, err := rs.resourceStore.IsResourceServerExist(resourceServerID)
	if err != nil {
		logger.Error("Failed to check resource server", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if !rsExists {
		return nil, &ErrorResourceServerNotFound
	}

	// Validate resource exists
	resExists, err := rs.resourceStore.IsResourceExist(resourceID, resourceServerID)
	if err != nil {
		logger.Error("Failed to check resource", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if !resExists {
		return nil, &ErrorResourceNotFound
	}

	// Get current action at resource level
	currentAction, err := rs.resourceStore.GetAction(id, resourceServerID, &resourceID)
	if err != nil {
		if errors.Is(err, ErrActionNotFound) {
			return nil, &ErrorActionNotFound
		}
		logger.Error("Failed to get action", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	// Update only name and description (handle is immutable)
	updateAction := Action{
		Name:        action.Name,
		Handle:      currentAction.Handle, // Immutable - preserve
		Description: action.Description,
		ResourceID:  &resourceID,
	}

	if err := rs.resourceStore.UpdateAction(id, resourceServerID, &resourceID, updateAction); err != nil {
		logger.Error("Failed to update action", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	updatedAction := &Action{
		ID:          id,
		Name:        updateAction.Name,
		Handle:      updateAction.Handle,
		Description: updateAction.Description,
		ResourceID:  &resourceID,
	}

	return updatedAction, nil
}

// DeleteActionAtResourceServer deletes an action at resource server level.
func (rs *resourceService) DeleteActionAtResourceServer(resourceServerID, id string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if id == "" || resourceServerID == "" {
		return &ErrorMissingID
	}

	// Validate resource server exists
	rsExists, err := rs.resourceStore.IsResourceServerExist(resourceServerID)
	if err != nil {
		logger.Error("Failed to check resource server", log.Error(err))
		return &serviceerror.InternalServerError
	}
	if !rsExists {
		return nil // Idempotent delete
	}

	// Check if action exists at resource server level (not at resource level)
	exists, err := rs.resourceStore.IsActionExist(id, resourceServerID, nil)
	if err != nil {
		logger.Error("Failed to check action existence", log.Error(err))
		return &serviceerror.InternalServerError
	}
	if !exists {
		return nil // Idempotent delete
	}

	if err := rs.resourceStore.DeleteAction(id, resourceServerID, nil); err != nil {
		logger.Error("Failed to delete action", log.Error(err))
		return &serviceerror.InternalServerError
	}

	return nil
}

// DeleteActionAtResource deletes an action at resource level.
func (rs *resourceService) DeleteActionAtResource(
	resourceServerID, resourceID, id string,
) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if id == "" || resourceServerID == "" || resourceID == "" {
		return &ErrorMissingID
	}

	// Validate resource server exists
	rsExists, err := rs.resourceStore.IsResourceServerExist(resourceServerID)
	if err != nil {
		logger.Error("Failed to check resource server", log.Error(err))
		return &serviceerror.InternalServerError
	}
	if !rsExists {
		return nil // Idempotent delete
	}

	// Validate resource exists
	resExists, err := rs.resourceStore.IsResourceExist(resourceID, resourceServerID)
	if err != nil {
		logger.Error("Failed to check resource", log.Error(err))
		return &serviceerror.InternalServerError
	}
	if !resExists {
		return nil // Idempotent delete
	}

	// Check if action exists at resource level
	exists, err := rs.resourceStore.IsActionExist(id, resourceServerID, &resourceID)
	if err != nil {
		logger.Error("Failed to check action existence", log.Error(err))
		return &serviceerror.InternalServerError
	}
	if !exists {
		return nil // Idempotent delete
	}

	if err := rs.resourceStore.DeleteAction(id, resourceServerID, &resourceID); err != nil {
		logger.Error("Failed to delete action", log.Error(err))
		return &serviceerror.InternalServerError
	}

	return nil
}

// Validation helper methods

func (rs *resourceService) validateResourceServerCreate(resourceServer ResourceServer) *serviceerror.ServiceError {
	if resourceServer.Name == "" {
		return &ErrorInvalidRequestFormat
	}
	if resourceServer.OrganizationUnitID == "" {
		return &ErrorInvalidRequestFormat
	}
	return nil
}

func (rs *resourceService) validateResourceServerUpdate(resourceServer ResourceServer) *serviceerror.ServiceError {
	if resourceServer.Name == "" {
		return &ErrorInvalidRequestFormat
	}
	if resourceServer.OrganizationUnitID == "" {
		return &ErrorInvalidRequestFormat
	}
	return nil
}

func (rs *resourceService) validateResourceCreate(resource Resource) *serviceerror.ServiceError {
	if resource.Name == "" {
		return &ErrorInvalidRequestFormat
	}
	if resource.Handle == "" {
		return &ErrorInvalidRequestFormat
	}
	return nil
}

func (rs *resourceService) validateActionCreate(action Action) *serviceerror.ServiceError {
	if action.Name == "" {
		return &ErrorInvalidRequestFormat
	}
	if action.Handle == "" {
		return &ErrorInvalidRequestFormat
	}
	return nil
}

func validatePaginationParams(limit, offset int) *serviceerror.ServiceError {
	if limit < 1 || limit > serverconst.MaxPageSize {
		return &ErrorInvalidLimit
	}
	if offset < 0 {
		return &ErrorInvalidOffset
	}
	return nil
}

func buildPaginationLinks(base string, limit, offset, totalCount int) []Link {
	links := make([]Link, 0)

	if offset > 0 {
		links = append(links, Link{
			Href: fmt.Sprintf("%s?offset=0&limit=%d", base, limit),
			Rel:  "first",
		})

		prevOffset := offset - limit
		if prevOffset < 0 {
			prevOffset = 0
		}
		links = append(links, Link{
			Href: fmt.Sprintf("%s?offset=%d&limit=%d", base, prevOffset, limit),
			Rel:  "prev",
		})
	}

	if offset+limit < totalCount {
		nextOffset := offset + limit
		links = append(links, Link{
			Href: fmt.Sprintf("%s?offset=%d&limit=%d", base, nextOffset, limit),
			Rel:  "next",
		})
	}

	lastPageOffset := ((totalCount - 1) / limit) * limit
	if offset < lastPageOffset {
		links = append(links, Link{
			Href: fmt.Sprintf("%s?offset=%d&limit=%d", base, lastPageOffset, limit),
			Rel:  "last",
		})
	}

	return links
}
