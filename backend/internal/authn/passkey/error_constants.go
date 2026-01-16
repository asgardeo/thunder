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

import "github.com/asgardeo/thunder/internal/system/error/serviceerror"

// Client errors for passkey authentication service

var (
	// ErrorEmptyUserIdentifier is returned when both userID and username are empty.
	ErrorEmptyUserIdentifier = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "PSK-1001",
		Error:            "Empty user identifier",
		ErrorDescription: "Either user ID or username must be provided",
	}
	// ErrorEmptyRelyingPartyID is returned when the relying party ID is empty.
	ErrorEmptyRelyingPartyID = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "PSK-1002",
		Error:            "Empty relying party ID",
		ErrorDescription: "The relying party ID is required",
	}
	// ErrorEmptyCredentialID is returned when the credential ID is empty.
	ErrorEmptyCredentialID = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "PSK-1003",
		Error:            "Empty credential ID",
		ErrorDescription: "The credential ID is required",
	}
	// ErrorEmptyCredentialType is returned when the credential type is empty.
	ErrorEmptyCredentialType = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "PSK-1004",
		Error:            "Empty credential type",
		ErrorDescription: "The credential type is required",
	}
	// ErrorInvalidAuthenticatorResponse is returned when the authenticator response is invalid.
	ErrorInvalidAuthenticatorResponse = serviceerror.ServiceError{
		Type:  serviceerror.ClientErrorType,
		Code:  "PSK-1005",
		Error: "Invalid authenticator response",
		ErrorDescription: "The authenticator response is missing required fields " +
			"(clientDataJSON, authenticatorData, or signature)",
	}
	// ErrorEmptySessionToken is returned when the session token is empty.
	ErrorEmptySessionToken = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "PSK-1006",
		Error:            "Empty session token",
		ErrorDescription: "The session token is required",
	}
	// ErrorInvalidFinishData is returned when the finish data is nil.
	ErrorInvalidFinishData = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "PSK-1007",
		Error:            "Invalid finish data",
		ErrorDescription: "The finish data cannot be null",
	}
	// ErrorInvalidChallenge is returned when the challenge validation fails.
	ErrorInvalidChallenge = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "PSK-1008",
		Error:            "Invalid challenge",
		ErrorDescription: "The challenge in the response does not match the expected challenge",
	}
	// ErrorInvalidSignature is returned when signature verification fails.
	ErrorInvalidSignature = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "PSK-1009",
		Error:            "Invalid signature",
		ErrorDescription: "The signature verification failed",
	}
	// ErrorCredentialNotFound is returned when the credential is not found.
	ErrorCredentialNotFound = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "PSK-1010",
		Error:            "WebauthnCredential not found",
		ErrorDescription: "The specified credential was not found for the user",
	}
	// ErrorInvalidAttestationResponse is returned when the attestation response is invalid.
	ErrorInvalidAttestationResponse = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "PSK-1011",
		Error:            "Invalid attestation response",
		ErrorDescription: "The attestation response is missing required fields (clientDataJSON or attestationObject)",
	}
	// ErrorUserNotFound is returned when the user is not found.
	ErrorUserNotFound = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "PSK-1012",
		Error:            "User not found",
		ErrorDescription: "The specified user was not found",
	}
	// ErrorInvalidSessionToken is returned when the session token is invalid.
	ErrorInvalidSessionToken = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "PSK-1013",
		Error:            "Invalid session token",
		ErrorDescription: "The session token is invalid or malformed",
	}
	// ErrorSessionExpired is returned when the session has expired.
	ErrorSessionExpired = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "PSK-1014",
		Error:            "Session expired",
		ErrorDescription: "The session has expired. Please start a new session",
	}
	// ErrorNoCredentialsFound is returned when no credentials are found for the user.
	ErrorNoCredentialsFound = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "PSK-1015",
		Error:            "No credentials found",
		ErrorDescription: "No credentials found for the user. Please register a credential first",
	}
)
