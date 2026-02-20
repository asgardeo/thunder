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
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	authncm "github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/notification"
	notifcommon "github.com/asgardeo/thunder/internal/notification/common"
	"github.com/asgardeo/thunder/internal/observability"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/userprovider"
)

// MobileNumberInput is the input definition for mobile number collection.
var MobileNumberInput = common.Input{
	Ref:        "mobile_number_input",
	Identifier: userAttributeMobileNumber,
	Type:       common.InputTypePhone,
	Required:   true,
}

// smsOTPAuthExecutor implements the ExecutorInterface for SMS OTP authentication.
type smsOTPAuthExecutor struct {
	core.ExecutorInterface
	identifyingExecutorInterface
	userProvider     userprovider.UserProviderInterface
	otpService       notification.OTPServiceInterface
	observabilitySvc observability.ObservabilityServiceInterface
	logger           *log.Logger
}

var _ core.ExecutorInterface = (*smsOTPAuthExecutor)(nil)
var _ identifyingExecutorInterface = (*smsOTPAuthExecutor)(nil)

// newSMSOTPAuthExecutor creates a new instance of SMSOTPAuthExecutor.
func newSMSOTPAuthExecutor(
	flowFactory core.FlowFactoryInterface,
	otpService notification.OTPServiceInterface,
	observabilitySvc observability.ObservabilityServiceInterface,
	userProvider userprovider.UserProviderInterface,
) *smsOTPAuthExecutor {
	defaultInputs := []common.Input{
		{
			Ref:        "otp_input",
			Identifier: userInputOTP,
			Type:       common.InputTypeOTP,
			Required:   true,
		},
	}
	prerequisites := []common.Input{
		MobileNumberInput,
	}

	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "SMSOTPAuthExecutor"),
		log.String(log.LoggerKeyExecutorName, ExecutorNameSMSAuth))

	identifyExec := newIdentifyingExecutor(ExecutorNameSMSAuth, defaultInputs, prerequisites,
		flowFactory, userProvider)
	base := flowFactory.CreateExecutor(ExecutorNameSMSAuth, common.ExecutorTypeAuthentication,
		defaultInputs, prerequisites)

	return &smsOTPAuthExecutor{
		ExecutorInterface:            base,
		identifyingExecutorInterface: identifyExec,
		userProvider:                 userProvider,
		otpService:                   otpService,
		observabilitySvc:             observabilitySvc,
		logger:                       logger,
	}
}

// Execute executes the SMS OTP authentication logic.
func (s *smsOTPAuthExecutor) Execute(ctx *core.NodeContext) (*common.ExecutorResponse, error) {
	logger := s.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Executing SMS OTP authentication executor")

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	if !s.ValidatePrerequisites(ctx, execResp) {
		logger.Debug("Prerequisites not met for SMS OTP authentication executor")
		return execResp, nil
	}

	// Determine the executor mode
	switch ctx.ExecutorMode {
	case ExecutorModeSend:
		return s.executeSend(ctx, execResp)
	case ExecutorModeVerify:
		return s.executeVerify(ctx, execResp)
	default:
		return execResp, fmt.Errorf("invalid executor mode: %s", ctx.ExecutorMode)
	}
}

// executeSend executes the OTP sending step.
func (s *smsOTPAuthExecutor) executeSend(ctx *core.NodeContext,
	execResp *common.ExecutorResponse) (*common.ExecutorResponse, error) {
	logger := s.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))

	err := s.InitiateOTP(ctx, execResp)
	if err != nil {
		return execResp, err
	}

	logger.Debug("SMS OTP send completed", log.String("status", string(execResp.Status)))

	return execResp, nil
}

// executeVerify executes the OTP verification step.
func (s *smsOTPAuthExecutor) executeVerify(ctx *core.NodeContext,
	execResp *common.ExecutorResponse) (*common.ExecutorResponse, error) {
	logger := s.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))

	if !s.HasRequiredInputs(ctx, execResp) {
		logger.Debug("Required inputs for SMS OTP verification are not provided")
		execResp.Status = common.ExecUserInputRequired
		return execResp, nil
	}

	err := s.ProcessAuthFlowResponse(ctx, execResp)
	if err != nil {
		return execResp, err
	}

	logger.Debug("SMS OTP verify completed",
		log.String("status", string(execResp.Status)),
		log.Bool("isAuthenticated", execResp.AuthenticatedUser.IsAuthenticated))

	return execResp, nil
}

