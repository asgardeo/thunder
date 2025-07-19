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

package authz

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/authz/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
)

type AuthzValidatorTestSuite struct {
	suite.Suite
	validator AuthorizationValidatorInterface
}

func TestAuthzValidatorTestSuite(t *testing.T) {
	suite.Run(t, new(AuthzValidatorTestSuite))
}

func (suite *AuthzValidatorTestSuite) SetupTest() {
	suite.validator = NewAuthorizationValidator()
}

func (suite *AuthzValidatorTestSuite) TestValidateInitialAuthorizationRequest_ValidRequest() {
	// Create a valid OAuth application
	app := &appmodel.OAuthApplication{
		ClientID:          "test_client",
		ClientSecret:      "test_secret",
		RedirectURIs:      []string{"http://localhost:8080/callback"},
		AllowedGrantTypes: []string{constants.GrantTypeAuthorizationCode},
	}

	// Create a valid OAuth message
	msg := &model.OAuthMessage{
		RequestType: constants.TypeInitialAuthorizationRequest,
		RequestQueryParams: map[string]string{
			constants.ClientID:     "test_client",
			constants.ResponseType: constants.ResponseTypeCode,
			constants.RedirectURI:  "http://localhost:8080/callback",
		},
	}

	hasError, errorCode, errorDesc := suite.validator.validateInitialAuthorizationRequest(msg, app)

	assert.False(suite.T(), hasError)
	assert.Empty(suite.T(), errorCode)
	assert.Empty(suite.T(), errorDesc)
}

func (suite *AuthzValidatorTestSuite) TestValidateInitialAuthorizationRequest_MissingClientID() {
	app := &appmodel.OAuthApplication{
		ClientID:          "test_client",
		AllowedGrantTypes: []string{constants.GrantTypeAuthorizationCode},
		RedirectURIs:      []string{"http://localhost:8080/callback"},
	}

	msg := &model.OAuthMessage{
		RequestType: constants.TypeInitialAuthorizationRequest,
		RequestQueryParams: map[string]string{
			constants.ResponseType: constants.ResponseTypeCode,
			constants.RedirectURI:  "http://localhost:8080/callback",
		},
	}

	hasError, errorCode, errorDesc := suite.validator.validateInitialAuthorizationRequest(msg, app)

	assert.False(suite.T(), hasError) // Should be false because client validation happens before redirect
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, errorCode)
	assert.Equal(suite.T(), "Missing client_id parameter", errorDesc)
}

func (suite *AuthzValidatorTestSuite) TestValidateInitialAuthorizationRequest_GrantTypeNotAllowed() {
	app := &appmodel.OAuthApplication{
		ClientID:          "test_client",
		AllowedGrantTypes: []string{constants.GrantTypeClientCredentials}, // Not authorization code
		RedirectURIs:      []string{"http://localhost:8080/callback"},
	}

	msg := &model.OAuthMessage{
		RequestType: constants.TypeInitialAuthorizationRequest,
		RequestQueryParams: map[string]string{
			constants.ClientID:     "test_client",
			constants.ResponseType: constants.ResponseTypeCode,
			constants.RedirectURI:  "http://localhost:8080/callback",
		},
	}

	hasError, errorCode, errorDesc := suite.validator.validateInitialAuthorizationRequest(msg, app)

	assert.False(suite.T(), hasError) // Should be false because client validation happens before redirect
	assert.Equal(suite.T(), constants.ErrorUnsupportedGrantType, errorCode)
	assert.Equal(suite.T(), "Authorization code grant type is not allowed for the client", errorDesc)
}

func (suite *AuthzValidatorTestSuite) TestValidateInitialAuthorizationRequest_InvalidRedirectURI() {
	app := &appmodel.OAuthApplication{
		ClientID:          "test_client",
		AllowedGrantTypes: []string{constants.GrantTypeAuthorizationCode},
		RedirectURIs:      []string{"http://localhost:8080/callback"},
	}

	msg := &model.OAuthMessage{
		RequestType: constants.TypeInitialAuthorizationRequest,
		RequestQueryParams: map[string]string{
			constants.ClientID:     "test_client",
			constants.ResponseType: constants.ResponseTypeCode,
			constants.RedirectURI:  "http://malicious.com/callback", // Invalid redirect URI
		},
	}

	hasError, errorCode, errorDesc := suite.validator.validateInitialAuthorizationRequest(msg, app)

	assert.False(suite.T(), hasError) // Should be false because client validation happens before redirect
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, errorCode)
	assert.Equal(suite.T(), "Invalid redirect URI", errorDesc)
}

