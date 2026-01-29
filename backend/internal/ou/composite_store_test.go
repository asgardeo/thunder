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

package ou

import (
	"errors"
	"testing"

	"github.com/asgardeo/thunder/internal/system/declarative_resource/entity"

	"github.com/stretchr/testify/suite"
)

// CompositeStoreTestSuite tests the composite OU store functionality.
type CompositeStoreTestSuite struct {
	suite.Suite
	fileStore      organizationUnitStoreInterface
	dbStoreMock    *organizationUnitStoreInterfaceMock
	compositeStore *compositeOUStore
}

// SetupTest sets up the test environment.
func (suite *CompositeStoreTestSuite) SetupTest() {
	// Clear the singleton entity store to avoid state leakage between tests
	_ = entity.GetInstance().Clear()

	// Create NEW file-based store for each test to avoid state leakage
	suite.fileStore = newFileBasedStore()

	// Create mock DB store
	suite.dbStoreMock = newOrganizationUnitStoreInterfaceMock(suite.T())

	// Create composite store
	suite.compositeStore = newCompositeOUStore(suite.fileStore, suite.dbStoreMock)
}

// TestCompositeStore_GetOrganizationUnit tests retrieving OUs from composite store.
func (suite *CompositeStoreTestSuite) TestCompositeStore_GetOrganizationUnit() {
	testCases := []struct {
		name           string
		ouID           string
		setupFileStore func()
		setupDBStore   func()
		want           OrganizationUnit
		wantErr        bool
	}{
		{
			name: "retrieves from DB store - marks as mutable",
			ouID: "db-ou-1",
			setupFileStore: func() {
				// File store doesn't have this OU
			},
			setupDBStore: func() {
				suite.dbStoreMock.On("GetOrganizationUnit", "db-ou-1").
					Return(OrganizationUnit{
						ID:     "db-ou-1",
						Handle: "db-handle",
						Name:   "DB OU",
					}, nil).
					Once()
			},
			want: OrganizationUnit{
				ID:     "db-ou-1",
				Handle: "db-handle",
				Name:   "DB OU",
			},
		},
		{
			name: "retrieves from file store - marks as immutable",
			ouID: "file-ou-1",
			setupFileStore: func() {
				// Add OU to file store
				err := suite.fileStore.CreateOrganizationUnit(OrganizationUnit{
					ID:     "file-ou-1",
					Handle: "file-handle",
					Name:   "File OU",
				})
				suite.NoError(err)
			},
			setupDBStore: func() {
				suite.dbStoreMock.On("GetOrganizationUnit", "file-ou-1").
					Return(OrganizationUnit{}, ErrOrganizationUnitNotFound).
					Once()
			},
			want: OrganizationUnit{
				ID:     "file-ou-1",
				Handle: "file-handle",
				Name:   "File OU",
			},
		},
		{
			name: "not found in either store",
			ouID: "nonexistent",
			setupFileStore: func() {
				// OU not in file store
			},
			setupDBStore: func() {
				suite.dbStoreMock.On("GetOrganizationUnit", "nonexistent").
					Return(OrganizationUnit{}, ErrOrganizationUnitNotFound).
					Once()
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest() // Fresh setup for each test
			tc.setupFileStore()
			tc.setupDBStore()

			got, err := suite.compositeStore.GetOrganizationUnit(tc.ouID)

			if tc.wantErr {
				suite.Error(err)
				suite.Equal(ErrOrganizationUnitNotFound, err)
			} else {
				suite.NoError(err)
				suite.Equal(tc.want.ID, got.ID)
				suite.Equal(tc.want.Handle, got.Handle)
				suite.Equal(tc.want.Name, got.Name)
			}
		})
	}
}

// TestCompositeStore_CreateOrganizationUnit tests creating OUs.
func (suite *CompositeStoreTestSuite) TestCompositeStore_CreateOrganizationUnit() {
	suite.Run("creates in DB store", func() {
		ou := OrganizationUnit{
			ID:     "new-ou-1",
			Handle: "new-handle",
			Name:   "New OU",
		}

		suite.dbStoreMock.On("CreateOrganizationUnit", ou).
			Return(nil).
			Once()

		err := suite.compositeStore.CreateOrganizationUnit(ou)
		suite.NoError(err)
	})
}

