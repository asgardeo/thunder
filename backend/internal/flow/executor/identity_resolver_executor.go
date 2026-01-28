/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/user"
)

// identityResolverExecutor implements the ExecutorInterface for resolving user identity from username.
// This is a standalone executor that can be used in flows to resolve username → userID.
type identityResolverExecutor struct {
	core.ExecutorInterface
	identifyingExecutorInterface
	userService user.UserServiceInterface
	logger      *log.Logger
}

var _ core.ExecutorInterface = (*identityResolverExecutor)(nil)

// newIdentityResolverExecutor creates a new instance of IdentityResolverExecutor.
func newIdentityResolverExecutor(
	flowFactory core.FlowFactoryInterface,
	userService user.UserServiceInterface,
) *identityResolverExecutor {
	defaultInputs := []common.Input{
		{
			Identifier: userAttributeUsername,
			Type:       "string",
			Required:   true,
		},
	}

	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "IdentityResolverExecutor"),
		log.String(log.LoggerKeyExecutorName, ExecutorNameIdentityResolver))

	identifyExec := newIdentifyingExecutor(ExecutorNameIdentityResolver, defaultInputs, []common.Input{},
		flowFactory, userService)
	base := flowFactory.CreateExecutor(ExecutorNameIdentityResolver, common.ExecutorTypeUtility,
		defaultInputs, []common.Input{})

	return &identityResolverExecutor{
		ExecutorInterface:            base,
		identifyingExecutorInterface: identifyExec,
		userService:                  userService,
		logger:                       logger,
	}
}

// Execute executes the identity resolution logic.
// It takes a username input and resolves it to a userID, storing it in RuntimeData.
func (i *identityResolverExecutor) Execute(ctx *core.NodeContext) (*common.ExecutorResponse, error) {
	logger := i.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Executing identity resolver executor")

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	// Check if required inputs are provided
	if !i.HasRequiredInputs(ctx, execResp) {
		logger.Debug("Required inputs for identity resolver is not provided")
		execResp.Status = common.ExecUserInputRequired
		return execResp, nil
	}

	// Get username from inputs
	username := ctx.UserInputs[userAttributeUsername]
	if username == "" {
		username = ctx.RuntimeData[userAttributeUsername]
	}
	if username == "" {
		logger.Debug("Username not provided")
		execResp.Status = common.ExecUserInputRequired
		return execResp, nil
	}

	// Try to identify the user
	// TODO: Remove context.TODO() when context is available
	filters := map[string]interface{}{userAttributeUsername: username}
	userID, err := i.IdentifyUser(context.TODO(), filters, execResp)
	if err != nil {
		logger.Error("Failed to identify user", log.Error(err))
		execResp.Status = common.ExecFailure
		execResp.FailureReason = failureReasonFailedToIdentifyUser
		return execResp, nil
	}

	if userID == nil || *userID == "" {
		logger.Debug("User not found for the provided username")
		execResp.Status = common.ExecFailure
		execResp.FailureReason = failureReasonUserNotFound
		return execResp, nil
	}

	// Store the resolved userID in RuntimeData for subsequent executors
	execResp.RuntimeData[userAttributeUserID] = *userID
	execResp.Status = common.ExecComplete

	logger.Debug("Identity resolver executor completed successfully",
		log.String("username", username),
		log.String("userID", *userID))

	return execResp, nil
}
