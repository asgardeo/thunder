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
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

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

func (suite *RoleStoreTestSuite) TestGetRoleListCount() {
	testCases := []struct {
		name          string
		setupMocks    func()
		expectedCount int
		shouldErr     bool
		checkError    func(error) bool
	}{
		{
			name: "Success",
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryGetRoleListCount, testDeploymentID).
					Return([]map[string]interface{}{
						{"total": int64(10)},
					}, nil)
			},
			expectedCount: 10,
			shouldErr:     false,
		},
		{
			name: "QueryError",
			setupMocks: func() {
				queryError := errors.New("query error")
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryGetRoleListCount, testDeploymentID).
					Return(nil, queryError)
			},
			expectedCount: 0,
			shouldErr:     true,
			checkError: func(err error) bool {
				suite.Contains(err.Error(), "failed to execute count query")
				return true
			},
		},
		{
			name: "DBClientError",
			setupMocks: func() {
				dbError := errors.New("db client error")
				suite.mockDBProvider.On("GetConfigDBClient").Return(nil, dbError)
			},
			expectedCount: 0,
			shouldErr:     true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Create fresh mocks for each test case
			suite.mockDBProvider = providermock.NewDBProviderInterfaceMock(suite.T())
			suite.mockDBClient = providermock.NewDBClientInterfaceMock(suite.T())
			suite.store = &roleStore{
				dbProvider:   suite.mockDBProvider,
				deploymentID: testDeploymentID,
			}

			tc.setupMocks()

			count, err := suite.store.GetRoleListCount(context.Background())

			if tc.shouldErr {
				suite.Error(err)
				suite.Equal(tc.expectedCount, count)
				if tc.checkError != nil {
					tc.checkError(err)
				}
			} else {
				suite.NoError(err)
				suite.Equal(tc.expectedCount, count)
			}
		})
	}
}

func (suite *RoleStoreTestSuite) TestGetRoleList() {
	testCases := []struct {
		name          string
		limit         int
		offset        int
		setupMocks    func()
		expectedRoles []Role
		shouldErr     bool
		checkError    func(error) bool
	}{
		{
			name:   "Success",
			limit:  10,
			offset: 0,
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryGetRoleList, 10, 0, testDeploymentID).
					Return([]map[string]interface{}{
						{"role_id": "role1", "name": "Admin", "description": "Admin role", "ou_id": "ou1"},
						{"role_id": "role2", "name": "User", "description": "User role", "ou_id": "ou1"},
					}, nil)
			},
			expectedRoles: []Role{
				{ID: "role1", Name: "Admin"},
				{ID: "role2", Name: "User"},
			},
			shouldErr: false,
		},
		{
			name:   "QueryError",
			limit:  10,
			offset: 0,
			setupMocks: func() {
				queryError := errors.New("query error")
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryGetRoleList, 10, 0, testDeploymentID).
					Return(nil, queryError)
			},
			shouldErr: true,
		},
		{
			name:   "InvalidRowData",
			limit:  10,
			offset: 0,
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryGetRoleList, 10, 0, testDeploymentID).
					Return([]map[string]interface{}{
						{
							"role_id": 123, "name": "Admin",
							"description": "Admin role", "ou_id": "ou1",
						},
					}, nil)
			},
			shouldErr: true,
			checkError: func(err error) bool {
				suite.Contains(err.Error(), "failed to build role from result row")
				return true
			},
		},
		{
			name:   "DBClientError",
			limit:  10,
			offset: 0,
			setupMocks: func() {
				dbError := errors.New("db client error")
				suite.mockDBProvider.On("GetConfigDBClient").Return(nil, dbError)
			},
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Create fresh mocks for each test case
			suite.mockDBProvider = providermock.NewDBProviderInterfaceMock(suite.T())
			suite.mockDBClient = providermock.NewDBClientInterfaceMock(suite.T())
			suite.store = &roleStore{
				dbProvider:   suite.mockDBProvider,
				deploymentID: testDeploymentID,
			}

			tc.setupMocks()

			roles, err := suite.store.GetRoleList(context.Background(), tc.limit, tc.offset)

			if tc.shouldErr {
				suite.Error(err)
				suite.Nil(roles)
				if tc.checkError != nil {
					tc.checkError(err)
				}
			} else {
				suite.NoError(err)
				suite.Len(roles, len(tc.expectedRoles))
				if len(tc.expectedRoles) > 0 {
					suite.Equal(tc.expectedRoles[0].ID, roles[0].ID)
					suite.Equal(tc.expectedRoles[0].Name, roles[0].Name)
				}
			}
		})
	}
}

