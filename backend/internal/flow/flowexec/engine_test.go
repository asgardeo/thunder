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

package flowexec

import (
	"testing"

	"github.com/stretchr/testify/suite"

	authncm "github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/tests/mocks/flow/coremock"
	"github.com/asgardeo/thunder/tests/mocks/observability/observabilitymock"
)

type EngineTestSuite struct {
	suite.Suite
}

func TestEngineTestSuite(t *testing.T) {
	suite.Run(t, new(EngineTestSuite))
}

func (s *EngineTestSuite) TestGetNodeInputs_ExecutorBackedNode() {
	t := s.T()
	mockNode := coremock.NewExecutorBackedNodeInterfaceMock(t)
	expectedInputs := []common.Input{
		{Identifier: "username", Type: "string", Required: true},
		{Identifier: "password", Type: "string", Required: true},
	}
	mockNode.On("GetInputs").Return(expectedInputs)

	inputs := getNodeInputs(mockNode)

	s.NotNil(inputs)
	s.Len(inputs, 2)
	s.Equal("username", inputs[0].Identifier)
	s.Equal("password", inputs[1].Identifier)
}

func (s *EngineTestSuite) TestGetNodeInputs_PromptNode() {
	t := s.T()
	mockNode := coremock.NewPromptNodeInterfaceMock(t)
	prompts := []common.Prompt{
		{
			Inputs: []common.Input{
				{Identifier: "email", Type: "string", Required: true},
			},
		},
		{
			Inputs: []common.Input{
				{Identifier: "code", Type: "string", Required: true},
			},
		},
	}
	mockNode.On("GetPrompts").Return(prompts)

	inputs := getNodeInputs(mockNode)

	s.NotNil(inputs)
	s.Len(inputs, 2)
	s.Equal("email", inputs[0].Identifier)
	s.Equal("code", inputs[1].Identifier)
}

func (s *EngineTestSuite) TestGetNodeInputs_RegularNode() {
	mockNode := coremock.NewNodeInterfaceMock(s.T())

	inputs := getNodeInputs(mockNode)

	s.Nil(inputs)
}

func (s *EngineTestSuite) TestGetNodeInputs_NilNode() {
	inputs := getNodeInputs(nil)

	s.Nil(inputs)
}

func (s *EngineTestSuite) TestUpdateContextWithNodeResponse_AdditionalData() {
	t := s.T()
	mockObservability := observabilitymock.NewObservabilityServiceInterfaceMock(t)
	mockObservability.On("IsEnabled").Return(false).Maybe()

	fe := &flowEngine{
		observabilitySvc: mockObservability,
	}

	ctx := &EngineContext{
		RuntimeData: make(map[string]string),
	}

	nodeResp := &common.NodeResponse{
		Status: common.NodeStatusComplete,
		AdditionalData: map[string]string{
			"passkeyChallenge":       `{"challenge": "abc123"}`,
			"passkeyCreationOptions": `{"rpId": "example.com"}`,
		},
	}

	fe.updateContextWithNodeResponse(ctx, nodeResp)

	s.NotNil(ctx.AdditionalData)
	s.Equal(`{"challenge": "abc123"}`, ctx.AdditionalData["passkeyChallenge"])
	s.Equal(`{"rpId": "example.com"}`, ctx.AdditionalData["passkeyCreationOptions"])
}

func (s *EngineTestSuite) TestUpdateContextWithNodeResponse_MergesAdditionalData() {
	t := s.T()
	mockObservability := observabilitymock.NewObservabilityServiceInterfaceMock(t)
	mockObservability.On("IsEnabled").Return(false).Maybe()

	fe := &flowEngine{
		observabilitySvc: mockObservability,
	}

	ctx := &EngineContext{
		RuntimeData: make(map[string]string),
		AdditionalData: map[string]string{
			"existingKey": "existingValue",
		},
	}

	nodeResp := &common.NodeResponse{
		Status: common.NodeStatusComplete,
		AdditionalData: map[string]string{
			"newKey": "newValue",
		},
	}

	fe.updateContextWithNodeResponse(ctx, nodeResp)

	s.NotNil(ctx.AdditionalData)
	s.Equal("existingValue", ctx.AdditionalData["existingKey"])
	s.Equal("newValue", ctx.AdditionalData["newKey"])
}

func (s *EngineTestSuite) TestUpdateContextWithNodeResponse_ClearsActionOnComplete() {
	t := s.T()
	mockObservability := observabilitymock.NewObservabilityServiceInterfaceMock(t)
	mockObservability.On("IsEnabled").Return(false).Maybe()

	fe := &flowEngine{
		observabilitySvc: mockObservability,
	}

	ctx := &EngineContext{
		CurrentAction: "someAction",
		RuntimeData:   make(map[string]string),
	}

	nodeResp := &common.NodeResponse{
		Status: common.NodeStatusComplete,
	}

	fe.updateContextWithNodeResponse(ctx, nodeResp)

	s.Empty(ctx.CurrentAction)
}

