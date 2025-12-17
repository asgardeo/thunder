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

package role

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	dbmodel "github.com/asgardeo/thunder/internal/system/database/model"

	"github.com/asgardeo/thunder/tests/mocks/database/modelmock"
	"github.com/asgardeo/thunder/tests/mocks/database/providermock"
)

const testDeploymentID = "test-deployment-id"

// mockResult is a simple mock implementation of sql.Result.
type mockResult struct {
	lastInsertID int64
	rowsAffected int64
}

func (m *mockResult) LastInsertId() (int64, error) {
	return m.lastInsertID, nil
}

func (m *mockResult) RowsAffected() (int64, error) {
	return m.rowsAffected, nil
}

var _ sql.Result = (*mockResult)(nil)

// RoleStoreTestSuite is the test suite for roleStore.
type RoleStoreTestSuite struct {
	suite.Suite
	mockDBProvider *providermock.DBProviderInterfaceMock
	mockDBClient   *providermock.DBClientInterfaceMock
	mockTx         *modelmock.TxInterfaceMock
	store          *roleStore
}

// TestRoleStoreTestSuite runs the test suite.
func TestRoleStoreTestSuite(t *testing.T) {
	suite.Run(t, new(RoleStoreTestSuite))
}

// SetupTest sets up the test suite.
func (suite *RoleStoreTestSuite) SetupTest() {
	suite.mockDBProvider = providermock.NewDBProviderInterfaceMock(suite.T())
	suite.mockDBClient = providermock.NewDBClientInterfaceMock(suite.T())
	suite.mockTx = modelmock.NewTxInterfaceMock(suite.T())
	suite.store = &roleStore{
		dbProvider:   suite.mockDBProvider,
		deploymentID: testDeploymentID,
	}
}

func (suite *RoleStoreTestSuite) TestGetRoleListCount_Success() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetRoleListCount, testDeploymentID).Return([]map[string]interface{}{
		{"total": int64(10)},
	}, nil)

	count, err := suite.store.GetRoleListCount()

	suite.NoError(err)
	suite.Equal(10, count)
}

func (suite *RoleStoreTestSuite) TestGetRoleListCount_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetRoleListCount, testDeploymentID).Return(nil, queryError)

	count, err := suite.store.GetRoleListCount()

	suite.Error(err)
	suite.Equal(0, count)
	suite.Contains(err.Error(), "failed to execute count query")
}

func (suite *RoleStoreTestSuite) TestGetRoleList_Success() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetRoleList, 10, 0, testDeploymentID).Return([]map[string]interface{}{
		{"role_id": "role1", "name": "Admin", "description": "Admin role", "ou_id": "ou1"},
		{"role_id": "role2", "name": "User", "description": "User role", "ou_id": "ou1"},
	}, nil)

	roles, err := suite.store.GetRoleList(10, 0)

	suite.NoError(err)
	suite.Len(roles, 2)
	suite.Equal("role1", roles[0].ID)
	suite.Equal("Admin", roles[0].Name)
}

func (suite *RoleStoreTestSuite) TestGetRoleList_QueryError() {
	queryError := errors.New("query error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetRoleList, 10, 0, testDeploymentID).Return(nil, queryError)

	roles, err := suite.store.GetRoleList(10, 0)

	suite.Error(err)
	suite.Nil(roles)
}

func (suite *RoleStoreTestSuite) TestGetRoleList_InvalidRowData() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetRoleList, 10, 0, testDeploymentID).Return([]map[string]interface{}{
		{"role_id": 123, "name": "Admin", "description": "Admin role", "ou_id": "ou1"}, // Invalid role_id type
	}, nil)

	roles, err := suite.store.GetRoleList(10, 0)

	suite.Error(err)
	suite.Nil(roles)
	suite.Contains(err.Error(), "failed to build role from result row")
}

