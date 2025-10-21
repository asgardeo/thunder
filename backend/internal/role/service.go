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
	"errors"
	"fmt"

	"github.com/asgardeo/thunder/internal/group"
	oupkg "github.com/asgardeo/thunder/internal/ou"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
	"github.com/asgardeo/thunder/internal/user"
)

const loggerComponentName = "RoleMgtService"

// RoleServiceInterface defines the interface for the role service.
type RoleServiceInterface interface {
	GetRoleList(limit, offset int) (*RoleListResponse, *serviceerror.ServiceError)
	CreateRole(request CreateRoleRequest) (*Role, *serviceerror.ServiceError)
	GetRole(roleID string) (*Role, *serviceerror.ServiceError)
	UpdateRole(roleID string, request UpdateRoleRequest) (*Role, *serviceerror.ServiceError)
	DeleteRole(roleID string) *serviceerror.ServiceError
	GetRoleAssignments(roleID string, limit, offset int) (*AssignmentListResponse, *serviceerror.ServiceError)
	AddAssignments(roleID string, request AssignmentsRequest) *serviceerror.ServiceError
	RemoveAssignments(roleID string, request AssignmentsRequest) *serviceerror.ServiceError
}

// roleService is the default implementation of the RoleServiceInterface.
type roleService struct {
	roleStore    roleStoreInterface
	userService  user.UserServiceInterface
	groupService group.GroupServiceInterface
	ouService    oupkg.OrganizationUnitServiceInterface
}

// newRoleService creates a new instance of RoleService with injected dependencies.
func newRoleService(
	roleStore roleStoreInterface,
	userService user.UserServiceInterface,
	groupService group.GroupServiceInterface,
	ouService oupkg.OrganizationUnitServiceInterface,
) RoleServiceInterface {
	return &roleService{
		roleStore:    roleStore,
		userService:  userService,
		groupService: groupService,
		ouService:    ouService,
	}
}

// GetRoleList retrieves a list of roles.
func (rs *roleService) GetRoleList(limit, offset int) (*RoleListResponse, *serviceerror.ServiceError) {
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

	response := &RoleListResponse{
		TotalResults: totalCount,
		Roles:        roles,
		StartIndex:   offset + 1,
		Count:        len(roles),
		Links:        buildPaginationLinks("/roles", limit, offset, totalCount),
	}

	return response, nil
}

