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

// Package provision provides the implementation for user provisioning in a flow.
package provision

import (
	"encoding/json"
	"fmt"
	"slices"

	authncm "github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/executor/identify"
	flowconst "github.com/asgardeo/thunder/internal/flow/common/constants"
	flowmodel "github.com/asgardeo/thunder/internal/flow/common/model"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/user"
)

const (
	loggerComponentName   = "ProvisioningExecutor"
	passwordAttributeName = "password"
)

var nonUserAttributes = []string{"userID", "code", "nonce", "state", "flowID",
	"otp", "attemptCount", "expiryTimeInMillis", "value"}

// ProvisioningExecutor implements the ExecutorInterface for user provisioning in a flow.
type ProvisioningExecutor struct {
	*identify.IdentifyingExecutor
	internal    flowmodel.Executor
	userService user.UserServiceInterface
}

var _ flowmodel.ExecutorInterface = (*ProvisioningExecutor)(nil)

// NewProvisioningExecutor creates a new instance of ProvisioningExecutor.
func NewProvisioningExecutor(id, name string, properties map[string]string) *ProvisioningExecutor {
	return &ProvisioningExecutor{
		IdentifyingExecutor: identify.NewIdentifyingExecutor(id, name, properties),
		internal: *flowmodel.NewExecutor(id, name, flowconst.ExecutorTypeRegistration,
			[]flowmodel.InputData{}, []flowmodel.InputData{},
			properties),
		userService: user.GetUserService(),
	}
}

// GetID returns the ID of the ProvisioningExecutor.
func (p *ProvisioningExecutor) GetID() string {
	return p.internal.GetID()
}

// GetName returns the name of the ProvisioningExecutor.
func (p *ProvisioningExecutor) GetName() string {
	return p.internal.GetName()
}

// GetProperties returns the properties of the ProvisioningExecutor.
func (p *ProvisioningExecutor) GetProperties() flowmodel.ExecutorProperties {
	return p.internal.GetProperties()
}

// Execute executes the user provisioning logic based on the inputs provided.
func (p *ProvisioningExecutor) Execute(ctx *flowmodel.NodeContext) (*flowmodel.ExecutorResponse, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorID, p.GetID()),
		log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Executing user provisioning executor")

	execResp := &flowmodel.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	if ctx.FlowType != flowconst.FlowTypeRegistration {
		logger.Warn("ProvisioningExecutor is only applicable for registration flows, skipping execution")
		execResp.Status = flowconst.ExecComplete
		return execResp, nil
	}

	if p.CheckInputData(ctx, execResp) {
		if execResp.Status == flowconst.ExecFailure {
			return execResp, nil
		}

		logger.Debug("Required input data for provisioning executor is not provided")
		execResp.Status = flowconst.ExecUserInputRequired
		return execResp, nil
	}

	userAttributes := p.getInputAttributes(ctx)
	if len(userAttributes) == 0 {
		logger.Debug("No user attributes provided for provisioning")
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = "No user attributes provided for provisioning"
		return execResp, nil
	}

	userID, err := p.IdentifyUser(userAttributes, execResp)
	if err != nil {
		logger.Error("Failed to identify user", log.Error(err))
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = "Failed to identify user"
		return execResp, nil
	}
	if execResp.Status == flowconst.ExecFailure && execResp.FailureReason != "User not found" {
		return execResp, nil
	}
	if userID != nil && *userID != "" {
		logger.Debug("User already exists", log.String("userID", *userID))
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = "User already exists"
		return execResp, nil
	}

	// Create the user in the store.
	p.appendNonIdentifyingAttributes(ctx, &userAttributes)
	createdUser, err := p.createUserInStore(ctx.FlowID, userAttributes)
	if err != nil {
		logger.Error("Failed to create user in the store", log.Error(err))
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = "Failed to create user"
		return execResp, nil
	}
	if createdUser == nil || createdUser.ID == "" {
		logger.Error("Created user is nil or has no ID")
		execResp.Status = flowconst.ExecFailure
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
	execResp.Status = flowconst.ExecComplete

	return execResp, nil
}

