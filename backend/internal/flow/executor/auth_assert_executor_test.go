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

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	authnassert "github.com/asgardeo/thunder/internal/authn/assert"
	authncm "github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/authnprovider"
	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/userprovider"
	"github.com/asgardeo/thunder/tests/mocks/authn/assertmock"
	"github.com/asgardeo/thunder/tests/mocks/authnprovidermock"
	"github.com/asgardeo/thunder/tests/mocks/flow/coremock"
	"github.com/asgardeo/thunder/tests/mocks/jose/jwtmock"
	"github.com/asgardeo/thunder/tests/mocks/oumock"
	"github.com/asgardeo/thunder/tests/mocks/userprovidermock"
)

type AuthAssertExecutorTestSuite struct {
	suite.Suite
	mockJWTService      *jwtmock.JWTServiceInterfaceMock
	mockOUService       *oumock.OrganizationUnitServiceInterfaceMock
	mockAssertGenerator *assertmock.AuthAssertGeneratorInterfaceMock
	mockAuthnProvider   *authnprovidermock.AuthnProviderInterfaceMock
	mockUserProvider    *userprovidermock.UserProviderInterfaceMock
	mockFlowFactory     *coremock.FlowFactoryInterfaceMock
	executor            *authAssertExecutor
}

func TestAuthAssertExecutorSuite(t *testing.T) {
	suite.Run(t, new(AuthAssertExecutorTestSuite))
}

func (suite *AuthAssertExecutorTestSuite) SetupTest() {
	// Initialize Thunder runtime for JWT config access
	_ = initializeTestRuntime()

	suite.mockJWTService = jwtmock.NewJWTServiceInterfaceMock(suite.T())
	suite.mockOUService = oumock.NewOrganizationUnitServiceInterfaceMock(suite.T())
	suite.mockAssertGenerator = assertmock.NewAuthAssertGeneratorInterfaceMock(suite.T())
	suite.mockAuthnProvider = authnprovidermock.NewAuthnProviderInterfaceMock(suite.T())
	suite.mockUserProvider = userprovidermock.NewUserProviderInterfaceMock(suite.T())
	suite.mockFlowFactory = coremock.NewFlowFactoryInterfaceMock(suite.T())

	mockExec := createMockExecutorSimple(suite.T(), ExecutorNameAuthAssert, common.ExecutorTypeUtility)
	suite.mockFlowFactory.On("CreateExecutor", ExecutorNameAuthAssert, common.ExecutorTypeUtility,
		[]common.Input{}, []common.Input{}).Return(mockExec)

	suite.executor = newAuthAssertExecutor(suite.mockFlowFactory, suite.mockJWTService,
		suite.mockOUService, suite.mockAssertGenerator, suite.mockAuthnProvider, suite.mockUserProvider)
}

func createMockExecutorSimple(t *testing.T, name string,
	executorType common.ExecutorType) core.ExecutorInterface {
	mockExec := coremock.NewExecutorInterfaceMock(t)
	mockExec.On("GetName").Return(name).Maybe()
	mockExec.On("GetType").Return(executorType).Maybe()
	mockExec.On("GetDefaultInputs").Return([]common.Input{}).Maybe()
	mockExec.On("GetPrerequisites").Return([]common.Input{}).Maybe()
	return mockExec
}

func initializeTestRuntime() error {
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "https://test.thunder.io",
			ValidityPeriod: 3600,
		},
	}
	return config.InitializeThunderRuntime("/tmp/test", testConfig)
}

func (suite *AuthAssertExecutorTestSuite) TestNewAuthAssertExecutor() {
	assert.NotNil(suite.T(), suite.executor)
	assert.NotNil(suite.T(), suite.executor.jwtService)
	assert.NotNil(suite.T(), suite.executor.authnProvider)
	assert.NotNil(suite.T(), suite.executor.userProvider)
	assert.NotNil(suite.T(), suite.executor.authAssertGenerator)
}

