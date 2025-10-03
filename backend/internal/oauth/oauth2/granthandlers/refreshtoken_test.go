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
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/tests/mocks/jwtmock"
	usersvcmock "github.com/asgardeo/thunder/tests/mocks/user/servicemock"
)

type RefreshTokenGrantHandlerTestSuite struct {
	suite.Suite
	handler           *refreshTokenGrantHandler
	mockJWTService    *jwtmock.JWTServiceInterfaceMock
	mockUserService   *usersvcmock.UserServiceInterfaceMock
	oauthApp          *appmodel.OAuthAppConfigProcessedDTO
	validRefreshToken string
	validClaims       map[string]interface{}
	testTokenReq      *model.TokenRequest
}

func TestRefreshTokenGrantHandlerSuite(t *testing.T) {
	suite.Run(t, new(RefreshTokenGrantHandlerTestSuite))
}

func (suite *RefreshTokenGrantHandlerTestSuite) SetupTest() {
	// Reset ThunderRuntime before initializing with test config
	config.ResetThunderRuntime()

	// Initialize Thunder Runtime config with basic test config
	testConfig := &config.Config{
		OAuth: config.OAuthConfig{
			JWT: config.JWTConfig{
				ValidityPeriod: 3600,
			},
			RefreshToken: config.RefreshTokenConfig{
				ValidityPeriod: 86400,
				RenewOnGrant:   false,
			},
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)

	suite.mockJWTService = &jwtmock.JWTServiceInterfaceMock{}
	suite.mockUserService = usersvcmock.NewUserServiceInterfaceMock(suite.T())

	suite.handler = &refreshTokenGrantHandler{
		JWTService:  suite.mockJWTService,
		UserService: suite.mockUserService,
	}

	suite.oauthApp = &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:           "test-client-id",
		HashedClientSecret: "hashed-secret",
		GrantTypes:         []constants.GrantType{constants.GrantTypeRefreshToken},
		TokenEndpointAuthMethod: []constants.TokenEndpointAuthMethod{
			constants.TokenEndpointAuthMethodClientSecretPost},
		Token: &appmodel.OAuthTokenConfig{
			AccessToken: &appmodel.TokenConfig{
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
		"access_token_sub": "test-user-id",
		"access_token_aud": "test-audience",
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
	handler := newRefreshTokenGrantHandler()
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
	// Mock JWT service to return nil public key (simulating signature verification failure)
	suite.mockJWTService.On("VerifyJWT", "valid.refresh.token", "", "").Return(errors.New("public key not available"))

	ctx := &model.TokenContext{
		TokenAttributes: make(map[string]interface{}),
	}

	response, err := suite.handler.HandleGrant(suite.testTokenReq, suite.oauthApp, ctx)

	assert.Nil(suite.T(), response)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, err.Error)
	assert.Equal(suite.T(), "Invalid refresh token", err.ErrorDescription)
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestIssueRefreshToken_Success() {
	// Mock JWT service for refresh token generation
	suite.mockJWTService.On("GenerateJWT", "test-client-id", "test-client-id", mock.Anything,
		int64(3600), mock.Anything).Return("new.refresh.token",
		int64(1234567890), nil)

	tokenResponse := &model.TokenResponseDTO{}
	ctx := &model.TokenContext{
		TokenAttributes: map[string]interface{}{
			"sub": "test-user-id",
			"aud": "test-audience",
		},
	}

	err := suite.handler.IssueRefreshToken(tokenResponse, suite.oauthApp, ctx,
		"authorization_code", []string{"read", "write"})

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), tokenResponse.RefreshToken)
	assert.Equal(suite.T(), "new.refresh.token", tokenResponse.RefreshToken.Token)
	assert.Equal(suite.T(), constants.TokenTypeBearer, tokenResponse.RefreshToken.TokenType)
	assert.Equal(suite.T(), int64(1234567890), tokenResponse.RefreshToken.IssuedAt)
	assert.Equal(suite.T(), int64(3600), tokenResponse.RefreshToken.ExpiresIn)
	assert.Equal(suite.T(), []string{"read", "write"}, tokenResponse.RefreshToken.Scopes)
	assert.Equal(suite.T(), "test-client-id", tokenResponse.RefreshToken.ClientID)
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestIssueRefreshToken_JWTGenerationError() {
	// Mock JWT service to return error
	suite.mockJWTService.On("GenerateJWT", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything).Return("", int64(0), errors.New("JWT generation failed"))

	tokenResponse := &model.TokenResponseDTO{}
	ctx := &model.TokenContext{
		TokenAttributes: make(map[string]interface{}),
	}

	err := suite.handler.IssueRefreshToken(tokenResponse, suite.oauthApp, ctx, "authorization_code",
		[]string{"read"})

	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorServerError, err.Error)
	assert.Equal(suite.T(), "Failed to generate refresh token", err.ErrorDescription)
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestExtractScopes_NoScopesInRefreshToken() {
	claims := map[string]interface{}{
		"client_id": "test-client-id",
	}

	refreshScopes, newScopes, err := suite.handler.extractScopes("read", claims, log.GetLogger())

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), []string{}, refreshScopes)
	assert.Equal(suite.T(), []string{}, newScopes)
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestExtractScopes_InvalidScopeType() {
	// Create a mock logger
	logger := log.GetLogger()

	claims := map[string]interface{}{
		"scopes": 123, // Invalid type
	}

	refreshScopes, newScopes, err := suite.handler.extractScopes("read", claims, logger)

	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, err.Error)
	assert.Equal(suite.T(), "Invalid refresh token", err.ErrorDescription)
	assert.Nil(suite.T(), refreshScopes)
	assert.Nil(suite.T(), newScopes)
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestExtractScopes_RequestedScopeNotInRefreshToken() {
	// Create a mock logger
	logger := log.GetLogger()

	claims := map[string]interface{}{
		"scopes": "read write",
	}

	refreshScopes, newScopes, err := suite.handler.extractScopes("admin", claims, logger)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), []string{"read", "write"}, refreshScopes)
	assert.Equal(suite.T(), []string{}, newScopes) // admin not in refresh token scopes
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestExtractScopes_PartialScopeMatch() {
	// Create a mock logger
	logger := log.GetLogger()

	claims := map[string]interface{}{
		"scopes": "read write admin",
	}

	refreshScopes, newScopes, err := suite.handler.extractScopes("read admin invalid", claims, logger)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), []string{"read", "write", "admin"}, refreshScopes)
	assert.Equal(suite.T(), []string{"read", "admin"}, newScopes) // Only valid scopes
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestExtractScopes_NoRequestedScopes() {
	// Create a mock logger
	logger := log.GetLogger()

	claims := map[string]interface{}{
		"scopes": "read write",
	}

	refreshScopes, newScopes, err := suite.handler.extractScopes("", claims, logger)

	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), []string{"read", "write"}, refreshScopes)
	assert.Equal(suite.T(), []string{"read", "write"}, newScopes) // All refresh token scopes granted
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestValidateTimeClaim_ValidClaim() {
	// Create a mock logger
	logger := log.GetLogger()

	now := time.Now().Unix()
	claims := map[string]interface{}{
		"exp": float64(now + 3600), // Valid future time
	}

	err := suite.handler.validateTimeClaim(claims, "exp",
		func(now, claim int64) bool { return now > claim }, // Expired check
		"Token has expired", "Expired token", logger)

	assert.Nil(suite.T(), err)
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestValidateTimeClaim_ExpiredToken() {
	// Create a mock logger
	logger := log.GetLogger()

	now := time.Now().Unix()
	claims := map[string]interface{}{
		"exp": float64(now - 3600), // Expired time
	}

	err := suite.handler.validateTimeClaim(claims, "exp",
		func(now, claim int64) bool { return now > claim }, // Expired check
		"Token has expired", "Expired token", logger)

	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, err.Error)
	assert.Equal(suite.T(), "Expired token", err.ErrorDescription)
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestValidateTimeClaim_MissingClaim() {
	// Create a mock logger
	logger := log.GetLogger()

	claims := map[string]interface{}{}

	err := suite.handler.validateTimeClaim(claims, "exp",
		func(now, claim int64) bool { return now > claim },
		"Token has expired", "Expired token", logger)

	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, err.Error)
	assert.Equal(suite.T(), "Invalid refresh token", err.ErrorDescription)
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestValidateTimeClaim_InvalidType() {
	// Create a mock logger
	logger := log.GetLogger()

	claims := map[string]interface{}{
		"exp": "invalid-time", // Invalid type
	}

	err := suite.handler.validateTimeClaim(claims, "exp",
		func(now, claim int64) bool { return now > claim },
		"Token has expired", "Expired token", logger)

	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, err.Error)
	assert.Equal(suite.T(), "Invalid refresh token", err.ErrorDescription)
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestValidateNBF_ValidTime() {
	// Create a mock logger
	logger := log.GetLogger()

	now := time.Now().Unix()
	claims := map[string]interface{}{
		"nbf": float64(now - 3600), // Valid past time
	}

	err := suite.handler.validateNBF(claims, logger)
	assert.Nil(suite.T(), err)
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestValidateNBF_FutureTime() {
	// Create a mock logger
	logger := log.GetLogger()

	now := time.Now().Unix()
	claims := map[string]interface{}{
		"nbf": float64(now + 3600), // Future time - not valid yet
	}

	err := suite.handler.validateNBF(claims, logger)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, err.Error)
	assert.Equal(suite.T(), "Refresh token not valid yet", err.ErrorDescription)
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestValidateNBF_MissingClaim() {
	// Create a mock logger
	logger := log.GetLogger()

	claims := map[string]interface{}{}

	err := suite.handler.validateNBF(claims, logger)
	assert.Nil(suite.T(), err) // NBF is optional
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestValidateNBF_InvalidType() {
	// Create a mock logger
	logger := log.GetLogger()

	claims := map[string]interface{}{
		"nbf": "invalid", // Invalid type
	}

	err := suite.handler.validateNBF(claims, logger)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, err.Error)
	assert.Equal(suite.T(), "Invalid refresh token", err.ErrorDescription)
}
