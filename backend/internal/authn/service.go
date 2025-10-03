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

// Package authn implements the authentication service for authenticating users against different methods.
package authn

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/authn/github"
	"github.com/asgardeo/thunder/internal/authn/google"
	"github.com/asgardeo/thunder/internal/authn/oauth"
	"github.com/asgardeo/thunder/internal/authn/oidc"
	"github.com/asgardeo/thunder/internal/idp"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/jwt"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
	usermodel "github.com/asgardeo/thunder/internal/user/model"
)

const svcLoggerComponentName = "AuthenticationService"

// crossAllowedTypes is the list of IDP types that allow cross-type authentication.
var crossAllowedTypes = []idp.IDPType{idp.IDPTypeOAuth, idp.IDPTypeOIDC}

// AuthenticationServiceInterface defines the interface for the authentication service.
type AuthenticationServiceInterface interface {
	StartAuthentication(requestedType idp.IDPType, idpID string) (
		*IDPAuthInitData, *serviceerror.ServiceError)
	FinishAuthentication(requestedType idp.IDPType, sessionToken, code string) (
		*common.AuthenticationResponse, *serviceerror.ServiceError)
}

// authenticationService is the default implementation of the AuthenticationServiceInterface.
type authenticationService struct {
	idpService    idp.IDPServiceInterface
	jwtService    jwt.JWTServiceInterface
	oauthService  oauth.OAuthAuthnServiceInterface
	oidcService   oidc.OIDCAuthnServiceInterface
	googleService google.GoogleOIDCAuthnServiceInterface
	githubService github.GithubOAuthAuthnServiceInterface
}

// NewAuthenticationService creates a new instance of AuthenticationService.
func NewAuthenticationService() AuthenticationServiceInterface {
	return &authenticationService{
		idpService:    idp.NewIDPService(),
		jwtService:    jwt.GetJWTService(),
		oauthService:  oauth.NewOAuthAuthnService(nil, nil, oauth.OAuthEndpoints{}),
		oidcService:   oidc.NewOIDCAuthnService(nil, nil),
		googleService: google.NewGoogleOIDCAuthnService(nil),
		githubService: github.NewGithubOAuthAuthnService(nil, nil),
	}
}

// StartAuthentication initiates authentication against an IDP.
func (as *authenticationService) StartAuthentication(requestedType idp.IDPType, idpID string) (
	*IDPAuthInitData, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, svcLoggerComponentName))
	logger.Debug("Starting IDP authentication", log.String("idpId", idpID))

	if strings.TrimSpace(idpID) == "" {
		return nil, &ErrorInvalidIDPID
	}

	identityProvider, svcErr := as.idpService.GetIdentityProvider(idpID)
	if svcErr != nil {
		return nil, as.handleIDPServiceError(idpID, svcErr, logger)
	}

	if svcErr := as.validateIDPType(requestedType, identityProvider.Type, logger); svcErr != nil {
		return nil, svcErr
	}

	// Route to appropriate service based on IDP type
	var redirectURL string
	switch identityProvider.Type {
	case idp.IDPTypeOAuth:
		redirectURL, svcErr = as.oauthService.BuildAuthorizeURL(idpID)
	case idp.IDPTypeOIDC:
		redirectURL, svcErr = as.oidcService.BuildAuthorizeURL(idpID)
	case idp.IDPTypeGoogle:
		redirectURL, svcErr = as.googleService.BuildAuthorizeURL(idpID)
	case idp.IDPTypeGitHub:
		redirectURL, svcErr = as.githubService.BuildAuthorizeURL(idpID)
	default:
		logger.Error("Unsupported IDP type", log.String("idpId", idpID),
			log.String("type", string(identityProvider.Type)))
		return nil, &ErrorInternalServerError
	}

	if svcErr != nil {
		return nil, svcErr
	}

	// Generate session token
	sessionToken, err := as.createSessionToken(idpID, identityProvider.Type)
	if err != nil {
		logger.Error("Failed to create session token", log.String("idpId", idpID), log.Error(err))
		return nil, &ErrorInternalServerError
	}

	return &IDPAuthInitData{
		RedirectURL:  redirectURL,
		SessionToken: sessionToken,
	}, nil
}

