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

// Package service provides the implementation for group management operations.
package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/asgardeo/thunder/internal/group/constants"
	"github.com/asgardeo/thunder/internal/group/model"
	"github.com/asgardeo/thunder/internal/group/store"
	ouPkg "github.com/asgardeo/thunder/internal/ou"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
	userservice "github.com/asgardeo/thunder/internal/user/service"
)

const loggerComponentName = "GroupMgtService"

// GroupServiceInterface defines the interface for the group service.
type GroupServiceInterface interface {
	GetGroupList(limit, offset int) (*model.GroupListResponse, *serviceerror.ServiceError)
	GetGroupsByPath(handlePath string, limit, offset int) (*model.GroupListResponse, *serviceerror.ServiceError)
	CreateGroup(request model.CreateGroupRequest) (*model.Group, *serviceerror.ServiceError)
	CreateGroupByPath(handlePath string, request model.CreateGroupByPathRequest) (*model.Group, *serviceerror.ServiceError)
	GetGroup(groupID string) (*model.Group, *serviceerror.ServiceError)
	UpdateGroup(groupID string, request model.UpdateGroupRequest) (*model.Group, *serviceerror.ServiceError)
	DeleteGroup(groupID string) *serviceerror.ServiceError
	GetGroupMembers(groupID string, limit, offset int) (*model.MemberListResponse, *serviceerror.ServiceError)
}

// GroupService is the default implementation of the GroupServiceInterface.
type GroupService struct{}

// GetGroupService creates a new instance of GroupService.
func GetGroupService() GroupServiceInterface {
	return &GroupService{}
}

// GetGroupList retrieves a list of groups. limit should be a positive integer & offset should be non-negative
// integer
func (gs *GroupService) GetGroupList(limit, offset int) (*model.GroupListResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}

	totalCount, err := store.GetGroupListCount()
	if err != nil {
		logger.Error("Failed to get group count", log.Error(err))
		return nil, &constants.ErrorInternalServerError
	}

	groups, err := store.GetGroupList(limit, offset)
	if err != nil {
		logger.Error("Failed to list groups", log.Error(err))
		return nil, &constants.ErrorInternalServerError
	}

	groupBasics := make([]model.GroupBasic, 0, len(groups))
	for _, groupDAO := range groups {
		groupBasics = append(groupBasics, buildGroupBasic(groupDAO))
	}

	response := &model.GroupListResponse{
		TotalResults: totalCount,
		Groups:       groupBasics,
		StartIndex:   offset + 1,
		Count:        len(groupBasics),
		Links:        buildPaginationLinks("/groups", limit, offset, totalCount),
	}

	return response, nil
}

// GetGroupsByPath retrieves a list of groups by hierarchical handle path.
func (gs *GroupService) GetGroupsByPath(
	handlePath string, limit, offset int,
) (*model.GroupListResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Getting groups by path", log.String("path", handlePath))

	serviceError := gs.validateAndProcessHandlePath(handlePath)
	if serviceError != nil {
		return nil, serviceError
	}

	ouService := ouPkg.GetOrganizationUnitService()
	ou, svcErr := ouService.GetOrganizationUnitByPath(handlePath)
	if svcErr != nil {
		if svcErr.Code == ouPkg.ErrorOrganizationUnitNotFound.Code {
			return nil, &constants.ErrorGroupNotFound
		}
		return nil, svcErr
	}
	organizationUnitID := ou.ID

	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}

	totalCount, err := store.GetGroupsByOrganizationUnitCount(organizationUnitID)
	if err != nil {
		logger.Error("Failed to get group count by organization unit", log.Error(err))
		return nil, &constants.ErrorInternalServerError
	}

	groups, err := store.GetGroupsByOrganizationUnit(organizationUnitID, limit, offset)
	if err != nil {
		logger.Error("Failed to list groups by organization unit", log.Error(err))
		return nil, &constants.ErrorInternalServerError
	}

	groupBasics := make([]model.GroupBasic, 0, len(groups))
	for _, groupDAO := range groups {
		groupBasics = append(groupBasics, buildGroupBasic(groupDAO))
	}

	response := &model.GroupListResponse{
		TotalResults: totalCount,
		Groups:       groupBasics,
		StartIndex:   offset + 1,
		Count:        len(groupBasics),
		Links:        buildPaginationLinks("/groups", limit, offset, totalCount),
	}

	return response, nil
}

