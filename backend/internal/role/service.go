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

	"github.com/asgardeo/thunder/internal/group"
	oupkg "github.com/asgardeo/thunder/internal/ou"
	resourcepkg "github.com/asgardeo/thunder/internal/resource"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
	"github.com/asgardeo/thunder/internal/user"
)

const loggerComponentName = "RoleMgtService"

// RoleServiceInterface defines the interface for the role service.
type RoleServiceInterface interface {
	GetRoleList(limit, offset int) (*RoleList, *serviceerror.ServiceError)
	CreateRole(role RoleCreationDetail) (*RoleWithPermissionsAndAssignments, *serviceerror.ServiceError)
	GetRoleWithPermissions(id string) (*RoleWithPermissions, *serviceerror.ServiceError)
	UpdateRoleWithPermissions(id string, role RoleUpdateDetail) (*RoleWithPermissions, *serviceerror.ServiceError)
	DeleteRole(id string) *serviceerror.ServiceError
	GetRoleAssignments(id string, limit, offset int,
		includeDisplay bool) (*AssignmentList, *serviceerror.ServiceError)
	AddAssignments(id string, assignments []RoleAssignment) *serviceerror.ServiceError
	RemoveAssignments(id string, assignments []RoleAssignment) *serviceerror.ServiceError
	GetAuthorizedPermissions(
		userID string, groups []string, requestedPermissions []string,
	) ([]string, *serviceerror.ServiceError)
}

// roleService is the default implementation of the RoleServiceInterface.
type roleService struct {
	roleStore       roleStoreInterface
	userService     user.UserServiceInterface
	groupService    group.GroupServiceInterface
	ouService       oupkg.OrganizationUnitServiceInterface
	resourceService resourcepkg.ResourceServiceInterface
}

// newRoleService creates a new instance of RoleService with injected dependencies.
func newRoleService(
	roleStore roleStoreInterface,
	userService user.UserServiceInterface,
	groupService group.GroupServiceInterface,
	ouService oupkg.OrganizationUnitServiceInterface,
	resourceService resourcepkg.ResourceServiceInterface,
) RoleServiceInterface {
	return &roleService{
		roleStore:       roleStore,
		userService:     userService,
		groupService:    groupService,
		ouService:       ouService,
		resourceService: resourceService,
	}
}

