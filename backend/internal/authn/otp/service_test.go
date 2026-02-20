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

package otp

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/authn/common"
	notifcommon "github.com/asgardeo/thunder/internal/notification/common"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/userprovider"
	"github.com/asgardeo/thunder/tests/mocks/notification/notificationmock"
	"github.com/asgardeo/thunder/tests/mocks/userprovidermock"
)

const (
	testSenderID     = "sender123"
	testSessionToken = "token123"
)

type OTPAuthnServiceTestSuite struct {
	suite.Suite
	mockOTPService  *notificationmock.OTPServiceInterfaceMock
	mockUserService *userprovidermock.UserProviderInterfaceMock
	service         OTPAuthnServiceInterface
}

func TestOTPAuthnServiceTestSuite(t *testing.T) {
	suite.Run(t, new(OTPAuthnServiceTestSuite))
}

func (suite *OTPAuthnServiceTestSuite) SetupTest() {
	suite.mockOTPService = notificationmock.NewOTPServiceInterfaceMock(suite.T())
	suite.mockUserService = userprovidermock.NewUserProviderInterfaceMock(suite.T())
	suite.service = newOTPAuthnService(suite.mockOTPService, suite.mockUserService)
}

func (suite *OTPAuthnServiceTestSuite) TestSendOTPSuccess() {
	channel := notifcommon.ChannelTypeSMS
	recipient := "+1234567890"

	result := &notifcommon.SendOTPResultDTO{
		SessionToken: testSessionToken,
	}

	suite.mockOTPService.On("SendOTP", mock.MatchedBy(func(dto notifcommon.SendOTPDTO) bool {
		return dto.SenderID == testSenderID && dto.Channel == string(channel) && dto.Recipient == recipient
	})).Return(result, nil)

	token, err := suite.service.SendOTP(testSenderID, channel, recipient)
	suite.Nil(err)
	suite.Equal(testSessionToken, token)
}

func (suite *OTPAuthnServiceTestSuite) TestSendOTPInvalidInputs() {
	tests := []struct {
		name         string
		senderID     string
		channel      notifcommon.ChannelType
		recipient    string
		expectedCode string
	}{
		{
			"EmptySenderID",
			"",
			notifcommon.ChannelTypeSMS,
			"+1234567890",
			ErrorInvalidSenderID.Code,
		},
		{
			"EmptyRecipient",
			testSenderID,
			notifcommon.ChannelTypeSMS,
			"",
			ErrorInvalidRecipient.Code,
		},
		{
			"UnsupportedChannel",
			testSenderID,
			notifcommon.ChannelType("email"),
			"test@example.com",
			ErrorUnsupportedChannel.Code,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			token, err := suite.service.SendOTP(tc.senderID, tc.channel, tc.recipient)
			suite.Empty(token)
			suite.NotNil(err)
			suite.Equal(tc.expectedCode, err.Code)
		})
	}
}

func (suite *OTPAuthnServiceTestSuite) TestSendOTPWithServiceError() {
	tests := []struct {
		name               string
		mockReturnErr      *serviceerror.ServiceError
		expectedErrCode    string
		expectedDescSubstr string
	}{
		{
			name: "ServiceError",
			mockReturnErr: &serviceerror.ServiceError{
				Type:             serviceerror.ServerErrorType,
				Code:             "INTERNAL_ERROR",
				ErrorDescription: "Service unavailable",
			},
			expectedErrCode: serviceerror.InternalServerError.Code,
		},
		{
			name: "ClientError",
			mockReturnErr: &serviceerror.ServiceError{
				Type:             serviceerror.ClientErrorType,
				Code:             "INVALID_FORMAT",
				ErrorDescription: "Invalid phone number format",
			},
			expectedErrCode:    ErrorClientErrorFromOTPService.Code,
			expectedDescSubstr: "Invalid phone number format",
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			freshOTP := notificationmock.NewOTPServiceInterfaceMock(suite.T())
			suite.service = newOTPAuthnService(freshOTP, suite.mockUserService)
			freshOTP.On("SendOTP", mock.Anything).Return(nil, tc.mockReturnErr)

			token, err := suite.service.SendOTP(testSenderID, notifcommon.ChannelTypeSMS, "+1234567890")
			suite.Empty(token)
			suite.NotNil(err)
			suite.Equal(tc.expectedErrCode, err.Code)

			if tc.expectedDescSubstr != "" {
				suite.Contains(err.ErrorDescription, tc.expectedDescSubstr)
			}
		})
	}
}

func (suite *OTPAuthnServiceTestSuite) TestVerifyOTPSuccess() {
	otp := "123456"
	recipient := "+1234567890"
	userID := "user123"
	orgUnit := "test-ou"

	verifyResult := &notifcommon.VerifyOTPResultDTO{
		Status:    notifcommon.OTPVerifyStatusVerified,
		Recipient: recipient,
	}
	user := &userprovider.User{
		UserID:             userID,
		UserType:           "person",
		OrganizationUnitID: orgUnit,
	}

	suite.mockOTPService.On("VerifyOTP", mock.MatchedBy(func(dto notifcommon.VerifyOTPDTO) bool {
		return dto.SessionToken == testSessionToken && dto.OTPCode == otp
	})).Return(verifyResult, nil)
	suite.mockUserService.On("IdentifyUser", mock.MatchedBy(func(filters map[string]interface{}) bool {
		return filters["mobileNumber"] == recipient
	})).Return(&userID, nil)
	suite.mockUserService.On("GetUser", userID).Return(user, nil)

	result, err := suite.service.VerifyOTP(testSessionToken, otp)
	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(userID, result.UserID)
	suite.Equal(orgUnit, result.OrganizationUnitID)
}

