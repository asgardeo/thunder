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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/usertype"
	"github.com/asgardeo/thunder/tests/mocks/flow/coremock"
	"github.com/asgardeo/thunder/tests/mocks/usertypemock"
)

type UserTypeResolverTestSuite struct {
	suite.Suite
	mockUserTypeService *usertypemock.UserTypeServiceInterfaceMock
	mockFlowFactory     *coremock.FlowFactoryInterfaceMock
	executor            *userTypeResolver
}

func TestUserTypeResolverSuite(t *testing.T) {
	suite.Run(t, new(UserTypeResolverTestSuite))
}

func (suite *UserTypeResolverTestSuite) SetupTest() {
	suite.mockUserTypeService = usertypemock.NewUserTypeServiceInterfaceMock(suite.T())
	suite.mockFlowFactory = coremock.NewFlowFactoryInterfaceMock(suite.T())

	defaultInputs := []common.Input{
		{
			Ref:        "usertype_input",
			Identifier: userTypeKey,
			Type:       "SELECT",
			Required:   true,
		},
	}

	// Mock the CreateExecutor method to return a base executor
	suite.mockFlowFactory.On("CreateExecutor", ExecutorNameUserTypeResolver, common.ExecutorTypeRegistration,
		defaultInputs, []common.Input{}).
		Return(createMockUserTypeResolverExecutor(suite.T()))

	suite.executor = newUserTypeResolver(suite.mockFlowFactory, suite.mockUserTypeService)
}

func createMockUserTypeResolverExecutor(t *testing.T) core.ExecutorInterface {
	defaultInputs := []common.Input{
		{
			Ref:        "usertype_input",
			Identifier: userTypeKey,
			Type:       "SELECT",
			Required:   true,
		},
	}
	mockExec := coremock.NewExecutorInterfaceMock(t)
	mockExec.On("GetName").Return(ExecutorNameUserTypeResolver).Maybe()
	mockExec.On("GetType").Return(common.ExecutorTypeRegistration).Maybe()
	mockExec.On("GetDefaultInputs").Return(defaultInputs).Maybe()
	mockExec.On("GetDefaultMeta").Return(nil).Maybe()
	mockExec.On("GetPrerequisites").Return([]common.Input{}).Maybe()

	// HasRequiredInputs returns true if userType input is present
	mockExec.On("HasRequiredInputs", mock.Anything, mock.Anything).Return(
		func(ctx *core.NodeContext, execResp *common.ExecutorResponse) bool {
			if val, ok := ctx.UserInputs[userTypeKey]; ok && val != "" {
				return true
			}
			return false
		},
	).Maybe()

	return mockExec
}

func (suite *UserTypeResolverTestSuite) TestNewUserTypeResolver() {
	mockFlowFactory := coremock.NewFlowFactoryInterfaceMock(suite.T())
	mockUserTypeService := usertypemock.NewUserTypeServiceInterfaceMock(suite.T())

	defaultInputs := []common.Input{
		{
			Ref:        "usertype_input",
			Identifier: userTypeKey,
			Type:       "SELECT",
			Required:   true,
		},
	}

	mockFlowFactory.On("CreateExecutor", ExecutorNameUserTypeResolver, common.ExecutorTypeRegistration,
		defaultInputs, []common.Input{}).
		Return(createMockUserTypeResolverExecutor(suite.T()))

	executor := newUserTypeResolver(mockFlowFactory, mockUserTypeService)

	assert.NotNil(suite.T(), executor)
	assert.Equal(suite.T(), ExecutorNameUserTypeResolver, executor.GetName())
}

func (suite *UserTypeResolverTestSuite) TestExecute_AuthenticationFlow_WithAllowedUserTypes() {
	suite.SetupTest()

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		Application: appmodel.Application{
			AllowedUserTypes: []string{"employee", "customer"},
		},
		RuntimeData: map[string]string{},
	}

	result, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), common.ExecComplete, result.Status)
	assert.Empty(suite.T(), result.RuntimeData[userTypeKey])
	suite.mockUserTypeService.AssertNotCalled(suite.T(), "GetUserTypeByName")
}

