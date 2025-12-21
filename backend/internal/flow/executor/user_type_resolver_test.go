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
	"github.com/stretchr/testify/suite"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/userschema"
	"github.com/asgardeo/thunder/tests/mocks/flow/coremock"
	"github.com/asgardeo/thunder/tests/mocks/userschemamock"
)

type UserTypeResolverTestSuite struct {
	suite.Suite
	mockUserSchemaService *userschemamock.UserSchemaServiceInterfaceMock
	mockFlowFactory       *coremock.FlowFactoryInterfaceMock
	executor              *userTypeResolver
}

func TestUserTypeResolverSuite(t *testing.T) {
	suite.Run(t, new(UserTypeResolverTestSuite))
}

func (suite *UserTypeResolverTestSuite) SetupTest() {
	suite.mockUserSchemaService = userschemamock.NewUserSchemaServiceInterfaceMock(suite.T())
	suite.mockFlowFactory = coremock.NewFlowFactoryInterfaceMock(suite.T())

	// Mock the CreateExecutor method to return a base executor
	suite.mockFlowFactory.On("CreateExecutor", ExecutorNameUserTypeResolver, common.ExecutorTypeRegistration,
		[]common.Input{}, []common.Input{}).
		Return(createMockUserTypeResolverExecutor(suite.T()))

	suite.executor = newUserTypeResolver(suite.mockFlowFactory, suite.mockUserSchemaService)
}

func createMockUserTypeResolverExecutor(t *testing.T) core.ExecutorInterface {
	mockExec := coremock.NewExecutorInterfaceMock(t)
	mockExec.On("GetName").Return(ExecutorNameUserTypeResolver).Maybe()
	mockExec.On("GetType").Return(common.ExecutorTypeRegistration).Maybe()
	mockExec.On("GetDefaultInputs").Return([]common.Input{}).Maybe()
	mockExec.On("GetPrerequisites").Return([]common.Input{}).Maybe()
	return mockExec
}

func (suite *UserTypeResolverTestSuite) TestNewUserTypeResolver() {
	mockFlowFactory := coremock.NewFlowFactoryInterfaceMock(suite.T())
	mockUserSchemaService := userschemamock.NewUserSchemaServiceInterfaceMock(suite.T())

	mockFlowFactory.On("CreateExecutor", ExecutorNameUserTypeResolver, common.ExecutorTypeRegistration,
		[]common.Input{}, []common.Input{}).
		Return(createMockUserTypeResolverExecutor(suite.T()))

	executor := newUserTypeResolver(mockFlowFactory, mockUserSchemaService)

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
	suite.mockUserSchemaService.AssertNotCalled(suite.T(), "GetUserSchemaByName")
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
	suite.mockUserSchemaService.AssertNotCalled(suite.T(), "GetUserSchemaByName")
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
			suite.mockUserSchemaService.AssertNotCalled(suite.T(), "GetUserSchemaByName")
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

			userSchema := &userschema.UserSchema{
				ID:                    "schema-123",
				Name:                  tc.providedUserType,
				OrganizationUnitID:    tc.expectedOUID,
				AllowSelfRegistration: true,
			}
			suite.mockUserSchemaService.On("GetUserSchemaByName", tc.providedUserType).
				Return(userSchema, nil)

			result, err := suite.executor.Execute(ctx)

			assert.NoError(suite.T(), err)
			assert.NotNil(suite.T(), result)
			assert.Equal(suite.T(), common.ExecComplete, result.Status)
			assert.Equal(suite.T(), tc.providedUserType, result.RuntimeData[userTypeKey])
			assert.Equal(suite.T(), tc.expectedOUID, result.RuntimeData[defaultOUIDKey])

			suite.mockUserSchemaService.AssertExpectations(suite.T())
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

	userSchema := &userschema.UserSchema{
		ID:                    "schema-123",
		Name:                  "employee",
		OrganizationUnitID:    "",
		AllowSelfRegistration: true,
	}
	suite.mockUserSchemaService.On("GetUserSchemaByName", "employee").
		Return(userSchema, nil)

	result, err := suite.executor.Execute(ctx)

	assert.Error(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "no organization unit found for user type")
	suite.mockUserSchemaService.AssertExpectations(suite.T())
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
	suite.mockUserSchemaService.AssertNotCalled(suite.T(), "GetUserSchemaByName")
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
	suite.mockUserSchemaService.On("GetUserSchemaByName", "employee").
		Return(nil, svcErr)

	result, err := suite.executor.Execute(ctx)

	assert.Error(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "failed to resolve user schema")
	suite.mockUserSchemaService.AssertExpectations(suite.T())
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
	suite.mockUserSchemaService.AssertNotCalled(suite.T(), "GetUserSchemaByName")
}

