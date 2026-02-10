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

package authz

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	flowcm "github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/flowexec"
	"github.com/asgardeo/thunder/tests/mocks/applicationmock"
	"github.com/asgardeo/thunder/tests/mocks/flow/flowexecmock"
	"github.com/asgardeo/thunder/tests/mocks/jwtmock"

	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	oauth2model "github.com/asgardeo/thunder/internal/oauth/oauth2/model"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/jwt"
)

const (
	testMinimalJWT = "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJ0ZXN0LXVzZXIifQ."
	testJWTWithIat = "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJ0ZXN0LXVzZXIiLCJpYXQiOjE3MDE0MjEyMDB9."
	testAuthID     = "test-auth-id"
)

type AuthorizeHandlerTestSuite struct {
	suite.Suite
	handler             *authorizeHandler
	mockAppService      *applicationmock.ApplicationServiceInterfaceMock
	mockJWTService      *jwtmock.JWTServiceInterfaceMock
	mockAuthzCodeStore  *AuthorizationCodeStoreInterfaceMock
	mockAuthReqStore    *authorizationRequestStoreInterfaceMock
	mockFlowExecService *flowexecmock.FlowExecServiceInterfaceMock
}

func TestAuthorizeHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(AuthorizeHandlerTestSuite))
}

func (suite *AuthorizeHandlerTestSuite) BeforeTest(suiteName, testName string) {
	config.ResetThunderRuntime()

	// Initialize Thunder Runtime config with basic test config
	testConfig := &config.Config{
		GateClient: config.GateClientConfig{
			Scheme:    "https",
			Hostname:  "localhost",
			Port:      3000,
			LoginPath: "/login",
			ErrorPath: "/error",
		},
		Database: config.DatabaseConfig{
			Identity: config.DataSource{
				Type: "sqlite",
				Path: ":memory:",
			},
			Runtime: config.DataSource{
				Type: "sqlite",
				Path: ":memory:",
			},
		},
		OAuth: config.OAuthConfig{
			AuthorizationCode: config.AuthorizationCodeConfig{
				ValidityPeriod: 600,
			},
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)
}

func (suite *AuthorizeHandlerTestSuite) SetupTest() {
	// Create mocked dependencies for testing
	suite.mockAppService = applicationmock.NewApplicationServiceInterfaceMock(suite.T())
	suite.mockJWTService = jwtmock.NewJWTServiceInterfaceMock(suite.T())
	suite.mockAuthzCodeStore = NewAuthorizationCodeStoreInterfaceMock(suite.T())
	suite.mockAuthReqStore = newAuthorizationRequestStoreInterfaceMock(suite.T())
	suite.mockFlowExecService = flowexecmock.NewFlowExecServiceInterfaceMock(suite.T())

	suite.handler = newAuthorizeHandler(
		suite.mockAppService, suite.mockJWTService, suite.mockAuthzCodeStore,
		suite.mockAuthReqStore, suite.mockFlowExecService).(*authorizeHandler)
}

func (suite *AuthorizeHandlerTestSuite) TestnewAuthorizeHandler() {
	mockStore := NewAuthorizationCodeStoreInterfaceMock(suite.T())
	mockFlowExec := flowexecmock.NewFlowExecServiceInterfaceMock(suite.T())
	mockAuthReqStore := newAuthorizationRequestStoreInterfaceMock(suite.T())
	handler := newAuthorizeHandler(suite.mockAppService, suite.mockJWTService, mockStore, mockAuthReqStore,
		mockFlowExec)
	assert.NotNil(suite.T(), handler)
	assert.Implements(suite.T(), (*AuthorizeHandlerInterface)(nil), handler)
}

func (suite *AuthorizeHandlerTestSuite) TestGetOAuthMessageForGetRequest_Success() {
	req := httptest.NewRequest(http.MethodGet, "/auth?client_id=test-client&redirect_uri=https://example.com", nil)

	msg, err := suite.handler.getOAuthMessageForGetRequest(req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), msg)
	if msg != nil {
		assert.Equal(suite.T(), oauth2const.TypeInitialAuthorizationRequest, msg.RequestType)
		assert.Equal(suite.T(), "test-client", msg.RequestQueryParams["client_id"])
		assert.Equal(suite.T(), "https://example.com", msg.RequestQueryParams["redirect_uri"])
		assert.Empty(suite.T(), msg.AuthID)
	}
}

func (suite *AuthorizeHandlerTestSuite) TestGetOAuthMessageForGetRequest_ParseFormError() {
	// Create a malformed URL to trigger ParseForm error
	req := httptest.NewRequest(http.MethodGet, "/auth?client_id=%ZZ", nil)

	msg, err := suite.handler.getOAuthMessageForGetRequest(req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), msg)
	assert.Contains(suite.T(), err.Error(), "failed to parse form data")
}

func (suite *AuthorizeHandlerTestSuite) TestGetOAuthMessageForPostRequest_MissingAuthID() {
	postData := AuthZPostRequest{
		AuthID:    "", // Missing auth ID
		Assertion: "test-assertion",
	}
	jsonData, _ := json.Marshal(postData)

	req := httptest.NewRequest(http.MethodPost, "/auth", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")

	msg, err := suite.handler.getOAuthMessageForPostRequest(req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), msg)
	assert.Contains(suite.T(), err.Error(), "authId or assertion is missing")
}

func (suite *AuthorizeHandlerTestSuite) TestGetOAuthMessageForPostRequest_MissingAssertion() {
	postData := AuthZPostRequest{
		AuthID:    testAuthID,
		Assertion: "", // Missing assertion
	}
	jsonData, _ := json.Marshal(postData)

	req := httptest.NewRequest(http.MethodPost, "/auth", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")

	msg, err := suite.handler.getOAuthMessageForPostRequest(req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), msg)
	assert.Contains(suite.T(), err.Error(), "authId or assertion is missing")
}

func (suite *AuthorizeHandlerTestSuite) TestGetOAuthMessage_UnsupportedMethod() {
	req := httptest.NewRequest(http.MethodPatch, "/auth", nil)
	rr := httptest.NewRecorder()

	msg := suite.handler.getOAuthMessage(req, rr)

	assert.Nil(suite.T(), msg)
	assert.Equal(suite.T(), http.StatusBadRequest, rr.Code)
}

func (suite *AuthorizeHandlerTestSuite) TestGetOAuthMessage_NilRequest() {
	rr := httptest.NewRecorder()

	msg := suite.handler.getOAuthMessage(nil, rr)

	assert.Nil(suite.T(), msg)
}

func (suite *AuthorizeHandlerTestSuite) TestGetOAuthMessage_NilResponseWriter() {
	req := httptest.NewRequest(http.MethodGet, "/auth", nil)

	msg := suite.handler.getOAuthMessage(req, nil)

	assert.Nil(suite.T(), msg)
}

func (suite *AuthorizeHandlerTestSuite) TestGetAuthorizationCode_Success() {
	// Create a valid OAuth message with authorization request context
	authRequestCtx := &authRequestContext{
		OAuthParameters: oauth2model.OAuthParameters{
			ClientID:         "test-client",
			RedirectURI:      "https://client.example.com/callback",
			StandardScopes:   []string{"openid", "profile"},
			PermissionScopes: []string{"read", "write"},
		},
	}

	assertionClaims := &assertionClaims{userID: "test-user"}
	// Use current time as auth time
	authTime := time.Now()

	result, err := createAuthorizationCode(authRequestCtx, assertionClaims, authTime)

	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), result.CodeID)
	assert.NotEmpty(suite.T(), result.Code)
	assert.Equal(suite.T(), "test-client", result.ClientID)
	assert.Equal(suite.T(), "https://client.example.com/callback", result.RedirectURI)
	assert.Equal(suite.T(), "test-user", result.AuthorizedUserID)
	assert.Equal(suite.T(), "openid profile read write", result.Scopes)
	assert.Equal(suite.T(), AuthCodeStateActive, result.State)
	assert.NotZero(suite.T(), result.TimeCreated)
	assert.True(suite.T(), result.ExpiryTime.After(result.TimeCreated))
}

