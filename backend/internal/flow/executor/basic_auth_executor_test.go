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

	"github.com/asgardeo/thunder/internal/authnprovider"
	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/userprovider"
	"github.com/asgardeo/thunder/tests/mocks/authn/credentialsmock"
	"github.com/asgardeo/thunder/tests/mocks/flow/coremock"
	"github.com/asgardeo/thunder/tests/mocks/observabilitymock"
	"github.com/asgardeo/thunder/tests/mocks/userprovidermock"
)

type BasicAuthExecutorTestSuite struct {
	suite.Suite
	mockUserProvider  *userprovidermock.UserProviderInterfaceMock
	mockCredsService  *credentialsmock.CredentialsAuthnServiceInterfaceMock
	mockFlowFactory   *coremock.FlowFactoryInterfaceMock
	mockObservability *observabilitymock.ObservabilityServiceInterfaceMock
	executor          *basicAuthExecutor
}

func TestBasicAuthExecutorSuite(t *testing.T) {
	suite.Run(t, new(BasicAuthExecutorTestSuite))
}

func (suite *BasicAuthExecutorTestSuite) SetupTest() {
	suite.mockUserProvider = userprovidermock.NewUserProviderInterfaceMock(suite.T())
	suite.mockCredsService = credentialsmock.NewCredentialsAuthnServiceInterfaceMock(suite.T())
	suite.mockFlowFactory = coremock.NewFlowFactoryInterfaceMock(suite.T())
	suite.mockObservability = observabilitymock.NewObservabilityServiceInterfaceMock(suite.T())

	defaultInputs := []common.Input{
		{Identifier: userAttributeUsername, Type: common.InputTypeText, Required: true},
		{Identifier: userAttributePassword, Type: common.InputTypePassword, Required: true},
	}

	// Mock the embedded identifying executor first
	identifyingMock := createMockIdentifyingExecutor(suite.T())
	suite.mockFlowFactory.On("CreateExecutor", ExecutorNameIdentifying, common.ExecutorTypeUtility,
		mock.Anything, mock.Anything).Return(identifyingMock).Maybe()

	mockExec := createMockBasicAuthExecutor(suite.T())
	suite.mockFlowFactory.On("CreateExecutor", ExecutorNameBasicAuth, common.ExecutorTypeAuthentication,
		defaultInputs, []common.Input{}).Return(mockExec)

	suite.executor = newBasicAuthExecutor(suite.mockFlowFactory, suite.mockUserProvider, suite.mockCredsService,
		suite.mockObservability)
}

func (suite *BasicAuthExecutorTestSuite) BeforeTest(suiteName, testName string) {
	suite.mockObservability.ExpectedCalls = nil
	suite.mockObservability.On("IsEnabled").Return(false).Maybe()
}

func createMockIdentifyingExecutor(t *testing.T) core.ExecutorInterface {
	mockExec := coremock.NewExecutorInterfaceMock(t)
	mockExec.On("GetName").Return(ExecutorNameIdentifying).Maybe()
	mockExec.On("GetType").Return(common.ExecutorTypeUtility).Maybe()
	mockExec.On("GetDefaultInputs").Return([]common.Input{}).Maybe()
	mockExec.On("GetPrerequisites").Return([]common.Input{}).Maybe()
	return mockExec
}

func createMockExecutorWithCustomInputs(t *testing.T, name string,
	inputs []common.Input) core.ExecutorInterface {
	mockExec := coremock.NewExecutorInterfaceMock(t)
	mockExec.On("GetName").Return(name).Maybe()
	mockExec.On("GetType").Return(common.ExecutorTypeAuthentication).Maybe()
	mockExec.On("GetDefaultInputs").Return(inputs).Maybe()
	mockExec.On("GetRequiredInputs", mock.Anything).Return(inputs).Maybe()
	mockExec.On("GetPrerequisites").Return([]common.Input{}).Maybe()
	mockExec.On("HasRequiredInputs", mock.Anything, mock.Anything).Return(
		func(ctx *core.NodeContext, execResp *common.ExecutorResponse) bool {
			for _, input := range inputs {
				if input.Required {
					value, exists := ctx.UserInputs[input.Identifier]
					if !exists || value == "" {
						execResp.Inputs = inputs
						execResp.Status = common.ExecUserInputRequired
						return false
					}
				}
			}
			return true
		}).Maybe()
	return mockExec
}