func (suite *UserTypeResolverTestSuite) TestExecute_AuthenticationFlow_NoAllowedUserTypes() {
	suite.SetupTest()

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		Application: appmodel.Application{
			AllowedUserTypes: []string{},
		},
		RuntimeData: map[string]string{},
	}

	result, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), common.ExecFailure, result.Status)
	assert.Equal(suite.T(), "Authentication not available for this application", result.FailureReason)
	suite.mockUserTypeService.AssertNotCalled(suite.T(), "GetUserTypeByName")
}

func (suite *UserTypeResolverTestSuite) TestExecute_UnsupportedFlowType() {
	testCases := []struct {
		name     string
		flowType common.FlowType
	}{
		{
			name:     "UnknownFlowType",
			flowType: common.FlowType("UNKNOWN"),
		},
		{
			name:     "EmptyFlowType",
			flowType: common.FlowType(""),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			ctx := &core.NodeContext{
				FlowID:   "flow-123",
				FlowType: tc.flowType,
				Application: appmodel.Application{
					AllowedUserTypes: []string{"employee"},
				},
				RuntimeData: map[string]string{},
			}

			result, err := suite.executor.Execute(ctx)

			assert.NoError(suite.T(), err)
			assert.NotNil(suite.T(), result)
			assert.Equal(suite.T(), common.ExecComplete, result.Status)
			assert.Empty(suite.T(), result.RuntimeData[userTypeKey])
			suite.mockUserTypeService.AssertNotCalled(suite.T(), "GetUserTypeByName")
		})
	}
}

func (suite *UserTypeResolverTestSuite) TestExecute_UserTypeProvidedInInput_Success() {
	testCases := []struct {
		name             string
		allowedUserTypes []string
		providedUserType string
		expectedOUID     string
	}{
		{
			name:             "Valid user type with OU",
			allowedUserTypes: []string{"employee", "customer"},
			providedUserType: "employee",
			expectedOUID:     "ou-123",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			ctx := &core.NodeContext{
				FlowID:   "flow-123",
				FlowType: common.FlowTypeRegistration,
				Application: appmodel.Application{
					AllowedUserTypes: tc.allowedUserTypes,
				},
				UserInputs: map[string]string{
					userTypeKey: tc.providedUserType,
				},
				RuntimeData: map[string]string{},
			}

			userType := &usertype.UserType{
				ID:                    "schema-123",
				Name:                  tc.providedUserType,
				OrganizationUnitID:    tc.expectedOUID,
				AllowSelfRegistration: true,
			}
			suite.mockUserTypeService.On("GetUserTypeByName", tc.providedUserType).
				Return(userType, nil)

			result, err := suite.executor.Execute(ctx)

			assert.NoError(suite.T(), err)
			assert.NotNil(suite.T(), result)
			assert.Equal(suite.T(), common.ExecComplete, result.Status)
			assert.Equal(suite.T(), tc.providedUserType, result.RuntimeData[userTypeKey])
			assert.Equal(suite.T(), tc.expectedOUID, result.RuntimeData[defaultOUIDKey])

			suite.mockUserTypeService.AssertExpectations(suite.T())
		})
	}
}

func (suite *UserTypeResolverTestSuite) TestExecute_UserTypeProvidedInInput_NoOU() {
	suite.SetupTest()

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeRegistration,
		Application: appmodel.Application{
			AllowedUserTypes: []string{"employee", "customer"},
		},
		UserInputs: map[string]string{
			userTypeKey: "employee",
		},
		RuntimeData: map[string]string{},
	}

	userType := &usertype.UserType{
		ID:                    "schema-123",
		Name:                  "employee",
		OrganizationUnitID:    "",
		AllowSelfRegistration: true,
	}
	suite.mockUserTypeService.On("GetUserTypeByName", "employee").
		Return(userType, nil)

	result, err := suite.executor.Execute(ctx)

	assert.Error(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "no organization unit found for user type")
	suite.mockUserTypeService.AssertExpectations(suite.T())
}

