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
	"context"
	"fmt"

	"github.com/asgardeo/thunder/internal/application"
	"github.com/asgardeo/thunder/internal/flow/common"
	flowmgt "github.com/asgardeo/thunder/internal/flow/mgt"

	"github.com/asgardeo/thunder/internal/system/config"
	sysContext "github.com/asgardeo/thunder/internal/system/context"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/observability"
	"github.com/asgardeo/thunder/internal/system/observability/event"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

// FlowExecServiceInterface defines the interface for flow orchestration and acts as the
// entry point for flow execution
type FlowExecServiceInterface interface {
	Execute(ctx context.Context, appID, flowID, flowType string, verbose bool,
		action string, inputs map[string]string) (*FlowStep, *serviceerror.ServiceError)
	InitiateFlow(initContext *FlowInitContext) (string, *serviceerror.ServiceError)
}

// flowExecService is the implementation of FlowExecServiceInterface
type flowExecService struct {
	flowEngine       flowEngineInterface
	flowMgtService   flowmgt.FlowMgtServiceInterface
	flowStore        flowStoreInterface
	appService       application.ApplicationServiceInterface
	observabilitySvc observability.ObservabilityServiceInterface
}

func newFlowExecService(flowMgtService flowmgt.FlowMgtServiceInterface,
	flowStore flowStoreInterface, flowEngine flowEngineInterface,
	applicationService application.ApplicationServiceInterface,
	observabilitySvc observability.ObservabilityServiceInterface) FlowExecServiceInterface {
	return &flowExecService{
		flowMgtService:   flowMgtService,
		flowStore:        flowStore,
		flowEngine:       flowEngine,
		appService:       applicationService,
		observabilitySvc: observabilitySvc,
	}
}

// Execute executes a flow with the given data
func (s *flowExecService) Execute(ctx context.Context,
	appID, flowID, flowType string, verbose bool, action string, inputs map[string]string) (
	*FlowStep, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "FlowExecService"))

	// Get trace ID from context
	traceID := sysContext.GetTraceID(ctx)

	var context *EngineContext
	var loadErr *serviceerror.ServiceError

	if isNewFlow(flowID) {
		context, loadErr = s.loadNewContext(appID, flowType, verbose, action, inputs, logger)
		if loadErr != nil {
			logger.Error("Failed to load new flow context",
				log.String("appID", appID),
				log.String("flowType", flowType),
				log.String("error", loadErr.Error))

			if s.observabilitySvc.IsEnabled() {
				evt := event.NewEvent(
					traceID,
					string(event.EventTypeFlowFailed),
					event.ComponentFlowEngine,
				).
					WithStatus(event.StatusFailure).
					WithData(event.DataKey.AppID, appID).
					WithData(event.DataKey.FlowType, flowType).
					WithData(event.DataKey.Error, loadErr.Error).
					WithData(event.DataKey.ErrorCode, loadErr.Code).
					WithData(event.DataKey.ErrorType, string(loadErr.Type))

				if loadErr.ErrorDescription != "" {
					evt.WithData(event.DataKey.Message, loadErr.ErrorDescription)
				}
				s.observabilitySvc.PublishEvent(evt)
			}
			return nil, loadErr
		}
	} else {
		context, loadErr = s.loadPrevContext(flowID, action, inputs, logger)
		if loadErr != nil {
			logger.Error("Failed to load previous flow context",
				log.String("flowID", flowID),
				log.String("error", loadErr.Error))
			return nil, loadErr
		}
	}

	// Set trace ID and HTTP context to context
	context.TraceID = traceID
	context.HTTPContext = ctx

	flowStep, flowErr := s.flowEngine.Execute(context)

	if flowErr != nil {
		if !isNewFlow(flowID) {
			if removeErr := s.removeContext(context.FlowID, logger); removeErr != nil {
				logger.Error("Failed to remove flow context after engine failure",
					log.String("flowID", context.FlowID), log.Error(removeErr))
				return nil, &serviceerror.InternalServerError
			}
		}
		return nil, flowErr
	}

	if isComplete(flowStep) {
		if !isNewFlow(flowID) {
			if removeErr := s.removeContext(context.FlowID, logger); removeErr != nil {
				logger.Error("Failed to remove flow context after completion",
					log.String("flowID", context.FlowID), log.Error(removeErr))
				return nil, &serviceerror.InternalServerError
			}
		}
	} else {
		if isNewFlow(flowID) {
			if storeErr := s.storeContext(context, logger); storeErr != nil {
				logger.Error("Failed to store initial flow context",
					log.String("flowID", context.FlowID), log.Error(storeErr))
				return nil, &serviceerror.InternalServerError
			}
		} else {
			if updateErr := s.updateContext(context, &flowStep, logger); updateErr != nil {
				logger.Error("Failed to update flow context", log.String("flowID", context.FlowID),
					log.Error(updateErr))
				return nil, &serviceerror.InternalServerError
			}
		}
	}

	return &flowStep, nil
}

