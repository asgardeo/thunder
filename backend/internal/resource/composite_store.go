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

package resource

import (
	"context"
	"errors"
	"math"

	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
)

// compositeResourceStore implements a composite store that combines file-based (immutable) and
// database (mutable) stores.
// - Read operations query both stores and merge results
// - Write operations (Create/Update/Delete) only affect the database store
// - Declarative resource servers (from YAML files) cannot be modified or deleted
type compositeResourceStore struct {
	fileStore resourceStoreInterface
	dbStore   resourceStoreInterface
}

// newCompositeResourceStore creates a new composite store with both file-based and database stores.
func newCompositeResourceStore(fileStore, dbStore resourceStoreInterface) *compositeResourceStore {
	return &compositeResourceStore{
		fileStore: fileStore,
		dbStore:   dbStore,
	}
}

// Resource Server operations

func (c *compositeResourceStore) CreateResourceServer(ctx context.Context, id string, rs ResourceServer) error {
	return c.dbStore.CreateResourceServer(ctx, id, rs)
}

func (c *compositeResourceStore) GetResourceServer(ctx context.Context, id string) (ResourceServer, error) {
	// Try DB store first
	server, err := c.dbStore.GetResourceServer(ctx, id)
	if err == nil {
		server.IsReadOnly = false
		return server, nil
	}

	// If not found in DB, try file store
	if errors.Is(err, errResourceServerNotFound) {
		server, err := c.fileStore.GetResourceServer(ctx, id)
		if err == nil {
			server.IsReadOnly = true
			return server, nil
		}
		return ResourceServer{}, err
	}

	return ResourceServer{}, err
}

func (c *compositeResourceStore) GetResourceServerList(
	ctx context.Context, limit, offset int) ([]ResourceServer, error) {
	// Fetch all servers from both stores using math.MaxInt to ensure no LIMIT 0 issue
	dbServers, err := c.dbStore.GetResourceServerList(ctx, math.MaxInt, 0)
	if err != nil {
		return nil, err
	}

	fileServers, err := c.fileStore.GetResourceServerList(ctx, math.MaxInt, 0)
	if err != nil {
		return nil, err
	}

	// Merge and deduplicate
	merged := mergeAndDeduplicateResourceServers(dbServers, fileServers)

	// Apply pagination
	start := offset
	end := offset + limit
	if start > len(merged) {
		return []ResourceServer{}, nil
	}
	if end > len(merged) {
		end = len(merged)
	}

	return merged[start:end], nil
}

func (c *compositeResourceStore) GetResourceServerListCount(ctx context.Context) (int, error) {
	// Fetch all servers from both stores using math.MaxInt to avoid LIMIT issues
	dbServers, err := c.dbStore.GetResourceServerList(ctx, math.MaxInt, 0)
	if err != nil {
		return 0, err
	}

	fileServers, err := c.fileStore.GetResourceServerList(ctx, math.MaxInt, 0)
	if err != nil {
		return 0, err
	}

	// Merge and deduplicate, then return count
	merged := mergeAndDeduplicateResourceServers(dbServers, fileServers)
	return len(merged), nil
}

func (c *compositeResourceStore) UpdateResourceServer(ctx context.Context, id string, rs ResourceServer) error {
	return c.dbStore.UpdateResourceServer(ctx, id, rs)
}

func (c *compositeResourceStore) DeleteResourceServer(ctx context.Context, id string) error {
	return c.dbStore.DeleteResourceServer(ctx, id)
}

func (c *compositeResourceStore) CheckResourceServerNameExists(ctx context.Context, name string) (bool, error) {
	return declarativeresource.CompositeBooleanCheckHelper(
		func() (bool, error) { return c.fileStore.CheckResourceServerNameExists(ctx, name) },
		func() (bool, error) { return c.dbStore.CheckResourceServerNameExists(ctx, name) },
	)
}

func (c *compositeResourceStore) CheckResourceServerIdentifierExists(
	ctx context.Context, identifier string) (bool, error) {
	return declarativeresource.CompositeBooleanCheckHelper(
		func() (bool, error) { return c.fileStore.CheckResourceServerIdentifierExists(ctx, identifier) },
		func() (bool, error) { return c.dbStore.CheckResourceServerIdentifierExists(ctx, identifier) },
	)
}

