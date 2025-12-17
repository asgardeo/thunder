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

// Package executor defines executors that can be used during flow executions for authentication, registration
// and other purposes.
package executor

import (
	"encoding/json"
	"errors"

	authncm "github.com/asgardeo/thunder/internal/authn/common"
	authncreds "github.com/asgardeo/thunder/internal/authn/credentials"
	flowcm "github.com/asgardeo/thunder/internal/flow/common"
	flowcore "github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/user"
)

const (
	basicAuthLoggerComponentName = "BasicAuthExecutor"
	inputDataTypePassword        = "PASSWORD_INPUT"
)

// basicAuthExecutor implements the ExecutorInterface for basic authentication.
type basicAuthExecutor struct {
	flowcore.ExecutorInterface
	identifyingExecutorInterface
	credsAuthSvc authncreds.CredentialsAuthnServiceInterface
	logger       *log.Logger
}

var _ flowcore.ExecutorInterface = (*basicAuthExecutor)(nil)
var _ identifyingExecutorInterface = (*basicAuthExecutor)(nil)

// newBasicAuthExecutor creates a new instance of BasicAuthExecutor.
func newBasicAuthExecutor(
	flowFactory flowcore.FlowFactoryInterface,
	userService user.UserServiceInterface,
	credsAuthSvc authncreds.CredentialsAuthnServiceInterface,
) *basicAuthExecutor {
	defaultInputs := []flowcm.InputData{
		{
			Name:     userAttributeUsername,
			Type:     "string",
			Required: true,
		},
		{
			Name:     userAttributePassword,
			Type:     inputDataTypePassword,
			Required: true,
		},
	}

	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, basicAuthLoggerComponentName),
		log.String(log.LoggerKeyExecutorName, ExecutorNameBasicAuth))

	identifyExec := newIdentifyingExecutor(ExecutorNameBasicAuth, defaultInputs, []flowcm.InputData{},
		flowFactory, userService)
	base := flowFactory.CreateExecutor(ExecutorNameBasicAuth, flowcm.ExecutorTypeAuthentication,
		defaultInputs, []flowcm.InputData{})

	return &basicAuthExecutor{
		ExecutorInterface:            base,
		identifyingExecutorInterface: identifyExec,
		credsAuthSvc:                 credsAuthSvc,
		logger:                       logger,
	}
}

// Execute executes the basic authentication logic.
func (b *basicAuthExecutor) Execute(ctx *flowcore.NodeContext) (*flowcm.ExecutorResponse, error) {
	logger := b.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Executing basic authentication executor")

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	// Validate for the required input data.
	if b.CheckInputData(ctx, execResp) {
		// If required input data is not provided, return incomplete status.
		logger.Debug("Required input data for basic authentication executor is not provided")
		execResp.Status = flowcm.ExecUserInputRequired
		return execResp, nil
	}

	// TODO: Should handle client errors here. Service should return a ServiceError and
	//  client errors should be appended as a failure.
	//  For the moment handling returned error as a authentication failure.
	authenticatedUser, err := b.getAuthenticatedUser(ctx, execResp)
	if err != nil {
		execResp.Status = flowcm.ExecFailure
		execResp.FailureReason = "Failed to authenticate user: " + err.Error()
		return execResp, nil
	}
	if execResp.Status == flowcm.ExecFailure {
		return execResp, nil
	}
	if authenticatedUser == nil {
		execResp.Status = flowcm.ExecFailure
		execResp.FailureReason = "Authenticated user not found."
		return execResp, nil
	}
	if !authenticatedUser.IsAuthenticated && ctx.FlowType != flowcm.FlowTypeRegistration {
		execResp.Status = flowcm.ExecFailure
		execResp.FailureReason = "User authentication failed."
		return execResp, nil
	}

	execResp.AuthenticatedUser = *authenticatedUser
	execResp.Status = flowcm.ExecComplete

	logger.Debug("Basic authentication executor execution completed",
		log.String("status", string(execResp.Status)),
		log.Bool("isAuthenticated", execResp.AuthenticatedUser.IsAuthenticated))

	return execResp, nil
}

// getAuthenticatedUser perform authentication based on the provided identifying and
// credential attributes and returns the authenticated user details.
func (b *basicAuthExecutor) getAuthenticatedUser(ctx *flowcore.NodeContext,
	execResp *flowcm.ExecutorResponse) (*authncm.AuthenticatedUser, error) {
	logger := b.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))

	userSearchAttributes := map[string]interface{}{}
	userAuthenticateAttributes := map[string]interface{}{}

	for _, inputData := range b.GetRequiredData(ctx) {
		if value, ok := ctx.UserInputData[inputData.Name]; ok {
			if inputData.Type != inputDataTypePassword {
				userSearchAttributes[inputData.Name] = value
			}
			userAuthenticateAttributes[inputData.Name] = value
		}
	}

	// Identify the user based on the provided attributes.
	userID, err := b.IdentifyUser(userSearchAttributes, execResp)
	if err != nil {
		return nil, err
	}

	// Handle registration flows.
	if ctx.FlowType == flowcm.FlowTypeRegistration {
		if execResp.Status == flowcm.ExecFailure {
			if execResp.FailureReason == failureReasonUserNotFound {
				logger.Debug("User not found for the provided attributes. Proceeding with registration flow.")
				execResp.Status = flowcm.ExecComplete

				return &authncm.AuthenticatedUser{
					IsAuthenticated: false,
					Attributes:      userSearchAttributes,
				}, nil
			}
			return nil, err
		}

		// At this point, a unique user is found in the system. Hence fail the execution.
		execResp.Status = flowcm.ExecFailure
		execResp.FailureReason = "User already exists with the provided attributes."
		return nil, nil
	}

	if execResp.Status == flowcm.ExecFailure {
		return nil, nil
	}

	// Authenticate the user based on all the provided attributes including credentials.
	user, svcErr := b.credsAuthSvc.Authenticate(userAuthenticateAttributes)
	if svcErr != nil {
		if svcErr.Type == serviceerror.ClientErrorType {
			execResp.Status = flowcm.ExecFailure
			execResp.FailureReason = "Failed to authenticate user: " + svcErr.ErrorDescription
			return nil, nil
		}
		logger.Error("Failed to authenticate user", log.String("userID", *userID),
			log.String("errorCode", svcErr.Code), log.String("errorDescription", svcErr.ErrorDescription))
		return nil, errors.New("failed to authenticate user")
	}

	var attrs map[string]interface{}
	if err := json.Unmarshal(user.Attributes, &attrs); err != nil {
		logger.Error("Failed to unmarshal user attributes", log.Error(err))
		return nil, err
	}

	authenticatedUser := authncm.AuthenticatedUser{
		IsAuthenticated:    true,
		UserID:             user.ID,
		OrganizationUnitID: user.OrganizationUnit,
		UserType:           user.Type,
		Attributes:         attrs,
	}
	return &authenticatedUser, nil
}
