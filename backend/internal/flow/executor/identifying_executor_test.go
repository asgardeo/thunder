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
	"github.com/asgardeo/thunder/internal/userprovider"
	"github.com/asgardeo/thunder/tests/mocks/flow/coremock"
	"github.com/asgardeo/thunder/tests/mocks/userprovidermock"
)

type IdentifyingExecutorTestSuite struct {
	suite.Suite
	mockUserProvider *userprovidermock.UserProviderInterfaceMock
	mockFlowFactory  *coremock.FlowFactoryInterfaceMock
	executor         *identifyingExecutor
}

func TestIdentifyingExecutorSuite(t *testing.T) {
	suite.Run(t, new(IdentifyingExecutorTestSuite))
}

func (suite *IdentifyingExecutorTestSuite) SetupTest() {
	suite.mockUserProvider = userprovidermock.NewUserProviderInterfaceMock(suite.T())
	suite.mockFlowFactory = coremock.NewFlowFactoryInterfaceMock(suite.T())

	mockExec := createMockExecutor(suite.T(), ExecutorNameIdentifying, common.ExecutorTypeUtility)
	suite.mockFlowFactory.On("CreateExecutor", ExecutorNameIdentifying, common.ExecutorTypeUtility,
		[]common.Input{}, []common.Input{}).Return(mockExec)

	suite.executor = newIdentifyingExecutor(ExecutorNameIdentifying, []common.Input{},
		[]common.Input{}, suite.mockFlowFactory, suite.mockUserProvider)
}

func (suite *IdentifyingExecutorTestSuite) TestNewIdentifyingExecutor() {
	assert.NotNil(suite.T(), suite.executor)
	assert.NotNil(suite.T(), suite.executor.userProvider)

	// Test default name
	exec := newIdentifyingExecutor(
		"",
		[]common.Input{},
		[]common.Input{},
		suite.mockFlowFactory,
		suite.mockUserProvider,
	)
	assert.NotNil(suite.T(), exec)
}

func (suite *IdentifyingExecutorTestSuite) TestIdentifyUser_Success() {
	filters := map[string]interface{}{"username": "testuser"}
	execResp := &common.ExecutorResponse{
		RuntimeData: make(map[string]string),
	}
	// Use package-level testUserID constant
	userID := testUserID
	suite.mockUserProvider.On("IdentifyUser", filters).Return(&userID, nil)

	result, err := suite.executor.IdentifyUser(filters, execResp)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), testUserID, *result)
	suite.mockUserProvider.AssertExpectations(suite.T())
}

func (suite *IdentifyingExecutorTestSuite) TestIdentifyUser_UserNotFound() {
	filters := map[string]interface{}{"username": "nonexistent"}
	execResp := &common.ExecutorResponse{
		RuntimeData: make(map[string]string),
	}

	suite.mockUserProvider.On("IdentifyUser", filters).Return(nil,
		userprovider.NewUserProviderError(userprovider.ErrorCodeUserNotFound, "", ""))

	result, err := suite.executor.IdentifyUser(filters, execResp)

	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), common.ExecFailure, execResp.Status)
	assert.Equal(suite.T(), failureReasonUserNotFound, execResp.FailureReason)
	suite.mockUserProvider.AssertExpectations(suite.T())
}

func (suite *IdentifyingExecutorTestSuite) TestIdentifyUser_ServiceError() {
	filters := map[string]interface{}{"username": "testuser"}
	execResp := &common.ExecutorResponse{
		RuntimeData: make(map[string]string),
	}

	suite.mockUserProvider.On("IdentifyUser", filters).Return(nil,
		userprovider.NewUserProviderError(userprovider.ErrorCodeSystemError, "", ""))

	result, err := suite.executor.IdentifyUser(filters, execResp)

	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), common.ExecFailure, execResp.Status)
	assert.Contains(suite.T(), execResp.FailureReason, "Failed to identify user")
	suite.mockUserProvider.AssertExpectations(suite.T())
}

