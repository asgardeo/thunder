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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	authncm "github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	notifcommon "github.com/asgardeo/thunder/internal/notification/common"
	"github.com/asgardeo/thunder/internal/observability/event"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/user"
	"github.com/asgardeo/thunder/tests/mocks/flow/coremock"
	"github.com/asgardeo/thunder/tests/mocks/notification/notificationmock"
	"github.com/asgardeo/thunder/tests/mocks/observabilitymock"
	"github.com/asgardeo/thunder/tests/mocks/usermock"
)

const (
	testSessionToken = "session-123"
)

type SMSAuthExecutorTestSuite struct {
	suite.Suite
	mockUserService   *usermock.UserServiceInterfaceMock
	mockOTPService    *notificationmock.OTPServiceInterfaceMock
	mockFlowFactory   *coremock.FlowFactoryInterfaceMock
	mockObservability *observabilitymock.ObservabilityServiceInterfaceMock
	executor          *smsOTPAuthExecutor
}

func TestSMSAuthExecutorSuite(t *testing.T) {
	suite.Run(t, new(SMSAuthExecutorTestSuite))
}

func (suite *SMSAuthExecutorTestSuite) SetupTest() {
	suite.mockUserService = usermock.NewUserServiceInterfaceMock(suite.T())
	suite.mockOTPService = notificationmock.NewOTPServiceInterfaceMock(suite.T())
	suite.mockFlowFactory = coremock.NewFlowFactoryInterfaceMock(suite.T())
	suite.mockObservability = observabilitymock.NewObservabilityServiceInterfaceMock(suite.T())

	// Default behavior for observability: disabled
	suite.mockObservability.On("IsEnabled").Return(false).Maybe()

	defaultInputs := []common.Input{
		{
			Identifier: userInputOTP,
			Type:       "string",
			Required:   true,
		},
	}
	prerequisites := []common.Input{
		{
			Identifier: userAttributeMobileNumber,
			Type:       "string",
			Required:   true,
		},
	}

	// Mock identifying executor
	identifyingMock := createMockIdentifyingExecutor(suite.T())
	suite.mockFlowFactory.On("CreateExecutor", ExecutorNameIdentifying, common.ExecutorTypeUtility,
		mock.Anything, mock.Anything).Return(identifyingMock).Maybe()

	// Mock base executor
	mockExec := coremock.NewExecutorInterfaceMock(suite.T())
	mockExec.On("GetName").Return(ExecutorNameSMSAuth).Maybe()
	mockExec.On("GetType").Return(common.ExecutorTypeAuthentication).Maybe()
	mockExec.On("GetDefaultInputs").Return(defaultInputs).Maybe()
	mockExec.On("GetRequiredInputs", mock.Anything).Return(defaultInputs).Maybe()
	mockExec.On("GetPrerequisites").Return(prerequisites).Maybe()
	mockExec.On("ValidatePrerequisites", mock.Anything, mock.Anything).Return(true).Maybe()
	mockExec.On("HasRequiredInputs", mock.Anything, mock.Anything).Return(
		func(ctx *core.NodeContext, execResp *common.ExecutorResponse) bool {
			otp, exists := ctx.UserInputs[userInputOTP]
			if !exists || otp == "" {
				execResp.Inputs = defaultInputs
				execResp.Status = common.ExecUserInputRequired
				return false
			}
			return true
		}).Maybe()

	suite.mockFlowFactory.On("CreateExecutor", ExecutorNameSMSAuth, common.ExecutorTypeAuthentication,
		defaultInputs, prerequisites).Return(mockExec)

	suite.executor = newSMSOTPAuthExecutor(suite.mockFlowFactory, suite.mockUserService,
		suite.mockOTPService, suite.mockObservability)
	// Inject the mock base executor
	suite.executor.ExecutorInterface = mockExec
}

func (suite *SMSAuthExecutorTestSuite) BeforeTest(suiteName, testName string) {
	suite.mockObservability.ExpectedCalls = nil
	suite.mockObservability.On("IsEnabled").Return(false).Maybe()
}

func (suite *SMSAuthExecutorTestSuite) TestExecute_Observability_Success() {
	// Enable observability
	suite.mockObservability.ExpectedCalls = nil
	suite.mockObservability.On("IsEnabled").Return(true)

	userID := testUserID
	mobileNumber := "+1234567890"
	sessionToken := testSessionToken

	// Mock dependencies for validation flow
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("GetRequiredInputs", mock.Anything).Return([]common.Input{
		{Identifier: userInputOTP, Type: "string", Required: true},
	}).Maybe()
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("HasRequiredInputs", mock.Anything, mock.Anything).
		Return(func(ctx *core.NodeContext, execResp *common.ExecutorResponse) bool {
			otp := ctx.UserInputs[userInputOTP]
			if otp == "" {
				execResp.Inputs = []common.Input{{Identifier: userInputOTP, Type: "string", Required: true}}
				execResp.Status = common.ExecUserInputRequired
				return false
			}
			return true
		})

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(true)

	ctx := &core.NodeContext{
		FlowID:        "flow-123",
		AppID:         "app-1",
		CurrentNodeID: "node-1",
		FlowType:      common.FlowTypeAuthentication,
		ExecutorMode:  smsOTPExecutorModeVerify,
		UserInputs: map[string]string{
			userInputOTP: "123456",
		},
		RuntimeData: map[string]string{
			userAttributeUserID:       userID,
			userAttributeMobileNumber: mobileNumber,
			"otpSessionToken":         sessionToken,
		},
	}

	attrs := map[string]interface{}{userAttributeMobileNumber: mobileNumber}
	attrsJSON, _ := json.Marshal(attrs)
	userObj := &user.User{
		ID:         testUserID,
		Attributes: attrsJSON,
	}

	suite.mockUserService.On("GetUser", userID).Return(userObj, nil)

	suite.mockOTPService.On("VerifyOTP", notifcommon.VerifyOTPDTO{
		SessionToken: sessionToken,
		OTPCode:      "123456",
	}).Return(&notifcommon.VerifyOTPResultDTO{
		Status: notifcommon.OTPVerifyStatusVerified,
	}, nil)

	// Expect Started event
	suite.mockObservability.On("PublishEvent", mock.MatchedBy(func(evt *event.Event) bool {
		return evt.Type == string(event.EventTypeFlowNodeExecutionStarted) &&
			evt.Status == event.StatusInProgress
	})).Return()

	// Expect Completed event
	suite.mockObservability.On("PublishEvent", mock.MatchedBy(func(evt *event.Event) bool {
		return evt.Type == string(event.EventTypeFlowNodeExecutionCompleted) &&
			evt.Status == event.StatusSuccess &&
			evt.Data[event.DataKey.UserID] == userID
	})).Return()

	resp, err := suite.executor.Execute(ctx)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)

	suite.mockObservability.AssertExpectations(suite.T())
}

