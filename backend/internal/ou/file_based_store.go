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

	"github.com/asgardeo/thunder/internal/system/file_based_runtime/entity"
	"github.com/asgardeo/thunder/internal/system/log"
)

type ouFileBasedStore struct {
	storage entity.StoreInterface
}

// newOUFileBasedStore creates a new instance of a file-based store.
func newOUFileBasedStore() organizationUnitStoreInterface {
	store := entity.GetInstance()
	return &ouFileBasedStore{
		storage: store,
	}
}

// CreateOrganizationUnit stores an organization unit in the file-based store.
func (f *ouFileBasedStore) CreateOrganizationUnit(ou OrganizationUnit) error {
	ouKey := entity.NewCompositeKey(ou.ID, entity.KeyTypeOrganizationUnit)
	return f.storage.Set(ouKey, &ou)
}

// GetOrganizationUnitListCount retrieves the total count of organization units.
func (f *ouFileBasedStore) GetOrganizationUnitListCount() (int, error) {
	return f.storage.CountByType(entity.KeyTypeOrganizationUnit)
}

// GetOrganizationUnitList retrieves organization units with pagination.
func (f *ouFileBasedStore) GetOrganizationUnitList(limit, offset int) ([]OrganizationUnitBasic, error) {
	list, err := f.storage.ListByType(entity.KeyTypeOrganizationUnit)
	if err != nil {
		return nil, err
	}

	// Filter for root organization units (parent is nil)
	var rootOUs []OrganizationUnitBasic
	for _, item := range list {
		if ou, ok := item.Data.(*OrganizationUnit); ok && ou.Parent == nil {
			rootOUs = append(rootOUs, OrganizationUnitBasic{
				ID:          ou.ID,
				Handle:      ou.Handle,
				Name:        ou.Name,
				Description: ou.Description,
			})
		}
	}

	// Apply pagination
	start := offset
	if start > len(rootOUs) {
		return []OrganizationUnitBasic{}, nil
	}

	end := start + limit
	if end > len(rootOUs) {
		end = len(rootOUs)
	}

	return rootOUs[start:end], nil
}

// GetOrganizationUnit retrieves an organization unit by its ID.
func (f *ouFileBasedStore) GetOrganizationUnit(id string) (OrganizationUnit, error) {
	entity, err := f.storage.Get(entity.NewCompositeKey(id, entity.KeyTypeOrganizationUnit))
	if err != nil {
		return OrganizationUnit{}, ErrOrganizationUnitNotFound
	}

	ou, ok := entity.Data.(*OrganizationUnit)
	if !ok {
		log.GetLogger().Error("Type assertion failed while retrieving organization unit by ID",
			log.String("ouID", id))
		return OrganizationUnit{}, errors.New("organization unit data corrupted")
	}

	return *ou, nil
}

// GetOrganizationUnitByPath retrieves an organization unit by its hierarchical handle path.
func (f *ouFileBasedStore) GetOrganizationUnitByPath(handles []string) (OrganizationUnit, error) {
	if len(handles) == 0 {
		return OrganizationUnit{}, ErrOrganizationUnitNotFound
	}

	list, err := f.storage.ListByType(entity.KeyTypeOrganizationUnit)
	if err != nil {
		return OrganizationUnit{}, err
	}

	// Convert list to map for easier lookup
	ouMap := make(map[string]*OrganizationUnit)
	for _, item := range list {
		if ou, ok := item.Data.(*OrganizationUnit); ok {
			ouMap[ou.ID] = ou
		}
	}

	// Traverse the path
	var currentOU *OrganizationUnit
	var parentID *string

	for i, handle := range handles {
		found := false
		for _, ou := range ouMap {
			// Match handle and parent
			if ou.Handle == handle {
				if (parentID == nil && ou.Parent == nil) ||
					(parentID != nil && ou.Parent != nil && *parentID == *ou.Parent) {
					currentOU = ou
					parentID = &ou.ID
					found = true
					break
				}
			}
		}

		if !found {
			log.GetLogger().Debug("Organization unit not found in path",
				log.String("handle", handle),
				log.Int("pathIndex", i))
			return OrganizationUnit{}, ErrOrganizationUnitNotFound
		}
	}

	if currentOU == nil {
		return OrganizationUnit{}, ErrOrganizationUnitNotFound
	}

	return *currentOU, nil
}

// IsOrganizationUnitExists checks if an organization unit exists by ID.
func (f *ouFileBasedStore) IsOrganizationUnitExists(id string) (bool, error) {
	_, err := f.storage.Get(entity.NewCompositeKey(id, entity.KeyTypeOrganizationUnit))
	if err != nil {
		return false, nil
	}
	return true, nil
}

