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

package declarativeresource

import (
	"errors"
	"fmt"
)

// CompositeGetHelper implements the dual-store Get pattern for composite stores.
//
// Type Parameter:
//   - T: The domain entity type (e.g., OrganizationUnit, Application, IDPDTO)
//
// Parameters:
//   - dbGetter: Function that retrieves from database store
//   - fileGetter: Function that retrieves from file store
//   - notFoundError: Domain-specific "not found" error to return
//
// Returns:
//   - Entity of type T
//   - Error if not found in either store or other errors occur
//
// Behavior:
//  1. Try database store first (runtime/mutable resources)
//  2. If found, return it
//  3. If not found, try file store (YAML/declarative resources)
//  4. If found in file, return it
//  5. If not found in either, return notFoundError
//
// Usage:
//   - GetByID: CompositeGetHelper(func() { return db.Get(id) }, func() { return file.Get(id) }, err)
//   - GetByPath: CompositeGetHelper(func() { return db.GetByPath(p) }, ...)
//   - GetByName: CompositeGetHelper(func() { return db.GetByName(n) }, ...)
func CompositeGetHelper[T any](
	dbGetter func() (T, error),
	fileGetter func() (T, error),
	notFoundError error,
) (T, error) {
	var zero T // Zero value for type T

	// Phase 1: Check database store (mutable resources)
	entity, err := dbGetter()
	if err == nil {
		return entity, nil
	}

	// Phase 2: If not found in DB, check file store (declarative resources)
	if errors.Is(err, notFoundError) {
		entity, fileErr := fileGetter()
		if fileErr == nil {
			return entity, nil
		}
		// Not found in either store
		return zero, notFoundError
	}

	// Database error (not "not found")
	return zero, err
}

// CompositeCreateHelper implements Create operation with immutable conflict checking.
//
// Type Parameter:
//   - T: The domain entity type
//
// Parameters:
//   - entity: The entity to create
//   - getID: Function to extract ID from entity
//   - fileExists: Function to check if ID exists in file store
//   - dbCreate: Function to create entity in database store
//
// Returns:
//   - Error if ID conflicts with immutable resource or creation fails
//
// Behavior:
//  1. Extract ID from entity
//  2. Check if ID exists in file store (immutable)
//  3. If exists in file store, return conflict error
//  4. If error occurs during check, return that error
//  5. If not found, proceed with creation
//
// This ensures you cannot create a runtime resource with the same ID
// as an immutable resource, and properly handles file store errors.
func CompositeCreateHelper[T any](
	entity T,
	getID func(T) string,
	fileExists func(string) (bool, error),
	dbCreate func(T) error,
) error {
	id := getID(entity)

	// Check if ID conflicts with immutable resource
	exists, err := fileExists(id)
	if err != nil {
		return err
	}
	if exists {
		// Resource exists in file store - conflict
		return fmt.Errorf("resource with ID %s already exists as declarative resource", id)
	}

	// Resource not found in file store - safe to create in database
	return dbCreate(entity)
}

// CompositeUpdateHelper blocks updates to declarative resources.
//
// Type Parameter:
//   - T: The domain entity type
//
// Parameters:
//   - entity: The entity to update
//   - getID: Function to extract ID from entity
//   - fileExists: Function to check if ID exists in file store
//   - dbUpdate: Function to update entity in database store
//   - immutableError: Error to return when trying to update immutable resource
//
// Returns:
//   - Error if resource is immutable or update fails
//
// Behavior:
//  1. Extract ID from entity
//  2. Check if exists in file store
//  3. If immutable, return immutableError
//  4. If error occurs during check, return that error
//  5. If not found, proceed with update
func CompositeUpdateHelper[T any](
	entity T,
	getID func(T) string,
	fileExists func(string) (bool, error),
	dbUpdate func(T) error,
	immutableError error,
) error {
	id := getID(entity)

	// Check if it's an immutable resource
	exists, err := fileExists(id)
	if err != nil {
		return err
	}
	if exists {
		// Resource exists in file store - immutable
		return immutableError
	}

	// Resource not found in file store - safe to update in database
	return dbUpdate(entity)
}

// CompositeDeleteHelper blocks deletion of declarative resources.
//
// Parameters:
//   - id: The resource ID to delete
//   - fileExists: Function to check if ID exists in file store
//   - dbDelete: Function to delete from database store
//   - immutableError: Error to return when trying to delete immutable resource
//
// Returns:
//   - Error if resource is immutable or deletion fails
//
// Behavior:
//  1. Check if exists in file store
//  2. If immutable, return immutableError
//  3. If error occurs during check, return that error
//  4. If not found, proceed with deletion
func CompositeDeleteHelper(
	id string,
	fileExists func(string) (bool, error),
	dbDelete func(string) error,
	immutableError error,
) error {
	// Check if it's an immutable resource
	exists, err := fileExists(id)
	if err != nil {
		return err
	}
	if exists {
		// Resource exists in file store - immutable
		return immutableError
	}

	// Resource not found in file store - safe to delete from database
	return dbDelete(id)
}

