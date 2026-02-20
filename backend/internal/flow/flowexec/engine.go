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

package flowexec

import (
	"errors"
	"fmt"
	"time"

	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/flow/executor"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/observability"
	"github.com/asgardeo/thunder/internal/system/observability/event"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

// flowEngineInterface defines the interface for the flow engine.
type flowEngineInterface interface {
	Execute(ctx *EngineContext) (FlowStep, *serviceerror.ServiceError)
}

// FlowEngine is the main engine implementation for orchestrating flow executions.
type flowEngine struct {
	executorRegistry executor.ExecutorRegistryInterface
	observabilitySvc observability.ObservabilityServiceInterface
}

// newFlowEngine creates a new flow engine with the given dependencies.
func newFlowEngine(
	executorRegistry executor.ExecutorRegistryInterface,
	observabilitySvc observability.ObservabilityServiceInterface,
) flowEngineInterface {
	return &flowEngine{
		executorRegistry: executorRegistry,
		observabilitySvc: observabilitySvc,
	}
}

// Execute executes a step in the flow
func (fe *flowEngine) Execute(ctx *EngineContext) (FlowStep, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "FlowEngine"))

	flowStep := FlowStep{
		FlowID: ctx.FlowID,
	}

	// Track flow execution start time
	flowStartTime := time.Now().UnixMilli()

	// Publish flow started event (only if this is the first execution - check if ExecutionHistory is empty)
	if len(ctx.ExecutionHistory) == 0 {
		publishFlowStartedEvent(ctx, fe.observabilitySvc)
	}

	currentNode, err := fe.setCurrentExecutionNode(ctx, logger)
	if err != nil {
		// Publish flow failed event before returning error
		publishFlowFailedEvent(ctx, err, flowStartTime, time.Now().UnixMilli(), fe.observabilitySvc)
		return flowStep, err
	}

	// Execute the graph nodes until a terminal condition is met or currentNode is nil
	for currentNode != nil {
		logger.Debug("Executing node", log.String("nodeID", currentNode.GetID()),
			log.String("nodeType", string(currentNode.GetType())))

		nodeCtx := &core.NodeContext{
			Context:           ctx.Context,
			FlowID:            ctx.FlowID,
			FlowType:          ctx.FlowType,
			AppID:             ctx.AppID,
			CurrentAction:     ctx.CurrentAction,
			Verbose:           ctx.Verbose,
			NodeInputs:        getNodeInputs(ctx.CurrentNode),
			UserInputs:        ctx.UserInputs,
			CurrentNodeID:     ctx.CurrentNode.GetID(),
			RuntimeData:       ctx.RuntimeData,
			ForwardedData:     ctx.ForwardedData,
			HTTPContext:       ctx.HTTPContext,
			Application:       ctx.Application,
			AuthenticatedUser: ctx.AuthenticatedUser,
			ExecutionHistory:  ctx.ExecutionHistory,
		}
		if nodeCtx.NodeInputs == nil {
			nodeCtx.NodeInputs = make([]common.Input, 0)
		}
		if nodeCtx.UserInputs == nil {
			nodeCtx.UserInputs = make(map[string]string)
		}
		if nodeCtx.RuntimeData == nil {
			nodeCtx.RuntimeData = make(map[string]string)
		}
		if nodeCtx.ForwardedData == nil {
			nodeCtx.ForwardedData = make(map[string]interface{})
		}

		// Clear ForwardedData from engine context after passing to node context
		// This ensures ForwardedData is only available to the immediate next node
		ctx.ForwardedData = nil

		// Check if the node should be executed based on its condition
		if !currentNode.ShouldExecute(nodeCtx) {
			logger.Debug("Skipping node due to unmet condition", log.String("nodeID", currentNode.GetID()))
			nextNode, svcErr := fe.skipToNextNode(ctx, currentNode, logger)
			if svcErr != nil {
				return flowStep, svcErr
			}
			currentNode = nextNode
			continue
		}

		svcErr := fe.setNodeExecutor(currentNode, logger)
		if svcErr != nil {
			return flowStep, svcErr
		}

		executionStartTime := time.Now().UnixMilli()

		// Publish node execution started event
		publishNodeExecutionStartedEvent(ctx, currentNode, fe.observabilitySvc)

		nodeResp, nodeErr := currentNode.Execute(nodeCtx)
		executionEndTime := time.Now().UnixMilli()

		// Clear sensitive inputs from context after executor has consumed them.
		fe.clearSensitiveInputs(ctx, currentNode)

		recordNodeExecution(ctx, currentNode, nodeResp, nodeErr, executionStartTime, executionEndTime)

		// Publish node execution completed or failed event
		publishNodeExecutionCompletedEvent(
			ctx, currentNode, nodeResp, nodeErr,
			executionStartTime, executionEndTime, fe.observabilitySvc,
		)

		if nodeErr != nil {
			// Publish flow failed event before returning error
			publishFlowFailedEvent(ctx, nodeErr, flowStartTime, time.Now().UnixMilli(), fe.observabilitySvc)
			return flowStep, nodeErr
		}

		fe.updateContextWithNodeResponse(ctx, nodeResp)

		nextNode, continueExecution, svcErr := fe.processNodeResponse(ctx, nodeResp, &flowStep, logger)
		if svcErr != nil {
			// Publish flow failed event before returning error
			publishFlowFailedEvent(ctx, svcErr, flowStartTime, time.Now().UnixMilli(), fe.observabilitySvc)
			return flowStep, svcErr
		}
		if !continueExecution {
			// Check if flow failed or just incomplete
			if flowStep.Status == common.FlowStatusError {
				publishFlowFailedEvent(ctx, nil, flowStartTime, time.Now().UnixMilli(), fe.observabilitySvc)
			}
			// Don't publish completed event here - flow is incomplete (waiting for user input)
			return flowStep, nil
		}
		currentNode = nextNode
	}

	// If we reach here, it means the flow has been executed successfully.
	flowStep.Status = common.FlowStatusComplete
	if ctx.Assertion != "" {
		flowStep.Assertion = ctx.Assertion
	}

	// Publish flow completed event
	flowEndTime := time.Now().UnixMilli()
	publishFlowCompletedEvent(ctx, flowStartTime, flowEndTime, fe.observabilitySvc)

	return flowStep, nil
}

