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

package granthandlers

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/model"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/tests/mocks/jwtmock"
)

// nolint:gosec // Test token, not a real credential
const testJWTToken = "test-jwt-token-123"
const testResourceURL = "https://mcp.example.com/mcp"

type ClientCredentialsGrantHandlerTestSuite struct {
	suite.Suite
	mockJWTService *jwtmock.JWTServiceInterfaceMock
	handler        *clientCredentialsGrantHandler
	oauthApp       *appmodel.OAuthAppConfigProcessedDTO
}

func TestClientCredentialsGrantHandlerSuite(t *testing.T) {
	suite.Run(t, new(ClientCredentialsGrantHandlerTestSuite))
}

func (suite *ClientCredentialsGrantHandlerTestSuite) SetupTest() {
	// Initialize Thunder Runtime for tests
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "https://test.thunder.io",
			ValidityPeriod: 3600,
		},
	}
	err := config.InitializeThunderRuntime("", testConfig)
	assert.NoError(suite.T(), err)

	suite.mockJWTService = jwtmock.NewJWTServiceInterfaceMock(suite.T())
	suite.handler = &clientCredentialsGrantHandler{
		jwtService: suite.mockJWTService,
	}

	suite.oauthApp = &appmodel.OAuthAppConfigProcessedDTO{
		AppID:                   "app123",
		ClientID:                "client123",
		HashedClientSecret:      "hashedsecret123",
		RedirectURIs:            []string{"https://example.com/callback"},
		GrantTypes:              []constants.GrantType{constants.GrantTypeClientCredentials},
		ResponseTypes:           []constants.ResponseType{constants.ResponseTypeCode},
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretBasic,
	}
}

func (suite *ClientCredentialsGrantHandlerTestSuite) TestNewClientCredentialsGrantHandler() {
	handler := newClientCredentialsGrantHandler(suite.mockJWTService)
	assert.NotNil(suite.T(), handler)
	assert.Implements(suite.T(), (*GrantHandlerInterface)(nil), handler)
}

func (suite *ClientCredentialsGrantHandlerTestSuite) TestValidateGrant_Success() {
	tokenRequest := &model.TokenRequest{
		GrantType:    "client_credentials",
		ClientID:     "client123",
		ClientSecret: "secret123",
		Scope:        "read",
	}

	result := suite.handler.ValidateGrant(tokenRequest, suite.oauthApp)
	assert.Nil(suite.T(), result)
}

func (suite *ClientCredentialsGrantHandlerTestSuite) TestValidateGrant_WrongGrantType() {
	tokenRequest := &model.TokenRequest{
		GrantType:    "authorization_code",
		ClientID:     "client123",
		ClientSecret: "secret123",
	}

	result := suite.handler.ValidateGrant(tokenRequest, suite.oauthApp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), constants.ErrorUnsupportedGrantType, result.Error)
	assert.Equal(suite.T(), "Unsupported grant type", result.ErrorDescription)
}

func (suite *ClientCredentialsGrantHandlerTestSuite) TestValidateGrant_MissingClientID() {
	tokenRequest := &model.TokenRequest{
		GrantType:    "client_credentials",
		ClientID:     "",
		ClientSecret: "secret123",
	}

	result := suite.handler.ValidateGrant(tokenRequest, suite.oauthApp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, result.Error)
	assert.Equal(suite.T(), "Client Id and secret are required", result.ErrorDescription)
}

func (suite *ClientCredentialsGrantHandlerTestSuite) TestValidateGrant_MissingClientSecret() {
	tokenRequest := &model.TokenRequest{
		GrantType:    "client_credentials",
		ClientID:     "client123",
		ClientSecret: "",
	}

	result := suite.handler.ValidateGrant(tokenRequest, suite.oauthApp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, result.Error)
	assert.Equal(suite.T(), "Client Id and secret are required", result.ErrorDescription)
}

