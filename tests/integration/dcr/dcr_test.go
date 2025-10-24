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
