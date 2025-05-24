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

// Package flow provides the FlowService interface and its implementation.
package flow

import (
	"sync"

	"github.com/asgardeo/thunder/internal/flow/composer"
	"github.com/asgardeo/thunder/internal/flow/engine"
	"github.com/asgardeo/thunder/internal/flow/model"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/services/apierror"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"

	"github.com/google/uuid"
)

var (
	instance *FlowService
	once     sync.Once
)

// FlowServiceInterface defines the interface for flow orchestration and acts as the entry point for flow execution
type FlowServiceInterface interface {
	Execute(appID, callBackURL, flowID, actionID string,
		inputData map[string]string) (*model.FlowStep, *model.FlowServiceError)
}

// FlowService is the implementation of FlowServiceInterface
type FlowService struct {
	store map[string]model.FlowContext
	mu    sync.Mutex
}

// GetFlowService returns a singleton instance of FlowService
func GetFlowService() FlowServiceInterface {
	once.Do(func() {
		instance = &FlowService{
			store: make(map[string]model.FlowContext),
		}
	})
	return instance
}

// Execute executes a flow with the given data
func (s *FlowService) Execute(appID, callBackURL, flowID, actionID string,
	inputData map[string]string) (*model.FlowStep, *model.FlowServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "FlowService"))
	context := model.FlowContext{}

	// Check whether it is a initial flow execution request
	if flowID == "" && actionID == "" && len(inputData) == 0 {
		// Validate for the required parameters
		if appID == "" {
			return nil, &model.FlowServiceError{
				Type:             apierror.ClientErrorType,
				Error:            "Invalid Request",
				ErrorDescription: "appID is required",
			}
		}
		if callBackURL == "" {
			return nil, &model.FlowServiceError{
				Type:             apierror.ClientErrorType,
				Error:            "Invalid Request",
				ErrorDescription: "callBackURL is required",
			}
		}

		// Generate a new flow ID and initialize a flow context
		// TODO: Replace with the new UUID generator.
		flowID = uuid.New().String()
		context.FlowID = flowID

		// Load the graph from the composer
		composer := composer.GetFlowComposer()
		graph, ok := composer.GetGraph("auth_flow_config")
		if !ok {
			logger.Error("Graph not found")
			return nil, &model.FlowServiceError{
				Type:             "server",
				Error:            "Graph Not Found",
				ErrorDescription: "Graph not found for the graph ID",
			}
		}
		context.Graph = graph

		context.AppID = appID
		context.CallBackURL = callBackURL
	} else {
		// Validate for the required parameters
		if flowID == "" {
			return nil, &model.FlowServiceError{
				Type:             apierror.ClientErrorType,
				Error:            "Invalid Request",
				ErrorDescription: "flowID is required",
			}
		}
		if actionID == "" {
			return nil, &model.FlowServiceError{
				Type:             apierror.ClientErrorType,
				Error:            "Invalid Request",
				ErrorDescription: "actionID is required",
			}
		}
		if len(inputData) == 0 {
			return nil, &model.FlowServiceError{
				Type:             apierror.ClientErrorType,
				Error:            "Invalid Request",
				ErrorDescription: "One or more input data is required",
			}
		}

		// Load the flow context from the store
		s.mu.Lock()
		defer s.mu.Unlock()

		context, ok := s.store[flowID]
		if !ok {
			logger.Error("Flow context not found in the store")
			return nil, &model.FlowServiceError{
				Type:             apierror.ClientErrorType,
				Error:            "Invalid Request",
				ErrorDescription: "Flow context not found for the flow ID",
			}
		}

		// Remove the flow context from the store
		delete(s.store, flowID)

		// Append user inputs to the context
		sysutils.MergeStringMaps(context.UserInputData, inputData)

		context.CurrentActionID = actionID
	}

	// Execute the flow by invoking the engine
	engine := engine.GetFlowEngine()
	flowStep, err := engine.Execute(&context)
	if err != nil {
		logger.Error("Failed to execute flow", log.Error(err))

		// Remove the flow context from the store
		s.mu.Lock()
		defer s.mu.Unlock()
		delete(s.store, context.FlowID)

		return nil, &model.FlowServiceError{
			Type:             "server",
			Error:            "Flow Execution Error",
			ErrorDescription: "An error occurred while executing the flow: " + err.Error(),
		}
	}

	// Check if the flow execution is complete
	if flowStep.Status != "" && flowStep.Status == FlowStatusComplete {
		// Flow execution is complete, remove the flow context from the store.
		s.mu.Lock()
		defer s.mu.Unlock()

		delete(s.store, context.FlowID)
	} else {
		// Flow execution is incomplete, add the flow context to the store.
		s.mu.Lock()
		defer s.mu.Unlock()

		s.store[context.FlowID] = context
	}

	return &flowStep, nil
}
