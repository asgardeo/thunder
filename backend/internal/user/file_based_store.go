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

package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/declarative_resource/entity"
)

type userResource struct {
	User        User
	Credentials Credentials
}

type userFileBasedStore struct {
	*declarativeresource.GenericFileBasedStore
}

// newUserFileBasedStore creates a new instance of a file-based store.
func newUserFileBasedStore() userStoreInterface {
	genericStore := declarativeresource.NewGenericFileBasedStore(entity.KeyTypeUser)
	return &userFileBasedStore{
		GenericFileBasedStore: genericStore,
	}
}

// Create implements declarativeresource.Storer interface for resource loader
func (f *userFileBasedStore) Create(id string, data interface{}) error {
	resource, ok := data.(*userResource)
	if !ok {
		declarativeresource.LogTypeAssertionError("user", id)
		return errors.New("invalid data type: expected *userResource")
	}
	return f.CreateUser(context.Background(), resource.User, resource.Credentials)
}

// CreateUser implements userStoreInterface.
func (f *userFileBasedStore) CreateUser(ctx context.Context, user User, credentials Credentials) error {
	resource := &userResource{
		User:        user,
		Credentials: credentials,
	}
	return f.GenericFileBasedStore.Create(user.ID, resource)
}

// GetUser retrieves a user by ID.
func (f *userFileBasedStore) GetUser(ctx context.Context, id string) (User, error) {
	data, err := f.GenericFileBasedStore.Get(id)
	if err != nil {
		return User{}, ErrUserNotFound
	}
	resource, ok := data.(*userResource)
	if !ok {
		declarativeresource.LogTypeAssertionError("user", id)
		return User{}, errors.New("user data corrupted")
	}
	return resource.User, nil
}

// GetUserListCount retrieves the total count of users from the file store.
func (f *userFileBasedStore) GetUserListCount(ctx context.Context, filters map[string]interface{}) (int, error) {
	resources, err := f.listUserResources()
	if err != nil {
		return 0, err
	}

	count := 0
	for _, resource := range resources {
		if matchesFilters(resource.User.Attributes, filters) {
			count++
		}
	}
	return count, nil
}

// GetUserList retrieves users from the file store with pagination and filtering.
func (f *userFileBasedStore) GetUserList(
	ctx context.Context, limit, offset int, filters map[string]interface{},
) ([]User, error) {
	resources, err := f.listUserResources()
	if err != nil {
		return nil, err
	}

	users := make([]User, 0)
	for _, resource := range resources {
		if matchesFilters(resource.User.Attributes, filters) {
			users = append(users, resource.User)
		}
	}

	return applyPagination(users, limit, offset), nil
}

// GetUserListCountByOUIDs retrieves the total count of users by OU IDs.
func (f *userFileBasedStore) GetUserListCountByOUIDs(
	ctx context.Context, ouIDs []string, filters map[string]interface{},
) (int, error) {
	resources, err := f.listUserResources()
	if err != nil {
		return 0, err
	}

	ouIDSet := make(map[string]struct{}, len(ouIDs))
	for _, id := range ouIDs {
		ouIDSet[id] = struct{}{}
	}

	count := 0
	for _, resource := range resources {
		if _, ok := ouIDSet[resource.User.OrganizationUnit]; !ok {
			continue
		}
		if matchesFilters(resource.User.Attributes, filters) {
			count++
		}
	}
	return count, nil
}

// GetUserListByOUIDs retrieves users scoped to OU IDs with pagination and filtering.
func (f *userFileBasedStore) GetUserListByOUIDs(
	ctx context.Context, ouIDs []string, limit, offset int, filters map[string]interface{},
) ([]User, error) {
	resources, err := f.listUserResources()
	if err != nil {
		return nil, err
	}

	ouIDSet := make(map[string]struct{}, len(ouIDs))
	for _, id := range ouIDs {
		ouIDSet[id] = struct{}{}
	}

	users := make([]User, 0)
	for _, resource := range resources {
		if _, ok := ouIDSet[resource.User.OrganizationUnit]; !ok {
			continue
		}
		if matchesFilters(resource.User.Attributes, filters) {
			users = append(users, resource.User)
		}
	}

	return applyPagination(users, limit, offset), nil
}

// GetGroupCountForUser retrieves the count of groups for a given user.
func (f *userFileBasedStore) GetGroupCountForUser(ctx context.Context, userID string) (int, error) {
	return 0, nil
}

// GetUserGroups retrieves groups for a given user.
func (f *userFileBasedStore) GetUserGroups(
	ctx context.Context, userID string, limit, offset int,
) ([]UserGroup, error) {
	return []UserGroup{}, nil
}

// UpdateUser updates a user in the file store (unsupported).
func (f *userFileBasedStore) UpdateUser(ctx context.Context, user *User) error {
	return errors.New("UpdateUser is not supported in file-based store")
}

