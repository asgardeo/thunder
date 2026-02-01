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
	"fmt"
	"time"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/jwt"
)

// TokenValidatorInterface defines the interface for validating tokens.
type TokenValidatorInterface interface {
	ValidateRefreshToken(token string, clientID string) (*RefreshTokenClaims, error)
	ValidateSubjectToken(token string, oauthApp *appmodel.OAuthAppConfigProcessedDTO) (*SubjectTokenClaims, error)
}

// TokenValidator implements TokenValidatorInterface.
type tokenValidator struct {
	jwtService jwt.JWTServiceInterface
}

// NewTokenValidator creates a new TokenValidator instance.
func newTokenValidator(jwtService jwt.JWTServiceInterface) TokenValidatorInterface {
	return &tokenValidator{
		jwtService: jwtService,
	}
}

// ValidateRefreshToken validates a refresh token and extracts the claims.
func (tv *tokenValidator) ValidateRefreshToken(token string, clientID string) (*RefreshTokenClaims, error) {
	if err := tv.jwtService.VerifyJWT(token, "", ""); err != nil {
		return nil, fmt.Errorf("invalid refresh token: %v", err)
	}

	claims, err := jwt.DecodeJWTPayload(token)
	if err != nil {
		return nil, fmt.Errorf("failed to decode refresh token: %w", err)
	}

	if err := tv.validateOAuth2RefreshClaims(claims, clientID); err != nil {
		return nil, err
	}

	// Extract claims
	sub, _ := extractStringClaim(claims, "access_token_sub")
	aud, _ := extractStringClaim(claims, "access_token_aud")
	grantType, _ := extractStringClaim(claims, "grant_type")
	iat, _ := extractInt64Claim(claims, "iat")
	scopes := extractScopesFromClaims(claims, false)

	// Extract user attributes if present
	var userAttributes map[string]interface{}
	if userAttrs, ok := claims["access_token_user_attributes"].(map[string]interface{}); ok {
		userAttributes = userAttrs
	}

	// Extract user type and organizational unit details if present
	return &RefreshTokenClaims{
		Sub:            sub,
		Aud:            aud,
		GrantType:      grantType,
		Scopes:         scopes,
		UserAttributes: userAttributes,
		Iat:            iat,
	}, nil
}

// ValidateSubjectToken validates a subject token for token exchange.
func (tv *tokenValidator) ValidateSubjectToken(
	token string,
	oauthApp *appmodel.OAuthAppConfigProcessedDTO,
) (*SubjectTokenClaims, error) {
	claims, err := jwt.DecodeJWTPayload(token)
	if err != nil {
		return nil, fmt.Errorf("failed to decode token: %w", err)
	}

	iss, err := extractStringClaim(claims, "iss")
	if err != nil {
		return nil, fmt.Errorf("subject token is missing 'iss' claim: %w", err)
	}

	if err := validateIssuer(iss, oauthApp); err != nil {
		return nil, err
	}

	if err := tv.verifyTokenSignatureByIssuer(token, iss, oauthApp); err != nil {
		return nil, fmt.Errorf("invalid subject token signature: %w", err)
	}

	sub, err := extractStringClaim(claims, "sub")
	if err != nil {
		return nil, fmt.Errorf("missing or invalid 'sub' claim: %w", err)
	}

	// Validate time-based claims
	if err := tv.validateTimeClaims(claims); err != nil {
		return nil, err
	}

	// Determine if this is an auth assertion
	isAuthAssertion := tv.isAuthAssertion(claims)

	// Extract and validate audience claim
	var aud string
	if isAuthAssertion {
		// For auth assertions, audience is required and must match config default or requesting client's app_id
		aud, err = extractStringClaim(claims, "aud")
		if err != nil {
			return nil, fmt.Errorf("auth assertion is missing 'aud' claim: %w", err)
		}

		defaultAudience := config.GetThunderRuntime().Config.JWT.Audience
		clientAppID := oauthApp.AppID

		if aud != defaultAudience && aud != clientAppID {
			return nil, fmt.Errorf("auth assertion audience mismatch")
		}
	} else {
		aud, _ = extractStringClaim(claims, "aud")
	}

	// Extract scopes
	scopes := extractScopesFromClaims(claims, isAuthAssertion)

	// Extract user attributes
	userAttributes := ExtractUserAttributes(claims)

	// Extract nested act claim if present
	var nestedAct map[string]interface{}
	if actClaim, ok := claims["act"].(map[string]interface{}); ok {
		nestedAct = actClaim
	}

	return &SubjectTokenClaims{
		Sub:            sub,
		Iss:            iss,
		Aud:            aud,
		Scopes:         scopes,
		UserAttributes: userAttributes,
		NestedAct:      nestedAct,
	}, nil
}

// verifyTokenSignatureByIssuer verifies JWT signature using issuer-specific verification method.
func (tv *tokenValidator) verifyTokenSignatureByIssuer(
	token string,
	issuer string,
	oauthApp *appmodel.OAuthAppConfigProcessedDTO,
) error {
	issuers := getValidIssuers(oauthApp)
	if issuers[issuer] {
		svcErr := tv.jwtService.VerifyJWTSignature(token)
		if svcErr != nil {
			return fmt.Errorf("failed to verify token signature: %v", svcErr)
		}
		return nil
	}

	// TODO: Implement JWKS-based verification for external federated issuers
	return fmt.Errorf("no verification method configured for issuer: %s", issuer)
}

// validateTimeClaims validates time-based claims (exp, nbf).
func (tv *tokenValidator) validateTimeClaims(claims map[string]interface{}) error {
	// Get leeway from config to account for clock skew
	leeway := config.GetThunderRuntime().Config.JWT.Leeway
	now := time.Now().Unix()

	exp, err := extractInt64Claim(claims, "exp")
	if err != nil {
		return fmt.Errorf("missing or invalid 'exp' claim: %w", err)
	}
	if now >= exp+leeway {
		return fmt.Errorf("token has expired")
	}

	nbf, err := extractInt64Claim(claims, "nbf")
	if err == nil {
		if now < nbf-leeway {
			return fmt.Errorf("token not yet valid")
		}
	}

	return nil
}

// validateOAuth2RefreshClaims validates OAuth2-specific refresh token claims.
func (tv *tokenValidator) validateOAuth2RefreshClaims(claims map[string]interface{}, clientID string) error {
	sub, err := extractStringClaim(claims, "sub")
	if err != nil {
		return fmt.Errorf("missing or invalid 'sub' claim: %w", err)
	}

	if sub != clientID {
		return fmt.Errorf("refresh token does not belong to the requesting client")
	}

	// Validate required refresh token claims
	if _, err := extractStringClaim(claims, "access_token_sub"); err != nil {
		return fmt.Errorf("missing or invalid 'access_token_sub' claim: %w", err)
	}

	if _, err := extractStringClaim(claims, "access_token_aud"); err != nil {
		return fmt.Errorf("missing or invalid 'access_token_aud' claim: %w", err)
	}

	if _, err := extractStringClaim(claims, "grant_type"); err != nil {
		return fmt.Errorf("missing or invalid 'grant_type' claim: %w", err)
	}

	return nil
}

// isAuthAssertion determines if a JWT token is an auth assertion.
func (tv *tokenValidator) isAuthAssertion(
	claims map[string]interface{},
) bool {
	// TODO: Revisit this once we have a proper way to determine if a token is an auth assertion.
	if _, hasAssurance := claims["assurance"]; hasAssurance {
		return true
	}

	return false
}
