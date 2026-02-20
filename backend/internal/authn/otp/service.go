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

// Package otp implements the OTP authentication service.
package otp

import (
	"fmt"
	"slices"
	"strings"

	"github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/notification"
	notifcommon "github.com/asgardeo/thunder/internal/notification/common"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/userprovider"
)

const (
	loggerComponentName       = "OTPAuthnService"
	userAttributeMobileNumber = "mobileNumber"
)

var supportedChannels = []notifcommon.ChannelType{notifcommon.ChannelTypeSMS}

// OTPAuthnServiceInterface defines the interface for OTP authentication operations.
// This is a wrapper over the notification.OTPServiceInterface to perform user authentication.
type OTPAuthnServiceInterface interface {
	SendOTP(senderID string, channel notifcommon.ChannelType, recipient string) (string, *serviceerror.ServiceError)
	VerifyOTP(sessionToken, otp string) (*userprovider.User, *serviceerror.ServiceError)
}

// otpAuthnService is the default implementation of OTPAuthnServiceInterface.
type otpAuthnService struct {
	otpService   notification.OTPServiceInterface
	userProvider userprovider.UserProviderInterface
}

// newOTPAuthnService creates a new instance of OTPAuthnService.
func newOTPAuthnService(otpSvc notification.OTPServiceInterface,
	userProvider userprovider.UserProviderInterface) OTPAuthnServiceInterface {
	service := &otpAuthnService{
		otpService:   otpSvc,
		userProvider: userProvider,
	}
	common.RegisterAuthenticator(service.getMetadata())

	return service
}

// SendOTP sends an OTP to the specified recipient using the provided sender.
func (s *otpAuthnService) SendOTP(senderID string, channel notifcommon.ChannelType,
	recipient string) (string, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Sending OTP for authentication", log.String("recipient", log.MaskString(recipient)),
		log.String("channel", string(channel)))

	if svcErr := s.validateOTPSendRequest(senderID, channel, recipient); svcErr != nil {
		return "", svcErr
	}

	otpData := notifcommon.SendOTPDTO{
		SenderID:  senderID,
		Channel:   string(channel),
		Recipient: recipient,
	}
	result, svcErr := s.otpService.SendOTP(otpData)
	if svcErr != nil {
		return "", s.handleOTPServiceError(svcErr, false, logger)
	}

	logger.Debug("OTP sent successfully, session token generated")
	return result.SessionToken, nil
}

// VerifyOTP verifies the provided OTP against the session token and returns the authenticated user.
func (s *otpAuthnService) VerifyOTP(sessionToken, otp string) (*userprovider.User, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Verifying OTP for authentication")

	if svcErr := s.validateOTPVerifyRequest(sessionToken, otp); svcErr != nil {
		return nil, svcErr
	}

	verifyData := notifcommon.VerifyOTPDTO{
		SessionToken: sessionToken,
		OTPCode:      otp,
	}
	result, svcErr := s.otpService.VerifyOTP(verifyData)
	if svcErr != nil {
		return nil, s.handleOTPServiceError(svcErr, true, logger)
	}

	return s.handleVerifyOTPResponse(result, logger)
}

// validateOTPSendRequest validates the parameters for sending an OTP.
func (s *otpAuthnService) validateOTPSendRequest(senderID string, channel notifcommon.ChannelType,
	recipient string) *serviceerror.ServiceError {
	if strings.TrimSpace(senderID) == "" {
		return &ErrorInvalidSenderID
	}
	if strings.TrimSpace(recipient) == "" {
		return &ErrorInvalidRecipient
	}
	if !slices.Contains(supportedChannels, channel) {
		return &ErrorUnsupportedChannel
	}
	return nil
}

