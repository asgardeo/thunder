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

package tokenservice

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/tests/mocks/jwtmock"
)

const (
	testAccessToken  = "test-access-token"  //nolint:gosec // Test token, not a real credential
	testRefreshToken = "test-refresh-token" //nolint:gosec // Test token, not a real credential
	testIDToken      = "test-id-token"      //nolint:gosec // Test token, not a real credential
	testUserName     = "John Doe"
	testAppID        = "app123"
)

type TokenBuilderTestSuite struct {
	suite.Suite
	mockJWTService *jwtmock.JWTServiceInterfaceMock
	builder        *tokenBuilder
	oauthApp       *appmodel.OAuthAppConfigProcessedDTO
}

func TestTokenBuilderTestSuite(t *testing.T) {
	suite.Run(t, new(TokenBuilderTestSuite))
}

func (suite *TokenBuilderTestSuite) SetupTest() {
	// Initialize Thunder Runtime for tests
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "https://thunder.io",
			ValidityPeriod: 3600,
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)

	suite.mockJWTService = jwtmock.NewJWTServiceInterfaceMock(suite.T())
	suite.builder = &tokenBuilder{
		jwtService: suite.mockJWTService,
	}

	suite.oauthApp = &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Token: &appmodel.OAuthTokenConfig{
			Issuer: "https://thunder.io",
			AccessToken: &appmodel.AccessTokenConfig{
				ValidityPeriod: 3600,
				UserAttributes: []string{"name"}, // Configure user attributes for tests
			},
		},
	}
}

func (suite *TokenBuilderTestSuite) TestNewTokenBuilder() {
	jwtService := jwtmock.NewJWTServiceInterfaceMock(suite.T())
	builder := newTokenBuilder(jwtService)

	assert.NotNil(suite.T(), builder)
	assert.Implements(suite.T(), (*TokenBuilderInterface)(nil), builder)
}

