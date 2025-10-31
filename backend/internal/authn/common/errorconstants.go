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

package common

import (
	"github.com/asgardeo/thunder/internal/system/error/apierror"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
)

// API errors

// APIErrorInvalidRequestFormat is returned when the request body is malformed.
var APIErrorInvalidRequestFormat = apierror.ErrorResponse{
	Code:        "AUTHN-1000",
	Message:     "Invalid request format",
	Description: "The request body is malformed or contains invalid data",
}

// Client errors for the service
var (
	// ErrorInvalidIDPID is the error returned when the provided IDP ID is invalid.
	ErrorInvalidIDPID = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "AUTHN-1001",
		Error:            "Invalid identity provider ID",
		ErrorDescription: "The provided identity provider ID is invalid or empty",
	}
	// ErrorClientErrorWhileRetrievingIDP is the error returned when there is a client error while retrieving the IDP.
	ErrorClientErrorWhileRetrievingIDP = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "AUTHN-1002",
		Error:            "Error retrieving identity provider",
		ErrorDescription: "An error occurred while retrieving the identity provider",
	}
	// ErrorInvalidIDPType is the error returned when the provided IDP type is invalid.
	ErrorInvalidIDPType = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "AUTHN-1003",
		Error:            "Invalid identity provider type",
		ErrorDescription: "The requested identity provider type is invalid",
	}
	// ErrorEmptySessionToken is the error returned when the provided session token is invalid.
	ErrorEmptySessionToken = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "AUTHN-1004",
		Error:            "Empty session token",
		ErrorDescription: "The provided session token is empty",
	}
	// ErrorEmptyAuthCode is the error returned when the provided authorization code is empty.
	ErrorEmptyAuthCode = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "AUTHN-1005",
		Error:            "Empty authorization code",
		ErrorDescription: "The provided authorization code is empty",
	}
	// ErrorInvalidSessionToken is the error returned when the provided session token is invalid.
	ErrorInvalidSessionToken = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "AUTHN-1006",
		Error:            "Invalid session token",
		ErrorDescription: "The provided session token is invalid or has expired",
	}
	// ErrorSubClaimNotFound is the error returned when the 'sub' claim is not found in the ID token.
	ErrorSubClaimNotFound = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "AUTHN-1007",
		Error:            "user subject not found",
		ErrorDescription: "The 'sub' claim is not found in the ID token claims",
	}
	// ErrorUserNotFound is the error when no user is found with the provided attributes.
	ErrorUserNotFound = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "AUTHN-1008",
		Error:            "User not found",
		ErrorDescription: "No user found with the provided attributes",
	}
	// ErrorInvalidAssertion is the error returned when the provided assertion token is invalid.
	ErrorInvalidAssertion = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "AUTHN-1009",
		Error:            "Invalid assertion",
		ErrorDescription: "The provided assertion token is invalid",
	}
	// ErrorAssertionSubjectMismatch is the error returned when the assertion subject doesn't match
	// the authenticated user.
	ErrorAssertionSubjectMismatch = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "AUTHN-1010",
		Error:            "Assertion subject mismatch",
		ErrorDescription: "The subject in the assertion does not match the authenticated user",
	}
)

// Server errors for the service

// ErrorInternalServerError is returned when an unexpected error occurs on the server.
var ErrorInternalServerError = serviceerror.ServiceError{
	Type:             serviceerror.ServerErrorType,
	Code:             "AUTHN-5000",
	Error:            "Internal server error",
	ErrorDescription: "An unexpected error occurred while processing the request",
}
