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

package pkce

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type PKCETestSuite struct {
	suite.Suite
}

func TestPKCESuite(t *testing.T) {
	suite.Run(t, new(PKCETestSuite))
}

func (suite *PKCETestSuite) TestValidatePKCE() {
	tests := []struct {
		name                string
		codeChallenge       string
		codeChallengeMethod string
		codeVerifier        string
		expectError         bool
		expectedError       error
	}{
		{
			name:                "Valid S256 challenge",
			codeChallenge:       "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
			codeChallengeMethod: CodeChallengeMethodS256,
			codeVerifier:        "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			expectError:         false,
			expectedError:       nil,
		},
		{
			name:                "Plain method is rejected",
			codeChallenge:       "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			codeChallengeMethod: "plain",
			codeVerifier:        "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			expectError:         true,
			expectedError:       ErrInvalidChallengeMethod,
		},
		{
			name:                "Invalid S256 challenge",
			codeChallenge:       "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			codeChallengeMethod: CodeChallengeMethodS256,
			codeVerifier:        "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk_different_verifier_long_enough",
			expectError:         true,
			expectedError:       ErrPKCEValidationFailed,
		},
		{
			name:                "Empty code verifier",
			codeChallenge:       "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			codeChallengeMethod: CodeChallengeMethodS256,
			codeVerifier:        "",
			expectError:         true,
			expectedError:       ErrInvalidCodeVerifier,
		},
		{
			name:                "Empty code challenge",
			codeChallenge:       "",
			codeChallengeMethod: CodeChallengeMethodS256,
			codeVerifier:        "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			expectError:         true,
			expectedError:       ErrInvalidCodeChallenge,
		},
		{
			name:                "Invalid challenge method",
			codeChallenge:       "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			codeChallengeMethod: "invalid",
			codeVerifier:        "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			expectError:         true,
			expectedError:       ErrInvalidChallengeMethod,
		},
		{
			name:                "Code verifier too short",
			codeChallenge:       "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			codeChallengeMethod: CodeChallengeMethodS256,
			codeVerifier:        "short",
			expectError:         true,
			expectedError:       ErrInvalidCodeVerifier,
		},
		{
			name:                "Default method when empty uses S256",
			codeChallenge:       "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
			codeChallengeMethod: "",
			codeVerifier:        "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			expectError:         false,
			expectedError:       nil,
		},
		{
			name:                "Unicode characters rejected",
			codeChallenge:       "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
			codeChallengeMethod: CodeChallengeMethodS256,
			codeVerifier:        "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk中文",
			expectError:         true,
			expectedError:       ErrInvalidCodeVerifier,
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			err := ValidatePKCE(tt.codeChallenge, tt.codeChallengeMethod, tt.codeVerifier)

			if tt.expectError {
				assert.Error(t, err, "Expected error but got none")
				if tt.expectedError != nil {
					assert.ErrorIs(t, err, tt.expectedError,
						"Expected specific error: %v, got: %v", tt.expectedError, err)
				}
			} else {
				assert.NoError(t, err, "Expected no error but got: %v", err)
			}
		})
	}
}

func (suite *PKCETestSuite) TestGenerateCodeChallenge() {
	tests := []struct {
		name          string
		codeVerifier  string
		method        string
		expectError   bool
		expectedError error
	}{
		{
			name:          "Generate S256 challenge",
			codeVerifier:  "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			method:        CodeChallengeMethodS256,
			expectError:   false,
			expectedError: nil,
		},
		{
			name:          "Plain method is rejected",
			codeVerifier:  "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			method:        "plain",
			expectError:   true,
			expectedError: ErrInvalidChallengeMethod,
		},
		{
			name:          "Invalid method",
			codeVerifier:  "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			method:        "invalid",
			expectError:   true,
			expectedError: ErrInvalidChallengeMethod,
		},
		{
			name:          "Empty code verifier",
			codeVerifier:  "",
			method:        CodeChallengeMethodS256,
			expectError:   true,
			expectedError: ErrInvalidCodeVerifier,
		},
		{
			name:          "Code verifier too short",
			codeVerifier:  "short",
			method:        CodeChallengeMethodS256,
			expectError:   true,
			expectedError: ErrInvalidCodeVerifier,
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			challenge, err := GenerateCodeChallenge(tt.codeVerifier, tt.method)

			if tt.expectError {
				assert.Error(t, err, "Expected error but got none")
				assert.Empty(t, challenge, "Challenge should be empty on error")
				if tt.expectedError != nil {
					assert.ErrorIs(t, err, tt.expectedError,
						"Expected specific error: %v, got: %v", tt.expectedError, err)
				}
			} else {
				assert.NoError(t, err, "Expected no error but got: %v", err)
				assert.NotEmpty(t, challenge, "Challenge should not be empty")

				err = ValidatePKCE(challenge, tt.method, tt.codeVerifier)
				assert.NoError(t, err, "Generated challenge validation failed: %v", err)
			}
		})
	}
}

