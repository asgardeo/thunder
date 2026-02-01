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

package userschema

import (
	"errors"

	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/declarative_resource/entity"
)

type userSchemaFileBasedStore struct {
	*declarativeresource.GenericFileBasedStore
}

// Create implements declarative_resource.Storer interface for resource loader
func (f *userSchemaFileBasedStore) Create(id string, data interface{}) error {
	schema := data.(*UserSchema)
	return f.CreateUserSchema(*schema)
}

// CreateUserSchema implements userSchemaStoreInterface.
func (f *userSchemaFileBasedStore) CreateUserSchema(schema UserSchema) error {
	return f.GenericFileBasedStore.Create(schema.ID, &schema)
}

// DeleteUserSchemaByID implements userSchemaStoreInterface.
func (f *userSchemaFileBasedStore) DeleteUserSchemaByID(id string) error {
	return errors.New("DeleteUserSchemaByID is not supported in file-based store")
}

// GetUserSchemaByID implements userSchemaStoreInterface.
func (f *userSchemaFileBasedStore) GetUserSchemaByID(schemaID string) (UserSchema, error) {
	data, err := f.GenericFileBasedStore.Get(schemaID)
	if err != nil {
		return UserSchema{}, ErrUserSchemaNotFound
	}
	schema, ok := data.(*UserSchema)
	if !ok {
		declarativeresource.LogTypeAssertionError("user schema", schemaID)
		return UserSchema{}, errors.New("user schema data corrupted")
	}
	return *schema, nil
}

// GetUserSchemaByName implements userSchemaStoreInterface.
func (f *userSchemaFileBasedStore) GetUserSchemaByName(schemaName string) (UserSchema, error) {
	data, err := f.GenericFileBasedStore.GetByField(schemaName, func(d interface{}) string {
		return d.(*UserSchema).Name
	})
	if err != nil {
		return UserSchema{}, ErrUserSchemaNotFound
	}
	return *data.(*UserSchema), nil
}

// GetUserSchemaList implements userSchemaStoreInterface.
func (f *userSchemaFileBasedStore) GetUserSchemaList(limit, offset int) ([]UserSchemaListItem, error) {
	list, err := f.GenericFileBasedStore.List()
	if err != nil {
		return nil, err
	}

	var schemaList []UserSchemaListItem
	for _, item := range list {
		if schema, ok := item.Data.(*UserSchema); ok {
			listItem := UserSchemaListItem{
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
		return []UserSchemaListItem{}, nil
	}
	if end > len(schemaList) {
		end = len(schemaList)
	}

	return schemaList[start:end], nil
}

// GetUserSchemaListCount implements userSchemaStoreInterface.
func (f *userSchemaFileBasedStore) GetUserSchemaListCount() (int, error) {
	return f.GenericFileBasedStore.Count()
}

// UpdateUserSchemaByID implements userSchemaStoreInterface.
func (f *userSchemaFileBasedStore) UpdateUserSchemaByID(schemaID string, schema UserSchema) error {
	return errors.New("UpdateUserSchemaByID is not supported in file-based store")
}

// newUserSchemaFileBasedStore creates a new instance of a file-based store.
func newUserSchemaFileBasedStore() userSchemaStoreInterface {
	genericStore := declarativeresource.NewGenericFileBasedStore(entity.KeyTypeUserSchema)
	return &userSchemaFileBasedStore{
		GenericFileBasedStore: genericStore,
	}
}
