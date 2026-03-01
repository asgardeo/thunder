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

package authz

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	appmodel "github.com/asgardeo/thunder/internal/application/model"

	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
)

type AuthorizationValidatorTestSuite struct {
	suite.Suite
	validator AuthorizationValidatorInterface
	oauthApp  *appmodel.OAuthAppConfigProcessedDTO
}

func TestAuthorizationValidatorTestSuite(t *testing.T) {
	suite.Run(t, new(AuthorizationValidatorTestSuite))
}

func (suite *AuthorizationValidatorTestSuite) SetupTest() {
	suite.validator = newAuthorizationValidator()

	suite.oauthApp = &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:                "test-client-id",
		HashedClientSecret:      "hashed-secret",
		RedirectURIs:            []string{"https://client.example.com/callback"},
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
		ResponseTypes:           []constants.ResponseType{constants.ResponseTypeCode},
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
	}
}

func (suite *AuthorizationValidatorTestSuite) TestnewAuthorizationValidator() {
	validator := newAuthorizationValidator()
	assert.NotNil(suite.T(), validator)
	assert.Implements(suite.T(), (*AuthorizationValidatorInterface)(nil), validator)
}

func (suite *AuthorizationValidatorTestSuite) TestValidateInitialAuthorizationRequest_Success() {
	msg := &OAuthMessage{
		RequestQueryParams: map[string]string{
			constants.RequestParamClientID:     "test-client-id",
			constants.RequestParamRedirectURI:  "https://client.example.com/callback",
			constants.RequestParamResponseType: string(constants.ResponseTypeCode),
		},
	}

	sendErrorToApp, errorCode, errorMessage := suite.validator.validateInitialAuthorizationRequest(
		msg, suite.oauthApp)

	assert.False(suite.T(), sendErrorToApp)
	assert.Empty(suite.T(), errorCode)
	assert.Empty(suite.T(), errorMessage)
}

func (suite *AuthorizationValidatorTestSuite) TestValidateInitialAuthorizationRequest_MissingClientID() {
	msg := &OAuthMessage{
		RequestQueryParams: map[string]string{
			constants.RequestParamRedirectURI:  "https://client.example.com/callback",
			constants.RequestParamResponseType: string(constants.ResponseTypeCode),
		},
	}

	sendErrorToApp, errorCode, errorMessage := suite.validator.validateInitialAuthorizationRequest(
		msg, suite.oauthApp)

	assert.False(suite.T(), sendErrorToApp)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, errorCode)
	assert.Equal(suite.T(), "Missing client_id parameter", errorMessage)
}

func (suite *AuthorizationValidatorTestSuite) TestValidateInitialAuthorizationRequest_InvalidRedirectURI() {
	msg := &OAuthMessage{
		RequestQueryParams: map[string]string{
			constants.RequestParamClientID:     "test-client-id",
			constants.RequestParamRedirectURI:  "https://malicious.example.com/callback", // not in allowed list
			constants.RequestParamResponseType: string(constants.ResponseTypeCode),
		},
	}

	sendErrorToApp, errorCode, errorMessage := suite.validator.validateInitialAuthorizationRequest(
		msg, suite.oauthApp)

	assert.False(suite.T(), sendErrorToApp)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, errorCode)
	assert.Equal(suite.T(), "Invalid redirect URI", errorMessage)
}

