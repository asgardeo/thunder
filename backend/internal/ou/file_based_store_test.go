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
	"testing"

	"github.com/asgardeo/thunder/internal/system/file_based_runtime/entity"
	"github.com/stretchr/testify/suite"
)

type OUFileBasedStoreTestSuite struct {
	suite.Suite
	store organizationUnitStoreInterface
}

func TestOUFileBasedStoreTestSuite(t *testing.T) {
	suite.Run(t, new(OUFileBasedStoreTestSuite))
}

func (suite *OUFileBasedStoreTestSuite) SetupTest() {
	// Create a new store for each test
	storage := entity.NewStore()
	suite.store = &ouFileBasedStore{
		storage: storage,
	}
}

func (suite *OUFileBasedStoreTestSuite) TestCreateOrganizationUnit_Success() {
	ou := OrganizationUnit{
		ID:          "ou-001",
		Handle:      "engineering",
		Name:        "Engineering",
		Description: "Engineering department",
		Parent:      nil,
	}

	err := suite.store.CreateOrganizationUnit(ou)
	suite.NoError(err)

	// Verify it was stored
	retrieved, err := suite.store.GetOrganizationUnit("ou-001")
	suite.NoError(err)
	suite.Equal(ou.ID, retrieved.ID)
	suite.Equal(ou.Handle, retrieved.Handle)
	suite.Equal(ou.Name, retrieved.Name)
	suite.Equal(ou.Description, retrieved.Description)
	suite.Nil(retrieved.Parent)
}

func (suite *OUFileBasedStoreTestSuite) TestCreateOrganizationUnit_WithParent() {
	// Create parent OU
	parentOU := OrganizationUnit{
		ID:     "ou-parent",
		Handle: "engineering",
		Name:   "Engineering",
		Parent: nil,
	}
	err := suite.store.CreateOrganizationUnit(parentOU)
	suite.NoError(err)

	// Create child OU
	parentID := "ou-parent"
	childOU := OrganizationUnit{
		ID:     "ou-child",
		Handle: "backend",
		Name:   "Backend Team",
		Parent: &parentID,
	}
	err = suite.store.CreateOrganizationUnit(childOU)
	suite.NoError(err)

	// Verify child was stored with parent reference
	retrieved, err := suite.store.GetOrganizationUnit("ou-child")
	suite.NoError(err)
	suite.NotNil(retrieved.Parent)
	suite.Equal("ou-parent", *retrieved.Parent)
}

func (suite *OUFileBasedStoreTestSuite) TestGetOrganizationUnit_NotFound() {
	_, err := suite.store.GetOrganizationUnit("non-existent")
	suite.Error(err)
	suite.Equal(ErrOrganizationUnitNotFound, err)
}

func (suite *OUFileBasedStoreTestSuite) TestGetOrganizationUnitByPath_SingleLevel() {
	ou := OrganizationUnit{
		ID:     "ou-001",
		Handle: "engineering",
		Name:   "Engineering",
		Parent: nil,
	}
	err := suite.store.CreateOrganizationUnit(ou)
	suite.NoError(err)

	retrieved, err := suite.store.GetOrganizationUnitByPath([]string{"engineering"})
	suite.NoError(err)
	suite.Equal("ou-001", retrieved.ID)
	suite.Equal("engineering", retrieved.Handle)
}

