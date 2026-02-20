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
	authncm "github.com/asgardeo/thunder/internal/authn/common"
	authnoauth "github.com/asgardeo/thunder/internal/authn/oauth"
	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/idp"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/userprovider"
	"github.com/asgardeo/thunder/internal/userschema"
	"github.com/asgardeo/thunder/tests/mocks/authn/oauthmock"
	"github.com/asgardeo/thunder/tests/mocks/flow/coremock"
	"github.com/asgardeo/thunder/tests/mocks/idp/idpmock"
	"github.com/asgardeo/thunder/tests/mocks/userschemamock"
)

type OAuthExecutorTestSuite struct {
	suite.Suite
	mockOAuthService      *oauthmock.OAuthAuthnCoreServiceInterfaceMock
	mockIDPService        *idpmock.IDPServiceInterfaceMock
	mockUserSchemaService *userschemamock.UserSchemaServiceInterfaceMock
	mockFlowFactory       *coremock.FlowFactoryInterfaceMock
	executor              oAuthExecutorInterface
}

func TestOAuthExecutorSuite(t *testing.T) {
	suite.Run(t, new(OAuthExecutorTestSuite))
}

func (suite *OAuthExecutorTestSuite) SetupTest() {
	suite.mockOAuthService = oauthmock.NewOAuthAuthnCoreServiceInterfaceMock(suite.T())
	suite.mockIDPService = idpmock.NewIDPServiceInterfaceMock(suite.T())
	suite.mockUserSchemaService = userschemamock.NewUserSchemaServiceInterfaceMock(suite.T())
	suite.mockFlowFactory = coremock.NewFlowFactoryInterfaceMock(suite.T())

	defaultInputs := []common.Input{{Identifier: "code", Type: "string", Required: true}}
	mockExec := createMockAuthExecutor(suite.T(), ExecutorNameOAuth)
	suite.mockFlowFactory.On("CreateExecutor", ExecutorNameOAuth, common.ExecutorTypeAuthentication,
		defaultInputs, []common.Input{}).Return(mockExec)

	suite.executor = newOAuthExecutor(ExecutorNameOAuth, defaultInputs, []common.Input{},
		suite.mockFlowFactory, suite.mockIDPService, suite.mockUserSchemaService, suite.mockOAuthService)
}

func (suite *OAuthExecutorTestSuite) TestNewOAuthExecutor() {
	assert.NotNil(suite.T(), suite.executor)
}

