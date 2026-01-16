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

package security

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/tests/mocks/jwtmock"
)

// JWTAuthenticatorTestSuite defines the test suite for JWTAuthenticator
type JWTAuthenticatorTestSuite struct {
	suite.Suite
	mockJWT       *jwtmock.JWTServiceInterfaceMock
	authenticator *jwtAuthenticator
}

func (suite *JWTAuthenticatorTestSuite) SetupTest() {
	suite.mockJWT = jwtmock.NewJWTServiceInterfaceMock(suite.T())
	suite.authenticator = newJWTAuthenticator(suite.mockJWT)
}

func (suite *JWTAuthenticatorTestSuite) TearDownTest() {
	suite.mockJWT.AssertExpectations(suite.T())
}

// Run the test suite
func TestJWTAuthenticatorSuite(t *testing.T) {
	suite.Run(t, new(JWTAuthenticatorTestSuite))
}

func (suite *JWTAuthenticatorTestSuite) TestCanHandle() {
	tests := []struct {
		name           string
		authHeader     string
		expectedResult bool
	}{
		{
			name:           "Valid Bearer token",
			authHeader:     "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.abc",
			expectedResult: true,
		},
		{
			name:           "No Authorization header",
			authHeader:     "",
			expectedResult: false,
		},
		{
			name:           "Basic auth header",
			authHeader:     "Basic dXNlcjpwYXNz",
			expectedResult: false,
		},
		{
			name:           "Bearer without token",
			authHeader:     "Bearer",
			expectedResult: false,
		},
		{
			name:           "Lowercase bearer",
			authHeader:     "bearer token123",
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			req := httptest.NewRequest(http.MethodGet, "/users", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			result := suite.authenticator.CanHandle(req)
			assert.Equal(suite.T(), tt.expectedResult, result)
		})
	}
}

