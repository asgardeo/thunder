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
	"github.com/asgardeo/thunder/internal/observability/event"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/user"
	"github.com/asgardeo/thunder/tests/mocks/authn/credentialsmock"
	"github.com/asgardeo/thunder/tests/mocks/flow/coremock"
	"github.com/asgardeo/thunder/tests/mocks/observabilitymock"
	"github.com/asgardeo/thunder/tests/mocks/usermock"
)

type BasicAuthExecutorTestSuite struct {
	suite.Suite
	mockUserService   *usermock.UserServiceInterfaceMock
	mockCredsService  *credentialsmock.CredentialsAuthnServiceInterfaceMock
	mockFlowFactory   *coremock.FlowFactoryInterfaceMock
	mockObservability *observabilitymock.ObservabilityServiceInterfaceMock
	executor          *basicAuthExecutor
}

func TestBasicAuthExecutorSuite(t *testing.T) {
	suite.Run(t, new(BasicAuthExecutorTestSuite))
}

func (suite *BasicAuthExecutorTestSuite) SetupTest() {
	suite.mockUserService = usermock.NewUserServiceInterfaceMock(suite.T())
	suite.mockCredsService = credentialsmock.NewCredentialsAuthnServiceInterfaceMock(suite.T())
	suite.mockFlowFactory = coremock.NewFlowFactoryInterfaceMock(suite.T())
	suite.mockObservability = observabilitymock.NewObservabilityServiceInterfaceMock(suite.T())

	// Default behavior for observability: disabled
	suite.mockObservability.On("IsEnabled").Return(false).Maybe()

	defaultInputs := []flowcm.InputData{
		{Name: userAttributeUsername, Type: "string", Required: true},
		{Name: userAttributePassword, Type: inputDataTypePassword, Required: true},
	}

	// Mock the embedded identifying executor first
	identifyingMock := createMockIdentifyingExecutor(suite.T())
	suite.mockFlowFactory.On("CreateExecutor", ExecutorNameIdentifying, flowcm.ExecutorTypeUtility,
		mock.Anything, mock.Anything).Return(identifyingMock).Maybe()

	mockExec := createMockBasicAuthExecutor(suite.T())
	suite.mockFlowFactory.On("CreateExecutor", ExecutorNameBasicAuth, flowcm.ExecutorTypeAuthentication,
		defaultInputs, []flowcm.InputData{}).Return(mockExec)

	suite.executor = newBasicAuthExecutor(suite.mockFlowFactory, suite.mockUserService, suite.mockCredsService,
		suite.mockObservability)
}

func createMockIdentifyingExecutor(t *testing.T) flowcore.ExecutorInterface {
	mockExec := coremock.NewExecutorInterfaceMock(t)
	mockExec.On("GetName").Return(ExecutorNameIdentifying).Maybe()
	mockExec.On("GetType").Return(flowcm.ExecutorTypeUtility).Maybe()
	mockExec.On("GetDefaultExecutorInputs").Return([]flowcm.InputData{}).Maybe()
	mockExec.On("GetPrerequisites").Return([]flowcm.InputData{}).Maybe()
	return mockExec
}

func createMockExecutorWithCustomInputs(t *testing.T, name string,
	inputs []flowcm.InputData) flowcore.ExecutorInterface {
	mockExec := coremock.NewExecutorInterfaceMock(t)
	mockExec.On("GetName").Return(name).Maybe()
	mockExec.On("GetType").Return(flowcm.ExecutorTypeAuthentication).Maybe()
	mockExec.On("GetDefaultExecutorInputs").Return(inputs).Maybe()
	mockExec.On("GetPrerequisites").Return([]flowcm.InputData{}).Maybe()
	mockExec.On("GetRequiredData", mock.Anything).Return(
		func(ctx *flowcore.NodeContext) []flowcm.InputData {
			return inputs
		}).Maybe()
	mockExec.On("CheckInputData", mock.Anything, mock.Anything).Return(
		func(ctx *flowcore.NodeContext, execResp *flowcm.ExecutorResponse) bool {
			for _, input := range inputs {
				if input.Required {
					value, exists := ctx.UserInputData[input.Name]
					if !exists || value == "" {
						execResp.RequiredData = inputs
						return true
					}
				}
			}
			return false
		}).Maybe()
	return mockExec
}

