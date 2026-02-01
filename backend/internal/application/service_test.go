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

package application

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/cert"
	flowcommon "github.com/asgardeo/thunder/internal/flow/common"
	flowmgt "github.com/asgardeo/thunder/internal/flow/mgt"
	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/userschema"
	"github.com/asgardeo/thunder/tests/mocks/certmock"
	"github.com/asgardeo/thunder/tests/mocks/flow/flowmgtmock"
	"github.com/asgardeo/thunder/tests/mocks/userschemamock"
)

const testAppIDForRollback = "app123"

type ServiceTestSuite struct {
	suite.Suite
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}

func (suite *ServiceTestSuite) TestBuildBasicApplicationResponse() {
	app := model.BasicApplicationDTO{
		ID:                        "app-123",
		Name:                      "Test App",
		Description:               "Test Description",
		AuthFlowID:                "auth_flow_1",
		RegistrationFlowID:        "reg_flow_1",
		IsRegistrationFlowEnabled: true,
		ClientID:                  "client-123",
	}

	result := buildBasicApplicationResponse(app)

	assert.Equal(suite.T(), "app-123", result.ID)
	assert.Equal(suite.T(), "Test App", result.Name)
	assert.Equal(suite.T(), "Test Description", result.Description)
	assert.Equal(suite.T(), "auth_flow_1", result.AuthFlowID)
	assert.Equal(suite.T(), "reg_flow_1", result.RegistrationFlowID)
	assert.True(suite.T(), result.IsRegistrationFlowEnabled)
	assert.Equal(suite.T(), "client-123", result.ClientID)
}

func (suite *ServiceTestSuite) TestBuildBasicApplicationResponse_WithTemplate() {
	app := model.BasicApplicationDTO{
		ID:                        "app-123",
		Name:                      "Test App",
		Description:               "Test Description",
		AuthFlowID:                "auth_flow_1",
		RegistrationFlowID:        "reg_flow_1",
		IsRegistrationFlowEnabled: true,
		BrandingID:                "brand-123",
		Template:                  "spa",
		ClientID:                  "client-123",
		LogoURL:                   "https://example.com/logo.png",
	}

	result := buildBasicApplicationResponse(app)

	assert.Equal(suite.T(), "app-123", result.ID)
	assert.Equal(suite.T(), "Test App", result.Name)
	assert.Equal(suite.T(), "brand-123", result.BrandingID)
	assert.Equal(suite.T(), "spa", result.Template)
	assert.Equal(suite.T(), "client-123", result.ClientID)
	assert.Equal(suite.T(), "https://example.com/logo.png", result.LogoURL)
}

func (suite *ServiceTestSuite) TestBuildBasicApplicationResponse_WithEmptyTemplate() {
	app := model.BasicApplicationDTO{
		ID:                        "app-123",
		Name:                      "Test App",
		Description:               "Test Description",
		AuthFlowID:                "auth_flow_1",
		RegistrationFlowID:        "reg_flow_1",
		IsRegistrationFlowEnabled: true,
		Template:                  "",
		ClientID:                  "client-123",
	}

	result := buildBasicApplicationResponse(app)

	assert.Equal(suite.T(), "app-123", result.ID)
	assert.Equal(suite.T(), "", result.Template)
}

func (suite *ServiceTestSuite) TestGetDefaultTokenConfigFromDeployment() {
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "https://test-issuer.com",
			ValidityPeriod: 7200,
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetThunderRuntime()

	result := getDefaultTokenConfigFromDeployment()

	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "https://test-issuer.com", result.Issuer)
	assert.Equal(suite.T(), int64(7200), result.ValidityPeriod)
}

func (suite *ServiceTestSuite) TestProcessTokenConfiguration() {
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "https://default-issuer.com",
			ValidityPeriod: 3600,
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetThunderRuntime()

	tests := []struct {
		name                    string
		app                     *model.ApplicationDTO
		expectedRootIssuer      string
		expectedRootValidity    int64
		expectedAccessValidity  int64
		expectedIDTokenValidity int64
		expectedTokenIssuer     string
	}{
		{
			name: "No token config - uses defaults",
			app: &model.ApplicationDTO{
				Name: "Test App",
			},
			expectedRootIssuer:      "https://default-issuer.com",
			expectedRootValidity:    3600,
			expectedAccessValidity:  3600,
			expectedIDTokenValidity: 3600,
			expectedTokenIssuer:     "https://default-issuer.com",
		},
		{
			name: "Custom root token config",
			app: &model.ApplicationDTO{
				Name: "Test App",
				Token: &model.TokenConfig{
					Issuer:         "https://custom-issuer.com",
					ValidityPeriod: 7200,
					UserAttributes: []string{"email", "name"},
				},
			},
			expectedRootIssuer:      "https://custom-issuer.com",
			expectedRootValidity:    7200,
			expectedAccessValidity:  7200,
			expectedIDTokenValidity: 7200,
			expectedTokenIssuer:     "https://custom-issuer.com",
		},
		{
			name: "Partial root token config",
			app: &model.ApplicationDTO{
				Name: "Test App",
				Token: &model.TokenConfig{
					ValidityPeriod: 5000,
				},
			},
			expectedRootIssuer:      "https://default-issuer.com",
			expectedRootValidity:    5000,
			expectedAccessValidity:  5000,
			expectedIDTokenValidity: 5000,
			expectedTokenIssuer:     "https://default-issuer.com",
		},
		{
			name: "OAuth token config with custom issuer",
			app: &model.ApplicationDTO{
				Name: "Test App",
				InboundAuthConfig: []model.InboundAuthConfigDTO{
					{
						Type: model.OAuthInboundAuthType,
						OAuthAppConfig: &model.OAuthAppConfigDTO{
							Token: &model.OAuthTokenConfig{
								Issuer: "https://oauth-issuer.com",
								AccessToken: &model.AccessTokenConfig{
									ValidityPeriod: 1800,
								},
								IDToken: &model.IDTokenConfig{
									ValidityPeriod: 900,
								},
							},
						},
					},
				},
			},
			expectedRootIssuer:      "https://default-issuer.com",
			expectedRootValidity:    3600,
			expectedAccessValidity:  1800,
			expectedIDTokenValidity: 900,
			expectedTokenIssuer:     "https://oauth-issuer.com",
		},
		{
			name: "OAuth token with only access token config",
			app: &model.ApplicationDTO{
				Name: "Test App",
				InboundAuthConfig: []model.InboundAuthConfigDTO{
					{
						Type: model.OAuthInboundAuthType,
						OAuthAppConfig: &model.OAuthAppConfigDTO{
							Token: &model.OAuthTokenConfig{
								AccessToken: &model.AccessTokenConfig{
									ValidityPeriod: 2400,
									UserAttributes: []string{"sub"},
								},
							},
						},
					},
				},
			},
			expectedRootIssuer:      "https://default-issuer.com",
			expectedRootValidity:    3600,
			expectedAccessValidity:  2400,
			expectedIDTokenValidity: 3600,
			expectedTokenIssuer:     "https://default-issuer.com",
		},
		{
			name: "OAuth token with issuer but no root token",
			app: &model.ApplicationDTO{
				Name: "Test App",
				InboundAuthConfig: []model.InboundAuthConfigDTO{
					{
						Type: model.OAuthInboundAuthType,
						OAuthAppConfig: &model.OAuthAppConfigDTO{
							Token: &model.OAuthTokenConfig{
								Issuer: "https://oauth-only-issuer.com",
							},
						},
					},
				},
			},
			expectedRootIssuer:      "https://default-issuer.com",
			expectedRootValidity:    3600,
			expectedAccessValidity:  3600,
			expectedIDTokenValidity: 3600,
			expectedTokenIssuer:     "https://oauth-only-issuer.com",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			rootToken, accessToken, idToken, tokenIssuer := processTokenConfiguration(tt.app)

			assert.Equal(suite.T(), tt.expectedRootIssuer, rootToken.Issuer)
			assert.Equal(suite.T(), tt.expectedRootValidity, rootToken.ValidityPeriod)
			assert.NotNil(suite.T(), rootToken.UserAttributes)

			assert.Equal(suite.T(), tt.expectedAccessValidity, accessToken.ValidityPeriod)
			assert.NotNil(suite.T(), accessToken.UserAttributes)

			assert.Equal(suite.T(), tt.expectedIDTokenValidity, idToken.ValidityPeriod)
			assert.NotNil(suite.T(), idToken.UserAttributes)
			assert.NotNil(suite.T(), idToken.ScopeClaims)

			assert.Equal(suite.T(), tt.expectedTokenIssuer, tokenIssuer)
		})
	}
}

func (suite *ServiceTestSuite) TestValidateRedirectURIs() {
	tests := []struct {
		name        string
		oauthConfig *model.OAuthAppConfigDTO
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid redirect URIs",
			oauthConfig: &model.OAuthAppConfigDTO{
				RedirectURIs: []string{"https://example.com/callback", "https://example.com/callback2"},
			},
			expectError: false,
		},
		{
			name: "Empty redirect URIs with client credentials grant",
			oauthConfig: &model.OAuthAppConfigDTO{
				RedirectURIs: []string{},
				GrantTypes:   []oauth2const.GrantType{oauth2const.GrantTypeClientCredentials},
			},
			expectError: false,
		},
		{
			name: "Empty redirect URIs with authorization code grant",
			oauthConfig: &model.OAuthAppConfigDTO{
				RedirectURIs: []string{},
				GrantTypes:   []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
			},
			expectError: true,
			errorMsg:    "authorization_code grant type requires redirect URIs",
		},
		{
			name: "Redirect URI with fragment",
			oauthConfig: &model.OAuthAppConfigDTO{
				RedirectURIs: []string{"https://example.com/callback#fragment"},
			},
			expectError: true,
			errorMsg:    "Redirect URIs must not contain a fragment component",
		},
		{
			name: "Multiple redirect URIs with one having fragment",
			oauthConfig: &model.OAuthAppConfigDTO{
				RedirectURIs: []string{"https://example.com/callback", "https://example.com/callback2#fragment"},
			},
			expectError: true,
			errorMsg:    "Redirect URIs must not contain a fragment component",
		},
		{
			name: "Invalid redirect URI missing scheme",
			oauthConfig: &model.OAuthAppConfigDTO{
				RedirectURIs: []string{"example.com/callback"},
			},
			expectError: true,
		},
		{
			name: "Invalid redirect URI missing host",
			oauthConfig: &model.OAuthAppConfigDTO{
				RedirectURIs: []string{"https:///callback"},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := validateRedirectURIs(tt.oauthConfig)

			if tt.expectError {
				assert.NotNil(suite.T(), err)
				if tt.errorMsg != "" {
					assert.Contains(suite.T(), err.ErrorDescription, tt.errorMsg)
				}
			} else {
				assert.Nil(suite.T(), err)
			}
		})
	}
}

func (suite *ServiceTestSuite) TestValidateGrantTypesAndResponseTypes() {
	tests := []struct {
		name          string
		oauthConfig   *model.OAuthAppConfigDTO
		expectError   bool
		errorContains string
	}{
		{
			name: "Valid authorization code flow",
			oauthConfig: &model.OAuthAppConfigDTO{
				GrantTypes:    []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
				ResponseTypes: []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
			},
			expectError: false,
		},
		{
			name: "Valid implicit flow",
			oauthConfig: &model.OAuthAppConfigDTO{
				GrantTypes:    []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
				ResponseTypes: []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
			},
			expectError: false,
		},
		{
			name: "Valid client credentials",
			oauthConfig: &model.OAuthAppConfigDTO{
				GrantTypes:    []oauth2const.GrantType{oauth2const.GrantTypeClientCredentials},
				ResponseTypes: []oauth2const.ResponseType{},
			},
			expectError: false,
		},
		{
			name: "Authorization code without any response type",
			oauthConfig: &model.OAuthAppConfigDTO{
				GrantTypes:    []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
				ResponseTypes: []oauth2const.ResponseType{},
			},
			expectError:   true,
			errorContains: "authorization_code grant type requires 'code' response type",
		},
		{
			name: "Invalid grant type",
			oauthConfig: &model.OAuthAppConfigDTO{
				GrantTypes:    []oauth2const.GrantType{"invalid_grant"},
				ResponseTypes: []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
			},
			expectError: true,
		},
		{
			name: "Invalid response type",
			oauthConfig: &model.OAuthAppConfigDTO{
				GrantTypes:    []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
				ResponseTypes: []oauth2const.ResponseType{"invalid_response"},
			},
			expectError: true,
		},
		{
			name: "Client credentials with response types",
			oauthConfig: &model.OAuthAppConfigDTO{
				GrantTypes:    []oauth2const.GrantType{oauth2const.GrantTypeClientCredentials},
				ResponseTypes: []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := validateGrantTypesAndResponseTypes(tt.oauthConfig)

			if tt.expectError {
				assert.NotNil(suite.T(), err)
				if tt.errorContains != "" {
					assert.Contains(suite.T(), err.ErrorDescription, tt.errorContains)
				}
			} else {
				assert.Nil(suite.T(), err)
			}
		})
	}
}