func (c *compositeResourceStore) CheckResourceServerHasDependencies(
	ctx context.Context, resServerID string) (bool, error) {
	// Check in DB store first
	hasDeps, err := c.dbStore.CheckResourceServerHasDependencies(ctx, resServerID)
	if err != nil {
		return false, err
	}
	if hasDeps {
		return true, nil
	}

	// Also check in file store
	return c.fileStore.CheckResourceServerHasDependencies(ctx, resServerID)
}

func (c *compositeResourceStore) IsResourceServerDeclarative(id string) bool {
	return declarativeresource.CompositeIsDeclarativeHelper(
		id,
		func(id string) (bool, error) {
			_, err := c.fileStore.GetResourceServer(context.Background(), id)
			return err == nil, nil
		},
	)
}

// Resource operations

func (c *compositeResourceStore) CreateResource(
	ctx context.Context, uuid string, resServerID string, parentID *string, res Resource) error {
	return c.dbStore.CreateResource(ctx, uuid, resServerID, parentID, res)
}

func (c *compositeResourceStore) GetResource(
	ctx context.Context, id string, resServerID string) (Resource, error) {
	// Try DB store first
	resource, err := c.dbStore.GetResource(ctx, id, resServerID)
	if err == nil {
		return resource, nil
	}

	// If not found in DB, try file store
	if errors.Is(err, errResourceNotFound) {
		return c.fileStore.GetResource(ctx, id, resServerID)
	}

	return Resource{}, err
}

func (c *compositeResourceStore) GetResourceList(
	ctx context.Context, resServerID string, limit, offset int) ([]Resource, error) {
	// Fetch all resources from both stores using math.MaxInt to ensure no LIMIT 0 issue
	dbResources, err := c.dbStore.GetResourceList(ctx, resServerID, math.MaxInt, 0)
	if err != nil {
		return nil, err
	}

	fileResources, err := c.fileStore.GetResourceList(ctx, resServerID, math.MaxInt, 0)
	if err != nil {
		return nil, err
	}

	// Merge and deduplicate
	merged := mergeAndDeduplicateResources(dbResources, fileResources)

	// Apply pagination
	start := offset
	end := offset + limit
	if start > len(merged) {
		return []Resource{}, nil
	}
	if end > len(merged) {
		end = len(merged)
	}

	return merged[start:end], nil
}

func (c *compositeResourceStore) GetResourceListByParent(
	ctx context.Context, resServerID string, parentID *string, limit, offset int,
) ([]Resource, error) {
	// Fetch all resources from both stores using math.MaxInt to ensure no LIMIT 0 issue
	dbResources, err := c.dbStore.GetResourceListByParent(ctx, resServerID, parentID, math.MaxInt, 0)
	if err != nil {
		return nil, err
	}

	fileResources, err := c.fileStore.GetResourceListByParent(
		ctx, resServerID, parentID, math.MaxInt, 0)
	if err != nil {
		return nil, err
	}

	// Merge and deduplicate
	merged := mergeAndDeduplicateResources(dbResources, fileResources)

	// Apply pagination
	start := offset
	end := offset + limit
	if start > len(merged) {
		return []Resource{}, nil
	}
	if end > len(merged) {
		end = len(merged)
	}

	return merged[start:end], nil
}

func (c *compositeResourceStore) GetResourceListCount(ctx context.Context, resServerID string) (int, error) {
	// Fetch all resources from both stores using math.MaxInt to avoid LIMIT issues
	dbResources, err := c.dbStore.GetResourceList(ctx, resServerID, math.MaxInt, 0)
	if err != nil {
		return 0, err
	}

	fileResources, err := c.fileStore.GetResourceList(ctx, resServerID, math.MaxInt, 0)
	if err != nil {
		return 0, err
	}

	// Merge and deduplicate, then return count
	merged := mergeAndDeduplicateResources(dbResources, fileResources)
	return len(merged), nil
}