func (suite *AuthAssertExecutorTestSuite) TestExecute_UserAuthenticated_Success() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		AppID:    "app-123",
		FlowType: common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated:    true,
			UserID:             "user-123",
			OrganizationUnitID: "ou-123",
			UserType:           "INTERNAL",
		},
		ExecutionHistory: map[string]*common.NodeExecutionRecord{
			"node1": {
				ExecutorName: ExecutorNameBasicAuth,
				ExecutorType: common.ExecutorTypeAuthentication,
				Status:       common.FlowStatusComplete,
				Step:         1,
				EndTime:      1234567890,
			},
		},
		Application: appmodel.Application{
			Assertion: &appmodel.AssertionConfig{
				UserAttributes: []string{"userType", "ouId"},
			},
		},
	}

	suite.mockAssertGenerator.On("GenerateAssertion", mock.MatchedBy(func(refs []authncm.AuthenticatorReference) bool {
		return len(refs) == 1 && refs[0].Authenticator == authncm.AuthenticatorCredentials
	})).Return(&authnassert.AssertionResult{
		Context: &authnassert.AssuranceContext{},
	}, nil)

	suite.mockJWTService.On("GenerateJWT", "user-123", "app-123", mock.Anything, mock.Anything,
		mock.Anything).Return("jwt-token", int64(3600), nil)

	suite.mockOUService.On("GetOrganizationUnit", "ou-123").Return(ou.OrganizationUnit{ID: "ou-123"}, nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	assert.Equal(suite.T(), "jwt-token", resp.Assertion)
	suite.mockAssertGenerator.AssertExpectations(suite.T())
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *AuthAssertExecutorTestSuite) TestExecute_UserNotAuthenticated() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: false,
		},
	}

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Equal(suite.T(), failureReasonUserNotAuthenticated, resp.FailureReason)
}

func (suite *AuthAssertExecutorTestSuite) TestExecute_WithAuthorizedPermissions() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		AppID:    "app-123",
		FlowType: common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
			UserID:          "user-123",
		},
		RuntimeData: map[string]string{
			"authorized_permissions": "read:documents write:documents",
		},
		ExecutionHistory: map[string]*common.NodeExecutionRecord{},
		Application:      appmodel.Application{},
	}

	suite.mockJWTService.On("GenerateJWT", "user-123", "app-123", mock.Anything, mock.Anything,
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			perms, ok := claims["authorized_permissions"]
			return ok && perms == "read:documents write:documents"
		})).Return("jwt-token", int64(3600), nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	assert.Equal(suite.T(), "jwt-token", resp.Assertion)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *AuthAssertExecutorTestSuite) TestExecute_WithUserAttributes() {
	attrs := map[string]interface{}{"email": "test@example.com", "phone": "1234567890"}
	attrsJSON, _ := json.Marshal(attrs)

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		AppID:    "app-123",
		FlowType: common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
			UserID:          "user-123",
			Attributes:      map[string]interface{}{"email": "test@example.com"},
		},
		ExecutionHistory: map[string]*common.NodeExecutionRecord{},
		Application: appmodel.Application{
			Assertion: &appmodel.AssertionConfig{
				UserAttributes: []string{"email", "phone"},
			},
		},
	}

	existingUser := &userprovider.User{
		UserID:     "user-123",
		Attributes: attrsJSON,
	}

	suite.mockUserProvider.On("GetUser", "user-123").Return(existingUser, nil)
	suite.mockJWTService.On("GenerateJWT", "user-123", "app-123", mock.Anything, mock.Anything,
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			return claims["email"] == "test@example.com" && claims["phone"] == "1234567890"
		})).Return("jwt-token", int64(3600), nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	suite.mockUserProvider.AssertExpectations(suite.T())
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *AuthAssertExecutorTestSuite) TestExecute_JWTGenerationFails() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		AppID:    "app-123",
		FlowType: common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
			UserID:          "user-123",
		},
		ExecutionHistory: map[string]*common.NodeExecutionRecord{},
		Application:      appmodel.Application{},
	}

	suite.mockJWTService.On("GenerateJWT", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything).Return("", int64(0), &serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "JWT_GENERATION_FAILED",
		Error:            "JWT generation failed",
		ErrorDescription: "Failed to generate JWT token",
	})

	_, err := suite.executor.Execute(ctx)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to generate JWT token")
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *AuthAssertExecutorTestSuite) TestExecute_AssertionGenerationFails_ServerError() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		AppID:    "app-123",
		FlowType: common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
			UserID:          "user-123",
		},
		ExecutionHistory: map[string]*common.NodeExecutionRecord{
			"node1": {
				ExecutorName: ExecutorNameBasicAuth,
				ExecutorType: common.ExecutorTypeAuthentication,
				Status:       common.FlowStatusComplete,
				Step:         1,
			},
		},
		Application: appmodel.Application{},
	}

	suite.mockAssertGenerator.On("GenerateAssertion", mock.Anything).
		Return(nil, &serviceerror.ServiceError{
			Type:  serviceerror.ServerErrorType,
			Error: "internal error",
		})

	_, err := suite.executor.Execute(ctx)

	assert.Error(suite.T(), err)
	suite.mockAssertGenerator.AssertExpectations(suite.T())
}