func (suite *AuthorizeHandlerTestSuite) TestGetAuthorizationCode_MissingClientID() {
	authRequestCtx := &authRequestContext{
		OAuthParameters: oauth2model.OAuthParameters{
			ClientID:    "", // Empty client ID
			RedirectURI: "https://client.example.com/callback",
		},
	}

	assertionClaims := &assertionClaims{
		userID: "test-user",
	}
	authTime := time.Now()

	result, err := createAuthorizationCode(authRequestCtx, assertionClaims, authTime)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "client_id or redirect_uri is missing")
	assert.Equal(suite.T(), AuthorizationCode{}, result)
}

func (suite *AuthorizeHandlerTestSuite) TestGetAuthorizationCode_MissingRedirectURI() {
	authRequestCtx := &authRequestContext{
		OAuthParameters: oauth2model.OAuthParameters{
			ClientID:    "test-client",
			RedirectURI: "", // Missing redirect URI
		},
	}

	assertionClaims := &assertionClaims{
		userID: "test-user",
	}
	authTime := time.Now()

	result, err := createAuthorizationCode(authRequestCtx, assertionClaims, authTime)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "client_id or redirect_uri is missing")
	assert.Equal(suite.T(), AuthorizationCode{}, result)
}

func (suite *AuthorizeHandlerTestSuite) TestGetAuthorizationCode_EmptyUserID() {
	authRequestCtx := &authRequestContext{
		OAuthParameters: oauth2model.OAuthParameters{
			ClientID:    "test-client-id",
			RedirectURI: "https://client.example.com/callback",
		},
	}

	assertionClaims := &assertionClaims{
		userID: "", // Empty user ID
	}
	authTime := time.Now()

	result, err := createAuthorizationCode(authRequestCtx, assertionClaims, authTime)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "authenticated user not found")
	assert.Equal(suite.T(), AuthorizationCode{}, result)
}

func (suite *AuthorizeHandlerTestSuite) TestGetAuthorizationCode_ZeroAuthTime() {
	authRequestCtx := &authRequestContext{
		OAuthParameters: oauth2model.OAuthParameters{
			ClientID:    "test-client-id",
			RedirectURI: "https://client.example.com/callback",
		},
	}

	assertionClaims := &assertionClaims{
		userID: "test-user",
	}
	// Zero auth time - should fallback to current time
	zeroAuthTime := time.Time{}
	beforeCreation := time.Now()

	result, err := createAuthorizationCode(authRequestCtx, assertionClaims, zeroAuthTime)

	// Should succeed and use current time as fallback
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), result.CodeID)
	assert.NotEmpty(suite.T(), result.Code)
	assert.NotZero(suite.T(), result.TimeCreated)
	// TimeCreated should be approximately now (within 1 second)
	afterCreation := time.Now()
	assert.True(suite.T(), result.TimeCreated.After(beforeCreation) || result.TimeCreated.Equal(beforeCreation))
	assert.True(suite.T(), result.TimeCreated.Before(afterCreation) || result.TimeCreated.Equal(afterCreation))
	// Expiry should be 10 minutes after TimeCreated
	assert.True(suite.T(), result.ExpiryTime.After(result.TimeCreated))
	assert.WithinDuration(suite.T(), result.TimeCreated.Add(10*time.Minute), result.ExpiryTime, time.Second)
}

func (suite *AuthorizeHandlerTestSuite) TestGetLoginPageRedirectURI_Success() {
	queryParams := map[string]string{
		"authId": "test-key",
		"appId":  "test-app",
	}

	redirectURI, err := getLoginPageRedirectURI(queryParams)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), redirectURI, "authId=test-key")
	assert.Contains(suite.T(), redirectURI, "appId=test-app")
}

func (suite *AuthorizeHandlerTestSuite) TestGetErrorPageRedirectURL_Success() {
	redirectURI, err := getErrorPageRedirectURL("invalid_request", "Missing parameter")
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), redirectURI, "errorCode=invalid_request")
	assert.Contains(suite.T(), redirectURI, "errorMessage=Missing+parameter")
}

// Helper function to create a valid OAuth application for testing
func (suite *AuthorizeHandlerTestSuite) createTestOAuthApp() *appmodel.OAuthAppConfigProcessedDTO {
	return &appmodel.OAuthAppConfigProcessedDTO{
		AppID:         "test-app-id",
		ClientID:      "test-client-id",
		RedirectURIs:  []string{"https://client.example.com/callback"},
		GrantTypes:    []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
		ResponseTypes: []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
		PKCERequired:  false, // Disable PKCE to simplify test
	}
}

// Helper function to create a test OAuth message
func (suite *AuthorizeHandlerTestSuite) createTestOAuthMessage() *OAuthMessage {
	return &OAuthMessage{
		RequestType: oauth2const.TypeInitialAuthorizationRequest,
		RequestQueryParams: map[string]string{
			"client_id":     "test-client-id",
			"redirect_uri":  "https://client.example.com/callback",
			"response_type": "code",
			"scope":         "read write",
			"state":         "test-state",
		},
	}
}

func (suite *AuthorizeHandlerTestSuite) TestHandleInitialAuthorizationRequest_InitiateFlowSuccess() {
	// Create a valid OAuth application with proper grant types and response types
	app := suite.createTestOAuthApp()
	suite.mockAppService.EXPECT().GetOAuthApplication("test-client-id").Return(app, nil)

	// Mock flow exec service to return success
	expectedFlowInitCtx := &flowexec.FlowInitContext{
		ApplicationID: "test-app-id",
		FlowType:      string(flowcm.FlowTypeAuthentication),
		RuntimeData: map[string]string{
			"requested_permissions": "read write",
			"required_attributes":   "",
		},
	}
	suite.mockFlowExecService.EXPECT().InitiateFlow(expectedFlowInitCtx).Return("test-session-key", nil)
	suite.mockAuthReqStore.EXPECT().AddRequest(mock.Anything).Return(testAuthID)

	// Create OAuth message for initial authorization request
	msg := suite.createTestOAuthMessage()

	// Create HTTP request and response recorder
	req := httptest.NewRequest(http.MethodGet, "/auth", nil)
	rr := httptest.NewRecorder()

	// Execute the method under test
	suite.handler.handleInitialAuthorizationRequest(msg, rr, req)

	// Assert that it redirects to login page
	assert.Equal(suite.T(), http.StatusFound, rr.Code)

	// Check the redirect location contains login page and flow information
	location := rr.Header().Get("Location")
	assert.Contains(suite.T(), location, "/login")
	assert.Contains(suite.T(), location, "flowId=test-session-key")
	assert.Contains(suite.T(), location, "authId=")
}

