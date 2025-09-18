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

// Package smsauth provides the implementation of SMS OTP authentication executor.
package smsauth

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"time"

	authndto "github.com/asgardeo/thunder/internal/authn/dto"
	"github.com/asgardeo/thunder/internal/executor/identify"
	flowconst "github.com/asgardeo/thunder/internal/flow/constants"
	flowmodel "github.com/asgardeo/thunder/internal/flow/model"
	"github.com/asgardeo/thunder/internal/notification/message/model"
	msgsenderprovider "github.com/asgardeo/thunder/internal/notification/message/provider"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/user/service"
)

const (
	loggerComponentName       = "SMSOTPAuthExecutor"
	userAttributeUserID       = "userID"
	userAttributeUsername     = "username"
	userAttributeMobileNumber = "mobileNumber"
	userAttributeEmail        = "email"
	userInputOTP              = "otp"
	errorInvalidOTP           = "invalid OTP provided"
)

// SMSOTPAuthExecutor implements the ExecutorInterface for SMS OTP authentication.
type SMSOTPAuthExecutor struct {
	*identify.IdentifyingExecutor
	internal    flowmodel.Executor
	userService service.UserServiceInterface
}

var _ flowmodel.ExecutorInterface = (*SMSOTPAuthExecutor)(nil)

// NewSMSOTPAuthExecutor creates a new instance of SMSOTPAuthExecutor.
func NewSMSOTPAuthExecutor(id, name string, properties map[string]string) *SMSOTPAuthExecutor {
	defaultInputs := []flowmodel.InputData{
		{
			Name:     userInputOTP,
			Type:     "string",
			Required: true,
		},
	}
	prerequisites := []flowmodel.InputData{
		{
			Name:     userAttributeMobileNumber,
			Type:     "string",
			Required: true,
		},
	}

	return &SMSOTPAuthExecutor{
		IdentifyingExecutor: identify.NewIdentifyingExecutor(id, name, properties),
		internal: *flowmodel.NewExecutor(id, name, defaultInputs, prerequisites,
			properties),
		userService: service.GetUserService(),
	}
}

// GetID returns the ID of the SMSOTPAuthExecutor.
func (s *SMSOTPAuthExecutor) GetID() string {
	return s.internal.GetID()
}

// GetName returns the name of the SMSOTPAuthExecutor.
func (s *SMSOTPAuthExecutor) GetName() string {
	return s.internal.GetName()
}

// GetProperties returns the properties of the SMSOTPAuthExecutor.
func (s *SMSOTPAuthExecutor) GetProperties() flowmodel.ExecutorProperties {
	return s.internal.Properties
}

// Execute executes the SMS OTP authentication logic.
func (s *SMSOTPAuthExecutor) Execute(ctx *flowmodel.NodeContext) (*flowmodel.ExecutorResponse, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorID, s.GetID()),
		log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Executing SMS OTP authentication executor")

	execResp := &flowmodel.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	if !s.ValidatePrerequisites(ctx, execResp) {
		logger.Debug("Prerequisites not met for SMS OTP authentication executor")
		return execResp, nil
	}

	if s.CheckInputData(ctx, execResp) {
		err := s.InitiateOTP(ctx, execResp)
		if err != nil {
			return execResp, err
		}
	} else {
		err := s.ProcessAuthFlowResponse(ctx, execResp)
		if err != nil {
			return execResp, err
		}
	}

	logger.Debug("SMS OTP authentication executor execution completed",
		log.String("status", string(execResp.Status)),
		log.Bool("isAuthenticated", execResp.AuthenticatedUser.IsAuthenticated))

	return execResp, nil
}

