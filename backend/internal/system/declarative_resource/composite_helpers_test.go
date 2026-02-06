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
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test entity for generic testing
type compositeTestEntity struct {
	ID   string
	Name string
}

var (
	errNotFound    = errors.New("not found")
	errDeclarative = errors.New("declarative resource")
	errIOError     = errors.New("I/O error")
	errParseError  = errors.New("parse error")
)

// TestCompositeGetHelper tests the CompositeGetHelper function.
func TestCompositeGetHelper(t *testing.T) {
	t.Run("retrieves from DB store", func(t *testing.T) {
		id := "test-id"
		dbGetter := func() (compositeTestEntity, error) {
			return compositeTestEntity{ID: id, Name: "DB Entity"}, nil
		}
		fileGetter := func() (compositeTestEntity, error) {
			return compositeTestEntity{}, errNotFound
		}

		result, err := CompositeGetHelper(dbGetter, fileGetter, errNotFound)
		assert.NoError(t, err)
		assert.Equal(t, "test-id", result.ID)
		assert.Equal(t, "DB Entity", result.Name)
	})

	t.Run("falls back to file store when not in DB", func(t *testing.T) {
		id := "test-id"
		dbGetter := func() (compositeTestEntity, error) {
			return compositeTestEntity{}, errNotFound
		}
		fileGetter := func() (compositeTestEntity, error) {
			return compositeTestEntity{ID: id, Name: "File Entity"}, nil
		}

		result, err := CompositeGetHelper(dbGetter, fileGetter, errNotFound)
		assert.NoError(t, err)
		assert.Equal(t, "test-id", result.ID)
		assert.Equal(t, "File Entity", result.Name)
	})

	t.Run("returns not found when in neither store", func(t *testing.T) {
		dbGetter := func() (compositeTestEntity, error) {
			return compositeTestEntity{}, errNotFound
		}
		fileGetter := func() (compositeTestEntity, error) {
			return compositeTestEntity{}, errNotFound
		}

		result, err := CompositeGetHelper(dbGetter, fileGetter, errNotFound)
		assert.Error(t, err)
		assert.Equal(t, errNotFound, err)
		assert.Empty(t, result.ID)
	})

	t.Run("propagates DB errors other than not found", func(t *testing.T) {
		dbErr := errors.New("database connection error")
		dbGetter := func() (compositeTestEntity, error) {
			return compositeTestEntity{}, dbErr
		}
		fileGetter := func() (compositeTestEntity, error) {
			return compositeTestEntity{}, errNotFound
		}

		result, err := CompositeGetHelper(dbGetter, fileGetter, errNotFound)
		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.Empty(t, result.ID)
	})
}

// TestCompositeCreateHelper tests the CompositeCreateHelper function.
func TestCompositeCreateHelper(t *testing.T) {
	t.Run("creates when ID not in file store", func(t *testing.T) {
		entity := compositeTestEntity{ID: "new-1", Name: "New Entity"}
		created := false

		getID := func(e compositeTestEntity) string { return e.ID }
		fileExists := func(id string) (bool, error) {
			return false, nil
		}
		dbCreate := func(e compositeTestEntity) error {
			created = true
			return nil
		}

		err := CompositeCreateHelper(entity, getID, fileExists, dbCreate)
		assert.NoError(t, err)
		assert.True(t, created)
	})

	t.Run("fails when ID exists in file store", func(t *testing.T) {
		entity := compositeTestEntity{ID: "declarative-1", Name: "Trying to create"}
		created := false

		getID := func(e compositeTestEntity) string { return e.ID }
		fileExists := func(id string) (bool, error) {
			return true, nil
		}
		dbCreate := func(e compositeTestEntity) error {
			created = true
			return nil
		}

		err := CompositeCreateHelper(entity, getID, fileExists, dbCreate)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists as declarative resource")
		assert.False(t, created)
	})

	t.Run("propagates file store I/O error", func(t *testing.T) {
		entity := compositeTestEntity{ID: "new-1", Name: "New Entity"}
		created := false

		getID := func(e compositeTestEntity) string { return e.ID }
		fileExists := func(id string) (bool, error) {
			return false, errIOError
		}
		dbCreate := func(e compositeTestEntity) error {
			created = true
			return nil
		}

		err := CompositeCreateHelper(entity, getID, fileExists, dbCreate)
		assert.Error(t, err)
		assert.Equal(t, errIOError, err)
		assert.False(t, created)
	})

	t.Run("propagates file store parse error", func(t *testing.T) {
		entity := compositeTestEntity{ID: "new-1", Name: "New Entity"}
		created := false

		getID := func(e compositeTestEntity) string { return e.ID }
		fileExists := func(id string) (bool, error) {
			return false, errParseError
		}
		dbCreate := func(e compositeTestEntity) error {
			created = true
			return nil
		}

		err := CompositeCreateHelper(entity, getID, fileExists, dbCreate)
		assert.Error(t, err)
		assert.Equal(t, errParseError, err)
		assert.False(t, created)
	})
}