func (suite *TokenBuilderTestSuite) TestBuildAccessToken_Success_Basic() {
	ctx := &AccessTokenBuildContext{
		Subject:        "user123",
		Audience:       "app123",
		ClientID:       "test-client",
		Scopes:         []string{"read", "write"},
		UserAttributes: map[string]interface{}{"name": testUserName},
		GrantType:      string(constants.GrantTypeAuthorizationCode),
		OAuthApp:       suite.oauthApp,
	}

	expectedToken := testAccessToken
	expectedIat := time.Now().Unix()

	suite.mockJWTService.On("GenerateJWT",
		"user123",
		"app123",
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			return claims["scope"] == "read write" &&
				claims["client_id"] == "test-client" &&
				claims["grant_type"] == string(constants.GrantTypeAuthorizationCode) &&
				claims["name"] == testUserName
		}),
	).Return(expectedToken, expectedIat, nil)

	result, err := suite.builder.BuildAccessToken(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), expectedToken, result.Token)
	assert.Equal(suite.T(), constants.TokenTypeBearer, result.TokenType)
	assert.Equal(suite.T(), expectedIat, result.IssuedAt)
	assert.Equal(suite.T(), int64(3600), result.ExpiresIn)
	assert.Equal(suite.T(), []string{"read", "write"}, result.Scopes)
	assert.Equal(suite.T(), "test-client", result.ClientID)
	assert.Equal(suite.T(), map[string]interface{}{"name": testUserName}, result.UserAttributes)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenBuilderTestSuite) TestBuildAccessToken_Success_WithActorClaim() {
	actorClaims := &SubjectTokenClaims{
		Sub:            "actor123",
		Iss:            "https://actor-issuer.com",
		Aud:            "",
		UserAttributes: map[string]interface{}{},
		NestedAct:      nil,
	}

	ctx := &AccessTokenBuildContext{
		Subject:        "user123",
		Audience:       "app123",
		ClientID:       "test-client",
		Scopes:         []string{"read"},
		UserAttributes: map[string]interface{}{},
		GrantType:      string(constants.GrantTypeTokenExchange),
		OAuthApp:       suite.oauthApp,
		ActorClaims:    actorClaims,
	}

	expectedToken := testAccessToken
	expectedIat := time.Now().Unix()

	suite.mockJWTService.On("GenerateJWT",
		"user123",
		"app123",
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			act, ok := claims["act"].(map[string]interface{})
			return ok && act["sub"] == "actor123" && act["iss"] == "https://actor-issuer.com"
		}),
	).Return(expectedToken, expectedIat, nil)

	result, err := suite.builder.BuildAccessToken(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenBuilderTestSuite) TestBuildAccessToken_Success_WithNestedActorClaim() {
	nestedActorClaims := &SubjectTokenClaims{
		Sub:            "nested-actor",
		Iss:            "https://nested-issuer.com",
		Aud:            "",
		UserAttributes: map[string]interface{}{},
		NestedAct: map[string]interface{}{
			"sub": "original-actor",
		},
	}

	ctx := &AccessTokenBuildContext{
		Subject:        "user123",
		Audience:       "app123",
		ClientID:       "test-client",
		Scopes:         []string{"read"},
		UserAttributes: map[string]interface{}{},
		GrantType:      string(constants.GrantTypeTokenExchange),
		OAuthApp:       suite.oauthApp,
		ActorClaims:    nestedActorClaims,
	}

	expectedToken := testAccessToken
	expectedIat := time.Now().Unix()

	suite.mockJWTService.On("GenerateJWT",
		"user123",
		"app123",
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			act, ok := claims["act"].(map[string]interface{})
			return ok && act["sub"] == "nested-actor" && act["act"] != nil
		}),
	).Return(expectedToken, expectedIat, nil)

	result, err := suite.builder.BuildAccessToken(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenBuilderTestSuite) TestBuildAccessToken_Success_EmptyScopes() {
	ctx := &AccessTokenBuildContext{
		Subject:        "user123",
		Audience:       "app123",
		ClientID:       "test-client",
		Scopes:         []string{},
		UserAttributes: map[string]interface{}{},
		GrantType:      string(constants.GrantTypeAuthorizationCode),
		OAuthApp:       suite.oauthApp,
	}

	expectedToken := testAccessToken
	expectedIat := time.Now().Unix()

	suite.mockJWTService.On("GenerateJWT",
		"user123",
		"app123",
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			_, hasScope := claims["scope"]
			return !hasScope
		}),
	).Return(expectedToken, expectedIat, nil)

	result, err := suite.builder.BuildAccessToken(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenBuilderTestSuite) TestBuildAccessToken_Success_EmptyClientID() {
	ctx := &AccessTokenBuildContext{
		Subject:        "user123",
		Audience:       "app123",
		ClientID:       "",
		Scopes:         []string{"read"},
		UserAttributes: map[string]interface{}{},
		GrantType:      string(constants.GrantTypeAuthorizationCode),
		OAuthApp:       suite.oauthApp,
	}

	expectedToken := testAccessToken
	expectedIat := time.Now().Unix()

	suite.mockJWTService.On("GenerateJWT",
		"user123",
		"app123",
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			_, hasClientID := claims["client_id"]
			return !hasClientID
		}),
	).Return(expectedToken, expectedIat, nil)

	result, err := suite.builder.BuildAccessToken(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenBuilderTestSuite) TestBuildAccessToken_Success_EmptyGrantType() {
	ctx := &AccessTokenBuildContext{
		Subject:        "user123",
		Audience:       "app123",
		ClientID:       "test-client",
		Scopes:         []string{"read"},
		UserAttributes: map[string]interface{}{},
		GrantType:      "",
		OAuthApp:       suite.oauthApp,
	}

	expectedToken := testAccessToken
	expectedIat := time.Now().Unix()

	suite.mockJWTService.On("GenerateJWT",
		"user123",
		"app123",
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			_, hasGrantType := claims["grant_type"]
			return !hasGrantType
		}),
	).Return(expectedToken, expectedIat, nil)

	result, err := suite.builder.BuildAccessToken(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenBuilderTestSuite) TestBuildAccessToken_Success_CustomIssuer() {
	customOAuthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Token: &appmodel.OAuthTokenConfig{
			Issuer: "https://custom.thunder.io", // OAuth-level issuer used for all tokens
			AccessToken: &appmodel.AccessTokenConfig{
				ValidityPeriod: 7200,
			},
		},
	}

	ctx := &AccessTokenBuildContext{
		Subject:        "user123",
		Audience:       "app123",
		ClientID:       "test-client",
		Scopes:         []string{"read"},
		UserAttributes: map[string]interface{}{},
		GrantType:      string(constants.GrantTypeAuthorizationCode),
		OAuthApp:       customOAuthApp,
	}

	expectedToken := testAccessToken
	expectedIat := time.Now().Unix()

	suite.mockJWTService.On("GenerateJWT",
		"user123",
		"app123",
		"https://custom.thunder.io", // OAuth-level issuer
		int64(7200),
		mock.Anything,
	).Return(expectedToken, expectedIat, nil)

	result, err := suite.builder.BuildAccessToken(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), int64(7200), result.ExpiresIn)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenBuilderTestSuite) TestBuildAccessToken_Error_NilContext() {
	result, err := suite.builder.BuildAccessToken(nil)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "build context cannot be nil")
}