func (suite *ServiceTestSuite) TestValidateTokenEndpointAuthMethod() {
	tests := []struct {
		name        string
		oauthConfig *model.OAuthAppConfigDTO
		expectError bool
	}{
		{
			name: "Valid client_secret_basic",
			oauthConfig: &model.OAuthAppConfigDTO{
				TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
				PublicClient:            false,
			},
			expectError: false,
		},
		{
			name: "Valid client_secret_post",
			oauthConfig: &model.OAuthAppConfigDTO{
				TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretPost,
				PublicClient:            false,
			},
			expectError: false,
		},
		{
			name: "Valid none for public client",
			oauthConfig: &model.OAuthAppConfigDTO{
				TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodNone,
				PublicClient:            true,
			},
			expectError: false,
		},
		{
			name: "Invalid none for client credentials grant",
			oauthConfig: &model.OAuthAppConfigDTO{
				TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodNone,
				GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeClientCredentials},
				PublicClient:            false,
			},
			expectError: true,
		},
		{
			name: "Invalid empty auth method",
			oauthConfig: &model.OAuthAppConfigDTO{
				TokenEndpointAuthMethod: "",
				PublicClient:            false,
			},
			expectError: true,
		},
		{
			name: "Invalid auth method value",
			oauthConfig: &model.OAuthAppConfigDTO{
				TokenEndpointAuthMethod: "invalid_method",
				PublicClient:            false,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := validateTokenEndpointAuthMethod(tt.oauthConfig)

			if tt.expectError {
				assert.NotNil(suite.T(), err)
			} else {
				assert.Nil(suite.T(), err)
			}
		})
	}
}

func (suite *ServiceTestSuite) TestValidatePublicClientConfiguration() {
	tests := []struct {
		name        string
		oauthConfig *model.OAuthAppConfigDTO
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid public client",
			oauthConfig: &model.OAuthAppConfigDTO{
				PublicClient:            true,
				ClientSecret:            "",
				TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodNone,
				PKCERequired:            true,
			},
			expectError: false,
		},
		{
			name: "Public client with client credentials grant",
			oauthConfig: &model.OAuthAppConfigDTO{
				PublicClient:            true,
				ClientSecret:            "",
				GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeClientCredentials},
				TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodNone,
				PKCERequired:            true,
			},
			expectError: true,
			errorMsg:    "Public clients cannot use the client_credentials grant type",
		},
		{
			name: "Public client with auth method other than none",
			oauthConfig: &model.OAuthAppConfigDTO{
				PublicClient:            true,
				ClientSecret:            "",
				TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
				PKCERequired:            true,
			},
			expectError: true,
			errorMsg:    "Public clients must use 'none' as token endpoint authentication method",
		},
		{
			name: "Public client with client secret",
			oauthConfig: &model.OAuthAppConfigDTO{
				PublicClient:            true,
				ClientSecret:            "should-not-have-secret",
				TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodNone,
			},
			expectError: true,
			errorMsg:    "Public clients cannot have client secrets",
		},
		{
			name: "Public client without PKCE required",
			oauthConfig: &model.OAuthAppConfigDTO{
				PublicClient:            true,
				ClientSecret:            "",
				TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodNone,
				PKCERequired:            false,
			},
			expectError: true,
			errorMsg:    "Public clients must have PKCE required set to true",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := validatePublicClientConfiguration(tt.oauthConfig)

			if tt.expectError {
				assert.NotNil(suite.T(), err)
				if tt.errorMsg != "" {
					assert.Contains(suite.T(), err.ErrorDescription, tt.errorMsg)
				}
			} else {
				assert.Nil(suite.T(), err)
			}
		})
	}
}

func (suite *ServiceTestSuite) TestGetProcessedClientSecret() {
	tests := []struct {
		name           string
		oauthConfig    *model.OAuthAppConfigDTO
		expectEmpty    bool
		expectNonEmpty bool
	}{
		{
			name: "Public client - no secret",
			oauthConfig: &model.OAuthAppConfigDTO{
				PublicClient: true,
				ClientSecret: "",
			},
			expectEmpty: true,
		},
		{
			name: "Confidential client with provided secret",
			oauthConfig: &model.OAuthAppConfigDTO{
				PublicClient: false,
				ClientSecret: "my-secret-123",
			},
			expectNonEmpty: true,
		},
		{
			name: "Confidential client without provided secret - generates new",
			oauthConfig: &model.OAuthAppConfigDTO{
				PublicClient: false,
				ClientSecret: "",
			},
			expectNonEmpty: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := getProcessedClientSecret(tt.oauthConfig)

			if tt.expectEmpty {
				assert.Empty(suite.T(), result)
			}
			if tt.expectNonEmpty {
				assert.NotEmpty(suite.T(), result)
			}
		})
	}
}

func (suite *ServiceTestSuite) TestValidateAuthFlowID_WithValidFlowID() {
	service, _, _, mockFlowMgtService := suite.setupTestService()

	app := &model.ApplicationDTO{
		AuthFlowID: "auth-flow-123",
	}

	mockFlowMgtService.EXPECT().IsValidFlow("auth-flow-123").Return(true)

	svcErr := service.validateAuthFlowID(app)

	assert.Nil(suite.T(), svcErr)
	assert.Equal(suite.T(), "auth-flow-123", app.AuthFlowID)
}

func (suite *ServiceTestSuite) TestValidateAuthFlowID_WithInvalidFlowID() {
	service, _, _, mockFlowMgtService := suite.setupTestService()

	app := &model.ApplicationDTO{
		AuthFlowID: "invalid-flow",
	}

	mockFlowMgtService.EXPECT().IsValidFlow("invalid-flow").Return(false)

	svcErr := service.validateAuthFlowID(app)

	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &ErrorInvalidAuthFlowID, svcErr)
}

func (suite *ServiceTestSuite) TestValidateAuthFlowID_WithEmptyFlowID_SetsDefault() {
	testConfig := &config.Config{
		Flow: config.FlowConfig{
			DefaultAuthFlowHandle: "default_auth_flow",
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetThunderRuntime()

	service, _, _, mockFlowMgtService := suite.setupTestService()

	app := &model.ApplicationDTO{
		AuthFlowID: "",
	}

	defaultFlow := &flowmgt.CompleteFlowDefinition{
		ID:     "default-flow-id-123",
		Handle: "default_auth_flow",
	}
	mockFlowMgtService.EXPECT().GetFlowByHandle("default_auth_flow", flowcommon.FlowTypeAuthentication).
		Return(defaultFlow, nil)

	svcErr := service.validateAuthFlowID(app)

	assert.Nil(suite.T(), svcErr)
	assert.Equal(suite.T(), "default-flow-id-123", app.AuthFlowID)
}

func (suite *ServiceTestSuite) TestValidateAuthFlowID_WithEmptyFlowID_ErrorRetrievingDefault() {
	testConfig := &config.Config{
		Flow: config.FlowConfig{
			DefaultAuthFlowHandle: "default_auth_flow",
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetThunderRuntime()

	service, _, _, mockFlowMgtService := suite.setupTestService()

	app := &model.ApplicationDTO{
		AuthFlowID: "",
	}

	mockFlowMgtService.EXPECT().GetFlowByHandle("default_auth_flow", flowcommon.FlowTypeAuthentication).
		Return(nil, &serviceerror.ServiceError{Type: serviceerror.ClientErrorType})

	svcErr := service.validateAuthFlowID(app)

	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &ErrorWhileRetrievingFlowDefinition, svcErr)
}

func (suite *ServiceTestSuite) TestValidateRegistrationFlowID_WithValidFlowID() {
	service, _, _, mockFlowMgtService := suite.setupTestService()

	app := &model.ApplicationDTO{
		RegistrationFlowID: "reg-flow-123",
	}

	mockFlowMgtService.EXPECT().IsValidFlow("reg-flow-123").Return(true)

	svcErr := service.validateRegistrationFlowID(app)

	assert.Nil(suite.T(), svcErr)
	assert.Equal(suite.T(), "reg-flow-123", app.RegistrationFlowID)
}

func (suite *ServiceTestSuite) TestValidateRegistrationFlowID_WithInvalidFlowID() {
	service, _, _, mockFlowMgtService := suite.setupTestService()

	app := &model.ApplicationDTO{
		RegistrationFlowID: "invalid-reg-flow",
	}

	mockFlowMgtService.EXPECT().IsValidFlow("invalid-reg-flow").Return(false)

	svcErr := service.validateRegistrationFlowID(app)

	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &ErrorInvalidRegistrationFlowID, svcErr)
}

func (suite *ServiceTestSuite) TestValidateRegistrationFlowID_WithEmptyFlowID_InfersFromAuthFlow() {
	service, _, _, mockFlowMgtService := suite.setupTestService()

	app := &model.ApplicationDTO{
		AuthFlowID:         "auth-flow-123",
		RegistrationFlowID: "",
	}

	authFlow := &flowmgt.CompleteFlowDefinition{
		ID:     "auth-flow-123",
		Handle: "basic_auth",
	}
	regFlow := &flowmgt.CompleteFlowDefinition{
		ID:     "reg-flow-456",
		Handle: "basic_auth",
	}

	mockFlowMgtService.EXPECT().GetFlow("auth-flow-123").Return(authFlow, nil)
	mockFlowMgtService.EXPECT().GetFlowByHandle("basic_auth", flowcommon.FlowTypeRegistration).
		Return(regFlow, nil)

	svcErr := service.validateRegistrationFlowID(app)

	assert.Nil(suite.T(), svcErr)
	assert.Equal(suite.T(), "reg-flow-456", app.RegistrationFlowID)
}

func (suite *ServiceTestSuite) TestValidateRegistrationFlowID_ErrorRetrievingAuthFlow() {
	service, _, _, mockFlowMgtService := suite.setupTestService()

	app := &model.ApplicationDTO{
		AuthFlowID:         "auth-flow-123",
		RegistrationFlowID: "",
	}

	mockFlowMgtService.EXPECT().GetFlow("auth-flow-123").
		Return(nil, &serviceerror.ServiceError{Type: serviceerror.ServerErrorType})

	svcErr := service.validateRegistrationFlowID(app)

	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &serviceerror.InternalServerError, svcErr)
}

func (suite *ServiceTestSuite) TestValidateRegistrationFlowID_ErrorRetrievingRegistrationFlow() {
	service, _, _, mockFlowMgtService := suite.setupTestService()

	app := &model.ApplicationDTO{
		AuthFlowID:         "auth-flow-123",
		RegistrationFlowID: "",
	}

	authFlow := &flowmgt.CompleteFlowDefinition{
		ID:     "auth-flow-123",
		Handle: "basic_auth",
	}

	mockFlowMgtService.EXPECT().GetFlow("auth-flow-123").Return(authFlow, nil)
	mockFlowMgtService.EXPECT().GetFlowByHandle("basic_auth", flowcommon.FlowTypeRegistration).
		Return(nil, &serviceerror.ServiceError{Type: serviceerror.ClientErrorType})

	svcErr := service.validateRegistrationFlowID(app)

	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &ErrorWhileRetrievingFlowDefinition, svcErr)
}

func (suite *ServiceTestSuite) TestValidateRegistrationFlowID_ClientErrorRetrievingAuthFlow() {
	service, _, _, mockFlowMgtService := suite.setupTestService()

	app := &model.ApplicationDTO{
		AuthFlowID:         "auth-flow-123",
		RegistrationFlowID: "",
	}

	mockFlowMgtService.EXPECT().GetFlow("auth-flow-123").
		Return(nil, &serviceerror.ServiceError{Type: serviceerror.ClientErrorType})

	svcErr := service.validateRegistrationFlowID(app)

	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &ErrorWhileRetrievingFlowDefinition, svcErr)
}

func (suite *ServiceTestSuite) TestValidateRegistrationFlowID_ServerErrorRetrievingRegistrationFlow() {
	service, _, _, mockFlowMgtService := suite.setupTestService()

	app := &model.ApplicationDTO{
		AuthFlowID:         "auth-flow-123",
		RegistrationFlowID: "",
	}

	authFlow := &flowmgt.CompleteFlowDefinition{
		ID:     "auth-flow-123",
		Handle: "basic_auth",
	}

	mockFlowMgtService.EXPECT().GetFlow("auth-flow-123").Return(authFlow, nil)
	mockFlowMgtService.EXPECT().GetFlowByHandle("basic_auth", flowcommon.FlowTypeRegistration).
		Return(nil, &serviceerror.ServiceError{Type: serviceerror.ServerErrorType})

	svcErr := service.validateRegistrationFlowID(app)

	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &serviceerror.InternalServerError, svcErr)
}