func (suite *PKCETestSuite) TestValidateCodeChallenge() {
	tests := []struct {
		name                string
		codeChallenge       string
		codeChallengeMethod string
		expectError         bool
		expectedError       error
	}{
		{
			name:                "Plain method is rejected",
			codeChallenge:       "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			codeChallengeMethod: "plain",
			expectError:         true,
			expectedError:       ErrInvalidChallengeMethod,
		},
		{
			name:                "Valid S256 challenge",
			codeChallenge:       "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
			codeChallengeMethod: CodeChallengeMethodS256,
			expectError:         false,
			expectedError:       nil,
		},
		{
			name:                "Empty code challenge with S256",
			codeChallenge:       "",
			codeChallengeMethod: CodeChallengeMethodS256,
			expectError:         true,
			expectedError:       ErrInvalidCodeChallenge,
		},
		{
			name:                "Invalid challenge method",
			codeChallenge:       "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			codeChallengeMethod: "invalid",
			expectError:         true,
			expectedError:       ErrInvalidChallengeMethod,
		},
		{
			name:                "Default method when empty uses S256",
			codeChallenge:       "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
			codeChallengeMethod: "",
			expectError:         false,
			expectedError:       nil,
		},
		{
			name:                "S256 challenge with invalid characters",
			codeChallenge:       "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM!",
			codeChallengeMethod: CodeChallengeMethodS256,
			expectError:         true,
			expectedError:       ErrInvalidCodeChallenge,
		},
		{
			name:                "S256 challenge wrong length",
			codeChallenge:       "short",
			codeChallengeMethod: CodeChallengeMethodS256,
			expectError:         true,
			expectedError:       ErrInvalidCodeChallenge,
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			err := ValidateCodeChallenge(tt.codeChallenge, tt.codeChallengeMethod)

			if tt.expectError {
				assert.Error(t, err, "Expected error but got none")
				if tt.expectedError != nil {
					assert.ErrorIs(t, err, tt.expectedError,
						"Expected specific error: %v, got: %v", tt.expectedError, err)
				}
			} else {
				assert.NoError(t, err, "Expected no error but got: %v", err)
			}
		})
	}
}

func (suite *PKCETestSuite) TestValidateCodeChallenge_InvalidMethod() {
	// Test that unsupported methods are rejected
	err := ValidateCodeChallenge("valid-challenge", "unsupported_method")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrInvalidChallengeMethod, err)
}

func (suite *PKCETestSuite) TestGenerateCodeChallenge_InvalidMethod() {
	// Test the default case in generateCodeChallenge
	challenge, err := GenerateCodeChallenge("dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk", "unsupported_method")
	assert.Error(suite.T(), err)
	assert.Empty(suite.T(), challenge)
	assert.Equal(suite.T(), ErrInvalidChallengeMethod, err)
}

func (suite *PKCETestSuite) TestValidateCodeChallenge_S256InvalidCharacters() {
	// Test S256 challenge with invalid base64URL characters
	// S256 challenge must be exactly 43 characters and only base64URL characters
	invalidChallenges := []string{
		"E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM!", // Invalid character
		"E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM ", // Space
		"E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM+", // Plus sign (not base64URL)
		"E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM/", // Slash (not base64URL)
	}

	for _, challenge := range invalidChallenges {
		suite.T().Run("Invalid_"+challenge, func(t *testing.T) {
			err := ValidateCodeChallenge(challenge, CodeChallengeMethodS256)
			assert.Error(t, err)
			assert.Equal(t, ErrInvalidCodeChallenge, err)
		})
	}
}

func (suite *PKCETestSuite) TestGetSupportedCodeChallengeMethods() {
	// Test GetSupportedCodeChallengeMethods function
	methods := GetSupportedCodeChallengeMethods()

	assert.NotNil(suite.T(), methods)
	assert.Equal(suite.T(), 1, len(methods))
	assert.Contains(suite.T(), methods, CodeChallengeMethodS256)
	assert.NotContains(suite.T(), methods, "plain")
}
