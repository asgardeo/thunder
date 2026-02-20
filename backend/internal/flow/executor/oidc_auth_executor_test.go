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
	"github.com/asgardeo/thunder/tests/mocks/authn/oidcmock"
	"github.com/asgardeo/thunder/tests/mocks/flow/coremock"
	"github.com/asgardeo/thunder/tests/mocks/idp/idpmock"
	"github.com/asgardeo/thunder/tests/mocks/userschemamock"
)

type OIDCAuthExecutorTestSuite struct {
	suite.Suite
	mockOIDCService       *oidcmock.OIDCAuthnCoreServiceInterfaceMock
	mockIDPService        *idpmock.IDPServiceInterfaceMock
	mockUserSchemaService *userschemamock.UserSchemaServiceInterfaceMock
	mockFlowFactory       *coremock.FlowFactoryInterfaceMock
	executor              oidcAuthExecutorInterface
}

func TestOIDCAuthExecutorSuite(t *testing.T) {
	suite.Run(t, new(OIDCAuthExecutorTestSuite))
}

func (suite *OIDCAuthExecutorTestSuite) SetupTest() {
	suite.mockOIDCService = oidcmock.NewOIDCAuthnCoreServiceInterfaceMock(suite.T())
	suite.mockIDPService = idpmock.NewIDPServiceInterfaceMock(suite.T())
	suite.mockUserSchemaService = userschemamock.NewUserSchemaServiceInterfaceMock(suite.T())
	suite.mockFlowFactory = coremock.NewFlowFactoryInterfaceMock(suite.T())

	defaultInputs := []common.Input{{Identifier: "code", Type: "string", Required: true}}
	mockExec := createMockAuthExecutor(suite.T(), ExecutorNameOIDCAuth)
	suite.mockFlowFactory.On("CreateExecutor", ExecutorNameOIDCAuth, common.ExecutorTypeAuthentication,
		defaultInputs, []common.Input{}).Return(mockExec)

	suite.executor = newOIDCAuthExecutor(ExecutorNameOIDCAuth, defaultInputs, []common.Input{},
		suite.mockFlowFactory, suite.mockIDPService, suite.mockUserSchemaService, suite.mockOIDCService)
}

func (suite *OIDCAuthExecutorTestSuite) TestNewOIDCAuthExecutor() {
	assert.NotNil(suite.T(), suite.executor)
}

