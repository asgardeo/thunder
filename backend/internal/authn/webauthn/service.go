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

// Package webauthn implements the WebAuthn/Passkey authentication service.
package webauthn

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/user"
)

const (
	// loggerComponentName is the component name for logging.
	loggerComponentName = "WebAuthnAuthnService"
)

// Note: WebAuthn credentials are now stored in the user table's CREDENTIALS column
// using the CredentialMap structure with credential type "passkey".

// AuthenticatorSelection represents criteria for selecting authenticators during registration.
type AuthenticatorSelection struct {
	AuthenticatorAttachment string
	RequireResidentKey      bool
	ResidentKey             string
	UserVerification        string
}

// WebAuthnRegisterStartRequest represents the request to start WebAuthn credential registration.
type WebAuthnRegisterStartRequest struct {
	UserID                 string
	RelyingPartyID         string
	RelyingPartyName       string
	AuthenticatorSelection *AuthenticatorSelection
	Attestation            string
}

// WebAuthnRegisterStartData represents the data returned when initiating WebAuthn registration.
type WebAuthnRegisterStartData struct {
	PublicKeyCredentialCreationOptions PublicKeyCredentialCreationOptions `json:"publicKeyCredentialCreationOptions"`
	SessionToken                       string                             `json:"sessionToken"`
}

// PublicKeyCredentialCreationOptions represents the options for credential creation.
type PublicKeyCredentialCreationOptions struct {
	Challenge              string                            `json:"challenge"`
	RelyingParty           protocol.RelyingPartyEntity       `json:"rp"`
	User                   protocol.UserEntity               `json:"user"`
	Parameters             []protocol.CredentialParameter    `json:"pubKeyCredParams"`
	AuthenticatorSelection protocol.AuthenticatorSelection   `json:"authenticatorSelection,omitempty"`
	Timeout                int                               `json:"timeout,omitempty"`
	CredentialExcludeList  []protocol.CredentialDescriptor   `json:"excludeCredentials,omitempty"`
	Extensions             protocol.AuthenticationExtensions `json:"extensions,omitempty"`
	Attestation            protocol.ConveyancePreference     `json:"attestation,omitempty"`
}

// WebAuthnRegisterFinishRequest represents the request to finish WebAuthn credential registration.
type WebAuthnRegisterFinishRequest struct {
	CredentialID      string
	CredentialType    string
	ClientDataJSON    string
	AttestationObject string
	SessionToken      string
	CredentialName    string
}

// WebAuthnRegisterFinishData represents the data returned after completing WebAuthn registration.
type WebAuthnRegisterFinishData struct {
	CredentialID   string
	CredentialName string
	CreatedAt      string
}

// WebAuthnStartData represents the data returned when initiating WebAuthn authentication.
type WebAuthnStartData struct {
	PublicKeyCredentialRequestOptions PublicKeyCredentialRequestOptions `json:"publicKeyCredentialRequestOptions"`
	SessionToken                      string                            `json:"sessionToken"`
}

// PublicKeyCredentialRequestOptions represents the options for credential assertion.
type PublicKeyCredentialRequestOptions struct {
	Challenge        string                               `json:"challenge"`
	Timeout          int                                  `json:"timeout,omitempty"`
	RelyingPartyID   string                               `json:"rpId,omitempty"`
	AllowCredentials []protocol.CredentialDescriptor      `json:"allowCredentials,omitempty"`
	UserVerification protocol.UserVerificationRequirement `json:"userVerification,omitempty"`
	Extensions       protocol.AuthenticationExtensions    `json:"extensions,omitempty"`
}

// CredentialDescriptor represents a WebAuthn credential descriptor.
type CredentialDescriptor struct {
	Type       string
	ID         string
	Transports []string
}

// WebAuthnFinishRequest represents the request to complete WebAuthn authentication.
type WebAuthnFinishRequest struct {
	PublicKeyCredential *protocol.ParsedCredentialAssertionData
	SessionToken        string
	SkipAssertion       bool
	Assertion           string
}

