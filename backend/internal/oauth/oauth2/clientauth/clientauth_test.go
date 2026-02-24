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

package clientauth

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/cert"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/discovery"
	"github.com/asgardeo/thunder/internal/system/crypto/hash"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/tests/mocks/applicationmock"
	"github.com/asgardeo/thunder/tests/mocks/jose/jwtmock"
	"github.com/asgardeo/thunder/tests/mocks/oauth/oauth2/discoverymock"
)

const (
	testClientID     = "test-client-id"
	testClientSecret = "test-secret"
)

type ClientAuthTestSuite struct {
	suite.Suite
	mockAppService       *applicationmock.ApplicationServiceInterfaceMock
	mockJwtService       *jwtmock.JWTServiceInterfaceMock
	mockDiscoveryService *discoverymock.DiscoveryServiceInterfaceMock
}

func TestClientAuthTestSuite(t *testing.T) {
	suite.Run(t, new(ClientAuthTestSuite))
}

func (suite *ClientAuthTestSuite) SetupTest() {
	suite.mockAppService = applicationmock.NewApplicationServiceInterfaceMock(suite.T())
	suite.mockJwtService = jwtmock.NewJWTServiceInterfaceMock(suite.T())
	suite.mockDiscoveryService = discoverymock.NewDiscoveryServiceInterfaceMock(suite.T())
}

func (suite *ClientAuthTestSuite) TestAuthenticate_Success_ClientSecretPost() {
	clientSecret := testClientSecret
	hashedSecret := hash.GenerateThumbprintFromString(clientSecret)
	mockApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:                testClientID,
		HashedClientSecret:      hashedSecret,
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
	}

	suite.mockAppService.On("GetOAuthApplication", testClientID).
		Return(mockApp, nil).Once()

	formData := url.Values{}
	formData.Set("client_id", testClientID)
	formData.Set("client_secret", clientSecret)

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()

	clientInfo, authErr := authenticate(req, suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	assert.Nil(suite.T(), authErr)
	assert.NotNil(suite.T(), clientInfo)
	if clientInfo != nil {
		assert.Equal(suite.T(), testClientID, clientInfo.ClientID)
		assert.Equal(suite.T(), clientSecret, clientInfo.ClientSecret)
		assert.NotNil(suite.T(), clientInfo.OAuthApp)
		assert.Equal(suite.T(), testClientID, clientInfo.OAuthApp.ClientID)
	}
}

func (suite *ClientAuthTestSuite) TestAuthenticate_Success_ClientSecretBasic() {
	clientSecret := testClientSecret
	hashedSecret := hash.GenerateThumbprintFromString(clientSecret)
	mockApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:                testClientID,
		HashedClientSecret:      hashedSecret,
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretBasic,
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
	}

	suite.mockAppService.On("GetOAuthApplication", testClientID).
		Return(mockApp, nil).Once()

	req, _ := http.NewRequest("POST", "/test", nil)
	req.SetBasicAuth(testClientID, clientSecret)

	clientInfo, authErr := authenticate(req, suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	assert.Nil(suite.T(), authErr)
	assert.NotNil(suite.T(), clientInfo)
	if clientInfo != nil {
		assert.Equal(suite.T(), testClientID, clientInfo.ClientID)
		assert.Equal(suite.T(), clientSecret, clientInfo.ClientSecret)
	}
}

func (suite *ClientAuthTestSuite) TestAuthenticate_Success_PublicClient() {
	mockApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:                "public-client-id",
		HashedClientSecret:      "",
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodNone,
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
		PublicClient:            true,
	}

	suite.mockAppService.On("GetOAuthApplication", "public-client-id").
		Return(mockApp, nil).Once()

	formData := url.Values{}
	formData.Set("client_id", "public-client-id")

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()

	clientInfo, authErr := authenticate(req, suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	assert.Nil(suite.T(), authErr)
	assert.NotNil(suite.T(), clientInfo)
	if clientInfo != nil {
		assert.Equal(suite.T(), "public-client-id", clientInfo.ClientID)
		assert.Equal(suite.T(), "", clientInfo.ClientSecret)
	}
}

func (suite *ClientAuthTestSuite) TestAuthenticate_MissingClientID() {
	req, _ := http.NewRequest("POST", "/test", nil)
	_ = req.ParseForm()

	clientInfo, authErr := authenticate(req, suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errMissingClientID, authErr)
	assert.Equal(suite.T(), constants.ErrorInvalidClient, authErr.ErrorCode)
	assert.Equal(suite.T(), "Missing client_id parameter", authErr.ErrorDescription)
}

func (suite *ClientAuthTestSuite) TestAuthenticate_MissingClientSecret() {
	formData := url.Values{}
	formData.Set("client_id", testClientID)

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()

	// This should succeed for public clients, but fail for confidential clients
	// Since we don't have an app yet, it will fail at app retrieval
	suite.mockAppService.On("GetOAuthApplication", testClientID).
		Return(nil, nil).Once()

	clientInfo, authErr := authenticate(req, suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errInvalidClientCredentials, authErr)
}

