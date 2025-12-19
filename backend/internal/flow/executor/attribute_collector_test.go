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

	authncm "github.com/asgardeo/thunder/internal/authn/common"
	flowcm "github.com/asgardeo/thunder/internal/flow/common"
	flowcore "github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/user"
	"github.com/asgardeo/thunder/tests/mocks/flow/coremock"
	"github.com/asgardeo/thunder/tests/mocks/usermock"
)

const testUserID = "user-123"

type AttributeCollectorTestSuite struct {
	suite.Suite
	mockUserService *usermock.UserServiceInterfaceMock
	mockFlowFactory *coremock.FlowFactoryInterfaceMock
	executor        *attributeCollector
}

func TestAttributeCollectorSuite(t *testing.T) {
	suite.Run(t, new(AttributeCollectorTestSuite))
}

func (suite *AttributeCollectorTestSuite) SetupTest() {
	suite.mockUserService = usermock.NewUserServiceInterfaceMock(suite.T())
	suite.mockFlowFactory = coremock.NewFlowFactoryInterfaceMock(suite.T())

	prerequisites := []flowcm.InputData{{Name: "userID", Type: "string", Required: true}}
	mockExec := createMockExecutorForAttrCollector(suite.T(), ExecutorNameAttributeCollect,
		flowcm.ExecutorTypeUtility, prerequisites)

	suite.mockFlowFactory.On("CreateExecutor", ExecutorNameAttributeCollect, flowcm.ExecutorTypeUtility,
		[]flowcm.InputData{}, prerequisites).Return(mockExec)

	suite.executor = newAttributeCollector(suite.mockFlowFactory, suite.mockUserService)
}

func createMockExecutorForAttrCollector(t *testing.T, name string,
	executorType flowcm.ExecutorType, prerequisites []flowcm.InputData) flowcore.ExecutorInterface {
	mockExec := coremock.NewExecutorInterfaceMock(t)
	mockExec.On("GetName").Return(name).Maybe()
	mockExec.On("GetType").Return(executorType).Maybe()
	mockExec.On("GetDefaultExecutorInputs").Return([]flowcm.InputData{}).Maybe()
	mockExec.On("GetPrerequisites").Return(prerequisites).Maybe()
	mockExec.On("GetRequiredData", mock.Anything).Return([]flowcm.InputData{}).Maybe()
	mockExec.On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(func(ctx *flowcore.NodeContext, execResp *flowcm.ExecutorResponse) bool {
			return ctx.RuntimeData != nil && ctx.RuntimeData[userAttributeUserID] != ""
		}).Maybe()
	mockExec.On("CheckInputData", mock.Anything, mock.Anything).
		Return(func(ctx *flowcore.NodeContext, execResp *flowcm.ExecutorResponse) bool {
			if len(ctx.NodeInputData) == 0 {
				return false
			}
			for _, input := range ctx.NodeInputData {
				if _, ok := ctx.UserInputData[input.Name]; !ok {
					if _, ok := ctx.RuntimeData[input.Name]; !ok {
						execResp.RequiredData = append(execResp.RequiredData, input)
					}
				}
			}
			return len(execResp.RequiredData) > 0
		}).Maybe()
	mockExec.On("GetUserIDFromContext", mock.Anything).
		Return(func(ctx *flowcore.NodeContext) string {
			if ctx.RuntimeData != nil {
				return ctx.RuntimeData[userAttributeUserID]
			}
			return ""
		}).Maybe()
	return mockExec
}

func (suite *AttributeCollectorTestSuite) TestNewAttributeCollector() {
	assert.NotNil(suite.T(), suite.executor)
	assert.NotNil(suite.T(), suite.executor.userService)
}

func (suite *AttributeCollectorTestSuite) TestExecute_RegistrationFlow() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeRegistration,
	}

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), flowcm.ExecComplete, resp.Status)
}

func (suite *AttributeCollectorTestSuite) TestExecute_UserNotAuthenticated() {
	ctx := &flowcore.NodeContext{
		FlowID:            "flow-123",
		FlowType:          flowcm.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{IsAuthenticated: false},
	}

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), flowcm.ExecFailure, resp.Status)
	assert.Equal(suite.T(), failureReasonUserNotAuthenticated, resp.FailureReason)
}

