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

package tokenservice

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/user"
)

// ParseScopes parses a space-separated scope string into a slice of scope strings.
func ParseScopes(scopeString string) []string {
	trimmed := strings.TrimSpace(scopeString)
	if trimmed == "" {
		return []string{}
	}

	// Split by space and filter out empty strings
	parts := strings.Split(trimmed, " ")
	scopes := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			scopes = append(scopes, part)
		}
	}
	return scopes
}

// JoinScopes joins a slice of scope strings into a space-separated string.
func JoinScopes(scopes []string) string {
	return strings.Join(scopes, " ")
}

// resolveTokenConfig resolves the token configuration from the OAuth app or falls back to global config.
// Both access and ID tokens use the same OAuth-level issuer.
func resolveTokenConfig(oauthApp *appmodel.OAuthAppConfigProcessedDTO, tokenType TokenType) *TokenConfig {
	conf := config.GetThunderRuntime().Config

	tokenConfig := &TokenConfig{
		Issuer:         conf.JWT.Issuer,
		ValidityPeriod: conf.JWT.ValidityPeriod,
	}

	// Use OAuth-level issuer for all token types if app config is available
	if oauthApp != nil && oauthApp.Token != nil && oauthApp.Token.Issuer != "" {
		tokenConfig.Issuer = oauthApp.Token.Issuer
	}

	// Override with token-type specific configuration if available
	switch tokenType {
	case TokenTypeAccess:
		if oauthApp != nil && oauthApp.Token != nil && oauthApp.Token.AccessToken != nil {
			if oauthApp.Token.AccessToken.ValidityPeriod > 0 {
				tokenConfig.ValidityPeriod = oauthApp.Token.AccessToken.ValidityPeriod
			}
		}
	case TokenTypeID:
		if oauthApp != nil && oauthApp.Token != nil && oauthApp.Token.IDToken != nil {
			if oauthApp.Token.IDToken.ValidityPeriod > 0 {
				tokenConfig.ValidityPeriod = oauthApp.Token.IDToken.ValidityPeriod
			}
		}
	case TokenTypeRefresh:
		if conf.OAuth.RefreshToken.ValidityPeriod > 0 {
			tokenConfig.ValidityPeriod = conf.OAuth.RefreshToken.ValidityPeriod
		}
	}

	return tokenConfig
}

// extractStringClaim safely extracts a string claim from a claims map.
func extractStringClaim(claims map[string]interface{}, key string) (string, error) {
	value, ok := claims[key]
	if !ok {
		return "", fmt.Errorf("missing claim: %s", key)
	}

	strValue, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("claim %s is not a string", key)
	}

	return strValue, nil
}

// extractInt64Claim safely extracts an int64 claim from a claims map.
func extractInt64Claim(claims map[string]interface{}, key string) (int64, error) {
	value, ok := claims[key]
	if !ok {
		return 0, fmt.Errorf("missing claim: %s", key)
	}

	// JSON numbers are decoded as float64
	switch v := value.(type) {
	case float64:
		return int64(v), nil
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	default:
		return 0, fmt.Errorf("claim %s is not a number", key)
	}
}

// extractScopesFromClaims extracts and parses scopes from a claims map.
func extractScopesFromClaims(claims map[string]interface{}, isAuthAssertion bool) []string {
	scopeValue, ok := claims["scope"]
	if ok {
		scopeString, ok := scopeValue.(string)
		if ok && scopeString != "" {
			return ParseScopes(scopeString)
		}
	}

	// This allows auth assertions with authorized_permissions to be used in token exchange
	if isAuthAssertion {
		authorizedPermsValue, ok := claims["authorized_permissions"]
		if ok {
			authorizedPermsString, ok := authorizedPermsValue.(string)
			if ok && authorizedPermsString != "" {
				return ParseScopes(authorizedPermsString)
			}
		}
	}

	return []string{}
}

// DetermineAudience determines the audience for a token based on priority.
func DetermineAudience(audience, resource, tokenAud, defaultAudience string) string {
	if audience != "" {
		return audience
	}
	if resource != "" {
		return resource
	}
	if tokenAud != "" {
		return tokenAud
	}
	return defaultAudience
}