func (suite *RoleStoreTestSuite) TestCreateRole() {
	testCases := []struct {
		name       string
		roleID     string
		roleDetail RoleCreationDetail
		setupMocks func()
		shouldErr  bool
		checkError func(error) bool
	}{
		{
			name:   "Success",
			roleID: "role1",
			roleDetail: RoleCreationDetail{
				Name:               "Test Role",
				Description:        "Test Description",
				OrganizationUnitID: "ou1",
				Permissions: []ResourcePermissions{
					{ResourceServerID: "rs1", Permissions: []string{"perm1", "perm2"}},
				},
				Assignments: []RoleAssignment{{ID: "user1", Type: AssigneeTypeUser}},
			},
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On(
					"ExecuteWithReturningContext", mock.Anything, queryCreateRole, "role1", "ou1", "Test Role",
					"Test Description", testDeploymentID).Return([]map[string]interface{}{{"id": int64(1)}}, nil)
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryCreateRolePermission, 1, "rs1",
					"perm1", testDeploymentID).Return(int64(1), nil)
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryCreateRolePermission, 1, "rs1",
					"perm2", testDeploymentID).Return(int64(1), nil)
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryCreateRoleAssignment, 1,
					AssigneeTypeUser, "user1", testDeploymentID).Return(int64(1), nil)
			},
			shouldErr: false,
		},
		{
			name:   "MultipleResourceServers",
			roleID: "role1",
			roleDetail: RoleCreationDetail{
				Name:               "Test Role",
				Description:        "Test Description",
				OrganizationUnitID: "ou1",
				Permissions: []ResourcePermissions{
					{ResourceServerID: "rs1", Permissions: []string{"perm1"}},
					{ResourceServerID: "rs2", Permissions: []string{"perm2", "perm3"}},
				},
				Assignments: []RoleAssignment{},
			},
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On(
					"ExecuteWithReturningContext", mock.Anything, queryCreateRole, "role1", "ou1", "Test Role",
					"Test Description", testDeploymentID).Return([]map[string]interface{}{{"id": int64(1)}}, nil)
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryCreateRolePermission, 1, "rs1",
					"perm1", testDeploymentID).Return(int64(1), nil)
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryCreateRolePermission, 1, "rs2",
					"perm2", testDeploymentID).Return(int64(1), nil)
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryCreateRolePermission, 1, "rs2",
					"perm3", testDeploymentID).Return(int64(1), nil)
			},
			shouldErr: false,
		},
		{
			name:   "ExecError",
			roleID: "role1",
			roleDetail: RoleCreationDetail{
				Name:               "Test Role",
				Description:        "Test Description",
				OrganizationUnitID: "ou1",
				Permissions:        []ResourcePermissions{},
				Assignments:        []RoleAssignment{},
			},
			setupMocks: func() {
				execError := errors.New("insert failed")
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On(
					"ExecuteWithReturningContext", mock.Anything, queryCreateRole, "role1", "ou1", "Test Role",
					"Test Description", testDeploymentID).Return(nil, execError)
			},
			shouldErr: true,
			checkError: func(err error) bool {
				suite.Contains(err.Error(), "failed to execute query")
				return true
			},
		},
		{
			name:   "PermissionError",
			roleID: "role1",
			roleDetail: RoleCreationDetail{
				Name:               "Test Role",
				Description:        "Test Description",
				OrganizationUnitID: "ou1",
				Permissions: []ResourcePermissions{
					{ResourceServerID: "rs1", Permissions: []string{"perm1"}},
				},
				Assignments: []RoleAssignment{},
			},
			setupMocks: func() {
				permError := errors.New("permission insert failed")
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On(
					"ExecuteWithReturningContext", mock.Anything, queryCreateRole, "role1", "ou1", "Test Role",
					"Test Description", testDeploymentID).Return([]map[string]interface{}{{"id": int64(1)}}, nil)
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryCreateRolePermission, 1, "rs1",
					"perm1", testDeploymentID).Return(int64(0), permError)
			},
			shouldErr: true,
			checkError: func(err error) bool {
				suite.Contains(err.Error(), "failed to add permission to role")
				return true
			},
		},
		{
			name:   "AssignmentError",
			roleID: "role1",
			roleDetail: RoleCreationDetail{
				Name:               "Test Role",
				Description:        "Test Description",
				OrganizationUnitID: "ou1",
				Permissions:        []ResourcePermissions{},
				Assignments:        []RoleAssignment{{ID: "user1", Type: AssigneeTypeUser}},
			},
			setupMocks: func() {
				assignError := errors.New("assignment insert failed")
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On(
					"ExecuteWithReturningContext", mock.Anything, queryCreateRole, "role1", "ou1", "Test Role",
					"Test Description", testDeploymentID).Return([]map[string]interface{}{{"id": int64(1)}}, nil)
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryCreateRoleAssignment, 1,
					AssigneeTypeUser, "user1", testDeploymentID).
					Return(int64(0), assignError)
			},
			shouldErr: true,
			checkError: func(err error) bool {
				suite.Contains(err.Error(), "failed to add assignment to role")
				return true
			},
		},
		{
			name:       "DBClientError",
			roleID:     "role1",
			roleDetail: RoleCreationDetail{},
			setupMocks: func() {
				dbError := errors.New("db client error")
				suite.mockDBProvider.On("GetConfigDBClient").Return(nil, dbError)
			},
			shouldErr: true,
		},
		{
			name:   "NoRowsReturned",
			roleID: "role1",
			roleDetail: RoleCreationDetail{
				Name:               "Test Role",
				Description:        "Test Description",
				OrganizationUnitID: "ou1",
			},
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On(
					"ExecuteWithReturningContext", mock.Anything, queryCreateRole, "role1", "ou1", "Test Role",
					"Test Description", testDeploymentID).Return([]map[string]interface{}{}, nil)
			},
			shouldErr: true,
			checkError: func(err error) bool {
				suite.Contains(err.Error(), "failed to create role: no rows returned")
				return true
			},
		},
		{
			name:   "ResolveIDError",
			roleID: "role1",
			roleDetail: RoleCreationDetail{
				Name:               "Test Role",
				Description:        "Test Description",
				OrganizationUnitID: "ou1",
			},
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On(
					"ExecuteWithReturningContext", mock.Anything, queryCreateRole, "role1", "ou1", "Test Role",
					"Test Description", testDeploymentID).Return([]map[string]interface{}{{"id": "invalid"}}, nil)
			},
			shouldErr: true,
			checkError: func(err error) bool {
				suite.Contains(err.Error(), "failed to resolve role internal ID")
				return true
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Create fresh mocks for each test case
			suite.mockDBProvider = providermock.NewDBProviderInterfaceMock(suite.T())
			suite.mockDBClient = providermock.NewDBClientInterfaceMock(suite.T())
			suite.mockTx = modelmock.NewTxInterfaceMock(suite.T())
			suite.store = &roleStore{
				dbProvider:   suite.mockDBProvider,
				deploymentID: testDeploymentID,
			}

			tc.setupMocks()

			err := suite.store.CreateRole(context.Background(), tc.roleID, tc.roleDetail)

			if tc.shouldErr {
				suite.Error(err)
				if tc.checkError != nil {
					tc.checkError(err)
				}
			} else {
				suite.NoError(err)
			}
		})
	}
}

