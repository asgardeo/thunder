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

package flowexec

import (
	"testing"

	"github.com/stretchr/testify/assert"

	authncm "github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/tests/mocks/flow/coremock"
	"github.com/asgardeo/thunder/tests/mocks/observabilitymock"
)

func TestGetNodeInputs_ExecutorBackedNode(t *testing.T) {
	mockNode := coremock.NewExecutorBackedNodeInterfaceMock(t)
	expectedInputs := []common.Input{
		{Identifier: "username", Type: "string", Required: true},
		{Identifier: "password", Type: "string", Required: true},
	}
	mockNode.On("GetInputs").Return(expectedInputs)

	inputs := getNodeInputs(mockNode)

	assert.NotNil(t, inputs)
	assert.Len(t, inputs, 2)
	assert.Equal(t, "username", inputs[0].Identifier)
	assert.Equal(t, "password", inputs[1].Identifier)
}

func TestGetNodeInputs_PromptNode(t *testing.T) {
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

	assert.NotNil(t, inputs)
	assert.Len(t, inputs, 2)
	assert.Equal(t, "email", inputs[0].Identifier)
	assert.Equal(t, "code", inputs[1].Identifier)
}

func TestGetNodeInputs_RegularNode(t *testing.T) {
	mockNode := coremock.NewNodeInterfaceMock(t)

	inputs := getNodeInputs(mockNode)

	assert.Nil(t, inputs)
}

func TestGetNodeInputs_NilNode(t *testing.T) {
	inputs := getNodeInputs(nil)

	assert.Nil(t, inputs)
}

func TestUpdateContextWithNodeResponse_AdditionalData(t *testing.T) {
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

	assert.NotNil(t, ctx.AdditionalData)
	assert.Equal(t, `{"challenge": "abc123"}`, ctx.AdditionalData["passkeyChallenge"])
	assert.Equal(t, `{"rpId": "example.com"}`, ctx.AdditionalData["passkeyCreationOptions"])
}

func TestUpdateContextWithNodeResponse_MergesAdditionalData(t *testing.T) {
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

	assert.NotNil(t, ctx.AdditionalData)
	assert.Equal(t, "existingValue", ctx.AdditionalData["existingKey"])
	assert.Equal(t, "newValue", ctx.AdditionalData["newKey"])
}

func TestUpdateContextWithNodeResponse_ClearsActionOnComplete(t *testing.T) {
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

	assert.Empty(t, ctx.CurrentAction)
}

func TestUpdateContextWithNodeResponse_ClearsActionOnForward(t *testing.T) {
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

	assert.Empty(t, ctx.CurrentAction)
}

func TestUpdateContextWithNodeResponse_PreservesActionOnIncomplete(t *testing.T) {
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

	assert.Equal(t, "passkeyChallenge", ctx.CurrentAction)
}

func TestResolveStepForRedirection_WithAdditionalData(t *testing.T) {
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

	assert.NoError(t, err)
	assert.Equal(t, "https://example.com/auth", flowStep.Data.RedirectURL)
	assert.NotNil(t, flowStep.Data.AdditionalData)
	assert.Equal(t, `{"challenge": "xyz789"}`, flowStep.Data.AdditionalData["passkeyChallenge"])
	assert.Equal(t, "abc123", flowStep.Data.AdditionalData["sessionToken"])
}

func TestResolveStepForRedirection_NoAdditionalData(t *testing.T) {
	fe := &flowEngine{}

	ctx := &EngineContext{}

	nodeResp := &common.NodeResponse{
		RedirectURL: "https://example.com/auth",
	}

	flowStep := &FlowStep{
		Data: FlowData{},
	}

	err := fe.resolveStepForRedirection(ctx, nodeResp, flowStep)

	assert.NoError(t, err)
	assert.Equal(t, "https://example.com/auth", flowStep.Data.RedirectURL)
	assert.Nil(t, flowStep.Data.AdditionalData)
}

func TestResolveStepForRedirection_NilNodeResponse(t *testing.T) {
	fe := &flowEngine{}
	ctx := &EngineContext{}
	flowStep := &FlowStep{}

	err := fe.resolveStepForRedirection(ctx, nil, flowStep)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "node response is nil")
}