func (suite *UserTypeResolverTestSuite) TestExecute_UserTypeProvidedInInput_NotAllowed() {
	suite.SetupTest()

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeRegistration,
		Application: appmodel.Application{
			AllowedUserTypes: []string{"employee", "customer"},
		},
		UserInputs: map[string]string{
			userTypeKey: "partner",
		},
		RuntimeData: map[string]string{},
	}

	result, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), common.ExecFailure, result.Status)
	assert.Equal(suite.T(), "Application does not allow registration for the user type", result.FailureReason)
	suite.mockUserTypeService.AssertNotCalled(suite.T(), "GetUserTypeByName")
}

func (suite *UserTypeResolverTestSuite) TestExecute_UserTypeProvidedInInput_OUResolutionFails() {
	suite.SetupTest()

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeRegistration,
		Application: appmodel.Application{
			AllowedUserTypes: []string{"employee"},
		},
		UserInputs: map[string]string{
			userTypeKey: "employee",
		},
		RuntimeData: map[string]string{},
	}

	svcErr := &serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "SCHEMA-500",
		Error:            "Internal Server Error",
		ErrorDescription: "Failed to retrieve OU",
	}
	suite.mockUserTypeService.On("GetUserTypeByName", "employee").
		Return(nil, svcErr)

	result, err := suite.executor.Execute(ctx)

	assert.Error(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "failed to resolve user type")
	suite.mockUserTypeService.AssertExpectations(suite.T())
}

func (suite *UserTypeResolverTestSuite) TestExecute_NoAllowedUserTypes() {
	suite.SetupTest()

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeRegistration,
		Application: appmodel.Application{
			AllowedUserTypes: []string{},
		},
		UserInputs:  map[string]string{},
		RuntimeData: map[string]string{},
	}

	result, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), common.ExecFailure, result.Status)
	assert.Equal(suite.T(), "Self-registration not available for this application", result.FailureReason)
	suite.mockUserTypeService.AssertNotCalled(suite.T(), "GetUserTypeByName")
}

func (suite *UserTypeResolverTestSuite) TestExecute_SingleAllowedUserType_Success() {
	suite.SetupTest()

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeRegistration,
		Application: appmodel.Application{
			AllowedUserTypes: []string{"employee"},
		},
		UserInputs:  map[string]string{},
		RuntimeData: map[string]string{},
	}

	userType := &usertype.UserType{
		ID:                    "schema-123",
		Name:                  "employee",
		OrganizationUnitID:    "ou-123",
		AllowSelfRegistration: true,
	}
	suite.mockUserTypeService.On("GetUserTypeByName", "employee").
		Return(userType, nil)

	result, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), common.ExecComplete, result.Status)
	assert.Equal(suite.T(), "employee", result.RuntimeData[userTypeKey])
	assert.Equal(suite.T(), "ou-123", result.RuntimeData[defaultOUIDKey])

	suite.mockUserTypeService.AssertExpectations(suite.T())
}

func (suite *UserTypeResolverTestSuite) TestExecute_SingleAllowedUserType_NoOU() {
	suite.SetupTest()

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeRegistration,
		Application: appmodel.Application{
			AllowedUserTypes: []string{"employee"},
		},
		UserInputs:  map[string]string{},
		RuntimeData: map[string]string{},
	}

	userType := &usertype.UserType{
		ID:                    "schema-123",
		Name:                  "employee",
		OrganizationUnitID:    "",
		AllowSelfRegistration: true,
	}
	suite.mockUserTypeService.On("GetUserTypeByName", "employee").
		Return(userType, nil)

	result, err := suite.executor.Execute(ctx)

	assert.Error(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "no organization unit found for user type")
	suite.mockUserTypeService.AssertExpectations(suite.T())
}

func (suite *UserTypeResolverTestSuite) TestExecute_SingleAllowedUserType_OUResolutionFails() {
	suite.SetupTest()

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeRegistration,
		Application: appmodel.Application{
			AllowedUserTypes: []string{"employee"},
		},
		UserInputs:  map[string]string{},
		RuntimeData: map[string]string{},
	}

	svcErr := &serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "SCHEMA-500",
		Error:            "Internal Server Error",
		ErrorDescription: "Failed to retrieve OU",
	}
	suite.mockUserTypeService.On("GetUserTypeByName", "employee").
		Return(nil, svcErr)

	result, err := suite.executor.Execute(ctx)

	assert.Error(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "failed to resolve user type")
	suite.mockUserTypeService.AssertExpectations(suite.T())
}

