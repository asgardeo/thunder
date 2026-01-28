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
	"encoding/json"
	"errors"
	"sort"

	"github.com/asgardeo/thunder/internal/authn/assert"
	authncm "github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/jwt"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/user"
)

const (
	authAssertLoggerComponentName = "AuthAssertExecutor"
)

// authAssertExecutor is an executor that handles authentication assertions in the flow.
type authAssertExecutor struct {
	core.ExecutorInterface
	jwtService          jwt.JWTServiceInterface
	userService         user.UserServiceInterface
	ouService           ou.OrganizationUnitServiceInterface
	authAssertGenerator assert.AuthAssertGeneratorInterface
	logger              *log.Logger
}

var _ core.ExecutorInterface = (*authAssertExecutor)(nil)

// newAuthAssertExecutor creates a new instance of AuthAssertExecutor.
func newAuthAssertExecutor(
	flowFactory core.FlowFactoryInterface,
	jwtService jwt.JWTServiceInterface,
	userService user.UserServiceInterface,
	ouService ou.OrganizationUnitServiceInterface,
	assertGenerator assert.AuthAssertGeneratorInterface,
) *authAssertExecutor {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, authAssertLoggerComponentName),
		log.String(log.LoggerKeyExecutorName, ExecutorNameAuthAssert))

	base := flowFactory.CreateExecutor(ExecutorNameAuthAssert, common.ExecutorTypeUtility,
		[]common.Input{}, []common.Input{})

	return &authAssertExecutor{
		ExecutorInterface:   base,
		jwtService:          jwtService,
		userService:         userService,
		ouService:           ouService,
		authAssertGenerator: assertGenerator,
		logger:              logger,
	}
}

// Execute executes the authentication assertion logic.
func (a *authAssertExecutor) Execute(ctx *core.NodeContext) (*common.ExecutorResponse, error) {
	logger := a.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Executing authentication assertion executor")

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	if ctx.AuthenticatedUser.IsAuthenticated {
		token, err := a.generateAuthAssertion(ctx, logger)
		if err != nil {
			return nil, err
		}

		logger.Debug("Generated JWT token for authentication assertion")

		execResp.Status = common.ExecComplete
		execResp.Assertion = token
	} else {
		execResp.Status = common.ExecFailure
		execResp.FailureReason = failureReasonUserNotAuthenticated
	}

	logger.Debug("Authentication assertion executor execution completed",
		log.String("status", string(execResp.Status)))

	return execResp, nil
}

// generateAuthAssertion generates the authentication assertion token.
func (a *authAssertExecutor) generateAuthAssertion(ctx *core.NodeContext, logger *log.Logger) (string, error) {
	tokenSub := ""
	if ctx.AuthenticatedUser.UserID != "" {
		tokenSub = ctx.AuthenticatedUser.UserID
	}

	jwtClaims := make(map[string]interface{})
	jwtConfig := config.GetThunderRuntime().Config.JWT
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

	authenticatorRefs := a.extractAuthenticatorReferences(ctx.ExecutionHistory)

	// Generate assertion from engaged authenticators
	if len(authenticatorRefs) > 0 {
		assertionResult, svcErr := a.authAssertGenerator.GenerateAssertion(authenticatorRefs)
		if svcErr != nil {
			if svcErr.Type == serviceerror.ServerErrorType {
				logger.Error("Failed to generate auth assertion",
					log.String("error", svcErr.Error))
				return "", errors.New("something went wrong while generating auth assertion")
			}
			return "", errors.New("failed to generate auth assertion: " + svcErr.Error)
		}

		jwtClaims["assurance"] = assertionResult.Context
	}

	// Include authorized permissions in JWT if present in runtime data
	// The "authorized_permissions" claim contains space-separated permission strings.
	// This claim will be present only if the authorization executor has run before this executor in the flow
	// and has set the authorized permissions in the runtime data.
	if permissions, exists := ctx.RuntimeData["authorized_permissions"]; exists && permissions != "" {
		jwtClaims["authorized_permissions"] = permissions
	}

	if err := a.appendUserDetailsToClaims(ctx, jwtClaims); err != nil {
		return "", err
	}

	// Add user type to the claims
	if ctx.AuthenticatedUser.UserType != "" {
		jwtClaims[userTypeKey] = ctx.AuthenticatedUser.UserType
	}

	if ctx.AuthenticatedUser.OrganizationUnitID != "" {
		if err := a.appendOUDetailsToClaims(ctx.AuthenticatedUser.OrganizationUnitID, jwtClaims); err != nil {
			return "", err
		}
	}

	token, _, err := a.jwtService.GenerateJWT(tokenSub, ctx.AppID, iss, validityPeriod, jwtClaims)
	if err != nil {
		logger.Error("Failed to generate JWT token", log.String("error", err.Error))
		return "", errors.New("failed to generate JWT token: " + err.Error)
	}

	return token, nil
}

