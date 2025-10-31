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

package authz

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
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
	clientID     = "authz_test_client_123"
	clientSecret = "authz_test_secret_123"
	appName      = "AuthzTestApp"
	redirectURI  = "https://localhost:3000"
)

var (
	testOUID string
)

type AuthzTestSuite struct {
	suite.Suite
	applicationID string
	client        *http.Client
}

func TestAuthzTestSuite(t *testing.T) {
	suite.Run(t, new(AuthzTestSuite))
}

func (ts *AuthzTestSuite) SetupSuite() {

	ts.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	app := map[string]interface{}{
		"name":                         appName,
		"description":                  "Application for authorization integration tests",
		"auth_flow_graph_id":           "auth_flow_config_basic",
		"registration_flow_graph_id":   "registration_flow_config_basic",
		"is_registration_flow_enabled": true,
		"inbound_auth_config": []map[string]interface{}{
			{
				"type": "oauth2",
				"config": map[string]interface{}{
					"client_id":     clientID,
					"client_secret": clientSecret,
					"redirect_uris": []string{redirectURI},
					"grant_types": []string{
						"client_credentials",
						"authorization_code",
						"refresh_token",
					},
					"response_types": []string{
						"code",
					},
					"token_endpoint_auth_method": "client_secret_basic",
				},
			},
		},
	}

	jsonData, err := json.Marshal(app)
	if err != nil {
		ts.T().Fatalf("Failed to marshal application data: %v", err)
	}

	req, err := http.NewRequest("POST", testServerURL+"/applications", bytes.NewBuffer(jsonData))
	if err != nil {
		ts.T().Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.client.Do(req)
	if err != nil {
		ts.T().Fatalf("Failed to create application: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		ts.T().Fatalf("Failed to create application. Status: %d, Response: %s", resp.StatusCode, string(bodyBytes))
	}

	var respData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		ts.T().Fatalf("Failed to parse response: %v", err)
	}

	ts.applicationID = respData["id"].(string)
	ts.T().Logf("Created test application with ID: %s", ts.applicationID)

	// Create test organization unit for user creation
	ouData := map[string]interface{}{
		"handle":      "oauth2-authz-test-ou",
		"name":        "OAuth2 Authorization Test OU",
		"description": "Organization unit for OAuth2 authorization testing",
		"parent":      nil,
	}

	ouJSON, err := json.Marshal(ouData)
	if err != nil {
		ts.T().Fatalf("Failed to marshal OU data: %v", err)
	}

	ouReqBody := bytes.NewReader(ouJSON)
	ouReq, err := http.NewRequest("POST", testServerURL+"/organization-units", ouReqBody)
	if err != nil {
		ts.T().Fatalf("Failed to create OU request: %v", err)
	}
	ouReq.Header.Set("Content-Type", "application/json")

	ouResp, err := ts.client.Do(ouReq)
	if err != nil {
		ts.T().Fatalf("Failed to send OU request: %v", err)
	}
	defer ouResp.Body.Close()

	if ouResp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(ouResp.Body)
		ts.T().Fatalf("Failed to create OU. Status: %d, Response: %s", ouResp.StatusCode, string(bodyBytes))
	}

	var ouRespData map[string]interface{}
	if err := json.NewDecoder(ouResp.Body).Decode(&ouRespData); err != nil {
		ts.T().Fatalf("Failed to parse OU response: %v", err)
	}

	testOUID = ouRespData["id"].(string)
	ts.T().Logf("Created test organization unit with ID: %s", testOUID)
}

