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
	"time"

	"github.com/asgardeo/thunder/internal/authn/assert"
	"github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/authn/credentials"
	"github.com/asgardeo/thunder/internal/authn/github"
	"github.com/asgardeo/thunder/internal/authn/google"
	"github.com/asgardeo/thunder/internal/authn/oauth"
	"github.com/asgardeo/thunder/internal/authn/oidc"
	"github.com/asgardeo/thunder/internal/authn/otp"
	"github.com/asgardeo/thunder/internal/authn/passkey"
	"github.com/asgardeo/thunder/internal/idp"
	notifcommon "github.com/asgardeo/thunder/internal/notification/common"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/jose/jwt"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
	"github.com/asgardeo/thunder/internal/userprovider"
)

const svcLoggerComponentName = "AuthenticationService"

// crossAllowedIDPTypes is the list of IDP types that allow cross-type authentication.
var crossAllowedIDPTypes = []idp.IDPType{idp.IDPTypeOAuth, idp.IDPTypeOIDC}

// AuthenticationServiceInterface defines the interface for the authentication service.
type AuthenticationServiceInterface interface {
	AuthenticateWithCredentials(identifiers, credentials map[string]interface{},
		skipAssertion bool, existingAssertion string) (*common.AuthenticationResponse, *serviceerror.ServiceError)
	SendOTP(senderID string, channel notifcommon.ChannelType, recipient string) (
		string, *serviceerror.ServiceError)
	VerifyOTP(sessionToken string, skipAssertion bool, existingAssertion, otp string) (
		*common.AuthenticationResponse, *serviceerror.ServiceError)
	StartIDPAuthentication(requestedType idp.IDPType, idpID string) (
		*IDPAuthInitData, *serviceerror.ServiceError)
	FinishIDPAuthentication(requestedType idp.IDPType, sessionToken string, skipAssertion bool,
		existingAssertion, code string) (*common.AuthenticationResponse, *serviceerror.ServiceError)
	// Passkey methods
	StartPasskeyRegistration(userID, relyingPartyID, relyingPartyName string,
		authSelection *PasskeyAuthenticatorSelectionDTO, attestation string) (interface{}, *serviceerror.ServiceError)
	FinishPasskeyRegistration(credential PasskeyPublicKeyCredentialDTO, sessionToken,
		credentialName string) (interface{}, *serviceerror.ServiceError)
	StartPasskeyAuthentication(userID, relyingPartyID string) (interface{}, *serviceerror.ServiceError)
	FinishPasskeyAuthentication(
		credentialID, credentialType string,
		response PasskeyCredentialResponseDTO,
		sessionToken string,
		skipAssertion bool,
		existingAssertion string,
	) (*common.AuthenticationResponse, *serviceerror.ServiceError)
}

// authenticationService is the default implementation of the AuthenticationServiceInterface.
type authenticationService struct {
	idpService             idp.IDPServiceInterface
	jwtService             jwt.JWTServiceInterface
	authAssertionGenerator assert.AuthAssertGeneratorInterface
	credentialsService     credentials.CredentialsAuthnServiceInterface
	otpService             otp.OTPAuthnServiceInterface
	oauthService           oauth.OAuthAuthnServiceInterface
	oidcService            oidc.OIDCAuthnServiceInterface
	googleService          google.GoogleOIDCAuthnServiceInterface
	githubService          github.GithubOAuthAuthnServiceInterface
	passkeyService         passkey.PasskeyServiceInterface
}

