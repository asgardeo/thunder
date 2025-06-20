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

// Package store provides the implementation for group persistence operations.
package store

import (
	"github.com/asgardeo/thunder/internal/group/model"
	"github.com/asgardeo/thunder/internal/system/database/client"
	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/log"

	"fmt"
)

const loggerComponentName = "GroupStore"

// GroupType represents the type group entity.
const GroupType = "group"

// GetGroupList retrieves all root groups.
func GetGroupList() ([]model.GroupBasic, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	dbClient, err := provider.NewDBProvider().GetDBClient("identity")
	if err != nil {
		logger.Error("Failed to get database client", log.Error(err))
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}
	defer func() {
		if closeErr := dbClient.Close(); closeErr != nil {
			logger.Error("Failed to close database client", log.Error(closeErr))
		}
	}()

	var results []map[string]interface{}

	results, err = dbClient.Query(QueryGetGroupList)
	if err != nil {
		logger.Error("Failed to execute query", log.Error(err))
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	groups := make([]model.GroupBasic, 0)
	for _, row := range results {
		group, err := buildGroupFromResultRow(row, logger)
		if err != nil {
			logger.Error("Failed to build group from result row", log.Error(err))
			return nil, fmt.Errorf("failed to build group from result row: %w", err)
		}

		groupBasic := model.GroupBasic{
			ID:          group.ID,
			Name:        group.Name,
			Description: group.Description,
			Parent:      group.Parent,
		}

		groups = append(groups, groupBasic)
	}

	return groups, nil
}

// CreateGroup creates a new group in the database.
func CreateGroup(group model.Group) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	dbClient, err := provider.NewDBProvider().GetDBClient("identity")
	if err != nil {
		logger.Error("Failed to get database client", log.Error(err))
		return fmt.Errorf("failed to get database client: %w", err)
	}
	defer func() {
		if closeErr := dbClient.Close(); closeErr != nil {
			logger.Error("Failed to close database client", log.Error(closeErr))
		}
	}()

	// Check for name conflicts
	err = checkGroupNameConflictForCreate(dbClient, group.Name, group.Parent, logger)
	if err != nil {
		return err
	}

	// Determine the parent group id and OU id
	var parentGroupID *string
	var ouID string

	if group.Parent.Type == GroupType {
		parentGroupID = &group.Parent.ID
		// Get the OU id from the parent group
		parentGroup, err := GetGroup(group.Parent.ID)
		if err != nil {
			logger.Error("Failed to get parent group", log.Error(err))
			return model.ErrParentNotFound
		}
		// Convert Group to GroupBasic for getOUFromPath function
		parentGroupBasic := model.GroupBasic{
			ID:          parentGroup.ID,
			Name:        parentGroup.Name,
			Description: parentGroup.Description,
			Parent:      parentGroup.Parent,
		}
		ouID = getOUFromPath(parentGroupBasic)
	} else {
		ouID = group.Parent.ID
	}

	// Generate path
	path := generateGroupPath(group.Name, group.Parent)

	_, err = dbClient.Execute(
		QueryCreateGroup,
		group.ID,
		parentGroupID,
		ouID,
		group.Name,
		group.Description,
		path,
	)
	if err != nil {
		logger.Error("Failed to execute create group query", log.Error(err))
		return fmt.Errorf("failed to execute query: %w", err)
	}

	// Add users to the group
	err = addUsersToGroup(dbClient, group.ID, group.Users, logger)
	if err != nil {
		return err
	}

	return nil
}

// GetGroup retrieves a group by its id.
func GetGroup(id string) (model.Group, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	dbClient, err := provider.NewDBProvider().GetDBClient("identity")
	if err != nil {
		logger.Error("Failed to get database client", log.Error(err))
		return model.Group{}, fmt.Errorf("failed to get database client: %w", err)
	}
	defer func() {
		if closeErr := dbClient.Close(); closeErr != nil {
			logger.Error("Failed to close database client", log.Error(closeErr))
		}
	}()

	results, err := dbClient.Query(QueryGetGroupByID, id)
	if err != nil {
		logger.Error("Failed to execute query", log.Error(err))
		return model.Group{}, fmt.Errorf("failed to execute query: %w", err)
	}

	if len(results) == 0 {
		logger.Error("Group not found with id: " + id)
		return model.Group{}, model.ErrGroupNotFound
	}

	if len(results) != 1 {
		logger.Error("Unexpected number of results")
		return model.Group{}, fmt.Errorf("unexpected number of results: %d", len(results))
	}

	row := results[0]
	group, err := buildGroupFromResultRow(row, logger)
	if err != nil {
		return model.Group{}, err
	}

	// Get child groups
	childGroups, err := GetChildGroups(id)
	if err != nil {
		return model.Group{}, err
	}
	group.Groups = *childGroups

	// Get users
	users, err := getGroupUsers(dbClient, id, logger)
	if err != nil {
		return model.Group{}, err
	}
	group.Users = users

	return group, nil
}