func (suite *ClientAuthTestSuite) TestAuthenticate_InvalidBasicAuth() {
	req, _ := http.NewRequest("POST", "/test", nil)
	req.Header.Set("Authorization", "Basic invalid_base64")

	clientInfo, authErr := authenticate(req, suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errInvalidAuthorizationHeader, authErr)
}

func (suite *ClientAuthTestSuite) TestAuthenticate_InvalidAuthorizationHeader() {
	req, _ := http.NewRequest("POST", "/test", nil)
	req.Header.Set("Authorization", "Bearer token")

	clientInfo, authErr := authenticate(req, suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errInvalidAuthorizationHeader, authErr)
	assert.NotNil(suite.T(), authErr.ResponseHeaders)
	assert.Equal(suite.T(), "Basic", authErr.ResponseHeaders["WWW-Authenticate"])
}

func (suite *ClientAuthTestSuite) TestAuthenticate_BothHeaderAndBody() {
	formData := url.Values{}
	formData.Set("client_id", testClientID)
	formData.Set("client_secret", testClientSecret)

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(testClientID, testClientSecret)
	_ = req.ParseForm()

	clientInfo, authErr := authenticate(req, suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errMultipleAuthMethods, authErr)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, authErr.ErrorCode)
}

func (suite *ClientAuthTestSuite) TestAuthenticate_ClientNotFound() {
	formData := url.Values{}
	formData.Set("client_id", "non-existent-client")
	formData.Set("client_secret", testClientSecret)

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()

	suite.mockAppService.On("GetOAuthApplication", "non-existent-client").
		Return(nil, nil).Once()

	clientInfo, authErr := authenticate(req, suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errInvalidClientCredentials, authErr)
}

func (suite *ClientAuthTestSuite) TestAuthenticate_InvalidClientSecret() {
	clientSecret := testClientSecret
	wrongSecret := "wrong-secret"
	hashedSecret := hash.GenerateThumbprintFromString(clientSecret)
	mockApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:                testClientID,
		HashedClientSecret:      hashedSecret,
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
	}

	suite.mockAppService.On("GetOAuthApplication", testClientID).
		Return(mockApp, nil).Once()

	formData := url.Values{}
	formData.Set("client_id", testClientID)
	formData.Set("client_secret", wrongSecret)

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()

	clientInfo, authErr := authenticate(req, suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errInvalidClientCredentials, authErr)
}

func (suite *ClientAuthTestSuite) TestAuthenticate_WrongAuthMethod() {
	clientSecret := testClientSecret
	hashedSecret := hash.GenerateThumbprintFromString(clientSecret)
	mockApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:                testClientID,
		HashedClientSecret:      hashedSecret,
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
	}

	suite.mockAppService.On("GetOAuthApplication", testClientID).
		Return(mockApp, nil).Once()

	// Try to use client_secret_basic when app only allows client_secret_post
	req, _ := http.NewRequest("POST", "/test", nil)
	req.SetBasicAuth(testClientID, clientSecret)

	clientInfo, authErr := authenticate(req, suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errUnauthorizedAuthMethod, authErr)
	assert.Equal(suite.T(), constants.ErrorUnauthorizedClient, authErr.ErrorCode)
}

func (suite *ClientAuthTestSuite) TestAuthenticate_PublicClientWithSecret() {
	mockApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:                "public-client-id",
		HashedClientSecret:      "",
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodNone,
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
		PublicClient:            true,
	}

	suite.mockAppService.On("GetOAuthApplication", "public-client-id").
		Return(mockApp, nil).Once()

	formData := url.Values{}
	formData.Set("client_id", "public-client-id")
	formData.Set("client_secret", "some-secret")

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()

	// Try to use client_secret_post with public client
	clientInfo, authErr := authenticate(req, suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errUnauthorizedAuthMethod, authErr)
}

func (suite *ClientAuthTestSuite) TestAuthenticate_PublicClientMissingSecret() {
	mockApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:                "public-client-id",
		HashedClientSecret:      "",
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodNone,
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
		PublicClient:            true,
	}

	suite.mockAppService.On("GetOAuthApplication", "public-client-id").
		Return(mockApp, nil).Once()

	formData := url.Values{}
	formData.Set("client_id", "public-client-id")

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()

	// Public client with authMethod = none should succeed
	clientInfo, authErr := authenticate(req, suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	assert.Nil(suite.T(), authErr)
	assert.NotNil(suite.T(), clientInfo)
}

func (suite *ClientAuthTestSuite) TestAuthenticate_ClientIDMismatch() {
	clientSecret := testClientSecret
	hashedSecret := hash.GenerateThumbprintFromString(clientSecret)
	mockApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:                "different-client-id",
		HashedClientSecret:      hashedSecret,
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
	}

	suite.mockAppService.On("GetOAuthApplication", testClientID).
		Return(mockApp, nil).Once()

	formData := url.Values{}
	formData.Set("client_id", testClientID)
	formData.Set("client_secret", clientSecret)

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()

	clientInfo, authErr := authenticate(req, suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errInvalidClientCredentials, authErr)
}