func (ts *AuthzTestSuite) TearDownSuite() {
	if ts.applicationID == "" {
		return
	}

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/applications/%s", testServerURL, ts.applicationID), nil)
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
		ts.T().Errorf("Failed to delete application. Status: %d", resp.StatusCode)
		return
	}

	ts.T().Logf("Successfully deleted test application with ID: %s", ts.applicationID)

	// Delete test organization unit
	if testOUID != "" {
		ouReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/organization-units/%s", testServerURL, testOUID), nil)
		if err != nil {
			ts.T().Errorf("Failed to create OU delete request: %v", err)
			return
		}

		ouResp, err := ts.client.Do(ouReq)
		if err != nil {
			ts.T().Errorf("Failed to delete organization unit: %v", err)
			return
		}
		defer ouResp.Body.Close()

		if ouResp.StatusCode != http.StatusNoContent {
			ts.T().Errorf("Failed to delete organization unit. Status: %d", ouResp.StatusCode)
		} else {
			ts.T().Logf("Successfully deleted test organization unit with ID: %s", testOUID)
		}
	}

}

// TestBasicAuthorizationRequest tests the basic authorization request flow
func (ts *AuthzTestSuite) TestBasicAuthorizationRequest() {
	testCases := []TestCase{
		{
			Name:           "Valid Request",
			ClientID:       clientID,
			RedirectURI:    redirectURI,
			ResponseType:   "code",
			Scope:          "openid",
			State:          "test_state_123",
			ExpectedStatus: http.StatusFound,
		},
		{
			Name:           "Invalid Client ID",
			ClientID:       "invalid_client",
			RedirectURI:    redirectURI,
			ResponseType:   "code",
			Scope:          "openid",
			State:          "test_state_456",
			ExpectedStatus: http.StatusFound,
			ExpectedError:  "invalid_client",
		},
		{
			Name:           "Invalid Response Type",
			ClientID:       clientID,
			RedirectURI:    redirectURI,
			ResponseType:   "invalid_type",
			Scope:          "openid",
			State:          "test_state_789",
			ExpectedStatus: http.StatusFound,
			ExpectedError:  "unsupported_response_type",
		},
		{
			Name:           "Missing Client ID",
			ClientID:       "",
			RedirectURI:    redirectURI,
			ResponseType:   "code",
			Scope:          "openid",
			State:          "test_state_missing_client",
			ExpectedStatus: http.StatusFound,
			ExpectedError:  "invalid_request",
		},
		{
			Name:           "Missing Redirect URI",
			ClientID:       clientID,
			RedirectURI:    "",
			ResponseType:   "code",
			Scope:          "openid",
			State:          "test_state_missing_redirect",
			ExpectedStatus: http.StatusFound,
		},
		{
			Name:           "Missing Response Type",
			ClientID:       clientID,
			RedirectURI:    redirectURI,
			ResponseType:   "",
			Scope:          "openid",
			State:          "test_state_missing_response",
			ExpectedStatus: http.StatusFound,
			ExpectedError:  "invalid_request",
		},
		{
			Name:           "Missing State Parameter",
			ClientID:       clientID,
			RedirectURI:    redirectURI,
			ResponseType:   "code",
			Scope:          "openid",
			State:          "",
			ExpectedStatus: http.StatusFound,
		},
	}

	for _, tc := range testCases {
		ts.Run(tc.Name, func() {
			resp, err := initiateAuthorizationFlow(tc.ClientID, tc.RedirectURI, tc.ResponseType, "openid", tc.State)
			ts.NoError(err, "Failed to initiate authorization flow")
			defer resp.Body.Close()

			ts.Equal(tc.ExpectedStatus, resp.StatusCode, "Expected status code")

			if tc.ExpectedStatus == http.StatusFound {
				location := resp.Header.Get("Location")
				ts.NotEmpty(location, "Expected redirect location header")
				if tc.ExpectedError != "" {
					err := validateOAuth2ErrorRedirect(location, tc.ExpectedError, "")
					ts.NoError(err, "OAuth2 error redirect validation failed")
				} else {
					sessionDataKey, _, err := extractSessionData(location)
					ts.NoError(err, "Failed to extract session data")
					ts.NotEmpty(sessionDataKey, "sessionDataKey should be present")
				}
			} else {
				bodyBytes, _ := io.ReadAll(resp.Body)
				ts.T().Logf("Error response body: %s", string(bodyBytes))
			}
		})
	}
}