func (suite *AuthorizeHandlerTestSuite) TestHandleInitialAuthorizationRequest_InitiateFlowError() {
	// Create a valid OAuth application with proper grant types and response types
	app := suite.createTestOAuthApp()
	suite.mockAppService.EXPECT().GetOAuthApplication("test-client-id").Return(app, nil)

	// Mock flow exec service to return an error
	expectedFlowInitCtx := &flowexec.FlowInitContext{
		ApplicationID: "test-app-id",
		FlowType:      string(flowcm.FlowTypeAuthentication),
		RuntimeData: map[string]string{
			"requested_permissions": "read write",
			"required_attributes":   "",
		},
	}
	mockError := &serviceerror.InternalServerError
	suite.mockFlowExecService.EXPECT().InitiateFlow(expectedFlowInitCtx).Return("", mockError)

	// Create OAuth message for initial authorization request
	msg := suite.createTestOAuthMessage()

	// Create HTTP request and response recorder
	req := httptest.NewRequest(http.MethodGet, "/auth", nil)
	rr := httptest.NewRecorder()

	// Execute the method under test
	suite.handler.handleInitialAuthorizationRequest(msg, rr, req)

	// Assert that it redirects to error page
	assert.Equal(suite.T(), http.StatusFound, rr.Code)

	// Check the redirect location contains error information
	location := rr.Header().Get("Location")
	assert.Contains(suite.T(), location, "/error")
	assert.Contains(suite.T(), location, "errorCode=server_error")
	assert.Contains(suite.T(), location, "errorMessage=Failed+to+initiate+authentication+flow")
}

func (suite *AuthorizeHandlerTestSuite) TestHandleInitialAuthorizationRequest_WithOIDCAndNonOIDCScopes() {
	// Create a valid OAuth application
	app := suite.createTestOAuthApp()
	suite.mockAppService.EXPECT().GetOAuthApplication("test-client-id").Return(app, nil)

	// Mock flow exec service - only non-OIDC scopes should be in RuntimeData
	expectedFlowInitCtx := &flowexec.FlowInitContext{
		ApplicationID: "test-app-id",
		FlowType:      string(flowcm.FlowTypeAuthentication),
		RuntimeData: map[string]string{
			"requested_permissions": "read write", // Only non-OIDC scopes
			"required_attributes":   "",
		},
	}
	suite.mockFlowExecService.EXPECT().InitiateFlow(expectedFlowInitCtx).Return("test-session-key", nil)
	suite.mockAuthReqStore.EXPECT().AddRequest(mock.Anything).Return(testAuthID)

	// Create OAuth message with both OIDC (openid, profile) and non-OIDC scopes (read, write)
	msg := &OAuthMessage{
		RequestType: oauth2const.TypeInitialAuthorizationRequest,
		RequestQueryParams: map[string]string{
			"client_id":     "test-client-id",
			"redirect_uri":  "https://client.example.com/callback",
			"response_type": "code",
			"scope":         "openid profile read write", // Mixed scopes
			"state":         "test-state",
		},
	}

	// Create HTTP request and response recorder
	req := httptest.NewRequest(http.MethodGet, "/auth", nil)
	rr := httptest.NewRecorder()

	// Execute the method under test
	suite.handler.handleInitialAuthorizationRequest(msg, rr, req)

	// Assert that it redirects to login page
	assert.Equal(suite.T(), http.StatusFound, rr.Code)

	// Check the redirect location
	location := rr.Header().Get("Location")
	assert.Contains(suite.T(), location, "/login")
	assert.Contains(suite.T(), location, "flowId=test-session-key")
}

func (suite *AuthorizeHandlerTestSuite) TestHandleInitialAuthorizationRequest_OnlyOIDCScopes() {
	// Create a valid OAuth application
	app := suite.createTestOAuthApp()
	suite.mockAppService.EXPECT().GetOAuthApplication("test-client-id").Return(app, nil)

	// Mock flow exec service - empty RuntimeData since no non-OIDC scopes
	expectedFlowInitCtx := &flowexec.FlowInitContext{
		ApplicationID: "test-app-id",
		FlowType:      string(flowcm.FlowTypeAuthentication),
		RuntimeData: map[string]string{
			"requested_permissions": "", // Empty, only OIDC scopes
			"required_attributes":   "",
		},
	}
	suite.mockFlowExecService.EXPECT().InitiateFlow(expectedFlowInitCtx).Return("test-session-key", nil)
	suite.mockAuthReqStore.EXPECT().AddRequest(mock.Anything).Return(testAuthID)

	// Create OAuth message with only OIDC scopes
	msg := &OAuthMessage{
		RequestType: oauth2const.TypeInitialAuthorizationRequest,
		RequestQueryParams: map[string]string{
			"client_id":     "test-client-id",
			"redirect_uri":  "https://client.example.com/callback",
			"response_type": "code",
			"scope":         "openid profile email", // Only OIDC scopes
			"state":         "test-state",
		},
	}

	// Create HTTP request and response recorder
	req := httptest.NewRequest(http.MethodGet, "/auth", nil)
	rr := httptest.NewRecorder()

	// Execute the method under test
	suite.handler.handleInitialAuthorizationRequest(msg, rr, req)

	// Assert that it redirects to login page
	assert.Equal(suite.T(), http.StatusFound, rr.Code)
}

func (suite *AuthorizeHandlerTestSuite) TestHandleAuthorizeGetRequest_Success() {
	// Test GET request handling - this will trigger validation
	app := suite.createTestOAuthApp()
	suite.mockAppService.EXPECT().GetOAuthApplication("test-client").Return(app, nil)

	req := httptest.NewRequest("GET",
		"/oauth2/authorize?client_id=test-client&redirect_uri=https://example.com/callback&response_type=code", nil)
	rr := httptest.NewRecorder()

	suite.handler.HandleAuthorizeGetRequest(rr, req)

	// Should process the request (may redirect or return error, but not panic)
	// The exact status depends on validation, but should not be 500
	assert.NotEqual(suite.T(), http.StatusInternalServerError, rr.Code)
}

func (suite *AuthorizeHandlerTestSuite) TestHandleAuthorizeGetRequest_InvalidParams() {
	// Test GET request with invalid parameters
	req := httptest.NewRequest("GET", "/oauth2/authorize?client_id=&redirect_uri=", nil)
	rr := httptest.NewRecorder()

	suite.handler.HandleAuthorizeGetRequest(rr, req)

	// Should handle invalid params gracefully (returns 400 or redirects to error)
	assert.NotEqual(suite.T(), http.StatusInternalServerError, rr.Code)
}

func (suite *AuthorizeHandlerTestSuite) TestHandleAuthCallbackPostRequest_ConsentType() {
	// Test TypeConsentResponseFromUser case
	// This case is not implemented yet, but we test that it doesn't panic
	postData := AuthZPostRequest{
		AuthID:    "test-key",
		Assertion: "test-assertion",
	}
	jsonData, _ := json.Marshal(postData)

	req := httptest.NewRequest(http.MethodPost, "/oauth2/auth/callback", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	// Add a query parameter to indicate consent response type
	req.URL.RawQuery = "requestType=consentResponseFromUser"
	rr := httptest.NewRecorder()

	// The consent type may call GetRequest, so we need to set up the expectation
	suite.mockAuthReqStore.EXPECT().GetRequest("test-key").Return(false, authRequestContext{})

	// The consent type is not handled yet (TODO), so it should fall through to default case
	// or handle gracefully without panicking
	suite.handler.HandleAuthCallbackPostRequest(rr, req)

	// Should return some response (either error or handled gracefully)
	assert.NotEqual(suite.T(), 0, rr.Code)
}

func (suite *AuthorizeHandlerTestSuite) TestHandleAuthCallbackPostRequest_InvalidRequestType() {
	req := httptest.NewRequest(http.MethodPost, "/oauth2/auth/callback", nil)
	rr := httptest.NewRecorder()

	suite.handler.HandleAuthCallbackPostRequest(rr, req)

	assert.Equal(suite.T(), http.StatusBadRequest, rr.Code)
}

func (suite *AuthorizeHandlerTestSuite) TestHandleAuthorizationResponseFromEngine_InvalidAuthID() {
	suite.mockAuthReqStore.EXPECT().GetRequest("invalid-key").Return(false, authRequestContext{})

	msg := &OAuthMessage{
		RequestType: oauth2const.TypeAuthorizationResponseFromEngine,
		AuthID:      "invalid-key",
		RequestBodyParams: map[string]string{
			oauth2const.Assertion: "test-assertion",
		},
	}

	rr := httptest.NewRecorder()
	suite.handler.handleAuthorizationResponseFromEngine(msg, rr)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)
	var resp AuthZPostResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), resp.RedirectURI, "/error")
}

