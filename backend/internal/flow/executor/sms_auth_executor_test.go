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

	flowcm "github.com/asgardeo/thunder/internal/flow/common"
	flowcore "github.com/asgardeo/thunder/internal/flow/core"
	notifcommon "github.com/asgardeo/thunder/internal/notification/common"
	"github.com/asgardeo/thunder/internal/observability/event"
	"github.com/asgardeo/thunder/internal/user"
	"github.com/asgardeo/thunder/tests/mocks/flow/coremock"
	"github.com/asgardeo/thunder/tests/mocks/notification/notificationmock"
	"github.com/asgardeo/thunder/tests/mocks/observabilitymock"
	"github.com/asgardeo/thunder/tests/mocks/usermock"
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

	defaultInputs := []flowcm.InputData{
		{
			Name:     userInputOTP,
			Type:     "string",
			Required: true,
		},
	}
	prerequisites := []flowcm.InputData{
		{
			Name:     userAttributeMobileNumber,
			Type:     "string",
			Required: true,
		},
	}

	// Mock identifying executor
	identifyingMock := createMockIdentifyingExecutor(suite.T())
	suite.mockFlowFactory.On("CreateExecutor", ExecutorNameIdentifying, flowcm.ExecutorTypeUtility,
		mock.Anything, mock.Anything).Return(identifyingMock).Maybe()

	// Mock base executor
	mockExec := coremock.NewExecutorInterfaceMock(suite.T())
	mockExec.On("GetName").Return(ExecutorNameSMSAuth).Maybe()
	mockExec.On("GetType").Return(flowcm.ExecutorTypeAuthentication).Maybe()
	mockExec.On("GetDefaultExecutorInputs").Return(defaultInputs).Maybe()
	mockExec.On("GetPrerequisites").Return(prerequisites).Maybe()
	mockExec.On("ValidatePrerequisites", mock.Anything, mock.Anything).Return(true).Maybe()
	mockExec.On("CheckInputData", mock.Anything, mock.Anything).Return(true).Maybe()

	suite.mockFlowFactory.On("CreateExecutor", ExecutorNameSMSAuth, flowcm.ExecutorTypeAuthentication,
		defaultInputs, prerequisites).Return(mockExec)

	suite.executor = newSMSOTPAuthExecutor(suite.mockFlowFactory, suite.mockUserService,
		suite.mockOTPService, suite.mockObservability)
	// Inject the mock base executor
	suite.executor.ExecutorInterface = mockExec
}

func (suite *SMSAuthExecutorTestSuite) TestExecute_Observability_Success() {
	// Enable observability
	suite.mockObservability.ExpectedCalls = nil
	suite.mockObservability.On("IsEnabled").Return(true)

	userID := testUserID
	mobileNumber := "+1234567890"
	sessionToken := "session-123"

	// Mock dependencies for validation flow
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("CheckInputData", mock.Anything, mock.Anything).
		Return(false)

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(true)

	ctx := &flowcore.NodeContext{
		FlowID:        "flow-123",
		AppID:         "app-1",
		CurrentNodeID: "node-1",
		FlowType:      flowcm.FlowTypeAuthentication,
		UserInputData: map[string]string{
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

	suite.mockUserService.On("GetUser", mock.Anything, userID).Return(userObj, nil)

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
	assert.Equal(suite.T(), flowcm.ExecComplete, resp.Status)

	suite.mockObservability.AssertExpectations(suite.T())
}

func (suite *SMSAuthExecutorTestSuite) TestExecute_Observability_InitiateOTP() {
	// Enable observability
	suite.mockObservability.ExpectedCalls = nil
	suite.mockObservability.On("IsEnabled").Return(true)

	userID := testUserID
	mobileNumber := "+1234567890"

	// Mock CheckInputData to return TRUE (Inputs missing -> Initiate OTP)
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("CheckInputData", mock.Anything, mock.Anything).
		Return(true)

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(true)

	// Expect IdentifyUser to be called because user is not authenticated in context
	suite.mockUserService.On("IdentifyUser", mock.Anything, map[string]interface{}{
		userAttributeMobileNumber: mobileNumber,
	}).Return(&userID, nil)

	ctx := &flowcore.NodeContext{
		FlowID:        "flow-123",
		AppID:         "app-1",
		CurrentNodeID: "node-1",
		FlowType:      flowcm.FlowTypeAuthentication,
		UserInputData: map[string]string{},
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

	// Expect UserInputRequired event
	suite.mockObservability.On("PublishEvent", mock.MatchedBy(func(evt *event.Event) bool {
		return evt.Type == string(event.EventTypeFlowUserInputRequired) &&
			evt.Status == event.StatusPending &&
			evt.Data[event.DataKey.UserID] == userID
	})).Return()

	resp, err := suite.executor.Execute(ctx)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), flowcm.ExecUserInputRequired, resp.Status)
	suite.mockObservability.AssertExpectations(suite.T())
}

func (suite *SMSAuthExecutorTestSuite) TestExecute_Observability_Failure() {
	// Enable observability
	suite.mockObservability.ExpectedCalls = nil
	suite.mockObservability.On("IsEnabled").Return(true)

	userID := testUserID
	sessionToken := "session-123"

	// Mock CheckInputData to return FALSE (Process Response)
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).ExpectedCalls = nil
	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("CheckInputData", mock.Anything, mock.Anything).
		Return(false)

	suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock).
		On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(true)

	ctx := &flowcore.NodeContext{
		FlowID:        "flow-123",
		AppID:         "app-1",
		CurrentNodeID: "node-1",
		FlowType:      flowcm.FlowTypeAuthentication,
		UserInputData: map[string]string{
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
	assert.Equal(suite.T(), flowcm.ExecFailure, resp.Status)

	suite.mockObservability.AssertExpectations(suite.T())
}
