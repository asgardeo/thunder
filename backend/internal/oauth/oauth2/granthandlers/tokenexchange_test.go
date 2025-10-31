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
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/tests/mocks/jwtmock"
)

const (
	testTokenExchangeJWT = "test-token-exchange-jwt" //nolint:gosec
	testScopeReadWrite   = "read write"
)

type TokenExchangeGrantHandlerTestSuite struct {
	suite.Suite
	mockJWTService *jwtmock.JWTServiceInterfaceMock
	handler        *tokenExchangeGrantHandler
	oauthApp       *appmodel.OAuthAppConfigProcessedDTO
}

func TestTokenExchangeGrantHandlerSuite(t *testing.T) {
	suite.Run(t, new(TokenExchangeGrantHandlerTestSuite))
}

func (suite *TokenExchangeGrantHandlerTestSuite) SetupTest() {
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "https://test.thunder.io",
			ValidityPeriod: 3600,
		},
	}
	err := config.InitializeThunderRuntime("", testConfig)
	assert.NoError(suite.T(), err)

	suite.mockJWTService = jwtmock.NewJWTServiceInterfaceMock(suite.T())
	suite.handler = &tokenExchangeGrantHandler{
		jwtService: suite.mockJWTService,
	}

	suite.oauthApp = &appmodel.OAuthAppConfigProcessedDTO{
		AppID:                   "app123",
		ClientID:                "client123",
		HashedClientSecret:      "hashedsecret123",
		RedirectURIs:            []string{"https://example.com/callback"},
		GrantTypes:              []constants.GrantType{constants.GrantTypeTokenExchange},
		ResponseTypes:           []constants.ResponseType{constants.ResponseTypeCode},
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretBasic,
		Token: &appmodel.OAuthTokenConfig{
			AccessToken: &appmodel.TokenConfig{
				Issuer:         "https://custom.issuer.com",
				ValidityPeriod: 7200,
			},
		},
	}
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
		ClientID:         "client123",
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
	}
}

// Helper function to setup JWT mock and execute HandleGrant with common assertions for invalid token errors
func (suite *TokenExchangeGrantHandlerTestSuite) executeHandleGrantWithInvalidToken(
	tokenRequest *model.TokenRequest,
	expectedError string,
	expectedErrorDesc string,
) {
	suite.mockJWTService.On("VerifyJWTSignature", tokenRequest.SubjectToken).Return(nil)

	ctx := &model.TokenContext{}
	result, errResp := suite.handler.HandleGrant(tokenRequest, suite.oauthApp, ctx)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), errResp)
	assert.Equal(suite.T(), expectedError, errResp.Error)
	assert.Contains(suite.T(), errResp.ErrorDescription, expectedErrorDesc)
}

// Helper function to setup JWT mock for successful token generation with audience check
func (suite *TokenExchangeGrantHandlerTestSuite) setupSuccessfulJWTMock(
	subjectToken string,
	expectedAudience string,
	now int64,
) {
	suite.mockJWTService.On("VerifyJWTSignature", subjectToken).Return(nil)
	suite.mockJWTService.On("GenerateJWT",
		"user123",
		expectedAudience,
		"https://custom.issuer.com",
		int64(7200),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			// aud is set from the GenerateJWT parameter, not in claims map
			return true
		}),
	).Return(testTokenExchangeJWT, now, nil)
}

// Helper function to setup JWT mock for successful token generation with scope check
func (suite *TokenExchangeGrantHandlerTestSuite) setupSuccessfulJWTMockWithScope(
	subjectToken string,
	expectedAudience string,
	expectedScope string,
	now int64,
) {
	suite.mockJWTService.On("VerifyJWTSignature", subjectToken).Return(nil)
	suite.mockJWTService.On("GenerateJWT",
		"user123",
		expectedAudience,
		"https://custom.issuer.com",
		int64(7200),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			return claims["scope"] == expectedScope
		}),
	).Return(testTokenExchangeJWT, now, nil)
}

// TestNewTokenExchangeGrantHandler tests the constructor
func (suite *TokenExchangeGrantHandlerTestSuite) TestNewTokenExchangeGrantHandler() {
	handler := newTokenExchangeGrantHandler(suite.mockJWTService)
	assert.NotNil(suite.T(), handler)
	assert.Implements(suite.T(), (*GrantHandlerInterface)(nil), handler)
}

// ============================================================================
// ValidateGrant Tests
// ============================================================================

