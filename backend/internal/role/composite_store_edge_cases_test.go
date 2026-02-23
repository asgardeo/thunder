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

package role

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// CompositeRoleStoreEdgeCaseTestSuite contains edge case tests for the composite role store.
type CompositeRoleStoreEdgeCaseTestSuite struct {
	suite.Suite
	mockDBStore   *roleStoreInterfaceMock
	mockFileStore *roleStoreInterfaceMock
	store         roleStoreInterface
	ctx           context.Context
}

func TestCompositeRoleStoreEdgeCaseTestSuite(t *testing.T) {
	suite.Run(t, new(CompositeRoleStoreEdgeCaseTestSuite))
}

func (suite *CompositeRoleStoreEdgeCaseTestSuite) SetupTest() {
	suite.mockDBStore = newRoleStoreInterfaceMock(suite.T())
	suite.mockFileStore = newRoleStoreInterfaceMock(suite.T())
	suite.store = newCompositeRoleStore(suite.mockFileStore, suite.mockDBStore)
	suite.ctx = context.Background()
}

// Test CreateRole delegates to database store
func (suite *CompositeRoleStoreEdgeCaseTestSuite) TestCreateRole_DelegatesToDB() {
	suite.mockDBStore.On("CreateRole", suite.ctx, "role1", mock.Anything).Return(nil)

	err := suite.store.CreateRole(suite.ctx, "role1", RoleCreationDetail{
		Name:               "Test",
		OrganizationUnitID: "ou1",
	})

	assert.NoError(suite.T(), err)
	suite.mockDBStore.AssertExpectations(suite.T())
	suite.mockFileStore.AssertNotCalled(suite.T(), "CreateRole")
}

// Test GetRole from DB when found
func (suite *CompositeRoleStoreEdgeCaseTestSuite) TestGetRole_FromDB() {
	expectedRole := RoleWithPermissions{
		ID:   "role1",
		Name: "Admin",
	}
	suite.mockDBStore.On("GetRole", suite.ctx, "role1").Return(expectedRole, nil)

	result, err := suite.store.GetRole(suite.ctx, "role1")

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedRole, result)
	suite.mockDBStore.AssertExpectations(suite.T())
	suite.mockFileStore.AssertNotCalled(suite.T(), "GetRole")
}

// Test GetRole falls back to file store when not in DB
func (suite *CompositeRoleStoreEdgeCaseTestSuite) TestGetRole_FallbackToFile() {
	expectedRole := RoleWithPermissions{
		ID:   "role1",
		Name: "Admin",
	}
	suite.mockDBStore.On("GetRole", suite.ctx, "role1").Return(RoleWithPermissions{}, ErrRoleNotFound)
	suite.mockFileStore.On("GetRole", suite.ctx, "role1").Return(expectedRole, nil)

	result, err := suite.store.GetRole(suite.ctx, "role1")

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedRole, result)
	suite.mockDBStore.AssertExpectations(suite.T())
	suite.mockFileStore.AssertExpectations(suite.T())
}

// Test GetRole returns DB error when not found in either store
func (suite *CompositeRoleStoreEdgeCaseTestSuite) TestGetRole_NotFound() {
	suite.mockDBStore.On("GetRole", suite.ctx, "nonexistent").Return(RoleWithPermissions{}, ErrRoleNotFound)
	suite.mockFileStore.On("GetRole", suite.ctx, "nonexistent").Return(RoleWithPermissions{}, ErrRoleNotFound)

	result, err := suite.store.GetRole(suite.ctx, "nonexistent")

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), RoleWithPermissions{}, result)
}

// Test GetRole returns DB error when DB has error other than not found
func (suite *CompositeRoleStoreEdgeCaseTestSuite) TestGetRole_DBError() {
	dbErr := errors.New("database connection error")
	suite.mockDBStore.On("GetRole", suite.ctx, "role1").Return(RoleWithPermissions{}, dbErr)

	result, err := suite.store.GetRole(suite.ctx, "role1")

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), dbErr, err)
	assert.Equal(suite.T(), RoleWithPermissions{}, result)
	suite.mockFileStore.AssertNotCalled(suite.T(), "GetRole")
}

