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

// graphDefinition represents the direct graph structure from JSON
type graphDefinition struct {
	ID    string           `json:"id"`
	Type  string           `json:"type"`
	Nodes []nodeDefinition `json:"nodes"`
}

// nodeDefinition represents a node in the graph definition
type nodeDefinition struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Meta       interface{}            `json:"meta,omitempty"`
	Inputs     []inputDefinition      `json:"inputs"`
	Actions    []actionDefinition     `json:"actions,omitempty"`
	Executor   executorDefinition     `json:"executor"`
	OnSuccess  string                 `json:"onSuccess,omitempty"`
	OnFailure  string                 `json:"onFailure,omitempty"`
	Condition  *conditionDefinition   `json:"condition,omitempty"`
}

// inputDefinition represents an input parameter for a node
type inputDefinition struct {
	Ref        string `json:"ref,omitempty"`
	Identifier string `json:"identifier"`
	Type       string `json:"type"`
	Required   bool   `json:"required"`
}

// actionDefinition represents an action that can be triggered from a node
type actionDefinition struct {
	Ref      string `json:"ref"`
	NextNode string `json:"nextNode"`
}

// executorDefinition represents the executor configuration for a node
type executorDefinition struct {
	Name string `json:"name"`
}

// conditionDefinition represents a condition that must be met for a node to execute.
// If specified, the node will only execute when the resolved value of Key matches Value.
// OnSkip specifies which node to skip to if the condition is not met.
type conditionDefinition struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	OnSkip string `json:"onSkip"`
}