func (suite *IdentifyingExecutorTestSuite) TestIdentifyUser_EmptyUserID() {
	filters := map[string]interface{}{"username": "testuser"}
	execResp := &common.ExecutorResponse{
		RuntimeData: make(map[string]string),
	}
	emptyID := ""

	suite.mockUserProvider.On("IdentifyUser", filters).Return(&emptyID, nil)

	result, err := suite.executor.IdentifyUser(filters, execResp)

	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), common.ExecFailure, execResp.Status)
	assert.Equal(suite.T(), failureReasonUserNotFound, execResp.FailureReason)
	suite.mockUserProvider.AssertExpectations(suite.T())
}

func (suite *IdentifyingExecutorTestSuite) TestIdentifyUser_FilterNonSearchableAttributes() {
	filters := map[string]interface{}{
		"username": "testuser",
		"password": "secret123",
		"code":     "auth-code",
		"nonce":    "nonce-value",
		"otp":      "123456",
	}
	execResp := &common.ExecutorResponse{
		RuntimeData: make(map[string]string),
	}
	// Use package-level testUserID constant
	suite.mockUserProvider.On("IdentifyUser", map[string]interface{}{
		"username": "testuser",
	}).Return(func() *string {
		userID := testUserID
		return &userID
	}(), nil)

	result, err := suite.executor.IdentifyUser(filters, execResp)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), testUserID, *result)
	suite.mockUserProvider.AssertExpectations(suite.T())
}

func (suite *IdentifyingExecutorTestSuite) TestIdentifyUser_WithEmail() {
	filters := map[string]interface{}{"email": "test@example.com"}
	execResp := &common.ExecutorResponse{
		RuntimeData: make(map[string]string),
	}
	emailUserID := "user-456"

	suite.mockUserProvider.On("IdentifyUser", filters).Return(&emailUserID, nil)

	result, err := suite.executor.IdentifyUser(filters, execResp)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "user-456", *result)
	suite.mockUserProvider.AssertExpectations(suite.T())
}

func (suite *IdentifyingExecutorTestSuite) TestIdentifyUser_WithMobileNumber() {
	filters := map[string]interface{}{"mobileNumber": "+1234567890"}
	execResp := &common.ExecutorResponse{
		RuntimeData: make(map[string]string),
	}
	mobileUserID := "user-789"

	suite.mockUserProvider.On("IdentifyUser", filters).Return(&mobileUserID, nil)

	result, err := suite.executor.IdentifyUser(filters, execResp)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "user-789", *result)
	suite.mockUserProvider.AssertExpectations(suite.T())
}

func (suite *IdentifyingExecutorTestSuite) TestExecute_Success_UserInputs() {
	ctx := &core.NodeContext{
		FlowID:     "flow-123",
		UserInputs: map[string]string{"username": "testuser"},
	}
	// Use package-level testUserID constant
	// Configure mock base executor
	mockBase := suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock)
	mockBase.On("HasRequiredInputs", mock.Anything, mock.Anything).Return(true)
	mockBase.On("GetRequiredInputs", mock.Anything).Return([]common.Input{
		{Identifier: "username", Type: "string", Required: true},
	})

	suite.mockUserProvider.On("IdentifyUser", map[string]interface{}{
		"username": "testuser",
	}).Return(func() *string {
		userID := testUserID
		return &userID
	}(), nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	assert.Equal(suite.T(), testUserID, resp.RuntimeData[userAttributeUserID])
}

func (suite *IdentifyingExecutorTestSuite) TestExecute_Success_RuntimeData() {
	ctx := &core.NodeContext{
		FlowID:      "flow-123",
		UserInputs:  make(map[string]string),
		RuntimeData: map[string]string{"username": "testuser"},
	}
	// Use package-level testUserID constant
	mockBase := suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock)
	mockBase.On("HasRequiredInputs", mock.Anything, mock.Anything).Return(true)
	mockBase.On("GetRequiredInputs", mock.Anything).Return([]common.Input{
		{Identifier: "username", Type: "string", Required: true},
	})

	suite.mockUserProvider.On("IdentifyUser", map[string]interface{}{
		"username": "testuser",
	}).Return(func() *string {
		userID := testUserID
		return &userID
	}(), nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	assert.Equal(suite.T(), testUserID, resp.RuntimeData[userAttributeUserID])
}

