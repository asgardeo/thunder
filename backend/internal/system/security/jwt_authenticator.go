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
	"net/http"
	"strings"

	"github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/jwt"
)

// jwtAuthenticator handles authentication and authorization using JWT Bearer tokens.
type jwtAuthenticator struct {
	jwtService jwt.JWTServiceInterface
}

// newJWTAuthenticator creates a new JWT authenticator.
func newJWTAuthenticator(jwtService jwt.JWTServiceInterface) *jwtAuthenticator {
	return &jwtAuthenticator{
		jwtService: jwtService,
	}
}

// CanHandle checks if the request contains a Bearer token in the Authorization header.
func (h *jwtAuthenticator) CanHandle(r *http.Request) bool {
	authHeader := r.Header.Get("Authorization")
	return strings.HasPrefix(authHeader, "Bearer ")
}

// Authenticate validates the JWT token and builds a SecurityContext.
func (h *jwtAuthenticator) Authenticate(r *http.Request) (*SecurityContext, error) {
	// Step 1: Extract Bearer token
	authHeader := r.Header.Get(constants.AuthorizationHeaderName)
	token, err := extractToken(authHeader)
	if err != nil {
		return nil, err
	}

	if token == "" {
		return nil, errInvalidToken
	}

	// Step 2: Verify JWT signature
	if err := h.jwtService.VerifyJWTSignature(token); err != nil {
		return nil, errInvalidToken
	}

	// Step 3: Decode JWT payload to extract attributes
	attributes, err := jwt.DecodeJWTPayload(token)
	if err != nil {
		return nil, errInvalidToken
	}

	// Step 4: Extract user information and build SecurityContext
	userID := ""
	if sub, ok := attributes["sub"].(string); ok && sub != "" {
		userID = sub
	}

	ouID := extractAttribute(attributes, "ou_id")
	appID := extractAttribute(attributes, "app_id")

	// Create immutable SecurityContext
	return newSecurityContext(userID, ouID, appID, token, attributes), nil
}

// Authorize verifies the authenticated user has the required scopes for the request.
func (h *jwtAuthenticator) Authorize(r *http.Request, securityCtx *SecurityContext) error {
	if securityCtx == nil {
		return errUnauthorized
	}

	ctx := r.Context()
	attributes := map[string]interface{}{}

	if scope := GetAttribute(ctx, "scope"); scope != nil {
		attributes["scope"] = scope
	}
	if scopes := GetAttribute(ctx, "scopes"); scopes != nil {
		attributes["scopes"] = scopes
	}
	if perms := GetAttribute(ctx, "authorized_permissions"); perms != nil {
		attributes["authorized_permissions"] = perms
	}

	scopes := extractScopes(attributes)
	requiredScopes := h.getRequiredScopes(r)

	if len(requiredScopes) > 0 && !hasAnyScope(scopes, requiredScopes) {
		return errInsufficientScopes
	}

	return nil
}

// extractToken extracts the Bearer token from the Authorization header.
func extractToken(authHeader string) (string, error) {
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", errMissingAuthHeader
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	token = strings.TrimSpace(token)
	return token, nil
}

// extractScopes extracts OAuth2 scopes from JWT claims.
// Scopes can be in "scope" (string with space-separated values) or "scopes" (array) claim.
func extractScopes(attributes map[string]interface{}) []string {
	// Try "scope" claim (OAuth2 standard - space-separated string)
	if scopeStr, ok := attributes["scope"].(string); ok && scopeStr != "" {
		return strings.Fields(scopeStr)
	}

	// Try "scopes" claim (array format)
	if scopesRaw, ok := attributes["scopes"]; ok {
		switch scopes := scopesRaw.(type) {
		case []interface{}:
			result := make([]string, 0, len(scopes))
			for _, s := range scopes {
				if str, ok := s.(string); ok {
					result = append(result, str)
				}
			}
			return result
		case []string:
			return scopes
		}
	}

	// Try "authorized_permissions" from the Thunder assertion
	if permsStr, ok := attributes["authorized_permissions"].(string); ok && permsStr != "" {
		return strings.Fields(permsStr)
	}

	return []string{}
}

// extractAttribute extracts a string claim from JWT claims map.
func extractAttribute(attributes map[string]interface{}, key string) string {
	if value, ok := attributes[key].(string); ok {
		return value
	}
	return ""
}

// getRequiredScopes returns the required scopes for a given route path.
func (h *jwtAuthenticator) getRequiredScopes(r *http.Request) []string {
	// User self service endpoints don't require scopes
	if strings.HasPrefix(r.URL.Path, "/users/me") {
		return []string{}
	}

	// Passkey registration endpoints don't require scopes
	if strings.HasPrefix(r.URL.Path, "/register/passkey/") {
		return []string{}
	}

	// Default required scope for other endpoints
	return []string{"system"}
}

// hasAnyScope checks if the user has at least one of the required scopes.
func hasAnyScope(userScopes, requiredScopes []string) bool {
	if len(requiredScopes) == 0 {
		return true
	}

	scopeMap := make(map[string]bool, len(userScopes))
	for _, scope := range userScopes {
		scopeMap[scope] = true
	}

	for _, required := range requiredScopes {
		if scopeMap[required] {
			return true
		}
	}

	return false
}
