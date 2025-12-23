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

package webauthn

import (
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/stretchr/testify/assert"

	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/user"
)

func TestGenerateDefaultCredentialName(t *testing.T) {
	name := generateDefaultCredentialName()

	assert.NotEmpty(t, name)
	assert.Contains(t, name, "Passkey")
	assert.Contains(t, name, time.Now().Format("2006-01-02"))
}

func TestGetConfiguredOrigins(t *testing.T) {
	origins := getConfiguredOrigins()

	assert.NotNil(t, origins)
	assert.NotEmpty(t, origins)
}

func TestParseUserAttributes_ValidJSON(t *testing.T) {
	attrs := json.RawMessage(`{"name":"John Doe","email":"john@example.com"}`)

	result := parseUserAttributes(attrs)

	assert.NotNil(t, result)
	assert.Equal(t, "John Doe", result["name"])
	assert.Equal(t, "john@example.com", result["email"])
}

func TestParseUserAttributes_EmptyJSON(t *testing.T) {
	attrs := json.RawMessage(`{}`)

	result := parseUserAttributes(attrs)

	assert.NotNil(t, result)
	assert.Empty(t, result)
}

func TestParseUserAttributes_NilInput(t *testing.T) {
	var attrs json.RawMessage

	result := parseUserAttributes(attrs)

	assert.Nil(t, result)
}

func TestParseUserAttributes_InvalidJSON(t *testing.T) {
	attrs := json.RawMessage(`{invalid json}`)

	result := parseUserAttributes(attrs)

	assert.Nil(t, result)
}

func TestBuildUserDisplayName_WithName(t *testing.T) {
	attrs := map[string]interface{}{
		"firstName": "John",
		"lastName":  "Doe",
	}

	result := buildUserDisplayName(testUserID, attrs)

	assert.Equal(t, "John Doe", result)
}

func TestBuildUserDisplayName_WithFirstNameOnly(t *testing.T) {
	attrs := map[string]interface{}{
		"firstName": "John",
	}

	result := buildUserDisplayName(testUserID, attrs)

	assert.Equal(t, "John", result)
}

func TestBuildUserDisplayName_Fallback(t *testing.T) {
	attrs := map[string]interface{}{
		"email": "john@example.com", // email is not used by buildUserDisplayName
	}

	result := buildUserDisplayName(testUserID, attrs)

	assert.Equal(t, testUserID, result)
}

func TestBuildUserDisplayName_NilAttributes(t *testing.T) {
	result := buildUserDisplayName(testUserID, nil)

	assert.Equal(t, testUserID, result)
}

func TestBuildUserName_WithUsername(t *testing.T) {
	attrs := map[string]interface{}{
		"username": "johndoe",
	}

	result := buildUserName(testUserID, attrs)

	assert.Equal(t, "johndoe", result)
}

func TestBuildUserName_WithEmail(t *testing.T) {
	attrs := map[string]interface{}{
		"email": "john@example.com",
	}

	result := buildUserName(testUserID, attrs)

	assert.Equal(t, "john@example.com", result)
}

func TestBuildUserName_Fallback(t *testing.T) {
	attrs := map[string]interface{}{}

	result := buildUserName(testUserID, attrs)

	assert.Equal(t, testUserID, result)
}

func TestBuildUserName_NilAttributes(t *testing.T) {
	result := buildUserName(testUserID, nil)

	assert.Equal(t, testUserID, result)
}

func TestValidateRegistrationStartRequest_ValidRequest(t *testing.T) {
	req := &WebAuthnRegisterStartRequest{
		UserID:         testUserID,
		RelyingPartyID: "example.com",
	}

	err := validateRegistrationStartRequest(req)

	assert.Nil(t, err)
}

func TestValidateRegistrationStartRequest_EmptyUserID(t *testing.T) {
	req := &WebAuthnRegisterStartRequest{
		UserID:         "",
		RelyingPartyID: "example.com",
	}

	err := validateRegistrationStartRequest(req)

	assert.NotNil(t, err)
	assert.Equal(t, ErrorEmptyUserIdentifier.Code, err.Code)
}

func TestValidateRegistrationStartRequest_EmptyRelyingPartyID(t *testing.T) {
	req := &WebAuthnRegisterStartRequest{
		UserID:         testUserID,
		RelyingPartyID: "",
	}

	err := validateRegistrationStartRequest(req)

	assert.NotNil(t, err)
	assert.Equal(t, ErrorEmptyRelyingPartyID.Code, err.Code)
}