func (suite *IdentifyingExecutorTestSuite) TestExecute_UserInputRequired() {
	ctx := &core.NodeContext{
		FlowID:     "flow-123",
		UserInputs: map[string]string{},
	}

	mockBase := suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock)
	mockBase.On("HasRequiredInputs", mock.Anything, mock.Anything).Return(false)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecUserInputRequired, resp.Status)
}

func (suite *IdentifyingExecutorTestSuite) TestExecute_Failure_IdentifyUserError() {
	ctx := &core.NodeContext{
		FlowID:     "flow-123",
		UserInputs: map[string]string{"username": "testuser"},
	}

	mockBase := suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock)
	mockBase.On("HasRequiredInputs", mock.Anything, mock.Anything).Return(true)
	mockBase.On("GetRequiredInputs", mock.Anything).Return([]common.Input{
		{Identifier: "username", Type: "string", Required: true},
	})

	suite.mockUserProvider.On("IdentifyUser", map[string]interface{}{
		"username": "testuser",
	}).Return(nil, userprovider.NewUserProviderError(userprovider.ErrorCodeUserNotFound, "", ""))

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	// IdentifyUser method in implementation swallows the error and returns nil, nil.
	// Then Execute checks for nil userID and returns UserNotFound.
	// So we should expect failureReasonUserNotFound
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Equal(suite.T(), failureReasonUserNotFound, resp.FailureReason)
}

func (suite *IdentifyingExecutorTestSuite) TestExecute_Failure_UserNotFound() {
	ctx := &core.NodeContext{
		FlowID:     "flow-123",
		UserInputs: map[string]string{"username": "nonexistent"},
	}

	mockBase := suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock)
	mockBase.On("HasRequiredInputs", mock.Anything, mock.Anything).Return(true)
	mockBase.On("GetRequiredInputs", mock.Anything).Return([]common.Input{
		{Identifier: "username", Type: "string", Required: true},
	})

	emptyID := ""
	suite.mockUserProvider.On("IdentifyUser", map[string]interface{}{
		"username": "nonexistent",
	}).Return(&emptyID, nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Equal(suite.T(), failureReasonUserNotFound, resp.FailureReason)
}

// TestExecute_Success_WithVariousAttributes tests successful user identification with different attributes.
func (suite *IdentifyingExecutorTestSuite) TestExecute_Success_WithVariousAttributes() {
	testCases := []struct {
		name       string
		attribute  string
		value      string
		expectedID string
	}{
		{"email", "email", "test@example.com", "user-email-456"},
		{"mobileNumber", "mobileNumber", "+1234567890", "user-mobile-789"},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			ctx := &core.NodeContext{
				FlowID:     "flow-123",
				UserInputs: map[string]string{tc.attribute: tc.value},
			}

			mockBase := suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock)
			mockBase.On("HasRequiredInputs", mock.Anything, mock.Anything).Return(true)
			mockBase.On("GetRequiredInputs", mock.Anything).Return([]common.Input{
				{Identifier: tc.attribute, Type: "string", Required: true},
			})

			suite.mockUserProvider.On("IdentifyUser", map[string]interface{}{
				tc.attribute: tc.value,
			}).Return(&tc.expectedID, nil)

			resp, err := suite.executor.Execute(ctx)

			assert.NoError(suite.T(), err)
			assert.NotNil(suite.T(), resp)
			assert.Equal(suite.T(), common.ExecComplete, resp.Status)
			assert.Equal(suite.T(), tc.expectedID, resp.RuntimeData[userAttributeUserID])
			suite.mockUserProvider.AssertExpectations(suite.T())
		})
	}
}

