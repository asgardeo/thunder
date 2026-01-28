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
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
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
	"github.com/asgardeo/thunder/tests/mocks/jwtmock"
	"github.com/asgardeo/thunder/tests/mocks/oauth/oauth2/tokenservicemock"
)

const (
	testTokenExchangeJWT = "test-token-exchange-jwt" //nolint:gosec
	testScopeReadWrite   = "read write"
	testCustomIssuer     = "https://custom.issuer.com"
	testUserEmail        = "user@example.com"
	testClientID         = "client123"
	testUserID           = "user123"
	testScopeRead        = "read"
)

type TokenExchangeGrantHandlerTestSuite struct {
	suite.Suite
	mockJWTService     *jwtmock.JWTServiceInterfaceMock
	mockTokenBuilder   *tokenservicemock.TokenBuilderInterfaceMock
	mockTokenValidator *tokenservicemock.TokenValidatorInterfaceMock
	handler            *tokenExchangeGrantHandler
	oauthApp           *appmodel.OAuthAppConfigProcessedDTO
}

func TestTokenExchangeGrantHandlerSuite(t *testing.T) {
	suite.Run(t, new(TokenExchangeGrantHandlerTestSuite))
}

func (suite *TokenExchangeGrantHandlerTestSuite) SetupTest() {
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "https://test.thunder.io",
			ValidityPeriod: 3600,
			Audience:       "application", // Default audience for tests
		},
	}
	err := config.InitializeThunderRuntime("", testConfig)
	assert.NoError(suite.T(), err)

	suite.mockJWTService = jwtmock.NewJWTServiceInterfaceMock(suite.T())
	suite.mockTokenBuilder = tokenservicemock.NewTokenBuilderInterfaceMock(suite.T())
	suite.mockTokenValidator = tokenservicemock.NewTokenValidatorInterfaceMock(suite.T())
	suite.handler = &tokenExchangeGrantHandler{
		tokenBuilder:   suite.mockTokenBuilder,
		tokenValidator: suite.mockTokenValidator,
	}

	suite.oauthApp = &appmodel.OAuthAppConfigProcessedDTO{
		AppID:                   "app123",
		ClientID:                testClientID,
		HashedClientSecret:      "hashedsecret123",
		RedirectURIs:            []string{"https://example.com/callback"},
		GrantTypes:              []constants.GrantType{constants.GrantTypeTokenExchange},
		ResponseTypes:           []constants.ResponseType{constants.ResponseTypeCode},
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretBasic,
		Token: &appmodel.OAuthTokenConfig{
			AccessToken: &appmodel.AccessTokenConfig{
				ValidityPeriod: 7200,
			},
		},
	}
}

// getDefaultAudience is a helper function to get the configured default audience from runtime.
func (suite *TokenExchangeGrantHandlerTestSuite) getDefaultAudience() string {
	runtime := config.GetThunderRuntime()
	if runtime == nil {
		suite.T().Skip("ThunderRuntime not initialized")
		return ""
	}
	defaultAudience := runtime.Config.JWT.Audience
	if defaultAudience == "" {
		suite.T().Skip("Default audience not configured in runtime")
		return ""
	}
	return defaultAudience
}

// Helper function to create a test JWT token
func (suite *TokenExchangeGrantHandlerTestSuite) createTestJWT(claims map[string]interface{}) string {
	header := map[string]interface{}{
		"alg": "RS256",
		"typ": "JWT",
	}

	headerJSON, _ := json.Marshal(header)
	claimsJSON, _ := json.Marshal(claims)

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	return fmt.Sprintf("%s.%s.signature", headerB64, claimsB64)
}

// Helper function to create a basic token request for testing
func (suite *TokenExchangeGrantHandlerTestSuite) createBasicTokenRequest(subjectToken string) *model.TokenRequest {
	return &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         testClientID,
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
	}
}

// Helper function to setup token validator and token builder mocks for successful token generation with audience check
func (suite *TokenExchangeGrantHandlerTestSuite) setupSuccessfulJWTMock(
	subjectToken string,
	expectedAudience string,
	now int64,
) {
	suite.mockTokenValidator.On("ValidateSubjectToken", mock.Anything, subjectToken, suite.oauthApp).
		Return(&tokenservice.SubjectTokenClaims{
			Sub:            testUserID,
			Iss:            testCustomIssuer,
			Scopes:         []string{"read", "write"},
			UserAttributes: map[string]interface{}{},
			NestedAct:      nil,
		}, nil)
	suite.mockTokenBuilder.On("BuildAccessToken", mock.Anything,
		mock.MatchedBy(func(ctx *tokenservice.AccessTokenBuildContext) bool {
			return ctx.Subject == testUserID && ctx.Audience == expectedAudience
		})).Return(&model.TokenDTO{
		Token:     testTokenExchangeJWT,
		TokenType: constants.TokenTypeBearer,
		IssuedAt:  now,
		ExpiresIn: 7200,
		Scopes:    []string{"read", "write"},
		ClientID:  testClientID,
	}, nil)
}

// Helper function to setup token validator and token builder mocks for successful token generation with scope check
func (suite *TokenExchangeGrantHandlerTestSuite) setupSuccessfulJWTMockWithScope(
	subjectToken string,
	expectedAudience string,
	expectedScope string,
	now int64,
) {
	suite.mockTokenValidator.On("ValidateSubjectToken", mock.Anything, subjectToken, suite.oauthApp).
		Return(&tokenservice.SubjectTokenClaims{
			Sub:            testUserID,
			Iss:            testCustomIssuer,
			Scopes:         tokenservice.ParseScopes(expectedScope),
			UserAttributes: map[string]interface{}{},
			NestedAct:      nil,
		}, nil)
	suite.mockTokenBuilder.On("BuildAccessToken", mock.Anything,
		mock.MatchedBy(func(ctx *tokenservice.AccessTokenBuildContext) bool {
			return ctx.Subject == testUserID && ctx.Audience == expectedAudience &&
				tokenservice.JoinScopes(ctx.Scopes) == expectedScope
		})).Return(&model.TokenDTO{
		Token:     testTokenExchangeJWT,
		TokenType: constants.TokenTypeBearer,
		IssuedAt:  now,
		ExpiresIn: 7200,
		Scopes:    tokenservice.ParseScopes(expectedScope),
		ClientID:  testClientID,
	}, nil)
}

