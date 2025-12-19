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
	"encoding/json"
	"fmt"
	"slices"

	authncm "github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/flow/common"
	flowcm "github.com/asgardeo/thunder/internal/flow/common"
	flowcore "github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/user"
)

const (
	provisioningLoggerComponentName = "ProvisioningExecutor"
)

// provisioningExecutor implements the ExecutorInterface for user provisioning in a flow.
type provisioningExecutor struct {
	flowcore.ExecutorInterface
	identifyingExecutorInterface
	userService user.UserServiceInterface
	logger      *log.Logger
}

var _ flowcore.ExecutorInterface = (*provisioningExecutor)(nil)
var _ identifyingExecutorInterface = (*provisioningExecutor)(nil)

// newProvisioningExecutor creates a new instance of ProvisioningExecutor.
func newProvisioningExecutor(
	flowFactory flowcore.FlowFactoryInterface,
	userService user.UserServiceInterface,
) *provisioningExecutor {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, provisioningLoggerComponentName),
		log.String(log.LoggerKeyExecutorName, ExecutorNameProvisioning))

	base := flowFactory.CreateExecutor(ExecutorNameProvisioning, flowcm.ExecutorTypeRegistration,
		[]flowcm.InputData{}, []flowcm.InputData{})

	identifyingExec := newIdentifyingExecutor(ExecutorNameProvisioning,
		[]flowcm.InputData{}, []flowcm.InputData{}, flowFactory, userService)

	return &provisioningExecutor{
		ExecutorInterface:            base,
		identifyingExecutorInterface: identifyingExec,
		userService:                  userService,
		logger:                       logger,
	}
}

// Execute executes the user provisioning logic based on the inputs provided.
func (p *provisioningExecutor) Execute(ctx *flowcore.NodeContext) (*flowcm.ExecutorResponse, error) {
	logger := p.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Executing user provisioning executor")

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	// If it's an authentication flow, skip execution if the user is not eligible for provisioning
	if ctx.FlowType == flowcm.FlowTypeAuthentication {
		eligible, ok := ctx.RuntimeData[common.RuntimeKeyUserEligibleForProvisioning]
		if !ok || eligible != dataValueTrue {
			logger.Debug("User is not eligible for provisioning, skipping execution")
			execResp.Status = flowcm.ExecComplete
			return execResp, nil
		}
	}

	if p.CheckInputData(ctx, execResp) {
		if execResp.Status == flowcm.ExecFailure {
			return execResp, nil
		}

		logger.Debug("Required input data for provisioning executor is not provided")
		execResp.Status = flowcm.ExecUserInputRequired
		return execResp, nil
	}

	userAttributes := p.getAttributesForProvisioning(ctx)
	if len(userAttributes) == 0 {
		logger.Debug("No user attributes provided for provisioning")
		execResp.Status = flowcm.ExecFailure
		execResp.FailureReason = "No user attributes provided for provisioning"
		return execResp, nil
	}

	userID, err := p.IdentifyUser(ctx, userAttributes, execResp)
	if err != nil {
		logger.Error("Failed to identify user", log.Error(err))
		execResp.Status = flowcm.ExecFailure
		execResp.FailureReason = "Failed to identify user"
		return execResp, nil
	}
	if execResp.Status == flowcm.ExecFailure && execResp.FailureReason != failureReasonUserNotFound {
		return execResp, nil
	}
	if userID != nil && *userID != "" {
		logger.Debug("User already exists", log.String("userID", *userID))

		// If it's a registration flow, check if proceeding with an existing user
		if ctx.FlowType == flowcm.FlowTypeRegistration {
			existing, ok := ctx.RuntimeData[common.RuntimeKeySkipProvisioning]
			if ok && existing == dataValueTrue {
				logger.Debug("Proceeding with an existing user in registration flow, skipping execution")
				execResp.RuntimeData[userAttributeUserID] = *userID
				execResp.Status = flowcm.ExecComplete
				return execResp, nil
			}
		}

		execResp.Status = flowcm.ExecFailure
		execResp.FailureReason = "User already exists"
		return execResp, nil
	}

	// Create the user in the store.
	p.appendNonIdentifyingAttributes(ctx, &userAttributes)
	createdUser, err := p.createUserInStore(ctx, userAttributes)
	if err != nil {
		logger.Error("Failed to create user in the store", log.Error(err))
		execResp.Status = flowcm.ExecFailure
		execResp.FailureReason = "Failed to create user"
		return execResp, nil
	}
	if createdUser == nil || createdUser.ID == "" {
		logger.Error("Created user is nil or has no ID")
		execResp.Status = flowcm.ExecFailure
		execResp.FailureReason = "Something went wrong while creating the user"
		return execResp, nil
	}

	logger.Debug("User created successfully", log.String("userID", createdUser.ID))

	var retAttributes map[string]interface{}
	if err := json.Unmarshal(createdUser.Attributes, &retAttributes); err != nil {
		logger.Error("Failed to unmarshal user attributes", log.Error(err))
		return nil, err
	}

	authenticatedUser := authncm.AuthenticatedUser{
		IsAuthenticated:    true,
		UserID:             createdUser.ID,
		OrganizationUnitID: createdUser.OrganizationUnit,
		UserType:           createdUser.Type,
		Attributes:         retAttributes,
	}
	execResp.AuthenticatedUser = authenticatedUser
	execResp.Status = flowcm.ExecComplete

	// Set user id in runtime data
	execResp.RuntimeData[userAttributeUserID] = createdUser.ID

	// Set the auto-provisioned flag if it's a user auto provisioning scenario
	if ctx.FlowType == flowcm.FlowTypeAuthentication {
		execResp.RuntimeData[common.RuntimeKeyUserAutoProvisioned] = dataValueTrue
	}

	return execResp, nil
}

