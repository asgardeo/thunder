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

// Package userinfo provides functionality for the OIDC UserInfo endpoint.
package userinfo

import (
	"context"
	"crypto"
	"encoding/json"
	"io"
	"net/http"
	"slices"

	"github.com/asgardeo/thunder/internal/attributecache"
	certmodel "github.com/asgardeo/thunder/internal/cert"
	"github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	syshttp "github.com/asgardeo/thunder/internal/system/http"
	"github.com/asgardeo/thunder/internal/system/jose/jwe"
	"github.com/asgardeo/thunder/internal/system/jose/jws"
	"github.com/asgardeo/thunder/internal/system/jose/jwt"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/transaction"
	"github.com/asgardeo/thunder/pkg/inboundclient"
	inboundmodel "github.com/asgardeo/thunder/pkg/inboundclient/model"
	"github.com/asgardeo/thunder/pkg/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/pkg/oauth/oauth2/model"
	"github.com/asgardeo/thunder/pkg/oauth/oauth2/tokenservice"
	oauth2utils "github.com/asgardeo/thunder/pkg/oauth/oauth2/utils"
)

const serviceLoggerComponentName = "UserInfoService"

// userInfoServiceInterface defines the interface for OIDC UserInfo endpoint.
type userInfoServiceInterface interface {
	GetUserInfo(ctx context.Context, accessToken string) (*UserInfoResponse, *serviceerror.ServiceError)
}

// userInfoService implements the userInfoServiceInterface.
type userInfoService struct {
	jwtService        jwt.JWTServiceInterface
	jweService        jwe.JWEServiceInterface
	httpClient        syshttp.HTTPClientInterface
	tokenValidator    tokenservice.TokenValidatorInterface
	inboundClient     inboundclient.InboundClientServiceInterface
	ouService         ou.OrganizationUnitServiceInterface
	attributeCacheSvc attributecache.AttributeCacheServiceInterface
	transactioner     transaction.Transactioner
	logger            *log.Logger
}

// newUserInfoService creates a new userInfoService instance.
func newUserInfoService(
	jwtService jwt.JWTServiceInterface,
	jweService jwe.JWEServiceInterface,
	httpClient syshttp.HTTPClientInterface,
	tokenValidator tokenservice.TokenValidatorInterface,
	inboundClient inboundclient.InboundClientServiceInterface,
	ouService ou.OrganizationUnitServiceInterface,
	attributeCacheSvc attributecache.AttributeCacheServiceInterface,
	transactioner transaction.Transactioner,
) userInfoServiceInterface {
	return &userInfoService{
		jwtService:        jwtService,
		jweService:        jweService,
		httpClient:        httpClient,
		tokenValidator:    tokenValidator,
		inboundClient:     inboundClient,
		ouService:         ouService,
		attributeCacheSvc: attributeCacheSvc,
		transactioner:     transactioner,
		logger:            log.GetLogger().With(log.String(log.LoggerKeyComponentName, serviceLoggerComponentName)),
	}
}