// TestNewTokenExchangeGrantHandler tests the constructor
func (suite *TokenExchangeGrantHandlerTestSuite) TestNewTokenExchangeGrantHandler() {
	handler := newTokenExchangeGrantHandler(suite.mockTokenBuilder, suite.mockTokenValidator)
	assert.NotNil(suite.T(), handler)
	assert.Implements(suite.T(), (*GrantHandlerInterface)(nil), handler)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestValidateGrant_Success() {
	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         testClientID,
		ClientSecret:     "secret123",
		SubjectToken:     "subject-token",
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
	}

	result := suite.handler.ValidateGrant(context.TODO(), tokenRequest, suite.oauthApp)
	assert.Nil(suite.T(), result)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestValidateGrant_WrongGrantType() {
	tokenRequest := &model.TokenRequest{
		GrantType:        "authorization_code",
		ClientID:         testClientID,
		SubjectToken:     "subject-token",
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
	}

	result := suite.handler.ValidateGrant(context.TODO(), tokenRequest, suite.oauthApp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), constants.ErrorUnsupportedGrantType, result.Error)
	assert.Equal(suite.T(), "Unsupported grant type", result.ErrorDescription)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestValidateGrant_MissingSubjectToken() {
	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         testClientID,
		SubjectToken:     "",
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
	}

	result := suite.handler.ValidateGrant(context.TODO(), tokenRequest, suite.oauthApp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, result.Error)
	assert.Equal(suite.T(), "Missing required parameter: subject_token", result.ErrorDescription)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestValidateGrant_MissingSubjectTokenType() {
	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         testClientID,
		SubjectToken:     "subject-token",
		SubjectTokenType: "",
	}

	result := suite.handler.ValidateGrant(context.TODO(), tokenRequest, suite.oauthApp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, result.Error)
	assert.Equal(suite.T(), "Missing required parameter: subject_token_type", result.ErrorDescription)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestValidateGrant_UnsupportedSubjectTokenType() {
	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         testClientID,
		SubjectToken:     "subject-token",
		SubjectTokenType: "urn:ietf:params:oauth:token-type:saml2",
	}

	result := suite.handler.ValidateGrant(context.TODO(), tokenRequest, suite.oauthApp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, result.Error)
	assert.Contains(suite.T(), result.ErrorDescription, "Unsupported subject_token_type")
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestValidateGrant_MissingActorTokenType() {
	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         testClientID,
		SubjectToken:     "subject-token",
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
		ActorToken:       "actor-token",
		ActorTokenType:   "",
	}

	result := suite.handler.ValidateGrant(context.TODO(), tokenRequest, suite.oauthApp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, result.Error)
	assert.Equal(suite.T(), "actor_token_type is required when actor_token is provided", result.ErrorDescription)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestValidateGrant_UnsupportedActorTokenType() {
	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         testClientID,
		SubjectToken:     "subject-token",
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
		ActorToken:       "actor-token",
		ActorTokenType:   "urn:ietf:params:oauth:token-type:saml1",
	}

	result := suite.handler.ValidateGrant(context.TODO(), tokenRequest, suite.oauthApp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, result.Error)
	assert.Contains(suite.T(), result.ErrorDescription, "Unsupported actor_token_type")
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestValidateGrant_ActorTokenTypeWithoutActorToken() {
	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         testClientID,
		SubjectToken:     "subject-token",
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
		ActorToken:       "",
		ActorTokenType:   string(constants.TokenTypeIdentifierAccessToken),
	}

	result := suite.handler.ValidateGrant(context.TODO(), tokenRequest, suite.oauthApp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, result.Error)
	assert.Equal(suite.T(), "actor_token_type must not be provided without actor_token", result.ErrorDescription)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestValidateGrant_InvalidResourceURI() {
	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         testClientID,
		SubjectToken:     "subject-token",
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
		Resource:         "not-a-valid-uri",
	}

	result := suite.handler.ValidateGrant(context.TODO(), tokenRequest, suite.oauthApp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, result.Error)
	assert.Contains(suite.T(), result.ErrorDescription, "Invalid resource parameter")
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestValidateGrant_ResourceURIWithFragment() {
	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         testClientID,
		SubjectToken:     "subject-token",
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
		Resource:         "https://api.example.com/resource#fragment",
	}

	result := suite.handler.ValidateGrant(context.TODO(), tokenRequest, suite.oauthApp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, result.Error)
	assert.Contains(suite.T(), result.ErrorDescription, "must not contain a fragment component")
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestValidateGrant_ValidResourceURI() {
	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         testClientID,
		SubjectToken:     "subject-token",
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
		Resource:         "https://api.example.com/resource",
	}

	result := suite.handler.ValidateGrant(context.TODO(), tokenRequest, suite.oauthApp)
	assert.Nil(suite.T(), result)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestValidateGrant_UnsupportedRequestedTokenType() {
	tokenRequest := &model.TokenRequest{
		GrantType:          string(constants.GrantTypeTokenExchange),
		ClientID:           testClientID,
		SubjectToken:       "subject-token",
		SubjectTokenType:   string(constants.TokenTypeIdentifierAccessToken),
		RequestedTokenType: "urn:ietf:params:oauth:token-type:saml2",
	}

	result := suite.handler.ValidateGrant(context.TODO(), tokenRequest, suite.oauthApp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, result.Error)
	assert.Contains(suite.T(), result.ErrorDescription, "Unsupported requested_token_type")
}

// ============================================================================
// HandleGrant Tests - Success Cases
// ============================================================================

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_Success_Basic() {
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub":   "user123",
		"iss":   testCustomIssuer,
		"aud":   "app123",
		"exp":   float64(now + 3600),
		"nbf":   float64(now - 60),
		"scope": "read write",
		"email": "user@example.com",
	})

	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         testClientID,
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
	}

	suite.mockTokenValidator.On("ValidateSubjectToken", mock.Anything, subjectToken, suite.oauthApp).
		Return(&tokenservice.SubjectTokenClaims{
			Sub:            testUserID,
			Iss:            testCustomIssuer,
			Scopes:         []string{"read", "write"},
			UserAttributes: map[string]interface{}{"email": testUserEmail},
			NestedAct:      nil,
		}, nil)
	suite.mockTokenBuilder.On("BuildAccessToken", mock.Anything,
		mock.MatchedBy(func(ctx *tokenservice.AccessTokenBuildContext) bool {
			return ctx.Subject == testUserID &&
				ctx.Audience == testClientID && // Default audience is clientID when no resource/audience parameter
				ctx.ClientID == testClientID &&
				ctx.UserAttributes["email"] == testUserEmail &&
				tokenservice.JoinScopes(ctx.Scopes) == testScopeReadWrite
		})).Return(&model.TokenDTO{
		Token:     testTokenExchangeJWT,
		TokenType: constants.TokenTypeBearer,
		IssuedAt:  now,
		ExpiresIn: 7200,
		Scopes:    []string{"read", "write"},
		ClientID:  testClientID,
		UserAttributes: map[string]interface{}{
			"email": "user@example.com",
		},
	}, nil)

	result, errResp := suite.handler.HandleGrant(context.TODO(), tokenRequest, suite.oauthApp)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), testTokenExchangeJWT, result.AccessToken.Token)
	assert.Equal(suite.T(), constants.TokenTypeBearer, result.AccessToken.TokenType)
	assert.Equal(suite.T(), int64(7200), result.AccessToken.ExpiresIn)
	assert.Equal(suite.T(), []string{"read", "write"}, result.AccessToken.Scopes)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_Success_WithScopeDownscoping() {
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub":   "user123",
		"iss":   testCustomIssuer,
		"exp":   float64(now + 3600),
		"nbf":   float64(now - 60),
		"scope": "read write delete",
	})

	tokenRequest := suite.createBasicTokenRequest(subjectToken)
	tokenRequest.Scope = testScopeReadWrite

	suite.setupSuccessfulJWTMockWithScope(subjectToken, testClientID, testScopeReadWrite, now)

	result, errResp := suite.handler.HandleGrant(context.TODO(), tokenRequest, suite.oauthApp)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), []string{"read", "write"}, result.AccessToken.Scopes)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_Success_WithActorToken() {
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub": "user123",
		"iss": testCustomIssuer,
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
	})

	actorToken := suite.createTestJWT(map[string]interface{}{
		"sub": "service456",
		"iss": testCustomIssuer,
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
	})

	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         testClientID,
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
		ActorToken:       actorToken,
		ActorTokenType:   string(constants.TokenTypeIdentifierAccessToken),
	}

	suite.mockTokenValidator.On("ValidateSubjectToken", mock.Anything, subjectToken, suite.oauthApp).
		Return(&tokenservice.SubjectTokenClaims{
			Sub:            testUserID,
			Iss:            testCustomIssuer,
			UserAttributes: map[string]interface{}{},
			NestedAct:      nil,
		}, nil)
	suite.mockTokenValidator.On("ValidateSubjectToken", mock.Anything, actorToken, suite.oauthApp).
		Return(&tokenservice.SubjectTokenClaims{
			Sub:            "service456",
			Iss:            testCustomIssuer,
			UserAttributes: map[string]interface{}{},
			NestedAct:      nil,
		}, nil)
	suite.mockTokenBuilder.On("BuildAccessToken", mock.Anything,
		mock.MatchedBy(func(ctx *tokenservice.AccessTokenBuildContext) bool {
			return ctx.Subject == testUserID &&
				ctx.Audience == testClientID &&
				ctx.ActorClaims != nil &&
				ctx.ActorClaims.Sub == "service456" &&
				ctx.ActorClaims.Iss == testCustomIssuer
		})).Return(&model.TokenDTO{
		Token:     testTokenExchangeJWT,
		TokenType: constants.TokenTypeBearer,
		IssuedAt:  now,
		ExpiresIn: 7200,
		Scopes:    []string{},
		ClientID:  testClientID,
	}, nil)

	result, errResp := suite.handler.HandleGrant(context.TODO(), tokenRequest, suite.oauthApp)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_Success_WithActorChaining() {
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub": "user123",
		"iss": testCustomIssuer,
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
		"act": map[string]interface{}{
			"sub": "service789",
			"iss": "https://existing-actor.com",
		},
	})

	actorToken := suite.createTestJWT(map[string]interface{}{
		"sub": "service456",
		"iss": testCustomIssuer,
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
	})

	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         testClientID,
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
		ActorToken:       actorToken,
		ActorTokenType:   string(constants.TokenTypeIdentifierAccessToken),
	}

	suite.mockTokenValidator.On("ValidateSubjectToken", mock.Anything, subjectToken, suite.oauthApp).
		Return(&tokenservice.SubjectTokenClaims{
			Sub:            "user123",
			Iss:            testCustomIssuer,
			UserAttributes: map[string]interface{}{},
			NestedAct: map[string]interface{}{
				"sub": "service789",
				"iss": "https://existing-actor.com",
			},
		}, nil)
	suite.mockTokenValidator.On("ValidateSubjectToken", mock.Anything, actorToken, suite.oauthApp).
		Return(&tokenservice.SubjectTokenClaims{
			Sub:            "service456",
			Iss:            testCustomIssuer,
			UserAttributes: map[string]interface{}{},
			NestedAct:      nil,
		}, nil)
	suite.mockTokenBuilder.On("BuildAccessToken", mock.Anything,
		mock.MatchedBy(func(ctx *tokenservice.AccessTokenBuildContext) bool {
			if ctx.ActorClaims == nil {
				return false
			}
			// Check new actor (from actor token)
			return ctx.ActorClaims.Sub == "service456" && ctx.ActorClaims.Iss == testCustomIssuer
		})).Return(&model.TokenDTO{
		Token:     testTokenExchangeJWT,
		TokenType: constants.TokenTypeBearer,
		IssuedAt:  now,
		ExpiresIn: 7200,
		Scopes:    []string{},
		ClientID:  testClientID,
	}, nil)

	result, errResp := suite.handler.HandleGrant(context.TODO(), tokenRequest, suite.oauthApp)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_Success_WithAudienceParameter() {
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub": "user123",
		"iss": testCustomIssuer,
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
	})

	tokenRequest := suite.createBasicTokenRequest(subjectToken)
	tokenRequest.Audience = "https://api.example.com"

	suite.setupSuccessfulJWTMock(subjectToken, "https://api.example.com", now)

	result, errResp := suite.handler.HandleGrant(context.TODO(), tokenRequest, suite.oauthApp)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_Success_WithResourceParameter() {
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub": "user123",
		"iss": testCustomIssuer,
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
	})

	tokenRequest := suite.createBasicTokenRequest(subjectToken)
	tokenRequest.Resource = "https://resource.example.com"

	suite.setupSuccessfulJWTMock(subjectToken, "https://resource.example.com", now)

	result, errResp := suite.handler.HandleGrant(context.TODO(), tokenRequest, suite.oauthApp)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_Success_WithMultipleSpacesInScope() {
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub":   "user123",
		"iss":   testCustomIssuer,
		"exp":   float64(now + 3600),
		"nbf":   float64(now - 60),
		"scope": "read write",
	})

	tokenRequest := suite.createBasicTokenRequest(subjectToken)
	tokenRequest.Scope = "  read    write  "

	suite.setupSuccessfulJWTMockWithScope(subjectToken, testClientID, testScopeReadWrite, now)

	result, errResp := suite.handler.HandleGrant(context.TODO(), tokenRequest, suite.oauthApp)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), []string{"read", "write"}, result.AccessToken.Scopes)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_Success_PreservesUserAttributes() {
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub":   "user123",
		"iss":   testCustomIssuer,
		"exp":   float64(now + 3600),
		"nbf":   float64(now - 60),
		"email": "user@example.com",
		"name":  "Test User",
		"roles": []string{"admin", "user"},
	})

	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         testClientID,
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
	}

	suite.mockTokenValidator.On("ValidateSubjectToken", mock.Anything, subjectToken, suite.oauthApp).
		Return(&tokenservice.SubjectTokenClaims{
			Sub:    testUserID,
			Iss:    testCustomIssuer,
			Scopes: []string{},
			UserAttributes: map[string]interface{}{
				"email": testUserEmail,
				"name":  "Test User",
				"roles": []string{"admin", "user"},
			},
			NestedAct: nil,
		}, nil)
	suite.mockTokenBuilder.On("BuildAccessToken", mock.Anything,
		mock.MatchedBy(func(ctx *tokenservice.AccessTokenBuildContext) bool {
			return ctx.Subject == testUserID &&
				ctx.Audience == testClientID &&
				ctx.UserAttributes["email"] == "user@example.com" &&
				ctx.UserAttributes["name"] == "Test User"
		})).Return(&model.TokenDTO{
		Token:     testTokenExchangeJWT,
		TokenType: constants.TokenTypeBearer,
		IssuedAt:  now,
		ExpiresIn: 7200,
		Scopes:    []string{},
		ClientID:  testClientID,
		UserAttributes: map[string]interface{}{
			"email": "user@example.com",
			"name":  "Test User",
		},
	}, nil)

	result, errResp := suite.handler.HandleGrant(context.TODO(), tokenRequest, suite.oauthApp)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), testUserEmail, result.AccessToken.UserAttributes["email"])
	assert.Equal(suite.T(), "Test User", result.AccessToken.UserAttributes["name"])
}