func (suite *OIDCAuthExecutorTestSuite) TestExecute_CodeNotProvided_BuildsAuthorizeURL() {
	ctx := &core.NodeContext{
		FlowID:     "flow-123",
		FlowType:   common.FlowTypeAuthentication,
		UserInputs: map[string]string{},
		NodeInputs: []common.Input{{Identifier: "code", Type: "string", Required: true}},
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	suite.mockOIDCService.On("BuildAuthorizeURL", "idp-123").
		Return("https://oidc.provider.com/authorize?client_id=abc&scope=openid", nil)

	suite.mockIDPService.On("GetIdentityProvider", "idp-123").
		Return(&idp.IDPDTO{ID: "idp-123", Name: "TestOIDCProvider"}, nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecExternalRedirection, resp.Status)
	assert.Contains(suite.T(), resp.RedirectURL, "https://oidc.provider.com/authorize")
	assert.Equal(suite.T(), "TestOIDCProvider", resp.AdditionalData[common.DataIDPName])
	suite.mockOIDCService.AssertExpectations(suite.T())
	suite.mockIDPService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestExecute_CodeProvided_ValidIDToken_AuthenticatesUser() {
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
		IDToken:     "id_token_jwt_123",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub":   "user-sub-123",
		"email": "test@example.com",
		"name":  "Test User",
		"iss":   "https://oidc.provider.com",
		"aud":   "client-id-123",
		"exp":   1234567890,
		"iat":   1234567800,
	}

	existingUser := &userprovider.User{
		UserID:             "user-123",
		OrganizationUnitID: "ou-123",
		UserType:           "INTERNAL",
	}

	oauthConfig := &authnoauth.OAuthClientConfig{
		Scopes: []string{"openid"},
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt_123").
		Return(idTokenClaims, nil)
	suite.mockOIDCService.On("GetInternalUser", "user-sub-123").
		Return(existingUser, nil)
	suite.mockOIDCService.On("GetOAuthClientConfig", "idp-123").
		Return(oauthConfig, nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), common.ExecComplete, resp.Status)
	assert.True(suite.T(), resp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), "user-123", resp.AuthenticatedUser.UserID)
	assert.Equal(suite.T(), "ou-123", resp.AuthenticatedUser.OrganizationUnitID)
	assert.Equal(suite.T(), "test@example.com", resp.RuntimeData["email"])
	suite.mockOIDCService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_ValidIDToken_Success() {
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
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub":   "user-sub-456",
		"email": "user@example.com",
		"iss":   "https://provider.com",
		"aud":   "client-id",
	}

	existingUser := &userprovider.User{
		UserID:             "user-456",
		OrganizationUnitID: "ou-456",
		UserType:           "INTERNAL",
	}

	oauthConfig := &authnoauth.OAuthClientConfig{
		Scopes: []string{"openid"},
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)
	suite.mockOIDCService.On("GetInternalUser", "user-sub-456").
		Return(existingUser, nil)
	suite.mockOIDCService.On("GetOAuthClientConfig", "idp-123").
		Return(oauthConfig, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecComplete, execResp.Status)
	assert.True(suite.T(), execResp.AuthenticatedUser.IsAuthenticated)
	suite.mockOIDCService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_InvalidNonce() {
	ctx := &core.NodeContext{
		FlowID:   "flow-123",
		FlowType: common.FlowTypeAuthentication,
		UserInputs: map[string]string{
			"code":  "auth_code_123",
			"nonce": "expected_nonce_123",
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
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub":   "user-sub-123",
		"nonce": "different_nonce_456",
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecFailure, execResp.Status)
	assert.Contains(suite.T(), execResp.FailureReason, "Nonce mismatch")
	suite.mockOIDCService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_NoSubClaim() {
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
		Scope:       "openid",
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"email": "test@example.com",
		"name":  "Test User",
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecFailure, execResp.Status)
	assert.Contains(suite.T(), execResp.FailureReason, "sub claim not found")
	suite.mockOIDCService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_RegistrationFlow_UserNotFound() {
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
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub":   "new-user-sub",
		"email": "newuser@example.com",
		"name":  "New User",
	}

	oauthConfig := &authnoauth.OAuthClientConfig{
		Scopes: []string{"openid"},
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)
	suite.mockOIDCService.On("GetInternalUser", "new-user-sub").
		Return(nil, &serviceerror.ServiceError{
			Code: authncm.ErrorUserNotFound.Code,
			Type: serviceerror.ClientErrorType,
		})
	suite.mockOIDCService.On("GetOAuthClientConfig", "idp-123").
		Return(oauthConfig, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecComplete, execResp.Status)
	assert.False(suite.T(), execResp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), "new-user-sub", execResp.RuntimeData["sub"])
	suite.mockOIDCService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_AuthFlow_UserNotFound() {
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
		Scope:       "openid",
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub": "unknown-user",
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)
	suite.mockOIDCService.On("GetInternalUser", "unknown-user").
		Return(nil, &serviceerror.ServiceError{
			Code: authncm.ErrorUserNotFound.Code,
			Type: serviceerror.ClientErrorType,
		})

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecFailure, execResp.Status)
	assert.Equal(suite.T(), failureReasonUserNotFound, execResp.FailureReason)
	suite.mockOIDCService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_UserAlreadyExists_RegistrationFlow() {
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
		Scope:       "openid",
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub": "existing-user-sub",
	}

	existingUser := &userprovider.User{
		UserID:             "user-789",
		OrganizationUnitID: "ou-789",
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)
	suite.mockOIDCService.On("GetInternalUser", "existing-user-sub").
		Return(existingUser, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecFailure, execResp.Status)
	assert.Contains(suite.T(), execResp.FailureReason, "User already exists")
	suite.mockOIDCService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestGetIDTokenClaims_Success() {
	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	expectedClaims := map[string]interface{}{
		"sub":   "user-sub-123",
		"email": "test@example.com",
		"name":  "Test User",
		"iss":   "https://provider.com",
		"aud":   "client-id",
	}

	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(expectedClaims, nil)

	claims, err := suite.executor.GetIDTokenClaims(execResp, "id_token_jwt")

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), claims)
	assert.Equal(suite.T(), "user-sub-123", claims["sub"])
	assert.Equal(suite.T(), "test@example.com", claims["email"])
	suite.mockOIDCService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestGetIDTokenClaims_Errors() {
	tests := []struct {
		name               string
		token              string
		serviceError       *serviceerror.ServiceError
		expectError        bool
		expectedStatus     common.ExecutorStatus
		expectedFailReason string
		errorContains      string
	}{
		{
			name:  "ClientError",
			token: "invalid_token",
			serviceError: &serviceerror.ServiceError{
				Type:             serviceerror.ClientErrorType,
				ErrorDescription: "Invalid ID token",
			},
			expectError:        false,
			expectedStatus:     common.ExecFailure,
			expectedFailReason: "Invalid ID token",
		},
		{
			name:  "ServerError",
			token: "id_token",
			serviceError: &serviceerror.ServiceError{
				Type:             serviceerror.ServerErrorType,
				Code:             "OIDC-5000",
				ErrorDescription: "Failed to extract claims",
			},
			expectError:   true,
			errorContains: "failed to extract claims from the ID token",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			execResp := &common.ExecutorResponse{
				AdditionalData: make(map[string]string),
				RuntimeData:    make(map[string]string),
			}

			suite.mockOIDCService.On("GetIDTokenClaims", tt.token).
				Return(nil, tt.serviceError).Once()

			claims, err := suite.executor.GetIDTokenClaims(execResp, tt.token)

			assert.Nil(suite.T(), claims)
			if tt.expectError {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), tt.errorContains)
			} else {
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), tt.expectedStatus, execResp.Status)
				assert.Equal(suite.T(), tt.expectedFailReason, execResp.FailureReason)
			}
			suite.mockOIDCService.AssertExpectations(suite.T())
		})
	}
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_WithAdditionalScopes_FetchesUserInfo() {
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
		Scope:       "openid profile email",
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub": "user-sub-123",
	}

	userInfo := map[string]interface{}{
		"sub":     "user-sub-123",
		"email":   "user@example.com",
		"phone":   "+1234567890",
		"address": "123 Main St",
	}

	existingUser := &userprovider.User{
		UserID:             "user-123",
		OrganizationUnitID: "ou-123",
		UserType:           "INTERNAL",
	}

	oauthConfig := &authnoauth.OAuthClientConfig{
		Scopes: []string{"openid", "profile", "email"},
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)
	suite.mockOIDCService.On("GetInternalUser", "user-sub-123").
		Return(existingUser, nil)
	suite.mockOIDCService.On("GetOAuthClientConfig", "idp-123").
		Return(oauthConfig, nil)
	suite.mockOIDCService.On("FetchUserInfo", "idp-123", "access_token_123").
		Return(userInfo, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecComplete, execResp.Status)
	assert.True(suite.T(), execResp.AuthenticatedUser.IsAuthenticated)
	assert.Contains(suite.T(), execResp.AuthenticatedUser.Attributes, "email")
	assert.Contains(suite.T(), execResp.AuthenticatedUser.Attributes, "phone")
	suite.mockOIDCService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_NoCodeProvided() {
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

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_FiltersNonUserClaimsFromIDToken() {
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
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub":     "user-sub-123",
		"email":   "user@example.com",
		"name":    "User Name",
		"iss":     "https://provider.com",
		"aud":     "client-id",
		"exp":     1234567890,
		"iat":     1234567800,
		"at_hash": "hash_value",
		"nonce":   "nonce_value",
	}

	existingUser := &userprovider.User{
		UserID:             "user-123",
		OrganizationUnitID: "ou-123",
		UserType:           "INTERNAL",
	}

	oauthConfig := &authnoauth.OAuthClientConfig{
		Scopes: []string{"openid"},
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)
	suite.mockOIDCService.On("GetInternalUser", "user-sub-123").
		Return(existingUser, nil)
	suite.mockOIDCService.On("GetOAuthClientConfig", "idp-123").
		Return(oauthConfig, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecComplete, execResp.Status)
	assert.Contains(suite.T(), execResp.AuthenticatedUser.Attributes, "email")
	assert.Contains(suite.T(), execResp.AuthenticatedUser.Attributes, "name")
	assert.NotContains(suite.T(), execResp.AuthenticatedUser.Attributes, "iss")
	assert.NotContains(suite.T(), execResp.AuthenticatedUser.Attributes, "aud")
	assert.NotContains(suite.T(), execResp.AuthenticatedUser.Attributes, "exp")
	assert.NotContains(suite.T(), execResp.AuthenticatedUser.Attributes, "iat")
	assert.NotContains(suite.T(), execResp.AuthenticatedUser.Attributes, "at_hash")
	assert.NotContains(suite.T(), execResp.AuthenticatedUser.Attributes, "nonce")
	assert.NotContains(suite.T(), execResp.AuthenticatedUser.Attributes, "sub")
	suite.mockOIDCService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_EmailInIDToken() {
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
		Scope:       "openid email",
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub":   "user-sub-789",
		"email": "user@test.com",
		"iss":   "https://provider.com",
		"aud":   "client-id",
	}

	existingUser := &userprovider.User{
		UserID:             "user-789",
		OrganizationUnitID: "ou-789",
		UserType:           "INTERNAL",
	}

	oauthConfig := &authnoauth.OAuthClientConfig{
		Scopes: []string{"openid"},
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)
	suite.mockOIDCService.On("GetInternalUser", "user-sub-789").
		Return(existingUser, nil)
	suite.mockOIDCService.On("GetOAuthClientConfig", "idp-123").
		Return(oauthConfig, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecComplete, execResp.Status)
	assert.True(suite.T(), execResp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), "user@test.com", execResp.RuntimeData["email"])
	assert.Equal(suite.T(), "user@test.com", execResp.AuthenticatedUser.Attributes["email"])
	suite.mockOIDCService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_NoEmailInIDToken() {
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
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub":  "user-sub-789",
		"name": "Test User",
		"iss":  "https://provider.com",
		"aud":  "client-id",
	}

	existingUser := &userprovider.User{
		UserID:             "user-789",
		OrganizationUnitID: "ou-789",
		UserType:           "INTERNAL",
	}

	oauthConfig := &authnoauth.OAuthClientConfig{
		Scopes: []string{"openid"},
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)
	suite.mockOIDCService.On("GetInternalUser", "user-sub-789").
		Return(existingUser, nil)
	suite.mockOIDCService.On("GetOAuthClientConfig", "idp-123").
		Return(oauthConfig, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecComplete, execResp.Status)
	assert.True(suite.T(), execResp.AuthenticatedUser.IsAuthenticated)
	assert.NotContains(suite.T(), execResp.RuntimeData, "email")
	assert.NotContains(suite.T(), execResp.AuthenticatedUser.Attributes, "email")
	suite.mockOIDCService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_EmptyEmailInIDToken() {
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
		Scope:       "openid email",
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub":   "user-sub-789",
		"email": "",
		"iss":   "https://provider.com",
		"aud":   "client-id",
	}

	existingUser := &userprovider.User{
		UserID:             "user-789",
		OrganizationUnitID: "ou-789",
		UserType:           "INTERNAL",
	}

	oauthConfig := &authnoauth.OAuthClientConfig{
		Scopes: []string{"openid"},
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)
	suite.mockOIDCService.On("GetInternalUser", "user-sub-789").
		Return(existingUser, nil)
	suite.mockOIDCService.On("GetOAuthClientConfig", "idp-123").
		Return(oauthConfig, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecComplete, execResp.Status)
	assert.True(suite.T(), execResp.AuthenticatedUser.IsAuthenticated)
	assert.NotContains(suite.T(), execResp.RuntimeData, "email")
	assert.Equal(suite.T(), "", execResp.AuthenticatedUser.Attributes["email"])
	suite.mockOIDCService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_RegistrationFlow_WithEmail() {
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
		Scope:       "openid email",
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub":   "new-user-sub",
		"email": "newuser@example.com",
		"name":  "New User",
		"iss":   "https://provider.com",
		"aud":   "client-id",
	}

	oauthConfig := &authnoauth.OAuthClientConfig{
		Scopes: []string{"openid"},
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)
	suite.mockOIDCService.On("GetInternalUser", "new-user-sub").
		Return(nil, &serviceerror.ServiceError{
			Code: authncm.ErrorUserNotFound.Code,
			Type: serviceerror.ClientErrorType,
		})
	suite.mockOIDCService.On("GetOAuthClientConfig", "idp-123").
		Return(oauthConfig, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecComplete, execResp.Status)
	assert.False(suite.T(), execResp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), "new-user-sub", execResp.RuntimeData["sub"])
	assert.Equal(suite.T(), "newuser@example.com", execResp.RuntimeData["email"])
	assert.Equal(suite.T(), "newuser@example.com", execResp.AuthenticatedUser.Attributes["email"])
	suite.mockOIDCService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_EmailFromUserInfo() {
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
		Scope:       "openid profile email",
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub":  "user-sub-789",
		"name": "Test User",
		"iss":  "https://provider.com",
		"aud":  "client-id",
	}

	userInfo := map[string]interface{}{
		"sub":   "user-sub-789",
		"email": "fromUserInfo@example.com",
		"name":  "Test User",
	}

	existingUser := &userprovider.User{
		UserID:             "user-789",
		OrganizationUnitID: "ou-789",
		UserType:           "INTERNAL",
	}

	oauthConfig := &authnoauth.OAuthClientConfig{
		Scopes: []string{"openid", "profile", "email"},
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)
	suite.mockOIDCService.On("GetInternalUser", "user-sub-789").
		Return(existingUser, nil)
	suite.mockOIDCService.On("GetOAuthClientConfig", "idp-123").
		Return(oauthConfig, nil)
	suite.mockOIDCService.On("FetchUserInfo", "idp-123", "access_token_123").
		Return(userInfo, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecComplete, execResp.Status)
	assert.True(suite.T(), execResp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), "fromUserInfo@example.com", execResp.RuntimeData["email"])
	assert.Equal(suite.T(), "fromUserInfo@example.com", execResp.AuthenticatedUser.Attributes["email"])
	suite.mockOIDCService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_EmailInIDToken_NilRuntimeData() {
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
		RuntimeData:    nil, // Explicitly nil
	}

	tokenResp := &authnoauth.TokenResponse{
		AccessToken: "access_token_123",
		TokenType:   "Bearer",
		Scope:       "openid email",
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub":   "user-sub-999",
		"email": "niltest@example.com",
		"iss":   "https://provider.com",
		"aud":   "client-id",
	}

	existingUser := &userprovider.User{
		UserID:             "user-999",
		OrganizationUnitID: "ou-999",
		UserType:           "INTERNAL",
	}

	oauthConfig := &authnoauth.OAuthClientConfig{
		Scopes: []string{"openid"},
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)
	suite.mockOIDCService.On("GetInternalUser", "user-sub-999").
		Return(existingUser, nil)
	suite.mockOIDCService.On("GetOAuthClientConfig", "idp-123").
		Return(oauthConfig, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecComplete, execResp.Status)
	assert.True(suite.T(), execResp.AuthenticatedUser.IsAuthenticated)
	assert.NotNil(suite.T(), execResp.RuntimeData, "RuntimeData should be initialized")
	assert.Equal(suite.T(), "niltest@example.com", execResp.RuntimeData["email"])
	assert.Equal(suite.T(), "niltest@example.com", execResp.AuthenticatedUser.Attributes["email"])
	suite.mockOIDCService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_AllowAuthWithoutLocalUser() {
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
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub":   "new-user-sub",
		"email": "newuser@example.com",
		"name":  "New User",
		"iss":   "https://provider.com",
		"aud":   "client-123",
		"exp":   float64(1234567890),
		"iat":   float64(1234567000),
	}

	oauthConfig := &authnoauth.OAuthClientConfig{
		Scopes: []string{"openid"},
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)
	suite.mockOIDCService.On("GetInternalUser", "new-user-sub").
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
	suite.mockOIDCService.On("GetOAuthClientConfig", "idp-123").
		Return(oauthConfig, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecComplete, execResp.Status)
	assert.False(suite.T(), execResp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), dataValueTrue, execResp.RuntimeData[common.RuntimeKeyUserEligibleForProvisioning])
	assert.Equal(suite.T(), "new-user-sub", execResp.RuntimeData["sub"])
	assert.NotNil(suite.T(), execResp.AuthenticatedUser.Attributes)
	suite.mockOIDCService.AssertExpectations(suite.T())
	suite.mockUserSchemaService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_PreventAuthWithoutLocalUser() {
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
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub": "new-user-sub",
		"iss": "https://provider.com",
		"aud": "client-123",
		"exp": float64(1234567890),
		"iat": float64(1234567000),
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)
	suite.mockOIDCService.On("GetInternalUser", "new-user-sub").
		Return(nil, &serviceerror.ServiceError{
			Code: authncm.ErrorUserNotFound.Code,
			Type: serviceerror.ClientErrorType,
		})

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecFailure, execResp.Status)
	assert.Equal(suite.T(), failureReasonUserNotFound, execResp.FailureReason)
	suite.mockOIDCService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_AllowRegistrationWithExistingUser() {
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
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub":   "existing-user-sub",
		"email": "existing@example.com",
		"name":  "Existing User",
		"iss":   "https://provider.com",
		"aud":   "client-123",
		"exp":   float64(1234567890),
		"iat":   float64(1234567000),
	}

	existingUser := &userprovider.User{
		UserID:             "user-123",
		OrganizationUnitID: "ou-123",
		UserType:           "INTERNAL",
	}

	oauthConfig := &authnoauth.OAuthClientConfig{
		Scopes: []string{"openid"},
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)
	suite.mockOIDCService.On("GetInternalUser", "existing-user-sub").
		Return(existingUser, nil)
	suite.mockOIDCService.On("GetOAuthClientConfig", "idp-123").
		Return(oauthConfig, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecComplete, execResp.Status)
	assert.True(suite.T(), execResp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), "user-123", execResp.AuthenticatedUser.UserID)
	assert.Equal(suite.T(), dataValueTrue, execResp.RuntimeData[common.RuntimeKeySkipProvisioning])
	suite.mockOIDCService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_PreventRegistrationWithExistingUser() {
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
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub": "existing-user-sub",
		"iss": "https://provider.com",
		"aud": "client-123",
		"exp": float64(1234567890),
		"iat": float64(1234567000),
	}

	existingUser := &userprovider.User{
		UserID:             "user-123",
		OrganizationUnitID: "ou-123",
		UserType:           "INTERNAL",
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)
	suite.mockOIDCService.On("GetInternalUser", "existing-user-sub").
		Return(existingUser, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecFailure, execResp.Status)
	assert.Equal(suite.T(), "User already exists with the provided sub claim.", execResp.FailureReason)
	suite.mockOIDCService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestGetContextUserAttributes_OAuthClientConfigErrors() {
	tests := []struct {
		name               string
		serviceError       *serviceerror.ServiceError
		idTokenClaims      map[string]interface{}
		expectGoError      bool
		expectedExecStatus common.ExecutorStatus
		errorContains      string
	}{
		{
			name: "ClientError",
			serviceError: &serviceerror.ServiceError{
				Code:             "CONFIG_ERROR",
				ErrorDescription: "Configuration not found",
				Type:             serviceerror.ClientErrorType,
			},
			idTokenClaims: map[string]interface{}{
				"sub":   "user-sub",
				"email": "user@example.com",
				"iss":   "https://provider.com",
				"aud":   "client-123",
			},
			expectGoError:      false,
			expectedExecStatus: common.ExecFailure,
			errorContains:      "failed to retrieve OAuth client configuration",
		},
		{
			name: "ServerError",
			serviceError: &serviceerror.ServiceError{
				Code:             "SERVER_ERROR",
				ErrorDescription: "Internal server error",
				Type:             serviceerror.ServerErrorType,
			},
			idTokenClaims: map[string]interface{}{
				"sub": "user-sub",
			},
			expectGoError:      true,
			expectedExecStatus: common.ExecutorStatus(""),
			errorContains:      "failed to retrieve OAuth client configuration",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Clear expectations before each test
			suite.mockOIDCService.ExpectedCalls = nil

			ctx := &core.NodeContext{
				FlowID: "flow-123",
				NodeProperties: map[string]interface{}{
					"idpId": "idp-123",
				},
			}

			execResp := &common.ExecutorResponse{
				AdditionalData: make(map[string]string),
				RuntimeData:    make(map[string]string),
			}

			suite.mockOIDCService.On("GetOAuthClientConfig", "idp-123").
				Return(nil, tt.serviceError)

			attributes, err := suite.executor.(*oidcAuthExecutor).getContextUserAttributes(
				ctx, execResp, tt.idTokenClaims, "access-token")

			if tt.expectGoError {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), tt.errorContains)
			} else {
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), tt.expectedExecStatus, execResp.Status)
				assert.Contains(suite.T(), execResp.FailureReason, tt.errorContains)
			}

			assert.Nil(suite.T(), attributes)
			suite.mockOIDCService.AssertExpectations(suite.T())
		})
	}
}

