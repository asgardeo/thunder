/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
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
	"errors"
	"fmt"
	"strconv"

	authncm "github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/authn/magiclink"
	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/jose/jwt"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
	"github.com/asgardeo/thunder/internal/userprovider"
)

// emailInput is the input definition for email collection.
var emailInput = common.Input{
	Ref:        "email_input",
	Identifier: userAttributeEmail,
	Type:       common.InputTypeText,
	Required:   true,
}

// Runtime data keys for MagicLink token tracking
const (
	runtimeKeyMagicLinkJti = "magicLinkJti"
	invalidMagicLinkToken  = "Invalid magic link token"
)

// magicLinkAuthExecutor implements the ExecutorInterface for Magic Link authentication.
type magicLinkAuthExecutor struct {
	core.ExecutorInterface
	identifyingExecutorInterface
	userProvider     userprovider.UserProviderInterface
	magicLinkService magiclink.MagicLinkAuthnServiceInterface
	logger           *log.Logger
}

var _ core.ExecutorInterface = (*magicLinkAuthExecutor)(nil)
var _ identifyingExecutorInterface = (*magicLinkAuthExecutor)(nil)

// newMagicLinkExecutorResponse creates a new instance of ExecutorResponse for Magic Link authentication.
func newMagicLinkExecutorResponse() *common.ExecutorResponse {
	return &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}
}

// newMagicLinkAuthExecutor creates a new instance of MagicLinkAuthExecutor.
func newMagicLinkAuthExecutor(
	flowFactory core.FlowFactoryInterface,
	magicLinkService magiclink.MagicLinkAuthnServiceInterface,
	userProvider userprovider.UserProviderInterface,
) *magicLinkAuthExecutor {
	defaultInputs := []common.Input{{
		Ref:        "magic_link_token_input",
		Identifier: userInputMagicLinkToken,
		Type:       common.InputTypeHidden,
		Required:   true,
	}}
	prerequisites := []common.Input{
		emailInput,
	}

	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "MagicLinkAuthExecutor"),
		log.String(log.LoggerKeyExecutorName, ExecutorNameMagicLinkAuth))

	identifyExec := newIdentifyingExecutor(ExecutorNameMagicLinkAuth, defaultInputs, prerequisites,
		flowFactory, userProvider)
	base := flowFactory.CreateExecutor(ExecutorNameMagicLinkAuth, common.ExecutorTypeAuthentication,
		defaultInputs, prerequisites)

	return &magicLinkAuthExecutor{
		ExecutorInterface:            base,
		identifyingExecutorInterface: identifyExec,
		userProvider:                 userProvider,
		magicLinkService:             magicLinkService,
		logger:                       logger,
	}
}

// Execute executes the Magic Link authentication logic.
func (m *magicLinkAuthExecutor) Execute(ctx *core.NodeContext) (*common.ExecutorResponse, error) {
	logger := m.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Executing Magic Link authentication executor")

	execResp := newMagicLinkExecutorResponse()

	if !m.ValidatePrerequisites(ctx, execResp) {
		logger.Debug("Prerequisites not met for Magic Link authentication executor")
		return execResp, nil
	}

	switch ctx.ExecutorMode {
	case ExecutorModeSend:
		return m.executeSend(ctx)
	case ExecutorModeVerify:
		return m.executeVerify(ctx)
	default:
		return execResp, fmt.Errorf("invalid executor mode: %s", ctx.ExecutorMode)
	}
}

// executeSend executes the magic link sending step.
func (m *magicLinkAuthExecutor) executeSend(ctx *core.NodeContext) (*common.ExecutorResponse, error) {
	logger := m.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))

	execResp, err := m.InitiateMagicLink(ctx)
	if err != nil {
		return execResp, err
	}

	logger.Debug("Magic link send completed", log.String("status", string(execResp.Status)))

	return execResp, nil
}

// executeVerify executes the magic link verification step.
func (m *magicLinkAuthExecutor) executeVerify(ctx *core.NodeContext) (*common.ExecutorResponse, error) {
	logger := m.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))

	execResp := newMagicLinkExecutorResponse()
	if !m.HasRequiredInputs(ctx, execResp) {
		logger.Debug("Required inputs for Magic Link verification are not provided")
		execResp.Status = common.ExecUserInputRequired
		return execResp, nil
	}

	execResp, err := m.ProcessAuthFlowResponse(ctx)
	if err != nil {
		return execResp, err
	}

	logger.Debug("Magic link verify completed",
		log.String("status", string(execResp.Status)),
		log.Bool("isAuthenticated", execResp.AuthenticatedUser.IsAuthenticated))

	return execResp, nil
}

