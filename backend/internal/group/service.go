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
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/asgardeo/thunder/internal/entity"
	oupkg "github.com/asgardeo/thunder/internal/ou"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/security"
	"github.com/asgardeo/thunder/internal/system/sysauthz"
	"github.com/asgardeo/thunder/internal/system/transaction"
	"github.com/asgardeo/thunder/internal/system/utils"
	"github.com/asgardeo/thunder/internal/userschema"
)

const loggerComponentName = "GroupMgtService"

// GroupServiceInterface defines the interface for the group service.
type GroupServiceInterface interface {
	GetGroupList(ctx context.Context, limit, offset int,
		includeDisplay bool) (*GroupListResponse, *serviceerror.ServiceError)
	GetGroupsByPath(ctx context.Context, handlePath string, limit, offset int, includeDisplay bool) (
		*GroupListResponse, *serviceerror.ServiceError)
	CreateGroup(ctx context.Context, request CreateGroupRequest) (*Group, *serviceerror.ServiceError)
	CreateGroupByPath(ctx context.Context, handlePath string, request CreateGroupByPathRequest) (
		*Group, *serviceerror.ServiceError)
	GetGroup(ctx context.Context, groupID string, includeDisplay bool) (*Group, *serviceerror.ServiceError)
	UpdateGroup(ctx context.Context, groupID string, request UpdateGroupRequest) (*Group, *serviceerror.ServiceError)
	DeleteGroup(ctx context.Context, groupID string) *serviceerror.ServiceError
	GetGroupMembers(ctx context.Context, groupID string, limit, offset int, includeDisplay bool) (
		*MemberListResponse, *serviceerror.ServiceError)
	ValidateGroupIDs(ctx context.Context, groupIDs []string) *serviceerror.ServiceError
	GetGroupsByIDs(ctx context.Context, groupIDs []string) (map[string]*Group, *serviceerror.ServiceError)
	AddGroupMembers(ctx context.Context, groupID string, members []Member) (*Group, *serviceerror.ServiceError)
	RemoveGroupMembers(ctx context.Context, groupID string, members []Member) (*Group, *serviceerror.ServiceError)
}

// groupService is the default implementation of the GroupServiceInterface.
type groupService struct {
	groupStore        groupStoreInterface
	ouService         oupkg.OrganizationUnitServiceInterface
	entityService     entity.EntityServiceInterface
	userSchemaService userschema.UserSchemaServiceInterface
	transactioner     transaction.Transactioner
	authzService      sysauthz.SystemAuthorizationServiceInterface
}

// newGroupServiceWithStore creates a new instance of GroupService with an externally provided store.
func newGroupServiceWithStore(
	store groupStoreInterface,
	ouService oupkg.OrganizationUnitServiceInterface,
	entityService entity.EntityServiceInterface,
	userSchemaService userschema.UserSchemaServiceInterface,
	authzService sysauthz.SystemAuthorizationServiceInterface,
	transactioner transaction.Transactioner,
) GroupServiceInterface {
	return &groupService{
		groupStore:        store,
		ouService:         ouService,
		entityService:     entityService,
		userSchemaService: userSchemaService,
		authzService:      authzService,
		transactioner:     transactioner,
	}
}

// GetGroupList retrieves a list of groups. limit should be a positive integer & offset should be non-negative
// integer
func (gs *groupService) GetGroupList(ctx context.Context, limit, offset int, includeDisplay bool) (
	*GroupListResponse, *serviceerror.ServiceError) {
	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}

	accessibleOUs, svcErr := gs.getAccessibleOUs(ctx, security.ActionListGroups)
	if svcErr != nil {
		return nil, svcErr
	}

	if accessibleOUs.AllAllowed {
		return gs.listAllGroups(ctx, limit, offset, includeDisplay)
	}

	return gs.listGroupsByOUIDs(ctx, accessibleOUs.IDs, limit, offset, includeDisplay)
}