func (suite *RoleStoreTestSuite) TestGetRole() {
	testCases := []struct {
		name               string
		roleID             string
		setupMocks         func()
		expectedRole       *RoleWithPermissions
		shouldErr          bool
		checkError         func(error) bool
		checkPermissions   bool
		expectedPermCount  int
		expectedPermLength []int
	}{
		{
			name:   "Success",
			roleID: "role1",
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryGetRoleByID, "role1", testDeploymentID).
					Return([]map[string]interface{}{
						{
							"id": int64(1), "role_id": "role1", "name": "Admin",
							"description": "Admin role", "ou_id": "ou1",
						},
					}, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryGetRolePermissions, 1,
					testDeploymentID).
					Return([]map[string]interface{}{
						{"resource_server_id": "rs1", "permission": "perm1"},
						{"resource_server_id": "rs1", "permission": "perm2"},
					}, nil)
			},
			expectedRole: &RoleWithPermissions{
				ID:                 "role1",
				Name:               "Admin",
				Description:        "Admin role",
				OrganizationUnitID: "ou1",
			},
			shouldErr:          false,
			checkPermissions:   true,
			expectedPermCount:  1,
			expectedPermLength: []int{2},
		},
		{
			name:   "MultipleResourceServers",
			roleID: "role1",
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryGetRoleByID, "role1", testDeploymentID).
					Return([]map[string]interface{}{{
						"id":          int64(1),
						"role_id":     "role1",
						"name":        "Admin",
						"description": "Admin role",
						"ou_id":       "ou1",
					}}, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryGetRolePermissions, 1,
					testDeploymentID).
					Return([]map[string]interface{}{
						{"resource_server_id": "rs1", "permission": "read:users"},
						{"resource_server_id": "rs1", "permission": "write:users"},
						{"resource_server_id": "rs2", "permission": "read:posts"},
						{"resource_server_id": "rs2", "permission": "write:posts"},
					}, nil)
			},
			shouldErr:          false,
			checkPermissions:   true,
			expectedPermCount:  2,
			expectedPermLength: []int{2, 2},
		},
		{
			name:   "EmptyPermissions",
			roleID: "role1",
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryGetRoleByID, "role1",
					testDeploymentID).Return([]map[string]interface{}{
					{"id": int64(1), "role_id": "role1", "name": "Admin", "description": "Admin role", "ou_id": "ou1"},
				}, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryGetRolePermissions, 1,
					testDeploymentID).
					Return([]map[string]interface{}{}, nil)
			},
			shouldErr:         false,
			checkPermissions:  true,
			expectedPermCount: 0,
		},
		{
			name:   "NotFound",
			roleID: "nonexistent",
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryGetRoleByID, "nonexistent",
					testDeploymentID).
					Return([]map[string]interface{}{}, nil)
			},
			shouldErr: true,
			checkError: func(err error) bool {
				suite.Equal(ErrRoleNotFound, err)
				return true
			},
		},
		{
			name:   "MultipleResults",
			roleID: "role1",
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryGetRoleByID, "role1",
					testDeploymentID).Return([]map[string]interface{}{
					{"id": int64(1), "role_id": "role1", "name": "Admin", "description": "Admin role", "ou_id": "ou1"},
					{"id": int64(1), "role_id": "role1", "name": "Admin", "description": "Admin role", "ou_id": "ou1"},
				}, nil)
			},
			shouldErr: true,
			checkError: func(err error) bool {
				suite.Contains(err.Error(), "unexpected number of results")
				return true
			},
		},
		{
			name:   "InvalidPermissionType",
			roleID: "role1",
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryGetRoleByID, "role1",
					testDeploymentID).Return([]map[string]interface{}{
					{"id": int64(1), "role_id": "role1", "name": "Admin", "description": "Admin role", "ou_id": "ou1"},
				}, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryGetRolePermissions, 1,
					testDeploymentID).Return([]map[string]interface{}{
					{"resource_server_id": "rs1", "permission": 123}, // Invalid type
				}, nil)
			},
			shouldErr: true,
			checkError: func(err error) bool {
				suite.Contains(err.Error(), "failed to parse permission as string")
				return true
			},
		},
		{
			name:   "PermissionsQueryError",
			roleID: "role1",
			setupMocks: func() {
				dbError := errors.New("database query error")
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryGetRoleByID, "role1",
					testDeploymentID).Return([]map[string]interface{}{
					{"id": int64(1), "role_id": "role1", "name": "Admin", "description": "Admin role", "ou_id": "ou1"},
				}, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryGetRolePermissions, 1,
					testDeploymentID).
					Return(nil, dbError)
			},
			shouldErr: true,
			checkError: func(err error) bool {
				suite.Contains(err.Error(), "failed to get role permissions")
				return true
			},
		},
		{
			name:   "DBClientError",
			roleID: "role1",
			setupMocks: func() {
				dbError := errors.New("db client error")
				suite.mockDBProvider.On("GetConfigDBClient").Return(nil, dbError)
			},
			shouldErr: true,
		},
		{
			name:   "QueryError",
			roleID: "role1",
			setupMocks: func() {
				queryError := errors.New("query failed")
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryGetRoleByID, "role1", testDeploymentID).
					Return(nil, queryError)
			},
			shouldErr: true,
		},
		{
			name:   "ResolveIDError",
			roleID: "role1",
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryGetRoleByID, "role1", testDeploymentID).
					Return([]map[string]interface{}{
						{
							"id": "invalid", "role_id": "role1", "name": "Admin",
							"description": "Admin role", "ou_id": "ou1",
						},
					}, nil)
			},
			shouldErr: true,
			checkError: func(err error) bool {
				suite.Contains(err.Error(), "unexpected internal ID type")
				return true
			},
		},
		{
			name:   "InvalidRowData",
			roleID: "role1",
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryGetRoleByID, "role1",
					testDeploymentID).Return([]map[string]interface{}{
					{
						"id": 123, "role_id": 123, "name": "Admin",
						"description": "Admin role", "ou_id": "ou1", // Invalid type
					},
				}, nil)
			},
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Create fresh mocks for each test case
			suite.mockDBProvider = providermock.NewDBProviderInterfaceMock(suite.T())
			suite.mockDBClient = providermock.NewDBClientInterfaceMock(suite.T())
			suite.store = &roleStore{
				dbProvider:   suite.mockDBProvider,
				deploymentID: testDeploymentID,
			}

			tc.setupMocks()

			role, err := suite.store.GetRole(context.Background(), tc.roleID)

			if tc.shouldErr {
				suite.Error(err)
				suite.Empty(role.ID)
				if tc.checkError != nil {
					tc.checkError(err)
				}
			} else {
				suite.NoError(err)
				if tc.expectedRole != nil {
					suite.Equal(tc.expectedRole.ID, role.ID)
					suite.Equal(tc.expectedRole.Name, role.Name)
				}
				if tc.checkPermissions {
					suite.Len(role.Permissions, tc.expectedPermCount)
					for i, expectedLen := range tc.expectedPermLength {
						suite.Len(role.Permissions[i].Permissions, expectedLen)
					}
				}
			}
		})
	}
}