func (suite *AttributeCollectorTestSuite) TestExecute_PrerequisitesNotMet() {
	ctx := &flowcore.NodeContext{
		FlowID:            "flow-123",
		FlowType:          flowcm.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{IsAuthenticated: true},
		RuntimeData:       map[string]string{},
	}

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), flowcm.ExecFailure, resp.Status)
}

func (suite *AttributeCollectorTestSuite) TestExecute_UserInputRequired() {
	attrs := map[string]interface{}{"phone": "1234567890"}
	attrsJSON, _ := json.Marshal(attrs)

	existingUser := &user.User{
		ID:         testUserID,
		Attributes: attrsJSON,
	}

	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(existingUser, nil)

	ctx := &flowcore.NodeContext{
		FlowID:            "flow-123",
		FlowType:          flowcm.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{IsAuthenticated: true},
		RuntimeData:       map[string]string{userAttributeUserID: testUserID},
		NodeInputData:     []flowcm.InputData{{Name: "email", Type: "string", Required: true}},
		UserInputData:     map[string]string{},
	}

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), flowcm.ExecUserInputRequired, resp.Status)
	assert.NotEmpty(suite.T(), resp.RequiredData)
	suite.mockUserService.AssertExpectations(suite.T())
}

func (suite *AttributeCollectorTestSuite) TestExecute_Success() {
	ctx := &flowcore.NodeContext{
		FlowID:            "flow-123",
		FlowType:          flowcm.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{IsAuthenticated: true},
		RuntimeData:       map[string]string{userAttributeUserID: testUserID},
		NodeInputData:     []flowcm.InputData{{Name: "email", Type: "string", Required: true}},
		UserInputData:     map[string]string{"email": "test@example.com"},
	}

	existingUser := &user.User{
		ID:               testUserID,
		OrganizationUnit: "ou-123",
		Type:             "INTERNAL",
		Attributes:       json.RawMessage(`{}`),
	}

	updatedAttrs := map[string]interface{}{"email": "test@example.com"}
	updatedAttrsJSON, _ := json.Marshal(updatedAttrs)
	updatedUser := &user.User{
		ID:               testUserID,
		OrganizationUnit: "ou-123",
		Type:             "INTERNAL",
		Attributes:       updatedAttrsJSON,
	}

	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(existingUser, nil)
	suite.mockUserService.On("UpdateUser", mock.Anything, testUserID, mock.MatchedBy(func(u *user.User) bool {
		return u.ID == testUserID && u.Attributes != nil
	})).Return(updatedUser, nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), flowcm.ExecComplete, resp.Status)
	suite.mockUserService.AssertExpectations(suite.T())
}

func (suite *AttributeCollectorTestSuite) TestExecute_UpdateUserFails() {
	ctx := &flowcore.NodeContext{
		FlowID:            "flow-123",
		FlowType:          flowcm.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{IsAuthenticated: true},
		RuntimeData:       map[string]string{userAttributeUserID: testUserID},
		NodeInputData:     []flowcm.InputData{{Name: "email", Type: "string", Required: true}},
		UserInputData:     map[string]string{"email": "test@example.com"},
	}

	existingUser := &user.User{
		ID:               testUserID,
		OrganizationUnit: "ou-123",
		Type:             "INTERNAL",
		Attributes:       json.RawMessage(`{}`),
	}

	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(existingUser, nil)
	suite.mockUserService.On("UpdateUser", mock.Anything, testUserID, mock.Anything).
		Return(nil, &serviceerror.ServiceError{Error: "update failed"})

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), flowcm.ExecFailure, resp.Status)
	assert.Contains(suite.T(), resp.FailureReason, "Failed to update user attributes")
	suite.mockUserService.AssertExpectations(suite.T())
}

