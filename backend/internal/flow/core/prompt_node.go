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
	"strings"

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
	GetVariant() string
	SetVariant(variant string)
}

// promptNode represents a node that prompts for user input/ action in the flow execution.
type promptNode struct {
	*node
	prompts []common.Prompt
	meta    interface{}
	variant string
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
			// Clear this prompt's inputs and current action (uses unfiltered prompts)
			for _, input := range getAllInputsFrom(n.prompts) {
				delete(ctx.UserInputs, input.Identifier)
			}
			ctx.CurrentAction = ""
		}
	}

	// For login_options nodes, compute the filtered+ordered prompt set as a local variable.
	// This avoids mutating the shared n.prompts field, which would be a race condition when
	// multiple goroutines execute the same cached graph node concurrently.
	effectivePrompts := n.prompts
	if n.variant == common.NodeVariantLoginOptions {
		requestedACRs := parseACRValues(ctx.RuntimeData[common.RuntimeKeyRequestedAuthClasses])
		effectivePrompts = n.filterAndOrderPrompts(requestedACRs)
	}

	if n.resolvePromptInputs(ctx, nodeResp, effectivePrompts) {
		logger.Debug("All required inputs and action are available, returning complete status")

		if ctx.CurrentAction != "" {
			if nextNode := getNextNodeForActionRef(effectivePrompts, ctx.CurrentAction, logger); nextNode != "" {
				nodeResp.NextNodeID = nextNode
				// Record the completed ACR for ID token generation.
				// Search unfiltered n.prompts so the lookup works for any valid action.
				if n.variant == common.NodeVariantLoginOptions {
					for _, p := range n.prompts {
						if p.Action != nil && p.Action.Ref == ctx.CurrentAction && p.ACR != "" {
							nodeResp.RuntimeData[common.RuntimeKeySelectedAuthClass] = p.ACR
							break
						}
					}
				}
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
		if n.variant == common.NodeVariantLoginOptions {
			nodeResp.Meta = n.filteredMeta(effectivePrompts)
		} else {
			nodeResp.Meta = n.GetMeta()
		}
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

// GetVariant returns the variant of the prompt node
func (n *promptNode) GetVariant() string {
	return n.variant
}

// SetVariant sets the variant of the prompt node
func (n *promptNode) SetVariant(variant string) {
	n.variant = variant
}

// resolvePromptInputs resolves the inputs and actions for the prompt node.
// It checks for missing required inputs, validates action selection, attempts auto-selection
// if applicable, and enriches inputs with dynamic data from ForwardedData.
// Returns true if all required inputs are available and a valid action is selected, otherwise false.
func (n *promptNode) resolvePromptInputs(ctx *NodeContext, nodeResp *common.NodeResponse,
	prompts []common.Prompt) bool {
	// Check for required inputs and collect missing ones
	hasAllInputs := n.hasRequiredInputs(ctx, nodeResp, prompts)

	// Enrich inputs from ForwardedData
	n.enrichInputsFromForwardedData(ctx, nodeResp)

	// Check for action selection
	hasAction := hasSelectedAction(ctx, nodeResp, prompts)

	// If inputs are satisfied but no action selected, try to auto-select single action
	if hasAllInputs && !hasAction && n.tryAutoSelectSingleAction(ctx, prompts) {
		hasAction = true
		// Clear actions from response since we auto-selected
		nodeResp.Actions = make([]common.Action, 0)
	}

	return hasAllInputs && hasAction
}

// hasRequiredInputs checks if all required inputs are available in the context. Adds missing
// inputs to the node response. Returns true if all required inputs are available, otherwise false.
func (n *promptNode) hasRequiredInputs(ctx *NodeContext, nodeResp *common.NodeResponse,
	prompts []common.Prompt) bool {
	logger := n.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))

	if nodeResp.Inputs == nil {
		nodeResp.Inputs = make([]common.Input, 0)
	}

	// Check if an action is selected
	if ctx.CurrentAction != "" {
		// If the selected action matches a prompt, validate inputs for that prompt only
		for _, prompt := range prompts {
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
	return !n.appendMissingInputs(ctx, nodeResp, getAllInputsFrom(prompts))
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
func hasSelectedAction(ctx *NodeContext, nodeResp *common.NodeResponse, prompts []common.Prompt) bool {
	actions := getAllActionsFrom(prompts)
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
// defined, no action has been selected, and either inputs are defined or the node is a
// login_options ACR chooser with an active acr_values filter. The ACR case implements AC-13:
// when acr_values narrows the chooser down to exactly one option, that option is selected
// automatically so the user goes directly to the credential prompt. For all other nodes,
// auto-select is skipped when no inputs are defined (confirmation-only prompts must wait for an
// explicit user action).
// Returns true if an action was auto-selected, otherwise false.
func (n *promptNode) tryAutoSelectSingleAction(ctx *NodeContext, prompts []common.Prompt) bool {
	actions := getAllActionsFrom(prompts)
	allInputs := getAllInputsFrom(prompts)

	// ACR auto-select (AC-13): login_options node with an active acr_values filter
	isACRAutoSelect := n.variant == common.NodeVariantLoginOptions &&
		ctx.RuntimeData[common.RuntimeKeyRequestedAuthClasses] != ""

	// Auto-select when: single action, no action selected, and has inputs defined OR ACR auto-select applies
	if len(actions) == 1 && ctx.CurrentAction == "" && (len(allInputs) > 0 || isACRAutoSelect) {
		ctx.CurrentAction = actions[0].Ref
		n.logger.Debug("Auto-selected single action", log.String(log.LoggerKeyFlowID, ctx.FlowID),
			log.String("actionRef", actions[0].Ref))
		return true
	}
	return false
}

// getAllInputsFrom returns all unique inputs from the given prompts, deduplicated by Identifier.
func getAllInputsFrom(prompts []common.Prompt) []common.Input {
	seen := make(map[string]struct{})
	inputs := make([]common.Input, 0)
	for _, prompt := range prompts {
		for _, input := range prompt.Inputs {
			if _, exists := seen[input.Identifier]; !exists {
				seen[input.Identifier] = struct{}{}
				inputs = append(inputs, input)
			}
		}
	}

	return inputs
}

// getAllActionsFrom returns all actions from the given prompts.
func getAllActionsFrom(prompts []common.Prompt) []common.Action {
	actions := make([]common.Action, 0)
	for _, prompt := range prompts {
		if prompt.Action != nil {
			actions = append(actions, *prompt.Action)
		}
	}
	return actions
}

// getNextNodeForActionRef finds the next node for the given action reference.
func getNextNodeForActionRef(prompts []common.Prompt, actionRef string, logger *log.Logger) string {
	actions := getAllActionsFrom(prompts)
	for i := range actions {
		if actions[i].Ref == actionRef {
			logger.Debug("Action selected successfully", log.String("actionRef", actions[i].Ref),
				log.String("nextNode", actions[i].NextNode))
			return actions[i].NextNode
		}
	}
	return ""
}

// filterAndOrderPrompts returns a filtered and preference-ordered subset of n.prompts for a
// login_options node. Prompts whose ACR is in requestedACRs are included in preference order;
// prompts without an ACR tag are always appended. If the filtered result is empty (flow
// misconfiguration), all prompts are returned as a graceful fallback.
func (n *promptNode) filterAndOrderPrompts(requestedACRs []string) []common.Prompt {
	if len(requestedACRs) == 0 {
		return n.prompts
	}

	acrToPrompt := make(map[string]common.Prompt)
	for _, p := range n.prompts {
		if p.ACR != "" {
			acrToPrompt[p.ACR] = p
		}
	}

	result := make([]common.Prompt, 0)
	for _, acr := range requestedACRs {
		if p, ok := acrToPrompt[acr]; ok {
			result = append(result, p)
		}
	}

	// Prompts without an ACR tag are not ACR-gated and are always included.
	for _, p := range n.prompts {
		if p.ACR == "" {
			result = append(result, p)
		}
	}

	if len(result) == 0 {
		return n.prompts // graceful fallback for misconfigured flows
	}
	return result
}

// filteredMeta returns a copy of n.meta with ACTION components filtered and reordered to match
// the given prompts. Non-ACTION components (display text, images, etc.) are always kept in their
// original relative positions. ACTION components are replaced in-place with the next action in
// ACR preference order — filtered-out actions are dropped. If meta cannot be interpreted as
// expected, the original meta is returned unchanged.
func (n *promptNode) filteredMeta(prompts []common.Prompt) interface{} {
	metaMap, ok := n.meta.(map[string]interface{})
	if !ok {
		return n.meta
	}
	components, ok := metaMap["components"].([]interface{})
	if !ok {
		return n.meta
	}

	// Build map from action ref to ACTION component.
	actionCompMap := make(map[string]interface{}, len(prompts))
	for _, comp := range components {
		compMap, ok := comp.(map[string]interface{})
		if !ok {
			continue
		}
		if compMap["type"] == "ACTION" {
			if id, ok := compMap["id"].(string); ok {
				actionCompMap[id] = comp
			}
		}
	}

	// Build ordered slice of ACTION components following the ACR-preference order of prompts.
	orderedActions := make([]interface{}, 0, len(prompts))
	for _, p := range prompts {
		if p.Action != nil && p.Action.Ref != "" {
			if comp, ok := actionCompMap[p.Action.Ref]; ok {
				orderedActions = append(orderedActions, comp)
			}
		}
	}

	// Reconstruct the components list: non-ACTION components stay in their original positions;
	// each ACTION slot is replaced by the next entry from orderedActions (filtered-out actions
	// are simply dropped by consuming no slot in orderedActions).
	actionIdx := 0
	result := make([]interface{}, 0, len(components))
	for _, comp := range components {
		compMap, ok := comp.(map[string]interface{})
		if !ok || compMap["type"] != "ACTION" {
			result = append(result, comp)
			continue
		}
		if actionIdx < len(orderedActions) {
			result = append(result, orderedActions[actionIdx])
			actionIdx++
		}
		// Filtered-out ACTION component: drop it by emitting nothing.
	}

	resultMap := make(map[string]interface{}, len(metaMap))
	for k, v := range metaMap {
		resultMap[k] = v
	}
	resultMap["components"] = result
	return resultMap
}

// parseACRValues splits the space-separated ACR values string from RuntimeData into an ordered slice.
func parseACRValues(raw string) []string {
	if raw == "" {
		return nil
	}
	return strings.Fields(raw)
}
