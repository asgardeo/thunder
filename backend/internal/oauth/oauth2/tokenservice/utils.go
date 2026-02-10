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
	"github.com/asgardeo/thunder/internal/oauth/oauth2/model"
	"github.com/asgardeo/thunder/internal/ou"
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

// FetchUserAttributes fetches user attributes and merges default claims and groups into the return map.
// Callers should log errors with their own context.
func FetchUserAttributes(
	userService user.UserServiceInterface,
	ouService ou.OrganizationUnitServiceInterface,
	userID string,
	allowedClaims []string,
) (map[string]interface{}, error) {
	userData, svcErr := userService.GetUser(context.TODO(), userID)
	if svcErr != nil {
		return nil, fmt.Errorf("failed to fetch user: %s", svcErr.Error)
	}

	// Parse user attributes from JSON
	var attrs map[string]interface{}
	if userData.Attributes != nil {
		if err := json.Unmarshal(userData.Attributes, &attrs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal user attributes: %w", err)
		}
	}
	if attrs == nil {
		attrs = make(map[string]interface{})
	}

	// Helper to check if a claim should be included
	shouldInclude := func(claimName string) bool {
		if len(allowedClaims) == 0 {
			return false // Only add special claims if explicitly allowed
		}
		return slices.Contains(allowedClaims, claimName)
	}

	// Add default claim - user type
	if userData.Type != "" && shouldInclude(constants.ClaimUserType) {
		attrs[constants.ClaimUserType] = userData.Type
	}

	if userData.OrganizationUnit != "" {
		// Add default claim - ouId
		if shouldInclude(constants.ClaimOUID) {
			attrs[constants.ClaimOUID] = userData.OrganizationUnit
		}

		// Only fetch OU details if ouHandle or ouName are requested
		needsOUDetails := shouldInclude(constants.ClaimOUHandle) || shouldInclude(constants.ClaimOUName)
		if needsOUDetails && ouService != nil {
			ouDetails, ouErr := ouService.GetOrganizationUnit(userData.OrganizationUnit)
			if ouErr != nil {
				return nil, fmt.Errorf("failed to fetch organization unit details: %s", ouErr.Error)
			}

			if shouldInclude(constants.ClaimOUHandle) {
				attrs[constants.ClaimOUHandle] = ouDetails.Handle
			}
			if shouldInclude(constants.ClaimOUName) {
				attrs[constants.ClaimOUName] = ouDetails.Name
			}
		}
	}

	// Fetch and add groups if requested
	if shouldInclude(constants.UserAttributeGroups) {
		groups, svcErr := userService.GetUserGroups(context.TODO(), userID, constants.DefaultGroupListLimit, 0)
		if svcErr != nil {
			return nil, fmt.Errorf("failed to fetch user groups: %s", svcErr.Error)
		}
		if len(groups.Groups) > 0 {
			groupNames := make([]string, 0, len(groups.Groups))
			for _, group := range groups.Groups {
				groupNames = append(groupNames, group.Name)
			}
			attrs[constants.UserAttributeGroups] = groupNames
		}
	}

	return attrs, nil
}

// BuildClaims builds claims by merging scope-based claims with explicit claims request.
// Explicit claims override scope claims. Returns empty if allowedUserAttributes is not configured.
// The requestedClaims should contain only the relevant claims map (IDToken or UserInfo) for the target.
func BuildClaims(
	scopes []string,
	requestedClaims map[string]*model.IndividualClaimRequest,
	userAttributes map[string]interface{},
	scopeClaimsMapping map[string][]string,
	allowedUserAttributes []string,
) map[string]interface{} {
	result := make(map[string]interface{})

	// Check for openid scope first
	hasOpenIDScope := slices.Contains(scopes, "openid")
	if !hasOpenIDScope || userAttributes == nil {
		return result
	}

	// Build scope claims
	scopeClaims := buildClaimsFromScopes(scopes, userAttributes, scopeClaimsMapping, allowedUserAttributes)

	// Process explicit claims request if present
	if requestedClaims != nil {
		explicitClaims := buildClaimsFromRequest(requestedClaims, userAttributes, allowedUserAttributes)

		// Add scope claims that are not explicitly requested
		for k, v := range scopeClaims {
			if _, explicitlyRequested := requestedClaims[k]; !explicitlyRequested {
				result[k] = v
			}
		}

		// Add validated explicit claims (takes precedence over scope claims)
		for claimName, value := range explicitClaims {
			result[claimName] = value
		}
	} else {
		// No explicit claims request, add all scope claims
		for k, v := range scopeClaims {
			result[k] = v
		}
	}

	return result
}

// buildClaimsFromScopes builds claims from OIDC scopes based on scope-to-claims mapping.
func buildClaimsFromScopes(
	scopes []string,
	userAttributes map[string]interface{},
	scopeClaimsMapping map[string][]string,
	allowedUserAttributes []string,
) map[string]interface{} {
	claims := make(map[string]interface{})

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

// buildClaimsFromRequest builds claims from explicit claims parameter.
// Returns empty if allowedUserAttributes is not configured.
// Filters claims by availability, allowed attributes, and value/values constraints.
func buildClaimsFromRequest(
	requestedClaims map[string]*model.IndividualClaimRequest,
	userAttributes map[string]interface{},
	allowedUserAttributes []string,
) map[string]interface{} {
	result := make(map[string]interface{})

	if requestedClaims == nil || userAttributes == nil {
		return result
	}

	// Return empty if no allowed attributes configured
	if len(allowedUserAttributes) == 0 {
		return result
	}

	// Process each requested claim
	for claimName, claimReq := range requestedClaims {
		// Check if claim value is available in user attributes
		value, exists := userAttributes[claimName]
		if !exists || value == nil {
			// Per OIDC spec, it's not an error to not return a requested claim
			continue
		}

		// Check if this claim is allowed by app config
		if !slices.Contains(allowedUserAttributes, claimName) {
			continue
		}

		// Check value/values constraints if specified
		if claimReq != nil && !claimReq.MatchesValue(value) {
			// Value doesn't match the requested constraint, skip this claim
			continue
		}

		// TODO: Revisit "essential" claim handling if needed.

		result[claimName] = value
	}

	return result
}
