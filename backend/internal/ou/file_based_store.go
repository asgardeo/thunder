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

	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/declarative_resource/entity"
)

type fileBasedStore struct {
	*declarativeresource.GenericFileBasedStore
}

// newFileBasedStore creates a new instance of a file-based store.
func newFileBasedStore() organizationUnitStoreInterface {
	genericStore := declarativeresource.NewGenericFileBasedStore(entity.KeyTypeOU)
	return &fileBasedStore{
		GenericFileBasedStore: genericStore,
	}
}

// Create implements declarativeresource.Storer interface for resource loader
func (f *fileBasedStore) Create(id string, data interface{}) error {
	ou := data.(*OrganizationUnit)
	return f.CreateOrganizationUnit(*ou)
}

// CreateOrganizationUnit implements organizationUnitStoreInterface.
func (f *fileBasedStore) CreateOrganizationUnit(ou OrganizationUnit) error {
	return f.GenericFileBasedStore.Create(ou.ID, &ou)
}

// DeleteOrganizationUnit implements organizationUnitStoreInterface.
func (f *fileBasedStore) DeleteOrganizationUnit(id string) error {
	return errors.New("DeleteOrganizationUnit is not supported in file-based store")
}

// GetOrganizationUnit implements organizationUnitStoreInterface.
func (f *fileBasedStore) GetOrganizationUnit(id string) (OrganizationUnit, error) {
	data, err := f.GenericFileBasedStore.Get(id)
	if err != nil {
		return OrganizationUnit{}, ErrOrganizationUnitNotFound
	}
	ou, ok := data.(*OrganizationUnit)
	if !ok {
		declarativeresource.LogTypeAssertionError("organization unit", id)
		return OrganizationUnit{}, errors.New("organization unit data corrupted")
	}
	return *ou, nil
}

// GetOrganizationUnitByPath implements organizationUnitStoreInterface.
func (f *fileBasedStore) GetOrganizationUnitByPath(handles []string) (OrganizationUnit, error) {
	list, err := f.GenericFileBasedStore.List()
	if err != nil {
		return OrganizationUnit{}, err
	}

	// Build the path by traversing the hierarchy
	var currentOU *OrganizationUnit
	var currentParent *string

	for _, handle := range handles {
		found := false
		for _, item := range list {
			if ou, ok := item.Data.(*OrganizationUnit); ok {
				// Check if this OU has the right handle and parent
				parentMatch := (currentParent == nil && ou.Parent == nil) ||
					(currentParent != nil && ou.Parent != nil && *currentParent == *ou.Parent)

				if ou.Handle == handle && parentMatch {
					currentOU = ou
					currentParent = &ou.ID
					found = true
					break
				}
			}
		}
		if !found {
			return OrganizationUnit{}, ErrOrganizationUnitNotFound
		}
	}

	if currentOU == nil {
		return OrganizationUnit{}, ErrOrganizationUnitNotFound
	}

	return *currentOU, nil
}

// GetOrganizationUnitList implements organizationUnitStoreInterface.
func (f *fileBasedStore) GetOrganizationUnitList(limit, offset int) ([]OrganizationUnitBasic, error) {
	list, err := f.GenericFileBasedStore.List()
	if err != nil {
		return nil, err
	}

	var ouList []OrganizationUnitBasic
	for _, item := range list {
		if ou, ok := item.Data.(*OrganizationUnit); ok {
			// Only include root OUs (those without a parent)
			if ou.Parent == nil {
				basicOU := OrganizationUnitBasic{
					ID:          ou.ID,
					Handle:      ou.Handle,
					Name:        ou.Name,
					Description: ou.Description,
				}
				ouList = append(ouList, basicOU)
			}
		}
	}

	// Apply pagination
	start := offset
	if start > len(ouList) {
		return []OrganizationUnitBasic{}, nil
	}
	end := start + limit
	if end > len(ouList) {
		end = len(ouList)
	}

	return ouList[start:end], nil
}

// GetOrganizationUnitListCount implements organizationUnitStoreInterface.
func (f *fileBasedStore) GetOrganizationUnitListCount() (int, error) {
	list, err := f.GenericFileBasedStore.List()
	if err != nil {
		return 0, err
	}

	count := 0
	for _, item := range list {
		if ou, ok := item.Data.(*OrganizationUnit); ok {
			// Only count root OUs (those without a parent)
			if ou.Parent == nil {
				count++
			}
		}
	}

	return count, nil
}