func (suite *OUFileBasedStoreTestSuite) TestGetOrganizationUnitByPath_MultiLevel() {
	// Create parent
	parentOU := OrganizationUnit{
		ID:     "ou-parent",
		Handle: "engineering",
		Name:   "Engineering",
		Parent: nil,
	}
	err := suite.store.CreateOrganizationUnit(parentOU)
	suite.NoError(err)

	// Create child
	parentID := "ou-parent"
	childOU := OrganizationUnit{
		ID:     "ou-child",
		Handle: "backend",
		Name:   "Backend Team",
		Parent: &parentID,
	}
	err = suite.store.CreateOrganizationUnit(childOU)
	suite.NoError(err)

	// Create grandchild
	childID := "ou-child"
	grandchildOU := OrganizationUnit{
		ID:     "ou-grandchild",
		Handle: "api",
		Name:   "API Team",
		Parent: &childID,
	}
	err = suite.store.CreateOrganizationUnit(grandchildOU)
	suite.NoError(err)

	// Retrieve by path
	retrieved, err := suite.store.GetOrganizationUnitByPath([]string{"engineering", "backend", "api"})
	suite.NoError(err)
	suite.Equal("ou-grandchild", retrieved.ID)
	suite.Equal("api", retrieved.Handle)
}

func (suite *OUFileBasedStoreTestSuite) TestGetOrganizationUnitByPath_NotFound() {
	_, err := suite.store.GetOrganizationUnitByPath([]string{"non-existent"})
	suite.Error(err)
	suite.Equal(ErrOrganizationUnitNotFound, err)
}

func (suite *OUFileBasedStoreTestSuite) TestGetOrganizationUnitByPath_EmptyPath() {
	_, err := suite.store.GetOrganizationUnitByPath([]string{})
	suite.Error(err)
	suite.Equal(ErrOrganizationUnitNotFound, err)
}

func (suite *OUFileBasedStoreTestSuite) TestIsOrganizationUnitExists() {
	ou := OrganizationUnit{
		ID:     "ou-001",
		Handle: "engineering",
		Name:   "Engineering",
		Parent: nil,
	}
	err := suite.store.CreateOrganizationUnit(ou)
	suite.NoError(err)

	exists, err := suite.store.IsOrganizationUnitExists("ou-001")
	suite.NoError(err)
	suite.True(exists)

	exists, err = suite.store.IsOrganizationUnitExists("non-existent")
	suite.NoError(err)
	suite.False(exists)
}

func (suite *OUFileBasedStoreTestSuite) TestCheckOrganizationUnitNameConflict_SameParent() {
	parentID := "ou-parent"
	ou1 := OrganizationUnit{
		ID:     "ou-001",
		Handle: "team1",
		Name:   "Team One",
		Parent: &parentID,
	}
	err := suite.store.CreateOrganizationUnit(ou1)
	suite.NoError(err)

	// Check for conflict with same name and parent
	conflict, err := suite.store.CheckOrganizationUnitNameConflict("Team One", &parentID)
	suite.NoError(err)
	suite.True(conflict)

	// Check for no conflict with different name
	conflict, err = suite.store.CheckOrganizationUnitNameConflict("Team Two", &parentID)
	suite.NoError(err)
	suite.False(conflict)
}

func (suite *OUFileBasedStoreTestSuite) TestCheckOrganizationUnitNameConflict_DifferentParent() {
	parentID1 := "ou-parent1"
	ou1 := OrganizationUnit{
		ID:     "ou-001",
		Handle: "team1",
		Name:   "Team One",
		Parent: &parentID1,
	}
	err := suite.store.CreateOrganizationUnit(ou1)
	suite.NoError(err)

	// Check for conflict with same name but different parent (should not conflict)
	parentID2 := "ou-parent2"
	conflict, err := suite.store.CheckOrganizationUnitNameConflict("Team One", &parentID2)
	suite.NoError(err)
	suite.False(conflict)
}

func (suite *OUFileBasedStoreTestSuite) TestCheckOrganizationUnitNameConflict_RootLevel() {
	ou1 := OrganizationUnit{
		ID:     "ou-001",
		Handle: "engineering",
		Name:   "Engineering",
		Parent: nil,
	}
	err := suite.store.CreateOrganizationUnit(ou1)
	suite.NoError(err)

	// Check for conflict at root level
	conflict, err := suite.store.CheckOrganizationUnitNameConflict("Engineering", nil)
	suite.NoError(err)
	suite.True(conflict)

	// No conflict with different name
	conflict, err = suite.store.CheckOrganizationUnitNameConflict("Sales", nil)
	suite.NoError(err)
	suite.False(conflict)
}

