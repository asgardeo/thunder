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

// Package engine provides the flow engine for orchestrating flow executions.
package engine

import (
	"sync"

	"github.com/asgardeo/thunder/internal/flow/model"
)

var (
	instance *FlowEngine
	once     sync.Once
)

// FlowEngineInterface defines the interface for the flow engine.
type FlowEngineInterface interface {
	Execute(ctx *model.FlowContext) (model.FlowStep, error)
}

// FlowEngine is the main engine implementation for orchestrating flow executions.
type FlowEngine struct{}

// GetFlowEngine returns a singleton instance of FlowEngine.
func GetFlowEngine() FlowEngineInterface {
	once.Do(func() {
		instance = &FlowEngine{}
	})
	return instance
}

// Execute executes a step in the flow
func (e *FlowEngine) Execute(ctx *model.FlowContext) (model.FlowStep, error) {
	// TODO: Implement the execution logic for the flow engine
	return model.FlowStep{}, nil
}
