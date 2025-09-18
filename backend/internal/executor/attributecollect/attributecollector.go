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

// Package attributecollect provides the implementation for collecting user attributes and updating the user profile.
package attributecollect

import (
	"encoding/json"
	"errors"
	"fmt"

	flowconst "github.com/asgardeo/thunder/internal/flow/constants"
	flowmodel "github.com/asgardeo/thunder/internal/flow/model"
	"github.com/asgardeo/thunder/internal/system/log"
	usermodel "github.com/asgardeo/thunder/internal/user/model"
	"github.com/asgardeo/thunder/internal/user/service"
)

const (
	loggerComponentName   = "AttributeCollector"
	userAttributeUserID   = "userID"
	userAttributePassword = "password"
)

// TODO: Need to handle complex attributes and nested structures in the user profile.
//  Currently executor only takes string inputs.

// AttributeCollector is an executor that collects user attributes and updates the user profile.
type AttributeCollector struct {
	internal    flowmodel.Executor
	userService service.UserServiceInterface
}

var _ flowmodel.ExecutorInterface = (*AttributeCollector)(nil)

// NewAttributeCollector creates a new instance of AttributeCollector.
func NewAttributeCollector(id, name string, properties map[string]string) *AttributeCollector {
	prerequisites := []flowmodel.InputData{
		{
			Name:     "userID",
			Type:     "string",
			Required: true,
		},
	}

	return &AttributeCollector{
		internal:    *flowmodel.NewExecutor(id, name, []flowmodel.InputData{}, prerequisites, properties),
		userService: service.GetUserService(),
	}
}

// GetID returns the ID of the AttributeCollector.
func (a *AttributeCollector) GetID() string {
	return a.internal.GetID()
}

// GetName returns the name of the AttributeCollector.
func (a *AttributeCollector) GetName() string {
	return a.internal.GetName()
}

// GetProperties returns the properties of the AttributeCollector.
func (a *AttributeCollector) GetProperties() flowmodel.ExecutorProperties {
	return a.internal.Properties
}

// Execute executes the attribute collection logic.
func (a *AttributeCollector) Execute(ctx *flowmodel.NodeContext) (*flowmodel.ExecutorResponse, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorID, a.GetID()),
		log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Executing attribute collect executor")

	execResp := &flowmodel.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	if ctx.FlowType == flowconst.FlowTypeRegistration {
		logger.Debug("Flow type is registration, skipping attribute collection")
		execResp.Status = flowconst.ExecComplete
		return execResp, nil
	}

	if !ctx.AuthenticatedUser.IsAuthenticated {
		logger.Debug("User is not authenticated, cannot collect attributes")
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = "User is not authenticated"
		return execResp, nil
	}

	if !a.ValidatePrerequisites(ctx, execResp) {
		logger.Debug("Prerequisites validation failed for attribute collector")
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = "Prerequisites validation failed for attribute collector"
		return execResp, nil
	}

	if a.CheckInputData(ctx, execResp) {
		if execResp.Status == flowconst.ExecFailure {
			return execResp, nil
		}

		logger.Debug("Required input data for attribute collector is not provided")
		execResp.Status = flowconst.ExecUserInputRequired
		return execResp, nil
	}
	if execResp.Status == flowconst.ExecComplete {
		logger.Debug("Attribute collection is complete, no further action required")
		return execResp, nil
	}

	if err := a.updateUserInStore(ctx); err != nil {
		logger.Error("Failed to update user attributes", log.Error(err))
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = "Failed to update user attributes"
		return execResp, nil
	}

	logger.Debug("User attributes updated successfully")
	execResp.Status = flowconst.ExecComplete
	return execResp, nil
}

// GetDefaultExecutorInputs returns the default inputs for the AttributeCollector.
func (a *AttributeCollector) GetDefaultExecutorInputs() []flowmodel.InputData {
	return a.internal.GetDefaultExecutorInputs()
}

