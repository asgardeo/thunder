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
	"github.com/asgardeo/thunder/internal/oauth/oauth2/authz"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/model"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/user"
	"github.com/asgardeo/thunder/tests/mocks/jwtmock"
	"github.com/asgardeo/thunder/tests/mocks/oauth/oauth2/authzmock"
	usersvcmock "github.com/asgardeo/thunder/tests/mocks/usermock"
)

type AuthorizationCodeGrantHandlerTestSuite struct {
	suite.Suite
	handler          *authorizationCodeGrantHandler
	mockJWTService   *jwtmock.JWTServiceInterfaceMock
	mockAuthzService *authzmock.AuthorizeServiceInterfaceMock
	mockUserService  *usersvcmock.UserServiceInterfaceMock
	oauthApp         *appmodel.OAuthAppConfigProcessedDTO
	testAuthzCode    authz.AuthorizationCode
	testTokenReq     *model.TokenRequest
}

func TestAuthorizationCodeGrantHandlerSuite(t *testing.T) {
	suite.Run(t, new(AuthorizationCodeGrantHandlerTestSuite))
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) SetupTest() {
	// Initialize Thunder Runtime config with basic test config
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			ValidityPeriod: 3600,
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)

	suite.mockJWTService = &jwtmock.JWTServiceInterfaceMock{}
	suite.mockAuthzService = &authzmock.AuthorizeServiceInterfaceMock{}
	suite.mockUserService = usersvcmock.NewUserServiceInterfaceMock(suite.T())

	suite.handler = &authorizationCodeGrantHandler{
		jwtService:   suite.mockJWTService,
		authzService: suite.mockAuthzService,
		userService:  suite.mockUserService,
	}

	suite.oauthApp = &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:                "test-client-id",
		HashedClientSecret:      "hashed-secret",
		RedirectURIs:            []string{"https://client.example.com/callback"},
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
		ResponseTypes:           []constants.ResponseType{constants.ResponseTypeCode},
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
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

	suite.testAuthzCode = authz.AuthorizationCode{
		CodeID:           "test-code-id",
		Code:             "test-auth-code",
		ClientID:         "test-client-id",
		RedirectURI:      "https://client.example.com/callback",
		AuthorizedUserID: "test-user-id",
		TimeCreated:      time.Now().Add(-5 * time.Minute),
		ExpiryTime:       time.Now().Add(5 * time.Minute),
		Scopes:           "read write",
		State:            authz.AuthCodeStateActive,
	}
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestNewAuthorizationCodeGrantHandler() {
	handler := newAuthorizationCodeGrantHandler(suite.mockJWTService, suite.mockUserService, suite.mockAuthzService)
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
	suite.mockAuthzService.On("GetAuthorizationCodeDetails", "test-client-id", "test-auth-code").
		Return(&suite.testAuthzCode, nil)

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

	// Create token request with matching resource
	tokenReqWithResource := *suite.testTokenReq
	tokenReqWithResource.Resource = testResourceURL

	result, err := suite.handler.HandleGrant(&tokenReqWithResource, suite.oauthApp, ctx)

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

	suite.mockAuthzService.AssertExpectations(suite.T())
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestHandleGrant_InvalidAuthorizationCode() {
	// Mock authorization code store to return error
	suite.mockAuthzService.On("GetAuthorizationCodeDetails", "test-client-id", "test-auth-code").
		Return(nil, errors.New("invalid authorization code"))

	ctx := &model.TokenContext{
		TokenAttributes: make(map[string]interface{}),
	}

	// Create token request with matching resource
	tokenReqWithResource := *suite.testTokenReq
	tokenReqWithResource.Resource = testResourceURL

	result, err := suite.handler.HandleGrant(&tokenReqWithResource, suite.oauthApp, ctx)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorInvalidGrant, err.Error)
	assert.Equal(suite.T(), "Invalid authorization code", err.ErrorDescription)

	suite.mockAuthzService.AssertExpectations(suite.T())
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestHandleGrant_JWTGenerationError() {
	// Mock authorization code store to return valid code
	suite.mockAuthzService.On("GetAuthorizationCodeDetails", "test-client-id", "test-auth-code").
		Return(&suite.testAuthzCode, nil)

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

	// Create token request with matching resource
	tokenReqWithResource := *suite.testTokenReq
	tokenReqWithResource.Resource = testResourceURL

	result, err := suite.handler.HandleGrant(&tokenReqWithResource, suite.oauthApp, ctx)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorServerError, err.Error)
	assert.Equal(suite.T(), "Failed to generate token", err.ErrorDescription)

	suite.mockAuthzService.AssertExpectations(suite.T())
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestHandleGrant_EmptyScopes() {
	// Test with empty scopes
	authzCodeWithEmptyScopes := suite.testAuthzCode
	authzCodeWithEmptyScopes.Scopes = ""

	suite.mockAuthzService.On("GetAuthorizationCodeDetails", "test-client-id", "test-auth-code").
		Return(&authzCodeWithEmptyScopes, nil)

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

	// Create token request with matching resource
	tokenReqWithResource := *suite.testTokenReq
	tokenReqWithResource.Resource = testResourceURL

	result, err := suite.handler.HandleGrant(&tokenReqWithResource, suite.oauthApp, ctx)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Empty(suite.T(), result.AccessToken.Scopes)

	suite.mockAuthzService.AssertExpectations(suite.T())
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestHandleGrant_NilTokenAttributes() {
	// Test with nil token attributes
	suite.mockAuthzService.On("GetAuthorizationCodeDetails", "test-client-id", "test-auth-code").
		Return(&suite.testAuthzCode, nil)

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

	// Create token request with matching resource
	tokenReqWithResource := *suite.testTokenReq
	tokenReqWithResource.Resource = testResourceURL

	result, err := suite.handler.HandleGrant(&tokenReqWithResource, suite.oauthApp, ctx)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)

	// Should have initialized TokenAttributes
	assert.NotNil(suite.T(), ctx.TokenAttributes)
	assert.Equal(suite.T(), "test-user-id", ctx.TokenAttributes["sub"])
	assert.Equal(suite.T(), "test-client-id", ctx.TokenAttributes["aud"])

	suite.mockAuthzService.AssertExpectations(suite.T())
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
	inactiveCode.State = authz.AuthCodeStateInactive

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

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestHandleGrant_WithGroups() {
	testCases := []struct {
		name                 string
		includeInAccessToken bool
		includeInIDToken     bool
		includeOpenIDScope   bool
		scopeClaimsForGroups bool
		expectedGroups       []string
		mockGroups           []user.UserGroup
		description          string
	}{
		{
			name:                 "Groups in access token with ID token config",
			includeInAccessToken: true,
			includeInIDToken:     true,
			includeOpenIDScope:   false,
			scopeClaimsForGroups: false,
			expectedGroups:       []string{"Admin", "Users"},
			mockGroups: []user.UserGroup{
				{ID: "group1", Name: "Admin"},
				{ID: "group2", Name: "Users"},
			},
			description: "Should include groups in access token when configured (IDToken config " +
				"present but openid scope not requested)",
		},
		{
			name:                 "Groups in both access and ID tokens",
			includeInAccessToken: true,
			includeInIDToken:     true,
			includeOpenIDScope:   true,
			scopeClaimsForGroups: true,
			expectedGroups:       []string{"Admin", "Users"},
			mockGroups: []user.UserGroup{
				{ID: "group1", Name: "Admin"},
				{ID: "group2", Name: "Users"},
			},
			description: "Should include groups in both tokens when configured with openid scope and scope claims",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Reset mocks for each test case
			suite.mockAuthzService = &authzmock.AuthorizeServiceInterfaceMock{}
			suite.mockUserService = usersvcmock.NewUserServiceInterfaceMock(suite.T())
			suite.mockJWTService = &jwtmock.JWTServiceInterfaceMock{}
			suite.handler = &authorizationCodeGrantHandler{
				jwtService:   suite.mockJWTService,
				authzService: suite.mockAuthzService,
				userService:  suite.mockUserService,
			}

			accessTokenAttrs := []string{"email", "username"}
			if tc.includeInAccessToken {
				accessTokenAttrs = append(accessTokenAttrs, "groups")
			}
			var idTokenConfig *appmodel.IDTokenConfig
			if tc.includeInIDToken {
				if tc.scopeClaimsForGroups {
					// Include groups in ID token config with scope claims mapping
					idTokenConfig = &appmodel.IDTokenConfig{
						UserAttributes: []string{"email", "username", "groups"},
						ScopeClaims: map[string][]string{
							"openid": {"email", "username", "groups"},
						},
					}
				} else {
					idTokenConfig = &appmodel.IDTokenConfig{
						UserAttributes: []string{"email", "username"},
					}
				}
			}

			oauthAppWithGroups := &appmodel.OAuthAppConfigProcessedDTO{
				ClientID:                "test-client-id",
				HashedClientSecret:      "hashed-secret",
				RedirectURIs:            []string{"https://client.example.com/callback"},
				GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
				ResponseTypes:           []constants.ResponseType{constants.ResponseTypeCode},
				TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
				Token: &appmodel.OAuthTokenConfig{
					AccessToken: &appmodel.TokenConfig{
						UserAttributes: accessTokenAttrs,
					},
					IDToken: idTokenConfig,
				},
			}

			authzCode := suite.testAuthzCode
			if tc.includeOpenIDScope {
				authzCode.Scopes = "openid read write"
			}

			suite.mockAuthzService.On("GetAuthorizationCodeDetails", "test-client-id", "test-auth-code").
				Return(&authzCode, nil)

			mockUser := &user.User{
				ID:         "test-user-id",
				Attributes: json.RawMessage(`{"email":"test@example.com","username":"testuser"}`),
			}
			suite.mockUserService.On("GetUser", "test-user-id").Return(mockUser, nil)

			mockGroups := &user.UserGroupListResponse{
				TotalResults: len(tc.mockGroups),
				StartIndex:   0,
				Count:        len(tc.mockGroups),
				Groups:       tc.mockGroups,
			}
			suite.mockUserService.On("GetUserGroups", "test-user-id", DefaultGroupListLimit, 0).
				Return(mockGroups, nil)

			var capturedAccessTokenClaims map[string]interface{}
			var capturedIDTokenClaims map[string]interface{}

			// Mock access token generation
			suite.mockJWTService.On("GenerateJWT", "test-user-id", "test-client-id",
				mock.AnythingOfType("string"), mock.AnythingOfType("int64"),
				mock.AnythingOfType("map[string]interface {}")).
				Run(func(args mock.Arguments) {
					capturedAccessTokenClaims = args.Get(4).(map[string]interface{})
				}).
				Return("test-jwt-token", int64(3600), nil).Once()

			// Mock ID token generation if openid scope is present
			if tc.includeOpenIDScope {
				suite.mockJWTService.On("GenerateJWT", "test-user-id", "test-client-id",
					mock.AnythingOfType("string"), mock.AnythingOfType("int64"),
					mock.AnythingOfType("map[string]interface {}")).
					Run(func(args mock.Arguments) {
						capturedIDTokenClaims = args.Get(4).(map[string]interface{})
					}).
					Return("test-id-token", int64(3600), nil).Once()
			}

			ctx := &model.TokenContext{
				TokenAttributes: make(map[string]interface{}),
			}

			result, err := suite.handler.HandleGrant(suite.testTokenReq, oauthAppWithGroups, ctx)

			assert.Nil(suite.T(), err, tc.description)
			assert.NotNil(suite.T(), result, tc.description)

			// Verify access token groups
			if tc.includeInAccessToken {
				assert.NotNil(suite.T(), capturedAccessTokenClaims["groups"], tc.description)
				groupsInClaims, ok := capturedAccessTokenClaims["groups"].([]string)
				assert.True(suite.T(), ok, tc.description)
				assert.Equal(suite.T(), tc.expectedGroups, groupsInClaims, tc.description)

				assert.NotNil(suite.T(), result.AccessToken.UserAttributes["groups"], tc.description)
				groupsInAttrs, ok := result.AccessToken.UserAttributes["groups"].([]string)
				assert.True(suite.T(), ok, tc.description)
				assert.Equal(suite.T(), tc.expectedGroups, groupsInAttrs, tc.description)
			} else {
				assert.Nil(suite.T(), capturedAccessTokenClaims["groups"], tc.description)
				assert.Nil(suite.T(), result.AccessToken.UserAttributes["groups"], tc.description)
			}

			// Verify ID token groups
			if tc.includeInIDToken && tc.includeOpenIDScope && tc.scopeClaimsForGroups {
				assert.NotNil(suite.T(), result.IDToken.Token, tc.description)
				assert.NotNil(suite.T(), capturedIDTokenClaims["groups"], tc.description)
				groupsInIDToken, ok := capturedIDTokenClaims["groups"].([]string)
				assert.True(suite.T(), ok, tc.description)
				assert.Equal(suite.T(), tc.expectedGroups, groupsInIDToken, tc.description)
			} else if tc.includeOpenIDScope {
				assert.NotNil(suite.T(), result.IDToken.Token, tc.description)
			} else {
				assert.Empty(suite.T(), result.IDToken.Token, tc.description)
			}

			suite.mockAuthzService.AssertExpectations(suite.T())
			suite.mockUserService.AssertExpectations(suite.T())
			suite.mockJWTService.AssertExpectations(suite.T())
		})
	}
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestHandleGrant_WithEmptyGroups() {
	testCases := []struct {
		name                 string
		includeInAccessToken bool
		includeInIDToken     bool
		includeOpenIDScope   bool
		scopeClaimsForGroups bool
		description          string
	}{
		{
			name:                 "Empty groups in access token",
			includeInAccessToken: true,
			includeInIDToken:     true,
			includeOpenIDScope:   false,
			scopeClaimsForGroups: false,
			description:          "Should not include groups claim in access token when user has no groups",
		},
		{
			name:                 "Empty groups with both tokens",
			includeInAccessToken: true,
			includeInIDToken:     true,
			includeOpenIDScope:   true,
			scopeClaimsForGroups: true,
			description:          "Should not include groups claim in either token when user has no groups",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.mockAuthzService = &authzmock.AuthorizeServiceInterfaceMock{}
			suite.mockUserService = usersvcmock.NewUserServiceInterfaceMock(suite.T())
			suite.mockJWTService = &jwtmock.JWTServiceInterfaceMock{}
			suite.handler = &authorizationCodeGrantHandler{
				jwtService:   suite.mockJWTService,
				authzService: suite.mockAuthzService,
				userService:  suite.mockUserService,
			}

			accessTokenAttrs := []string{"email", "username"}
			if tc.includeInAccessToken {
				accessTokenAttrs = append(accessTokenAttrs, "groups")
			}
			var idTokenConfig *appmodel.IDTokenConfig
			if tc.includeInIDToken {
				if tc.scopeClaimsForGroups {
					idTokenConfig = &appmodel.IDTokenConfig{
						UserAttributes: []string{"email", "username", "groups"},
						ScopeClaims: map[string][]string{
							"openid": {"email", "username", "groups"},
						},
					}
				} else {
					idTokenConfig = &appmodel.IDTokenConfig{
						UserAttributes: []string{"email", "username"},
					}
				}
			}

			oauthAppWithGroups := &appmodel.OAuthAppConfigProcessedDTO{
				ClientID:                "test-client-id",
				HashedClientSecret:      "hashed-secret",
				RedirectURIs:            []string{"https://client.example.com/callback"},
				GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
				ResponseTypes:           []constants.ResponseType{constants.ResponseTypeCode},
				TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
				Token: &appmodel.OAuthTokenConfig{
					AccessToken: &appmodel.TokenConfig{
						UserAttributes: accessTokenAttrs,
					},
					IDToken: idTokenConfig,
				},
			}

			authzCode := suite.testAuthzCode
			if tc.includeOpenIDScope {
				authzCode.Scopes = "openid read write"
			}

			suite.mockAuthzService.On("GetAuthorizationCodeDetails", "test-client-id", "test-auth-code").
				Return(&authzCode, nil)

			mockUser := &user.User{
				ID:         "test-user-id",
				Attributes: json.RawMessage(`{"email":"test@example.com","username":"testuser"}`),
			}
			suite.mockUserService.On("GetUser", "test-user-id").Return(mockUser, nil)

			mockGroups := &user.UserGroupListResponse{
				TotalResults: 0,
				StartIndex:   0,
				Count:        0,
				Groups:       []user.UserGroup{}, // Empty groups
			}
			suite.mockUserService.On("GetUserGroups", "test-user-id", DefaultGroupListLimit, 0).
				Return(mockGroups, nil)

			var capturedAccessTokenClaims map[string]interface{}
			var capturedIDTokenClaims map[string]interface{}

			// Mock access token generation
			suite.mockJWTService.On("GenerateJWT", "test-user-id", "test-client-id",
				mock.AnythingOfType("string"), mock.AnythingOfType("int64"),
				mock.AnythingOfType("map[string]interface {}")).
				Run(func(args mock.Arguments) {
					capturedAccessTokenClaims = args.Get(4).(map[string]interface{})
				}).
				Return("test-jwt-token", int64(3600), nil).Once()

			// Mock ID token generation if openid scope is present
			if tc.includeOpenIDScope {
				suite.mockJWTService.On("GenerateJWT", "test-user-id", "test-client-id",
					mock.AnythingOfType("string"), mock.AnythingOfType("int64"),
					mock.AnythingOfType("map[string]interface {}")).
					Run(func(args mock.Arguments) {
						capturedIDTokenClaims = args.Get(4).(map[string]interface{})
					}).
					Return("test-id-token", int64(3600), nil).Once()
			}

			ctx := &model.TokenContext{
				TokenAttributes: make(map[string]interface{}),
			}

			result, err := suite.handler.HandleGrant(suite.testTokenReq, oauthAppWithGroups, ctx)

			assert.Nil(suite.T(), err, tc.description)
			assert.NotNil(suite.T(), result, tc.description)

			assert.Nil(suite.T(), capturedAccessTokenClaims["groups"], tc.description)
			assert.Nil(suite.T(), result.AccessToken.UserAttributes["groups"], tc.description)

			// Verify ID token
			if tc.includeOpenIDScope {
				assert.NotNil(suite.T(), result.IDToken.Token, tc.description)
				assert.Nil(suite.T(), capturedIDTokenClaims["groups"], tc.description)
			} else {
				assert.Empty(suite.T(), result.IDToken.Token, tc.description)
			}

			suite.mockAuthzService.AssertExpectations(suite.T())
			suite.mockUserService.AssertExpectations(suite.T())
			suite.mockJWTService.AssertExpectations(suite.T())
		})
	}
}

// Resource Parameter Tests (RFC 8707)

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestHandleGrant_WithResourceParameterInAuthCode() {
	// Set up auth code with resource parameter
	authCodeWithResource := suite.testAuthzCode
	authCodeWithResource.Resource = testResourceURL

	suite.mockAuthzService.On("GetAuthorizationCodeDetails", "test-client-id", "test-auth-code").
		Return(&authCodeWithResource, nil)

	// Mock user service to return user
	mockUser := &user.User{
		ID:         "test-user-id",
		Attributes: json.RawMessage(`{"email":"test@example.com","username":"testuser"}`),
	}
	suite.mockUserService.On("GetUser", "test-user-id").Return(mockUser, nil)

	capturedClaims := map[string]interface{}{}
	suite.mockJWTService.On("GenerateJWT", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.MatchedBy(func(claims map[string]interface{}) bool {
			for k, v := range claims {
				capturedClaims[k] = v
			}
			return true
		})).Return("mock-jwt-token", int64(12345), nil)

	ctx := &model.TokenContext{
		TokenAttributes: make(map[string]interface{}),
	}

	// Create token request with matching resource
	tokenReqWithResource := *suite.testTokenReq
	tokenReqWithResource.Resource = testResourceURL

	result, err := suite.handler.HandleGrant(&tokenReqWithResource, suite.oauthApp, ctx)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)

	// Verify resource was included in audience claim
	assert.NotNil(suite.T(), capturedClaims["aud"])
	assert.Equal(suite.T(), testResourceURL, capturedClaims["aud"])
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestHandleGrant_ResourceParameterMismatch() {
	// Set up auth code with different resource than token request
	authCodeWithResource := suite.testAuthzCode
	authCodeWithResource.Resource = "https://api.example.com/resource"

	suite.mockAuthzService.On("GetAuthorizationCodeDetails", "test-client-id", "test-auth-code").
		Return(&authCodeWithResource, nil)

	// Create token request with different resource
	tokenReqWithResource := *suite.testTokenReq
	tokenReqWithResource.Resource = testResourceURL

	ctx := &model.TokenContext{
		TokenAttributes: make(map[string]interface{}),
	}

	result, err := suite.handler.HandleGrant(&tokenReqWithResource, suite.oauthApp, ctx)

	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), constants.ErrorInvalidTarget, err.Error)
	assert.Equal(suite.T(), "Resource parameter mismatch", err.ErrorDescription)
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestHandleGrant_ResourceParameterMatch() {
	// Set up auth code with resource parameter
	authCodeWithResource := suite.testAuthzCode
	authCodeWithResource.Resource = testResourceURL

	suite.mockAuthzService.On("GetAuthorizationCodeDetails", "test-client-id", "test-auth-code").
		Return(&authCodeWithResource, nil)

	// Mock user service to return user
	mockUser := &user.User{
		ID:         "test-user-id",
		Attributes: json.RawMessage(`{"email":"test@example.com","username":"testuser"}`),
	}
	suite.mockUserService.On("GetUser", "test-user-id").Return(mockUser, nil)

	capturedClaims := map[string]interface{}{}
	suite.mockJWTService.On("GenerateJWT", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.MatchedBy(func(claims map[string]interface{}) bool {
			for k, v := range claims {
				capturedClaims[k] = v
			}
			return true
		})).Return("mock-jwt-token", int64(12345), nil)

	// Create token request with matching resource
	tokenReqWithResource := *suite.testTokenReq
	tokenReqWithResource.Resource = testResourceURL

	ctx := &model.TokenContext{
		TokenAttributes: make(map[string]interface{}),
	}

	result, err := suite.handler.HandleGrant(&tokenReqWithResource, suite.oauthApp, ctx)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), testResourceURL, capturedClaims["aud"])
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestHandleGrant_NoResourceParameter() {
	// Auth code without resource parameter
	suite.mockAuthzService.On("GetAuthorizationCodeDetails", "test-client-id", "test-auth-code").
		Return(&suite.testAuthzCode, nil)

	// Mock user service to return user
	mockUser := &user.User{
		ID:         "test-user-id",
		Attributes: json.RawMessage(`{"email":"test@example.com","username":"testuser"}`),
	}
	suite.mockUserService.On("GetUser", "test-user-id").Return(mockUser, nil)

	capturedClaims := map[string]interface{}{}
	suite.mockJWTService.On("GenerateJWT", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.MatchedBy(func(claims map[string]interface{}) bool {
			for k, v := range claims {
				capturedClaims[k] = v
			}
			return true
		})).Return("mock-jwt-token", int64(12345), nil)

	ctx := &model.TokenContext{
		TokenAttributes: make(map[string]interface{}),
	}

	// Create token request with matching resource
	tokenReqWithResource := *suite.testTokenReq
	tokenReqWithResource.Resource = testResourceURL

	result, err := suite.handler.HandleGrant(&tokenReqWithResource, suite.oauthApp, ctx)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)

	// Verify no audience claim
	assert.Nil(suite.T(), capturedClaims["aud"])
}