func (s *EngineTestSuite) TestUpdateContextWithNodeResponse_ClearsActionOnForward() {
	t := s.T()
	mockObservability := observabilitymock.NewObservabilityServiceInterfaceMock(t)
	mockObservability.On("IsEnabled").Return(false).Maybe()

	fe := &flowEngine{
		observabilitySvc: mockObservability,
	}

	ctx := &EngineContext{
		CurrentAction: "someAction",
		RuntimeData:   make(map[string]string),
	}

	nodeResp := &common.NodeResponse{
		Status: common.NodeStatusForward,
	}

	fe.updateContextWithNodeResponse(ctx, nodeResp)

	s.Empty(ctx.CurrentAction)
}

func (s *EngineTestSuite) TestUpdateContextWithNodeResponse_PreservesActionOnIncomplete() {
	t := s.T()
	mockObservability := observabilitymock.NewObservabilityServiceInterfaceMock(t)
	mockObservability.On("IsEnabled").Return(false).Maybe()

	fe := &flowEngine{
		observabilitySvc: mockObservability,
	}

	ctx := &EngineContext{
		CurrentAction: "passkeyChallenge",
		RuntimeData:   make(map[string]string),
	}

	nodeResp := &common.NodeResponse{
		Status: common.NodeStatusIncomplete,
	}

	fe.updateContextWithNodeResponse(ctx, nodeResp)

	s.Equal("passkeyChallenge", ctx.CurrentAction)
}

func (s *EngineTestSuite) TestResolveStepForRedirection_WithAdditionalData() {
	fe := &flowEngine{}

	ctx := &EngineContext{
		AdditionalData: map[string]string{
			"passkeyChallenge": `{"challenge": "xyz789"}`,
			"sessionToken":     "abc123",
		},
	}

	nodeResp := &common.NodeResponse{
		RedirectURL: "https://example.com/auth",
	}

	flowStep := &FlowStep{
		Data: FlowData{},
	}

	err := fe.resolveStepForRedirection(ctx, nodeResp, flowStep)

	s.NoError(err)
	s.Equal("https://example.com/auth", flowStep.Data.RedirectURL)
	s.NotNil(flowStep.Data.AdditionalData)
	s.Equal(`{"challenge": "xyz789"}`, flowStep.Data.AdditionalData["passkeyChallenge"])
	s.Equal("abc123", flowStep.Data.AdditionalData["sessionToken"])
}

func (s *EngineTestSuite) TestResolveStepForRedirection_NoAdditionalData() {
	fe := &flowEngine{}

	ctx := &EngineContext{}

	nodeResp := &common.NodeResponse{
		RedirectURL: "https://example.com/auth",
	}

	flowStep := &FlowStep{
		Data: FlowData{},
	}

	err := fe.resolveStepForRedirection(ctx, nodeResp, flowStep)

	s.NoError(err)
	s.Equal("https://example.com/auth", flowStep.Data.RedirectURL)
	s.Nil(flowStep.Data.AdditionalData)
}

func (s *EngineTestSuite) TestResolveStepForRedirection_NilNodeResponse() {
	fe := &flowEngine{}
	ctx := &EngineContext{}
	flowStep := &FlowStep{}

	err := fe.resolveStepForRedirection(ctx, nil, flowStep)

	s.Error(err)
	s.Contains(err.Error(), "node response is nil")
}

func (s *EngineTestSuite) TestResolveStepForRedirection_EmptyRedirectURL() {
	fe := &flowEngine{}
	ctx := &EngineContext{}
	nodeResp := &common.NodeResponse{
		RedirectURL: "",
	}
	flowStep := &FlowStep{}

	err := fe.resolveStepForRedirection(ctx, nodeResp, flowStep)

	s.Error(err)
	s.Contains(err.Error(), "redirect URL not found")
}

func (s *EngineTestSuite) TestResolveStepDetailsForPrompt_WithAdditionalData() {
	fe := &flowEngine{}

	ctx := &EngineContext{
		AdditionalData: map[string]string{
			"passkeyCreationOptions": `{"rpId": "example.com"}`,
		},
	}

	nodeResp := &common.NodeResponse{
		Inputs: []common.Input{
			{Identifier: "username", Type: "string", Required: true},
		},
	}

	flowStep := &FlowStep{
		Data: FlowData{},
	}

	err := fe.resolveStepDetailsForPrompt(ctx, nodeResp, flowStep)

	s.NoError(err)
	s.NotNil(flowStep.Data.AdditionalData)
	s.Equal(`{"rpId": "example.com"}`, flowStep.Data.AdditionalData["passkeyCreationOptions"])
}

