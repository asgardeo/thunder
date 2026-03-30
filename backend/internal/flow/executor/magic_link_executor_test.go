/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	authncm "github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/authn/magiclink"
	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/userprovider"
	"github.com/asgardeo/thunder/tests/mocks/authn/magiclinkmock"
	"github.com/asgardeo/thunder/tests/mocks/flow/coremock"
	"github.com/asgardeo/thunder/tests/mocks/observability/observabilitymock"
	"github.com/asgardeo/thunder/tests/mocks/userprovidermock"
)

const (
	magicLinkTestUserID       = "user-123"
	magicLinkTestEmail        = "test@example.com"
	magicLinkTestMagicToken   = "magic-token-123"
	magicLinkTestFlowID       = "flow-123"
	magicLinkTestOUID         = "ou-123"
	magicLinkTestUserType     = "INTERNAL"
	magicLinkTestMagicLinkURL = "https://example.com/verify"
)

type MagicLinkAuthExecutorTestSuite struct {
	suite.Suite
	mockMagicLinkService *magiclinkmock.MagicLinkAuthnServiceInterfaceMock
	mockFlowFactory      *coremock.FlowFactoryInterfaceMock
	mockObservability    *observabilitymock.ObservabilityServiceInterfaceMock
	mockUserProvider     *userprovidermock.UserProviderInterfaceMock
	executor             *magicLinkAuthExecutor
}

func TestMagicLinkAuthExecutorSuite(t *testing.T) {
	suite.Run(t, new(MagicLinkAuthExecutorTestSuite))
}

func (suite *MagicLinkAuthExecutorTestSuite) SetupTest() {
	suite.mockMagicLinkService = magiclinkmock.NewMagicLinkAuthnServiceInterfaceMock(suite.T())
	suite.mockFlowFactory = coremock.NewFlowFactoryInterfaceMock(suite.T())
	suite.mockObservability = observabilitymock.NewObservabilityServiceInterfaceMock(suite.T())
	suite.mockUserProvider = userprovidermock.NewUserProviderInterfaceMock(suite.T())

	suite.mockObservability.On("IsEnabled").Return(false).Maybe()

	defaultInputs := []common.Input{MagicLinkTokenInput}
	prerequisites := []common.Input{EmailInput}

	identifyingMock := createMockIdentifyingExecutor(suite.T())
	suite.mockFlowFactory.On("CreateExecutor", ExecutorNameIdentifying, common.ExecutorTypeUtility,
		mock.Anything, mock.Anything).Return(identifyingMock).Maybe()

	mockExec := createMockMagicLinkAuthExecutor(suite.T())
	suite.mockFlowFactory.On("CreateExecutor", ExecutorNameMagicLinkAuth, common.ExecutorTypeAuthentication,
		defaultInputs, prerequisites).Return(mockExec)

	suite.executor = newMagicLinkAuthExecutor(suite.mockFlowFactory, suite.mockMagicLinkService,
		suite.mockObservability, suite.mockUserProvider)
	suite.executor.ExecutorInterface = mockExec
}

func createMockMagicLinkAuthExecutor(t *testing.T) core.ExecutorInterface {
	mockExec := coremock.NewExecutorInterfaceMock(t)
	mockExec.On("GetName").Return(ExecutorNameMagicLinkAuth).Maybe()
	mockExec.On("GetType").Return(common.ExecutorTypeAuthentication).Maybe()
	mockExec.On("GetDefaultInputs").Return([]common.Input{MagicLinkTokenInput}).Maybe()
	mockExec.On("GetRequiredInputs", mock.Anything).Return([]common.Input{MagicLinkTokenInput}).Maybe()
	mockExec.On("GetPrerequisites").Return([]common.Input{EmailInput}).Maybe()
	mockExec.On("ValidatePrerequisites", mock.Anything, mock.Anything).Return(true).Maybe()
	mockExec.On("HasRequiredInputs", mock.Anything, mock.Anything).Return(
		func(ctx *core.NodeContext, execResp *common.ExecutorResponse) bool {
			token, exists := ctx.UserInputs[userInputMagicLinkToken]
			if !exists || token == "" {
				execResp.Inputs = []common.Input{MagicLinkTokenInput}
				execResp.Status = common.ExecUserInputRequired
				return false
			}
			return true
		}).Maybe()
	return mockExec
}