// CheckInputData checks if the required input data is provided in the context.
// If the attributes are not found, it adds the required data to the executor response.
func (p *provisioningExecutor) CheckInputData(ctx *flowcore.NodeContext, execResp *flowcm.ExecutorResponse) bool {
	logger := p.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Checking input data for the provisioning executor")

	inputRequired := p.ExecutorInterface.CheckInputData(ctx, execResp)
	if !inputRequired {
		return false
	}
	if len(execResp.RequiredData) == 0 {
		return false
	}

	// Update the executor response with the required data retrieved from authenticated user attributes.
	authnUserAttrs := ctx.AuthenticatedUser.Attributes
	if len(authnUserAttrs) > 0 {
		logger.Debug("Authenticated user attributes found, updating executor response required data")

		// Clear the required data in the executor response to avoid duplicates.
		missingAttributes := execResp.RequiredData
		execResp.RequiredData = make([]flowcm.InputData, 0)
		if execResp.RuntimeData == nil {
			execResp.RuntimeData = make(map[string]string)
		}

		for _, inputData := range missingAttributes {
			attribute, exists := authnUserAttrs[inputData.Name]
			if exists {
				attributeStr, ok := attribute.(string)
				if ok {
					logger.Debug("Attribute exists in authenticated user attributes, adding to runtime data",
						log.String("attributeName", inputData.Name))
					execResp.RuntimeData[inputData.Name] = attributeStr
				}
			} else {
				logger.Debug("Attribute does not exist in authenticated user attributes, adding to required data",
					log.String("attributeName", inputData.Name))
				execResp.RequiredData = append(execResp.RequiredData, inputData)
			}
		}

		if len(execResp.RequiredData) == 0 {
			logger.Debug("All required attributes are available in authenticated user attributes, " +
				"no further action needed")
			return false
		}
	}

	return true
}