func (suite *RoleStoreTestSuite) TestCreateRole_Success() {
	roleDetail := RoleCreationDetail{
		Name:               "Test Role",
		Description:        "Test Description",
		OrganizationUnitID: "ou1",
		Permissions:        []string{"perm1", "perm2"},
		Assignments:        []RoleAssignment{{ID: "user1", Type: AssigneeTypeUser}},
	}

	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("BeginTx").Return(suite.mockTx, nil)
	suite.mockTx.On("Exec", queryCreateRole, mock.Anything, "ou1", "Test Role", "Test Description",
		testDeploymentID).Return(&mockResult{}, nil)
	suite.mockTx.On("Exec", queryCreateRolePermission, mock.Anything, "perm1", testDeploymentID).
		Return(&mockResult{}, nil)
	suite.mockTx.On("Exec", queryCreateRolePermission, mock.Anything, "perm2", testDeploymentID).
		Return(&mockResult{}, nil)
	suite.mockTx.On("Exec", queryCreateRoleAssignment, mock.Anything, AssigneeTypeUser, "user1",
		testDeploymentID).Return(&mockResult{}, nil)
	suite.mockTx.On("Commit").Return(nil)

	err := suite.store.CreateRole("role1", roleDetail)

	suite.NoError(err)
}

func (suite *RoleStoreTestSuite) TestCreateRole_ExecError() {
	roleDetail := RoleCreationDetail{
		Name:               "Test Role",
		Description:        "Test Description",
		OrganizationUnitID: "ou1",
		Permissions:        []string{},
		Assignments:        []RoleAssignment{},
	}

	execError := errors.New("insert failed")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("BeginTx").Return(suite.mockTx, nil)
	suite.mockTx.On("Exec", queryCreateRole, mock.Anything, "ou1", "Test Role", "Test Description",
		testDeploymentID).Return(nil, execError)
	suite.mockTx.On("Rollback").Return(nil)

	err := suite.store.CreateRole("role1", roleDetail)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to execute query")
}

func (suite *RoleStoreTestSuite) TestGetRole_Success() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetRoleByID, "role1", testDeploymentID).Return([]map[string]interface{}{
		{"role_id": "role1", "name": "Admin", "description": "Admin role", "ou_id": "ou1"},
	}, nil)
	suite.mockDBClient.On("Query", queryGetRolePermissions, "role1", testDeploymentID).Return([]map[string]interface{}{
		{"permission": "perm1"},
		{"permission": "perm2"},
	}, nil)

	role, err := suite.store.GetRole("role1")

	suite.NoError(err)
	suite.Equal("role1", role.ID)
	suite.Equal("Admin", role.Name)
	suite.Len(role.Permissions, 2)
}

func (suite *RoleStoreTestSuite) TestGetRole_NotFound() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetRoleByID, "nonexistent", testDeploymentID).
		Return([]map[string]interface{}{}, nil)

	role, err := suite.store.GetRole("nonexistent")

	suite.Error(err)
	suite.Equal(ErrRoleNotFound, err)
	suite.Empty(role.ID)
}

func (suite *RoleStoreTestSuite) TestGetRole_MultipleResults() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetRoleByID, "role1", testDeploymentID).Return([]map[string]interface{}{
		{"role_id": "role1", "name": "Admin", "description": "Admin role", "ou_id": "ou1"},
		{"role_id": "role1", "name": "Admin", "description": "Admin role", "ou_id": "ou1"},
	}, nil)

	role, err := suite.store.GetRole("role1")

	suite.Error(err)
	suite.Contains(err.Error(), "unexpected number of results")
	suite.Empty(role.ID)
}

func (suite *RoleStoreTestSuite) TestIsRoleExist_Exists() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckRoleExists, "role1", testDeploymentID).Return([]map[string]interface{}{
		{"count": int64(1)},
	}, nil)

	exists, err := suite.store.IsRoleExist("role1")

	suite.NoError(err)
	suite.True(exists)
}

func (suite *RoleStoreTestSuite) TestIsRoleExist_DoesNotExist() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckRoleExists, "nonexistent", testDeploymentID).
		Return([]map[string]interface{}{
			{"count": int64(0)},
		}, nil)

	exists, err := suite.store.IsRoleExist("nonexistent")

	suite.NoError(err)
	suite.False(exists)
}

func (suite *RoleStoreTestSuite) TestGetRoleAssignments_Success() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetRoleAssignments, "role1", 10, 0, testDeploymentID).
		Return([]map[string]interface{}{
			{"assignee_id": "user1", "assignee_type": "user"},
			{"assignee_id": "group1", "assignee_type": "group"},
		}, nil)

	assignments, err := suite.store.GetRoleAssignments("role1", 10, 0)

	suite.NoError(err)
	suite.Len(assignments, 2)
	suite.Equal("user1", assignments[0].ID)
	suite.Equal(AssigneeTypeUser, assignments[0].Type)
}