func (suite *AuthAssertExecutorTestSuite) TestExtractAuthenticatorReferences() {
	history := map[string]*common.NodeExecutionRecord{
		"node1": {
			ExecutorName: ExecutorNameBasicAuth,
			ExecutorType: common.ExecutorTypeAuthentication,
			Status:       common.FlowStatusComplete,
			Step:         3,
			EndTime:      1000,
		},
		"node2": {
			ExecutorName: ExecutorNameSMSAuth,
			ExecutorType: common.ExecutorTypeAuthentication,
			Status:       common.FlowStatusComplete,
			Step:         1,
			EndTime:      2000,
		},
		"node3": {
			ExecutorName: ExecutorNameProvisioning,
			ExecutorType: common.ExecutorTypeRegistration,
			Status:       common.FlowStatusComplete,
			Step:         2,
		},
		"node4": {
			ExecutorName: ExecutorNameOAuth,
			ExecutorType: common.ExecutorTypeAuthentication,
			Status:       common.FlowStatusError,
			Step:         4,
		},
	}

	refs := suite.executor.extractAuthenticatorReferences(history)

	assert.Len(suite.T(), refs, 2)
	assert.Equal(suite.T(), authncm.AuthenticatorSMSOTP, refs[0].Authenticator)
	assert.Equal(suite.T(), 1, refs[0].Step)
	assert.Equal(suite.T(), authncm.AuthenticatorCredentials, refs[1].Authenticator)
	assert.Equal(suite.T(), 2, refs[1].Step)
}

func (suite *AuthAssertExecutorTestSuite) TestExtractAuthenticatorReferences_EmptyHistory() {
	history := map[string]*common.NodeExecutionRecord{}

	refs := suite.executor.extractAuthenticatorReferences(history)

	assert.Empty(suite.T(), refs)
}

func (suite *AuthAssertExecutorTestSuite) TestExtractAuthenticatorReferences_UnknownExecutor() {
	history := map[string]*common.NodeExecutionRecord{
		"node1": {
			ExecutorName: "UnknownExecutor",
			ExecutorType: common.ExecutorTypeAuthentication,
			Status:       common.FlowStatusComplete,
			Step:         1,
		},
	}

	refs := suite.executor.extractAuthenticatorReferences(history)

	assert.Empty(suite.T(), refs)
}