// TestCompositeStore_UpdateOrganizationUnit tests updating OUs.
func (suite *CompositeStoreTestSuite) TestCompositeStore_UpdateOrganizationUnit() {
	suite.Run("updates DB OU successfully", func() {
		ou := OrganizationUnit{
			ID:     "db-ou-1",
			Handle: "updated-handle",
			Name:   "Updated OU",
		}

		suite.dbStoreMock.On("UpdateOrganizationUnit", ou).
			Return(nil).
			Once()

		err := suite.compositeStore.UpdateOrganizationUnit(ou)
		suite.NoError(err)
	})
}

// TestCompositeStore_DeleteOrganizationUnit tests deleting OUs.
func (suite *CompositeStoreTestSuite) TestCompositeStore_DeleteOrganizationUnit() {
	suite.Run("deletes DB OU successfully", func() {
		// Children validation is done at service layer, not store layer
		suite.dbStoreMock.On("DeleteOrganizationUnit", "db-ou-1").
			Return(nil).
			Once()

		err := suite.compositeStore.DeleteOrganizationUnit("db-ou-1")
		suite.NoError(err)
	})
}

// TestCompositeStore_IsOrganizationUnitExists tests existence checks across both stores.
func (suite *CompositeStoreTestSuite) TestCompositeStore_IsOrganizationUnitExists() {
	suite.Run("exists in DB store", func() {
		suite.dbStoreMock.On("IsOrganizationUnitExists", "db-ou-1").
			Return(true, nil).
			Once()

		exists, err := suite.compositeStore.IsOrganizationUnitExists("db-ou-1")
		suite.NoError(err)
		suite.True(exists)
	})

	suite.Run("exists in file store", func() {
		// Add OU to file store
		err := suite.fileStore.CreateOrganizationUnit(OrganizationUnit{
			ID:     "file-ou-1",
			Handle: "file-handle",
			Name:   "File OU",
		})
		suite.NoError(err)

		exists, err := suite.compositeStore.IsOrganizationUnitExists("file-ou-1")
		suite.NoError(err)
		suite.True(exists)
	})

	suite.Run("not found in either store", func() {
		suite.dbStoreMock.On("IsOrganizationUnitExists", "nonexistent").
			Return(false, nil).
			Once()

		exists, err := suite.compositeStore.IsOrganizationUnitExists("nonexistent")
		suite.NoError(err)
		suite.False(exists)
	})
}

// TestCompositeStore_ConflictChecks tests name and handle conflict detection across both stores.
func (suite *CompositeStoreTestSuite) TestCompositeStore_ConflictChecks() {
	suite.Run("detects name conflict in DB store", func() {
		// File store checked first - returns false (no conflict)
		// Then DB store checked - returns true (conflict)
		suite.dbStoreMock.On("CheckOrganizationUnitNameConflict", "test-name", (*string)(nil)).
			Return(true, nil).
			Once()

		conflict, err := suite.compositeStore.CheckOrganizationUnitNameConflict("test-name", nil)
		suite.NoError(err)
		suite.True(conflict)
	})

	suite.Run("detects name conflict in file store", func() {
		// Add OU to file store
		err := suite.fileStore.CreateOrganizationUnit(OrganizationUnit{
			ID:     "file-ou-1",
			Handle: "file-handle",
			Name:   "Conflict Name",
		})
		suite.NoError(err)

		// File store checked first - returns true (conflict found)
		// DB store not called since file store found conflict
		// No mock expectation for dbStore

		conflict, err := suite.compositeStore.CheckOrganizationUnitNameConflict("Conflict Name", nil)
		suite.NoError(err)
		suite.True(conflict)
	})

	suite.Run("no name conflict in either store", func() {
		suite.dbStoreMock.On("CheckOrganizationUnitNameConflict", "unique-name", (*string)(nil)).
			Return(false, nil).
			Once()

		conflict, err := suite.compositeStore.CheckOrganizationUnitNameConflict("unique-name", nil)
		suite.NoError(err)
		suite.False(conflict)
	})
}