// WebAuthnAuthnServiceInterface defines the interface for WebAuthn authentication operations.
type WebAuthnAuthnServiceInterface interface {
	// Registration methods
	StartRegistration(req *WebAuthnRegisterStartRequest) (*WebAuthnRegisterStartData, *serviceerror.ServiceError)
	FinishRegistration(req *WebAuthnRegisterFinishRequest) (*WebAuthnRegisterFinishData, *serviceerror.ServiceError)

	// Authentication methods
	StartAuthentication(userID, relyingPartyID string) (*WebAuthnStartData, *serviceerror.ServiceError)
	FinishAuthentication(
		credentialID, credentialType, clientDataJSON, authenticatorData, signature, userHandle,
		sessionToken string,
		skipAssertion bool,
		existingAssertion string,
	) (*common.AuthenticationResponse, *serviceerror.ServiceError)
}

// webAuthnAuthnService is the default implementation of WebAuthnAuthnServiceInterface.
type webAuthnAuthnService struct {
	userService  user.UserServiceInterface
	sessionStore sessionStoreInterface
}

// NewWebAuthnAuthnService creates a new instance of WebAuthn authenticator service.
func NewWebAuthnAuthnService(userSvc user.UserServiceInterface) WebAuthnAuthnServiceInterface {
	if userSvc == nil {
		userSvc = user.GetUserService()
	}

	service := &webAuthnAuthnService{
		userService:  userSvc,
		sessionStore: newSessionStore(),
	}
	common.RegisterAuthenticator(service.getMetadata())

	return service
}

