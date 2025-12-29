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

import "github.com/asgardeo/thunder/internal/system/error/serviceerror"

// Client errors for WebAuthn authentication service

var (
	// ErrorEmptyUserIdentifier is returned when both userID and username are empty.
	ErrorEmptyUserIdentifier = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "WEBAUTHN-1001",
		Error:            "Empty user identifier",
		ErrorDescription: "Either user ID or username must be provided",
	}
	// ErrorEmptyRelyingPartyID is returned when the relying party ID is empty.
	ErrorEmptyRelyingPartyID = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "WEBAUTHN-1002",
		Error:            "Empty relying party ID",
		ErrorDescription: "The relying party ID is required for WebAuthn operations",
	}
	// ErrorEmptyCredentialID is returned when the credential ID is empty.
	ErrorEmptyCredentialID = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "WEBAUTHN-1003",
		Error:            "Empty credential ID",
		ErrorDescription: "The credential ID is required to complete WebAuthn authentication",
	}
	// ErrorEmptyCredentialType is returned when the credential type is empty.
	ErrorEmptyCredentialType = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "WEBAUTHN-1004",
		Error:            "Empty credential type",
		ErrorDescription: "The credential type is required to complete WebAuthn authentication",
	}
	// ErrorInvalidAuthenticatorResponse is returned when the authenticator response is invalid.
	ErrorInvalidAuthenticatorResponse = serviceerror.ServiceError{
		Type:  serviceerror.ClientErrorType,
		Code:  "WEBAUTHN-1005",
		Error: "Invalid authenticator response",
		ErrorDescription: "The authenticator response is missing required fields " +
			"(clientDataJSON, authenticatorData, or signature)",
	}
	// ErrorEmptySessionToken is returned when the session token is empty.
	ErrorEmptySessionToken = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "WEBAUTHN-1006",
		Error:            "Empty session token",
		ErrorDescription: "The session token is required to complete WebAuthn authentication",
	}
	// ErrorInvalidFinishData is returned when the finish data is nil.
	ErrorInvalidFinishData = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "WEBAUTHN-1007",
		Error:            "Invalid finish data",
		ErrorDescription: "The authentication finish data cannot be null",
	}
	// ErrorInvalidChallenge is returned when the challenge validation fails.
	ErrorInvalidChallenge = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "WEBAUTHN-1008",
		Error:            "Invalid challenge",
		ErrorDescription: "The challenge in the response does not match the expected challenge",
	}
	// ErrorInvalidSignature is returned when signature verification fails.
	ErrorInvalidSignature = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "WEBAUTHN-1009",
		Error:            "Invalid signature",
		ErrorDescription: "The WebAuthn signature verification failed",
	}
	// ErrorCredentialNotFound is returned when the credential is not found.
	ErrorCredentialNotFound = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "WEBAUTHN-1010",
		Error:            "Credential not found",
		ErrorDescription: "The specified credential was not found for the user",
	}
	// ErrorInvalidAttestationResponse is returned when the attestation response is invalid.
	ErrorInvalidAttestationResponse = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "WEBAUTHN-1011",
		Error:            "Invalid attestation response",
		ErrorDescription: "The attestation response is missing required fields (clientDataJSON or attestationObject)",
	}
	// ErrorUserNotFound is returned when the user is not found.
	ErrorUserNotFound = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "WEBAUTHN-1012",
		Error:            "User not found",
		ErrorDescription: "The specified user was not found",
	}
	// ErrorInvalidSessionToken is returned when the session token is invalid.
	ErrorInvalidSessionToken = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "WEBAUTHN-1013",
		Error:            "Invalid session token",
		ErrorDescription: "The session token is invalid or malformed",
	}
	// ErrorSessionExpired is returned when the session has expired.
	ErrorSessionExpired = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "WEBAUTHN-1014",
		Error:            "Session expired",
		ErrorDescription: "The WebAuthn session has expired. Please start a new registration or authentication",
	}
	// ErrorNoCredentialsFound is returned when no credentials are found for the user.
	ErrorNoCredentialsFound = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "WEBAUTHN-1015",
		Error:            "No credentials found",
		ErrorDescription: "No WebAuthn credentials found for the user. Please register a credential first",
	}
)

// Server errors for WebAuthn authentication service

var (
	// ErrorNotImplemented is returned when WebAuthn authentication is not yet implemented.
	ErrorNotImplemented = serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "WEBAUTHN-5001",
		Error:            "WebAuthn authentication not yet implemented",
		ErrorDescription: "WebAuthn/Passkey authentication feature is not yet implemented",
	}
	// ErrorInternalServerError is returned when an internal server error occurs.
	ErrorInternalServerError = serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "WEBAUTHN-5002",
		Error:            "Internal server error",
		ErrorDescription: "An internal server error occurred while processing the request",
	}
)
