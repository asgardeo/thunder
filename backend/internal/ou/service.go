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

// Package ou handles the organization unit management operations.
package ou

import (
	"context"
	"errors"
	"fmt"
	"strings"

	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/security"
	"github.com/asgardeo/thunder/internal/system/sysauthz"
	"github.com/asgardeo/thunder/internal/system/utils"
)

const loggerComponentNameService = "OrganizationUnitService"

// OrganizationUnitServiceInterface defines the interface for organization unit service operations.
type OrganizationUnitServiceInterface interface {
	GetOrganizationUnitList(
		ctx context.Context, limit, offset int,
	) (*OrganizationUnitListResponse, *serviceerror.ServiceError)
	CreateOrganizationUnit(
		ctx context.Context, request OrganizationUnitRequest,
	) (OrganizationUnit, *serviceerror.ServiceError)
	GetOrganizationUnit(ctx context.Context, id string) (OrganizationUnit, *serviceerror.ServiceError)
	GetOrganizationUnitByPath(ctx context.Context, handlePath string) (OrganizationUnit, *serviceerror.ServiceError)
	IsOrganizationUnitExists(ctx context.Context, id string) (bool, *serviceerror.ServiceError)
	IsOrganizationUnitDeclarative(id string) bool
	IsParent(ctx context.Context, parentID, childID string) (bool, *serviceerror.ServiceError)
	UpdateOrganizationUnit(
		ctx context.Context, id string, request OrganizationUnitRequest,
	) (OrganizationUnit, *serviceerror.ServiceError)
	UpdateOrganizationUnitByPath(
		ctx context.Context, handlePath string, request OrganizationUnitRequest,
	) (OrganizationUnit, *serviceerror.ServiceError)
	DeleteOrganizationUnit(ctx context.Context, id string) *serviceerror.ServiceError
	DeleteOrganizationUnitByPath(ctx context.Context, handlePath string) *serviceerror.ServiceError
	GetOrganizationUnitChildren(
		ctx context.Context, id string, limit, offset int,
	) (*OrganizationUnitListResponse, *serviceerror.ServiceError)
	GetOrganizationUnitChildrenByPath(
		ctx context.Context, handlePath string, limit, offset int,
	) (*OrganizationUnitListResponse, *serviceerror.ServiceError)
	GetOrganizationUnitUsers(
		ctx context.Context, id string, limit, offset int,
	) (*UserListResponse, *serviceerror.ServiceError)
	GetOrganizationUnitUsersByPath(
		ctx context.Context, handlePath string, limit, offset int,
	) (*UserListResponse, *serviceerror.ServiceError)
	GetOrganizationUnitGroups(
		ctx context.Context, id string, limit, offset int,
	) (*GroupListResponse, *serviceerror.ServiceError)
	GetOrganizationUnitGroupsByPath(
		ctx context.Context, handlePath string, limit, offset int,
	) (*GroupListResponse, *serviceerror.ServiceError)
}

// OrganizationUnitService provides organization unit management operations.
type organizationUnitService struct {
	authzService sysauthz.SystemAuthorizationServiceInterface
	ouStore      organizationUnitStoreInterface
}

// newOrganizationUnitService creates a new instance of OrganizationUnitService.
func newOrganizationUnitService(
	authzService sysauthz.SystemAuthorizationServiceInterface,
	ouStore organizationUnitStoreInterface,
) OrganizationUnitServiceInterface {
	return &organizationUnitService{
		authzService: authzService,
		ouStore:      ouStore,
	}
}

// GetOrganizationUnitList retrieves a list of organization units.
// limit should be a positive integer and offset should be non-negative.
func (ous *organizationUnitService) GetOrganizationUnitList(ctx context.Context, limit, offset int) (
	*OrganizationUnitListResponse, *serviceerror.ServiceError,
) {
	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}

	// Resolve the set of organization units the caller is authorized to see.
	accessible, svcErr := ous.authzService.GetAccessibleResources(
		ctx, security.ActionListOUs, security.ResourceTypeOU)
	if svcErr != nil {
		return nil, &ErrorInternalServerError
	}

	// Unfiltered path: the caller can see all organization units.
	if accessible.AllAllowed {
		return ous.listAllOrganizationUnits(limit, offset)
	}

	// Filtered path: the caller has a restricted set of accessible organization units.
	return ous.listAccessibleOrganizationUnits(accessible.IDs, limit, offset)
}