// TestTokenRequestValidation tests the validation of token request parameters
func (ts *AuthzTestSuite) TestTokenRequestValidation() {
	// Create test user and get authorization code
	username := "token_test_user"
	password := "testpass123"

	user := testutils.User{
		OrganizationUnit: testOUID,
		Type:             "person",
		Attributes: json.RawMessage(fmt.Sprintf(`{
			"username": "%s",
			"password": "%s",
			"email": "%s@example.com",
			"firstName": "Test",
			"lastName": "User"
		}`, username, password, username)),
	}
	userID, err := testutils.CreateUser(user)
	ts.NoError(err, "Failed to create test user")
	defer func() {
		if err := testutils.DeleteUser(userID); err != nil {
			ts.T().Logf("Warning: Failed to delete test user: %v", err)
		}
	}()

	// Get a valid authorization code first
	validAuthzCode := initiateAuthorizeFlowAndRetrieveAuthzCode(ts, username, password)
	anotherValidAuthzCode := initiateAuthorizeFlowAndRetrieveAuthzCode(ts, username, password)

	testCases := []struct {
		Name           string
		ClientID       string
		ClientSecret   string
		Code           string
		RedirectURI    string
		GrantType      string
		ExpectedStatus int
		ExpectedError  string
	}{
		{
			Name:           "Missing Authorization Code",
			ClientID:       clientID,
			ClientSecret:   clientSecret,
			Code:           "",
			RedirectURI:    redirectURI,
			GrantType:      "authorization_code",
			ExpectedStatus: http.StatusBadRequest,
			ExpectedError:  "invalid_grant",
		},
		{
			Name:           "No Client ID",
			ClientID:       "",
			ClientSecret:   clientSecret,
			Code:           validAuthzCode,
			RedirectURI:    redirectURI,
			GrantType:      "authorization_code",
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedError:  "invalid_client",
		},
		{
			Name:           "No Client ID and Secret",
			ClientID:       "",
			ClientSecret:   "",
			Code:           validAuthzCode,
			RedirectURI:    redirectURI,
			GrantType:      "authorization_code",
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedError:  "invalid_client",
		},
		{
			Name:           "No Client Secret",
			ClientID:       clientID,
			ClientSecret:   "",
			Code:           validAuthzCode,
			RedirectURI:    redirectURI,
			GrantType:      "authorization_code",
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedError:  "invalid_client",
		},
		{
			Name:           "Invalid Client Credentials",
			ClientID:       clientID,
			ClientSecret:   "wrong_secret",
			Code:           validAuthzCode,
			RedirectURI:    redirectURI,
			GrantType:      "authorization_code",
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedError:  "invalid_client",
		},
		{
			Name:           "Missing Grant Type",
			ClientID:       clientID,
			ClientSecret:   clientSecret,
			Code:           validAuthzCode,
			RedirectURI:    redirectURI,
			GrantType:      "",
			ExpectedStatus: http.StatusBadRequest,
			ExpectedError:  "invalid_request",
		},
		{
			Name:           "Invalid Authorization Code",
			ClientID:       clientID,
			ClientSecret:   clientSecret,
			Code:           "invalid_code_12345",
			RedirectURI:    redirectURI,
			GrantType:      "authorization_code",
			ExpectedStatus: http.StatusBadRequest,
			ExpectedError:  "invalid_grant",
		},
		{
			Name:           "Invalid Client ID",
			ClientID:       "invalid_client_id",
			ClientSecret:   clientSecret,
			Code:           validAuthzCode,
			RedirectURI:    redirectURI,
			GrantType:      "authorization_code",
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedError:  "invalid_client",
		},
		{
			Name:           "Mismatched Redirect URI",
			ClientID:       clientID,
			ClientSecret:   clientSecret,
			Code:           anotherValidAuthzCode,
			RedirectURI:    "https://localhost:3001",
			GrantType:      "authorization_code",
			ExpectedStatus: http.StatusBadRequest,
			ExpectedError:  "invalid_grant",
		},
		{
			Name:           "Used unsuccessful Authz Code",
			ClientID:       clientID,
			ClientSecret:   clientSecret,
			Code:           anotherValidAuthzCode,
			RedirectURI:    redirectURI,
			GrantType:      "authorization_code",
			ExpectedStatus: http.StatusBadRequest,
			ExpectedError:  "invalid_grant",
		},
		{
			Name:           "Invalid Grant Type Format",
			ClientID:       clientID,
			ClientSecret:   clientSecret,
			Code:           validAuthzCode,
			RedirectURI:    redirectURI,
			GrantType:      "invalid_grant_type",
			ExpectedStatus: http.StatusBadRequest,
			ExpectedError:  "unsupported_grant_type",
		},
		{
			Name:           "Valid Token Request",
			ClientID:       clientID,
			ClientSecret:   clientSecret,
			Code:           validAuthzCode,
			RedirectURI:    redirectURI,
			GrantType:      "authorization_code",
			ExpectedStatus: http.StatusOK,
			ExpectedError:  "",
		},
		{
			Name:           "Used successful Authz Code",
			ClientID:       clientID,
			ClientSecret:   clientSecret,
			Code:           validAuthzCode,
			RedirectURI:    redirectURI,
			GrantType:      "authorization_code",
			ExpectedStatus: http.StatusBadRequest,
			ExpectedError:  "invalid_grant",
		},
	}

	for _, tc := range testCases {
		ts.Run(tc.Name, func() {
			result, err := requestToken(tc.ClientID, tc.ClientSecret, tc.Code, tc.RedirectURI, tc.GrantType)
			ts.NoError(err, "Token request should not error at transport level")
			ts.Equal(tc.ExpectedStatus, result.StatusCode, "Expected status code")

			if tc.ExpectedStatus == http.StatusOK {
				ts.NotNil(result.Token, "Token payload should be present on success")

				tokenResponse := result.Token
				ts.NotEmpty(tokenResponse.AccessToken, "Access token should be present")
				ts.Equal("Bearer", tokenResponse.TokenType, "Token type should be Bearer")
				ts.True(tokenResponse.ExpiresIn > 0, "Expires in should be greater than 0")

				parts := strings.Split(tokenResponse.AccessToken, ".")
				ts.Len(parts, 3, "Access token should be a JWT with 3 parts")

				payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
				ts.NoError(err, "Failed to decode JWT payload")

				var claims map[string]interface{}
				err = json.Unmarshal(payloadBytes, &claims)
				ts.NoError(err, "Failed to unmarshal JWT claims")

				ts.Equal(tc.ClientID, claims["aud"], "Audience claim should match client_id")
				ts.Equal("openid", claims["scope"], "Scope claim should match requested scope")
				ts.Equal(userID, claims["sub"], "Subject claim should match authenticated user ID")
			} else if tc.ExpectedError != "" {
				var errorResponse map[string]interface{}
				err := json.Unmarshal(result.Body, &errorResponse)
				ts.NoError(err, "Failed to unmarshal error response")
				ts.Contains(errorResponse, "error", "Error response should contain error field")
				ts.Equal(tc.ExpectedError, errorResponse["error"], "Expected error should match")
			}
		})
	}
}