func (suite *ServiceTestSuite) TestGetDefaultAuthFlowID_Success() {
	testConfig := &config.Config{
		Flow: config.FlowConfig{
			DefaultAuthFlowHandle: "custom_auth_flow",
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetThunderRuntime()

	service, _, _, mockFlowMgtService := suite.setupTestService()

	defaultFlow := &flowmgt.CompleteFlowDefinition{
		ID:     "flow-id-789",
		Handle: "custom_auth_flow",
	}
	mockFlowMgtService.EXPECT().GetFlowByHandle("custom_auth_flow", flowcommon.FlowTypeAuthentication).
		Return(defaultFlow, nil)

	result, svcErr := service.getDefaultAuthFlowID()

	assert.Nil(suite.T(), svcErr)
	assert.Equal(suite.T(), "flow-id-789", result)
}

func (suite *ServiceTestSuite) TestGetDefaultAuthFlowID_ErrorRetrieving() {
	testConfig := &config.Config{
		Flow: config.FlowConfig{
			DefaultAuthFlowHandle: "custom_auth_flow",
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetThunderRuntime()

	service, _, _, mockFlowMgtService := suite.setupTestService()

	mockFlowMgtService.EXPECT().GetFlowByHandle("custom_auth_flow", flowcommon.FlowTypeAuthentication).
		Return(nil, &serviceerror.ServiceError{Type: serviceerror.ClientErrorType})

	result, svcErr := service.getDefaultAuthFlowID()

	assert.NotNil(suite.T(), svcErr)
	assert.Empty(suite.T(), result)
	assert.Equal(suite.T(), &ErrorWhileRetrievingFlowDefinition, svcErr)
}

func (suite *ServiceTestSuite) TestGetDefaultAuthFlowID_ServerError() {
	testConfig := &config.Config{
		Flow: config.FlowConfig{
			DefaultAuthFlowHandle: "custom_auth_flow",
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetThunderRuntime()

	service, _, _, mockFlowMgtService := suite.setupTestService()

	mockFlowMgtService.EXPECT().GetFlowByHandle("custom_auth_flow", flowcommon.FlowTypeAuthentication).
		Return(nil, &serviceerror.ServiceError{Type: serviceerror.ServerErrorType})

	result, svcErr := service.getDefaultAuthFlowID()

	assert.NotNil(suite.T(), svcErr)
	assert.Empty(suite.T(), result)
	assert.Equal(suite.T(), &serviceerror.InternalServerError, svcErr)
}

func (suite *ServiceTestSuite) setupTestService() (
	*applicationService,
	*applicationStoreInterfaceMock,
	*certmock.CertificateServiceInterfaceMock,
	*flowmgtmock.FlowMgtServiceInterfaceMock,
) {
	mockStore := newApplicationStoreInterfaceMock(suite.T())
	mockCertService := certmock.NewCertificateServiceInterfaceMock(suite.T())
	mockFlowMgtService := flowmgtmock.NewFlowMgtServiceInterfaceMock(suite.T())
	mockUserSchemaService := userschemamock.NewUserSchemaServiceInterfaceMock(suite.T())
	service := &applicationService{
		appStore:          mockStore,
		certService:       mockCertService,
		flowMgtService:    mockFlowMgtService,
		userSchemaService: mockUserSchemaService,
	}
	return service, mockStore, mockCertService, mockFlowMgtService
}

func (suite *ServiceTestSuite) TestGetOAuthApplication_EmptyClientID() {
	service, _, _, _ := suite.setupTestService()

	result, svcErr := service.GetOAuthApplication("")

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestGetOAuthApplication_NotFound() {
	service, mockStore, _, _ := suite.setupTestService()

	mockStore.On("GetOAuthApplication", "client123").Return(nil, model.ApplicationNotFoundError)

	result, svcErr := service.GetOAuthApplication("client123")

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestGetOAuthApplication_StoreError() {
	service, mockStore, _, _ := suite.setupTestService()

	mockStore.On("GetOAuthApplication", "client123").Return(nil, errors.New("store error"))

	result, svcErr := service.GetOAuthApplication("client123")

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestGetOAuthApplication_NilApp() {
	service, mockStore, _, _ := suite.setupTestService()

	mockStore.On("GetOAuthApplication", "client123").Return(nil, nil)

	result, svcErr := service.GetOAuthApplication("client123")

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestGetOAuthApplication_Success() {
	service, mockStore, _, _ := suite.setupTestService()

	oauthApp := &model.OAuthAppConfigProcessedDTO{
		AppID:    "app123",
		ClientID: "client123",
	}

	mockStore.On("GetOAuthApplication", "client123").Return(oauthApp, nil)

	result, svcErr := service.GetOAuthApplication("client123")

	assert.NotNil(suite.T(), result)
	assert.Nil(suite.T(), svcErr)
	assert.Equal(suite.T(), "client123", result.ClientID)
}

func (suite *ServiceTestSuite) TestGetApplication_EmptyAppID() {
	service, _, _, _ := suite.setupTestService()

	result, svcErr := service.GetApplication("")

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestGetApplication_NotFound() {
	service, mockStore, _, _ := suite.setupTestService()

	mockStore.On("GetApplicationByID", "app123").Return(nil, model.ApplicationNotFoundError)

	result, svcErr := service.GetApplication("app123")

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestGetApplication_StoreError() {
	service, mockStore, _, _ := suite.setupTestService()

	mockStore.On("GetApplicationByID", "app123").Return(nil, errors.New("store error"))

	result, svcErr := service.GetApplication("app123")

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestGetApplication_Success() {
	service, mockStore, mockCertService, _ := suite.setupTestService()

	app := &model.ApplicationProcessedDTO{
		ID:   "app123",
		Name: "Test App",
	}

	mockStore.On("GetApplicationByID", "app123").Return(app, nil)
	mockCertService.EXPECT().GetCertificateByReference(
		cert.CertificateReferenceTypeApplication, "app123").Return(nil, &cert.ErrorCertificateNotFound)

	result, svcErr := service.GetApplication("app123")

	assert.NotNil(suite.T(), result)
	assert.Nil(suite.T(), svcErr)
	assert.Equal(suite.T(), "app123", result.ID)
}

func (suite *ServiceTestSuite) TestGetApplicationList_Success() {
	service, mockStore, _, _ := suite.setupTestService()

	apps := []model.BasicApplicationDTO{
		{
			ID:   "app1",
			Name: "App 1",
		},
		{
			ID:   "app2",
			Name: "App 2",
		},
	}

	mockStore.On("GetTotalApplicationCount").Return(2, nil)
	mockStore.On("GetApplicationList").Return(apps, nil)

	result, svcErr := service.GetApplicationList()

	assert.NotNil(suite.T(), result)
	assert.Nil(suite.T(), svcErr)
	assert.Equal(suite.T(), 2, result.TotalResults)
	assert.Equal(suite.T(), 2, result.Count)
	assert.Len(suite.T(), result.Applications, 2)
}

func (suite *ServiceTestSuite) TestGetApplicationList_CountError() {
	service, mockStore, _, _ := suite.setupTestService()

	mockStore.On("GetTotalApplicationCount").Return(0, errors.New("count error"))

	result, svcErr := service.GetApplicationList()

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestGetApplicationList_ListError() {
	service, mockStore, _, _ := suite.setupTestService()

	mockStore.On("GetTotalApplicationCount").Return(2, nil)
	mockStore.On("GetApplicationList").Return(nil, errors.New("list error"))

	result, svcErr := service.GetApplicationList()

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestValidateApplication_NilApp() {
	service, _, _, _ := suite.setupTestService()

	result, inboundAuth, svcErr := service.ValidateApplication(nil)

	assert.Nil(suite.T(), result)
	assert.Nil(suite.T(), inboundAuth)
	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestValidateApplication_EmptyName() {
	service, _, _, _ := suite.setupTestService()

	app := &model.ApplicationDTO{
		Name: "",
	}

	result, inboundAuth, svcErr := service.ValidateApplication(app)

	assert.Nil(suite.T(), result)
	assert.Nil(suite.T(), inboundAuth)
	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestValidateApplication_ExistingName() {
	service, mockStore, _, _ := suite.setupTestService()

	app := &model.ApplicationDTO{
		Name: "Existing App",
	}

	existingApp := &model.ApplicationProcessedDTO{
		ID:   "existing-id",
		Name: "Existing App",
	}

	mockStore.On("GetApplicationByName", "Existing App").Return(existingApp, nil)

	result, inboundAuth, svcErr := service.ValidateApplication(app)

	assert.Nil(suite.T(), result)
	assert.Nil(suite.T(), inboundAuth)
	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestDeleteApplication_EmptyAppID() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetThunderRuntime()

	service, _, _, _ := suite.setupTestService()

	svcErr := service.DeleteApplication("")

	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestDeleteApplication_NotFound() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetThunderRuntime()

	service, mockStore, _, _ := suite.setupTestService()

	mockStore.On("DeleteApplication", "app123").Return(model.ApplicationNotFoundError)

	svcErr := service.DeleteApplication("app123")

	// Should return nil (not error) when app not found
	assert.Nil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestDeleteApplication_StoreError() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetThunderRuntime()

	service, mockStore, _, _ := suite.setupTestService()

	mockStore.On("DeleteApplication", "app123").Return(errors.New("store error"))

	svcErr := service.DeleteApplication("app123")

	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestDeleteApplication_Success() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetThunderRuntime()

	service, mockStore, mockCertService, _ := suite.setupTestService()

	mockStore.On("DeleteApplication", "app123").Return(nil)
	mockCertService.EXPECT().DeleteCertificateByReference(cert.CertificateReferenceTypeApplication,
		"app123").Return(nil)

	svcErr := service.DeleteApplication("app123")

	assert.Nil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestDeleteApplication_CertError() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetThunderRuntime()

	service, mockStore, mockCertService, _ := suite.setupTestService()

	mockStore.On("DeleteApplication", "app123").Return(nil)
	mockCertService.EXPECT().
		DeleteCertificateByReference(cert.CertificateReferenceTypeApplication, "app123").
		Return(&serviceerror.ServiceError{Type: serviceerror.ClientErrorType})

	svcErr := service.DeleteApplication("app123")

	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestGetApplicationCertificate_NotFound() {
	service, _, mockCertService, _ := suite.setupTestService()

	svcErr := &cert.ErrorCertificateNotFound

	mockCertService.EXPECT().
		GetCertificateByReference(cert.CertificateReferenceTypeApplication, "app123").
		Return(nil, svcErr)

	result, err := service.getApplicationCertificate("app123")

	assert.NotNil(suite.T(), result)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), cert.CertificateTypeNone, result.Type)
}

func (suite *ServiceTestSuite) TestGetApplicationCertificate_NilCertificate() {
	service, _, mockCertService, _ := suite.setupTestService()

	mockCertService.EXPECT().GetCertificateByReference(cert.CertificateReferenceTypeApplication,
		"app123").Return(nil, nil)

	result, err := service.getApplicationCertificate("app123")

	assert.NotNil(suite.T(), result)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), cert.CertificateTypeNone, result.Type)
}

func (suite *ServiceTestSuite) TestGetApplicationCertificate_Success() {
	service, _, mockCertService, _ := suite.setupTestService()

	certificate := &cert.Certificate{
		Type:  cert.CertificateTypeJWKS,
		Value: `{"keys":[]}`,
	}

	mockCertService.EXPECT().
		GetCertificateByReference(cert.CertificateReferenceTypeApplication, "app123").
		Return(certificate, nil)

	result, err := service.getApplicationCertificate("app123")

	assert.NotNil(suite.T(), result)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), cert.CertificateTypeJWKS, result.Type)
}

func (suite *ServiceTestSuite) TestCreateApplicationCertificate_Success() {
	service, _, mockCertService, _ := suite.setupTestService()

	certificate := &cert.Certificate{
		Type:  cert.CertificateTypeJWKS,
		Value: `{"keys":[]}`,
	}

	mockCertService.EXPECT().CreateCertificate(certificate).Return(certificate, nil)

	result, svcErr := service.createApplicationCertificate(certificate)

	assert.NotNil(suite.T(), result)
	assert.Nil(suite.T(), svcErr)
	assert.Equal(suite.T(), cert.CertificateTypeJWKS, result.Type)
}

func (suite *ServiceTestSuite) TestCreateApplicationCertificate_Nil() {
	service, _, _, _ := suite.setupTestService()

	result, svcErr := service.createApplicationCertificate(nil)

	assert.NotNil(suite.T(), result)
	assert.Nil(suite.T(), svcErr)
	assert.Equal(suite.T(), cert.CertificateTypeNone, result.Type)
}

func (suite *ServiceTestSuite) TestCreateApplicationCertificate_ClientError() {
	service, _, mockCertService, _ := suite.setupTestService()

	certificate := &cert.Certificate{
		Type:  cert.CertificateTypeJWKS,
		Value: `{"keys":[]}`,
	}

	svcErr := &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		ErrorDescription: "Invalid certificate",
	}

	mockCertService.EXPECT().CreateCertificate(certificate).Return(nil, svcErr)

	result, err := service.createApplicationCertificate(certificate)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
}

func (suite *ServiceTestSuite) TestRollbackAppCertificateCreation_Success() {
	service, _, mockCertService, _ := suite.setupTestService()

	mockCertService.EXPECT().DeleteCertificateByReference(cert.CertificateReferenceTypeApplication,
		"app123").Return(nil)

	svcErr := service.rollbackAppCertificateCreation("app123")

	assert.Nil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestRollbackAppCertificateCreation_ClientError() {
	service, _, mockCertService, _ := suite.setupTestService()

	svcErr := &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		ErrorDescription: "Certificate not found",
	}

	mockCertService.EXPECT().
		DeleteCertificateByReference(cert.CertificateReferenceTypeApplication, "app123").
		Return(svcErr)

	err := service.rollbackAppCertificateCreation("app123")

	assert.NotNil(suite.T(), err)
}

func (suite *ServiceTestSuite) TestGetValidatedCertificateForCreate_None() {
	service, _, _, _ := suite.setupTestService()

	app := &model.ApplicationDTO{
		ID: "app123",
		Certificate: &model.ApplicationCertificate{
			Type: "NONE",
		},
	}

	result, svcErr := service.getValidatedCertificateForCreate("app123", app)

	assert.Nil(suite.T(), result)
	assert.Nil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestGetValidatedCertificateForCreate_JWKS() {
	service, _, _, _ := suite.setupTestService()

	app := &model.ApplicationDTO{
		ID: "app123",
		Certificate: &model.ApplicationCertificate{
			Type:  "JWKS",
			Value: `{"keys":[]}`,
		},
	}

	result, svcErr := service.getValidatedCertificateForCreate("app123", app)

	assert.NotNil(suite.T(), result)
	assert.Nil(suite.T(), svcErr)
	assert.Equal(suite.T(), cert.CertificateTypeJWKS, result.Type)
}

func (suite *ServiceTestSuite) TestGetValidatedCertificateForCreate_JWKS_EmptyValue() {
	service, _, _, _ := suite.setupTestService()

	app := &model.ApplicationDTO{
		ID: "app123",
		Certificate: &model.ApplicationCertificate{
			Type:  "JWKS",
			Value: "",
		},
	}

	result, svcErr := service.getValidatedCertificateForCreate("app123", app)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestGetValidatedCertificateForCreate_JWKSUri() {
	service, _, _, _ := suite.setupTestService()

	app := &model.ApplicationDTO{
		ID: "app123",
		Certificate: &model.ApplicationCertificate{
			Type:  "JWKS_URI",
			Value: "https://example.com/jwks",
		},
	}

	result, svcErr := service.getValidatedCertificateForCreate("app123", app)

	assert.NotNil(suite.T(), result)
	assert.Nil(suite.T(), svcErr)
	assert.Equal(suite.T(), cert.CertificateTypeJWKSURI, result.Type)
}

func (suite *ServiceTestSuite) TestGetValidatedCertificateForCreate_InvalidType() {
	service, _, _, _ := suite.setupTestService()

	app := &model.ApplicationDTO{
		ID: "app123",
		Certificate: &model.ApplicationCertificate{
			Type:  "INVALID",
			Value: "some-value",
		},
	}

	result, svcErr := service.getValidatedCertificateForCreate("app123", app)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestValidateOAuthParamsForCreateAndUpdate_EmptyInboundAuth() {
	app := &model.ApplicationDTO{
		Name: "Test App",
	}

	result, svcErr := validateOAuthParamsForCreateAndUpdate(app)

	assert.Nil(suite.T(), result)
	assert.Nil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestValidateOAuthParamsForCreateAndUpdate_InvalidType() {
	app := &model.ApplicationDTO{
		Name: "Test App",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: "invalid_type",
			},
		},
	}

	result, svcErr := validateOAuthParamsForCreateAndUpdate(app)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestValidateOAuthParamsForCreateAndUpdate_NilOAuthConfig() {
	app := &model.ApplicationDTO{
		Name: "Test App",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type:           model.OAuthInboundAuthType,
				OAuthAppConfig: nil,
			},
		},
	}

	result, svcErr := validateOAuthParamsForCreateAndUpdate(app)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestValidateRedirectURIs_InvalidParsedURI() {
	oauthConfig := &model.OAuthAppConfigDTO{
		RedirectURIs: []string{"://invalid"},
	}

	err := validateRedirectURIs(oauthConfig)

	assert.NotNil(suite.T(), err)
}

func (suite *ServiceTestSuite) TestProcessTokenConfiguration_WithOAuthIDToken() {
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "https://default-issuer.com",
			ValidityPeriod: 3600,
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetThunderRuntime()

	app := &model.ApplicationDTO{
		Name: "Test App",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					Token: &model.OAuthTokenConfig{
						IDToken: &model.IDTokenConfig{
							ValidityPeriod: 1200,
							UserAttributes: []string{"email"},
							ScopeClaims:    map[string][]string{"scope1": {"claim1"}},
						},
					},
				},
			},
		},
	}

	rootToken, accessToken, idToken, tokenIssuer := processTokenConfiguration(app)

	assert.NotNil(suite.T(), rootToken)
	assert.NotNil(suite.T(), accessToken)
	assert.NotNil(suite.T(), idToken)
	assert.Equal(suite.T(), int64(1200), idToken.ValidityPeriod)
	assert.Equal(suite.T(), []string{"email"}, idToken.UserAttributes)
	assert.NotNil(suite.T(), idToken.ScopeClaims)
	assert.Equal(suite.T(), "https://default-issuer.com", tokenIssuer)
}

func (suite *ServiceTestSuite) TestGetApplicationCertificate_ClientError() {
	service, _, mockCertService, _ := suite.setupTestService()

	svcErr := &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		ErrorDescription: "Invalid certificate",
	}

	mockCertService.EXPECT().
		GetCertificateByReference(cert.CertificateReferenceTypeApplication, "app123").
		Return(nil, svcErr)

	result, err := service.getApplicationCertificate("app123")

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
}

func (suite *ServiceTestSuite) TestGetApplicationCertificate_ServerError() {
	service, _, mockCertService, _ := suite.setupTestService()

	svcErr := &serviceerror.ServiceError{
		Type: serviceerror.ServerErrorType,
	}

	mockCertService.EXPECT().
		GetCertificateByReference(cert.CertificateReferenceTypeApplication, "app123").
		Return(nil, svcErr)

	result, err := service.getApplicationCertificate("app123")

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
}

func (suite *ServiceTestSuite) TestRollbackAppCertificateCreation_ServerError() {
	service, _, mockCertService, _ := suite.setupTestService()

	svcErr := &serviceerror.ServiceError{
		Type: serviceerror.ServerErrorType,
	}

	mockCertService.EXPECT().
		DeleteCertificateByReference(cert.CertificateReferenceTypeApplication, "app123").
		Return(svcErr)

	err := service.rollbackAppCertificateCreation("app123")

	assert.NotNil(suite.T(), err)
}

func (suite *ServiceTestSuite) TestCreateApplicationCertificate_ServerError() {
	service, _, mockCertService, _ := suite.setupTestService()

	certificate := &cert.Certificate{
		Type:  cert.CertificateTypeJWKS,
		Value: `{"keys":[]}`,
	}

	svcErr := &serviceerror.ServiceError{
		Type: serviceerror.ServerErrorType,
	}

	mockCertService.EXPECT().CreateCertificate(certificate).Return(nil, svcErr)

	result, err := service.createApplicationCertificate(certificate)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
}

func (suite *ServiceTestSuite) TestGetValidatedCertificateForCreate_EmptyType() {
	service, _, _, _ := suite.setupTestService()

	app := &model.ApplicationDTO{
		ID: "app123",
		Certificate: &model.ApplicationCertificate{
			Type: "",
		},
	}

	result, svcErr := service.getValidatedCertificateForCreate("app123", app)

	assert.Nil(suite.T(), result)
	assert.Nil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestGetValidatedCertificateForCreate_NilCertificate() {
	service, _, _, _ := suite.setupTestService()

	app := &model.ApplicationDTO{
		ID:          "app123",
		Certificate: nil,
	}

	result, svcErr := service.getValidatedCertificateForCreate("app123", app)

	assert.Nil(suite.T(), result)
	assert.Nil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestGetValidatedCertificateForCreate_JWKSURI_InvalidURI() {
	service, _, _, _ := suite.setupTestService()

	app := &model.ApplicationDTO{
		ID: "app123",
		Certificate: &model.ApplicationCertificate{
			Type:  "JWKS_URI",
			Value: "not-a-valid-uri",
		},
	}

	result, svcErr := service.getValidatedCertificateForCreate("app123", app)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestDeleteApplicationCertificate_Success() {
	service, _, mockCertService, _ := suite.setupTestService()

	mockCertService.EXPECT().DeleteCertificateByReference(cert.CertificateReferenceTypeApplication,
		"app123").Return(nil)

	svcErr := service.deleteApplicationCertificate("app123")

	assert.Nil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestDeleteApplicationCertificate_ClientError() {
	service, _, mockCertService, _ := suite.setupTestService()

	svcErr := &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		ErrorDescription: "Certificate not found",
	}

	mockCertService.EXPECT().
		DeleteCertificateByReference(cert.CertificateReferenceTypeApplication, "app123").
		Return(svcErr)

	err := service.deleteApplicationCertificate("app123")

	assert.NotNil(suite.T(), err)
}

func (suite *ServiceTestSuite) TestDeleteApplicationCertificate_ServerError() {
	service, _, mockCertService, _ := suite.setupTestService()

	svcErr := &serviceerror.ServiceError{
		Type: serviceerror.ServerErrorType,
	}

	mockCertService.EXPECT().
		DeleteCertificateByReference(cert.CertificateReferenceTypeApplication, "app123").
		Return(svcErr)

	err := service.deleteApplicationCertificate("app123")

	assert.NotNil(suite.T(), err)
}

func (suite *ServiceTestSuite) TestGetApplicationCertificate_ClientError_NonNotFound() {
	service, _, mockCertService, _ := suite.setupTestService()

	svcErr := &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "CES-1001",
		ErrorDescription: "Invalid certificate",
	}

	mockCertService.EXPECT().
		GetCertificateByReference(cert.CertificateReferenceTypeApplication, "app123").
		Return(nil, svcErr)

	result, err := service.getApplicationCertificate("app123")

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
}

func (suite *ServiceTestSuite) TestValidateOAuthParamsForCreateAndUpdate_WithDefaults() {
	app := &model.ApplicationDTO{
		Name: "Test App",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{},
					ResponseTypes:           []oauth2const.ResponseType{},
					TokenEndpointAuthMethod: "",
				},
			},
		},
	}

	result, svcErr := validateOAuthParamsForCreateAndUpdate(app)

	assert.NotNil(suite.T(), result)
	assert.Nil(suite.T(), svcErr)
	assert.Len(suite.T(), result.OAuthAppConfig.GrantTypes, 1)
	assert.Equal(suite.T(), oauth2const.GrantTypeAuthorizationCode, result.OAuthAppConfig.GrantTypes[0])
	assert.Equal(
		suite.T(),
		oauth2const.TokenEndpointAuthMethodClientSecretBasic,
		result.OAuthAppConfig.TokenEndpointAuthMethod,
	)
}

func (suite *ServiceTestSuite) TestValidateOAuthParamsForCreateAndUpdate_WithResponseTypeDefault() {
	app := &model.ApplicationDTO{
		Name: "Test App",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
					ResponseTypes:           []oauth2const.ResponseType{},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
				},
			},
		},
	}

	result, svcErr := validateOAuthParamsForCreateAndUpdate(app)

	assert.NotNil(suite.T(), result)
	assert.Nil(suite.T(), svcErr)
	assert.Len(suite.T(), result.OAuthAppConfig.ResponseTypes, 1)
	assert.Equal(suite.T(), oauth2const.ResponseTypeCode, result.OAuthAppConfig.ResponseTypes[0])
}

func (suite *ServiceTestSuite) TestValidateOAuthParamsForCreateAndUpdate_WithGrantTypeButNoResponseType() {
	app := &model.ApplicationDTO{
		Name: "Test App",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeClientCredentials},
					ResponseTypes:           []oauth2const.ResponseType{},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
				},
			},
		},
	}

	result, svcErr := validateOAuthParamsForCreateAndUpdate(app)

	assert.NotNil(suite.T(), result)
	assert.Nil(suite.T(), svcErr)
	assert.Len(suite.T(), result.OAuthAppConfig.ResponseTypes, 0)
}

func (suite *ServiceTestSuite) TestGetValidatedCertificateInput_JWKS() {
	app := &model.ApplicationDTO{
		ID: "app123",
		Certificate: &model.ApplicationCertificate{
			Type:  "JWKS",
			Value: `{"keys":[]}`,
		},
	}

	result, svcErr := getValidatedCertificateInput("app123", "cert123", app)

	assert.NotNil(suite.T(), result)
	assert.Nil(suite.T(), svcErr)
	assert.Equal(suite.T(), cert.CertificateTypeJWKS, result.Type)
	assert.Equal(suite.T(), "cert123", result.ID)
}

func (suite *ServiceTestSuite) TestGetValidatedCertificateInput_JWKSURI() {
	app := &model.ApplicationDTO{
		ID: "app123",
		Certificate: &model.ApplicationCertificate{
			Type:  "JWKS_URI",
			Value: "https://example.com/jwks",
		},
	}

	result, svcErr := getValidatedCertificateInput("app123", "cert123", app)

	assert.NotNil(suite.T(), result)
	assert.Nil(suite.T(), svcErr)
	assert.Equal(suite.T(), cert.CertificateTypeJWKSURI, result.Type)
}

func (suite *ServiceTestSuite) TestGetValidatedCertificateInput_InvalidType() {
	app := &model.ApplicationDTO{
		ID: "app123",
		Certificate: &model.ApplicationCertificate{
			Type:  "INVALID",
			Value: "some-value",
		},
	}

	result, svcErr := getValidatedCertificateInput("app123", "cert123", app)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestGetValidatedCertificateInput_JWKSURI_InvalidURI() {
	app := &model.ApplicationDTO{
		ID: "app123",
		Certificate: &model.ApplicationCertificate{
			Type:  "JWKS_URI",
			Value: "not-a-valid-uri",
		},
	}

	result, svcErr := getValidatedCertificateInput("app123", "cert123", app)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestGetValidatedCertificateInput_JWKS_EmptyValue() {
	app := &model.ApplicationDTO{
		ID: "app123",
		Certificate: &model.ApplicationCertificate{
			Type:  "JWKS",
			Value: "",
		},
	}

	result, svcErr := getValidatedCertificateInput("app123", "cert123", app)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestDeleteApplication_DeclarativeResourcesEnabled() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: true,
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetThunderRuntime()

	service, _, _, _ := suite.setupTestService()

	svcErr := service.DeleteApplication("app123")

	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestEnrichApplicationWithCertificate_Error() {
	service, _, mockCertService, _ := suite.setupTestService()

	app := &model.Application{
		ID:   "app123",
		Name: "Test App",
	}

	svcErr := &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		ErrorDescription: "Invalid certificate",
	}

	mockCertService.EXPECT().
		GetCertificateByReference(cert.CertificateReferenceTypeApplication, "app123").
		Return(nil, svcErr)

	result, err := service.enrichApplicationWithCertificate(app)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
}

func (suite *ServiceTestSuite) TestEnrichApplicationWithCertificate_Success() {
	service, _, mockCertService, _ := suite.setupTestService()

	app := &model.Application{
		ID:   "app123",
		Name: "Test App",
	}

	certificate := &cert.Certificate{
		Type:  cert.CertificateTypeJWKS,
		Value: `{"keys":[]}`,
	}

	mockCertService.EXPECT().
		GetCertificateByReference(cert.CertificateReferenceTypeApplication, "app123").
		Return(certificate, nil)

	result, err := service.enrichApplicationWithCertificate(app)

	assert.NotNil(suite.T(), result)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), cert.CertificateTypeJWKS, result.Certificate.Type)
}

func (suite *ServiceTestSuite) TestProcessTokenConfiguration_WithRootToken() {
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "https://default-issuer.com",
			ValidityPeriod: 3600,
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetThunderRuntime()

	app := &model.ApplicationDTO{
		Name: "Test App",
		Token: &model.TokenConfig{
			Issuer:         "https://custom-issuer.com",
			ValidityPeriod: 1800,
			UserAttributes: []string{"email", "name"},
		},
	}

	rootToken, accessToken, idToken, tokenIssuer := processTokenConfiguration(app)

	assert.NotNil(suite.T(), rootToken)
	assert.NotNil(suite.T(), accessToken)
	assert.NotNil(suite.T(), idToken)
	assert.Equal(suite.T(), "https://custom-issuer.com", rootToken.Issuer)
	assert.Equal(suite.T(), int64(1800), rootToken.ValidityPeriod)
	assert.Equal(suite.T(), []string{"email", "name"}, rootToken.UserAttributes)
	assert.Equal(suite.T(), "https://custom-issuer.com", tokenIssuer)
}

func (suite *ServiceTestSuite) TestProcessTokenConfiguration_WithRootTokenDefaults() {
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "https://default-issuer.com",
			ValidityPeriod: 3600,
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetThunderRuntime()

	app := &model.ApplicationDTO{
		Name: "Test App",
		Token: &model.TokenConfig{
			Issuer:         "",
			ValidityPeriod: 0,
		},
	}

	rootToken, accessToken, idToken, tokenIssuer := processTokenConfiguration(app)

	assert.NotNil(suite.T(), rootToken)
	assert.NotNil(suite.T(), accessToken)
	assert.NotNil(suite.T(), idToken)
	assert.Equal(suite.T(), "https://default-issuer.com", rootToken.Issuer)
	assert.Equal(suite.T(), int64(3600), rootToken.ValidityPeriod)
	assert.Equal(suite.T(), "https://default-issuer.com", tokenIssuer)
}

func (suite *ServiceTestSuite) TestProcessTokenConfiguration_WithOAuthAccessToken() {
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "https://default-issuer.com",
			ValidityPeriod: 3600,
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetThunderRuntime()

	app := &model.ApplicationDTO{
		Name: "Test App",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					Token: &model.OAuthTokenConfig{
						AccessToken: &model.AccessTokenConfig{
							ValidityPeriod: 2400,
							UserAttributes: []string{"sub", "email"},
						},
					},
				},
			},
		},
	}

	rootToken, accessToken, idToken, tokenIssuer := processTokenConfiguration(app)

	assert.NotNil(suite.T(), rootToken)
	assert.NotNil(suite.T(), accessToken)
	assert.NotNil(suite.T(), idToken)
	assert.Equal(suite.T(), int64(2400), accessToken.ValidityPeriod)
	assert.Equal(suite.T(), []string{"sub", "email"}, accessToken.UserAttributes)
	assert.Equal(suite.T(), "https://default-issuer.com", tokenIssuer)
}

