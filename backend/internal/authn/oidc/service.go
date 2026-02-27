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

// Package oidc implements an authentication service for authenticating via an OIDC-based identity provider.
package oidc

import (
	"strings"

	authncm "github.com/asgardeo/thunder/internal/authn/common"
	authnoauth "github.com/asgardeo/thunder/internal/authn/oauth"
	"github.com/asgardeo/thunder/internal/idp"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	httpservice "github.com/asgardeo/thunder/internal/system/http"
	"github.com/asgardeo/thunder/internal/system/jose/jwt"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/userprovider"
)

const (
	loggerComponentName = "OIDCAuthnService"
)

// OIDCAuthnCoreServiceInterface defines the core contract for OIDC based authenticator services.
type OIDCAuthnCoreServiceInterface interface {
	authnoauth.OAuthAuthnCoreServiceInterface
	ValidateIDToken(idpID, idToken string) *serviceerror.ServiceError
	GetIDTokenClaims(idToken string) (map[string]interface{}, *serviceerror.ServiceError)
}

// OIDCAuthnServiceInterface defines the contract for OIDC based authenticator services.
type OIDCAuthnServiceInterface interface {
	OIDCAuthnCoreServiceInterface
	ValidateTokenResponse(idpID string, tokenResp *authnoauth.TokenResponse,
		validateIDToken bool) *serviceerror.ServiceError
}

// oidcAuthnService is the default implementation of OIDCAuthnServiceInterface.
type oidcAuthnService struct {
	internal   authnoauth.OAuthAuthnServiceInterface
	jwtService jwt.JWTServiceInterface
	logger     *log.Logger
}

// newOIDCAuthnService creates a new instance of OIDC authenticator service.
func newOIDCAuthnService(httpClient httpservice.HTTPClientInterface,
	idpSvc idp.IDPServiceInterface, userProvider userprovider.UserProviderInterface,
	jwtSvc jwt.JWTServiceInterface) OIDCAuthnServiceInterface {
	internal := authnoauth.NewOAuthAuthnService(httpClient, idpSvc, userProvider)

	service := &oidcAuthnService{
		internal:   internal,
		jwtService: jwtSvc,
		logger:     log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName)),
	}
	authncm.RegisterAuthenticator(service.getMetadata())

	return service
}

// NewOIDCAuthnService creates a new instance of OIDC authenticator service.
// [Deprecated: use dependency injection to get the instance instead].
// TODO: Should be removed when executors are migrated to di pattern.
func NewOIDCAuthnService(httpClient httpservice.HTTPClientInterface,
	idpSvc idp.IDPServiceInterface, userProvider userprovider.UserProviderInterface,
	jwtSvc jwt.JWTServiceInterface) OIDCAuthnServiceInterface {
	return newOIDCAuthnService(httpClient, idpSvc, userProvider, jwtSvc)
}

// GetOAuthClientConfig retrieves the OAuth client configuration for the given identity provider ID.
func (s *oidcAuthnService) GetOAuthClientConfig(idpID string) (
	*authnoauth.OAuthClientConfig, *serviceerror.ServiceError) {
	return s.internal.GetOAuthClientConfig(idpID)
}

// BuildAuthorizeURL constructs the authorization request URL for the external identity provider.
func (s *oidcAuthnService) BuildAuthorizeURL(idpID string) (string, *serviceerror.ServiceError) {
	return s.internal.BuildAuthorizeURL(idpID)
}

// ExchangeCodeForToken exchanges the authorization code for a token with the external identity provider
// and validates the token response if validateResponse is true.
func (s *oidcAuthnService) ExchangeCodeForToken(idpID, code string, validateResponse bool) (
	*authnoauth.TokenResponse, *serviceerror.ServiceError) {
	tokenResp, svcErr := s.internal.ExchangeCodeForToken(idpID, code, false)
	if svcErr != nil {
		return nil, svcErr
	}

	if validateResponse {
		svcErr = s.ValidateTokenResponse(idpID, tokenResp, true)
		if svcErr != nil {
			return nil, svcErr
		}
	}

	return tokenResp, nil
}