func (s *EngineTestSuite) TestResolveStepDetailsForPrompt_WithActions() {
	fe := &flowEngine{}

	ctx := &EngineContext{}

	nodeResp := &common.NodeResponse{
		Actions: []common.Action{
			{Ref: "submit-action", NextNode: "next-node"},
		},
	}

	flowStep := &FlowStep{
		Data: FlowData{},
	}

	err := fe.resolveStepDetailsForPrompt(ctx, nodeResp, flowStep)

	s.NoError(err)
	s.Len(flowStep.Data.Actions, 1)
	s.Equal("submit-action", flowStep.Data.Actions[0].Ref)
}

func (s *EngineTestSuite) TestResolveStepDetailsForPrompt_NilNodeResponse() {
	fe := &flowEngine{}
	ctx := &EngineContext{}
	flowStep := &FlowStep{}

	err := fe.resolveStepDetailsForPrompt(ctx, nil, flowStep)

	s.Error(err)
	s.Contains(err.Error(), "node response is nil")
}

func (s *EngineTestSuite) TestResolveStepDetailsForPrompt_NoInputsOrActions() {
	fe := &flowEngine{}
	ctx := &EngineContext{}
	nodeResp := &common.NodeResponse{}
	flowStep := &FlowStep{}

	err := fe.resolveStepDetailsForPrompt(ctx, nodeResp, flowStep)

	s.Error(err)
	s.Contains(err.Error(), "no required data or actions found")
}

func (s *EngineTestSuite) TestUpdateContextWithNodeResponse_RuntimeData() {
	t := s.T()
	mockObservability := observabilitymock.NewObservabilityServiceInterfaceMock(t)
	mockObservability.On("IsEnabled").Return(false).Maybe()

	fe := &flowEngine{
		observabilitySvc: mockObservability,
	}

	ctx := &EngineContext{
		RuntimeData: map[string]string{"existing": "value"},
	}

	nodeResp := &common.NodeResponse{
		Status: common.NodeStatusComplete,
		RuntimeData: map[string]string{
			"newKey": "newValue",
		},
	}

	fe.updateContextWithNodeResponse(ctx, nodeResp)

	s.Equal("value", ctx.RuntimeData["existing"])
	s.Equal("newValue", ctx.RuntimeData["newKey"])
}

func (s *EngineTestSuite) TestUpdateContextWithNodeResponse_RuntimeDataNilContext() {
	t := s.T()
	mockObservability := observabilitymock.NewObservabilityServiceInterfaceMock(t)
	mockObservability.On("IsEnabled").Return(false).Maybe()

	fe := &flowEngine{
		observabilitySvc: mockObservability,
	}

	ctx := &EngineContext{} // No RuntimeData initialized

	nodeResp := &common.NodeResponse{
		Status: common.NodeStatusComplete,
		RuntimeData: map[string]string{
			"userID": "user-123",
		},
	}

	fe.updateContextWithNodeResponse(ctx, nodeResp)

	s.NotNil(ctx.RuntimeData)
	s.Equal("user-123", ctx.RuntimeData["userID"])
}

func (s *EngineTestSuite) TestUpdateContextWithNodeResponse_Assertion() {
	t := s.T()
	mockObservability := observabilitymock.NewObservabilityServiceInterfaceMock(t)
	mockObservability.On("IsEnabled").Return(false).Maybe()

	fe := &flowEngine{
		observabilitySvc: mockObservability,
	}

	ctx := &EngineContext{}

	nodeResp := &common.NodeResponse{
		Status:    common.NodeStatusComplete,
		Assertion: "test-assertion-token",
	}

	fe.updateContextWithNodeResponse(ctx, nodeResp)

	s.Equal("test-assertion-token", ctx.Assertion)
}

func (s *EngineTestSuite) TestUpdateContextWithNodeResponse_AuthenticatedUserUpdate() {
	t := s.T()
	mockObservability := observabilitymock.NewObservabilityServiceInterfaceMock(t)
	mockObservability.On("IsEnabled").Return(false).Maybe()

	mockExecutor := coremock.NewExecutorInterfaceMock(t)
	mockExecutor.On("GetType").Return(common.ExecutorTypeAuthentication)

	mockNode := coremock.NewExecutorBackedNodeInterfaceMock(t)
	mockNode.On("GetType").Return(common.NodeTypeTaskExecution)
	mockNode.On("GetExecutor").Return(mockExecutor)

	fe := &flowEngine{
		observabilitySvc: mockObservability,
	}

	ctx := &EngineContext{
		CurrentNode: mockNode,
		FlowType:    common.FlowTypeAuthentication,
	}

	nodeResp := &common.NodeResponse{
		Status: common.NodeStatusComplete,
		AuthenticatedUser: authncm.AuthenticatedUser{
			UserID:          "user-123",
			IsAuthenticated: true,
		},
	}

	fe.updateContextWithNodeResponse(ctx, nodeResp)

	s.True(ctx.AuthenticatedUser.IsAuthenticated)
	s.Equal("user-123", ctx.AuthenticatedUser.UserID)
	s.Equal("user-123", ctx.RuntimeData["userID"])
}

