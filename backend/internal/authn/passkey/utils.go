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

package passkey

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/user"
)

const (
	// defaultCredentialNameFormat is the format for auto-generated credential names.
	defaultCredentialNameFormat = "Passkey %s"
	// defaultCredentialDateFormat is the date format for credential names.
	defaultCredentialDateFormat = "2006-01-02" // nolint:gosec // This is a date format, not a credential
	// defaultOriginHTTP is the default HTTP origin for local development.
	defaultOriginHTTP = "https://localhost:8090"
)

// generateDefaultCredentialName generates a default credential name with the current date.
func generateDefaultCredentialName() string {
	return fmt.Sprintf(defaultCredentialNameFormat, time.Now().Format(defaultCredentialDateFormat))
}

// getConfiguredOrigins retrieves the allowed origins from runtime configuration.
func getConfiguredOrigins() []string {
	// Default origins if not configured
	defaultOrigins := []string{defaultOriginHTTP}

	// Try to get runtime configuration with panic recovery
	var originList []string
	func() {
		defer func() {
			if r := recover(); r != nil {
				// If configuration access fails, originList stays nil
				originList = nil
			}
		}()

		runtime := config.GetThunderRuntime()
		if runtime != nil {
			originList = runtime.Config.Passkey.AllowedOrigins
		}
	}()

	// If no origins configured, return defaults
	if len(originList) == 0 {
		return defaultOrigins
	}

	return originList
}