// getAttributesForProvisioning retrieves the input attributes from the context to be stored in user profile.
func (p *provisioningExecutor) getAttributesForProvisioning(ctx *flowcore.NodeContext) map[string]interface{} {
	attributesMap := make(map[string]interface{})
	requiredInputAttrs := p.GetRequiredData(ctx)

	// If no input attributes are defined, get all user attributes from the context.
	if len(requiredInputAttrs) == 0 {
		for key, value := range ctx.UserInputData {
			if !slices.Contains(nonUserAttributes, key) {
				attributesMap[key] = value
			}
		}
		for key, value := range ctx.AuthenticatedUser.Attributes {
			if !slices.Contains(nonUserAttributes, key) {
				attributesMap[key] = value
			}
		}
		for key, value := range ctx.RuntimeData {
			if !slices.Contains(nonUserAttributes, key) {
				attributesMap[key] = value
			}
		}
		return attributesMap
	}

	// Otherwise, filter the required input attributes and get their values from the context.
	for _, inputAttr := range requiredInputAttrs {
		if slices.Contains(nonUserAttributes, inputAttr.Name) {
			continue
		}

		value, exists := ctx.UserInputData[inputAttr.Name]
		if exists {
			attributesMap[inputAttr.Name] = value
		} else if runtimeValue, exists := ctx.RuntimeData[inputAttr.Name]; exists {
			attributesMap[inputAttr.Name] = runtimeValue
		} else if authnValue, exists := ctx.AuthenticatedUser.Attributes[inputAttr.Name]; exists {
			attributesMap[inputAttr.Name] = authnValue
		}
	}

	return attributesMap
}

// appendNonIdentifyingAttributes appends non-identifying attributes to the provided attributes map.
func (p *provisioningExecutor) appendNonIdentifyingAttributes(ctx *flowcore.NodeContext,
	attributes *map[string]interface{}) {
	if value, exists := ctx.UserInputData[userAttributePassword]; exists {
		(*attributes)[userAttributePassword] = value
	} else if runtimeValue, exists := ctx.RuntimeData[userAttributePassword]; exists {
		(*attributes)[userAttributePassword] = runtimeValue
	}
}

// createUserInStore creates a new user in the user store with the provided attributes.
func (p *provisioningExecutor) createUserInStore(ctx *flowcore.NodeContext,
	userAttributes map[string]interface{}) (*user.User, error) {
	logger := p.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Creating the user account")

	ouID := p.getOuID(ctx)
	if ouID == "" {
		return nil, fmt.Errorf("organization unit ID not found")
	}
	userType := p.getUserType(ctx)
	if userType == "" {
		return nil, fmt.Errorf("user type not found")
	}

	newUser := user.User{
		OrganizationUnit: ouID,
		Type:             userType,
	}

	// Convert the user attributes to JSON.
	attributesJSON, err := json.Marshal(userAttributes)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user attributes: %w", err)
	}
	newUser.Attributes = attributesJSON

	retUser, svcErr := p.userService.CreateUser(ctx.RequestContext, &newUser)
	if svcErr != nil {
		return nil, fmt.Errorf("failed to create user in the store: %s", svcErr.Error)
	}
	if retUser != nil && retUser.ID != "" {
		logger.Debug("User account created successfully", log.String("userID", retUser.ID))
	}

	return retUser, nil
}

// getOuID retrieves the organization unit ID from runtime data.
func (p *provisioningExecutor) getOuID(ctx *flowcore.NodeContext) string {
	ouID := ""
	// Check for ouId in runtime data
	if val, ok := ctx.RuntimeData[ouIDKey]; ok && val != "" {
		ouID = val
	}
	// If not found, check for defaultOUID in runtime data
	if ouID == "" {
		if val, ok := ctx.RuntimeData[defaultOUIDKey]; ok && val != "" {
			ouID = val
		}
	}

	return ouID
}

// getUserType retrieves the user type from runtime data.
func (p *provisioningExecutor) getUserType(ctx *flowcore.NodeContext) string {
	userType := ""
	if val, ok := ctx.RuntimeData[userTypeKey]; ok && val != "" {
		userType = val
	}

	return userType
}