func (s *EngineTestSuite) TestUpdateContextWithNodeResponse_MergesUserAttributes() {
	t := s.T()
	mockObservability := observabilitymock.NewObservabilityServiceInterfaceMock(t)
	mockObservability.On("IsEnabled").Return(false).Maybe()

	mockExecutor := coremock.NewExecutorInterfaceMock(t)
	mockExecutor.On("GetType").Return(common.ExecutorTypeAuthentication)

	mockNode := coremock.NewExecutorBackedNodeInterfaceMock(t)
	mockNode.On("GetType").Return(common.NodeTypeTaskExecution)
	mockNode.On("GetExecutor").Return(mockExecutor)

	fe := &flowEngine{
		observabilitySvc: mockObservability,
	}

	ctx := &EngineContext{
		CurrentNode: mockNode,
		FlowType:    common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			Attributes: map[string]interface{}{
				"existingAttr": "existingValue",
			},
		},
	}

	nodeResp := &common.NodeResponse{
		Status: common.NodeStatusComplete,
		AuthenticatedUser: authncm.AuthenticatedUser{
			UserID:          "user-456",
			IsAuthenticated: true,
			Attributes: map[string]interface{}{
				"newAttr": "newValue",
			},
		},
	}

	fe.updateContextWithNodeResponse(ctx, nodeResp)

	s.True(ctx.AuthenticatedUser.IsAuthenticated)
	s.Equal("existingValue", ctx.AuthenticatedUser.Attributes["existingAttr"])
	s.Equal("newValue", ctx.AuthenticatedUser.Attributes["newAttr"])
}

func (s *EngineTestSuite) TestUpdateContextWithNodeResponse_PreservesExistingUserID() {
	t := s.T()
	mockObservability := observabilitymock.NewObservabilityServiceInterfaceMock(t)
	mockObservability.On("IsEnabled").Return(false).Maybe()

	mockExecutor := coremock.NewExecutorInterfaceMock(t)
	mockExecutor.On("GetType").Return(common.ExecutorTypeAuthentication)

	mockNode := coremock.NewExecutorBackedNodeInterfaceMock(t)
	mockNode.On("GetType").Return(common.NodeTypeTaskExecution)
	mockNode.On("GetExecutor").Return(mockExecutor)

	fe := &flowEngine{
		observabilitySvc: mockObservability,
	}

	ctx := &EngineContext{
		CurrentNode: mockNode,
		FlowType:    common.FlowTypeAuthentication,
		RuntimeData: map[string]string{
			"userID": "existing-user-id",
		},
	}

	nodeResp := &common.NodeResponse{
		Status: common.NodeStatusComplete,
		AuthenticatedUser: authncm.AuthenticatedUser{
			UserID:          "new-user-id",
			IsAuthenticated: true,
		},
	}

	fe.updateContextWithNodeResponse(ctx, nodeResp)

	// Existing userID in RuntimeData should be preserved
	s.Equal("existing-user-id", ctx.RuntimeData["userID"])
}

func (s *EngineTestSuite) TestUpdateContextWithNodeResponse_PreviousAttrsNilNewAttrs() {
	t := s.T()
	mockObservability := observabilitymock.NewObservabilityServiceInterfaceMock(t)
	mockObservability.On("IsEnabled").Return(false).Maybe()

	mockExecutor := coremock.NewExecutorInterfaceMock(t)
	mockExecutor.On("GetType").Return(common.ExecutorTypeAuthentication)

	mockNode := coremock.NewExecutorBackedNodeInterfaceMock(t)
	mockNode.On("GetType").Return(common.NodeTypeTaskExecution)
	mockNode.On("GetExecutor").Return(mockExecutor)

	fe := &flowEngine{
		observabilitySvc: mockObservability,
	}

	ctx := &EngineContext{
		CurrentNode: mockNode,
		FlowType:    common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			Attributes: map[string]interface{}{
				"prevAttr": "prevValue",
			},
		},
	}

	// Node response has nil Attributes
	nodeResp := &common.NodeResponse{
		Status: common.NodeStatusComplete,
		AuthenticatedUser: authncm.AuthenticatedUser{
			UserID:          "user-789",
			IsAuthenticated: true,
			Attributes:      nil,
		},
	}

	fe.updateContextWithNodeResponse(ctx, nodeResp)

	// Previous attributes should be preserved when new ones are nil
	s.Equal("prevValue", ctx.AuthenticatedUser.Attributes["prevAttr"])
}