func (suite *RoleStoreTestSuite) TestGetRoleAssignmentsCount_Success() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryGetRoleAssignmentsCount, "role1", testDeploymentID).
		Return([]map[string]interface{}{
			{"total": int64(5)},
		}, nil)

	count, err := suite.store.GetRoleAssignmentsCount("role1")

	suite.NoError(err)
	suite.Equal(5, count)
}

func (suite *RoleStoreTestSuite) TestUpdateRole_Success() {
	roleDetail := RoleUpdateDetail{
		Name:               "Updated Role",
		Description:        "Updated Description",
		OrganizationUnitID: "ou1",
		Permissions:        []string{"perm1"},
	}

	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("BeginTx").Return(suite.mockTx, nil)
	suite.mockTx.On("Exec", queryUpdateRole, "ou1", "Updated Role", "Updated Description", "role1",
		testDeploymentID).Return(&mockResult{rowsAffected: 1}, nil)
	suite.mockTx.On("Exec", queryDeleteRolePermissions, "role1", testDeploymentID).Return(&mockResult{}, nil)
	suite.mockTx.On("Exec", queryCreateRolePermission, "role1", "perm1", testDeploymentID).
		Return(&mockResult{}, nil)
	suite.mockTx.On("Commit").Return(nil)

	err := suite.store.UpdateRole("role1", roleDetail)

	suite.NoError(err)
}

func (suite *RoleStoreTestSuite) TestUpdateRole_NotFound() {
	roleDetail := RoleUpdateDetail{
		Name:               "Updated Role",
		Description:        "Updated Description",
		OrganizationUnitID: "ou1",
		Permissions:        []string{},
	}

	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("BeginTx").Return(suite.mockTx, nil)
	suite.mockTx.On(
		"Exec", queryUpdateRole, "ou1", "Updated Role", "Updated Description", "nonexistent", testDeploymentID,
	).Return(&mockResult{rowsAffected: 0}, nil)
	suite.mockTx.On("Rollback").Return(nil)

	err := suite.store.UpdateRole("nonexistent", roleDetail)

	suite.Error(err)
	suite.Equal(ErrRoleNotFound, err)
}

func (suite *RoleStoreTestSuite) TestDeleteRole_Success() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryDeleteRole, "role1", testDeploymentID).Return(int64(1), nil)

	err := suite.store.DeleteRole("role1")

	suite.NoError(err)
}

func (suite *RoleStoreTestSuite) TestDeleteRole_ExecuteError() {
	execError := errors.New("delete failed")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", queryDeleteRole, "role1", testDeploymentID).Return(int64(0), execError)

	err := suite.store.DeleteRole("role1")

	suite.Error(err)
}

func (suite *RoleStoreTestSuite) TestAddAssignments_Success() {
	assignments := []RoleAssignment{
		{ID: "user1", Type: AssigneeTypeUser},
	}

	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("BeginTx").Return(suite.mockTx, nil)
	suite.mockTx.On("Exec", queryCreateRoleAssignment, "role1", AssigneeTypeUser, "user1", testDeploymentID).
		Return(&mockResult{}, nil)
	suite.mockTx.On("Commit").Return(nil)

	err := suite.store.AddAssignments("role1", assignments)

	suite.NoError(err)
}

func (suite *RoleStoreTestSuite) TestRemoveAssignments_Success() {
	assignments := []RoleAssignment{
		{ID: "user1", Type: AssigneeTypeUser},
	}

	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("BeginTx").Return(suite.mockTx, nil)
	suite.mockTx.On("Exec", queryDeleteRoleAssignmentsByIDs, "role1", AssigneeTypeUser, "user1",
		testDeploymentID).Return(&mockResult{}, nil)
	suite.mockTx.On("Commit").Return(nil)

	err := suite.store.RemoveAssignments("role1", assignments)

	suite.NoError(err)
}

func (suite *RoleStoreTestSuite) TestCheckRoleNameExists_Exists() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckRoleNameExists, "ou1", "Admin", testDeploymentID).
		Return([]map[string]interface{}{
			{"count": int64(1)},
		}, nil)

	exists, err := suite.store.CheckRoleNameExists("ou1", "Admin")

	suite.NoError(err)
	suite.True(exists)
}

func (suite *RoleStoreTestSuite) TestCheckRoleNameExistsExcludingID_DoesNotExist() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", queryCheckRoleNameExistsExcludingID, "ou1", "Admin", "role1", testDeploymentID).
		Return([]map[string]interface{}{
			{"count": int64(0)},
		}, nil)

	exists, err := suite.store.CheckRoleNameExistsExcludingID("ou1", "Admin", "role1")

	suite.NoError(err)
	suite.False(exists)
}