func (suite *OAuthExecutorTestSuite) TestExecute_CodeNotProvided_BuildsAuthorizeURL() {
	ctx := &core.NodeContext{
		FlowID:     "flow-123",
		FlowType:   common.FlowTypeAuthentication,
		UserInputs: map[string]string{},
		NodeInputs: []common.Input{{Identifier: "code", Type: "string", Required: true}},
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	suite.mockOAuthService.On("BuildAuthorizeURL", "idp-123").
		Return("https://oauth.provider.com/authorize?client_id=abc", nil)

	suite.mockIDPService.On("GetIdentityProvider", "idp-123").
		Return(&idp.IDPDTO{ID: "idp-123", Name: "TestIDP"}, nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecExternalRedirection, resp.Status)
	assert.Equal(suite.T(), "https://oauth.provider.com/authorize?client_id=abc", resp.RedirectURL)
	assert.Equal(suite.T(), "TestIDP", resp.AdditionalData[common.DataIDPName])
	suite.mockOAuthService.AssertExpectations(suite.T())
	suite.mockIDPService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestExecute_CodeProvided_AuthenticatesUser() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			"code": "auth_code_123",
		},
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	tokenResp := &authnoauth.TokenResponse{
		AccessToken: "access_token_123",
		TokenType:   "Bearer",
		Scope:       "openid profile email",
		ExpiresIn:   3600,
	}

	userInfo := map[string]interface{}{
		"sub":   "user-sub-123",
		"email": "test@example.com",
		"name":  "Test User",
	}

	existingUser := &userprovider.User{
		UserID:             "user-123",
		OrganizationUnitID: "ou-123",
		UserType:           "INTERNAL",
	}

	suite.mockOAuthService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOAuthService.On("FetchUserInfo", "idp-123", "access_token_123").
		Return(userInfo, nil)
	suite.mockOAuthService.On("GetInternalUser", "user-sub-123").
		Return(existingUser, nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	assert.True(suite.T(), resp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), "user-123", resp.AuthenticatedUser.UserID)
	assert.Equal(suite.T(), "ou-123", resp.AuthenticatedUser.OrganizationUnitID)
	assert.Equal(suite.T(), "test@example.com", resp.RuntimeData["email"])
	suite.mockOAuthService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestBuildAuthorizeFlow_Success() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	suite.mockOAuthService.On("BuildAuthorizeURL", "idp-123").
		Return("https://oauth.provider.com/authorize", nil)
	suite.mockIDPService.On("GetIdentityProvider", "idp-123").
		Return(&idp.IDPDTO{ID: "idp-123", Name: "GoogleIDP"}, nil)

	err := suite.executor.BuildAuthorizeFlow(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecExternalRedirection, execResp.Status)
	assert.Equal(suite.T(), "https://oauth.provider.com/authorize", execResp.RedirectURL)
	assert.Equal(suite.T(), "GoogleIDP", execResp.AdditionalData[common.DataIDPName])
	suite.mockOAuthService.AssertExpectations(suite.T())
	suite.mockIDPService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestBuildAuthorizeFlow_IDPNotConfigured() {
	ctx := &core.NodeContext{
		FlowID:         "flow-123",
		FlowType:       common.FlowTypeAuthentication,
		NodeProperties: map[string]interface{}{},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	err := suite.executor.BuildAuthorizeFlow(ctx, execResp)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "idpId is not configured")
}

func (suite *OAuthExecutorTestSuite) TestBuildAuthorizeFlow_BuildURLClientError() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	suite.mockOAuthService.On("BuildAuthorizeURL", "idp-123").
		Return("", &serviceerror.ServiceError{
			Type:             serviceerror.ClientErrorType,
			ErrorDescription: "Invalid IDP configuration",
		})

	err := suite.executor.BuildAuthorizeFlow(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecFailure, execResp.Status)
	assert.Equal(suite.T(), "Invalid IDP configuration", execResp.FailureReason)
	suite.mockOAuthService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestBuildAuthorizeFlow_BuildURLServerError() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	suite.mockOAuthService.On("BuildAuthorizeURL", "idp-123").
		Return("", &serviceerror.ServiceError{
			Type:             serviceerror.ServerErrorType,
			Code:             "OAUTH-5000",
			ErrorDescription: "Internal server error",
		})

	err := suite.executor.BuildAuthorizeFlow(ctx, execResp)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to build authorize URL")
	suite.mockOAuthService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestExchangeCodeForToken_Success() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	tokenResp := &authnoauth.TokenResponse{
		AccessToken:  "access_token_123",
		TokenType:    "Bearer",
		Scope:        "openid profile",
		RefreshToken: "refresh_token_123",
		IDToken:      "id_token_123",
		ExpiresIn:    3600,
	}

	suite.mockOAuthService.On("ExchangeCodeForToken", "idp-123", "auth_code", true).
		Return(tokenResp, nil)

	result, err := suite.executor.ExchangeCodeForToken(ctx, execResp, "auth_code")

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "access_token_123", result.AccessToken)
	assert.Equal(suite.T(), "Bearer", result.TokenType)
	assert.Equal(suite.T(), "openid profile", result.Scope)
	assert.Equal(suite.T(), 3600, result.ExpiresIn)
	suite.mockOAuthService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestExchangeCodeForToken_ClientError() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	suite.mockOAuthService.On("ExchangeCodeForToken", "idp-123", "invalid_code", true).
		Return(nil, &serviceerror.ServiceError{
			Type:             serviceerror.ClientErrorType,
			ErrorDescription: "Invalid authorization code",
		})

	result, err := suite.executor.ExchangeCodeForToken(ctx, execResp, "invalid_code")

	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), common.ExecFailure, execResp.Status)
	assert.Equal(suite.T(), "Invalid authorization code", execResp.FailureReason)
	suite.mockOAuthService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestExchangeCodeForToken_ServerError() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	suite.mockOAuthService.On("ExchangeCodeForToken", "idp-123", "auth_code", true).
		Return(nil, &serviceerror.ServiceError{
			Type:             serviceerror.ServerErrorType,
			Code:             "OAUTH-5000",
			ErrorDescription: "Token exchange failed",
		})

	result, err := suite.executor.ExchangeCodeForToken(ctx, execResp, "auth_code")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "failed to exchange code for token")
	suite.mockOAuthService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestGetUserInfo_Success() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	userInfo := map[string]interface{}{
		"sub":   "user-sub-123",
		"email": "test@example.com",
		"name":  "Test User",
	}

	suite.mockOAuthService.On("FetchUserInfo", "idp-123", "access_token").
		Return(userInfo, nil)

	result, err := suite.executor.GetUserInfo(ctx, execResp, "access_token")

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "user-sub-123", result["sub"])
	assert.Equal(suite.T(), "test@example.com", result["email"])
	assert.Equal(suite.T(), "Test User", result["name"])
	suite.mockOAuthService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestGetUserInfo_ClientError() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	suite.mockOAuthService.On("FetchUserInfo", "idp-123", "invalid_token").
		Return(nil, &serviceerror.ServiceError{
			Type:             serviceerror.ClientErrorType,
			ErrorDescription: "Invalid access token",
		})

	result, err := suite.executor.GetUserInfo(ctx, execResp, "invalid_token")

	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), common.ExecFailure, execResp.Status)
	assert.Equal(suite.T(), "Invalid access token", execResp.FailureReason)
	suite.mockOAuthService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestGetUserInfo_ServerError() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	suite.mockOAuthService.On("FetchUserInfo", "idp-123", "access_token").
		Return(nil, &serviceerror.ServiceError{
			Type:             serviceerror.ServerErrorType,
			Code:             "OAUTH-5000",
			ErrorDescription: "Failed to fetch user info",
		})

	result, err := suite.executor.GetUserInfo(ctx, execResp, "access_token")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "failed to fetch user information")
	suite.mockOAuthService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestGetIdpID_Success() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	idpID, err := suite.executor.GetIdpID(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "idp-123", idpID)
}

