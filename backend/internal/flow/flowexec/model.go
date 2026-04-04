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
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	authncm "github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/authnprovider"
	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/system/crypto/encrypt"
)

// EngineContext holds the overall context used by the flow engine during execution.
type EngineContext struct {
	Context context.Context

	FlowID         string
	FlowType       common.FlowType
	AppID          string
	Verbose        bool
	UserInputs     map[string]string
	RuntimeData    map[string]string
	ForwardedData  map[string]interface{}
	AdditionalData map[string]string
	TraceID        string

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

// FlowContextDB represents the database row for a flow context.
type FlowContextDB struct {
	FlowID     string
	Context    string
	ExpiryTime time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// flowContextContent holds all flow state serialized into the CONTEXT JSON column.
type flowContextContent struct {
	AppID               string  `json:"appId"`
	Verbose             bool    `json:"verbose"`
	CurrentNodeID       *string `json:"currentNodeId,omitempty"`
	CurrentAction       *string `json:"currentAction,omitempty"`
	GraphID             string  `json:"graphId"`
	RuntimeData         *string `json:"runtimeData,omitempty"`
	ExecutionHistory    *string `json:"executionHistory,omitempty"`
	IsAuthenticated     bool    `json:"isAuthenticated"`
	UserID              *string `json:"userId,omitempty"`
	OUID                *string `json:"ouId,omitempty"`
	UserType            *string `json:"userType,omitempty"`
	UserInputs          *string `json:"userInputs,omitempty"`
	UserAttributes      *string `json:"userAttributes,omitempty"`
	Token               *string `json:"token,omitempty"`
	AvailableAttributes *string `json:"availableAttributes,omitempty"`
}

// GetGraphID extracts the graph ID from the context JSON.
func (f *FlowContextDB) GetGraphID() (string, error) {
	var content flowContextContent
	if err := json.Unmarshal([]byte(f.Context), &content); err != nil {
		return "", err
	}
	return content.GraphID, nil
}

// ToEngineContext converts the database model to the flow engine context.
func (f *FlowContextDB) ToEngineContext(graph core.GraphInterface) (EngineContext, error) {
	var content flowContextContent
	if err := json.Unmarshal([]byte(f.Context), &content); err != nil {
		return EngineContext{}, err
	}
	// Parse user inputs
	var userInputs map[string]string
	if content.UserInputs != nil {
		if err := json.Unmarshal([]byte(*content.UserInputs), &userInputs); err != nil {
			return EngineContext{}, err
		}
	} else {
		userInputs = make(map[string]string)
	}

	// Parse runtime data
	var runtimeData map[string]string
	if content.RuntimeData != nil {
		if err := json.Unmarshal([]byte(*content.RuntimeData), &runtimeData); err != nil {
			return EngineContext{}, err
		}
	} else {
		runtimeData = make(map[string]string)
	}

	// Parse authenticated user attributes
	var userAttributes map[string]interface{}
	if content.UserAttributes != nil {
		if err := json.Unmarshal([]byte(*content.UserAttributes), &userAttributes); err != nil {
			return EngineContext{}, err
		}
	} else {
		userAttributes = make(map[string]interface{})
	}

	// Decrypt token if present
	var token string
	if content.Token != nil && *content.Token != "" {
		encryptionService := encrypt.GetEncryptionService()
		decrypted, err := encryptionService.DecryptString(*content.Token)
		if err != nil {
			return EngineContext{}, err
		}
		token = decrypted
	}

	// Parse available attributes
	var availableAttributes *authnprovider.AvailableAttributes
	if content.AvailableAttributes != nil && strings.TrimSpace(*content.AvailableAttributes) != "" {
		var attrs authnprovider.AvailableAttributes
		if err := json.Unmarshal([]byte(*content.AvailableAttributes), &attrs); err != nil {
			return EngineContext{}, err
		}
		availableAttributes = &attrs
	}

	// Build authenticated user
	authenticatedUser := authncm.AuthenticatedUser{
		IsAuthenticated:     content.IsAuthenticated,
		UserID:              "",
		Attributes:          userAttributes,
		Token:               token,
		AvailableAttributes: availableAttributes,
	}
	if content.UserID != nil {
		authenticatedUser.UserID = *content.UserID
	}
	if content.OUID != nil {
		authenticatedUser.OUID = *content.OUID
	}
	if content.UserType != nil {
		authenticatedUser.UserType = *content.UserType
	}

	// Parse execution history
	var executionHistory map[string]*common.NodeExecutionRecord
	if content.ExecutionHistory != nil {
		if err := json.Unmarshal([]byte(*content.ExecutionHistory), &executionHistory); err != nil {
			return EngineContext{}, err
		}
	} else {
		executionHistory = make(map[string]*common.NodeExecutionRecord)
	}

	// Get current node from graph if available
	var currentNode core.NodeInterface
	if content.CurrentNodeID != nil {
		if node, exists := graph.GetNode(*content.CurrentNodeID); exists {
			currentNode = node
		}
	}

	// Get current action
	currentAction := ""
	if content.CurrentAction != nil {
		currentAction = *content.CurrentAction
	}

	return EngineContext{
		FlowID:            f.FlowID,
		TraceID:           "", // TraceID is transient and set from request context
		FlowType:          graph.GetType(),
		AppID:             content.AppID,
		Verbose:           content.Verbose,
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
func FromEngineContext(ctx EngineContext) (*FlowContextDB, error) {
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
	var oUID *string
	if ctx.AuthenticatedUser.OUID != "" {
		oUID = &ctx.AuthenticatedUser.OUID
	}

	// Get user type
	var userType *string
	if ctx.AuthenticatedUser.UserType != "" {
		userType = &ctx.AuthenticatedUser.UserType
	}

	// Encrypt and store token if present
	var encryptedToken *string
	if ctx.AuthenticatedUser.Token != "" {
		encryptionService := encrypt.GetEncryptionService()
		encrypted, err := encryptionService.EncryptString(ctx.AuthenticatedUser.Token)
		if err != nil {
			return nil, err
		}
		encryptedToken = &encrypted
	}

	// Serialize available attributes
	var availableAttributes *string
	if ctx.AuthenticatedUser.AvailableAttributes != nil {
		availableAttrsJSON, err := json.Marshal(ctx.AuthenticatedUser.AvailableAttributes)
		if err != nil {
			return nil, err
		}
		availableAttrsStr := string(availableAttrsJSON)
		availableAttributes = &availableAttrsStr
	}

	// Get graph ID
	if ctx.Graph == nil || ctx.Graph.GetID() == "" {
		return nil, fmt.Errorf("graph with a valid ID is required to persist engine context")
	}
	graphID := ctx.Graph.GetID()

	content := flowContextContent{
		AppID:               ctx.AppID,
		Verbose:             ctx.Verbose,
		CurrentNodeID:       currentNodeID,
		CurrentAction:       currentAction,
		GraphID:             graphID,
		RuntimeData:         &runtimeData,
		ExecutionHistory:    &executionHistory,
		IsAuthenticated:     ctx.AuthenticatedUser.IsAuthenticated,
		UserID:              authenticatedUserID,
		OUID:                oUID,
		UserType:            userType,
		UserInputs:          &userInputs,
		UserAttributes:      &userAttributes,
		Token:               encryptedToken,
		AvailableAttributes: availableAttributes,
	}

	contextJSON, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}

	return &FlowContextDB{
		FlowID:  ctx.FlowID,
		Context: string(contextJSON),
	}, nil
}