func (suite *MagicLinkAuthExecutorTestSuite) TestNewMagicLinkAuthExecutor() {
	assert.NotNil(suite.T(), suite.executor)
	assert.NotNil(suite.T(), suite.executor.magicLinkService)
	assert.NotNil(suite.T(), suite.executor.userProvider)
}

// Test Send Mode

func (suite *MagicLinkAuthExecutorTestSuite) TestExecute_SendMode_Success_AuthenticationFlow() {
	ctx := &core.NodeContext{
		Context:      context.Background(),
		FlowID:       magicLinkTestFlowID,
		FlowType:     common.FlowTypeAuthentication,
		ExecutorMode: ExecutorModeSend,
		UserInputs: map[string]string{
			userAttributeEmail: magicLinkTestEmail,
		},
		RuntimeData: make(map[string]string),
	}

	suite.mockMagicLinkService.On("SendMagicLink", ctx.Context, magicLinkTestEmail,
		int64(magiclink.DefaultExpirySeconds), magicLinkTestFlowID, "").Return(nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	assert.Empty(suite.T(), resp.RuntimeData[userAttributeUserID])
	suite.mockUserProvider.AssertNotCalled(suite.T(), "IdentifyUser", map[string]interface{}{
		userAttributeEmail: magicLinkTestEmail,
	})
	suite.mockMagicLinkService.AssertExpectations(suite.T())
}

func (suite *MagicLinkAuthExecutorTestSuite) TestExecute_SendMode_Success_RegistrationFlow_NewUser() {
	ctx := &core.NodeContext{
		Context:      context.Background(),
		FlowID:       magicLinkTestFlowID,
		FlowType:     common.FlowTypeRegistration,
		ExecutorMode: ExecutorModeSend,
		UserInputs: map[string]string{
			userAttributeEmail: magicLinkTestEmail,
		},
		RuntimeData: make(map[string]string),
	}

	suite.mockUserProvider.On("IdentifyUser", map[string]interface{}{
		userAttributeEmail: magicLinkTestEmail,
	}).Return(nil, userprovider.NewUserProviderError(userprovider.ErrorCodeUserNotFound, "", ""))

	suite.mockMagicLinkService.On("SendMagicLink", ctx.Context, magicLinkTestEmail,
		int64(magiclink.DefaultExpirySeconds), magicLinkTestFlowID, "").Return(nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	suite.mockUserProvider.AssertExpectations(suite.T())
	suite.mockMagicLinkService.AssertExpectations(suite.T())
}

func (suite *MagicLinkAuthExecutorTestSuite) TestExecute_SendMode_Failure_RegistrationFlow_UserExists() {
	ctx := &core.NodeContext{
		Context:      context.Background(),
		FlowID:       magicLinkTestFlowID,
		FlowType:     common.FlowTypeRegistration,
		ExecutorMode: ExecutorModeSend,
		UserInputs: map[string]string{
			userAttributeEmail: magicLinkTestEmail,
		},
		RuntimeData: make(map[string]string),
	}

	userID := magicLinkTestUserID
	suite.mockUserProvider.On("IdentifyUser", map[string]interface{}{
		userAttributeEmail: magicLinkTestEmail,
	}).Return(&userID, nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Equal(suite.T(), "User already exists with the provided email.", resp.FailureReason)
	suite.mockUserProvider.AssertExpectations(suite.T())
}

func (suite *MagicLinkAuthExecutorTestSuite) TestExecute_SendMode_Success_WithCustomTokenExpiry() {
	ctx := &core.NodeContext{
		Context:      context.Background(),
		FlowID:       magicLinkTestFlowID,
		FlowType:     common.FlowTypeAuthentication,
		ExecutorMode: ExecutorModeSend,
		UserInputs: map[string]string{
			userAttributeEmail: magicLinkTestEmail,
		},
		RuntimeData: make(map[string]string),
		NodeProperties: map[string]interface{}{
			propertyKeyTokenExpiry: "600",
		},
	}

	suite.mockMagicLinkService.On("SendMagicLink", ctx.Context, magicLinkTestEmail,
		int64(600), magicLinkTestFlowID, "").Return(nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	suite.mockMagicLinkService.AssertExpectations(suite.T())
}

func (suite *MagicLinkAuthExecutorTestSuite) TestExecute_SendMode_Success_WithCustomMagicLinkURL() {
	ctx := &core.NodeContext{
		Context:      context.Background(),
		FlowID:       magicLinkTestFlowID,
		FlowType:     common.FlowTypeAuthentication,
		ExecutorMode: ExecutorModeSend,
		UserInputs: map[string]string{
			userAttributeEmail: magicLinkTestEmail,
		},
		RuntimeData: make(map[string]string),
		NodeProperties: map[string]interface{}{
			propertyKeyMagicLinkURL: magicLinkTestMagicLinkURL,
		},
	}

	suite.mockMagicLinkService.On("SendMagicLink", ctx.Context, magicLinkTestEmail,
		int64(magiclink.DefaultExpirySeconds), magicLinkTestFlowID, magicLinkTestMagicLinkURL).Return(nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	suite.mockMagicLinkService.AssertExpectations(suite.T())
}

func (suite *MagicLinkAuthExecutorTestSuite) TestExecute_SendMode_Failure_MagicLinkServiceNotConfigured() {
	suite.executor.magicLinkService = nil

	ctx := &core.NodeContext{
		Context:      context.Background(),
		FlowID:       magicLinkTestFlowID,
		FlowType:     common.FlowTypeAuthentication,
		ExecutorMode: ExecutorModeSend,
		UserInputs: map[string]string{
			userAttributeEmail: magicLinkTestEmail,
		},
		RuntimeData: make(map[string]string),
	}

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Equal(suite.T(), "Authentication method is currently unavailable", resp.FailureReason)
}

func (suite *MagicLinkAuthExecutorTestSuite) TestExecute_SendMode_Failure_SendMagicLinkError() {
	ctx := &core.NodeContext{
		Context:      context.Background(),
		FlowID:       magicLinkTestFlowID,
		FlowType:     common.FlowTypeAuthentication,
		ExecutorMode: ExecutorModeSend,
		UserInputs: map[string]string{
			userAttributeEmail: magicLinkTestEmail,
		},
		RuntimeData: make(map[string]string),
	}

	suite.mockMagicLinkService.On("SendMagicLink", ctx.Context, magicLinkTestEmail,
		int64(magiclink.DefaultExpirySeconds), magicLinkTestFlowID, "").Return(&serviceerror.I18nServiceError{
		Code: "AUTHN-ML-5002",
	})

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Equal(suite.T(), "Failed to send magic link", resp.FailureReason)
	suite.mockMagicLinkService.AssertExpectations(suite.T())
}

func (suite *MagicLinkAuthExecutorTestSuite) TestExecute_SendMode_Success_AuthenticationFlow_DoesNotIdentifyUser() {
	ctx := &core.NodeContext{
		Context:      context.Background(),
		FlowID:       magicLinkTestFlowID,
		FlowType:     common.FlowTypeAuthentication,
		ExecutorMode: ExecutorModeSend,
		UserInputs: map[string]string{
			userAttributeEmail: magicLinkTestEmail,
		},
		RuntimeData: make(map[string]string),
	}

	suite.mockMagicLinkService.On("SendMagicLink", ctx.Context, magicLinkTestEmail,
		int64(magiclink.DefaultExpirySeconds), magicLinkTestFlowID, "").Return(nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	suite.mockUserProvider.AssertNotCalled(suite.T(), "IdentifyUser", map[string]interface{}{
		userAttributeEmail: magicLinkTestEmail,
	})
	suite.mockMagicLinkService.AssertExpectations(suite.T())
}

func (suite *MagicLinkAuthExecutorTestSuite) TestExecute_SendMode_Success_WithAuthenticatedUser() {
	ctx := &core.NodeContext{
		Context:      context.Background(),
		FlowID:       magicLinkTestFlowID,
		FlowType:     common.FlowTypeAuthentication,
		ExecutorMode: ExecutorModeSend,
		UserInputs: map[string]string{
			userAttributeEmail: magicLinkTestEmail,
		},
		RuntimeData: map[string]string{
			userAttributeUserID: magicLinkTestUserID,
		},
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
			UserID:          magicLinkTestUserID,
		},
	}

	mockExec := coremock.NewExecutorInterfaceMock(suite.T())
	mockExec.On("GetName").Return(ExecutorNameMagicLinkAuth).Maybe()
	mockExec.On("GetType").Return(common.ExecutorTypeAuthentication).Maybe()
	mockExec.On("GetDefaultInputs").Return([]common.Input{MagicLinkTokenInput}).Maybe()
	mockExec.On("GetRequiredInputs", mock.Anything).Return([]common.Input{MagicLinkTokenInput}).Maybe()
	mockExec.On("GetPrerequisites").Return([]common.Input{EmailInput}).Maybe()
	mockExec.On("ValidatePrerequisites", mock.Anything, mock.Anything).Return(true).Maybe()
	mockExec.On("GetUserIDFromContext", mock.Anything).Return(magicLinkTestUserID).Maybe()
	suite.executor.ExecutorInterface = mockExec

	suite.mockMagicLinkService.On("SendMagicLink", ctx.Context, magicLinkTestEmail,
		int64(magiclink.DefaultExpirySeconds), magicLinkTestFlowID, "").Return(nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	suite.mockMagicLinkService.AssertExpectations(suite.T())
}

// Test Verify Mode

func (suite *MagicLinkAuthExecutorTestSuite) TestExecute_VerifyMode_Success() {
	attrs := map[string]interface{}{
		"email": magicLinkTestEmail,
	}
	attrsJSON, _ := json.Marshal(attrs)

	ctx := &core.NodeContext{
		Context:      context.Background(),
		FlowID:       magicLinkTestFlowID,
		FlowType:     common.FlowTypeAuthentication,
		ExecutorMode: ExecutorModeVerify,
		UserInputs: map[string]string{
			userInputMagicLinkToken: magicLinkTestMagicToken,
		},
		RuntimeData: make(map[string]string),
	}

	user := &userprovider.User{
		UserID:     magicLinkTestUserID,
		UserType:   magicLinkTestUserType,
		OUID:       magicLinkTestOUID,
		Attributes: attrsJSON,
	}

	suite.mockMagicLinkService.On("VerifyMagicLink", ctx.Context, magicLinkTestMagicToken).Return(user, nil)
	suite.mockUserProvider.On("GetUser", magicLinkTestUserID).Return(user, nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	assert.True(suite.T(), resp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), magicLinkTestUserID, resp.AuthenticatedUser.UserID)
	assert.Equal(suite.T(), magicLinkTestUserType, resp.AuthenticatedUser.UserType)
	assert.Equal(suite.T(), magicLinkTestOUID, resp.AuthenticatedUser.OUID)
	suite.mockMagicLinkService.AssertExpectations(suite.T())
	suite.mockUserProvider.AssertExpectations(suite.T())
}

func (suite *MagicLinkAuthExecutorTestSuite) TestExecute_VerifyMode_Failure_TokenNotProvided() {
	ctx := &core.NodeContext{
		Context:      context.Background(),
		FlowID:       magicLinkTestFlowID,
		FlowType:     common.FlowTypeAuthentication,
		ExecutorMode: ExecutorModeVerify,
		UserInputs:   make(map[string]string),
		RuntimeData:  make(map[string]string),
	}

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecUserInputRequired, resp.Status)
	assert.Len(suite.T(), resp.Inputs, 1)
	assert.Equal(suite.T(), userInputMagicLinkToken, resp.Inputs[0].Identifier)
}

func (suite *MagicLinkAuthExecutorTestSuite) TestExecute_VerifyMode_Failure_InvalidToken() {
	ctx := &core.NodeContext{
		Context:      context.Background(),
		FlowID:       magicLinkTestFlowID,
		FlowType:     common.FlowTypeAuthentication,
		ExecutorMode: ExecutorModeVerify,
		UserInputs: map[string]string{
			userInputMagicLinkToken: "invalid-token",
		},
		RuntimeData: make(map[string]string),
	}

	suite.mockMagicLinkService.On("VerifyMagicLink", ctx.Context, "invalid-token").Return(
		nil, &serviceerror.I18nServiceError{Code: "AUTHN-ML-1002"})

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Equal(suite.T(), failureReasonInvalidMagicLinkToken, resp.FailureReason)
	suite.mockMagicLinkService.AssertExpectations(suite.T())
}

func (suite *MagicLinkAuthExecutorTestSuite) TestExecute_VerifyMode_Failure_UserNotFoundAfterVerification() {
	ctx := &core.NodeContext{
		Context:      context.Background(),
		FlowID:       magicLinkTestFlowID,
		FlowType:     common.FlowTypeAuthentication,
		ExecutorMode: ExecutorModeVerify,
		UserInputs: map[string]string{
			userInputMagicLinkToken: magicLinkTestMagicToken,
		},
		RuntimeData: make(map[string]string),
	}

	suite.mockMagicLinkService.On("VerifyMagicLink", ctx.Context, magicLinkTestMagicToken).Return(nil, nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecFailure, resp.Status)
	assert.Equal(suite.T(), failureReasonUserNotFound, resp.FailureReason)
	suite.mockMagicLinkService.AssertExpectations(suite.T())
}

func (suite *MagicLinkAuthExecutorTestSuite) TestExecute_VerifyMode_Failure_GetUserError() {
	ctx := &core.NodeContext{
		Context:      context.Background(),
		FlowID:       magicLinkTestFlowID,
		FlowType:     common.FlowTypeAuthentication,
		ExecutorMode: ExecutorModeVerify,
		UserInputs: map[string]string{
			userInputMagicLinkToken: magicLinkTestMagicToken,
		},
		RuntimeData: make(map[string]string),
	}

	user := &userprovider.User{
		UserID:   magicLinkTestUserID,
		UserType: magicLinkTestUserType,
		OUID:     magicLinkTestOUID,
	}

	suite.mockMagicLinkService.On("VerifyMagicLink", ctx.Context, magicLinkTestMagicToken).Return(user, nil)
	suite.mockUserProvider.On("GetUser", magicLinkTestUserID).Return(nil,
		userprovider.NewUserProviderError(userprovider.ErrorCodeSystemError, "database error", ""))

	_, err := suite.executor.Execute(ctx)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to get user")
	suite.mockMagicLinkService.AssertExpectations(suite.T())
	suite.mockUserProvider.AssertExpectations(suite.T())
}

// Test Invalid Executor Mode

func (suite *MagicLinkAuthExecutorTestSuite) TestExecute_InvalidMode() {
	ctx := &core.NodeContext{
		Context:      context.Background(),
		FlowID:       magicLinkTestFlowID,
		FlowType:     common.FlowTypeAuthentication,
		ExecutorMode: "invalid-mode",
		UserInputs: map[string]string{
			userAttributeEmail: magicLinkTestEmail,
		},
		RuntimeData: make(map[string]string),
	}

	resp, err := suite.executor.Execute(ctx)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "invalid executor mode")
	assert.NotNil(suite.T(), resp)
}

// Test Prerequisites Not Met

func (suite *MagicLinkAuthExecutorTestSuite) TestExecute_PrerequisitesNotMet() {
	mockExec := coremock.NewExecutorInterfaceMock(suite.T())
	mockExec.On("GetName").Return(ExecutorNameMagicLinkAuth).Maybe()
	mockExec.On("GetType").Return(common.ExecutorTypeAuthentication).Maybe()
	mockExec.On("GetDefaultInputs").Return([]common.Input{MagicLinkTokenInput}).Maybe()
	mockExec.On("GetPrerequisites").Return([]common.Input{EmailInput}).Maybe()
	mockExec.On("ValidatePrerequisites", mock.Anything, mock.Anything).Return(false)
	suite.executor.ExecutorInterface = mockExec

	ctx := &core.NodeContext{
		Context:      context.Background(),
		FlowID:       magicLinkTestFlowID,
		FlowType:     common.FlowTypeAuthentication,
		ExecutorMode: ExecutorModeSend,
		UserInputs:   make(map[string]string),
		RuntimeData:  make(map[string]string),
	}

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
}

// Test Helper Methods

func (suite *MagicLinkAuthExecutorTestSuite) TestGetUserEmailFromContext_FromUserInputs() {
	ctx := &core.NodeContext{
		UserInputs: map[string]string{
			userAttributeEmail: magicLinkTestEmail,
		},
		RuntimeData: make(map[string]string),
	}

	email, err := suite.executor.getUserEmailFromContext(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), magicLinkTestEmail, email)
}

func (suite *MagicLinkAuthExecutorTestSuite) TestGetUserEmailFromContext_FromRuntimeData() {
	ctx := &core.NodeContext{
		UserInputs: make(map[string]string),
		RuntimeData: map[string]string{
			userAttributeEmail: magicLinkTestEmail,
		},
	}

	email, err := suite.executor.getUserEmailFromContext(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), magicLinkTestEmail, email)
}

func (suite *MagicLinkAuthExecutorTestSuite) TestGetUserEmailFromContext_NotFound() {
	ctx := &core.NodeContext{
		UserInputs:  make(map[string]string),
		RuntimeData: make(map[string]string),
	}

	email, err := suite.executor.getUserEmailFromContext(ctx)

	assert.Error(suite.T(), err)
	assert.Empty(suite.T(), email)
	assert.Contains(suite.T(), err.Error(), "email not found")
}

func (suite *MagicLinkAuthExecutorTestSuite) TestGetAuthenticatedUser_Success() {
	attrs := map[string]interface{}{
		"email": magicLinkTestEmail,
	}
	attrsJSON, _ := json.Marshal(attrs)

	execResp := &common.ExecutorResponse{
		RuntimeData: map[string]string{
			userAttributeUserID: magicLinkTestUserID,
		},
	}

	user := &userprovider.User{
		UserID:     magicLinkTestUserID,
		UserType:   magicLinkTestUserType,
		OUID:       magicLinkTestOUID,
		Attributes: attrsJSON,
	}

	suite.mockUserProvider.On("GetUser", magicLinkTestUserID).Return(user, nil)

	authenticatedUser, err := suite.executor.getAuthenticatedUser(execResp)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), authenticatedUser)
	assert.True(suite.T(), authenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), magicLinkTestUserID, authenticatedUser.UserID)
	assert.Equal(suite.T(), magicLinkTestUserType, authenticatedUser.UserType)
	assert.Equal(suite.T(), magicLinkTestOUID, authenticatedUser.OUID)
	suite.mockUserProvider.AssertExpectations(suite.T())
}

