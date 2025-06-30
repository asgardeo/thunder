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

package constants

import "github.com/asgardeo/thunder/internal/system/error/serviceerror"

// Client errors for message notification sender operations.
var (
	// ErrorSenderNotFound is the error returned when a notification sender is not found.
	ErrorSenderNotFound = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "MNS-1001",
		Error:            "Sender not found",
		ErrorDescription: "The requested message notification sender could not be found",
	}
	// ErrorInvalidSenderID is the error returned when an invalid sender ID is provided.
	ErrorInvalidSenderID = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "MNS-1002",
		Error:            "Invalid sender ID",
		ErrorDescription: "The provided sender ID is invalid",
	}
	// ErrorInvalidSenderName is the error returned when an invalid sender name is provided.
	ErrorInvalidSenderName = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "MNS-1003",
		Error:            "Invalid sender Name",
		ErrorDescription: "The provided sender name is invalid",
	}
	// ErrorInvalidProvider is the error returned when an unsupported provider is specified.
	ErrorInvalidProvider = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "MNS-1004",
		Error:            "Invalid provider",
		ErrorDescription: "The specified provider is not supported",
	}
	// ErrorDuplicateSenderName is the error returned when a sender with the same name already exists.
	ErrorDuplicateSenderName = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "MNS-1005",
		Error:            "Duplicate sender name",
		ErrorDescription: "A sender with the same name already exists",
	}
	// ErrorInvalidRequestFormat is the error returned when the request format is invalid.
	ErrorInvalidRequestFormat = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "MNS-1006",
		Error:            "Invalid request format",
		ErrorDescription: "The request body is malformed or contains invalid data",
	}
)

// Server errors for message notification sender operations.
var (
	// ErrorInternalServerError is the error returned when an internal server error occurs.
	ErrorInternalServerError = serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "MNS-5001",
		Error:            "Internal server error",
		ErrorDescription: "An unexpected error occurred while processing the request",
	}
)
