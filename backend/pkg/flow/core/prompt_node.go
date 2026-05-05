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
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/pkg/flow/common"
)

// PromptNodeInterface extends NodeInterface for nodes that require user interaction.
type PromptNodeInterface interface {
	NodeInterface
	GetPrompts() []common.Prompt
	SetPrompts(prompts []common.Prompt)
	GetMeta() interface{}
	SetMeta(meta interface{})
	GetNextNode() string
	SetNextNode(nextNode string)
	GetMessage() string
	SetMessage(message string)
	IsDisplayOnly() bool
}

// promptNode represents a node that prompts for user input/ action in the flow execution.
type promptNode struct {
	*node
	prompts  []common.Prompt
	meta     interface{}
	nextNode string
	message  string
	logger   *log.Logger
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
	logger := n.logger.With(log.String(log.LoggerKeyExecutionID, ctx.ExecutionID))
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
			// Clear this prompt's inputs and current action
			for _, input := range n.getAllInputs() {
				delete(ctx.UserInputs, input.Identifier)
			}
			ctx.CurrentAction = ""
		}
	}

	// Check if this is a display-only prompt node
	if n.IsDisplayOnly() {
		logger.Debug("Display-only prompt node, returning display content")

		if ctx.Verbose && n.GetMeta() != nil {
			nodeResp.Meta = n.GetMeta()
		}

		if n.message != "" {
			if nodeResp.AdditionalData == nil {
				nodeResp.AdditionalData = make(map[string]string)
			}
			nodeResp.AdditionalData[common.DataPromptMessage] = n.message
		}

		nodeResp.Status = common.NodeStatusComplete
		nodeResp.Type = common.NodeResponseTypeView
		return nodeResp, nil
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

		// Forward the action type to the next node
		if actionType := n.getActionTypeForRef(ctx.CurrentAction); actionType != "" {
			if nodeResp.ForwardedData == nil {
				nodeResp.ForwardedData = make(map[string]interface{})
			}
			nodeResp.ForwardedData[common.ForwardedDataKeyActionType] = actionType
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
		trimmed := n.trimMetaToRequestedInputs(nodeResp.Inputs, nodeResp.Actions)
		nodeResp.Meta = n.appendSyntheticMetaComponents(trimmed, nodeResp.Inputs)
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

// GetNextNode returns the next node ID for display-only prompt nodes.
func (n *promptNode) GetNextNode() string {
	return n.nextNode
}

// SetNextNode sets the next node ID for display-only prompt nodes.
func (n *promptNode) SetNextNode(nextNode string) {
	n.nextNode = nextNode
}

// GetMessage returns the display message for display-only prompt nodes.
func (n *promptNode) GetMessage() string {
	return n.message
}

// SetMessage sets the display message for display-only prompt nodes.
func (n *promptNode) SetMessage(message string) {
	n.message = message
}

// IsDisplayOnly returns true if this is a display-only prompt node.
// A prompt node is considered display-only if it has a next node, but no prompts (inputs or actions).
func (n *promptNode) IsDisplayOnly() bool {
	return n.nextNode != "" && len(n.prompts) == 0
}

// resolvePromptInputs resolves the inputs and actions for the prompt node.
// It checks for missing required inputs, validates action selection, attempts auto-selection
// if applicable, and enriches inputs with dynamic data from ForwardedData.
// Returns true if all required inputs are available and a valid action is selected, otherwise false.
func (n *promptNode) resolvePromptInputs(ctx *NodeContext, nodeResp *common.NodeResponse) bool {
	// Check for required inputs and collect missing ones
	hasAllInputs := n.hasRequiredInputs(ctx, nodeResp)

	// Enrich inputs from ForwardedData — may append dynamically derived inputs not in node prompts.
	// If any new inputs are added they are unsatisfied by definition, so the node is incomplete.
	prevCount := len(nodeResp.Inputs)
	n.enrichInputsFromForwardedData(ctx, nodeResp)
	if len(nodeResp.Inputs) > prevCount {
		hasAllInputs = false
	}

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
	logger := n.logger.With(log.String(log.LoggerKeyExecutionID, ctx.ExecutionID))

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
	return !n.appendMissingInputs(ctx, nodeResp, n.getAllInputs())
}

// appendMissingInputs appends the missing required inputs to the node response.
// Returns true if any required data is found missing, otherwise false.
func (n *promptNode) appendMissingInputs(ctx *NodeContext, nodeResp *common.NodeResponse,
	requiredInputs []common.Input) bool {
	logger := log.GetLogger().With(log.String(log.LoggerKeyExecutionID, ctx.ExecutionID))

	requireInputs := false
	for _, input := range requiredInputs {
		if _, ok := ctx.UserInputs[input.Identifier]; !ok {
			if _, ok := ctx.RuntimeData[input.Identifier]; ok {
				logger.Debug("Input available in runtime data, skipping",
					log.String("identifier", input.Identifier), log.Bool("isRequired", input.Required))
				continue
			}
			if value, ok := ctx.ForwardedData[input.Identifier]; ok {
				if _, isString := value.(string); isString {
					logger.Debug("Input available in forwarded data, skipping",
						log.String("identifier", input.Identifier), log.Bool("isRequired", input.Required))
					continue
				}
			}
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
// from ForwardedData. Inputs present in ForwardedData but absent from the node response are
// appended (dynamically derived inputs). Options are propagated for all matched inputs.
func (n *promptNode) enrichInputsFromForwardedData(ctx *NodeContext, nodeResp *common.NodeResponse) {
	if ctx.ForwardedData == nil {
		return
	}

	// Check if ForwardedData contains inputs.
	forwardedInputsData, ok := ctx.ForwardedData[common.ForwardedDataKeyInputs]
	if !ok {
		return
	}

	// Type assert to []common.Input.
	forwardedInputs, ok := forwardedInputsData.([]common.Input)
	if !ok {
		n.logger.Debug("ForwardedData contains 'inputs' key but value is not []common.Input, skipping enrichment")
		return
	}

	// Build an index map of identifiers already in the response for O(1) lookup and in-place update.
	existingIndexMap := make(map[string]int, len(nodeResp.Inputs))
	for i, inp := range nodeResp.Inputs {
		existingIndexMap[inp.Identifier] = i
	}

	// Single pass: upsert forwarded inputs — replace existing entries (updating required/options)
	// or append dynamically derived inputs not yet satisfied by the user.
	for _, fwdInput := range forwardedInputs {
		if idx, exists := existingIndexMap[fwdInput.Identifier]; exists {
			if fwdInput.Required && !nodeResp.Inputs[idx].Required {
				nodeResp.Inputs[idx].Required = true
				n.logger.Debug("Updated input required flag from ForwardedData",
					log.String("identifier", fwdInput.Identifier))
			}
			if fwdInput.Type == common.InputTypeSelect &&
				nodeResp.Inputs[idx].Type == common.InputTypeSelect &&
				len(fwdInput.Options) > 0 {
				nodeResp.Inputs[idx].Options = fwdInput.Options
				n.logger.Debug("Enriched input with options from ForwardedData",
					log.String("identifier", fwdInput.Identifier),
					log.Int("optionsCount", len(fwdInput.Options)))
			}
			continue
		}
		if _, ok := ctx.UserInputs[fwdInput.Identifier]; ok {
			continue
		}
		if _, ok := ctx.RuntimeData[fwdInput.Identifier]; ok {
			continue
		}
		if value, ok := ctx.ForwardedData[fwdInput.Identifier]; ok {
			if _, isString := value.(string); isString {
				continue
			}
		}
		nodeResp.Inputs = append(nodeResp.Inputs, fwdInput)
		existingIndexMap[fwdInput.Identifier] = len(nodeResp.Inputs) - 1
		n.logger.Debug("Added dynamically-derived input from ForwardedData",
			log.String("identifier", fwdInput.Identifier))
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
		n.logger.Debug("Auto-selected single action", log.String(log.LoggerKeyExecutionID, ctx.ExecutionID),
			log.String("actionRef", actions[0].Ref))
		return true
	}
	return false
}

// getAllInputs returns all unique inputs from prompts, deduplicated by Identifier.
func (n *promptNode) getAllInputs() []common.Input {
	seen := make(map[string]struct{})
	inputs := make([]common.Input, 0)
	for _, prompt := range n.prompts {
		for _, input := range prompt.Inputs {
			if _, exists := seen[input.Identifier]; !exists {
				seen[input.Identifier] = struct{}{}
				inputs = append(inputs, input)
			}
		}
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

// getActionTypeForRef finds the action type for the given action reference.
func (n *promptNode) getActionTypeForRef(actionRef string) string {
	for _, prompt := range n.prompts {
		if prompt.Action != nil && prompt.Action.Ref == actionRef {
			return prompt.Action.Type
		}
	}
	return ""
}

// trimMetaToRequestedInputs returns a copy of n.meta with the "components" list trimmed to only
// include components matching the given inputs and actions (plus structural components like TEXT
// and BLOCK containers that are not themselves inputs or actions).
func (n *promptNode) trimMetaToRequestedInputs(inputs []common.Input, actions []common.Action) interface{} {
	metaMap, ok := n.meta.(map[string]interface{})
	if !ok {
		return n.meta
	}

	allowedRefs := make(map[string]struct{})
	for _, input := range inputs {
		if input.Ref != "" {
			allowedRefs[input.Ref] = struct{}{}
		}
	}
	for _, action := range actions {
		if action.Ref != "" {
			allowedRefs[action.Ref] = struct{}{}
		}
	}

	knownInputActionRefs := make(map[string]struct{})
	for _, input := range n.getAllInputs() {
		if input.Ref != "" {
			knownInputActionRefs[input.Ref] = struct{}{}
		}
	}
	for _, action := range n.getAllActions() {
		if action.Ref != "" {
			knownInputActionRefs[action.Ref] = struct{}{}
		}
	}

	trimmed := make(map[string]interface{}, len(metaMap))
	for k, v := range metaMap {
		trimmed[k] = v
	}
	if comps, ok := metaMap["components"]; ok {
		if compSlice, ok := comps.([]interface{}); ok {
			trimmed["components"] = filterMetaComponents(compSlice, allowedRefs, knownInputActionRefs)
		}
	}
	return trimmed
}

// filterMetaComponents filters a meta components slice, dropping satisfied input/action components
// while keeping structural components (TEXT, BLOCK containers, etc.) and recursively trimming
// their children.
func filterMetaComponents(comps []interface{}, allowedRefs, knownInputActionRefs map[string]struct{}) []interface{} {
	result := make([]interface{}, 0, len(comps))
	for _, comp := range comps {
		compMap, ok := comp.(map[string]interface{})
		if !ok {
			result = append(result, comp)
			continue
		}

		id, _ := compMap["id"].(string)
		if _, isKnown := knownInputActionRefs[id]; isKnown {
			if _, isAllowed := allowedRefs[id]; isAllowed {
				result = append(result, comp)
			}
			continue
		}

		// Structural component — always keep; recurse into children if present.
		if childComps, hasChildren := compMap["components"]; hasChildren {
			if childSlice, ok := childComps.([]interface{}); ok {
				trimmedComp := make(map[string]interface{}, len(compMap))
				for k, v := range compMap {
					trimmedComp[k] = v
				}
				trimmedComp["components"] = filterMetaComponents(childSlice, allowedRefs, knownInputActionRefs)
				result = append(result, trimmedComp)
				continue
			}
		}
		result = append(result, comp)
	}
	return result
}

// appendSyntheticMetaComponents ensures every input in the list has a corresponding meta
// component. For inputs whose component already exists (matched by ref or id), the required
// field is updated in-place if the input marks it required. For inputs with no existing
// component, a minimal synthetic component is created and inserted into the first BLOCK
// before any ACTION. If no BLOCK exists, a new one is appended.
// The label uses DisplayName when set, falling back to Identifier.
func (n *promptNode) appendSyntheticMetaComponents(trimmedMeta interface{}, inputs []common.Input) interface{} {
	metaMap, ok := trimmedMeta.(map[string]interface{})
	if !ok {
		return trimmedMeta
	}

	// Build a set of refs/ids from the node's own configured prompt inputs —
	// used to suppress synthesis for node-defined inputs with no meta component.
	nodeInputRefs := make(map[string]struct{})
	for _, inp := range n.getAllInputs() {
		if inp.Ref != "" {
			nodeInputRefs[inp.Ref] = struct{}{}
		}
		nodeInputRefs[inp.Identifier] = struct{}{}
	}

	// Build ref/id → component map for O(1) lookup and in-place required update.
	metaCompByRef := make(map[string]map[string]interface{})
	collectMetaComponentMap(metaMap["components"], metaCompByRef)

	// Single pass: update required on existing components; collect synthetic for missing ones.
	// For each input, try to find its meta component by Identifier first, then by Ref.
	// Node-configured inputs with no meta component are skipped (no synthesis).
	synthetic := make([]interface{}, 0, len(inputs))
	for _, input := range inputs {
		comp, inMeta := metaCompByRef[input.Identifier]
		if !inMeta && input.Ref != "" {
			comp, inMeta = metaCompByRef[input.Ref]
		}
		if inMeta {
			if input.Required {
				comp["required"] = true
			}
			continue
		}
		if _, isNodeInput := nodeInputRefs[input.Identifier]; isNodeInput {
			continue
		}
		label := input.DisplayName
		if label == "" {
			label = input.Identifier
		}
		inputType := input.Type
		if inputType == "" {
			inputType = common.InputTypeText
		}
		synthetic = append(synthetic, map[string]interface{}{
			"id":       input.Identifier,
			"ref":      input.Identifier,
			"type":     inputType,
			"label":    label,
			"required": input.Required,
		})
	}

	// Walk the meta tree and either replace a DYNAMIC_INPUT_PLACEHOLDER with synthetic
	// inputs (preferred — exact insertion point), or insert before the first ACTION in
	// the first BLOCK (fallback). The placeholder is always stripped even when there are
	// no synthetic inputs, so it never leaks to the client.
	// Input components are only rendered inside a BLOCK by the UI renderer.
	result := make(map[string]interface{}, len(metaMap))
	for k, v := range metaMap {
		result[k] = v
	}
	existing, _ := metaMap["components"].([]interface{})
	updatedComponents := make([]interface{}, len(existing))
	copy(updatedComponents, existing)

	placeholderStripped := false
	inserted := false
	for i, comp := range updatedComponents {
		compMap, ok := comp.(map[string]interface{})
		if !ok || compMap["type"] != common.MetaComponentTypeBlock {
			continue
		}
		children, _ := compMap["components"].([]interface{})

		// Preferred path: find and replace DYNAMIC_INPUT_PLACEHOLDER.
		// Always strip it, inserting synthetic inputs in its place (may be empty).
		for j, child := range children {
			childMap, ok := child.(map[string]interface{})
			if !ok || childMap["type"] != common.MetaComponentTypeDynamicInputPlaceholder {
				continue
			}
			newChildren := make([]interface{}, 0, len(children)-1+len(synthetic))
			newChildren = append(newChildren, children[:j]...)
			newChildren = append(newChildren, synthetic...)
			newChildren = append(newChildren, children[j+1:]...)
			newBlock := make(map[string]interface{}, len(compMap))
			for k, v := range compMap {
				newBlock[k] = v
			}
			newBlock["components"] = newChildren
			updatedComponents[i] = newBlock
			placeholderStripped = true
			inserted = true
			break
		}
		if placeholderStripped {
			break
		}

		// Fallback (no placeholder): insert before the first ACTION, only when there
		// are synthetic inputs to add.
		if len(synthetic) == 0 {
			break
		}
		insertIdx := len(children)
		for j, child := range children {
			childMap, ok := child.(map[string]interface{})
			if ok && childMap["type"] == common.MetaComponentTypeAction {
				insertIdx = j
				break
			}
		}
		newChildren := make([]interface{}, 0, len(children)+len(synthetic))
		newChildren = append(newChildren, children[:insertIdx]...)
		newChildren = append(newChildren, synthetic...)
		newChildren = append(newChildren, children[insertIdx:]...)
		newBlock := make(map[string]interface{}, len(compMap))
		for k, v := range compMap {
			newBlock[k] = v
		}
		newBlock["components"] = newChildren
		updatedComponents[i] = newBlock
		inserted = true
		break
	}

	if len(synthetic) == 0 && !placeholderStripped {
		return trimmedMeta
	}

	if !inserted && len(synthetic) > 0 {
		// No BLOCK found — wrap synthetic inputs in a new BLOCK so the UI renderer
		// can display them (input components are only supported inside a BLOCK).
		updatedComponents = append(updatedComponents, map[string]interface{}{
			"id":         "block_schema_dynamic",
			"type":       common.MetaComponentTypeBlock,
			"components": synthetic,
		})
	}

	result["components"] = updatedComponents
	return result
}

// collectMetaComponentMap recursively walks the meta components tree and builds a
// ref/id → component map. ref takes priority over id for the same component so that
// identifier-based lookups match the data-binding key rather than the element id.
func collectMetaComponentMap(comps interface{}, compMap map[string]map[string]interface{}) {
	compSlice, ok := comps.([]interface{})
	if !ok {
		return
	}
	for _, comp := range compSlice {
		cm, ok := comp.(map[string]interface{})
		if !ok {
			continue
		}
		if ref, ok := cm["ref"].(string); ok && ref != "" {
			compMap[ref] = cm
		}
		if id, ok := cm["id"].(string); ok && id != "" {
			if _, exists := compMap[id]; !exists {
				compMap[id] = cm
			}
		}
		collectMetaComponentMap(cm["components"], compMap)
	}
}
