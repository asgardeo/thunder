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
	"github.com/asgardeo/thunder/internal/group"
	"github.com/asgardeo/thunder/internal/role"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/userprovider"
	"github.com/asgardeo/thunder/tests/mocks/flow/coremock"
	"github.com/asgardeo/thunder/tests/mocks/groupmock"
	"github.com/asgardeo/thunder/tests/mocks/rolemock"
	"github.com/asgardeo/thunder/tests/mocks/userprovidermock"
)

const (
	testUserType  = "INTERNAL"
	testNewUserID = "user-new"
)

type ProvisioningExecutorTestSuite struct {
	suite.Suite
	mockGroupService *groupmock.GroupServiceInterfaceMock
	mockRoleService  *rolemock.RoleServiceInterfaceMock
	mockFlowFactory  *coremock.FlowFactoryInterfaceMock
	mockUserProvider *userprovidermock.UserProviderInterfaceMock
	executor         *provisioningExecutor
}

func TestProvisioningExecutorSuite(t *testing.T) {
	suite.Run(t, new(ProvisioningExecutorTestSuite))
}

func (suite *ProvisioningExecutorTestSuite) SetupTest() {
	suite.mockGroupService = groupmock.NewGroupServiceInterfaceMock(suite.T())
	suite.mockRoleService = rolemock.NewRoleServiceInterfaceMock(suite.T())
	suite.mockFlowFactory = coremock.NewFlowFactoryInterfaceMock(suite.T())
	suite.mockUserProvider = userprovidermock.NewUserProviderInterfaceMock(suite.T())

	// Mock the embedded identifying executor first
	identifyingMock := suite.createMockIdentifyingExecutor()
	suite.mockFlowFactory.On("CreateExecutor", ExecutorNameIdentifying, common.ExecutorTypeUtility,
		mock.Anything, mock.Anything).Return(identifyingMock).Maybe()

	mockExec := suite.createMockProvisioningExecutor()
	suite.mockFlowFactory.On("CreateExecutor", ExecutorNameProvisioning, common.ExecutorTypeRegistration,
		[]common.Input{}, []common.Input{}).Return(mockExec)

	suite.executor = newProvisioningExecutor(suite.mockFlowFactory,
		suite.mockGroupService, suite.mockRoleService, suite.mockUserProvider)
}

func (suite *ProvisioningExecutorTestSuite) createMockIdentifyingExecutor() core.ExecutorInterface {
	mockExec := coremock.NewExecutorInterfaceMock(suite.T())
	mockExec.On("GetName").Return(ExecutorNameIdentifying).Maybe()
	mockExec.On("GetType").Return(common.ExecutorTypeUtility).Maybe()
	mockExec.On("GetDefaultInputs").Return([]common.Input{}).Maybe()
	mockExec.On("GetPrerequisites").Return([]common.Input{}).Maybe()
	return mockExec
}

func (suite *ProvisioningExecutorTestSuite) createMockProvisioningExecutor() core.ExecutorInterface {
	mockExec := coremock.NewExecutorInterfaceMock(suite.T())
	mockExec.On("GetName").Return(ExecutorNameProvisioning).Maybe()
	mockExec.On("GetType").Return(common.ExecutorTypeRegistration).Maybe()
	mockExec.On("GetDefaultInputs").Return([]common.Input{}).Maybe()
	mockExec.On("GetPrerequisites").Return([]common.Input{}).Maybe()
	mockExec.On("HasRequiredInputs", mock.Anything, mock.Anything).Return(
		func(ctx *core.NodeContext, execResp *common.ExecutorResponse) bool {
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
	mockExec.On("GetInputs", mock.Anything).Return([]common.Input{}).Maybe()
	mockExec.On("GetRequiredInputs", mock.Anything).Return([]common.Input{}).Maybe()
	return mockExec
}

func (suite *ProvisioningExecutorTestSuite) TestExecute_NonRegistrationFlow() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
	}

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
}

func (suite *ProvisioningExecutorTestSuite) TestExecute_Success() {
	attrs := map[string]interface{}{"username": "newuser", "email": "new@example.com"}
	attrsJSON, _ := json.Marshal(attrs)

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeRegistration,
		UserInputs: map[string]string{
			"username": "newuser",
			"email":    "new@example.com",
		},
		RuntimeData: map[string]string{
			ouIDKey:     testOUID,
			userTypeKey: testUserType,
		},
		NodeInputs: []common.Input{
			{Identifier: "username", Type: "string", Required: true},
			{Identifier: "email", Type: "string", Required: true},
		},
		NodeProperties: map[string]interface{}{
			"assignGroup": "test-group-id",
			"assignRole":  "test-role-id",
		},
	}

	suite.mockUserProvider.On("IdentifyUser", map[string]interface{}{
		"username": "newuser",
		"email":    "new@example.com",
	}).Return(nil, userprovider.NewUserProviderError(userprovider.ErrorCodeUserNotFound, "", ""))

	createdUser := &userprovider.User{
		UserID:             testNewUserID,
		OrganizationUnitID: testOUID,
		UserType:           testUserType,
		Attributes:         attrsJSON,
	}

	suite.mockUserProvider.On("CreateUser", mock.MatchedBy(func(u *userprovider.User) bool {
		return u.OrganizationUnitID == testOUID && u.UserType == testUserType
	})).Return(createdUser, nil)

	// Mock group assignment
	suite.mockGroupService.On("AddGroupMembers", mock.Anything, "test-group-id",
		mock.MatchedBy(func(members []group.Member) bool {
			return len(members) == 1 &&
				members[0].ID == testNewUserID &&
				members[0].Type == group.MemberTypeUser
		})).Return(nil, nil)

	// Mock role assignment
	suite.mockRoleService.On("AddAssignments", mock.Anything, "test-role-id",
		mock.MatchedBy(func(assignments []role.RoleAssignment) bool {
			return len(assignments) == 1 &&
				assignments[0].ID == testNewUserID &&
				assignments[0].Type == role.AssigneeTypeUser
		})).Return(nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	assert.True(suite.T(), resp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), testNewUserID, resp.AuthenticatedUser.UserID)
	suite.mockUserProvider.AssertExpectations(suite.T())
	suite.mockGroupService.AssertExpectations(suite.T())
	suite.mockRoleService.AssertExpectations(suite.T())
}

