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

package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type OAuth2ModelTestSuite struct {
	suite.Suite
}

func TestOAuth2ModelSuite(t *testing.T) {
	suite.Run(t, new(OAuth2ModelTestSuite))
}

func (suite *OAuth2ModelTestSuite) TestOAuthParameters_Creation() {
	params := OAuthParameters{
		SessionDataKey: "test-session-key",
		State:          "test-state",
		ClientID:       "test-client-id",
		RedirectURI:    "https://example.com/callback",
		ResponseType:   "code",
		Scopes:         "read write",
	}

	assert.Equal(suite.T(), "test-session-key", params.SessionDataKey)
	assert.Equal(suite.T(), "test-state", params.State)
	assert.Equal(suite.T(), "test-client-id", params.ClientID)
	assert.Equal(suite.T(), "https://example.com/callback", params.RedirectURI)
	assert.Equal(suite.T(), "code", params.ResponseType)
	assert.Equal(suite.T(), "read write", params.Scopes)
}

func (suite *OAuth2ModelTestSuite) TestOAuthParameters_EmptyValues() {
	params := OAuthParameters{}

	assert.Empty(suite.T(), params.SessionDataKey)
	assert.Empty(suite.T(), params.State)
	assert.Empty(suite.T(), params.ClientID)
	assert.Empty(suite.T(), params.RedirectURI)
	assert.Empty(suite.T(), params.ResponseType)
	assert.Empty(suite.T(), params.Scopes)
}

func (suite *OAuth2ModelTestSuite) TestErrorResponse_JSONSerialization() {
	errorResp := ErrorResponse{
		Error:            "invalid_request",
		ErrorDescription: "The request is missing a required parameter",
		ErrorURI:         "https://example.com/error",
	}

	jsonData, err := json.Marshal(errorResp)
	assert.NoError(suite.T(), err)

	var unmarshaled ErrorResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), errorResp.Error, unmarshaled.Error)
	assert.Equal(suite.T(), errorResp.ErrorDescription, unmarshaled.ErrorDescription)
	assert.Equal(suite.T(), errorResp.ErrorURI, unmarshaled.ErrorURI)
}

func (suite *OAuth2ModelTestSuite) TestErrorResponse_OptionalFields() {
	// Test with only required field
	errorResp := ErrorResponse{
		Error: "invalid_client",
	}

	jsonData, err := json.Marshal(errorResp)
	assert.NoError(suite.T(), err)

	expected := `{"error":"invalid_client"}`
	assert.JSONEq(suite.T(), expected, string(jsonData))
}

func (suite *OAuth2ModelTestSuite) TestTokenRequest_JSONDeserialization() {
	jsonStr := `{
		"grant_type": "authorization_code",
		"client_id": "test-client",
		"client_secret": "test-secret",
		"code": "test-auth-code",
		"redirect_uri": "https://example.com/callback",
		"scope": "read write"
	}`

	var tokenReq TokenRequest
	err := json.Unmarshal([]byte(jsonStr), &tokenReq)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "authorization_code", tokenReq.GrantType)
	assert.Equal(suite.T(), "test-client", tokenReq.ClientID)
	assert.Equal(suite.T(), "test-secret", tokenReq.ClientSecret)
	assert.Equal(suite.T(), "test-auth-code", tokenReq.Code)
	assert.Equal(suite.T(), "https://example.com/callback", tokenReq.RedirectURI)
	assert.Equal(suite.T(), "read write", tokenReq.Scope)
}

func (suite *OAuth2ModelTestSuite) TestTokenRequest_RefreshTokenGrant() {
	tokenReq := TokenRequest{
		GrantType:    "refresh_token",
		ClientID:     "test-client",
		RefreshToken: "test-refresh-token",
		Scope:        "read",
	}

	jsonData, err := json.Marshal(tokenReq)
	assert.NoError(suite.T(), err)

	var unmarshaled TokenRequest
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), tokenReq.GrantType, unmarshaled.GrantType)
	assert.Equal(suite.T(), tokenReq.ClientID, unmarshaled.ClientID)
	assert.Equal(suite.T(), tokenReq.RefreshToken, unmarshaled.RefreshToken)
	assert.Equal(suite.T(), tokenReq.Scope, unmarshaled.Scope)
}

func (suite *OAuth2ModelTestSuite) TestTokenRequest_ClientCredentialsGrant() {
	tokenReq := TokenRequest{
		GrantType:    "client_credentials",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		Scope:        "api:read api:write",
	}

	assert.Equal(suite.T(), "client_credentials", tokenReq.GrantType)
	assert.Equal(suite.T(), "test-client", tokenReq.ClientID)
	assert.Equal(suite.T(), "test-secret", tokenReq.ClientSecret)
	assert.Equal(suite.T(), "api:read api:write", tokenReq.Scope)
}

func (suite *OAuth2ModelTestSuite) TestTokenResponse_JSONSerialization() {
	tokenResp := TokenResponse{
		AccessToken:  "test-access-token",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		RefreshToken: "test-refresh-token",
		Scope:        "read write",
	}

	jsonData, err := json.Marshal(tokenResp)
	assert.NoError(suite.T(), err)

	var unmarshaled TokenResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), tokenResp.AccessToken, unmarshaled.AccessToken)
	assert.Equal(suite.T(), tokenResp.TokenType, unmarshaled.TokenType)
	assert.Equal(suite.T(), tokenResp.ExpiresIn, unmarshaled.ExpiresIn)
	assert.Equal(suite.T(), tokenResp.RefreshToken, unmarshaled.RefreshToken)
	assert.Equal(suite.T(), tokenResp.Scope, unmarshaled.Scope)
}

