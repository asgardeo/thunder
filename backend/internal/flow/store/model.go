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

// Package store provides the implementation for flow context persistence operations.
package store

import (
	"encoding/json"
	"time"

	authndto "github.com/asgardeo/thunder/internal/authn/dto"
	"github.com/asgardeo/thunder/internal/flow/model"
)

// FlowContextWithUserDataDB represents the combined flow context and user data.
type FlowContextWithUserDataDB struct {
	FlowID          string
	AppID           string
	CurrentNodeID   *string
	CurrentActionID *string
	GraphID         string
	RuntimeData     *string
	IsAuthenticated bool
	UserID          *string
	UserInputs      *string
	UserAttributes  *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// ToEngineContext converts the database model to the flow engine context.
func (f *FlowContextWithUserDataDB) ToEngineContext(graph model.GraphInterface) (model.EngineContext, error) {
	// Parse user input data
	var userInputData map[string]string
	if f.UserInputs != nil {
		if err := json.Unmarshal([]byte(*f.UserInputs), &userInputData); err != nil {
			return model.EngineContext{}, err
		}
	} else {
		userInputData = make(map[string]string)
	}

	// Parse runtime data
	var runtimeData map[string]string
	if f.RuntimeData != nil {
		if err := json.Unmarshal([]byte(*f.RuntimeData), &runtimeData); err != nil {
			return model.EngineContext{}, err
		}
	} else {
		runtimeData = make(map[string]string)
	}

	// Parse authenticated user attributes
	var userAttributes map[string]string
	if f.UserAttributes != nil {
		if err := json.Unmarshal([]byte(*f.UserAttributes), &userAttributes); err != nil {
			return model.EngineContext{}, err
		}
	} else {
		userAttributes = make(map[string]string)
	}

	// Build authenticated user
	authenticatedUser := authndto.AuthenticatedUser{
		IsAuthenticated: f.IsAuthenticated,
		UserID:          "",
		Attributes:      userAttributes,
	}
	if f.UserID != nil {
		authenticatedUser.UserID = *f.UserID
	}

	// Get current node from graph if available
	var currentNode model.NodeInterface
	if f.CurrentNodeID != nil && graph != nil {
		if node, exists := graph.GetNode(*f.CurrentNodeID); exists {
			currentNode = node
		}
	}

	// Get current action ID
	currentActionID := ""
	if f.CurrentActionID != nil {
		currentActionID = *f.CurrentActionID
	}

	return model.EngineContext{
		FlowID:            f.FlowID,
		FlowType:          graph.GetType(),
		AppID:             f.AppID,
		UserInputData:     userInputData,
		RuntimeData:       runtimeData,
		CurrentNode:       currentNode,
		CurrentActionID:   currentActionID,
		Graph:             graph,
		AuthenticatedUser: authenticatedUser,
	}, nil
}

// FromEngineContext creates a database model from the flow engine context.
func FromEngineContext(ctx model.EngineContext) (*FlowContextWithUserDataDB, error) {
	// Serialize user input data
	userInputDataJSON, err := json.Marshal(ctx.UserInputData)
	if err != nil {
		return nil, err
	}
	userInputData := string(userInputDataJSON)

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

	// Get current node ID
	var currentNodeID *string
	if ctx.CurrentNode != nil {
		nodeID := ctx.CurrentNode.GetID()
		currentNodeID = &nodeID
	}

	// Get current action ID
	var currentActionID *string
	if ctx.CurrentActionID != "" {
		currentActionID = &ctx.CurrentActionID
	}

	// Get authenticated user ID
	var authenticatedUserID *string
	if ctx.AuthenticatedUser.UserID != "" {
		authenticatedUserID = &ctx.AuthenticatedUser.UserID
	}

	// Get graph ID
	graphID := ""
	if ctx.Graph != nil {
		graphID = ctx.Graph.GetID()
	}

	return &FlowContextWithUserDataDB{
		FlowID:          ctx.FlowID,
		AppID:           ctx.AppID,
		CurrentNodeID:   currentNodeID,
		CurrentActionID: currentActionID,
		GraphID:         graphID,
		RuntimeData:     &runtimeData,
		IsAuthenticated: ctx.AuthenticatedUser.IsAuthenticated,
		UserID:          authenticatedUserID,
		UserInputs:      &userInputData,
		UserAttributes:  &userAttributes,
	}, nil
}