// TestCompositeUpdateHelper tests the CompositeUpdateHelper function.
func TestCompositeUpdateHelper(t *testing.T) {
	t.Run("updates when not in file store", func(t *testing.T) {
		entity := compositeTestEntity{ID: "db-1", Name: "Updated Entity"}
		updated := false

		getID := func(e compositeTestEntity) string { return e.ID }
		fileExists := func(id string) (bool, error) {
			return false, nil
		}
		dbUpdate := func(e compositeTestEntity) error {
			updated = true
			return nil
		}

		err := CompositeUpdateHelper(entity, getID, fileExists, dbUpdate, errDeclarative)
		assert.NoError(t, err)
		assert.True(t, updated)
	})

	t.Run("fails when trying to update declarative resource", func(t *testing.T) {
		entity := compositeTestEntity{ID: "declarative-1", Name: "Trying to update"}
		updated := false

		getID := func(e compositeTestEntity) string { return e.ID }
		fileExists := func(id string) (bool, error) {
			return true, nil
		}
		dbUpdate := func(e compositeTestEntity) error {
			updated = true
			return nil
		}

		err := CompositeUpdateHelper(entity, getID, fileExists, dbUpdate, errDeclarative)
		assert.Error(t, err)
		assert.Equal(t, errDeclarative, err)
		assert.False(t, updated)
	})

	t.Run("propagates file store I/O error", func(t *testing.T) {
		entity := compositeTestEntity{ID: "db-1", Name: "Updated Entity"}
		updated := false

		getID := func(e compositeTestEntity) string { return e.ID }
		fileExists := func(id string) (bool, error) {
			return false, errIOError
		}
		dbUpdate := func(e compositeTestEntity) error {
			updated = true
			return nil
		}

		err := CompositeUpdateHelper(entity, getID, fileExists, dbUpdate, errDeclarative)
		assert.Error(t, err)
		assert.Equal(t, errIOError, err)
		assert.False(t, updated)
	})

	t.Run("propagates file store parse error", func(t *testing.T) {
		entity := compositeTestEntity{ID: "db-1", Name: "Updated Entity"}
		updated := false

		getID := func(e compositeTestEntity) string { return e.ID }
		fileExists := func(id string) (bool, error) {
			return false, errParseError
		}
		dbUpdate := func(e compositeTestEntity) error {
			updated = true
			return nil
		}

		err := CompositeUpdateHelper(entity, getID, fileExists, dbUpdate, errDeclarative)
		assert.Error(t, err)
		assert.Equal(t, errParseError, err)
		assert.False(t, updated)
	})
}

