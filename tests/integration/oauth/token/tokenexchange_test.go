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

package token

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/asgardeo/thunder/tests/integration/testutils"
	"github.com/stretchr/testify/suite"
)

const (
	tokenExchangeClientID     = "token_exchange_test_client"
	tokenExchangeClientSecret = "token_exchange_test_secret"
	tokenExchangeAppName      = "TokenExchangeTestApp"
	tokenExchangeTestUser     = "te_test_user"
	tokenExchangeTestPassword = "TePassword123!"
	tokenExchangeTestEmail    = "te_test@example.com"
)

type TokenExchangeTestSuite struct {
	suite.Suite
	applicationID  string
	userID         string
	client         *http.Client
	assertionToken string
}

func TestTokenExchangeTestSuite(t *testing.T) {
	suite.Run(t, new(TokenExchangeTestSuite))
}

func (ts *TokenExchangeTestSuite) SetupSuite() {
	// Create HTTP client that skips TLS verification
	ts.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// Create test user
	ts.userID = ts.createTestUser()

	// Create OAuth application with token exchange grant type
	ts.applicationID = ts.createTestApplication()

	// Authenticate user to get assertion token for tests
	ts.assertionToken = ts.getUserAssertion()
}

func (ts *TokenExchangeTestSuite) TearDownSuite() {
	// Clean up application
	if ts.applicationID != "" {
		ts.deleteApplication(ts.applicationID)
	}

	// Clean up user
	if ts.userID != "" {
		testutils.DeleteUser(ts.userID)
	}
}

func (ts *TokenExchangeTestSuite) createTestUser() string {
	attributes := map[string]interface{}{
		"username": tokenExchangeTestUser,
		"password": tokenExchangeTestPassword,
		"email":    tokenExchangeTestEmail,
	}

	attributesJSON, err := json.Marshal(attributes)
	ts.Require().NoError(err, "Failed to marshal user attributes")

	user := testutils.User{
		Type:             "person",
		OrganizationUnit: "root",
		Attributes:       json.RawMessage(attributesJSON),
	}

	userID, err := testutils.CreateUser(user)
	ts.Require().NoError(err, "Failed to create test user")
	ts.T().Logf("Created test user with ID: %s", userID)

	return userID
}

func (ts *TokenExchangeTestSuite) createTestApplication() string {
	app := map[string]interface{}{
		"name":                         tokenExchangeAppName,
		"description":                  "Application for token exchange integration tests",
		"auth_flow_graph_id":           "auth_flow_config_basic",
		"registration_flow_graph_id":   "registration_flow_config_basic",
		"is_registration_flow_enabled": true,
		"inbound_auth_config": []map[string]interface{}{
			{
				"type": "oauth2",
				"config": map[string]interface{}{
					"client_id":     tokenExchangeClientID,
					"client_secret": tokenExchangeClientSecret,
					"redirect_uris": []string{"https://localhost:3000"},
					"grant_types": []string{
						"urn:ietf:params:oauth:grant-type:token-exchange",
						"authorization_code",
					},
					"response_types":             []string{"code"},
					"token_endpoint_auth_method": "client_secret_basic",
					"scopes":                     []string{"openid", "profile", "email", "read", "write"},
				},
			},
		},
	}

	jsonData, err := json.Marshal(app)
	ts.Require().NoError(err, "Failed to marshal application data")

	req, err := http.NewRequest("POST", testutils.TestServerURL+"/applications", bytes.NewBuffer(jsonData))
	ts.Require().NoError(err, "Failed to create request")
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.client.Do(req)
	ts.Require().NoError(err, "Failed to create application")
	defer resp.Body.Close()

	ts.Require().Equal(http.StatusCreated, resp.StatusCode, "Failed to create application")

	var respData map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&respData)
	ts.Require().NoError(err, "Failed to parse response")

	appID := respData["id"].(string)
	ts.T().Logf("Created test application with ID: %s", appID)
	return appID
}

func (ts *TokenExchangeTestSuite) deleteApplication(appID string) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/applications/%s", testutils.TestServerURL, appID), nil)
	if err != nil {
		ts.T().Errorf("Failed to create delete request: %v", err)
		return
	}

	resp, err := ts.client.Do(req)
	if err != nil {
		ts.T().Errorf("Failed to delete application: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		ts.T().Errorf("Failed to delete application. Status: %d, Response: %s", resp.StatusCode, string(bodyBytes))
	} else {
		ts.T().Logf("Successfully deleted test application with ID: %s", appID)
	}
}

