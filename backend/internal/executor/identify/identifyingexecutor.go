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

// Package identify provides the IdentifyingExecutor for identifying users based on provided attributes.
package identify

import (
	"slices"

	flowconst "github.com/asgardeo/thunder/internal/flow/common/constants"
	flowmodel "github.com/asgardeo/thunder/internal/flow/common/model"
	"github.com/asgardeo/thunder/internal/system/log"

	"github.com/asgardeo/thunder/internal/user"
)

const loggerComponentName = "IdentifyingExecutor"

var nonSearchableAttributes = []string{"password", "code", "nonce", "otp"}

// IdentifyingExecutor implements the ExecutorInterface for identifying users based on provided attributes.
type IdentifyingExecutor struct {
	internal    flowmodel.Executor
	userService user.UserServiceInterface
}

// NewIdentifyingExecutor creates a new instance of IdentifyingExecutor.
func NewIdentifyingExecutor(id, name string, properties map[string]string) *IdentifyingExecutor {
	return &IdentifyingExecutor{
		internal: *flowmodel.NewExecutor(id, name, flowconst.ExecutorTypeUtility,
			[]flowmodel.InputData{}, []flowmodel.InputData{}, properties),
		userService: user.GetUserService(),
	}
}

// IdentifyUser identifies a user based on the provided attributes.
func (i *IdentifyingExecutor) IdentifyUser(filters map[string]interface{},
	execResp *flowmodel.ExecutorResponse) (*string, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Identifying user with filters")

	// filter out non-searchable attributes
	var searchableFilter = make(map[string]interface{})
	for key, value := range filters {
		if !slices.Contains(nonSearchableAttributes, key) {
			searchableFilter[key] = value
		}
	}

	userID, svcErr := i.userService.IdentifyUser(searchableFilter)
	if svcErr != nil {
		if svcErr.Code == user.ErrorUserNotFound.Code {
			logger.Debug("User not found for the provided filters")
			execResp.Status = flowconst.ExecFailure
			execResp.FailureReason = "User not found"
			return nil, nil
		} else {
			logger.Debug("Failed to identify user due to error: " + svcErr.Error)
			execResp.Status = flowconst.ExecFailure
			execResp.FailureReason = "Failed to identify user"
			return nil, nil
		}
	}

	if userID == nil || *userID == "" {
		logger.Debug("User not found for the provided filter")
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = "User not found"
		return nil, nil
	}

	return userID, nil
}