func (suite *ClientAuthTestSuite) TestAuthenticate_ServiceError() {
	serviceErr := &serviceerror.ServiceError{
		Code:             "APP-5001",
		Type:             serviceerror.ServerErrorType,
		Error:            "server_error",
		ErrorDescription: "Internal server error",
	}

	suite.mockAppService.On("GetOAuthApplication", testClientID).
		Return(nil, serviceErr).Once()

	formData := url.Values{}
	formData.Set("client_id", testClientID)
	formData.Set("client_secret", testClientSecret)

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()

	clientInfo, authErr := authenticate(req, suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errInvalidClientCredentials, authErr)
}

// buildTestJWT constructs a fake JWT string (header.payload.signature) for testing purposes.
// It accepts header claims and payload claims as maps.
func buildTestJWT(headerClaims, payloadClaims map[string]any) string {
	headerJSON, _ := json.Marshal(headerClaims)
	payloadJSON, _ := json.Marshal(payloadClaims)
	header := base64.RawURLEncoding.EncodeToString(headerJSON)
	payload := base64.RawURLEncoding.EncodeToString(payloadJSON)
	signature := base64.RawURLEncoding.EncodeToString([]byte("fake-signature"))
	return header + "." + payload + "." + signature
}

// buildTestRSAJWKS generates an RSA key pair and returns the JWKS JSON string for the given kid.
func buildTestRSAJWKS(kid string) string {
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	pubKey := &privKey.PublicKey
	nBytes := pubKey.N.Bytes()
	eBytes := big.NewInt(int64(pubKey.E)).Bytes()
	jwk := map[string]any{
		"kty": "RSA",
		"kid": kid,
		"n":   base64.RawURLEncoding.EncodeToString(nBytes),
		"e":   base64.RawURLEncoding.EncodeToString(eBytes),
	}
	jwks := map[string]any{"keys": []map[string]any{jwk}}
	jwksJSON, _ := json.Marshal(jwks)
	return string(jwksJSON)
}

// buildFakeJWTWithSub constructs a fake JWT with a given subject and kid in the header.
func buildFakeJWTWithSub(subject string) string {
	return buildTestJWT(
		map[string]any{"alg": "RS256", "kid": "test-kid", "typ": "JWT"},
		map[string]any{"sub": subject, "aud": "https://token"},
	)
}

// buildFakeJWTWithPayload constructs a fake JWT string with a custom payload for testing purposes.
func buildFakeJWTWithPayload(payloadJSON string) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","kid":"test-kid","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(payloadJSON))
	signature := base64.RawURLEncoding.EncodeToString([]byte("fake-signature"))
	return header + "." + payload + "." + signature
}

func (suite *ClientAuthTestSuite) TestAuthenticate_Success_PrivateKeyJWT() {
	jwksJSON := buildTestRSAJWKS("test-kid")
	assertion := buildFakeJWTWithSub(testClientID)
	mockApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:                testClientID,
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodPrivateKeyJWT,
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
		Certificate:             &appmodel.ApplicationCertificate{Value: jwksJSON},
	}

	suite.mockAppService.On("GetOAuthApplication", testClientID).
		Return(mockApp, nil).Once()
	suite.mockDiscoveryService.EXPECT().
		GetOAuth2AuthorizationServerMetadata().
		Return(&discovery.OAuth2AuthorizationServerMetadata{
			TokenEndpoint: "https://localhost:9443/oauth2/token",
		})
	suite.mockJwtService.EXPECT().
		VerifyJWTWithPublicKey(assertion, mock.Anything, "https://localhost:9443/oauth2/token", testClientID).
		Return(nil)

	formData := url.Values{}
	formData.Set("client_assertion_type", constants.SupportedClientAssertionType)
	formData.Set("client_assertion", assertion)

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()

	clientInfo, authErr := authenticate(req, suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	assert.Nil(suite.T(), authErr)
	assert.NotNil(suite.T(), clientInfo)
	if clientInfo != nil {
		assert.Equal(suite.T(), testClientID, clientInfo.ClientID)
		assert.NotNil(suite.T(), clientInfo.OAuthApp)
		assert.Equal(suite.T(), testClientID, clientInfo.OAuthApp.ClientID)
	}
}

func (suite *ClientAuthTestSuite) TestAuthenticate_Success_PrivateKeyJWT_WithClientIDInBody() {
	jwksJSON := buildTestRSAJWKS("test-kid")
	assertion := buildFakeJWTWithSub(testClientID)
	mockApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:                testClientID,
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodPrivateKeyJWT,
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
		Certificate:             &appmodel.ApplicationCertificate{Value: jwksJSON},
	}

	suite.mockAppService.On("GetOAuthApplication", testClientID).
		Return(mockApp, nil).Once()
	suite.mockDiscoveryService.EXPECT().
		GetOAuth2AuthorizationServerMetadata().
		Return(&discovery.OAuth2AuthorizationServerMetadata{
			TokenEndpoint: "https://localhost:9443/oauth2/token",
		})
	suite.mockJwtService.EXPECT().
		VerifyJWTWithPublicKey(assertion, mock.Anything, "https://localhost:9443/oauth2/token", testClientID).
		Return(nil)

	formData := url.Values{}
	formData.Set("client_id", testClientID)
	formData.Set("client_assertion_type", constants.SupportedClientAssertionType)
	formData.Set("client_assertion", assertion)

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()

	clientInfo, authErr := authenticate(req, suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	assert.Nil(suite.T(), authErr)
	assert.NotNil(suite.T(), clientInfo)
	if clientInfo != nil {
		assert.Equal(suite.T(), testClientID, clientInfo.ClientID)
	}
}

