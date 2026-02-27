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

package flowmgt

import (
	"errors"

	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
)

// Client errors for flow management operations.
var (
	// ErrorInvalidRequestFormat is the error returned when the request format is invalid.
	ErrorInvalidRequestFormat = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "FLM-1001",
		Error:            "Invalid request format",
		ErrorDescription: "The request body is malformed or contains invalid data",
	}
	// ErrorMissingFlowID is the error returned when flow ID is missing.
	ErrorMissingFlowID = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "FLM-1002",
		Error:            "Invalid flow ID",
		ErrorDescription: "The flow ID must be provided",
	}
	// ErrorFlowNotFound is the error returned when a flow is not found.
	ErrorFlowNotFound = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "FLM-1003",
		Error:            "Flow not found",
		ErrorDescription: "The flow with the specified id does not exist",
	}
	// ErrorInvalidFlowType is the error returned when flow type is invalid.
	ErrorInvalidFlowType = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "FLM-1004",
		Error:            "Invalid flow type",
		ErrorDescription: "The specified flow type is invalid",
	}
	// ErrorInvalidFlowData is the error returned when flow data is invalid.
	ErrorInvalidFlowData = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "FLM-1005",
		Error:            "Invalid flow data",
		ErrorDescription: "The flow definition contains invalid data",
	}
	// ErrorInvalidLimit is the error returned when limit parameter is invalid.
	ErrorInvalidLimit = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "FLM-1006",
		Error:            "Invalid pagination parameter",
		ErrorDescription: "The limit parameter must be a positive integer",
	}
	// ErrorInvalidOffset is the error returned when offset parameter is invalid.
	ErrorInvalidOffset = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "FLM-1007",
		Error:            "Invalid pagination parameter",
		ErrorDescription: "The offset parameter must be a non-negative integer",
	}
	// ErrorVersionNotFound is the error returned when a flow version is not found.
	ErrorVersionNotFound = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "FLM-1008",
		Error:            "Flow version not found",
		ErrorDescription: "The requested flow version does not exist",
	}
	// ErrorInvalidVersion is the error returned when a flow version is invalid.
	ErrorInvalidVersion = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "FLM-1009",
		Error:            "Invalid flow version",
		ErrorDescription: "The specified flow version is invalid",
	}
	// ErrorMissingFlowHandle is the error returned when flow handle is missing.
	ErrorMissingFlowHandle = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "FLM-1010",
		Error:            "Invalid flow handle",
		ErrorDescription: "The flow handle must be provided",
	}
	// ErrorMissingFlowName is the error returned when flow name is missing.
	ErrorMissingFlowName = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "FLM-1011",
		Error:            "Invalid flow name",
		ErrorDescription: "The flow name must be provided",
	}
	// ErrorCannotUpdateFlowType is the error returned when trying to update flow type.
	ErrorCannotUpdateFlowType = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "FLM-1012",
		Error:            "Invalid update request",
		ErrorDescription: "The flow type cannot be changed once created",
	}
	// ErrorDuplicateFlowHandle is the error returned when a flow with the same handle and type already exists.
	ErrorDuplicateFlowHandle = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "FLM-1013",
		Error:            "Duplicate flow handle",
		ErrorDescription: "A flow with this handle already exists for the given flow type",
	}
	// ErrorHandleUpdateNotAllowed is the error returned when attempting to update an immutable handle.
	ErrorHandleUpdateNotAllowed = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "FLM-1014",
		Error:            "Invalid update request",
		ErrorDescription: "The flow handle cannot be modified after creation",
	}
	// ErrorInvalidFlowHandleFormat is the error returned when handle format is invalid.
	ErrorInvalidFlowHandleFormat = serviceerror.ServiceError{
		Type:  serviceerror.ClientErrorType,
		Code:  "FLM-1015",
		Error: "Invalid flow handle format",
		ErrorDescription: "The flow handle must be lowercase, alphanumeric, and can only contain " +
			"underscores or dashes",
	}

	// ErrorGraphBuildFailure is the error returned when graph building fails.
	// TODO: This should be removed and instead should return InternalServerError
	// for graph build failures. Ideally there should be a graph validation step during
	// flow creation/update to catch such errors early.
	ErrorGraphBuildFailure = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "FLM-1016",
		Error:            "Graph build failure",
		ErrorDescription: "Failed to build executable graph from flow definition",
	}
	// ErrorFlowDeclarativeReadOnly is the error returned when trying to modify a declarative (immutable) flow.
	ErrorFlowDeclarativeReadOnly = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "FLM-1017",
		Error:            "Flow is immutable",
		ErrorDescription: "Declarative flows cannot be modified or deleted",
	}
)

// Internal errors
var (
	errFlowNotFound    = errors.New("flow not found")
	errVersionNotFound = errors.New("version not found")
)
