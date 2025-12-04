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

package ou

import (
	"errors"
	"fmt"

	"github.com/asgardeo/thunder/internal/system/log"
)

const hybridStoreLoggerComponentName = "OrganizationUnitHybridStore"

// ouHybridStore implements organizationUnitStoreInterface by combining file-based and database stores.
// Read operations merge results from both stores, write operations only go to the database store.
type ouHybridStore struct {
	fileStore organizationUnitStoreInterface // Immutable OUs from YAML files
	dbStore   organizationUnitStoreInterface // Runtime OUs from database
}

// newOUHybridStore creates a new hybrid store instance.
func newOUHybridStore(
	fileStore organizationUnitStoreInterface,
	dbStore organizationUnitStoreInterface,
) organizationUnitStoreInterface {
	return &ouHybridStore{
		fileStore: fileStore,
		dbStore:   dbStore,
	}
}

// GetOrganizationUnitListCount retrieves the total count of root organization units from both stores.
func (h *ouHybridStore) GetOrganizationUnitListCount() (int, error) {
	fileCount, err := h.fileStore.GetOrganizationUnitListCount()
	if err != nil {
		return 0, fmt.Errorf("failed to get file store count: %w", err)
	}

	_, err = h.dbStore.GetOrganizationUnitListCount()
	if err != nil {
		return 0, fmt.Errorf("failed to get database store count: %w", err)
	}

	// Get all OUs to check for duplicates
	fileOUs, err := h.fileStore.GetOrganizationUnitList(9999, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to get file store list: %w", err)
	}

	dbOUs, err := h.dbStore.GetOrganizationUnitList(9999, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to get database store list: %w", err)
	}

	// Create set of file-based IDs
	fileIDs := make(map[string]bool)
	for _, ou := range fileOUs {
		fileIDs[ou.ID] = true
	}

	// Count unique DB OUs (exclude duplicates)
	uniqueDBCount := 0
	for _, ou := range dbOUs {
		if !fileIDs[ou.ID] {
			uniqueDBCount++
		}
	}

	return fileCount + uniqueDBCount, nil
}

// GetOrganizationUnitList retrieves root organization units from both stores with pagination.
// File-based OUs are returned first, followed by database OUs.
func (h *ouHybridStore) GetOrganizationUnitList(limit, offset int) ([]OrganizationUnitBasic, error) {
	// Get all from both stores
	fileOUs, err := h.fileStore.GetOrganizationUnitList(9999, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get file store list: %w", err)
	}

	dbOUs, err := h.dbStore.GetOrganizationUnitList(9999, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get database store list: %w", err)
	}

	// Merge: file-based first, then DB (avoid duplicates by ID)
	seen := make(map[string]bool)
	var merged []OrganizationUnitBasic

	for _, ou := range fileOUs {
		merged = append(merged, ou)
		seen[ou.ID] = true
	}

	for _, ou := range dbOUs {
		if !seen[ou.ID] {
			merged = append(merged, ou)
		}
	}

	// Apply pagination to merged results
	start := offset
	if start > len(merged) {
		return []OrganizationUnitBasic{}, nil
	}

	end := start + limit
	if end > len(merged) {
		end = len(merged)
	}

	return merged[start:end], nil
}

// CreateOrganizationUnit creates a new organization unit in the database store only.
// It checks for conflicts with file-based OUs before creating.
func (h *ouHybridStore) CreateOrganizationUnit(ou OrganizationUnit) error {
	// Check if ID exists in file store (immutable)
	_, err := h.fileStore.GetOrganizationUnit(ou.ID)
	if err == nil {
		return errors.New("cannot create organization unit: ID conflicts with immutable file-based OU")
	}
	if err != ErrOrganizationUnitNotFound {
		return fmt.Errorf("error checking file store: %w", err)
	}

	// Check handle conflict in file store
	hasConflict, err := h.fileStore.CheckOrganizationUnitHandleConflict(ou.Handle, ou.Parent)
	if err != nil {
		return fmt.Errorf("error checking handle conflict in file store: %w", err)
	}
	if hasConflict {
		return errors.New("cannot create organization unit: handle conflicts with immutable file-based OU")
	}

	// Check name conflict in file store
	hasConflict, err = h.fileStore.CheckOrganizationUnitNameConflict(ou.Name, ou.Parent)
	if err != nil {
		return fmt.Errorf("error checking name conflict in file store: %w", err)
	}
	if hasConflict {
		return errors.New("cannot create organization unit: name conflicts with immutable file-based OU")
	}

	// Create in DB store only
	return h.dbStore.CreateOrganizationUnit(ou)
}

// GetOrganizationUnit retrieves an organization unit by ID from either store.
// File store is checked first, then database store.
func (h *ouHybridStore) GetOrganizationUnit(id string) (OrganizationUnit, error) {
	// Try file store first
	ou, err := h.fileStore.GetOrganizationUnit(id)
	if err == nil {
		return ou, nil
	}
	if err != ErrOrganizationUnitNotFound {
		return OrganizationUnit{}, fmt.Errorf("error checking file store: %w", err)
	}

	// Try DB store
	return h.dbStore.GetOrganizationUnit(id)
}

