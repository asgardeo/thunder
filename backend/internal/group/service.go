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

// Package group provides group management functionality.
package group

import (
	"context"
	"errors"
	"fmt"
	"strings"

	oupkg "github.com/asgardeo/thunder/internal/ou"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/database/transaction"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
	"github.com/asgardeo/thunder/internal/user"
)

const loggerComponentName = "GroupMgtService"

// GroupServiceInterface defines the interface for the group service.
type GroupServiceInterface interface {
	GetGroupList(ctx context.Context, limit, offset int) (*GroupListResponse, *serviceerror.ServiceError)
	GetGroupsByPath(ctx context.Context, handlePath string, limit, offset int) (
		*GroupListResponse, *serviceerror.ServiceError)
	CreateGroup(ctx context.Context, request CreateGroupRequest) (*Group, *serviceerror.ServiceError)
	CreateGroupByPath(ctx context.Context, handlePath string, request CreateGroupByPathRequest) (
		*Group, *serviceerror.ServiceError)
	GetGroup(ctx context.Context, groupID string) (*Group, *serviceerror.ServiceError)
	UpdateGroup(ctx context.Context, groupID string, request UpdateGroupRequest) (*Group, *serviceerror.ServiceError)
	DeleteGroup(ctx context.Context, groupID string) *serviceerror.ServiceError
	GetGroupMembers(ctx context.Context, groupID string, limit, offset int) (
		*MemberListResponse, *serviceerror.ServiceError)
	ValidateGroupIDs(ctx context.Context, groupIDs []string) *serviceerror.ServiceError
	AddGroupMembers(ctx context.Context, groupID string, members []Member) (*Group, *serviceerror.ServiceError)
	RemoveGroupMembers(ctx context.Context, groupID string, members []Member) (*Group, *serviceerror.ServiceError)
}

// groupService is the default implementation of the GroupServiceInterface.
type groupService struct {
	groupStore    groupStoreInterface
	ouService     oupkg.OrganizationUnitServiceInterface
	userService   user.UserServiceInterface
	transactioner transaction.Transactioner
}

// newGroupService creates a new instance of GroupService with injected dependencies.
func newGroupService(
	ouService oupkg.OrganizationUnitServiceInterface,
	userService user.UserServiceInterface,
	transactioner transaction.Transactioner,
) GroupServiceInterface {
	return &groupService{
		groupStore:    newGroupStore(),
		ouService:     ouService,
		userService:   userService,
		transactioner: transactioner,
	}
}

