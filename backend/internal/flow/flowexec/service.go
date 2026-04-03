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
	"encoding/json"
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
	"github.com/asgardeo/thunder/internal/system/transaction"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

// FlowExecServiceInterface defines the interface for flow orchestration and acts as the
// entry point for flow execution
type FlowExecServiceInterface interface {
	Execute(ctx context.Context, appID, flowID, flowType string, verbose bool,
		action string, inputs map[string]string) (*FlowStep, *serviceerror.ServiceError)
	InitiateFlow(ctx context.Context, initContext *FlowInitContext) (string, *serviceerror.ServiceError)
	GetFlowRuntimeData(ctx context.Context, flowID string) (map[string]string, *serviceerror.ServiceError)
}

const (
	defaultAuthFlowExpiry           int64 = 1800  // 30 minutes in seconds
	defaultRegistrationFlowExpiry   int64 = 3600  // 60 minutes in seconds
	defaultUserOnboardingFlowExpiry int64 = 86400 // 24 hours in seconds
)

// flowExecService is the implementation of FlowExecServiceInterface
type flowExecService struct {
	flowEngine       flowEngineInterface
	flowMgtService   flowmgt.FlowMgtServiceInterface
	flowStore        flowStoreInterface
	appService       application.ApplicationServiceInterface
	observabilitySvc observability.ObservabilityServiceInterface
	transactioner    transaction.Transactioner
}

func newFlowExecService(flowMgtService flowmgt.FlowMgtServiceInterface,
	flowStore flowStoreInterface, flowEngine flowEngineInterface,
	applicationService application.ApplicationServiceInterface,
	observabilitySvc observability.ObservabilityServiceInterface,
	transactioner transaction.Transactioner) FlowExecServiceInterface {
	return &flowExecService{
		flowMgtService:   flowMgtService,
		flowStore:        flowStore,
		flowEngine:       flowEngine,
		appService:       applicationService,
		observabilitySvc: observabilitySvc,
		transactioner:    transactioner,
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
		context, loadErr = s.loadNewContext(ctx, appID, flowType, verbose, action, inputs, logger)
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
		context, loadErr = s.loadPrevContext(ctx, flowID, action, inputs, logger)
		if loadErr != nil {
			logger.Error("Failed to load previous flow context",
				log.String("flowID", flowID),
				log.String("error", loadErr.Error))
			return nil, loadErr
		}
	}

	// Set trace ID to engine context (request context is already set during context loading)
	context.TraceID = traceID

	flowStep, flowErr := s.flowEngine.Execute(context)

	if flowErr != nil {
		if !isNewFlow(flowID) {
			if removeErr := s.removeContext(ctx, context.FlowID, logger); removeErr != nil {
				logger.Error("Failed to remove flow context after engine failure",
					log.String("flowID", context.FlowID), log.Error(removeErr))
				return nil, &serviceerror.InternalServerError
			}
		}
		return nil, flowErr
	}

	if isComplete(flowStep) {
		if !isNewFlow(flowID) {
			if removeErr := s.removeContext(ctx, context.FlowID, logger); removeErr != nil {
				logger.Error("Failed to remove flow context after completion",
					log.String("flowID", context.FlowID), log.Error(removeErr))
				return nil, &serviceerror.InternalServerError
			}
		}
	} else {
		if isNewFlow(flowID) {
			if storeErr := s.storeContext(ctx, context, logger); storeErr != nil {
				logger.Error("Failed to store initial flow context",
					log.String("flowID", context.FlowID), log.Error(storeErr))
				return nil, &serviceerror.InternalServerError
			}
		} else {
			if updateErr := s.updateContext(ctx, context, &flowStep, logger); updateErr != nil {
				logger.Error("Failed to update flow context", log.String("flowID", context.FlowID),
					log.Error(updateErr))
				return nil, &serviceerror.InternalServerError
			}
		}
	}

	return &flowStep, nil
}

// initContext initializes a new flow context with the given details.
func (s *flowExecService) loadNewContext(ctx context.Context, appID, flowTypeStr string, verbose bool,
	action string, inputs map[string]string, logger *log.Logger) (
	*EngineContext, *serviceerror.ServiceError) {
	flowType, err := validateFlowType(flowTypeStr)
	if err != nil {
		return nil, err
	}

	engineCtx, err := s.initContext(ctx, appID, flowType, verbose, logger)
	if err != nil {
		return nil, err
	}

	prepareContext(engineCtx, action, inputs)
	return engineCtx, nil
}