// InitiateOTP initiates the OTP sending process to the user's mobile number.
func (s *smsOTPAuthExecutor) InitiateOTP(ctx *core.NodeContext,
	execResp *common.ExecutorResponse) error {
	logger := s.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Sending SMS OTP to user")

	mobileNumber, err := s.getUserMobileFromContext(ctx)
	if err != nil {
		return err
	}

	var userID *string
	if ctx.AuthenticatedUser.IsAuthenticated {
		userIDVal := s.GetUserIDFromContext(ctx)
		if userIDVal == "" {
			return errors.New("user ID is empty in the context")
		}
		userID = &userIDVal
	} else {
		// Identify user by mobile number if not authenticated
		if mobileNumber == "" {
			logger.Error("Mobile number is empty in the context")
		}

		filter := map[string]interface{}{userAttributeMobileNumber: mobileNumber}
		userID, err = s.IdentifyUser(filter, execResp)
		if err != nil {
			logger.Error("Failed to identify user", log.Error(err))
			return fmt.Errorf("failed to identify user: %w", err)
		}
	}

	// Handle registration flows.
	if ctx.FlowType == common.FlowTypeRegistration {
		if execResp.Status == common.ExecFailure && execResp.FailureReason != failureReasonUserNotFound {
			logger.Error("Failed to identify user during registration flow", log.Error(err))
			return fmt.Errorf("failed to identify user during registration flow: %w", err)
		}

		if userID != nil && *userID != "" {
			// At this point, a unique user is found in the system. Hence fail the execution.
			execResp.Status = common.ExecFailure
			execResp.FailureReason = "User already exists with the provided mobile number."
			return nil
		}

		execResp.Status = ""
		execResp.FailureReason = ""
	} else {
		if execResp.Status == common.ExecFailure {
			return nil
		}
		execResp.RuntimeData[userAttributeUserID] = *userID
	}

	// Send the OTP to the user's mobile number.
	if err := s.generateAndSendOTP(mobileNumber, ctx, execResp, logger); err != nil {
		logger.Error("Failed to send OTP", log.Error(err))
		return fmt.Errorf("failed to send OTP: %w", err)
	}
	if execResp.Status == common.ExecFailure {
		return nil
	}

	logger.Debug("SMS OTP sent successfully")
	execResp.Status = common.ExecComplete

	return nil
}

// ProcessAuthFlowResponse processes the authentication flow response for SMS OTP.
func (s *smsOTPAuthExecutor) ProcessAuthFlowResponse(ctx *core.NodeContext,
	execResp *common.ExecutorResponse) error {
	logger := s.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Processing authentication flow response for SMS OTP")

	err := s.validateOTP(ctx, execResp, logger)
	if err != nil {
		logger.Error("Error validating OTP", log.Error(err))
		return fmt.Errorf("error validating OTP: %w", err)
	}
	if execResp.Status == common.ExecFailure {
		return nil
	}

	authenticatedUser, err := s.getAuthenticatedUser(ctx, execResp)
	if err != nil {
		logger.Error("Failed to get authenticated user details", log.Error(err))
		return fmt.Errorf("failed to get authenticated user details: %w", err)
	}

	execResp.AuthenticatedUser = *authenticatedUser
	execResp.Status = common.ExecComplete

	return nil
}

// ValidatePrerequisites validates whether the prerequisites for the SMSOTPAuthExecutor are met.
func (s *smsOTPAuthExecutor) ValidatePrerequisites(ctx *core.NodeContext,
	execResp *common.ExecutorResponse) bool {
	preReqMet := s.ExecutorInterface.ValidatePrerequisites(ctx, execResp)
	if preReqMet {
		return true
	}

	logger := s.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))

	if ctx.FlowType == common.FlowTypeRegistration {
		logger.Debug("Prerequisites not met for registration flow, prompting for mobile number")
		execResp.Status = common.ExecUserInputRequired
		execResp.Inputs = []common.Input{MobileNumberInput}
		return false
	}

	logger.Debug("Trying to satisfy prerequisites for SMS OTP authentication executor")

	s.satisfyPrerequisites(ctx, execResp)
	if execResp.Status == common.ExecFailure {
		return false
	}

	return s.ExecutorInterface.ValidatePrerequisites(ctx, execResp)
}