// InitiateMagicLink initiates the magic link sending process to the user's email address.
func (m *magicLinkAuthExecutor) InitiateMagicLink(ctx *core.NodeContext) (*common.ExecutorResponse, error) {
	logger := m.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Sending magic link to user")

	execResp := newMagicLinkExecutorResponse()

	email, emailAttr, err := m.getUserEmailFromContext(ctx)
	if err != nil {
		return execResp, err
	}

	if ctx.FlowType == common.FlowTypeRegistration {
		filter := map[string]interface{}{emailAttr: email}
		userID, identifyErr := m.IdentifyUser(filter, execResp)
		if identifyErr != nil {
			return execResp, fmt.Errorf("failed to identify user during registration flow: %w", identifyErr)
		}

		if execResp.Status == common.ExecFailure && execResp.FailureReason != failureReasonUserNotFound {
			return execResp, errors.New("failed to identify user during registration flow")
		}

		if userID != nil && *userID != "" {
			execResp.Status = common.ExecFailure
			execResp.FailureReason = "User already exists with the provided email."
			return execResp, nil
		}

		execResp.Status = ""
		execResp.FailureReason = ""
	} else if ctx.AuthenticatedUser.IsAuthenticated {
		userIDVal := m.GetUserIDFromContext(ctx)
		if userIDVal == "" {
			return execResp, errors.New("user ID is empty in the context")
		}
		execResp.RuntimeData[userAttributeUserID] = userIDVal
	}

	failure, err := m.sendMagicLink(emailAttr, email, ctx, logger)
	if err != nil {
		return execResp, err
	}
	if failure != "" {
		execResp.Status = common.ExecFailure
		execResp.FailureReason = failure
		return execResp, nil
	}

	logger.Debug("Magic link sent successfully")
	execResp.Status = common.ExecComplete

	return execResp, nil
}

// ProcessAuthFlowResponse processes the authentication flow response for Magic Link.
func (m *magicLinkAuthExecutor) ProcessAuthFlowResponse(ctx *core.NodeContext) (*common.ExecutorResponse, error) {
	logger := m.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Processing authentication flow response for Magic Link")

	execResp := newMagicLinkExecutorResponse()

	userID, tokenJTI, failure, err := m.validateMagicLinkToken(ctx, logger)
	if err != nil {
		return execResp, err
	}
	if failure != "" {
		execResp.Status = common.ExecFailure
		execResp.FailureReason = failure
		return execResp, nil
	}

	authenticatedUser, err := m.getAuthenticatedUser(userID)
	if err != nil {
		return execResp, fmt.Errorf("failed to get authenticated user details: %w", err)
	}

	execResp.RuntimeData[runtimeKeyMagicLinkJti] = tokenJTI
	execResp.AuthenticatedUser = *authenticatedUser
	execResp.Status = common.ExecComplete

	logger.Debug("User authenticated successfully with Magic Link")

	return execResp, nil
}

// sendMagicLink sends a magic link to the user's email.
func (m *magicLinkAuthExecutor) sendMagicLink(emailAttr string, email string, ctx *core.NodeContext,
	logger *log.Logger) (string, error) {
	logger.Debug("Sending magic link to user", log.String("email", log.MaskString(email)))

	expirySeconds := m.getTokenExpiry(ctx)
	magicLinkURL := m.getMagicLinkURL(ctx)
	queryParams := map[string]string{
		"flowId": ctx.FlowID,
	}

	svcErr := m.magicLinkService.SendMagicLink(ctx.Context, emailAttr, email, expirySeconds, queryParams, magicLinkURL)
	if svcErr != nil {
		if svcErr.Type == serviceerror.ClientErrorType {
			return svcErr.ErrorDescription.DefaultValue, nil
		}
		return "", errors.New("failed to send magic link")
	}

	logger.Debug("Magic link sent successfully")

	return "", nil
}

