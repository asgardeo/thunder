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

	"github.com/stretchr/testify/suite"
)

type OUHybridStoreTestSuite struct {
	suite.Suite
	fileStore *organizationUnitStoreInterfaceMock
	dbStore   *organizationUnitStoreInterfaceMock
	hybrid    organizationUnitStoreInterface
}

func TestOUHybridStoreTestSuite(t *testing.T) {
	suite.Run(t, new(OUHybridStoreTestSuite))
}

func (suite *OUHybridStoreTestSuite) SetupTest() {
	suite.fileStore = newOrganizationUnitStoreInterfaceMock(suite.T())
	suite.dbStore = newOrganizationUnitStoreInterfaceMock(suite.T())
	suite.hybrid = newOUHybridStore(suite.fileStore, suite.dbStore)
}

// Read Operation Tests

func (suite *OUHybridStoreTestSuite) TestHybridStore_GetOrganizationUnit_FromFileStore() {
	ouID := "file-ou-001"
	expectedOU := OrganizationUnit{
		ID:     ouID,
		Handle: "engineering",
		Name:   "Engineering",
	}

	suite.fileStore.On("GetOrganizationUnit", ouID).Return(expectedOU, nil).Once()

	result, err := suite.hybrid.GetOrganizationUnit(ouID)

	suite.Require().NoError(err)
	suite.Require().Equal(expectedOU, result)
	suite.fileStore.AssertExpectations(suite.T())
	suite.dbStore.AssertNotCalled(suite.T(), "GetOrganizationUnit")
}

func (suite *OUHybridStoreTestSuite) TestHybridStore_GetOrganizationUnit_FromDBStore() {
	ouID := "db-ou-001"
	expectedOU := OrganizationUnit{
		ID:     ouID,
		Handle: "sales",
		Name:   "Sales",
	}

	suite.fileStore.On("GetOrganizationUnit", ouID).Return(OrganizationUnit{}, ErrOrganizationUnitNotFound).Once()
	suite.dbStore.On("GetOrganizationUnit", ouID).Return(expectedOU, nil).Once()

	result, err := suite.hybrid.GetOrganizationUnit(ouID)

	suite.Require().NoError(err)
	suite.Require().Equal(expectedOU, result)
	suite.fileStore.AssertExpectations(suite.T())
	suite.dbStore.AssertExpectations(suite.T())
}

func (suite *OUHybridStoreTestSuite) TestHybridStore_GetOrganizationUnit_NotFound() {
	ouID := "non-existent"

	suite.fileStore.On("GetOrganizationUnit", ouID).Return(OrganizationUnit{}, ErrOrganizationUnitNotFound).Once()
	suite.dbStore.On("GetOrganizationUnit", ouID).Return(OrganizationUnit{}, ErrOrganizationUnitNotFound).Once()

	result, err := suite.hybrid.GetOrganizationUnit(ouID)

	suite.Require().Error(err)
	suite.Require().Equal(ErrOrganizationUnitNotFound, err)
	suite.Require().Equal(OrganizationUnit{}, result)
}

func (suite *OUHybridStoreTestSuite) TestHybridStore_GetOrganizationUnitList_MergedResults() {
	fileOUs := []OrganizationUnitBasic{
		{ID: "file-ou-1", Handle: "engineering", Name: "Engineering"},
		{ID: "file-ou-2", Handle: "product", Name: "Product"},
	}
	dbOUs := []OrganizationUnitBasic{
		{ID: "db-ou-1", Handle: "sales", Name: "Sales"},
		{ID: "db-ou-2", Handle: "marketing", Name: "Marketing"},
	}

	suite.fileStore.On("GetOrganizationUnitList", 9999, 0).Return(fileOUs, nil).Once()
	suite.dbStore.On("GetOrganizationUnitList", 9999, 0).Return(dbOUs, nil).Once()

	result, err := suite.hybrid.GetOrganizationUnitList(10, 0)

	suite.Require().NoError(err)
	suite.Require().Len(result, 4)
	// File-based OUs come first
	suite.Require().Equal("file-ou-1", result[0].ID)
	suite.Require().Equal("file-ou-2", result[1].ID)
	suite.Require().Equal("db-ou-1", result[2].ID)
	suite.Require().Equal("db-ou-2", result[3].ID)
}

