/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
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

	"github.com/asgardeo/thunder/internal/group/constants"
	"github.com/asgardeo/thunder/internal/group/model"
	"github.com/asgardeo/thunder/internal/group/store"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
)

const loggerComponentName = "GroupMgtService"

// GroupServiceInterface defines the interface for the group service.
type GroupServiceInterface interface {
	GetGroupList() ([]model.GroupBasic, *serviceerror.ServiceError)
	CreateGroup(request model.CreateGroupRequest) (*model.Group, *serviceerror.ServiceError)
	GetGroup(groupID string) (*model.Group, *serviceerror.ServiceError)
	UpdateGroup(groupID string, request model.UpdateGroupRequest) (*model.Group, *serviceerror.ServiceError)
	DeleteGroup(groupID string) *serviceerror.ServiceError
}

// GroupService is the default implementation of the GroupServiceInterface.
type GroupService struct{}

// GetGroupService creates a new instance of GroupService.
func GetGroupService() GroupServiceInterface {
	return &GroupService{}
}

// GetGroupList retrieves a list of root groups.
func (gs *GroupService) GetGroupList() ([]model.GroupBasic, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Listing all groups")

	groups, err := store.GetGroupList()
	if err != nil {
		logger.Error("Failed to list groups", log.Error(err))
		return nil, &constants.ErrorInternalServerError
	}

	return groups, nil
}

// CreateGroup creates a new group.
func (gs *GroupService) CreateGroup(request model.CreateGroupRequest) (*model.Group, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Creating group", log.String("name", request.Name))

	// Validate request
	if err := gs.validateCreateGroupRequest(request); err != nil {
		return nil, err
	}

	// Check if group with same name already exists under the same parent
	// Note: This check could be improved by adding a GetGroupByNameAndParent method to the store
	// For now, we'll rely on the database constraints to catch duplicates
	logger.Debug("Skipping duplicate name check - relying on database constraints")

	// Validate parent exists
	if err := gs.validateParentExists(request.Parent); err != nil {
		return nil, err
	}

	// Create group object
	group := model.Group{
		ID:          utils.GenerateUUID(),
		Name:        request.Name,
		Description: request.Description,
		Parent:      request.Parent,
		Users:       request.Users,
		Groups:      []string{},
	}

	// Create group in the database
	if err := store.CreateGroup(group); err != nil {
		logger.Error("Failed to create group", log.Error(err))
		return nil, &constants.ErrorInternalServerError
	}

	logger.Debug("Successfully created group", log.String("id", group.ID), log.String("name", group.Name))
	return &group, nil
}

// GetGroup retrieves a specific group by its id.
func (gs *GroupService) GetGroup(groupID string) (*model.Group, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Retrieving group", log.String("id", groupID))

	if groupID == "" {
		return nil, &constants.ErrorMissingGroupID
	}

	group, err := store.GetGroup(groupID)
	if err != nil {
		if errors.Is(err, model.ErrGroupNotFound) {
			logger.Debug("Group not found", log.String("id", groupID))
			return nil, &constants.ErrorGroupNotFound
		}
		logger.Error("Failed to retrieve group", log.String("id", groupID), log.Error(err))
		return nil, &constants.ErrorInternalServerError
	}

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

	// Validate request
	if err := gs.validateUpdateGroupRequest(request); err != nil {
		return nil, err
	}

	// Get existing group to ensure it exists
	existingGroup, err := store.GetGroup(groupID)
	if err != nil {
		if errors.Is(err, model.ErrGroupNotFound) {
			logger.Debug("Group not found", log.String("id", groupID))
			return nil, &constants.ErrorGroupNotFound
		}
		logger.Error("Failed to retrieve group", log.String("id", groupID), log.Error(err))
		return nil, &constants.ErrorInternalServerError
	}

	// Check for duplicate name under the same parent
	// Note: This check could be improved by adding a GetGroupByNameAndParent method to the store
	// For now, we'll rely on the database constraints to catch duplicates if the name changes
	if existingGroup.Name != request.Name {
		logger.Debug("Group name changed - relying on database constraints for duplicate check")
	}

	// Validate parent exists if parent is changed
	if existingGroup.Parent.ID != request.Parent.ID || existingGroup.Parent.Type != request.Parent.Type {
		if err := gs.validateParentExists(request.Parent); err != nil {
			return nil, err
		}
	}

	// Create updated group object
	updatedGroup := model.Group{
		ID:          existingGroup.ID,
		Name:        request.Name,
		Description: request.Description,
		Parent:      request.Parent,
		Users:       request.Users,
		Groups:      request.Groups,
	}

	// Update group in the database
	if err := store.UpdateGroup(updatedGroup); err != nil {
		logger.Error("Failed to update group", log.Error(err))
		return nil, &constants.ErrorInternalServerError
	}

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

	// Check if group exists
	_, err := store.GetGroup(groupID)
	if err != nil {
		if errors.Is(err, model.ErrGroupNotFound) {
			logger.Debug("Group not found", log.String("id", groupID))
			return &constants.ErrorGroupNotFound
		}
		logger.Error("Failed to retrieve group", log.String("id", groupID), log.Error(err))
		return &constants.ErrorInternalServerError
	}

	// Check if group can be deleted (no child groups)
	if err := gs.validateForDeleteGroup(groupID); err != nil {
		return err
	}

	// Delete the group
	if err := store.DeleteGroup(groupID); err != nil {
		logger.Error("Failed to delete group", log.String("id", groupID), log.Error(err))
		return &constants.ErrorInternalServerError
	}

	logger.Debug("Successfully deleted group", log.String("id", groupID))
	return nil
}

