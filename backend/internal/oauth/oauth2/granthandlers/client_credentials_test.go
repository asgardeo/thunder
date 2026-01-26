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
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/tokenservice"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/tests/mocks/jwtmock"
	"github.com/asgardeo/thunder/tests/mocks/oauth/oauth2/tokenservicemock"
)

// nolint:gosec // Test token, not a real credential
const testJWTToken = "test-jwt-token-123"
const testResourceURL = "https://mcp.example.com/mcp"

type ClientCredentialsGrantHandlerTestSuite struct {
	suite.Suite
	mockJWTService   *jwtmock.JWTServiceInterfaceMock
	mockTokenBuilder *tokenservicemock.TokenBuilderInterfaceMock
	handler          *clientCredentialsGrantHandler
	oauthApp         *appmodel.OAuthAppConfigProcessedDTO
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
	suite.mockTokenBuilder = tokenservicemock.NewTokenBuilderInterfaceMock(suite.T())
	suite.handler = &clientCredentialsGrantHandler{
		tokenBuilder: suite.mockTokenBuilder,
	}

	suite.oauthApp = &appmodel.OAuthAppConfigProcessedDTO{
		AppID:                   "app123",
		ClientID:                testClientID,
		HashedClientSecret:      "hashedsecret123",
		RedirectURIs:            []string{"https://example.com/callback"},
		GrantTypes:              []constants.GrantType{constants.GrantTypeClientCredentials},
		ResponseTypes:           []constants.ResponseType{constants.ResponseTypeCode},
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretBasic,
	}
}

func (suite *ClientCredentialsGrantHandlerTestSuite) TestNewClientCredentialsGrantHandler() {
	handler := newClientCredentialsGrantHandler(suite.mockTokenBuilder)
	assert.NotNil(suite.T(), handler)
	assert.Implements(suite.T(), (*GrantHandlerInterface)(nil), handler)
}

func (suite *ClientCredentialsGrantHandlerTestSuite) TestValidateGrant_Success() {
	tokenRequest := &model.TokenRequest{
		GrantType:    "client_credentials",
		ClientID:     testClientID,
		ClientSecret: "secret123",
		Scope:        "read",
	}

	result := suite.handler.ValidateGrant(context.TODO(), tokenRequest, suite.oauthApp)
	assert.Nil(suite.T(), result)
}

func (suite *ClientCredentialsGrantHandlerTestSuite) TestValidateGrant_WrongGrantType() {
	tokenRequest := &model.TokenRequest{
		GrantType:    "authorization_code",
		ClientID:     testClientID,
		ClientSecret: "secret123",
	}

	result := suite.handler.ValidateGrant(context.TODO(), tokenRequest, suite.oauthApp)
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

	result := suite.handler.ValidateGrant(context.TODO(), tokenRequest, suite.oauthApp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, result.Error)
	assert.Equal(suite.T(), "Client Id and secret are required", result.ErrorDescription)
}

