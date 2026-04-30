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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ServerUtilTestSuite struct {
	suite.Suite
}

func TestServerUtilSuite(t *testing.T) {
	suite.Run(t, new(ServerUtilTestSuite))
}

func (suite *ServerUtilTestSuite) TestGetAllowedOrigin() {
	testCases := []struct {
		name           string
		allowedOrigins []string
		redirectURI    string
		expected       string
	}{
		{
			name:           "EmptyAllowedOrigins",
			allowedOrigins: []string{},
			redirectURI:    "https://example.com/callback",
			expected:       "",
		},
		{
			name:           "ExactMatch",
			allowedOrigins: []string{"https://example.com"},
			redirectURI:    "https://example.com/callback",
			expected:       "https://example.com",
		},
		{
			name:           "NoMatch",
			allowedOrigins: []string{"https://example.com"},
			redirectURI:    "https://malicious.com/callback",
			expected:       "",
		},
		{
			name:           "MultipleAllowedOriginsWithMatch",
			allowedOrigins: []string{"https://example1.com", "https://example2.com", "https://example3.com"},
			redirectURI:    "https://example2.com/auth/callback",
			expected:       "https://example2.com",
		},
		{
			name:           "SubdomainMatch",
			allowedOrigins: []string{"example.com"},
			redirectURI:    "https://subdomain.example.com/callback",
			expected:       "example.com",
		},
		{
			name:           "PartialStringMatch",
			allowedOrigins: []string{"example"},
			redirectURI:    "https://example.com/callback",
			expected:       "example",
		},
		{
			name:           "NullRedirectURI",
			allowedOrigins: []string{"https://example.com"},
			redirectURI:    "",
			expected:       "",
		},
		{
			name:           "CaseSensitiveNoMatch",
			allowedOrigins: []string{"https://EXAMPLE.com"},
			redirectURI:    "https://example.com/callback",
			expected:       "",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			result := GetAllowedOrigin(tc.allowedOrigins, tc.redirectURI)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func (suite *ServerUtilTestSuite) TestGetAllowedOriginWithProtocolVariations() {
	allowedOrigins := []string{"https://example.com"}
	testCases := []struct {
		name        string
		redirectURI string
		expected    string
	}{
		{
			name:        "HTTPSProtocol",
			redirectURI: "https://example.com/callback",
			expected:    "https://example.com",
		},
		{
			name:        "HTTPProtocol",
			redirectURI: "http://example.com/callback",
			expected:    "", // No match because protocol is different
		},
		{
			name:        "WithPortNumber",
			redirectURI: "https://example.com:8443/callback",
			expected:    "https://example.com",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			result := GetAllowedOrigin(allowedOrigins, tc.redirectURI)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func (suite *ServerUtilTestSuite) TestGetAllowedOriginWithComplexURLs() {
	testCases := []struct {
		name           string
		allowedOrigins []string
		redirectURI    string
		expected       string
	}{
		{
			name:           "URLWithQueryParameters",
			allowedOrigins: []string{"https://example.com"},
			redirectURI:    "https://example.com/callback?code=abc&state=123",
			expected:       "https://example.com",
		},
		{
			name:           "URLWithFragment",
			allowedOrigins: []string{"https://example.com"},
			redirectURI:    "https://example.com/callback#token=xyz",
			expected:       "https://example.com",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			result := GetAllowedOrigin(tc.allowedOrigins, tc.redirectURI)
			assert.Equal(t, tc.expected, result)
		})
	}
}
