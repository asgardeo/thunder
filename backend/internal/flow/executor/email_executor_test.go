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
	"github.com/asgardeo/thunder/internal/system/email"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/template"
	"github.com/asgardeo/thunder/tests/mocks/emailmock"
	"github.com/asgardeo/thunder/tests/mocks/flow/coremock"
	"github.com/asgardeo/thunder/tests/mocks/templatemock"
)

type EmailExecutorTestSuite struct {
	suite.Suite
	mockFlowFactory     *coremock.FlowFactoryInterfaceMock
	mockEmailClient     *emailmock.EmailClientInterfaceMock
	mockTemplateService *templatemock.TemplateServiceInterfaceMock
	executor            *emailExecutor
}

func (suite *EmailExecutorTestSuite) SetupTest() {
	suite.mockFlowFactory = coremock.NewFlowFactoryInterfaceMock(suite.T())
	mockBaseExecutor := coremock.NewExecutorInterfaceMock(suite.T())
	suite.mockEmailClient = emailmock.NewEmailClientInterfaceMock(suite.T())
	suite.mockTemplateService = templatemock.NewTemplateServiceInterfaceMock(suite.T())

	suite.mockFlowFactory.On("CreateExecutor",
		ExecutorNameEmailExecutor,
		common.ExecutorTypeUtility,
		[]common.Input{},
		[]common.Input{
			{Identifier: userAttributeEmail, Type: common.InputTypeText, Required: true},
		},
	).Return(mockBaseExecutor)

	suite.executor = newEmailExecutor(suite.mockFlowFactory, suite.mockEmailClient, suite.mockTemplateService)
}

func (suite *EmailExecutorTestSuite) TestExecute_SendMode_UserInviteTemplate_Success() {
	ctx := &core.NodeContext{
		ExecutionID:  "test-flow-id",
		FlowType:     common.FlowTypeUserOnboarding,
		ExecutorMode: ExecutorModeSend,
		UserInputs: map[string]string{
			"email": "user@example.com",
		},
		RuntimeData: map[string]string{
			common.RuntimeKeyInviteLink: "https://localhost:5190/gate/invite?executionId=test&inviteToken=abc",
		},
		NodeProperties: map[string]interface{}{
			"emailTemplate": "USER_INVITE",
		},
	}

	suite.mockTemplateService.On("Render",
		mock.Anything,
		template.ScenarioUserInvite,
		template.TemplateTypeEmail,
		template.TemplateData{
			"inviteLink": "https://localhost:5190/gate/invite?executionId=test&inviteToken=abc",
			"appName":    "",
		},
	).Return(&template.RenderedTemplate{
		Subject: "You're Invited to Register",
		Body:    "<html><body>Complete Registration</body></html>",
		IsHTML:  true,
	}, nil)

	var sentEmail email.EmailData
	suite.mockEmailClient.On("Send", mock.Anything).Run(func(args mock.Arguments) {
		sentEmail = args.Get(0).(email.EmailData)
	}).Return(nil)

	resp, err := suite.executor.Execute(ctx)

	suite.NoError(err)
	suite.Equal(common.ExecComplete, resp.Status)
	suite.Equal(dataValueTrue, resp.AdditionalData[common.DataEmailSent])
	suite.Equal([]string{"user@example.com"}, sentEmail.To)
	suite.Equal("You're Invited to Register", sentEmail.Subject)
	suite.True(sentEmail.IsHTML)
	suite.Contains(sentEmail.Body, "Complete Registration")
}

func (suite *EmailExecutorTestSuite) TestExecute_SendMode_SelfRegistration_InviteLinkNotExposed() {
	ctx := &core.NodeContext{
		ExecutionID:  "test-flow-id",
		FlowType:     common.FlowTypeRegistration,
		ExecutorMode: ExecutorModeSend,
		UserInputs: map[string]string{
			"email": "user@example.com",
		},
		RuntimeData: map[string]string{
			common.RuntimeKeyInviteLink: "https://localhost:5190/gate/invite?executionId=test&inviteToken=abc",
		},
		NodeProperties: map[string]interface{}{
			"emailTemplate": "SELF_REGISTRATION",
		},
	}

	suite.mockTemplateService.On("Render",
		mock.Anything,
		template.ScenarioSelfRegistration,
		template.TemplateTypeEmail,
		template.TemplateData{
			"inviteLink": "https://localhost:5190/gate/invite?executionId=test&inviteToken=abc",
			"appName":    "",
		},
	).Return(&template.RenderedTemplate{
		Subject: "Complete Your Registration",
		Body:    "<html><body>Click to register</body></html>",
		IsHTML:  true,
	}, nil)

	suite.mockEmailClient.On("Send", mock.Anything).Return(nil)

	resp, err := suite.executor.Execute(ctx)

	suite.NoError(err)
	suite.Equal(common.ExecComplete, resp.Status)
	suite.Equal(dataValueTrue, resp.AdditionalData[common.DataEmailSent])
	// For SELF_REGISTRATION, invite link must NOT be exposed in AdditionalData
	suite.Empty(resp.AdditionalData[common.DataInviteLink])
}