// Test UpdateRole delegates to database store
func (suite *CompositeRoleStoreEdgeCaseTestSuite) TestUpdateRole_DelegatesToDB() {
	suite.mockDBStore.On("UpdateRole", suite.ctx, "role1", mock.Anything).Return(nil)

	err := suite.store.UpdateRole(suite.ctx, "role1", RoleUpdateDetail{
		Name: "Updated",
	})

	assert.NoError(suite.T(), err)
	suite.mockDBStore.AssertExpectations(suite.T())
	suite.mockFileStore.AssertNotCalled(suite.T(), "UpdateRole")
}

// Test DeleteRole delegates to database store
func (suite *CompositeRoleStoreEdgeCaseTestSuite) TestDeleteRole_DelegatesToDB() {
	suite.mockDBStore.On("DeleteRole", suite.ctx, "role1").Return(nil)

	err := suite.store.DeleteRole(suite.ctx, "role1")

	assert.NoError(suite.T(), err)
	suite.mockDBStore.AssertExpectations(suite.T())
	suite.mockFileStore.AssertNotCalled(suite.T(), "DeleteRole")
}

// Test AddAssignments delegates to database store
func (suite *CompositeRoleStoreEdgeCaseTestSuite) TestAddAssignments_DelegatesToDB() {
	suite.mockDBStore.On("AddAssignments", suite.ctx, "role1", mock.Anything).Return(nil)

	err := suite.store.AddAssignments(suite.ctx, "role1", []RoleAssignment{
		{ID: "user1", Type: AssigneeTypeUser},
	})

	assert.NoError(suite.T(), err)
	suite.mockDBStore.AssertExpectations(suite.T())
}

// Test RemoveAssignments delegates to database store
func (suite *CompositeRoleStoreEdgeCaseTestSuite) TestRemoveAssignments_DelegatesToDB() {
	suite.mockDBStore.On("RemoveAssignments", suite.ctx, "role1", mock.Anything).Return(nil)

	err := suite.store.RemoveAssignments(suite.ctx, "role1", []RoleAssignment{
		{ID: "user1", Type: AssigneeTypeUser},
	})

	assert.NoError(suite.T(), err)
	suite.mockDBStore.AssertExpectations(suite.T())
}

// Test CheckRoleNameExists checks file store first, returns true if found
func (suite *CompositeRoleStoreEdgeCaseTestSuite) TestCheckRoleNameExists_ChecksBothStores() {
	// CompositeBooleanCheckHelper checks fileStore first. If it returns true, it stops.
	suite.mockFileStore.On("CheckRoleNameExists", suite.ctx, "ou1", "Admin").Return(true, nil)
	// DBStore should not be called since fileStore returns true

	exists, err := suite.store.CheckRoleNameExists(suite.ctx, "ou1", "Admin")

	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)
	suite.mockFileStore.AssertExpectations(suite.T())
	suite.mockDBStore.AssertNotCalled(suite.T(), "CheckRoleNameExists")
}

// Test CheckRoleNameExists returns true if found in DB
func (suite *CompositeRoleStoreEdgeCaseTestSuite) TestCheckRoleNameExists_FoundInDB() {
	suite.mockDBStore.On("CheckRoleNameExists", suite.ctx, "ou1", "Admin").Return(true, nil)
	suite.mockFileStore.On("CheckRoleNameExists", suite.ctx, "ou1", "Admin").Return(false, nil)

	exists, err := suite.store.CheckRoleNameExists(suite.ctx, "ou1", "Admin")

	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)
}

// Test CheckRoleNameExists returns true if found in file store
func (suite *CompositeRoleStoreEdgeCaseTestSuite) TestCheckRoleNameExists_FoundInFile() {
	// FileStore returns true, so DBStore is not called
	suite.mockFileStore.On("CheckRoleNameExists", suite.ctx, "ou1", "Admin").Return(true, nil)

	exists, err := suite.store.CheckRoleNameExists(suite.ctx, "ou1", "Admin")

	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)
	suite.mockFileStore.AssertExpectations(suite.T())
	suite.mockDBStore.AssertNotCalled(suite.T(), "CheckRoleNameExists")
}

// Test CheckRoleNameExistsExcludingID checks both stores
func (suite *CompositeRoleStoreEdgeCaseTestSuite) TestCheckRoleNameExistsExcludingID_ChecksBothStores() {
	suite.mockDBStore.On("CheckRoleNameExistsExcludingID", suite.ctx, "ou1", "Admin", "role1").Return(false, nil)
	suite.mockFileStore.On("CheckRoleNameExistsExcludingID", suite.ctx, "ou1", "Admin", "role1").Return(false, nil)

	exists, err := suite.store.CheckRoleNameExistsExcludingID(suite.ctx, "ou1", "Admin", "role1")

	assert.NoError(suite.T(), err)
	assert.False(suite.T(), exists)
}