func (suite *ServiceTestSuite) TestProcessTokenConfiguration_WithOAuthAccessTokenDefaults() {
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "https://default-issuer.com",
			ValidityPeriod: 3600,
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetThunderRuntime()

	app := &model.ApplicationDTO{
		Name: "Test App",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					Token: &model.OAuthTokenConfig{
						AccessToken: &model.AccessTokenConfig{
							ValidityPeriod: 0,
							UserAttributes: nil,
						},
					},
				},
			},
		},
	}

	rootToken, accessToken, idToken, _ := processTokenConfiguration(app)

	assert.NotNil(suite.T(), rootToken)
	assert.NotNil(suite.T(), accessToken)
	assert.NotNil(suite.T(), idToken)
	assert.Equal(suite.T(), int64(3600), accessToken.ValidityPeriod)
	assert.NotNil(suite.T(), accessToken.UserAttributes)
	assert.Len(suite.T(), accessToken.UserAttributes, 0)
}

func (suite *ServiceTestSuite) TestProcessTokenConfiguration_WithOAuthIDTokenDefaults() {
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "https://default-issuer.com",
			ValidityPeriod: 3600,
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetThunderRuntime()

	app := &model.ApplicationDTO{
		Name: "Test App",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					Token: &model.OAuthTokenConfig{
						IDToken: &model.IDTokenConfig{
							ValidityPeriod: 0,
							UserAttributes: nil,
							ScopeClaims:    nil,
						},
					},
				},
			},
		},
	}

	rootToken, accessToken, idToken, _ := processTokenConfiguration(app)

	assert.NotNil(suite.T(), rootToken)
	assert.NotNil(suite.T(), accessToken)
	assert.NotNil(suite.T(), idToken)
	assert.Equal(suite.T(), int64(3600), idToken.ValidityPeriod)
	assert.NotNil(suite.T(), idToken.UserAttributes)
	assert.Len(suite.T(), idToken.UserAttributes, 0)
	assert.NotNil(suite.T(), idToken.ScopeClaims)
}

