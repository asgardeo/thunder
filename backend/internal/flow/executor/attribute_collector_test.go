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
	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/userprovider"
	"github.com/asgardeo/thunder/tests/mocks/flow/coremock"
	"github.com/asgardeo/thunder/tests/mocks/userprovidermock"
)

const testUserID = "user-123"

type AttributeCollectorTestSuite struct {
	suite.Suite
	mockUserProvider *userprovidermock.UserProviderInterfaceMock
	mockFlowFactory  *coremock.FlowFactoryInterfaceMock
	executor         *attributeCollector
}

func TestAttributeCollectorSuite(t *testing.T) {
	suite.Run(t, new(AttributeCollectorTestSuite))
}

func (suite *AttributeCollectorTestSuite) SetupTest() {
	suite.mockUserProvider = userprovidermock.NewUserProviderInterfaceMock(suite.T())
	suite.mockFlowFactory = coremock.NewFlowFactoryInterfaceMock(suite.T())

	prerequisites := []common.Input{{Identifier: "userID", Type: "string", Required: true}}
	mockExec := createMockExecutorForAttrCollector(suite.T(), ExecutorNameAttributeCollect,
		common.ExecutorTypeUtility, prerequisites)

	suite.mockFlowFactory.On("CreateExecutor", ExecutorNameAttributeCollect, common.ExecutorTypeUtility,
		[]common.Input{}, prerequisites).Return(mockExec)

	suite.executor = newAttributeCollector(suite.mockFlowFactory, suite.mockUserProvider)
}

func createMockExecutorForAttrCollector(t *testing.T, name string,
	executorType common.ExecutorType, prerequisites []common.Input) core.ExecutorInterface {
	mockExec := coremock.NewExecutorInterfaceMock(t)
	mockExec.On("GetName").Return(name).Maybe()
	mockExec.On("GetType").Return(executorType).Maybe()
	mockExec.On("GetDefaultInputs").Return([]common.Input{}).Maybe()
	mockExec.On("GetPrerequisites").Return(prerequisites).Maybe()
	mockExec.On("GetInputs", mock.Anything).Return([]common.Input{}).Maybe()
	mockExec.On("ValidatePrerequisites", mock.Anything, mock.Anything).
		Return(func(ctx *core.NodeContext, execResp *common.ExecutorResponse) bool {
			return ctx.RuntimeData != nil && ctx.RuntimeData[userAttributeUserID] != ""
		}).Maybe()
	mockExec.On("HasRequiredInputs", mock.Anything, mock.Anything).
		Return(func(ctx *core.NodeContext, execResp *common.ExecutorResponse) bool {
			if len(ctx.NodeInputs) == 0 {
				return true
			}
			for _, input := range ctx.NodeInputs {
				if _, ok := ctx.UserInputs[input.Identifier]; !ok {
					if _, ok := ctx.RuntimeData[input.Identifier]; !ok {
						execResp.Inputs = append(execResp.Inputs, input)
					}
				}
			}
			return len(execResp.Inputs) == 0
		}).Maybe()
	mockExec.On("GetUserIDFromContext", mock.Anything).
		Return(func(ctx *core.NodeContext) string {
			if ctx.RuntimeData != nil {
				return ctx.RuntimeData[userAttributeUserID]
			}
			return ""
		}).Maybe()
	return mockExec
}

func (suite *AttributeCollectorTestSuite) TestNewAttributeCollector() {
	assert.NotNil(suite.T(), suite.executor)
	assert.NotNil(suite.T(), suite.executor.userProvider)
}

func (suite *AttributeCollectorTestSuite) TestExecute_RegistrationFlow() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeRegistration,
	}

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
}

func (suite *AttributeCollectorTestSuite) TestExecute_UserNotAuthenticated() {
	ctx := &core.NodeContext{
		FlowID:            "flow-123",
		FlowType:          common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{IsAuthenticated: false},
	}

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Equal(suite.T(), failureReasonUserNotAuthenticated, resp.FailureReason)
}