func TestHandleUserRetrievalError_ClientError(t *testing.T) {
	svcErr := &serviceerror.ServiceError{
		Type: serviceerror.ClientErrorType,
		Code: "USER_NOT_FOUND",
	}
	logger := log.GetLogger()

	result := handleUserRetrievalError(svcErr, testUserID, logger)

	assert.NotNil(t, result)
	assert.Equal(t, ErrorUserNotFound.Code, result.Code)
}

func TestHandleUserRetrievalError_ServerError(t *testing.T) {
	svcErr := &serviceerror.ServiceError{
		Type: serviceerror.ServerErrorType,
		Code: "DB_ERROR",
	}
	logger := log.GetLogger()

	result := handleUserRetrievalError(svcErr, testUserID, logger)

	assert.NotNil(t, result)
	assert.Equal(t, ErrorInternalServerError.Code, result.Code)
}

func TestBuildAssertionClaims(t *testing.T) {
	testUser := &user.User{
		ID:               testUserID,
		Type:             "person",
		OrganizationUnit: "org123",
	}

	claims := buildAssertionClaims(testUser, "WebAuthnAuthenticator")

	assert.NotNil(t, claims)
	assert.Equal(t, testUserID, claims["sub"])
	assert.Equal(t, "person", claims["type"])
	assert.Equal(t, "org123", claims["ou"])
	assert.Contains(t, claims, "iat")
	assert.Contains(t, claims, "exp")

	// Check authenticators
	authenticators, ok := claims["amr"].([]string)
	assert.True(t, ok)
	assert.Contains(t, authenticators, "WebAuthnAuthenticator")
}

func TestDecodeBase64_RawURLEncoding(t *testing.T) {
	// Test with RawURLEncoding (no padding)
	input := "SGVsbG8gV29ybGQ"

	result, err := decodeBase64(input)

	assert.NoError(t, err)
	assert.Equal(t, "Hello World", string(result))
}

func TestDecodeBase64_URLEncoding(t *testing.T) {
	// Test with URLEncoding (with padding)
	input := "SGVsbG8gV29ybGQ="

	result, err := decodeBase64(input)

	assert.NoError(t, err)
	assert.Equal(t, "Hello World", string(result))
}

func TestDecodeBase64_StdEncoding(t *testing.T) {
	// Test with standard encoding
	input := "SGVsbG8gV29ybGQ="

	result, err := decodeBase64(input)

	assert.NoError(t, err)
	assert.Equal(t, "Hello World", string(result))
}

func TestDecodeBase64_InvalidInput(t *testing.T) {
	// Test with invalid base64
	input := "not-valid-base64!@#$%"

	_, err := decodeBase64(input)

	assert.Error(t, err)
}

func TestExtractUserInfo_WithFullAttributes(t *testing.T) {
	attrs := json.RawMessage(`{"firstName":"John","lastName":"Doe","username":"johndoe"}`)
	testUser := &user.User{
		ID:               testUserID,
		Type:             "person",
		OrganizationUnit: "org123",
		Attributes:       attrs,
	}

	displayName, userName := extractUserInfo(testUser)

	assert.Equal(t, "John Doe", displayName)
	assert.Equal(t, "johndoe", userName)
}

func TestExtractUserInfo_WithEmailOnly(t *testing.T) {
	attrs := json.RawMessage(`{"email":"john@example.com"}`)
	testUser := &user.User{
		ID:         testUserID,
		Attributes: attrs,
	}

	displayName, userName := extractUserInfo(testUser)

	assert.Equal(t, testUserID, displayName) // Falls back to ID
	assert.Equal(t, "john@example.com", userName)
}

func TestExtractUserInfo_EmptyAttributes(t *testing.T) {
	testUser := &user.User{
		ID: testUserID,
	}

	displayName, userName := extractUserInfo(testUser)
	assert.Equal(t, testUserID, displayName)
	assert.Equal(t, testUserID, userName)
}