func (suite *AuthAssertExecutorTestSuite) TestExtractAuthenticatorReferences_SMSOTPSendVerifyMode() {
	history := map[string]*common.NodeExecutionRecord{
		"sms_send_node": {
			ExecutorName: ExecutorNameSMSAuth,
			ExecutorType: common.ExecutorTypeAuthentication,
			ExecutorMode: "send",
			Status:       common.FlowStatusComplete,
			Step:         1,
			EndTime:      1000,
		},
		"sms_verify_node": {
			ExecutorName: ExecutorNameSMSAuth,
			ExecutorType: common.ExecutorTypeAuthentication,
			ExecutorMode: "verify",
			Status:       common.FlowStatusComplete,
			Step:         2,
			EndTime:      2000,
		},
	}

	refs := suite.executor.extractAuthenticatorReferences(history)

	// Should only have one SMS OTP authenticator, not two
	assert.Len(suite.T(), refs, 1)
	assert.Equal(suite.T(), authncm.AuthenticatorSMSOTP, refs[0].Authenticator)
	assert.Equal(suite.T(), 1, refs[0].Step)
}

func (suite *AuthAssertExecutorTestSuite) TestGetUserAttributes_Success() {
	attrs := map[string]interface{}{"email": "test@example.com", "name": "Test User"}
	attrsJSON, _ := json.Marshal(attrs)

	existingUser := &userprovider.User{
		UserID:     "user-123",
		Attributes: attrsJSON,
	}

	suite.mockUserProvider.On("GetUser", "user-123").Return(existingUser, nil)

	resultAttrs, err := suite.executor.getUserAttributes("user-123", "", nil, nil)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resultAttrs)
	assert.Equal(suite.T(), "test@example.com", resultAttrs["email"])
	assert.Equal(suite.T(), "Test User", resultAttrs["name"])
	suite.mockUserProvider.AssertExpectations(suite.T())
}

func (suite *AuthAssertExecutorTestSuite) TestGetUserAttributes_ServiceError() {
	suite.mockUserProvider.On("GetUser", "user-123").
		Return(nil, &userprovider.UserProviderError{Message: "user not found"})

	resultAttrs, err := suite.executor.getUserAttributes("user-123", "", nil, nil)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), resultAttrs)
	suite.mockUserProvider.AssertExpectations(suite.T())
}

func (suite *AuthAssertExecutorTestSuite) TestGetUserAttributes_InvalidJSON() {
	existingUser := &userprovider.User{
		UserID:     "user-123",
		Attributes: json.RawMessage(`invalid json`),
	}

	suite.mockUserProvider.On("GetUser", "user-123").Return(existingUser, nil)

	resultAttrs, err := suite.executor.getUserAttributes("user-123", "", nil, nil)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), resultAttrs)
	suite.mockUserProvider.AssertExpectations(suite.T())
}

func (suite *AuthAssertExecutorTestSuite) TestGetUserAttributes_WithToken_Success() {
	attrs := map[string]interface{}{"email": "test@example.com", "name": "Test User"}
	attrsJSON, _ := json.Marshal(attrs)

	res := authnprovider.GetAttributesResult{
		Attributes: attrsJSON,
	}

	suite.mockAuthnProvider.On("GetAttributes", "token-123", []string{"email", "name"},
		(*authnprovider.GetAttributesMetadata)(nil)).Return(&res, nil)

	resultAttrs, err := suite.executor.getUserAttributes("user-123", "token-123", []string{"email", "name"}, nil)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resultAttrs)
	assert.Equal(suite.T(), "test@example.com", resultAttrs["email"])
	assert.Equal(suite.T(), "Test User", resultAttrs["name"])
	suite.mockAuthnProvider.AssertExpectations(suite.T())
}