// GetDefaultExecutorInputs returns the default inputs for the ProvisioningExecutor.
func (p *ProvisioningExecutor) GetDefaultExecutorInputs() []flowmodel.InputData {
	return p.internal.GetDefaultExecutorInputs()
}

// GetPrerequisites returns the prerequisites for the ProvisioningExecutor.
func (p *ProvisioningExecutor) GetPrerequisites() []flowmodel.InputData {
	return p.internal.GetPrerequisites()
}

// CheckInputData checks if the required input data is provided in the context.
// If the attributes are not found, it adds the required data to the executor response.
func (p *ProvisioningExecutor) CheckInputData(ctx *flowmodel.NodeContext, execResp *flowmodel.ExecutorResponse) bool {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorID, p.GetID()),
		log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Checking input data for the provisioning executor")

	inputRequired := p.internal.CheckInputData(ctx, execResp)
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
		execResp.RequiredData = make([]flowmodel.InputData, 0)
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

// ValidatePrerequisites validates the prerequisites for the ProvisioningExecutor.
func (p *ProvisioningExecutor) ValidatePrerequisites(ctx *flowmodel.NodeContext,
	execResp *flowmodel.ExecutorResponse) bool {
	return p.internal.ValidatePrerequisites(ctx, execResp)
}

// GetUserIDFromContext retrieves the user ID from the context.
func (p *ProvisioningExecutor) GetUserIDFromContext(ctx *flowmodel.NodeContext) (string, error) {
	return p.internal.GetUserIDFromContext(ctx)
}

// GetRequiredData returns the required input data for the AttributeCollector.
func (p *ProvisioningExecutor) GetRequiredData(ctx *flowmodel.NodeContext) []flowmodel.InputData {
	return p.internal.GetRequiredData(ctx)
}

// getInputAttributes retrieves the input attributes from the context to be stored in user profile.
func (p *ProvisioningExecutor) getInputAttributes(ctx *flowmodel.NodeContext) map[string]interface{} {
	attributesMap := make(map[string]interface{})
	requiredInputAttrs := p.GetRequiredData(ctx)

	// If no input attributes are defined, get all user attributes from the context.
	if len(requiredInputAttrs) == 0 {
		for key, value := range ctx.UserInputData {
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
		}
	}

	return attributesMap
}

// appendNonIdentifyingAttributes appends non-identifying attributes to the provided attributes map.
func (p *ProvisioningExecutor) appendNonIdentifyingAttributes(ctx *flowmodel.NodeContext,
	attributes *map[string]interface{}) {
	if value, exists := ctx.UserInputData[passwordAttributeName]; exists {
		(*attributes)[passwordAttributeName] = value
	} else if runtimeValue, exists := ctx.RuntimeData[passwordAttributeName]; exists {
		(*attributes)[passwordAttributeName] = runtimeValue
	}
}

// createUserInStore creates a new user in the user store with the provided attributes.
func (p *ProvisioningExecutor) createUserInStore(flowID string,
	userAttributes map[string]interface{}) (*user.User, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorID, p.GetID()),
		log.String(log.LoggerKeyFlowID, flowID))
	logger.Debug("Creating the user account")

	// TODO: Use a hard coded ou for the moment. This needs to be resolved properly
	//  when the support is implemented.
	newUser := user.User{
		OrganizationUnit: "00000000-0000-0000-0000-000000000000", // This should be configurable
	}

	// Takes the user type from the context, if available.
	if userType, exists := userAttributes["type"]; exists {
		newUser.Type = userType.(string)
	} else {
		// TODO: Use a hard coded type for the moment. This needs to be resolved accordingly.
		newUser.Type = "person"
	}

	// Convert the user attributes to JSON.
	attributesJSON, err := json.Marshal(userAttributes)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user attributes: %w", err)
	}
	newUser.Attributes = attributesJSON

	retUser, svcErr := p.userService.CreateUser(&newUser)
	if svcErr != nil {
		return nil, fmt.Errorf("failed to create user in the store: %s", svcErr.Error)
	}
	logger.Debug("User account created successfully", log.String("userID", retUser.ID))

	return retUser, nil
}
