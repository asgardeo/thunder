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

package model

import (
	authncm "github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/flow/common/constants"
	"github.com/asgardeo/thunder/internal/system/log"
)

const (
	userAttributeUserID = "userID"
)

// ExecutorResponse represents the response from an executor
type ExecutorResponse struct {
	Status            constants.ExecutorStatus  `json:"status"`
	RequiredData      []InputData               `json:"required_data,omitempty"`
	AdditionalData    map[string]string         `json:"additional_data,omitempty"`
	RedirectURL       string                    `json:"redirect_url,omitempty"`
	RuntimeData       map[string]string         `json:"runtime_data,omitempty"`
	AuthenticatedUser authncm.AuthenticatedUser `json:"authenticated_user,omitempty"`
	Assertion         string                    `json:"assertion,omitempty"`
	FailureReason     string                    `json:"failure_reason,omitempty"`
	ExecutionRecord   *NodeExecutionRecord      `json:"execution_record,omitempty"`
}

// ExecutorProperties holds the properties of an executor.
type ExecutorProperties struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        constants.ExecutorType `json:"type,omitempty"`
	DisplayName string                 `json:"display_name"`
	Properties  map[string]string      `json:"properties,omitempty"`
}

// ExecutorConfig holds the configuration for an executor.
type ExecutorConfig struct {
	Name       string            `json:"name"`
	IdpName    string            `json:"idp_name,omitempty"`
	Properties map[string]string `json:"properties,omitempty"`
	Executor   ExecutorInterface
}

// ExecutorInterface defines the interface for executors.
type ExecutorInterface interface {
	Execute(ctx *NodeContext) (*ExecutorResponse, error)
	GetID() string
	GetName() string
	GetProperties() ExecutorProperties
	GetDefaultExecutorInputs() []InputData
	GetPrerequisites() []InputData
	CheckInputData(ctx *NodeContext, execResp *ExecutorResponse) bool
	ValidatePrerequisites(ctx *NodeContext, execResp *ExecutorResponse) bool
	GetUserIDFromContext(ctx *NodeContext) (string, error)
	GetRequiredData(ctx *NodeContext) []InputData
}

var _ ExecutorInterface = (*Executor)(nil)

// Executor represents the basic implementation of an executor.
type Executor struct {
	Properties            ExecutorProperties
	DefaultExecutorInputs []InputData
	Prerequisites         []InputData
}

// NewExecutor creates a new instance of Executor with the given properties.
func NewExecutor(id, name string, executorType constants.ExecutorType, defaultInputs []InputData,
	prerequisites []InputData, properties map[string]string) *Executor {
	return &Executor{
		Properties: ExecutorProperties{
			ID:         id,
			Name:       name,
			Type:       executorType,
			Properties: properties,
		},
		DefaultExecutorInputs: defaultInputs,
		Prerequisites:         prerequisites,
	}
}

// GetID returns the ID of the executor.
func (e *Executor) GetID() string {
	return e.Properties.ID
}

// GetName returns the name of the executor.
func (e *Executor) GetName() string {
	return e.Properties.Name
}

// GetProperties returns the properties of the executor.
func (e *Executor) GetProperties() ExecutorProperties {
	return e.Properties
}

// Execute executes the executor logic.
func (e *Executor) Execute(ctx *NodeContext) (*ExecutorResponse, error) {
	// Implement the logic for executing the executor here.
	// This is just a placeholder implementation
	return nil, nil
}

// GetDefaultExecutorInputs returns the default required input data for the executor.
func (e *Executor) GetDefaultExecutorInputs() []InputData {
	return e.DefaultExecutorInputs
}

// GetPrerequisites returns the prerequisites for the executor.
func (e *Executor) GetPrerequisites() []InputData {
	return e.Prerequisites
}

// CheckInputData checks if the required input data is provided in the context.
// If not, it adds the required data to the executor response and returns true.
func (e *Executor) CheckInputData(ctx *NodeContext, execResp *ExecutorResponse) bool {
	requiredData := e.GetRequiredData(ctx)

	if execResp.RequiredData == nil {
		execResp.RequiredData = make([]InputData, 0)
	}
	if len(ctx.UserInputData) == 0 && len(ctx.RuntimeData) == 0 {
		execResp.RequiredData = append(execResp.RequiredData, requiredData...)
		return true
	}

	return e.appendRequiredData(ctx, execResp, requiredData)
}

// ValidatePrerequisites validates whether the prerequisites for the executor are met.
// Returns true if all prerequisites are met, otherwise returns false and updates the executor response.
func (e *Executor) ValidatePrerequisites(ctx *NodeContext, execResp *ExecutorResponse) bool {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "Executor"),
		log.String(log.LoggerKeyExecutorID, e.GetID()),
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
				execResp.Status = constants.ExecFailure
				execResp.FailureReason = "Prerequisite not met: " + prerequisite.Name
				return false
			}
		}
	}

	return true
}

// GetUserIDFromContext retrieves the user ID from the context.
func (e *Executor) GetUserIDFromContext(ctx *NodeContext) (string, error) {
	userID := ctx.AuthenticatedUser.UserID
	if userID == "" {
		userID = ctx.RuntimeData[userAttributeUserID]
	}
	if userID == "" {
		userID = ctx.UserInputData[userAttributeUserID]
	}

	return userID, nil
}

// GetRequiredData returns the required input data for the executor.
// It combines the default executor inputs with the node input data, ensuring no duplicates.
func (e *Executor) GetRequiredData(ctx *NodeContext) []InputData {
	executorReqData := e.GetDefaultExecutorInputs()
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

// appendRequiredData appends the required input data to the executor response if not present in the context.
// returns true if any required data is missing, false otherwise.
func (e *Executor) appendRequiredData(ctx *NodeContext, execResp *ExecutorResponse, requiredData []InputData) bool {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "Executor"),
		log.String(log.LoggerKeyExecutorID, e.GetID()),
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
