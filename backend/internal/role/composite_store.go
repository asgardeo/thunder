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

	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
)

// compositeRoleStore implements a composite store that combines file-based (immutable) and
// database (mutable) stores.
// - Read operations query both stores and merge results
// - Write operations (Create/Update/Delete) only affect the database store
// - Declarative roles (from YAML files) cannot be modified or deleted
type compositeRoleStore struct {
	fileStore roleStoreInterface
	dbStore   roleStoreInterface
}

// newCompositeRoleStore creates a new composite store with both file-based and database stores.
func newCompositeRoleStore(fileStore, dbStore roleStoreInterface) roleStoreInterface {
	return &compositeRoleStore{
		fileStore: fileStore,
		dbStore:   dbStore,
	}
}

// GetRoleListCount retrieves the total count of roles from both stores.
func (c *compositeRoleStore) GetRoleListCount(ctx context.Context) (int, error) {
	dbCount, err := c.dbStore.GetRoleListCount(ctx)
	if err != nil {
		return 0, err
	}

	fileCount, err := c.fileStore.GetRoleListCount(ctx)
	if err != nil {
		return 0, err
	}

	dbRoles, err := c.dbStore.GetRoleList(ctx, dbCount, 0)
	if err != nil {
		return 0, err
	}

	fileRoles, err := c.fileStore.GetRoleList(ctx, fileCount, 0)
	if err != nil {
		return 0, err
	}

	mergedRoles := mergeRoles(dbRoles, fileRoles)
	return len(mergedRoles), nil
}

// GetRoleList retrieves roles from both stores and merges them.
func (c *compositeRoleStore) GetRoleList(ctx context.Context, limit, offset int) ([]Role, error) {
	dbCount, err := c.dbStore.GetRoleListCount(ctx)
	if err != nil {
		return nil, err
	}

	fileCount, err := c.fileStore.GetRoleListCount(ctx)
	if err != nil {
		return nil, err
	}

	dbRoles, err := c.dbStore.GetRoleList(ctx, dbCount, 0)
	if err != nil {
		return nil, err
	}

	fileRoles, err := c.fileStore.GetRoleList(ctx, fileCount, 0)
	if err != nil {
		return nil, err
	}

	// Merge roles from both stores
	mergedRoles := mergeRoles(dbRoles, fileRoles)

	start := offset
	if start > len(mergedRoles) {
		start = len(mergedRoles)
	}
	end := start + limit
	if end > len(mergedRoles) {
		end = len(mergedRoles)
	}

	return mergedRoles[start:end], nil
}

// CreateRole creates a new role in the database store only.
func (c *compositeRoleStore) CreateRole(ctx context.Context, id string, role RoleCreationDetail) error {
	return c.dbStore.CreateRole(ctx, id, role)
}

// GetRole retrieves a role from either store.
// Checks database store first, then falls back to file store.
func (c *compositeRoleStore) GetRole(ctx context.Context, id string) (RoleWithPermissions, error) {
	role, err := declarativeresource.CompositeGetHelper(
		func() (*RoleWithPermissions, error) {
			r, err := c.dbStore.GetRole(ctx, id)
			return &r, err
		},
		func() (*RoleWithPermissions, error) {
			r, err := c.fileStore.GetRole(ctx, id)
			return &r, err
		},
		ErrRoleNotFound,
	)
	if role != nil {
		return *role, err
	}
	return RoleWithPermissions{}, err
}

// IsRoleExist checks if a role exists in either store.
func (c *compositeRoleStore) IsRoleExist(ctx context.Context, id string) (bool, error) {
	return declarativeresource.CompositeBooleanCheckHelper(
		func() (bool, error) { return c.fileStore.IsRoleExist(ctx, id) },
		func() (bool, error) { return c.dbStore.IsRoleExist(ctx, id) },
	)
}

// GetRoleAssignments retrieves role assignments from both stores.
func (c *compositeRoleStore) GetRoleAssignments(
	ctx context.Context,
	id string,
	limit, offset int,
) ([]RoleAssignment, error) {
	dbCount, err := c.dbStore.GetRoleAssignmentsCount(ctx, id)
	if err != nil {
		return nil, err
	}

	fileCount, err := c.fileStore.GetRoleAssignmentsCount(ctx, id)
	if err != nil {
		return nil, err
	}

	dbAssignments, err := c.dbStore.GetRoleAssignments(ctx, id, dbCount, 0)
	if err != nil {
		return nil, err
	}

	fileAssignments, err := c.fileStore.GetRoleAssignments(ctx, id, fileCount, 0)
	if err != nil {
		return nil, err
	}

	// Merge assignments from both stores
	mergedAssignments := mergeAssignments(dbAssignments, fileAssignments)

	start := offset
	if start > len(mergedAssignments) {
		start = len(mergedAssignments)
	}
	end := start + limit
	if end > len(mergedAssignments) {
		end = len(mergedAssignments)
	}

	return mergedAssignments[start:end], nil
}

// GetRoleAssignmentsCount retrieves the count of role assignments from both stores.
func (c *compositeRoleStore) GetRoleAssignmentsCount(ctx context.Context, id string) (int, error) {
	dbCount, err := c.dbStore.GetRoleAssignmentsCount(ctx, id)
	if err != nil {
		return 0, err
	}

	fileCount, err := c.fileStore.GetRoleAssignmentsCount(ctx, id)
	if err != nil {
		return 0, err
	}

	dbAssignments, err := c.dbStore.GetRoleAssignments(ctx, id, dbCount, 0)
	if err != nil {
		return 0, err
	}

	fileAssignments, err := c.fileStore.GetRoleAssignments(ctx, id, fileCount, 0)
	if err != nil {
		return 0, err
	}

	mergedAssignments := mergeAssignments(dbAssignments, fileAssignments)
	return len(mergedAssignments), nil
}