func (suite *OUHybridStoreTestSuite) TestHybridStore_GetOrganizationUnitList_Pagination() {
	fileOUs := []OrganizationUnitBasic{
		{ID: "file-ou-1", Handle: "engineering", Name: "Engineering"},
		{ID: "file-ou-2", Handle: "product", Name: "Product"},
	}
	dbOUs := []OrganizationUnitBasic{
		{ID: "db-ou-1", Handle: "sales", Name: "Sales"},
	}

	suite.fileStore.On("GetOrganizationUnitList", 9999, 0).Return(fileOUs, nil).Once()
	suite.dbStore.On("GetOrganizationUnitList", 9999, 0).Return(dbOUs, nil).Once()

	// Get with limit 2, offset 1
	result, err := suite.hybrid.GetOrganizationUnitList(2, 1)

	suite.Require().NoError(err)
	suite.Require().Len(result, 2)
	suite.Require().Equal("file-ou-2", result[0].ID) // Second file OU
	suite.Require().Equal("db-ou-1", result[1].ID)   // First DB OU
}

func (suite *OUHybridStoreTestSuite) TestHybridStore_IsOrganizationUnitExists_InFileStore() {
	ouID := "file-ou-001"

	suite.fileStore.On("IsOrganizationUnitExists", ouID).Return(true, nil).Once()

	exists, err := suite.hybrid.IsOrganizationUnitExists(ouID)

	suite.Require().NoError(err)
	suite.Require().True(exists)
	suite.fileStore.AssertExpectations(suite.T())
	suite.dbStore.AssertNotCalled(suite.T(), "IsOrganizationUnitExists")
}

func (suite *OUHybridStoreTestSuite) TestHybridStore_IsOrganizationUnitExists_InDBStore() {
	ouID := "db-ou-001"

	suite.fileStore.On("IsOrganizationUnitExists", ouID).Return(false, nil).Once()
	suite.dbStore.On("IsOrganizationUnitExists", ouID).Return(true, nil).Once()

	exists, err := suite.hybrid.IsOrganizationUnitExists(ouID)

	suite.Require().NoError(err)
	suite.Require().True(exists)
	suite.fileStore.AssertExpectations(suite.T())
	suite.dbStore.AssertExpectations(suite.T())
}

func (suite *OUHybridStoreTestSuite) TestHybridStore_CheckNameConflict_InFileStore() {
	name := "Engineering"
	var parent *string

	suite.fileStore.On("CheckOrganizationUnitNameConflict", name, parent).Return(true, nil).Once()

	hasConflict, err := suite.hybrid.CheckOrganizationUnitNameConflict(name, parent)

	suite.Require().NoError(err)
	suite.Require().True(hasConflict)
	suite.fileStore.AssertExpectations(suite.T())
	suite.dbStore.AssertNotCalled(suite.T(), "CheckOrganizationUnitNameConflict")
}

func (suite *OUHybridStoreTestSuite) TestHybridStore_CheckNameConflict_InDBStore() {
	name := "Sales"
	var parent *string

	suite.fileStore.On("CheckOrganizationUnitNameConflict", name, parent).Return(false, nil).Once()
	suite.dbStore.On("CheckOrganizationUnitNameConflict", name, parent).Return(true, nil).Once()

	hasConflict, err := suite.hybrid.CheckOrganizationUnitNameConflict(name, parent)

	suite.Require().NoError(err)
	suite.Require().True(hasConflict)
	suite.fileStore.AssertExpectations(suite.T())
	suite.dbStore.AssertExpectations(suite.T())
}

// Write Operation Tests

func (suite *OUHybridStoreTestSuite) TestHybridStore_CreateOrganizationUnit_Success() {
	ou := OrganizationUnit{
		ID:     "new-ou-001",
		Handle: "support",
		Name:   "Support",
	}

	suite.fileStore.On("GetOrganizationUnit", ou.ID).Return(OrganizationUnit{}, ErrOrganizationUnitNotFound).Once()
	suite.fileStore.On("CheckOrganizationUnitHandleConflict", ou.Handle, ou.Parent).Return(false, nil).Once()
	suite.fileStore.On("CheckOrganizationUnitNameConflict", ou.Name, ou.Parent).Return(false, nil).Once()
	suite.dbStore.On("CreateOrganizationUnit", ou).Return(nil).Once()

	err := suite.hybrid.CreateOrganizationUnit(ou)

	suite.Require().NoError(err)
	suite.fileStore.AssertExpectations(suite.T())
	suite.dbStore.AssertExpectations(suite.T())
}

