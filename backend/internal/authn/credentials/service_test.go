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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/authnprovider"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/tests/mocks/authnprovidermock"
)

const (
	testUserID = "user123"
	testToken  = "test_token"
)

type CredentialsAuthnServiceTestSuite struct {
	suite.Suite
	mockAuthnProvider *authnprovidermock.AuthnProviderInterfaceMock
	service           CredentialsAuthnServiceInterface
}

func TestCredentialsAuthnServiceTestSuite(t *testing.T) {
	suite.Run(t, new(CredentialsAuthnServiceTestSuite))
}

func (suite *CredentialsAuthnServiceTestSuite) SetupTest() {
	suite.mockAuthnProvider = authnprovidermock.NewAuthnProviderInterfaceMock(suite.T())
	suite.service = newCredentialsAuthnService(suite.mockAuthnProvider)
}

func (suite *CredentialsAuthnServiceTestSuite) TestAuthenticateSuccess() {
	identifiers := map[string]interface{}{
		"username": "testuser",
	}
	credentials := map[string]interface{}{
		"password": "testpass",
	}

	userID := testUserID
	orgUnit := "test-ou"
	userType := "person"
	userToken := "test-token"

	availableAttributes := []authnprovider.AvailableAttribute{
		{
			Name:        "username",
			DisplayName: "Username",
			Verified:    false,
		},
	}

	metadata := &authnprovider.AuthnMetadata{
		AppMetadata: map[string]interface{}{"key": "value"},
	}

	providerResponse := &authnprovider.AuthnResult{
		UserID:              userID,
		UserType:            userType,
		OrganizationUnitID:  orgUnit,
		Token:               userToken,
		AvailableAttributes: availableAttributes,
	}

	suite.mockAuthnProvider.On("Authenticate", identifiers, credentials, metadata).Return(providerResponse, nil)

	result, err := suite.service.Authenticate(identifiers, credentials, metadata)
	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(userID, result.UserID)
	suite.Equal(orgUnit, result.OrganizationUnitID)
	suite.Equal(userType, result.UserType)
	suite.Equal(userToken, result.Token)
	suite.Equal(availableAttributes, result.AvailableAttributes)
	suite.mockAuthnProvider.AssertExpectations(suite.T())
}

func (suite *CredentialsAuthnServiceTestSuite) TestAuthenticateWithNilMetadata() {
	identifiers := map[string]interface{}{
		"username": "testuser",
	}
	credentials := map[string]interface{}{
		"password": "testpass",
	}

	userID := testUserID
	orgUnit := "test-ou"
	userType := "person"
	userToken := "test-token"

	availableAttributes := []authnprovider.AvailableAttribute{
		{
			Name:        "username",
			DisplayName: "Username",
			Verified:    false,
		},
	}

	providerResponse := &authnprovider.AuthnResult{
		UserID:              userID,
		UserType:            userType,
		OrganizationUnitID:  orgUnit,
		Token:               userToken,
		AvailableAttributes: availableAttributes,
	}

	suite.mockAuthnProvider.On("Authenticate", identifiers, credentials, (*authnprovider.AuthnMetadata)(nil)).
		Return(providerResponse, nil)

	result, err := suite.service.Authenticate(identifiers, credentials, nil)
	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(userID, result.UserID)
	suite.mockAuthnProvider.AssertExpectations(suite.T())
}