// CompositeBooleanCheckHelper performs a boolean check across both stores.
//
// Generic helper for any boolean check operation (existence, conflict, has-children, etc.)
//
// Parameters:
//   - fileChecker: Function to check in file store
//   - dbChecker: Function to check in database store
//
// Returns:
//   - true if check succeeds in either store, false otherwise
//   - Error if check fails
//
// Behavior:
//  1. Check file store first (declarative resources)
//  2. If true, return true immediately
//  3. If error occurs during file check, return that error
//  4. Otherwise, check database store (mutable resources)
//  5. Return database check result
//
// Usage:
//   - Existence check: CompositeBooleanCheckHelper(func() { return file.Exists(id) }, func() { return db.Exists(id) })
//   - Conflict check: CompositeBooleanCheckHelper(func() { return file.HasConflict(name) },
//     func() { return db.HasConflict(name) })
//   - Has children: CompositeBooleanCheckHelper(func() { return file.HasChildren(id) },
//     func() { return db.HasChildren(id) })
func CompositeBooleanCheckHelper(
	fileChecker func() (bool, error),
	dbChecker func() (bool, error),
) (bool, error) {
	// Check file store first
	fileResult, err := fileChecker()
	if err != nil {
		return false, err
	}
	if fileResult {
		return true, nil
	}

	// Check database store
	return dbChecker()
}

// CompositeMergeCountHelper merges counts from both stores.
//
// Parameters:
//   - dbCounter: Function to get count from database store
//   - fileCounter: Function to get count from file store
//
// Returns:
//   - Sum of counts from both stores
//   - Error if either count operation fails
//
// Behavior:
//  1. Get count from database store
//  2. Get count from file store
//  3. Return sum of both counts
func CompositeMergeCountHelper(
	dbCounter func() (int, error),
	fileCounter func() (int, error),
) (int, error) {
	dbCount, err := dbCounter()
	if err != nil {
		return 0, err
	}

	fileCount, err := fileCounter()
	if err != nil {
		return 0, err
	}

	return dbCount + fileCount, nil
}

// CompositeMergeListHelper retrieves and merges lists from both stores with pagination.
//
// Parameters:
//   - firstCounter: Function to get count from first store (typically file store)
//   - secondCounter: Function to get count from second store (typically DB store)
//   - firstFetcher: Function to fetch items from first store
//   - secondFetcher: Function to fetch items from second store
//   - merger: Function to merge and deduplicate the two lists
//   - limit: Maximum number of items to return
//   - offset: Number of items to skip
//
// Returns:
//   - Merged and paginated list
//   - Error if any operation fails
//
// Behavior:
//  1. Get counts from both stores
//  2. If offset exceeds total count, return empty list
//  3. Fetch all items from both stores (acceptable for limited root-level resources)
//  4. Merge and deduplicate using provided merger function
//  5. Apply pagination to merged results
func CompositeMergeListHelper[T any](
	firstCounter func() (int, error),
	secondCounter func() (int, error),
	firstFetcher func(count int) ([]T, error),
	secondFetcher func(count int) ([]T, error),
	merger func([]T, []T) []T,
	limit int,
	offset int,
) ([]T, error) {
	// Validate limit and offset
	if limit < 0 {
		return nil, fmt.Errorf("limit must be non-negative, got %d", limit)
	}
	if offset < 0 {
		return nil, fmt.Errorf("offset must be non-negative, got %d", offset)
	}

	// Get counts from both stores
	firstCount, err := firstCounter()
	if err != nil {
		return nil, err
	}

	secondCount, err := secondCounter()
	if err != nil {
		return nil, err
	}

	totalCount := firstCount + secondCount

	// If offset is beyond total count, return empty
	if offset >= totalCount {
		return []T{}, nil
	}

	// Fetch all items from both stores
	firstItems, err := firstFetcher(firstCount)
	if err != nil {
		return nil, err
	}

	secondItems, err := secondFetcher(secondCount)
	if err != nil {
		return nil, err
	}

	// Merge and deduplicate
	allItems := merger(firstItems, secondItems)

	// Apply pagination to merged results
	if offset >= len(allItems) {
		return []T{}, nil
	}

	end := offset + limit
	if end > len(allItems) {
		end = len(allItems)
	}

	return allItems[offset:end], nil
}

// CompositeHasChildrenHelper checks if a resource has children in either store.
//
// Parameters:
//   - fileChecker: Function to check if resource has children in file store
//   - dbChecker: Function to check if resource has children in database store
//
// Returns:
//   - true if resource has children in either store
//   - Error if check fails
func CompositeHasChildrenHelper(
	fileChecker func() (bool, error),
	dbChecker func() (bool, error),
) (bool, error) {
	return CompositeBooleanCheckHelper(fileChecker, dbChecker)
}