// initContext initializes a new flow context with the given details.
func (s *flowExecService) loadNewContext(appID, flowTypeStr string, verbose bool,
	action string, inputs map[string]string, logger *log.Logger) (
	*EngineContext, *serviceerror.ServiceError) {
	flowType, err := validateFlowType(flowTypeStr)
	if err != nil {
		return nil, err
	}

	ctx, err := s.initContext(appID, flowType, verbose, logger)
	if err != nil {
		return nil, err
	}

	prepareContext(ctx, action, inputs)
	return ctx, nil
}

// initContext initializes a new flow context with the given details.
func (s *flowExecService) initContext(appID string, flowType common.FlowType,
	verbose bool, logger *log.Logger) (*EngineContext, *serviceerror.ServiceError) {
	graphID, svcErr := s.getFlowGraph(appID, flowType, logger)
	if svcErr != nil {
		return nil, svcErr
	}

	ctx := EngineContext{}
	flowID, err := sysutils.GenerateUUIDv7()
	if err != nil {
		logger.Error("Failed to generate UUID", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	ctx.FlowID = flowID

	graph, svcErr := s.flowMgtService.GetGraph(graphID)
	if svcErr != nil {
		logger.Error("Error retrieving flow graph from flow management service",
			log.String("graphID", graphID), log.String("error", svcErr.Error))
		return nil, &serviceerror.InternalServerError
	}

	ctx.FlowType = graph.GetType()
	ctx.Graph = graph
	ctx.AppID = appID
	ctx.Verbose = verbose

	// Set application context if required
	if err := s.setApplicationToContext(&ctx, logger); err != nil {
		return nil, err
	}

	return &ctx, nil
}

// loadPrevContext retrieves the flow context from the store based on the given details.
func (s *flowExecService) loadPrevContext(flowID, action string, inputs map[string]string,
	logger *log.Logger) (*EngineContext, *serviceerror.ServiceError) {
	ctx, err := s.loadContextFromStore(flowID, logger)
	if err != nil {
		return nil, err
	}

	prepareContext(ctx, action, inputs)
	return ctx, nil
}

// loadContextFromStore retrieves the flow context from the store based on the given details.
func (s *flowExecService) loadContextFromStore(flowID string, logger *log.Logger) (
	*EngineContext, *serviceerror.ServiceError) {
	if flowID == "" {
		return nil, &ErrorInvalidFlowID
	}

	dbModel, err := s.flowStore.GetFlowContext(flowID)
	if err != nil {
		logger.Error("Error retrieving flow context from store", log.String("flowID", flowID),
			log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	if dbModel == nil {
		return nil, &ErrorInvalidFlowID
	}

	graph, svcErr := s.flowMgtService.GetGraph(dbModel.GraphID)
	if svcErr != nil {
		logger.Error("Error retrieving flow graph from flow management service",
			log.String("graphID", dbModel.GraphID), log.String("error", svcErr.Error))
		return nil, &serviceerror.InternalServerError
	}

	engineContext, err := dbModel.ToEngineContext(graph)
	if err != nil {
		logger.Error("Failed to convert flow context from database format",
			log.String("flowID", flowID), log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	// Set application context if required
	if err := s.setApplicationToContext(&engineContext, logger); err != nil {
		return nil, err
	}

	return &engineContext, nil
}

// setApplicationToContext retrieves the application and sets it to the flow context.
func (s *flowExecService) setApplicationToContext(ctx *EngineContext,
	logger *log.Logger) *serviceerror.ServiceError {
	// Skip application loading for app-independent flows
	if ctx.FlowType == common.FlowTypeUserOnboarding {
		return nil
	}

	app, err := s.appService.GetApplication(ctx.AppID)
	if err != nil {
		if err.Code == application.ErrorApplicationNotFound.Code {
			return &ErrorInvalidAppID
		}
		if err.Type == serviceerror.ClientErrorType {
			svcErr := &ErrorApplicationRetrievalClientError
			svcErr.ErrorDescription = fmt.Sprintf("Error while retrieving application: %s", err.ErrorDescription)
			return svcErr
		}

		logger.Error("Server error while retrieving application", log.String("appID", ctx.AppID),
			log.String("errorCode", err.Code), log.String("errorDescription", err.ErrorDescription))
		return &serviceerror.InternalServerError
	}
	if app == nil {
		logger.Error("Application not found while setting to flow context", log.String("appID", ctx.AppID))
		return &serviceerror.InternalServerError
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
func (s *flowExecService) updateContext(ctx *EngineContext, flowStep *FlowStep, logger *log.Logger) error {
	if flowStep.Status == common.FlowStatusComplete {
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
func (s *flowExecService) storeContext(ctx *EngineContext, logger *log.Logger) error {
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

// getFlowGraph checks if the provided application ID is valid and returns the associated flow ID.
func (s *flowExecService) getFlowGraph(appID string, flowType common.FlowType,
	logger *log.Logger) (string, *serviceerror.ServiceError) {
	// Handle app-independent system flows
	if flowType == common.FlowTypeUserOnboarding {
		return s.getSystemFlowGraph(flowType, logger)
	}

	if appID == "" {
		return "", &ErrorInvalidAppID
	}

	app, err := s.appService.GetApplication(appID)
	if err != nil {
		if err.Code == application.ErrorApplicationNotFound.Code {
			return "", &ErrorInvalidAppID
		}
		if err.Type == serviceerror.ClientErrorType {
			return "", &ErrorApplicationRetrievalClientError
		}

		logger.Error("Server error while retrieving application", log.String("appID", appID),
			log.String("errorCode", err.Code), log.String("errorDescription", err.ErrorDescription))
		return "", &serviceerror.InternalServerError
	}
	if app == nil {
		return "", &ErrorInvalidAppID
	}

	if flowType == common.FlowTypeRegistration {
		if !app.IsRegistrationFlowEnabled {
			return "", &ErrorRegistrationFlowDisabled
		} else if app.RegistrationFlowID == "" {
			logger.Error("Registration flow is not configured for the application",
				log.String("appID", appID))
			return "", &serviceerror.InternalServerError
		}
		return app.RegistrationFlowID, nil
	}

	// Default to authentication flow ID
	if app.AuthFlowID == "" {
		logger.Error("Authentication flow is not configured for the application",
			log.String("appID", appID))
		return "", &serviceerror.InternalServerError
	}

	return app.AuthFlowID, nil
}

// validateFlowType validates the provided flow type string and returns the corresponding FlowType.
func validateFlowType(flowTypeStr string) (common.FlowType, *serviceerror.ServiceError) {
	switch common.FlowType(flowTypeStr) {
	case common.FlowTypeAuthentication, common.FlowTypeRegistration, common.FlowTypeUserOnboarding:
		return common.FlowType(flowTypeStr), nil
	default:
		return "", &ErrorInvalidFlowType
	}
}

// isNewFlow checks if the flow is a new flow based on the provided input.
func isNewFlow(flowID string) bool {
	return flowID == ""
}

// getSystemFlowGraph retrieves the flow graph for system flows by handle.
func (s *flowExecService) getSystemFlowGraph(flowType common.FlowType,
	logger *log.Logger) (string, *serviceerror.ServiceError) {
	handle := ""
	switch flowType {
	case common.FlowTypeUserOnboarding:
		handle = config.GetThunderRuntime().Config.Flow.UserOnboardingFlowHandle
	default:
		return "", &ErrorInvalidFlowType
	}

	flow, err := s.flowMgtService.GetFlowByHandle(handle, flowType)
	if err != nil {
		logger.Error("Failed to get system flow by handle",
			log.String("handle", handle), log.String("flowType", string(flowType)))
		return "", err
	}
	return flow.ID, nil
}

// isComplete checks if the flow step status indicates completion.
func isComplete(step FlowStep) bool {
	return step.Status == common.FlowStatusComplete
}

// prepareContext prepares the flow context by merging any data.
func prepareContext(ctx *EngineContext, action string, inputs map[string]string) {
	// Append any inputs present to the context
	if len(inputs) > 0 {
		ctx.UserInputs = sysutils.MergeStringMaps(ctx.UserInputs, inputs)
	}

	if ctx.UserInputs == nil {
		ctx.UserInputs = make(map[string]string)
	}
	if ctx.RuntimeData == nil {
		ctx.RuntimeData = make(map[string]string)
	}

	// Set the action if provided
	if action != "" {
		ctx.CurrentAction = action
	}
}

// InitiateFlow initiates a new flow with the provided context and returns the flowID without executing the flow.
// This allows external components to pre-initialize a flow with runtime data before actual execution begins.
func (s *flowExecService) InitiateFlow(initContext *FlowInitContext) (string, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "FlowExecService"))

	if initContext == nil || initContext.FlowType == "" {
		return "", &ErrorInvalidFlowInitContext
	}

	// Validate flow type
	flowType, err := validateFlowType(initContext.FlowType)
	if err != nil {
		return "", err
	}

	// Application ID is required for all flows except Invite Registration
	if flowType != common.FlowTypeUserOnboarding && initContext.ApplicationID == "" {
		return "", &ErrorInvalidFlowInitContext
	}

	// Initialize the engine context
	// This uses verbose true to ensure step layouts are returned during execution
	ctx, err := s.initContext(initContext.ApplicationID, flowType, true, logger)
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
		return "", &serviceerror.InternalServerError
	}

	logger.Debug("Flow initiated successfully", log.String("flowID", ctx.FlowID))
	return ctx.FlowID, nil
}