func (suite *OUFileBasedStoreTestSuite) TestCheckOrganizationUnitHandleConflict_SameParent() {
	parentID := "ou-parent"
	ou1 := OrganizationUnit{
		ID:     "ou-001",
		Handle: "team1",
		Name:   "Team One",
		Parent: &parentID,
	}
	err := suite.store.CreateOrganizationUnit(ou1)
	suite.NoError(err)

	// Check for conflict with same handle and parent
	conflict, err := suite.store.CheckOrganizationUnitHandleConflict("team1", &parentID)
	suite.NoError(err)
	suite.True(conflict)

	// Check for no conflict with different handle
	conflict, err = suite.store.CheckOrganizationUnitHandleConflict("team2", &parentID)
	suite.NoError(err)
	suite.False(conflict)
}

func (suite *OUFileBasedStoreTestSuite) TestGetOrganizationUnitListCount() {
	// Create multiple OUs, some with parents
	ou1 := OrganizationUnit{ID: "ou-001", Handle: "eng", Name: "Engineering", Parent: nil}
	ou2 := OrganizationUnit{ID: "ou-002", Handle: "sales", Name: "Sales", Parent: nil}
	parentID := "ou-001"
	ou3 := OrganizationUnit{ID: "ou-003", Handle: "backend", Name: "Backend", Parent: &parentID}

	suite.NoError(suite.store.CreateOrganizationUnit(ou1))
	suite.NoError(suite.store.CreateOrganizationUnit(ou2))
	suite.NoError(suite.store.CreateOrganizationUnit(ou3))

	count, err := suite.store.GetOrganizationUnitListCount()
	suite.NoError(err)
	suite.Equal(3, count)
}

func (suite *OUFileBasedStoreTestSuite) TestGetOrganizationUnitList_Pagination() {
	// Create multiple root OUs
	for i := 1; i <= 5; i++ {
		ou := OrganizationUnit{
			ID:     "ou-00" + string(rune('0'+i)),
			Handle: "handle" + string(rune('0'+i)),
			Name:   "OU " + string(rune('0'+i)),
			Parent: nil,
		}
		suite.NoError(suite.store.CreateOrganizationUnit(ou))
	}

	// Test pagination
	list, err := suite.store.GetOrganizationUnitList(2, 0)
	suite.NoError(err)
	suite.Len(list, 2)

	list, err = suite.store.GetOrganizationUnitList(2, 2)
	suite.NoError(err)
	suite.Len(list, 2)

	list, err = suite.store.GetOrganizationUnitList(2, 4)
	suite.NoError(err)
	suite.Len(list, 1)

	// Offset beyond available items
	list, err = suite.store.GetOrganizationUnitList(2, 10)
	suite.NoError(err)
	suite.Len(list, 0)
}

func (suite *OUFileBasedStoreTestSuite) TestGetOrganizationUnitChildrenCount() {
	// Create parent
	parent := OrganizationUnit{ID: "ou-parent", Handle: "parent", Name: "Parent", Parent: nil}
	suite.NoError(suite.store.CreateOrganizationUnit(parent))

	// Create children
	parentID := "ou-parent"
	for i := 1; i <= 3; i++ {
		child := OrganizationUnit{
			ID:     "ou-child-00" + string(rune('0'+i)),
			Handle: "child" + string(rune('0'+i)),
			Name:   "Child " + string(rune('0'+i)),
			Parent: &parentID,
		}
		suite.NoError(suite.store.CreateOrganizationUnit(child))
	}

	count, err := suite.store.GetOrganizationUnitChildrenCount("ou-parent")
	suite.NoError(err)
	suite.Equal(3, count)

	// Non-existent parent should return 0
	count, err = suite.store.GetOrganizationUnitChildrenCount("non-existent")
	suite.NoError(err)
	suite.Equal(0, count)
}