func (suite *AttributeCollectorTestSuite) TestExecute_PrerequisitesNotMet() {
	ctx := &core.NodeContext{
		FlowID:            "flow-123",
		FlowType:          common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{IsAuthenticated: true},
		RuntimeData:       map[string]string{},
	}

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
}

func (suite *AttributeCollectorTestSuite) TestExecute_UserInputRequired() {
	attrs := map[string]interface{}{"phone": "1234567890"}
	attrsJSON, _ := json.Marshal(attrs)

	existingUser := &userprovider.User{
		UserID:     testUserID,
		Attributes: attrsJSON,
	}

	suite.mockUserProvider.On("GetUser", testUserID).Return(existingUser, nil)

	ctx := &core.NodeContext{
		FlowID:            "flow-123",
		FlowType:          common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{IsAuthenticated: true},
		RuntimeData:       map[string]string{userAttributeUserID: testUserID},
		NodeInputs:        []common.Input{{Identifier: "email", Type: "string", Required: true}},
		UserInputs:        map[string]string{},
	}

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecUserInputRequired, resp.Status)
	assert.NotEmpty(suite.T(), resp.Inputs)
	suite.mockUserProvider.AssertExpectations(suite.T())
}

func (suite *AttributeCollectorTestSuite) TestExecute_Success() {
	ctx := &core.NodeContext{
		FlowID:            "flow-123",
		FlowType:          common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{IsAuthenticated: true},
		RuntimeData:       map[string]string{userAttributeUserID: testUserID},
		NodeInputs:        []common.Input{{Identifier: "email", Type: "string", Required: true}},
		UserInputs:        map[string]string{"email": "test@example.com"},
	}

	existingUser := &userprovider.User{
		UserID:             testUserID,
		OrganizationUnitID: "ou-123",
		UserType:           "INTERNAL",
		Attributes:         json.RawMessage(`{}`),
	}

	updatedAttrs := map[string]interface{}{"email": "test@example.com"}
	updatedAttrsJSON, _ := json.Marshal(updatedAttrs)
	updatedUser := &userprovider.User{
		UserID:             testUserID,
		OrganizationUnitID: "ou-123",
		UserType:           "INTERNAL",
		Attributes:         updatedAttrsJSON,
	}

	suite.mockUserProvider.On("GetUser", testUserID).Return(existingUser, nil)
	suite.mockUserProvider.On("UpdateUser", testUserID, mock.MatchedBy(func(u *userprovider.User) bool {
		return u.UserID == testUserID && u.Attributes != nil
	})).Return(updatedUser, nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	suite.mockUserProvider.AssertExpectations(suite.T())
}

func (suite *AttributeCollectorTestSuite) TestExecute_UpdateUserFails() {
	ctx := &core.NodeContext{
		FlowID:            "flow-123",
		FlowType:          common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{IsAuthenticated: true},
		RuntimeData:       map[string]string{userAttributeUserID: testUserID},
		NodeInputs:        []common.Input{{Identifier: "email", Type: "string", Required: true}},
		UserInputs:        map[string]string{"email": "test@example.com"},
	}

	existingUser := &userprovider.User{
		UserID:             testUserID,
		OrganizationUnitID: "ou-123",
		UserType:           "INTERNAL",
		Attributes:         json.RawMessage(`{}`),
	}

	suite.mockUserProvider.On("GetUser", testUserID).Return(existingUser, nil)
	suite.mockUserProvider.On("UpdateUser", testUserID, mock.Anything).
		Return(nil, &userprovider.UserProviderError{Message: "update failed"})

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Contains(suite.T(), resp.FailureReason, "Failed to update user attributes")
	suite.mockUserProvider.AssertExpectations(suite.T())
}

func (suite *AttributeCollectorTestSuite) TestHasRequiredInputs_AttributesInAuthenticatedUser() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
			Attributes:      map[string]interface{}{"email": "test@example.com"},
		},
		NodeInputs:  []common.Input{{Identifier: "email", Type: "string", Required: true}},
		RuntimeData: map[string]string{},
	}

	execResp := &common.ExecutorResponse{
		Inputs:      []common.Input{{Identifier: "email", Type: "string", Required: true}},
		RuntimeData: make(map[string]string),
	}

	result := suite.executor.HasRequiredInputs(ctx, execResp)

	assert.True(suite.T(), result)
	assert.Empty(suite.T(), execResp.Inputs)
	assert.Equal(suite.T(), "test@example.com", execResp.RuntimeData["email"])
}