func (suite *ProvisioningExecutorTestSuite) TestExecute_UserAlreadyExists() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeRegistration,
		UserInputs: map[string]string{
			"username": "existinguser",
		},
		NodeInputs: []common.Input{{Identifier: "username", Type: "string", Required: true}},
	}

	userID := "user-existing"
	suite.mockUserProvider.On("IdentifyUser", map[string]interface{}{
		"username": "existinguser",
	}).Return(&userID, nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Contains(suite.T(), resp.FailureReason, "User already exists")
	suite.mockUserProvider.AssertExpectations(suite.T())
}

func (suite *ProvisioningExecutorTestSuite) TestExecute_NoUserAttributes() {
	ctx := &core.NodeContext{
		FlowID:     "flow-123",
		FlowType:   common.FlowTypeRegistration,
		UserInputs: map[string]string{},
		NodeInputs: []common.Input{},
	}

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Contains(suite.T(), resp.FailureReason, "No user attributes provided")
}

func (suite *ProvisioningExecutorTestSuite) TestExecute_CreateUserFails() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeRegistration,
		UserInputs: map[string]string{
			"username": "newuser",
		},
		RuntimeData: map[string]string{
			ouIDKey:     testOUID,
			userTypeKey: testUserType,
		},
		NodeInputs: []common.Input{{Identifier: "username", Type: "string", Required: true}},
	}

	suite.mockUserProvider.On("IdentifyUser", mock.Anything).Return(nil,
		userprovider.NewUserProviderError(userprovider.ErrorCodeUserNotFound, "", ""))
	suite.mockUserProvider.On("CreateUser", mock.Anything).
		Return(nil, userprovider.NewUserProviderError(userprovider.ErrorCodeSystemError, "creation failed", ""))

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Contains(suite.T(), resp.FailureReason, "Failed to create user")
	suite.mockUserProvider.AssertExpectations(suite.T())
}

func (suite *ProvisioningExecutorTestSuite) TestHasRequiredInputs_AttributesFromAuthUser() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeRegistration,
		AuthenticatedUser: authncm.AuthenticatedUser{
			Attributes: map[string]interface{}{"email": "test@example.com"},
		},
		NodeInputs: []common.Input{{Identifier: "email", Type: "string", Required: true}},
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

func (suite *ProvisioningExecutorTestSuite) TestGetAttributesForProvisioning_FromMultipleSources() {
	ctx := &core.NodeContext{
		UserInputs:  map[string]string{"username": "testuser", "code": "auth-code"},
		RuntimeData: map[string]string{"email": "test@example.com"},
		NodeInputs:  []common.Input{},
	}

	result := suite.executor.getAttributesForProvisioning(ctx)

	assert.Contains(suite.T(), result, "username")
	assert.Contains(suite.T(), result, "email")
	assert.NotContains(suite.T(), result, "code")
}

func (suite *ProvisioningExecutorTestSuite) TestGetAttributesForProvisioning_FilterNonUserAttributes() {
	ctx := &core.NodeContext{
		UserInputs: map[string]string{
			"username": "testuser",
			"userID":   "user-123",
			"code":     "auth-code",
			"nonce":    "test-nonce",
		},
		NodeInputs: []common.Input{},
	}

	result := suite.executor.getAttributesForProvisioning(ctx)

	assert.Contains(suite.T(), result, "username")
	assert.NotContains(suite.T(), result, "userID")
	assert.NotContains(suite.T(), result, "code")
	assert.NotContains(suite.T(), result, "nonce")
}

func (suite *ProvisioningExecutorTestSuite) TestGetAttributesForProvisioning_WithAuthenticatedUserAttributes() {
	ctx := &core.NodeContext{
		UserInputs: map[string]string{"username": "testuser"},
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
			UserID:          "user-123",
			Attributes: map[string]interface{}{
				"email":       "authenticated@example.com",
				"given_name":  "Test",
				"family_name": "User",
			},
		},
		RuntimeData: map[string]string{"phone": "+1234567890"},
		NodeInputs:  []common.Input{},
	}

	result := suite.executor.getAttributesForProvisioning(ctx)

	// Should include attributes from all three sources
	assert.Contains(suite.T(), result, "username")
	assert.Contains(suite.T(), result, "email")
	assert.Contains(suite.T(), result, "given_name")
	assert.Contains(suite.T(), result, "family_name")
	assert.Contains(suite.T(), result, "phone")
	assert.Equal(suite.T(), "testuser", result["username"])
	assert.Equal(suite.T(), "authenticated@example.com", result["email"])
	assert.Equal(suite.T(), "Test", result["given_name"])
}

