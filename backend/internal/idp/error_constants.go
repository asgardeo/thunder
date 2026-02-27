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

package idp

import (
	"errors"

	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
)

// ErrIDPNotFound is returned when the IdP is not found in the system.
var ErrIDPNotFound = errors.New("IdP not found")

// Client errors for identity provider operations.
var (
	// ErrorIDPNotFound is the error returned when an identity provider is not found.
	ErrorIDPNotFound = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "IDP-1001",
		Error:            "Identity provider not found",
		ErrorDescription: "The requested identity provider could not be found",
	}
	// ErrorInvalidIDPID is the error returned when an invalid identity provider ID is provided.
	ErrorInvalidIDPID = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "IDP-1002",
		Error:            "Invalid identity provider ID",
		ErrorDescription: "The provided identity provider ID is invalid or empty",
	}
	// ErrorInvalidIDPName is the error returned when an invalid identity provider name is provided.
	ErrorInvalidIDPName = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "IDP-1003",
		Error:            "Invalid identity provider name",
		ErrorDescription: "The provided identity provider name is invalid or empty",
	}
	// ErrorInvalidIDPType is the error returned when an invalid identity provider type is provided.
	ErrorInvalidIDPType = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "IDP-1004",
		Error:            "Invalid identity provider type",
		ErrorDescription: "The provided identity provider type is invalid or empty",
	}
	// ErrorIDPAlreadyExists is the error returned when an identity provider with the same name already exists.
	ErrorIDPAlreadyExists = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "IDP-1005",
		Error:            "Identity provider already exists",
		ErrorDescription: "An identity provider with the same name already exists",
	}
	// ErrorInvalidIDPProperty is the error returned when an invalid identity provider property is provided.
	ErrorInvalidIDPProperty = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "IDP-1006",
		Error:            "Invalid identity provider property",
		ErrorDescription: "One or more identity provider properties are invalid or empty",
	}
	// ErrorUnsupportedIDPProperty is the error returned when an unsupported identity provider property is provided.
	ErrorUnsupportedIDPProperty = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "IDP-1007",
		Error:            "Unsupported identity provider property",
		ErrorDescription: "One or more identity provider properties are not supported",
	}
	// ErrorIDPNil is the error returned when the identity provider object is nil.
	ErrorIDPNil = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "IDP-1008",
		Error:            "Identity provider cannot be null",
		ErrorDescription: "The identity provider object cannot be null or empty",
	}
	// ErrorInvalidRequestFormat is the error returned when the request format is invalid.
	ErrorInvalidRequestFormat = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "IDP-1009",
		Error:            "Invalid request format",
		ErrorDescription: "The request body is malformed or contains invalid data",
	}
	// ErrorIDPDeclarativeReadOnly is the error returned when attempting to modify a declarative (immutable) IDP.
	ErrorIDPDeclarativeReadOnly = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "IDP-1010",
		Error:            "Identity provider is immutable",
		ErrorDescription: "The requested identity provider is declarative and cannot be modified or deleted",
	}
)