func (suite *OAuthExecutorTestSuite) TestGetIdpID_NotConfigured() {
	ctx := &core.NodeContext{
		FlowID:         "flow-123",
		FlowType:       common.FlowTypeAuthentication,
		NodeProperties: map[string]interface{}{},
	}

	idpID, err := suite.executor.GetIdpID(ctx)

	assert.Error(suite.T(), err)
	assert.Empty(suite.T(), idpID)
	assert.Contains(suite.T(), err.Error(), "idpId is not configured")
}

func (suite *OAuthExecutorTestSuite) TestProcessAuthFlowResponse_RegistrationFlow_UserNotFound() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeRegistration,
		UserInputs: map[string]string{
			"code": "auth_code_123",
		},
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	tokenResp := &authnoauth.TokenResponse{
		AccessToken: "access_token_123",
		TokenType:   "Bearer",
		Scope:       "openid profile email",
		ExpiresIn:   3600,
	}

	userInfo := map[string]interface{}{
		"sub":   "new-user-sub",
		"email": "newuser@example.com",
		"name":  "New User",
	}

	suite.mockOAuthService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOAuthService.On("FetchUserInfo", "idp-123", "access_token_123").
		Return(userInfo, nil)
	suite.mockOAuthService.On("GetInternalUser", "new-user-sub").
		Return(nil, &serviceerror.ServiceError{
			Code: authncm.ErrorUserNotFound.Code,
			Type: serviceerror.ClientErrorType,
		})

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecComplete, execResp.Status)
	assert.False(suite.T(), execResp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), "new-user-sub", execResp.RuntimeData["sub"])
	assert.NotNil(suite.T(), execResp.AuthenticatedUser.Attributes)
	suite.mockOAuthService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestProcessAuthFlowResponse_AuthFlow_UserNotFound() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			"code": "auth_code_123",
		},
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	tokenResp := &authnoauth.TokenResponse{
		AccessToken: "access_token_123",
		TokenType:   "Bearer",
		Scope:       "openid profile",
		ExpiresIn:   3600,
	}

	userInfo := map[string]interface{}{
		"sub":   "unknown-user",
		"email": "unknown@example.com",
	}

	suite.mockOAuthService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOAuthService.On("FetchUserInfo", "idp-123", "access_token_123").
		Return(userInfo, nil)
	suite.mockOAuthService.On("GetInternalUser", "unknown-user").
		Return(nil, &serviceerror.ServiceError{
			Code: authncm.ErrorUserNotFound.Code,
			Type: serviceerror.ClientErrorType,
		})

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecFailure, execResp.Status)
	assert.Equal(suite.T(), "User not found", execResp.FailureReason)
	suite.mockOAuthService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestProcessAuthFlowResponse_UserAlreadyExists_RegistrationFlow() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeRegistration,
		UserInputs: map[string]string{
			"code": "auth_code_123",
		},
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	tokenResp := &authnoauth.TokenResponse{
		AccessToken: "access_token_123",
		TokenType:   "Bearer",
		Scope:       "openid profile",
		ExpiresIn:   3600,
	}

	userInfo := map[string]interface{}{
		"sub":   "existing-user-sub",
		"email": "existing@example.com",
	}

	existingUser := &userprovider.User{
		UserID:             "user-456",
		OrganizationUnitID: "ou-456",
	}

	suite.mockOAuthService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOAuthService.On("FetchUserInfo", "idp-123", "access_token_123").
		Return(userInfo, nil)
	suite.mockOAuthService.On("GetInternalUser", "existing-user-sub").
		Return(existingUser, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecFailure, execResp.Status)
	assert.Contains(suite.T(), execResp.FailureReason, "User already exists")
	suite.mockOAuthService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestProcessAuthFlowResponse_NoCodeProvided() {
	ctx := &core.NodeContext{
		FlowID:     "flow-123",
		FlowType:   common.FlowTypeAuthentication,
		UserInputs: map[string]string{},
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.False(suite.T(), execResp.AuthenticatedUser.IsAuthenticated)
}

func (suite *OAuthExecutorTestSuite) TestProcessAuthFlowResponse_EmptyScope() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			"code": "auth_code_123",
		},
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	tokenResp := &authnoauth.TokenResponse{
		AccessToken: "access_token_123",
		TokenType:   "Bearer",
		Scope:       "",
		ExpiresIn:   3600,
	}

	suite.mockOAuthService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecFailure, execResp.Status)
	assert.False(suite.T(), execResp.AuthenticatedUser.IsAuthenticated)
	suite.mockOAuthService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestProcessAuthFlowResponse_NoSubClaim() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			"code": "auth_code_123",
		},
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	tokenResp := &authnoauth.TokenResponse{
		AccessToken: "access_token_123",
		TokenType:   "Bearer",
		Scope:       "openid profile",
		ExpiresIn:   3600,
	}

	userInfo := map[string]interface{}{
		"email": "test@example.com",
		"name":  "Test User",
	}

	suite.mockOAuthService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOAuthService.On("FetchUserInfo", "idp-123", "access_token_123").
		Return(userInfo, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecFailure, execResp.Status)
	assert.Contains(suite.T(), execResp.FailureReason, "sub claim not found")
	suite.mockOAuthService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestHasRequiredInputs_CodeProvided() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			"code": "auth_code_123",
		},
	}

	execResp := &common.ExecutorResponse{
		Inputs: []common.Input{},
	}

	result := suite.executor.HasRequiredInputs(ctx, execResp)

	assert.True(suite.T(), result)
}