func (suite *SMSAuthExecutorTestSuite) TestExecute_Observability_InitiateOTP() {
	// Enable observability
	suite.mockObservability.ExpectedCalls = nil
	suite.mockObservability.On("IsEnabled").Return(true)

	userID := testUserID
	mobileNumber := "+1234567890"

	// Mock GetRequiredInputs for send mode
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("GetRequiredInputs", mock.Anything).Return([]common.Input{
		{Identifier: userInputOTP, Type: "string", Required: true},
	}).Maybe()

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(true)

	// Expect IdentifyUser to be called because user is not authenticated in context
	suite.mockUserService.On("IdentifyUser", map[string]interface{}{
		userAttributeMobileNumber: mobileNumber,
	}).Return(&userID, nil)

	ctx := &core.NodeContext{
		FlowID:        "flow-123",
		AppID:         "app-1",
		CurrentNodeID: "node-1",
		FlowType:      common.FlowTypeAuthentication,
		ExecutorMode:  smsOTPExecutorModeSend,
		UserInputs:    map[string]string{},
		RuntimeData: map[string]string{
			userAttributeUserID:       userID,
			userAttributeMobileNumber: mobileNumber,
		},
		NodeProperties: map[string]interface{}{
			"senderId": "testSender",
		},
	}

	// InitiateOTP -> validateAttempts -> OK
	// InitiateOTP -> generateAndSendOTP -> SendOTP
	suite.mockOTPService.On("SendOTP", mock.MatchedBy(func(dto notifcommon.SendOTPDTO) bool {
		return dto.Recipient == mobileNumber
	})).Return(&notifcommon.SendOTPResultDTO{
		SessionToken: "new-session",
	}, nil)

	// Expect Started event
	suite.mockObservability.On("PublishEvent", mock.MatchedBy(func(evt *event.Event) bool {
		return evt.Type == string(event.EventTypeFlowNodeExecutionStarted)
	})).Return()

	// Expect Completed event (with pending status after OTP send)
	suite.mockObservability.On("PublishEvent", mock.MatchedBy(func(evt *event.Event) bool {
		return evt.Type == string(event.EventTypeFlowNodeExecutionCompleted) &&
			evt.Status == event.StatusPending &&
			evt.Data[event.DataKey.UserID] == userID
	})).Return()

	resp, err := suite.executor.Execute(ctx)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	suite.mockObservability.AssertExpectations(suite.T())
}

func (suite *SMSAuthExecutorTestSuite) TestExecute_Observability_Failure() {
	// Enable observability
	suite.mockObservability.ExpectedCalls = nil
	suite.mockObservability.On("IsEnabled").Return(true)

	userID := testUserID
	sessionToken := testSessionToken

	// Mock GetRequiredInputs to return FALSE (Process Response)
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("GetRequiredInputs", mock.Anything).Return([]common.Input{
		{Identifier: userInputOTP, Type: "string", Required: true},
	}).Maybe()
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("HasRequiredInputs", mock.Anything, mock.Anything).
		Return(func(ctx *core.NodeContext, execResp *common.ExecutorResponse) bool {
			otp := ctx.UserInputs[userInputOTP]
			if otp == "" {
				execResp.Inputs = []common.Input{{Identifier: userInputOTP, Type: "string", Required: true}}
				execResp.Status = common.ExecUserInputRequired
				return false
			}
			return true
		})

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(true)

	ctx := &core.NodeContext{
		FlowID:        "flow-123",
		AppID:         "app-1",
		CurrentNodeID: "node-1",
		FlowType:      common.FlowTypeAuthentication,
		ExecutorMode:  smsOTPExecutorModeVerify,
		UserInputs: map[string]string{
			userInputOTP: "wrong-otp",
		},
		RuntimeData: map[string]string{
			userAttributeUserID: userID,
			"otpSessionToken":   sessionToken,
		},
	}

	suite.mockOTPService.On("VerifyOTP", mock.Anything).Return(&notifcommon.VerifyOTPResultDTO{
		Status: notifcommon.OTPVerifyStatusInvalid,
	}, nil)

	// Expect Started event
	suite.mockObservability.On("PublishEvent", mock.MatchedBy(func(evt *event.Event) bool {
		return evt.Type == string(event.EventTypeFlowNodeExecutionStarted)
	})).Return()

	// Expect Failed event
	suite.mockObservability.On("PublishEvent", mock.MatchedBy(func(evt *event.Event) bool {
		return evt.Type == string(event.EventTypeFlowNodeExecutionFailed) &&
			evt.Status == event.StatusFailure &&
			evt.Data[event.DataKey.FailureReason] == errorInvalidOTP
	})).Return()

	resp, err := suite.executor.Execute(ctx)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)

	suite.mockObservability.AssertExpectations(suite.T())
}

func (suite *SMSAuthExecutorTestSuite) TestExecute_EmptyMode_ReturnsError() {
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(true)

	ctx := &core.NodeContext{
		FlowID:        "flow-123",
		AppID:         "app-1",
		CurrentNodeID: "node-1",
		FlowType:      common.FlowTypeAuthentication,
		ExecutorMode:  "", // Empty mode should return error
		UserInputs:    map[string]string{},
		RuntimeData: map[string]string{
			userAttributeMobileNumber: "+1234567890",
		},
	}

	resp, err := suite.executor.Execute(ctx)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "invalid executor mode")
	assert.NotNil(suite.T(), resp)
}

func (suite *SMSAuthExecutorTestSuite) TestExecute_UnknownMode() {
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(true)

	ctx := &core.NodeContext{
		FlowID:        "flow-123",
		AppID:         "app-1",
		CurrentNodeID: "node-1",
		FlowType:      common.FlowTypeAuthentication,
		ExecutorMode:  "unknown_mode", // Unknown mode
		UserInputs:    map[string]string{},
		RuntimeData: map[string]string{
			userAttributeMobileNumber: "+1234567890",
		},
	}

	resp, err := suite.executor.Execute(ctx)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "invalid executor mode: unknown_mode")
	assert.NotNil(suite.T(), resp)
}

func (suite *SMSAuthExecutorTestSuite) TestValidatePrerequisites_AlreadyMet() {
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(true)

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		RuntimeData: map[string]string{
			userAttributeMobileNumber: "+1234567890",
		},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	result := suite.executor.ValidatePrerequisites(ctx, execResp)
	assert.True(suite.T(), result)
}

