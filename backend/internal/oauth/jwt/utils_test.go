/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
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
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/config"
)

type JWTUtilsTestSuite struct {
	suite.Suite
}

func TestJWTUtilsSuite(t *testing.T) {
	suite.Run(t, new(JWTUtilsTestSuite))
}

func (suite *JWTUtilsTestSuite) SetupTest() {
	// Initialize runtime configuration for testing
	testConfig := &config.Config{
		OAuth: config.OAuthConfig{
			JWT: config.JWTConfig{
				ValidityPeriod: 7200,
			},
		},
	}
	err := config.InitializeThunderRuntime("/tmp", testConfig)
	assert.NoError(suite.T(), err)
}

func (suite *JWTUtilsTestSuite) TestGetJWTTokenValidityPeriod() {
	validityPeriod := GetJWTTokenValidityPeriod()
	assert.Equal(suite.T(), int64(7200), validityPeriod)
}

func (suite *JWTUtilsTestSuite) TestGetJWTTokenValidityPeriod_DefaultValue() {
	// Test with config having 0 validity period
	testConfig := &config.Config{
		OAuth: config.OAuthConfig{
			JWT: config.JWTConfig{
				ValidityPeriod: 0,
			},
		},
	}
	err := config.InitializeThunderRuntime("/tmp", testConfig)
	assert.NoError(suite.T(), err)

	validityPeriod := GetJWTTokenValidityPeriod()
	assert.Equal(suite.T(), int64(defaultTokenValidity), validityPeriod)
}

func (suite *JWTUtilsTestSuite) TestDecodeJWT_Success() {
	// Create a test JWT manually
	header := map[string]interface{}{
		"alg": "RS256",
		"typ": "JWT",
		"kid": "test-key-id",
	}
	payload := map[string]interface{}{
		"sub":   "test-subject",
		"aud":   "test-audience",
		"iss":   "test-issuer",
		"exp":   1234567890,
		"iat":   1234567800,
		"scope": "read write",
	}

	headerBytes, _ := json.Marshal(header)
	payloadBytes, _ := json.Marshal(payload)

	headerB64 := base64.RawURLEncoding.EncodeToString(headerBytes)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadBytes)
	signature := "fake-signature"

	token := headerB64 + "." + payloadB64 + "." + signature

	decodedHeader, decodedPayload, err := DecodeJWT(token)
	
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "RS256", decodedHeader["alg"])
	assert.Equal(suite.T(), "JWT", decodedHeader["typ"])
	assert.Equal(suite.T(), "test-key-id", decodedHeader["kid"])
	assert.Equal(suite.T(), "test-subject", decodedPayload["sub"])
	assert.Equal(suite.T(), "test-audience", decodedPayload["aud"])
	assert.Equal(suite.T(), "test-issuer", decodedPayload["iss"])
	assert.Equal(suite.T(), "read write", decodedPayload["scope"])
}

func (suite *JWTUtilsTestSuite) TestDecodeJWT_InvalidFormat() {
	testCases := []struct {
		name  string
		token string
	}{
		{
			name:  "TwoParts",
			token: "header.payload",
		},
		{
			name:  "FourParts",
			token: "header.payload.signature.extra",
		},
		{
			name:  "Empty",
			token: "",
		},
		{
			name:  "SinglePart",
			token: "header",
		},
		{
			name:  "OnlyDots",
			token: "..",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			_, _, err := DecodeJWT(tc.token)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid JWT format")
		})
	}
}

func (suite *JWTUtilsTestSuite) TestDecodeJWT_InvalidBase64Header() {
	// Create invalid base64 header but valid payload
	payloadBytes, _ := json.Marshal(map[string]interface{}{"sub": "test"})
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadBytes)
	
	token := "invalid-base64." + payloadB64 + ".signature"
	_, _, err := DecodeJWT(token)
	
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to decode JWT header")
}

func (suite *JWTUtilsTestSuite) TestDecodeJWT_InvalidBase64Payload() {
	headerBytes, _ := json.Marshal(map[string]interface{}{"alg": "RS256"})
	headerB64 := base64.RawURLEncoding.EncodeToString(headerBytes)
	
	token := headerB64 + ".invalid-base64.signature"
	_, _, err := DecodeJWT(token)
	
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to decode JWT payload")
}

func (suite *JWTUtilsTestSuite) TestDecodeJWT_InvalidJSONHeader() {
	// Create base64 encoded invalid JSON for header
	headerB64 := base64.RawURLEncoding.EncodeToString([]byte("invalid json"))
	payloadBytes, _ := json.Marshal(map[string]interface{}{"sub": "test"})
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadBytes)
	
	token := headerB64 + "." + payloadB64 + ".signature"
	_, _, err := DecodeJWT(token)
	
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to unmarshal JWT header")
}