// listAllOrganizationUnits retrieves organization units without authorization filtering.
func (ous *organizationUnitService) listAllOrganizationUnits(
	limit, offset int,
) (*OrganizationUnitListResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentNameService))
	totalCount, err := ous.ouStore.GetOrganizationUnitListCount()
	if err != nil {
		logger.Error("Failed to get organization unit count", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	ouList, err := ous.ouStore.GetOrganizationUnitList(limit, offset)
	if err != nil {
		// Check if it's a limit exceeded error
		if errors.Is(err, ErrResultLimitExceededInCompositeMode) {
			return nil, &ErrorResultLimitExceeded
		}
		logger.Error("Failed to list organization units", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	return &OrganizationUnitListResponse{
		TotalResults:      totalCount,
		OrganizationUnits: ouList,
		StartIndex:        offset + 1,
		Count:             len(ouList),
		Links:             buildPaginationLinks(limit, offset, totalCount),
	}, nil
}

// listAccessibleOrganizationUnits retrieves only the organization units the caller is authorized to access.
func (ous *organizationUnitService) listAccessibleOrganizationUnits(
	ids []string, limit, offset int,
) (*OrganizationUnitListResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentNameService))
	total := len(ids)
	if total == 0 {
		return &OrganizationUnitListResponse{
			TotalResults:      0,
			OrganizationUnits: []OrganizationUnitBasic{},
			StartIndex:        1,
			Count:             0,
			Links:             buildPaginationLinks(limit, offset, 0),
		}, nil
	}

	// Paginate the ID list before hitting the store.
	start := offset
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}
	pageIDs := ids[start:end]

	if len(pageIDs) == 0 {
		return &OrganizationUnitListResponse{
			TotalResults:      total,
			OrganizationUnits: []OrganizationUnitBasic{},
			StartIndex:        offset + 1,
			Count:             0,
			Links:             buildPaginationLinks(limit, offset, total),
		}, nil
	}

	// Fetch only the organization units needed for this page.
	pageOUs, err := ous.ouStore.GetOrganizationUnitsByIDs(pageIDs)
	if err != nil {
		logger.Error("Failed to get organization units by IDs", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	return &OrganizationUnitListResponse{
		TotalResults:      total,
		OrganizationUnits: pageOUs,
		StartIndex:        offset + 1,
		Count:             len(pageOUs),
		Links:             buildPaginationLinks(limit, offset, total),
	}, nil
}

// CreateOrganizationUnit creates a new organization unit.
func (ous *organizationUnitService) CreateOrganizationUnit(
	ctx context.Context, request OrganizationUnitRequest,
) (OrganizationUnit, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentNameService))
	logger.Debug("Creating organization unit", log.String("name", request.Name))

	// Fail if store is in declarative mode
	if isDeclarativeModeEnabled() {
		return OrganizationUnit{}, &ErrorCannotModifyDeclarativeResource
	}

	if err := ous.validateOUName(request.Name); err != nil {
		return OrganizationUnit{}, err
	}

	if err := ous.validateOUHandle(request.Handle); err != nil {
		return OrganizationUnit{}, err
	}

	if request.Parent != nil {
		if svcErr := ous.checkOUAccess(ctx, security.ActionCreateOU, *request.Parent); svcErr != nil {
			return OrganizationUnit{}, svcErr
		}
		exists, err := ous.ouStore.IsOrganizationUnitExists(*request.Parent)
		if err != nil {
			logger.Error("Failed to check parent organization unit existence", log.Error(err))
			return OrganizationUnit{}, &ErrorInternalServerError
		}
		if !exists {
			return OrganizationUnit{}, &ErrorParentOrganizationUnitNotFound
		}
	} else {
		if svcErr := ous.checkOUAccess(ctx, security.ActionCreateOU, ""); svcErr != nil {
			return OrganizationUnit{}, svcErr
		}
	}

	conflict, err := ous.ouStore.CheckOrganizationUnitNameConflict(request.Name, request.Parent)
	if err != nil {
		logger.Error("Failed to check organization unit name conflict", log.Error(err))
		return OrganizationUnit{}, &ErrorInternalServerError
	}
	if conflict {
		return OrganizationUnit{}, &ErrorOrganizationUnitNameConflict
	}

	handleConflict, err := ous.ouStore.CheckOrganizationUnitHandleConflict(request.Handle, request.Parent)
	if err != nil {
		logger.Error("Failed to check organization unit handle conflict", log.Error(err))
		return OrganizationUnit{}, &ErrorInternalServerError
	}
	if handleConflict {
		return OrganizationUnit{}, &ErrorOrganizationUnitHandleConflict
	}

	ouID, err := utils.GenerateUUIDv7()
	if err != nil {
		logger.Error("Failed to generate UUID", log.Error(err))
		return OrganizationUnit{}, &ErrorInternalServerError
	}

	ou := OrganizationUnit{
		ID:              ouID,
		Handle:          request.Handle,
		Name:            request.Name,
		Description:     request.Description,
		Parent:          request.Parent,
		ThemeID:         request.ThemeID,
		LayoutID:        request.LayoutID,
		LogoURL:         request.LogoURL,
		TosURI:          request.TosURI,
		PolicyURI:       request.PolicyURI,
		CookiePolicyURI: request.CookiePolicyURI,
	}

	err = ous.ouStore.CreateOrganizationUnit(ou)
	if err != nil {
		logger.Error("Failed to create organization unit", log.Error(err))
		return OrganizationUnit{}, &ErrorInternalServerError
	}

	logger.Debug("Successfully created organization unit", log.String("ouID", ouID))

	return ou, nil
}

