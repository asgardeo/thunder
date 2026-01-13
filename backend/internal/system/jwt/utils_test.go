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

package jwt

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/crypto/sign"
)

type JWTUtilsTestSuite struct {
	suite.Suite
	rsaPrivateKey *rsa.PrivateKey
	rsaPublicKey  *rsa.PublicKey
	ecPrivateKey  *ecdsa.PrivateKey
	ecPublicKey   *ecdsa.PublicKey
	edPrivateKey  ed25519.PrivateKey
	edPublicKey   ed25519.PublicKey
	validJWT      string
	invalidJWT    string
	testServer    *httptest.Server
}

func TestJWTUtilsSuite(t *testing.T) {
	suite.Run(t, new(JWTUtilsTestSuite))
}

func (suite *JWTUtilsTestSuite) SetupTest() {
	err := config.InitializeThunderRuntime("", &config.Config{
		JWT: config.JWTConfig{},
	})
	assert.NoError(suite.T(), err)

	// Generate RSA key pair for testing
	suite.rsaPrivateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		suite.T().Fatalf("Failed to generate RSA key: %v", err)
	}
	suite.rsaPublicKey = &suite.rsaPrivateKey.PublicKey

	// Generate ECDSA P-256 key pair
	suite.ecPrivateKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		suite.T().Fatalf("Failed to generate ECDSA key: %v", err)
	}
	suite.ecPublicKey = &suite.ecPrivateKey.PublicKey

	// Generate Ed25519 key pair
	suite.edPublicKey, suite.edPrivateKey, err = ed25519.GenerateKey(rand.Reader)
	if err != nil {
		suite.T().Fatalf("Failed to generate Ed25519 key: %v", err)
	}

	// Create a valid JWT token
	suite.validJWT = suite.createValidJWT()

	// Create an invalid JWT token
	suite.invalidJWT = "invalid.jwt.token"
}

func (suite *JWTUtilsTestSuite) TearDownTest() {
	// Clean up the test server
	if suite.testServer != nil {
		suite.testServer.Close()
	}
}

// Helper method to create a valid JWT token for testing
func (suite *JWTUtilsTestSuite) createValidJWT() string {
	header := map[string]interface{}{
		"alg": "RS256",
		"typ": "JWT",
		"kid": "test-key-id",
	}

	payload := map[string]interface{}{
		"sub":  "1234567890",
		"name": "Test User",
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(time.Hour).Unix(),
	}

	headerJSON, _ := json.Marshal(header)
	payloadJSON, _ := json.Marshal(payload)

	headerBase64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadBase64 := base64.RawURLEncoding.EncodeToString(payloadJSON)
	signingInput := headerBase64 + "." + payloadBase64

	signature, err := sign.Generate([]byte(signingInput), sign.RSASHA256, suite.rsaPrivateKey)
	if err != nil {
		suite.T().Fatalf("Failed to sign JWT: %v", err)
	}
	signatureBase64 := base64.RawURLEncoding.EncodeToString(signature)

	return headerBase64 + "." + payloadBase64 + "." + signatureBase64
}

func (suite *JWTUtilsTestSuite) TestDecodeJWT() {
	tests := []struct {
		name            string
		token           string
		expectError     bool
		expectedHeader  map[string]interface{}
		expectedPayload map[string]interface{}
		errorContains   string
	}{
		{
			name: "WithValidToken",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." +
				"eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.signature",
			expectError:    false,
			expectedHeader: map[string]interface{}{"alg": "HS256", "typ": "JWT"},
			expectedPayload: map[string]interface{}{"sub": "1234567890", "name": "John Doe",
				"iat": float64(1516239022)},
		},
		{
			name:          "WithInvalidTokenFormat",
			token:         "part1.part2",
			expectError:   true,
			errorContains: "invalid JWT format",
		},
		{
			name:        "WithInvalidBase64InHeader",
			token:       "invalid_base64.eyJzdWIiOiIxMjM0NTY3ODkwIn0.signature",
			expectError: true,
		},
		{
			name:        "WithInvalidBase64InPayload",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid_base64.signature",
			expectError: true,
		},
		{
			name:        "WithInvalidJSONInHeader",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVH0.eyJzdWIiOiIxMjM0NTY3ODkwIn0.signature",
			expectError: true,
		},
		{
			name:        "WithInvalidJSONInPayload",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0c.signature",
			expectError: true,
		},
		{
			name:          "EmptyToken",
			token:         "",
			expectError:   true,
			errorContains: "invalid JWT format",
		},
	}

	for _, tc := range tests {
		suite.T().Run(tc.name, func(t *testing.T) {
			header, payload, err := DecodeJWT(tc.token)

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedHeader, header)
				assert.Equal(t, tc.expectedPayload, payload)
			}
		})
	}
}

func (suite *JWTUtilsTestSuite) TestParseJWTClaims() {
	claims, err := DecodeJWTPayload(suite.validJWT)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), claims)
	assert.Equal(suite.T(), "Test User", claims["name"])
	assert.Equal(suite.T(), "1234567890", claims["sub"])
}

func (suite *JWTUtilsTestSuite) TestParseJWTClaimsInvalid() {
	testCases := []struct {
		name  string
		token string
	}{
		{"InvalidFormat", "invalid.format"},
		{"EmptyToken", ""},
		{"MalformedPayload", "header.notbase64encoded.signature"},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			claims, err := DecodeJWTPayload(tc.token)

			assert.Error(t, err)
			assert.Nil(t, claims)
		})
	}
}

