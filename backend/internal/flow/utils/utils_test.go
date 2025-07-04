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

package utils

import (
	"testing"

	"github.com/asgardeo/thunder/internal/flow/constants"
	"github.com/asgardeo/thunder/internal/flow/jsonmodel"
	"github.com/asgardeo/thunder/internal/flow/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UtilsTestSuite struct {
	suite.Suite
}

func TestUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(UtilsTestSuite))
}

// BuildGraphFromDefinition tests

func (suite *UtilsTestSuite) TestBuildGraphFromDefinition_NilDefinition() {
	graph, err := BuildGraphFromDefinition(nil)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), graph)
	assert.Contains(suite.T(), err.Error(), "graph definition is nil or has no nodes")
}

func (suite *UtilsTestSuite) TestBuildGraphFromDefinition_EmptyNodes() {
	definition := &jsonmodel.GraphDefinition{
		ID:    "test-graph",
		Type:  "authentication",
		Nodes: []jsonmodel.NodeDefinition{},
	}

	graph, err := BuildGraphFromDefinition(definition)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), graph)
	assert.Contains(suite.T(), err.Error(), "graph definition is nil or has no nodes")
}

func (suite *UtilsTestSuite) TestBuildGraphFromDefinition_InvalidGraphType() {
	definition := &jsonmodel.GraphDefinition{
		ID:   "test-graph",
		Type: "invalid-type",
		Nodes: []jsonmodel.NodeDefinition{
			{
				ID:   "test-node",
				Type: "PROMPT_ONLY",
			},
		},
	}

	graph, err := BuildGraphFromDefinition(definition)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), graph)
	assert.Contains(suite.T(), err.Error(), "error while retrieving graph type")
}

func (suite *UtilsTestSuite) TestBuildGraphFromDefinition_ValidSingleNode() {
	definition := &jsonmodel.GraphDefinition{
		ID:   "test-graph",
		Type: "AUTHENTICATION",
		Nodes: []jsonmodel.NodeDefinition{
			{
				ID:   "start-node",
				Type: "PROMPT_ONLY",
				InputData: []jsonmodel.InputDefinition{
					{Name: "username", Type: "string", Required: true},
				},
			},
		},
	}

	graph, err := BuildGraphFromDefinition(definition)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), graph)
	assert.Equal(suite.T(), "test-graph", graph.GetID())
	assert.Equal(suite.T(), constants.FlowTypeAuthentication, graph.GetType())

	// Should have one node
	nodes := graph.GetNodes()
	assert.Len(suite.T(), nodes, 1)
	assert.Contains(suite.T(), nodes, "start-node")

	// Should set start node
	assert.Equal(suite.T(), "start-node", graph.GetStartNodeID())
}

func (suite *UtilsTestSuite) TestBuildGraphFromDefinition_ValidMultipleNodes() {
	definition := &jsonmodel.GraphDefinition{
		ID:   "test-graph",
		Type: "AUTHENTICATION",
		Nodes: []jsonmodel.NodeDefinition{
			{
				ID:   "start-node",
				Type: "PROMPT_ONLY",
				Next: []string{"end-node"},
			},
			{
				ID:   "end-node",
				Type: "TASK_EXECUTION",
			},
		},
	}

	graph, err := BuildGraphFromDefinition(definition)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), graph)

	// Should have two nodes
	nodes := graph.GetNodes()
	assert.Len(suite.T(), nodes, 2)
	assert.Contains(suite.T(), nodes, "start-node")
	assert.Contains(suite.T(), nodes, "end-node")

	// Should have edge
	edges := graph.GetEdges()
	assert.Contains(suite.T(), edges, "start-node")
	assert.Contains(suite.T(), edges["start-node"], "end-node")

	// Should set start node (node with no previous nodes)
	assert.Equal(suite.T(), "start-node", graph.GetStartNodeID())
}

func (suite *UtilsTestSuite) TestBuildGraphFromDefinition_InvalidNodeType() {
	definition := &jsonmodel.GraphDefinition{
		ID:   "test-graph",
		Type: "AUTHENTICATION",
		Nodes: []jsonmodel.NodeDefinition{
			{
				ID:   "invalid-node",
				Type: "INVALID_TYPE",
			},
		},
	}

	graph, err := BuildGraphFromDefinition(definition)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), graph)
	assert.Contains(suite.T(), err.Error(), "failed to create node")
}

func (suite *UtilsTestSuite) TestBuildGraphFromDefinition_WithExecutor() {
	definition := &jsonmodel.GraphDefinition{
		ID:   "test-graph",
		Type: "AUTHENTICATION",
		Nodes: []jsonmodel.NodeDefinition{
			{
				ID:   "exec-node",
				Type: "TASK_EXECUTION",
				Executor: jsonmodel.ExecutorDefinition{
					Name: "AttributeCollector",
					Properties: map[string]string{
						"attribute": "email",
					},
				},
			},
		},
	}

	graph, err := BuildGraphFromDefinition(definition)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), graph)

	node, exists := graph.GetNode("exec-node")
	assert.True(suite.T(), exists)
	assert.NotNil(suite.T(), node.GetExecutorConfig())
	assert.Equal(suite.T(), "AttributeCollector", node.GetExecutorConfig().Name)
}