// CreateGroup creates a new group.
func (gs *GroupService) CreateGroup(request model.CreateGroupRequest) (*model.Group, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Creating group", log.String("name", request.Name))

	if err := gs.validateCreateGroupRequest(request); err != nil {
		return nil, err
	}

	if err := gs.validateOU(request.OrganizationUnitID); err != nil {
		return nil, err
	}

	var userIDs []string
	var groupIDs []string
	for _, member := range request.Members {
		switch member.Type {
		case model.MemberTypeUser:
			userIDs = append(userIDs, member.ID)
		case model.MemberTypeGroup:
			groupIDs = append(groupIDs, member.ID)
		}
	}

	if err := gs.validateUserIDs(userIDs); err != nil {
		return nil, err
	}

	if err := gs.validateGroupIDs(groupIDs); err != nil {
		return nil, err
	}

	if err := store.CheckGroupNameConflictForCreate(request.Name, request.OrganizationUnitID); err != nil {
		if errors.Is(err, constants.ErrGroupNameConflict) {
			logger.Debug("Group name conflict detected", log.String("name", request.Name))
			return nil, &constants.ErrorGroupNameConflict
		}
		logger.Error("Failed to check group name conflict", log.Error(err))
		return nil, &constants.ErrorInternalServerError
	}

	groupDAO := model.GroupDAO{
		ID:                 utils.GenerateUUID(),
		Name:               request.Name,
		Description:        request.Description,
		OrganizationUnitID: request.OrganizationUnitID,
		Members:            request.Members,
	}

	if err := store.CreateGroup(groupDAO); err != nil {
		logger.Error("Failed to create group", log.Error(err))
		return nil, &constants.ErrorInternalServerError
	}

	group := convertGroupDAOToGroup(groupDAO)
	logger.Debug("Successfully created group", log.String("id", groupDAO.ID), log.String("name", groupDAO.Name))
	return &group, nil
}

// CreateGroupByPath creates a new group under the organization unit specified by the handle path.
func (gs *GroupService) CreateGroupByPath(
	handlePath string, request model.CreateGroupByPathRequest,
) (*model.Group, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Creating group by path", log.String("path", handlePath), log.String("name", request.Name))

	serviceError := gs.validateAndProcessHandlePath(handlePath)
	if serviceError != nil {
		return nil, serviceError
	}

	ouService := ouPkg.GetOrganizationUnitService()
	ou, svcErr := ouService.GetOrganizationUnitByPath(handlePath)
	if svcErr != nil {
		if svcErr.Code == ouPkg.ErrorOrganizationUnitNotFound.Code {
			return nil, &constants.ErrorGroupNotFound
		}
		return nil, svcErr
	}

	// Convert CreateGroupByPathRequest to CreateGroupRequest
	createRequest := model.CreateGroupRequest{
		Name:               request.Name,
		Description:        request.Description,
		OrganizationUnitID: ou.ID,
		Members:            request.Members,
	}

	return gs.CreateGroup(createRequest)
}

// GetGroup retrieves a specific group by its id.
func (gs *GroupService) GetGroup(groupID string) (*model.Group, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Retrieving group", log.String("id", groupID))

	if groupID == "" {
		return nil, &constants.ErrorMissingGroupID
	}

	groupDAO, err := store.GetGroup(groupID)
	if err != nil {
		if errors.Is(err, constants.ErrGroupNotFound) {
			logger.Debug("Group not found", log.String("id", groupID))
			return nil, &constants.ErrorGroupNotFound
		}
		logger.Error("Failed to retrieve group", log.String("id", groupID), log.Error(err))
		return nil, &constants.ErrorInternalServerError
	}

	group := convertGroupDAOToGroup(groupDAO)
	logger.Debug("Successfully retrieved group", log.String("id", group.ID), log.String("name", group.Name))
	return &group, nil
}