func createMockBasicAuthExecutor(t *testing.T) core.ExecutorInterface {
	mockExec := coremock.NewExecutorInterfaceMock(t)
	mockExec.On("GetName").Return(ExecutorNameBasicAuth).Maybe()
	mockExec.On("GetType").Return(common.ExecutorTypeAuthentication).Maybe()
	mockExec.On("GetDefaultInputs").Return([]common.Input{
		{Identifier: userAttributeUsername, Type: common.InputTypeText, Required: true},
		{Identifier: userAttributePassword, Type: common.InputTypePassword, Required: true},
	}).Maybe()
	mockExec.On("GetRequiredInputs", mock.Anything).Return([]common.Input{
		{Identifier: userAttributeUsername, Type: common.InputTypeText, Required: true},
		{Identifier: userAttributePassword, Type: common.InputTypePassword, Required: true},
	}).Maybe()
	mockExec.On("GetPrerequisites").Return([]common.Input{}).Maybe()
	mockExec.On("HasRequiredInputs", mock.Anything, mock.Anything).Return(
		func(ctx *core.NodeContext, execResp *common.ExecutorResponse) bool {
			username, hasUsername := ctx.UserInputs[userAttributeUsername]
			password, hasPassword := ctx.UserInputs[userAttributePassword]
			if !hasUsername || username == "" || !hasPassword || password == "" {
				execResp.Inputs = []common.Input{
					{Identifier: userAttributeUsername, Type: common.InputTypeText, Required: true},
					{Identifier: userAttributePassword, Type: common.InputTypePassword, Required: true},
				}
				execResp.Status = common.ExecUserInputRequired
				return false
			}
			return true
		}).Maybe()
	return mockExec
}

func (suite *BasicAuthExecutorTestSuite) TestNewBasicAuthExecutor() {
	assert.NotNil(suite.T(), suite.executor)
	assert.NotNil(suite.T(), suite.executor.credsAuthSvc)
	assert.NotNil(suite.T(), suite.executor.userProvider)
}

func (suite *BasicAuthExecutorTestSuite) TestExecute_Success_AuthenticationFlow() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			userAttributeUsername: "testuser",
			userAttributePassword: "password123",
		},
		RuntimeData: make(map[string]string),
	}

	authenticateResult := &authnprovider.AuthnResult{
		UserID:             testUserID,
		UserType:           "person",
		OrganizationUnitID: "ou-123",
		Token:              "test-token",
		AvailableAttributes: []authnprovider.AvailableAttribute{
			{Name: "username", DisplayName: "username", Verified: true},
		},
	}

	suite.mockCredsService.On("Authenticate", map[string]interface{}{
		userAttributeUsername: "testuser",
	}, map[string]interface{}{
		userAttributePassword: "password123",
	}, mock.Anything).Return(authenticateResult, nil)

	suite.mockUserProvider.On("GetUser", testUserID).Return(nil,
		userprovider.NewUserProviderError(userprovider.ErrorCodeNotImplemented, "", ""))

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	assert.True(suite.T(), resp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), testUserID, resp.AuthenticatedUser.UserID)
	suite.mockCredsService.AssertExpectations(suite.T())
}

func (suite *BasicAuthExecutorTestSuite) TestExecute_Success_WithEmailAttribute() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		},
		RuntimeData: make(map[string]string),
	}

	// Override GetRequiredInputs to return email and password as required fields
	originalInputs := []common.Input{
		{Identifier: "email", Type: common.InputTypeText, Required: true},
		{Identifier: "password", Type: common.InputTypePassword, Required: true},
	}
	suite.executor.ExecutorInterface = createMockExecutorWithCustomInputs(
		suite.T(), ExecutorNameBasicAuth, originalInputs)

	authenticatedUser := &authnprovider.AuthnResult{
		UserID:             testUserID,
		UserType:           "person",
		OrganizationUnitID: "ou-123",
		Token:              "test-token",
		AvailableAttributes: []authnprovider.AvailableAttribute{
			{Name: "email", DisplayName: "email", Verified: true},
		},
	}

	suite.mockCredsService.On("Authenticate", map[string]interface{}{
		"email": "test@example.com",
	}, map[string]interface{}{
		"password": "password123",
	}, mock.Anything).Return(authenticatedUser, nil)

	suite.mockUserProvider.On("GetUser", testUserID).Return(nil,
		userprovider.NewUserProviderError(userprovider.ErrorCodeNotImplemented, "", ""))

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	assert.True(suite.T(), resp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), testUserID, resp.AuthenticatedUser.UserID)
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

	suite.mockUserProvider.On("IdentifyUser", map[string]interface{}{
		userAttributeUsername: "newuser",
	}).Return(nil, userprovider.NewUserProviderError(userprovider.ErrorCodeUserNotFound, "", ""))

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	assert.False(suite.T(), resp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), "newuser", resp.AuthenticatedUser.Attributes[userAttributeUsername])
	suite.mockUserProvider.AssertExpectations(suite.T())
}

