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

package oidc

import (
	"testing"

	"github.com/stretchr/testify/suite"

	authncm "github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/authn/oauth"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/userprovider"
	"github.com/asgardeo/thunder/tests/mocks/authn/oauthmock"
	"github.com/asgardeo/thunder/tests/mocks/httpmock"
	"github.com/asgardeo/thunder/tests/mocks/idp/idpmock"
	"github.com/asgardeo/thunder/tests/mocks/jose/jwtmock"
	"github.com/asgardeo/thunder/tests/mocks/userprovidermock"
)

const (
	testOIDCIDPID = "idp123"
)

type OIDCAuthnServiceTestSuite struct {
	suite.Suite
	mockOAuthService *oauthmock.OAuthAuthnServiceInterfaceMock
	mockHTTPClient   *httpmock.HTTPClientInterfaceMock
	mockIdpService   *idpmock.IDPServiceInterfaceMock
	mockUserProvider *userprovidermock.UserProviderInterfaceMock
	mockJWTService   *jwtmock.JWTServiceInterfaceMock
	endpoints        oauth.OAuthEndpoints
	service          oidcAuthnService
}

func TestOIDCAuthnServiceTestSuite(t *testing.T) {
	suite.Run(t, new(OIDCAuthnServiceTestSuite))
}

func (suite *OIDCAuthnServiceTestSuite) SetupTest() {
	suite.mockOAuthService = oauthmock.NewOAuthAuthnServiceInterfaceMock(suite.T())
	suite.mockHTTPClient = httpmock.NewHTTPClientInterfaceMock(suite.T())
	suite.mockIdpService = idpmock.NewIDPServiceInterfaceMock(suite.T())
	suite.mockUserProvider = userprovidermock.NewUserProviderInterfaceMock(suite.T())
	suite.mockJWTService = jwtmock.NewJWTServiceInterfaceMock(suite.T())
	suite.endpoints = oauth.OAuthEndpoints{
		AuthorizationEndpoint: "https://localhost:8090/oauth/authorize",
		TokenEndpoint:         "https://localhost:8090/oauth/token",
		UserInfoEndpoint:      "https://localhost:8090/oauth/userinfo",
	}

	service := newOIDCAuthnService(suite.mockHTTPClient, suite.mockIdpService,
		suite.mockUserProvider, suite.mockJWTService)

	cast, ok := service.(*oidcAuthnService)
	suite.True(ok, "service is not of type *oidcAuthnService")
	suite.service = *cast
	suite.service.internal = suite.mockOAuthService
}

func (suite *OIDCAuthnServiceTestSuite) TestGetOAuthClientConfigWithOpenIDScope() {
	idpID := testOIDCIDPID
	config := &oauth.OAuthClientConfig{
		ClientID:     "client123",
		ClientSecret: "secret",
		Scopes:       []string{"openid", "profile", "email"},
	}
	suite.mockOAuthService.On("GetOAuthClientConfig", idpID).Return(config, nil)

	result, err := suite.service.GetOAuthClientConfig(idpID)
	suite.Nil(err)
	suite.NotNil(result)

	suite.Contains(result.Scopes, "openid")
	suite.Contains(result.Scopes, "profile")
	suite.Contains(result.Scopes, "email")

	// Ensure openid is not duplicated
	suite.Equal(3, len(result.Scopes))
}

func (suite *OIDCAuthnServiceTestSuite) TestGetOAuthClientConfigWithoutOpenIDScope() {
	idpID := testOIDCIDPID
	config := &oauth.OAuthClientConfig{
		ClientID:     "client123",
		ClientSecret: "secret",
		Scopes:       []string{"profile"},
	}
	suite.mockOAuthService.On("GetOAuthClientConfig", idpID).Return(config, nil)

	result, err := suite.service.GetOAuthClientConfig(idpID)
	suite.Nil(err)
	suite.NotNil(result)

	// Scopes come from IDP config as-is, no automatic addition of openid scope
	suite.NotContains(result.Scopes, "openid")
	suite.Contains(result.Scopes, "profile")
}

func (suite *OIDCAuthnServiceTestSuite) TestBuildAuthorizeURLSuccess() {
	expectedURL := "https://example.com/authorize?client_id=test"
	suite.mockOAuthService.On("BuildAuthorizeURL", testOIDCIDPID).Return(expectedURL, nil)

	url, err := suite.service.BuildAuthorizeURL(testOIDCIDPID)
	suite.Nil(err)
	suite.Equal(expectedURL, url)
}

