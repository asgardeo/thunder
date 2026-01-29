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
	"strconv"
	"testing"

	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/declarative_resource/entity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	testParentOUID = "parent-1"
	testRootOUID   = "root-1"
)

type FileBasedStoreTestSuite struct {
	suite.Suite
	store *fileBasedStore
}

func TestFileBasedStoreTestSuite(t *testing.T) {
	suite.Run(t, new(FileBasedStoreTestSuite))
}

func (s *FileBasedStoreTestSuite) SetupTest() {
	// Create a file-based store with test instance
	genericStore := declarativeresource.NewGenericFileBasedStoreForTest(entity.KeyTypeOU)
	s.store = &fileBasedStore{
		GenericFileBasedStore: genericStore,
	}
}

func (s *FileBasedStoreTestSuite) TestCreateOrganizationUnit() {
	ou := OrganizationUnit{
		ID:          "test-ou-1",
		Handle:      "test",
		Name:        "Test OU",
		Description: "Test organization unit",
		Parent:      nil,
	}

	err := s.store.CreateOrganizationUnit(ou)
	assert.NoError(s.T(), err)

	// Verify it was created
	retrieved, err := s.store.GetOrganizationUnit("test-ou-1")
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), ou.ID, retrieved.ID)
	assert.Equal(s.T(), ou.Name, retrieved.Name)
	assert.Equal(s.T(), ou.Handle, retrieved.Handle)
}

func (s *FileBasedStoreTestSuite) TestGetOrganizationUnitNotFound() {
	_, err := s.store.GetOrganizationUnit("non-existent")
	assert.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, ErrOrganizationUnitNotFound)
}

func (s *FileBasedStoreTestSuite) TestGetOrganizationUnitList() {
	// Create root OUs
	ou1 := OrganizationUnit{
		ID:     "root-1",
		Handle: "root1",
		Name:   "Root 1",
		Parent: nil,
	}
	ou2 := OrganizationUnit{
		ID:     "root-2",
		Handle: "root2",
		Name:   "Root 2",
		Parent: nil,
	}

	err := s.store.CreateOrganizationUnit(ou1)
	assert.NoError(s.T(), err)
	err = s.store.CreateOrganizationUnit(ou2)
	assert.NoError(s.T(), err)

	// Create child OU (should not be in root list)
	parentID := testRootOUID
	child := OrganizationUnit{
		ID:     "child-1",
		Handle: "child1",
		Name:   "Child 1",
		Parent: &parentID,
	}
	err = s.store.CreateOrganizationUnit(child)
	assert.NoError(s.T(), err)

	// Get list should only return root OUs
	list, err := s.store.GetOrganizationUnitList(10, 0)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), list, 2)
}

func (s *FileBasedStoreTestSuite) TestUpdateNotSupported() {
	ou := OrganizationUnit{
		ID:     "test-ou-1",
		Handle: "test",
		Name:   "Test OU",
	}

	err := s.store.UpdateOrganizationUnit(ou)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "not supported")
}

func (s *FileBasedStoreTestSuite) TestDeleteNotSupported() {
	err := s.store.DeleteOrganizationUnit("test-ou-1")
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "not supported")
}

func (s *FileBasedStoreTestSuite) TestCheckOrganizationUnitNameConflict() {
	ou := OrganizationUnit{
		ID:     "test-ou-1",
		Handle: "test",
		Name:   "Test OU",
		Parent: nil,
	}

	err := s.store.CreateOrganizationUnit(ou)
	assert.NoError(s.T(), err)

	// Check for conflict with same name and parent
	conflict, err := s.store.CheckOrganizationUnitNameConflict("Test OU", nil)
	assert.NoError(s.T(), err)
	assert.True(s.T(), conflict)

	// No conflict with different name
	conflict, err = s.store.CheckOrganizationUnitNameConflict("Different Name", nil)
	assert.NoError(s.T(), err)
	assert.False(s.T(), conflict)
}