func (s *EngineTestSuite) TestShouldUpdateAuthenticatedUser_NilNode() {
	fe := &flowEngine{}
	ctx := &EngineContext{
		CurrentNode: nil,
	}

	result := fe.shouldUpdateAuthenticatedUser(ctx)

	s.False(result)
}

func (s *EngineTestSuite) TestShouldUpdateAuthenticatedUser_NonTaskExecutionNode() {
	mockNode := coremock.NewNodeInterfaceMock(s.T())
	mockNode.On("GetType").Return(common.NodeTypePrompt)

	fe := &flowEngine{}
	ctx := &EngineContext{
		CurrentNode: mockNode,
	}

	result := fe.shouldUpdateAuthenticatedUser(ctx)

	s.False(result)
}

func (s *EngineTestSuite) TestShouldUpdateAuthenticatedUser_NonExecutorBackedNode() {
	mockNode := coremock.NewNodeInterfaceMock(s.T())
	mockNode.On("GetType").Return(common.NodeTypeTaskExecution)

	fe := &flowEngine{}
	ctx := &EngineContext{
		CurrentNode: mockNode,
	}

	result := fe.shouldUpdateAuthenticatedUser(ctx)

	s.False(result)
}

func (s *EngineTestSuite) TestShouldUpdateAuthenticatedUser_NilExecutor() {
	mockNode := coremock.NewExecutorBackedNodeInterfaceMock(s.T())
	mockNode.On("GetType").Return(common.NodeTypeTaskExecution)
	mockNode.On("GetExecutor").Return(nil)

	fe := &flowEngine{}
	ctx := &EngineContext{
		CurrentNode: mockNode,
	}

	result := fe.shouldUpdateAuthenticatedUser(ctx)

	s.False(result)
}

func (s *EngineTestSuite) TestShouldUpdateAuthenticatedUser_AuthFlowWithAuthExecutor() {
	t := s.T()
	mockExecutor := coremock.NewExecutorInterfaceMock(t)
	mockExecutor.On("GetType").Return(common.ExecutorTypeAuthentication)

	mockNode := coremock.NewExecutorBackedNodeInterfaceMock(t)
	mockNode.On("GetType").Return(common.NodeTypeTaskExecution)
	mockNode.On("GetExecutor").Return(mockExecutor)

	fe := &flowEngine{}
	ctx := &EngineContext{
		CurrentNode: mockNode,
		FlowType:    common.FlowTypeAuthentication,
	}

	result := fe.shouldUpdateAuthenticatedUser(ctx)

	s.True(result)
}

func (s *EngineTestSuite) TestShouldUpdateAuthenticatedUser_AuthFlowWithProvisioningExecutor() {
	t := s.T()
	mockExecutor := coremock.NewExecutorInterfaceMock(t)
	mockExecutor.On("GetType").Return(common.ExecutorTypeRegistration)
	mockExecutor.On("GetName").Return("ProvisioningExecutor")

	mockNode := coremock.NewExecutorBackedNodeInterfaceMock(t)
	mockNode.On("GetType").Return(common.NodeTypeTaskExecution)
	mockNode.On("GetExecutor").Return(mockExecutor)

	fe := &flowEngine{}
	ctx := &EngineContext{
		CurrentNode: mockNode,
		FlowType:    common.FlowTypeAuthentication,
		RuntimeData: map[string]string{
			common.RuntimeKeyUserEligibleForProvisioning: "true",
		},
	}

	result := fe.shouldUpdateAuthenticatedUser(ctx)

	s.True(result)
}

func (s *EngineTestSuite) TestShouldUpdateAuthenticatedUser_AuthFlowWithNonAuthExecutor() {
	t := s.T()
	mockExecutor := coremock.NewExecutorInterfaceMock(t)
	mockExecutor.On("GetType").Return(common.ExecutorTypeUtility)

	mockNode := coremock.NewExecutorBackedNodeInterfaceMock(t)
	mockNode.On("GetType").Return(common.NodeTypeTaskExecution)
	mockNode.On("GetExecutor").Return(mockExecutor)

	fe := &flowEngine{}
	ctx := &EngineContext{
		CurrentNode: mockNode,
		FlowType:    common.FlowTypeAuthentication,
	}

	result := fe.shouldUpdateAuthenticatedUser(ctx)

	s.False(result)
}

