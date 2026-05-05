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
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/pkg/application/model"
	inboundmodel "github.com/asgardeo/thunder/pkg/inboundclient/model"
	oauth2const "github.com/asgardeo/thunder/pkg/oauth/oauth2/constants"
)

// ApplicationStoreTestSuite covers helpers remaining in the application package after the
// store layer moved to inboundclient. Comprehensive store CRUD tests now live in the
// inboundclient package.
type ApplicationStoreTestSuite struct {
	suite.Suite
}

func TestApplicationStoreTestSuite(t *testing.T) {
	suite.Run(t, new(ApplicationStoreTestSuite))
}

func (suite *ApplicationStoreTestSuite) createTestApplication() model.ApplicationProcessedDTO {
	return model.ApplicationProcessedDTO{
		ID:                        "app1",
		Name:                      "Test App 1",
		Description:               "Test application description",
		AuthFlowID:                "auth_flow_1",
		RegistrationFlowID:        "reg_flow_1",
		IsRegistrationFlowEnabled: true,
		URL:                       "https://example.com",
		LogoURL:                   "https://example.com/logo.png",
		TosURI:                    "https://example.com/tos",
		PolicyURI:                 "https://example.com/policy",
		Contacts:                  []string{"contact@example.com", "support@example.com"},
		Assertion: &inboundmodel.AssertionConfig{
			ValidityPeriod: 3600,
			UserAttributes: []string{"email", "name", "sub"},
		},
		InboundAuthConfig: []model.InboundAuthConfigProcessedDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &inboundmodel.OAuthClient{
					AppID:        "app1",
					ClientID:     "client_app1",
					RedirectURIs: []string{"https://example.com/callback", "https://example.com/cb2"},
					GrantTypes: []oauth2const.GrantType{
						oauth2const.GrantTypeAuthorizationCode,
						oauth2const.GrantTypeRefreshToken,
					},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretPost,
					PKCERequired:            true,
					PublicClient:            false,
					Scopes:                  []string{"openid", "profile", "email"},
					Token: &inboundmodel.OAuthTokenConfig{
						AccessToken: &inboundmodel.AccessTokenConfig{
							ValidityPeriod: 7200,
							UserAttributes: []string{"sub", "email", "name"},
						},
						IDToken: &inboundmodel.IDTokenConfig{
							ValidityPeriod: 3600,
							UserAttributes: []string{"sub", "email", "name", "given_name"},
						},
					},
					ScopeClaims: map[string][]string{
						"profile": {"name", "given_name", "family_name"},
						"email":   {"email", "email_verified"},
					},
				},
			},
		},
	}
}

// --- Tests for buildOAuthConfigData ---

func (suite *ApplicationStoreTestSuite) TestBuildOAuthConfigData_Success() {
	app := suite.createTestApplication()
	inboundAuthConfig := app.InboundAuthConfig[0]

	cfg := buildOAuthConfigData(inboundAuthConfig)

	suite.NotNil(cfg)
	suite.Equal([]string{"authorization_code", "refresh_token"}, cfg.GrantTypes)
	suite.Equal([]string{"code"}, cfg.ResponseTypes)
	suite.Equal("client_secret_post", cfg.TokenEndpointAuthMethod)
	suite.True(cfg.PKCERequired)
	suite.False(cfg.PublicClient)
	suite.Len(cfg.RedirectURIs, 2)
	suite.NotNil(cfg.Token)
	suite.NotNil(cfg.Token.AccessToken)
	suite.Equal(int64(7200), cfg.Token.AccessToken.ValidityPeriod)
	suite.NotNil(cfg.Token.IDToken)
	suite.Equal(int64(3600), cfg.Token.IDToken.ValidityPeriod)
	suite.NotNil(cfg.ScopeClaims)
}

func (suite *ApplicationStoreTestSuite) TestBuildOAuthConfigData_WithoutToken() {
	app := suite.createTestApplication()
	inboundAuthConfig := app.InboundAuthConfig[0]
	inboundAuthConfig.OAuthAppConfig.Token = nil

	cfg := buildOAuthConfigData(inboundAuthConfig)

	suite.NotNil(cfg)
	suite.Nil(cfg.Token)
}

func (suite *ApplicationStoreTestSuite) TestBuildOAuthConfigData_WithoutAccessToken() {
	app := suite.createTestApplication()
	inboundAuthConfig := app.InboundAuthConfig[0]
	inboundAuthConfig.OAuthAppConfig.Token.AccessToken = nil

	cfg := buildOAuthConfigData(inboundAuthConfig)

	suite.NotNil(cfg)
	suite.NotNil(cfg.Token)
	suite.Nil(cfg.Token.AccessToken)
	suite.NotNil(cfg.Token.IDToken)
}

func (suite *ApplicationStoreTestSuite) TestBuildOAuthConfigData_WithoutIDToken() {
	app := suite.createTestApplication()
	inboundAuthConfig := app.InboundAuthConfig[0]
	inboundAuthConfig.OAuthAppConfig.Token.IDToken = nil

	cfg := buildOAuthConfigData(inboundAuthConfig)

	suite.NotNil(cfg)
	suite.NotNil(cfg.Token)
	suite.NotNil(cfg.Token.AccessToken)
	suite.Nil(cfg.Token.IDToken)
}

func (suite *ApplicationStoreTestSuite) TestBuildOAuthConfigData_WithUserInfo() {
	app := suite.createTestApplication()
	inboundAuthConfig := app.InboundAuthConfig[0]
	inboundAuthConfig.OAuthAppConfig.UserInfo = &inboundmodel.UserInfoConfig{
		ResponseType:   "jwt",
		UserAttributes: []string{"email", "name"},
	}

	cfg := buildOAuthConfigData(inboundAuthConfig)

	suite.NotNil(cfg)
	suite.NotNil(cfg.UserInfo)
	suite.Equal(inboundmodel.UserInfoResponseType("jwt"), cfg.UserInfo.ResponseType)
	suite.Len(cfg.UserInfo.UserAttributes, 2)
}

func (suite *ApplicationStoreTestSuite) TestBuildOAuthConfigData_NilOAuthConfig() {
	cfg := buildOAuthConfigData(model.InboundAuthConfigProcessedDTO{})
	suite.Nil(cfg)
}
