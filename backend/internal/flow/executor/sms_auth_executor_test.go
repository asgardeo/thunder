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
	"github.com/asgardeo/thunder/tests/mocks/notification/notificationmock"
	"github.com/asgardeo/thunder/tests/mocks/observabilitymock"
	"github.com/asgardeo/thunder/tests/mocks/userprovidermock"
)

type SMSAuthExecutorTestSuite struct {
	suite.Suite
	mockOTPService    *notificationmock.OTPServiceInterfaceMock
	mockFlowFactory   *coremock.FlowFactoryInterfaceMock
	mockObservability *observabilitymock.ObservabilityServiceInterfaceMock
	mockUserProvider  *userprovidermock.UserProviderInterfaceMock
	executor          *smsOTPAuthExecutor
}

func TestSMSAuthExecutorSuite(t *testing.T) {
	suite.Run(t, new(SMSAuthExecutorTestSuite))
}

func (suite *SMSAuthExecutorTestSuite) SetupTest() {
	suite.mockOTPService = notificationmock.NewOTPServiceInterfaceMock(suite.T())
	suite.mockFlowFactory = coremock.NewFlowFactoryInterfaceMock(suite.T())
	suite.mockObservability = observabilitymock.NewObservabilityServiceInterfaceMock(suite.T())
	suite.mockUserProvider = userprovidermock.NewUserProviderInterfaceMock(suite.T())

	// Default behavior for observability: disabled
	suite.mockObservability.On("IsEnabled").Return(false).Maybe()

	defaultInputs := []common.Input{
		{
			Ref:        "otp_input",
			Identifier: userInputOTP,
			Type:       common.InputTypeOTP,
			Required:   true,
		},
	}
	prerequisites := []common.Input{
		MobileNumberInput,
	}

	// Mock identifying executor
	identifyingMock := createMockIdentifyingExecutor(suite.T())
	suite.mockFlowFactory.On("CreateExecutor", ExecutorNameIdentifying, common.ExecutorTypeUtility,
		mock.Anything, mock.Anything).Return(identifyingMock).Maybe()

	// Mock base executor
	mockExec := coremock.NewExecutorInterfaceMock(suite.T())
	mockExec.On("GetName").Return(ExecutorNameSMSAuth).Maybe()
	mockExec.On("GetType").Return(common.ExecutorTypeAuthentication).Maybe()
	mockExec.On("GetDefaultInputs").Return(defaultInputs).Maybe()
	mockExec.On("GetRequiredInputs", mock.Anything).Return(defaultInputs).Maybe()
	mockExec.On("GetPrerequisites").Return(prerequisites).Maybe()
	mockExec.On("ValidatePrerequisites", mock.Anything, mock.Anything).Return(true).Maybe()
	mockExec.On("HasRequiredInputs", mock.Anything, mock.Anything).Return(
		func(ctx *core.NodeContext, execResp *common.ExecutorResponse) bool {
			otp, exists := ctx.UserInputs[userInputOTP]
			if !exists || otp == "" {
				execResp.Inputs = defaultInputs
				execResp.Status = common.ExecUserInputRequired
				return false
			}
			return true
		}).Maybe()

	suite.mockFlowFactory.On("CreateExecutor", ExecutorNameSMSAuth, common.ExecutorTypeAuthentication,
		defaultInputs, prerequisites).Return(mockExec)

	suite.executor = newSMSOTPAuthExecutor(suite.mockFlowFactory,
		suite.mockOTPService, suite.mockObservability, suite.mockUserProvider)
	// Inject the mock base executor
	suite.executor.ExecutorInterface = mockExec
}

func (suite *SMSAuthExecutorTestSuite) TestValidatePrerequisites_RegistrationFlow_PromptsMobileNumber() {
	// Create a mock that returns false for ValidatePrerequisites (prerequisites not met)
	mockExec := coremock.NewExecutorInterfaceMock(suite.T())
	mockExec.On("ValidatePrerequisites", mock.Anything, mock.Anything).Return(false)
	suite.executor.ExecutorInterface = mockExec

	ctx := &core.NodeContext{
		FlowID:      "test-flow-123",
		FlowType:    common.FlowTypeRegistration,
		UserInputs:  make(map[string]string),
		RuntimeData: make(map[string]string),
	}
	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	result := suite.executor.ValidatePrerequisites(ctx, execResp)

	// Should return false (prerequisites not met)
	assert.False(suite.T(), result)

	// Should set status to ExecUserInputRequired
	assert.Equal(suite.T(), common.ExecUserInputRequired, execResp.Status)

	// Should return mobile number input
	assert.Len(suite.T(), execResp.Inputs, 1)
	assert.Equal(suite.T(), userAttributeMobileNumber, execResp.Inputs[0].Identifier)
	assert.Equal(suite.T(), "PHONE_INPUT", execResp.Inputs[0].Type)
	assert.Equal(suite.T(), "mobile_number_input", execResp.Inputs[0].Ref)
	assert.True(suite.T(), execResp.Inputs[0].Required)
}