// handleOTPServiceError handles errors from the OTP service.
func (s *otpAuthnService) handleOTPServiceError(svcErr *serviceerror.ServiceError, isVerify bool,
	logger *log.Logger) *serviceerror.ServiceError {
	if svcErr.Type == serviceerror.ClientErrorType {
		if isVerify {
			return serviceerror.CustomServiceError(ErrorClientErrorFromOTPService,
				fmt.Sprintf("Error verifying OTP: %s", svcErr.ErrorDescription))
		} else {
			return serviceerror.CustomServiceError(ErrorClientErrorFromOTPService,
				fmt.Sprintf("Error sending OTP: %s", svcErr.ErrorDescription))
		}
	}

	if isVerify {
		logger.Error("Error occurred while verifying OTP", log.Any("error", svcErr))
	} else {
		logger.Error("Error occurred while sending OTP", log.Any("error", svcErr))
	}
	return &serviceerror.InternalServerError
}

// validateOTPVerifyRequest validates the parameters for verifying an OTP.
func (s *otpAuthnService) validateOTPVerifyRequest(sessionToken, otp string) *serviceerror.ServiceError {
	if strings.TrimSpace(sessionToken) == "" {
		return &ErrorInvalidSessionToken
	}
	if strings.TrimSpace(otp) == "" {
		return &ErrorInvalidOTP
	}
	return nil
}

// handleVerifyOTPResponse processes the OTP verification result and resolves the user.
func (s *otpAuthnService) handleVerifyOTPResponse(result *notifcommon.VerifyOTPResultDTO,
	logger *log.Logger) (*userprovider.User, *serviceerror.ServiceError) {
	if result.Status != notifcommon.OTPVerifyStatusVerified {
		return nil, &ErrorIncorrectOTP
	}

	if result.Recipient == "" {
		logger.Error("Recipient not found in OTP verification result")
		return nil, &serviceerror.InternalServerError
	}

	user, svcErr := s.resolveUser(result.Recipient, notifcommon.ChannelTypeSMS, logger)
	if svcErr != nil {
		return nil, svcErr
	}

	return user, nil
}

// resolveUser retrieves a user by their recipient identifier (e.g., mobile number).
func (s *otpAuthnService) resolveUser(recipient string, channel notifcommon.ChannelType,
	logger *log.Logger) (*userprovider.User, *serviceerror.ServiceError) {
	logger.Debug("Resolving user from recipient", log.String("recipient", log.MaskString(recipient)),
		log.String("channel", string(channel)))

	// Build filter based on channel type
	filters := make(map[string]interface{})
	switch channel {
	case notifcommon.ChannelTypeSMS:
		filters[userAttributeMobileNumber] = recipient
	default:
		return nil, &ErrorUnsupportedChannel
	}

	userID, upErr := s.userProvider.IdentifyUser(filters)
	if upErr != nil {
		return nil, s.handleUserProviderError(upErr, logger)
	}
	if userID == nil || *userID == "" {
		logger.Debug("No user found for recipient", log.String("recipient", log.MaskString(recipient)))
		return nil, &common.ErrorUserNotFound
	}

	user, upErr := s.userProvider.GetUser(*userID)
	if upErr != nil {
		return nil, s.handleUserProviderError(upErr, logger)
	}

	logger.Debug("User resolved from recipient", log.String("userId", user.UserID))
	return user, nil
}

// handleUserProviderError handles errors from the user provider.
func (s *otpAuthnService) handleUserProviderError(upErr *userprovider.UserProviderError,
	logger *log.Logger) *serviceerror.ServiceError {
	if upErr.Code == userprovider.ErrorCodeUserNotFound {
		return &common.ErrorUserNotFound
	}
	if upErr.Code == userprovider.ErrorCodeSystemError {
		logger.Error("Error occurred while retrieving user", log.Any("error", upErr))
		return &serviceerror.InternalServerError
	}
	return serviceerror.CustomServiceError(ErrorClientErrorWhileResolvingUser,
		fmt.Sprintf("An error occurred while retrieving user: %s", upErr.Description))
}

// getMetadata returns the authenticator metadata for OTP authenticator.
func (s *otpAuthnService) getMetadata() common.AuthenticatorMeta {
	return common.AuthenticatorMeta{
		Name:    common.AuthenticatorSMSOTP,
		Factors: []common.AuthenticationFactor{common.FactorPossession},
	}
}