func (suite *UserTypeResolverTestSuite) TestExecute_MultipleAllowedUserTypes_PromptUser() {
	suite.SetupTest()

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeRegistration,
		Application: appmodel.Application{
			AllowedUserTypes: []string{"employee", "customer", "partner"},
		},
		UserInputs:  map[string]string{},
		RuntimeData: map[string]string{},
	}

	// Mock all three user types with self registration enabled
	for _, userTypeName := range []string{"employee", "customer", "partner"} {
		userType := &usertype.UserType{
			ID:                    "schema-" + userTypeName,
			Name:                  userTypeName,
			OrganizationUnitID:    "ou-" + userTypeName,
			AllowSelfRegistration: true,
		}
		suite.mockUserTypeService.On("GetUserTypeByName", userTypeName).
			Return(userType, nil)
	}

	result, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), common.ExecUserInputRequired, result.Status)
	assert.NotEmpty(suite.T(), result.Inputs)
	assert.Len(suite.T(), result.Inputs, 1)

	requiredInput := result.Inputs[0]
	assert.Equal(suite.T(), userTypeKey, requiredInput.Identifier)
	assert.Equal(suite.T(), "SELECT", requiredInput.Type)
	assert.Equal(suite.T(), "usertype_input", requiredInput.Ref)
	assert.True(suite.T(), requiredInput.Required)
	assert.ElementsMatch(suite.T(), []string{"employee", "customer", "partner"}, requiredInput.Options)

	suite.mockUserTypeService.AssertExpectations(suite.T())
}

func (suite *UserTypeResolverTestSuite) TestExecute_EmptyUserTypeInput() {
	suite.SetupTest()

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeRegistration,
		Application: appmodel.Application{
			AllowedUserTypes: []string{"employee", "customer"},
		},
		UserInputs: map[string]string{
			userTypeKey: "",
		},
		RuntimeData: map[string]string{},
	}

	// Mock both user types with self registration enabled
	for _, userTypeName := range []string{"employee", "customer"} {
		userType := &usertype.UserType{
			ID:                    "schema-" + userTypeName,
			Name:                  userTypeName,
			OrganizationUnitID:    "ou-" + userTypeName,
			AllowSelfRegistration: true,
		}
		suite.mockUserTypeService.On("GetUserTypeByName", userTypeName).
			Return(userType, nil)
	}

	result, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), common.ExecUserInputRequired, result.Status)
	assert.NotEmpty(suite.T(), result.Inputs)
	assert.Len(suite.T(), result.Inputs, 1)

	requiredInput := result.Inputs[0]
	assert.Equal(suite.T(), userTypeKey, requiredInput.Identifier)
	assert.Equal(suite.T(), "SELECT", requiredInput.Type)

	suite.mockUserTypeService.AssertExpectations(suite.T())
}

func (suite *UserTypeResolverTestSuite) TestExecute_UserTypeProvidedInInput_SelfRegistrationDisabled() {
	suite.SetupTest()

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeRegistration,
		Application: appmodel.Application{
			AllowedUserTypes: []string{"employee"},
		},
		UserInputs: map[string]string{
			userTypeKey: "employee",
		},
		RuntimeData: map[string]string{},
	}

	userType := &usertype.UserType{
		ID:                    "schema-123",
		Name:                  "employee",
		OrganizationUnitID:    "ou-123",
		AllowSelfRegistration: false,
	}
	suite.mockUserTypeService.On("GetUserTypeByName", "employee").
		Return(userType, nil)

	result, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), common.ExecFailure, result.Status)
	assert.Equal(suite.T(), "Self-registration not enabled for the user type", result.FailureReason)
	suite.mockUserTypeService.AssertExpectations(suite.T())
}

