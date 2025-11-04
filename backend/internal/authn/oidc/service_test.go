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

	"github.com/asgardeo/thunder/internal/authn/oauth"
	"github.com/asgardeo/thunder/internal/idp"
	"github.com/asgardeo/thunder/internal/system/cmodels"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/user"
	"github.com/asgardeo/thunder/tests/mocks/authn/oauthmock"
	"github.com/asgardeo/thunder/tests/mocks/idp/idpmock"
	"github.com/asgardeo/thunder/tests/mocks/jwtmock"
)

const (
	testOIDCIDPID = "idp123"
)

type OIDCAuthnServiceTestSuite struct {
	suite.Suite
	mockOAuthService *oauthmock.OAuthAuthnServiceInterfaceMock
	mockIDPService   *idpmock.IDPServiceInterfaceMock
	mockJWTService   *jwtmock.JWTServiceInterfaceMock
	service          OIDCAuthnServiceInterface
}

func TestOIDCAuthnServiceTestSuite(t *testing.T) {
	suite.Run(t, new(OIDCAuthnServiceTestSuite))
}

func (suite *OIDCAuthnServiceTestSuite) SetupTest() {
	suite.mockOAuthService = oauthmock.NewOAuthAuthnServiceInterfaceMock(suite.T())
	suite.mockIDPService = idpmock.NewIDPServiceInterfaceMock(suite.T())
	suite.mockJWTService = jwtmock.NewJWTServiceInterfaceMock(suite.T())
	suite.service = NewOIDCAuthnServiceWithIDPService(suite.mockOAuthService, suite.mockIDPService, suite.mockJWTService)
}

func createTestOIDCIDPDTO(idpID string) *idp.IDPDTO {
	clientIDProp, _ := cmodels.NewProperty("client_id", "test_client", false)
	clientSecretProp, _ := cmodels.NewProperty("client_secret", "test_secret", false)
	redirectURIProp, _ := cmodels.NewProperty("redirect_uri", "https://app.com/callback", false)
	scopesProp, _ := cmodels.NewProperty("scopes", "openid profile", false)
	authzEndpointProp, _ := cmodels.NewProperty("authorization_endpoint", "https://idp.com/authorize", false)
	tokenEndpointProp, _ := cmodels.NewProperty("token_endpoint", "https://idp.com/token", false)

	return &idp.IDPDTO{
		ID:   idpID,
		Name: "Test OIDC IDP",
		Type: idp.IDPTypeOIDC,
		Properties: []cmodels.Property{
			*clientIDProp, *clientSecretProp, *redirectURIProp, *scopesProp,
			*authzEndpointProp, *tokenEndpointProp,
		},
	}
}

func (suite *OIDCAuthnServiceTestSuite) TestGetOAuthClientConfigWithOpenIDScope() {
	idpID := testOIDCIDPID
	idpDTO := createTestOIDCIDPDTO(idpID)
	
	suite.mockIDPService.On("GetIdentityProvider", idpID).Return(idpDTO, nil)

	result, err := suite.service.GetOAuthClientConfig(idpID)
	suite.Nil(err)
	suite.NotNil(result)

	suite.Contains(result.Scopes, "openid")
	suite.Contains(result.Scopes, "profile")
}