func (suite *EmailExecutorTestSuite) TestExecute_SendMode_UsesUserInputOverRuntimeRecipient() {
	ctx := &core.NodeContext{
		ExecutionID:  "test-flow-id",
		ExecutorMode: ExecutorModeSend,
		UserInputs: map[string]string{
			"email": "user@example.com",
		},
		RuntimeData: map[string]string{
			"email":                     "runtime@example.com",
			common.RuntimeKeyInviteLink: "https://localhost:5190/gate/invite?executionId=test&inviteToken=abc",
		},
		NodeProperties: map[string]interface{}{
			"emailTemplate": "USER_INVITE",
		},
	}

	suite.mockTemplateService.On("Render",
		mock.Anything,
		template.ScenarioUserInvite,
		template.TemplateTypeEmail,
		template.TemplateData{
			"inviteLink": "https://localhost:5190/gate/invite?executionId=test&inviteToken=abc",
			"appName":    "",
		},
	).Return(&template.RenderedTemplate{
		Subject: "You're Invited to Register",
		Body:    "<html><body>Complete Registration</body></html>",
		IsHTML:  true,
	}, nil)

	var sentEmail email.EmailData
	suite.mockEmailClient.On("Send", mock.Anything).Run(func(args mock.Arguments) {
		sentEmail = args.Get(0).(email.EmailData)
	}).Return(nil)

	resp, err := suite.executor.Execute(ctx)

	suite.NoError(err)
	suite.Equal(common.ExecComplete, resp.Status)
	suite.Equal([]string{"user@example.com"}, sentEmail.To)
}

func (suite *EmailExecutorTestSuite) TestExecute_SendMode_EmailFromRuntimeData() {
	ctx := &core.NodeContext{
		ExecutionID:  "test-flow-id",
		ExecutorMode: ExecutorModeSend,
		UserInputs:   make(map[string]string),
		RuntimeData: map[string]string{
			"email":                     "runtime@example.com",
			common.RuntimeKeyInviteLink: "https://localhost:5190/gate/invite?executionId=test&inviteToken=abc",
		},
		NodeProperties: map[string]interface{}{
			"emailTemplate": "USER_INVITE",
		},
	}

	suite.mockTemplateService.On("Render",
		mock.Anything,
		template.ScenarioUserInvite,
		template.TemplateTypeEmail,
		template.TemplateData{
			"inviteLink": "https://localhost:5190/gate/invite?executionId=test&inviteToken=abc",
			"appName":    "",
		},
	).Return(&template.RenderedTemplate{
		Subject: "You're Invited to Register",
		Body:    "<html><body>Complete Registration</body></html>",
		IsHTML:  true,
	}, nil)

	var sentEmail email.EmailData
	suite.mockEmailClient.On("Send", mock.Anything).Run(func(args mock.Arguments) {
		sentEmail = args.Get(0).(email.EmailData)
	}).Return(nil)

	resp, err := suite.executor.Execute(ctx)

	suite.NoError(err)
	suite.Equal(common.ExecComplete, resp.Status)
	suite.Equal([]string{"runtime@example.com"}, sentEmail.To)
}

func (suite *EmailExecutorTestSuite) TestExecute_SendMode_MissingRecipient() {
	ctx := &core.NodeContext{
		ExecutionID:  "test-flow-id",
		ExecutorMode: ExecutorModeSend,
		UserInputs:   make(map[string]string),
		RuntimeData: map[string]string{
			common.RuntimeKeyInviteLink: "https://localhost:5190/gate/invite?executionId=test&inviteToken=abc",
		},
		NodeProperties: map[string]interface{}{
			"emailTemplate": "USER_INVITE",
		},
	}

	resp, err := suite.executor.Execute(ctx)

	suite.NoError(err)
	suite.Equal(common.ExecFailure, resp.Status)
	suite.Equal("Email recipient is required", resp.FailureReason)
	suite.mockEmailClient.AssertNotCalled(suite.T(), "Send", mock.Anything)
}