func (suite *ClientAuthTestSuite) TestAuthenticate_PrivateKeyJWT_UnsupportedAssertionType() {
	formData := url.Values{}
	formData.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:saml2-bearer")
	formData.Set("client_assertion", "some-assertion")

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()

	clientInfo, authErr := authenticate(req, suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errInvalidClientAssertion, authErr)
	assert.Equal(suite.T(), constants.ErrorInvalidClient, authErr.ErrorCode)
	assert.Equal(suite.T(), "Invalid client assertion", authErr.ErrorDescription)
}

func (suite *ClientAuthTestSuite) TestAuthenticate_PrivateKeyJWT_OnlyAssertionTypeProvided() {
	formData := url.Values{}
	formData.Set("client_assertion_type", constants.SupportedClientAssertionType)

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()

	clientInfo, authErr := authenticate(req, suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errInvalidClientAssertion, authErr)
	assert.Equal(suite.T(), constants.ErrorInvalidClient, authErr.ErrorCode)
	assert.Equal(suite.T(), "Invalid client assertion", authErr.ErrorDescription)
}

func (suite *ClientAuthTestSuite) TestAuthenticate_PrivateKeyJWT_OnlyAssertionProvided() {
	assertion := buildFakeJWTWithSub(testClientID)

	formData := url.Values{}
	formData.Set("client_assertion", assertion)

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()

	// Only client_assertion without client_assertion_type: assertion_type is empty,
	// but the code detects it as private_key_jwt since client_assertion is present.
	// Then it checks assertion_type != SupportedClientAssertionType, which fails.
	clientInfo, authErr := authenticate(req, suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errInvalidClientAssertion, authErr)
	assert.Equal(suite.T(), constants.ErrorInvalidClient, authErr.ErrorCode)
	assert.Equal(suite.T(), "Invalid client assertion", authErr.ErrorDescription)
}

func (suite *ClientAuthTestSuite) TestAuthenticate_PrivateKeyJWT_InvalidJWTFormat() {
	formData := url.Values{}
	formData.Set("client_assertion_type", constants.SupportedClientAssertionType)
	formData.Set("client_assertion", "not-a-valid-jwt")

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()

	clientInfo, authErr := authenticate(req, suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errInvalidClientAssertion, authErr)
	assert.Equal(suite.T(), constants.ErrorInvalidClient, authErr.ErrorCode)
	assert.Equal(suite.T(), "Invalid client assertion", authErr.ErrorDescription)
}

func (suite *ClientAuthTestSuite) TestAuthenticate_PrivateKeyJWT_MissingSubClaim() {
	// Build a JWT with no sub claim
	assertion := buildFakeJWTWithPayload(`{"aud":"https://token","iss":"some-issuer"}`)

	formData := url.Values{}
	formData.Set("client_assertion_type", constants.SupportedClientAssertionType)
	formData.Set("client_assertion", assertion)

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()

	clientInfo, authErr := authenticate(req, suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errInvalidClientAssertion, authErr)
	assert.Equal(suite.T(), constants.ErrorInvalidClient, authErr.ErrorCode)
	assert.Equal(suite.T(), "Invalid client assertion", authErr.ErrorDescription)
}

func (suite *ClientAuthTestSuite) TestAuthenticate_PrivateKeyJWT_EmptySubClaim() {
	// Build a JWT with empty sub claim
	assertion := buildFakeJWTWithPayload(`{"sub":"","aud":"https://token"}`)

	formData := url.Values{}
	formData.Set("client_assertion_type", constants.SupportedClientAssertionType)
	formData.Set("client_assertion", assertion)

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()

	clientInfo, authErr := authenticate(req, suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errInvalidClientAssertion, authErr)
	assert.Equal(suite.T(), constants.ErrorInvalidClient, authErr.ErrorCode)
	assert.Equal(suite.T(), "Invalid client assertion", authErr.ErrorDescription)
}

func (suite *ClientAuthTestSuite) TestAuthenticate_PrivateKeyJWT_NonStringSubClaim() {
	// Build a JWT with non-string sub claim
	assertion := buildFakeJWTWithPayload(`{"sub":12345,"aud":"https://token"}`)

	formData := url.Values{}
	formData.Set("client_assertion_type", constants.SupportedClientAssertionType)
	formData.Set("client_assertion", assertion)

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()

	clientInfo, authErr := authenticate(req, suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errInvalidClientAssertion, authErr)
	assert.Equal(suite.T(), constants.ErrorInvalidClient, authErr.ErrorCode)
	assert.Equal(suite.T(), "Invalid client assertion", authErr.ErrorDescription)
}

