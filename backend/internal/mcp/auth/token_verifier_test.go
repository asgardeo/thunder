/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mcpauth "github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/tests/mocks/jwtmock"
)

const (
	testIssuer = "https://localhost:8090"
	testMCPURL = "https://localhost:8090/mcp"
)

type TokenVerifierTestSuite struct {
	suite.Suite
	mockJWTService *jwtmock.JWTServiceInterfaceMock
}

func TestTokenVerifierTestSuite(t *testing.T) {
	suite.Run(t, new(TokenVerifierTestSuite))
}

func (suite *TokenVerifierTestSuite) SetupTest() {
	suite.mockJWTService = jwtmock.NewJWTServiceInterfaceMock(suite.T())
}

// createTestJWT creates a test JWT token with the given payload.
func createTestJWT(payload map[string]interface{}) string {
	header := map[string]string{"alg": "RS256", "typ": "JWT"}
	headerJSON, _ := json.Marshal(header)
	payloadJSON, _ := json.Marshal(payload)

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

	// Signature doesn't matter for these tests as we mock VerifyJWT
	return headerB64 + "." + payloadB64 + ".signature"
}

func (suite *TokenVerifierTestSuite) TestNewTokenVerifier_Success() {
	// Create a valid token payload
	futureExp := time.Now().Add(1 * time.Hour).Unix()
	payload := map[string]interface{}{
		"sub":   "user123",
		"iss":   testIssuer,
		"aud":   testMCPURL,
		"exp":   float64(futureExp),
		"scope": "system openid",
	}
	token := createTestJWT(payload)

	// Setup mock
	suite.mockJWTService.EXPECT().VerifyJWT(token, testMCPURL, testIssuer).Return(nil)

	// Create verifier and test
	verifier := NewTokenVerifier(suite.mockJWTService, testIssuer, testMCPURL)
	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)

	tokenInfo, err := verifier(context.Background(), token, req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), tokenInfo)
	assert.Equal(suite.T(), "user123", tokenInfo.UserID)
	assert.Equal(suite.T(), []string{"system", "openid"}, tokenInfo.Scopes)
	assert.Equal(suite.T(), time.Unix(futureExp, 0), tokenInfo.Expiration)
}

func (suite *TokenVerifierTestSuite) TestNewTokenVerifier_JWTVerificationFails() {
	token := createTestJWT(map[string]interface{}{
		"sub": "user123",
	})

	// Setup mock to return an error
	suite.mockJWTService.EXPECT().
		VerifyJWT(token, testMCPURL, testIssuer).
		Return(&serviceerror.ServiceError{Error: "invalid token"})

	// Create verifier and test
	verifier := NewTokenVerifier(suite.mockJWTService, testIssuer, testMCPURL)
	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)

	tokenInfo, err := verifier(context.Background(), token, req)

	assert.ErrorIs(suite.T(), err, mcpauth.ErrInvalidToken)
	assert.Nil(suite.T(), tokenInfo)
}

func (suite *TokenVerifierTestSuite) TestNewTokenVerifier_InvalidTokenFormat() {
	// Invalid JWT format (not 3 parts)
	invalidToken := "invalid.token"

	// Setup mock
	suite.mockJWTService.EXPECT().VerifyJWT(invalidToken, testMCPURL, testIssuer).Return(nil)

	// Create verifier and test
	verifier := NewTokenVerifier(suite.mockJWTService, testIssuer, testMCPURL)
	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)

	tokenInfo, err := verifier(context.Background(), invalidToken, req)

	// DecodeJWTPayload will fail
	assert.ErrorIs(suite.T(), err, mcpauth.ErrInvalidToken)
	assert.Nil(suite.T(), tokenInfo)
}

func (suite *TokenVerifierTestSuite) TestNewTokenVerifier_InvalidPayloadBase64() {
	// Valid structure but invalid base64 payload
	malformedJWT := "eyJhbGciOiJSUzI1NiJ9.!!!invalid-base64!!!.signature"

	// Setup mock
	suite.mockJWTService.EXPECT().VerifyJWT(malformedJWT, testMCPURL, testIssuer).Return(nil)

	// Create verifier and test
	verifier := NewTokenVerifier(suite.mockJWTService, testIssuer, testMCPURL)
	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)

	tokenInfo, err := verifier(context.Background(), malformedJWT, req)

	assert.ErrorIs(suite.T(), err, mcpauth.ErrInvalidToken)
	assert.Nil(suite.T(), tokenInfo)
}