// ============================================================================
// HandleGrant Tests - Error Cases
// ============================================================================

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_InvalidSubjectToken_SignatureError() {
	now := time.Now().Unix()
	// Create a token that decodes successfully and has valid issuer, but invalid signature
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub": "user123",
		"iss": testCustomIssuer,
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
	})

	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         testClientID,
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
	}

	// Token will pass issuer validation but fail signature verification
	suite.mockTokenValidator.On("ValidateSubjectToken", mock.Anything, subjectToken, suite.oauthApp).
		Return(nil, errors.New("invalid subject token signature: invalid signature"))

	result, errResp := suite.handler.HandleGrant(context.TODO(), tokenRequest, suite.oauthApp)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), errResp)
	assert.Equal(suite.T(), constants.ErrorInvalidGrant, errResp.Error)
	assert.Contains(suite.T(), errResp.ErrorDescription, "Invalid subject_token")
	assert.Contains(suite.T(), errResp.ErrorDescription, "invalid signature")
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_InvalidSubjectToken_MissingSubClaim() {
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"iss": testCustomIssuer,
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
	})

	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         testClientID,
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
	}

	suite.mockTokenValidator.On("ValidateSubjectToken", mock.Anything, subjectToken, suite.oauthApp).
		Return(nil, errors.New("missing or invalid 'sub' claim"))

	result, errResp := suite.handler.HandleGrant(context.TODO(), tokenRequest, suite.oauthApp)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), errResp)
	assert.Equal(suite.T(), constants.ErrorInvalidGrant, errResp.Error)
	assert.Contains(suite.T(), errResp.ErrorDescription, "missing or invalid 'sub' claim")
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_InvalidSubjectToken_DecodeError() {
	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         testClientID,
		SubjectToken:     "invalid.jwt.format",
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
	}

	// Mock token validator to return decode error
	suite.mockTokenValidator.On("ValidateSubjectToken", mock.Anything, "invalid.jwt.format", suite.oauthApp).
		Return(nil, errors.New("invalid token format"))

	result, errResp := suite.handler.HandleGrant(context.TODO(), tokenRequest, suite.oauthApp)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), errResp)
	assert.Equal(suite.T(), constants.ErrorInvalidGrant, errResp.Error)
	assert.Contains(suite.T(), errResp.ErrorDescription, "Invalid subject_token")
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_InvalidSubjectToken_Expired() {
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub": "user123",
		"iss": testCustomIssuer,
		"exp": float64(now - 3600),
		"nbf": float64(now - 7200),
	})

	tokenRequest := suite.createBasicTokenRequest(subjectToken)
	suite.mockTokenValidator.On("ValidateSubjectToken", mock.Anything, subjectToken, suite.oauthApp).
		Return(nil, errors.New("token has expired"))

	result, errResp := suite.handler.HandleGrant(context.TODO(), tokenRequest, suite.oauthApp)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), errResp)
	assert.Equal(suite.T(), constants.ErrorInvalidGrant, errResp.Error)
	assert.Contains(suite.T(), errResp.ErrorDescription, "token has expired")
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_InvalidSubjectToken_NotYetValid() {
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub": "user123",
		"iss": testCustomIssuer,
		"exp": float64(now + 3600),
		"nbf": float64(now + 1800),
	})

	tokenRequest := suite.createBasicTokenRequest(subjectToken)
	suite.mockTokenValidator.On("ValidateSubjectToken", mock.Anything, subjectToken, suite.oauthApp).
		Return(nil, errors.New("token not yet valid"))

	result, errResp := suite.handler.HandleGrant(context.TODO(), tokenRequest, suite.oauthApp)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), errResp)
	assert.Equal(suite.T(), constants.ErrorInvalidGrant, errResp.Error)
	assert.Contains(suite.T(), errResp.ErrorDescription, "token not yet valid")
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_InvalidActorToken() {
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub": "user123",
		"iss": testCustomIssuer,
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
	})

	// Create a valid JWT format actor token that passes issuer validation but fails signature verification
	actorToken := suite.createTestJWT(map[string]interface{}{
		"sub": "service456",
		"iss": testCustomIssuer,
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
	})

	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         testClientID,
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
		ActorToken:       actorToken,
		ActorTokenType:   string(constants.TokenTypeIdentifierAccessToken),
	}

	suite.mockTokenValidator.On("ValidateSubjectToken", mock.Anything, subjectToken, suite.oauthApp).
		Return(&tokenservice.SubjectTokenClaims{
			Sub:            testUserID,
			Iss:            testCustomIssuer,
			UserAttributes: map[string]interface{}{},
			NestedAct:      nil,
		}, nil)
	suite.mockTokenValidator.On("ValidateSubjectToken", mock.Anything, actorToken, suite.oauthApp).
		Return(nil, errors.New("invalid subject token signature: invalid signature"))

	result, errResp := suite.handler.HandleGrant(context.TODO(), tokenRequest, suite.oauthApp)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), errResp)
	assert.Equal(suite.T(), constants.ErrorInvalidGrant, errResp.Error)
	assert.Contains(suite.T(), errResp.ErrorDescription, "Invalid actor_token")
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_InvalidScope() {
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub":   "user123",
		"iss":   testCustomIssuer,
		"exp":   float64(now + 3600),
		"nbf":   float64(now - 60),
		"scope": "read write",
	})

	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         testClientID,
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
		Scope:            "read write delete", // "delete" is not in subject token
	}

	suite.mockTokenValidator.On("ValidateSubjectToken", mock.Anything, subjectToken, suite.oauthApp).
		Return(&tokenservice.SubjectTokenClaims{
			Sub:            testUserID,
			Iss:            testCustomIssuer,
			Scopes:         []string{"read", "write"},
			UserAttributes: map[string]interface{}{},
			NestedAct:      nil,
		}, nil)
	// Expect token generation with only valid scopes ("read write", filtering out "delete")
	suite.mockTokenBuilder.On("BuildAccessToken", mock.Anything,
		mock.MatchedBy(func(ctx *tokenservice.AccessTokenBuildContext) bool {
			// Verify only valid scopes are included (filtering out "delete")
			return tokenservice.JoinScopes(ctx.Scopes) == "read write"
		})).Return(&model.TokenDTO{
		Token:     testTokenExchangeJWT,
		TokenType: constants.TokenTypeBearer,
		IssuedAt:  now,
		ExpiresIn: 7200,
		Scopes:    []string{"read", "write"},
		ClientID:  testClientID,
	}, nil)

	result, errResp := suite.handler.HandleGrant(context.TODO(), tokenRequest, suite.oauthApp)

	// Should succeed with only valid scopes filtered in
	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), []string{"read", "write"}, result.AccessToken.Scopes)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_ScopeEscalationPrevention() {
	now := time.Now().Unix()
	// Subject token has NO scopes
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub": "user123",
		"iss": testCustomIssuer,
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
	})

	// Request tries to add scopes
	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         testClientID,
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
		Scope:            "read write",
	}

	suite.mockTokenValidator.On("ValidateSubjectToken", mock.Anything, subjectToken, suite.oauthApp).
		Return(&tokenservice.SubjectTokenClaims{
			Sub:            testUserID,
			Iss:            testCustomIssuer,
			Scopes:         []string{}, // No scopes in subject token
			UserAttributes: map[string]interface{}{},
			NestedAct:      nil,
		}, nil)

	result, errResp := suite.handler.HandleGrant(context.TODO(), tokenRequest, suite.oauthApp)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), errResp)
	assert.Equal(suite.T(), constants.ErrorInvalidScope, errResp.Error)
	assert.Contains(suite.T(), errResp.ErrorDescription, "Cannot request scopes when the subject token has no scopes")
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_JWTGenerationError() {
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub": "user123",
		"iss": testCustomIssuer,
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
	})

	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         testClientID,
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
	}

	suite.mockTokenValidator.On("ValidateSubjectToken", mock.Anything, subjectToken, suite.oauthApp).
		Return(&tokenservice.SubjectTokenClaims{
			Sub:            testUserID,
			Iss:            testCustomIssuer,
			Scopes:         []string{},
			UserAttributes: map[string]interface{}{},
			NestedAct:      nil,
		}, nil)
	suite.mockTokenBuilder.On("BuildAccessToken", mock.Anything, mock.Anything).
		Return(nil, errors.New("failed to sign token"))

	result, errResp := suite.handler.HandleGrant(context.TODO(), tokenRequest, suite.oauthApp)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), errResp)
	assert.Equal(suite.T(), constants.ErrorServerError, errResp.Error)
	assert.Equal(suite.T(), "Failed to generate token", errResp.ErrorDescription)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_UsesDefaultConfig() {
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub": "user123",
		"iss": "https://test.thunder.io", // Use default config issuer since oauthApp has no Token config
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
	})

	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         testClientID,
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
	}

	// Use app without custom token config
	oauthAppNoConfig := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:   testClientID,
		GrantTypes: []constants.GrantType{constants.GrantTypeTokenExchange},
	}

	suite.mockTokenValidator.On("ValidateSubjectToken", mock.Anything, subjectToken, oauthAppNoConfig).
		Return(&tokenservice.SubjectTokenClaims{
			Sub:            testUserID,
			Iss:            testCustomIssuer,
			Scopes:         []string{},
			UserAttributes: map[string]interface{}{},
			NestedAct:      nil,
		}, nil)
	suite.mockTokenBuilder.On("BuildAccessToken", mock.Anything, mock.Anything).
		Return(&model.TokenDTO{
			Token:     testTokenExchangeJWT,
			TokenType: constants.TokenTypeBearer,
			IssuedAt:  now,
			ExpiresIn: 3600,
			Scopes:    []string{},
			ClientID:  testClientID,
		}, nil)

	result, errResp := suite.handler.HandleGrant(context.TODO(), tokenRequest, oauthAppNoConfig)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_Success_WithJWTTokenType() {
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub": "user123",
		"iss": testCustomIssuer,
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
	})

	tokenRequest := &model.TokenRequest{
		GrantType:          string(constants.GrantTypeTokenExchange),
		ClientID:           testClientID,
		SubjectToken:       subjectToken,
		SubjectTokenType:   string(constants.TokenTypeIdentifierAccessToken),
		RequestedTokenType: string(constants.TokenTypeIdentifierJWT),
	}

	suite.mockTokenValidator.On("ValidateSubjectToken", mock.Anything, subjectToken, suite.oauthApp).
		Return(&tokenservice.SubjectTokenClaims{
			Sub:            testUserID,
			Iss:            testCustomIssuer,
			Scopes:         []string{},
			UserAttributes: map[string]interface{}{},
			NestedAct:      nil,
		}, nil)
	suite.mockTokenBuilder.On("BuildAccessToken", mock.Anything, mock.Anything).
		Return(&model.TokenDTO{
			Token:     testTokenExchangeJWT,
			TokenType: constants.TokenTypeBearer,
			IssuedAt:  now,
			ExpiresIn: 7200,
			Scopes:    []string{},
			ClientID:  testClientID,
		}, nil)

	result, errResp := suite.handler.HandleGrant(context.TODO(), tokenRequest, suite.oauthApp)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)
	assert.NotEmpty(suite.T(), result.AccessToken.Token)
	// IssuedTokenType is determined at the token handler level, not the grant handler level
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestValidateGrant_UnsupportedIDTokenType() {
	tokenRequest := &model.TokenRequest{
		GrantType:          string(constants.GrantTypeTokenExchange),
		ClientID:           testClientID,
		SubjectToken:       "subject-token",
		SubjectTokenType:   string(constants.TokenTypeIdentifierAccessToken),
		RequestedTokenType: string(constants.TokenTypeIdentifierIDToken),
	}

	errResp := suite.handler.ValidateGrant(context.TODO(), tokenRequest, suite.oauthApp)
	assert.NotNil(suite.T(), errResp)
	assert.Equal(suite.T(), constants.ErrorInvalidTarget, errResp.Error)
	assert.Contains(suite.T(), errResp.ErrorDescription, "not supported")
	assert.Contains(suite.T(), errResp.ErrorDescription, string(constants.TokenTypeIdentifierIDToken))
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestValidateGrant_UnsupportedRefreshTokenType() {
	tokenRequest := &model.TokenRequest{
		GrantType:          string(constants.GrantTypeTokenExchange),
		ClientID:           testClientID,
		SubjectToken:       "subject-token",
		SubjectTokenType:   string(constants.TokenTypeIdentifierAccessToken),
		RequestedTokenType: string(constants.TokenTypeIdentifierRefreshToken),
	}

	// Test ValidateGrant first (which is called before HandleGrant in production)
	errResp := suite.handler.ValidateGrant(context.TODO(), tokenRequest, suite.oauthApp)
	assert.NotNil(suite.T(), errResp)
	assert.Equal(suite.T(), constants.ErrorInvalidTarget, errResp.Error)
	assert.Contains(suite.T(), errResp.ErrorDescription, "not supported")
	assert.Contains(suite.T(), errResp.ErrorDescription, string(constants.TokenTypeIdentifierRefreshToken))
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestRFC8693_CompleteTokenExchangeFlow() {
	// RFC 8693 Section 2.2: Verify all required response parameters
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub":   "user123",
		"iss":   testCustomIssuer,
		"aud":   "original-audience",
		"exp":   float64(now + 3600),
		"nbf":   float64(now - 60),
		"scope": "read write",
		"email": "user@example.com",
		"name":  "John Doe",
	})

	tokenRequest := &model.TokenRequest{
		GrantType:          string(constants.GrantTypeTokenExchange),
		ClientID:           testClientID,
		SubjectToken:       subjectToken,
		SubjectTokenType:   string(constants.TokenTypeIdentifierAccessToken),
		RequestedTokenType: string(constants.TokenTypeIdentifierAccessToken),
		Audience:           "https://target-service.com",
		Scope:              "read",
	}

	suite.mockTokenValidator.On("ValidateSubjectToken", mock.Anything, subjectToken, suite.oauthApp).
		Return(&tokenservice.SubjectTokenClaims{
			Sub:    testUserID,
			Iss:    testCustomIssuer,
			Scopes: []string{"read", "write"},
			UserAttributes: map[string]interface{}{
				"email": testUserEmail,
				"name":  "John Doe",
			},
			NestedAct: nil,
		}, nil)
	suite.mockTokenBuilder.On("BuildAccessToken", mock.Anything, mock.MatchedBy(func(ctx *tokenservice.AccessTokenBuildContext) bool {
		// Verify claims structure per RFC 8693
		return ctx.Subject == testUserID &&
			ctx.Audience == "https://target-service.com" &&
			ctx.ClientID == testClientID &&
			tokenservice.JoinScopes(ctx.Scopes) == testScopeRead &&
			ctx.UserAttributes["email"] == "user@example.com" &&
			ctx.UserAttributes["name"] == "John Doe"
	})).Return(&model.TokenDTO{
		Token:     testTokenExchangeJWT,
		TokenType: constants.TokenTypeBearer,
		IssuedAt:  now,
		ExpiresIn: 7200,
		Scopes:    []string{"read"},
		ClientID:  testClientID,
		UserAttributes: map[string]interface{}{
			"email": "user@example.com",
			"name":  "John Doe",
		},
	}, nil)

	result, errResp := suite.handler.HandleGrant(context.TODO(), tokenRequest, suite.oauthApp)

	// RFC 8693 Section 2.2: Verify required response parameters
	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)
	assert.NotEmpty(suite.T(), result.AccessToken.Token)                             // access_token - REQUIRED
	assert.Equal(suite.T(), constants.TokenTypeBearer, result.AccessToken.TokenType) // token_type - REQUIRED
	assert.NotZero(suite.T(), result.AccessToken.ExpiresIn)                          // expires_in - RECOMMENDED
	assert.Equal(suite.T(), []string{"read"}, result.AccessToken.Scopes)
	// issued_token_type - REQUIRED
	// IssuedTokenType is determined at the token handler level, not the grant handler level
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestRFC8693_AudiencePriority() {
	// RFC 8693: Test audience parameter priority (audience > resource > token.aud > client_id)
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub": "user123",
		"iss": testCustomIssuer,
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
		"aud": "token-audience",
	})

	// Test 1: Audience parameter takes priority
	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         testClientID,
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
		Audience:         "request-audience",
		Resource:         "https://resource.example.com",
	}

	suite.mockTokenValidator.On("ValidateSubjectToken", mock.Anything, subjectToken, suite.oauthApp).
		Return(&tokenservice.SubjectTokenClaims{
			Sub:            testUserID,
			Iss:            testCustomIssuer,
			Scopes:         []string{},
			UserAttributes: map[string]interface{}{},
			NestedAct:      nil,
		}, nil)
	suite.mockTokenBuilder.On("BuildAccessToken", mock.Anything, mock.MatchedBy(func(ctx *tokenservice.AccessTokenBuildContext) bool {
		// Should use request audience, not resource or token aud
		return ctx.Subject == testUserID && ctx.Audience == "request-audience"
	})).Return(&model.TokenDTO{
		Token:     testTokenExchangeJWT,
		TokenType: constants.TokenTypeBearer,
		IssuedAt:  now,
		ExpiresIn: 7200,
		Scopes:    []string{},
		ClientID:  testClientID,
	}, nil)

	result, errResp := suite.handler.HandleGrant(context.TODO(), tokenRequest, suite.oauthApp)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestRFC8693_ActorDelegationChain() {
	// RFC 8693 Section 4.1: Test nested actor delegation chains
	now := time.Now().Unix()

	// Subject token with existing actor
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub": "user123",
		"iss": testCustomIssuer,
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
		"act": map[string]interface{}{
			"sub": "previous-actor",
			"iss": "https://previous-issuer.com",
		},
	})

	// Actor token with its own actor chain
	actorToken := suite.createTestJWT(map[string]interface{}{
		"sub": "current-actor",
		"iss": testCustomIssuer,
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
		"act": map[string]interface{}{
			"sub": "actor-of-actor",
			"iss": "https://nested-issuer.com",
		},
	})

	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         testClientID,
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
		ActorToken:       actorToken,
		ActorTokenType:   string(constants.TokenTypeIdentifierAccessToken),
	}

	suite.mockTokenValidator.On("ValidateSubjectToken", mock.Anything, subjectToken, suite.oauthApp).
		Return(&tokenservice.SubjectTokenClaims{
			Sub:            "user123",
			Iss:            testCustomIssuer,
			UserAttributes: map[string]interface{}{},
			NestedAct: map[string]interface{}{
				"sub": "previous-actor",
				"iss": "https://previous-issuer.com",
			},
		}, nil)
	suite.mockTokenValidator.On("ValidateSubjectToken", mock.Anything, actorToken, suite.oauthApp).
		Return(&tokenservice.SubjectTokenClaims{
			Sub:            "current-actor",
			Iss:            testCustomIssuer,
			UserAttributes: map[string]interface{}{},
			NestedAct: map[string]interface{}{
				"sub": "actor-of-actor",
				"iss": "https://nested-issuer.com",
			},
		}, nil)
	suite.mockTokenBuilder.On("BuildAccessToken", mock.Anything,
		mock.MatchedBy(func(ctx *tokenservice.AccessTokenBuildContext) bool {
			// Verify nested delegation chain per RFC 8693
			if ctx.ActorClaims == nil {
				return false
			}
			// Current actor
			if ctx.ActorClaims.Sub != "current-actor" || ctx.ActorClaims.Iss != testCustomIssuer {
				return false
			}
			// Check that actor token's act claim is preserved
			if len(ctx.ActorClaims.NestedAct) > 0 {
				return ctx.ActorClaims.NestedAct["sub"] == "actor-of-actor"
			}
			return false
		})).Return(&model.TokenDTO{
		Token:     testTokenExchangeJWT,
		TokenType: constants.TokenTypeBearer,
		IssuedAt:  now,
		ExpiresIn: 7200,
		Scopes:    []string{},
		ClientID:  testClientID,
	}, nil)

	result, errResp := suite.handler.HandleGrant(context.TODO(), tokenRequest, suite.oauthApp)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_Success_WithActorTokenHasActButSubjectHasNoAct() {
	// Test case: Actor token has its own act claim, but subject token has no act claim
	now := time.Now().Unix()

	// Subject token WITHOUT act claim
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub": "user123",
		"iss": testCustomIssuer,
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
	})

	// Actor token WITH its own act claim
	actorToken := suite.createTestJWT(map[string]interface{}{
		"sub": "current-actor",
		"iss": testCustomIssuer,
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
		"act": map[string]interface{}{
			"sub": "actor-of-actor",
			"iss": "https://nested-issuer.com",
		},
	})

	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         testClientID,
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
		ActorToken:       actorToken,
		ActorTokenType:   string(constants.TokenTypeIdentifierAccessToken),
	}

	suite.mockTokenValidator.On("ValidateSubjectToken", mock.Anything, subjectToken, suite.oauthApp).
		Return(&tokenservice.SubjectTokenClaims{
			Sub:            testUserID,
			Iss:            testCustomIssuer,
			UserAttributes: map[string]interface{}{},
			NestedAct:      nil,
		}, nil)
	suite.mockTokenValidator.On("ValidateSubjectToken", mock.Anything, actorToken, suite.oauthApp).
		Return(&tokenservice.SubjectTokenClaims{
			Sub:            "current-actor",
			Iss:            testCustomIssuer,
			UserAttributes: map[string]interface{}{},
			NestedAct: map[string]interface{}{
				"sub": "actor-of-actor",
				"iss": "https://nested-issuer.com",
			},
		}, nil)
	suite.mockTokenBuilder.On("BuildAccessToken", mock.Anything, mock.MatchedBy(func(ctx *tokenservice.AccessTokenBuildContext) bool {
		// Verify actor claim structure
		if ctx.ActorClaims == nil {
			return false
		}
		// Current actor should be present
		if ctx.ActorClaims.Sub != "current-actor" || ctx.ActorClaims.Iss != testCustomIssuer {
			return false
		}
		// Actor's act claim should be preserved directly
		if len(ctx.ActorClaims.NestedAct) > 0 {
			nestedAct := ctx.ActorClaims.NestedAct
			nestedSub := nestedAct["sub"] == "actor-of-actor"
			nestedIss := nestedAct["iss"] == "https://nested-issuer.com"
			if !nestedSub || !nestedIss {
				return false
			}
			// Subject has no act claim, so it should not be nested
			_, hasFurtherNesting := ctx.ActorClaims.NestedAct["act"]
			return !hasFurtherNesting
		}
		return false
	})).Return(&model.TokenDTO{
		Token:     testTokenExchangeJWT,
		TokenType: constants.TokenTypeBearer,
		IssuedAt:  now,
		ExpiresIn: 7200,
		Scopes:    []string{},
		ClientID:  testClientID,
	}, nil)

	result, errResp := suite.handler.HandleGrant(context.TODO(), tokenRequest, suite.oauthApp)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestRFC8693_ScopeDownscopingEnforcement() {
	// RFC 8693 Section 5: Verify scope downscoping (security consideration)
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub":   "user123",
		"iss":   testCustomIssuer,
		"exp":   float64(now + 3600),
		"nbf":   float64(now - 60),
		"scope": "read write delete",
	})

	// Test 1: Valid downscoping (subset of scopes)
	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         testClientID,
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
		Scope:            "read",
	}

	suite.mockTokenValidator.On("ValidateSubjectToken", mock.Anything, subjectToken, suite.oauthApp).
		Return(&tokenservice.SubjectTokenClaims{
			Sub:            testUserID,
			Iss:            testCustomIssuer,
			Scopes:         []string{"read", "write", "delete"},
			UserAttributes: map[string]interface{}{},
			NestedAct:      nil,
		}, nil)
	suite.mockTokenBuilder.On("BuildAccessToken", mock.Anything,
		mock.MatchedBy(func(ctx *tokenservice.AccessTokenBuildContext) bool {
			return tokenservice.JoinScopes(ctx.Scopes) == testScopeRead
		})).Return(&model.TokenDTO{
		Token:     testTokenExchangeJWT,
		TokenType: constants.TokenTypeBearer,
		IssuedAt:  now,
		ExpiresIn: 7200,
		Scopes:    []string{"read"},
		ClientID:  testClientID,
	}, nil)

	result, errResp := suite.handler.HandleGrant(context.TODO(), tokenRequest, suite.oauthApp)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), []string{"read"}, result.AccessToken.Scopes)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestRFC8693_ResourceParameterValidation() {
	// RFC 8693 Section 2.1: Resource must be absolute URI without fragment
	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         testClientID,
		SubjectToken:     "subject-token",
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
		Resource:         "https://api.example.com/v1/resource",
	}

	result := suite.handler.ValidateGrant(context.TODO(), tokenRequest, suite.oauthApp)
	assert.Nil(suite.T(), result)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestRFC8693_NoTokenLinkage() {
	// RFC 8693 Section 2.1: "exchange has no impact on the validity of the subject token"
	// This is a design verification test - token exchange should not invalidate input tokens
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub": "user123",
		"iss": testCustomIssuer,
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
	})

	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         testClientID,
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
	}

	suite.mockTokenValidator.On("ValidateSubjectToken", mock.Anything, subjectToken, suite.oauthApp).
		Return(&tokenservice.SubjectTokenClaims{
			Sub:            testUserID,
			Iss:            testCustomIssuer,
			Scopes:         []string{},
			UserAttributes: map[string]interface{}{},
			NestedAct:      nil,
		}, nil)
	suite.mockTokenBuilder.On("BuildAccessToken", mock.Anything, mock.Anything).
		Return(&model.TokenDTO{
			Token:     testTokenExchangeJWT,
			TokenType: constants.TokenTypeBearer,
			IssuedAt:  now,
			ExpiresIn: 7200,
			Scopes:    []string{},
			ClientID:  testClientID,
		}, nil)

	result1, errResp1 := suite.handler.HandleGrant(context.TODO(), tokenRequest, suite.oauthApp)

	assert.Nil(suite.T(), errResp1)
	assert.NotNil(suite.T(), result1)

	// Use same subject token again - should succeed (no linkage/invalidation)
	result2, errResp2 := suite.handler.HandleGrant(context.TODO(), tokenRequest, suite.oauthApp)

	assert.Nil(suite.T(), errResp2)
	assert.NotNil(suite.T(), result2)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestRFC8693_ClaimPreservation() {
	// Verify non-standard claims are preserved through token exchange
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub":          "user123",
		"iss":          testCustomIssuer,
		"exp":          float64(now + 3600),
		"nbf":          float64(now - 60),
		"email":        "user@example.com",
		"given_name":   "John",
		"family_name":  "Doe",
		"roles":        []interface{}{"admin", "user"},
		"organization": "ACME Corp",
	})

	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         testClientID,
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
	}

	suite.mockTokenValidator.On("ValidateSubjectToken", mock.Anything, subjectToken, suite.oauthApp).
		Return(&tokenservice.SubjectTokenClaims{
			Sub:    testUserID,
			Iss:    testCustomIssuer,
			Scopes: []string{},
			UserAttributes: map[string]interface{}{
				"email":        "user@example.com",
				"given_name":   "John",
				"family_name":  "Doe",
				"roles":        []interface{}{"admin", "user"},
				"organization": "ACME Corp",
			},
			NestedAct: nil,
		}, nil)
	suite.mockTokenBuilder.On("BuildAccessToken", mock.Anything,
		mock.MatchedBy(func(ctx *tokenservice.AccessTokenBuildContext) bool {
			// Verify all custom claims are preserved in user attributes
			return ctx.UserAttributes["email"] == testUserEmail &&
				ctx.UserAttributes["given_name"] == "John" &&
				ctx.UserAttributes["family_name"] == "Doe" &&
				ctx.UserAttributes["organization"] == "ACME Corp"
		})).Return(&model.TokenDTO{
		Token:     testTokenExchangeJWT,
		TokenType: constants.TokenTypeBearer,
		IssuedAt:  now,
		ExpiresIn: 7200,
		Scopes:    []string{},
		ClientID:  testClientID,
		UserAttributes: map[string]interface{}{
			"email":        "user@example.com",
			"given_name":   "John",
			"family_name":  "Doe",
			"roles":        []interface{}{"admin", "user"},
			"organization": "ACME Corp",
		},
	}, nil)

	result, errResp := suite.handler.HandleGrant(context.TODO(), tokenRequest, suite.oauthApp)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)

	// Verify user attributes in response
	assert.Equal(suite.T(), testUserEmail, result.AccessToken.UserAttributes["email"])
	assert.Equal(suite.T(), "John", result.AccessToken.UserAttributes["given_name"])
	assert.Equal(suite.T(), "Doe", result.AccessToken.UserAttributes["family_name"])
	assert.Equal(suite.T(), "ACME Corp", result.AccessToken.UserAttributes["organization"])
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestIsSupportedTokenType() {
	assert.True(suite.T(), constants.TokenTypeIdentifierAccessToken.IsValid())
	assert.True(suite.T(), constants.TokenTypeIdentifierRefreshToken.IsValid())
	assert.True(suite.T(), constants.TokenTypeIdentifierIDToken.IsValid())
	assert.True(suite.T(), constants.TokenTypeIdentifierJWT.IsValid())
	assert.False(suite.T(), constants.TokenTypeIdentifier("urn:ietf:params:oauth:token-type:saml2").IsValid())
	assert.False(suite.T(), constants.TokenTypeIdentifier("invalid").IsValid())
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestGetAudience_WithClientIDFallback() {
	// Test that DetermineAudience falls back to clientID when no audience, resource, or token.aud provided
	result := tokenservice.DetermineAudience("", "", "", testClientID)
	assert.Equal(suite.T(), testClientID, result)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestExtractUserAttributes() {
	claims := map[string]interface{}{
		"sub":       testUserID,
		"iss":       "issuer",
		"aud":       "audience",
		"exp":       float64(123456789),
		"nbf":       float64(123456789),
		"iat":       float64(123456789),
		"jti":       "jwt-id",
		"scope":     "read write",
		"client_id": testClientID,
		"act":       map[string]interface{}{"sub": "actor"},
		"email":     testUserEmail,
		"name":      "Test User",
		"custom":    "value",
	}

	// Use the utility function from tokenservice
	userAttrs := tokenservice.ExtractUserAttributes(claims)

	assert.Equal(suite.T(), 3, len(userAttrs))
	assert.Equal(suite.T(), testUserEmail, userAttrs["email"])
	assert.Equal(suite.T(), "Test User", userAttrs["name"])
	assert.Equal(suite.T(), "value", userAttrs["custom"])
	assert.NotContains(suite.T(), userAttrs, "sub")
	assert.NotContains(suite.T(), userAttrs, "iss")
	assert.NotContains(suite.T(), userAttrs, "scope")
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestDetermineAudience_Priority() {
	// Audience parameter has highest priority (RFC 8693)
	aud := tokenservice.DetermineAudience("request-audience", "request-resource", "token-aud", testClientID)
	assert.Equal(suite.T(), "request-audience", aud)

	// Resource parameter is second priority
	aud = tokenservice.DetermineAudience("", "request-resource", "token-aud", testClientID)
	assert.Equal(suite.T(), "request-resource", aud)

	// Token.aud is third priority
	aud = tokenservice.DetermineAudience("", "", "token-aud", testClientID)
	assert.Equal(suite.T(), "token-aud", aud)

	// Client ID is fallback when neither audience, resource, nor token.aud provided
	aud = tokenservice.DetermineAudience("", "", "", testClientID)
	assert.Equal(suite.T(), testClientID, aud)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_ThunderAuthAssertion_Success() {
	defaultAudience := suite.getDefaultAudience()
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub":                    testUserID,
		"iss":                    testCustomIssuer,
		"aud":                    defaultAudience, // Match default audience
		"exp":                    float64(now + 3600),
		"nbf":                    float64(now - 60),
		"assurance":              map[string]interface{}{"aal": "AAL1", "ial": "IAL1"}, // Make it an auth assertion
		"authorized_permissions": "read:documents write:documents",
		"userType":               "person",
	}
	subjectToken := suite.createTestJWT(claims)

	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         testClientID,
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierJWT),
	}

	suite.mockTokenValidator.On("ValidateSubjectToken", mock.Anything, subjectToken, suite.oauthApp).
		Return(&tokenservice.SubjectTokenClaims{
			Sub:            testUserID,
			Iss:            testCustomIssuer,
			Aud:            defaultAudience,
			Scopes:         []string{"read:documents", "write:documents"}, // Mapped from authorized_permissions
			UserAttributes: map[string]interface{}{"userType": "person"},
			NestedAct:      nil,
		}, nil)

	suite.mockTokenBuilder.On("BuildAccessToken", mock.Anything,
		mock.MatchedBy(func(ctx *tokenservice.AccessTokenBuildContext) bool {
			return ctx.Subject == testUserID &&
				len(ctx.Scopes) == 2 &&
				ctx.Scopes[0] == "read:documents" &&
				ctx.Scopes[1] == "write:documents"
		})).Return(&model.TokenDTO{
		Token:     testTokenExchangeJWT,
		TokenType: constants.TokenTypeBearer,
		IssuedAt:  now,
		ExpiresIn: 7200,
		Scopes:    []string{"read:documents", "write:documents"},
		ClientID:  testClientID,
	}, nil)

	result, errResp := suite.handler.HandleGrant(context.TODO(), tokenRequest, suite.oauthApp)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), testTokenExchangeJWT, result.AccessToken.Token)
	assert.Equal(suite.T(), []string{"read:documents", "write:documents"}, result.AccessToken.Scopes)
	suite.mockTokenValidator.AssertExpectations(suite.T())
	suite.mockTokenBuilder.AssertExpectations(suite.T())
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_ThunderAuthAssertion_AudienceMismatch() {
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub":       testUserID,
		"iss":       testCustomIssuer,
		"aud":       "different-audience", // Doesn't match default audience or client app_id
		"exp":       float64(now + 3600),
		"nbf":       float64(now - 60),
		"assurance": map[string]interface{}{"aal": "AAL1", "ial": "IAL1"}, // Make it an auth assertion
	}
	subjectToken := suite.createTestJWT(claims)

	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         testClientID,
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierJWT),
	}

	suite.mockTokenValidator.On("ValidateSubjectToken", mock.Anything, subjectToken, suite.oauthApp).
		Return(nil, fmt.Errorf("auth assertion audience mismatch"))

	result, errResp := suite.handler.HandleGrant(context.TODO(), tokenRequest, suite.oauthApp)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), errResp)
	assert.Equal(suite.T(), constants.ErrorInvalidGrant, errResp.Error)
	assert.Contains(suite.T(), errResp.ErrorDescription, "auth assertion audience mismatch")
	suite.mockTokenValidator.AssertExpectations(suite.T())
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_ThunderAuthAssertion_MissingAudience() {
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub": testUserID,
		"iss": testCustomIssuer,
		// Missing aud claim
		"exp":       float64(now + 3600),
		"nbf":       float64(now - 60),
		"assurance": map[string]interface{}{"aal": "AAL1", "ial": "IAL1"}, // Make it an auth assertion
	}
	subjectToken := suite.createTestJWT(claims)

	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         testClientID,
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierJWT),
	}

	suite.mockTokenValidator.On("ValidateSubjectToken", mock.Anything, subjectToken, suite.oauthApp).
		Return(nil, fmt.Errorf("thunder auth assertion is missing 'aud' claim"))

	result, errResp := suite.handler.HandleGrant(context.TODO(), tokenRequest, suite.oauthApp)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), errResp)
	assert.Equal(suite.T(), constants.ErrorInvalidGrant, errResp.Error)
	assert.Contains(suite.T(), errResp.ErrorDescription, "missing 'aud' claim")
	suite.mockTokenValidator.AssertExpectations(suite.T())
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_ThunderAuthAssertion_ClientIDMatch() {
	defaultAudience := suite.getDefaultAudience()
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub":                    testUserID,
		"iss":                    testCustomIssuer,
		"aud":                    defaultAudience, // Match default audience
		"exp":                    float64(now + 3600),
		"nbf":                    float64(now - 60),
		"assurance":              map[string]interface{}{"aal": "AAL1", "ial": "IAL1"}, // Make it an auth assertion
		"authorized_permissions": "read write",
	}
	subjectToken := suite.createTestJWT(claims)

	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         testClientID,
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierJWT),
	}

	suite.mockTokenValidator.On("ValidateSubjectToken", mock.Anything, subjectToken, suite.oauthApp).
		Return(&tokenservice.SubjectTokenClaims{
			Sub:            testUserID,
			Iss:            testCustomIssuer,
			Aud:            defaultAudience,
			Scopes:         []string{"read", "write"},
			UserAttributes: map[string]interface{}{},
			NestedAct:      nil,
		}, nil)

	suite.mockTokenBuilder.On("BuildAccessToken", mock.Anything, mock.Anything).Return(&model.TokenDTO{
		Token:     testTokenExchangeJWT,
		TokenType: constants.TokenTypeBearer,
		IssuedAt:  now,
		ExpiresIn: 7200,
		Scopes:    []string{"read", "write"},
		ClientID:  testClientID,
	}, nil)

	result, errResp := suite.handler.HandleGrant(context.TODO(), tokenRequest, suite.oauthApp)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)
	suite.mockTokenValidator.AssertExpectations(suite.T())
	suite.mockTokenBuilder.AssertExpectations(suite.T())
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_ThunderAuthAssertion_WithClientAppID() {
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"sub":                    testUserID,
		"iss":                    testCustomIssuer,
		"aud":                    suite.oauthApp.AppID, // Match client app_id
		"exp":                    float64(now + 3600),
		"nbf":                    float64(now - 60),
		"assurance":              map[string]interface{}{"aal": "AAL1", "ial": "IAL1"}, // Make it an auth assertion
		"authorized_permissions": "read:documents write:documents",
		"userType":               "person",
	}
	subjectToken := suite.createTestJWT(claims)

	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         testClientID,
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierJWT),
	}

	suite.mockTokenValidator.On("ValidateSubjectToken", mock.Anything, subjectToken, suite.oauthApp).
		Return(&tokenservice.SubjectTokenClaims{
			Sub:            testUserID,
			Iss:            testCustomIssuer,
			Aud:            suite.oauthApp.AppID,
			Scopes:         []string{"read:documents", "write:documents"},
			UserAttributes: map[string]interface{}{"userType": "person"},
			NestedAct:      nil,
		}, nil)

	suite.mockTokenBuilder.On("BuildAccessToken", mock.Anything,
		mock.MatchedBy(func(ctx *tokenservice.AccessTokenBuildContext) bool {
			return ctx.Subject == testUserID &&
				len(ctx.Scopes) == 2 &&
				ctx.Scopes[0] == "read:documents" &&
				ctx.Scopes[1] == "write:documents"
		})).Return(&model.TokenDTO{
		Token:     testTokenExchangeJWT,
		TokenType: constants.TokenTypeBearer,
		IssuedAt:  now,
		ExpiresIn: 7200,
		Scopes:    []string{"read:documents", "write:documents"},
		ClientID:  testClientID,
	}, nil)

	result, errResp := suite.handler.HandleGrant(context.TODO(), tokenRequest, suite.oauthApp)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), testTokenExchangeJWT, result.AccessToken.Token)
	assert.Equal(suite.T(), []string{"read:documents", "write:documents"}, result.AccessToken.Scopes)
	suite.mockTokenValidator.AssertExpectations(suite.T())
	suite.mockTokenBuilder.AssertExpectations(suite.T())
}