// GetOrganizationUnit retrieves an organization unit by ID.
func (ous *organizationUnitService) GetOrganizationUnit(
	ctx context.Context, id string,
) (OrganizationUnit, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentNameService))
	logger.Debug("Getting organization unit", log.String("ouID", id))

	if svcErr := ous.checkOUAccess(ctx, security.ActionReadOU, id); svcErr != nil {
		return OrganizationUnit{}, svcErr
	}

	ou, err := ous.ouStore.GetOrganizationUnit(id)
	if err != nil {
		if errors.Is(err, ErrOrganizationUnitNotFound) {
			return OrganizationUnit{}, &ErrorOrganizationUnitNotFound
		}
		logger.Error("Failed to get organization unit", log.Error(err))
		return OrganizationUnit{}, &ErrorInternalServerError
	}

	return ou, nil
}

// GetOrganizationUnitByPath retrieves an organization unit by hierarchical handle path.
func (ous *organizationUnitService) GetOrganizationUnitByPath(
	ctx context.Context, handlePath string,
) (OrganizationUnit, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentNameService))
	logger.Debug("Getting organization unit by path", log.String("path", handlePath))

	handles, serviceError := validateAndProcessHandlePath(handlePath)
	if serviceError != nil {
		return OrganizationUnit{}, serviceError
	}

	ou, err := ous.ouStore.GetOrganizationUnitByPath(handles)
	if err != nil {
		if errors.Is(err, ErrOrganizationUnitNotFound) {
			return OrganizationUnit{}, &ErrorOrganizationUnitNotFound
		}
		logger.Error("Failed to get organization unit by path", log.Error(err))
		return OrganizationUnit{}, &ErrorInternalServerError
	}

	if svcErr := ous.checkOUAccess(ctx, security.ActionReadOU, ou.ID); svcErr != nil {
		return OrganizationUnit{}, svcErr
	}

	return ou, nil
}

// IsOrganizationUnitExists checks if an organization unit exists by ID.
func (ous *organizationUnitService) IsOrganizationUnitExists(
	ctx context.Context, id string,
) (bool, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentNameService))
	logger.Debug("Checking if organization unit exists", log.String("ouID", id))

	exists, err := ous.ouStore.IsOrganizationUnitExists(id)
	if err != nil {
		logger.Error("Failed to check organization unit existence", log.Error(err))
		return false, &ErrorInternalServerError
	}

	return exists, nil
}

func (ous *organizationUnitService) IsOrganizationUnitDeclarative(id string) bool {
	return ous.ouStore.IsOrganizationUnitDeclarative(id)
}

// IsParent checks whether the provided parentID is an ancestor of childID.
// Returns true if the parent and child are the same or if parentID is an ancestor of childID.
func (ous *organizationUnitService) IsParent(
	ctx context.Context, parentID, childID string,
) (bool, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentNameService))

	if strings.TrimSpace(parentID) == "" || strings.TrimSpace(childID) == "" {
		return false, &ErrorInvalidRequestFormat
	}

	currentParent := &childID
	for currentParent != nil {
		if *currentParent == parentID {
			return true, nil
		}

		parentOU, err := ous.ouStore.GetOrganizationUnit(*currentParent)
		if err != nil {
			if errors.Is(err, ErrOrganizationUnitNotFound) {
				logger.Debug("Encountered missing organization unit in hierarchy", log.String("ouID", *currentParent))
				return false, &ErrorOrganizationUnitNotFound
			}
			logger.Error("Failed to traverse organization unit hierarchy", log.Error(err))
			return false, &ErrorInternalServerError
		}

		currentParent = parentOU.Parent
	}

	return false, nil
}

