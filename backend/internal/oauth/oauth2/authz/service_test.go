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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	oauth2model "github.com/asgardeo/thunder/internal/oauth/oauth2/model"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/jose/jwt"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/tests/mocks/applicationmock"
	"github.com/asgardeo/thunder/tests/mocks/flow/flowexecmock"
	"github.com/asgardeo/thunder/tests/mocks/jose/jwtmock"
)

// JWT constants used in service tests.
const (
	// Header: {"alg":"none","typ":"JWT"}   Payload: {"sub":"test-user","iat":1701421200}
	svcJWTWithIat = "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJ0ZXN0LXVzZXIiLCJpYXQiOjE3MDE0MjEyMDB9."
	// Header: {"alg":"none","typ":"JWT"}   Payload: {"sub":"test-user"}
	svcJWTMinimal = "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJ0ZXN0LXVzZXIifQ."
)

type AuthorizeServiceTestSuite struct {
	suite.Suite
	mockAppService      *applicationmock.ApplicationServiceInterfaceMock
	mockJWTService      *jwtmock.JWTServiceInterfaceMock
	mockAuthzCodeStore  *AuthorizationCodeStoreInterfaceMock
	mockAuthReqStore    *authorizationRequestStoreInterfaceMock
	mockFlowExecService *flowexecmock.FlowExecServiceInterfaceMock
	mockValidator       *AuthorizationValidatorInterfaceMock
}

func TestAuthorizeServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AuthorizeServiceTestSuite))
}

func (suite *AuthorizeServiceTestSuite) BeforeTest(suiteName, testName string) {
	config.ResetThunderRuntime()
	testConfig := &config.Config{
		GateClient: config.GateClientConfig{
			Scheme:    "https",
			Hostname:  "localhost",
			Port:      3000,
			LoginPath: "/login",
			ErrorPath: "/error",
		},
		Database: config.DatabaseConfig{
			Identity: config.DataSource{Type: "sqlite", Path: ":memory:"},
			Runtime:  config.DataSource{Type: "sqlite", Path: ":memory:"},
		},
		OAuth: config.OAuthConfig{
			AuthorizationCode: config.AuthorizationCodeConfig{ValidityPeriod: 600},
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)
}

func (suite *AuthorizeServiceTestSuite) SetupTest() {
	suite.mockAppService = applicationmock.NewApplicationServiceInterfaceMock(suite.T())
	suite.mockJWTService = jwtmock.NewJWTServiceInterfaceMock(suite.T())
	suite.mockAuthzCodeStore = NewAuthorizationCodeStoreInterfaceMock(suite.T())
	suite.mockAuthReqStore = newAuthorizationRequestStoreInterfaceMock(suite.T())
	suite.mockFlowExecService = flowexecmock.NewFlowExecServiceInterfaceMock(suite.T())
	suite.mockValidator = NewAuthorizationValidatorInterfaceMock(suite.T())
}

// newService builds an authorizeService with all mocked dependencies.
func (suite *AuthorizeServiceTestSuite) newService() *authorizeService {
	return &authorizeService{
		appService:      suite.mockAppService,
		authZValidator:  suite.mockValidator,
		authCodeStore:   suite.mockAuthzCodeStore,
		authReqStore:    suite.mockAuthReqStore,
		jwtService:      suite.mockJWTService,
		flowExecService: suite.mockFlowExecService,
		logger:          log.GetLogger().With(log.String(log.LoggerKeyComponentName, "AuthorizeServiceTest")),
	}
}

// testApp returns a minimal OAuthAppConfigProcessedDTO for use in tests.
func (suite *AuthorizeServiceTestSuite) testApp() *appmodel.OAuthAppConfigProcessedDTO {
	return &appmodel.OAuthAppConfigProcessedDTO{
		AppID:        "test-app-id",
		ClientID:     "test-client-id",
		RedirectURIs: []string{"https://client.example.com/callback"},
		GrantTypes:   []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
		PKCERequired: false,
	}
}

// testMsg returns a minimal OAuthMessage for initial authorization requests.
func (suite *AuthorizeServiceTestSuite) testMsg() *OAuthMessage {
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

func (suite *AuthorizeServiceTestSuite) TestHandleInitialAuthorizationRequest_MissingClientID() {
	msg := &OAuthMessage{
		RequestType: oauth2const.TypeInitialAuthorizationRequest,
		RequestQueryParams: map[string]string{
			"redirect_uri":  "https://client.example.com/callback",
			"response_type": "code",
		},
	}

	svc := suite.newService()
	result, authErr := svc.HandleInitialAuthorizationRequest(msg)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), authErr)
	assert.Equal(suite.T(), oauth2const.ErrorInvalidRequest, authErr.Code)
	assert.Contains(suite.T(), authErr.Message, "Missing client_id")
}

