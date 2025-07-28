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

package model

import (
	"testing"

	"github.com/asgardeo/thunder/internal/flow/constants"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const testNodeID = "test-node"

type ModelTestSuite struct {
	suite.Suite
}

func TestModelTestSuite(t *testing.T) {
	suite.Run(t, new(ModelTestSuite))
}

// Basic Graph functionality tests - these should work with the existing Graph implementation

func (suite *ModelTestSuite) TestNewGraph_WithIDAndType() {
	graphID := "test-graph-123"
	graphType := constants.FlowTypeAuthentication

	graph := NewGraph(graphID, graphType)

	assert.NotNil(suite.T(), graph)
	assert.Equal(suite.T(), graphID, graph.GetID())
	assert.Equal(suite.T(), graphType, graph.GetType())
	assert.NotNil(suite.T(), graph.GetNodes())
	assert.Empty(suite.T(), graph.GetNodes())
	assert.NotNil(suite.T(), graph.GetEdges())
	assert.Empty(suite.T(), graph.GetEdges())
}

func (suite *ModelTestSuite) TestNewGraph_WithEmptyID() {
	graphType := constants.FlowTypeAuthentication

	graph := NewGraph("", graphType)

	assert.NotNil(suite.T(), graph)
	assert.NotEmpty(suite.T(), graph.GetID(), "Should generate UUID when ID is empty")
	assert.Equal(suite.T(), graphType, graph.GetType())
}

func (suite *ModelTestSuite) TestNewGraph_WithEmptyType() {
	graphID := "test-graph-123"

	graph := NewGraph(graphID, "")

	assert.NotNil(suite.T(), graph)
	assert.Equal(suite.T(), graphID, graph.GetID())
	assert.Equal(suite.T(), constants.FlowTypeAuthentication, graph.GetType(), "Should default to authentication type")
}

func (suite *ModelTestSuite) TestGraph_AddNode_Success() {
	graph := NewGraph("test-graph", constants.FlowTypeAuthentication)
	node := NewPromptOnlyNode("prompt-node", false, false)

	err := graph.AddNode(node)

	assert.NoError(suite.T(), err)
	nodes := graph.GetNodes()
	assert.Len(suite.T(), nodes, 1)
	assert.Contains(suite.T(), nodes, "prompt-node")
	assert.Equal(suite.T(), node, nodes["prompt-node"])
}

func (suite *ModelTestSuite) TestGraph_AddNode_NilNode() {
	graph := NewGraph("test-graph", constants.FlowTypeAuthentication)

	err := graph.AddNode(nil)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "node cannot be nil")
	assert.Empty(suite.T(), graph.GetNodes())
}

func (suite *ModelTestSuite) TestGraph_GetNode_Exists() {
	graph := NewGraph("test-graph", constants.FlowTypeAuthentication)
	node := NewPromptOnlyNode("prompt-node", false, false)
	err := graph.AddNode(node)
	assert.NoError(suite.T(), err)

	retrievedNode, exists := graph.GetNode("prompt-node")

	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), node, retrievedNode)
}

func (suite *ModelTestSuite) TestGraph_GetNode_NotExists() {
	graph := NewGraph("test-graph", constants.FlowTypeAuthentication)

	retrievedNode, exists := graph.GetNode("non-existent")

	assert.False(suite.T(), exists)
	assert.Nil(suite.T(), retrievedNode)
}

// Comprehensive Edge Management Tests
func (suite *ModelTestSuite) TestGraph_AddEdge_Success() {
	graph := NewGraph("test-graph", constants.FlowTypeAuthentication)
	node1 := NewPromptOnlyNode("node1", false, false)
	node2 := NewPromptOnlyNode("node2", false, false)
	err := graph.AddNode(node1)
	assert.NoError(suite.T(), err)
	err = graph.AddNode(node2)
	assert.NoError(suite.T(), err)

	err = graph.AddEdge("node1", "node2")

	assert.NoError(suite.T(), err)
	edges := graph.GetEdges()
	assert.Contains(suite.T(), edges, "node1")
	assert.Contains(suite.T(), edges["node1"], "node2")
}

func (suite *ModelTestSuite) TestGraph_AddEdge_EmptyFromNodeID() {
	graph := NewGraph("test-graph", constants.FlowTypeAuthentication)

	err := graph.AddEdge("", "node2")

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "fromNodeID and toNodeID cannot be empty")
}

