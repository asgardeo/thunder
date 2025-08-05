/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
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

package constants

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type OAuth2ConstantsTestSuite struct {
	suite.Suite
}

func TestOAuth2ConstantsSuite(t *testing.T) {
	suite.Run(t, new(OAuth2ConstantsTestSuite))
}

func (suite *OAuth2ConstantsTestSuite) TestRequestParameterConstants() {
	// Test that all request parameter constants have expected values
	assert.Equal(suite.T(), "grant_type", RequestParamGrantType)
	assert.Equal(suite.T(), "client_id", RequestParamClientID)
	assert.Equal(suite.T(), "client_secret", RequestParamClientSecret)
	assert.Equal(suite.T(), "redirect_uri", RequestParamRedirectURI)
	assert.Equal(suite.T(), "username", RequestParamUsername)
	assert.Equal(suite.T(), "password", RequestParamPassword)
	assert.Equal(suite.T(), "scope", RequestParamScope)
	assert.Equal(suite.T(), "code", RequestParamCode)
	assert.Equal(suite.T(), "code_verifier", RequestParamCodeVerifier)
	assert.Equal(suite.T(), "refresh_token", RequestParamRefreshToken)
	assert.Equal(suite.T(), "response_type", RequestParamResponseType)
	assert.Equal(suite.T(), "state", RequestParamState)
	assert.Equal(suite.T(), "error", RequestParamError)
	assert.Equal(suite.T(), "error_description", RequestParamErrorDescription)
}

func (suite *OAuth2ConstantsTestSuite) TestServerOAuthConstants() {
	// Test server OAuth constants
	assert.Equal(suite.T(), "sessionDataKey", SessionDataKey)
	assert.Equal(suite.T(), "sessionDataKeyConsent", SessionDataKeyConsent)
	assert.Equal(suite.T(), "showInsecureWarning", ShowInsecureWarning)
	assert.Equal(suite.T(), "applicationId", AppID)
	assert.Equal(suite.T(), "assertion", Assertion)
}

func (suite *OAuth2ConstantsTestSuite) TestMessageTypeConstants() {
	// Test OAuth message type constants
	assert.Equal(suite.T(), "initialAuthorizationRequest", TypeInitialAuthorizationRequest)
	assert.Equal(suite.T(), "authorizationResponseFromEngine", TypeAuthorizationResponseFromEngine)
	assert.Equal(suite.T(), "consentResponseFromUser", TypeConsentResponseFromUser)
}

func (suite *OAuth2ConstantsTestSuite) TestEndpointConstants() {
	// Test OAuth2 endpoint constants
	assert.Equal(suite.T(), "/oauth2/token", OAuth2TokenEndpoint)
	assert.Equal(suite.T(), "/oauth2/authorize", OAuth2AuthorizationEndpoint)
	assert.Equal(suite.T(), "/oauth2/introspect", OAuth2IntrospectionEndpoint)
	assert.Equal(suite.T(), "/oauth2/revoke", OAuth2RevokeEndpoint)
	assert.Equal(suite.T(), "/oauth2/userinfo", OAuth2UserInfoEndpoint)
	assert.Equal(suite.T(), "/oauth2/jwks", OAuth2JWKSEndpoint)
	assert.Equal(suite.T(), "/oauth2/logout", OAuth2LogoutEndpoint)
}

func (suite *OAuth2ConstantsTestSuite) TestGrantTypeConstants() {
	// Test grant type constants
	assert.Equal(suite.T(), GrantType("authorization_code"), GrantTypeAuthorizationCode)
	assert.Equal(suite.T(), GrantType("client_credentials"), GrantTypeClientCredentials)
	assert.Equal(suite.T(), GrantType("password"), GrantTypePassword)
	assert.Equal(suite.T(), GrantType("implicit"), GrantTypeImplicit)
	assert.Equal(suite.T(), GrantType("refresh_token"), GrantTypeRefreshToken)
}