func (suite *RoleStoreTestSuite) TestIsRoleExist() {
	testCases := []struct {
		name       string
		roleID     string
		setupMocks func()
		expectedID int
		shouldErr  bool
	}{
		{
			name:   "Exists",
			roleID: "role1",
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryCheckRoleExists, "role1", testDeploymentID).
					Return([]map[string]interface{}{
						{"id": int64(1)},
					}, nil)
			},
			expectedID: 1,
			shouldErr:  false,
		},
		{
			name:   "DoesNotExist",
			roleID: "nonexistent",
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryCheckRoleExists, "nonexistent",
					testDeploymentID).
					Return([]map[string]interface{}{}, nil)
			},
			expectedID: 0,
			shouldErr:  false,
		},
		{
			name:   "QueryError",
			roleID: "role1",
			setupMocks: func() {
				queryError := errors.New("query failed")
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryCheckRoleExists, "role1", testDeploymentID).
					Return(nil, queryError)
			},
			expectedID: 0,
			shouldErr:  true,
		},
		{
			name:   "DBClientError",
			roleID: "role1",
			setupMocks: func() {
				dbError := errors.New("db client error")
				suite.mockDBProvider.On("GetConfigDBClient").Return(nil, dbError)
			},
			expectedID: 0,
			shouldErr:  true,
		},
		{
			name:   "MultipleRolesFound",
			roleID: "role1",
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryCheckRoleExists, "role1", testDeploymentID).
					Return([]map[string]interface{}{
						{"id": int64(1)},
						{"id": int64(2)},
					}, nil)
			},
			expectedID: 0,
			shouldErr:  true,
		},
		{
			name:   "ResolveIDError",
			roleID: "role1",
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryCheckRoleExists, "role1", testDeploymentID).
					Return([]map[string]interface{}{
						{"id": "invalid"},
					}, nil)
			},
			expectedID: 0,
			shouldErr:  true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.mockDBProvider = providermock.NewDBProviderInterfaceMock(suite.T())
			suite.mockDBClient = providermock.NewDBClientInterfaceMock(suite.T())
			suite.store = &roleStore{
				dbProvider:   suite.mockDBProvider,
				deploymentID: testDeploymentID,
			}

			tc.setupMocks()

			exists, internalID, err := suite.store.IsRoleExist(context.Background(), tc.roleID)

			if tc.shouldErr {
				suite.Error(err)
			} else {
				suite.NoError(err)
				if tc.expectedID > 0 {
					suite.True(exists)
					suite.Equal(tc.expectedID, internalID)
				} else {
					suite.False(exists)
					suite.Equal(0, internalID)
				}
			}
		})
	}
}

func (suite *RoleStoreTestSuite) TestGetRoleAssignments_Success() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("QueryContext", mock.Anything, queryGetRoleAssignments, 1, 10, 0, testDeploymentID).
		Return([]map[string]interface{}{
			{"assignee_id": "user1", "assignee_type": "user"},
			{"assignee_id": "group1", "assignee_type": "group"},
		}, nil)

	assignments, err := suite.store.GetRoleAssignments(context.Background(), 1, 10, 0)

	suite.NoError(err)
	suite.Len(assignments, 2)
	suite.Equal("user1", assignments[0].ID)
	suite.Equal(AssigneeTypeUser, assignments[0].Type)
}

func (suite *RoleStoreTestSuite) TestGetRoleAssignmentsCount_Success() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("QueryContext", mock.Anything, queryGetRoleAssignmentsCount, 1, testDeploymentID).
		Return([]map[string]interface{}{
			{"total": int64(5)},
		}, nil)

	count, err := suite.store.GetRoleAssignmentsCount(context.Background(), 1)

	suite.NoError(err)
	suite.Equal(5, count)
}

func (suite *RoleStoreTestSuite) TestDeleteRole() {
	testCases := []struct {
		name       string
		roleID     int
		setupMocks func()
		shouldErr  bool
	}{
		{
			name:   "Success",
			roleID: 1,
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryDeleteRole, 1, testDeploymentID).
					Return(int64(1), nil)
			},
			shouldErr: false,
		},
		{
			name:   "NotFound",
			roleID: 1,
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryDeleteRole, 1, testDeploymentID).
					Return(int64(0), nil)
			},
			shouldErr: false,
		},
		{
			name:   "ExecuteError",
			roleID: 1,
			setupMocks: func() {
				execError := errors.New("delete failed")
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryDeleteRole, 1, testDeploymentID).
					Return(int64(0), execError)
			},
			shouldErr: true,
		},
		{
			name:   "DBClientError",
			roleID: 1,
			setupMocks: func() {
				dbError := errors.New("db client error")
				suite.mockDBProvider.On("GetConfigDBClient").Return(nil, dbError)
			},
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.mockDBProvider = providermock.NewDBProviderInterfaceMock(suite.T())
			suite.mockDBClient = providermock.NewDBClientInterfaceMock(suite.T())
			suite.store = &roleStore{
				dbProvider:   suite.mockDBProvider,
				deploymentID: testDeploymentID,
			}

			tc.setupMocks()

			err := suite.store.DeleteRole(context.Background(), tc.roleID)

			if tc.shouldErr {
				suite.Error(err)
			} else {
				suite.NoError(err)
			}
		})
	}
}