func (suite *IdentifyingExecutorTestSuite) TestExecute_Success_WithMultipleAttributes() {
	ctx := &core.NodeContext{
		FlowID: "flow-123",
		UserInputs: map[string]string{
			"username": "testuser",
			"email":    "test@example.com",
		},
	}
	multiAttrUserID := "user-multi-123"

	mockBase := suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock)
	mockBase.On("HasRequiredInputs", mock.Anything, mock.Anything).Return(true)
	mockBase.On("GetRequiredInputs", mock.Anything).Return([]common.Input{
		{Identifier: "username", Type: "string", Required: true},
		{Identifier: "email", Type: "string", Required: true},
	})

	suite.mockUserProvider.On("IdentifyUser", map[string]interface{}{
		"username": "testuser",
		"email":    "test@example.com",
	}).Return(&multiAttrUserID, nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	assert.Equal(suite.T(), multiAttrUserID, resp.RuntimeData[userAttributeUserID])
	suite.mockUserProvider.AssertExpectations(suite.T())
}

// TestExecute_Failure_UserNotFoundByAttribute tests failure handling when user is not found by different attributes.
func (suite *IdentifyingExecutorTestSuite) TestExecute_Failure_UserNotFoundByAttribute() {
	testCases := []struct {
		name      string
		attribute string
		value     string
	}{
		{"email", "email", "nonexistent@example.com"},
		{"mobileNumber", "mobileNumber", "+0000000000"},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			ctx := &core.NodeContext{
				FlowID:     "flow-123",
				UserInputs: map[string]string{tc.attribute: tc.value},
			}

			mockBase := suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock)
			mockBase.On("HasRequiredInputs", mock.Anything, mock.Anything).Return(true)
			mockBase.On("GetRequiredInputs", mock.Anything).Return([]common.Input{
				{Identifier: tc.attribute, Type: "string", Required: true},
			})

			suite.mockUserProvider.On("IdentifyUser", map[string]interface{}{
				tc.attribute: tc.value,
			}).Return(nil, userprovider.NewUserProviderError(userprovider.ErrorCodeUserNotFound, "", ""))

			resp, err := suite.executor.Execute(ctx)

			assert.NoError(suite.T(), err)
			assert.NotNil(suite.T(), resp)
			assert.Equal(suite.T(), common.ExecFailure, resp.Status)
			assert.Equal(suite.T(), failureReasonUserNotFound, resp.FailureReason)
			suite.mockUserProvider.AssertExpectations(suite.T())
		})
	}
}

// TestExecute_Success_FromRuntimeData tests successful identification when attributes come from RuntimeData.
func (suite *IdentifyingExecutorTestSuite) TestExecute_Success_FromRuntimeData() {
	testCases := []struct {
		name       string
		attribute  string
		value      string
		expectedID string
	}{
		{"email", "email", "runtime@example.com", "user-runtime-email-456"},
		{"mobileNumber", "mobileNumber", "+9876543210", "user-runtime-mobile-789"},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			ctx := &core.NodeContext{
				FlowID:      "flow-123",
				UserInputs:  make(map[string]string),
				RuntimeData: map[string]string{tc.attribute: tc.value},
			}

			mockBase := suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock)
			mockBase.On("HasRequiredInputs", mock.Anything, mock.Anything).Return(true)
			mockBase.On("GetRequiredInputs", mock.Anything).Return([]common.Input{
				{Identifier: tc.attribute, Type: "string", Required: true},
			})

			suite.mockUserProvider.On("IdentifyUser", map[string]interface{}{
				tc.attribute: tc.value,
			}).Return(&tc.expectedID, nil)

			resp, err := suite.executor.Execute(ctx)

			assert.NoError(suite.T(), err)
			assert.NotNil(suite.T(), resp)
			assert.Equal(suite.T(), common.ExecComplete, resp.Status)
			assert.Equal(suite.T(), tc.expectedID, resp.RuntimeData[userAttributeUserID])
			suite.mockUserProvider.AssertExpectations(suite.T())
		})
	}
}