func (suite *TokenExchangeGrantHandlerTestSuite) TestValidateGrant_Success() {
	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         "client123",
		ClientSecret:     "secret123",
		SubjectToken:     "subject-token",
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
	}

	result := suite.handler.ValidateGrant(tokenRequest, suite.oauthApp)
	assert.Nil(suite.T(), result)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestValidateGrant_WrongGrantType() {
	tokenRequest := &model.TokenRequest{
		GrantType:        "authorization_code",
		ClientID:         "client123",
		SubjectToken:     "subject-token",
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
	}

	result := suite.handler.ValidateGrant(tokenRequest, suite.oauthApp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), constants.ErrorUnsupportedGrantType, result.Error)
	assert.Equal(suite.T(), "Unsupported grant type", result.ErrorDescription)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestValidateGrant_MissingSubjectToken() {
	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         "client123",
		SubjectToken:     "",
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
	}

	result := suite.handler.ValidateGrant(tokenRequest, suite.oauthApp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, result.Error)
	assert.Equal(suite.T(), "Missing required parameter: subject_token", result.ErrorDescription)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestValidateGrant_MissingSubjectTokenType() {
	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         "client123",
		SubjectToken:     "subject-token",
		SubjectTokenType: "",
	}

	result := suite.handler.ValidateGrant(tokenRequest, suite.oauthApp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, result.Error)
	assert.Equal(suite.T(), "Missing required parameter: subject_token_type", result.ErrorDescription)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestValidateGrant_UnsupportedSubjectTokenType() {
	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         "client123",
		SubjectToken:     "subject-token",
		SubjectTokenType: "urn:ietf:params:oauth:token-type:saml2",
	}

	result := suite.handler.ValidateGrant(tokenRequest, suite.oauthApp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, result.Error)
	assert.Contains(suite.T(), result.ErrorDescription, "Unsupported subject_token_type")
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestValidateGrant_MissingActorTokenType() {
	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         "client123",
		SubjectToken:     "subject-token",
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
		ActorToken:       "actor-token",
		ActorTokenType:   "",
	}

	result := suite.handler.ValidateGrant(tokenRequest, suite.oauthApp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, result.Error)
	assert.Equal(suite.T(), "actor_token_type is required when actor_token is provided", result.ErrorDescription)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestValidateGrant_UnsupportedActorTokenType() {
	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         "client123",
		SubjectToken:     "subject-token",
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
		ActorToken:       "actor-token",
		ActorTokenType:   "urn:ietf:params:oauth:token-type:saml1",
	}

	result := suite.handler.ValidateGrant(tokenRequest, suite.oauthApp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, result.Error)
	assert.Contains(suite.T(), result.ErrorDescription, "Unsupported actor_token_type")
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestValidateGrant_ActorTokenTypeWithoutActorToken() {
	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         "client123",
		SubjectToken:     "subject-token",
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
		ActorToken:       "",
		ActorTokenType:   string(constants.TokenTypeIdentifierAccessToken),
	}

	result := suite.handler.ValidateGrant(tokenRequest, suite.oauthApp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, result.Error)
	assert.Equal(suite.T(), "actor_token_type must not be provided without actor_token", result.ErrorDescription)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestValidateGrant_InvalidResourceURI() {
	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         "client123",
		SubjectToken:     "subject-token",
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
		Resource:         "not-a-valid-uri",
	}

	result := suite.handler.ValidateGrant(tokenRequest, suite.oauthApp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, result.Error)
	assert.Contains(suite.T(), result.ErrorDescription, "Invalid resource parameter")
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestValidateGrant_ResourceURIWithFragment() {
	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         "client123",
		SubjectToken:     "subject-token",
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
		Resource:         "https://api.example.com/resource#fragment",
	}

	result := suite.handler.ValidateGrant(tokenRequest, suite.oauthApp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, result.Error)
	assert.Contains(suite.T(), result.ErrorDescription, "must not contain a fragment component")
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestValidateGrant_ValidResourceURI() {
	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         "client123",
		SubjectToken:     "subject-token",
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
		Resource:         "https://api.example.com/resource",
	}

	result := suite.handler.ValidateGrant(tokenRequest, suite.oauthApp)
	assert.Nil(suite.T(), result)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestValidateGrant_UnsupportedRequestedTokenType() {
	tokenRequest := &model.TokenRequest{
		GrantType:          string(constants.GrantTypeTokenExchange),
		ClientID:           "client123",
		SubjectToken:       "subject-token",
		SubjectTokenType:   string(constants.TokenTypeIdentifierAccessToken),
		RequestedTokenType: "urn:ietf:params:oauth:token-type:saml2",
	}

	result := suite.handler.ValidateGrant(tokenRequest, suite.oauthApp)
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
		"iss":   "https://issuer.com",
		"aud":   "app123",
		"exp":   float64(now + 3600),
		"nbf":   float64(now - 60),
		"scope": "read write",
		"email": "user@example.com",
	})

	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         "client123",
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
	}

	suite.mockJWTService.On("VerifyJWTSignature", subjectToken).Return(nil)
	suite.mockJWTService.On("GenerateJWT",
		"user123",
		"app123",
		"https://custom.issuer.com",
		int64(7200),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			// aud is set from the GenerateJWT parameter, not in claims map
			return claims["scope"] == testScopeReadWrite &&
				claims["client_id"] == "client123" &&
				claims["email"] == "user@example.com"
		}),
	).Return(testTokenExchangeJWT, now, nil)

	ctx := &model.TokenContext{
		TokenAttributes: make(map[string]interface{}),
	}

	result, errResp := suite.handler.HandleGrant(tokenRequest, suite.oauthApp, ctx)

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
		"exp":   float64(now + 3600),
		"nbf":   float64(now - 60),
		"scope": "read write delete",
	})

	tokenRequest := suite.createBasicTokenRequest(subjectToken)
	tokenRequest.Scope = testScopeReadWrite

	suite.setupSuccessfulJWTMockWithScope(subjectToken, "client123", testScopeReadWrite, now)

	ctx := &model.TokenContext{}
	result, errResp := suite.handler.HandleGrant(tokenRequest, suite.oauthApp, ctx)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), []string{"read", "write"}, result.AccessToken.Scopes)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_Success_WithActorToken() {
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub": "user123",
		"iss": "https://subject-issuer.com",
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
	})

	actorToken := suite.createTestJWT(map[string]interface{}{
		"sub": "service456",
		"iss": "https://actor-issuer.com",
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
	})

	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         "client123",
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
		ActorToken:       actorToken,
		ActorTokenType:   string(constants.TokenTypeIdentifierAccessToken),
	}

	suite.mockJWTService.On("VerifyJWTSignature", subjectToken).Return(nil)
	suite.mockJWTService.On("VerifyJWTSignature", actorToken).Return(nil)
	suite.mockJWTService.On("GenerateJWT",
		"user123",
		"client123",
		"https://custom.issuer.com",
		int64(7200),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			act, ok := claims["act"].(map[string]interface{})
			return ok && act["sub"] == "service456" && act["iss"] == "https://actor-issuer.com"
		}),
	).Return(testTokenExchangeJWT, now, nil)

	ctx := &model.TokenContext{}
	result, errResp := suite.handler.HandleGrant(tokenRequest, suite.oauthApp, ctx)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_Success_WithActorChaining() {
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub": "user123",
		"iss": "https://subject-issuer.com",
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
		"act": map[string]interface{}{
			"sub": "service789",
			"iss": "https://existing-actor.com",
		},
	})

	actorToken := suite.createTestJWT(map[string]interface{}{
		"sub": "service456",
		"iss": "https://new-actor.com",
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
	})

	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         "client123",
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
		ActorToken:       actorToken,
		ActorTokenType:   string(constants.TokenTypeIdentifierAccessToken),
	}

	suite.mockJWTService.On("VerifyJWTSignature", subjectToken).Return(nil)
	suite.mockJWTService.On("VerifyJWTSignature", actorToken).Return(nil)
	suite.mockJWTService.On("GenerateJWT",
		"user123",
		"client123",
		"https://custom.issuer.com",
		int64(7200),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			act, ok := claims["act"].(map[string]interface{})
			if !ok {
				return false
			}
			// New actor
			if act["sub"] != "service456" || act["iss"] != "https://new-actor.com" {
				return false
			}
			// Chained actor
			chainedAct, ok := act["act"].(map[string]interface{})
			return ok && chainedAct["sub"] == "service789" && chainedAct["iss"] == "https://existing-actor.com"
		}),
	).Return(testTokenExchangeJWT, now, nil)

	ctx := &model.TokenContext{}
	result, errResp := suite.handler.HandleGrant(tokenRequest, suite.oauthApp, ctx)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_Success_WithAudienceParameter() {
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub": "user123",
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
	})

	tokenRequest := suite.createBasicTokenRequest(subjectToken)
	tokenRequest.Audience = "https://api.example.com"

	suite.setupSuccessfulJWTMock(subjectToken, "https://api.example.com", now)

	ctx := &model.TokenContext{}
	result, errResp := suite.handler.HandleGrant(tokenRequest, suite.oauthApp, ctx)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_Success_WithResourceParameter() {
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub": "user123",
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
	})

	tokenRequest := suite.createBasicTokenRequest(subjectToken)
	tokenRequest.Resource = "https://resource.example.com"

	suite.setupSuccessfulJWTMock(subjectToken, "https://resource.example.com", now)

	ctx := &model.TokenContext{}
	result, errResp := suite.handler.HandleGrant(tokenRequest, suite.oauthApp, ctx)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_Success_WithMultipleSpacesInScope() {
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub":   "user123",
		"exp":   float64(now + 3600),
		"nbf":   float64(now - 60),
		"scope": "read write",
	})

	tokenRequest := suite.createBasicTokenRequest(subjectToken)
	tokenRequest.Scope = "  read    write  "

	suite.setupSuccessfulJWTMockWithScope(subjectToken, "client123", testScopeReadWrite, now)

	ctx := &model.TokenContext{}
	result, errResp := suite.handler.HandleGrant(tokenRequest, suite.oauthApp, ctx)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), []string{"read", "write"}, result.AccessToken.Scopes)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_Success_PreservesUserAttributes() {
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub":   "user123",
		"exp":   float64(now + 3600),
		"nbf":   float64(now - 60),
		"email": "user@example.com",
		"name":  "Test User",
		"roles": []string{"admin", "user"},
	})

	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         "client123",
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
	}

	suite.mockJWTService.On("VerifyJWTSignature", subjectToken).Return(nil)
	suite.mockJWTService.On("GenerateJWT",
		"user123",
		"client123",
		"https://custom.issuer.com",
		int64(7200),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			return claims["email"] == "user@example.com" &&
				claims["name"] == "Test User"
		}),
	).Return(testTokenExchangeJWT, now, nil)

	ctx := &model.TokenContext{}
	result, errResp := suite.handler.HandleGrant(tokenRequest, suite.oauthApp, ctx)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "user@example.com", result.AccessToken.UserAttributes["email"])
	assert.Equal(suite.T(), "Test User", result.AccessToken.UserAttributes["name"])
}

