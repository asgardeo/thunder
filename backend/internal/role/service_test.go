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
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/group"
	oupkg "github.com/asgardeo/thunder/internal/ou"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/user"
	"github.com/asgardeo/thunder/tests/mocks/groupmock"
	"github.com/asgardeo/thunder/tests/mocks/oumock"
	"github.com/asgardeo/thunder/tests/mocks/resourcemock"
	"github.com/asgardeo/thunder/tests/mocks/usermock"
)

const (
	testUserID1 = "user1"
)

// Test Suite
type RoleServiceTestSuite struct {
	suite.Suite
	mockStore           *roleStoreInterfaceMock
	mockUserService     *usermock.UserServiceInterfaceMock
	mockGroupService    *groupmock.GroupServiceInterfaceMock
	mockOUService       *oumock.OrganizationUnitServiceInterfaceMock
	mockResourceService *resourcemock.ResourceServiceInterfaceMock
	service             RoleServiceInterface
}

func TestRoleServiceTestSuite(t *testing.T) {
	suite.Run(t, new(RoleServiceTestSuite))
}

func (suite *RoleServiceTestSuite) SetupTest() {
	suite.mockStore = newRoleStoreInterfaceMock(suite.T())
	suite.mockUserService = usermock.NewUserServiceInterfaceMock(suite.T())
	suite.mockGroupService = groupmock.NewGroupServiceInterfaceMock(suite.T())
	suite.mockOUService = oumock.NewOrganizationUnitServiceInterfaceMock(suite.T())
	suite.mockResourceService = resourcemock.NewResourceServiceInterfaceMock(suite.T())
	suite.service = newRoleService(
		suite.mockStore,
		suite.mockUserService,
		suite.mockGroupService,
		suite.mockOUService,
		suite.mockResourceService,
	)
}

// GetRoleList Tests
func (suite *RoleServiceTestSuite) TestGetRoleList_Success() {
	expectedRoles := []Role{
		{ID: "role1", Name: "Admin", OrganizationUnitID: "ou1"},
		{ID: "role2", Name: "User", OrganizationUnitID: "ou1"},
	}

	suite.mockStore.On("GetRoleListCount").Return(2, nil)
	suite.mockStore.On("GetRoleList", 10, 0).Return(expectedRoles, nil)

	result, err := suite.service.GetRoleList(10, 0)

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(2, result.TotalResults)
	suite.Equal(2, result.Count)
	suite.Equal(1, result.StartIndex)
	suite.Equal(2, len(result.Roles))
	suite.Equal("role1", result.Roles[0].ID)
	suite.Equal("Admin", result.Roles[0].Name)
	suite.Equal("role2", result.Roles[1].ID)
	suite.Equal("User", result.Roles[1].Name)
}

func (suite *RoleServiceTestSuite) TestGetRoleList_InvalidPagination() {
	testCases := []struct {
		name    string
		limit   int
		offset  int
		errCode string
	}{
		{"InvalidLimit_Zero", 0, 0, ErrorInvalidLimit.Code},
		{"InvalidLimit_TooLarge", serverconst.MaxPageSize + 1, 0, ErrorInvalidLimit.Code},
		{"InvalidOffset_Negative", 10, -1, ErrorInvalidOffset.Code},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			result, err := suite.service.GetRoleList(tc.limit, tc.offset)
			suite.Nil(result)
			suite.NotNil(err)
			suite.Equal(tc.errCode, err.Code)
		})
	}
}

func (suite *RoleServiceTestSuite) TestGetRoleList_StoreErrors() {
	testCases := []struct {
		name      string
		mockSetup func()
	}{
		{
			name: "CountError",
			mockSetup: func() {
				suite.mockStore.On("GetRoleListCount").Return(0, errors.New("database error")).Once()
			},
		},
		{
			name: "GetListError",
			mockSetup: func() {
				suite.mockStore.On("GetRoleListCount").Return(10, nil).Once()
				suite.mockStore.On("GetRoleList", 10, 0).Return([]Role{}, errors.New("database error")).Once()
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.mockSetup()

			result, err := suite.service.GetRoleList(10, 0)

			suite.Nil(result)
			suite.NotNil(err)
			suite.Equal(ErrorInternalServerError.Code, err.Code)
		})
	}
}

// CreateRole Tests
func (suite *RoleServiceTestSuite) TestCreateRole_Success() {
	request := RoleCreationDetail{
		Name:               "Test Role",
		Description:        "Test Description",
		OrganizationUnitID: "ou1",
		Permissions:        []ResourcePermissions{{ResourceServerID: "rs1", Permissions: []string{"perm1", "perm2"}}},
		Assignments: []RoleAssignment{
			{ID: testUserID1, Type: AssigneeTypeUser},
		},
	}

	ou := oupkg.OrganizationUnit{ID: "ou1", Name: "Test OU"}
	suite.mockOUService.On("GetOrganizationUnit", "ou1").Return(ou, nil)
	suite.mockResourceService.On("ValidatePermissions", "rs1", []string{"perm1", "perm2"}).Return([]string{}, nil)
	suite.mockStore.On("CheckRoleNameExists", "ou1", "Test Role").Return(false, nil)
	suite.mockUserService.On("ValidateUserIDs", mock.Anything, []string{testUserID1}).Return([]string{}, nil)
	suite.mockStore.On("CreateRole", mock.AnythingOfType("string"),
		mock.AnythingOfType("RoleCreationDetail")).Return(nil)

	result, err := suite.service.CreateRole(request)

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal("Test Role", result.Name)
	suite.Equal("Test Description", result.Description)
	suite.Equal("ou1", result.OrganizationUnitID)
	suite.Equal(1, len(result.Permissions))
	suite.Equal(2, len(result.Permissions[0].Permissions))
	// Verify permission validation was called
	suite.mockResourceService.AssertCalled(suite.T(), "ValidatePermissions", "rs1", []string{"perm1", "perm2"})
}