func (suite *BasicAuthExecutorTestSuite) TestExecute_Success_WithMultipleAttributes() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			"email":    "test@example.com",
			"phone":    "+1234567890",
			"password": "password123",
		},
		RuntimeData: make(map[string]string),
	}

	// Override GetRequiredInputs to return email, phone, and password as required fields
	customInputs := []common.Input{
		{Identifier: "email", Type: common.InputTypeText, Required: true},
		{Identifier: "phone", Type: common.InputTypeText, Required: true},
		{Identifier: "password", Type: common.InputTypePassword, Required: true},
	}
	suite.executor.ExecutorInterface = createMockExecutorWithCustomInputs(
		suite.T(), ExecutorNameBasicAuth, customInputs)

	authenticatedUser := &authnprovider.AuthnResult{
		UserID:             testUserID,
		UserType:           "person",
		OrganizationUnitID: "ou-123",
		Token:              "test-token",
		AvailableAttributes: []authnprovider.AvailableAttribute{
			{Name: "email", DisplayName: "email", Verified: true},
			{Name: "phone", DisplayName: "phone", Verified: true},
		},
	}

	suite.mockCredsService.On("Authenticate", map[string]interface{}{
		"email": "test@example.com",
		"phone": "+1234567890",
	}, map[string]interface{}{
		"password": "password123",
	}, mock.Anything).Return(authenticatedUser, nil)

	suite.mockUserProvider.On("GetUser", testUserID).Return(nil,
		userprovider.NewUserProviderError(userprovider.ErrorCodeNotImplemented, "", ""))

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	assert.True(suite.T(), resp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), testUserID, resp.AuthenticatedUser.UserID)
	suite.mockCredsService.AssertExpectations(suite.T())
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

	suite.mockCredsService.On("Authenticate", map[string]interface{}{
		userAttributeUsername: "testuser",
	}, map[string]interface{}{
		userAttributePassword: "wrongpassword",
	}, mock.Anything).Return(nil, &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		ErrorDescription: "Invalid credentials",
	})

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Contains(suite.T(), resp.FailureReason, "Failed to authenticate user")
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

	// Authenticate internally calls IdentifyUser and returns user not found error
	suite.mockCredsService.On("Authenticate", map[string]interface{}{
		userAttributeUsername: "nonexistent",
	}, map[string]interface{}{
		userAttributePassword: "password123",
	}, mock.Anything).Return(nil, &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		ErrorDescription: "User not found",
	})

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	suite.mockCredsService.AssertExpectations(suite.T())
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
	suite.mockUserProvider.On("IdentifyUser", map[string]interface{}{
		userAttributeUsername: "existinguser",
	}).Return(&userID, nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Contains(suite.T(), resp.FailureReason, "User already exists")
	suite.mockUserProvider.AssertExpectations(suite.T())
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

	// Authenticate returns a server error (e.g., database error)
	suite.mockCredsService.On("Authenticate", map[string]interface{}{
		userAttributeUsername: "testuser",
	}, map[string]interface{}{
		userAttributePassword: "password123",
	}, mock.Anything).Return(nil, &serviceerror.ServiceError{
		Type:  serviceerror.ServerErrorType,
		Error: "database error",
	})

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	suite.mockCredsService.AssertExpectations(suite.T())
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

	suite.mockCredsService.On("Authenticate", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, &serviceerror.ServiceError{
			Type:  serviceerror.ServerErrorType,
			Error: "internal server error",
		})

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Contains(suite.T(), resp.FailureReason, "Failed to authenticate user")
	suite.mockCredsService.AssertExpectations(suite.T())
}

