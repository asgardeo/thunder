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
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var testPublicPaths = []string{
	"/health/**",
	"/auth/**",
	"/flow/execute/**",
	"/oauth2/**",
	"/.well-known/openid-configuration/**",
	"/.well-known/oauth-authorization-server/**",
	"/gate/**",
	"/develop/**",
	"/error/**",
	"/design/resolve/**",
	"/i18n/languages",
	"/i18n/languages/*/translations/resolve",
	"/i18n/languages/*/translations/ns/*/keys/*/resolve",
}

// SecurityServiceTestSuite defines the test suite for SecurityService
type SecurityServiceTestSuite struct {
	suite.Suite
	service   *securityService
	mockAuth1 *AuthenticatorInterfaceMock
	mockAuth2 *AuthenticatorInterfaceMock
	testCtx   *SecurityContext
}

func (suite *SecurityServiceTestSuite) SetupTest() {
	suite.mockAuth1 = &AuthenticatorInterfaceMock{}
	suite.mockAuth2 = &AuthenticatorInterfaceMock{}

	var err error
	suite.service, err = newSecurityService(
		[]AuthenticatorInterface{suite.mockAuth1, suite.mockAuth2}, testPublicPaths, apiPermissionEntries)
	suite.Require().NoError(err)

	// Create test authentication context with "system" permission so that
	// the service-level authorization check passes for protected /api/* paths.
	suite.testCtx = newSecurityContext(
		"user123",
		"ou456",
		"test_token",
		[]string{"system"},
		map[string]interface{}{
			"scope": []string{"read", "write"},
			"role":  "admin",
		},
	)
}

func (suite *SecurityServiceTestSuite) TearDownTest() {
	suite.mockAuth1.AssertExpectations(suite.T())
	suite.mockAuth2.AssertExpectations(suite.T())
}

// Run the test suite
func TestSecurityServiceSuite(t *testing.T) {
	suite.Run(t, new(SecurityServiceTestSuite))
}

// Test Process method with public paths
func (suite *SecurityServiceTestSuite) TestProcess_PublicPaths() {
	testCases := []struct {
		name string
		path string
	}{
		{"Auth path", "/auth/login"},
		{"Auth path with subpath", "/auth/register/user"},
		{"OAuth2 token", "/oauth2/token"},
		{"OAuth2 authorize", "/oauth2/authorize"},
		{"OAuth2 well-known", "/oauth2/.well-known/openid_configuration"},
		{"OAuth2 JWKS", "/oauth2/jwks"},
		{"OAuth2 register", "/oauth2/register"},
		{"Health check liveness", "/health/liveness"},
		{"Health check readiness", "/health/readiness"},
		{"Signin path", "/gate/verify"},
		{"Signin path with subpath", "/gate/forgot-password"},
		{"Develop path", "/develop/dashboard"},
		{"Develop path with subpath", "/develop/api/test"},
		{"Auth without trailing slash", "/auth"},
		{"OAuth2 token without params", "/oauth2/token"},
		{"Signin without trailing slash", "/gate/signin"},
		{"Develop without trailing slash", "/develop"},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)

			// With optional auth, authenticators are checked even for public paths
			// Since there is no auth header, CanHandle returns false
			suite.mockAuth1.On("CanHandle", req).Return(false)
			suite.mockAuth2.On("CanHandle", req).Return(false)

			ctx, err := suite.service.Process(req)

			assert.NoError(suite.T(), err)
			assert.NotNil(suite.T(), ctx)
			assert.True(suite.T(), IsRuntimeContext(ctx), "public path should return a runtime context")
		})
	}
}

// Test Process method with non-public paths and successful authentication
func (suite *SecurityServiceTestSuite) TestProcess_SuccessfulAuthentication_FirstAuthenticator() {
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)

	// First authenticator can handle the request
	suite.mockAuth1.On("CanHandle", req).Return(true)
	suite.mockAuth1.On("Authenticate", req).Return(suite.testCtx, nil)

	ctx, err := suite.service.Process(req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), ctx)

	// Verify authentication context is added to the context
	userID := GetSubject(ctx)
	assert.Equal(suite.T(), "user123", userID)

	ouID := GetOUID(ctx)
	assert.Equal(suite.T(), "ou456", ouID)

	// Second authenticator should not be called
	suite.mockAuth2.AssertNotCalled(suite.T(), "CanHandle")
	suite.mockAuth2.AssertNotCalled(suite.T(), "Authenticate")
}