// UpdateUserCredentials updates a user's credentials in the file store (unsupported).
func (f *userFileBasedStore) UpdateUserCredentials(
	ctx context.Context, userID string, credentials Credentials,
) error {
	return errors.New("UpdateUserCredentials is not supported in file-based store")
}

// DeleteUser deletes a user from the file store (unsupported).
func (f *userFileBasedStore) DeleteUser(ctx context.Context, id string) error {
	return errors.New("DeleteUser is not supported in file-based store")
}

// IdentifyUser identifies a user with the given filters.
func (f *userFileBasedStore) IdentifyUser(
	ctx context.Context, filters map[string]interface{},
) (*string, error) {
	resources, err := f.listUserResources()
	if err != nil {
		return nil, err
	}

	var matches []string
	for _, resource := range resources {
		if matchesFilters(resource.User.Attributes, filters) {
			matches = append(matches, resource.User.ID)
		}
	}

	if len(matches) == 0 {
		return nil, ErrUserNotFound
	}
	if len(matches) > 1 {
		return nil, fmt.Errorf("unexpected number of results: %d", len(matches))
	}

	return &matches[0], nil
}

// GetCredentials retrieves the credentials for a user.
func (f *userFileBasedStore) GetCredentials(
	ctx context.Context, id string,
) (User, Credentials, error) {
	data, err := f.GenericFileBasedStore.Get(id)
	if err != nil {
		return User{}, nil, ErrUserNotFound
	}
	resource, ok := data.(*userResource)
	if !ok {
		declarativeresource.LogTypeAssertionError("user", id)
		return User{}, nil, errors.New("user data corrupted")
	}
	return resource.User, resource.Credentials, nil
}

// ValidateUserIDs checks if all provided user IDs exist.
func (f *userFileBasedStore) ValidateUserIDs(ctx context.Context, userIDs []string) ([]string, error) {
	invalid := make([]string, 0)
	for _, id := range userIDs {
		_, err := f.GetUser(ctx, id)
		if err != nil {
			if errors.Is(err, ErrUserNotFound) {
				invalid = append(invalid, id)
				continue
			}
			return nil, err
		}
	}
	return invalid, nil
}

// IsUserDeclarative checks if a user is immutable (exists in file store).
func (f *userFileBasedStore) IsUserDeclarative(ctx context.Context, id string) (bool, error) {
	_, err := f.GetUser(ctx, id)
	if err == nil {
		return true, nil
	}
	// If not found in file store, it's not declarative
	if strings.Contains(err.Error(), "not found") {
		return false, nil
	}
	return false, err
}

func (f *userFileBasedStore) listUserResources() ([]*userResource, error) {
	list, err := f.GenericFileBasedStore.List()
	if err != nil {
		return nil, err
	}

	resources := make([]*userResource, 0, len(list))
	for _, item := range list {
		resource, ok := item.Data.(*userResource)
		if !ok {
			declarativeresource.LogTypeAssertionError("user", item.ID.ID)
			return nil, errors.New("user data corrupted")
		}
		resources = append(resources, resource)
	}

	return resources, nil
}

func applyPagination(users []User, limit, offset int) []User {
	if limit < 0 {
		return []User{}
	}
	if offset < 0 {
		offset = 0
	}
	if offset >= len(users) {
		return []User{}
	}

	end := offset + limit
	if limit == 0 {
		end = len(users)
	}
	if end > len(users) {
		end = len(users)
	}

	return users[offset:end]
}

func matchesFilters(attributes json.RawMessage, filters map[string]interface{}) bool {
	if len(filters) == 0 {
		return true
	}
	if len(attributes) == 0 {
		return false
	}

	var attrsMap map[string]interface{}
	if err := json.Unmarshal(attributes, &attrsMap); err != nil {
		return false
	}

	for key, expected := range filters {
		value, ok := getNestedValue(attrsMap, key)
		if !ok || !valuesEqual(value, expected) {
			return false
		}
	}

	return true
}

func getNestedValue(data map[string]interface{}, key string) (interface{}, bool) {
	parts := strings.Split(key, ".")
	current := interface{}(data)

	for _, part := range parts {
		obj, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}
		value, exists := obj[part]
		if !exists {
			return nil, false
		}
		current = value
	}

	return current, true
}

func valuesEqual(actual interface{}, expected interface{}) bool {
	switch actualValue := actual.(type) {
	case float64:
		switch expectedValue := expected.(type) {
		case int64:
			return actualValue == float64(expectedValue)
		case float64:
			return actualValue == expectedValue
		case int:
			return actualValue == float64(expectedValue)
		}
	case string:
		if expectedValue, ok := expected.(string); ok {
			return actualValue == expectedValue
		}
	case bool:
		if expectedValue, ok := expected.(bool); ok {
			return actualValue == expectedValue
		}
	}

	return false
}