func (suite *EmailExecutorTestSuite) TestExecute_SendMode_MissingInviteLink() {
	ctx := &core.NodeContext{
		ExecutionID:  "test-flow-id",
		ExecutorMode: ExecutorModeSend,
		UserInputs: map[string]string{
			"email": "user@example.com",
		},
		RuntimeData: make(map[string]string),
		NodeProperties: map[string]interface{}{
			"emailTemplate": "USER_INVITE",
		},
	}

	resp, err := suite.executor.Execute(ctx)

	suite.Error(err)
	suite.Nil(resp)
	suite.Contains(err.Error(), "invite link not found")
	suite.mockEmailClient.AssertNotCalled(suite.T(), "Send", mock.Anything)
}

func (suite *EmailExecutorTestSuite) TestExecute_SendMode_SelfRegistration_MissingInviteLink() {
	ctx := &core.NodeContext{
		ExecutionID:  "test-flow-id",
		FlowType:     common.FlowTypeRegistration,
		ExecutorMode: ExecutorModeSend,
		UserInputs: map[string]string{
			"email": "user@example.com",
		},
		RuntimeData: make(map[string]string),
		NodeProperties: map[string]interface{}{
			"emailTemplate": "SELF_REGISTRATION",
		},
	}

	resp, err := suite.executor.Execute(ctx)

	suite.Error(err)
	suite.Nil(resp)
	suite.Contains(err.Error(), "invite link not found")
	suite.mockEmailClient.AssertNotCalled(suite.T(), "Send", mock.Anything)
}

func (suite *EmailExecutorTestSuite) TestExecute_SendMode_MissingTemplateProperty_DefaultsToUserInvite() {
	ctx := &core.NodeContext{
		ExecutionID:  "test-flow-id",
		ExecutorMode: ExecutorModeSend,
		UserInputs: map[string]string{
			"email": "user@example.com",
		},
		RuntimeData: map[string]string{
			common.RuntimeKeyInviteLink: "https://localhost:5190/gate/invite?executionId=test&inviteToken=abc",
		},
		NodeProperties: map[string]interface{}{},
	}

	// Verify that Render is called with ScenarioUserInvite even when the property is absent.
	suite.mockTemplateService.On("Render",
		mock.Anything,
		template.ScenarioUserInvite,
		template.TemplateTypeEmail,
		template.TemplateData{
			"inviteLink": "https://localhost:5190/gate/invite?executionId=test&inviteToken=abc",
			"appName":    "",
		},
	).Return(&template.RenderedTemplate{
		Subject: "You're Invited to Register",
		Body:    "<html><body>Complete Registration</body></html>",
		IsHTML:  true,
	}, nil)

	var sentEmail email.EmailData
	suite.mockEmailClient.On("Send", mock.Anything).Run(func(args mock.Arguments) {
		sentEmail = args.Get(0).(email.EmailData)
	}).Return(nil)

	resp, err := suite.executor.Execute(ctx)

	suite.NoError(err)
	suite.Equal(common.ExecComplete, resp.Status)
	suite.Equal([]string{"user@example.com"}, sentEmail.To)
}

func (suite *EmailExecutorTestSuite) TestExecute_SendMode_EmptyTemplateString_DefaultsToUserInvite() {
	ctx := &core.NodeContext{
		ExecutionID:  "test-flow-id",
		ExecutorMode: ExecutorModeSend,
		UserInputs: map[string]string{
			"email": "user@example.com",
		},
		RuntimeData: map[string]string{
			common.RuntimeKeyInviteLink: "https://localhost:5190/gate/invite?executionId=test&inviteToken=abc",
		},
		NodeProperties: map[string]interface{}{
			"emailTemplate": "",
		},
	}

	suite.mockTemplateService.On("Render",
		mock.Anything,
		template.ScenarioUserInvite,
		template.TemplateTypeEmail,
		template.TemplateData{
			"inviteLink": "https://localhost:5190/gate/invite?executionId=test&inviteToken=abc",
			"appName":    "",
		},
	).Return(&template.RenderedTemplate{
		Subject: "You're Invited to Register",
		Body:    "<html><body>Complete Registration</body></html>",
		IsHTML:  true,
	}, nil)

	var sentEmail email.EmailData
	suite.mockEmailClient.On("Send", mock.Anything).Run(func(args mock.Arguments) {
		sentEmail = args.Get(0).(email.EmailData)
	}).Return(nil)

	resp, err := suite.executor.Execute(ctx)

	suite.NoError(err)
	suite.Equal(common.ExecComplete, resp.Status)
	suite.Equal([]string{"user@example.com"}, sentEmail.To)
}