func (s *FileBasedStoreTestSuite) TestGetOrganizationUnitChildren() {
	// Create parent
	parent := OrganizationUnit{
		ID:     testParentOUID,
		Handle: "parent",
		Name:   "Parent OU",
		Parent: nil,
	}
	err := s.store.CreateOrganizationUnit(parent)
	assert.NoError(s.T(), err)

	// Create children
	parentID := testParentOUID
	child1 := OrganizationUnit{
		ID:     "child-1",
		Handle: "child1",
		Name:   "Child 1",
		Parent: &parentID,
	}
	child2 := OrganizationUnit{
		ID:     "child-2",
		Handle: "child2",
		Name:   "Child 2",
		Parent: &parentID,
	}

	err = s.store.CreateOrganizationUnit(child1)
	assert.NoError(s.T(), err)
	err = s.store.CreateOrganizationUnit(child2)
	assert.NoError(s.T(), err)

	// Get children
	children, err := s.store.GetOrganizationUnitChildrenList(testParentOUID, 10, 0)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), children, 2)

	// Get children count
	count, err := s.store.GetOrganizationUnitChildrenCount(testParentOUID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 2, count)
}

func (s *FileBasedStoreTestSuite) TestGetOrganizationUnitByPath() {
	// Create hierarchy: root -> engineering -> backend
	root := OrganizationUnit{
		ID:     "root-1",
		Handle: "root",
		Name:   "Root",
		Parent: nil,
	}
	rootID := "root-1"
	engineering := OrganizationUnit{
		ID:     "eng-1",
		Handle: "engineering",
		Name:   "Engineering",
		Parent: &rootID,
	}
	engID := "eng-1"
	backend := OrganizationUnit{
		ID:     "backend-1",
		Handle: "backend",
		Name:   "Backend",
		Parent: &engID,
	}

	err := s.store.CreateOrganizationUnit(root)
	assert.NoError(s.T(), err)
	err = s.store.CreateOrganizationUnit(engineering)
	assert.NoError(s.T(), err)
	err = s.store.CreateOrganizationUnit(backend)
	assert.NoError(s.T(), err)

	// Test getting by path
	ou, err := s.store.GetOrganizationUnitByPath([]string{"root"})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "root-1", ou.ID)

	ou, err = s.store.GetOrganizationUnitByPath([]string{"root", "engineering"})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "eng-1", ou.ID)

	ou, err = s.store.GetOrganizationUnitByPath([]string{"root", "engineering", "backend"})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "backend-1", ou.ID)
}

func (s *FileBasedStoreTestSuite) TestGetOrganizationUnitByPath_NotFound() {
	root := OrganizationUnit{
		ID:     "root-1",
		Handle: "root",
		Name:   "Root",
		Parent: nil,
	}
	err := s.store.CreateOrganizationUnit(root)
	assert.NoError(s.T(), err)

	// Test invalid path
	_, err = s.store.GetOrganizationUnitByPath([]string{"root", "nonexistent"})
	assert.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, ErrOrganizationUnitNotFound)

	// Test completely invalid path
	_, err = s.store.GetOrganizationUnitByPath([]string{"invalid"})
	assert.Error(s.T(), err)
	assert.ErrorIs(s.T(), err, ErrOrganizationUnitNotFound)
}

func (s *FileBasedStoreTestSuite) TestIsOrganizationUnitExists() {
	ou := OrganizationUnit{
		ID:     "test-ou-1",
		Handle: "test",
		Name:   "Test OU",
		Parent: nil,
	}
	err := s.store.CreateOrganizationUnit(ou)
	assert.NoError(s.T(), err)

	// Test existing OU
	exists, err := s.store.IsOrganizationUnitExists("test-ou-1")
	assert.NoError(s.T(), err)
	assert.True(s.T(), exists)

	// Test non-existent OU
	exists, err = s.store.IsOrganizationUnitExists("non-existent")
	assert.NoError(s.T(), err)
	assert.False(s.T(), exists)
}

func (s *FileBasedStoreTestSuite) TestCheckOrganizationUnitHandleConflict() {
	ou := OrganizationUnit{
		ID:     "test-ou-1",
		Handle: "test-handle",
		Name:   "Test OU",
		Parent: nil,
	}
	err := s.store.CreateOrganizationUnit(ou)
	assert.NoError(s.T(), err)

	// Check for conflict with same handle and parent
	conflict, err := s.store.CheckOrganizationUnitHandleConflict("test-handle", nil)
	assert.NoError(s.T(), err)
	assert.True(s.T(), conflict)

	// No conflict with different handle
	conflict, err = s.store.CheckOrganizationUnitHandleConflict("different-handle", nil)
	assert.NoError(s.T(), err)
	assert.False(s.T(), conflict)

	// Test with parent context
	parentID := testParentOUID
	child := OrganizationUnit{
		ID:     "child-1",
		Handle: "child-handle",
		Name:   "Child",
		Parent: &parentID,
	}
	err = s.store.CreateOrganizationUnit(child)
	assert.NoError(s.T(), err)

	conflict, err = s.store.CheckOrganizationUnitHandleConflict("child-handle", &parentID)
	assert.NoError(s.T(), err)
	assert.True(s.T(), conflict)

	// Different parent, same handle should not conflict
	differentParent := "different-parent"
	conflict, err = s.store.CheckOrganizationUnitHandleConflict("child-handle", &differentParent)
	assert.NoError(s.T(), err)
	assert.False(s.T(), conflict)
}

