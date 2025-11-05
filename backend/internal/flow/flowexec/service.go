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

// Package flowexec provides the FlowExecService interface and its implementation.
package flowexec

import (
	"fmt"

	"github.com/asgardeo/thunder/internal/application"
	"github.com/asgardeo/thunder/internal/flow/common/constants"
	"github.com/asgardeo/thunder/internal/flow/common/model"
	"github.com/asgardeo/thunder/internal/flow/flowmgt"
	"github.com/asgardeo/thunder/internal/observability"
	"github.com/asgardeo/thunder/internal/observability/event"

	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

// FlowExecServiceInterface defines the interface for flow orchestration and acts as the entry point for flow execution
type FlowExecServiceInterface interface {
	Execute(appID, flowID, actionID, flowType string, inputData map[string]string) (
		*model.FlowStep, *serviceerror.ServiceError)
	InitiateFlow(initContext *model.FlowInitContext) (string, *serviceerror.ServiceError)
}

// flowExecService is the implementation of FlowExecServiceInterface
type flowExecService struct {
	flowEngine     flowEngineInterface
	flowMgtService flowmgt.FlowMgtServiceInterface
	flowStore      FlowStoreInterface
	appService     application.ApplicationServiceInterface
}

func newFlowExecService(flowMgtService flowmgt.FlowMgtServiceInterface,
	applicationService application.ApplicationServiceInterface,
	flowEngine flowEngineInterface) FlowExecServiceInterface {
	return &flowExecService{
		flowMgtService: flowMgtService,
		flowStore:      NewFlowStore(),
		appService:     applicationService,
		flowEngine:     flowEngine,
	}
}

// Execute executes a flow with the given data
func (s *flowExecService) Execute(appID, flowID, actionID, flowType string, inputData map[string]string) (
	*model.FlowStep, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "FlowExecService"))

	var context *model.EngineContext
	var loadErr *serviceerror.ServiceError

	if isNewFlow(flowID) {
		context, loadErr = s.loadNewContext(appID, actionID, flowType, inputData)
		if loadErr != nil {
			logger.Error("Failed to load new flow context",
				log.String("appID", appID),
				log.String("flowType", flowType),
				log.String("error", loadErr.Error))
			return nil, loadErr
		}

		// Publish flow started event for new flows
		startEvt := event.NewEvent(context.FlowID, string(event.EventTypeFlowStarted), event.ComponentFlowExecutionService)
		startEvt.WithData(event.DataKey.FlowID, context.FlowID).
			WithData(event.DataKey.FlowType, string(context.FlowType)).
			WithData(event.DataKey.AppID, appID).
			WithStatus(event.StatusInProgress)
		observability.GetService().PublishEvent(startEvt)
	} else {
		context, loadErr = s.loadPrevContext(flowID, actionID, inputData)
		if loadErr != nil {
			logger.Error("Failed to load previous flow context",
				log.String("flowID", flowID),
				log.String("error", loadErr.Error))
			return nil, loadErr
		}
	}

	flowStep, flowErr := s.flowEngine.Execute(context)

	if flowErr != nil {
		// Publish flow failed event
		failEvt := event.NewEvent(context.FlowID, string(event.EventTypeFlowFailed), event.ComponentFlowExecutionService)
		failEvt.WithData(event.DataKey.FlowID, context.FlowID).
			WithData(event.DataKey.FlowType, string(context.FlowType)).
			WithData(event.DataKey.AppID, context.AppID).
			WithData(event.DataKey.Error, flowErr.ErrorDescription).
			WithData(event.DataKey.ErrorCode, flowErr.Code).
			WithStatus(event.StatusFailure)
		observability.GetService().PublishEvent(failEvt)

		if !isNewFlow(flowID) {
			if removeErr := s.removeContext(context.FlowID, logger); removeErr != nil {
				logger.Error("Failed to remove flow context after engine failure",
					log.String("flowID", context.FlowID), log.Error(removeErr))
				return nil, &constants.ErrorUpdatingContextInStore
			}
		}
		return nil, flowErr
	}

	if isComplete(flowStep) {
		// Publish flow completed event
		completedEvt := event.NewEvent(
			context.FlowID,
			string(event.EventTypeFlowCompleted),
			event.ComponentFlowExecutionService,
		)
		completedEvt.WithData(event.DataKey.FlowID, context.FlowID).
			WithData(event.DataKey.FlowType, string(context.FlowType)).
			WithData(event.DataKey.AppID, context.AppID).
			WithStatus(event.StatusSuccess)

		// Add user ID if authenticated
		if context.AuthenticatedUser.IsAuthenticated && context.AuthenticatedUser.UserID != "" {
			completedEvt.WithData(event.DataKey.UserID, context.AuthenticatedUser.UserID)
		}
		observability.GetService().PublishEvent(completedEvt)

		if !isNewFlow(flowID) {
			if removeErr := s.removeContext(context.FlowID, logger); removeErr != nil {
				logger.Error("Failed to remove flow context after completion",
					log.String("flowID", context.FlowID), log.Error(removeErr))
				return nil, &constants.ErrorUpdatingContextInStore
			}
		}
	} else {
		if isNewFlow(flowID) {
			if storeErr := s.storeContext(context, logger); storeErr != nil {
				logger.Error("Failed to store initial flow context", log.String("flowID", context.FlowID), log.Error(storeErr))
				return nil, &constants.ErrorUpdatingContextInStore
			}
		} else {
			if updateErr := s.updateContext(context, &flowStep, logger); updateErr != nil {
				logger.Error("Failed to update flow context", log.String("flowID", context.FlowID), log.Error(updateErr))
				return nil, &constants.ErrorUpdatingContextInStore
			}
		}
	}

	return &flowStep, nil
}

