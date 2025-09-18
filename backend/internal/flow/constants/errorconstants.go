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

package constants

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

// ErrorInvalidFlowID defines the error response for invalid flow ID errors.
var ErrorInvalidFlowID = serviceerror.ServiceError{
	Code:             "FES-1004",
	Type:             serviceerror.ClientErrorType,
	Error:            "Invalid request",
	ErrorDescription: "Invalid flow ID provided in the request",
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

// Server error structs

// ErrorFlowGraphNotInitialized defines the error response for uninitialized flow graph errors.
var ErrorFlowGraphNotInitialized = serviceerror.ServiceError{
	Code:             "FES-5001",
	Type:             serviceerror.ServerErrorType,
	Error:            "Something went wrong",
	ErrorDescription: "Flow graph is not initialized or is nil",
}

// ErrorFlowGraphNotFound defines the error response for flow graph not found errors.
var ErrorFlowGraphNotFound = serviceerror.ServiceError{
	Code:             "FES-5002",
	Type:             serviceerror.ServerErrorType,
	Error:            "Something went wrong",
	ErrorDescription: "Flow graph not found for the graph ID",
}

// ErrorStartNodeNotFoundInGraph defines the error response for start node not found in the flow graph.
var ErrorStartNodeNotFoundInGraph = serviceerror.ServiceError{
	Code:             "FES-5003",
	Type:             serviceerror.ServerErrorType,
	Error:            "Something went wrong",
	ErrorDescription: "Start node not found in the flow graph",
}

// ErrorNodeResponseStatusNotFound defines the error response for node response status not found in the flow graph.
var ErrorNodeResponseStatusNotFound = serviceerror.ServiceError{
	Code:             "FES-5004",
	Type:             serviceerror.ServerErrorType,
	Error:            "Something went wrong",
	ErrorDescription: "Node response status not found in the flow graph",
}

// ErrorMovingToNextNode defines the error response for errors while moving to the next node in the flow graph.
var ErrorMovingToNextNode = serviceerror.ServiceError{
	Code:             "FES-5005",
	Type:             serviceerror.ServerErrorType,
	Error:            "Something went wrong",
	ErrorDescription: "Error while moving to the next node",
}

// ErrorResolvingStepForRedirection defines the error response for errors while resolving step for redirection.
var ErrorResolvingStepForRedirection = serviceerror.ServiceError{
	Code:             "FES-5006",
	Type:             serviceerror.ServerErrorType,
	Error:            "Something went wrong",
	ErrorDescription: "Error while resolving step for redirection",
}

// ErrorResolvingStepForPrompt defines the error response for errors while resolving step for prompt.
var ErrorResolvingStepForPrompt = serviceerror.ServiceError{
	Code:             "FES-5007",
	Type:             serviceerror.ServerErrorType,
	Error:            "Something went wrong",
	ErrorDescription: "Error while resolving step for prompt",
}

// ErrorUnsupportedNodeResponseType defines the error response for unsupported response type from the node.
var ErrorUnsupportedNodeResponseType = serviceerror.ServiceError{
	Code:             "FES-5008",
	Type:             serviceerror.ServerErrorType,
	Error:            "Something went wrong",
	ErrorDescription: "Unsupported response type returned from the node",
}

// ErrorUnsupportedNodeResponseStatus defines the error response for unsupported response status from the node.
var ErrorUnsupportedNodeResponseStatus = serviceerror.ServiceError{
	Code:             "FES-5009",
	Type:             serviceerror.ServerErrorType,
	Error:            "Something went wrong",
	ErrorDescription: "Unsupported response status returned from the node",
}

// ErrorNodeExecutorNotFound defines the error response for node executor not found errors.
var ErrorNodeExecutorNotFound = serviceerror.ServiceError{
	Code:             "FES-5010",
	Type:             serviceerror.ServerErrorType,
	Error:            "Something went wrong",
	ErrorDescription: "An executor not found for the node",
}

// ErrorConstructingNodeExecutor defines the error response for errors while constructing the node executor.
var ErrorConstructingNodeExecutor = serviceerror.ServiceError{
	Code:             "FES-5011",
	Type:             serviceerror.ServerErrorType,
	Error:            "Something went wrong",
	ErrorDescription: "Error constructing the node executor",
}

// ErrorNodeExecutorExecError defines the error response for errors while executing the node executor.
var ErrorNodeExecutorExecError = serviceerror.ServiceError{
	Code:             "FES-5012",
	Type:             serviceerror.ServerErrorType,
	Error:            "Executor Execution Error",
	ErrorDescription: "Error executing the node executor",
}

// ErrorNilResponseFromExecutor defines the error response for nil response from the executor.
var ErrorNilResponseFromExecutor = serviceerror.ServiceError{
	Code:             "FES-5013",
	Type:             serviceerror.ServerErrorType,
	Error:            "Executor Response Error",
	ErrorDescription: "Received nil response from the executor",
}

// ErrorUpdatingContextInStore defines the error response for errors while updating the flow context in the store.
var ErrorUpdatingContextInStore = serviceerror.ServiceError{
	Code:             "FES-5014",
	Type:             serviceerror.ServerErrorType,
	Error:            "Something went wrong",
	ErrorDescription: "Error updating flow context in the store",
}

// ErrorAuthFlowNotConfiguredForApplication defines the error response for applications without
// an authentication flow graph configured.
var ErrorAuthFlowNotConfiguredForApplication = serviceerror.ServiceError{
	Code:             "FES-5015",
	Type:             serviceerror.ServerErrorType,
	Error:            "Invalid configuration",
	ErrorDescription: "No authentication flow graph is configured for the application",
}

// ErrorInvalidAuthFlowConfiguredForApplication defines the error response for applications with an invalid
// authentication flow graph configured.
var ErrorInvalidAuthFlowConfiguredForApplication = serviceerror.ServiceError{
	Code:             "FES-5016",
	Type:             serviceerror.ServerErrorType,
	Error:            "Invalid configuration",
	ErrorDescription: "The configured flow graph is not valid for the application authentication flow",
}

// ErrorNoActionsDefinedForNode defines the error response for nodes without any actions defined.
var ErrorNoActionsDefinedForNode = serviceerror.ServiceError{
	Code:             "FES-5017",
	Type:             serviceerror.ServerErrorType,
	Error:            "Invalid configuration",
	ErrorDescription: "No actions defined for the node",
}

// ErrorRegisFlowNotConfiguredForApplication defines the error response for applications without
// a registration flow graph configured.
var ErrorRegisFlowNotConfiguredForApplication = serviceerror.ServiceError{
	Code:             "FES-5018",
	Type:             serviceerror.ServerErrorType,
	Error:            "Invalid configuration",
	ErrorDescription: "No registration flow graph is configured for the application",
}

// ErrorFlowContextConversionFailed defines the error response for failed flow context conversion.
var ErrorFlowContextConversionFailed = serviceerror.ServiceError{
	Code:             "FES-5019",
	Type:             serviceerror.ServerErrorType,
	Error:            "Something went wrong",
	ErrorDescription: "Failed to convert flow context from database format",
}

// ErrorApplicationRetrievalServerError defines the error response for application retrieval server errors.
var ErrorApplicationRetrievalServerError = serviceerror.ServiceError{
	Code:             "FES-5020",
	Type:             serviceerror.ServerErrorType,
	Error:            "Application retrieval error",
	ErrorDescription: "Server error while retrieving application details",
}