func (suite *OAuthExecutorTestSuite) TestHasRequiredInputs_CodeNotProvided() {
	ctx := &core.NodeContext{
		FlowID:     "flow-123",
		FlowType:   common.FlowTypeAuthentication,
		UserInputs: map[string]string{},
		NodeInputs: []common.Input{{Identifier: "code", Type: "string", Required: true}},
	}

	execResp := &common.ExecutorResponse{
		Inputs: []common.Input{},
	}

	result := suite.executor.HasRequiredInputs(ctx, execResp)

	assert.False(suite.T(), result)
	assert.NotEmpty(suite.T(), execResp.Inputs)
}

func (suite *OAuthExecutorTestSuite) TestGetContextUserAttributes_WithEmail() {
	userInfo := map[string]string{
		"sub":      "user-sub-123",
		"email":    "test@example.com",
		"name":     "Test User",
		"username": "testuser",
	}

	execResp := &common.ExecutorResponse{
		RuntimeData: make(map[string]string),
	}

	attributes := suite.executor.(*oAuthExecutor).getContextUserAttributes(execResp, userInfo)

	assert.NotNil(suite.T(), attributes)
	assert.Equal(suite.T(), "test@example.com", attributes["email"])
	assert.Equal(suite.T(), "Test User", attributes["name"])
	assert.NotContains(suite.T(), attributes, "sub")
	assert.NotContains(suite.T(), attributes, "username")
	assert.Equal(suite.T(), "test@example.com", execResp.RuntimeData["email"])
}

func (suite *OAuthExecutorTestSuite) TestGetContextUserAttributes_WithoutEmail() {
	userInfo := map[string]string{
		"sub":  "user-sub-123",
		"name": "Test User",
	}

	execResp := &common.ExecutorResponse{
		RuntimeData: make(map[string]string),
	}

	attributes := suite.executor.(*oAuthExecutor).getContextUserAttributes(execResp, userInfo)

	assert.NotNil(suite.T(), attributes)
	assert.Equal(suite.T(), "Test User", attributes["name"])
	assert.NotContains(suite.T(), attributes, "email")
	assert.NotContains(suite.T(), execResp.RuntimeData, "email")
}

func (suite *OAuthExecutorTestSuite) TestGetContextUserAttributes_WithEmptyEmail() {
	userInfo := map[string]string{
		"sub":   "user-sub-123",
		"email": "",
		"name":  "Test User",
	}

	execResp := &common.ExecutorResponse{
		RuntimeData: make(map[string]string),
	}

	attributes := suite.executor.(*oAuthExecutor).getContextUserAttributes(execResp, userInfo)

	assert.NotNil(suite.T(), attributes)
	assert.Equal(suite.T(), "", attributes["email"])
	assert.NotContains(suite.T(), execResp.RuntimeData, "email")
}