func (suite *AuthorizationValidatorTestSuite) TestValidateAuthzRequest_CodeGrantNotAllowed() {
	// Create an app that doesn't allow authorization code grant type
	restrictedApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:                "test-client-id",
		HashedClientSecret:      "hashed-secret",
		RedirectURIs:            []string{"https://client.example.com/callback"},
		GrantTypes:              []constants.GrantType{constants.GrantTypeClientCredentials}, // no auth code
		ResponseTypes:           []constants.ResponseType{constants.ResponseTypeCode},
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
	}

	msg := &OAuthMessage{
		RequestQueryParams: map[string]string{
			constants.RequestParamClientID:     "test-client-id",
			constants.RequestParamRedirectURI:  "https://client.example.com/callback",
			constants.RequestParamResponseType: string(constants.ResponseTypeCode),
		},
	}

	sendErrorToApp, errorCode, errorMessage := suite.validator.validateInitialAuthorizationRequest(
		msg, restrictedApp)

	assert.True(suite.T(), sendErrorToApp)
	assert.Equal(suite.T(), constants.ErrorUnsupportedGrantType, errorCode)
	assert.Equal(suite.T(), "Authorization code grant type is not allowed for the client", errorMessage)
}

func (suite *AuthorizationValidatorTestSuite) TestValidateInitialAuthorizationRequest_MissingResponseType() {
	msg := &OAuthMessage{
		RequestQueryParams: map[string]string{
			constants.RequestParamClientID:    "test-client-id",
			constants.RequestParamRedirectURI: "https://client.example.com/callback",
		},
	}

	sendErrorToApp, errorCode, errorMessage := suite.validator.validateInitialAuthorizationRequest(
		msg, suite.oauthApp)

	assert.True(suite.T(), sendErrorToApp)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, errorCode)
	assert.Equal(suite.T(), "Missing response_type parameter", errorMessage)
}

func (suite *AuthorizationValidatorTestSuite) TestValidateInitialAuthorizationRequest_UnsupportedResponseType() {
	// Create an app that doesn't support "code" response type
	restrictedApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:                "test-client-id",
		HashedClientSecret:      "hashed-secret",
		RedirectURIs:            []string{"https://client.example.com/callback"},
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
		ResponseTypes:           []constants.ResponseType{}, // no response types allowed
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
	}

	msg := &OAuthMessage{
		RequestQueryParams: map[string]string{
			constants.RequestParamClientID:     "test-client-id",
			constants.RequestParamRedirectURI:  "https://client.example.com/callback",
			constants.RequestParamResponseType: string(constants.ResponseTypeCode),
		},
	}

	sendErrorToApp, errorCode, errorMessage := suite.validator.validateInitialAuthorizationRequest(
		msg, restrictedApp)

	assert.True(suite.T(), sendErrorToApp)
	assert.Equal(suite.T(), constants.ErrorUnsupportedResponseType, errorCode)
	assert.Equal(suite.T(), "Unsupported response type", errorMessage)
}

func (suite *AuthorizationValidatorTestSuite) TestValidateInitialAuthorizationRequest_EmptyRedirectURI() {
	msg := &OAuthMessage{
		RequestQueryParams: map[string]string{
			constants.RequestParamClientID:     "test-client-id",
			constants.RequestParamRedirectURI:  "", // empty redirect URI should be OK if app has only one registered
			constants.RequestParamResponseType: string(constants.ResponseTypeCode),
		},
	}

	sendErrorToApp, errorCode, errorMessage := suite.validator.validateInitialAuthorizationRequest(
		msg, suite.oauthApp)

	assert.False(suite.T(), sendErrorToApp)
	assert.Empty(suite.T(), errorCode)
	assert.Empty(suite.T(), errorMessage)
}

// Resource Parameter Validation Tests

func (suite *AuthorizationValidatorTestSuite) TestValidateInitialAuthorizationRequest_ValidResource() {
	msg := &OAuthMessage{
		RequestQueryParams: map[string]string{
			constants.RequestParamClientID:     "test-client-id",
			constants.RequestParamRedirectURI:  "https://client.example.com/callback",
			constants.RequestParamResponseType: string(constants.ResponseTypeCode),
			constants.RequestParamResource:     "https://api.example.com/resource",
		},
	}

	sendErrorToApp, errorCode, errorMessage := suite.validator.validateInitialAuthorizationRequest(
		msg, suite.oauthApp)

	assert.False(suite.T(), sendErrorToApp)
	assert.Empty(suite.T(), errorCode)
	assert.Empty(suite.T(), errorMessage)
}