func (suite *SMSAuthExecutorTestSuite) TestValidatePrerequisites_RegistrationFlow_EnforcesFailure() {
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(false) // Prerequisites not met initially

	ctx := &core.NodeContext{
		FlowID:      "flow-123",
		FlowType:    common.FlowTypeRegistration, // Registration flow
		RuntimeData: map[string]string{},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	result := suite.executor.ValidatePrerequisites(ctx, execResp)
	assert.False(suite.T(), result) // Should return false without trying to satisfy
}

func (suite *SMSAuthExecutorTestSuite) TestValidatePrerequisites_AuthFlow_SatisfiesPrerequisites() {
	userID := testUserID
	mobileNumber := "+1234567890"
	attrs := map[string]interface{}{userAttributeMobileNumber: mobileNumber}
	attrsJSON, _ := json.Marshal(attrs)

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	// First call returns false, second call (after satisfaction) returns true
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(false).Once()
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(true).Once()
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("GetUserIDFromContext", mock.Anything).
		Return("") // No user ID in context initially

	// Mock user lookup for prerequisite satisfaction
	suite.mockUserService.On("IdentifyUser", map[string]interface{}{
		userAttributeUsername: "testuser",
	}).Return(&userID, nil)

	suite.mockUserService.On("GetUser", userID).Return(&user.User{
		ID:         userID,
		Attributes: attrsJSON,
	}, nil)

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			userAttributeUsername: "testuser",
		},
		RuntimeData: map[string]string{},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	result := suite.executor.ValidatePrerequisites(ctx, execResp)
	assert.True(suite.T(), result)
	assert.Equal(suite.T(), mobileNumber, ctx.RuntimeData[userAttributeMobileNumber])
}

func (suite *SMSAuthExecutorTestSuite) TestValidatePrerequisites_AuthFlow_SatisfyFails() {
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(false) // Prerequisites not met
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("GetUserIDFromContext", mock.Anything).
		Return("") // No user ID in context

	ctx := &core.NodeContext{
		FlowID:      "flow-123",
		FlowType:    common.FlowTypeAuthentication,
		UserInputs:  map[string]string{}, // No inputs to resolve user ID
		RuntimeData: map[string]string{},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	result := suite.executor.ValidatePrerequisites(ctx, execResp)
	assert.False(suite.T(), result)
	assert.Equal(suite.T(), common.ExecFailure, execResp.Status)
	assert.Contains(suite.T(), execResp.FailureReason, "User ID could not be resolved")
}

func (suite *SMSAuthExecutorTestSuite) TestExecuteVerify_MissingInputs() {
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("GetRequiredInputs", mock.Anything).Return([]common.Input{
		{Identifier: userInputOTP, Type: "string", Required: true},
	}).Maybe()
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("HasRequiredInputs", mock.Anything, mock.Anything).
		Return(func(ctx *core.NodeContext, execResp *common.ExecutorResponse) bool {
			otp := ctx.UserInputs[userInputOTP]
			if otp == "" {
				execResp.Inputs = []common.Input{{Identifier: userInputOTP, Type: "string", Required: true}}
				execResp.Status = common.ExecUserInputRequired
				return false
			}
			return true
		})
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(true)

	ctx := &core.NodeContext{
		FlowID:        "flow-123",
		AppID:         "app-1",
		CurrentNodeID: "node-1",
		FlowType:      common.FlowTypeAuthentication,
		ExecutorMode:  smsOTPExecutorModeVerify,
		UserInputs:    map[string]string{}, // No OTP provided
		RuntimeData: map[string]string{
			userAttributeUserID:       testUserID,
			userAttributeMobileNumber: "+1234567890",
		},
	}

	resp, err := suite.executor.Execute(ctx)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecUserInputRequired, resp.Status)
	assert.NotEmpty(suite.T(), resp.Inputs)
}

func (suite *SMSAuthExecutorTestSuite) TestExecute_SendMode_RegistrationFlow_UserExists() {
	userID := testUserID
	mobileNumber := "+1234567890"

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("GetRequiredInputs", mock.Anything).Return([]common.Input{
		{Identifier: userInputOTP, Type: "string", Required: true},
	}).Maybe()

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(true)

	// User is found during registration - should fail
	suite.mockUserService.On("IdentifyUser", map[string]interface{}{
		userAttributeMobileNumber: mobileNumber,
	}).Return(&userID, nil)

	ctx := &core.NodeContext{
		FlowID:        "flow-123",
		AppID:         "app-1",
		CurrentNodeID: "node-1",
		FlowType:      common.FlowTypeRegistration,
		ExecutorMode:  smsOTPExecutorModeSend,
		UserInputs:    map[string]string{},
		RuntimeData: map[string]string{
			userAttributeMobileNumber: mobileNumber,
		},
		NodeProperties: map[string]interface{}{
			"senderId": "testSender",
		},
	}

	resp, err := suite.executor.Execute(ctx)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Contains(suite.T(), resp.FailureReason, "User already exists")
}

func (suite *SMSAuthExecutorTestSuite) TestExecute_SendMode_AuthFlow_Success() {
	userID := testUserID
	mobileNumber := "+1234567890"

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("GetRequiredInputs", mock.Anything).Return([]common.Input{
		{Identifier: userInputOTP, Type: "string", Required: true},
	}).Maybe()

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(true)

	// User is found during authentication
	suite.mockUserService.On("IdentifyUser", map[string]interface{}{
		userAttributeMobileNumber: mobileNumber,
	}).Return(&userID, nil)

	suite.mockOTPService.On("SendOTP", mock.MatchedBy(func(dto notifcommon.SendOTPDTO) bool {
		return dto.Recipient == mobileNumber && dto.SenderID == "testSender"
	})).Return(&notifcommon.SendOTPResultDTO{
		SessionToken: "session-token-123",
	}, nil)

	ctx := &core.NodeContext{
		FlowID:        "flow-123",
		AppID:         "app-1",
		CurrentNodeID: "node-1",
		FlowType:      common.FlowTypeAuthentication,
		ExecutorMode:  smsOTPExecutorModeSend,
		UserInputs:    map[string]string{},
		RuntimeData: map[string]string{
			userAttributeMobileNumber: mobileNumber,
		},
		NodeProperties: map[string]interface{}{
			"senderId": "testSender",
		},
	}

	resp, err := suite.executor.Execute(ctx)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	assert.Equal(suite.T(), "session-token-123", resp.RuntimeData["otpSessionToken"])
	assert.Equal(suite.T(), userID, resp.RuntimeData[userAttributeUserID])
}

func (suite *SMSAuthExecutorTestSuite) TestExecute_SendMode_RegistrationFlow_Success() {
	mobileNumber := "+1234567890"

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("GetRequiredInputs", mock.Anything).Return([]common.Input{
		{Identifier: userInputOTP, Type: "string", Required: true},
	}).Maybe()

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(true)

	// User is NOT found during registration - should proceed
	suite.mockUserService.On("IdentifyUser", map[string]interface{}{
		userAttributeMobileNumber: mobileNumber,
	}).Return(nil, &user.ErrorUserNotFound)

	suite.mockOTPService.On("SendOTP", mock.MatchedBy(func(dto notifcommon.SendOTPDTO) bool {
		return dto.Recipient == mobileNumber && dto.SenderID == "testSender"
	})).Return(&notifcommon.SendOTPResultDTO{
		SessionToken: "session-token-123",
	}, nil)

	ctx := &core.NodeContext{
		FlowID:        "flow-123",
		AppID:         "app-1",
		CurrentNodeID: "node-1",
		FlowType:      common.FlowTypeRegistration,
		ExecutorMode:  smsOTPExecutorModeSend,
		UserInputs:    map[string]string{},
		RuntimeData: map[string]string{
			userAttributeMobileNumber: mobileNumber,
		},
		NodeProperties: map[string]interface{}{
			"senderId": "testSender",
		},
	}

	resp, err := suite.executor.Execute(ctx)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	assert.Equal(suite.T(), "session-token-123", resp.RuntimeData["otpSessionToken"])
}

