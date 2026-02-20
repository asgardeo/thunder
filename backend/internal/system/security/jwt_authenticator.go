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
	"github.com/asgardeo/thunder/internal/system/jose/jwt"
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

	// Step 4: Extract subject information and build SecurityContext
	subject := ""
	if sub, ok := attributes["sub"].(string); ok && sub != "" {
		subject = sub
	}

	ouID := extractAttribute(attributes, "ou_id")

	// Step 5: Extract scopes from JWT claims
	scopes := extractScopes(attributes)

	// Create immutable SecurityContext
	return newSecurityContext(subject, ouID, token, scopes, attributes), nil
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

// extractScopes extracts permissions from JWT claims.
// Permissions can be in "scope" (string with space-separated values), "scopes" (array) claim,
// or "authorized_permissions" (Thunder-specific) claim.
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