func (suite *TokenVerifierTestSuite) TestNewTokenVerifier_MissingScopes() {
	// Token without scope claim
	futureExp := time.Now().Add(1 * time.Hour).Unix()
	payload := map[string]interface{}{
		"sub": "user123",
		"iss": testIssuer,
		"aud": testMCPURL,
		"exp": float64(futureExp),
	}
	token := createTestJWT(payload)

	// Setup mock
	suite.mockJWTService.EXPECT().VerifyJWT(token, testMCPURL, testIssuer).Return(nil)

	// Create verifier and test
	verifier := NewTokenVerifier(suite.mockJWTService, testIssuer, testMCPURL)
	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)

	tokenInfo, err := verifier(context.Background(), token, req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), tokenInfo)
	assert.Equal(suite.T(), "user123", tokenInfo.UserID)
	assert.Empty(suite.T(), tokenInfo.Scopes)
}

func (suite *TokenVerifierTestSuite) TestNewTokenVerifier_EmptyScopes() {
	// Token with empty scope string
	futureExp := time.Now().Add(1 * time.Hour).Unix()
	payload := map[string]interface{}{
		"sub":   "user123",
		"iss":   testIssuer,
		"aud":   testMCPURL,
		"exp":   float64(futureExp),
		"scope": "",
	}
	token := createTestJWT(payload)

	// Setup mock
	suite.mockJWTService.EXPECT().VerifyJWT(token, testMCPURL, testIssuer).Return(nil)

	// Create verifier and test
	verifier := NewTokenVerifier(suite.mockJWTService, testIssuer, testMCPURL)
	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)

	tokenInfo, err := verifier(context.Background(), token, req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), tokenInfo)
	assert.Empty(suite.T(), tokenInfo.Scopes)
}

func (suite *TokenVerifierTestSuite) TestNewTokenVerifier_MissingSubClaim() {
	// Token without sub claim
	futureExp := time.Now().Add(1 * time.Hour).Unix()
	payload := map[string]interface{}{
		"iss":   testIssuer,
		"aud":   testMCPURL,
		"exp":   float64(futureExp),
		"scope": "system",
	}
	token := createTestJWT(payload)

	// Setup mock
	suite.mockJWTService.EXPECT().VerifyJWT(token, testMCPURL, testIssuer).Return(nil)

	// Create verifier and test
	verifier := NewTokenVerifier(suite.mockJWTService, testIssuer, testMCPURL)
	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)

	tokenInfo, err := verifier(context.Background(), token, req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), tokenInfo)
	assert.Empty(suite.T(), tokenInfo.UserID)
	assert.Equal(suite.T(), []string{"system"}, tokenInfo.Scopes)
}

func (suite *TokenVerifierTestSuite) TestNewTokenVerifier_MissingExpClaim() {
	// Token without exp claim
	payload := map[string]interface{}{
		"sub":   "user123",
		"iss":   testIssuer,
		"aud":   testMCPURL,
		"scope": "system",
	}
	token := createTestJWT(payload)

	// Setup mock
	suite.mockJWTService.EXPECT().VerifyJWT(token, testMCPURL, testIssuer).Return(nil)

	// Create verifier and test
	verifier := NewTokenVerifier(suite.mockJWTService, testIssuer, testMCPURL)
	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)

	tokenInfo, err := verifier(context.Background(), token, req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), tokenInfo)
	assert.Equal(suite.T(), "user123", tokenInfo.UserID)
	// Expiration should be zero value when not present
	assert.True(suite.T(), tokenInfo.Expiration.IsZero())
}

func (suite *TokenVerifierTestSuite) TestNewTokenVerifier_MultipleScopes() {
	// Token with multiple scopes
	futureExp := time.Now().Add(1 * time.Hour).Unix()
	payload := map[string]interface{}{
		"sub":   "user123",
		"iss":   testIssuer,
		"aud":   testMCPURL,
		"exp":   float64(futureExp),
		"scope": "system openid profile email",
	}
	token := createTestJWT(payload)

	// Setup mock
	suite.mockJWTService.EXPECT().VerifyJWT(token, testMCPURL, testIssuer).Return(nil)

	// Create verifier and test
	verifier := NewTokenVerifier(suite.mockJWTService, testIssuer, testMCPURL)
	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)

	tokenInfo, err := verifier(context.Background(), token, req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), tokenInfo)
	assert.Equal(suite.T(), []string{"system", "openid", "profile", "email"}, tokenInfo.Scopes)
}