func (suite *JWTUtilsTestSuite) TestParseJWTHeader() {
	header, err := DecodeJWTHeader(suite.validJWT)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), header)
	assert.Equal(suite.T(), "RS256", header["alg"])
	assert.Equal(suite.T(), "JWT", header["typ"])
	assert.Equal(suite.T(), "test-key-id", header["kid"])
}

func (suite *JWTUtilsTestSuite) TestParseJWTHeaderInvalid() {
	header, err := DecodeJWTHeader(suite.invalidJWT)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), header)
}

func (suite *JWTUtilsTestSuite) TestJWKToPublicKey() {
	// Create a JWK from the RSA public key
	n := base64.RawURLEncoding.EncodeToString(suite.rsaPublicKey.N.Bytes())

	// Convert exponent to bytes
	eBytes := []byte{1, 0, 1} // 65537 in big-endian
	e := base64.RawURLEncoding.EncodeToString(eBytes)

	jwk := map[string]interface{}{
		"kty": "RSA",
		"n":   n,
		"e":   e,
		"kid": "test-key-id",
	}

	pubKey, err := jwkToPublicKey(jwk)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), pubKey)

	rsaKey, ok := pubKey.(*rsa.PublicKey)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), suite.rsaPublicKey.E, rsaKey.E)
}

func (suite *JWTUtilsTestSuite) TestJWKToPublicKeyInvalid() {
	// Create a JWK from the RSA public key
	n := base64.RawURLEncoding.EncodeToString(suite.rsaPublicKey.N.Bytes())

	// Convert exponent to bytes
	eBytes := []byte{1, 0, 1} // 65537 in big-endian
	e := base64.RawURLEncoding.EncodeToString(eBytes)

	// Test with missing parameters
	invalidTestCases := []struct {
		name string
		jwk  map[string]interface{}
	}{
		{"MissingN", map[string]interface{}{"e": e}},
		{"MissingE", map[string]interface{}{"n": n}},
		{"InvalidN", map[string]interface{}{"n": "invalid@base64!", "e": e}},
		{"InvalidE", map[string]interface{}{"n": n, "e": "invalid@base64!"}},
		{"NonStringN", map[string]interface{}{"n": 123, "e": e}},
		{"NonStringE", map[string]interface{}{"n": n, "e": 456}},
	}

	for _, tc := range invalidTestCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			pubKey, err := jwkToPublicKey(tc.jwk)

			assert.Error(t, err)
			assert.Nil(t, pubKey)
		})
	}
}

func (suite *JWTUtilsTestSuite) TestJWKToPublicKeyEC() {
	// Generate P-256 key
	x := base64.RawURLEncoding.EncodeToString(suite.ecPublicKey.X.Bytes())
	y := base64.RawURLEncoding.EncodeToString(suite.ecPublicKey.Y.Bytes())

	jwk := map[string]interface{}{
		"kty": "EC",
		"crv": "P-256",
		"x":   x,
		"y":   y,
		"kid": "ec-key-id",
	}

	pubKey, err := jwkToPublicKey(jwk)
	assert.NoError(suite.T(), err)

	ecPub, ok := pubKey.(*ecdsa.PublicKey)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "P-256", ecPub.Curve.Params().Name)
}

func (suite *JWTUtilsTestSuite) TestJWKToPublicKeyECInvalid() {
	// Test with missing parameters
	invalidTestCases := []struct {
		name string
		jwk  map[string]interface{}
	}{
		{"MissingX", map[string]interface{}{"crv": "P-256", "y": "valid"}},
		{"MissingY", map[string]interface{}{"crv": "P-256", "x": "valid"}},
		{"MissingCrv", map[string]interface{}{"x": "valid", "y": "valid"}},
		{"InvalidX", map[string]interface{}{"crv": "P-256", "x": "invalid@base64!", "y": "valid"}},
		{"InvalidY", map[string]interface{}{"crv": "P-256", "x": "valid", "y": "invalid@base64!"}},
		{"UnsupportedCrv", map[string]interface{}{"crv": "P-999", "x": "valid", "y": "valid"}},
	}

	for _, tc := range invalidTestCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			pubKey, err := jwkToPublicKey(tc.jwk)

			assert.Error(t, err)
			assert.Nil(t, pubKey)
		})
	}
}

func (suite *JWTUtilsTestSuite) TestJWKToPublicKeyEd25519() {
	x := base64.RawURLEncoding.EncodeToString(suite.edPublicKey)

	jwk := map[string]interface{}{
		"kty": "OKP",
		"crv": "Ed25519",
		"x":   x,
		"kid": "ed25519-key-id",
	}

	pubKey, err := jwkToPublicKey(jwk)
	assert.NoError(suite.T(), err)

	edPub, ok := pubKey.(ed25519.PublicKey)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), suite.edPublicKey, edPub)
}

func (suite *JWTUtilsTestSuite) TestJWKToPublicKeyEd25519Invalid() {
	// Test with missing parameters
	invalidTestCases := []struct {
		name string
		jwk  map[string]interface{}
	}{
		{"MissingX", map[string]interface{}{"crv": "Ed25519"}},
		{"MissingCrv", map[string]interface{}{"x": "valid"}},
		{"InvalidX", map[string]interface{}{"crv": "Ed25519", "x": "invalid@base64!"}},
		{"UnsupportedCrv", map[string]interface{}{"crv": "X25519", "x": "valid"}},
	}

	for _, tc := range invalidTestCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			pubKey, err := jwkToPublicKey(tc.jwk)

			assert.Error(t, err)
			assert.Nil(t, pubKey)
		})
	}
}