// TestExecute_Failure_EmptyInput tests failure handling when input value is an empty string.
func (suite *IdentifyingExecutorTestSuite) TestExecute_Failure_EmptyInput() {
	testCases := []struct {
		name      string
		attribute string
	}{
		{"username", "username"},
		{"email", "email"},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			ctx := &core.NodeContext{
				FlowID:     "flow-123",
				UserInputs: map[string]string{tc.attribute: ""},
			}

			mockBase := suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock)
			mockBase.On("HasRequiredInputs", mock.Anything, mock.Anything).Return(true)
			mockBase.On("GetRequiredInputs", mock.Anything).Return([]common.Input{
				{Identifier: tc.attribute, Type: "string", Required: true},
			})

			emptyID := ""
			suite.mockUserProvider.On("IdentifyUser", map[string]interface{}{
				tc.attribute: "",
			}).Return(&emptyID, nil)

			resp, err := suite.executor.Execute(ctx)

			assert.NoError(suite.T(), err)
			assert.NotNil(suite.T(), resp)
			assert.Equal(suite.T(), common.ExecFailure, resp.Status)
			assert.Equal(suite.T(), failureReasonUserNotFound, resp.FailureReason)
			suite.mockUserProvider.AssertExpectations(suite.T())
		})
	}
}

// TestExecute_UserInputsPriorityOverRuntimeData tests that UserInputs takes priority over RuntimeData.
func (suite *IdentifyingExecutorTestSuite) TestExecute_UserInputsPriorityOverRuntimeData() {
	testCases := []struct {
		name           string
		attribute      string
		userInputValue string
		runtimeValue   string
		expectedID     string
	}{
		{"username", "username", "userinput-user", "runtime-user", "user-from-userinput-123"},
		{"email", "email", "userinput@example.com", "runtime@example.com", "user-from-email-userinput-456"},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			// Both UserInputs and RuntimeData have the same key
			// UserInputs should take priority
			ctx := &core.NodeContext{
				FlowID:      "flow-123",
				UserInputs:  map[string]string{tc.attribute: tc.userInputValue},
				RuntimeData: map[string]string{tc.attribute: tc.runtimeValue},
			}

			mockBase := suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock)
			mockBase.On("HasRequiredInputs", mock.Anything, mock.Anything).Return(true)
			mockBase.On("GetRequiredInputs", mock.Anything).Return([]common.Input{
				{Identifier: tc.attribute, Type: "string", Required: true},
			})

			// The mock should be called with the UserInputs value, not the RuntimeData value
			suite.mockUserProvider.On("IdentifyUser", map[string]interface{}{
				tc.attribute: tc.userInputValue,
			}).Return(&tc.expectedID, nil)

			resp, err := suite.executor.Execute(ctx)

			assert.NoError(suite.T(), err)
			assert.NotNil(suite.T(), resp)
			assert.Equal(suite.T(), common.ExecComplete, resp.Status)
			assert.Equal(suite.T(), tc.expectedID, resp.RuntimeData[userAttributeUserID])
			suite.mockUserProvider.AssertExpectations(suite.T())
		})
	}
}

// Resolve mode tests

func (suite *IdentifyingExecutorTestSuite) TestExecuteResolve_UniqueUser() {
	ctx := &core.NodeContext{
		FlowID:       "flow-123",
		ExecutorMode: ExecutorModeResolve,
		UserInputs:   map[string]string{"username": "testuser"},
		RuntimeData:  make(map[string]string),
	}

	mockBase := suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock)
	mockBase.On("HasRequiredInputs", mock.Anything, mock.Anything).Return(true)
	mockBase.On("GetRequiredInputs", mock.Anything).Return([]common.Input{
		{Identifier: "username", Type: "string", Required: true},
	})

	suite.mockUserProvider.On("SearchUsers", map[string]interface{}{
		"username": "testuser",
	}).Return([]*userprovider.User{
		{UserID: "user-1", Attributes: json.RawMessage(`{"username":"testuser"}`)},
	}, nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	assert.Equal(suite.T(), "user-1", resp.RuntimeData[userAttributeUserID])
}