func (suite *ClientCredentialsGrantHandlerTestSuite) TestValidateGrant_MissingClientSecret() {
	tokenRequest := &model.TokenRequest{
		GrantType:    "client_credentials",
		ClientID:     testClientID,
		ClientSecret: "",
	}

	result := suite.handler.ValidateGrant(context.TODO(), tokenRequest, suite.oauthApp)
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

	result := suite.handler.ValidateGrant(context.TODO(), tokenRequest, suite.oauthApp)
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
				ClientID:     testClientID,
				ClientSecret: "secret123",
				Scope:        tc.scope,
			}

			expectedToken := testJWTToken
			suite.mockTokenBuilder.On("BuildAccessToken", mock.Anything,
				mock.MatchedBy(func(ctx *tokenservice.AccessTokenBuildContext) bool {
					return ctx.Subject == testClientID &&
						ctx.Audience == testClientID &&
						ctx.ClientID == testClientID &&
						tokenservice.JoinScopes(ctx.Scopes) == tokenservice.JoinScopes(tc.expectedScopes)
				})).Return(&model.TokenDTO{
				Token:     expectedToken,
				TokenType: constants.TokenTypeBearer,
				IssuedAt:  int64(1234567890),
				ExpiresIn: 3600,
				Scopes:    tc.expectedScopes,
				ClientID:  testClientID,
				Subject:   testClientID,
				Audience:  testClientID,
			}, nil)

			result, errResp := suite.handler.HandleGrant(context.TODO(), tokenRequest, suite.oauthApp)

			assert.Nil(t, errResp)
			assert.NotNil(t, result)
			assert.Equal(t, expectedToken, result.AccessToken.Token)
			assert.Equal(t, constants.TokenTypeBearer, result.AccessToken.TokenType)
			assert.Equal(t, int64(3600), result.AccessToken.ExpiresIn)
			assert.Equal(t, tc.expectedScopes, result.AccessToken.Scopes)
			assert.Equal(t, testClientID, result.AccessToken.ClientID)

			// Verify token attributes
			assert.Equal(t, testClientID, result.AccessToken.Subject)
			assert.Equal(t, testClientID, result.AccessToken.Audience)

			suite.mockTokenBuilder.AssertExpectations(t)
		})
	}
}

func (suite *ClientCredentialsGrantHandlerTestSuite) TestHandleGrant_JWTGenerationError() {
	tokenRequest := &model.TokenRequest{
		GrantType:    "client_credentials",
		ClientID:     testClientID,
		ClientSecret: "secret123",
		Scope:        "read",
	}

	suite.mockTokenBuilder.On("BuildAccessToken", mock.Anything, mock.Anything).
		Return(nil, errors.New("JWT generation failed"))

	result, errResp := suite.handler.HandleGrant(context.TODO(), tokenRequest, suite.oauthApp)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), errResp)
	assert.Equal(suite.T(), constants.ErrorServerError, errResp.Error)
	assert.Equal(suite.T(), "Failed to generate token", errResp.ErrorDescription)

	suite.mockTokenBuilder.AssertExpectations(suite.T())
}

func (suite *ClientCredentialsGrantHandlerTestSuite) TestHandleGrant_NilTokenAttributes() {
	tokenRequest := &model.TokenRequest{
		GrantType:    "client_credentials",
		ClientID:     testClientID,
		ClientSecret: "secret123",
		Scope:        "read",
	}

	expectedToken := testJWTToken
	suite.mockTokenBuilder.On("BuildAccessToken", mock.Anything,
		mock.MatchedBy(func(ctx *tokenservice.AccessTokenBuildContext) bool {
			return ctx.Subject == testClientID && ctx.Audience == testClientID &&
				tokenservice.JoinScopes(ctx.Scopes) == testScopeRead
		})).Return(&model.TokenDTO{
		Token:     expectedToken,
		TokenType: constants.TokenTypeBearer,
		IssuedAt:  int64(1234567890),
		ExpiresIn: 3600,
		Scopes:    []string{"read"},
		ClientID:  "client123",
		Subject:   testClientID,
		Audience:  testClientID,
	}, nil)

	result, errResp := suite.handler.HandleGrant(context.TODO(), tokenRequest, suite.oauthApp)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), expectedToken, result.AccessToken.Token)

	// Verify token attributes
	assert.Equal(suite.T(), testClientID, result.AccessToken.Subject)
	assert.Equal(suite.T(), testClientID, result.AccessToken.Audience)

	suite.mockTokenBuilder.AssertExpectations(suite.T())
}