func (suite *JWTUtilsTestSuite) TestDecodeJWT_InvalidJSONPayload() {
	headerBytes, _ := json.Marshal(map[string]interface{}{"alg": "RS256"})
	headerB64 := base64.RawURLEncoding.EncodeToString(headerBytes)
	// Create base64 encoded invalid JSON for payload
	payloadB64 := base64.RawURLEncoding.EncodeToString([]byte("invalid json"))
	
	token := headerB64 + "." + payloadB64 + ".signature"
	_, _, err := DecodeJWT(token)
	
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to unmarshal JWT payload")
}

func (suite *JWTUtilsTestSuite) TestDecodeJWT_EmptyJSONObjects() {
	// Test with empty but valid JSON objects
	headerBytes, _ := json.Marshal(map[string]interface{}{})
	payloadBytes, _ := json.Marshal(map[string]interface{}{})

	headerB64 := base64.RawURLEncoding.EncodeToString(headerBytes)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadBytes)

	token := headerB64 + "." + payloadB64 + ".signature"
	decodedHeader, decodedPayload, err := DecodeJWT(token)
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), decodedHeader)
	assert.NotNil(suite.T(), decodedPayload)
	assert.Len(suite.T(), decodedHeader, 0)
	assert.Len(suite.T(), decodedPayload, 0)
}

func (suite *JWTUtilsTestSuite) TestDecodeJWT_WithSpecialCharacters() {
	// Test with payload containing special characters and various data types
	header := map[string]interface{}{
		"alg": "RS256",
		"typ": "JWT",
	}
	payload := map[string]interface{}{
		"sub":     "user@example.com",
		"roles":   []string{"admin", "user"},
		"isAdmin": true,
		"count":   42,
		"data":    map[string]interface{}{"nested": "value"},
		"special": "hello world!@#$%^&*()",
	}

	headerBytes, _ := json.Marshal(header)
	payloadBytes, _ := json.Marshal(payload)

	headerB64 := base64.RawURLEncoding.EncodeToString(headerBytes)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadBytes)

	token := headerB64 + "." + payloadB64 + ".signature"
	decodedHeader, decodedPayload, err := DecodeJWT(token)
	
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "RS256", decodedHeader["alg"])
	assert.Equal(suite.T(), "user@example.com", decodedPayload["sub"])
	assert.Equal(suite.T(), true, decodedPayload["isAdmin"])
	assert.Equal(suite.T(), "hello world!@#$%^&*()", decodedPayload["special"])
	
	// Check array handling
	roles, ok := decodedPayload["roles"].([]interface{})
	assert.True(suite.T(), ok)
	assert.Len(suite.T(), roles, 2)
	assert.Equal(suite.T(), "admin", roles[0])
	assert.Equal(suite.T(), "user", roles[1])
}

func (suite *JWTUtilsTestSuite) TestDecodeJWT_RealWorldExample() {
	// Test with a more realistic JWT structure
	header := map[string]interface{}{
		"alg": "RS256",
		"typ": "JWT",
		"kid": "2011-04-29",
	}
	payload := map[string]interface{}{
		"iss":   "https://server.example.com",
		"sub":   "248289761001",
		"aud":   "s6BhdRkqt3",
		"nonce": "n-0S6_WzA2Mj",
		"exp":   1311281970,
		"iat":   1311280970,
		"name":  "Jane Doe",
		"email": "jane.doe@example.com",
		"scope": "openid profile email",
		"custom_claims": map[string]interface{}{
			"department": "Engineering",
			"level":      "Senior",
		},
	}

	headerBytes, _ := json.Marshal(header)
	payloadBytes, _ := json.Marshal(payload)

	headerB64 := base64.RawURLEncoding.EncodeToString(headerBytes)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadBytes)

	token := headerB64 + "." + payloadB64 + ".signature"
	decodedHeader, decodedPayload, err := DecodeJWT(token)
	
	assert.NoError(suite.T(), err)
	
	// Verify header
	assert.Equal(suite.T(), "RS256", decodedHeader["alg"])
	assert.Equal(suite.T(), "JWT", decodedHeader["typ"])
	assert.Equal(suite.T(), "2011-04-29", decodedHeader["kid"])
	
	// Verify payload
	assert.Equal(suite.T(), "https://server.example.com", decodedPayload["iss"])
	assert.Equal(suite.T(), "248289761001", decodedPayload["sub"])
	assert.Equal(suite.T(), "s6BhdRkqt3", decodedPayload["aud"])
	assert.Equal(suite.T(), "Jane Doe", decodedPayload["name"])
	assert.Equal(suite.T(), "jane.doe@example.com", decodedPayload["email"])
	assert.Equal(suite.T(), "openid profile email", decodedPayload["scope"])
	
	// Verify nested custom claims
	customClaims, ok := decodedPayload["custom_claims"].(map[string]interface{})
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "Engineering", customClaims["department"])
	assert.Equal(suite.T(), "Senior", customClaims["level"])
}