func (suite *JWTAuthenticatorTestSuite) TestAuthenticate() {
	// Valid JWT token with attributes (simplified representation)
	// Payload: {"sub":"user123","scope":"system users:read","ou_id":"ou1","app_id":"app1"}
	//nolint:gosec,lll // Test data, not a real credential
	validToken := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyMTIzIiwic2NvcGUiOiJzeXN0ZW0gdXNlcnM6cmVhZCIsIm91X2lkIjoib3UxIiwiYXBwX2lkIjoiYXBwMSJ9.signature"

	tests := []struct {
		name           string
		authHeader     string
		setupMock      func(*jwtmock.JWTServiceInterfaceMock)
		expectedError  error
		validateResult func(*testing.T, *SecurityContext)
	}{
		{
			name:       "Successful authentication with system scope",
			authHeader: "Bearer " + validToken,
			setupMock: func(m *jwtmock.JWTServiceInterfaceMock) {
				m.On("VerifyJWTSignature", validToken).Return(nil)
			},
			expectedError: nil,
			validateResult: func(t *testing.T, ctx *SecurityContext) {
				baseCtx := withSecurityContext(context.Background(), ctx)
				assert.Equal(t, "user123", GetUserID(baseCtx))
				assert.Equal(t, "ou1", GetOUID(baseCtx))
				assert.Equal(t, "app1", GetAppID(baseCtx))
			},
		},
		{
			name:          "Missing Authorization header",
			authHeader:    "",
			setupMock:     func(m *jwtmock.JWTServiceInterfaceMock) {},
			expectedError: errMissingAuthHeader,
		},
		{
			name:          "Invalid header format",
			authHeader:    "Basic dXNlcjpwYXNz",
			setupMock:     func(m *jwtmock.JWTServiceInterfaceMock) {},
			expectedError: errMissingAuthHeader,
		},
		{
			name:          "Empty token",
			authHeader:    "Bearer   ",
			setupMock:     func(m *jwtmock.JWTServiceInterfaceMock) {},
			expectedError: errInvalidToken,
		},
		{
			name:       "Invalid JWT signature",
			authHeader: "Bearer invalid.jwt.token",
			setupMock: func(m *jwtmock.JWTServiceInterfaceMock) {
				m.On("VerifyJWTSignature", "invalid.jwt.token").Return(&serviceerror.ServiceError{
					Type:             serviceerror.ServerErrorType,
					Code:             "INVALID_SIGNATURE",
					Error:            "Invalid signature",
					ErrorDescription: "The JWT signature is invalid",
				})
			},
			expectedError: errInvalidToken,
		},
		{
			name:       "Invalid JWT format - decoding error",
			authHeader: "Bearer invalidjwtformat", // Not 3 parts separated by dots
			setupMock: func(m *jwtmock.JWTServiceInterfaceMock) {
				m.On("VerifyJWTSignature", "invalidjwtformat").Return(nil)
			},
			expectedError: errInvalidToken,
		},
		{
			name:       "Invalid JWT payload - malformed base64",
			authHeader: "Bearer eyJhbGciOiJIUzI1NiJ9.invalid!base64!payload.signature",
			setupMock: func(m *jwtmock.JWTServiceInterfaceMock) {
				m.On("VerifyJWTSignature", "eyJhbGciOiJIUzI1NiJ9.invalid!base64!payload.signature").Return(nil)
			},
			expectedError: errInvalidToken,
		},
		{
			name:       "Invalid JWT payload - malformed JSON",
			authHeader: "Bearer eyJhbGciOiJIUzI1NiJ9.bm90X3ZhbGlkX2pzb24.signature", // "not_valid_json" base64 encoded
			setupMock: func(m *jwtmock.JWTServiceInterfaceMock) {
				m.On("VerifyJWTSignature", "eyJhbGciOiJIUzI1NiJ9.bm90X3ZhbGlkX2pzb24.signature").Return(nil)
			},
			expectedError: errInvalidToken,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Reset mock for each test case
			suite.mockJWT = jwtmock.NewJWTServiceInterfaceMock(suite.T())
			if tt.setupMock != nil {
				tt.setupMock(suite.mockJWT)
			}
			suite.authenticator = newJWTAuthenticator(suite.mockJWT)

			req := httptest.NewRequest(http.MethodGet, "/users", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			authCtx, err := suite.authenticator.Authenticate(req)

			if tt.expectedError != nil {
				assert.ErrorIs(suite.T(), err, tt.expectedError)
				assert.Nil(suite.T(), authCtx)
			} else {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), authCtx)
				if tt.validateResult != nil {
					tt.validateResult(suite.T(), authCtx)
				}
			}

			suite.mockJWT.AssertExpectations(suite.T())
		})
	}
}

func (suite *JWTAuthenticatorTestSuite) TestAuthorize() {
	tests := []struct {
		name          string
		attributes    map[string]interface{}
		expectedError error
	}{
		{
			name: "Has system scope in scope attribute",
			attributes: map[string]interface{}{
				"scope": "system users:read",
			},
			expectedError: nil,
		},
		{
			name: "Has authorized_permissions attribute",
			attributes: map[string]interface{}{
				"authorized_permissions": "other system",
			},
			expectedError: nil,
		},
		{
			name: "Missing required scopes",
			attributes: map[string]interface{}{
				"scope": "users:read",
			},
			expectedError: errInsufficientScopes,
		},
		{
			name:          "Nil authentication context",
			attributes:    nil,
			expectedError: errUnauthorized,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			req := httptest.NewRequest(http.MethodGet, "/users", nil)

			var authCtx *SecurityContext
			if tt.attributes != nil {
				authCtx = newSecurityContext("user123", "ou1", "app1", "token", tt.attributes)
				ctx := withSecurityContext(req.Context(), authCtx)
				req = req.WithContext(ctx)
			}

			err := suite.authenticator.Authorize(req, authCtx)

			if tt.expectedError != nil {
				assert.ErrorIs(suite.T(), err, tt.expectedError)
			} else {
				assert.NoError(suite.T(), err)
			}
		})
	}
}