// setCurrentExecutionNode sets the current execution node in the context and returns it.
func (fe *flowEngine) setCurrentExecutionNode(ctx *EngineContext, logger *log.Logger) (
	core.NodeInterface, *serviceerror.ServiceError) {
	graph := ctx.Graph
	if graph == nil {
		logger.Error("Flow graph is not initialized in the context")
		return nil, &serviceerror.InternalServerError
	}

	currentNode := ctx.CurrentNode
	if currentNode == nil {
		logger.Debug("Current node is nil. Setting start node as the current node.")
		var err error
		currentNode, err = graph.GetStartNode()
		if err != nil {
			logger.Error("Start node not found in the flow graph", log.Error(err))
			return nil, &serviceerror.InternalServerError
		}
		ctx.CurrentNode = currentNode
	}

	// Initialize execution history map if needed
	if ctx.ExecutionHistory == nil {
		ctx.ExecutionHistory = make(map[string]*common.NodeExecutionRecord)
	}

	return currentNode, nil
}

// getNodeInputs extracts required inputs for a node.
func getNodeInputs(node core.NodeInterface) []common.Input {
	if execNode, ok := node.(core.ExecutorBackedNodeInterface); ok {
		return execNode.GetInputs()
	}
	if promptNode, ok := node.(core.PromptNodeInterface); ok {
		var inputs []common.Input
		for _, prompt := range promptNode.GetPrompts() {
			inputs = append(inputs, prompt.Inputs...)
		}
		return inputs
	}
	return nil
}

// setNodeExecutor sets the executor for the given node if it is not already set.
func (fe *flowEngine) setNodeExecutor(node core.NodeInterface, logger *log.Logger) *serviceerror.ServiceError {
	if node.GetType() != common.NodeTypeTaskExecution {
		return nil
	}
	executableNode, ok := node.(core.ExecutorBackedNodeInterface)
	if !ok {
		logger.Error("Task execution node does not implement ExecutorBackedNodeInterface",
			log.String("nodeID", node.GetID()))
		return &serviceerror.InternalServerError
	}

	// Return if executor is already set
	if executableNode.GetExecutor() != nil {
		return nil
	}

	logger.Debug("Executor not set for the node. Constructing executor.", log.String("nodeID", node.GetID()))

	executorName := executableNode.GetExecutorName()
	if executorName == "" {
		logger.Error("Executor name not configured for executable node", log.String("nodeID", node.GetID()))
		return &serviceerror.InternalServerError
	}

	executor, err := fe.getExecutorByName(executorName)
	if err != nil {
		logger.Error("Error constructing executor for node", log.String("nodeID", node.GetID()),
			log.String("executorName", executorName), log.Error(err))
		return &serviceerror.InternalServerError
	}

	executableNode.SetExecutor(executor)
	return nil
}