func createMockBasicAuthExecutor(t *testing.T) flowcore.ExecutorInterface {
	mockExec := coremock.NewExecutorInterfaceMock(t)
	mockExec.On("GetName").Return(ExecutorNameBasicAuth).Maybe()
	mockExec.On("GetType").Return(flowcm.ExecutorTypeAuthentication).Maybe()
	mockExec.On("GetDefaultExecutorInputs").Return([]flowcm.InputData{
		{Name: userAttributeUsername, Type: "string", Required: true},
		{Name: userAttributePassword, Type: inputDataTypePassword, Required: true},
	}).Maybe()
	mockExec.On("GetPrerequisites").Return([]flowcm.InputData{}).Maybe()
	mockExec.On("GetRequiredData", mock.Anything).Return(
		func(ctx *flowcore.NodeContext) []flowcm.InputData {
			return []flowcm.InputData{
				{Name: userAttributeUsername, Type: "string", Required: true},
				{Name: userAttributePassword, Type: inputDataTypePassword, Required: true},
			}
		}).Maybe()
	mockExec.On("CheckInputData", mock.Anything, mock.Anything).Return(
		func(ctx *flowcore.NodeContext, execResp *flowcm.ExecutorResponse) bool {
			username, hasUsername := ctx.UserInputData[userAttributeUsername]
			password, hasPassword := ctx.UserInputData[userAttributePassword]
			if !hasUsername || username == "" || !hasPassword || password == "" {
				execResp.RequiredData = []flowcm.InputData{
					{Name: userAttributeUsername, Type: "string", Required: true},
					{Name: userAttributePassword, Type: inputDataTypePassword, Required: true},
				}
				return true
			}
			return false
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

	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		UserInputData: map[string]string{
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
	assert.Equal(suite.T(), flowcm.ExecComplete, resp.Status)
	assert.True(suite.T(), resp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), testUserID, resp.AuthenticatedUser.UserID)
	suite.mockUserService.AssertExpectations(suite.T())
	suite.mockCredsService.AssertExpectations(suite.T())
}

func (suite *BasicAuthExecutorTestSuite) TestExecute_Success_WithEmailAttribute() {
	attrs := map[string]interface{}{"phone": "+1234567890"}
	attrsJSON, _ := json.Marshal(attrs)

	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		UserInputData: map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		},
		RuntimeData: make(map[string]string),
	}

	// Override GetRequiredData to return email and password as required fields
	originalInputs := []flowcm.InputData{
		{Name: "email", Type: "string", Required: true},
		{Name: "password", Type: inputDataTypePassword, Required: true},
	}
	suite.executor.ExecutorInterface = createMockExecutorWithCustomInputs(
		suite.T(), ExecutorNameBasicAuth, originalInputs)

	userID := testUserID
	suite.mockUserService.On("IdentifyUser", map[string]interface{}{
		"email": "test@example.com",
	}).Return(&userID, nil)

	authenticatedUser := &user.User{
		ID:               testUserID,
		OrganizationUnit: "ou-123",
		Type:             "INTERNAL",
		Attributes:       attrsJSON,
	}

	suite.mockCredsService.On("Authenticate", map[string]interface{}{
		"email":    "test@example.com",
		"password": "password123",
	}).Return(authenticatedUser, nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), flowcm.ExecComplete, resp.Status)
	assert.True(suite.T(), resp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), testUserID, resp.AuthenticatedUser.UserID)
	suite.mockUserService.AssertExpectations(suite.T())
	suite.mockCredsService.AssertExpectations(suite.T())
}

func (suite *BasicAuthExecutorTestSuite) TestExecute_Success_RegistrationFlow() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeRegistration,
		UserInputData: map[string]string{
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
	assert.Equal(suite.T(), flowcm.ExecComplete, resp.Status)
	assert.False(suite.T(), resp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), "newuser", resp.AuthenticatedUser.Attributes[userAttributeUsername])
	suite.mockUserService.AssertExpectations(suite.T())
}