func (suite *SMSAuthExecutorTestSuite) TestExecute_SendMode_MissingSenderID() {
	mobileNumber := "+1234567890"

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("GetRequiredInputs", mock.Anything).Return([]common.Input{
		{Identifier: userInputOTP, Type: "string", Required: true},
	}).Maybe()

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(true)

	suite.mockUserService.On("IdentifyUser", map[string]interface{}{
		userAttributeMobileNumber: mobileNumber,
	}).Return(nil, &user.ErrorUserNotFound)

	ctx := &core.NodeContext{
		FlowID:        "flow-123",
		AppID:         "app-1",
		CurrentNodeID: "node-1",
		FlowType:      common.FlowTypeRegistration,
		ExecutorMode:  smsOTPExecutorModeSend,
		UserInputs:    map[string]string{},
		RuntimeData: map[string]string{
			userAttributeMobileNumber: mobileNumber,
		},
		NodeProperties: map[string]interface{}{}, // No senderId
	}

	_, err := suite.executor.Execute(ctx)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "message sender id is not configured")
	suite.mockUserService.AssertExpectations(suite.T())
}

func (suite *SMSAuthExecutorTestSuite) TestExecute_SendMode_OTPServiceError() {
	userID := testUserID
	mobileNumber := "+1234567890"

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("GetRequiredInputs", mock.Anything).Return([]common.Input{
		{Identifier: userInputOTP, Type: "string", Required: true},
	}).Maybe()

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(true)

	suite.mockUserService.On("IdentifyUser", map[string]interface{}{
		userAttributeMobileNumber: mobileNumber,
	}).Return(&userID, nil)

	suite.mockOTPService.On("SendOTP", mock.Anything).Return(nil,
		&serviceerror.ServiceError{
			Code:  "OTP_SEND_FAILED",
			Error: "Failed to send OTP",
		})

	ctx := &core.NodeContext{
		FlowID:        "flow-123",
		AppID:         "app-1",
		CurrentNodeID: "node-1",
		FlowType:      common.FlowTypeAuthentication,
		ExecutorMode:  smsOTPExecutorModeSend,
		UserInputs:    map[string]string{},
		RuntimeData: map[string]string{
			userAttributeMobileNumber: mobileNumber,
		},
		NodeProperties: map[string]interface{}{
			"senderId": "testSender",
		},
	}

	_, err := suite.executor.Execute(ctx)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to send OTP")
	suite.mockUserService.AssertExpectations(suite.T())
	suite.mockOTPService.AssertExpectations(suite.T())
}

func (suite *SMSAuthExecutorTestSuite) TestExecute_SendMode_MaxAttemptsReached() {
	userID := testUserID
	mobileNumber := "+1234567890"

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("GetRequiredInputs", mock.Anything).Return([]common.Input{
		{Identifier: userInputOTP, Type: "string", Required: true},
	}).Maybe()

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(true)

	suite.mockUserService.On("IdentifyUser", map[string]interface{}{
		userAttributeMobileNumber: mobileNumber,
	}).Return(&userID, nil)

	ctx := &core.NodeContext{
		FlowID:        "flow-123",
		AppID:         "app-1",
		CurrentNodeID: "node-1",
		FlowType:      common.FlowTypeAuthentication,
		ExecutorMode:  smsOTPExecutorModeSend,
		UserInputs:    map[string]string{},
		RuntimeData: map[string]string{
			userAttributeMobileNumber: mobileNumber,
			"attemptCount":            "3", // Max attempts reached
		},
		NodeProperties: map[string]interface{}{
			"senderId": "testSender",
		},
	}

	resp, err := suite.executor.Execute(ctx)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Contains(suite.T(), resp.FailureReason, "maximum OTP attempts reached")
	suite.mockUserService.AssertExpectations(suite.T())
}

func (suite *SMSAuthExecutorTestSuite) TestExecute_VerifyMode_Success_AuthFlow() {
	userID := testUserID
	mobileNumber := "+1234567890"
	sessionToken := testSessionToken

	attrs := map[string]interface{}{userAttributeMobileNumber: mobileNumber}
	attrsJSON, _ := json.Marshal(attrs)

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("GetRequiredInputs", mock.Anything).Return([]common.Input{
		{Identifier: userInputOTP, Type: "string", Required: true},
	}).Maybe()
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("HasRequiredInputs", mock.Anything, mock.Anything).
		Return(func(ctx *core.NodeContext, execResp *common.ExecutorResponse) bool {
			otp := ctx.UserInputs[userInputOTP]
			if otp == "" {
				execResp.Inputs = []common.Input{{Identifier: userInputOTP, Type: "string", Required: true}}
				execResp.Status = common.ExecUserInputRequired
				return false
			}
			return true
		})

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(true)

	ctx := &core.NodeContext{
		FlowID:        "flow-123",
		AppID:         "app-1",
		CurrentNodeID: "node-1",
		FlowType:      common.FlowTypeAuthentication,
		ExecutorMode:  smsOTPExecutorModeVerify,
		UserInputs: map[string]string{
			userInputOTP: "123456",
		},
		RuntimeData: map[string]string{
			userAttributeUserID:       userID,
			userAttributeMobileNumber: mobileNumber,
			"otpSessionToken":         sessionToken,
		},
	}

	suite.mockOTPService.On("VerifyOTP", notifcommon.VerifyOTPDTO{
		SessionToken: sessionToken,
		OTPCode:      "123456",
	}).Return(&notifcommon.VerifyOTPResultDTO{
		Status: notifcommon.OTPVerifyStatusVerified,
	}, nil)

	suite.mockUserService.On("GetUser", userID).Return(&user.User{
		ID:               userID,
		OrganizationUnit: "ou-123",
		Type:             "INTERNAL",
		Attributes:       attrsJSON,
	}, nil)

	resp, err := suite.executor.Execute(ctx)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	assert.True(suite.T(), resp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), userID, resp.AuthenticatedUser.UserID)
}

