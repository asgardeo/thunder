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

package flowexec

import (
	"github.com/asgardeo/thunder/internal/system/error/apierror"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
)

// Client error structs

// APIErrorFlowRequestJSONDecodeError defines the error response for json decode errors.
var APIErrorFlowRequestJSONDecodeError = apierror.ErrorResponse{
	Code:        "FES-1001",
	Message:     "Invalid request payload",
	Description: "Failed to decode request payload",
}

// ErrorNodeResponse defines the error response for errors received from nodes.
var ErrorNodeResponse = serviceerror.ServiceError{
	Code:             "FES-1002",
	Type:             serviceerror.ClientErrorType,
	Error:            "Invalid node response",
	ErrorDescription: "Error response received from the node",
}

// ErrorInvalidAppID defines the error response for invalid app ID errors.
var ErrorInvalidAppID = serviceerror.ServiceError{
	Code:             "FES-1003",
	Type:             serviceerror.ClientErrorType,
	Error:            "Invalid request",
	ErrorDescription: "Invalid app ID provided in the request",
}

// ErrorInvalidExecutionID defines the error response for invalid execution ID errors.
var ErrorInvalidExecutionID = serviceerror.ServiceError{
	Code:             "FES-1004",
	Type:             serviceerror.ClientErrorType,
	Error:            "Invalid request",
	ErrorDescription: "Invalid flow execution ID provided in the request",
}

// ErrorInvalidFlowType defines the error response for invalid flow type errors.
var ErrorInvalidFlowType = serviceerror.ServiceError{
	Code:             "FES-1005",
	Type:             serviceerror.ClientErrorType,
	Error:            "Invalid request",
	ErrorDescription: "Invalid flow type provided in the request",
}

// ErrorRegistrationFlowDisabled defines the error response for registration flow disabled errors.
var ErrorRegistrationFlowDisabled = serviceerror.ServiceError{
	Code:             "FES-1006",
	Type:             serviceerror.ClientErrorType,
	Error:            "Registration not allowed",
	ErrorDescription: "Registration flow is disabled for the application",
}

// ErrorApplicationRetrievalClientError defines the error response for application retrieval client errors.
var ErrorApplicationRetrievalClientError = serviceerror.ServiceError{
	Code:             "FES-1007",
	Type:             serviceerror.ClientErrorType,
	Error:            "Application retrieval error",
	ErrorDescription: "Error while retrieving application details",
}

// ErrorInvalidFlowInitContext defines the error response for invalid flow init context.
var ErrorInvalidFlowInitContext = serviceerror.ServiceError{
	Code:             "FES-1008",
	Type:             serviceerror.ClientErrorType,
	Error:            "Invalid request",
	ErrorDescription: "Invalid flow initialization context provided",
}