func (suite *ClientCredentialsGrantHandlerTestSuite) TestValidateGrant_MissingBothCredentials() {
	tokenRequest := &model.TokenRequest{
		GrantType:    "client_credentials",
		ClientID:     "",
		ClientSecret: "",
	}

	result := suite.handler.ValidateGrant(tokenRequest, suite.oauthApp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, result.Error)
	assert.Equal(suite.T(), "Client Id and secret are required", result.ErrorDescription)
}

func (suite *ClientCredentialsGrantHandlerTestSuite) TestHandleGrant_Success() {
	testCases := []struct {
		name              string
		scope             string
		expectedJWTClaims map[string]interface{}
		expectedScopes    []string
	}{
		{
			name:              "WithValidScope",
			scope:             "read write",
			expectedJWTClaims: map[string]interface{}{"scope": "read write"},
			expectedScopes:    []string{"read", "write"},
		},
		{
			name:              "WithoutScope",
			scope:             "",
			expectedJWTClaims: map[string]interface{}{},
			expectedScopes:    []string{},
		},
		{
			name:              "WithWhitespaceScope",
			scope:             "   ",
			expectedJWTClaims: map[string]interface{}{},
			expectedScopes:    []string{},
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Reset mock for each test case
			suite.mockJWTService.Mock = mock.Mock{}

			tokenRequest := &model.TokenRequest{
				GrantType:    "client_credentials",
				ClientID:     "client123",
				ClientSecret: "secret123",
				Scope:        tc.scope,
			}

			ctx := &model.TokenContext{
				TokenAttributes: make(map[string]interface{}),
			}

			expectedToken := testJWTToken
			suite.mockJWTService.On("GenerateJWT", "client123", "client123", mock.AnythingOfType("string"),
				int64(3600), tc.expectedJWTClaims).Return(expectedToken, int64(1234567890), nil)

			result, errResp := suite.handler.HandleGrant(tokenRequest, suite.oauthApp, ctx)

			assert.Nil(t, errResp)
			assert.NotNil(t, result)
			assert.Equal(t, expectedToken, result.AccessToken.Token)
			assert.Equal(t, constants.TokenTypeBearer, result.AccessToken.TokenType)
			assert.Equal(t, int64(3600), result.AccessToken.ExpiresIn)
			assert.Equal(t, tc.expectedScopes, result.AccessToken.Scopes)
			assert.Equal(t, "client123", result.AccessToken.ClientID)

			// Verify context attributes
			assert.Equal(t, "client123", ctx.TokenAttributes["sub"])
			assert.Equal(t, "client123", ctx.TokenAttributes["aud"])

			suite.mockJWTService.AssertExpectations(t)
		})
	}
}

