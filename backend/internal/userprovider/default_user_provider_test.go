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

package userprovider

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/user"
	"github.com/asgardeo/thunder/tests/mocks/usermock"
)

type DefaultUserProviderTestSuite struct {
	suite.Suite
	mockService *usermock.UserServiceInterfaceMock
	provider    UserProviderInterface
}

func (suite *DefaultUserProviderTestSuite) SetupTest() {
	suite.mockService = usermock.NewUserServiceInterfaceMock(suite.T())
	suite.provider = newDefaultUserProvider(suite.mockService)
}

func TestDefaultUserProviderTestSuite(t *testing.T) {
	suite.Run(t, new(DefaultUserProviderTestSuite))
}

const testUserID = "user123"

func (suite *DefaultUserProviderTestSuite) TestIdentifyUser() {
	// Test Success
	filters := map[string]interface{}{"email": "test@example.com"}
	userIDAddr := testUserID
	suite.mockService.On("IdentifyUser", mock.Anything, filters).Return(&userIDAddr, (*serviceerror.ServiceError)(nil)).
		Once()

	userID, err := suite.provider.IdentifyUser(filters)
	suite.Nil(err)
	suite.Equal(testUserID, *userID)

	// Test Not Found
	notFoundErr := &user.ErrorUserNotFound
	suite.mockService.On("IdentifyUser", mock.Anything, filters).Return(nil, notFoundErr).Once()

	userID, err = suite.provider.IdentifyUser(filters)
	suite.Nil(userID)
	suite.NotNil(err)
	suite.Equal(ErrorCodeUserNotFound, err.Code)

	// Test System Error
	sysErr := &serviceerror.ServiceError{Code: "SYS_ERR", Error: "System Error"}
	suite.mockService.On("IdentifyUser", mock.Anything, filters).Return(nil, sysErr).Once()

	userID, err = suite.provider.IdentifyUser(filters)
	suite.Nil(userID)
	suite.NotNil(err)
	suite.Equal(ErrorCodeSystemError, err.Code)
}

func (suite *DefaultUserProviderTestSuite) TestGetUser() {
	userID := testUserID
	expectedUser := &user.User{
		ID:               userID,
		Type:             "customer",
		OrganizationUnit: "ou1",
		Attributes:       json.RawMessage(`{"attr":"val"}`),
	}

	// Test Success
	suite.mockService.On("GetUser", mock.Anything, userID).Return(expectedUser, (*serviceerror.ServiceError)(nil)).
		Once()

	u, err := suite.provider.GetUser(userID)
	suite.Nil(err)
	suite.Equal(userID, u.UserID)
	suite.Equal("customer", u.UserType)
	suite.Equal("ou1", u.OrganizationUnitID)

	// Test Not Found
	suite.mockService.On("GetUser", mock.Anything, userID).Return(nil, &user.ErrorUserNotFound).
		Once()

	u, err = suite.provider.GetUser(userID)
	suite.Nil(u)
	suite.NotNil(err)
	suite.Equal(ErrorCodeUserNotFound, err.Code)
}

func (suite *DefaultUserProviderTestSuite) TestGetUserGroups() {
	userID := testUserID
	limit := 10
	offset := 0

	groupListResponse := &user.UserGroupListResponse{
		Groups: []user.UserGroup{
			{ID: "g1", Name: "Group 1", OrganizationUnitID: "ou1"},
		},
		Links: []user.Link{
			{Href: "/groups/next", Rel: "next"},
		},
		TotalResults: 1,
	}

	// Test Success
	suite.mockService.On("GetUserGroups", mock.Anything, userID, limit, offset).
		Return(groupListResponse, (*serviceerror.ServiceError)(nil)).Once()

	resp, err := suite.provider.GetUserGroups(userID, limit, offset)
	suite.Nil(err)
	suite.Equal(1, len(resp.Groups))
	suite.Equal("g1", resp.Groups[0].ID)
	suite.Equal("Group 1", resp.Groups[0].Name)
	suite.Equal("ou1", resp.Groups[0].OrganizationUnitID)

	// Test User Not Found
	suite.mockService.On("GetUserGroups", mock.Anything, userID, limit, offset).Return(nil, &user.ErrorUserNotFound).
		Once()

	resp, err = suite.provider.GetUserGroups(userID, limit, offset)
	suite.Nil(resp)
	suite.NotNil(err)
	suite.Equal(ErrorCodeUserNotFound, err.Code)
}