// InitiateOTP initiates the OTP sending process to the user's mobile number.
func (s *SMSOTPAuthExecutor) InitiateOTP(ctx *flowmodel.NodeContext,
	execResp *flowmodel.ExecutorResponse) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorID, s.GetID()), log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Sending SMS OTP to user")

	mobileNumber, err := s.getUserMobileFromContext(ctx)
	if err != nil {
		return err
	}

	var userID *string
	if ctx.AuthenticatedUser.IsAuthenticated {
		userIDVal, err := s.GetUserIDFromContext(ctx)
		if err != nil {
			logger.Error("Failed to retrieve user ID from context", log.Error(err))
			return fmt.Errorf("failed to retrieve user ID from context: %w", err)
		}
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
	if ctx.FlowType == flowconst.FlowTypeRegistration {
		if execResp.Status == flowconst.ExecFailure && execResp.FailureReason != "User not found" {
			logger.Error("Failed to identify user during registration flow", log.Error(err))
			return fmt.Errorf("failed to identify user during registration flow: %w", err)
		}

		if userID != nil && *userID != "" {
			// At this point, a unique user is found in the system. Hence fail the execution.
			execResp.Status = flowconst.ExecFailure
			execResp.FailureReason = "User already exists with the provided mobile number."
			return nil
		}

		execResp.Status = ""
		execResp.FailureReason = ""
	} else {
		if execResp.Status == flowconst.ExecFailure {
			return nil
		}
		ctx.RuntimeData[userAttributeUserID] = *userID
	}

	// Send the OTP to the user's mobile number.
	if err := s.generateAndSendOTP(mobileNumber, ctx, execResp, logger); err != nil {
		logger.Error("Failed to send OTP", log.Error(err))
		return fmt.Errorf("failed to send OTP: %w", err)
	}
	if execResp.Status == flowconst.ExecFailure {
		return nil
	}

	logger.Debug("SMS OTP sent successfully")
	execResp.Status = flowconst.ExecUserInputRequired
	return nil
}

// ProcessAuthFlowResponse processes the authentication flow response for SMS OTP.
func (s *SMSOTPAuthExecutor) ProcessAuthFlowResponse(ctx *flowmodel.NodeContext,
	execResp *flowmodel.ExecutorResponse) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorID, s.GetID()), log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Processing authentication flow response for SMS OTP")

	err := s.validateOTP(ctx, execResp, logger)
	if err != nil {
		logger.Error("Error validating OTP", log.Error(err))
		return fmt.Errorf("error validating OTP: %w", err)
	}
	if execResp.Status == flowconst.ExecFailure {
		return nil
	}

	authenticatedUser, err := s.getAuthenticatedUser(ctx, execResp)
	if err != nil {
		logger.Error("Failed to get authenticated user details", log.Error(err))
		return fmt.Errorf("failed to get authenticated user details: %w", err)
	}

	execResp.AuthenticatedUser = *authenticatedUser
	execResp.Status = flowconst.ExecComplete
	return nil
}

// GetDefaultExecutorInputs returns the default required input data for the SMSOTPAuthExecutor.
func (s *SMSOTPAuthExecutor) GetDefaultExecutorInputs() []flowmodel.InputData {
	return s.internal.DefaultExecutorInputs
}

// GetPrerequisites returns the prerequisites for the SMSOTPAuthExecutor.
func (s *SMSOTPAuthExecutor) GetPrerequisites() []flowmodel.InputData {
	return s.internal.Prerequisites
}

// CheckInputData checks if the required input data is provided in the context.
func (s *SMSOTPAuthExecutor) CheckInputData(ctx *flowmodel.NodeContext, execResp *flowmodel.ExecutorResponse) bool {
	return s.internal.CheckInputData(ctx, execResp)
}

// ValidatePrerequisites validates whether the prerequisites for the SMSOTPAuthExecutor are met.
func (s *SMSOTPAuthExecutor) ValidatePrerequisites(ctx *flowmodel.NodeContext,
	execResp *flowmodel.ExecutorResponse) bool {
	preReqMet := s.internal.ValidatePrerequisites(ctx, execResp)
	if preReqMet {
		return true
	}

	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorID, s.GetID()), log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Trying to satisfy prerequisites for SMS OTP authentication executor")

	// If the flow type is registration, skip the prerequisite validation.
	if ctx.FlowType == flowconst.FlowTypeRegistration {
		logger.Debug("Mobile number not provided in registration flow, prompting user for input")
		execResp.Status = flowconst.ExecUserInputRequired
		execResp.RequiredData = []flowmodel.InputData{
			{
				Name:     userAttributeMobileNumber,
				Type:     "string",
				Required: true,
			},
		}
		return false
	}

	s.satisfyPrerequisites(ctx, execResp)
	if execResp.Status == flowconst.ExecFailure || execResp.Status == flowconst.ExecUserInputRequired {
		return false
	}

	return s.internal.ValidatePrerequisites(ctx, execResp)
}

// GetUserIDFromContext retrieves the user ID from the context.
func (s *SMSOTPAuthExecutor) GetUserIDFromContext(ctx *flowmodel.NodeContext) (string, error) {
	return s.internal.GetUserIDFromContext(ctx)
}

