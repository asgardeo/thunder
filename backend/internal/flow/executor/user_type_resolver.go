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

package executor

import (
	"context"
	"fmt"
	"slices"

	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/userschema"
)

const (
	userTypeResolverLoggerComponentName = "UserTypeResolver"
)

// schemaWithOU represents a user schema along with its associated organization unit ID.
type schemaWithOU struct {
	userSchema *userschema.UserSchema
	ouID       string
}

// userTypeResolver is a registration-flow executor that resolves the user type at flow start.
type userTypeResolver struct {
	core.ExecutorInterface
	userSchemaService userschema.UserSchemaServiceInterface
	logger            *log.Logger
}

var _ core.ExecutorInterface = (*userTypeResolver)(nil)

// newUserTypeResolver creates a new instance of the UserTypeResolver executor.
func newUserTypeResolver(
	flowFactory core.FlowFactoryInterface,
	userSchemaService userschema.UserSchemaServiceInterface,
) *userTypeResolver {
	logger := log.GetLogger().With(
		log.String(log.LoggerKeyComponentName, userTypeResolverLoggerComponentName),
		log.String(log.LoggerKeyExecutorName, ExecutorNameUserTypeResolver))

	defaultInputs := []common.Input{
		{
			Ref:        "usertype_input",
			Identifier: userTypeKey,
			Type:       "SELECT",
			Required:   true,
		},
	}

	base := flowFactory.CreateExecutor(ExecutorNameUserTypeResolver, common.ExecutorTypeRegistration,
		defaultInputs, []common.Input{})

	return &userTypeResolver{
		ExecutorInterface: base,
		userSchemaService: userSchemaService,
		logger:            logger,
	}
}

// Execute resolves the user type from inputs or prompts the user to select one.
func (u *userTypeResolver) Execute(ctx *core.NodeContext) (*common.ExecutorResponse, error) {
	logger := u.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Executing user type resolver")

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
		ForwardedData:  make(map[string]interface{}),
	}

	switch ctx.FlowType {
	case common.FlowTypeAuthentication:
		return u.handleAuthenticationFlows(ctx, execResp)
	case common.FlowTypeRegistration:
		return u.handleRegistrationFlows(ctx, execResp)
	case common.FlowTypeUserOnboarding:
		return u.handleUserOnboardingFlows(ctx, execResp)
	default:
		logger.Debug("User type resolver is not applicable for the flow type",
			log.String("flowType", string(ctx.FlowType)))
		execResp.Status = common.ExecComplete
		return execResp, nil
	}
}

// handleAuthenticationFlows handles user type resolution for authentication flows.
func (u *userTypeResolver) handleAuthenticationFlows(ctx *core.NodeContext, execResp *common.ExecutorResponse) (
	*common.ExecutorResponse, error) {
	logger := u.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))

	// Validate that allowed user types are defined
	if len(ctx.Application.AllowedUserTypes) == 0 {
		logger.Debug("No allowed user types configured for authentication")
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "Authentication not available for this application"
		return execResp, nil
	}

	execResp.Status = common.ExecComplete
	return execResp, nil
}

// handleRegistrationFlows handles user type resolution for registration flows.
func (u *userTypeResolver) handleRegistrationFlows(ctx *core.NodeContext, execResp *common.ExecutorResponse) (
	*common.ExecutorResponse, error) {
	logger := u.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	allowed := ctx.Application.AllowedUserTypes

	// Check for allowed user types to decide next steps
	if len(allowed) == 0 {
		// TODO: This should be improved to fallback to the application's ou when the support is available.
		//  userType has an attached ou. Need to find userType from the application's ou.
		//  Also should check if self registration is enabled for the user type when the support is available.

		logger.Debug("No allowed user types found for the application")
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "Self-registration not available for this application"
		return execResp, nil
	}

	// Check if userType is provided in inputs
	if u.HasRequiredInputs(ctx, execResp) {
		err := u.resolveUserTypeFromInput(context.TODO(), execResp, ctx.UserInputs[userTypeKey], allowed)
		return execResp, err
	}

	// If only one allowed user type, select it automatically
	if len(allowed) == 1 {
		err := u.resolveUserTypeFromSingleAllowed(context.TODO(), execResp, allowed[0])
		return execResp, err
	}

	// If multiple allowed user types, prompt the user to select one
	err := u.resolveUserTypeFromMultipleAllowed(context.TODO(), execResp, allowed)

	return execResp, err
}