func (suite *OIDCAuthnServiceTestSuite) TestGetOAuthClientConfigWithoutOpenIDScope() {
	idpID := testOIDCIDPID
	clientIDProp, _ := cmodels.NewProperty("client_id", "test_client", false)
	clientSecretProp, _ := cmodels.NewProperty("client_secret", "test_secret", false)
	redirectURIProp, _ := cmodels.NewProperty("redirect_uri", "https://app.com/callback", false)
	scopesProp, _ := cmodels.NewProperty("scopes", "profile", false)
	authzEndpointProp, _ := cmodels.NewProperty("authorization_endpoint", "https://idp.com/authorize", false)
	tokenEndpointProp, _ := cmodels.NewProperty("token_endpoint", "https://idp.com/token", false)

	idpDTO := &idp.IDPDTO{
		ID:   idpID,
		Name: "Test OIDC IDP",
		Type: idp.IDPTypeOIDC,
		Properties: []cmodels.Property{
			*clientIDProp, *clientSecretProp, *redirectURIProp, *scopesProp,
			*authzEndpointProp, *tokenEndpointProp,
		},
	}
	
	suite.mockIDPService.On("GetIdentityProvider", idpID).Return(idpDTO, nil)

	result, err := suite.service.GetOAuthClientConfig(idpID)
	suite.Nil(err)
	suite.NotNil(result)

	suite.Contains(result.Scopes, "openid")
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
				
				clientIDProp, _ := cmodels.NewProperty("client_id", "test_client", false)
				clientSecretProp, _ := cmodels.NewProperty("client_secret", "test_secret", false)
				redirectURIProp, _ := cmodels.NewProperty("redirect_uri", "https://app.com/callback", false)
				scopesProp, _ := cmodels.NewProperty("scopes", "openid profile", false)
				authzEndpointProp, _ := cmodels.NewProperty("authorization_endpoint", "https://idp.com/authorize", false)
				tokenEndpointProp, _ := cmodels.NewProperty("token_endpoint", "https://idp.com/token", false)
				jwksEndpointProp, _ := cmodels.NewProperty("jwks_endpoint", "https://example.com/jwks", false)

				idpDTO := &idp.IDPDTO{
					ID:   testOIDCIDPID,
					Name: "Test OIDC IDP",
					Type: idp.IDPTypeOIDC,
					Properties: []cmodels.Property{
						*clientIDProp, *clientSecretProp, *redirectURIProp, *scopesProp,
						*authzEndpointProp, *tokenEndpointProp, *jwksEndpointProp,
					},
				}
				
				suite.mockIDPService.On("GetIdentityProvider", testOIDCIDPID).Return(idpDTO, nil)
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
			suite.mockIDPService = idpmock.NewIDPServiceInterfaceMock(suite.T())
			suite.mockJWTService = jwtmock.NewJWTServiceInterfaceMock(suite.T())
			suite.service = NewOIDCAuthnServiceWithIDPService(suite.mockOAuthService, suite.mockIDPService, suite.mockJWTService)

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
				clientIDProp, _ := cmodels.NewProperty("client_id", "test_client", false)
				clientSecretProp, _ := cmodels.NewProperty("client_secret", "test_secret", false)
				redirectURIProp, _ := cmodels.NewProperty("redirect_uri", "https://app.com/callback", false)
				scopesProp, _ := cmodels.NewProperty("scopes", "openid profile", false)
				authzEndpointProp, _ := cmodels.NewProperty("authorization_endpoint", "https://idp.com/authorize", false)
				tokenEndpointProp, _ := cmodels.NewProperty("token_endpoint", "https://idp.com/token", false)
				jwksEndpointProp, _ := cmodels.NewProperty("jwks_endpoint", "https://example.com/jwks", false)

				idpDTO := &idp.IDPDTO{
					ID:   testOIDCIDPID,
					Name: "Test OIDC IDP",
					Type: idp.IDPTypeOIDC,
					Properties: []cmodels.Property{
						*clientIDProp, *clientSecretProp, *redirectURIProp, *scopesProp,
						*authzEndpointProp, *tokenEndpointProp, *jwksEndpointProp,
					},
				}
				
				suite.mockIDPService.On("GetIdentityProvider", testOIDCIDPID).Return(idpDTO, nil)
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
			suite.mockIDPService = idpmock.NewIDPServiceInterfaceMock(suite.T())
			suite.mockJWTService = jwtmock.NewJWTServiceInterfaceMock(suite.T())
			suite.service = NewOIDCAuthnServiceWithIDPService(suite.mockOAuthService, suite.mockIDPService, suite.mockJWTService)

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
				clientIDProp, _ := cmodels.NewProperty("client_id", "test_client", false)
				clientSecretProp, _ := cmodels.NewProperty("client_secret", "test_secret", false)
				redirectURIProp, _ := cmodels.NewProperty("redirect_uri", "https://app.com/callback", false)
				scopesProp, _ := cmodels.NewProperty("scopes", "openid profile", false)
				authzEndpointProp, _ := cmodels.NewProperty("authorization_endpoint", "https://idp.com/authorize", false)
				tokenEndpointProp, _ := cmodels.NewProperty("token_endpoint", "https://idp.com/token", false)
				jwksEndpointProp, _ := cmodels.NewProperty("jwks_endpoint", "https://example.com/jwks", false)

				idpDTO := &idp.IDPDTO{
					ID:   testOIDCIDPID,
					Name: "Test OIDC IDP",
					Type: idp.IDPTypeOIDC,
					Properties: []cmodels.Property{
						*clientIDProp, *clientSecretProp, *redirectURIProp, *scopesProp,
						*authzEndpointProp, *tokenEndpointProp, *jwksEndpointProp,
					},
				}
				
				suite.mockIDPService.On("GetIdentityProvider", testOIDCIDPID).Return(idpDTO, nil)
				suite.mockJWTService.On("VerifyJWTWithJWKS", "valid_id_token",
					"https://example.com/jwks", "", "").Return(nil)
			},
		},
		{
			name: "WithoutJWKSEndpoint",
			setupMocks: func() {
				idpDTO := createTestOIDCIDPDTO(testOIDCIDPID)
				suite.mockIDPService.On("GetIdentityProvider", testOIDCIDPID).Return(idpDTO, nil)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.mockOAuthService = oauthmock.NewOAuthAuthnServiceInterfaceMock(suite.T())
			suite.mockIDPService = idpmock.NewIDPServiceInterfaceMock(suite.T())
			suite.mockJWTService = jwtmock.NewJWTServiceInterfaceMock(suite.T())
			suite.service = NewOIDCAuthnServiceWithIDPService(suite.mockOAuthService, suite.mockIDPService, suite.mockJWTService)

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
	user := &user.User{
		ID:   "user123",
		Type: "person",
	}
	suite.mockOAuthService.On("GetInternalUser", sub).Return(user, nil)

	result, err := suite.service.GetInternalUser(sub)
	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(user.ID, result.ID)
}

func (suite *OIDCAuthnServiceTestSuite) TestGetOAuthClientConfigWithoutUserInfoEndpoint() {
	idpID := testOIDCIDPID
	clientIDProp, _ := cmodels.NewProperty("client_id", "test_client", false)
	clientSecretProp, _ := cmodels.NewProperty("client_secret", "test_secret", false)
	redirectURIProp, _ := cmodels.NewProperty("redirect_uri", "https://app.com/callback", false)
	scopesProp, _ := cmodels.NewProperty("scopes", "openid profile", false)
	authzEndpointProp, _ := cmodels.NewProperty("authorization_endpoint", "https://idp.com/authorize", false)
	tokenEndpointProp, _ := cmodels.NewProperty("token_endpoint", "https://idp.com/token", false)

	idpDTO := &idp.IDPDTO{
		ID:   idpID,
		Name: "Test OIDC IDP",
		Type: idp.IDPTypeOIDC,
		Properties: []cmodels.Property{
			*clientIDProp, *clientSecretProp, *redirectURIProp, *scopesProp,
			*authzEndpointProp, *tokenEndpointProp,
		},
	}
	
	suite.mockIDPService.On("GetIdentityProvider", idpID).Return(idpDTO, nil)

	result, err := suite.service.GetOAuthClientConfig(idpID)
	suite.Nil(err)
	suite.NotNil(result)
	suite.Contains(result.Scopes, "openid")
	suite.Contains(result.Scopes, "profile")
	suite.Equal("", result.OAuthEndpoints.UserInfoEndpoint)
}