func (suite *AuthAssertExecutorTestSuite) TestExecute_WithUserTypeAndOU() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		AppID:    "app-123",
		FlowType: common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated:    true,
			UserID:             "user-123",
			UserType:           "EXTERNAL",
			OrganizationUnitID: "ou-456",
		},
		ExecutionHistory: map[string]*common.NodeExecutionRecord{},
		Application: appmodel.Application{
			Assertion: &appmodel.AssertionConfig{
				UserAttributes: []string{"userType", "ouId"},
			},
		},
	}

	suite.mockJWTService.On("GenerateJWT", "user-123", "app-123", mock.Anything, mock.Anything,
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			return claims[oauth2const.ClaimUserType] == "EXTERNAL" && claims[oauth2const.ClaimOUID] == "ou-456"
		})).Return("jwt-token", int64(3600), nil)

	suite.mockOUService.On("GetOrganizationUnit", "ou-456").Return(ou.OrganizationUnit{ID: "ou-456"}, nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *AuthAssertExecutorTestSuite) TestExecute_WithCustomTokenConfig() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		AppID:    "app-123",
		FlowType: common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
			UserID:          "user-123",
		},
		ExecutionHistory: map[string]*common.NodeExecutionRecord{},
		Application: appmodel.Application{
			Assertion: &appmodel.AssertionConfig{
				Issuer:         "custom-issuer",
				ValidityPeriod: 7200,
			},
		},
	}

	suite.mockJWTService.On("GenerateJWT", "user-123", "app-123", "custom-issuer", int64(7200),
		mock.Anything).Return("jwt-token", int64(7200), nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *AuthAssertExecutorTestSuite) TestExecute_WithOUNameAndHandle() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		AppID:    "app-123",
		FlowType: common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated:    true,
			UserID:             "user-123",
			OrganizationUnitID: "ou-789",
		},
		ExecutionHistory: map[string]*common.NodeExecutionRecord{},
		Application: appmodel.Application{
			Assertion: &appmodel.AssertionConfig{
				UserAttributes: []string{"ouId", "ouName", "ouHandle"},
			},
		},
	}

	suite.mockOUService.On("GetOrganizationUnit", "ou-789").Return(ou.OrganizationUnit{
		ID:     "ou-789",
		Name:   "Engineering",
		Handle: "eng",
	}, nil)

	suite.mockJWTService.On("GenerateJWT", "user-123", "app-123", mock.Anything, mock.Anything,
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			return claims[oauth2const.ClaimOUID] == "ou-789" &&
				claims[oauth2const.ClaimOUName] == "Engineering" &&
				claims[oauth2const.ClaimOUHandle] == "eng"
		})).Return("jwt-token", int64(3600), nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	assert.Equal(suite.T(), "jwt-token", resp.Assertion)
	suite.mockOUService.AssertExpectations(suite.T())
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *AuthAssertExecutorTestSuite) TestExecute_AppendUserDetailsToClaimsFails() {
	attrs := map[string]interface{}{"email": "test@example.com"}
	attrsJSON, _ := json.Marshal(attrs)

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		AppID:    "app-123",
		FlowType: common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
			UserID:          "user-123",
		},
		ExecutionHistory: map[string]*common.NodeExecutionRecord{},
		Application: appmodel.Application{
			Assertion: &appmodel.AssertionConfig{
				UserAttributes: []string{"email"},
			},
		},
	}

	// Test case 1: GetUser returns service error
	suite.mockUserProvider.On("GetUser", "user-123").
		Return(nil, &userprovider.UserProviderError{
			Message:     "user_not_found",
			Description: "user not found",
		})

	_, err := suite.executor.Execute(ctx)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "something went wrong while fetching user attributes")
	suite.mockUserProvider.AssertExpectations(suite.T())

	// Reset mock for test case 2
	suite.mockUserProvider = userprovidermock.NewUserProviderInterfaceMock(suite.T())
	suite.executor.userProvider = suite.mockUserProvider

	// Test case 2: Invalid JSON in user attributes
	existingUser := &userprovider.User{
		UserID:     "user-123",
		Attributes: json.RawMessage(`{invalid json}`),
	}

	suite.mockUserProvider.On("GetUser", "user-123").Return(existingUser, nil)

	_, err = suite.executor.Execute(ctx)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "something went wrong while unmarshalling user attributes")
	suite.mockUserProvider.AssertExpectations(suite.T())

	// Test success case for comparison
	suite.mockUserProvider = userprovidermock.NewUserProviderInterfaceMock(suite.T())
	suite.executor.userProvider = suite.mockUserProvider

	existingUser.Attributes = attrsJSON
	suite.mockUserProvider.On("GetUser", "user-123").Return(existingUser, nil)
	suite.mockJWTService.On("GenerateJWT", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything).Return("jwt-token", int64(3600), nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
}