func TestResolveStepForRedirection_EmptyRedirectURL(t *testing.T) {
	fe := &flowEngine{}
	ctx := &EngineContext{}
	nodeResp := &common.NodeResponse{
		RedirectURL: "",
	}
	flowStep := &FlowStep{}

	err := fe.resolveStepForRedirection(ctx, nodeResp, flowStep)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redirect URL not found")
}

func TestResolveStepDetailsForPrompt_WithAdditionalData(t *testing.T) {
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

	assert.NoError(t, err)
	assert.NotNil(t, flowStep.Data.AdditionalData)
	assert.Equal(t, `{"rpId": "example.com"}`, flowStep.Data.AdditionalData["passkeyCreationOptions"])
}

func TestResolveStepDetailsForPrompt_WithActions(t *testing.T) {
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

	assert.NoError(t, err)
	assert.Len(t, flowStep.Data.Actions, 1)
	assert.Equal(t, "submit-action", flowStep.Data.Actions[0].Ref)
}

func TestResolveStepDetailsForPrompt_NilNodeResponse(t *testing.T) {
	fe := &flowEngine{}
	ctx := &EngineContext{}
	flowStep := &FlowStep{}

	err := fe.resolveStepDetailsForPrompt(ctx, nil, flowStep)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "node response is nil")
}

func TestResolveStepDetailsForPrompt_NoInputsOrActions(t *testing.T) {
	fe := &flowEngine{}
	ctx := &EngineContext{}
	nodeResp := &common.NodeResponse{}
	flowStep := &FlowStep{}

	err := fe.resolveStepDetailsForPrompt(ctx, nodeResp, flowStep)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no required data or actions found")
}

func TestUpdateContextWithNodeResponse_RuntimeData(t *testing.T) {
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

	assert.Equal(t, "value", ctx.RuntimeData["existing"])
	assert.Equal(t, "newValue", ctx.RuntimeData["newKey"])
}

func TestUpdateContextWithNodeResponse_RuntimeDataNilContext(t *testing.T) {
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

	assert.NotNil(t, ctx.RuntimeData)
	assert.Equal(t, "user-123", ctx.RuntimeData["userID"])
}

func TestUpdateContextWithNodeResponse_Assertion(t *testing.T) {
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

	assert.Equal(t, "test-assertion-token", ctx.Assertion)
}

func TestUpdateContextWithNodeResponse_AuthenticatedUserUpdate(t *testing.T) {
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

	assert.True(t, ctx.AuthenticatedUser.IsAuthenticated)
	assert.Equal(t, "user-123", ctx.AuthenticatedUser.UserID)
	assert.Equal(t, "user-123", ctx.RuntimeData["userID"])
}

func TestUpdateContextWithNodeResponse_MergesUserAttributes(t *testing.T) {
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

	assert.True(t, ctx.AuthenticatedUser.IsAuthenticated)
	assert.Equal(t, "existingValue", ctx.AuthenticatedUser.Attributes["existingAttr"])
	assert.Equal(t, "newValue", ctx.AuthenticatedUser.Attributes["newAttr"])
}

func TestUpdateContextWithNodeResponse_PreservesExistingUserID(t *testing.T) {
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
	assert.Equal(t, "existing-user-id", ctx.RuntimeData["userID"])
}

func TestUpdateContextWithNodeResponse_PreviousAttrsNilNewAttrs(t *testing.T) {
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
	assert.Equal(t, "prevValue", ctx.AuthenticatedUser.Attributes["prevAttr"])
}

func TestShouldUpdateAuthenticatedUser_NilNode(t *testing.T) {
	fe := &flowEngine{}
	ctx := &EngineContext{
		CurrentNode: nil,
	}

	result := fe.shouldUpdateAuthenticatedUser(ctx)

	assert.False(t, result)
}

