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

package model

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sysconfig "github.com/asgardeo/thunder/internal/system/config"
	oauth2const "github.com/asgardeo/thunder/pkg/oauth/oauth2/constants"
)

const (
	errRedirectURIFragment          = "redirect URI must not contain a fragment component"
	errRedirectURINotRegistered     = "your application's redirect URL does not match with the registered redirect URLs"
	errRedirectURIRequired          = "redirect URI is required in the authorization request"
	errRedirectURINotFullyQualified = "registered redirect URI is not fully qualified"
)

type OAuthAppConfigDTOTestSuite struct {
	suite.Suite
}

func TestOAuthAppConfigDTOTestSuite(t *testing.T) {
	suite.Run(t, new(OAuthAppConfigDTOTestSuite))
}

func (suite *OAuthAppConfigDTOTestSuite) SetupTest() {
	sysconfig.ResetServerRuntime()
	err := sysconfig.InitializeServerRuntime("/tmp/test", &sysconfig.Config{
		OAuth: sysconfig.OAuthConfig{AllowWildcardRedirectURI: true},
	})
	suite.Require().NoError(err)
}

func (suite *OAuthAppConfigDTOTestSuite) TearDownTest() {
	sysconfig.ResetServerRuntime()
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedGrantType_AuthorizationCode() {
	config := &OAuthAppConfigDTO{
		GrantTypes: []oauth2const.GrantType{
			oauth2const.GrantTypeAuthorizationCode,
			oauth2const.GrantTypeRefreshToken,
		},
	}

	suite.True(config.IsAllowedGrantType(oauth2const.GrantTypeAuthorizationCode))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedGrantType_ClientCredentials() {
	config := &OAuthAppConfigDTO{
		GrantTypes: []oauth2const.GrantType{
			oauth2const.GrantTypeClientCredentials,
		},
	}

	suite.True(config.IsAllowedGrantType(oauth2const.GrantTypeClientCredentials))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedGrantType_RefreshToken() {
	config := &OAuthAppConfigDTO{
		GrantTypes: []oauth2const.GrantType{
			oauth2const.GrantTypeRefreshToken,
		},
	}

	suite.True(config.IsAllowedGrantType(oauth2const.GrantTypeRefreshToken))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedGrantType_TokenExchange() {
	config := &OAuthAppConfigDTO{
		GrantTypes: []oauth2const.GrantType{
			oauth2const.GrantTypeTokenExchange,
		},
	}

	suite.True(config.IsAllowedGrantType(oauth2const.GrantTypeTokenExchange))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedGrantType_NotAllowed() {
	config := &OAuthAppConfigDTO{
		GrantTypes: []oauth2const.GrantType{
			oauth2const.GrantTypeAuthorizationCode,
		},
	}

	suite.False(config.IsAllowedGrantType(oauth2const.GrantTypeClientCredentials))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedGrantType_EmptyGrantType() {
	config := &OAuthAppConfigDTO{
		GrantTypes: []oauth2const.GrantType{
			oauth2const.GrantTypeAuthorizationCode,
		},
	}

	suite.False(config.IsAllowedGrantType(""))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedGrantType_EmptyGrantTypesList() {
	config := &OAuthAppConfigDTO{
		GrantTypes: []oauth2const.GrantType{},
	}

	suite.False(config.IsAllowedGrantType(oauth2const.GrantTypeAuthorizationCode))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedGrantType_NilGrantTypesList() {
	config := &OAuthAppConfigDTO{
		GrantTypes: nil,
	}

	suite.False(config.IsAllowedGrantType(oauth2const.GrantTypeAuthorizationCode))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedGrantType_MultipleGrantTypes() {
	config := &OAuthAppConfigDTO{
		GrantTypes: []oauth2const.GrantType{
			oauth2const.GrantTypeAuthorizationCode,
			oauth2const.GrantTypeClientCredentials,
			oauth2const.GrantTypeRefreshToken,
			oauth2const.GrantTypeTokenExchange,
		},
	}

	suite.True(config.IsAllowedGrantType(oauth2const.GrantTypeAuthorizationCode))
	suite.True(config.IsAllowedGrantType(oauth2const.GrantTypeClientCredentials))
	suite.True(config.IsAllowedGrantType(oauth2const.GrantTypeRefreshToken))
	suite.True(config.IsAllowedGrantType(oauth2const.GrantTypeTokenExchange))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedResponseType_Code() {
	config := &OAuthAppConfigDTO{
		ResponseTypes: []oauth2const.ResponseType{
			oauth2const.ResponseTypeCode,
		},
	}

	suite.True(config.IsAllowedResponseType("code"))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedResponseType_NotAllowed() {
	config := &OAuthAppConfigDTO{
		ResponseTypes: []oauth2const.ResponseType{
			oauth2const.ResponseTypeCode,
		},
	}

	suite.False(config.IsAllowedResponseType("token"))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedResponseType_EmptyResponseType() {
	config := &OAuthAppConfigDTO{
		ResponseTypes: []oauth2const.ResponseType{
			oauth2const.ResponseTypeCode,
		},
	}

	suite.False(config.IsAllowedResponseType(""))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedResponseType_EmptyResponseTypesList() {
	config := &OAuthAppConfigDTO{
		ResponseTypes: []oauth2const.ResponseType{},
	}

	suite.False(config.IsAllowedResponseType("code"))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedResponseType_NilResponseTypesList() {
	config := &OAuthAppConfigDTO{
		ResponseTypes: nil,
	}

	suite.False(config.IsAllowedResponseType("code"))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedResponseType_MultipleResponseTypes() {
	config := &OAuthAppConfigDTO{
		ResponseTypes: []oauth2const.ResponseType{
			oauth2const.ResponseTypeCode,
			"token",
			"id_token",
		},
	}

	suite.True(config.IsAllowedResponseType("code"))
	suite.True(config.IsAllowedResponseType("token"))
	suite.True(config.IsAllowedResponseType("id_token"))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedTokenEndpointAuthMethod_ClientSecretBasic() {
	config := &OAuthAppConfigDTO{
		TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
	}

	suite.True(config.IsAllowedTokenEndpointAuthMethod(oauth2const.TokenEndpointAuthMethodClientSecretBasic))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedTokenEndpointAuthMethod_ClientSecretPost() {
	config := &OAuthAppConfigDTO{
		TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretPost,
	}

	suite.True(config.IsAllowedTokenEndpointAuthMethod(oauth2const.TokenEndpointAuthMethodClientSecretPost))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedTokenEndpointAuthMethod_None() {
	config := &OAuthAppConfigDTO{
		TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodNone,
	}

	suite.True(config.IsAllowedTokenEndpointAuthMethod(oauth2const.TokenEndpointAuthMethodNone))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedTokenEndpointAuthMethod_NotAllowed() {
	config := &OAuthAppConfigDTO{
		TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
	}

	suite.False(config.IsAllowedTokenEndpointAuthMethod(oauth2const.TokenEndpointAuthMethodClientSecretPost))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedTokenEndpointAuthMethod_Empty() {
	config := &OAuthAppConfigDTO{
		TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
	}

	suite.False(config.IsAllowedTokenEndpointAuthMethod(""))
}

func (suite *OAuthAppConfigDTOTestSuite) TestValidateRedirectURI_ValidWithSingleRegisteredURI() {
	config := &OAuthAppConfigDTO{
		RedirectURIs: []string{"https://example.com/callback"},
	}

	err := config.ValidateRedirectURI("https://example.com/callback")
	suite.NoError(err)
}

func (suite *OAuthAppConfigDTOTestSuite) TestValidateRedirectURI_ValidHTTPLocalhostWithPort() {
	config := &OAuthAppConfigDTO{
		RedirectURIs: []string{"http://localhost:3000/callback"},
	}

	err := config.ValidateRedirectURI("http://localhost:3000/callback")
	suite.NoError(err)
}

func (suite *OAuthAppConfigDTOTestSuite) TestValidateRedirectURI_ValidHTTPSWithPath() {
	config := &OAuthAppConfigDTO{
		RedirectURIs: []string{"https://app.example.com/oauth/callback"},
	}

	err := config.ValidateRedirectURI("https://app.example.com/oauth/callback")
	suite.NoError(err)
}

func (suite *OAuthAppConfigDTOTestSuite) TestValidateRedirectURI_ValidCustomScheme() {
	config := &OAuthAppConfigDTO{
		RedirectURIs: []string{"myapp://callback"},
	}

	err := config.ValidateRedirectURI("myapp://callback")
	suite.NoError(err)
}

func (suite *OAuthAppConfigDTOTestSuite) TestValidateRedirectURI_ValidWithQueryParameters() {
	config := &OAuthAppConfigDTO{
		RedirectURIs: []string{"https://example.com/callback?param=value"},
	}

	err := config.ValidateRedirectURI("https://example.com/callback?param=value")
	suite.NoError(err)
}

func (suite *OAuthAppConfigDTOTestSuite) TestValidateRedirectURI_InvalidWithFragment() {
	config := &OAuthAppConfigDTO{
		RedirectURIs: []string{"https://example.com/callback#fragment"},
	}

	err := config.ValidateRedirectURI("https://example.com/callback#fragment")
	suite.Error(err)
	suite.Equal(errRedirectURIFragment, err.Error())
}

func (suite *OAuthAppConfigDTOTestSuite) TestValidateRedirectURI_NotRegistered() {
	config := &OAuthAppConfigDTO{
		RedirectURIs: []string{"https://example.com/callback"},
	}

	err := config.ValidateRedirectURI("https://different.com/callback")
	suite.Error(err)
	suite.Equal(errRedirectURINotRegistered, err.Error())
}

func (suite *OAuthAppConfigDTOTestSuite) TestValidateRedirectURI_EmptyWithSingleFullyQualifiedURI() {
	config := &OAuthAppConfigDTO{
		RedirectURIs: []string{"https://example.com/callback"},
	}

	err := config.ValidateRedirectURI("")
	suite.NoError(err)
}

func (suite *OAuthAppConfigDTOTestSuite) TestValidateRedirectURI_EmptyWithMultipleURIs() {
	config := &OAuthAppConfigDTO{
		RedirectURIs: []string{
			"https://example.com/callback",
			"https://example.com/callback2",
		},
	}

	err := config.ValidateRedirectURI("")
	suite.Error(err)
	suite.Equal(errRedirectURIRequired, err.Error())
}

func (suite *OAuthAppConfigDTOTestSuite) TestValidateRedirectURI_EmptyWithPartialRegisteredURI() {
	config := &OAuthAppConfigDTO{
		RedirectURIs: []string{"/callback"},
	}

	err := config.ValidateRedirectURI("")
	suite.Error(err)
	suite.Equal(errRedirectURINotFullyQualified, err.Error())
}

func (suite *OAuthAppConfigDTOTestSuite) TestValidateRedirectURI_EmptyWithInvalidRegisteredURI() {
	config := &OAuthAppConfigDTO{
		RedirectURIs: []string{"://invalid"},
	}

	err := config.ValidateRedirectURI("")
	suite.Error(err)
	suite.Equal(errRedirectURINotFullyQualified, err.Error())
}

func (suite *OAuthAppConfigDTOTestSuite) TestValidateRedirectURI_EmptyRedirectURIsList() {
	config := &OAuthAppConfigDTO{
		RedirectURIs: []string{},
	}

	err := config.ValidateRedirectURI("")
	suite.Error(err)
	suite.Equal(errRedirectURIRequired, err.Error())
}

func (suite *OAuthAppConfigDTOTestSuite) TestValidateRedirectURI_NilRedirectURIsList() {
	config := &OAuthAppConfigDTO{
		RedirectURIs: nil,
	}

	err := config.ValidateRedirectURI("")
	suite.Error(err)
	suite.Equal(errRedirectURIRequired, err.Error())
}
