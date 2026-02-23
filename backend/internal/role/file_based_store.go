/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

package role

import (
	"context"
	"errors"
	"strings"

	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/declarative_resource/entity"
	"github.com/asgardeo/thunder/internal/system/log"
)

type fileBasedStore struct {
	*declarativeresource.GenericFileBasedStore
}

// newFileBasedStore creates a new file-based store for roles.
func newFileBasedStore() roleStoreInterface {
	return &fileBasedStore{
		GenericFileBasedStore: declarativeresource.NewGenericFileBasedStore(entity.KeyTypeRole),
	}
}

// Create implements declarativeresource.Storer interface for resource loader.
func (f *fileBasedStore) Create(id string, data interface{}) error {
	role, ok := data.(*RoleWithPermissionsAndAssignments)
	if !ok {
		return errors.New("role data corrupted")
	}
	if role.ID == "" {
		role.ID = id
	}
	return f.GenericFileBasedStore.Create(id, role)
}

// GetRoleListCount returns the total count of roles in the file-based store.
func (f *fileBasedStore) GetRoleListCount(ctx context.Context) (int, error) {
	return f.GenericFileBasedStore.Count()
}

// GetRoleList returns the list of roles from the file-based store.
func (f *fileBasedStore) GetRoleList(ctx context.Context, limit, offset int) ([]Role, error) {
	if limit <= 0 {
		return []Role{}, nil
	}
	if offset < 0 {
		offset = 0
	}

	list, err := f.GenericFileBasedStore.List()
	if err != nil {
		return nil, err
	}

	roles := make([]Role, 0, len(list))
	for _, item := range list {
		roleData, err := roleFromDeclarativeData(item.ID.ID, item.Data)
		if err != nil {
			// Log warning for malformed declarative entry
			log.GetLogger().Warn("Skipping malformed role in GetRoleList",
				log.String("roleID", item.ID.ID),
				log.Error(err))
			continue
		}
		roles = append(roles, Role{
			ID:                 roleData.ID,
			Name:               roleData.Name,
			Description:        roleData.Description,
			OrganizationUnitID: roleData.OrganizationUnitID,
		})
	}

	start := offset
	if start >= len(roles) {
		return []Role{}, nil
	}
	end := start + limit
	if end > len(roles) {
		end = len(roles)
	}

	return roles[start:end], nil
}

// CreateRole is not supported in file-based store.
func (f *fileBasedStore) CreateRole(ctx context.Context, id string, role RoleCreationDetail) error {
	return errors.New("CreateRole is not supported in file-based store")
}

// GetRole returns a role from the file-based store.
func (f *fileBasedStore) GetRole(ctx context.Context, id string) (RoleWithPermissions, error) {
	data, err := f.GenericFileBasedStore.Get(id)
	if err != nil {
		// Distinguish "not found" from other storage errors
		if isEntityNotFoundError(err) {
			return RoleWithPermissions{}, ErrRoleNotFound
		}
		// Propagate other storage errors
		return RoleWithPermissions{}, err
	}

	roleData, err := roleFromDeclarativeData(id, data)
	if err != nil {
		// Propagate parsing errors
		return RoleWithPermissions{}, err
	}

	return RoleWithPermissions{
		ID:                 roleData.ID,
		Name:               roleData.Name,
		Description:        roleData.Description,
		OrganizationUnitID: roleData.OrganizationUnitID,
		Permissions:        roleData.Permissions,
	}, nil
}

// IsRoleExist checks if a role exists in the file-based store.
func (f *fileBasedStore) IsRoleExist(ctx context.Context, id string) (bool, error) {
	_, err := f.GenericFileBasedStore.Get(id)
	if err != nil {
		// Distinguish "not found" from other storage errors
		if isEntityNotFoundError(err) {
			return false, nil
		}
		// Propagate other storage errors
		return false, err
	}
	return true, nil
}

// GetRoleAssignments returns role assignments from the file-based store.
func (f *fileBasedStore) GetRoleAssignments(
	ctx context.Context,
	id string,
	limit, offset int,
) ([]RoleAssignment, error) {
	if limit <= 0 {
		return []RoleAssignment{}, nil
	}
	if offset < 0 {
		offset = 0
	}

	data, err := f.GenericFileBasedStore.Get(id)
	if err != nil {
		// Distinguish "not found" from other storage errors
		if isEntityNotFoundError(err) {
			return []RoleAssignment{}, nil
		}
		// Propagate other storage errors
		return nil, err
	}

	roleData, err := roleFromDeclarativeData(id, data)
	if err != nil {
		// Propagate parsing errors
		return nil, err
	}

	assignments := roleData.Assignments
	start := offset
	if start >= len(assignments) {
		return []RoleAssignment{}, nil
	}
	end := start + limit
	if end > len(assignments) {
		end = len(assignments)
	}

	return assignments[start:end], nil
}

// GetRoleAssignmentsCount returns the assignment count for a role in the file-based store.
func (f *fileBasedStore) GetRoleAssignmentsCount(ctx context.Context, id string) (int, error) {
	data, err := f.GenericFileBasedStore.Get(id)
	if err != nil {
		// Distinguish "not found" from other storage errors
		if isEntityNotFoundError(err) {
			return 0, nil
		}
		// Propagate other storage errors
		return 0, err
	}

	roleData, err := roleFromDeclarativeData(id, data)
	if err != nil {
		// Propagate parsing errors
		return 0, err
	}

	return len(roleData.Assignments), nil
}