func (suite *ClientAuthTestSuite) TestAuthenticate_PrivateKeyJWT_ClientNotFound() {
	assertion := buildFakeJWTWithSub("unknown-client")

	suite.mockAppService.On("GetOAuthApplication", "unknown-client").
		Return(nil, nil).Once()

	formData := url.Values{}
	formData.Set("client_assertion_type", constants.SupportedClientAssertionType)
	formData.Set("client_assertion", assertion)

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()

	clientInfo, authErr := authenticate(req, suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errInvalidClientCredentials, authErr)
}

func (suite *ClientAuthTestSuite) TestAuthenticate_PrivateKeyJWT_AuthMethodNotAllowed() {
	assertion := buildFakeJWTWithSub(testClientID)
	// App only allows client_secret_post, not private_key_jwt
	mockApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:                testClientID,
		HashedClientSecret:      hash.GenerateThumbprintFromString(testClientSecret),
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
	}

	suite.mockAppService.On("GetOAuthApplication", testClientID).
		Return(mockApp, nil).Once()

	formData := url.Values{}
	formData.Set("client_assertion_type", constants.SupportedClientAssertionType)
	formData.Set("client_assertion", assertion)

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()

	clientInfo, authErr := authenticate(req, suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errUnauthorizedAuthMethod, authErr)
	assert.Equal(suite.T(), constants.ErrorUnauthorizedClient, authErr.ErrorCode)
}

func (suite *ClientAuthTestSuite) TestAuthenticate_PrivateKeyJWT_AssertionValidationFails() {
	assertion := buildFakeJWTWithSub(testClientID)
	mockApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:                testClientID,
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodPrivateKeyJWT,
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
		// No certificate configured, so ValidateClientAssertion will return false
	}

	suite.mockAppService.On("GetOAuthApplication", testClientID).
		Return(mockApp, nil).Once()

	formData := url.Values{}
	formData.Set("client_assertion_type", constants.SupportedClientAssertionType)
	formData.Set("client_assertion", assertion)

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()

	clientInfo, authErr := authenticate(req, suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errInvalidClientAssertion, authErr)
	assert.Equal(suite.T(), constants.ErrorInvalidClient, authErr.ErrorCode)
	assert.Equal(suite.T(), "Invalid client assertion", authErr.ErrorDescription)
}

func (suite *ClientAuthTestSuite) TestAuthenticate_PrivateKeyJWT_ClientIDMismatch() {
	assertion := buildFakeJWTWithSub("different-client-id")

	formData := url.Values{}
	formData.Set("client_id", testClientID)
	formData.Set("client_assertion_type", constants.SupportedClientAssertionType)
	formData.Set("client_assertion", assertion)

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()

	clientInfo, authErr := authenticate(req, suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errClientIDMismatch, authErr)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, authErr.ErrorCode)
}

func (suite *ClientAuthTestSuite) TestAuthenticate_PrivateKeyJWT_WithBasicAuth_MultipleAuthMethods() {
	assertion := buildFakeJWTWithSub(testClientID)

	formData := url.Values{}
	formData.Set("client_assertion_type", constants.SupportedClientAssertionType)
	formData.Set("client_assertion", assertion)

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(testClientID, testClientSecret)
	_ = req.ParseForm()

	clientInfo, authErr := authenticate(req, suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errMultipleAuthMethods, authErr)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, authErr.ErrorCode)
}

func (suite *ClientAuthTestSuite) TestAuthenticate_PrivateKeyJWT_WithClientSecret_MultipleAuthMethods() {
	assertion := buildFakeJWTWithSub(testClientID)

	formData := url.Values{}
	formData.Set("client_id", testClientID)
	formData.Set("client_secret", testClientSecret)
	formData.Set("client_assertion_type", constants.SupportedClientAssertionType)
	formData.Set("client_assertion", assertion)

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()

	clientInfo, authErr := authenticate(req, suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errMultipleAuthMethods, authErr)
	assert.Equal(suite.T(), constants.ErrorInvalidRequest, authErr.ErrorCode)
}

func (suite *ClientAuthTestSuite) TestAuthenticate_PrivateKeyJWT_ServiceError() {
	assertion := buildFakeJWTWithSub(testClientID)

	serviceErr := &serviceerror.ServiceError{
		Code:             "APP-5001",
		Type:             serviceerror.ServerErrorType,
		Error:            "server_error",
		ErrorDescription: "Internal server error",
	}

	suite.mockAppService.On("GetOAuthApplication", testClientID).
		Return(nil, serviceErr).Once()

	formData := url.Values{}
	formData.Set("client_assertion_type", constants.SupportedClientAssertionType)
	formData.Set("client_assertion", assertion)

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()

	clientInfo, authErr := authenticate(req, suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errInvalidClientCredentials, authErr)
}

func (suite *ClientAuthTestSuite) TestAuthenticate_PrivateKeyJWT_InvalidBase64Payload() {
	// Build a JWT with invalid base64 in the payload segment
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256"}`))
	signature := base64.RawURLEncoding.EncodeToString([]byte("fake-signature"))
	invalidJWT := header + ".!!!invalid-base64!!!." + signature

	formData := url.Values{}
	formData.Set("client_assertion_type", constants.SupportedClientAssertionType)
	formData.Set("client_assertion", invalidJWT)

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()

	clientInfo, authErr := authenticate(req, suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errInvalidClientAssertion, authErr)
	assert.Equal(suite.T(), constants.ErrorInvalidClient, authErr.ErrorCode)
	assert.Equal(suite.T(), "Invalid client assertion", authErr.ErrorDescription)
}