// GetRequiredData returns the required input data for the SMSOTPAuthExecutor.
func (s *SMSOTPAuthExecutor) GetRequiredData(ctx *flowmodel.NodeContext) []flowmodel.InputData {
	return s.internal.GetRequiredData(ctx)
}

// getUserMobileFromContext retrieves the user's mobile number from the context.
func (s *SMSOTPAuthExecutor) getUserMobileFromContext(ctx *flowmodel.NodeContext) (string, error) {
	mobileNumber := ctx.RuntimeData[userAttributeMobileNumber]
	if mobileNumber == "" {
		mobileNumber = ctx.UserInputData[userAttributeMobileNumber]
	}
	if mobileNumber == "" {
		return "", errors.New("mobile number not found in context")
	}
	return mobileNumber, nil
}

// satisfyPrerequisites tries to satisfy the prerequisites for the SMSOTPAuthExecutor.
func (s *SMSOTPAuthExecutor) satisfyPrerequisites(ctx *flowmodel.NodeContext,
	execResp *flowmodel.ExecutorResponse) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorID, s.GetID()), log.String(log.LoggerKeyFlowID, ctx.FlowID))

	execResp.Status = ""
	execResp.FailureReason = ""

	logger.Debug("Trying to resolve user ID from context data")
	userIDResolved, err := s.resolveUserID(ctx)
	if err != nil {
		logger.Error("Failed to resolve user ID from context data", log.Error(err))
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = "Failed to resolve user ID from context data"
		return
	}
	if !userIDResolved {
		logger.Debug("User ID could not be resolved from context data")
		execResp.Status = flowconst.ExecFailure
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
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = "Failed to retrieve mobile number"
		return
	}
	if execResp.Status == flowconst.ExecFailure || execResp.Status == flowconst.ExecUserInputRequired {
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
func (s *SMSOTPAuthExecutor) resolveUserID(ctx *flowmodel.NodeContext) (bool, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorID, s.GetID()), log.String(log.LoggerKeyFlowID, ctx.FlowID))

	// First, check if the user ID is already available in the context.
	userID, err := s.GetUserIDFromContext(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to retrieve user ID from context: %w", err)
	}
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
	userIDResolved, err = s.resolveUserIDFromAttribute(ctx, userAttributeMobileNumber, logger)
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
func (s *SMSOTPAuthExecutor) resolveUserIDFromAttribute(ctx *flowmodel.NodeContext,
	attributeName string, logger *log.Logger) (bool, error) {
	logger.Debug("Resolving user ID from attribute", log.String("attributeName", attributeName))

	attributeValue := ctx.UserInputData[attributeName]
	if attributeValue == "" {
		attributeValue = ctx.RuntimeData[attributeName]
	}
	if attributeValue != "" {
		filters := map[string]interface{}{attributeName: attributeValue}
		userID, svcErr := s.userService.IdentifyUser(filters)
		if svcErr != nil {
			return false, fmt.Errorf("failed to identify user by %s: %s", attributeName, svcErr.Error)
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
func (s *SMSOTPAuthExecutor) getUserMobileNumber(userID string, ctx *flowmodel.NodeContext,
	execResp *flowmodel.ExecutorResponse) (string, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorID, s.GetID()), log.String(log.LoggerKeyFlowID, ctx.FlowID),
		log.String("userID", userID))
	logger.Debug("Retrieving user mobile number")

	var err error
	user, svcErr := s.userService.GetUser(userID)
	if svcErr != nil {
		return "", fmt.Errorf("failed to retrieve user details: %s", svcErr.Error)
	}

	// Extract mobile number from user attributes
	var attrs map[string]interface{}
	if err := json.Unmarshal(user.Attributes, &attrs); err != nil {
		return "", fmt.Errorf("failed to unmarshal user attributes: %w", err)
	}

	mobileNumber := ""
	mobileNumberAttr := attrs[userAttributeMobileNumber]
	if mobileNumberAttr != nil && mobileNumberAttr != "" {
		mobileNumber = mobileNumberAttr.(string)
	}

	if mobileNumber == "" {
		// If the user is not authenticated, return an error.
		// TODO: Revisit this logic when implementing registration flows/ JIT provisioning.
		if !ctx.AuthenticatedUser.IsAuthenticated {
			logger.Debug("Mobile number not found in user attributes")
			execResp.Status = flowconst.ExecFailure
			execResp.FailureReason = "Mobile number not found in user attributes"
			return "", nil
		}

		// If the user is authenticated, try to retrieve the mobile number from context or prompt for user input.
		mobileNumber, err = s.getUserMobileFromContext(ctx)
		if err != nil {
			logger.Debug("Mobile number not found in user attributes or context. Prompting user for input")
			execResp.Status = flowconst.ExecUserInputRequired
			// Here mobile number is treated as a special attribute. Hence rest of the data will be overridden.
			execResp.RequiredData = []flowmodel.InputData{
				{
					Name:     userAttributeMobileNumber,
					Type:     "string",
					Required: true,
				},
			}
			return "", nil
		}
	}

	return mobileNumber, nil
}

// generateAndSendOTP generates an OTP and sends it to the user's mobile number.
func (s *SMSOTPAuthExecutor) generateAndSendOTP(mobileNumber string, ctx *flowmodel.NodeContext,
	execResp *flowmodel.ExecutorResponse, logger *log.Logger) error {
	attemptCount, err := s.validateAttempts(ctx, execResp, logger)
	if err != nil {
		return fmt.Errorf("failed to validate OTP attempts: %w", err)
	}
	if execResp.Status == flowconst.ExecFailure {
		return nil
	}

	otp, err := s.generateOTP()
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	logger.Debug("Sending OTP to user's mobile")

	smsData := model.SMSData{
		To: mobileNumber,
		// TODO: This should be handled with a SMS template.
		Body: fmt.Sprintf("Your verification code is: %s. This code is valid for %d minutes.",
			otp.Value, otp.ValidityPeriodInMillis/60000),
	}

	// Get the message sender name from executor properties.
	execProps := s.GetProperties().Properties
	if len(execProps) == 0 {
		return errors.New("message sender name is not configured in executor properties")
	}
	senderName, ok := execProps["senderName"]
	if !ok || senderName == "" {
		return errors.New("message sender name is not configured in executor properties")
	}

	// Send the SMS OTP.
	provider := msgsenderprovider.NewNotificationServiceProvider()
	service := provider.GetMessageClientService()
	msgClient, svcErr := service.GetMessageClientByName(senderName)
	if svcErr != nil {
		logger.Error("Failed to get message client", log.String("senderName", senderName),
			log.Any("serviceError", svcErr))
		return fmt.Errorf("failed to get message client: %s", svcErr.ErrorDescription)
	}
	if msgClient == nil {
		return fmt.Errorf("message client %s not found", senderName)
	}

	if err := msgClient.SendSMS(smsData); err != nil {
		return fmt.Errorf("failed to send SMS: %w", err)
	}

	// Store runtime data.
	if execResp.RuntimeData == nil {
		execResp.RuntimeData = make(map[string]string)
	}
	execResp.RuntimeData["value"] = otp.Value
	execResp.RuntimeData["expiryTimeInMillis"] = fmt.Sprintf("%d", otp.ExpiryTimeInMillis)
	execResp.RuntimeData["attemptCount"] = strconv.Itoa(attemptCount + 1)

	return nil
}

// validateAttempts checks if the maximum number of OTP attempts has been reached.
func (s *SMSOTPAuthExecutor) validateAttempts(ctx *flowmodel.NodeContext, execResp *flowmodel.ExecutorResponse,
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
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = fmt.Sprintf("maximum OTP attempts reached: %d", attemptCount)
		return 0, nil
	}

	return attemptCount, nil
}