func (suite *AttributeCollectorTestSuite) TestHasRequiredInputs_AttributesInUserProfile() {
	attrs := map[string]interface{}{"email": "profile@example.com"}
	attrsJSON, _ := json.Marshal(attrs)

	ctx := &core.NodeContext{
		FlowID:            "flow-123",
		FlowType:          common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{IsAuthenticated: true},
		RuntimeData:       map[string]string{userAttributeUserID: testUserID},
		NodeInputs:        []common.Input{{Identifier: "email", Type: "string", Required: true}},
	}

	execResp := &common.ExecutorResponse{
		Inputs:      []common.Input{{Identifier: "email", Type: "string", Required: true}},
		RuntimeData: make(map[string]string),
	}

	existingUser := &userprovider.User{
		UserID:     testUserID,
		Attributes: attrsJSON,
	}

	suite.mockUserProvider.On("GetUser", testUserID).Return(existingUser, nil)

	result := suite.executor.HasRequiredInputs(ctx, execResp)

	assert.True(suite.T(), result)
	assert.Empty(suite.T(), execResp.Inputs)
	assert.Equal(suite.T(), "profile@example.com", execResp.RuntimeData["email"])
	suite.mockUserProvider.AssertExpectations(suite.T())
}

func (suite *AttributeCollectorTestSuite) TestGetUserAttributes_Success() {
	attrs := map[string]interface{}{"email": "test@example.com", "phone": "1234567890"}
	attrsJSON, _ := json.Marshal(attrs)

	ctx := &core.NodeContext{
		RuntimeData: map[string]string{userAttributeUserID: testUserID},
	}

	existingUser := &userprovider.User{
		UserID:     testUserID,
		Attributes: attrsJSON,
	}

	suite.mockUserProvider.On("GetUser", testUserID).Return(existingUser, nil)

	result, err := suite.executor.getUserAttributes(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "test@example.com", result["email"])
	assert.Equal(suite.T(), "1234567890", result["phone"])
	suite.mockUserProvider.AssertExpectations(suite.T())
}

func (suite *AttributeCollectorTestSuite) TestGetUserAttributes_UserNotFound() {
	ctx := &core.NodeContext{
		RuntimeData: map[string]string{userAttributeUserID: testUserID},
	}

	suite.mockUserProvider.On("GetUser", testUserID).
		Return(nil, &userprovider.UserProviderError{Message: "user not found"})

	result, err := suite.executor.getUserAttributes(ctx)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	suite.mockUserProvider.AssertExpectations(suite.T())
}

func (suite *AttributeCollectorTestSuite) TestGetUserAttributes_InvalidJSON() {
	ctx := &core.NodeContext{
		RuntimeData: map[string]string{userAttributeUserID: testUserID},
	}

	existingUser := &userprovider.User{
		UserID:     testUserID,
		Attributes: json.RawMessage(`invalid json`),
	}

	suite.mockUserProvider.On("GetUser", testUserID).Return(existingUser, nil)

	result, err := suite.executor.getUserAttributes(ctx)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	suite.mockUserProvider.AssertExpectations(suite.T())
}