// Test Process method with second authenticator handling the request
func (suite *SecurityServiceTestSuite) TestProcess_SuccessfulAuthentication_SecondAuthenticator() {
	req := httptest.NewRequest(http.MethodPost, "/api/groups", nil)

	// First authenticator cannot handle the request, second can
	suite.mockAuth1.On("CanHandle", req).Return(false)
	suite.mockAuth2.On("CanHandle", req).Return(true)
	suite.mockAuth2.On("Authenticate", req).Return(suite.testCtx, nil)

	ctx, err := suite.service.Process(req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), ctx)

	// Verify authentication context is added
	userID := GetSubject(ctx)
	assert.Equal(suite.T(), "user123", userID)
}

// Test Process method when no authenticator can handle the request
func (suite *SecurityServiceTestSuite) TestProcess_NoHandlerFound() {
	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)

	// Both authenticators cannot handle the request
	suite.mockAuth1.On("CanHandle", req).Return(false)
	suite.mockAuth2.On("CanHandle", req).Return(false)

	ctx, err := suite.service.Process(req)

	assert.Nil(suite.T(), ctx)
	assert.Equal(suite.T(), errNoHandlerFound, err)

	// Verify neither authenticate method was called
	suite.mockAuth1.AssertNotCalled(suite.T(), "Authenticate")
	suite.mockAuth2.AssertNotCalled(suite.T(), "Authenticate")
}

// Test Process method when authentication fails
func (suite *SecurityServiceTestSuite) TestProcess_AuthenticationFailure() {
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	authError := errors.New("invalid credentials")

	suite.mockAuth1.On("CanHandle", req).Return(true)
	suite.mockAuth1.On("Authenticate", req).Return(nil, authError)

	ctx, err := suite.service.Process(req)

	assert.Nil(suite.T(), ctx)
	assert.Equal(suite.T(), authError, err)
}

// Test Process method with specific security errors
func (suite *SecurityServiceTestSuite) TestProcess_SecurityErrors() {
	testCases := []struct {
		name  string
		error error
	}{
		{"Unauthorized error", errUnauthorized},
		{"Forbidden error", errForbidden},
		{"Invalid token error", errInvalidToken},
		{"Insufficient permissions error", errInsufficientPermissions},
		{"Missing auth header error", errMissingAuthHeader},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)

			// Reset mocks for each test case
			suite.mockAuth1 = &AuthenticatorInterfaceMock{}
			suite.mockAuth2 = &AuthenticatorInterfaceMock{}
			suite.service.authenticators = []AuthenticatorInterface{suite.mockAuth1, suite.mockAuth2}

			suite.mockAuth1.On("CanHandle", req).Return(true)
			suite.mockAuth1.On("Authenticate", req).Return(nil, tc.error)

			ctx, err := suite.service.Process(req)

			assert.Nil(suite.T(), ctx)
			assert.Equal(suite.T(), tc.error, err)

			suite.mockAuth1.AssertExpectations(suite.T())
		})
	}
}

// Test Process method with nil authenticator context.
// A nil SecurityContext means no permissions are available; the service-level
// authorization check allows the request through when the path requires no
// special permission (e.g. /users/me).
func (suite *SecurityServiceTestSuite) TestProcess_NilSecurityContext() {
	// /users/me requires no special permission, so a nil context must still pass.
	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)

	suite.mockAuth1.On("CanHandle", req).Return(true)
	suite.mockAuth1.On("Authenticate", req).Return(nil, nil)

	ctx, err := suite.service.Process(req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), ctx)

	// Verify empty context values when auth context is nil
	userID := GetSubject(ctx)
	assert.Empty(suite.T(), userID)
}