// GetOrganizationUnitByPath retrieves an organization unit by its hierarchical path.
// Supports mixed paths where parents can be in file store and children in DB store or vice versa.
func (h *ouHybridStore) GetOrganizationUnitByPath(handlePath []string) (OrganizationUnit, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, hybridStoreLoggerComponentName))

	if len(handlePath) == 0 {
		return OrganizationUnit{}, ErrOrganizationUnitNotFound
	}

	// Try file store first (might be faster for full paths)
	ou, err := h.fileStore.GetOrganizationUnitByPath(handlePath)
	if err == nil {
		return ou, nil
	}

	// Try DB store
	ou, err = h.dbStore.GetOrganizationUnitByPath(handlePath)
	if err == nil {
		return ou, nil
	}

	// Handle mixed path (traverse step by step across both stores)
	logger.Debug("Attempting mixed path traversal", log.Any("handlePath", handlePath))
	return h.traverseMixedPath(handlePath)
}

// traverseMixedPath traverses a hierarchical path across both file and database stores.
func (h *ouHybridStore) traverseMixedPath(handlePath []string) (OrganizationUnit, error) {
	var currentParentID *string

	for i, handle := range handlePath {
		// Try to find in file store
		ou, err := h.findOUByHandleAndParent(handle, currentParentID, h.fileStore)
		if err == nil {
			currentParentID = &ou.ID
			if i == len(handlePath)-1 {
				return ou, nil
			}
			continue
		}

		// Try DB store
		ou, err = h.findOUByHandleAndParent(handle, currentParentID, h.dbStore)
		if err != nil {
			return OrganizationUnit{}, ErrOrganizationUnitNotFound
		}

		currentParentID = &ou.ID
		if i == len(handlePath)-1 {
			return ou, nil
		}
	}

	return OrganizationUnit{}, ErrOrganizationUnitNotFound
}

// findOUByHandleAndParent finds an OU by handle and parent ID in a specific store.
func (h *ouHybridStore) findOUByHandleAndParent(
	handle string,
	parentID *string,
	store organizationUnitStoreInterface,
) (OrganizationUnit, error) {
	// Get all children of parent (or root if parent is nil)
	children, err := store.GetOrganizationUnitChildrenList("", 9999, 0)
	if err != nil {
		return OrganizationUnit{}, err
	}

	// If we have a parent, get its children
	if parentID != nil {
		children, err = store.GetOrganizationUnitChildrenList(*parentID, 9999, 0)
		if err != nil {
			return OrganizationUnit{}, err
		}
	}

	// Find matching handle
	for _, child := range children {
		if child.Handle == handle {
			return store.GetOrganizationUnit(child.ID)
		}
	}

	return OrganizationUnit{}, ErrOrganizationUnitNotFound
}

// IsOrganizationUnitExists checks if an organization unit exists in either store.
func (h *ouHybridStore) IsOrganizationUnitExists(id string) (bool, error) {
	// Check file store
	exists, err := h.fileStore.IsOrganizationUnitExists(id)
	if err != nil {
		return false, fmt.Errorf("error checking file store: %w", err)
	}
	if exists {
		return true, nil
	}

	// Check DB store
	return h.dbStore.IsOrganizationUnitExists(id)
}

// CheckOrganizationUnitNameConflict checks for name conflicts in both stores.
func (h *ouHybridStore) CheckOrganizationUnitNameConflict(name string, parent *string) (bool, error) {
	// Check file store
	hasConflict, err := h.fileStore.CheckOrganizationUnitNameConflict(name, parent)
	if err != nil {
		return false, fmt.Errorf("error checking file store: %w", err)
	}
	if hasConflict {
		return true, nil
	}

	// Check DB store
	return h.dbStore.CheckOrganizationUnitNameConflict(name, parent)
}

// CheckOrganizationUnitHandleConflict checks for handle conflicts in both stores.
func (h *ouHybridStore) CheckOrganizationUnitHandleConflict(handle string, parent *string) (bool, error) {
	// Check file store
	hasConflict, err := h.fileStore.CheckOrganizationUnitHandleConflict(handle, parent)
	if err != nil {
		return false, fmt.Errorf("error checking file store: %w", err)
	}
	if hasConflict {
		return true, nil
	}

	// Check DB store
	return h.dbStore.CheckOrganizationUnitHandleConflict(handle, parent)
}

// UpdateOrganizationUnit updates an organization unit in the database store only.
// Returns error if the OU exists in the file store (immutable).
func (h *ouHybridStore) UpdateOrganizationUnit(ou OrganizationUnit) error {
	// Check if exists in file store (immutable)
	_, err := h.fileStore.GetOrganizationUnit(ou.ID)
	if err == nil {
		return errors.New("cannot update organization unit: exists in immutable file-based store")
	}
	if err != ErrOrganizationUnitNotFound {
		return fmt.Errorf("error checking file store: %w", err)
	}

	// Update in DB store only
	return h.dbStore.UpdateOrganizationUnit(ou)
}