// handleUserOnboardingFlows handles user type resolution for user onboarding flows.
func (u *userTypeResolver) handleUserOnboardingFlows(ctx *core.NodeContext,
	execResp *common.ExecutorResponse) (*common.ExecutorResponse, error) {
	logger := u.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))

	// If userType already provided, validate and set runtime data
	if userType, ok := ctx.UserInputs[userTypeKey]; ok && userType != "" {
		userSchema, ouID, err := u.getUserSchemaAndOU(context.TODO(), userType)
		if err != nil {
			execResp.Status = common.ExecFailure
			execResp.FailureReason = "Invalid user type"
			return execResp, nil
		}

		execResp.RuntimeData[userTypeKey] = userType
		execResp.RuntimeData[defaultOUIDKey] = ouID
		logger.Debug("User type resolved for user onboarding", log.String(userTypeKey, userType),
			log.String(ouIDKey, userSchema.OrganizationUnitID))
		execResp.Status = common.ExecComplete
		return execResp, nil
	}

	// List all available user schemas
	schemas, svcErr := u.userSchemaService.GetUserSchemaList(context.TODO(), 100, 0)
	if svcErr != nil {
		logger.Debug("Failed to list user schemas", log.String("error", svcErr.Error))
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "Failed to retrieve user types"
		return execResp, nil
	}

	if len(schemas.Schemas) == 0 {
		logger.Debug("No user schemas available")
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "No user types available"
		return execResp, nil
	}

	options := make([]string, 0, len(schemas.Schemas))
	for _, schema := range schemas.Schemas {
		options = append(options, schema.Name)
	}

	u.promptUserSelection(execResp, options)
	return execResp, nil
}

// resolveUserTypeFromInput resolves the user type from input and updates the executor response.
func (u *userTypeResolver) resolveUserTypeFromInput(ctx context.Context, execResp *common.ExecutorResponse,
	userType string, allowed []string) error {
	logger := u.logger
	if slices.Contains(allowed, userType) {
		logger.Debug("User type resolved from input", log.String(userTypeKey, userType))

		userSchema, ouID, err := u.getUserSchemaAndOU(ctx, userType)
		if err != nil {
			return err
		}
		if !userSchema.AllowSelfRegistration {
			logger.Debug("Self registration not enabled for user type", log.String(userTypeKey, userType))
			execResp.Status = common.ExecFailure
			execResp.FailureReason = "Self-registration not enabled for the user type"
			return nil
		}

		// Add userType and ouID to runtime data
		execResp.RuntimeData[userTypeKey] = userType
		execResp.RuntimeData[defaultOUIDKey] = ouID

		execResp.Status = common.ExecComplete
		return nil
	}

	execResp.Status = common.ExecFailure
	execResp.FailureReason = "Application does not allow registration for the user type"
	return nil
}

// resolveUserTypeFromSingleAllowed resolves the user type when there is only a single allowed user type.
func (u *userTypeResolver) resolveUserTypeFromSingleAllowed(ctx context.Context, execResp *common.ExecutorResponse,
	allowedUserType string) error {
	logger := u.logger
	userSchema, ouID, err := u.getUserSchemaAndOU(ctx, allowedUserType)
	if err != nil {
		return err
	}

	if !userSchema.AllowSelfRegistration {
		logger.Debug("Self registration not enabled for user type", log.String(userTypeKey, allowedUserType))
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "Self-registration not enabled for the user type"
		return nil
	}

	logger.Debug("User type resolved from allowed list", log.String(userTypeKey, allowedUserType))

	// Add userType and ouID to runtime data
	execResp.RuntimeData[userTypeKey] = allowedUserType
	execResp.RuntimeData[defaultOUIDKey] = ouID

	execResp.Status = common.ExecComplete
	return nil
}