func (suite *ServiceTestSuite) TestProcessTokenConfiguration_WithOAuthTokenIssuer() {
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "https://default-issuer.com",
			ValidityPeriod: 3600,
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetThunderRuntime()

	app := &model.ApplicationDTO{
		Name: "Test App",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					Token: &model.OAuthTokenConfig{
						Issuer: "https://oauth-issuer.com",
					},
				},
			},
		},
	}

	rootToken, accessToken, idToken, tokenIssuer := processTokenConfiguration(app)

	assert.NotNil(suite.T(), rootToken)
	assert.NotNil(suite.T(), accessToken)
	assert.NotNil(suite.T(), idToken)
	assert.Equal(suite.T(), "https://oauth-issuer.com", tokenIssuer)
}

func (suite *ServiceTestSuite) TestProcessTokenConfiguration_WithAccessTokenNilUserAttributes() {
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "https://default-issuer.com",
			ValidityPeriod: 3600,
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetThunderRuntime()

	app := &model.ApplicationDTO{
		Name: "Test App",
		Token: &model.TokenConfig{
			Issuer:         "https://root-issuer.com",
			ValidityPeriod: 1800,
			UserAttributes: []string{"email", "name"},
		},
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					Token: &model.OAuthTokenConfig{
						AccessToken: &model.AccessTokenConfig{
							ValidityPeriod: 2400,
							UserAttributes: nil, // nil UserAttributes
						},
					},
				},
			},
		},
	}

	rootToken, accessToken, idToken, tokenIssuer := processTokenConfiguration(app)

	assert.NotNil(suite.T(), rootToken)
	assert.NotNil(suite.T(), accessToken)
	assert.NotNil(suite.T(), idToken)
	// nil UserAttributes should be initialized to empty slice
	assert.NotNil(suite.T(), accessToken.UserAttributes)
	assert.Len(suite.T(), accessToken.UserAttributes, 0)
	assert.Equal(suite.T(), int64(2400), accessToken.ValidityPeriod)
	assert.Equal(suite.T(), "https://root-issuer.com", tokenIssuer)
}

func (suite *ServiceTestSuite) TestProcessTokenConfiguration_WithAccessTokenEmptyUserAttributes() {
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "https://default-issuer.com",
			ValidityPeriod: 3600,
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetThunderRuntime()

	app := &model.ApplicationDTO{
		Name: "Test App",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					Token: &model.OAuthTokenConfig{
						AccessToken: &model.AccessTokenConfig{
							ValidityPeriod: 2400,
							UserAttributes: []string{}, // empty slice
						},
					},
				},
			},
		},
	}

	rootToken, accessToken, idToken, tokenIssuer := processTokenConfiguration(app)

	assert.NotNil(suite.T(), rootToken)
	assert.NotNil(suite.T(), accessToken)
	assert.NotNil(suite.T(), idToken)
	assert.NotNil(suite.T(), accessToken.UserAttributes)
	assert.Len(suite.T(), accessToken.UserAttributes, 0)
	assert.Equal(suite.T(), int64(2400), accessToken.ValidityPeriod)
	assert.Equal(suite.T(), "https://default-issuer.com", tokenIssuer)
}

func (suite *ServiceTestSuite) TestValidateOAuthParamsForCreateAndUpdate_RedirectURIError() {
	app := &model.ApplicationDTO{
		Name: "Test App",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					RedirectURIs:            []string{"://invalid"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
				},
			},
		},
	}

	result, svcErr := validateOAuthParamsForCreateAndUpdate(app)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestValidateOAuthParamsForCreateAndUpdate_GrantTypeError() {
	app := &model.ApplicationDTO{
		Name: "Test App",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeClientCredentials},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
				},
			},
		},
	}

	result, svcErr := validateOAuthParamsForCreateAndUpdate(app)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestValidateOAuthParamsForCreateAndUpdate_TokenEndpointAuthMethodError() {
	app := &model.ApplicationDTO{
		Name: "Test App",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretPost,
					PublicClient:            true,
					PKCERequired:            true,
				},
			},
		},
	}

	result, svcErr := validateOAuthParamsForCreateAndUpdate(app)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestValidateOAuthParamsForCreateAndUpdate_PublicClientError() {
	app := &model.ApplicationDTO{
		Name: "Test App",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeClientCredentials},
					ResponseTypes:           []oauth2const.ResponseType{},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodNone,
					PublicClient:            true,
					PKCERequired:            true,
					ClientSecret:            "secret",
				},
			},
		},
	}

	result, svcErr := validateOAuthParamsForCreateAndUpdate(app)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestValidateOAuthParamsForCreateAndUpdate_PublicClientSuccess() {
	app := &model.ApplicationDTO{
		Name: "Test App",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodNone,
					PublicClient:            true,
					PKCERequired:            true,
				},
			},
		},
	}

	result, svcErr := validateOAuthParamsForCreateAndUpdate(app)

	assert.NotNil(suite.T(), result)
	assert.Nil(suite.T(), svcErr)
	assert.True(suite.T(), result.OAuthAppConfig.PublicClient)
}