// TestCompositeDeleteHelper tests the CompositeDeleteHelper function.
func TestCompositeDeleteHelper(t *testing.T) {
	t.Run("deletes when not in file store", func(t *testing.T) {
		deleted := false

		fileExists := func(id string) (bool, error) {
			return false, nil
		}
		dbDelete := func(id string) error {
			deleted = true
			return nil
		}

		err := CompositeDeleteHelper("db-1", fileExists, dbDelete, errDeclarative)
		assert.NoError(t, err)
		assert.True(t, deleted)
	})

	t.Run("fails when trying to delete declarative resource", func(t *testing.T) {
		deleted := false

		fileExists := func(id string) (bool, error) {
			return true, nil
		}
		dbDelete := func(id string) error {
			deleted = true
			return nil
		}

		err := CompositeDeleteHelper("declarative-1", fileExists, dbDelete, errDeclarative)
		assert.Error(t, err)
		assert.Equal(t, errDeclarative, err)
		assert.False(t, deleted)
	})

	t.Run("propagates file store I/O error", func(t *testing.T) {
		deleted := false

		fileExists := func(id string) (bool, error) {
			return false, errIOError
		}
		dbDelete := func(id string) error {
			deleted = true
			return nil
		}

		err := CompositeDeleteHelper("db-1", fileExists, dbDelete, errDeclarative)
		assert.Error(t, err)
		assert.Equal(t, errIOError, err)
		assert.False(t, deleted)
	})

	t.Run("propagates file store parse error", func(t *testing.T) {
		deleted := false

		fileExists := func(id string) (bool, error) {
			return false, errParseError
		}
		dbDelete := func(id string) error {
			deleted = true
			return nil
		}

		err := CompositeDeleteHelper("db-1", fileExists, dbDelete, errDeclarative)
		assert.Error(t, err)
		assert.Equal(t, errParseError, err)
		assert.False(t, deleted)
	})
}

