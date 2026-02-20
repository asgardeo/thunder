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
	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/observability"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/userprovider"
)

// basicAuthExecutor implements the ExecutorInterface for basic authentication.
type basicAuthExecutor struct {
	core.ExecutorInterface
	identifyingExecutorInterface
	userProvider     userprovider.UserProviderInterface
	credsAuthSvc     authncreds.CredentialsAuthnServiceInterface
	observabilitySvc observability.ObservabilityServiceInterface
	logger           *log.Logger
}

var _ core.ExecutorInterface = (*basicAuthExecutor)(nil)
var _ identifyingExecutorInterface = (*basicAuthExecutor)(nil)

// newBasicAuthExecutor creates a new instance of BasicAuthExecutor.
func newBasicAuthExecutor(
	flowFactory core.FlowFactoryInterface,
	userProvider userprovider.UserProviderInterface,
	credsAuthSvc authncreds.CredentialsAuthnServiceInterface,
	observabilitySvc observability.ObservabilityServiceInterface,
) *basicAuthExecutor {
	defaultInputs := []common.Input{
		{
			Identifier: userAttributeUsername,
			Type:       common.InputTypeText,
			Required:   true,
		},
		{
			Identifier: userAttributePassword,
			Type:       common.InputTypePassword,
			Required:   true,
		},
	}

	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "BasicAuthExecutor"),
		log.String(log.LoggerKeyExecutorName, ExecutorNameBasicAuth))

	identifyExec := newIdentifyingExecutor(ExecutorNameBasicAuth, defaultInputs, []common.Input{},
		flowFactory, userProvider)
	base := flowFactory.CreateExecutor(ExecutorNameBasicAuth, common.ExecutorTypeAuthentication,
		defaultInputs, []common.Input{})

	return &basicAuthExecutor{
		ExecutorInterface:            base,
		identifyingExecutorInterface: identifyExec,
		userProvider:                 userProvider,
		credsAuthSvc:                 credsAuthSvc,
		observabilitySvc:             observabilitySvc,
		logger:                       logger,
	}
}

// Execute executes the basic authentication logic.
func (b *basicAuthExecutor) Execute(ctx *core.NodeContext) (*common.ExecutorResponse, error) {
	logger := b.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Executing basic authentication executor")

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	if !b.HasRequiredInputs(ctx, execResp) {
		logger.Debug("Required inputs for basic authentication executor is not provided")
		execResp.Status = common.ExecUserInputRequired
		return execResp, nil
	}

	// TODO: Should handle client errors here. Service should return a ServiceError and
	//  client errors should be appended as a failure.
	//  For the moment handling returned error as a authentication failure.
	authenticatedUser, err := b.getAuthenticatedUser(ctx, execResp)
	if err != nil {
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "Failed to authenticate user: " + err.Error()
		return execResp, nil
	}
	if execResp.Status == common.ExecFailure {
		return execResp, nil
	}
	if authenticatedUser == nil {
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "Authenticated user not found."
		return execResp, nil
	}
	if !authenticatedUser.IsAuthenticated && ctx.FlowType != common.FlowTypeRegistration {
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "User authentication failed."
		return execResp, nil
	}

	execResp.AuthenticatedUser = *authenticatedUser
	execResp.Status = common.ExecComplete

	logger.Debug("Basic authentication executor execution completed",
		log.String("status", string(execResp.Status)),
		log.Bool("isAuthenticated", execResp.AuthenticatedUser.IsAuthenticated))

	return execResp, nil
}

// getAuthenticatedUser perform authentication based on the provided identifying and
// credential attributes and returns the authenticated user details.
func (b *basicAuthExecutor) getAuthenticatedUser(ctx *core.NodeContext,
	execResp *common.ExecutorResponse) (*authncm.AuthenticatedUser, error) {
	logger := b.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))

	userIdentifiers := map[string]interface{}{}
	userCredentials := map[string]interface{}{}

	for _, inputData := range b.GetRequiredInputs(ctx) {
		if value, ok := ctx.UserInputs[inputData.Identifier]; ok {
			if inputData.IsSensitive() {
				userCredentials[inputData.Identifier] = value
			} else {
				userIdentifiers[inputData.Identifier] = value
			}
		}
	}

	// For registration flows, only check if user exists.
	if ctx.FlowType == common.FlowTypeRegistration {
		_, err := b.IdentifyUser(userIdentifiers, execResp)
		if err != nil {
			return nil, err
		}
		if execResp.Status == common.ExecFailure {
			if execResp.FailureReason == failureReasonUserNotFound {
				logger.Debug("User not found for the provided attributes. Proceeding with registration flow.")
				execResp.Status = common.ExecComplete
				return &authncm.AuthenticatedUser{
					IsAuthenticated: false,
					Attributes:      userIdentifiers,
				}, nil
			}
			return nil, nil
		}
		// User found - fail registration.
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "User already exists with the provided attributes."
		return nil, nil
	}

	// For authentication flows, call Authenticate directly.
	authnResult, svcErr := b.credsAuthSvc.Authenticate(userIdentifiers, userCredentials, nil)
	if svcErr != nil {
		if svcErr.Type == serviceerror.ClientErrorType {
			execResp.Status = common.ExecFailure
			execResp.FailureReason = "Failed to authenticate user: " + svcErr.ErrorDescription
			return nil, nil
		}
		logger.Error("Failed to authenticate user",
			log.String("errorCode", svcErr.Code), log.String("errorDescription", svcErr.ErrorDescription))
		return nil, errors.New("failed to authenticate user")
	}

	// Try to retrieve the user and get the attributes
	userAttributes := map[string]interface{}{}
	user, err := b.userProvider.GetUser(authnResult.UserID)

	if err != nil {
		if err.Code != userprovider.ErrorCodeNotImplemented {
			logger.Error("Failed to get user attributes", log.Error(err))
			return nil, errors.New("failed to get user attributes")
		}
		logger.Debug("User provider is not implemented. User attributes will be empty.")
	}

	if err == nil && user != nil {
		if err := json.Unmarshal(user.Attributes, &userAttributes); err != nil {
			logger.Error("Failed to unmarshal user attributes", log.Error(err))
			return nil, errors.New("failed to unmarshal user attributes")
		}
	}

	return &authncm.AuthenticatedUser{
		IsAuthenticated:     true,
		UserID:              authnResult.UserID,
		OrganizationUnitID:  authnResult.OrganizationUnitID,
		UserType:            authnResult.UserType,
		Attributes:          userAttributes,
		AvailableAttributes: authnResult.AvailableAttributes,
		Token:               authnResult.Token,
	}, nil
}