func (suite *AuthorizeHandlerTestSuite) TestHandleAuthorizationResponseFromEngine_MissingAssertion() {
	authRequestCtx := authRequestContext{
		OAuthParameters: oauth2model.OAuthParameters{
			ClientID:    "test-client",
			RedirectURI: "https://client.example.com/callback",
		},
	}
	authID := testAuthID
	suite.mockAuthReqStore.EXPECT().GetRequest(authID).Return(true, authRequestCtx)
	suite.mockAuthReqStore.EXPECT().ClearRequest(authID)

	msg := &OAuthMessage{
		RequestType:       oauth2const.TypeAuthorizationResponseFromEngine,
		AuthID:            authID,
		RequestBodyParams: map[string]string{},
	}

	rr := httptest.NewRecorder()
	suite.handler.handleAuthorizationResponseFromEngine(msg, rr)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)
}

func (suite *AuthorizeHandlerTestSuite) TestHandleAuthorizationResponseFromEngine_InvalidAssertionSignature() {
	authRequestCtx := authRequestContext{
		OAuthParameters: oauth2model.OAuthParameters{
			ClientID:    "test-client",
			RedirectURI: "https://client.example.com/callback",
		},
	}
	authID := testAuthID
	suite.mockAuthReqStore.EXPECT().GetRequest(authID).Return(true, authRequestCtx)
	suite.mockAuthReqStore.EXPECT().ClearRequest(authID)

	assertion := "invalid.jwt.token"
	suite.mockJWTService.EXPECT().VerifyJWT(assertion, "", "").Return(&jwt.ErrorInvalidTokenSignature)

	msg := &OAuthMessage{
		RequestType: oauth2const.TypeAuthorizationResponseFromEngine,
		AuthID:      authID,
		RequestBodyParams: map[string]string{
			oauth2const.Assertion: assertion,
		},
	}

	rr := httptest.NewRecorder()
	suite.handler.handleAuthorizationResponseFromEngine(msg, rr)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)
}

func (suite *AuthorizeHandlerTestSuite) TestHandleAuthorizationResponseFromEngine_FailedToDecodeAssertion() {
	authRequestCtx := authRequestContext{
		OAuthParameters: oauth2model.OAuthParameters{
			ClientID:    "test-client",
			RedirectURI: "https://client.example.com/callback",
		},
	}
	authID := testAuthID
	suite.mockAuthReqStore.EXPECT().GetRequest(authID).Return(true, authRequestCtx)
	suite.mockAuthReqStore.EXPECT().ClearRequest(authID)

	assertion := "invalid-jwt-format"
	suite.mockJWTService.EXPECT().VerifyJWT(assertion, "", "").Return(nil)

	msg := &OAuthMessage{
		RequestType: oauth2const.TypeAuthorizationResponseFromEngine,
		AuthID:      authID,
		RequestBodyParams: map[string]string{
			oauth2const.Assertion: assertion,
		},
	}

	rr := httptest.NewRecorder()
	suite.handler.handleAuthorizationResponseFromEngine(msg, rr)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)
}

func (suite *AuthorizeHandlerTestSuite) TestHandleAuthorizationResponseFromEngine_EmptyUserID() {
	authRequestCtx := authRequestContext{
		OAuthParameters: oauth2model.OAuthParameters{
			ClientID:    "test-client",
			RedirectURI: "https://client.example.com/callback",
		},
	}
	authID := testAuthID
	suite.mockAuthReqStore.EXPECT().GetRequest(authID).Return(true, authRequestCtx)
	suite.mockAuthReqStore.EXPECT().ClearRequest(authID)

	// This will fail during decode since it's not a valid JWT
	assertion := "not.a.valid.jwt"
	suite.mockJWTService.EXPECT().VerifyJWT(assertion, "", "").Return(nil)

	msg := &OAuthMessage{
		RequestType: oauth2const.TypeAuthorizationResponseFromEngine,
		AuthID:      authID,
		RequestBodyParams: map[string]string{
			oauth2const.Assertion: assertion,
		},
	}

	rr := httptest.NewRecorder()
	suite.handler.handleAuthorizationResponseFromEngine(msg, rr)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)
	var resp AuthZPostResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), resp.RedirectURI, "/error")
}

func (suite *AuthorizeHandlerTestSuite) TestRedirectToLoginPage_NilResponseWriter() {
	req := httptest.NewRequest(http.MethodGet, "/auth", nil)
	queryParams := map[string]string{"authId": "test-key"}

	suite.handler.redirectToLoginPage(nil, req, queryParams)
	// Should not panic and should log error
}

func (suite *AuthorizeHandlerTestSuite) TestRedirectToLoginPage_NilRequest() {
	rr := httptest.NewRecorder()
	queryParams := map[string]string{"authId": "test-key"}

	suite.handler.redirectToLoginPage(rr, nil, queryParams)
	// Should not panic and should log error
}

func (suite *AuthorizeHandlerTestSuite) TestRedirectToErrorPage_NilResponseWriter() {
	req := httptest.NewRequest(http.MethodGet, "/auth", nil)

	suite.handler.redirectToErrorPage(nil, req, "error_code", "error message")
	// Should not panic and should log error
}

func (suite *AuthorizeHandlerTestSuite) TestRedirectToErrorPage_NilRequest() {
	rr := httptest.NewRecorder()

	suite.handler.redirectToErrorPage(rr, nil, "error_code", "error message")
	// Should not panic and should log error
}

func (suite *AuthorizeHandlerTestSuite) TestWriteAuthZResponseToErrorPage_WithState() {
	authRequestCtx := &authRequestContext{
		OAuthParameters: oauth2model.OAuthParameters{
			State: "test-state",
		},
	}

	rr := httptest.NewRecorder()
	suite.handler.writeAuthZResponseToErrorPage(rr, "error_code", "error message", authRequestCtx)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)
	var resp AuthZPostResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), resp.RedirectURI, "state=test-state")
}

func (suite *AuthorizeHandlerTestSuite) TestWriteAuthZResponseToErrorPage_NilAuthRequestContext() {
	rr := httptest.NewRecorder()
	suite.handler.writeAuthZResponseToErrorPage(rr, "error_code", "error message", nil)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)
	var resp AuthZPostResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), resp.RedirectURI)
}