func (suite *OAuthExecutorTestSuite) TestGetContextUserAttributes_FilterSkipAttributes() {
	userInfo := map[string]string{
		"sub":      "user-sub-123",
		"email":    "test@example.com",
		"name":     "Test User",
		"username": "testuser",
		"id":       "some-id",
	}

	execResp := &common.ExecutorResponse{
		RuntimeData: make(map[string]string),
	}

	attributes := suite.executor.(*oAuthExecutor).getContextUserAttributes(execResp, userInfo)

	assert.NotNil(suite.T(), attributes)
	assert.Equal(suite.T(), "test@example.com", attributes["email"])
	assert.Equal(suite.T(), "Test User", attributes["name"])
	assert.NotContains(suite.T(), attributes, "sub")
	assert.NotContains(suite.T(), attributes, "username")
	assert.NotContains(suite.T(), attributes, "id")
	assert.Equal(suite.T(), "test@example.com", execResp.RuntimeData["email"])
}

func (suite *OAuthExecutorTestSuite) TestProcessAuthFlowResponse_RegistrationFlow_WithEmail() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeRegistration,
		UserInputs: map[string]string{
			"code": "auth_code_123",
		},
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	tokenResp := &authnoauth.TokenResponse{
		AccessToken: "access_token_123",
		TokenType:   "Bearer",
		Scope:       "openid profile email",
		ExpiresIn:   3600,
	}

	userInfo := map[string]interface{}{
		"sub":   "new-user-sub",
		"email": "newuser@example.com",
		"name":  "New User",
	}

	suite.mockOAuthService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOAuthService.On("FetchUserInfo", "idp-123", "access_token_123").
		Return(userInfo, nil)
	suite.mockOAuthService.On("GetInternalUser", "new-user-sub").
		Return(nil, &serviceerror.ServiceError{
			Code: authncm.ErrorUserNotFound.Code,
			Type: serviceerror.ClientErrorType,
		})

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecComplete, execResp.Status)
	assert.False(suite.T(), execResp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), "new-user-sub", execResp.RuntimeData["sub"])
	assert.Equal(suite.T(), "newuser@example.com", execResp.RuntimeData["email"])
	assert.Equal(suite.T(), "newuser@example.com", execResp.AuthenticatedUser.Attributes["email"])
	suite.mockOAuthService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestGetContextUserAttributes_WithEmail_NilRuntimeData() {
	userInfo := map[string]string{
		"sub":   "user-sub-123",
		"email": "test@example.com",
		"name":  "Test User",
	}

	execResp := &common.ExecutorResponse{
		RuntimeData: nil, // Explicitly nil
	}

	attributes := suite.executor.(*oAuthExecutor).getContextUserAttributes(execResp, userInfo)

	assert.NotNil(suite.T(), attributes)
	assert.Equal(suite.T(), "test@example.com", attributes["email"])
	assert.Equal(suite.T(), "Test User", attributes["name"])
	assert.NotNil(suite.T(), execResp.RuntimeData, "RuntimeData should be initialized")
	assert.Equal(suite.T(), "test@example.com", execResp.RuntimeData["email"])
}