func (suite *OIDCAuthnServiceTestSuite) TestBuildAuthorizeURLError() {
	svcErr := &serviceerror.ServiceError{
		Code:             "ERROR",
		ErrorDescription: "Failed to build URL",
	}
	suite.mockOAuthService.On("BuildAuthorizeURL", testOIDCIDPID).Return("", svcErr)

	url, err := suite.service.BuildAuthorizeURL(testOIDCIDPID)
	suite.Empty(url)
	suite.NotNil(err)
	suite.Equal(svcErr.Code, err.Code)
}

func (suite *OIDCAuthnServiceTestSuite) TestExchangeCodeForTokenSuccess() {
	tests := []struct {
		name             string
		validateResponse bool
		setupMocks       func()
	}{
		{
			name:             "WithValidation",
			validateResponse: true,
			setupMocks: func() {
				code := "auth_code"
				tokenResp := &oauth.TokenResponse{
					AccessToken: "access_token",
					IDToken:     "id_token",
					TokenType:   "Bearer",
				}
				suite.mockOAuthService.On("ExchangeCodeForToken", testOIDCIDPID, code, false).Return(tokenResp, nil)
				suite.mockOAuthService.On("GetOAuthClientConfig", testOIDCIDPID).Return(&oauth.OAuthClientConfig{
					OAuthEndpoints: oauth.OAuthEndpoints{JwksEndpoint: "https://example.com/jwks"},
				}, nil)
				suite.mockJWTService.On("VerifyJWTWithJWKS", "id_token",
					"https://example.com/jwks", "", "").Return(nil)
			},
		},
		{
			name:             "WithoutValidation",
			validateResponse: false,
			setupMocks: func() {
				code := "auth_code"
				tokenResp := &oauth.TokenResponse{
					AccessToken: "access_token",
					IDToken:     "id_token",
					TokenType:   "Bearer",
				}
				suite.mockOAuthService.On("ExchangeCodeForToken", testOIDCIDPID, code, false).Return(tokenResp, nil)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.mockOAuthService = oauthmock.NewOAuthAuthnServiceInterfaceMock(suite.T())
			suite.mockJWTService = jwtmock.NewJWTServiceInterfaceMock(suite.T())

			service := newOIDCAuthnService(suite.mockHTTPClient, suite.mockIdpService,
				suite.mockUserProvider, suite.mockJWTService)
			cast, ok := service.(*oidcAuthnService)
			suite.True(ok, "service is not of type *oidcAuthnService")
			suite.service = *cast
			suite.service.internal = suite.mockOAuthService

			tc.setupMocks()

			result, err := suite.service.ExchangeCodeForToken(testOIDCIDPID, "auth_code", tc.validateResponse)
			suite.Nil(err)
			suite.NotNil(result)
			suite.Equal("access_token", result.AccessToken)
		})
	}
}

func (suite *OIDCAuthnServiceTestSuite) TestValidateTokenResponseSuccess() {
	tests := []struct {
		name            string
		validateIDToken bool
		setupMocks      func()
	}{
		{
			name:            "WithIDTokenValidation",
			validateIDToken: true,
			setupMocks: func() {
				suite.mockOAuthService.On("GetOAuthClientConfig", testOIDCIDPID).Return(&oauth.OAuthClientConfig{
					OAuthEndpoints: oauth.OAuthEndpoints{JwksEndpoint: "https://example.com/jwks"},
				}, nil)
				suite.mockJWTService.On("VerifyJWTWithJWKS", "id_token",
					"https://example.com/jwks", "", "").Return(nil)
			},
		},
		{
			name:            "WithoutIDTokenValidation",
			validateIDToken: false,
			setupMocks:      func() {},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.mockOAuthService = oauthmock.NewOAuthAuthnServiceInterfaceMock(suite.T())
			suite.mockJWTService = jwtmock.NewJWTServiceInterfaceMock(suite.T())

			service := newOIDCAuthnService(suite.mockHTTPClient, suite.mockIdpService,
				suite.mockUserProvider, suite.mockJWTService)
			cast, ok := service.(*oidcAuthnService)
			suite.True(ok, "service is not of type *oidcAuthnService")
			suite.service = *cast
			suite.service.internal = suite.mockOAuthService

			tc.setupMocks()

			tokenResp := &oauth.TokenResponse{
				AccessToken: "access_token",
				IDToken:     "id_token",
				TokenType:   "Bearer",
			}
			err := suite.service.ValidateTokenResponse(testOIDCIDPID, tokenResp, tc.validateIDToken)
			suite.Nil(err)
		})
	}
}