// CheckOrganizationUnitNameConflict checks if an organization unit name conflicts under the same parent.
func (f *ouFileBasedStore) CheckOrganizationUnitNameConflict(name string, parentID *string) (bool, error) {
	list, err := f.storage.ListByType(entity.KeyTypeOrganizationUnit)
	if err != nil {
		return false, err
	}

	for _, item := range list {
		if ou, ok := item.Data.(*OrganizationUnit); ok {
			if ou.Name == name {
				// Check if parents match
				if (parentID == nil && ou.Parent == nil) ||
					(parentID != nil && ou.Parent != nil && *parentID == *ou.Parent) {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// CheckOrganizationUnitHandleConflict checks if an organization unit handle conflicts under the same parent.
func (f *ouFileBasedStore) CheckOrganizationUnitHandleConflict(handle string, parentID *string) (bool, error) {
	list, err := f.storage.ListByType(entity.KeyTypeOrganizationUnit)
	if err != nil {
		return false, err
	}

	for _, item := range list {
		if ou, ok := item.Data.(*OrganizationUnit); ok {
			if ou.Handle == handle {
				// Check if parents match
				if (parentID == nil && ou.Parent == nil) ||
					(parentID != nil && ou.Parent != nil && *parentID == *ou.Parent) {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// UpdateOrganizationUnit is not supported in file-based store.
func (f *ouFileBasedStore) UpdateOrganizationUnit(ou OrganizationUnit) error {
	return errors.New("UpdateOrganizationUnit is not supported in file-based store")
}

// DeleteOrganizationUnit is not supported in file-based store.
func (f *ouFileBasedStore) DeleteOrganizationUnit(id string) error {
	return errors.New("DeleteOrganizationUnit is not supported in file-based store")
}

// CheckOrganizationUnitHasChildResources checks if an organization unit has child OUs.
func (f *ouFileBasedStore) CheckOrganizationUnitHasChildResources(id string) (bool, error) {
	list, err := f.storage.ListByType(entity.KeyTypeOrganizationUnit)
	if err != nil {
		return false, err
	}

	for _, item := range list {
		if ou, ok := item.Data.(*OrganizationUnit); ok {
			if ou.Parent != nil && *ou.Parent == id {
				return true, nil
			}
		}
	}

	return false, nil
}

// GetOrganizationUnitChildrenCount retrieves the total count of child organization units.
func (f *ouFileBasedStore) GetOrganizationUnitChildrenCount(parentID string) (int, error) {
	list, err := f.storage.ListByType(entity.KeyTypeOrganizationUnit)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, item := range list {
		if ou, ok := item.Data.(*OrganizationUnit); ok {
			if ou.Parent != nil && *ou.Parent == parentID {
				count++
			}
		}
	}

	return count, nil
}

// GetOrganizationUnitChildrenList retrieves a paginated list of child organization units.
func (f *ouFileBasedStore) GetOrganizationUnitChildrenList(
	parentID string, limit, offset int,
) ([]OrganizationUnitBasic, error) {
	list, err := f.storage.ListByType(entity.KeyTypeOrganizationUnit)
	if err != nil {
		return nil, err
	}

	// Filter children
	var children []OrganizationUnitBasic
	for _, item := range list {
		if ou, ok := item.Data.(*OrganizationUnit); ok {
			if ou.Parent != nil && *ou.Parent == parentID {
				children = append(children, OrganizationUnitBasic{
					ID:          ou.ID,
					Handle:      ou.Handle,
					Name:        ou.Name,
					Description: ou.Description,
				})
			}
		}
	}

	// Apply pagination
	start := offset
	if start > len(children) {
		return []OrganizationUnitBasic{}, nil
	}

	end := start + limit
	if end > len(children) {
		end = len(children)
	}

	return children[start:end], nil
}

// GetOrganizationUnitUsersCount returns 0 for file-based store (no runtime users).
func (f *ouFileBasedStore) GetOrganizationUnitUsersCount(ouID string) (int, error) {
	// File-based OUs don't contain runtime users
	return 0, nil
}

// GetOrganizationUnitUsersList returns empty list for file-based store (no runtime users).
func (f *ouFileBasedStore) GetOrganizationUnitUsersList(ouID string, limit, offset int) ([]User, error) {
	// File-based OUs don't contain runtime users
	return []User{}, nil
}

// GetOrganizationUnitGroupsCount returns 0 for file-based store (no runtime groups).
func (f *ouFileBasedStore) GetOrganizationUnitGroupsCount(ouID string) (int, error) {
	// File-based OUs don't contain runtime groups
	return 0, nil
}

// GetOrganizationUnitGroupsList returns empty list for file-based store (no runtime groups).
func (f *ouFileBasedStore) GetOrganizationUnitGroupsList(ouID string, limit, offset int) ([]Group, error) {
	// File-based OUs don't contain runtime groups
	return []Group{}, nil
}
