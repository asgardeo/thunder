/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
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

// Package constants provides constants for the authentication module.
package constants

import (
	"github.com/asgardeo/thunder/internal/system/error/apierror"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
)

// Client error structs

// APIErrorJSONDecodeError defines the error response for json decode errors.
var APIErrorJSONDecodeError = apierror.ErrorResponse{
	Code:        "ANE-1001",
	Message:     "Invalid request payload",
	Description: "Failed to decode request payload",
}

// ErrorInvalidAuthFlow defines the service error for invalid authentication flow.
var ErrorInvalidAuthFlow = serviceerror.ServiceError{
	Code:             "ANE-1002",
	Type:             serviceerror.ClientErrorType,
	Error:            "Invalid authentication flow",
	ErrorDescription: "The authentication flow is invalid or malformed",
}

// ErrorSessionNotFound defines the service error for session not found errors.
var ErrorSessionNotFound = serviceerror.ServiceError{
	Code:             "ANE-1003",
	Type:             serviceerror.ClientErrorType,
	Error:            "Session not found",
	ErrorDescription: "The session data could not be found for the provided key",
}

// ErrorAppIDNotFound defines the service error for application ID not found errors.
var ErrorAppIDNotFound = serviceerror.ServiceError{
	Code:             "ANE-1004",
	Type:             serviceerror.ClientErrorType,
	Error:            "Application ID not found",
	ErrorDescription: "The application ID could not be found in the session data",
}

// ErrorFlowExecutionClientError defines the service error for client-side flow execution errors.
var ErrorFlowExecutionClientError = serviceerror.ServiceError{
	Code:             "ANE-1005",
	Type:             serviceerror.ClientErrorType,
	Error:            "Flow execution error",
	ErrorDescription: "An error occurred while executing the authentication flow",
}

// Server error structs

// ErrorFlowExecutionServerError defines the service error for server-side flow execution errors.
var ErrorFlowExecutionServerError = serviceerror.ServiceError{
	Code:             "ANE-5001",
	Type:             serviceerror.ServerErrorType,
	Error:            "Flow execution error",
	ErrorDescription: "An error occurred while executing the authentication flow",
}

// ErrorWhileDecodingFlowAssertion defines the service error for errors while decoding the flow assertion.
var ErrorWhileDecodingFlowAssertion = serviceerror.ServiceError{
	Code:             "ANE-5002",
	Type:             serviceerror.ServerErrorType,
	Error:            "Something went wrong while decoding the flow assertion",
	ErrorDescription: "The flow assertion could not be decoded",
}

// ErrorConstructingRedirectURI defines the service error for errors while constructing the redirect URI.
var ErrorConstructingRedirectURI = serviceerror.ServiceError{
	Code:             "ANE-5004",
	Type:             serviceerror.ServerErrorType,
	Error:            "Something went wrong while constructing the redirect URI",
	ErrorDescription: "Failed to construct the redirect URI for the authentication flow",
}