func (suite *RoleServiceTestSuite) TestCreateRole_ValidationErrors() {
	testCases := []struct {
		name    string
		request RoleCreationDetail
		errCode string
	}{
		{
			name: "MissingName",
			request: RoleCreationDetail{
				OrganizationUnitID: "ou1",
				Permissions: []ResourcePermissions{{
					ResourceServerID: "rs1",
					Permissions:      []string{"perm1"},
				}},
			},
			errCode: ErrorInvalidRequestFormat.Code,
		},
		{
			name: "MissingOrgUnit",
			request: RoleCreationDetail{
				Name: "Role",
				Permissions: []ResourcePermissions{{
					ResourceServerID: "rs1",
					Permissions:      []string{"perm1"},
				}},
			},
			errCode: ErrorInvalidRequestFormat.Code,
		},
		{
			name: "InvalidAssignmentType",
			request: RoleCreationDetail{
				Name:               "Role",
				OrganizationUnitID: "ou1",
				Permissions:        []ResourcePermissions{{ResourceServerID: "rs1", Permissions: []string{"perm1"}}},
				Assignments:        []RoleAssignment{{ID: testUserID1, Type: "invalid"}},
			},
			errCode: ErrorInvalidRequestFormat.Code,
		},
		{
			name: "EmptyAssignmentID",
			request: RoleCreationDetail{
				Name:               "Role",
				OrganizationUnitID: "ou1",
				Permissions:        []ResourcePermissions{{ResourceServerID: "rs1", Permissions: []string{"perm1"}}},
				Assignments:        []RoleAssignment{{ID: "", Type: AssigneeTypeUser}},
			},
			errCode: ErrorInvalidRequestFormat.Code,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			result, err := suite.service.CreateRole(tc.request)
			suite.Nil(result)
			suite.NotNil(err)
			suite.Equal(tc.errCode, err.Code)
		})
	}
}

func (suite *RoleServiceTestSuite) TestCreateRole_PermissionValidationErrors() {
	testCases := []struct {
		name          string
		request       RoleCreationDetail
		setupMocks    func()
		expectedError *serviceerror.ServiceError
	}{
		{
			name: "InvalidPermissions",
			request: RoleCreationDetail{
				Name:               "Test Role",
				OrganizationUnitID: "ou1",
				Permissions:        []ResourcePermissions{{ResourceServerID: "rs1", Permissions: []string{"perm1"}}},
			},
			setupMocks: func() {
				suite.mockResourceService.On("ValidatePermissions", "rs1", []string{"perm1"}).
					Return([]string{"perm1"}, nil).Once()
			},
			expectedError: &ErrorInvalidPermissions,
		},
		{
			name: "PermissionValidationServiceError",
			request: RoleCreationDetail{
				Name:               "Test Role",
				OrganizationUnitID: "ou1",
				Permissions:        []ResourcePermissions{{ResourceServerID: "rs1", Permissions: []string{"perm1"}}},
			},
			setupMocks: func() {
				suite.mockResourceService.On("ValidatePermissions", "rs1", []string{"perm1"}).
					Return([]string{}, &serviceerror.ServiceError{Code: "INTERNAL_ERROR"}).Once()
			},
			expectedError: &ErrorInternalServerError,
		},
		{
			name: "EmptyResourceServerID",
			request: RoleCreationDetail{
				Name:               "Test Role",
				OrganizationUnitID: "ou1",
				Permissions:        []ResourcePermissions{{ResourceServerID: "", Permissions: []string{"perm1"}}},
			},
			setupMocks: func() {
				// Resource service should not be called for empty resource server ID
			},
			expectedError: &ErrorInvalidPermissions,
		},
		{
			name: "EmptyPermissionsArray",
			request: RoleCreationDetail{
				Name:               "Test Role",
				Description:        "Test Description",
				OrganizationUnitID: "ou1",
				Permissions:        []ResourcePermissions{},
			},
			setupMocks: func() {
				ou := oupkg.OrganizationUnit{ID: "ou1"}
				suite.mockOUService.On("GetOrganizationUnit", "ou1").Return(ou, nil).Once()
				suite.mockStore.On("CheckRoleNameExists", "ou1", "Test Role").Return(false, nil).Once()
				suite.mockStore.On("CreateRole", mock.AnythingOfType("string"),
					mock.AnythingOfType("RoleCreationDetail")).Return(nil).Once()
				// Resource service should NOT be called for empty permissions
			},
			expectedError: nil, // Success case
		},
		{
			name: "MultipleResourceServers",
			request: RoleCreationDetail{
				Name:               "Test Role",
				Description:        "Test Description",
				OrganizationUnitID: "ou1",
				Permissions: []ResourcePermissions{
					{ResourceServerID: "rs1", Permissions: []string{"perm1"}},
					{ResourceServerID: "rs2", Permissions: []string{"perm2"}},
				},
			},
			setupMocks: func() {
				ou := oupkg.OrganizationUnit{ID: "ou1"}
				suite.mockOUService.On("GetOrganizationUnit", "ou1").Return(ou, nil).Once()
				suite.mockResourceService.On("ValidatePermissions", "rs1", []string{"perm1"}).
					Return([]string{}, nil).Once()
				suite.mockResourceService.On("ValidatePermissions", "rs2", []string{"perm2"}).
					Return([]string{}, nil).Once()
				suite.mockStore.On("CheckRoleNameExists", "ou1", "Test Role").Return(false, nil).Once()
				suite.mockStore.On("CreateRole", mock.AnythingOfType("string"),
					mock.AnythingOfType("RoleCreationDetail")).Return(nil).Once()
			},
			expectedError: nil, // Success case
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Setup fresh mocks for this test case
			suite.SetupTest()
			tc.setupMocks()

			result, err := suite.service.CreateRole(tc.request)

			if tc.expectedError != nil {
				suite.Nil(result)
				suite.NotNil(err)
				suite.Equal(tc.expectedError.Code, err.Code)
			} else {
				suite.Nil(err)
				suite.NotNil(result)
			}
		})
	}
}

func (suite *RoleServiceTestSuite) TestCreateRole_OrganizationUnitNotFound() {
	request := RoleCreationDetail{
		Name:               "Test Role",
		OrganizationUnitID: "nonexistent",
		Permissions:        []ResourcePermissions{{ResourceServerID: "rs1", Permissions: []string{"perm1"}}},
	}

	suite.mockResourceService.On("ValidatePermissions", "rs1", []string{"perm1"}).Return([]string{}, nil)
	suite.mockOUService.On("GetOrganizationUnit", "nonexistent").
		Return(oupkg.OrganizationUnit{}, &oupkg.ErrorOrganizationUnitNotFound)

	result, err := suite.service.CreateRole(request)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorOrganizationUnitNotFound.Code, err.Code)
}

func (suite *RoleServiceTestSuite) TestCreateRole_InvalidUserID() {
	request := RoleCreationDetail{
		Name:               "Test Role",
		OrganizationUnitID: "ou1",
		Permissions:        []ResourcePermissions{{ResourceServerID: "rs1", Permissions: []string{"perm1"}}},
		Assignments:        []RoleAssignment{{ID: "invalid_user", Type: AssigneeTypeUser}},
	}

	// Assignment validation now happens before OU and name checks
	suite.mockResourceService.On("ValidatePermissions", "rs1", []string{"perm1"}).Return([]string{}, nil)
	suite.mockUserService.On("ValidateUserIDs", mock.Anything, []string{"invalid_user"}).
		Return([]string{"invalid_user"}, nil)

	result, err := suite.service.CreateRole(request)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInvalidAssignmentID.Code, err.Code)
}