// UpdateOrganizationUnit updates an organization unit.
func (ous *organizationUnitService) UpdateOrganizationUnit(
	ctx context.Context, id string, request OrganizationUnitRequest,
) (OrganizationUnit, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentNameService))
	logger.Debug("Updating organization unit", log.String("ouID", id))

	// Fail if store is in declarative mode
	if isDeclarativeModeEnabled() {
		return OrganizationUnit{}, &ErrorCannotModifyDeclarativeResource
	}

	if svcErr := ous.checkOUAccess(ctx, security.ActionUpdateOU, id); svcErr != nil {
		return OrganizationUnit{}, svcErr
	}

	existingOU, err := ous.ouStore.GetOrganizationUnit(id)
	if err != nil {
		if errors.Is(err, ErrOrganizationUnitNotFound) {
			return OrganizationUnit{}, &ErrorOrganizationUnitNotFound
		}
		logger.Error("Failed to get organization unit", log.Error(err))
		return OrganizationUnit{}, &ErrorInternalServerError
	}

	updatedOU, serviceError := ous.updateOUInternal(id, request, existingOU, logger)
	if serviceError != nil {
		return OrganizationUnit{}, serviceError
	}

	logger.Debug("Successfully updated organization unit", log.String("ouID", id))
	return updatedOU, nil
}

// UpdateOrganizationUnitByPath updates an organization unit by hierarchical handle path.
func (ous *organizationUnitService) UpdateOrganizationUnitByPath(
	ctx context.Context, handlePath string, request OrganizationUnitRequest,
) (OrganizationUnit, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentNameService))
	logger.Debug("Updating organization unit by path", log.String("path", handlePath))

	if err := declarativeresource.CheckDeclarativeUpdate(); err != nil {
		return OrganizationUnit{}, err
	}

	handles, serviceError := validateAndProcessHandlePath(handlePath)
	if serviceError != nil {
		return OrganizationUnit{}, serviceError
	}

	existingOU, err := ous.ouStore.GetOrganizationUnitByPath(handles)
	if err != nil {
		if errors.Is(err, ErrOrganizationUnitNotFound) {
			return OrganizationUnit{}, &ErrorOrganizationUnitNotFound
		}
		logger.Error("Failed to get organization unit by path", log.Error(err))
		return OrganizationUnit{}, &ErrorInternalServerError
	}

	if svcErr := ous.checkOUAccess(ctx, security.ActionUpdateOU, existingOU.ID); svcErr != nil {
		return OrganizationUnit{}, svcErr
	}

	// Check if OU is declarative (for composite mode)
	if ous.ouStore.IsOrganizationUnitDeclarative(existingOU.ID) {
		return OrganizationUnit{}, &ErrorCannotModifyDeclarativeResource
	}

	updatedOU, serviceError := ous.updateOUInternal(existingOU.ID, request, existingOU, logger)
	if serviceError != nil {
		return OrganizationUnit{}, serviceError
	}

	logger.Debug("Successfully updated organization unit by path", log.String("ouID", existingOU.ID))
	return updatedOU, nil
}