func (suite *RoleStoreTestSuite) TestGetAuthorizedPermissions_Success() {
	userID := "user1"
	groupIDs := []string{"group1"}
	requestedPermissions := []string{"perm1", "perm2"}

	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", mock.Anything, testDeploymentID, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything).
		Return(
			[]map[string]interface{}{
				{"permission": "perm1"},
			}, nil)

	permissions, err := suite.store.GetAuthorizedPermissions(userID, groupIDs, requestedPermissions)

	suite.NoError(err)
	suite.Len(permissions, 1)
	suite.Equal("perm1", permissions[0])
}

func (suite *RoleStoreTestSuite) TestGetAuthorizedPermissions_NilGroupsHandled() {
	userID := "user1"
	requestedPermissions := []string{"perm1"}

	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", mock.Anything, testDeploymentID, mock.Anything, mock.Anything).
		Return([]map[string]interface{}{{"permission": "perm1"}}, nil)

	permissions, err := suite.store.GetAuthorizedPermissions(userID, nil, requestedPermissions)

	suite.NoError(err)
	suite.Len(permissions, 1)
}

// Test buildRoleBasicInfoFromResultRow

func (suite *RoleStoreTestSuite) TestBuildRoleBasicInfoFromResultRow_Success() {
	row := map[string]interface{}{
		"role_id":     "role1",
		"name":        "Admin",
		"description": "Admin role",
		"ou_id":       "ou1",
	}

	role, err := buildRoleBasicInfoFromResultRow(row)

	suite.NoError(err)
	suite.Equal("role1", role.ID)
	suite.Equal("Admin", role.Name)
	suite.Equal("Admin role", role.Description)
	suite.Equal("ou1", role.OrganizationUnitID)
}

func (suite *RoleStoreTestSuite) TestBuildRoleBasicInfoFromResultRow_InvalidData() {
	row := map[string]interface{}{
		"role_id":     123, // Invalid type
		"name":        "Admin",
		"description": "Admin role",
		"ou_id":       "ou1",
	}

	role, err := buildRoleBasicInfoFromResultRow(row)

	suite.Error(err)
	suite.Empty(role.ID)
}

// Test Helper Functions

func (suite *RoleStoreTestSuite) TestGetIdentityDBClient_Success() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)

	client, err := suite.store.getIdentityDBClient()

	suite.NoError(err)
	suite.NotNil(client)
	suite.Equal(suite.mockDBClient, client)
}