func (suite *BasicAuthExecutorTestSuite) TestExecute_Success_WithMultipleAttributes() {
	attrs := map[string]interface{}{"name": "Test User", "role": "admin"}
	attrsJSON, _ := json.Marshal(attrs)

	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		UserInputData: map[string]string{
			"email":    "test@example.com",
			"phone":    "+1234567890",
			"password": "password123",
		},
		RuntimeData: make(map[string]string),
	}

	// Override GetRequiredData to return email, phone, and password as required fields
	customInputs := []flowcm.InputData{
		{Name: "email", Type: "string", Required: true},
		{Name: "phone", Type: "string", Required: true},
		{Name: "password", Type: inputDataTypePassword, Required: true},
	}
	suite.executor.ExecutorInterface = createMockExecutorWithCustomInputs(
		suite.T(), ExecutorNameBasicAuth, customInputs)

	userID := testUserID
	suite.mockUserService.On("IdentifyUser", map[string]interface{}{
		"email": "test@example.com",
		"phone": "+1234567890",
	}).Return(&userID, nil)

	authenticatedUser := &user.User{
		ID:               testUserID,
		OrganizationUnit: "ou-123",
		Type:             "INTERNAL",
		Attributes:       attrsJSON,
	}

	suite.mockCredsService.On("Authenticate", map[string]interface{}{
		"email":    "test@example.com",
		"phone":    "+1234567890",
		"password": "password123",
	}).Return(authenticatedUser, nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), flowcm.ExecComplete, resp.Status)
	assert.True(suite.T(), resp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), testUserID, resp.AuthenticatedUser.UserID)
	suite.mockUserService.AssertExpectations(suite.T())
	suite.mockCredsService.AssertExpectations(suite.T())
}

func (suite *BasicAuthExecutorTestSuite) TestExecute_UserInputRequired() {
	ctx := &flowcore.NodeContext{
		FlowID:        "flow-123",
		FlowType:      flowcm.FlowTypeAuthentication,
		UserInputData: map[string]string{},
		RuntimeData:   make(map[string]string),
	}

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), flowcm.ExecUserInputRequired, resp.Status)
	assert.NotEmpty(suite.T(), resp.RequiredData)
}

func (suite *BasicAuthExecutorTestSuite) TestExecute_AuthenticationFailed() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		UserInputData: map[string]string{
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
	assert.Equal(suite.T(), flowcm.ExecFailure, resp.Status)
	assert.Contains(suite.T(), resp.FailureReason, "Failed to authenticate user")
	suite.mockUserService.AssertExpectations(suite.T())
	suite.mockCredsService.AssertExpectations(suite.T())
}

func (suite *BasicAuthExecutorTestSuite) TestExecute_UserNotFound_AuthenticationFlow() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		UserInputData: map[string]string{
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
	assert.Equal(suite.T(), flowcm.ExecFailure, resp.Status)
	suite.mockUserService.AssertExpectations(suite.T())
}

func (suite *BasicAuthExecutorTestSuite) TestExecute_UserAlreadyExists_RegistrationFlow() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeRegistration,
		UserInputData: map[string]string{
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
	assert.Equal(suite.T(), flowcm.ExecFailure, resp.Status)
	assert.Contains(suite.T(), resp.FailureReason, "User already exists")
	suite.mockUserService.AssertExpectations(suite.T())
}

func (suite *BasicAuthExecutorTestSuite) TestExecute_ServiceError() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		UserInputData: map[string]string{
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
	assert.Equal(suite.T(), flowcm.ExecFailure, resp.Status)
	suite.mockUserService.AssertExpectations(suite.T())
}

func (suite *BasicAuthExecutorTestSuite) TestExecute_AuthenticationServiceError() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		UserInputData: map[string]string{
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
	assert.Equal(suite.T(), flowcm.ExecFailure, resp.Status)
	assert.Contains(suite.T(), resp.FailureReason, "Failed to authenticate user")
	suite.mockUserService.AssertExpectations(suite.T())
	suite.mockCredsService.AssertExpectations(suite.T())
}