// IsOrganizationUnitExists implements organizationUnitStoreInterface.
func (f *fileBasedStore) IsOrganizationUnitExists(id string) (bool, error) {
	_, err := f.GetOrganizationUnit(id)
	if err != nil {
		if errors.Is(err, ErrOrganizationUnitNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// IsOrganizationUnitDeclarative checks if an organization unit is immutable.
// File-based resources are always immutable, returns true if exists.
func (f *fileBasedStore) IsOrganizationUnitDeclarative(id string) bool {
	exists, err := f.IsOrganizationUnitExists(id)
	return err == nil && exists
}

// CheckOrganizationUnitNameConflict implements organizationUnitStoreInterface.
func (f *fileBasedStore) CheckOrganizationUnitNameConflict(name string, parent *string) (bool, error) {
	list, err := f.GenericFileBasedStore.List()
	if err != nil {
		return false, err
	}

	for _, item := range list {
		if ou, ok := item.Data.(*OrganizationUnit); ok {
			parentMatch := (parent == nil && ou.Parent == nil) ||
				(parent != nil && ou.Parent != nil && *parent == *ou.Parent)

			if ou.Name == name && parentMatch {
				return true, nil
			}
		}
	}

	return false, nil
}

// CheckOrganizationUnitHandleConflict implements organizationUnitStoreInterface.
func (f *fileBasedStore) CheckOrganizationUnitHandleConflict(handle string, parent *string) (bool, error) {
	list, err := f.GenericFileBasedStore.List()
	if err != nil {
		return false, err
	}

	for _, item := range list {
		if ou, ok := item.Data.(*OrganizationUnit); ok {
			parentMatch := (parent == nil && ou.Parent == nil) ||
				(parent != nil && ou.Parent != nil && *parent == *ou.Parent)

			if ou.Handle == handle && parentMatch {
				return true, nil
			}
		}
	}

	return false, nil
}

// UpdateOrganizationUnit implements organizationUnitStoreInterface.
func (f *fileBasedStore) UpdateOrganizationUnit(ou OrganizationUnit) error {
	return errors.New("UpdateOrganizationUnit is not supported in file-based store")
}

// CheckOrganizationUnitHasChildResources implements organizationUnitStoreInterface.
func (f *fileBasedStore) CheckOrganizationUnitHasChildResources(id string) (bool, error) {
	// In file-based mode, we check if there are any child OUs
	list, err := f.GenericFileBasedStore.List()
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

// GetOrganizationUnitChildrenCount implements organizationUnitStoreInterface.
func (f *fileBasedStore) GetOrganizationUnitChildrenCount(id string) (int, error) {
	list, err := f.GenericFileBasedStore.List()
	if err != nil {
		return 0, err
	}

	count := 0
	for _, item := range list {
		if ou, ok := item.Data.(*OrganizationUnit); ok {
			if ou.Parent != nil && *ou.Parent == id {
				count++
			}
		}
	}

	return count, nil
}

// GetOrganizationUnitChildrenList implements organizationUnitStoreInterface.
func (f *fileBasedStore) GetOrganizationUnitChildrenList(
	id string, limit, offset int) ([]OrganizationUnitBasic, error) {
	list, err := f.GenericFileBasedStore.List()
	if err != nil {
		return nil, err
	}

	var children []OrganizationUnitBasic
	for _, item := range list {
		if ou, ok := item.Data.(*OrganizationUnit); ok {
			if ou.Parent != nil && *ou.Parent == id {
				basicOU := OrganizationUnitBasic{
					ID:          ou.ID,
					Handle:      ou.Handle,
					Name:        ou.Name,
					Description: ou.Description,
				}
				children = append(children, basicOU)
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

// GetOrganizationUnitUsersCount implements organizationUnitStoreInterface.
func (f *fileBasedStore) GetOrganizationUnitUsersCount(id string) (int, error) {
	// In file-based mode, users are not stored with OUs
	return 0, nil
}

// GetOrganizationUnitUsersList implements organizationUnitStoreInterface.
func (f *fileBasedStore) GetOrganizationUnitUsersList(id string, limit, offset int) ([]User, error) {
	// In file-based mode, users are not stored with OUs
	return []User{}, nil
}

// GetOrganizationUnitGroupsCount implements organizationUnitStoreInterface.
func (f *fileBasedStore) GetOrganizationUnitGroupsCount(id string) (int, error) {
	// In file-based mode, groups are not stored with OUs
	return 0, nil
}

// GetOrganizationUnitGroupsList implements organizationUnitStoreInterface.
func (f *fileBasedStore) GetOrganizationUnitGroupsList(id string, limit, offset int) ([]Group, error) {
	// In file-based mode, groups are not stored with OUs
	return []Group{}, nil
}