func (suite *TokenBuilderTestSuite) TestBuildAccessToken_Error_JWTGenerationFailed() {
	ctx := &AccessTokenBuildContext{
		Subject:        "user123",
		Audience:       "app123",
		ClientID:       "test-client",
		Scopes:         []string{"read"},
		UserAttributes: map[string]interface{}{},
		GrantType:      string(constants.GrantTypeAuthorizationCode),
		OAuthApp:       suite.oauthApp,
	}

	suite.mockJWTService.On("GenerateJWT",
		"user123",
		"app123",
		"https://thunder.io",
		int64(3600),
		mock.Anything,
	).Return("", int64(0), &serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "JWT_GENERATION_FAILED",
		Error:            "JWT generation failed",
		ErrorDescription: "Failed to generate JWT token",
	})

	result, err := suite.builder.BuildAccessToken(ctx)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "failed to generate access token")
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenBuilderTestSuite) TestBuildRefreshToken_Success_Basic() {
	// Create OAuth app with user attributes configured
	oauthAppWithUserAttrs := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Token: &appmodel.OAuthTokenConfig{
			Issuer: "https://thunder.io",
			AccessToken: &appmodel.AccessTokenConfig{
				ValidityPeriod: 3600,
				UserAttributes: []string{"name"}, // Configure user attributes
			},
		},
	}

	ctx := &RefreshTokenBuildContext{
		ClientID:             "test-client",
		Scopes:               []string{"read", "write"},
		GrantType:            string(constants.GrantTypeAuthorizationCode),
		AccessTokenSubject:   "user123",
		AccessTokenAudience:  "app123",
		AccessTokenUserAttrs: map[string]interface{}{"name": testUserName},
		OAuthApp:             oauthAppWithUserAttrs,
	}

	expectedToken := testRefreshToken
	expectedIat := time.Now().Unix()

	suite.mockJWTService.On("GenerateJWT",
		"test-client",
		"https://thunder.io",
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			return claims["scope"] == "read write" &&
				claims["access_token_sub"] == "user123" &&
				claims["access_token_aud"] == testAppID &&
				claims["grant_type"] == string(constants.GrantTypeAuthorizationCode) &&
				claims["access_token_user_attributes"].(map[string]interface{})["name"] == testUserName
		}),
	).Return(expectedToken, expectedIat, nil)

	result, err := suite.builder.BuildRefreshToken(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), expectedToken, result.Token)
	assert.Equal(suite.T(), expectedIat, result.IssuedAt)
	assert.Equal(suite.T(), int64(3600), result.ExpiresIn)
	assert.Equal(suite.T(), []string{"read", "write"}, result.Scopes)
	assert.Equal(suite.T(), "test-client", result.ClientID)
	assert.Equal(suite.T(), "https://thunder.io", result.Audience)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenBuilderTestSuite) TestBuildRefreshToken_Success_WithoutUserAttributes() {
	ctx := &RefreshTokenBuildContext{
		ClientID:             "test-client",
		Scopes:               []string{"read"},
		GrantType:            string(constants.GrantTypeAuthorizationCode),
		AccessTokenSubject:   "user123",
		AccessTokenAudience:  "app123",
		AccessTokenUserAttrs: map[string]interface{}{},
		OAuthApp:             suite.oauthApp,
	}

	expectedToken := testRefreshToken
	expectedIat := time.Now().Unix()

	suite.mockJWTService.On("GenerateJWT",
		"test-client",
		"https://thunder.io",
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			_, hasUserAttrs := claims["access_token_user_attributes"]
			return !hasUserAttrs
		}),
	).Return(expectedToken, expectedIat, nil)

	result, err := suite.builder.BuildRefreshToken(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenBuilderTestSuite) TestBuildRefreshToken_Success_WithNilOAuthApp() {
	ctx := &RefreshTokenBuildContext{
		ClientID:             "test-client",
		Scopes:               []string{"read"},
		GrantType:            string(constants.GrantTypeAuthorizationCode),
		AccessTokenSubject:   "user123",
		AccessTokenAudience:  "app123",
		AccessTokenUserAttrs: map[string]interface{}{"name": testUserName},
		OAuthApp:             nil,
	}

	expectedToken := testRefreshToken
	expectedIat := time.Now().Unix()

	suite.mockJWTService.On("GenerateJWT",
		"test-client",
		"https://thunder.io",
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			_, hasUserAttrs := claims["access_token_user_attributes"]
			return !hasUserAttrs // Should not include user attrs when OAuthApp is nil
		}),
	).Return(expectedToken, expectedIat, nil)

	result, err := suite.builder.BuildRefreshToken(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenBuilderTestSuite) TestBuildRefreshToken_Success_EmptyScopes() {
	ctx := &RefreshTokenBuildContext{
		ClientID:             "test-client",
		Scopes:               []string{},
		GrantType:            string(constants.GrantTypeAuthorizationCode),
		AccessTokenSubject:   "user123",
		AccessTokenAudience:  "app123",
		AccessTokenUserAttrs: map[string]interface{}{},
		OAuthApp:             suite.oauthApp,
	}

	expectedToken := testRefreshToken
	expectedIat := time.Now().Unix()

	suite.mockJWTService.On("GenerateJWT",
		"test-client",
		"https://thunder.io",
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			_, hasScope := claims["scope"]
			return !hasScope
		}),
	).Return(expectedToken, expectedIat, nil)

	result, err := suite.builder.BuildRefreshToken(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenBuilderTestSuite) TestBuildRefreshToken_Success_CustomIssuer() {
	customOAuthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Token: &appmodel.OAuthTokenConfig{
			Issuer:      "https://custom.thunder.io",
			AccessToken: &appmodel.AccessTokenConfig{
				// AccessToken uses OAuth-level issuer
			},
		},
	}

	ctx := &RefreshTokenBuildContext{
		ClientID:             "test-client",
		Scopes:               []string{"read"},
		GrantType:            string(constants.GrantTypeAuthorizationCode),
		AccessTokenSubject:   "user123",
		AccessTokenAudience:  "app123",
		AccessTokenUserAttrs: map[string]interface{}{},
		OAuthApp:             customOAuthApp,
	}

	expectedToken := testRefreshToken
	expectedIat := time.Now().Unix()

	suite.mockJWTService.On("GenerateJWT",
		"test-client",
		"https://custom.thunder.io",
		"https://custom.thunder.io",
		int64(3600),
		mock.Anything,
	).Return(expectedToken, expectedIat, nil)

	result, err := suite.builder.BuildRefreshToken(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenBuilderTestSuite) TestBuildRefreshToken_Success_WithNilAccessToken() {
	oauthAppWithNilAccessToken := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Token: &appmodel.OAuthTokenConfig{
			// Token exists but AccessToken is nil
			AccessToken: nil,
		},
	}

	ctx := &RefreshTokenBuildContext{
		ClientID:             "test-client",
		Scopes:               []string{"read"},
		GrantType:            string(constants.GrantTypeAuthorizationCode),
		AccessTokenSubject:   "user123",
		AccessTokenAudience:  "app123",
		AccessTokenUserAttrs: map[string]interface{}{"name": testUserName},
		OAuthApp:             oauthAppWithNilAccessToken,
	}

	expectedToken := testRefreshToken
	expectedIat := time.Now().Unix()

	suite.mockJWTService.On("GenerateJWT",
		"test-client",
		"https://thunder.io",
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			_, hasUserAttrs := claims["access_token_user_attributes"]
			return !hasUserAttrs // Should not include user attrs when AccessToken is nil
		}),
	).Return(expectedToken, expectedIat, nil)

	result, err := suite.builder.BuildRefreshToken(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenBuilderTestSuite) TestBuildRefreshToken_Error_NilContext() {
	result, err := suite.builder.BuildRefreshToken(nil)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "build context cannot be nil")
}

