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

// Package engine provides the flow engine for orchestrating flow executions.
package engine

import (
	"errors"
	"sync"

	"github.com/asgardeo/thunder/internal/flow/constants"
	"github.com/asgardeo/thunder/internal/flow/model"
	"github.com/asgardeo/thunder/internal/flow/utils"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

var (
	instance *FlowEngine
	once     sync.Once
)

// FlowEngineInterface defines the interface for the flow engine.
type FlowEngineInterface interface {
	Execute(ctx *model.EngineContext) (model.FlowStep, *serviceerror.ServiceError)
}

// FlowEngine is the main engine implementation for orchestrating flow executions.
type FlowEngine struct{}

// GetFlowEngine returns a singleton instance of FlowEngine.
func GetFlowEngine() FlowEngineInterface {
	once.Do(func() {
		instance = &FlowEngine{}
	})
	return instance
}

// Execute executes a step in the flow
func (fe *FlowEngine) Execute(ctx *model.EngineContext) (model.FlowStep, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "FlowEngine"))

	flowStep := model.FlowStep{
		FlowID: ctx.FlowID,
	}

	currentNode, err := setCurrentExecutionNode(ctx, logger)
	if err != nil {
		return flowStep, err
	}

	// Execute the graph nodes until a terminal condition is met or currentNode is nil
	for currentNode != nil {
		logger.Debug("Executing node", log.String("nodeID", currentNode.GetID()),
			log.String("nodeType", string(currentNode.GetType())))

		svcErr := setNodeExecutor(currentNode, logger)
		if svcErr != nil {
			return flowStep, svcErr
		}

		nodeCtx := &model.NodeContext{
			FlowID:            ctx.FlowID,
			FlowType:          ctx.FlowType,
			AppID:             ctx.AppID,
			CurrentActionID:   ctx.CurrentActionID,
			NodeInputData:     ctx.CurrentNode.GetInputData(),
			UserInputData:     ctx.UserInputData,
			RuntimeData:       ctx.RuntimeData,
			Application:       ctx.Application,
			AuthenticatedUser: ctx.AuthenticatedUser,
		}
		if nodeCtx.NodeInputData == nil {
			nodeCtx.NodeInputData = make([]model.InputData, 0)
		}
		if nodeCtx.UserInputData == nil {
			nodeCtx.UserInputData = make(map[string]string)
		}
		if nodeCtx.RuntimeData == nil {
			nodeCtx.RuntimeData = make(map[string]string)
		}

		nodeResp, nodeErr := currentNode.Execute(nodeCtx)
		if nodeErr != nil {
			return flowStep, nodeErr
		}

		updateContextWithNodeResponse(ctx, nodeResp)

		nextNode, continueExecution, svcErr := fe.processNodeResponse(ctx, currentNode, nodeResp, &flowStep)
		if svcErr != nil {
			return flowStep, svcErr
		}
		if !continueExecution {
			return flowStep, nil
		}
		currentNode = nextNode
	}

	// If we reach here, it means the flow has been executed successfully.
	flowStep.Status = constants.FlowStatusComplete
	if ctx.CurrentNodeResponse != nil && ctx.CurrentNodeResponse.Assertion != "" {
		flowStep.Assertion = ctx.CurrentNodeResponse.Assertion
	}

	return flowStep, nil
}

// setCurrentExecutionNode sets the current execution node in the context and returns it.
func setCurrentExecutionNode(ctx *model.EngineContext, logger *log.Logger) (model.NodeInterface,
	*serviceerror.ServiceError) {
	graph := ctx.Graph
	if graph == nil {
		return nil, &constants.ErrorFlowGraphNotInitialized
	}

	currentNode := ctx.CurrentNode
	if currentNode == nil {
		logger.Debug("Current node is nil. Setting start node as the current node.")
		var err error
		currentNode, err = graph.GetStartNode()
		if err != nil {
			return nil, &constants.ErrorStartNodeNotFoundInGraph
		}
		ctx.CurrentNode = currentNode
	}

	return currentNode, nil
}

