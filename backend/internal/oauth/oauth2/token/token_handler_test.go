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

package token

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	applicationmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/clientauth"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/model"
	"github.com/asgardeo/thunder/internal/oauth/scope"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/tests/mocks/applicationmock"
	"github.com/asgardeo/thunder/tests/mocks/jwtmock"
	"github.com/asgardeo/thunder/tests/mocks/oauth/oauth2/granthandlersmock"
	"github.com/asgardeo/thunder/tests/mocks/oauth/scopemock"
	"github.com/asgardeo/thunder/tests/mocks/observabilitymock"
	"github.com/asgardeo/thunder/tests/mocks/usermock"
)

type TokenHandlerTestSuite struct {
	suite.Suite
	mockJWTService     *jwtmock.JWTServiceInterfaceMock
	mockUserService    *usermock.UserServiceInterfaceMock
	mockAppService     *applicationmock.ApplicationServiceInterfaceMock
	mockGrantProvider  *granthandlersmock.GrantHandlerProviderInterfaceMock
	mockScopeValidator *scopemock.ScopeValidatorInterfaceMock
	mockGrantHandler   *granthandlersmock.GrantHandlerInterfaceMock
	mockObsSvc         *observabilitymock.ObservabilityServiceInterfaceMock
}

func TestTokenHandlerSuite(t *testing.T) {
	suite.Run(t, new(TokenHandlerTestSuite))
}

func (suite *TokenHandlerTestSuite) SetupTest() {
	// Initialize Thunder Runtime config with basic test config
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			ValidityPeriod: 3600,
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)
	suite.mockJWTService = jwtmock.NewJWTServiceInterfaceMock(suite.T())
	suite.mockUserService = usermock.NewUserServiceInterfaceMock(suite.T())
	suite.mockGrantProvider = granthandlersmock.NewGrantHandlerProviderInterfaceMock(suite.T())
	suite.mockAppService = applicationmock.NewApplicationServiceInterfaceMock(suite.T())
	suite.mockScopeValidator = scopemock.NewScopeValidatorInterfaceMock(suite.T())
	suite.mockGrantHandler = granthandlersmock.NewGrantHandlerInterfaceMock(suite.T())

	// Setup common mock for GetGrantHandler that can be used across tests
	// Using Maybe() allows tests to override this if needed
	suite.mockGrantProvider.On("GetGrantHandler", constants.GrantTypeAuthorizationCode).
		Return(suite.mockGrantHandler, nil).Maybe()

	suite.mockObsSvc = observabilitymock.NewObservabilityServiceInterfaceMock(suite.T())
	suite.mockObsSvc.On("IsEnabled").Return(true).Maybe()
	suite.mockObsSvc.On("PublishEvent", mock.Anything).Return().Maybe()
}

func (suite *TokenHandlerTestSuite) TestnewTokenHandler() {
	handler := newTokenHandler(suite.mockAppService, suite.mockGrantProvider, suite.mockScopeValidator,
		suite.mockObsSvc)
	assert.NotNil(suite.T(), handler)
	assert.Implements(suite.T(), (*TokenHandlerInterface)(nil), handler)
}