func (suite *OUHybridStoreTestSuite) TestHybridStore_CreateOrganizationUnit_ConflictWithFileStore() {
	ou := OrganizationUnit{
		ID:     "file-ou-001",
		Handle: "engineering",
		Name:   "Engineering",
	}

	suite.fileStore.On("GetOrganizationUnit", ou.ID).Return(ou, nil).Once()

	err := suite.hybrid.CreateOrganizationUnit(ou)

	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "ID conflicts with immutable file-based OU")
	suite.fileStore.AssertExpectations(suite.T())
	suite.dbStore.AssertNotCalled(suite.T(), "CreateOrganizationUnit")
}

func (suite *OUHybridStoreTestSuite) TestHybridStore_CreateOrganizationUnit_HandleConflict() {
	ou := OrganizationUnit{
		ID:     "new-ou-001",
		Handle: "engineering",
		Name:   "New Engineering",
	}

	suite.fileStore.On("GetOrganizationUnit", ou.ID).Return(OrganizationUnit{}, ErrOrganizationUnitNotFound).Once()
	suite.fileStore.On("CheckOrganizationUnitHandleConflict", ou.Handle, ou.Parent).Return(true, nil).Once()

	err := suite.hybrid.CreateOrganizationUnit(ou)

	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "handle conflicts with immutable file-based OU")
	suite.fileStore.AssertExpectations(suite.T())
	suite.dbStore.AssertNotCalled(suite.T(), "CreateOrganizationUnit")
}

func (suite *OUHybridStoreTestSuite) TestHybridStore_UpdateOrganizationUnit_Success() {
	ou := OrganizationUnit{
		ID:     "db-ou-001",
		Handle: "sales-updated",
		Name:   "Sales Updated",
	}

	suite.fileStore.On("GetOrganizationUnit", ou.ID).Return(OrganizationUnit{}, ErrOrganizationUnitNotFound).Once()
	suite.dbStore.On("UpdateOrganizationUnit", ou).Return(nil).Once()

	err := suite.hybrid.UpdateOrganizationUnit(ou)

	suite.Require().NoError(err)
	suite.fileStore.AssertExpectations(suite.T())
	suite.dbStore.AssertExpectations(suite.T())
}

func (suite *OUHybridStoreTestSuite) TestHybridStore_UpdateOrganizationUnit_RejectFileStoreOU() {
	ou := OrganizationUnit{
		ID:     "file-ou-001",
		Handle: "engineering",
		Name:   "Engineering Updated",
	}

	suite.fileStore.On("GetOrganizationUnit", ou.ID).Return(ou, nil).Once()

	err := suite.hybrid.UpdateOrganizationUnit(ou)

	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "cannot update organization unit: exists in immutable file-based store")
	suite.fileStore.AssertExpectations(suite.T())
	suite.dbStore.AssertNotCalled(suite.T(), "UpdateOrganizationUnit")
}

func (suite *OUHybridStoreTestSuite) TestHybridStore_DeleteOrganizationUnit_Success() {
	ouID := "db-ou-001"

	suite.fileStore.On("GetOrganizationUnit", ouID).Return(OrganizationUnit{}, ErrOrganizationUnitNotFound).Once()
	suite.dbStore.On("DeleteOrganizationUnit", ouID).Return(nil).Once()

	err := suite.hybrid.DeleteOrganizationUnit(ouID)

	suite.Require().NoError(err)
	suite.fileStore.AssertExpectations(suite.T())
	suite.dbStore.AssertExpectations(suite.T())
}

func (suite *OUHybridStoreTestSuite) TestHybridStore_DeleteOrganizationUnit_RejectFileStoreOU() {
	ouID := "file-ou-001"
	fileOU := OrganizationUnit{
		ID:     ouID,
		Handle: "engineering",
		Name:   "Engineering",
	}

	suite.fileStore.On("GetOrganizationUnit", ouID).Return(fileOU, nil).Once()

	err := suite.hybrid.DeleteOrganizationUnit(ouID)

	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "cannot delete organization unit: exists in immutable file-based store")
	suite.fileStore.AssertExpectations(suite.T())
	suite.dbStore.AssertNotCalled(suite.T(), "DeleteOrganizationUnit")
}

