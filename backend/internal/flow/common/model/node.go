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

package model

import (
	"errors"
	"fmt"

	authncm "github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/flow/common/constants"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

// NodeResponse represents the response from a node execution
type NodeResponse struct {
	Status            constants.NodeStatus       `json:"status"`
	Type              constants.NodeResponseType `json:"type"`
	FailureReason     string                     `json:"failure_reason,omitempty"`
	RequiredData      []InputData                `json:"required_data,omitempty"`
	AdditionalData    map[string]string          `json:"additional_data,omitempty"`
	RedirectURL       string                     `json:"redirect_url,omitempty"`
	Actions           []Action                   `json:"actions,omitempty"`
	NextNodeID        string                     `json:"next_node_id,omitempty"`
	RuntimeData       map[string]string          `json:"runtime_data,omitempty"`
	AuthenticatedUser authncm.AuthenticatedUser  `json:"authenticated_user,omitempty"`
	Assertion         string                     `json:"assertion,omitempty"`
	ExecutionRecord   *NodeExecutionRecord       `json:"execution_record,omitempty"`
}

// NodeInterface defines the interface for nodes in the graph
type NodeInterface interface {
	sysutils.ClonableInterface
	Execute(ctx *NodeContext) (*NodeResponse, *serviceerror.ServiceError)
	GetID() string
	GetType() constants.NodeType
	IsStartNode() bool
	SetAsStartNode()
	IsFinalNode() bool
	SetAsFinalNode()
	GetNextNodeList() []string
	SetNextNodeList(nextNodeIDList []string)
	AddNextNodeID(nextNodeID string)
	RemoveNextNodeID(nextNodeID string)
	GetPreviousNodeList() []string
	SetPreviousNodeList(previousNodeIDList []string)
	AddPreviousNodeID(previousNodeID string)
	RemovePreviousNodeID(previousNodeID string)
	GetInputData() []InputData
	SetInputData(inputData []InputData)
	GetExecutorConfig() *ExecutorConfig
	SetExecutorConfig(executorConfig *ExecutorConfig)
	GetExecutor() ExecutorInterface
	SetExecutor(executor ExecutorInterface)
}

// Node implements the NodeInterface
type Node struct {
	id               string
	_type            constants.NodeType
	isStartNode      bool
	isFinalNode      bool
	nextNodeList     []string
	previousNodeList []string
	inputData        []InputData
	executorConfig   *ExecutorConfig
}

var _ NodeInterface = (*Node)(nil)

// NewNode creates a new Node with the given type and properties.
func NewNode(id string, _type string, isStartNode bool, isFinalNode bool) (NodeInterface, error) {
	var nodeType constants.NodeType
	if _type == "" {
		return nil, errors.New("node type cannot be empty")
	} else {
		nodeType = constants.NodeType(_type)
	}

	switch nodeType {
	case constants.NodeTypeTaskExecution:
		return NewTaskExecutionNode(id, isStartNode, isFinalNode), nil
	case constants.NodeTypeDecision:
		return NewDecisionNode(id, isStartNode, isFinalNode), nil
	case constants.NodeTypePromptOnly:
		return NewPromptOnlyNode(id, isStartNode, isFinalNode), nil
	case constants.NodeTypeAuthSuccess:
		return NewTaskExecutionNode(id, isStartNode, isFinalNode), nil
	default:
		return nil, errors.New("unsupported node type: " + _type)
	}
}

// Execute executes the node
func (n *Node) Execute(ctx *NodeContext) (*NodeResponse, *serviceerror.ServiceError) {
	return nil, nil
}

// GetID returns the node's ID
func (n *Node) GetID() string {
	return n.id
}

// GetType returns the node's type
func (n *Node) GetType() constants.NodeType {
	return n._type
}

// IsStartNode checks if the node is a start node
func (n *Node) IsStartNode() bool {
	return n.isStartNode
}

// SetAsStartNode sets the node as a start node
func (n *Node) SetAsStartNode() {
	n.isStartNode = true
}

// IsFinalNode checks if the node is a final node
func (n *Node) IsFinalNode() bool {
	return n.isFinalNode
}

// SetAsFinalNode sets the node as a final node
func (n *Node) SetAsFinalNode() {
	n.isFinalNode = true
}

// GetNextNodeList returns the list of next node IDs
func (n *Node) GetNextNodeList() []string {
	if n.nextNodeList == nil {
		return []string{}
	}
	return n.nextNodeList
}

// SetNextNodeList sets the list of next node IDs
func (n *Node) SetNextNodeList(nextNodeIDList []string) {
	if nextNodeIDList == nil {
		n.nextNodeList = []string{}
	} else {
		n.nextNodeList = nextNodeIDList
	}
}

// AddNextNodeID adds a next node ID to the list
func (n *Node) AddNextNodeID(nextNodeID string) {
	if nextNodeID == "" {
		return
	}
	if n.nextNodeList == nil {
		n.nextNodeList = []string{}
	}
	// Check for duplicates before adding
	for _, id := range n.nextNodeList {
		if id == nextNodeID {
			return
		}
	}
	n.nextNodeList = append(n.nextNodeList, nextNodeID)
}

// RemoveNextNodeID removes a next node ID from the list
func (n *Node) RemoveNextNodeID(nextNodeID string) {
	if nextNodeID == "" || n.nextNodeList == nil {
		return
	}

	for i, id := range n.nextNodeList {
		if id == nextNodeID {
			n.nextNodeList = append(n.nextNodeList[:i], n.nextNodeList[i+1:]...)
			return
		}
	}
}

// GetPreviousNodeList returns the list of previous node IDs
func (n *Node) GetPreviousNodeList() []string {
	if n.previousNodeList == nil {
		return []string{}
	}
	return n.previousNodeList
}

// SetPreviousNodeList sets the list of previous node IDs
func (n *Node) SetPreviousNodeList(previousNodeIDList []string) {
	if previousNodeIDList == nil {
		n.previousNodeList = []string{}
	} else {
		n.previousNodeList = previousNodeIDList
	}
}

// AddPreviousNodeID adds a previous node ID to the list
func (n *Node) AddPreviousNodeID(previousNodeID string) {
	if previousNodeID == "" {
		return
	}
	if n.previousNodeList == nil {
		n.previousNodeList = []string{}
	}
	// Check for duplicates before adding
	for _, id := range n.previousNodeList {
		if id == previousNodeID {
			return
		}
	}
	n.previousNodeList = append(n.previousNodeList, previousNodeID)
}

// RemovePreviousNodeID removes a previous node ID from the list
func (n *Node) RemovePreviousNodeID(previousNodeID string) {
	if previousNodeID == "" || n.previousNodeList == nil {
		return
	}

	for i, id := range n.previousNodeList {
		if id == previousNodeID {
			n.previousNodeList = append(n.previousNodeList[:i], n.previousNodeList[i+1:]...)
			return
		}
	}
}

// GetInputData returns the input data for the node
func (n *Node) GetInputData() []InputData {
	return n.inputData
}

// SetInputData sets the input data for the node
func (n *Node) SetInputData(inputData []InputData) {
	n.inputData = inputData
}

// GetExecutorConfig returns the executor configuration for the node
func (n *Node) GetExecutorConfig() *ExecutorConfig {
	return n.executorConfig
}

// SetExecutorConfig sets the executor configuration for the node
func (n *Node) SetExecutorConfig(executorConfig *ExecutorConfig) {
	n.executorConfig = executorConfig
}

// GetExecutor returns the executor associated with the node
func (n *Node) GetExecutor() ExecutorInterface {
	if n.executorConfig == nil {
		return nil
	}
	return n.executorConfig.Executor
}

// SetExecutor sets the executor for the node
func (n *Node) SetExecutor(executor ExecutorInterface) {
	if n.executorConfig == nil {
		n.executorConfig = &ExecutorConfig{}
		n.executorConfig.Name = executor.GetName()
	}
	n.executorConfig.Executor = executor
}

// Clone creates a deep copy of the Node
func (n *Node) Clone() (sysutils.ClonableInterface, error) {
	nextCopy := append([]string{}, n.nextNodeList...)
	prevCopy := append([]string{}, n.previousNodeList...)
	inputCopy := append([]InputData{}, n.inputData...)

	var execConfigCopy *ExecutorConfig
	if n.executorConfig != nil {
		execConfigCopy = &ExecutorConfig{
			Name:       n.executorConfig.Name,
			IdpName:    n.executorConfig.IdpName,
			Properties: sysutils.DeepCopyMapOfStrings(n.executorConfig.Properties),
			Executor:   n.executorConfig.Executor,
		}
	}

	nodeCopy, err := NewNode(n.id, string(n._type), n.isStartNode, n.isFinalNode)
	if err != nil {
		return nil, fmt.Errorf("failed to clone node: %w", err)
	}

	nodeCopy.SetNextNodeList(nextCopy)
	nodeCopy.SetPreviousNodeList(prevCopy)
	nodeCopy.SetInputData(inputCopy)
	nodeCopy.SetExecutorConfig(execConfigCopy)

	return nodeCopy, nil
}