func (suite *OIDCAuthnServiceTestSuite) TestValidateTokenResponseWithError() {
	tests := []struct {
		name string
		resp *oauth.TokenResponse
	}{
		{
			name: "NilResponse",
			resp: nil,
		},
		{
			name: "EmptyAccessToken",
			resp: &oauth.TokenResponse{AccessToken: "", IDToken: "id_token"},
		},
		{
			name: "EmptyIDToken",
			resp: &oauth.TokenResponse{AccessToken: "access_token", IDToken: ""},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			err := suite.service.ValidateTokenResponse(testOIDCIDPID, tc.resp, false)
			suite.NotNil(err)
			suite.Equal(oauth.ErrorInvalidTokenResponse.Code, err.Code)
		})
	}
}

func (suite *OIDCAuthnServiceTestSuite) TestValidateIDTokenSuccess() {
	tests := []struct {
		name       string
		setupMocks func()
	}{
		{
			name: "WithJWKSEndpoint",
			setupMocks: func() {
				suite.mockOAuthService.On("GetOAuthClientConfig", testOIDCIDPID).Return(&oauth.OAuthClientConfig{
					OAuthEndpoints: oauth.OAuthEndpoints{JwksEndpoint: "https://example.com/jwks"},
				}, nil)
				suite.mockJWTService.On("VerifyJWTWithJWKS", "valid_id_token",
					"https://example.com/jwks", "", "").Return(nil)
			},
		},
		{
			name: "WithoutJWKSEndpoint",
			setupMocks: func() {
				suite.mockOAuthService.On("GetOAuthClientConfig", testOIDCIDPID).Return(&oauth.OAuthClientConfig{
					OAuthEndpoints: oauth.OAuthEndpoints{},
				}, nil)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.mockOAuthService = oauthmock.NewOAuthAuthnServiceInterfaceMock(suite.T())
			suite.mockJWTService = jwtmock.NewJWTServiceInterfaceMock(suite.T())

			service := newOIDCAuthnService(suite.mockHTTPClient, suite.mockIdpService,
				suite.mockUserProvider, suite.mockJWTService)
			cast, ok := service.(*oidcAuthnService)
			suite.True(ok, "service is not of type *oidcAuthnService")
			suite.service = *cast
			suite.service.internal = suite.mockOAuthService

			tc.setupMocks()

			err := suite.service.ValidateIDToken(testOIDCIDPID, "valid_id_token")
			suite.Nil(err)
		})
	}
}

func (suite *OIDCAuthnServiceTestSuite) TestValidateIDTokenEmptyToken() {
	err := suite.service.ValidateIDToken(testOIDCIDPID, "")
	suite.NotNil(err)
	suite.Equal(ErrorInvalidIDToken.Code, err.Code)
}

func (suite *OIDCAuthnServiceTestSuite) TestGetIDTokenClaimsSuccess() {
	// Create a valid JWT token (base64 encoded header.payload.signature)
	idToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." +
		"eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ." +
		"SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

	claims, err := suite.service.GetIDTokenClaims(idToken)
	suite.Nil(err)
	suite.NotNil(claims)
	suite.Equal("1234567890", claims["sub"])
}

func (suite *OIDCAuthnServiceTestSuite) TestGetIDTokenClaimsEmptyToken() {
	claims, err := suite.service.GetIDTokenClaims("")
	suite.Nil(claims)
	suite.NotNil(err)
	suite.Equal(ErrorInvalidIDToken.Code, err.Code)
}

func (suite *OIDCAuthnServiceTestSuite) TestFetchUserInfoSuccess() {
	accessToken := "access_token"
	userInfo := map[string]interface{}{
		"sub":   "user123",
		"email": "user@example.com",
	}
	suite.mockOAuthService.On("FetchUserInfo", testOIDCIDPID, accessToken).Return(userInfo, nil)

	result, err := suite.service.FetchUserInfo(testOIDCIDPID, accessToken)
	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(userInfo["sub"], result["sub"])
}

func (suite *OIDCAuthnServiceTestSuite) TestGetInternalUserSuccess() {
	sub := "user123"
	user := &userprovider.User{
		UserID:   "user123",
		UserType: "person",
	}
	suite.mockOAuthService.On("GetInternalUser", sub).Return(user, nil)

	result, err := suite.service.GetInternalUser(sub)
	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(user.UserID, result.UserID)
}

func (suite *OIDCAuthnServiceTestSuite) TestExchangeCodeForTokenInternalError() {
	suite.mockOAuthService.On("ExchangeCodeForToken", testOIDCIDPID, "auth_code", false).
		Return(nil, &serviceerror.ServiceError{Code: "INT-ERR"})

	result, err := suite.service.ExchangeCodeForToken(testOIDCIDPID, "auth_code", false)
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal("INT-ERR", err.Code)
}