func (suite *AuthorizeServiceTestSuite) TestHandleInitialAuthorizationRequest_InvalidClient() {
	notFound := &serviceerror.ServiceError{Type: serviceerror.ClientErrorType, Error: "Application not found"}
	suite.mockAppService.EXPECT().GetOAuthApplication("invalid-client").Return(nil, notFound)

	msg := &OAuthMessage{
		RequestType: oauth2const.TypeInitialAuthorizationRequest,
		RequestQueryParams: map[string]string{
			"client_id":     "invalid-client",
			"redirect_uri":  "https://client.example.com/callback",
			"response_type": "code",
		},
	}

	svc := suite.newService()
	result, authErr := svc.HandleInitialAuthorizationRequest(msg)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), authErr)
	assert.Equal(suite.T(), oauth2const.ErrorInvalidClient, authErr.Code)
}

func (suite *AuthorizeServiceTestSuite) TestHandleInitialAuthorizationRequest_InvalidClaimsParameter() {
	app := suite.testApp()
	suite.mockAppService.EXPECT().GetOAuthApplication("test-client-id").Return(app, nil)

	msg := &OAuthMessage{
		RequestType: oauth2const.TypeInitialAuthorizationRequest,
		RequestQueryParams: map[string]string{
			"client_id":    "test-client-id",
			"redirect_uri": "https://client.example.com/callback",
			"claims":       "{invalid json}",
		},
	}

	svc := suite.newService()
	result, authErr := svc.HandleInitialAuthorizationRequest(msg)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), authErr)
	assert.Equal(suite.T(), oauth2const.ErrorInvalidRequest, authErr.Code)
}

func (suite *AuthorizeServiceTestSuite) TestHandleInitialAuthorizationRequest_ValidationError_NoClientRedirect() {
	app := suite.testApp()
	suite.mockAppService.EXPECT().GetOAuthApplication("test-client-id").Return(app, nil)

	// Validator rejects; sendErrorToApp=false → error goes to error page, not client.
	suite.mockValidator.On("validateInitialAuthorizationRequest", mock.Anything, app).
		Return(false, oauth2const.ErrorInvalidRequest, "Missing required parameter")

	msg := suite.testMsg()
	svc := suite.newService()
	result, authErr := svc.HandleInitialAuthorizationRequest(msg)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), authErr)
	assert.Equal(suite.T(), oauth2const.ErrorInvalidRequest, authErr.Code)
	assert.False(suite.T(), authErr.SendErrorToClient)
	assert.Equal(suite.T(), "test-state", authErr.State)
}

func (suite *AuthorizeServiceTestSuite) TestHandleInitialAuthorizationRequest_ValidationError_SendToClient() {
	app := suite.testApp()
	suite.mockAppService.EXPECT().GetOAuthApplication("test-client-id").Return(app, nil)

	// sendErrorToApp=true + redirect_uri present → error forwarded to client.
	suite.mockValidator.On("validateInitialAuthorizationRequest", mock.Anything, app).
		Return(true, oauth2const.ErrorUnsupportedResponseType, "Unsupported response_type value")

	msg := suite.testMsg()
	svc := suite.newService()
	result, authErr := svc.HandleInitialAuthorizationRequest(msg)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), authErr)
	assert.Equal(suite.T(), oauth2const.ErrorUnsupportedResponseType, authErr.Code)
	assert.True(suite.T(), authErr.SendErrorToClient)
	assert.Equal(suite.T(), "https://client.example.com/callback", authErr.ClientRedirectURI)
	assert.Equal(suite.T(), "test-state", authErr.State)
}