func (suite *RoleStoreTestSuite) TestUpdateRole() {
	testCases := []struct {
		name         string
		roleID       int
		roleDetail   RoleUpdateDetail
		setupMocks   func()
		shouldErr    bool
		errorMessage string
	}{
		{
			name:   "Success",
			roleID: 1,
			roleDetail: RoleUpdateDetail{
				Name:               "Updated Role",
				Description:        "Updated Description",
				OrganizationUnitID: "ou1",
				Permissions: []ResourcePermissions{
					{ResourceServerID: "rs1", Permissions: []string{"perm1"}},
				},
			},
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryUpdateRole, "ou1", "Updated Role",
					"Updated Description", 1, testDeploymentID).
					Return(int64(1), nil)
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryDeleteRolePermissions, 1,
					testDeploymentID).
					Return(int64(1), nil)
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryCreateRolePermission, 1, "rs1",
					"perm1", testDeploymentID).
					Return(int64(1), nil)
			},
			shouldErr: false,
		},
		{
			name:   "NotFound",
			roleID: 999,
			roleDetail: RoleUpdateDetail{
				Name:               "Updated Role",
				Description:        "Updated Description",
				OrganizationUnitID: "ou1",
				Permissions:        []ResourcePermissions{},
			},
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryUpdateRole, "ou1", "Updated Role",
					"Updated Description", 999, testDeploymentID).
					Return(int64(0), nil)
			},
			shouldErr:    true,
			errorMessage: ErrRoleNotFound.Error(),
		},
		{
			name:   "MultipleResourceServers",
			roleID: 1,
			roleDetail: RoleUpdateDetail{
				Name:               "Updated Role",
				Description:        "Updated Description",
				OrganizationUnitID: "ou1",
				Permissions: []ResourcePermissions{
					{ResourceServerID: "rs1", Permissions: []string{"perm1"}},
					{ResourceServerID: "rs2", Permissions: []string{"perm2"}},
				},
			},
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryUpdateRole, "ou1", "Updated Role",
					"Updated Description", 1, testDeploymentID).
					Return(int64(1), nil)
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryDeleteRolePermissions, 1,
					testDeploymentID).
					Return(int64(1), nil)
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryCreateRolePermission, 1, "rs1",
					"perm1", testDeploymentID).
					Return(int64(1), nil)
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryCreateRolePermission, 1, "rs2",
					"perm2", testDeploymentID).
					Return(int64(1), nil)
			},
			shouldErr: false,
		},
		{
			name:   "DeletePermissionsError",
			roleID: 1,
			roleDetail: RoleUpdateDetail{
				Name:               "Updated Role",
				Description:        "Updated Description",
				OrganizationUnitID: "ou1",
				Permissions:        []ResourcePermissions{},
			},
			setupMocks: func() {
				deleteError := errors.New("delete permissions failed")
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryUpdateRole, "ou1", "Updated Role",
					"Updated Description", 1, testDeploymentID).
					Return(int64(1), nil)
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryDeleteRolePermissions, 1,
					testDeploymentID).
					Return(int64(0), deleteError)
			},
			shouldErr:    true,
			errorMessage: "failed to delete existing role permissions",
		},
		{
			name:   "AddPermissionsError",
			roleID: 1,
			roleDetail: RoleUpdateDetail{
				Name:               "Updated Role",
				Description:        "Updated Description",
				OrganizationUnitID: "ou1",
				Permissions: []ResourcePermissions{
					{ResourceServerID: "rs1", Permissions: []string{"perm1"}},
				},
			},
			setupMocks: func() {
				addError := errors.New("add permissions failed")
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryUpdateRole, "ou1", "Updated Role",
					"Updated Description", 1, testDeploymentID).Return(int64(1), nil)
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryDeleteRolePermissions, 1,
					testDeploymentID).Return(int64(1), nil)
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryCreateRolePermission, 1, "rs1",
					"perm1", testDeploymentID).
					Return(int64(0), addError)
			},
			shouldErr:    true,
			errorMessage: "failed to assign permissions to role",
		},
		{
			name:       "DBClientError",
			roleID:     1,
			roleDetail: RoleUpdateDetail{},
			setupMocks: func() {
				dbError := errors.New("db client error")
				suite.mockDBProvider.On("GetConfigDBClient").Return(nil, dbError)
			},
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.mockDBProvider = providermock.NewDBProviderInterfaceMock(suite.T())
			suite.mockDBClient = providermock.NewDBClientInterfaceMock(suite.T())
			suite.mockTx = modelmock.NewTxInterfaceMock(suite.T())
			suite.store = &roleStore{
				dbProvider:   suite.mockDBProvider,
				deploymentID: testDeploymentID,
			}

			tc.setupMocks()

			err := suite.store.UpdateRole(context.Background(), tc.roleID, tc.roleDetail)

			if tc.shouldErr {
				suite.Error(err)
				if tc.errorMessage != "" {
					suite.Contains(err.Error(), tc.errorMessage)
				}
			} else {
				suite.NoError(err)
			}
		})
	}
}

func (suite *RoleStoreTestSuite) TestAddAssignments() {
	testCases := []struct {
		name         string
		roleID       int
		assignments  []RoleAssignment
		setupMocks   func()
		shouldErr    bool
		errorMessage string
	}{
		{
			name:   "Success",
			roleID: 1,
			assignments: []RoleAssignment{
				{ID: testUserID1, Type: AssigneeTypeUser},
			},
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryCreateRoleAssignment, 1,
					AssigneeTypeUser, testUserID1, testDeploymentID).Return(int64(1), nil)
			},
			shouldErr: false,
		},
		{
			name:   "ExecError",
			roleID: 1,
			assignments: []RoleAssignment{
				{ID: testUserID1, Type: AssigneeTypeUser},
			},
			setupMocks: func() {
				execError := errors.New("insert failed")
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryCreateRoleAssignment, 1,
					AssigneeTypeUser, testUserID1, testDeploymentID).Return(int64(0), execError)
			},
			shouldErr:    true,
			errorMessage: "failed to add assignment to role",
		},
		{
			name:        "DBClientError",
			roleID:      1,
			assignments: []RoleAssignment{},
			setupMocks: func() {
				dbError := errors.New("db client error")
				suite.mockDBProvider.On("GetConfigDBClient").Return(nil, dbError)
			},
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.mockDBProvider = providermock.NewDBProviderInterfaceMock(suite.T())
			suite.mockDBClient = providermock.NewDBClientInterfaceMock(suite.T())
			suite.mockTx = modelmock.NewTxInterfaceMock(suite.T())
			suite.store = &roleStore{
				dbProvider:   suite.mockDBProvider,
				deploymentID: testDeploymentID,
			}

			tc.setupMocks()

			err := suite.store.AddAssignments(context.Background(), tc.roleID, tc.assignments)

			if tc.shouldErr {
				suite.Error(err)
				if tc.errorMessage != "" {
					suite.Contains(err.Error(), tc.errorMessage)
				}
			} else {
				suite.NoError(err)
			}
		})
	}
}