// TestCompositeStore_ChildrenOperations tests child OU operations across both stores.
func (suite *CompositeStoreTestSuite) TestCompositeStore_ChildrenOperations() {
	suite.Run("counts children from both stores", func() {
		suite.SetupTest() // Fresh setup
		parentID := "parent-ou"

		suite.dbStoreMock.On("GetOrganizationUnitChildrenCount", parentID).
			Return(2, nil).
			Once()

		// Add parent and child to file store
		err := suite.fileStore.CreateOrganizationUnit(OrganizationUnit{
			ID:     parentID,
			Handle: "parent",
			Name:   "Parent",
		})
		suite.NoError(err)
		err = suite.fileStore.CreateOrganizationUnit(OrganizationUnit{
			ID:     "file-child-1",
			Handle: "child1",
			Name:   "Child 1",
			Parent: &parentID,
		})
		suite.NoError(err)

		count, err := suite.compositeStore.GetOrganizationUnitChildrenCount(parentID)
		suite.NoError(err)
		suite.Equal(3, count) // 2 from DB + 1 from file
	})

	suite.Run("checks if OU has children in either store", func() {
		suite.dbStoreMock.On("CheckOrganizationUnitHasChildResources", "parent-ou").
			Return(true, nil).
			Once()

		hasChildren, err := suite.compositeStore.CheckOrganizationUnitHasChildResources("parent-ou")
		suite.NoError(err)
		suite.True(hasChildren)
	})
}

// TestCompositeStore_ListOperations tests list operations include both stores.
func (suite *CompositeStoreTestSuite) TestCompositeStore_ListOperations() {
	suite.Run("GetOrganizationUnitListCount returns count from both stores", func() {
		// Add OUs to file store
		_ = suite.fileStore.CreateOrganizationUnit(OrganizationUnit{
			ID: "file-ou-1", Handle: "file-1", Name: "File OU 1",
		})
		_ = suite.fileStore.CreateOrganizationUnit(OrganizationUnit{
			ID: "file-ou-2", Handle: "file-2", Name: "File OU 2",
		})

		// Mock DB store count - CompositeMergeCountHelper calls dbStore.GetOrganizationUnitListCount once
		suite.dbStoreMock.On("GetOrganizationUnitListCount").Return(3, nil).Once()

		count, err := suite.compositeStore.GetOrganizationUnitListCount()
		suite.NoError(err)
		suite.Equal(5, count) // 3 from DB + 2 from file
	})

	suite.Run("GetOrganizationUnitList returns merged results", func() {
		// Reset state from previous sub-test by reinitializing stores
		_ = entity.GetInstance().Clear()
		suite.fileStore = newFileBasedStore()
		suite.compositeStore = newCompositeOUStore(suite.fileStore, suite.dbStoreMock)

		// Add 1 OU to file store
		_ = suite.fileStore.CreateOrganizationUnit(OrganizationUnit{
			ID: "file-ou-1", Handle: "file-1", Name: "File OU 1",
		})

		// Mock DB store - note that implementation calls count to determine how many to fetch from each
		// dbCount will be called to fetch all from DB
		suite.dbStoreMock.On("GetOrganizationUnitListCount").Return(2, nil).Once()
		// Then fileCount from fileStore (real store) returns 1
		// Total = 3, so it will try to fetch all 3 (dbCount=2, fileCount=1)
		suite.dbStoreMock.On("GetOrganizationUnitList", 2, 0).Return([]OrganizationUnitBasic{
			{ID: "db-ou-1", Handle: "db-1", Name: "DB OU 1"},
			{ID: "db-ou-2", Handle: "db-2", Name: "DB OU 2"},
		}, nil).Once()
		// fileStore.GetOrganizationUnitList will be called with (1, 0) and return the file OU

		list, err := suite.compositeStore.GetOrganizationUnitList(10, 0)
		suite.NoError(err)
		suite.Len(list, 3) // 2 from DB + 1 from file

		// Verify all OUs are present
		ids := make(map[string]bool)
		for _, ou := range list {
			ids[ou.ID] = true
		}
		suite.True(ids["db-ou-1"])
		suite.True(ids["db-ou-2"])
		suite.True(ids["file-ou-1"])
	})

	suite.Run("user/group operations use DB store only", func() {
		suite.dbStoreMock.On("GetOrganizationUnitUsersCount", "ou-1").
			Return(10, nil).
			Once()

		count, err := suite.compositeStore.GetOrganizationUnitUsersCount("ou-1")
		suite.NoError(err)
		suite.Equal(10, count)
	})
}

