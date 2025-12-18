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
	"strings"

	oupkg "github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/system/config"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
)

const (
	loggerComponentName = "ResourceMgtService"

	// validDelimiterCharacters defines the allowed characters for delimiters.
	// Allowed: . _ : - /
	validDelimiterCharacters = "._:-/"

	// validPermissionCharacters defines the allowed characters for permission strings.
	// Allowed: a-z A-Z 0-9 and delimiter characters (. _ : - /)
	validPermissionCharacters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" +
		validDelimiterCharacters
)

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
	CreateAction(resourceServerID string, resourceID *string, action Action) (*Action, *serviceerror.ServiceError)
	GetAction(resourceServerID string, resourceID *string, id string) (*Action, *serviceerror.ServiceError)
	GetActionList(
		resourceServerID string, resourceID *string, limit, offset int,
	) (*ActionList, *serviceerror.ServiceError)
	UpdateAction(
		resourceServerID string, resourceID *string, id string, action Action,
	) (*Action, *serviceerror.ServiceError)
	DeleteAction(resourceServerID string, resourceID *string, id string) *serviceerror.ServiceError
	ValidatePermissions(resourceServerID string, permissions []string) ([]string, *serviceerror.ServiceError)
}

// resourceService is the default implementation of ResourceServiceInterface.
type resourceService struct {
	logger           log.Logger
	resourceStore    resourceStoreInterface
	ouService        oupkg.OrganizationUnitServiceInterface
	defaultDelimiter string
}

// newResourceService creates a new instance of ResourceService.
func newResourceService(
	resourceStore resourceStoreInterface,
	ouService oupkg.OrganizationUnitServiceInterface,
) (ResourceServiceInterface, error) {
	// Load default delimiter from config
	defaultDelimiter := getDefaultDelimiter()
	if err := validateDelimiter(defaultDelimiter); err != nil {
		return nil, fmt.Errorf("configured permission delimiter is invalid")
	}

	return &resourceService{
		logger:           *log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName)),
		resourceStore:    resourceStore,
		ouService:        ouService,
		defaultDelimiter: defaultDelimiter,
	}, nil
}

// Resource Server Methods

