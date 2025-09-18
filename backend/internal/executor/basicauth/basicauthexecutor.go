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

// Package basicauth provides the basic authentication executor for handling username and password authentication.
package basicauth

import (
	"encoding/json"
	"fmt"

	authndto "github.com/asgardeo/thunder/internal/authn/dto"
	"github.com/asgardeo/thunder/internal/executor/identify"
	flowconst "github.com/asgardeo/thunder/internal/flow/constants"
	flowmodel "github.com/asgardeo/thunder/internal/flow/model"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/user/service"
)

const (
	loggerComponentName   = "BasicAuthExecutor"
	userAttributeUsername = "username"
	userAttributePassword = "password"
)

// BasicAuthExecutor implements the ExecutorInterface for basic authentication.
type BasicAuthExecutor struct {
	*identify.IdentifyingExecutor
	internal    flowmodel.Executor
	userService service.UserServiceInterface
}

var _ flowmodel.ExecutorInterface = (*BasicAuthExecutor)(nil)

// NewBasicAuthExecutor creates a new instance of BasicAuthExecutor.
func NewBasicAuthExecutor(id, name string, properties map[string]string) *BasicAuthExecutor {
	defaultInputs := []flowmodel.InputData{
		{
			Name:     userAttributeUsername,
			Type:     "string",
			Required: true,
		},
		{
			Name:     userAttributePassword,
			Type:     "string",
			Required: true,
		},
	}
	return &BasicAuthExecutor{
		IdentifyingExecutor: identify.NewIdentifyingExecutor(id, name, properties),
		internal:            *flowmodel.NewExecutor(id, name, defaultInputs, []flowmodel.InputData{}, properties),
		userService:         service.GetUserService(),
	}
}

// GetID returns the ID of the BasicAuthExecutor.
func (b *BasicAuthExecutor) GetID() string {
	return b.internal.GetID()
}

// GetName returns the name of the BasicAuthExecutor.
func (b *BasicAuthExecutor) GetName() string {
	return b.internal.GetName()
}

// GetProperties returns the properties of the BasicAuthExecutor.
func (b *BasicAuthExecutor) GetProperties() flowmodel.ExecutorProperties {
	return b.internal.Properties
}

// Execute executes the basic authentication logic.
func (b *BasicAuthExecutor) Execute(ctx *flowmodel.NodeContext) (*flowmodel.ExecutorResponse, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorID, b.GetID()),
		log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Executing basic authentication executor")

	execResp := &flowmodel.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	// Validate for the required input data.
	if b.CheckInputData(ctx, execResp) {
		// If required input data is not provided, return incomplete status.
		logger.Debug("Required input data for basic authentication executor is not provided")
		execResp.Status = flowconst.ExecUserInputRequired
		return execResp, nil
	}

	// TODO: Should handle client errors here. Service should return a ServiceError and
	//  client errors should be appended as a failure.
	//  For the moment handling returned error as a authentication failure.
	authenticatedUser, err := b.getAuthenticatedUser(ctx, execResp)
	if err != nil {
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = "Failed to authenticate user: " + err.Error()
		return execResp, nil
	}
	if execResp.Status == flowconst.ExecFailure {
		return execResp, nil
	}
	if authenticatedUser == nil {
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = "Authenticated user not found."
		return execResp, nil
	}
	if !authenticatedUser.IsAuthenticated && ctx.FlowType != flowconst.FlowTypeRegistration {
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = "User authentication failed."
		return execResp, nil
	}

	execResp.AuthenticatedUser = *authenticatedUser
	execResp.Status = flowconst.ExecComplete

	logger.Debug("Basic authentication executor execution completed",
		log.String("status", string(execResp.Status)),
		log.Bool("isAuthenticated", execResp.AuthenticatedUser.IsAuthenticated))

	return execResp, nil
}

