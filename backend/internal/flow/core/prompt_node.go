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

const promptNodeLoggerComponentName = "PromptNode"

// promptNode represents a node that prompts for user input/ action in the flow execution.
type promptNode struct {
	*node
	actions []common.Action
}

// newPromptNode creates a new instance of PromptNode with the given details.
func newPromptNode(id string, properties map[string]interface{},
	isStartNode bool, isFinalNode bool) NodeInterface {
	return &promptNode{
		node: &node{
			id:               id,
			_type:            common.NodeTypePrompt,
			properties:       properties,
			isStartNode:      isStartNode,
			isFinalNode:      isFinalNode,
			nextNodeList:     []string{},
			previousNodeList: []string{},
			inputs:           []common.Input{},
		},
		actions: []common.Action{},
	}
}

// Execute executes the prompt node logic based on the current context.
func (n *promptNode) Execute(ctx *NodeContext) (*common.NodeResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, promptNodeLoggerComponentName),
		log.String(log.LoggerKeyNodeID, n.GetID()),
		log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Executing prompt node")

	nodeResp := &common.NodeResponse{
		Inputs:         make([]common.Input, 0),
		AdditionalData: make(map[string]string),
		Actions:        make([]common.Action, 0),
		RuntimeData:    make(map[string]string),
	}

	// Check if this prompt is handling a failure
	if ctx.RuntimeData != nil {
		if failureReason, exists := ctx.RuntimeData["failureReason"]; exists && failureReason != "" {
			logger.Debug("Prompt node is handling a failure", log.String("failureReason", failureReason))
			nodeResp.FailureReason = failureReason
			delete(ctx.RuntimeData, "failureReason")
		}
	}

	hasAllInputs := n.hasRequiredInputs(ctx, nodeResp)
	hasAction := n.hasSelectedAction(ctx, nodeResp)

	// If both inputs and action are satisfied, complete the node
	if hasAllInputs && hasAction {
		logger.Debug("All required inputs and action are available, returning complete status")

		// If an action was selected, set the next node
		if ctx.CurrentAction != "" {
			if nextNode := n.getNextNodeForActionRef(ctx.CurrentAction, logger); nextNode != "" {
				nodeResp.NextNodeID = nextNode
			} else {
				logger.Debug("Invalid action selected", log.String("actionRef", ctx.CurrentAction))
				nodeResp.Status = common.NodeStatusFailure
				nodeResp.FailureReason = "Invalid action selected"
				return nodeResp, nil
			}
		}

		nodeResp.Status = common.NodeStatusComplete
		nodeResp.Type = ""
		return nodeResp, nil
	}

	// If required inputs or action is not yet available, prompt for user interaction
	logger.Debug("Required inputs or action not available, prompting user",
		log.Any("inputs", nodeResp.Inputs), log.Any("actions", nodeResp.Actions))

	nodeResp.Status = common.NodeStatusIncomplete
	nodeResp.Type = common.NodeResponseTypeView
	return nodeResp, nil
}

// GetActions returns the actions available for the prompt node
func (n *promptNode) GetActions() []common.Action {
	return n.actions
}

// SetActions sets the actions available for the prompt node
func (n *promptNode) SetActions(actions []common.Action) {
	n.actions = actions
}

// hasRequiredInputs checks if all required inputs are available in the context. Adds missing
// inputs to the node response.
// Returns true if all required inputs are available, otherwise false.
func (n *promptNode) hasRequiredInputs(ctx *NodeContext, nodeResp *common.NodeResponse) bool {
	requiredInputs := n.GetInputs()
	if len(requiredInputs) == 0 {
		return true
	}

	if nodeResp.Inputs == nil {
		nodeResp.Inputs = make([]common.Input, 0)
	}

	if len(ctx.UserInputs) == 0 {
		nodeResp.Inputs = append(nodeResp.Inputs, requiredInputs...)
		return false
	}

	return !n.appendMissingInputs(ctx, nodeResp, requiredInputs)
}

// appendMissingInputs appends the missing required inputs to the node response.
// Returns true if any required data is found missing, otherwise false.
func (n *promptNode) appendMissingInputs(ctx *NodeContext, nodeResp *common.NodeResponse,
	requiredInputs []common.Input) bool {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, promptNodeLoggerComponentName),
		log.String(log.LoggerKeyNodeID, n.GetID()),
		log.String(log.LoggerKeyFlowID, ctx.FlowID))

	requireInputs := false
	for _, input := range requiredInputs {
		if _, ok := ctx.UserInputs[input.Identifier]; !ok {
			if input.Required {
				requireInputs = true
			}
			nodeResp.Inputs = append(nodeResp.Inputs, input)
			logger.Debug("Input not available in the context",
				log.String("identifier", input.Identifier), log.Bool("isRequired", input.Required))
		}
	}

	return requireInputs
}

// hasSelectedAction checks if an action has been selected when actions are defined. Adds actions
// to the response if they haven't been selected yet.
// Returns true if an action is already selected or no actions are defined, otherwise false.
func (n *promptNode) hasSelectedAction(ctx *NodeContext, nodeResp *common.NodeResponse) bool {
	actions := n.GetActions()
	if len(actions) == 0 {
		return true
	}

	// Returns true if an action is already selected
	if ctx.CurrentAction != "" {
		return true
	}

	// If not yet selected, add actions to response
	nodeResp.Actions = append(nodeResp.Actions, actions...)
	return false
}

// getNextNodeForActionRef finds the next node for the given action reference
func (n *promptNode) getNextNodeForActionRef(actionRef string, logger *log.Logger) string {
	actions := n.GetActions()
	for i := range actions {
		if actions[i].Ref == actionRef {
			logger.Debug("Action selected successfully", log.String("actionRef", actions[i].Ref),
				log.String("nextNode", actions[i].NextNode))
			return actions[i].NextNode
		}
	}
	return ""
}
