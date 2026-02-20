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

package authnprovider

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/user"
	"github.com/asgardeo/thunder/tests/mocks/usermock"
)

type DefaultAuthnProviderTestSuite struct {
	suite.Suite
	mockService *usermock.UserServiceInterfaceMock
	provider    AuthnProviderInterface
}

func (suite *DefaultAuthnProviderTestSuite) SetupTest() {
	suite.mockService = usermock.NewUserServiceInterfaceMock(suite.T())
	suite.provider = newDefaultAuthnProvider(suite.mockService)
}

func TestDefaultAuthnProviderTestSuite(t *testing.T) {
	suite.Run(t, new(DefaultAuthnProviderTestSuite))
}

func (suite *DefaultAuthnProviderTestSuite) TestAuthenticate_Success() {
	identifiers := map[string]interface{}{"username": "testuser"}
	credentials := map[string]interface{}{"password": "password123"}

	authReq := user.AuthenticateUserRequest{
		"username": "testuser",
		"password": "password123",
	}

	authResp := &user.AuthenticateUserResponse{
		ID:               "user123",
		Type:             "customer",
		OrganizationUnit: "ou1",
	}

	userObj := &user.User{
		ID:               "user123",
		Type:             "customer",
		OrganizationUnit: "ou1",
		Attributes:       json.RawMessage(`{"email":"test@example.com"}`),
	}

	// Expect AuthenticateUser call
	suite.mockService.On("AuthenticateUser", mock.Anything, authReq).
		Return(authResp, (*serviceerror.ServiceError)(nil)).Once()
	// Expect GetUser call
	suite.mockService.On("GetUser", mock.Anything, "user123").Return(userObj, (*serviceerror.ServiceError)(nil)).Once()

	result, err := suite.provider.Authenticate(identifiers, credentials, nil)

	suite.Nil(err)
	suite.Equal("user123", result.UserID)
	suite.Equal("user123", result.Token)
	suite.Equal("customer", result.UserType)
	suite.Equal("ou1", result.OrganizationUnitID)
	suite.Len(result.AvailableAttributes, 1)
	suite.Equal("email", result.AvailableAttributes[0].Name)
}

func (suite *DefaultAuthnProviderTestSuite) TestAuthenticate_UserNotFound() {
	identifiers := map[string]interface{}{"username": "unknown"}
	credentials := map[string]interface{}{"password": "password"}

	authReq := user.AuthenticateUserRequest{
		"username": "unknown",
		"password": "password",
	}

	userNotFoundErr := &user.ErrorUserNotFound

	suite.mockService.On("AuthenticateUser", mock.Anything, authReq).Return(nil, userNotFoundErr).Once()

	result, err := suite.provider.Authenticate(identifiers, credentials, nil)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorCodeUserNotFound, err.Code)
}

func (suite *DefaultAuthnProviderTestSuite) TestAuthenticate_AuthenticationFailed() {
	identifiers := map[string]interface{}{"username": "testuser"}
	credentials := map[string]interface{}{"password": "wrongpassword"}

	authReq := user.AuthenticateUserRequest{
		"username": "testuser",
		"password": "wrongpassword",
	}

	authFailedErr := &user.ErrorAuthenticationFailed

	suite.mockService.On("AuthenticateUser", mock.Anything, authReq).Return(nil, authFailedErr).Once()

	result, err := suite.provider.Authenticate(identifiers, credentials, nil)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorCodeAuthenticationFailed, err.Code)
}

func (suite *DefaultAuthnProviderTestSuite) TestAuthenticate_SystemError_Prepare() {
	identifiers := map[string]interface{}{"username": "testuser"}
	credentials := map[string]interface{}{"password": "password"}

	authReq := user.AuthenticateUserRequest{
		"username": "testuser",
		"password": "password",
	}

	sysErr := &serviceerror.ServiceError{Code: "SYS_ERR", Type: serviceerror.ServerErrorType, Error: "System Error"}

	suite.mockService.On("AuthenticateUser", mock.Anything, authReq).Return(nil, sysErr).Once()

	result, err := suite.provider.Authenticate(identifiers, credentials, nil)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorCodeSystemError, err.Code)
}

func (suite *DefaultAuthnProviderTestSuite) TestGetAttributes_Success_All() {
	token := "user123"
	userObj := &user.User{
		ID:               "user123",
		Type:             "customer",
		OrganizationUnit: "ou1",
		Attributes:       json.RawMessage(`{"email":"test@example.com", "age": 30}`),
	}

	suite.mockService.On("GetUser", mock.Anything, token).Return(userObj, (*serviceerror.ServiceError)(nil)).Once()

	result, err := suite.provider.GetAttributes(token, nil, nil)

	suite.Nil(err)
	suite.Equal("user123", result.UserID)
	suite.JSONEq(`{"email":"test@example.com", "age": 30}`, string(result.Attributes))
}

func (suite *DefaultAuthnProviderTestSuite) TestGetAttributes_Success_Filtered() {
	token := "user123"
	userObj := &user.User{
		ID:               "user123",
		Type:             "customer",
		OrganizationUnit: "ou1",
		Attributes:       json.RawMessage(`{"email":"test@example.com", "age": 30}`),
	}

	suite.mockService.On("GetUser", mock.Anything, token).Return(userObj, (*serviceerror.ServiceError)(nil)).Once()

	result, err := suite.provider.GetAttributes(token, []string{"email"}, nil)

	suite.Nil(err)
	suite.Equal("user123", result.UserID)
	suite.JSONEq(`{"email":"test@example.com"}`, string(result.Attributes))
}

func (suite *DefaultAuthnProviderTestSuite) TestGetAttributes_InvalidToken() {
	token := "invalid"
	notFoundErr := &user.ErrorUserNotFound

	suite.mockService.On("GetUser", mock.Anything, token).Return(nil, notFoundErr).Once()

	result, err := suite.provider.GetAttributes(token, nil, nil)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorCodeInvalidToken, err.Code)
}

func (suite *DefaultAuthnProviderTestSuite) TestAuthenticate_GetUserNotFound() {
	identifiers := map[string]interface{}{"username": "testuser"}
	credentials := map[string]interface{}{"password": "password123"}

	authReq := user.AuthenticateUserRequest{
		"username": "testuser",
		"password": "password123",
	}

	authResp := &user.AuthenticateUserResponse{
		ID:               "user123",
		Type:             "customer",
		OrganizationUnit: "ou1",
	}

	// Expect AuthenticateUser call - Success
	suite.mockService.On("AuthenticateUser", mock.Anything, authReq).
		Return(authResp, (*serviceerror.ServiceError)(nil)).Once()

	// Expect GetUser call - Fail with UserNotFound
	userNotFoundErr := &user.ErrorUserNotFound
	suite.mockService.On("GetUser", mock.Anything, "user123").Return(nil, userNotFoundErr).Once()

	result, err := suite.provider.Authenticate(identifiers, credentials, nil)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorCodeUserNotFound, err.Code)
}