func (suite *OAuthExecutorTestSuite) TestProcessAuthFlowResponse_AllowAuthWithoutLocalUser() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			"code": "auth_code_123",
		},
		NodeProperties: map[string]interface{}{
			"idpId":                               "idp-123",
			"allowAuthenticationWithoutLocalUser": true,
		},
		Application: appmodel.Application{
			AllowedUserTypes: []string{"INTERNAL"},
		},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	tokenResp := &authnoauth.TokenResponse{
		AccessToken: "access_token_123",
		TokenType:   "Bearer",
		Scope:       "openid profile",
		ExpiresIn:   3600,
	}

	userInfo := map[string]interface{}{
		"sub":   "new-user-sub",
		"email": "newuser@example.com",
		"name":  "New User",
	}

	suite.mockOAuthService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOAuthService.On("FetchUserInfo", "idp-123", "access_token_123").
		Return(userInfo, nil)
	suite.mockOAuthService.On("GetInternalUser", "new-user-sub").
		Return(nil, &serviceerror.ServiceError{
			Code: authncm.ErrorUserNotFound.Code,
			Type: serviceerror.ClientErrorType,
		})
	suite.mockUserSchemaService.On("GetUserSchemaByName", "INTERNAL").
		Return(&userschema.UserSchema{
			Name:                  "INTERNAL",
			AllowSelfRegistration: true,
			OrganizationUnitID:    "ou-123",
		}, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecComplete, execResp.Status)
	assert.False(suite.T(), execResp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), dataValueTrue, execResp.RuntimeData[common.RuntimeKeyUserEligibleForProvisioning])
	assert.Equal(suite.T(), "new-user-sub", execResp.RuntimeData["sub"])
	assert.NotNil(suite.T(), execResp.AuthenticatedUser.Attributes)
	suite.mockOAuthService.AssertExpectations(suite.T())
	suite.mockUserSchemaService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestProcessAuthFlowResponse_PreventAuthWithoutLocalUser() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			"code": "auth_code_123",
		},
		NodeProperties: map[string]interface{}{
			"idpId":                               "idp-123",
			"allowAuthenticationWithoutLocalUser": false,
		},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	tokenResp := &authnoauth.TokenResponse{
		AccessToken: "access_token_123",
		TokenType:   "Bearer",
		Scope:       "openid profile",
		ExpiresIn:   3600,
	}

	userInfo := map[string]interface{}{
		"sub":   "new-user-sub",
		"email": "newuser@example.com",
	}

	suite.mockOAuthService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOAuthService.On("FetchUserInfo", "idp-123", "access_token_123").
		Return(userInfo, nil)
	suite.mockOAuthService.On("GetInternalUser", "new-user-sub").
		Return(nil, &serviceerror.ServiceError{
			Code: authncm.ErrorUserNotFound.Code,
			Type: serviceerror.ClientErrorType,
		})

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecFailure, execResp.Status)
	assert.Equal(suite.T(), "User not found", execResp.FailureReason)
	suite.mockOAuthService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestProcessAuthFlowResponse_AllowRegistrationWithExistingUser() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeRegistration,
		UserInputs: map[string]string{
			"code": "auth_code_123",
		},
		NodeProperties: map[string]interface{}{
			"idpId":                             "idp-123",
			"allowRegistrationWithExistingUser": true,
		},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	tokenResp := &authnoauth.TokenResponse{
		AccessToken: "access_token_123",
		TokenType:   "Bearer",
		Scope:       "openid profile email",
		ExpiresIn:   3600,
	}

	userInfo := map[string]interface{}{
		"sub":   "existing-user-sub",
		"email": "existing@example.com",
		"name":  "Existing User",
	}

	existingUser := &userprovider.User{
		UserID:             "user-123",
		OrganizationUnitID: "ou-123",
		UserType:           "INTERNAL",
	}

	suite.mockOAuthService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOAuthService.On("FetchUserInfo", "idp-123", "access_token_123").
		Return(userInfo, nil)
	suite.mockOAuthService.On("GetInternalUser", "existing-user-sub").
		Return(existingUser, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecComplete, execResp.Status)
	assert.True(suite.T(), execResp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), "user-123", execResp.AuthenticatedUser.UserID)
	assert.Equal(suite.T(), dataValueTrue, execResp.RuntimeData[common.RuntimeKeySkipProvisioning])
	suite.mockOAuthService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestProcessAuthFlowResponse_PreventRegistrationWithExistingUser() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeRegistration,
		UserInputs: map[string]string{
			"code": "auth_code_123",
		},
		NodeProperties: map[string]interface{}{
			"idpId":                             "idp-123",
			"allowRegistrationWithExistingUser": false,
		},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	tokenResp := &authnoauth.TokenResponse{
		AccessToken: "access_token_123",
		TokenType:   "Bearer",
		Scope:       "openid profile",
		ExpiresIn:   3600,
	}

	userInfo := map[string]interface{}{
		"sub":   "existing-user-sub",
		"email": "existing@example.com",
	}

	existingUser := &userprovider.User{
		UserID:             "user-123",
		OrganizationUnitID: "ou-123",
		UserType:           "INTERNAL",
	}

	suite.mockOAuthService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOAuthService.On("FetchUserInfo", "idp-123", "access_token_123").
		Return(userInfo, nil)
	suite.mockOAuthService.On("GetInternalUser", "existing-user-sub").
		Return(existingUser, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecFailure, execResp.Status)
	assert.Equal(suite.T(), "User already exists with the provided sub claim.", execResp.FailureReason)
	suite.mockOAuthService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestResolveUserTypeForAutoProvisioning() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		Application: appmodel.Application{
			AllowedUserTypes: []string{"INTERNAL"},
		},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	suite.mockUserSchemaService.On("GetUserSchemaByName", "INTERNAL").
		Return(&userschema.UserSchema{
			Name:                  "INTERNAL",
			AllowSelfRegistration: true,
			OrganizationUnitID:    "ou-123",
		}, nil)

	err := suite.executor.(*oAuthExecutor).resolveUserTypeForAutoProvisioning(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.NotEqual(suite.T(), common.ExecFailure, execResp.Status)
	assert.Equal(suite.T(), "INTERNAL", execResp.RuntimeData[userTypeKey])
	assert.Equal(suite.T(), "ou-123", execResp.RuntimeData[defaultOUIDKey])
	suite.mockUserSchemaService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestResolveUserTypeForAutoProvisioning_Failures() {
	tests := []struct {
		name             string
		allowedUserTypes []string
		mockSetup        func()
	}{
		{
			name:             "NoAllowedUserTypes",
			allowedUserTypes: []string{},
			mockSetup:        func() {},
		},
		{
			name:             "NoSelfRegistrationEnabled",
			allowedUserTypes: []string{"INTERNAL"},
			mockSetup: func() {
				suite.mockUserSchemaService.On("GetUserSchemaByName", "INTERNAL").
					Return(&userschema.UserSchema{
						Name:                  "INTERNAL",
						AllowSelfRegistration: false,
						OrganizationUnitID:    "ou-123",
					}, nil).Once()
			},
		},
		{
			name:             "MultipleSelfRegistrationEnabled",
			allowedUserTypes: []string{"INTERNAL", "CUSTOMER"},
			mockSetup: func() {
				suite.mockUserSchemaService.On("GetUserSchemaByName", "INTERNAL").
					Return(&userschema.UserSchema{
						Name:                  "INTERNAL",
						AllowSelfRegistration: true,
						OrganizationUnitID:    "ou-123",
					}, nil).Once()
				suite.mockUserSchemaService.On("GetUserSchemaByName", "CUSTOMER").
					Return(&userschema.UserSchema{
						Name:                  "CUSTOMER",
						AllowSelfRegistration: true,
						OrganizationUnitID:    "ou-456",
					}, nil).Once()
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ctx := &core.NodeContext{
				FlowID:   "flow-123",
				FlowType: common.FlowTypeAuthentication,
				Application: appmodel.Application{
					AllowedUserTypes: tt.allowedUserTypes,
				},
			}

			execResp := &common.ExecutorResponse{
				AdditionalData: make(map[string]string),
				RuntimeData:    make(map[string]string),
			}

			tt.mockSetup()

			err := suite.executor.(*oAuthExecutor).resolveUserTypeForAutoProvisioning(ctx, execResp)

			assert.NoError(suite.T(), err)
			assert.Equal(suite.T(), common.ExecFailure, execResp.Status)
			assert.Equal(suite.T(), errCannotProvisionUserAutomatically, execResp.FailureReason)
			suite.mockUserSchemaService.AssertExpectations(suite.T())
		})
	}
}

func (suite *OAuthExecutorTestSuite) TestResolveUserTypeForAutoProvisioning_GetUserSchemaError() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		Application: appmodel.Application{
			AllowedUserTypes: []string{"INTERNAL"},
		},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	suite.mockUserSchemaService.On("GetUserSchemaByName", "INTERNAL").
		Return(nil, &serviceerror.ServiceError{
			Type:             serviceerror.ServerErrorType,
			Code:             "SCHEMA-5000",
			ErrorDescription: "Internal error",
		})

	err := suite.executor.(*oAuthExecutor).resolveUserTypeForAutoProvisioning(ctx, execResp)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "error while retrieving user schema")
	suite.mockUserSchemaService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestGetInternalUser_Errors() {
	tests := []struct {
		name               string
		sub                string
		serviceError       *serviceerror.ServiceError
		expectError        bool
		expectedStatus     common.ExecutorStatus
		expectedFailReason string
		errorContains      string
	}{
		{
			name: "UserNotFoundError",
			sub:  "unknown-sub",
			serviceError: &serviceerror.ServiceError{
				Code: authncm.ErrorUserNotFound.Code,
				Type: serviceerror.ClientErrorType,
			},
			expectError:    false,
			expectedStatus: "", // Status should not be set to failure for user not found
		},
		{
			name: "ClientError",
			sub:  "invalid-sub",
			serviceError: &serviceerror.ServiceError{
				Type:             serviceerror.ClientErrorType,
				ErrorDescription: "Invalid sub claim",
			},
			expectError:        false,
			expectedStatus:     common.ExecFailure,
			expectedFailReason: "Invalid sub claim",
		},
		{
			name: "ServerError",
			sub:  "some-sub",
			serviceError: &serviceerror.ServiceError{
				Type:             serviceerror.ServerErrorType,
				Code:             "USER-5000",
				ErrorDescription: "Internal error",
			},
			expectError:   true,
			errorContains: "error while retrieving internal user",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			execResp := &common.ExecutorResponse{
				AdditionalData: make(map[string]string),
				RuntimeData:    make(map[string]string),
			}

			suite.mockOAuthService.On("GetInternalUser", tt.sub).
				Return(nil, tt.serviceError).Once()

			result, err := suite.executor.GetInternalUser(tt.sub, execResp)

			assert.Nil(suite.T(), result)
			if tt.expectError {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), tt.errorContains)
			} else {
				assert.NoError(suite.T(), err)
				if tt.expectedStatus != "" {
					assert.Equal(suite.T(), tt.expectedStatus, execResp.Status)
					assert.Equal(suite.T(), tt.expectedFailReason, execResp.FailureReason)
				} else {
					assert.NotEqual(suite.T(), common.ExecFailure, execResp.Status)
				}
			}
			suite.mockOAuthService.AssertExpectations(suite.T())
		})
	}
}