func (suite *RoleServiceTestSuite) TestCreateRole_InvalidGroupID() {
	request := RoleCreationDetail{
		Name:               "Test Role",
		OrganizationUnitID: "ou1",
		Permissions:        []ResourcePermissions{{ResourceServerID: "rs1", Permissions: []string{"perm1"}}},
		Assignments:        []RoleAssignment{{ID: "invalid_group", Type: AssigneeTypeGroup}},
	}

	// Assignment validation now happens before OU and name checks
	suite.mockResourceService.On("ValidatePermissions", "rs1", []string{"perm1"}).Return([]string{}, nil)
	suite.mockGroupService.On("ValidateGroupIDs", []string{"invalid_group"}).Return(&group.ErrorInvalidGroupMemberID)

	result, err := suite.service.CreateRole(request)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInvalidAssignmentID.Code, err.Code)
}

func (suite *RoleServiceTestSuite) TestCreateRole_StoreError() {
	request := RoleCreationDetail{
		Name:               "Test Role",
		OrganizationUnitID: "ou1",
		Permissions:        []ResourcePermissions{{ResourceServerID: "rs1", Permissions: []string{"perm1"}}},
	}

	ou := oupkg.OrganizationUnit{ID: "ou1"}
	suite.mockOUService.On("GetOrganizationUnit", "ou1").Return(ou, nil)
	suite.mockResourceService.On("ValidatePermissions", "rs1", []string{"perm1"}).Return([]string{}, nil)
	suite.mockStore.On("CheckRoleNameExists", "ou1", "Test Role").Return(false, nil)
	suite.mockStore.On("CreateRole", mock.AnythingOfType("string"),
		mock.AnythingOfType("RoleCreationDetail")).Return(errors.New("database error"))

	result, err := suite.service.CreateRole(request)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInternalServerError.Code, err.Code)
}

func (suite *RoleServiceTestSuite) TestCreateRole_NameConflict() {
	request := RoleCreationDetail{
		Name:               "Test Role",
		OrganizationUnitID: "ou1",
		Permissions:        []ResourcePermissions{{ResourceServerID: "rs1", Permissions: []string{"perm1"}}},
	}

	ou := oupkg.OrganizationUnit{ID: "ou1"}
	suite.mockOUService.On("GetOrganizationUnit", "ou1").Return(ou, nil)
	suite.mockResourceService.On("ValidatePermissions", "rs1", []string{"perm1"}).Return([]string{}, nil)
	suite.mockStore.On("CheckRoleNameExists", "ou1", "Test Role").Return(true, nil)

	result, err := suite.service.CreateRole(request)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorRoleNameConflict.Code, err.Code)
}

func (suite *RoleServiceTestSuite) TestCreateRole_CheckNameExistsError() {
	request := RoleCreationDetail{
		Name:               "Test Role",
		OrganizationUnitID: "ou1",
		Permissions:        []ResourcePermissions{{ResourceServerID: "rs1", Permissions: []string{"perm1"}}},
	}

	ou := oupkg.OrganizationUnit{ID: "ou1"}
	suite.mockOUService.On("GetOrganizationUnit", "ou1").Return(ou, nil)
	suite.mockResourceService.On("ValidatePermissions", "rs1", []string{"perm1"}).Return([]string{}, nil)
	suite.mockStore.On("CheckRoleNameExists", "ou1", "Test Role").Return(false, errors.New("database error"))

	result, err := suite.service.CreateRole(request)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInternalServerError.Code, err.Code)
}

// GetRoleWithPermissions Tests
func (suite *RoleServiceTestSuite) TestGetRole_Success() {
	expectedRole := RoleWithPermissions{
		ID:                 "role1",
		Name:               "Admin",
		Description:        "Administrator role",
		OrganizationUnitID: "ou1",
		Permissions:        []ResourcePermissions{{ResourceServerID: "rs1", Permissions: []string{"perm1", "perm2"}}},
	}

	suite.mockStore.On("GetRole", "role1").Return(expectedRole, nil)

	result, err := suite.service.GetRoleWithPermissions("role1")

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(expectedRole.ID, result.ID)
	suite.Equal(expectedRole.Name, result.Name)
}

func (suite *RoleServiceTestSuite) TestGetRole_MissingID() {
	result, err := suite.service.GetRoleWithPermissions("")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorMissingRoleID.Code, err.Code)
}

func (suite *RoleServiceTestSuite) TestGetRole_NotFound() {
	suite.mockStore.On("GetRole", "nonexistent").Return(RoleWithPermissions{}, ErrRoleNotFound)

	result, err := suite.service.GetRoleWithPermissions("nonexistent")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorRoleNotFound.Code, err.Code)
}

func (suite *RoleServiceTestSuite) TestGetRole_StoreError() {
	suite.mockStore.On("GetRole", "role1").Return(RoleWithPermissions{}, errors.New("database error"))

	result, err := suite.service.GetRoleWithPermissions("role1")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInternalServerError.Code, err.Code)
}

// UpdateRole Tests
func (suite *RoleServiceTestSuite) TestUpdateRole_MissingRoleID() {
	request := RoleUpdateDetail{
		Name:               "New Name",
		OrganizationUnitID: "ou1",
		Permissions:        []ResourcePermissions{{ResourceServerID: "rs1", Permissions: []string{"perm1"}}},
	}

	result, err := suite.service.UpdateRoleWithPermissions("", request)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorMissingRoleID.Code, err.Code)
}

func (suite *RoleServiceTestSuite) TestUpdateRole_ValidationErrors() {
	testCases := []struct {
		name    string
		request RoleUpdateDetail
		errCode string
	}{
		{
			name: "MissingName",
			request: RoleUpdateDetail{
				OrganizationUnitID: "ou1",
				Permissions: []ResourcePermissions{{
					ResourceServerID: "rs1",
					Permissions:      []string{"perm1"},
				}},
			},
			errCode: ErrorInvalidRequestFormat.Code,
		},
		{
			name: "MissingOrgUnit",
			request: RoleUpdateDetail{
				Name: "Role",
				Permissions: []ResourcePermissions{{
					ResourceServerID: "rs1",
					Permissions:      []string{"perm1"},
				}},
			},
			errCode: ErrorInvalidRequestFormat.Code,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			result, err := suite.service.UpdateRoleWithPermissions("role1", tc.request)
			suite.Nil(result)
			suite.NotNil(err)
			suite.Equal(tc.errCode, err.Code)
		})
	}
}

func (suite *RoleServiceTestSuite) TestUpdateRole_GetRoleError() {
	request := RoleUpdateDetail{
		Name:               "New Name",
		OrganizationUnitID: "ou1",
		Permissions:        []ResourcePermissions{{ResourceServerID: "rs1", Permissions: []string{"perm1"}}},
	}

	suite.mockResourceService.On("ValidatePermissions", "rs1", []string{"perm1"}).Return([]string{}, nil)
	suite.mockStore.On("IsRoleExist", "role1").Return(false, errors.New("database error"))

	result, err := suite.service.UpdateRoleWithPermissions("role1", request)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInternalServerError.Code, err.Code)
}