func (suite *ServiceTestSuite) TestValidateApplication_StoreErrorNonNotFound() {
	service, mockStore, _, _ := suite.setupTestService()

	app := &model.ApplicationDTO{
		Name: "Test App",
	}

	// Return an error that's not ApplicationNotFoundError
	mockStore.On("GetApplicationByName", "Test App").Return(nil, errors.New("database connection error"))

	result, inboundAuth, svcErr := service.ValidateApplication(app)

	assert.Nil(suite.T(), result)
	assert.Nil(suite.T(), inboundAuth)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &ErrorInternalServerError, svcErr)
}

//nolint:dupl // Testing different URL validation scenarios
func (suite *ServiceTestSuite) TestValidateApplication_InvalidURL() {
	testConfig := &config.Config{
		Flow: config.FlowConfig{
			DefaultAuthFlowHandle: "default_auth_flow",
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetThunderRuntime()

	service, mockStore, _, mockFlowMgtService := suite.setupTestService()

	app := &model.ApplicationDTO{
		Name:       "Test App",
		URL:        "not-a-valid-uri",
		AuthFlowID: "edc013d0-e893-4dc0-990c-3e1d203e005b",
	}

	mockStore.On("GetApplicationByName", "Test App").Return(nil, model.ApplicationNotFoundError)
	mockFlowMgtService.EXPECT().IsValidFlow("edc013d0-e893-4dc0-990c-3e1d203e005b").Return(true)
	mockFlowMgtService.EXPECT().GetFlow("edc013d0-e893-4dc0-990c-3e1d203e005b").Return(&flowmgt.CompleteFlowDefinition{
		ID:     "edc013d0-e893-4dc0-990c-3e1d203e005b",
		Handle: "basic_auth",
	}, nil).Maybe()

	// Return success for registration flow so URL validation runs
	mockFlowMgtService.EXPECT().GetFlowByHandle("basic_auth", flowcommon.FlowTypeRegistration).Return(
		&flowmgt.CompleteFlowDefinition{
			ID:     "reg_flow_basic",
			Handle: "basic_auth",
		}, nil).Maybe()
	mockFlowMgtService.EXPECT().IsValidFlow(mock.Anything).Return(true).Maybe()

	result, inboundAuth, svcErr := service.ValidateApplication(app)

	assert.Nil(suite.T(), result)
	assert.Nil(suite.T(), inboundAuth)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &ErrorInvalidApplicationURL, svcErr)
}

//nolint:dupl // Testing different URL validation scenarios
func (suite *ServiceTestSuite) TestValidateApplication_InvalidLogoURL() {
	testConfig := &config.Config{
		Flow: config.FlowConfig{
			DefaultAuthFlowHandle: "default_auth_flow",
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetThunderRuntime()

	service, mockStore, _, mockFlowMgtService := suite.setupTestService()

	app := &model.ApplicationDTO{
		Name:       "Test App",
		LogoURL:    "not-a-valid-uri",
		AuthFlowID: "edc013d0-e893-4dc0-990c-3e1d203e005b",
	}

	mockStore.On("GetApplicationByName", "Test App").Return(nil, model.ApplicationNotFoundError)
	mockFlowMgtService.EXPECT().IsValidFlow("edc013d0-e893-4dc0-990c-3e1d203e005b").Return(true)
	mockFlowMgtService.EXPECT().GetFlow("edc013d0-e893-4dc0-990c-3e1d203e005b").Return(&flowmgt.CompleteFlowDefinition{
		ID:     "edc013d0-e893-4dc0-990c-3e1d203e005b",
		Handle: "basic_auth",
	}, nil).Maybe()

	// Return success for registration flow so URL validation runs
	mockFlowMgtService.EXPECT().GetFlowByHandle("basic_auth", flowcommon.FlowTypeRegistration).Return(
		&flowmgt.CompleteFlowDefinition{
			ID:     "reg_flow_basic",
			Handle: "basic_auth",
		}, nil).Maybe()
	mockFlowMgtService.EXPECT().IsValidFlow(mock.Anything).Return(true).Maybe()

	result, inboundAuth, svcErr := service.ValidateApplication(app)

	assert.Nil(suite.T(), result)
	assert.Nil(suite.T(), inboundAuth)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &ErrorInvalidLogoURL, svcErr)
}

func (suite *ServiceTestSuite) TestCreateApplication_StoreErrorWithRollback() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
		Flow: config.FlowConfig{
			DefaultAuthFlowHandle: "default_auth_flow",
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetThunderRuntime()

	service, mockStore, mockCertService, mockFlowMgtService := suite.setupTestService()

	app := &model.ApplicationDTO{
		Name:               "Test App",
		AuthFlowID:         "edc013d0-e893-4dc0-990c-3e1d203e005b",
		RegistrationFlowID: "80024fb3-29ed-4c33-aa48-8aee5e96d522",
		Certificate: &model.ApplicationCertificate{
			Type:  "JWKS",
			Value: `{"keys":[]}`,
		},
	}

	mockStore.On("GetApplicationByName", "Test App").Return(nil, model.ApplicationNotFoundError)
	mockFlowMgtService.EXPECT().IsValidFlow("edc013d0-e893-4dc0-990c-3e1d203e005b").Return(true)
	mockFlowMgtService.EXPECT().IsValidFlow("80024fb3-29ed-4c33-aa48-8aee5e96d522").Return(true)
	mockCertService.EXPECT().CreateCertificate(mock.Anything).Return(&cert.Certificate{Type: "JWKS"}, nil)
	mockStore.On("CreateApplication", mock.Anything).Return(errors.New("store error"))
	mockCertService.EXPECT().
		DeleteCertificateByReference(cert.CertificateReferenceTypeApplication, mock.Anything).
		Return(nil)

	result, svcErr := service.CreateApplication(app)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &ErrorInternalServerError, svcErr)
}

func (suite *ServiceTestSuite) TestCreateApplication_StoreErrorWithRollbackFailure() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
		Flow: config.FlowConfig{
			DefaultAuthFlowHandle: "default_auth_flow",
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetThunderRuntime()

	service, mockStore, mockCertService, mockFlowMgtService := suite.setupTestService()

	app := &model.ApplicationDTO{
		Name:               "Test App",
		AuthFlowID:         "edc013d0-e893-4dc0-990c-3e1d203e005b",
		RegistrationFlowID: "80024fb3-29ed-4c33-aa48-8aee5e96d522",
		Certificate: &model.ApplicationCertificate{
			Type:  "JWKS",
			Value: `{"keys":[]}`,
		},
	}

	mockStore.On("GetApplicationByName", "Test App").Return(nil, model.ApplicationNotFoundError)
	mockFlowMgtService.EXPECT().IsValidFlow("edc013d0-e893-4dc0-990c-3e1d203e005b").Return(true)
	mockFlowMgtService.EXPECT().IsValidFlow("80024fb3-29ed-4c33-aa48-8aee5e96d522").Return(true)
	mockCertService.EXPECT().CreateCertificate(mock.Anything).Return(&cert.Certificate{Type: "JWKS"}, nil)
	mockStore.On("CreateApplication", mock.Anything).Return(errors.New("store error"))
	rollbackErr := &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		ErrorDescription: "Failed to rollback",
	}
	mockCertService.EXPECT().
		DeleteCertificateByReference(cert.CertificateReferenceTypeApplication, mock.Anything).
		Return(rollbackErr)

	result, svcErr := service.CreateApplication(app)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
	// Should return the rollback error
	assert.Equal(suite.T(), serviceerror.ClientErrorType, svcErr.Type)
}

func (suite *ServiceTestSuite) TestUpdateApplication_StoreErrorNonNotFound() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetThunderRuntime()

	service, mockStore, _, _ := suite.setupTestService()

	app := &model.ApplicationDTO{
		Name: "Updated App",
	}

	// Return an error that's not ApplicationNotFoundError
	mockStore.On("GetApplicationByID", "app123").Return(nil, errors.New("database connection error"))

	result, svcErr := service.UpdateApplication("app123", app)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &ErrorInternalServerError, svcErr)
}

func (suite *ServiceTestSuite) TestUpdateApplication_StoreErrorWhenCheckingName() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetThunderRuntime()

	service, mockStore, _, _ := suite.setupTestService()

	existingApp := &model.ApplicationProcessedDTO{
		ID:   "app123",
		Name: "Old App",
	}

	app := &model.ApplicationDTO{
		Name: "New App",
	}

	mockStore.On("GetApplicationByID", "app123").Return(existingApp, nil)
	// Return an error that's not ApplicationNotFoundError when checking name
	mockStore.On("GetApplicationByName", "New App").Return(nil, errors.New("database connection error"))

	result, svcErr := service.UpdateApplication("app123", app)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &ErrorInternalServerError, svcErr)
}

func (suite *ServiceTestSuite) TestUpdateApplication_StoreErrorWhenCheckingClientID() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
		Flow: config.FlowConfig{
			DefaultAuthFlowHandle: "default_auth_flow",
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetThunderRuntime()

	service, mockStore, _, mockFlowMgtService := suite.setupTestService()

	existingApp := &model.ApplicationProcessedDTO{
		ID:   "app123",
		Name: "Test App",
		InboundAuthConfig: []model.InboundAuthConfigProcessedDTO{
			{
				OAuthAppConfig: &model.OAuthAppConfigProcessedDTO{
					ClientID: "old-client-id",
				},
			},
		},
	}

	app := &model.ApplicationDTO{
		Name: "Test App",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					ClientID:                "new-client-id",
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
				},
			},
		},
	}

	mockStore.On("GetApplicationByID", "app123").Return(existingApp, nil)
	mockFlowMgtService.EXPECT().IsValidFlow(mock.Anything).Return(true).Maybe()
	// Return an error that's not ApplicationNotFoundError when checking client ID
	mockStore.On("GetOAuthApplication", "new-client-id").Return(nil, errors.New("database connection error"))

	result, svcErr := service.UpdateApplication("app123", app)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &ErrorInternalServerError, svcErr)
}

func (suite *ServiceTestSuite) TestUpdateApplication_StoreErrorWithRollback() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
		Flow: config.FlowConfig{
			DefaultAuthFlowHandle: "default_auth_flow",
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetThunderRuntime()

	service, mockStore, mockCertService, mockFlowMgtService := suite.setupTestService()

	existingApp := &model.ApplicationProcessedDTO{
		ID:   "app123",
		Name: "Test App",
	}

	app := &model.ApplicationDTO{
		ID:                 "app123",
		Name:               "Test App",
		AuthFlowID:         "edc013d0-e893-4dc0-990c-3e1d203e005b",
		RegistrationFlowID: "80024fb3-29ed-4c33-aa48-8aee5e96d522",
		Certificate: &model.ApplicationCertificate{
			Type:  "JWKS",
			Value: `{"keys":[]}`,
		},
	}

	mockStore.On("GetApplicationByID", "app123").Return(existingApp, nil)
	mockFlowMgtService.EXPECT().IsValidFlow("edc013d0-e893-4dc0-990c-3e1d203e005b").Return(true)
	mockFlowMgtService.EXPECT().IsValidFlow("80024fb3-29ed-4c33-aa48-8aee5e96d522").Return(true)
	mockCertService.EXPECT().
		GetCertificateByReference(cert.CertificateReferenceTypeApplication, "app123").
		Return(nil, &cert.ErrorCertificateNotFound)
	mockCertService.EXPECT().CreateCertificate(mock.Anything).Return(&cert.Certificate{Type: "JWKS"}, nil)
	mockStore.On("UpdateApplication", mock.Anything, mock.Anything).Return(errors.New("store error"))
	mockCertService.EXPECT().
		DeleteCertificateByReference(cert.CertificateReferenceTypeApplication, "app123").
		Return(nil)

	result, svcErr := service.UpdateApplication("app123", app)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &ErrorInternalServerError, svcErr)
}

// TestRollbackApplicationCertificateUpdate_UpdateCertificateClientError tests rollback when
// UpdateCertificateByID fails with ClientErrorType
func (suite *ServiceTestSuite) TestRollbackApplicationCertificateUpdate_UpdateCertificateClientError() {
	service, _, mockCertService, _ := suite.setupTestService()

	appID := testAppIDForRollback
	existingCert := &cert.Certificate{
		ID:    "cert-existing-123",
		Type:  cert.CertificateTypeJWKS,
		Value: `{"keys":[]}`,
	}
	updatedCert := &cert.Certificate{
		ID:    "cert-updated-123",
		Type:  cert.CertificateTypeJWKS,
		Value: `{"keys":[{"kty":"RSA"}]}`,
	}

	clientError := &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "CERT-1001",
		Error:            "Certificate validation failed",
		ErrorDescription: "Invalid certificate format",
	}

	mockCertService.EXPECT().
		UpdateCertificateByID(existingCert.ID, existingCert).
		Return(nil, clientError).
		Once()

	svcErr := service.rollbackApplicationCertificateUpdate(appID, existingCert, updatedCert)

	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), ErrorCertificateClientError.Code, svcErr.Code)
	assert.Equal(suite.T(), serviceerror.ClientErrorType, svcErr.Type)
	assert.Contains(suite.T(), svcErr.ErrorDescription, "Failed to revert application certificate update")
	assert.Contains(suite.T(), svcErr.ErrorDescription, "Invalid certificate format")
}