func (suite *SMSAuthExecutorTestSuite) TestExecute_VerifyMode_Success_RegistrationFlow() {
	mobileNumber := "+1234567890"
	sessionToken := testSessionToken

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("GetRequiredInputs", mock.Anything).Return([]common.Input{
		{Identifier: userInputOTP, Type: "string", Required: true},
	}).Maybe()
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("HasRequiredInputs", mock.Anything, mock.Anything).
		Return(func(ctx *core.NodeContext, execResp *common.ExecutorResponse) bool {
			otp := ctx.UserInputs[userInputOTP]
			if otp == "" {
				execResp.Inputs = []common.Input{{Identifier: userInputOTP, Type: "string", Required: true}}
				execResp.Status = common.ExecUserInputRequired
				return false
			}
			return true
		})

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(true)

	ctx := &core.NodeContext{
		FlowID:        "flow-123",
		AppID:         "app-1",
		CurrentNodeID: "node-1",
		FlowType:      common.FlowTypeRegistration,
		ExecutorMode:  smsOTPExecutorModeVerify,
		UserInputs: map[string]string{
			userInputOTP: "123456",
		},
		RuntimeData: map[string]string{
			userAttributeMobileNumber: mobileNumber,
			"otpSessionToken":         sessionToken,
		},
	}

	suite.mockOTPService.On("VerifyOTP", notifcommon.VerifyOTPDTO{
		SessionToken: sessionToken,
		OTPCode:      "123456",
	}).Return(&notifcommon.VerifyOTPResultDTO{
		Status: notifcommon.OTPVerifyStatusVerified,
	}, nil)

	resp, err := suite.executor.Execute(ctx)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	assert.False(suite.T(), resp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), mobileNumber, resp.AuthenticatedUser.Attributes[userAttributeMobileNumber])
}

func (suite *SMSAuthExecutorTestSuite) TestExecute_VerifyMode_InvalidOTP() {
	userID := testUserID
	mobileNumber := "+1234567890"
	sessionToken := testSessionToken

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("GetRequiredInputs", mock.Anything).Return([]common.Input{
		{Identifier: userInputOTP, Type: "string", Required: true},
	}).Maybe()
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("HasRequiredInputs", mock.Anything, mock.Anything).
		Return(func(ctx *core.NodeContext, execResp *common.ExecutorResponse) bool {
			otp := ctx.UserInputs[userInputOTP]
			if otp == "" {
				execResp.Inputs = []common.Input{{Identifier: userInputOTP, Type: "string", Required: true}}
				execResp.Status = common.ExecUserInputRequired
				return false
			}
			return true
		})

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(true)

	ctx := &core.NodeContext{
		FlowID:        "flow-123",
		AppID:         "app-1",
		CurrentNodeID: "node-1",
		FlowType:      common.FlowTypeAuthentication,
		ExecutorMode:  smsOTPExecutorModeVerify,
		UserInputs: map[string]string{
			userInputOTP: "wrong-otp",
		},
		RuntimeData: map[string]string{
			userAttributeUserID:       userID,
			userAttributeMobileNumber: mobileNumber,
			"otpSessionToken":         sessionToken,
		},
	}

	suite.mockOTPService.On("VerifyOTP", notifcommon.VerifyOTPDTO{
		SessionToken: sessionToken,
		OTPCode:      "wrong-otp",
	}).Return(&notifcommon.VerifyOTPResultDTO{
		Status: notifcommon.OTPVerifyStatusInvalid,
	}, nil)

	resp, err := suite.executor.Execute(ctx)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Equal(suite.T(), errorInvalidOTP, resp.FailureReason)
}

func (suite *SMSAuthExecutorTestSuite) TestExecute_VerifyMode_OTPServiceError() {
	userID := testUserID
	mobileNumber := "+1234567890"
	sessionToken := testSessionToken

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("GetRequiredInputs", mock.Anything).Return([]common.Input{
		{Identifier: userInputOTP, Type: "string", Required: true},
	}).Maybe()
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("HasRequiredInputs", mock.Anything, mock.Anything).
		Return(func(ctx *core.NodeContext, execResp *common.ExecutorResponse) bool {
			otp := ctx.UserInputs[userInputOTP]
			if otp == "" {
				execResp.Inputs = []common.Input{{Identifier: userInputOTP, Type: "string", Required: true}}
				execResp.Status = common.ExecUserInputRequired
				return false
			}
			return true
		})

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(true)

	ctx := &core.NodeContext{
		FlowID:        "flow-123",
		AppID:         "app-1",
		CurrentNodeID: "node-1",
		FlowType:      common.FlowTypeAuthentication,
		ExecutorMode:  smsOTPExecutorModeVerify,
		UserInputs: map[string]string{
			userInputOTP: "123456",
		},
		RuntimeData: map[string]string{
			userAttributeUserID:       userID,
			userAttributeMobileNumber: mobileNumber,
			"otpSessionToken":         sessionToken,
		},
	}

	suite.mockOTPService.On("VerifyOTP", mock.Anything).Return(nil,
		&serviceerror.ServiceError{
			Code:  "OTP_VERIFY_FAILED",
			Error: "Failed to verify OTP",
		})

	_, err := suite.executor.Execute(ctx)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to verify OTP")
	suite.mockOTPService.AssertExpectations(suite.T())
}

func (suite *SMSAuthExecutorTestSuite) TestExecute_VerifyMode_MissingSessionToken() {
	userID := testUserID
	mobileNumber := "+1234567890"

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("GetRequiredInputs", mock.Anything).Return([]common.Input{
		{Identifier: userInputOTP, Type: "string", Required: true},
	}).Maybe()
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("HasRequiredInputs", mock.Anything, mock.Anything).
		Return(func(ctx *core.NodeContext, execResp *common.ExecutorResponse) bool {
			otp := ctx.UserInputs[userInputOTP]
			if otp == "" {
				execResp.Inputs = []common.Input{{Identifier: userInputOTP, Type: "string", Required: true}}
				execResp.Status = common.ExecUserInputRequired
				return false
			}
			return true
		})

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(true)

	ctx := &core.NodeContext{
		FlowID:        "flow-123",
		AppID:         "app-1",
		CurrentNodeID: "node-1",
		FlowType:      common.FlowTypeAuthentication,
		ExecutorMode:  smsOTPExecutorModeVerify,
		UserInputs: map[string]string{
			userInputOTP: "123456",
		},
		RuntimeData: map[string]string{
			userAttributeUserID:       userID,
			userAttributeMobileNumber: mobileNumber,
			// Missing otpSessionToken
		},
	}

	_, err := suite.executor.Execute(ctx)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "no session token found")
}

func (suite *SMSAuthExecutorTestSuite) TestExecute_VerifyMode_UserServiceError() {
	userID := testUserID
	mobileNumber := "+1234567890"
	sessionToken := testSessionToken

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("GetRequiredInputs", mock.Anything).Return([]common.Input{
		{Identifier: userInputOTP, Type: "string", Required: true},
	}).Maybe()
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("HasRequiredInputs", mock.Anything, mock.Anything).
		Return(func(ctx *core.NodeContext, execResp *common.ExecutorResponse) bool {
			otp := ctx.UserInputs[userInputOTP]
			if otp == "" {
				execResp.Inputs = []common.Input{{Identifier: userInputOTP, Type: "string", Required: true}}
				execResp.Status = common.ExecUserInputRequired
				return false
			}
			return true
		})

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(true)

	ctx := &core.NodeContext{
		FlowID:        "flow-123",
		AppID:         "app-1",
		CurrentNodeID: "node-1",
		FlowType:      common.FlowTypeAuthentication,
		ExecutorMode:  smsOTPExecutorModeVerify,
		UserInputs: map[string]string{
			userInputOTP: "123456",
		},
		RuntimeData: map[string]string{
			userAttributeUserID:       userID,
			userAttributeMobileNumber: mobileNumber,
			"otpSessionToken":         sessionToken,
		},
	}

	suite.mockOTPService.On("VerifyOTP", notifcommon.VerifyOTPDTO{
		SessionToken: sessionToken,
		OTPCode:      "123456",
	}).Return(&notifcommon.VerifyOTPResultDTO{
		Status: notifcommon.OTPVerifyStatusVerified,
	}, nil)

	suite.mockUserService.On("GetUser", userID).Return(nil,
		&serviceerror.ServiceError{
			Code:  "USER_NOT_FOUND",
			Error: "User not found",
		})

	_, err := suite.executor.Execute(ctx)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to get user details")
	suite.mockOTPService.AssertExpectations(suite.T())
	suite.mockUserService.AssertExpectations(suite.T())
}

