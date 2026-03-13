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
	"context"
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
	return f.CreateOrganizationUnit(context.TODO(), *ou)
}

// CreateOrganizationUnit implements organizationUnitStoreInterface.
func (f *fileBasedStore) CreateOrganizationUnit(ctx context.Context, ou OrganizationUnit) error {
	return f.GenericFileBasedStore.Create(ou.ID, &ou)
}

// DeleteOrganizationUnit implements organizationUnitStoreInterface.
func (f *fileBasedStore) DeleteOrganizationUnit(ctx context.Context, id string) error {
	return errors.New("DeleteOrganizationUnit is not supported in file-based store")
}

// GetOrganizationUnit implements organizationUnitStoreInterface.
func (f *fileBasedStore) GetOrganizationUnit(ctx context.Context, id string) (OrganizationUnit, error) {
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

// GetOrganizationUnitByHandle implements organizationUnitStoreInterface.
func (f *fileBasedStore) GetOrganizationUnitByHandle(
	ctx context.Context, handle string, parent *string,
) (OrganizationUnit, error) {
	list, err := f.GenericFileBasedStore.List()
	if err != nil {
		return OrganizationUnit{}, err
	}

	for _, item := range list {
		ou, ok := item.Data.(*OrganizationUnit)
		if !ok {
			continue
		}

		parentMatch := (parent == nil && ou.Parent == nil) ||
			(parent != nil && ou.Parent != nil && *parent == *ou.Parent)
		if ou.Handle == handle && parentMatch {
			return *ou, nil
		}
	}

	return OrganizationUnit{}, ErrOrganizationUnitNotFound
}

// GetOrganizationUnitByPath implements organizationUnitStoreInterface.
func (f *fileBasedStore) GetOrganizationUnitByPath(ctx context.Context, handles []string) (OrganizationUnit, error) {
	var currentOU *OrganizationUnit
	var currentParent *string

	for _, handle := range handles {
		ou, err := f.GetOrganizationUnitByHandle(ctx, handle, currentParent)
		if err != nil {
			return OrganizationUnit{}, ErrOrganizationUnitNotFound
		}

		currentOU = &ou
		currentParent = &ou.ID
	}

	if currentOU == nil {
		return OrganizationUnit{}, ErrOrganizationUnitNotFound
	}

	return *currentOU, nil
}

// GetOrganizationUnitList implements organizationUnitStoreInterface.
func (f *fileBasedStore) GetOrganizationUnitList(
	ctx context.Context, limit, offset int,
) ([]OrganizationUnitBasic, error) {
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
					LogoURL:     ou.LogoURL,
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
func (f *fileBasedStore) GetOrganizationUnitListCount(ctx context.Context) (int, error) {
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

// GetOrganizationUnitsByIDs implements organizationUnitStoreInterface.
func (f *fileBasedStore) GetOrganizationUnitsByIDs(ctx context.Context, ids []string) ([]OrganizationUnitBasic, error) {
	if len(ids) == 0 {
		return []OrganizationUnitBasic{}, nil
	}

	idSet := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
	}

	list, err := f.GenericFileBasedStore.List()
	if err != nil {
		return nil, err
	}

	var result []OrganizationUnitBasic
	for _, item := range list {
		if ou, ok := item.Data.(*OrganizationUnit); ok {
			if _, found := idSet[ou.ID]; found {
				result = append(result, OrganizationUnitBasic{
					ID:          ou.ID,
					Handle:      ou.Handle,
					Name:        ou.Name,
					Description: ou.Description,
					LogoURL:     ou.LogoURL,
				})
			}
		}
	}

	return result, nil
}

// IsOrganizationUnitExists implements organizationUnitStoreInterface.
func (f *fileBasedStore) IsOrganizationUnitExists(ctx context.Context, id string) (bool, error) {
	_, err := f.GetOrganizationUnit(ctx, id)
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
func (f *fileBasedStore) IsOrganizationUnitDeclarative(ctx context.Context, id string) bool {
	exists, err := f.IsOrganizationUnitExists(ctx, id)
	return err == nil && exists
}

// CheckOrganizationUnitNameConflict implements organizationUnitStoreInterface.
func (f *fileBasedStore) CheckOrganizationUnitNameConflict(
	ctx context.Context, name string, parent *string,
) (bool, error) {
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
func (f *fileBasedStore) CheckOrganizationUnitHandleConflict(
	ctx context.Context, handle string, parent *string,
) (bool, error) {
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
func (f *fileBasedStore) UpdateOrganizationUnit(ctx context.Context, ou OrganizationUnit) error {
	return errors.New("UpdateOrganizationUnit is not supported in file-based store")
}

// GetOrganizationUnitChildrenCount implements organizationUnitStoreInterface.
func (f *fileBasedStore) GetOrganizationUnitChildrenCount(ctx context.Context, id string) (int, error) {
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
	ctx context.Context, id string, limit, offset int) ([]OrganizationUnitBasic, error) {
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
					LogoURL:     ou.LogoURL,
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