func (suite *AuthorizationValidatorTestSuite) TestValidateInitialAuthorizationRequest_ValidMCPServerResource() {
	msg := &OAuthMessage{
		RequestQueryParams: map[string]string{
			constants.RequestParamClientID:     "test-client-id",
			constants.RequestParamRedirectURI:  "https://client.example.com/callback",
			constants.RequestParamResponseType: string(constants.ResponseTypeCode),
			constants.RequestParamResource:     "https://mcp.example.com/mcp",
		},
	}

	sendErrorToApp, errorCode, errorMessage := suite.validator.validateInitialAuthorizationRequest(
		msg, suite.oauthApp)

	assert.False(suite.T(), sendErrorToApp)
	assert.Empty(suite.T(), errorCode)
	assert.Empty(suite.T(), errorMessage)
}

func (suite *AuthorizationValidatorTestSuite) TestValidateInitialAuthorizationRequest_ValidResourceWithPort() {
	msg := &OAuthMessage{
		RequestQueryParams: map[string]string{
			constants.RequestParamClientID:     "test-client-id",
			constants.RequestParamRedirectURI:  "https://client.example.com/callback",
			constants.RequestParamResponseType: string(constants.ResponseTypeCode),
			constants.RequestParamResource:     "https://mcp.example.com:8443",
		},
	}

	sendErrorToApp, errorCode, errorMessage := suite.validator.validateInitialAuthorizationRequest(
		msg, suite.oauthApp)

	assert.False(suite.T(), sendErrorToApp)
	assert.Empty(suite.T(), errorCode)
	assert.Empty(suite.T(), errorMessage)
}

func (suite *AuthorizationValidatorTestSuite) TestValidateInitialAuthorizationRequest_EmptyResource() {
	msg := &OAuthMessage{
		RequestQueryParams: map[string]string{
			constants.RequestParamClientID:     "test-client-id",
			constants.RequestParamRedirectURI:  "https://client.example.com/callback",
			constants.RequestParamResponseType: string(constants.ResponseTypeCode),
			constants.RequestParamResource:     "",
		},
	}

	sendErrorToApp, errorCode, errorMessage := suite.validator.validateInitialAuthorizationRequest(
		msg, suite.oauthApp)

	assert.False(suite.T(), sendErrorToApp)
	assert.Empty(suite.T(), errorCode)
	assert.Empty(suite.T(), errorMessage)
}

func (suite *AuthorizationValidatorTestSuite) TestValidateInitialAuthorizationRequest_ResourceMissingScheme() {
	msg := &OAuthMessage{
		RequestQueryParams: map[string]string{
			constants.RequestParamClientID:     "test-client-id",
			constants.RequestParamRedirectURI:  "https://client.example.com/callback",
			constants.RequestParamResponseType: string(constants.ResponseTypeCode),
			constants.RequestParamResource:     "api.example.com/resource",
		},
	}

	sendErrorToApp, errorCode, errorMessage := suite.validator.validateInitialAuthorizationRequest(
		msg, suite.oauthApp)

	assert.True(suite.T(), sendErrorToApp)
	assert.Equal(suite.T(), constants.ErrorInvalidTarget, errorCode)
	assert.Contains(suite.T(), errorMessage, "absolute URI with a scheme")
}

func (suite *AuthorizationValidatorTestSuite) TestValidateInitialAuthorizationRequest_ResourceWithFragment() {
	msg := &OAuthMessage{
		RequestQueryParams: map[string]string{
			constants.RequestParamClientID:     "test-client-id",
			constants.RequestParamRedirectURI:  "https://client.example.com/callback",
			constants.RequestParamResponseType: string(constants.ResponseTypeCode),
			constants.RequestParamResource:     "https://api.example.com/resource#fragment",
		},
	}

	sendErrorToApp, errorCode, errorMessage := suite.validator.validateInitialAuthorizationRequest(
		msg, suite.oauthApp)

	assert.True(suite.T(), sendErrorToApp)
	assert.Equal(suite.T(), constants.ErrorInvalidTarget, errorCode)
	assert.Contains(suite.T(), errorMessage, "fragment component")
}