func (suite *TokenHandlerTestSuite) TestHandleTokenRequest_InvalidFormData() {
	handler := newTokenHandler(suite.mockAppService, suite.mockGrantProvider, suite.mockScopeValidator,
		suite.mockObsSvc)
	req, _ := http.NewRequest("POST", "/token", strings.NewReader("invalid-form-data%"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	handler.HandleTokenRequest(rr, req)

	assert.Equal(suite.T(), http.StatusBadRequest, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "invalid_request", response["error"])
	assert.Equal(suite.T(), "Failed to parse request body", response["error_description"])
}

func (suite *TokenHandlerTestSuite) TestHandleTokenRequest_MissingGrantType() {
	formData := url.Values{}
	formData.Set("client_id", "test-client-id")
	formData.Set("client_secret", "test-secret")

	suite.testTokenRequestError(formData, http.StatusBadRequest, "invalid_request",
		"Missing grant_type parameter")
}

// Helper function to test token request error scenarios
func (suite *TokenHandlerTestSuite) testTokenRequestError(formData url.Values,
	expectedStatusCode int, expectedError, expectedErrorDescription string) {
	handler := newTokenHandler(suite.mockAppService, suite.mockGrantProvider, suite.mockScopeValidator,
		suite.mockObsSvc)

	req, _ := http.NewRequest("POST", "/token", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Set up OAuth client in context (simulating middleware)
	mockApp := &applicationmodel.OAuthAppConfigProcessedDTO{
		ClientID:                "test-client-id",
		HashedClientSecret:      "hashed-secret",
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
	}
	clientInfo := &clientauth.OAuthClientInfo{
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		OAuthApp:     mockApp,
	}
	ctx := context.WithValue(req.Context(), clientauth.OAuthClientKey, clientInfo)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	handler.HandleTokenRequest(rr, req)

	assert.Equal(suite.T(), expectedStatusCode, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedError, response["error"])
	assert.Equal(suite.T(), expectedErrorDescription, response["error_description"])
}

func (suite *TokenHandlerTestSuite) TestHandleTokenRequest_InvalidGrantType() {
	formData := url.Values{}
	formData.Set("grant_type", "invalid_grant")
	formData.Set("client_id", "test-client-id")
	formData.Set("client_secret", "test-secret")

	suite.testTokenRequestError(formData, http.StatusBadRequest, "unsupported_grant_type",
		"Invalid grant_type parameter")
}

func (suite *TokenHandlerTestSuite) TestHandleTokenRequest_MissingClientID() {
	handler := newTokenHandler(suite.mockAppService, suite.mockGrantProvider, suite.mockScopeValidator,
		suite.mockObsSvc)
	formData := url.Values{}
	formData.Set("grant_type", "authorization_code")

	req, _ := http.NewRequest("POST", "/token", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	handler.HandleTokenRequest(rr, req)

	// Handler should return server error when context is missing
	assert.Equal(suite.T(), http.StatusInternalServerError, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "server_error", response["error"])
}

func (suite *TokenHandlerTestSuite) TestHandleTokenRequest_UnsupportedGrantTypeError() {
	formData := url.Values{}
	formData.Set("grant_type", "authorization_code")

	handler := newTokenHandler(suite.mockAppService, suite.mockGrantProvider, suite.mockScopeValidator,
		suite.mockObsSvc)
	req, _ := http.NewRequest("POST", "/token", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	mockApp := &applicationmodel.OAuthAppConfigProcessedDTO{
		ClientID:                "test-client-id",
		HashedClientSecret:      "hashed-secret",
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
	}
	clientInfo := &clientauth.OAuthClientInfo{
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		OAuthApp:     mockApp,
	}
	ctx := context.WithValue(req.Context(), clientauth.OAuthClientKey, clientInfo)
	req = req.WithContext(ctx)

	// Mock grant provider to return unsupported grant type error
	suite.mockGrantProvider.ExpectedCalls = nil
	suite.mockGrantProvider.On("GetGrantHandler", constants.GrantTypeAuthorizationCode).
		Return(nil, constants.UnSupportedGrantTypeError)

	rr := httptest.NewRecorder()
	handler.HandleTokenRequest(rr, req)

	assert.Equal(suite.T(), http.StatusBadRequest, rr.Code)
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "unsupported_grant_type", response["error"])
}

func (suite *TokenHandlerTestSuite) TestHandleTokenRequest_GrantHandlerProviderError() {
	formData := url.Values{}
	formData.Set("grant_type", "authorization_code")

	handler := newTokenHandler(suite.mockAppService, suite.mockGrantProvider, suite.mockScopeValidator,
		suite.mockObsSvc)
	req, _ := http.NewRequest("POST", "/token", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	mockApp := &applicationmodel.OAuthAppConfigProcessedDTO{
		ClientID:                "test-client-id",
		HashedClientSecret:      "hashed-secret",
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
	}
	clientInfo := &clientauth.OAuthClientInfo{
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		OAuthApp:     mockApp,
	}
	ctx := context.WithValue(req.Context(), clientauth.OAuthClientKey, clientInfo)
	req = req.WithContext(ctx)

	// Mock grant provider to return a generic error
	suite.mockGrantProvider.ExpectedCalls = nil
	suite.mockGrantProvider.On("GetGrantHandler", constants.GrantTypeAuthorizationCode).
		Return(nil, errors.New("internal error"))

	rr := httptest.NewRecorder()
	handler.HandleTokenRequest(rr, req)

	assert.Equal(suite.T(), http.StatusInternalServerError, rr.Code)
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "server_error", response["error"])
}

func (suite *TokenHandlerTestSuite) TestHandleTokenRequest_UnauthorizedClient() {
	formData := url.Values{}
	formData.Set("grant_type", "client_credentials")

	handler := newTokenHandler(suite.mockAppService, suite.mockGrantProvider, suite.mockScopeValidator,
		suite.mockObsSvc)
	req, _ := http.NewRequest("POST", "/token", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Mock app that doesn't allow client_credentials grant type
	mockApp := &applicationmodel.OAuthAppConfigProcessedDTO{
		ClientID:                "test-client-id",
		HashedClientSecret:      "hashed-secret",
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode}, // Only auth code
	}
	clientInfo := &clientauth.OAuthClientInfo{
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		OAuthApp:     mockApp,
	}
	ctx := context.WithValue(req.Context(), clientauth.OAuthClientKey, clientInfo)
	req = req.WithContext(ctx)

	mockClientCredentialsHandler := granthandlersmock.NewGrantHandlerInterfaceMock(suite.T())
	suite.mockGrantProvider.ExpectedCalls = nil
	suite.mockGrantProvider.On("GetGrantHandler", constants.GrantTypeClientCredentials).
		Return(mockClientCredentialsHandler, nil)

	rr := httptest.NewRecorder()
	handler.HandleTokenRequest(rr, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, rr.Code)
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "unauthorized_client", response["error"])
}

func (suite *TokenHandlerTestSuite) TestHandleTokenRequest_ValidateGrantError() {
	formData := url.Values{}
	formData.Set("grant_type", "authorization_code")
	formData.Set("code", "test-code")

	handler := newTokenHandler(suite.mockAppService, suite.mockGrantProvider, suite.mockScopeValidator,
		suite.mockObsSvc)
	req, _ := http.NewRequest("POST", "/token", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	mockApp := &applicationmodel.OAuthAppConfigProcessedDTO{
		ClientID:                "test-client-id",
		HashedClientSecret:      "hashed-secret",
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
	}
	clientInfo := &clientauth.OAuthClientInfo{
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		OAuthApp:     mockApp,
	}
	ctx := context.WithValue(req.Context(), clientauth.OAuthClientKey, clientInfo)
	req = req.WithContext(ctx)

	// Mock grant handler to return validation error
	suite.mockGrantProvider.ExpectedCalls = nil
	suite.mockGrantProvider.On("GetGrantHandler", constants.GrantTypeAuthorizationCode).
		Return(suite.mockGrantHandler, nil)

	errorResponse := &model.ErrorResponse{
		Error:            "invalid_grant",
		ErrorDescription: "Invalid authorization code",
	}
	suite.mockGrantHandler.On("ValidateGrant", mock.Anything, mockApp).
		Return(errorResponse)

	rr := httptest.NewRecorder()
	handler.HandleTokenRequest(rr, req)

	assert.Equal(suite.T(), http.StatusBadRequest, rr.Code)
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "invalid_grant", response["error"])
	assert.Equal(suite.T(), "Invalid authorization code", response["error_description"])
}

func (suite *TokenHandlerTestSuite) TestHandleTokenRequest_ScopeValidationError() {
	formData := url.Values{}
	formData.Set("grant_type", "authorization_code")
	formData.Set("code", "test-code")
	formData.Set("scope", "invalid_scope")

	handler := newTokenHandler(suite.mockAppService, suite.mockGrantProvider, suite.mockScopeValidator,
		suite.mockObsSvc)
	req, _ := http.NewRequest("POST", "/token", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	mockApp := &applicationmodel.OAuthAppConfigProcessedDTO{
		ClientID:                "test-client-id",
		HashedClientSecret:      "hashed-secret",
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
	}
	clientInfo := &clientauth.OAuthClientInfo{
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		OAuthApp:     mockApp,
	}
	ctx := context.WithValue(req.Context(), clientauth.OAuthClientKey, clientInfo)
	req = req.WithContext(ctx)

	suite.mockGrantProvider.ExpectedCalls = nil
	suite.mockGrantProvider.On("GetGrantHandler", constants.GrantTypeAuthorizationCode).
		Return(suite.mockGrantHandler, nil)

	suite.mockGrantHandler.On("ValidateGrant", mock.Anything, mockApp).
		Return(nil)

	scopeError := &scope.ScopeError{
		Error:            "invalid_scope",
		ErrorDescription: "Invalid scope requested",
	}
	suite.mockScopeValidator.On("ValidateScopes", "invalid_scope", "test-client-id").
		Return("", scopeError)

	rr := httptest.NewRecorder()
	handler.HandleTokenRequest(rr, req)

	assert.Equal(suite.T(), http.StatusBadRequest, rr.Code)
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "invalid_scope", response["error"])
}

func (suite *TokenHandlerTestSuite) TestHandleTokenRequest_Success() {
	formData := url.Values{}
	formData.Set("grant_type", "authorization_code")
	formData.Set("code", "test-code")
	formData.Set("scope", "openid profile")

	handler := newTokenHandler(suite.mockAppService, suite.mockGrantProvider, suite.mockScopeValidator,
		suite.mockObsSvc)
	req, _ := http.NewRequest("POST", "/token", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	mockApp := &applicationmodel.OAuthAppConfigProcessedDTO{
		ClientID:                "test-client-id",
		HashedClientSecret:      "hashed-secret",
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
	}
	clientInfo := &clientauth.OAuthClientInfo{
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		OAuthApp:     mockApp,
	}
	ctx := context.WithValue(req.Context(), clientauth.OAuthClientKey, clientInfo)
	req = req.WithContext(ctx)

	suite.mockGrantProvider.ExpectedCalls = nil
	suite.mockGrantProvider.On("GetGrantHandler", constants.GrantTypeAuthorizationCode).
		Return(suite.mockGrantHandler, nil)

	suite.mockGrantHandler.On("ValidateGrant", mock.Anything, mockApp).
		Return(nil)

	suite.mockScopeValidator.On("ValidateScopes", "openid profile", "test-client-id").
		Return("openid profile", nil)

	tokenResponse := &model.TokenResponseDTO{
		AccessToken: model.TokenDTO{
			Token:     "access-token-123",
			TokenType: "Bearer",
			ExpiresIn: 3600,
			Scopes:    []string{"openid", "profile"},
		},
		RefreshToken: model.TokenDTO{
			Token: "",
		},
		IDToken: model.TokenDTO{
			Token: "",
		},
	}
	suite.mockGrantHandler.On("HandleGrant", mock.Anything, mockApp).
		Return(tokenResponse, nil)

	rr := httptest.NewRecorder()
	handler.HandleTokenRequest(rr, req)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)
	assert.Equal(suite.T(), "application/json", rr.Header().Get("Content-Type"))
	assert.Equal(suite.T(), "no-store", rr.Header().Get("Cache-Control"))
	assert.Equal(suite.T(), "no-cache", rr.Header().Get("Pragma"))

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "access-token-123", response["access_token"])
	assert.Equal(suite.T(), "Bearer", response["token_type"])
	assert.Equal(suite.T(), float64(3600), response["expires_in"])
	assert.Equal(suite.T(), "openid profile", response["scope"])
}

func (suite *TokenHandlerTestSuite) TestHandleTokenRequest_HandleGrantError() {
	formData := url.Values{}
	formData.Set("grant_type", "authorization_code")
	formData.Set("code", "test-code")
	formData.Set("scope", "openid")

	handler := newTokenHandler(suite.mockAppService, suite.mockGrantProvider, suite.mockScopeValidator,
		suite.mockObsSvc)
	req, _ := http.NewRequest("POST", "/token", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	mockApp := &applicationmodel.OAuthAppConfigProcessedDTO{
		ClientID:                "test-client-id",
		HashedClientSecret:      "hashed-secret",
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
	}
	clientInfo := &clientauth.OAuthClientInfo{
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		OAuthApp:     mockApp,
	}
	ctx := context.WithValue(req.Context(), clientauth.OAuthClientKey, clientInfo)
	req = req.WithContext(ctx)

	suite.mockGrantProvider.ExpectedCalls = nil
	suite.mockGrantProvider.On("GetGrantHandler", constants.GrantTypeAuthorizationCode).
		Return(suite.mockGrantHandler, nil)

	suite.mockGrantHandler.On("ValidateGrant", mock.Anything, mockApp).
		Return(nil)

	suite.mockScopeValidator.On("ValidateScopes", "openid", "test-client-id").
		Return("openid", nil)

	errorResponse := &model.ErrorResponse{
		Error:            "invalid_grant",
		ErrorDescription: "Authorization code expired",
	}
	suite.mockGrantHandler.On("HandleGrant", mock.Anything, mockApp).
		Return(nil, errorResponse)

	rr := httptest.NewRecorder()
	handler.HandleTokenRequest(rr, req)

	assert.Equal(suite.T(), http.StatusBadRequest, rr.Code)
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "invalid_grant", response["error"])
}

func (suite *TokenHandlerTestSuite) TestHandleTokenRequest_WithRefreshToken() {
	formData := url.Values{}
	formData.Set("grant_type", "authorization_code")
	formData.Set("code", "test-code")
	formData.Set("scope", "openid")

	handler := newTokenHandler(suite.mockAppService, suite.mockGrantProvider, suite.mockScopeValidator,
		suite.mockObsSvc)
	req, _ := http.NewRequest("POST", "/token", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Mock app that allows both authorization_code and refresh_token
	mockApp := &applicationmodel.OAuthAppConfigProcessedDTO{
		ClientID:                "test-client-id",
		HashedClientSecret:      "hashed-secret",
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
		GrantTypes: []constants.GrantType{
			constants.GrantTypeAuthorizationCode,
			constants.GrantTypeRefreshToken,
		},
	}
	clientInfo := &clientauth.OAuthClientInfo{
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		OAuthApp:     mockApp,
	}
	ctx := context.WithValue(req.Context(), clientauth.OAuthClientKey, clientInfo)
	req = req.WithContext(ctx)

	suite.mockGrantProvider.ExpectedCalls = nil
	suite.mockGrantProvider.On("GetGrantHandler", constants.GrantTypeAuthorizationCode).
		Return(suite.mockGrantHandler, nil)

	mockRefreshHandler := granthandlersmock.NewRefreshTokenGrantHandlerInterfaceMock(suite.T())
	suite.mockGrantProvider.On("GetGrantHandler", constants.GrantTypeRefreshToken).
		Return(mockRefreshHandler, nil)

	suite.mockGrantHandler.On("ValidateGrant", mock.Anything, mockApp).
		Return(nil)

	suite.mockScopeValidator.On("ValidateScopes", "openid", "test-client-id").
		Return("openid", nil)

	tokenResponse := &model.TokenResponseDTO{
		AccessToken: model.TokenDTO{
			Token:     "access-token-123",
			TokenType: "Bearer",
			ExpiresIn: 3600,
			Scopes:    []string{"openid"},
			Subject:   "user123",
			Audience:  "test-audience",
		},
		RefreshToken: model.TokenDTO{
			Token: "",
		},
		IDToken: model.TokenDTO{
			Token: "",
		},
	}
	suite.mockGrantHandler.On("HandleGrant", mock.Anything, mockApp).
		Return(tokenResponse, nil)

	mockRefreshHandler.On("IssueRefreshToken", tokenResponse, mockApp, "user123", "test-audience",
		"authorization_code", mock.MatchedBy(func(scopes []string) bool {
			return len(scopes) == 1 && scopes[0] == "openid"
		})).
		Return(nil)

	rr := httptest.NewRecorder()
	handler.HandleTokenRequest(rr, req)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "access-token-123", response["access_token"])
}

func (suite *TokenHandlerTestSuite) TestHandleTokenRequest_RefreshTokenIssuanceError() {
	formData := url.Values{}
	formData.Set("grant_type", "authorization_code")
	formData.Set("code", "test-code")
	formData.Set("scope", "openid")

	handler := newTokenHandler(suite.mockAppService, suite.mockGrantProvider, suite.mockScopeValidator,
		suite.mockObsSvc)
	req, _ := http.NewRequest("POST", "/token", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	mockApp := &applicationmodel.OAuthAppConfigProcessedDTO{
		ClientID:                "test-client-id",
		HashedClientSecret:      "hashed-secret",
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
		GrantTypes: []constants.GrantType{
			constants.GrantTypeAuthorizationCode,
			constants.GrantTypeRefreshToken,
		},
	}
	clientInfo := &clientauth.OAuthClientInfo{
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		OAuthApp:     mockApp,
	}
	ctx := context.WithValue(req.Context(), clientauth.OAuthClientKey, clientInfo)
	req = req.WithContext(ctx)

	suite.mockGrantProvider.ExpectedCalls = nil
	suite.mockGrantProvider.On("GetGrantHandler", constants.GrantTypeAuthorizationCode).
		Return(suite.mockGrantHandler, nil)

	mockRefreshHandler := granthandlersmock.NewRefreshTokenGrantHandlerInterfaceMock(suite.T())
	suite.mockGrantProvider.On("GetGrantHandler", constants.GrantTypeRefreshToken).
		Return(mockRefreshHandler, nil)

	suite.mockGrantHandler.On("ValidateGrant", mock.Anything, mockApp).
		Return(nil)

	suite.mockScopeValidator.On("ValidateScopes", "openid", "test-client-id").
		Return("openid", nil)

	tokenResponse := &model.TokenResponseDTO{
		AccessToken: model.TokenDTO{
			Token:     "access-token-123",
			TokenType: "Bearer",
			ExpiresIn: 3600,
			Scopes:    []string{"openid"},
			Subject:   "user123",
			Audience:  "test-audience",
		},
		RefreshToken: model.TokenDTO{
			Token: "",
		},
		IDToken: model.TokenDTO{
			Token: "",
		},
	}
	suite.mockGrantHandler.On("HandleGrant", mock.Anything, mockApp).
		Return(tokenResponse, nil)

	refreshError := &model.ErrorResponse{
		Error:            "server_error",
		ErrorDescription: "Failed to issue refresh token",
	}
	mockRefreshHandler.On("IssueRefreshToken", tokenResponse, mockApp, "user123", "test-audience",
		"authorization_code", mock.MatchedBy(func(scopes []string) bool {
			return len(scopes) == 1 && scopes[0] == "openid"
		})).
		Return(refreshError)

	rr := httptest.NewRecorder()
	handler.HandleTokenRequest(rr, req)

	assert.Equal(suite.T(), http.StatusInternalServerError, rr.Code)
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "server_error", response["error"])
}

func (suite *TokenHandlerTestSuite) TestHandleTokenRequest_RefreshTokenHandlerNotFound() {
	formData := url.Values{}
	formData.Set("grant_type", "authorization_code")
	formData.Set("code", "test-code")
	formData.Set("scope", "openid")

	handler := newTokenHandler(suite.mockAppService, suite.mockGrantProvider, suite.mockScopeValidator,
		suite.mockObsSvc)
	req, _ := http.NewRequest("POST", "/token", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	mockApp := &applicationmodel.OAuthAppConfigProcessedDTO{
		ClientID:                "test-client-id",
		HashedClientSecret:      "hashed-secret",
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
		GrantTypes: []constants.GrantType{
			constants.GrantTypeAuthorizationCode,
			constants.GrantTypeRefreshToken,
		},
	}
	clientInfo := &clientauth.OAuthClientInfo{
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		OAuthApp:     mockApp,
	}
	ctx := context.WithValue(req.Context(), clientauth.OAuthClientKey, clientInfo)
	req = req.WithContext(ctx)

	suite.mockGrantProvider.ExpectedCalls = nil
	suite.mockGrantProvider.On("GetGrantHandler", constants.GrantTypeAuthorizationCode).
		Return(suite.mockGrantHandler, nil)

	// Return an error when getting refresh token handler
	suite.mockGrantProvider.On("GetGrantHandler", constants.GrantTypeRefreshToken).
		Return(nil, errors.New("refresh handler not found"))

	suite.mockGrantHandler.On("ValidateGrant", mock.Anything, mockApp).
		Return(nil)

	suite.mockScopeValidator.On("ValidateScopes", "openid", "test-client-id").
		Return("openid", nil)

	tokenResponse := &model.TokenResponseDTO{
		AccessToken: model.TokenDTO{
			Token:     "access-token-123",
			TokenType: "Bearer",
			ExpiresIn: 3600,
			Scopes:    []string{"openid"},
			Subject:   "user123",
			Audience:  "test-audience",
		},
		RefreshToken: model.TokenDTO{
			Token: "",
		},
		IDToken: model.TokenDTO{
			Token: "",
		},
	}
	suite.mockGrantHandler.On("HandleGrant", mock.Anything, mockApp).
		Return(tokenResponse, nil)

	rr := httptest.NewRecorder()
	handler.HandleTokenRequest(rr, req)

	assert.Equal(suite.T(), http.StatusInternalServerError, rr.Code)
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "server_error", response["error"])
}

func (suite *TokenHandlerTestSuite) TestHandleTokenRequest_RefreshTokenHandlerCastFailure() {
	formData := url.Values{}
	formData.Set("grant_type", "authorization_code")
	formData.Set("code", "test-code")
	formData.Set("scope", "openid")

	handler := newTokenHandler(suite.mockAppService, suite.mockGrantProvider, suite.mockScopeValidator,
		suite.mockObsSvc)
	req, _ := http.NewRequest("POST", "/token", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	mockApp := &applicationmodel.OAuthAppConfigProcessedDTO{
		ClientID:                "test-client-id",
		HashedClientSecret:      "hashed-secret",
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
		GrantTypes: []constants.GrantType{
			constants.GrantTypeAuthorizationCode,
			constants.GrantTypeRefreshToken,
		},
	}
	clientInfo := &clientauth.OAuthClientInfo{
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		OAuthApp:     mockApp,
	}
	ctx := context.WithValue(req.Context(), clientauth.OAuthClientKey, clientInfo)
	req = req.WithContext(ctx)

	suite.mockGrantProvider.ExpectedCalls = nil
	suite.mockGrantProvider.On("GetGrantHandler", constants.GrantTypeAuthorizationCode).
		Return(suite.mockGrantHandler, nil)

	// Return a handler that doesn't implement RefreshTokenGrantHandlerInterface
	suite.mockGrantProvider.On("GetGrantHandler", constants.GrantTypeRefreshToken).
		Return(suite.mockGrantHandler, nil)

	suite.mockGrantHandler.On("ValidateGrant", mock.Anything, mockApp).
		Return(nil)

	suite.mockScopeValidator.On("ValidateScopes", "openid", "test-client-id").
		Return("openid", nil)

	tokenResponse := &model.TokenResponseDTO{
		AccessToken: model.TokenDTO{
			Token:     "access-token-123",
			TokenType: "Bearer",
			ExpiresIn: 3600,
			Scopes:    []string{"openid"},
			Subject:   "user123",
			Audience:  "test-audience",
		},
		RefreshToken: model.TokenDTO{
			Token: "",
		},
		IDToken: model.TokenDTO{
			Token: "",
		},
	}
	suite.mockGrantHandler.On("HandleGrant", mock.Anything, mockApp).
		Return(tokenResponse, nil)

	rr := httptest.NewRecorder()
	handler.HandleTokenRequest(rr, req)

	assert.Equal(suite.T(), http.StatusInternalServerError, rr.Code)
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "server_error", response["error"])
}

func (suite *TokenHandlerTestSuite) TestHandleTokenRequest_TokenExchange() {
	formData := url.Values{}
	formData.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
	formData.Set("subject_token", "subject-token")
	formData.Set("requested_token_type", "urn:ietf:params:oauth:token-type:access_token")

	handler := newTokenHandler(suite.mockAppService, suite.mockGrantProvider, suite.mockScopeValidator,
		suite.mockObsSvc)
	req, _ := http.NewRequest("POST", "/token", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	mockApp := &applicationmodel.OAuthAppConfigProcessedDTO{
		ClientID:                "test-client-id",
		HashedClientSecret:      "hashed-secret",
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
		GrantTypes:              []constants.GrantType{constants.GrantTypeTokenExchange},
	}
	clientInfo := &clientauth.OAuthClientInfo{
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		OAuthApp:     mockApp,
	}
	ctx := context.WithValue(req.Context(), clientauth.OAuthClientKey, clientInfo)
	req = req.WithContext(ctx)

	mockTokenExchangeHandler := granthandlersmock.NewGrantHandlerInterfaceMock(suite.T())
	suite.mockGrantProvider.ExpectedCalls = nil
	suite.mockGrantProvider.On("GetGrantHandler", constants.GrantTypeTokenExchange).
		Return(mockTokenExchangeHandler, nil)

	mockTokenExchangeHandler.On("ValidateGrant", mock.Anything, mockApp).
		Return(nil)

	suite.mockScopeValidator.On("ValidateScopes", "", "test-client-id").
		Return("", nil)

	tokenResponse := &model.TokenResponseDTO{
		AccessToken: model.TokenDTO{
			Token:     "exchanged-token",
			TokenType: "Bearer",
			ExpiresIn: 3600,
			Scopes:    []string{},
		},
		RefreshToken: model.TokenDTO{
			Token: "",
		},
		IDToken: model.TokenDTO{
			Token: "",
		},
	}
	mockTokenExchangeHandler.On("HandleGrant", mock.Anything, mockApp).
		Return(tokenResponse, nil)

	rr := httptest.NewRecorder()
	handler.HandleTokenRequest(rr, req)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "exchanged-token", response["access_token"])
	assert.Equal(suite.T(), "urn:ietf:params:oauth:token-type:access_token", response["issued_token_type"])
}

func (suite *TokenHandlerTestSuite) TestHandleTokenRequest_TokenExchangeWithJWTTokenType() {
	formData := url.Values{}
	formData.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
	formData.Set("subject_token", "subject-token")
	formData.Set("requested_token_type", "urn:ietf:params:oauth:token-type:jwt")

	handler := newTokenHandler(suite.mockAppService, suite.mockGrantProvider, suite.mockScopeValidator,
		suite.mockObsSvc)
	req, _ := http.NewRequest("POST", "/token", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	mockApp := &applicationmodel.OAuthAppConfigProcessedDTO{
		ClientID:                "test-client-id",
		HashedClientSecret:      "hashed-secret",
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
		GrantTypes:              []constants.GrantType{constants.GrantTypeTokenExchange},
	}
	clientInfo := &clientauth.OAuthClientInfo{
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		OAuthApp:     mockApp,
	}
	ctx := context.WithValue(req.Context(), clientauth.OAuthClientKey, clientInfo)
	req = req.WithContext(ctx)

	mockTokenExchangeHandler := granthandlersmock.NewGrantHandlerInterfaceMock(suite.T())
	suite.mockGrantProvider.ExpectedCalls = nil
	suite.mockGrantProvider.On("GetGrantHandler", constants.GrantTypeTokenExchange).
		Return(mockTokenExchangeHandler, nil)

	mockTokenExchangeHandler.On("ValidateGrant", mock.Anything, mockApp).
		Return(nil)

	suite.mockScopeValidator.On("ValidateScopes", "", "test-client-id").
		Return("", nil)

	tokenResponse := &model.TokenResponseDTO{
		AccessToken: model.TokenDTO{
			Token:     "exchanged-token",
			TokenType: "Bearer",
			ExpiresIn: 3600,
			Scopes:    []string{},
		},
		RefreshToken: model.TokenDTO{
			Token: "",
		},
		IDToken: model.TokenDTO{
			Token: "",
		},
	}
	mockTokenExchangeHandler.On("HandleGrant", mock.Anything, mockApp).
		Return(tokenResponse, nil)

	rr := httptest.NewRecorder()
	handler.HandleTokenRequest(rr, req)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "urn:ietf:params:oauth:token-type:jwt", response["issued_token_type"])
}