// Test isPublicPath method directly
func (suite *SecurityServiceTestSuite) TestIsPublicPath() {
	testCases := []struct {
		name     string
		path     string
		expected bool
	}{
		// Public paths - should return true
		{"Auth root", "/auth/", true},
		{"Auth login", "/auth/credentials/authenticate", true},
		{"OAuth2 token", "/oauth2/token", true},
		{"OAuth2 authorize", "/oauth2/authorize", true},
		{"OAuth2 well-known", "/.well-known/openid-configuration", true},
		{"OAuth2 JWKS", "/oauth2/jwks", true},
		{"OAuth2 register", "/oauth2/register", true},
		{"Health check", "/health/liveness", true},
		{"Signin root", "/gate/signin", true},
		{"Signin logo", "/gate/signin/logo/123", true},
		{"Develop root", "/develop/", true},
		{"Develop dashboard", "/develop/dashboard", true},
		{"I18n languages", "/i18n/languages", true},

		// Exact matches without trailing slash
		{"Auth exact", "/auth", true},
		{"OAuth2 token exact", "/oauth2/token", true},
		{"Signin exact", "/gate/signin", true},
		{"Develop exact", "/develop", true},

		// Non-public paths - should return false
		{"API users", "/api/users", false},
		{"API groups", "/api/groups", false},
		{"Admin panel", "/admin/dashboard", false},
		{"Root path", "/", false},
		{"Random path", "/random/path", false},
		{"Similar but not exact", "/authentication", false},
		{"Similar prefix", "/oauth", false},
		{"Not allowed sub prefix", "/flow", false},

		// Edge cases
		{"Empty path", "", false},
		{"Just slash", "/", false},

		// Parameterized paths
		{"Parameterized path match", "/i18n/languages/en/translations/resolve", true},
		{"Parameterized path mismatch prefix", "/i18n/languages/en/translations", false},
		{"Parameterized path mismatch suffix", "/i18n/languages/en/translations/resolve/extra", false},
		{"Parameterized path empty param", "/i18n/languages//translations/resolve", false},

		// Multi-parameter paths
		{"Multi-param path match", "/i18n/languages/en/translations/ns/common/keys/btn.submit/resolve", true},
		{"Multi-param path mismatch namespace", "/i18n/languages/en/translations/ns//keys/btn.submit/resolve", false},
		{"Multi-param path mismatch key", "/i18n/languages/en/translations/ns/common/keys//resolve", false},
		{"Multi-param path mismatch structure", "/i18n/languages/en/translations/ns/common/keys/btn.submit", false},

		// Special characters in parameters
		{"Parameterized path with hyphen", "/i18n/languages/en-US/translations/resolve", true},
		{"Multi-param path with dots", "/i18n/languages/en/translations/ns/common/keys/btn.submit.label/resolve", true},

		// Performance/Robustness edge cases
		{"Long parameter value within limit",
			"/i18n/languages/" + strings.Repeat("a", 255) + "/translations/resolve", true},
		{"Exceeds max path length", "/i18n/languages/" + strings.Repeat("a", 4096) + "/translations/resolve", false},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			result := suite.service.isPublicPath(tc.path)
			assert.Equal(suite.T(), tc.expected, result, "Path: %s", tc.path)
		})
	}
}

// Test SecurityService with empty authenticators list
func (suite *SecurityServiceTestSuite) TestProcess_EmptyAuthenticators() {
	service, err := newSecurityService([]AuthenticatorInterface{}, testPublicPaths, apiPermissionEntries)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)

	ctx, err := service.Process(req)

	assert.Nil(suite.T(), ctx)
	assert.Equal(suite.T(), errNoHandlerFound, err)
}

// Test SecurityService with nil authenticators list
func (suite *SecurityServiceTestSuite) TestProcess_NilAuthenticators() {
	service, err := newSecurityService(nil, testPublicPaths, apiPermissionEntries)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)

	ctx, err := service.Process(req)

	assert.Nil(suite.T(), ctx)
	assert.Equal(suite.T(), errNoHandlerFound, err)
}

// Test Process with different HTTP methods
func (suite *SecurityServiceTestSuite) TestProcess_DifferentHTTPMethods() {
	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
		http.MethodHead,
	}

	for _, method := range methods {
		suite.Run("Method_"+method, func() {
			req := httptest.NewRequest(method, "/api/test", nil)

			// Reset mocks for each test case
			suite.mockAuth1 = &AuthenticatorInterfaceMock{}
			suite.mockAuth2 = &AuthenticatorInterfaceMock{}
			suite.service.authenticators = []AuthenticatorInterface{suite.mockAuth1, suite.mockAuth2}

			suite.mockAuth1.On("CanHandle", req).Return(true)
			suite.mockAuth1.On("Authenticate", req).Return(suite.testCtx, nil)

			ctx, err := suite.service.Process(req)

			assert.NoError(suite.T(), err)
			assert.NotNil(suite.T(), ctx)

			userID := GetSubject(ctx)
			assert.Equal(suite.T(), "user123", userID)

			suite.mockAuth1.AssertExpectations(suite.T())
		})
	}
}