func (gs *groupService) listAllGroups(ctx context.Context, limit, offset int, includeDisplay bool) (
	*GroupListResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
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

	if includeDisplay {
		gs.populateGroupOUHandles(ctx, groupBasics, logger)
	}

	displayQuery := utils.DisplayQueryParam(includeDisplay)
	response := &GroupListResponse{
		TotalResults: totalCount,
		Groups:       groupBasics,
		StartIndex:   offset + 1,
		Count:        len(groupBasics),
		Links:        utils.BuildPaginationLinks("/groups", limit, offset, totalCount, displayQuery),
	}

	return response, nil
}

func (gs *groupService) listGroupsByOUIDs(ctx context.Context, ouIDs []string, limit, offset int,
	includeDisplay bool) (*GroupListResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	displayQuery := utils.DisplayQueryParam(includeDisplay)

	if len(ouIDs) == 0 {
		return &GroupListResponse{
			TotalResults: 0,
			Groups:       []GroupBasic{},
			StartIndex:   offset + 1,
			Count:        0,
			Links:        []utils.Link{},
		}, nil
	}

	totalCount, err := gs.groupStore.GetGroupListCountByOUIDs(ctx, ouIDs)
	if err != nil {
		logger.Error("Failed to get group count by OU IDs", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	if totalCount == 0 {
		return &GroupListResponse{
			TotalResults: 0,
			Groups:       []GroupBasic{},
			StartIndex:   offset + 1,
			Count:        0,
			Links:        []utils.Link{},
		}, nil
	}

	groups, err := gs.groupStore.GetGroupListByOUIDs(ctx, ouIDs, limit, offset)
	if err != nil {
		logger.Error("Failed to list groups by OU IDs", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	groupBasics := make([]GroupBasic, 0, len(groups))
	for _, groupDAO := range groups {
		groupBasics = append(groupBasics, buildGroupBasic(groupDAO))
	}

	if includeDisplay {
		gs.populateGroupOUHandles(ctx, groupBasics, logger)
	}

	response := &GroupListResponse{
		TotalResults: totalCount,
		Groups:       groupBasics,
		StartIndex:   offset + 1,
		Count:        len(groupBasics),
		Links:        utils.BuildPaginationLinks("/groups", limit, offset, totalCount, displayQuery),
	}

	return response, nil
}

// GetGroupsByPath retrieves a list of groups by hierarchical handle path.
func (gs *groupService) GetGroupsByPath(
	ctx context.Context, handlePath string, limit, offset int, includeDisplay bool,
) (*GroupListResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Getting groups by path", log.String("path", handlePath))

	serviceError := gs.validateAndProcessHandlePath(handlePath)
	if serviceError != nil {
		return nil, serviceError
	}

	ou, svcErr := gs.ouService.GetOrganizationUnitByPath(ctx, handlePath)
	if svcErr != nil {
		if svcErr.Code == oupkg.ErrorOrganizationUnitNotFound.Code {
			return nil, &ErrorGroupNotFound
		}
		return nil, svcErr
	}
	oUID := ou.ID

	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}

	if err := gs.checkGroupAccess(ctx, security.ActionListGroups, oUID, ""); err != nil {
		return nil, err
	}

	totalCount, err := gs.groupStore.GetGroupsByOrganizationUnitCount(ctx, oUID)
	if err != nil {
		logger.Error("Failed to get group count by organization unit", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	groups, err := gs.groupStore.GetGroupsByOrganizationUnit(ctx, oUID, limit, offset)
	if err != nil {
		logger.Error("Failed to list groups by organization unit", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	groupBasics := make([]GroupBasic, 0, len(groups))
	for _, groupDAO := range groups {
		g := buildGroupBasic(groupDAO)
		if includeDisplay {
			g.OUHandle = ou.Handle
		}
		groupBasics = append(groupBasics, g)
	}

	displayQuery := utils.DisplayQueryParam(includeDisplay)
	response := &GroupListResponse{
		TotalResults: totalCount,
		Groups:       groupBasics,
		StartIndex:   offset + 1,
		Count:        len(groupBasics),
		Links:        utils.BuildPaginationLinks("/groups/tree/"+handlePath, limit, offset, totalCount, displayQuery),
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

	if err := gs.validateOU(ctx, request.OUID); err != nil {
		return nil, err
	}

	if err := gs.checkGroupAccess(ctx, security.ActionCreateGroup, request.OUID, ""); err != nil {
		return nil, err
	}

	var userIDs []string
	var groupIDs []string
	var appIDs []string
	for _, member := range request.Members {
		switch member.Type {
		case MemberTypeUser:
			userIDs = append(userIDs, member.ID)
		case MemberTypeGroup:
			groupIDs = append(groupIDs, member.ID)
		case MemberTypeApp:
			appIDs = append(appIDs, member.ID)
		}
	}

	if err := gs.validateUserIDsWithAccess(ctx, userIDs); err != nil {
		return nil, err
	}

	if err := gs.validateAppIDs(ctx, appIDs); err != nil {
		return nil, err
	}

	if err := gs.ValidateGroupIDs(ctx, groupIDs); err != nil {
		return nil, err
	}

	var createdGroup *Group
	var capturedSvcErr *serviceerror.ServiceError

	err := gs.transactioner.Transact(ctx, func(txCtx context.Context) error {
		if err := gs.groupStore.CheckGroupNameConflictForCreate(
			txCtx, request.Name, request.OUID); err != nil {
			if errors.Is(err, ErrGroupNameConflict) {
				logger.Debug("Group name conflict detected", log.String("name", request.Name))
				capturedSvcErr = &ErrorGroupNameConflict
				return errors.New("rollback for group name conflict")
			}
			return err
		}

		groupDaoID, err := utils.GenerateUUIDv7()
		if err != nil {
			return err
		}

		groupDAO := GroupDAO{
			ID:          groupDaoID,
			Name:        request.Name,
			Description: request.Description,
			OUID:        request.OUID,
			Members:     request.Members,
		}

		if err := gs.groupStore.CreateGroup(txCtx, groupDAO); err != nil {
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
		logger.Error("Failed to create group", log.Error(err), log.String("name", request.Name))
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

	ou, svcErr := gs.ouService.GetOrganizationUnitByPath(ctx, handlePath)
	if svcErr != nil {
		if svcErr.Code == oupkg.ErrorOrganizationUnitNotFound.Code {
			return nil, &ErrorGroupNotFound
		}
		return nil, svcErr
	}

	// Convert CreateGroupByPathRequest to CreateGroupRequest
	createRequest := CreateGroupRequest{
		Name:        request.Name,
		Description: request.Description,
		OUID:        ou.ID,
		Members:     request.Members,
	}

	return gs.CreateGroup(ctx, createRequest)
}

// GetGroup retrieves a specific group by its id.
func (gs *groupService) GetGroup(
	ctx context.Context, groupID string, includeDisplay bool,
) (*Group, *serviceerror.ServiceError) {
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

	if err := gs.checkGroupAccess(ctx, security.ActionReadGroup, groupDAO.OUID, groupID); err != nil {
		return nil, err
	}

	group := convertGroupDAOToGroup(groupDAO)

	if includeDisplay {
		handleMap, svcErr := gs.ouService.GetOrganizationUnitHandlesByIDs(
			ctx, []string{group.OUID})
		if svcErr != nil {
			logger.Warn("Failed to resolve OU handle for group, skipping",
				log.String("id", groupID), log.Any("error", svcErr))
		} else if handle, ok := handleMap[group.OUID]; ok {
			group.OUHandle = handle
		}
	}

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
			return err
		}

		existingGroup := convertGroupDAOToGroup(existingGroupDAO)
		updateOUID := existingGroupDAO.OUID

		if gs.isOrganizationUnitChanged(existingGroup, request) {
			if err := gs.validateOU(txCtx, request.OUID); err != nil {
				capturedSvcErr = err
				return errors.New("rollback for invalid OU")
			}
			updateOUID = request.OUID
		}

		if err := gs.checkGroupAccess(
			txCtx,
			security.ActionUpdateGroup,
			existingGroupDAO.OUID,
			groupID,
		); err != nil {
			capturedSvcErr = err
			return errors.New("rollback for unauthorized access")
		}

		if updateOUID != existingGroupDAO.OUID {
			if err := gs.checkGroupAccess(
				txCtx,
				security.ActionUpdateGroup,
				updateOUID,
				groupID,
			); err != nil {
				capturedSvcErr = err
				return errors.New("rollback for unauthorized access to target OU")
			}
		}

		if existingGroup.Name != request.Name || existingGroup.OUID != request.OUID {
			err := gs.groupStore.CheckGroupNameConflictForUpdate(
				txCtx, request.Name, request.OUID, groupID)
			if err != nil {
				if errors.Is(err, ErrGroupNameConflict) {
					logger.Debug("Group name conflict detected during update", log.String("name", request.Name))
					capturedSvcErr = &ErrorGroupNameConflict
					return errors.New("rollback for group name conflict")
				}
				return err
			}
		}

		updatedGroupDAO := GroupDAO{
			ID:          existingGroup.ID,
			Name:        request.Name,
			Description: request.Description,
			OUID:        updateOUID,
		}

		if err := gs.groupStore.UpdateGroup(txCtx, updatedGroupDAO); err != nil {
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
		logger.Error("Failed to update group", log.Error(err), log.String("groupID", groupID))
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
		existingGroupDAO, err := gs.groupStore.GetGroup(txCtx, groupID)
		if err != nil {
			if errors.Is(err, ErrGroupNotFound) {
				logger.Debug("Group not found", log.String("id", groupID))
				capturedSvcErr = &ErrorGroupNotFound
				return errors.New("rollback for group not found")
			}
			return err
		}

		if err := gs.checkGroupAccess(
			txCtx,
			security.ActionDeleteGroup,
			existingGroupDAO.OUID,
			groupID,
		); err != nil {
			capturedSvcErr = err
			return errors.New("rollback for unauthorized access")
		}

		if err := gs.groupStore.DeleteGroup(txCtx, groupID); err != nil {
			return err
		}
		return nil
	})

	if capturedSvcErr != nil {
		return capturedSvcErr
	}

	if err != nil {
		logger.Error("Failed to delete group", log.Error(err), log.String("groupID", groupID))
		return &ErrorInternalServerError
	}

	logger.Debug("Successfully deleted group", log.String("id", groupID))
	return nil
}

// GetGroupMembers retrieves members of a group with pagination.
func (gs *groupService) GetGroupMembers(ctx context.Context, groupID string, limit, offset int,
	includeDisplay bool) (*MemberListResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}

	if groupID == "" {
		return nil, &ErrorMissingGroupID
	}

	existingGroupDAO, err := gs.groupStore.GetGroup(ctx, groupID)
	if err != nil {
		if errors.Is(err, ErrGroupNotFound) {
			logger.Debug("Group not found", log.String("id", groupID))
			return nil, &ErrorGroupNotFound
		}
		logger.Error("Failed to retrieve group", log.String("id", groupID), log.Error(err))
		return nil, &ErrorInternalServerError
	}

	if err := gs.checkGroupAccess(
		ctx,
		security.ActionReadGroup,
		existingGroupDAO.OUID,
		groupID,
	); err != nil {
		return nil, err
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

	if includeDisplay {
		gs.populateMemberDisplayNames(ctx, members, logger)
	}

	baseURL := fmt.Sprintf("/groups/%s/members", groupID)
	links := utils.BuildPaginationLinks(baseURL, limit, offset, totalCount, utils.DisplayQueryParam(includeDisplay))

	response := &MemberListResponse{
		TotalResults: totalCount,
		Members:      members,
		StartIndex:   offset + 1,
		Count:        len(members),
		Links:        links,
	}

	return response, nil
}

// populateMemberDisplayNames resolves display names for group members.
// For user members, it uses the schema-configured display attribute path.
// For group members, it uses the group name.
func (gs *groupService) populateMemberDisplayNames(ctx context.Context, members []Member, logger *log.Logger) {
	if len(members) == 0 {
		return
	}

	// Separate user, group, and app member IDs.
	var userIDs, groupIDs, appIDs []string
	for _, m := range members {
		switch m.Type {
		case MemberTypeUser:
			userIDs = append(userIDs, m.ID)
		case MemberTypeGroup:
			groupIDs = append(groupIDs, m.ID)
		case MemberTypeApp:
			appIDs = append(appIDs, m.ID)
		}
	}

	// Batch-fetch users and resolve display attribute paths.
	var entityMap map[string]*entity.Entity
	var displayAttrPaths map[string]string
	if len(userIDs) > 0 {
		entities, err := gs.entityService.GetEntitiesByIDs(ctx, userIDs)
		if err != nil {
			logger.Warn("Failed to batch-fetch users for display resolution", log.Error(err))
		} else {
			entityMap = make(map[string]*entity.Entity, len(entities))
			var userTypes []string
			for i := range entities {
				entityMap[entities[i].ID] = &entities[i]
				userTypes = append(userTypes, entities[i].Type)
			}
			displayAttrPaths = resolveDisplayAttributePaths(ctx, userTypes, gs.userSchemaService, logger)
		}
	}

	// Batch-fetch groups for group member display names.
	var groupsMap map[string]*Group
	if len(groupIDs) > 0 {
		var svcErr *serviceerror.ServiceError
		groupsMap, svcErr = gs.GetGroupsByIDs(ctx, groupIDs)
		if svcErr != nil {
			logger.Warn("Failed to batch-fetch groups for display resolution", log.Any("error", svcErr))
		}
	}

	// Batch-fetch app entities for app member display names.
	appNamesMap := make(map[string]string)
	if len(appIDs) > 0 && gs.entityService != nil {
		entities, err := gs.entityService.GetEntitiesByIDs(ctx, appIDs)
		if err != nil {
			logger.Warn("Failed to batch-fetch app entities for display resolution", log.Any("error", err))
		} else {
			for _, e := range entities {
				appNamesMap[e.ID] = resolveEntityDisplayName(e)
			}
		}
	}

	// Set display on each member.
	for i := range members {
		switch members[i].Type {
		case MemberTypeUser:
			if entityMap != nil {
				if e, ok := entityMap[members[i].ID]; ok {
					members[i].Display = utils.ResolveDisplay(e.ID, e.Type, e.Attributes, displayAttrPaths)
					continue
				}
			}
			members[i].Display = members[i].ID
		case MemberTypeGroup:
			if groupsMap != nil {
				if g, ok := groupsMap[members[i].ID]; ok && g.Name != "" {
					members[i].Display = g.Name
					continue
				}
			}
			members[i].Display = members[i].ID
		case MemberTypeApp:
			if name, ok := appNamesMap[members[i].ID]; ok && name != "" {
				members[i].Display = name
				continue
			}
			members[i].Display = members[i].ID
		}
	}
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
		existingGroupDAO, err := gs.groupStore.GetGroup(txCtx, groupID)
		if err != nil {
			if errors.Is(err, ErrGroupNotFound) {
				logger.Debug("Group not found", log.String("id", groupID))
				capturedSvcErr = &ErrorGroupNotFound
				return errors.New("rollback for group not found")
			}
			return err
		}

		if err := gs.checkGroupAccess(
			txCtx,
			security.ActionUpdateGroup,
			existingGroupDAO.OUID,
			groupID,
		); err != nil {
			capturedSvcErr = err
			return errors.New("rollback for unauthorized access")
		}

		var userIDs []string
		var groupIDs []string
		var appIDs []string
		for _, member := range members {
			switch member.Type {
			case MemberTypeUser:
				userIDs = append(userIDs, member.ID)
			case MemberTypeGroup:
				groupIDs = append(groupIDs, member.ID)
			case MemberTypeApp:
				appIDs = append(appIDs, member.ID)
			}
		}

		if svcErr := gs.validateUserIDsWithAccess(txCtx, userIDs); svcErr != nil {
			capturedSvcErr = svcErr
			return errors.New("rollback for invalid user IDs")
		}

		if svcErr := gs.validateAppIDs(txCtx, appIDs); svcErr != nil {
			capturedSvcErr = svcErr
			return errors.New("rollback for invalid app IDs")
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
		existingGroupDAO, err := gs.groupStore.GetGroup(txCtx, groupID)
		if err != nil {
			if errors.Is(err, ErrGroupNotFound) {
				logger.Debug("Group not found", log.String("id", groupID))
				capturedSvcErr = &ErrorGroupNotFound
				return errors.New("rollback for group not found")
			}
			return err
		}

		if err := gs.checkGroupAccess(
			txCtx,
			security.ActionUpdateGroup,
			existingGroupDAO.OUID,
			groupID,
		); err != nil {
			capturedSvcErr = err
			return errors.New("rollback for unauthorized access")
		}

		if err := gs.groupStore.RemoveGroupMembers(txCtx, groupID, members); err != nil {
			return err
		}

		groupDAO, err := gs.groupStore.GetGroup(txCtx, groupID)
		if err != nil {
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

	if request.OUID == "" {
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

	if request.OUID == "" {
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
	return existingGroup.OUID != request.OUID
}

// validateOU validates that provided organization unit ID exist.
func (gs *groupService) validateOU(ctx context.Context, ouID string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	isExists, err := gs.ouService.IsOrganizationUnitExists(ctx, ouID)
	if err != nil {
		logger.Error("Failed to check organization unit existence", log.Any("error: ", err))
		return &ErrorInternalServerError
	}

	if !isExists {
		return &ErrorInvalidOUID
	}

	return nil
}

// validateUserIDs validates that all provided user IDs exist.
func (gs *groupService) validateUserIDs(ctx context.Context, userIDs []string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	invalidUserIDs, err := gs.entityService.ValidateEntityIDs(ctx, userIDs)
	if err != nil {
		logger.Error("Failed to validate user IDs", log.Error(err))
		return &ErrorInternalServerError
	}

	if len(invalidUserIDs) > 0 {
		logger.Debug("Invalid user IDs found", log.Any("invalidUserIDs", invalidUserIDs))
		return &ErrorInvalidUserMemberID
	}

	return nil
}

// validateUserIDsWithAccess validates user IDs in two steps:
//  1. Existence check — returns ErrorInvalidUserMemberID for any unknown user ID.
//  2. OU-scope check — returns ErrorUnauthorized for any user ID outside the caller's accessible OUs
//     (OUs where the caller holds group-update permission).
//
// When the caller is a full admin (AllAllowed), only step 1 is performed.
func (gs *groupService) validateUserIDsWithAccess(
	ctx context.Context, userIDs []string,
) *serviceerror.ServiceError {
	if len(userIDs) == 0 {
		return nil
	}
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if err := gs.validateUserIDs(ctx, userIDs); err != nil {
		return err
	}

	accessibleOUs, svcErr := gs.getAccessibleOUs(ctx, security.ActionUpdateGroup)
	if svcErr != nil {
		return svcErr
	}

	if accessibleOUs.AllAllowed {
		// Full admin — no scope restriction.
		return nil
	}

	outOfScopeIDs, err := gs.entityService.ValidateEntityIDsInOUs(ctx, userIDs, accessibleOUs.IDs)
	if err != nil {
		logger.Error("Failed to validate user IDs in OUs", log.Error(err))
		return &ErrorInternalServerError
	}

	if len(outOfScopeIDs) > 0 {
		logger.Debug("User IDs outside accessible OUs", log.Any("outOfScopeIDs", outOfScopeIDs))
		return &serviceerror.ErrorUnauthorized
	}

	return nil
}

// validateAppIDs validates that all provided app entity IDs exist.
func (gs *groupService) validateAppIDs(ctx context.Context, appIDs []string) *serviceerror.ServiceError {
	if len(appIDs) == 0 {
		return nil
	}
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if gs.entityService == nil {
		logger.Error("Entity service not configured, cannot validate app IDs")
		return &ErrorInternalServerError
	}

	invalidIDs, err := gs.entityService.ValidateEntityIDs(ctx, appIDs)
	if err != nil {
		logger.Error("Failed to validate app IDs", log.String("error", err.Error()))
		return &ErrorInternalServerError
	}

	if len(invalidIDs) > 0 {
		logger.Debug("Invalid app IDs found", log.Any("invalidAppIDs", invalidIDs))
		return &ErrorInvalidUserMemberID
	}

	return nil
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
// Falls back to the entity ID if no name is found.
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

// GetGroupsByIDs retrieves groups by a list of IDs.
func (gs *groupService) GetGroupsByIDs(
	ctx context.Context, groupIDs []string,
) (map[string]*Group, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if len(groupIDs) == 0 {
		return map[string]*Group{}, nil
	}

	// Deduplicate IDs before passing to store.
	seen := make(map[string]struct{}, len(groupIDs))
	uniqueIDs := make([]string, 0, len(groupIDs))
	for _, id := range groupIDs {
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			uniqueIDs = append(uniqueIDs, id)
		}
	}

	groupDAOs, err := gs.groupStore.GetGroupsByIDs(ctx, uniqueIDs)
	if err != nil {
		logger.Error("Failed to get groups by IDs", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	result := make(map[string]*Group, len(groupDAOs))
	for _, dao := range groupDAOs {
		group := convertGroupDAOToGroup(GroupDAO{
			ID:          dao.ID,
			Name:        dao.Name,
			Description: dao.Description,
			OUID:        dao.OUID,
		})
		result[dao.ID] = &group
	}

	return result, nil
}

// convertGroupDAOToGroup constructs a Group from a GroupDAO.
func convertGroupDAOToGroup(groupDAO GroupDAO) Group {
	return Group{
		ID:          groupDAO.ID,
		Name:        groupDAO.Name,
		Description: groupDAO.Description,
		OUID:        groupDAO.OUID,
		Members:     groupDAO.Members,
	}
}

// buildGroupBasic constructs a GroupBasic from a GroupBasicDAO.
func buildGroupBasic(groupDAO GroupBasicDAO) GroupBasic {
	return GroupBasic{
		ID:          groupDAO.ID,
		Name:        groupDAO.Name,
		Description: groupDAO.Description,
		OUID:        groupDAO.OUID,
	}
}

// populateGroupOUHandles resolves OU handles for a slice of groups in-place.
func (gs *groupService) populateGroupOUHandles(ctx context.Context, groups []GroupBasic, logger *log.Logger) {
	ouIDs := make([]string, 0, len(groups))
	seen := make(map[string]bool, len(groups))
	for _, g := range groups {
		if g.OUID != "" && !seen[g.OUID] {
			ouIDs = append(ouIDs, g.OUID)
			seen[g.OUID] = true
		}
	}

	handleMap, svcErr := gs.ouService.GetOrganizationUnitHandlesByIDs(ctx, ouIDs)
	if svcErr != nil {
		logger.Warn("Failed to resolve OU handles for groups, skipping", log.Any("error", svcErr))
		return
	}

	for i := range groups {
		if handle, ok := handleMap[groups[i].OUID]; ok {
			groups[i].OUHandle = handle
		}
	}
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

// checkGroupAccess performs an authorization check on the group resource against the current caller.
func (gs *groupService) checkGroupAccess(
	ctx context.Context, action security.Action, ouID string, groupID string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	actionCtx := sysauthz.ActionContext{
		ResourceType: security.ResourceTypeGroup,
		OUID:         ouID,
		ResourceID:   groupID,
	}

	hasAccess, err := gs.authzService.IsActionAllowed(ctx, action, &actionCtx)
	if err != nil {
		logger.Error("Failed to check authorization", log.String("err", err.Error))
		return &ErrorInternalServerError
	}
	if !hasAccess {
		return &serviceerror.ErrorUnauthorized
	}
	return nil
}

// getAccessibleOUs retrieves the accessible resources for the group.
func (gs *groupService) getAccessibleOUs(
	ctx context.Context, action security.Action) (*sysauthz.AccessibleResources, *serviceerror.ServiceError) {
	accessibleResources, err := gs.authzService.GetAccessibleResources(ctx, action, security.ResourceTypeOU)
	if err != nil {
		return nil, err
	}
	return accessibleResources, nil
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