func (c *compositeResourceStore) GetResourceListCountByParent(
	ctx context.Context, resServerID string, parentID *string) (int, error) {
	// Fetch all resources from both stores using math.MaxInt to avoid LIMIT issues
	dbResources, err := c.dbStore.GetResourceListByParent(ctx, resServerID, parentID, math.MaxInt, 0)
	if err != nil {
		return 0, err
	}

	fileResources, err := c.fileStore.GetResourceListByParent(
		ctx, resServerID, parentID, math.MaxInt, 0)
	if err != nil {
		return 0, err
	}

	// Merge and deduplicate, then return count
	merged := mergeAndDeduplicateResources(dbResources, fileResources)
	return len(merged), nil
}

func (c *compositeResourceStore) UpdateResource(
	ctx context.Context, id string, resServerID string, res Resource) error {
	return c.dbStore.UpdateResource(ctx, id, resServerID, res)
}

func (c *compositeResourceStore) DeleteResource(
	ctx context.Context, id string, resServerID string) error {
	return c.dbStore.DeleteResource(ctx, id, resServerID)
}

func (c *compositeResourceStore) CheckResourceHandleExists(
	ctx context.Context, resServerID string, handle string, parentID *string,
) (bool, error) {
	return declarativeresource.CompositeBooleanCheckHelper(
		func() (bool, error) {
			return c.fileStore.CheckResourceHandleExists(ctx, resServerID, handle, parentID)
		},
		func() (bool, error) {
			return c.dbStore.CheckResourceHandleExists(ctx, resServerID, handle, parentID)
		},
	)
}

func (c *compositeResourceStore) CheckResourceHasDependencies(ctx context.Context, resID string) (bool, error) {
	return c.dbStore.CheckResourceHasDependencies(ctx, resID)
}

func (c *compositeResourceStore) CheckCircularDependency(
	ctx context.Context, resourceID, newParentID string) (bool, error) {
	return c.dbStore.CheckCircularDependency(ctx, resourceID, newParentID)
}

// Action operations

func (c *compositeResourceStore) CreateAction(
	ctx context.Context, uuid string, resServerID string, resID *string, action Action) error {
	return c.dbStore.CreateAction(ctx, uuid, resServerID, resID, action)
}

func (c *compositeResourceStore) GetAction(
	ctx context.Context, id string, resServerID string, resID *string) (Action, error) {
	action, err := declarativeresource.CompositeGetHelper(
		func() (Action, error) { return c.dbStore.GetAction(ctx, id, resServerID, resID) },
		func() (Action, error) { return c.fileStore.GetAction(ctx, id, resServerID, resID) },
		errActionNotFound,
	)
	return action, err
}

func (c *compositeResourceStore) GetActionList(
	ctx context.Context, resServerID string, resID *string, limit, offset int) ([]Action, error) {
	// Fetch all actions from both stores using math.MaxInt to ensure no LIMIT 0 issue
	dbActions, err := c.dbStore.GetActionList(ctx, resServerID, resID, math.MaxInt, 0)
	if err != nil {
		return nil, err
	}

	fileActions, err := c.fileStore.GetActionList(ctx, resServerID, resID, math.MaxInt, 0)
	if err != nil {
		return nil, err
	}

	// Merge and deduplicate
	merged := mergeAndDeduplicateActions(dbActions, fileActions)

	// Apply pagination
	start := offset
	end := offset + limit
	if start > len(merged) {
		return []Action{}, nil
	}
	if end > len(merged) {
		end = len(merged)
	}

	return merged[start:end], nil
}

func (c *compositeResourceStore) GetActionListCount(
	ctx context.Context, resServerID string, resID *string) (int, error) {
	// Fetch all actions from both stores using math.MaxInt to avoid LIMIT issues
	dbActions, err := c.dbStore.GetActionList(ctx, resServerID, resID, math.MaxInt, 0)
	if err != nil {
		return 0, err
	}

	fileActions, err := c.fileStore.GetActionList(ctx, resServerID, resID, math.MaxInt, 0)
	if err != nil {
		return 0, err
	}

	// Merge and deduplicate, then return count
	merged := mergeAndDeduplicateActions(dbActions, fileActions)
	return len(merged), nil
}

func (c *compositeResourceStore) UpdateAction(
	ctx context.Context, id string, resServerID string, resID *string, action Action) error {
	return c.dbStore.UpdateAction(ctx, id, resServerID, resID, action)
}