// ============================================================================
// HandleGrant Tests - Error Cases
// ============================================================================

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_InvalidSubjectToken_SignatureError() {
	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         "client123",
		SubjectToken:     "invalid-token",
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
	}

	suite.mockJWTService.On("VerifyJWTSignature", "invalid-token").
		Return(errors.New("invalid signature"))

	ctx := &model.TokenContext{}
	result, errResp := suite.handler.HandleGrant(tokenRequest, suite.oauthApp, ctx)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), errResp)
	assert.Equal(suite.T(), constants.ErrorInvalidGrant, errResp.Error)
	assert.Contains(suite.T(), errResp.ErrorDescription, "Invalid subject_token")
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_InvalidSubjectToken_MissingSubClaim() {
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
	})

	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         "client123",
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
	}

	suite.mockJWTService.On("VerifyJWTSignature", subjectToken).Return(nil)

	ctx := &model.TokenContext{}
	result, errResp := suite.handler.HandleGrant(tokenRequest, suite.oauthApp, ctx)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), errResp)
	assert.Equal(suite.T(), constants.ErrorInvalidGrant, errResp.Error)
	assert.Contains(suite.T(), errResp.ErrorDescription, "missing or invalid 'sub' claim")
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_InvalidSubjectToken_DecodeError() {
	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         "client123",
		SubjectToken:     "invalid.jwt.format",
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
	}

	suite.mockJWTService.On("VerifyJWTSignature", "invalid.jwt.format").Return(nil)

	ctx := &model.TokenContext{}
	result, errResp := suite.handler.HandleGrant(tokenRequest, suite.oauthApp, ctx)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), errResp)
	assert.Equal(suite.T(), constants.ErrorInvalidGrant, errResp.Error)
	assert.Contains(suite.T(), errResp.ErrorDescription, "failed to decode token")
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_InvalidSubjectToken_Expired() {
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub": "user123",
		"exp": float64(now - 3600),
		"nbf": float64(now - 7200),
	})

	tokenRequest := suite.createBasicTokenRequest(subjectToken)
	suite.executeHandleGrantWithInvalidToken(tokenRequest, constants.ErrorInvalidGrant, "token has expired")
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_InvalidSubjectToken_NotYetValid() {
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub": "user123",
		"exp": float64(now + 3600),
		"nbf": float64(now + 1800),
	})

	tokenRequest := suite.createBasicTokenRequest(subjectToken)
	suite.executeHandleGrantWithInvalidToken(tokenRequest, constants.ErrorInvalidGrant, "token not yet valid")
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_InvalidActorToken() {
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub": "user123",
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
	})

	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         "client123",
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
		ActorToken:       "invalid-actor-token",
		ActorTokenType:   string(constants.TokenTypeIdentifierAccessToken),
	}

	suite.mockJWTService.On("VerifyJWTSignature", subjectToken).Return(nil)
	suite.mockJWTService.On("VerifyJWTSignature", "invalid-actor-token").
		Return(errors.New("invalid signature"))

	ctx := &model.TokenContext{}
	result, errResp := suite.handler.HandleGrant(tokenRequest, suite.oauthApp, ctx)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), errResp)
	assert.Equal(suite.T(), constants.ErrorInvalidGrant, errResp.Error)
	assert.Contains(suite.T(), errResp.ErrorDescription, "Invalid actor_token")
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_InvalidScope() {
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub":   "user123",
		"exp":   float64(now + 3600),
		"nbf":   float64(now - 60),
		"scope": "read write",
	})

	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         "client123",
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
		Scope:            "read write delete",
	}

	suite.mockJWTService.On("VerifyJWTSignature", subjectToken).Return(nil)

	ctx := &model.TokenContext{}
	result, errResp := suite.handler.HandleGrant(tokenRequest, suite.oauthApp, ctx)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), errResp)
	assert.Equal(suite.T(), constants.ErrorInvalidScope, errResp.Error)
	assert.Contains(suite.T(), errResp.ErrorDescription, "not present in subject token")
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_ScopeEscalationPrevention() {
	now := time.Now().Unix()
	// Subject token has NO scopes
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub": "user123",
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
	})

	// Request tries to add scopes
	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         "client123",
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
		Scope:            "read write",
	}

	suite.mockJWTService.On("VerifyJWTSignature", subjectToken).Return(nil)

	ctx := &model.TokenContext{}
	result, errResp := suite.handler.HandleGrant(tokenRequest, suite.oauthApp, ctx)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), errResp)
	assert.Equal(suite.T(), constants.ErrorInvalidScope, errResp.Error)
	assert.Contains(suite.T(), errResp.ErrorDescription, "Cannot request scopes when subject token has no scopes")
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_JWTGenerationError() {
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub": "user123",
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
	})

	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         "client123",
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
	}

	suite.mockJWTService.On("VerifyJWTSignature", subjectToken).Return(nil)
	suite.mockJWTService.On("GenerateJWT",
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
	).Return("", int64(0), errors.New("failed to sign token"))

	ctx := &model.TokenContext{}
	result, errResp := suite.handler.HandleGrant(tokenRequest, suite.oauthApp, ctx)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), errResp)
	assert.Equal(suite.T(), constants.ErrorServerError, errResp.Error)
	assert.Equal(suite.T(), "Failed to generate token", errResp.ErrorDescription)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_UsesDefaultConfig() {
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub": "user123",
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
	})

	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         "client123",
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
	}

	// Use app without custom token config
	oauthAppNoConfig := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:   "client123",
		GrantTypes: []constants.GrantType{constants.GrantTypeTokenExchange},
	}

	suite.mockJWTService.On("VerifyJWTSignature", subjectToken).Return(nil)
	suite.mockJWTService.On("GenerateJWT",
		"user123",
		"client123",
		"https://test.thunder.io",
		int64(3600),
		mock.Anything,
	).Return(testTokenExchangeJWT, now, nil)

	ctx := &model.TokenContext{}
	result, errResp := suite.handler.HandleGrant(tokenRequest, oauthAppNoConfig, ctx)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_Success_WithJWTTokenType() {
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub": "user123",
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
	})

	tokenRequest := &model.TokenRequest{
		GrantType:          string(constants.GrantTypeTokenExchange),
		ClientID:           "client123",
		SubjectToken:       subjectToken,
		SubjectTokenType:   string(constants.TokenTypeIdentifierAccessToken),
		RequestedTokenType: string(constants.TokenTypeIdentifierJWT),
	}

	suite.mockJWTService.On("VerifyJWTSignature", subjectToken).Return(nil)
	suite.mockJWTService.On("GenerateJWT",
		"user123",
		"client123",
		"https://custom.issuer.com",
		int64(7200),
		mock.Anything,
	).Return(testTokenExchangeJWT, now, nil)

	ctx := &model.TokenContext{}
	result, errResp := suite.handler.HandleGrant(tokenRequest, suite.oauthApp, ctx)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)
	assert.NotEmpty(suite.T(), result.AccessToken.Token)
	assert.Equal(suite.T(), string(constants.TokenTypeIdentifierJWT), ctx.TokenAttributes["issued_token_type"])
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestValidateGrant_UnsupportedIDTokenType() {
	tokenRequest := &model.TokenRequest{
		GrantType:          string(constants.GrantTypeTokenExchange),
		ClientID:           "client123",
		SubjectToken:       "subject-token",
		SubjectTokenType:   string(constants.TokenTypeIdentifierAccessToken),
		RequestedTokenType: string(constants.TokenTypeIdentifierIDToken),
	}

	errResp := suite.handler.ValidateGrant(tokenRequest, suite.oauthApp)
	assert.NotNil(suite.T(), errResp)
	assert.Equal(suite.T(), constants.ErrorInvalidTarget, errResp.Error)
	assert.Contains(suite.T(), errResp.ErrorDescription, "not supported")
	assert.Contains(suite.T(), errResp.ErrorDescription, string(constants.TokenTypeIdentifierIDToken))
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestValidateGrant_UnsupportedRefreshTokenType() {
	tokenRequest := &model.TokenRequest{
		GrantType:          string(constants.GrantTypeTokenExchange),
		ClientID:           "client123",
		SubjectToken:       "subject-token",
		SubjectTokenType:   string(constants.TokenTypeIdentifierAccessToken),
		RequestedTokenType: string(constants.TokenTypeIdentifierRefreshToken),
	}

	// Test ValidateGrant first (which is called before HandleGrant in production)
	errResp := suite.handler.ValidateGrant(tokenRequest, suite.oauthApp)
	assert.NotNil(suite.T(), errResp)
	assert.Equal(suite.T(), constants.ErrorInvalidTarget, errResp.Error)
	assert.Contains(suite.T(), errResp.ErrorDescription, "not supported")
	assert.Contains(suite.T(), errResp.ErrorDescription, string(constants.TokenTypeIdentifierRefreshToken))
}

// ============================================================================
// RFC 8693 Compliance Integration Tests
// ============================================================================

func (suite *TokenExchangeGrantHandlerTestSuite) TestRFC8693_CompleteTokenExchangeFlow() {
	// RFC 8693 Section 2.2: Verify all required response parameters
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub":   "user123",
		"iss":   "https://original-issuer.com",
		"aud":   "original-audience",
		"exp":   float64(now + 3600),
		"nbf":   float64(now - 60),
		"scope": "read write",
		"email": "user@example.com",
		"name":  "John Doe",
	})

	tokenRequest := &model.TokenRequest{
		GrantType:          string(constants.GrantTypeTokenExchange),
		ClientID:           "client123",
		SubjectToken:       subjectToken,
		SubjectTokenType:   string(constants.TokenTypeIdentifierAccessToken),
		RequestedTokenType: string(constants.TokenTypeIdentifierAccessToken),
		Audience:           "https://target-service.com",
		Scope:              "read",
	}

	suite.mockJWTService.On("VerifyJWTSignature", subjectToken).Return(nil)
	suite.mockJWTService.On("GenerateJWT",
		"user123",
		"https://target-service.com",
		"https://custom.issuer.com",
		int64(7200),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			// Verify claims structure per RFC 8693
			// aud is passed as GenerateJWT parameter, not in claims map
			assert.Equal(suite.T(), "client123", claims["client_id"])
			assert.Equal(suite.T(), "read", claims["scope"])
			assert.Equal(suite.T(), "user@example.com", claims["email"])
			assert.Equal(suite.T(), "John Doe", claims["name"])
			return true
		}),
	).Return(testTokenExchangeJWT, now, nil)

	ctx := &model.TokenContext{}
	result, errResp := suite.handler.HandleGrant(tokenRequest, suite.oauthApp, ctx)

	// RFC 8693 Section 2.2: Verify required response parameters
	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)
	assert.NotEmpty(suite.T(), result.AccessToken.Token)                             // access_token - REQUIRED
	assert.Equal(suite.T(), constants.TokenTypeBearer, result.AccessToken.TokenType) // token_type - REQUIRED
	assert.NotZero(suite.T(), result.AccessToken.ExpiresIn)                          // expires_in - RECOMMENDED
	assert.Equal(suite.T(), []string{"read"}, result.AccessToken.Scopes)
	// issued_token_type - REQUIRED
	assert.Equal(suite.T(), string(constants.TokenTypeIdentifierAccessToken), ctx.TokenAttributes["issued_token_type"])
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestRFC8693_AudiencePriority() {
	// RFC 8693: Test audience parameter priority (audience > resource > token.aud > client_id)
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub": "user123",
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
		"aud": "token-audience",
	})

	// Test 1: Audience parameter takes priority
	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         "client123",
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
		Audience:         "request-audience",
		Resource:         "https://resource.example.com",
	}

	suite.mockJWTService.On("VerifyJWTSignature", subjectToken).Return(nil)
	suite.mockJWTService.On("GenerateJWT",
		"user123",
		"request-audience", // Should use request audience, not resource or token aud
		mock.Anything,
		mock.Anything,
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			// aud is set from the GenerateJWT parameter, not in claims map
			return true
		}),
	).Return(testTokenExchangeJWT, now, nil)

	ctx := &model.TokenContext{}
	result, errResp := suite.handler.HandleGrant(tokenRequest, suite.oauthApp, ctx)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestRFC8693_ActorDelegationChain() {
	// RFC 8693 Section 4.1: Test nested actor delegation chains
	now := time.Now().Unix()

	// Subject token with existing actor
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub": "user123",
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
		"iss": "https://actor-issuer.com",
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
		"act": map[string]interface{}{
			"sub": "actor-of-actor",
			"iss": "https://nested-issuer.com",
		},
	})

	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         "client123",
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
		ActorToken:       actorToken,
		ActorTokenType:   string(constants.TokenTypeIdentifierAccessToken),
	}

	suite.mockJWTService.On("VerifyJWTSignature", subjectToken).Return(nil)
	suite.mockJWTService.On("VerifyJWTSignature", actorToken).Return(nil)
	suite.mockJWTService.On("GenerateJWT",
		"user123",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			// Verify nested delegation chain per RFC 8693
			act, ok := claims["act"].(map[string]interface{})
			if !ok {
				return false
			}

			// Current actor
			assert.Equal(suite.T(), "current-actor", act["sub"])
			assert.Equal(suite.T(), "https://actor-issuer.com", act["iss"])

			// Nested chain: actor's act -> subject's act
			nestedAct, ok := act["act"].(map[string]interface{})
			if !ok {
				return false
			}
			assert.Equal(suite.T(), "actor-of-actor", nestedAct["sub"])

			furtherNested, ok := nestedAct["act"].(map[string]interface{})
			if !ok {
				return false
			}
			assert.Equal(suite.T(), "previous-actor", furtherNested["sub"])

			return true
		}),
	).Return(testTokenExchangeJWT, now, nil)

	ctx := &model.TokenContext{}
	result, errResp := suite.handler.HandleGrant(tokenRequest, suite.oauthApp, ctx)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestHandleGrant_Success_WithActorTokenHasActButSubjectHasNoAct() {
	// Test case: Actor token has its own act claim, but subject token has no act claim
	// This covers lines 358-359 where actClaim["act"] = actorAct
	now := time.Now().Unix()

	// Subject token WITHOUT act claim
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub": "user123",
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
	})

	// Actor token WITH its own act claim
	actorToken := suite.createTestJWT(map[string]interface{}{
		"sub": "current-actor",
		"iss": "https://actor-issuer.com",
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
		"act": map[string]interface{}{
			"sub": "actor-of-actor",
			"iss": "https://nested-issuer.com",
		},
	})

	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         "client123",
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
		ActorToken:       actorToken,
		ActorTokenType:   string(constants.TokenTypeIdentifierAccessToken),
	}

	suite.mockJWTService.On("VerifyJWTSignature", subjectToken).Return(nil)
	suite.mockJWTService.On("VerifyJWTSignature", actorToken).Return(nil)
	suite.mockJWTService.On("GenerateJWT",
		"user123",
		"client123",
		"https://custom.issuer.com",
		int64(7200),
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			// Verify actor claim structure
			act, ok := claims["act"].(map[string]interface{})
			if !ok {
				return false
			}

			// Current actor should be present
			if act["sub"] != "current-actor" || act["iss"] != "https://actor-issuer.com" {
				return false
			}

			// Actor's act claim should be preserved directly (lines 358-359)
			actorAct, ok := act["act"].(map[string]interface{})
			if !ok {
				return false
			}

			// Verify the actor's act claim is preserved as-is
			if actorAct["sub"] != "actor-of-actor" || actorAct["iss"] != "https://nested-issuer.com" {
				return false
			}

			// Subject has no act claim, so it should not be nested
			_, hasFurtherNesting := actorAct["act"]
			return !hasFurtherNesting
		}),
	).Return(testTokenExchangeJWT, now, nil)

	ctx := &model.TokenContext{}
	result, errResp := suite.handler.HandleGrant(tokenRequest, suite.oauthApp, ctx)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestRFC8693_ScopeDownscopingEnforcement() {
	// RFC 8693 Section 5: Verify scope downscoping (security consideration)
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub":   "user123",
		"exp":   float64(now + 3600),
		"nbf":   float64(now - 60),
		"scope": "read write delete",
	})

	// Test 1: Valid downscoping (subset of scopes)
	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         "client123",
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
		Scope:            "read",
	}

	suite.mockJWTService.On("VerifyJWTSignature", subjectToken).Return(nil)
	suite.mockJWTService.On("GenerateJWT",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			assert.Equal(suite.T(), "read", claims["scope"])
			return true
		}),
	).Return(testTokenExchangeJWT, now, nil)

	ctx := &model.TokenContext{}
	result, errResp := suite.handler.HandleGrant(tokenRequest, suite.oauthApp, ctx)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), []string{"read"}, result.AccessToken.Scopes)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestRFC8693_ResourceParameterValidation() {
	// RFC 8693 Section 2.1: Resource must be absolute URI without fragment
	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         "client123",
		SubjectToken:     "subject-token",
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
		Resource:         "https://api.example.com/v1/resource",
	}

	result := suite.handler.ValidateGrant(tokenRequest, suite.oauthApp)
	assert.Nil(suite.T(), result)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestRFC8693_NoTokenLinkage() {
	// RFC 8693 Section 2.1: "exchange has no impact on the validity of the subject token"
	// This is a design verification test - token exchange should not invalidate input tokens
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub": "user123",
		"exp": float64(now + 3600),
		"nbf": float64(now - 60),
	})

	tokenRequest := &model.TokenRequest{
		GrantType:        string(constants.GrantTypeTokenExchange),
		ClientID:         "client123",
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
	}

	suite.mockJWTService.On("VerifyJWTSignature", subjectToken).Return(nil)
	suite.mockJWTService.On("GenerateJWT",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(testTokenExchangeJWT, now, nil)

	ctx := &model.TokenContext{}
	result1, errResp1 := suite.handler.HandleGrant(tokenRequest, suite.oauthApp, ctx)

	assert.Nil(suite.T(), errResp1)
	assert.NotNil(suite.T(), result1)

	// Use same subject token again - should succeed (no linkage/invalidation)
	ctx2 := &model.TokenContext{}
	result2, errResp2 := suite.handler.HandleGrant(tokenRequest, suite.oauthApp, ctx2)

	assert.Nil(suite.T(), errResp2)
	assert.NotNil(suite.T(), result2)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestRFC8693_ClaimPreservation() {
	// Verify non-standard claims are preserved through token exchange
	now := time.Now().Unix()
	subjectToken := suite.createTestJWT(map[string]interface{}{
		"sub":          "user123",
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
		ClientID:         "client123",
		SubjectToken:     subjectToken,
		SubjectTokenType: string(constants.TokenTypeIdentifierAccessToken),
	}

	suite.mockJWTService.On("VerifyJWTSignature", subjectToken).Return(nil)
	suite.mockJWTService.On("GenerateJWT",
		"user123",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			// Verify all custom claims are preserved
			assert.Equal(suite.T(), "user@example.com", claims["email"])
			assert.Equal(suite.T(), "John", claims["given_name"])
			assert.Equal(suite.T(), "Doe", claims["family_name"])
			assert.Equal(suite.T(), "ACME Corp", claims["organization"])

			roles, ok := claims["roles"].([]interface{})
			assert.True(suite.T(), ok)
			assert.Equal(suite.T(), 2, len(roles))

			return true
		}),
	).Return(testTokenExchangeJWT, now, nil)

	ctx := &model.TokenContext{}
	result, errResp := suite.handler.HandleGrant(tokenRequest, suite.oauthApp, ctx)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)

	// Verify user attributes in response
	assert.Equal(suite.T(), "user@example.com", result.AccessToken.UserAttributes["email"])
	assert.Equal(suite.T(), "John", result.AccessToken.UserAttributes["given_name"])
	assert.Equal(suite.T(), "Doe", result.AccessToken.UserAttributes["family_name"])
	assert.Equal(suite.T(), "ACME Corp", result.AccessToken.UserAttributes["organization"])
}

