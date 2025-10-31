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

package dcr

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/asgardeo/thunder/tests/integration/testutils"
	"github.com/stretchr/testify/suite"
)

const (
	testServerURL = "https://localhost:8095"
	dcrEndpoint   = "/oauth2/dcr/register"
)

type DCRTestSuite struct {
	suite.Suite
	registeredAppIDs []string
}

func TestDCRTestSuite(t *testing.T) {
	suite.Run(t, new(DCRTestSuite))
}

func (ts *DCRTestSuite) TearDownSuite() {
	for _, appID := range ts.registeredAppIDs {
		if appID != "" {
			err := testutils.DeleteApplication(appID)
			if err != nil {
				ts.T().Logf("Failed to delete application during teardown: %v", err)
			}
		}
	}
}

// TestDCRRegistrationWithAllFields verifies successful registration with all RFC 7591 metadata fields populated.
func (ts *DCRTestSuite) TestDCRRegistrationWithAllFields() {
	request := DCRRegistrationRequest{
		RedirectURIs:            []string{"https://client.example.com/callback", "https://client.example.com/callback2"},
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		ResponseTypes:           []string{"code"},
		ClientName:              "Test Client Full",
		ClientURI:               "https://client.example.com",
		LogoURI:                 "https://client.example.com/logo.png",
		TokenEndpointAuthMethod: "client_secret_basic",
		Scope:                   "openid profile email",
		Contacts:                []string{"admin@example.com", "support@example.com"},
		TosURI:                  "https://client.example.com/tos",
		PolicyURI:               "https://client.example.com/policy",
	}

	response, statusCode := ts.registerClient(request)

	ts.Assert().Equal(http.StatusCreated, statusCode)
	ts.Assert().NotEmpty(response.ClientID)
	ts.Assert().NotEmpty(response.ClientSecret)
	ts.Assert().Equal(int64(0), response.ClientSecretExpiresAt)
	ts.Assert().NotEmpty(response.AppID)
	ts.Assert().Equal(request.RedirectURIs, response.RedirectURIs)
	ts.Assert().Equal(request.GrantTypes, response.GrantTypes)
	ts.Assert().Equal(request.ResponseTypes, response.ResponseTypes)
	ts.Assert().Equal(request.ClientName, response.ClientName)
	ts.Assert().Equal(request.ClientURI, response.ClientURI)
	ts.Assert().Equal(request.LogoURI, response.LogoURI)
	ts.Assert().Equal(request.TokenEndpointAuthMethod, response.TokenEndpointAuthMethod)
	ts.Assert().Equal(request.Scope, response.Scope)
	ts.Assert().Equal(request.Contacts, response.Contacts)
	ts.Assert().Equal(request.TosURI, response.TosURI)
	ts.Assert().Equal(request.PolicyURI, response.PolicyURI)

	ts.registeredAppIDs = append(ts.registeredAppIDs, response.AppID)
}

// TestDCRRegistrationMinimalFields verifies registration with only redirect URIs and auto-generated client_name.
func (ts *DCRTestSuite) TestDCRRegistrationMinimalFields() {
	request := DCRRegistrationRequest{
		RedirectURIs: []string{"https://minimal.example.com/callback"},
	}

	response, statusCode := ts.registerClient(request)

	ts.Assert().Equal(http.StatusCreated, statusCode)
	ts.Assert().NotEmpty(response.ClientID)
	ts.Assert().NotEmpty(response.ClientSecret)
	ts.Assert().Equal(int64(0), response.ClientSecretExpiresAt)
	ts.Assert().Equal(request.RedirectURIs, response.RedirectURIs)
	ts.Assert().Equal([]string{"authorization_code"}, response.GrantTypes)
	ts.Assert().Equal([]string{"code"}, response.ResponseTypes)
	ts.Assert().Equal("client_secret_basic", response.TokenEndpointAuthMethod)
	ts.Assert().NotEmpty(response.ClientName)

	ts.registeredAppIDs = append(ts.registeredAppIDs, response.AppID)
}

// TestDCRRegistrationPublicClient verifies public client registration with token_endpoint_auth_method=none.
func (ts *DCRTestSuite) TestDCRRegistrationPublicClient() {
	request := DCRRegistrationRequest{
		RedirectURIs:            []string{"https://public.example.com/callback"},
		ClientName:              "Public Client",
		TokenEndpointAuthMethod: "none",
		GrantTypes:              []string{"authorization_code"},
		ResponseTypes:           []string{"code"},
	}

	response, statusCode := ts.registerClient(request)

	ts.Assert().Equal(http.StatusCreated, statusCode)
	ts.Assert().NotEmpty(response.ClientID)
	ts.Assert().Empty(response.ClientSecret)
	ts.Assert().Equal("none", response.TokenEndpointAuthMethod)

	ts.registeredAppIDs = append(ts.registeredAppIDs, response.AppID)
}

// TestDCRRegistrationWithClientCredentialsGrant verifies M2M client registration without redirect URIs.
func (ts *DCRTestSuite) TestDCRRegistrationWithClientCredentialsGrant() {
	request := DCRRegistrationRequest{
		GrantTypes:              []string{"client_credentials"},
		ClientName:              "Client Credentials App",
		TokenEndpointAuthMethod: "client_secret_post",
	}

	response, statusCode := ts.registerClient(request)

	ts.Assert().Equal(http.StatusCreated, statusCode)
	ts.Assert().NotEmpty(response.ClientID)
	ts.Assert().NotEmpty(response.ClientSecret)
	ts.Assert().Equal([]string{"client_credentials"}, response.GrantTypes)
	ts.Assert().Equal("client_secret_post", response.TokenEndpointAuthMethod)
	ts.Assert().Empty(response.ResponseTypes)
	ts.Assert().Empty(response.RedirectURIs)

	ts.registeredAppIDs = append(ts.registeredAppIDs, response.AppID)
}

// TestDCRRegistrationWithMultipleGrantTypes verifies registration with multiple OAuth grant types.
func (ts *DCRTestSuite) TestDCRRegistrationWithMultipleGrantTypes() {
	request := DCRRegistrationRequest{
		RedirectURIs:  []string{"https://multi.example.com/callback"},
		GrantTypes:    []string{"authorization_code", "refresh_token", "client_credentials"},
		ResponseTypes: []string{"code"},
		ClientName:    "Multi Grant Client",
	}

	response, statusCode := ts.registerClient(request)

	ts.Assert().Equal(http.StatusCreated, statusCode)
	ts.Assert().NotEmpty(response.ClientID)
	ts.Assert().Equal(request.GrantTypes, response.GrantTypes)

	ts.registeredAppIDs = append(ts.registeredAppIDs, response.AppID)
}

// TestDCRRegistrationWithScopes verifies registration with custom OAuth scopes.
func (ts *DCRTestSuite) TestDCRRegistrationWithScopes() {
	request := DCRRegistrationRequest{
		RedirectURIs: []string{"https://scopes.example.com/callback"},
		ClientName:   "Scoped Client",
		Scope:        "openid profile email address phone",
	}

	response, statusCode := ts.registerClient(request)

	ts.Assert().Equal(http.StatusCreated, statusCode)
	ts.Assert().NotEmpty(response.ClientID)
	ts.Assert().Equal(request.Scope, response.Scope)

	ts.registeredAppIDs = append(ts.registeredAppIDs, response.AppID)
}