func (suite *RoleServiceTestSuite) TestUpdateRole_OUNotFound() {
	request := RoleUpdateDetail{
		Name:               "New Name",
		OrganizationUnitID: "nonexistent_ou",
		Permissions:        []ResourcePermissions{{ResourceServerID: "rs1", Permissions: []string{"perm1"}}},
	}

	suite.mockResourceService.On("ValidatePermissions", "rs1", []string{"perm1"}).Return([]string{}, nil)
	suite.mockStore.On("IsRoleExist", "role1").Return(true, nil)
	suite.mockOUService.On("GetOrganizationUnit", "nonexistent_ou").
		Return(oupkg.OrganizationUnit{}, &oupkg.ErrorOrganizationUnitNotFound)

	result, err := suite.service.UpdateRoleWithPermissions("role1", request)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorOrganizationUnitNotFound.Code, err.Code)
}

func (suite *RoleServiceTestSuite) TestUpdateRole_OUServiceError() {
	request := RoleUpdateDetail{
		Name:               "New Name",
		OrganizationUnitID: "ou1",
		Permissions:        []ResourcePermissions{{ResourceServerID: "rs1", Permissions: []string{"perm1"}}},
	}

	suite.mockResourceService.On("ValidatePermissions", "rs1", []string{"perm1"}).Return([]string{}, nil)
	suite.mockStore.On("IsRoleExist", "role1").Return(true, nil)
	suite.mockOUService.On("GetOrganizationUnit", "ou1").
		Return(oupkg.OrganizationUnit{}, &serviceerror.ServiceError{Code: "INTERNAL_ERROR"})

	result, err := suite.service.UpdateRoleWithPermissions("role1", request)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInternalServerError.Code, err.Code)
}

func (suite *RoleServiceTestSuite) TestUpdateRole_UpdateStoreError() {
	request := RoleUpdateDetail{
		Name:               "New Name",
		OrganizationUnitID: "ou1",
		Permissions:        []ResourcePermissions{{ResourceServerID: "rs1", Permissions: []string{"perm1"}}},
	}

	ou := oupkg.OrganizationUnit{ID: "ou1"}
	suite.mockResourceService.On("ValidatePermissions", "rs1", []string{"perm1"}).Return([]string{}, nil)
	suite.mockStore.On("IsRoleExist", "role1").Return(true, nil)
	suite.mockOUService.On("GetOrganizationUnit", "ou1").Return(ou, nil)
	suite.mockStore.On("CheckRoleNameExistsExcludingID", "ou1", "New Name", "role1").Return(false, nil)
	suite.mockStore.On("UpdateRole", mock.AnythingOfType("string"),
		mock.AnythingOfType("RoleUpdateDetail")).Return(errors.New("update error"))

	result, err := suite.service.UpdateRoleWithPermissions("role1", request)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInternalServerError.Code, err.Code)
}

func (suite *RoleServiceTestSuite) TestUpdateRole_Success() {
	// This test also verifies permission validation is called correctly during update
	request := RoleUpdateDetail{
		Name:               "New Name",
		Description:        "Updated description",
		OrganizationUnitID: "ou1",
		Permissions:        []ResourcePermissions{{ResourceServerID: "rs1", Permissions: []string{"perm1", "perm2"}}},
	}

	ou := oupkg.OrganizationUnit{ID: "ou1"}
	suite.mockResourceService.On("ValidatePermissions", "rs1", []string{"perm1", "perm2"}).Return([]string{}, nil)
	suite.mockStore.On("IsRoleExist", "role1").Return(true, nil)
	suite.mockOUService.On("GetOrganizationUnit", "ou1").Return(ou, nil)
	suite.mockStore.On("CheckRoleNameExistsExcludingID", "ou1", "New Name", "role1").Return(false, nil)
	suite.mockStore.On("UpdateRole", mock.AnythingOfType("string"), mock.AnythingOfType("RoleUpdateDetail")).Return(nil)

	result, err := suite.service.UpdateRoleWithPermissions("role1", request)

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal("New Name", result.Name)
	suite.Equal("Updated description", result.Description)
	// Verify permission validation was called
	suite.mockResourceService.AssertCalled(suite.T(), "ValidatePermissions", "rs1", []string{"perm1", "perm2"})
}

func (suite *RoleServiceTestSuite) TestUpdateRole_RoleNotFound() {
	request := RoleUpdateDetail{
		Name:               "New Name",
		OrganizationUnitID: "ou1",
		Permissions:        []ResourcePermissions{{ResourceServerID: "rs1", Permissions: []string{"perm1"}}},
	}

	suite.mockResourceService.On("ValidatePermissions", "rs1", []string{"perm1"}).Return([]string{}, nil)
	suite.mockStore.On("IsRoleExist", "nonexistent").Return(false, nil)

	result, err := suite.service.UpdateRoleWithPermissions("nonexistent", request)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorRoleNotFound.Code, err.Code)
}

func (suite *RoleServiceTestSuite) TestUpdateRole_NameConflict() {
	request := RoleUpdateDetail{
		Name:               "Conflicting Name",
		OrganizationUnitID: "ou1",
		Permissions:        []ResourcePermissions{{ResourceServerID: "rs1", Permissions: []string{"perm1"}}},
	}

	ou := oupkg.OrganizationUnit{ID: "ou1"}
	suite.mockResourceService.On("ValidatePermissions", "rs1", []string{"perm1"}).Return([]string{}, nil)
	suite.mockStore.On("IsRoleExist", "role1").Return(true, nil)
	suite.mockOUService.On("GetOrganizationUnit", "ou1").Return(ou, nil)
	suite.mockStore.On("CheckRoleNameExistsExcludingID", "ou1", "Conflicting Name", "role1").Return(true, nil)

	result, err := suite.service.UpdateRoleWithPermissions("role1", request)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorRoleNameConflict.Code, err.Code)
}

func (suite *RoleServiceTestSuite) TestUpdateRole_CheckNameExistsError() {
	request := RoleUpdateDetail{
		Name:               "New Name",
		OrganizationUnitID: "ou1",
		Permissions:        []ResourcePermissions{{ResourceServerID: "rs1", Permissions: []string{"perm1"}}},
	}

	ou := oupkg.OrganizationUnit{ID: "ou1"}
	suite.mockResourceService.On("ValidatePermissions", "rs1", []string{"perm1"}).Return([]string{}, nil)
	suite.mockStore.On("IsRoleExist", "role1").Return(true, nil)
	suite.mockOUService.On("GetOrganizationUnit", "ou1").Return(ou, nil)
	suite.mockStore.On("CheckRoleNameExistsExcludingID", "ou1", "New Name", "role1").
		Return(false, errors.New("database error"))

	result, err := suite.service.UpdateRoleWithPermissions("role1", request)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInternalServerError.Code, err.Code)
}