func (suite *OAuth2ModelTestSuite) TestTokenResponse_OptionalFields() {
	tokenResp := TokenResponse{
		AccessToken: "test-access-token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
	}

	jsonData, err := json.Marshal(tokenResp)
	assert.NoError(suite.T(), err)

	// Should not include omitempty fields when they're empty
	assert.NotContains(suite.T(), string(jsonData), "refresh_token")
	assert.NotContains(suite.T(), string(jsonData), "scope")
}

func (suite *OAuth2ModelTestSuite) TestTokenContext_Creation() {
	context := TokenContext{
		TokenAttributes: map[string]interface{}{
			"aud":    "test-audience",
			"sub":    "test-subject",
			"custom": "custom-value",
		},
	}

	assert.NotNil(suite.T(), context.TokenAttributes)
	assert.Equal(suite.T(), "test-audience", context.TokenAttributes["aud"])
	assert.Equal(suite.T(), "test-subject", context.TokenAttributes["sub"])
	assert.Equal(suite.T(), "custom-value", context.TokenAttributes["custom"])
}

func (suite *OAuth2ModelTestSuite) TestTokenContext_EmptyAttributes() {
	context := TokenContext{}
	assert.Nil(suite.T(), context.TokenAttributes)
}

func (suite *OAuth2ModelTestSuite) TestTokenDTO_Creation() {
	tokenDTO := TokenDTO{
		Token:     "test-token-value",
		TokenType: "Bearer",
		IssuedAt:  1234567890,
		ExpiresIn: 3600,
		Scopes:    []string{"read", "write", "admin"},
		ClientID:  "test-client",
	}

	assert.Equal(suite.T(), "test-token-value", tokenDTO.Token)
	assert.Equal(suite.T(), "Bearer", tokenDTO.TokenType)
	assert.Equal(suite.T(), int64(1234567890), tokenDTO.IssuedAt)
	assert.Equal(suite.T(), int64(3600), tokenDTO.ExpiresIn)
	assert.Equal(suite.T(), []string{"read", "write", "admin"}, tokenDTO.Scopes)
	assert.Equal(suite.T(), "test-client", tokenDTO.ClientID)
}

func (suite *OAuth2ModelTestSuite) TestTokenDTO_JSONSerialization() {
	tokenDTO := TokenDTO{
		Token:     "test-token",
		TokenType: "Bearer",
		IssuedAt:  1234567890,
		ExpiresIn: 3600,
		Scopes:    []string{"read", "write"},
		ClientID:  "test-client",
	}

	jsonData, err := json.Marshal(tokenDTO)
	assert.NoError(suite.T(), err)

	var unmarshaled TokenDTO
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), tokenDTO.Token, unmarshaled.Token)
	assert.Equal(suite.T(), tokenDTO.TokenType, unmarshaled.TokenType)
	assert.Equal(suite.T(), tokenDTO.IssuedAt, unmarshaled.IssuedAt)
	assert.Equal(suite.T(), tokenDTO.ExpiresIn, unmarshaled.ExpiresIn)
	assert.Equal(suite.T(), tokenDTO.Scopes, unmarshaled.Scopes)
	assert.Equal(suite.T(), tokenDTO.ClientID, unmarshaled.ClientID)
}

func (suite *OAuth2ModelTestSuite) TestTokenResponseDTO_Creation() {
	accessToken := TokenDTO{
		Token:     "access-token-value",
		TokenType: "Bearer",
		IssuedAt:  1234567890,
		ExpiresIn: 3600,
		Scopes:    []string{"read", "write"},
		ClientID:  "test-client",
	}

	refreshToken := TokenDTO{
		Token:     "refresh-token-value",
		TokenType: "refresh",
		IssuedAt:  1234567890,
		ExpiresIn: 7200,
		ClientID:  "test-client",
	}

	tokenResponseDTO := TokenResponseDTO{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	assert.Equal(suite.T(), accessToken, tokenResponseDTO.AccessToken)
	assert.Equal(suite.T(), refreshToken, tokenResponseDTO.RefreshToken)
}

func (suite *OAuth2ModelTestSuite) TestTokenResponseDTO_JSONSerialization() {
	accessToken := TokenDTO{
		Token:     "access-token-value",
		TokenType: "Bearer",
		IssuedAt:  1234567890,
		ExpiresIn: 3600,
		ClientID:  "test-client",
	}

	tokenResponseDTO := TokenResponseDTO{
		AccessToken: accessToken,
	}

	jsonData, err := json.Marshal(tokenResponseDTO)
	assert.NoError(suite.T(), err)

	var unmarshaled TokenResponseDTO
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), tokenResponseDTO.AccessToken, unmarshaled.AccessToken)
	// RefreshToken should be omitted if empty
	assert.Empty(suite.T(), unmarshaled.RefreshToken.Token)
}

func (suite *OAuth2ModelTestSuite) TestTokenResponseDTO_WithRefreshToken() {
	accessToken := TokenDTO{
		Token:     "access-token-value",
		TokenType: "Bearer",
		IssuedAt:  1234567890,
		ExpiresIn: 3600,
		ClientID:  "test-client",
	}

	refreshToken := TokenDTO{
		Token:     "refresh-token-value",
		TokenType: "refresh",
		IssuedAt:  1234567890,
		ExpiresIn: 7200,
		ClientID:  "test-client",
	}

	tokenResponseDTO := TokenResponseDTO{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	jsonData, err := json.Marshal(tokenResponseDTO)
	assert.NoError(suite.T(), err)

	var unmarshaled TokenResponseDTO
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), tokenResponseDTO.AccessToken, unmarshaled.AccessToken)
	assert.Equal(suite.T(), tokenResponseDTO.RefreshToken, unmarshaled.RefreshToken)
}