// GetUserInfo validates the access token and returns user information based on authorized scopes.
func (s *userInfoService) GetUserInfo(
	ctx context.Context, accessToken string,
) (*UserInfoResponse, *serviceerror.ServiceError) {
	if accessToken == "" {
		return nil, &errorInvalidAccessToken
	}

	accessTokenClaims, err := s.tokenValidator.ValidateAccessToken(accessToken)
	if err != nil {
		s.logger.Debug("Failed to verify access token", log.Error(err))
		return nil, &errorInvalidAccessToken
	}
	tokenClaims := accessTokenClaims.Claims
	sub := accessTokenClaims.Sub

	if svcErr := s.validateGrantType(tokenClaims); svcErr != nil {
		return nil, svcErr
	}

	scopes := s.extractScopes(tokenClaims)

	// Validate that the 'openid' scope is present
	if svcErr := s.validateOpenIDScope(scopes); svcErr != nil {
		return nil, svcErr
	}

	oauthApp := s.getOAuthApp(ctx, tokenClaims)

	// Extract allowed user attributes
	var allowedUserAttributes []string
	if oauthApp != nil && oauthApp.UserInfo != nil {
		allowedUserAttributes = oauthApp.UserInfo.UserAttributes
	}

	attributeCacheID := ""
	if val, ok := tokenClaims["aci"].(string); ok {
		attributeCacheID = val
	}

	// Fetch user attributes with groups and default claims
	userAttributes, err := tokenservice.FetchUserAttributes(ctx, s.attributeCacheSvc,
		allowedUserAttributes, attributeCacheID)
	if err != nil {
		s.logger.Error("Failed to fetch user attributes", log.MaskedString(log.LoggerKeyUserID, sub), log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	response, svcErr := s.buildUserInfoResponse(sub, scopes, userAttributes, oauthApp, tokenClaims)
	if svcErr != nil {
		return nil, svcErr
	}

	var userInfoCfg *inboundmodel.UserInfoConfig
	var certificate *inboundmodel.Certificate
	if oauthApp != nil {
		userInfoCfg = oauthApp.UserInfo
		certificate = oauthApp.Certificate
	}

	responseType := inboundmodel.UserInfoResponseTypeJSON
	if userInfoCfg != nil {
		responseType = userInfoCfg.ResponseType
	}
	switch responseType {
	case inboundmodel.UserInfoResponseTypeNESTEDJWT:
		return s.generateNestedJWTUserInfo(ctx, sub, tokenClaims, response, userInfoCfg, certificate)
	case inboundmodel.UserInfoResponseTypeJWE:
		return s.generateJWEUserInfo(ctx, response, userInfoCfg, certificate)
	case inboundmodel.UserInfoResponseTypeJWS:
		return s.generateJWSUserInfo(sub, tokenClaims, response, userInfoCfg)
	default:
		return &UserInfoResponse{Type: inboundmodel.UserInfoResponseTypeJSON, JSONBody: response}, nil
	}
}

// resolveRPPublicKey resolves the RP's public key from the application certificate.
// It returns the public key and the kid from the matching JWK entry (empty string when absent).
// encryptionAlg is the key-management algorithm (e.g. "RSA-OAEP-256") used to filter incompatible keys.
func (s *userInfoService) resolveRPPublicKey(
	ctx context.Context, certificate *inboundmodel.Certificate, encryptionAlg string,
) (crypto.PublicKey, string, *serviceerror.ServiceError) {
	if certificate == nil || certificate.Type == "" {
		s.logger.Error("No certificate configured for userinfo encryption")
		return nil, "", &serviceerror.InternalServerError
	}

	var jwksData []byte
	switch certificate.Type {
	case certmodel.CertificateTypeJWKS:
		jwksData = []byte(certificate.Value)
	case certmodel.CertificateTypeJWKSURI:
		body, svcErr := s.fetchJWKS(ctx, certificate.Value)
		if svcErr != nil {
			return nil, "", svcErr
		}
		jwksData = body
	default:
		s.logger.Error("Unsupported certificate type for userinfo encryption",
			log.String("type", string(certificate.Type)))
		return nil, "", &serviceerror.InternalServerError
	}

	return s.parseEncryptionKeyFromJWKS(jwksData, encryptionAlg)
}

// fetchJWKS fetches the JWKS document from the given URI with SSRF protection and a 1 MB size cap.
func (s *userInfoService) fetchJWKS(ctx context.Context, jwksURI string) ([]byte, *serviceerror.ServiceError) {
	if err := syshttp.IsSSRFSafeURL(jwksURI); err != nil {
		s.logger.Error("JWKS URI is not SSRF-safe", log.String("uri", jwksURI), log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, jwksURI, nil)
	if err != nil {
		s.logger.Error("Failed to build JWKS request", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error("Failed to fetch JWKS from URI", log.String("uri", jwksURI), log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		s.logger.Error("JWKS URI returned non-200 status",
			log.String("uri", jwksURI), log.Int("statusCode", resp.StatusCode))
		return nil, &serviceerror.InternalServerError
	}
	const maxJWKSBytes = 1 << 20 // 1 MB — guards against OOM from a malicious endpoint
	limitedReader := io.LimitReader(resp.Body, maxJWKSBytes+1)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		s.logger.Error("Failed to read JWKS response body", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if len(body) > maxJWKSBytes {
		s.logger.Error("JWKS URI response exceeds 1 MB size limit", log.String("uri", jwksURI))
		return nil, &serviceerror.InternalServerError
	}
	return body, nil
}

// parseEncryptionKeyFromJWKS finds the first RSA enc key in the JWKS that matches encryptionAlg.
// Returns the public key and its kid (empty when absent in the JWK entry).
func (s *userInfoService) parseEncryptionKeyFromJWKS(
	jwksData []byte, encryptionAlg string,
) (crypto.PublicKey, string, *serviceerror.ServiceError) {
	var jwksObj struct {
		Keys []map[string]interface{} `json:"keys"`
	}
	if err := json.Unmarshal(jwksData, &jwksObj); err != nil {
		s.logger.Error("Failed to parse JWKS for userinfo encryption", log.Error(err))
		return nil, "", &serviceerror.InternalServerError
	}

	for _, key := range jwksObj.Keys {
		use, _ := key["use"].(string)
		if use != "enc" {
			continue
		}
		kty, _ := key["kty"].(string)
		if kty != "RSA" {
			continue
		}
		if keyAlg, _ := key["alg"].(string); keyAlg != "" && keyAlg != encryptionAlg {
			continue
		}
		pub, err := jws.JWKToPublicKey(key)
		if err == nil && pub != nil {
			kid, _ := key["kid"].(string)
			return pub, kid, nil
		}
	}

	s.logger.Error("No suitable encryption key found in JWKS")
	return nil, "", &serviceerror.InternalServerError
}

// generateJWEUserInfo creates an encrypted JWE UserInfo response.
func (s *userInfoService) generateJWEUserInfo(
	ctx context.Context,
	response map[string]interface{},
	cfg *inboundmodel.UserInfoConfig,
	certificate *inboundmodel.Certificate,
) (*UserInfoResponse, *serviceerror.ServiceError) {
	rpKey, rpKID, svcErr := s.resolveRPPublicKey(ctx, certificate, cfg.EncryptionAlg)
	if svcErr != nil {
		return nil, svcErr
	}

	payload, err := json.Marshal(response)
	if err != nil {
		s.logger.Error("Failed to marshal userinfo claims for JWE")
		return nil, &serviceerror.InternalServerError
	}

	compact, svcErr := s.jweService.Encrypt(
		payload, rpKey,
		jwe.KeyEncAlgorithm(cfg.EncryptionAlg),
		jwe.ContentEncAlgorithm(cfg.EncryptionEnc),
		"json",
		rpKID,
	)
	if svcErr != nil {
		s.logger.Error("Failed to encrypt userinfo JWE")
		return nil, svcErr
	}

	return &UserInfoResponse{Type: inboundmodel.UserInfoResponseTypeJWE, JWTBody: compact}, nil
}

// generateNestedJWTUserInfo creates a sign-then-encrypt Nested JWT UserInfo response.
func (s *userInfoService) generateNestedJWTUserInfo(
	ctx context.Context,
	sub string,
	tokenClaims map[string]interface{},
	response map[string]interface{},
	cfg *inboundmodel.UserInfoConfig,
	certificate *inboundmodel.Certificate,
) (*UserInfoResponse, *serviceerror.ServiceError) {
	jwsResp, svcErr := s.generateJWSUserInfo(sub, tokenClaims, response, cfg)
	if svcErr != nil {
		return nil, svcErr
	}

	rpKey, rpKID, svcErr := s.resolveRPPublicKey(ctx, certificate, cfg.EncryptionAlg)
	if svcErr != nil {
		return nil, svcErr
	}

	compact, svcErr := s.jweService.Encrypt(
		[]byte(jwsResp.JWTBody), rpKey,
		jwe.KeyEncAlgorithm(cfg.EncryptionAlg),
		jwe.ContentEncAlgorithm(cfg.EncryptionEnc),
		"JWT",
		rpKID,
	)
	if svcErr != nil {
		s.logger.Error("Failed to encrypt nested JWT userinfo JWE")
		return nil, svcErr
	}

	return &UserInfoResponse{Type: inboundmodel.UserInfoResponseTypeNESTEDJWT, JWTBody: compact}, nil
}

// generateJWSUserInfo creates a signed JWT UserInfo response
// based on the application configuration.
func (s *userInfoService) generateJWSUserInfo(
	sub string,
	tokenClaims map[string]interface{},
	response map[string]interface{},
	cfg *inboundmodel.UserInfoConfig,
) (*UserInfoResponse, *serviceerror.ServiceError) {
	clientID := ""
	if cid, ok := tokenClaims["client_id"].(string); ok {
		clientID = cid
	}

	runtime := config.GetServerRuntime()

	issuer := runtime.Config.JWT.Issuer
	validity := runtime.Config.JWT.ValidityPeriod

	response["aud"] = clientID
	signingAlg := ""
	if cfg != nil {
		signingAlg = cfg.SigningAlg
	}

	signedJWT, _, err := s.jwtService.GenerateJWT(
		sub,
		issuer,
		validity,
		response,
		jwt.TokenTypeJWT,
		signingAlg,
	)
	if err != nil {
		if err.Code == jwt.ErrorUnsupportedJWSAlgorithm.Code {
			s.logger.Error("UserInfo signing algorithm is not supported by the server key",
				log.String("alg", signingAlg), log.String("error", err.Error.DefaultValue))
		} else {
			s.logger.Error("Failed to generate signed UserInfo JWT",
				log.String("error", err.Error.DefaultValue))
		}
		return nil, &serviceerror.InternalServerError
	}

	return &UserInfoResponse{
		Type:    inboundmodel.UserInfoResponseTypeJWS,
		JWTBody: signedJWT,
	}, nil
}

// validateGrantType validates that the token was not issued using client_credentials grant.
func (s *userInfoService) validateGrantType(claims map[string]interface{}) *serviceerror.ServiceError {
	grantTypeValue, ok := claims["grant_type"]
	if !ok {
		return nil
	}

	grantTypeString, ok := grantTypeValue.(string)
	if !ok {
		return nil
	}

	if constants.GrantType(grantTypeString) == constants.GrantTypeClientCredentials {
		s.logger.Debug("UserInfo endpoint called with client_credentials grant token",
			log.String("grant_type", grantTypeString))
		return &errorClientCredentialsNotSupported
	}

	return nil
}

// extractScopes extracts scopes from the token claims.
func (s *userInfoService) extractScopes(claims map[string]interface{}) []string {
	scopeValue, ok := claims["scope"]
	if !ok {
		return nil
	}

	scopeString, ok := scopeValue.(string)
	if !ok {
		return nil
	}

	return tokenservice.ParseScopes(scopeString)
}

// validateOpenIDScope validates that the access token contains the required 'openid' scope.
func (s *userInfoService) validateOpenIDScope(scopes []string) *serviceerror.ServiceError {
	if !slices.Contains(scopes, constants.ScopeOpenID) {
		s.logger.Debug("UserInfo request missing required 'openid' scope",
			log.String("scopes", tokenservice.JoinScopes(scopes)))
		return &errorInsufficientScope
	}
	return nil
}

// getOAuthApp retrieves the OAuth client configuration if client_id is present in claims.
// Returns nil when no client_id is present, on error, or when the app is not found.
func (s *userInfoService) getOAuthApp(
	ctx context.Context, claims map[string]interface{},
) *inboundmodel.OAuthClient {
	clientID, ok := claims["client_id"].(string)
	if !ok || clientID == "" {
		return nil
	}

	app, err := s.inboundClient.GetOAuthClientByClientID(ctx, clientID)
	if err != nil || app == nil {
		return nil
	}

	return app
}

// buildUserInfoResponse builds the final UserInfo response from sub, scopes, and user attributes.
// It also processes any explicit claims request embedded in the access token.
func (s *userInfoService) buildUserInfoResponse(
	sub string,
	scopes []string,
	userAttributes map[string]interface{},
	oauthApp *inboundmodel.OAuthClient,
	tokenClaims map[string]interface{},
) (map[string]interface{}, *serviceerror.ServiceError) {
	response := map[string]interface{}{
		"sub": sub,
	}

	// Build claims from scopes and explicit claims request
	// Extract only the UserInfo claims map from the access token
	claimsRequest, svcErr := s.extractClaimsRequest(tokenClaims)
	if svcErr != nil {
		return nil, svcErr
	}
	var userInfoClaims map[string]*model.IndividualClaimRequest
	if claimsRequest != nil {
		userInfoClaims = claimsRequest.UserInfo
	}

	// Get scope claims mapping and allowed user attributes from app config
	var scopeClaimsMapping map[string][]string
	var allowedUserAttributes []string
	if oauthApp != nil {
		scopeClaimsMapping = oauthApp.ScopeClaims
		if oauthApp.UserInfo != nil && len(oauthApp.UserInfo.UserAttributes) > 0 {
			allowedUserAttributes = oauthApp.UserInfo.UserAttributes
		}
	}

	claimData := tokenservice.BuildClaims(
		scopes,
		userInfoClaims,
		userAttributes,
		scopeClaimsMapping,
		allowedUserAttributes,
	)
	for key, value := range claimData {
		response[key] = value
	}

	return response, nil
}

// extractClaimsRequest extracts the claims request from the access token if present.
func (s *userInfoService) extractClaimsRequest(
	tokenClaims map[string]interface{},
) (*model.ClaimsRequest, *serviceerror.ServiceError) {
	claimsRequestStr, ok := tokenClaims[constants.ClaimClaimsRequest].(string)
	if !ok || claimsRequestStr == "" {
		return nil, nil
	}

	claimsRequest, err := oauth2utils.ParseClaimsRequest(claimsRequestStr)
	if err != nil {
		s.logger.Error("Failed to parse claims request from access token", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	return claimsRequest, nil
}