// TestRollbackApplicationCertificateUpdate_UpdateCertificateServerError tests rollback when
// UpdateCertificateByID fails with ServerErrorType
func (suite *ServiceTestSuite) TestRollbackApplicationCertificateUpdate_UpdateCertificateServerError() {
	service, _, mockCertService, _ := suite.setupTestService()

	appID := testAppIDForRollback
	existingCert := &cert.Certificate{
		ID:    "cert-existing-123",
		Type:  cert.CertificateTypeJWKS,
		Value: `{"keys":[]}`,
	}
	updatedCert := &cert.Certificate{
		ID:    "cert-updated-123",
		Type:  cert.CertificateTypeJWKS,
		Value: `{"keys":[{"kty":"RSA"}]}`,
	}

	serverError := &serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "CERT-5001",
		Error:            "Database error",
		ErrorDescription: "Failed to update certificate in database",
	}

	mockCertService.EXPECT().
		UpdateCertificateByID(existingCert.ID, existingCert).
		Return(nil, serverError).
		Once()

	svcErr := service.rollbackApplicationCertificateUpdate(appID, existingCert, updatedCert)

	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &ErrorCertificateServerError, svcErr)
}

// TestRollbackApplicationCertificateUpdate_DeleteCertificateClientError tests rollback when
// DeleteCertificateByReference fails with ClientErrorType
func (suite *ServiceTestSuite) TestRollbackApplicationCertificateUpdate_DeleteCertificateClientError() {
	service, _, mockCertService, _ := suite.setupTestService()

	appID := testAppIDForRollback
	updatedCert := &cert.Certificate{
		ID:    "cert-new-123",
		Type:  cert.CertificateTypeJWKS,
		Value: `{"keys":[{"kty":"RSA"}]}`,
	}

	clientError := &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "CERT-1002",
		Error:            "Certificate not found",
		ErrorDescription: "Certificate does not exist",
	}

	mockCertService.EXPECT().
		DeleteCertificateByReference(cert.CertificateReferenceTypeApplication, appID).
		Return(clientError).
		Once()

	svcErr := service.rollbackApplicationCertificateUpdate(appID, nil, updatedCert)

	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), ErrorCertificateClientError.Code, svcErr.Code)
	assert.Equal(suite.T(), serviceerror.ClientErrorType, svcErr.Type)
	assert.Contains(suite.T(), svcErr.ErrorDescription, "Failed to delete application certificate")
	assert.Contains(suite.T(), svcErr.ErrorDescription, "after update failure")
	assert.Contains(suite.T(), svcErr.ErrorDescription, "Certificate does not exist")
}

// TestRollbackApplicationCertificateUpdate_DeleteCertificateServerError tests rollback when
// DeleteCertificateByReference fails with ServerErrorType
func (suite *ServiceTestSuite) TestRollbackApplicationCertificateUpdate_DeleteCertificateServerError() {
	service, _, mockCertService, _ := suite.setupTestService()

	appID := testAppIDForRollback
	updatedCert := &cert.Certificate{
		ID:    "cert-new-123",
		Type:  cert.CertificateTypeJWKS,
		Value: `{"keys":[{"kty":"RSA"}]}`,
	}

	serverError := &serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "CERT-5002",
		Error:            "Database error",
		ErrorDescription: "Failed to delete certificate from database",
	}

	mockCertService.EXPECT().
		DeleteCertificateByReference(cert.CertificateReferenceTypeApplication, appID).
		Return(serverError).
		Once()

	svcErr := service.rollbackApplicationCertificateUpdate(appID, nil, updatedCert)

	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &ErrorCertificateServerError, svcErr)
}

// TestRollbackApplicationCertificateUpdate_CreateCertificateClientError tests rollback when
// CreateCertificate fails with ClientErrorType
func (suite *ServiceTestSuite) TestRollbackApplicationCertificateUpdate_CreateCertificateClientError() {
	service, _, mockCertService, _ := suite.setupTestService()

	appID := testAppIDForRollback
	existingCert := &cert.Certificate{
		ID:    "cert-existing-123",
		Type:  cert.CertificateTypeJWKS,
		Value: `{"keys":[]}`,
	}

	clientError := &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "CERT-1003",
		Error:            "Certificate validation failed",
		ErrorDescription: "Invalid certificate data",
	}

	mockCertService.EXPECT().
		CreateCertificate(existingCert).
		Return(nil, clientError).
		Once()

	svcErr := service.rollbackApplicationCertificateUpdate(appID, existingCert, nil)

	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), ErrorCertificateClientError.Code, svcErr.Code)
	assert.Equal(suite.T(), serviceerror.ClientErrorType, svcErr.Type)
	assert.Contains(suite.T(), svcErr.ErrorDescription, "Failed to revert application certificate creation")
	assert.Contains(suite.T(), svcErr.ErrorDescription, "Invalid certificate data")
}

// TestRollbackApplicationCertificateUpdate_CreateCertificateServerError tests rollback when
// CreateCertificate fails with ServerErrorType
func (suite *ServiceTestSuite) TestRollbackApplicationCertificateUpdate_CreateCertificateServerError() {
	service, _, mockCertService, _ := suite.setupTestService()

	appID := testAppIDForRollback
	existingCert := &cert.Certificate{
		ID:    "cert-existing-123",
		Type:  cert.CertificateTypeJWKS,
		Value: `{"keys":[]}`,
	}

	serverError := &serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "CERT-5003",
		Error:            "Database error",
		ErrorDescription: "Failed to create certificate in database",
	}

	mockCertService.EXPECT().
		CreateCertificate(existingCert).
		Return(nil, serverError).
		Once()

	svcErr := service.rollbackApplicationCertificateUpdate(appID, existingCert, nil)

	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &ErrorCertificateServerError, svcErr)
}

// TestRollbackApplicationCertificateUpdate_Success_UpdateExisting tests successful rollback
// when updating existing certificate
func (suite *ServiceTestSuite) TestRollbackApplicationCertificateUpdate_Success_UpdateExisting() {
	service, _, mockCertService, _ := suite.setupTestService()

	appID := testAppIDForRollback
	existingCert := &cert.Certificate{
		ID:    "cert-existing-123",
		Type:  cert.CertificateTypeJWKS,
		Value: `{"keys":[]}`,
	}
	updatedCert := &cert.Certificate{
		ID:    "cert-updated-123",
		Type:  cert.CertificateTypeJWKS,
		Value: `{"keys":[{"kty":"RSA"}]}`,
	}

	mockCertService.EXPECT().
		UpdateCertificateByID(existingCert.ID, existingCert).
		Return(&cert.Certificate{ID: existingCert.ID}, nil).
		Once()

	svcErr := service.rollbackApplicationCertificateUpdate(appID, existingCert, updatedCert)

	assert.Nil(suite.T(), svcErr)
}

// TestRollbackApplicationCertificateUpdate_Success_DeleteNew tests successful rollback
// when deleting newly created certificate
func (suite *ServiceTestSuite) TestRollbackApplicationCertificateUpdate_Success_DeleteNew() {
	service, _, mockCertService, _ := suite.setupTestService()

	appID := testAppIDForRollback
	updatedCert := &cert.Certificate{
		ID:    "cert-new-123",
		Type:  cert.CertificateTypeJWKS,
		Value: `{"keys":[{"kty":"RSA"}]}`,
	}

	mockCertService.EXPECT().
		DeleteCertificateByReference(cert.CertificateReferenceTypeApplication, appID).
		Return(nil).
		Once()

	svcErr := service.rollbackApplicationCertificateUpdate(appID, nil, updatedCert)

	assert.Nil(suite.T(), svcErr)
}

// TestRollbackApplicationCertificateUpdate_Success_CreateExisting tests successful rollback
// when recreating previously deleted certificate
func (suite *ServiceTestSuite) TestRollbackApplicationCertificateUpdate_Success_CreateExisting() {
	service, _, mockCertService, _ := suite.setupTestService()

	appID := testAppIDForRollback
	existingCert := &cert.Certificate{
		ID:    "cert-existing-123",
		Type:  cert.CertificateTypeJWKS,
		Value: `{"keys":[]}`,
	}

	mockCertService.EXPECT().
		CreateCertificate(existingCert).
		Return(&cert.Certificate{ID: existingCert.ID}, nil).
		Once()

	svcErr := service.rollbackApplicationCertificateUpdate(appID, existingCert, nil)

	assert.Nil(suite.T(), svcErr)
}

// TestRollbackApplicationCertificateUpdate_NoOp tests rollback when no certificate changes were made
func (suite *ServiceTestSuite) TestRollbackApplicationCertificateUpdate_NoOp() {
	service, _, _, _ := suite.setupTestService()

	appID := testAppIDForRollback

	// No certificates - nothing to rollback
	svcErr := service.rollbackApplicationCertificateUpdate(appID, nil, nil)

	assert.Nil(suite.T(), svcErr)
}

// TestUpdateApplicationCertificate_GetCertificateClientError tests when GetCertificateByReference
// fails with ClientErrorType (non-NotFound)
func (suite *ServiceTestSuite) TestUpdateApplicationCertificate_GetCertificateClientError() {
	service, _, mockCertService, _ := suite.setupTestService()

	app := &model.ApplicationDTO{
		ID:   testAppIDForRollback,
		Name: "Test App",
	}

	clientError := &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "CERT-1001",
		Error:            "Certificate validation failed",
		ErrorDescription: "Invalid certificate reference",
	}

	mockCertService.EXPECT().
		GetCertificateByReference(cert.CertificateReferenceTypeApplication, testAppIDForRollback).
		Return(nil, clientError).
		Once()

	existingCert, updatedCert, returnCert, svcErr := service.updateApplicationCertificate(app)

	assert.Nil(suite.T(), existingCert)
	assert.Nil(suite.T(), updatedCert)
	assert.Nil(suite.T(), returnCert)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), ErrorCertificateClientError.Code, svcErr.Code)
	assert.Equal(suite.T(), serviceerror.ClientErrorType, svcErr.Type)
	assert.Contains(suite.T(), svcErr.ErrorDescription, "Failed to retrieve application certificate")
	assert.Contains(suite.T(), svcErr.ErrorDescription, "Invalid certificate reference")
}

// TestUpdateApplicationCertificate_GetCertificateServerError tests when GetCertificateByReference
// fails with ServerErrorType (non-NotFound)
func (suite *ServiceTestSuite) TestUpdateApplicationCertificate_GetCertificateServerError() {
	service, _, mockCertService, _ := suite.setupTestService()

	app := &model.ApplicationDTO{
		ID:   testAppIDForRollback,
		Name: "Test App",
	}

	serverError := &serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "CERT-5001",
		Error:            "Database error",
		ErrorDescription: "Failed to retrieve certificate from database",
	}

	mockCertService.EXPECT().
		GetCertificateByReference(cert.CertificateReferenceTypeApplication, testAppIDForRollback).
		Return(nil, serverError).
		Once()

	existingCert, updatedCert, returnCert, svcErr := service.updateApplicationCertificate(app)

	assert.Nil(suite.T(), existingCert)
	assert.Nil(suite.T(), updatedCert)
	assert.Nil(suite.T(), returnCert)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &ErrorCertificateServerError, svcErr)
}

// TestUpdateApplicationCertificate_UpdateCertificateClientError tests when UpdateCertificateByID
// fails with ClientErrorType
func (suite *ServiceTestSuite) TestUpdateApplicationCertificate_UpdateCertificateClientError() {
	service, _, mockCertService, _ := suite.setupTestService()

	existingCert := &cert.Certificate{
		ID:    "cert-existing-123",
		Type:  cert.CertificateTypeJWKS,
		Value: `{"keys":[]}`,
	}

	app := &model.ApplicationDTO{
		ID:   testAppIDForRollback,
		Name: "Test App",
		Certificate: &model.ApplicationCertificate{
			Type:  cert.CertificateTypeJWKS,
			Value: `{"keys":[{"kty":"RSA"}]}`,
		},
	}

	clientError := &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "CERT-1002",
		Error:            "Certificate validation failed",
		ErrorDescription: "Invalid certificate format",
	}

	mockCertService.EXPECT().
		GetCertificateByReference(cert.CertificateReferenceTypeApplication, testAppIDForRollback).
		Return(existingCert, nil).
		Once()
	mockCertService.EXPECT().
		UpdateCertificateByID(existingCert.ID, mock.Anything).
		Return(nil, clientError).
		Once()

	existingCertResult, updatedCert, returnCert, svcErr := service.updateApplicationCertificate(app)

	assert.Nil(suite.T(), existingCertResult)
	assert.Nil(suite.T(), updatedCert)
	assert.Nil(suite.T(), returnCert)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), ErrorCertificateClientError.Code, svcErr.Code)
	assert.Equal(suite.T(), serviceerror.ClientErrorType, svcErr.Type)
	assert.Contains(suite.T(), svcErr.ErrorDescription, "Failed to update application certificate")
	assert.Contains(suite.T(), svcErr.ErrorDescription, "Invalid certificate format")
}