func (suite *TokenBuilderTestSuite) TestBuildRefreshToken_Error_JWTGenerationFailed() {
	ctx := &RefreshTokenBuildContext{
		ClientID:             "test-client",
		Scopes:               []string{"read"},
		GrantType:            string(constants.GrantTypeAuthorizationCode),
		AccessTokenSubject:   "user123",
		AccessTokenAudience:  "app123",
		AccessTokenUserAttrs: map[string]interface{}{},
		OAuthApp:             suite.oauthApp,
	}

	suite.mockJWTService.On("GenerateJWT",
		"test-client",
		"https://thunder.io",
		"https://thunder.io",
		int64(3600),
		mock.Anything,
	).Return("", int64(0), &serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "JWT_GENERATION_FAILED",
		Error:            "JWT generation failed",
		ErrorDescription: "Failed to generate JWT token",
	})

	result, err := suite.builder.BuildRefreshToken(ctx)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "failed to generate refresh token")
	suite.mockJWTService.AssertExpectations(suite.T())
}

// ============================================================================
// BuildIDToken Tests - Success Cases
// ============================================================================

func (suite *TokenBuilderTestSuite) TestBuildIDToken_Success_Basic() {
	ctx := &IDTokenBuildContext{
		Subject:        "user123",
		Audience:       "app123",
		Scopes:         []string{"openid", "profile"},
		UserAttributes: map[string]interface{}{"sub": "user123", "name": testUserName},
		AuthTime:       time.Now().Unix(),
		OAuthApp:       suite.oauthApp,
	}

	expectedToken := testIDToken
	expectedIat := time.Now().Unix()

	suite.mockJWTService.On("GenerateJWT",
		"user123",
		"app123",
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			// sub is passed as first arg to GenerateJWT, not in claims map
			return claims["auth_time"] == ctx.AuthTime
		}),
	).Return(expectedToken, expectedIat, nil)

	result, err := suite.builder.BuildIDToken(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), expectedToken, result.Token)
	assert.Equal(suite.T(), "", result.TokenType) // ID tokens are not bearer tokens
	assert.Equal(suite.T(), expectedIat, result.IssuedAt)
	assert.Equal(suite.T(), int64(3600), result.ExpiresIn)
	assert.Equal(suite.T(), []string{"openid", "profile"}, result.Scopes)
	assert.Equal(suite.T(), "app123", result.ClientID)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenBuilderTestSuite) TestBuildIDToken_Success_NoAuthTime() {
	ctx := &IDTokenBuildContext{
		Subject:        "user123",
		Audience:       "app123",
		Scopes:         []string{"openid"},
		UserAttributes: map[string]interface{}{"sub": "user123"},
		AuthTime:       0,
		OAuthApp:       suite.oauthApp,
	}

	expectedToken := testIDToken
	expectedIat := time.Now().Unix()

	suite.mockJWTService.On("GenerateJWT",
		"user123",
		"app123",
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			_, hasAuthTime := claims["auth_time"]
			return !hasAuthTime
		}),
	).Return(expectedToken, expectedIat, nil)

	result, err := suite.builder.BuildIDToken(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenBuilderTestSuite) TestBuildIDToken_Success_WithScopeClaims() {
	oauthAppWithScopeClaims := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Token: &appmodel.OAuthTokenConfig{
			IDToken: &appmodel.IDTokenConfig{
				ValidityPeriod: 3600,
				UserAttributes: []string{"name", "email"},
			},
		},
		ScopeClaims: map[string][]string{
			"profile": {"name", "email"},
		},
	}

	ctx := &IDTokenBuildContext{
		Subject:        "user123",
		Audience:       "app123",
		Scopes:         []string{"openid", "profile"},
		UserAttributes: map[string]interface{}{"sub": "user123", "name": testUserName, "email": "john@example.com"},
		AuthTime:       time.Now().Unix(),
		OAuthApp:       oauthAppWithScopeClaims,
	}

	expectedToken := testIDToken
	expectedIat := time.Now().Unix()

	suite.mockJWTService.On("GenerateJWT",
		"user123",
		"app123",
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			return claims["name"] == testUserName && claims["email"] == "john@example.com"
		}),
	).Return(expectedToken, expectedIat, nil)

	result, err := suite.builder.BuildIDToken(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenBuilderTestSuite) TestBuildIDToken_Success_WithStandardOIDCScopes() {
	oauthAppWithUserAttrs := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Token: &appmodel.OAuthTokenConfig{
			IDToken: &appmodel.IDTokenConfig{
				ValidityPeriod: 3600,
				UserAttributes: []string{"name", "email"},
			},
		},
	}

	ctx := &IDTokenBuildContext{
		Subject:        "user123",
		Audience:       "app123",
		Scopes:         []string{"openid", "profile", "email"}, // Added email scope
		UserAttributes: map[string]interface{}{"sub": "user123", "name": testUserName, "email": "john@example.com"},
		AuthTime:       time.Now().Unix(),
		OAuthApp:       oauthAppWithUserAttrs,
	}

	expectedToken := testIDToken
	expectedIat := time.Now().Unix()

	suite.mockJWTService.On("GenerateJWT",
		"user123",
		"app123",
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			// Check that both name (from profile scope) and email (from email scope) are present
			return claims["name"] == testUserName && claims["email"] == "john@example.com"
		}),
	).Return(expectedToken, expectedIat, nil)

	result, err := suite.builder.BuildIDToken(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenBuilderTestSuite) TestBuildIDToken_Success_NoUserAttributes() {
	ctx := &IDTokenBuildContext{
		Subject:        "user123",
		Audience:       "app123",
		Scopes:         []string{"openid"},
		UserAttributes: nil,
		AuthTime:       time.Now().Unix(),
		OAuthApp:       suite.oauthApp,
	}

	expectedToken := testIDToken
	expectedIat := time.Now().Unix()

	suite.mockJWTService.On("GenerateJWT",
		"user123",
		"app123",
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			return claims["auth_time"] == ctx.AuthTime
		}),
	).Return(expectedToken, expectedIat, nil)

	result, err := suite.builder.BuildIDToken(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenBuilderTestSuite) TestBuildIDToken_Success_EmptyUserAttributes() {
	oauthAppWithEmptyUserAttrs := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Token: &appmodel.OAuthTokenConfig{
			IDToken: &appmodel.IDTokenConfig{
				UserAttributes: []string{},
			},
		},
	}

	ctx := &IDTokenBuildContext{
		Subject:        "user123",
		Audience:       "app123",
		Scopes:         []string{"openid", "profile"},
		UserAttributes: map[string]interface{}{"name": testUserName},
		AuthTime:       time.Now().Unix(),
		OAuthApp:       oauthAppWithEmptyUserAttrs,
	}

	expectedToken := testIDToken
	expectedIat := time.Now().Unix()

	suite.mockJWTService.On("GenerateJWT",
		"user123",
		"app123",
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			_, hasName := claims["name"]
			return claims["auth_time"] == ctx.AuthTime &&
				!hasName // Should not include name if not in UserAttributes config
		}),
	).Return(expectedToken, expectedIat, nil)

	result, err := suite.builder.BuildIDToken(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenBuilderTestSuite) TestBuildIDToken_Success_CustomValidityPeriod() {
	oauthAppWithCustomValidity := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Token: &appmodel.OAuthTokenConfig{
			IDToken: &appmodel.IDTokenConfig{
				ValidityPeriod: 7200,
			},
		},
	}

	ctx := &IDTokenBuildContext{
		Subject:        "user123",
		Audience:       "app123",
		Scopes:         []string{"openid"},
		UserAttributes: map[string]interface{}{"sub": "user123"},
		AuthTime:       time.Now().Unix(),
		OAuthApp:       oauthAppWithCustomValidity,
	}

	expectedToken := testIDToken
	expectedIat := time.Now().Unix()

	suite.mockJWTService.On("GenerateJWT",
		"user123",
		"app123",
		"https://thunder.io",
		int64(7200),
		mock.Anything,
	).Return(expectedToken, expectedIat, nil)

	result, err := suite.builder.BuildIDToken(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), int64(7200), result.ExpiresIn)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenBuilderTestSuite) TestBuildIDToken_Error_NilContext() {
	result, err := suite.builder.BuildIDToken(nil)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "build context cannot be nil")
}

func (suite *TokenBuilderTestSuite) TestBuildIDToken_Error_JWTGenerationFailed() {
	ctx := &IDTokenBuildContext{
		Subject:        "user123",
		Audience:       "app123",
		Scopes:         []string{"openid"},
		UserAttributes: map[string]interface{}{"sub": "user123"},
		AuthTime:       time.Now().Unix(),
		OAuthApp:       suite.oauthApp,
	}

	suite.mockJWTService.On("GenerateJWT",
		"user123",
		"app123",
		"https://thunder.io",
		int64(3600),
		mock.Anything,
	).Return("", int64(0), &serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "JWT_GENERATION_FAILED",
		Error:            "JWT generation failed",
		ErrorDescription: "Failed to generate JWT token",
	})

	result, err := suite.builder.BuildIDToken(ctx)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "failed to generate ID token")
	suite.mockJWTService.AssertExpectations(suite.T())
}
