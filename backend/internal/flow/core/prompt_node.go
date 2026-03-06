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

// PromptNodeInterface extends NodeInterface for nodes that require user interaction.
type PromptNodeInterface interface {
	NodeInterface
	GetPrompts() []common.Prompt
	SetPrompts(prompts []common.Prompt)
	GetMeta() interface{}
	SetMeta(meta interface{})
}

// promptNode represents a node that prompts for user input/ action in the flow execution.
type promptNode struct {
	*node
	prompts []common.Prompt
	meta    interface{}
	logger  *log.Logger
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
		},
		prompts: []common.Prompt{},
		logger: log.GetLogger().With(log.String(log.LoggerKeyComponentName, "PromptNode"),
			log.String(log.LoggerKeyNodeID, id)),
	}
}

// Execute executes the prompt node logic based on the current context.
func (n *promptNode) Execute(ctx *NodeContext) (*common.NodeResponse, *serviceerror.ServiceError) {
	logger := n.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
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

	if n.resolvePromptInputs(ctx, nodeResp) {
		logger.Debug("All required inputs and action are available, returning complete status")

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

	// Include meta in the response if verbose mode is enabled
	if ctx.Verbose && n.GetMeta() != nil {
		nodeResp.Meta = n.GetMeta()
	}

	nodeResp.Status = common.NodeStatusIncomplete
	nodeResp.Type = common.NodeResponseTypeView
	return nodeResp, nil
}

// GetPrompts returns the prompts for the prompt node
func (n *promptNode) GetPrompts() []common.Prompt {
	return n.prompts
}

// SetPrompts sets the prompts for the prompt node
func (n *promptNode) SetPrompts(prompts []common.Prompt) {
	n.prompts = prompts
}

// GetMeta returns the meta object for the prompt node
func (n *promptNode) GetMeta() interface{} {
	return n.meta
}

// SetMeta sets the meta object for the prompt node
func (n *promptNode) SetMeta(meta interface{}) {
	n.meta = meta
}

// resolvePromptInputs resolves the inputs and actions for the prompt node.
// It checks for missing required inputs, validates action selection, attempts auto-selection
// if applicable, and enriches inputs with dynamic data from ForwardedData.
// Returns true if all required inputs are available and a valid action is selected, otherwise false.
func (n *promptNode) resolvePromptInputs(ctx *NodeContext, nodeResp *common.NodeResponse) bool {
	// Check for required inputs and collect missing ones
	hasAllInputs := n.hasRequiredInputs(ctx, nodeResp)

	// Enrich inputs from ForwardedData
	n.enrichInputsFromForwardedData(ctx, nodeResp)

	// Check for action selection
	hasAction := n.hasSelectedAction(ctx, nodeResp)

	// If inputs are satisfied but no action selected, try to auto-select single action
	if hasAllInputs && !hasAction && n.tryAutoSelectSingleAction(ctx) {
		hasAction = true
		// Clear actions from response since we auto-selected
		nodeResp.Actions = make([]common.Action, 0)
	}

	return hasAllInputs && hasAction
}

// hasRequiredInputs checks if all required inputs are available in the context. Adds missing
// inputs to the node response. Returns true if all required inputs are available, otherwise false.
func (n *promptNode) hasRequiredInputs(ctx *NodeContext, nodeResp *common.NodeResponse) bool {
	logger := n.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))

	if nodeResp.Inputs == nil {
		nodeResp.Inputs = make([]common.Input, 0)
	}

	// Check if an action is selected
	if ctx.CurrentAction != "" {
		// If the selected action matches a prompt, validate inputs for that prompt only
		for _, prompt := range n.prompts {
			if prompt.Action != nil && prompt.Action.Ref == ctx.CurrentAction {
				return !n.appendMissingInputs(ctx, nodeResp, prompt.Inputs)
			}
		}
		logger.Debug("Selected action not found in prompts, treating as no action selected",
			log.String("action", ctx.CurrentAction))
	} else {
		logger.Debug("No action selected, checking inputs from all prompts")
	}

	// If no action selected or action not found, validate inputs from all prompts
	allInputs := make([]common.Input, 0)
	for _, prompt := range n.prompts {
		allInputs = append(allInputs, prompt.Inputs...)
	}

	return !n.appendMissingInputs(ctx, nodeResp, allInputs)
}