func (suite *MagicLinkAuthExecutorTestSuite) TestGetAuthenticatedUser_UserIDNotFound() {
	execResp := &common.ExecutorResponse{
		RuntimeData: make(map[string]string),
	}

	authenticatedUser, err := suite.executor.getAuthenticatedUser(execResp)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), authenticatedUser)
	assert.Contains(suite.T(), err.Error(), "user ID not found in runtime data")
}

func (suite *MagicLinkAuthExecutorTestSuite) TestGetAuthenticatedUser_GetUserError() {
	execResp := &common.ExecutorResponse{
		RuntimeData: map[string]string{
			userAttributeUserID: magicLinkTestUserID,
		},
	}

	suite.mockUserProvider.On("GetUser", magicLinkTestUserID).Return(nil,
		userprovider.NewUserProviderError(userprovider.ErrorCodeSystemError, "user not found", ""))

	authenticatedUser, err := suite.executor.getAuthenticatedUser(execResp)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), authenticatedUser)
	assert.Contains(suite.T(), err.Error(), "failed to get user")
	suite.mockUserProvider.AssertExpectations(suite.T())
}

// Test Property Getters

func (suite *MagicLinkAuthExecutorTestSuite) TestGetTokenExpiry_DefaultValue() {
	ctx := &core.NodeContext{
		NodeProperties: nil,
	}

	expiry := suite.executor.getTokenExpiry(ctx)

	assert.Equal(suite.T(), int64(magiclink.DefaultExpirySeconds), expiry)
}