func (suite *UserTypeResolverTestSuite) TestExecute_NoAllowedUserTypes_WithUserTypeInput() {
	suite.SetupTest()

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeRegistration,
		Application: appmodel.Application{
			AllowedUserTypes: []string{},
		},
		UserInputs: map[string]string{
			userTypeKey: "employee",
		},
		RuntimeData: map[string]string{},
	}

	userSchema := &userschema.UserSchema{
		ID:                    "schema-123",
		Name:                  "employee",
		OrganizationUnitID:    "ou-123",
		AllowSelfRegistration: true,
	}
	suite.mockUserSchemaService.On("GetUserSchemaByName", "employee").
		Return(userSchema, nil)

	result, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), common.ExecComplete, result.Status)
	assert.Equal(suite.T(), "employee", result.RuntimeData[userTypeKey])
	assert.Equal(suite.T(), "ou-123", result.RuntimeData[defaultOUIDKey])
	suite.mockUserSchemaService.AssertExpectations(suite.T())
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

	userSchema := &userschema.UserSchema{
		ID:                    "schema-123",
		Name:                  "employee",
		OrganizationUnitID:    "ou-123",
		AllowSelfRegistration: true,
	}
	suite.mockUserSchemaService.On("GetUserSchemaByName", "employee").
		Return(userSchema, nil)

	result, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), common.ExecComplete, result.Status)
	assert.Equal(suite.T(), "employee", result.RuntimeData[userTypeKey])
	assert.Equal(suite.T(), "ou-123", result.RuntimeData[defaultOUIDKey])

	suite.mockUserSchemaService.AssertExpectations(suite.T())
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

	userSchema := &userschema.UserSchema{
		ID:                    "schema-123",
		Name:                  "employee",
		OrganizationUnitID:    "",
		AllowSelfRegistration: true,
	}
	suite.mockUserSchemaService.On("GetUserSchemaByName", "employee").
		Return(userSchema, nil)

	result, err := suite.executor.Execute(ctx)

	assert.Error(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "no organization unit found for user type")
	suite.mockUserSchemaService.AssertExpectations(suite.T())
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
	suite.mockUserSchemaService.On("GetUserSchemaByName", "employee").
		Return(nil, svcErr)

	result, err := suite.executor.Execute(ctx)

	assert.Error(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "failed to resolve user schema")
	suite.mockUserSchemaService.AssertExpectations(suite.T())
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
	for _, userType := range []string{"employee", "customer", "partner"} {
		userSchema := &userschema.UserSchema{
			ID:                    "schema-" + userType,
			Name:                  userType,
			OrganizationUnitID:    "ou-" + userType,
			AllowSelfRegistration: true,
		}
		suite.mockUserSchemaService.On("GetUserSchemaByName", userType).
			Return(userSchema, nil)
	}

	result, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), common.ExecUserInputRequired, result.Status)
	assert.NotEmpty(suite.T(), result.Inputs)
	assert.Len(suite.T(), result.Inputs, 1)

	requiredInput := result.Inputs[0]
	assert.Equal(suite.T(), userTypeKey, requiredInput.Identifier)
	assert.Equal(suite.T(), "dropdown", requiredInput.Type)
	assert.True(suite.T(), requiredInput.Required)
	assert.ElementsMatch(suite.T(), []string{"employee", "customer", "partner"}, requiredInput.Options)

	suite.mockUserSchemaService.AssertExpectations(suite.T())
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
	for _, userType := range []string{"employee", "customer"} {
		userSchema := &userschema.UserSchema{
			ID:                    "schema-" + userType,
			Name:                  userType,
			OrganizationUnitID:    "ou-" + userType,
			AllowSelfRegistration: true,
		}
		suite.mockUserSchemaService.On("GetUserSchemaByName", userType).
			Return(userSchema, nil)
	}

	result, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), common.ExecUserInputRequired, result.Status)
	assert.NotEmpty(suite.T(), result.Inputs)
	assert.Len(suite.T(), result.Inputs, 1)

	requiredInput := result.Inputs[0]
	assert.Equal(suite.T(), userTypeKey, requiredInput.Identifier)
	assert.Equal(suite.T(), "dropdown", requiredInput.Type)

	suite.mockUserSchemaService.AssertExpectations(suite.T())
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

	userSchema := &userschema.UserSchema{
		ID:                    "schema-123",
		Name:                  "employee",
		OrganizationUnitID:    "ou-123",
		AllowSelfRegistration: false,
	}
	suite.mockUserSchemaService.On("GetUserSchemaByName", "employee").
		Return(userSchema, nil)

	result, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), common.ExecFailure, result.Status)
	assert.Equal(suite.T(), "Self-registration not enabled for the user type", result.FailureReason)
	suite.mockUserSchemaService.AssertExpectations(suite.T())
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

	userSchema := &userschema.UserSchema{
		ID:                    "schema-123",
		Name:                  "employee",
		OrganizationUnitID:    "ou-123",
		AllowSelfRegistration: false,
	}
	suite.mockUserSchemaService.On("GetUserSchemaByName", "employee").
		Return(userSchema, nil)

	result, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), common.ExecFailure, result.Status)
	assert.Equal(suite.T(), "Self-registration not enabled for the user type", result.FailureReason)
	suite.mockUserSchemaService.AssertExpectations(suite.T())
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
	employeeSchema := &userschema.UserSchema{
		ID:                    "schema-employee",
		Name:                  "employee",
		OrganizationUnitID:    "ou-employee",
		AllowSelfRegistration: false,
	}
	customerSchema := &userschema.UserSchema{
		ID:                    "schema-customer",
		Name:                  "customer",
		OrganizationUnitID:    "ou-customer",
		AllowSelfRegistration: true,
	}
	partnerSchema := &userschema.UserSchema{
		ID:                    "schema-partner",
		Name:                  "partner",
		OrganizationUnitID:    "ou-partner",
		AllowSelfRegistration: false,
	}

	suite.mockUserSchemaService.On("GetUserSchemaByName", "employee").
		Return(employeeSchema, nil)
	suite.mockUserSchemaService.On("GetUserSchemaByName", "customer").
		Return(customerSchema, nil)
	suite.mockUserSchemaService.On("GetUserSchemaByName", "partner").
		Return(partnerSchema, nil)

	result, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), common.ExecComplete, result.Status)
	assert.Equal(suite.T(), "customer", result.RuntimeData[userTypeKey])
	assert.Equal(suite.T(), "ou-customer", result.RuntimeData[defaultOUIDKey])
	suite.mockUserSchemaService.AssertExpectations(suite.T())
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
	employeeSchema := &userschema.UserSchema{
		ID:                    "schema-employee",
		Name:                  "employee",
		OrganizationUnitID:    "ou-employee",
		AllowSelfRegistration: false,
	}
	customerSchema := &userschema.UserSchema{
		ID:                    "schema-customer",
		Name:                  "customer",
		OrganizationUnitID:    "ou-customer",
		AllowSelfRegistration: false,
	}

	suite.mockUserSchemaService.On("GetUserSchemaByName", "employee").
		Return(employeeSchema, nil)
	suite.mockUserSchemaService.On("GetUserSchemaByName", "customer").
		Return(customerSchema, nil)

	result, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), common.ExecFailure, result.Status)
	assert.Equal(suite.T(), "Self-registration not available for this application", result.FailureReason)
	suite.mockUserSchemaService.AssertExpectations(suite.T())
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
	employeeSchema := &userschema.UserSchema{
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

	suite.mockUserSchemaService.On("GetUserSchemaByName", "employee").
		Return(employeeSchema, nil)
	suite.mockUserSchemaService.On("GetUserSchemaByName", "customer").
		Return(nil, svcErr)

	result, err := suite.executor.Execute(ctx)

	assert.Error(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "failed to resolve user schema for user type")
	suite.mockUserSchemaService.AssertExpectations(suite.T())
}

