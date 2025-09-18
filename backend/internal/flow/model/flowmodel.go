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

// Package model defines the data structures and interfaces for flow execution and graph representation.
package model

import (
	appmodel "github.com/asgardeo/thunder/internal/application/model"
	authndto "github.com/asgardeo/thunder/internal/authn/dto"
	"github.com/asgardeo/thunder/internal/flow/constants"
)

// EngineContext holds the overall context used by the flow engine during execution.
type EngineContext struct {
	FlowID        string
	FlowType      constants.FlowType
	AppID         string
	UserInputData map[string]string
	RuntimeData   map[string]string

	CurrentNode         NodeInterface
	CurrentNodeResponse *NodeResponse
	CurrentActionID     string

	Graph       GraphInterface
	Application appmodel.ApplicationProcessedDTO

	AuthenticatedUser authndto.AuthenticatedUser
}

// NodeContext holds the context for a specific node in the flow execution.
type NodeContext struct {
	FlowID          string
	FlowType        constants.FlowType
	AppID           string
	CurrentActionID string

	NodeInputData []InputData
	UserInputData map[string]string
	RuntimeData   map[string]string

	Application       appmodel.ApplicationProcessedDTO
	AuthenticatedUser authndto.AuthenticatedUser
}

// FlowStep represents the outcome of a individual flow step
type FlowStep struct {
	FlowID        string
	StepID        string
	Type          constants.FlowStepType
	Status        constants.FlowStatus
	Data          FlowData
	Assertion     string
	FailureReason string
}

// FlowData holds the data returned by a flow execution step
type FlowData struct {
	Inputs         []InputData       `json:"inputs,omitempty"`
	RedirectURL    string            `json:"redirectURL,omitempty"`
	Actions        []Action          `json:"actions,omitempty"`
	AdditionalData map[string]string `json:"additionalData,omitempty"`
}

// InputData represents the input data required for a flow step
type InputData struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

// Action represents an action to be executed in a flow step
type Action struct {
	Type constants.ActionType `json:"type"`
	ID   string               `json:"id"`
	// Executor *ExecutorModel `json:"executor,omitempty"`
}

// ExecutorModel represents an executor configuration within an action
type ExecutorModel struct {
	Name string `json:"name"`
}

// FlowRequest represents the flow execution API request body
type FlowRequest struct {
	ApplicationID string            `json:"applicationId"`
	FlowType      string            `json:"flowType"`
	FlowID        string            `json:"flowId"`
	ActionID      string            `json:"actionId"`
	Inputs        map[string]string `json:"inputs"`
}

// FlowResponse represents the flow execution API response body
type FlowResponse struct {
	FlowID        string   `json:"flowId"`
	StepID        string   `json:"stepId,omitempty"`
	FlowStatus    string   `json:"flowStatus"`
	Type          string   `json:"type,omitempty"`
	Data          FlowData `json:"data,omitempty"`
	Assertion     string   `json:"assertion,omitempty"`
	FailureReason string   `json:"failureReason,omitempty"`
}