func initiateAuthorizeFlowAndRetrieveAuthzCode(ts *AuthzTestSuite, username string, password string) string {
	resp, err := initiateAuthorizationFlow(clientID, redirectURI, "code", "openid", "token_test_state")
	ts.NoError(err, "Failed to initiate authorization flow")
	defer resp.Body.Close()

	ts.Equal(http.StatusFound, resp.StatusCode, "Expected redirect status")
	location := resp.Header.Get("Location")
	sessionDataKey, _, err := extractSessionData(location)
	ts.NoError(err, "Failed to extract session data")

	// Execute authentication flow
	flowStep, err := ExecuteAuthenticationFlow(ts.applicationID, map[string]string{
		"username": username,
		"password": password,
	})
	ts.NoError(err, "Failed to execute authentication flow")
	ts.Equal("COMPLETE", flowStep.FlowStatus, "Flow should complete successfully")

	// Complete authorization
	authzResponse, err := completeAuthorization(sessionDataKey, flowStep.Assertion)
	ts.NoError(err, "Failed to complete authorization")
	validAuthzCode, err := extractAuthorizationCode(authzResponse.RedirectURI)
	ts.NoError(err, "Failed to extract authorization code")
	return validAuthzCode
}

// TestRedirectURIValidation tests the redirect URI validation in OAuth2 flows
func (ts *AuthzTestSuite) TestRedirectURIValidation() {
	testCases := []struct {
		Name           string
		ClientID       string
		RedirectURI    string
		ResponseType   string
		Scope          string
		State          string
		ExpectedStatus int
		ExpectedError  string
		Description    string
	}{
		{
			Name:           "Valid HTTPS Redirect URI",
			ClientID:       clientID,
			RedirectURI:    redirectURI,
			ResponseType:   "code",
			Scope:          "openid",
			State:          "redirect_test_valid_https",
			ExpectedStatus: http.StatusFound,
			Description:    "Standard HTTPS localhost should be valid",
		},
		{
			Name:           "Valid HTTPS with Path",
			ClientID:       clientID,
			RedirectURI:    "https://localhost:3000/callback",
			ResponseType:   "code",
			Scope:          "openid",
			State:          "redirect_test_valid_path",
			ExpectedStatus: http.StatusFound,
			ExpectedError:  "invalid_request",
			Description:    "HTTPS with callback path should be rejected (not registered)",
		},
		{
			Name:           "HTTP Redirect URI",
			ClientID:       clientID,
			RedirectURI:    "http://localhost:3000",
			ResponseType:   "code",
			Scope:          "openid",
			State:          "redirect_test_http",
			ExpectedStatus: http.StatusFound,
			ExpectedError:  "invalid_request",
			Description:    "HTTP should be rejected for security",
		},
		{
			Name:           "Invalid Protocol",
			ClientID:       clientID,
			RedirectURI:    "invalid://localhost:3000",
			ResponseType:   "code",
			Scope:          "openid",
			State:          "redirect_test_invalid_protocol",
			ExpectedStatus: http.StatusFound,
			ExpectedError:  "invalid_request",
			Description:    "Invalid protocol should be rejected",
		},
		{
			Name:           "External Domain",
			ClientID:       clientID,
			RedirectURI:    "https://malicious.com/callback",
			ResponseType:   "code",
			Scope:          "openid",
			State:          "redirect_test_malicious_domain",
			ExpectedStatus: http.StatusFound,
			ExpectedError:  "invalid_request",
			Description:    "External malicious domain should be rejected",
		},
	}

	for _, tc := range testCases {
		ts.Run(tc.Name, func() {
			resp, err := initiateAuthorizationFlow(tc.ClientID, tc.RedirectURI, tc.ResponseType, tc.Scope, tc.State)
			ts.NoError(err, "Failed to initiate authorization flow")
			defer resp.Body.Close()

			ts.Equal(tc.ExpectedStatus, resp.StatusCode, "Expected status code")

			if tc.ExpectedStatus == http.StatusFound {
				location := resp.Header.Get("Location")
				ts.NotEmpty(location, "Expected redirect location header")

				if tc.ExpectedError != "" {
					if tc.RedirectURI != redirectURI {
						parsedLocation, parseErr := url.Parse(location)
						ts.NoError(parseErr, "Failed to parse redirect location")

						parsedTestURI, parseErr := url.Parse(tc.RedirectURI)
						ts.NoError(parseErr, "Failed to parse test case redirect URI")

						ts.NotEqual(parsedTestURI.Host, parsedLocation.Host,
							"System redirected to invalid domain '%s' instead of authorization server",
							parsedTestURI.Host)
					}

					err := validateOAuth2ErrorRedirect(location, tc.ExpectedError, "")
					ts.NoError(err, "OAuth2 error redirect validation failed")

				} else {
					sessionDataKey, _, err := extractSessionData(location)
					ts.NoError(err, "Failed to extract session data")
					ts.NotEmpty(sessionDataKey, "sessionDataKey should be present")
				}
			}
		})
	}
}