func (suite *AuthorizeServiceTestSuite) TestHandleInitialAuthorizationRequest_FlowInitError() {
	app := suite.testApp()
	suite.mockAppService.EXPECT().GetOAuthApplication("test-client-id").Return(app, nil)
	suite.mockValidator.On("validateInitialAuthorizationRequest", mock.Anything, app).
		Return(false, "", "")
	suite.mockFlowExecService.EXPECT().InitiateFlow(mock.Anything).Return("", &serviceerror.InternalServerError)

	svc := suite.newService()
	result, authErr := svc.HandleInitialAuthorizationRequest(suite.testMsg())

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), authErr)
	assert.Equal(suite.T(), oauth2const.ErrorServerError, authErr.Code)
}

func (suite *AuthorizeServiceTestSuite) TestHandleInitialAuthorizationRequest_Success() {
	app := suite.testApp()
	suite.mockAppService.EXPECT().GetOAuthApplication("test-client-id").Return(app, nil)
	suite.mockValidator.On("validateInitialAuthorizationRequest", mock.Anything, app).
		Return(false, "", "")
	suite.mockFlowExecService.EXPECT().InitiateFlow(mock.Anything).Return("test-flow-id", nil)
	suite.mockAuthReqStore.EXPECT().AddRequest(mock.Anything).Return(testAuthID)

	svc := suite.newService()
	result, authErr := svc.HandleInitialAuthorizationRequest(suite.testMsg())

	assert.Nil(suite.T(), authErr)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), testAuthID, result.QueryParams[oauth2const.AuthID])
	assert.Equal(suite.T(), "test-app-id", result.QueryParams[oauth2const.AppID])
	assert.Equal(suite.T(), "test-flow-id", result.QueryParams[oauth2const.FlowID])
}

func (suite *AuthorizeServiceTestSuite) TestHandleInitialAuthorizationRequest_InsecureRedirectURI() {
	app := suite.testApp()
	app.RedirectURIs = []string{"http://client.example.com/callback"}
	suite.mockAppService.EXPECT().GetOAuthApplication("test-client-id").Return(app, nil)
	suite.mockValidator.On("validateInitialAuthorizationRequest", mock.Anything, app).
		Return(false, "", "")
	suite.mockFlowExecService.EXPECT().InitiateFlow(mock.Anything).Return("test-flow-id", nil)
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

	svc := suite.newService()
	result, authErr := svc.HandleInitialAuthorizationRequest(msg)

	assert.Nil(suite.T(), authErr)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "true", result.QueryParams[oauth2const.ShowInsecureWarning])
}

func (suite *AuthorizeServiceTestSuite) TestHandleInitialAuthorizationRequest_EmptyRedirectURIUsesAppDefault() {
	app := suite.testApp() // RedirectURIs: ["https://client.example.com/callback"]
	suite.mockAppService.EXPECT().GetOAuthApplication("test-client-id").Return(app, nil)
	suite.mockValidator.On("validateInitialAuthorizationRequest", mock.Anything, app).
		Return(false, "", "")
	suite.mockFlowExecService.EXPECT().InitiateFlow(mock.Anything).Return("test-flow-id", nil)
	suite.mockAuthReqStore.EXPECT().AddRequest(mock.Anything).Return(testAuthID)

	msg := &OAuthMessage{
		RequestType: oauth2const.TypeInitialAuthorizationRequest,
		RequestQueryParams: map[string]string{
			"client_id":     "test-client-id",
			"response_type": "code",
			"scope":         "read write",
			// No redirect_uri — service should use app.RedirectURIs[0].
		},
	}

	svc := suite.newService()
	result, authErr := svc.HandleInitialAuthorizationRequest(msg)

	assert.Nil(suite.T(), authErr)
	assert.NotNil(suite.T(), result)
}