func (suite *AttributeCollectorTestSuite) TestCheckInputData_AttributesInAuthenticatedUser() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
			Attributes:      map[string]interface{}{"email": "test@example.com"},
		},
		NodeInputData: []flowcm.InputData{{Name: "email", Type: "string", Required: true}},
		RuntimeData:   map[string]string{},
	}

	execResp := &flowcm.ExecutorResponse{
		RequiredData: []flowcm.InputData{{Name: "email", Type: "string", Required: true}},
		RuntimeData:  make(map[string]string),
	}

	result := suite.executor.CheckInputData(ctx, execResp)

	assert.False(suite.T(), result)
	assert.Empty(suite.T(), execResp.RequiredData)
	assert.Equal(suite.T(), "test@example.com", execResp.RuntimeData["email"])
}

func (suite *AttributeCollectorTestSuite) TestCheckInputData_AttributesInUserProfile() {
	attrs := map[string]interface{}{"email": "profile@example.com"}
	attrsJSON, _ := json.Marshal(attrs)

	ctx := &flowcore.NodeContext{
		FlowID:            "flow-123",
		FlowType:          flowcm.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{IsAuthenticated: true},
		RuntimeData:       map[string]string{userAttributeUserID: testUserID},
		NodeInputData:     []flowcm.InputData{{Name: "email", Type: "string", Required: true}},
	}

	execResp := &flowcm.ExecutorResponse{
		RequiredData: []flowcm.InputData{{Name: "email", Type: "string", Required: true}},
		RuntimeData:  make(map[string]string),
	}

	existingUser := &user.User{
		ID:         testUserID,
		Attributes: attrsJSON,
	}

	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(existingUser, nil)

	result := suite.executor.CheckInputData(ctx, execResp)

	assert.False(suite.T(), result)
	assert.Empty(suite.T(), execResp.RequiredData)
	assert.Equal(suite.T(), "profile@example.com", execResp.RuntimeData["email"])
	suite.mockUserService.AssertExpectations(suite.T())
}

func (suite *AttributeCollectorTestSuite) TestGetUserAttributes_Success() {
	attrs := map[string]interface{}{"email": "test@example.com", "phone": "1234567890"}
	attrsJSON, _ := json.Marshal(attrs)

	ctx := &flowcore.NodeContext{
		RuntimeData: map[string]string{userAttributeUserID: testUserID},
	}

	existingUser := &user.User{
		ID:         testUserID,
		Attributes: attrsJSON,
	}

	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(existingUser, nil)

	result, err := suite.executor.getUserAttributes(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "test@example.com", result["email"])
	assert.Equal(suite.T(), "1234567890", result["phone"])
	suite.mockUserService.AssertExpectations(suite.T())
}

func (suite *AttributeCollectorTestSuite) TestGetUserAttributes_UserNotFound() {
	ctx := &flowcore.NodeContext{
		RuntimeData: map[string]string{userAttributeUserID: testUserID},
	}

	suite.mockUserService.On("GetUser", mock.Anything, testUserID).
		Return(nil, &serviceerror.ServiceError{Error: "user not found"})

	result, err := suite.executor.getUserAttributes(ctx)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	suite.mockUserService.AssertExpectations(suite.T())
}

func (suite *AttributeCollectorTestSuite) TestGetUserAttributes_InvalidJSON() {
	ctx := &flowcore.NodeContext{
		RuntimeData: map[string]string{userAttributeUserID: testUserID},
	}

	existingUser := &user.User{
		ID:         testUserID,
		Attributes: json.RawMessage(`invalid json`),
	}

	suite.mockUserService.On("GetUser", mock.Anything, testUserID).Return(existingUser, nil)

	result, err := suite.executor.getUserAttributes(ctx)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	suite.mockUserService.AssertExpectations(suite.T())
}

func (suite *AttributeCollectorTestSuite) TestGetUpdatedUserObject_NewAttributes() {
	ctx := &flowcore.NodeContext{
		UserInputData: map[string]string{"email": "new@example.com"},
		NodeInputData: []flowcm.InputData{{Name: "email", Type: "string", Required: true}},
	}

	existingUser := &user.User{
		ID:               testUserID,
		OrganizationUnit: "ou-123",
		Type:             "INTERNAL",
		Attributes:       json.RawMessage(`{}`),
	}

	updateRequired, updatedUser, err := suite.executor.getUpdatedUserObject(ctx, existingUser)

	assert.NoError(suite.T(), err)
	assert.True(suite.T(), updateRequired)
	assert.NotNil(suite.T(), updatedUser)
	assert.Equal(suite.T(), testUserID, updatedUser.ID)

	var attrs map[string]interface{}
	err = json.Unmarshal(updatedUser.Attributes, &attrs)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "new@example.com", attrs["email"])
}

