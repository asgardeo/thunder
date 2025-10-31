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
	"github.com/asgardeo/thunder/internal/flow/common/constants"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
)

// TaskExecutionNode represents a node that executes a task
type TaskExecutionNode struct {
	*Node
}

// NewTaskExecutionNode creates a new TaskExecutionNode with the given details.
func NewTaskExecutionNode(id string, isStartNode bool, isFinalNode bool) NodeInterface {
	return &TaskExecutionNode{
		Node: &Node{
			id:               id,
			_type:            constants.NodeTypeTaskExecution,
			isStartNode:      isStartNode,
			isFinalNode:      isFinalNode,
			nextNodeList:     []string{},
			previousNodeList: []string{},
			inputData:        []InputData{},
			executorConfig:   &ExecutorConfig{},
		},
	}
}

// Execute executes the node's executor.
func (n *TaskExecutionNode) Execute(ctx *NodeContext) (*NodeResponse, *serviceerror.ServiceError) {
	if n.executorConfig == nil || n.executorConfig.Executor == nil {
		return nil, &constants.ErrorNodeExecutorNotFound
	}

	execResp, svcErr := n.triggerExecutor(ctx)
	if svcErr != nil {
		return nil, svcErr
	}

	return buildNodeResponse(execResp), nil
}

// triggerExecutor triggers the executor configured for the node.
func (n *TaskExecutionNode) triggerExecutor(ctx *NodeContext) (*ExecutorResponse, *serviceerror.ServiceError) {
	execResp, err := n.executorConfig.Executor.Execute(ctx)
	if err != nil {
		svcErr := constants.ErrorNodeExecutorExecError
		svcErr.ErrorDescription = "Error executing node executor: " + err.Error()
		return nil, &svcErr
	}
	if execResp == nil {
		return nil, &constants.ErrorNilResponseFromExecutor
	}

	return execResp, nil
}

// buildNodeResponse constructs a NodeResponse from the ExecutorResponse.
func buildNodeResponse(execResp *ExecutorResponse) *NodeResponse {
	nodeResp := &NodeResponse{
		FailureReason:     execResp.FailureReason,
		RequiredData:      execResp.RequiredData,
		AdditionalData:    execResp.AdditionalData,
		RedirectURL:       execResp.RedirectURL,
		RuntimeData:       execResp.RuntimeData,
		AuthenticatedUser: execResp.AuthenticatedUser,
		Assertion:         execResp.Assertion,
		ExecutionRecord:   execResp.ExecutionRecord,
	}
	if nodeResp.AdditionalData == nil {
		nodeResp.AdditionalData = make(map[string]string)
	}
	if nodeResp.RuntimeData == nil {
		nodeResp.RuntimeData = make(map[string]string)
	}
	if nodeResp.RequiredData == nil {
		nodeResp.RequiredData = make([]InputData, 0)
	}
	if nodeResp.Actions == nil {
		nodeResp.Actions = make([]Action, 0)
	}

	if execResp.Status == constants.ExecComplete {
		nodeResp.Status = constants.NodeStatusComplete
		nodeResp.Type = ""
	} else if execResp.Status == constants.ExecUserInputRequired {
		nodeResp.Status = constants.NodeStatusIncomplete
		nodeResp.Type = constants.NodeResponseTypeView
	} else if execResp.Status == constants.ExecExternalRedirection {
		nodeResp.Status = constants.NodeStatusIncomplete
		nodeResp.Type = constants.NodeResponseTypeRedirection
	} else if execResp.Status == constants.ExecRetry {
		nodeResp.Status = constants.NodeStatusIncomplete
		nodeResp.Type = constants.NodeResponseTypeRetry
	} else if execResp.Status == constants.ExecFailure {
		nodeResp.Status = constants.NodeStatusFailure
		nodeResp.Type = ""
	} else {
		nodeResp.Status = constants.NodeStatusIncomplete
		nodeResp.Type = ""
	}

	return nodeResp
}