func (suite *AuthorizationValidatorTestSuite) TestValidateInitialAuthorizationRequest_ResourceRelativeURI() {
	msg := &OAuthMessage{
		RequestQueryParams: map[string]string{
			constants.RequestParamClientID:     "test-client-id",
			constants.RequestParamRedirectURI:  "https://client.example.com/callback",
			constants.RequestParamResponseType: string(constants.ResponseTypeCode),
			constants.RequestParamResource:     "/api/resource",
		},
	}

	sendErrorToApp, errorCode, errorMessage := suite.validator.validateInitialAuthorizationRequest(
		msg, suite.oauthApp)

	assert.True(suite.T(), sendErrorToApp)
	assert.Equal(suite.T(), constants.ErrorInvalidTarget, errorCode)
	assert.Contains(suite.T(), errorMessage, "absolute URI with a scheme")
}

func (suite *AuthorizationValidatorTestSuite) TestValidateInitialAuthorizationRequest_ResourceInvalidURI() {
	msg := &OAuthMessage{
		RequestQueryParams: map[string]string{
			constants.RequestParamClientID:     "test-client-id",
			constants.RequestParamRedirectURI:  "https://client.example.com/callback",
			constants.RequestParamResponseType: string(constants.ResponseTypeCode),
			constants.RequestParamResource:     "not a valid uri format",
		},
	}

	sendErrorToApp, errorCode, errorMessage := suite.validator.validateInitialAuthorizationRequest(
		msg, suite.oauthApp)

	assert.True(suite.T(), sendErrorToApp)
	assert.Equal(suite.T(), constants.ErrorInvalidTarget, errorCode)
	assert.Contains(suite.T(), errorMessage, "absolute URI")
}

func (suite *AuthorizationValidatorTestSuite) TestValidateInitialAuthorizationRequest_ResourceParameterWithQuery() {
	// Test resource parameter with query component (should be valid per RFC 8707)
	msg := &OAuthMessage{
		RequestQueryParams: map[string]string{
			constants.RequestParamClientID:     "test-client-id",
			constants.RequestParamRedirectURI:  "https://client.example.com/callback",
			constants.RequestParamResponseType: string(constants.ResponseTypeCode),
			constants.RequestParamResource:     "https://api.example.com/resource?param=value",
		},
	}

	sendErrorToApp, errorCode, errorMessage := suite.validator.validateInitialAuthorizationRequest(
		msg, suite.oauthApp)

	assert.False(suite.T(), sendErrorToApp)
	assert.Empty(suite.T(), errorCode)
	assert.Empty(suite.T(), errorMessage)
}

func (suite *AuthorizationValidatorTestSuite) TestValidateAuthzReq_PKCERequired_MissingCodeChallenge() {
	// Create an app that requires PKCE
	pkceApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:                "test-client-id",
		HashedClientSecret:      "hashed-secret",
		RedirectURIs:            []string{"https://client.example.com/callback"},
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
		ResponseTypes:           []constants.ResponseType{constants.ResponseTypeCode},
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
		PKCERequired:            true,
	}

	msg := &OAuthMessage{
		RequestQueryParams: map[string]string{
			constants.RequestParamClientID:     "test-client-id",
			constants.RequestParamRedirectURI:  "https://client.example.com/callback",
			constants.RequestParamResponseType: string(constants.ResponseTypeCode),
			// Missing code_challenge
		},
	}

	sendErrorToApp, errorCode, errorMessage := suite.validator.validateInitialAuthorizationRequest(
		msg, pkceApp)

	assert.True(suite.T(), sendErrorToApp)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, errorCode)
	assert.Equal(suite.T(), "code_challenge is required for this application", errorMessage)
}