func (suite *SMSAuthExecutorTestSuite) TestExecute_PrerequisitesNotMet() {
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(false)
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("GetUserIDFromContext", mock.Anything).
		Return("").Maybe()

	ctx := &core.NodeContext{
		FlowID:        "flow-123",
		AppID:         "app-1",
		CurrentNodeID: "node-1",
		FlowType:      common.FlowTypeAuthentication,
		ExecutorMode:  smsOTPExecutorModeSend,
		UserInputs:    map[string]string{},
		RuntimeData:   map[string]string{},
	}

	resp, err := suite.executor.Execute(ctx)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	// Should return early without completing the flow
}

func (suite *SMSAuthExecutorTestSuite) TestInitiateOTP_AuthenticatedUser_EmptyUserID() {
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(true)
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("GetUserIDFromContext", mock.Anything).
		Return("")

	ctx := &core.NodeContext{
		FlowID:        "flow-123",
		AppID:         "app-1",
		CurrentNodeID: "node-1",
		FlowType:      common.FlowTypeAuthentication,
		ExecutorMode:  smsOTPExecutorModeSend,
		UserInputs:    map[string]string{},
		RuntimeData: map[string]string{
			userAttributeMobileNumber: "+1234567890",
		},
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
		},
		NodeProperties: map[string]interface{}{
			"senderId": "testSender",
		},
	}

	_, err := suite.executor.Execute(ctx)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "user ID is empty in the context")
}

func (suite *SMSAuthExecutorTestSuite) TestInitiateOTP_MissingMobileNumber() {
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(true)

	ctx := &core.NodeContext{
		FlowID:        "flow-123",
		AppID:         "app-1",
		CurrentNodeID: "node-1",
		FlowType:      common.FlowTypeAuthentication,
		ExecutorMode:  smsOTPExecutorModeSend,
		UserInputs:    map[string]string{},
		RuntimeData:   map[string]string{},
		NodeProperties: map[string]interface{}{
			"senderId": "testSender",
		},
	}

	_, err := suite.executor.Execute(ctx)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "mobile number not found in context")
}

func (suite *SMSAuthExecutorTestSuite) TestInitiateOTP_RegistrationFlow_IdentifyUserError() {
	mobileNumber := "+1234567890"

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(true)

	suite.mockUserService.On("IdentifyUser", map[string]interface{}{
		userAttributeMobileNumber: mobileNumber,
	}).Return(nil, &serviceerror.ServiceError{
		Code:  "SERVICE_ERROR",
		Error: "Some service error",
	})

	ctx := &core.NodeContext{
		FlowID:        "flow-123",
		AppID:         "app-1",
		CurrentNodeID: "node-1",
		FlowType:      common.FlowTypeRegistration,
		ExecutorMode:  smsOTPExecutorModeSend,
		UserInputs:    map[string]string{},
		RuntimeData: map[string]string{
			userAttributeMobileNumber: mobileNumber,
		},
		NodeProperties: map[string]interface{}{
			"senderId": "testSender",
		},
	}

	_, err := suite.executor.Execute(ctx)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to identify user during registration flow")
}

func (suite *SMSAuthExecutorTestSuite) TestInitiateOTP_AuthFlow_UserNotFound() {
	mobileNumber := "+1234567890"

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(true)

	suite.mockUserService.On("IdentifyUser", map[string]interface{}{
		userAttributeMobileNumber: mobileNumber,
	}).Return(nil, &user.ErrorUserNotFound)

	ctx := &core.NodeContext{
		FlowID:        "flow-123",
		AppID:         "app-1",
		CurrentNodeID: "node-1",
		FlowType:      common.FlowTypeAuthentication,
		ExecutorMode:  smsOTPExecutorModeSend,
		UserInputs:    map[string]string{},
		RuntimeData: map[string]string{
			userAttributeMobileNumber: mobileNumber,
		},
		NodeProperties: map[string]interface{}{
			"senderId": "testSender",
		},
	}

	resp, err := suite.executor.Execute(ctx)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
}

func (suite *SMSAuthExecutorTestSuite) TestGetUserMobileFromContext_FromUserInputs() {
	mobileNumber := "+1234567890"
	userID := testUserID

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(true)

	suite.mockUserService.On("IdentifyUser", map[string]interface{}{
		userAttributeMobileNumber: mobileNumber,
	}).Return(&userID, nil)

	suite.mockOTPService.On("SendOTP", mock.Anything).Return(&notifcommon.SendOTPResultDTO{
		SessionToken: "session-token-123",
	}, nil)

	ctx := &core.NodeContext{
		FlowID:        "flow-123",
		AppID:         "app-1",
		CurrentNodeID: "node-1",
		FlowType:      common.FlowTypeAuthentication,
		ExecutorMode:  smsOTPExecutorModeSend,
		UserInputs: map[string]string{
			userAttributeMobileNumber: mobileNumber,
		},
		RuntimeData: map[string]string{},
		NodeProperties: map[string]interface{}{
			"senderId": "testSender",
		},
	}

	resp, err := suite.executor.Execute(ctx)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
}

func (suite *SMSAuthExecutorTestSuite) TestSatisfyPrerequisites_ResolveUserIDError() {
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(false)
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("GetUserIDFromContext", mock.Anything).
		Return("")

	suite.mockUserService.On("IdentifyUser", map[string]interface{}{
		userAttributeEmail: "test@example.com",
	}).Return(nil, &serviceerror.ServiceError{
		Code:  "SERVICE_ERROR",
		Error: "Service error",
	})

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			userAttributeEmail: "test@example.com",
		},
		RuntimeData: map[string]string{},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	result := suite.executor.ValidatePrerequisites(ctx, execResp)
	assert.False(suite.T(), result)
	assert.Equal(suite.T(), common.ExecFailure, execResp.Status)
	assert.Contains(suite.T(), execResp.FailureReason, "Failed to resolve user ID")
}