func (ts *TokenExchangeTestSuite) getUserAssertion() string {
	authRequest := map[string]interface{}{
		"username": tokenExchangeTestUser,
		"password": tokenExchangeTestPassword,
	}

	requestJSON, err := json.Marshal(authRequest)
	ts.Require().NoError(err, "Failed to marshal auth request")

	req, err := http.NewRequest("POST", testutils.TestServerURL+"/auth/credentials/authenticate", bytes.NewReader(requestJSON))
	ts.Require().NoError(err, "Failed to create auth request")
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.client.Do(req)
	ts.Require().NoError(err, "Failed to authenticate user")
	defer resp.Body.Close()

	ts.Require().Equal(http.StatusOK, resp.StatusCode, "Authentication failed")

	var authResponse testutils.AuthenticationResponse
	err = json.NewDecoder(resp.Body).Decode(&authResponse)
	ts.Require().NoError(err, "Failed to parse auth response")
	ts.Require().NotEmpty(authResponse.Assertion, "Assertion token should not be empty")

	return authResponse.Assertion
}

// getAccessTokenWithScopes gets an access token with the specified scopes using authorization code grant
func (ts *TokenExchangeTestSuite) getAccessTokenWithScopes(requestedScopes string) string {
	redirectURI := "https://localhost:3000"
	state := "test_state_123"

	authzURL := testutils.TestServerURL + "/oauth2/authorize"
	params := url.Values{}
	params.Set("client_id", tokenExchangeClientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("response_type", "code")
	params.Set("scope", requestedScopes)
	params.Set("state", state)

	req, err := http.NewRequest("GET", authzURL+"?"+params.Encode(), nil)
	ts.Require().NoError(err, "Failed to create authorization request")

	// Use a client that doesn't automatically follow redirects
	authzClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	authzResp, err := authzClient.Do(req)
	ts.Require().NoError(err, "Failed to send authorization request")
	defer authzResp.Body.Close()

	ts.Require().Equal(http.StatusFound, authzResp.StatusCode, "Expected redirect")
	location := authzResp.Header.Get("Location")
	ts.Require().NotEmpty(location, "Expected redirect location")

	// Extract sessionDataKey from redirect
	parsedURL, err := url.Parse(location)
	ts.Require().NoError(err)
	sessionDataKey := parsedURL.Query().Get("sessionDataKey")
	ts.Require().NotEmpty(sessionDataKey, "Expected sessionDataKey")

	// Step 2: Execute authentication flow
	flowData := map[string]interface{}{
		"applicationId": ts.applicationID,
		"flowType":      "AUTHENTICATION",
		"inputs": map[string]string{
			"username": tokenExchangeTestUser,
			"password": tokenExchangeTestPassword,
		},
	}

	flowJSON, err := json.Marshal(flowData)
	ts.Require().NoError(err)

	flowReq, err := http.NewRequest("POST", testutils.TestServerURL+"/flow/execute", bytes.NewReader(flowJSON))
	ts.Require().NoError(err)
	flowReq.Header.Set("Content-Type", "application/json")

	flowResp, err := ts.client.Do(flowReq)
	ts.Require().NoError(err)
	defer flowResp.Body.Close()

	ts.Require().Equal(http.StatusOK, flowResp.StatusCode)
	var flowStep map[string]interface{}
	err = json.NewDecoder(flowResp.Body).Decode(&flowStep)
	ts.Require().NoError(err)
	assertion := flowStep["assertion"].(string)
	ts.Require().NotEmpty(assertion)

	// Step 3: Complete authorization
	authzCompleteData := map[string]interface{}{
		"sessionDataKey": sessionDataKey,
		"assertion":      assertion,
	}

	authzCompleteJSON, err := json.Marshal(authzCompleteData)
	ts.Require().NoError(err)

	authzCompleteReq, err := http.NewRequest("POST", testutils.TestServerURL+"/oauth2/authorize", bytes.NewReader(authzCompleteJSON))
	ts.Require().NoError(err)
	authzCompleteReq.Header.Set("Content-Type", "application/json")

	authzCompleteResp, err := ts.client.Do(authzCompleteReq)
	ts.Require().NoError(err)
	defer authzCompleteResp.Body.Close()

	ts.Require().Equal(http.StatusOK, authzCompleteResp.StatusCode)
	var authzCompleteResponse map[string]interface{}
	err = json.NewDecoder(authzCompleteResp.Body).Decode(&authzCompleteResponse)
	ts.Require().NoError(err)

	redirectURIWithCode, ok := authzCompleteResponse["redirect_uri"].(string)
	ts.Require().True(ok, "redirect_uri should be present and a string")
	ts.Require().NotEmpty(redirectURIWithCode, "redirect_uri should not be empty")
	parsedRedirect, err := url.Parse(redirectURIWithCode)
	ts.Require().NoError(err)
	code := parsedRedirect.Query().Get("code")
	ts.Require().NotEmpty(code, "Expected authorization code")

	// Step 4: Exchange authorization code for access token
	tokenData := url.Values{}
	tokenData.Set("grant_type", "authorization_code")
	tokenData.Set("code", code)
	tokenData.Set("redirect_uri", redirectURI)

	tokenReq, err := http.NewRequest("POST", testutils.TestServerURL+"/oauth2/token", strings.NewReader(tokenData.Encode()))
	ts.Require().NoError(err)
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	tokenReq.SetBasicAuth(tokenExchangeClientID, tokenExchangeClientSecret)

	tokenResp, err := ts.client.Do(tokenReq)
	ts.Require().NoError(err)
	defer tokenResp.Body.Close()

	ts.Require().Equal(http.StatusOK, tokenResp.StatusCode)

	var tokenResponse map[string]interface{}
	err = json.NewDecoder(tokenResp.Body).Decode(&tokenResponse)
	ts.Require().NoError(err)

	accessToken := tokenResponse["access_token"].(string)
	ts.Require().NotEmpty(accessToken, "Expected access token")

	return accessToken
}

func (ts *TokenExchangeTestSuite) exchangeToken(requestBody string, authHeader string) (*TokenExchangeResponse, int, error) {
	req, err := http.NewRequest("POST", testutils.TestServerURL+"/oauth2/token", strings.NewReader(requestBody))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := ts.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	var tokenResp TokenExchangeResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusOK {
		err = json.Unmarshal(bodyBytes, &tokenResp)
		if err != nil {
			return nil, resp.StatusCode, err
		}
		return &tokenResp, resp.StatusCode, nil
	}

	// Parse error response
	var errorResp map[string]interface{}
	_ = json.Unmarshal(bodyBytes, &errorResp)
	tokenResp.Error = fmt.Sprintf("%v", errorResp["error"])
	if desc, ok := errorResp["error_description"]; ok {
		tokenResp.ErrorDescription = fmt.Sprintf("%v", desc)
	}
	return &tokenResp, resp.StatusCode, nil
}

type TokenExchangeResponse struct {
	AccessToken      string `json:"access_token,omitempty"`
	TokenType        string `json:"token_type,omitempty"`
	ExpiresIn        int64  `json:"expires_in,omitempty"`
	IssuedTokenType  string `json:"issued_token_type,omitempty"`
	Scope            string `json:"scope,omitempty"`
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
}

// TestTokenExchange_BasicSuccess tests basic successful token exchange
func (ts *TokenExchangeTestSuite) TestTokenExchange_BasicSuccess() {
	formData := url.Values{}
	formData.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
	formData.Set("subject_token", ts.assertionToken)
	formData.Set("subject_token_type", "urn:ietf:params:oauth:token-type:jwt")

	authHeader := "Basic " + basicAuth(tokenExchangeClientID, tokenExchangeClientSecret)

	resp, statusCode, err := ts.exchangeToken(formData.Encode(), authHeader)
	ts.Require().NoError(err)
	ts.Equal(http.StatusOK, statusCode)
	ts.NotEmpty(resp.AccessToken, "Access token should be present")
	ts.Equal("Bearer", resp.TokenType, "Token type should be Bearer")
	ts.NotZero(resp.ExpiresIn, "Expires in should be set")
	ts.Equal("urn:ietf:params:oauth:token-type:access_token", resp.IssuedTokenType)

	// Verify the access token is a valid JWT
	claims, err := testutils.DecodeJWT(resp.AccessToken)
	ts.Require().NoError(err, "Access token should be a valid JWT")
	ts.Equal(ts.userID, claims.Sub, "Subject should match user ID")
}

// TestTokenExchange_WithScopes tests token exchange with scope downscoping
// First gets an access token with scopes, then exchanges it requesting a subset
func (ts *TokenExchangeTestSuite) TestTokenExchange_WithScopes() {
	// First, get an access token with scopes using authorization code grant
	accessTokenWithScopes := ts.getAccessTokenWithScopes("read write profile email")

	// Now use that access token as the subject token and request a subset (downscoping)
	formData := url.Values{}
	formData.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
	formData.Set("subject_token", accessTokenWithScopes)
	formData.Set("subject_token_type", "urn:ietf:params:oauth:token-type:access_token")
	formData.Set("scope", "read write") // Request subset of scopes

	authHeader := "Basic " + basicAuth(tokenExchangeClientID, tokenExchangeClientSecret)

	resp, statusCode, err := ts.exchangeToken(formData.Encode(), authHeader)
	ts.Require().NoError(err)
	ts.Equal(http.StatusOK, statusCode)
	ts.NotEmpty(resp.AccessToken)
	ts.NotEmpty(resp.Scope, "Scope should be present in response")

	scopes := strings.Fields(resp.Scope)
	ts.Contains(scopes, "read", "Response should contain 'read' scope")
	ts.Contains(scopes, "write", "Response should contain 'write' scope")
	ts.NotContains(scopes, "profile", "Response should not contain 'profile' scope (downscoped)")

	// Verify scopes in JWT
	claims, err := testutils.DecodeJWT(resp.AccessToken)
	ts.Require().NoError(err)
	if scopeClaim, ok := claims.Additional["scope"].(string); ok {
		scopeList := strings.Fields(scopeClaim)
		ts.Contains(scopeList, "read")
		ts.Contains(scopeList, "write")
		ts.NotContains(scopeList, "profile")
	}
}

// TestTokenExchange_WithAudience tests token exchange with audience parameter
func (ts *TokenExchangeTestSuite) TestTokenExchange_WithAudience() {
	formData := url.Values{}
	formData.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
	formData.Set("subject_token", ts.assertionToken)
	formData.Set("subject_token_type", "urn:ietf:params:oauth:token-type:jwt")
	formData.Set("audience", "https://api.example.com")

	authHeader := "Basic " + basicAuth(tokenExchangeClientID, tokenExchangeClientSecret)

	resp, statusCode, err := ts.exchangeToken(formData.Encode(), authHeader)
	ts.Require().NoError(err)
	ts.Equal(http.StatusOK, statusCode)
	ts.NotEmpty(resp.AccessToken)

	// Verify audience in JWT
	claims, err := testutils.DecodeJWT(resp.AccessToken)
	ts.Require().NoError(err)
	ts.Equal("https://api.example.com", claims.Aud, "Audience should match requested audience")
}

// TestTokenExchange_WithResource tests token exchange with resource parameter
func (ts *TokenExchangeTestSuite) TestTokenExchange_WithResource() {
	formData := url.Values{}
	formData.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
	formData.Set("subject_token", ts.assertionToken)
	formData.Set("subject_token_type", "urn:ietf:params:oauth:token-type:jwt")
	formData.Set("resource", "https://resource.example.com")

	authHeader := "Basic " + basicAuth(tokenExchangeClientID, tokenExchangeClientSecret)

	resp, statusCode, err := ts.exchangeToken(formData.Encode(), authHeader)
	ts.Require().NoError(err)
	ts.Equal(http.StatusOK, statusCode)
	ts.NotEmpty(resp.AccessToken)

	// Verify resource is used as audience
	claims, err := testutils.DecodeJWT(resp.AccessToken)
	ts.Require().NoError(err)
	ts.Equal("https://resource.example.com", claims.Aud, "Audience should match resource parameter")
}

// TestTokenExchange_WithRequestedTokenType tests token exchange with requested_token_type
func (ts *TokenExchangeTestSuite) TestTokenExchange_WithRequestedTokenType() {
	formData := url.Values{}
	formData.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
	formData.Set("subject_token", ts.assertionToken)
	formData.Set("subject_token_type", "urn:ietf:params:oauth:token-type:jwt")
	formData.Set("requested_token_type", "urn:ietf:params:oauth:token-type:access_token")

	authHeader := "Basic " + basicAuth(tokenExchangeClientID, tokenExchangeClientSecret)

	resp, statusCode, err := ts.exchangeToken(formData.Encode(), authHeader)
	ts.Require().NoError(err)
	ts.Equal(http.StatusOK, statusCode)
	ts.NotEmpty(resp.AccessToken)
	ts.Equal("urn:ietf:params:oauth:token-type:access_token", resp.IssuedTokenType)
}

// TestTokenExchange_MissingSubjectToken tests error when subject_token is missing
func (ts *TokenExchangeTestSuite) TestTokenExchange_MissingSubjectToken() {
	formData := url.Values{}
	formData.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
	formData.Set("subject_token_type", "urn:ietf:params:oauth:token-type:jwt")

	authHeader := "Basic " + basicAuth(tokenExchangeClientID, tokenExchangeClientSecret)

	resp, statusCode, err := ts.exchangeToken(formData.Encode(), authHeader)
	ts.Require().NoError(err)
	ts.Equal(http.StatusBadRequest, statusCode)
	ts.Equal("invalid_request", resp.Error)
	ts.Contains(resp.ErrorDescription, "subject_token")
}

// TestTokenExchange_MissingSubjectTokenType tests error when subject_token_type is missing
func (ts *TokenExchangeTestSuite) TestTokenExchange_MissingSubjectTokenType() {
	formData := url.Values{}
	formData.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
	formData.Set("subject_token", ts.assertionToken)

	authHeader := "Basic " + basicAuth(tokenExchangeClientID, tokenExchangeClientSecret)

	resp, statusCode, err := ts.exchangeToken(formData.Encode(), authHeader)
	ts.Require().NoError(err)
	ts.Equal(http.StatusBadRequest, statusCode)
	ts.Equal("invalid_request", resp.Error)
	ts.Contains(resp.ErrorDescription, "subject_token_type")
}

// TestTokenExchange_InvalidSubjectToken tests error with invalid subject token
func (ts *TokenExchangeTestSuite) TestTokenExchange_InvalidSubjectToken() {
	formData := url.Values{}
	formData.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
	formData.Set("subject_token", "invalid.jwt.token")
	formData.Set("subject_token_type", "urn:ietf:params:oauth:token-type:jwt")

	authHeader := "Basic " + basicAuth(tokenExchangeClientID, tokenExchangeClientSecret)

	resp, statusCode, err := ts.exchangeToken(formData.Encode(), authHeader)
	ts.Require().NoError(err)
	ts.Equal(http.StatusBadRequest, statusCode)
	ts.Equal("invalid_grant", resp.Error)
	ts.Contains(resp.ErrorDescription, "Invalid subject_token")
}

// TestTokenExchange_UnsupportedSubjectTokenType tests error with unsupported token type
func (ts *TokenExchangeTestSuite) TestTokenExchange_UnsupportedSubjectTokenType() {
	formData := url.Values{}
	formData.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
	formData.Set("subject_token", ts.assertionToken)
	formData.Set("subject_token_type", "urn:ietf:params:oauth:token-type:saml2")

	authHeader := "Basic " + basicAuth(tokenExchangeClientID, tokenExchangeClientSecret)

	resp, statusCode, err := ts.exchangeToken(formData.Encode(), authHeader)
	ts.Require().NoError(err)
	ts.Equal(http.StatusBadRequest, statusCode)
	ts.Equal("invalid_request", resp.Error)
	ts.Contains(resp.ErrorDescription, "Unsupported subject_token_type")
}

// TestTokenExchange_InvalidClientCredentials tests error with invalid client credentials
func (ts *TokenExchangeTestSuite) TestTokenExchange_InvalidClientCredentials() {
	formData := url.Values{}
	formData.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
	formData.Set("subject_token", ts.assertionToken)
	formData.Set("subject_token_type", "urn:ietf:params:oauth:token-type:jwt")

	authHeader := "Basic " + basicAuth("invalid_client", "invalid_secret")

	resp, statusCode, err := ts.exchangeToken(formData.Encode(), authHeader)
	ts.Require().NoError(err)
	ts.Equal(http.StatusUnauthorized, statusCode)
	ts.Equal("invalid_client", resp.Error)
}

// TestTokenExchange_ApplicationNotRegisteredForGrantType tests error when app doesn't have grant type
func (ts *TokenExchangeTestSuite) TestTokenExchange_ApplicationNotRegisteredForGrantType() {
	// Create an application without token exchange grant type
	app := map[string]interface{}{
		"name":                         tokenExchangeAppName + "_no_te",
		"description":                  "Application without token exchange",
		"auth_flow_graph_id":           "auth_flow_config_basic",
		"registration_flow_graph_id":   "registration_flow_config_basic",
		"is_registration_flow_enabled": true,
		"inbound_auth_config": []map[string]interface{}{
			{
				"type": "oauth2",
				"config": map[string]interface{}{
					"client_id":                  tokenExchangeClientID + "_no_te",
					"client_secret":              tokenExchangeClientSecret,
					"redirect_uris":              []string{"https://localhost:3000"},
					"grant_types":                []string{"authorization_code"},
					"response_types":             []string{"code"},
					"token_endpoint_auth_method": "client_secret_basic",
				},
			},
		},
	}

	jsonData, _ := json.Marshal(app)
	req, _ := http.NewRequest("POST", testutils.TestServerURL+"/applications", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp, err := ts.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		var respData map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&respData)
		appID := respData["id"].(string)

		defer func() {
			req, _ := http.NewRequest("DELETE", testutils.TestServerURL+"/applications/"+appID, nil)
			ts.client.Do(req)
		}()

		// Try token exchange with app that doesn't support it
		formData := url.Values{}
		formData.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
		formData.Set("subject_token", ts.assertionToken)
		formData.Set("subject_token_type", "urn:ietf:params:oauth:token-type:jwt")

		authHeader := "Basic " + basicAuth(tokenExchangeClientID+"_no_te", tokenExchangeClientSecret)

		tokenResp, statusCode, err := ts.exchangeToken(formData.Encode(), authHeader)
		ts.Require().NoError(err)
		ts.Equal(http.StatusUnauthorized, statusCode)
		ts.Equal("unauthorized_client", tokenResp.Error)
		ts.Contains(tokenResp.ErrorDescription, "not authorized")
	}
}