func (suite *OUHybridStoreTestSuite) TestHybridStore_GetOrganizationUnitChildren_MergedResults() {
	parentID := "parent-ou-001"
	fileChildren := []OrganizationUnitBasic{
		{ID: "file-child-1", Handle: "team-a", Name: "Team A"},
	}
	dbChildren := []OrganizationUnitBasic{
		{ID: "db-child-1", Handle: "team-b", Name: "Team B"},
	}

	suite.fileStore.On("GetOrganizationUnitChildrenList", parentID, 9999, 0).Return(fileChildren, nil).Once()
	suite.dbStore.On("GetOrganizationUnitChildrenList", parentID, 9999, 0).Return(dbChildren, nil).Once()

	result, err := suite.hybrid.GetOrganizationUnitChildrenList(parentID, 10, 0)

	suite.Require().NoError(err)
	suite.Require().Len(result, 2)
	suite.Require().Equal("file-child-1", result[0].ID)
	suite.Require().Equal("db-child-1", result[1].ID)
}

func (suite *OUHybridStoreTestSuite) TestHybridStore_UserOperations_DBOnly() {
	ouID := "test-ou-001"
	expectedUsers := []User{{ID: "user-1"}}

	suite.dbStore.On("GetOrganizationUnitUsersCount", ouID).Return(1, nil).Once()
	suite.dbStore.On("GetOrganizationUnitUsersList", ouID, 10, 0).Return(expectedUsers, nil).Once()

	count, err := suite.hybrid.GetOrganizationUnitUsersCount(ouID)
	suite.Require().NoError(err)
	suite.Require().Equal(1, count)

	users, err := suite.hybrid.GetOrganizationUnitUsersList(ouID, 10, 0)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedUsers, users)

	suite.dbStore.AssertExpectations(suite.T())
	suite.fileStore.AssertNotCalled(suite.T(), "GetOrganizationUnitUsersCount")
	suite.fileStore.AssertNotCalled(suite.T(), "GetOrganizationUnitUsersList")
}

func (suite *OUHybridStoreTestSuite) TestHybridStore_GroupOperations_DBOnly() {
	ouID := "test-ou-001"
	expectedGroups := []Group{{ID: "group-1", Name: "Group 1"}}

	suite.dbStore.On("GetOrganizationUnitGroupsCount", ouID).Return(1, nil).Once()
	suite.dbStore.On("GetOrganizationUnitGroupsList", ouID, 10, 0).Return(expectedGroups, nil).Once()

	count, err := suite.hybrid.GetOrganizationUnitGroupsCount(ouID)
	suite.Require().NoError(err)
	suite.Require().Equal(1, count)

	groups, err := suite.hybrid.GetOrganizationUnitGroupsList(ouID, 10, 0)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedGroups, groups)

	suite.dbStore.AssertExpectations(suite.T())
	suite.fileStore.AssertNotCalled(suite.T(), "GetOrganizationUnitGroupsCount")
	suite.fileStore.AssertNotCalled(suite.T(), "GetOrganizationUnitGroupsList")
}

func (suite *OUHybridStoreTestSuite) TestHybridStore_CheckHasChildResources_BothStores() {
	ouID := "test-ou-001"

	suite.fileStore.On("CheckOrganizationUnitHasChildResources", ouID).Return(false, nil).Once()
	suite.dbStore.On("CheckOrganizationUnitHasChildResources", ouID).Return(true, nil).Once()

	hasChildren, err := suite.hybrid.CheckOrganizationUnitHasChildResources(ouID)

	suite.Require().NoError(err)
	suite.Require().True(hasChildren)
	suite.fileStore.AssertExpectations(suite.T())
	suite.dbStore.AssertExpectations(suite.T())
}

// Error Handling Tests

func (suite *OUHybridStoreTestSuite) TestHybridStore_GetOrganizationUnit_FileStoreError() {
	ouID := "test-ou-001"

	suite.fileStore.On("GetOrganizationUnit", ouID).Return(OrganizationUnit{}, errors.New("file store error")).Once()

	result, err := suite.hybrid.GetOrganizationUnit(ouID)

	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "file store error")
	suite.Require().Equal(OrganizationUnit{}, result)
	suite.dbStore.AssertNotCalled(suite.T(), "GetOrganizationUnit")
}

func (suite *OUHybridStoreTestSuite) TestHybridStore_CreateOrganizationUnit_FileStoreCheckError() {
	ou := OrganizationUnit{
		ID:     "new-ou-001",
		Handle: "support",
		Name:   "Support",
	}

	suite.fileStore.On("GetOrganizationUnit", ou.ID).Return(OrganizationUnit{}, errors.New("check error")).Once()

	err := suite.hybrid.CreateOrganizationUnit(ou)

	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "check error")
	suite.dbStore.AssertNotCalled(suite.T(), "CreateOrganizationUnit")
}