func (suite *AttributeCollectorTestSuite) TestGetUpdatedUserObject_NewAttributes() {
	ctx := &core.NodeContext{
		UserInputs: map[string]string{"email": "new@example.com"},
		NodeInputs: []common.Input{{Identifier: "email", Type: "string", Required: true}},
	}

	existingUser := &userprovider.User{
		UserID:             testUserID,
		OrganizationUnitID: "ou-123",
		UserType:           "INTERNAL",
		Attributes:         json.RawMessage(`{}`),
	}

	updateRequired, updatedUser, err := suite.executor.getUpdatedUserObject(ctx, existingUser)

	assert.NoError(suite.T(), err)
	assert.True(suite.T(), updateRequired)
	assert.NotNil(suite.T(), updatedUser)
	assert.Equal(suite.T(), testUserID, updatedUser.UserID)

	var attrs map[string]interface{}
	err = json.Unmarshal(updatedUser.Attributes, &attrs)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "new@example.com", attrs["email"])
}

func (suite *AttributeCollectorTestSuite) TestGetUpdatedUserObject_NoNewAttributes() {
	ctx := &core.NodeContext{
		UserInputs: map[string]string{},
		NodeInputs: []common.Input{{Identifier: "email", Type: "string", Required: true}},
	}

	existingUser := &userprovider.User{
		UserID:             testUserID,
		OrganizationUnitID: "ou-123",
		UserType:           "INTERNAL",
		Attributes:         json.RawMessage(`{"existing": "value"}`),
	}

	updateRequired, updatedUser, err := suite.executor.getUpdatedUserObject(ctx, existingUser)

	assert.NoError(suite.T(), err)
	assert.False(suite.T(), updateRequired)
	assert.Equal(suite.T(), existingUser, updatedUser)
}

func (suite *AttributeCollectorTestSuite) TestGetUpdatedUserObject_MergeAttributes() {
	existingAttrs := map[string]interface{}{"existing": "value"}
	existingAttrsJSON, _ := json.Marshal(existingAttrs)

	ctx := &core.NodeContext{
		UserInputs: map[string]string{"email": "new@example.com"},
		NodeInputs: []common.Input{{Identifier: "email", Type: "string", Required: true}},
	}

	existingUser := &userprovider.User{
		UserID:             testUserID,
		OrganizationUnitID: "ou-123",
		UserType:           "INTERNAL",
		Attributes:         existingAttrsJSON,
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
	ctx := &core.NodeContext{
		UserInputs:  map[string]string{"email": "test@example.com", "phone": "1234567890"},
		RuntimeData: map[string]string{},
		NodeInputs: []common.Input{
			{Identifier: "email", Type: "string", Required: true},
			{Identifier: "phone", Type: "string", Required: true},
		},
	}

	result := suite.executor.getInputAttributes(ctx)

	assert.Len(suite.T(), result, 2)
	assert.Equal(suite.T(), "test@example.com", result["email"])
	assert.Equal(suite.T(), "1234567890", result["phone"])
}

func (suite *AttributeCollectorTestSuite) TestGetInputAttributes_FromRuntimeData() {
	ctx := &core.NodeContext{
		UserInputs:  map[string]string{},
		RuntimeData: map[string]string{"email": "runtime@example.com"},
		NodeInputs:  []common.Input{{Identifier: "email", Type: "string", Required: true}},
	}

	result := suite.executor.getInputAttributes(ctx)

	assert.Len(suite.T(), result, 1)
	assert.Equal(suite.T(), "runtime@example.com", result["email"])
}

func (suite *AttributeCollectorTestSuite) TestGetInputAttributes_SkipUserID() {
	ctx := &core.NodeContext{
		UserInputs:  map[string]string{"userID": testUserID, "email": "test@example.com"},
		RuntimeData: map[string]string{},
		NodeInputs: []common.Input{
			{Identifier: "userID", Type: "string", Required: true},
			{Identifier: "email", Type: "string", Required: true},
		},
	}

	result := suite.executor.getInputAttributes(ctx)

	assert.Len(suite.T(), result, 1)
	assert.Equal(suite.T(), "test@example.com", result["email"])
	assert.NotContains(suite.T(), result, "userID")
}