func (suite *ClientAuthTestSuite) TestAuthenticate_PrivateKeyJWT_InvalidJSONPayload() {
	// Build a JWT with valid base64 but invalid JSON in the payload
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`not-json`))
	signature := base64.RawURLEncoding.EncodeToString([]byte("fake-signature"))
	invalidJWT := header + "." + payload + "." + signature

	formData := url.Values{}
	formData.Set("client_assertion_type", constants.SupportedClientAssertionType)
	formData.Set("client_assertion", invalidJWT)

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()

	clientInfo, authErr := authenticate(req, suite.mockAppService, suite.mockJwtService, suite.mockDiscoveryService)

	assert.NotNil(suite.T(), authErr)
	assert.Nil(suite.T(), clientInfo)
	assert.Equal(suite.T(), errInvalidClientAssertion, authErr)
	assert.Equal(suite.T(), constants.ErrorInvalidClient, authErr.ErrorCode)
	assert.Equal(suite.T(), "Invalid client assertion", authErr.ErrorDescription)
}

// validateClientAssertion tests

func (suite *ClientAuthTestSuite) TestValidateClientAssertion_NilCertificate() {
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID:    "test-client",
		Certificate: nil,
	}

	err := validateClientAssertion(oauthApp, suite.mockJwtService, suite.mockDiscoveryService, "test-client", "some.jwt.token")
	assert.NotNil(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "No certificate configured")
}

func (suite *ClientAuthTestSuite) TestValidateClientAssertion_CertificateTypeNone() {
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Certificate: &appmodel.ApplicationCertificate{
			Type:  cert.CertificateTypeNone,
			Value: "",
		},
	}

	err := validateClientAssertion(oauthApp, suite.mockJwtService, suite.mockDiscoveryService, "test-client", "some.jwt.token")
	assert.NotNil(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "No certificate configured")
}

func (suite *ClientAuthTestSuite) TestValidateClientAssertion_JWKSURI_Success() {
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Certificate: &appmodel.ApplicationCertificate{
			Type:  cert.CertificateTypeJWKSURI,
			Value: "https://example.com/.well-known/jwks.json",
		},
	}

	assertion := buildFakeJWTWithSub("test-client")

	suite.mockDiscoveryService.EXPECT().
		GetOAuth2AuthorizationServerMetadata().
		Return(&discovery.OAuth2AuthorizationServerMetadata{
			TokenEndpoint: "https://localhost:9443/oauth2/token",
		})
	suite.mockJwtService.EXPECT().
		VerifyJWTWithJWKS(assertion, "https://example.com/.well-known/jwks.json", "https://localhost:9443/oauth2/token", "test-client").
		Return(nil)

	err := validateClientAssertion(oauthApp, suite.mockJwtService, suite.mockDiscoveryService, "test-client", assertion)
	assert.Nil(suite.T(), err)
}

func (suite *ClientAuthTestSuite) TestValidateClientAssertion_JWKSURI_VerificationFails() {
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Certificate: &appmodel.ApplicationCertificate{
			Type:  cert.CertificateTypeJWKSURI,
			Value: "https://example.com/.well-known/jwks.json",
		},
	}

	assertion := buildFakeJWTWithSub("test-client")

	suite.mockDiscoveryService.EXPECT().
		GetOAuth2AuthorizationServerMetadata().
		Return(&discovery.OAuth2AuthorizationServerMetadata{
			TokenEndpoint: "https://localhost:9443/oauth2/token",
		})
	suite.mockJwtService.EXPECT().
		VerifyJWTWithJWKS(assertion, "https://example.com/.well-known/jwks.json", "https://localhost:9443/oauth2/token", "test-client").
		Return(&serviceerror.ServiceError{Error: "verification failed"})

	err := validateClientAssertion(oauthApp, suite.mockJwtService, suite.mockDiscoveryService, "test-client", assertion)
	assert.NotNil(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "Client assertion verification with JWKS URI failed")
}

func (suite *ClientAuthTestSuite) TestValidateClientAssertion_InvalidJWKSJSON() {
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Certificate: &appmodel.ApplicationCertificate{
			Type:  "jwks",
			Value: "not-valid-json",
		},
	}

	suite.mockDiscoveryService.EXPECT().
		GetOAuth2AuthorizationServerMetadata().
		Return(&discovery.OAuth2AuthorizationServerMetadata{
			TokenEndpoint: "https://localhost:9443/oauth2/token",
		})

	err := validateClientAssertion(oauthApp, suite.mockJwtService, suite.mockDiscoveryService, "test-client", "some.jwt.token")
	assert.NotNil(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "Invalid jwks certificate format")
}