func (suite *OAuthExecutorTestSuite) TestGetInternalUser_Success() {
	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	existingUser := &userprovider.User{
		UserID:             "user-123",
		OrganizationUnitID: "ou-123",
		UserType:           "INTERNAL",
	}

	suite.mockOAuthService.On("GetInternalUser", "user-sub-123").
		Return(existingUser, nil)

	result, err := suite.executor.GetInternalUser("user-sub-123", execResp)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "user-123", result.UserID)
	suite.mockOAuthService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestGetContextUserForRegistration_WithExistingUser_SkipProvisioningFlag() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeRegistration,
		RuntimeData: map[string]string{
			common.RuntimeKeySkipProvisioning: dataValueTrue,
		},
		NodeProperties: map[string]interface{}{
			"allowRegistrationWithExistingUser": true,
		},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	existingUser := &userprovider.User{
		UserID:             "user-456",
		OrganizationUnitID: "ou-456",
		UserType:           "INTERNAL",
	}

	contextUser, err := suite.executor.(*oAuthExecutor).getContextUserForRegistration(
		ctx, execResp, "test-sub", existingUser)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), contextUser)
	assert.True(suite.T(), contextUser.IsAuthenticated)
	assert.Equal(suite.T(), "user-456", contextUser.UserID)
	assert.Equal(suite.T(), dataValueTrue, execResp.RuntimeData[common.RuntimeKeySkipProvisioning])
	assert.Equal(suite.T(), common.ExecComplete, execResp.Status)
}