func (suite *OIDCAuthnServiceTestSuite) TestValidateTokenResponseValidateIDTokenFailure() {
	// Setup: Token response valid but ValidateIDToken will fail due to signature verification error
	suite.mockOAuthService = oauthmock.NewOAuthAuthnServiceInterfaceMock(suite.T())
	suite.mockJWTService = jwtmock.NewJWTServiceInterfaceMock(suite.T())

	service := newOIDCAuthnService(suite.mockHTTPClient, suite.mockIdpService,
		suite.mockUserProvider, suite.mockJWTService)
	cast, ok := service.(*oidcAuthnService)
	suite.True(ok)
	suite.service = *cast
	suite.service.internal = suite.mockOAuthService

	// GetOAuthClientConfig returns a config with jwks endpoint
	suite.mockOAuthService.On("GetOAuthClientConfig", testOIDCIDPID).Return(&oauth.OAuthClientConfig{
		OAuthEndpoints: oauth.OAuthEndpoints{JwksEndpoint: "https://example.com/jwks"},
	}, nil)

	// jwt service fails verification
	suite.mockJWTService.On("VerifyJWTWithJWKS", "id_token", "https://example.com/jwks", "", "").
		Return(&serviceerror.ServiceError{
			Type:             serviceerror.ServerErrorType,
			Code:             "SIGNATURE_INVALID",
			Error:            "Signature invalid",
			ErrorDescription: "signature invalid",
		})

	tokenResp := &oauth.TokenResponse{AccessToken: "access", IDToken: "id_token"}
	err := suite.service.ValidateTokenResponse(testOIDCIDPID, tokenResp, true)
	suite.NotNil(err)
	suite.Equal(ErrorInvalidIDTokenSignature.Code, err.Code)
}

func (suite *OIDCAuthnServiceTestSuite) TestNewOIDCAuthnService_HttpClientFallback() {
	svc := NewOIDCAuthnService(nil, suite.mockIdpService, suite.mockUserProvider, suite.mockJWTService)
	suite.NotNil(svc)
	// type assert to concrete to inspect metadata
	if concrete, ok := svc.(*oidcAuthnService); ok {
		meta := concrete.getMetadata()
		suite.Equal(authncm.AuthenticatorOIDC, meta.Name)
	}
}

func (suite *OIDCAuthnServiceTestSuite) TestGetIDTokenClaimsMalformedToken() {
	// Malformed token (not three parts) should return invalid token error
	claims, err := suite.service.GetIDTokenClaims("not.a.valid.token")
	suite.Nil(claims)
	suite.NotNil(err)
	suite.Equal(ErrorInvalidIDToken.Code, err.Code)
}

func (suite *OIDCAuthnServiceTestSuite) TestValidateIDTokenWithJWKSEndpoint() {
	// Test that JWKS endpoint is used when configured
	idToken := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyMTIzI" +
		"iwibmFtZSI6IlRlc3QgVXNlciIsImlhdCI6MTUxNjIzOTAyMn0.signature"

	config := &oauth.OAuthClientConfig{
		ClientID:     "test_client",
		ClientSecret: "test_secret",
		RedirectURI:  "https://app.com/callback",
		Scopes:       []string{"openid"},
		OAuthEndpoints: oauth.OAuthEndpoints{
			JwksEndpoint: "https://idp.com/jwks",
		},
	}

	suite.mockOAuthService.On("GetOAuthClientConfig", testOIDCIDPID).Return(config, nil)
	suite.mockJWTService.On("VerifyJWTWithJWKS", idToken, "https://idp.com/jwks", "", "").Return(nil)

	err := suite.service.ValidateIDToken(testOIDCIDPID, idToken)
	suite.Nil(err)
}

func (suite *OIDCAuthnServiceTestSuite) TestValidateIDTokenWithoutJWKSEndpoint() {
	// Test that validation succeeds when JWKS endpoint is not configured (skips signature validation)
	idToken := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyMTIzIiw" +
		"ibmFtZSI6IlRlc3QgVXNlciIsImlhdCI6MTUxNjIzOTAyMn0.signature"

	config := &oauth.OAuthClientConfig{
		ClientID:     "test_client",
		ClientSecret: "test_secret",
		RedirectURI:  "https://app.com/callback",
		Scopes:       []string{"openid"},
		OAuthEndpoints: oauth.OAuthEndpoints{
			JwksEndpoint: "", // Empty JWKS endpoint
		},
	}

	suite.mockOAuthService.On("GetOAuthClientConfig", testOIDCIDPID).Return(config, nil)
	// VerifyJWTWithJWKS should not be called when JWKS endpoint is empty

	err := suite.service.ValidateIDToken(testOIDCIDPID, idToken)
	suite.Nil(err)
}