func (suite *IdentifyingExecutorTestSuite) TestExecuteResolve_AmbiguousUser() {
	ctx := &core.NodeContext{
		FlowID:       "flow-123",
		ExecutorMode: ExecutorModeResolve,
		UserInputs:   map[string]string{"username": "john"},
		RuntimeData:  make(map[string]string),
	}

	mockBase := suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock)
	mockBase.On("HasRequiredInputs", mock.Anything, mock.Anything).Return(true)
	mockBase.On("GetRequiredInputs", mock.Anything).Return([]common.Input{
		{Identifier: "username", Type: "string", Required: true},
	})

	suite.mockUserProvider.On("SearchUsers", map[string]interface{}{
		"username": "john",
	}).Return([]*userprovider.User{
		{UserID: "user-1", Attributes: json.RawMessage(`{"username":"john","firstName":"John"}`)},
		{UserID: "user-2", Attributes: json.RawMessage(`{"username":"john","firstName":"Jane"}`)},
		{UserID: "user-3", Attributes: json.RawMessage(`{"username":"john","firstName":"John"}`)},
	}, nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecUserInputRequired, resp.Status)
	assert.NotEmpty(suite.T(), resp.RuntimeData[runtimeKeyCandidateUsers])
}

func (suite *IdentifyingExecutorTestSuite) TestExecuteResolve_FilteredToOne() {
	candidates := []*userprovider.User{
		{UserID: "user-1", Attributes: json.RawMessage(`{"username":"john","firstName":"John"}`)},
		{UserID: "user-2", Attributes: json.RawMessage(`{"username":"john","firstName":"Jane"}`)},
	}
	candidatesJSON, _ := json.Marshal(candidates)

	ctx := &core.NodeContext{
		FlowID:       "flow-123",
		ExecutorMode: ExecutorModeResolve,
		UserInputs:   map[string]string{"username": "john", "firstName": "Jane"},
		RuntimeData:  map[string]string{runtimeKeyCandidateUsers: string(candidatesJSON)},
	}

	mockBase := suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock)
	mockBase.On("HasRequiredInputs", mock.Anything, mock.Anything).Return(true)
	mockBase.On("GetRequiredInputs", mock.Anything).Return([]common.Input{
		{Identifier: "username", Type: "string", Required: true},
		{Identifier: "firstName", Type: "string", Required: true},
	})

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	assert.Equal(suite.T(), "user-2", resp.RuntimeData[userAttributeUserID])
}

func (suite *IdentifyingExecutorTestSuite) TestExecuteResolve_StillAmbiguous() {
	candidates := []*userprovider.User{
		{UserID: "user-1", Attributes: json.RawMessage(`{"username":"john","firstName":"John"}`)},
		{UserID: "user-2", Attributes: json.RawMessage(`{"username":"john","firstName":"John"}`)},
		{UserID: "user-3", Attributes: json.RawMessage(`{"username":"john","firstName":"Jane"}`)},
	}
	candidatesJSON, _ := json.Marshal(candidates)

	ctx := &core.NodeContext{
		FlowID:       "flow-123",
		ExecutorMode: ExecutorModeResolve,
		UserInputs:   map[string]string{"username": "john", "firstName": "John"},
		RuntimeData:  map[string]string{runtimeKeyCandidateUsers: string(candidatesJSON)},
	}

	mockBase := suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock)
	mockBase.On("HasRequiredInputs", mock.Anything, mock.Anything).Return(true)
	mockBase.On("GetRequiredInputs", mock.Anything).Return([]common.Input{
		{Identifier: "username", Type: "string", Required: true},
		{Identifier: "firstName", Type: "string", Required: true},
	})

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecUserInputRequired, resp.Status)
	var storedCandidates []*userprovider.User
	err = json.Unmarshal([]byte(resp.RuntimeData[runtimeKeyCandidateUsers]), &storedCandidates)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), storedCandidates, 2)
}

