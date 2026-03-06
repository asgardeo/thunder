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

package role

import (
	"errors"

	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
)

// Client errors for role management operations.
var (
	// ErrorInvalidRequestFormat is the error returned when the request format is invalid.
	ErrorInvalidRequestFormat = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "ROL-1001",
		Error:            "Invalid request format",
		ErrorDescription: "The request body is malformed or contains invalid data",
	}
	// ErrorMissingRoleID is the error returned when role ID is missing.
	ErrorMissingRoleID = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "ROL-1002",
		Error:            "Invalid request format",
		ErrorDescription: "Role ID is required",
	}
	// ErrorRoleNotFound is the error returned when a role is not found.
	ErrorRoleNotFound = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "ROL-1003",
		Error:            "Role not found",
		ErrorDescription: "The role with the specified id does not exist",
	}
	// ErrorRoleNameConflict is the error returned when a role name already exists in the organization unit.
	ErrorRoleNameConflict = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "ROL-1004",
		Error:            "Role name conflict",
		ErrorDescription: "A role with the same name exists under the same organization unit",
	}
	// ErrorOrganizationUnitNotFound is the error returned when organization unit is not found.
	ErrorOrganizationUnitNotFound = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "ROL-1005",
		Error:            "Organization unit not found",
		ErrorDescription: "Organization unit not found",
	}
	// ErrorCannotDeleteRole is the error returned when role cannot be deleted.
	ErrorCannotDeleteRole = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "ROL-1006",
		Error:            "Cannot delete role",
		ErrorDescription: "Cannot delete role that is currently assigned to users or groups",
	}
	// ErrorInvalidAssignmentID is the error returned when assignment ID is invalid.
	ErrorInvalidAssignmentID = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "ROL-1007",
		Error:            "Invalid assignment ID",
		ErrorDescription: "One or more assignment IDs in the request do not exist",
	}
	// ErrorInvalidLimit is the error returned when limit parameter is invalid.
	ErrorInvalidLimit = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "ROL-1008",
		Error:            "Invalid limit parameter",
		ErrorDescription: "The limit parameter must be a positive integer",
	}
	// ErrorInvalidOffset is the error returned when offset parameter is invalid.
	ErrorInvalidOffset = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "ROL-1009",
		Error:            "Invalid offset parameter",
		ErrorDescription: "The offset parameter must be a non-negative integer",
	}
	// ErrorEmptyAssignments is the error returned when assignments list is empty.
	ErrorEmptyAssignments = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "ROL-1010",
		Error:            "Empty assignments list",
		ErrorDescription: "At least one assignment must be provided",
	}
	// ErrorMissingUserOrGroups is the error returned when both user ID and groups are missing.
	ErrorMissingUserOrGroups = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "ROL-1011",
		Error:            "Invalid request format",
		ErrorDescription: "Either userId or groups must be provided for authorization check",
	}
	// ErrorInvalidPermissions is returned when one or more permissions are invalid.
	ErrorInvalidPermissions = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "ROL-1012",
		Error:            "Invalid permissions",
		ErrorDescription: "One or more permissions do not exist in the resource management system",
	}
	// ErrorImmutableRole is the error returned when attempting to modify a declarative role.
	ErrorImmutableRole = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "ROL-1013",
		Error:            "Cannot modify declarative role",
		ErrorDescription: "The role is defined in declarative configuration and cannot be modified",
	}
	// ErrorImmutableAssignment is the error returned when attempting to modify a declarative assignment.
	ErrorImmutableAssignment = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "ROL-1014",
		Error:            "Cannot modify declarative assignment",
		ErrorDescription: "The assignment is defined in declarative configuration and cannot be modified",
	}
	// ErrorDeclarativeModeCreateNotAllowed is the error returned when attempting to create
	// a role in declarative-only mode.
	ErrorDeclarativeModeCreateNotAllowed = serviceerror.ServiceError{
		Type:  serviceerror.ClientErrorType,
		Code:  "ROL-1015",
		Error: "Cannot create role in declarative-only mode",
		ErrorDescription: "Role creation is not allowed when running in declarative-only mode. " +
			"Roles must be defined in declarative configuration files",
	}
)

// Server errors for role management operations.
var (
	// ErrorInternalServerError is the error returned when an internal server error occurs.
	ErrorInternalServerError = serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "ROL-5000",
		Error:            "Internal server error",
		ErrorDescription: "An unexpected error occurred while processing the request",
	}
)

// Internal error constants for role management operations.
var (
	// ErrRoleNotFound is returned when the role is not found in the system.
	ErrRoleNotFound = errors.New("role not found")

	// ErrRoleHasAssignments is returned when attempting to delete a role that has active assignments.
	ErrRoleHasAssignments = errors.New("role has active assignments")

	// ErrRoleNameConflict is returned when a role with the same name already exists in the organization unit.
	ErrRoleNameConflict = errors.New("role name conflict")
)
