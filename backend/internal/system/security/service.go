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

// Package security provides authentication and authorization for Thunder APIs.
package security

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/asgardeo/thunder/internal/system/log"
)

const loggerComponentName = "SecurityService"

// SecurityServiceInterface defines the contract for security processing services.
type SecurityServiceInterface interface {
	Process(r *http.Request) (context.Context, error)
}

// securityService orchestrates authentication and authorization for HTTP requests.
type securityService struct {
	authenticators []AuthenticatorInterface
	logger         *log.Logger
	compiledPaths  []*regexp.Regexp
	skipSecurity   bool
}

// newSecurityService creates a new instance of the security service.
//
// Parameters:
//   - authenticators: A slice of AuthenticatorInterface implementations to handle request authentication.
//   - publicPaths: A slice of string patterns representing paths that are exempt from authentication.
//
// Returns:
//   - *securityService: A pointer to the created securityService instance.
//   - error: An error if any of the provided public paths are invalid and cannot be compiled.
func newSecurityService(authenticators []AuthenticatorInterface, publicPaths []string) (*securityService, error) {
	compiledPaths, err := compilePathPatterns(publicPaths)
	if err != nil {
		return nil, err
	}

	// Check if security enforcement should be skipped via environment variable
	skipSecurity := os.Getenv("THUNDER_SKIP_SECURITY") == "true"

	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if skipSecurity {
		logger.Warn("============================================================")
		logger.Warn("|       WARNING: SECURITY ENFORCEMENT DISABLED             |")
		logger.Warn("|                                                          |")
		logger.Warn("|        THUNDER_SKIP_SECURITY is set to 'true'            |")
		logger.Warn("|  This is NOT RECOMMENDED for production environments!    |")
		logger.Warn("| Endpoints accessible without auth, but tokens processed  |")
		logger.Warn("|                                                          |")
		logger.Warn("============================================================")
	}

	return &securityService{
		authenticators: authenticators,
		logger:         logger,
		compiledPaths:  compiledPaths,
		skipSecurity:   skipSecurity,
	}, nil
}

// Process handles the complete security flow: authentication and authorization.
// Returns an enriched context on success, or an error if authentication or authorization fails.
func (s *securityService) Process(r *http.Request) (context.Context, error) {
	isPublic := s.isPublicPath(r.URL.Path)

	// Check if the request is options (CORS preflight)
	if r.Method == http.MethodOptions {
		return r.Context(), nil
	}

	// Find an authenticator that can process this request
	var authenticator AuthenticatorInterface
	for _, a := range s.authenticators {
		if a.CanHandle(r) {
			authenticator = a
			break
		}
	}

	// If no authenticator found
	if authenticator == nil {
		return s.handleAuthError(r.Context(), r.URL.Path, errNoHandlerFound, isPublic, s.skipSecurity)
	}

	// Authenticate the request
	securityCtx, err := authenticator.Authenticate(r)
	if err != nil {
		return s.handleAuthError(r.Context(), r.URL.Path, err, isPublic, s.skipSecurity)
	}

	// Add authentication context to request context if available
	ctx := r.Context()
	if securityCtx != nil {
		ctx = withSecurityContext(ctx, securityCtx)
	}

	// Authorize the authenticated principal based on the permissions carried in the security context.
	if err := s.authorize(r.WithContext(ctx)); err != nil {
		return s.handleAuthError(ctx, r.URL.Path, err, isPublic, s.skipSecurity)
	}

	return ctx, nil
}

// authorize checks whether the permissions stored in the request context satisfy
// the requirements for the requested path.
func (s *securityService) authorize(r *http.Request) error {
	permissions := GetPermissions(r.Context())
	required := s.getRequiredPermissions(r)

	if len(required) > 0 && !hasAnyPermission(permissions, required) {
		return errInsufficientPermissions
	}

	return nil
}

// getRequiredPermissions returns the permissions that a caller must hold to access
// the requested path. An empty slice means the path is open to any authenticated user.
func (s *securityService) getRequiredPermissions(r *http.Request) []string {
	// User self-service endpoints are accessible to any authenticated user.
	if r.URL.Path == "/users/me" || strings.HasPrefix(r.URL.Path, "/users/me/") {
		return []string{}
	}

	// Passkey registration endpoints are accessible to any authenticated user.
	if strings.HasPrefix(r.URL.Path, "/register/passkey/") {
		return []string{}
	}

	// All other endpoints require the "system" permission by default.
	return []string{"system"}
}

// hasAnyPermission reports whether userPermissions contains at least one entry
// from requiredPermissions. An empty required list is always satisfied.
func hasAnyPermission(userPermissions, requiredPermissions []string) bool {
	if len(requiredPermissions) == 0 {
		return true
	}

	permissionSet := make(map[string]bool, len(userPermissions))
	for _, p := range userPermissions {
		permissionSet[p] = true
	}

	for _, required := range requiredPermissions {
		if permissionSet[required] {
			return true
		}
	}

	return false
}

// isPublicPath checks if the given request path matches any of the configured public path patterns.
func (s *securityService) isPublicPath(requestPath string) bool {
	if len(requestPath) > maxPublicPathLength {
		s.logger.Warn("Path length exceeds maximum allowed length",
			log.Int("limit", maxPublicPathLength),
			log.Int("length", len(requestPath)))
		return false
	}

	for _, regex := range s.compiledPaths {
		if regex.MatchString(requestPath) {
			return true
		}
	}

	return false
}

// compilePathPatterns compiles the path patterns into regular expressions safely.
// It returns an error if any pattern is invalid.
func compilePathPatterns(patterns []string) ([]*regexp.Regexp, error) {
	compiled := make([]*regexp.Regexp, 0, len(patterns))

	for _, pattern := range patterns {
		var regexPattern string

		// Check for recursive wildcard usage
		if strings.Contains(pattern, "**") {
			// Ensure "**" is only used as a suffix "/**"
			if !strings.HasSuffix(pattern, "/**") {
				return nil,
					fmt.Errorf("invalid pattern: recursive wildcard '**' is only allowed as a suffix: %s", pattern)
			}

			// Ensure "**" appears only once
			if strings.Count(pattern, "**") > 1 {
				return nil, fmt.Errorf("invalid pattern: recursive wildcard '**' can only appear once: %s", pattern)
			}

			base := strings.TrimSuffix(pattern, "/**")
			baseRegex := regexp.QuoteMeta(base)
			baseRegex = strings.ReplaceAll(baseRegex, "\\*", "[^/]+")
			regexPattern = "^" + baseRegex + "(?:/.*)?$"
		} else {
			// Normal pattern (no recursive wildcards)
			regexPattern = regexp.QuoteMeta(pattern)
			regexPattern = strings.ReplaceAll(regexPattern, "\\*", "[^/]+")
			regexPattern = "^" + regexPattern + "$"
		}

		re, err := regexp.Compile(regexPattern)
		if err != nil {
			return nil, fmt.Errorf("error compiling public path regex for pattern %s: %w", pattern, err)
		}

		compiled = append(compiled, re)
	}

	return compiled, nil
}

// handleAuthError handles authentication/authorization errors based on whether
// the path is public or security is skipped.
func (s *securityService) handleAuthError(
	ctx context.Context,
	path string,
	err error,
	isPublic bool,
	skipSecurity bool,
) (context.Context, error) {
	if isPublic {
		return ctx, nil
	}

	if skipSecurity {
		s.logger.Debug(
			"Proceeding without authentication/authorization enforcement as skipSecurity is enabled",
			log.Error(err),
			log.String("path", path))
		return ctx, nil
	}

	return nil, err
}