func (suite *OAuth2ConstantsTestSuite) TestGrantType_IsValid() {
	// Test valid grant types
	validGrantTypes := []GrantType{
		GrantTypeAuthorizationCode,
		GrantTypeClientCredentials,
		GrantTypePassword,
		GrantTypeImplicit,
		GrantTypeRefreshToken,
	}

	for _, gt := range validGrantTypes {
		assert.True(suite.T(), gt.IsValid(), "Grant type %s should be valid", gt)
	}

	// Test invalid grant types
	invalidGrantTypes := []GrantType{
		"invalid_grant",
		"",
		"some_random_grant",
		"AUTHORIZATION_CODE", // case sensitive
	}

	for _, gt := range invalidGrantTypes {
		assert.False(suite.T(), gt.IsValid(), "Grant type %s should be invalid", gt)
	}
}

func (suite *OAuth2ConstantsTestSuite) TestResponseTypeConstants() {
	// Test response type constants
	assert.Equal(suite.T(), ResponseType("code"), ResponseTypeCode)
	assert.Equal(suite.T(), ResponseType("token"), ResponseTypeToken)
}

func (suite *OAuth2ConstantsTestSuite) TestResponseType_IsValid() {
	// Test valid response types
	assert.True(suite.T(), ResponseTypeCode.IsValid())

	// Test invalid response types
	invalidResponseTypes := []ResponseType{
		ResponseTypeToken, // Currently not supported according to the IsValid implementation
		"invalid_response",
		"",
		"CODE", // case sensitive
	}

	for _, rt := range invalidResponseTypes {
		assert.False(suite.T(), rt.IsValid(), "Response type %s should be invalid", rt)
	}
}

func (suite *OAuth2ConstantsTestSuite) TestTokenEndpointAuthMethodConstants() {
	// Test token endpoint auth method constants
	assert.Equal(suite.T(), TokenEndpointAuthMethod("client_secret_basic"), TokenEndpointAuthMethodClientSecretBasic)
	assert.Equal(suite.T(), TokenEndpointAuthMethod("client_secret_post"), TokenEndpointAuthMethodClientSecretPost)
	assert.Equal(suite.T(), TokenEndpointAuthMethod("none"), TokenEndpointAuthMethodNone)
}

func (suite *OAuth2ConstantsTestSuite) TestTokenEndpointAuthMethod_IsValid() {
	// Test valid token endpoint auth methods
	validAuthMethods := []TokenEndpointAuthMethod{
		TokenEndpointAuthMethodClientSecretBasic,
		TokenEndpointAuthMethodClientSecretPost,
		TokenEndpointAuthMethodNone,
	}

	for _, tam := range validAuthMethods {
		assert.True(suite.T(), tam.IsValid(), "Token endpoint auth method %s should be valid", tam)
	}

	// Test invalid token endpoint auth methods
	invalidAuthMethods := []TokenEndpointAuthMethod{
		"invalid_method",
		"",
		"CLIENT_SECRET_BASIC", // case sensitive
		"client_secret_jwt",   // not supported
	}

	for _, tam := range invalidAuthMethods {
		assert.False(suite.T(), tam.IsValid(), "Token endpoint auth method %s should be invalid", tam)
	}
}

func (suite *OAuth2ConstantsTestSuite) TestTokenTypeConstants() {
	// Test token type constants
	assert.Equal(suite.T(), "Bearer", TokenTypeBearer)
}

func (suite *OAuth2ConstantsTestSuite) TestErrorCodeConstants() {
	// Test OAuth2 error code constants
	assert.Equal(suite.T(), "invalid_request", ErrorInvalidRequest)
	assert.Equal(suite.T(), "invalid_client", ErrorInvalidClient)
	assert.Equal(suite.T(), "invalid_grant", ErrorInvalidGrant)
	assert.Equal(suite.T(), "unauthorized_client", ErrorUnauthorizedClient)
	assert.Equal(suite.T(), "unsupported_grant_type", ErrorUnsupportedGrantType)
	assert.Equal(suite.T(), "invalid_scope", ErrorInvalidScope)
	assert.Equal(suite.T(), "server_error", ErrorServerError)
	assert.Equal(suite.T(), "unsupported_response_type", ErrorUnsupportedResponseType)
	assert.Equal(suite.T(), "access_denied", ErrorAccessDenied)
}

