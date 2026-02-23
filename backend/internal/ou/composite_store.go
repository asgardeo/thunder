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
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
)

// compositeOUStore implements a composite store that combines file-based (immutable) and database (mutable) stores.
// - Read operations query both stores and merge results
// - Write operations (Create/Update/Delete) only affect the database store
// - Declarative OUs (from YAML files) cannot be modified or deleted
type compositeOUStore struct {
	fileStore organizationUnitStoreInterface
	dbStore   organizationUnitStoreInterface
}

// newCompositeOUStore creates a new composite store with both file-based and database stores.
func newCompositeOUStore(fileStore, dbStore organizationUnitStoreInterface) *compositeOUStore {
	return &compositeOUStore{
		fileStore: fileStore,
		dbStore:   dbStore,
	}
}

// GetOrganizationUnitListCount retrieves the total count of organization units from both stores.
func (c *compositeOUStore) GetOrganizationUnitListCount() (int, error) {
	return declarativeresource.CompositeMergeCountHelper(
		func() (int, error) { return c.dbStore.GetOrganizationUnitListCount() },
		func() (int, error) { return c.fileStore.GetOrganizationUnitListCount() },
	)
}

// GetOrganizationUnitList retrieves organization units from both stores with pagination.
// Applies the 1000-record limit in composite mode to prevent memory exhaustion.
// Returns ErrResultLimitExceededInCompositeMode if the limit is exceeded.
func (c *compositeOUStore) GetOrganizationUnitList(limit, offset int) ([]OrganizationUnitBasic, error) {
	items, limitExceeded, err := declarativeresource.CompositeMergeListHelperWithLimit(
		func() (int, error) { return c.dbStore.GetOrganizationUnitListCount() },
		func() (int, error) { return c.fileStore.GetOrganizationUnitListCount() },
		func(count int) ([]OrganizationUnitBasic, error) { return c.dbStore.GetOrganizationUnitList(count, 0) },
		func(count int) ([]OrganizationUnitBasic, error) { return c.fileStore.GetOrganizationUnitList(count, 0) },
		mergeAndDeduplicateOUs,
		limit,
		offset,
		serverconst.MaxCompositeStoreRecords, // Apply 1000-record limit
	)
	if err != nil {
		return nil, err
	}
	// Return limit exceeded as an error
	if limitExceeded {
		return nil, ErrResultLimitExceededInCompositeMode
	}
	return items, nil
}

// GetOrganizationUnitsByIDs retrieves organization units matching the given IDs from both stores.
func (c *compositeOUStore) GetOrganizationUnitsByIDs(ids []string) ([]OrganizationUnitBasic, error) {
	if len(ids) == 0 {
		return []OrganizationUnitBasic{}, nil
	}

	dbOUs, err := c.dbStore.GetOrganizationUnitsByIDs(ids)
	if err != nil {
		return nil, err
	}

	fileOUs, err := c.fileStore.GetOrganizationUnitsByIDs(ids)
	if err != nil {
		return nil, err
	}

	return mergeAndDeduplicateOUs(dbOUs, fileOUs), nil
}

// CreateOrganizationUnit creates a new organization unit in the database store only.
// Conflict checking and parent validation are handled at the service layer.
func (c *compositeOUStore) CreateOrganizationUnit(ou OrganizationUnit) error {
	return c.dbStore.CreateOrganizationUnit(ou)
}

// GetOrganizationUnit retrieves an organization unit by ID from either store.
// Checks database store first, then falls back to file store.
func (c *compositeOUStore) GetOrganizationUnit(id string) (OrganizationUnit, error) {
	return declarativeresource.CompositeGetHelper(
		func() (OrganizationUnit, error) { return c.dbStore.GetOrganizationUnit(id) },
		func() (OrganizationUnit, error) { return c.fileStore.GetOrganizationUnit(id) },
		ErrOrganizationUnitNotFound,
	)
}