// StartRegistration initiates WebAuthn credential registration for a user.
func (w *webAuthnAuthnService) StartRegistration(
	req *WebAuthnRegisterStartRequest,
) (*WebAuthnRegisterStartData, *serviceerror.ServiceError) {
	// Check for nil request to prevent panic
	if req == nil {
		return nil, &ErrorInvalidFinishData
	}

	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Starting WebAuthn credential registration",
		log.String("userID", log.MaskString(req.UserID)),
		log.String("relyingPartyID", req.RelyingPartyID))

	// Validate input
	if svcErr := validateRegistrationStartRequest(req); svcErr != nil {
		return nil, svcErr
	}

	// Retrieve user information
	userInfo, svcErr := w.userService.GetUser(req.UserID)
	if svcErr != nil {
		return nil, handleUserRetrievalError(svcErr, req.UserID, logger)
	}

	// Build WebAuthn relying party display name
	rpDisplayName := req.RelyingPartyName
	if rpDisplayName == "" {
		rpDisplayName = req.RelyingPartyID
	}

	// Retrieve user's existing WebAuthn credentials from database
	credentials, err := w.getWebAuthnCredentialsFromDB(req.UserID)
	if err != nil {
		logger.Error("Failed to retrieve credentials from database", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	logger.Debug("Retrieved existing credentials",
		log.String("userID", req.UserID),
		log.Int("credentialCount", len(credentials)))

	// Create WebAuthn user adapter
	webAuthnUserAdapter := createWebAuthnUserFromUserInfo(userInfo, credentials)

	// Initialize WebAuthn library service
	webAuthnLibService, err := newWebAuthnLibraryService(req.RelyingPartyID, rpDisplayName)
	if err != nil {
		logger.Error("Failed to initialize WebAuthn library service", log.String("error", err.Error()))
		return nil, &ErrorInternalServerError
	}

	// Build registration options
	registrationOptions := buildRegistrationOptions(req)

	// Begin registration ceremony using the WebAuthn library service
	options, sessionData, err := webAuthnLibService.BeginRegistration(webAuthnUserAdapter, registrationOptions)
	if err != nil {
		logger.Error("Failed to begin WebAuthn registration", log.String("error", err.Error()))
		return nil, &ErrorInternalServerError
	}

	// Store session data in cache with TTL
	sessionToken, svcErr := w.storeSessionData(req.UserID, req.RelyingPartyID, sessionData)
	if svcErr != nil {
		logger.Error("Failed to store session data", log.String("error", svcErr.Error))
		return nil, svcErr
	}

	logger.Debug("WebAuthn credential creation options generated successfully",
		log.String("userID", log.MaskString(req.UserID)),
		log.Int("excludeCredentialsCount", len(credentials)))

	// Convert to our custom structure with properly encoded challenge
	creationOptions := PublicKeyCredentialCreationOptions{
		Challenge:              base64.RawURLEncoding.EncodeToString(options.Response.Challenge),
		RelyingParty:           options.Response.RelyingParty,
		User:                   options.Response.User,
		Parameters:             options.Response.Parameters,
		AuthenticatorSelection: options.Response.AuthenticatorSelection,
		Timeout:                options.Response.Timeout,
		CredentialExcludeList:  options.Response.CredentialExcludeList,
		Extensions:             options.Response.Extensions,
		Attestation:            options.Response.Attestation,
	}

	return &WebAuthnRegisterStartData{
		PublicKeyCredentialCreationOptions: creationOptions,
		SessionToken:                       sessionToken,
	}, nil
}

// FinishRegistration completes WebAuthn credential registration.
func (w *webAuthnAuthnService) FinishRegistration(req *WebAuthnRegisterFinishRequest) (
	*WebAuthnRegisterFinishData, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Finishing WebAuthn credential registration")

	// Validate input
	if req == nil {
		logger.Debug("Registration finish request is nil")
		return nil, &ErrorInvalidFinishData
	}
	if strings.TrimSpace(req.SessionToken) == "" {
		logger.Debug("Session token is empty")
		return nil, &ErrorEmptySessionToken
	}
	if strings.TrimSpace(req.CredentialID) == "" {
		logger.Debug("Credential ID is empty")
		return nil, &ErrorInvalidFinishData
	}
	if strings.TrimSpace(req.ClientDataJSON) == "" {
		logger.Debug("ClientDataJSON is empty")
		return nil, &ErrorInvalidFinishData
	}
	if strings.TrimSpace(req.AttestationObject) == "" {
		logger.Debug("AttestationObject is empty")
		return nil, &ErrorInvalidFinishData
	}

	// Default credential type to "public-key" if not provided
	credentialType := strings.TrimSpace(req.CredentialType)
	if credentialType == "" {
		credentialType = "public-key"
	}

	logger.Debug("Parsing attestation response",
		log.String("credentialID", req.CredentialID),
		log.String("credentialType", credentialType),
		log.Int("clientDataJSONLen", len(req.ClientDataJSON)),
		log.Int("attestationObjectLen", len(req.AttestationObject)))

	// Parse the attestation response using our custom parser
	// This ensures all required fields including the Raw field are properly populated
	parsedCredential, err := parseAttestationResponse(
		req.CredentialID,
		credentialType,
		req.ClientDataJSON,
		req.AttestationObject,
	)
	if err != nil {
		logger.Error("Failed to parse attestation response",
			log.String("error", err.Error()),
			log.String("credentialID", req.CredentialID),
			log.String("credentialType", credentialType))
		return nil, &ErrorInvalidAttestationResponse
	}

	logger.Debug("Successfully parsed attestation response",
		log.String("credentialID", parsedCredential.ID),
		log.String("credentialType", parsedCredential.Type))

	// Retrieve session data from cache
	sessionData, userID, relyingPartyID, svcErr := w.retrieveSessionData(req.SessionToken)
	if svcErr != nil {
		logger.Error("Failed to retrieve session data", log.String("error", svcErr.Error))
		return nil, svcErr
	}

	// Get user information
	userInfo, svcErr := w.userService.GetUser(userID)
	if svcErr != nil {
		logger.Error("Failed to retrieve user", log.String("error", svcErr.Error))
		return nil, &ErrorInternalServerError
	}

	// Retrieve existing credentials from database
	credentials, err := w.getWebAuthnCredentialsFromDB(userID)
	if err != nil {
		logger.Error("Failed to retrieve credentials from database", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	logger.Debug("Retrieved existing credentials for user",
		log.String("userID", userID),
		log.Int("credentialCount", len(credentials)))

	// Create WebAuthn user adapter from user info
	webAuthnUserAdapter := createWebAuthnUserFromUserInfo(userInfo, credentials)

	// Initialize WebAuthn library service with relying party configuration
	webAuthnLibService, err := newWebAuthnLibraryService(relyingPartyID, relyingPartyID)
	if err != nil {
		logger.Error("Failed to initialize WebAuthn library service", log.String("error", err.Error()))
		return nil, &ErrorInternalServerError
	}

	// Verify the credential creation response using WebAuthn library service
	// The ParsedCredentialCreationData has already been validated by the protocol package
	// Now we use the library service to perform full WebAuthn verification
	logger.Debug("Calling FinishRegistration with session data",
		log.String("sessionChallenge", sessionData.Challenge),
		log.String("sessionRPID", sessionData.RelyingPartyID),
		log.String("userID", string(sessionData.UserID)))

	logger.Debug("PublicKeyCredential details",
		log.String("credentialID", parsedCredential.ID),
		log.String("credentialType", parsedCredential.Type),
		log.Int("rawIDLength", len(parsedCredential.RawID)),
		log.String("clientDataType", string(parsedCredential.Response.CollectedClientData.Type)),
		log.String("clientDataChallenge", parsedCredential.Response.CollectedClientData.Challenge),
		log.String("clientDataOrigin", parsedCredential.Response.CollectedClientData.Origin))

	credential, err := webAuthnLibService.FinishRegistration(webAuthnUserAdapter, *sessionData, parsedCredential)
	if err != nil {
		logger.Error("Failed to verify and create credential", log.String("error", err.Error()))
		return nil, &ErrorInvalidAttestationResponse
	}

	// Generate credential name if not provided
	credentialName := req.CredentialName
	if credentialName == "" {
		credentialName = generateDefaultCredentialName()
	}

	// Encode credential ID to base64url
	credentialID := base64.StdEncoding.EncodeToString(credential.ID)

	logger.Debug("WebAuthn credential verified and created",
		log.String("credentialID", credentialID),
		log.String("aaguid", base64.StdEncoding.EncodeToString(credential.Authenticator.AAGUID)))

	// Store credential in database using user service
	if err := w.storeWebAuthnCredentialInDB(userID, credential); err != nil {
		logger.Error("Failed to store credential in database", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	logger.Debug("Stored credential in database",
		log.String("userID", userID),
		log.String("credentialID", credentialID))

	// Clear session data
	w.clearSessionData(req.SessionToken)

	logger.Debug("WebAuthn credential registration completed successfully",
		log.String("userID", log.MaskString(userID)),
		log.String("credentialID", credentialID))

	return &WebAuthnRegisterFinishData{
		CredentialID:   credentialID,
		CredentialName: credentialName,
		CreatedAt:      time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// StartAuthentication initiates WebAuthn authentication for a user.
func (w *webAuthnAuthnService) StartAuthentication(userID, relyingPartyID string) (
	*WebAuthnStartData, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Starting WebAuthn authentication",
		log.String("userID", log.MaskString(userID)),
		log.String("relyingPartyID", relyingPartyID))

	// Validate input
	if strings.TrimSpace(userID) == "" {
		return nil, &ErrorEmptyUserIdentifier
	}
	if strings.TrimSpace(relyingPartyID) == "" {
		return nil, &ErrorEmptyRelyingPartyID
	}

	// Retrieve user by userID to verify user exists
	userInfo, svcErr := w.userService.GetUser(userID)
	if svcErr != nil {
		if svcErr.Type == serviceerror.ClientErrorType {
			logger.Debug("User not found", log.String("userID", log.MaskString(userID)))
			return nil, &ErrorUserNotFound
		}
		logger.Error("Failed to retrieve user", log.String("error", svcErr.Error))
		return nil, &ErrorInternalServerError
	}

	// Retrieve user's registered WebAuthn credentials from database
	credentials, err := w.getWebAuthnCredentialsFromDB(userID)
	if err != nil {
		logger.Error("Failed to retrieve credentials from database", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	logger.Debug("Retrieved credentials for authentication",
		log.String("userID", userID),
		log.Int("credentialCount", len(credentials)))

	if len(credentials) == 0 {
		logger.Debug("No credentials found for user", log.String("userID", userID))
		return nil, &ErrorNoCredentialsFound
	}

	// Create WebAuthn user adapter from user info
	webAuthnUserAdapter := createWebAuthnUserFromUserInfo(userInfo, credentials)

	// Initialize WebAuthn library service with relying party configuration
	webAuthnLibService, err := newWebAuthnLibraryService(relyingPartyID, relyingPartyID)
	if err != nil {
		logger.Error("Failed to initialize WebAuthn library service", log.String("error", err.Error()))
		return nil, &ErrorInternalServerError
	}

	// Begin login ceremony using the WebAuthn library service
	// The library service will generate challenge and set timeout automatically
	options, sessionData, err := webAuthnLibService.BeginAuthentication(webAuthnUserAdapter)
	if err != nil {
		logger.Error("Failed to begin WebAuthn login", log.String("error", err.Error()))
		return nil, &ErrorInternalServerError
	}

	// Store session data in cache with TTL
	sessionToken, svcErr := w.storeSessionData(userID, relyingPartyID, sessionData)
	if svcErr != nil {
		logger.Error("Failed to store session data", log.String("error", svcErr.Error))
		return nil, svcErr
	}

	logger.Debug("WebAuthn authentication options generated successfully",
		log.String("userID", log.MaskString(userID)),
		log.Int("allowedCredentialsCount", len(credentials)))

	// Convert to our custom structure with properly encoded challenge
	requestOptions := PublicKeyCredentialRequestOptions{
		Challenge:        base64.RawURLEncoding.EncodeToString(options.Response.Challenge),
		Timeout:          options.Response.Timeout,
		RelyingPartyID:   options.Response.RelyingPartyID,
		AllowCredentials: options.Response.AllowedCredentials,
		UserVerification: options.Response.UserVerification,
		Extensions:       options.Response.Extensions,
	}

	return &WebAuthnStartData{
		PublicKeyCredentialRequestOptions: requestOptions,
		SessionToken:                      sessionToken,
	}, nil
}

// FinishAuthentication completes WebAuthn authentication.
func (w *webAuthnAuthnService) FinishAuthentication(credentialID, credentialType, clientDataJSON,
	authenticatorData, signature, userHandle, sessionToken string, skipAssertion bool,
	existingAssertion string) (*common.AuthenticationResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Finishing WebAuthn authentication")

	// Validate input
	if strings.TrimSpace(credentialID) == "" {
		return nil, &ErrorEmptyCredentialID
	}
	if strings.TrimSpace(credentialType) == "" {
		return nil, &ErrorEmptyCredentialType
	}
	if strings.TrimSpace(clientDataJSON) == "" || strings.TrimSpace(authenticatorData) == "" ||
		strings.TrimSpace(signature) == "" {
		return nil, &ErrorInvalidAuthenticatorResponse
	}
	if strings.TrimSpace(sessionToken) == "" {
		return nil, &ErrorEmptySessionToken
	}

	// Retrieve session data from cache
	sessionData, userID, relyingPartyID, svcErr := w.retrieveSessionData(sessionToken)
	if svcErr != nil {
		logger.Error("Failed to retrieve session data", log.String("error", svcErr.Error))
		return nil, svcErr
	}

	logger.Debug("Processing WebAuthn authentication",
		log.String("userID", log.MaskString(userID)),
		log.String("relyingPartyID", relyingPartyID))

	// Get user information
	userInfo, svcErr := w.userService.GetUser(userID)
	if svcErr != nil {
		logger.Error("Failed to retrieve user", log.String("error", svcErr.Error))
		return nil, &ErrorInternalServerError
	}

	// Retrieve user's credentials from database
	credentials, err := w.getWebAuthnCredentialsFromDB(userID)
	if err != nil {
		logger.Error("Failed to retrieve credentials from database", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	logger.Debug("Retrieved credentials for authentication verification",
		log.String("userID", userID),
		log.Int("credentialCount", len(credentials)))

	if len(credentials) == 0 {
		logger.Debug("No credentials found for user", log.String("userID", userID))
		return nil, &ErrorNoCredentialsFound
	}

	// Create WebAuthn user adapter from user info
	webAuthnUserAdapter := createWebAuthnUserFromUserInfo(userInfo, credentials)

	// Initialize WebAuthn library service with relying party configuration
	webAuthnLibService, err := newWebAuthnLibraryService(relyingPartyID, relyingPartyID)
	if err != nil {
		logger.Error("Failed to initialize WebAuthn library service", log.String("error", err.Error()))
		return nil, &ErrorInternalServerError
	}

	// Parse the assertion response from the raw credential data
	// The protocol package expects properly formatted data
	parsedResponse, err := parseAssertionResponse(credentialID, credentialType,
		clientDataJSON, authenticatorData, signature, userHandle)
	if err != nil {
		logger.Error("Failed to parse assertion response", log.String("error", err.Error()))
		return nil, &ErrorInvalidAuthenticatorResponse
	}

	// Verify the credential assertion using WebAuthn library service
	credential, err := webAuthnLibService.FinishAuthentication(webAuthnUserAdapter, *sessionData, parsedResponse)
	if err != nil {
		logger.Error("Failed to validate WebAuthn assertion", log.String("error", err.Error()))
		return nil, &ErrorInvalidSignature
	}

	logger.Debug("WebAuthn authentication verified successfully",
		log.String("credentialID", base64.StdEncoding.EncodeToString(credential.ID)),
		log.Any("signCount", credential.Authenticator.SignCount))

	// Update credential in database to prevent replay attacks
	if err := w.updateWebAuthnCredentialInDB(userID, credential); err != nil {
		logger.Error("Failed to update credential sign count in database", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	logger.Debug("Updated credential sign count in database",
		log.String("userID", userID),
		log.String("credentialID", base64.StdEncoding.EncodeToString(credential.ID)),
		log.Any("newSignCount", credential.Authenticator.SignCount))

	// Clear session data
	w.clearSessionData(sessionToken)

	// Build authentication response
	authResponse := &common.AuthenticationResponse{
		ID:               userInfo.ID,
		Type:             userInfo.Type,
		OrganizationUnit: userInfo.OrganizationUnit,
	}

	// Generate JWT assertion if not skipped
	if !skipAssertion {
		var assertion string
		if existingAssertion != "" {
			// Enrich existing assertion
			assertion, svcErr = w.enrichAssertion(existingAssertion, userInfo, common.AuthenticatorWebAuthn)
		} else {
			// Generate new assertion
			assertion, svcErr = w.generateAssertion(userInfo, common.AuthenticatorWebAuthn)
		}
		if svcErr != nil {
			logger.Error("Failed to generate assertion", log.String("error", svcErr.Error))
			return nil, svcErr
		}
		authResponse.Assertion = assertion
	}

	logger.Debug("WebAuthn authentication completed successfully",
		log.String("userID", log.MaskString(userID)))

	return authResponse, nil
}

// getMetadata returns the metadata for WebAuthn authenticator.
func (w *webAuthnAuthnService) getMetadata() common.AuthenticatorMeta {
	return common.AuthenticatorMeta{
		Name:    common.AuthenticatorWebAuthn,
		Factors: []common.AuthenticationFactor{common.FactorPossession},
	}
}

// generateAssertion generates an assertion token containing user authentication info.
// Note: In production, this should generate a proper JWT token
func (w *webAuthnAuthnService) generateAssertion(
	userInfo *user.User, authenticator string,
) (string, *serviceerror.ServiceError) {
	claims := buildAssertionClaims(userInfo, authenticator)
	jsonData, err := json.Marshal(claims)
	if err != nil {
		return "", &ErrorInternalServerError
	}
	assertion := base64.StdEncoding.EncodeToString(jsonData)
	return assertion, nil
}

// enrichAssertion enriches an existing assertion with WebAuthn authentication.
func (w *webAuthnAuthnService) enrichAssertion(
	existingAssertion string, userInfo *user.User, authenticator string,
) (string, *serviceerror.ServiceError) {
	// For now, just generate a new assertion
	// In production, decode existing JWT, enrich claims, and re-sign
	_ = existingAssertion // Will be used for JWT decoding in full implementation
	return w.generateAssertion(userInfo, authenticator)
}

// getWebAuthnCredentialsFromDB retrieves WebAuthn credentials for a user from the database.
func (w *webAuthnAuthnService) getWebAuthnCredentialsFromDB(userID string) ([]webauthn.Credential, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	// Get passkey credentials from user service
	passkeyCredentials, svcErr := w.userService.GetUserCredentialsByType(userID, user.CredentialTypePasskey.String())
	if svcErr != nil {
		logger.Error("Failed to get passkey credentials",
			log.String("userID", userID),
			log.String("error", svcErr.Error))
		return nil, fmt.Errorf("failed to get passkey credentials: %s", svcErr.Error)
	}

	// Convert user.Credential to generic maps for processing
	storedCreds := make([]map[string]interface{}, 0, len(passkeyCredentials))
	for _, cred := range passkeyCredentials {
		storedCreds = append(storedCreds, map[string]interface{}{
			"storageType":       cred.StorageType,
			"storageAlgo":       cred.StorageAlgo,
			"storageAlgoParams": cred.StorageAlgoParams,
			"value":             cred.Value,
		})
	}

	// Convert stored credentials to WebAuthn credentials
	credentials := make([]webauthn.Credential, 0, len(storedCreds))
	for _, storedCred := range storedCreds {
		// Get the credential value
		credValueStr, ok := storedCred["value"].(string)
		if !ok {
			logger.Error("Failed to get credential value",
				log.String("userID", userID))
			continue
		}

		var credential webauthn.Credential
		if err := json.Unmarshal([]byte(credValueStr), &credential); err != nil {
			// Log error but continue processing other credentials
			logger.Error("Failed to unmarshal WebAuthn credential",
				log.String("userID", userID),
				log.Error(err))
			continue
		}
		credentials = append(credentials, credential)
	}

	logger.Debug("Retrieved WebAuthn credentials from database",
		log.String("userID", userID),
		log.Int("credentialCount", len(credentials)))

	return credentials, nil
}

// storeWebAuthnCredentialInDB stores a WebAuthn credential in the database.
func (w *webAuthnAuthnService) storeWebAuthnCredentialInDB(userID string, credential *webauthn.Credential) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	// Serialize the WebAuthn credential to JSON for storage in the Value field
	credentialJSON, err := json.Marshal(credential)
	if err != nil {
		logger.Error("Failed to marshal credential",
			log.String("userID", userID),
			log.Error(err))
		return fmt.Errorf("failed to marshal credential: %w", err)
	}

	// Get existing passkey credentials to append to
	existingCredentials, svcErr := w.userService.GetUserCredentialsByType(userID, user.CredentialTypePasskey.String())
	if svcErr != nil {
		logger.Error("Failed to get existing passkey credentials",
			log.String("userID", userID),
			log.String("error", svcErr.Error))
		return fmt.Errorf("failed to get existing passkey credentials: %s", svcErr.Error)
	}

	// Create a new credential entry
	// For passkey credentials, only the Value field is set (contains the WebAuthn credential as JSON)
	// StorageType, StorageAlgo, and StorageAlgoParams are kept empty since passkeys don't use hashing
	newCredential := user.Credential{
		Value: string(credentialJSON),
	}

	// Append the new credential to existing ones
	existingCredentials = append(existingCredentials, newCredential)

	// Update credentials in the database
	svcErr = w.userService.UpdateUserCredentials(userID, user.CredentialTypePasskey.String(), existingCredentials)
	if svcErr != nil {
		logger.Error("Failed to update passkey credentials",
			log.String("userID", userID),
			log.String("error", svcErr.Error))
		return fmt.Errorf("failed to update passkey credentials: %s", svcErr.Error)
	}

	logger.Debug("Successfully stored WebAuthn credential in database",
		log.String("userID", userID),
		log.String("credentialID", base64.StdEncoding.EncodeToString(credential.ID)))

	return nil
}

// updateWebAuthnCredentialInDB updates an existing WebAuthn credential in the database.
// This is primarily used to update the sign count after successful authentication.
func (w *webAuthnAuthnService) updateWebAuthnCredentialInDB(
	userID string, updatedCredential *webauthn.Credential,
) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	// Get all existing passkey credentials
	existingCredentials, svcErr := w.userService.GetUserCredentialsByType(userID, user.CredentialTypePasskey.String())
	if svcErr != nil {
		logger.Error("Failed to get existing credentials",
			log.String("userID", userID),
			log.String("error", svcErr.Error))
		return fmt.Errorf("failed to get existing credentials: %s", svcErr.Error)
	}

	// Find and update the matching credential
	found := false
	updatedCredentials := make([]user.Credential, 0, len(existingCredentials))

	for _, storedCred := range existingCredentials {
		var credential webauthn.Credential
		if err := json.Unmarshal([]byte(storedCred.Value), &credential); err != nil {
			// Keep the credential as-is if we can't unmarshal it
			logger.Warn("Failed to unmarshal credential, keeping original",
				log.String("userID", userID),
				log.Error(err))
			updatedCredentials = append(updatedCredentials, storedCred)
			continue
		}

		// Check if this is the credential to update
		if string(credential.ID) == string(updatedCredential.ID) {
			// Serialize the updated credential
			credentialJSON, err := json.Marshal(updatedCredential)
			if err != nil {
				logger.Error("Failed to marshal updated credential",
					log.String("userID", userID),
					log.Error(err))
				return fmt.Errorf("failed to marshal updated credential: %w", err)
			}

			// Create updated credential entry
			updatedCred := user.Credential{
				StorageType:       storedCred.StorageType,
				StorageAlgo:       storedCred.StorageAlgo,
				StorageAlgoParams: storedCred.StorageAlgoParams,
				Value:             string(credentialJSON),
			}
			updatedCredentials = append(updatedCredentials, updatedCred)
			found = true

			logger.Debug("Updated credential in memory",
				log.String("userID", userID),
				log.String("credentialID", base64.StdEncoding.EncodeToString(updatedCredential.ID)),
				log.Any("newSignCount", updatedCredential.Authenticator.SignCount))
		} else {
			// Keep the credential as-is
			updatedCredentials = append(updatedCredentials, storedCred)
		}
	}

	if !found {
		logger.Warn("Credential not found for update",
			log.String("userID", userID),
			log.String("credentialID", base64.StdEncoding.EncodeToString(updatedCredential.ID)))
		return fmt.Errorf("credential not found for update")
	}

	// Update all passkey credentials in the database
	svcErr = w.userService.UpdateUserCredentials(userID, user.CredentialTypePasskey.String(), updatedCredentials)
	if svcErr != nil {
		logger.Error("Failed to update credentials",
			log.String("userID", userID),
			log.String("error", svcErr.Error))
		return fmt.Errorf("failed to update credentials: %s", svcErr.Error)
	}

	logger.Debug("Successfully updated WebAuthn credential in database",
		log.String("userID", userID),
		log.String("credentialID", base64.StdEncoding.EncodeToString(updatedCredential.ID)),
		log.Any("newSignCount", updatedCredential.Authenticator.SignCount))

	return nil
}
