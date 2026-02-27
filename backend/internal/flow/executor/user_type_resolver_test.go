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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

	suite.executor = newUserTypeResolver(suite.mockFlowFactory, suite.mockUserSchemaService)
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
	mockUserSchemaService := userschemamock.NewUserSchemaServiceInterfaceMock(suite.T())

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
			suite.mockUserSchemaService.On("GetUserSchemaByName", context.TODO(), tc.providedUserType).
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
	suite.mockUserSchemaService.On("GetUserSchemaByName", context.TODO(), "employee").
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
	suite.mockUserSchemaService.On("GetUserSchemaByName", context.TODO(), "employee").
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
	suite.mockUserSchemaService.On("GetUserSchemaByName", context.TODO(), "employee").
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
	suite.mockUserSchemaService.On("GetUserSchemaByName", context.TODO(), "employee").
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
	suite.mockUserSchemaService.On("GetUserSchemaByName", context.TODO(), "employee").
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
		suite.mockUserSchemaService.On("GetUserSchemaByName", context.TODO(), userType).
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
	assert.Equal(suite.T(), "SELECT", requiredInput.Type)
	assert.Equal(suite.T(), "usertype_input", requiredInput.Ref)
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
		suite.mockUserSchemaService.On("GetUserSchemaByName", context.TODO(), userType).
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
	assert.Equal(suite.T(), "SELECT", requiredInput.Type)

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
	suite.mockUserSchemaService.On("GetUserSchemaByName", mock.Anything, "employee").
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
	suite.mockUserSchemaService.On("GetUserSchemaByName", mock.Anything, "employee").
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

	suite.mockUserSchemaService.On("GetUserSchemaByName", context.TODO(), "employee").
		Return(employeeSchema, nil)
	suite.mockUserSchemaService.On("GetUserSchemaByName", context.TODO(), "customer").
		Return(customerSchema, nil)
	suite.mockUserSchemaService.On("GetUserSchemaByName", context.TODO(), "partner").
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

	suite.mockUserSchemaService.On("GetUserSchemaByName", context.TODO(), "employee").
		Return(employeeSchema, nil)
	suite.mockUserSchemaService.On("GetUserSchemaByName", context.TODO(), "customer").
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

	suite.mockUserSchemaService.On("GetUserSchemaByName", context.TODO(), "employee").
		Return(employeeSchema, nil)
	suite.mockUserSchemaService.On("GetUserSchemaByName", context.TODO(), "customer").
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
	suite.mockUserSchemaService.On("GetUserSchemaByName", context.TODO(), "employee").
		Return(userSchema, nil)

	schema, ouID, err := suite.executor.getUserSchemaAndOU(context.TODO(), "employee")

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
	suite.mockUserSchemaService.On("GetUserSchemaByName", context.TODO(), "employee").
		Return(userSchema, nil)

	schema, ouID, err := suite.executor.getUserSchemaAndOU(context.TODO(), "employee")

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
	suite.mockUserSchemaService.On("GetUserSchemaByName", context.TODO(), "employee").
		Return(nil, svcErr)

	schema, ouID, err := suite.executor.getUserSchemaAndOU(context.TODO(), "employee")

	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), schema)
	assert.Equal(suite.T(), "", ouID)
	assert.Contains(suite.T(), err.Error(), "failed to resolve user schema for user type")
	suite.mockUserSchemaService.AssertExpectations(suite.T())
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

	userSchema := &userschema.UserSchema{
		ID:                 "schema-123",
		Name:               "employee",
		OrganizationUnitID: "ou-123",
	}
	suite.mockUserSchemaService.On("GetUserSchemaByName", context.TODO(), "employee").
		Return(userSchema, nil)

	result, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), common.ExecComplete, result.Status)
	assert.Equal(suite.T(), "employee", result.RuntimeData[userTypeKey])
	assert.Equal(suite.T(), "ou-123", result.RuntimeData[defaultOUIDKey])

	suite.mockUserSchemaService.AssertExpectations(suite.T())
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
		ErrorDescription: "User schema not found",
	}
	suite.mockUserSchemaService.On("GetUserSchemaByName", context.TODO(), "invalid_user").
		Return(nil, svcErr)

	result, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err) // Logic returns ExecFailure, not error
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), common.ExecFailure, result.Status)
	assert.Equal(suite.T(), "Invalid user type", result.FailureReason)

	suite.mockUserSchemaService.AssertExpectations(suite.T())
}