func (suite *BasicAuthExecutorTestSuite) TestGetAuthenticatedUser_SuccessfulAuthentication() {
	attrs := map[string]interface{}{"email": "test@example.com", "phone": "1234567890"}
	attrsJSON, _ := json.Marshal(attrs)

	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		UserInputData: map[string]string{
			userAttributeUsername: "testuser",
			userAttributePassword: "password123",
		},
	}

	execResp := &flowcm.ExecutorResponse{
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
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		UserInputData: map[string]string{
			userAttributeUsername: "testuser",
			userAttributePassword: "password123",
		},
	}

	execResp := &flowcm.ExecutorResponse{
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

func (suite *BasicAuthExecutorTestSuite) TestExecute_Observability_Success() {
	// Enable observability for this test
	suite.mockObservability.ExpectedCalls = nil
	suite.mockObservability.On("IsEnabled").Return(true)

	userID := testUserID
	attrs := map[string]interface{}{"email": "test@example.com"}
	attrsJSON, _ := json.Marshal(attrs)

	ctx := &flowcore.NodeContext{
		FlowID:        "flow-123",
		AppID:         "app-1",
		CurrentNodeID: "node-1",
		FlowType:      flowcm.FlowTypeAuthentication,
		UserInputData: map[string]string{
			userAttributeUsername: "testuser",
			userAttributePassword: "password123",
		},
	}

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

	// Expect Started event
	suite.mockObservability.On("PublishEvent", mock.MatchedBy(func(evt *event.Event) bool {
		return evt.Type == string(event.EventTypeFlowNodeExecutionStarted) &&
			evt.Status == event.StatusInProgress &&
			evt.Data[event.DataKey.FlowID] == "flow-123" &&
			evt.Data[event.DataKey.AppID] == "app-1" &&
			evt.Data[event.DataKey.NodeID] == "node-1"
	})).Return()

	// Expect Completed event
	suite.mockObservability.On("PublishEvent", mock.MatchedBy(func(evt *event.Event) bool {
		return evt.Type == string(event.EventTypeFlowNodeExecutionCompleted) &&
			evt.Status == event.StatusSuccess &&
			evt.Data[event.DataKey.FlowID] == "flow-123" &&
			evt.Data[event.DataKey.AppID] == "app-1" &&
			evt.Data[event.DataKey.NodeID] == "node-1" &&
			evt.Data[event.DataKey.UserID] == testUserID
	})).Return()

	resp, err := suite.executor.Execute(ctx)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), flowcm.ExecComplete, resp.Status)

	suite.mockUserService.AssertExpectations(suite.T())
	suite.mockCredsService.AssertExpectations(suite.T())
	suite.mockObservability.AssertExpectations(suite.T())
}

func (suite *BasicAuthExecutorTestSuite) TestExecute_Observability_Failure() {
	// Enable observability for this test
	suite.mockObservability.ExpectedCalls = nil
	suite.mockObservability.On("IsEnabled").Return(true)

	userID := testUserID
	ctx := &flowcore.NodeContext{
		FlowID:        "flow-123",
		AppID:         "app-1",
		CurrentNodeID: "node-1",
		FlowType:      flowcm.FlowTypeAuthentication,
		UserInputData: map[string]string{
			userAttributeUsername: "testuser",
			userAttributePassword: "wrongpassword",
		},
	}

	suite.mockUserService.On("IdentifyUser", map[string]interface{}{
		userAttributeUsername: "testuser",
	}).Return(&userID, nil)

	suite.mockCredsService.On("Authenticate", map[string]interface{}{
		userAttributeUsername: "testuser",
		userAttributePassword: "wrongpassword",
	}).Return(nil, &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		ErrorDescription: "Invalid credentials",
		Error:            "invalid_credentials",
		Code:             "1001",
	})

	// Expect Started event
	suite.mockObservability.On("PublishEvent", mock.MatchedBy(func(evt *event.Event) bool {
		return evt.Type == string(event.EventTypeFlowNodeExecutionStarted) &&
			evt.Status == event.StatusInProgress
	})).Return()

	// Expect Failed event
	suite.mockObservability.On("PublishEvent", mock.MatchedBy(func(evt *event.Event) bool {
		return evt.Type == string(event.EventTypeFlowNodeExecutionFailed) &&
			evt.Status == event.StatusFailure &&
			evt.Data[event.DataKey.Error] == "invalid_credentials" &&
			evt.Data[event.DataKey.ErrorCode] == "1001" &&
			evt.Data[event.DataKey.ErrorType] == string(serviceerror.ClientErrorType) &&
			evt.Data[event.DataKey.Message] == "Invalid credentials"
	})).Return()

	resp, err := suite.executor.Execute(ctx)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), flowcm.ExecFailure, resp.Status)

	suite.mockUserService.AssertExpectations(suite.T())
	suite.mockCredsService.AssertExpectations(suite.T())
	suite.mockObservability.AssertExpectations(suite.T())
}