// getUserMobileFromContext retrieves the user's mobile number from the context.
func (s *smsOTPAuthExecutor) getUserMobileFromContext(ctx *core.NodeContext) (string, error) {
	mobileNumber := ctx.RuntimeData[userAttributeMobileNumber]

	if mobileNumber == "" {
		mobileNumber = ctx.UserInputs[userAttributeMobileNumber]
	}

	if mobileNumber == "" && ctx.AuthenticatedUser.Attributes != nil {
		if mobile, ok := ctx.AuthenticatedUser.Attributes[userAttributeMobileNumber]; ok {
			if mobileStr, valid := mobile.(string); valid && mobileStr != "" {
				mobileNumber = mobileStr
			}
		}
	}

	if mobileNumber == "" {
		return "", errors.New("mobile number not found in context")
	}
	return mobileNumber, nil
}

// satisfyPrerequisites tries to satisfy the prerequisites for the SMSOTPAuthExecutor.
func (s *smsOTPAuthExecutor) satisfyPrerequisites(ctx *core.NodeContext,
	execResp *common.ExecutorResponse) {
	logger := s.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))

	execResp.Status = ""
	execResp.FailureReason = ""

	logger.Debug("Trying to resolve user ID from context data")
	userIDResolved, err := s.resolveUserID(ctx)
	if err != nil {
		logger.Error("Failed to resolve user ID from context data", log.Error(err))
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "Failed to resolve user ID from context data"
		return
	}
	if !userIDResolved {
		logger.Debug("User ID could not be resolved from context data")
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "User ID could not be resolved from context data"
		return
	}
	userID := ctx.RuntimeData[userAttributeUserID]

	// TODO: If the mobile number is not found, but the user is authenticated, this method will
	//  prompt the user to enter their mobile number.
	//  We should verify whether this is the expected behavior.

	logger.Debug("Retrieving mobile number from user ID", log.String("userID", userID))
	mobileNumber, err := s.getUserMobileNumber(userID, ctx, execResp)
	if err != nil {
		logger.Error("Failed to retrieve mobile number", log.String("userID", userID), log.Error(err))
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "Failed to retrieve mobile number"
		return
	}
	if execResp.Status == common.ExecFailure {
		return
	}

	logger.Debug("Mobile number retrieved successfully", log.String("userID", userID))
	ctx.RuntimeData[userAttributeMobileNumber] = mobileNumber

	// Reset the executor response status and failure reason.
	execResp.Status = ""
	execResp.FailureReason = ""
}

// resolveUserID resolves the user ID from the context based on various attributes.
// TODO: Move to a separate resolver when the support is added.
func (s *smsOTPAuthExecutor) resolveUserID(ctx *core.NodeContext) (bool, error) {
	logger := s.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))

	// First, check if the user ID is already available in the context.
	userID := s.GetUserIDFromContext(ctx)
	if userID != "" {
		logger.Debug("User ID found in context data", log.String("userID", userID))
		if ctx.RuntimeData == nil {
			ctx.RuntimeData = make(map[string]string)
		}
		ctx.RuntimeData[userAttributeUserID] = userID

		return true, nil
	}

	userIDResolved := false

	// Try to resolve user ID from mobile number next.
	userIDResolved, err := s.resolveUserIDFromAttribute(ctx, userAttributeMobileNumber, logger)
	if err != nil {
		return false, err
	}
	if userIDResolved {
		return true, nil
	}

	// Try to resolve user ID from username first.
	userIDResolved, err = s.resolveUserIDFromAttribute(ctx, userAttributeUsername, logger)
	if err != nil {
		return false, err
	}
	if userIDResolved {
		return true, nil
	}

	// Try to resolve user ID from email next.
	userIDResolved, err = s.resolveUserIDFromAttribute(ctx, userAttributeEmail, logger)
	if err != nil {
		return false, err
	}
	if userIDResolved {
		return true, nil
	}

	return false, nil
}

// resolveUserIDFromAttribute attempts to resolve the user ID from a specific attribute in the context.
func (s *smsOTPAuthExecutor) resolveUserIDFromAttribute(ctx *core.NodeContext,
	attributeName string, logger *log.Logger) (bool, error) {
	logger.Debug("Resolving user ID from attribute", log.String("attributeName", attributeName))

	attributeValue := ctx.UserInputs[attributeName]
	if attributeValue == "" {
		attributeValue = ctx.RuntimeData[attributeName]
	}
	if attributeValue != "" {
		filters := map[string]interface{}{attributeName: attributeValue}
		userID, providerErr := s.userProvider.IdentifyUser(filters)
		if providerErr != nil {
			return false, fmt.Errorf("failed to identify user by %s: %s", attributeName, providerErr.Error())
		}
		if userID != nil && *userID != "" {
			logger.Debug("User ID resolved from attribute", log.String("attributeName", attributeName),
				log.String("userID", *userID))
			if ctx.RuntimeData == nil {
				ctx.RuntimeData = make(map[string]string)
			}
			ctx.RuntimeData[userAttributeUserID] = *userID
			return true, nil
		}
	}

	return false, nil
}

