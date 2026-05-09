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

// Package role provides role management functionality.
package role

import (
	"context"
	"errors"
	"fmt"

	"encoding/json"

	"github.com/asgardeo/thunder/internal/entity"
	"github.com/asgardeo/thunder/internal/entitytype"
	"github.com/asgardeo/thunder/internal/group"
	oupkg "github.com/asgardeo/thunder/internal/ou"
	resourcepkg "github.com/asgardeo/thunder/internal/resource"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/transaction"
	"github.com/asgardeo/thunder/internal/system/utils"
)

const loggerComponentName = "RoleMgtService"

// RoleServiceInterface defines the interface for the role service.
type RoleServiceInterface interface {
	GetRoleList(ctx context.Context, limit, offset int) (*RoleList, *serviceerror.ServiceError)
	CreateRole(ctx context.Context, role RoleCreationDetail) (
		*RoleWithPermissionsAndAssignments, *serviceerror.ServiceError)
	GetRoleWithPermissions(ctx context.Context, id string) (*RoleWithPermissions, *serviceerror.ServiceError)
	UpdateRoleWithPermissions(ctx context.Context, id string, role RoleUpdateDetail) (
		*RoleWithPermissions, *serviceerror.ServiceError)
	DeleteRole(ctx context.Context, id string) *serviceerror.ServiceError
	GetRoleAssignments(ctx context.Context, id string, limit, offset int,
		includeDisplay bool) (*AssignmentList, *serviceerror.ServiceError)
	GetRoleAssignmentsByType(ctx context.Context, id string, limit, offset int,
		includeDisplay bool, assigneeType string) (*AssignmentList, *serviceerror.ServiceError)
	AddAssignments(ctx context.Context, id string, assignments []RoleAssignment) *serviceerror.ServiceError
	RemoveAssignments(ctx context.Context, id string, assignments []RoleAssignment) *serviceerror.ServiceError
	IsRoleDeclarative(ctx context.Context, id string) (bool, *serviceerror.ServiceError)
	GetAuthorizedPermissions(
		ctx context.Context, entityID string, groups []string, requestedPermissions []string,
	) ([]string, *serviceerror.ServiceError)
	GetUserRoles(ctx context.Context, entityID string, groupIDs []string) ([]string, *serviceerror.ServiceError)
}

// roleService is the default implementation of the RoleServiceInterface.
type roleService struct {
	roleStore         roleStoreInterface
	entityService     entity.EntityServiceInterface
	groupService      group.GroupServiceInterface
	ouService         oupkg.OrganizationUnitServiceInterface
	resourceService   resourcepkg.ResourceServiceInterface
	entityTypeService entitytype.EntityTypeServiceInterface
	transactioner     transaction.Transactioner
}

// newRoleService creates a new instance of RoleService with injected dependencies.
func newRoleService(
	roleStore roleStoreInterface,
	entityService entity.EntityServiceInterface,
	groupService group.GroupServiceInterface,
	ouService oupkg.OrganizationUnitServiceInterface,
	resourceService resourcepkg.ResourceServiceInterface,
	entityTypeService entitytype.EntityTypeServiceInterface,
	transactioner transaction.Transactioner,
) RoleServiceInterface {
	return &roleService{
		roleStore:         roleStore,
		entityService:     entityService,
		groupService:      groupService,
		ouService:         ouService,
		resourceService:   resourceService,
		entityTypeService: entityTypeService,
		transactioner:     transactioner,
	}
}