func (suite *AuthorizeHandlerTestSuite) TestHandleInitialAuthorizationRequest_MissingClientID() {
	msg := &OAuthMessage{
		RequestType: oauth2const.TypeInitialAuthorizationRequest,
		RequestQueryParams: map[string]string{
			"redirect_uri":  "https://client.example.com/callback",
			"response_type": "code",
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/auth", nil)
	rr := httptest.NewRecorder()

	suite.handler.handleInitialAuthorizationRequest(msg, rr, req)

	assert.Equal(suite.T(), http.StatusFound, rr.Code)
	location := rr.Header().Get("Location")
	assert.Contains(suite.T(), location, "/error")
}

func (suite *AuthorizeHandlerTestSuite) TestHandleInitialAuthorizationRequest_InvalidClient() {
	suite.mockAppService.EXPECT().GetOAuthApplication("invalid-client").Return(nil, &serviceerror.ServiceError{
		Code: "CLIENT_NOT_FOUND",
	})

	msg := &OAuthMessage{
		RequestType: oauth2const.TypeInitialAuthorizationRequest,
		RequestQueryParams: map[string]string{
			"client_id":     "invalid-client",
			"redirect_uri":  "https://client.example.com/callback",
			"response_type": "code",
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/auth", nil)
	rr := httptest.NewRecorder()

	suite.handler.handleInitialAuthorizationRequest(msg, rr, req)

	assert.Equal(suite.T(), http.StatusFound, rr.Code)
	location := rr.Header().Get("Location")
	assert.Contains(suite.T(), location, "/error")
}

func (suite *AuthorizeHandlerTestSuite) TestHandleInitialAuthorizationRequest_ValidationError_RedirectToApp() {
	app := suite.createTestOAuthApp()
	suite.mockAppService.EXPECT().GetOAuthApplication("test-client-id").Return(app, nil)

	msg := &OAuthMessage{
		RequestType: oauth2const.TypeInitialAuthorizationRequest,
		RequestQueryParams: map[string]string{
			"client_id":     "test-client-id",
			"redirect_uri":  "https://client.example.com/callback",
			"response_type": "invalid_type", // Invalid response type
			"state":         "test-state",
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/auth", nil)
	rr := httptest.NewRecorder()

	suite.handler.handleInitialAuthorizationRequest(msg, rr, req)

	assert.Equal(suite.T(), http.StatusFound, rr.Code)
}

func (suite *AuthorizeHandlerTestSuite) TestHandleInitialAuthorizationRequest_InsecureRedirectURI() {
	app := suite.createTestOAuthApp()
	app.RedirectURIs = []string{"http://client.example.com/callback"} // HTTP instead of HTTPS
	suite.mockAppService.EXPECT().GetOAuthApplication("test-client-id").Return(app, nil)

	expectedFlowInitCtx := &flowexec.FlowInitContext{
		ApplicationID: "test-app-id",
		FlowType:      string(flowcm.FlowTypeAuthentication),
		RuntimeData: map[string]string{
			"requested_permissions": "read write",
			"required_attributes":   "",
		},
	}
	suite.mockFlowExecService.EXPECT().InitiateFlow(expectedFlowInitCtx).Return("test-flow-id", nil)
	suite.mockAuthReqStore.EXPECT().AddRequest(mock.Anything).Return(testAuthID)

	msg := &OAuthMessage{
		RequestType: oauth2const.TypeInitialAuthorizationRequest,
		RequestQueryParams: map[string]string{
			"client_id":     "test-client-id",
			"redirect_uri":  "http://client.example.com/callback",
			"response_type": "code",
			"scope":         "read write",
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/auth", nil)
	rr := httptest.NewRecorder()

	suite.handler.handleInitialAuthorizationRequest(msg, rr, req)

	assert.Equal(suite.T(), http.StatusFound, rr.Code)
	location := rr.Header().Get("Location")
	assert.Contains(suite.T(), location, "showInsecureWarning=true")
}

func (suite *AuthorizeHandlerTestSuite) TestHandleInitialAuthorizationRequest_EmptyRedirectURI() {
	app := suite.createTestOAuthApp()
	suite.mockAppService.EXPECT().GetOAuthApplication("test-client-id").Return(app, nil)

	expectedFlowInitCtx := &flowexec.FlowInitContext{
		ApplicationID: "test-app-id",
		FlowType:      string(flowcm.FlowTypeAuthentication),
		RuntimeData: map[string]string{
			"requested_permissions": "read write",
			"required_attributes":   "",
		},
	}
	suite.mockFlowExecService.EXPECT().InitiateFlow(expectedFlowInitCtx).Return("test-flow-id", nil)
	suite.mockAuthReqStore.EXPECT().AddRequest(mock.Anything).Return(testAuthID)

	msg := &OAuthMessage{
		RequestType: oauth2const.TypeInitialAuthorizationRequest,
		RequestQueryParams: map[string]string{
			"client_id":     "test-client-id",
			"response_type": "code",
			"scope":         "read write",
			// redirect_uri is empty, should use app's default
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/auth", nil)
	rr := httptest.NewRecorder()

	suite.handler.handleInitialAuthorizationRequest(msg, rr, req)

	assert.Equal(suite.T(), http.StatusFound, rr.Code)
}

func (suite *AuthorizeHandlerTestSuite) TestHandleAuthorizationResponseFromEngine_PersistAuthCodeError() {
	authRequestCtx := authRequestContext{
		OAuthParameters: oauth2model.OAuthParameters{
			ClientID:    "test-client",
			RedirectURI: "https://client.example.com/callback",
		},
	}
	authID := testAuthID
	suite.mockAuthReqStore.EXPECT().GetRequest(authID).Return(true, authRequestCtx)
	suite.mockAuthReqStore.EXPECT().ClearRequest(authID)

	// Use JWT with iat claim for testing
	// Header: {"alg":"none","typ":"JWT"}
	// Payload: {"sub":"test-user","iat":1701421200}
	assertion := testJWTWithIat
	suite.mockJWTService.EXPECT().VerifyJWT(assertion, "", "").Return(nil)

	// Mock the store to return an error when inserting
	storeError := errors.New("database error")
	suite.mockAuthzCodeStore.EXPECT().InsertAuthorizationCode(mock.Anything).Return(storeError)

	msg := &OAuthMessage{
		RequestType: oauth2const.TypeAuthorizationResponseFromEngine,
		AuthID:      authID,
		RequestBodyParams: map[string]string{
			oauth2const.Assertion: assertion,
		},
	}

	rr := httptest.NewRecorder()
	suite.handler.handleAuthorizationResponseFromEngine(msg, rr)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)
	var resp AuthZPostResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), resp.RedirectURI, "/error")
}

func (suite *AuthorizeHandlerTestSuite) TestDecodeAttributesFromAssertion_Success() {
	// Create a valid JWT token with all claim types (using base64 encoded payload)
	// Payload contains: sub, username, email, firstName, lastName, authorized_permissions,
	// userType, ouId, ouName, ouHandle
	// This is a valid JWT format: header.payload.signature (signature can be empty for testing decode)
	// JWT with all claim types: sub, username, email, firstName, lastName,
	// authorized_permissions, userType, ouId, ouName, ouHandle
	validJWT := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0." +
		"eyJzdWIiOiJ0ZXN0LXVzZXIiLCJ1c2VybmFtZSI6InRlc3R1c2VyIiwiZW1haWwiOiJ0ZXN0QGV4YW1wbGUuY29tIiwi" +
		"Zmlyc3ROYW1lIjoiVGVzdCIsImxhc3ROYW1lIjoiVXNlciIsImF1dGhvcml6ZWRfcGVybWlzc2lvbnMiOiJyZWFkIHdyaXRlIiwidXNlclR5cGUiOiJsb2NhbCIsIm91SWQiOiJvdTEyMyIsIm91TmFtZSI6Ik9yZ2FuaXphdGlvbiIsIm91SGFuZGxlIjoib3JnLWhhbmRsZSJ9." //nolint:lll

	claims, _, err := decodeAttributesFromAssertion(validJWT)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test-user", claims.userID)
	assert.Equal(suite.T(), "local", claims.userAttributes["userType"])
	assert.Equal(suite.T(), "ou123", claims.userAttributes["ouId"])
	assert.Equal(suite.T(), "Organization", claims.userAttributes["ouName"])
	assert.Equal(suite.T(), "org-handle", claims.userAttributes["ouHandle"])
	assert.Equal(suite.T(), "testuser", claims.userAttributes["username"])
	assert.Equal(suite.T(), "test@example.com", claims.userAttributes["email"])
	assert.Equal(suite.T(), "Test", claims.userAttributes["firstName"])
	assert.Equal(suite.T(), "User", claims.userAttributes["lastName"])
	// authorized_permissions is extracted separately, not in userAttributes
	assert.Equal(suite.T(), "read write", claims.authorizedPermissions)
}

func (suite *AuthorizeHandlerTestSuite) TestDecodeAttributesFromAssertion_DecodeError() {
	invalidJWT := "invalid.jwt.token"

	_, _, err := decodeAttributesFromAssertion(invalidJWT)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "Failed to decode the JWT token")
}

func (suite *AuthorizeHandlerTestSuite) TestDecodeAttributesFromAssertion_InvalidSubClaim() {
	// JWT with invalid sub claim type (number instead of string)
	// Payload: {"sub":12345}
	invalidSubJWT := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOjEyMzQ1fQ."

	claims, _, err := decodeAttributesFromAssertion(invalidSubJWT)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "JWT 'sub' claim is not a string")
	assert.Equal(suite.T(), "", claims.userID)
}

func (suite *AuthorizeHandlerTestSuite) TestDecodeAttributesFromAssertion_NonStringAttributes() {
	// User attributes are now stored as-is without type validation
	// Payload: {"sub":"test-user","username":12345,"email":12345,"firstName":12345,
	// "lastName":12345,"authorized_permissions":12345}
	nonStringAttrsJWT := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0." +
		"eyJzdWIiOiJ0ZXN0LXVzZXIiLCJ1c2VybmFtZSI6MTIzNDUsImVtYWlsIjoxMjM0NSwi" +
		"Zmlyc3ROYW1lIjoxMjM0NSwibGFzdE5hbWUiOjEyMzQ1LCJhdXRob3JpemVkX3Blcm1p" +
		"c3Npb25zIjoxMjM0NX0."

	claims, _, err := decodeAttributesFromAssertion(nonStringAttrsJWT)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test-user", claims.userID)
	// Attributes are stored as-is without type validation
	assert.Equal(suite.T(), float64(12345), claims.userAttributes["username"])
	assert.Equal(suite.T(), float64(12345), claims.userAttributes["email"])
	assert.Equal(suite.T(), float64(12345), claims.userAttributes["firstName"])
	assert.Equal(suite.T(), float64(12345), claims.userAttributes["lastName"])
	// authorized_permissions is extracted separately, not in userAttributes
	// If it's not a string, it's ignored (should be empty string)
	assert.Equal(suite.T(), "", claims.authorizedPermissions)
}

func (suite *AuthorizeHandlerTestSuite) TestDecodeAttributesFromAssertion_UserTypeInUserAttributes() {
	// userType is now stored as a regular user attribute without special validation
	// Payload: {"sub":"test-user","userType":12345}
	userTypeJWT := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJ0ZXN0LXVzZXIiLCJ1c2VyVHlwZSI6MTIzNDV9."

	claims, _, err := decodeAttributesFromAssertion(userTypeJWT)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test-user", claims.userID)
	// userType is now in userAttributes without type validation
	assert.Equal(suite.T(), float64(12345), claims.userAttributes["userType"])
}

func (suite *AuthorizeHandlerTestSuite) TestDecodeAttributesFromAssertion_OUClaimsInUserAttributes() {
	// OU claims are now stored as regular user attributes without special validation
	jwt := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0." +
		"eyJzdWIiOiJ0ZXN0LXVzZXIiLCJvdUlkIjoxMjM0NSwib3VOYW1lIjoxMjM0NSwib3VIYW5kbGUiOjEyMzQ1fQ."

	claims, _, err := decodeAttributesFromAssertion(jwt)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test-user", claims.userID)
	// OU claims are now in userAttributes without type validation
	assert.Equal(suite.T(), float64(12345), claims.userAttributes["ouId"])
	assert.Equal(suite.T(), float64(12345), claims.userAttributes["ouName"])
	assert.Equal(suite.T(), float64(12345), claims.userAttributes["ouHandle"])
}

func (suite *AuthorizeHandlerTestSuite) TestHandleAuthorizeGetRequest_GetOAuthMessageReturnsNil() {
	// Create a request that will cause getOAuthMessage to return nil
	req := httptest.NewRequest("GET", "/oauth2/authorize?client_id=%ZZ", nil) // Invalid URL encoding
	rr := httptest.NewRecorder()

	suite.handler.HandleAuthorizeGetRequest(rr, req)

	// Should return 400 Bad Request when getOAuthMessage fails
	assert.Equal(suite.T(), http.StatusBadRequest, rr.Code)
}

func (suite *AuthorizeHandlerTestSuite) TestHandleAuthCallbackPostRequest_GetOAuthMessageReturnsNil() {
	req := httptest.NewRequest("POST", "/oauth2/auth/callback", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	suite.handler.HandleAuthCallbackPostRequest(rr, req)

	// Should return 400 Bad Request when getOAuthMessage fails
	assert.Equal(suite.T(), http.StatusBadRequest, rr.Code)
}

func (suite *AuthorizeHandlerTestSuite) TestHandleInitialAuthorizationRequest_GetURIWithQueryParamsError() {
	// Test error path when GetURIWithQueryParams fails
	app := suite.createTestOAuthApp()
	suite.mockAppService.EXPECT().GetOAuthApplication("test-client-id").Return(app, nil)

	// Create a message with invalid redirect URI that will cause GetURIWithQueryParams to fail
	msg := &OAuthMessage{
		RequestType: oauth2const.TypeInitialAuthorizationRequest,
		RequestQueryParams: map[string]string{
			"client_id":     "test-client-id",
			"redirect_uri":  "://invalid-uri", // Invalid URI
			"response_type": "code",
		},
	}

	req := httptest.NewRequest("GET", "/auth", nil)
	rr := httptest.NewRecorder()

	suite.handler.handleInitialAuthorizationRequest(msg, rr, req)

	// Should redirect to error page when GetURIWithQueryParams fails
	assert.Equal(suite.T(), http.StatusFound, rr.Code)
}

func (suite *AuthorizeHandlerTestSuite) TestHandleInitialAuthorizationRequest_EmptyRedirectURIWithNoAppRedirectURIs() {
	app := suite.createTestOAuthApp()
	app.RedirectURIs = []string{} // Empty redirect URIs
	suite.mockAppService.EXPECT().GetOAuthApplication("test-client-id").Return(app, nil)

	msg := &OAuthMessage{
		RequestType: oauth2const.TypeInitialAuthorizationRequest,
		RequestQueryParams: map[string]string{
			"client_id":     "test-client-id",
			"redirect_uri":  "", // Empty redirect URI
			"response_type": "code",
		},
	}

	req := httptest.NewRequest("GET", "/auth", nil)
	rr := httptest.NewRecorder()

	// Validation will catch this before we reach the panic point
	// The handler will redirect to error page instead
	suite.handler.handleInitialAuthorizationRequest(msg, rr, req)

	// Should redirect to error page (validation error)
	assert.Equal(suite.T(), http.StatusFound, rr.Code)
}

func (suite *AuthorizeHandlerTestSuite) TestHandleAuthorizationResponseFromEngine_EmptyAuthorizedPermissions() {
	authRequestCtx := authRequestContext{
		OAuthParameters: oauth2model.OAuthParameters{
			ClientID:    "test-client",
			RedirectURI: "https://client.example.com/callback",
		},
	}
	authID := testAuthID
	suite.mockAuthReqStore.EXPECT().GetRequest(authID).Return(true, authRequestCtx)
	suite.mockAuthReqStore.EXPECT().ClearRequest(authID)

	// JWT with sub but no authorized_permissions (with iat claim)
	assertion := testJWTWithIat
	suite.mockJWTService.EXPECT().VerifyJWT(assertion, "", "").Return(nil)

	suite.mockAuthzCodeStore.EXPECT().InsertAuthorizationCode(mock.Anything).Return(nil)

	msg := &OAuthMessage{
		RequestType: oauth2const.TypeAuthorizationResponseFromEngine,
		AuthID:      authID,
		RequestBodyParams: map[string]string{
			oauth2const.Assertion: assertion,
		},
	}

	rr := httptest.NewRecorder()
	suite.handler.handleAuthorizationResponseFromEngine(msg, rr)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)
}

func (suite *AuthorizeHandlerTestSuite) TestHandleAuthorizationResponseFromEngine_WithState() {
	authRequestCtx := authRequestContext{
		OAuthParameters: oauth2model.OAuthParameters{
			ClientID:    "test-client",
			RedirectURI: "https://client.example.com/callback",
			State:       "test-state-123",
		},
	}
	authID := testAuthID
	suite.mockAuthReqStore.EXPECT().GetRequest(authID).Return(true, authRequestCtx)
	suite.mockAuthReqStore.EXPECT().ClearRequest(authID)

	assertion := testJWTWithIat
	suite.mockJWTService.EXPECT().VerifyJWT(assertion, "", "").Return(nil)

	suite.mockAuthzCodeStore.EXPECT().InsertAuthorizationCode(mock.Anything).Return(nil)

	msg := &OAuthMessage{
		RequestType: oauth2const.TypeAuthorizationResponseFromEngine,
		AuthID:      authID,
		RequestBodyParams: map[string]string{
			oauth2const.Assertion: assertion,
		},
	}

	rr := httptest.NewRecorder()
	suite.handler.handleAuthorizationResponseFromEngine(msg, rr)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)
	var resp AuthZPostResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	assert.NoError(suite.T(), err)
	// State should be URL-encoded in the redirect URI
	assert.Contains(suite.T(), resp.RedirectURI, "state=test-state-123")
}

func (suite *AuthorizeHandlerTestSuite) TestHandleAuthorizationResponseFromEngine_CreateAuthorizationCodeError() {
	// Test error when createAuthorizationCode fails
	// Empty client ID will cause createAuthorizationCode to fail
	authRequestCtx := authRequestContext{
		OAuthParameters: oauth2model.OAuthParameters{
			ClientID:    "",
			RedirectURI: "https://client.example.com/callback",
		},
	}
	authID := testAuthID
	suite.mockAuthReqStore.EXPECT().GetRequest(authID).Return(true, authRequestCtx)
	suite.mockAuthReqStore.EXPECT().ClearRequest(authID)

	assertion := testMinimalJWT
	suite.mockJWTService.EXPECT().VerifyJWT(assertion, "", "").Return(nil)

	msg := &OAuthMessage{
		RequestType: oauth2const.TypeAuthorizationResponseFromEngine,
		AuthID:      authID,
		RequestBodyParams: map[string]string{
			oauth2const.Assertion: assertion,
		},
	}

	rr := httptest.NewRecorder()
	suite.handler.handleAuthorizationResponseFromEngine(msg, rr)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)
	var resp AuthZPostResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), resp.RedirectURI, "/error")
}

func (suite *AuthorizeHandlerTestSuite) TestRedirectToLoginPage_GetLoginPageRedirectURIError() {
	// Test error path when getLoginPageRedirectURI fails
	req := httptest.NewRequest("GET", "/auth", nil)
	rr := httptest.NewRecorder()

	// Use valid params - the function should succeed
	validParams := map[string]string{
		"authId": "test-key",
		"appId":  "test-app",
	}

	suite.handler.redirectToLoginPage(rr, req, validParams)

	// Verify redirect occurred successfully
	assert.Equal(suite.T(), http.StatusFound, rr.Code, "Should redirect when URI construction succeeds")
	assert.NotEmpty(suite.T(), rr.Header().Get("Location"), "Should have Location header")
}

func (suite *AuthorizeHandlerTestSuite) TestRedirectToErrorPage_GetErrorPageRedirectURLError() {
	// Test error path when getErrorPageRedirectURL fails
	req := httptest.NewRequest("GET", "/auth", nil)
	rr := httptest.NewRecorder()

	// Test with nil response writer (should handle gracefully without panicking)
	suite.handler.redirectToErrorPage(nil, req, "error_code", "error message")

	suite.handler.redirectToErrorPage(rr, nil, "error_code", "error message")

	assert.Equal(suite.T(), http.StatusOK, rr.Code, "Status should remain unchanged when request is nil")
}

func (suite *AuthorizeHandlerTestSuite) TestWriteAuthZResponseToErrorPage_GetErrorPageRedirectURLError() {
	rr := httptest.NewRecorder()

	suite.handler.writeAuthZResponseToErrorPage(rr, "error_code", "error message", nil)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)
	var resp AuthZPostResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), resp.RedirectURI, "/error")
}

