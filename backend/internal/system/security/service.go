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
	"net/http"
	"strings"
)

// SecurityServiceInterface defines the contract for security processing services.
type SecurityServiceInterface interface {
	Process(r *http.Request) (context.Context, error)
}

// securityService orchestrates authentication and authorization for HTTP requests.
type securityService struct {
	authenticators []AuthenticatorInterface
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

// isPublicPath checks if the given path is a public endpoint that doesn't require authentication.
func (s *securityService) isPublicPath(path string) bool {
	publicPaths := []string{
		"/health/",
		"/auth/",
		"/flow/execute",
		"/oauth2/",
		"/.well-known/openid-configuration",
		"/.well-known/oauth-authorization-server",
		"/gate/",    // Gate application (login UI)
		"/develop/", // Develop application
		"/error",
		"/branding/resolve",
	}

	for _, publicPath := range publicPaths {
		if strings.HasPrefix(path, publicPath) {
			return true
		}
		// Exact match for paths without trailing slash
		if path == strings.TrimSuffix(publicPath, "/") {
			return true
		}
	}

	return false
}