func TestShouldUpdateAuthenticatedUser_NonTaskExecutionNode(t *testing.T) {
	mockNode := coremock.NewNodeInterfaceMock(t)
	mockNode.On("GetType").Return(common.NodeTypePrompt)

	fe := &flowEngine{}
	ctx := &EngineContext{
		CurrentNode: mockNode,
	}

	result := fe.shouldUpdateAuthenticatedUser(ctx)

	assert.False(t, result)
}

func TestShouldUpdateAuthenticatedUser_NonExecutorBackedNode(t *testing.T) {
	mockNode := coremock.NewNodeInterfaceMock(t)
	mockNode.On("GetType").Return(common.NodeTypeTaskExecution)

	fe := &flowEngine{}
	ctx := &EngineContext{
		CurrentNode: mockNode,
	}

	result := fe.shouldUpdateAuthenticatedUser(ctx)

	assert.False(t, result)
}

func TestShouldUpdateAuthenticatedUser_NilExecutor(t *testing.T) {
	mockNode := coremock.NewExecutorBackedNodeInterfaceMock(t)
	mockNode.On("GetType").Return(common.NodeTypeTaskExecution)
	mockNode.On("GetExecutor").Return(nil)

	fe := &flowEngine{}
	ctx := &EngineContext{
		CurrentNode: mockNode,
	}

	result := fe.shouldUpdateAuthenticatedUser(ctx)

	assert.False(t, result)
}

func TestShouldUpdateAuthenticatedUser_AuthFlowWithAuthExecutor(t *testing.T) {
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

	assert.True(t, result)
}

func TestShouldUpdateAuthenticatedUser_AuthFlowWithProvisioningExecutor(t *testing.T) {
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

	assert.True(t, result)
}

func TestShouldUpdateAuthenticatedUser_AuthFlowWithNonAuthExecutor(t *testing.T) {
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

	assert.False(t, result)
}

func TestShouldUpdateAuthenticatedUser_RegistrationFlowWithProvisioning(t *testing.T) {
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

	assert.True(t, result)
}

func TestShouldUpdateAuthenticatedUser_RegistrationFlowSkipProvisioning(t *testing.T) {
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

	assert.True(t, result)
}

func TestResolveStepForRedirection_WithInputs(t *testing.T) {
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

	assert.NoError(t, err)
	assert.Len(t, flowStep.Data.Inputs, 1)
	assert.Equal(t, "code", flowStep.Data.Inputs[0].Identifier)
	assert.Equal(t, common.FlowStatusIncomplete, flowStep.Status)
	assert.Equal(t, common.StepTypeRedirection, flowStep.Type)
}

func TestResolveStepForRedirection_AppendsInputs(t *testing.T) {
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

	assert.NoError(t, err)
	assert.Len(t, flowStep.Data.Inputs, 2)
}

func TestResolveStepDetailsForPrompt_WithMeta(t *testing.T) {
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

	assert.NoError(t, err)
	assert.NotNil(t, flowStep.Data.Meta)
}

func TestResolveStepDetailsForPrompt_WithFailureReason(t *testing.T) {
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

	assert.NoError(t, err)
	assert.Equal(t, "Invalid OTP provided", flowStep.FailureReason)
	assert.Equal(t, common.FlowStatusIncomplete, flowStep.Status)
	assert.Equal(t, common.StepTypeView, flowStep.Type)
}

func TestResolveStepDetailsForPrompt_AppendsInputs(t *testing.T) {
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

	assert.NoError(t, err)
	assert.Len(t, flowStep.Data.Inputs, 2)
}

func TestResolveStepDetailsForPrompt_ExistingActions(t *testing.T) {
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

	assert.NoError(t, err)
	// Actions are replaced, not appended
	assert.Len(t, flowStep.Data.Actions, 1)
	assert.Equal(t, "submit-action", flowStep.Data.Actions[0].Ref)
}

func TestGetNodeInputs_PromptNodeEmptyInputs(t *testing.T) {
	mockNode := coremock.NewPromptNodeInterfaceMock(t)
	prompts := []common.Prompt{
		{
			Inputs: []common.Input{},
		},
	}
	mockNode.On("GetPrompts").Return(prompts)

	inputs := getNodeInputs(mockNode)

	assert.Nil(t, inputs)
}
