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
	"errors"
	"sort"
	"strings"

	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
)

// compositeUserStore implements a composite store that combines file-based (immutable) and
// database (mutable) stores.
// - Read operations query both stores and merge results
// - Write operations (Create/Update/Delete) only affect the database store
// - Declarative users (from YAML files) cannot be modified or deleted
type compositeUserStore struct {
	fileStore userStoreInterface
	dbStore   userStoreInterface
}

// newCompositeUserStore creates a new composite store with both file-based and database stores.
func newCompositeUserStore(fileStore, dbStore userStoreInterface) *compositeUserStore {
	return &compositeUserStore{
		fileStore: fileStore,
		dbStore:   dbStore,
	}
}

// GetUserListCount retrieves the total count of users from both stores.
func (c *compositeUserStore) GetUserListCount(ctx context.Context, filters map[string]interface{}) (int, error) {
	return c.getDistinctUserCount(
		func() (int, error) { return c.dbStore.GetUserListCount(ctx, filters) },
		func() (int, error) { return c.fileStore.GetUserListCount(ctx, filters) },
		func(count int) ([]User, error) { return c.dbStore.GetUserList(ctx, count, 0, filters) },
		func(count int) ([]User, error) { return c.fileStore.GetUserList(ctx, count, 0, filters) },
	)
}

// GetUserList retrieves users from both stores with pagination.
// Applies the 1000-record limit in composite mode to prevent memory exhaustion.
// Returns errResultLimitExceededInCompositeMode if the limit is exceeded.
func (c *compositeUserStore) GetUserList(
	ctx context.Context, limit, offset int, filters map[string]interface{},
) ([]User, error) {
	users, limitExceeded, err := declarativeresource.CompositeMergeListHelperWithLimit(
		func() (int, error) { return c.dbStore.GetUserListCount(ctx, filters) },
		func() (int, error) { return c.fileStore.GetUserListCount(ctx, filters) },
		func(count int) ([]User, error) {
			return c.dbStore.GetUserList(ctx, count, 0, filters)
		},
		func(count int) ([]User, error) {
			return c.fileStore.GetUserList(ctx, count, 0, filters)
		},
		mergeAndDeduplicateUsers,
		limit,
		offset,
		serverconst.MaxCompositeStoreRecords,
	)
	if err != nil {
		return nil, err
	}
	if limitExceeded {
		return nil, errResultLimitExceededInCompositeMode
	}
	return users, nil
}

// GetUserListCountByOUIDs retrieves the total count of users by OU IDs from both stores.
func (c *compositeUserStore) GetUserListCountByOUIDs(
	ctx context.Context, ouIDs []string, filters map[string]interface{},
) (int, error) {
	return c.getDistinctUserCount(
		func() (int, error) { return c.dbStore.GetUserListCountByOUIDs(ctx, ouIDs, filters) },
		func() (int, error) { return c.fileStore.GetUserListCountByOUIDs(ctx, ouIDs, filters) },
		func(count int) ([]User, error) { return c.dbStore.GetUserListByOUIDs(ctx, ouIDs, count, 0, filters) },
		func(count int) ([]User, error) { return c.fileStore.GetUserListByOUIDs(ctx, ouIDs, count, 0, filters) },
	)
}

// getDistinctUserCount retrieves the count of distinct users
// from both stores based on provided count and list functions.
func (c *compositeUserStore) getDistinctUserCount(
	dbCount func() (int, error),
	fileCount func() (int, error),
	dbList func(count int) ([]User, error),
	fileList func(count int) ([]User, error),
) (int, error) {
	count, err := dbCount()
	if err != nil {
		return 0, err
	}
	fileCountValue, err := fileCount()
	if err != nil {
		return 0, err
	}

	userIDs := make(map[string]struct{}, count+fileCountValue)
	if count > 0 {
		users, err := dbList(count)
		if err != nil {
			return 0, err
		}
		for _, user := range users {
			userIDs[user.ID] = struct{}{}
		}
	}

	if fileCountValue > 0 {
		users, err := fileList(fileCountValue)
		if err != nil {
			return 0, err
		}
		for _, user := range users {
			userIDs[user.ID] = struct{}{}
		}
	}

	return len(userIDs), nil
}

// GetUserListByOUIDs retrieves users scoped to OU IDs from both stores with pagination.
// Applies the 1000-record limit in composite mode to prevent memory exhaustion.
// Returns errResultLimitExceededInCompositeMode if the limit is exceeded.
func (c *compositeUserStore) GetUserListByOUIDs(
	ctx context.Context, ouIDs []string, limit, offset int, filters map[string]interface{},
) ([]User, error) {
	users, limitExceeded, err := declarativeresource.CompositeMergeListHelperWithLimit(
		func() (int, error) { return c.dbStore.GetUserListCountByOUIDs(ctx, ouIDs, filters) },
		func() (int, error) { return c.fileStore.GetUserListCountByOUIDs(ctx, ouIDs, filters) },
		func(count int) ([]User, error) {
			return c.dbStore.GetUserListByOUIDs(ctx, ouIDs, count, 0, filters)
		},
		func(count int) ([]User, error) {
			return c.fileStore.GetUserListByOUIDs(ctx, ouIDs, count, 0, filters)
		},
		mergeAndDeduplicateUsers,
		limit,
		offset,
		serverconst.MaxCompositeStoreRecords,
	)
	if err != nil {
		return nil, err
	}
	if limitExceeded {
		return nil, errResultLimitExceededInCompositeMode
	}
	return users, nil
}