func (suite *ProvisioningExecutorTestSuite) TestGetAttributesForProvisioning_AttributePriority() {
	ctx := &core.NodeContext{
		UserInputs: map[string]string{
			"email": "userinput@example.com",
		},
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
			Attributes: map[string]interface{}{
				"email": "authenticated@example.com",
				"name":  "Authenticated Name",
			},
		},
		RuntimeData: map[string]string{
			"email": "runtime@example.com",
			"phone": "+1234567890",
		},
		NodeInputs: []common.Input{},
	}

	result := suite.executor.getAttributesForProvisioning(ctx)

	// RuntimeData comes last in the loop, so it overwrites for 'email'
	assert.Equal(suite.T(), "runtime@example.com", result["email"])
	// AuthenticatedUser.Attributes should provide 'name' (not in other sources)
	assert.Equal(suite.T(), "Authenticated Name", result["name"])
	// RuntimeData should provide 'phone' (not in other sources)
	assert.Equal(suite.T(), "+1234567890", result["phone"])
}

func (suite *ProvisioningExecutorTestSuite) TestGetAttributesForProvisioning_WithRequiredInputs_FromAuthUser() {
	ctx := &core.NodeContext{
		UserInputs: map[string]string{
			"username": "testuser",
		},
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
			Attributes: map[string]interface{}{
				"email":      "authenticated@example.com",
				"given_name": "Test",
			},
		},
		RuntimeData: map[string]string{
			"phone": "+1234567890",
		},
		NodeInputs: []common.Input{
			{Identifier: "username", Type: "string", Required: true},
			{Identifier: "email", Type: "string", Required: true},
			{Identifier: "phone", Type: "string", Required: false},
		},
	}

	result := suite.executor.getAttributesForProvisioning(ctx)

	// Note: GetInputs is mocked to return empty, so this test behaves like no required inputs
	// All attributes from all sources will be included
	assert.Contains(suite.T(), result, "username")
	assert.Contains(suite.T(), result, "email")
	assert.Contains(suite.T(), result, "phone")
	assert.Contains(suite.T(), result, "given_name") // Will be included since GetInputs returns empty
}

func (suite *ProvisioningExecutorTestSuite) TestGetAttributesForProvisioning_WithRequiredInputs_Priority() {
	ctx := &core.NodeContext{
		UserInputs: map[string]string{
			"email": "userinput@example.com",
		},
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
			Attributes: map[string]interface{}{
				"email": "authenticated@example.com",
				"phone": "+9999999999",
			},
		},
		RuntimeData: map[string]string{
			"phone": "+1234567890",
		},
		NodeInputs: []common.Input{
			{Identifier: "email", Type: "string", Required: true},
			{Identifier: "phone", Type: "string", Required: true},
		},
	}

	result := suite.executor.getAttributesForProvisioning(ctx)

	// Note: GetInputs is mocked to return empty, so RuntimeData overwrites
	// RuntimeData comes last in the loop and overwrites for 'phone'
	assert.Equal(suite.T(), "+1234567890", result["phone"])
	// email exists in all three, RuntimeData wins (no 'email' in RuntimeData, so AuthenticatedUser wins)
	assert.Equal(suite.T(), "authenticated@example.com", result["email"])
}

func (suite *ProvisioningExecutorTestSuite) TestGetAttributesForProvisioning_FilterNonUserAttributesFromAuthUser() {
	ctx := &core.NodeContext{
		UserInputs: map[string]string{},
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
			Attributes: map[string]interface{}{
				"email":  "authenticated@example.com",
				"userID": "should-be-filtered",
				"code":   "should-be-filtered",
				"nonce":  "should-be-filtered",
			},
		},
		RuntimeData: map[string]string{},
		NodeInputs:  []common.Input{},
	}

	result := suite.executor.getAttributesForProvisioning(ctx)

	assert.Contains(suite.T(), result, "email")
	assert.NotContains(suite.T(), result, "userID")
	assert.NotContains(suite.T(), result, "code")
	assert.NotContains(suite.T(), result, "nonce")
}

func (suite *ProvisioningExecutorTestSuite) TestExecute_SkipProvisioning_UserAlreadyExists() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeRegistration,
		UserInputs: map[string]string{
			"username": "existinguser",
		},
		RuntimeData: map[string]string{
			common.RuntimeKeySkipProvisioning: dataValueTrue,
		},
		NodeInputs: []common.Input{
			{Identifier: "username", Type: "string", Required: true},
		},
	}

	userID := "existing-user-123"
	attrs := map[string]interface{}{
		"username": "existinguser",
	}
	suite.mockUserProvider.On("IdentifyUser", attrs).Return(&userID, nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	assert.Equal(suite.T(), "existing-user-123", resp.RuntimeData[userAttributeUserID])
	// Verify that CreateUser was not called (provisioning was skipped)
	// Verify that CreateUser was not called (provisioning was skipped)
	suite.mockUserProvider.AssertExpectations(suite.T())
	suite.mockUserProvider.AssertNotCalled(suite.T(), "CreateUser")
}