func (suite *UserTypeResolverTestSuite) TestExecute_SingleAllowedUserType_SelfRegistrationDisabled() {
	suite.SetupTest()

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeRegistration,
		Application: appmodel.Application{
			AllowedUserTypes: []string{"employee"},
		},
		UserInputs:  map[string]string{},
		RuntimeData: map[string]string{},
	}

	userType := &usertype.UserType{
		ID:                    "schema-123",
		Name:                  "employee",
		OrganizationUnitID:    "ou-123",
		AllowSelfRegistration: false,
	}
	suite.mockUserTypeService.On("GetUserTypeByName", "employee").
		Return(userType, nil)

	result, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), common.ExecFailure, result.Status)
	assert.Equal(suite.T(), "Self-registration not enabled for the user type", result.FailureReason)
	suite.mockUserTypeService.AssertExpectations(suite.T())
}

func (suite *UserTypeResolverTestSuite) TestExecute_MultipleAllowedUserTypes_OnlyOneSelfRegEnabled() {
	suite.SetupTest()

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeRegistration,
		Application: appmodel.Application{
			AllowedUserTypes: []string{"employee", "customer", "partner"},
		},
		UserInputs:  map[string]string{},
		RuntimeData: map[string]string{},
	}

	// Only customer has self-registration enabled
	employeeSchema := &usertype.UserType{
		ID:                    "schema-employee",
		Name:                  "employee",
		OrganizationUnitID:    "ou-employee",
		AllowSelfRegistration: false,
	}
	customerSchema := &usertype.UserType{
		ID:                    "schema-customer",
		Name:                  "customer",
		OrganizationUnitID:    "ou-customer",
		AllowSelfRegistration: true,
	}
	partnerSchema := &usertype.UserType{
		ID:                    "schema-partner",
		Name:                  "partner",
		OrganizationUnitID:    "ou-partner",
		AllowSelfRegistration: false,
	}

	suite.mockUserTypeService.On("GetUserTypeByName", "employee").
		Return(employeeSchema, nil)
	suite.mockUserTypeService.On("GetUserTypeByName", "customer").
		Return(customerSchema, nil)
	suite.mockUserTypeService.On("GetUserTypeByName", "partner").
		Return(partnerSchema, nil)

	result, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), common.ExecComplete, result.Status)
	assert.Equal(suite.T(), "customer", result.RuntimeData[userTypeKey])
	assert.Equal(suite.T(), "ou-customer", result.RuntimeData[defaultOUIDKey])
	suite.mockUserTypeService.AssertExpectations(suite.T())
}

func (suite *UserTypeResolverTestSuite) TestExecute_MultipleAllowedUserTypes_NoSelfRegEnabled() {
	suite.SetupTest()

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeRegistration,
		Application: appmodel.Application{
			AllowedUserTypes: []string{"employee", "customer"},
		},
		UserInputs:  map[string]string{},
		RuntimeData: map[string]string{},
	}

	// None have self-registration enabled
	employeeSchema := &usertype.UserType{
		ID:                    "schema-employee",
		Name:                  "employee",
		OrganizationUnitID:    "ou-employee",
		AllowSelfRegistration: false,
	}
	customerSchema := &usertype.UserType{
		ID:                    "schema-customer",
		Name:                  "customer",
		OrganizationUnitID:    "ou-customer",
		AllowSelfRegistration: false,
	}

	suite.mockUserTypeService.On("GetUserTypeByName", "employee").
		Return(employeeSchema, nil)
	suite.mockUserTypeService.On("GetUserTypeByName", "customer").
		Return(customerSchema, nil)

	result, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), common.ExecFailure, result.Status)
	assert.Equal(suite.T(), "Self-registration not available for this application", result.FailureReason)
	suite.mockUserTypeService.AssertExpectations(suite.T())
}