func (suite *OTPAuthnServiceTestSuite) TestVerifyOTPWithInvalidInputs() {
	tests := []struct {
		name         string
		sessionToken string
		otp          string
		expectedCode string
	}{
		{
			"EmptySessionToken",
			"",
			"123456",
			ErrorInvalidSessionToken.Code,
		},
		{
			"EmptyOTP",
			testSessionToken,
			"",
			ErrorInvalidOTP.Code,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result, err := suite.service.VerifyOTP(tc.sessionToken, tc.otp)
			suite.Nil(result)
			suite.NotNil(err)
			suite.Equal(tc.expectedCode, err.Code)
		})
	}
}

func (suite *OTPAuthnServiceTestSuite) TestVerifyOTPWithIncorrectOTP() {
	verifyResult := &notifcommon.VerifyOTPResultDTO{
		Status:    notifcommon.OTPVerifyStatusInvalid,
		Recipient: "+1234567890",
	}

	suite.mockOTPService.On("VerifyOTP", mock.Anything).Return(verifyResult, nil)

	result, err := suite.service.VerifyOTP(testSessionToken, "123456")
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorIncorrectOTP.Code, err.Code)
}

func (suite *OTPAuthnServiceTestSuite) TestVerifyOTPWithOTPServiceError() {
	tests := []struct {
		name               string
		mockReturnErr      *serviceerror.ServiceError
		expectedErrCode    string
		expectedDescSubstr string
	}{
		{
			name: "ServiceError",
			mockReturnErr: &serviceerror.ServiceError{
				Type:             serviceerror.ServerErrorType,
				Code:             "INTERNAL_ERROR",
				ErrorDescription: "Service unavailable",
			},
			expectedErrCode: serviceerror.InternalServerError.Code,
		},
		{
			name: "ClientError",
			mockReturnErr: &serviceerror.ServiceError{
				Type:             serviceerror.ClientErrorType,
				Code:             "OTP_EXPIRED",
				ErrorDescription: "OTP has expired",
			},
			expectedErrCode:    ErrorClientErrorFromOTPService.Code,
			expectedDescSubstr: "OTP has expired",
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			freshOTP := notificationmock.NewOTPServiceInterfaceMock(suite.T())
			suite.service = newOTPAuthnService(freshOTP, suite.mockUserService)
			freshOTP.On("VerifyOTP", mock.Anything).Return(nil, tc.mockReturnErr)

			result, err := suite.service.VerifyOTP(testSessionToken, "123456")
			suite.Nil(result)
			suite.NotNil(err)
			suite.Equal(tc.expectedErrCode, err.Code)

			if tc.expectedDescSubstr != "" {
				suite.Contains(err.ErrorDescription, tc.expectedDescSubstr)
			}
		})
	}
}

func (suite *OTPAuthnServiceTestSuite) TestVerifyOTPWithUserServiceError() {
	verifyResult := &notifcommon.VerifyOTPResultDTO{
		Status:    notifcommon.OTPVerifyStatusVerified,
		Recipient: "+1234567890",
	}
	serverErr := &userprovider.UserProviderError{
		Code:        userprovider.ErrorCodeSystemError,
		Description: "Database unavailable",
	}

	// Prepare a userID for cases that require a valid identify result
	userID := "user123"

	tests := []struct {
		name         string
		identifyRet  *string
		identifyErr  *userprovider.UserProviderError
		getUserRet   *userprovider.User
		getUserErr   *userprovider.UserProviderError
		expectedCode string
	}{
		{
			"NonExistentUser",
			nil,
			&userprovider.UserProviderError{Code: userprovider.ErrorCodeUserNotFound},
			nil,
			nil,
			common.ErrorUserNotFound.Code,
		},
		{
			"IdentifyServerError",
			nil,
			serverErr,
			nil,
			nil,
			serviceerror.InternalServerError.Code,
		},
		{
			"GetUserServerError",
			&userID,
			nil,
			nil,
			serverErr,
			serviceerror.InternalServerError.Code,
		},
		{
			"UserIDNil",
			nil,
			(*userprovider.UserProviderError)(nil),
			nil,
			nil,
			common.ErrorUserNotFound.Code,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			freshOTP := notificationmock.NewOTPServiceInterfaceMock(suite.T())
			freshUser := userprovidermock.NewUserProviderInterfaceMock(suite.T())
			suite.service = newOTPAuthnService(freshOTP, freshUser)

			freshOTP.On("VerifyOTP", mock.Anything).Return(verifyResult, nil)
			freshUser.On("IdentifyUser", mock.Anything).Return(tc.identifyRet, tc.identifyErr)

			// only set GetUser expectation when identify returns a user id
			if tc.getUserRet != nil || tc.getUserErr != nil {
				freshUser.On("GetUser", *tc.identifyRet).Return(tc.getUserRet, tc.getUserErr)
			}

			result, err := suite.service.VerifyOTP(testSessionToken, "123456")
			suite.Nil(result)
			suite.NotNil(err)
			suite.Equal(tc.expectedCode, err.Code)
		})
	}
}

func (suite *OTPAuthnServiceTestSuite) TestVerifyOTPWithEmptyRecipient() {
	verifyResult := &notifcommon.VerifyOTPResultDTO{
		Status:    notifcommon.OTPVerifyStatusVerified,
		Recipient: "",
	}
	suite.mockOTPService.On("VerifyOTP", mock.Anything).Return(verifyResult, nil)

	result, err := suite.service.VerifyOTP(testSessionToken, "123456")
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}