func (suite *MagicLinkAuthExecutorTestSuite) TestGetTokenExpiry_CustomValue() {
	ctx := &core.NodeContext{
		NodeProperties: map[string]interface{}{
			propertyKeyTokenExpiry: "600",
		},
	}

	expiry := suite.executor.getTokenExpiry(ctx)

	assert.Equal(suite.T(), int64(600), expiry)
}

func (suite *MagicLinkAuthExecutorTestSuite) TestGetTokenExpiry_InvalidValue_UsesDefault() {
	ctx := &core.NodeContext{
		NodeProperties: map[string]interface{}{
			propertyKeyTokenExpiry: "invalid",
		},
	}

	expiry := suite.executor.getTokenExpiry(ctx)

	assert.Equal(suite.T(), int64(magiclink.DefaultExpirySeconds), expiry)
}

func (suite *MagicLinkAuthExecutorTestSuite) TestGetTokenExpiry_NegativeValue_UsesDefault() {
	ctx := &core.NodeContext{
		NodeProperties: map[string]interface{}{
			propertyKeyTokenExpiry: "-100",
		},
	}

	expiry := suite.executor.getTokenExpiry(ctx)

	assert.Equal(suite.T(), int64(magiclink.DefaultExpirySeconds), expiry)
}

func (suite *MagicLinkAuthExecutorTestSuite) TestGetTokenExpiry_EmptyString_UsesDefault() {
	ctx := &core.NodeContext{
		NodeProperties: map[string]interface{}{
			propertyKeyTokenExpiry: "",
		},
	}

	expiry := suite.executor.getTokenExpiry(ctx)

	assert.Equal(suite.T(), int64(magiclink.DefaultExpirySeconds), expiry)
}