func (suite *RoleStoreTestSuite) TestRemoveAssignments() {
	testCases := []struct {
		name         string
		roleID       int
		assignments  []RoleAssignment
		setupMocks   func()
		shouldErr    bool
		errorMessage string
	}{
		{
			name:   "Success",
			roleID: 1,
			assignments: []RoleAssignment{
				{ID: "user1", Type: AssigneeTypeUser},
			},
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryDeleteRoleAssignmentsByIDs, 1,
					AssigneeTypeUser, "user1", testDeploymentID).Return(int64(1), nil)
			},
			shouldErr: false,
		},
		{
			name:   "ExecError",
			roleID: 1,
			assignments: []RoleAssignment{
				{ID: "user1", Type: AssigneeTypeUser},
			},
			setupMocks: func() {
				execError := errors.New("delete failed")
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryDeleteRoleAssignmentsByIDs, 1,
					AssigneeTypeUser, "user1", testDeploymentID).Return(int64(0), execError)
			},
			shouldErr:    true,
			errorMessage: "failed to remove assignment from role",
		},
		{
			name:        "DBClientError",
			roleID:      1,
			assignments: []RoleAssignment{},
			setupMocks: func() {
				dbError := errors.New("db client error")
				suite.mockDBProvider.On("GetConfigDBClient").Return(nil, dbError)
			},
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.mockDBProvider = providermock.NewDBProviderInterfaceMock(suite.T())
			suite.mockDBClient = providermock.NewDBClientInterfaceMock(suite.T())
			suite.mockTx = modelmock.NewTxInterfaceMock(suite.T())
			suite.store = &roleStore{
				dbProvider:   suite.mockDBProvider,
				deploymentID: testDeploymentID,
			}

			tc.setupMocks()

			err := suite.store.RemoveAssignments(context.Background(), tc.roleID, tc.assignments)

			if tc.shouldErr {
				suite.Error(err)
				if tc.errorMessage != "" {
					suite.Contains(err.Error(), tc.errorMessage)
				}
			} else {
				suite.NoError(err)
			}
		})
	}
}

func (suite *RoleStoreTestSuite) TestCheckRoleNameExists() {
	testCases := []struct {
		name          string
		ouID          string
		roleName      string
		setupMocks    func()
		expectedExist bool
		shouldErr     bool
	}{
		{
			name:     "Exists",
			ouID:     "ou1",
			roleName: "Admin",
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryCheckRoleNameExists, "ou1", "Admin",
					testDeploymentID).
					Return([]map[string]interface{}{
						{"count": int64(1)},
					}, nil)
			},
			expectedExist: true,
			shouldErr:     false,
		},
		{
			name:     "DoesNotExist",
			ouID:     "ou1",
			roleName: "Admin",
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryCheckRoleNameExists, "ou1", "Admin",
					testDeploymentID).
					Return([]map[string]interface{}{
						{"count": int64(0)},
					}, nil)
			},
			expectedExist: false,
			shouldErr:     false,
		},
		{
			name:     "QueryError",
			ouID:     "ou1",
			roleName: "Admin",
			setupMocks: func() {
				queryError := errors.New("query failed")
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryCheckRoleNameExists, "ou1", "Admin",
					testDeploymentID).
					Return(nil, queryError)
			},
			expectedExist: false,
			shouldErr:     true,
		},
		{
			name:     "DBClientError",
			ouID:     "ou1",
			roleName: "Admin",
			setupMocks: func() {
				dbError := errors.New("db client error")
				suite.mockDBProvider.On("GetConfigDBClient").Return(nil, dbError)
			},
			expectedExist: false,
			shouldErr:     true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.mockDBProvider = providermock.NewDBProviderInterfaceMock(suite.T())
			suite.mockDBClient = providermock.NewDBClientInterfaceMock(suite.T())
			suite.store = &roleStore{
				dbProvider:   suite.mockDBProvider,
				deploymentID: testDeploymentID,
			}

			tc.setupMocks()

			exists, err := suite.store.CheckRoleNameExists(context.Background(), tc.ouID, tc.roleName)

			if tc.shouldErr {
				suite.Error(err)
				suite.Equal(tc.expectedExist, exists)
			} else {
				suite.NoError(err)
				suite.Equal(tc.expectedExist, exists)
			}
		})
	}
}