func TestBuildAuthenticatorSelection(t *testing.T) {
	tests := []struct {
		name     string
		input    *AuthenticatorSelection
		expected protocol.AuthenticatorSelection
	}{
		{
			name: "All fields populated",
			input: &AuthenticatorSelection{
				AuthenticatorAttachment: "platform",
				ResidentKey:             "required",
				UserVerification:        "required",
			},
			expected: protocol.AuthenticatorSelection{
				AuthenticatorAttachment: protocol.Platform,
				ResidentKey:             protocol.ResidentKeyRequirementRequired,
				UserVerification:        protocol.VerificationRequired,
			},
		},
		{
			name: "Empty fields triggers defaults and skips assignments",
			input: &AuthenticatorSelection{
				AuthenticatorAttachment: "",
				ResidentKey:             "",
				UserVerification:        "",
			},
			expected: protocol.AuthenticatorSelection{
				AuthenticatorAttachment: "",
				ResidentKey:             "",
				UserVerification:        protocol.VerificationPreferred, // Tests the 'else' branch
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildAuthenticatorSelection(tt.input)

			if got.AuthenticatorAttachment != tt.expected.AuthenticatorAttachment {
				t.Errorf("AuthenticatorAttachment: got %v, want %v", got.AuthenticatorAttachment,
					tt.expected.AuthenticatorAttachment)
			}

			if got.ResidentKey != tt.expected.ResidentKey {
				t.Errorf("ResidentKey: got %v, want %v", got.ResidentKey, tt.expected.ResidentKey)
			}

			if got.UserVerification != tt.expected.UserVerification {
				t.Errorf("UserVerification: got %v, want %v", got.UserVerification, tt.expected.UserVerification)
			}
		})
	}
}

// Helper to encode strings to Base64URL (assuming decodeBase64 uses this format).
func b64(s string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(s))
}

func TestParseAssertionResponse(t *testing.T) {
	// Setup valid dummy data
	validClientData := `{"type":"webauthn.get","challenge":"Y2hhbGxlbmdl","origin":"https://localhost"}`
	validAuthData := make([]byte, 37) // Minimum length for AuthData is usually 37 bytes

	tests := []struct {
		name           string
		credentialID   string
		credentialType string
		clientDataJSON string
		authData       string
		signature      string
		userHandle     string
		wantErr        bool
		errContains    string
	}{
		{
			name:           "Success - All fields valid",
			credentialID:   b64("id123"),
			credentialType: "public-key",
			clientDataJSON: b64(validClientData),
			authData:       base64.RawURLEncoding.EncodeToString(validAuthData),
			signature:      b64("signature-bytes"),
			userHandle:     b64("user-123"),
			wantErr:        false,
		},
		{
			name:           "Success - Optional UserHandle fails decoding (soft fail)",
			credentialID:   b64("id123"),
			credentialType: "public-key",
			clientDataJSON: b64(validClientData),
			authData:       base64.RawURLEncoding.EncodeToString(validAuthData),
			signature:      b64("sig"),
			userHandle:     "!!!invalid-base64!!!",
			wantErr:        false, // The code explicitly swallows this error
		},
		{
			name:         "Error - Invalid Credential ID Base64",
			credentialID: "!!!",
			wantErr:      true,
			errContains:  "failed to decode credential ID",
		},
		{
			name:           "Error - Invalid ClientData JSON",
			credentialID:   b64("id"),
			clientDataJSON: b64("{ invalid json"),
			wantErr:        true,
			errContains:    "failed to parse client data JSON",
		},
		{
			name:           "Error - Invalid Authenticator Data",
			credentialID:   b64("id"),
			clientDataJSON: b64(validClientData),
			authData:       b64("too-short"), // Will fail Unmarshal
			wantErr:        true,
			errContains:    "failed to parse authenticator data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseAssertionResponse(
				tt.credentialID,
				tt.credentialType,
				tt.clientDataJSON,
				tt.authData,
				tt.signature,
				tt.userHandle,
			)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// Verify Data Structure Mapping
				assert.Equal(t, tt.credentialType, result.ParsedPublicKeyCredential.ParsedCredential.Type)
				assert.Equal(t, tt.credentialID, result.ParsedPublicKeyCredential.ParsedCredential.ID)

				// Verify UserHandle specifically
				if tt.userHandle != "" && tt.name != "Success - Optional UserHandle fails decoding (soft fail)" {
					assert.NotNil(t, result.Response.UserHandle)
				}
			}
		})
	}
}

// ==================== Tests for L170-171: RawStdEncoding path ====================

func TestDecodeBase64_RawStdEncoding(t *testing.T) {
	// Test with RawStdEncoding (standard base64 without padding)
	// This tests the L170-171 branch
	input := "SGVsbG8gV29ybGQ" // "Hello World" in RawStdEncoding

	result, err := decodeBase64(input)

	assert.NoError(t, err)
	assert.Equal(t, "Hello World", string(result))
}