func (suite *ClientCredentialsGrantHandlerTestSuite) TestHandleGrant_TokenTimingValidation() {
	tokenRequest := &model.TokenRequest{
		GrantType:    "client_credentials",
		ClientID:     testClientID,
		ClientSecret: "secret123",
		Scope:        "read",
	}

	expectedToken := testJWTToken
	now := time.Now().Unix()
	suite.mockTokenBuilder.On("BuildAccessToken", mock.Anything, mock.Anything).
		Return(&model.TokenDTO{
			Token:     expectedToken,
			TokenType: constants.TokenTypeBearer,
			IssuedAt:  now,
			ExpiresIn: 3600,
			Scopes:    []string{"read"},
			ClientID:  testClientID,
		}, nil)

	startTime := time.Now().Unix()
	result, errResp := suite.handler.HandleGrant(context.TODO(), tokenRequest, suite.oauthApp)
	endTime := time.Now().Unix()

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)

	// Verify the issued time is within reasonable bounds
	assert.GreaterOrEqual(suite.T(), result.AccessToken.IssuedAt, startTime)
	assert.LessOrEqual(suite.T(), result.AccessToken.IssuedAt, endTime)

	suite.mockTokenBuilder.AssertExpectations(suite.T())
}

// Resource Parameter Tests (RFC 8707) for Client Credentials Grant

func (suite *ClientCredentialsGrantHandlerTestSuite) TestHandleGrant_WithResourceParameter() {
	tokenRequest := &model.TokenRequest{
		GrantType:    "client_credentials",
		ClientID:     testClientID,
		ClientSecret: "secret123",
		Scope:        "read",
		Resource:     "https://mcp.example.com/mcp",
	}

	var capturedAudience string
	suite.mockTokenBuilder.On("BuildAccessToken", mock.Anything,
		mock.MatchedBy(func(ctx *tokenservice.AccessTokenBuildContext) bool {
			capturedAudience = ctx.Audience
			return ctx.Subject == testClientID && ctx.Audience == "https://mcp.example.com/mcp"
		})).Return(&model.TokenDTO{
		Token:     testJWTToken,
		TokenType: constants.TokenTypeBearer,
		IssuedAt:  int64(1234567890),
		ExpiresIn: 3600,
		Scopes:    []string{"read"},
		ClientID:  "client123",
		Subject:   testClientID,
		Audience:  "https://mcp.example.com/mcp",
	}, nil)

	result, errResp := suite.handler.HandleGrant(context.TODO(), tokenRequest, suite.oauthApp)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)

	// Verify resource was included in audience
	assert.Equal(suite.T(), "https://mcp.example.com/mcp", capturedAudience)

	// Verify token attributes use resource as audience
	assert.Equal(suite.T(), "https://mcp.example.com/mcp", result.AccessToken.Audience)
}

func (suite *ClientCredentialsGrantHandlerTestSuite) TestHandleGrant_WithoutResourceParameter() {
	tokenRequest := &model.TokenRequest{
		GrantType:    "client_credentials",
		ClientID:     testClientID,
		ClientSecret: "secret123",
		Scope:        "read",
	}

	var capturedAudience string
	suite.mockTokenBuilder.On("BuildAccessToken", mock.Anything,
		mock.MatchedBy(func(ctx *tokenservice.AccessTokenBuildContext) bool {
			capturedAudience = ctx.Audience
			return ctx.Subject == testClientID && ctx.Audience == testClientID
		})).Return(&model.TokenDTO{
		Token:     testJWTToken,
		TokenType: constants.TokenTypeBearer,
		IssuedAt:  int64(1234567890),
		ExpiresIn: 3600,
		Scopes:    []string{"read"},
		ClientID:  "client123",
		Subject:   testClientID,
		Audience:  testClientID,
	}, nil)

	result, errResp := suite.handler.HandleGrant(context.TODO(), tokenRequest, suite.oauthApp)

	assert.Nil(suite.T(), errResp)
	assert.NotNil(suite.T(), result)

	// Verify default audience (client_id) when no resource parameter
	assert.Equal(suite.T(), testClientID, capturedAudience)

	// Verify token attributes use client ID as audience when no resource
	assert.Equal(suite.T(), testClientID, result.AccessToken.Audience)
}