func (ts *AuthzTestSuite) TestCompleteAuthorizationCodeFlow() {
	testCases := []TestCase{
		{
			Name:         "Successful Flow",
			ClientID:     clientID,
			RedirectURI:  "https://localhost:3000",
			ResponseType: "code",
			Scope:        "openid",
			State:        "test_state_456",
			Username:     "testuser",
			Password:     "testpass123",
		},
	}

	for _, tc := range testCases {
		ts.Run(tc.Name, func() {
			// Create test user with credentials
			user := testutils.User{
				OrganizationUnit: testOUID,
				Type:             "person",
				Attributes: json.RawMessage(fmt.Sprintf(`{
					"username": "%s",
					"password": "%s",
					"email": "%s@example.com",
					"firstName": "Test",
					"lastName": "User"
				}`, tc.Username, tc.Password, tc.Username)),
			}
			userID, err := testutils.CreateUser(user)
			if err != nil {
				ts.T().Fatalf("Failed to create test user: %v", err)
			}
			if userID == "" {
				ts.T().Fatalf("Expected user ID, got empty string")
			}

			defer func() {
				if err := testutils.DeleteUser(userID); err != nil {
					ts.T().Logf("Warning: Failed to delete test user: %v", err)
				}
			}()

			// Start authorization flow
			resp, err := initiateAuthorizationFlow(tc.ClientID, tc.RedirectURI, tc.ResponseType, tc.Scope, tc.State)
			ts.NoError(err, "Failed to initiate authorization flow")
			defer resp.Body.Close()

			ts.Equal(http.StatusFound, resp.StatusCode, "Expected redirect status")

			location := resp.Header.Get("Location")
			ts.NotEmpty(location, "Expected redirect location header")

			// Extract session data
			sessionDataKey, _, err := extractSessionData(location)
			if err != nil {
				ts.T().Fatalf("Failed to extract session data: %v", err)
			}
			if sessionDataKey == "" {
				ts.T().Fatalf("Expected sessionDataKey, got empty string")
			}

			// Execute authentication flow
			flowStep, err := ExecuteAuthenticationFlow(ts.applicationID, map[string]string{
				"username": tc.Username,
				"password": tc.Password,
			})
			if err != nil {
				ts.T().Fatalf("Failed to execute authentication flow: %v", err)
			}
			if flowStep == nil {
				ts.T().Fatalf("Expected flow step, got nil")
			}

			if flowStep.FlowID == "" {
				ts.T().Fatalf("Expected flow ID, got empty string")
			}
			if flowStep.FlowStatus != "COMPLETE" {
				ts.T().Fatalf("Expected flow status COMPLETE, got %s", flowStep.FlowStatus)
			}

			if flowStep.Assertion == "" {
				ts.T().Fatalf("Expected assertion, got empty string")
			}

			// Complete authorization
			authzResponse, err := completeAuthorization(sessionDataKey, flowStep.Assertion)
			ts.NoError(err, "Failed to complete authorization")
			ts.NotEmpty(authzResponse.RedirectURI, "Redirect URI should be present")

			authzCode, err := extractAuthorizationCode(authzResponse.RedirectURI)
			ts.NoError(err, "Failed to extract authorization code")
			ts.NotEmpty(authzCode, "Authorization code should be present")

			// Exchange authorization code for access token
			result, err := requestToken(clientID, clientSecret, authzCode, tc.RedirectURI, "authorization_code")
			ts.NoError(err, "Failed to exchange code for token")
			ts.Equal(http.StatusOK, result.StatusCode, "Token request should succeed")
			tokenResponse := result.Token

			// Verify token response
			ts.NotEmpty(tokenResponse.AccessToken, "Access token should be present")
			ts.Equal("Bearer", tokenResponse.TokenType, "Token type should be Bearer")
			ts.True(tokenResponse.ExpiresIn > 0, "Expires in should be greater than 0")

			parts := strings.Split(tokenResponse.AccessToken, ".")
			ts.Len(parts, 3, "Access token should be a JWT with 3 parts")

			payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
			ts.NoError(err, "Failed to decode JWT payload")

			var claims map[string]interface{}
			err = json.Unmarshal(payloadBytes, &claims)
			ts.NoError(err, "Failed to unmarshal JWT claims")

			ts.Equal(tc.ClientID, claims["aud"], "Audience claim should match client_id")
			ts.Equal(tc.Scope, claims["scope"], "Scope claim should match requested scope")
			ts.Equal(userID, claims["sub"], "Subject claim should match authenticated user ID")
		})
	}
}