// resolveUserTypeFromMultipleAllowed resolves the user type when multiple allowed user types exist.
func (u *userTypeResolver) resolveUserTypeFromMultipleAllowed(ctx context.Context, execResp *common.ExecutorResponse,
	allowed []string) error {
	logger := u.logger

	// Filter self registration enabled user types
	selfRegEnabledUserTypes := make([]schemaWithOU, 0)
	for _, userType := range allowed {
		userSchema, ouID, err := u.getUserSchemaAndOU(ctx, userType)
		if err != nil {
			return err
		}
		if userSchema.AllowSelfRegistration {
			selfRegEnabledUserTypes = append(selfRegEnabledUserTypes, schemaWithOU{
				userSchema: userSchema,
				ouID:       ouID,
			})
		}
	}

	// Fail if no user types have self registration enabled
	if len(selfRegEnabledUserTypes) == 0 {
		logger.Debug("No user types with self registration enabled")
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "Self-registration not available for this application"
		return nil
	}

	// If only one user type has self registration enabled, select it automatically
	if len(selfRegEnabledUserTypes) == 1 {
		record := selfRegEnabledUserTypes[0]
		logger.Debug("User type auto-selected", log.String(userTypeKey, record.userSchema.Name))

		// Add userType and ouID to runtime data
		execResp.RuntimeData[userTypeKey] = record.userSchema.Name
		execResp.RuntimeData[defaultOUIDKey] = record.ouID

		execResp.Status = common.ExecComplete
		return nil
	}

	// If multiple user types are allowed, prompt the user to select one
	selfRegUserTypes := make([]string, 0, len(selfRegEnabledUserTypes))
	for _, record := range selfRegEnabledUserTypes {
		selfRegUserTypes = append(selfRegUserTypes, record.userSchema.Name)
	}

	logger.Debug("Prompting for user type selection as multiple user types are available for self registration",
		log.Any("userTypes", selfRegUserTypes))

	u.promptUserSelection(execResp, selfRegUserTypes)
	return nil
}

// getUserSchemaAndOU retrieves the user schema by name and returns the schema and organization unit ID.
func (u *userTypeResolver) getUserSchemaAndOU(
	ctx context.Context, userType string,
) (*userschema.UserSchema, string, error) {
	logger := u.logger.With(log.String(userTypeKey, userType))

	userSchema, svcErr := u.userSchemaService.GetUserSchemaByName(ctx, userType)
	if svcErr != nil {
		logger.Error("Failed to resolve user schema for user type",
			log.String(userTypeKey, userType), log.String("error", svcErr.Error))
		return nil, "", fmt.Errorf("failed to resolve user schema for user type: %s", userType)
	}

	if userSchema.OrganizationUnitID == "" {
		logger.Error("No organization unit found for user type", log.String(userTypeKey, userType))
		return nil, "", fmt.Errorf("no organization unit found for user type: %s", userType)
	}

	logger.Debug("User schema resolved for user type", log.String(userTypeKey, userType),
		log.String(ouIDKey, userSchema.OrganizationUnitID))
	return userSchema, userSchema.OrganizationUnitID, nil
}

// promptUserSelection prompts the user to select a user type from the provided options.
func (u *userTypeResolver) promptUserSelection(execResp *common.ExecutorResponse, options []string) {
	u.logger.Debug("Prompting user for user type selection", log.Any("userTypes", options))

	execResp.Status = common.ExecUserInputRequired

	// Use the default input configuration
	inputs := u.GetDefaultInputs()
	if len(inputs) > 0 {
		input := inputs[0]
		input.Options = options
		execResp.Inputs = []common.Input{input}

		// Forward the input with options to the next node
		execResp.ForwardedData[common.ForwardedDataKeyInputs] = execResp.Inputs
	}
}