// Test Process with various public path variations
func (suite *SecurityServiceTestSuite) TestProcess_PublicPathVariations() {
	testCases := []struct {
		name string
		path string
	}{
		// Test case sensitivity and exact matching
		{"OAuth2 with query params", "/oauth2/token?grant_type=authorization_code"},
		{"Auth with fragment", "/auth/login#section"},
		{"Well-known with path", "/oauth2/.well-known/openid_configuration"},
		{"Nested signin path", "/gate/forgot-password/confirm"},
		{"Deep develop path", "/develop/api/v1/test"},
		{"Health check with query", "/health/liveness?detailed=true"},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)

			// With optional auth, authenticators are checked even for public paths
			suite.mockAuth1.On("CanHandle", req).Return(false)
			suite.mockAuth2.On("CanHandle", req).Return(false)

			ctx, err := suite.service.Process(req)

			assert.NoError(suite.T(), err, "Path should be public: %s", tc.path)
			assert.NotNil(suite.T(), ctx)
			assert.True(suite.T(), IsRuntimeContext(ctx), "public path should return a runtime context: %s", tc.path)
		})
	}
}

// Test OPTIONS method bypasses authentication
func (suite *SecurityServiceTestSuite) TestProcess_OptionsMethod() {
	req := httptest.NewRequest(http.MethodOptions, "/api/protected", nil)

	ctx, err := suite.service.Process(req)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), req.Context(), ctx)

	// Verify no authenticators were called for OPTIONS method
	suite.mockAuth1.AssertNotCalled(suite.T(), "CanHandle")
	suite.mockAuth2.AssertNotCalled(suite.T(), "CanHandle")
}

// TestNewSecurityService_Error verifies that newSecurityService returns an error
// when either the public path patterns or the API permission entry patterns are invalid.
func (suite *SecurityServiceTestSuite) TestNewSecurityService_Error() {
	tests := []struct {
		name        string
		publicPaths []string
		apiPerms    []apiPermissionEntry
		errContains string
	}{
		{
			name:        "invalid public path pattern",
			publicPaths: []string{"/valid", "/invalid/**/middle/**"},
			apiPerms:    apiPermissionEntries,
			errContains: "invalid pattern",
		},
		{
			name:        "invalid API permission entry pattern",
			publicPaths: []string{},
			apiPerms:    []apiPermissionEntry{{"GET /invalid/**/middle/**", PermissionUser}},
			errContains: "invalid pattern",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			service, err := newSecurityService(nil, tt.publicPaths, tt.apiPerms)
			assert.Error(suite.T(), err)
			assert.Nil(suite.T(), service)
			assert.Contains(suite.T(), err.Error(), tt.errContains)
		})
	}
}

// Test Process method with public path and valid token
func (suite *SecurityServiceTestSuite) TestProcess_PublicPath_WithToken() {
	req := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
	// Add Authorization header to simulate optional auth
	req.Header.Add("Authorization", "Bearer valid_token")

	// First authenticator can handle the request
	suite.mockAuth1.On("CanHandle", req).Return(true)
	suite.mockAuth1.On("Authenticate", req).Return(suite.testCtx, nil)

	ctx, err := suite.service.Process(req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), ctx)

	// Verify authentication context is added
	userID := GetSubject(ctx)
	assert.Equal(suite.T(), "user123", userID)
}

// Test Process method with public path and invalid token (optional auth failure)
func (suite *SecurityServiceTestSuite) TestProcess_PublicPath_WithInvalidToken() {
	req := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
	// Add Authorization header
	req.Header.Add("Authorization", "Bearer invalid_token")

	authError := errors.New("invalid token")

	// First authenticator handles it but fails
	suite.mockAuth1.On("CanHandle", req).Return(true)
	suite.mockAuth1.On("Authenticate", req).Return(nil, authError)

	ctx, err := suite.service.Process(req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), ctx)
	assert.True(suite.T(), IsRuntimeContext(ctx), "public path with invalid token should return a runtime context")

	userID := GetSubject(ctx)
	assert.Empty(suite.T(), userID)
}

