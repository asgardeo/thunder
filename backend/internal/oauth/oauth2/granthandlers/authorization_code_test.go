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
	"github.com/asgardeo/thunder/internal/oauth/oauth2/tokenservice"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/user"
	"github.com/asgardeo/thunder/tests/mocks/jwtmock"
	"github.com/asgardeo/thunder/tests/mocks/oauth/oauth2/authzmock"
	"github.com/asgardeo/thunder/tests/mocks/oauth/oauth2/tokenservicemock"
	usersvcmock "github.com/asgardeo/thunder/tests/mocks/usermock"
)

const (
	oidcReadWriteScopes         = "openid read write"
	testCodeChallenge           = "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
	testCodeChallengeMethodS256 = "S256"
	testClientCallbackURL       = "https://client.example.com/callback"
)

// convertToStringSlice converts groups from various formats to []string for testing.
func convertToStringSlice(groups interface{}) []string {
	if groups == nil {
		return nil
	}
	switch v := groups.(type) {
	case []string:
		return v
	case []interface{}:
		result := make([]string, 0, len(v))
		for _, item := range v {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	default:
		return nil
	}
}

type AuthorizationCodeGrantHandlerTestSuite struct {
	suite.Suite
	handler          *authorizationCodeGrantHandler
	mockJWTService   *jwtmock.JWTServiceInterfaceMock
	mockTokenBuilder *tokenservicemock.TokenBuilderInterfaceMock
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

	suite.mockJWTService = jwtmock.NewJWTServiceInterfaceMock(suite.T())
	suite.mockTokenBuilder = tokenservicemock.NewTokenBuilderInterfaceMock(suite.T())
	suite.mockAuthzService = authzmock.NewAuthorizeServiceInterfaceMock(suite.T())
	suite.mockUserService = usersvcmock.NewUserServiceInterfaceMock(suite.T())

	suite.handler = &authorizationCodeGrantHandler{
		tokenBuilder: suite.mockTokenBuilder,
		authzService: suite.mockAuthzService,
		userService:  suite.mockUserService,
	}

	suite.oauthApp = &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:                testClientID,
		HashedClientSecret:      "hashed-secret",
		RedirectURIs:            []string{"https://client.example.com/callback"},
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
		ResponseTypes:           []constants.ResponseType{constants.ResponseTypeCode},
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
		Token: &appmodel.OAuthTokenConfig{
			AccessToken: &appmodel.AccessTokenConfig{
				UserAttributes: []string{"email", "username"},
			},
		},
	}

	suite.testTokenReq = &model.TokenRequest{
		GrantType:   string(constants.GrantTypeAuthorizationCode),
		ClientID:    testClientID,
		Code:        "test-auth-code",
		RedirectURI: "https://client.example.com/callback",
	}

	suite.testAuthzCode = authz.AuthorizationCode{
		CodeID:           "test-code-id",
		Code:             "test-auth-code",
		ClientID:         testClientID,
		RedirectURI:      "https://client.example.com/callback",
		AuthorizedUserID: testUserID,
		UserAttributes: map[string]interface{}{
			"email":    "test@example.com",
			"username": "testuser",
		},
		TimeCreated: time.Now().Add(-5 * time.Minute),
		ExpiryTime:  time.Now().Add(5 * time.Minute),
		Scopes:      "read write",
		State:       authz.AuthCodeStateActive,
	}
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestNewAuthorizationCodeGrantHandler() {
	handler := newAuthorizationCodeGrantHandler(suite.mockUserService,
		suite.mockAuthzService, suite.mockTokenBuilder)
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
		ClientID:  testClientID,
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
		ClientID:  testClientID,
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
		ClientID:  testClientID,
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
		ClientID:    testClientID,
		Code:        "test-code",
		RedirectURI: "", // Missing redirect URI
	}

	err := suite.handler.ValidateGrant(tokenReq, suite.oauthApp)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, err.Error)
	assert.Equal(suite.T(), "Redirect URI is required", err.ErrorDescription)
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestHandleGrant_Success() {
	// Create authorization code with resource
	authCodeWithResource := suite.testAuthzCode
	authCodeWithResource.Resource = testResourceURL

	// Mock authorization code store to return valid code with resource
	suite.mockAuthzService.On("GetAuthorizationCodeDetails", testClientID, "test-auth-code").
		Return(&authCodeWithResource, nil)

	// Mock token builder to generate access token
	suite.mockTokenBuilder.On("BuildAccessToken", mock.MatchedBy(func(ctx *tokenservice.AccessTokenBuildContext) bool {
		return ctx.Subject == testUserID &&
			ctx.Audience == testResourceURL &&
			ctx.ClientID == testClientID &&
			ctx.GrantType == string(constants.GrantTypeAuthorizationCode)
	})).Return(&model.TokenDTO{
		Token:     "test-jwt-token",
		TokenType: constants.TokenTypeBearer,
		IssuedAt:  time.Now().Unix(),
		ExpiresIn: 3600,
		Scopes:    []string{"read", "write"},
		ClientID:  testClientID,
		Subject:   testUserID,
		Audience:  testResourceURL,
	}, nil)

	// Create token request with matching resource
	tokenReqWithResource := *suite.testTokenReq
	tokenReqWithResource.Resource = testResourceURL

	result, err := suite.handler.HandleGrant(&tokenReqWithResource, suite.oauthApp)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "test-jwt-token", result.AccessToken.Token)
	assert.Equal(suite.T(), constants.TokenTypeBearer, result.AccessToken.TokenType)
	assert.Equal(suite.T(), int64(3600), result.AccessToken.ExpiresIn)
	assert.Equal(suite.T(), []string{"read", "write"}, result.AccessToken.Scopes)
	assert.Equal(suite.T(), testClientID, result.AccessToken.ClientID)

	// Check token attributes
	assert.Equal(suite.T(), testUserID, result.AccessToken.Subject)
	assert.Equal(suite.T(), testResourceURL, result.AccessToken.Audience)

	suite.mockAuthzService.AssertExpectations(suite.T())
	suite.mockTokenBuilder.AssertExpectations(suite.T())
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestHandleGrant_InvalidAuthorizationCode() {
	// Mock authorization code store to return error
	suite.mockAuthzService.On("GetAuthorizationCodeDetails", testClientID, "test-auth-code").
		Return(nil, errors.New("invalid authorization code"))

	// Create token request with matching resource
	tokenReqWithResource := *suite.testTokenReq
	tokenReqWithResource.Resource = testResourceURL

	result, err := suite.handler.HandleGrant(&tokenReqWithResource, suite.oauthApp)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorInvalidGrant, err.Error)
	assert.Equal(suite.T(), "Invalid authorization code", err.ErrorDescription)

	suite.mockAuthzService.AssertExpectations(suite.T())
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestHandleGrant_JWTGenerationError() {
	// Mock authorization code store to return valid code
	suite.mockAuthzService.On("GetAuthorizationCodeDetails", testClientID, "test-auth-code").
		Return(&suite.testAuthzCode, nil)

	// Mock token builder to fail token generation
	suite.mockTokenBuilder.On("BuildAccessToken", mock.Anything).Return(nil, errors.New("jwt generation failed"))

	// Create token request with matching resource
	tokenReqWithResource := *suite.testTokenReq
	tokenReqWithResource.Resource = testResourceURL

	result, err := suite.handler.HandleGrant(&tokenReqWithResource, suite.oauthApp)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorServerError, err.Error)
	assert.Equal(suite.T(), "Failed to generate token", err.ErrorDescription)

	suite.mockAuthzService.AssertExpectations(suite.T())
	suite.mockTokenBuilder.AssertExpectations(suite.T())
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestHandleGrant_EmptyScopes() {
	// Test with empty scopes
	authzCodeWithEmptyScopes := suite.testAuthzCode
	authzCodeWithEmptyScopes.Scopes = ""

	suite.mockAuthzService.On("GetAuthorizationCodeDetails", testClientID, "test-auth-code").
		Return(&authzCodeWithEmptyScopes, nil)

	suite.mockTokenBuilder.On("BuildAccessToken", mock.Anything).Return(&model.TokenDTO{
		Token:     "test-jwt-token",
		TokenType: constants.TokenTypeBearer,
		IssuedAt:  time.Now().Unix(),
		ExpiresIn: 3600,
		Scopes:    []string{},
		ClientID:  testClientID,
	}, nil)

	// Create token request with matching resource
	tokenReqWithResource := *suite.testTokenReq
	tokenReqWithResource.Resource = testResourceURL

	result, err := suite.handler.HandleGrant(&tokenReqWithResource, suite.oauthApp)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Empty(suite.T(), result.AccessToken.Scopes)

	suite.mockAuthzService.AssertExpectations(suite.T())
	suite.mockTokenBuilder.AssertExpectations(suite.T())
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestHandleGrant_NilTokenAttributes() {
	// Test with nil token attributes
	suite.mockAuthzService.On("GetAuthorizationCodeDetails", testClientID, "test-auth-code").
		Return(&suite.testAuthzCode, nil)

	suite.mockTokenBuilder.On("BuildAccessToken", mock.Anything).Return(&model.TokenDTO{
		Token:     "test-jwt-token",
		TokenType: constants.TokenTypeBearer,
		IssuedAt:  time.Now().Unix(),
		ExpiresIn: 3600,
		Scopes:    []string{"read", "write"},
		ClientID:  testClientID,
		Subject:   testUserID,
		Audience:  testClientID,
	}, nil)

	// Create token request with matching resource
	tokenReqWithResource := *suite.testTokenReq
	tokenReqWithResource.Resource = testResourceURL

	result, err := suite.handler.HandleGrant(&tokenReqWithResource, suite.oauthApp)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)

	// Check token attributes
	assert.Equal(suite.T(), testUserID, result.AccessToken.Subject)
	assert.Equal(suite.T(), testClientID, result.AccessToken.Audience)

	suite.mockAuthzService.AssertExpectations(suite.T())
	suite.mockTokenBuilder.AssertExpectations(suite.T())
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
		ClientID:    testClientID,
		RedirectURI: "https://wrong.example.com/callback", // Wrong redirect URI
	}

	err := validateAuthorizationCode(invalidTokenReq, suite.testAuthzCode)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorInvalidGrant, err.Error)
	assert.Equal(suite.T(), "Invalid redirect URI", err.ErrorDescription)
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestValidateAuthorizationCode_EmptyRedirectURIInCode() {
	authzCodeWithEmptyURI := suite.testAuthzCode
	authzCodeWithEmptyURI.RedirectURI = ""

	tokenReq := &model.TokenRequest{
		ClientID:    testClientID,
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
			suite.mockAuthzService = authzmock.NewAuthorizeServiceInterfaceMock(suite.T())
			suite.mockUserService = usersvcmock.NewUserServiceInterfaceMock(suite.T())
			suite.mockJWTService = jwtmock.NewJWTServiceInterfaceMock(suite.T())
			suite.mockTokenBuilder = tokenservicemock.NewTokenBuilderInterfaceMock(suite.T())
			suite.handler = &authorizationCodeGrantHandler{
				tokenBuilder: suite.mockTokenBuilder,
				authzService: suite.mockAuthzService,
				userService:  suite.mockUserService,
			}

			accessTokenAttrs := []string{"email", "username"}
			if tc.includeInAccessToken {
				accessTokenAttrs = append(accessTokenAttrs, constants.UserAttributeGroups)
			}
			var idTokenConfig *appmodel.IDTokenConfig
			if tc.includeInIDToken {
				if tc.scopeClaimsForGroups {
					// Include groups in ID token config with scope claims mapping
					idTokenConfig = &appmodel.IDTokenConfig{
						UserAttributes: []string{"email", "username", constants.UserAttributeGroups},
						ScopeClaims: map[string][]string{
							"openid": {"email", "username", constants.UserAttributeGroups},
						},
					}
				} else {
					idTokenConfig = &appmodel.IDTokenConfig{
						UserAttributes: []string{"email", "username"},
					}
				}
			}

			oauthAppWithGroups := &appmodel.OAuthAppConfigProcessedDTO{
				ClientID:                testClientID,
				HashedClientSecret:      "hashed-secret",
				RedirectURIs:            []string{"https://client.example.com/callback"},
				GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
				ResponseTypes:           []constants.ResponseType{constants.ResponseTypeCode},
				TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
				Token: &appmodel.OAuthTokenConfig{
					AccessToken: &appmodel.AccessTokenConfig{
						UserAttributes: accessTokenAttrs,
					},
					IDToken: idTokenConfig,
				},
			}

			authzCode := suite.testAuthzCode
			if tc.includeOpenIDScope {
				authzCode.Scopes = oidcReadWriteScopes
			}

			// Add user attributes to authz code (including groups from assertion)
			userGroups := make([]string, 0, len(tc.mockGroups))
			for _, group := range tc.mockGroups {
				userGroups = append(userGroups, group.Name)
			}
			authzCode.UserAttributes = map[string]interface{}{
				"email":    "test@example.com",
				"username": "testuser",
			}
			if len(userGroups) > 0 {
				authzCode.UserAttributes[constants.UserAttributeGroups] = userGroups
			}

			suite.mockAuthzService.On("GetAuthorizationCodeDetails", testClientID, "test-auth-code").
				Return(&authzCode, nil)

			// Groups come from the authorization code (extracted from assertion during authorization)
			// No need to fetch from DB

			var capturedAccessTokenClaims map[string]interface{}
			var capturedIDTokenClaims map[string]interface{}

			// Mock access token generation - use function return to access context at call time
			suite.mockTokenBuilder.On("BuildAccessToken",
				mock.MatchedBy(func(ctx *tokenservice.AccessTokenBuildContext) bool {
					// Capture user attributes and groups (simulate filtering that happens in BuildAccessToken)
					capturedAccessTokenClaims = make(map[string]interface{})
					for k, v := range ctx.UserAttributes {
						capturedAccessTokenClaims[k] = v
					}
					// Verify GrantType is authorization_code
					return ctx.GrantType == string(constants.GrantTypeAuthorizationCode)
				})).Return(func(ctx *tokenservice.AccessTokenBuildContext) (*model.TokenDTO, error) {
				// Simulate filtering that happens in BuildAccessToken
				userAttrs := make(map[string]interface{})
				for k, v := range ctx.UserAttributes {
					userAttrs[k] = v
				}
				return &model.TokenDTO{
					Token:          "test-jwt-token",
					TokenType:      constants.TokenTypeBearer,
					IssuedAt:       time.Now().Unix(),
					ExpiresIn:      3600,
					Scopes:         []string{"read", "write"},
					ClientID:       testClientID,
					UserAttributes: userAttrs,
				}, nil
			}).Once()

			// Mock ID token generation if openid scope is present
			if tc.includeOpenIDScope {
				suite.mockTokenBuilder.On("BuildIDToken",
					mock.MatchedBy(func(ctx *tokenservice.IDTokenBuildContext) bool {
						// Capture ID token claims
						capturedIDTokenClaims = make(map[string]interface{})
						for k, v := range ctx.UserAttributes {
							capturedIDTokenClaims[k] = v
						}
						// Groups are already in UserAttributes if configured
						// Token builder will extract and add them if needed
						return true
					})).Return(&model.TokenDTO{
					Token:     "test-id-token",
					TokenType: "",
					IssuedAt:  time.Now().Unix(),
					ExpiresIn: 3600,
					Scopes:    []string{"read", "write", "openid"},
					ClientID:  testClientID,
				}, nil).Once()
			}

			result, err := suite.handler.HandleGrant(suite.testTokenReq, oauthAppWithGroups)

			assert.Nil(suite.T(), err, tc.description)
			assert.NotNil(suite.T(), result, tc.description)

			// Verify access token groups
			if tc.includeInAccessToken {
				assert.NotNil(suite.T(), capturedAccessTokenClaims[constants.UserAttributeGroups], tc.description)
				groupsInClaims := convertToStringSlice(capturedAccessTokenClaims[constants.UserAttributeGroups])
				assert.Equal(suite.T(), tc.expectedGroups, groupsInClaims, tc.description)

				assert.NotNil(suite.T(),
					result.AccessToken.UserAttributes[constants.UserAttributeGroups], tc.description)
				groupsInAttrs := convertToStringSlice(result.AccessToken.UserAttributes[constants.UserAttributeGroups])
				assert.Equal(suite.T(), tc.expectedGroups, groupsInAttrs, tc.description)
			} else {
				assert.Nil(suite.T(), capturedAccessTokenClaims[constants.UserAttributeGroups], tc.description)
				assert.Nil(suite.T(), result.AccessToken.UserAttributes[constants.UserAttributeGroups], tc.description)
			}

			// Verify ID token groups
			if tc.includeInIDToken && tc.includeOpenIDScope && tc.scopeClaimsForGroups {
				assert.NotNil(suite.T(), result.IDToken.Token, tc.description)
				assert.NotNil(suite.T(), capturedIDTokenClaims[constants.UserAttributeGroups], tc.description)
				groupsInIDToken := convertToStringSlice(capturedIDTokenClaims[constants.UserAttributeGroups])
				assert.Equal(suite.T(), tc.expectedGroups, groupsInIDToken, tc.description)
			} else if tc.includeOpenIDScope {
				assert.NotNil(suite.T(), result.IDToken.Token, tc.description)
			} else {
				assert.Empty(suite.T(), result.IDToken.Token, tc.description)
			}

			suite.mockAuthzService.AssertExpectations(suite.T())
			suite.mockUserService.AssertExpectations(suite.T())
			suite.mockTokenBuilder.AssertExpectations(suite.T())
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
			suite.mockAuthzService = authzmock.NewAuthorizeServiceInterfaceMock(suite.T())
			suite.mockUserService = usersvcmock.NewUserServiceInterfaceMock(suite.T())
			suite.mockJWTService = jwtmock.NewJWTServiceInterfaceMock(suite.T())
			suite.mockTokenBuilder = tokenservicemock.NewTokenBuilderInterfaceMock(suite.T())
			suite.handler = &authorizationCodeGrantHandler{
				tokenBuilder: suite.mockTokenBuilder,
				authzService: suite.mockAuthzService,
				userService:  suite.mockUserService,
			}

			accessTokenAttrs := []string{"email", "username"}
			if tc.includeInAccessToken {
				accessTokenAttrs = append(accessTokenAttrs, constants.UserAttributeGroups)
			}
			var idTokenConfig *appmodel.IDTokenConfig
			if tc.includeInIDToken {
				if tc.scopeClaimsForGroups {
					idTokenConfig = &appmodel.IDTokenConfig{
						UserAttributes: []string{"email", "username", constants.UserAttributeGroups},
						ScopeClaims: map[string][]string{
							"openid": {"email", "username", constants.UserAttributeGroups},
						},
					}
				} else {
					idTokenConfig = &appmodel.IDTokenConfig{
						UserAttributes: []string{"email", "username"},
					}
				}
			}

			oauthAppWithGroups := &appmodel.OAuthAppConfigProcessedDTO{
				ClientID:                testClientID,
				HashedClientSecret:      "hashed-secret",
				RedirectURIs:            []string{"https://client.example.com/callback"},
				GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
				ResponseTypes:           []constants.ResponseType{constants.ResponseTypeCode},
				TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
				Token: &appmodel.OAuthTokenConfig{
					AccessToken: &appmodel.AccessTokenConfig{
						UserAttributes: accessTokenAttrs,
					},
					IDToken: idTokenConfig,
				},
			}

			authzCode := suite.testAuthzCode
			if tc.includeOpenIDScope {
				authzCode.Scopes = oidcReadWriteScopes
			}

			// Add user attributes to authz code (groups will be empty array or not present)
			authzCode.UserAttributes = map[string]interface{}{
				"email":    "test@example.com",
				"username": "testuser",
			}
			// Empty groups - not added to UserAttributes (user has no groups)

			suite.mockAuthzService.On("GetAuthorizationCodeDetails", testClientID, "test-auth-code").
				Return(&authzCode, nil)

			// Groups come from the authorization code (extracted from assertion during authorization)
			// If groups are not in auth code, user has no groups (empty array)
			// No need to fetch from DB

			var capturedAccessTokenClaims map[string]interface{}
			var capturedIDTokenClaims map[string]interface{}

			// Mock access token generation - use function return to access context at call time
			suite.mockTokenBuilder.On("BuildAccessToken",
				mock.MatchedBy(func(ctx *tokenservice.AccessTokenBuildContext) bool {
					// Capture user attributes which contain groups
					capturedAccessTokenClaims = make(map[string]interface{})
					for k, v := range ctx.UserAttributes {
						capturedAccessTokenClaims[k] = v
					}
					// Verify GrantType is authorization_code
					return ctx.GrantType == string(constants.GrantTypeAuthorizationCode)
				})).Return(func(ctx *tokenservice.AccessTokenBuildContext) (*model.TokenDTO, error) {
				// Return user attributes from the actual call context
				userAttrs := make(map[string]interface{})
				for k, v := range ctx.UserAttributes {
					userAttrs[k] = v
				}
				return &model.TokenDTO{
					Token:          "test-jwt-token",
					TokenType:      constants.TokenTypeBearer,
					IssuedAt:       time.Now().Unix(),
					ExpiresIn:      3600,
					Scopes:         []string{"read", "write"},
					ClientID:       testClientID,
					UserAttributes: userAttrs,
				}, nil
			}).Once()

			// Mock ID token generation if openid scope is present
			if tc.includeOpenIDScope {
				suite.mockTokenBuilder.On("BuildIDToken",
					mock.MatchedBy(func(ctx *tokenservice.IDTokenBuildContext) bool {
						// Capture ID token claims from user attributes
						capturedIDTokenClaims = ctx.UserAttributes
						return true
					})).Return(&model.TokenDTO{
					Token:     "test-id-token",
					TokenType: "",
					IssuedAt:  time.Now().Unix(),
					ExpiresIn: 3600,
					Scopes:    []string{"read", "write", "openid"},
					ClientID:  testClientID,
				}, nil).Once()
			}

			result, err := suite.handler.HandleGrant(suite.testTokenReq, oauthAppWithGroups)

			assert.Nil(suite.T(), err, tc.description)
			assert.NotNil(suite.T(), result, tc.description)

			assert.Nil(suite.T(), capturedAccessTokenClaims[constants.UserAttributeGroups], tc.description)
			assert.Nil(suite.T(), result.AccessToken.UserAttributes[constants.UserAttributeGroups], tc.description)

			// Verify ID token
			if tc.includeOpenIDScope {
				assert.NotNil(suite.T(), result.IDToken.Token, tc.description)
				assert.Nil(suite.T(), capturedIDTokenClaims[constants.UserAttributeGroups], tc.description)
			} else {
				assert.Empty(suite.T(), result.IDToken.Token, tc.description)
			}

			suite.mockAuthzService.AssertExpectations(suite.T())
			suite.mockUserService.AssertExpectations(suite.T())
			suite.mockTokenBuilder.AssertExpectations(suite.T())
		})
	}
}