// newAuthenticationService creates a new instance of AuthenticationService.
func newAuthenticationService(
	idpSvc idp.IDPServiceInterface,
	jwtSvc jwt.JWTServiceInterface,
	authAssertGen assert.AuthAssertGeneratorInterface,
	credentialsAuthnSvc credentials.CredentialsAuthnServiceInterface,
	otpAuthnSvc otp.OTPAuthnServiceInterface,
	oauthAuthnSvc oauth.OAuthAuthnServiceInterface,
	oidcAuthnSvc oidc.OIDCAuthnServiceInterface,
	googleAuthnSvc google.GoogleOIDCAuthnServiceInterface,
	githubAuthnSvc github.GithubOAuthAuthnServiceInterface,
	passkeySvc passkey.PasskeyServiceInterface,
) AuthenticationServiceInterface {
	return &authenticationService{
		idpService:             idpSvc,
		jwtService:             jwtSvc,
		authAssertionGenerator: authAssertGen,
		credentialsService:     credentialsAuthnSvc,
		otpService:             otpAuthnSvc,
		oauthService:           oauthAuthnSvc,
		oidcService:            oidcAuthnSvc,
		googleService:          googleAuthnSvc,
		githubService:          githubAuthnSvc,
		passkeyService:         passkeySvc,
	}
}

// AuthenticateWithCredentials authenticates a user using credentials.
func (as *authenticationService) AuthenticateWithCredentials(identifiers, credentials map[string]interface{},
	skipAssertion bool, existingAssertion string) (
	*common.AuthenticationResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, svcLoggerComponentName))
	logger.Debug("Authenticating with credentials")

	authenticateResp, svcErr := as.credentialsService.Authenticate(identifiers, credentials, nil)
	if svcErr != nil {
		return nil, svcErr
	}

	if authenticateResp == nil {
		logger.Error("Credentials authenticate response is nil")
		return nil, &serviceerror.InternalServerError
	}

	requestedAttributes := []string{}
	for _, attr := range authenticateResp.AvailableAttributes {
		requestedAttributes = append(requestedAttributes, attr.Name)
	}

	userAttributes, svcErr := as.credentialsService.GetAttributes(authenticateResp.Token, requestedAttributes, nil)
	if svcErr != nil {
		return nil, svcErr
	}

	if userAttributes == nil {
		logger.Error("Credentials get attributes response is nil")
		return nil, &serviceerror.InternalServerError
	}

	authResponse := &common.AuthenticationResponse{
		ID:               userAttributes.UserID,
		Type:             userAttributes.UserType,
		OrganizationUnit: userAttributes.OrganizationUnitID,
	}

	// Generate assertion if not skipped
	if !skipAssertion {
		authenticatedUser := &userprovider.User{
			UserID:             userAttributes.UserID,
			UserType:           userAttributes.UserType,
			OrganizationUnitID: userAttributes.OrganizationUnitID,
			Attributes:         userAttributes.Attributes,
		}
		svcErr = as.validateAndAppendAuthAssertion(authResponse, authenticatedUser, common.AuthenticatorCredentials,
			existingAssertion, logger)
		if svcErr != nil {
			return nil, svcErr
		}
	}

	return authResponse, nil
}

// SendOTP sends an OTP to the specified recipient for authentication.
func (as *authenticationService) SendOTP(senderID string, channel notifcommon.ChannelType,
	recipient string) (string, *serviceerror.ServiceError) {
	return as.otpService.SendOTP(senderID, channel, recipient)
}

// VerifyOTP verifies an OTP and returns the authenticated user.
func (as *authenticationService) VerifyOTP(sessionToken string, skipAssertion bool,
	existingAssertion, otpCode string) (*common.AuthenticationResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, svcLoggerComponentName))
	logger.Debug("Verifying OTP for authentication")

	user, svcErr := as.otpService.VerifyOTP(sessionToken, otpCode)
	if svcErr != nil {
		return nil, svcErr
	}

	authResponse := &common.AuthenticationResponse{
		ID:               user.UserID,
		Type:             user.UserType,
		OrganizationUnit: user.OrganizationUnitID,
	}

	// Generate assertion if not skipped
	if !skipAssertion {
		svcErr = as.validateAndAppendAuthAssertion(authResponse, user, common.AuthenticatorSMSOTP,
			existingAssertion, logger)
		if svcErr != nil {
			return nil, svcErr
		}
	}

	return authResponse, nil
}

