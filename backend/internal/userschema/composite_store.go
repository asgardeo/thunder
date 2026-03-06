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

package userschema

import (
	"context"

	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
)

// compositeUserSchemaStore implements a composite store that combines file-based (immutable) and
// database (mutable) stores.
// - Read operations query both stores and merge results
// - Write operations (Create/Update/Delete) only affect the database store
// - Declarative user schemas (from YAML files) cannot be modified or deleted
type compositeUserSchemaStore struct {
	fileStore userSchemaStoreInterface
	dbStore   userSchemaStoreInterface
}

// newCompositeUserSchemaStore creates a new composite store with both file-based and database stores.
func newCompositeUserSchemaStore(fileStore, dbStore userSchemaStoreInterface) *compositeUserSchemaStore {
	return &compositeUserSchemaStore{
		fileStore: fileStore,
		dbStore:   dbStore,
	}
}

// GetUserSchemaListCount retrieves the total count of user schemas from both stores.
func (c *compositeUserSchemaStore) GetUserSchemaListCount(ctx context.Context) (int, error) {
	return declarativeresource.CompositeMergeCountHelper(
		func() (int, error) { return c.dbStore.GetUserSchemaListCount(ctx) },
		func() (int, error) { return c.fileStore.GetUserSchemaListCount(ctx) },
	)
}