func (suite *RoleServiceTestSuite) TestUpdateRole_PermissionValidationErrors() {
	testCases := []struct {
		name          string
		request       RoleUpdateDetail
		setupMocks    func()
		expectedError *serviceerror.ServiceError
	}{
		{
			name: "InvalidPermissionsOnUpdate",
			request: RoleUpdateDetail{
				Name:               "Updated Role",
				OrganizationUnitID: "ou1",
				Permissions:        []ResourcePermissions{{ResourceServerID: "rs1", Permissions: []string{"perm1"}}},
			},
			setupMocks: func() {
				// Permission validation happens before IsRoleExist check in UpdateRole
				suite.mockResourceService.On("ValidatePermissions", "rs1", []string{"perm1"}).
					Return([]string{"perm1"}, nil).Once()
			},
			expectedError: &ErrorInvalidPermissions,
		},
		{
			name: "PermissionValidationServiceError",
			request: RoleUpdateDetail{
				Name:               "Updated Role",
				OrganizationUnitID: "ou1",
				Permissions:        []ResourcePermissions{{ResourceServerID: "rs1", Permissions: []string{"perm1"}}},
			},
			setupMocks: func() {
				// Permission validation happens before IsRoleExist check in UpdateRole
				suite.mockResourceService.On("ValidatePermissions", "rs1", []string{"perm1"}).
					Return([]string{}, &serviceerror.ServiceError{Code: "INTERNAL_ERROR"}).Once()
			},
			expectedError: &ErrorInternalServerError,
		},
		{
			name: "EmptyResourceServerIDOnUpdate",
			request: RoleUpdateDetail{
				Name:               "Updated Role",
				OrganizationUnitID: "ou1",
				Permissions:        []ResourcePermissions{{ResourceServerID: "", Permissions: []string{"perm1"}}},
			},
			setupMocks: func() {
				// Resource service should not be called for empty resource server ID
				// Early validation should fail before any other calls
			},
			expectedError: &ErrorInvalidPermissions,
		},
		{
			name: "MultipleResourceServersOnUpdate",
			request: RoleUpdateDetail{
				Name:               "Updated Role",
				Description:        "Updated description",
				OrganizationUnitID: "ou1",
				Permissions: []ResourcePermissions{
					{ResourceServerID: "rs1", Permissions: []string{"perm1"}},
					{ResourceServerID: "rs2", Permissions: []string{"perm2"}},
				},
			},
			setupMocks: func() {
				ou := oupkg.OrganizationUnit{ID: "ou1"}
				suite.mockStore.On("IsRoleExist", "role1").Return(true, nil).Once()
				suite.mockResourceService.On("ValidatePermissions", "rs1", []string{"perm1"}).
					Return([]string{}, nil).Once()
				suite.mockResourceService.On("ValidatePermissions", "rs2", []string{"perm2"}).
					Return([]string{}, nil).Once()
				suite.mockOUService.On("GetOrganizationUnit", "ou1").Return(ou, nil).Once()
				suite.mockStore.On("CheckRoleNameExistsExcludingID", "ou1",
					"Updated Role", "role1").Return(false, nil).Once()
				suite.mockStore.On("UpdateRole", mock.AnythingOfType("string"),
					mock.AnythingOfType("RoleUpdateDetail")).Return(nil).Once()
			},
			expectedError: nil, // Success case
		},
		{
			name: "EmptyPermissionsArrayOnUpdate",
			request: RoleUpdateDetail{
				Name:               "Updated Role",
				Description:        "Updated description",
				OrganizationUnitID: "ou1",
				Permissions:        []ResourcePermissions{},
			},
			setupMocks: func() {
				ou := oupkg.OrganizationUnit{ID: "ou1"}
				suite.mockStore.On("IsRoleExist", "role1").Return(true, nil).Once()
				suite.mockOUService.On("GetOrganizationUnit", "ou1").Return(ou, nil).Once()
				suite.mockStore.On("CheckRoleNameExistsExcludingID", "ou1",
					"Updated Role", "role1").Return(false, nil).Once()
				suite.mockStore.On("UpdateRole", mock.AnythingOfType("string"),
					mock.AnythingOfType("RoleUpdateDetail")).Return(nil).Once()
				// Resource service should NOT be called for empty permissions
			},
			expectedError: nil, // Success case
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Setup fresh mocks for this test case
			suite.SetupTest()
			tc.setupMocks()

			result, err := suite.service.UpdateRoleWithPermissions("role1", tc.request)

			if tc.expectedError != nil {
				suite.Nil(result)
				suite.NotNil(err)
				suite.Equal(tc.expectedError.Code, err.Code)
			} else {
				suite.Nil(err)
				suite.NotNil(result)
			}
		})
	}
}

// DeleteRole Tests
func (suite *RoleServiceTestSuite) TestDeleteRole_Success() {
	suite.mockStore.On("IsRoleExist", "role1").Return(true, nil)
	suite.mockStore.On("GetRoleAssignmentsCount", "role1").Return(0, nil)
	suite.mockStore.On("DeleteRole", "role1").Return(nil)

	err := suite.service.DeleteRole("role1")

	suite.Nil(err)
}

func (suite *RoleServiceTestSuite) TestDeleteRole_WithAssignments() {
	suite.mockStore.On("IsRoleExist", "role1").Return(true, nil)
	suite.mockStore.On("GetRoleAssignmentsCount", "role1").Return(5, nil)

	err := suite.service.DeleteRole("role1")

	suite.NotNil(err)
	suite.Equal(ErrorCannotDeleteRole.Code, err.Code)
}

func (suite *RoleServiceTestSuite) TestDeleteRole_NotFound_ReturnsNil() {
	suite.mockStore.On("IsRoleExist", "nonexistent").Return(false, nil)

	err := suite.service.DeleteRole("nonexistent")

	suite.Nil(err)
}

func (suite *RoleServiceTestSuite) TestDeleteRole_MissingID() {
	err := suite.service.DeleteRole("")

	suite.NotNil(err)
	suite.Equal(ErrorMissingRoleID.Code, err.Code)
}

func (suite *RoleServiceTestSuite) TestDeleteRole_GetRoleError() {
	suite.mockStore.On("IsRoleExist", "role1").Return(false, errors.New("database error"))

	err := suite.service.DeleteRole("role1")

	suite.NotNil(err)
	suite.Equal(ErrorInternalServerError.Code, err.Code)
}