func (s *EngineTestSuite) TestShouldUpdateAuthenticatedUser_RegistrationFlowWithProvisioning() {
	t := s.T()
	mockExecutor := coremock.NewExecutorInterfaceMock(t)
	mockExecutor.On("GetName").Return("ProvisioningExecutor")

	mockNode := coremock.NewExecutorBackedNodeInterfaceMock(t)
	mockNode.On("GetType").Return(common.NodeTypeTaskExecution)
	mockNode.On("GetExecutor").Return(mockExecutor)

	fe := &flowEngine{}
	ctx := &EngineContext{
		CurrentNode: mockNode,
		FlowType:    common.FlowTypeRegistration,
	}

	result := fe.shouldUpdateAuthenticatedUser(ctx)

	s.True(result)
}

func (s *EngineTestSuite) TestShouldUpdateAuthenticatedUser_RegistrationFlowSkipProvisioning() {
	t := s.T()
	mockExecutor := coremock.NewExecutorInterfaceMock(t)
	mockExecutor.On("GetType").Return(common.ExecutorTypeAuthentication)

	mockNode := coremock.NewExecutorBackedNodeInterfaceMock(t)
	mockNode.On("GetType").Return(common.NodeTypeTaskExecution)
	mockNode.On("GetExecutor").Return(mockExecutor)

	fe := &flowEngine{}
	ctx := &EngineContext{
		CurrentNode: mockNode,
		FlowType:    common.FlowTypeRegistration,
		RuntimeData: map[string]string{
			common.RuntimeKeySkipProvisioning: "true",
		},
	}

	result := fe.shouldUpdateAuthenticatedUser(ctx)

	s.True(result)
}

func (s *EngineTestSuite) TestResolveStepForRedirection_WithInputs() {
	fe := &flowEngine{}

	ctx := &EngineContext{}

	nodeResp := &common.NodeResponse{
		RedirectURL: "https://example.com/auth",
		Inputs: []common.Input{
			{Identifier: "code", Type: "string", Required: true},
		},
	}

	flowStep := &FlowStep{
		Data: FlowData{},
	}

	err := fe.resolveStepForRedirection(ctx, nodeResp, flowStep)

	s.NoError(err)
	s.Len(flowStep.Data.Inputs, 1)
	s.Equal("code", flowStep.Data.Inputs[0].Identifier)
	s.Equal(common.FlowStatusIncomplete, flowStep.Status)
	s.Equal(common.StepTypeRedirection, flowStep.Type)
}

func (s *EngineTestSuite) TestResolveStepForRedirection_AppendsInputs() {
	fe := &flowEngine{}

	ctx := &EngineContext{}

	nodeResp := &common.NodeResponse{
		RedirectURL: "https://example.com/auth",
		Inputs: []common.Input{
			{Identifier: "code", Type: "string", Required: true},
		},
	}

	flowStep := &FlowStep{
		Data: FlowData{
			Inputs: []common.Input{
				{Identifier: "state", Type: "string", Required: true},
			},
		},
	}

	err := fe.resolveStepForRedirection(ctx, nodeResp, flowStep)

	s.NoError(err)
	s.Len(flowStep.Data.Inputs, 2)
}

func (s *EngineTestSuite) TestResolveStepDetailsForPrompt_WithMeta() {
	fe := &flowEngine{}

	ctx := &EngineContext{}

	nodeResp := &common.NodeResponse{
		Inputs: []common.Input{
			{Identifier: "username", Type: "string", Required: true},
		},
		Meta: map[string]interface{}{
			"title":       "Login",
			"description": "Enter your credentials",
		},
	}

	flowStep := &FlowStep{
		Data: FlowData{},
	}

	err := fe.resolveStepDetailsForPrompt(ctx, nodeResp, flowStep)

	s.NoError(err)
	s.NotNil(flowStep.Data.Meta)
}

func (s *EngineTestSuite) TestResolveStepDetailsForPrompt_WithFailureReason() {
	fe := &flowEngine{}

	ctx := &EngineContext{}

	nodeResp := &common.NodeResponse{
		Inputs: []common.Input{
			{Identifier: "otp", Type: "string", Required: true},
		},
		FailureReason: "Invalid OTP provided",
	}

	flowStep := &FlowStep{
		Data: FlowData{},
	}

	err := fe.resolveStepDetailsForPrompt(ctx, nodeResp, flowStep)

	s.NoError(err)
	s.Equal("Invalid OTP provided", flowStep.FailureReason)
	s.Equal(common.FlowStatusIncomplete, flowStep.Status)
	s.Equal(common.StepTypeView, flowStep.Type)
}

