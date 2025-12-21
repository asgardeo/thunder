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
	"encoding/json"
	"time"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	authncm "github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
)

// EngineContext holds the overall context used by the flow engine during execution.
type EngineContext struct {
	FlowID      string
	FlowType    common.FlowType
	AppID       string
	Verbose     bool
	UserInputs  map[string]string
	RuntimeData map[string]string
	TraceID     string

	CurrentNode         core.NodeInterface
	CurrentNodeResponse *common.NodeResponse
	CurrentAction       string

	Graph       core.GraphInterface
	Application appmodel.Application

	AuthenticatedUser authncm.AuthenticatedUser
	Assertion         string
	ExecutionHistory  map[string]*common.NodeExecutionRecord
}

// FlowStep represents the outcome of a individual flow step
type FlowStep struct {
	FlowID        string
	StepID        string
	Type          common.FlowStepType
	Status        common.FlowStatus
	Data          FlowData
	Assertion     string
	FailureReason string
}

// FlowData holds the data returned by a flow execution step
type FlowData struct {
	Inputs         []common.Input    `json:"inputs,omitempty"`
	RedirectURL    string            `json:"redirectURL,omitempty"`
	Actions        []common.Action   `json:"actions,omitempty"`
	Meta           interface{}       `json:"meta,omitempty"`
	AdditionalData map[string]string `json:"additionalData,omitempty"`
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

// FlowRequest represents the flow execution API request body
type FlowRequest struct {
	ApplicationID string            `json:"applicationId"`
	FlowType      string            `json:"flowType"`
	Verbose       bool              `json:"verbose,omitempty"`
	FlowID        string            `json:"flowId"`
	Action        string            `json:"action"`
	Inputs        map[string]string `json:"inputs"`
}

// FlowInitContext represents the context for initiating a new flow with runtime data
type FlowInitContext struct {
	ApplicationID string
	FlowType      string
	RuntimeData   map[string]string
}

// FlowContextWithUserDataDB represents the combined flow context and user data.
type FlowContextWithUserDataDB struct {
	FlowID             string
	AppID              string
	Verbose            bool
	CurrentNodeID      *string
	CurrentAction      *string
	GraphID            string
	RuntimeData        *string
	IsAuthenticated    bool
	UserID             *string
	OrganizationUnitID *string
	UserType           *string
	UserInputs         *string
	UserAttributes     *string
	ExecutionHistory   *string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// ToEngineContext converts the database model to the flow engine context.
func (f *FlowContextWithUserDataDB) ToEngineContext(graph core.GraphInterface) (EngineContext, error) {
	// Parse user inputs
	var userInputs map[string]string
	if f.UserInputs != nil {
		if err := json.Unmarshal([]byte(*f.UserInputs), &userInputs); err != nil {
			return EngineContext{}, err
		}
	} else {
		userInputs = make(map[string]string)
	}

	// Parse runtime data
	var runtimeData map[string]string
	if f.RuntimeData != nil {
		if err := json.Unmarshal([]byte(*f.RuntimeData), &runtimeData); err != nil {
			return EngineContext{}, err
		}
	} else {
		runtimeData = make(map[string]string)
	}

	// Parse authenticated user attributes
	var userAttributes map[string]interface{}
	if f.UserAttributes != nil {
		if err := json.Unmarshal([]byte(*f.UserAttributes), &userAttributes); err != nil {
			return EngineContext{}, err
		}
	} else {
		userAttributes = make(map[string]interface{})
	}

	// Build authenticated user
	authenticatedUser := authncm.AuthenticatedUser{
		IsAuthenticated: f.IsAuthenticated,
		UserID:          "",
		Attributes:      userAttributes,
	}
	if f.UserID != nil {
		authenticatedUser.UserID = *f.UserID
	}
	if f.OrganizationUnitID != nil {
		authenticatedUser.OrganizationUnitID = *f.OrganizationUnitID
	}
	if f.UserType != nil {
		authenticatedUser.UserType = *f.UserType
	}

	// Parse execution history
	var executionHistory map[string]*common.NodeExecutionRecord
	if f.ExecutionHistory != nil {
		if err := json.Unmarshal([]byte(*f.ExecutionHistory), &executionHistory); err != nil {
			return EngineContext{}, err
		}
	} else {
		executionHistory = make(map[string]*common.NodeExecutionRecord)
	}

	// Get current node from graph if available
	var currentNode core.NodeInterface
	if f.CurrentNodeID != nil && graph != nil {
		if node, exists := graph.GetNode(*f.CurrentNodeID); exists {
			currentNode = node
		}
	}

	// Get current action
	currentAction := ""
	if f.CurrentAction != nil {
		currentAction = *f.CurrentAction
	}

	return EngineContext{
		FlowID:            f.FlowID,
		TraceID:           "", // TraceID is transient and set from request context
		FlowType:          graph.GetType(),
		AppID:             f.AppID,
		Verbose:           f.Verbose,
		UserInputs:        userInputs,
		RuntimeData:       runtimeData,
		CurrentNode:       currentNode,
		CurrentAction:     currentAction,
		Graph:             graph,
		AuthenticatedUser: authenticatedUser,
		ExecutionHistory:  executionHistory,
	}, nil
}

// FromEngineContext creates a database model from the flow engine context.
func FromEngineContext(ctx EngineContext) (*FlowContextWithUserDataDB, error) {
	// Serialize user inputs
	userInputsJSON, err := json.Marshal(ctx.UserInputs)
	if err != nil {
		return nil, err
	}
	userInputs := string(userInputsJSON)

	// Serialize runtime data
	runtimeDataJSON, err := json.Marshal(ctx.RuntimeData)
	if err != nil {
		return nil, err
	}
	runtimeData := string(runtimeDataJSON)

	// Serialize authenticated user attributes
	userAttributesJSON, err := json.Marshal(ctx.AuthenticatedUser.Attributes)
	if err != nil {
		return nil, err
	}
	userAttributes := string(userAttributesJSON)

	// Serialize execution history
	executionHistoryJSON, err := json.Marshal(ctx.ExecutionHistory)
	if err != nil {
		return nil, err
	}
	executionHistory := string(executionHistoryJSON)

	// Get current node ID
	var currentNodeID *string
	if ctx.CurrentNode != nil {
		nodeID := ctx.CurrentNode.GetID()
		currentNodeID = &nodeID
	}

	// Get current action
	var currentAction *string
	if ctx.CurrentAction != "" {
		currentAction = &ctx.CurrentAction
	}

	// Get authenticated user ID
	var authenticatedUserID *string
	if ctx.AuthenticatedUser.UserID != "" {
		authenticatedUserID = &ctx.AuthenticatedUser.UserID
	}

	// Get organization unit ID
	var organizationUnitID *string
	if ctx.AuthenticatedUser.OrganizationUnitID != "" {
		organizationUnitID = &ctx.AuthenticatedUser.OrganizationUnitID
	}

	// Get user type
	var userType *string
	if ctx.AuthenticatedUser.UserType != "" {
		userType = &ctx.AuthenticatedUser.UserType
	}

	// Get graph ID
	graphID := ""
	if ctx.Graph != nil {
		graphID = ctx.Graph.GetID()
	}

	return &FlowContextWithUserDataDB{
		FlowID:             ctx.FlowID,
		AppID:              ctx.AppID,
		Verbose:            ctx.Verbose,
		CurrentNodeID:      currentNodeID,
		CurrentAction:      currentAction,
		GraphID:            graphID,
		RuntimeData:        &runtimeData,
		IsAuthenticated:    ctx.AuthenticatedUser.IsAuthenticated,
		UserID:             authenticatedUserID,
		OrganizationUnitID: organizationUnitID,
		UserType:           userType,
		UserInputs:         &userInputs,
		UserAttributes:     &userAttributes,
		ExecutionHistory:   &executionHistory,
	}, nil
}
