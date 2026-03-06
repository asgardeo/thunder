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

// Package google implements an authentication service for authenticating via Google OIDC.
package google

import (
	"strings"
	"time"

	"github.com/asgardeo/thunder/internal/authn/common"
	authnoauth "github.com/asgardeo/thunder/internal/authn/oauth"
	authnoidc "github.com/asgardeo/thunder/internal/authn/oidc"
	"github.com/asgardeo/thunder/internal/idp"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	syshttp "github.com/asgardeo/thunder/internal/system/http"
	"github.com/asgardeo/thunder/internal/system/jose/jwt"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/userprovider"
)

const (
	loggerComponentName = "GoogleOIDCAuthnService"
)

// GoogleOIDCAuthnServiceInterface defines the contract for Google OIDC based authenticator services.
type GoogleOIDCAuthnServiceInterface interface {
	authnoidc.OIDCAuthnCoreServiceInterface
}

// googleOIDCAuthnService is the default implementation of GoogleOIDCAuthnServiceInterface.
type googleOIDCAuthnService struct {
	internal   authnoidc.OIDCAuthnServiceInterface
	jwtService jwt.JWTServiceInterface
	logger     *log.Logger
}

// newGoogleOIDCAuthnService creates a new instance of Google OIDC authenticator service.
func newGoogleOIDCAuthnService(idpSvc idp.IDPServiceInterface, userProvider userprovider.UserProviderInterface,
	jwtSvc jwt.JWTServiceInterface) GoogleOIDCAuthnServiceInterface {
	httpClient := syshttp.NewHTTPClient()
	internal := authnoidc.NewOIDCAuthnService(httpClient, idpSvc, userProvider, jwtSvc)

	service := &googleOIDCAuthnService{
		internal:   internal,
		jwtService: jwtSvc,
		logger:     log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName)),
	}
	common.RegisterAuthenticator(service.getMetadata())

	return service
}

// BuildAuthorizeURL constructs the authorization request URL for Google OIDC authentication.
func (g *googleOIDCAuthnService) BuildAuthorizeURL(idpID string) (string, *serviceerror.ServiceError) {
	return g.internal.BuildAuthorizeURL(idpID)
}

// ExchangeCodeForToken exchanges the authorization code for a token with Google
// and validates the token response if validateResponse is true.
func (g *googleOIDCAuthnService) ExchangeCodeForToken(idpID, code string, validateResponse bool) (
	*authnoauth.TokenResponse, *serviceerror.ServiceError) {
	tokenResp, svcErr := g.internal.ExchangeCodeForToken(idpID, code, false)
	if svcErr != nil {
		return nil, svcErr
	}

	if validateResponse {
		svcErr = g.ValidateTokenResponse(idpID, tokenResp)
		if svcErr != nil {
			return nil, svcErr
		}
	}

	return tokenResp, nil
}

// ValidateTokenResponse validates the token response returned from Google.
// ExchangeCodeForToken method calls this method to validate the token response if validateResponse is set
// to true. Hence generally you may not need to call this method explicitly.
func (g *googleOIDCAuthnService) ValidateTokenResponse(idpID string,
	tokenResp *authnoauth.TokenResponse) *serviceerror.ServiceError {
	svcErr := g.internal.ValidateTokenResponse(idpID, tokenResp, false)
	if svcErr != nil {
		return svcErr
	}

	return g.ValidateIDToken(idpID, tokenResp.IDToken)
}