// GetRoleList retrieves a list of roles.
func (rs *roleService) GetRoleList(ctx context.Context, limit, offset int) (*RoleList, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}

	totalCount, err := rs.roleStore.GetRoleListCount(ctx)
	if err != nil {
		if errors.Is(err, errResultLimitExceededInCompositeMode) {
			return nil, &ResultLimitExceededInCompositeMode
		}
		logger.Error("Failed to get role count", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	roles, err := rs.roleStore.GetRoleList(ctx, limit, offset)
	if err != nil {
		if errors.Is(err, errResultLimitExceededInCompositeMode) {
			return nil, &ResultLimitExceededInCompositeMode
		}
		logger.Error("Failed to list roles", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	if len(roles) > 0 {
		seen := make(map[string]struct{}, len(roles))
		ouIDs := make([]string, 0, len(roles))
		for _, r := range roles {
			if r.OUID != "" {
				if _, exists := seen[r.OUID]; !exists {
					ouIDs = append(ouIDs, r.OUID)
					seen[r.OUID] = struct{}{}
				}
			}
		}
		ouHandles, svcErr := rs.ouService.GetOrganizationUnitHandlesByIDs(ctx, ouIDs)
		if svcErr != nil {
			logger.Warn("Failed to resolve OU handles for roles, skipping", log.Any("error", svcErr))
		} else {
			for i := range roles {
				roles[i].OUHandle = ouHandles[roles[i].OUID]
			}
		}
	}

	response := &RoleList{
		TotalResults: totalCount,
		Roles:        roles,
		StartIndex:   offset + 1,
		Count:        len(roles),
		Links:        utils.BuildPaginationLinks("/roles", limit, offset, totalCount, ""),
	}

	return response, nil
}

// CreateRole creates a new role.
func (rs *roleService) CreateRole(
	ctx context.Context, role RoleCreationDetail,
) (*RoleWithPermissionsAndAssignments, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Creating role", log.String("name", role.Name))

	// Check if role creation is allowed (not in declarative-only mode)
	if isDeclarativeModeEnabled() {
		logger.Debug("Cannot create role in declarative-only mode")
		return nil, &ErrorDeclarativeModeCreateNotAllowed
	}

	if err := rs.validateCreateRoleRequest(role); err != nil {
		return nil, err
	}

	responseAssignments := role.Assignments

	// Validate organization unit exists using OU service
	ou, svcErr := rs.ouService.GetOrganizationUnit(ctx, role.OUID)
	if svcErr != nil {
		if svcErr.Code == oupkg.ErrorOrganizationUnitNotFound.Code {
			logger.Debug("Organization unit not found", log.String("ouID", role.OUID))
			return nil, &ErrorOrganizationUnitNotFound
		}
		logger.Error("Failed to validate organization unit", log.String("error", svcErr.Error.DefaultValue))
		return nil, &serviceerror.InternalServerError
	}

	// Validate permissions exist in resource management system
	if err := rs.validatePermissions(ctx, role.Permissions); err != nil {
		return nil, err
	}

	// Validate assignment IDs (existence + category check) before normalization.
	if len(role.Assignments) > 0 {
		if err := rs.validateAssignmentIDs(ctx, role.Assignments); err != nil {
			return nil, err
		}
	}

	role.Assignments = normalizeAssignments(role.Assignments)

	// Check if role name already exists in the organization unit
	nameExists, err := rs.roleStore.CheckRoleNameExists(ctx, role.OUID, role.Name)
	if err != nil {
		logger.Error("Failed to check role name existence", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if nameExists {
		logger.Debug("Role name already exists in organization unit",
			log.String("name", role.Name), log.String("ouID", role.OUID))
		return nil, &ErrorRoleNameConflict
	}

	id := role.ID
	if id == "" {
		id, err = utils.GenerateUUIDv7()
		if err != nil {
			logger.Error("Failed to generate UUID", log.Error(err))
			return nil, &serviceerror.InternalServerError
		}
	} else {
		_, err = rs.roleStore.GetRole(ctx, id)
		if err != nil && !errors.Is(err, ErrRoleNotFound) {
			logger.Error("Failed to check role ID existence", log.Error(err))
			return nil, &serviceerror.InternalServerError
		}
		if err == nil {
			logger.Debug("Role ID already exists", log.String("id", id))
			return nil, &ErrorRoleIDConflict
		}
	}

	serviceRole := &RoleWithPermissionsAndAssignments{
		ID:          id,
		Name:        role.Name,
		Description: role.Description,
		OUID:        role.OUID,
		OUHandle:    ou.Handle,
		Permissions: role.Permissions,
		Assignments: responseAssignments,
	}

	err = rs.transactioner.Transact(ctx, func(txCtx context.Context) error {
		if err := rs.roleStore.CreateRole(txCtx, id, role); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		logger.Error("Failed to create role", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	logger.Debug("Successfully created role", log.String("id", id), log.String("name", role.Name))
	return serviceRole, nil
}

// GetRoleWithPermissions retrieves a specific role by its id.
func (rs *roleService) GetRoleWithPermissions(ctx context.Context, id string) (
	*RoleWithPermissions, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Retrieving role", log.String("id", id))

	if id == "" {
		return nil, &ErrorMissingRoleID
	}

	role, err := rs.roleStore.GetRole(ctx, id)
	if err != nil {
		if errors.Is(err, ErrRoleNotFound) {
			logger.Debug("Role not found", log.String("id", id))
			return nil, &ErrorRoleNotFound
		}
		logger.Error("Failed to retrieve role", log.String("id", id), log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	ou, svcErr := rs.ouService.GetOrganizationUnit(ctx, role.OUID)
	if svcErr != nil {
		logger.Warn("Failed to resolve OU handle for role, skipping",
			log.String("id", id), log.Any("error", svcErr))
	} else {
		role.OUHandle = ou.Handle
	}

	logger.Debug("Successfully retrieved role", log.String("id", role.ID), log.String("name", role.Name))
	return &role, nil
}

// UpdateRole updates an existing role.
func (rs *roleService) UpdateRoleWithPermissions(
	ctx context.Context, id string, role RoleUpdateDetail) (*RoleWithPermissions, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Updating role", log.String("id", id), log.String("name", role.Name))

	if id == "" {
		return nil, &ErrorMissingRoleID
	}

	if err := rs.validateUpdateRoleRequest(role); err != nil {
		return nil, err
	}

	// Validate permissions exist in resource management system
	if err := rs.validatePermissions(ctx, role.Permissions); err != nil {
		return nil, err
	}

	exists, err := rs.roleStore.IsRoleExist(ctx, id)
	if err != nil {
		logger.Error("Failed to check role existence", log.String("id", id), log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if !exists {
		logger.Debug("Role not found", log.String("id", id))
		return nil, &ErrorRoleNotFound
	}

	// Check if role is declarative - cannot modify declarative roles
	if rs.isRoleDeclarative(ctx, id) {
		logger.Debug("Cannot modify declarative role", log.String("id", id))
		return nil, &ErrorImmutableRole
	}

	// Validate organization unit exists using OU service
	ou, svcErr := rs.ouService.GetOrganizationUnit(ctx, role.OUID)
	if svcErr != nil {
		if svcErr.Code == oupkg.ErrorOrganizationUnitNotFound.Code {
			logger.Debug("Organization unit not found", log.String("ouID", role.OUID))
			return nil, &ErrorOrganizationUnitNotFound
		}
		logger.Error("Failed to validate organization unit", log.String("error", svcErr.Error.DefaultValue))
		return nil, &serviceerror.InternalServerError
	}

	// Check if role name already exists in the organization unit (excluding the current role)
	nameExists, err := rs.roleStore.CheckRoleNameExistsExcludingID(ctx, role.OUID, role.Name, id)
	if err != nil {
		logger.Error("Failed to check role name existence", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if nameExists {
		logger.Debug("Role name already exists in organization unit",
			log.String("name", role.Name), log.String("ouID", role.OUID))
		return nil, &ErrorRoleNameConflict
	}

	err = rs.transactioner.Transact(ctx, func(txCtx context.Context) error {
		return rs.roleStore.UpdateRole(txCtx, id, role)
	})

	if err != nil {
		logger.Error("Failed to update role", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	logger.Debug("Successfully updated role", log.String("id", id), log.String("name", role.Name))
	return &RoleWithPermissions{
		ID:          id,
		Name:        role.Name,
		Description: role.Description,
		OUID:        role.OUID,
		OUHandle:    ou.Handle,
		Permissions: role.Permissions,
	}, nil
}

// DeleteRole delete the specified role by its id.
func (rs *roleService) DeleteRole(ctx context.Context, id string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Deleting role", log.String("id", id))

	if id == "" {
		return &ErrorMissingRoleID
	}

	exists, err := rs.roleStore.IsRoleExist(ctx, id)
	if err != nil {
		logger.Error("Failed to check role existence", log.String("id", id), log.Error(err))
		return &serviceerror.InternalServerError
	}
	if !exists {
		logger.Debug("Role not found", log.String("id", id))
		return nil
	}

	// Check if role is declarative - cannot delete declarative roles
	if rs.isRoleDeclarative(ctx, id) {
		logger.Debug("Cannot delete declarative role", log.String("id", id))
		return &ErrorImmutableRole
	}

	// Check if role has any assignments before deleting
	assignmentCount, err := rs.roleStore.GetRoleAssignmentsCount(ctx, id)
	if err != nil {
		if errors.Is(err, errResultLimitExceededInCompositeMode) {
			return &ResultLimitExceededInCompositeMode
		}
		logger.Error("Failed to get role assignments count", log.String("id", id), log.Error(err))
		return &serviceerror.InternalServerError
	}

	if assignmentCount > 0 {
		logger.Debug("Cannot delete role with active assignments",
			log.String("id", id), log.Int("assignmentCount", assignmentCount))
		return &ErrorCannotDeleteRole
	}

	if err := rs.roleStore.DeleteRole(ctx, id); err != nil {
		logger.Error("Failed to delete role", log.String("id", id), log.Error(err))
		return &serviceerror.InternalServerError
	}

	logger.Debug("Successfully deleted role", log.String("id", id))
	return nil
}

// GetRoleAssignments retrieves assignments for a role with pagination.
func (rs *roleService) GetRoleAssignments(ctx context.Context, id string, limit, offset int,
	includeDisplay bool) (*AssignmentList, *serviceerror.ServiceError) {
	return rs.GetRoleAssignmentsByType(ctx, id, limit, offset, includeDisplay, "")
}

// GetRoleAssignmentsByType retrieves assignments for a role filtered by assignee type with pagination.
func (rs *roleService) GetRoleAssignmentsByType(ctx context.Context, id string, limit, offset int,
	includeDisplay bool, assigneeType string) (*AssignmentList, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}

	if id == "" {
		return nil, &ErrorMissingRoleID
	}

	exists, err := rs.roleStore.IsRoleExist(ctx, id)
	if err != nil {
		logger.Error("Failed to check role existence", log.String("id", id), log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if !exists {
		logger.Debug("Role not found", log.String("id", id))
		return nil, &ErrorRoleNotFound
	}

	// user/app/agent filters require fetching all entity assignments and post-filtering by category.
	if assigneeType == string(entity.EntityCategoryUser) ||
		assigneeType == string(entity.EntityCategoryApp) ||
		assigneeType == string(entity.EntityCategoryAgent) {
		return rs.getAssignmentsByEntityCategory(ctx, id, limit, offset, includeDisplay, assigneeType, logger)
	}

	// For no filter or 'group' filter, use DB-level pagination directly.
	var totalCount int
	var assignments []RoleAssignment
	if assigneeType != "" {
		totalCount, err = rs.roleStore.GetRoleAssignmentsCountByType(ctx, id, assigneeType)
	} else {
		totalCount, err = rs.roleStore.GetRoleAssignmentsCount(ctx, id)
	}
	if err != nil {
		if errors.Is(err, errResultLimitExceededInCompositeMode) {
			return nil, &ResultLimitExceededInCompositeMode
		}
		logger.Error("Failed to get role assignments count", log.String("id", id), log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	if assigneeType != "" {
		assignments, err = rs.roleStore.GetRoleAssignmentsByType(ctx, id, limit, offset, assigneeType)
	} else {
		assignments, err = rs.roleStore.GetRoleAssignments(ctx, id, limit, offset)
	}
	if err != nil {
		if errors.Is(err, errResultLimitExceededInCompositeMode) {
			return nil, &ResultLimitExceededInCompositeMode
		}
		logger.Error("Failed to get role assignments", log.String("id", id), log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	serviceAssignments, svcErr := rs.resolveAssignments(ctx, assignments, includeDisplay)
	if svcErr != nil {
		return nil, svcErr
	}

	baseURL := fmt.Sprintf("/roles/%s/assignments", id)
	extraQuery := utils.DisplayQueryParam(includeDisplay)
	if assigneeType != "" {
		extraQuery += "&type=" + assigneeType
	}
	links := utils.BuildPaginationLinks(baseURL, limit, offset, totalCount, extraQuery)

	return &AssignmentList{
		TotalResults: totalCount,
		Assignments:  serviceAssignments,
		StartIndex:   offset + 1,
		Count:        len(serviceAssignments),
		Links:        links,
	}, nil
}

// getAssignmentsByEntityCategory handles ?type=user and ?type=app filter cases.
// Since both are stored as 'entity' internally, it fetches all entity assignments,
// resolves their category, and paginates the filtered results in memory.
func (rs *roleService) getAssignmentsByEntityCategory(
	ctx context.Context, id string, limit, offset int,
	includeDisplay bool, category string, logger *log.Logger,
) (*AssignmentList, *serviceerror.ServiceError) {
	totalEntityCount, err := rs.roleStore.GetRoleAssignmentsCountByType(ctx, id, string(assigneeTypeEntity))
	if err != nil {
		if errors.Is(err, errResultLimitExceededInCompositeMode) {
			return nil, &ResultLimitExceededInCompositeMode
		}
		logger.Error("Failed to get entity assignments count", log.String("id", id), log.Error(err))
		return nil, &ErrorInternalServerError
	}

	var allEntityAssignments []RoleAssignment
	if totalEntityCount > 0 {
		allEntityAssignments, err = rs.roleStore.GetRoleAssignmentsByType(
			ctx, id, totalEntityCount, 0, string(assigneeTypeEntity))
		if err != nil {
			if errors.Is(err, errResultLimitExceededInCompositeMode) {
				return nil, &ResultLimitExceededInCompositeMode
			}
			logger.Error("Failed to get entity assignments", log.String("id", id), log.Error(err))
			return nil, &ErrorInternalServerError
		}
	}

	// Batch-resolve entity categories.
	entityCategoryMap := make(map[string]string)
	if len(allEntityAssignments) > 0 {
		entityIDs := make([]string, len(allEntityAssignments))
		for i, a := range allEntityAssignments {
			entityIDs[i] = a.ID
		}
		entities, fetchErr := rs.entityService.GetEntitiesByIDs(ctx, entityIDs)
		if fetchErr != nil {
			logger.Error("Failed to batch fetch entities for category filter", log.Error(fetchErr))
			return nil, &ErrorInternalServerError
		}
		for _, e := range entities {
			entityCategoryMap[e.ID] = string(e.Category)
		}
	}

	// Filter to matching category and paginate in memory.
	var filtered []RoleAssignment
	for _, a := range allEntityAssignments {
		if entityCategoryMap[a.ID] == category {
			filtered = append(filtered, a)
		}
	}

	totalCount := len(filtered)
	start := offset
	if start > totalCount {
		start = totalCount
	}
	end := start + limit
	if end > totalCount {
		end = totalCount
	}
	page := filtered[start:end]

	serviceAssignments, svcErr := rs.resolveAssignments(ctx, page, includeDisplay)
	if svcErr != nil {
		return nil, svcErr
	}

	baseURL := fmt.Sprintf("/roles/%s/assignments", id)
	extraQuery := utils.DisplayQueryParam(includeDisplay) + "&type=" + category
	links := utils.BuildPaginationLinks(baseURL, limit, offset, totalCount, extraQuery)

	return &AssignmentList{
		TotalResults: totalCount,
		Assignments:  serviceAssignments,
		StartIndex:   offset + 1,
		Count:        len(serviceAssignments),
		Links:        links,
	}, nil
}

// AddAssignments adds assignments to a role.
func (rs *roleService) AddAssignments(
	ctx context.Context, id string, assignments []RoleAssignment) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Adding assignments to role", log.String("id", id))

	normalized, svcErr := rs.prepareAssignments(ctx, id, assignments)
	if svcErr != nil {
		return svcErr
	}

	if err := rs.transactioner.Transact(ctx, func(txCtx context.Context) error {
		return rs.roleStore.AddAssignments(txCtx, id, normalized)
	}); err != nil {
		logger.Error("Failed to add assignments to role", log.String("id", id), log.Error(err))
		return &serviceerror.InternalServerError
	}

	logger.Debug("Successfully added assignments to role", log.String("id", id))
	return nil
}

// RemoveAssignments removes assignments from a role.
func (rs *roleService) RemoveAssignments(
	ctx context.Context, id string, assignments []RoleAssignment) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Removing assignments from role", log.String("id", id))

	normalized, svcErr := rs.prepareAssignments(ctx, id, assignments)
	if svcErr != nil {
		return svcErr
	}

	if err := rs.transactioner.Transact(ctx, func(txCtx context.Context) error {
		return rs.roleStore.RemoveAssignments(txCtx, id, normalized)
	}); err != nil {
		logger.Error("Failed to remove assignments from role", log.String("id", id), log.Error(err))
		return &serviceerror.InternalServerError
	}

	logger.Debug("Successfully removed assignments from role", log.String("id", id))
	return nil
}

// prepareAssignments validates, normalizes, and checks role accessibility before an assignment mutation.
// It returns the normalized assignments ready for storage, or a service error.
func (rs *roleService) prepareAssignments(
	ctx context.Context, id string, assignments []RoleAssignment,
) ([]RoleAssignment, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if id == "" {
		return nil, &ErrorMissingRoleID
	}

	if err := rs.validateAssignmentsRequest(assignments); err != nil {
		return nil, err
	}

	exists, err := rs.roleStore.IsRoleExist(ctx, id)
	if err != nil {
		logger.Error("Failed to check role existence", log.String("id", id), log.Error(err))
		return nil, &ErrorInternalServerError
	}
	if !exists {
		logger.Debug("Role not found", log.String("id", id))
		return nil, &ErrorRoleNotFound
	}

	if rs.isRoleDeclarative(ctx, id) {
		logger.Debug("Cannot modify assignments for declarative role", log.String("id", id))
		return nil, &ErrorImmutableAssignment
	}

	if err := rs.validateAssignmentIDs(ctx, assignments); err != nil {
		return nil, err
	}

	normalized := normalizeAssignments(assignments)

	return normalized, nil
}

// GetAuthorizedPermissions checks which requested permissions are authorized for the entity based on roles.
func (rs *roleService) GetAuthorizedPermissions(
	ctx context.Context, entityID string, groups []string, requestedPermissions []string,
) ([]string, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Authorizing permissions",
		log.MaskedString(log.LoggerKeyUserID, entityID), log.Int("groupCount", len(groups)))

	// Handle nil groups slice
	if groups == nil {
		groups = []string{}
	}

	// Validate that at least entityID or groups is provided
	if entityID == "" && len(groups) == 0 {
		return nil, &ErrorMissingEntityOrGroups
	}

	// Return empty list if no permissions requested
	if len(requestedPermissions) == 0 {
		return []string{}, nil
	}

	// Get authorized permissions from store
	authorizedPermissions, err := rs.roleStore.GetAuthorizedPermissions(ctx, entityID, groups, requestedPermissions)
	if err != nil {
		logger.Error("Failed to get authorized permissions",
			log.MaskedString(log.LoggerKeyUserID, entityID),
			log.Int("groupCount", len(groups)),
			log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	logger.Debug("Retrieved authorized permissions",
		log.MaskedString(log.LoggerKeyUserID, entityID),
		log.Int("groupCount", len(groups)),
		log.Int("requestedCount", len(requestedPermissions)),
		log.Int("authorizedCount", len(authorizedPermissions)))

	return authorizedPermissions, nil
}

// GetUserRoles retrieves the names of roles assigned to an entity directly and/or through group membership.
func (rs *roleService) GetUserRoles(
	ctx context.Context, entityID string, groupIDs []string,
) ([]string, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Getting entity roles", log.MaskedString("entityID", entityID), log.Int("groupCount", len(groupIDs)))

	if groupIDs == nil {
		groupIDs = []string{}
	}

	if entityID == "" && len(groupIDs) == 0 {
		return []string{}, nil
	}

	roles, err := rs.roleStore.GetUserRoles(ctx, entityID, groupIDs)
	if err != nil {
		logger.Error("Failed to get entity roles",
			log.MaskedString("entityID", entityID), log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	return roles, nil
}

// IsRoleDeclarative returns true if the role is declarative.
func (rs *roleService) IsRoleDeclarative(ctx context.Context, id string) (bool, *serviceerror.ServiceError) {
	isDeclarative, err := rs.roleStore.IsRoleDeclarative(ctx, id)
	if err != nil {
		return false, &serviceerror.InternalServerError
	}

	return isDeclarative, nil
}

// validateCreateRoleRequest validates the create role request.
func (rs *roleService) validateCreateRoleRequest(role RoleCreationDetail) *serviceerror.ServiceError {
	if role.Name == "" {
		return &ErrorInvalidRequestFormat
	}

	if role.OUID == "" {
		return &ErrorInvalidRequestFormat
	}

	if len(role.Assignments) > 0 {
		if err := rs.validateAssignmentsRequest(role.Assignments); err != nil {
			return err
		}
	}

	return nil
}

// validateUpdateRoleRequest validates the update role request.
func (rs *roleService) validateUpdateRoleRequest(request RoleUpdateDetail) *serviceerror.ServiceError {
	if request.Name == "" {
		return &ErrorInvalidRequestFormat
	}

	if request.OUID == "" {
		return &ErrorInvalidRequestFormat
	}

	return nil
}

// validateAssignmentsRequest validates the assignments request.
// Accepts public types 'user', 'app', 'group'.
func (rs *roleService) validateAssignmentsRequest(assignments []RoleAssignment) *serviceerror.ServiceError {
	if len(assignments) == 0 {
		return &ErrorEmptyAssignments
	}

	for _, assignment := range assignments {
		if !assignment.Type.IsEntityType() && assignment.Type != AssigneeTypeGroup {
			return &ErrorInvalidAssigneeType
		}
		if assignment.ID == "" {
			return &ErrorInvalidRequestFormat
		}
	}

	return nil
}

// normalizeAssignments converts public 'user'/'app' types to the internal 'entity' type.
func normalizeAssignments(assignments []RoleAssignment) []RoleAssignment {
	normalized := make([]RoleAssignment, len(assignments))
	for i, a := range assignments {
		t := a.Type
		if t.IsEntityType() {
			t = assigneeTypeEntity
		}
		normalized[i] = RoleAssignment{ID: a.ID, Type: t}
	}
	return normalized
}

// validateAssignmentIDs validates assignment IDs before normalization.
// For user/app assignments it checks existence and verifies the claimed type matches the actual
// entity category. For group assignments it checks existence via the group service.
func (rs *roleService) validateAssignmentIDs(
	ctx context.Context, assignments []RoleAssignment) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	typeByID := make(map[string]AssigneeType)
	var groupIDs []string

	for _, a := range assignments {
		if a.Type.IsEntityType() {
			if existing, ok := typeByID[a.ID]; ok && existing != a.Type {
				return &ErrorInvalidAssignmentID
			}
			typeByID[a.ID] = a.Type
		} else if a.Type == AssigneeTypeGroup {
			groupIDs = append(groupIDs, a.ID)
		}
	}

	groupIDs = utils.UniqueStrings(groupIDs)

	if len(typeByID) > 0 {
		entityIDs := make([]string, 0, len(typeByID))
		for id := range typeByID {
			entityIDs = append(entityIDs, id)
		}

		entities, err := rs.entityService.GetEntitiesByIDs(ctx, entityIDs)
		if err != nil {
			logger.Error("Failed to fetch entities for assignment validation", log.Error(err))
			return &ErrorInternalServerError
		}

		if len(entities) != len(entityIDs) {
			return &ErrorInvalidAssignmentID
		}

		for _, e := range entities {
			claimed := typeByID[e.ID]
			actual := AssigneeType(e.Category)
			if claimed != actual {
				logger.Debug("Assignment type mismatch", log.String("id", e.ID),
					log.String("claimed", string(claimed)), log.String("actual", string(actual)))
				return &ErrorInvalidAssignmentID
			}
		}
	}

	if len(groupIDs) > 0 {
		if err := rs.groupService.ValidateGroupIDs(ctx, groupIDs); err != nil {
			if err.Code == group.ErrorInvalidGroupMemberID.Code {
				logger.Debug("Invalid group member IDs found")
				return &ErrorInvalidAssignmentID
			}
			logger.Error("Failed to validate group IDs", log.String("error", err.Error.DefaultValue))
			return &serviceerror.InternalServerError
		}
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

// resolveAssignments resolves the public types and optionally display names for role assignments.
// It batch-fetches entity categories to translate the internal 'entity' type into the public
// 'user' or 'app' value.
func (rs *roleService) resolveAssignments(
	ctx context.Context,
	assignments []RoleAssignment,
	includeDisplay bool,
) ([]RoleAssignmentWithDisplay, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	var entityIDs, groupIDs []string
	for _, a := range assignments {
		switch a.Type {
		case assigneeTypeEntity:
			entityIDs = append(entityIDs, a.ID)
		case AssigneeTypeGroup:
			groupIDs = append(groupIDs, a.ID)
		}
	}

	// Always batch-fetch entities to resolve their category (user vs app) for the API response type.
	var entityMap map[string]*entity.Entity
	if len(entityIDs) > 0 {
		entities, err := rs.entityService.GetEntitiesByIDs(ctx, entityIDs)
		if err != nil {
			logger.Error("Failed to batch fetch entities for assignments", log.Error(err))
			return nil, &ErrorInternalServerError
		}
		entityMap = make(map[string]*entity.Entity, len(entities))
		for i := range entities {
			entityMap[entities[i].ID] = &entities[i]
		}
	}

	var groupsMap map[string]*group.Group
	if includeDisplay && len(groupIDs) > 0 {
		var svcErr *serviceerror.ServiceError
		groupsMap, svcErr = rs.groupService.GetGroupsByIDs(ctx, groupIDs)
		if svcErr != nil {
			logger.Warn("Failed to batch fetch groups for display names", log.Any("error", svcErr))
		}
	}

	// Resolve display attribute paths for user-category entities.
	var displayAttrPaths map[string]string
	if includeDisplay && entityMap != nil {
		var userTypes []string
		for _, e := range entityMap {
			if e.Category == entity.EntityCategoryUser {
				userTypes = append(userTypes, e.Type)
			}
		}
		displayAttrPaths = resolveDisplayAttributePaths(ctx, userTypes, rs.entityTypeService, logger)
	}

	// Build the result slice, skipping orphaned entity assignments.
	result := make([]RoleAssignmentWithDisplay, 0, len(assignments))
	for _, a := range assignments {
		ra := RoleAssignmentWithDisplay{ID: a.ID}
		switch a.Type {
		case assigneeTypeEntity:
			e, ok := entityMap[a.ID]
			if !ok {
				logger.Warn("Skipping orphaned entity assignment", log.String("id", a.ID))
				continue
			}
			// Set the public type from the entity category ("user", "app", or "agent").
			ra.Type = AssigneeType(e.Category)
			if includeDisplay {
				if e.Category == entity.EntityCategoryUser {
					ra.Display = utils.ResolveDisplay(e.ID, e.Type, e.Attributes, displayAttrPaths)
				} else {
					ra.Display = resolveAppDisplay(*e)
				}
			}
		case AssigneeTypeGroup:
			ra.Type = AssigneeTypeGroup
			if includeDisplay {
				if groupsMap != nil {
					if g, ok := groupsMap[a.ID]; ok {
						ra.Display = g.Name
					} else {
						ra.Display = a.ID
					}
				} else {
					ra.Display = a.ID
				}
			}
		default:
			ra.Type = a.Type
			ra.Display = a.ID
		}
		result = append(result, ra)
	}
	return result, nil
}

// resolveDisplayAttributePaths collects unique user types and resolves their display
// attribute paths from the entity type service.
func resolveDisplayAttributePaths(
	ctx context.Context, userTypes []string, schemaService entitytype.EntityTypeServiceInterface,
	logger *log.Logger,
) map[string]string {
	if schemaService == nil || len(userTypes) == 0 {
		return nil
	}

	uniqueTypes := utils.UniqueNonEmptyStrings(userTypes)
	if len(uniqueTypes) == 0 {
		return nil
	}

	displayPaths, svcErr := schemaService.GetDisplayAttributesByNames(ctx, entitytype.TypeCategoryUser, uniqueTypes)
	if svcErr != nil {
		if logger != nil {
			logger.Warn("Failed to resolve display attribute paths, skipping display resolution",
				log.Any("error", svcErr))
		}
		return nil
	}

	return displayPaths
}

// resolveAppDisplay extracts a display name for an app entity from its system attributes.
func resolveAppDisplay(e entity.Entity) string {
	if len(e.SystemAttributes) > 0 {
		var sysAttrs map[string]interface{}
		if err := json.Unmarshal(e.SystemAttributes, &sysAttrs); err == nil {
			if name, ok := sysAttrs["name"].(string); ok && name != "" {
				return name
			}
		}
	}
	return e.ID
}

// validatePermissions validates that all permissions exist in the resource management system.
func (rs *roleService) validatePermissions(
	ctx context.Context, permissions []ResourcePermissions,
) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if len(permissions) == 0 {
		return nil
	}

	// Validate each resource server's permissions
	for _, resPerm := range permissions {
		if resPerm.ResourceServerID == "" {
			logger.Debug("Empty resource server ID")
			return &ErrorInvalidPermissions
		}

		if len(resPerm.Permissions) == 0 {
			continue
		}

		// Call resource service to validate permissions
		invalidPerms, svcErr := rs.resourceService.ValidatePermissions(
			ctx,
			resPerm.ResourceServerID,
			resPerm.Permissions,
		)

		if svcErr != nil {
			logger.Error("Failed to validate permissions",
				log.String("resourceServerId", resPerm.ResourceServerID),
				log.String("error", svcErr.Error.DefaultValue))
			return &serviceerror.InternalServerError
		}

		// If any permissions are invalid, return error
		if len(invalidPerms) > 0 {
			logger.Debug("Invalid permissions found",
				log.String("resourceServerId", resPerm.ResourceServerID),
				log.Any("invalidPermissions", invalidPerms),
				log.Int("count", len(invalidPerms)))
			return &ErrorInvalidPermissions
		}
	}

	return nil
}

// isRoleDeclarative checks if a role is defined in declarative configuration.
func (rs *roleService) isRoleDeclarative(ctx context.Context, roleID string) bool {
	// Check the store mode - if it's mutable, no roles are declarative
	storeMode := getRoleStoreMode()
	if storeMode == serverconst.StoreModeMutable {
		return false
	}

	// For declarative and composite modes, check with store
	// Note: This is a placeholder implementation
	// Actual implementation would check against declarative config
	isDeclarative, err := rs.roleStore.IsRoleDeclarative(ctx, roleID)
	if err != nil {
		// Log at Warn level and fail open - treat as non-declarative on error
		// RISK: In composite mode, this could allow modification of declarative roles if the check fails
		logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
		logger.Warn("Failed to check if role is declarative", log.String("roleID", roleID), log.Error(err))
		return false
	}

	return isDeclarative
}