// testBooleanCheckerHelper is a helper to test boolean checker functions (conflict, hasChildren).
func testBooleanCheckerHelper(
	t *testing.T,
	checkerFunc func(func() (bool, error), func() (bool, error)) (bool, error),
) {
	t.Run("true in DB store", func(t *testing.T) {
		dbChecker := func() (bool, error) { return true, nil }
		fileChecker := func() (bool, error) { return false, nil }
		result, err := checkerFunc(dbChecker, fileChecker)
		assert.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("true in file store only", func(t *testing.T) {
		dbChecker := func() (bool, error) { return false, nil }
		fileChecker := func() (bool, error) { return true, nil }
		result, err := checkerFunc(dbChecker, fileChecker)
		assert.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("false in either store", func(t *testing.T) {
		dbChecker := func() (bool, error) { return false, nil }
		fileChecker := func() (bool, error) { return false, nil }
		result, err := checkerFunc(dbChecker, fileChecker)
		assert.NoError(t, err)
		assert.False(t, result)
	})

	t.Run("propagates DB error", func(t *testing.T) {
		dbErr := errIOError
		dbChecker := func() (bool, error) { return false, dbErr }
		fileChecker := func() (bool, error) { return false, nil }
		result, err := checkerFunc(dbChecker, fileChecker)
		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.False(t, result)
	})

	t.Run("propagates file store error", func(t *testing.T) {
		fileErr := errParseError
		dbChecker := func() (bool, error) { return false, nil }
		fileChecker := func() (bool, error) { return false, fileErr }
		result, err := checkerFunc(dbChecker, fileChecker)
		assert.Error(t, err)
		assert.Equal(t, fileErr, err)
		assert.False(t, result)
	})
}

// TestCompositeMergeCountHelper tests the CompositeMergeCountHelper function.
func TestCompositeMergeCountHelper(t *testing.T) {
	t.Run("merges counts from both stores", func(t *testing.T) {
		dbCounter := func() (int, error) {
			return 5, nil
		}
		fileCounter := func() (int, error) {
			return 3, nil
		}

		count, err := CompositeMergeCountHelper(dbCounter, fileCounter)
		assert.NoError(t, err)
		assert.Equal(t, 8, count)
	})

	t.Run("handles zero counts", func(t *testing.T) {
		dbCounter := func() (int, error) {
			return 0, nil
		}
		fileCounter := func() (int, error) {
			return 0, nil
		}

		count, err := CompositeMergeCountHelper(dbCounter, fileCounter)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("propagates DB error", func(t *testing.T) {
		dbErr := errIOError
		dbCounter := func() (int, error) {
			return 0, dbErr
		}
		fileCounter := func() (int, error) {
			return 3, nil
		}

		count, err := CompositeMergeCountHelper(dbCounter, fileCounter)
		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.Equal(t, 0, count)
	})

	t.Run("propagates file store error", func(t *testing.T) {
		fileErr := errParseError
		dbCounter := func() (int, error) {
			return 5, nil
		}
		fileCounter := func() (int, error) {
			return 0, fileErr
		}

		count, err := CompositeMergeCountHelper(dbCounter, fileCounter)
		assert.Error(t, err)
		assert.Equal(t, fileErr, err)
		assert.Equal(t, 0, count)
	})
}

// TestCompositeHasChildrenHelper tests the CompositeHasChildrenHelper function.
func TestCompositeHasChildrenHelper(t *testing.T) {
	testBooleanCheckerHelper(t, CompositeHasChildrenHelper)
}

// TestCompositeIsDeclarativeHelper tests the CompositeIsDeclarativeHelper function.
func TestCompositeIsDeclarativeHelper(t *testing.T) {
	t.Run("returns true when resource exists in file store", func(t *testing.T) {
		fileExists := func(id string) (bool, error) {
			return true, nil
		}

		isDeclarative := CompositeIsDeclarativeHelper("test-id", fileExists)
		assert.True(t, isDeclarative)
	})

	t.Run("returns false when resource not in file store", func(t *testing.T) {
		fileExists := func(id string) (bool, error) {
			return false, nil
		}

		isDeclarative := CompositeIsDeclarativeHelper("test-id", fileExists)
		assert.False(t, isDeclarative)
	})

	t.Run("returns false on file store error", func(t *testing.T) {
		fileExists := func(id string) (bool, error) {
			return false, errIOError
		}

		isDeclarative := CompositeIsDeclarativeHelper("test-id", fileExists)
		assert.False(t, isDeclarative)
	})

	t.Run("returns false on parse error", func(t *testing.T) {
		fileExists := func(id string) (bool, error) {
			return false, errParseError
		}

		isDeclarative := CompositeIsDeclarativeHelper("test-id", fileExists)
		assert.False(t, isDeclarative)
	})
}

// TestCompositeMergeListHelper tests the CompositeMergeListHelper function.
func TestCompositeMergeListHelper(t *testing.T) {
	merger := func(list1, list2 []compositeTestEntity) []compositeTestEntity {
		result := append([]compositeTestEntity{}, list1...)
		return append(result, list2...)
	}

	t.Run("returns error for negative limit", func(t *testing.T) {
		firstCounter := func() (int, error) { return 10, nil }
		secondCounter := func() (int, error) { return 15, nil }
		firstFetcher := func(count int) ([]compositeTestEntity, error) {
			return []compositeTestEntity{}, nil
		}
		secondFetcher := func(count int) ([]compositeTestEntity, error) {
			return []compositeTestEntity{}, nil
		}

		result, err := CompositeMergeListHelper(
			firstCounter, secondCounter, firstFetcher, secondFetcher,
			merger, -5, 0,
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "limit must be non-negative")
		assert.Contains(t, err.Error(), "-5")
		assert.Nil(t, result)
	})

	t.Run("returns error for negative offset", func(t *testing.T) {
		firstCounter := func() (int, error) { return 10, nil }
		secondCounter := func() (int, error) { return 15, nil }
		firstFetcher := func(count int) ([]compositeTestEntity, error) {
			return []compositeTestEntity{}, nil
		}
		secondFetcher := func(count int) ([]compositeTestEntity, error) {
			return []compositeTestEntity{}, nil
		}

		result, err := CompositeMergeListHelper(
			firstCounter, secondCounter, firstFetcher, secondFetcher,
			merger, 5, -10,
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "offset must be non-negative")
		assert.Contains(t, err.Error(), "-10")
		assert.Nil(t, result)
	})

	t.Run("basic pagination works with valid inputs", func(t *testing.T) {
		firstCounter := func() (int, error) { return 3, nil }
		secondCounter := func() (int, error) { return 2, nil }
		firstFetcher := func(count int) ([]compositeTestEntity, error) {
			return []compositeTestEntity{
				{ID: "f1", Name: "File1"},
				{ID: "f2", Name: "File2"},
				{ID: "f3", Name: "File3"},
			}, nil
		}
		secondFetcher := func(count int) ([]compositeTestEntity, error) {
			return []compositeTestEntity{
				{ID: "d1", Name: "DB1"},
				{ID: "d2", Name: "DB2"},
			}, nil
		}

		result, err := CompositeMergeListHelper(
			firstCounter, secondCounter, firstFetcher, secondFetcher,
			merger, 2, 1,
		)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "f2", result[0].ID)
		assert.Equal(t, "f3", result[1].ID)
	})
}

// Helper function to create test fetchers for CompositeMergeListHelperWithLimit tests.
func createTestFetchers(firstCount, secondCount int) (
	func() (int, error),
	func() (int, error),
	func(int) ([]compositeTestEntity, error),
	func(int) ([]compositeTestEntity, error),
) {
	firstCounter := func() (int, error) { return firstCount, nil }
	secondCounter := func() (int, error) { return secondCount, nil }
	firstFetcher := func(limit int) ([]compositeTestEntity, error) {
		items := make([]compositeTestEntity, limit)
		for i := 0; i < limit; i++ {
			items[i] = compositeTestEntity{ID: "f" + string(rune(i)), Name: "File"}
		}
		return items, nil
	}
	secondFetcher := func(limit int) ([]compositeTestEntity, error) {
		items := make([]compositeTestEntity, limit)
		for i := 0; i < limit; i++ {
			items[i] = compositeTestEntity{ID: "d" + string(rune(i)), Name: "DB"}
		}
		return items, nil
	}
	return firstCounter, secondCounter, firstFetcher, secondFetcher
}

// TestCompositeMergeListHelperWithLimit tests the CompositeMergeListHelperWithLimit function.
func TestCompositeMergeListHelperWithLimit(t *testing.T) {
	// Helper to create merger function
	merger := func(first, second []compositeTestEntity) []compositeTestEntity {
		seen := make(map[string]bool)
		result := make([]compositeTestEntity, 0)
		for _, item := range first {
			if !seen[item.ID] {
				seen[item.ID] = true
				result = append(result, item)
			}
		}
		for _, item := range second {
			if !seen[item.ID] {
				seen[item.ID] = true
				result = append(result, item)
			}
		}
		return result
	}

	t.Run("basic pagination without limit", func(t *testing.T) {
		firstCounter := func() (int, error) { return 3, nil }
		secondCounter := func() (int, error) { return 2, nil }
		firstFetcher := func(limit int) ([]compositeTestEntity, error) {
			return []compositeTestEntity{
				{ID: "f1", Name: "File1"},
				{ID: "f2", Name: "File2"},
				{ID: "f3", Name: "File3"},
			}, nil
		}
		secondFetcher := func(limit int) ([]compositeTestEntity, error) {
			return []compositeTestEntity{
				{ID: "d1", Name: "DB1"},
				{ID: "d2", Name: "DB2"},
			}, nil
		}

		result, limitExceeded, err := CompositeMergeListHelperWithLimit(
			firstCounter, secondCounter, firstFetcher, secondFetcher,
			merger, 2, 1, 0,
		)

		assert.NoError(t, err)
		assert.False(t, limitExceeded)
		assert.Len(t, result, 2)
		assert.Equal(t, "f2", result[0].ID)
		assert.Equal(t, "f3", result[1].ID)
	})

	t.Run("applies maxRecords limit and returns limitExceeded=true", func(t *testing.T) {
		firstCounter, secondCounter, firstFetcher, secondFetcher := createTestFetchers(600, 700)

		result, limitExceeded, err := CompositeMergeListHelperWithLimit(
			firstCounter, secondCounter, firstFetcher, secondFetcher,
			merger, 10, 0, 1000,
		)

		assert.NoError(t, err)
		assert.True(t, limitExceeded)
		assert.Len(t, result, 10)
	})

	t.Run("offset beyond effective total returns empty", func(t *testing.T) {
		firstCounter := func() (int, error) { return 600, nil }
		secondCounter := func() (int, error) { return 700, nil }
		firstFetcher := func(limit int) ([]compositeTestEntity, error) {
			return []compositeTestEntity{}, nil
		}
		secondFetcher := func(limit int) ([]compositeTestEntity, error) {
			return []compositeTestEntity{}, nil
		}

		// offset=1100 > effectiveTotal=1000
		result, limitExceeded, err := CompositeMergeListHelperWithLimit(
			firstCounter, secondCounter, firstFetcher, secondFetcher,
			merger, 10, 1100, 1000,
		)

		assert.NoError(t, err)
		assert.True(t, limitExceeded)
		assert.Empty(t, result)
	})

	t.Run("scatter-gather at high offset with limit", func(t *testing.T) {
		firstCounter := func() (int, error) { return 500, nil }
		secondCounter := func() (int, error) { return 500, nil }

		// Track what limits are requested
		var firstRequestedLimit, secondRequestedLimit int

		firstFetcher := func(limit int) ([]compositeTestEntity, error) {
			firstRequestedLimit = limit
			items := make([]compositeTestEntity, limit)
			for i := 0; i < limit; i++ {
				items[i] = compositeTestEntity{ID: "f" + string(rune(i)), Name: "File"}
			}
			return items, nil
		}
		secondFetcher := func(limit int) ([]compositeTestEntity, error) {
			secondRequestedLimit = limit
			items := make([]compositeTestEntity, limit)
			for i := 0; i < limit; i++ {
				items[i] = compositeTestEntity{ID: "d" + string(rune(i)), Name: "DB"}
			}
			return items, nil
		}

		// offset=900, limit=100, maxRecords=1000
		// depth = 1000, should fetch min(1000, 500) = 500 from each
		result, limitExceeded, err := CompositeMergeListHelperWithLimit(
			firstCounter, secondCounter, firstFetcher, secondFetcher,
			merger, 100, 900, 1000,
		)

		assert.NoError(t, err)
		assert.False(t, limitExceeded) // 1000 total = 1000 limit exactly
		assert.Equal(t, 500, firstRequestedLimit)
		assert.Equal(t, 500, secondRequestedLimit)
		assert.Len(t, result, 100) // Should get 100 items from offset 900
	})

	t.Run("handles first counter error", func(t *testing.T) {
		firstCounter := func() (int, error) { return 0, errIOError }
		secondCounter := func() (int, error) { return 5, nil }
		firstFetcher := func(limit int) ([]compositeTestEntity, error) {
			return []compositeTestEntity{}, nil
		}
		secondFetcher := func(limit int) ([]compositeTestEntity, error) {
			return []compositeTestEntity{}, nil
		}

		result, limitExceeded, err := CompositeMergeListHelperWithLimit(
			firstCounter, secondCounter, firstFetcher, secondFetcher,
			merger, 10, 0, 0,
		)

		assert.Error(t, err)
		assert.Equal(t, errIOError, err)
		assert.False(t, limitExceeded)
		assert.Nil(t, result)
	})

	t.Run("handles second counter error", func(t *testing.T) {
		firstCounter := func() (int, error) { return 5, nil }
		secondCounter := func() (int, error) { return 0, errIOError }
		firstFetcher := func(limit int) ([]compositeTestEntity, error) {
			return []compositeTestEntity{}, nil
		}
		secondFetcher := func(limit int) ([]compositeTestEntity, error) {
			return []compositeTestEntity{}, nil
		}

		result, limitExceeded, err := CompositeMergeListHelperWithLimit(
			firstCounter, secondCounter, firstFetcher, secondFetcher,
			merger, 10, 0, 0,
		)

		assert.Error(t, err)
		assert.Equal(t, errIOError, err)
		assert.False(t, limitExceeded)
		assert.Nil(t, result)
	})

	t.Run("handles first fetcher error", func(t *testing.T) {
		firstCounter := func() (int, error) { return 5, nil }
		secondCounter := func() (int, error) { return 5, nil }
		firstFetcher := func(limit int) ([]compositeTestEntity, error) {
			return nil, errIOError
		}
		secondFetcher := func(limit int) ([]compositeTestEntity, error) {
			return []compositeTestEntity{{ID: "d1", Name: "DB1"}}, nil
		}

		result, limitExceeded, err := CompositeMergeListHelperWithLimit(
			firstCounter, secondCounter, firstFetcher, secondFetcher,
			merger, 10, 0, 0,
		)

		assert.Error(t, err)
		assert.Equal(t, errIOError, err)
		assert.False(t, limitExceeded)
		assert.Nil(t, result)
	})

	t.Run("handles second fetcher error", func(t *testing.T) {
		firstCounter := func() (int, error) { return 5, nil }
		secondCounter := func() (int, error) { return 5, nil }
		firstFetcher := func(limit int) ([]compositeTestEntity, error) {
			return []compositeTestEntity{{ID: "f1", Name: "File1"}}, nil
		}
		secondFetcher := func(limit int) ([]compositeTestEntity, error) {
			return nil, errIOError
		}

		result, limitExceeded, err := CompositeMergeListHelperWithLimit(
			firstCounter, secondCounter, firstFetcher, secondFetcher,
			merger, 10, 0, 0,
		)

		assert.Error(t, err)
		assert.Equal(t, errIOError, err)
		assert.False(t, limitExceeded)
		assert.Nil(t, result)
	})

	t.Run("deduplicates merged results", func(t *testing.T) {
		firstCounter := func() (int, error) { return 3, nil }
		secondCounter := func() (int, error) { return 3, nil }
		firstFetcher := func(limit int) ([]compositeTestEntity, error) {
			return []compositeTestEntity{
				{ID: "1", Name: "File1"},
				{ID: "2", Name: "File2"},
				{ID: "3", Name: "File3"},
			}, nil
		}
		secondFetcher := func(limit int) ([]compositeTestEntity, error) {
			return []compositeTestEntity{
				{ID: "2", Name: "DB2"}, // Duplicate ID
				{ID: "4", Name: "DB4"},
				{ID: "5", Name: "DB5"},
			}, nil
		}

		result, limitExceeded, err := CompositeMergeListHelperWithLimit(
			firstCounter, secondCounter, firstFetcher, secondFetcher,
			merger, 10, 0, 0,
		)

		assert.NoError(t, err)
		assert.False(t, limitExceeded)
		assert.Len(t, result, 5) // 6 items - 1 duplicate = 5

		// Verify no duplicate IDs
		ids := make(map[string]bool)
		for _, item := range result {
			assert.False(t, ids[item.ID], "Found duplicate ID: "+item.ID)
			ids[item.ID] = true
		}
	})

	t.Run("empty stores return empty result", func(t *testing.T) {
		firstCounter := func() (int, error) { return 0, nil }
		secondCounter := func() (int, error) { return 0, nil }
		firstFetcher := func(limit int) ([]compositeTestEntity, error) {
			return []compositeTestEntity{}, nil
		}
		secondFetcher := func(limit int) ([]compositeTestEntity, error) {
			return []compositeTestEntity{}, nil
		}

		result, limitExceeded, err := CompositeMergeListHelperWithLimit(
			firstCounter, secondCounter, firstFetcher, secondFetcher,
			merger, 10, 0, 0,
		)

		assert.NoError(t, err)
		assert.False(t, limitExceeded)
		assert.Empty(t, result)
	})

	t.Run("maxRecords=0 means no limit", func(t *testing.T) {
		firstCounter, secondCounter, firstFetcher, secondFetcher := createTestFetchers(10, 15)

		result, limitExceeded, err := CompositeMergeListHelperWithLimit(
			firstCounter, secondCounter, firstFetcher, secondFetcher,
			merger, 5, 0, 0,
		)

		assert.NoError(t, err)
		assert.False(t, limitExceeded)
		assert.Len(t, result, 5)
	})

	t.Run("returns error for negative limit", func(t *testing.T) {
		firstCounter, secondCounter, firstFetcher, secondFetcher := createTestFetchers(10, 15)

		result, limitExceeded, err := CompositeMergeListHelperWithLimit(
			firstCounter, secondCounter, firstFetcher, secondFetcher,
			merger, -5, 0, 0,
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "limit must be non-negative")
		assert.Contains(t, err.Error(), "-5")
		assert.False(t, limitExceeded)
		assert.Nil(t, result)
	})

	t.Run("returns error for negative offset", func(t *testing.T) {
		firstCounter, secondCounter, firstFetcher, secondFetcher := createTestFetchers(10, 15)

		result, limitExceeded, err := CompositeMergeListHelperWithLimit(
			firstCounter, secondCounter, firstFetcher, secondFetcher,
			merger, 5, -10, 0,
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "offset must be non-negative")
		assert.Contains(t, err.Error(), "-10")
		assert.False(t, limitExceeded)
		assert.Nil(t, result)
	})
}
