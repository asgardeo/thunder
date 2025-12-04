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

	"github.com/asgardeo/thunder/internal/system/file_based_runtime/entity"
	"github.com/asgardeo/thunder/internal/system/log"
)

type userSchemaFileBasedStore struct {
	storage entity.StoreInterface
}

// CreateUserSchema implements userSchemaStoreInterface.
func (f *userSchemaFileBasedStore) CreateUserSchema(schema UserSchema) error {
	schemaKey := entity.NewCompositeKey(schema.ID, entity.KeyTypeUserSchema)
	return f.storage.Set(schemaKey, &schema)
}

// DeleteUserSchemaByID implements userSchemaStoreInterface.
func (f *userSchemaFileBasedStore) DeleteUserSchemaByID(id string) error {
	return errors.New("delete operation not supported in immutable mode")
}

// GetUserSchemaByID implements userSchemaStoreInterface.
func (f *userSchemaFileBasedStore) GetUserSchemaByID(schemaID string) (UserSchema, error) {
	entity, err := f.storage.Get(entity.NewCompositeKey(schemaID, entity.KeyTypeUserSchema))
	if err != nil {
		return UserSchema{}, ErrUserSchemaNotFound
	}
	schema, ok := entity.Data.(*UserSchema)
	if !ok {
		log.GetLogger().Error("Type assertion failed while retrieving user schema by ID",
			log.String("schemaID", schemaID))
		return UserSchema{}, errors.New("user schema data corrupted")
	}
	return *schema, nil
}

// GetUserSchemaByName implements userSchemaStoreInterface.
func (f *userSchemaFileBasedStore) GetUserSchemaByName(schemaName string) (UserSchema, error) {
	list, err := f.storage.ListByType(entity.KeyTypeUserSchema)
	if err != nil {
		return UserSchema{}, err
	}

	for _, item := range list {
		if schema, ok := item.Data.(*UserSchema); ok && schema.Name == schemaName {
			return *schema, nil
		}
	}

	return UserSchema{}, ErrUserSchemaNotFound
}

// GetUserSchemaList implements userSchemaStoreInterface.
func (f *userSchemaFileBasedStore) GetUserSchemaList(limit, offset int) ([]UserSchemaListItem, error) {
	list, err := f.storage.ListByType(entity.KeyTypeUserSchema)
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
	list, err := f.storage.ListByType(entity.KeyTypeUserSchema)
	if err != nil {
		return 0, err
	}
	return len(list), nil
}

// UpdateUserSchemaByID implements userSchemaStoreInterface.
func (f *userSchemaFileBasedStore) UpdateUserSchemaByID(schemaID string, schema UserSchema) error {
	return errors.New("update operation not supported in immutable mode")
}

// newUserSchemaFileBasedStore creates a new instance of a file-based store.
func newUserSchemaFileBasedStore() userSchemaStoreInterface {
	storage := entity.NewStore()
	return &userSchemaFileBasedStore{
		storage: storage,
	}
}