// FinishAuthentication completes authentication against an IDP.
func (as *authenticationService) FinishAuthentication(requestedType idp.IDPType, sessionToken, code string) (
	*common.AuthenticationResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, svcLoggerComponentName))
	logger.Debug("Finishing IDP authentication")

	if strings.TrimSpace(sessionToken) == "" {
		return nil, &ErrorEmptySessionToken
	}
	if strings.TrimSpace(code) == "" {
		return nil, &ErrorEmptyAuthCode
	}

	// Verify and decode session token
	sessionData, svcErr := as.verifyAndDecodeSessionToken(sessionToken, logger)
	if svcErr != nil {
		return nil, svcErr
	}

	if svcErr := as.validateIDPType(requestedType, sessionData.IDPType, logger); svcErr != nil {
		return nil, svcErr
	}

	// Route to appropriate service based on IDP type from session
	var user *usermodel.User
	switch sessionData.IDPType {
	case idp.IDPTypeOAuth:
		_, user, svcErr = as.finishOAuthAuthentication(sessionData.IDPID, code, logger)
	case idp.IDPTypeOIDC:
		_, user, svcErr = as.finishOIDCAuthentication(sessionData.IDPID, code, logger)
	case idp.IDPTypeGoogle:
		_, user, svcErr = as.finishGoogleAuthentication(sessionData.IDPID, code, logger)
	case idp.IDPTypeGitHub:
		_, user, svcErr = as.finishGithubAuthentication(sessionData.IDPID, code, logger)
	default:
		logger.Error("Unsupported IDP type in session", log.String("idpId", sessionData.IDPID),
			log.String("type", string(sessionData.IDPType)))
		return nil, &ErrorInternalServerError
	}

	if svcErr != nil {
		return nil, svcErr
	}

	return &common.AuthenticationResponse{
		ID:               user.ID,
		Type:             user.Type,
		OrganizationUnit: user.OrganizationUnit,
	}, nil
}

// finishOAuthAuthentication handles OAuth authentication completion.
func (as *authenticationService) finishOAuthAuthentication(idpID, code string, logger *log.Logger) (
	string, *usermodel.User, *serviceerror.ServiceError) {
	tokenResp, svcErr := as.oauthService.ExchangeCodeForToken(idpID, code, true)
	if svcErr != nil {
		return "", nil, svcErr
	}

	userInfo, svcErr := as.oauthService.FetchUserInfo(idpID, tokenResp.AccessToken)
	if svcErr != nil {
		return "", nil, svcErr
	}

	sub, svcErr := as.getSubClaim(userInfo, logger)
	if svcErr != nil {
		return "", nil, svcErr
	}

	user, svcErr := as.oauthService.GetInternalUser(sub)
	if svcErr != nil {
		return "", nil, svcErr
	}

	return sub, user, nil
}

// finishOIDCAuthentication handles OIDC authentication completion.
func (as *authenticationService) finishOIDCAuthentication(idpID, code string, logger *log.Logger) (
	string, *usermodel.User, *serviceerror.ServiceError) {
	tokenResp, svcErr := as.oidcService.ExchangeCodeForToken(idpID, code, true)
	if svcErr != nil {
		return "", nil, svcErr
	}

	claims, svcErr := as.oidcService.GetIDTokenClaims(tokenResp.IDToken)
	if svcErr != nil {
		return "", nil, svcErr
	}

	// TODO: Fetch user info if more claims are needed. Implement when the IDP requested attribute
	//  support is added

	sub, svcErr := as.getSubClaim(claims, logger)
	if svcErr != nil {
		return "", nil, svcErr
	}

	user, svcErr := as.oidcService.GetInternalUser(sub)
	if svcErr != nil {
		return "", nil, svcErr
	}

	return sub, user, nil
}

// finishGoogleAuthentication handles Google authentication completion.
func (as *authenticationService) finishGoogleAuthentication(idpID, code string, logger *log.Logger) (
	string, *usermodel.User, *serviceerror.ServiceError) {
	tokenResp, svcErr := as.googleService.ExchangeCodeForToken(idpID, code, true)
	if svcErr != nil {
		return "", nil, svcErr
	}

	claims, svcErr := as.googleService.GetIDTokenClaims(tokenResp.IDToken)
	if svcErr != nil {
		return "", nil, svcErr
	}

	// TODO: Fetch user info if more claims are needed. Implement when the IDP requested attribute
	//  support is added

	sub, svcErr := as.getSubClaim(claims, logger)
	if svcErr != nil {
		return "", nil, svcErr
	}

	user, svcErr := as.googleService.GetInternalUser(sub)
	if svcErr != nil {
		return "", nil, svcErr
	}

	return sub, user, nil
}

// finishGithubAuthentication handles GitHub authentication completion.
func (as *authenticationService) finishGithubAuthentication(idpID, code string, logger *log.Logger) (
	string, *usermodel.User, *serviceerror.ServiceError) {
	tokenResp, svcErr := as.githubService.ExchangeCodeForToken(idpID, code, true)
	if svcErr != nil {
		return "", nil, svcErr
	}

	userInfo, svcErr := as.githubService.FetchUserInfo(idpID, tokenResp.AccessToken)
	if svcErr != nil {
		return "", nil, svcErr
	}

	sub, svcErr := as.getSubClaim(userInfo, logger)
	if svcErr != nil {
		return "", nil, svcErr
	}

	user, svcErr := as.githubService.GetInternalUser(sub)
	if svcErr != nil {
		return "", nil, svcErr
	}

	return sub, user, nil
}