func (suite *SMSAuthExecutorTestSuite) TestSatisfyPrerequisites_GetUserMobileNumberError() {
	userID := testUserID

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(false).Once()
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("GetUserIDFromContext", mock.Anything).
		Return("")

	suite.mockUserService.On("IdentifyUser", map[string]interface{}{
		userAttributeUsername: "testuser",
	}).Return(&userID, nil)

	suite.mockUserService.On("GetUser", userID).Return(nil, &serviceerror.ServiceError{
		Code:  "USER_NOT_FOUND",
		Error: "User not found",
	})

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			userAttributeUsername: "testuser",
		},
		RuntimeData: map[string]string{},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	result := suite.executor.ValidatePrerequisites(ctx, execResp)
	assert.False(suite.T(), result)
	assert.Equal(suite.T(), common.ExecFailure, execResp.Status)
	assert.Contains(suite.T(), execResp.FailureReason, "Failed to retrieve mobile number")
}

func (suite *SMSAuthExecutorTestSuite) TestGetUserMobileNumber_InvalidJSON() {
	userID := testUserID

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(false).Once()
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("GetUserIDFromContext", mock.Anything).
		Return("")

	suite.mockUserService.On("IdentifyUser", map[string]interface{}{
		userAttributeUsername: "testuser",
	}).Return(&userID, nil)

	suite.mockUserService.On("GetUser", userID).Return(&user.User{
		ID:         userID,
		Attributes: []byte("invalid json"),
	}, nil)

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			userAttributeUsername: "testuser",
		},
		RuntimeData: map[string]string{},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	result := suite.executor.ValidatePrerequisites(ctx, execResp)
	assert.False(suite.T(), result)
	assert.Equal(suite.T(), common.ExecFailure, execResp.Status)
}

func (suite *SMSAuthExecutorTestSuite) TestGetUserMobileNumber_NotAuthenticatedNoMobile() {
	userID := testUserID

	suite.setupMobileNumberErrorMocks(userAttributeUsername, "testuser", userID,
		map[string]interface{}{})

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			userAttributeUsername: "testuser",
		},
		RuntimeData: map[string]string{},
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: false,
		},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	result := suite.executor.ValidatePrerequisites(ctx, execResp)
	assert.False(suite.T(), result)
	assert.Equal(suite.T(), common.ExecFailure, execResp.Status)
	assert.Contains(suite.T(), execResp.FailureReason, "Mobile number not found in user attributes")
}

func (suite *SMSAuthExecutorTestSuite) TestGetUserMobileNumber_AuthenticatedUserWithContextMobile() {
	userID := testUserID
	mobileNumber := "+1234567890"
	attrs := map[string]interface{}{}
	attrsJSON, _ := json.Marshal(attrs)

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(false).Once()
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(true).Once()
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("GetUserIDFromContext", mock.Anything).
		Return("")

	// When mobile number is in UserInputs, resolveUserID will try to identify user by mobile first
	suite.mockUserService.On("IdentifyUser", map[string]interface{}{
		userAttributeMobileNumber: mobileNumber,
	}).Return(&userID, nil)

	suite.mockUserService.On("GetUser", userID).Return(&user.User{
		ID:         userID,
		Attributes: attrsJSON,
	}, nil)

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			userAttributeMobileNumber: mobileNumber,
		},
		RuntimeData: map[string]string{},
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
		},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	result := suite.executor.ValidatePrerequisites(ctx, execResp)
	assert.True(suite.T(), result)
	assert.Equal(suite.T(), mobileNumber, ctx.RuntimeData[userAttributeMobileNumber])
}

func (suite *SMSAuthExecutorTestSuite) TestResolveUserID_FromRuntimeDataMobile() {
	userID := testUserID
	mobileNumber := "+1234567890"

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(false).Once()
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(true).Once()
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("GetUserIDFromContext", mock.Anything).
		Return("")

	attrs := map[string]interface{}{userAttributeMobileNumber: mobileNumber}
	attrsJSON, _ := json.Marshal(attrs)

	suite.mockUserService.On("IdentifyUser", map[string]interface{}{
		userAttributeMobileNumber: mobileNumber,
	}).Return(&userID, nil)

	suite.mockUserService.On("GetUser", userID).Return(&user.User{
		ID:         userID,
		Attributes: attrsJSON,
	}, nil)

	ctx := &core.NodeContext{
		FlowID:     "flow-123",
		FlowType:   common.FlowTypeAuthentication,
		UserInputs: map[string]string{},
		RuntimeData: map[string]string{
			userAttributeMobileNumber: mobileNumber,
		},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	result := suite.executor.ValidatePrerequisites(ctx, execResp)
	assert.True(suite.T(), result)
}

func (suite *SMSAuthExecutorTestSuite) TestResolveUserID_FromContext() {
	userID := testUserID
	mobileNumber := "+1234567890"

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(false).Once()
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(true).Once()
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("GetUserIDFromContext", mock.Anything).
		Return(userID)

	attrs := map[string]interface{}{userAttributeMobileNumber: mobileNumber}
	attrsJSON, _ := json.Marshal(attrs)

	suite.mockUserService.On("GetUser", userID).Return(&user.User{
		ID:         userID,
		Attributes: attrsJSON,
	}, nil)

	ctx := &core.NodeContext{
		FlowID:      "flow-123",
		FlowType:    common.FlowTypeAuthentication,
		UserInputs:  map[string]string{},
		RuntimeData: nil, // nil RuntimeData to test initialization
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	result := suite.executor.ValidatePrerequisites(ctx, execResp)
	assert.True(suite.T(), result)
	assert.NotNil(suite.T(), ctx.RuntimeData)
	assert.Equal(suite.T(), userID, ctx.RuntimeData[userAttributeUserID])
}

func (suite *SMSAuthExecutorTestSuite) TestResolveUserID_FromUsername() {
	userID := testUserID
	mobileNumber := "+1234567890"
	username := "testuser"

	suite.setupUserResolutionMocks(userAttributeUsername, username, userID,
		map[string]interface{}{userAttributeMobileNumber: mobileNumber})

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			userAttributeUsername: username,
		},
		RuntimeData: map[string]string{},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	result := suite.executor.ValidatePrerequisites(ctx, execResp)
	assert.True(suite.T(), result)
	assert.Equal(suite.T(), mobileNumber, ctx.RuntimeData[userAttributeMobileNumber])
}

func (suite *SMSAuthExecutorTestSuite) TestResolveUserID_FromEmail() {
	userID := testUserID
	mobileNumber := "+1234567890"
	email := "test@example.com"

	suite.setupUserResolutionMocks(userAttributeEmail, email, userID,
		map[string]interface{}{userAttributeMobileNumber: mobileNumber})

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			userAttributeEmail: email,
		},
		RuntimeData: map[string]string{},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	result := suite.executor.ValidatePrerequisites(ctx, execResp)
	assert.True(suite.T(), result)
	assert.Equal(suite.T(), mobileNumber, ctx.RuntimeData[userAttributeMobileNumber])
}

