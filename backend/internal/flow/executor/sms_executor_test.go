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
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	notifcm "github.com/asgardeo/thunder/internal/notification/common"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/tests/mocks/flow/coremock"
	"github.com/asgardeo/thunder/tests/mocks/notification/notificationmock"
)

type SMSExecutorTestSuite struct {
	suite.Suite
	mockFlowFactory  *coremock.FlowFactoryInterfaceMock
	mockSMSSenderSvc *notificationmock.NotificationSenderServiceInterfaceMock
	executor         *smsExecutor
}

func (suite *SMSExecutorTestSuite) SetupTest() {
	suite.mockFlowFactory = coremock.NewFlowFactoryInterfaceMock(suite.T())
	mockBaseExecutor := coremock.NewExecutorInterfaceMock(suite.T())
	suite.mockSMSSenderSvc = notificationmock.NewNotificationSenderServiceInterfaceMock(suite.T())

	suite.mockFlowFactory.On("CreateExecutor",
		ExecutorNameSMSExecutor,
		common.ExecutorTypeUtility,
		[]common.Input{},
		[]common.Input{
			{Identifier: userAttributeMobileNumber, Type: common.InputTypePhone, Required: true},
		},
	).Return(mockBaseExecutor)

	suite.executor = newSMSExecutor(suite.mockFlowFactory, suite.mockSMSSenderSvc)
}

func (suite *SMSExecutorTestSuite) TestExecute_SendMode_Success() {
	ctx := &core.NodeContext{
		FlowID:       "test-flow-id",
		ExecutorMode: ExecutorModeSend,
		UserInputs: map[string]string{
			userAttributeMobileNumber: "+94714627887",
		},
		RuntimeData: make(map[string]string),
		NodeProperties: map[string]interface{}{
			propertyKeyNotificationSenderID: "sender-uuid-001",
		},
	}

	suite.mockSMSSenderSvc.On("Send",
		mock.Anything, mock.Anything, "sender-uuid-001",
		notifcm.NotificationData{Recipient: "+94714627887", Body: smsDefaultMessage},
	).Return(nil)

	resp, err := suite.executor.Execute(ctx)

	suite.NoError(err)
	suite.Equal(common.ExecComplete, resp.Status)
	suite.Equal(dataValueTrue, resp.AdditionalData[common.DataSMSSent])
}

func (suite *SMSExecutorTestSuite) TestExecute_SendMode_RecipientFromRuntimeData() {
	ctx := &core.NodeContext{
		FlowID:       "test-flow-id",
		ExecutorMode: ExecutorModeSend,
		UserInputs:   make(map[string]string),
		RuntimeData: map[string]string{
			userAttributeMobileNumber: "+94714627887",
		},
		NodeProperties: map[string]interface{}{
			propertyKeyNotificationSenderID: "sender-uuid-001",
		},
	}

	suite.mockSMSSenderSvc.On("Send",
		mock.Anything, mock.Anything, "sender-uuid-001",
		notifcm.NotificationData{Recipient: "+94714627887", Body: smsDefaultMessage},
	).Return(nil)

	resp, err := suite.executor.Execute(ctx)

	suite.NoError(err)
	suite.Equal(common.ExecComplete, resp.Status)
	suite.Equal(dataValueTrue, resp.AdditionalData[common.DataSMSSent])
}

func (suite *SMSExecutorTestSuite) TestExecute_SendMode_UserInputOverridesRuntimeData() {
	ctx := &core.NodeContext{
		FlowID:       "test-flow-id",
		ExecutorMode: ExecutorModeSend,
		UserInputs: map[string]string{
			userAttributeMobileNumber: "+94714627887",
		},
		RuntimeData: map[string]string{
			userAttributeMobileNumber: "+94771111111",
		},
		NodeProperties: map[string]interface{}{
			propertyKeyNotificationSenderID: "sender-uuid-001",
		},
	}

	suite.mockSMSSenderSvc.On("Send",
		mock.Anything, mock.Anything, "sender-uuid-001",
		notifcm.NotificationData{Recipient: "+94714627887", Body: smsDefaultMessage},
	).Return(nil)

	resp, err := suite.executor.Execute(ctx)

	suite.NoError(err)
	suite.Equal(common.ExecComplete, resp.Status)
	suite.Equal(dataValueTrue, resp.AdditionalData[common.DataSMSSent])
}

func (suite *SMSExecutorTestSuite) TestExecute_SendMode_CustomPhoneAttribute() {
	ctx := &core.NodeContext{
		FlowID:       "test-flow-id",
		ExecutorMode: ExecutorModeSend,
		UserInputs: map[string]string{
			"phoneNumber": "+94714627887",
		},
		RuntimeData: make(map[string]string),
		NodeProperties: map[string]interface{}{
			propertyKeyNotificationSenderID: "sender-uuid-001",
			propertyKeyPhoneAttribute:       "phoneNumber",
		},
	}

	suite.mockSMSSenderSvc.On("Send",
		mock.Anything, mock.Anything, "sender-uuid-001",
		notifcm.NotificationData{Recipient: "+94714627887", Body: smsDefaultMessage},
	).Return(nil)

	resp, err := suite.executor.Execute(ctx)

	suite.NoError(err)
	suite.Equal(common.ExecComplete, resp.Status)
	suite.Equal(dataValueTrue, resp.AdditionalData[common.DataSMSSent])
}