func (ous *organizationUnitService) updateOUInternal(
	id string,
	request OrganizationUnitRequest,
	existingOU OrganizationUnit,
	logger *log.Logger,
) (OrganizationUnit, *serviceerror.ServiceError) {
	// Check if OU is immutable (for composite mode)
	if ous.ouStore.IsOrganizationUnitDeclarative(id) {
		return OrganizationUnit{}, &ErrorCannotModifyDeclarativeResource
	}

	if err := ous.validateOUName(request.Name); err != nil {
		return OrganizationUnit{}, err
	}

	if err := ous.validateOUHandle(request.Handle); err != nil {
		return OrganizationUnit{}, err
	}

	if request.Parent != nil {
		exists, err := ous.ouStore.IsOrganizationUnitExists(*request.Parent)
		if err != nil {
			logger.Error("Failed to check parent organization unit existence", log.Error(err))
			return OrganizationUnit{}, &ErrorInternalServerError
		}
		if !exists {
			return OrganizationUnit{}, &ErrorParentOrganizationUnitNotFound
		}
	}

	if err := ous.checkCircularDependency(id, request.Parent); err != nil {
		return OrganizationUnit{}, err
	}

	parentChanged := !stringPtrEqual(existingOU.Parent, request.Parent)

	var nameConflict bool
	var err error
	if parentChanged || existingOU.Name != request.Name {
		nameConflict, err = ous.ouStore.CheckOrganizationUnitNameConflict(request.Name, request.Parent)
		if err != nil {
			logger.Error("Failed to check organization unit name conflict", log.Error(err))
			return OrganizationUnit{}, &ErrorInternalServerError
		}
	}

	if nameConflict {
		return OrganizationUnit{}, &ErrorOrganizationUnitNameConflict
	}

	var handleConflict bool
	if parentChanged || existingOU.Handle != request.Handle {
		handleConflict, err = ous.ouStore.CheckOrganizationUnitHandleConflict(request.Handle, request.Parent)
		if err != nil {
			logger.Error("Failed to check organization unit handle conflict", log.Error(err))
			return OrganizationUnit{}, &ErrorInternalServerError
		}
	}

	if handleConflict {
		return OrganizationUnit{}, &ErrorOrganizationUnitHandleConflict
	}

	updatedOU := OrganizationUnit{
		ID:              existingOU.ID,
		Handle:          request.Handle,
		Name:            request.Name,
		Description:     request.Description,
		Parent:          request.Parent,
		ThemeID:         request.ThemeID,
		LayoutID:        request.LayoutID,
		LogoURL:         request.LogoURL,
		TosURI:          request.TosURI,
		PolicyURI:       request.PolicyURI,
		CookiePolicyURI: request.CookiePolicyURI,
	}

	err = ous.ouStore.UpdateOrganizationUnit(updatedOU)
	if err != nil {
		if errors.Is(err, ErrOrganizationUnitNotFound) {
			return OrganizationUnit{}, &ErrorOrganizationUnitNotFound
		}
		logger.Error("Failed to update organization unit", log.Error(err))
		return OrganizationUnit{}, &ErrorInternalServerError
	}
	return updatedOU, nil
}

// DeleteOrganizationUnit deletes an organization unit.
func (ous *organizationUnitService) DeleteOrganizationUnit(ctx context.Context, id string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentNameService))
	logger.Debug("Deleting organization unit", log.String("ouID", id))

	// Fail if store is in declarative mode
	if isDeclarativeModeEnabled() {
		return &ErrorCannotModifyDeclarativeResource
	}

	if svcErr := ous.checkOUAccess(ctx, security.ActionDeleteOU, id); svcErr != nil {
		return svcErr
	}

	// Check if organization unit exists
	exists, err := ous.ouStore.IsOrganizationUnitExists(id)
	if err != nil {
		logger.Error("Failed to check organization unit existence", log.Error(err))
		return &ErrorInternalServerError
	}
	if !exists {
		return &ErrorOrganizationUnitNotFound
	}

	serviceError := ous.deleteOUInternal(id, logger)
	if serviceError != nil {
		return serviceError
	}

	logger.Debug("Successfully deleted organization unit", log.String("ouID", id))
	return nil
}

// DeleteOrganizationUnitByPath deletes an organization unit by hierarchical handle path.
func (ous *organizationUnitService) DeleteOrganizationUnitByPath(
	ctx context.Context, handlePath string,
) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentNameService))
	logger.Debug("Deleting organization unit by path", log.String("path", handlePath))

	if err := declarativeresource.CheckDeclarativeDelete(); err != nil {
		return err
	}

	handles, serviceError := validateAndProcessHandlePath(handlePath)
	if serviceError != nil {
		return serviceError
	}

	existingOU, err := ous.ouStore.GetOrganizationUnitByPath(handles)
	if err != nil {
		if errors.Is(err, ErrOrganizationUnitNotFound) {
			return &ErrorOrganizationUnitNotFound
		}
		logger.Error("Failed to get organization unit by path", log.Error(err))
		return &ErrorInternalServerError
	}

	if svcErr := ous.checkOUAccess(ctx, security.ActionDeleteOU, existingOU.ID); svcErr != nil {
		return svcErr
	}

	// Check if OU is declarative (for composite mode)
	if ous.ouStore.IsOrganizationUnitDeclarative(existingOU.ID) {
		return &ErrorCannotModifyDeclarativeResource
	}

	if svcErr := ous.deleteOUInternal(existingOU.ID, logger); svcErr != nil {
		return svcErr
	}

	logger.Debug("Successfully deleted organization unit by path", log.String("ouID", existingOU.ID))
	return nil
}