// TestCompositeStore_IsOrganizationUnitDeclarative tests checking if an OU is immutable.
func (suite *CompositeStoreTestSuite) TestCompositeStore_IsOrganizationUnitDeclarative() {
	suite.Run("returns true for immutable OU (exists in file store)", func() {
		// Add OU to file store
		err := suite.fileStore.CreateOrganizationUnit(OrganizationUnit{
			ID:     "immutable-ou-1",
			Handle: "immutable-handle",
			Name:   "Declarative OU",
		})
		suite.NoError(err)

		isDeclarative := suite.compositeStore.IsOrganizationUnitDeclarative("immutable-ou-1")
		suite.True(isDeclarative)
	})

	suite.Run("returns false for mutable OU (not in file store)", func() {
		isDeclarative := suite.compositeStore.IsOrganizationUnitDeclarative("db-ou-1")
		suite.False(isDeclarative)
	})

	suite.Run("returns false for non-existent OU", func() {
		isDeclarative := suite.compositeStore.IsOrganizationUnitDeclarative("nonexistent")
		suite.False(isDeclarative)
	})
}

// TestCompositeStoreTestSuite runs the test suite.
func TestCompositeStoreTestSuite(t *testing.T) {
	// Initialize entity store for file-based store
	_ = entity.GetInstance()
	suite.Run(t, new(CompositeStoreTestSuite))
}

const (
	testCoverageParentOUID = "parent-ou"
	testCoverageChild1ID   = "child-1"
)

// CompositeStoreCoverageTestSuite provides additional coverage tests for composite store.
type CompositeStoreCoverageTestSuite struct {
	suite.Suite
	fileStore      organizationUnitStoreInterface
	dbStoreMock    *organizationUnitStoreInterfaceMock
	compositeStore *compositeOUStore
}

func (suite *CompositeStoreCoverageTestSuite) SetupTest() {
	// Clear the singleton entity store to avoid state leakage between tests
	_ = entity.GetInstance().Clear()

	// Create NEW file-based store for each test to avoid state leakage
	suite.fileStore = newFileBasedStore()
	suite.dbStoreMock = newOrganizationUnitStoreInterfaceMock(suite.T())
	suite.compositeStore = newCompositeOUStore(suite.fileStore, suite.dbStoreMock)
}

// TestCompositeStore_GetOrganizationUnitList tests paginated list retrieval.
func (suite *CompositeStoreCoverageTestSuite) TestCompositeStore_GetOrganizationUnitList() {
	suite.Run("retrieves paginated list from DB store only", func() {
		expectedList := []OrganizationUnitBasic{
			{ID: "ou-1", Handle: "handle-1", Name: "OU 1"},
			{ID: "ou-2", Handle: "handle-2", Name: "OU 2"},
		}

		suite.dbStoreMock.On("GetOrganizationUnitListCount").
			Return(2, nil).
			Once()
		suite.dbStoreMock.On("GetOrganizationUnitList", 2, 0).
			Return(expectedList, nil).
			Once()

		result, err := suite.compositeStore.GetOrganizationUnitList(10, 0)
		suite.NoError(err)
		suite.Equal(expectedList, result)
	})

	suite.Run("propagates DB store error", func() {
		dbErr := errors.New("database error")
		suite.dbStoreMock.On("GetOrganizationUnitListCount").
			Return(0, dbErr).
			Once()

		result, err := suite.compositeStore.GetOrganizationUnitList(10, 0)
		suite.Error(err)
		suite.Equal(dbErr, err)
		suite.Empty(result)
	})
}