func (suite *RoleServiceTestSuite) TestDeleteRole_GetAssignmentsCountError() {
	suite.mockStore.On("IsRoleExist", "role1").Return(true, nil)
	suite.mockStore.On("GetRoleAssignmentsCount", "role1").Return(0, errors.New("database error"))

	err := suite.service.DeleteRole("role1")

	suite.NotNil(err)
	suite.Equal(ErrorInternalServerError.Code, err.Code)
}

func (suite *RoleServiceTestSuite) TestDeleteRole_StoreError() {
	suite.mockStore.On("IsRoleExist", "role1").Return(true, nil)
	suite.mockStore.On("GetRoleAssignmentsCount", "role1").Return(0, nil)
	suite.mockStore.On("DeleteRole", "role1").Return(errors.New("delete error"))

	err := suite.service.DeleteRole("role1")

	suite.NotNil(err)
	suite.Equal(ErrorInternalServerError.Code, err.Code)
}

// GetRoleAssignments Tests
func (suite *RoleServiceTestSuite) TestGetRoleAssignments_Success() {
	expectedAssignments := []RoleAssignment{
		{ID: testUserID1, Type: AssigneeTypeUser},
		{ID: "group1", Type: AssigneeTypeGroup},
	}

	suite.mockStore.On("IsRoleExist", "role1").Return(true, nil)
	suite.mockStore.On("GetRoleAssignmentsCount", "role1").Return(2, nil)
	suite.mockStore.On("GetRoleAssignments", "role1", 10, 0).Return(expectedAssignments, nil)

	result, err := suite.service.GetRoleAssignments("role1", 10, 0, false)

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(2, result.TotalResults)
	suite.Equal(2, result.Count)
	suite.Equal(2, len(result.Assignments))
	suite.Equal("user1", result.Assignments[0].ID)
	suite.Equal(AssigneeTypeUser, result.Assignments[0].Type)
	suite.Equal("group1", result.Assignments[1].ID)
	suite.Equal(AssigneeTypeGroup, result.Assignments[1].Type)
}

func (suite *RoleServiceTestSuite) TestGetRoleAssignments_MissingID() {
	result, err := suite.service.GetRoleAssignments("", 10, 0, false)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorMissingRoleID.Code, err.Code)
}

func (suite *RoleServiceTestSuite) TestGetRoleAssignments_InvalidPagination() {
	result, err := suite.service.GetRoleAssignments("role1", 0, 0, false)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInvalidLimit.Code, err.Code)
}

func (suite *RoleServiceTestSuite) TestGetRoleAssignments_RoleNotFound() {
	suite.mockStore.On("IsRoleExist", "nonexistent").Return(false, nil)

	result, err := suite.service.GetRoleAssignments("nonexistent", 10, 0, false)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorRoleNotFound.Code, err.Code)
}

func (suite *RoleServiceTestSuite) TestGetRoleAssignments_GetRoleError() {
	suite.mockStore.On("IsRoleExist", "role1").Return(false, errors.New("database error"))

	result, err := suite.service.GetRoleAssignments("role1", 10, 0, false)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInternalServerError.Code, err.Code)
}

func (suite *RoleServiceTestSuite) TestGetRoleAssignments_CountError() {
	suite.mockStore.On("IsRoleExist", "role1").Return(true, nil)
	suite.mockStore.On("GetRoleAssignmentsCount", "role1").Return(0, errors.New("count error"))

	result, err := suite.service.GetRoleAssignments("role1", 10, 0, false)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInternalServerError.Code, err.Code)
}

func (suite *RoleServiceTestSuite) TestGetRoleAssignments_GetListError() {
	suite.mockStore.On("IsRoleExist", "role1").Return(true, nil)
	suite.mockStore.On("GetRoleAssignmentsCount", "role1").Return(2, nil)
	suite.mockStore.On("GetRoleAssignments", "role1", 10, 0).Return([]RoleAssignment{}, errors.New("list error"))

	result, err := suite.service.GetRoleAssignments("role1", 10, 0, false)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInternalServerError.Code, err.Code)
}

func (suite *RoleServiceTestSuite) TestGetRoleAssignments_WithDisplay_Success() {
	expectedAssignments := []RoleAssignment{
		{ID: testUserID1, Type: AssigneeTypeUser},
		{ID: "group1", Type: AssigneeTypeGroup},
	}

	suite.mockStore.On("IsRoleExist", "role1").Return(true, nil)
	suite.mockStore.On("GetRoleAssignmentsCount", "role1").Return(2, nil)
	suite.mockStore.On("GetRoleAssignments", "role1", 10, 0).Return(expectedAssignments, nil)
	suite.mockUserService.On("GetUser", mock.Anything, testUserID1).Return(&user.User{ID: testUserID1}, nil).Once()
	suite.mockGroupService.On("GetGroup", "group1").Return(&group.Group{Name: "Test Group"}, nil).Once()

	result, err := suite.service.GetRoleAssignments("role1", 10, 0, true)

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(2, result.TotalResults)
	suite.Equal(2, result.Count)
	suite.Equal(testUserID1, result.Assignments[0].Display)
	suite.Equal("Test Group", result.Assignments[1].Display)
}

func (suite *RoleServiceTestSuite) TestGetRoleAssignments_WithDisplay_FetchErrors() {
	testCases := []struct {
		name            string
		assignment      RoleAssignment
		setupMock       func()
		expectedDisplay string
	}{
		{
			name:       "User fetch error",
			assignment: RoleAssignment{ID: testUserID1, Type: AssigneeTypeUser},
			setupMock: func() {
				suite.mockUserService.On("GetUser", mock.Anything, testUserID1).
					Return(nil, &serviceerror.ServiceError{Code: "USER_NOT_FOUND"}).Once()
			},
			expectedDisplay: "",
		},
		{
			name:       "Group fetch error",
			assignment: RoleAssignment{ID: "group1", Type: AssigneeTypeGroup},
			setupMock: func() {
				suite.mockGroupService.On("GetGroup", "group1").
					Return(nil, &serviceerror.ServiceError{Code: "GROUP_NOT_FOUND"}).Once()
			},
			expectedDisplay: "",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			expectedAssignments := []RoleAssignment{tc.assignment}

			suite.mockStore.On("IsRoleExist", "role1").Return(true, nil).Once()
			suite.mockStore.On("GetRoleAssignmentsCount", "role1").Return(1, nil).Once()
			suite.mockStore.On("GetRoleAssignments", "role1", 10, 0).
				Return(expectedAssignments, nil).Once()
			tc.setupMock()

			result, err := suite.service.GetRoleAssignments("role1", 10, 0, true)

			// Should succeed but with empty display name on error
			suite.Nil(err)
			suite.NotNil(result)
			suite.Equal(1, result.TotalResults)
			suite.Equal(1, result.Count)
			suite.Equal(tc.expectedDisplay, result.Assignments[0].Display)
		})
	}
}