func (suite *RoleStoreTestSuite) TestGetIdentityDBClient_Error() {
	dbError := errors.New("database connection error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(nil, dbError)

	client, err := suite.store.getIdentityDBClient()

	suite.Error(err)
	suite.Nil(client)
	suite.Contains(err.Error(), "failed to get database client")
}

func (suite *RoleStoreTestSuite) TestParseCountResult_Success() {
	results := []map[string]interface{}{
		{"total": int64(42)},
	}

	count, err := parseCountResult(results)

	suite.NoError(err)
	suite.Equal(42, count)
}

func (suite *RoleStoreTestSuite) TestParseCountResult_EmptyResults() {
	results := []map[string]interface{}{}

	count, err := parseCountResult(results)

	suite.NoError(err)
	suite.Equal(0, count)
}

func (suite *RoleStoreTestSuite) TestParseCountResult_TypeAssertionError() {
	results := []map[string]interface{}{
		{"total": "not_a_number"},
	}

	count, err := parseCountResult(results)

	suite.Error(err)
	suite.Equal(0, count)
	suite.Contains(err.Error(), "failed to parse total")
}

func (suite *RoleStoreTestSuite) TestParseBoolFromCount_True() {
	results := []map[string]interface{}{
		{"count": int64(5)},
	}

	exists, err := parseBoolFromCount(results)

	suite.NoError(err)
	suite.True(exists)
}

func (suite *RoleStoreTestSuite) TestParseBoolFromCount_False() {
	results := []map[string]interface{}{
		{"count": int64(0)},
	}

	exists, err := parseBoolFromCount(results)

	suite.NoError(err)
	suite.False(exists)
}

func (suite *RoleStoreTestSuite) TestParseBoolFromCount_EmptyResults() {
	results := []map[string]interface{}{}

	exists, err := parseBoolFromCount(results)

	suite.NoError(err)
	suite.False(exists)
}

func (suite *RoleStoreTestSuite) TestParseBoolFromCount_TypeError() {
	results := []map[string]interface{}{
		{"count": "invalid"},
	}

	exists, err := parseBoolFromCount(results)

	suite.Error(err)
	suite.False(exists)
}

func (suite *RoleStoreTestSuite) TestParseStringField_Success() {
	row := map[string]interface{}{
		"name": "test_value",
	}

	value, err := parseStringField(row, "name")

	suite.NoError(err)
	suite.Equal("test_value", value)
}

func (suite *RoleStoreTestSuite) TestParseStringField_TypeError() {
	row := map[string]interface{}{
		"name": 123,
	}

	value, err := parseStringField(row, "name")

	suite.Error(err)
	suite.Empty(value)
	suite.Contains(err.Error(), "failed to parse name")
}

func (suite *RoleStoreTestSuite) TestParseStringFields_Success() {
	row := map[string]interface{}{
		"role_id":     "role1",
		"name":        "Admin",
		"description": "Admin role",
		"ou_id":       "ou1",
	}

	values, err := parseStringFields(row, "role_id", "name", "description", "ou_id")

	suite.NoError(err)
	suite.Len(values, 4)
	suite.Equal("role1", values[0])
	suite.Equal("Admin", values[1])
	suite.Equal("Admin role", values[2])
	suite.Equal("ou1", values[3])
}

func (suite *RoleStoreTestSuite) TestParseStringFields_PartialError() {
	row := map[string]interface{}{
		"role_id": "role1",
		"name":    123, // Invalid type
	}

	values, err := parseStringFields(row, "role_id", "name")

	suite.Error(err)
	suite.Nil(values)
	suite.Contains(err.Error(), "failed to parse name")
}

func (suite *RoleStoreTestSuite) TestExecuteInTransaction_Success() {
	suite.mockDBClient.On("BeginTx").Return(suite.mockTx, nil)
	suite.mockTx.On("Commit").Return(nil)

	operationCalled := false
	err := suite.store.executeInTransaction(suite.mockDBClient, func(tx dbmodel.TxInterface) error {
		operationCalled = true
		return nil
	})

	suite.NoError(err)
	suite.True(operationCalled)
}

func (suite *RoleStoreTestSuite) TestExecuteInTransaction_BeginError() {
	beginError := errors.New("begin transaction failed")
	suite.mockDBClient.On("BeginTx").Return(nil, beginError)

	err := suite.store.executeInTransaction(suite.mockDBClient, func(tx dbmodel.TxInterface) error {
		suite.Fail("Operation should not be called")
		return nil
	})

	suite.Error(err)
	suite.Contains(err.Error(), "failed to begin transaction")
}

func (suite *RoleStoreTestSuite) TestExecuteInTransaction_OperationError() {
	operationError := errors.New("operation failed")
	suite.mockDBClient.On("BeginTx").Return(suite.mockTx, nil)
	suite.mockTx.On("Rollback").Return(nil)

	err := suite.store.executeInTransaction(suite.mockDBClient, func(tx dbmodel.TxInterface) error {
		return operationError
	})

	suite.Error(err)
	suite.Equal(operationError, err)
}

func (suite *RoleStoreTestSuite) TestExecuteInTransaction_RollbackError() {
	operationError := errors.New("operation failed")
	rollbackError := errors.New("rollback failed")
	suite.mockDBClient.On("BeginTx").Return(suite.mockTx, nil)
	suite.mockTx.On("Rollback").Return(rollbackError)

	err := suite.store.executeInTransaction(suite.mockDBClient, func(tx dbmodel.TxInterface) error {
		return operationError
	})

	suite.Error(err)
	suite.Contains(err.Error(), "operation failed")
	suite.Contains(err.Error(), "rollback")
}

func (suite *RoleStoreTestSuite) TestExecuteInTransaction_CommitError() {
	commitError := errors.New("commit failed")
	suite.mockDBClient.On("BeginTx").Return(suite.mockTx, nil)
	suite.mockTx.On("Commit").Return(commitError)

	err := suite.store.executeInTransaction(suite.mockDBClient, func(tx dbmodel.TxInterface) error {
		return nil
	})

	suite.Error(err)
	suite.Contains(err.Error(), "failed to commit transaction")
}