// TestCompositeStore_GetOrganizationUnitByPath tests path-based retrieval.
func (suite *CompositeStoreCoverageTestSuite) TestCompositeStore_GetOrganizationUnitByPath() {
	handles := []string{"level1", "level2"}

	suite.Run("retrieves from DB store by path", func() {
		expectedOU := OrganizationUnit{
			ID:     "ou-path-1",
			Handle: "level2",
			Name:   "Level 2 OU",
		}

		suite.dbStoreMock.On("GetOrganizationUnitByPath", handles).
			Return(expectedOU, nil).
			Once()

		result, err := suite.compositeStore.GetOrganizationUnitByPath(handles)
		suite.NoError(err)
		suite.Equal(expectedOU, result)
	})

	suite.Run("falls back to file store when not in DB", func() {
		// Create parent and child in file store
		parentID := "file-parent"
		err := suite.fileStore.CreateOrganizationUnit(OrganizationUnit{
			ID:     parentID,
			Handle: "level1",
			Name:   "Parent",
		})
		suite.NoError(err)

		err = suite.fileStore.CreateOrganizationUnit(OrganizationUnit{
			ID:     "file-child",
			Handle: "level2",
			Name:   "Child",
			Parent: &parentID,
		})
		suite.NoError(err)

		suite.dbStoreMock.On("GetOrganizationUnitByPath", handles).
			Return(OrganizationUnit{}, ErrOrganizationUnitNotFound).
			Once()

		result, err := suite.compositeStore.GetOrganizationUnitByPath(handles)
		suite.NoError(err)
		suite.Equal("file-child", result.ID)
		suite.Equal("level2", result.Handle)
	})

	suite.Run("returns not found when in neither store", func() {
		suite.SetupTest() // Fresh setup
		handlesTest := []string{"nonexistent1", "nonexistent2"}

		suite.dbStoreMock.On("GetOrganizationUnitByPath", handlesTest).
			Return(OrganizationUnit{}, ErrOrganizationUnitNotFound).
			Once()

		result, err := suite.compositeStore.GetOrganizationUnitByPath(handlesTest)
		suite.Error(err)
		suite.Equal(ErrOrganizationUnitNotFound, err)
		suite.Empty(result.ID)
	})

	suite.Run("propagates DB errors other than not found", func() {
		dbErr := errors.New("database connection error")
		suite.dbStoreMock.On("GetOrganizationUnitByPath", handles).
			Return(OrganizationUnit{}, dbErr).
			Once()

		result, err := suite.compositeStore.GetOrganizationUnitByPath(handles)
		suite.Error(err)
		suite.Equal(dbErr, err)
		suite.Empty(result.ID)
	})
}

// TestCompositeStore_CheckOrganizationUnitHandleConflict tests handle conflict detection.
func (suite *CompositeStoreCoverageTestSuite) TestCompositeStore_CheckOrganizationUnitHandleConflict() {
	suite.Run("detects handle conflict in DB store", func() {
		// File store checked first - returns false (no conflict)
		// Then DB store checked - returns true (conflict)
		suite.dbStoreMock.On("CheckOrganizationUnitHandleConflict", "test-handle", (*string)(nil)).
			Return(true, nil).
			Once()

		conflict, err := suite.compositeStore.CheckOrganizationUnitHandleConflict("test-handle", nil)
		suite.NoError(err)
		suite.True(conflict)
	})

	suite.Run("detects handle conflict in file store", func() {
		err := suite.fileStore.CreateOrganizationUnit(OrganizationUnit{
			ID:     "file-ou",
			Handle: "conflict-handle",
			Name:   "File OU",
		})
		suite.NoError(err)

		// File store checked first - returns true (conflict found)
		// DB store not called since file store found conflict
		// No mock expectation for dbStore

		conflict, err := suite.compositeStore.CheckOrganizationUnitHandleConflict("conflict-handle", nil)
		suite.NoError(err)
		suite.True(conflict)
	})

	suite.Run("no handle conflict in either store", func() {
		suite.dbStoreMock.On("CheckOrganizationUnitHandleConflict", "unique-handle", (*string)(nil)).
			Return(false, nil).
			Once()

		conflict, err := suite.compositeStore.CheckOrganizationUnitHandleConflict("unique-handle", nil)
		suite.NoError(err)
		suite.False(conflict)
	})

	suite.Run("propagates DB error", func() {
		dbErr := errors.New("db error")
		suite.dbStoreMock.On("CheckOrganizationUnitHandleConflict", "test", (*string)(nil)).
			Return(false, dbErr).
			Once()

		conflict, err := suite.compositeStore.CheckOrganizationUnitHandleConflict("test", nil)
		suite.Error(err)
		suite.Equal(dbErr, err)
		suite.False(conflict)
	})
}

func TestCompositeStoreCoverageTestSuite(t *testing.T) {
	suite.Run(t, new(CompositeStoreCoverageTestSuite))
}

