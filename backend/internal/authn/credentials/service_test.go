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

package credentials

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/user"
	"github.com/asgardeo/thunder/tests/mocks/usermock"
)

const (
	testUserID = "user123"
)

type CredentialsAuthnServiceTestSuite struct {
	suite.Suite
	mockUserService *usermock.UserServiceInterfaceMock
	service         CredentialsAuthnServiceInterface
}

func TestCredentialsAuthnServiceTestSuite(t *testing.T) {
	suite.Run(t, new(CredentialsAuthnServiceTestSuite))
}

func (suite *CredentialsAuthnServiceTestSuite) SetupTest() {
	suite.mockUserService = usermock.NewUserServiceInterfaceMock(suite.T())
	suite.service = newCredentialsAuthnService(suite.mockUserService)
}

func (suite *CredentialsAuthnServiceTestSuite) TestAuthenticateSuccess() {
	attributes := map[string]interface{}{
		"username": "testuser",
		"password": "testpass",
	}
	userID := testUserID
	orgUnit := "test-ou"

	authResp := &user.AuthenticateUserResponse{
		ID: userID,
	}
	user := &user.User{
		ID:               userID,
		Type:             "person",
		OrganizationUnit: orgUnit,
	}

	suite.mockUserService.On("AuthenticateUser", mock.Anything, mock.Anything).Return(authResp, nil)
	suite.mockUserService.On("GetUser", mock.Anything, userID).Return(user, nil)

	result, err := suite.service.Authenticate(attributes)
	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(userID, result.ID)
	suite.Equal(orgUnit, result.OrganizationUnit)
}

func (suite *CredentialsAuthnServiceTestSuite) TestAuthenticateFailures() {
	cases := []struct {
		name              string
		attributes        map[string]interface{}
		setupMock         func(m *usermock.UserServiceInterfaceMock)
		expectedErrorCode string
	}{
		{
			name:              "EmptyAttributes",
			attributes:        map[string]interface{}{},
			setupMock:         nil,
			expectedErrorCode: ErrorEmptyAttributesOrCredentials.Code,
		},
		{
			name: "UserNotFound",
			attributes: map[string]interface{}{
				"username": "nonexistent",
				"password": "testpass",
			},
			setupMock: func(m *usermock.UserServiceInterfaceMock) {
				m.On("AuthenticateUser", mock.Anything, mock.Anything).Return(nil, &user.ErrorUserNotFound)
			},
			expectedErrorCode: common.ErrorUserNotFound.Code,
		},
		{
			name: "InvalidCredentials",
			attributes: map[string]interface{}{
				"username": "testuser",
				"password": "wrongpass",
			},
			setupMock: func(m *usermock.UserServiceInterfaceMock) {
				m.On("AuthenticateUser", mock.Anything, mock.Anything).Return(nil, &user.ErrorAuthenticationFailed)
			},
			expectedErrorCode: ErrorInvalidCredentials.Code,
		},
	}

	for _, tc := range cases {
		suite.T().Run(tc.name, func(t *testing.T) {
			m := usermock.NewUserServiceInterfaceMock(t)
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			svc := newCredentialsAuthnService(m)

			result, err := svc.Authenticate(tc.attributes)
			suite.Nil(result)
			suite.NotNil(err)
			suite.Equal(tc.expectedErrorCode, err.Code)
			m.AssertExpectations(t)
		})
	}
}

func (suite *CredentialsAuthnServiceTestSuite) TestAuthenticateWithServiceErrors() {
	cases := []struct {
		name               string
		attributes         map[string]interface{}
		setupMock          func(m *usermock.UserServiceInterfaceMock)
		expectedErrorCode  string
		expectedErrContain string
	}{
		{
			name: "UserServiceServerError",
			attributes: map[string]interface{}{
				"username": "testuser",
				"password": "testpass",
			},
			setupMock: func(m *usermock.UserServiceInterfaceMock) {
				serverErr := &serviceerror.ServiceError{
					Type:             serviceerror.ServerErrorType,
					Code:             "INTERNAL_ERROR",
					ErrorDescription: "Database connection failed",
				}
				m.On("AuthenticateUser", mock.Anything, mock.Anything).Return(nil, serverErr)
			},
			expectedErrorCode: serviceerror.InternalServerError.Code,
		},
		{
			name: "UserServiceClientError",
			attributes: map[string]interface{}{
				"username": "testuser",
				"password": "testpass",
			},
			setupMock: func(m *usermock.UserServiceInterfaceMock) {
				clientErr := &serviceerror.ServiceError{
					Type:             serviceerror.ClientErrorType,
					Code:             "CUSTOM_ERROR",
					ErrorDescription: "Custom error message",
				}
				m.On("AuthenticateUser", mock.Anything, mock.Anything).Return(nil, clientErr)
			},
			expectedErrorCode:  ErrorClientErrorFromUserSvcAuthentication.Code,
			expectedErrContain: "Custom error message",
		},
		{
			name: "GetUserServerError",
			attributes: map[string]interface{}{
				"username": "testuser",
				"password": "testpass",
			},
			setupMock: func(m *usermock.UserServiceInterfaceMock) {
				userID := testUserID
				authResp := &user.AuthenticateUserResponse{ID: userID}
				serverErr := &serviceerror.ServiceError{
					Type:             serviceerror.ServerErrorType,
					Code:             "INTERNAL_ERROR",
					ErrorDescription: "Database connection failed",
				}
				m.On("AuthenticateUser", mock.Anything, mock.Anything).Return(authResp, nil)
				m.On("GetUser", mock.Anything, userID).Return(nil, serverErr)
			},
			expectedErrorCode: serviceerror.InternalServerError.Code,
		},
		{
			name: "GetUserClientError",
			attributes: map[string]interface{}{
				"username": "testuser",
				"password": "testpass",
			},
			setupMock: func(m *usermock.UserServiceInterfaceMock) {
				userID := testUserID
				authResp := &user.AuthenticateUserResponse{ID: userID}
				clientErr := &serviceerror.ServiceError{
					Type:             serviceerror.ClientErrorType,
					Code:             "CUSTOM_ERROR",
					ErrorDescription: "User locked",
				}
				m.On("AuthenticateUser", mock.Anything, mock.Anything).Return(authResp, nil)
				m.On("GetUser", mock.Anything, userID).Return(nil, clientErr)
			},
			expectedErrorCode:  ErrorClientErrorFromUserSvcAuthentication.Code,
			expectedErrContain: "User locked",
		},
	}

	for _, tc := range cases {
		suite.T().Run(tc.name, func(t *testing.T) {
			m := usermock.NewUserServiceInterfaceMock(t)
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			svc := newCredentialsAuthnService(m)

			result, err := svc.Authenticate(tc.attributes)
			suite.Nil(result)
			suite.NotNil(err)
			suite.Equal(tc.expectedErrorCode, err.Code)

			if tc.expectedErrContain != "" {
				suite.Contains(err.ErrorDescription, tc.expectedErrContain)
			}
			m.AssertExpectations(t)
		})
	}
}
