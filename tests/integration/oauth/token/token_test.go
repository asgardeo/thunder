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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

const (
	testServerURL = "https://localhost:8095"
	clientId      = "token_test_client_123"
	clientSecret  = "token_test_secret_123"
	appName       = "TokenTestApp"
)

type TokenTestSuite struct {
	suite.Suite
	applicationIDBasic string
	applicationIDPost  string
	client             *http.Client
}

func TestTokenTestSuite(t *testing.T) {
	suite.Run(t, new(TokenTestSuite))
}

func (ts *TokenTestSuite) SetupSuite() {
	// Create a client that skips TLS verification
	ts.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// Create applications for different authentication methods
	ts.applicationIDBasic = ts.createTestApplication("client_secret_basic")
	ts.applicationIDPost = ts.createTestApplication("client_secret_post")
}

func (ts *TokenTestSuite) createTestApplication(authMethod string) string {
	// Create a unique application name and client ID based on auth method
	appName := appName + "_" + authMethod
	clientId := clientId + "_" + authMethod

	// Create a new application for testing
	app := map[string]interface{}{
		"name":                         appName,
		"description":                  "Application for token integration tests",
		"auth_flow_graph_id":           "auth_flow_config_basic",
		"registration_flow_graph_id":   "registration_flow_config_basic",
		"is_registration_flow_enabled": true,
		"inbound_auth_config": []map[string]interface{}{
			{
				"type": "oauth2",
				"config": map[string]interface{}{
					"client_id":     clientId,
					"client_secret": clientSecret,
					"redirect_uris": []string{"https://localhost:3000"},
					"grant_types": []string{
						"client_credentials",
						"authorization_code",
						"refresh_token",
					},
					"token_endpoint_auth_method": authMethod,
				},
			},
		},
	}

	jsonData, err := json.Marshal(app)
	if err != nil {
		ts.T().Fatalf("Failed to marshal application data: %v", err)
	}

	// Send the request to create the application
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

	// Parse the response to get the application ID
	var respData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		ts.T().Fatalf("Failed to parse response: %v", err)
	}

	appId := respData["id"].(string)
	ts.T().Logf("Created test application with ID: %s", appId)
	return appId
}

func (ts *TokenTestSuite) TearDownSuite() {
	// Clean up both applications
	if ts.applicationIDBasic != "" {
		ts.deleteApplication(ts.applicationIDBasic)
	}
	if ts.applicationIDPost != "" {
		ts.deleteApplication(ts.applicationIDPost)
	}
}

func (ts *TokenTestSuite) deleteApplication(appId string) {
	// Delete the application
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/applications/%s", testServerURL, appId), nil)
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
		ts.T().Logf("Successfully deleted test application with ID: %s", appId)
	}
}

func (ts *TokenTestSuite) runClientCredentialsTestCase(request *http.Request,
	expectedStatus int, expectedScopes []string, expectedError string) {

	// Send the request using the suite's client
	resp, err := ts.client.Do(request)
	if err != nil {
		ts.T().Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Validate the response status.
	if resp.StatusCode != expectedStatus {
		ts.T().Fatalf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
	}

	// Parse the response body.
	var respBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		ts.T().Fatalf("Failed to parse response body: %v", err)
	}

	// Validate the response content.
	if expectedStatus == http.StatusOK {
		if _, ok := respBody["access_token"]; !ok {
			ts.T().Fatalf("Response does not contain access_token")
		}
		if _, ok := respBody["token_type"]; !ok {
			ts.T().Fatalf("Response does not contain token_type")
		}
		if _, ok := respBody["expires_in"]; !ok {
			ts.T().Fatalf("Response does not contain expires_in")
		}
		if len(expectedScopes) > 0 {
			if _, ok := respBody["scope"]; !ok {
				ts.T().Fatalf("Response does not contain scope")
			}
			scopes := strings.Fields(respBody["scope"].(string))
			if len(scopes) != len(expectedScopes) {
				ts.T().Fatalf("Expected %d scopes, got %d", len(expectedScopes), len(scopes))
			}
			for _, expectedScope := range expectedScopes {
				found := false
				for _, scope := range scopes {
					if scope == expectedScope {
						found = true
						break
					}
				}
				if !found {
					ts.T().Fatalf("Expected scope %s not found in response", expectedScope)
				}
			}
		} else if _, ok := respBody["scope"]; ok {
			ts.T().Fatalf("Response should not contain scope when no scopes are requested")
		}
	} else if expectedStatus == http.StatusBadRequest {
		if _, ok := respBody["error"]; !ok {
			ts.T().Fatalf("Response does not contain error")
		}
		if respBody["error"] != expectedError {
			ts.T().Fatalf("Expected error '%s', got '%v'", expectedError, respBody["error"])
		}
	}
}