// AddAssignments Tests
func (suite *RoleServiceTestSuite) TestAddAssignments_MissingRoleID() {
	request := []RoleAssignment{
		{ID: testUserID1, Type: AssigneeTypeUser},
	}

	err := suite.service.AddAssignments("", request)

	suite.NotNil(err)
	suite.Equal(ErrorMissingRoleID.Code, err.Code)
}

func (suite *RoleServiceTestSuite) TestAddAssignments_EmptyAssignments() {
	request := []RoleAssignment{}

	err := suite.service.AddAssignments("role1", request)

	suite.NotNil(err)
	suite.Equal(ErrorEmptyAssignments.Code, err.Code)
}

func (suite *RoleServiceTestSuite) TestAddAssignments_InvalidAssignmentFormat() {
	testCases := []struct {
		name       string
		assignment RoleAssignment
	}{
		{
			name:       "InvalidType",
			assignment: RoleAssignment{ID: testUserID1, Type: "invalid_type"},
		},
		{
			name:       "EmptyID",
			assignment: RoleAssignment{ID: "", Type: AssigneeTypeUser},
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			request := []RoleAssignment{
				tc.assignment,
			}

			err := suite.service.AddAssignments("role1", request)

			suite.NotNil(err)
			suite.Equal(ErrorInvalidRequestFormat.Code, err.Code)
		})
	}
}

func (suite *RoleServiceTestSuite) TestAddAssignments_RoleNotFound() {
	request := []RoleAssignment{
		{ID: testUserID1, Type: AssigneeTypeUser},
	}

	suite.mockStore.On("IsRoleExist", "nonexistent").Return(false, nil)

	err := suite.service.AddAssignments("nonexistent", request)

	suite.NotNil(err)
	suite.Equal(ErrorRoleNotFound.Code, err.Code)
}

func (suite *RoleServiceTestSuite) TestAddAssignments_GetRoleError() {
	request := []RoleAssignment{
		{ID: testUserID1, Type: AssigneeTypeUser},
	}

	suite.mockStore.On("IsRoleExist", "role1").Return(false, errors.New("database error"))

	err := suite.service.AddAssignments("role1", request)

	suite.NotNil(err)
	suite.Equal(ErrorInternalServerError.Code, err.Code)
}

func (suite *RoleServiceTestSuite) TestAddAssignments_StoreError() {
	request := []RoleAssignment{
		{ID: testUserID1, Type: AssigneeTypeUser},
	}

	suite.mockStore.On("IsRoleExist", "role1").Return(true, nil)
	suite.mockUserService.On("ValidateUserIDs", mock.Anything, []string{testUserID1}).Return([]string{}, nil)
	suite.mockStore.On("AddAssignments", "role1", request).Return(errors.New("store error"))

	err := suite.service.AddAssignments("role1", request)

	suite.NotNil(err)
	suite.Equal(ErrorInternalServerError.Code, err.Code)
}

func (suite *RoleServiceTestSuite) TestAddAssignments_Success() {
	request := []RoleAssignment{
		{ID: testUserID1, Type: AssigneeTypeUser},
	}

	suite.mockStore.On("IsRoleExist", "role1").Return(true, nil)
	suite.mockUserService.On("ValidateUserIDs", mock.Anything, []string{testUserID1}).Return([]string{}, nil)
	suite.mockStore.On("AddAssignments", "role1", request).Return(nil)

	err := suite.service.AddAssignments("role1", request)

	suite.Nil(err)
}

// RemoveAssignments Tests
func (suite *RoleServiceTestSuite) TestRemoveAssignments_MissingRoleID() {
	request := []RoleAssignment{
		{ID: "user1", Type: AssigneeTypeUser},
	}

	err := suite.service.RemoveAssignments("", request)

	suite.NotNil(err)
	suite.Equal(ErrorMissingRoleID.Code, err.Code)
}

func (suite *RoleServiceTestSuite) TestRemoveAssignments_EmptyAssignments() {
	request := []RoleAssignment{}

	err := suite.service.RemoveAssignments("role1", request)

	suite.NotNil(err)
	suite.Equal(ErrorEmptyAssignments.Code, err.Code)
}

func (suite *RoleServiceTestSuite) TestRemoveAssignments_RoleNotFound() {
	request := []RoleAssignment{
		{ID: "user1", Type: AssigneeTypeUser},
	}

	suite.mockStore.On("IsRoleExist", "nonexistent").Return(false, nil)

	err := suite.service.RemoveAssignments("nonexistent", request)

	suite.NotNil(err)
	suite.Equal(ErrorRoleNotFound.Code, err.Code)
}

func (suite *RoleServiceTestSuite) TestRemoveAssignments_GetRoleError() {
	request := []RoleAssignment{
		{ID: "user1", Type: AssigneeTypeUser},
	}

	suite.mockStore.On("IsRoleExist", "role1").Return(false, errors.New("database error"))

	err := suite.service.RemoveAssignments("role1", request)

	suite.NotNil(err)
	suite.Equal(ErrorInternalServerError.Code, err.Code)
}

func (suite *RoleServiceTestSuite) TestRemoveAssignments_StoreError() {
	request := []RoleAssignment{
		{ID: "user1", Type: AssigneeTypeUser},
	}

	suite.mockStore.On("IsRoleExist", "role1").Return(true, nil)
	suite.mockStore.On("RemoveAssignments", "role1", request).Return(errors.New("store error"))

	err := suite.service.RemoveAssignments("role1", request)

	suite.NotNil(err)
	suite.Equal(ErrorInternalServerError.Code, err.Code)
}

func (suite *RoleServiceTestSuite) TestRemoveAssignments_Success() {
	request := []RoleAssignment{
		{ID: "user1", Type: AssigneeTypeUser},
	}

	suite.mockStore.On("IsRoleExist", "role1").Return(true, nil)
	suite.mockStore.On("RemoveAssignments", "role1", request).Return(nil)

	err := suite.service.RemoveAssignments("role1", request)

	suite.Nil(err)
}

// validateAssignmentIDs Tests
func (suite *RoleServiceTestSuite) TestValidateAssignmentIDs_UserServiceError() {
	request := RoleCreationDetail{
		Name:               "Test Role",
		OrganizationUnitID: "ou1",
		Permissions:        []ResourcePermissions{{ResourceServerID: "rs1", Permissions: []string{"perm1"}}},
		Assignments:        []RoleAssignment{{ID: "user1", Type: AssigneeTypeUser}},
	}

	// Assignment validation now happens before OU and name checks
	suite.mockResourceService.On("ValidatePermissions", "rs1", []string{"perm1"}).Return([]string{}, nil)
	suite.mockUserService.On("ValidateUserIDs", mock.Anything, []string{"user1"}).
		Return([]string{}, &serviceerror.ServiceError{Code: "INTERNAL_ERROR"})

	result, err := suite.service.CreateRole(request)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInternalServerError.Code, err.Code)
}