// deleteOUInternal deletes an organization unit by ID after checking if it has child resources.
func (ous *organizationUnitService) deleteOUInternal(id string, logger *log.Logger) *serviceerror.ServiceError {
	// Check if OU is immutable (for composite mode)
	if ous.ouStore.IsOrganizationUnitDeclarative(id) {
		return &ErrorCannotModifyDeclarativeResource
	}

	hasChildren, err := ous.ouStore.CheckOrganizationUnitHasChildResources(id)
	if err != nil {
		logger.Error("Failed to check if organization unit has children", log.Error(err))
		return &ErrorInternalServerError
	}
	if hasChildren {
		return &ErrorCannotDeleteOrganizationUnit
	}

	err = ous.ouStore.DeleteOrganizationUnit(id)
	if err != nil {
		if errors.Is(err, ErrOrganizationUnitNotFound) {
			return &ErrorOrganizationUnitNotFound
		}
		logger.Error("Failed to delete organization unit", log.Error(err))
		return &ErrorInternalServerError
	}
	return nil
}

// checkOUAccess validates that the caller is authorized to perform the given action on an organization unit.
// Pass an empty ouID when there is no specific resource context (e.g. creating a root-level OU).
func (ous *organizationUnitService) checkOUAccess(
	ctx context.Context, action security.Action, ouID string,
) *serviceerror.ServiceError {
	allowed, svcErr := ous.authzService.IsActionAllowed(ctx, action,
		&sysauthz.ActionContext{ResourceType: security.ResourceTypeOU, OuID: ouID})
	if svcErr != nil {
		return &ErrorInternalServerError
	}
	if !allowed {
		return &serviceerror.ErrorUnauthorized
	}
	return nil
}

// GetOrganizationUnitUsers retrieves a list of users for a given organization unit ID.
func (ous *organizationUnitService) GetOrganizationUnitUsers(
	ctx context.Context, id string, limit, offset int,
) (*UserListResponse, *serviceerror.ServiceError) {
	if svcErr := ous.checkOUAccess(ctx, security.ActionReadUser, id); svcErr != nil {
		return nil, svcErr
	}

	items, totalCount, svcErr := ous.getResourceListWithExistenceCheck(
		id, limit, offset, "users",
		func(id string, limit, offset int) (interface{}, error) {
			return ous.ouStore.GetOrganizationUnitUsersList(id, limit, offset)
		},
		ous.ouStore.GetOrganizationUnitUsersCount,
		false, // No composite error mapping for users
	)
	if svcErr != nil {
		return nil, svcErr
	}
	return buildUserListResponse(items, totalCount, limit, offset)
}

// GetOrganizationUnitGroups retrieves a list of groups for a given organization unit ID.
func (ous *organizationUnitService) GetOrganizationUnitGroups(
	ctx context.Context, id string, limit, offset int,
) (*GroupListResponse, *serviceerror.ServiceError) {
	if svcErr := ous.checkOUAccess(ctx, security.ActionReadGroup, id); svcErr != nil {
		return nil, svcErr
	}

	items, totalCount, svcErr := ous.getResourceListWithExistenceCheck(
		id, limit, offset, "groups",
		func(id string, limit, offset int) (interface{}, error) {
			return ous.ouStore.GetOrganizationUnitGroupsList(id, limit, offset)
		},
		ous.ouStore.GetOrganizationUnitGroupsCount,
		false, // No composite error mapping for groups
	)
	if svcErr != nil {
		return nil, svcErr
	}
	return buildGroupListResponse(items, totalCount, limit, offset)
}

// GetOrganizationUnitChildren retrieves a list of child organization units for a given organization unit ID.
func (ous *organizationUnitService) GetOrganizationUnitChildren(
	ctx context.Context, id string, limit, offset int,
) (*OrganizationUnitListResponse, *serviceerror.ServiceError) {
	if svcErr := ous.checkOUAccess(ctx, security.ActionListChildOUs, id); svcErr != nil {
		return nil, svcErr
	}

	items, totalCount, svcErr := ous.getResourceListWithExistenceCheck(
		id, limit, offset, "child organization units",
		func(id string, limit, offset int) (interface{}, error) {
			return ous.ouStore.GetOrganizationUnitChildrenList(id, limit, offset)
		},
		ous.ouStore.GetOrganizationUnitChildrenCount,
		true, // Map composite limit error for children
	)
	if svcErr != nil {
		return nil, svcErr
	}
	return buildOrganizationUnitListResponse(items, totalCount, limit, offset)
}