func (suite *ProvisioningExecutorTestSuite) TestExecute_SkipProvisioning_ProceedsNormally() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeRegistration,
		UserInputs: map[string]string{
			"username": "newuser",
			"email":    "new@example.com",
		},
		RuntimeData: map[string]string{
			common.RuntimeKeySkipProvisioning: "false",
			ouIDKey:                           testOUID,
			userTypeKey:                       testUserType,
		},
		NodeInputs: []common.Input{
			{Identifier: "username", Type: "string", Required: true},
			{Identifier: "email", Type: "string", Required: true},
		},
		// No NodeProperties - should skip group/role assignment
	}

	attrs := map[string]interface{}{
		"username": "newuser",
		"email":    "new@example.com",
	}
	attrsJSON, _ := json.Marshal(attrs)

	createdUser := &userprovider.User{
		UserID:             testNewUserID,
		OrganizationUnitID: testOUID,
		UserType:           testUserType,
		Attributes:         attrsJSON,
	}

	suite.mockUserProvider.On("IdentifyUser", attrs).Return(nil,
		userprovider.NewUserProviderError(userprovider.ErrorCodeUserNotFound, "", ""))
	suite.mockUserProvider.On("CreateUser", mock.MatchedBy(func(u *userprovider.User) bool {
		return u.OrganizationUnitID == testOUID && u.UserType == testUserType
	})).Return(createdUser, nil)

	// No group/role assignment mocks - assignments should be skipped

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	assert.True(suite.T(), resp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), testNewUserID, resp.AuthenticatedUser.UserID)
	// userAutoProvisioned flag is not set in registration flows
	assert.Equal(suite.T(), testNewUserID, resp.AuthenticatedUser.UserID)
	// userAutoProvisioned flag is not set in registration flows
	suite.mockUserProvider.AssertExpectations(suite.T())

	// Verify no group/role methods were called
	suite.mockGroupService.AssertNotCalled(suite.T(), "GetGroup")
	suite.mockRoleService.AssertNotCalled(suite.T(), "AddAssignments")
}

func (suite *ProvisioningExecutorTestSuite) TestExecute_UserEligibleForProvisioning() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			"username": "provisioneduser",
			"email":    "provisioned@example.com",
		},
		RuntimeData: map[string]string{
			common.RuntimeKeyUserEligibleForProvisioning: dataValueTrue,
			ouIDKey:     testOUID,
			userTypeKey: testUserType,
		},
		NodeInputs: []common.Input{
			{Identifier: "username", Type: "string", Required: true},
			{Identifier: "email", Type: "string", Required: true},
		},
	}

	attrs := map[string]interface{}{
		"username": "provisioneduser",
		"email":    "provisioned@example.com",
	}
	attrsJSON, _ := json.Marshal(attrs)

	createdUser := &userprovider.User{
		UserID:             "user-provisioned",
		OrganizationUnitID: testOUID,
		UserType:           testUserType,
		Attributes:         attrsJSON,
	}

	suite.mockUserProvider.On("IdentifyUser", attrs).Return(nil,
		userprovider.NewUserProviderError(userprovider.ErrorCodeUserNotFound, "", ""))
	suite.mockUserProvider.On("CreateUser", mock.MatchedBy(func(u *userprovider.User) bool {
		return u.OrganizationUnitID == testOUID && u.UserType == testUserType
	})).Return(createdUser, nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	assert.True(suite.T(), resp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), "user-provisioned", resp.AuthenticatedUser.UserID)
	assert.Equal(suite.T(), dataValueTrue, resp.RuntimeData[common.RuntimeKeyUserAutoProvisioned])
	suite.mockUserProvider.AssertExpectations(suite.T())
}

func (suite *ProvisioningExecutorTestSuite) TestExecute_UserAutoProvisionedFlag_SetAfterCreation() {
	attrs := map[string]interface{}{"username": "newuser", "email": "new@example.com"}
	attrsJSON, _ := json.Marshal(attrs)

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			"username": "newuser",
			"email":    "new@example.com",
		},
		RuntimeData: map[string]string{
			ouIDKey:     testOUID,
			userTypeKey: testUserType,
			common.RuntimeKeyUserEligibleForProvisioning: dataValueTrue,
		},
		NodeInputs: []common.Input{
			{Identifier: "username", Type: "string", Required: true},
			{Identifier: "email", Type: "string", Required: true},
		},
	}

	createdUser := &userprovider.User{
		UserID:             testNewUserID,
		OrganizationUnitID: testOUID,
		UserType:           testUserType,
		Attributes:         attrsJSON,
	}

	suite.mockUserProvider.On("IdentifyUser", attrs).Return(nil,
		userprovider.NewUserProviderError(userprovider.ErrorCodeUserNotFound, "", ""))
	suite.mockUserProvider.On("CreateUser", mock.Anything).Return(createdUser, nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	assert.Equal(suite.T(), dataValueTrue, resp.RuntimeData[common.RuntimeKeyUserAutoProvisioned],
		"userAutoProvisioned flag should be set to true after successful provisioning")
	suite.mockUserProvider.AssertExpectations(suite.T())
}

