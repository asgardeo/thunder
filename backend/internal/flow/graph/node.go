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

package graph

import (
	"errors"

	"github.com/asgardeo/thunder/internal/flow/executors"
	"github.com/asgardeo/thunder/internal/flow/model"
)

// NodeInterface defines the interface for nodes in the graph
type NodeInterface interface {
	Execute(ctx *model.FlowContext) (*model.ExecutorResponse, error)
}

// Node implements the NodeInterface
type Node struct {
	ID             string
	Type           string
	IsStartNode    bool
	NextNodeID     string
	PreviousNodeID string
	Executor       executors.ExecutorInterface
	Page           model.Step
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
func (n *Node) Execute(ctx *model.FlowContext) (*model.ExecutorResponse, error) {
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