// ValidateTokenResponse validates the token response returned by the identity provider.
// ExchangeCodeForToken method calls this method to validate the token response if validateResponse is set
// to true. Hence generally you may not need to call this method explicitly.
func (s *oidcAuthnService) ValidateTokenResponse(idpID string, tokenResp *authnoauth.TokenResponse,
	validateIDToken bool) *serviceerror.ServiceError {
	logger := s.logger
	logger.Debug("Validating token response")

	if tokenResp == nil {
		logger.Debug("Empty token response received from identity provider")
		return &authnoauth.ErrorInvalidTokenResponse
	}
	if tokenResp.AccessToken == "" {
		logger.Debug("Access token is empty in the token response")
		return &authnoauth.ErrorInvalidTokenResponse
	}
	if tokenResp.IDToken == "" {
		logger.Debug("ID token is empty in the token response")
		return &authnoauth.ErrorInvalidTokenResponse
	}

	if validateIDToken {
		svcErr := s.ValidateIDToken(idpID, tokenResp.IDToken)
		if svcErr != nil {
			return svcErr
		}
	}

	return nil
}

// ValidateIDToken validates the ID token from the OIDC provider.
// ValidateTokenResponse method calls this method to validate the token response if validateIDToken is set
// to true. Hence generally you may not need to call this method explicitly if ExchangeCodeForToken method
// is called with validateResponse set to true.
func (s *oidcAuthnService) ValidateIDToken(idpID, idToken string) *serviceerror.ServiceError {
	logger := s.logger.With(log.String("idpId", idpID))
	logger.Debug("Validating ID token")

	if strings.TrimSpace(idToken) == "" {
		logger.Debug("ID token is empty")
		return &ErrorInvalidIDToken
	}

	oAuthClientConfig, svcErr := s.GetOAuthClientConfig(idpID)
	if svcErr != nil {
		return svcErr
	}

	// Validate ID token signature using JWKS endpoint if available
	if oAuthClientConfig.OAuthEndpoints.JwksEndpoint != "" {
		err := s.jwtService.VerifyJWTWithJWKS(idToken, oAuthClientConfig.OAuthEndpoints.JwksEndpoint, "", "")
		if err != nil {
			logger.Debug("ID token signature validation failed", log.String("error", err.Error))
			return &ErrorInvalidIDTokenSignature
		}
	} else {
		logger.Debug("Skipping ID token signature validation as JWKS endpoint is not configured")
	}

	// TODO: Should mandate ID token validation when the support is available through a IDP configuration.
	//  Additionally should switch the validation method based on the configurations.
	//  For now, assumes validation is only performed if the JWKS endpoint is available.

	return nil
}

// GetIDTokenClaims extracts and returns the claims from the ID token.
func (s *oidcAuthnService) GetIDTokenClaims(idToken string) (
	map[string]interface{}, *serviceerror.ServiceError) {
	logger := s.logger
	logger.Debug("Extracting claims from ID token")

	if strings.TrimSpace(idToken) == "" {
		logger.Debug("ID token is empty")
		return nil, &ErrorInvalidIDToken
	}

	claims, err := jwt.DecodeJWTPayload(idToken)
	if err != nil {
		logger.Debug("Failed to decode ID token payload", log.Error(err))
		return nil, &ErrorInvalidIDToken
	}

	return claims, nil
}

// FetchUserInfo retrieves user information from the external identity provider.
func (s *oidcAuthnService) FetchUserInfo(idpID, accessToken string) (
	map[string]interface{}, *serviceerror.ServiceError) {
	return s.internal.FetchUserInfo(idpID, accessToken)
}

// GetInternalUser retrieves the internal user based on the external subject identifier.
func (s *oidcAuthnService) GetInternalUser(sub string) (*userprovider.User, *serviceerror.ServiceError) {
	return s.internal.GetInternalUser(sub)
}

// getMetadata returns the authenticator metadata for OIDC authenticator.
func (s *oidcAuthnService) getMetadata() authncm.AuthenticatorMeta {
	return authncm.AuthenticatorMeta{
		Name:          authncm.AuthenticatorOIDC,
		Factors:       []authncm.AuthenticationFactor{authncm.FactorKnowledge},
		AssociatedIDP: idp.IDPTypeOIDC,
	}
}