func (suite *ProvisioningExecutorTestSuite) TestAppendNonIdentifyingAttributes() {
	tests := []struct {
		name               string
		userInputs         map[string]string
		runtimeData        map[string]string
		expectedPassword   string
		shouldHavePassword bool
	}{
		{
			name: "PasswordInUserInput",
			userInputs: map[string]string{
				"username": "testuser",
				"password": "secure123",
			},
			runtimeData:        map[string]string{},
			expectedPassword:   "secure123",
			shouldHavePassword: true,
		},
		{
			name: "PasswordInRuntimeData",
			userInputs: map[string]string{
				"username": "testuser",
			},
			runtimeData: map[string]string{
				"password": "runtime-password",
			},
			expectedPassword:   "runtime-password",
			shouldHavePassword: true,
		},
		{
			name: "NoPassword",
			userInputs: map[string]string{
				"username": "testuser",
			},
			runtimeData:        map[string]string{},
			shouldHavePassword: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ctx := &core.NodeContext{
				UserInputs:  tt.userInputs,
				RuntimeData: tt.runtimeData,
			}

			attributes := map[string]interface{}{
				"username": "testuser",
			}

			suite.executor.appendNonIdentifyingAttributes(ctx, &attributes)

			if tt.shouldHavePassword {
				assert.Contains(suite.T(), attributes, "password")
				assert.Equal(suite.T(), tt.expectedPassword, attributes["password"])
			} else {
				assert.NotContains(suite.T(), attributes, "password")
				assert.Equal(suite.T(), 1, len(attributes)) // Only username
			}
		})
	}
}

func (suite *ProvisioningExecutorTestSuite) TestExecute_RegistrationFlow_SkipProvisioningWithExistingUser() {
	userID := "existing-user-id"
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeRegistration,
		UserInputs: map[string]string{
			"username": "existinguser",
		},
		RuntimeData: map[string]string{
			common.RuntimeKeySkipProvisioning: dataValueTrue,
		},
		NodeInputs: []common.Input{
			{Identifier: "username", Type: "string", Required: true},
		},
	}

	attrs := map[string]interface{}{
		"username": "existinguser",
	}
	suite.mockUserProvider.On("IdentifyUser", attrs).Return(&userID, nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	assert.Equal(suite.T(), userID, resp.RuntimeData[userAttributeUserID])
	assert.Empty(suite.T(), resp.FailureReason)
	suite.mockUserProvider.AssertNotCalled(suite.T(), "CreateUser")
	suite.mockUserProvider.AssertExpectations(suite.T())
}

func (suite *ProvisioningExecutorTestSuite) TestExecute_MissingInputs() {
	tests := []struct {
		name        string
		runtimeData map[string]string
	}{
		{
			name: "MissingOuID",
			runtimeData: map[string]string{
				userTypeKey: testUserType,
			},
		},
		{
			name: "MissingUserType",
			runtimeData: map[string]string{
				ouIDKey: testOUID,
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ctx := &core.NodeContext{
				FlowID:   "flow-123",
				FlowType: common.FlowTypeRegistration,
				UserInputs: map[string]string{
					"username": "newuser",
				},
				RuntimeData: tt.runtimeData,
				NodeInputs: []common.Input{
					{Identifier: "username", Type: "string", Required: true},
				},
			}

			attrs := map[string]interface{}{
				"username": "newuser",
			}
			suite.mockUserProvider.On("IdentifyUser", attrs).Return(nil,
				userprovider.NewUserProviderError(userprovider.ErrorCodeUserNotFound, "", ""))

			resp, err := suite.executor.Execute(ctx)

			assert.NoError(suite.T(), err)
			assert.NotNil(suite.T(), resp)
			assert.Equal(suite.T(), common.ExecFailure, resp.Status)
			assert.Equal(suite.T(), "Failed to create user", resp.FailureReason)
			suite.mockUserProvider.AssertNotCalled(suite.T(), "CreateUser")
			suite.mockUserProvider.AssertExpectations(suite.T())
		})
	}
}