// TestUpdateApplicationCertificate_UpdateCertificateServerError tests when UpdateCertificateByID
// fails with ServerErrorType
func (suite *ServiceTestSuite) TestUpdateApplicationCertificate_UpdateCertificateServerError() {
	service, _, mockCertService, _ := suite.setupTestService()

	existingCert := &cert.Certificate{
		ID:    "cert-existing-123",
		Type:  cert.CertificateTypeJWKS,
		Value: `{"keys":[]}`,
	}

	app := &model.ApplicationDTO{
		ID:   testAppIDForRollback,
		Name: "Test App",
		Certificate: &model.ApplicationCertificate{
			Type:  cert.CertificateTypeJWKS,
			Value: `{"keys":[{"kty":"RSA"}]}`,
		},
	}

	serverError := &serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "CERT-5002",
		Error:            "Database error",
		ErrorDescription: "Failed to update certificate in database",
	}

	mockCertService.EXPECT().
		GetCertificateByReference(cert.CertificateReferenceTypeApplication, testAppIDForRollback).
		Return(existingCert, nil).
		Once()
	mockCertService.EXPECT().
		UpdateCertificateByID(existingCert.ID, mock.Anything).
		Return(nil, serverError).
		Once()

	existingCertResult, updatedCert, returnCert, svcErr := service.updateApplicationCertificate(app)

	assert.Nil(suite.T(), existingCertResult)
	assert.Nil(suite.T(), updatedCert)
	assert.Nil(suite.T(), returnCert)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &ErrorCertificateServerError, svcErr)
}

// TestUpdateApplicationCertificate_CreateCertificateClientError tests when CreateCertificate
// fails with ClientErrorType (when creating new certificate)
func (suite *ServiceTestSuite) TestUpdateApplicationCertificate_CreateCertificateClientError() {
	service, _, mockCertService, _ := suite.setupTestService()

	app := &model.ApplicationDTO{
		ID:   testAppIDForRollback,
		Name: "Test App",
		Certificate: &model.ApplicationCertificate{
			Type:  cert.CertificateTypeJWKS,
			Value: `{"keys":[{"kty":"RSA"}]}`,
		},
	}

	clientError := &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "CERT-1003",
		Error:            "Certificate validation failed",
		ErrorDescription: "Invalid certificate data",
	}

	mockCertService.EXPECT().
		GetCertificateByReference(cert.CertificateReferenceTypeApplication, testAppIDForRollback).
		Return(nil, &cert.ErrorCertificateNotFound).
		Once()
	mockCertService.EXPECT().
		CreateCertificate(mock.Anything).
		Return(nil, clientError).
		Once()

	existingCert, updatedCert, returnCert, svcErr := service.updateApplicationCertificate(app)

	assert.Nil(suite.T(), existingCert)
	assert.Nil(suite.T(), updatedCert)
	assert.Nil(suite.T(), returnCert)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), ErrorCertificateClientError.Code, svcErr.Code)
	assert.Equal(suite.T(), serviceerror.ClientErrorType, svcErr.Type)
	assert.Contains(suite.T(), svcErr.ErrorDescription, "Failed to create application certificate")
	assert.Contains(suite.T(), svcErr.ErrorDescription, "Invalid certificate data")
}

// TestUpdateApplicationCertificate_CreateCertificateServerError tests when CreateCertificate
// fails with ServerErrorType (when creating new certificate)
func (suite *ServiceTestSuite) TestUpdateApplicationCertificate_CreateCertificateServerError() {
	service, _, mockCertService, _ := suite.setupTestService()

	app := &model.ApplicationDTO{
		ID:   testAppIDForRollback,
		Name: "Test App",
		Certificate: &model.ApplicationCertificate{
			Type:  cert.CertificateTypeJWKS,
			Value: `{"keys":[{"kty":"RSA"}]}`,
		},
	}

	serverError := &serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "CERT-5003",
		Error:            "Database error",
		ErrorDescription: "Failed to create certificate in database",
	}

	mockCertService.EXPECT().
		GetCertificateByReference(cert.CertificateReferenceTypeApplication, testAppIDForRollback).
		Return(nil, &cert.ErrorCertificateNotFound).
		Once()
	mockCertService.EXPECT().
		CreateCertificate(mock.Anything).
		Return(nil, serverError).
		Once()

	existingCert, updatedCert, returnCert, svcErr := service.updateApplicationCertificate(app)

	assert.Nil(suite.T(), existingCert)
	assert.Nil(suite.T(), updatedCert)
	assert.Nil(suite.T(), returnCert)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &ErrorCertificateServerError, svcErr)
}

// TestUpdateApplicationCertificate_DeleteCertificateClientError tests when DeleteCertificateByReference
// fails with ClientErrorType (when removing existing certificate)
func (suite *ServiceTestSuite) TestUpdateApplicationCertificate_DeleteCertificateClientError() {
	service, _, mockCertService, _ := suite.setupTestService()

	existingCert := &cert.Certificate{
		ID:    "cert-existing-123",
		Type:  cert.CertificateTypeJWKS,
		Value: `{"keys":[]}`,
	}

	app := &model.ApplicationDTO{
		ID:   testAppIDForRollback,
		Name: "Test App",
		// No certificate provided - should delete existing
	}

	clientError := &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "CERT-1004",
		Error:            "Certificate not found",
		ErrorDescription: "Certificate does not exist",
	}

	mockCertService.EXPECT().
		GetCertificateByReference(cert.CertificateReferenceTypeApplication, testAppIDForRollback).
		Return(existingCert, nil).
		Once()
	mockCertService.EXPECT().
		DeleteCertificateByReference(cert.CertificateReferenceTypeApplication, testAppIDForRollback).
		Return(clientError).
		Once()

	existingCertResult, updatedCert, returnCert, svcErr := service.updateApplicationCertificate(app)

	assert.Nil(suite.T(), existingCertResult)
	assert.Nil(suite.T(), updatedCert)
	assert.Nil(suite.T(), returnCert)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), ErrorCertificateClientError.Code, svcErr.Code)
	assert.Equal(suite.T(), serviceerror.ClientErrorType, svcErr.Type)
	assert.Contains(suite.T(), svcErr.ErrorDescription, "Failed to delete application certificate")
	assert.Contains(suite.T(), svcErr.ErrorDescription, "Certificate does not exist")
}

// TestUpdateApplicationCertificate_DeleteCertificateServerError tests when DeleteCertificateByReference
// fails with ServerErrorType (when removing existing certificate)
func (suite *ServiceTestSuite) TestUpdateApplicationCertificate_DeleteCertificateServerError() {
	service, _, mockCertService, _ := suite.setupTestService()

	existingCert := &cert.Certificate{
		ID:    "cert-existing-123",
		Type:  cert.CertificateTypeJWKS,
		Value: `{"keys":[]}`,
	}

	app := &model.ApplicationDTO{
		ID:   testAppIDForRollback,
		Name: "Test App",
		// No certificate provided - should delete existing
	}

	serverError := &serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "CERT-5004",
		Error:            "Database error",
		ErrorDescription: "Failed to delete certificate from database",
	}

	mockCertService.EXPECT().
		GetCertificateByReference(cert.CertificateReferenceTypeApplication, testAppIDForRollback).
		Return(existingCert, nil).
		Once()
	mockCertService.EXPECT().
		DeleteCertificateByReference(cert.CertificateReferenceTypeApplication, testAppIDForRollback).
		Return(serverError).
		Once()

	existingCertResult, updatedCert, returnCert, svcErr := service.updateApplicationCertificate(app)

	assert.Nil(suite.T(), existingCertResult)
	assert.Nil(suite.T(), updatedCert)
	assert.Nil(suite.T(), returnCert)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &ErrorCertificateServerError, svcErr)
}

// TestValidateAllowedUserTypes_EmptyString tests when an empty string is provided
// in allowedUserTypes, which should be treated as invalid
func (suite *ServiceTestSuite) TestValidateAllowedUserTypes_EmptyString() {
	// Mock GetUserSchemaList to return an empty list
	mockStore := newApplicationStoreInterfaceMock(suite.T())
	mockCertService := certmock.NewCertificateServiceInterfaceMock(suite.T())
	mockFlowMgtService := flowmgtmock.NewFlowMgtServiceInterfaceMock(suite.T())
	mockUserSchemaService := userschemamock.NewUserSchemaServiceInterfaceMock(suite.T())

	// Mock GetUserSchemaList to return empty list (first call)
	mockUserSchemaService.EXPECT().
		GetUserSchemaList(mock.Anything, 0).
		Return(&userschema.UserSchemaListResponse{
			TotalResults: 0,
			Count:        0,
			Schemas:      []userschema.UserSchemaListItem{},
		}, nil).
		Once()

	serviceWithMock := &applicationService{
		appStore:          mockStore,
		certService:       mockCertService,
		flowMgtService:    mockFlowMgtService,
		userSchemaService: mockUserSchemaService,
	}

	// Test with empty string in allowedUserTypes
	allowedUserTypes := []string{""}
	svcErr := serviceWithMock.validateAllowedUserTypes(allowedUserTypes)

	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &ErrorInvalidUserType, svcErr)
}

// TestValidateAllowedUserTypes_EmptyStringWithValidTypes tests when an empty string
// is provided along with valid user types
func (suite *ServiceTestSuite) TestValidateAllowedUserTypes_EmptyStringWithValidTypes() {
	mockStore := newApplicationStoreInterfaceMock(suite.T())
	mockCertService := certmock.NewCertificateServiceInterfaceMock(suite.T())
	mockFlowMgtService := flowmgtmock.NewFlowMgtServiceInterfaceMock(suite.T())
	mockUserSchemaService := userschemamock.NewUserSchemaServiceInterfaceMock(suite.T())

	// Mock GetUserSchemaList to return a list with one valid user type
	mockUserSchemaService.EXPECT().
		GetUserSchemaList(mock.Anything, 0).
		Return(&userschema.UserSchemaListResponse{
			TotalResults: 1,
			Count:        1,
			Schemas: []userschema.UserSchemaListItem{
				{
					Name: "validUserType",
				},
			},
		}, nil).
		Once()

	serviceWithMock := &applicationService{
		appStore:          mockStore,
		certService:       mockCertService,
		flowMgtService:    mockFlowMgtService,
		userSchemaService: mockUserSchemaService,
	}

	// Test with empty string and valid user type
	allowedUserTypes := []string{"", "validUserType"}
	svcErr := serviceWithMock.validateAllowedUserTypes(allowedUserTypes)

	// Should still fail because empty string is invalid
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &ErrorInvalidUserType, svcErr)
}

func (suite *ServiceTestSuite) TestValidateRegistrationFlowID_NoPrefix() {
	testConfig := &config.Config{
		Flow: config.FlowConfig{
			DefaultAuthFlowHandle: "default_auth_flow",
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetThunderRuntime()

	service, mockStore, _, mockFlowMgtService := suite.setupTestService()

	app := &model.ApplicationDTO{
		Name:               "Test App",
		AuthFlowID:         "invalid_flow_id", // Doesn't have prefix
		RegistrationFlowID: "",                // Empty, should infer from auth flow
	}

	mockStore.On("GetApplicationByName", "Test App").Return(nil, model.ApplicationNotFoundError)
	mockFlowMgtService.EXPECT().IsValidFlow("invalid_flow_id").Return(true)
	mockFlowMgtService.EXPECT().GetFlow("invalid_flow_id").Return(&flowmgt.CompleteFlowDefinition{
		ID:     "invalid_flow_id",
		Handle: "test_flow",
	}, nil).Maybe()
	mockFlowMgtService.EXPECT().GetFlowByHandle(mock.Anything, flowcommon.FlowTypeRegistration).Return(
		nil, &serviceerror.ServiceError{Type: serviceerror.ClientErrorType}).Maybe()

	result, inboundAuth, svcErr := service.ValidateApplication(app)

	assert.Nil(suite.T(), result)
	assert.Nil(suite.T(), inboundAuth)
	assert.NotNil(suite.T(), svcErr)
	// When registration flow can't be inferred from auth flow, we get ErrorWhileRetrievingFlowDefinition
	assert.Equal(suite.T(), &ErrorWhileRetrievingFlowDefinition, svcErr)
}
