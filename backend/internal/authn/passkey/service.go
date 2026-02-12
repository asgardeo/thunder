/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

// Package passkey implements the Passkey authentication service.
package passkey

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/user"
)

const (
	// loggerComponentName is the component name for logging.
	loggerComponentName = "PasskeyService"
)

// PasskeyServiceInterface defines the interface for passkey authentication and registration operations.
type PasskeyServiceInterface interface {
	// Registration methods
	StartRegistration(req *PasskeyRegistrationStartRequest) (*PasskeyRegistrationStartData, *serviceerror.ServiceError)
	FinishRegistration(
		req *PasskeyRegistrationFinishRequest) (*PasskeyRegistrationFinishData, *serviceerror.ServiceError)

	// Authentication methods
	StartAuthentication(
		req *PasskeyAuthenticationStartRequest) (*PasskeyAuthenticationStartData, *serviceerror.ServiceError)
	FinishAuthentication(
		req *PasskeyAuthenticationFinishRequest) (*common.AuthenticationResponse, *serviceerror.ServiceError)
}

// passkeyService is the default implementation of PasskeyServiceInterface.
type passkeyService struct {
	userService  user.UserServiceInterface
	sessionStore sessionStoreInterface
	logger       *log.Logger
}

// newPasskeyService creates a new instance of passkey service.
func newPasskeyService(userSvc user.UserServiceInterface, sessionStore sessionStoreInterface) PasskeyServiceInterface {
	if userSvc == nil {
		userSvc = user.GetUserService()
	}

	service := &passkeyService{
		userService:  userSvc,
		sessionStore: sessionStore,
		logger:       log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName)),
	}
	common.RegisterAuthenticator(service.getMetadata())

	return service
}