// TestDCRRegistrationWithMultipleContacts verifies registration with multiple contact email addresses.
func (ts *DCRTestSuite) TestDCRRegistrationWithMultipleContacts() {
	request := DCRRegistrationRequest{
		RedirectURIs: []string{"https://contacts.example.com/callback"},
		ClientName:   "Multi Contact Client",
		Contacts:     []string{"admin@example.com", "support@example.com", "security@example.com"},
	}

	response, statusCode := ts.registerClient(request)

	ts.Assert().Equal(http.StatusCreated, statusCode)
	ts.Assert().NotEmpty(response.ClientID)
	ts.Assert().Equal(request.Contacts, response.Contacts)

	ts.registeredAppIDs = append(ts.registeredAppIDs, response.AppID)
}

// TestDCRRegistrationEmptyRedirectURIs verifies rejection when redirect URIs are required but empty.
func (ts *DCRTestSuite) TestDCRRegistrationEmptyRedirectURIs() {
	request := DCRRegistrationRequest{
		RedirectURIs: []string{},
		ClientName:   "No Redirect URI Client",
	}

	_, statusCode, errResp := ts.registerClientWithError(request)

	ts.Assert().Equal(http.StatusBadRequest, statusCode)
	ts.Assert().NotEmpty(errResp.Error)
}

// TestDCRRegistrationInvalidRedirectURI verifies rejection of malformed redirect URI values.
func (ts *DCRTestSuite) TestDCRRegistrationInvalidRedirectURI() {
	request := DCRRegistrationRequest{
		RedirectURIs: []string{"not-a-valid-uri"},
		ClientName:   "Invalid URI Client",
	}

	_, statusCode, errResp := ts.registerClientWithError(request)

	ts.Assert().Equal(http.StatusBadRequest, statusCode)
	ts.Assert().NotEmpty(errResp.Error)
}

// TestDCRRegistrationFragmentInRedirectURI verifies rejection of redirect URIs with fragments per RFC 6749.
func (ts *DCRTestSuite) TestDCRRegistrationFragmentInRedirectURI() {
	request := DCRRegistrationRequest{
		RedirectURIs: []string{"https://example.com/callback#fragment"},
		ClientName:   "Fragment URI Client",
	}

	_, statusCode, errResp := ts.registerClientWithError(request)

	ts.Assert().Equal(http.StatusBadRequest, statusCode)
	ts.Assert().NotEmpty(errResp.Error)
}

