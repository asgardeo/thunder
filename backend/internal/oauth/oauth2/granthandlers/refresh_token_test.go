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
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/tokenservice"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/tests/mocks/jwtmock"
	"github.com/asgardeo/thunder/tests/mocks/oauth/oauth2/tokenservicemock"
	usersvcmock "github.com/asgardeo/thunder/tests/mocks/usermock"
)

// testUserID and testAudience are declared in tokenexchange_test.go
const testRefreshTokenUserID = "test-user-id"
const testRefreshTokenAudience = "test-audience"

type RefreshTokenGrantHandlerTestSuite struct {
	suite.Suite
	handler            *refreshTokenGrantHandler
	mockJWTService     *jwtmock.JWTServiceInterfaceMock
	mockUserService    *usersvcmock.UserServiceInterfaceMock
	mockTokenBuilder   *tokenservicemock.TokenBuilderInterfaceMock
	mockTokenValidator *tokenservicemock.TokenValidatorInterfaceMock
	oauthApp           *appmodel.OAuthAppConfigProcessedDTO
	validRefreshToken  string
	validClaims        map[string]interface{}
	testTokenReq       *model.TokenRequest
}

func TestRefreshTokenGrantHandlerSuite(t *testing.T) {
	suite.Run(t, new(RefreshTokenGrantHandlerTestSuite))
}

func (suite *RefreshTokenGrantHandlerTestSuite) SetupTest() {
	// Reset ThunderRuntime before initializing with test config
	config.ResetThunderRuntime()

	// Initialize Thunder Runtime config with basic test config
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			ValidityPeriod: 3600,
		},
		OAuth: config.OAuthConfig{
			RefreshToken: config.RefreshTokenConfig{
				ValidityPeriod: 86400,
				RenewOnGrant:   false,
			},
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)

	suite.mockJWTService = jwtmock.NewJWTServiceInterfaceMock(suite.T())
	suite.mockUserService = usersvcmock.NewUserServiceInterfaceMock(suite.T())
	suite.mockTokenBuilder = tokenservicemock.NewTokenBuilderInterfaceMock(suite.T())
	suite.mockTokenValidator = tokenservicemock.NewTokenValidatorInterfaceMock(suite.T())

	suite.handler = &refreshTokenGrantHandler{
		jwtService:     suite.mockJWTService,
		userService:    suite.mockUserService,
		tokenBuilder:   suite.mockTokenBuilder,
		tokenValidator: suite.mockTokenValidator,
	}

	suite.oauthApp = &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:                "test-client-id",
		HashedClientSecret:      "hashed-secret",
		GrantTypes:              []constants.GrantType{constants.GrantTypeRefreshToken},
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
		Token: &appmodel.OAuthTokenConfig{
			AccessToken: &appmodel.AccessTokenConfig{
				UserAttributes: []string{"email", "username"},
			},
		},
	}

	suite.validRefreshToken = "valid.refresh.token"
	now := time.Now().Unix()
	suite.validClaims = map[string]interface{}{
		"iat":              float64(now - 3600),
		"exp":              float64(now + 86400),
		"client_id":        "test-client-id",
		"grant_type":       "authorization_code",
		"scopes":           "read write",
		"access_token_sub": testRefreshTokenUserID,
		"access_token_aud": testRefreshTokenAudience,
	}

	suite.testTokenReq = &model.TokenRequest{
		GrantType:    string(constants.GrantTypeRefreshToken),
		ClientID:     "test-client-id",
		RefreshToken: suite.validRefreshToken,
		Scope:        "read",
	}
}