// Resource Parameter Tests (RFC 8707)

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestHandleGrant_ResourceParameterMismatch() {
	// Set up auth code with different resource than token request
	authCodeWithResource := suite.testAuthzCode
	authCodeWithResource.Resource = "https://api.example.com/resource"

	suite.mockAuthzService.On("GetAuthorizationCodeDetails", testClientID, "test-auth-code").
		Return(&authCodeWithResource, nil)

	// Create token request with different resource
	tokenReqWithResource := *suite.testTokenReq
	tokenReqWithResource.Resource = testResourceURL

	result, err := suite.handler.HandleGrant(&tokenReqWithResource, suite.oauthApp)

	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), constants.ErrorInvalidTarget, err.Error)
	assert.Equal(suite.T(), "Resource parameter mismatch", err.ErrorDescription)
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestHandleGrant_ResourceParameterMatch() {
	// Set up auth code with resource parameter
	authCodeWithResource := suite.testAuthzCode
	authCodeWithResource.Resource = testResourceURL

	suite.mockAuthzService.On("GetAuthorizationCodeDetails", testClientID, "test-auth-code").
		Return(&authCodeWithResource, nil)

	var capturedAudience string
	suite.mockTokenBuilder.On("BuildAccessToken", mock.MatchedBy(func(ctx *tokenservice.AccessTokenBuildContext) bool {
		capturedAudience = ctx.Audience
		return true
	})).Return(&model.TokenDTO{
		Token:     "mock-jwt-token",
		TokenType: constants.TokenTypeBearer,
		IssuedAt:  time.Now().Unix(),
		ExpiresIn: 3600,
		Scopes:    []string{"read", "write"},
		ClientID:  testClientID,
	}, nil)

	// Create token request with matching resource
	tokenReqWithResource := *suite.testTokenReq
	tokenReqWithResource.Resource = testResourceURL

	result, err := suite.handler.HandleGrant(&tokenReqWithResource, suite.oauthApp)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), testResourceURL, capturedAudience)
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestHandleGrant_NoResourceParameter() {
	// Auth code without resource parameter
	suite.mockAuthzService.On("GetAuthorizationCodeDetails", testClientID, "test-auth-code").
		Return(&suite.testAuthzCode, nil)

	var capturedAudience string
	suite.mockTokenBuilder.On("BuildAccessToken", mock.MatchedBy(func(ctx *tokenservice.AccessTokenBuildContext) bool {
		capturedAudience = ctx.Audience
		return true
	})).Return(&model.TokenDTO{
		Token:     "mock-jwt-token",
		TokenType: constants.TokenTypeBearer,
		IssuedAt:  time.Now().Unix(),
		ExpiresIn: 3600,
		Scopes:    []string{"read", "write"},
		ClientID:  testClientID,
	}, nil)

	// Create token request with matching resource
	tokenReqWithResource := *suite.testTokenReq
	tokenReqWithResource.Resource = testResourceURL

	result, err := suite.handler.HandleGrant(&tokenReqWithResource, suite.oauthApp)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)

	// Verify audience defaults to client ID when no resource parameter
	assert.Equal(suite.T(), testClientID, capturedAudience)
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestHandleGrant_FetchUserOUFailed() {
	suite.T().Skip("OU service is no longer used - OU details are retrieved from authz code " +
		"which comes from JWT claims")
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestHandleGrant_IDTokenGenerationFailed() {
	// Create auth code with openid scope
	authzCodeWithOpenID := suite.testAuthzCode
	authzCodeWithOpenID.Scopes = oidcReadWriteScopes

	suite.mockAuthzService.On("GetAuthorizationCodeDetails", testClientID, "test-auth-code").
		Return(&authzCodeWithOpenID, nil)

	// Mock access token generation succeeds
	suite.mockTokenBuilder.On("BuildAccessToken", mock.Anything).Return(&model.TokenDTO{
		Token:     "test-jwt-token",
		TokenType: constants.TokenTypeBearer,
		IssuedAt:  time.Now().Unix(),
		ExpiresIn: 3600,
		Scopes:    []string{"openid", "read", "write"},
		ClientID:  testClientID,
	}, nil)

	// Mock ID token generation fails
	suite.mockTokenBuilder.On("BuildIDToken", mock.Anything).
		Return(nil, errors.New("failed to generate ID token"))

	result, err := suite.handler.HandleGrant(suite.testTokenReq, suite.oauthApp)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorServerError, err.Error)
	assert.Equal(suite.T(), "Failed to generate ID token", err.ErrorDescription)

	suite.mockAuthzService.AssertExpectations(suite.T())
	suite.mockUserService.AssertExpectations(suite.T())
	suite.mockTokenBuilder.AssertExpectations(suite.T())
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestValidateGrant_ResourceWithFragment() {
	// Test resource parameter with fragment component
	tokenReq := &model.TokenRequest{
		GrantType:   string(constants.GrantTypeAuthorizationCode),
		ClientID:    testClientID,
		Code:        "test-code",
		RedirectURI: "https://client.example.com/callback",
		Resource:    "https://api.example.com/resource#fragment", // Fragment not allowed
	}

	err := suite.handler.ValidateGrant(tokenReq, suite.oauthApp)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, err.Error)
	assert.Contains(suite.T(), err.ErrorDescription, "fragment component")
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestValidateGrant_ResourceParseError() {
	// Test resource parameter that fails to parse
	tokenReq := &model.TokenRequest{
		GrantType:   string(constants.GrantTypeAuthorizationCode),
		ClientID:    testClientID,
		Code:        "test-code",
		RedirectURI: "https://client.example.com/callback",
		Resource:    "://invalid-uri", // Invalid URI format
	}

	err := suite.handler.ValidateGrant(tokenReq, suite.oauthApp)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, err.Error)
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestHandleGrant_FetchUserGroupsError() {
	// Test that groups are retrieved from authorization code (not fetched from DB)
	// Create OAuth app with groups configured
	oauthAppWithGroups := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:                testClientID,
		HashedClientSecret:      "hashed-secret",
		RedirectURIs:            []string{"https://client.example.com/callback"},
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
		ResponseTypes:           []constants.ResponseType{constants.ResponseTypeCode},
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
		Token: &appmodel.OAuthTokenConfig{
			AccessToken: &appmodel.AccessTokenConfig{
				UserAttributes: []string{"email", "username", constants.UserAttributeGroups},
			},
		},
	}

	// Add groups to auth code (from assertion)
	authzCodeWithGroups := suite.testAuthzCode
	authzCodeWithGroups.UserAttributes = map[string]interface{}{
		"email":    "test@example.com",
		"username": "testuser",
		"groups":   []string{"Admin", "Users"},
	}

	suite.mockAuthzService.On("GetAuthorizationCodeDetails", testClientID, "test-auth-code").
		Return(&authzCodeWithGroups, nil)

	// Groups come from the authorization code (extracted from assertion during authorization)
	// Token builder will extract groups from UserAttributes - verify groups are in UserAttributes
	suite.mockTokenBuilder.On("BuildAccessToken", mock.MatchedBy(func(ctx *tokenservice.AccessTokenBuildContext) bool {
		// Verify groups are in UserAttributes (will be extracted by token builder)
		groupsValue, ok := ctx.UserAttributes[constants.UserAttributeGroups]
		if !ok {
			return false
		}
		groupsArray := convertToStringSlice(groupsValue)
		return len(groupsArray) == 2 &&
			groupsArray[0] == "Admin" &&
			groupsArray[1] == "Users"
	})).Return(&model.TokenDTO{
		Token:     "test-jwt-token",
		TokenType: constants.TokenTypeBearer,
		IssuedAt:  time.Now().Unix(),
		ExpiresIn: 3600,
		Scopes:    []string{"read", "write"},
		ClientID:  testClientID,
	}, nil)

	result, err := suite.handler.HandleGrant(suite.testTokenReq, oauthAppWithGroups)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "test-jwt-token", result.AccessToken.Token)

	suite.mockAuthzService.AssertExpectations(suite.T())
	suite.mockTokenBuilder.AssertExpectations(suite.T())
	// Verify GetUserGroups was NOT called
	suite.mockUserService.AssertNotCalled(suite.T(), "GetUserGroups")
}