func (suite *ModelTestSuite) TestGraph_AddEdge_EmptyToNodeID() {
	graph := NewGraph("test-graph", constants.FlowTypeAuthentication)

	err := graph.AddEdge("node1", "")

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "fromNodeID and toNodeID cannot be empty")
}

func (suite *ModelTestSuite) TestGraph_AddEdge_FromNodeNotExists() {
	graph := NewGraph("test-graph", constants.FlowTypeAuthentication)
	node2 := NewPromptOnlyNode("node2", false, false)
	err := graph.AddNode(node2)
	assert.NoError(suite.T(), err)

	err = graph.AddEdge("nonexistent", "node2")

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "node with fromNodeID does not exist")
}

func (suite *ModelTestSuite) TestGraph_AddEdge_ToNodeNotExists() {
	graph := NewGraph("test-graph", constants.FlowTypeAuthentication)
	node1 := NewPromptOnlyNode("node1", false, false)
	err := graph.AddNode(node1)
	assert.NoError(suite.T(), err)

	err = graph.AddEdge("node1", "nonexistent")

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "node with toNodeID does not exist")
}

func (suite *ModelTestSuite) TestGraph_RemoveEdge_Success() {
	graph := NewGraph("test-graph", constants.FlowTypeAuthentication)
	node1 := NewPromptOnlyNode("node1", false, false)
	node2 := NewPromptOnlyNode("node2", false, false)
	err := graph.AddNode(node1)
	assert.NoError(suite.T(), err)
	err = graph.AddNode(node2)
	assert.NoError(suite.T(), err)
	err = graph.AddEdge("node1", "node2")
	assert.NoError(suite.T(), err)

	err = graph.RemoveEdge("node1", "node2")

	assert.NoError(suite.T(), err)
	edges := graph.GetEdges()
	if edgeList, exists := edges["node1"]; exists {
		assert.NotContains(suite.T(), edgeList, "node2")
	}
}

func (suite *ModelTestSuite) TestGraph_RemoveEdge_EmptyFromNodeID() {
	graph := NewGraph("test-graph", constants.FlowTypeAuthentication)

	err := graph.RemoveEdge("", "node2")

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "fromNodeID and toNodeID cannot be empty")
}

func (suite *ModelTestSuite) TestGraph_RemoveEdge_FromNodeNotExists() {
	graph := NewGraph("test-graph", constants.FlowTypeAuthentication)
	node2 := NewPromptOnlyNode("node2", false, false)
	err := graph.AddNode(node2)
	assert.NoError(suite.T(), err)

	err = graph.RemoveEdge("nonexistent", "node2")

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "node with fromNodeID does not exist")
}

// Node/Edge Management Tests
func (suite *ModelTestSuite) TestGraph_SetNodes_WithValidNodes() {
	graph := NewGraph("test-graph", constants.FlowTypeAuthentication)
	node1 := NewPromptOnlyNode("node1", false, false)
	node2 := NewPromptOnlyNode("node2", false, false)
	nodes := map[string]NodeInterface{
		"node1": node1,
		"node2": node2,
	}

	graph.SetNodes(nodes)

	retrievedNodes := graph.GetNodes()
	assert.Equal(suite.T(), nodes, retrievedNodes)
	assert.Len(suite.T(), retrievedNodes, 2)
}

func (suite *ModelTestSuite) TestGraph_SetNodes_WithNilNodes() {
	graph := NewGraph("test-graph", constants.FlowTypeAuthentication)

	graph.SetNodes(nil)

	retrievedNodes := graph.GetNodes()
	assert.NotNil(suite.T(), retrievedNodes)
	assert.Empty(suite.T(), retrievedNodes)
}

func (suite *ModelTestSuite) TestGraph_SetEdges_WithValidEdges() {
	graph := NewGraph("test-graph", constants.FlowTypeAuthentication)
	edges := map[string][]string{
		"node1": {"node2", "node3"},
		"node2": {"node3"},
	}

	graph.SetEdges(edges)

	retrievedEdges := graph.GetEdges()
	assert.Equal(suite.T(), edges, retrievedEdges)
}

func (suite *ModelTestSuite) TestGraph_SetEdges_WithNilEdges() {
	graph := NewGraph("test-graph", constants.FlowTypeAuthentication)

	graph.SetEdges(nil)

	retrievedEdges := graph.GetEdges()
	assert.NotNil(suite.T(), retrievedEdges)
	assert.Empty(suite.T(), retrievedEdges)
}

