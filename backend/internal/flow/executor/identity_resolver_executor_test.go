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
	"github.com/asgardeo/thunder/internal/user"
	"github.com/asgardeo/thunder/tests/mocks/flow/coremock"
	"github.com/asgardeo/thunder/tests/mocks/usermock"
)

const (
	testIdentityResolverFlowID   = "identity-resolver-flow-123"
	testIdentityResolverUsername = "testuser"
	testIdentityResolverUserID   = "user-id-456"
)

type IdentityResolverExecutorTestSuite struct {
	suite.Suite
	mockUserService *usermock.UserServiceInterfaceMock
	mockFlowFactory *coremock.FlowFactoryInterfaceMock
	executor        *identityResolverExecutor
}

func TestIdentityResolverExecutorSuite(t *testing.T) {
	suite.Run(t, new(IdentityResolverExecutorTestSuite))
}

func (suite *IdentityResolverExecutorTestSuite) SetupTest() {
	suite.mockUserService = usermock.NewUserServiceInterfaceMock(suite.T())
	suite.mockFlowFactory = coremock.NewFlowFactoryInterfaceMock(suite.T())

	// Create mock identifying executor
	identifyingMock := coremock.NewExecutorInterfaceMock(suite.T())
	identifyingMock.On("GetName").Return(ExecutorNameIdentifying).Maybe()
	identifyingMock.On("GetType").Return(common.ExecutorTypeUtility).Maybe()
	identifyingMock.On("GetDefaultInputs").Return([]common.Input{}).Maybe()
	identifyingMock.On("GetPrerequisites").Return([]common.Input{}).Maybe()
	suite.mockFlowFactory.On("CreateExecutor", ExecutorNameIdentifying, common.ExecutorTypeUtility,
		mock.Anything, mock.Anything).Return(identifyingMock).Maybe()

	// Create mock base executor for identity resolver
	mockExec := coremock.NewExecutorInterfaceMock(suite.T())
	mockExec.On("GetName").Return(ExecutorNameIdentityResolver).Maybe()
	mockExec.On("GetType").Return(common.ExecutorTypeUtility).Maybe()
	mockExec.On("GetDefaultInputs").Return([]common.Input{
		{Identifier: userAttributeUsername, Type: "string", Required: true},
	}).Maybe()
	mockExec.On("GetPrerequisites").Return([]common.Input{}).Maybe()
	mockExec.On("GetRequiredInputs", mock.Anything).Return([]common.Input{
		{Identifier: userAttributeUsername, Type: "string", Required: true},
	}).Maybe()
	mockExec.On("HasRequiredInputs", mock.Anything, mock.Anything).Return(
		func(ctx *core.NodeContext, execResp *common.ExecutorResponse) bool {
			if _, exists := ctx.UserInputs[userAttributeUsername]; exists {
				return true
			}
			if _, exists := ctx.RuntimeData[userAttributeUsername]; exists {
				return true
			}
			execResp.Status = common.ExecUserInputRequired
			execResp.Inputs = []common.Input{
				{Identifier: userAttributeUsername, Type: "string", Required: true},
			}
			return false
		}).Maybe()
	mockExec.On("ValidatePrerequisites", mock.Anything, mock.Anything).Return(true).Maybe()

	suite.mockFlowFactory.On("CreateExecutor", ExecutorNameIdentityResolver, common.ExecutorTypeUtility,
		mock.Anything, mock.Anything).Return(mockExec)

	suite.executor = newIdentityResolverExecutor(suite.mockFlowFactory, suite.mockUserService)
}

// Helper to create a node context for identity resolver
func createIdentityResolverNodeContext() *core.NodeContext {
	return &core.NodeContext{
		FlowID:      testIdentityResolverFlowID,
		FlowType:    common.FlowTypeAuthentication,
		UserInputs:  make(map[string]string),
		RuntimeData: make(map[string]string),
	}
}

func (suite *IdentityResolverExecutorTestSuite) TestNewIdentityResolverExecutor() {
	assert.NotNil(suite.T(), suite.executor)
	assert.NotNil(suite.T(), suite.executor.userService)
	assert.NotNil(suite.T(), suite.executor.identifyingExecutorInterface)
}

