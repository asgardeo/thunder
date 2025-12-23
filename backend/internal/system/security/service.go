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
}

// NewSecurityService creates a new instance of the security service.
//
// Parameters:
//   - authenticators: A slice of AuthenticatorInterface implementations to handle request authentication.
//   - publicPaths: A slice of string patterns representing paths that are exempt from authentication.
//
// Returns:
//   - *securityService: A pointer to the created securityService instance.
//   - error: An error if any of the provided public paths are invalid and cannot be compiled.
func NewSecurityService(authenticators []AuthenticatorInterface, publicPaths []string) (*securityService, error) {
	compiledPaths, err := compilePathPatterns(publicPaths)
	if err != nil {
		return nil, err
	}

	return &securityService{
		authenticators: authenticators,
		logger:         log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName)),
		compiledPaths:  compiledPaths,
	}, nil
}

// Process handles the complete security flow: authentication and authorization.
// Returns an enriched context on success, or an error if authentication or authorization fails.
func (s *securityService) Process(r *http.Request) (context.Context, error) {
	// Check if the path is public (skip authentication)
	if s.isPublicPath(r.URL.Path) {
		return r.Context(), nil
	}

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

	// If no authenticator found, request is unauthorized
	if authenticator == nil {
		return nil, errNoHandlerFound
	}

	// Authenticate the request
	securityCtx, err := authenticator.Authenticate(r)
	if err != nil {
		return nil, err
	}

	// Add authentication context to request context if available
	ctx := r.Context()
	if securityCtx != nil {
		ctx = withSecurityContext(ctx, securityCtx)
	}

	// Authorize the authenticated principal
	if err := authenticator.Authorize(r.WithContext(ctx), securityCtx); err != nil {
		return nil, err
	}

	return ctx, nil
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