// Test IsRoleExist checks file store first (uses CompositeBooleanCheckHelper)
func (suite *CompositeRoleStoreEdgeCaseTestSuite) TestIsRoleExist_ChecksBothStores() {
	// CompositeBooleanCheckHelper checks fileStore first. If true, returns without checking dbStore.
	suite.mockFileStore.On("IsRoleExist", suite.ctx, "role1").Return(true, nil)

	exists, err := suite.store.IsRoleExist(suite.ctx, "role1")

	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)
	suite.mockFileStore.AssertExpectations(suite.T())
	suite.mockDBStore.AssertNotCalled(suite.T(), "IsRoleExist")
}

// Test IsRoleExist returns true if found in DB
func (suite *CompositeRoleStoreEdgeCaseTestSuite) TestIsRoleExist_FoundInDB() {
	suite.mockDBStore.On("IsRoleExist", suite.ctx, "role1").Return(true, nil)
	suite.mockFileStore.On("IsRoleExist", suite.ctx, "role1").Return(false, nil)

	exists, err := suite.store.IsRoleExist(suite.ctx, "role1")

	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)
}

// Test IsRoleExist returns true if found in file store
func (suite *CompositeRoleStoreEdgeCaseTestSuite) TestIsRoleExist_FoundInFile() {
	// FileStore returns true, so DBStore is not called
	suite.mockFileStore.On("IsRoleExist", suite.ctx, "role1").Return(true, nil)

	exists, err := suite.store.IsRoleExist(suite.ctx, "role1")

	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)
	suite.mockFileStore.AssertExpectations(suite.T())
	suite.mockDBStore.AssertNotCalled(suite.T(), "IsRoleExist")
}

// Test GetRoleListCount merges and deduplicates
func (suite *CompositeRoleStoreEdgeCaseTestSuite) TestGetRoleListCount_MergesAndDeduplicates() {
	dbRoles := []Role{
		{ID: "role1", Name: "Admin"},
		{ID: "role2", Name: "Editor"},
	}
	fileRoles := []Role{
		{ID: "role2", Name: "Editor"},
		{ID: "role3", Name: "Viewer"},
	}

	// GetRoleListCount first calls GetRoleListCount on both stores
	suite.mockDBStore.On("GetRoleListCount", suite.ctx).Return(2, nil)
	suite.mockFileStore.On("GetRoleListCount", suite.ctx).Return(3, nil)
	// Then calls GetRoleList with the counts as limits and 0 offset
	suite.mockDBStore.On("GetRoleList", suite.ctx, 2, 0).Return(dbRoles, nil)
	suite.mockFileStore.On("GetRoleList", suite.ctx, 3, 0).Return(fileRoles, nil)

	count, err := suite.store.GetRoleListCount(suite.ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3, count)
}

// Test GetRoleList merges and applies pagination
func (suite *CompositeRoleStoreEdgeCaseTestSuite) TestGetRoleList_MergesAndPaginates() {
	dbRoles := []Role{
		{ID: "role1", Name: "Admin"},
		{ID: "role2", Name: "Editor"},
	}
	fileRoles := []Role{
		{ID: "role3", Name: "Viewer"},
		{ID: "role4", Name: "Guest"},
	}

	// GetRoleList calls GetRoleListCount first, then GetRoleList with the counts
	suite.mockDBStore.On("GetRoleListCount", suite.ctx).Return(2, nil)
	suite.mockFileStore.On("GetRoleListCount", suite.ctx).Return(2, nil)
	suite.mockDBStore.On("GetRoleList", suite.ctx, 2, 0).Return(dbRoles, nil)
	suite.mockFileStore.On("GetRoleList", suite.ctx, 2, 0).Return(fileRoles, nil)

	// Test page 1
	result, err := suite.store.GetRoleList(suite.ctx, 2, 0)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)

	// For the second page test, need fresh mock setup
	suite.mockDBStore.On("GetRoleListCount", suite.ctx).Return(2, nil)
	suite.mockFileStore.On("GetRoleListCount", suite.ctx).Return(2, nil)
	suite.mockDBStore.On("GetRoleList", suite.ctx, 2, 0).Return(dbRoles, nil)
	suite.mockFileStore.On("GetRoleList", suite.ctx, 2, 0).Return(fileRoles, nil)

	// Test page 2
	result, err = suite.store.GetRoleList(suite.ctx, 2, 2)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
}