// TestCompositeStore_CheckOrganizationUnitHasChildResources tests child resource detection.
func (suite *CompositeStoreCoverageTestSuite) TestCompositeStore_CheckOrganizationUnitHasChildResources() {
	suite.Run("has children in DB store", func() {
		// File store checked first - returns false (no children)
		// Then DB store checked - returns true (has children)
		suite.dbStoreMock.On("CheckOrganizationUnitHasChildResources", "ou-1").
			Return(true, nil).
			Once()

		hasChildren, err := suite.compositeStore.CheckOrganizationUnitHasChildResources("ou-1")
		suite.NoError(err)
		suite.True(hasChildren)
	})

	suite.Run("has children in file store only", func() {
		parentID := testCoverageParentOUID
		err := suite.fileStore.CreateOrganizationUnit(OrganizationUnit{
			ID:     parentID,
			Handle: "parent",
			Name:   "Parent",
		})
		suite.NoError(err)
		err = suite.fileStore.CreateOrganizationUnit(OrganizationUnit{
			ID:     "child-ou",
			Handle: "child",
			Name:   "Child",
			Parent: &parentID,
		})
		suite.NoError(err)

		// File store checked first - returns true (has children)
		// DB store not called since file store returned true
		// No mock expectation for dbStore

		hasChildren, err := suite.compositeStore.CheckOrganizationUnitHasChildResources(parentID)
		suite.NoError(err)
		suite.True(hasChildren)
	})

	suite.Run("no children in either store", func() {
		suite.dbStoreMock.On("CheckOrganizationUnitHasChildResources", "childless-ou").
			Return(false, nil).
			Once()

		hasChildren, err := suite.compositeStore.CheckOrganizationUnitHasChildResources("childless-ou")
		suite.NoError(err)
		suite.False(hasChildren)
	})

	suite.Run("propagates DB error", func() {
		dbErr := errors.New("db error")
		suite.dbStoreMock.On("CheckOrganizationUnitHasChildResources", "ou-1").
			Return(false, dbErr).
			Once()

		hasChildren, err := suite.compositeStore.CheckOrganizationUnitHasChildResources("ou-1")
		suite.Error(err)
		suite.Equal(dbErr, err)
		suite.False(hasChildren)
	})
}

