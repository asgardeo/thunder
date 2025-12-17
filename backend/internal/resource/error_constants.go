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

package resource

import (
	"errors"

	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
)

// Client errors for resource management operations.
var (
	// ErrorInvalidRequestFormat is returned when the request format is invalid.
	ErrorInvalidRequestFormat = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "RES-1001",
		Error:            "Invalid request format",
		ErrorDescription: "The request body is malformed or contains invalid data",
	}
	// ErrorMissingID is returned when resource server/resource/action ID is missing.
	ErrorMissingID = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "RES-1002",
		Error:            "Invalid request format",
		ErrorDescription: "ID is required",
	}
	// ErrorResourceServerNotFound is returned when a resource server is not found.
	ErrorResourceServerNotFound = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "RES-1003",
		Error:            "Resource server not found",
		ErrorDescription: "The resource server with the specified id does not exist",
	}
	// ErrorNameConflict is returned when a name already exists.
	ErrorNameConflict = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "RES-1004",
		Error:            "Name conflict",
		ErrorDescription: "A resource server with the same name already exists",
	}
	// ErrorParentResourceNotFound is returned when a parent resource is not found.
	ErrorParentResourceNotFound = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "RES-1005",
		Error:            "Parent resource not found",
		ErrorDescription: "The specified parent resource does not exist",
	}
	// ErrorCannotDelete is returned when resource server/resource cannot be deleted.
	ErrorCannotDelete = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "RES-1006",
		Error:            "Cannot delete",
		ErrorDescription: "Cannot delete resource server/resource that has dependencies",
	}
	// ErrorCircularDependency is returned when a circular dependency is detected.
	ErrorCircularDependency = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "RES-1007",
		Error:            "Circular dependency detected",
		ErrorDescription: "Setting this parent would create a circular dependency",
	}
	// ErrorResourceNotFound is returned when a resource is not found.
	ErrorResourceNotFound = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "RES-1008",
		Error:            "Resource not found",
		ErrorDescription: "The resource with the specified id does not exist",
	}
	// ErrorActionNotFound is returned when an action is not found.
	ErrorActionNotFound = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "RES-1009",
		Error:            "Action not found",
		ErrorDescription: "The action with the specified id does not exist",
	}
	// ErrorOrganizationUnitNotFound is returned when organization unit is not found.
	ErrorOrganizationUnitNotFound = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "RES-1010",
		Error:            "Organization unit not found",
		ErrorDescription: "The specified organization unit does not exist",
	}
	// ErrorInvalidLimit is returned when limit parameter is invalid.
	ErrorInvalidLimit = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "RES-1011",
		Error:            "Invalid limit parameter",
		ErrorDescription: "The limit parameter must be a positive integer",
	}
	// ErrorInvalidOffset is returned when offset parameter is invalid.
	ErrorInvalidOffset = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "RES-1012",
		Error:            "Invalid offset parameter",
		ErrorDescription: "The offset parameter must be a non-negative integer",
	}
	// ErrorIdentifierConflict is returned when an identifier already exists.
	ErrorIdentifierConflict = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "RES-1013",
		Error:            "Identifier conflict",
		ErrorDescription: "A resource server with the same identifier already exists",
	}
	// ErrorHandleConflict is returned when a handle already exists.
	ErrorHandleConflict = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "RES-1014",
		Error:            "Handle conflict",
		ErrorDescription: "The same handle already exists within the specified resource",
	}
	// ErrorInvalidDelimiter is returned when delimiter is invalid.
	ErrorInvalidDelimiter = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "RES-1015",
		Error:            "Invalid delimiter",
		ErrorDescription: "Delimiter must be a single valid character (a-z A-Z 0-9 . _ : - /)",
	}
	// ErrorInvalidHandle is returned when handle contains invalid characters.
	ErrorInvalidHandle = serviceerror.ServiceError{
		Type:  serviceerror.ClientErrorType,
		Code:  "RES-1016",
		Error: "Invalid handle",
		ErrorDescription: "Handle length must be less than 100 characters " +
			"and contain valid characters (a-z A-Z 0-9 . _ : - /)",
	}
	// ErrorDelimiterInHandle is returned when handle contains invalid characters.
	ErrorDelimiterInHandle = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "RES-1017",
		Error:            "Delimiter conflict in handle",
		ErrorDescription: "Handle cannot contain the delimiter character",
	}
)

// Internal error constants.
var (
	// errResourceServerNotFound is returned when the resource server is not found.
	errResourceServerNotFound = errors.New("resource server not found")

	// errResourceNotFound is returned when the resource is not found.
	errResourceNotFound = errors.New("resource not found")

	// errActionNotFound is returned when the action is not found.
	errActionNotFound = errors.New("action not found")
)
