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

package granthandlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/model"
	"github.com/asgardeo/thunder/internal/system/log"
)

type RefreshTokenGrantHandlerTestSuite struct {
	suite.Suite
	handler *RefreshTokenGrantHandler
}

type ClientCredentialsGrantHandlerTestSuite struct {
	suite.Suite
	handler *ClientCredentialsGrantHandler
}

type AuthorizationCodeGrantHandlerTestSuite struct {
	suite.Suite
	handler *AuthorizationCodeGrantHandler
}

func TestRefreshTokenGrantHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(RefreshTokenGrantHandlerTestSuite))
}

func TestClientCredentialsGrantHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ClientCredentialsGrantHandlerTestSuite))
}

func TestAuthorizationCodeGrantHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(AuthorizationCodeGrantHandlerTestSuite))
}

func (suite *RefreshTokenGrantHandlerTestSuite) SetupTest() {
	suite.handler = &RefreshTokenGrantHandler{}
}

func (suite *ClientCredentialsGrantHandlerTestSuite) SetupTest() {
	suite.handler = &ClientCredentialsGrantHandler{}
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) SetupTest() {
	suite.handler = &AuthorizationCodeGrantHandler{}
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestValidateGrant_ValidRequest() {
	tokenRequest := &model.TokenRequest{
		GrantType:    constants.GrantTypeRefreshToken,
		ClientID:     "test_client",
		ClientSecret: "test_secret",
		RefreshToken: "valid_refresh_token",
	}

	oauthApp := &appmodel.OAuthApplication{
		ClientID:     "test_client",
		ClientSecret: "test_secret",
	}

	errorResponse := suite.handler.ValidateGrant(tokenRequest, oauthApp)

	assert.Nil(suite.T(), errorResponse)
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestValidateGrant_InvalidGrantType() {
	tokenRequest := &model.TokenRequest{
		GrantType:    constants.GrantTypeAuthorizationCode, // Wrong grant type
		ClientID:     "test_client",
		ClientSecret: "test_secret",
		RefreshToken: "valid_refresh_token",
	}

	oauthApp := &appmodel.OAuthApplication{
		ClientID:     "test_client",
		ClientSecret: "test_secret",
	}

	errorResponse := suite.handler.ValidateGrant(tokenRequest, oauthApp)

	assert.NotNil(suite.T(), errorResponse)
	assert.Equal(suite.T(), constants.ErrorUnsupportedGrantType, errorResponse.Error)
	assert.Equal(suite.T(), "Unsupported grant type", errorResponse.ErrorDescription)
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestValidateGrant_MissingRefreshToken() {
	tokenRequest := &model.TokenRequest{
		GrantType:    constants.GrantTypeRefreshToken,
		ClientID:     "test_client",
		ClientSecret: "test_secret",
		// Missing RefreshToken
	}

	oauthApp := &appmodel.OAuthApplication{
		ClientID:     "test_client",
		ClientSecret: "test_secret",
	}

	errorResponse := suite.handler.ValidateGrant(tokenRequest, oauthApp)

	assert.NotNil(suite.T(), errorResponse)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, errorResponse.Error)
	assert.Equal(suite.T(), "Refresh token is required", errorResponse.ErrorDescription)
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestValidateGrant_MissingClientID() {
	tokenRequest := &model.TokenRequest{
		GrantType:    constants.GrantTypeRefreshToken,
		ClientSecret: "test_secret",
		RefreshToken: "valid_refresh_token",
		// Missing ClientID
	}

	oauthApp := &appmodel.OAuthApplication{
		ClientID:     "test_client",
		ClientSecret: "test_secret",
	}

	errorResponse := suite.handler.ValidateGrant(tokenRequest, oauthApp)

	assert.NotNil(suite.T(), errorResponse)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, errorResponse.Error)
	assert.Equal(suite.T(), "Client ID is required", errorResponse.ErrorDescription)
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestValidateGrant_InvalidClientCredentials() {
	tokenRequest := &model.TokenRequest{
		GrantType:    constants.GrantTypeRefreshToken,
		ClientID:     "wrong_client",
		ClientSecret: "wrong_secret",
		RefreshToken: "valid_refresh_token",
	}

	oauthApp := &appmodel.OAuthApplication{
		ClientID:     "test_client",
		ClientSecret: "test_secret",
	}

	errorResponse := suite.handler.ValidateGrant(tokenRequest, oauthApp)

	assert.NotNil(suite.T(), errorResponse)
	assert.Equal(suite.T(), constants.ErrorInvalidClient, errorResponse.Error)
	assert.Equal(suite.T(), "Invalid client credentials", errorResponse.ErrorDescription)
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestValidateGrant_MismatchedClientID() {
	tokenRequest := &model.TokenRequest{
		GrantType:    constants.GrantTypeRefreshToken,
		ClientID:     "different_client",
		ClientSecret: "test_secret", // Secret matches but ID doesn't
		RefreshToken: "valid_refresh_token",
	}

	oauthApp := &appmodel.OAuthApplication{
		ClientID:     "test_client",
		ClientSecret: "test_secret",
	}

	errorResponse := suite.handler.ValidateGrant(tokenRequest, oauthApp)

	assert.NotNil(suite.T(), errorResponse)
	assert.Equal(suite.T(), constants.ErrorInvalidClient, errorResponse.Error)
	assert.Equal(suite.T(), "Invalid client credentials", errorResponse.ErrorDescription)
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestValidateGrant_MismatchedClientSecret() {
	tokenRequest := &model.TokenRequest{
		GrantType:    constants.GrantTypeRefreshToken,
		ClientID:     "test_client",
		ClientSecret: "wrong_secret", // Secret doesn't match
		RefreshToken: "valid_refresh_token",
	}

	oauthApp := &appmodel.OAuthApplication{
		ClientID:     "test_client",
		ClientSecret: "test_secret",
	}

	errorResponse := suite.handler.ValidateGrant(tokenRequest, oauthApp)

	assert.NotNil(suite.T(), errorResponse)
	assert.Equal(suite.T(), constants.ErrorInvalidClient, errorResponse.Error)
	assert.Equal(suite.T(), "Invalid client credentials", errorResponse.ErrorDescription)
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestValidateGrant_EmptyStrings() {
	tokenRequest := &model.TokenRequest{
		GrantType:    constants.GrantTypeRefreshToken,
		ClientID:     "",
		ClientSecret: "",
		RefreshToken: "",
	}

	oauthApp := &appmodel.OAuthApplication{
		ClientID:     "test_client",
		ClientSecret: "test_secret",
	}

	errorResponse := suite.handler.ValidateGrant(tokenRequest, oauthApp)

	// Should fail on refresh token validation first
	assert.NotNil(suite.T(), errorResponse)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, errorResponse.Error)
	assert.Equal(suite.T(), "Refresh token is required", errorResponse.ErrorDescription)
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestValidateTimeClaim_ValidClaim() {
	claims := map[string]interface{}{
		"iat": float64(1609459200), // 2021-01-01 00:00:00 UTC
	}

	logger := log.GetLogger()

	// Create a comparison function that should return false (no error)
	cmp := func(now, claim int64) bool {
		return claim > now // This would be true if claim is in the future, indicating an error
	}

	errorResponse := suite.handler.validateTimeClaim(
		claims, "iat", cmp, constants.ErrorInvalidRequest, "Test error", logger,
	)

	assert.Nil(suite.T(), errorResponse)
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestValidateTimeClaim_MissingClaim() {
	claims := map[string]interface{}{} // Missing the expected claim

	logger := log.GetLogger()

	cmp := func(now, claim int64) bool {
		return false
	}

	errorResponse := suite.handler.validateTimeClaim(
		claims, "iat", cmp, constants.ErrorInvalidRequest, "Test error", logger,
	)

	assert.NotNil(suite.T(), errorResponse) // Missing claims should cause errors based on the code
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, errorResponse.Error)
	assert.Equal(suite.T(), "Invalid refresh token", errorResponse.ErrorDescription)
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestValidateTimeClaim_InvalidClaimType() {
	claims := map[string]interface{}{
		"iat": "not_a_number", // Invalid type
	}

	logger := log.GetLogger()

	cmp := func(now, claim int64) bool {
		return false
	}

	errorResponse := suite.handler.validateTimeClaim(
		claims, "iat", cmp, constants.ErrorInvalidRequest, "Test error", logger,
	)

	assert.NotNil(suite.T(), errorResponse)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, errorResponse.Error)
	assert.Equal(suite.T(), "Invalid refresh token", errorResponse.ErrorDescription)
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestValidateGrant_EmptyRefreshToken() {
	tokenRequest := &model.TokenRequest{
		GrantType:    constants.GrantTypeRefreshToken,
		ClientID:     "test_client",
		ClientSecret: "test_secret",
		RefreshToken: "",
	}

	oauthApp := &appmodel.OAuthApplication{
		ClientID:     "test_client",
		ClientSecret: "test_secret",
	}

	errorResponse := suite.handler.ValidateGrant(tokenRequest, oauthApp)

	assert.NotNil(suite.T(), errorResponse)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, errorResponse.Error)
	assert.Equal(suite.T(), "Refresh token is required", errorResponse.ErrorDescription)
}

func (suite *RefreshTokenGrantHandlerTestSuite) TestValidateTimeClaim_ComparisonFails() {
	claims := map[string]interface{}{
		"iat": float64(1609459200),
	}

	logger := log.GetLogger()

	// Create a comparison function that will return true (indicating error)
	cmp := func(now, claim int64) bool {
		return true // Always return true to simulate error condition
	}

	errorResponse := suite.handler.validateTimeClaim(
		claims, "iat", cmp, constants.ErrorInvalidRequest, "Test error", logger,
	)

	assert.NotNil(suite.T(), errorResponse)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, errorResponse.Error)
	assert.Equal(suite.T(), "Test error", errorResponse.ErrorDescription)
}

// Client Credentials Grant Handler Tests
func (suite *ClientCredentialsGrantHandlerTestSuite) TestValidateGrant_ValidRequest() {
	tokenRequest := &model.TokenRequest{
		GrantType:    constants.GrantTypeClientCredentials,
		ClientID:     "test_client",
		ClientSecret: "test_secret",
	}

	oauthApp := &appmodel.OAuthApplication{
		ClientID:     "test_client",
		ClientSecret: "test_secret",
	}

	errorResponse := suite.handler.ValidateGrant(tokenRequest, oauthApp)

	assert.Nil(suite.T(), errorResponse)
}

func (suite *ClientCredentialsGrantHandlerTestSuite) TestValidateGrant_InvalidGrantType() {
	tokenRequest := &model.TokenRequest{
		GrantType:    constants.GrantTypeAuthorizationCode, // Wrong grant type
		ClientID:     "test_client",
		ClientSecret: "test_secret",
	}

	oauthApp := &appmodel.OAuthApplication{
		ClientID:     "test_client",
		ClientSecret: "test_secret",
	}

	errorResponse := suite.handler.ValidateGrant(tokenRequest, oauthApp)

	assert.NotNil(suite.T(), errorResponse)
	assert.Equal(suite.T(), constants.ErrorUnsupportedGrantType, errorResponse.Error)
	assert.Equal(suite.T(), "Unsupported grant type", errorResponse.ErrorDescription)
}

func (suite *ClientCredentialsGrantHandlerTestSuite) TestValidateGrant_MissingClientID() {
	tokenRequest := &model.TokenRequest{
		GrantType:    constants.GrantTypeClientCredentials,
		ClientSecret: "test_secret",
		// Missing ClientID
	}

	oauthApp := &appmodel.OAuthApplication{
		ClientID:     "test_client",
		ClientSecret: "test_secret",
	}

	errorResponse := suite.handler.ValidateGrant(tokenRequest, oauthApp)

	assert.NotNil(suite.T(), errorResponse)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, errorResponse.Error)
	assert.Equal(suite.T(), "Client Id and secret are required", errorResponse.ErrorDescription)
}

func (suite *ClientCredentialsGrantHandlerTestSuite) TestValidateGrant_MissingClientSecret() {
	tokenRequest := &model.TokenRequest{
		GrantType: constants.GrantTypeClientCredentials,
		ClientID:  "test_client",
		// Missing ClientSecret
	}

	oauthApp := &appmodel.OAuthApplication{
		ClientID:     "test_client",
		ClientSecret: "test_secret",
	}

	errorResponse := suite.handler.ValidateGrant(tokenRequest, oauthApp)

	assert.NotNil(suite.T(), errorResponse)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, errorResponse.Error)
	assert.Equal(suite.T(), "Client Id and secret are required", errorResponse.ErrorDescription)
}

func (suite *ClientCredentialsGrantHandlerTestSuite) TestValidateGrant_InvalidClientCredentials() {
	tokenRequest := &model.TokenRequest{
		GrantType:    constants.GrantTypeClientCredentials,
		ClientID:     "wrong_client",
		ClientSecret: "wrong_secret",
	}

	oauthApp := &appmodel.OAuthApplication{
		ClientID:     "test_client",
		ClientSecret: "test_secret",
	}

	errorResponse := suite.handler.ValidateGrant(tokenRequest, oauthApp)

	assert.NotNil(suite.T(), errorResponse)
	assert.Equal(suite.T(), constants.ErrorInvalidClient, errorResponse.Error)
	assert.Equal(suite.T(), "Invalid client credentials", errorResponse.ErrorDescription)
}

func (suite *ClientCredentialsGrantHandlerTestSuite) TestValidateGrant_MismatchedClientID() {
	tokenRequest := &model.TokenRequest{
		GrantType:    constants.GrantTypeClientCredentials,
		ClientID:     "different_client",
		ClientSecret: "test_secret", // Secret matches but ID doesn't
	}

	oauthApp := &appmodel.OAuthApplication{
		ClientID:     "test_client",
		ClientSecret: "test_secret",
	}

	errorResponse := suite.handler.ValidateGrant(tokenRequest, oauthApp)

	assert.NotNil(suite.T(), errorResponse)
	assert.Equal(suite.T(), constants.ErrorInvalidClient, errorResponse.Error)
	assert.Equal(suite.T(), "Invalid client credentials", errorResponse.ErrorDescription)
}

func (suite *ClientCredentialsGrantHandlerTestSuite) TestValidateGrant_MismatchedClientSecret() {
	tokenRequest := &model.TokenRequest{
		GrantType:    constants.GrantTypeClientCredentials,
		ClientID:     "test_client",
		ClientSecret: "wrong_secret", // Secret doesn't match
	}

	oauthApp := &appmodel.OAuthApplication{
		ClientID:     "test_client",
		ClientSecret: "test_secret",
	}

	errorResponse := suite.handler.ValidateGrant(tokenRequest, oauthApp)

	assert.NotNil(suite.T(), errorResponse)
	assert.Equal(suite.T(), constants.ErrorInvalidClient, errorResponse.Error)
	assert.Equal(suite.T(), "Invalid client credentials", errorResponse.ErrorDescription)
}

// Authorization Code Grant Handler Tests
// Note: The ValidateGrant method for authorization code requires database interaction
// for code validation, so we focus on testing the basic validation logic that
// doesn't require database dependencies.

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestValidateGrant_MissingGrantType() {
	tokenRequest := &model.TokenRequest{
		// Missing GrantType
		ClientID:     "test_client",
		ClientSecret: "test_secret",
		Code:         "valid_auth_code",
		RedirectURI:  "http://localhost:8080/callback",
	}

	oauthApp := &appmodel.OAuthApplication{
		ClientID:     "test_client",
		ClientSecret: "test_secret",
	}

	errorResponse := suite.handler.ValidateGrant(tokenRequest, oauthApp)

	assert.NotNil(suite.T(), errorResponse)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, errorResponse.Error)
	assert.Equal(suite.T(), "Missing grant type", errorResponse.ErrorDescription)
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestValidateGrant_InvalidGrantType() {
	tokenRequest := &model.TokenRequest{
		GrantType:    constants.GrantTypeClientCredentials, // Wrong grant type
		ClientID:     "test_client",
		ClientSecret: "test_secret",
		Code:         "valid_auth_code",
		RedirectURI:  "http://localhost:8080/callback",
	}

	oauthApp := &appmodel.OAuthApplication{
		ClientID:     "test_client",
		ClientSecret: "test_secret",
	}

	errorResponse := suite.handler.ValidateGrant(tokenRequest, oauthApp)

	assert.NotNil(suite.T(), errorResponse)
	assert.Equal(suite.T(), constants.ErrorUnsupportedGrantType, errorResponse.Error)
	assert.Equal(suite.T(), "Unsupported grant type", errorResponse.ErrorDescription)
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestValidateGrant_MissingAuthorizationCode() {
	tokenRequest := &model.TokenRequest{
		GrantType:    constants.GrantTypeAuthorizationCode,
		ClientID:     "test_client",
		ClientSecret: "test_secret",
		// Missing Code
		RedirectURI: "http://localhost:8080/callback",
	}

	oauthApp := &appmodel.OAuthApplication{
		ClientID:     "test_client",
		ClientSecret: "test_secret",
	}

	errorResponse := suite.handler.ValidateGrant(tokenRequest, oauthApp)

	assert.NotNil(suite.T(), errorResponse)
	assert.Equal(suite.T(), constants.ErrorInvalidGrant, errorResponse.Error)
	assert.Equal(suite.T(), "Authorization code is required", errorResponse.ErrorDescription)
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestValidateGrant_MissingClientID() {
	tokenRequest := &model.TokenRequest{
		GrantType: constants.GrantTypeAuthorizationCode,
		// Missing ClientID
		ClientSecret: "test_secret",
		Code:         "valid_auth_code",
		RedirectURI:  "http://localhost:8080/callback",
	}

	oauthApp := &appmodel.OAuthApplication{
		ClientID:     "test_client",
		ClientSecret: "test_secret",
	}

	errorResponse := suite.handler.ValidateGrant(tokenRequest, oauthApp)

	assert.NotNil(suite.T(), errorResponse)
	assert.Equal(suite.T(), constants.ErrorInvalidClient, errorResponse.Error)
	assert.Equal(suite.T(), "Client Id is required", errorResponse.ErrorDescription)
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestValidateGrant_MissingRedirectURI() {
	tokenRequest := &model.TokenRequest{
		GrantType:    constants.GrantTypeAuthorizationCode,
		ClientID:     "test_client",
		ClientSecret: "test_secret",
		Code:         "valid_auth_code",
		// Missing RedirectURI
	}

	oauthApp := &appmodel.OAuthApplication{
		ClientID:     "test_client",
		ClientSecret: "test_secret",
	}

	errorResponse := suite.handler.ValidateGrant(tokenRequest, oauthApp)

	assert.NotNil(suite.T(), errorResponse)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, errorResponse.Error)
	assert.Equal(suite.T(), "Redirect URI is required", errorResponse.ErrorDescription)
}

func (suite *AuthorizationCodeGrantHandlerTestSuite) TestValidateGrant_InvalidClientCredentials() {
	tokenRequest := &model.TokenRequest{
		GrantType:    constants.GrantTypeAuthorizationCode,
		ClientID:     "wrong_client",
		ClientSecret: "wrong_secret",
		Code:         "valid_auth_code",
		RedirectURI:  "http://localhost:8080/callback",
	}

	oauthApp := &appmodel.OAuthApplication{
		ClientID:     "test_client",
		ClientSecret: "test_secret",
	}

	errorResponse := suite.handler.ValidateGrant(tokenRequest, oauthApp)

	assert.NotNil(suite.T(), errorResponse)
	assert.Equal(suite.T(), constants.ErrorInvalidClient, errorResponse.Error)
	assert.Equal(suite.T(), "Invalid client credentials", errorResponse.ErrorDescription)
}