// generateOTP generates a random OTP based on the configurations.
func (s *SMSOTPAuthExecutor) generateOTP() (model.OTP, error) {
	charSet := s.getOTPCharset()
	otpLength := s.getOTPLength()

	chars := []rune(charSet)
	result := make([]rune, otpLength)

	for i := 0; i < otpLength; i++ {
		max := big.NewInt(int64(len(chars)))
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return model.OTP{}, fmt.Errorf("failed to generate random number: %w", err)
		}
		result[i] = chars[n.Int64()]
	}

	token := string(result)
	currentTime := time.Now().UnixMilli()
	validityPeriod := s.getOTPValidityPeriodInMillis()

	return model.OTP{
		Value:                  token,
		GeneratedTimeInMillis:  currentTime,
		ValidityPeriodInMillis: validityPeriod,
		ExpiryTimeInMillis:     currentTime + validityPeriod,
	}, nil
}

// getOTPCharset returns the character set for OTP generation.
func (s *SMSOTPAuthExecutor) getOTPCharset() string {
	if s.useOnlyNumericChars() {
		return "9245378016"
	}
	return "KIGXHOYSPRWCEFMVUQLZDNABJT9245378016"
}

// getOTPMaxAttempts returns the maximum number of attempts allowed for OTP validation.
func (s *SMSOTPAuthExecutor) getOTPMaxAttempts() int {
	// TODO: This needs to be configured as a IDP property.
	return 3
}