// initContext initializes a new flow context with the given details.
func (s *flowExecService) loadNewContext(appID, actionID, flowTypeStr string,
	inputData map[string]string) (*model.EngineContext, *serviceerror.ServiceError) {
	flowType, err := validateFlowType(flowTypeStr)
	if err != nil {
		return nil, err
	}

	ctx, err := s.initContext(appID, flowType)
	if err != nil {
		return nil, err
	}

	prepareContext(ctx, actionID, inputData)
	return ctx, nil
}

// initContext initializes a new flow context with the given details.
func (s *flowExecService) initContext(appID string,
	flowType constants.FlowType) (*model.EngineContext, *serviceerror.ServiceError) {
	graphID, svcErr := s.getFlowGraph(appID, flowType)
	if svcErr != nil {
		return nil, svcErr
	}

	ctx := model.EngineContext{}
	flowID := sysutils.GenerateUUID()
	ctx.FlowID = flowID

	graph, ok := s.flowMgtService.GetGraph(graphID)
	if !ok {
		return nil, &constants.ErrorFlowGraphNotFound
	}
	ctx.FlowType = graph.GetType()
	ctx.Graph = graph
	ctx.AppID = appID

	svcErr = s.setApplicationToContext(&ctx)
	if svcErr != nil {
		return nil, svcErr
	}

	return &ctx, nil
}

// loadPrevContext retrieves the flow context from the store based on the given details.
func (s *flowExecService) loadPrevContext(flowID, actionID string, inputData map[string]string) (
	*model.EngineContext, *serviceerror.ServiceError) {
	ctx, err := s.loadContextFromStore(flowID)
	if err != nil {
		return nil, err
	}

	prepareContext(ctx, actionID, inputData)
	return ctx, nil
}

// loadContextFromStore retrieves the flow context from the store based on the given details.
func (s *flowExecService) loadContextFromStore(flowID string) (*model.EngineContext,
	*serviceerror.ServiceError) {
	if flowID == "" {
		return nil, &constants.ErrorInvalidFlowID
	}

	dbModel, err := s.flowStore.GetFlowContext(flowID)
	if err != nil {
		return nil, &constants.ErrorRetrievingContextInStore
	}

	if dbModel == nil {
		return nil, &constants.ErrorInvalidFlowID
	}

	graph, exists := s.flowMgtService.GetGraph(dbModel.GraphID)
	if !exists {
		return nil, &constants.ErrorFlowGraphNotFound
	}

	engineContext, err := dbModel.ToEngineContext(graph)
	if err != nil {
		return nil, &constants.ErrorFlowContextConversionFailed
	}

	svcErr := s.setApplicationToContext(&engineContext)
	if svcErr != nil {
		return nil, svcErr
	}

	return &engineContext, nil
}

// setApplicationToContext retrieves the application and sets it to the flow context.
func (s *flowExecService) setApplicationToContext(ctx *model.EngineContext) *serviceerror.ServiceError {
	app, err := s.appService.GetApplication(ctx.AppID)
	if err != nil {
		if err.Code == application.ErrorApplicationNotFound.Code {
			return &constants.ErrorInvalidAppID
		}
		if err.Type == serviceerror.ClientErrorType {
			svcErr := &constants.ErrorApplicationRetrievalClientError
			svcErr.ErrorDescription = fmt.Sprintf("Error while retrieving application: %s", err.ErrorDescription)
			return svcErr
		}
		return &constants.ErrorApplicationRetrievalServerError
	}
	if app == nil {
		svcErr := &constants.ErrorApplicationRetrievalServerError
		svcErr.ErrorDescription = "Application not found"
		return svcErr
	}
	ctx.Application = *app
	return nil
}

// removeContext removes the flow context from the store.
func (s *flowExecService) removeContext(flowID string, logger *log.Logger) error {
	if flowID == "" {
		return fmt.Errorf("flow ID cannot be empty")
	}

	err := s.flowStore.DeleteFlowContext(flowID)
	if err != nil {
		return fmt.Errorf("failed to remove flow context from database: %w", err)
	}

	logger.Debug("Flow context removed successfully from database", log.String("flowID", flowID))
	return nil
}