// getExecutorByName retrieves executor instance from the executor registry.
func (fe *flowEngine) getExecutorByName(executorName string) (core.ExecutorInterface, error) {
	exec, err := fe.executorRegistry.GetExecutor(executorName)
	if err != nil {
		return nil, fmt.Errorf("failed to get executor '%s': %w", executorName, err)
	}

	return exec, nil
}

// clearSensitiveInputs removes sensitive user inputs from the engine context after a node has executed.
// This cleanup is only applied for authentication flows.
func (fe *flowEngine) clearSensitiveInputs(ctx *EngineContext, node core.NodeInterface) {
	if ctx.FlowType != common.FlowTypeAuthentication {
		return
	}

	execNode, ok := node.(core.ExecutorBackedNodeInterface)
	if !ok {
		return
	}

	// Get inputs from the node configuration. If the node does not define its own inputs,
	// fall back to the executor's default inputs.
	inputs := execNode.GetInputs()
	if len(inputs) == 0 {
		if executor := execNode.GetExecutor(); executor != nil {
			inputs = executor.GetDefaultInputs()
		}
	}

	for _, input := range inputs {
		if input.IsSensitive() {
			delete(ctx.UserInputs, input.Identifier)
		}
	}
}

// updateContextWithNodeResponse updates the engine context with the node response and authenticated user.
func (fe *flowEngine) updateContextWithNodeResponse(engineCtx *EngineContext, nodeResp *common.NodeResponse) {
	engineCtx.CurrentNodeResponse = nodeResp

	// Clear action only when node completes or forwards
	if nodeResp.Status == common.NodeStatusComplete || nodeResp.Status == common.NodeStatusForward {
		engineCtx.CurrentAction = ""
	}

	// Handle runtime data from the node response
	if len(nodeResp.RuntimeData) > 0 {
		if engineCtx.RuntimeData == nil {
			engineCtx.RuntimeData = make(map[string]string)
		}
		engineCtx.RuntimeData = sysutils.MergeStringMaps(engineCtx.RuntimeData, nodeResp.RuntimeData)
	}

	// Handle additional data from the node response (e.g., passkeyCreationOptions, passkeyChallenge)
	if len(nodeResp.AdditionalData) > 0 {
		if engineCtx.AdditionalData == nil {
			engineCtx.AdditionalData = make(map[string]string)
		}
		engineCtx.AdditionalData = sysutils.MergeStringMaps(engineCtx.AdditionalData, nodeResp.AdditionalData)
	}

	// Handle authenticated user from the node response
	if fe.shouldUpdateAuthenticatedUser(engineCtx) {
		prevAuthnUserAttrs := engineCtx.AuthenticatedUser.Attributes
		engineCtx.AuthenticatedUser = nodeResp.AuthenticatedUser

		// If engine context already had authenticated user attributes, merge them with the new ones.
		// Here if the same attribute exists in both, the one from the node response will take precedence.
		if len(prevAuthnUserAttrs) > 0 {
			if engineCtx.AuthenticatedUser.Attributes == nil {
				engineCtx.AuthenticatedUser.Attributes = prevAuthnUserAttrs
			} else {
				engineCtx.AuthenticatedUser.Attributes = sysutils.MergeInterfaceMaps(
					prevAuthnUserAttrs, engineCtx.AuthenticatedUser.Attributes)
			}
		}

		// Append user ID as a runtime data if not already set
		if engineCtx.AuthenticatedUser.UserID != "" {
			userID := engineCtx.RuntimeData["userID"]
			if userID == "" {
				if engineCtx.RuntimeData == nil {
					engineCtx.RuntimeData = make(map[string]string)
				}
				engineCtx.RuntimeData["userID"] = engineCtx.AuthenticatedUser.UserID
			}
		}
	}

	// Add assertion to the context
	if nodeResp.Assertion != "" {
		engineCtx.Assertion = nodeResp.Assertion
	}

	// Handle forwarded data from the node response
	// It replaces any existing forwarded data rather than merging
	if len(nodeResp.ForwardedData) > 0 {
		engineCtx.ForwardedData = nodeResp.ForwardedData
	}
}