func (suite *AuthorizeServiceTestSuite) TestHandleInitialAuthorizationRequest_WithClaimsLocales() {
	app := suite.testApp()
	suite.mockAppService.EXPECT().GetOAuthApplication("test-client-id").Return(app, nil)
	suite.mockValidator.On("validateInitialAuthorizationRequest", mock.Anything, app).
		Return(false, "", "")
	suite.mockFlowExecService.EXPECT().InitiateFlow(mock.Anything).Return("test-flow-id", nil)
	suite.mockAuthReqStore.EXPECT().AddRequest(mock.Anything).Return(testAuthID)

	msg := &OAuthMessage{
		RequestType: oauth2const.TypeInitialAuthorizationRequest,
		RequestQueryParams: map[string]string{
			"client_id":      "test-client-id",
			"redirect_uri":   "https://client.example.com/callback",
			"response_type":  "code",
			"scope":          "openid read write",
			"claims_locales": "en-US fr-CA",
		},
	}

	svc := suite.newService()
	result, authErr := svc.HandleInitialAuthorizationRequest(msg)

	assert.Nil(suite.T(), authErr)
	assert.NotNil(suite.T(), result)
}

func (suite *AuthorizeServiceTestSuite) TestHandleAuthorizationCallback_InvalidAuthID() {
	suite.mockAuthReqStore.EXPECT().GetRequest("invalid-key").Return(false, authRequestContext{})

	svc := suite.newService()
	redirectURI, authErr := svc.HandleAuthorizationCallback("invalid-key", "test-assertion")

	assert.Empty(suite.T(), redirectURI)
	assert.NotNil(suite.T(), authErr)
	assert.Equal(suite.T(), oauth2const.ErrorInvalidRequest, authErr.Code)
}

func (suite *AuthorizeServiceTestSuite) TestHandleAuthorizationCallback_MissingAssertion() {
	authCtx := authRequestContext{
		OAuthParameters: oauth2model.OAuthParameters{
			ClientID:    "test-client",
			RedirectURI: "https://client.example.com/callback",
			State:       "test-state",
		},
	}
	suite.mockAuthReqStore.EXPECT().GetRequest(testAuthID).Return(true, authCtx)
	suite.mockAuthReqStore.EXPECT().ClearRequest(testAuthID)

	svc := suite.newService()
	redirectURI, authErr := svc.HandleAuthorizationCallback(testAuthID, "")

	assert.Empty(suite.T(), redirectURI)
	assert.NotNil(suite.T(), authErr)
	assert.Equal(suite.T(), oauth2const.ErrorInvalidRequest, authErr.Code)
	assert.Equal(suite.T(), "test-state", authErr.State)
}

func (suite *AuthorizeServiceTestSuite) TestHandleAuthorizationCallback_InvalidAssertionSignature() {
	authCtx := authRequestContext{
		OAuthParameters: oauth2model.OAuthParameters{
			ClientID:    "test-client",
			RedirectURI: "https://client.example.com/callback",
			State:       "test-state",
		},
	}
	suite.mockAuthReqStore.EXPECT().GetRequest(testAuthID).Return(true, authCtx)
	suite.mockAuthReqStore.EXPECT().ClearRequest(testAuthID)
	suite.mockJWTService.EXPECT().VerifyJWT("invalid-assertion", "", "").Return(&jwt.ErrorInvalidTokenSignature)

	svc := suite.newService()
	redirectURI, authErr := svc.HandleAuthorizationCallback(testAuthID, "invalid-assertion")

	assert.Empty(suite.T(), redirectURI)
	assert.NotNil(suite.T(), authErr)
	assert.Equal(suite.T(), oauth2const.ErrorInvalidRequest, authErr.Code)
	assert.Equal(suite.T(), "test-state", authErr.State)
}

func (suite *AuthorizeServiceTestSuite) TestHandleAuthorizationCallback_FailedToDecodeAssertion() {
	authCtx := authRequestContext{
		OAuthParameters: oauth2model.OAuthParameters{
			ClientID:    "test-client",
			RedirectURI: "https://client.example.com/callback",
			State:       "test-state",
		},
	}
	suite.mockAuthReqStore.EXPECT().GetRequest(testAuthID).Return(true, authCtx)
	suite.mockAuthReqStore.EXPECT().ClearRequest(testAuthID)
	// VerifyJWT succeeds but "not.valid.jwt" cannot be decoded as a valid JWT payload.
	suite.mockJWTService.EXPECT().VerifyJWT("not.valid.jwt", "", "").Return(nil)

	svc := suite.newService()
	redirectURI, authErr := svc.HandleAuthorizationCallback(testAuthID, "not.valid.jwt")

	assert.Empty(suite.T(), redirectURI)
	assert.NotNil(suite.T(), authErr)
	assert.Equal(suite.T(), "test-state", authErr.State)
}