// UpdateRole updates a role in the database store only.
// Immutability checks are handled at the service layer.
func (c *compositeRoleStore) UpdateRole(ctx context.Context, id string, role RoleUpdateDetail) error {
	return c.dbStore.UpdateRole(ctx, id, role)
}

// DeleteRole deletes a role from the database store only.
// Immutability checks are handled at the service layer.
func (c *compositeRoleStore) DeleteRole(ctx context.Context, id string) error {
	return c.dbStore.DeleteRole(ctx, id)
}

// AddAssignments adds assignments to a role in the database store only.
func (c *compositeRoleStore) AddAssignments(ctx context.Context, id string, assignments []RoleAssignment) error {
	return c.dbStore.AddAssignments(ctx, id, assignments)
}

// RemoveAssignments removes assignments from a role in the database store only.
func (c *compositeRoleStore) RemoveAssignments(ctx context.Context, id string, assignments []RoleAssignment) error {
	return c.dbStore.RemoveAssignments(ctx, id, assignments)
}

// CheckRoleNameExists checks if a role with the given name exists in either store.
func (c *compositeRoleStore) CheckRoleNameExists(ctx context.Context, ouID, name string) (bool, error) {
	return declarativeresource.CompositeBooleanCheckHelper(
		func() (bool, error) { return c.fileStore.CheckRoleNameExists(ctx, ouID, name) },
		func() (bool, error) { return c.dbStore.CheckRoleNameExists(ctx, ouID, name) },
	)
}

// CheckRoleNameExistsExcludingID checks if a role exists excluding a specific ID in either store.
func (c *compositeRoleStore) CheckRoleNameExistsExcludingID(
	ctx context.Context,
	ouID, name, excludeRoleID string,
) (bool, error) {
	return declarativeresource.CompositeBooleanCheckHelper(
		func() (bool, error) {
			return c.fileStore.CheckRoleNameExistsExcludingID(ctx, ouID, name, excludeRoleID)
		},
		func() (bool, error) { return c.dbStore.CheckRoleNameExistsExcludingID(ctx, ouID, name, excludeRoleID) },
	)
}

// GetAuthorizedPermissions retrieves authorized permissions from both stores.
func (c *compositeRoleStore) GetAuthorizedPermissions(
	ctx context.Context,
	userID string,
	groupIDs []string,
	requestPermissions []string,
) ([]string, error) {
	dbPerms, err := c.dbStore.GetAuthorizedPermissions(ctx, userID, groupIDs, requestPermissions)
	if err != nil {
		return nil, err
	}

	filePerms, err := c.fileStore.GetAuthorizedPermissions(ctx, userID, groupIDs, requestPermissions)
	if err != nil {
		return nil, err
	}

	// Merge permissions from both stores and deduplicate
	return mergePermissions(dbPerms, filePerms), nil
}

// IsRoleDeclarative checks if a role is immutable (exists in file store).
func (c *compositeRoleStore) IsRoleDeclarative(ctx context.Context, roleID string) (bool, error) {
	fileExists, err := c.fileStore.IsRoleExist(ctx, roleID)
	if err != nil {
		return false, err
	}
	return fileExists, nil
}

// mergeRoles deduplicates and merges roles from database and file stores.
// Database roles take precedence over file-based roles with the same ID.
func mergeRoles(dbRoles, fileRoles []Role) []Role {
	seen := make(map[string]bool)
	result := make([]Role, 0, len(dbRoles)+len(fileRoles))

	// Add database roles first (they take precedence)
	for _, role := range dbRoles {
		result = append(result, role)
		seen[role.ID] = true
	}

	// Add file-based roles only if not already seen
	for _, role := range fileRoles {
		if !seen[role.ID] {
			result = append(result, role)
			seen[role.ID] = true
		}
	}

	return result
}

// mergeAssignments deduplicates and merges assignments from database and file stores.
// Database assignments take precedence over file-based assignments with the same ID and type.
func mergeAssignments(dbAssignments, fileAssignments []RoleAssignment) []RoleAssignment {
	seen := make(map[string]bool)
	result := make([]RoleAssignment, 0, len(dbAssignments)+len(fileAssignments))

	// Add database assignments first (they take precedence)
	for _, assignment := range dbAssignments {
		key := string(assignment.Type) + ":" + assignment.ID
		result = append(result, assignment)
		seen[key] = true
	}

	// Add file-based assignments only if not already seen
	for _, assignment := range fileAssignments {
		key := string(assignment.Type) + ":" + assignment.ID
		if !seen[key] {
			result = append(result, assignment)
			seen[key] = true
		}
	}

	return result
}

// mergePermissions deduplicates and merges permissions from database and file stores.
func mergePermissions(dbPerms, filePerms []string) []string {
	permMap := make(map[string]bool)

	for _, perm := range filePerms {
		permMap[perm] = true
	}

	for _, perm := range dbPerms {
		permMap[perm] = true
	}

	result := make([]string, 0, len(permMap))
	for perm := range permMap {
		result = append(result, perm)
	}
	return result
}
