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

package flowmeta

import (
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
)

// Error constants for flow metadata service

// ErrorInvalidType defines the error response for invalid type parameter.
var ErrorInvalidType = serviceerror.ServiceError{
	Code:             "FM-1001",
	Type:             serviceerror.ClientErrorType,
	Error:            "Invalid request",
	ErrorDescription: "The 'type' parameter must be either 'APP' or 'OU'",
}

// ErrorApplicationNotFound defines the error response for application not found.
var ErrorApplicationNotFound = serviceerror.ServiceError{
	Code:             "FM-1002",
	Type:             serviceerror.ClientErrorType,
	Error:            "Resource not found",
	ErrorDescription: "The specified application does not exist",
}

// ErrorOUNotFound defines the error response for organization unit not found.
var ErrorOUNotFound = serviceerror.ServiceError{
	Code:             "FM-1003",
	Type:             serviceerror.ClientErrorType,
	Error:            "Resource not found",
	ErrorDescription: "The specified organization unit does not exist",
}

// ErrorMissingType defines the error response for missing type parameter.
var ErrorMissingType = serviceerror.ServiceError{
	Code:             "FM-1004",
	Type:             serviceerror.ClientErrorType,
	Error:            "Missing required parameter",
	ErrorDescription: "The 'type' query parameter is required",
}

// ErrorMissingID defines the error response for missing id parameter.
var ErrorMissingID = serviceerror.ServiceError{
	Code:             "FM-1005",
	Type:             serviceerror.ClientErrorType,
	Error:            "Missing required parameter",
	ErrorDescription: "The 'id' query parameter is required",
}

// ErrorApplicationFetchFailed defines the error response for application fetch failure.
var ErrorApplicationFetchFailed = serviceerror.ServiceError{
	Code:             "FM-5001",
	Type:             serviceerror.ServerErrorType,
	Error:            "Internal server error",
	ErrorDescription: "Failed to retrieve application information",
}

// ErrorOUFetchFailed defines the error response for organization unit fetch failure.
var ErrorOUFetchFailed = serviceerror.ServiceError{
	Code:             "FM-5002",
	Type:             serviceerror.ServerErrorType,
	Error:            "Internal server error",
	ErrorDescription: "Failed to retrieve organization unit information",
}