// createPKCEApp creates a test OAuth app with PKCE required
func (suite *AuthorizationCodeGrantHandlerTestSuite) createPKCEApp() *appmodel.OAuthAppConfigProcessedDTO {
	return &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:                testClientID,
		HashedClientSecret:      "hashed-secret",
		RedirectURIs:            []string{testClientCallbackURL},
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
		ResponseTypes:           []constants.ResponseType{constants.ResponseTypeCode},
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
		PKCERequired:            true,
	}
}

// createAuthCodeWithPKCE creates an authorization code with PKCE parameters
func (suite *AuthorizationCodeGrantHandlerTestSuite) createAuthCodeWithPKCE() authz.AuthorizationCode {
	authCodeWithPKCE := suite.testAuthzCode
	authCodeWithPKCE.CodeChallenge = testCodeChallenge
	authCodeWithPKCE.CodeChallengeMethod = testCodeChallengeMethodS256
	return authCodeWithPKCE
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestRetrieveAndValidateAuthCode_PKCERequiredMissingVerifier() {
	// Test PKCE required but code_verifier missing
	pkceApp := suite.createPKCEApp()
	authCodeWithPKCE := suite.createAuthCodeWithPKCE()

	suite.mockAuthzService.On("GetAuthorizationCodeDetails", testClientID, "test-auth-code").
		Return(&authCodeWithPKCE, nil)

	tokenReq := *suite.testTokenReq
	tokenReq.CodeVerifier = "" // Missing code verifier

	result, err := suite.handler.HandleGrant(&tokenReq, pkceApp)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, err.Error)
	assert.Contains(suite.T(), err.ErrorDescription, "code_verifier is required")

	suite.mockAuthzService.AssertExpectations(suite.T())
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestRetrieveAndValidateAuthCode_PKCEValidationFailed() {
	// Test PKCE validation failure
	pkceApp := suite.createPKCEApp()
	authCodeWithPKCE := suite.createAuthCodeWithPKCE()

	suite.mockAuthzService.On("GetAuthorizationCodeDetails", testClientID, "test-auth-code").
		Return(&authCodeWithPKCE, nil)

	tokenReq := *suite.testTokenReq
	tokenReq.CodeVerifier = "wrong-verifier-dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk" // Wrong verifier

	result, err := suite.handler.HandleGrant(&tokenReq, pkceApp)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorInvalidGrant, err.Error)
	assert.Contains(suite.T(), err.ErrorDescription, "Invalid code verifier")

	suite.mockAuthzService.AssertExpectations(suite.T())
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestRetrieveAndValidateAuthCode_CodeChallengeNotRequired() {
	authCodeWithPKCE := suite.testAuthzCode
	authCodeWithPKCE.CodeChallenge = testCodeChallenge
	authCodeWithPKCE.CodeChallengeMethod = testCodeChallengeMethodS256

	suite.mockAuthzService.On("GetAuthorizationCodeDetails", testClientID, "test-auth-code").
		Return(&authCodeWithPKCE, nil)

	// Mock token builder
	suite.mockTokenBuilder.On("BuildAccessToken", mock.Anything).Return(&model.TokenDTO{
		Token:     "test-jwt-token",
		TokenType: constants.TokenTypeBearer,
		IssuedAt:  time.Now().Unix(),
		ExpiresIn: 3600,
		Scopes:    []string{"read", "write"},
		ClientID:  testClientID,
	}, nil)

	tokenReq := *suite.testTokenReq
	tokenReq.CodeVerifier = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk" // Valid verifier

	result, err := suite.handler.HandleGrant(&tokenReq, suite.oauthApp)

	// Should succeed even though PKCE is not required, because code challenge was provided
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)

	suite.mockAuthzService.AssertExpectations(suite.T())
	suite.mockUserService.AssertExpectations(suite.T())
	suite.mockTokenBuilder.AssertExpectations(suite.T())
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestValidateAuthorizationCode_ResourceMismatch() {
	// Test resource parameter mismatch
	authCodeWithResource := suite.testAuthzCode
	authCodeWithResource.Resource = "https://api.example.com/resource"
	authCodeWithResource.RedirectURI = testClientCallbackURL

	tokenReq := &model.TokenRequest{
		ClientID:    testClientID,
		RedirectURI: "https://client.example.com/callback", // Must match auth code
		Resource:    testResourceURL,                       // Different resource
	}

	err := validateAuthorizationCode(tokenReq, authCodeWithResource)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorInvalidTarget, err.Error)
	assert.Equal(suite.T(), "Resource parameter mismatch", err.ErrorDescription)
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestValidateAuthorizationCode_ResourceMatch() {
	// Test resource parameter match
	authCodeWithResource := suite.testAuthzCode
	authCodeWithResource.Resource = testResourceURL
	authCodeWithResource.RedirectURI = testClientCallbackURL // Must match

	tokenReq := &model.TokenRequest{
		ClientID:    testClientID,
		RedirectURI: "https://client.example.com/callback", // Must match auth code
		Resource:    testResourceURL,                       // Matching resource
	}

	err := validateAuthorizationCode(tokenReq, authCodeWithResource)
	assert.Nil(suite.T(), err)
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestValidateAuthorizationCode_EmptyResourceInCode() {
	authCodeWithEmptyResource := suite.testAuthzCode
	authCodeWithEmptyResource.Resource = ""
	authCodeWithEmptyResource.RedirectURI = "https://client.example.com/callback" // Must match

	tokenReq := &model.TokenRequest{
		ClientID:    testClientID,
		RedirectURI: "https://client.example.com/callback", // Must match auth code
		Resource:    testResourceURL,                       // Any resource should be OK
	}

	err := validateAuthorizationCode(tokenReq, authCodeWithEmptyResource)
	assert.Nil(suite.T(), err)
}
