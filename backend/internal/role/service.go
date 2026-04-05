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
	"github.com/asgardeo/thunder/internal/group"
	oupkg "github.com/asgardeo/thunder/internal/ou"
	resourcepkg "github.com/asgardeo/thunder/internal/resource"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/transaction"
	"github.com/asgardeo/thunder/internal/system/utils"
	"github.com/asgardeo/thunder/internal/userschema"
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
		ctx context.Context, userID string, groups []string, requestedPermissions []string,
	) ([]string, *serviceerror.ServiceError)
	GetUserRoles(ctx context.Context, userID string, groupIDs []string) ([]string, *serviceerror.ServiceError)
}

// roleService is the default implementation of the RoleServiceInterface.
type roleService struct {
	roleStore         roleStoreInterface
	entityService     entity.EntityServiceInterface
	groupService      group.GroupServiceInterface
	ouService         oupkg.OrganizationUnitServiceInterface
	resourceService   resourcepkg.ResourceServiceInterface
	userSchemaService userschema.UserSchemaServiceInterface
	transactioner     transaction.Transactioner
}

// newRoleService creates a new instance of RoleService with injected dependencies.
func newRoleService(
	roleStore roleStoreInterface,
	entityService entity.EntityServiceInterface,
	groupService group.GroupServiceInterface,
	ouService oupkg.OrganizationUnitServiceInterface,
	resourceService resourcepkg.ResourceServiceInterface,
	userSchemaService userschema.UserSchemaServiceInterface,
	transactioner transaction.Transactioner,
) RoleServiceInterface {
	return &roleService{
		roleStore:         roleStore,
		entityService:     entityService,
		groupService:      groupService,
		ouService:         ouService,
		resourceService:   resourceService,
		userSchemaService: userSchemaService,
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
		return nil, &ErrorInternalServerError
	}

	roles, err := rs.roleStore.GetRoleList(ctx, limit, offset)
	if err != nil {
		if errors.Is(err, errResultLimitExceededInCompositeMode) {
			return nil, &ResultLimitExceededInCompositeMode
		}
		logger.Error("Failed to list roles", log.Error(err))
		return nil, &ErrorInternalServerError
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

	// Validate permissions exist in resource management system
	if err := rs.validatePermissions(ctx, role.Permissions); err != nil {
		return nil, err
	}

	// Validate assignment IDs early to avoid unnecessary database operations
	if len(role.Assignments) > 0 {
		if err := rs.validateAssignmentIDs(ctx, role.Assignments); err != nil {
			return nil, err
		}
	}

	// Validate organization unit exists using OU service
	_, svcErr := rs.ouService.GetOrganizationUnit(ctx, role.OUID)
	if svcErr != nil {
		if svcErr.Code == oupkg.ErrorOrganizationUnitNotFound.Code {
			logger.Debug("Organization unit not found", log.String("ouID", role.OUID))
			return nil, &ErrorOrganizationUnitNotFound
		}
		logger.Error("Failed to validate organization unit", log.String("error", svcErr.Error))
		return nil, &ErrorInternalServerError
	}

	// Check if role name already exists in the organization unit
	nameExists, err := rs.roleStore.CheckRoleNameExists(ctx, role.OUID, role.Name)
	if err != nil {
		logger.Error("Failed to check role name existence", log.Error(err))
		return nil, &ErrorInternalServerError
	}
	if nameExists {
		logger.Debug("Role name already exists in organization unit",
			log.String("name", role.Name), log.String("ouID", role.OUID))
		return nil, &ErrorRoleNameConflict
	}

	id, err := utils.GenerateUUIDv7()
	if err != nil {
		logger.Error("Failed to generate UUID", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	serviceRole := &RoleWithPermissionsAndAssignments{
		ID:          id,
		Name:        role.Name,
		Description: role.Description,
		OUID:        role.OUID,
		Permissions: role.Permissions,
		Assignments: role.Assignments,
	}

	err = rs.transactioner.Transact(ctx, func(txCtx context.Context) error {
		if err := rs.roleStore.CreateRole(txCtx, id, role); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		logger.Error("Failed to create role", log.Error(err))
		return nil, &ErrorInternalServerError
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
		return nil, &ErrorInternalServerError
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
		return nil, &ErrorInternalServerError
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
	_, svcErr := rs.ouService.GetOrganizationUnit(ctx, role.OUID)
	if svcErr != nil {
		if svcErr.Code == oupkg.ErrorOrganizationUnitNotFound.Code {
			logger.Debug("Organization unit not found", log.String("ouID", role.OUID))
			return nil, &ErrorOrganizationUnitNotFound
		}
		logger.Error("Failed to validate organization unit", log.String("error", svcErr.Error))
		return nil, &ErrorInternalServerError
	}

	// Check if role name already exists in the organization unit (excluding the current role)
	nameExists, err := rs.roleStore.CheckRoleNameExistsExcludingID(ctx, role.OUID, role.Name, id)
	if err != nil {
		logger.Error("Failed to check role name existence", log.Error(err))
		return nil, &ErrorInternalServerError
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
		return nil, &ErrorInternalServerError
	}

	logger.Debug("Successfully updated role", log.String("id", id), log.String("name", role.Name))
	return &RoleWithPermissions{
		ID:          id,
		Name:        role.Name,
		Description: role.Description,
		OUID:        role.OUID,
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
		return &ErrorInternalServerError
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
		return &ErrorInternalServerError
	}

	if assignmentCount > 0 {
		logger.Debug("Cannot delete role with active assignments",
			log.String("id", id), log.Int("assignmentCount", assignmentCount))
		return &ErrorCannotDeleteRole
	}

	if err := rs.roleStore.DeleteRole(ctx, id); err != nil {
		logger.Error("Failed to delete role", log.String("id", id), log.Error(err))
		return &ErrorInternalServerError
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
		return nil, &ErrorInternalServerError
	}
	if !exists {
		logger.Debug("Role not found", log.String("id", id))
		return nil, &ErrorRoleNotFound
	}

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
		return nil, &ErrorInternalServerError
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
		return nil, &ErrorInternalServerError
	}

	// Convert to service layer assignments
	serviceAssignments := make([]RoleAssignmentWithDisplay, len(assignments))

	if includeDisplay {
		rs.populateDisplayNames(ctx, assignments, serviceAssignments)
	} else {
		for i := range assignments {
			serviceAssignments[i].ID = assignments[i].ID
			serviceAssignments[i].Type = assignments[i].Type
		}
	}
	baseURL := fmt.Sprintf("/roles/%s/assignments", id)
	extraQuery := utils.DisplayQueryParam(includeDisplay)
	if assigneeType != "" {
		extraQuery += "&type=" + assigneeType
	}
	links := utils.BuildPaginationLinks(baseURL, limit, offset, totalCount, extraQuery)

	response := &AssignmentList{
		TotalResults: totalCount,
		Assignments:  serviceAssignments,
		StartIndex:   offset + 1,
		Count:        len(serviceAssignments),
		Links:        links,
	}

	return response, nil
}

// AddAssignments adds assignments to a role.
func (rs *roleService) AddAssignments(
	ctx context.Context, id string, assignments []RoleAssignment) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Adding assignments to role", log.String("id", id))

	if id == "" {
		return &ErrorMissingRoleID
	}

	if err := rs.validateAssignmentsRequest(assignments); err != nil {
		return err
	}

	exists, err := rs.roleStore.IsRoleExist(ctx, id)
	if err != nil {
		logger.Error("Failed to check role existence", log.String("id", id), log.Error(err))
		return &ErrorInternalServerError
	}
	if !exists {
		logger.Debug("Role not found", log.String("id", id))
		return &ErrorRoleNotFound
	}

	// Check if role is declarative - cannot modify assignments for declarative roles
	if rs.isRoleDeclarative(ctx, id) {
		logger.Debug("Cannot modify assignments for declarative role", log.String("id", id))
		return &ErrorImmutableAssignment
	}

	// Validate assignment IDs
	if err := rs.validateAssignmentIDs(ctx, assignments); err != nil {
		return err
	}

	err = rs.transactioner.Transact(ctx, func(txCtx context.Context) error {
		return rs.roleStore.AddAssignments(txCtx, id, assignments)
	})

	if err != nil {
		logger.Error("Failed to add assignments to role", log.String("id", id), log.Error(err))
		return &ErrorInternalServerError
	}

	logger.Debug("Successfully added assignments to role", log.String("id", id))
	return nil
}

// RemoveAssignments removes assignments from a role.
func (rs *roleService) RemoveAssignments(
	ctx context.Context, id string, assignments []RoleAssignment) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Removing assignments from role", log.String("id", id))

	if id == "" {
		return &ErrorMissingRoleID
	}

	if err := rs.validateAssignmentsRequest(assignments); err != nil {
		return err
	}

	exists, err := rs.roleStore.IsRoleExist(ctx, id)
	if err != nil {
		logger.Error("Failed to check role existence", log.String("id", id), log.Error(err))
		return &ErrorInternalServerError
	}
	if !exists {
		logger.Debug("Role not found", log.String("id", id))
		return &ErrorRoleNotFound
	}

	// Check if role is declarative - cannot modify assignments for declarative roles
	if rs.isRoleDeclarative(ctx, id) {
		logger.Debug("Cannot modify assignments for declarative role", log.String("id", id))
		return &ErrorImmutableAssignment
	}

	err = rs.transactioner.Transact(ctx, func(txCtx context.Context) error {
		return rs.roleStore.RemoveAssignments(txCtx, id, assignments)
	})

	if err != nil {
		logger.Error("Failed to remove assignments from role", log.String("id", id), log.Error(err))
		return &ErrorInternalServerError
	}

	logger.Debug("Successfully removed assignments from role", log.String("id", id))
	return nil
}

// GetAuthorizedPermissions checks which of the requested permissions are authorized for the user based on their roles.
func (rs *roleService) GetAuthorizedPermissions(
	ctx context.Context, userID string, groups []string, requestedPermissions []string,
) ([]string, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Authorizing permissions", log.String("userID", userID), log.Int("groupCount", len(groups)))

	// Handle nil groups slice
	if groups == nil {
		groups = []string{}
	}

	// Validate that at least userID or groups is provided
	if userID == "" && len(groups) == 0 {
		return nil, &ErrorMissingUserOrGroups
	}

	// Return empty list if no permissions requested
	if len(requestedPermissions) == 0 {
		return []string{}, nil
	}

	// Get authorized permissions from store
	authorizedPermissions, err := rs.roleStore.GetAuthorizedPermissions(ctx, userID, groups, requestedPermissions)
	if err != nil {
		logger.Error("Failed to get authorized permissions",
			log.String("userID", userID),
			log.Int("groupCount", len(groups)),
			log.Error(err))
		return nil, &ErrorInternalServerError
	}

	logger.Debug("Retrieved authorized permissions",
		log.String("userID", userID),
		log.Int("groupCount", len(groups)),
		log.Int("requestedCount", len(requestedPermissions)),
		log.Int("authorizedCount", len(authorizedPermissions)))

	return authorizedPermissions, nil
}

// GetUserRoles retrieves the names of roles assigned to a user directly and/or through group membership.
func (rs *roleService) GetUserRoles(
	ctx context.Context, userID string, groupIDs []string,
) ([]string, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Getting user roles", log.String("userID", userID), log.Int("groupCount", len(groupIDs)))

	if groupIDs == nil {
		groupIDs = []string{}
	}

	if userID == "" && len(groupIDs) == 0 {
		return []string{}, nil
	}

	roles, err := rs.roleStore.GetUserRoles(ctx, userID, groupIDs)
	if err != nil {
		logger.Error("Failed to get user roles",
			log.String("userID", userID), log.Error(err))
		return nil, &ErrorInternalServerError
	}

	return roles, nil
}

// IsRoleDeclarative returns true if the role is declarative.
func (rs *roleService) IsRoleDeclarative(ctx context.Context, id string) (bool, *serviceerror.ServiceError) {
	isDeclarative, err := rs.roleStore.IsRoleDeclarative(ctx, id)
	if err != nil {
		return false, &ErrorInternalServerError
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
func (rs *roleService) validateAssignmentsRequest(assignments []RoleAssignment) *serviceerror.ServiceError {
	if len(assignments) == 0 {
		return &ErrorEmptyAssignments
	}

	for _, assignment := range assignments {
		if assignment.Type != AssigneeTypeUser && assignment.Type != AssigneeTypeGroup &&
			assignment.Type != AssigneeTypeApp {
			return &ErrorInvalidRequestFormat
		}
		if assignment.ID == "" {
			return &ErrorInvalidRequestFormat
		}
	}

	return nil
}

// validateAssignmentIDs validates that all provided assignment IDs exist.
func (rs *roleService) validateAssignmentIDs(
	ctx context.Context, assignments []RoleAssignment) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	var userIDs []string
	var groupIDs []string

	// Collect user and group IDs
	for _, assignment := range assignments {
		switch assignment.Type {
		case AssigneeTypeUser:
			userIDs = append(userIDs, assignment.ID)
		case AssigneeTypeGroup:
			groupIDs = append(groupIDs, assignment.ID)
		}
	}

	// Deduplicate IDs
	userIDs = utils.UniqueStrings(userIDs)
	groupIDs = utils.UniqueStrings(groupIDs)

	// Validate user IDs using entity service
	if len(userIDs) > 0 {
		invalidUserIDs, err := rs.entityService.ValidateEntityIDs(ctx, userIDs)
		if err != nil {
			logger.Error("Failed to validate user IDs", log.Error(err))
			return &ErrorInternalServerError
		}

		if len(invalidUserIDs) > 0 {
			logger.Debug("Invalid user IDs found", log.Any("invalidUserIDs", invalidUserIDs))
			return &ErrorInvalidAssignmentID
		}
	}

	// Validate group IDs using group service
	if len(groupIDs) > 0 {
		if err := rs.groupService.ValidateGroupIDs(ctx, groupIDs); err != nil {
			if err.Code == group.ErrorInvalidGroupMemberID.Code {
				logger.Debug("Invalid group member IDs found")
				return &ErrorInvalidAssignmentID
			}
			logger.Error("Failed to validate group IDs", log.String("error", err.Error))
			return &ErrorInternalServerError
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

// populateDisplayNames batch-fetches display names for all assignments using GetUsersByIDs/GetGroupsByIDs.
func (rs *roleService) populateDisplayNames(
	ctx context.Context, assignments []RoleAssignment,
	serviceAssignments []RoleAssignmentWithDisplay,
) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	// Collect IDs by type
	var userIDs, groupIDs, appIDs []string
	for _, a := range assignments {
		switch a.Type {
		case AssigneeTypeUser:
			userIDs = append(userIDs, a.ID)
		case AssigneeTypeGroup:
			groupIDs = append(groupIDs, a.ID)
		case AssigneeTypeApp:
			appIDs = append(appIDs, a.ID)
		}
	}

	// Batch fetch users and groups
	var entityMap map[string]*entity.Entity
	var groupsMap map[string]*group.Group

	if len(userIDs) > 0 {
		entities, err := rs.entityService.GetEntitiesByIDs(ctx, userIDs)
		if err != nil {
			logger.Warn("Failed to batch fetch users for display names", log.Error(err))
		} else {
			entityMap = make(map[string]*entity.Entity, len(entities))
			for i := range entities {
				entityMap[entities[i].ID] = &entities[i]
			}
		}
	}

	if len(groupIDs) > 0 {
		var svcErr *serviceerror.ServiceError
		groupsMap, svcErr = rs.groupService.GetGroupsByIDs(ctx, groupIDs)
		if svcErr != nil {
			logger.Warn("Failed to batch fetch groups for display names", log.Any("error", svcErr))
		}
	}

	// Batch fetch app entities for app assignee display names.
	appNamesMap := make(map[string]string)
	if len(appIDs) > 0 && rs.entityService != nil {
		entities, err := rs.entityService.GetEntitiesByIDs(ctx, appIDs)
		if err != nil {
			logger.Warn("Failed to batch fetch app entities for display names", log.Error(err))
		} else {
			for _, e := range entities {
				appNamesMap[e.ID] = resolveEntityDisplayName(e)
			}
		}
	}

	// Resolve display attribute paths for user types
	userTypes := make([]string, 0, len(entityMap))
	for _, e := range entityMap {
		userTypes = append(userTypes, e.Type)
	}
	displayAttrPaths := resolveDisplayAttributePaths(ctx, userTypes, rs.userSchemaService, logger)

	for i := range assignments {
		serviceAssignments[i].ID = assignments[i].ID
		serviceAssignments[i].Type = assignments[i].Type

		switch assignments[i].Type {
		case AssigneeTypeUser:
			if entityMap != nil {
				if e, ok := entityMap[assignments[i].ID]; ok {
					serviceAssignments[i].Display = utils.ResolveDisplay(
						e.ID, e.Type, e.Attributes, displayAttrPaths)
					continue
				}
			}
			serviceAssignments[i].Display = assignments[i].ID
		case AssigneeTypeGroup:
			if groupsMap != nil {
				if g, ok := groupsMap[assignments[i].ID]; ok {
					serviceAssignments[i].Display = g.Name
					continue
				}
			}
			serviceAssignments[i].Display = assignments[i].ID
		case AssigneeTypeApp:
			if name, ok := appNamesMap[assignments[i].ID]; ok && name != "" {
				serviceAssignments[i].Display = name
				continue
			}
			serviceAssignments[i].Display = assignments[i].ID
		default:
			serviceAssignments[i].Display = assignments[i].ID
		}
	}
}

// resolveDisplayAttributePaths collects unique user types and resolves their display
// attribute paths from the user schema service.
func resolveDisplayAttributePaths(
	ctx context.Context, userTypes []string, schemaService userschema.UserSchemaServiceInterface,
	logger *log.Logger,
) map[string]string {
	if schemaService == nil || len(userTypes) == 0 {
		return nil
	}

	uniqueTypes := utils.UniqueNonEmptyStrings(userTypes)
	if len(uniqueTypes) == 0 {
		return nil
	}

	displayPaths, svcErr := schemaService.GetDisplayAttributesByNames(ctx, uniqueTypes)
	if svcErr != nil {
		if logger != nil {
			logger.Warn("Failed to resolve display attribute paths, skipping display resolution",
				log.Any("error", svcErr))
		}
		return nil
	}

	return displayPaths
}

// resolveEntityDisplayName extracts a display name from an entity's system attributes.
func resolveEntityDisplayName(e entity.Entity) string {
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
				log.String("error", svcErr.Error))
			return &ErrorInternalServerError
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