// Test skipSecurity flag allows requests to proceed even with auth failures
func (suite *SecurityServiceTestSuite) TestProcess_SkipSecurity_AllowsAccessWithoutToken() {
	_ = os.Setenv("THUNDER_SKIP_SECURITY", "true")
	defer func() { _ = os.Unsetenv("THUNDER_SKIP_SECURITY") }()

	service, err := newSecurityService([]AuthenticatorInterface{suite.mockAuth1}, testPublicPaths, apiPermissionEntries)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)

	// No authenticator can handle the request (no token)
	suite.mockAuth1.On("CanHandle", req).Return(false)

	ctx, err := service.Process(req)

	// Should succeed without authentication
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), ctx)
	// Security was skipped: context must carry the skipped marker and have no subject.
	assert.True(suite.T(), IsSecuritySkipped(ctx))
	assert.Empty(suite.T(), GetSubject(ctx))
}

// setupSkipSecurityTest is a helper function to set up skip security tests
func (suite *SecurityServiceTestSuite) setupSkipSecurityTest() (*http.Request, SecurityServiceInterface) {
	_ = os.Setenv("THUNDER_SKIP_SECURITY", "true")
	suite.T().Cleanup(func() { _ = os.Unsetenv("THUNDER_SKIP_SECURITY") })

	service, err := newSecurityService([]AuthenticatorInterface{suite.mockAuth1}, testPublicPaths, apiPermissionEntries)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
	req.Header.Set("Authorization", "Bearer valid_token")
	return req, service
}

// Test skipSecurity flag enriches context when valid token is present
func (suite *SecurityServiceTestSuite) TestProcess_SkipSecurity_EnrichesContextWithValidToken() {
	req, service := suite.setupSkipSecurityTest()

	// Authenticator handles request and succeeds
	suite.mockAuth1.On("CanHandle", req).Return(true)
	suite.mockAuth1.On("Authenticate", req).Return(suite.testCtx, nil)

	ctx, err := service.Process(req)

	// Should succeed and context should be enriched
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), ctx)

	userID := GetSubject(ctx)
	assert.Equal(suite.T(), "user123", userID)
}

// Test skipSecurity flag allows access even with invalid token
func (suite *SecurityServiceTestSuite) TestProcess_SkipSecurity_AllowsAccessWithInvalidToken() {
	_ = os.Setenv("THUNDER_SKIP_SECURITY", "true")
	defer func() { _ = os.Unsetenv("THUNDER_SKIP_SECURITY") }()

	service, err := newSecurityService([]AuthenticatorInterface{suite.mockAuth1}, testPublicPaths, apiPermissionEntries)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid_token")

	// Authenticator handles request but authentication fails
	suite.mockAuth1.On("CanHandle", req).Return(true)
	suite.mockAuth1.On("Authenticate", req).Return(nil, errInvalidToken)

	ctx, err := service.Process(req)

	// Should succeed even though authentication failed
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), ctx)
	// Security was skipped: context must carry the skipped marker and have no subject.
	assert.True(suite.T(), IsSecuritySkipped(ctx))
	userID := GetSubject(ctx)
	assert.Empty(suite.T(), userID)
}

// Test skipSecurity flag allows access even when the caller has no permissions.
// A SecurityContext without "system" permission would normally trigger
// errInsufficientPermissions on the /api/protected path, but skipSecurity
// bypasses the enforcement.
func (suite *SecurityServiceTestSuite) TestProcess_SkipSecurity_AllowsAccessWithInsufficientPermissions() {
	req, service := suite.setupSkipSecurityTest()

	// Create a context with no permissions — authorization will fail on /api/protected.
	unprivCtx := newSecurityContext("user123", "ou456", "test_token", []string{}, nil)

	// Authenticator handles request and authentication succeeds.
	suite.mockAuth1.On("CanHandle", req).Return(true)
	suite.mockAuth1.On("Authenticate", req).Return(unprivCtx, nil)

	ctx, err := service.Process(req)

	// Should succeed even though authorization would have failed.
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), ctx)
	// Security was skipped: context must carry the skipped marker.
	assert.True(suite.T(), IsSecuritySkipped(ctx))
	// Context should still be enriched (authentication succeeded).
	userID := GetSubject(ctx)
	assert.Equal(suite.T(), "user123", userID)
}

