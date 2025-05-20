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
	"github.com/asgardeo/thunder/internal/flow/engine"
	"github.com/asgardeo/thunder/internal/flow/model"
)

// FlowServiceInterface defines the interface for flow orchestration and acts as the entry point for flow execution
type FlowServiceInterface interface {
	Execute(flowID string, actionID string, inputData map[string]string) (*model.FlowStep, error)
}

// FlowService is the implementation of FlowServiceInterface
type FlowService struct {
	engine engine.FlowEngineInterface
}

// NewFlowService creates a new FlowService
func NewFlowService() FlowServiceInterface {
	flowEngine := engine.NewFlowEngine()

	return &FlowService{
		engine: flowEngine,
	}
}

// Execute executes a flow with the given data
// TODO: Need to decide and modify the function parameters accordingly
func (s *FlowService) Execute(flowID string, actionID string, inputData map[string]string) (*model.FlowStep, error) {
	// Create a new flow context
	ctx := model.NewFlowContext(flowID)

	flowStep, err := s.engine.Execute(ctx)
	if err != nil {
		return nil, err
	}

	// Check state.

	return flowStep, nil
}
