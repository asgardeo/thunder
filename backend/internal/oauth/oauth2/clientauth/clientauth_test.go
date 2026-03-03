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

package clientauth

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/system/crypto/hash"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/tests/mocks/applicationmock"
)

const (
	testClientID     = "test-client-id"
	testClientSecret = "test-secret"
)

type ClientAuthTestSuite struct {
	suite.Suite
	mockAppService *applicationmock.ApplicationServiceInterfaceMock
}

func TestClientAuthTestSuite(t *testing.T) {
	suite.Run(t, new(ClientAuthTestSuite))
}

func (suite *ClientAuthTestSuite) SetupTest() {
	suite.mockAppService = applicationmock.NewApplicationServiceInterfaceMock(suite.T())
}

func (suite *ClientAuthTestSuite) TestAuthenticate_Success_ClientSecretPost() {
	clientSecret := testClientSecret
	hashedSecret := hash.GenerateThumbprintFromString(clientSecret)
	mockApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:                testClientID,
		HashedClientSecret:      hashedSecret,
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
	}

	suite.mockAppService.On("GetOAuthApplication", testClientID).
		Return(mockApp, nil).Once()

	formData := url.Values{}
	formData.Set("client_id", testClientID)
	formData.Set("client_secret", clientSecret)

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()

	clientInfo, authErr := authenticate(req, suite.mockAppService)

	assert.Nil(suite.T(), authErr)
	assert.NotNil(suite.T(), clientInfo)
	if clientInfo != nil {
		assert.Equal(suite.T(), testClientID, clientInfo.ClientID)
		assert.Equal(suite.T(), clientSecret, clientInfo.ClientSecret)
		assert.NotNil(suite.T(), clientInfo.OAuthApp)
		assert.Equal(suite.T(), testClientID, clientInfo.OAuthApp.ClientID)
	}
}

func (suite *ClientAuthTestSuite) TestAuthenticate_Success_ClientSecretBasic() {
	clientSecret := testClientSecret
	hashedSecret := hash.GenerateThumbprintFromString(clientSecret)
	mockApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:                testClientID,
		HashedClientSecret:      hashedSecret,
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretBasic,
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
	}

	suite.mockAppService.On("GetOAuthApplication", testClientID).
		Return(mockApp, nil).Once()

	req, _ := http.NewRequest("POST", "/test", nil)
	req.SetBasicAuth(testClientID, clientSecret)

	clientInfo, authErr := authenticate(req, suite.mockAppService)

	assert.Nil(suite.T(), authErr)
	assert.NotNil(suite.T(), clientInfo)
	if clientInfo != nil {
		assert.Equal(suite.T(), testClientID, clientInfo.ClientID)
		assert.Equal(suite.T(), clientSecret, clientInfo.ClientSecret)
	}
}

func (suite *ClientAuthTestSuite) TestAuthenticate_Success_PublicClient() {
	mockApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:                "public-client-id",
		HashedClientSecret:      "",
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodNone,
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
		PublicClient:            true,
	}

	suite.mockAppService.On("GetOAuthApplication", "public-client-id").
		Return(mockApp, nil).Once()

	formData := url.Values{}
	formData.Set("client_id", "public-client-id")

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()

	clientInfo, authErr := authenticate(req, suite.mockAppService)

	assert.Nil(suite.T(), authErr)
	assert.NotNil(suite.T(), clientInfo)
	if clientInfo != nil {
		assert.Equal(suite.T(), "public-client-id", clientInfo.ClientID)
		assert.Equal(suite.T(), "", clientInfo.ClientSecret)
	}
}

func (suite *ClientAuthTestSuite) TestAuthenticate_MissingClientID() {
	req, _ := http.NewRequest("POST", "/test", nil)
	_ = req.ParseForm()

	clientInfo, authErr := authenticate(req, suite.mockAppService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errMissingClientID, authErr)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, authErr.ErrorCode)
	assert.Equal(suite.T(), "Missing client_id parameter", authErr.ErrorDescription)
}

func (suite *ClientAuthTestSuite) TestAuthenticate_MissingClientSecret() {
	formData := url.Values{}
	formData.Set("client_id", testClientID)

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()

	// This should succeed for public clients, but fail for confidential clients
	// Since we don't have an app yet, it will fail at app retrieval
	suite.mockAppService.On("GetOAuthApplication", testClientID).
		Return(nil, nil).Once()

	clientInfo, authErr := authenticate(req, suite.mockAppService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errInvalidClientCredentials, authErr)
}

func (suite *ClientAuthTestSuite) TestAuthenticate_InvalidBasicAuth() {
	req, _ := http.NewRequest("POST", "/test", nil)
	req.Header.Set("Authorization", "Basic invalid_base64")

	clientInfo, authErr := authenticate(req, suite.mockAppService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errInvalidAuthorizationHeader, authErr)
}

func (suite *ClientAuthTestSuite) TestAuthenticate_InvalidAuthorizationHeader() {
	req, _ := http.NewRequest("POST", "/test", nil)
	req.Header.Set("Authorization", "Bearer token")

	clientInfo, authErr := authenticate(req, suite.mockAppService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errInvalidAuthorizationHeader, authErr)
	assert.NotNil(suite.T(), authErr.ResponseHeaders)
	assert.Equal(suite.T(), "Basic", authErr.ResponseHeaders["WWW-Authenticate"])
}