// UpdateGroup updates an existing group.
func (gs *GroupService) UpdateGroup(
	groupID string, request model.UpdateGroupRequest) (*model.Group, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Updating group", log.String("id", groupID), log.String("name", request.Name))

	if groupID == "" {
		return nil, &constants.ErrorMissingGroupID
	}

	if err := gs.validateUpdateGroupRequest(request); err != nil {
		return nil, err
	}

	existingGroupDAO, err := store.GetGroup(groupID)
	if err != nil {
		if errors.Is(err, constants.ErrGroupNotFound) {
			logger.Debug("Group not found", log.String("id", groupID))
			return nil, &constants.ErrorGroupNotFound
		}
		logger.Error("Failed to retrieve group", log.String("id", groupID), log.Error(err))
		return nil, &constants.ErrorInternalServerError
	}

	existingGroup := convertGroupDAOToGroup(existingGroupDAO)
	updateOrganizationUnitID := existingGroupDAO.OrganizationUnitID

	if gs.isOrganizationUnitChanged(existingGroup, request) {
		if err := gs.validateOU(request.OrganizationUnitID); err != nil {
			return nil, err
		}
		updateOrganizationUnitID = request.OrganizationUnitID
	}

	var userIDs []string
	var groupIDs []string
	for _, member := range request.Members {
		switch member.Type {
		case model.MemberTypeUser:
			userIDs = append(userIDs, member.ID)
		case model.MemberTypeGroup:
			groupIDs = append(groupIDs, member.ID)
		}
	}

	if err := gs.validateUserIDs(userIDs); err != nil {
		return nil, err
	}

	if err := gs.validateGroupIDs(groupIDs); err != nil {
		return nil, err
	}

	if existingGroup.Name != request.Name || existingGroup.OrganizationUnitID != request.OrganizationUnitID {
		if err := store.CheckGroupNameConflictForUpdate(request.Name, request.OrganizationUnitID, groupID); err != nil {
			if errors.Is(err, constants.ErrGroupNameConflict) {
				logger.Debug("Group name conflict detected during update", log.String("name", request.Name))
				return nil, &constants.ErrorGroupNameConflict
			}
			logger.Error("Failed to check group name conflict during update", log.Error(err))
			return nil, &constants.ErrorInternalServerError
		}
	}

	updatedGroupDAO := model.GroupDAO{
		ID:                 existingGroup.ID,
		Name:               request.Name,
		Description:        request.Description,
		OrganizationUnitID: updateOrganizationUnitID,
		Members:            request.Members,
	}

	if err := store.UpdateGroup(updatedGroupDAO); err != nil {
		logger.Error("Failed to update group", log.Error(err))
		return nil, &constants.ErrorInternalServerError
	}

	updatedGroup := convertGroupDAOToGroup(updatedGroupDAO)
	logger.Debug("Successfully updated group", log.String("id", groupID), log.String("name", request.Name))
	return &updatedGroup, nil
}

// DeleteGroup delete the specified group by its id.
func (gs *GroupService) DeleteGroup(groupID string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Deleting group", log.String("id", groupID))

	if groupID == "" {
		return &constants.ErrorMissingGroupID
	}

	_, err := store.GetGroup(groupID)
	if err != nil {
		if errors.Is(err, constants.ErrGroupNotFound) {
			logger.Debug("Group not found", log.String("id", groupID))
			return &constants.ErrorGroupNotFound
		}
		logger.Error("Failed to retrieve group", log.String("id", groupID), log.Error(err))
		return &constants.ErrorInternalServerError
	}

	if err := store.DeleteGroup(groupID); err != nil {
		logger.Error("Failed to delete group", log.String("id", groupID), log.Error(err))
		return &constants.ErrorInternalServerError
	}

	logger.Debug("Successfully deleted group", log.String("id", groupID))
	return nil
}