// Test that the skipped marker is NOT present when authentication and authorization succeed normally.
func (suite *SecurityServiceTestSuite) TestProcess_SecurityNotSkipped_WhenAuthSucceeds() {
	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)

	suite.mockAuth1.On("CanHandle", req).Return(true)
	suite.mockAuth1.On("Authenticate", req).Return(suite.testCtx, nil)

	ctx, err := suite.service.Process(req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), ctx)
	// Normal successful flow must NOT mark the context as skipped.
	assert.False(suite.T(), IsSecuritySkipped(ctx))
	assert.Equal(suite.T(), "user123", GetSubject(ctx))
}

// Test that the service returns errInsufficientPermissions when the authenticated
// subject lacks the required permission for a protected path.
func (suite *SecurityServiceTestSuite) TestProcess_AuthorizationFailure_InsufficientPermissions() {
	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)

	// Subject has no useful permissions — /api/protected requires "system".
	unprivCtx := newSecurityContext("user123", "ou456", "test_token", []string{"other"}, nil)

	suite.mockAuth1.On("CanHandle", req).Return(true)
	suite.mockAuth1.On("Authenticate", req).Return(unprivCtx, nil)

	ctx, err := suite.service.Process(req)

	assert.Nil(suite.T(), ctx)
	assert.ErrorIs(suite.T(), err, errInsufficientPermissions)
}

// Test GetRequiredPermissionForAPI returns the expected permission for various method+path combos.
func (suite *SecurityServiceTestSuite) TestGetRequiredPermissionForAPI() {
	tests := []struct {
		name     string
		method   string
		path     string
		expected string
	}{
		{
			name:     "GET /organization-units requires ou:view",
			method:   http.MethodGet,
			path:     "/organization-units",
			expected: PermissionOUView,
		},
		{
			name:     "POST /organization-units requires ou",
			method:   http.MethodPost,
			path:     "/organization-units",
			expected: PermissionOU,
		},
		{
			name:     "GET /organization-units/{id} requires ou:view",
			method:   http.MethodGet,
			path:     "/organization-units/ou-123",
			expected: PermissionOUView,
		},
		{
			name:     "DELETE /organization-units/{id} requires ou",
			method:   http.MethodDelete,
			path:     "/organization-units/ou-123",
			expected: PermissionOU,
		},
		{
			name:     "GET /users requires user:view",
			method:   http.MethodGet,
			path:     "/users",
			expected: PermissionUserView,
		},
		{
			name:     "POST /users requires user",
			method:   http.MethodPost,
			path:     "/users",
			expected: PermissionUser,
		},
		{
			name:     "Unmapped path falls back to system",
			method:   http.MethodGet,
			path:     "/applications",
			expected: SystemPermission,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := suite.service.getRequiredPermissionForAPI(tt.method, tt.path)
			assert.Equal(suite.T(), tt.expected, result)
		})
	}
}

// Test HasSufficientPermission covers hierarchical scope matching.
func (suite *SecurityServiceTestSuite) TestHasSufficientPermission() {
	tests := []struct {
		name            string
		userPermissions []string
		required        string
		expected        bool
	}{
		{
			name:            "Exact match",
			userPermissions: []string{"system:ou:view"},
			required:        "system:ou:view",
			expected:        true,
		},
		{
			name:            "Parent scope satisfies child scope",
			userPermissions: []string{"system:ou"},
			required:        "system:ou:view",
			expected:        true,
		},
		{
			name:            "Root system scope satisfies any child scope",
			userPermissions: []string{"system"},
			required:        "system:ou:view",
			expected:        true,
		},
		{
			name:            "Child scope does not satisfy parent scope",
			userPermissions: []string{"system:ou:view"},
			required:        "system:ou",
			expected:        false,
		},
		{
			name:            "Unrelated sibling scope does not satisfy",
			userPermissions: []string{"system:user"},
			required:        "system:ou:view",
			expected:        false,
		},
		{
			name:            "Empty user permissions",
			userPermissions: []string{},
			required:        "system",
			expected:        false,
		},
		{
			name:            "Partial prefix should not match",
			userPermissions: []string{"sys"},
			required:        "system:ou",
			expected:        false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := HasSufficientPermission(tt.userPermissions, tt.required)
			assert.Equal(suite.T(), tt.expected, result)
		})
	}
}
