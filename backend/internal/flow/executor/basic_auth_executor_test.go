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

	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/user"
	"github.com/asgardeo/thunder/tests/mocks/authn/credentialsmock"
	"github.com/asgardeo/thunder/tests/mocks/flow/coremock"
	"github.com/asgardeo/thunder/tests/mocks/usermock"
)

type BasicAuthExecutorTestSuite struct {
	suite.Suite
	mockUserService  *usermock.UserServiceInterfaceMock
	mockCredsService *credentialsmock.CredentialsAuthnServiceInterfaceMock
	mockFlowFactory  *coremock.FlowFactoryInterfaceMock
	executor         *basicAuthExecutor
}

func TestBasicAuthExecutorSuite(t *testing.T) {
	suite.Run(t, new(BasicAuthExecutorTestSuite))
}

func (suite *BasicAuthExecutorTestSuite) SetupTest() {
	suite.mockUserService = usermock.NewUserServiceInterfaceMock(suite.T())
	suite.mockCredsService = credentialsmock.NewCredentialsAuthnServiceInterfaceMock(suite.T())
	suite.mockFlowFactory = coremock.NewFlowFactoryInterfaceMock(suite.T())

	defaultInputs := []common.Input{
		{Identifier: userAttributeUsername, Type: "string", Required: true},
		{Identifier: userAttributePassword, Type: "string", Required: true},
	}

	// Mock the embedded identifying executor first
	identifyingMock := createMockIdentifyingExecutor(suite.T())
	suite.mockFlowFactory.On("CreateExecutor", ExecutorNameIdentifying, common.ExecutorTypeUtility,
		mock.Anything, mock.Anything).Return(identifyingMock).Maybe()

	mockExec := createMockBasicAuthExecutor(suite.T())
	suite.mockFlowFactory.On("CreateExecutor", ExecutorNameBasicAuth, common.ExecutorTypeAuthentication,
		defaultInputs, []common.Input{}).Return(mockExec)

	suite.executor = newBasicAuthExecutor(suite.mockFlowFactory, suite.mockUserService, suite.mockCredsService)
}

func createMockIdentifyingExecutor(t *testing.T) core.ExecutorInterface {
	mockExec := coremock.NewExecutorInterfaceMock(t)
	mockExec.On("GetName").Return(ExecutorNameIdentifying).Maybe()
	mockExec.On("GetType").Return(common.ExecutorTypeUtility).Maybe()
	mockExec.On("GetDefaultInputs").Return([]common.Input{}).Maybe()
	mockExec.On("GetPrerequisites").Return([]common.Input{}).Maybe()
	return mockExec
}

func createMockBasicAuthExecutor(t *testing.T) core.ExecutorInterface {
	mockExec := coremock.NewExecutorInterfaceMock(t)
	mockExec.On("GetName").Return(ExecutorNameBasicAuth).Maybe()
	mockExec.On("GetType").Return(common.ExecutorTypeAuthentication).Maybe()
	mockExec.On("GetDefaultInputs").Return([]common.Input{
		{Identifier: userAttributeUsername, Type: "string", Required: true},
		{Identifier: userAttributePassword, Type: "string", Required: true},
	}).Maybe()
	mockExec.On("GetPrerequisites").Return([]common.Input{}).Maybe()
	mockExec.On("HasRequiredInputs", mock.Anything, mock.Anything).Return(
		func(ctx *core.NodeContext, execResp *common.ExecutorResponse) bool {
			username, hasUsername := ctx.UserInputs[userAttributeUsername]
			password, hasPassword := ctx.UserInputs[userAttributePassword]
			if !hasUsername || username == "" || !hasPassword || password == "" {
				execResp.Inputs = []common.Input{
					{Identifier: userAttributeUsername, Type: "string", Required: true},
					{Identifier: userAttributePassword, Type: "string", Required: true},
				}
				return false
			}
			return true
		}).Maybe()
	return mockExec
}

func (suite *BasicAuthExecutorTestSuite) TestNewBasicAuthExecutor() {
	assert.NotNil(suite.T(), suite.executor)
	assert.NotNil(suite.T(), suite.executor.credsAuthSvc)
}