func (suite *ProvisioningExecutorTestSuite) TestExecute_CreateUserFailures() {
	tests := []struct {
		name               string
		createdUser        *userprovider.User
		createUserError    *userprovider.UserProviderError
		expectedFailReason string
	}{
		{
			name:        "ServiceReturnsError",
			createdUser: nil,
			createUserError: userprovider.NewUserProviderError(
				userprovider.ErrorCodeSystemError, "Database error", ""),
			expectedFailReason: "Failed to create user",
		},
		{
			name:               "CreatedUserIsNil",
			createdUser:        nil,
			createUserError:    nil,
			expectedFailReason: "Something went wrong while creating the user",
		},
		{
			name: "CreatedUserHasEmptyID",
			createdUser: &userprovider.User{
				UserID:             "",
				OrganizationUnitID: testOUID,
				UserType:           testUserType,
				Attributes:         []byte(`{"username":"newuser"}`),
			},
			createUserError:    nil,
			expectedFailReason: "Something went wrong while creating the user",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Clear expectations before each test
			suite.mockUserProvider.ExpectedCalls = nil

			ctx := &core.NodeContext{
				FlowID:   "flow-123",
				FlowType: common.FlowTypeRegistration,
				UserInputs: map[string]string{
					"username": "newuser",
				},
				RuntimeData: map[string]string{
					ouIDKey:     testOUID,
					userTypeKey: testUserType,
				},
				NodeInputs: []common.Input{
					{Identifier: "username", Type: "string", Required: true},
				},
			}

			attrs := map[string]interface{}{
				"username": "newuser",
			}
			suite.mockUserProvider.On("IdentifyUser", attrs).Return(nil,
				userprovider.NewUserProviderError(userprovider.ErrorCodeUserNotFound, "", ""))
			suite.mockUserProvider.On("CreateUser", mock.Anything).
				Return(tt.createdUser, tt.createUserError)

			resp, err := suite.executor.Execute(ctx)

			assert.NoError(suite.T(), err)
			assert.NotNil(suite.T(), resp)
			assert.Equal(suite.T(), common.ExecFailure, resp.Status)
			assert.Equal(suite.T(), tt.expectedFailReason, resp.FailureReason)
			suite.mockUserProvider.AssertExpectations(suite.T())
		})
	}
}

func (suite *ProvisioningExecutorTestSuite) TestGetOuID() {
	tests := []struct {
		name        string
		runtimeData map[string]string
		expected    string
	}{
		{
			name: "FromOuIDKey",
			runtimeData: map[string]string{
				ouIDKey:        "ou-from-ouIDKey",
				defaultOUIDKey: "ou-from-defaultOUIDKey",
			},
			expected: "ou-from-ouIDKey",
		},
		{
			name: "FromDefaultOUIDKey",
			runtimeData: map[string]string{
				defaultOUIDKey: "ou-from-defaultOUIDKey",
			},
			expected: "ou-from-defaultOUIDKey",
		},
		{
			name:        "NotFound",
			runtimeData: map[string]string{},
			expected:    "",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ctx := &core.NodeContext{
				RuntimeData: tt.runtimeData,
			}

			ouID := suite.executor.getOuID(ctx)

			assert.Equal(suite.T(), tt.expected, ouID)
		})
	}
}

func (suite *ProvisioningExecutorTestSuite) TestGetUserType() {
	tests := []struct {
		name        string
		runtimeData map[string]string
		expected    string
	}{
		{
			name: "Found",
			runtimeData: map[string]string{
				userTypeKey: "CUSTOM_USER_TYPE",
			},
			expected: "CUSTOM_USER_TYPE",
		},
		{
			name:        "NotFound",
			runtimeData: map[string]string{},
			expected:    "",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ctx := &core.NodeContext{
				RuntimeData: tt.runtimeData,
			}

			userType := suite.executor.getUserType(ctx)

			assert.Equal(suite.T(), tt.expected, userType)
		})
	}
}