// GetPrerequisites returns the prerequisites for the AttributeCollector.
func (a *AttributeCollector) GetPrerequisites() []flowmodel.InputData {
	return a.internal.GetPrerequisites()
}

// CheckInputData checks if the required input data is provided in the context.
// If not present, it tries to retrieve user attributes from the user profile.
// If the attributes are not found, it adds the required data to the executor response.
func (a *AttributeCollector) CheckInputData(ctx *flowmodel.NodeContext, execResp *flowmodel.ExecutorResponse) bool {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorID, a.GetID()),
		log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Checking input data for the attribute collector")

	inputRequired := a.internal.CheckInputData(ctx, execResp)
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
				// If the attribute is a password, do not retrieve it from the profile.
				if inputData.Name == userAttributePassword {
					continue
				}

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

	// Update the executor response with the required data by checking the user profile.
	userAttributes, err := a.getUserAttributes(ctx)
	if err != nil {
		logger.Error("Failed to retrieve user attributes", log.Error(err))
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = "Failed to retrieve user attributes from user profile"
		return true
	}
	if userAttributes == nil {
		logger.Debug("No user attributes found in the user profile, proceeding with required data")
		return true
	}

	// Clear the required data in the executor response to avoid duplicates.
	missingAttributes := execResp.RequiredData
	execResp.RequiredData = make([]flowmodel.InputData, 0)
	if execResp.RuntimeData == nil {
		execResp.RuntimeData = make(map[string]string)
	}

	for _, inputData := range missingAttributes {
		attribute, exists := userAttributes[inputData.Name]
		if exists {
			// If the attribute is a password, do not retrieve it from the profile.
			if inputData.Name == userAttributePassword {
				continue
			}
			logger.Debug("Attribute exists in user profile, adding to runtime data",
				log.String("attributeName", inputData.Name))

			// TODO: This conversion should be modified according to the storage mechanism of the
			//  user store implementation.
			if strVal, ok := attribute.(string); ok {
				execResp.RuntimeData[inputData.Name] = strVal
			} else {
				execResp.RuntimeData[inputData.Name] = fmt.Sprintf("%v", attribute)
			}
		} else {
			logger.Debug("Attribute does not exist in user profile, adding to required data",
				log.String("attributeName", inputData.Name))
			execResp.RequiredData = append(execResp.RequiredData, inputData)
		}
	}

	if len(execResp.RequiredData) == 0 {
		logger.Debug("All required attributes are available in the user profile, no further action needed")
		return false
	}

	return true
}

// ValidatePrerequisites validates the prerequisites for the AttributeCollector.
func (a *AttributeCollector) ValidatePrerequisites(ctx *flowmodel.NodeContext,
	execResp *flowmodel.ExecutorResponse) bool {
	return a.internal.ValidatePrerequisites(ctx, execResp)
}

// GetUserIDFromContext retrieves the user ID from the context.
func (a *AttributeCollector) GetUserIDFromContext(ctx *flowmodel.NodeContext) (string, error) {
	return a.internal.GetUserIDFromContext(ctx)
}

// GetRequiredData returns the required input data for the AttributeCollector.
func (a *AttributeCollector) GetRequiredData(ctx *flowmodel.NodeContext) []flowmodel.InputData {
	return a.getRequiredData(ctx)
}

// getUserAttributes retrieves the user attributes from the user profile.
func (a *AttributeCollector) getUserAttributes(ctx *flowmodel.NodeContext) (map[string]interface{}, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorID, a.GetID()),
		log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Retrieving user attributes from the user profile")

	user, err := a.getUserFromStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user from store: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Unmarshal the user attributes if they exist
	var userAttributes map[string]interface{}
	if user.Attributes != nil {
		if err := json.Unmarshal(user.Attributes, &userAttributes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal user attributes: %w", err)
		}
	} else {
		userAttributes = make(map[string]interface{})
	}
	logger.Debug("User attributes retrieved successfully")

	return userAttributes, nil
}

