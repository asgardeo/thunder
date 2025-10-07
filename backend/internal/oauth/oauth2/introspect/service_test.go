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

package introspect_test

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/introspect"
	"github.com/asgardeo/thunder/tests/mocks/jwtmock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type TokenIntrospectionServiceTestSuite struct {
	suite.Suite
	jwtServiceMock     *jwtmock.JWTServiceInterfaceMock
	introspectService  introspect.TokenIntrospectionServiceInterface
	validToken         string
	expiredToken       string
	notBeforeToken     string
	missingClaimsToken string
	privateKey         *rsa.PrivateKey
}

func TestTokenIntrospectionServiceTestSuite(t *testing.T) {
	suite.Run(t, new(TokenIntrospectionServiceTestSuite))
}

func (s *TokenIntrospectionServiceTestSuite) SetupTest() {
	s.jwtServiceMock = jwtmock.NewJWTServiceInterfaceMock(s.T())

	// Create a private key for signing JWT tokens
	var err error
	s.privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		s.T().Fatal("Error generating RSA key:", err)
	}

	s.introspectService = introspect.NewTokenIntrospectionService(s.jwtServiceMock)

	s.validToken = s.createValidToken()
	s.expiredToken = s.createExpiredToken()
	s.notBeforeToken = s.createNotBeforeToken()
	s.missingClaimsToken = s.createMissingClaimsToken()
}

func (s *TokenIntrospectionServiceTestSuite) TestIntrospectToken_EmptyToken() {
	response, err := s.introspectService.IntrospectToken("", "")
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "token is required")
	assert.Nil(s.T(), response)
}

func (s *TokenIntrospectionServiceTestSuite) TestIntrospectToken_PublicKeyNotAvailable() {
	s.jwtServiceMock.On("GetPublicKey").Return(nil).Maybe()
	s.jwtServiceMock.On("VerifyJWT", mock.Anything, "", "").Return(
		errors.New("public key not available"))

	response, err := s.introspectService.IntrospectToken(s.validToken, "")
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), response)
	assert.False(s.T(), response.Active)
	s.jwtServiceMock.AssertExpectations(s.T())
}

func (s *TokenIntrospectionServiceTestSuite) TestIntrospectToken_InvalidSignature() {
	s.jwtServiceMock.On("GetPublicKey").Return(&s.privateKey.PublicKey).Maybe()

	// Use a different private key to create an invalid signature
	differentKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	header := map[string]interface{}{"alg": "RS256", "typ": "JWT"}
	claims := map[string]interface{}{
		"exp": float64(time.Now().Add(time.Hour).Unix()),
		"nbf": float64(time.Now().Add(-time.Minute).Unix()),
	}

	headerBytes, _ := json.Marshal(header)
	claimsBytes, _ := json.Marshal(claims)
	headerEncoded := base64.RawURLEncoding.EncodeToString(headerBytes)
	claimsEncoded := base64.RawURLEncoding.EncodeToString(claimsBytes)

	signingInput := headerEncoded + "." + claimsEncoded
	hashed := sha256.Sum256([]byte(signingInput))
	signature, _ := rsa.SignPKCS1v15(rand.Reader, differentKey, crypto.SHA256, hashed[:])
	signatureEncoded := base64.RawURLEncoding.EncodeToString(signature)

	invalidToken := signingInput + "." + signatureEncoded

	s.jwtServiceMock.On("VerifyJWT", invalidToken, "", "").Return(
		errors.New("invalid signature"))

	// Test with a token having invalid signature
	response, err := s.introspectService.IntrospectToken(invalidToken, "")
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), response)
	assert.False(s.T(), response.Active)
	s.jwtServiceMock.AssertExpectations(s.T())
}

