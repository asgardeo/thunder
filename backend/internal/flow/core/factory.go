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

package core

import (
	"errors"
	"fmt"

	"github.com/asgardeo/thunder/internal/flow/common"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

// FlowFactoryInterface defines the interface for creating flow graph components.
type FlowFactoryInterface interface {
	CreateNode(id, _type string, properties map[string]interface{}, isStartNode, isFinalNode bool) (
		NodeInterface, error)
	CreateGraph(id string, _type common.FlowType) GraphInterface
	CreateExecutor(name string, executorType common.ExecutorType,
		defaultInputs, prerequisites []common.Input) ExecutorInterface
	CloneNode(source NodeInterface) (NodeInterface, error)
	CloneNodes(nodes map[string]NodeInterface) (map[string]NodeInterface, error)
}

// flowFactory is the concrete implementation of FlowFactoryInterface
type flowFactory struct{}

func newFlowFactory() FlowFactoryInterface {
	return &flowFactory{}
}

// CreateNode creates a new node based on the provided type and properties
func (f *flowFactory) CreateNode(id, _type string, properties map[string]interface{},
	isStartNode, isFinalNode bool) (NodeInterface, error) {
	var nodeType common.NodeType
	if _type == "" {
		return nil, errors.New("node type cannot be empty")
	} else {
		nodeType = common.NodeType(_type)
	}
	if properties == nil {
		properties = make(map[string]interface{})
	}

	switch nodeType {
	case common.NodeTypeTaskExecution:
		return newTaskExecutionNode(id, properties, isStartNode, isFinalNode), nil
	case common.NodeTypePrompt:
		return newPromptNode(id, properties, isStartNode, isFinalNode), nil
	case common.NodeTypeAuthSuccess:
		return newTaskExecutionNode(id, properties, isStartNode, isFinalNode), nil
	case common.NodeTypeRegistrationStart:
		return newTaskExecutionNode(id, properties, isStartNode, isFinalNode), nil
	default:
		return nil, errors.New("unsupported node type: " + _type)
	}
}

// CreateGraph creates a new graph with the given ID and type
func (f *flowFactory) CreateGraph(id string, _type common.FlowType) GraphInterface {
	if id == "" {
		id = sysutils.GenerateUUID()
	}
	if _type == "" {
		_type = common.FlowTypeAuthentication
	}

	return &graph{
		id:    id,
		_type: _type,
		nodes: make(map[string]NodeInterface),
		edges: make(map[string][]string),
	}
}

// CreateExecutor creates a new executor with the given properties
func (f *flowFactory) CreateExecutor(name string, executorType common.ExecutorType,
	defaultInputs, prerequisites []common.Input) ExecutorInterface {
	return newExecutor(name, executorType, defaultInputs, prerequisites)
}

// CloneNode creates a deep copy of a given node
func (f *flowFactory) CloneNode(source NodeInterface) (NodeInterface, error) {
	if source == nil {
		return nil, errors.New("source node cannot be nil")
	}

	// Deep copy properties
	var propertiesCopy map[string]interface{}
	if source.GetProperties() != nil {
		propertiesCopy = sysutils.DeepCopyMap(source.GetProperties())
	} else {
		propertiesCopy = make(map[string]interface{})
	}

	// Create new node
	nodeCopy, err := f.CreateNode(
		source.GetID(),
		string(source.GetType()),
		propertiesCopy,
		source.IsStartNode(),
		source.IsFinalNode(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to clone node: %w", err)
	}

	// Copy node connections and metadata
	nodeCopy.SetNextNodeList(append([]string{}, source.GetNextNodeList()...))
	nodeCopy.SetPreviousNodeList(append([]string{}, source.GetPreviousNodeList()...))
	nodeCopy.SetInputs(append([]common.Input{}, source.GetInputs()...))

	// Copy condition if present
	if sourceCondition := source.GetCondition(); sourceCondition != nil {
		nodeCopy.SetCondition(&NodeCondition{
			Key:   sourceCondition.Key,
			Value: sourceCondition.Value,
		})
	}

	// Copy executor name if the node is executor-backed
	if executableSource, ok := source.(ExecutorBackedNodeInterface); ok {
		if executableCopy, ok := nodeCopy.(ExecutorBackedNodeInterface); ok {
			executableCopy.SetExecutorName(executableSource.GetExecutorName())
		} else {
			return nil, errors.New("mismatch in node types during cloning. copy is not executor-backed")
		}
	}

	return nodeCopy, nil
}

// CloneNodes creates deep copies of a map of nodes
func (f *flowFactory) CloneNodes(nodes map[string]NodeInterface) (map[string]NodeInterface, error) {
	if nodes == nil {
		return nil, nil
	}

	clonedNodes := make(map[string]NodeInterface, len(nodes))
	for id, node := range nodes {
		clonedNode, err := f.CloneNode(node)
		if err != nil {
			return nil, fmt.Errorf("failed to clone node %s: %w", id, err)
		}
		clonedNodes[id] = clonedNode
	}

	return clonedNodes, nil
}