func (suite *UserTypeResolverTestSuite) TestExecute_UserOnboardingFlow_NoUserType_SchemaListEmpty() {
	suite.SetupTest()

	ctx := &core.NodeContext{
		FlowID:      "flow-123",
		FlowType:    common.FlowTypeUserOnboarding,
		UserInputs:  map[string]string{},
		RuntimeData: map[string]string{},
	}

	// Mock GetUserSchemaList returning empty list
	emptyList := &userschema.UserSchemaListResponse{
		Schemas: []userschema.UserSchemaListItem{},
	}
	suite.mockUserSchemaService.On("GetUserSchemaList", context.TODO(), 100, 0).
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
	suite.mockUserSchemaService.On("GetUserSchemaList", context.TODO(), 100, 0).
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

	// Mock GetUserSchemaList returning schemas
	schemaList := &userschema.UserSchemaListResponse{
		Schemas: []userschema.UserSchemaListItem{
			{Name: "employee"},
			{Name: "customer"},
		},
	}
	suite.mockUserSchemaService.On("GetUserSchemaList", context.TODO(), 100, 0).
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

func (suite *UserTypeResolverTestSuite) TestPromptUserSelection_ForwardsInputsInForwardedData() {
	suite.SetupTest()

	options := []string{"employee", "customer", "partner"}
	execResp := &common.ExecutorResponse{
		ForwardedData: map[string]interface{}{},
	}

	suite.executor.promptUserSelection(execResp, options)

	// Verify status and inputs are set
	assert.Equal(suite.T(), common.ExecUserInputRequired, execResp.Status)
	assert.Len(suite.T(), execResp.Inputs, 1)
	assert.Equal(suite.T(), userTypeKey, execResp.Inputs[0].Identifier)
	assert.ElementsMatch(suite.T(), options, execResp.Inputs[0].Options)

	// Verify ForwardedData is set with inputs
	assert.NotNil(suite.T(), execResp.ForwardedData)
	forwardedInputs, ok := execResp.ForwardedData[common.ForwardedDataKeyInputs]
	assert.True(suite.T(), ok, "ForwardedData should contain 'inputs' key")

	// Type assert and verify
	inputsSlice, ok := forwardedInputs.([]common.Input)
	assert.True(suite.T(), ok, "ForwardedData['inputs'] should be []common.Input")
	assert.Len(suite.T(), inputsSlice, 1)
	assert.Equal(suite.T(), userTypeKey, inputsSlice[0].Identifier)
	assert.ElementsMatch(suite.T(), options, inputsSlice[0].Options)
}

func (suite *UserTypeResolverTestSuite) TestPromptUserSelection_WithEmptyOptions() {
	suite.SetupTest()

	options := []string{}
	execResp := &common.ExecutorResponse{
		ForwardedData: map[string]interface{}{},
	}

	suite.executor.promptUserSelection(execResp, options)

	// Verify status is set
	assert.Equal(suite.T(), common.ExecUserInputRequired, execResp.Status)
	assert.Len(suite.T(), execResp.Inputs, 1)

	// Verify ForwardedData is still set even with empty options
	assert.NotNil(suite.T(), execResp.ForwardedData)
	forwardedInputs, ok := execResp.ForwardedData[common.ForwardedDataKeyInputs]
	assert.True(suite.T(), ok, "ForwardedData should contain 'inputs' key even with empty options")

	inputsSlice, ok := forwardedInputs.([]common.Input)
	assert.True(suite.T(), ok)
	assert.Len(suite.T(), inputsSlice, 1)
	assert.Empty(suite.T(), inputsSlice[0].Options)
}

func (suite *UserTypeResolverTestSuite) TestPromptUserSelection_PreservesExistingForwardedData() {
	suite.SetupTest()

	options := []string{"employee"}
	execResp := &common.ExecutorResponse{
		ForwardedData: map[string]interface{}{
			"existingKey": "existingValue",
		},
	}

	suite.executor.promptUserSelection(execResp, options)

	// Verify existing ForwardedData is preserved
	assert.NotNil(suite.T(), execResp.ForwardedData)
	assert.Equal(suite.T(), "existingValue", execResp.ForwardedData["existingKey"])

	// Verify new inputs key is added
	forwardedInputs, ok := execResp.ForwardedData[common.ForwardedDataKeyInputs]
	assert.True(suite.T(), ok)
	assert.NotNil(suite.T(), forwardedInputs)
}