// StartIDPAuthentication initiates authentication against an IDP.
func (as *authenticationService) StartIDPAuthentication(requestedType idp.IDPType, idpID string) (
	*IDPAuthInitData, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, svcLoggerComponentName))
	logger.Debug("Starting IDP authentication", log.String("idpId", idpID))

	if strings.TrimSpace(idpID) == "" {
		return nil, &common.ErrorInvalidIDPID
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
		return nil, &serviceerror.InternalServerError
	}

	if svcErr != nil {
		return nil, svcErr
	}

	// Generate session token
	sessionToken, err := as.createSessionToken(idpID, identityProvider.Type)
	if err != nil {
		logger.Error("Failed to create session token", log.String("idpId", idpID), log.String("error", err.Error))
		return nil, &serviceerror.InternalServerError
	}

	return &IDPAuthInitData{
		RedirectURL:  redirectURL,
		SessionToken: sessionToken,
	}, nil
}

// FinishIDPAuthentication completes authentication against an IDP.
func (as *authenticationService) FinishIDPAuthentication(requestedType idp.IDPType, sessionToken string,
	skipAssertion bool, existingAssertion, code string) (*common.AuthenticationResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, svcLoggerComponentName))
	logger.Debug("Finishing IDP authentication")

	if strings.TrimSpace(sessionToken) == "" {
		return nil, &common.ErrorEmptySessionToken
	}
	if strings.TrimSpace(code) == "" {
		return nil, &common.ErrorEmptyAuthCode
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
	var user *userprovider.User
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
		return nil, &serviceerror.InternalServerError
	}

	if svcErr != nil {
		return nil, svcErr
	}

	authResponse := &common.AuthenticationResponse{
		ID:               user.UserID,
		Type:             user.UserType,
		OrganizationUnit: user.OrganizationUnitID,
	}

	// Generate assertion if not skipped
	if !skipAssertion {
		authenticatorName, err := common.GetAuthenticatorNameForIDPType(sessionData.IDPType)
		if err != nil {
			logger.Error("Failed to get authenticator name for IDP type",
				log.String("idpType", string(sessionData.IDPType)), log.Error(err))
			return nil, &serviceerror.InternalServerError
		}

		svcErr = as.validateAndAppendAuthAssertion(authResponse, user, authenticatorName,
			existingAssertion, logger)
		if svcErr != nil {
			return nil, svcErr
		}
	}

	return authResponse, nil
}

// validateAndAppendAuthAssertion validates and appends a generated auth assertion to the authentication response.
func (as *authenticationService) validateAndAppendAuthAssertion(authResponse *common.AuthenticationResponse,
	user *userprovider.User, authenticator string, existingAssertion string,
	logger *log.Logger) *serviceerror.ServiceError {
	logger.Debug("Generating auth assertion", log.String("userId", user.UserID))

	authenticatorRef := &common.AuthenticatorReference{
		Authenticator: authenticator,
		Timestamp:     time.Now().Unix(),
	}

	// Extract existing assurance if provided and set appropriate step number
	var existingAssurance *assert.AssuranceContext
	if strings.TrimSpace(existingAssertion) != "" {
		var assertionSub string
		var svcErr *serviceerror.ServiceError
		existingAssurance, assertionSub, svcErr = as.extractClaimsFromAssertion(existingAssertion, logger)
		if svcErr != nil {
			return svcErr
		}

		// Validate that the assertion subject matches the current user
		if assertionSub != user.UserID {
			logger.Debug("Assertion subject mismatch", log.String("assertionSub", assertionSub),
				log.String("userId", user.UserID))
			return &common.ErrorAssertionSubjectMismatch
		}

		if existingAssurance != nil {
			authenticatorRef.Step = len(existingAssurance.Authenticators) + 1
		} else {
			authenticatorRef.Step = 1
		}
	} else {
		authenticatorRef.Step = 1
	}

	// Prepare JWT claims
	jwtClaims := make(map[string]interface{})
	if user.UserType != "" {
		jwtClaims["userType"] = user.UserType
	}
	if user.OrganizationUnitID != "" {
		jwtClaims["organizationUnit"] = user.OrganizationUnitID
	}

	// Get authentication assertion result
	assertionResult, svcErr := as.getAssertionResult(existingAssurance, authenticatorRef)
	if svcErr != nil {
		return svcErr
	}

	if assertionResult != nil {
		jwtClaims["assurance"] = assertionResult.Context
	}

	// Generate auth assertion JWT
	jwtConfig := config.GetThunderRuntime().Config.JWT
	token, _, err := as.jwtService.GenerateJWT(user.UserID, jwtConfig.Audience, jwtConfig.Issuer,
		jwtConfig.ValidityPeriod, jwtClaims)
	if err != nil {
		logger.Error("Failed to generate auth assertion", log.String("error", err.Error))
		return &serviceerror.InternalServerError
	}

	authResponse.Assertion = token
	return nil
}