// TestCompositeStore_GetOrganizationUnitChildrenList tests paginated children retrieval.
func (suite *CompositeStoreCoverageTestSuite) TestCompositeStore_GetOrganizationUnitChildrenList() {
	parentID := "parent-ou"

	suite.Run("merges children from both stores with pagination", func() {
		// Setup DB children
		dbChildren := []OrganizationUnitBasic{
			{ID: "db-child-1", Handle: "db1", Name: "DB Child 1"},
			{ID: "db-child-2", Handle: "db2", Name: "DB Child 2"},
		}

		// Setup file store parent and children
		err := suite.fileStore.CreateOrganizationUnit(OrganizationUnit{
			ID:     parentID,
			Handle: "parent",
			Name:   "Parent",
		})
		suite.NoError(err)
		err = suite.fileStore.CreateOrganizationUnit(OrganizationUnit{
			ID:     "file-child-1",
			Handle: "file1",
			Name:   "File Child 1",
			Parent: &parentID,
		})
		suite.NoError(err)

		suite.dbStoreMock.On("GetOrganizationUnitChildrenCount", parentID).
			Return(2, nil).
			Once()
		suite.dbStoreMock.On("GetOrganizationUnitChildrenList", parentID, 2, 0).
			Return(dbChildren, nil).
			Once()

		// Request first 2 items
		result, err := suite.compositeStore.GetOrganizationUnitChildrenList(parentID, 2, 0)
		suite.NoError(err)
		suite.Len(result, 2)
	})

	suite.Run("handles offset beyond total count", func() {
		suite.dbStoreMock.On("GetOrganizationUnitChildrenCount", parentID).
			Return(2, nil).
			Once()

		result, err := suite.compositeStore.GetOrganizationUnitChildrenList(parentID, 10, 100)
		suite.NoError(err)
		suite.Empty(result)
	})

	suite.Run("handles pagination with offset", func() {
		suite.SetupTest() // Fresh setup
		testParentID := "pagination-parent"
		dbChildren := []OrganizationUnitBasic{
			{ID: "db-child-1", Handle: "db1", Name: "DB Child 1"},
			{ID: "db-child-2", Handle: "db2", Name: "DB Child 2"},
			{ID: "db-child-3", Handle: "db3", Name: "DB Child 3"},
		}

		// Setup file store parent
		err := suite.fileStore.CreateOrganizationUnit(OrganizationUnit{
			ID:     testParentID,
			Handle: "parent",
			Name:   "Parent",
		})
		suite.NoError(err)

		suite.dbStoreMock.On("GetOrganizationUnitChildrenCount", testParentID).
			Return(3, nil).
			Once()
		suite.dbStoreMock.On("GetOrganizationUnitChildrenList", testParentID, 3, 0).
			Return(dbChildren, nil).
			Once()

		// Get second page (offset=2, limit=2)
		result, err := suite.compositeStore.GetOrganizationUnitChildrenList(testParentID, 2, 2)
		suite.NoError(err)
		suite.Len(result, 1) // Only 1 item remaining
		suite.Equal("db-child-3", result[0].ID)
	})

	suite.Run("propagates DB count error", func() {
		suite.SetupTest() // Fresh setup
		testParentID := "count-error-parent"
		err := suite.fileStore.CreateOrganizationUnit(OrganizationUnit{
			ID:     testParentID,
			Handle: "parent",
			Name:   "Parent",
		})
		suite.NoError(err)

		dbErr := errors.New("count error")
		suite.dbStoreMock.On("GetOrganizationUnitChildrenCount", testParentID).
			Return(0, dbErr).
			Once()

		result, err := suite.compositeStore.GetOrganizationUnitChildrenList(testParentID, 10, 0)
		suite.Error(err)
		suite.Equal(dbErr, err)
		suite.Nil(result)
	})

	suite.Run("propagates DB list error", func() {
		suite.SetupTest() // Fresh setup
		testParentID := "error-parent"
		// Create parent in file store
		err := suite.fileStore.CreateOrganizationUnit(OrganizationUnit{
			ID:     testParentID,
			Handle: "parent",
			Name:   "Parent",
		})
		suite.NoError(err)

		dbErr := errors.New("list error")
		suite.dbStoreMock.On("GetOrganizationUnitChildrenCount", testParentID).
			Return(2, nil).
			Once()
		suite.dbStoreMock.On("GetOrganizationUnitChildrenList", testParentID, 2, 0).
			Return([]OrganizationUnitBasic{}, dbErr).
			Once()

		result, err := suite.compositeStore.GetOrganizationUnitChildrenList(testParentID, 10, 0)
		suite.Error(err)
		suite.Equal(dbErr, err)
		suite.Nil(result)
	})

	suite.Run("deduplicates children by ID", func() {
		suite.SetupTest() // Fresh setup
		dedupeParentID := "dedupe-parent"

		// Setup DB children
		dbChildren := []OrganizationUnitBasic{
			{ID: "child-1", Handle: "db-handle", Name: "DB Child"},
			{ID: "child-2", Handle: "db2", Name: "DB Child 2"},
		}

		// Setup file store with duplicate ID
		err := suite.fileStore.CreateOrganizationUnit(OrganizationUnit{
			ID:     dedupeParentID,
			Handle: "parent",
			Name:   "Parent",
		})
		suite.NoError(err)
		err = suite.fileStore.CreateOrganizationUnit(OrganizationUnit{
			ID:     "child-1", // Same ID as DB child
			Handle: "file-handle",
			Name:   "File Child",
			Parent: &dedupeParentID,
		})
		suite.NoError(err)

		suite.dbStoreMock.On("GetOrganizationUnitChildrenCount", dedupeParentID).
			Return(2, nil).
			Once()
		suite.dbStoreMock.On("GetOrganizationUnitChildrenList", dedupeParentID, 2, 0).
			Return(dbChildren, nil).
			Once()

		result, err := suite.compositeStore.GetOrganizationUnitChildrenList(dedupeParentID, 10, 0)
		suite.NoError(err)
		// Should have 2 children (duplicate removed)
		suite.Len(result, 2)
		// DB child should take precedence
		for _, child := range result {
			if child.ID == testCoverageChild1ID {
				suite.Equal("db-handle", child.Handle)
			}
		}
	})
}