func (ts *AuthzTestSuite) TestAuthorizationCodeErrorScenarios() {
	testCases := []TestCase{
		{
			Name:           "Reused Authorization Code",
			ClientID:       clientID,
			RedirectURI:    "https://localhost:3000",
			ResponseType:   "code",
			Scope:          "openid",
			State:          "test_state_error",
			Username:       "testuser_error",
			Password:       "testpass123",
			ExpectedStatus: http.StatusBadRequest,
			ExpectedError:  "invalid_grant",
		}}

	for _, tc := range testCases {
		ts.Run(tc.Name, func() {
			// Create test user
			user := testutils.User{
				OrganizationUnit: testOUID,
				Type:             "person",
				Attributes: json.RawMessage(fmt.Sprintf(`{
					"username": "%s",
					"password": "%s",
					"email": "%s@example.com",
					"firstName": "Test",
					"lastName": "User"
				}`, tc.Username, tc.Password, tc.Username)),
			}
			userID, err := testutils.CreateUser(user)
			ts.NoError(err, "Failed to create test user")
			defer func() {
				if err := testutils.DeleteUser(userID); err != nil {
					ts.T().Logf("Warning: Failed to delete test user: %v", err)
				}
			}()

			// Start authorization flow
			resp, err := initiateAuthorizationFlow(tc.ClientID, tc.RedirectURI, tc.ResponseType, tc.Scope, tc.State)
			ts.NoError(err, "Failed to initiate authorization flow")
			defer resp.Body.Close()

			ts.Equal(http.StatusFound, resp.StatusCode, "Expected redirect status")

			location := resp.Header.Get("Location")
			ts.NotEmpty(location, "Expected redirect location header")

			sessionDataKey, _, err := extractSessionData(location)
			ts.NoError(err, "Failed to extract session data")

			// Execute authentication flow
			flowStep, err := ExecuteAuthenticationFlow(ts.applicationID, map[string]string{
				"username": tc.Username,
				"password": tc.Password,
			})
			if err != nil {
				ts.T().Fatalf("Failed to execute authentication flow: %v", err)
			}
			if flowStep.FlowStatus != "COMPLETE" {
				ts.T().Fatalf("Expected flow status COMPLETE, got %s", flowStep.FlowStatus)
			}

			authzResponse, err := completeAuthorization(sessionDataKey, flowStep.Assertion)
			if err != nil {
				ts.T().Fatalf("Failed to complete authorization: %v", err)
			}

			// Extract authorization code
			authzCode, err := extractAuthorizationCode(authzResponse.RedirectURI)
			ts.NoError(err, "Failed to extract authorization code")

			if tc.Name == "Reused Authorization Code" {
				result, err := requestToken(clientID, clientSecret, authzCode, tc.RedirectURI, "authorization_code")
				ts.NoError(err, "First token exchange should succeed")
				ts.Equal(http.StatusOK, result.StatusCode, "First token exchange should succeed")

				// Second attempt should fail with invalid_grant
				result2, err := requestToken(clientID, clientSecret, authzCode, tc.RedirectURI, "authorization_code")
				ts.NoError(err, "Second token exchange should not error at transport level")
				ts.Equal(http.StatusBadRequest, result2.StatusCode, "Second token exchange should fail with bad request")

				// Check error response
				var errorResponse map[string]interface{}
				err = json.Unmarshal(result2.Body, &errorResponse)
				ts.NoError(err, "Failed to unmarshal error response")
				ts.Equal("invalid_grant", errorResponse["error"], "Expected invalid_grant error")
			}
		})
	}
}