// updateUserInStore updates the user profile with the collected attributes.
func (a *AttributeCollector) updateUserInStore(ctx *flowmodel.NodeContext) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorID, a.GetID()),
		log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Updating user attributes")

	user, err := a.getUserFromStore(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve user from store: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}
	userID := user.ID

	updateRequired, updatedUser, err := a.getUpdatedUserObject(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to get updated user object: %w", err)
	}
	if !updateRequired {
		logger.Debug("No updates required for user attributes, skipping update")
		return nil
	}
	if updatedUser == nil {
		return errors.New("failed to create updated user object")
	}

	if _, svcErr := a.userService.UpdateUser(userID, updatedUser); svcErr != nil {
		return fmt.Errorf("failed to update user attributes: %s", svcErr.Error)
	}
	logger.Debug("User attributes updated successfully", log.String("userID", userID))

	return nil
}

// getUserFromStore retrieves the user profile from the user store.
func (a *AttributeCollector) getUserFromStore(ctx *flowmodel.NodeContext) (*usermodel.User, error) {
	userID, err := a.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user ID: %w", err)
	}
	if userID == "" {
		return nil, errors.New("user ID is not available in the context")
	}

	user, svcErr := a.userService.GetUser(userID)
	if svcErr != nil {
		return nil, fmt.Errorf("failed to get user by ID: %s", svcErr.Error)
	}

	return user, nil
}

// getUpdatedUserObject creates a new user object with the updated attributes.
func (a *AttributeCollector) getUpdatedUserObject(ctx *flowmodel.NodeContext,
	user *usermodel.User) (bool, *usermodel.User, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorID, a.GetID()),
		log.String(log.LoggerKeyFlowID, ctx.FlowID))

	updatedUser := &usermodel.User{
		ID:               user.ID,
		OrganizationUnit: user.OrganizationUnit,
		Type:             user.Type,
	}

	// Get the existing attributes
	var existingAttrs map[string]interface{}
	if user.Attributes != nil {
		if err := json.Unmarshal(user.Attributes, &existingAttrs); err != nil {
			return false, nil, fmt.Errorf("failed to unmarshal existing user attributes: %w", err)
		}
	} else {
		existingAttrs = make(map[string]interface{})
	}

	// Get new attributes from input
	newAttrs := a.getInputAttributes(ctx)
	if len(newAttrs) == 0 {
		logger.Debug("No new attributes provided, returning existing user")
		return false, user, nil
	}

	// Merge attributes
	for k, v := range newAttrs {
		existingAttrs[k] = v
	}

	// Marshal the merged attributes back to JSON
	if len(existingAttrs) > 0 {
		mergedAttrs, err := json.Marshal(existingAttrs)
		if err != nil {
			return false, nil, fmt.Errorf("failed to marshal merged attributes: %w", err)
		} else {
			updatedUser.Attributes = mergedAttrs
		}
	}

	return true, updatedUser, nil
}

// getInputAttributes retrieves the input attributes from the context.
func (a *AttributeCollector) getInputAttributes(ctx *flowmodel.NodeContext) map[string]interface{} {
	attributesMap := make(map[string]interface{})
	requiredInputAttrs := a.getRequiredData(ctx)

	for _, inputAttr := range requiredInputAttrs {
		// Skip special attributes that shouldn't be stored/ updated in the user profile
		if inputAttr.Name == userAttributeUserID {
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

// getRequiredData returns the required input data for the AttributeCollector.
func (a *AttributeCollector) getRequiredData(ctx *flowmodel.NodeContext) []flowmodel.InputData {
	executorReqData := a.GetDefaultExecutorInputs()
	requiredData := ctx.NodeInputData

	if len(requiredData) == 0 {
		requiredData = executorReqData
	} else {
		// Append the default required data if not already present.
		for _, inputData := range executorReqData {
			exists := false
			for _, existingInputData := range requiredData {
				if existingInputData.Name == inputData.Name {
					exists = true
					break
				}
			}
			// If the input data already exists, skip adding it again.
			if !exists {
				requiredData = append(requiredData, inputData)
			}
		}
	}

	return requiredData
}