func (suite *AuthorizationValidatorTestSuite) TestValidateAuthzReq_PKCERequired_InvalidCodeChallenge() {
	// Create an app that requires PKCE
	pkceApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:                "test-client-id",
		HashedClientSecret:      "hashed-secret",
		RedirectURIs:            []string{"https://client.example.com/callback"},
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
		ResponseTypes:           []constants.ResponseType{constants.ResponseTypeCode},
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
		PKCERequired:            true,
	}

	msg := &OAuthMessage{
		RequestQueryParams: map[string]string{
			constants.RequestParamClientID:            "test-client-id",
			constants.RequestParamRedirectURI:         "https://client.example.com/callback",
			constants.RequestParamResponseType:        string(constants.ResponseTypeCode),
			constants.RequestParamCodeChallenge:       "invalid-challenge",
			constants.RequestParamCodeChallengeMethod: "plain", // Not supported per OAuth 2.0 Security BCP
		},
	}

	sendErrorToApp, errorCode, errorMessage := suite.validator.validateInitialAuthorizationRequest(
		msg, pkceApp)

	assert.True(suite.T(), sendErrorToApp)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, errorCode)
	assert.Equal(suite.T(), "Invalid PKCE parameters", errorMessage)
}

func (suite *AuthorizationValidatorTestSuite) TestValidateInitialAuthorizationRequest_PKCERequired_ValidPKCE() {
	// Create an app that requires PKCE
	pkceApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:                "test-client-id",
		HashedClientSecret:      "hashed-secret",
		RedirectURIs:            []string{"https://client.example.com/callback"},
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
		ResponseTypes:           []constants.ResponseType{constants.ResponseTypeCode},
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
		PKCERequired:            true,
	}

	// Use a valid S256 code challenge (base64url encoded SHA256 hash)
	// This is a valid format for testing
	msg := &OAuthMessage{
		RequestQueryParams: map[string]string{
			constants.RequestParamClientID:            "test-client-id",
			constants.RequestParamRedirectURI:         "https://client.example.com/callback",
			constants.RequestParamResponseType:        string(constants.ResponseTypeCode),
			constants.RequestParamCodeChallenge:       "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
			constants.RequestParamCodeChallengeMethod: "S256",
		},
	}

	sendErrorToApp, errorCode, errorMessage := suite.validator.validateInitialAuthorizationRequest(
		msg, pkceApp)

	assert.False(suite.T(), sendErrorToApp)
	assert.Empty(suite.T(), errorCode)
	assert.Empty(suite.T(), errorMessage)
}

func (suite *AuthorizationValidatorTestSuite) TestValidateInitialAuthorizationRequest_PKCENotRequired() {
	// Create an app that doesn't require PKCE
	nonPKCEApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:                "test-client-id",
		HashedClientSecret:      "hashed-secret",
		RedirectURIs:            []string{"https://client.example.com/callback"},
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
		ResponseTypes:           []constants.ResponseType{constants.ResponseTypeCode},
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
		PKCERequired:            false,
	}

	msg := &OAuthMessage{
		RequestQueryParams: map[string]string{
			constants.RequestParamClientID:     "test-client-id",
			constants.RequestParamRedirectURI:  "https://client.example.com/callback",
			constants.RequestParamResponseType: string(constants.ResponseTypeCode),
			// No PKCE parameters - should be OK since PKCE is not required
		},
	}

	sendErrorToApp, errorCode, errorMessage := suite.validator.validateInitialAuthorizationRequest(
		msg, nonPKCEApp)

	assert.False(suite.T(), sendErrorToApp)
	assert.Empty(suite.T(), errorCode)
	assert.Empty(suite.T(), errorMessage)
}