// CreateUser creates a new user in the database store only.
// Conflict checking is handled at the service layer.
func (c *compositeUserStore) CreateUser(ctx context.Context, user User, credentials Credentials) error {
	return c.dbStore.CreateUser(ctx, user, credentials)
}

// GetUser retrieves a user by ID from either store.
// Checks database store first, then falls back to file store.
func (c *compositeUserStore) GetUser(ctx context.Context, id string) (User, error) {
	return declarativeresource.CompositeGetHelper(
		func() (User, error) { return c.dbStore.GetUser(ctx, id) },
		func() (User, error) { return c.fileStore.GetUser(ctx, id) },
		ErrUserNotFound,
	)
}

// GetGroupCountForUser retrieves the count of groups for a given user from both stores.
func (c *compositeUserStore) GetGroupCountForUser(ctx context.Context, userID string) (int, error) {
	// Groups are only associated with database users
	return c.dbStore.GetGroupCountForUser(ctx, userID)
}

// GetUserGroups retrieves groups for a given user from both stores.
func (c *compositeUserStore) GetUserGroups(
	ctx context.Context, userID string, limit, offset int,
) ([]UserGroup, error) {
	// Groups are only associated with database users
	return c.dbStore.GetUserGroups(ctx, userID, limit, offset)
}

// UpdateUser updates a user in the database store only.
// Cannot update declarative users.
func (c *compositeUserStore) UpdateUser(ctx context.Context, user *User) error {
	return c.dbStore.UpdateUser(ctx, user)
}

// UpdateUserCredentials updates user credentials in the database store only.
// Cannot update credentials for declarative users.
func (c *compositeUserStore) UpdateUserCredentials(
	ctx context.Context, userID string, credentials Credentials,
) error {
	return c.dbStore.UpdateUserCredentials(ctx, userID, credentials)
}

// DeleteUser deletes a user from the database store only.
// Cannot delete declarative users.
func (c *compositeUserStore) DeleteUser(ctx context.Context, id string) error {
	return c.dbStore.DeleteUser(ctx, id)
}

// IdentifyUser identifies a user with the given filters from either store.
// Checks database store first, then falls back to file store.
func (c *compositeUserStore) IdentifyUser(
	ctx context.Context, filters map[string]interface{},
) (*string, error) {
	// Try database store first
	userID, err := c.dbStore.IdentifyUser(ctx, filters)
	if err == nil {
		return userID, nil
	}
	if !errors.Is(err, ErrUserNotFound) {
		return nil, err
	}

	// Fall back to file store
	return c.fileStore.IdentifyUser(ctx, filters)
}

// GetCredentials retrieves user credentials from either store.
// Checks database store first, then falls back to file store.
func (c *compositeUserStore) GetCredentials(
	ctx context.Context, id string,
) (User, Credentials, error) {
	// Try database store first
	user, credentials, err := c.dbStore.GetCredentials(ctx, id)
	if err == nil {
		return user, credentials, nil
	}
	if !errors.Is(err, ErrUserNotFound) {
		return User{}, nil, err
	}

	// Fall back to file store
	return c.fileStore.GetCredentials(ctx, id)
}

// ValidateUserIDs checks if all provided user IDs exist in either store.
func (c *compositeUserStore) ValidateUserIDs(ctx context.Context, userIDs []string) ([]string, error) {
	invalidIDs := make([]string, 0)

	for _, id := range userIDs {
		_, err := c.GetUser(ctx, id)
		if err != nil {
			if errors.Is(err, ErrUserNotFound) {
				invalidIDs = append(invalidIDs, id)
				continue
			}
			return nil, err
		}
	}

	return invalidIDs, nil
}

// mergeAndDeduplicateUsers merges and deduplicates users from two lists.
// Database users take precedence over file-based users when IDs conflict.
func mergeAndDeduplicateUsers(dbUsers, fileUsers []User) []User {
	userMap := make(map[string]User)

	// Add file users first
	for _, user := range fileUsers {
		userMap[user.ID] = user
	}

	// Add/overwrite with db users (db takes precedence)
	for _, user := range dbUsers {
		userMap[user.ID] = user
	}

	ids := make([]string, 0, len(userMap))
	for id := range userMap {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	result := make([]User, 0, len(ids))
	for _, id := range ids {
		result = append(result, userMap[id])
	}

	return result
}

// IsUserDeclarative checks if a user is declarative (exists in file store) or mutable (DB store only).
func (c *compositeUserStore) IsUserDeclarative(ctx context.Context, id string) (bool, error) {
	// Check if user exists in file store first
	isDeclarative, err := c.fileStore.IsUserDeclarative(ctx, id)
	if err != nil {
		// If not found in file store, check DB (will be mutable)
		if strings.Contains(err.Error(), "not found") {
			isDeclarative, err = c.dbStore.IsUserDeclarative(ctx, id)
			if err != nil && strings.Contains(err.Error(), "not found") {
				// User doesn't exist in either store
				return false, err
			}
			return isDeclarative, err
		}
		return false, err
	}

	// User exists in file store, it's declarative
	return isDeclarative, nil
}