// ValidateIDToken validates the ID token from Google with additional Google-specific validations.
// ValidateTokenResponse method calls this method to validate the token response if validateIDToken is set
// to true. Hence generally you may not need to call this method explicitly if ExchangeCodeForToken method
// is called with validateResponse set to true.
func (g *googleOIDCAuthnService) ValidateIDToken(idpID, idToken string) *serviceerror.ServiceError {
	logger := g.logger.With(log.String("idpId", idpID))
	logger.Debug("Validating ID token")

	if strings.TrimSpace(idToken) == "" {
		logger.Debug("ID token is empty")
		return &authnoidc.ErrorInvalidIDToken
	}

	// Get the OAuth client config for token validations
	oAuthClientConfig, svcErr := g.internal.GetOAuthClientConfig(idpID)
	if svcErr != nil {
		return svcErr
	}

	// Validate ID token signature using JWKS endpoint if available
	if oAuthClientConfig.OAuthEndpoints.JwksEndpoint != "" {
		err := g.jwtService.VerifyJWTSignatureWithJWKS(idToken, oAuthClientConfig.OAuthEndpoints.JwksEndpoint)
		if err != nil {
			logger.Debug("ID token signature validation failed", log.String("error", err.Error))
			return &authnoidc.ErrorInvalidIDTokenSignature
		}
	} else {
		logger.Debug("Skipping ID token signature validation as JWKS endpoint is not configured")
	}

	logger.Debug("Validating Google specific ID token claims")

	// Extract ID token claims for Google-specific validation
	claims, err := jwt.DecodeJWTPayload(idToken)
	if err != nil {
		return &authnoidc.ErrorInvalidIDToken
	}

	// Validate issuer
	iss, ok := claims["iss"].(string)
	if !ok || (iss != Issuer1 && iss != Issuer2) {
		logger.Debug("Invalid ID token issuer", log.String("issuer", iss))
		return serviceerror.CustomServiceError(authnoidc.ErrorInvalidIDToken,
			"The issuer of the ID token is not a valid Google issuer")
	}

	// Validate audience
	aud, ok := claims["aud"].(string)
	if !ok || aud != oAuthClientConfig.ClientID {
		logger.Debug("Invalid ID token audience", log.String("audience", aud),
			log.String("clientId", log.MaskString(oAuthClientConfig.ClientID)))
		return serviceerror.CustomServiceError(authnoidc.ErrorInvalidIDToken,
			"The ID token audience does not match the expected client ID")
	}

	// Get leeway from config to account for clock skew
	leeway := config.GetThunderRuntime().Config.JWT.Leeway

	// Validate expiration time
	exp, ok := claims["exp"].(float64)
	if !ok {
		logger.Debug("Invalid ID token expiration claim", log.Any("exp", claims["exp"]))
		return serviceerror.CustomServiceError(authnoidc.ErrorInvalidIDToken,
			"The ID token expiration claim is missing or invalid")
	}
	if time.Now().Unix() >= int64(exp)+leeway {
		logger.Debug("ID token has expired", log.Int("exp", int(exp)))
		return serviceerror.CustomServiceError(authnoidc.ErrorInvalidIDToken, "The ID token has expired")
	}

	// Check if token was issued in the future (with leeway for clock skew)
	iat, ok := claims["iat"].(float64)
	if !ok {
		logger.Debug("Invalid ID token issued-at claim", log.Any("iat", claims["iat"]))
		return serviceerror.CustomServiceError(authnoidc.ErrorInvalidIDToken,
			"The ID token issued-at (iat) claim is missing or invalid")
	}
	if time.Now().Unix() < int64(iat)-leeway {
		logger.Debug("ID token was issued in the future", log.Int("iat", int(iat)))
		return serviceerror.CustomServiceError(authnoidc.ErrorInvalidIDToken,
			"The ID token was issued in the future")
	}

	// Check for specific domain if configured in additional params
	if hd, found := claims["hd"]; found {
		logger.Debug("hd claim found in ID token")
		if domain, exists := oAuthClientConfig.AdditionalParams["hd"]; exists && domain != "" {
			logger.Debug("Validating hosted domain (hd) claim")
			if hdStr, ok := hd.(string); !ok || hdStr != domain {
				logger.Debug("Invalid hosted domain (hd) claim", log.String("hd", hdStr),
					log.String("expectedDomain", domain))
				return serviceerror.CustomServiceError(authnoidc.ErrorInvalidIDToken,
					"The ID token is not from the expected hosted domain: "+domain)
			}
		}
	}

	return nil
}

// GetIDTokenClaims extracts and returns the claims from the Google ID token.
func (g *googleOIDCAuthnService) GetIDTokenClaims(idToken string) (
	map[string]interface{}, *serviceerror.ServiceError) {
	return g.internal.GetIDTokenClaims(idToken)
}

// FetchUserInfo retrieves user information from Google, ensuring email resolution if necessary.
func (g *googleOIDCAuthnService) FetchUserInfo(idpID, accessToken string) (
	map[string]interface{}, *serviceerror.ServiceError) {
	return g.internal.FetchUserInfo(idpID, accessToken)
}

// GetInternalUser retrieves the internal user based on the external subject identifier.
func (g *googleOIDCAuthnService) GetInternalUser(sub string) (*userprovider.User, *serviceerror.ServiceError) {
	return g.internal.GetInternalUser(sub)
}

// GetOAuthClientConfig retrieves and validates the OAuth client configuration for the given identity provider ID.
func (g *googleOIDCAuthnService) GetOAuthClientConfig(idpID string) (
	*authnoauth.OAuthClientConfig, *serviceerror.ServiceError) {
	return g.internal.GetOAuthClientConfig(idpID)
}

// getMetadata returns the authenticator metadata for Google OIDC authenticator.
func (g *googleOIDCAuthnService) getMetadata() common.AuthenticatorMeta {
	return common.AuthenticatorMeta{
		Name:          common.AuthenticatorGoogle,
		Factors:       []common.AuthenticationFactor{common.FactorKnowledge},
		AssociatedIDP: idp.IDPTypeGoogle,
	}
}
