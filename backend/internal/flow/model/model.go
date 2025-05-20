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

// Package model defines the data structures and models used in the flow execution
package model

// FlowContext holds the context for flow execution
type FlowContext struct {
}

// NewFlowContext creates a new flow context
func NewFlowContext(flowID string) *FlowContext {
	return &FlowContext{}
}

// FlowStep represents the result of a flow execution
type FlowStep struct {
	ID       string
	Type     string
	StepData StepData
}

// StepData holds the data for a step in the flow
type StepData struct {
	Components     []Component
	RequiredParams []string
	AdditionalData map[string]string
}

// Component represents a component in a step
// type Component struct {
// 	ID         string
// 	Type       string
// 	Components []Component
// }

// ExecutorResponse represents the response from an executor
type ExecutorResponse struct{}