// appendMissingInputs appends the missing required inputs to the node response.
// Returns true if any required data is found missing, otherwise false.
func (n *promptNode) appendMissingInputs(ctx *NodeContext, nodeResp *common.NodeResponse,
	requiredInputs []common.Input) bool {
	logger := log.GetLogger().With(log.String(log.LoggerKeyFlowID, ctx.FlowID))

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

// enrichInputsFromForwardedData enriches the inputs in the node response with dynamic data
// from ForwardedData. Currently only enriches Options for inputs that match by Identifier.
func (n *promptNode) enrichInputsFromForwardedData(ctx *NodeContext, nodeResp *common.NodeResponse) {
	if ctx.ForwardedData == nil || len(nodeResp.Inputs) == 0 {
		return
	}

	// Check if ForwardedData contains inputs
	forwardedInputsData, ok := ctx.ForwardedData[common.ForwardedDataKeyInputs]
	if !ok {
		return
	}

	// Type assert to []common.Input
	forwardedInputs, ok := forwardedInputsData.([]common.Input)
	if !ok {
		n.logger.Debug("ForwardedData contains 'inputs' key but value is not []common.Input, skipping enrichment")
		return
	}

	// Build a map of forwarded inputs by Identifier for quick lookup
	forwardedInputMap := make(map[string]common.Input)
	for _, fwdInput := range forwardedInputs {
		forwardedInputMap[fwdInput.Identifier] = fwdInput
	}

	// Enrich each prompt input with data from matching forwarded input
	for i := range nodeResp.Inputs {
		if fwdInput, found := forwardedInputMap[nodeResp.Inputs[i].Identifier]; found {
			// Only enrich Options - do not overwrite other fields like Ref, Type, Required
			if len(fwdInput.Options) > 0 {
				nodeResp.Inputs[i].Options = fwdInput.Options
				n.logger.Debug("Enriched input with options from ForwardedData",
					log.String("identifier", nodeResp.Inputs[i].Identifier),
					log.Int("optionsCount", len(fwdInput.Options)))
			}
		}
	}
}

// hasSelectedAction checks if a valid action has been selected when actions are defined. Adds actions
// to the response if they haven't been selected yet.
// Returns true if an action is already selected or no actions are defined, otherwise false.
func (n *promptNode) hasSelectedAction(ctx *NodeContext, nodeResp *common.NodeResponse) bool {
	actions := n.getAllActions()
	if len(actions) == 0 {
		return true
	}

	// Check if a valid action is selected
	if ctx.CurrentAction != "" {
		for _, action := range actions {
			if action.Ref == ctx.CurrentAction {
				return true
			}
		}
	}

	// If no action selected or invalid action, add actions to response
	nodeResp.Actions = append(nodeResp.Actions, actions...)
	return false
}

// tryAutoSelectSingleAction attempts to auto-select the action when there's exactly one action
// defined, no action has been selected, and inputs are defined. If no inputs are defined
// (confirmation-only prompts), we should not auto-select as the prompt is meant to wait for
// explicit user action.
// Returns true if an action was auto-selected, otherwise false.
func (n *promptNode) tryAutoSelectSingleAction(ctx *NodeContext) bool {
	actions := n.getAllActions()
	allInputs := n.getAllInputs()

	// Auto-select only when: single action, no action selected, and has inputs defined
	// Skip auto-select for confirmation prompts (no inputs) - they should wait for explicit action
	if len(actions) == 1 && ctx.CurrentAction == "" && len(allInputs) > 0 {
		ctx.CurrentAction = actions[0].Ref
		n.logger.Debug("Auto-selected single action", log.String(log.LoggerKeyFlowID, ctx.FlowID),
			log.String("actionRef", actions[0].Ref))
		return true
	}
	return false
}

// getAllInputs returns all inputs from prompts.
func (n *promptNode) getAllInputs() []common.Input {
	inputs := make([]common.Input, 0)
	for _, prompt := range n.prompts {
		inputs = append(inputs, prompt.Inputs...)
	}
	return inputs
}

// getAllActions returns all actions from prompts.
func (n *promptNode) getAllActions() []common.Action {
	actions := make([]common.Action, 0)
	for _, prompt := range n.prompts {
		if prompt.Action != nil {
			actions = append(actions, *prompt.Action)
		}
	}
	return actions
}

// getNextNodeForActionRef finds the next node for the given action reference.
func (n *promptNode) getNextNodeForActionRef(actionRef string, logger *log.Logger) string {
	actions := n.getAllActions()
	for i := range actions {
		if actions[i].Ref == actionRef {
			logger.Debug("Action selected successfully", log.String("actionRef", actions[i].Ref),
				log.String("nextNode", actions[i].NextNode))
			return actions[i].NextNode
		}
	}
	return ""
}