// GetOrganizationUnitChildrenByPath retrieves a list of child organization units by hierarchical handle path.
func (ous *organizationUnitService) GetOrganizationUnitChildrenByPath(
	ctx context.Context, handlePath string, limit, offset int,
) (*OrganizationUnitListResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentNameService))
	logger.Debug("Getting organization unit children by path", log.String("path", handlePath))

	handles, serviceError := validateAndProcessHandlePath(handlePath)
	if serviceError != nil {
		return nil, serviceError
	}

	ou, err := ous.ouStore.GetOrganizationUnitByPath(handles)
	if err != nil {
		if errors.Is(err, ErrOrganizationUnitNotFound) {
			return nil, &ErrorOrganizationUnitNotFound
		}
		logger.Error("Failed to get organization unit by path", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	return ous.GetOrganizationUnitChildren(ctx, ou.ID, limit, offset)
}

// GetOrganizationUnitUsersByPath retrieves a list of users by hierarchical handle path.
func (ous *organizationUnitService) GetOrganizationUnitUsersByPath(
	ctx context.Context, handlePath string, limit, offset int,
) (*UserListResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentNameService))
	logger.Debug("Getting organization unit users by path", log.String("path", handlePath))

	handles, serviceError := validateAndProcessHandlePath(handlePath)
	if serviceError != nil {
		return nil, serviceError
	}

	ou, err := ous.ouStore.GetOrganizationUnitByPath(handles)
	if err != nil {
		if errors.Is(err, ErrOrganizationUnitNotFound) {
			return nil, &ErrorOrganizationUnitNotFound
		}
		logger.Error("Failed to get organization unit by path", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	return ous.GetOrganizationUnitUsers(ctx, ou.ID, limit, offset)
}

// GetOrganizationUnitGroupsByPath retrieves a list of groups by hierarchical handle path.
func (ous *organizationUnitService) GetOrganizationUnitGroupsByPath(
	ctx context.Context, handlePath string, limit, offset int,
) (*GroupListResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentNameService))
	logger.Debug("Getting organization unit groups by path", log.String("path", handlePath))

	handles, serviceError := validateAndProcessHandlePath(handlePath)
	if serviceError != nil {
		return nil, serviceError
	}

	ou, err := ous.ouStore.GetOrganizationUnitByPath(handles)
	if err != nil {
		if errors.Is(err, ErrOrganizationUnitNotFound) {
			return nil, &ErrorOrganizationUnitNotFound
		}
		logger.Error("Failed to get organization unit by path", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	return ous.GetOrganizationUnitGroups(ctx, ou.ID, limit, offset)
}

// checkCircularDependency checks if setting the parent would create a circular dependency.
func (ous *organizationUnitService) checkCircularDependency(ouID string, parentID *string) *serviceerror.ServiceError {
	if parentID == nil {
		return nil
	}

	if ouID == *parentID {
		return &ErrorCircularDependency
	}

	currentParentID := parentID
	for currentParentID != nil {
		if *currentParentID == ouID {
			return &ErrorCircularDependency
		}

		parentOU, err := ous.ouStore.GetOrganizationUnit(*currentParentID)
		if err != nil {
			if errors.Is(err, ErrOrganizationUnitNotFound) {
				break
			}
			return &ErrorInternalServerError
		}

		currentParentID = parentOU.Parent
	}

	return nil
}

// validateOUName validates organization unit name.
func (ous *organizationUnitService) validateOUName(name string) *serviceerror.ServiceError {
	if strings.TrimSpace(name) == "" {
		return &ErrorInvalidRequestFormat
	}

	return nil
}

// validateOUHandle validates organization unit handle.
func (ous *organizationUnitService) validateOUHandle(handle string) *serviceerror.ServiceError {
	trimmed := strings.TrimSpace(handle)
	if trimmed == "" {
		return &ErrorInvalidRequestFormat
	}

	if strings.Contains(trimmed, "/") {
		return &ErrorInvalidRequestFormat
	}

	return nil
}

func validateAndProcessHandlePath(handlePath string) ([]string, *serviceerror.ServiceError) {
	if strings.TrimSpace(handlePath) == "" {
		return nil, &ErrorInvalidHandlePath
	}

	trimmed := strings.Trim(handlePath, "/")
	if trimmed == "" {
		return nil, &ErrorInvalidHandlePath
	}

	handles := strings.Split(trimmed, "/")
	var validHandles []string
	for _, handle := range handles {
		if strings.TrimSpace(handle) != "" {
			validHandles = append(validHandles, strings.TrimSpace(handle))
		}
	}
	return validHandles, nil
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

// buildPaginationLinks builds pagination links for the response.
func buildPaginationLinks(limit, offset, totalCount int) []Link {
	links := make([]Link, 0)

	if offset > 0 {
		links = append(links, Link{
			Href: fmt.Sprintf("/organization-units?offset=0&limit=%d", limit),
			Rel:  "first",
		})

		prevOffset := offset - limit
		if prevOffset < 0 {
			prevOffset = 0
		}
		links = append(links, Link{
			Href: fmt.Sprintf("/organization-units?offset=%d&limit=%d", prevOffset, limit),
			Rel:  "prev",
		})
	}

	if offset+limit < totalCount {
		nextOffset := offset + limit
		links = append(links, Link{
			Href: fmt.Sprintf("/organization-units?offset=%d&limit=%d", nextOffset, limit),
			Rel:  "next",
		})
	}

	lastPageOffset := ((totalCount - 1) / limit) * limit
	if offset < lastPageOffset {
		links = append(links, Link{
			Href: fmt.Sprintf("/organization-units?offset=%d&limit=%d", lastPageOffset, limit),
			Rel:  "last",
		})
	}

	return links
}

// getResourceListWithExistenceCheck is a generic function to get resources for an
// organization unit with existence check.
// If mapCompositeError is true, it will map ErrResultLimitExceededInCompositeMode to ErrorResultLimitExceeded.
func (ous *organizationUnitService) getResourceListWithExistenceCheck(
	id string, limit, offset int, resourceType string,
	getListFunc func(string, int, int) (interface{}, error),
	getCountFunc func(string) (int, error),
	mapCompositeError bool,
) (interface{}, int, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentNameService))
	logger.Debug("Getting resource for organization unit", log.String("resource_type", resourceType),
		log.String("ouID", id))

	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, 0, err
	}

	// Check if the organization unit exists
	exists, err := ous.ouStore.IsOrganizationUnitExists(id)
	if err != nil {
		logger.Error("Failed to check organization unit existence", log.Error(err))
		return nil, 0, &ErrorInternalServerError
	}
	if !exists {
		return nil, 0, &ErrorOrganizationUnitNotFound
	}

	items, err := getListFunc(id, limit, offset)
	if err != nil {
		// Map composite limit error if requested
		if mapCompositeError && errors.Is(err, ErrResultLimitExceededInCompositeMode) {
			return nil, 0, &ErrorResultLimitExceeded
		}
		logger.Error("Failed to list resource", log.String("resource_type", resourceType), log.Error(err))
		return nil, 0, &ErrorInternalServerError
	}

	totalCount, err := getCountFunc(id)
	if err != nil {
		logger.Error("Failed to get resource count", log.String("resource_type", resourceType), log.Error(err))
		return nil, 0, &ErrorInternalServerError
	}

	return items, totalCount, nil
}

