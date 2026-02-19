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

package usertype

import (
	"errors"

	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
)

// Client errors for user type management operations.
var (
	// ErrorInvalidRequestFormat is the error returned when the request format is invalid.
	ErrorInvalidRequestFormat = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "USRS-1001",
		Error:            "Invalid request format",
		ErrorDescription: "The request body is malformed or contains invalid data",
	}
	// ErrorUserTypeNotFound is the error returned when a user type is not found.
	ErrorUserTypeNotFound = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "USRS-1002",
		Error:            "User type not found",
		ErrorDescription: "The user type with the specified id does not exist",
	}
	// ErrorUserTypeNameConflict is the error returned when user type name already exists.
	ErrorUserTypeNameConflict = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "USRS-1003",
		Error:            "User type name conflict",
		ErrorDescription: "A user type with the same name already exists",
	}
	// ErrorInvalidUserTypeRequest is the error returned when user type request is invalid.
	ErrorInvalidUserTypeRequest = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "USRS-1004",
		Error:            "Invalid user type request",
		ErrorDescription: "The user type request contains invalid or missing required fields",
	}
	// ErrorInvalidLimit is the error returned when limit parameter is invalid.
	ErrorInvalidLimit = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "USRS-1005",
		Error:            "Invalid pagination parameter",
		ErrorDescription: "The limit parameter must be a positive integer",
	}
	// ErrorInvalidOffset is the error returned when offset parameter is invalid.
	ErrorInvalidOffset = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "USRS-1006",
		Error:            "Invalid pagination parameter",
		ErrorDescription: "The offset parameter must be a non-negative integer",
	}
	// ErrorUserValidationFailed is the error returned when user attributes do not conform to the schema.
	ErrorUserValidationFailed = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "USRS-1007",
		Error:            "User validation failed",
		ErrorDescription: "User attributes do not conform to the required schema",
	}
)

// Server errors for user type management operations.
var (
	// ErrorInternalServerError is the error returned when an internal server error occurs.
	ErrorInternalServerError = serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "USRS-5000",
		Error:            "Internal server error",
		ErrorDescription: "An unexpected error occurred while processing the request",
	}
)

// Error variables for user type operations.
var (
	// ErrUserTypeNotFound is returned when the user type is not found in the system.
	ErrUserTypeNotFound = errors.New("user type not found")

	// ErrUserTypeAlreadyExists is returned when a user type with the same name already exists.
	ErrUserTypeAlreadyExists = errors.New("usertype already exists")

	// ErrInvalidSchemaDefinition is returned when the schema definition is invalid.
	ErrInvalidSchemaDefinition = errors.New("invalid schema definition")
)