// getStandardJWTClaims returns the standard JWT claims that should be excluded from user attributes.
func getStandardJWTClaims() map[string]bool {
	return map[string]bool{
		"sub":       true,
		"iss":       true,
		"aud":       true,
		"exp":       true,
		"nbf":       true,
		"iat":       true,
		"jti":       true,
		"scope":     true,
		"client_id": true,
		"act":       true,
	}
}

// ExtractUserAttributes extracts user attributes from JWT claims by filtering out standard claims.
func ExtractUserAttributes(claims map[string]interface{}) map[string]interface{} {
	standardClaims := getStandardJWTClaims()

	userAttributes := make(map[string]interface{})
	for key, value := range claims {
		if !standardClaims[key] {
			userAttributes[key] = value
		}
	}

	return userAttributes
}

// getValidIssuers collects all valid/trusted issuers for the given OAuth application.
func getValidIssuers(oauthApp *appmodel.OAuthAppConfigProcessedDTO) map[string]bool {
	validIssuers := make(map[string]bool)

	tokenConfig := resolveTokenConfig(oauthApp, TokenTypeAccess)
	validIssuers[tokenConfig.Issuer] = true

	// TODO: Add support for external issuers
	return validIssuers
}

// validateIssuer validates that a token issuer is trusted by checking against configured issuers.
func validateIssuer(issuer string, oauthApp *appmodel.OAuthAppConfigProcessedDTO) error {
	validIssuers := getValidIssuers(oauthApp)
	if !validIssuers[issuer] {
		return fmt.Errorf("token issuer '%s' is not supported", issuer)
	}
	return nil
}

// BuildOIDCClaimsFromScopes builds OIDC claims based on scopes, user attributes, and app configuration.
func BuildOIDCClaimsFromScopes(
	scopes []string,
	userAttributes map[string]interface{},
	oauthApp *appmodel.OAuthAppConfigProcessedDTO,
) map[string]interface{} {
	claims := make(map[string]interface{})

	// Extract allowed user attributes and scope-to-claims mapping from ID token config
	var allowedUserAttributes []string
	var scopeClaimsMapping map[string][]string
	if oauthApp != nil && oauthApp.Token != nil && oauthApp.Token.IDToken != nil {
		allowedUserAttributes = oauthApp.Token.IDToken.UserAttributes
		scopeClaimsMapping = oauthApp.Token.IDToken.ScopeClaims
	}

	if len(allowedUserAttributes) == 0 || userAttributes == nil || len(scopes) == 0 {
		return claims
	}

	// For each scope, get the claims associated with that scope
	for _, scope := range scopes {
		var scopeClaims []string

		// Check app-specific scope claims first
		if scopeClaimsMapping != nil {
			if appClaims, exists := scopeClaimsMapping[scope]; exists {
				scopeClaims = appClaims
			}
		}

		// Fall back to standard OIDC scopes if no app-specific mapping
		if scopeClaims == nil {
			if standardScope, exists := constants.StandardOIDCScopes[scope]; exists {
				scopeClaims = standardScope.Claims
			}
		}

		// Add claims if they're in user attributes and allowed in config
		for _, claim := range scopeClaims {
			if slices.Contains(allowedUserAttributes, claim) {
				if value, ok := userAttributes[claim]; ok && value != nil {
					claims[claim] = value
				}
			}
		}
	}

	return claims
}

// FetchUserAttributesAndGroups fetches user attributes and groups from the user service.
// It returns user attributes, user groups, and an error if any.
// Callers should log errors with their own context.
func FetchUserAttributesAndGroups(userService user.UserServiceInterface, userID string,
	includeGroups bool) (map[string]interface{}, []string, error) {
	user, svcErr := userService.GetUser(context.TODO(), userID)
	if svcErr != nil {
		return nil, nil, fmt.Errorf("failed to fetch user: %s", svcErr.Error)
	}

	var attrs map[string]interface{}
	if user.Attributes != nil {
		if err := json.Unmarshal(user.Attributes, &attrs); err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal user attributes: %w", err)
		}
	}

	if !includeGroups {
		return attrs, []string{}, nil
	}

	groups, svcErr := userService.GetUserGroups(context.TODO(), userID, constants.DefaultGroupListLimit, 0)
	if svcErr != nil {
		return nil, nil, fmt.Errorf("failed to fetch user groups: %s", svcErr.Error)
	}

	userGroups := make([]string, 0, len(groups.Groups))
	for _, group := range groups.Groups {
		userGroups = append(userGroups, group.Name)
	}

	return attrs, userGroups, nil
}
