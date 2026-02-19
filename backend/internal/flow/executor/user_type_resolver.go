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
	"fmt"
	"slices"

	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/usertype"
)

const (
	userTypeResolverLoggerComponentName = "UserTypeResolver"
)

// schemaWithOU represents a user type along with its associated organization unit ID.
type schemaWithOU struct {
	userType *usertype.UserType
	ouID     string
}

// userTypeResolver is a registration-flow executor that resolves the user type at flow start.
type userTypeResolver struct {
	core.ExecutorInterface
	userTypeService usertype.UserTypeServiceInterface
	logger          *log.Logger
}

var _ core.ExecutorInterface = (*userTypeResolver)(nil)

// newUserTypeResolver creates a new instance of the UserTypeResolver executor.
func newUserTypeResolver(
	flowFactory core.FlowFactoryInterface,
	userTypeService usertype.UserTypeServiceInterface,
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
		userTypeService:   userTypeService,
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
	}

	allowed := ctx.Application.AllowedUserTypes

	if ctx.FlowType == common.FlowTypeAuthentication {
		// For authentication flows, validate that allowed user types are defined
		if len(allowed) == 0 {
			logger.Debug("No allowed user types configured for authentication")
			execResp.Status = common.ExecFailure
			execResp.FailureReason = "Authentication not available for this application"
			return execResp, nil
		}

		execResp.Status = common.ExecComplete
		return execResp, nil
	} else if ctx.FlowType == common.FlowTypeUserOnboarding {
		return u.resolveForUserOnboarding(ctx, execResp)
	} else if ctx.FlowType != common.FlowTypeRegistration {
		logger.Debug("User type resolver is only applicable for registration, user onboarding and authentication flows")
		execResp.Status = common.ExecComplete
		return execResp, nil
	}

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
		err := u.resolveUserTypeFromInput(execResp, ctx.UserInputs[userTypeKey], allowed)
		return execResp, err
	}

	if len(allowed) == 1 {
		err := u.resolveUserTypeFromSingleAllowed(execResp, allowed[0])
		return execResp, err
	}

	err := u.resolveUserTypeFromMultipleAllowed(execResp, allowed)
	return execResp, err
}

// GetDefaultMeta returns the default meta structure for the user type selection page.
func (u *userTypeResolver) GetDefaultMeta() interface{} {
	return core.NewMetaBuilder().
		WithIDPrefix("usertype").
		WithHeading("{{ t(signup:heading) }}").
		WithInput(u.GetDefaultInputs()[0], core.MetaInputConfig{
			Label:       "{{ t(elements:fields.usertype.label) }}",
			Placeholder: "{{ t(elements:fields.usertype.placeholder) }}",
		}).
		WithSubmitButton("{{ t(elements:buttons.submit.text) }}").
		Build()
}

// resolveUserTypeFromInput resolves the user type from input and updates the executor response.
func (u *userTypeResolver) resolveUserTypeFromInput(execResp *common.ExecutorResponse,
	userTypeName string, allowed []string) error {
	logger := u.logger
	if slices.Contains(allowed, userTypeName) {
		logger.Debug("User type resolved from input", log.String(userTypeKey, userTypeName))

		userType, ouID, err := u.getUserTypeAndOU(userTypeName)
		if err != nil {
			return err
		}
		if !userType.AllowSelfRegistration {
			logger.Debug("Self registration not enabled for user type", log.String(userTypeKey, userType.Name))
			execResp.Status = common.ExecFailure
			execResp.FailureReason = "Self-registration not enabled for the user type"
			return nil
		}

		// Add userType and ouID to runtime data
		execResp.RuntimeData[userTypeKey] = userType.Name
		execResp.RuntimeData[defaultOUIDKey] = ouID

		execResp.Status = common.ExecComplete
		return nil
	}

	execResp.Status = common.ExecFailure
	execResp.FailureReason = "Application does not allow registration for the user type"
	return nil
}