func (suite *ClientAuthTestSuite) TestValidateClientAssertion_InvalidJWTFormat() {
	jwksJSON := buildTestRSAJWKS("test-kid")
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Certificate: &appmodel.ApplicationCertificate{
			Type:  "jwks",
			Value: jwksJSON,
		},
	}

	suite.mockDiscoveryService.EXPECT().
		GetOAuth2AuthorizationServerMetadata().
		Return(&discovery.OAuth2AuthorizationServerMetadata{
			TokenEndpoint: "https://localhost:9443/oauth2/token",
		})

	err := validateClientAssertion(oauthApp, suite.mockJwtService, suite.mockDiscoveryService, "test-client", "invalid-jwt")
	assert.NotNil(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "Failed to decode header")
}

func (suite *ClientAuthTestSuite) TestValidateClientAssertion_MissingKidInHeader() {
	jwksJSON := buildTestRSAJWKS("test-kid")
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Certificate: &appmodel.ApplicationCertificate{
			Type:  "jwks",
			Value: jwksJSON,
		},
	}

	fakeJWT := buildTestJWT(map[string]any{"alg": "RS256", "typ": "JWT"}, map[string]any{"sub": "test-client"})

	suite.mockDiscoveryService.EXPECT().
		GetOAuth2AuthorizationServerMetadata().
		Return(&discovery.OAuth2AuthorizationServerMetadata{
			TokenEndpoint: "https://localhost:9443/oauth2/token",
		})

	err := validateClientAssertion(oauthApp, suite.mockJwtService, suite.mockDiscoveryService, "test-client", fakeJWT)
	assert.NotNil(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "JWT header missing 'kid' claim")
}

func (suite *ClientAuthTestSuite) TestValidateClientAssertion_EmptyKidInHeader() {
	jwksJSON := buildTestRSAJWKS("test-kid")
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Certificate: &appmodel.ApplicationCertificate{
			Type:  "jwks",
			Value: jwksJSON,
		},
	}

	fakeJWT := buildTestJWT(map[string]any{"alg": "RS256", "kid": "", "typ": "JWT"}, map[string]any{"sub": "test-client"})

	suite.mockDiscoveryService.EXPECT().
		GetOAuth2AuthorizationServerMetadata().
		Return(&discovery.OAuth2AuthorizationServerMetadata{
			TokenEndpoint: "https://localhost:9443/oauth2/token",
		})

	err := validateClientAssertion(oauthApp, suite.mockJwtService, suite.mockDiscoveryService, "test-client", fakeJWT)
	assert.NotNil(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "JWT header missing 'kid' claim")
}

func (suite *ClientAuthTestSuite) TestValidateClientAssertion_KidNotAString() {
	jwksJSON := buildTestRSAJWKS("test-kid")
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Certificate: &appmodel.ApplicationCertificate{
			Type:  "jwks",
			Value: jwksJSON,
		},
	}

	fakeJWT := buildTestJWT(map[string]any{"alg": "RS256", "kid": 12345, "typ": "JWT"}, map[string]any{"sub": "test-client"})

	suite.mockDiscoveryService.EXPECT().
		GetOAuth2AuthorizationServerMetadata().
		Return(&discovery.OAuth2AuthorizationServerMetadata{
			TokenEndpoint: "https://localhost:9443/oauth2/token",
		})

	err := validateClientAssertion(oauthApp, suite.mockJwtService, suite.mockDiscoveryService, "test-client", fakeJWT)
	assert.NotNil(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "JWT header missing 'kid' claim")
}

func (suite *ClientAuthTestSuite) TestValidateClientAssertion_NoMatchingKidInJWKS() {
	jwksJSON := buildTestRSAJWKS("different-kid")
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Certificate: &appmodel.ApplicationCertificate{
			Type:  "jwks",
			Value: jwksJSON,
		},
	}

	fakeJWT := buildTestJWT(map[string]any{"alg": "RS256", "kid": "test-kid", "typ": "JWT"}, map[string]any{"sub": "test-client"})

	suite.mockDiscoveryService.EXPECT().
		GetOAuth2AuthorizationServerMetadata().
		Return(&discovery.OAuth2AuthorizationServerMetadata{
			TokenEndpoint: "https://localhost:9443/oauth2/token",
		})

	err := validateClientAssertion(oauthApp, suite.mockJwtService, suite.mockDiscoveryService, "test-client", fakeJWT)
	assert.NotNil(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "No matching key found in JWKS")
}

func (suite *ClientAuthTestSuite) TestValidateClientAssertion_InvalidJWKCannotConvertToPublicKey() {
	invalidJWK := map[string]any{
		"kty": "RSA",
		"kid": "test-kid",
	}
	jwks := map[string]any{"keys": []map[string]any{invalidJWK}}
	jwksJSON, _ := json.Marshal(jwks)

	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Certificate: &appmodel.ApplicationCertificate{
			Type:  "jwks",
			Value: string(jwksJSON),
		},
	}

	fakeJWT := buildTestJWT(map[string]any{"alg": "RS256", "kid": "test-kid", "typ": "JWT"}, map[string]any{"sub": "test-client"})

	suite.mockDiscoveryService.EXPECT().
		GetOAuth2AuthorizationServerMetadata().
		Return(&discovery.OAuth2AuthorizationServerMetadata{
			TokenEndpoint: "https://localhost:9443/oauth2/token",
		})

	err := validateClientAssertion(oauthApp, suite.mockJwtService, suite.mockDiscoveryService, "test-client", fakeJWT)
	assert.NotNil(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "Failed to convert JWK to public key")
}