func (suite *DefaultUserProviderTestSuite) TestUpdateUser() {
	userID := testUserID
	updateUser := &User{
		UserID:             userID,
		UserType:           "customer",
		OrganizationUnitID: "ou1",
		Attributes:         json.RawMessage(`{"updated":"true"}`),
	}

	backendUser := &user.User{
		ID:               userID,
		Type:             "customer",
		OrganizationUnit: "ou1",
		Attributes:       json.RawMessage(`{"updated":"true"}`),
	}

	// Test Success
	suite.mockService.On("UpdateUser", mock.Anything, userID, mock.MatchedBy(func(u *user.User) bool {
		return u.ID == userID && u.Type == "customer"
	})).Return(backendUser, (*serviceerror.ServiceError)(nil)).Once()

	u, err := suite.provider.UpdateUser(userID, updateUser)
	suite.Nil(err)
	suite.Equal(userID, u.UserID)

	// Test Invalid Request
	suite.mockService.On("UpdateUser", mock.Anything, userID, mock.Anything).
		Return(nil, &user.ErrorInvalidRequestFormat).Once()
	u, err = suite.provider.UpdateUser(userID, updateUser)
	suite.Nil(u)
	suite.NotNil(err)
	suite.Equal(ErrorCodeInvalidRequestFormat, err.Code)

	// Test Attribute Conflict
	suite.mockService.On("UpdateUser", mock.Anything, userID, mock.Anything).
		Return(nil, &user.ErrorAttributeConflict).Once()
	u, err = suite.provider.UpdateUser(userID, updateUser)
	suite.Nil(u)
	suite.NotNil(err)
	suite.Equal(ErrorCodeAttributeConflict, err.Code)

	// Test Nil Configuration
	u, err = suite.provider.UpdateUser(userID, nil)
	suite.Nil(u)
	suite.NotNil(err)
	suite.Equal(ErrorCodeInvalidRequestFormat, err.Code)
}

func (suite *DefaultUserProviderTestSuite) TestCreateUser() {
	newUser := &User{
		UserType:           "customer",
		OrganizationUnitID: "ou1",
		Attributes:         json.RawMessage(`{"new":"true"}`),
	}

	createdBackendUser := &user.User{
		ID:               testUserID,
		Type:             "customer",
		OrganizationUnit: "ou1",
		Attributes:       json.RawMessage(`{"new":"true"}`),
	}

	// Test Success
	suite.mockService.On("CreateUser", mock.Anything, mock.MatchedBy(func(u *user.User) bool {
		return u.Type == "customer"
	})).Return(createdBackendUser, (*serviceerror.ServiceError)(nil)).Once()

	u, err := suite.provider.CreateUser(newUser)
	suite.Nil(err)
	suite.Equal(testUserID, u.UserID)

	// Test Organization Unit Mismatch
	suite.mockService.On("CreateUser", mock.Anything, mock.Anything).
		Return(nil, &user.ErrorOrganizationUnitMismatch).Once()
	u, err = suite.provider.CreateUser(newUser)
	suite.Nil(u)
	suite.NotNil(err)
	suite.Equal(ErrorCodeOrganizationUnitMismatch, err.Code)

	// Test Nil Configuration
	u, err = suite.provider.CreateUser(nil)
	suite.Nil(u)
	suite.NotNil(err)
	suite.Equal(ErrorCodeInvalidRequestFormat, err.Code)
}

func (suite *DefaultUserProviderTestSuite) TestUpdateUserCredentials() {
	userID := testUserID
	creds := json.RawMessage(`{"password":"newpassword"}`)

	// Test Success
	suite.mockService.On("UpdateUserCredentials", mock.Anything, userID, creds).
		Return((*serviceerror.ServiceError)(nil)).Once()

	err := suite.provider.UpdateUserCredentials(userID, creds)
	suite.Nil(err)

	// Test Missing Credentials
	suite.mockService.On("UpdateUserCredentials", mock.Anything, userID, creds).
		Return(&user.ErrorMissingCredentials).Once()

	err = suite.provider.UpdateUserCredentials(userID, creds)
	suite.NotNil(err)
	suite.Equal(ErrorCodeMissingCredentials, err.Code)
}