func (suite *ProvisioningExecutorTestSuite) TestHasRequiredInputs_AllAttributesInRuntimeData() {
	ctx := &core.NodeContext{
		FlowID:     "flow-123",
		UserInputs: map[string]string{},
		RuntimeData: map[string]string{
			"email":    "user@example.com",
			"username": "testuser",
		},
		AuthenticatedUser: authncm.AuthenticatedUser{
			Attributes: map[string]interface{}{},
		},
		NodeInputs: []common.Input{
			{Identifier: "email", Type: "string", Required: true},
			{Identifier: "username", Type: "string", Required: true},
		},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	inputRequired := suite.executor.HasRequiredInputs(ctx, execResp)

	assert.True(suite.T(), inputRequired)
	assert.Equal(suite.T(), 0, len(execResp.Inputs))
}

// Test group assignment failure - provisioning should fail, but role assignment should still be attempted
func (suite *ProvisioningExecutorTestSuite) TestExecute_Failure_GroupAssignmentFails() {
	attrs := map[string]interface{}{"username": "newuser"}
	attrsJSON, _ := json.Marshal(attrs)

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeRegistration,
		UserInputs: map[string]string{
			"username": "newuser",
		},
		RuntimeData: map[string]string{
			ouIDKey:     testOUID,
			userTypeKey: testUserType,
		},
		NodeInputs: []common.Input{
			{Identifier: "username", Type: "string", Required: true},
		},
		NodeProperties: map[string]interface{}{
			"assignGroup": "test-group-id",
			"assignRole":  "test-role-id",
		},
	}

	suite.mockUserProvider.On("IdentifyUser", attrs).Return(nil,
		userprovider.NewUserProviderError(userprovider.ErrorCodeUserNotFound, "", ""))

	createdUser := &userprovider.User{
		UserID:             testNewUserID,
		OrganizationUnitID: testOUID,
		UserType:           testUserType,
		Attributes:         attrsJSON,
	}

	suite.mockUserProvider.On("CreateUser", mock.Anything).Return(createdUser, nil)

	// Mock group assignment fails (e.g., group doesn't exist)
	suite.mockGroupService.On("AddGroupMembers", mock.Anything, "test-group-id", mock.Anything).
		Return(nil, &serviceerror.ServiceError{Error: "Group not found"})

	// Role assignment should still be attempted
	suite.mockRoleService.On("AddAssignments", mock.Anything, "test-role-id", mock.Anything).Return(nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Contains(suite.T(), resp.FailureReason, "Failed to assign groups and roles")
	assert.Contains(suite.T(), resp.FailureReason, "group")

	// Verify role assignment WAS attempted despite group failure
	suite.mockRoleService.AssertExpectations(suite.T())
}

// Test both group and role assignment failure - provisioning should fail with combined error
func (suite *ProvisioningExecutorTestSuite) TestExecute_Failure_BothGroupAndRoleAssignmentFail() {
	attrs := map[string]interface{}{"username": "newuser"}
	attrsJSON, _ := json.Marshal(attrs)

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeRegistration,
		UserInputs: map[string]string{
			"username": "newuser",
		},
		RuntimeData: map[string]string{
			ouIDKey:     testOUID,
			userTypeKey: testUserType,
		},
		NodeInputs: []common.Input{
			{Identifier: "username", Type: "string", Required: true},
		},
		NodeProperties: map[string]interface{}{
			"assignGroup": "test-group-id",
			"assignRole":  "test-role-id",
		},
	}

	suite.mockUserProvider.On("IdentifyUser", attrs).Return(nil,
		userprovider.NewUserProviderError(userprovider.ErrorCodeUserNotFound, "", ""))

	createdUser := &userprovider.User{
		UserID:             testNewUserID,
		OrganizationUnitID: testOUID,
		UserType:           testUserType,
		Attributes:         attrsJSON,
	}

	suite.mockUserProvider.On("CreateUser", mock.Anything).Return(createdUser, nil)

	// Mock group assignment fails
	suite.mockGroupService.On("AddGroupMembers", mock.Anything, "test-group-id", mock.Anything).
		Return(nil, &serviceerror.ServiceError{Error: "Group not found"})

	// Mock role assignment also fails
	suite.mockRoleService.On("AddAssignments", mock.Anything, "test-role-id", mock.Anything).
		Return(&serviceerror.ServiceError{Error: "Role not found"})

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Equal(suite.T(), "Failed to assign groups and roles", resp.FailureReason)

	// Verify both services were called (new behavior: try both even if one fails)
	suite.mockGroupService.AssertExpectations(suite.T())
	suite.mockRoleService.AssertExpectations(suite.T())
}

// Test role assignment failure - provisioning should fail, but group assignment succeeds
func (suite *ProvisioningExecutorTestSuite) TestExecute_Failure_RoleAssignmentFails() {
	attrs := map[string]interface{}{"username": "newuser"}
	attrsJSON, _ := json.Marshal(attrs)

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeRegistration,
		UserInputs: map[string]string{
			"username": "newuser",
		},
		RuntimeData: map[string]string{
			ouIDKey:     testOUID,
			userTypeKey: testUserType,
		},
		NodeInputs: []common.Input{
			{Identifier: "username", Type: "string", Required: true},
		},
		NodeProperties: map[string]interface{}{
			"assignGroup": "test-group-id",
			"assignRole":  "test-role-id",
		},
	}

	suite.mockUserProvider.On("IdentifyUser", attrs).Return(nil,
		userprovider.NewUserProviderError(userprovider.ErrorCodeUserNotFound, "", ""))

	createdUser := &userprovider.User{
		UserID:             testNewUserID,
		OrganizationUnitID: testOUID,
		UserType:           testUserType,
		Attributes:         attrsJSON,
	}

	suite.mockUserProvider.On("CreateUser", mock.Anything).Return(createdUser, nil)

	// Group assignment succeeds
	suite.mockGroupService.On("AddGroupMembers", mock.Anything, "test-group-id", mock.Anything).
		Return(nil, nil)

	// Role assignment fails (e.g., role doesn't exist)
	suite.mockRoleService.On("AddAssignments", mock.Anything, "test-role-id", mock.Anything).
		Return(&serviceerror.ServiceError{Error: "Role not found"})

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Contains(suite.T(), resp.FailureReason, "Failed to assign groups and roles")
	assert.Contains(suite.T(), resp.FailureReason, "role")

	// Verify both group and role services were called
	suite.mockGroupService.AssertExpectations(suite.T())
	suite.mockRoleService.AssertExpectations(suite.T())
}

// Test group with existing members - user should be appended
func (suite *ProvisioningExecutorTestSuite) TestExecute_GroupWithExistingMembers() {
	attrs := map[string]interface{}{"username": "newuser"}
	attrsJSON, _ := json.Marshal(attrs)

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeRegistration,
		UserInputs: map[string]string{
			"username": "newuser",
		},
		RuntimeData: map[string]string{
			ouIDKey:     testOUID,
			userTypeKey: testUserType,
		},
		NodeInputs: []common.Input{
			{Identifier: "username", Type: "string", Required: true},
		},
		NodeProperties: map[string]interface{}{
			"assignGroup": "test-group-id",
			"assignRole":  "test-role-id",
		},
	}

	suite.mockUserProvider.On("IdentifyUser", attrs).Return(nil,
		userprovider.NewUserProviderError(userprovider.ErrorCodeUserNotFound, "", ""))

	createdUser := &userprovider.User{
		UserID:             testNewUserID,
		OrganizationUnitID: testOUID,
		UserType:           testUserType,
		Attributes:         attrsJSON,
	}

	suite.mockUserProvider.On("CreateUser", mock.Anything).Return(createdUser, nil)

	// Mock group assignment - AddGroupMembers only adds the new user, not existing members
	suite.mockGroupService.On("AddGroupMembers", mock.Anything, "test-group-id",
		mock.MatchedBy(func(members []group.Member) bool {
			return len(members) == 1 &&
				members[0].ID == testNewUserID &&
				members[0].Type == group.MemberTypeUser
		})).Return(nil, nil)

	suite.mockRoleService.On("AddAssignments", mock.Anything, "test-role-id", mock.Anything).Return(nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	suite.mockGroupService.AssertExpectations(suite.T())
}