func (s *TokenIntrospectionServiceTestSuite) TestIntrospectToken() {
	testCases := []struct {
		name           string
		token          string
		tokenFn        func(*TokenIntrospectionServiceTestSuite) string
		expectError    bool
		errorContains  string
		active         bool
		expectedFields map[string]interface{}
	}{
		{
			name:        "InvalidTokenFormat",
			token:       "not-a-valid-jwt-token",
			expectError: false,
			active:      false,
		},
		{
			name:        "ExpiredToken",
			tokenFn:     func(s *TokenIntrospectionServiceTestSuite) string { return s.expiredToken },
			expectError: false,
			active:      false,
		},
		{
			name:        "FutureToken",
			tokenFn:     func(s *TokenIntrospectionServiceTestSuite) string { return s.notBeforeToken },
			expectError: false,
			active:      false,
		},
		{
			name:        "ValidToken",
			tokenFn:     func(s *TokenIntrospectionServiceTestSuite) string { return s.validToken },
			expectError: false,
			active:      true,
			expectedFields: map[string]interface{}{
				"TokenType": constants.TokenTypeBearer,
				"Scope":     "openid profile",
				"ClientID":  "client123",
				"Username":  "user@example.com",
				"Sub":       "user123",
				"Aud":       "api.example.com",
				"Iss":       "https://example.com",
				"Jti":       "token-id-123",
			},
		},
		{
			name: "TokenWithMissingExpClaim",
			tokenFn: func(s *TokenIntrospectionServiceTestSuite) string {
				claims := map[string]interface{}{
					"nbf": float64(time.Now().Add(-time.Minute).Unix()),
					"iat": float64(time.Now().Unix()),
				}
				return s.createToken(claims)
			},
			expectError: false,
			active:      false,
		},
		{
			name: "TokenWithMissingNbfClaim",
			tokenFn: func(s *TokenIntrospectionServiceTestSuite) string {
				claims := map[string]interface{}{
					"exp": float64(time.Now().Add(time.Hour).Unix()),
					"iat": float64(time.Now().Unix()),
				}
				return s.createToken(claims)
			},
			expectError: false,
			active:      false,
		},
		{
			name:        "TokenWithMissingOptionalClaims",
			tokenFn:     func(s *TokenIntrospectionServiceTestSuite) string { return s.missingClaimsToken },
			expectError: false,
			active:      true,
			expectedFields: map[string]interface{}{
				"TokenType": constants.TokenTypeBearer,
				"Scope":     "",
				"ClientID":  "",
				"Username":  "",
				"Sub":       "",
				"Aud":       "",
				"Iss":       "",
				"Jti":       "",
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.jwtServiceMock.On("GetPublicKey").Return(&s.privateKey.PublicKey).Maybe()

			var token string
			if tc.token != "" {
				token = tc.token
			} else if tc.tokenFn != nil {
				token = tc.tokenFn(s)
			}

			// Mock VerifyJWT based on test case
			switch tc.name {
			case "InvalidTokenFormat":
				s.jwtServiceMock.On("VerifyJWT", token, "", "").Return(
					errors.New("invalid token format"))
			case "ExpiredToken":
				s.jwtServiceMock.On("VerifyJWT", token, "", "").Return(
					errors.New("token has expired"))
			case "FutureToken":
				s.jwtServiceMock.On("VerifyJWT", token, "", "").Return(
					errors.New("token not valid yet (nbf)"))
			case "TokenWithMissingExpClaim":
				s.jwtServiceMock.On("VerifyJWT", token, "", "").Return(
					errors.New("missing or invalid 'exp' claim"))
			case "TokenWithMissingNbfClaim":
				s.jwtServiceMock.On("VerifyJWT", token, "", "").Return(
					errors.New("missing or invalid 'nbf' claim"))
			case "ValidToken", "TokenWithMissingOptionalClaims":
				s.jwtServiceMock.On("VerifyJWT", token, "", "").Return(nil)
			default:
				s.jwtServiceMock.On("VerifyJWT", token, "", "").Return(nil)
			}

			response, err := s.introspectService.IntrospectToken(token, "")

			if tc.expectError {
				assert.Error(s.T(), err)
				if tc.errorContains != "" {
					assert.Contains(s.T(), err.Error(), tc.errorContains)
				}
			} else {
				assert.NoError(s.T(), err)
				assert.NotNil(s.T(), response)
				assert.Equal(s.T(), tc.active, response.Active)

				if tc.expectedFields != nil && tc.active {
					// Verify expected fields
					for field, value := range tc.expectedFields {
						switch field {
						case "TokenType":
							assert.Equal(s.T(), value, response.TokenType)
						case "Scope":
							assert.Equal(s.T(), value, response.Scope)
						case "ClientID":
							assert.Equal(s.T(), value, response.ClientID)
						case "Username":
							assert.Equal(s.T(), value, response.Username)
						case "Sub":
							assert.Equal(s.T(), value, response.Sub)
						case "Aud":
							assert.Equal(s.T(), value, response.Aud)
						case "Iss":
							assert.Equal(s.T(), value, response.Iss)
						case "Jti":
							assert.Equal(s.T(), value, response.Jti)
						}
					}
				}
			}
			s.jwtServiceMock.AssertExpectations(s.T())
		})
	}
}