func (s *EngineTestSuite) TestResolveStepDetailsForPrompt_AppendsInputs() {
	fe := &flowEngine{}

	ctx := &EngineContext{}

	nodeResp := &common.NodeResponse{
		Inputs: []common.Input{
			{Identifier: "password", Type: "string", Required: true},
		},
	}

	flowStep := &FlowStep{
		Data: FlowData{
			Inputs: []common.Input{
				{Identifier: "username", Type: "string", Required: true},
			},
		},
	}

	err := fe.resolveStepDetailsForPrompt(ctx, nodeResp, flowStep)

	s.NoError(err)
	s.Len(flowStep.Data.Inputs, 2)
}

func (s *EngineTestSuite) TestResolveStepDetailsForPrompt_ExistingActions() {
	fe := &flowEngine{}

	ctx := &EngineContext{}

	nodeResp := &common.NodeResponse{
		Actions: []common.Action{
			{Ref: "submit-action"},
		},
	}

	flowStep := &FlowStep{
		Data: FlowData{
			Actions: []common.Action{
				{Ref: "existing-action"},
			},
		},
	}

	err := fe.resolveStepDetailsForPrompt(ctx, nodeResp, flowStep)

	s.NoError(err)
	// Actions are replaced, not appended
	s.Len(flowStep.Data.Actions, 1)
	s.Equal("submit-action", flowStep.Data.Actions[0].Ref)
}

func (s *EngineTestSuite) TestGetNodeInputs_PromptNodeEmptyInputs() {
	mockNode := coremock.NewPromptNodeInterfaceMock(s.T())
	prompts := []common.Prompt{
		{
			Inputs: []common.Input{},
		},
	}
	mockNode.On("GetPrompts").Return(prompts)

	inputs := getNodeInputs(mockNode)

	s.Nil(inputs)
}

func (s *EngineTestSuite) TestShouldUpdateAuthenticatedUser_UserOnboardingFlowWithAuthExecutor() {
	t := s.T()
	mockExecutor := coremock.NewExecutorInterfaceMock(t)
	mockExecutor.On("GetType").Return(common.ExecutorTypeAuthentication)

	mockNode := coremock.NewExecutorBackedNodeInterfaceMock(t)
	mockNode.On("GetType").Return(common.NodeTypeTaskExecution)
	mockNode.On("GetExecutor").Return(mockExecutor)

	fe := &flowEngine{}
	ctx := &EngineContext{
		CurrentNode: mockNode,
		FlowType:    common.FlowTypeUserOnboarding,
	}

	result := fe.shouldUpdateAuthenticatedUser(ctx)

	s.True(result)
}

func (s *EngineTestSuite) TestShouldUpdateAuthenticatedUser_UserOnboardingFlowWithProvisioningExecutor() {
	t := s.T()
	mockExecutor := coremock.NewExecutorInterfaceMock(t)
	mockExecutor.On("GetType").Return(common.ExecutorTypeRegistration)
	mockExecutor.On("GetName").Return("ProvisioningExecutor")

	mockNode := coremock.NewExecutorBackedNodeInterfaceMock(t)
	mockNode.On("GetType").Return(common.NodeTypeTaskExecution)
	mockNode.On("GetExecutor").Return(mockExecutor)

	fe := &flowEngine{}
	ctx := &EngineContext{
		CurrentNode: mockNode,
		FlowType:    common.FlowTypeUserOnboarding,
	}

	result := fe.shouldUpdateAuthenticatedUser(ctx)

	s.True(result)
}

func (s *EngineTestSuite) TestShouldUpdateAuthenticatedUser_UserOnboardingFlowWithOtherExecutor() {
	t := s.T()
	mockExecutor := coremock.NewExecutorInterfaceMock(t)
	mockExecutor.On("GetType").Return(common.ExecutorTypeUtility)
	mockExecutor.On("GetName").Return("SomeOtherExecutor")

	mockNode := coremock.NewExecutorBackedNodeInterfaceMock(t)
	mockNode.On("GetType").Return(common.NodeTypeTaskExecution)
	mockNode.On("GetExecutor").Return(mockExecutor)

	fe := &flowEngine{}
	ctx := &EngineContext{
		CurrentNode: mockNode,
		FlowType:    common.FlowTypeUserOnboarding,
	}

	result := fe.shouldUpdateAuthenticatedUser(ctx)

	s.False(result)
}