// CompositeIsDeclarativeHelper checks if a resource is immutable (exists in file store).
//
// Parameters:
//   - id: The resource ID to check
//   - fileExists: Function to check if ID exists in file store
//
// Returns:
//   - true if resource exists in file store (immutable), false otherwise
//
// Behavior:
//  1. Check if ID exists in file store
//  2. If exists, return true (immutable)
//  3. Otherwise, return false (mutable or doesn't exist)
func CompositeIsDeclarativeHelper(
	id string,
	fileExists func(string) (bool, error),
) bool {
	exists, err := fileExists(id)
	return err == nil && exists
}

// CompositeMergeListHelperWithLimit retrieves and merges lists from both stores with pagination and hard limit.
// This implements a "scatter-gather" pattern optimized for composite store operations with a maximum record cap.
//
// Type Parameter:
//   - T: The domain entity type
//
// Parameters:
//   - firstCounter: Function to get count from first store (typically DB store)
//   - secondCounter: Function to get count from second store (typically file store)
//   - firstFetcher: Function to fetch items from first store, accepts limit parameter (typically DB store)
//   - secondFetcher: Function to fetch items from second store, accepts limit parameter (typically file store)
//   - merger: Function to merge and deduplicate the two lists
//   - limit: Maximum number of items to return
//   - offset: Number of items to skip
//   - maxRecords: Hard limit on total records to fetch (0 = no limit, typically 1000 for hybrid mode)
//
// Returns:
//   - Merged and paginated list
//   - limitExceeded: true if total count exceeds maxRecords
//   - Error if any operation fails
//
// Behavior:
//  1. Get counts from both stores
//  2. Calculate effective total count (capped at maxRecords if > 0)
//  3. If offset exceeds effective count, return empty list
//  4. Calculate depth = offset + limit (how many items needed from start)
//  5. Fetch min(depth, effectiveCount, storeCount) from each store (scatter-gather)
//  6. Merge and deduplicate using provided merger function
//  7. Apply pagination (slice offset:offset+limit)
//  8. Return limitExceeded=true if actual total > maxRecords
//
// Note: Results are returned in the default order from each store (typically by name).
//
// Example: offset=20, limit=10, store1=50, store2=100, maxRecords=1000
//   - depth = 30, fetch 30 from each store
//   - Merge (up to 60 items), slice [20:30]
//
// Example: offset=900, limit=100, store1=500, store2=700, maxRecords=1000
//   - effectiveTotal = 1000 (capped), depth = 1000
//   - Fetch 500 from store1, 500 from store2 (capped at store counts)
//   - Merge (up to 1000), slice [900:1000]
//   - Return limitExceeded=true (actual 1200 > 1000)
func CompositeMergeListHelperWithLimit[T any](
	firstCounter func() (int, error),
	secondCounter func() (int, error),
	firstFetcher func(limit int) ([]T, error),
	secondFetcher func(limit int) ([]T, error),
	merger func([]T, []T) []T,
	limit int,
	offset int,
	maxRecords int,
) ([]T, bool, error) {
	// Validate limit and offset
	if limit < 0 {
		return nil, false, fmt.Errorf("limit must be non-negative, got %d", limit)
	}
	if offset < 0 {
		return nil, false, fmt.Errorf("offset must be non-negative, got %d", offset)
	}

	// Get counts from both stores
	firstCount, err := firstCounter()
	if err != nil {
		return nil, false, err
	}

	secondCount, err := secondCounter()
	if err != nil {
		return nil, false, err
	}

	actualTotal := firstCount + secondCount
	limitExceeded := false

	// Apply maxRecords cap if specified
	effectiveTotal := actualTotal
	if maxRecords > 0 && actualTotal > maxRecords {
		effectiveTotal = maxRecords
		limitExceeded = true
	}

	// If offset is beyond effective total, return empty
	if offset >= effectiveTotal {
		return []T{}, limitExceeded, nil
	}

	// Calculate depth: how many items we need from the start to fulfill offset+limit
	depth := offset + limit
	if depth > effectiveTotal {
		depth = effectiveTotal
	}

	// Determine how many items to fetch from each store
	// Strategy: Fetch up to 'depth' from each store to ensure we have enough for merging
	// This is necessary because we don't know the merge order until we combine results
	firstFetchLimit := depth
	if firstFetchLimit > firstCount {
		firstFetchLimit = firstCount
	}

	secondFetchLimit := depth
	if secondFetchLimit > secondCount {
		secondFetchLimit = secondCount
	}

	// Scatter: Fetch items from both stores
	firstItems, err := firstFetcher(firstFetchLimit)
	if err != nil {
		return nil, limitExceeded, err
	}

	secondItems, err := secondFetcher(secondFetchLimit)
	if err != nil {
		return nil, limitExceeded, err
	}

	// Gather: Merge and deduplicate
	allItems := merger(firstItems, secondItems)

	// Apply pagination to merged results
	if offset >= len(allItems) {
		return []T{}, limitExceeded, nil
	}

	end := offset + limit
	if end > len(allItems) {
		end = len(allItems)
	}

	return allItems[offset:end], limitExceeded, nil
}