func (suite *JWTAuthenticatorTestSuite) TestExtractScopes() {
	tests := []struct {
		name           string
		attributes     map[string]interface{}
		expectedScopes []string
	}{
		{
			name: "OAuth2 standard scope attribute (space-separated)",
			attributes: map[string]interface{}{
				"scope": "users:read users:write applications:manage",
			},
			expectedScopes: []string{"users:read", "users:write", "applications:manage"},
		},
		{
			name: "Scopes as array of strings",
			attributes: map[string]interface{}{
				"scopes": []string{"users:read", "users:write"},
			},
			expectedScopes: []string{"users:read", "users:write"},
		},
		{
			name: "Scopes as array of interfaces",
			attributes: map[string]interface{}{
				"scopes": []interface{}{"users:read", "users:write"},
			},
			expectedScopes: []string{"users:read", "users:write"},
		},
		{
			name: "Empty scope attribute",
			attributes: map[string]interface{}{
				"scope": "",
			},
			expectedScopes: []string{},
		},
		{
			name:           "No scope attribute",
			attributes:     map[string]interface{}{},
			expectedScopes: []string{},
		},
		{
			name: "Single scope",
			attributes: map[string]interface{}{
				"scope": "users:read",
			},
			expectedScopes: []string{"users:read"},
		},
		{
			name: "Thunder assertion authorized_permissions attribute",
			attributes: map[string]interface{}{
				"authorized_permissions": "perm1 perm2 perm3",
			},
			expectedScopes: []string{"perm1", "perm2", "perm3"},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			scopes := extractScopes(tt.attributes)
			assert.ElementsMatch(suite.T(), tt.expectedScopes, scopes)
		})
	}
}

func (suite *JWTAuthenticatorTestSuite) TestExtractAttribute() {
	tests := []struct {
		name          string
		attributes    map[string]interface{}
		key           string
		expectedValue string
	}{
		{
			name:          "Existing string attribute",
			attributes:    map[string]interface{}{"ou_id": "ou123"},
			key:           "ou_id",
			expectedValue: "ou123",
		},
		{
			name:          "Non-existent attribute",
			attributes:    map[string]interface{}{"other": "value"},
			key:           "ou_id",
			expectedValue: "",
		},
		{
			name:          "Non-string attribute value",
			attributes:    map[string]interface{}{"ou_id": 123},
			key:           "ou_id",
			expectedValue: "",
		},
		{
			name:          "Empty attributes",
			attributes:    map[string]interface{}{},
			key:           "ou_id",
			expectedValue: "",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := extractAttribute(tt.attributes, tt.key)
			assert.Equal(suite.T(), tt.expectedValue, result)
		})
	}
}

func (suite *JWTAuthenticatorTestSuite) TestHasAnyScope() {
	tests := []struct {
		name           string
		userScopes     []string
		requiredScopes []string
		expectedResult bool
	}{
		{
			name:           "User has one of the required scopes",
			userScopes:     []string{"users:read", "groups:manage"},
			requiredScopes: []string{"users:read", "users:write"},
			expectedResult: true,
		},
		{
			name:           "User has all required scopes",
			userScopes:     []string{"users:read", "users:write"},
			requiredScopes: []string{"users:read", "users:write"},
			expectedResult: true,
		},
		{
			name:           "User has none of the required scopes",
			userScopes:     []string{"groups:manage"},
			requiredScopes: []string{"users:read", "users:write"},
			expectedResult: false,
		},
		{
			name:           "No required scopes",
			userScopes:     []string{"users:read"},
			requiredScopes: []string{},
			expectedResult: true,
		},
		{
			name:           "Empty user scopes",
			userScopes:     []string{},
			requiredScopes: []string{"users:read"},
			expectedResult: false,
		},
		{
			name:           "Both empty",
			userScopes:     []string{},
			requiredScopes: []string{},
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := hasAnyScope(tt.userScopes, tt.requiredScopes)
			assert.Equal(suite.T(), tt.expectedResult, result)
		})
	}
}

