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

package utils

import (
	"encoding/base64"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type OAuthUtilTestSuite struct {
	suite.Suite
}

func TestOAuthUtilSuite(t *testing.T) {
	suite.Run(t, new(OAuthUtilTestSuite))
}

func (suite *OAuthUtilTestSuite) TestGenerateOAuth2ClientID() {
	clientID, err := GenerateOAuth2ClientID()
	
	// Should not return an error
	assert.NoError(suite.T(), err, "GenerateOAuth2ClientID should not return an error")
	assert.NotEmpty(suite.T(), clientID, "Generated client ID should not be empty")

	// Verify format - should be base64url without padding
	base64URLPattern := regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	assert.True(suite.T(), base64URLPattern.MatchString(clientID), 
		"Client ID should contain only base64url characters (A-Z, a-z, 0-9, -, _)")

	// Should not contain padding characters
	assert.False(suite.T(), strings.Contains(clientID, "="), 
		"Client ID should not contain padding characters")

	// Verify length - 16 bytes base64url encoded without padding should be ~22 characters
	expectedLength := base64.RawURLEncoding.EncodedLen(OAuth2ClientIDLength)
	assert.Equal(suite.T(), expectedLength, len(clientID), 
		"Client ID should have the expected encoded length")

	// Verify it can be decoded back to original byte length
	decoded, err := base64.RawURLEncoding.DecodeString(clientID)
	assert.NoError(suite.T(), err, "Generated client ID should be valid base64url")
	assert.Equal(suite.T(), OAuth2ClientIDLength, len(decoded), 
		"Decoded client ID should have the expected byte length")
}

func (suite *OAuthUtilTestSuite) TestGenerateOAuth2ClientSecret() {
	clientSecret, err := GenerateOAuth2ClientSecret()
	
	// Should not return an error
	assert.NoError(suite.T(), err, "GenerateOAuth2ClientSecret should not return an error")
	assert.NotEmpty(suite.T(), clientSecret, "Generated client secret should not be empty")

	// Verify format - should be base64url without padding
	base64URLPattern := regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	assert.True(suite.T(), base64URLPattern.MatchString(clientSecret), 
		"Client secret should contain only base64url characters (A-Z, a-z, 0-9, -, _)")

	// Should not contain padding characters
	assert.False(suite.T(), strings.Contains(clientSecret, "="), 
		"Client secret should not contain padding characters")

	// Verify length - 32 bytes base64url encoded without padding should be ~43 characters
	expectedLength := base64.RawURLEncoding.EncodedLen(OAuth2ClientSecretLength)
	assert.Equal(suite.T(), expectedLength, len(clientSecret), 
		"Client secret should have the expected encoded length")

	// Verify it can be decoded back to original byte length
	decoded, err := base64.RawURLEncoding.DecodeString(clientSecret)
	assert.NoError(suite.T(), err, "Generated client secret should be valid base64url")
	assert.Equal(suite.T(), OAuth2ClientSecretLength, len(decoded), 
		"Decoded client secret should have the expected byte length")
}

func (suite *OAuthUtilTestSuite) TestGenerateOAuth2ClientIDUniqueness() {
	clientIDs := make(map[string]bool)
	
	// Generate multiple client IDs and verify uniqueness
	for i := 0; i < 1000; i++ {
		clientID, err := GenerateOAuth2ClientID()
		assert.NoError(suite.T(), err, "Should not return an error during generation")
		
		_, exists := clientIDs[clientID]
		assert.False(suite.T(), exists, "Generated client IDs should be unique")
		clientIDs[clientID] = true
	}

	assert.Equal(suite.T(), 1000, len(clientIDs), "Should have generated 1000 unique client IDs")
}

func (suite *OAuthUtilTestSuite) TestGenerateOAuth2ClientSecretUniqueness() {
	clientSecrets := make(map[string]bool)
	
	// Generate multiple client secrets and verify uniqueness
	for i := 0; i < 1000; i++ {
		clientSecret, err := GenerateOAuth2ClientSecret()
		assert.NoError(suite.T(), err, "Should not return an error during generation")
		
		_, exists := clientSecrets[clientSecret]
		assert.False(suite.T(), exists, "Generated client secrets should be unique")
		clientSecrets[clientSecret] = true
	}

	assert.Equal(suite.T(), 1000, len(clientSecrets), "Should have generated 1000 unique client secrets")
}

func (suite *OAuthUtilTestSuite) TestOAuth2CredentialsDifferentFromUUID() {
	// Generate OAuth credentials
	clientID, err := GenerateOAuth2ClientID()
	assert.NoError(suite.T(), err)
	
	clientSecret, err := GenerateOAuth2ClientSecret()
	assert.NoError(suite.T(), err)
	
	// Generate UUID for comparison
	uuid := GenerateUUID()
	
	// OAuth credentials should have different format than UUID
	uuidPattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	
	assert.False(suite.T(), uuidPattern.MatchString(clientID), 
		"OAuth client ID should not match UUID format")
	assert.False(suite.T(), uuidPattern.MatchString(clientSecret), 
		"OAuth client secret should not match UUID format")
	assert.True(suite.T(), uuidPattern.MatchString(uuid), 
		"UUID should match expected UUID format")
	
	// OAuth credentials should be shorter/different than UUID
	assert.True(suite.T(), len(clientID) < len(uuid), 
		"OAuth client ID should be shorter than UUID")
	assert.True(suite.T(), len(clientSecret) > len(uuid), 
		"OAuth client secret should be longer than UUID for better security")
}

func (suite *OAuthUtilTestSuite) TestOAuth2URLSafety() {
	// Generate credentials multiple times to test URL safety
	for i := 0; i < 100; i++ {
		clientID, err := GenerateOAuth2ClientID()
		assert.NoError(suite.T(), err)
		
		clientSecret, err := GenerateOAuth2ClientSecret()
		assert.NoError(suite.T(), err)
		
		// Should not contain URL-unsafe characters
		urlUnsafeChars := []string{"+", "/", "=", " ", "&", "?", "#"}
		for _, char := range urlUnsafeChars {
			assert.False(suite.T(), strings.Contains(clientID, char), 
				"Client ID should not contain URL-unsafe character: %s", char)
			assert.False(suite.T(), strings.Contains(clientSecret, char), 
				"Client secret should not contain URL-unsafe character: %s", char)
		}
	}
}

func (suite *OAuthUtilTestSuite) TestOAuth2EntropyLevels() {
	clientID, err := GenerateOAuth2ClientID()
	assert.NoError(suite.T(), err)
	
	clientSecret, err := GenerateOAuth2ClientSecret()
	assert.NoError(suite.T(), err)
	
	// Decode to verify entropy
	clientIDBytes, err := base64.RawURLEncoding.DecodeString(clientID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), OAuth2ClientIDLength, len(clientIDBytes), 
		"Client ID should have 16 bytes (128 bits) of entropy")
	
	clientSecretBytes, err := base64.RawURLEncoding.DecodeString(clientSecret)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), OAuth2ClientSecretLength, len(clientSecretBytes), 
		"Client secret should have 32 bytes (256 bits) of entropy")
	
	// Client secret should have more entropy than client ID
	assert.True(suite.T(), len(clientSecretBytes) > len(clientIDBytes), 
		"Client secret should have more entropy than client ID")
}