func (suite *SMSExecutorTestSuite) TestExecute_SendMode_MissingRecipient() {
	ctx := &core.NodeContext{
		FlowID:       "test-flow-id",
		ExecutorMode: ExecutorModeSend,
		UserInputs:   make(map[string]string),
		RuntimeData:  make(map[string]string),
		NodeProperties: map[string]interface{}{
			propertyKeyNotificationSenderID: "sender-uuid-001",
		},
	}

	resp, err := suite.executor.Execute(ctx)

	suite.NoError(err)
	suite.Equal(common.ExecFailure, resp.Status)
	suite.Equal("SMS recipient is required", resp.FailureReason)
	suite.mockSMSSenderSvc.AssertNotCalled(suite.T(), "Send",
		mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func (suite *SMSExecutorTestSuite) TestExecute_SendMode_MissingSenderID() {
	ctx := &core.NodeContext{
		FlowID:       "test-flow-id",
		ExecutorMode: ExecutorModeSend,
		UserInputs: map[string]string{
			userAttributeMobileNumber: "+94714627887",
		},
		RuntimeData:    make(map[string]string),
		NodeProperties: map[string]interface{}{},
	}

	resp, err := suite.executor.Execute(ctx)

	suite.Error(err)
	suite.Nil(resp)
	suite.Contains(err.Error(), "senderId is not configured")
	suite.mockSMSSenderSvc.AssertNotCalled(suite.T(), "Send",
		mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func (suite *SMSExecutorTestSuite) TestExecute_SendMode_InvalidSenderIDType() {
	ctx := &core.NodeContext{
		FlowID:       "test-flow-id",
		ExecutorMode: ExecutorModeSend,
		UserInputs: map[string]string{
			userAttributeMobileNumber: "+94714627887",
		},
		RuntimeData: make(map[string]string),
		NodeProperties: map[string]interface{}{
			propertyKeyNotificationSenderID: 123,
		},
	}

	resp, err := suite.executor.Execute(ctx)

	suite.Error(err)
	suite.Nil(resp)
	suite.Contains(err.Error(), "senderId is not configured")
}

func (suite *SMSExecutorTestSuite) TestExecute_SendMode_NilSMSSenderService_ReturnsError() {
	mockBaseExecutor := coremock.NewExecutorInterfaceMock(suite.T())
	mockFactory := coremock.NewFlowFactoryInterfaceMock(suite.T())
	mockFactory.On("CreateExecutor",
		ExecutorNameSMSExecutor,
		common.ExecutorTypeUtility,
		[]common.Input{},
		[]common.Input{
			{Identifier: userAttributeMobileNumber, Type: common.InputTypePhone, Required: true},
		},
	).Return(mockBaseExecutor)

	noServiceExecutor := newSMSExecutor(mockFactory, nil)

	ctx := &core.NodeContext{
		FlowID:       "test-flow-id",
		ExecutorMode: ExecutorModeSend,
		UserInputs: map[string]string{
			userAttributeMobileNumber: "+94714627887",
		},
		RuntimeData: make(map[string]string),
		NodeProperties: map[string]interface{}{
			propertyKeyNotificationSenderID: "sender-uuid-001",
		},
	}

	resp, err := noServiceExecutor.Execute(ctx)

	suite.Error(err)
	suite.Nil(resp)
	suite.EqualError(err, "notification sender service is not configured")
}

func (suite *SMSExecutorTestSuite) TestExecute_SendMode_ClientError() {
	ctx := &core.NodeContext{
		FlowID:       "test-flow-id",
		ExecutorMode: ExecutorModeSend,
		UserInputs: map[string]string{
			userAttributeMobileNumber: "+94714627887",
		},
		RuntimeData: make(map[string]string),
		NodeProperties: map[string]interface{}{
			propertyKeyNotificationSenderID: "sender-uuid-001",
		},
	}

	clientErr := &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "MNS-1001",
		Error:            "Sender not found",
		ErrorDescription: "The requested notification sender could not be found",
	}
	suite.mockSMSSenderSvc.On("Send", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(clientErr)

	resp, err := suite.executor.Execute(ctx)

	suite.NoError(err)
	suite.Equal(common.ExecFailure, resp.Status)
	suite.Equal("The requested notification sender could not be found", resp.FailureReason)
}

func (suite *SMSExecutorTestSuite) TestExecute_SendMode_ServerError() {
	ctx := &core.NodeContext{
		FlowID:       "test-flow-id",
		ExecutorMode: ExecutorModeSend,
		UserInputs: map[string]string{
			userAttributeMobileNumber: "+94714627887",
		},
		RuntimeData: make(map[string]string),
		NodeProperties: map[string]interface{}{
			propertyKeyNotificationSenderID: "sender-uuid-001",
		},
	}

	serverErr := &serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "MNS-5000",
		ErrorDescription: "internal server error",
	}
	suite.mockSMSSenderSvc.On("Send", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(serverErr)

	resp, err := suite.executor.Execute(ctx)

	suite.Error(err)
	suite.Nil(resp)
	suite.Contains(err.Error(), "SMS send failed")
}

func TestSMSExecutorSuite(t *testing.T) {
	suite.Run(t, new(SMSExecutorTestSuite))
}