// TestAuthorizationCodeFlowWithResourceParameter tests RFC 8707 resource parameter implementation
func (ts *AuthzTestSuite) TestAuthorizationCodeFlowWithResourceParameter() {
	// Test that resource parameter is properly stored and used as audience in access token
	resourceURL := "https://mcp.example.com/mcp"

	// Create test user
	user := testutils.User{
		OrganizationUnit: testOUID,
		Type:             "person",
		Attributes: json.RawMessage(`{
			"username": "resourcetest",
			"password": "testpass123",
			"email": "resourcetest@example.com",
			"firstName": "Resource",
			"lastName": "Test"
		}`),
	}
	userID, err := testutils.CreateUser(user)
	ts.NoError(err, "Failed to create test user")
	defer func() {
		if err := testutils.DeleteUser(userID); err != nil {
			ts.T().Logf("Warning: Failed to delete test user: %v", err)
		}
	}()

	// Start authorization flow with resource parameter
	resp, err := initiateAuthorizationFlowWithResource(
		clientID,
		redirectURI,
		"code",
		"openid",
		"test_resource_state",
		resourceURL,
	)
	ts.NoError(err, "Failed to initiate authorization flow with resource")
	defer resp.Body.Close()

	ts.Equal(http.StatusFound, resp.StatusCode, "Expected redirect status")
	location := resp.Header.Get("Location")
	ts.NotEmpty(location, "Expected redirect location header")

	sessionDataKey, _, err := extractSessionData(location)
	ts.NoError(err, "Failed to extract session data")

	// Execute authentication flow
	flowStep, err := ExecuteAuthenticationFlow(ts.applicationID, map[string]string{
		"username": "resourcetest",
		"password": "testpass123",
	})
	ts.NoError(err, "Failed to execute authentication flow")
	ts.Equal("COMPLETE", flowStep.FlowStatus, "Expected flow status COMPLETE")

	// Complete authorization
	authzResponse, err := completeAuthorization(sessionDataKey, flowStep.Assertion)
	ts.NoError(err, "Failed to complete authorization")

	// Extract authorization code
	authzCode, err := extractAuthorizationCode(authzResponse.RedirectURI)
	ts.NoError(err, "Failed to extract authorization code")

	// Request token with resource parameter
	tokenReq := url.Values{}
	tokenReq.Set("grant_type", "authorization_code")
	tokenReq.Set("code", authzCode)
	tokenReq.Set("redirect_uri", redirectURI)
	tokenReq.Set("resource", resourceURL)

	req, err := http.NewRequest("POST", testServerURL+"/oauth2/token", bytes.NewBufferString(tokenReq.Encode()))
	ts.NoError(err, "Failed to create token request")

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientID, clientSecret)

	tokenResp, err := ts.client.Do(req)
	ts.NoError(err, "Failed to send token request")
	defer tokenResp.Body.Close()

	ts.Equal(http.StatusOK, tokenResp.StatusCode, "Token request should succeed")

	var tokenResponse map[string]interface{}
	err = json.NewDecoder(tokenResp.Body).Decode(&tokenResponse)
	ts.NoError(err, "Failed to decode token response")

	// Extract and decode the access token
	accessToken, ok := tokenResponse["access_token"].(string)
	ts.True(ok, "Access token should be present")

	// Decode JWT to verify audience claim
	parts := strings.Split(accessToken, ".")
	ts.Len(parts, 3, "Access token should be a JWT")

	// Decode the payload (second part)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	ts.NoError(err, "Failed to decode JWT payload")

	var claims map[string]interface{}
	err = json.Unmarshal(payload, &claims)
	ts.NoError(err, "Failed to unmarshal JWT claims")

	// Verify the audience claim matches the resource parameter
	aud, ok := claims["aud"]
	ts.True(ok, "Audience claim should be present in access token")
	ts.Equal(resourceURL, aud, "Audience should match the resource parameter")
}