func (suite *UserTypeResolverTestSuite) TestExecute_MultipleAllowedUserTypes_SchemaResolutionFails() {
	suite.SetupTest()

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeRegistration,
		Application: appmodel.Application{
			AllowedUserTypes: []string{"employee", "customer"},
		},
		UserInputs:  map[string]string{},
		RuntimeData: map[string]string{},
	}

	// First schema succeeds, second fails
	employeeSchema := &usertype.UserType{
		ID:                    "schema-employee",
		Name:                  "employee",
		OrganizationUnitID:    "ou-employee",
		AllowSelfRegistration: true,
	}
	svcErr := &serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "SCHEMA-500",
		Error:            "Internal Server Error",
		ErrorDescription: "Failed to retrieve schema",
	}

	suite.mockUserTypeService.On("GetUserTypeByName", "employee").
		Return(employeeSchema, nil)
	suite.mockUserTypeService.On("GetUserTypeByName", "customer").
		Return(nil, svcErr)

	result, err := suite.executor.Execute(ctx)

	assert.Error(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "failed to resolve user type for user type")
	suite.mockUserTypeService.AssertExpectations(suite.T())
}

func (suite *UserTypeResolverTestSuite) TestGetUserTypeAndOU_Success() {
	suite.SetupTest()

	userType := &usertype.UserType{
		ID:                    "schema-123",
		Name:                  "employee",
		OrganizationUnitID:    "ou-123",
		AllowSelfRegistration: true,
	}
	suite.mockUserTypeService.On("GetUserTypeByName", "employee").
		Return(userType, nil)

	schema, ouID, err := suite.executor.getUserTypeAndOU("employee")

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), schema)
	assert.Equal(suite.T(), "ou-123", ouID)
	assert.Equal(suite.T(), "employee", schema.Name)
	suite.mockUserTypeService.AssertExpectations(suite.T())
}

func (suite *UserTypeResolverTestSuite) TestGetUserTypeAndOU_NoOUFound() {
	suite.SetupTest()

	userType := &usertype.UserType{
		ID:                    "schema-123",
		Name:                  "employee",
		OrganizationUnitID:    "",
		AllowSelfRegistration: true,
	}
	suite.mockUserTypeService.On("GetUserTypeByName", "employee").
		Return(userType, nil)

	schema, ouID, err := suite.executor.getUserTypeAndOU("employee")

	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), schema)
	assert.Equal(suite.T(), "", ouID)
	assert.Contains(suite.T(), err.Error(), "no organization unit found for user type")
	suite.mockUserTypeService.AssertExpectations(suite.T())
}

func (suite *UserTypeResolverTestSuite) TestGetUserTypeAndOU_SchemaNotFound() {
	suite.SetupTest()

	svcErr := &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "SCHEMA-404",
		Error:            "Not Found",
		ErrorDescription: "User type not found",
	}
	suite.mockUserTypeService.On("GetUserTypeByName", "employee").
		Return(nil, svcErr)

	schema, ouID, err := suite.executor.getUserTypeAndOU("employee")

	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), schema)
	assert.Equal(suite.T(), "", ouID)
	assert.Contains(suite.T(), err.Error(), "failed to resolve user type for user type")
	suite.mockUserTypeService.AssertExpectations(suite.T())
}

func (suite *UserTypeResolverTestSuite) TestGetDefaultMeta() {
	suite.SetupTest()

	meta := suite.executor.GetDefaultMeta()

	// Verify meta is returned and has proper structure
	assert.NotNil(suite.T(), meta)
	metaStruct, ok := meta.(core.MetaStructure)
	assert.True(suite.T(), ok, "Meta should be of type core.MetaStructure")
	assert.NotEmpty(suite.T(), metaStruct.Components, "Meta should contain components")

	// MetaBuilder creates TEXT (heading), BLOCK (containing inputs), and ACTION (submit button) components
	// Verify we have at least a BLOCK component (which contains the inputs)
	foundBlock := false
	for _, component := range metaStruct.Components {
		if component.Type == "BLOCK" {
			foundBlock = true
			assert.NotEmpty(suite.T(), component.Components, "BLOCK should contain nested components (inputs)")
			break
		}
	}
	assert.True(suite.T(), foundBlock, "Meta should contain a BLOCK component with inputs")
}