func (suite *AuthorizeServiceTestSuite) TestHandleAuthorizationCallback_PersistAuthCodeError() {
	authCtx := authRequestContext{
		OAuthParameters: oauth2model.OAuthParameters{
			ClientID:    "test-client",
			RedirectURI: "https://client.example.com/callback",
			State:       "test-state",
		},
	}
	suite.mockAuthReqStore.EXPECT().GetRequest(testAuthID).Return(true, authCtx)
	suite.mockAuthReqStore.EXPECT().ClearRequest(testAuthID)
	suite.mockJWTService.EXPECT().VerifyJWT(svcJWTWithIat, "", "").Return(nil)
	suite.mockAuthzCodeStore.EXPECT().InsertAuthorizationCode(mock.Anything).Return(errors.New("db error"))

	svc := suite.newService()
	redirectURI, authErr := svc.HandleAuthorizationCallback(testAuthID, svcJWTWithIat)

	assert.Empty(suite.T(), redirectURI)
	assert.NotNil(suite.T(), authErr)
	assert.Equal(suite.T(), oauth2const.ErrorServerError, authErr.Code)
	assert.Equal(suite.T(), "test-state", authErr.State)
}

func (suite *AuthorizeServiceTestSuite) TestHandleAuthorizationCallback_Success() {
	authCtx := authRequestContext{
		OAuthParameters: oauth2model.OAuthParameters{
			ClientID:    "test-client",
			RedirectURI: "https://client.example.com/callback",
		},
	}
	suite.mockAuthReqStore.EXPECT().GetRequest(testAuthID).Return(true, authCtx)
	suite.mockAuthReqStore.EXPECT().ClearRequest(testAuthID)
	suite.mockJWTService.EXPECT().VerifyJWT(svcJWTWithIat, "", "").Return(nil)
	suite.mockAuthzCodeStore.EXPECT().InsertAuthorizationCode(mock.Anything).Return(nil)

	svc := suite.newService()
	redirectURI, authErr := svc.HandleAuthorizationCallback(testAuthID, svcJWTWithIat)

	assert.Nil(suite.T(), authErr)
	assert.Contains(suite.T(), redirectURI, "https://client.example.com/callback")
	assert.Contains(suite.T(), redirectURI, "code=")
}

func (suite *AuthorizeServiceTestSuite) TestHandleAuthorizationCallback_WithState() {
	authCtx := authRequestContext{
		OAuthParameters: oauth2model.OAuthParameters{
			ClientID:    "test-client",
			RedirectURI: "https://client.example.com/callback",
			State:       "test-state-123",
		},
	}
	suite.mockAuthReqStore.EXPECT().GetRequest(testAuthID).Return(true, authCtx)
	suite.mockAuthReqStore.EXPECT().ClearRequest(testAuthID)
	suite.mockJWTService.EXPECT().VerifyJWT(svcJWTWithIat, "", "").Return(nil)
	suite.mockAuthzCodeStore.EXPECT().InsertAuthorizationCode(mock.Anything).Return(nil)

	svc := suite.newService()
	redirectURI, authErr := svc.HandleAuthorizationCallback(testAuthID, svcJWTWithIat)

	assert.Nil(suite.T(), authErr)
	assert.Contains(suite.T(), redirectURI, "state=test-state-123")
	assert.Contains(suite.T(), redirectURI, "code=")
}