// DeleteOrganizationUnit deletes an organization unit from the database store only.
// Returns error if the OU exists in the file store (immutable).
func (h *ouHybridStore) DeleteOrganizationUnit(id string) error {
	// Check if exists in file store (immutable)
	_, err := h.fileStore.GetOrganizationUnit(id)
	if err == nil {
		return errors.New("cannot delete organization unit: exists in immutable file-based store")
	}
	if err != ErrOrganizationUnitNotFound {
		return fmt.Errorf("error checking file store: %w", err)
	}

	// Delete from DB store only
	return h.dbStore.DeleteOrganizationUnit(id)
}

// CheckOrganizationUnitHasChildResources checks if an OU has child resources in either store.
func (h *ouHybridStore) CheckOrganizationUnitHasChildResources(id string) (bool, error) {
	// Check file store
	hasChildren, err := h.fileStore.CheckOrganizationUnitHasChildResources(id)
	if err != nil {
		return false, fmt.Errorf("error checking file store: %w", err)
	}
	if hasChildren {
		return true, nil
	}

	// Check DB store
	return h.dbStore.CheckOrganizationUnitHasChildResources(id)
}

// GetOrganizationUnitChildrenCount retrieves the total count of child OUs from both stores.
func (h *ouHybridStore) GetOrganizationUnitChildrenCount(parentID string) (int, error) {
	fileCount, err := h.fileStore.GetOrganizationUnitChildrenCount(parentID)
	if err != nil {
		return 0, fmt.Errorf("failed to get file store children count: %w", err)
	}

	_, err = h.dbStore.GetOrganizationUnitChildrenCount(parentID)
	if err != nil {
		return 0, fmt.Errorf("failed to get database store children count: %w", err)
	}

	// Get all children to check for duplicates
	fileChildren, err := h.fileStore.GetOrganizationUnitChildrenList(parentID, 9999, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to get file store children list: %w", err)
	}

	dbChildren, err := h.dbStore.GetOrganizationUnitChildrenList(parentID, 9999, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to get database store children list: %w", err)
	}

	// Create set of file-based IDs
	fileIDs := make(map[string]bool)
	for _, child := range fileChildren {
		fileIDs[child.ID] = true
	}

	// Count unique DB children (exclude duplicates)
	uniqueDBCount := 0
	for _, child := range dbChildren {
		if !fileIDs[child.ID] {
			uniqueDBCount++
		}
	}

	return fileCount + uniqueDBCount, nil
}

// GetOrganizationUnitChildrenList retrieves child OUs from both stores with pagination.
func (h *ouHybridStore) GetOrganizationUnitChildrenList(
	parentID string, limit, offset int,
) ([]OrganizationUnitBasic, error) {
	// Get all from both stores
	fileChildren, err := h.fileStore.GetOrganizationUnitChildrenList(parentID, 9999, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get file store children: %w", err)
	}

	dbChildren, err := h.dbStore.GetOrganizationUnitChildrenList(parentID, 9999, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get database store children: %w", err)
	}

	// Merge: file-based first, then DB (avoid duplicates by ID)
	seen := make(map[string]bool)
	var merged []OrganizationUnitBasic

	for _, child := range fileChildren {
		merged = append(merged, child)
		seen[child.ID] = true
	}

	for _, child := range dbChildren {
		if !seen[child.ID] {
			merged = append(merged, child)
		}
	}

	// Apply pagination to merged results
	start := offset
	if start > len(merged) {
		return []OrganizationUnitBasic{}, nil
	}

	end := start + limit
	if end > len(merged) {
		end = len(merged)
	}

	return merged[start:end], nil
}

// GetOrganizationUnitUsersCount retrieves user count from database store only.
// File-based OUs don't have runtime user associations.
func (h *ouHybridStore) GetOrganizationUnitUsersCount(ouID string) (int, error) {
	return h.dbStore.GetOrganizationUnitUsersCount(ouID)
}

// GetOrganizationUnitUsersList retrieves users from database store only.
// File-based OUs don't have runtime user associations.
func (h *ouHybridStore) GetOrganizationUnitUsersList(ouID string, limit, offset int) ([]User, error) {
	return h.dbStore.GetOrganizationUnitUsersList(ouID, limit, offset)
}

// GetOrganizationUnitGroupsCount retrieves group count from database store only.
// File-based OUs don't have runtime group associations.
func (h *ouHybridStore) GetOrganizationUnitGroupsCount(ouID string) (int, error) {
	return h.dbStore.GetOrganizationUnitGroupsCount(ouID)
}

// GetOrganizationUnitGroupsList retrieves groups from database store only.
// File-based OUs don't have runtime group associations.
func (h *ouHybridStore) GetOrganizationUnitGroupsList(ouID string, limit, offset int) ([]Group, error) {
	return h.dbStore.GetOrganizationUnitGroupsList(ouID, limit, offset)
}