// ============================================================================
// Helper Method Tests
// ============================================================================

func (suite *TokenExchangeGrantHandlerTestSuite) TestIsSupportedTokenType() {
	assert.True(suite.T(), constants.TokenTypeIdentifierAccessToken.IsValid())
	assert.True(suite.T(), constants.TokenTypeIdentifierRefreshToken.IsValid())
	assert.True(suite.T(), constants.TokenTypeIdentifierIDToken.IsValid())
	assert.True(suite.T(), constants.TokenTypeIdentifierJWT.IsValid())
	assert.False(suite.T(), constants.TokenTypeIdentifier("urn:ietf:params:oauth:token-type:saml2").IsValid())
	assert.False(suite.T(), constants.TokenTypeIdentifier("invalid").IsValid())
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestGetAudience_WithSubjectTokenStringAudience() {
	// Test that getAudience returns string audience as-is
	tokenRequest := &model.TokenRequest{
		GrantType: string(constants.GrantTypeTokenExchange),
		ClientID:  "client123",
	}

	subjectClaims := map[string]interface{}{
		"aud": "string-audience",
	}

	result := suite.handler.getAudience(tokenRequest, subjectClaims)
	assert.Equal(suite.T(), "string-audience", result)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestExtractUserAttributes() {
	claims := map[string]interface{}{
		"sub":       "user123",
		"iss":       "issuer",
		"aud":       "audience",
		"exp":       float64(123456789),
		"nbf":       float64(123456789),
		"iat":       float64(123456789),
		"jti":       "jwt-id",
		"scope":     "read write",
		"client_id": "client123",
		"act":       map[string]interface{}{"sub": "actor"},
		"email":     "user@example.com",
		"name":      "Test User",
		"custom":    "value",
	}

	userAttrs := suite.handler.extractUserAttributes(claims)

	assert.Equal(suite.T(), 3, len(userAttrs))
	assert.Equal(suite.T(), "user@example.com", userAttrs["email"])
	assert.Equal(suite.T(), "Test User", userAttrs["name"])
	assert.Equal(suite.T(), "value", userAttrs["custom"])
	assert.NotContains(suite.T(), userAttrs, "sub")
	assert.NotContains(suite.T(), userAttrs, "iss")
	assert.NotContains(suite.T(), userAttrs, "scope")
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestgetAudience_Priority() {
	claims := map[string]interface{}{
		"aud": "token-audience",
	}

	tokenRequest := &model.TokenRequest{
		ClientID: "client123",
		Audience: "request-audience",
		Resource: "request-resource",
	}

	// Audience parameter has highest priority
	aud := suite.handler.getAudience(tokenRequest, claims)
	assert.Equal(suite.T(), "request-audience", aud)

	// Resource parameter is second priority
	tokenRequest.Audience = ""
	aud = suite.handler.getAudience(tokenRequest, claims)
	assert.Equal(suite.T(), "request-resource", aud)

	// Token audience is third priority
	tokenRequest.Resource = ""
	aud = suite.handler.getAudience(tokenRequest, claims)
	assert.Equal(suite.T(), "token-audience", aud)

	// Client ID is fallback
	delete(claims, "aud")
	aud = suite.handler.getAudience(tokenRequest, claims)
	assert.Equal(suite.T(), "client123", aud)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestgetScopes_EmptyRequest() {
	claims := map[string]interface{}{
		"scope": "read write",
	}

	tokenRequest := &model.TokenRequest{
		Scope: "",
	}

	scopes, errResp := suite.handler.getScopes(tokenRequest, claims)
	assert.Nil(suite.T(), errResp)
	assert.Equal(suite.T(), []string{"read", "write"}, scopes)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestgetScopes_EmptyTokenScope() {
	claims := map[string]interface{}{}

	tokenRequest := &model.TokenRequest{
		Scope: "",
	}

	scopes, errResp := suite.handler.getScopes(tokenRequest, claims)
	assert.Nil(suite.T(), errResp)
	assert.Equal(suite.T(), []string{}, scopes)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestgetScopes_AllWhitespace() {
	claims := map[string]interface{}{
		"scope": "read write",
	}

	tokenRequest := &model.TokenRequest{
		Scope: "   ",
	}

	scopes, errResp := suite.handler.getScopes(tokenRequest, claims)
	assert.Nil(suite.T(), errResp)
	assert.Equal(suite.T(), []string{}, scopes)
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestBuildTokenClaims_NoActorNoSubjectActor() {
	tokenRequest := &model.TokenRequest{
		ClientID: "client123",
	}

	userAttrs := map[string]interface{}{
		"email": "user@example.com",
	}

	claims := suite.handler.buildTokenClaims(
		tokenRequest,
		nil,
		nil,
		userAttrs,
		[]string{"read", "write"},
	)

	assert.Equal(suite.T(), testScopeReadWrite, claims["scope"])
	assert.Equal(suite.T(), "client123", claims["client_id"])
	assert.Equal(suite.T(), "user@example.com", claims["email"])
	assert.NotContains(suite.T(), claims, "act")
	assert.NotContains(suite.T(), claims, "aud") // aud is set separately in HandleGrant
}

func (suite *TokenExchangeGrantHandlerTestSuite) TestBuildTokenClaims_WithSubjectActorNoNewActor() {
	tokenRequest := &model.TokenRequest{
		ClientID: "client123",
	}

	subjectActor := map[string]interface{}{
		"sub": "existing-actor",
		"iss": "https://existing-issuer.com",
	}

	claims := suite.handler.buildTokenClaims(
		tokenRequest,
		nil,
		subjectActor,
		map[string]interface{}{},
		[]string{},
	)

	assert.Equal(suite.T(), subjectActor, claims["act"])
}
