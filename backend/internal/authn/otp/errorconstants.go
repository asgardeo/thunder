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

package otp

import "github.com/asgardeo/thunder/internal/system/error/serviceerror"

// Client errors for OTP authentication service
var (
	// ErrorInvalidSenderID is the error returned when the provided sender ID is invalid.
	ErrorInvalidSenderID = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "AUTHN-OTP-1001",
		Error:            "Invalid sender ID",
		ErrorDescription: "The provided sender ID is invalid or empty",
	}
	// ErrorInvalidRecipient is the error returned when the provided recipient is invalid.
	ErrorInvalidRecipient = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "AUTHN-OTP-1002",
		Error:            "Invalid recipient",
		ErrorDescription: "The provided recipient is invalid or empty",
	}
	// ErrorUnsupportedChannel is the error returned when the provided channel is not supported.
	ErrorUnsupportedChannel = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "AUTHN-OTP-1003",
		Error:            "Unsupported channel",
		ErrorDescription: "The provided channel is not supported for OTP authentication",
	}
	// ErrorInvalidSessionToken is the error returned when the provided session token is invalid.
	ErrorInvalidSessionToken = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "AUTHN-OTP-1004",
		Error:            "Invalid session token",
		ErrorDescription: "The provided session token is invalid or empty",
	}
	// ErrorInvalidOTP is the error returned when the provided OTP is invalid.
	ErrorInvalidOTP = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "AUTHN-OTP-1005",
		Error:            "Invalid OTP",
		ErrorDescription: "The provided OTP is invalid or empty",
	}
	// ErrorUserNotFound is the error returned when no user is found for the recipient.
	ErrorUserNotFound = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "AUTHN-OTP-1006",
		Error:            "User not found",
		ErrorDescription: "No user found for the provided recipient",
	}
	// ErrorClientErrorWhileSendingOTP is the error returned when there is a client error while sending OTP.
	ErrorClientErrorWhileSendingOTP = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "AUTHN-OTP-1007",
		Error:            "Error processing OTP",
		ErrorDescription: "An error occurred while processing the OTP request",
	}
	// ErrorClientErrorWhileResolvingUser is the error returned when there is a client error while resolving the user.
	ErrorClientErrorWhileResolvingUser = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "AUTHN-OTP-1008",
		Error:            "Error resolving user",
		ErrorDescription: "An error occurred while resolving the user for the receipient",
	}
)

// Server errors for OTP authentication service
var (
	// ErrorInternalServerError is returned when an unexpected error occurs on the server.
	ErrorInternalServerError = serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "AUTHN-OTP-5000",
		Error:            "Internal server error",
		ErrorDescription: "An unexpected error occurred while processing the request",
	}
)
