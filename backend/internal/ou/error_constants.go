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

package ou

import (
	"errors"

	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
)

// Client errors for organization unit management operations.
var (
	// ErrorInvalidRequestFormat is the error returned when the request format is invalid.
	ErrorInvalidRequestFormat = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "OU-1001",
		Error:            "Invalid request format",
		ErrorDescription: "The request body is malformed, contains invalid data, or required fields are missing/empty",
	}
	// ErrorMissingOrganizationUnitID is the error returned when organization unit ID is missing.
	ErrorMissingOrganizationUnitID = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "OU-1002",
		Error:            "Invalid request format",
		ErrorDescription: "Organization unit ID is required",
	}
	// ErrorOrganizationUnitNotFound is the error returned when an organization unit is not found.
	ErrorOrganizationUnitNotFound = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "OU-1003",
		Error:            "Organization unit not found",
		ErrorDescription: "The organization unit with the specified id does not exist",
	}
	// ErrorOrganizationUnitNameConflict is the error returned when an organization unit name conflicts.
	ErrorOrganizationUnitNameConflict = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "OU-1004",
		Error:            "Organization unit name conflict",
		ErrorDescription: "An organization unit with the same name exists under the same parent",
	}
	// ErrorParentOrganizationUnitNotFound is the error returned when parent organization unit is not found.
	ErrorParentOrganizationUnitNotFound = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "OU-1005",
		Error:            "Parent organization unit not found",
		ErrorDescription: "Parent organization unit not found",
	}
	// ErrorCannotDeleteOrganizationUnit is the error returned when organization unit cannot be deleted.
	ErrorCannotDeleteOrganizationUnit = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "OU-1006",
		Error:            "Organization unit has children",
		ErrorDescription: "Cannot delete organization unit with children or users/groups",
	}
	// ErrorCircularDependency is the error returned when a circular dependency is detected.
	ErrorCircularDependency = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "OU-1007",
		Error:            "Circular dependency detected",
		ErrorDescription: "Setting this parent would create a circular dependency",
	}
	// ErrorOrganizationUnitHandleConflict is the error returned when an organization unit handle conflicts.
	ErrorOrganizationUnitHandleConflict = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "OU-1008",
		Error:            "Organization unit handle conflict",
		ErrorDescription: "An organization unit with the same handle already exists under the same parent",
	}
	// ErrorInvalidHandlePath is the error returned when handle path is invalid.
	ErrorInvalidHandlePath = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "OU-1009",
		Error:            "Invalid handle path",
		ErrorDescription: "The specified handle path does not exist",
	}
	// ErrorInvalidLimit is the error returned when limit parameter is invalid.
	ErrorInvalidLimit = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "OU-1010",
		Error:            "Invalid limit parameter",
		ErrorDescription: "The limit parameter must be a positive integer",
	}
	// ErrorInvalidOffset is the error returned when offset parameter is invalid.
	ErrorInvalidOffset = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "OU-1011",
		Error:            "Invalid offset parameter",
		ErrorDescription: "The offset parameter must be a non-negative integer",
	}
	// ErrorCannotModifyDeclarativeResource is the error returned when trying to modify a declarative resource.
	ErrorCannotModifyDeclarativeResource = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "OU-1012",
		Error:            "Cannot modify declarative resource",
		ErrorDescription: "The organization unit is declarative and cannot be modified or deleted",
	}
)

// Server errors for organization unit management operations.
var (
	// ErrorInternalServerError is the error returned when an internal server error occurs.
	ErrorInternalServerError = serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "OU-5000",
		Error:            "Internal server error",
		ErrorDescription: "An unexpected error occurred while processing the request",
	}
)

// Error variables
var (
	// ErrOrganizationUnitNotFound is returned when the organization unit is not found in the system.
	ErrOrganizationUnitNotFound = errors.New("organization unit not found")
	// ErrCannotUpdateDeclarativeOU is returned when attempting to update a declarative organization unit.
	ErrCannotUpdateDeclarativeOU = errors.New("cannot update declarative organization unit")
	// ErrCannotDeleteDeclarativeOU is returned when attempting to delete a declarative organization unit.
	ErrCannotDeleteDeclarativeOU = errors.New("cannot delete declarative organization unit")
)