func (suite *AuthzValidatorTestSuite) TestValidateInitialAuthorizationRequest_MissingResponseType() {
	app := &appmodel.OAuthApplication{
		ClientID:          "test_client",
		AllowedGrantTypes: []string{constants.GrantTypeAuthorizationCode},
		RedirectURIs:      []string{"http://localhost:8080/callback"},
	}

	msg := &model.OAuthMessage{
		RequestType: constants.TypeInitialAuthorizationRequest,
		RequestQueryParams: map[string]string{
			constants.ClientID:    "test_client",
			constants.RedirectURI: "http://localhost:8080/callback",
			// Missing response_type
		},
	}

	hasError, errorCode, errorDesc := suite.validator.validateInitialAuthorizationRequest(msg, app)

	assert.True(suite.T(), hasError) // Should be true because this is redirectable error
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, errorCode)
	assert.Equal(suite.T(), "Missing response_type parameter", errorDesc)
}

func (suite *AuthzValidatorTestSuite) TestValidateInitialAuthorizationRequest_UnsupportedResponseType() {
	app := &appmodel.OAuthApplication{
		ClientID:          "test_client",
		AllowedGrantTypes: []string{constants.GrantTypeAuthorizationCode},
		RedirectURIs:      []string{"http://localhost:8080/callback"},
	}

	msg := &model.OAuthMessage{
		RequestType: constants.TypeInitialAuthorizationRequest,
		RequestQueryParams: map[string]string{
			constants.ClientID:     "test_client",
			constants.ResponseType: "token", // Unsupported response type
			constants.RedirectURI:  "http://localhost:8080/callback",
		},
	}

	hasError, errorCode, errorDesc := suite.validator.validateInitialAuthorizationRequest(msg, app)

	assert.True(suite.T(), hasError) // Should be true because this is redirectable error
	assert.Equal(suite.T(), constants.ErrorUnsupportedResponseType, errorCode)
	assert.Equal(suite.T(), "Unsupported response type", errorDesc)
}

func (suite *AuthzValidatorTestSuite) TestValidateInitialAuthorizationRequest_EmptyRedirectURI() {
	// Test case where redirect URI is empty but only one is registered
	app := &appmodel.OAuthApplication{
		ClientID:          "test_client",
		AllowedGrantTypes: []string{constants.GrantTypeAuthorizationCode},
		RedirectURIs:      []string{"http://localhost:8080/callback"},
	}

	msg := &model.OAuthMessage{
		RequestType: constants.TypeInitialAuthorizationRequest,
		RequestQueryParams: map[string]string{
			constants.ClientID:     "test_client",
			constants.ResponseType: constants.ResponseTypeCode,
			// No redirect URI specified
		},
	}

	hasError, errorCode, errorDesc := suite.validator.validateInitialAuthorizationRequest(msg, app)

	assert.False(suite.T(), hasError)
	assert.Empty(suite.T(), errorCode)
	assert.Empty(suite.T(), errorDesc)
}

func (suite *AuthzValidatorTestSuite) TestNewAuthorizationValidator() {
	validator := NewAuthorizationValidator()
	assert.NotNil(suite.T(), validator)
	assert.IsType(suite.T(), &AuthorizationValidator{}, validator)
}

func (suite *AuthzValidatorTestSuite) TestValidateInitialAuthorizationRequest_EmptyClientID() {
	msg := &model.OAuthMessage{
		RequestQueryParams: map[string]string{
			constants.ResponseType: constants.ResponseTypeCode,
			constants.RedirectURI:  "https://example.com/callback",
			constants.ClientID:     "", // Empty client ID
		},
	}

	app := &appmodel.OAuthApplication{
		ClientID:          "test_client",
		AllowedGrantTypes: []string{constants.GrantTypeAuthorizationCode},
		RedirectURIs:      []string{"https://example.com/callback"},
	}

	hasError, errorCode, errorDesc := suite.validator.validateInitialAuthorizationRequest(msg, app)

	assert.False(suite.T(), hasError)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, errorCode)
	assert.Equal(suite.T(), "Missing client_id parameter", errorDesc)
}

func (suite *AuthzValidatorTestSuite) TestValidateInitialAuthorizationRequest_EmptyResponseType() {
	msg := &model.OAuthMessage{
		RequestQueryParams: map[string]string{
			constants.ClientID:     "test_client",
			constants.RedirectURI:  "https://example.com/callback",
			constants.ResponseType: "", // Empty response type
		},
	}

	app := &appmodel.OAuthApplication{
		ClientID:          "test_client",
		AllowedGrantTypes: []string{constants.GrantTypeAuthorizationCode},
		RedirectURIs:      []string{"https://example.com/callback"},
	}

	hasError, errorCode, errorDesc := suite.validator.validateInitialAuthorizationRequest(msg, app)

	assert.True(suite.T(), hasError)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, errorCode)
	assert.Equal(suite.T(), "Missing response_type parameter", errorDesc)
}
