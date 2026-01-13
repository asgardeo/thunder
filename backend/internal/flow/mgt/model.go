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
	"encoding/json"

	"gopkg.in/yaml.v3"

	"github.com/asgardeo/thunder/internal/flow/common"
)

// FlowDefinition represents the structure of a flow definition.
type FlowDefinition struct {
	Handle   string           `json:"handle" validate:"required"`
	Name     string           `json:"name" validate:"required"`
	FlowType common.FlowType  `json:"flowType" validate:"required"`
	Nodes    []NodeDefinition `json:"nodes" validate:"required"`
}

// CompleteFlowDefinition represents a complete flow definition with all details.
type CompleteFlowDefinition struct {
	ID            string           `json:"id" yaml:"id"`
	Handle        string           `json:"handle" yaml:"handle"`
	Name          string           `json:"name" yaml:"name"`
	FlowType      common.FlowType  `json:"flowType" yaml:"flowType"`
	ActiveVersion int              `json:"activeVersion" yaml:"activeVersion"`
	Nodes         []NodeDefinition `json:"nodes" yaml:"nodes"`
	CreatedAt     string           `json:"createdAt" yaml:"createdAt"`
	UpdatedAt     string           `json:"updatedAt" yaml:"updatedAt"`
}

// BasicFlowDefinition represents basic information about a flow definition.
type BasicFlowDefinition struct {
	ID            string          `json:"id"`
	Handle        string          `json:"handle"`
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
	Handle    string           `json:"handle"`
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

// NodeLayout represents the layout information for a node in the flow composer UI.
type NodeLayout struct {
	Size     *NodeSize     `json:"size,omitempty" yaml:"size,omitempty"`
	Position *NodePosition `json:"position,omitempty" yaml:"position,omitempty"`
}

// NodeSize represents the dimensions of a node.
type NodeSize struct {
	Width  float64 `json:"width" yaml:"width"`
	Height float64 `json:"height" yaml:"height"`
}

// NodePosition represents the position of a node on the canvas.
type NodePosition struct {
	X float64 `json:"x" yaml:"x"`
	Y float64 `json:"y" yaml:"y"`
}

// NodeDefinition represents a single node in a flow definition.
type NodeDefinition struct {
	ID         string                 `json:"id" yaml:"id"`
	Type       string                 `json:"type" yaml:"type"`
	Layout     *NodeLayout            `json:"layout,omitempty" yaml:"layout,omitempty"`
	Meta       interface{}            `json:"meta,omitempty" yaml:"meta,omitempty"`
	Inputs     []InputDefinition      `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Actions    []ActionDefinition     `json:"actions,omitempty" yaml:"actions,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty" yaml:"properties,omitempty"`
	Executor   *ExecutorDefinition    `json:"executor,omitempty" yaml:"executor,omitempty"`
	OnSuccess  string                 `json:"onSuccess,omitempty" yaml:"onSuccess,omitempty"`
	OnFailure  string                 `json:"onFailure,omitempty" yaml:"onFailure,omitempty"`
	Condition  *ConditionDefinition   `json:"condition,omitempty" yaml:"condition,omitempty"`
}

// InputDefinition represents an input parameter for a node.
type InputDefinition struct {
	Ref        string `json:"ref,omitempty" yaml:"ref,omitempty"`
	Type       string `json:"type" yaml:"type"`
	Identifier string `json:"identifier" yaml:"identifier"`
	Required   bool   `json:"required" yaml:"required"`
}

// ActionDefinition represents an action to be executed by a node.
type ActionDefinition struct {
	Ref      string `json:"ref" yaml:"ref"`
	NextNode string `json:"nextNode" yaml:"nextNode"`
}

// ExecutorDefinition represents the executor configuration for a node.
type ExecutorDefinition struct {
	Name string `json:"name" yaml:"name"`
	Mode string `json:"mode,omitempty" yaml:"mode,omitempty"` // Execution mode for multi-step executors
}

// ConditionDefinition represents a condition for node execution.
type ConditionDefinition struct {
	Key    string `json:"key" yaml:"key"`
	Value  string `json:"value" yaml:"value"`
	OnSkip string `json:"onSkip" yaml:"onSkip"`
}

// nodeDefinitionAlias is used to avoid infinite recursion during marshaling/unmarshaling.
type nodeDefinitionAlias NodeDefinition

// MarshalYAML implements custom YAML marshaling for NodeDefinition.
// It converts the Meta interface{} field to a JSON-encoded string for proper serialization.
func (nd *NodeDefinition) MarshalYAML() (interface{}, error) {
	// Create an alias to avoid infinite recursion
	alias := nodeDefinitionAlias(*nd)

	// If Meta is nil or empty, marshal as-is
	if alias.Meta == nil {
		return alias, nil
	}

	// JSON-encode the Meta field to preserve its structure
	metaJSON, err := json.Marshal(alias.Meta)
	if err != nil {
		return nil, err
	}

	// Replace Meta with the JSON string
	alias.Meta = string(metaJSON)

	return alias, nil
}

// UnmarshalYAML implements custom YAML unmarshaling for NodeDefinition.
// It parses the Meta field from a JSON-encoded string back to interface{}.
func (nd *NodeDefinition) UnmarshalYAML(value *yaml.Node) error {
	// Create an alias to avoid infinite recursion
	var alias nodeDefinitionAlias

	// Unmarshal into the alias
	if err := value.Decode(&alias); err != nil {
		return err
	}

	// Copy all fields from alias to nd
	*nd = NodeDefinition(alias)

	// If Meta is a string, try to parse it as JSON
	if metaStr, ok := nd.Meta.(string); ok && metaStr != "" {
		var metaData interface{}
		if err := json.Unmarshal([]byte(metaStr), &metaData); err != nil {
			// If JSON parsing fails, keep the string value
			// This allows backward compatibility with non-JSON Meta values
			return nil
		}
		nd.Meta = metaData
	}

	return nil
}