func (suite *IdentifyingExecutorTestSuite) TestExecuteResolve_FilteredToNone() {
	candidates := []*userprovider.User{
		{UserID: "user-1", Attributes: json.RawMessage(`{"username":"john","firstName":"John"}`)},
		{UserID: "user-2", Attributes: json.RawMessage(`{"username":"john","firstName":"Jane"}`)},
	}
	candidatesJSON, _ := json.Marshal(candidates)

	ctx := &core.NodeContext{
		FlowID:       "flow-123",
		ExecutorMode: ExecutorModeResolve,
		UserInputs:   map[string]string{"username": "john", "firstName": "Nobody"},
		RuntimeData:  map[string]string{runtimeKeyCandidateUsers: string(candidatesJSON)},
	}

	mockBase := suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock)
	mockBase.On("HasRequiredInputs", mock.Anything, mock.Anything).Return(true)
	mockBase.On("GetRequiredInputs", mock.Anything).Return([]common.Input{
		{Identifier: "username", Type: "string", Required: true},
		{Identifier: "firstName", Type: "string", Required: true},
	})

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Equal(suite.T(), failureReasonUserNotFound, resp.FailureReason)
}

func (suite *IdentifyingExecutorTestSuite) TestExecuteResolve_SearchUsersNotFound() {
	ctx := &core.NodeContext{
		FlowID:       "flow-123",
		ExecutorMode: ExecutorModeResolve,
		UserInputs:   map[string]string{"username": "missing"},
		RuntimeData:  make(map[string]string),
	}

	mockBase := suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock)
	mockBase.On("HasRequiredInputs", mock.Anything, mock.Anything).Return(true)
	mockBase.On("GetRequiredInputs", mock.Anything).Return([]common.Input{
		{Identifier: "username", Type: "string", Required: true},
	})

	suite.mockUserProvider.On("SearchUsers", map[string]interface{}{
		"username": "missing",
	}).Return(nil, userprovider.NewUserProviderError(
		userprovider.ErrorCodeUserNotFound, "User not found", "No users match"))

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Equal(suite.T(), failureReasonUserNotFound, resp.FailureReason)
}

func (suite *IdentifyingExecutorTestSuite) TestExecute_IdentifyMode_AmbiguousUser() {
	ctx := &core.NodeContext{
		FlowID:       "flow-123",
		ExecutorMode: ExecutorModeIdentify,
		UserInputs:   map[string]string{"username": "john"},
	}

	mockBase := suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock)
	mockBase.On("HasRequiredInputs", mock.Anything, mock.Anything).Return(true)
	mockBase.On("GetRequiredInputs", mock.Anything).Return([]common.Input{
		{Identifier: "username", Type: "string", Required: true},
	})

	suite.mockUserProvider.On("IdentifyUser", map[string]interface{}{
		"username": "john",
	}).Return((*string)(nil), userprovider.NewUserProviderError(
		userprovider.ErrorCodeAmbiguousUser, "Ambiguous user", "Multiple users match"))

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Equal(suite.T(), failureReasonAmbiguousUser, resp.FailureReason)
}

func (suite *IdentifyingExecutorTestSuite) TestExecute_DefaultMode_AmbiguousUser() {
	ctx := &core.NodeContext{
		FlowID:     "flow-123",
		UserInputs: map[string]string{"username": "john"},
	}

	mockBase := suite.executor.ExecutorInterface.(*coremock.ExecutorInterfaceMock)
	mockBase.On("HasRequiredInputs", mock.Anything, mock.Anything).Return(true)
	mockBase.On("GetRequiredInputs", mock.Anything).Return([]common.Input{
		{Identifier: "username", Type: "string", Required: true},
	})

	suite.mockUserProvider.On("IdentifyUser", map[string]interface{}{
		"username": "john",
	}).Return((*string)(nil), userprovider.NewUserProviderError(
		userprovider.ErrorCodeAmbiguousUser, "Ambiguous user", "Multiple users match"))

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Equal(suite.T(), failureReasonAmbiguousUser, resp.FailureReason)
}