// Helper methods to create tokens with specific claims
func (s *TokenIntrospectionServiceTestSuite) createToken(claims map[string]interface{}) string {
	header := map[string]interface{}{
		"alg": "RS256",
		"typ": "JWT",
	}

	headerBytes, _ := json.Marshal(header)
	claimsBytes, _ := json.Marshal(claims)

	headerEncoded := base64.RawURLEncoding.EncodeToString(headerBytes)
	claimsEncoded := base64.RawURLEncoding.EncodeToString(claimsBytes)

	signingInput := headerEncoded + "." + claimsEncoded
	hashed := sha256.Sum256([]byte(signingInput))
	signature, err := rsa.SignPKCS1v15(rand.Reader, s.privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		s.T().Fatal("Error signing token:", err)
	}
	signatureEncoded := base64.RawURLEncoding.EncodeToString(signature)

	return signingInput + "." + signatureEncoded
}

func (s *TokenIntrospectionServiceTestSuite) createValidToken() string {
	claims := map[string]interface{}{
		"exp":       float64(time.Now().Add(time.Hour).Unix()),
		"nbf":       float64(time.Now().Add(-time.Minute).Unix()),
		"iat":       float64(time.Now().Unix()),
		"jti":       "token-id-123",
		"scope":     "openid profile",
		"client_id": "client123",
		"username":  "user@example.com",
		"sub":       "user123",
		"aud":       "api.example.com",
		"iss":       "https://example.com",
	}

	return s.createToken(claims)
}

func (s *TokenIntrospectionServiceTestSuite) createExpiredToken() string {
	claims := map[string]interface{}{
		"exp":       float64(time.Now().Add(-time.Hour).Unix()), // Expired
		"nbf":       float64(time.Now().Add(-time.Hour * 2).Unix()),
		"iat":       float64(time.Now().Add(-time.Hour * 3).Unix()),
		"jti":       "expired-token-123",
		"scope":     "openid profile",
		"client_id": "client123",
		"username":  "user@example.com",
		"sub":       "user123",
		"aud":       "api.example.com",
		"iss":       "https://example.com",
	}

	return s.createToken(claims)
}

func (s *TokenIntrospectionServiceTestSuite) createNotBeforeToken() string {
	claims := map[string]interface{}{
		"exp":       float64(time.Now().Add(time.Hour).Unix()),
		"nbf":       float64(time.Now().Add(time.Hour).Unix()), // Not valid yet
		"iat":       float64(time.Now().Unix()),
		"jti":       "future-token-123",
		"scope":     "openid profile",
		"client_id": "client123",
		"username":  "user@example.com",
		"sub":       "user123",
		"aud":       "api.example.com",
		"iss":       "https://example.com",
	}

	return s.createToken(claims)
}

func (s *TokenIntrospectionServiceTestSuite) createMissingClaimsToken() string {
	claims := map[string]interface{}{
		"exp": float64(time.Now().Add(time.Hour).Unix()),
		"nbf": float64(time.Now().Add(-time.Minute).Unix()),
	}

	return s.createToken(claims)
}