// Test GetRoleList returns empty when offset exceeds results
func (suite *CompositeRoleStoreEdgeCaseTestSuite) TestGetRoleList_OffsetBeyondResults() {
	dbRoles := []Role{
		{ID: "role1", Name: "Admin"},
	}
	fileRoles := []Role{}

	suite.mockDBStore.On("GetRoleListCount", suite.ctx).Return(1, nil)
	suite.mockFileStore.On("GetRoleListCount", suite.ctx).Return(0, nil)
	suite.mockDBStore.On("GetRoleList", suite.ctx, 1, 0).Return(dbRoles, nil)
	suite.mockFileStore.On("GetRoleList", suite.ctx, 0, 0).Return(fileRoles, nil)

	result, err := suite.store.GetRoleList(suite.ctx, 10, 100)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 0)
}

// Test GetRoleAssignmentsCount merges and deduplicates
func (suite *CompositeRoleStoreEdgeCaseTestSuite) TestGetRoleAssignmentsCount_MergesAndDeduplicates() {
	dbAssignments := []RoleAssignment{
		{ID: "user1", Type: AssigneeTypeUser},
		{ID: "group1", Type: AssigneeTypeGroup},
		{ID: "user2", Type: AssigneeTypeUser},
	}
	fileAssignments := []RoleAssignment{
		{ID: "user2", Type: AssigneeTypeUser},
		{ID: "group2", Type: AssigneeTypeGroup},
	}

	// GetRoleAssignmentsCount calls GetRoleAssignmentsCount first
	suite.mockDBStore.On("GetRoleAssignmentsCount", suite.ctx, "role1").Return(3, nil)
	suite.mockFileStore.On("GetRoleAssignmentsCount", suite.ctx, "role1").Return(2, nil)
	// Then calls GetRoleAssignments with the counts
	suite.mockDBStore.On("GetRoleAssignments", suite.ctx, "role1", 3, 0).Return(dbAssignments, nil)
	suite.mockFileStore.On("GetRoleAssignments", suite.ctx, "role1", 2, 0).Return(fileAssignments, nil)

	count, err := suite.store.GetRoleAssignmentsCount(suite.ctx, "role1")

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 4, count)
}

// Test GetRoleAssignments merges and applies pagination
func (suite *CompositeRoleStoreEdgeCaseTestSuite) TestGetRoleAssignments_MergesAndPaginates() {
	dbAssignments := []RoleAssignment{
		{ID: "user1", Type: AssigneeTypeUser},
		{ID: "user2", Type: AssigneeTypeUser},
	}
	fileAssignments := []RoleAssignment{
		{ID: "group1", Type: AssigneeTypeGroup},
		{ID: "group2", Type: AssigneeTypeGroup},
	}

	// GetRoleAssignments calls GetRoleAssignmentsCount first
	suite.mockDBStore.On("GetRoleAssignmentsCount", suite.ctx, "role1").Return(2, nil)
	suite.mockFileStore.On("GetRoleAssignmentsCount", suite.ctx, "role1").Return(2, nil)
	// Then calls GetRoleAssignments with the counts
	suite.mockDBStore.On("GetRoleAssignments", suite.ctx, "role1", 2, 0).Return(dbAssignments, nil)
	suite.mockFileStore.On("GetRoleAssignments", suite.ctx, "role1", 2, 0).Return(fileAssignments, nil)

	result, err := suite.store.GetRoleAssignments(suite.ctx, "role1", 2, 1)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
}

// Test IsRoleDeclarative checks file store
func (suite *CompositeRoleStoreEdgeCaseTestSuite) TestIsRoleDeclarative_ChecksFileStore() {
	suite.mockFileStore.On("IsRoleExist", suite.ctx, "role1").Return(true, nil)

	isDeclarative, err := suite.store.IsRoleDeclarative(suite.ctx, "role1")

	assert.NoError(suite.T(), err)
	assert.True(suite.T(), isDeclarative)
}

// Test IsRoleDeclarative returns false for non-existent role
func (suite *CompositeRoleStoreEdgeCaseTestSuite) TestIsRoleDeclarative_NonExistent() {
	suite.mockFileStore.On("IsRoleExist", suite.ctx, "nonexistent").Return(false, nil)

	isDeclarative, err := suite.store.IsRoleDeclarative(suite.ctx, "nonexistent")

	assert.NoError(suite.T(), err)
	assert.False(suite.T(), isDeclarative)
}