// CreateRole creates a new role.
func (rs *roleService) CreateRole(request CreateRoleRequest) (*Role, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Creating role", log.String("name", request.Name))

	if err := rs.validateCreateRoleRequest(request); err != nil {
		return nil, err
	}

	// Validate organization unit exists using OU service
	_, svcErr := rs.ouService.GetOrganizationUnit(request.OrganizationUnitID)
	if svcErr != nil {
		if svcErr.Code == oupkg.ErrorOrganizationUnitNotFound.Code {
			logger.Debug("Organization unit not found", log.String("ouID", request.OrganizationUnitID))
			return nil, &ErrorOrganizationUnitNotFound
		}
		logger.Error("Failed to validate organization unit", log.String("error", svcErr.Error))
		return nil, &ErrorInternalServerError
	}

	if err := rs.validateAssignmentIDs(request.Assignments); err != nil {
		return nil, err
	}

	role := Role{
		ID:                 utils.GenerateUUID(),
		Name:               request.Name,
		Description:        request.Description,
		OrganizationUnitID: request.OrganizationUnitID,
		Permissions:        request.Permissions,
		Assignments:        request.Assignments,
	}

	if err := rs.roleStore.CreateRole(role); err != nil {
		logger.Error("Failed to create role", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	logger.Debug("Successfully created role", log.String("id", role.ID), log.String("name", role.Name))
	return &role, nil
}

// GetRole retrieves a specific role by its id.
func (rs *roleService) GetRole(roleID string) (*Role, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Retrieving role", log.String("id", roleID))

	if roleID == "" {
		return nil, &ErrorMissingRoleID
	}

	role, err := rs.roleStore.GetRole(roleID)
	if err != nil {
		if errors.Is(err, ErrRoleNotFound) {
			logger.Debug("Role not found", log.String("id", roleID))
			return nil, &ErrorRoleNotFound
		}
		logger.Error("Failed to retrieve role", log.String("id", roleID), log.Error(err))
		return nil, &ErrorInternalServerError
	}

	logger.Debug("Successfully retrieved role", log.String("id", role.ID), log.String("name", role.Name))
	return &role, nil
}

// UpdateRole updates an existing role.
func (rs *roleService) UpdateRole(
	roleID string, request UpdateRoleRequest) (*Role, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Updating role", log.String("id", roleID), log.String("name", request.Name))

	if roleID == "" {
		return nil, &ErrorMissingRoleID
	}

	if err := rs.validateUpdateRoleRequest(request); err != nil {
		return nil, err
	}

	_, err := rs.roleStore.GetRole(roleID)
	if err != nil {
		if errors.Is(err, ErrRoleNotFound) {
			logger.Debug("Role not found", log.String("id", roleID))
			return nil, &ErrorRoleNotFound
		}
		logger.Error("Failed to retrieve role", log.String("id", roleID), log.Error(err))
		return nil, &ErrorInternalServerError
	}

	// Validate organization unit exists using OU service
	_, svcErr := rs.ouService.GetOrganizationUnit(request.OrganizationUnitID)
	if svcErr != nil {
		if svcErr.Code == oupkg.ErrorOrganizationUnitNotFound.Code {
			logger.Debug("Organization unit not found", log.String("ouID", request.OrganizationUnitID))
			return nil, &ErrorOrganizationUnitNotFound
		}
		logger.Error("Failed to validate organization unit", log.String("error", svcErr.Error))
		return nil, &ErrorInternalServerError
	}

	if err := rs.validateAssignmentIDs(request.Assignments); err != nil {
		return nil, err
	}

	updatedRole := Role{
		ID:                 roleID,
		Name:               request.Name,
		Description:        request.Description,
		OrganizationUnitID: request.OrganizationUnitID,
		Permissions:        request.Permissions,
		Assignments:        request.Assignments,
	}

	if err := rs.roleStore.UpdateRole(updatedRole); err != nil {
		logger.Error("Failed to update role", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	logger.Debug("Successfully updated role", log.String("id", roleID), log.String("name", request.Name))
	return &updatedRole, nil
}

// DeleteRole delete the specified role by its id.
func (rs *roleService) DeleteRole(roleID string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Deleting role", log.String("id", roleID))

	if roleID == "" {
		return &ErrorMissingRoleID
	}

	_, err := rs.roleStore.GetRole(roleID)
	if err != nil {
		if errors.Is(err, ErrRoleNotFound) {
			logger.Debug("Role not found", log.String("id", roleID))
			return nil
		}
		logger.Error("Failed to retrieve role", log.String("id", roleID), log.Error(err))
		return &ErrorInternalServerError
	}

	// Check if role has any assignments before deleting
	assignmentCount, err := rs.roleStore.GetRoleAssignmentsCount(roleID)
	if err != nil {
		logger.Error("Failed to get role assignments count", log.String("id", roleID), log.Error(err))
		return &ErrorInternalServerError
	}

	if assignmentCount > 0 {
		logger.Debug("Cannot delete role with active assignments",
			log.String("id", roleID), log.Int("assignmentCount", assignmentCount))
		return &ErrorCannotDeleteRole
	}

	if err := rs.roleStore.DeleteRole(roleID); err != nil {
		logger.Error("Failed to delete role", log.String("id", roleID), log.Error(err))
		return &ErrorInternalServerError
	}

	logger.Debug("Successfully deleted role", log.String("id", roleID))
	return nil
}

// GetRoleAssignments retrieves assignments for a role with pagination.
func (rs *roleService) GetRoleAssignments(roleID string, limit, offset int) (
	*AssignmentListResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}

	if roleID == "" {
		return nil, &ErrorMissingRoleID
	}

	_, err := rs.roleStore.GetRole(roleID)
	if err != nil {
		if errors.Is(err, ErrRoleNotFound) {
			logger.Debug("Role not found", log.String("id", roleID))
			return nil, &ErrorRoleNotFound
		}
		logger.Error("Failed to retrieve role", log.String("id", roleID), log.Error(err))
		return nil, &ErrorInternalServerError
	}

	totalCount, err := rs.roleStore.GetRoleAssignmentsCount(roleID)
	if err != nil {
		logger.Error("Failed to get role assignments count", log.String("roleID", roleID), log.Error(err))
		return nil, &ErrorInternalServerError
	}

	assignments, err := rs.roleStore.GetRoleAssignments(roleID, limit, offset)
	if err != nil {
		logger.Error("Failed to get role assignments", log.String("roleID", roleID), log.Error(err))
		return nil, &ErrorInternalServerError
	}

	baseURL := fmt.Sprintf("/roles/%s/assignments", roleID)
	links := buildPaginationLinks(baseURL, limit, offset, totalCount)

	response := &AssignmentListResponse{
		TotalResults: totalCount,
		Assignments:  assignments,
		StartIndex:   offset + 1,
		Count:        len(assignments),
		Links:        links,
	}

	return response, nil
}

// AddAssignments adds assignments to a role.
func (rs *roleService) AddAssignments(roleID string, request AssignmentsRequest) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Adding assignments to role", log.String("roleID", roleID))

	if roleID == "" {
		return &ErrorMissingRoleID
	}

	if err := rs.validateAssignmentsRequest(request); err != nil {
		return err
	}

	_, err := rs.roleStore.GetRole(roleID)
	if err != nil {
		if errors.Is(err, ErrRoleNotFound) {
			logger.Debug("Role not found", log.String("id", roleID))
			return &ErrorRoleNotFound
		}
		logger.Error("Failed to retrieve role", log.String("id", roleID), log.Error(err))
		return &ErrorInternalServerError
	}

	if err := rs.validateAssignmentIDs(request.Assignments); err != nil {
		return err
	}

	if err := rs.roleStore.AddAssignments(roleID, request.Assignments); err != nil {
		logger.Error("Failed to add assignments to role", log.String("roleID", roleID), log.Error(err))
		return &ErrorInternalServerError
	}

	logger.Debug("Successfully added assignments to role", log.String("roleID", roleID))
	return nil
}

// RemoveAssignments removes assignments from a role.
func (rs *roleService) RemoveAssignments(roleID string, request AssignmentsRequest) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Removing assignments from role", log.String("roleID", roleID))

	if roleID == "" {
		return &ErrorMissingRoleID
	}

	if err := rs.validateAssignmentsRequest(request); err != nil {
		return err
	}

	_, err := rs.roleStore.GetRole(roleID)
	if err != nil {
		if errors.Is(err, ErrRoleNotFound) {
			logger.Debug("Role not found", log.String("id", roleID))
			return &ErrorRoleNotFound
		}
		logger.Error("Failed to retrieve role", log.String("id", roleID), log.Error(err))
		return &ErrorInternalServerError
	}

	if err := rs.roleStore.RemoveAssignments(roleID, request.Assignments); err != nil {
		logger.Error("Failed to remove assignments from role", log.String("roleID", roleID), log.Error(err))
		return &ErrorInternalServerError
	}

	logger.Debug("Successfully removed assignments from role", log.String("roleID", roleID))
	return nil
}

// validateCreateRoleRequest validates the create role request.
func (rs *roleService) validateCreateRoleRequest(request CreateRoleRequest) *serviceerror.ServiceError {
	if request.Name == "" {
		return &ErrorInvalidRequestFormat
	}

	if request.OrganizationUnitID == "" {
		return &ErrorInvalidRequestFormat
	}

	if len(request.Permissions) == 0 {
		return &ErrorEmptyPermissions
	}

	for _, assignment := range request.Assignments {
		if assignment.Type != AssigneeTypeUser && assignment.Type != AssigneeTypeGroup {
			return &ErrorInvalidRequestFormat
		}
		if assignment.ID == "" {
			return &ErrorInvalidRequestFormat
		}
	}

	return nil
}

// validateUpdateRoleRequest validates the update role request.
func (rs *roleService) validateUpdateRoleRequest(request UpdateRoleRequest) *serviceerror.ServiceError {
	if request.Name == "" {
		return &ErrorInvalidRequestFormat
	}

	if request.OrganizationUnitID == "" {
		return &ErrorInvalidRequestFormat
	}

	if len(request.Permissions) == 0 {
		return &ErrorEmptyPermissions
	}

	for _, assignment := range request.Assignments {
		if assignment.Type != AssigneeTypeUser && assignment.Type != AssigneeTypeGroup {
			return &ErrorInvalidRequestFormat
		}
		if assignment.ID == "" {
			return &ErrorInvalidRequestFormat
		}
	}

	return nil
}

// validateAssignmentsRequest validates the assignments request.
func (rs *roleService) validateAssignmentsRequest(request AssignmentsRequest) *serviceerror.ServiceError {
	if len(request.Assignments) == 0 {
		return &ErrorEmptyAssignments
	}

	for _, assignment := range request.Assignments {
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
func (rs *roleService) validateAssignmentIDs(assignments []Assignment) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	var userIDs []string
	var groupIDs []string
	for _, assignment := range assignments {
		switch assignment.Type {
		case AssigneeTypeUser:
			userIDs = append(userIDs, assignment.ID)
		case AssigneeTypeGroup:
			groupIDs = append(groupIDs, assignment.ID)
		}
	}

	// Validate user IDs using user service
	if len(userIDs) > 0 {
		invalidUserIDs, svcErr := rs.userService.ValidateUserIDs(userIDs)
		if svcErr != nil {
			logger.Error("Failed to validate user IDs", log.String("error", svcErr.Error), log.String("code", svcErr.Code))
			return &ErrorInternalServerError
		}

		if len(invalidUserIDs) > 0 {
			logger.Debug("Invalid user IDs found", log.Any("invalidUserIDs", invalidUserIDs))
			return &ErrorInvalidAssignmentID
		}
	}

	// Validate group IDs using group service
	if len(groupIDs) > 0 {
		for _, groupID := range groupIDs {
			_, svcErr := rs.groupService.GetGroup(groupID)
			if svcErr != nil {
				if svcErr.Code == group.ErrorGroupNotFound.Code {
					logger.Debug("Invalid group ID found", log.String("groupID", groupID))
					return &ErrorInvalidAssignmentID
				}
				logger.Error("Failed to validate group ID", log.String("groupID", groupID), log.String("error", svcErr.Error))
				return &ErrorInternalServerError
			}
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