func (c *compositeResourceStore) DeleteAction(
	ctx context.Context, id string, resServerID string, resID *string) error {
	return c.dbStore.DeleteAction(ctx, id, resServerID, resID)
}

func (c *compositeResourceStore) IsActionExist(
	ctx context.Context, id string, resServerID string, resID *string) (bool, error) {
	return declarativeresource.CompositeBooleanCheckHelper(
		func() (bool, error) { return c.fileStore.IsActionExist(ctx, id, resServerID, resID) },
		func() (bool, error) { return c.dbStore.IsActionExist(ctx, id, resServerID, resID) },
	)
}

func (c *compositeResourceStore) CheckActionHandleExists(
	ctx context.Context, resServerID string, resID *string, handle string,
) (bool, error) {
	return declarativeresource.CompositeBooleanCheckHelper(
		func() (bool, error) {
			return c.fileStore.CheckActionHandleExists(ctx, resServerID, resID, handle)
		},
		func() (bool, error) {
			return c.dbStore.CheckActionHandleExists(ctx, resServerID, resID, handle)
		},
	)
}

func (c *compositeResourceStore) ValidatePermissions(
	ctx context.Context, resServerID string, permissions []string) ([]string, error) {
	// Call db store
	dbInvalid, err := c.dbStore.ValidatePermissions(ctx, resServerID, permissions)
	if err != nil {
		return nil, err
	}

	// Call file store (declarative store)
	fileInvalid, err := c.fileStore.ValidatePermissions(ctx, resServerID, permissions)
	if err != nil {
		return nil, err
	}

	// Create set of file invalid permissions for efficient lookup
	fileInvalidSet := make(map[string]struct{})
	for _, perm := range fileInvalid {
		fileInvalidSet[perm] = struct{}{}
	}

	// Return only permissions that are invalid in both stores (intersection)
	// A permission is valid if present in either store
	var result []string
	for _, perm := range dbInvalid {
		if _, ok := fileInvalidSet[perm]; ok {
			result = append(result, perm)
		}
	}

	return result, nil
}

func mergeAndDeduplicateResourceServers(dbServers, fileServers []ResourceServer) []ResourceServer {
	seen := make(map[string]bool)
	result := make([]ResourceServer, 0, len(dbServers)+len(fileServers))

	// Add DB servers first (they take precedence) - mark as mutable (isReadOnly=false)
	for i := range dbServers {
		if !seen[dbServers[i].ID] {
			seen[dbServers[i].ID] = true
			dbServers[i].IsReadOnly = false
			result = append(result, dbServers[i])
		}
	}

	// Add file servers if not already present - mark as immutable (isReadOnly=true)
	for i := range fileServers {
		if !seen[fileServers[i].ID] {
			seen[fileServers[i].ID] = true
			fileServers[i].IsReadOnly = true
			result = append(result, fileServers[i])
		}
	}

	return result
}

func mergeAndDeduplicateResources(dbResources, fileResources []Resource) []Resource {
	seen := make(map[string]bool)
	result := make([]Resource, 0, len(dbResources)+len(fileResources))

	// Add DB resources first
	for i := range dbResources {
		if !seen[dbResources[i].ID] {
			seen[dbResources[i].ID] = true
			result = append(result, dbResources[i])
		}
	}

	// Add file resources if not already present
	for i := range fileResources {
		if !seen[fileResources[i].ID] {
			seen[fileResources[i].ID] = true
			result = append(result, fileResources[i])
		}
	}

	return result
}

func mergeAndDeduplicateActions(dbActions, fileActions []Action) []Action {
	seen := make(map[string]bool)
	result := make([]Action, 0, len(dbActions)+len(fileActions))

	// Add DB actions first
	for i := range dbActions {
		if !seen[dbActions[i].ID] {
			seen[dbActions[i].ID] = true
			result = append(result, dbActions[i])
		}
	}

	// Add file actions if not already present
	for i := range fileActions {
		if !seen[fileActions[i].ID] {
			seen[fileActions[i].ID] = true
			result = append(result, fileActions[i])
		}
	}

	return result
}
