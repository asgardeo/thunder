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

package common

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/idp"
)

type AuthenticatorTestSuite struct {
	suite.Suite
}

func TestAuthenticatorTestSuite(t *testing.T) {
	suite.Run(t, new(AuthenticatorTestSuite))
}

func (suite *AuthenticatorTestSuite) TestGetAuthenticatorMetaData() {
	testCases := []struct {
		name              string
		authenticator     string
		expectNil         bool
		expectedName      string
		expectedAALWeight int
	}{
		{
			name:              "Credentials authenticator",
			authenticator:     AuthenticatorCredentials,
			expectNil:         false,
			expectedName:      AuthenticatorCredentials,
			expectedAALWeight: 1,
		},
		{
			name:              "SMS OTP authenticator",
			authenticator:     AuthenticatorSMSOTP,
			expectNil:         false,
			expectedName:      AuthenticatorSMSOTP,
			expectedAALWeight: 1,
		},
		{
			name:              "Google authenticator",
			authenticator:     AuthenticatorGoogle,
			expectNil:         false,
			expectedName:      AuthenticatorGoogle,
			expectedAALWeight: 1,
		},
		{
			name:              "GitHub authenticator",
			authenticator:     AuthenticatorGithub,
			expectNil:         false,
			expectedName:      AuthenticatorGithub,
			expectedAALWeight: 1,
		},
		{
			name:              "OAuth authenticator",
			authenticator:     AuthenticatorOAuth,
			expectNil:         false,
			expectedName:      AuthenticatorOAuth,
			expectedAALWeight: 1,
		},
		{
			name:              "OIDC authenticator",
			authenticator:     AuthenticatorOIDC,
			expectNil:         false,
			expectedName:      AuthenticatorOIDC,
			expectedAALWeight: 1,
		},
		{
			name:          "Unknown authenticator",
			authenticator: "UnknownAuthenticator",
			expectNil:     true,
		},
		{
			name:          "Empty authenticator name",
			authenticator: "",
			expectNil:     true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			result := getAuthenticatorMetaData(tc.authenticator)

			if tc.expectNil {
				suite.Nil(result)
			} else {
				suite.NotNil(result)
				suite.Equal(tc.expectedName, result.Name)
				suite.Equal(tc.expectedAALWeight, result.AALWeight)
			}
		})
	}
}

func (suite *AuthenticatorTestSuite) TestGetAuthenticatorWeight() {
	testCases := []struct {
		name           string
		authenticator  string
		expectedWeight int
	}{
		{
			name:           "Credentials authenticator",
			authenticator:  AuthenticatorCredentials,
			expectedWeight: 1,
		},
		{
			name:           "SMS OTP authenticator",
			authenticator:  AuthenticatorSMSOTP,
			expectedWeight: 1,
		},
		{
			name:           "Google authenticator",
			authenticator:  AuthenticatorGoogle,
			expectedWeight: 1,
		},
		{
			name:           "GitHub authenticator",
			authenticator:  AuthenticatorGithub,
			expectedWeight: 1,
		},
		{
			name:           "OAuth authenticator",
			authenticator:  AuthenticatorOAuth,
			expectedWeight: 1,
		},
		{
			name:           "OIDC authenticator",
			authenticator:  AuthenticatorOIDC,
			expectedWeight: 1,
		},
		{
			name:           "Unknown authenticator returns 0",
			authenticator:  "UnknownAuthenticator",
			expectedWeight: 0,
		},
		{
			name:           "Empty authenticator name returns 0",
			authenticator:  "",
			expectedWeight: 0,
		},
		{
			name:           "Random string returns 0",
			authenticator:  "RandomString123",
			expectedWeight: 0,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			result := GetAuthenticatorWeight(tc.authenticator)
			suite.Equal(tc.expectedWeight, result)
		})
	}
}