func (suite *AuthAssertExecutorTestSuite) TestExecute_AppendOUDetailsToClaimsFails() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		AppID:    "app-123",
		FlowType: common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated:    true,
			UserID:             "user-123",
			OrganizationUnitID: "ou-123",
		},
		ExecutionHistory: map[string]*common.NodeExecutionRecord{},
		Application: appmodel.Application{
			Assertion: &appmodel.AssertionConfig{
				UserAttributes: []string{oauth2const.ClaimOUID},
			},
		},
	}

	suite.mockOUService.On("GetOrganizationUnit", "ou-123").
		Return(ou.OrganizationUnit{}, &serviceerror.ServiceError{
			Error:            "ou_not_found",
			ErrorDescription: "organization unit not found",
		})

	_, err := suite.executor.Execute(ctx)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "something went wrong while fetching organization unit")
	suite.mockOUService.AssertExpectations(suite.T())
}

func (suite *AuthAssertExecutorTestSuite) TestAppendUserDetailsToClaims_GetUserAttributesFails() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		AppID:    "app-123",
		FlowType: common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
			UserID:          "user-123",
			Attributes:      map[string]interface{}{"email": "test@example.com"},
		},
		ExecutionHistory: map[string]*common.NodeExecutionRecord{},
		Application: appmodel.Application{
			Assertion: &appmodel.AssertionConfig{
				UserAttributes: []string{"email", "phone"},
			},
		},
	}

	suite.mockUserProvider.On("GetUser", "user-123").
		Return(nil, &userprovider.UserProviderError{
			Message:     "database_error",
			Description: "failed to fetch user",
		})

	_, err := suite.executor.Execute(ctx)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "something went wrong while fetching user attributes")
	suite.mockUserProvider.AssertExpectations(suite.T())
}

func (suite *AuthAssertExecutorTestSuite) TestAppendOUDetailsToClaims_GetOrganizationUnitFails() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		AppID:    "app-123",
		FlowType: common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated:    true,
			UserID:             "user-123",
			OrganizationUnitID: "ou-invalid",
		},
		ExecutionHistory: map[string]*common.NodeExecutionRecord{},
		Application: appmodel.Application{
			Assertion: &appmodel.AssertionConfig{
				UserAttributes: []string{oauth2const.ClaimOUID},
			},
		},
	}

	suite.mockOUService.On("GetOrganizationUnit", "ou-invalid").
		Return(ou.OrganizationUnit{}, &serviceerror.ServiceError{
			Error:            "ou_not_found",
			ErrorDescription: "organization unit does not exist",
		})

	_, err := suite.executor.Execute(ctx)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "something went wrong while fetching organization unit")
	assert.Contains(suite.T(), err.Error(), "organization unit does not exist")
	suite.mockOUService.AssertExpectations(suite.T())
}