func (suite *ClientCredentialsGrantHandlerTestSuite) TestHandleGrant_JWTGenerationError() {
	tokenRequest := &model.TokenRequest{
		GrantType:    "client_credentials",
		ClientID:     "client123",
		ClientSecret: "secret123",
		Scope:        "read",
	}

	ctx := &model.TokenContext{
		TokenAttributes: make(map[string]interface{}),
	}

	suite.mockJWTService.On("GenerateJWT", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return("", int64(0), errors.New("JWT generation failed"))

	result, errResp := suite.handler.HandleGrant(tokenRequest, suite.oauthApp, ctx)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), errResp)
	assert.Equal(suite.T(), constants.ErrorServerError, errResp.Error)
	assert.Equal(suite.T(), "Failed to generate token", errResp.ErrorDescription)

	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *ClientCredentialsGrantHandlerTestSuite) TestHandleGrant_NilTokenAttributes() {
	tokenRequest := &model.TokenRequest{
		GrantType:    "client_credentials",
		ClientID:     "client123",
		ClientSecret: "secret123",
		Scope:        "read",
	}

	ctx := &model.TokenContext{
		TokenAttributes: nil,
	}

	expectedToken := testJWTToken
	suite.mockJWTService.On("GenerateJWT", "client123", "client123", mock.AnythingOfType("string"), int64(3600),
		map[string]interface{}{"scope": "read"}).Return(expectedToken, int64(1234567890), nil)

	result, errResp := suite.handler.HandleGrant(tokenRequest, suite.oauthApp, ctx)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), expectedToken, result.AccessToken.Token)

	// Verify context attributes were initialized and set
	assert.NotNil(suite.T(), ctx.TokenAttributes)
	assert.Equal(suite.T(), "client123", ctx.TokenAttributes["sub"])
	assert.Equal(suite.T(), "client123", ctx.TokenAttributes["aud"])

	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *ClientCredentialsGrantHandlerTestSuite) TestHandleGrant_TokenTimingValidation() {
	tokenRequest := &model.TokenRequest{
		GrantType:    "client_credentials",
		ClientID:     "client123",
		ClientSecret: "secret123",
		Scope:        "read",
	}

	ctx := &model.TokenContext{
		TokenAttributes: make(map[string]interface{}),
	}

	expectedToken := testJWTToken
	suite.mockJWTService.On("GenerateJWT", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedToken, int64(1234567890), nil)

	startTime := time.Now().Unix()
	result, errResp := suite.handler.HandleGrant(tokenRequest, suite.oauthApp, ctx)
	endTime := time.Now().Unix()

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)

	// Verify the issued time is within reasonable bounds
	assert.GreaterOrEqual(suite.T(), result.AccessToken.IssuedAt, startTime)
	assert.LessOrEqual(suite.T(), result.AccessToken.IssuedAt, endTime)

	suite.mockJWTService.AssertExpectations(suite.T())
}

// Resource Parameter Tests (RFC 8707) for Client Credentials Grant

func (suite *ClientCredentialsGrantHandlerTestSuite) TestHandleGrant_WithResourceParameter() {
	tokenRequest := &model.TokenRequest{
		GrantType:    "client_credentials",
		ClientID:     "client123",
		ClientSecret: "secret123",
		Scope:        "read",
		Resource:     "https://mcp.example.com/mcp",
	}

	ctx := &model.TokenContext{
		TokenAttributes: make(map[string]interface{}),
	}

	capturedClaims := map[string]interface{}{}
	suite.mockJWTService.On("GenerateJWT", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.MatchedBy(func(claims map[string]interface{}) bool {
			for k, v := range claims {
				capturedClaims[k] = v
			}
			return true
		})).Return(testJWTToken, int64(1234567890), nil)

	result, errResp := suite.handler.HandleGrant(tokenRequest, suite.oauthApp, ctx)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)

	// Verify resource was included in audience claim
	assert.NotNil(suite.T(), capturedClaims["aud"])
	assert.Equal(suite.T(), "https://mcp.example.com/mcp", capturedClaims["aud"])

	// Verify context attributes use resource as audience
	assert.Equal(suite.T(), "https://mcp.example.com/mcp", ctx.TokenAttributes["aud"])
}

func (suite *ClientCredentialsGrantHandlerTestSuite) TestHandleGrant_WithoutResourceParameter() {
	tokenRequest := &model.TokenRequest{
		GrantType:    "client_credentials",
		ClientID:     "client123",
		ClientSecret: "secret123",
		Scope:        "read",
	}

	ctx := &model.TokenContext{
		TokenAttributes: make(map[string]interface{}),
	}

	capturedClaims := map[string]interface{}{}
	suite.mockJWTService.On("GenerateJWT", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.MatchedBy(func(claims map[string]interface{}) bool {
			for k, v := range claims {
				capturedClaims[k] = v
			}
			return true
		})).Return(testJWTToken, int64(1234567890), nil)

	result, errResp := suite.handler.HandleGrant(tokenRequest, suite.oauthApp, ctx)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)

	// Verify no audience claim in JWT
	assert.Nil(suite.T(), capturedClaims["aud"])

	// Verify context attributes use client ID as audience when no resource
	assert.Equal(suite.T(), "client123", ctx.TokenAttributes["aud"])
}