func (suite *UtilsTestSuite) TestBuildGraphFromDefinition_WithAuthSuccessNode() {
	definition := &jsonmodel.GraphDefinition{
		ID:   "test-graph",
		Type: "AUTHENTICATION",
		Nodes: []jsonmodel.NodeDefinition{
			{
				ID:   "auth-success",
				Type: "AUTHENTICATION_SUCCESS",
			},
		},
	}

	graph, err := BuildGraphFromDefinition(definition)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), graph)

	node, exists := graph.GetNode("auth-success")
	assert.True(suite.T(), exists)
	// AUTH_SUCCESS nodes should get default AuthAssertExecutor
	assert.NotNil(suite.T(), node.GetExecutorConfig())
	assert.Equal(suite.T(), "AuthAssertExecutor", node.GetExecutorConfig().Name)
}

func (suite *UtilsTestSuite) TestBuildGraphFromDefinition_InvalidExecutor() {
	definition := &jsonmodel.GraphDefinition{
		ID:   "test-graph",
		Type: "AUTHENTICATION",
		Nodes: []jsonmodel.NodeDefinition{
			{
				ID:   "exec-node",
				Type: "TASK_EXECUTION",
				Executor: jsonmodel.ExecutorDefinition{
					Name: "InvalidExecutor",
				},
			},
		},
	}

	graph, err := BuildGraphFromDefinition(definition)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), graph)
	assert.Contains(suite.T(), err.Error(), "error while getting executor")
}

func (suite *UtilsTestSuite) TestBuildGraphFromDefinition_InvalidEdge() {
	definition := &jsonmodel.GraphDefinition{
		ID:   "test-graph",
		Type: "AUTHENTICATION",
		Nodes: []jsonmodel.NodeDefinition{
			{
				ID:   "start-node",
				Type: "PROMPT_ONLY",
				Next: []string{"nonexistent-node"},
			},
		},
	}

	graph, err := BuildGraphFromDefinition(definition)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), graph)
	assert.Contains(suite.T(), err.Error(), "failed to add edge")
}

func (suite *UtilsTestSuite) TestBuildGraphFromDefinition_RegistrationType() {
	definition := &jsonmodel.GraphDefinition{
		ID:   "reg-graph",
		Type: "REGISTRATION",
		Nodes: []jsonmodel.NodeDefinition{
			{
				ID:   "reg-node",
				Type: "PROMPT_ONLY",
			},
		},
	}

	graph, err := BuildGraphFromDefinition(definition)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), graph)
	assert.Equal(suite.T(), constants.FlowTypeRegistration, graph.GetType())
}

// GetExecutorByName tests

func (suite *UtilsTestSuite) TestGetExecutorByName_NilConfig() {
	executor, err := GetExecutorByName(nil)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), executor)
	assert.Contains(suite.T(), err.Error(), "executor configuration cannot be nil")
}

func (suite *UtilsTestSuite) TestGetExecutorByName_EmptyName() {
	config := &model.ExecutorConfig{
		Name: "",
	}

	executor, err := GetExecutorByName(config)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), executor)
	assert.Contains(suite.T(), err.Error(), "executor name cannot be empty")
}

func (suite *UtilsTestSuite) TestGetExecutorByName_UnknownExecutor() {
	config := &model.ExecutorConfig{
		Name: "UnknownExecutor",
	}

	executor, err := GetExecutorByName(config)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), executor)
	assert.Contains(suite.T(), err.Error(), "executor with name UnknownExecutor not found")
}

func (suite *UtilsTestSuite) TestGetExecutorByName_AttributeCollector() {
	config := &model.ExecutorConfig{
		Name: "AttributeCollector",
		Properties: map[string]string{
			"attribute": "email",
		},
	}

	executor, err := GetExecutorByName(config)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), executor)
}

func (suite *UtilsTestSuite) TestGetExecutorByName_ProvisioningExecutor() {
	config := &model.ExecutorConfig{
		Name: "ProvisioningExecutor",
		Properties: map[string]string{
			"userstore": "primary",
		},
	}

	executor, err := GetExecutorByName(config)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), executor)
}

func (suite *UtilsTestSuite) TestGetExecutorByName_AuthAssertExecutor() {
	config := &model.ExecutorConfig{
		Name: "AuthAssertExecutor",
		Properties: map[string]string{
			"assertion_type": "saml",
		},
	}

	executor, err := GetExecutorByName(config)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), executor)
}

// Note: Some executor tests may fail due to missing IDP setup, but that's expected in unit tests
// The important thing is that we're testing the code paths and error handling