// getAssertionResult generates or updates an assertion result based on existing context.
func (as *authenticationService) getAssertionResult(existingContext *assert.AssuranceContext,
	newAuthenticator *common.AuthenticatorReference) (
	*assert.AssertionResult, *serviceerror.ServiceError) {
	var assertionResult *assert.AssertionResult
	var svcErr *serviceerror.ServiceError
	if existingContext != nil && newAuthenticator != nil {
		// Update existing assurance with new authenticator
		assertionResult, svcErr = as.authAssertionGenerator.UpdateAssertion(
			existingContext, *newAuthenticator)
	} else if newAuthenticator != nil {
		// Generate new assurance from authenticator
		assertionResult, svcErr = as.authAssertionGenerator.GenerateAssertion(
			[]common.AuthenticatorReference{*newAuthenticator})
	}

	return assertionResult, svcErr
}

// extractClaimsFromAssertion extracts assurance context and subject from an existing JWT assertion.
func (as *authenticationService) extractClaimsFromAssertion(assertion string,
	logger *log.Logger) (*assert.AssuranceContext, string, *serviceerror.ServiceError) {
	jwtConfig := config.GetThunderRuntime().Config.JWT

	if err := as.jwtService.VerifyJWT(assertion, "", jwtConfig.Issuer); err != nil {
		logger.Debug("Failed to verify JWT signature of the assertion", log.String("error", err.Error))
		return nil, "", &common.ErrorInvalidAssertion
	}

	payload, err := jwt.DecodeJWTPayload(assertion)
	if err != nil {
		logger.Debug("Failed to decode JWT assertion", log.Error(err))
		return nil, "", &common.ErrorInvalidAssertion
	}

	// Extract subject claim
	subClaim, ok := payload["sub"]
	if !ok {
		logger.Debug("No 'sub' claim found in JWT assertion")
		return nil, "", &common.ErrorInvalidAssertion
	}
	sub, ok := subClaim.(string)
	if !ok || strings.TrimSpace(sub) == "" {
		logger.Debug("Invalid 'sub' claim in JWT assertion")
		return nil, "", &common.ErrorInvalidAssertion
	}

	// Extract assurance claim
	assuranceClaim, ok := payload["assurance"]
	if !ok {
		logger.Debug("No assurance claim found in JWT assertion")
		return nil, "", &common.ErrorInvalidAssertion
	}

	// Convert assurance claim to AssuranceContext
	assuranceBytes, err := json.Marshal(assuranceClaim)
	if err != nil {
		logger.Debug("Failed to marshal assurance claim", log.Error(err))
		return nil, "", &common.ErrorInvalidAssertion
	}

	var assuranceCtx assert.AssuranceContext
	if err := json.Unmarshal(assuranceBytes, &assuranceCtx); err != nil {
		logger.Debug("Failed to unmarshal assurance claim to AssuranceContext", log.Error(err))
		return nil, "", &common.ErrorInvalidAssertion
	}

	return &assuranceCtx, sub, nil
}