func (suite *OAuth2ConstantsTestSuite) TestGrantTypeStringConversion() {
	// Test that grant type can be converted to string
	gt := GrantTypeAuthorizationCode
	assert.Equal(suite.T(), "authorization_code", string(gt))
}

func (suite *OAuth2ConstantsTestSuite) TestResponseTypeStringConversion() {
	// Test that response type can be converted to string
	rt := ResponseTypeCode
	assert.Equal(suite.T(), "code", string(rt))
}

func (suite *OAuth2ConstantsTestSuite) TestTokenEndpointAuthMethodStringConversion() {
	// Test that token endpoint auth method can be converted to string
	tam := TokenEndpointAuthMethodClientSecretBasic
	assert.Equal(suite.T(), "client_secret_basic", string(tam))
}

func (suite *OAuth2ConstantsTestSuite) TestConstantsAreNotEmpty() {
	// Verify that all constants are not empty strings
	constants := []string{
		RequestParamGrantType,
		RequestParamClientID,
		RequestParamClientSecret,
		RequestParamRedirectURI,
		RequestParamUsername,
		RequestParamPassword,
		RequestParamScope,
		RequestParamCode,
		RequestParamCodeVerifier,
		RequestParamRefreshToken,
		RequestParamResponseType,
		RequestParamState,
		RequestParamError,
		RequestParamErrorDescription,
		SessionDataKey,
		SessionDataKeyConsent,
		ShowInsecureWarning,
		AppID,
		Assertion,
		TypeInitialAuthorizationRequest,
		TypeAuthorizationResponseFromEngine,
		TypeConsentResponseFromUser,
		OAuth2TokenEndpoint,
		OAuth2AuthorizationEndpoint,
		OAuth2IntrospectionEndpoint,
		OAuth2RevokeEndpoint,
		OAuth2UserInfoEndpoint,
		OAuth2JWKSEndpoint,
		OAuth2LogoutEndpoint,
		TokenTypeBearer,
		ErrorInvalidRequest,
		ErrorInvalidClient,
		ErrorInvalidGrant,
		ErrorUnauthorizedClient,
		ErrorUnsupportedGrantType,
		ErrorInvalidScope,
		ErrorServerError,
		ErrorUnsupportedResponseType,
		ErrorAccessDenied,
	}

	for _, constant := range constants {
		assert.NotEmpty(suite.T(), constant, "Constant should not be empty")
	}
}

func (suite *OAuth2ConstantsTestSuite) TestEndpointsStartWithSlash() {
	// Verify that all endpoint constants start with "/"
	endpoints := []string{
		OAuth2TokenEndpoint,
		OAuth2AuthorizationEndpoint,
		OAuth2IntrospectionEndpoint,
		OAuth2RevokeEndpoint,
		OAuth2UserInfoEndpoint,
		OAuth2JWKSEndpoint,
		OAuth2LogoutEndpoint,
	}

	for _, endpoint := range endpoints {
		assert.True(suite.T(), len(endpoint) > 0 && endpoint[0] == '/', "Endpoint %s should start with /", endpoint)
	}
}

func (suite *OAuth2ConstantsTestSuite) TestCustomTypeValidation_EdgeCases() {
	// Test GrantType with empty string
	var emptyGT GrantType
	assert.False(suite.T(), emptyGT.IsValid())

	// Test ResponseType with empty string
	var emptyRT ResponseType
	assert.False(suite.T(), emptyRT.IsValid())

	// Test TokenEndpointAuthMethod with empty string
	var emptyTAM TokenEndpointAuthMethod
	assert.False(suite.T(), emptyTAM.IsValid())
}