func (suite *RefreshTokenGrantHandlerTestSuite) TearDownTest() {
	config.ResetThunderRuntime()
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestNewRefreshTokenGrantHandler() {
	handler := newRefreshTokenGrantHandler(
		suite.mockJWTService,
		suite.mockUserService,
		suite.mockTokenBuilder,
		suite.mockTokenValidator,
	)
	assert.NotNil(suite.T(), handler)
	assert.Implements(suite.T(), (*RefreshTokenGrantHandlerInterface)(nil), handler)
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestValidateGrant_Success() {
	err := suite.handler.ValidateGrant(suite.testTokenReq, suite.oauthApp)
	assert.Nil(suite.T(), err)
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestValidateGrant_InvalidGrantType() {
	tokenReq := &model.TokenRequest{
		GrantType:    "invalid_grant",
		ClientID:     "test-client-id",
		RefreshToken: "token",
	}

	err := suite.handler.ValidateGrant(tokenReq, suite.oauthApp)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorUnsupportedGrantType, err.Error)
	assert.Equal(suite.T(), "Unsupported grant type", err.ErrorDescription)
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestValidateGrant_MissingRefreshToken() {
	tokenReq := &model.TokenRequest{
		GrantType: string(constants.GrantTypeRefreshToken),
		ClientID:  "test-client-id",
	}

	err := suite.handler.ValidateGrant(tokenReq, suite.oauthApp)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, err.Error)
	assert.Equal(suite.T(), "Refresh token is required", err.ErrorDescription)
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestValidateGrant_MissingClientID() {
	tokenReq := &model.TokenRequest{
		GrantType:    string(constants.GrantTypeRefreshToken),
		RefreshToken: "token",
	}

	err := suite.handler.ValidateGrant(tokenReq, suite.oauthApp)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, err.Error)
	assert.Equal(suite.T(), "Client ID is required", err.ErrorDescription)
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestHandleGrant_InvalidSignature() {
	// Mock token validator to return error (simulating signature verification failure)
	suite.mockTokenValidator.On("ValidateRefreshToken", suite.validRefreshToken, "test-client-id").
		Return(nil, errors.New("public key not available"))

	response, err := suite.handler.HandleGrant(suite.testTokenReq, suite.oauthApp)

	assert.Nil(suite.T(), response)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorInvalidGrant, err.Error)
	assert.Equal(suite.T(), "Invalid refresh token", err.ErrorDescription)
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestIssueRefreshToken_Success() {
	// Mock token builder for refresh token generation
	suite.mockTokenBuilder.On("BuildRefreshToken", mock.MatchedBy(
		func(ctx *tokenservice.RefreshTokenBuildContext) bool {
			return ctx.ClientID == "test-client-id" &&
				ctx.GrantType == "authorization_code" &&
				ctx.AccessTokenSubject == testRefreshTokenUserID &&
				ctx.AccessTokenAudience == testRefreshTokenAudience
		})).Return(&model.TokenDTO{
		Token:     "new.refresh.token",
		TokenType: "",
		IssuedAt:  int64(1234567890),
		ExpiresIn: 3600,
		Scopes:    []string{"read", "write"},
		ClientID:  "test-client-id",
	}, nil)

	tokenResponse := &model.TokenResponseDTO{}

	err := suite.handler.IssueRefreshToken(tokenResponse, suite.oauthApp,
		testRefreshTokenUserID, testRefreshTokenAudience,
		"authorization_code", []string{"read", "write"})

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), tokenResponse.RefreshToken)
	assert.Equal(suite.T(), "new.refresh.token", tokenResponse.RefreshToken.Token)
	assert.Equal(suite.T(), "", tokenResponse.RefreshToken.TokenType)
	assert.Equal(suite.T(), int64(1234567890), tokenResponse.RefreshToken.IssuedAt)
	assert.Equal(suite.T(), int64(3600), tokenResponse.RefreshToken.ExpiresIn)
	assert.Equal(suite.T(), []string{"read", "write"}, tokenResponse.RefreshToken.Scopes)
	assert.Equal(suite.T(), "test-client-id", tokenResponse.RefreshToken.ClientID)
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestIssueRefreshToken_JWTGenerationError() {
	// Mock token builder to return error
	suite.mockTokenBuilder.On("BuildRefreshToken", mock.Anything).
		Return(nil, errors.New("JWT generation failed"))

	tokenResponse := &model.TokenResponseDTO{}

	err := suite.handler.IssueRefreshToken(tokenResponse, suite.oauthApp, "", "",
		"authorization_code", []string{"read"})

	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorServerError, err.Error)
	assert.Equal(suite.T(), "Failed to generate refresh token", err.ErrorDescription)
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestIssueRefreshToken_WithEmptyTokenAttributes() {
	// Mock token builder with matcher that checks for empty sub and aud
	suite.mockTokenBuilder.On("BuildRefreshToken", mock.MatchedBy(
		func(ctx *tokenservice.RefreshTokenBuildContext) bool {
			return ctx.AccessTokenSubject == "" && ctx.AccessTokenAudience == ""
		})).Return(&model.TokenDTO{
		Token:    "new.refresh.token",
		IssuedAt: int64(1234567890),
	}, nil)

	tokenResponse := &model.TokenResponseDTO{}

	err := suite.handler.IssueRefreshToken(tokenResponse, suite.oauthApp, "", "",
		"authorization_code", []string{"read"})

	assert.Nil(suite.T(), err)
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestHandleGrant_Success_WithRenewOnGrantDisabled() {
	// Mock successful refresh token validation
	suite.mockTokenValidator.On("ValidateRefreshToken", suite.validRefreshToken, "test-client-id").
		Return(&tokenservice.RefreshTokenClaims{
			Sub:            testRefreshTokenUserID,
			Aud:            testRefreshTokenAudience,
			Scopes:         []string{"read", "write"},
			GrantType:      "authorization_code",
			UserAttributes: map[string]interface{}{"email": "test@example.com"},
			Iat:            int64(suite.validClaims["iat"].(float64)),
		}, nil)

	// Mock successful access token generation
	suite.mockTokenBuilder.On("BuildAccessToken", mock.MatchedBy(
		func(ctx *tokenservice.AccessTokenBuildContext) bool {
			return ctx.Subject == testRefreshTokenUserID &&
				ctx.ClientID == "test-client-id" &&
				len(ctx.Scopes) == 1 && ctx.Scopes[0] == "read"
		})).Return(&model.TokenDTO{
		Token:     "new.access.token",
		IssuedAt:  time.Now().Unix(),
		ExpiresIn: 3600,
		Scopes:    []string{"read"},
	}, nil)

	response, err := suite.handler.HandleGrant(suite.testTokenReq, suite.oauthApp)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), response)
	assert.Equal(suite.T(), "new.access.token", response.AccessToken.Token)
	assert.Equal(suite.T(), suite.validRefreshToken, response.RefreshToken.Token)
	assert.Equal(suite.T(), []string{"read", "write"}, response.RefreshToken.Scopes)
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestHandleGrant_Success_WithRenewOnGrantEnabled() {
	// Enable RenewOnGrant in config
	config.GetThunderRuntime().Config.OAuth.RefreshToken.RenewOnGrant = true

	// Mock successful refresh token validation
	suite.mockTokenValidator.On("ValidateRefreshToken", suite.validRefreshToken, "test-client-id").
		Return(&tokenservice.RefreshTokenClaims{
			Sub:            testRefreshTokenUserID,
			Aud:            testRefreshTokenAudience,
			Scopes:         []string{"read", "write"},
			GrantType:      "authorization_code",
			UserAttributes: map[string]interface{}{"email": "test@example.com"},
			Iat:            int64(suite.validClaims["iat"].(float64)),
		}, nil)

	// Mock successful access token generation
	suite.mockTokenBuilder.On("BuildAccessToken", mock.Anything).Return(&model.TokenDTO{
		Token:          "new.access.token",
		IssuedAt:       time.Now().Unix(),
		ExpiresIn:      3600,
		Scopes:         []string{"read"},
		UserAttributes: map[string]interface{}{"email": "test@example.com"},
	}, nil)

	// Mock successful refresh token generation
	suite.mockTokenBuilder.On("BuildRefreshToken", mock.MatchedBy(
		func(ctx *tokenservice.RefreshTokenBuildContext) bool {
			return ctx.AccessTokenSubject == testRefreshTokenUserID &&
				ctx.AccessTokenAudience == testRefreshTokenAudience
		})).Return(&model.TokenDTO{
		Token:     "new.refresh.token",
		IssuedAt:  time.Now().Unix(),
		ExpiresIn: 86400,
		Scopes:    []string{"read"},
	}, nil)

	response, err := suite.handler.HandleGrant(suite.testTokenReq, suite.oauthApp)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), response)
	assert.Equal(suite.T(), "new.access.token", response.AccessToken.Token)
	assert.Equal(suite.T(), "new.refresh.token", response.RefreshToken.Token)
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestHandleGrant_BuildAccessTokenError() {
	// Mock successful refresh token validation
	suite.mockTokenValidator.On("ValidateRefreshToken", suite.validRefreshToken, "test-client-id").
		Return(&tokenservice.RefreshTokenClaims{
			Sub:            testRefreshTokenUserID,
			Aud:            testRefreshTokenAudience,
			Scopes:         []string{"read"},
			GrantType:      "authorization_code",
			UserAttributes: map[string]interface{}{},
			Iat:            int64(suite.validClaims["iat"].(float64)),
		}, nil)

	// Mock failed access token generation
	suite.mockTokenBuilder.On("BuildAccessToken", mock.Anything).
		Return(nil, errors.New("failed to sign JWT"))

	response, err := suite.handler.HandleGrant(suite.testTokenReq, suite.oauthApp)

	assert.Nil(suite.T(), response)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorServerError, err.Error)
	assert.Equal(suite.T(), "Failed to generate access token", err.ErrorDescription)
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestHandleGrant_IssueRefreshTokenError() {
	// Enable RenewOnGrant in config
	config.GetThunderRuntime().Config.OAuth.RefreshToken.RenewOnGrant = true

	// Mock successful refresh token validation
	suite.mockTokenValidator.On("ValidateRefreshToken", suite.validRefreshToken, "test-client-id").
		Return(&tokenservice.RefreshTokenClaims{
			Sub:            testRefreshTokenUserID,
			Aud:            testRefreshTokenAudience,
			Scopes:         []string{"read"},
			GrantType:      "authorization_code",
			UserAttributes: map[string]interface{}{},
			Iat:            int64(suite.validClaims["iat"].(float64)),
		}, nil)

	// Mock successful access token generation
	suite.mockTokenBuilder.On("BuildAccessToken", mock.Anything).Return(&model.TokenDTO{
		Token:          "new.access.token",
		IssuedAt:       time.Now().Unix(),
		ExpiresIn:      3600,
		Scopes:         []string{"read"},
		UserAttributes: map[string]interface{}{},
	}, nil)

	// Mock failed refresh token generation
	suite.mockTokenBuilder.On("BuildRefreshToken", mock.Anything).
		Return(nil, errors.New("refresh token generation failed"))

	response, err := suite.handler.HandleGrant(suite.testTokenReq, suite.oauthApp)

	assert.Nil(suite.T(), response)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorServerError, err.Error)
	assert.Contains(suite.T(), err.ErrorDescription, "Error while issuing refresh token")
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestHandleGrant_ExtractIatClaimError() {
	// RenewOnGrant is disabled by default in SetupTest

	// Mock validator to return error when iat is missing (validation fails)
	suite.mockTokenValidator.On("ValidateRefreshToken", suite.validRefreshToken, "test-client-id").
		Return(nil, errors.New("missing or invalid 'iat' claim"))

	response, err := suite.handler.HandleGrant(suite.testTokenReq, suite.oauthApp)

	assert.Nil(suite.T(), response)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorInvalidGrant, err.Error)
	assert.Equal(suite.T(), "Invalid refresh token", err.ErrorDescription)
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestApplyScopeDownscoping_NoScopesRequested() {
	refreshTokenScopes := []string{"read", "write", "delete"}
	logger := log.GetLogger()

	result := suite.handler.applyScopeDownscoping("", refreshTokenScopes, logger)

	assert.Equal(suite.T(), refreshTokenScopes, result)
	assert.Len(suite.T(), result, 3)
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestApplyScopeDownscoping_RequestedScopesSubset() {
	refreshTokenScopes := []string{"read", "write", "delete"}
	logger := log.GetLogger()

	result := suite.handler.applyScopeDownscoping("read write", refreshTokenScopes, logger)

	assert.Len(suite.T(), result, 2)
	assert.Contains(suite.T(), result, "read")
	assert.Contains(suite.T(), result, "write")
	assert.NotContains(suite.T(), result, "delete")
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestApplyScopeDownscoping_SomeRequestedScopesNotInRefreshToken() {
	refreshTokenScopes := []string{"read", "write"}
	logger := log.GetLogger()

	// Request "read", "write", and "delete" - but "delete" is not in refresh token
	result := suite.handler.applyScopeDownscoping("read write delete admin", refreshTokenScopes, logger)

	assert.Len(suite.T(), result, 2)
	assert.Contains(suite.T(), result, "read")
	assert.Contains(suite.T(), result, "write")
	assert.NotContains(suite.T(), result, "delete")
	assert.NotContains(suite.T(), result, "admin")
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestApplyScopeDownscoping_NoMatchingScopes() {
	refreshTokenScopes := []string{"read", "write"}
	logger := log.GetLogger()

	result := suite.handler.applyScopeDownscoping("admin delete", refreshTokenScopes, logger)

	assert.Empty(suite.T(), result)
}