// UpdateRole is not supported in file-based store.
func (f *fileBasedStore) UpdateRole(ctx context.Context, id string, role RoleUpdateDetail) error {
	return errors.New("UpdateRole is not supported in file-based store")
}

// DeleteRole is not supported in file-based store.
func (f *fileBasedStore) DeleteRole(ctx context.Context, id string) error {
	return errors.New("DeleteRole is not supported in file-based store")
}

// AddAssignments is not supported in file-based store.
func (f *fileBasedStore) AddAssignments(ctx context.Context, id string, assignments []RoleAssignment) error {
	return errors.New("AddAssignments is not supported in file-based store")
}

// RemoveAssignments is not supported in file-based store.
func (f *fileBasedStore) RemoveAssignments(ctx context.Context, id string, assignments []RoleAssignment) error {
	return errors.New("RemoveAssignments is not supported in file-based store")
}

// CheckRoleNameExists checks if a role with the given name exists in the file-based store.
func (f *fileBasedStore) CheckRoleNameExists(ctx context.Context, ouID, name string) (bool, error) {
	list, err := f.GenericFileBasedStore.List()
	if err != nil {
		return false, err
	}

	for _, item := range list {
		roleData, err := roleFromDeclarativeData(item.ID.ID, item.Data)
		if err != nil {
			// Log warning for malformed declarative entry
			log.GetLogger().Warn("Skipping malformed role in CheckRoleNameExists",
				log.String("roleID", item.ID.ID),
				log.Error(err))
			continue
		}
		if roleData.OrganizationUnitID == ouID && roleData.Name == name {
			return true, nil
		}
	}

	return false, nil
}

// CheckRoleNameExistsExcludingID checks for a role name conflict excluding a specific role ID.
func (f *fileBasedStore) CheckRoleNameExistsExcludingID(
	ctx context.Context,
	ouID, name, excludeRoleID string,
) (bool, error) {
	list, err := f.GenericFileBasedStore.List()
	if err != nil {
		return false, err
	}

	for _, item := range list {
		roleData, err := roleFromDeclarativeData(item.ID.ID, item.Data)
		if err != nil {
			// Log warning for malformed declarative entry
			log.GetLogger().Warn("Skipping malformed role in CheckRoleNameExistsExcludingID",
				log.String("roleID", item.ID.ID),
				log.Error(err))
			continue
		}
		if roleData.ID == excludeRoleID {
			continue
		}
		if roleData.OrganizationUnitID == ouID && roleData.Name == name {
			return true, nil
		}
	}

	return false, nil
}

// GetAuthorizedPermissions returns permissions from roles assigned to the user or groups in the file store.
func (f *fileBasedStore) GetAuthorizedPermissions(
	ctx context.Context,
	userID string,
	groupIDs []string,
	requestPermissions []string,
) ([]string, error) {
	if len(requestPermissions) == 0 {
		return []string{}, nil
	}
	if userID == "" && len(groupIDs) == 0 {
		return []string{}, nil
	}

	list, err := f.GenericFileBasedStore.List()
	if err != nil {
		return nil, err
	}

	requestedSet := make(map[string]bool, len(requestPermissions))
	for _, perm := range requestPermissions {
		requestedSet[perm] = true
	}

	groupSet := make(map[string]bool, len(groupIDs))
	for _, groupID := range groupIDs {
		groupSet[groupID] = true
	}

	permitted := make(map[string]bool)
	for _, item := range list {
		roleData, err := roleFromDeclarativeData(item.ID.ID, item.Data)
		if err != nil {
			// Log warning for malformed declarative entry
			log.GetLogger().Warn("Skipping malformed role in GetAuthorizedPermissions",
				log.String("roleID", item.ID.ID),
				log.Error(err))
			continue
		}
		if !matchesAssignee(roleData.Assignments, userID, groupSet) {
			continue
		}
		for _, resourcePerms := range roleData.Permissions {
			for _, perm := range resourcePerms.Permissions {
				if requestedSet[perm] {
					permitted[perm] = true
				}
			}
		}
	}

	result := make([]string, 0, len(permitted))
	for _, perm := range requestPermissions {
		if permitted[perm] {
			result = append(result, perm)
		}
	}
	return result, nil
}

// IsRoleDeclarative returns true for roles in the file-based store because they are declarative.
func (f *fileBasedStore) IsRoleDeclarative(ctx context.Context, roleID string) (bool, error) {
	exists, err := f.IsRoleExist(ctx, roleID)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// roleFromDeclarativeData converts raw data from the declarative store into a RoleWithPermissionsAndAssignments struct.
func roleFromDeclarativeData(id string, data interface{}) (RoleWithPermissionsAndAssignments, error) {
	role, ok := data.(*RoleWithPermissionsAndAssignments)
	if !ok || role == nil {
		declarativeresource.LogTypeAssertionError("role", id)
		return RoleWithPermissionsAndAssignments{}, errors.New("role data corrupted")
	}

	return *role, nil
}

// isEntityNotFoundError checks if an error indicates an entity was not found.
// This function checks for error messages from both GenericFileBasedStore
// and the underlying entity store.
func isEntityNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return errMsg == "entity not found" || strings.Contains(errMsg, "not found")
}

// matchesAssignee returns true when the user or any of the user's groups is assigned.
func matchesAssignee(assignments []RoleAssignment, userID string, groupSet map[string]bool) bool {
	for _, assignment := range assignments {
		if assignment.Type == AssigneeTypeUser && assignment.ID == userID {
			return true
		}
		if assignment.Type == AssigneeTypeGroup && groupSet[assignment.ID] {
			return true
		}
	}
	return false
}