func (suite *RoleStoreTestSuite) TestCheckRoleNameExistsExcludingID() {
	testCases := []struct {
		name          string
		ouID          string
		roleName      string
		excludeID     string
		setupMocks    func()
		expectedExist bool
		shouldErr     bool
	}{
		{
			name:      "DoesNotExist",
			ouID:      "ou1",
			roleName:  "Admin",
			excludeID: "role1",
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryCheckRoleNameExistsExcludingID, "ou1",
					"Admin", "role1", testDeploymentID).Return([]map[string]interface{}{
					{"count": int64(0)},
				}, nil)
			},
			expectedExist: false,
			shouldErr:     false,
		},
		{
			name:      "Exists",
			ouID:      "ou1",
			roleName:  "Admin",
			excludeID: "role1",
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryCheckRoleNameExistsExcludingID, "ou1",
					"Admin", "role1", testDeploymentID).Return([]map[string]interface{}{
					{"count": int64(1)},
				}, nil)
			},
			expectedExist: true,
			shouldErr:     false,
		},
		{
			name:      "QueryError",
			ouID:      "ou1",
			roleName:  "Admin",
			excludeID: "role1",
			setupMocks: func() {
				queryError := errors.New("query failed")
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryCheckRoleNameExistsExcludingID, "ou1",
					"Admin", "role1", testDeploymentID).Return(nil, queryError)
			},
			expectedExist: false,
			shouldErr:     true,
		},
		{
			name:      "DBClientError",
			ouID:      "ou1",
			roleName:  "Admin",
			excludeID: "role1",
			setupMocks: func() {
				dbError := errors.New("db client error")
				suite.mockDBProvider.On("GetConfigDBClient").Return(nil, dbError)
			},
			expectedExist: false,
			shouldErr:     true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.mockDBProvider = providermock.NewDBProviderInterfaceMock(suite.T())
			suite.mockDBClient = providermock.NewDBClientInterfaceMock(suite.T())
			suite.store = &roleStore{
				dbProvider:   suite.mockDBProvider,
				deploymentID: testDeploymentID,
			}

			tc.setupMocks()

			exists, err := suite.store.CheckRoleNameExistsExcludingID(context.Background(), tc.ouID, tc.roleName,
				tc.excludeID)

			if tc.shouldErr {
				suite.Error(err)
				suite.Equal(tc.expectedExist, exists)
			} else {
				suite.NoError(err)
				suite.Equal(tc.expectedExist, exists)
			}
		})
	}
}

func (suite *RoleStoreTestSuite) TestGetAuthorizedPermissions_Success() {
	userID := testUserID1
	groupIDs := []string{"group1"}
	requestedPermissions := []string{"perm1", "perm2"}

	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("QueryContext", mock.Anything, mock.Anything, testDeploymentID, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything).
		Return(
			[]map[string]interface{}{
				{"permission": "perm1"},
			}, nil)

	permissions, err := suite.store.GetAuthorizedPermissions(context.Background(), userID, groupIDs,
		requestedPermissions)

	suite.NoError(err)
	suite.Len(permissions, 1)
	suite.Equal("perm1", permissions[0])
}

func (suite *RoleStoreTestSuite) TestGetAuthorizedPermissions_NilGroupsHandled() {
	userID := testUserID1
	requestedPermissions := []string{"perm1"}

	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("QueryContext", mock.Anything, mock.Anything, testDeploymentID, mock.Anything, mock.Anything).
		Return([]map[string]interface{}{{"permission": "perm1"}}, nil)

	permissions, err := suite.store.GetAuthorizedPermissions(context.Background(), userID, nil, requestedPermissions)

	suite.NoError(err)
	suite.Len(permissions, 1)
}

func (suite *RoleStoreTestSuite) TestGetAuthorizedPermissions_QueryError() {
	userID := testUserID1
	groupIDs := []string{"group1"}
	requestedPermissions := []string{"perm1"}

	queryError := errors.New("query failed")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("QueryContext", mock.Anything, mock.Anything, testDeploymentID, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything).Return(nil, queryError)

	permissions, err := suite.store.GetAuthorizedPermissions(context.Background(), userID, groupIDs,
		requestedPermissions)

	suite.Error(err)
	suite.Nil(permissions)
	suite.Contains(err.Error(), "failed to get authorized permissions")
}

func (suite *RoleStoreTestSuite) TestGetAuthorizedPermissions_DBClientError() {
	userID := testUserID1
	groupIDs := []string{"group1"}
	requestedPermissions := []string{"perm1"}

	dbError := errors.New("db client error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(nil, dbError)

	permissions, err := suite.store.GetAuthorizedPermissions(context.Background(), userID, groupIDs,
		requestedPermissions)

	suite.Error(err)
	suite.Nil(permissions)
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

// Tests for addPermissionsToRole with resource servers

func (suite *RoleStoreTestSuite) TestAddPermissionsToRole() {
	testCases := []struct {
		name         string
		permissions  []ResourcePermissions
		setupMocks   func()
		shouldErr    bool
		errorMessage string
	}{
		{
			name: "MultipleResourceServers",
			permissions: []ResourcePermissions{
				{ResourceServerID: "rs1", Permissions: []string{"perm1", "perm2"}},
				{ResourceServerID: "rs2", Permissions: []string{"perm3"}},
			},
			setupMocks: func() {
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryCreateRolePermission, 1, "rs1",
					"perm1", testDeploymentID).
					Return(int64(1), nil)
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryCreateRolePermission, 1, "rs1",
					"perm2", testDeploymentID).
					Return(int64(1), nil)
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryCreateRolePermission, 1, "rs2",
					"perm3", testDeploymentID).
					Return(int64(1), nil)
			},
			shouldErr: false,
		},
		{
			name: "EmptyResourceServerID",
			permissions: []ResourcePermissions{
				{ResourceServerID: "", Permissions: []string{"legacy:perm"}},
			},
			setupMocks: func() {
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryCreateRolePermission, 1, "",
					"legacy:perm", testDeploymentID).
					Return(int64(1), nil)
			},
			shouldErr: false,
		},
		{
			name:        "EmptyPermissionsList",
			permissions: []ResourcePermissions{},
			setupMocks:  func() {},
			shouldErr:   false,
		},
		{
			name: "ExecError",
			permissions: []ResourcePermissions{
				{ResourceServerID: "rs1", Permissions: []string{"perm1"}},
			},
			setupMocks: func() {
				execError := errors.New("insert permission failed")
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryCreateRolePermission, 1, "rs1",
					"perm1", testDeploymentID).
					Return(int64(0), execError)
			},
			shouldErr:    true,
			errorMessage: "failed to add permission to role",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.mockDBClient = providermock.NewDBClientInterfaceMock(suite.T())

			if tc.setupMocks != nil {
				tc.setupMocks()
			}

			err := addPermissionsToRole(context.Background(), suite.mockDBClient, 1, tc.permissions,
				testDeploymentID)

			if tc.shouldErr {
				suite.Error(err)
				if tc.errorMessage != "" {
					suite.Contains(err.Error(), tc.errorMessage)
				}
			} else {
				suite.NoError(err)
			}
		})
	}
}