// GetUserSchemaList retrieves user schemas from both stores with pagination.
// Applies the 1000-record limit in composite mode to prevent memory exhaustion.
// Returns errResultLimitExceededInCompositeMode if the limit is exceeded.
func (c *compositeUserSchemaStore) GetUserSchemaList(
	ctx context.Context, limit, offset int,
) ([]UserSchemaListItem, error) {
	items, limitExceeded, err := declarativeresource.CompositeMergeListHelperWithLimit(
		func() (int, error) { return c.dbStore.GetUserSchemaListCount(ctx) },
		func() (int, error) { return c.fileStore.GetUserSchemaListCount(ctx) },
		func(count int) ([]UserSchemaListItem, error) {
			return c.dbStore.GetUserSchemaList(ctx, count, 0)
		},
		func(count int) ([]UserSchemaListItem, error) {
			return c.fileStore.GetUserSchemaList(ctx, count, 0)
		},
		mergeAndDeduplicateUserSchemas,
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
	return items, nil
}

// GetUserSchemaListCountByOUIDs retrieves the total count of user schemas filtered by OU IDs from both stores.
func (c *compositeUserSchemaStore) GetUserSchemaListCountByOUIDs(ctx context.Context, ouIDs []string) (int, error) {
	return declarativeresource.CompositeMergeCountHelper(
		func() (int, error) { return c.dbStore.GetUserSchemaListCountByOUIDs(ctx, ouIDs) },
		func() (int, error) { return c.fileStore.GetUserSchemaListCountByOUIDs(ctx, ouIDs) },
	)
}

// GetUserSchemaListByOUIDs retrieves user schemas filtered by OU IDs from both stores with pagination.
// Applies the 1000-record limit in composite mode to prevent memory exhaustion.
// Returns errResultLimitExceededInCompositeMode if the limit is exceeded.
func (c *compositeUserSchemaStore) GetUserSchemaListByOUIDs(
	ctx context.Context, ouIDs []string, limit, offset int,
) ([]UserSchemaListItem, error) {
	items, limitExceeded, err := declarativeresource.CompositeMergeListHelperWithLimit(
		func() (int, error) { return c.dbStore.GetUserSchemaListCountByOUIDs(ctx, ouIDs) },
		func() (int, error) { return c.fileStore.GetUserSchemaListCountByOUIDs(ctx, ouIDs) },
		func(count int) ([]UserSchemaListItem, error) {
			return c.dbStore.GetUserSchemaListByOUIDs(ctx, ouIDs, count, 0)
		},
		func(count int) ([]UserSchemaListItem, error) {
			return c.fileStore.GetUserSchemaListByOUIDs(ctx, ouIDs, count, 0)
		},
		mergeAndDeduplicateUserSchemas,
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
	return items, nil
}

// CreateUserSchema creates a new user schema in the database store only.
// Conflict checking is handled at the service layer.
func (c *compositeUserSchemaStore) CreateUserSchema(ctx context.Context, schema UserSchema) error {
	return c.dbStore.CreateUserSchema(ctx, schema)
}

// GetUserSchemaByID retrieves a user schema by ID from either store.
// Checks database store first, then falls back to file store.
func (c *compositeUserSchemaStore) GetUserSchemaByID(ctx context.Context, schemaID string) (UserSchema, error) {
	return declarativeresource.CompositeGetHelper(
		func() (UserSchema, error) { return c.dbStore.GetUserSchemaByID(ctx, schemaID) },
		func() (UserSchema, error) { return c.fileStore.GetUserSchemaByID(ctx, schemaID) },
		ErrUserSchemaNotFound,
	)
}

// GetUserSchemaByName retrieves a user schema by name from either store.
// Checks database store first, then falls back to file store.
func (c *compositeUserSchemaStore) GetUserSchemaByName(ctx context.Context, schemaName string) (UserSchema, error) {
	return declarativeresource.CompositeGetHelper(
		func() (UserSchema, error) { return c.dbStore.GetUserSchemaByName(ctx, schemaName) },
		func() (UserSchema, error) { return c.fileStore.GetUserSchemaByName(ctx, schemaName) },
		ErrUserSchemaNotFound,
	)
}

// UpdateUserSchemaByID updates a user schema in the database store only.
// Immutability checks are handled at the service layer.
func (c *compositeUserSchemaStore) UpdateUserSchemaByID(
	ctx context.Context, schemaID string, schema UserSchema,
) error {
	return c.dbStore.UpdateUserSchemaByID(ctx, schemaID, schema)
}

// DeleteUserSchemaByID deletes a user schema from the database store only.
// Immutability checks are handled at the service layer.
func (c *compositeUserSchemaStore) DeleteUserSchemaByID(ctx context.Context, schemaID string) error {
	return c.dbStore.DeleteUserSchemaByID(ctx, schemaID)
}

// IsUserSchemaDeclarative checks if a user schema is immutable (exists in file store).
func (c *compositeUserSchemaStore) IsUserSchemaDeclarative(schemaID string) bool {
	return declarativeresource.CompositeIsDeclarativeHelper(
		schemaID,
		func(id string) (bool, error) {
			_, err := c.fileStore.GetUserSchemaByID(context.Background(), id)
			if err != nil {
				return false, nil
			}
			return true, nil
		},
	)
}

// mergeAndDeduplicateUserSchemas merges user schemas from both stores and removes duplicates by ID.
// While duplicates shouldn't exist by design (a schema exists in only one store), this provides
// defensive programming against misconfigurations or bugs.
func mergeAndDeduplicateUserSchemas(
	dbSchemas, fileSchemas []UserSchemaListItem,
) []UserSchemaListItem {
	seen := make(map[string]bool)
	result := make([]UserSchemaListItem, 0, len(dbSchemas)+len(fileSchemas))

	// Add DB schemas first (they take precedence) - mark as mutable (isReadOnly=false)
	for i := range dbSchemas {
		if !seen[dbSchemas[i].ID] {
			seen[dbSchemas[i].ID] = true
			schemaCopy := dbSchemas[i]
			schemaCopy.IsReadOnly = false
			result = append(result, schemaCopy)
		}
	}

	// Add file schemas if not already present - mark as immutable (isReadOnly=true)
	for i := range fileSchemas {
		if !seen[fileSchemas[i].ID] {
			seen[fileSchemas[i].ID] = true
			schemaCopy := fileSchemas[i]
			schemaCopy.IsReadOnly = true
			result = append(result, schemaCopy)
		}
	}

	return result
}
