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

package userprovider

import (
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
)

// ErrorCode represents a user provider error code.
type ErrorCode string

// Predefined error codes for user provider.
// Implementors should use these constants with NewError().
const (
	CodeSystemError              ErrorCode = "UP-0001"
	CodeUserNotFound             ErrorCode = "UP-0002"
	CodeInvalidRequestFormat     ErrorCode = "UP-0003"
	CodeOrganizationUnitMismatch ErrorCode = "UP-0004"
	CodeAttributeConflict        ErrorCode = "UP-0005"
	CodeMissingRequiredFields    ErrorCode = "UP-0006"
	CodeMissingCredentials       ErrorCode = "UP-0007"
	CodeNotImplemented           ErrorCode = "UP-0008"
)

// errorDefinition holds the fixed Type and Error message for each error code.
type errorDefinition struct {
	Type  serviceerror.ServiceErrorType
	Error string
}

// validErrorCodes maps error codes to their fixed Type and Error message.
var validErrorCodes = map[ErrorCode]errorDefinition{
	CodeSystemError:              {serviceerror.ServerErrorType, "System error"},
	CodeUserNotFound:             {serviceerror.ClientErrorType, "User not found"},
	CodeInvalidRequestFormat:     {serviceerror.ClientErrorType, "Invalid request format"},
	CodeOrganizationUnitMismatch: {serviceerror.ClientErrorType, "Organization unit mismatch"},
	CodeAttributeConflict:        {serviceerror.ClientErrorType, "Attribute conflict"},
	CodeMissingRequiredFields:    {serviceerror.ClientErrorType, "Missing required fields"},
	CodeMissingCredentials:       {serviceerror.ClientErrorType, "Missing credentials"},
	CodeNotImplemented:           {serviceerror.ServerErrorType, "Not implemented"},
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

// User provider errors.
var (
	// ErrorSystemError is the error returned when a system error occurs.
	ErrorSystemError = serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "UP-0001",
		Error:            "System error",
		ErrorDescription: "An unexpected system error occurred",
	}

	// ErrorUserNotFound is the error returned when the user is not found.
	ErrorUserNotFound = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "UP-0002",
		Error:            "User not found",
		ErrorDescription: "The requested user could not be found",
	}

	// ErrorInvalidRequestFormat is the error returned when the request format is invalid.
	ErrorInvalidRequestFormat = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "UP-0003",
		Error:            "Invalid request format",
		ErrorDescription: "The request format is invalid",
	}

	// ErrorOrganizationUnitMismatch is the error returned when organization unit does not match.
	ErrorOrganizationUnitMismatch = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "UP-0004",
		Error:            "Organization unit mismatch",
		ErrorDescription: "The organization unit does not match",
	}

	// ErrorAttributeConflict is the error returned when there is an attribute conflict.
	ErrorAttributeConflict = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "UP-0005",
		Error:            "Attribute conflict",
		ErrorDescription: "One or more attributes conflict with existing values",
	}

	// ErrorMissingRequiredFields is the error returned when required fields are missing.
	ErrorMissingRequiredFields = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "UP-0006",
		Error:            "Missing required fields",
		ErrorDescription: "One or more required fields are missing",
	}

	// ErrorMissingCredentials is the error returned when credentials are missing.
	ErrorMissingCredentials = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "UP-0007",
		Error:            "Missing credentials",
		ErrorDescription: "Credentials are missing or invalid",
	}

	// ErrorNotImplemented is the error returned when a feature is not implemented.
	ErrorNotImplemented = serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "UP-0008",
		Error:            "Not implemented",
		ErrorDescription: "This feature is not implemented",
	}
)