// GetOrganizationUnitByPath retrieves an organization unit by hierarchical path from either store.
func (c *compositeOUStore) GetOrganizationUnitByPath(handles []string) (OrganizationUnit, error) {
	return declarativeresource.CompositeGetHelper(
		func() (OrganizationUnit, error) { return c.dbStore.GetOrganizationUnitByPath(handles) },
		func() (OrganizationUnit, error) { return c.fileStore.GetOrganizationUnitByPath(handles) },
		ErrOrganizationUnitNotFound,
	)
}

// IsOrganizationUnitExists checks if an organization unit exists in either store.
func (c *compositeOUStore) IsOrganizationUnitExists(id string) (bool, error) {
	return declarativeresource.CompositeBooleanCheckHelper(
		func() (bool, error) { return c.fileStore.IsOrganizationUnitExists(id) },
		func() (bool, error) { return c.dbStore.IsOrganizationUnitExists(id) },
	)
}

// IsOrganizationUnitDeclarative checks if an organization unit is immutable (exists in file store).
func (c *compositeOUStore) IsOrganizationUnitDeclarative(id string) bool {
	return declarativeresource.CompositeIsDeclarativeHelper(
		id,
		func(id string) (bool, error) { return c.fileStore.IsOrganizationUnitExists(id) },
	)
}

// CheckOrganizationUnitNameConflict checks for name conflicts in both stores.
func (c *compositeOUStore) CheckOrganizationUnitNameConflict(name string, parent *string) (bool, error) {
	return declarativeresource.CompositeBooleanCheckHelper(
		func() (bool, error) { return c.fileStore.CheckOrganizationUnitNameConflict(name, parent) },
		func() (bool, error) { return c.dbStore.CheckOrganizationUnitNameConflict(name, parent) },
	)
}

// CheckOrganizationUnitHandleConflict checks for handle conflicts in both stores.
func (c *compositeOUStore) CheckOrganizationUnitHandleConflict(handle string, parent *string) (bool, error) {
	return declarativeresource.CompositeBooleanCheckHelper(
		func() (bool, error) { return c.fileStore.CheckOrganizationUnitHandleConflict(handle, parent) },
		func() (bool, error) { return c.dbStore.CheckOrganizationUnitHandleConflict(handle, parent) },
	)
}

// UpdateOrganizationUnit updates an organization unit in the database store only.
// Immutability checks and parent validation are handled at the service layer.
func (c *compositeOUStore) UpdateOrganizationUnit(ou OrganizationUnit) error {
	return c.dbStore.UpdateOrganizationUnit(ou)
}

// DeleteOrganizationUnit deletes an organization unit from the database store only.
// Immutability and children validation are handled at the service layer.
func (c *compositeOUStore) DeleteOrganizationUnit(id string) error {
	return c.dbStore.DeleteOrganizationUnit(id)
}

// CheckOrganizationUnitHasChildResources checks if an OU has child resources in either store.
func (c *compositeOUStore) CheckOrganizationUnitHasChildResources(id string) (bool, error) {
	return declarativeresource.CompositeBooleanCheckHelper(
		func() (bool, error) { return c.fileStore.CheckOrganizationUnitHasChildResources(id) },
		func() (bool, error) { return c.dbStore.CheckOrganizationUnitHasChildResources(id) },
	)
}

// GetOrganizationUnitChildrenCount retrieves the count of child OUs from both stores.
func (c *compositeOUStore) GetOrganizationUnitChildrenCount(id string) (int, error) {
	return declarativeresource.CompositeMergeCountHelper(
		func() (int, error) { return c.dbStore.GetOrganizationUnitChildrenCount(id) },
		func() (int, error) { return c.fileStore.GetOrganizationUnitChildrenCount(id) },
	)
}