func (suite *SMSAuthExecutorTestSuite) TestValidateAttempts_ParseError() {
	mobileNumber := "+1234567890"
	userID := testUserID

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(true)

	suite.mockUserService.On("IdentifyUser", map[string]interface{}{
		userAttributeMobileNumber: mobileNumber,
	}).Return(&userID, nil)

	ctx := &core.NodeContext{
		FlowID:        "flow-123",
		AppID:         "app-1",
		CurrentNodeID: "node-1",
		FlowType:      common.FlowTypeAuthentication,
		ExecutorMode:  smsOTPExecutorModeSend,
		UserInputs:    map[string]string{},
		RuntimeData: map[string]string{
			userAttributeMobileNumber: mobileNumber,
			"attemptCount":            "invalid",
		},
		NodeProperties: map[string]interface{}{
			"senderId": "testSender",
		},
	}

	_, err := suite.executor.Execute(ctx)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to parse attempt count")
}

func (suite *SMSAuthExecutorTestSuite) TestGetUserMobileNumber_AuthenticatedUser_ContextError() {
	userID := testUserID

	suite.setupMobileNumberErrorMocks(userAttributeUsername, "testuser", userID,
		map[string]interface{}{})

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			userAttributeUsername: "testuser",
		},
		RuntimeData: map[string]string{},
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
		},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	result := suite.executor.ValidatePrerequisites(ctx, execResp)
	assert.False(suite.T(), result)
	assert.Equal(suite.T(), common.ExecFailure, execResp.Status)
	assert.Contains(suite.T(), execResp.FailureReason, "Mobile number not found in user attributes or context")
}

func (suite *SMSAuthExecutorTestSuite) TestGenerateAndSendOTP_EmptyNodeProperties() {
	mobileNumber := "+1234567890"
	userID := testUserID

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(true)

	suite.mockUserService.On("IdentifyUser", map[string]interface{}{
		userAttributeMobileNumber: mobileNumber,
	}).Return(&userID, nil)

	ctx := &core.NodeContext{
		FlowID:        "flow-123",
		AppID:         "app-1",
		CurrentNodeID: "node-1",
		FlowType:      common.FlowTypeAuthentication,
		ExecutorMode:  smsOTPExecutorModeSend,
		UserInputs:    map[string]string{},
		RuntimeData: map[string]string{
			userAttributeMobileNumber: mobileNumber,
		},
		NodeProperties: map[string]interface{}{}, // Empty node properties
	}

	_, err := suite.executor.Execute(ctx)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "message sender id is not configured")
}

func (suite *SMSAuthExecutorTestSuite) TestValidateOTP_EmptyOTP() {
	userID := testUserID
	sessionToken := testSessionToken

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(true)
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("HasRequiredInputs", mock.Anything, mock.Anything).
		Return(true)

	ctx := &core.NodeContext{
		FlowID:        "flow-123",
		AppID:         "app-1",
		CurrentNodeID: "node-1",
		FlowType:      common.FlowTypeAuthentication,
		ExecutorMode:  smsOTPExecutorModeVerify,
		UserInputs: map[string]string{
			userInputOTP: "", // Empty OTP
		},
		RuntimeData: map[string]string{
			userAttributeUserID:       userID,
			userAttributeMobileNumber: "+1234567890",
			"otpSessionToken":         sessionToken,
		},
	}

	resp, err := suite.executor.Execute(ctx)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Equal(suite.T(), errorInvalidOTP, resp.FailureReason)
}

func (suite *SMSAuthExecutorTestSuite) TestGetAuthenticatedUser_EmptyUserID() {
	sessionToken := testSessionToken

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(true)
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("HasRequiredInputs", mock.Anything, mock.Anything).
		Return(true)

	suite.mockOTPService.On("VerifyOTP", notifcommon.VerifyOTPDTO{
		SessionToken: sessionToken,
		OTPCode:      "123456",
	}).Return(&notifcommon.VerifyOTPResultDTO{
		Status: notifcommon.OTPVerifyStatusVerified,
	}, nil)

	ctx := &core.NodeContext{
		FlowID:        "flow-123",
		AppID:         "app-1",
		CurrentNodeID: "node-1",
		FlowType:      common.FlowTypeAuthentication,
		ExecutorMode:  smsOTPExecutorModeVerify,
		UserInputs: map[string]string{
			userInputOTP: "123456",
		},
		RuntimeData: map[string]string{
			userAttributeMobileNumber: "+1234567890",
			"otpSessionToken":         sessionToken,
		},
	}

	_, err := suite.executor.Execute(ctx)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "user ID is empty")
}

func (suite *SMSAuthExecutorTestSuite) TestGetAuthenticatedUser_InvalidUserAttributes() {
	userID := testUserID
	mobileNumber := "+1234567890"
	sessionToken := testSessionToken

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(true)
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("HasRequiredInputs", mock.Anything, mock.Anything).
		Return(true)

	suite.mockOTPService.On("VerifyOTP", notifcommon.VerifyOTPDTO{
		SessionToken: sessionToken,
		OTPCode:      "123456",
	}).Return(&notifcommon.VerifyOTPResultDTO{
		Status: notifcommon.OTPVerifyStatusVerified,
	}, nil)

	suite.mockUserService.On("GetUser", userID).Return(&user.User{
		ID:         userID,
		Attributes: []byte("invalid json"),
	}, nil)

	ctx := &core.NodeContext{
		FlowID:        "flow-123",
		AppID:         "app-1",
		CurrentNodeID: "node-1",
		FlowType:      common.FlowTypeAuthentication,
		ExecutorMode:  smsOTPExecutorModeVerify,
		UserInputs: map[string]string{
			userInputOTP: "123456",
		},
		RuntimeData: map[string]string{
			userAttributeUserID:       userID,
			userAttributeMobileNumber: mobileNumber,
			"otpSessionToken":         sessionToken,
		},
	}

	_, err := suite.executor.Execute(ctx)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to unmarshal user attributes")
}

// Helper function to setup mocks for user resolution by attribute tests
func (suite *SMSAuthExecutorTestSuite) setupUserResolutionMocks(
	identifyAttr string, identifyValue string, userID string, userAttrs map[string]interface{}) {
	attrsJSON, _ := json.Marshal(userAttrs)

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(false).Once()
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(true).Once()
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("GetUserIDFromContext", mock.Anything).
		Return("")

	suite.mockUserService.On("IdentifyUser", map[string]interface{}{
		identifyAttr: identifyValue,
	}).Return(&userID, nil)

	suite.mockUserService.On("GetUser", userID).Return(&user.User{
		ID:         userID,
		Attributes: attrsJSON,
	}, nil)
}

// Helper function to setup mocks for mobile number validation error tests
func (suite *SMSAuthExecutorTestSuite) setupMobileNumberErrorMocks(
	identifyAttr string, identifyValue string, userID string, userAttrs map[string]interface{}) {
	attrsJSON, _ := json.Marshal(userAttrs)

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(false).Once()
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("GetUserIDFromContext", mock.Anything).
		Return("")

	suite.mockUserService.On("IdentifyUser", map[string]interface{}{
		identifyAttr: identifyValue,
	}).Return(&userID, nil)

	suite.mockUserService.On("GetUser", userID).Return(&user.User{
		ID:         userID,
		Attributes: attrsJSON,
	}, nil)
}
