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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/tests/mocks/flow/coremock"
)

type InviteExecutorTestSuite struct {
	suite.Suite
	mockFlowFactory *coremock.FlowFactoryInterfaceMock
	executor        *inviteExecutor
}

func (suite *InviteExecutorTestSuite) SetupTest() {
	// Initialize runtime config for tests
	err := config.InitializeThunderRuntime(".", &config.Config{
		GateClient: config.GateClientConfig{
			Scheme:   "https",
			Hostname: "localhost",
			Port:     5190,
			Path:     "/gate",
		},
	})
	suite.Require().NoError(err)

	suite.mockFlowFactory = coremock.NewFlowFactoryInterfaceMock(suite.T())
	mockBaseExecutor := coremock.NewExecutorInterfaceMock(suite.T())

	// Set up expectations for CreateExecutor (called in constructor)
	suite.mockFlowFactory.On("CreateExecutor",
		ExecutorNameInviteExecutor,
		common.ExecutorTypeUtility,
		[]common.Input{
			{
				Identifier: userInputInviteToken,
				Type:       "HIDDEN",
				Required:   true,
			},
		},
		[]common.Input{}).Return(mockBaseExecutor)

	suite.executor = newInviteExecutor(suite.mockFlowFactory)
}

func (suite *InviteExecutorTestSuite) TearDownTest() {
	config.ResetThunderRuntime()
}

func (suite *InviteExecutorTestSuite) TestExecute_GenerateToken_AdminPhase() {
	ctx := &core.NodeContext{
		FlowID:      "test-flow-id",
		UserInputs:  make(map[string]string),
		RuntimeData: make(map[string]string),
	}

	mockExecutor := suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock)
	mockExecutor.On("HasRequiredInputs", ctx, mock.Anything).Return(false)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecUserInputRequired, resp.Status)
	assert.NotEmpty(suite.T(), resp.RuntimeData[runtimeKeyStoredInviteToken])
	assert.Contains(suite.T(), resp.AdditionalData["inviteLink"], "inviteToken=")
	assert.Contains(suite.T(), resp.AdditionalData["inviteLink"], "flowId=test-flow-id")
}

func (suite *InviteExecutorTestSuite) TestExecute_Idempotency_AdminRetry() {
	existingToken := "existing-token-123"
	ctx := &core.NodeContext{
		FlowID:     "test-flow-id",
		UserInputs: make(map[string]string),
		RuntimeData: map[string]string{
			runtimeKeyStoredInviteToken: existingToken,
		},
	}

	mockExecutor := suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock)
	mockExecutor.On("HasRequiredInputs", ctx, mock.Anything).Return(false)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecUserInputRequired, resp.Status)
	assert.Equal(suite.T(), existingToken, resp.RuntimeData[runtimeKeyStoredInviteToken])
	assert.Contains(suite.T(), resp.AdditionalData["inviteLink"], existingToken)
}

func (suite *InviteExecutorTestSuite) TestExecute_ValidationSuccess_UserPhase() {
	token := "valid-token"
	ctx := &core.NodeContext{
		FlowID: "test-flow-id",
		UserInputs: map[string]string{
			userInputInviteToken: token,
		},
		RuntimeData: map[string]string{
			runtimeKeyStoredInviteToken: token,
		},
	}

	mockExecutor := suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock)
	mockExecutor.On("HasRequiredInputs", ctx, mock.Anything).Return(true)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
}

func (suite *InviteExecutorTestSuite) TestExecute_ValidationFailure_Mismatch() {
	ctx := &core.NodeContext{
		FlowID: "test-flow-id",
		UserInputs: map[string]string{
			userInputInviteToken: "wrong-token",
		},
		RuntimeData: map[string]string{
			runtimeKeyStoredInviteToken: "correct-token",
		},
	}

	mockExecutor := suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock)
	mockExecutor.On("HasRequiredInputs", ctx, mock.Anything).Return(true)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Equal(suite.T(), "Invalid invite token", resp.FailureReason)
}

func (suite *InviteExecutorTestSuite) TestExecute_ValidationFailure_NoStoredToken() {
	ctx := &core.NodeContext{
		FlowID: "test-flow-id",
		UserInputs: map[string]string{
			userInputInviteToken: "some-token",
		},
		RuntimeData: make(map[string]string),
	}

	mockExecutor := suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock)
	mockExecutor.On("HasRequiredInputs", ctx, mock.Anything).Return(true)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Equal(suite.T(), "Invalid invite token", resp.FailureReason)
}

func TestInviteExecutorSuite(t *testing.T) {
	suite.Run(t, new(InviteExecutorTestSuite))
}