// finishOAuthAuthentication handles OAuth authentication completion.
func (as *authenticationService) finishOAuthAuthentication(idpID, code string, logger *log.Logger) (
	string, *userprovider.User, *serviceerror.ServiceError) {
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
	string, *userprovider.User, *serviceerror.ServiceError) {
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
	string, *userprovider.User, *serviceerror.ServiceError) {
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
	string, *userprovider.User, *serviceerror.ServiceError) {
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
		return serviceerror.CustomServiceError(common.ErrorClientErrorWhileRetrievingIDP,
			fmt.Sprintf("An error occurred while retrieving the identity provider with ID %s: %s",
				idpID, svcErr.ErrorDescription))
	}

	logger.Error("Error occurred while retrieving IDP", log.String("idpId", idpID), log.Any("error", svcErr))
	return &serviceerror.InternalServerError
}

// validateIDPType validates that the requested IDP type matches the actual IDP type.
func (as *authenticationService) validateIDPType(requestedType, actualType idp.IDPType,
	logger *log.Logger) *serviceerror.ServiceError {
	if requestedType != "" && requestedType != actualType {
		// Allow cross-type authentication for certain types
		if slices.Contains(crossAllowedIDPTypes, requestedType) &&
			slices.Contains(crossAllowedIDPTypes, actualType) {
			return nil
		}

		logger.Debug("IDP type mismatch", log.String("requested", string(requestedType)),
			log.String("actual", string(actualType)))
		return &common.ErrorInvalidIDPType
	}

	return nil
}

// createSessionToken creates a JWT session token with authentication session data.
func (as *authenticationService) createSessionToken(idpID string, idpType idp.IDPType) (
	string, *serviceerror.ServiceError) {
	sessionData := AuthSessionData{
		IDPID:   idpID,
		IDPType: idpType,
	}
	claims := map[string]interface{}{
		"auth_data": sessionData,
	}

	jwtConfig := config.GetThunderRuntime().Config.JWT
	token, _, err := as.jwtService.GenerateJWT("auth-svc", "auth-svc", jwtConfig.Issuer, 600, claims)
	if err != nil {
		return "", err
	}

	return token, nil
}

// verifyAndDecodeSessionToken verifies the JWT signature and decodes the auth session data.
func (as *authenticationService) verifyAndDecodeSessionToken(token string, logger *log.Logger) (
	*AuthSessionData, *serviceerror.ServiceError) {
	// Verify JWT signature and claims
	jwtConfig := config.GetThunderRuntime().Config.JWT
	svcErr := as.jwtService.VerifyJWT(token, "auth-svc", jwtConfig.Issuer)
	if svcErr != nil {
		logger.Debug("Error verifying session token", log.String("error", svcErr.Error))
		return nil, &common.ErrorInvalidSessionToken
	}

	// Parse and extract authentication session data
	payload, err := jwt.DecodeJWTPayload(token)
	if err != nil {
		logger.Debug("Error decoding session token payload", log.Error(err))
		return nil, &common.ErrorInvalidSessionToken
	}

	authDataClaim, ok := payload["auth_data"]
	if !ok {
		logger.Debug("auth_data claim not found in session token")
		return nil, &common.ErrorInvalidSessionToken
	}

	authDataBytes, err := json.Marshal(authDataClaim)
	if err != nil {
		logger.Debug("Error marshaling auth_data claim", log.Error(err))
		return nil, &common.ErrorInvalidSessionToken
	}

	var sessionData AuthSessionData
	err = json.Unmarshal(authDataBytes, &sessionData)
	if err != nil {
		logger.Debug("Error marshaling auth_data claim", log.Error(err))
		return nil, &common.ErrorInvalidSessionToken
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
	return "", &common.ErrorSubClaimNotFound
}

// StartPasskeyRegistration starts the passkey registration process.
func (as *authenticationService) StartPasskeyRegistration(userID, relyingPartyID, relyingPartyName string,
	authSelection *PasskeyAuthenticatorSelectionDTO, attestation string) (interface{}, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, svcLoggerComponentName))
	logger.Debug("Starting Passkey registration")

	var passkeyAuthSelection *passkey.AuthenticatorSelection
	if authSelection != nil {
		passkeyAuthSelection = &passkey.AuthenticatorSelection{
			AuthenticatorAttachment: authSelection.AuthenticatorAttachment,
			RequireResidentKey:      authSelection.RequireResidentKey,
			ResidentKey:             authSelection.ResidentKey,
			UserVerification:        authSelection.UserVerification,
		}
	}

	req := &passkey.PasskeyRegistrationStartRequest{
		UserID:                 userID,
		RelyingPartyID:         relyingPartyID,
		RelyingPartyName:       relyingPartyName,
		AuthenticatorSelection: passkeyAuthSelection,
		Attestation:            attestation,
	}

	return as.passkeyService.StartRegistration(req)
}