// Test authentication flow with auto-provisioning still assigns groups/roles
func (suite *ProvisioningExecutorTestSuite) TestExecute_AuthFlow_AutoProvisioning_AssignsGroupsAndRoles() {
	attrs := map[string]interface{}{"username": "provisioneduser"}
	attrsJSON, _ := json.Marshal(attrs)

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			"username": "provisioneduser",
		},
		RuntimeData: map[string]string{
			common.RuntimeKeyUserEligibleForProvisioning: dataValueTrue,
			ouIDKey:     testOUID,
			userTypeKey: testUserType,
		},
		NodeInputs: []common.Input{
			{Identifier: "username", Type: "string", Required: true},
		},
		NodeProperties: map[string]interface{}{
			"assignGroup": "test-group-id",
			"assignRole":  "test-role-id",
		},
	}

	suite.mockUserProvider.On("IdentifyUser", attrs).Return(nil,
		userprovider.NewUserProviderError(userprovider.ErrorCodeUserNotFound, "", ""))

	createdUser := &userprovider.User{
		UserID:             "user-provisioned",
		OrganizationUnitID: testOUID,
		UserType:           testUserType,
		Attributes:         attrsJSON,
	}

	suite.mockUserProvider.On("CreateUser", mock.Anything).Return(createdUser, nil)

	// Mock successful group and role assignment
	suite.mockGroupService.On("AddGroupMembers", mock.Anything, "test-group-id", mock.Anything).
		Return(nil, nil)
	suite.mockRoleService.On("AddAssignments", mock.Anything, "test-role-id", mock.Anything).Return(nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	assert.Equal(suite.T(), dataValueTrue, resp.RuntimeData[common.RuntimeKeyUserAutoProvisioned])

	// Verify assignments were made
	suite.mockGroupService.AssertExpectations(suite.T())
	suite.mockRoleService.AssertExpectations(suite.T())
}

// Test successful provisioning with both group and role assignment (detailed verification)
func (suite *ProvisioningExecutorTestSuite) TestExecute_Success_WithGroupAndRoleAssignment() {
	attrs := map[string]interface{}{"username": "newuser", "email": "new@example.com"}
	attrsJSON, _ := json.Marshal(attrs)

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeRegistration,
		UserInputs: map[string]string{
			"username": "newuser",
			"email":    "new@example.com",
		},
		RuntimeData: map[string]string{
			ouIDKey:     testOUID,
			userTypeKey: testUserType,
		},
		NodeInputs: []common.Input{
			{Identifier: "username", Type: "string", Required: true},
			{Identifier: "email", Type: "string", Required: true},
		},
		NodeProperties: map[string]interface{}{
			"assignGroup": "test-group-id",
			"assignRole":  "test-role-id",
		},
	}

	suite.mockUserProvider.On("IdentifyUser", map[string]interface{}{
		"username": "newuser",
		"email":    "new@example.com",
	}).Return(nil, userprovider.NewUserProviderError(userprovider.ErrorCodeUserNotFound, "", ""))

	createdUser := &userprovider.User{
		UserID:             testNewUserID,
		OrganizationUnitID: testOUID,
		UserType:           testUserType,
		Attributes:         attrsJSON,
	}

	suite.mockUserProvider.On("CreateUser", mock.MatchedBy(func(u *userprovider.User) bool {
		return u.OrganizationUnitID == testOUID && u.UserType == testUserType
	})).Return(createdUser, nil)

	// Mock group assignment
	suite.mockGroupService.On("AddGroupMembers", mock.Anything, "test-group-id",
		mock.MatchedBy(func(members []group.Member) bool {
			return len(members) == 1 &&
				members[0].ID == testNewUserID &&
				members[0].Type == group.MemberTypeUser
		})).Return(nil, nil)

	// Mock role assignment
	suite.mockRoleService.On("AddAssignments", mock.Anything, "test-role-id",
		mock.MatchedBy(func(assignments []role.RoleAssignment) bool {
			return len(assignments) == 1 &&
				assignments[0].ID == testNewUserID &&
				assignments[0].Type == role.AssigneeTypeUser
		})).Return(nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	assert.Equal(suite.T(), testNewUserID, resp.AuthenticatedUser.UserID)

	// Verify all mocks were called
	suite.mockUserProvider.AssertExpectations(suite.T())
	suite.mockGroupService.AssertExpectations(suite.T())
	suite.mockRoleService.AssertExpectations(suite.T())
}