// GetDefaultExecutorInputs returns the default required input data for the BasicAuthExecutor.
func (b *BasicAuthExecutor) GetDefaultExecutorInputs() []flowmodel.InputData {
	return b.internal.DefaultExecutorInputs
}

// GetPrerequisites returns the prerequisites for the BasicAuthExecutor.
func (b *BasicAuthExecutor) GetPrerequisites() []flowmodel.InputData {
	return b.internal.Prerequisites
}

// CheckInputData checks if the required input data is provided in the context.
func (b *BasicAuthExecutor) CheckInputData(ctx *flowmodel.NodeContext, execResp *flowmodel.ExecutorResponse) bool {
	return b.internal.CheckInputData(ctx, execResp)
}

// ValidatePrerequisites validates whether the prerequisites for the BasicAuthExecutor are met.
func (b *BasicAuthExecutor) ValidatePrerequisites(ctx *flowmodel.NodeContext,
	execResp *flowmodel.ExecutorResponse) bool {
	return b.internal.ValidatePrerequisites(ctx, execResp)
}

// GetUserIDFromContext retrieves the user ID from the context.
func (b *BasicAuthExecutor) GetUserIDFromContext(ctx *flowmodel.NodeContext) (string, error) {
	return b.internal.GetUserIDFromContext(ctx)
}

// GetRequiredData returns the required input data for the BasicAuthExecutor.
func (b *BasicAuthExecutor) GetRequiredData(ctx *flowmodel.NodeContext) []flowmodel.InputData {
	return b.internal.GetRequiredData(ctx)
}

// getAuthenticatedUser perform authentication based on the provided username and password and return
// authenticated user details.
func (b *BasicAuthExecutor) getAuthenticatedUser(ctx *flowmodel.NodeContext,
	execResp *flowmodel.ExecutorResponse) (*authndto.AuthenticatedUser, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorID, b.GetID()))

	username := ctx.UserInputData[userAttributeUsername]
	filters := map[string]interface{}{userAttributeUsername: username}
	userID, err := b.IdentifyUser(filters, execResp)
	if err != nil {
		return nil, err
	}

	// Handle registration flows.
	if ctx.FlowType == flowconst.FlowTypeRegistration {
		if execResp.Status == flowconst.ExecFailure {
			if execResp.FailureReason == "User not found" {
				logger.Debug("User not found for the provided username. Proceeding with registration flow.")
				execResp.Status = flowconst.ExecComplete

				return &authndto.AuthenticatedUser{
					IsAuthenticated: false,
					Attributes: map[string]interface{}{
						userAttributeUsername: username,
					},
				}, nil
			}
			return nil, err
		}

		// At this point, a unique user is found in the system. Hence fail the execution.
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = "User already exists with the provided username."
		return nil, nil
	}

	if execResp.Status == flowconst.ExecFailure {
		return nil, nil
	}

	credentials := map[string]interface{}{
		userAttributePassword: ctx.UserInputData[userAttributePassword],
	}

	user, svcErr := b.userService.VerifyUser(*userID, credentials)
	if svcErr != nil {
		logger.Error("Failed to verify user credentials",
			log.String("userID", *userID),
			log.String("error", svcErr.Error),
			log.String("code", svcErr.Code))
		return nil, fmt.Errorf("failed to verify user credentials: %s", svcErr.Error)
	}

	var authenticatedUser authndto.AuthenticatedUser
	if user == nil {
		authenticatedUser = authndto.AuthenticatedUser{
			IsAuthenticated: false,
		}
	} else {
		var attrs map[string]interface{}
		if err := json.Unmarshal(user.Attributes, &attrs); err != nil {
			logger.Error("Failed to unmarshal user attributes", log.Error(err))
			return nil, err
		}

		authenticatedUser = authndto.AuthenticatedUser{
			IsAuthenticated: true,
			UserID:          user.ID,
			Attributes:      attrs,
		}
	}
	return &authenticatedUser, nil
}