func (s *FileBasedStoreTestSuite) TestCheckOrganizationUnitNameConflict_WithParent() {
	parentID := testParentOUID
	parent := OrganizationUnit{
		ID:     parentID,
		Handle: "parent",
		Name:   "Parent",
		Parent: nil,
	}
	err := s.store.CreateOrganizationUnit(parent)
	assert.NoError(s.T(), err)

	child := OrganizationUnit{
		ID:     "child-1",
		Handle: "child",
		Name:   "Child Name",
		Parent: &parentID,
	}
	err = s.store.CreateOrganizationUnit(child)
	assert.NoError(s.T(), err)

	// Same name, same parent - should conflict
	conflict, err := s.store.CheckOrganizationUnitNameConflict("Child Name", &parentID)
	assert.NoError(s.T(), err)
	assert.True(s.T(), conflict)

	// Same name, different parent - should not conflict
	differentParent := "different-parent"
	conflict, err = s.store.CheckOrganizationUnitNameConflict("Child Name", &differentParent)
	assert.NoError(s.T(), err)
	assert.False(s.T(), conflict)

	// Same name, nil parent vs actual parent - should not conflict
	conflict, err = s.store.CheckOrganizationUnitNameConflict("Child Name", nil)
	assert.NoError(s.T(), err)
	assert.False(s.T(), conflict)
}

func (s *FileBasedStoreTestSuite) TestCheckOrganizationUnitHasChildResources() {
	parent := OrganizationUnit{
		ID:     testParentOUID,
		Handle: "parent",
		Name:   "Parent",
		Parent: nil,
	}
	err := s.store.CreateOrganizationUnit(parent)
	assert.NoError(s.T(), err)

	// No children initially
	hasChildren, err := s.store.CheckOrganizationUnitHasChildResources(testParentOUID)
	assert.NoError(s.T(), err)
	assert.False(s.T(), hasChildren)

	// Add a child
	parentID := testParentOUID
	child := OrganizationUnit{
		ID:     "child-1",
		Handle: "child",
		Name:   "Child",
		Parent: &parentID,
	}
	err = s.store.CreateOrganizationUnit(child)
	assert.NoError(s.T(), err)

	// Should have children now
	hasChildren, err = s.store.CheckOrganizationUnitHasChildResources(testParentOUID)
	assert.NoError(s.T(), err)
	assert.True(s.T(), hasChildren)
}

func (s *FileBasedStoreTestSuite) TestGetOrganizationUnitListCount() {
	// Initially empty
	count, err := s.store.GetOrganizationUnitListCount()
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 0, count)

	// Add root OUs
	root1 := OrganizationUnit{
		ID:     "root-1",
		Handle: "root1",
		Name:   "Root 1",
		Parent: nil,
	}
	root2 := OrganizationUnit{
		ID:     "root-2",
		Handle: "root2",
		Name:   "Root 2",
		Parent: nil,
	}
	err = s.store.CreateOrganizationUnit(root1)
	assert.NoError(s.T(), err)
	err = s.store.CreateOrganizationUnit(root2)
	assert.NoError(s.T(), err)

	// Should count only root OUs
	count, err = s.store.GetOrganizationUnitListCount()
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 2, count)

	// Add child OU (should not be counted)
	parentID := "root-1"
	child := OrganizationUnit{
		ID:     "child-1",
		Handle: "child",
		Name:   "Child",
		Parent: &parentID,
	}
	err = s.store.CreateOrganizationUnit(child)
	assert.NoError(s.T(), err)

	// Count should still be 2
	count, err = s.store.GetOrganizationUnitListCount()
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 2, count)
}