// getTokenExpiry returns the magic link token expiry in seconds from node properties,
// falling back to the default if not configured or invalid.
func (m *magicLinkAuthExecutor) getTokenExpiry(ctx *core.NodeContext) *int64 {
	if ctx.NodeProperties != nil {
		if val, ok := ctx.NodeProperties[propertyKeyTokenExpiry]; ok {
			if str, valid := val.(string); valid && str != "" {
				if parsed, err := strconv.ParseInt(str, 10, 64); err == nil && parsed > 0 {
					return &parsed
				}
			}
		}
	}
	return nil
}

// getMagicLinkURL returns the magic link URL prefix from node properties,
// returning nil if not configured.
func (m *magicLinkAuthExecutor) getMagicLinkURL(ctx *core.NodeContext) *string {
	if ctx.NodeProperties != nil {
		if val, ok := ctx.NodeProperties[propertyKeyMagicLinkURL]; ok {
			if str, valid := val.(string); valid && str != "" {
				return &str
			}
		}
	}
	return nil
}

// validateMagicLinkToken validates the magic link token provided by the user.
func (m *magicLinkAuthExecutor) validateMagicLinkToken(ctx *core.NodeContext,
	logger *log.Logger) (string, string, string, error) {
	token, ok := ctx.UserInputs[userInputMagicLinkToken]
	if !ok || token == "" {
		logger.Debug("Magic link token not found in user inputs")
		return "", "", "Magic link token is required", nil
	}

	payload, decodeErr := jwt.DecodeJWTPayload(token)
	if decodeErr != nil {
		logger.Debug("Failed to decode magic link token", log.Error(decodeErr))
		return "", "", invalidMagicLinkToken, nil
	}

	flowIDClaim := utils.ConvertInterfaceValueToString(payload["flowId"])
	if flowIDClaim == "" {
		logger.Debug("Magic link token missing flowId claim")
		return "", "", invalidMagicLinkToken, nil
	}
	if flowIDClaim != ctx.FlowID {
		logger.Debug("Magic link token flowId mismatch",
			log.String("expected", ctx.FlowID),
			log.String("actual", flowIDClaim))
		return "", "", invalidMagicLinkToken, nil
	}

	jtiClaim := utils.ConvertInterfaceValueToString(payload["jti"])
	if jtiClaim == "" {
		logger.Debug("Magic link token missing jti claim")
		return "", "", invalidMagicLinkToken, nil
	}

	if usedJti, exists := ctx.RuntimeData[runtimeKeyMagicLinkJti]; exists && usedJti != "" {
		if usedJti == jtiClaim {
			logger.Debug("Magic link token has already been used",
				log.String("jti", jtiClaim))
			return "", "", "Magic link has already been used", nil
		}
		logger.Debug("New magic link token detected, replacing previous token")
	}

	user, svcErr := m.magicLinkService.VerifyMagicLink(ctx.Context, token)
	if svcErr != nil {
		if svcErr.Type == serviceerror.ClientErrorType {
			return "", "", svcErr.ErrorDescription.DefaultValue, nil
		}
		return "", "", "", errors.New("failed to verify magic link token")
	}

	if user == nil {
		return "", "", failureReasonUserNotFound, nil
	}

	logger.Debug("Magic link token validated successfully", log.String("userID", user.UserID))
	return user.UserID, jtiClaim, "", nil
}

// getUserEmailFromContext retrieves the user's email from the context.
func (m *magicLinkAuthExecutor) getUserEmailFromContext(ctx *core.NodeContext) (string, string, error) {
	emailAttr := userAttributeEmail
	for _, input := range m.GetPrerequisites() {
		if input.Ref == emailInput.Ref {
			emailAttr = input.Identifier
			break
		}
	}
	if email, ok := ctx.UserInputs[emailAttr]; ok && email != "" {
		return email, emailAttr, nil
	}
	if email, ok := ctx.RuntimeData[emailAttr]; ok && email != "" {
		return email, emailAttr, nil
	}
	return "", "", errors.New("email not found in user inputs or runtime data")
}

// getAuthenticatedUser retrieves the authenticated user details from the user provider.
func (m *magicLinkAuthExecutor) getAuthenticatedUser(
	userID string) (*authncm.AuthenticatedUser, error) {
	if userID == "" {
		return nil, errors.New("user ID is empty")
	}

	user, err := m.userProvider.GetUser(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &authncm.AuthenticatedUser{
		IsAuthenticated: true,
		UserID:          user.UserID,
		UserType:        user.UserType,
		OUID:            user.OUID,
	}, nil
}