func (ts *TokenTestSuite) TestClientCredentialsGrantWithHeaderCredentials() {

	testCases := []struct {
		testName        string
		requestedScopes string
		expectedStatus  int
		expectedScopes  []string
	}{
		{
			testName:        "WithAuthorizedScopes",
			requestedScopes: "internal_user_mgt_view internal_user_mgt_edit internal_group_mgt_view",
			expectedStatus:  http.StatusOK,
			expectedScopes:  []string{"internal_user_mgt_view", "internal_user_mgt_edit", "internal_group_mgt_view"},
		},
		{
			testName:        "WithoutScopes",
			requestedScopes: "",
			expectedStatus:  http.StatusOK,
			expectedScopes:  nil,
		},
		{
			testName:        "WithUnknownScopes",
			requestedScopes: "unknown_scope",
			expectedStatus:  http.StatusOK,
			expectedScopes:  []string{"unknown_scope"},
		},
		{
			testName:        "WithAuthorizedAndUnknownScopes",
			requestedScopes: "internal_user_mgt_view unknown_scope",
			expectedStatus:  http.StatusOK,
			expectedScopes:  []string{"internal_user_mgt_view", "unknown_scope"},
		},
	}

	for _, tc := range testCases {
		ts.Run(tc.testName, func() {
			// Prepare the request.
			reqBody := strings.NewReader("grant_type=client_credentials&scope=" + tc.requestedScopes)
			request, err := http.NewRequest("POST", testServerURL+"/oauth2/token", reqBody)
			if err != nil {
				ts.T().Fatalf("Failed to create request: %v", err)
			}
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			request.SetBasicAuth(clientId+"_client_secret_basic", clientSecret)

			// Run the test.
			ts.runClientCredentialsTestCase(request, tc.expectedStatus, tc.expectedScopes, "")
		})
	}
}

func (ts *TokenTestSuite) TestClientCredentialsGrantWithBodyCredentials() {

	testCases := []struct {
		testName        string
		requestedScopes string
		expectedStatus  int
		expectedScopes  []string
	}{
		{
			testName:        "WithAuthorizedScopes",
			requestedScopes: "internal_user_mgt_view internal_user_mgt_edit",
			expectedStatus:  http.StatusOK,
			expectedScopes:  []string{"internal_user_mgt_view", "internal_user_mgt_edit"},
		},
		{
			testName:        "WithoutScopes",
			requestedScopes: "",
			expectedStatus:  http.StatusOK,
			expectedScopes:  nil,
		},
	}

	for _, tc := range testCases {
		ts.Run(tc.testName, func() {
			reqBody := strings.NewReader("grant_type=client_credentials&scope=" + tc.requestedScopes +
				"&client_id=" + clientId + "_client_secret_post" + "&client_secret=" + clientSecret)
			request, err := http.NewRequest("POST", testServerURL+"/oauth2/token", reqBody)
			if err != nil {
				ts.T().Fatalf("Failed to create request: %v", err)
			}
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			ts.runClientCredentialsTestCase(request, tc.expectedStatus, tc.expectedScopes, "")
		})
	}
}

func (ts *TokenTestSuite) TestClientCredentialsGrantNegativeCases() {

	testCases := []struct {
		testName       string
		requestBody    string
		authHeader     string
		expectedStatus int
		expectedError  string
	}{
		{
			testName:       "InvalidHeaderCredentials",
			requestBody:    "grant_type=client_credentials",
			authHeader:     "Basic " + basicAuth("invalid", "invalid"),
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid_client",
		},
		{
			testName:       "IncorrectHeaderCredentials",
			requestBody:    "grant_type=client_credentials",
			authHeader:     "Basic invalid_base64",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid_client",
		},
		{
			testName:       "InvalidHeaderCredentials",
			requestBody:    "grant_type=client_credentials",
			authHeader:     "Basic invalid_base64",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid_client",
		},
		{
			testName:       "InvalidCredentialsInBody",
			requestBody:    "grant_type=client_credentials&client_id=invalid&client_secret=invalid",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid_client",
		},
		{
			testName:       "MissingCredentialsInBody",
			requestBody:    "grant_type=client_credentials",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid_client",
		},
		{
			testName:       "InvalidGrantType",
			requestBody:    "grant_type=invalid_grant",
			authHeader:     "Basic " + basicAuth(clientId+"_client_secret_basic", clientSecret),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "unsupported_grant_type",
		},
		{
			testName:       "MissingGrantType",
			requestBody:    "",
			authHeader:     "Basic " + basicAuth(clientId+"_client_secret_basic", clientSecret),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid_request",
		},
	}

	for _, tc := range testCases {
		ts.Run(tc.testName, func() {
			// Prepare the request.
			reqBody := strings.NewReader(tc.requestBody)
			request, err := http.NewRequest("POST", testServerURL+"/oauth2/token", reqBody)
			if err != nil {
				ts.T().Fatalf("Failed to create request: %v", err)
			}
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			if tc.authHeader != "" {
				request.Header.Set("Authorization", tc.authHeader)
			}

			// Run the test.
			ts.runClientCredentialsTestCase(request, tc.expectedStatus, nil, tc.expectedError)
		})
	}
}

func basicAuth(username, password string) string {

	return base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
}