// initContext initializes a new flow context with the given details.
func (s *flowExecService) initContext(ctx context.Context, appID string, flowType common.FlowType,
	verbose bool, logger *log.Logger) (*EngineContext, *serviceerror.ServiceError) {
	graphID, svcErr := s.getFlowGraph(ctx, appID, flowType, logger)
	if svcErr != nil {
		return nil, svcErr
	}

	engineCtx := EngineContext{}
	flowID, err := sysutils.GenerateUUIDv7()
	if err != nil {
		logger.Error("Failed to generate UUID", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	engineCtx.FlowID = flowID

	graph, svcErr := s.flowMgtService.GetGraph(ctx, graphID)
	if svcErr != nil {
		logger.Error("Error retrieving flow graph from flow management service",
			log.String("graphID", graphID), log.String("error", svcErr.Error))
		return nil, &serviceerror.InternalServerError
	}

	engineCtx.FlowType = graph.GetType()
	engineCtx.Graph = graph
	engineCtx.Context = ctx
	engineCtx.AppID = appID
	engineCtx.Verbose = verbose

	// Set application context if required
	if err := s.setApplicationToContext(&engineCtx, logger); err != nil {
		return nil, err
	}

	return &engineCtx, nil
}

// getFlowExpirySeconds returns the expiry time for a flow in seconds.
func (s *flowExecService) getFlowExpirySeconds(flowType common.FlowType) int64 {
	switch flowType {
	case common.FlowTypeAuthentication:
		return defaultAuthFlowExpiry
	case common.FlowTypeRegistration:
		return defaultRegistrationFlowExpiry
	case common.FlowTypeUserOnboarding:
		return defaultUserOnboardingFlowExpiry
	default:
		// Fallback to auth flow expiry
		return defaultAuthFlowExpiry
	}
}

// loadPrevContext retrieves the flow context from the store based on the given details.
func (s *flowExecService) loadPrevContext(ctx context.Context, flowID, action string,
	inputs map[string]string, logger *log.Logger) (*EngineContext, *serviceerror.ServiceError) {
	engineCtx, err := s.loadContextFromStore(ctx, flowID, logger)
	if err != nil {
		return nil, err
	}

	prepareContext(engineCtx, action, inputs)
	return engineCtx, nil
}

// loadContextFromStore retrieves the flow context from the store based on the given details.
func (s *flowExecService) loadContextFromStore(ctx context.Context, flowID string, logger *log.Logger) (
	*EngineContext, *serviceerror.ServiceError) {
	if flowID == "" {
		return nil, &ErrorInvalidFlowID
	}

	dbModel, err := s.flowStore.GetFlowContext(ctx, flowID)
	if err != nil {
		logger.Error("Error retrieving flow context from store", log.String("flowID", flowID),
			log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	if dbModel == nil {
		return nil, &ErrorInvalidFlowID
	}

	graph, svcErr := s.flowMgtService.GetGraph(ctx, dbModel.GraphID)
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

	// Embed the request context so downstream helpers can read it from engineContext.Context.
	engineContext.Context = ctx

	// Set application context if required
	if err := s.setApplicationToContext(&engineContext, logger); err != nil {
		return nil, err
	}

	return &engineContext, nil
}

// setApplicationToContext retrieves the application and sets it to the flow context.
// It uses the request context stored in engineCtx.Context.
func (s *flowExecService) setApplicationToContext(engineCtx *EngineContext,
	logger *log.Logger) *serviceerror.ServiceError {
	// Skip application loading for app-independent flows
	if engineCtx.FlowType == common.FlowTypeUserOnboarding {
		return nil
	}

	app, err := s.appService.GetApplication(engineCtx.Context, engineCtx.AppID)
	if err != nil {
		if err.Code == application.ErrorApplicationNotFound.Code {
			return &ErrorInvalidAppID
		}
		if err.Type == serviceerror.ClientErrorType {
			svcErr := &ErrorApplicationRetrievalClientError
			svcErr.ErrorDescription = fmt.Sprintf("Error while retrieving application: %s", err.ErrorDescription)
			return svcErr
		}

		logger.Error("Server error while retrieving application", log.String("appID", engineCtx.AppID),
			log.String("errorCode", err.Code), log.String("errorDescription", err.ErrorDescription))
		return &serviceerror.InternalServerError
	}
	if app == nil {
		logger.Error("Application not found while setting to flow context", log.String("appID", engineCtx.AppID))
		return &serviceerror.InternalServerError
	}
	engineCtx.Application = *app
	return nil
}

// removeContext removes the flow context from the store.
func (s *flowExecService) removeContext(ctx context.Context, flowID string, logger *log.Logger) error {
	if flowID == "" {
		return fmt.Errorf("flow ID cannot be empty")
	}

	txErr := s.transactioner.Transact(ctx, func(txCtx context.Context) error {
		return s.flowStore.DeleteFlowContext(txCtx, flowID)
	})
	if txErr != nil {
		return fmt.Errorf("failed to remove flow context from database: %w", txErr)
	}

	logger.Debug("Flow context removed successfully from database", log.String("flowID", flowID))
	return nil
}

// updateContext updates the flow context in the store based on the flow step status.
func (s *flowExecService) updateContext(ctx context.Context, engineCtx *EngineContext,
	flowStep *FlowStep, logger *log.Logger) error {
	if flowStep.Status == common.FlowStatusComplete {
		return s.removeContext(ctx, engineCtx.FlowID, logger)
	} else {
		logger.Debug("Flow execution is incomplete, updating the flow context",
			log.String("flowID", engineCtx.FlowID))

		if engineCtx.FlowID == "" {
			return fmt.Errorf("flow ID cannot be empty")
		}

		txErr := s.transactioner.Transact(ctx, func(txCtx context.Context) error {
			return s.flowStore.UpdateFlowContext(txCtx, *engineCtx)
		})
		if txErr != nil {
			return fmt.Errorf("failed to update flow context in database: %w", txErr)
		}

		logger.Debug("Flow context updated successfully in database",
			log.String("flowID", engineCtx.FlowID))
		return nil
	}
}

// storeContext stores the flow context in the store.
func (s *flowExecService) storeContext(ctx context.Context, engineCtx *EngineContext,
	logger *log.Logger) error {
	if engineCtx.FlowID == "" {
		return fmt.Errorf("flow ID cannot be empty")
	}

	expirySeconds := s.getFlowExpirySeconds(engineCtx.FlowType)

	txErr := s.transactioner.Transact(ctx, func(txCtx context.Context) error {
		return s.flowStore.StoreFlowContext(txCtx, *engineCtx, expirySeconds)
	})
	if txErr != nil {
		return fmt.Errorf("failed to store flow context in database: %w", txErr)
	}

	logger.Debug("Flow context stored successfully in database", log.String("flowID", engineCtx.FlowID))
	return nil
}

// getFlowGraph checks if the provided application ID is valid and returns the associated flow ID.
func (s *flowExecService) getFlowGraph(ctx context.Context, appID string, flowType common.FlowType,
	logger *log.Logger) (string, *serviceerror.ServiceError) {
	// Handle app-independent system flows
	if flowType == common.FlowTypeUserOnboarding {
		return s.getSystemFlowGraph(ctx, flowType, logger)
	}

	if appID == "" {
		return "", &ErrorInvalidAppID
	}

	app, err := s.appService.GetApplication(ctx, appID)
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
func (s *flowExecService) getSystemFlowGraph(ctx context.Context, flowType common.FlowType,
	logger *log.Logger) (string, *serviceerror.ServiceError) {
	handle := ""
	switch flowType {
	case common.FlowTypeUserOnboarding:
		handle = config.GetThunderRuntime().Config.Flow.UserOnboardingFlowHandle
	default:
		return "", &ErrorInvalidFlowType
	}

	flow, err := s.flowMgtService.GetFlowByHandle(ctx, handle, flowType)
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
func (s *flowExecService) InitiateFlow(ctx context.Context,
	initContext *FlowInitContext) (string, *serviceerror.ServiceError) {
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
	engineCtx, err := s.initContext(ctx, initContext.ApplicationID, flowType, true, logger)
	if err != nil {
		logger.Error("Failed to initialize flow context",
			log.String("appID", initContext.ApplicationID),
			log.String("flowType", initContext.FlowType),
			log.String("error", err.Error))
		return "", err
	}

	// Replace the RuntimeData with initContext RuntimeData
	engineCtx.RuntimeData = initContext.RuntimeData

	// Store the context without executing the flow
	if storeErr := s.storeContext(ctx, engineCtx, logger); storeErr != nil {
		logger.Error("Failed to store initial flow context",
			log.String("flowID", engineCtx.FlowID),
			log.Error(storeErr))
		return "", &serviceerror.InternalServerError
	}

	logger.Debug("Flow initiated successfully", log.String("flowID", engineCtx.FlowID))
	return engineCtx.FlowID, nil
}

// GetFlowRuntimeData returns the RuntimeData map for the given flow ID.
// Returns nil, nil if the flow context does not exist.
func (s *flowExecService) GetFlowRuntimeData(
	ctx context.Context, flowID string) (map[string]string, *serviceerror.ServiceError) {
	flowCtx, err := s.flowStore.GetFlowContext(ctx, flowID)
	if err != nil {
		return nil, &serviceerror.InternalServerError
	}
	if flowCtx == nil || flowCtx.RuntimeData == nil {
		return nil, nil
	}
	var runtimeData map[string]string
	if jsonErr := json.Unmarshal([]byte(*flowCtx.RuntimeData), &runtimeData); jsonErr != nil {
		return nil, &serviceerror.InternalServerError
	}
	return runtimeData, nil
}
