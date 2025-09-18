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

// Package authassert provides the authentication assertion executor for a flow.
package authassert

import (
	"encoding/json"
	"errors"

	flowconst "github.com/asgardeo/thunder/internal/flow/constants"
	flowmodel "github.com/asgardeo/thunder/internal/flow/model"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/jwt"
	"github.com/asgardeo/thunder/internal/system/log"
	usermodel "github.com/asgardeo/thunder/internal/user/model"
	userservice "github.com/asgardeo/thunder/internal/user/service"
)

const loggerComponentName = "AuthAssertExecutor"

// AuthAssertExecutor is an executor that handles authentication assertions in the flow.
type AuthAssertExecutor struct {
	internal    flowmodel.Executor
	JWTService  jwt.JWTServiceInterface
	UserService userservice.UserServiceInterface
}

var _ flowmodel.ExecutorInterface = (*AuthAssertExecutor)(nil)

// NewAuthAssertExecutor creates a new instance of AuthAssertExecutor.
func NewAuthAssertExecutor(id, name string, properties map[string]string) *AuthAssertExecutor {
	return &AuthAssertExecutor{
		internal:    *flowmodel.NewExecutor(id, name, []flowmodel.InputData{}, []flowmodel.InputData{}, properties),
		JWTService:  jwt.GetJWTService(),
		UserService: userservice.GetUserService(),
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
		token, err := a.generateAuthAssertion(ctx, logger)
		if err != nil {
			return nil, err
		}

		logger.Debug("Generated JWT token for authentication assertion")

		execResp.Status = flowconst.ExecComplete
		execResp.Assertion = token
	} else {
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = "User is not authenticated"
	}

	logger.Debug("Authentication assertion executor execution completed",
		log.String("status", string(execResp.Status)))

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

// generateAuthAssertion generates the authentication assertion token.
func (a *AuthAssertExecutor) generateAuthAssertion(ctx *flowmodel.NodeContext, logger *log.Logger) (string, error) {
	tokenSub := ""
	if ctx.AuthenticatedUser.UserID != "" {
		tokenSub = ctx.AuthenticatedUser.UserID
	}

	jwtClaims := make(map[string]interface{})
	jwtConfig := config.GetThunderRuntime().Config.OAuth.JWT
	iss := ""
	validityPeriod := int64(0)

	if ctx.Application.Token != nil {
		iss = ctx.Application.Token.Issuer
		validityPeriod = ctx.Application.Token.ValidityPeriod
	}
	if iss == "" {
		iss = jwtConfig.Issuer
	}
	if validityPeriod == 0 {
		validityPeriod = jwtConfig.ValidityPeriod
	}

	if ctx.Application.Token != nil && len(ctx.Application.Token.UserAttributes) > 0 &&
		ctx.AuthenticatedUser.UserID != "" {
		var user *usermodel.User
		var attrs map[string]interface{}

		for _, attr := range ctx.Application.Token.UserAttributes {
			// check for the attribute in authenticated user attributes
			if val, ok := ctx.AuthenticatedUser.Attributes[attr]; ok {
				jwtClaims[attr] = val
				continue
			}

			// fetch user details only once
			if user == nil {
				var err error
				user, attrs, err = a.getUserAttributes(ctx.AuthenticatedUser.UserID, logger)
				if err != nil {
					return "", err
				}
			}

			// check for the attribute in user store attributes
			if val, ok := attrs[attr]; ok {
				jwtClaims[attr] = val
			}
		}
	}

	token, _, err := a.JWTService.GenerateJWT(tokenSub, ctx.AppID, iss, validityPeriod, jwtClaims)
	if err != nil {
		logger.Error("Failed to generate JWT token", log.Error(err))
		return "", errors.New("failed to generate JWT token: " + err.Error())
	}

	return token, nil
}

// getUserAttributes retrieves user details and unmarshal the attributes.
func (a *AuthAssertExecutor) getUserAttributes(userID string, logger *log.Logger) (
	*usermodel.User, map[string]interface{}, error) {
	var svcErr *serviceerror.ServiceError
	user, svcErr := a.UserService.GetUser(userID)
	if svcErr != nil {
		logger.Error("Failed to fetch user attributes",
			log.String("userID", userID), log.Any("error", svcErr))
		return nil, nil, errors.New("something went wrong while fetching user attributes: " +
			svcErr.ErrorDescription)
	}

	var attrs map[string]interface{}
	if err := json.Unmarshal(user.Attributes, &attrs); err != nil {
		logger.Error("Failed to unmarshal user attributes", log.String("userID", userID),
			log.Error(err))
		return nil, nil, errors.New("something went wrong while unmarshalling user attributes: " + err.Error())
	}

	return user, attrs, nil
}