// extractAuthenticatorReferences extracts authenticator references from execution history.
func (a *authAssertExecutor) extractAuthenticatorReferences(
	history map[string]*common.NodeExecutionRecord) []authncm.AuthenticatorReference {
	refs := make([]authncm.AuthenticatorReference, 0)
	seenAuthenticators := make(map[string]bool)

	for _, record := range history {
		if record.ExecutorType != common.ExecutorTypeAuthentication {
			continue
		}
		if record.Status != common.FlowStatusComplete {
			continue
		}

		// Map executor name to the authn service name
		authnServiceName := getAuthnServiceName(record.ExecutorName)
		if authnServiceName == "" {
			continue
		}

		// Skip if we've already seen this authenticator
		if seenAuthenticators[authnServiceName] {
			continue
		}
		seenAuthenticators[authnServiceName] = true

		refs = append(refs, authncm.AuthenticatorReference{
			Authenticator: authnServiceName,
			Step:          record.Step,
			Timestamp:     record.EndTime,
		})
	}

	// Sort by step field
	sort.Slice(refs, func(i, j int) bool {
		return refs[i].Step < refs[j].Step
	})

	// Renumber Step field to be auth step
	for i := range refs {
		refs[i].Step = i + 1
	}

	return refs
}

// appendUserDetailsToClaims appends user details to the JWT claims.
func (a *authAssertExecutor) appendUserDetailsToClaims(ctx *core.NodeContext,
	jwtClaims map[string]interface{}) error {
	if ctx.Application.Token != nil && len(ctx.Application.Token.UserAttributes) > 0 &&
		ctx.AuthenticatedUser.UserID != "" {
		var user *user.User
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
				user, attrs, err = a.getUserAttributes(ctx.AuthenticatedUser.UserID)
				if err != nil {
					return err
				}
			}

			// check for the attribute in user store attributes
			if val, ok := attrs[attr]; ok {
				jwtClaims[attr] = val
			}
		}
	}

	return nil
}

// getUserAttributes retrieves user details and unmarshal the attributes.
func (a *authAssertExecutor) getUserAttributes(userID string) (*user.User, map[string]interface{}, error) {
	logger := a.logger.With(log.String("userID", userID))

	var svcErr *serviceerror.ServiceError
	user, svcErr := a.userService.GetUser(context.TODO(), userID)
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

// appendOUDetailsToClaims appends organization unit details to the JWT claims.
func (a *authAssertExecutor) appendOUDetailsToClaims(ouID string, jwtClaims map[string]interface{}) error {
	logger := a.logger.With(log.String(ouIDKey, ouID))

	organizationUnit, svcErr := a.ouService.GetOrganizationUnit(ouID)
	if svcErr != nil {
		logger.Error("Failed to fetch organization unit details",
			log.String(ouIDKey, ouID), log.Any("error", svcErr))
		return errors.New("something went wrong while fetching organization unit: " + svcErr.ErrorDescription)
	}

	jwtClaims[ouIDKey] = organizationUnit.ID
	if organizationUnit.Name != "" {
		jwtClaims[userInputOuName] = organizationUnit.Name
	}
	if organizationUnit.Handle != "" {
		jwtClaims[userInputOuHandle] = organizationUnit.Handle
	}

	return nil
}