// setNodeExecutor sets the executor for the given node if it is not already set.
func setNodeExecutor(node model.NodeInterface, logger *log.Logger) *serviceerror.ServiceError {
	if node.GetType() != constants.NodeTypeTaskExecution {
		return nil
	}

	if node.GetExecutor() == nil {
		logger.Debug("Executor not set for the node. Constructing executor.", log.String("nodeID", node.GetID()))

		executor, err := utils.GetExecutorByName(node.GetExecutorConfig())
		if err != nil {
			logger.Error("Error constructing executor for node", log.String("nodeID", node.GetID()),
				log.String("executorName", node.GetExecutorConfig().Name), log.Error(err))
			return &constants.ErrorConstructingNodeExecutor
		}
		node.SetExecutor(executor)
	}

	return nil
}

// updateContextWithNodeResponse updates the engine context with the node response and authenticated user.
func updateContextWithNodeResponse(engineCtx *model.EngineContext, nodeResp *model.NodeResponse) {
	engineCtx.CurrentNodeResponse = nodeResp
	engineCtx.CurrentActionID = ""

	// Handle runtime data from the node response
	if len(nodeResp.RuntimeData) > 0 {
		if engineCtx.RuntimeData == nil {
			engineCtx.RuntimeData = make(map[string]string)
		}
		engineCtx.RuntimeData = sysutils.MergeStringMaps(engineCtx.RuntimeData, nodeResp.RuntimeData)
	}

	// Handle authenticated user from the node response
	if nodeResp.AuthenticatedUser.IsAuthenticated || engineCtx.FlowType == constants.FlowTypeRegistration {
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
}

// processNodeResponse processes the node response and determines the next action.
// Returns:
// - The next node to execute.
// - Whether to continue execution.
// - Any service error.
func (fe *FlowEngine) processNodeResponse(ctx *model.EngineContext, currentNode model.NodeInterface,
	nodeResp *model.NodeResponse, flowStep *model.FlowStep) (model.NodeInterface, bool, *serviceerror.ServiceError) {
	if nodeResp.Status == "" {
		return nil, false, &constants.ErrorNodeResponseStatusNotFound
	}
	if nodeResp.Status == constants.NodeStatusComplete {
		nextNode, svcErr := fe.handleCompletedResponse(ctx, currentNode, nodeResp)
		if svcErr != nil {
			return nil, false, svcErr
		}
		return nextNode, true, nil
	} else if nodeResp.Status == constants.NodeStatusIncomplete {
		svcErr := fe.handleIncompleteResponse(nodeResp, flowStep)
		if svcErr != nil {
			return nil, false, svcErr
		}
		return nil, false, nil
	} else if nodeResp.Status == constants.NodeStatusFailure {
		flowStep.Status = constants.FlowStatusError
		flowStep.FailureReason = nodeResp.FailureReason
		return nil, false, nil
	} else {
		svcErr := constants.ErrorUnsupportedNodeResponseStatus
		svcErr.ErrorDescription = "unsupported status returned from the node: " + string(nodeResp.Status)
		return nil, false, &svcErr
	}
}

// handleCompletedResponse handles the completed node and returns the next node to execute.
func (fe *FlowEngine) handleCompletedResponse(ctx *model.EngineContext, currentNode model.NodeInterface,
	nodeResp *model.NodeResponse) (model.NodeInterface, *serviceerror.ServiceError) {
	nextNode, err := fe.resolveToNextNode(ctx.Graph, currentNode, nodeResp)
	if err != nil {
		svcErr := constants.ErrorMovingToNextNode
		svcErr.ErrorDescription = "error moving to next node: " + err.Error()
		return nil, &svcErr
	}
	ctx.CurrentNode = nextNode
	return nextNode, nil
}

// handleIncompleteResponse handles the node response when the status is incomplete.
// It resolves the flow step details based on the type of node response. The same node will be executed again
// in the next request with the required data.
func (fe *FlowEngine) handleIncompleteResponse(nodeResp *model.NodeResponse,
	flowStep *model.FlowStep) *serviceerror.ServiceError {
	if nodeResp.Type == constants.NodeResponseTypeRedirection {
		err := fe.resolveStepForRedirection(nodeResp, flowStep)
		if err != nil {
			svcErr := constants.ErrorResolvingStepForRedirection
			svcErr.ErrorDescription = "error resolving step for redirection: " + err.Error()
			return &svcErr
		}
		return nil
	} else if nodeResp.Type == constants.NodeResponseTypeView {
		err := fe.resolveStepDetailsForPrompt(nodeResp, flowStep)
		if err != nil {
			svcErr := constants.ErrorResolvingStepForPrompt
			svcErr.ErrorDescription = "error resolving step for prompt: " + err.Error()
			return &svcErr
		}
		return nil
	} else {
		svcErr := constants.ErrorUnsupportedNodeResponseType
		svcErr.ErrorDescription = "unsupported node response type: " + string(nodeResp.Type)
		return &svcErr
	}
	// TODO: Handle retry scenarios with nodeResp.Type == constants.NodeResponseTypeRetry
}

// resolveToNextNode resolves the next node to execute based on the current node.
func (fe *FlowEngine) resolveToNextNode(graph model.GraphInterface, currentNode model.NodeInterface,
	nodeResp *model.NodeResponse) (model.NodeInterface, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "FlowEngine"))

	nextNodeID := ""
	if currentNode.GetType() == constants.NodeTypeDecision {
		logger.Debug("Current node is a decision node. Trying to resolve next node based on decision.")
		if nodeResp == nil || nodeResp.NextNodeID == "" {
			logger.Debug("No next node ID found in the node response. Returning nil.")
			return nil, nil
		}
		nextNodeID = nodeResp.NextNodeID
	} else {
		// Set the first element of the next node list assuming only decision nodes can have multiple next nodes.
		if len(currentNode.GetNextNodeList()) == 0 {
			logger.Debug("No next node found in the current node. Returning nil.")
			return nil, nil
		}
		nextNodeID = currentNode.GetNextNodeList()[0]
	}
	if nextNodeID == "" {
		logger.Debug("No next node found. Returning nil.")
		return nil, nil
	}

	nextNode, ok := graph.GetNode(nextNodeID)
	if !ok {
		return nil, errors.New("next node not found in the graph")
	}

	logger.Debug("Moving to next node", log.String("nextNodeID", nextNode.GetID()))
	return nextNode, nil
}