// GetGroupMembers retrieves members of a group with pagination.
func (gs *GroupService) GetGroupMembers(groupID string, limit, offset int) (
	*model.MemberListResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}

	if groupID == "" {
		return nil, &constants.ErrorMissingGroupID
	}

	_, err := store.GetGroup(groupID)
	if err != nil {
		if errors.Is(err, constants.ErrGroupNotFound) {
			logger.Debug("Group not found", log.String("id", groupID))
			return nil, &constants.ErrorGroupNotFound
		}
		logger.Error("Failed to retrieve group", log.String("id", groupID), log.Error(err))
		return nil, &constants.ErrorInternalServerError
	}

	totalCount, err := store.GetGroupMemberCount(groupID)
	if err != nil {
		logger.Error("Failed to get group member count", log.String("groupID", groupID), log.Error(err))
		return nil, &constants.ErrorInternalServerError
	}

	members, err := store.GetGroupMembers(groupID, limit, offset)
	if err != nil {
		logger.Error("Failed to get group members", log.String("groupID", groupID), log.Error(err))
		return nil, &constants.ErrorInternalServerError
	}

	baseURL := fmt.Sprintf("/groups/%s/members", groupID)
	links := buildPaginationLinks(baseURL, limit, offset, totalCount)

	response := &model.MemberListResponse{
		TotalResults: totalCount,
		Members:      members,
		StartIndex:   offset + 1,
		Count:        len(members),
		Links:        links,
	}

	return response, nil
}

// validateCreateGroupRequest validates the create group request.
func (gs *GroupService) validateCreateGroupRequest(request model.CreateGroupRequest) *serviceerror.ServiceError {
	if request.Name == "" {
		return &constants.ErrorInvalidRequestFormat
	}

	if request.OrganizationUnitID == "" {
		return &constants.ErrorInvalidRequestFormat
	}

	for _, member := range request.Members {
		if member.Type != model.MemberTypeUser && member.Type != model.MemberTypeGroup {
			return &constants.ErrorInvalidRequestFormat
		}
		if member.ID == "" {
			return &constants.ErrorInvalidRequestFormat
		}
	}

	return nil
}

// validateUpdateGroupRequest validates the update group request.
func (gs *GroupService) validateUpdateGroupRequest(request model.UpdateGroupRequest) *serviceerror.ServiceError {
	if request.Name == "" {
		return &constants.ErrorInvalidRequestFormat
	}

	if request.OrganizationUnitID == "" {
		return &constants.ErrorInvalidRequestFormat
	}

	for _, member := range request.Members {
		if member.Type != model.MemberTypeUser && member.Type != model.MemberTypeGroup {
			return &constants.ErrorInvalidRequestFormat
		}
		if member.ID == "" {
			return &constants.ErrorInvalidRequestFormat
		}
	}

	return nil
}

// isOrganizationUnitChanged checks if the organization unit of the group has changed during an update.
func (gs *GroupService) isOrganizationUnitChanged(existingGroup model.Group, request model.UpdateGroupRequest) bool {
	return existingGroup.OrganizationUnitID != request.OrganizationUnitID
}

// validateOU validates that provided organization unit ID exist.
func (gs *GroupService) validateOU(ouID string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	ouService := ouPkg.GetOrganizationUnitService()
	_, err := ouService.GetOrganizationUnit(ouID)
	if err != nil {
		if err.Code == ouPkg.ErrorOrganizationUnitNotFound.Code {
			return &constants.ErrorInvalidOUID
		} else {
			logger.Error("Failed to get organization unit", log.Any("error: ", err))
			return &constants.ErrorInternalServerError
		}
	}

	return nil
}

