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
	"github.com/asgardeo/thunder/tests/mocks/jose/jwtmock"
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
				assert.Equal(t, "user123", GetSubject(baseCtx))
				assert.Equal(t, "ou1", GetOUID(baseCtx))
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

func (suite *JWTAuthenticatorTestSuite) TestExtractPermissionsFromJWTClaims() {
	tests := []struct {
		name                string
		attributes          map[string]interface{}
		expectedPermissions []string
	}{
		{
			name: "OAuth2 standard scope attribute (space-separated)",
			attributes: map[string]interface{}{
				"scope": "users:read users:write applications:manage",
			},
			expectedPermissions: []string{"users:read", "users:write", "applications:manage"},
		},
		{
			name: "Scopes as array of strings",
			attributes: map[string]interface{}{
				"scopes": []string{"users:read", "users:write"},
			},
			expectedPermissions: []string{"users:read", "users:write"},
		},
		{
			name: "Scopes as array of interfaces",
			attributes: map[string]interface{}{
				"scopes": []interface{}{"users:read", "users:write"},
			},
			expectedPermissions: []string{"users:read", "users:write"},
		},
		{
			name: "Empty scope attribute",
			attributes: map[string]interface{}{
				"scope": "",
			},
			expectedPermissions: []string{},
		},
		{
			name:                "No scope attribute",
			attributes:          map[string]interface{}{},
			expectedPermissions: []string{},
		},
		{
			name: "Single scope",
			attributes: map[string]interface{}{
				"scope": "users:read",
			},
			expectedPermissions: []string{"users:read"},
		},
		{
			name: "Thunder assertion authorized_permissions attribute",
			attributes: map[string]interface{}{
				"authorized_permissions": "perm1 perm2 perm3",
			},
			expectedPermissions: []string{"perm1", "perm2", "perm3"},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			permissions := extractScopes(tt.attributes)
			assert.ElementsMatch(suite.T(), tt.expectedPermissions, permissions)
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

func (suite *JWTAuthenticatorTestSuite) TestExtractPermissionsFromJWTClaims_EdgeCases() {
	tests := []struct {
		name                string
		attributes          map[string]interface{}
		expectedPermissions []string
	}{
		{
			name: "Scopes array with mixed types (should filter non-strings)",
			attributes: map[string]interface{}{
				"scopes": []interface{}{"valid", 123, true, "another_valid"},
			},
			expectedPermissions: []string{"valid", "another_valid"},
		},
		{
			name: "Scopes as non-array, non-string type",
			attributes: map[string]interface{}{
				"scopes": map[string]string{"invalid": "format"},
			},
			expectedPermissions: []string{},
		},
		{
			name: "Scope attribute with extra whitespace",
			attributes: map[string]interface{}{
				"scope": "  users:read   users:write  ",
			},
			expectedPermissions: []string{"users:read", "users:write"},
		},
		{
			name: "Both scope and scopes present (scope takes precedence)",
			attributes: map[string]interface{}{
				"scope":  "from_scope",
				"scopes": []string{"from_scopes"},
			},
			expectedPermissions: []string{"from_scope"},
		},
		{
			name: "Scope as non-string type",
			attributes: map[string]interface{}{
				"scope": 12345,
			},
			expectedPermissions: []string{},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			permissions := extractScopes(tt.attributes)
			assert.ElementsMatch(suite.T(), tt.expectedPermissions, permissions)
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