// Tests for UpdateRole with resource servers

// Tests for GetAuthorizedPermissions edge cases

func (suite *RoleStoreTestSuite) TestGetAuthorizedPermissions_EmptyGroupIDs() {
	userID := "user1"
	requestedPermissions := []string{"perm1"}

	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("QueryContext", mock.Anything, mock.Anything, testDeploymentID, mock.Anything, mock.Anything).
		Return([]map[string]interface{}{{"permission": "perm1"}}, nil)

	permissions, err := suite.store.GetAuthorizedPermissions(context.Background(), userID, []string{},
		requestedPermissions)

	suite.NoError(err)
	suite.Len(permissions, 1)
}

func (suite *RoleStoreTestSuite) TestGetAuthorizedPermissions_EmptyUserID() {
	groupIDs := []string{"group1"}
	requestedPermissions := []string{"perm1"}

	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("QueryContext", mock.Anything, mock.Anything, testDeploymentID, mock.Anything, mock.Anything).
		Return([]map[string]interface{}{{"permission": "perm1"}}, nil)

	permissions, err := suite.store.GetAuthorizedPermissions(context.Background(), "", groupIDs, requestedPermissions)

	suite.NoError(err)
	suite.Len(permissions, 1)
}

func (suite *RoleStoreTestSuite) TestGetAuthorizedPermissions_MultipleGroups() {
	userID := "user1"
	groupIDs := []string{"group1", "group2", "group3"}
	requestedPermissions := []string{"perm1", "perm2"}

	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("QueryContext", mock.Anything, mock.Anything, testDeploymentID, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return([]map[string]interface{}{
			{"permission": "perm1"},
			{"permission": "perm2"},
		}, nil)

	permissions, err := suite.store.GetAuthorizedPermissions(context.Background(), userID, groupIDs,
		requestedPermissions)

	suite.NoError(err)
	suite.Len(permissions, 2)
}

func (suite *RoleStoreTestSuite) TestGetAuthorizedPermissions_InvalidPermissionType() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("QueryContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything).
		Return([]map[string]interface{}{
			{"permission": 123}, // Invalid type
		}, nil)

	permissions, err := suite.store.GetAuthorizedPermissions(context.Background(), "user1", []string{"group1"},
		[]string{"perm1"})

	suite.NoError(err)
	suite.Len(permissions, 0) // Non-string permissions are skipped
}

// Additional edge case tests

func (suite *RoleStoreTestSuite) TestGetRoleAssignments_InvalidAssigneeID() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("QueryContext", mock.Anything, queryGetRoleAssignments, 1, 10, 0, testDeploymentID).
		Return([]map[string]interface{}{
			{"assignee_id": 123, "assignee_type": "user"}, // Invalid type
		}, nil)

	assignments, err := suite.store.GetRoleAssignments(context.Background(), 1, 10, 0)

	suite.Error(err)
	suite.Nil(assignments)
}

func (suite *RoleStoreTestSuite) TestGetRoleAssignments_InvalidAssigneeType() {
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("QueryContext", mock.Anything, queryGetRoleAssignments, 1, 10, 0, testDeploymentID).
		Return([]map[string]interface{}{
			{"assignee_id": "user1", "assignee_type": 456}, // Invalid type
		}, nil)

	assignments, err := suite.store.GetRoleAssignments(context.Background(), 1, 10, 0)

	suite.Error(err)
	suite.Nil(assignments)
}

func (suite *RoleStoreTestSuite) TestGetRoleAssignments_QueryError() {
	queryError := errors.New("query failed")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("QueryContext", mock.Anything, queryGetRoleAssignments, 1, 10, 0, testDeploymentID).
		Return(nil, queryError)

	assignments, err := suite.store.GetRoleAssignments(context.Background(), 1, 10, 0)

	suite.Error(err)
	suite.Nil(assignments)
}

func (suite *RoleStoreTestSuite) TestGetRoleAssignmentsCount_QueryError() {
	queryError := errors.New("query failed")
	suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("QueryContext", mock.Anything, queryGetRoleAssignmentsCount, 1, testDeploymentID).
		Return(nil, queryError)

	count, err := suite.store.GetRoleAssignmentsCount(context.Background(), 1)

	suite.Error(err)
	suite.Equal(0, count)
}
func (suite *RoleStoreTestSuite) TestGetRoleAssignments_DBClientError() {
	dbError := errors.New("db client error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(nil, dbError)

	assignments, err := suite.store.GetRoleAssignments(context.Background(), 1, 10, 0)

	suite.Error(err)
	suite.Nil(assignments)
}

func (suite *RoleStoreTestSuite) TestGetRoleAssignmentsCount_DBClientError() {
	dbError := errors.New("db client error")
	suite.mockDBProvider.On("GetConfigDBClient").Return(nil, dbError)

	count, err := suite.store.GetRoleAssignmentsCount(context.Background(), 1)

	suite.Error(err)
	suite.Equal(0, count)
}

func (suite *RoleStoreTestSuite) TestResolveInternalID() {
	tests := []struct {
		name      string
		row       map[string]interface{}
		expected  int
		shouldErr bool
	}{
		{
			name:      "Int",
			row:       map[string]interface{}{"id": 123},
			expected:  123,
			shouldErr: false,
		},
		{
			name:      "Int64",
			row:       map[string]interface{}{"id": int64(123)},
			expected:  123,
			shouldErr: false,
		},
		{
			name:      "Float64",
			row:       map[string]interface{}{"id": float64(123.0)},
			expected:  123,
			shouldErr: false,
		},
		{
			name:      "InvalidType",
			row:       map[string]interface{}{"id": "123"},
			expected:  0,
			shouldErr: true,
		},
		{
			name:      "MissingField",
			row:       map[string]interface{}{},
			expected:  0,
			shouldErr: true,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result, err := resolveInternalID(tc.row)
			if tc.shouldErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Equal(tc.expected, result)
			}
		})
	}
}