func (suite *IdentityResolverExecutorTestSuite) TestExecute_Success_UsernameInUserInputs() {
	ctx := createIdentityResolverNodeContext()
	ctx.UserInputs[userAttributeUsername] = testIdentityResolverUsername

	userID := testIdentityResolverUserID
	suite.mockUserService.On("IdentifyUser", mock.Anything, map[string]interface{}{
		userAttributeUsername: testIdentityResolverUsername,
	}).Return(&userID, nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	assert.Equal(suite.T(), testIdentityResolverUserID, resp.RuntimeData[userAttributeUserID])
}

func (suite *IdentityResolverExecutorTestSuite) TestExecute_Success_UsernameInRuntimeData() {
	ctx := createIdentityResolverNodeContext()
	ctx.RuntimeData[userAttributeUsername] = testIdentityResolverUsername

	userID := testIdentityResolverUserID
	suite.mockUserService.On("IdentifyUser", mock.Anything, map[string]interface{}{
		userAttributeUsername: testIdentityResolverUsername,
	}).Return(&userID, nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	assert.Equal(suite.T(), testIdentityResolverUserID, resp.RuntimeData[userAttributeUserID])
}

func (suite *IdentityResolverExecutorTestSuite) TestExecute_UserInputRequired_NoUsername() {
	ctx := createIdentityResolverNodeContext()
	// No username provided in UserInputs or RuntimeData

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecUserInputRequired, resp.Status)
}

func (suite *IdentityResolverExecutorTestSuite) TestExecute_UserInputRequired_EmptyUsername() {
	ctx := createIdentityResolverNodeContext()
	ctx.UserInputs[userAttributeUsername] = "" // Empty username

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecUserInputRequired, resp.Status)
}

func (suite *IdentityResolverExecutorTestSuite) TestExecute_Failure_UserNotFound() {
	ctx := createIdentityResolverNodeContext()
	ctx.UserInputs[userAttributeUsername] = "nonexistent_user"

	suite.mockUserService.On("IdentifyUser", mock.Anything, map[string]interface{}{
		userAttributeUsername: "nonexistent_user",
	}).Return(nil, &user.ErrorUserNotFound)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Contains(suite.T(), resp.FailureReason, failureReasonUserNotFound)
}

func (suite *IdentityResolverExecutorTestSuite) TestExecute_Failure_NilUserID() {
	ctx := createIdentityResolverNodeContext()
	ctx.UserInputs[userAttributeUsername] = testIdentityResolverUsername

	// Return nil userID (edge case)
	suite.mockUserService.On("IdentifyUser", mock.Anything, map[string]interface{}{
		userAttributeUsername: testIdentityResolverUsername,
	}).Return(nil, nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Contains(suite.T(), resp.FailureReason, failureReasonUserNotFound)
}

func (suite *IdentityResolverExecutorTestSuite) TestExecute_Failure_EmptyUserID() {
	ctx := createIdentityResolverNodeContext()
	ctx.UserInputs[userAttributeUsername] = testIdentityResolverUsername

	// Return empty userID
	emptyUserID := ""
	suite.mockUserService.On("IdentifyUser", mock.Anything, map[string]interface{}{
		userAttributeUsername: testIdentityResolverUsername,
	}).Return(&emptyUserID, nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Contains(suite.T(), resp.FailureReason, failureReasonUserNotFound)
}

func (suite *IdentityResolverExecutorTestSuite) TestExecute_Failure_ServiceError() {
	ctx := createIdentityResolverNodeContext()
	ctx.UserInputs[userAttributeUsername] = testIdentityResolverUsername

	suite.mockUserService.On("IdentifyUser", mock.Anything, map[string]interface{}{
		userAttributeUsername: testIdentityResolverUsername,
	}).Return(nil, &user.ErrorInternalServerError)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
}

func (suite *IdentityResolverExecutorTestSuite) TestExecute_PreservesExistingRuntimeData() {
	ctx := createIdentityResolverNodeContext()
	ctx.UserInputs[userAttributeUsername] = testIdentityResolverUsername
	ctx.RuntimeData["existingKey"] = "existingValue"

	userID := testIdentityResolverUserID
	suite.mockUserService.On("IdentifyUser", mock.Anything, map[string]interface{}{
		userAttributeUsername: testIdentityResolverUsername,
	}).Return(&userID, nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	// The resolved userID should be in RuntimeData
	assert.Equal(suite.T(), testIdentityResolverUserID, resp.RuntimeData[userAttributeUserID])
}

func (suite *IdentityResolverExecutorTestSuite) TestExecute_UserInputsPrioritizedOverRuntimeData() {
	ctx := createIdentityResolverNodeContext()
	// Set username in both places - UserInputs should take priority
	ctx.UserInputs[userAttributeUsername] = "user_from_inputs"
	ctx.RuntimeData[userAttributeUsername] = "user_from_runtime"

	userID := testIdentityResolverUserID
	suite.mockUserService.On("IdentifyUser", mock.Anything, map[string]interface{}{
		userAttributeUsername: "user_from_inputs",
	}).Return(&userID, nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
}