func (suite *JWTAuthenticatorTestSuite) TestGetRequiredScopes() {
	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{
			name:     "Users endpoint",
			path:     "/users",
			expected: []string{"system"},
		},
		{
			name:     "Applications endpoint",
			path:     "/applications",
			expected: []string{"system"},
		},
		{
			name:     "Groups endpoint",
			path:     "/groups",
			expected: []string{"system"},
		},
		{
			name:     "Root path",
			path:     "/",
			expected: []string{"system"},
		},
		{
			name:     "Any other path",
			path:     "/some/other/path",
			expected: []string{"system"},
		},
		{
			name:     "User self-service endpoint - /users/me",
			path:     "/users/me",
			expected: []string{},
		},
		{
			name:     "User self-service endpoint - /users/me/update-credentials",
			path:     "/users/me/update-credentials",
			expected: []string{},
		},
		{
			name:     "Passkey registration start endpoint",
			path:     "/register/passkey/start",
			expected: []string{},
		},
		{
			name:     "Passkey registration finish endpoint",
			path:     "/register/passkey/finish",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			result := suite.authenticator.getRequiredScopes(req)
			assert.Equal(suite.T(), tt.expected, result)
		})
	}
}

func (suite *JWTAuthenticatorTestSuite) TestExtractScopes_EdgeCases() {
	tests := []struct {
		name           string
		attributes     map[string]interface{}
		expectedScopes []string
	}{
		{
			name: "Scopes array with mixed types (should filter non-strings)",
			attributes: map[string]interface{}{
				"scopes": []interface{}{"valid", 123, true, "another_valid"},
			},
			expectedScopes: []string{"valid", "another_valid"},
		},
		{
			name: "Scopes as non-array, non-string type",
			attributes: map[string]interface{}{
				"scopes": map[string]string{"invalid": "format"},
			},
			expectedScopes: []string{},
		},
		{
			name: "Scope attribute with extra whitespace",
			attributes: map[string]interface{}{
				"scope": "  users:read   users:write  ",
			},
			expectedScopes: []string{"users:read", "users:write"},
		},
		{
			name: "Both scope and scopes present (scope takes precedence)",
			attributes: map[string]interface{}{
				"scope":  "from_scope",
				"scopes": []string{"from_scopes"},
			},
			expectedScopes: []string{"from_scope"},
		},
		{
			name: "Scope as non-string type",
			attributes: map[string]interface{}{
				"scope": 12345,
			},
			expectedScopes: []string{},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			scopes := extractScopes(tt.attributes)
			assert.ElementsMatch(suite.T(), tt.expectedScopes, scopes)
		})
	}
}

func (suite *JWTAuthenticatorTestSuite) TestNewJWTAuthenticator() {
	mockJWTService := jwtmock.NewJWTServiceInterfaceMock(suite.T())

	authenticator := newJWTAuthenticator(mockJWTService)

	assert.NotNil(suite.T(), authenticator)
	assert.Equal(suite.T(), mockJWTService, authenticator.jwtService)
}

func (suite *JWTAuthenticatorTestSuite) TestCanHandle_EdgeCases() {
	tests := []struct {
		name           string
		setupRequest   func() *http.Request
		expectedResult bool
	}{
		{
			name: "Bearer with space but no token",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/users", nil)
				req.Header.Set("Authorization", "Bearer ")
				return req
			},
			expectedResult: true, // CanHandle only checks prefix, validation is in Authenticate
		},
		{
			name: "Bearer with tab character",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/users", nil)
				req.Header.Set("Authorization", "Bearer\ttoken123")
				return req
			},
			expectedResult: false,
		},
		{
			name: "Multiple Authorization headers",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/users", nil)
				req.Header.Add("Authorization", "Basic xyz")
				req.Header.Add("Authorization", "Bearer token123")
				return req
			},
			expectedResult: false, // Get() returns first header
		},
		{
			name: "Case sensitive Bearer",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/users", nil)
				req.Header.Set("Authorization", "BEARER token123")
				return req
			},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			req := tt.setupRequest()
			result := suite.authenticator.CanHandle(req)
			assert.Equal(suite.T(), tt.expectedResult, result)
		})
	}
}