// handleIDPServiceError handles errors from IDP service.
func (as *authenticationService) handleIDPServiceError(idpID string, svcErr *serviceerror.ServiceError,
	logger *log.Logger) *serviceerror.ServiceError {
	if svcErr.Type == serviceerror.ClientErrorType {
		return serviceerror.CustomServiceError(ErrorClientErrorWhileRetrievingIDP,
			fmt.Sprintf("An error occurred while retrieving the identity provider with ID %s: %s",
				idpID, svcErr.ErrorDescription))
	}

	logger.Error("Error occurred while retrieving IDP", log.String("idpId", idpID), log.Any("error", svcErr))
	return &ErrorInternalServerError
}

// validateIDPType validates that the requested IDP type matches the actual IDP type.
func (as *authenticationService) validateIDPType(requestedType, actualType idp.IDPType,
	logger *log.Logger) *serviceerror.ServiceError {
	if requestedType != "" && requestedType != actualType {
		// Allow cross-type authentication for certain types
		if slices.Contains(crossAllowedTypes, requestedType) &&
			slices.Contains(crossAllowedTypes, actualType) {
			return nil
		}

		logger.Debug("IDP type mismatch", log.String("requested", string(requestedType)),
			log.String("actual", string(actualType)))
		return &ErrorInvalidIDPType
	}

	return nil
}

// createSessionToken creates a JWT session token with authentication session data.
func (as *authenticationService) createSessionToken(idpID string, idpType idp.IDPType) (string, error) {
	sessionData := AuthSessionData{
		IDPID:   idpID,
		IDPType: idpType,
	}
	claims := map[string]interface{}{
		"auth_data": sessionData,
	}

	jwtConfig := config.GetThunderRuntime().Config.OAuth.JWT
	token, _, err := as.jwtService.GenerateJWT("auth-svc", "auth-svc", jwtConfig.Issuer, 600, claims)
	if err != nil {
		return "", err
	}

	return token, nil
}

// verifyAndDecodeSessionToken verifies the JWT signature and decodes the auth session data.
func (as *authenticationService) verifyAndDecodeSessionToken(token string, logger *log.Logger) (
	*AuthSessionData, *serviceerror.ServiceError) {
	publicKey := as.jwtService.GetPublicKey()
	if publicKey == nil {
		logger.Error("Error verifying session token: JWT public key is not available")
		return nil, &ErrorInternalServerError
	}

	// Verify JWT signature and claims
	jwtConfig := config.GetThunderRuntime().Config.OAuth.JWT
	err := as.jwtService.VerifyJWT(token, publicKey, "auth-svc", jwtConfig.Issuer)
	if err != nil {
		logger.Debug("Error verifying session token", log.Error(err))
		return nil, &ErrorInvalidSessionToken
	}

	// Parse and extract authentication session data
	payload, err := jwt.DecodeJWTPayload(token)
	if err != nil {
		logger.Debug("Error decoding session token payload", log.Error(err))
		return nil, &ErrorInvalidSessionToken
	}

	authDataClaim, ok := payload["auth_data"]
	if !ok {
		logger.Debug("auth_data claim not found in session token")
		return nil, &ErrorInvalidSessionToken
	}

	authDataBytes, err := json.Marshal(authDataClaim)
	if err != nil {
		logger.Debug("Error marshaling auth_data claim", log.Error(err))
		return nil, &ErrorInvalidSessionToken
	}

	var sessionData AuthSessionData
	err = json.Unmarshal(authDataBytes, &sessionData)
	if err != nil {
		logger.Debug("Error marshaling auth_data claim", log.Error(err))
		return nil, &ErrorInvalidSessionToken
	}

	return &sessionData, nil
}

// getSubClaim extracts the 'sub' claim from user info claims.
func (as *authenticationService) getSubClaim(userClaims map[string]interface{}, logger *log.Logger) (
	string, *serviceerror.ServiceError) {
	sub, ok := userClaims["sub"]
	if ok && sub != nil {
		if subStr, ok := sub.(string); ok && subStr != "" {
			return subStr, nil
		}
	}

	// Try 'id' field as fallback
	id, ok := userClaims["id"]
	if ok && id != nil {
		if idStr := sysutils.ConvertInterfaceValueToString(id); idStr != "" {
			return idStr, nil
		}
	}

	logger.Debug("sub claim not found in user info claims")
	return "", &ErrorSubClaimNotFound
}