func (suite *EmailExecutorTestSuite) TestExecute_SendMode_InvalidTemplateType_ReturnsError() {
	ctx := &core.NodeContext{
		ExecutionID:  "test-flow-id",
		ExecutorMode: ExecutorModeSend,
		UserInputs: map[string]string{
			"email": "user@example.com",
		},
		RuntimeData: map[string]string{
			common.RuntimeKeyInviteLink: "https://localhost:5190/gate/invite?executionId=test&inviteToken=abc",
		},
		NodeProperties: map[string]interface{}{
			"emailTemplate": 123,
		},
	}

	resp, err := suite.executor.Execute(ctx)
	if suite.Error(err) {
		suite.Contains(err.Error(), "invalid type for emailTemplate")
	}
	suite.Nil(resp)
}

func (suite *EmailExecutorTestSuite) TestExecute_SendMode_TemplateRenderError() {
	ctx := &core.NodeContext{
		ExecutionID:  "test-flow-id",
		ExecutorMode: ExecutorModeSend,
		UserInputs: map[string]string{
			"email": "user@example.com",
		},
		RuntimeData: map[string]string{
			common.RuntimeKeyInviteLink: "https://localhost:5190/gate/invite?executionId=test&inviteToken=abc",
		},
		NodeProperties: map[string]interface{}{
			"emailTemplate": "USER_INVITE",
		},
	}

	suite.mockTemplateService.On("Render",
		mock.Anything,
		template.ScenarioUserInvite,
		template.TemplateTypeEmail,
		mock.Anything,
	).Return(nil, &serviceerror.I18nServiceError{Code: "TMP-5000"})

	resp, err := suite.executor.Execute(ctx)
	if suite.Error(err) {
		suite.Contains(err.Error(), "failed to render email template: TMP-5000")
	}
	suite.Nil(resp)
	suite.mockEmailClient.AssertNotCalled(suite.T(), "Send", mock.Anything)
}

func (suite *EmailExecutorTestSuite) TestExecute_SendMode_NilTemplateService() {
	mockBaseExecutor := coremock.NewExecutorInterfaceMock(suite.T())
	mockFactory := coremock.NewFlowFactoryInterfaceMock(suite.T())
	mockFactory.On("CreateExecutor",
		ExecutorNameEmailExecutor,
		common.ExecutorTypeUtility,
		[]common.Input{},
		[]common.Input{
			{Identifier: userAttributeEmail, Type: common.InputTypeText, Required: true},
		},
	).Return(mockBaseExecutor)

	noServiceExecutor := newEmailExecutor(mockFactory, suite.mockEmailClient, nil)

	ctx := &core.NodeContext{
		ExecutionID:  "test-flow-id",
		ExecutorMode: ExecutorModeSend,
		UserInputs: map[string]string{
			"email": "user@example.com",
		},
		RuntimeData: map[string]string{
			common.RuntimeKeyInviteLink: "https://localhost:5190/gate/invite?executionId=test&inviteToken=abc",
		},
		NodeProperties: map[string]interface{}{
			"emailTemplate": "USER_INVITE",
		},
	}

	resp, err := noServiceExecutor.Execute(ctx)
	if suite.Error(err) {
		suite.Contains(err.Error(), "template service is not configured")
	}
	suite.Nil(resp)
}

func (suite *EmailExecutorTestSuite) TestExecute_SendMode_ClientError() {
	ctx := &core.NodeContext{
		ExecutionID:  "test-flow-id",
		ExecutorMode: ExecutorModeSend,
		UserInputs: map[string]string{
			"email": "user@example.com",
		},
		RuntimeData: map[string]string{
			common.RuntimeKeyInviteLink: "https://localhost:5190/gate/invite?executionId=test&inviteToken=abc",
		},
		NodeProperties: map[string]interface{}{
			"emailTemplate": "USER_INVITE",
		},
	}

	suite.mockTemplateService.On("Render",
		mock.Anything,
		template.ScenarioUserInvite,
		template.TemplateTypeEmail,
		mock.Anything,
	).Return(&template.RenderedTemplate{
		Subject: "You're Invited to Register",
		Body:    "<html><body>Complete Registration</body></html>",
		IsHTML:  true,
	}, nil)

	suite.mockEmailClient.On("Send", mock.Anything).Return(email.ErrorInvalidRecipient)

	resp, err := suite.executor.Execute(ctx)

	suite.NoError(err)
	suite.Equal(common.ExecFailure, resp.Status)
	suite.Equal("Failed to send email", resp.FailureReason)
}