func (suite *BasicAuthExecutorTestSuite) TestExecute_Success_AuthenticationFlow() {
	attrs := map[string]interface{}{"email": "test@example.com"}
	attrsJSON, _ := json.Marshal(attrs)

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			userAttributeUsername: "testuser",
			userAttributePassword: "password123",
		},
		RuntimeData: make(map[string]string),
	}

	userID := testUserID
	suite.mockUserService.On("IdentifyUser", map[string]interface{}{
		userAttributeUsername: "testuser",
	}).Return(&userID, nil)

	authenticatedUser := &user.User{
		ID:               testUserID,
		OrganizationUnit: "ou-123",
		Type:             "INTERNAL",
		Attributes:       attrsJSON,
	}

	suite.mockCredsService.On("Authenticate", map[string]interface{}{
		userAttributeUsername: "testuser",
		userAttributePassword: "password123",
	}).Return(authenticatedUser, nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	assert.True(suite.T(), resp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), testUserID, resp.AuthenticatedUser.UserID)
	suite.mockUserService.AssertExpectations(suite.T())
	suite.mockCredsService.AssertExpectations(suite.T())
}

func (suite *BasicAuthExecutorTestSuite) TestExecute_Success_RegistrationFlow() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeRegistration,
		UserInputs: map[string]string{
			userAttributeUsername: "newuser",
			userAttributePassword: "password123",
		},
		RuntimeData: make(map[string]string),
	}

	suite.mockUserService.On("IdentifyUser", map[string]interface{}{
		userAttributeUsername: "newuser",
	}).Return(nil, &user.ErrorUserNotFound)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	assert.False(suite.T(), resp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), "newuser", resp.AuthenticatedUser.Attributes[userAttributeUsername])
	suite.mockUserService.AssertExpectations(suite.T())
}

func (suite *BasicAuthExecutorTestSuite) TestExecute_UserInputRequired() {
	ctx := &core.NodeContext{
		FlowID:      "flow-123",
		FlowType:    common.FlowTypeAuthentication,
		UserInputs:  map[string]string{},
		RuntimeData: make(map[string]string),
	}

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecUserInputRequired, resp.Status)
	assert.NotEmpty(suite.T(), resp.Inputs)
}

func (suite *BasicAuthExecutorTestSuite) TestExecute_AuthenticationFailed() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			userAttributeUsername: "testuser",
			userAttributePassword: "wrongpassword",
		},
		RuntimeData: make(map[string]string),
	}

	userID := testUserID
	suite.mockUserService.On("IdentifyUser", map[string]interface{}{
		userAttributeUsername: "testuser",
	}).Return(&userID, nil)

	suite.mockCredsService.On("Authenticate", map[string]interface{}{
		userAttributeUsername: "testuser",
		userAttributePassword: "wrongpassword",
	}).Return(nil, &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		ErrorDescription: "Invalid credentials",
	})

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Contains(suite.T(), resp.FailureReason, "Failed to authenticate user")
	suite.mockUserService.AssertExpectations(suite.T())
	suite.mockCredsService.AssertExpectations(suite.T())
}

func (suite *BasicAuthExecutorTestSuite) TestExecute_UserNotFound_AuthenticationFlow() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			userAttributeUsername: "nonexistent",
			userAttributePassword: "password123",
		},
		RuntimeData: make(map[string]string),
	}

	suite.mockUserService.On("IdentifyUser", map[string]interface{}{
		userAttributeUsername: "nonexistent",
	}).Return(nil, &user.ErrorUserNotFound)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	suite.mockUserService.AssertExpectations(suite.T())
}

func (suite *BasicAuthExecutorTestSuite) TestExecute_UserAlreadyExists_RegistrationFlow() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeRegistration,
		UserInputs: map[string]string{
			userAttributeUsername: "existinguser",
			userAttributePassword: "password123",
		},
		RuntimeData: make(map[string]string),
	}

	userID := testUserID
	suite.mockUserService.On("IdentifyUser", map[string]interface{}{
		userAttributeUsername: "existinguser",
	}).Return(&userID, nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Contains(suite.T(), resp.FailureReason, "User already exists")
	suite.mockUserService.AssertExpectations(suite.T())
}