// TestTokenExchange_PreservesUserAttributes tests that user attributes are preserved in exchanged token
func (ts *TokenExchangeTestSuite) TestTokenExchange_PreservesUserAttributes() {
	formData := url.Values{}
	formData.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
	formData.Set("subject_token", ts.assertionToken)
	formData.Set("subject_token_type", "urn:ietf:params:oauth:token-type:jwt")

	authHeader := "Basic " + basicAuth(tokenExchangeClientID, tokenExchangeClientSecret)

	resp, statusCode, err := ts.exchangeToken(formData.Encode(), authHeader)
	ts.Require().NoError(err)
	ts.Equal(http.StatusOK, statusCode)

	// Verify user attributes are preserved
	claims, err := testutils.DecodeJWT(resp.AccessToken)
	ts.Require().NoError(err)
	ts.Equal(ts.userID, claims.Sub, "Subject should match user ID")
	// Check that user type and other attributes from assertion are present
	if userType, ok := claims.Additional["userType"].(string); ok {
		ts.NotEmpty(userType, "User type should be preserved")
	}
}

// TestTokenExchange_DefaultIssuedTokenType tests that default issued_token_type is used when not specified
func (ts *TokenExchangeTestSuite) TestTokenExchange_DefaultIssuedTokenType() {
	formData := url.Values{}
	formData.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
	formData.Set("subject_token", ts.assertionToken)
	formData.Set("subject_token_type", "urn:ietf:params:oauth:token-type:jwt")
	// Note: requested_token_type is not set

	authHeader := "Basic " + basicAuth(tokenExchangeClientID, tokenExchangeClientSecret)

	resp, statusCode, err := ts.exchangeToken(formData.Encode(), authHeader)
	ts.Require().NoError(err)
	ts.Equal(http.StatusOK, statusCode)
	// Should default to access_token type
	ts.Equal("urn:ietf:params:oauth:token-type:access_token", resp.IssuedTokenType)
}