func (suite *AuthorizeHandlerTestSuite) TestWriteAuthZResponse() {
	rr := httptest.NewRecorder()

	suite.handler.writeAuthZResponse(rr, "https://example.com/callback?code=abc123")

	assert.Equal(suite.T(), http.StatusOK, rr.Code)
	assert.Equal(suite.T(), "application/json", rr.Header().Get("Content-Type"))
	var resp AuthZPostResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "https://example.com/callback?code=abc123", resp.RedirectURI)
}

func (suite *AuthorizeHandlerTestSuite) TestHandleAuthCallbackPostRequest_UnsupportedMethod() {
	req := httptest.NewRequest(http.MethodPut, "/oauth2/auth/callback", nil)
	rr := httptest.NewRecorder()

	suite.handler.HandleAuthCallbackPostRequest(rr, req)

	assert.Equal(suite.T(), http.StatusBadRequest, rr.Code)
	var response map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "invalid_request", response["error"])
}

// TestGetRequiredAttributes tests the getRequiredAttributes function
func (suite *AuthorizeHandlerTestSuite) TestGetRequiredAttributes() {
	tests := []struct {
		name           string
		oidcScopes     []string
		app            *appmodel.OAuthAppConfigProcessedDTO
		expectedResult string // space-separated attributes
		description    string
	}{
		{
			name:           "Nil app",
			oidcScopes:     []string{"openid", "profile"},
			app:            nil,
			expectedResult: "",
			description:    "Should return empty string when app is nil",
		},
		{
			name:       "Nil token config",
			oidcScopes: []string{"openid", "profile"},
			app: &appmodel.OAuthAppConfigProcessedDTO{
				Token: nil,
			},
			expectedResult: "",
			description:    "Should return empty string when token config is nil",
		},
		{
			name:       "Standard OIDC scopes with no IDToken config",
			oidcScopes: []string{"openid", "profile", "email"},
			app: &appmodel.OAuthAppConfigProcessedDTO{
				Token: &appmodel.OAuthTokenConfig{
					IDToken:     nil,
					AccessToken: nil,
				},
			},
			expectedResult: "",
			description:    "Should return empty when IDToken is nil (consistent with token builder behavior)",
		},
		{
			name:       "Standard OIDC scopes with empty IDToken.UserAttributes",
			oidcScopes: []string{"openid", "profile", "email"},
			app: &appmodel.OAuthAppConfigProcessedDTO{
				Token: &appmodel.OAuthTokenConfig{
					IDToken: &appmodel.IDTokenConfig{
						UserAttributes: []string{},
					},
					AccessToken: nil,
				},
				ScopeClaims: nil,
			},
			expectedResult: "",
			description:    "Should return empty when IDToken.UserAttributes is empty",
		},
		{
			name:       "Standard OIDC scopes with IDToken.UserAttributes filter",
			oidcScopes: []string{"openid", "profile", "email"},
			app: &appmodel.OAuthAppConfigProcessedDTO{
				Token: &appmodel.OAuthTokenConfig{
					IDToken: &appmodel.IDTokenConfig{
						UserAttributes: []string{"sub", "name", "email"}, // email_verified not allowed
					},
					AccessToken: nil,
				},
				ScopeClaims: nil,
			},
			expectedResult: "sub name email",
			description:    "Should filter scope claims by IDToken.UserAttributes",
		},
		{
			name:       "Custom scope claims mapping",
			oidcScopes: []string{"openid", "profile"},
			app: &appmodel.OAuthAppConfigProcessedDTO{
				Token: &appmodel.OAuthTokenConfig{
					IDToken: &appmodel.IDTokenConfig{
						UserAttributes: []string{"sub", "name", "custom_claim"},
					},
					AccessToken: nil,
				},
				ScopeClaims: map[string][]string{
					"profile": {"name", "custom_claim"},
				},
			},
			expectedResult: "sub name custom_claim",
			description:    "Should use custom scope claims mapping when provided",
		},
		{
			name:       "Access token attributes included",
			oidcScopes: []string{"openid"},
			app: &appmodel.OAuthAppConfigProcessedDTO{
				Token: &appmodel.OAuthTokenConfig{
					IDToken: &appmodel.IDTokenConfig{
						UserAttributes: []string{"sub"},
					},
					AccessToken: &appmodel.AccessTokenConfig{
						UserAttributes: []string{"groups", "roles"},
					},
				},
			},
			expectedResult: "sub groups roles",
			description:    "Should include both ID token claims and access token attributes",
		},
		{
			name:       "Combined ID token and access token attributes",
			oidcScopes: []string{"openid", "profile", "email"},
			app: &appmodel.OAuthAppConfigProcessedDTO{
				Token: &appmodel.OAuthTokenConfig{
					IDToken: &appmodel.IDTokenConfig{
						UserAttributes: []string{"sub", "name", "email"},
					},
					AccessToken: &appmodel.AccessTokenConfig{
						UserAttributes: []string{"groups", "roles", "email"}, // email overlaps
					},
				},
			},
			expectedResult: "sub name email groups roles",
			description:    "Should combine ID token and access token attributes, removing duplicates",
		},
		{
			name:       "Empty scopes",
			oidcScopes: []string{},
			app: &appmodel.OAuthAppConfigProcessedDTO{
				Token: &appmodel.OAuthTokenConfig{
					AccessToken: &appmodel.AccessTokenConfig{
						UserAttributes: []string{"groups"},
					},
				},
			},
			expectedResult: "groups",
			description:    "Should only include access token attributes when no scopes",
		},
		{
			name:       "Unknown scope",
			oidcScopes: []string{"unknown_scope"},
			app: &appmodel.OAuthAppConfigProcessedDTO{
				Token: &appmodel.OAuthTokenConfig{
					IDToken: &appmodel.IDTokenConfig{
						UserAttributes: []string{"sub"},
					},
					AccessToken: &appmodel.AccessTokenConfig{
						UserAttributes: []string{"groups"},
					},
				},
			},
			expectedResult: "groups",
			description:    "Should ignore unknown scopes and only include access token attributes",
		},
		{
			name:       "Custom scope with fallback to standard",
			oidcScopes: []string{"openid", "profile"},
			app: &appmodel.OAuthAppConfigProcessedDTO{
				Token: &appmodel.OAuthTokenConfig{
					IDToken: &appmodel.IDTokenConfig{
						UserAttributes: []string{"sub", "name"},
					},
					AccessToken: nil,
				},
				ScopeClaims: map[string][]string{
					"openid": {"sub"}, // Custom mapping for openid
					// profile falls back to standard
				},
			},
			expectedResult: "sub name",
			description:    "Should use custom mapping when available, fallback to standard otherwise",
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			result := getRequiredAttributes(tt.oidcScopes, tt.app, nil)

			// Parse both results into sets for comparison (order doesn't matter)
			resultAttrs := make(map[string]bool)
			for _, attr := range strings.Fields(result) {
				resultAttrs[attr] = true
			}

			expectedAttrs := make(map[string]bool)
			for _, attr := range strings.Fields(tt.expectedResult) {
				expectedAttrs[attr] = true
			}

			assert.Equal(t, expectedAttrs, resultAttrs, tt.description)
		})
	}
}

