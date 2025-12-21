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

	authncm "github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/user"
)

// ExecutorNameInvitationRedemption is the name of the invitation redemption executor.
const ExecutorNameInvitationRedemption = "InvitationRedemptionExecutor"

// Runtime data keys for invitation flow
const (
	RuntimeKeyInvitedUserID    = "invitedUserID"
	RuntimeKeyIsInvitationFlow = "isInvitationFlow"
)

// invitationRedemptionExecutor implements the ExecutorInterface for handling invited user credential setup.
// This executor is used in the registration flow to set credentials for pre-created invited users.
type invitationRedemptionExecutor struct {
	core.ExecutorInterface
	userService user.UserServiceInterface
	logger      *log.Logger
}

var _ core.ExecutorInterface = (*invitationRedemptionExecutor)(nil)

// newInvitationRedemptionExecutor creates a new instance of invitationRedemptionExecutor.
func newInvitationRedemptionExecutor(
	flowFactory core.FlowFactoryInterface,
	userService user.UserServiceInterface,
) *invitationRedemptionExecutor {
	logger := log.GetLogger().With(
		log.String(log.LoggerKeyComponentName, ExecutorNameInvitationRedemption),
		log.String(log.LoggerKeyExecutorName, ExecutorNameInvitationRedemption))

	// Required input is password for setting credentials on the invited user
	requiredInputs := []common.Input{
		{Identifier: userAttributePassword, Type: "password", Required: true},
	}

	base := flowFactory.CreateExecutor(ExecutorNameInvitationRedemption, common.ExecutorTypeRegistration,
		requiredInputs, []common.Input{})

	return &invitationRedemptionExecutor{
		ExecutorInterface: base,
		userService:       userService,
		logger:            logger,
	}
}

// Execute handles the invitation redemption process:
// 1. Checks if this is an invitation flow (via runtime data)
// 2. Gets the invited user ID from runtime data
// 3. Collects credentials from the user
// 4. Sets credentials on the invited user and activates them
func (e *invitationRedemptionExecutor) Execute(ctx *core.NodeContext) (*common.ExecutorResponse, error) {
	logger := e.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Executing invitation redemption executor")

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	// Check if this is an invitation flow
	isInvitationFlow, ok := ctx.RuntimeData[RuntimeKeyIsInvitationFlow]
	if !ok || isInvitationFlow != dataValueTrue {
		logger.Debug("Not an invitation flow, skipping execution")
		execResp.Status = common.ExecComplete
		return execResp, nil
	}

	// Get the invited user ID from runtime data
	userID := ctx.RuntimeData[RuntimeKeyInvitedUserID]
	if userID == "" {
		logger.Error("Invited user ID not found in runtime data")
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "Invalid invitation context"
		return execResp, nil
	}

	// Set the user ID in runtime data for downstream executors
	execResp.RuntimeData[userAttributeUserID] = userID

	// Check if password is provided
	if !e.HasRequiredInputs(ctx, execResp) {
		logger.Debug("Password not provided, requesting user input")
		execResp.Status = common.ExecUserInputRequired
		return execResp, nil
	}

	// Get the password from user inputs
	password, exists := ctx.UserInputs[userAttributePassword]
	if !exists || password == "" {
		logger.Debug("Password not provided in user inputs")
		execResp.Status = common.ExecUserInputRequired
		execResp.Inputs = []common.Input{
			{Identifier: userAttributePassword, Type: "password", Required: true},
		}
		return execResp, nil
	}

	// Set credentials on the user and activate them
	credentials := map[string]string{userAttributePassword: password}
	credentialsJSON, err := json.Marshal(credentials)
	if err != nil {
		logger.Error("Failed to marshal credentials", log.Error(err))
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "Failed to process credentials"
		return execResp, nil
	}

	// Set credentials and activate the user
	if svcErr := e.userService.SetUserCredentialsAndActivate(userID, credentialsJSON); svcErr != nil {
		logger.Error("Failed to set credentials and activate user",
			log.String("userID", userID), log.String("error", svcErr.Error))
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "Failed to set credentials"
		return execResp, nil
	}

	// Get the user details for the authenticated user response
	userDetails, svcErr := e.userService.GetUser(userID)
	if svcErr != nil {
		logger.Error("Failed to get user details", log.String("userID", userID), log.String("error", svcErr.Error))
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "Failed to retrieve user"
		return execResp, nil
	}

	var userAttributes map[string]interface{}
	if err := json.Unmarshal(userDetails.Attributes, &userAttributes); err != nil {
		logger.Error("Failed to unmarshal user attributes", log.Error(err))
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "Failed to process user data"
		return execResp, nil
	}

	// Set the authenticated user
	execResp.AuthenticatedUser = authncm.AuthenticatedUser{
		IsAuthenticated:    true,
		UserID:             userDetails.ID,
		OrganizationUnitID: userDetails.OrganizationUnit,
		UserType:           userDetails.Type,
		Attributes:         userAttributes,
	}

	execResp.Status = common.ExecComplete
	logger.Debug("Invitation redemption completed successfully", log.String("userID", userID))

	return execResp, nil
}
