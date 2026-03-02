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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/system/crypto/hash"
	"github.com/asgardeo/thunder/tests/mocks/applicationmock"
	"github.com/asgardeo/thunder/tests/mocks/jose/jwtmock"
	"github.com/asgardeo/thunder/tests/mocks/oauth/oauth2/discoverymock"
)

type ClientAuthMiddlewareTestSuite struct {
	suite.Suite
	mockAppService       *applicationmock.ApplicationServiceInterfaceMock
	mockJwtService       *jwtmock.JWTServiceInterfaceMock
	mockDiscoveryService *discoverymock.DiscoveryServiceInterfaceMock
}

func TestClientAuthMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(ClientAuthMiddlewareTestSuite))
}

func (suite *ClientAuthMiddlewareTestSuite) SetupTest() {
	suite.mockAppService = applicationmock.NewApplicationServiceInterfaceMock(suite.T())
	suite.mockJwtService = jwtmock.NewJWTServiceInterfaceMock(suite.T())
	suite.mockDiscoveryService = discoverymock.NewDiscoveryServiceInterfaceMock(suite.T())
}

func (suite *ClientAuthMiddlewareTestSuite) TestClientAuthMiddleware_Success_ClientSecretPost() {
	// Setup mock OAuth app with correctly hashed secret
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

	// Create middleware
	middleware := ClientAuthMiddleware(suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	// Create test handler that checks context
	var clientInfo *OAuthClientInfo
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientInfo = GetOAuthClient(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	// Create request with client_secret_post
	formData := url.Values{}
	formData.Set("client_id", testClientID)
	formData.Set("client_secret", clientSecret)

	req := httptest.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	// Execute middleware
	middleware(handler).ServeHTTP(w, req)

	// Verify
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.NotNil(suite.T(), clientInfo, "Expected client info in context")
	if clientInfo != nil {
		assert.Equal(suite.T(), testClientID, clientInfo.ClientID)
		assert.Equal(suite.T(), "test-secret", clientInfo.ClientSecret)
		assert.NotNil(suite.T(), clientInfo.OAuthApp)
	}
}

func (suite *ClientAuthMiddlewareTestSuite) TestClientAuthMiddleware_Success_ClientSecretBasic() {
	// Setup mock OAuth app with correctly hashed secret
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

	// Create middleware
	middleware := ClientAuthMiddleware(suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	// Create test handler
	var clientInfo *OAuthClientInfo
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientInfo = GetOAuthClient(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	// Create request with Basic Auth
	req := httptest.NewRequest("POST", "/test", nil)
	req.SetBasicAuth(testClientID, clientSecret)
	w := httptest.NewRecorder()

	// Execute middleware
	middleware(handler).ServeHTTP(w, req)

	// Verify
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.NotNil(suite.T(), clientInfo, "Expected client info in context")
	if clientInfo != nil {
		assert.Equal(suite.T(), testClientID, clientInfo.ClientID)
	}
}

func (suite *ClientAuthMiddlewareTestSuite) TestClientAuthMiddleware_MissingClientID() {
	// Create middleware
	middleware := ClientAuthMiddleware(suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create request without client_id
	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()

	// Execute middleware
	middleware(handler).ServeHTTP(w, req)

	// Verify error response
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "invalid_client", response["error"])
}

func (suite *ClientAuthMiddlewareTestSuite) TestClientAuthMiddleware_InvalidClient() {
	// Mock app service to return nil (client not found)
	suite.mockAppService.On("GetOAuthApplication", "invalid-client").
		Return(nil, nil).Once()

	// Create middleware
	middleware := ClientAuthMiddleware(suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create request with invalid client
	formData := url.Values{}
	formData.Set("client_id", "invalid-client")
	formData.Set("client_secret", "test-secret")

	req := httptest.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	// Execute middleware
	middleware(handler).ServeHTTP(w, req)

	// Verify error response
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "invalid_client", response["error"])
}

func (suite *ClientAuthMiddlewareTestSuite) TestClientAuthMiddleware_InvalidClientSecret() {
	// Setup mock OAuth app
	mockApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:                testClientID,
		HashedClientSecret:      "correct-hashed-secret",
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
	}

	suite.mockAppService.On("GetOAuthApplication", testClientID).
		Return(mockApp, nil).Once()

	// Create middleware
	middleware := ClientAuthMiddleware(suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create request with wrong client_secret
	formData := url.Values{}
	formData.Set("client_id", testClientID)
	formData.Set("client_secret", "wrong-secret")

	req := httptest.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	// Execute middleware
	middleware(handler).ServeHTTP(w, req)

	// Verify error response
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "invalid_client", response["error"])
}

func (suite *ClientAuthMiddlewareTestSuite) TestClientAuthMiddleware_HandlerNotCalledOnAuthFailure() {
	// Mock app service to return nil (client not found)
	suite.mockAppService.On("GetOAuthApplication", mock.Anything).
		Return(nil, nil).Once()

	// Create middleware
	middleware := ClientAuthMiddleware(suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	// Track if handler was called
	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// Create request with invalid client
	formData := url.Values{}
	formData.Set("client_id", "invalid-client")
	formData.Set("client_secret", "test-secret")

	req := httptest.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	// Execute middleware
	middleware(handler).ServeHTTP(w, req)

	// Verify handler was not called
	assert.False(suite.T(), handlerCalled, "Handler should not be called when authentication fails")
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

func (suite *ClientAuthMiddlewareTestSuite) TestClientAuthMiddleware_ContextPropagation() {
	// Setup mock OAuth app with correctly hashed secret
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

	// Create middleware
	middleware := ClientAuthMiddleware(suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	// Create nested handler that also checks context
	var clientInfo *OAuthClientInfo
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientInfo = GetOAuthClient(r.Context())
		// Verify context is available
		if clientInfo == nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	// Create request
	formData := url.Values{}
	formData.Set("client_id", testClientID)
	formData.Set("client_secret", clientSecret)

	req := httptest.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	// Execute middleware
	middleware(handler).ServeHTTP(w, req)

	// Verify context was propagated
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.NotNil(suite.T(), clientInfo)
}