// GetOrganizationUnitChildrenList retrieves child OUs from both stores with pagination.
// Applies the 1000-record limit in composite mode to prevent memory exhaustion.
// Returns ErrResultLimitExceededInCompositeMode if the limit is exceeded.
func (c *compositeOUStore) GetOrganizationUnitChildrenList(
	id string, limit, offset int) ([]OrganizationUnitBasic, error) {
	items, limitExceeded, err := declarativeresource.CompositeMergeListHelperWithLimit(
		func() (int, error) { return c.dbStore.GetOrganizationUnitChildrenCount(id) },
		func() (int, error) { return c.fileStore.GetOrganizationUnitChildrenCount(id) },
		func(count int) ([]OrganizationUnitBasic, error) {
			return c.dbStore.GetOrganizationUnitChildrenList(id, count, 0)
		},
		func(count int) ([]OrganizationUnitBasic, error) {
			return c.fileStore.GetOrganizationUnitChildrenList(id, count, 0)
		},
		mergeAndDeduplicateChildren,
		limit,
		offset,
		serverconst.MaxCompositeStoreRecords, // Apply 1000-record limit
	)
	if err != nil {
		return nil, err
	}
	// Return limit exceeded as an error
	if limitExceeded {
		return nil, ErrResultLimitExceededInCompositeMode
	}
	return items, nil
}

// GetOrganizationUnitUsersCount retrieves the count of users from the database store only.
func (c *compositeOUStore) GetOrganizationUnitUsersCount(id string) (int, error) {
	return c.dbStore.GetOrganizationUnitUsersCount(id)
}

// GetOrganizationUnitUsersList retrieves users from the database store only.
func (c *compositeOUStore) GetOrganizationUnitUsersList(id string, limit, offset int) ([]User, error) {
	return c.dbStore.GetOrganizationUnitUsersList(id, limit, offset)
}

// GetOrganizationUnitGroupsCount retrieves the count of groups from the database store only.
func (c *compositeOUStore) GetOrganizationUnitGroupsCount(id string) (int, error) {
	return c.dbStore.GetOrganizationUnitGroupsCount(id)
}

// GetOrganizationUnitGroupsList retrieves groups from the database store only.
func (c *compositeOUStore) GetOrganizationUnitGroupsList(id string, limit, offset int) ([]Group, error) {
	return c.dbStore.GetOrganizationUnitGroupsList(id, limit, offset)
}

// mergeAndDeduplicateOUs merges root-level OUs from both stores and removes duplicates by ID.
// While duplicates shouldn't exist by design (an OU exists in only one store), this provides
// defensive programming against misconfigurations or bugs.
func mergeAndDeduplicateOUs(dbOUs, fileOUs []OrganizationUnitBasic) []OrganizationUnitBasic {
	seen := make(map[string]bool)
	result := make([]OrganizationUnitBasic, 0, len(dbOUs)+len(fileOUs))

	// Add DB OUs first (they take precedence) - mark as mutable (isReadOnly=false)
	for i := range dbOUs {
		if !seen[dbOUs[i].ID] {
			seen[dbOUs[i].ID] = true
			dbOUs[i].IsReadOnly = false
			result = append(result, dbOUs[i])
		}
	}

	// Add file OUs if not already present - mark as immutable (isReadOnly=true)
	for i := range fileOUs {
		if !seen[fileOUs[i].ID] {
			seen[fileOUs[i].ID] = true
			fileOUs[i].IsReadOnly = true
			result = append(result, fileOUs[i])
		}
	}

	return result
}

// mergeAndDeduplicateChildren merges children from both stores and removes duplicates by ID.
// While duplicates shouldn't exist by design (an OU exists in only one store), this provides
// defensive programming against misconfigurations or bugs.
func mergeAndDeduplicateChildren(dbChildren, fileChildren []OrganizationUnitBasic) []OrganizationUnitBasic {
	seen := make(map[string]bool)
	result := make([]OrganizationUnitBasic, 0, len(dbChildren)+len(fileChildren))

	// Add DB children first (they take precedence) - mark as mutable (isReadOnly=false)
	for i := range dbChildren {
		if !seen[dbChildren[i].ID] {
			seen[dbChildren[i].ID] = true
			dbChildren[i].IsReadOnly = false
			result = append(result, dbChildren[i])
		}
	}

	// Add file children if not already present - mark as immutable (isReadOnly=true)
	for i := range fileChildren {
		if !seen[fileChildren[i].ID] {
			seen[fileChildren[i].ID] = true
			fileChildren[i].IsReadOnly = true
			result = append(result, fileChildren[i])
		}
	}

	return result
}
