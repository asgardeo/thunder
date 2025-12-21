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
	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
)

const taskExecNodeLoggerComponentName = "TaskExecutionNode"

// taskExecutionNode represents a node that executes a task via an executor
type taskExecutionNode struct {
	*node
	executorName string
	executor     ExecutorInterface
	mode         string
	onSuccess    string
	onFailure    string
}

// Ensure taskExecutionNode implements ExecutorBackedNodeInterface
var _ ExecutorBackedNodeInterface = (*taskExecutionNode)(nil)

// newTaskExecutionNode creates a new TaskExecutionNode with the given details.
func newTaskExecutionNode(id string, properties map[string]interface{}, isStartNode bool,
	isFinalNode bool) NodeInterface {
	return &taskExecutionNode{
		node: &node{
			id:               id,
			_type:            common.NodeTypeTaskExecution,
			properties:       properties,
			isStartNode:      isStartNode,
			isFinalNode:      isFinalNode,
			nextNodeList:     []string{},
			previousNodeList: []string{},
			inputs:           []common.Input{},
		},
		executorName: "",
		executor:     nil,
	}
}

// Execute executes the node's executor.
func (n *taskExecutionNode) Execute(ctx *NodeContext) (*common.NodeResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, taskExecNodeLoggerComponentName),
		log.String(log.LoggerKeyNodeID, n.GetID()),
		log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Executing task execution node")

	if n.executor == nil {
		logger.Error("No executor configured for the node")
		return nil, &serviceerror.InternalServerError
	}

	// Set node properties in context
	if len(n.GetProperties()) > 0 {
		ctx.NodeProperties = n.GetProperties()
	} else {
		ctx.NodeProperties = make(map[string]interface{})
	}

	// Set executor mode in context
	ctx.ExecutorMode = n.mode

	execResp, svcErr := n.triggerExecutor(ctx, logger)
	if svcErr != nil {
		return nil, svcErr
	}

	nodeResp := buildNodeResponse(execResp)

	// Set the next node ID based on execution outcome
	if nodeResp.Status == common.NodeStatusComplete {
		if n.onSuccess != "" {
			nodeResp.NextNodeID = n.onSuccess
		}
	} else if nodeResp.FailureReason != "" && n.onFailure != "" {
		// Change status to Forward so engine forwards execution to onFailure node
		nodeResp.Status = common.NodeStatusForward
		nodeResp.NextNodeID = n.onFailure

		// Store failure reason in RuntimeData so it's available to the onFailure handler
		if nodeResp.RuntimeData == nil {
			nodeResp.RuntimeData = make(map[string]string)
		}
		nodeResp.RuntimeData["failureReason"] = nodeResp.FailureReason
	}

	return nodeResp, nil
}

// triggerExecutor triggers the executor configured for the node.
func (n *taskExecutionNode) triggerExecutor(ctx *NodeContext, logger *log.Logger) (
	*common.ExecutorResponse, *serviceerror.ServiceError) {
	execResp, err := n.executor.Execute(ctx)
	if err != nil {
		logger.Error("Error executing node executor", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if execResp == nil {
		logger.Error("Executor returned a nil response")
		return nil, &serviceerror.InternalServerError
	}

	return execResp, nil
}

// buildNodeResponse constructs a NodeResponse from the ExecutorResponse.
func buildNodeResponse(execResp *common.ExecutorResponse) *common.NodeResponse {
	nodeResp := &common.NodeResponse{
		FailureReason:     execResp.FailureReason,
		Inputs:            execResp.Inputs,
		AdditionalData:    execResp.AdditionalData,
		RedirectURL:       execResp.RedirectURL,
		RuntimeData:       execResp.RuntimeData,
		AuthenticatedUser: execResp.AuthenticatedUser,
		Assertion:         execResp.Assertion,
	}
	if nodeResp.AdditionalData == nil {
		nodeResp.AdditionalData = make(map[string]string)
	}
	if nodeResp.RuntimeData == nil {
		nodeResp.RuntimeData = make(map[string]string)
	}
	if nodeResp.Inputs == nil {
		nodeResp.Inputs = make([]common.Input, 0)
	}
	if nodeResp.Actions == nil {
		nodeResp.Actions = make([]common.Action, 0)
	}

	switch execResp.Status {
	case common.ExecComplete:
		nodeResp.Status = common.NodeStatusComplete
		nodeResp.Type = ""
	case common.ExecUserInputRequired:
		nodeResp.Status = common.NodeStatusIncomplete
		nodeResp.Type = common.NodeResponseTypeView
	case common.ExecExternalRedirection:
		nodeResp.Status = common.NodeStatusIncomplete
		nodeResp.Type = common.NodeResponseTypeRedirection
	case common.ExecRetry:
		nodeResp.Status = common.NodeStatusIncomplete
		nodeResp.Type = common.NodeResponseTypeRetry
	case common.ExecFailure:
		nodeResp.Status = common.NodeStatusFailure
		nodeResp.Type = ""
	default:
		nodeResp.Status = common.NodeStatusIncomplete
		nodeResp.Type = ""
	}

	return nodeResp
}

// GetExecutorName returns the executor name for the task execution node
func (n *taskExecutionNode) GetExecutorName() string {
	return n.executorName
}

// SetExecutorName sets the executor name for the task execution node
func (n *taskExecutionNode) SetExecutorName(name string) {
	n.executorName = name
}

// GetExecutor returns the executor instance associated with the task execution node
func (n *taskExecutionNode) GetExecutor() ExecutorInterface {
	return n.executor
}

// SetExecutor sets the executor instance for the task execution node
func (n *taskExecutionNode) SetExecutor(executor ExecutorInterface) {
	n.executor = executor
	if executor != nil {
		n.executorName = executor.GetName()
	}
}

// GetOnSuccess returns the onSuccess node ID
func (n *taskExecutionNode) GetOnSuccess() string {
	return n.onSuccess
}

// SetOnSuccess sets the onSuccess node ID
func (n *taskExecutionNode) SetOnSuccess(nodeID string) {
	n.onSuccess = nodeID
}

// GetOnFailure returns the onFailure node ID
func (n *taskExecutionNode) GetOnFailure() string {
	return n.onFailure
}

// SetOnFailure sets the onFailure node ID
func (n *taskExecutionNode) SetOnFailure(nodeID string) {
	n.onFailure = nodeID
}

// GetMode returns the mode for the executor that supports multi-step execution
func (n *taskExecutionNode) GetMode() string {
	return n.mode
}

// SetMode sets the mode for the executor that supports multi-step execution
func (n *taskExecutionNode) SetMode(mode string) {
	n.mode = mode
}