// validateUserIDs validates that all provided user IDs exist.
func (gs *GroupService) validateUserIDs(userIDs []string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	userService := userservice.GetUserService()
	invalidUserIDs, svcErr := userService.ValidateUserIDs(userIDs)
	if svcErr != nil {
		logger.Error("Failed to validate user IDs", log.String("error", svcErr.Error), log.String("code", svcErr.Code))
		return &constants.ErrorInternalServerError
	}

	if len(invalidUserIDs) > 0 {
		logger.Debug("Invalid user IDs found", log.Any("invalidUserIDs", invalidUserIDs))
		return &constants.ErrorInvalidUserMemberID
	}

	return nil
}

// validateGroupIDs validates that all provided group IDs exist.
func (gs *GroupService) validateGroupIDs(groupIDs []string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	invalidGroupIDs, err := store.ValidateGroupIDs(groupIDs)
	if err != nil {
		logger.Error("Failed to validate group IDs", log.Error(err))
		return &constants.ErrorInternalServerError
	}

	if len(invalidGroupIDs) > 0 {
		logger.Debug("Invalid group IDs found", log.Any("invalidGroupIDs", invalidGroupIDs))
		return &constants.ErrorInvalidGroupMemberID
	}

	return nil
}

// convertGroupDAOToGroup constructs a model.Group from a model.GroupDAO.
func convertGroupDAOToGroup(groupDAO model.GroupDAO) model.Group {
	return model.Group(groupDAO)
}

// buildGroupBasic constructs a model.GroupBasic from a model.GroupBasicDAO.
func buildGroupBasic(groupDAO model.GroupBasicDAO) model.GroupBasic {
	return model.GroupBasic(groupDAO)
}

// validatePaginationParams validates pagination parameters.
func validatePaginationParams(limit, offset int) *serviceerror.ServiceError {
	if limit < 1 || limit > serverconst.MaxPageSize {
		return &constants.ErrorInvalidLimit
	}
	if offset < 0 {
		return &constants.ErrorInvalidOffset
	}
	return nil
}

// buildPaginationLinks builds pagination links for the response.
func buildPaginationLinks(base string, limit, offset, totalCount int) []model.Link {
	links := make([]model.Link, 0)

	if offset > 0 {
		links = append(links, model.Link{
			Href: fmt.Sprintf("%s?offset=0&limit=%d", base, limit),
			Rel:  "first",
		})

		prevOffset := offset - limit
		if prevOffset < 0 {
			prevOffset = 0
		}
		links = append(links, model.Link{
			Href: fmt.Sprintf("%s?offset=%d&limit=%d", base, prevOffset, limit),
			Rel:  "prev",
		})
	}

	if offset+limit < totalCount {
		nextOffset := offset + limit
		links = append(links, model.Link{
			Href: fmt.Sprintf("%s?offset=%d&limit=%d", base, nextOffset, limit),
			Rel:  "next",
		})
	}

	lastPageOffset := ((totalCount - 1) / limit) * limit
	if offset < lastPageOffset {
		links = append(links, model.Link{
			Href: fmt.Sprintf("%s?offset=%d&limit=%d", base, lastPageOffset, limit),
			Rel:  "last",
		})
	}

	return links
}

// validateAndProcessHandlePath validates and processes the handle path.
func (gs *GroupService) validateAndProcessHandlePath(handlePath string) *serviceerror.ServiceError {
	if strings.TrimSpace(handlePath) == "" {
		return &constants.ErrorInvalidRequestFormat
	}

	handles := strings.Split(strings.Trim(handlePath, "/"), "/")
	if len(handles) == 0 {
		return &constants.ErrorInvalidRequestFormat
	}

	for _, handle := range handles {
		if strings.TrimSpace(handle) == "" {
			return &constants.ErrorInvalidRequestFormat
		}
	}
	return nil
}