func (suite *OAuthExecutorTestSuite) TestResolveUserTypeForAutoProvisioning_FailureScenarios() {
	tests := []struct {
		name             string
		allowedUserTypes []string
		userSchemas      map[string]*userschema.UserSchema
	}{
		{
			name:             "NoAllowedUserTypes",
			allowedUserTypes: []string{},
			userSchemas:      nil,
		},
		{
			name:             "NoSelfRegistrationEnabled",
			allowedUserTypes: []string{"TYPE1", "TYPE2"},
			userSchemas: map[string]*userschema.UserSchema{
				"TYPE1": {
					Name:                  "TYPE1",
					AllowSelfRegistration: false,
				},
				"TYPE2": {
					Name:                  "TYPE2",
					AllowSelfRegistration: false,
				},
			},
		},
		{
			name:             "MultipleEligibleTypes",
			allowedUserTypes: []string{"TYPE1", "TYPE2"},
			userSchemas: map[string]*userschema.UserSchema{
				"TYPE1": {
					Name:                  "TYPE1",
					AllowSelfRegistration: true,
					OrganizationUnitID:    "ou-1",
				},
				"TYPE2": {
					Name:                  "TYPE2",
					AllowSelfRegistration: true,
					OrganizationUnitID:    "ou-2",
				},
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Clear expectations before each test
			suite.mockUserSchemaService.ExpectedCalls = nil

			ctx := &core.NodeContext{
				FlowID: "flow-123",
				Application: appmodel.Application{
					AllowedUserTypes: tt.allowedUserTypes,
				},
			}

			execResp := &common.ExecutorResponse{
				AdditionalData: make(map[string]string),
				RuntimeData:    make(map[string]string),
			}

			if tt.userSchemas != nil {
				for userType, schema := range tt.userSchemas {
					suite.mockUserSchemaService.On("GetUserSchemaByName", userType).Return(schema, nil)
				}
			}

			err := suite.executor.(*oAuthExecutor).resolveUserTypeForAutoProvisioning(ctx, execResp)

			assert.NoError(suite.T(), err)
			assert.Equal(suite.T(), common.ExecFailure, execResp.Status)
			assert.Equal(suite.T(), errCannotProvisionUserAutomatically, execResp.FailureReason)

			if tt.userSchemas != nil {
				suite.mockUserSchemaService.AssertExpectations(suite.T())
			}
		})
	}
}

func (suite *OAuthExecutorTestSuite) TestGetContextUserForAuthentication_WithoutLocalUser_NotAllowed() {
	ctx := &core.NodeContext{
		FlowID:         "flow-123",
		FlowType:       common.FlowTypeAuthentication,
		NodeProperties: map[string]interface{}{
			// allowAuthenticationWithoutLocalUser not set or false
		},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	contextUser, err := suite.executor.(*oAuthExecutor).getContextUserForAuthentication(
		ctx, execResp, "test-sub", nil)

	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), contextUser)
	assert.Equal(suite.T(), common.ExecFailure, execResp.Status)
	assert.Equal(suite.T(), "User not found", execResp.FailureReason)
}

func (suite *OAuthExecutorTestSuite) TestExecute_InvalidFlowType() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: "InvalidFlowType",
	}

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
}