// parseUserAttributes extracts user attributes from JSON.
func parseUserAttributes(attributes json.RawMessage) map[string]interface{} {
	if len(attributes) == 0 {
		return nil
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(attributes, &parsed); err != nil {
		return nil
	}

	return parsed
}

// buildUserDisplayName builds a display name from user attributes.
func buildUserDisplayName(userID string, attributes map[string]interface{}) string {
	if attributes == nil {
		return userID
	}

	firstName, firstOk := attributes["firstName"].(string)
	lastName, lastOk := attributes["lastName"].(string)

	if firstOk && firstName != "" {
		if lastOk && lastName != "" {
			return firstName + " " + lastName
		}
		return firstName
	}

	return userID
}

// resolveUserName builds a username from user attributes.
func resolveUserName(userID string, attributes map[string]interface{}) string {
	if attributes == nil {
		return userID
	}

	if username, ok := attributes["username"].(string); ok && username != "" {
		return username
	}

	if email, ok := attributes["email"].(string); ok && email != "" {
		return email
	}

	return userID
}

// extractCoreUser extracts display name and username from core user.
func extractCoreUser(coreUser *user.User) (displayName, userName string) {
	attributes := parseUserAttributes(coreUser.Attributes)
	displayName = buildUserDisplayName(coreUser.ID, attributes)
	userName = resolveUserName(coreUser.ID, attributes)
	return displayName, userName
}

// decodeBase64 attempts to decode a base64 string using multiple encodings.
func decodeBase64(s string) ([]byte, error) {
	// Try RawURLEncoding (RFC 4648 section 5, no padding) - preferred for WebAuthn
	if b, err := base64.RawURLEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	// Try URLEncoding (RFC 4648 section 5, with padding)
	if b, err := base64.URLEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	// Try RawStdEncoding (RFC 4648 section 4, no padding)
	if b, err := base64.RawStdEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	// Try StdEncoding (RFC 4648 section 4, with padding)
	return base64.StdEncoding.DecodeString(s)
}

// parseAssertionResponse converts raw string parameters to ParsedCredentialAssertionData.
func parseAssertionResponse(credentialID, credentialType, clientDataJSON,
	authenticatorData, signature, userHandle string) (*ParsedCredentialAssertionData, error) {
	// Decode all base64url encoded parameters
	rawID, err := decodeBase64(credentialID)
	if err != nil {
		return nil, fmt.Errorf("failed to decode credential ID: %w", err)
	}

	clientData, err := decodeBase64(clientDataJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to decode client data JSON: %w", err)
	}

	authData, err := decodeBase64(authenticatorData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode authenticator data: %w", err)
	}

	sig, err := decodeBase64(signature)
	if err != nil {
		return nil, fmt.Errorf("failed to decode signature: %w", err)
	}

	var userHandleBytes []byte
	if userHandle != "" {
		userHandleBytes, err = decodeBase64(userHandle)
		if err != nil {
			// User handle is optional, so we can continue without it
			userHandleBytes = nil
		}
	}

	// Parse client data JSON to extract required fields
	var clientDataParsed protocol.CollectedClientData
	if err := json.Unmarshal(clientData, &clientDataParsed); err != nil {
		return nil, fmt.Errorf("failed to parse client data JSON: %w", err)
	}

	// Parse authenticator data
	authDataParsed := protocol.AuthenticatorData{}
	if err := authDataParsed.Unmarshal(authData); err != nil {
		return nil, fmt.Errorf("failed to parse authenticator data: %w", err)
	}

	// Create the parsed credential assertion data structure
	parsed := &ParsedCredentialAssertionData{
		ParsedPublicKeyCredential: protocol.ParsedPublicKeyCredential{
			RawID: rawID,
			ParsedCredential: protocol.ParsedCredential{
				ID:   credentialID,
				Type: credentialType,
			},
		},
		Response: protocol.ParsedAssertionResponse{
			CollectedClientData: clientDataParsed,
			AuthenticatorData:   authDataParsed,
			Signature:           sig,
			UserHandle:          userHandleBytes,
		},
		Raw: protocol.CredentialAssertionResponse{
			PublicKeyCredential: protocol.PublicKeyCredential{
				Credential: protocol.Credential{
					ID:   credentialID,
					Type: credentialType,
				},
				RawID: rawID,
			},
			AssertionResponse: protocol.AuthenticatorAssertionResponse{
				AuthenticatorResponse: protocol.AuthenticatorResponse{
					ClientDataJSON: clientData,
				},
				AuthenticatorData: authData,
				Signature:         sig,
				UserHandle:        userHandleBytes,
			},
		},
	}

	return parsed, nil
}

// parseAttestationResponse converts raw credential data to ParsedCredentialCreationData.
// This function uses the protocol package's ParseCredentialCreationResponseBytes to properly
// parse the credential data with all required fields populated.
func parseAttestationResponse(
	credentialID, credentialType, clientDataJSON, attestationObject string,
) (*ParsedCredentialCreationData, error) {
	// Decode inputs to ensure we have valid bytes, regardless of input encoding
	clientDataBytes, err := decodeBase64(clientDataJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to decode client data JSON: %w", err)
	}

	attestationBytes, err := decodeBase64(attestationObject)
	if err != nil {
		return nil, fmt.Errorf("failed to decode attestation object: %w", err)
	}

	// Re-encode to RawURLEncoding to ensure consistent format for the protocol parser
	// The protocol package expects RawURLEncoding for these fields in the JSON
	clientDataEncoded := base64.RawURLEncoding.EncodeToString(clientDataBytes)
	attestationEncoded := base64.RawURLEncoding.EncodeToString(attestationBytes)

	// Construct the JSON payload for the protocol parser
	// The protocol package expects the data in this format
	payload := map[string]interface{}{
		"id":    credentialID,
		"rawId": credentialID,
		"type":  credentialType,
		"response": map[string]interface{}{
			"clientDataJSON":    clientDataEncoded,
			"attestationObject": attestationEncoded,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal credential: %w", err)
	}

	// Parse the attestation response using the protocol package's byte parser
	// This properly parses the CBOR-encoded attestation object and populates all fields
	parsed, err := protocol.ParseCredentialCreationResponseBytes(jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse credential creation response: %w", err)
	}

	return parsed, nil
}

// validateRegistrationStartRequest validates the registration start request.
func validateRegistrationStartRequest(req *PasskeyRegistrationStartRequest) *serviceerror.ServiceError {
	if strings.TrimSpace(req.UserID) == "" {
		return &ErrorEmptyUserIdentifier
	}
	if strings.TrimSpace(req.RelyingPartyID) == "" {
		return &ErrorEmptyRelyingPartyID
	}
	return nil
}

// validateRegistrationFinishRequest validates the registration finish request.
func validateRegistrationFinishRequest(req *PasskeyRegistrationFinishRequest) *serviceerror.ServiceError {
	if req == nil {
		return &ErrorInvalidFinishData
	}
	if strings.TrimSpace(req.SessionToken) == "" {
		return &ErrorEmptySessionToken
	}
	if strings.TrimSpace(req.CredentialID) == "" {
		return &ErrorInvalidFinishData
	}
	if strings.TrimSpace(req.ClientDataJSON) == "" {
		return &ErrorInvalidFinishData
	}
	if strings.TrimSpace(req.AttestationObject) == "" {
		return &ErrorInvalidFinishData
	}
	return nil
}

// validateAuthenticationStartRequest validates the authentication start request.
func validateAuthenticationStartRequest(req *PasskeyAuthenticationStartRequest) *serviceerror.ServiceError {
	if req == nil {
		return &ErrorInvalidFinishData
	}
	if strings.TrimSpace(req.UserID) == "" {
		return &ErrorEmptyUserIdentifier
	}
	if strings.TrimSpace(req.RelyingPartyID) == "" {
		return &ErrorEmptyRelyingPartyID
	}
	return nil
}

// validateAuthenticationFinishRequest validates the authentication finish request.
func validateAuthenticationFinishRequest(req *PasskeyAuthenticationFinishRequest) *serviceerror.ServiceError {
	if req == nil {
		return &ErrorInvalidFinishData
	}
	if strings.TrimSpace(req.CredentialID) == "" {
		return &ErrorEmptyCredentialID
	}
	if strings.TrimSpace(req.CredentialType) == "" {
		return &ErrorEmptyCredentialType
	}
	if strings.TrimSpace(req.ClientDataJSON) == "" ||
		strings.TrimSpace(req.AuthenticatorData) == "" ||
		strings.TrimSpace(req.Signature) == "" {
		return &ErrorInvalidAuthenticatorResponse
	}
	if strings.TrimSpace(req.SessionToken) == "" {
		return &ErrorEmptySessionToken
	}
	return nil
}

// handleUserRetrievalError handles errors from user retrieval.
func handleUserRetrievalError(
	svcErr *serviceerror.ServiceError, userID string, logger *log.Logger,
) *serviceerror.ServiceError {
	if svcErr.Type == serviceerror.ClientErrorType {
		logger.Debug("User not found", log.String("userID", log.MaskString(userID)))
		return &ErrorUserNotFound
	}
	logger.Error("Failed to retrieve user", log.String("error", svcErr.Error))
	return &serviceerror.InternalServerError
}

// buildRegistrationOptions builds registration options from the request.
func buildRegistrationOptions(req *PasskeyRegistrationStartRequest) []RegistrationOption {
	var registrationOptions []RegistrationOption

	// Set authenticator selection if provided
	if req.AuthenticatorSelection != nil {
		authSelection := buildAuthenticatorSelection(req.AuthenticatorSelection)
		registrationOptions = append(registrationOptions, webauthn.WithAuthenticatorSelection(authSelection))
	}

	// Set attestation conveyance preference
	if req.Attestation != "" {
		conveyance := protocol.ConveyancePreference(req.Attestation)
		registrationOptions = append(registrationOptions, webauthn.WithConveyancePreference(conveyance))
	}

	return registrationOptions
}

// buildAuthenticatorSelection builds authenticator selection from request.
func buildAuthenticatorSelection(sel *AuthenticatorSelection) protocol.AuthenticatorSelection {
	authSelection := protocol.AuthenticatorSelection{}

	if sel.AuthenticatorAttachment != "" {
		attachment := protocol.AuthenticatorAttachment(sel.AuthenticatorAttachment)
		authSelection.AuthenticatorAttachment = attachment
	}

	if sel.ResidentKey != "" {
		residentKey := protocol.ResidentKeyRequirement(sel.ResidentKey)
		authSelection.ResidentKey = residentKey
	}

	if sel.UserVerification != "" {
		authSelection.UserVerification = protocol.UserVerificationRequirement(sel.UserVerification)
	} else {
		authSelection.UserVerification = protocol.VerificationPreferred
	}

	return authSelection
}