func (suite *SMSAuthExecutorTestSuite) TestValidatePrerequisites_RegistrationFlow_PrerequisitesMet() {
	// Create a mock that returns true for ValidatePrerequisites (prerequisites met)
	mockExec := coremock.NewExecutorInterfaceMock(suite.T())
	mockExec.On("ValidatePrerequisites", mock.Anything, mock.Anything).Return(true)
	suite.executor.ExecutorInterface = mockExec

	ctx := &core.NodeContext{
		FlowID:   "test-flow-123",
		FlowType: common.FlowTypeRegistration,
		UserInputs: map[string]string{
			userAttributeMobileNumber: "+1234567890",
		},
		RuntimeData: make(map[string]string),
	}
	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	result := suite.executor.ValidatePrerequisites(ctx, execResp)

	// Should return true (prerequisites met)
	assert.True(suite.T(), result)

	// Status should NOT be set to ExecUserInputRequired
	assert.NotEqual(suite.T(), common.ExecUserInputRequired, execResp.Status)
}

func (suite *SMSAuthExecutorTestSuite) TestValidatePrerequisites_AuthenticationFlow_DoesNotPromptMobile() {
	// Create a mock that returns false initially (prerequisites not met)
	// and also mock additional methods that satisfyPrerequisites might call
	mockExec := coremock.NewExecutorInterfaceMock(suite.T())
	mockExec.On("ValidatePrerequisites", mock.Anything, mock.Anything).Return(false)
	mockExec.On("GetUserIDFromContext", mock.Anything).Return("").Maybe()
	suite.executor.ExecutorInterface = mockExec

	ctx := &core.NodeContext{
		FlowID:      "test-flow-123",
		FlowType:    common.FlowTypeAuthentication, // Authentication flow, NOT registration
		UserInputs:  make(map[string]string),
		RuntimeData: make(map[string]string),
	}
	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	result := suite.executor.ValidatePrerequisites(ctx, execResp)

	assert.False(suite.T(), result, "Should return false when prerequisites not met")
	assert.NotEqual(suite.T(), common.ExecUserInputRequired, execResp.Status,
		"Authentication flows should not prompt for mobile number directly")
}

// TestGetAuthenticatedUser_MFA_AddsMobileNumberToAttributes verifies that when user is already authenticated
func (suite *SMSAuthExecutorTestSuite) TestGetAuthenticatedUser_MFA_AddsMobileNumberToAttributes() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		RuntimeData: map[string]string{
			userAttributeMobileNumber: "+1234567890",
		},
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated:    true,
			UserID:             "user-123",
			OrganizationUnitID: "ou-123",
			UserType:           "INTERNAL",
			Attributes: map[string]interface{}{
				"email": "test@example.com",
				// Mobile number NOT in attributes yet
			},
		},
	}

	execResp := &common.ExecutorResponse{
		RuntimeData: make(map[string]string),
	}

	result, err := suite.executor.getAuthenticatedUser(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.True(suite.T(), result.IsAuthenticated)
	assert.Equal(suite.T(), "user-123", result.UserID)
	// Verify mobile number was added to attributes
	assert.Equal(suite.T(), "+1234567890", result.Attributes[userAttributeMobileNumber])
	assert.Equal(suite.T(), "test@example.com", result.Attributes["email"]) // Existing attributes preserved
}

// TestGetAuthenticatedUser_FetchFromStore_AddsMobileNumberToAttributes verifies that when fetching user
// from store, mobile number from RuntimeData is added to attributes if not present in stored attributes.
func (suite *SMSAuthExecutorTestSuite) TestGetAuthenticatedUser_FetchFromStore_AddsMobileNumberToAttributes() {
	attrs := map[string]interface{}{
		"email": "test@example.com",
		// Mobile number NOT in stored attributes
	}
	attrsJSON, _ := json.Marshal(attrs)

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		RuntimeData: map[string]string{
			userAttributeUserID:       "user-123",
			userAttributeMobileNumber: "+1234567890", // Mobile from RuntimeData
		},
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: false, // User not in context
		},
	}

	execResp := &common.ExecutorResponse{
		RuntimeData: make(map[string]string),
	}

	userFromStore := &userprovider.User{
		UserID:             "user-123",
		OrganizationUnitID: "ou-123",
		UserType:           "INTERNAL",
		Attributes:         attrsJSON,
	}

	suite.mockUserProvider.On("GetUser", "user-123").Return(userFromStore, nil)

	result, err := suite.executor.getAuthenticatedUser(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.True(suite.T(), result.IsAuthenticated)
	assert.Equal(suite.T(), "user-123", result.UserID)
	// Verify mobile number was added to attributes even though it wasn't in stored attributes
	assert.Equal(suite.T(), "+1234567890", result.Attributes[userAttributeMobileNumber])
	assert.Equal(suite.T(), "test@example.com", result.Attributes["email"]) // Existing attributes preserved
	suite.mockUserProvider.AssertExpectations(suite.T())
}