func (suite *BasicAuthExecutorTestSuite) TestExecute_ServiceError() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			userAttributeUsername: "testuser",
			userAttributePassword: "password123",
		},
		RuntimeData: make(map[string]string),
	}

	suite.mockUserService.On("IdentifyUser", map[string]interface{}{
		userAttributeUsername: "testuser",
	}).Return(nil, &serviceerror.ServiceError{Error: "database error"})

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	suite.mockUserService.AssertExpectations(suite.T())
}

func (suite *BasicAuthExecutorTestSuite) TestExecute_AuthenticationServiceError() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			userAttributeUsername: "testuser",
			userAttributePassword: "password123",
		},
		RuntimeData: make(map[string]string),
	}

	userID := testUserID
	suite.mockUserService.On("IdentifyUser", map[string]interface{}{
		userAttributeUsername: "testuser",
	}).Return(&userID, nil)

	suite.mockCredsService.On("Authenticate", mock.Anything).
		Return(nil, &serviceerror.ServiceError{
			Type:  serviceerror.ServerErrorType,
			Error: "internal server error",
		})

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Contains(suite.T(), resp.FailureReason, "Failed to authenticate user")
	suite.mockUserService.AssertExpectations(suite.T())
	suite.mockCredsService.AssertExpectations(suite.T())
}

func (suite *BasicAuthExecutorTestSuite) TestGetAuthenticatedUser_SuccessfulAuthentication() {
	attrs := map[string]interface{}{"email": "test@example.com", "phone": "1234567890"}
	attrsJSON, _ := json.Marshal(attrs)

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			userAttributeUsername: "testuser",
			userAttributePassword: "password123",
		},
	}

	execResp := &common.ExecutorResponse{
		RuntimeData: make(map[string]string),
	}

	userID := testUserID
	suite.mockUserService.On("IdentifyUser", map[string]interface{}{
		userAttributeUsername: "testuser",
	}).Return(&userID, nil)

	authenticatedUser := &user.User{
		ID:               testUserID,
		OrganizationUnit: "ou-123",
		Type:             "INTERNAL",
		Attributes:       attrsJSON,
	}

	suite.mockCredsService.On("Authenticate", map[string]interface{}{
		userAttributeUsername: "testuser",
		userAttributePassword: "password123",
	}).Return(authenticatedUser, nil)

	result, err := suite.executor.getAuthenticatedUser(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.True(suite.T(), result.IsAuthenticated)
	assert.Equal(suite.T(), testUserID, result.UserID)
	assert.Equal(suite.T(), "ou-123", result.OrganizationUnitID)
	assert.Equal(suite.T(), "INTERNAL", result.UserType)
	assert.Equal(suite.T(), "test@example.com", result.Attributes["email"])
	suite.mockUserService.AssertExpectations(suite.T())
	suite.mockCredsService.AssertExpectations(suite.T())
}

func (suite *BasicAuthExecutorTestSuite) TestGetAuthenticatedUser_InvalidJSONAttributes() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			userAttributeUsername: "testuser",
			userAttributePassword: "password123",
		},
	}

	execResp := &common.ExecutorResponse{
		RuntimeData: make(map[string]string),
	}

	userID := testUserID
	suite.mockUserService.On("IdentifyUser", map[string]interface{}{
		userAttributeUsername: "testuser",
	}).Return(&userID, nil)

	authenticatedUser := &user.User{
		ID:               testUserID,
		OrganizationUnit: "ou-123",
		Type:             "INTERNAL",
		Attributes:       json.RawMessage(`invalid json`),
	}

	suite.mockCredsService.On("Authenticate", mock.Anything).Return(authenticatedUser, nil)

	resp, err := suite.executor.getAuthenticatedUser(ctx, execResp)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), resp)
	suite.mockUserService.AssertExpectations(suite.T())
	suite.mockCredsService.AssertExpectations(suite.T())
}
