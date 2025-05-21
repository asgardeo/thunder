/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
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

// Package handler provides HTTP handlers for managing flow related API requests.
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/asgardeo/thunder/internal/flow"
	"github.com/asgardeo/thunder/internal/flow/model"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/services/apierror"
)

// FlowExecutionHandler handles flow execution requests.
type FlowExecutionHandler struct{}

// NewFlowExecutionHandler creates a new instance of FlowExecutionHandler.
func NewFlowExecutionHandler() *FlowExecutionHandler {
	return &FlowExecutionHandler{}
}

// HandleFlowExecutionRequest handles the flow execution request.
func (h *FlowExecutionHandler) HandleFlowExecutionRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "FlowExecutionHandler"))

	var flowExecutionRequest model.FlowRequest
	if err := json.NewDecoder(r.Body).Decode(&flowExecutionRequest); err != nil {
		errResponse := apierror.ErrorResponse{
			Message:     "Invalid request payload",
			Description: "Failed to decode request payload",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		if err := json.NewEncoder(w).Encode(errResponse); err != nil {
			logger.Error("Error encoding error response", log.Error(err))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}

		return
	}

	flowService := flow.GetFlowService()
	flowStep, flowErr := flowService.Execute(flowExecutionRequest.ApplicationID, flowExecutionRequest.CallbackURL,
		flowExecutionRequest.FlowID, flowExecutionRequest.ActionID, flowExecutionRequest.Inputs)

	if flowErr != nil {
		errResponse := apierror.ErrorResponse{
			Message:     flowErr.Error,
			Description: flowErr.ErrorDescription,
		}

		w.Header().Set("Content-Type", "application/json")
		if flowErr.Type == apierror.ClientErrorType {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		if err := json.NewEncoder(w).Encode(errResponse); err != nil {
			logger.Error("Error encoding error response", log.Error(err))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}

		return
	}

	flowExecutionResponse := model.FlowResponse{
		Type:       "test",
		FlowID:     flowExecutionRequest.FlowID,
		FlowStatus: flowStep.Status,
		Data: model.FlowResponseData{
			Components: flowStep.StepData.Components,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err := json.NewEncoder(w).Encode(flowExecutionResponse)
	if err != nil {
		logger.Error("Error encoding response", log.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Debug("Flow execution request handled successfully", log.String("flowID", flowExecutionRequest.FlowID))
}