// StartRegistration initiates passkey credential registration for a user.
func (w *passkeyService) StartRegistration(
	req *PasskeyRegistrationStartRequest,
) (*PasskeyRegistrationStartData, *serviceerror.ServiceError) {
	if req == nil {
		return nil, &ErrorInvalidFinishData
	}

	logger := w.logger.With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Starting passkey credential registration",
		log.String("userID", log.MaskString(req.UserID)),
		log.String("relyingPartyID", req.RelyingPartyID))

	// Validate input
	if svcErr := validateRegistrationStartRequest(req); svcErr != nil {
		return nil, svcErr
	}

	// Retrieve core user
	coreUser, svcErr := w.userService.GetUser(context.TODO(), req.UserID)
	if svcErr != nil {
		return nil, handleUserRetrievalError(svcErr, req.UserID, logger)
	}

	// Build relying party display name
	rpDisplayName := req.RelyingPartyName
	if rpDisplayName == "" {
		rpDisplayName = req.RelyingPartyID
	}

	// Retrieve user's existing passkey credentials from database
	credentials, err := w.getStoredPasskeyCredentials(req.UserID)
	if err != nil {
		logger.Error("Failed to retrieve credentials from database", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	logger.Debug("Retrieved existing credentials",
		log.String("userID", req.UserID),
		log.Int("credentialCount", len(credentials)))

	// Create passkey user
	webAuthnUser := newWebAuthnUserFromCoreUser(coreUser, credentials)

	// Initialize WebAuthn service with relying party configuration
	rpOrigins := getConfiguredOrigins()
	webAuthnService, err := newDefaultWebAuthnService(req.RelyingPartyID, rpDisplayName, rpOrigins)
	if err != nil {
		logger.Error("Failed to initialize WebAuthn service", log.String("error", err.Error()))
		return nil, &serviceerror.InternalServerError
	}

	// Configure registration options
	registrationOptions := buildRegistrationOptions(req)

	// Begin registration ceremony using the WebAuthn service
	// The WebAuthn service will generate challenge and set timeout automatically
	options, sessionData, err := webAuthnService.BeginRegistration(webAuthnUser, registrationOptions)
	if err != nil {
		logger.Error("Failed to begin passkey registration", log.String("error", err.Error()))
		return nil, &serviceerror.InternalServerError
	}

	// Store session data in cache with TTL
	sessionToken, svcErr := w.storeSessionData(req.UserID, req.RelyingPartyID, sessionData)
	if svcErr != nil {
		logger.Error("Failed to store session data", log.String("error", svcErr.Error))
		return nil, svcErr
	}

	logger.Debug("Passkey credential creation options generated successfully",
		log.String("userID", log.MaskString(req.UserID)),
		log.Int("credentialsCount", len(credentials)))

	// Convert to custom structure with properly encoded challenge
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

	return &PasskeyRegistrationStartData{
		PublicKeyCredentialCreationOptions: creationOptions,
		SessionToken:                       sessionToken,
	}, nil
}

// FinishRegistration completes passkey credential registration.
func (w *passkeyService) FinishRegistration(req *PasskeyRegistrationFinishRequest) (
	*PasskeyRegistrationFinishData, *serviceerror.ServiceError) {
	logger := w.logger.With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Finishing passkey credential registration")

	// Validate input
	if svcErr := validateRegistrationFinishRequest(req); svcErr != nil {
		logger.Debug("Registration finish request validation failed")
		return nil, svcErr
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

	// Parse the attestation response
	// This ensures all required fields including the Raw field are properly populated
	parsedCredential, err := parseAttestationResponse(
		req.CredentialID,
		credentialType,
		req.ClientDataJSON,
		req.AttestationObject,
	)
	if err != nil {
		logger.Debug("Failed to parse attestation response",
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

	// Get core user
	coreUser, svcErr := w.userService.GetUser(context.TODO(), userID)
	if svcErr != nil {
		logger.Error("Failed to retrieve user", log.String("error", svcErr.Error))
		return nil, &serviceerror.InternalServerError
	}

	// Retrieve existing credentials from database
	credentials, err := w.getStoredPasskeyCredentials(userID)
	if err != nil {
		logger.Error("Failed to retrieve credentials from database", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	logger.Debug("Retrieved existing credentials for user",
		log.String("userID", userID),
		log.Int("credentialCount", len(credentials)))

	// Create WebAuthn user from core user
	webAuthnUser := newWebAuthnUserFromCoreUser(coreUser, credentials)

	// Initialize WebAuthn service with relying party configuration
	rpOrigins := getConfiguredOrigins()
	webAuthnService, err := newDefaultWebAuthnService(relyingPartyID, relyingPartyID, rpOrigins)
	if err != nil {
		logger.Error("Failed to initialize WebAuthn service", log.String("error", err.Error()))
		return nil, &serviceerror.InternalServerError
	}

	// Verify the credential using WebAuthn service
	credential, err := webAuthnService.CreateCredential(webAuthnUser, *sessionData, parsedCredential)
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

	// Store credential in database using user service
	if err := w.storePasskeyCredential(userID, credential); err != nil {
		logger.Error("Failed to store credential in database", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	// Clear session data
	w.clearSessionData(req.SessionToken)

	return &PasskeyRegistrationFinishData{
		CredentialID:   credentialID,
		CredentialName: credentialName,
		CreatedAt:      time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// StartAuthentication initiates passkey authentication for a user.
func (w *passkeyService) StartAuthentication(req *PasskeyAuthenticationStartRequest) (
	*PasskeyAuthenticationStartData, *serviceerror.ServiceError) {
	if req == nil {
		return nil, &ErrorInvalidFinishData
	}

	logger := w.logger.With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Starting passkey authentication",
		log.String("userID", log.MaskString(req.UserID)),
		log.String("relyingPartyID", req.RelyingPartyID))

	// Validate input
	if svcErr := validateAuthenticationStartRequest(req); svcErr != nil {
		return nil, svcErr
	}

	// Retrieve user by userID to verify user exists
	coreUser, svcErr := w.userService.GetUser(context.TODO(), req.UserID)
	if svcErr != nil {
		return nil, handleUserRetrievalError(svcErr, req.UserID, logger)
	}

	// Retrieve user's registered passkey credentials from database
	credentials, err := w.getStoredPasskeyCredentials(req.UserID)
	if err != nil {
		logger.Error("Failed to retrieve credentials from database", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	logger.Debug("Retrieved credentials for authentication",
		log.String("userID", req.UserID),
		log.Int("credentialCount", len(credentials)))

	if len(credentials) == 0 {
		logger.Debug("No credentials found for user", log.String("userID", req.UserID))
		return nil, &ErrorNoCredentialsFound
	}

	// Create WebAuthn user from core user
	webAuthnUser := newWebAuthnUserFromCoreUser(coreUser, credentials)

	// Initialize WebAuthn service with relying party configuration
	rpOrigins := getConfiguredOrigins()
	webAuthnService, err := newDefaultWebAuthnService(req.RelyingPartyID, req.RelyingPartyID, rpOrigins)
	if err != nil {
		logger.Error("Failed to initialize WebAuthn service", log.String("error", err.Error()))
		return nil, &serviceerror.InternalServerError
	}

	// Begin login ceremony using the WebAuthn service
	// The WebAuthn service will generate challenge and set timeout automatically
	options, sessionData, err := webAuthnService.BeginLogin(webAuthnUser)
	if err != nil {
		logger.Error("Failed to begin passkey login", log.String("error", err.Error()))
		return nil, &serviceerror.InternalServerError
	}

	// Store session data in cache with TTL
	sessionToken, svcErr := w.storeSessionData(req.UserID, req.RelyingPartyID, sessionData)
	if svcErr != nil {
		logger.Error("Failed to store session data", log.String("error", svcErr.Error))
		return nil, svcErr
	}

	// Convert to custom structure with properly encoded challenge
	requestOptions := PublicKeyCredentialRequestOptions{
		Challenge:        base64.RawURLEncoding.EncodeToString(options.Response.Challenge),
		Timeout:          options.Response.Timeout,
		RelyingPartyID:   options.Response.RelyingPartyID,
		AllowCredentials: options.Response.AllowedCredentials,
		UserVerification: options.Response.UserVerification,
		Extensions:       options.Response.Extensions,
	}

	return &PasskeyAuthenticationStartData{
		PublicKeyCredentialRequestOptions: requestOptions,
		SessionToken:                      sessionToken,
	}, nil
}

// FinishAuthentication completes passkey authentication.
func (w *passkeyService) FinishAuthentication(req *PasskeyAuthenticationFinishRequest) (
	*common.AuthenticationResponse, *serviceerror.ServiceError) {
	logger := w.logger.With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Finishing passkey authentication")

	// Validate input
	if svcErr := validateAuthenticationFinishRequest(req); svcErr != nil {
		return nil, svcErr
	}

	// Retrieve session data from cache
	sessionData, userID, relyingPartyID, svcErr := w.retrieveSessionData(req.SessionToken)
	if svcErr != nil {
		logger.Error("Failed to retrieve session data", log.String("error", svcErr.Error))
		return nil, svcErr
	}

	logger.Debug("Processing passkey authentication",
		log.String("userID", log.MaskString(userID)),
		log.String("relyingPartyID", relyingPartyID))

	// Get core user
	coreUser, svcErr := w.userService.GetUser(context.TODO(), userID)
	if svcErr != nil {
		logger.Error("Failed to retrieve user", log.String("error", svcErr.Error))
		return nil, &serviceerror.InternalServerError
	}

	// Retrieve user's credentials from database
	credentials, err := w.getStoredPasskeyCredentials(userID)
	if err != nil {
		logger.Error("Failed to retrieve credentials from database", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	logger.Debug("Retrieved credentials for authentication verification",
		log.String("userID", userID),
		log.Int("credentialCount", len(credentials)))

	if len(credentials) == 0 {
		logger.Debug("No credentials found for user", log.String("userID", userID))
		return nil, &ErrorNoCredentialsFound
	}

	// Create WebAuthn user from core user
	webAuthnUser := newWebAuthnUserFromCoreUser(coreUser, credentials)

	// Initialize WebAuthn service with relying party configuration
	rpOrigins := getConfiguredOrigins()
	webAuthnService, err := newDefaultWebAuthnService(relyingPartyID, relyingPartyID, rpOrigins)
	if err != nil {
		logger.Error("Failed to initialize WebAuthn service", log.String("error", err.Error()))
		return nil, &serviceerror.InternalServerError
	}

	// Parse the assertion response from the raw credential data
	parsedResponse, err := parseAssertionResponse(req.CredentialID, req.CredentialType,
		req.ClientDataJSON, req.AuthenticatorData, req.Signature, req.UserHandle)
	if err != nil {
		logger.Debug("Failed to parse assertion response", log.String("error", err.Error()))
		return nil, &ErrorInvalidAuthenticatorResponse
	}

	// Verify the credential assertion using WebAuthn service
	credential, err := webAuthnService.ValidateLogin(webAuthnUser, *sessionData, parsedResponse)
	if err != nil {
		logger.Debug("Failed to validate WebAuthn assertion", log.String("error", err.Error()))
		return nil, &ErrorInvalidSignature
	}

	logger.Debug("Passkey authentication verified successfully",
		log.String("credentialID", base64.StdEncoding.EncodeToString(credential.ID)),
		log.Any("signCount", credential.Authenticator.SignCount))

	// Update credential in database to prevent replay attacks
	if err := w.updatePasskeyCredential(userID, credential); err != nil {
		logger.Error("Failed to update credential sign count in database", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	logger.Debug("Updated credential sign count in database",
		log.String("userID", userID),
		log.String("credentialID", base64.StdEncoding.EncodeToString(credential.ID)),
		log.Any("newSignCount", credential.Authenticator.SignCount))

	// Clear session data
	w.clearSessionData(req.SessionToken)

	// Build authentication response
	authResponse := &common.AuthenticationResponse{
		ID:               coreUser.ID,
		Type:             coreUser.Type,
		OrganizationUnit: coreUser.OrganizationUnit,
	}

	logger.Debug("Passkey authentication completed successfully",
		log.String("userID", log.MaskString(userID)))

	return authResponse, nil
}

// getMetadata returns the metadata for passkey authenticator.
func (w *passkeyService) getMetadata() common.AuthenticatorMeta {
	return common.AuthenticatorMeta{
		Name:    common.AuthenticatorPasskey,
		Factors: []common.AuthenticationFactor{common.FactorPossession, common.FactorInherence},
	}
}

// getStoredPasskeyCredentials retrieves passkey credentials for a user from the database.
func (w *passkeyService) getStoredPasskeyCredentials(userID string) ([]webauthnCredential, error) {
	logger := w.logger.With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	// Get passkey credentials from user service
	passkeyCredentials, svcErr := w.userService.GetUserCredentialsByType(
		context.TODO(), userID, user.CredentialTypePasskey.String())
	if svcErr != nil {
		logger.Error("Failed to get passkey credentials",
			log.String("userID", userID),
			log.String("error", svcErr.Error))
		return nil, fmt.Errorf("failed to get passkey credentials: %s", svcErr.Error)
	}

	// Convert passkey credentials to generic maps for processing
	storedCreds := make([]map[string]interface{}, 0, len(passkeyCredentials))
	for _, cred := range passkeyCredentials {
		storedCreds = append(storedCreds, map[string]interface{}{
			"storageType":       cred.StorageType,
			"storageAlgo":       cred.StorageAlgo,
			"storageAlgoParams": cred.StorageAlgoParams,
			"value":             cred.Value,
		})
	}

	// Convert stored credentials to passkey credentials
	credentials := make([]webauthnCredential, 0, len(storedCreds))
	for _, storedCred := range storedCreds {
		// Get the credential value
		credValueStr, ok := storedCred["value"].(string)
		if !ok {
			logger.Error("Failed to get credential value",
				log.String("userID", userID))
			continue
		}

		var credential webauthnCredential
		if err := json.Unmarshal([]byte(credValueStr), &credential); err != nil {
			// Log error but continue processing other credentials
			logger.Error("Failed to unmarshal passkey credential",
				log.String("userID", userID),
				log.Error(err))
			continue
		}
		credentials = append(credentials, credential)
	}

	logger.Debug("Retrieved passkey credentials from database",
		log.String("userID", userID),
		log.Int("credentialCount", len(credentials)))

	return credentials, nil
}

// storePasskeyCredential stores a passkey credential in the database.
func (w *passkeyService) storePasskeyCredential(userID string, credential *webauthnCredential) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	// Serialize the passkey credential to JSON for storage in the Value field
	credentialJSON, err := json.Marshal(credential)
	if err != nil {
		logger.Error("Failed to marshal credential",
			log.String("userID", userID),
			log.Error(err))
		return fmt.Errorf("failed to marshal credential: %w", err)
	}

	// Get existing passkey credentials to append to
	existingCredentials, svcErr := w.userService.GetUserCredentialsByType(
		context.TODO(), userID, user.CredentialTypePasskey.String())
	if svcErr != nil {
		logger.Error("Failed to get existing passkey credentials",
			log.String("userID", userID),
			log.String("error", svcErr.Error))
		return fmt.Errorf("failed to get existing passkey credentials: %s", svcErr.Error)
	}

	// Create a new credential entry
	newCredential := user.Credential{
		Value: string(credentialJSON),
	}

	// Append the new credential to existing ones
	existingCredentials = append(existingCredentials, newCredential)

	// Prepare credentials map for update
	credentialsMap := map[string][]user.Credential{
		user.CredentialTypePasskey.String(): existingCredentials,
	}
	credentialsJSON, err := json.Marshal(credentialsMap)
	if err != nil {
		logger.Error("Failed to marshal credentials",
			log.String("userID", userID),
			log.Error(err))
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	// Update credentials in the database
	svcErr = w.userService.UpdateUserCredentials(context.TODO(), userID, credentialsJSON)
	if svcErr != nil {
		logger.Error("Failed to update passkey credentials",
			log.String("userID", userID),
			log.String("error", svcErr.Error))
		return fmt.Errorf("failed to update passkey credentials: %s", svcErr.Error)
	}

	logger.Debug("Successfully stored passkey credential in database",
		log.String("userID", userID),
		log.String("credentialID", base64.StdEncoding.EncodeToString(credential.ID)))

	return nil
}

// updatePasskeyCredential updates an existing passkey credential in the database.
func (w *passkeyService) updatePasskeyCredential(
	userID string, updatedCredential *webauthnCredential,
) error {
	logger := w.logger.With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	// Get all existing passkey credentials
	existingCredentials, svcErr := w.userService.GetUserCredentialsByType(
		context.TODO(), userID, user.CredentialTypePasskey.String())
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
		var credential webauthnCredential
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
		logger.Warn("Passkey credential not found for update",
			log.String("userID", userID),
			log.String("credentialID", base64.StdEncoding.EncodeToString(updatedCredential.ID)))
		return fmt.Errorf("credential not found for update")
	}

	// Prepare credentials map for update
	credentialsMap := map[string][]user.Credential{
		user.CredentialTypePasskey.String(): updatedCredentials,
	}
	credentialsJSON, err := json.Marshal(credentialsMap)
	if err != nil {
		logger.Error("Failed to marshal credentials",
			log.String("userID", userID),
			log.Error(err))
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	// Update all passkey credentials in the database
	svcErr = w.userService.UpdateUserCredentials(context.TODO(), userID, credentialsJSON)
	if svcErr != nil {
		logger.Error("Failed to update credentials",
			log.String("userID", userID),
			log.String("error", svcErr.Error))
		return fmt.Errorf("failed to update credentials: %s", svcErr.Error)
	}

	logger.Debug("Successfully updated passkey credential in database",
		log.String("userID", userID),
		log.String("credentialID", base64.StdEncoding.EncodeToString(updatedCredential.ID)),
		log.Any("newSignCount", updatedCredential.Authenticator.SignCount))

	return nil
}