// Start Node Tests
func (suite *ModelTestSuite) TestGraph_GetStartNodeID_NotSet() {
	graph := NewGraph("test-graph", constants.FlowTypeAuthentication)

	startNodeID := graph.GetStartNodeID()

	assert.Empty(suite.T(), startNodeID)
}

func (suite *ModelTestSuite) TestGraph_SetStartNode_Success() {
	graph := NewGraph("test-graph", constants.FlowTypeAuthentication)
	node := NewPromptOnlyNode("start-node", false, false)
	err := graph.AddNode(node)
	assert.NoError(suite.T(), err)

	err = graph.SetStartNode("start-node")

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "start-node", graph.GetStartNodeID())
}

func (suite *ModelTestSuite) TestGraph_SetStartNode_NodeNotExists() {
	graph := NewGraph("test-graph", constants.FlowTypeAuthentication)

	err := graph.SetStartNode("nonexistent-node")

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "node with startNodeID does not exist")
}

func (suite *ModelTestSuite) TestGraph_GetStartNode_Success() {
	graph := NewGraph("test-graph", constants.FlowTypeAuthentication)
	node := NewPromptOnlyNode("start-node", false, false)
	err := graph.AddNode(node)
	assert.NoError(suite.T(), err)
	err = graph.SetStartNode("start-node")
	assert.NoError(suite.T(), err)

	startNode, err := graph.GetStartNode()

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), node, startNode)
}

func (suite *ModelTestSuite) TestGraph_GetStartNode_NotSet() {
	graph := NewGraph("test-graph", constants.FlowTypeAuthentication)

	startNode, err := graph.GetStartNode()

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "start node not set for the graph")
	assert.Nil(suite.T(), startNode)
}

func (suite *ModelTestSuite) TestGraph_GetStartNode_NodeNotExists() {
	graph := NewGraph("test-graph", constants.FlowTypeAuthentication)
	// Manually set start node ID without adding the actual node
	g := graph.(*Graph)
	g.startNodeID = "nonexistent-node"

	startNode, err := graph.GetStartNode()

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "start node does not exist in the graph")
	assert.Nil(suite.T(), startNode)
}

// ToJSON Tests
func (suite *ModelTestSuite) TestGraph_ToJSON_EmptyGraph() {
	graph := NewGraph("test-graph", constants.FlowTypeAuthentication)

	jsonStr, err := graph.ToJSON()

	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), jsonStr)
	// Should be valid JSON
	assert.Contains(suite.T(), jsonStr, "test-graph")
}

func (suite *ModelTestSuite) TestGraph_ToJSON_WithNodes() {
	graph := NewGraph("test-graph", constants.FlowTypeAuthentication)
	node := NewPromptOnlyNode("prompt-node", false, false)
	err := graph.AddNode(node)
	assert.NoError(suite.T(), err)

	jsonStr, err := graph.ToJSON()

	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), jsonStr)
	assert.Contains(suite.T(), jsonStr, "prompt-node")
}

// Node Tests
func (suite *ModelTestSuite) TestNewNode_Success() {
	nodeID := testNodeID
	nodeType := "PROMPT_ONLY"

	node, err := NewNode(nodeID, nodeType, false, false)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), node)
	assert.Equal(suite.T(), nodeID, node.GetID())
	assert.Equal(suite.T(), constants.NodeTypePromptOnly, node.GetType())
	assert.False(suite.T(), node.IsStartNode())
	assert.False(suite.T(), node.IsFinalNode())
}

func (suite *ModelTestSuite) TestNewNode_EmptyType() {
	nodeID := testNodeID

	node, err := NewNode(nodeID, "", false, false)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), node)
	assert.Contains(suite.T(), err.Error(), "node type cannot be empty")
}

func (suite *ModelTestSuite) TestNewNode_UnsupportedType() {
	nodeID := testNodeID

	node, err := NewNode(nodeID, "UNSUPPORTED_TYPE", false, false)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), node)
	assert.Contains(suite.T(), err.Error(), "unsupported node type")
}

func (suite *ModelTestSuite) TestNode_SetAsStartNode() {
	node := NewPromptOnlyNode(testNodeID, false, false)

	node.SetAsStartNode()

	assert.True(suite.T(), node.IsStartNode())
}

func (suite *ModelTestSuite) TestNode_SetAsFinalNode() {
	node := NewPromptOnlyNode(testNodeID, false, false)

	node.SetAsFinalNode()

	assert.True(suite.T(), node.IsFinalNode())
}