func (suite *AuthorizeServiceTestSuite) TestHandleAuthorizationCallback_EmptyAuthorizedPermissions() {
	// svcJWTWithIat has only "sub" and "iat" — no authorized_permissions.
	// Permission scopes in the auth context should be cleared.
	authCtx := authRequestContext{
		OAuthParameters: oauth2model.OAuthParameters{
			ClientID:         "test-client",
			RedirectURI:      "https://client.example.com/callback",
			PermissionScopes: []string{"read", "write"},
		},
	}
	suite.mockAuthReqStore.EXPECT().GetRequest(testAuthID).Return(true, authCtx)
	suite.mockAuthReqStore.EXPECT().ClearRequest(testAuthID)
	suite.mockJWTService.EXPECT().VerifyJWT(svcJWTWithIat, "", "").Return(nil)
	suite.mockAuthzCodeStore.EXPECT().InsertAuthorizationCode(mock.Anything).Return(nil)

	svc := suite.newService()
	redirectURI, authErr := svc.HandleAuthorizationCallback(testAuthID, svcJWTWithIat)

	assert.Nil(suite.T(), authErr)
	assert.NotEmpty(suite.T(), redirectURI)
}

func (suite *AuthorizeServiceTestSuite) TestHandleAuthorizationCallback_CreateAuthCodeError() {
	// Empty ClientID in auth context → createAuthorizationCode will fail.
	authCtx := authRequestContext{
		OAuthParameters: oauth2model.OAuthParameters{
			ClientID:    "",
			RedirectURI: "https://client.example.com/callback",
		},
	}
	suite.mockAuthReqStore.EXPECT().GetRequest(testAuthID).Return(true, authCtx)
	suite.mockAuthReqStore.EXPECT().ClearRequest(testAuthID)
	suite.mockJWTService.EXPECT().VerifyJWT(svcJWTMinimal, "", "").Return(nil)

	svc := suite.newService()
	redirectURI, authErr := svc.HandleAuthorizationCallback(testAuthID, svcJWTMinimal)

	assert.Empty(suite.T(), redirectURI)
	assert.NotNil(suite.T(), authErr)
	assert.Equal(suite.T(), oauth2const.ErrorServerError, authErr.Code)
}

func (suite *AuthorizeServiceTestSuite) TestGetAuthorizationCodeDetails_NotFound() {
	suite.mockAuthzCodeStore.EXPECT().GetAuthorizationCode("client-id", "invalid-code").
		Return(nil, ErrAuthorizationCodeNotFound)

	svc := suite.newService()
	result, err := svc.GetAuthorizationCodeDetails("client-id", "invalid-code")

	assert.Nil(suite.T(), result)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "invalid authorization code")
}

func (suite *AuthorizeServiceTestSuite) TestGetAuthorizationCodeDetails_GetError() {
	suite.mockAuthzCodeStore.EXPECT().GetAuthorizationCode("client-id", "code").
		Return(nil, errors.New("database error"))

	svc := suite.newService()
	result, err := svc.GetAuthorizationCodeDetails("client-id", "code")

	assert.Nil(suite.T(), result)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to retrieve authorization code")
}

func (suite *AuthorizeServiceTestSuite) TestGetAuthorizationCodeDetails_Success() {
	authCode := &AuthorizationCode{
		CodeID:           "code-id-123",
		Code:             "valid-code",
		ClientID:         "client-id",
		AuthorizedUserID: "user-123",
	}
	suite.mockAuthzCodeStore.EXPECT().GetAuthorizationCode("client-id", "valid-code").
		Return(authCode, nil)
	suite.mockAuthzCodeStore.EXPECT().DeactivateAuthorizationCode(*authCode).Return(nil)

	svc := suite.newService()
	result, err := svc.GetAuthorizationCodeDetails("client-id", "valid-code")

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "valid-code", result.Code)
	assert.Equal(suite.T(), "user-123", result.AuthorizedUserID)
}

func (suite *AuthorizeServiceTestSuite) TestGetAuthorizationCodeDetails_DeactivateError() {
	authCode := &AuthorizationCode{
		CodeID:           "code-id-123",
		Code:             "valid-code",
		ClientID:         "client-id",
		AuthorizedUserID: "user-123",
	}
	suite.mockAuthzCodeStore.EXPECT().GetAuthorizationCode("client-id", "valid-code").
		Return(authCode, nil)
	suite.mockAuthzCodeStore.EXPECT().DeactivateAuthorizationCode(*authCode).
		Return(errors.New("deactivate error"))

	svc := suite.newService()
	result, err := svc.GetAuthorizationCodeDetails("client-id", "valid-code")

	assert.Nil(suite.T(), result)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to invalidate authorization code")
}