// getUserMobileNumber retrieves the mobile number for the given user ID.
func (s *smsOTPAuthExecutor) getUserMobileNumber(userID string, ctx *core.NodeContext,
	execResp *common.ExecutorResponse) (string, error) {
	logger := s.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID), log.String("userID", userID))
	logger.Debug("Retrieving user mobile number")

	// Try to get mobile number from context
	mobileNumber, err := s.getUserMobileFromContext(ctx)
	if err == nil && mobileNumber != "" {
		logger.Debug("Mobile number found in context, skipping user store call")
		return mobileNumber, nil
	}

	// Mobile number not in context, fetch from user store
	logger.Debug("Mobile number not in context, fetching from user store")
	user, providerErr := s.userProvider.GetUser(userID)
	if providerErr != nil {
		return "", fmt.Errorf("failed to retrieve user details: %s", providerErr.Error())
	}

	// Extract mobile number from user attributes
	var attrs map[string]interface{}
	if err := json.Unmarshal(user.Attributes, &attrs); err != nil {
		return "", fmt.Errorf("failed to unmarshal user attributes: %w", err)
	}

	mobileNumber = ""
	mobileNumberAttr := attrs[userAttributeMobileNumber]
	if mobileNumberAttr != nil && mobileNumberAttr != "" {
		mobileNumber = mobileNumberAttr.(string)
	}

	if mobileNumber == "" {
		logger.Debug("Mobile number not found in user attributes or context")
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "Mobile number not found in user attributes or context"
		return "", nil
	}

	return mobileNumber, nil
}

// generateAndSendOTP generates an OTP and sends it to the user's mobile number.
func (s *smsOTPAuthExecutor) generateAndSendOTP(mobileNumber string, ctx *core.NodeContext,
	execResp *common.ExecutorResponse, logger *log.Logger) error {
	attemptCount, err := s.validateAttempts(ctx, execResp, logger)
	if err != nil {
		return fmt.Errorf("failed to validate OTP attempts: %w", err)
	}
	if execResp.Status == common.ExecFailure {
		return nil
	}

	// Get the message sender id from node properties.
	if len(ctx.NodeProperties) == 0 {
		return errors.New("message sender id is not configured in node properties")
	}

	senderID := ""
	if senderIDVal, ok := ctx.NodeProperties["senderId"]; ok {
		if sid, valid := senderIDVal.(string); valid && sid != "" {
			senderID = sid
		}
	}
	if senderID == "" {
		return errors.New("senderId is not configured in node properties")
	}

	// Send the OTP
	sendOTPRequest := notifcommon.SendOTPDTO{
		Recipient: mobileNumber,
		SenderID:  senderID,
		Channel:   string(notifcommon.ChannelTypeSMS),
	}

	sendResult, svcErr := s.otpService.SendOTP(sendOTPRequest)
	if svcErr != nil {
		return fmt.Errorf("failed to send OTP: %s", svcErr.ErrorDescription)
	}

	// Store runtime data
	if execResp.RuntimeData == nil {
		execResp.RuntimeData = make(map[string]string)
	}
	execResp.RuntimeData["otpSessionToken"] = sendResult.SessionToken
	execResp.RuntimeData["attemptCount"] = strconv.Itoa(attemptCount + 1)

	return nil
}

// validateAttempts checks if the maximum number of OTP attempts has been reached.
func (s *smsOTPAuthExecutor) validateAttempts(ctx *core.NodeContext, execResp *common.ExecutorResponse,
	logger *log.Logger) (int, error) {
	userID := ctx.RuntimeData[userAttributeUserID]
	attemptCount := 0

	attemptCountStr := ctx.RuntimeData["attemptCount"]
	if attemptCountStr != "" {
		count, err := strconv.Atoi(attemptCountStr)
		if err != nil {
			logger.Error("Failed to parse attempt count", log.Error(err))
			return 0, fmt.Errorf("failed to parse attempt count: %w", err)
		}
		attemptCount = count
	}

	if attemptCount >= s.getOTPMaxAttempts() {
		logger.Debug("Maximum OTP attempts reached", log.String("userID", userID),
			log.Int("attemptCount", attemptCount))
		execResp.Status = common.ExecFailure
		execResp.FailureReason = fmt.Sprintf("maximum OTP attempts reached: %d", attemptCount)
		return 0, nil
	}

	return attemptCount, nil
}

