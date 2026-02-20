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
	suite.service, err = newSecurityService([]AuthenticatorInterface{suite.mockAuth1, suite.mockAuth2}, testPublicPaths)
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
			assert.Equal(suite.T(), req.Context(), ctx)
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
	service, err := newSecurityService([]AuthenticatorInterface{}, testPublicPaths)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)

	ctx, err := service.Process(req)

	assert.Nil(suite.T(), ctx)
	assert.Equal(suite.T(), errNoHandlerFound, err)
}

// Test SecurityService with nil authenticators list
func (suite *SecurityServiceTestSuite) TestProcess_NilAuthenticators() {
	service, err := newSecurityService(nil, testPublicPaths)
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
			assert.Equal(suite.T(), req.Context(), ctx)
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

// Test newSecurityService returns error on invalid paths
func (suite *SecurityServiceTestSuite) TestNewSecurityService_Error() {
	invalidPaths := []string{"/valid", "/invalid/**/middle/**"}
	service, err := newSecurityService(nil, invalidPaths)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), service)
	assert.Contains(suite.T(), err.Error(), "invalid pattern")
}

func (suite *SecurityServiceTestSuite) TestCompilePathPatterns() {
	tests := []struct {
		name           string
		pattern        string
		expectedRegex  string
		shouldMatch    []string
		shouldNotMatch []string
	}{
		{
			name:           "Single wildcard segment",
			pattern:        "/api/*/users",
			expectedRegex:  "^/api/[^/]+/users$",
			shouldMatch:    []string{"/api/v1/users", "/api/test/users"},
			shouldNotMatch: []string{"/api/users", "/api/v1/v2/users"},
		},
		{
			name:           "Recursive wildcard suffix",
			pattern:        "/health/**",
			expectedRegex:  "^/health(?:/.*)?$",
			shouldMatch:    []string{"/health", "/health/", "/health/liveness", "/health/readiness/full"},
			shouldNotMatch: []string{"/healthz", "/other"},
		},
		{
			name:           "Multiple wildcards",
			pattern:        "/i18n/languages/*/translations/ns/*/keys/*/resolve",
			expectedRegex:  "^/i18n/languages/[^/]+/translations/ns/[^/]+/keys/[^/]+/resolve$",
			shouldMatch:    []string{"/i18n/languages/en/translations/ns/common/keys/btn.submit/resolve"},
			shouldNotMatch: []string{"/i18n/languages/en/translations/ns/common/keys/btn.submit/extra"},
		},
		{
			name:           "Special characters escaping",
			pattern:        "/api/v1.0/user",
			expectedRegex:  "^/api/v1\\.0/user$",
			shouldMatch:    []string{"/api/v1.0/user"},
			shouldNotMatch: []string{"/api/v1a0/user"},
		},
	}

	tests = append(tests, []struct {
		name           string
		pattern        string
		expectedRegex  string
		shouldMatch    []string
		shouldNotMatch []string
	}{
		{
			name:           "Invalid middle globstar (skipped)",
			pattern:        "/api/**/users",
			expectedRegex:  "", // Skipped
			shouldMatch:    nil,
			shouldNotMatch: nil,
		},
		{
			name:           "Multiple globstars (skipped)",
			pattern:        "/api/**/users/**",
			expectedRegex:  "", // Skipped
			shouldMatch:    nil,
			shouldNotMatch: nil,
		},
	}...)

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			compiled, err := compilePathPatterns([]string{tt.pattern})

			if tt.expectedRegex == "" {
				// Invalid pattern. Error is expected.
				assert.Error(suite.T(), err)
				assert.Nil(suite.T(), compiled)
			} else {
				assert.NoError(suite.T(), err)
				assert.Len(suite.T(), compiled, 1)
				regex := compiled[0]
				assert.Equal(suite.T(), tt.expectedRegex, regex.String())

				for _, matchPath := range tt.shouldMatch {
					assert.True(suite.T(), regex.MatchString(matchPath), "Should match: %s", matchPath)
				}

				for _, mismatchPath := range tt.shouldNotMatch {
					assert.False(suite.T(), regex.MatchString(mismatchPath), "Should not match: %s", mismatchPath)
				}
			}
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
	assert.Equal(suite.T(), req.Context(), ctx)

	userID := GetSubject(ctx)
	assert.Empty(suite.T(), userID)
}