func (suite *ModelTestSuite) TestNode_NextNodeManagement() {
	node := NewPromptOnlyNode(testNodeID, false, false)

	// Initially empty
	assert.Empty(suite.T(), node.GetNextNodeList())

	// Add next node
	node.AddNextNodeID("next-node")
	assert.Contains(suite.T(), node.GetNextNodeList(), "next-node")

	// Add duplicate - should not be added
	node.AddNextNodeID("next-node")
	assert.Len(suite.T(), node.GetNextNodeList(), 1) // Still only one

	// Add empty node ID - should be ignored
	node.AddNextNodeID("")
	assert.Len(suite.T(), node.GetNextNodeList(), 1)

	// Set next node list
	newList := []string{"node1", "node2"}
	node.SetNextNodeList(newList)
	assert.Equal(suite.T(), newList, node.GetNextNodeList())

	// Remove next node
	node.RemoveNextNodeID("node1")
	assert.NotContains(suite.T(), node.GetNextNodeList(), "node1")
	assert.Contains(suite.T(), node.GetNextNodeList(), "node2")

	// Remove empty ID - should be ignored
	node.RemoveNextNodeID("")
	assert.Contains(suite.T(), node.GetNextNodeList(), "node2") // Should still be there
}

func (suite *ModelTestSuite) TestNode_PreviousNodeManagement() {
	node := NewPromptOnlyNode(testNodeID, false, false)

	// Initially empty
	assert.Empty(suite.T(), node.GetPreviousNodeList())

	// Add previous node
	node.AddPreviousNodeID("prev-node")
	assert.Contains(suite.T(), node.GetPreviousNodeList(), "prev-node")

	// Add duplicate - should not be added
	node.AddPreviousNodeID("prev-node")
	assert.Len(suite.T(), node.GetPreviousNodeList(), 1) // Still only one

	// Set previous node list
	newList := []string{"prev1", "prev2"}
	node.SetPreviousNodeList(newList)
	assert.Equal(suite.T(), newList, node.GetPreviousNodeList())

	// Remove previous node
	node.RemovePreviousNodeID("prev1")
	assert.NotContains(suite.T(), node.GetPreviousNodeList(), "prev1")
	assert.Contains(suite.T(), node.GetPreviousNodeList(), "prev2")

	// Remove non-existent node - should be ignored
	node.RemovePreviousNodeID("nonexistent")
	assert.Contains(suite.T(), node.GetPreviousNodeList(), "prev2") // Should still be there
}

func (suite *ModelTestSuite) TestNode_InputDataManagement() {
	node := NewPromptOnlyNode(testNodeID, false, false)

	// Initially empty slice, not nil
	inputData := node.GetInputData()
	assert.NotNil(suite.T(), inputData)
	assert.Empty(suite.T(), inputData)

	// Set input data
	newInputData := []InputData{
		{Name: "username", Type: "string", Required: true},
		{Name: "password", Type: "string", Required: false},
	}
	node.SetInputData(newInputData)
	assert.Equal(suite.T(), newInputData, node.GetInputData())
}

func (suite *ModelTestSuite) TestNode_ExecutorManagement() {
	node := NewTaskExecutionNode(testNodeID, false, false)

	// Initially nil
	assert.Nil(suite.T(), node.GetExecutor())

	// Set executor - test with nil (basic test)
	node.SetExecutor(nil)
	assert.Nil(suite.T(), node.GetExecutor())
}

// Task Execution Node Tests
func (suite *ModelTestSuite) TestNewTaskExecutionNode() {
	nodeID := "task-node"

	node := NewTaskExecutionNode(nodeID, false, false)

	assert.NotNil(suite.T(), node)
	assert.Equal(suite.T(), nodeID, node.GetID())
	assert.Equal(suite.T(), constants.NodeTypeTaskExecution, node.GetType())
}

// Prompt Only Node Tests
func (suite *ModelTestSuite) TestNewPromptOnlyNode() {
	nodeID := "prompt-node"

	node := NewPromptOnlyNode(nodeID, false, false)

	assert.NotNil(suite.T(), node)
	assert.Equal(suite.T(), nodeID, node.GetID())
	assert.Equal(suite.T(), constants.NodeTypePromptOnly, node.GetType())
}

// Decision Node Tests
func (suite *ModelTestSuite) TestNewDecisionNode() {
	nodeID := "decision-node"

	node := NewDecisionNode(nodeID, false, false)

	assert.NotNil(suite.T(), node)
	assert.Equal(suite.T(), nodeID, node.GetID())
	assert.Equal(suite.T(), constants.NodeTypeDecision, node.GetType())
}