// GetRoleList retrieves a list of roles.
func (rs *roleService) GetRoleList(limit, offset int) (*RoleList, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}

	totalCount, err := rs.roleStore.GetRoleListCount()
	if err != nil {
		logger.Error("Failed to get role count", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	roles, err := rs.roleStore.GetRoleList(limit, offset)
	if err != nil {
		logger.Error("Failed to list roles", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	response := &RoleList{
		TotalResults: totalCount,
		Roles:        roles,
		StartIndex:   offset + 1,
		Count:        len(roles),
		Links:        buildPaginationLinks("/roles", limit, offset, totalCount),
	}

	return response, nil
}

// CreateRole creates a new role.
func (rs *roleService) CreateRole(
	role RoleCreationDetail,
) (*RoleWithPermissionsAndAssignments, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Creating role", log.String("name", role.Name))

	if err := rs.validateCreateRoleRequest(role); err != nil {
		return nil, err
	}

	// Validate permissions exist in resource management system
	if err := rs.validatePermissions(role.Permissions); err != nil {
		return nil, err
	}

	// Validate assignment IDs early to avoid unnecessary database operations
	if len(role.Assignments) > 0 {
		if err := rs.validateAssignmentIDs(role.Assignments); err != nil {
			return nil, err
		}
	}

	// Validate organization unit exists using OU service
	_, svcErr := rs.ouService.GetOrganizationUnit(role.OrganizationUnitID)
	if svcErr != nil {
		if svcErr.Code == oupkg.ErrorOrganizationUnitNotFound.Code {
			logger.Debug("Organization unit not found", log.String("ouID", role.OrganizationUnitID))
			return nil, &ErrorOrganizationUnitNotFound
		}
		logger.Error("Failed to validate organization unit", log.String("error", svcErr.Error))
		return nil, &ErrorInternalServerError
	}

	// Check if role name already exists in the organization unit
	nameExists, err := rs.roleStore.CheckRoleNameExists(role.OrganizationUnitID, role.Name)
	if err != nil {
		logger.Error("Failed to check role name existence", log.Error(err))
		return nil, &ErrorInternalServerError
	}
	if nameExists {
		logger.Debug("Role name already exists in organization unit",
			log.String("name", role.Name), log.String("ouID", role.OrganizationUnitID))
		return nil, &ErrorRoleNameConflict
	}

	id := utils.GenerateUUID()
	if err := rs.roleStore.CreateRole(id, role); err != nil {
		logger.Error("Failed to create role", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	serviceRole := &RoleWithPermissionsAndAssignments{
		ID:                 id,
		Name:               role.Name,
		Description:        role.Description,
		OrganizationUnitID: role.OrganizationUnitID,
		Permissions:        role.Permissions,
		Assignments:        role.Assignments,
	}

	logger.Debug("Successfully created role", log.String("id", id), log.String("name", role.Name))
	return serviceRole, nil
}

// GetRoleWithPermissions retrieves a specific role by its id.
func (rs *roleService) GetRoleWithPermissions(id string) (*RoleWithPermissions, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Retrieving role", log.String("id", id))

	if id == "" {
		return nil, &ErrorMissingRoleID
	}

	role, err := rs.roleStore.GetRole(id)
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
	id string, role RoleUpdateDetail) (*RoleWithPermissions, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Updating role", log.String("id", id), log.String("name", role.Name))

	if id == "" {
		return nil, &ErrorMissingRoleID
	}

	if err := rs.validateUpdateRoleRequest(role); err != nil {
		return nil, err
	}

	// Validate permissions exist in resource management system
	if err := rs.validatePermissions(role.Permissions); err != nil {
		return nil, err
	}

	exists, err := rs.roleStore.IsRoleExist(id)
	if err != nil {
		logger.Error("Failed to check role existence", log.String("id", id), log.Error(err))
		return nil, &ErrorInternalServerError
	}
	if !exists {
		logger.Debug("Role not found", log.String("id", id))
		return nil, &ErrorRoleNotFound
	}

	// Validate organization unit exists using OU service
	_, svcErr := rs.ouService.GetOrganizationUnit(role.OrganizationUnitID)
	if svcErr != nil {
		if svcErr.Code == oupkg.ErrorOrganizationUnitNotFound.Code {
			logger.Debug("Organization unit not found", log.String("ouID", role.OrganizationUnitID))
			return nil, &ErrorOrganizationUnitNotFound
		}
		logger.Error("Failed to validate organization unit", log.String("error", svcErr.Error))
		return nil, &ErrorInternalServerError
	}

	// Check if role name already exists in the organization unit (excluding the current role)
	nameExists, err := rs.roleStore.CheckRoleNameExistsExcludingID(role.OrganizationUnitID, role.Name, id)
	if err != nil {
		logger.Error("Failed to check role name existence", log.Error(err))
		return nil, &ErrorInternalServerError
	}
	if nameExists {
		logger.Debug("Role name already exists in organization unit",
			log.String("name", role.Name), log.String("ouID", role.OrganizationUnitID))
		return nil, &ErrorRoleNameConflict
	}

	if err := rs.roleStore.UpdateRole(id, role); err != nil {
		logger.Error("Failed to update role", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	logger.Debug("Successfully updated role", log.String("id", id), log.String("name", role.Name))
	return &RoleWithPermissions{
		ID:                 id,
		Name:               role.Name,
		Description:        role.Description,
		OrganizationUnitID: role.OrganizationUnitID,
		Permissions:        role.Permissions,
	}, nil
}

// DeleteRole delete the specified role by its id.
func (rs *roleService) DeleteRole(id string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Deleting role", log.String("id", id))

	if id == "" {
		return &ErrorMissingRoleID
	}

	exists, err := rs.roleStore.IsRoleExist(id)
	if err != nil {
		logger.Error("Failed to check role existence", log.String("id", id), log.Error(err))
		return &ErrorInternalServerError
	}
	if !exists {
		logger.Debug("Role not found", log.String("id", id))
		return nil
	}

	// Check if role has any assignments before deleting
	assignmentCount, err := rs.roleStore.GetRoleAssignmentsCount(id)
	if err != nil {
		logger.Error("Failed to get role assignments count", log.String("id", id), log.Error(err))
		return &ErrorInternalServerError
	}

	if assignmentCount > 0 {
		logger.Debug("Cannot delete role with active assignments",
			log.String("id", id), log.Int("assignmentCount", assignmentCount))
		return &ErrorCannotDeleteRole
	}

	if err := rs.roleStore.DeleteRole(id); err != nil {
		logger.Error("Failed to delete role", log.String("id", id), log.Error(err))
		return &ErrorInternalServerError
	}

	logger.Debug("Successfully deleted role", log.String("id", id))
	return nil
}

// GetRoleAssignments retrieves assignments for a role with pagination.
func (rs *roleService) GetRoleAssignments(id string, limit, offset int,
	includeDisplay bool) (*AssignmentList, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}

	if id == "" {
		return nil, &ErrorMissingRoleID
	}

	exists, err := rs.roleStore.IsRoleExist(id)
	if err != nil {
		logger.Error("Failed to check role existence", log.String("id", id), log.Error(err))
		return nil, &ErrorInternalServerError
	}
	if !exists {
		logger.Debug("Role not found", log.String("id", id))
		return nil, &ErrorRoleNotFound
	}

	totalCount, err := rs.roleStore.GetRoleAssignmentsCount(id)
	if err != nil {
		logger.Error("Failed to get role assignments count", log.String("id", id), log.Error(err))
		return nil, &ErrorInternalServerError
	}

	assignments, err := rs.roleStore.GetRoleAssignments(id, limit, offset)
	if err != nil {
		logger.Error("Failed to get role assignments", log.String("id", id), log.Error(err))
		return nil, &ErrorInternalServerError
	}

	// Convert to service layer assignments
	serviceAssignments := make([]RoleAssignmentWithDisplay, len(assignments))

	for i := range assignments {
		// Populate display names if requested
		displayName := ""
		if includeDisplay {
			displayName, err = rs.getDisplayNameForAssignment(&assignments[i])
			if err != nil {
				logger.Warn("Failed to get display name for assignment",
					log.String("assignmentID", assignments[i].ID),
					log.String("assignmentType", string(assignments[i].Type)),
					log.Error(err))
				// Continue with empty display name rather than failing the entire request
				displayName = ""
			}
		}
		serviceAssignments[i].ID = assignments[i].ID
		serviceAssignments[i].Type = assignments[i].Type
		serviceAssignments[i].Display = displayName
	}
	baseURL := fmt.Sprintf("/roles/%s/assignments", id)
	links := buildPaginationLinks(baseURL, limit, offset, totalCount)

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
func (rs *roleService) AddAssignments(id string, assignments []RoleAssignment) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Adding assignments to role", log.String("id", id))

	if id == "" {
		return &ErrorMissingRoleID
	}

	if err := rs.validateAssignmentsRequest(assignments); err != nil {
		return err
	}

	exists, err := rs.roleStore.IsRoleExist(id)
	if err != nil {
		logger.Error("Failed to check role existence", log.String("id", id), log.Error(err))
		return &ErrorInternalServerError
	}
	if !exists {
		logger.Debug("Role not found", log.String("id", id))
		return &ErrorRoleNotFound
	}

	if err := rs.validateAssignmentIDs(assignments); err != nil {
		return err
	}

	if err := rs.roleStore.AddAssignments(id, assignments); err != nil {
		logger.Error("Failed to add assignments to role", log.String("id", id), log.Error(err))
		return &ErrorInternalServerError
	}

	logger.Debug("Successfully added assignments to role", log.String("id", id))
	return nil
}

// RemoveAssignments removes assignments from a role.
func (rs *roleService) RemoveAssignments(id string, assignments []RoleAssignment) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Removing assignments from role", log.String("id", id))

	if id == "" {
		return &ErrorMissingRoleID
	}

	if err := rs.validateAssignmentsRequest(assignments); err != nil {
		return err
	}

	exists, err := rs.roleStore.IsRoleExist(id)
	if err != nil {
		logger.Error("Failed to check role existence", log.String("id", id), log.Error(err))
		return &ErrorInternalServerError
	}
	if !exists {
		logger.Debug("Role not found", log.String("id", id))
		return &ErrorRoleNotFound
	}

	if err := rs.roleStore.RemoveAssignments(id, assignments); err != nil {
		logger.Error("Failed to remove assignments from role", log.String("id", id), log.Error(err))
		return &ErrorInternalServerError
	}

	logger.Debug("Successfully removed assignments from role", log.String("id", id))
	return nil
}

// GetAuthorizedPermissions checks which of the requested permissions are authorized for the user based on their roles.
func (rs *roleService) GetAuthorizedPermissions(
	userID string, groups []string, requestedPermissions []string,
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
	authorizedPermissions, err := rs.roleStore.GetAuthorizedPermissions(userID, groups, requestedPermissions)
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

// validateCreateRoleRequest validates the create role request.
func (rs *roleService) validateCreateRoleRequest(role RoleCreationDetail) *serviceerror.ServiceError {
	if role.Name == "" {
		return &ErrorInvalidRequestFormat
	}

	if role.OrganizationUnitID == "" {
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

	if request.OrganizationUnitID == "" {
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
		if assignment.Type != AssigneeTypeUser && assignment.Type != AssigneeTypeGroup {
			return &ErrorInvalidRequestFormat
		}
		if assignment.ID == "" {
			return &ErrorInvalidRequestFormat
		}
	}

	return nil
}

// validateAssignmentIDs validates that all provided assignment IDs exist.
func (rs *roleService) validateAssignmentIDs(assignments []RoleAssignment) *serviceerror.ServiceError {
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

	// Validate user IDs using user service
	if len(userIDs) > 0 {
		invalidUserIDs, svcErr := rs.userService.ValidateUserIDs(context.TODO(), userIDs)
		if svcErr != nil {
			logger.Error("Failed to validate user IDs", log.String("error", svcErr.Error),
				log.String("code", svcErr.Code))
			return &ErrorInternalServerError
		}

		if len(invalidUserIDs) > 0 {
			logger.Debug("Invalid user IDs found", log.Any("invalidUserIDs", invalidUserIDs))
			return &ErrorInvalidAssignmentID
		}
	}

	// Validate group IDs using group service
	if len(groupIDs) > 0 {
		if err := rs.groupService.ValidateGroupIDs(groupIDs); err != nil {
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

// buildPaginationLinks builds pagination links for the response.
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

// getDisplayNameForAssignment retrieves the display name for a user or group assignment.
func (rs *roleService) getDisplayNameForAssignment(assignment *RoleAssignment) (string, error) {
	switch assignment.Type {
	case AssigneeTypeUser:
		userResp, svcErr := rs.userService.GetUser(context.TODO(), assignment.ID)
		if svcErr != nil {
			return "", fmt.Errorf("failed to get user: %w", errors.New(svcErr.Error))
		}
		// Return user ID as display name (since User doesn't have a username field)
		return userResp.ID, nil

	case AssigneeTypeGroup:
		groupResp, svcErr := rs.groupService.GetGroup(assignment.ID)
		if svcErr != nil {
			return "", fmt.Errorf("failed to get group: %w", errors.New(svcErr.Error))
		}
		// Return group name as display name
		return groupResp.Name, nil

	default:
		return "", fmt.Errorf("unknown assignment type: %s", assignment.Type)
	}
}

// validatePermissions validates that all permissions exist in the resource management system.
func (rs *roleService) validatePermissions(permissions []ResourcePermissions) *serviceerror.ServiceError {
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