// resolveUserTypeFromSingleAllowed resolves the user type when there is only a single allowed user type.
func (u *userTypeResolver) resolveUserTypeFromSingleAllowed(execResp *common.ExecutorResponse,
	allowedUserType string) error {
	logger := u.logger
	userType, ouID, err := u.getUserTypeAndOU(allowedUserType)
	if err != nil {
		return err
	}

	if !userType.AllowSelfRegistration {
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
func (u *userTypeResolver) resolveUserTypeFromMultipleAllowed(execResp *common.ExecutorResponse,
	allowed []string) error {
	logger := u.logger

	// Filter self registration enabled user types
	selfRegEnabledUserTypes := make([]schemaWithOU, 0)
	for _, userType := range allowed {
		userType, ouID, err := u.getUserTypeAndOU(userType)
		if err != nil {
			return err
		}
		if userType.AllowSelfRegistration {
			selfRegEnabledUserTypes = append(selfRegEnabledUserTypes, schemaWithOU{
				userType: userType,
				ouID:     ouID,
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
		logger.Debug("User type auto-selected", log.String(userTypeKey, record.userType.Name))

		// Add userType and ouID to runtime data
		execResp.RuntimeData[userTypeKey] = record.userType.Name
		execResp.RuntimeData[defaultOUIDKey] = record.ouID

		execResp.Status = common.ExecComplete
		return nil
	}

	// If multiple user types are allowed, prompt the user to select one
	selfRegUserTypes := make([]string, 0, len(selfRegEnabledUserTypes))
	for _, record := range selfRegEnabledUserTypes {
		selfRegUserTypes = append(selfRegUserTypes, record.userType.Name)
	}

	logger.Debug("Prompting for user type selection as multiple user types are available for self registration",
		log.Any("userTypes", selfRegUserTypes))

	u.promptUserSelection(execResp, selfRegUserTypes)
	return nil
}

// getUserTypeAndOU retrieves the user type by name and returns the schema and organization unit ID.
func (u *userTypeResolver) getUserTypeAndOU(userTypeName string) (*usertype.UserType, string, error) {
	logger := u.logger.With(log.String(userTypeKey, userTypeName))

	userType, svcErr := u.userTypeService.GetUserTypeByName(userTypeName)
	if svcErr != nil {
		logger.Error("Failed to resolve user type for user type",
			log.String(userTypeKey, userTypeName), log.String("error", svcErr.Error))
		return nil, "", fmt.Errorf("failed to resolve user type for user type: %s", userTypeName)
	}

	if userType.OrganizationUnitID == "" {
		logger.Error("No organization unit found for user type", log.String(userTypeKey, userTypeName))
		return nil, "", fmt.Errorf("no organization unit found for user type: %s", userTypeName)
	}

	logger.Debug("User type resolved for user type", log.String(userTypeKey, userTypeName),
		log.String(ouIDKey, userType.OrganizationUnitID))
	return userType, userType.OrganizationUnitID, nil
}

// resolveForUserOnboarding handles user type resolution for user onboarding flows.
func (u *userTypeResolver) resolveForUserOnboarding(ctx *core.NodeContext,
	execResp *common.ExecutorResponse) (*common.ExecutorResponse, error) {
	logger := u.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))

	// If userType already provided, validate and set runtime data
	if userTypeName, ok := ctx.UserInputs[userTypeKey]; ok && userTypeName != "" {
		userType, ouID, err := u.getUserTypeAndOU(userTypeName)
		if err != nil {
			execResp.Status = common.ExecFailure
			execResp.FailureReason = "Invalid user type"
			return execResp, nil
		}

		execResp.RuntimeData[userTypeKey] = userType.Name
		execResp.RuntimeData[defaultOUIDKey] = ouID
		logger.Debug("User type resolved for user onboarding", log.String(userTypeKey, userType.Name),
			log.String(ouIDKey, userType.OrganizationUnitID))
		execResp.Status = common.ExecComplete
		return execResp, nil
	}

	// List all available user types
	schemas, svcErr := u.userTypeService.GetUserTypeList(100, 0)
	if svcErr != nil {
		logger.Debug("Failed to list user types", log.String("error", svcErr.Error))
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "Failed to retrieve user types"
		return execResp, nil
	}

	if len(schemas.Types) == 0 {
		logger.Debug("No user types available")
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "No user types available"
		return execResp, nil
	}

	options := make([]string, 0, len(schemas.Types))
	for _, schema := range schemas.Types {
		options = append(options, schema.Name)
	}

	u.promptUserSelection(execResp, options)
	return execResp, nil
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
	}
}