func (suite *AuthAssertExecutorTestSuite) TestExecute_WithConfiguredUserAttributes() {
	attrs := map[string]interface{}{"email": "test@example.com", "username": "testuser", "firstName": "Test"}
	attrsJSON, _ := json.Marshal(attrs)

	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		AppID:    "app-123",
		FlowType: common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
			UserID:          "user-123",
		},
		ExecutionHistory: map[string]*common.NodeExecutionRecord{},
		Application: appmodel.Application{
			// Token config with user attributes configured
			Assertion: &appmodel.AssertionConfig{
				UserAttributes: []string{"email", "username", "firstName"},
			},
		},
	}

	existingUser := &userprovider.User{
		UserID:     "user-123",
		Attributes: attrsJSON,
	}

	suite.mockUserProvider.On("GetUser", "user-123").Return(existingUser, nil)
	suite.mockJWTService.On("GenerateJWT", "user-123", "app-123", mock.Anything, mock.Anything,
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			// Should contain the configured user attributes from the user store
			hasEmail := claims["email"] == "test@example.com"
			hasUsername := claims["username"] == "testuser"
			hasFirstName := claims["firstName"] == "Test"
			return hasEmail && hasUsername && hasFirstName
		})).Return("jwt-token", int64(3600), nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	suite.mockUserProvider.AssertExpectations(suite.T())
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *AuthAssertExecutorTestSuite) TestExecute_WithGroups() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		AppID:    "app-123",
		FlowType: common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
			UserID:          "user-123",
		},
		ExecutionHistory: map[string]*common.NodeExecutionRecord{},
		Application: appmodel.Application{
			Assertion: &appmodel.AssertionConfig{
				UserAttributes: []string{oauth2const.UserAttributeGroups},
			},
		},
	}

	userGroups := &userprovider.UserGroupListResponse{
		Groups: []userprovider.UserGroup{
			{Name: "admin"},
			{Name: "developer"},
			{Name: "viewer"},
		},
	}

	suite.mockUserProvider.On("GetUserGroups", "user-123", oauth2const.DefaultGroupListLimit, 0).
		Return(userGroups, nil)
	suite.mockJWTService.On("GenerateJWT", "user-123", "app-123", mock.Anything, mock.Anything,
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			// Should contain groups claim
			groups, ok := claims[oauth2const.UserAttributeGroups].([]string)
			if !ok {
				return false
			}
			return len(groups) == 3 && groups[0] == "admin" && groups[1] == "developer" && groups[2] == "viewer"
		})).Return("jwt-token", int64(3600), nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	suite.mockUserProvider.AssertExpectations(suite.T())
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *AuthAssertExecutorTestSuite) TestExecute_WithGroups_EmptyGroups() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		AppID:    "app-123",
		FlowType: common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
			UserID:          "user-123",
		},
		ExecutionHistory: map[string]*common.NodeExecutionRecord{},
		Application: appmodel.Application{
			Assertion: &appmodel.AssertionConfig{
				UserAttributes: []string{oauth2const.UserAttributeGroups},
			},
		},
	}

	userGroups := &userprovider.UserGroupListResponse{
		Groups: []userprovider.UserGroup{},
	}

	suite.mockUserProvider.On("GetUserGroups", "user-123", oauth2const.DefaultGroupListLimit, 0).
		Return(userGroups, nil)
	suite.mockJWTService.On("GenerateJWT", "user-123", "app-123", mock.Anything, mock.Anything,
		mock.MatchedBy(func(claims map[string]interface{}) bool {
			// Should NOT contain groups claim when groups list is empty
			_, ok := claims[oauth2const.UserAttributeGroups]
			return !ok
		})).Return("jwt-token", int64(3600), nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	suite.mockUserProvider.AssertExpectations(suite.T())
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *AuthAssertExecutorTestSuite) TestExecute_WithGroups_GetUserGroupsFails() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		AppID:    "app-123",
		FlowType: common.FlowTypeAuthentication,
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
			UserID:          "user-123",
		},
		ExecutionHistory: map[string]*common.NodeExecutionRecord{},
		Application: appmodel.Application{
			Assertion: &appmodel.AssertionConfig{
				UserAttributes: []string{oauth2const.UserAttributeGroups},
			},
		},
	}

	suite.mockUserProvider.On("GetUserGroups", "user-123", oauth2const.DefaultGroupListLimit, 0).
		Return(nil, &userprovider.UserProviderError{Message: "failed to fetch groups", Description: "database error"})

	resp, err := suite.executor.Execute(ctx)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), resp)
	assert.Contains(suite.T(), err.Error(), "something went wrong while fetching user groups")
	suite.mockUserProvider.AssertExpectations(suite.T())
}