// validateCreateGroupRequest validates the create group request.
func (gs *GroupService) validateCreateGroupRequest(request model.CreateGroupRequest) *serviceerror.ServiceError {
	if request.Name == "" {
		return &constants.ErrorInvalidRequestFormat
	}

	if request.Parent.Type == "" || request.Parent.ID == "" {
		return &constants.ErrorInvalidRequestFormat
	}

	// Use ParentType constants for validation
	if request.Parent.Type != model.ParentTypeGroup && request.Parent.Type != model.ParentTypeOrganizationUnit {
		return &constants.ErrorInvalidRequestFormat
	}

	return nil
}

// validateUpdateGroupRequest validates the update group request.
func (gs *GroupService) validateUpdateGroupRequest(request model.UpdateGroupRequest) *serviceerror.ServiceError {
	if request.Name == "" {
		return &constants.ErrorInvalidRequestFormat
	}

	if request.Parent.Type == "" || request.Parent.ID == "" {
		return &constants.ErrorInvalidRequestFormat
	}

	// Use ParentType constants for validation
	if request.Parent.Type != model.ParentTypeGroup && request.Parent.Type != model.ParentTypeOrganizationUnit {
		return &constants.ErrorInvalidRequestFormat
	}

	return nil
}

// validateParentExists validates that the parent group or organization unit exists.
func (gs *GroupService) validateParentExists(parent model.Parent) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if parent.Type == model.ParentTypeGroup {
		// Check if parent group exists
		_, err := store.GetGroup(parent.ID)
		if err != nil {
			if errors.Is(err, model.ErrGroupNotFound) {
				logger.Debug("Parent group not found", log.String("parentID", parent.ID))
				return &constants.ErrorParentNotFound
			}
			logger.Error("Failed to check parent group existence", log.String("parentID", parent.ID), log.Error(err))
			return &constants.ErrorInternalServerError
		}
	} else if parent.Type == model.ParentTypeOrganizationUnit {
		// TODO: Add validation for organization unit existence
		// For now, we'll assume it exists
		logger.Debug("Organization unit validation not implemented", log.String("parentID", parent.ID))
	}

	return nil
}

// validateForDeleteGroup checks if the group can be deleted.
func (gs *GroupService) validateForDeleteGroup(groupID string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	childGroups, err := store.GetChildGroups(groupID)
	if err != nil {
		logger.Error("Failed to check for child groups", log.String("groupID", groupID), log.Error(err))
		return &constants.ErrorInternalServerError
	}

	if childGroups != nil && len(*childGroups) > 0 {
		logger.Debug("Cannot delete group with child groups", log.String("groupID", groupID),
			log.Int("childCount", len(*childGroups)))
		return &constants.ErrorCannotDeleteGroup
	}

	return nil
}
