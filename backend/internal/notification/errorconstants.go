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

package notification

import "github.com/asgardeo/thunder/internal/system/error/serviceerror"

// Client errors for notification sender operations.
var (
	// ErrorSenderNotFound is the error returned when a notification sender is not found.
	ErrorSenderNotFound = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "MNS-1001",
		Error:            "Sender not found",
		ErrorDescription: "The requested notification sender could not be found",
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
		Error:            "Invalid sender name",
		ErrorDescription: "The provided sender name is invalid",
	}
	// ErrorInvalidProvider is the error returned when an invalid provider is specified.
	ErrorInvalidProvider = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "MNS-1004",
		Error:            "Invalid notification provider",
		ErrorDescription: "The specified notification provider is invalid or unsupported",
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
	// ErrorInvalidSenderType is the error returned when an invalid sender type is provided.
	ErrorInvalidSenderType = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "MNS-1007",
		Error:            "Invalid sender type",
		ErrorDescription: "The provided sender type is invalid or unsupported",
	}
	// ErrorSenderTypeUpdateNotAllowed is the error when trying to update the sender type.
	ErrorSenderTypeUpdateNotAllowed = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "MNS-1008",
		Error:            "Update not allowed",
		ErrorDescription: "Updating the sender type is not allowed",
	}
	// ErrorInvalidSenderProperty is the error returned when an invalid sender property is provided.
	ErrorInvalidSenderProperty = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "MNS-1016",
		Error:            "Invalid sender property",
		ErrorDescription: "One or more sender properties are invalid or empty",
	}
	// ErrorUnsupportedSenderProperty is the error returned when an unsupported sender property is provided.
	ErrorUnsupportedSenderProperty = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "MNS-1017",
		Error:            "Unsupported sender property",
		ErrorDescription: "One or more sender properties are not supported for the specified provider",
	}
	// ErrorRequestedSenderIsNotOfExpectedType is the error when the requested sender is not of the expected type.
	ErrorRequestedSenderIsNotOfExpectedType = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "MNS-1009",
		Error:            "Sender type mismatch",
		ErrorDescription: "The requested sender is not of the expected type",
	}
	// ErrorInvalidRecipient is the error returned when an invalid recipient is provided.
	ErrorInvalidRecipient = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "MNS-1010",
		Error:            "Invalid recipient",
		ErrorDescription: "The provided recipient is invalid",
	}
	// ErrorInvalidChannel is the error returned when an invalid channel is provided.
	ErrorInvalidChannel = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "MNS-1011",
		Error:            "Invalid channel",
		ErrorDescription: "The provided channel is invalid",
	}
	// ErrorUnsupportedChannel is the error returned when an unsupported channel is provided.
	ErrorUnsupportedChannel = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "MNS-1012",
		Error:            "Unsupported channel",
		ErrorDescription: "The provided channel is not supported",
	}
	// ErrorInvalidOTP is the error returned when an invalid OTP is provided.
	ErrorInvalidOTP = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "MNS-1013",
		Error:            "Invalid OTP",
		ErrorDescription: "The provided OTP is invalid",
	}
	// ErrorInvalidSessionToken is the error returned when an invalid session token is provided.
	ErrorInvalidSessionToken = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "MNS-1014",
		Error:            "Invalid session token",
		ErrorDescription: "The provided session token is invalid, malformed, or expired",
	}
	// ErrorClientErrorWhileRetrievingMessageClient is the error returned when a client error occurs
	// while retrieving the message client.
	ErrorClientErrorWhileRetrievingMessageClient = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "MNS-1015",
		Error:            "Error while retrieving message client",
		ErrorDescription: "An error occurred while retrieving the message client",
	}
)

// Server errors for notification sender operations.
var (
	// ErrorInternalServerError is the error returned when an internal server error occurs.
	ErrorInternalServerError = serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "MNS-5000",
		Error:            "Internal server error",
		ErrorDescription: "An unexpected error occurred while processing the request",
	}
)