// TestCompositeStore_UserAndGroupOperations tests user/group list operations.
func (suite *CompositeStoreCoverageTestSuite) TestCompositeStore_UserAndGroupOperations() {
	suite.Run("retrieves users list from DB store", func() {
		expectedUsers := []User{
			{ID: "user-1"},
			{ID: "user-2"},
		}

		suite.dbStoreMock.On("GetOrganizationUnitUsersList", "ou-1", 10, 0).
			Return(expectedUsers, nil).
			Once()

		result, err := suite.compositeStore.GetOrganizationUnitUsersList("ou-1", 10, 0)
		suite.NoError(err)
		suite.Equal(expectedUsers, result)
	})

	suite.Run("retrieves groups count from DB store", func() {
		suite.dbStoreMock.On("GetOrganizationUnitGroupsCount", "ou-1").
			Return(5, nil).
			Once()

		count, err := suite.compositeStore.GetOrganizationUnitGroupsCount("ou-1")
		suite.NoError(err)
		suite.Equal(5, count)
	})

	suite.Run("retrieves groups list from DB store", func() {
		expectedGroups := []Group{
			{ID: "group-1", Name: "Group 1"},
			{ID: "group-2", Name: "Group 2"},
		}

		suite.dbStoreMock.On("GetOrganizationUnitGroupsList", "ou-1", 10, 0).
			Return(expectedGroups, nil).
			Once()

		result, err := suite.compositeStore.GetOrganizationUnitGroupsList("ou-1", 10, 0)
		suite.NoError(err)
		suite.Equal(expectedGroups, result)
	})

	suite.Run("propagates errors for user/group operations", func() {
		dbErr := errors.New("db error")

		suite.dbStoreMock.On("GetOrganizationUnitUsersList", "ou-1", 10, 0).
			Return([]User{}, dbErr).
			Once()

		users, err := suite.compositeStore.GetOrganizationUnitUsersList("ou-1", 10, 0)
		suite.Error(err)
		suite.Empty(users)

		suite.dbStoreMock.On("GetOrganizationUnitGroupsCount", "ou-1").
			Return(0, dbErr).
			Once()

		count, err := suite.compositeStore.GetOrganizationUnitGroupsCount("ou-1")
		suite.Error(err)
		suite.Equal(0, count)

		suite.dbStoreMock.On("GetOrganizationUnitGroupsList", "ou-1", 10, 0).
			Return([]Group{}, dbErr).
			Once()

		groups, err := suite.compositeStore.GetOrganizationUnitGroupsList("ou-1", 10, 0)
		suite.Error(err)
		suite.Empty(groups)
	})
}

// TestMergeAndDeduplicateChildren tests the merge helper function.
func (suite *CompositeStoreCoverageTestSuite) TestMergeAndDeduplicateChildren() {
	suite.Run("merges without duplicates", func() {
		dbChildren := []OrganizationUnitBasic{
			{ID: "db-1", Handle: "db1", Name: "DB 1"},
			{ID: "db-2", Handle: "db2", Name: "DB 2"},
		}
		fileChildren := []OrganizationUnitBasic{
			{ID: "file-1", Handle: "file1", Name: "File 1"},
		}

		result := mergeAndDeduplicateChildren(dbChildren, fileChildren)
		suite.Len(result, 3)
	})

	suite.Run("removes duplicates and DB takes precedence", func() {
		dbChildren := []OrganizationUnitBasic{
			{ID: "child-1", Handle: "db-handle", Name: "DB Child"},
		}
		fileChildren := []OrganizationUnitBasic{
			{ID: "child-1", Handle: "file-handle", Name: "File Child"},
		}

		result := mergeAndDeduplicateChildren(dbChildren, fileChildren)
		suite.Len(result, 1)
		suite.Equal("db-handle", result[0].Handle)
	})

	suite.Run("handles empty slices", func() {
		result := mergeAndDeduplicateChildren([]OrganizationUnitBasic{}, []OrganizationUnitBasic{})
		suite.Empty(result)

		dbChildren := []OrganizationUnitBasic{{ID: "db-1", Handle: "db1", Name: "DB 1"}}
		result = mergeAndDeduplicateChildren(dbChildren, []OrganizationUnitBasic{})
		suite.Len(result, 1)

		fileChildren := []OrganizationUnitBasic{{ID: "file-1", Handle: "file1", Name: "File 1"}}
		result = mergeAndDeduplicateChildren([]OrganizationUnitBasic{}, fileChildren)
		suite.Len(result, 1)
	})

	suite.Run("handles multiple duplicates", func() {
		dbChildren := []OrganizationUnitBasic{
			{ID: "child-1", Handle: "db1", Name: "DB 1"},
			{ID: "child-2", Handle: "db2", Name: "DB 2"},
		}
		fileChildren := []OrganizationUnitBasic{
			{ID: "child-1", Handle: "file1", Name: "File 1"},
			{ID: "child-2", Handle: "file2", Name: "File 2"},
			{ID: "child-3", Handle: "file3", Name: "File 3"},
		}

		result := mergeAndDeduplicateChildren(dbChildren, fileChildren)
		suite.Len(result, 3) // child-1, child-2 from DB, child-3 from file
		for _, child := range result {
			if child.ID == "child-1" {
				suite.Equal("db1", child.Handle)
			}
			if child.ID == "child-2" {
				suite.Equal("db2", child.Handle)
			}
			if child.ID == "child-3" {
				suite.Equal("file3", child.Handle)
			}
		}
	})
}
