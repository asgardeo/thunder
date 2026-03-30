/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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
	"errors"
	"fmt"
	"strconv"

	authncm "github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/authn/magiclink"
	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/observability"
	"github.com/asgardeo/thunder/internal/userprovider"
)

// EmailInput is the input definition for email collection.
var EmailInput = common.Input{
	Ref:        "email_input",
	Identifier: userAttributeEmail,
	Type:       common.InputTypeText,
	Required:   true,
}

// MagicLinkTokenInput is the input definition for magic link token verification.
var MagicLinkTokenInput = common.Input{
	Ref:        "magic_link_token_input",
	Identifier: userInputMagicLinkToken,
	Type:       common.InputTypeText,
	Required:   true,
}

// magicLinkAuthExecutor implements the ExecutorInterface for Magic Link authentication.
type magicLinkAuthExecutor struct {
	core.ExecutorInterface
	identifyingExecutorInterface
	userProvider     userprovider.UserProviderInterface
	magicLinkService magiclink.MagicLinkAuthnServiceInterface
	observabilitySvc observability.ObservabilityServiceInterface
	logger           *log.Logger
}

var _ core.ExecutorInterface = (*magicLinkAuthExecutor)(nil)
var _ identifyingExecutorInterface = (*magicLinkAuthExecutor)(nil)

// newMagicLinkAuthExecutor creates a new instance of MagicLinkAuthExecutor.
func newMagicLinkAuthExecutor(
	flowFactory core.FlowFactoryInterface,
	magicLinkService magiclink.MagicLinkAuthnServiceInterface,
	observabilitySvc observability.ObservabilityServiceInterface,
	userProvider userprovider.UserProviderInterface,
) *magicLinkAuthExecutor {
	defaultInputs := []common.Input{
		MagicLinkTokenInput,
	}
	prerequisites := []common.Input{
		EmailInput,
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
		observabilitySvc:             observabilitySvc,
		logger:                       logger,
	}
}

// Execute executes the Magic Link authentication logic.
func (m *magicLinkAuthExecutor) Execute(ctx *core.NodeContext) (*common.ExecutorResponse, error) {
	logger := m.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Executing Magic Link authentication executor")

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	if !m.ValidatePrerequisites(ctx, execResp) {
		logger.Debug("Prerequisites not met for Magic Link authentication executor")
		return execResp, nil
	}

	switch ctx.ExecutorMode {
	case ExecutorModeSend:
		return m.executeSend(ctx, execResp)
	case ExecutorModeVerify:
		return m.executeVerify(ctx, execResp)
	default:
		return execResp, fmt.Errorf("invalid executor mode: %s", ctx.ExecutorMode)
	}
}

// executeSend executes the magic link sending step.
func (m *magicLinkAuthExecutor) executeSend(ctx *core.NodeContext,
	execResp *common.ExecutorResponse) (*common.ExecutorResponse, error) {
	logger := m.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))

	err := m.InitiateMagicLink(ctx, execResp)
	if err != nil {
		return execResp, err
	}

	logger.Debug("Magic link send completed", log.String("status", string(execResp.Status)))

	return execResp, nil
}

// executeVerify executes the magic link verification step.
func (m *magicLinkAuthExecutor) executeVerify(ctx *core.NodeContext,
	execResp *common.ExecutorResponse) (*common.ExecutorResponse, error) {
	logger := m.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))

	if !m.HasRequiredInputs(ctx, execResp) {
		logger.Debug("Required inputs for Magic Link verification are not provided")
		execResp.Status = common.ExecUserInputRequired
		return execResp, nil
	}

	err := m.ProcessAuthFlowResponse(ctx, execResp)
	if err != nil {
		return execResp, err
	}

	logger.Debug("Magic link verify completed",
		log.String("status", string(execResp.Status)),
		log.Bool("isAuthenticated", execResp.AuthenticatedUser.IsAuthenticated))

	return execResp, nil
}

// InitiateMagicLink initiates the magic link sending process to the user's email address.
func (m *magicLinkAuthExecutor) InitiateMagicLink(ctx *core.NodeContext,
	execResp *common.ExecutorResponse) error {
	logger := m.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Sending magic link to user")

	if m.magicLinkService == nil {
		logger.Error("Magic link service is not configured")
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "Authentication method is currently unavailable"
		return nil
	}

	email, err := m.getUserEmailFromContext(ctx)
	if err != nil {
		return err
	}

	if ctx.FlowType == common.FlowTypeRegistration {
		filter := map[string]interface{}{userAttributeEmail: email}
		userID, identifyErr := m.IdentifyUser(filter, execResp)
		if identifyErr != nil {
			logger.Error("Failed to identify user", log.Error(identifyErr))
			return fmt.Errorf("failed to identify user during registration flow: %w", identifyErr)
		}

		if execResp.Status == common.ExecFailure && execResp.FailureReason != failureReasonUserNotFound {
			return errors.New("failed to identify user during registration flow")
		}

		if userID != nil && *userID != "" {
			execResp.Status = common.ExecFailure
			execResp.FailureReason = "User already exists with the provided email."
			return nil
		}

		execResp.Status = ""
		execResp.FailureReason = ""
	} else if ctx.AuthenticatedUser.IsAuthenticated {
		userIDVal := m.GetUserIDFromContext(ctx)
		if userIDVal == "" {
			return errors.New("user ID is empty in the context")
		}
		execResp.RuntimeData[userAttributeUserID] = userIDVal
	}

	m.sendMagicLink(email, ctx, execResp, logger)
	if execResp.Status == common.ExecFailure {
		return nil
	}

	logger.Debug("Magic link sent successfully")
	execResp.Status = common.ExecComplete

	return nil
}

