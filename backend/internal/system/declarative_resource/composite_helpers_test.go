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