// GetGroupList retrieves a list of groups. limit should be a positive integer & offset should be non-negative
// integer
func (gs *groupService) GetGroupList(ctx context.Context, limit, offset int) (
	*GroupListResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}

	totalCount, err := gs.groupStore.GetGroupListCount(ctx)
	if err != nil {
		logger.Error("Failed to get group count", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	groups, err := gs.groupStore.GetGroupList(ctx, limit, offset)
	if err != nil {
		logger.Error("Failed to list groups", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	groupBasics := make([]GroupBasic, 0, len(groups))
	for _, groupDAO := range groups {
		groupBasics = append(groupBasics, buildGroupBasic(groupDAO))
	}

	response := &GroupListResponse{
		TotalResults: totalCount,
		Groups:       groupBasics,
		StartIndex:   offset + 1,
		Count:        len(groupBasics),
		Links:        buildPaginationLinks("/groups", limit, offset, totalCount),
	}

	return response, nil
}

// GetGroupsByPath retrieves a list of groups by hierarchical handle path.
func (gs *groupService) GetGroupsByPath(
	ctx context.Context, handlePath string, limit, offset int,
) (*GroupListResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Getting groups by path", log.String("path", handlePath))

	serviceError := gs.validateAndProcessHandlePath(handlePath)
	if serviceError != nil {
		return nil, serviceError
	}

	ou, svcErr := gs.ouService.GetOrganizationUnitByPath(handlePath)
	if svcErr != nil {
		if svcErr.Code == oupkg.ErrorOrganizationUnitNotFound.Code {
			return nil, &ErrorGroupNotFound
		}
		return nil, svcErr
	}
	organizationUnitID := ou.ID

	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}

	totalCount, err := gs.groupStore.GetGroupsByOrganizationUnitCount(ctx, organizationUnitID)
	if err != nil {
		logger.Error("Failed to get group count by organization unit", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	groups, err := gs.groupStore.GetGroupsByOrganizationUnit(ctx, organizationUnitID, limit, offset)
	if err != nil {
		logger.Error("Failed to list groups by organization unit", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	groupBasics := make([]GroupBasic, 0, len(groups))
	for _, groupDAO := range groups {
		groupBasics = append(groupBasics, buildGroupBasic(groupDAO))
	}

	response := &GroupListResponse{
		TotalResults: totalCount,
		Groups:       groupBasics,
		StartIndex:   offset + 1,
		Count:        len(groupBasics),
		Links:        buildPaginationLinks("/groups", limit, offset, totalCount),
	}

	return response, nil
}

// CreateGroup creates a new group.
func (gs *groupService) CreateGroup(ctx context.Context, request CreateGroupRequest) (
	*Group, *serviceerror.ServiceError) {
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
		case MemberTypeUser:
			userIDs = append(userIDs, member.ID)
		case MemberTypeGroup:
			groupIDs = append(groupIDs, member.ID)
		}
	}

	if err := gs.validateUserIDs(ctx, userIDs); err != nil {
		return nil, err
	}

	if err := gs.ValidateGroupIDs(ctx, groupIDs); err != nil {
		return nil, err
	}

	var createdGroup *Group
	var capturedSvcErr *serviceerror.ServiceError

	err := gs.transactioner.Transact(ctx, func(txCtx context.Context) error {
		if err := gs.groupStore.CheckGroupNameConflictForCreate(
			txCtx, request.Name, request.OrganizationUnitID); err != nil {
			if errors.Is(err, ErrGroupNameConflict) {
				logger.Debug("Group name conflict detected", log.String("name", request.Name))
				capturedSvcErr = &ErrorGroupNameConflict
				return errors.New("rollback for group name conflict")
			}
			logger.Error("Failed to check group name conflict", log.Error(err))
			capturedSvcErr = &ErrorInternalServerError
			return errors.New("rollback for internal server error")
		}

		groupDaoID, err := utils.GenerateUUIDv7()
		if err != nil {
			logger.Error("Failed to generate UUID", log.Error(err))
			capturedSvcErr = &serviceerror.InternalServerError
			return errors.New("rollback for UUID generation failure")
		}

		groupDAO := GroupDAO{
			ID:                 groupDaoID,
			Name:               request.Name,
			Description:        request.Description,
			OrganizationUnitID: request.OrganizationUnitID,
			Members:            request.Members,
		}

		if err := gs.groupStore.CreateGroup(txCtx, groupDAO); err != nil {
			logger.Error("Failed to create group", log.Error(err))
			capturedSvcErr = &ErrorInternalServerError
			return err
		}

		group := convertGroupDAOToGroup(groupDAO)
		createdGroup = &group
		return nil
	})

	if capturedSvcErr != nil {
		return nil, capturedSvcErr
	}

	if err != nil {
		return nil, &ErrorInternalServerError
	}

	logger.Debug("Successfully created group", log.String("id", createdGroup.ID), log.String("name", createdGroup.Name))
	return createdGroup, nil
}

// CreateGroupByPath creates a new group under the organization unit specified by the handle path.
func (gs *groupService) CreateGroupByPath(
	ctx context.Context, handlePath string, request CreateGroupByPathRequest,
) (*Group, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Creating group by path", log.String("path", handlePath), log.String("name", request.Name))

	serviceError := gs.validateAndProcessHandlePath(handlePath)
	if serviceError != nil {
		return nil, serviceError
	}

	ou, svcErr := gs.ouService.GetOrganizationUnitByPath(handlePath)
	if svcErr != nil {
		if svcErr.Code == oupkg.ErrorOrganizationUnitNotFound.Code {
			return nil, &ErrorGroupNotFound
		}
		return nil, svcErr
	}

	// Convert CreateGroupByPathRequest to CreateGroupRequest
	createRequest := CreateGroupRequest{
		Name:               request.Name,
		Description:        request.Description,
		OrganizationUnitID: ou.ID,
		Members:            request.Members,
	}

	return gs.CreateGroup(ctx, createRequest)
}

// GetGroup retrieves a specific group by its id.
func (gs *groupService) GetGroup(ctx context.Context, groupID string) (*Group, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Retrieving group", log.String("id", groupID))

	if groupID == "" {
		return nil, &ErrorMissingGroupID
	}

	groupDAO, err := gs.groupStore.GetGroup(ctx, groupID)
	if err != nil {
		if errors.Is(err, ErrGroupNotFound) {
			logger.Debug("Group not found", log.String("id", groupID))
			return nil, &ErrorGroupNotFound
		}
		logger.Error("Failed to retrieve group", log.String("id", groupID), log.Error(err))
		return nil, &ErrorInternalServerError
	}

	group := convertGroupDAOToGroup(groupDAO)
	logger.Debug("Successfully retrieved group", log.String("id", group.ID), log.String("name", group.Name))
	return &group, nil
}

// UpdateGroup updates an existing group.
func (gs *groupService) UpdateGroup(
	ctx context.Context, groupID string, request UpdateGroupRequest) (*Group, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Updating group", log.String("id", groupID), log.String("name", request.Name))

	if groupID == "" {
		return nil, &ErrorMissingGroupID
	}

	if err := gs.validateUpdateGroupRequest(request); err != nil {
		return nil, err
	}

	var updatedGroup *Group
	var capturedSvcErr *serviceerror.ServiceError

	err := gs.transactioner.Transact(ctx, func(txCtx context.Context) error {
		existingGroupDAO, err := gs.groupStore.GetGroup(txCtx, groupID)
		if err != nil {
			if errors.Is(err, ErrGroupNotFound) {
				logger.Debug("Group not found", log.String("id", groupID))
				capturedSvcErr = &ErrorGroupNotFound
				return errors.New("rollback for group not found")
			}
			logger.Error("Failed to retrieve group", log.String("id", groupID), log.Error(err))
			capturedSvcErr = &ErrorInternalServerError
			return err
		}

		existingGroup := convertGroupDAOToGroup(existingGroupDAO)
		updateOrganizationUnitID := existingGroupDAO.OrganizationUnitID

		if gs.isOrganizationUnitChanged(existingGroup, request) {
			if err := gs.validateOU(request.OrganizationUnitID); err != nil {
				capturedSvcErr = err
				return errors.New("rollback for invalid OU")
			}
			updateOrganizationUnitID = request.OrganizationUnitID
		}

		if existingGroup.Name != request.Name || existingGroup.OrganizationUnitID != request.OrganizationUnitID {
			err := gs.groupStore.CheckGroupNameConflictForUpdate(
				txCtx, request.Name, request.OrganizationUnitID, groupID)
			if err != nil {
				if errors.Is(err, ErrGroupNameConflict) {
					logger.Debug("Group name conflict detected during update", log.String("name", request.Name))
					capturedSvcErr = &ErrorGroupNameConflict
					return errors.New("rollback for group name conflict")
				}
				logger.Error("Failed to check group name conflict during update", log.Error(err))
				capturedSvcErr = &ErrorInternalServerError
				return err
			}
		}

		updatedGroupDAO := GroupDAO{
			ID:                 existingGroup.ID,
			Name:               request.Name,
			Description:        request.Description,
			OrganizationUnitID: updateOrganizationUnitID,
		}

		if err := gs.groupStore.UpdateGroup(txCtx, updatedGroupDAO); err != nil {
			logger.Error("Failed to update group", log.Error(err))
			capturedSvcErr = &ErrorInternalServerError
			return err
		}

		group := convertGroupDAOToGroup(updatedGroupDAO)
		updatedGroup = &group
		return nil
	})

	if capturedSvcErr != nil {
		return nil, capturedSvcErr
	}

	if err != nil {
		return nil, &ErrorInternalServerError
	}

	logger.Debug("Successfully updated group", log.String("id", groupID), log.String("name", request.Name))
	return updatedGroup, nil
}

// DeleteGroup delete the specified group by its id.
func (gs *groupService) DeleteGroup(ctx context.Context, groupID string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Deleting group", log.String("id", groupID))

	if groupID == "" {
		return &ErrorMissingGroupID
	}

	var capturedSvcErr *serviceerror.ServiceError

	err := gs.transactioner.Transact(ctx, func(txCtx context.Context) error {
		_, err := gs.groupStore.GetGroup(txCtx, groupID)
		if err != nil {
			if errors.Is(err, ErrGroupNotFound) {
				logger.Debug("Group not found", log.String("id", groupID))
				capturedSvcErr = &ErrorGroupNotFound
				return errors.New("rollback for group not found")
			}
			logger.Error("Failed to retrieve group", log.String("id", groupID), log.Error(err))
			capturedSvcErr = &ErrorInternalServerError
			return err
		}

		if err := gs.groupStore.DeleteGroup(txCtx, groupID); err != nil {
			logger.Error("Failed to delete group", log.String("id", groupID), log.Error(err))
			capturedSvcErr = &ErrorInternalServerError
			return err
		}
		return nil
	})

	if capturedSvcErr != nil {
		return capturedSvcErr
	}

	if err != nil {
		return &ErrorInternalServerError
	}

	logger.Debug("Successfully deleted group", log.String("id", groupID))
	return nil
}

// GetGroupMembers retrieves members of a group with pagination.
func (gs *groupService) GetGroupMembers(ctx context.Context, groupID string, limit, offset int) (
	*MemberListResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}

	if groupID == "" {
		return nil, &ErrorMissingGroupID
	}

	_, err := gs.groupStore.GetGroup(ctx, groupID)
	if err != nil {
		if errors.Is(err, ErrGroupNotFound) {
			logger.Debug("Group not found", log.String("id", groupID))
			return nil, &ErrorGroupNotFound
		}
		logger.Error("Failed to retrieve group", log.String("id", groupID), log.Error(err))
		return nil, &ErrorInternalServerError
	}

	totalCount, err := gs.groupStore.GetGroupMemberCount(ctx, groupID)
	if err != nil {
		logger.Error("Failed to get group member count", log.String("groupID", groupID), log.Error(err))
		return nil, &ErrorInternalServerError
	}

	members, err := gs.groupStore.GetGroupMembers(ctx, groupID, limit, offset)
	if err != nil {
		logger.Error("Failed to get group members", log.String("groupID", groupID), log.Error(err))
		return nil, &ErrorInternalServerError
	}

	baseURL := fmt.Sprintf("/groups/%s/members", groupID)
	links := buildPaginationLinks(baseURL, limit, offset, totalCount)

	response := &MemberListResponse{
		TotalResults: totalCount,
		Members:      members,
		StartIndex:   offset + 1,
		Count:        len(members),
		Links:        links,
	}

	return response, nil
}

// AddGroupMembers adds members to a group.
func (gs *groupService) AddGroupMembers(
	ctx context.Context, groupID string, members []Member) (*Group, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Adding members to group", log.String("id", groupID))

	if groupID == "" {
		return nil, &ErrorMissingGroupID
	}

	if len(members) == 0 {
		return nil, &ErrorEmptyMembers
	}

	if svcErr := validateMemberTypes(members); svcErr != nil {
		return nil, svcErr
	}

	var capturedSvcErr *serviceerror.ServiceError
	var updatedGroupDAO GroupDAO

	err := gs.transactioner.Transact(ctx, func(txCtx context.Context) error {
		_, err := gs.groupStore.GetGroup(txCtx, groupID)
		if err != nil {
			if errors.Is(err, ErrGroupNotFound) {
				logger.Debug("Group not found", log.String("id", groupID))
				capturedSvcErr = &ErrorGroupNotFound
				return errors.New("rollback for group not found")
			}
			logger.Error("Failed to check group existence", log.String("id", groupID), log.Error(err))
			capturedSvcErr = &ErrorInternalServerError
			return err
		}

		var userIDs []string
		var groupIDs []string
		for _, member := range members {
			switch member.Type {
			case MemberTypeUser:
				userIDs = append(userIDs, member.ID)
			case MemberTypeGroup:
				groupIDs = append(groupIDs, member.ID)
			}
		}

		if svcErr := gs.validateUserIDs(txCtx, userIDs); svcErr != nil {
			capturedSvcErr = svcErr
			return errors.New("rollback for invalid user IDs")
		}

		if svcErr := gs.ValidateGroupIDs(txCtx, groupIDs); svcErr != nil {
			capturedSvcErr = svcErr
			return errors.New("rollback for invalid group IDs")
		}

		if err := gs.groupStore.AddGroupMembers(txCtx, groupID, members); err != nil {
			return err
		}

		groupDAO, err := gs.groupStore.GetGroup(txCtx, groupID)
		if err != nil {
			logger.Error("Failed to retrieve updated group", log.String("id", groupID), log.Error(err))
			capturedSvcErr = &ErrorInternalServerError
			return err
		}
		updatedGroupDAO = groupDAO

		return nil
	})

	if capturedSvcErr != nil {
		return nil, capturedSvcErr
	}

	if err != nil {
		logger.Error("Failed to add members to group", log.String("id", groupID), log.Error(err))
		return nil, &ErrorInternalServerError
	}

	updatedGroup := convertGroupDAOToGroup(updatedGroupDAO)
	logger.Debug("Successfully added members to group", log.String("id", groupID))
	return &updatedGroup, nil
}

// RemoveGroupMembers removes members from a group.
func (gs *groupService) RemoveGroupMembers(
	ctx context.Context, groupID string, members []Member) (*Group, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Removing members from group", log.String("id", groupID))

	if groupID == "" {
		return nil, &ErrorMissingGroupID
	}

	if len(members) == 0 {
		return nil, &ErrorEmptyMembers
	}

	if svcErr := validateMemberTypes(members); svcErr != nil {
		return nil, svcErr
	}

	var capturedSvcErr *serviceerror.ServiceError
	var updatedGroupDAO GroupDAO

	err := gs.transactioner.Transact(ctx, func(txCtx context.Context) error {
		_, err := gs.groupStore.GetGroup(txCtx, groupID)
		if err != nil {
			if errors.Is(err, ErrGroupNotFound) {
				logger.Debug("Group not found", log.String("id", groupID))
				capturedSvcErr = &ErrorGroupNotFound
				return errors.New("rollback for group not found")
			}
			logger.Error("Failed to check group existence", log.String("id", groupID), log.Error(err))
			capturedSvcErr = &ErrorInternalServerError
			return err
		}

		if err := gs.groupStore.RemoveGroupMembers(txCtx, groupID, members); err != nil {
			return err
		}

		groupDAO, err := gs.groupStore.GetGroup(txCtx, groupID)
		if err != nil {
			logger.Error("Failed to retrieve updated group", log.String("id", groupID), log.Error(err))
			capturedSvcErr = &ErrorInternalServerError
			return err
		}
		updatedGroupDAO = groupDAO

		return nil
	})

	if capturedSvcErr != nil {
		return nil, capturedSvcErr
	}

	if err != nil {
		logger.Error("Failed to remove members from group", log.String("id", groupID), log.Error(err))
		return nil, &ErrorInternalServerError
	}

	updatedGroup := convertGroupDAOToGroup(updatedGroupDAO)
	logger.Debug("Successfully removed members from group", log.String("id", groupID))
	return &updatedGroup, nil
}

// validateCreateGroupRequest validates the create group request.
func (gs *groupService) validateCreateGroupRequest(request CreateGroupRequest) *serviceerror.ServiceError {
	if request.Name == "" {
		return &ErrorInvalidRequestFormat
	}

	if request.OrganizationUnitID == "" {
		return &ErrorInvalidRequestFormat
	}

	for _, member := range request.Members {
		if member.Type != MemberTypeUser && member.Type != MemberTypeGroup {
			return &ErrorInvalidRequestFormat
		}
		if member.ID == "" {
			return &ErrorInvalidRequestFormat
		}
	}

	return nil
}

// validateUpdateGroupRequest validates the update group request.
func (gs *groupService) validateUpdateGroupRequest(request UpdateGroupRequest) *serviceerror.ServiceError {
	if request.Name == "" {
		return &ErrorInvalidRequestFormat
	}

	if request.OrganizationUnitID == "" {
		return &ErrorInvalidRequestFormat
	}

	return nil
}

// validateMemberTypes validates that all members have a valid type.
func validateMemberTypes(members []Member) *serviceerror.ServiceError {
	for _, member := range members {
		if member.Type != MemberTypeUser && member.Type != MemberTypeGroup {
			return &ErrorInvalidRequestFormat
		}
		if member.ID == "" {
			return &ErrorInvalidRequestFormat
		}
	}
	return nil
}

// isOrganizationUnitChanged checks if the organization unit of the group has changed during an update.
func (gs *groupService) isOrganizationUnitChanged(existingGroup Group, request UpdateGroupRequest) bool {
	return existingGroup.OrganizationUnitID != request.OrganizationUnitID
}

// validateOU validates that provided organization unit ID exist.
func (gs *groupService) validateOU(ouID string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	_, err := gs.ouService.GetOrganizationUnit(ouID)
	if err != nil {
		if err.Code == oupkg.ErrorOrganizationUnitNotFound.Code {
			return &ErrorInvalidOUID
		} else {
			logger.Error("Failed to get organization unit", log.Any("error: ", err))
			return &ErrorInternalServerError
		}
	}

	return nil
}

// validateUserIDs validates that all provided user IDs exist.
func (gs *groupService) validateUserIDs(ctx context.Context, userIDs []string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	invalidUserIDs, svcErr := gs.userService.ValidateUserIDs(ctx, userIDs)
	if svcErr != nil {
		logger.Error("Failed to validate user IDs", log.String("error", svcErr.Error), log.String("code", svcErr.Code))
		return &ErrorInternalServerError
	}

	if len(invalidUserIDs) > 0 {
		logger.Debug("Invalid user IDs found", log.Any("invalidUserIDs", invalidUserIDs))
		return &ErrorInvalidUserMemberID
	}

	return nil
}

// ValidateGroupIDs validates that all provided group IDs exist.
func (gs *groupService) ValidateGroupIDs(ctx context.Context, groupIDs []string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	invalidGroupIDs, err := gs.groupStore.ValidateGroupIDs(ctx, groupIDs)
	if err != nil {
		logger.Error("Failed to validate group IDs", log.Error(err))
		return &ErrorInternalServerError
	}

	if len(invalidGroupIDs) > 0 {
		logger.Debug("Invalid group IDs found", log.Any("invalidGroupIDs", invalidGroupIDs))
		return &ErrorInvalidGroupMemberID
	}

	return nil
}

// convertGroupDAOToGroup constructs a Group from a GroupDAO.
func convertGroupDAOToGroup(groupDAO GroupDAO) Group {
	return Group(groupDAO)
}

// buildGroupBasic constructs a GroupBasic from a GroupBasicDAO.
func buildGroupBasic(groupDAO GroupBasicDAO) GroupBasic {
	return GroupBasic(groupDAO)
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

// validateAndProcessHandlePath validates and processes the handle path.
func (gs *groupService) validateAndProcessHandlePath(handlePath string) *serviceerror.ServiceError {
	trimmedPath := strings.TrimSpace(handlePath)
	if trimmedPath == "" {
		return &ErrorInvalidRequestFormat
	}

	trimmedPath = strings.Trim(trimmedPath, "/")
	if trimmedPath == "" {
		return &ErrorInvalidRequestFormat
	}

	handles := strings.Split(trimmedPath, "/")
	for _, handle := range handles {
		if strings.TrimSpace(handle) == "" {
			return &ErrorInvalidRequestFormat
		}
	}
	return nil
}
