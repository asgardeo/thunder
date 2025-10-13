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

package granthandlers

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	authzconstants "github.com/asgardeo/thunder/internal/oauth/oauth2/authz/constants"
	authzmodel "github.com/asgardeo/thunder/internal/oauth/oauth2/authz/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/model"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/user"
	"github.com/asgardeo/thunder/tests/mocks/jwtmock"
	"github.com/asgardeo/thunder/tests/mocks/oauth/oauth2/authz/storemock"
	usersvcmock "github.com/asgardeo/thunder/tests/mocks/usermock"
)

type AuthorizationCodeGrantHandlerTestSuite struct {
	suite.Suite
	handler         *authorizationCodeGrantHandler
	mockJWTService  *jwtmock.JWTServiceInterfaceMock
	mockAuthZStore  *storemock.AuthorizationCodeStoreInterfaceMock
	mockUserService *usersvcmock.UserServiceInterfaceMock
	oauthApp        *appmodel.OAuthAppConfigProcessedDTO
	testAuthzCode   authzmodel.AuthorizationCode
	testTokenReq    *model.TokenRequest
}

func TestAuthorizationCodeGrantHandlerSuite(t *testing.T) {
	suite.Run(t, new(AuthorizationCodeGrantHandlerTestSuite))
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) SetupTest() {
	// Initialize Thunder Runtime config with basic test config
	testConfig := &config.Config{
		OAuth: config.OAuthConfig{
			JWT: config.JWTConfig{
				ValidityPeriod: 3600,
			},
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)

	suite.mockJWTService = &jwtmock.JWTServiceInterfaceMock{}
	suite.mockAuthZStore = &storemock.AuthorizationCodeStoreInterfaceMock{}
	suite.mockUserService = usersvcmock.NewUserServiceInterfaceMock(suite.T())

	suite.handler = &authorizationCodeGrantHandler{
		JWTService:  suite.mockJWTService,
		AuthZStore:  suite.mockAuthZStore,
		UserService: suite.mockUserService,
	}

	suite.oauthApp = &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:           "test-client-id",
		HashedClientSecret: "hashed-secret",
		RedirectURIs:       []string{"https://client.example.com/callback"},
		GrantTypes:         []constants.GrantType{constants.GrantTypeAuthorizationCode},
		ResponseTypes:      []constants.ResponseType{constants.ResponseTypeCode},
		TokenEndpointAuthMethod: []constants.TokenEndpointAuthMethod{
			constants.TokenEndpointAuthMethodClientSecretPost},
		Token: &appmodel.OAuthTokenConfig{
			AccessToken: &appmodel.TokenConfig{
				UserAttributes: []string{"email", "username"},
			},
		},
	}

	suite.testTokenReq = &model.TokenRequest{
		GrantType:   string(constants.GrantTypeAuthorizationCode),
		ClientID:    "test-client-id",
		Code:        "test-auth-code",
		RedirectURI: "https://client.example.com/callback",
	}

	suite.testAuthzCode = authzmodel.AuthorizationCode{
		CodeID:           "test-code-id",
		Code:             "test-auth-code",
		ClientID:         "test-client-id",
		RedirectURI:      "https://client.example.com/callback",
		AuthorizedUserID: "test-user-id",
		TimeCreated:      time.Now().Add(-5 * time.Minute),
		ExpiryTime:       time.Now().Add(5 * time.Minute),
		Scopes:           "read write",
		State:            authzconstants.AuthCodeStateActive,
	}
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestNewAuthorizationCodeGrantHandler() {
	handler := newAuthorizationCodeGrantHandler()
	assert.NotNil(suite.T(), handler)
	assert.Implements(suite.T(), (*GrantHandlerInterface)(nil), handler)
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestValidateGrant_Success() {
	err := suite.handler.ValidateGrant(suite.testTokenReq, suite.oauthApp)
	assert.Nil(suite.T(), err)
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestValidateGrant_MissingGrantType() {
	tokenReq := &model.TokenRequest{
		GrantType: "", // Missing grant type
		ClientID:  "test-client-id",
		Code:      "test-code",
	}

	err := suite.handler.ValidateGrant(tokenReq, suite.oauthApp)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, err.Error)
	assert.Equal(suite.T(), "Missing grant type", err.ErrorDescription)
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestValidateGrant_UnsupportedGrantType() {
	tokenReq := &model.TokenRequest{
		GrantType: string(constants.GrantTypeClientCredentials), // Wrong grant type
		ClientID:  "test-client-id",
		Code:      "test-code",
	}

	err := suite.handler.ValidateGrant(tokenReq, suite.oauthApp)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorUnsupportedGrantType, err.Error)
	assert.Equal(suite.T(), "Unsupported grant type", err.ErrorDescription)
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestValidateGrant_MissingAuthorizationCode() {
	tokenReq := &model.TokenRequest{
		GrantType: string(constants.GrantTypeAuthorizationCode),
		ClientID:  "test-client-id",
		Code:      "", // Missing authorization code
	}

	err := suite.handler.ValidateGrant(tokenReq, suite.oauthApp)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorInvalidGrant, err.Error)
	assert.Equal(suite.T(), "Authorization code is required", err.ErrorDescription)
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestValidateGrant_MissingClientID() {
	tokenReq := &model.TokenRequest{
		GrantType: string(constants.GrantTypeAuthorizationCode),
		ClientID:  "", // Missing client ID
		Code:      "test-code",
	}

	err := suite.handler.ValidateGrant(tokenReq, suite.oauthApp)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorInvalidClient, err.Error)
	assert.Equal(suite.T(), "Client Id is required", err.ErrorDescription)
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestValidateGrant_MissingRedirectURI() {
	tokenReq := &model.TokenRequest{
		GrantType:   string(constants.GrantTypeAuthorizationCode),
		ClientID:    "test-client-id",
		Code:        "test-code",
		RedirectURI: "", // Missing redirect URI
	}

	err := suite.handler.ValidateGrant(tokenReq, suite.oauthApp)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, err.Error)
	assert.Equal(suite.T(), "Redirect URI is required", err.ErrorDescription)
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestHandleGrant_Success() {
	// Mock authorization code store to return valid code
	suite.mockAuthZStore.On("GetAuthorizationCode", "test-client-id", "test-auth-code").
		Return(suite.testAuthzCode, nil)
	suite.mockAuthZStore.On("DeactivateAuthorizationCode", suite.testAuthzCode).Return(nil)

	// Mock user service to return user for attributes
	mockUser := &user.User{
		ID:         "test-user-id",
		Attributes: json.RawMessage(`{"email":"test@example.com","username":"testuser"}`),
	}
	suite.mockUserService.On("GetUser", "test-user-id").Return(mockUser, nil)

	// Mock JWT service to generate token
	suite.mockJWTService.On("GenerateJWT", "test-user-id", "test-client-id",
		mock.AnythingOfType("string"), mock.AnythingOfType("int64"), mock.AnythingOfType("map[string]interface {}")).
		Return("test-jwt-token", int64(3600), nil)

	ctx := &model.TokenContext{
		TokenAttributes: make(map[string]interface{}),
	}

	result, err := suite.handler.HandleGrant(suite.testTokenReq, suite.oauthApp, ctx)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "test-jwt-token", result.AccessToken.Token)
	assert.Equal(suite.T(), constants.TokenTypeBearer, result.AccessToken.TokenType)
	assert.Equal(suite.T(), int64(3600), result.AccessToken.ExpiresIn)
	assert.Equal(suite.T(), []string{"read", "write"}, result.AccessToken.Scopes)
	assert.Equal(suite.T(), "test-client-id", result.AccessToken.ClientID)

	// Check context attributes
	assert.Equal(suite.T(), "test-user-id", ctx.TokenAttributes["sub"])
	assert.Equal(suite.T(), "test-client-id", ctx.TokenAttributes["aud"])

	suite.mockAuthZStore.AssertExpectations(suite.T())
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestHandleGrant_InvalidAuthorizationCode() {
	// Mock authorization code store to return error
	suite.mockAuthZStore.On("GetAuthorizationCode", "test-client-id", "test-auth-code").
		Return(authzmodel.AuthorizationCode{}, errors.New("code not found"))

	ctx := &model.TokenContext{
		TokenAttributes: make(map[string]interface{}),
	}

	result, err := suite.handler.HandleGrant(suite.testTokenReq, suite.oauthApp, ctx)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorInvalidGrant, err.Error)
	assert.Equal(suite.T(), "Invalid authorization code", err.ErrorDescription)

	suite.mockAuthZStore.AssertExpectations(suite.T())
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestHandleGrant_EmptyAuthorizationCode() {
	// Mock authorization code store to return empty code
	emptyCode := authzmodel.AuthorizationCode{Code: ""}
	suite.mockAuthZStore.On("GetAuthorizationCode", "test-client-id", "test-auth-code").
		Return(emptyCode, nil)

	ctx := &model.TokenContext{
		TokenAttributes: make(map[string]interface{}),
	}

	result, err := suite.handler.HandleGrant(suite.testTokenReq, suite.oauthApp, ctx)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorInvalidGrant, err.Error)
	assert.Equal(suite.T(), "Invalid authorization code", err.ErrorDescription)

	suite.mockAuthZStore.AssertExpectations(suite.T())
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestHandleGrant_DeactivateError() {
	// Mock authorization code store to return valid code but fail deactivation
	suite.mockAuthZStore.On("GetAuthorizationCode", "test-client-id", "test-auth-code").
		Return(suite.testAuthzCode, nil)
	suite.mockAuthZStore.On("DeactivateAuthorizationCode", suite.testAuthzCode).
		Return(errors.New("deactivate failed"))

	ctx := &model.TokenContext{
		TokenAttributes: make(map[string]interface{}),
	}

	result, err := suite.handler.HandleGrant(suite.testTokenReq, suite.oauthApp, ctx)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorServerError, err.Error)
	assert.Equal(suite.T(), "Failed to invalidate authorization code", err.ErrorDescription)

	suite.mockAuthZStore.AssertExpectations(suite.T())
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestHandleGrant_JWTGenerationError() {
	// Mock authorization code store to return valid code
	suite.mockAuthZStore.On("GetAuthorizationCode", "test-client-id", "test-auth-code").
		Return(suite.testAuthzCode, nil)
	suite.mockAuthZStore.On("DeactivateAuthorizationCode", suite.testAuthzCode).Return(nil)

	// Mock user service to return user for attributes
	mockUser := &user.User{
		ID:         "test-user-id",
		Attributes: json.RawMessage(`{"email":"test@example.com","username":"testuser"}`),
	}
	suite.mockUserService.On("GetUser", "test-user-id").Return(mockUser, nil)

	// Mock JWT service to fail token generation
	suite.mockJWTService.On("GenerateJWT", "test-user-id", "test-client-id",
		mock.AnythingOfType("string"), mock.AnythingOfType("int64"), mock.AnythingOfType("map[string]interface {}")).
		Return("", int64(0), errors.New("jwt generation failed"))

	ctx := &model.TokenContext{
		TokenAttributes: make(map[string]interface{}),
	}

	result, err := suite.handler.HandleGrant(suite.testTokenReq, suite.oauthApp, ctx)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorServerError, err.Error)
	assert.Equal(suite.T(), "Failed to generate token", err.ErrorDescription)

	suite.mockAuthZStore.AssertExpectations(suite.T())
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestHandleGrant_EmptyScopes() {
	// Test with empty scopes
	authzCodeWithEmptyScopes := suite.testAuthzCode
	authzCodeWithEmptyScopes.Scopes = ""

	suite.mockAuthZStore.On("GetAuthorizationCode", "test-client-id", "test-auth-code").
		Return(authzCodeWithEmptyScopes, nil)
	suite.mockAuthZStore.On("DeactivateAuthorizationCode", authzCodeWithEmptyScopes).Return(nil)

	// Mock user service to return user for attributes
	mockUser := &user.User{
		ID:         "test-user-id",
		Attributes: json.RawMessage(`{"email":"test@example.com","username":"testuser"}`),
	}
	suite.mockUserService.On("GetUser", "test-user-id").Return(mockUser, nil)

	suite.mockJWTService.On("GenerateJWT", "test-user-id", "test-client-id",
		mock.AnythingOfType("string"), mock.AnythingOfType("int64"), mock.AnythingOfType("map[string]interface {}")).
		Return("test-jwt-token", int64(3600), nil)

	ctx := &model.TokenContext{
		TokenAttributes: make(map[string]interface{}),
	}

	result, err := suite.handler.HandleGrant(suite.testTokenReq, suite.oauthApp, ctx)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Empty(suite.T(), result.AccessToken.Scopes)

	suite.mockAuthZStore.AssertExpectations(suite.T())
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestHandleGrant_NilTokenAttributes() {
	// Test with nil token attributes
	suite.mockAuthZStore.On("GetAuthorizationCode", "test-client-id", "test-auth-code").
		Return(suite.testAuthzCode, nil)
	suite.mockAuthZStore.On("DeactivateAuthorizationCode", suite.testAuthzCode).Return(nil)

	// Mock user service to return user for attributes
	mockUser := &user.User{
		ID:         "test-user-id",
		Attributes: json.RawMessage(`{"email":"test@example.com","username":"testuser"}`),
	}
	suite.mockUserService.On("GetUser", "test-user-id").Return(mockUser, nil)

	suite.mockJWTService.On("GenerateJWT", "test-user-id", "test-client-id",
		mock.AnythingOfType("string"), mock.AnythingOfType("int64"), mock.AnythingOfType("map[string]interface {}")).
		Return("test-jwt-token", int64(3600), nil)

	ctx := &model.TokenContext{
		TokenAttributes: nil, // Nil attributes
	}

	result, err := suite.handler.HandleGrant(suite.testTokenReq, suite.oauthApp, ctx)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)

	// Should have initialized TokenAttributes
	assert.NotNil(suite.T(), ctx.TokenAttributes)
	assert.Equal(suite.T(), "test-user-id", ctx.TokenAttributes["sub"])
	assert.Equal(suite.T(), "test-client-id", ctx.TokenAttributes["aud"])

	suite.mockAuthZStore.AssertExpectations(suite.T())
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestValidateAuthorizationCode_Success() {
	err := validateAuthorizationCode(suite.testTokenReq, suite.testAuthzCode)
	assert.Nil(suite.T(), err)
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestValidateAuthorizationCode_WrongClientID() {
	invalidTokenReq := &model.TokenRequest{
		ClientID: "wrong-client-id", // Wrong client ID
	}

	err := validateAuthorizationCode(invalidTokenReq, suite.testAuthzCode)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorInvalidClient, err.Error)
	assert.Equal(suite.T(), "Invalid client Id", err.ErrorDescription)
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestValidateAuthorizationCode_WrongRedirectURI() {
	invalidTokenReq := &model.TokenRequest{
		ClientID:    "test-client-id",
		RedirectURI: "https://wrong.example.com/callback", // Wrong redirect URI
	}

	err := validateAuthorizationCode(invalidTokenReq, suite.testAuthzCode)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorInvalidGrant, err.Error)
	assert.Equal(suite.T(), "Invalid redirect URI", err.ErrorDescription)
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestValidateAuthorizationCode_EmptyRedirectURIInCode() {
	// Test when authorization code has empty redirect URI (valid scenario)
	authzCodeWithEmptyURI := suite.testAuthzCode
	authzCodeWithEmptyURI.RedirectURI = ""

	tokenReq := &model.TokenRequest{
		ClientID:    "test-client-id",
		RedirectURI: "https://any.example.com/callback",
	}

	err := validateAuthorizationCode(tokenReq, authzCodeWithEmptyURI)
	assert.Nil(suite.T(), err)
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestValidateAuthorizationCode_InactiveCode() {
	inactiveCode := suite.testAuthzCode
	inactiveCode.State = authzconstants.AuthCodeStateInactive

	err := validateAuthorizationCode(suite.testTokenReq, inactiveCode)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorInvalidGrant, err.Error)
	assert.Equal(suite.T(), "Inactive authorization code", err.ErrorDescription)
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestValidateAuthorizationCode_InvalidState() {
	invalidStateCode := suite.testAuthzCode
	invalidStateCode.State = "INVALID_STATE"

	err := validateAuthorizationCode(suite.testTokenReq, invalidStateCode)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorInvalidGrant, err.Error)
	assert.Equal(suite.T(), "Inactive authorization code", err.ErrorDescription)
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestValidateAuthorizationCode_ExpiredCode() {
	expiredCode := suite.testAuthzCode
	expiredCode.ExpiryTime = time.Now().Add(-5 * time.Minute) // Expired

	err := validateAuthorizationCode(suite.testTokenReq, expiredCode)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorInvalidGrant, err.Error)
	assert.Equal(suite.T(), "Expired authorization code", err.ErrorDescription)
}
