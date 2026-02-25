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
	"context"
	"errors"
	"sort"

	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/declarative_resource/entity"
)

type userSchemaFileBasedStore struct {
	*declarativeresource.GenericFileBasedStore
}

// Create implements declarative_resource.Storer interface for resource loader
func (f *userSchemaFileBasedStore) Create(id string, data interface{}) error {
	schema := data.(*UserSchema)
	return f.CreateUserSchema(context.Background(), *schema)
}

// CreateUserSchema implements userSchemaStoreInterface.
func (f *userSchemaFileBasedStore) CreateUserSchema(ctx context.Context, schema UserSchema) error {
	return f.GenericFileBasedStore.Create(schema.ID, &schema)
}

// DeleteUserSchemaByID implements userSchemaStoreInterface.
func (f *userSchemaFileBasedStore) DeleteUserSchemaByID(ctx context.Context, id string) error {
	return errors.New("DeleteUserSchemaByID is not supported in file-based store")
}

// GetUserSchemaByID implements userSchemaStoreInterface.
func (f *userSchemaFileBasedStore) GetUserSchemaByID(ctx context.Context, schemaID string) (UserSchema, error) {
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
func (f *userSchemaFileBasedStore) GetUserSchemaByName(ctx context.Context, schemaName string) (UserSchema, error) {
	data, err := f.GenericFileBasedStore.GetByField(schemaName, func(d interface{}) string {
		return d.(*UserSchema).Name
	})
	if err != nil {
		return UserSchema{}, ErrUserSchemaNotFound
	}
	return *data.(*UserSchema), nil
}

// GetUserSchemaList implements userSchemaStoreInterface.
func (f *userSchemaFileBasedStore) GetUserSchemaList(
	ctx context.Context, limit, offset int,
) ([]UserSchemaListItem, error) {
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
func (f *userSchemaFileBasedStore) GetUserSchemaListCount(ctx context.Context) (int, error) {
	return f.GenericFileBasedStore.Count()
}

// GetUserSchemaListByOUIDs implements userSchemaStoreInterface.
func (f *userSchemaFileBasedStore) GetUserSchemaListByOUIDs(
	ctx context.Context, ouIDs []string, limit, offset int,
) ([]UserSchemaListItem, error) {
	ouIDSet := make(map[string]struct{}, len(ouIDs))
	for _, id := range ouIDs {
		ouIDSet[id] = struct{}{}
	}

	list, err := f.GenericFileBasedStore.List()
	if err != nil {
		return nil, err
	}

	var filtered []UserSchemaListItem
	for _, item := range list {
		if schema, ok := item.Data.(*UserSchema); ok {
			if _, exists := ouIDSet[schema.OrganizationUnitID]; exists {
				filtered = append(filtered, UserSchemaListItem{
					ID:                    schema.ID,
					Name:                  schema.Name,
					OrganizationUnitID:    schema.OrganizationUnitID,
					AllowSelfRegistration: schema.AllowSelfRegistration,
				})
			}
		}
	}

	// Sort the filtered list by name to ensure deterministic pagination
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Name < filtered[j].Name
	})

	// Apply pagination.
	start := offset
	end := offset + limit
	if start > len(filtered) {
		return []UserSchemaListItem{}, nil
	}
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[start:end], nil
}

// GetUserSchemaListCountByOUIDs implements userSchemaStoreInterface.
func (f *userSchemaFileBasedStore) GetUserSchemaListCountByOUIDs(ctx context.Context, ouIDs []string) (int, error) {
	ouIDSet := make(map[string]struct{}, len(ouIDs))
	for _, id := range ouIDs {
		ouIDSet[id] = struct{}{}
	}

	list, err := f.GenericFileBasedStore.List()
	if err != nil {
		return 0, err
	}

	count := 0
	for _, item := range list {
		if schema, ok := item.Data.(*UserSchema); ok {
			if _, exists := ouIDSet[schema.OrganizationUnitID]; exists {
				count++
			}
		}
	}

	return count, nil
}

// UpdateUserSchemaByID implements userSchemaStoreInterface.
func (f *userSchemaFileBasedStore) UpdateUserSchemaByID(ctx context.Context, schemaID string, schema UserSchema) error {
	return errors.New("UpdateUserSchemaByID is not supported in file-based store")
}

// IsUserSchemaDeclarative returns true as file-based schemas are always immutable.
func (f *userSchemaFileBasedStore) IsUserSchemaDeclarative(schemaID string) bool {
	return true
}

// newUserSchemaFileBasedStore creates a new instance of a file-based store.
func newUserSchemaFileBasedStore() userSchemaStoreInterface {
	genericStore := declarativeresource.NewGenericFileBasedStore(entity.KeyTypeUserSchema)
	return &userSchemaFileBasedStore{
		GenericFileBasedStore: genericStore,
	}
}
