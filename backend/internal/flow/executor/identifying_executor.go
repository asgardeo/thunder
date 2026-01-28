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
	"slices"

	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/system/log"

	"github.com/asgardeo/thunder/internal/user"
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
	userService user.UserServiceInterface
	logger      *log.Logger
}

var _ core.ExecutorInterface = (*identifyingExecutor)(nil)
var _ identifyingExecutorInterface = (*identifyingExecutor)(nil)

// newIdentifyingExecutor creates a new instance of IdentifyingExecutor.
func newIdentifyingExecutor(
	name string,
	defaultInputs, prerequisites []common.Input,
	flowFactory core.FlowFactoryInterface,
	userService user.UserServiceInterface,
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
		userService:       userService,
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

	userID, svcErr := i.userService.IdentifyUser(context.TODO(), searchableFilter)
	if svcErr != nil {
		if svcErr.Code == user.ErrorUserNotFound.Code {
			logger.Debug("User not found for the provided filters")
			execResp.Status = common.ExecFailure
			execResp.FailureReason = failureReasonUserNotFound
			return nil, nil
		} else {
			logger.Debug("Failed to identify user due to error: " + svcErr.Error)
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
