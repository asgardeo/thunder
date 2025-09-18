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

// Package flow provides the FlowExecService interface and its implementation.
package flow

import (
	"errors"
	"fmt"

	appconst "github.com/asgardeo/thunder/internal/application/constants"
	appservice "github.com/asgardeo/thunder/internal/application/service"
	"github.com/asgardeo/thunder/internal/flow/constants"
	"github.com/asgardeo/thunder/internal/flow/engine"
	"github.com/asgardeo/thunder/internal/flow/flowmgt"
	"github.com/asgardeo/thunder/internal/flow/model"
	"github.com/asgardeo/thunder/internal/flow/store"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

// FlowExecServiceInterface defines the interface for flow orchestration and acts as the entry point for flow execution
type FlowExecServiceInterface interface {
	Init() error
	Execute(appID, flowID, actionID, flowType string, inputData map[string]string) (
		*model.FlowStep, *serviceerror.ServiceError)
}

// FlowExecService is the implementation of FlowExecServiceInterface
type FlowExecService struct {
	flowStore  store.FlowStoreInterface
	AppService appservice.ApplicationServiceInterface
}

// GetFlowExecService returns a singleton instance of FlowExecService
func GetFlowExecService() FlowExecServiceInterface {
	return &FlowExecService{
		flowStore:  store.NewFlowStore(),
		AppService: appservice.GetApplicationService(),
	}
}

// Init initializes the FlowExecService by loading the necessary components.
func (s *FlowExecService) Init() error {
	flowMgt := flowmgt.GetFlowMgtService()
	if err := flowMgt.Init(); err != nil {
		return errors.New("failed to initialize flow service: " + err.Error())
	}

	if err := validateDefaultFlowConfigs(flowMgt); err != nil {
		return errors.New("default flow config validation failed: " + err.Error())
	}

	return nil
}

// Execute executes a flow with the given data
func (s *FlowExecService) Execute(appID, flowID, actionID, flowType string, inputData map[string]string) (
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
	} else {
		context, loadErr = s.loadPrevContext(flowID, actionID, inputData)
		if loadErr != nil {
			logger.Error("Failed to load previous flow context",
				log.String("flowID", flowID),
				log.String("error", loadErr.Error))
			return nil, loadErr
		}
	}

	flowStep, flowErr := engine.GetFlowEngine().Execute(context)

	if flowErr != nil {
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
func (s *FlowExecService) loadNewContext(appID, actionID, flowTypeStr string,
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
func (s *FlowExecService) initContext(appID string,
	flowType constants.FlowType) (*model.EngineContext, *serviceerror.ServiceError) {
	graphID, svcErr := s.getFlowGraph(appID, flowType)
	if svcErr != nil {
		return nil, svcErr
	}

	ctx := model.EngineContext{}
	flowID := sysutils.GenerateUUID()
	ctx.FlowID = flowID

	flowMgt := flowmgt.GetFlowMgtService()
	graph, ok := flowMgt.GetGraph(graphID)
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
func (s *FlowExecService) loadPrevContext(flowID, actionID string, inputData map[string]string) (
	*model.EngineContext, *serviceerror.ServiceError) {
	ctx, err := s.loadContextFromStore(flowID)
	if err != nil {
		return nil, err
	}

	prepareContext(ctx, actionID, inputData)
	return ctx, nil
}

// loadContextFromStore retrieves the flow context from the store based on the given details.
func (s *FlowExecService) loadContextFromStore(flowID string) (*model.EngineContext,
	*serviceerror.ServiceError) {
	if flowID == "" {
		return nil, &constants.ErrorInvalidFlowID
	}

	dbModel, err := s.flowStore.GetFlowContext(flowID)
	if err != nil {
		return nil, &constants.ErrorUpdatingContextInStore
	}

	if dbModel == nil {
		return nil, &constants.ErrorInvalidFlowID
	}

	flowMgt := flowmgt.GetFlowMgtService()
	graph, exists := flowMgt.GetGraph(dbModel.GraphID)
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
func (s *FlowExecService) setApplicationToContext(ctx *model.EngineContext) *serviceerror.ServiceError {
	app, err := s.AppService.GetApplication(ctx.AppID)
	if err != nil {
		if err.Code == appconst.ErrorApplicationNotFound.Code {
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
func (s *FlowExecService) removeContext(flowID string, logger *log.Logger) error {
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
func (s *FlowExecService) updateContext(ctx *model.EngineContext, flowStep *model.FlowStep, logger *log.Logger) error {
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
func (s *FlowExecService) storeContext(ctx *model.EngineContext, logger *log.Logger) error {
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
func (s *FlowExecService) getFlowGraph(appID string, flowType constants.FlowType) (string,
	*serviceerror.ServiceError) {
	if appID == "" {
		return "", &constants.ErrorInvalidAppID
	}

	app, err := s.AppService.GetApplication(appID)
	if err != nil {
		if err.Code == appconst.ErrorApplicationNotFound.Code {
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

// validateDefaultFlowConfigs validates the default flow configurations.
func validateDefaultFlowConfigs(flowMgt flowmgt.FlowMgtServiceInterface) error {
	flowConfig := config.GetThunderRuntime().Config.Flow

	// Validate auth flow.
	if flowConfig.Authn.DefaultFlow == "" {
		return errors.New("default authentication flow is not configured")
	}
	if !flowMgt.IsValidGraphID(flowConfig.Authn.DefaultFlow) {
		return errors.New("default authentication flow graph ID is invalid")
	}

	return nil
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