// TestGetAuthenticatedUser_FetchFromStore_PreservesExistingMobileNumber verifies that when fetching user
// from store, if mobile number already exists in stored attributes, it is preserved.
func (suite *SMSAuthExecutorTestSuite) TestGetAuthenticatedUser_FetchFromStore_PreservesExistingMobileNumber() {
	attrs := map[string]interface{}{
		"email":                   "test@example.com",
		userAttributeMobileNumber: "+9876543210", // Mobile already in stored attributes
	}
	attrsJSON, _ := json.Marshal(attrs)

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		RuntimeData: map[string]string{
			userAttributeUserID:       "user-123",
			userAttributeMobileNumber: "+1234567890", // Different mobile in RuntimeData
		},
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: false,
		},
	}

	execResp := &common.ExecutorResponse{
		RuntimeData: make(map[string]string),
	}

	userFromStore := &userprovider.User{
		UserID:             "user-123",
		OrganizationUnitID: "ou-123",
		UserType:           "INTERNAL",
		Attributes:         attrsJSON,
	}

	suite.mockUserProvider.On("GetUser", "user-123").Return(userFromStore, nil)

	result, err := suite.executor.getAuthenticatedUser(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.True(suite.T(), result.IsAuthenticated)
	// Verify stored mobile number is preserved (not overwritten by RuntimeData)
	assert.Equal(suite.T(), "+9876543210", result.Attributes[userAttributeMobileNumber])
	suite.mockUserProvider.AssertExpectations(suite.T())
}

// TestGetUserMobileNumber_NotFoundInAttributesOrContext verifies that when mobile number
// is not found in user attributes or context, the function sets failure status.
func (suite *SMSAuthExecutorTestSuite) TestGetUserMobileNumber_NotFoundInAttributesOrContext() {
	// User attributes without mobile number
	attrs := map[string]interface{}{
		"email": "test@example.com",
		// No mobile number
	}
	attrsJSON, _ := json.Marshal(attrs)

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		// No mobile number in UserInputs or RuntimeData
		UserInputs:  map[string]string{},
		RuntimeData: map[string]string{},
	}

	execResp := &common.ExecutorResponse{
		RuntimeData: make(map[string]string),
	}

	userFromStore := &userprovider.User{
		UserID:             "user-123",
		OrganizationUnitID: "ou-123",
		UserType:           "INTERNAL",
		Attributes:         attrsJSON,
	}

	suite.mockUserProvider.On("GetUser", "user-123").Return(userFromStore, nil)

	mobileNumber, err := suite.executor.getUserMobileNumber("user-123", ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), mobileNumber)
	assert.Equal(suite.T(), common.ExecFailure, execResp.Status)
	assert.Equal(suite.T(), "Mobile number not found in user attributes or context", execResp.FailureReason)
	suite.mockUserProvider.AssertExpectations(suite.T())
}

// TestGetAuthenticatedUser_MFA_NilAttributes verifies that when the authenticated user
// has nil Attributes map, it is initialized before adding mobile number.
func (suite *SMSAuthExecutorTestSuite) TestGetAuthenticatedUser_MFA_NilAttributes() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		RuntimeData: map[string]string{
			userAttributeMobileNumber: "+1234567890",
		},
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated:    true,
			UserID:             "user-123",
			OrganizationUnitID: "ou-123",
			UserType:           "INTERNAL",
			Attributes:         nil, // Explicitly nil
		},
	}

	execResp := &common.ExecutorResponse{
		RuntimeData: make(map[string]string),
	}

	result, err := suite.executor.getAuthenticatedUser(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.True(suite.T(), result.IsAuthenticated)
	assert.NotNil(suite.T(), result.Attributes) // Attributes should be initialized
	assert.Equal(suite.T(), "+1234567890", result.Attributes[userAttributeMobileNumber])
}

// TestGetAuthenticatedUser_FetchFromStore_NilAttrsAfterUnmarshal verifies that when
// user attributes unmarshal to nil, the attrs map is initialized before use.
func (suite *SMSAuthExecutorTestSuite) TestGetAuthenticatedUser_FetchFromStore_NilAttrsAfterUnmarshal() {
	// JSON null unmarshals to nil map
	attrsJSON := []byte("null")

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		RuntimeData: map[string]string{
			userAttributeUserID:       "user-123",
			userAttributeMobileNumber: "+1234567890",
		},
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: false, // Not authenticated, will fetch from store
		},
	}

	execResp := &common.ExecutorResponse{
		RuntimeData: make(map[string]string),
	}

	userFromStore := &userprovider.User{
		UserID:             "user-123",
		OrganizationUnitID: "ou-123",
		UserType:           "INTERNAL",
		Attributes:         attrsJSON, // null JSON
	}

	suite.mockUserProvider.On("GetUser", "user-123").Return(userFromStore, nil)

	result, err := suite.executor.getAuthenticatedUser(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.True(suite.T(), result.IsAuthenticated)
	assert.NotNil(suite.T(), result.Attributes) // Attrs should be initialized from nil
	assert.Equal(suite.T(), "+1234567890", result.Attributes[userAttributeMobileNumber])
	suite.mockUserProvider.AssertExpectations(suite.T())
}
