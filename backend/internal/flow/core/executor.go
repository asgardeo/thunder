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

package core

import (
	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/system/log"
)

const (
	userAttributeUserID = "userID"
)

// ExecutorInterface defines the interface for executors.
type ExecutorInterface interface {
	Execute(ctx *NodeContext) (*common.ExecutorResponse, error)
	GetName() string
	GetType() common.ExecutorType
	GetDefaultExecutorInputs() []common.InputData
	GetPrerequisites() []common.InputData
	CheckInputData(ctx *NodeContext, execResp *common.ExecutorResponse) bool
	ValidatePrerequisites(ctx *NodeContext, execResp *common.ExecutorResponse) bool
	GetUserIDFromContext(ctx *NodeContext) string
	GetRequiredData(ctx *NodeContext) []common.InputData
}

// executor represents the basic implementation of an executor.
type executor struct {
	Name                  string
	Type                  common.ExecutorType
	DefaultExecutorInputs []common.InputData
	Prerequisites         []common.InputData
}

var _ ExecutorInterface = (*executor)(nil)

// newExecutor creates a new instance of Executor with the given properties.
func newExecutor(name string, executorType common.ExecutorType, defaultInputs []common.InputData,
	prerequisites []common.InputData) ExecutorInterface {
	return &executor{
		Name:                  name,
		Type:                  executorType,
		DefaultExecutorInputs: defaultInputs,
		Prerequisites:         prerequisites,
	}
}

// GetName returns the name of the executor.
func (e *executor) GetName() string {
	return e.Name
}

// GetType returns the type of the executor.
func (e *executor) GetType() common.ExecutorType {
	return e.Type
}

// Execute executes the executor logic.
func (e *executor) Execute(ctx *NodeContext) (*common.ExecutorResponse, error) {
	// Implement the logic for executing the executor here.
	// This is just a placeholder implementation
	return nil, nil
}

// GetDefaultExecutorInputs returns the default required input data for the executor.
func (e *executor) GetDefaultExecutorInputs() []common.InputData {
	return e.DefaultExecutorInputs
}

// GetPrerequisites returns the prerequisites for the executor.
func (e *executor) GetPrerequisites() []common.InputData {
	return e.Prerequisites
}

// CheckInputData checks if the required input data is provided in the context.
// If not, it adds the required data to the executor response and returns true.
func (e *executor) CheckInputData(ctx *NodeContext, execResp *common.ExecutorResponse) bool {
	requiredData := e.GetRequiredData(ctx)

	if execResp.RequiredData == nil {
		execResp.RequiredData = make([]common.InputData, 0)
	}
	if len(ctx.UserInputData) == 0 && len(ctx.RuntimeData) == 0 {
		execResp.RequiredData = append(execResp.RequiredData, requiredData...)
		return true
	}

	return e.appendRequiredData(ctx, execResp, requiredData)
}

// ValidatePrerequisites validates whether the prerequisites for the executor are met.
// Returns true if all prerequisites are met, otherwise returns false and updates the executor response.
func (e *executor) ValidatePrerequisites(ctx *NodeContext, execResp *common.ExecutorResponse) bool {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "Executor"),
		log.String(log.LoggerKeyExecutorName, e.GetName()),
		log.String(log.LoggerKeyFlowID, ctx.FlowID))

	prerequisites := e.GetPrerequisites()
	if len(prerequisites) == 0 {
		return true
	}

	for _, prerequisite := range prerequisites {
		// Handle userID prerequisite specifically.
		if prerequisite.Name == userAttributeUserID {
			userID := ctx.AuthenticatedUser.UserID
			if userID != "" {
				continue
			}
		}

		if _, ok := ctx.UserInputData[prerequisite.Name]; !ok {
			if _, ok := ctx.RuntimeData[prerequisite.Name]; !ok {
				logger.Debug("Prerequisite not met for the executor", log.String("name", prerequisite.Name))
				execResp.Status = common.ExecFailure
				execResp.FailureReason = "Prerequisite not met: " + prerequisite.Name
				return false
			}
		}
	}

	return true
}

// GetUserIDFromContext retrieves the user ID from the context.
func (e *executor) GetUserIDFromContext(ctx *NodeContext) string {
	userID := ctx.AuthenticatedUser.UserID
	if userID == "" {
		userID = ctx.RuntimeData[userAttributeUserID]
	}
	if userID == "" {
		userID = ctx.UserInputData[userAttributeUserID]
	}

	return userID
}

// GetRequiredData returns the required input data for the executor.
func (e *executor) GetRequiredData(ctx *NodeContext) []common.InputData {
	requiredData := ctx.NodeInputData

	if len(requiredData) > 0 {
		return requiredData
	}

	return e.GetDefaultExecutorInputs()
}

// appendRequiredData appends the required input data to the executor response if not present
// in the context. Returns true if any required data is missing, false otherwise.
func (e *executor) appendRequiredData(ctx *NodeContext, execResp *common.ExecutorResponse,
	requiredData []common.InputData) bool {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "Executor"),
		log.String(log.LoggerKeyExecutorName, e.GetName()),
		log.String(log.LoggerKeyFlowID, ctx.FlowID))

	requireData := false
	for _, inputData := range requiredData {
		if _, ok := ctx.UserInputData[inputData.Name]; !ok {
			// If the input data is available in runtime data, skip adding it to the required data.
			if _, ok := ctx.RuntimeData[inputData.Name]; ok {
				logger.Debug("Input data available in runtime data, skipping required data addition",
					log.String("inputDataName", inputData.Name), log.Bool("isRequired", inputData.Required))
				continue
			}

			requireData = true
			execResp.RequiredData = append(execResp.RequiredData, inputData)
			logger.Debug("Input data not available in the context",
				log.String("inputDataName", inputData.Name), log.Bool("isRequired", inputData.Required))
		}
	}

	return requireData
}