func (suite *UserTypeResolverTestSuite) TestGetUserSchemaAndOU_Success() {
	suite.SetupTest()

	userSchema := &userschema.UserSchema{
		ID:                    "schema-123",
		Name:                  "employee",
		OrganizationUnitID:    "ou-123",
		AllowSelfRegistration: true,
	}
	suite.mockUserSchemaService.On("GetUserSchemaByName", "employee").
		Return(userSchema, nil)

	schema, ouID, err := suite.executor.getUserSchemaAndOU("employee")

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), schema)
	assert.Equal(suite.T(), "ou-123", ouID)
	assert.Equal(suite.T(), "employee", schema.Name)
	suite.mockUserSchemaService.AssertExpectations(suite.T())
}

func (suite *UserTypeResolverTestSuite) TestGetUserSchemaAndOU_NoOUFound() {
	suite.SetupTest()

	userSchema := &userschema.UserSchema{
		ID:                    "schema-123",
		Name:                  "employee",
		OrganizationUnitID:    "",
		AllowSelfRegistration: true,
	}
	suite.mockUserSchemaService.On("GetUserSchemaByName", "employee").
		Return(userSchema, nil)

	schema, ouID, err := suite.executor.getUserSchemaAndOU("employee")

	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), schema)
	assert.Equal(suite.T(), "", ouID)
	assert.Contains(suite.T(), err.Error(), "no organization unit found for user type")
	suite.mockUserSchemaService.AssertExpectations(suite.T())
}

func (suite *UserTypeResolverTestSuite) TestGetUserSchemaAndOU_SchemaNotFound() {
	suite.SetupTest()

	svcErr := &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "SCHEMA-404",
		Error:            "Not Found",
		ErrorDescription: "User schema not found",
	}
	suite.mockUserSchemaService.On("GetUserSchemaByName", "employee").
		Return(nil, svcErr)

	schema, ouID, err := suite.executor.getUserSchemaAndOU("employee")

	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), schema)
	assert.Equal(suite.T(), "", ouID)
	assert.Contains(suite.T(), err.Error(), "failed to resolve user schema for user type")
	suite.mockUserSchemaService.AssertExpectations(suite.T())
}