func buildUserListResponse(items interface{}, totalCount, limit, offset int) (
	*UserListResponse, *serviceerror.ServiceError,
) {
	users, ok := items.([]User)
	if !ok {
		return nil, &ErrorInternalServerError
	}
	response := &UserListResponse{
		TotalResults: totalCount,
		Users:        users,
		StartIndex:   offset + 1,
		Count:        len(users),
		Links:        buildPaginationLinks(limit, offset, totalCount),
	}
	return response, nil
}

func buildGroupListResponse(items interface{}, totalCount, limit, offset int) (
	*GroupListResponse, *serviceerror.ServiceError,
) {
	groups, ok := items.([]Group)
	if !ok {
		return nil, &ErrorInternalServerError
	}
	response := &GroupListResponse{
		TotalResults: totalCount,
		Groups:       groups,
		StartIndex:   offset + 1,
		Count:        len(groups),
		Links:        buildPaginationLinks(limit, offset, totalCount),
	}
	return response, nil
}

func buildOrganizationUnitListResponse(items interface{}, totalCount, limit, offset int) (
	*OrganizationUnitListResponse, *serviceerror.ServiceError,
) {
	children, ok := items.([]OrganizationUnitBasic)
	if !ok {
		return nil, &ErrorInternalServerError
	}
	response := &OrganizationUnitListResponse{
		TotalResults:      totalCount,
		OrganizationUnits: children,
		StartIndex:        offset + 1,
		Count:             len(children),
		Links:             buildPaginationLinks(limit, offset, totalCount),
	}
	return response, nil
}

// stringPtrEqual compares two string pointers by their values.
func stringPtrEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