// CreateResourceServer creates a new resource server.
func (rs *resourceService) CreateResourceServer(
	resourceServer ResourceServer,
) (*ResourceServer, *serviceerror.ServiceError) {
	rs.logger.Debug("Creating resource server", log.String("name", resourceServer.Name))

	if err := rs.validateResourceServerCreate(resourceServer); err != nil {
		return nil, err
	}

	// Validate organization unit exists
	_, svcErr := rs.ouService.GetOrganizationUnit(resourceServer.OrganizationUnitID)
	if svcErr != nil {
		if svcErr.Code == oupkg.ErrorOrganizationUnitNotFound.Code {
			rs.logger.Debug("Organization unit not found", log.String("ouID", resourceServer.OrganizationUnitID))
			return nil, &ErrorOrganizationUnitNotFound
		}
		rs.logger.Error("Failed to validate organization unit", log.String("error", svcErr.Error))
		return nil, &serviceerror.InternalServerError
	}

	// Check name uniqueness
	nameExists, err := rs.resourceStore.CheckResourceServerNameExists(resourceServer.Name)
	if err != nil {
		rs.logger.Error("Failed to check resource server name", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if nameExists {
		rs.logger.Debug("Resource server name already exists", log.String("name", resourceServer.Name))
		return nil, &ErrorNameConflict
	}

	// Check identifier uniqueness (if provided)
	if resourceServer.Identifier != "" {
		identifierExists, err := rs.resourceStore.CheckResourceServerIdentifierExists(resourceServer.Identifier)
		if err != nil {
			rs.logger.Error("Failed to check resource server identifier", log.Error(err))
			return nil, &serviceerror.InternalServerError
		}
		if identifierExists {
			rs.logger.Debug("Resource server identifier already exists",
				log.String("identifier", resourceServer.Identifier))
			return nil, &ErrorIdentifierConflict
		}
	}
	// Set default delimiter if not provided
	if resourceServer.Delimiter == "" {
		resourceServer.Delimiter = rs.defaultDelimiter
	}

	id, err := utils.GenerateUUIDv7()
	if err != nil {
		rs.logger.Error("Failed to generate UUID", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if err := rs.resourceStore.CreateResourceServer(id, resourceServer); err != nil {
		rs.logger.Error("Failed to create resource server", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	createdRS := &ResourceServer{
		ID:                 id,
		Name:               resourceServer.Name,
		Description:        resourceServer.Description,
		Identifier:         resourceServer.Identifier,
		OrganizationUnitID: resourceServer.OrganizationUnitID,
		Delimiter:          resourceServer.Delimiter,
	}

	rs.logger.Debug("Successfully created resource server", log.String("id", id))
	return createdRS, nil
}

// GetResourceServer retrieves a resource server by ID.
func (rs *resourceService) GetResourceServer(id string) (*ResourceServer, *serviceerror.ServiceError) {
	if id == "" {
		return nil, &ErrorMissingID
	}

	_, resourceServer, err := rs.resourceStore.GetResourceServer(id)
	if err != nil {
		if errors.Is(err, errResourceServerNotFound) {
			rs.logger.Debug("Resource server not found", log.String("id", id))
			return nil, &ErrorResourceServerNotFound
		}
		rs.logger.Error("Failed to get resource server", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	return &resourceServer, nil
}

// GetResourceServerList retrieves a paginated list of resource servers.
func (rs *resourceService) GetResourceServerList(limit, offset int) (*ResourceServerList, *serviceerror.ServiceError) {
	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}

	totalCount, err := rs.resourceStore.GetResourceServerListCount()
	if err != nil {
		rs.logger.Error("Failed to get resource server count", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	resourceServers, err := rs.resourceStore.GetResourceServerList(limit, offset)
	if err != nil {
		rs.logger.Error("Failed to list resource servers", log.Error(err))
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
	if id == "" {
		return nil, &ErrorMissingID
	}

	if err := rs.validateResourceServerUpdate(resourceServer); err != nil {
		return nil, err
	}

	_, existingResServer, err := rs.resourceStore.GetResourceServer(id)
	if err != nil {
		if errors.Is(err, errResourceServerNotFound) {
			rs.logger.Debug("Resource server not found", log.String("id", id))
			return nil, &ErrorResourceServerNotFound
		}
		rs.logger.Error("Failed to check resource server existence", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	// Preserve the immutable delimiter from existing record
	resourceServer.Delimiter = existingResServer.Delimiter

	// Validate organization unit
	_, svcErr := rs.ouService.GetOrganizationUnit(resourceServer.OrganizationUnitID)
	if svcErr != nil {
		if svcErr.Code == oupkg.ErrorOrganizationUnitNotFound.Code {
			return nil, &ErrorOrganizationUnitNotFound
		}
		return nil, &serviceerror.InternalServerError
	}

	// Check name uniqueness, if changed
	if existingResServer.Name != resourceServer.Name {
		nameExists, err := rs.resourceStore.CheckResourceServerNameExists(resourceServer.Name)
		if err != nil {
			rs.logger.Error("Failed to check resource server name", log.Error(err))
			return nil, &serviceerror.InternalServerError
		}
		if nameExists {
			return nil, &ErrorNameConflict
		}
	}

	// Check identifier uniqueness, if provided and changed
	if resourceServer.Identifier != "" && existingResServer.Identifier != resourceServer.Identifier {
		identifierExists, err := rs.resourceStore.CheckResourceServerIdentifierExists(resourceServer.Identifier)
		if err != nil {
			rs.logger.Error("Failed to check resource server identifier", log.Error(err))
			return nil, &serviceerror.InternalServerError
		}
		if identifierExists {
			rs.logger.Debug("Resource server identifier already exists",
				log.String("identifier", resourceServer.Identifier))
			return nil, &ErrorIdentifierConflict
		}
	}

	if err := rs.resourceStore.UpdateResourceServer(id, resourceServer); err != nil {
		rs.logger.Error("Failed to update resource server", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	updatedRS := &ResourceServer{
		ID:                 id,
		Name:               resourceServer.Name,
		Description:        resourceServer.Description,
		Identifier:         resourceServer.Identifier,
		OrganizationUnitID: resourceServer.OrganizationUnitID,
		Delimiter:          resourceServer.Delimiter,
	}

	return updatedRS, nil
}

// DeleteResourceServer deletes a resource server.
func (rs *resourceService) DeleteResourceServer(id string) *serviceerror.ServiceError {
	if id == "" {
		return &ErrorMissingID
	}

	resServerInternalID, _, err := rs.resourceStore.GetResourceServer(id)
	if err != nil {
		if errors.Is(err, errResourceServerNotFound) {
			return nil // Idempotent delete
		}
		rs.logger.Error("Failed to check resource server existence", log.Error(err))
		return &serviceerror.InternalServerError
	}

	// Check for dependencies
	hasDeps, err := rs.resourceStore.CheckResourceServerHasDependencies(resServerInternalID)
	if err != nil {
		rs.logger.Error("Failed to check dependencies", log.Error(err))
		return &serviceerror.InternalServerError
	}
	if hasDeps {
		return &ErrorCannotDelete
	}

	if err := rs.resourceStore.DeleteResourceServer(id); err != nil {
		rs.logger.Error("Failed to delete resource server", log.Error(err))
		return &serviceerror.InternalServerError
	}

	return nil
}

// Resource Methods

// CreateResource creates a new resource.
func (rs *resourceService) CreateResource(
	resourceServerID string, resource Resource,
) (*Resource, *serviceerror.ServiceError) {
	// Validate resource server exists and get internal ID
	resServerInternalID, resourceServer, svcErr := rs.validateAndGetResourceServerInternalID(resourceServerID)
	if svcErr != nil {
		return nil, svcErr
	}

	if err := rs.validateResourceCreate(resource, resourceServer.Delimiter); err != nil {
		return nil, err
	}

	// Validate parent if specified and get internal ID
	var parentInternalID *int
	var parentResource *Resource
	if resource.Parent != nil {
		parentID, res, err := rs.resourceStore.GetResource(*resource.Parent, resServerInternalID)
		if err != nil {
			if errors.Is(err, errResourceNotFound) {
				return nil, &ErrorParentResourceNotFound
			}
			rs.logger.Error("Failed to check parent resource", log.Error(err))
			return nil, &serviceerror.InternalServerError
		}
		parentInternalID = &parentID
		parentResource = &res
	}

	// Check handle uniqueness under parent
	handleExists, err := rs.resourceStore.CheckResourceHandleExists(
		resServerInternalID, resource.Handle, parentInternalID,
	)
	if err != nil {
		rs.logger.Error("Failed to check resource handle", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if handleExists {
		return nil, &ErrorHandleConflict
	}

	// Derive permission string based on hierarchy
	resource.Permission = derivePermission(resourceServer, parentResource, resource.Handle)

	id, err := utils.GenerateUUIDv7()
	if err != nil {
		rs.logger.Error("Failed to generate UUID", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if err := rs.resourceStore.CreateResource(id, resServerInternalID, parentInternalID, resource); err != nil {
		rs.logger.Error("Failed to create resource", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	createdResource := &Resource{
		ID:          id,
		Name:        resource.Name,
		Handle:      resource.Handle,
		Description: resource.Description,
		Parent:      resource.Parent,
		Permission:  resource.Permission,
	}

	return createdResource, nil
}

// GetResource retrieves a resource by ID.
func (rs *resourceService) GetResource(resourceServerID, id string) (*Resource, *serviceerror.ServiceError) {
	if id == "" || resourceServerID == "" {
		return nil, &ErrorMissingID
	}

	// Validate resource server exists and get internal ID
	resServerInternalID, _, svcErr := rs.validateAndGetResourceServerInternalID(resourceServerID)
	if svcErr != nil {
		return nil, svcErr
	}

	_, resource, err := rs.resourceStore.GetResource(id, resServerInternalID)
	if err != nil {
		if errors.Is(err, errResourceNotFound) {
			return nil, &ErrorResourceNotFound
		}
		rs.logger.Error("Failed to get resource", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	return &resource, nil
}

// GetResourceList retrieves a paginated list of resources.
func (rs *resourceService) GetResourceList(
	resourceServerID string, parentID *string, limit, offset int,
) (*ResourceList, *serviceerror.ServiceError) {
	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}
	if resourceServerID == "" {
		return nil, &ErrorMissingID
	}
	// Validate resource server exists and get internal ID
	resServerInternalID, _, svcErr := rs.validateAndGetResourceServerInternalID(resourceServerID)
	if svcErr != nil {
		return nil, svcErr
	}

	var totalCount int
	var resources []Resource
	var parentInternalID *int

	// Resolve internal ID for parent if specified
	if parentID != nil {
		// ParentID specified - validate and get internal ID
		parentIntID, _, svcErr := rs.validateAndGetResource(*parentID, resServerInternalID)
		if svcErr != nil {
			return nil, svcErr
		}
		parentInternalID = &parentIntID
	}

	totalCount, err := rs.resourceStore.GetResourceListCountByParent(resServerInternalID, parentInternalID)
	if err != nil {
		rs.logger.Error("Failed to get top-level resource count", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	resources, err = rs.resourceStore.GetResourceListByParent(resServerInternalID, parentInternalID, limit, offset)
	if err != nil {
		rs.logger.Error("Failed to list resources", log.Error(err))
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
	if id == "" || resourceServerID == "" {
		return nil, &ErrorMissingID
	}

	// Validate resource server exists and get internal ID
	resServerInternalID, _, svcErr := rs.validateAndGetResourceServerInternalID(resourceServerID)
	if svcErr != nil {
		return nil, svcErr
	}

	// Validate resource exists
	_, currentResource, err := rs.resourceStore.GetResource(id, resServerInternalID)
	if err != nil {
		if errors.Is(err, errResourceNotFound) {
			return nil, &ErrorResourceNotFound
		}
		rs.logger.Error("Failed to check resource existence", log.Error(err))
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

	if err := rs.resourceStore.UpdateResource(id, resServerInternalID, updateResource); err != nil {
		rs.logger.Error("Failed to update resource", log.Error(err))
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
	if id == "" || resourceServerID == "" {
		return &ErrorMissingID
	}

	// Validate resource server exists and get internal ID
	resServerInternalID, _, err := rs.resourceStore.GetResourceServer(resourceServerID)
	if err != nil {
		if errors.Is(err, errResourceServerNotFound) {
			return nil // Idempotent delete
		}
		rs.logger.Error("Failed to check resource server", log.Error(err))
		return &serviceerror.InternalServerError
	}

	// Check resource exists and get internal ID
	resInternalID, _, err := rs.resourceStore.GetResource(id, resServerInternalID)
	if err != nil {
		if errors.Is(err, errResourceNotFound) {
			return nil // Idempotent delete
		}
		rs.logger.Error("Failed to check resource existence", log.Error(err))
		return &serviceerror.InternalServerError
	}

	// Check for dependencies
	hasDeps, err := rs.resourceStore.CheckResourceHasDependencies(resInternalID)
	if err != nil {
		rs.logger.Error("Failed to check dependencies", log.Error(err))
		return &serviceerror.InternalServerError
	}
	if hasDeps {
		return &ErrorCannotDelete
	}

	if err := rs.resourceStore.DeleteResource(id, resServerInternalID); err != nil {
		rs.logger.Error("Failed to delete resource", log.Error(err))
		return &serviceerror.InternalServerError
	}

	return nil
}

// Action Methods

// CreateAction creates an action.
// If resourceID is nil, creates action at resource server level.
// If resourceID is provided, creates action at resource level.
func (rs *resourceService) CreateAction(
	resourceServerID string, resourceID *string, action Action,
) (*Action, *serviceerror.ServiceError) {
	// Validate resource server exists and get internal ID
	resServerInternalID, resourceServer, svcErr := rs.validateAndGetResourceServerInternalID(resourceServerID)
	if svcErr != nil {
		return nil, svcErr
	}

	// Validate resource if provided and get internal ID
	var resInternalID *int
	var resource *Resource
	if resourceID != nil {
		resID, res, svcErr := rs.validateAndGetResource(*resourceID, resServerInternalID)
		if svcErr != nil {
			return nil, svcErr
		}
		resInternalID = &resID
		resource = &res
	}

	if err := rs.validateActionCreate(action, resourceServer.Delimiter); err != nil {
		return nil, err
	}

	// Check handle uniqueness
	handleExists, err := rs.resourceStore.CheckActionHandleExists(resServerInternalID, resInternalID, action.Handle)
	if err != nil {
		rs.logger.Error("Failed to check action handle", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if handleExists {
		return nil, &ErrorHandleConflict
	}

	// Derive permission string based on hierarchy
	action.Permission = derivePermission(resourceServer, resource, action.Handle)

	id, err := utils.GenerateUUIDv7()
	if err != nil {
		rs.logger.Error("Failed to generate UUID", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if err := rs.resourceStore.CreateAction(id, resServerInternalID, resInternalID, action); err != nil {
		rs.logger.Error("Failed to create action", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	createdAction := &Action{
		ID:          id,
		Name:        action.Name,
		Handle:      action.Handle,
		Description: action.Description,
		Permission:  action.Permission,
	}
	return createdAction, nil
}

// GetAction retrieves an action by ID.
// If resourceID is nil, retrieves action at resource server level.
// If resourceID is provided, retrieves action at resource level.
func (rs *resourceService) GetAction(
	resourceServerID string, resourceID *string, id string,
) (*Action, *serviceerror.ServiceError) {
	if id == "" || resourceServerID == "" {
		return nil, &ErrorMissingID
	}

	if resourceID != nil && *resourceID == "" {
		return nil, &ErrorMissingID
	}

	// Validate resource server exists and get internal ID
	resServerInternalID, _, svcErr := rs.validateAndGetResourceServerInternalID(resourceServerID)
	if svcErr != nil {
		return nil, svcErr
	}

	// Validate resource if provided and get internal ID
	var resInternalID *int
	if resourceID != nil {
		resID, _, svcErr := rs.validateAndGetResource(*resourceID, resServerInternalID)
		if svcErr != nil {
			return nil, svcErr
		}
		resInternalID = &resID
	}

	action, err := rs.resourceStore.GetAction(id, resServerInternalID, resInternalID)
	if err != nil {
		if errors.Is(err, errActionNotFound) {
			return nil, &ErrorActionNotFound
		}
		rs.logger.Error("Failed to get action", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	return &action, nil
}

// GetActionList retrieves a paginated list of actions.
// If resourceID is nil, retrieves actions at resource server level.
// If resourceID is provided, retrieves actions at resource level.
func (rs *resourceService) GetActionList(
	resourceServerID string, resourceID *string, limit, offset int,
) (*ActionList, *serviceerror.ServiceError) {
	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}

	if resourceServerID == "" {
		return nil, &ErrorMissingID
	}

	if resourceID != nil && *resourceID == "" {
		return nil, &ErrorMissingID
	}

	// Validate resource server exists and get internal ID
	resServerInternalID, _, svcErr := rs.validateAndGetResourceServerInternalID(resourceServerID)
	if svcErr != nil {
		return nil, svcErr
	}

	// Validate resource if provided and get internal ID
	var resInternalID *int
	if resourceID != nil {
		resID, _, svcErr := rs.validateAndGetResource(*resourceID, resServerInternalID)
		if svcErr != nil {
			return nil, svcErr
		}
		resInternalID = &resID
	}

	totalCount, err := rs.resourceStore.GetActionListCount(resServerInternalID, resInternalID)
	if err != nil {
		rs.logger.Error("Failed to get action count", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	actions, err := rs.resourceStore.GetActionList(resServerInternalID, resInternalID, limit, offset)
	if err != nil {
		rs.logger.Error("Failed to list actions", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	// Build base URL based on whether resource ID is provided
	var baseURL string
	if resourceID == nil {
		baseURL = fmt.Sprintf("/resource-servers/%s/actions", resourceServerID)
	} else {
		baseURL = fmt.Sprintf("/resource-servers/%s/resources/%s/actions", resourceServerID, *resourceID)
	}

	response := &ActionList{
		TotalResults: totalCount,
		Actions:      actions,
		StartIndex:   offset + 1,
		Count:        len(actions),
		Links:        buildPaginationLinks(baseURL, limit, offset, totalCount),
	}

	return response, nil
}

// UpdateAction updates an action.
// If resourceID is nil, updates action at resource server level.
// If resourceID is provided, updates action at resource level.
func (rs *resourceService) UpdateAction(
	resourceServerID string, resourceID *string, id string, action Action,
) (*Action, *serviceerror.ServiceError) {
	if id == "" || resourceServerID == "" {
		return nil, &ErrorMissingID
	}

	if resourceID != nil && *resourceID == "" {
		return nil, &ErrorMissingID
	}

	// Validate resource server exists and get internal ID
	resServerInternalID, _, svcErr := rs.validateAndGetResourceServerInternalID(resourceServerID)
	if svcErr != nil {
		return nil, svcErr
	}

	// Validate resource if provided and get internal ID
	var resInternalID *int
	if resourceID != nil {
		resID, _, svcErr := rs.validateAndGetResource(*resourceID, resServerInternalID)
		if svcErr != nil {
			return nil, svcErr
		}
		resInternalID = &resID
	}

	// Get current action to preserve immutable fields
	currentAction, err := rs.resourceStore.GetAction(id, resServerInternalID, resInternalID)
	if err != nil {
		if errors.Is(err, errActionNotFound) {
			return nil, &ErrorActionNotFound
		}
		rs.logger.Error("Failed to get action", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	// Update only name and description (handle is immutable)
	updateAction := Action{
		Name:        action.Name,
		Handle:      currentAction.Handle, // Immutable - preserve
		Description: action.Description,
	}

	if err := rs.resourceStore.UpdateAction(id, resServerInternalID, resInternalID, updateAction); err != nil {
		rs.logger.Error("Failed to update action", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	updatedAction := &Action{
		ID:          id,
		Name:        updateAction.Name,
		Handle:      updateAction.Handle,
		Description: updateAction.Description,
	}

	return updatedAction, nil
}

// DeleteAction deletes an action.
// If resourceID is nil, deletes action at resource server level.
// If resourceID is provided, deletes action at resource level.
func (rs *resourceService) DeleteAction(
	resourceServerID string, resourceID *string, id string,
) *serviceerror.ServiceError {
	if id == "" || resourceServerID == "" {
		return &ErrorMissingID
	}

	if resourceID != nil && *resourceID == "" {
		return &ErrorMissingID
	}

	// Validate resource server exists and get internal ID
	resServerInternalID, _, svcErr := rs.validateAndGetResourceServerInternalID(resourceServerID)
	if svcErr != nil {
		if svcErr.Code == ErrorResourceServerNotFound.Code {
			return nil // Idempotent delete
		}
		return svcErr
	}

	// Validate resource if provided and get internal ID
	var resInternalID *int
	if resourceID != nil {
		resID, _, svcErr := rs.validateAndGetResource(*resourceID, resServerInternalID)
		if svcErr != nil {
			if svcErr.Code == ErrorResourceNotFound.Code {
				return nil // Idempotent delete
			}
			return svcErr
		}
		resInternalID = &resID
	}

	// Check if action exists
	exists, err := rs.resourceStore.IsActionExist(id, resServerInternalID, resInternalID)
	if err != nil {
		rs.logger.Error("Failed to check action existence", log.Error(err))
		return &serviceerror.InternalServerError
	}
	if !exists {
		return nil // Idempotent delete
	}

	if err := rs.resourceStore.DeleteAction(id, resServerInternalID, resInternalID); err != nil {
		rs.logger.Error("Failed to delete action", log.Error(err))
		return &serviceerror.InternalServerError
	}

	return nil
}

// ValidatePermissions checks if permissions exist for a given resource server.
// Returns array of invalid permissions (empty if all valid).
func (rs *resourceService) ValidatePermissions(
	resourceServerID string,
	permissions []string,
) ([]string, *serviceerror.ServiceError) {
	rs.logger.Debug("Validating permissions",
		log.String("resourceServerId", resourceServerID),
		log.Int("permissionCount", len(permissions)))

	if len(permissions) == 0 {
		return []string{}, nil
	}

	// Validate resource server exists
	resServerInternalID, _, err := rs.resourceStore.GetResourceServer(resourceServerID)
	if err != nil {
		if !errors.Is(err, errResourceServerNotFound) {
			rs.logger.Error("Failed to validate resource server existence",
				log.String("resourceServerId", resourceServerID),
				log.Error(err))
			return nil, &serviceerror.InternalServerError
		}
	}
	if resServerInternalID == 0 {
		rs.logger.Debug("Resource server not found",
			log.String("resourceServerId", resourceServerID))
		// Return all permissions as invalid if resource server doesn't exist
		return permissions, nil
	}

	// Call store to validate permissions
	invalidPermissions, storeErr := rs.resourceStore.ValidatePermissions(resServerInternalID, permissions)
	if storeErr != nil {
		rs.logger.Error("Failed to validate permissions in store",
			log.String("resourceServerId", resourceServerID),
			log.Error(storeErr))
		return nil, &serviceerror.InternalServerError
	}

	return invalidPermissions, nil
}

// Validation helper methods

// validateAndGetResourceServerInternalID validates resource server exists and returns its internal ID.
func (rs *resourceService) validateAndGetResourceServerInternalID(
	resourceServerID string,
) (int, ResourceServer, *serviceerror.ServiceError) {
	resServerInternalID, resourceServer, err := rs.resourceStore.GetResourceServer(resourceServerID)
	if err != nil {
		if errors.Is(err, errResourceServerNotFound) {
			return 0, ResourceServer{}, &ErrorResourceServerNotFound
		}
		rs.logger.Error("Failed to check resource server", log.Error(err))
		return 0, ResourceServer{}, &serviceerror.InternalServerError
	}
	return resServerInternalID, resourceServer, nil
}

// validateAndGetResource validates resource exists and returns its internal ID.
func (rs *resourceService) validateAndGetResource(
	resourceID string,
	resourceServerInternalID int,
) (int, Resource, *serviceerror.ServiceError) {
	resInternalID, resource, err := rs.resourceStore.GetResource(resourceID, resourceServerInternalID)
	if err != nil {
		if errors.Is(err, errResourceNotFound) {
			return 0, Resource{}, &ErrorResourceNotFound
		}
		rs.logger.Error("Failed to check resource", log.Error(err))
		return 0, Resource{}, &serviceerror.InternalServerError
	}
	return resInternalID, resource, nil
}

// validateResourceServerCreate validates the input for creating a resource server.
func (rs *resourceService) validateResourceServerCreate(resourceServer ResourceServer) *serviceerror.ServiceError {
	if resourceServer.Name == "" {
		return &ErrorInvalidRequestFormat
	}
	if resourceServer.OrganizationUnitID == "" {
		return &ErrorInvalidRequestFormat
	}
	if resourceServer.Delimiter != "" {
		if err := validateDelimiter(resourceServer.Delimiter); err != nil {
			return err
		}
	}
	return nil
}

// validateResourceServerUpdate validates the input for updating a resource server.
func (rs *resourceService) validateResourceServerUpdate(resourceServer ResourceServer) *serviceerror.ServiceError {
	if resourceServer.Name == "" {
		return &ErrorInvalidRequestFormat
	}
	if resourceServer.OrganizationUnitID == "" {
		return &ErrorInvalidRequestFormat
	}
	return nil
}

// validateResourceCreate validates the input for creating a resource.
func (rs *resourceService) validateResourceCreate(resource Resource, delimiter string) *serviceerror.ServiceError {
	if resource.Name == "" {
		return &ErrorInvalidRequestFormat
	}
	if resource.Handle == "" {
		return &ErrorInvalidRequestFormat
	}
	// Validate handle
	if err := validateHandle(resource.Handle, delimiter); err != nil {
		return err
	}
	return nil
}

// validateActionCreate validates the input for creating an action.
func (rs *resourceService) validateActionCreate(action Action, delimiter string) *serviceerror.ServiceError {
	if action.Name == "" {
		return &ErrorInvalidRequestFormat
	}
	if action.Handle == "" {
		return &ErrorInvalidRequestFormat
	}
	// Validate handle
	if err := validateHandle(action.Handle, delimiter); err != nil {
		return err
	}
	return nil
}

// validatePaginationParams validates pagination parameters.
func validatePaginationParams(limit, offset int) *serviceerror.ServiceError {
	if limit < 1 || limit > serverconst.MaxPageSize {
		return &ErrorInvalidLimit
	}
	if offset < 0 {
		return &ErrorInvalidOffset
	}

	return nil
}

// buildPaginationLinks constructs pagination links for a paginated response.
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

// isValidPermissionCharacter checks if a character is valid for permission strings.
// Allowed characters: a-z A-Z 0-9 . _ : - /
func isValidPermissionCharacter(c rune) bool {
	return strings.ContainsRune(validPermissionCharacters, c)
}

// validateDelimiter validates delimiter is a single valid delimiter character.
func validateDelimiter(delimiter string) *serviceerror.ServiceError {
	if len(delimiter) != 1 {
		return &ErrorInvalidDelimiter
	}
	if !strings.ContainsRune(validDelimiterCharacters, rune(delimiter[0])) {
		return &ErrorInvalidDelimiter
	}
	return nil
}

// validateHandle validates a handle string.
func validateHandle(handle string, delimiter string) *serviceerror.ServiceError {
	if len(handle) > 100 {
		return &ErrorInvalidHandle
	}
	for _, c := range handle {
		if !isValidPermissionCharacter(c) {
			return &ErrorInvalidHandle
		}
		if string(c) == delimiter {
			return &ErrorDelimiterInHandle
		}
	}
	return nil
}

// getDefaultDelimiter returns the default delimiter from configuration.
func getDefaultDelimiter() string {
	delimiter := config.GetThunderRuntime().Config.Resource.DefaultDelimiter
	if delimiter == "" {
		return ":" // Fallback default if not configured
	}
	return delimiter
}

// derivePermission builds permission string for a resource based on parent hierarchy.
func derivePermission(
	resourceServer ResourceServer,
	parentResource *Resource,
	handle string,
) string {
	if parentResource != nil {
		// Build permission: parent_permission + delimiter + handle
		return parentResource.Permission + resourceServer.Delimiter + handle
	}
	return handle // Top-level resource - permission is just the handle
}