func (s *EngineTestSuite) TestClearSensitiveInputs_AuthFlowRemovesPassword() {
	mockNode := coremock.NewExecutorBackedNodeInterfaceMock(s.T())
	mockNode.On("GetInputs").Return([]common.Input{
		{Identifier: "username", Type: "TEXT_INPUT", Required: true},
		{Identifier: "password", Type: common.InputTypePassword, Required: true},
	})
	mockNode.On("GetExecutor").Return(nil).Maybe()

	fe := &flowEngine{}
	ctx := &EngineContext{
		FlowType: common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			"username": "testuser",
			"password": "secret123",
		},
	}

	fe.clearSensitiveInputs(ctx, mockNode)

	s.Equal("testuser", ctx.UserInputs["username"])
	_, exists := ctx.UserInputs["password"]
	s.False(exists)
}

func (s *EngineTestSuite) TestClearSensitiveInputs_AuthFlowRemovesOTP() {
	mockNode := coremock.NewExecutorBackedNodeInterfaceMock(s.T())
	mockNode.On("GetInputs").Return([]common.Input{
		{Identifier: "otp", Type: common.InputTypeOTP, Required: true},
	})
	mockNode.On("GetExecutor").Return(nil).Maybe()

	fe := &flowEngine{}
	ctx := &EngineContext{
		FlowType: common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			"otp": "123456",
		},
	}

	fe.clearSensitiveInputs(ctx, mockNode)

	_, exists := ctx.UserInputs["otp"]
	s.False(exists)
}

func (s *EngineTestSuite) TestClearSensitiveInputs_RegistrationFlowRetainsPassword() {
	mockNode := coremock.NewExecutorBackedNodeInterfaceMock(s.T())
	mockNode.On("GetInputs").Return([]common.Input{
		{Identifier: "password", Type: common.InputTypePassword, Required: true},
	}).Maybe()

	fe := &flowEngine{}
	ctx := &EngineContext{
		FlowType: common.FlowTypeRegistration,
		UserInputs: map[string]string{
			"password": "secret123",
		},
	}

	fe.clearSensitiveInputs(ctx, mockNode)

	s.Equal("secret123", ctx.UserInputs["password"])
}

func (s *EngineTestSuite) TestClearSensitiveInputs_NoNodeInputs() {
	mockNode := coremock.NewNodeInterfaceMock(s.T())

	fe := &flowEngine{}
	ctx := &EngineContext{
		FlowType: common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			"password": "secret123",
		},
	}

	fe.clearSensitiveInputs(ctx, mockNode)

	// Password should remain since the node has no declared inputs
	s.Equal("secret123", ctx.UserInputs["password"])
}

func (s *EngineTestSuite) TestClearSensitiveInputs_NonSensitiveInputsRetained() {
	mockNode := coremock.NewExecutorBackedNodeInterfaceMock(s.T())
	mockNode.On("GetInputs").Return([]common.Input{
		{Identifier: "username", Type: "TEXT_INPUT", Required: true},
	})
	mockNode.On("GetExecutor").Return(nil).Maybe()

	fe := &flowEngine{}
	ctx := &EngineContext{
		FlowType: common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			"username": "testuser",
		},
	}

	fe.clearSensitiveInputs(ctx, mockNode)

	s.Equal("testuser", ctx.UserInputs["username"])
}

func (s *EngineTestSuite) TestClearSensitiveInputs_NoNodeInputsUsesExecutorDefaults() {
	// Node has no configured inputs, but executor defaults have PASSWORD_INPUT.
	mockExecutor := coremock.NewExecutorInterfaceMock(s.T())
	mockExecutor.On("GetDefaultInputs").Return([]common.Input{
		{Identifier: "username", Type: "string", Required: true},
		{Identifier: "password", Type: common.InputTypePassword, Required: true},
	})

	mockNode := coremock.NewExecutorBackedNodeInterfaceMock(s.T())
	mockNode.On("GetInputs").Return([]common.Input{})
	mockNode.On("GetExecutor").Return(mockExecutor)

	fe := &flowEngine{}
	ctx := &EngineContext{
		FlowType: common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			"username": "testuser",
			"password": "secret123",
		},
	}

	fe.clearSensitiveInputs(ctx, mockNode)

	s.Equal("testuser", ctx.UserInputs["username"])
	_, exists := ctx.UserInputs["password"]
	s.False(exists)
}

func (s *EngineTestSuite) TestClearSensitiveInputs_UserOnboardingFlowRetainsPassword() {
	mockNode := coremock.NewExecutorBackedNodeInterfaceMock(s.T())
	mockNode.On("GetInputs").Return([]common.Input{
		{Identifier: "password", Type: common.InputTypePassword, Required: true},
	}).Maybe()

	fe := &flowEngine{}
	ctx := &EngineContext{
		FlowType: common.FlowTypeUserOnboarding,
		UserInputs: map[string]string{
			"password": "secret123",
		},
	}

	fe.clearSensitiveInputs(ctx, mockNode)

	s.Equal("secret123", ctx.UserInputs["password"])
}