func (suite *AuthenticatorTestSuite) TestGetAuthenticatorNameForIDPType() {
	testCases := []struct {
		name             string
		idpType          idp.IDPType
		expectedAuthName string
	}{
		{
			name:             "Google IDP type",
			idpType:          idp.IDPTypeGoogle,
			expectedAuthName: AuthenticatorGoogle,
		},
		{
			name:             "GitHub IDP type",
			idpType:          idp.IDPTypeGitHub,
			expectedAuthName: AuthenticatorGithub,
		},
		{
			name:             "OAuth IDP type",
			idpType:          idp.IDPTypeOAuth,
			expectedAuthName: AuthenticatorOAuth,
		},
		{
			name:             "OIDC IDP type",
			idpType:          idp.IDPTypeOIDC,
			expectedAuthName: AuthenticatorOIDC,
		},
		{
			name:             "Unknown IDP type defaults to OAuth",
			idpType:          "UnknownIDPType",
			expectedAuthName: AuthenticatorOAuth,
		},
		{
			name:             "Empty IDP type defaults to OAuth",
			idpType:          "",
			expectedAuthName: AuthenticatorOAuth,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			result := GetAuthenticatorNameForIDPType(tc.idpType)
			suite.Equal(tc.expectedAuthName, result)
		})
	}
}

func (suite *AuthenticatorTestSuite) TestAuthenticatorRegistry() {
	suite.Run("Registry contains all expected authenticators", func() {
		expectedAuthenticators := []string{
			AuthenticatorCredentials,
			AuthenticatorSMSOTP,
			AuthenticatorGoogle,
			AuthenticatorGithub,
			AuthenticatorOAuth,
			AuthenticatorOIDC,
		}

		suite.Equal(len(expectedAuthenticators), len(authenticatorRegistry))

		for _, auth := range expectedAuthenticators {
			meta, exists := authenticatorRegistry[auth]
			suite.True(exists, "Authenticator %s should exist in registry", auth)
			suite.Equal(auth, meta.Name)
			suite.Equal(1, meta.AALWeight)
		}
	})
}

func (suite *AuthenticatorTestSuite) TestAuthenticatorMetaStructure() {
	suite.Run("AuthenticatorMeta has correct fields", func() {
		meta := AuthenticatorMeta{
			Name:      "TestAuthenticator",
			AALWeight: 2,
		}

		suite.Equal("TestAuthenticator", meta.Name)
		suite.Equal(2, meta.AALWeight)
	})
}

func (suite *AuthenticatorTestSuite) TestAuthenticatorReferenceStructure() {
	suite.Run("AuthenticatorReference has correct fields", func() {
		ref := AuthenticatorReference{
			Authenticator: AuthenticatorCredentials,
			Step:          1,
		}

		suite.Equal(AuthenticatorCredentials, ref.Authenticator)
		suite.Equal(1, ref.Step)
		suite.NotNil(ref.Timestamp)
	})
}

func (suite *AuthenticatorTestSuite) TestAuthenticatorConstants() {
	testCases := []struct {
		name     string
		constant string
		expected string
	}{
		{
			name:     "Credentials authenticator constant",
			constant: AuthenticatorCredentials,
			expected: "CredentialsAuthenticator",
		},
		{
			name:     "SMS OTP authenticator constant",
			constant: AuthenticatorSMSOTP,
			expected: "SMSOTPAuthenticator",
		},
		{
			name:     "Google authenticator constant",
			constant: AuthenticatorGoogle,
			expected: "GoogleOIDCAuthenticator",
		},
		{
			name:     "GitHub authenticator constant",
			constant: AuthenticatorGithub,
			expected: "GithubOAuthAuthenticator",
		},
		{
			name:     "OAuth authenticator constant",
			constant: AuthenticatorOAuth,
			expected: "OAuthAuthenticator",
		},
		{
			name:     "OIDC authenticator constant",
			constant: AuthenticatorOIDC,
			expected: "OIDCAuthenticator",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.Equal(tc.expected, tc.constant)
		})
	}
}

func (suite *AuthenticatorTestSuite) TestRegistryManipulation() {
	suite.Run("Directly modify and restore registry for testing", func() {
		// Save original registry
		originalRegistry := make(map[string]AuthenticatorMeta, len(authenticatorRegistry))
		for k, v := range authenticatorRegistry {
			originalRegistry[k] = v
		}
		suite.Equal(6, len(originalRegistry))

		// Add a test authenticator directly
		authenticatorRegistry["TestAuthenticator"] = AuthenticatorMeta{
			Name:      "TestAuthenticator",
			AALWeight: 2,
		}

		// Verify it was added
		weight := GetAuthenticatorWeight("TestAuthenticator")
		suite.Equal(2, weight)

		// Restore original registry
		authenticatorRegistry = originalRegistry

		// Verify test authenticator is removed
		weight = GetAuthenticatorWeight("TestAuthenticator")
		suite.Equal(0, weight)

		// Verify original authenticators still exist
		suite.Equal(1, GetAuthenticatorWeight(AuthenticatorCredentials))
	})
}