func (suite *CredentialsAuthnServiceTestSuite) TestAuthenticateFailures() {
	cases := []struct {
		name              string
		identifiers       map[string]interface{}
		credentials       map[string]interface{}
		setupMock         func(m *authnprovidermock.AuthnProviderInterfaceMock)
		expectedErrorCode string
	}{
		{
			name:              "EmptyIdentifiers",
			identifiers:       map[string]interface{}{},
			credentials:       map[string]interface{}{"password": "pass"},
			setupMock:         nil,
			expectedErrorCode: ErrorEmptyAttributesOrCredentials.Code,
		},
		{
			name:              "EmptyCredentials",
			identifiers:       map[string]interface{}{"username": "user"},
			credentials:       map[string]interface{}{},
			setupMock:         nil,
			expectedErrorCode: ErrorEmptyAttributesOrCredentials.Code,
		},
		{
			name:        "UserNotFound",
			identifiers: map[string]interface{}{"username": "nonexistent"},
			credentials: map[string]interface{}{"password": "testpass"},
			setupMock: func(m *authnprovidermock.AuthnProviderInterfaceMock) {
				m.On("Authenticate", mock.Anything, mock.Anything, mock.Anything).Return(nil, authnprovider.NewError(
					authnprovider.ErrorCodeUserNotFound, "User not found", "user not found description"))
			},
			expectedErrorCode: common.ErrorUserNotFound.Code,
		},
		{
			name:        "InvalidCredentials",
			identifiers: map[string]interface{}{"username": "testuser"},
			credentials: map[string]interface{}{"password": "wrongpass"},
			setupMock: func(m *authnprovidermock.AuthnProviderInterfaceMock) {
				m.On("Authenticate", mock.Anything, mock.Anything, mock.Anything).Return(nil, authnprovider.NewError(
					authnprovider.ErrorCodeAuthenticationFailed, "Invalid credentials",
					"invalid credentials description"))
			},
			expectedErrorCode: ErrorInvalidCredentials.Code,
		},
	}

	for _, tc := range cases {
		suite.T().Run(tc.name, func(t *testing.T) {
			m := authnprovidermock.NewAuthnProviderInterfaceMock(t)
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			svc := newCredentialsAuthnService(m)

			result, err := svc.Authenticate(tc.identifiers, tc.credentials, nil)
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
		identifiers        map[string]interface{}
		credentials        map[string]interface{}
		setupMock          func(m *authnprovidermock.AuthnProviderInterfaceMock)
		expectedErrorCode  string
		expectedErrContain string
	}{
		{
			name:        "AuthnProviderSystemError",
			identifiers: map[string]interface{}{"username": "testuser"},
			credentials: map[string]interface{}{"password": "testpass"},
			setupMock: func(m *authnprovidermock.AuthnProviderInterfaceMock) {
				m.On("Authenticate", mock.Anything, mock.Anything, mock.Anything).Return(nil, authnprovider.NewError(
					authnprovider.ErrorCodeSystemError, "System error", "Database failure"))
			},
			expectedErrorCode: serviceerror.InternalServerError.Code,
		},
		{
			name:        "AuthnProviderUnknownError",
			identifiers: map[string]interface{}{"username": "testuser"},
			credentials: map[string]interface{}{"password": "testpass"},
			setupMock: func(m *authnprovidermock.AuthnProviderInterfaceMock) {
				m.On("Authenticate", mock.Anything, mock.Anything, mock.Anything).Return(nil, authnprovider.NewError(
					"UNKNOWN_CODE", "Unknown error", "Something went wrong"))
			},
			expectedErrorCode: serviceerror.InternalServerError.Code,
		},
	}

	for _, tc := range cases {
		suite.T().Run(tc.name, func(t *testing.T) {
			m := authnprovidermock.NewAuthnProviderInterfaceMock(t)
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			svc := newCredentialsAuthnService(m)

			result, err := svc.Authenticate(tc.identifiers, tc.credentials, nil)
			suite.Nil(result)
			suite.NotNil(err)
			suite.Equal(tc.expectedErrorCode, err.Code)
			m.AssertExpectations(t)
		})
	}
}

func (suite *CredentialsAuthnServiceTestSuite) TestGetAttributesSuccess() {
	token := testToken
	requestedAttributes := []string{"attr1", "attr2"}
	metadata := &authnprovider.GetAttributesMetadata{
		AppMetadata: map[string]interface{}{"key": "value"},
		Locale:      "en",
	}

	expectedResult := &authnprovider.GetAttributesResult{
		UserID:             "user123",
		UserType:           "person",
		OrganizationUnitID: "ou1",
		Attributes:         json.RawMessage(`{"attr1":"val1"}`),
	}

	suite.mockAuthnProvider.On("GetAttributes", token, requestedAttributes, &authnprovider.GetAttributesMetadata{
		AppMetadata: metadata.AppMetadata,
		Locale:      metadata.Locale,
	}).Return(expectedResult, nil)

	result, err := suite.service.GetAttributes(token, requestedAttributes, metadata)

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(expectedResult.UserID, result.UserID)
	suite.Equal(expectedResult.UserType, result.UserType)
	suite.Equal(expectedResult.OrganizationUnitID, result.OrganizationUnitID)
	suite.Equal(expectedResult.Attributes, result.Attributes)
	suite.mockAuthnProvider.AssertExpectations(suite.T())
}

func (suite *CredentialsAuthnServiceTestSuite) TestGetAttributesWithNilMetadata() {
	token := testToken
	requestedAttributes := []string{"attr1"}

	expectedResult := &authnprovider.GetAttributesResult{
		UserID:             "user123",
		UserType:           "person",
		OrganizationUnitID: "ou1",
		Attributes:         json.RawMessage(`{"attr1":"val1"}`),
	}

	suite.mockAuthnProvider.On("GetAttributes", token, requestedAttributes,
		(*authnprovider.GetAttributesMetadata)(nil)).
		Return(expectedResult, nil)

	result, err := suite.service.GetAttributes(token, requestedAttributes, nil)

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(expectedResult.UserID, result.UserID)
	suite.mockAuthnProvider.AssertExpectations(suite.T())
}

func (suite *CredentialsAuthnServiceTestSuite) TestGetAttributesFailures() {
	token := testToken
	requestedAttributes := []string{"attr1"}
	metadata := &authnprovider.GetAttributesMetadata{}

	cases := []struct {
		name              string
		setupMock         func()
		expectedErrorCode string
	}{
		{
			name: "InvalidToken",
			setupMock: func() {
				suite.mockAuthnProvider.On("GetAttributes", token, requestedAttributes, mock.Anything).
					Return(nil, authnprovider.NewError(authnprovider.ErrorCodeInvalidToken, "Invalid token",
						"Token is expired or invalid"))
			},
			expectedErrorCode: ErrorInvalidToken.Code,
		},
		{
			name: "SystemError",
			setupMock: func() {
				suite.mockAuthnProvider.On("GetAttributes", token, requestedAttributes, mock.Anything).
					Return(nil, authnprovider.NewError(authnprovider.ErrorCodeSystemError, "System error",
						"DB connection failed"))
			},
			expectedErrorCode: serviceerror.InternalServerError.Code,
		},
	}

	for _, tc := range cases {
		suite.T().Run(tc.name, func(t *testing.T) {
			suite.mockAuthnProvider = authnprovidermock.NewAuthnProviderInterfaceMock(t)
			suite.service = newCredentialsAuthnService(suite.mockAuthnProvider)

			if tc.setupMock != nil {
				tc.setupMock()
			}

			result, err := suite.service.GetAttributes(token, requestedAttributes, metadata)

			suite.Nil(result)
			suite.NotNil(err)
			suite.Equal(tc.expectedErrorCode, err.Code)
			suite.mockAuthnProvider.AssertExpectations(t)
		})
	}
}