func (suite *ClientAuthTestSuite) TestValidateClientAssertion_VerificationFails() {
	jwksJSON := buildTestRSAJWKS("test-kid")
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Certificate: &appmodel.ApplicationCertificate{
			Type:  "jwks",
			Value: jwksJSON,
		},
	}

	fakeJWT := buildTestJWT(map[string]any{"alg": "RS256", "kid": "test-kid", "typ": "JWT"}, map[string]any{"sub": "test-client"})

	suite.mockDiscoveryService.EXPECT().
		GetOAuth2AuthorizationServerMetadata().
		Return(&discovery.OAuth2AuthorizationServerMetadata{
			TokenEndpoint: "https://localhost:9443/oauth2/token",
		})
	suite.mockJwtService.EXPECT().
		VerifyJWTWithPublicKey(fakeJWT, mock.Anything, "https://localhost:9443/oauth2/token", "test-client").
		Return(&serviceerror.ServiceError{
			Code:  "JWT-00001",
			Type:  serviceerror.ClientErrorType,
			Error: "invalid_token",
		})

	err := validateClientAssertion(oauthApp, suite.mockJwtService, suite.mockDiscoveryService, "test-client", fakeJWT)
	assert.NotNil(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "Client assertion verification failed")
}

func (suite *ClientAuthTestSuite) TestValidateClientAssertion_Success() {
	jwksJSON := buildTestRSAJWKS("test-kid")
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Certificate: &appmodel.ApplicationCertificate{
			Type:  "jwks",
			Value: jwksJSON,
		},
	}

	fakeJWT := buildTestJWT(map[string]any{"alg": "RS256", "kid": "test-kid", "typ": "JWT"}, map[string]any{"sub": "test-client"})

	suite.mockDiscoveryService.EXPECT().
		GetOAuth2AuthorizationServerMetadata().
		Return(&discovery.OAuth2AuthorizationServerMetadata{
			TokenEndpoint: "https://localhost:9443/oauth2/token",
		})
	suite.mockJwtService.EXPECT().
		VerifyJWTWithPublicKey(fakeJWT, mock.Anything, "https://localhost:9443/oauth2/token", "test-client").
		Return(nil)

	err := validateClientAssertion(oauthApp, suite.mockJwtService, suite.mockDiscoveryService, "test-client", fakeJWT)
	assert.Nil(suite.T(), err)
}

func (suite *ClientAuthTestSuite) TestValidateClientAssertion_EmptyJWKSKeys() {
	jwks := map[string]any{"keys": []map[string]any{}}
	jwksJSON, _ := json.Marshal(jwks)

	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Certificate: &appmodel.ApplicationCertificate{
			Type:  "jwks",
			Value: string(jwksJSON),
		},
	}

	fakeJWT := buildTestJWT(map[string]any{"alg": "RS256", "kid": "test-kid", "typ": "JWT"}, map[string]any{"sub": "test-client"})

	suite.mockDiscoveryService.EXPECT().
		GetOAuth2AuthorizationServerMetadata().
		Return(&discovery.OAuth2AuthorizationServerMetadata{
			TokenEndpoint: "https://localhost:9443/oauth2/token",
		})

	err := validateClientAssertion(oauthApp, suite.mockJwtService, suite.mockDiscoveryService, "test-client", fakeJWT)
	assert.NotNil(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "No matching key found in JWKS")
}

func (suite *ClientAuthTestSuite) TestValidateClientAssertion_MultipleKeysMatchesCorrectKid() {
	jwksJSON1 := buildTestRSAJWKS("kid-1")
	jwksJSON2 := buildTestRSAJWKS("kid-2")

	var jwks1, jwks2 struct {
		Keys []map[string]any `json:"keys"`
	}
	_ = json.Unmarshal([]byte(jwksJSON1), &jwks1)
	_ = json.Unmarshal([]byte(jwksJSON2), &jwks2)

	combinedJWKS := map[string]any{
		"keys": []map[string]any{jwks1.Keys[0], jwks2.Keys[0]},
	}
	combinedJSON, _ := json.Marshal(combinedJWKS)

	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Certificate: &appmodel.ApplicationCertificate{
			Type:  "jwks",
			Value: string(combinedJSON),
		},
	}

	fakeJWT := buildTestJWT(map[string]any{"alg": "RS256", "kid": "kid-2", "typ": "JWT"}, map[string]any{"sub": "test-client"})

	suite.mockDiscoveryService.EXPECT().
		GetOAuth2AuthorizationServerMetadata().
		Return(&discovery.OAuth2AuthorizationServerMetadata{
			TokenEndpoint: "https://localhost:9443/oauth2/token",
		})
	suite.mockJwtService.EXPECT().
		VerifyJWTWithPublicKey(fakeJWT, mock.Anything, "https://localhost:9443/oauth2/token", "test-client").
		Return(nil)

	err := validateClientAssertion(oauthApp, suite.mockJwtService, suite.mockDiscoveryService, "test-client", fakeJWT)
	assert.Nil(suite.T(), err)
}