func (suite *EmailExecutorTestSuite) TestExecute_SendMode_ServerError() {
	ctx := &core.NodeContext{
		ExecutionID:  "test-flow-id",
		ExecutorMode: ExecutorModeSend,
		UserInputs: map[string]string{
			"email": "user@example.com",
		},
		RuntimeData: map[string]string{
			common.RuntimeKeyInviteLink: "https://localhost:5190/gate/invite?executionId=test&inviteToken=abc",
		},
		NodeProperties: map[string]interface{}{
			"emailTemplate": "USER_INVITE",
		},
	}

	suite.mockTemplateService.On("Render",
		mock.Anything,
		template.ScenarioUserInvite,
		template.TemplateTypeEmail,
		mock.Anything,
	).Return(&template.RenderedTemplate{
		Subject: "You're Invited to Register",
		Body:    "<html><body>Complete Registration</body></html>",
		IsHTML:  true,
	}, nil)

	suite.mockEmailClient.On("Send", mock.Anything).Return(email.ErrorSMTPConnection)

	resp, err := suite.executor.Execute(ctx)

	suite.Error(err)
	suite.Nil(resp)
}

func (suite *EmailExecutorTestSuite) TestExecute_SendMode_NilEmailClient_NoOp() {
	// Create executor with nil email client (SMTP not configured)
	mockBaseExecutor := coremock.NewExecutorInterfaceMock(suite.T())
	mockFactory := coremock.NewFlowFactoryInterfaceMock(suite.T())
	mockFactory.On("CreateExecutor",
		ExecutorNameEmailExecutor,
		common.ExecutorTypeUtility,
		[]common.Input{},
		[]common.Input{
			{Identifier: userAttributeEmail, Type: common.InputTypeText, Required: true},
		},
	).Return(mockBaseExecutor)

	noEmailExecutor := newEmailExecutor(mockFactory, nil, suite.mockTemplateService)

	ctx := &core.NodeContext{
		ExecutionID:  "test-flow-id",
		FlowType:     common.FlowTypeUserOnboarding,
		ExecutorMode: ExecutorModeSend,
		UserInputs: map[string]string{
			"email": "user@example.com",
		},
		RuntimeData: map[string]string{
			common.RuntimeKeyInviteLink: "https://localhost:5190/gate/invite?executionId=test&inviteToken=abc",
		},
		NodeProperties: map[string]interface{}{
			"emailTemplate": "USER_INVITE",
		},
	}

	resp, err := noEmailExecutor.Execute(ctx)

	suite.NoError(err)
	suite.Equal(common.ExecComplete, resp.Status)
	suite.Equal(dataValueFalse, resp.AdditionalData[common.DataEmailSent])
}

func (suite *EmailExecutorTestSuite) TestExecute_SendMode_NilEmailClient_SelfRegistration_InviteLinkNotExposed() {
	noEmailExecutor := newEmailExecutor(suite.mockFlowFactory, nil, suite.mockTemplateService)

	ctx := &core.NodeContext{
		ExecutionID:  "test-flow-id",
		FlowType:     common.FlowTypeRegistration,
		ExecutorMode: ExecutorModeSend,
		UserInputs: map[string]string{
			"email": "user@example.com",
		},
		RuntimeData: map[string]string{
			common.RuntimeKeyInviteLink: "https://localhost:5190/gate/invite?executionId=test&inviteToken=abc",
		},
		NodeProperties: map[string]interface{}{
			"emailTemplate": "SELF_REGISTRATION",
		},
	}

	resp, err := noEmailExecutor.Execute(ctx)

	suite.NoError(err)
	suite.Equal(common.ExecComplete, resp.Status)
	suite.Equal(dataValueFalse, resp.AdditionalData[common.DataEmailSent])
	suite.Empty(resp.AdditionalData[common.DataInviteLink])
}

func (suite *EmailExecutorTestSuite) TestExecute_InvalidMode() {
	ctx := &core.NodeContext{
		ExecutionID:  "test-flow-id",
		ExecutorMode: "invalid",
		UserInputs:   make(map[string]string),
		RuntimeData:  make(map[string]string),
	}

	resp, err := suite.executor.Execute(ctx)
	if suite.Error(err) {
		suite.Contains(err.Error(), "invalid executor mode for EmailExecutor")
	}
	suite.Nil(resp)
}

func TestEmailExecutorSuite(t *testing.T) {
	suite.Run(t, new(EmailExecutorTestSuite))
}
