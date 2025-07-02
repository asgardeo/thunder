/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
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

// Package authassert provides the authentication assertion executor for a flow.
package authassert

import (
	"errors"

	flowconst "github.com/asgardeo/thunder/internal/flow/constants"
	flowmodel "github.com/asgardeo/thunder/internal/flow/model"
	"github.com/asgardeo/thunder/internal/oauth/jwt"
	"github.com/asgardeo/thunder/internal/system/log"
)

const loggerComponentName = "AuthAssertExecutor"

// AuthAssertExecutor is an executor that handles authentication assertions in the flow.
type AuthAssertExecutor struct {
	internal flowmodel.Executor
}

var _ flowmodel.ExecutorInterface = (*AuthAssertExecutor)(nil)

// NewAuthAssertExecutor creates a new instance of AuthAssertExecutor.
func NewAuthAssertExecutor(id, name string, properties map[string]string) *AuthAssertExecutor {
	return &AuthAssertExecutor{
		internal: *flowmodel.NewExecutor(id, name, []flowmodel.InputData{}, []flowmodel.InputData{}, properties),
	}
}

// GetID returns the ID of the AuthAssertExecutor.
func (a *AuthAssertExecutor) GetID() string {
	return a.internal.GetID()
}

// GetName returns the name of the AuthAssertExecutor.
func (a *AuthAssertExecutor) GetName() string {
	return a.internal.GetName()
}

// GetProperties returns the properties of the AuthAssertExecutor.
func (a *AuthAssertExecutor) GetProperties() flowmodel.ExecutorProperties {
	return a.internal.Properties
}

// Execute executes the authentication assertion logic.
func (a *AuthAssertExecutor) Execute(ctx *flowmodel.NodeContext) (*flowmodel.ExecutorResponse, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorID, a.GetID()),
		log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Executing authentication assertion executor")

	execResp := &flowmodel.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	if ctx.AuthenticatedUser.IsAuthenticated {
		tokenSub := ""
		if ctx.AuthenticatedUser.UserID != "" {
			tokenSub = ctx.AuthenticatedUser.UserID
		}

		token, _, err := jwt.GenerateJWT(tokenSub, ctx.AppID, jwt.GetJWTTokenValidityPeriod(),
			ctx.AuthenticatedUser.Attributes)
		if err != nil {
			logger.Error("Failed to generate JWT token", log.Error(err))
			return nil, errors.New("failed to generate JWT token: " + err.Error())
		}

		logger.Debug("Generated JWT token for authentication assertion")

		execResp.Status = flowconst.ExecComplete
		execResp.Assertion = token
	} else {
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = "User is not authenticated"
	}

	logger.Debug("Authentication assertion executor execution completed", log.String("status", string(execResp.Status)))

	return execResp, nil
}

// GetDefaultExecutorInputs returns the default required input data for the AuthAssertExecutor.
func (a *AuthAssertExecutor) GetDefaultExecutorInputs() []flowmodel.InputData {
	return a.internal.GetDefaultExecutorInputs()
}

// GetPrerequisites returns the prerequisites for the AuthAssertExecutor.
func (a *AuthAssertExecutor) GetPrerequisites() []flowmodel.InputData {
	return a.internal.GetPrerequisites()
}

// CheckInputData checks if the required input data is provided in the context.
func (a *AuthAssertExecutor) CheckInputData(ctx *flowmodel.NodeContext, execResp *flowmodel.ExecutorResponse) bool {
	return a.internal.CheckInputData(ctx, execResp)
}

// ValidatePrerequisites validates whether the prerequisites for the AuthAssertExecutor are met.
func (a *AuthAssertExecutor) ValidatePrerequisites(ctx *flowmodel.NodeContext,
	execResp *flowmodel.ExecutorResponse) bool {
	return a.internal.ValidatePrerequisites(ctx, execResp)
}

// GetUserIDFromContext retrieves the user ID from the context.
func (a *AuthAssertExecutor) GetUserIDFromContext(ctx *flowmodel.NodeContext) (string, error) {
	return a.internal.GetUserIDFromContext(ctx)
}

// GetRequiredData returns the required input data for the AuthAssertExecutor.
func (a *AuthAssertExecutor) GetRequiredData(ctx *flowmodel.NodeContext) []flowmodel.InputData {
	return a.internal.GetRequiredData(ctx)
}