// Test skipSecurity flag allows requests to proceed even with auth failures
func (suite *SecurityServiceTestSuite) TestProcess_SkipSecurity_AllowsAccessWithoutToken() {
	_ = os.Setenv("THUNDER_SKIP_SECURITY", "true")
	defer func() { _ = os.Unsetenv("THUNDER_SKIP_SECURITY") }()

	service, err := newSecurityService([]AuthenticatorInterface{suite.mockAuth1}, testPublicPaths)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)

	// No authenticator can handle the request (no token)
	suite.mockAuth1.On("CanHandle", req).Return(false)

	ctx, err := service.Process(req)

	// Should succeed without authentication
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), ctx)
	assert.Equal(suite.T(), req.Context(), ctx)
}

// setupSkipSecurityTest is a helper function to set up skip security tests
func (suite *SecurityServiceTestSuite) setupSkipSecurityTest() (*http.Request, SecurityServiceInterface) {
	_ = os.Setenv("THUNDER_SKIP_SECURITY", "true")
	suite.T().Cleanup(func() { _ = os.Unsetenv("THUNDER_SKIP_SECURITY") })

	service, err := newSecurityService([]AuthenticatorInterface{suite.mockAuth1}, testPublicPaths)
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

	service, err := newSecurityService([]AuthenticatorInterface{suite.mockAuth1}, testPublicPaths)
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
	assert.Equal(suite.T(), req.Context(), ctx)

	// Context should not be enriched
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

	// Context should still be enriched (authentication succeeded).
	userID := GetSubject(ctx)
	assert.Equal(suite.T(), "user123", userID)
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

// Test getRequiredPermissions returns the expected permission set for various paths.
func (suite *SecurityServiceTestSuite) TestGetRequiredPermissions() {
	tests := []struct {
		name     string
		path     string
		method   string
		expected []string
	}{
		{
			name:     "Protected API path requires system",
			path:     "/users",
			method:   http.MethodGet,
			expected: []string{"system"},
		},
		{
			name:     "Applications endpoint requires system",
			path:     "/applications",
			method:   http.MethodGet,
			expected: []string{"system"},
		},
		{
			name:     "Root path requires system",
			path:     "/",
			method:   http.MethodGet,
			expected: []string{"system"},
		},
		{
			name:     "User self-service /users/me needs no special permission",
			path:     "/users/me",
			method:   http.MethodGet,
			expected: []string{},
		},
		{
			name:     "User self-service /users/me123 needs system permission",
			path:     "/users/me123",
			method:   http.MethodGet,
			expected: []string{"system"},
		},
		{
			name:     "User credentials update needs no special permission",
			path:     "/users/me/update-credentials",
			method:   http.MethodPut,
			expected: []string{},
		},
		{
			name:     "Passkey registration start needs no special permission",
			path:     "/register/passkey/start",
			method:   http.MethodPost,
			expected: []string{},
		},
		{
			name:     "Passkey registration finish needs no special permission",
			path:     "/register/passkey/finish",
			method:   http.MethodPost,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			result := suite.service.getRequiredPermissions(req)
			assert.Equal(suite.T(), tt.expected, result)
		})
	}
}

// Test hasAnyPermission covers all edge cases.
func (suite *SecurityServiceTestSuite) TestHasAnyPermission() {
	tests := []struct {
		name                string
		userPermissions     []string
		requiredPermissions []string
		expected            bool
	}{
		{
			name:                "User has one of the required permissions",
			userPermissions:     []string{"users:read", "groups:manage"},
			requiredPermissions: []string{"users:read", "users:write"},
			expected:            true,
		},
		{
			name:                "User has all required permissions",
			userPermissions:     []string{"users:read", "users:write"},
			requiredPermissions: []string{"users:read", "users:write"},
			expected:            true,
		},
		{
			name:                "User has none of the required permissions",
			userPermissions:     []string{"groups:manage"},
			requiredPermissions: []string{"users:read", "users:write"},
			expected:            false,
		},
		{
			name:                "No required permissions — always allowed",
			userPermissions:     []string{"users:read"},
			requiredPermissions: []string{},
			expected:            true,
		},
		{
			name:                "Empty user permissions",
			userPermissions:     []string{},
			requiredPermissions: []string{"users:read"},
			expected:            false,
		},
		{
			name:                "Both empty — allowed",
			userPermissions:     []string{},
			requiredPermissions: []string{},
			expected:            true,
		},
		{
			name:                "System permission satisfies system requirement",
			userPermissions:     []string{"system"},
			requiredPermissions: []string{"system"},
			expected:            true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := hasAnyPermission(tt.userPermissions, tt.requiredPermissions)
			assert.Equal(suite.T(), tt.expected, result)
		})
	}
}