func (suite *OUFileBasedStoreTestSuite) TestGetOrganizationUnitChildrenList_Pagination() {
	// Create parent
	parent := OrganizationUnit{ID: "ou-parent", Handle: "parent", Name: "Parent", Parent: nil}
	suite.NoError(suite.store.CreateOrganizationUnit(parent))

	// Create children
	parentID := "ou-parent"
	for i := 1; i <= 5; i++ {
		child := OrganizationUnit{
			ID:     "ou-child-00" + string(rune('0'+i)),
			Handle: "child" + string(rune('0'+i)),
			Name:   "Child " + string(rune('0'+i)),
			Parent: &parentID,
		}
		suite.NoError(suite.store.CreateOrganizationUnit(child))
	}

	// Test pagination
	list, err := suite.store.GetOrganizationUnitChildrenList("ou-parent", 2, 0)
	suite.NoError(err)
	suite.Len(list, 2)

	list, err = suite.store.GetOrganizationUnitChildrenList("ou-parent", 3, 2)
	suite.NoError(err)
	suite.Len(list, 3)

	list, err = suite.store.GetOrganizationUnitChildrenList("ou-parent", 10, 0)
	suite.NoError(err)
	suite.Len(list, 5)
}

func (suite *OUFileBasedStoreTestSuite) TestCheckOrganizationUnitHasChildResources() {
	// Create parent
	parent := OrganizationUnit{ID: "ou-parent", Handle: "parent", Name: "Parent", Parent: nil}
	suite.NoError(suite.store.CreateOrganizationUnit(parent))

	// Should not have children initially
	hasChildren, err := suite.store.CheckOrganizationUnitHasChildResources("ou-parent")
	suite.NoError(err)
	suite.False(hasChildren)

	// Add a child
	parentID := "ou-parent"
	child := OrganizationUnit{
		ID:     "ou-child",
		Handle: "child",
		Name:   "Child",
		Parent: &parentID,
	}
	suite.NoError(suite.store.CreateOrganizationUnit(child))

	// Should now have children
	hasChildren, err = suite.store.CheckOrganizationUnitHasChildResources("ou-parent")
	suite.NoError(err)
	suite.True(hasChildren)
}

func (suite *OUFileBasedStoreTestSuite) TestUpdateOrganizationUnit_NotSupported() {
	ou := OrganizationUnit{ID: "ou-001", Handle: "eng", Name: "Engineering", Parent: nil}
	err := suite.store.UpdateOrganizationUnit(ou)
	suite.Error(err)
	suite.Contains(err.Error(), "not supported")
}

func (suite *OUFileBasedStoreTestSuite) TestDeleteOrganizationUnit_NotSupported() {
	err := suite.store.DeleteOrganizationUnit("ou-001")
	suite.Error(err)
	suite.Contains(err.Error(), "not supported")
}

func (suite *OUFileBasedStoreTestSuite) TestGetOrganizationUnitUsersCount_AlwaysZero() {
	count, err := suite.store.GetOrganizationUnitUsersCount("ou-001")
	suite.NoError(err)
	suite.Equal(0, count)
}

func (suite *OUFileBasedStoreTestSuite) TestGetOrganizationUnitUsersList_AlwaysEmpty() {
	users, err := suite.store.GetOrganizationUnitUsersList("ou-001", 10, 0)
	suite.NoError(err)
	suite.Empty(users)
}

func (suite *OUFileBasedStoreTestSuite) TestGetOrganizationUnitGroupsCount_AlwaysZero() {
	count, err := suite.store.GetOrganizationUnitGroupsCount("ou-001")
	suite.NoError(err)
	suite.Equal(0, count)
}

func (suite *OUFileBasedStoreTestSuite) TestGetOrganizationUnitGroupsList_AlwaysEmpty() {
	groups, err := suite.store.GetOrganizationUnitGroupsList("ou-001", 10, 0)
	suite.NoError(err)
	suite.Empty(groups)
}