func (suite *ClientAuthTestSuite) TestAuthenticate_BothHeaderAndBody() {
	formData := url.Values{}
	formData.Set("client_id", testClientID)
	formData.Set("client_secret", testClientSecret)

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(testClientID, testClientSecret)
	_ = req.ParseForm()

	clientInfo, authErr := authenticate(req, suite.mockAppService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errBothHeaderAndBody, authErr)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, authErr.ErrorCode)
}

func (suite *ClientAuthTestSuite) TestAuthenticate_PartialHeaderAndBody() {
	formData := url.Values{}
	formData.Set("client_id", testClientID)

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(testClientID, testClientSecret)
	_ = req.ParseForm()

	clientInfo, authErr := authenticate(req, suite.mockAppService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errBothHeaderAndBody, authErr)
}

func (suite *ClientAuthTestSuite) TestAuthenticate_ClientNotFound() {
	formData := url.Values{}
	formData.Set("client_id", "non-existent-client")
	formData.Set("client_secret", testClientSecret)

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()

	suite.mockAppService.On("GetOAuthApplication", "non-existent-client").
		Return(nil, nil).Once()

	clientInfo, authErr := authenticate(req, suite.mockAppService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errInvalidClientCredentials, authErr)
}

func (suite *ClientAuthTestSuite) TestAuthenticate_InvalidClientSecret() {
	clientSecret := testClientSecret
	wrongSecret := "wrong-secret"
	hashedSecret := hash.GenerateThumbprintFromString(clientSecret)
	mockApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:                testClientID,
		HashedClientSecret:      hashedSecret,
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
	}

	suite.mockAppService.On("GetOAuthApplication", testClientID).
		Return(mockApp, nil).Once()

	formData := url.Values{}
	formData.Set("client_id", testClientID)
	formData.Set("client_secret", wrongSecret)

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()

	clientInfo, authErr := authenticate(req, suite.mockAppService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errInvalidClientCredentials, authErr)
}

func (suite *ClientAuthTestSuite) TestAuthenticate_WrongAuthMethod() {
	clientSecret := testClientSecret
	hashedSecret := hash.GenerateThumbprintFromString(clientSecret)
	mockApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:                testClientID,
		HashedClientSecret:      hashedSecret,
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
	}

	suite.mockAppService.On("GetOAuthApplication", testClientID).
		Return(mockApp, nil).Once()

	// Try to use client_secret_basic when app only allows client_secret_post
	req, _ := http.NewRequest("POST", "/test", nil)
	req.SetBasicAuth(testClientID, clientSecret)

	clientInfo, authErr := authenticate(req, suite.mockAppService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errUnauthorizedAuthMethod, authErr)
	assert.Equal(suite.T(), constants.ErrorUnauthorizedClient, authErr.ErrorCode)
}

func (suite *ClientAuthTestSuite) TestAuthenticate_PublicClientWithSecret() {
	mockApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:                "public-client-id",
		HashedClientSecret:      "",
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodNone,
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
		PublicClient:            true,
	}

	suite.mockAppService.On("GetOAuthApplication", "public-client-id").
		Return(mockApp, nil).Once()

	formData := url.Values{}
	formData.Set("client_id", "public-client-id")
	formData.Set("client_secret", "some-secret")

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()

	// Try to use client_secret_post with public client
	clientInfo, authErr := authenticate(req, suite.mockAppService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errUnauthorizedAuthMethod, authErr)
}

func (suite *ClientAuthTestSuite) TestAuthenticate_PublicClientMissingSecret() {
	mockApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:                "public-client-id",
		HashedClientSecret:      "",
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodNone,
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
		PublicClient:            true,
	}

	suite.mockAppService.On("GetOAuthApplication", "public-client-id").
		Return(mockApp, nil).Once()

	formData := url.Values{}
	formData.Set("client_id", "public-client-id")

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()

	// Public client with authMethod = none should succeed
	clientInfo, authErr := authenticate(req, suite.mockAppService)

	assert.Nil(suite.T(), authErr)
	assert.NotNil(suite.T(), clientInfo)
}

func (suite *ClientAuthTestSuite) TestAuthenticate_ClientIDMismatch() {
	clientSecret := testClientSecret
	hashedSecret := hash.GenerateThumbprintFromString(clientSecret)
	mockApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:                "different-client-id",
		HashedClientSecret:      hashedSecret,
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
	}

	suite.mockAppService.On("GetOAuthApplication", testClientID).
		Return(mockApp, nil).Once()

	formData := url.Values{}
	formData.Set("client_id", testClientID)
	formData.Set("client_secret", clientSecret)

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()

	clientInfo, authErr := authenticate(req, suite.mockAppService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errInvalidClientCredentials, authErr)
}

func (suite *ClientAuthTestSuite) TestAuthenticate_ServiceError() {
	serviceErr := &serviceerror.ServiceError{
		Code:             "APP-5001",
		Type:             serviceerror.ServerErrorType,
		Error:            "server_error",
		ErrorDescription: "Internal server error",
	}

	suite.mockAppService.On("GetOAuthApplication", testClientID).
		Return(nil, serviceErr).Once()

	formData := url.Values{}
	formData.Set("client_id", testClientID)
	formData.Set("client_secret", testClientSecret)

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()

	clientInfo, authErr := authenticate(req, suite.mockAppService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errInvalidClientCredentials, authErr)
}