func (s *FileBasedStoreTestSuite) TestGetOrganizationUnitList_Pagination() {
	// Create multiple root OUs
	for i := 1; i <= 5; i++ {
		iStr := strconv.Itoa(i)
		ou := OrganizationUnit{
			ID:     "root-" + iStr,
			Handle: "root" + iStr,
			Name:   "Root " + iStr,
			Parent: nil,
		}
		err := s.store.CreateOrganizationUnit(ou)
		assert.NoError(s.T(), err)
	}

	// Test pagination - first page
	list, err := s.store.GetOrganizationUnitList(2, 0)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), list, 2)

	// Test pagination - second page
	list, err = s.store.GetOrganizationUnitList(2, 2)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), list, 2)

	// Test pagination - last page
	list, err = s.store.GetOrganizationUnitList(2, 4)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), list, 1)

	// Test offset beyond range
	list, err = s.store.GetOrganizationUnitList(10, 100)
	assert.NoError(s.T(), err)
	assert.Empty(s.T(), list)
}

func (s *FileBasedStoreTestSuite) TestGetOrganizationUnitChildrenList_Pagination() {
	// Create parent
	parent := OrganizationUnit{
		ID:     testParentOUID,
		Handle: "parent",
		Name:   "Parent",
		Parent: nil,
	}
	err := s.store.CreateOrganizationUnit(parent)
	assert.NoError(s.T(), err)

	// Create multiple children
	parentID := testParentOUID
	for i := 1; i <= 5; i++ {
		iStr := strconv.Itoa(i)
		child := OrganizationUnit{
			ID:     "child-" + iStr,
			Handle: "child" + iStr,
			Name:   "Child " + iStr,
			Parent: &parentID,
		}
		err := s.store.CreateOrganizationUnit(child)
		assert.NoError(s.T(), err)
	}

	// Test pagination
	children, err := s.store.GetOrganizationUnitChildrenList(testParentOUID, 2, 0)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), children, 2)

	children, err = s.store.GetOrganizationUnitChildrenList(testParentOUID, 2, 2)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), children, 2)

	// Test offset beyond range
	children, err = s.store.GetOrganizationUnitChildrenList(testParentOUID, 10, 100)
	assert.NoError(s.T(), err)
	assert.Empty(s.T(), children)
}

func (s *FileBasedStoreTestSuite) TestGetOrganizationUnitUsersCount() {
	// Users are not stored in file-based mode
	count, err := s.store.GetOrganizationUnitUsersCount("any-id")
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 0, count)
}

func (s *FileBasedStoreTestSuite) TestGetOrganizationUnitUsersList() {
	// Users are not stored in file-based mode
	users, err := s.store.GetOrganizationUnitUsersList("any-id", 10, 0)
	assert.NoError(s.T(), err)
	assert.Empty(s.T(), users)
}

func (s *FileBasedStoreTestSuite) TestGetOrganizationUnitGroupsCount() {
	// Groups are not stored in file-based mode
	count, err := s.store.GetOrganizationUnitGroupsCount("any-id")
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 0, count)
}

func (s *FileBasedStoreTestSuite) TestGetOrganizationUnitGroupsList() {
	// Groups are not stored in file-based mode
	groups, err := s.store.GetOrganizationUnitGroupsList("any-id", 10, 0)
	assert.NoError(s.T(), err)
	assert.Empty(s.T(), groups)
}

func (s *FileBasedStoreTestSuite) TestCreate_StorerInterface() {
	ou := &OrganizationUnit{
		ID:     "test-ou-1",
		Handle: "test",
		Name:   "Test OU",
		Parent: nil,
	}

	// Test the Create method from Storer interface
	err := s.store.Create("test-ou-1", ou)
	assert.NoError(s.T(), err)

	// Verify it was created
	retrieved, err := s.store.GetOrganizationUnit("test-ou-1")
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), ou.ID, retrieved.ID)
}

func (s *FileBasedStoreTestSuite) TestNewFileBasedStore() {
	// Test that newFileBasedStore creates a valid store
	store := newFileBasedStore()
	assert.NotNil(s.T(), store)

	// Verify it implements the interface by using it
	fbStore, ok := store.(*fileBasedStore)
	assert.True(s.T(), ok)
	assert.NotNil(s.T(), fbStore.GenericFileBasedStore)
}