func TestFilterUsersByAttributes(t *testing.T) {
	users := []*userprovider.User{
		{UserID: "u1", Attributes: json.RawMessage(`{"firstName":"John","lastName":"Doe"}`)},
		{UserID: "u2", Attributes: json.RawMessage(`{"firstName":"Jane","lastName":"Doe"}`)},
		{UserID: "u3", Attributes: json.RawMessage(`{"firstName":"John","lastName":"Smith"}`)},
	}

	t.Run("FilterByFirstName", func(t *testing.T) {
		result := filterUsersByAttributes(users, map[string]interface{}{"firstName": "John"})
		assert.Len(t, result, 2)
	})

	t.Run("FilterByLastName", func(t *testing.T) {
		result := filterUsersByAttributes(users, map[string]interface{}{"lastName": "Doe"})
		assert.Len(t, result, 2)
	})

	t.Run("FilterByBoth", func(t *testing.T) {
		result := filterUsersByAttributes(users, map[string]interface{}{
			"firstName": "John",
			"lastName":  "Doe",
		})
		assert.Len(t, result, 1)
		assert.Equal(t, "u1", result[0].UserID)
	})

	t.Run("NoMatch", func(t *testing.T) {
		result := filterUsersByAttributes(users, map[string]interface{}{"firstName": "Nobody"})
		assert.Len(t, result, 0)
	})

	t.Run("NonSearchableInputsSkipped", func(t *testing.T) {
		result := filterUsersByAttributes(users, map[string]interface{}{
			"firstName": "John",
			"password":  "secret",
		})
		assert.Len(t, result, 2)
	})

	t.Run("EmptyAttributes", func(t *testing.T) {
		emptyUsers := []*userprovider.User{
			{UserID: "u1", Attributes: json.RawMessage(`{}`)},
		}
		result := filterUsersByAttributes(emptyUsers, map[string]interface{}{"firstName": "John"})
		assert.Len(t, result, 0)
	})

	t.Run("FilterByUserType", func(t *testing.T) {
		typedUsers := []*userprovider.User{
			{UserID: "u1", UserType: "DEFAULT", Attributes: json.RawMessage(`{"firstName":"John"}`)},
			{UserID: "u2", UserType: "ADMIN", Attributes: json.RawMessage(`{"firstName":"John"}`)},
			{UserID: "u3", UserType: "DEFAULT", Attributes: json.RawMessage(`{"firstName":"Jane"}`)},
		}
		result := filterUsersByAttributes(typedUsers, map[string]interface{}{
			"firstName": "John",
			"userType":  "DEFAULT",
		})
		assert.Len(t, result, 1)
		assert.Equal(t, "u1", result[0].UserID)
	})

	t.Run("FilterByOUHandle", func(t *testing.T) {
		ouUsers := []*userprovider.User{
			{UserID: "u1", OUHandle: "/org1", Attributes: json.RawMessage(`{"firstName":"John"}`)},
			{UserID: "u2", OUHandle: "/org2", Attributes: json.RawMessage(`{"firstName":"John"}`)},
		}
		result := filterUsersByAttributes(ouUsers, map[string]interface{}{
			"firstName": "John",
			"ouHandle":  "/org1",
		})
		assert.Len(t, result, 1)
		assert.Equal(t, "u1", result[0].UserID)
	})

	t.Run("FilterByUserTypeAndOUHandle", func(t *testing.T) {
		mixedUsers := []*userprovider.User{
			{UserID: "u1", UserType: "DEFAULT", OUHandle: "/org1", Attributes: json.RawMessage(`{"firstName":"John"}`)},
			{UserID: "u2", UserType: "DEFAULT", OUHandle: "/org2", Attributes: json.RawMessage(`{"firstName":"John"}`)},
			{UserID: "u3", UserType: "ADMIN", OUHandle: "/org1", Attributes: json.RawMessage(`{"firstName":"John"}`)},
		}
		result := filterUsersByAttributes(mixedUsers, map[string]interface{}{
			"firstName": "John",
			"userType":  "DEFAULT",
			"ouHandle":  "/org1",
		})
		assert.Len(t, result, 1)
		assert.Equal(t, "u1", result[0].UserID)
	})

	t.Run("NilAttributesWithTopLevelFilter", func(t *testing.T) {
		nilAttrUsers := []*userprovider.User{
			{UserID: "u1", UserType: "DEFAULT"},
			{UserID: "u2", UserType: "ADMIN"},
		}
		result := filterUsersByAttributes(nilAttrUsers, map[string]interface{}{"userType": "DEFAULT"})
		assert.Len(t, result, 1)
		assert.Equal(t, "u1", result[0].UserID)
	})
}