// UpdateGroup updates an existing group.
func UpdateGroup(group model.Group) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	dbClient, err := provider.NewDBProvider().GetDBClient("identity")
	if err != nil {
		logger.Error("Failed to get database client", log.Error(err))
		return fmt.Errorf("failed to get database client: %w", err)
	}
	defer func() {
		if closeErr := dbClient.Close(); closeErr != nil {
			logger.Error("Failed to close database client", log.Error(closeErr))
		}
	}()

	// Check for name conflicts (excluding current group)
	err = checkGroupNameConflictForUpdate(dbClient, group.Name, group.Parent, group.ID, logger)
	if err != nil {
		return err
	}

	// Determine the parent group ID and OU ID
	var parentGroupID *string
	var ouID string

	if group.Parent.Type == GroupType {
		parentGroupID = &group.Parent.ID
		// Get the OU id from the parent group
		parentGroup, err := GetGroup(group.Parent.ID)
		if err != nil {
			logger.Error("Failed to get parent group", log.Error(err))
			return model.ErrParentNotFound
		}
		// Convert Group to GroupBasic for getOUFromPath function
		parentGroupBasic := model.GroupBasic{
			ID:          parentGroup.ID,
			Name:        parentGroup.Name,
			Description: parentGroup.Description,
			Parent:      parentGroup.Parent,
		}
		ouID = getOUFromPath(parentGroupBasic)
	} else {
		ouID = group.Parent.ID
	}

	// Generate path
	path := generateGroupPath(group.Name, group.Parent)

	rowsAffected, err := dbClient.Execute(
		QueryUpdateGroup,
		group.ID,
		parentGroupID,
		ouID,
		group.Name,
		group.Description,
		path,
	)
	if err != nil {
		logger.Error("Failed to execute update group query", log.Error(err))
		return fmt.Errorf("failed to execute query: %w", err)
	}

	if rowsAffected == 0 {
		logger.Error("Group not found with id: " + group.ID)
		return model.ErrGroupNotFound
	}

	// Update group users
	err = updateGroupUsers(dbClient, group.ID, group.Users, logger)
	if err != nil {
		return err
	}

	return nil
}

// DeleteGroup deletes a group.
func DeleteGroup(id string) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	dbClient, err := provider.NewDBProvider().GetDBClient("identity")
	if err != nil {
		logger.Error("Failed to get database client", log.Error(err))
		return fmt.Errorf("failed to get database client: %w", err)
	}
	defer func() {
		if closeErr := dbClient.Close(); closeErr != nil {
			logger.Error("Failed to close database client", log.Error(closeErr))
		}
	}()

	// Delete group users first
	_, err = dbClient.Execute(QueryDeleteGroupUsers, id)
	if err != nil {
		logger.Error("Failed to delete group users", log.Error(err))
		return fmt.Errorf("failed to delete group users: %w", err)
	}

	// Delete the group
	rowsAffected, err := dbClient.Execute(QueryDeleteGroup, id)
	if err != nil {
		logger.Error("Failed to execute delete group query", log.Error(err))
		return fmt.Errorf("failed to execute query: %w", err)
	}

	if rowsAffected == 0 {
		logger.Debug("Group not found with id: " + id)
	}

	return nil
}

// buildGroupFromResultRow constructs a model.Group from a database result row.
func buildGroupFromResultRow(row map[string]interface{}, logger *log.Logger) (model.Group, error) {
	groupID, ok := row["group_id"].(string)
	if !ok {
		logger.Error("Failed to parse group_id as string")
		return model.Group{}, fmt.Errorf("failed to parse group_id as string")
	}

	name, ok := row["name"].(string)
	if !ok {
		logger.Error("Failed to parse name as string")
		return model.Group{}, fmt.Errorf("failed to parse name as string")
	}

	description, ok := row["description"].(string)
	if !ok {
		logger.Error("Failed to parse description as string")
		return model.Group{}, fmt.Errorf("failed to parse description as string")
	}

	ouID, ok := row["ou_id"].(string)
	if !ok {
		logger.Error("Failed to parse ou_id as string")
		return model.Group{}, fmt.Errorf("failed to parse ou_id as string")
	}

	var parentGroupID *string
	if row["parent_id"] != nil {
		if pgid, ok := row["parent_id"].(string); ok {
			parentGroupID = &pgid
		}
	}

	// Determine parent
	var parent model.Parent
	if parentGroupID != nil {
		parent = model.Parent{
			Type: model.ParentTypeGroup,
			ID:   *parentGroupID,
		}
	} else {
		parent = model.Parent{
			Type: model.ParentTypeOrganizationUnit,
			ID:   ouID,
		}
	}

	group := model.Group{
		ID:          groupID,
		Name:        name,
		Description: description,
		Parent:      parent,
	}

	return group, nil
}