// getOTPLength returns the length of the OTP.
func (s *SMSOTPAuthExecutor) getOTPLength() int {
	// TODO: This needs to be configured as a IDP property.
	return 6
}

// useOnlyNumericChars determines whether to use only numeric characters.
func (s *SMSOTPAuthExecutor) useOnlyNumericChars() bool {
	// TODO: This needs to be configured as a IDP property.
	return true
}

// getOTPValidityPeriodInMillis returns the validity period of the OTP in milliseconds.
func (s *SMSOTPAuthExecutor) getOTPValidityPeriodInMillis() int64 {
	// TODO: This needs to be configured as a IDP property.
	return 120000 // 2 minutes
}

// validateOTP validates the OTP for the given user and mobile number.
func (s *SMSOTPAuthExecutor) validateOTP(ctx *flowmodel.NodeContext, execResp *flowmodel.ExecutorResponse,
	logger *log.Logger) error {
	userID := ctx.RuntimeData[userAttributeUserID]
	providedOTP := ctx.UserInputData[userInputOTP]

	logger.Debug("Validating OTP", log.String("userID", userID))

	if providedOTP == "" {
		logger.Debug("Provided OTP is empty", log.String("userID", userID))
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = errorInvalidOTP
		return nil
	}

	storedOTP := ctx.RuntimeData["value"]
	if storedOTP == "" {
		logger.Debug("No stored OTP found for validation", log.String("userID", userID))
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = errorInvalidOTP
		return nil
	}

	// Validate for the OTP value
	if providedOTP != storedOTP {
		logger.Debug("Provided OTP does not match stored OTP", log.String("userID", userID))
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = errorInvalidOTP
		return nil
	}

	// Validate for the expiry time of the OTP
	expiryTimeStr := ctx.RuntimeData["expiryTimeInMillis"]
	if expiryTimeStr == "" {
		logger.Debug("No expiry time found for stored OTP", log.String("userID", userID))
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = errorInvalidOTP
		return nil
	}
	expiryTime, err := strconv.ParseInt(expiryTimeStr, 10, 64)
	if err != nil {
		logger.Error("Failed to parse expiry time for OTP", log.String("userID", userID),
			log.Error(err))
		return errors.New("something went wrong while validating the OTP")
	}

	currentTime := time.Now().UnixMilli()
	if currentTime > expiryTime {
		execResp.RuntimeData["value"] = ""
		execResp.RuntimeData["expiryTimeInMillis"] = ""
		logger.Debug("OTP has expired", log.String("userID", userID))
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = "OTP has expired"
		return nil
	}

	execResp.RuntimeData["value"] = ""
	execResp.RuntimeData["expiryTimeInMillis"] = ""
	logger.Debug("OTP validated successfully", log.String("userID", userID))
	return nil
}

// getAuthenticatedUser returns the authenticated user details for the given user ID.
func (s *SMSOTPAuthExecutor) getAuthenticatedUser(ctx *flowmodel.NodeContext,
	execResp *flowmodel.ExecutorResponse) (*authndto.AuthenticatedUser, error) {
	mobileNumber, err := s.getUserMobileFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Handle registration flows.
	if ctx.FlowType == flowconst.FlowTypeRegistration {
		execResp.Status = flowconst.ExecComplete
		execResp.FailureReason = ""
		return &authndto.AuthenticatedUser{
			IsAuthenticated: false,
			Attributes: map[string]interface{}{
				userAttributeMobileNumber: mobileNumber,
			},
		}, nil
	}

	userID := ctx.RuntimeData[userAttributeUserID]
	if userID == "" {
		return nil, errors.New("user ID is empty")
	}

	user, svcErr := s.userService.GetUser(userID)
	if svcErr != nil {
		return nil, fmt.Errorf("failed to get user details: %s", svcErr.Error)
	}

	// Extract user attributes
	var attrs map[string]interface{}
	if err := json.Unmarshal(user.Attributes, &attrs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user attributes: %w", err)
	}

	authenticatedUser := &authndto.AuthenticatedUser{
		IsAuthenticated: true,
		UserID:          user.ID,
		Attributes:      attrs,
	}

	return authenticatedUser, nil
}