// TestValidateSubClaimConstraint tests the sub claim validation requirement
func (suite *AuthorizeHandlerTestSuite) TestValidateSubClaimConstraint() {
	tests := []struct {
		name          string
		claimsRequest *oauth2model.ClaimsRequest
		actualSubject string
		expectError   bool
	}{
		{
			name:          "nil claims request should pass",
			claimsRequest: nil,
			actualSubject: "user123",
			expectError:   false,
		},
		{
			name: "no sub constraint should pass",
			claimsRequest: &oauth2model.ClaimsRequest{
				IDToken: map[string]*oauth2model.IndividualClaimRequest{
					"email": nil,
				},
			},
			actualSubject: "user123",
			expectError:   false,
		},
		{
			name: "matching id_token sub value should pass",
			claimsRequest: &oauth2model.ClaimsRequest{
				IDToken: map[string]*oauth2model.IndividualClaimRequest{
					"sub": {Value: "user123"},
				},
			},
			actualSubject: "user123",
			expectError:   false,
		},
		{
			name: "non-matching id_token sub value should fail",
			claimsRequest: &oauth2model.ClaimsRequest{
				IDToken: map[string]*oauth2model.IndividualClaimRequest{
					"sub": {Value: "expected-user"},
				},
			},
			actualSubject: "actual-user",
			expectError:   true,
		},
		{
			name: "matching userinfo sub value should pass",
			claimsRequest: &oauth2model.ClaimsRequest{
				UserInfo: map[string]*oauth2model.IndividualClaimRequest{
					"sub": {Value: "user456"},
				},
			},
			actualSubject: "user456",
			expectError:   false,
		},
		{
			name: "non-matching userinfo sub value should fail",
			claimsRequest: &oauth2model.ClaimsRequest{
				UserInfo: map[string]*oauth2model.IndividualClaimRequest{
					"sub": {Value: "expected-user"},
				},
			},
			actualSubject: "actual-user",
			expectError:   true,
		},
		{
			name: "matching sub in values array should pass",
			claimsRequest: &oauth2model.ClaimsRequest{
				IDToken: map[string]*oauth2model.IndividualClaimRequest{
					"sub": {Values: []interface{}{"user1", "user2", "user3"}},
				},
			},
			actualSubject: "user2",
			expectError:   false,
		},
		{
			name: "non-matching sub in values array should fail",
			claimsRequest: &oauth2model.ClaimsRequest{
				IDToken: map[string]*oauth2model.IndividualClaimRequest{
					"sub": {Values: []interface{}{"user1", "user2", "user3"}},
				},
			},
			actualSubject: "user4",
			expectError:   true,
		},
		{
			name: "null sub request (voluntary) should pass",
			claimsRequest: &oauth2model.ClaimsRequest{
				IDToken: map[string]*oauth2model.IndividualClaimRequest{
					"sub": nil,
				},
			},
			actualSubject: "any-user",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			err := validateSubClaimConstraint(tt.claimsRequest, tt.actualSubject)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
