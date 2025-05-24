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
	FlowID        string
	AppID         string
	CallBackURL   string
	UserInputData map[string]string

	CurrentNode     *Node
	CurrentActionID string

	Graph *Graph
}

// FlowStep represents the result of a flow execution
type FlowStep struct {
	ID       string
	Type     string
	Status   string
	StepData StepData
}

// StepData holds the data for a step in the flow
type StepData struct {
	Components     []Component
	RequiredParams []string
	AdditionalData map[string]string
}

// Flow represents a execution flow
type Flow struct {
	ID    string `json:"id"`
	Steps []Step `json:"steps"`
}

// Step represents a step in a flow
type Step struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Size     Size     `json:"size"`
	Position Position `json:"position"`
	Data     Data     `json:"data"`
}

// Size represents the dimensions of a step model
type Size struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Position represents the coordinates of a step model
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Data represents the data of a step model
type Data struct {
	Components []Component `json:"components"`
}

// Component represents a component in a step model
type Component struct {
	ID         string                 `json:"id"`
	Category   string                 `json:"category"`
	Type       string                 `json:"type"`
	Variant    string                 `json:"variant"`
	Config     map[string]interface{} `json:"config"`
	Action     *Action                `json:"action,omitempty"`
	Components []Component            `json:"components"`
}

// Action represents an action configuration for a component
type Action struct {
	Type     string         `json:"type"`
	Next     string         `json:"next,omitempty"`
	Executor *ExecutorModel `json:"executor,omitempty"`
}

// ExecutorModel represents an executor configuration within an action
type ExecutorModel struct {
	Name string `json:"name"`
}

// FlowServiceError represents an error response from the flow service
type FlowServiceError struct {
	Type             string `json:"type,omitempty"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// FlowRequest represents the flow execution API request body
type FlowRequest struct {
	ApplicationID string            `json:"applicationId"`
	CallbackURL   string            `json:"callbackUrl"`
	FlowID        string            `json:"flowId"`
	ActionID      string            `json:"actionId"`
	Inputs        map[string]string `json:"inputs"`
}

// FlowResponse represents the flow execution API response body
type FlowResponse struct {
	Type       string           `json:"type"`
	FlowID     string           `json:"flowId"`
	FlowStatus string           `json:"flowStatus"`
	Data       FlowResponseData `json:"data"`
}

// FlowResponseData represents the data in the flow execution API response
type FlowResponseData struct {
	Components []Component `json:"components"`
}
