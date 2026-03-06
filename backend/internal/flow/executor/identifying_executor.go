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
	"slices"

	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/userprovider"
)

const (
	idfExecLoggerComponentName = "IdentifyingExecutor"
)

// identifyingExecutorInterface defines the interface for identifying executors.
type identifyingExecutorInterface interface {
	IdentifyUser(filters map[string]interface{},
		execResp *common.ExecutorResponse) (*string, error)
}

// identifyingExecutor implements the ExecutorInterface for identifying users based on provided attributes.
type identifyingExecutor struct {
	core.ExecutorInterface
	userProvider userprovider.UserProviderInterface
	logger       *log.Logger
}

var _ core.ExecutorInterface = (*identifyingExecutor)(nil)
var _ identifyingExecutorInterface = (*identifyingExecutor)(nil)

// newIdentifyingExecutor creates a new instance of IdentifyingExecutor.
func newIdentifyingExecutor(
	name string,
	defaultInputs, prerequisites []common.Input,
	flowFactory core.FlowFactoryInterface,
	userProvider userprovider.UserProviderInterface,
) *identifyingExecutor {
	if name == "" {
		name = ExecutorNameIdentifying
	}
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, idfExecLoggerComponentName),
		log.String(log.LoggerKeyExecutorName, name))

	base := flowFactory.CreateExecutor(ExecutorNameIdentifying, common.ExecutorTypeUtility,
		defaultInputs, prerequisites)
	return &identifyingExecutor{
		ExecutorInterface: base,
		userProvider:      userProvider,
		logger:            logger,
	}
}

// IdentifyUser identifies a user based on the provided attributes.
func (i *identifyingExecutor) IdentifyUser(filters map[string]interface{},
	execResp *common.ExecutorResponse) (*string, error) {
	logger := i.logger
	logger.Debug("Identifying user with filters")

	// filter out non-searchable attributes
	var searchableFilter = make(map[string]interface{})
	for key, value := range filters {
		if !slices.Contains(nonSearchableInputs, key) {
			searchableFilter[key] = value
		}
	}

	userID, err := i.userProvider.IdentifyUser(searchableFilter)
	if err != nil {
		if err.Code == userprovider.ErrorCodeUserNotFound {
			logger.Debug("User not found for the provided filters")
			execResp.Status = common.ExecFailure
			execResp.FailureReason = failureReasonUserNotFound
			return nil, nil
		} else {
			logger.Debug("Failed to identify user due to error: " + err.Error())
			execResp.Status = common.ExecFailure
			execResp.FailureReason = failureReasonFailedToIdentifyUser
			return nil, nil
		}
	}

	if userID == nil || *userID == "" {
		logger.Debug("User not found for the provided filter")
		execResp.Status = common.ExecFailure
		execResp.FailureReason = failureReasonUserNotFound
		return nil, nil
	}

	return userID, nil
}

// Execute executes the identifying executor logic.
func (i *identifyingExecutor) Execute(ctx *core.NodeContext) (*common.ExecutorResponse, error) {
	logger := i.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Executing identifying executor")

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	// Check if required inputs are provided
	if !i.HasRequiredInputs(ctx, execResp) {
		logger.Debug("Required inputs for identifying executor are not provided")
		execResp.Status = common.ExecUserInputRequired
		return execResp, nil
	}

	userSearchAttributes := map[string]interface{}{}

	for _, inputData := range i.GetRequiredInputs(ctx) {
		if value, ok := ctx.UserInputs[inputData.Identifier]; ok {
			userSearchAttributes[inputData.Identifier] = value
		} else if value, ok := ctx.RuntimeData[inputData.Identifier]; ok {
			// Fallback to RuntimeData if not in UserInputs
			userSearchAttributes[inputData.Identifier] = value
		}
	}

	// Try to identify the user
	userID, err := i.IdentifyUser(userSearchAttributes, execResp)

	if err != nil {
		logger.Debug("Failed to identify user due to error: " + err.Error())
		execResp.Status = common.ExecFailure
		execResp.FailureReason = failureReasonFailedToIdentifyUser
		return execResp, nil
	}

	if userID == nil || *userID == "" {
		logger.Debug("User not found for the provided attributes")
		execResp.Status = common.ExecFailure
		execResp.FailureReason = failureReasonUserNotFound
		return execResp, nil
	}

	// Store the resolved userID in RuntimeData for subsequent executors
	execResp.RuntimeData[userAttributeUserID] = *userID
	execResp.Status = common.ExecComplete

	logger.Debug("Identifying executor completed successfully",
		log.String("userID", log.MaskString(*userID)))

	return execResp, nil
}
