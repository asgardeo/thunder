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
			name:                "Valid plain challenge",
			codeChallenge:       "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			codeChallengeMethod: CodeChallengeMethodPlain,
			codeVerifier:        "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			expectError:         false,
			expectedError:       nil,
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
			name:                "Invalid plain challenge",
			codeChallenge:       "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			codeChallengeMethod: CodeChallengeMethodPlain,
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
			name:                "Default method when empty",
			codeChallenge:       "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			codeChallengeMethod: "",
			codeVerifier:        "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			expectError:         false,
			expectedError:       nil,
		},
		{
			name:                "Unicode characters rejected",
			codeChallenge:       "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			codeChallengeMethod: CodeChallengeMethodPlain,
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
					assert.ErrorIs(t, err, tt.expectedError, "Expected specific error: %v, got: %v", tt.expectedError, err)
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
			name:          "Generate plain challenge",
			codeVerifier:  "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			method:        CodeChallengeMethodPlain,
			expectError:   false,
			expectedError: nil,
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
					assert.ErrorIs(t, err, tt.expectedError, "Expected specific error: %v, got: %v", tt.expectedError, err)
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
			name:                "Valid plain challenge",
			codeChallenge:       "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			codeChallengeMethod: CodeChallengeMethodPlain,
			expectError:         false,
			expectedError:       nil,
		},
		{
			name:                "Valid S256 challenge",
			codeChallenge:       "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
			codeChallengeMethod: CodeChallengeMethodS256,
			expectError:         false,
			expectedError:       nil,
		},
		{
			name:                "Empty code challenge",
			codeChallenge:       "",
			codeChallengeMethod: CodeChallengeMethodPlain,
			expectError:         true,
			expectedError:       ErrInvalidCodeChallenge,
		},
		{
			name:                "Invalid challenge method",
			codeChallenge:       "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			codeChallengeMethod: "invalid",
			expectError:         true,
			expectedError:       ErrInvalidCodeChallenge,
		},
		{
			name:                "Default method when empty",
			codeChallenge:       "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			codeChallengeMethod: "",
			expectError:         false,
			expectedError:       nil,
		},
		{
			name:                "Plain challenge with invalid characters",
			codeChallenge:       "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk!",
			codeChallengeMethod: CodeChallengeMethodPlain,
			expectError:         true,
			expectedError:       ErrInvalidCodeChallenge,
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
		{
			name:                "Plain challenge too short",
			codeChallenge:       "short",
			codeChallengeMethod: CodeChallengeMethodPlain,
			expectError:         true,
			expectedError:       ErrInvalidCodeChallenge,
		},
		{
			name:                "Unicode characters rejected",
			codeChallenge:       "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk中文",
			codeChallengeMethod: CodeChallengeMethodPlain,
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
					assert.ErrorIs(t, err, tt.expectedError, "Expected specific error: %v, got: %v", tt.expectedError, err)
				}
			} else {
				assert.NoError(t, err, "Expected no error but got: %v", err)
			}
		})
	}
}