func (suite *AttributeCollectorTestSuite) TestGetUpdatedUserObject_NoNewAttributes() {
	ctx := &flowcore.NodeContext{
		UserInputData: map[string]string{},
		NodeInputData: []flowcm.InputData{{Name: "email", Type: "string", Required: true}},
	}

	existingUser := &user.User{
		ID:               testUserID,
		OrganizationUnit: "ou-123",
		Type:             "INTERNAL",
		Attributes:       json.RawMessage(`{"existing": "value"}`),
	}

	updateRequired, updatedUser, err := suite.executor.getUpdatedUserObject(ctx, existingUser)

	assert.NoError(suite.T(), err)
	assert.False(suite.T(), updateRequired)
	assert.Equal(suite.T(), existingUser, updatedUser)
}

func (suite *AttributeCollectorTestSuite) TestGetUpdatedUserObject_MergeAttributes() {
	existingAttrs := map[string]interface{}{"existing": "value"}
	existingAttrsJSON, _ := json.Marshal(existingAttrs)

	ctx := &flowcore.NodeContext{
		UserInputData: map[string]string{"email": "new@example.com"},
		NodeInputData: []flowcm.InputData{{Name: "email", Type: "string", Required: true}},
	}

	existingUser := &user.User{
		ID:               testUserID,
		OrganizationUnit: "ou-123",
		Type:             "INTERNAL",
		Attributes:       existingAttrsJSON,
	}

	updateRequired, updatedUser, err := suite.executor.getUpdatedUserObject(ctx, existingUser)

	assert.NoError(suite.T(), err)
	assert.True(suite.T(), updateRequired)
	assert.NotNil(suite.T(), updatedUser)

	var attrs map[string]interface{}
	err = json.Unmarshal(updatedUser.Attributes, &attrs)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "value", attrs["existing"])
	assert.Equal(suite.T(), "new@example.com", attrs["email"])
}

func (suite *AttributeCollectorTestSuite) TestGetInputAttributes_FromUserInput() {
	ctx := &flowcore.NodeContext{
		UserInputData: map[string]string{"email": "test@example.com", "phone": "1234567890"},
		RuntimeData:   map[string]string{},
		NodeInputData: []flowcm.InputData{
			{Name: "email", Type: "string", Required: true},
			{Name: "phone", Type: "string", Required: true},
		},
	}

	result := suite.executor.getInputAttributes(ctx)

	assert.Len(suite.T(), result, 2)
	assert.Equal(suite.T(), "test@example.com", result["email"])
	assert.Equal(suite.T(), "1234567890", result["phone"])
}

func (suite *AttributeCollectorTestSuite) TestGetInputAttributes_FromRuntimeData() {
	ctx := &flowcore.NodeContext{
		UserInputData: map[string]string{},
		RuntimeData:   map[string]string{"email": "runtime@example.com"},
		NodeInputData: []flowcm.InputData{{Name: "email", Type: "string", Required: true}},
	}

	result := suite.executor.getInputAttributes(ctx)

	assert.Len(suite.T(), result, 1)
	assert.Equal(suite.T(), "runtime@example.com", result["email"])
}

func (suite *AttributeCollectorTestSuite) TestGetInputAttributes_SkipUserID() {
	ctx := &flowcore.NodeContext{
		UserInputData: map[string]string{"userID": testUserID, "email": "test@example.com"},
		RuntimeData:   map[string]string{},
		NodeInputData: []flowcm.InputData{
			{Name: "userID", Type: "string", Required: true},
			{Name: "email", Type: "string", Required: true},
		},
	}

	result := suite.executor.getInputAttributes(ctx)

	assert.Len(suite.T(), result, 1)
	assert.Equal(suite.T(), "test@example.com", result["email"])
	assert.NotContains(suite.T(), result, "userID")
}