func (suite *MagicLinkAuthExecutorTestSuite) TestGetTokenExpiry_NonStringValue_UsesDefault() {
	ctx := &core.NodeContext{
		NodeProperties: map[string]interface{}{
			propertyKeyTokenExpiry: 123,
		},
	}

	expiry := suite.executor.getTokenExpiry(ctx)

	assert.Equal(suite.T(), int64(magiclink.DefaultExpirySeconds), expiry)
}

func (suite *MagicLinkAuthExecutorTestSuite) TestGetMagicLinkURL_DefaultEmpty() {
	ctx := &core.NodeContext{
		NodeProperties: nil,
	}

	url := suite.executor.getMagicLinkURL(ctx)

	assert.Equal(suite.T(), "", url)
}

func (suite *MagicLinkAuthExecutorTestSuite) TestGetMagicLinkURL_CustomValue() {
	ctx := &core.NodeContext{
		NodeProperties: map[string]interface{}{
			propertyKeyMagicLinkURL: magicLinkTestMagicLinkURL,
		},
	}

	url := suite.executor.getMagicLinkURL(ctx)

	assert.Equal(suite.T(), magicLinkTestMagicLinkURL, url)
}

func (suite *MagicLinkAuthExecutorTestSuite) TestGetMagicLinkURL_NonStringValue_ReturnsEmpty() {
	ctx := &core.NodeContext{
		NodeProperties: map[string]interface{}{
			propertyKeyMagicLinkURL: 12345,
		},
	}

	url := suite.executor.getMagicLinkURL(ctx)

	assert.Equal(suite.T(), "", url)
}