func (suite *OIDCAuthExecutorTestSuite) TestGetContextUserAttributes_OnlyOpenIDScope_NoUserInfoCall() {
	ctx := &core.NodeContext{
		FlowID: "flow-123",
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	idTokenClaims := map[string]interface{}{
		"sub":        "user-sub",
		"email":      "user@example.com",
		"name":       "Test User",
		"iss":        "https://provider.com",
		"aud":        "client-123",
		"exp":        float64(1234567890),
		"iat":        float64(1234567000),
		"given_name": "Test",
	}

	oauthConfig := &authnoauth.OAuthClientConfig{
		Scopes: []string{"openid"},
	}

	suite.mockOIDCService.On("GetOAuthClientConfig", "idp-123").
		Return(oauthConfig, nil)

	attributes, err := suite.executor.(*oidcAuthExecutor).getContextUserAttributes(
		ctx, execResp, idTokenClaims, "access-token")

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), attributes)
	// User attributes from ID token should be present
	assert.Equal(suite.T(), "user@example.com", attributes["email"])
	assert.Equal(suite.T(), "Test User", attributes["name"])
	assert.Equal(suite.T(), "Test", attributes["given_name"])
	// Non-user attributes should be filtered (including sub which is an identifier, not a user attribute)
	assert.NotContains(suite.T(), attributes, "sub")
	assert.NotContains(suite.T(), attributes, "iss")
	assert.NotContains(suite.T(), attributes, "aud")
	assert.NotContains(suite.T(), attributes, "exp")
	assert.NotContains(suite.T(), attributes, "iat")
	// Email should be added to runtime data
	assert.Equal(suite.T(), "user@example.com", execResp.RuntimeData["email"])
	// Verify GetUserInfo was NOT called since only openid scope is present
	suite.mockOIDCService.AssertNotCalled(suite.T(), "GetUserInfo")
	suite.mockOIDCService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_NonStringSubClaim() {
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
		Scope:       "openid",
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	// sub claim is not a string
	idTokenClaims := map[string]interface{}{
		"sub": 12345, // numeric sub instead of string
		"iss": "https://provider.com",
		"aud": "client-123",
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), common.ExecFailure, execResp.Status)
	assert.Equal(suite.T(), "sub claim not found in the ID token.", execResp.FailureReason)
	suite.mockOIDCService.AssertExpectations(suite.T())
}