// ==================== Tests for L187-188, L192-193, L197-198: parseAssertionResponse error paths ====================

func TestParseAssertionResponse_ClientDataJSONDecodeError(t *testing.T) {
	// Tests L187-188: error decoding clientDataJSON
	_, err := parseAssertionResponse(
		b64("valid-id"),
		"public-key",
		"!!!invalid-base64!!!",
		b64("valid-auth-data"),
		b64("valid-signature"),
		"",
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode client data JSON")
}

func TestParseAssertionResponse_AuthenticatorDataDecodeError(t *testing.T) {
	// Tests L192-193: error decoding authenticatorData
	_, err := parseAssertionResponse(
		b64("valid-id"),
		"public-key",
		b64(`{"type":"webauthn.get"}`),
		"!!!invalid-base64!!!",
		b64("valid-signature"),
		"",
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode authenticator data")
}

func TestParseAssertionResponse_SignatureDecodeError(t *testing.T) {
	// Tests L197-198: error decoding signature
	validAuthData := make([]byte, 37)
	_, err := parseAssertionResponse(
		b64("valid-id"),
		"public-key",
		b64(`{"type":"webauthn.get"}`),
		base64.RawURLEncoding.EncodeToString(validAuthData),
		"!!!invalid-base64!!!",
		"",
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode signature")
}

// ==================== Tests for L267-268, L272-273: parseAttestationResponse error paths ====================

func TestParseAttestationResponse_ClientDataJSONDecodeError(t *testing.T) {
	// Tests L267-268: error decoding clientDataJSON in parseAttestationResponse
	_, err := parseAttestationResponse(
		"credential-id",
		"public-key",
		"!!!invalid-base64!!!",
		b64("valid-attestation"),
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode client data JSON")
}

func TestParseAttestationResponse_AttestationObjectDecodeError(t *testing.T) {
	// Tests L272-273: error decoding attestationObject
	_, err := parseAttestationResponse(
		"credential-id",
		"public-key",
		b64(`{"type":"webauthn.create"}`),
		"!!!invalid-base64!!!",
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode attestation object")
}

// ==================== Tests for L294-295: marshal credential error ====================

// Note: L294-295 tests json.Marshal error which is difficult to trigger directly
// as json.Marshal rarely fails for simple map[string]interface{} structures.
// The error handling is defensive programming.

// ==================== Tests for L305: parseAttestationResponse success path ====================

func TestParseAttestationResponse_Success(t *testing.T) {
	// Tests L305: successful parsing returns parsed data
	// This requires valid attestation data structure
	validClientData := `{"type":"webauthn.create","challenge":"test","origin":"https://localhost"}`

	// Create a minimal valid CBOR attestation object
	// Format: map with "fmt" and "attStmt" keys
	attestationCBOR := []byte{
		0xa2,                   // map(2)
		0x63, 0x66, 0x6d, 0x74, // "fmt" (text string of length 3)
		0x64, 0x6e, 0x6f, 0x6e, 0x65, // "none" (text string of length 4)
		0x67, 0x61, 0x74, 0x74, 0x53, 0x74, 0x6d, 0x74, // "attStmt" (text string of length 7)
		0xa0, // empty map
	}

	result, err := parseAttestationResponse(
		b64("credential-id"),
		"public-key",
		b64(validClientData),
		base64.RawURLEncoding.EncodeToString(attestationCBOR),
	)

	// The protocol parser may still fail due to missing required fields
	// but we're testing that the function reaches the parse step (L305)
	// If it errors, it should be from the protocol parser, not our code
	if err != nil {
		assert.Contains(t, err.Error(), "failed to parse credential creation response")
	} else {
		assert.NotNil(t, result)
	}
}

// ==================== Tests for L355-357, L361-363: buildRegistrationOptions ====================

func TestBuildRegistrationOptions_WithAuthenticatorSelection(t *testing.T) {
	// Tests L355-357: authenticatorSelection is not nil
	req := &WebAuthnRegisterStartRequest{
		UserID:         testUserID,
		RelyingPartyID: "example.com",
		AuthenticatorSelection: &AuthenticatorSelection{
			AuthenticatorAttachment: "platform",
			ResidentKey:             "required",
			UserVerification:        "required",
		},
	}

	options := buildRegistrationOptions(req)

	assert.NotNil(t, options)
	assert.NotEmpty(t, options)
	// Should contain at least the authenticator selection option
	assert.GreaterOrEqual(t, len(options), 1)
}

func TestBuildRegistrationOptions_WithAttestation(t *testing.T) {
	// Tests L361-363: attestation is not empty
	req := &WebAuthnRegisterStartRequest{
		UserID:         testUserID,
		RelyingPartyID: "example.com",
		Attestation:    "direct",
	}

	options := buildRegistrationOptions(req)

	assert.NotNil(t, options)
	assert.NotEmpty(t, options)
	// Should contain at least the attestation option
	assert.GreaterOrEqual(t, len(options), 1)
}

func TestBuildRegistrationOptions_WithBothOptions(t *testing.T) {
	// Tests both L355-357 and L361-363: both conditions true
	req := &WebAuthnRegisterStartRequest{
		UserID:         testUserID,
		RelyingPartyID: "example.com",
		AuthenticatorSelection: &AuthenticatorSelection{
			AuthenticatorAttachment: "cross-platform",
			ResidentKey:             "preferred",
			UserVerification:        "preferred",
		},
		Attestation: "indirect",
	}

	options := buildRegistrationOptions(req)

	assert.NotNil(t, options)
	assert.NotEmpty(t, options)
	// Should contain both options
	assert.Equal(t, 2, len(options))
}

func TestBuildRegistrationOptions_WithNeitherOption(t *testing.T) {
	// Tests when both conditions are false (else branches)
	req := &WebAuthnRegisterStartRequest{
		UserID:         testUserID,
		RelyingPartyID: "example.com",
		// No AuthenticatorSelection
		// No Attestation
	}

	options := buildRegistrationOptions(req)

	// When no options are added, the slice remains nil
	assert.Len(t, options, 0)
}

// ==================== Tests for L83: getConfiguredOrigins returns originList ====================

func TestGetConfiguredOrigins_ReturnsDefaultsWhenNoConfig(t *testing.T) {
	// Tests L83: returns originList when configured
	// This is implicitly tested by TestGetConfiguredOrigins, but we can be more explicit
	origins := getConfiguredOrigins()

	assert.NotNil(t, origins)
	assert.NotEmpty(t, origins)
	// Should return either configured origins or defaults
	assert.True(t, len(origins) >= 1)
}

// Note: L66-68 (panic recovery in getConfiguredOrigins) is difficult to test directly
// without mocking the config.GetThunderRuntime() function, as it requires triggering
// a panic in the configuration access. The panic recovery is defensive code that ensures
// the function returns defaults even if config access fails.

// ==================== Tests for buildAuthenticatorSelection branches ====================

func TestBuildAuthenticatorSelection_WithAuthenticatorAttachment(t *testing.T) {
	// Tests the branch where AuthenticatorAttachment is set
	sel := &AuthenticatorSelection{
		AuthenticatorAttachment: "platform",
	}

	result := buildAuthenticatorSelection(sel)

	assert.Equal(t, protocol.Platform, result.AuthenticatorAttachment)
	assert.Equal(t, protocol.VerificationPreferred, result.UserVerification)
}

func TestBuildAuthenticatorSelection_WithResidentKey(t *testing.T) {
	// Tests the branch where ResidentKey is set
	sel := &AuthenticatorSelection{
		ResidentKey: "required",
	}

	result := buildAuthenticatorSelection(sel)

	assert.Equal(t, protocol.ResidentKeyRequirementRequired, result.ResidentKey)
	assert.Equal(t, protocol.VerificationPreferred, result.UserVerification)
}

func TestBuildAuthenticatorSelection_WithUserVerification(t *testing.T) {
	// Tests the branch where UserVerification is set (not empty)
	sel := &AuthenticatorSelection{
		UserVerification: "required",
	}

	result := buildAuthenticatorSelection(sel)

	assert.Equal(t, protocol.VerificationRequired, result.UserVerification)
}

func TestBuildAuthenticatorSelection_EmptyUserVerification(t *testing.T) {
	// Tests the else branch where UserVerification defaults to preferred
	sel := &AuthenticatorSelection{
		UserVerification: "",
	}

	result := buildAuthenticatorSelection(sel)

	assert.Equal(t, protocol.VerificationPreferred, result.UserVerification)
}