// FinishPasskeyRegistration completes the passkey registration process.
func (as *authenticationService) FinishPasskeyRegistration(credential PasskeyPublicKeyCredentialDTO,
	sessionToken, credentialName string) (interface{}, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, svcLoggerComponentName))
	logger.Debug("Finishing Passkey registration")

	req := &passkey.PasskeyRegistrationFinishRequest{
		CredentialID:      credential.ID,
		CredentialType:    credential.Type,
		ClientDataJSON:    credential.Response.ClientDataJSON,
		AttestationObject: credential.Response.AttestationObject,
		SessionToken:      sessionToken,
		CredentialName:    credentialName,
	}

	return as.passkeyService.FinishRegistration(req)
}

// StartPasskeyAuthentication starts the passkey authentication process.
func (as *authenticationService) StartPasskeyAuthentication(userID, relyingPartyID string) (
	interface{}, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, svcLoggerComponentName))
	logger.Debug("Starting Passkey authentication")

	req := &passkey.PasskeyAuthenticationStartRequest{
		UserID:         userID,
		RelyingPartyID: relyingPartyID,
	}
	return as.passkeyService.StartAuthentication(req)
}

// FinishPasskeyAuthentication completes the passkey authentication process.
func (as *authenticationService) FinishPasskeyAuthentication(credentialID, credentialType string,
	response PasskeyCredentialResponseDTO, sessionToken string, skipAssertion bool,
	existingAssertion string) (*common.AuthenticationResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, svcLoggerComponentName))
	logger.Debug("Finishing Passkey authentication")

	// Get authentication response from passkey service
	req := &passkey.PasskeyAuthenticationFinishRequest{
		CredentialID:      credentialID,
		CredentialType:    credentialType,
		ClientDataJSON:    response.ClientDataJSON,
		AuthenticatorData: response.AuthenticatorData,
		Signature:         response.Signature,
		UserHandle:        response.UserHandle,
		SessionToken:      sessionToken,
	}
	authResponse, svcErr := as.passkeyService.FinishAuthentication(req)
	if svcErr != nil {
		return nil, svcErr
	}

	// Generate assertion if not skipped
	if !skipAssertion {
		// Create user object from authResponse for assertion generation
		userForAssertion := &userprovider.User{
			UserID:             authResponse.ID,
			UserType:           authResponse.Type,
			OrganizationUnitID: authResponse.OrganizationUnit,
			Attributes:         nil, // Attributes not needed for assertion generation from passkey finish
		}

		svcErr = as.validateAndAppendAuthAssertion(authResponse, userForAssertion, common.AuthenticatorPasskey,
			existingAssertion, logger)
		if svcErr != nil {
			return nil, svcErr
		}
	}

	return authResponse, nil
}
