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
	"reflect"
	"testing"
	"time"

	"github.com/asgardeo/thunder/internal/system/i18n/core"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	inboundmodel "github.com/asgardeo/thunder/pkg/inboundclient/model"
	"github.com/asgardeo/thunder/pkg/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/tests/mocks/jose/jwtmock"
)

const (
	testAccessToken  = "test-access-token"  //nolint:gosec // Test token, not a real credential
	testRefreshToken = "test-refresh-token" //nolint:gosec // Test token, not a real credential
	testIDToken      = "test-id-token"      //nolint:gosec // Test token, not a real credential
	testUserName     = "John Doe"
	testAppID        = "app123"
	testCacheID      = "test-cache-id"
)

type TokenBuilderTestSuite struct {
	suite.Suite
	mockJWTService *jwtmock.JWTServiceInterfaceMock
	builder        *tokenBuilder
	oauthApp       *inboundmodel.OAuthClient
}

func TestTokenBuilderTestSuite(t *testing.T) {
	suite.Run(t, new(TokenBuilderTestSuite))
}

func (suite *TokenBuilderTestSuite) SetupTest() {
	// Initialize Runtime for tests
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "https://thunder.io",
			ValidityPeriod: 3600,
		},
	}
	_ = config.InitializeServerRuntime("test", testConfig)

	suite.mockJWTService = jwtmock.NewJWTServiceInterfaceMock(suite.T())
	suite.builder = &tokenBuilder{
		jwtService: suite.mockJWTService,
	}

	suite.oauthApp = &inboundmodel.OAuthClient{
		ClientID: "test-client",
		Token: &inboundmodel.OAuthTokenConfig{
			AccessToken: &inboundmodel.AccessTokenConfig{
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
		Audiences:      []string{"app123"},
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
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			return claims["scope"] == "read write" &&
				claims["client_id"] == "test-client" &&
				claims["grant_type"] == string(constants.GrantTypeAuthorizationCode) &&
				claims["name"] == testUserName
		}), mock.Anything, mock.Anything,
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
		Aud:            nil,
		UserAttributes: map[string]interface{}{},
		NestedAct:      nil,
	}

	ctx := &AccessTokenBuildContext{
		Subject:        "user123",
		Audiences:      []string{"app123"},
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
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			act, ok := claims["act"].(map[string]interface{})
			return ok && act["sub"] == "actor123" && act["iss"] == "https://actor-issuer.com"
		}), mock.Anything, mock.Anything,
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
		Aud:            nil,
		UserAttributes: map[string]interface{}{},
		NestedAct: map[string]interface{}{
			"sub": "original-actor",
		},
	}

	ctx := &AccessTokenBuildContext{
		Subject:        "user123",
		Audiences:      []string{"app123"},
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
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			act, ok := claims["act"].(map[string]interface{})
			return ok && act["sub"] == "nested-actor" && act["act"] != nil
		}), mock.Anything, mock.Anything,
	).Return(expectedToken, expectedIat, nil)

	result, err := suite.builder.BuildAccessToken(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenBuilderTestSuite) TestBuildAccessToken_Success_EmptyScopes() {
	ctx := &AccessTokenBuildContext{
		Subject:        "user123",
		Audiences:      []string{"app123"},
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
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			_, hasScope := claims["scope"]
			return !hasScope
		}), mock.Anything, mock.Anything,
	).Return(expectedToken, expectedIat, nil)

	result, err := suite.builder.BuildAccessToken(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenBuilderTestSuite) TestBuildAccessToken_Success_EmptyClientID() {
	ctx := &AccessTokenBuildContext{
		Subject:        "user123",
		Audiences:      []string{"app123"},
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
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			_, hasClientID := claims["client_id"]
			return !hasClientID
		}), mock.Anything, mock.Anything,
	).Return(expectedToken, expectedIat, nil)

	result, err := suite.builder.BuildAccessToken(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenBuilderTestSuite) TestBuildAccessToken_Success_EmptyGrantType() {
	ctx := &AccessTokenBuildContext{
		Subject:        "user123",
		Audiences:      []string{"app123"},
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
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			_, hasGrantType := claims["grant_type"]
			return !hasGrantType
		}), mock.Anything, mock.Anything,
	).Return(expectedToken, expectedIat, nil)

	result, err := suite.builder.BuildAccessToken(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenBuilderTestSuite) TestBuildAccessToken_Success_CustomValidityPeriod() {
	customOAuthApp := &inboundmodel.OAuthClient{
		ClientID: "test-client",
		Token: &inboundmodel.OAuthTokenConfig{
			AccessToken: &inboundmodel.AccessTokenConfig{
				ValidityPeriod: 7200,
			},
		},
	}

	ctx := &AccessTokenBuildContext{
		Subject:        "user123",
		Audiences:      []string{"app123"},
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
		"https://thunder.io", // Server-level issuer always used
		int64(7200),
		mock.Anything, mock.Anything, mock.Anything,
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
		Audiences:      []string{"app123"},
		ClientID:       "test-client",
		Scopes:         []string{"read"},
		UserAttributes: map[string]interface{}{},
		GrantType:      string(constants.GrantTypeAuthorizationCode),
		OAuthApp:       suite.oauthApp,
	}

	suite.mockJWTService.On("GenerateJWT",
		"user123",
		"https://thunder.io",
		int64(3600),
		mock.Anything, mock.Anything, mock.Anything,
	).Return("", int64(0), &serviceerror.ServiceError{
		Type: serviceerror.ServerErrorType,
		Code: "JWT_GENERATION_FAILED",
		Error: core.I18nMessage{
			Key: "error.test.jwt_generation_failed", DefaultValue: "JWT generation failed",
		},
		ErrorDescription: core.I18nMessage{
			Key: "error.test.failed_to_generate_jwt_token", DefaultValue: "Failed to generate JWT token",
		},
	})

	result, err := suite.builder.BuildAccessToken(ctx)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "failed to generate access token")
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenBuilderTestSuite) TestBuildAccessToken_Success_WithClaimsLocales() {
	ctx := &AccessTokenBuildContext{
		Subject:        "user123",
		Audiences:      []string{"app123"},
		ClientID:       "test-client",
		Scopes:         []string{"openid", "profile"},
		UserAttributes: map[string]interface{}{"name": testUserName},
		GrantType:      string(constants.GrantTypeAuthorizationCode),
		OAuthApp:       suite.oauthApp,
		ClaimsLocales:  "en-US fr-CA ja",
	}

	expectedToken := testAccessToken
	expectedIat := time.Now().Unix()

	suite.mockJWTService.On("GenerateJWT",
		"user123",
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			return claims["scope"] == "openid profile" &&
				claims["client_id"] == "test-client" &&
				claims["grant_type"] == string(constants.GrantTypeAuthorizationCode) &&
				claims["name"] == testUserName &&
				claims["claims_locales"] == "en-US fr-CA ja"
		}), mock.Anything, mock.Anything,
	).Return(expectedToken, expectedIat, nil)

	result, err := suite.builder.BuildAccessToken(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), expectedToken, result.Token)
	assert.Equal(suite.T(), "en-US fr-CA ja", result.ClaimsLocales)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenBuilderTestSuite) TestBuildRefreshToken_Success_Basic() {
	// Create OAuth app with user attributes configured
	oauthAppWithUserAttrs := &inboundmodel.OAuthClient{
		ClientID: "test-client",
		Token: &inboundmodel.OAuthTokenConfig{
			AccessToken: &inboundmodel.AccessTokenConfig{
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
		AccessTokenAudiences: []string{"app123"},
		AttributeCacheID:     testCacheID,
		OAuthApp:             oauthAppWithUserAttrs,
	}

	expectedToken := testRefreshToken
	expectedIat := time.Now().Unix()

	suite.mockJWTService.On("GenerateJWT",
		"test-client",
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			return claims["scope"] == "read write" &&
				claims["access_token_sub"] == "user123" &&
				reflect.DeepEqual(claims["access_token_aud"], []string{testAppID}) &&
				claims["grant_type"] == string(constants.GrantTypeAuthorizationCode) &&
				claims["aci"] == testCacheID
		}), mock.Anything, mock.Anything,
	).Return(expectedToken, expectedIat, nil)

	result, err := suite.builder.BuildRefreshToken(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), expectedToken, result.Token)
	assert.Equal(suite.T(), expectedIat, result.IssuedAt)
	assert.Equal(suite.T(), int64(3600), result.ExpiresIn)
	assert.Equal(suite.T(), []string{"read", "write"}, result.Scopes)
	assert.Equal(suite.T(), "test-client", result.ClientID)
	assert.Equal(suite.T(), []string{"https://thunder.io"}, result.Audiences)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenBuilderTestSuite) TestBuildRefreshToken_Success_WithoutUserAttributes() {
	ctx := &RefreshTokenBuildContext{
		ClientID:             "test-client",
		Scopes:               []string{"read"},
		GrantType:            string(constants.GrantTypeAuthorizationCode),
		AccessTokenSubject:   "user123",
		AccessTokenAudiences: []string{"app123"},
		AttributeCacheID:     "",
		OAuthApp:             suite.oauthApp,
	}

	expectedToken := testRefreshToken
	expectedIat := time.Now().Unix()

	suite.mockJWTService.On("GenerateJWT",
		"test-client",
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			_, hasAttrCacheID := claims["aci"]
			return !hasAttrCacheID
		}), mock.Anything, mock.Anything,
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
		AccessTokenAudiences: []string{"app123"},
		AttributeCacheID:     testCacheID,
		OAuthApp:             nil,
	}

	expectedToken := testRefreshToken
	expectedIat := time.Now().Unix()

	suite.mockJWTService.On("GenerateJWT",
		"test-client",
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			return claims["aci"] == testCacheID
		}), mock.Anything, mock.Anything,
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
		AccessTokenAudiences: []string{"app123"},
		AttributeCacheID:     "",
		OAuthApp:             suite.oauthApp,
	}

	expectedToken := testRefreshToken
	expectedIat := time.Now().Unix()

	suite.mockJWTService.On("GenerateJWT",
		"test-client",
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			_, hasScope := claims["scope"]
			return !hasScope
		}), mock.Anything, mock.Anything,
	).Return(expectedToken, expectedIat, nil)

	result, err := suite.builder.BuildRefreshToken(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenBuilderTestSuite) TestBuildRefreshToken_Success_WithTokenConfig() {
	customOAuthApp := &inboundmodel.OAuthClient{
		ClientID: "test-client",
		Token: &inboundmodel.OAuthTokenConfig{
			AccessToken: &inboundmodel.AccessTokenConfig{},
		},
	}

	ctx := &RefreshTokenBuildContext{
		ClientID:             "test-client",
		Scopes:               []string{"read"},
		GrantType:            string(constants.GrantTypeAuthorizationCode),
		AccessTokenSubject:   "user123",
		AccessTokenAudiences: []string{"app123"},
		AttributeCacheID:     "",
		OAuthApp:             customOAuthApp,
	}

	expectedToken := testRefreshToken
	expectedIat := time.Now().Unix()

	suite.mockJWTService.On("GenerateJWT",
		"test-client",
		"https://thunder.io",
		int64(3600),
		mock.Anything, mock.Anything, mock.Anything,
	).Return(expectedToken, expectedIat, nil)

	result, err := suite.builder.BuildRefreshToken(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenBuilderTestSuite) TestBuildRefreshToken_Success_WithNilAccessToken() {
	oauthAppWithNilAccessToken := &inboundmodel.OAuthClient{
		ClientID: "test-client",
		Token: &inboundmodel.OAuthTokenConfig{
			// Token exists but AccessToken is nil
			AccessToken: nil,
		},
	}

	ctx := &RefreshTokenBuildContext{
		ClientID:             "test-client",
		Scopes:               []string{"read"},
		GrantType:            string(constants.GrantTypeAuthorizationCode),
		AccessTokenSubject:   "user123",
		AccessTokenAudiences: []string{"app123"},
		AttributeCacheID:     testCacheID,
		OAuthApp:             oauthAppWithNilAccessToken,
	}

	expectedToken := testRefreshToken
	expectedIat := time.Now().Unix()

	suite.mockJWTService.On("GenerateJWT",
		"test-client",
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			return claims["aci"] == testCacheID
		}), mock.Anything, mock.Anything,
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
		AccessTokenAudiences: []string{"app123"},
		AttributeCacheID:     "",
		OAuthApp:             suite.oauthApp,
	}

	suite.mockJWTService.On("GenerateJWT",
		"test-client",
		"https://thunder.io",
		int64(3600),
		mock.Anything, mock.Anything, mock.Anything,
	).Return("", int64(0), &serviceerror.ServiceError{
		Type: serviceerror.ServerErrorType,
		Code: "JWT_GENERATION_FAILED",
		Error: core.I18nMessage{
			Key: "error.test.jwt_generation_failed", DefaultValue: "JWT generation failed",
		},
		ErrorDescription: core.I18nMessage{
			Key: "error.test.failed_to_generate_jwt_token", DefaultValue: "Failed to generate JWT token",
		},
	})

	result, err := suite.builder.BuildRefreshToken(ctx)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "failed to generate refresh token")
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenBuilderTestSuite) TestBuildRefreshToken_Success_WithClaimsLocales() {
	ctx := &RefreshTokenBuildContext{
		ClientID:             "test-client",
		Scopes:               []string{"openid", "profile"},
		GrantType:            string(constants.GrantTypeAuthorizationCode),
		AccessTokenSubject:   "user123",
		AccessTokenAudiences: []string{"app123"},
		AttributeCacheID:     "",
		OAuthApp:             suite.oauthApp,
		ClaimsLocales:        "en-US fr-CA ja",
	}

	expectedToken := testRefreshToken
	expectedIat := time.Now().Unix()

	suite.mockJWTService.On("GenerateJWT",
		"test-client",
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			return claims["scope"] == "openid profile" &&
				claims["access_token_sub"] == "user123" &&
				reflect.DeepEqual(claims["access_token_aud"], []string{testAppID}) &&
				claims["grant_type"] == string(constants.GrantTypeAuthorizationCode) &&
				claims["access_token_claims_locales"] == "en-US fr-CA ja"
		}), mock.Anything, mock.Anything,
	).Return(expectedToken, expectedIat, nil)

	result, err := suite.builder.BuildRefreshToken(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), expectedToken, result.Token)
	assert.Equal(suite.T(), "en-US fr-CA ja", result.ClaimsLocales)
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
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			// sub is passed as first arg to GenerateJWT, not in claims map
			return claims["auth_time"] == ctx.AuthTime
		}), mock.Anything, mock.Anything,
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