// resolveStepForRedirection resolves the flow step details for a redirection response.
func (fe *FlowEngine) resolveStepForRedirection(nodeResp *model.NodeResponse, flowStep *model.FlowStep) error {
	if nodeResp == nil {
		return errors.New("node response is nil")
	}
	if nodeResp.RedirectURL == "" {
		return errors.New("redirect URL not found in the node response")
	}

	if flowStep.Data.AdditionalData == nil {
		flowStep.Data.AdditionalData = make(map[string]string)
		flowStep.Data.AdditionalData = nodeResp.AdditionalData
	} else {
		// Append to the existing additional info
		for key, value := range nodeResp.AdditionalData {
			flowStep.Data.AdditionalData[key] = value
		}
	}

	flowStep.Data.RedirectURL = nodeResp.RedirectURL

	if flowStep.Data.Inputs == nil {
		flowStep.Data.Inputs = make([]model.InputData, 0)
		flowStep.Data.Inputs = nodeResp.RequiredData
	} else {
		// Append to the existing input data
		flowStep.Data.Inputs = append(flowStep.Data.Inputs, nodeResp.RequiredData...)
	}

	flowStep.Status = constants.FlowStatusIncomplete
	flowStep.Type = constants.StepTypeRedirection
	return nil
}

// resolveStepDetailsForPrompt resolves the step details for a user prompt response.
func (fe *FlowEngine) resolveStepDetailsForPrompt(nodeResp *model.NodeResponse, flowStep *model.FlowStep) error {
	if nodeResp == nil {
		return errors.New("node response is nil")
	}
	if len(nodeResp.RequiredData) == 0 && len(nodeResp.Actions) == 0 {
		return errors.New("no required data or actions found in the node response")
	}

	if len(nodeResp.RequiredData) > 0 {
		if flowStep.Data.Inputs == nil {
			flowStep.Data.Inputs = make([]model.InputData, 0)
			flowStep.Data.Inputs = nodeResp.RequiredData
		} else {
			// Append to the existing input data
			flowStep.Data.Inputs = append(flowStep.Data.Inputs, nodeResp.RequiredData...)
		}
	}

	if len(nodeResp.Actions) > 0 {
		if flowStep.Data.Actions == nil {
			flowStep.Data.Actions = make([]model.Action, 0)
		}
		flowStep.Data.Actions = nodeResp.Actions
	}

	flowStep.Status = constants.FlowStatusIncomplete
	flowStep.Type = constants.StepTypeView
	return nil
}