// Test Edge Cases

func (suite *MagicLinkAuthExecutorTestSuite) TestExecute_SendMode_EmptyUserID() {
	ctx := &core.NodeContext{
		Context:      context.Background(),
		FlowID:       magicLinkTestFlowID,
		FlowType:     common.FlowTypeAuthentication,
		ExecutorMode: ExecutorModeSend,
		UserInputs: map[string]string{
			userAttributeEmail: magicLinkTestEmail,
		},
		RuntimeData: make(map[string]string),
	}

	suite.mockMagicLinkService.On("SendMagicLink", ctx.Context, magicLinkTestEmail,
		int64(magiclink.DefaultExpirySeconds), magicLinkTestFlowID, "").Return(nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	suite.mockUserProvider.AssertNotCalled(suite.T(), "IdentifyUser", map[string]interface{}{
		userAttributeEmail: magicLinkTestEmail,
	})
	suite.mockMagicLinkService.AssertExpectations(suite.T())
}

func (suite *MagicLinkAuthExecutorTestSuite) TestExecute_SendMode_IdentifyUserSystemError() {
	ctx := &core.NodeContext{
		Context:      context.Background(),
		FlowID:       magicLinkTestFlowID,
		FlowType:     common.FlowTypeAuthentication,
		ExecutorMode: ExecutorModeSend,
		UserInputs: map[string]string{
			userAttributeEmail: magicLinkTestEmail,
		},
		RuntimeData: make(map[string]string),
	}

	suite.mockMagicLinkService.On("SendMagicLink", ctx.Context, magicLinkTestEmail,
		int64(magiclink.DefaultExpirySeconds), magicLinkTestFlowID, "").Return(nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	suite.mockUserProvider.AssertNotCalled(suite.T(), "IdentifyUser", map[string]interface{}{
		userAttributeEmail: magicLinkTestEmail,
	})
	suite.mockMagicLinkService.AssertExpectations(suite.T())
}

func (suite *MagicLinkAuthExecutorTestSuite) TestExecute_SendMode_AuthenticatedUser_EmptyUserID() {
	ctx := &core.NodeContext{
		Context:      context.Background(),
		FlowID:       magicLinkTestFlowID,
		FlowType:     common.FlowTypeAuthentication,
		ExecutorMode: ExecutorModeSend,
		UserInputs: map[string]string{
			userAttributeEmail: magicLinkTestEmail,
		},
		RuntimeData: make(map[string]string),
		AuthenticatedUser: authncm.AuthenticatedUser{
			IsAuthenticated: true,
			UserID:          "",
		},
	}

	mockExec := coremock.NewExecutorInterfaceMock(suite.T())
	mockExec.On("GetName").Return(ExecutorNameMagicLinkAuth).Maybe()
	mockExec.On("GetType").Return(common.ExecutorTypeAuthentication).Maybe()
	mockExec.On("GetDefaultInputs").Return([]common.Input{MagicLinkTokenInput}).Maybe()
	mockExec.On("GetPrerequisites").Return([]common.Input{EmailInput}).Maybe()
	mockExec.On("ValidatePrerequisites", mock.Anything, mock.Anything).Return(true).Maybe()
	mockExec.On("GetUserIDFromContext", mock.Anything).Return("")
	suite.executor.ExecutorInterface = mockExec

	_, err := suite.executor.Execute(ctx)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "user ID is empty")
}

func (suite *MagicLinkAuthExecutorTestSuite) TestExecute_VerifyMode_EmptyToken() {
	ctx := &core.NodeContext{
		Context:      context.Background(),
		FlowID:       magicLinkTestFlowID,
		FlowType:     common.FlowTypeAuthentication,
		ExecutorMode: ExecutorModeVerify,
		UserInputs: map[string]string{
			userInputMagicLinkToken: "",
		},
		RuntimeData: make(map[string]string),
	}

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecUserInputRequired, resp.Status)
}

func (suite *MagicLinkAuthExecutorTestSuite) TestExecute_SendMode_RegistrationFlow_IdentifyUserSystemError() {
	ctx := &core.NodeContext{
		Context:      context.Background(),
		FlowID:       magicLinkTestFlowID,
		FlowType:     common.FlowTypeRegistration,
		ExecutorMode: ExecutorModeSend,
		UserInputs: map[string]string{
			userAttributeEmail: magicLinkTestEmail,
		},
		RuntimeData: make(map[string]string),
	}

	suite.mockUserProvider.On("IdentifyUser", map[string]interface{}{
		userAttributeEmail: magicLinkTestEmail,
	}).Return(nil, userprovider.NewUserProviderError(userprovider.ErrorCodeSystemError, "", ""))

	_, err := suite.executor.Execute(ctx)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to identify user during registration flow")
	suite.mockUserProvider.AssertExpectations(suite.T())
}
