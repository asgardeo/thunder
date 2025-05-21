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
	"encoding/json"
	"errors"

	"github.com/google/uuid"
)

// GraphInterface defines the graph structure
type GraphInterface interface {
	AddNode(node Node)
	AddEdge(fromNodeID, toNodeID string)
	ToJSON() (string, error)
}

// Graph implements the GraphInterface for the flow execution
type Graph struct {
	ID          string
	Nodes       map[string]Node
	Edges       map[string][]string
	StartNodeID string
}

// NewGraph creates a new Graph
func NewGraph(startNodeID string) GraphInterface {
	return &Graph{
		ID:          uuid.New().String(),
		Nodes:       make(map[string]Node),
		Edges:       make(map[string][]string),
		StartNodeID: startNodeID,
	}
}

// AddNode adds a node to the graph
func (g *Graph) AddNode(node Node) {
	g.Nodes[node.ID] = node
}

// AddEdge adds an edge from one node to another
func (g *Graph) AddEdge(fromNodeID, toNodeID string) {
	if _, exists := g.Edges[fromNodeID]; !exists {
		g.Edges[fromNodeID] = []string{}
	}
	g.Edges[fromNodeID] = append(g.Edges[fromNodeID], toNodeID)
}

// ToJSON converts the graph to a JSON string representation
func (g *Graph) ToJSON() (string, error) {
	type JSONNode struct {
		ID   string `json:"id"`
		Page Step   `json:"page,omitempty"`
		Type string `json:"type"`
	}

	type JSONGraph struct {
		Nodes       map[string]JSONNode `json:"nodes"`
		Edges       map[string][]string `json:"edges"`
		StartNodeID string              `json:"startNodeId"`
	}

	jsonGraph := JSONGraph{
		Nodes:       make(map[string]JSONNode),
		Edges:       g.Edges,
		StartNodeID: g.StartNodeID,
	}

	// Convert nodes to JSONNode
	for id, node := range g.Nodes {
		jsonGraph.Nodes[id] = JSONNode{
			ID:   id,
			Page: node.Page,
			Type: node.Type,
		}
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(jsonGraph)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// NodeInterface defines the interface for nodes in the graph
type NodeInterface interface {
	Execute(ctx *FlowContext) (*ExecutorResponse, error)
}

// Node implements the NodeInterface
type Node struct {
	ID             string
	Type           string
	IsStartNode    bool
	IsFinalNode    bool
	NextNodeID     string
	PreviousNodeID string
	Executor       ExecutorInterface
	Page           Step
}

// NewNode creates a new node with the given details
func NewNode(id string, nodeType string, isStartNode bool) NodeInterface {
	return &Node{
		ID:          id,
		Type:        nodeType,
		IsStartNode: isStartNode,
	}
}

// Execute executes the node's executor
func (n *Node) Execute(ctx *FlowContext) (*ExecutorResponse, error) {
	if n.Executor == nil {
		return nil, errors.New("executor is not set")
	}
	return n.Executor.Execute(ctx)
}

// PromptNode represents a node that only takes user input
type PromptNode struct {
	*Node
}

// TaskExecutionNode represents a node that executes a task
type TaskExecutionNode struct {
	*Node
}