// getOTPMaxAttempts returns the maximum number of attempts allowed for OTP validation.
func (s *smsOTPAuthExecutor) getOTPMaxAttempts() int {
	// TODO: This needs to be configured as a IDP property.
	return 3
}

// validateOTP validates the OTP for the given user and mobile number.
func (s *smsOTPAuthExecutor) validateOTP(ctx *core.NodeContext, execResp *common.ExecutorResponse,
	logger *log.Logger) error {
	userID := ctx.RuntimeData[userAttributeUserID]
	providedOTP := ctx.UserInputs[userInputOTP]

	logger.Debug("Validating OTP", log.String("userID", userID))

	if providedOTP == "" {
		logger.Debug("Provided OTP is empty", log.String("userID", userID))
		execResp.Status = common.ExecFailure
		execResp.FailureReason = failureReasonInvalidOTP
		return nil
	}

	sessionToken := ctx.RuntimeData["otpSessionToken"]
	if sessionToken == "" {
		logger.Error("No session token found for OTP validation", log.String("userID", userID))
		return fmt.Errorf("no session token found for OTP validation")
	}

	// Use the OTP service to verify the OTP
	verifyOTPRequest := notifcommon.VerifyOTPDTO{
		SessionToken: sessionToken,
		OTPCode:      providedOTP,
	}

	verifyResult, svcErr := s.otpService.VerifyOTP(verifyOTPRequest)
	if svcErr != nil {
		logger.Error("Failed to verify OTP", log.String("userID", userID), log.Any("serviceError", svcErr))
		return fmt.Errorf("failed to verify OTP: %s", svcErr.ErrorDescription)
	}

	// Check verification result
	if verifyResult.Status != notifcommon.OTPVerifyStatusVerified {
		logger.Debug("OTP verification failed", log.String("userID", userID),
			log.String("status", string(verifyResult.Status)))

		execResp.Status = common.ExecFailure
		execResp.FailureReason = failureReasonInvalidOTP
		return nil
	}

	execResp.RuntimeData["otpSessionToken"] = ""
	logger.Debug("OTP validated successfully", log.String("userID", userID))
	return nil
}

// getAuthenticatedUser returns the authenticated user details for the given user ID.
func (s *smsOTPAuthExecutor) getAuthenticatedUser(ctx *core.NodeContext,
	execResp *common.ExecutorResponse) (*authncm.AuthenticatedUser, error) {
	logger := s.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))

	mobileNumber, err := s.getUserMobileFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Handle registration flows.
	if ctx.FlowType == common.FlowTypeRegistration {
		execResp.Status = common.ExecComplete
		execResp.FailureReason = ""
		return &authncm.AuthenticatedUser{
			IsAuthenticated: false,
			Attributes: map[string]interface{}{
				userAttributeMobileNumber: mobileNumber,
			},
		}, nil
	}

	// Check if user is already authenticated
	if ctx.AuthenticatedUser.IsAuthenticated && ctx.AuthenticatedUser.UserID != "" {
		if ctx.AuthenticatedUser.Attributes == nil {
			ctx.AuthenticatedUser.Attributes = make(map[string]interface{})
		}
		ctx.AuthenticatedUser.Attributes[userAttributeMobileNumber] = mobileNumber
		return &ctx.AuthenticatedUser, nil
	}

	userID := ctx.RuntimeData[userAttributeUserID]
	if userID == "" {
		return nil, errors.New("user ID is empty")
	}

	// User not available in context, fetch from user store
	logger.Debug("Fetching user details from user store", log.String("userID", userID))
	user, providerErr := s.userProvider.GetUser(userID)
	if providerErr != nil {
		return nil, fmt.Errorf("failed to get user details: %s", providerErr.Error())
	}

	// Extract user attributes
	var attrs map[string]interface{}
	if err := json.Unmarshal(user.Attributes, &attrs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user attributes: %w", err)
	}

	// Ensure mobile number is in attributes
	if attrs == nil {
		attrs = make(map[string]interface{})
	}
	if _, exists := attrs[userAttributeMobileNumber]; !exists {
		attrs[userAttributeMobileNumber] = mobileNumber
	}

	authenticatedUser := &authncm.AuthenticatedUser{
		IsAuthenticated:    true,
		UserID:             user.UserID,
		OrganizationUnitID: user.OrganizationUnitID,
		UserType:           user.UserType,
		Attributes:         attrs,
	}

	return authenticatedUser, nil
}