func (suite *BasicAuthExecutorTestSuite) TestGetAuthenticatedUser_SuccessfulAuthentication() {
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

	authenticatedUser := &authnprovider.AuthnResult{
		UserID:             testUserID,
		UserType:           "person",
		OrganizationUnitID: "ou-123",
		Token:              "test-token",
		AvailableAttributes: []authnprovider.AvailableAttribute{
			{Name: "email", DisplayName: "email", Verified: true},
			{Name: "phone", DisplayName: "phone", Verified: true},
		},
	}

	suite.mockCredsService.On("Authenticate", map[string]interface{}{
		userAttributeUsername: "testuser",
	}, map[string]interface{}{
		userAttributePassword: "password123",
	}, mock.Anything).Return(authenticatedUser, nil)

	suite.mockUserProvider.On("GetUser", testUserID).Return(nil,
		userprovider.NewUserProviderError(userprovider.ErrorCodeNotImplemented, "", ""))

	result, err := suite.executor.getAuthenticatedUser(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.True(suite.T(), result.IsAuthenticated)
	assert.Equal(suite.T(), testUserID, result.UserID)
	assert.Equal(suite.T(), "ou-123", result.OrganizationUnitID)
	assert.Equal(suite.T(), "person", result.UserType)
	assert.Equal(suite.T(), "email", result.AvailableAttributes[0].Name)
	assert.Equal(suite.T(), "phone", result.AvailableAttributes[1].Name)
	suite.mockCredsService.AssertExpectations(suite.T())
}

func (suite *BasicAuthExecutorTestSuite) TestGetAuthenticatedUser_Success_WithFetchedAttributes() {
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

	authenticateResult := &authnprovider.AuthnResult{
		UserID:             testUserID,
		UserType:           "person",
		OrganizationUnitID: "ou-123",
		Token:              "test-token",
		AvailableAttributes: []authnprovider.AvailableAttribute{
			{Name: "username", DisplayName: "username", Verified: true},
			{Name: "email", DisplayName: "email", Verified: true},
			{Name: "role", DisplayName: "role", Verified: true},
		},
	}

	suite.mockCredsService.On("Authenticate", map[string]interface{}{
		userAttributeUsername: "testuser",
	}, map[string]interface{}{
		userAttributePassword: "password123",
	}, mock.Anything).Return(authenticateResult, nil)

	// Mock UserProvider response
	attrs := map[string]interface{}{"username": "testuser", "email": "fetched@example.com", "role": "admin"}
	attrsJSON, _ := json.Marshal(attrs)
	user := &userprovider.User{
		UserID:     testUserID,
		Attributes: attrsJSON,
	}
	suite.mockUserProvider.On("GetUser", testUserID).Return(user, nil)

	result, err := suite.executor.getAuthenticatedUser(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.True(suite.T(), result.IsAuthenticated)
	assert.Equal(suite.T(), testUserID, result.UserID)
	assert.Equal(suite.T(), "testuser", result.Attributes["username"])
	assert.Equal(suite.T(), "fetched@example.com", result.Attributes["email"])
	assert.Equal(suite.T(), "admin", result.Attributes["role"])
	suite.mockCredsService.AssertExpectations(suite.T())
	suite.mockUserProvider.AssertExpectations(suite.T())
}

func (suite *BasicAuthExecutorTestSuite) TestGetAuthenticatedUser_AuthenticationFlow_NoRedundantIdentifyUser() {
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

	authenticatedUser := &authnprovider.AuthnResult{
		UserID:             testUserID,
		UserType:           "person",
		OrganizationUnitID: "ou-123",
		Token:              "test-token",
		AvailableAttributes: []authnprovider.AvailableAttribute{
			{Name: "email", DisplayName: "email", Verified: true},
			{Name: "phone", DisplayName: "phone", Verified: true},
		},
	}

	suite.mockCredsService.On("Authenticate", map[string]interface{}{
		userAttributeUsername: "testuser",
	}, map[string]interface{}{
		userAttributePassword: "password123",
	}, mock.Anything).Return(authenticatedUser, nil)

	suite.mockUserProvider.On("GetUser", testUserID).Return(nil,
		userprovider.NewUserProviderError(userprovider.ErrorCodeNotImplemented, "", ""))

	result, err := suite.executor.getAuthenticatedUser(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.True(suite.T(), result.IsAuthenticated)
	assert.Equal(suite.T(), testUserID, result.UserID)
	// Verify Authenticate was called (which handles IdentifyUser + VerifyUser internally)
	// This test verifies the optimization: no explicit IdentifyUser call for auth flows
	suite.mockCredsService.AssertExpectations(suite.T())
}

func (suite *BasicAuthExecutorTestSuite) TestGetAuthenticatedUser_RegistrationFlow_CallsIdentifyUser() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeRegistration,
		UserInputs: map[string]string{
			userAttributeUsername: "newuser",
			userAttributePassword: "password123",
		},
	}

	execResp := &common.ExecutorResponse{
		RuntimeData: make(map[string]string),
	}

	// For registration flows, IdentifyUser should be called to check if user exists
	suite.mockUserProvider.On("IdentifyUser", map[string]interface{}{
		userAttributeUsername: "newuser",
	}).Return(nil, userprovider.NewUserProviderError(userprovider.ErrorCodeUserNotFound, "", ""))

	result, err := suite.executor.getAuthenticatedUser(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.False(suite.T(), result.IsAuthenticated)
	assert.Equal(suite.T(), "newuser", result.Attributes[userAttributeUsername])
	// Verify IdentifyUser was called for registration flow
	suite.mockUserProvider.AssertExpectations(suite.T())
	// Verify Authenticate was NOT called for registration flow
	suite.mockCredsService.AssertNotCalled(suite.T(), "Authenticate")
}