// GetChildGroups retrieves the child groups of a given group ID.
func GetChildGroups(groupID string) (*[]string, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	dbClient, err := provider.NewDBProvider().GetDBClient("identity")
	if err != nil {
		logger.Error("Failed to get database client", log.Error(err))
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}
	defer func() {
		if closeErr := dbClient.Close(); closeErr != nil {
			logger.Error("Failed to close database client", log.Error(closeErr))
		}
	}()

	results, err := dbClient.Query(QueryGetChildGroups, groupID)
	if err != nil {
		logger.Error("Failed to get child groups", log.Error(err))
		return nil, fmt.Errorf("failed to get child groups: %w", err)
	}

	childGroups := make([]string, 0)
	for _, row := range results {
		if childID, ok := row["group_id"].(string); ok {
			childGroups = append(childGroups, childID)
		}
	}

	return &childGroups, nil
}

// getChildGroups retrieves the child groups of a given group ID.
func getGroupUsers(dbClient client.DBClientInterface, groupID string, logger *log.Logger) ([]string, error) {
	results, err := dbClient.Query(QueryGetGroupUsers, groupID)
	if err != nil {
		logger.Error("Failed to get group users", log.Error(err))
		return nil, fmt.Errorf("failed to get group users: %w", err)
	}

	users := make([]string, 0)
	for _, row := range results {
		if userID, ok := row["user_id"].(string); ok {
			users = append(users, userID)
		}
	}

	return users, nil
}

// addUsersToGroup adds a list of users to a group.
func addUsersToGroup(dbClient client.DBClientInterface, groupID string, users []string, logger *log.Logger) error {
	for _, userID := range users {
		_, err := dbClient.Execute(QueryAddUserToGroup, groupID, userID)
		if err != nil {
			logger.Error("Failed to add user to group", log.String("userID", userID), log.Error(err))
			return fmt.Errorf("failed to add user to group: %w", err)
		}
	}
	return nil
}

// updateGroupUsers updates the users in a group by first deleting existing users and then adding new ones.
func updateGroupUsers(dbClient client.DBClientInterface, groupID string, users []string, logger *log.Logger) error {
	// Delete existing users
	_, err := dbClient.Execute(QueryDeleteGroupUsers, groupID)
	if err != nil {
		logger.Error("Failed to delete existing group users", log.Error(err))
		return fmt.Errorf("failed to delete existing group users: %w", err)
	}

	// Add new users
	return addUsersToGroup(dbClient, groupID, users, logger)
}

// checkGroupNameConflictForCreate checks if the new group name conflicts with existing groups under the same parent.
func checkGroupNameConflictForCreate(
	dbClient client.DBClientInterface,
	name string,
	parent model.Parent,
	logger *log.Logger,
) error {
	var results []map[string]interface{}
	var err error

	// Use appropriate query based on parent type and whether this is an update operation
	if parent.Type == model.ParentTypeGroup {
		results, err = dbClient.Query(QueryCheckGroupNameConflict, name, parent.ID)
	} else if parent.Type == model.ParentTypeOrganizationUnit {
		results, err = dbClient.Query(QueryCheckGroupNameConflictInOU, name, parent.ID)
	} else {
		return fmt.Errorf("invalid parent type: %s", parent.Type)
	}

	if err != nil {
		logger.Error("Failed to check group name conflict", log.Error(err))
		return fmt.Errorf("failed to check group name conflict: %w", err)
	}

	if len(results) > 0 {
		if count, ok := results[0]["count"].(int64); ok && count > 0 {
			return model.ErrGroupNameConflict
		}
	}

	return nil
}

// checkGroupNameConflictForUpdate checks if the new group name conflicts with other groups under the same parent.
func checkGroupNameConflictForUpdate(
	dbClient client.DBClientInterface,
	name string,
	parent model.Parent,
	groupID string,
	logger *log.Logger,
) error {
	var results []map[string]interface{}
	var err error

	// Use appropriate query based on parent type and whether this is an update operation
	if parent.Type == model.ParentTypeGroup {
		results, err = dbClient.Query(QueryCheckGroupNameConflictForUpdate, name, parent.ID, groupID)
	} else if parent.Type == model.ParentTypeOrganizationUnit {
		results, err = dbClient.Query(QueryCheckGroupNameConflictInOUForUpdate, name, parent.ID, groupID)
	} else {
		return fmt.Errorf("invalid parent type: %s", parent.Type)
	}

	if err != nil {
		logger.Error("Failed to check group name conflict", log.Error(err))
		return fmt.Errorf("failed to check group name conflict: %w", err)
	}

	if len(results) > 0 {
		if count, ok := results[0]["count"].(int64); ok && count > 0 {
			return model.ErrGroupNameConflict
		}
	}

	return nil
}

// generateGroupPath generates the path for a group based on its name and parent.
func generateGroupPath(name string, parent model.Parent) string {
	// Simplified path generation - in a real implementation, you'd build the full path
	// from the root to this group
	if parent.Type == "group" {
		return fmt.Sprintf("/%s/%s", parent.ID, name)
	}
	return fmt.Sprintf("/%s", name)
}

// getOUFromPath extracts the OU ID from a group's path.
func getOUFromPath(group model.GroupBasic) string {
	// Simplified - in a real implementation, you'd extract the OU from the group's path
	// For now, return the parent OU ID
	if group.Parent.Type == "organizationUnit" {
		return group.Parent.ID
	}
	// Would need to traverse up the hierarchy to find the root OU
	return group.Parent.ID
}