func (suite *UserTypeResolverTestSuite) TestExecute_UserOnboardingFlow_UserTypeProvided_Success() {
	suite.SetupTest()

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeUserOnboarding, // User Onboarding Flow
		UserInputs: map[string]string{
			userTypeKey: "employee",
		},
		RuntimeData: map[string]string{},
	}

	userType := &usertype.UserType{
		ID:                 "schema-123",
		Name:               "employee",
		OrganizationUnitID: "ou-123",
	}
	suite.mockUserTypeService.On("GetUserTypeByName", "employee").
		Return(userType, nil)

	result, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), common.ExecComplete, result.Status)
	assert.Equal(suite.T(), "employee", result.RuntimeData[userTypeKey])
	assert.Equal(suite.T(), "ou-123", result.RuntimeData[defaultOUIDKey])

	suite.mockUserTypeService.AssertExpectations(suite.T())
}

func (suite *UserTypeResolverTestSuite) TestExecute_UserOnboardingFlow_UserTypeProvided_Invalid() {
	suite.SetupTest()

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeUserOnboarding,
		UserInputs: map[string]string{
			userTypeKey: "invalid_user",
		},
		RuntimeData: map[string]string{},
	}

	// Mock schema retrieval failing
	svcErr := &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "SCHEMA-404",
		Error:            "Not Found",
		ErrorDescription: "User type not found",
	}
	suite.mockUserTypeService.On("GetUserTypeByName", "invalid_user").
		Return(nil, svcErr)

	result, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err) // Logic returns ExecFailure, not error
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), common.ExecFailure, result.Status)
	assert.Equal(suite.T(), "Invalid user type", result.FailureReason)

	suite.mockUserTypeService.AssertExpectations(suite.T())
}

func (suite *UserTypeResolverTestSuite) TestExecute_UserOnboardingFlow_NoUserType_SchemaListEmpty() {
	suite.SetupTest()

	ctx := &core.NodeContext{
		FlowID:      "flow-123",
		FlowType:    common.FlowTypeUserOnboarding,
		UserInputs:  map[string]string{},
		RuntimeData: map[string]string{},
	}

	// Mock GetUserTypeList returning empty list
	emptyList := &usertype.UserTypeListResponse{
		Types: []usertype.UserTypeListItem{},
	}
	suite.mockUserTypeService.On("GetUserTypeList", 100, 0).
		Return(emptyList, nil)

	result, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), common.ExecFailure, result.Status)
	assert.Equal(suite.T(), "No user types available", result.FailureReason)
}

func (suite *UserTypeResolverTestSuite) TestExecute_UserOnboardingFlow_NoUserType_SchemaListError() {
	suite.SetupTest()

	ctx := &core.NodeContext{
		FlowID:      "flow-123",
		FlowType:    common.FlowTypeUserOnboarding,
		UserInputs:  map[string]string{},
		RuntimeData: map[string]string{},
	}

	svcErr := &serviceerror.ServiceError{
		Type:  serviceerror.ServerErrorType,
		Error: "Simulated Error",
	}
	suite.mockUserTypeService.On("GetUserTypeList", 100, 0).
		Return(nil, svcErr)

	result, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), common.ExecFailure, result.Status)
	assert.Equal(suite.T(), "Failed to retrieve user types", result.FailureReason)
}

func (suite *UserTypeResolverTestSuite) TestExecute_UserOnboardingFlow_NoUserType_PromptUser() {
	suite.SetupTest()

	ctx := &core.NodeContext{
		FlowID:      "flow-123",
		FlowType:    common.FlowTypeUserOnboarding,
		UserInputs:  map[string]string{},
		RuntimeData: map[string]string{},
	}

	// Mock GetUserTypeList returning schemas
	schemaList := &usertype.UserTypeListResponse{
		Types: []usertype.UserTypeListItem{
			{Name: "employee"},
			{Name: "customer"},
		},
	}
	suite.mockUserTypeService.On("GetUserTypeList", 100, 0).
		Return(schemaList, nil)

	result, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), common.ExecUserInputRequired, result.Status)
	assert.Len(suite.T(), result.Inputs, 1)

	// Verify options in the prompt
	requiredInput := result.Inputs[0]
	assert.Equal(suite.T(), userTypeKey, requiredInput.Identifier)
	assert.ElementsMatch(suite.T(), []string{"employee", "customer"}, requiredInput.Options)
}