// Test GetAuthorizedPermissions checks both stores
func (suite *CompositeRoleStoreEdgeCaseTestSuite) TestGetAuthorizedPermissions_ChecksBothStores() {
	perms := []string{"perm1", "perm2"}
	suite.mockDBStore.On(
		"GetAuthorizedPermissions", suite.ctx, "user1", []string{"group1"}, perms,
	).Return([]string{"perm1"}, nil)
	suite.mockFileStore.On(
		"GetAuthorizedPermissions", suite.ctx, "user1", []string{"group1"}, perms,
	).Return([]string{"perm1", "perm2"}, nil)

	result, err := suite.store.GetAuthorizedPermissions(
		suite.ctx, "user1", []string{"group1"}, perms,
	)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), []string{"perm1", "perm2"}, result)
}

// Test GetAuthorizedPermissions merges permissions from both stores (union)
func (suite *CompositeRoleStoreEdgeCaseTestSuite) TestGetAuthorizedPermissions_CommonPermissions() {
	perms := []string{"p1", "p2", "p3"}
	suite.mockDBStore.On(
		"GetAuthorizedPermissions", suite.ctx, "user1", []string{"group1"}, perms,
	).Return([]string{"p1", "p2"}, nil)
	suite.mockFileStore.On(
		"GetAuthorizedPermissions", suite.ctx, "user1", []string{"group1"}, perms,
	).Return([]string{"p2", "p3"}, nil)

	result, err := suite.store.GetAuthorizedPermissions(
		suite.ctx, "user1", []string{"group1"}, perms,
	)

	assert.NoError(suite.T(), err)
	// mergePermissions returns union of all unique permissions
	assert.Len(suite.T(), result, 3)
	assert.Contains(suite.T(), result, "p1")
	assert.Contains(suite.T(), result, "p2")
	assert.Contains(suite.T(), result, "p3")
}

// Test GetAuthorizedPermissions with empty result
func (suite *CompositeRoleStoreEdgeCaseTestSuite) TestGetAuthorizedPermissions_EmptyResult() {
	perms := []string{"perm1"}
	suite.mockDBStore.On(
		"GetAuthorizedPermissions", suite.ctx, "user1", []string{}, perms,
	).Return([]string{}, nil)
	suite.mockFileStore.On(
		"GetAuthorizedPermissions", suite.ctx, "user1", []string{}, perms,
	).Return([]string{}, nil)

	result, err := suite.store.GetAuthorizedPermissions(
		suite.ctx, "user1", []string{}, perms,
	)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 0)
}

// Test DB precedence in deduplication
func (suite *CompositeRoleStoreEdgeCaseTestSuite) TestMergeAndDeduplicateRoles_DBPrecedence() {
	dbRoles := []Role{
		{ID: "role1", Name: "AdminDB"},
	}
	fileRoles := []Role{
		{ID: "role1", Name: "AdminFile"},
	}

	suite.mockDBStore.On("GetRoleListCount", suite.ctx).Return(1, nil)
	suite.mockFileStore.On("GetRoleListCount", suite.ctx).Return(1, nil)
	suite.mockDBStore.On("GetRoleList", suite.ctx, 1, 0).Return(dbRoles, nil)
	suite.mockFileStore.On("GetRoleList", suite.ctx, 1, 0).Return(fileRoles, nil)

	result, err := suite.store.GetRoleList(suite.ctx, 10, 0)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 1)
	assert.Equal(suite.T(), "AdminDB", result[0].Name)
}

// Test DB error propagation in GetRoleList
func (suite *CompositeRoleStoreEdgeCaseTestSuite) TestGetRoleList_PropagatesDBError() {
	dbErr := errors.New("database error")
	suite.mockDBStore.On("GetRoleListCount", suite.ctx).Return(0, dbErr)

	result, err := suite.store.GetRoleList(suite.ctx, 10, 0)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), dbErr, err)
}

// Test file store error propagation in GetRoleList
func (suite *CompositeRoleStoreEdgeCaseTestSuite) TestGetRoleList_PropagatesFileError() {
	fileErr := errors.New("file store error")
	suite.mockDBStore.On("GetRoleListCount", suite.ctx).Return(1, nil)
	suite.mockFileStore.On("GetRoleListCount", suite.ctx).Return(0, fileErr)

	result, err := suite.store.GetRoleList(suite.ctx, 10, 0)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), fileErr, err)
}
