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

package authnprovider

import (
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
)

// ErrorCode represents an authn provider error code.
type ErrorCode string

// Predefined error codes for authn provider.
// Implementors should use these constants with NewError().
const (
	CodeSystemError          ErrorCode = "AUP-0001"
	CodeAuthenticationFailed ErrorCode = "AUP-0002"
	CodeUserNotFound         ErrorCode = "AUP-0003"
	CodeInvalidToken         ErrorCode = "AUP-0004"
	CodeNotImplemented       ErrorCode = "AUP-0005"
)

// errorDefinition holds the fixed Type and Error message for each error code.
type errorDefinition struct {
	Type  serviceerror.ServiceErrorType
	Error string
}

// validErrorCodes maps error codes to their fixed Type and Error message.
var validErrorCodes = map[ErrorCode]errorDefinition{
	CodeSystemError:          {serviceerror.ServerErrorType, "System error"},
	CodeAuthenticationFailed: {serviceerror.ClientErrorType, "Authentication failed"},
	CodeUserNotFound:         {serviceerror.ClientErrorType, "User not found"},
	CodeInvalidToken:         {serviceerror.ClientErrorType, "Invalid token"},
	CodeNotImplemented:       {serviceerror.ServerErrorType, "Not implemented"},
}

// IsValid checks if the error code is a valid predefined code.
func (ec ErrorCode) IsValid() bool {
	_, ok := validErrorCodes[ec]
	return ok
}

// NewError creates a ServiceError with the given code and description.
// If the code is not valid (not defined in validErrorCodes), it defaults to CodeSystemError.
// The Type and Error fields are set based on the predefined values for the code.
func NewError(code ErrorCode, description string) *serviceerror.ServiceError {
	if !code.IsValid() {
		code = CodeSystemError
	}

	def := validErrorCodes[code]
	return &serviceerror.ServiceError{
		Type:             def.Type,
		Code:             string(code),
		Error:            def.Error,
		ErrorDescription: description,
	}
}

// Authentication provider errors.
var (
	// ErrorSystemError is the error returned when a system error occurs.
	ErrorSystemError = serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "AUP-0001",
		Error:            "System error",
		ErrorDescription: "An unexpected system error occurred",
	}

	// ErrorAuthenticationFailed is the error returned when authentication fails.
	ErrorAuthenticationFailed = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "AUP-0002",
		Error:            "Authentication failed",
		ErrorDescription: "The provided credentials are invalid",
	}

	// ErrorUserNotFound is the error returned when the user is not found.
	ErrorUserNotFound = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "AUP-0003",
		Error:            "User not found",
		ErrorDescription: "The requested user could not be found",
	}

	// ErrorInvalidToken is the error returned when the token is invalid.
	ErrorInvalidToken = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "AUP-0004",
		Error:            "Invalid token",
		ErrorDescription: "The provided token is invalid or expired",
	}

	// ErrorNotImplemented is the error returned when a feature is not implemented.
	ErrorNotImplemented = serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "AUP-0005",
		Error:            "Not implemented",
		ErrorDescription: "This feature is not implemented",
	}
)
