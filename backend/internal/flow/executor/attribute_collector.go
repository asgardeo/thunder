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
	"errors"
	"fmt"

	flowcm "github.com/asgardeo/thunder/internal/flow/common"
	flowcore "github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/user"
)

const (
	attrCollectLoggerComponentName = "AttributeCollector"
)

// TODO: Need to handle complex attributes and nested structures in the user profile.
//  Currently executor only takes string inputs.

// attributeCollector is an executor that collects user attributes and updates the user profile.
type attributeCollector struct {
	flowcore.ExecutorInterface
	userService user.UserServiceInterface
	logger      *log.Logger
}

var _ flowcore.ExecutorInterface = (*attributeCollector)(nil)

// newAttributeCollector creates a new instance of AttributeCollector.
func newAttributeCollector(
	flowFactory flowcore.FlowFactoryInterface,
	userService user.UserServiceInterface,
) *attributeCollector {
	prerequisites := []flowcm.InputData{
		{
			Name:     "userID",
			Type:     "string",
			Required: true,
		},
	}
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, attrCollectLoggerComponentName),
		log.String(log.LoggerKeyExecutorName, ExecutorNameAttributeCollect))

	base := flowFactory.CreateExecutor(ExecutorNameAttributeCollect, flowcm.ExecutorTypeUtility,
		[]flowcm.InputData{}, prerequisites)

	return &attributeCollector{
		ExecutorInterface: base,
		userService:       userService,
		logger:            logger,
	}
}

// Execute executes the attribute collection logic.
func (a *attributeCollector) Execute(ctx *flowcore.NodeContext) (*flowcm.ExecutorResponse, error) {
	logger := a.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Executing attribute collect executor")

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	if ctx.FlowType == flowcm.FlowTypeRegistration {
		logger.Debug("Flow type is registration, skipping attribute collection")
		execResp.Status = flowcm.ExecComplete
		return execResp, nil
	}

	if !ctx.AuthenticatedUser.IsAuthenticated {
		logger.Debug("User is not authenticated, cannot collect attributes")
		execResp.Status = flowcm.ExecFailure
		execResp.FailureReason = failureReasonUserNotAuthenticated
		return execResp, nil
	}

	if !a.ValidatePrerequisites(ctx, execResp) {
		logger.Debug("Prerequisites validation failed for attribute collector")
		execResp.Status = flowcm.ExecFailure
		execResp.FailureReason = "Prerequisites validation failed for attribute collector"
		return execResp, nil
	}

	if a.CheckInputData(ctx, execResp) {
		if execResp.Status == flowcm.ExecFailure {
			return execResp, nil
		}

		logger.Debug("Required input data for attribute collector is not provided")
		execResp.Status = flowcm.ExecUserInputRequired
		return execResp, nil
	}
	if execResp.Status == flowcm.ExecComplete {
		logger.Debug("Attribute collection is complete, no further action required")
		return execResp, nil
	}

	if err := a.updateUserInStore(ctx); err != nil {
		logger.Error("Failed to update user attributes", log.Error(err))
		execResp.Status = flowcm.ExecFailure
		execResp.FailureReason = "Failed to update user attributes"
		return execResp, nil
	}

	logger.Debug("User attributes updated successfully")
	execResp.Status = flowcm.ExecComplete
	return execResp, nil
}

// CheckInputData checks if the required input data is provided in the context.
// If not present, it tries to retrieve user attributes from the user profile.
// If the attributes are not found, it adds the required data to the executor response.
func (a *attributeCollector) CheckInputData(ctx *flowcore.NodeContext, execResp *flowcm.ExecutorResponse) bool {
	logger := a.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Checking input data for the attribute collector")

	inputRequired := a.ExecutorInterface.CheckInputData(ctx, execResp)
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
		execResp.Status = flowcm.ExecFailure
		execResp.FailureReason = "Failed to retrieve user attributes from user profile"
		return true
	}
	if userAttributes == nil {
		logger.Debug("No user attributes found in the user profile, proceeding with required data")
		return true
	}

	// Clear the required data in the executor response to avoid duplicates.
	missingAttributes := execResp.RequiredData
	execResp.RequiredData = make([]flowcm.InputData, 0)
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

// getUserAttributes retrieves the user attributes from the user profile.
func (a *attributeCollector) getUserAttributes(ctx *flowcore.NodeContext) (map[string]interface{}, error) {
	logger := a.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
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
func (a *attributeCollector) updateUserInStore(ctx *flowcore.NodeContext) error {
	logger := a.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
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

	if _, svcErr := a.userService.UpdateUser(ctx.RequestContext, userID, updatedUser); svcErr != nil {
		return fmt.Errorf("failed to update user attributes: %s", svcErr.Error)
	}
	logger.Debug("User attributes updated successfully", log.String("userID", userID))

	return nil
}

// getUserFromStore retrieves the user profile from the user store.
func (a *attributeCollector) getUserFromStore(ctx *flowcore.NodeContext) (*user.User, error) {
	userID := a.GetUserIDFromContext(ctx)
	if userID == "" {
		return nil, errors.New("user ID is not available in the context")
	}

	user, svcErr := a.userService.GetUser(ctx.RequestContext, userID)
	if svcErr != nil {
		return nil, fmt.Errorf("failed to get user by ID: %s", svcErr.Error)
	}

	return user, nil
}

// getUpdatedUserObject creates a new user object with the updated attributes.
func (a *attributeCollector) getUpdatedUserObject(ctx *flowcore.NodeContext,
	userData *user.User) (bool, *user.User, error) {
	logger := a.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))

	updatedUser := &user.User{
		ID:               userData.ID,
		OrganizationUnit: userData.OrganizationUnit,
		Type:             userData.Type,
	}

	// Get the existing attributes
	var existingAttrs map[string]interface{}
	if userData.Attributes != nil {
		if err := json.Unmarshal(userData.Attributes, &existingAttrs); err != nil {
			return false, nil, fmt.Errorf("failed to unmarshal existing user attributes: %w", err)
		}
	} else {
		existingAttrs = make(map[string]interface{})
	}

	// Get new attributes from input
	newAttrs := a.getInputAttributes(ctx)
	if len(newAttrs) == 0 {
		logger.Debug("No new attributes provided, returning existing user")
		return false, userData, nil
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
func (a *attributeCollector) getInputAttributes(ctx *flowcore.NodeContext) map[string]interface{} {
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
func (a *attributeCollector) getRequiredData(ctx *flowcore.NodeContext) []flowcm.InputData {
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
