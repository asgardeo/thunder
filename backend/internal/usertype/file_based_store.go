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

package usertype

import (
	"errors"

	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/declarative_resource/entity"
)

type userTypeFileBasedStore struct {
	*declarativeresource.GenericFileBasedStore
}

// Create implements declarative_resource.Storer interface for resource loader
func (f *userTypeFileBasedStore) Create(id string, data interface{}) error {
	schema := data.(*UserType)
	return f.CreateUserType(*schema)
}

// CreateUserType implements userTypeStoreInterface.
func (f *userTypeFileBasedStore) CreateUserType(schema UserType) error {
	return f.GenericFileBasedStore.Create(schema.ID, &schema)
}

// DeleteUserTypeByID implements userTypeStoreInterface.
func (f *userTypeFileBasedStore) DeleteUserTypeByID(id string) error {
	return errors.New("DeleteUserTypeByID is not supported in file-based store")
}

// GetUserTypeByID implements userTypeStoreInterface.
func (f *userTypeFileBasedStore) GetUserTypeByID(schemaID string) (UserType, error) {
	data, err := f.GenericFileBasedStore.Get(schemaID)
	if err != nil {
		return UserType{}, ErrUserTypeNotFound
	}
	schema, ok := data.(*UserType)
	if !ok {
		declarativeresource.LogTypeAssertionError("user type", schemaID)
		return UserType{}, errors.New("user type data corrupted")
	}
	return *schema, nil
}

// GetUserTypeByName implements userTypeStoreInterface.
func (f *userTypeFileBasedStore) GetUserTypeByName(schemaName string) (UserType, error) {
	data, err := f.GenericFileBasedStore.GetByField(schemaName, func(d interface{}) string {
		return d.(*UserType).Name
	})
	if err != nil {
		return UserType{}, ErrUserTypeNotFound
	}
	return *data.(*UserType), nil
}

// GetUserTypeList implements userTypeStoreInterface.
func (f *userTypeFileBasedStore) GetUserTypeList(limit, offset int) ([]UserTypeListItem, error) {
	list, err := f.GenericFileBasedStore.List()
	if err != nil {
		return nil, err
	}

	var schemaList []UserTypeListItem
	for _, item := range list {
		if schema, ok := item.Data.(*UserType); ok {
			listItem := UserTypeListItem{
				ID:                    schema.ID,
				Name:                  schema.Name,
				OrganizationUnitID:    schema.OrganizationUnitID,
				AllowSelfRegistration: schema.AllowSelfRegistration,
			}
			schemaList = append(schemaList, listItem)
		}
	}

	// Apply pagination
	start := offset
	end := offset + limit
	if start > len(schemaList) {
		return []UserTypeListItem{}, nil
	}
	if end > len(schemaList) {
		end = len(schemaList)
	}

	return schemaList[start:end], nil
}

// GetUserTypeListCount implements userTypeStoreInterface.
func (f *userTypeFileBasedStore) GetUserTypeListCount() (int, error) {
	return f.GenericFileBasedStore.Count()
}

// UpdateUserTypeByID implements userTypeStoreInterface.
func (f *userTypeFileBasedStore) UpdateUserTypeByID(schemaID string, schema UserType) error {
	return errors.New("UpdateUserTypeByID is not supported in file-based store")
}

// newUserTypeFileBasedStore creates a new instance of a file-based store.
func newUserTypeFileBasedStore() userTypeStoreInterface {
	genericStore := declarativeresource.NewGenericFileBasedStore(entity.KeyTypeUserType)
	return &userTypeFileBasedStore{
		GenericFileBasedStore: genericStore,
	}
}