// updateContext updates the flow context in the store based on the flow step status.
func (s *flowExecService) updateContext(ctx *model.EngineContext, flowStep *model.FlowStep, logger *log.Logger) error {
	if flowStep.Status == constants.FlowStatusComplete {
		return s.removeContext(ctx.FlowID, logger)
	} else {
		logger.Debug("Flow execution is incomplete, updating the flow context",
			log.String("flowID", ctx.FlowID))

		if ctx.FlowID == "" {
			return fmt.Errorf("flow ID cannot be empty")
		}

		err := s.flowStore.UpdateFlowContext(*ctx)
		if err != nil {
			return fmt.Errorf("failed to update flow context in database: %w", err)
		}

		logger.Debug("Flow context updated successfully in database", log.String("flowID", ctx.FlowID))
		return nil
	}
}

// storeContext stores the flow context in the store.
func (s *flowExecService) storeContext(ctx *model.EngineContext, logger *log.Logger) error {
	if ctx.FlowID == "" {
		return fmt.Errorf("flow ID cannot be empty")
	}

	err := s.flowStore.StoreFlowContext(*ctx)
	if err != nil {
		return fmt.Errorf("failed to store flow context in database: %w", err)
	}

	logger.Debug("Flow context stored successfully in database", log.String("flowID", ctx.FlowID))
	return nil
}

// getFlowGraph checks if the provided application ID is valid and returns the associated flow graph.
func (s *flowExecService) getFlowGraph(appID string, flowType constants.FlowType) (string,
	*serviceerror.ServiceError) {
	if appID == "" {
		return "", &constants.ErrorInvalidAppID
	}

	app, err := s.appService.GetApplication(appID)
	if err != nil {
		if err.Code == application.ErrorApplicationNotFound.Code {
			return "", &constants.ErrorInvalidAppID
		}
		if err.Type == serviceerror.ClientErrorType {
			return "", &constants.ErrorApplicationRetrievalClientError
		}
		return "", &constants.ErrorApplicationRetrievalServerError
	}
	if app == nil {
		return "", &constants.ErrorInvalidAppID
	}

	if flowType == constants.FlowTypeRegistration {
		if app.RegistrationFlowGraphID == "" {
			return "", &constants.ErrorRegisFlowNotConfiguredForApplication
		} else if !app.IsRegistrationFlowEnabled {
			return "", &constants.ErrorRegistrationFlowDisabled
		}
		return app.RegistrationFlowGraphID, nil
	}

	// Default to authentication flow graph ID
	if app.AuthFlowGraphID == "" {
		return "", &constants.ErrorAuthFlowNotConfiguredForApplication
	}

	return app.AuthFlowGraphID, nil
}

// validateFlowType validates the provided flow type string and returns the corresponding FlowType.
func validateFlowType(flowTypeStr string) (constants.FlowType, *serviceerror.ServiceError) {
	switch constants.FlowType(flowTypeStr) {
	case constants.FlowTypeAuthentication, constants.FlowTypeRegistration:
		return constants.FlowType(flowTypeStr), nil
	default:
		return "", &constants.ErrorInvalidFlowType
	}
}

// isNewFlow checks if the flow is a new flow based on the provided input.
func isNewFlow(flowID string) bool {
	return flowID == ""
}

// isComplete checks if the flow step status indicates completion.
func isComplete(step model.FlowStep) bool {
	return step.Status == constants.FlowStatusComplete
}

// prepareContext prepares the flow context by merging any data.
func prepareContext(ctx *model.EngineContext, actionID string, inputData map[string]string) {
	// Append any input data present to the context
	if len(inputData) > 0 {
		ctx.UserInputData = sysutils.MergeStringMaps(ctx.UserInputData, inputData)
	}

	if ctx.UserInputData == nil {
		ctx.UserInputData = make(map[string]string)
	}
	if ctx.RuntimeData == nil {
		ctx.RuntimeData = make(map[string]string)
	}

	// Set the action ID if provided
	if actionID != "" {
		ctx.CurrentActionID = actionID
	}
}

// InitiateFlow initiates a new flow with the provided context and returns the flowID without executing the flow.
// This allows external components to pre-initialize a flow with runtime data before actual execution begins.
func (s *flowExecService) InitiateFlow(initContext *model.FlowInitContext) (string, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "FlowExecService"))

	if initContext == nil || initContext.ApplicationID == "" || initContext.FlowType == "" {
		return "", &constants.ErrorInvalidFlowInitContext
	}

	// Validate flow type
	flowType, err := validateFlowType(initContext.FlowType)
	if err != nil {
		return "", err
	}

	// Initialize the engine context
	ctx, err := s.initContext(initContext.ApplicationID, flowType)
	if err != nil {
		logger.Error("Failed to initialize flow context",
			log.String("appID", initContext.ApplicationID),
			log.String("flowType", initContext.FlowType),
			log.String("error", err.Error))
		return "", err
	}

	// Replace the RuntimeData with initContext RuntimeData
	ctx.RuntimeData = initContext.RuntimeData

	// Store the context without executing the flow
	if storeErr := s.storeContext(ctx, logger); storeErr != nil {
		logger.Error("Failed to store initial flow context",
			log.String("flowID", ctx.FlowID),
			log.Error(storeErr))
		return "", &constants.ErrorUpdatingContextInStore
	}

	logger.Debug("Flow initiated successfully", log.String("flowID", ctx.FlowID))
	return ctx.FlowID, nil
}