// shouldUpdateAuthenticatedUser determines if the authenticated user should be updated in the context.
func (fe *flowEngine) shouldUpdateAuthenticatedUser(engineCtx *EngineContext) bool {
	currentNode := engineCtx.CurrentNode
	if currentNode == nil {
		return false
	}
	if currentNode.GetType() != common.NodeTypeTaskExecution {
		return false
	}

	executableNode, ok := currentNode.(core.ExecutorBackedNodeInterface)
	if !ok {
		return false
	}
	executorInst := executableNode.GetExecutor()
	if executorInst == nil {
		return false
	}

	// For authentication flows, only update from authentication executors or conditionally
	// from provisioning executor
	if engineCtx.FlowType == common.FlowTypeAuthentication {
		if executorInst.GetType() == common.ExecutorTypeAuthentication {
			return true
		}
		if engineCtx.RuntimeData != nil &&
			engineCtx.RuntimeData[common.RuntimeKeyUserEligibleForProvisioning] == "true" {
			return executorInst.GetName() == executor.ExecutorNameProvisioning
		}
		return false
	}

	// For registration flows, only update from provisioning executor or conditionally
	// from authentication executors
	if engineCtx.FlowType == common.FlowTypeRegistration {
		if engineCtx.RuntimeData != nil && engineCtx.RuntimeData[common.RuntimeKeySkipProvisioning] == "true" {
			return executorInst.GetType() == common.ExecutorTypeAuthentication
		}

		return executorInst.GetName() == executor.ExecutorNameProvisioning
	}

	// For user onboarding flows, update from authentication executors or from provisioning executor.
	if engineCtx.FlowType == common.FlowTypeUserOnboarding {
		return executorInst.GetType() == common.ExecutorTypeAuthentication ||
			executorInst.GetName() == executor.ExecutorNameProvisioning
	}

	return false
}

// processNodeResponse processes the node response and determines the next action.
// Returns:
// - The next node to execute.
// - Whether to continue execution.
// - Any service error.
func (fe *flowEngine) processNodeResponse(ctx *EngineContext, nodeResp *common.NodeResponse,
	flowStep *FlowStep, logger *log.Logger) (
	core.NodeInterface, bool, *serviceerror.ServiceError) {
	if nodeResp.Status == "" {
		logger.Error("Node response status not found in the flow graph")
		return nil, false, &serviceerror.InternalServerError
	}

	switch nodeResp.Status {
	case common.NodeStatusComplete:
		nextNode, svcErr := fe.handleCompletedResponse(ctx, nodeResp, logger)
		if svcErr != nil {
			return nil, false, svcErr
		}
		return nextNode, true, nil
	case common.NodeStatusIncomplete:
		svcErr := fe.handleIncompleteResponse(ctx, nodeResp, flowStep, logger)
		if svcErr != nil {
			return nil, false, svcErr
		}
		return nil, false, nil
	case common.NodeStatusForward:
		nextNode, svcErr := fe.handleForwardResponse(ctx, nodeResp, logger)
		if svcErr != nil {
			return nil, false, svcErr
		}
		return nextNode, true, nil
	case common.NodeStatusFailure:
		flowStep.Status = common.FlowStatusError
		flowStep.FailureReason = nodeResp.FailureReason
		return nil, false, nil
	default:
		logger.Error("Unsupported response status returned from the node",
			log.String("status", string(nodeResp.Status)))
		return nil, false, &serviceerror.InternalServerError
	}
}