func (suite *TokenBuilderTestSuite) TestBuildIDToken_Success_WithNonce() {
	ctx := &IDTokenBuildContext{
		Subject:        "user123",
		Audience:       "app123",
		Scopes:         []string{"openid"},
		UserAttributes: map[string]interface{}{"sub": "user123"},
		AuthTime:       time.Now().Unix(),
		OAuthApp:       suite.oauthApp,
		Nonce:          "test-nonce-123",
	}

	expectedToken := testIDToken
	expectedIat := time.Now().Unix()

	suite.mockJWTService.On("GenerateJWT",
		"user123",
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			return claims["nonce"] == "test-nonce-123"
		}), mock.Anything, mock.Anything,
	).Return(expectedToken, expectedIat, nil)

	result, err := suite.builder.BuildIDToken(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenBuilderTestSuite) TestBuildIDToken_Success_WithoutNonce() {
	ctx := &IDTokenBuildContext{
		Subject:        "user123",
		Audience:       "app123",
		Scopes:         []string{"openid"},
		UserAttributes: map[string]interface{}{"sub": "user123"},
		AuthTime:       time.Now().Unix(),
		OAuthApp:       suite.oauthApp,
		Nonce:          "",
	}

	expectedToken := testIDToken
	expectedIat := time.Now().Unix()

	suite.mockJWTService.On("GenerateJWT",
		"user123",
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			_, exists := claims["nonce"]
			return !exists
		}), mock.Anything, mock.Anything,
	).Return(expectedToken, expectedIat, nil)

	result, err := suite.builder.BuildIDToken(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
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
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			_, hasAuthTime := claims["auth_time"]
			return !hasAuthTime
		}), mock.Anything, mock.Anything,
	).Return(expectedToken, expectedIat, nil)

	result, err := suite.builder.BuildIDToken(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenBuilderTestSuite) TestBuildIDToken_Success_WithScopeClaims() {
	oauthAppWithScopeClaims := &inboundmodel.OAuthClient{
		ClientID: "test-client",
		Token: &inboundmodel.OAuthTokenConfig{
			IDToken: &inboundmodel.IDTokenConfig{
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
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			return claims["name"] == testUserName && claims["email"] == "john@example.com"
		}), mock.Anything, mock.Anything,
	).Return(expectedToken, expectedIat, nil)

	result, err := suite.builder.BuildIDToken(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenBuilderTestSuite) TestBuildIDToken_Success_WithStandardOIDCScopes() {
	oauthAppWithUserAttrs := &inboundmodel.OAuthClient{
		ClientID: "test-client",
		Token: &inboundmodel.OAuthTokenConfig{
			IDToken: &inboundmodel.IDTokenConfig{
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
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			// Check that both name (from profile scope) and email (from email scope) are present
			return claims["name"] == testUserName && claims["email"] == "john@example.com"
		}), mock.Anything, mock.Anything,
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
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			return claims["auth_time"] == ctx.AuthTime
		}), mock.Anything, mock.Anything,
	).Return(expectedToken, expectedIat, nil)

	result, err := suite.builder.BuildIDToken(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenBuilderTestSuite) TestBuildIDToken_Success_EmptyUserAttributes() {
	oauthAppWithEmptyUserAttrs := &inboundmodel.OAuthClient{
		ClientID: "test-client",
		Token: &inboundmodel.OAuthTokenConfig{
			IDToken: &inboundmodel.IDTokenConfig{
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
		"https://thunder.io",
		int64(3600),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			_, hasName := claims["name"]
			return claims["auth_time"] == ctx.AuthTime &&
				!hasName // Should not include name if not in UserAttributes config
		}), mock.Anything, mock.Anything,
	).Return(expectedToken, expectedIat, nil)

	result, err := suite.builder.BuildIDToken(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *TokenBuilderTestSuite) TestBuildIDToken_Success_CustomValidityPeriod() {
	oauthAppWithCustomValidity := &inboundmodel.OAuthClient{
		ClientID: "test-client",
		Token: &inboundmodel.OAuthTokenConfig{
			IDToken: &inboundmodel.IDTokenConfig{
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
		"https://thunder.io",
		int64(7200),
		mock.Anything, mock.Anything, mock.Anything,
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
		"https://thunder.io",
		int64(3600),
		mock.Anything, mock.Anything, mock.Anything,
	).Return("", int64(0), &serviceerror.ServiceError{
		Type: serviceerror.ServerErrorType,
		Code: "JWT_GENERATION_FAILED",
		Error: core.I18nMessage{
			Key: "error.test.jwt_generation_failed", DefaultValue: "JWT generation failed",
		},
		ErrorDescription: core.I18nMessage{
			Key: "error.test.failed_to_generate_jwt_token", DefaultValue: "Failed to generate JWT token",
		},
	})

	result, err := suite.builder.BuildIDToken(ctx)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "failed to generate ID token")
	suite.mockJWTService.AssertExpectations(suite.T())
}