// TestDCRRegistrationInvalidJSON verifies rejection of malformed JSON request body.
func (ts *DCRTestSuite) TestDCRRegistrationInvalidJSON() {
	invalidJSON := []byte(`{"redirect_uris": ["https://example.com"], "invalid_json"}`)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("POST", testServerURL+dcrEndpoint, bytes.NewReader(invalidJSON))
	if err != nil {
		ts.T().Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		ts.T().Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	ts.Assert().Equal(http.StatusBadRequest, resp.StatusCode)
}

// TestDCRRegistrationInvalidTokenEndpointAuthMethod verifies rejection of unsupported auth methods.
func (ts *DCRTestSuite) TestDCRRegistrationInvalidTokenEndpointAuthMethod() {
	request := DCRRegistrationRequest{
		RedirectURIs:            []string{"https://invalid-auth.example.com/callback"},
		ClientName:              "Invalid Auth Method Client",
		TokenEndpointAuthMethod: "invalid_method",
	}

	_, statusCode, errResp := ts.registerClientWithError(request)

	ts.Assert().Equal(http.StatusBadRequest, statusCode)
	ts.Assert().NotEmpty(errResp.Error)
}

// TestDCRRegistrationWithPartialDefaults verifies correct default value application for omitted fields.
func (ts *DCRTestSuite) TestDCRRegistrationWithPartialDefaults() {
	request := DCRRegistrationRequest{
		RedirectURIs:            []string{"https://partial.example.com/callback"},
		ClientName:              "Partial Defaults Client",
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		TokenEndpointAuthMethod: "client_secret_post",
	}

	response, statusCode := ts.registerClient(request)

	ts.Assert().Equal(http.StatusCreated, statusCode)
	ts.Assert().NotEmpty(response.ClientID)
	ts.Assert().Equal(request.GrantTypes, response.GrantTypes)
	ts.Assert().Equal("client_secret_post", response.TokenEndpointAuthMethod)
	ts.Assert().Equal([]string{"code"}, response.ResponseTypes)

	ts.registeredAppIDs = append(ts.registeredAppIDs, response.AppID)
}

func (ts *DCRTestSuite) registerClient(request DCRRegistrationRequest) (*DCRRegistrationResponse, int) {
	requestJSON, err := json.Marshal(request)
	if err != nil {
		ts.T().Fatalf("Failed to marshal request: %v", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("POST", testServerURL+dcrEndpoint, bytes.NewReader(requestJSON))
	if err != nil {
		ts.T().Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		ts.T().Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		responseBody, _ := io.ReadAll(resp.Body)
		ts.T().Fatalf("Expected status 201, got %d. Response: %s", resp.StatusCode, string(responseBody))
	}

	var response DCRRegistrationResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		ts.T().Fatalf("Failed to decode response: %v", err)
	}

	return &response, resp.StatusCode
}

// TestDCRRegistrationInvalidGrantType verifies rejection of unknown OAuth grant type values.
func (ts *DCRTestSuite) TestDCRRegistrationInvalidGrantType() {
	request := DCRRegistrationRequest{
		RedirectURIs: []string{"https://example.com/callback"},
		ClientName:   "Invalid Grant Type Client",
		GrantTypes:   []string{"invalid_grant_type"},
	}

	_, statusCode, errResp := ts.registerClientWithError(request)

	ts.Assert().Equal(http.StatusBadRequest, statusCode)
	ts.Assert().NotEmpty(errResp.Error)
}

// TestDCRRegistrationInvalidResponseType verifies rejection of unknown OAuth response type values.
func (ts *DCRTestSuite) TestDCRRegistrationInvalidResponseType() {
	request := DCRRegistrationRequest{
		RedirectURIs:  []string{"https://example.com/callback"},
		ClientName:    "Invalid Response Type Client",
		ResponseTypes: []string{"invalid_response"},
	}

	_, statusCode, errResp := ts.registerClientWithError(request)

	ts.Assert().Equal(http.StatusBadRequest, statusCode)
	ts.Assert().NotEmpty(errResp.Error)
}

// TestDCRRegistrationJWKSAndJWKSUriConflict verifies rejection when both JWKS and JWKS URI are specified.
func (ts *DCRTestSuite) TestDCRRegistrationJWKSAndJWKSUriConflict() {
	request := DCRRegistrationRequest{
		RedirectURIs:            []string{"https://example.com/callback"},
		ClientName:              "JWKS Conflict Client",
		TokenEndpointAuthMethod: "private_key_jwt",
		JWKSUri:                 "https://example.com/jwks",
		JWKS: map[string]interface{}{
			"keys": []interface{}{
				map[string]interface{}{
					"kty": "RSA",
					"use": "sig",
					"kid": "test-key",
				},
			},
		},
	}

	_, statusCode, errResp := ts.registerClientWithError(request)

	ts.Assert().Equal(http.StatusBadRequest, statusCode)
	ts.Assert().NotEmpty(errResp.Error)
}

// TestDCRRegistrationJWKSUriNotHTTPS verifies rejection of non-HTTPS JWKS URI per RFC 7591.
func (ts *DCRTestSuite) TestDCRRegistrationJWKSUriNotHTTPS() {
	request := DCRRegistrationRequest{
		RedirectURIs:            []string{"https://example.com/callback"},
		ClientName:              "Non-HTTPS JWKS URI Client",
		TokenEndpointAuthMethod: "private_key_jwt",
		JWKSUri:                 "http://example.com/jwks",
	}

	_, statusCode, errResp := ts.registerClientWithError(request)

	ts.Assert().Equal(http.StatusBadRequest, statusCode)
	ts.Assert().NotEmpty(errResp.Error)
}

// TestDCRRegistrationMultipleRedirectURIs verifies registration with multiple redirect URI values.
func (ts *DCRTestSuite) TestDCRRegistrationMultipleRedirectURIs() {
	request := DCRRegistrationRequest{
		RedirectURIs: []string{
			"https://example.com/callback1",
			"https://example.com/callback2",
			"https://example.com/callback3",
		},
		ClientName: "Multiple Redirect URIs Client",
	}

	response, statusCode := ts.registerClient(request)

	ts.Assert().Equal(http.StatusCreated, statusCode)
	ts.Assert().NotEmpty(response.ClientID)
	ts.Assert().Equal(3, len(response.RedirectURIs))
	ts.Assert().Equal(request.RedirectURIs, response.RedirectURIs)

	ts.registeredAppIDs = append(ts.registeredAppIDs, response.AppID)
}

// TestDCRRegistrationRefreshTokenGrant verifies registration with refresh_token grant type.
func (ts *DCRTestSuite) TestDCRRegistrationRefreshTokenGrant() {
	request := DCRRegistrationRequest{
		RedirectURIs:  []string{"https://example.com/callback"},
		ClientName:    "Refresh Token Client",
		GrantTypes:    []string{"authorization_code", "refresh_token"},
		ResponseTypes: []string{"code"},
	}

	response, statusCode := ts.registerClient(request)

	ts.Assert().Equal(http.StatusCreated, statusCode)
	ts.Assert().NotEmpty(response.ClientID)
	ts.Assert().Equal(request.GrantTypes, response.GrantTypes)

	ts.registeredAppIDs = append(ts.registeredAppIDs, response.AppID)
}

// TestDCRRegistrationInvalidClientURI verifies rejection of malformed client_uri values.
func (ts *DCRTestSuite) TestDCRRegistrationInvalidClientURI() {
	request := DCRRegistrationRequest{
		RedirectURIs: []string{"https://example.com/callback"},
		ClientName:   "Invalid Client URI Client",
		ClientURI:    "not-a-valid-uri",
	}

	_, statusCode, errResp := ts.registerClientWithError(request)

	ts.Assert().Equal(http.StatusBadRequest, statusCode)
	ts.Assert().NotEmpty(errResp.Error)
}

// TestDCRRegistrationInvalidLogoURI verifies rejection of malformed logo_uri values.
func (ts *DCRTestSuite) TestDCRRegistrationInvalidLogoURI() {
	request := DCRRegistrationRequest{
		RedirectURIs: []string{"https://example.com/callback"},
		ClientName:   "Invalid Logo URI Client",
		LogoURI:      "not-a-valid-uri",
	}

	_, statusCode, errResp := ts.registerClientWithError(request)

	ts.Assert().Equal(http.StatusBadRequest, statusCode)
	ts.Assert().NotEmpty(errResp.Error)
}

// TestDCRRegistrationEmptyGrantTypesArray verifies default grant type application when array is empty.
func (ts *DCRTestSuite) TestDCRRegistrationEmptyGrantTypesArray() {
	request := DCRRegistrationRequest{
		RedirectURIs: []string{"https://example.com/callback"},
		ClientName:   "Empty Grant Types Client",
		GrantTypes:   []string{},
	}

	response, statusCode := ts.registerClient(request)

	ts.Assert().Equal(http.StatusCreated, statusCode)
	ts.Assert().NotEmpty(response.ClientID)
	ts.Assert().Equal([]string{"authorization_code"}, response.GrantTypes)
	ts.Assert().Equal([]string{"code"}, response.ResponseTypes)

	ts.registeredAppIDs = append(ts.registeredAppIDs, response.AppID)
}

func (ts *DCRTestSuite) registerClientWithError(request DCRRegistrationRequest) (*DCRRegistrationResponse, int, *DCRErrorResponse) {
	requestJSON, err := json.Marshal(request)
	if err != nil {
		ts.T().Fatalf("Failed to marshal request: %v", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("POST", testServerURL+dcrEndpoint, bytes.NewReader(requestJSON))
	if err != nil {
		ts.T().Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		ts.T().Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	responseBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusCreated {
		var successResp DCRRegistrationResponse
		json.Unmarshal(responseBody, &successResp)
		return &successResp, resp.StatusCode, nil
	}

	var errResp DCRErrorResponse
	json.Unmarshal(responseBody, &errResp)
	return nil, resp.StatusCode, &errResp
}

// TestDCRRegistrationWithJWKSURI tests OAuth client registration with JWKS_URI certificate.
func (ts *DCRTestSuite) TestDCRRegistrationWithJWKSURI() {
	request := DCRRegistrationRequest{
		RedirectURIs:            []string{"https://client.example.com/callback"},
		GrantTypes:              []string{"authorization_code"},
		ResponseTypes:           []string{"code"},
		ClientName:              "JWKS URI Test Client",
		TokenEndpointAuthMethod: "client_secret_basic",
		JWKSUri:                 "https://client.example.com/.well-known/jwks.json",
	}

	response, statusCode, _ := ts.registerClientWithError(request)

	ts.Assert().Equal(http.StatusCreated, statusCode)
	ts.Assert().NotEmpty(response.ClientID)
	ts.Assert().Equal(request.JWKSUri, response.JWKSUri)
	ts.Assert().Nil(response.JWKS)

	ts.registeredAppIDs = append(ts.registeredAppIDs, response.AppID)
}

// TestDCRRegistrationWithJWKS tests OAuth client registration with inline JWKS certificate.
func (ts *DCRTestSuite) TestDCRRegistrationWithJWKS() {
	jwks := map[string]interface{}{
		"keys": []map[string]interface{}{
			{
				"kty": "RSA",
				"use": "sig",
				"kid": "test-key-1",
				"n":   "0vx7agoebGcQSuuPiLJXZptN9nndrQmbXEps2aiAFbWhM78LhWx4cbbfAAtVT86zwu1RK7aPFFxuhDR1L6tSoc_BJECPebWKRXjBZCiFV4n3oknjhMstn64tZ_2W-5JsGY4Hc5n9yBXArwl93lqt7_RN5w6Cf0h4QyQ5v-65YGjQR0_FDW2QvzqY368QQMicAtaSqzs8KJZgnYb9c7d0zgdAZHzu6qMQvRL5hajrn1n91CbOpbISD08qNLyrdkt-bFTWhAI4vMQFh6WeZu0fM4lFd2NcRwr3XPksINHaQ-G_xBniIqbw0Ls1jF44-csFCur-kEgU8awapJzKnqDKgw",
				"e":   "AQAB",
			},
		},
	}

	request := DCRRegistrationRequest{
		RedirectURIs:            []string{"https://client.example.com/callback"},
		GrantTypes:              []string{"authorization_code"},
		ResponseTypes:           []string{"code"},
		ClientName:              "JWKS Inline Test Client",
		TokenEndpointAuthMethod: "client_secret_basic",
		JWKS:                    jwks,
	}

	response, statusCode, _ := ts.registerClientWithError(request)

	ts.Assert().Equal(http.StatusCreated, statusCode)
	ts.Assert().NotEmpty(response.ClientID)
	ts.Assert().Empty(response.JWKSUri)
	ts.Assert().NotNil(response.JWKS)
	ts.Assert().Contains(response.JWKS, "keys")
	keys, ok := response.JWKS["keys"].([]interface{})
	ts.Assert().True(ok, "keys should be an array")
	ts.Assert().Len(keys, 1, "should have one key")
	key, ok := keys[0].(map[string]interface{})
	ts.Assert().True(ok, "key should be a map")
	ts.Assert().Equal("RSA", key["kty"])
	ts.Assert().Equal("sig", key["use"])
	ts.Assert().Equal("test-key-1", key["kid"])

	ts.registeredAppIDs = append(ts.registeredAppIDs, response.AppID)
}

// TestDCRRegistrationWithJWKSURIAndRetrieve tests OAuth client registration with JWKS_URI and retrieves it to verify persistence.
func (ts *DCRTestSuite) TestDCRRegistrationWithJWKSURIAndRetrieve() {
	request := DCRRegistrationRequest{
		RedirectURIs:            []string{"https://retrieve-test.example.com/callback"},
		GrantTypes:              []string{"authorization_code"},
		ResponseTypes:           []string{"code"},
		ClientName:              "JWKS URI Retrieve Test Client",
		TokenEndpointAuthMethod: "client_secret_basic",
		JWKSUri:                 "https://retrieve-test.example.com/.well-known/jwks.json",
		Scope:                   "openid profile email",
	}

	// Register the client
	response, statusCode, _ := ts.registerClientWithError(request)
	ts.Assert().Equal(http.StatusCreated, statusCode)
	ts.Assert().NotEmpty(response.ClientID)
	ts.Assert().NotEmpty(response.AppID)
	ts.Assert().Equal(request.JWKSUri, response.JWKSUri)

	// Retrieve the application via application API to verify the JWKS URI is persisted
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	getReq, err := http.NewRequest("GET", testServerURL+"/applications/"+response.AppID, nil)
	ts.Require().NoError(err, "Failed to create GET request")

	getResp, err := client.Do(getReq)
	ts.Require().NoError(err, "Failed to send GET request")
	defer getResp.Body.Close()

	ts.Assert().Equal(http.StatusOK, getResp.StatusCode, "Expected status 200 when retrieving application")

	var retrievedApp map[string]interface{}
	err = json.NewDecoder(getResp.Body).Decode(&retrievedApp)
	ts.Require().NoError(err, "Failed to decode application response")

	// Verify the application-level certificate contains the JWKS URI
	certificate, ok := retrievedApp["certificate"].(map[string]interface{})
	ts.Assert().True(ok, "certificate should be present at application level")
	ts.Assert().NotNil(certificate, "certificate should not be nil")
	ts.Assert().Equal("JWKS_URI", certificate["type"], "certificate type should be JWKS_URI")
	ts.Assert().Equal(request.JWKSUri, certificate["value"], "certificate value should match the JWKS URI")

	ts.registeredAppIDs = append(ts.registeredAppIDs, response.AppID)
}

// TestDCRRegistrationJWKSAndJWKSURIConflictNew tests that providing both JWKS and JWKS_URI returns error.
func (ts *DCRTestSuite) TestDCRRegistrationJWKSAndJWKSURIConflictNew() {
	jwks := map[string]interface{}{
		"keys": []map[string]interface{}{
			{
				"kty": "RSA",
				"kid": "test-key",
			},
		},
	}

	request := DCRRegistrationRequest{
		RedirectURIs:            []string{"https://client.example.com/callback"},
		GrantTypes:              []string{"authorization_code"},
		ResponseTypes:           []string{"code"},
		ClientName:              "Conflict Test Client",
		TokenEndpointAuthMethod: "client_secret_basic",
		JWKSUri:                 "https://client.example.com/.well-known/jwks.json",
		JWKS:                    jwks,
	}

	_, statusCode, errResp := ts.registerClientWithError(request)

	ts.Assert().Equal(http.StatusBadRequest, statusCode)
	ts.Assert().NotNil(errResp)
	ts.Assert().Equal("invalid_client_metadata", errResp.Error)
}

// TestDCRRegistrationClientNameFallback tests that client_id is used as client_name when omitted.
func (ts *DCRTestSuite) TestDCRRegistrationClientNameFallback() {
	request := DCRRegistrationRequest{
		RedirectURIs: []string{"https://client.example.com/callback"},
		GrantTypes:   []string{"authorization_code"},
	}

	response, statusCode, _ := ts.registerClientWithError(request)

	ts.Assert().Equal(http.StatusCreated, statusCode)
	ts.Assert().NotEmpty(response.ClientID)
	ts.Assert().Equal(response.ClientID, response.ClientName)

	ts.registeredAppIDs = append(ts.registeredAppIDs, response.AppID)
}

// TestDCRRegistrationScopeConversion tests space-separated scope string to array conversion.
func (ts *DCRTestSuite) TestDCRRegistrationScopeConversion() {
	scopeString := "openid profile email address phone"

	request := DCRRegistrationRequest{
		RedirectURIs: []string{"https://client.example.com/callback"},
		GrantTypes:   []string{"authorization_code"},
		ClientName:   "Scope Test Client",
		Scope:        scopeString,
	}

	response, statusCode, _ := ts.registerClientWithError(request)

	ts.Assert().Equal(http.StatusCreated, statusCode)
	ts.Assert().Equal(scopeString, response.Scope)

	ts.registeredAppIDs = append(ts.registeredAppIDs, response.AppID)
}

// TestDCRRegistrationInvalidJWKSURI2 tests DCR with invalid JWKS_URI.
func (ts *DCRTestSuite) TestDCRRegistrationInvalidJWKSURI2() {
	request := DCRRegistrationRequest{
		RedirectURIs:            []string{"https://client.example.com/callback"},
		GrantTypes:              []string{"authorization_code"},
		ResponseTypes:           []string{"code"},
		ClientName:              "Invalid JWKS URI Test",
		TokenEndpointAuthMethod: "client_secret_basic",
		JWKSUri:                 "not-a-valid-uri",
	}

	_, statusCode, _ := ts.registerClientWithError(request)
	ts.Assert().Equal(http.StatusBadRequest, statusCode)
}

// TestDCRRegistrationRedirectURIWithFragment2 tests DCR rejects redirect URIs with fragments.
func (ts *DCRTestSuite) TestDCRRegistrationRedirectURIWithFragment2() {
	request := DCRRegistrationRequest{
		RedirectURIs:            []string{"https://client.example.com/callback#fragment"},
		GrantTypes:              []string{"authorization_code"},
		ResponseTypes:           []string{"code"},
		ClientName:              "URI Fragment Test",
		TokenEndpointAuthMethod: "client_secret_basic",
	}

	_, statusCode, _ := ts.registerClientWithError(request)
	ts.Assert().Equal(http.StatusBadRequest, statusCode)
}

// TestDCRRegistrationClientCredentialsWithResponseTypes2 tests client_credentials cannot have response_types.
func (ts *DCRTestSuite) TestDCRRegistrationClientCredentialsWithResponseTypes2() {
	request := DCRRegistrationRequest{
		RedirectURIs:            []string{"https://client.example.com/callback"},
		GrantTypes:              []string{"client_credentials"},
		ResponseTypes:           []string{"code"},
		ClientName:              "Client Credentials Response Type Test",
		TokenEndpointAuthMethod: "client_secret_basic",
	}

	_, statusCode, _ := ts.registerClientWithError(request)
	ts.Assert().Equal(http.StatusBadRequest, statusCode)
}

// TestDCRRegistrationEmptyScope tests DCR with empty scope string.
func (ts *DCRTestSuite) TestDCRRegistrationEmptyScope() {
	request := DCRRegistrationRequest{
		RedirectURIs:            []string{"https://client.example.com/callback"},
		GrantTypes:              []string{"authorization_code"},
		ResponseTypes:           []string{"code"},
		ClientName:              "Empty Scope Test",
		TokenEndpointAuthMethod: "client_secret_basic",
		Scope:                   "",
	}

	response, statusCode, _ := ts.registerClientWithError(request)
	ts.Assert().Equal(http.StatusCreated, statusCode)
	ts.Assert().Equal("", response.Scope)

	if response.AppID != "" {
		ts.registeredAppIDs = append(ts.registeredAppIDs, response.AppID)
	}
}

// TestDCRRegistrationMultipleScopesConversion tests scope array conversion.
func (ts *DCRTestSuite) TestDCRRegistrationMultipleScopesConversion() {
	request := DCRRegistrationRequest{
		RedirectURIs:            []string{"https://client.example.com/callback"},
		GrantTypes:              []string{"authorization_code"},
		ResponseTypes:           []string{"code"},
		ClientName:              "Multiple Scopes Test",
		TokenEndpointAuthMethod: "client_secret_basic",
		Scope:                   "openid profile email address phone",
	}

	response, statusCode, _ := ts.registerClientWithError(request)
	ts.Assert().Equal(http.StatusCreated, statusCode)
	ts.Assert().Contains(response.Scope, "openid")
	ts.Assert().Contains(response.Scope, "profile")
	ts.Assert().Contains(response.Scope, "email")

	if response.AppID != "" {
		ts.registeredAppIDs = append(ts.registeredAppIDs, response.AppID)
	}
}

// TestDCRRegistrationPublicClientWithWrongGrant tests public client validation.
func (ts *DCRTestSuite) TestDCRRegistrationPublicClientWithWrongGrant() {
	request := DCRRegistrationRequest{
		RedirectURIs:            []string{"https://client.example.com/callback"},
		GrantTypes:              []string{"client_credentials"},
		ResponseTypes:           []string{},
		ClientName:              "Public Client Wrong Grant",
		TokenEndpointAuthMethod: "none",
	}

	_, statusCode, _ := ts.registerClientWithError(request)
	ts.Assert().Equal(http.StatusBadRequest, statusCode)
}

// TestDCRRegistrationWithContacts tests DCR with contact information.
func (ts *DCRTestSuite) TestDCRRegistrationWithContacts() {
	request := DCRRegistrationRequest{
		RedirectURIs:            []string{"https://client.example.com/callback"},
		GrantTypes:              []string{"authorization_code"},
		ResponseTypes:           []string{"code"},
		ClientName:              "Contacts Test Client",
		TokenEndpointAuthMethod: "client_secret_basic",
		Contacts:                []string{"admin@example.com", "support@example.com"},
	}

	response, statusCode, _ := ts.registerClientWithError(request)
	ts.Assert().Equal(http.StatusCreated, statusCode)
	ts.Assert().Len(response.Contacts, 2)
	ts.Assert().Contains(response.Contacts, "admin@example.com")

	if response.AppID != "" {
		ts.registeredAppIDs = append(ts.registeredAppIDs, response.AppID)
	}
}

// TestDCRRegistrationWithPolicyAndTos tests DCR with policy and TOS URIs.
func (ts *DCRTestSuite) TestDCRRegistrationWithPolicyAndTos() {
	request := DCRRegistrationRequest{
		RedirectURIs:            []string{"https://client.example.com/callback"},
		GrantTypes:              []string{"authorization_code"},
		ResponseTypes:           []string{"code"},
		ClientName:              "Policy TOS Test Client",
		TokenEndpointAuthMethod: "client_secret_basic",
		PolicyURI:               "https://client.example.com/privacy",
		TosURI:                  "https://client.example.com/terms",
	}

	response, statusCode, _ := ts.registerClientWithError(request)
	ts.Assert().Equal(http.StatusCreated, statusCode)
	ts.Assert().Equal("https://client.example.com/privacy", response.PolicyURI)
	ts.Assert().Equal("https://client.example.com/terms", response.TosURI)

	if response.AppID != "" {
		ts.registeredAppIDs = append(ts.registeredAppIDs, response.AppID)
	}
}

// TestDCRRegistrationWithAllOptionalFields tests DCR with all optional fields.
func (ts *DCRTestSuite) TestDCRRegistrationWithAllOptionalFields() {
	jwks := map[string]interface{}{
		"keys": []map[string]interface{}{
			{
				"kty": "RSA", "use": "sig", "kid": "test-key-full",
				"n": "0vx7agoebGcQSuuPiLJXZptN9nndrQmbXEps2aiAFbWhM78LhWx4cbbfAAtVT86zwu1RK7aPFFxuhDR1L6tSoc_BJECPebWKRXjBZCiFV4n3oknjhMstn64tZ_2W-5JsGY4Hc5n9yBXArwl93lqt7_RN5w6Cf0h4QyQ5v-65YGjQR0_FDW2QvzqY368QQMicAtaSqzs8KJZgnYb9c7d0zgdAZHzu6qMQvRL5hajrn1n91CbOpbISD08qNLyrdkt-bFTWhAI4vMQFh6WeZu0fM4lFd2NcRwr3XPksINHaQ-G_xBniIqbw0Ls1jF44-csFCur-kEgU8awapJzKnqDKgw",
				"e": "AQAB",
			},
		},
	}

	request := DCRRegistrationRequest{
		RedirectURIs:            []string{"https://client.example.com/callback1", "https://client.example.com/callback2"},
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		ResponseTypes:           []string{"code"},
		ClientName:              "Full Optional Fields Test Client",
		ClientURI:               "https://client.example.com",
		LogoURI:                 "https://client.example.com/logo.png",
		PolicyURI:               "https://client.example.com/privacy",
		TosURI:                  "https://client.example.com/terms",
		Contacts:                []string{"admin@example.com", "support@example.com", "security@example.com"},
		TokenEndpointAuthMethod: "client_secret_basic",
		Scope:                   "openid profile email address phone offline_access",
		JWKS:                    jwks,
	}

	response, statusCode, _ := ts.registerClientWithError(request)
	ts.Assert().Equal(http.StatusCreated, statusCode)
	ts.Assert().NotEmpty(response.ClientID)
	ts.Assert().Equal(request.ClientName, response.ClientName)
	ts.Assert().Equal(request.ClientURI, response.ClientURI)
	ts.Assert().Equal(request.LogoURI, response.LogoURI)
	ts.Assert().Equal(request.PolicyURI, response.PolicyURI)
	ts.Assert().Equal(request.TosURI, response.TosURI)
	ts.Assert().Len(response.Contacts, 3)
	ts.Assert().Contains(response.Scope, "openid")
	ts.Assert().NotNil(response.JWKS)

	if response.AppID != "" {
		ts.registeredAppIDs = append(ts.registeredAppIDs, response.AppID)
	}
}

// TestDCRRegistrationAuthorizationCodeWithPKCE tests DCR with authorization_code and PKCE.
func (ts *DCRTestSuite) TestDCRRegistrationAuthorizationCodeWithPKCE() {
	request := DCRRegistrationRequest{
		RedirectURIs:            []string{"myapp://callback"},
		GrantTypes:              []string{"authorization_code"},
		ResponseTypes:           []string{"code"},
		ClientName:              "PKCE Native App",
		TokenEndpointAuthMethod: "none",
		Scope:                   "openid profile email",
	}

	response, statusCode, _ := ts.registerClientWithError(request)
	ts.Assert().Equal(http.StatusCreated, statusCode)
	ts.Assert().Equal("none", response.TokenEndpointAuthMethod)
	ts.Assert().Empty(response.ClientSecret) // Public client should not have secret

	if response.AppID != "" {
		ts.registeredAppIDs = append(ts.registeredAppIDs, response.AppID)
	}
}

// TestDCRRegistrationWithJWKSURIHTTPS tests JWKS_URI with HTTPS.
func (ts *DCRTestSuite) TestDCRRegistrationWithJWKSURIHTTPS() {
	request := DCRRegistrationRequest{
		RedirectURIs:            []string{"https://client.example.com/callback"},
		GrantTypes:              []string{"authorization_code"},
		ResponseTypes:           []string{"code"},
		ClientName:              "JWKS URI HTTPS Test",
		TokenEndpointAuthMethod: "client_secret_basic",
		JWKSUri:                 "https://secure.example.com/.well-known/jwks.json",
		Scope:                   "openid",
	}

	response, statusCode, _ := ts.registerClientWithError(request)
	ts.Assert().Equal(http.StatusCreated, statusCode)
	ts.Assert().Equal("https://secure.example.com/.well-known/jwks.json", response.JWKSUri)
	ts.Assert().Nil(response.JWKS)

	if response.AppID != "" {
		ts.registeredAppIDs = append(ts.registeredAppIDs, response.AppID)
	}
}

// TestDCRRegistrationSingleRedirectURI tests DCR with single redirect URI.
func (ts *DCRTestSuite) TestDCRRegistrationSingleRedirectURI() {
	request := DCRRegistrationRequest{
		RedirectURIs:            []string{"https://single.example.com/callback"},
		GrantTypes:              []string{"authorization_code"},
		ResponseTypes:           []string{"code"},
		ClientName:              "Single Redirect URI Test",
		TokenEndpointAuthMethod: "client_secret_post",
		Scope:                   "openid",
	}

	response, statusCode, _ := ts.registerClientWithError(request)
	ts.Assert().Equal(http.StatusCreated, statusCode)
	ts.Assert().Len(response.RedirectURIs, 1)
	ts.Assert().Equal("https://single.example.com/callback", response.RedirectURIs[0])

	if response.AppID != "" {
		ts.registeredAppIDs = append(ts.registeredAppIDs, response.AppID)
	}
}

// TestDCRRegistrationClientSecretPost tests DCR with client_secret_post auth method.
func (ts *DCRTestSuite) TestDCRRegistrationClientSecretPost() {
	request := DCRRegistrationRequest{
		RedirectURIs:            []string{"https://client.example.com/callback"},
		GrantTypes:              []string{"authorization_code"},
		ResponseTypes:           []string{"code"},
		ClientName:              "Client Secret Post Test",
		TokenEndpointAuthMethod: "client_secret_post",
		Scope:                   "openid profile",
	}

	response, statusCode, _ := ts.registerClientWithError(request)
	ts.Assert().Equal(http.StatusCreated, statusCode)
	ts.Assert().Equal("client_secret_post", response.TokenEndpointAuthMethod)
	ts.Assert().NotEmpty(response.ClientSecret) // Confidential client should have secret

	if response.AppID != "" {
		ts.registeredAppIDs = append(ts.registeredAppIDs, response.AppID)
	}
}

// TestDCRRegistrationWithAllMetadata tests DCR with comprehensive metadata.
func (ts *DCRTestSuite) TestDCRRegistrationWithAllMetadata() {
	request := DCRRegistrationRequest{
		RedirectURIs:            []string{"https://fullmeta.example.com/callback"},
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		ResponseTypes:           []string{"code"},
		ClientName:              "Full Metadata Client",
		ClientURI:               "https://fullmeta.example.com",
		LogoURI:                 "https://fullmeta.example.com/logo.png",
		TokenEndpointAuthMethod: "client_secret_basic",
		Scope:                   "openid profile email address phone",
		TosURI:                  "https://fullmeta.example.com/tos",
		PolicyURI:               "https://fullmeta.example.com/privacy",
		Contacts:                []string{"admin@fullmeta.example.com", "support@fullmeta.example.com"},
	}

	response, statusCode, _ := ts.registerClientWithError(request)
	ts.Assert().Equal(http.StatusCreated, statusCode)
	ts.Assert().Equal("Full Metadata Client", response.ClientName)
	ts.Assert().Equal("https://fullmeta.example.com", response.ClientURI)
	ts.Assert().Equal("https://fullmeta.example.com/logo.png", response.LogoURI)
	ts.Assert().Equal("https://fullmeta.example.com/tos", response.TosURI)
	ts.Assert().Equal("https://fullmeta.example.com/privacy", response.PolicyURI)
	ts.Assert().Equal([]string{"admin@fullmeta.example.com", "support@fullmeta.example.com"}, response.Contacts)
	ts.Assert().Equal("openid profile email address phone", response.Scope)

	if response.AppID != "" {
		ts.registeredAppIDs = append(ts.registeredAppIDs, response.AppID)
	}
}

// TestDCRRegistrationWithOnlyClientURI tests DCR with only client URI metadata.
func (ts *DCRTestSuite) TestDCRRegistrationWithOnlyClientURI() {
	request := DCRRegistrationRequest{
		RedirectURIs:            []string{"https://clienturi.example.com/callback"},
		GrantTypes:              []string{"authorization_code"},
		ResponseTypes:           []string{"code"},
		ClientURI:               "https://clienturi.example.com",
		TokenEndpointAuthMethod: "client_secret_basic",
		Scope:                   "openid",
	}

	response, statusCode, _ := ts.registerClientWithError(request)
	ts.Assert().Equal(http.StatusCreated, statusCode)
	ts.Assert().Equal("https://clienturi.example.com", response.ClientURI)
	ts.Assert().Empty(response.LogoURI)
	ts.Assert().Empty(response.TosURI)

	if response.AppID != "" {
		ts.registeredAppIDs = append(ts.registeredAppIDs, response.AppID)
	}
}

// TestDCRRegistrationPublicClientWithPKCE tests public client with PKCE.
func (ts *DCRTestSuite) TestDCRRegistrationPublicClientWithPKCE() {
	request := DCRRegistrationRequest{
		RedirectURIs:            []string{"https://public-pkce.example.com/callback"},
		GrantTypes:              []string{"authorization_code"},
		ResponseTypes:           []string{"code"},
		ClientName:              "Public PKCE Client",
		TokenEndpointAuthMethod: "none",
		Scope:                   "openid profile",
	}

	response, statusCode, _ := ts.registerClientWithError(request)
	ts.Assert().Equal(http.StatusCreated, statusCode)
	ts.Assert().Equal("none", response.TokenEndpointAuthMethod)
	ts.Assert().Empty(response.ClientSecret) // Public client should have no secret
	ts.Assert().NotEmpty(response.ClientID)

	if response.AppID != "" {
		ts.registeredAppIDs = append(ts.registeredAppIDs, response.AppID)
	}
}

// TestDCRRegistrationWithRefreshToken tests DCR with refresh_token grant.
func (ts *DCRTestSuite) TestDCRRegistrationWithRefreshToken() {
	request := DCRRegistrationRequest{
		RedirectURIs:            []string{"https://refresh.example.com/callback"},
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		ResponseTypes:           []string{"code"},
		ClientName:              "Refresh Token Coverage Client",
		TokenEndpointAuthMethod: "client_secret_basic",
		Scope:                   "openid offline_access",
	}

	response, statusCode, _ := ts.registerClientWithError(request)
	ts.Assert().Equal(http.StatusCreated, statusCode)
	ts.Assert().Contains(response.GrantTypes, "authorization_code")
	ts.Assert().Contains(response.GrantTypes, "refresh_token")
	ts.Assert().Contains(response.Scope, "offline_access")

	if response.AppID != "" {
		ts.registeredAppIDs = append(ts.registeredAppIDs, response.AppID)
	}
}

// TestDCRRegistrationClientCredentialsGrant tests DCR with client_credentials grant.
func (ts *DCRTestSuite) TestDCRRegistrationClientCredentialsGrant() {
	request := DCRRegistrationRequest{
		GrantTypes:              []string{"client_credentials"},
		ResponseTypes:           []string{},
		ClientName:              "Client Credentials Client",
		TokenEndpointAuthMethod: "client_secret_basic",
		Scope:                   "api:read api:write",
	}

	response, statusCode, _ := ts.registerClientWithError(request)
	ts.Assert().Equal(http.StatusCreated, statusCode)
	ts.Assert().Equal([]string{"client_credentials"}, response.GrantTypes)
	ts.Assert().Empty(response.RedirectURIs) // Client credentials doesn't need redirect URIs
	ts.Assert().NotEmpty(response.ClientSecret)

	if response.AppID != "" {
		ts.registeredAppIDs = append(ts.registeredAppIDs, response.AppID)
	}
}

// TestDCRRegistrationWithEmptyScope tests DCR with empty scope.
func (ts *DCRTestSuite) TestDCRRegistrationWithEmptyScope() {
	request := DCRRegistrationRequest{
		RedirectURIs:            []string{"https://noscope.example.com/callback"},
		GrantTypes:              []string{"authorization_code"},
		ResponseTypes:           []string{"code"},
		ClientName:              "No Scope Client",
		TokenEndpointAuthMethod: "client_secret_basic",
		Scope:                   "",
	}

	response, statusCode, _ := ts.registerClientWithError(request)
	ts.Assert().Equal(http.StatusCreated, statusCode)
	ts.Assert().Empty(response.Scope)

	if response.AppID != "" {
		ts.registeredAppIDs = append(ts.registeredAppIDs, response.AppID)
	}
}

// TestDCRRegistrationWithMultipleRedirectURIs tests DCR with multiple redirect URIs.
func (ts *DCRTestSuite) TestDCRRegistrationWithMultipleRedirectURIs() {
	request := DCRRegistrationRequest{
		RedirectURIs: []string{
			"https://multi.example.com/callback1",
			"https://multi.example.com/callback2",
			"https://multi.example.com/callback3",
		},
		GrantTypes:              []string{"authorization_code"},
		ResponseTypes:           []string{"code"},
		ClientName:              "Multi Redirect Client",
		TokenEndpointAuthMethod: "client_secret_basic",
		Scope:                   "openid",
	}

	response, statusCode, _ := ts.registerClientWithError(request)
	ts.Assert().Equal(http.StatusCreated, statusCode)
	ts.Assert().Len(response.RedirectURIs, 3)
	ts.Assert().Contains(response.RedirectURIs, "https://multi.example.com/callback1")
	ts.Assert().Contains(response.RedirectURIs, "https://multi.example.com/callback2")
	ts.Assert().Contains(response.RedirectURIs, "https://multi.example.com/callback3")

	if response.AppID != "" {
		ts.registeredAppIDs = append(ts.registeredAppIDs, response.AppID)
	}
}

// TestDCRRegistrationErrorInvalidGrantType tests error when invalid grant type is provided.
func (ts *DCRTestSuite) TestDCRRegistrationErrorInvalidGrantType() {
	request := DCRRegistrationRequest{
		RedirectURIs:            []string{"https://invalidgrant.example.com/callback"},
		GrantTypes:              []string{"invalid_grant_type"},
		ResponseTypes:           []string{"code"},
		ClientName:              "Invalid Grant Type Client",
		TokenEndpointAuthMethod: "client_secret_basic",
		Scope:                   "openid",
	}

	_, statusCode, _ := ts.registerClientWithError(request)
	ts.Assert().Equal(http.StatusBadRequest, statusCode)
}

// TestDCRRegistrationErrorInvalidResponseType tests error when invalid response type is provided.
func (ts *DCRTestSuite) TestDCRRegistrationErrorInvalidResponseType() {
	request := DCRRegistrationRequest{
		RedirectURIs:            []string{"https://invalidresponse.example.com/callback"},
		GrantTypes:              []string{"authorization_code"},
		ResponseTypes:           []string{"invalid_response_type"},
		ClientName:              "Invalid Response Type Client",
		TokenEndpointAuthMethod: "client_secret_basic",
		Scope:                   "openid",
	}

	_, statusCode, _ := ts.registerClientWithError(request)
	ts.Assert().Equal(http.StatusBadRequest, statusCode)
}

// TestDCRRegistrationErrorInvalidTokenEndpointAuthMethod tests error when invalid token endpoint auth method is provided.
func (ts *DCRTestSuite) TestDCRRegistrationErrorInvalidTokenEndpointAuthMethod() {
	request := DCRRegistrationRequest{
		RedirectURIs:            []string{"https://invalidauth.example.com/callback"},
		GrantTypes:              []string{"authorization_code"},
		ResponseTypes:           []string{"code"},
		ClientName:              "Invalid Auth Method Client",
		TokenEndpointAuthMethod: "invalid_auth_method",
		Scope:                   "openid",
	}

	_, statusCode, _ := ts.registerClientWithError(request)
	ts.Assert().Equal(http.StatusBadRequest, statusCode)
}

// TestDCRRegistrationErrorRedirectURIWithFragment tests error when redirect URI contains fragment.
func (ts *DCRTestSuite) TestDCRRegistrationErrorRedirectURIWithFragment() {
	request := DCRRegistrationRequest{
		RedirectURIs:            []string{"https://fragment.example.com/callback#fragment"},
		GrantTypes:              []string{"authorization_code"},
		ResponseTypes:           []string{"code"},
		ClientName:              "Fragment URI Client",
		TokenEndpointAuthMethod: "client_secret_basic",
		Scope:                   "openid",
	}

	_, statusCode, _ := ts.registerClientWithError(request)
	ts.Assert().Equal(http.StatusBadRequest, statusCode)
}

// TestDCRRegistrationErrorInvalidJWKSURI tests error when invalid JWKS URI is provided.
func (ts *DCRTestSuite) TestDCRRegistrationErrorInvalidJWKSURI() {
	request := DCRRegistrationRequest{
		RedirectURIs:            []string{"https://invalidjwks.example.com/callback"},
		GrantTypes:              []string{"authorization_code"},
		ResponseTypes:           []string{"code"},
		ClientName:              "Invalid JWKS URI Client",
		TokenEndpointAuthMethod: "client_secret_basic",
		JWKSUri:                 "not-a-valid-uri",
		Scope:                   "openid",
	}

	_, statusCode, _ := ts.registerClientWithError(request)
	ts.Assert().Equal(http.StatusBadRequest, statusCode)
}

// TestDCRRegistrationEmptyRequest tests that empty request body results in error handling.
func (ts *DCRTestSuite) TestDCRRegistrationEmptyRequest() {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("POST", testServerURL+dcrEndpoint, bytes.NewReader([]byte("{}")))
	ts.Require().NoError(err, "Failed to create request")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	ts.Require().NoError(err, "Failed to send request")
	defer resp.Body.Close()

	ts.Assert().Equal(http.StatusBadRequest, resp.StatusCode)

	var errResp DCRErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	ts.Require().NoError(err)
	ts.Assert().NotEmpty(errResp.Error)
}

// TestDCRRegistrationNullRedirectURIs tests that null redirect_uris field triggers proper error.
func (ts *DCRTestSuite) TestDCRRegistrationNullRedirectURIs() {
	// Create request with null redirect_uris by using a map
	requestData := map[string]interface{}{
		"redirect_uris": nil,
		"client_name":   "Null Redirect Test",
	}

	requestJSON, err := json.Marshal(requestData)
	ts.Require().NoError(err)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("POST", testServerURL+dcrEndpoint, bytes.NewReader(requestJSON))
	ts.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	ts.Require().NoError(err)
	defer resp.Body.Close()

	// Should return bad request
	ts.Assert().Equal(http.StatusBadRequest, resp.StatusCode)
}

// TestDCRRegistrationWithDuplicateClientName tests error when client name already exists.
func (ts *DCRTestSuite) TestDCRRegistrationWithDuplicateClientName() {
	duplicateName := "Duplicate Name Test Client"

	// First registration
	request1 := DCRRegistrationRequest{
		RedirectURIs:            []string{"https://example.com/callback"},
		GrantTypes:              []string{"authorization_code"},
		ClientName:              duplicateName,
		TokenEndpointAuthMethod: "client_secret_basic",
	}

	response1, statusCode1 := ts.registerClient(request1)
	ts.Assert().Equal(http.StatusCreated, statusCode1)
	ts.registeredAppIDs = append(ts.registeredAppIDs, response1.AppID)

	// Second registration with same name should fail
	request2 := DCRRegistrationRequest{
		RedirectURIs:            []string{"https://example.com/callback2"},
		GrantTypes:              []string{"authorization_code"},
		ClientName:              duplicateName,
		TokenEndpointAuthMethod: "client_secret_basic",
	}

	_, statusCode2, errResp := ts.registerClientWithError(request2)
	ts.Assert().Equal(http.StatusBadRequest, statusCode2)
	ts.Assert().NotNil(errResp)
	ts.Assert().NotEmpty(errResp.Error)
}

// TestDCRRegistrationWithInvalidScopeFormat tests that empty scope string is handled correctly.
func (ts *DCRTestSuite) TestDCRRegistrationWithInvalidScopeFormat() {
	// Test with valid but minimal scope
	request := DCRRegistrationRequest{
		RedirectURIs:            []string{"https://example.com/callback"},
		GrantTypes:              []string{"authorization_code"},
		ClientName:              "Minimal Scope Client",
		TokenEndpointAuthMethod: "client_secret_basic",
		Scope:                   "  ", // Whitespace only
	}

	response, statusCode := ts.registerClient(request)
	ts.Assert().Equal(http.StatusCreated, statusCode)
	ts.Assert().NotEmpty(response.ClientID)

	ts.registeredAppIDs = append(ts.registeredAppIDs, response.AppID)
}

// TestDCRRegistrationErrorResponseStructure tests the structure of error responses.
func (ts *DCRTestSuite) TestDCRRegistrationErrorResponseStructure() {
	testCases := []struct {
		name           string
		request        DCRRegistrationRequest
		expectedStatus int
	}{
		{
			name: "Invalid Logo URI Format",
			request: DCRRegistrationRequest{
				RedirectURIs: []string{"https://example.com/callback"},
				GrantTypes:   []string{"authorization_code"},
				ClientName:   "Invalid Logo URI Test",
				LogoURI:      "not-a-valid-uri",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing Required Redirect URIs",
			request: DCRRegistrationRequest{
				ClientName:              "Missing Redirect URI Test",
				GrantTypes:              []string{"authorization_code"},
				TokenEndpointAuthMethod: "client_secret_basic",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid Grant Type",
			request: DCRRegistrationRequest{
				RedirectURIs:            []string{"https://example.com/callback"},
				GrantTypes:              []string{"invalid_grant_type"},
				ClientName:              "Invalid Grant Type Test",
				TokenEndpointAuthMethod: "client_secret_basic",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid Response Type",
			request: DCRRegistrationRequest{
				RedirectURIs:            []string{"https://example.com/callback"},
				GrantTypes:              []string{"authorization_code"},
				ResponseTypes:           []string{"invalid_response_type"},
				ClientName:              "Invalid Response Type Test",
				TokenEndpointAuthMethod: "client_secret_basic",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		ts.Run(tc.name, func() {
			_, statusCode, errResp := ts.registerClientWithError(tc.request)
			ts.Assert().Equal(tc.expectedStatus, statusCode)
			ts.Assert().NotNil(errResp)
			ts.Assert().NotEmpty(errResp.Error)
			ts.Assert().NotNil(errResp.Error)
		})
	}
}

// TestDCRRegistrationWithClientCredentialsNoRedirects tests client_credentials without redirect URIs.
func (ts *DCRTestSuite) TestDCRRegistrationWithClientCredentialsNoRedirects() {
	request := DCRRegistrationRequest{
		GrantTypes:              []string{"client_credentials"},
		ClientName:              "Client Credentials No Redirects",
		TokenEndpointAuthMethod: "client_secret_basic",
	}

	response, statusCode := ts.registerClient(request)
	ts.Assert().Equal(http.StatusCreated, statusCode)
	ts.Assert().NotEmpty(response.ClientID)
	ts.Assert().NotEmpty(response.ClientSecret)
	ts.Assert().Equal([]string{"client_credentials"}, response.GrantTypes)
	ts.Assert().Empty(response.RedirectURIs)
	ts.Assert().Empty(response.ResponseTypes)

	ts.registeredAppIDs = append(ts.registeredAppIDs, response.AppID)
}

// TestDCRRegistrationWithEmptyJWKSAndJWKSUri tests that JWKS configuration conflict is caught.
func (ts *DCRTestSuite) TestDCRRegistrationWithEmptyJWKSAndJWKSUri() {
	// JWKS with empty keys array
	jwks := map[string]interface{}{
		"keys": []map[string]interface{}{},
	}

	request := DCRRegistrationRequest{
		RedirectURIs:            []string{"https://example.com/callback"},
		GrantTypes:              []string{"authorization_code"},
		ClientName:              "Empty JWKS Client",
		TokenEndpointAuthMethod: "client_secret_basic",
		JWKSUri:                 "https://example.com/.well-known/jwks.json",
		JWKS:                    jwks,
	}

	_, statusCode, errResp := ts.registerClientWithError(request)
	ts.Assert().Equal(http.StatusBadRequest, statusCode)
	ts.Assert().NotNil(errResp)
	ts.Assert().Contains(errResp.Error, "invalid_client_metadata")
}