// handleCompletedResponse handles the completed node and returns the next node to execute.
func (fe *flowEngine) handleCompletedResponse(ctx *EngineContext,
	nodeResp *common.NodeResponse, logger *log.Logger) (
	core.NodeInterface, *serviceerror.ServiceError) {
	nextNode, err := fe.resolveToNextNode(ctx.Graph, nodeResp)
	if err != nil {
		logger.Error("Error moving to the next node", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	ctx.CurrentNode = nextNode
	return nextNode, nil
}

// handleIncompleteResponse handles the node response when the status is incomplete.
// It resolves the flow step details based on the type of node response. The same node will be executed again
// in the next request with the required data.
func (fe *flowEngine) handleIncompleteResponse(ctx *EngineContext, nodeResp *common.NodeResponse,
	flowStep *FlowStep, logger *log.Logger) *serviceerror.ServiceError {
	switch nodeResp.Type {
	case common.NodeResponseTypeRedirection:
		err := fe.resolveStepForRedirection(ctx, nodeResp, flowStep)
		if err != nil {
			logger.Error("Error while resolving step for redirection", log.Error(err))
			return &serviceerror.InternalServerError
		}
		return nil
	case common.NodeResponseTypeView:
		err := fe.resolveStepDetailsForPrompt(ctx, nodeResp, flowStep)
		if err != nil {
			logger.Error("Error while resolving step details for prompt", log.Error(err))
			return &serviceerror.InternalServerError
		}
		return nil
	default:
		logger.Error("Unsupported response type returned from the node",
			log.String("responseType", string(nodeResp.Type)))
		return &serviceerror.InternalServerError
	}
	// TODO: Handle retry scenarios with nodeResp.Type == common.NodeResponseTypeRetry
}

// handleForwardResponse handles forwarding to the next node (e.g., onFailure handler)
func (fe *flowEngine) handleForwardResponse(ctx *EngineContext,
	nodeResp *common.NodeResponse, logger *log.Logger) (
	core.NodeInterface, *serviceerror.ServiceError) {
	logger.Debug("Forwarding to next node",
		log.String("nextNodeID", nodeResp.NextNodeID),
		log.String("failureReason", nodeResp.FailureReason))

	nextNode, err := fe.resolveToNextNode(ctx.Graph, nodeResp)
	if err != nil {
		logger.Error("Error resolving to next node", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	ctx.CurrentNode = nextNode
	return nextNode, nil
}

// skipToNextNode skips the current node and moves to the next node. It updates the context with the
// next node and returns it.
func (fe *flowEngine) skipToNextNode(ctx *EngineContext, currentNode core.NodeInterface,
	logger *log.Logger) (core.NodeInterface, *serviceerror.ServiceError) {
	condition := currentNode.GetCondition()

	// Condition must specify where to skip to
	if condition == nil || condition.OnSkip == "" {
		logger.Error("Node has condition but onSkip is not specified",
			log.String("nodeID", currentNode.GetID()))
		return nil, &serviceerror.InternalServerError
	}

	logger.Debug("Using condition's onSkip for skipped node",
		log.String("nodeID", currentNode.GetID()), log.String("onSkip", condition.OnSkip))

	nodeResp := &common.NodeResponse{NextNodeID: condition.OnSkip}

	// Resolve to the next node
	nextNode, err := fe.resolveToNextNode(ctx.Graph, nodeResp)
	if err != nil {
		logger.Error("Error moving to the next node after skipping", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	ctx.CurrentNode = nextNode
	return nextNode, nil
}

// resolveToNextNode resolves the next node to execute based on nodeResp.NextNodeID.
func (fe *flowEngine) resolveToNextNode(graph core.GraphInterface, nodeResp *common.NodeResponse) (
	core.NodeInterface, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "FlowEngine"))

	if nodeResp == nil || nodeResp.NextNodeID == "" {
		logger.Debug("No next node ID in response. Returning nil.")
		return nil, nil
	}

	nextNode, ok := graph.GetNode(nodeResp.NextNodeID)
	if !ok {
		return nil, errors.New("next node not found in the graph")
	}

	logger.Debug("Moving to next node", log.String("nextNodeID", nextNode.GetID()))
	return nextNode, nil
}

// resolveStepForRedirection resolves the flow step details for a redirection response.
func (fe *flowEngine) resolveStepForRedirection(ctx *EngineContext, nodeResp *common.NodeResponse,
	flowStep *FlowStep) error {
	if nodeResp == nil {
		return errors.New("node response is nil")
	}
	if nodeResp.RedirectURL == "" {
		return errors.New("redirect URL not found in the node response")
	}

	// Copy additional data from context (accumulated from all node responses)
	if len(ctx.AdditionalData) > 0 {
		flowStep.Data.AdditionalData = ctx.AdditionalData
	}

	flowStep.Data.RedirectURL = nodeResp.RedirectURL

	if flowStep.Data.Inputs == nil {
		flowStep.Data.Inputs = make([]common.Input, 0)
		flowStep.Data.Inputs = nodeResp.Inputs
	} else {
		// Append to the existing inputs
		flowStep.Data.Inputs = append(flowStep.Data.Inputs, nodeResp.Inputs...)
	}

	flowStep.Status = common.FlowStatusIncomplete
	flowStep.Type = common.StepTypeRedirection
	return nil
}

// resolveStepDetailsForPrompt resolves the step details for a user prompt response.
func (fe *flowEngine) resolveStepDetailsForPrompt(ctx *EngineContext, nodeResp *common.NodeResponse,
	flowStep *FlowStep) error {
	if nodeResp == nil {
		return errors.New("node response is nil")
	}
	if len(nodeResp.Inputs) == 0 && len(nodeResp.Actions) == 0 {
		return errors.New("no required data or actions found in the node response")
	}

	if len(nodeResp.Inputs) > 0 {
		if flowStep.Data.Inputs == nil {
			flowStep.Data.Inputs = make([]common.Input, 0)
			flowStep.Data.Inputs = nodeResp.Inputs
		} else {
			// Append to the existing inputs
			flowStep.Data.Inputs = append(flowStep.Data.Inputs, nodeResp.Inputs...)
		}
	}

	if len(nodeResp.Actions) > 0 {
		if flowStep.Data.Actions == nil {
			flowStep.Data.Actions = make([]common.Action, 0)
		}
		flowStep.Data.Actions = nodeResp.Actions
	}

	// Copy additional data from context (accumulated from all node responses)
	if len(ctx.AdditionalData) > 0 {
		flowStep.Data.AdditionalData = ctx.AdditionalData
	}

	// Include meta in the flow step if present
	if nodeResp.Meta != nil {
		flowStep.Data.Meta = nodeResp.Meta
	}

	// Include additionalData in the flow step if present
	if len(nodeResp.AdditionalData) > 0 {
		if flowStep.Data.AdditionalData == nil {
			flowStep.Data.AdditionalData = make(map[string]string)
		}
		for key, value := range nodeResp.AdditionalData {
			flowStep.Data.AdditionalData[key] = value
		}
	}

	// Set failure reason if present (e.g., when handling onFailure)
	if nodeResp.FailureReason != "" {
		flowStep.FailureReason = nodeResp.FailureReason
	}

	flowStep.Status = common.FlowStatusIncomplete
	flowStep.Type = common.StepTypeView
	return nil
}

// recordNodeExecution adds or updates execution record for the node.
func recordNodeExecution(ctx *EngineContext, node core.NodeInterface, nodeResp *common.NodeResponse,
	nodeErr *serviceerror.ServiceError, executionStartTime int64, executionEndTime int64) {
	nodeID := node.GetID()
	record := ctx.ExecutionHistory[nodeID]

	// Create new record if it does not exist
	if record == nil {
		nextStep := len(ctx.ExecutionHistory) + 1
		newRecord := createExecutionRecord(node, nextStep)
		ctx.ExecutionHistory[nodeID] = &newRecord
		record = &newRecord
	}

	attempt := createExecutionAttempt(record, nodeResp, nodeErr, executionStartTime, executionEndTime)
	record.Executions = append(record.Executions, attempt)

	record.Status = attempt.Status
	record.EndTime = attempt.EndTime
}

// createExecutionRecord creates a new node execution record.
func createExecutionRecord(node core.NodeInterface, step int) common.NodeExecutionRecord {
	record := common.NodeExecutionRecord{
		NodeID:     node.GetID(),
		NodeType:   string(node.GetType()),
		Step:       step,
		Status:     common.FlowStatusIncomplete,
		Executions: make([]common.ExecutionAttempt, 0),
		StartTime:  time.Now().Unix(),
	}

	// Set executor details if applicable (only for executor-backed nodes)
	if node.GetType() == common.NodeTypeTaskExecution {
		if executableNode, ok := node.(core.ExecutorBackedNodeInterface); ok {
			executor := executableNode.GetExecutor()
			if executor != nil {
				record.ExecutorName = executor.GetName()
				record.ExecutorType = executor.GetType()
			}
			record.ExecutorMode = executableNode.GetMode()
		}
	}

	return record
}

// createExecutionAttempt creates a new execution attempt.
func createExecutionAttempt(nodeRecord *common.NodeExecutionRecord, nodeResp *common.NodeResponse,
	nodeErr *serviceerror.ServiceError, executionStartTime int64, executionEndTime int64) common.ExecutionAttempt {
	attempt := common.ExecutionAttempt{
		Attempt:   len(nodeRecord.Executions) + 1,
		Timestamp: executionEndTime,
		StartTime: executionStartTime,
		EndTime:   executionEndTime,
	}

	// Determine status
	if nodeErr != nil {
		attempt.Status = common.FlowStatusError
	} else if nodeResp != nil {
		switch nodeResp.Status {
		case common.NodeStatusComplete:
			attempt.Status = common.FlowStatusComplete
		case common.NodeStatusIncomplete:
			attempt.Status = common.FlowStatusIncomplete
		case common.NodeStatusFailure:
			attempt.Status = common.FlowStatusError
		default:
			attempt.Status = common.FlowStatusIncomplete
		}
	}

	return attempt
}

// publishNodeExecutionStartedEvent publishes an observability event when node execution starts.
func publishNodeExecutionStartedEvent(
	ctx *EngineContext,
	node core.NodeInterface,
	obsSvc observability.ObservabilityServiceInterface,
) {
	if obsSvc == nil || !obsSvc.IsEnabled() {
		return
	}

	// Get node execution record to determine step number and attempt
	record := ctx.ExecutionHistory[node.GetID()]
	stepNumber := len(ctx.ExecutionHistory) + 1
	attemptNumber := 1
	if record != nil {
		stepNumber = record.Step
		attemptNumber = len(record.Executions) + 1
	}

	evt := event.NewEvent(
		ctx.FlowID, // Use FlowID as TraceID
		string(event.EventTypeFlowNodeExecutionStarted),
		event.ComponentFlowEngine,
	).
		WithStatus(event.StatusInProgress).
		WithData(event.DataKey.FlowID, ctx.FlowID).
		WithData(event.DataKey.FlowType, string(ctx.FlowType)).
		WithData(event.DataKey.NodeID, node.GetID()).
		WithData(event.DataKey.NodeType, string(node.GetType())).
		WithData(event.DataKey.StepNumber, fmt.Sprintf("%d", stepNumber)).
		WithData(event.DataKey.AttemptNumber, fmt.Sprintf("%d", attemptNumber)).
		WithData(event.DataKey.AppID, ctx.AppID)

	obsSvc.PublishEvent(evt)
}

// publishNodeExecutionCompletedEvent publishes an observability event when node execution completes or fails.
func publishNodeExecutionCompletedEvent(ctx *EngineContext, node core.NodeInterface,
	nodeResp *common.NodeResponse, nodeErr *serviceerror.ServiceError,
	executionStartTime int64, executionEndTime int64, obsSvc observability.ObservabilityServiceInterface) {
	if obsSvc == nil || !obsSvc.IsEnabled() {
		return
	}

	// Get node execution record to determine step number and attempt
	record := ctx.ExecutionHistory[node.GetID()]
	if record == nil {
		return
	}

	stepNumber := record.Step
	attemptNumber := len(record.Executions)

	// Determine event type and status based on outcome
	var eventType event.EventType
	var status string
	var nodeStatus string

	if nodeErr != nil {
		eventType = event.EventTypeFlowNodeExecutionFailed
		status = event.StatusFailure
		nodeStatus = string(common.FlowStatusError)
	} else if nodeResp != nil {
		switch nodeResp.Status {
		case common.NodeStatusComplete:
			eventType = event.EventTypeFlowNodeExecutionCompleted
			status = event.StatusSuccess
			nodeStatus = string(common.FlowStatusComplete)
		case common.NodeStatusIncomplete:
			eventType = event.EventTypeFlowNodeExecutionCompleted
			status = event.StatusSuccess
			nodeStatus = string(common.FlowStatusIncomplete)
		case common.NodeStatusFailure:
			eventType = event.EventTypeFlowNodeExecutionFailed
			status = event.StatusFailure
			nodeStatus = string(common.FlowStatusError)
		default:
			eventType = event.EventTypeFlowNodeExecutionCompleted
			status = event.StatusSuccess
			nodeStatus = string(nodeResp.Status)
		}
	} else {
		eventType = event.EventTypeFlowNodeExecutionCompleted
		status = event.StatusSuccess
		nodeStatus = string(common.FlowStatusComplete)
	}

	// Calculate duration in milliseconds
	durationMs := executionEndTime - executionStartTime

	evt := event.NewEvent(
		ctx.FlowID, // Use FlowID as TraceID
		string(eventType),
		event.ComponentFlowEngine,
	).
		WithStatus(status).
		WithData(event.DataKey.FlowID, ctx.FlowID).
		WithData(event.DataKey.FlowType, string(ctx.FlowType)).
		WithData(event.DataKey.NodeID, node.GetID()).
		WithData(event.DataKey.NodeType, string(node.GetType())).
		WithData(event.DataKey.NodeStatus, nodeStatus).
		WithData(event.DataKey.StepNumber, fmt.Sprintf("%d", stepNumber)).
		WithData(event.DataKey.AttemptNumber, fmt.Sprintf("%d", attemptNumber)).
		WithData(event.DataKey.DurationMs, fmt.Sprintf("%d", durationMs)).
		WithData(event.DataKey.AppID, ctx.AppID)

	// Add error or failure details
	if nodeErr != nil {
		evt.WithData(event.DataKey.Error, nodeErr.Error).
			WithData(event.DataKey.ErrorCode, nodeErr.Code).
			WithData(event.DataKey.ErrorType, string(nodeErr.Type))
		if nodeErr.ErrorDescription != "" {
			evt.WithData(event.DataKey.Message, nodeErr.ErrorDescription)
		}
	} else if nodeResp != nil && nodeResp.FailureReason != "" {
		evt.WithData(event.DataKey.FailureReason, nodeResp.FailureReason)
	}

	// Add user ID if authenticated
	if ctx.AuthenticatedUser.IsAuthenticated && ctx.AuthenticatedUser.UserID != "" {
		evt.WithData(event.DataKey.UserID, ctx.AuthenticatedUser.UserID)
	}

	obsSvc.PublishEvent(evt)
}

// publishFlowStartedEvent publishes an observability event when flow execution starts.
func publishFlowStartedEvent(ctx *EngineContext, obsSvc observability.ObservabilityServiceInterface) {
	if obsSvc == nil || !obsSvc.IsEnabled() {
		return
	}

	evt := event.NewEvent(
		ctx.TraceID, // Use TraceID from context
		string(event.EventTypeFlowStarted),
		event.ComponentFlowEngine,
	).
		WithStatus(event.StatusInProgress).
		WithData(event.DataKey.FlowID, ctx.FlowID).
		WithData(event.DataKey.FlowType, string(ctx.FlowType)).
		WithData(event.DataKey.AppID, ctx.AppID)

	// Add user ID if already authenticated
	if ctx.AuthenticatedUser.IsAuthenticated && ctx.AuthenticatedUser.UserID != "" {
		evt.WithData(event.DataKey.UserID, ctx.AuthenticatedUser.UserID)
	}

	obsSvc.PublishEvent(evt)
}

// publishFlowCompletedEvent publishes an observability event when flow execution completes successfully.
func publishFlowCompletedEvent(
	ctx *EngineContext,
	flowStartTime int64,
	flowEndTime int64,
	obsSvc observability.ObservabilityServiceInterface,
) {
	if obsSvc == nil || !obsSvc.IsEnabled() {
		return
	}

	// Calculate duration in milliseconds
	durationMs := flowEndTime - flowStartTime

	evt := event.NewEvent(
		ctx.FlowID, // Use FlowID as TraceID
		string(event.EventTypeFlowCompleted),
		event.ComponentFlowEngine,
	).
		WithStatus(event.StatusSuccess).
		WithData(event.DataKey.FlowID, ctx.FlowID).
		WithData(event.DataKey.FlowType, string(ctx.FlowType)).
		WithData(event.DataKey.AppID, ctx.AppID).
		WithData(event.DataKey.DurationMs, fmt.Sprintf("%d", durationMs))

	// Add user ID if authenticated
	if ctx.AuthenticatedUser.IsAuthenticated && ctx.AuthenticatedUser.UserID != "" {
		evt.WithData(event.DataKey.UserID, ctx.AuthenticatedUser.UserID)
	}

	obsSvc.PublishEvent(evt)
}

// publishFlowFailedEvent publishes an observability event when flow execution fails.
func publishFlowFailedEvent(ctx *EngineContext, svcErr *serviceerror.ServiceError,
	flowStartTime int64, flowEndTime int64, obsSvc observability.ObservabilityServiceInterface) {
	if obsSvc == nil || !obsSvc.IsEnabled() {
		return
	}

	// Calculate duration in milliseconds
	durationMs := flowEndTime - flowStartTime

	evt := event.NewEvent(
		ctx.TraceID, // Use TraceID from context
		string(event.EventTypeFlowFailed),
		event.ComponentFlowEngine,
	).
		WithStatus(event.StatusFailure).
		WithData(event.DataKey.FlowID, ctx.FlowID).
		WithData(event.DataKey.FlowType, string(ctx.FlowType)).
		WithData(event.DataKey.AppID, ctx.AppID).
		WithData(event.DataKey.DurationMs, fmt.Sprintf("%d", durationMs))

	// Add error details if available
	if svcErr != nil {
		evt.WithData(event.DataKey.Error, svcErr.Error).
			WithData(event.DataKey.ErrorCode, svcErr.Code).
			WithData(event.DataKey.ErrorType, string(svcErr.Type))
		if svcErr.ErrorDescription != "" {
			evt.WithData(event.DataKey.Message, svcErr.ErrorDescription)
		}
	}

	// Add user ID if authenticated
	if ctx.AuthenticatedUser.IsAuthenticated && ctx.AuthenticatedUser.UserID != "" {
		evt.WithData(event.DataKey.UserID, ctx.AuthenticatedUser.UserID)
	}

	obsSvc.PublishEvent(evt)
}
