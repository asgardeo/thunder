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

package flowmgt

import (
	"github.com/asgardeo/thunder/internal/flow/common"
)

// FlowDefinition represents the structure of a flow definition.
type FlowDefinition struct {
	Name     string           `json:"name" validate:"required"`
	FlowType common.FlowType  `json:"flowType" validate:"required"`
	Nodes    []NodeDefinition `json:"nodes" validate:"required"`
}

// CompleteFlowDefinition represents a complete flow definition with all details.
type CompleteFlowDefinition struct {
	ID            string           `json:"id"`
	Name          string           `json:"name"`
	FlowType      common.FlowType  `json:"flowType"`
	ActiveVersion int              `json:"activeVersion"`
	Nodes         []NodeDefinition `json:"nodes"`
	CreatedAt     string           `json:"createdAt"`
	UpdatedAt     string           `json:"updatedAt"`
}

// BasicFlowDefinition represents basic information about a flow definition.
type BasicFlowDefinition struct {
	ID            string          `json:"id"`
	FlowType      common.FlowType `json:"flowType"`
	Name          string          `json:"name"`
	ActiveVersion int             `json:"activeVersion"`
	CreatedAt     string          `json:"createdAt"`
	UpdatedAt     string          `json:"updatedAt"`
}

// FlowListResponse represents a paginated list of flow definitions.
type FlowListResponse struct {
	TotalResults int                   `json:"totalResults"`
	StartIndex   int                   `json:"startIndex"`
	Count        int                   `json:"count"`
	Flows        []BasicFlowDefinition `json:"flows"`
	Links        []Link                `json:"links"`
}

// FlowVersion represents a specific version of a flow definition.
type FlowVersion struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	FlowType  string           `json:"flowType"`
	Version   int              `json:"version"`
	IsActive  bool             `json:"isActive"`
	Nodes     []NodeDefinition `json:"nodes"`
	CreatedAt string           `json:"createdAt"`
}

// FlowVersionListResponse represents a list of flow versions.
type FlowVersionListResponse struct {
	TotalVersions int                `json:"totalVersions"`
	Versions      []BasicFlowVersion `json:"versions"`
}

// BasicFlowVersion represents basic information about a flow version.
type BasicFlowVersion struct {
	Version   int    `json:"version"`
	CreatedAt string `json:"createdAt"`
	IsActive  bool   `json:"isActive"`
}

// RestoreVersionRequest represents a request to restore a specific version.
type RestoreVersionRequest struct {
	Version int `json:"version" validate:"required"`
}

// Link represents a hypermedia link for pagination.
type Link struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
}

// NodeDefinition represents a single node in a flow definition.
type NodeDefinition struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Meta       interface{}            `json:"meta,omitempty"`
	Inputs     []InputDefinition      `json:"inputs,omitempty"`
	Actions    []ActionDefinition     `json:"actions,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Executor   *ExecutorDefinition    `json:"executor,omitempty"`
	OnSuccess  string                 `json:"onSuccess,omitempty"`
	OnFailure  string                 `json:"onFailure,omitempty"`
	Condition  *ConditionDefinition   `json:"condition,omitempty"`
}

// InputDefinition represents an input parameter for a node.
type InputDefinition struct {
	Ref        string `json:"ref,omitempty"`
	Type       string `json:"type"`
	Identifier string `json:"identifier"`
	Required   bool   `json:"required"`
}

// ActionDefinition represents an action to be executed by a node.
type ActionDefinition struct {
	Ref      string `json:"ref"`
	NextNode string `json:"nextNode"`
}

// ExecutorDefinition represents the executor configuration for a node.
type ExecutorDefinition struct {
	Name string `json:"name"`
}

// ConditionDefinition represents a condition for node execution.
type ConditionDefinition struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	OnSkip string `json:"onSkip"`
}