// ProcessAuthFlowResponse processes the authentication flow response for Magic Link.
func (m *magicLinkAuthExecutor) ProcessAuthFlowResponse(ctx *core.NodeContext,
	execResp *common.ExecutorResponse) error {
	logger := m.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Processing authentication flow response for Magic Link")

	m.validateMagicLinkToken(ctx, execResp, logger)
	if execResp.Status == common.ExecFailure {
		return nil
	}

	authenticatedUser, err := m.getAuthenticatedUser(execResp)
	if err != nil {
		logger.Error("Failed to get authenticated user details", log.Error(err))
		return fmt.Errorf("failed to get authenticated user details: %w", err)
	}

	execResp.AuthenticatedUser = *authenticatedUser
	execResp.Status = common.ExecComplete

	logger.Debug("User authenticated successfully with Magic Link")

	return nil
}

// sendMagicLink sends a magic link to the user's email.
func (m *magicLinkAuthExecutor) sendMagicLink(email string, ctx *core.NodeContext,
	execResp *common.ExecutorResponse, logger *log.Logger) {
	logger.Debug("Sending magic link to user", log.String("email", log.MaskString(email)))

	expirySeconds := m.getTokenExpiry(ctx)
	magicLinkURL := m.getMagicLinkURL(ctx)
	svcErr := m.magicLinkService.SendMagicLink(ctx.Context, email, expirySeconds, ctx.FlowID, magicLinkURL)
	if svcErr != nil {
		logger.Error("Failed to send magic link", log.String("error", svcErr.Error.String()))
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "Failed to send magic link"
		return
	}
}

// getTokenExpiry returns the magic link token expiry in seconds from node properties,
// falling back to the default if not configured or invalid.
func (m *magicLinkAuthExecutor) getTokenExpiry(ctx *core.NodeContext) int64 {
	if ctx.NodeProperties != nil {
		if val, ok := ctx.NodeProperties[propertyKeyTokenExpiry]; ok {
			if str, valid := val.(string); valid && str != "" {
				if parsed, err := strconv.ParseInt(str, 10, 64); err == nil && parsed > 0 {
					return parsed
				}
			}
		}
	}
	return int64(magiclink.DefaultExpirySeconds)
}

// getMagicLinkURL returns the magic link URL prefix from node properties,
// returning an empty string if not configured.
func (m *magicLinkAuthExecutor) getMagicLinkURL(ctx *core.NodeContext) string {
	if ctx.NodeProperties != nil {
		if val, ok := ctx.NodeProperties[propertyKeyMagicLinkURL]; ok {
			if str, valid := val.(string); valid {
				return str
			}
		}
	}
	return ""
}

// validateMagicLinkToken validates the magic link token provided by the user.
func (m *magicLinkAuthExecutor) validateMagicLinkToken(ctx *core.NodeContext,
	execResp *common.ExecutorResponse, logger *log.Logger) {
	token, ok := ctx.UserInputs[userInputMagicLinkToken]
	if !ok || token == "" {
		logger.Debug("Magic link token not found in user inputs")
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "Magic link token is required"
		return
	}

	logger.Debug("Verifying magic link token")
	user, svcErr := m.magicLinkService.VerifyMagicLink(ctx.Context, token)
	if svcErr != nil {
		logger.Error("Failed to verify magic link token", log.String("error", svcErr.Error.String()))
		execResp.Status = common.ExecFailure
		execResp.FailureReason = failureReasonInvalidMagicLinkToken
		return
	}

	if user == nil {
		logger.Error("User not found after magic link verification")
		execResp.Status = common.ExecFailure
		execResp.FailureReason = failureReasonUserNotFound
		return
	}

	execResp.RuntimeData[userAttributeUserID] = user.UserID
	logger.Debug("Magic link token validated successfully", log.String("userID", user.UserID))
}

// getUserEmailFromContext retrieves the user's email from the context.
func (m *magicLinkAuthExecutor) getUserEmailFromContext(ctx *core.NodeContext) (string, error) {
	if email, ok := ctx.UserInputs[userAttributeEmail]; ok && email != "" {
		return email, nil
	}
	if email, ok := ctx.RuntimeData[userAttributeEmail]; ok && email != "" {
		return email, nil
	}
	return "", errors.New("email not found in user inputs or runtime data")
}

// getAuthenticatedUser retrieves the authenticated user details from the user provider.
func (m *magicLinkAuthExecutor) getAuthenticatedUser(
	execResp *common.ExecutorResponse) (*authncm.AuthenticatedUser, error) {
	userID, ok := execResp.RuntimeData[userAttributeUserID]
	if !ok || userID == "" {
		return nil, errors.New("user ID not found in runtime data")
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