func (suite *RoleServiceTestSuite) TestValidateAssignmentIDs_GroupServiceError() {
	request := RoleCreationDetail{
		Name:               "Test Role",
		OrganizationUnitID: "ou1",
		Permissions:        []ResourcePermissions{{ResourceServerID: "rs1", Permissions: []string{"perm1"}}},
		Assignments:        []RoleAssignment{{ID: "group1", Type: AssigneeTypeGroup}},
	}

	// Assignment validation now happens before OU and name checks
	suite.mockResourceService.On("ValidatePermissions", "rs1", []string{"perm1"}).Return([]string{}, nil)
	suite.mockGroupService.On("ValidateGroupIDs", []string{"group1"}).
		Return(&serviceerror.ServiceError{Code: "INTERNAL_ERROR"})

	result, err := suite.service.CreateRole(request)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInternalServerError.Code, err.Code)
}

// Utility functions tests
func (suite *RoleServiceTestSuite) TestBuildPaginationLinks() {
	testCases := []struct {
		name        string
		base        string
		limit       int
		offset      int
		totalCount  int
		expectFirst bool
		expectPrev  bool
		expectNext  bool
		expectLast  bool
	}{
		{
			name:        "FirstPage",
			base:        "/roles",
			limit:       10,
			offset:      0,
			totalCount:  30,
			expectFirst: false,
			expectPrev:  false,
			expectNext:  true,
			expectLast:  true,
		},
		{
			name:        "MiddlePage",
			base:        "/roles",
			limit:       10,
			offset:      10,
			totalCount:  30,
			expectFirst: true,
			expectPrev:  true,
			expectNext:  true,
			expectLast:  true,
		},
		{
			name:        "LastPage",
			base:        "/roles",
			limit:       10,
			offset:      20,
			totalCount:  30,
			expectFirst: true,
			expectPrev:  true,
			expectNext:  false,
			expectLast:  false,
		},
		{
			name:        "SinglePage",
			base:        "/roles",
			limit:       10,
			offset:      0,
			totalCount:  5,
			expectFirst: false,
			expectPrev:  false,
			expectNext:  false,
			expectLast:  false,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			links := buildPaginationLinks(tc.base, tc.limit, tc.offset, tc.totalCount)

			hasFirst := false
			hasPrev := false
			hasNext := false
			hasLast := false

			for _, link := range links {
				switch link.Rel {
				case "first":
					hasFirst = true
				case "prev":
					hasPrev = true
				case "next":
					hasNext = true
				case "last":
					hasLast = true
				}
			}

			suite.Equal(tc.expectFirst, hasFirst, "first link mismatch")
			suite.Equal(tc.expectPrev, hasPrev, "prev link mismatch")
			suite.Equal(tc.expectNext, hasNext, "next link mismatch")
			suite.Equal(tc.expectLast, hasLast, "last link mismatch")
		})
	}
}

// GetAuthorizedPermissions Tests - Consolidated for efficiency while maintaining coverage
func (suite *RoleServiceTestSuite) TestGetAuthorizedPermissions() {
	testCases := []struct {
		name                 string
		userID               string
		groups               []string
		requestedPermissions []string
		mockReturn           []string
		mockError            error
		expectedPermissions  []string
		expectedError        *serviceerror.ServiceError
		skipMock             bool
	}{
		{
			name:                 "Success_UserAndGroups",
			userID:               testUserID1,
			groups:               []string{"group1", "group2"},
			requestedPermissions: []string{"perm1", "perm2", "perm3"},
			mockReturn:           []string{"perm1", "perm3"},
			expectedPermissions:  []string{"perm1", "perm3"},
		},
		{
			name:                 "Success_UserOnly_NilGroupsNormalized",
			userID:               testUserID1,
			groups:               nil, // Tests both nil and empty groups normalization
			requestedPermissions: []string{"perm1", "perm2"},
			mockReturn:           []string{"perm1"},
			expectedPermissions:  []string{"perm1"},
		},
		{
			name:                 "Success_GroupsOnly",
			userID:               "",
			groups:               []string{"group1", "group2"},
			requestedPermissions: []string{"perm1", "perm2"},
			mockReturn:           []string{"perm1"},
			expectedPermissions:  []string{"perm1"},
		},
		{
			name:                 "Success_NoAuthorizedPermissions",
			userID:               testUserID1,
			groups:               []string{"group1"},
			requestedPermissions: []string{"perm1", "perm2"},
			mockReturn:           []string{}, // User has no permissions
			expectedPermissions:  []string{},
		},
		{
			name:                 "Success_AllPermissionsAuthorized",
			userID:               testUserID1,
			groups:               []string{"group1"},
			requestedPermissions: []string{"perm1", "perm2"},
			mockReturn:           []string{"perm1", "perm2"}, // All permissions authorized
			expectedPermissions:  []string{"perm1", "perm2"},
		},
		{
			name:                 "EmptyAndNilRequestedPermissions_ReturnsEmpty",
			userID:               testUserID1,
			groups:               []string{"group1"},
			requestedPermissions: nil, // Also covers empty []string{} case
			expectedPermissions:  []string{},
			skipMock:             true, // No store call for empty permissions
		},
		{
			name:                 "MissingUserAndGroups_Error",
			userID:               "",
			groups:               nil, // Covers both nil and empty cases
			requestedPermissions: []string{"perm1", "perm2"},
			expectedError:        &ErrorMissingUserOrGroups,
			skipMock:             true,
		},
		{
			name:                 "StoreError_ReturnsInternalError",
			userID:               testUserID1,
			groups:               []string{"group1"},
			requestedPermissions: []string{"perm1", "perm2"},
			mockError:            errors.New("database error"),
			expectedError:        &ErrorInternalServerError,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			if !tc.skipMock {
				normalizedGroups := tc.groups
				if normalizedGroups == nil {
					normalizedGroups = []string{}
				}
				suite.mockStore.On("GetAuthorizedPermissions", tc.userID, normalizedGroups, tc.requestedPermissions).
					Return(tc.mockReturn, tc.mockError).Once()
			}

			result, err := suite.service.GetAuthorizedPermissions(tc.userID, tc.groups, tc.requestedPermissions)

			if tc.expectedError != nil {
				suite.NotNil(err)
				suite.Equal(tc.expectedError.Code, err.Code)
				suite.Nil(result)
			} else {
				suite.Nil(err)
				suite.NotNil(result)
				if len(tc.requestedPermissions) == 0 {
					suite.Equal(0, len(result))
				} else {
					suite.Equal(len(tc.expectedPermissions), len(result))
					suite.Equal(tc.expectedPermissions, result)
				}
			}
		})
	}
}
