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
 * software distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package userinfo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/asgardeo/thunder/tests/integration/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	testServerURL = "https://localhost:8095"
	clientID      = "userinfo_test_client_123"
	clientSecret  = "userinfo_test_secret_123"
	appName       = "UserInfoTestApp"
	redirectURI   = "https://localhost:3000"
)

var (
	testUserSchema = testutils.UserSchema{
		Name: "userinfo-person",
		Schema: map[string]interface{}{
			"username": map[string]interface{}{
				"type": "string",
			},
			"password": map[string]interface{}{
				"type": "string",
			},
			"email": map[string]interface{}{
				"type": "string",
			},
			"firstName": map[string]interface{}{
				"type": "string",
			},
			"lastName": map[string]interface{}{
				"type": "string",
			},
		},
	}
)

type UserInfoTestSuite struct {
	suite.Suite
	flowID        string
	applicationID string
	userSchemaID  string
	userID        string
	client        *http.Client
	ouID          string
}

func TestUserInfoTestSuite(t *testing.T) {
	suite.Run(t, new(UserInfoTestSuite))
}

func (ts *UserInfoTestSuite) SetupSuite() {
	ts.client = testutils.GetHTTPClient()

	// Create test organization unit
	ou := testutils.OrganizationUnit{
		Handle:      "userinfo-test-ou",
		Name:        "UserInfo Test OU",
		Description: "Organization unit for UserInfo integration testing",
		Parent:      nil,
	}
	ouID, err := testutils.CreateOrganizationUnit(ou)
	ts.Require().NoError(err, "Failed to create test organization unit")
	ts.ouID = ouID

	// Create user schema
	testUserSchema.OrganizationUnitId = ts.ouID
	schemaID, err := testutils.CreateUserType(testUserSchema)
	ts.Require().NoError(err, "Failed to create test user schema")
	ts.userSchemaID = schemaID

	// Create test user
	ts.userID = ts.createTestUser()

	// Create authentication flow
	ts.flowID = ts.createTestAuthenticationFlow()

	// Create OAuth application
	ts.applicationID = ts.createTestApplication(ts.flowID)
}

func (ts *UserInfoTestSuite) TearDownSuite() {
	// Clean up application
	if ts.applicationID != "" {
		ts.deleteApplication(ts.applicationID)
	}

	// Clean up authentication flow
	if ts.flowID != "" {
		if err := testutils.DeleteFlow(ts.flowID); err != nil {
			ts.T().Logf("Failed to delete authentication flow during teardown: %v", err)
		}
	}

	// Clean up user
	if ts.userID != "" {
		testutils.DeleteUser(ts.userID)
	}

	// Clean up organization unit
	if ts.ouID != "" {
		testutils.DeleteOrganizationUnit(ts.ouID)
	}

	// Clean up user schema
	if ts.userSchemaID != "" {
		if err := testutils.DeleteUserType(ts.userSchemaID); err != nil {
			ts.T().Logf("Failed to delete user schema during teardown: %v", err)
		}
	}
}

func (ts *UserInfoTestSuite) createTestUser() string {
	attributes := map[string]interface{}{
		"username":  "userinfo_test_user",
		"password":  "SecurePass123!",
		"email":     "userinfo_test@example.com",
		"firstName": "UserInfo",
		"lastName":  "Test",
	}

	attributesJSON, err := json.Marshal(attributes)
	ts.Require().NoError(err, "Failed to marshal user attributes")

	user := testutils.User{
		Type:             "userinfo-person",
		OrganizationUnit: ts.ouID,
		Attributes:       json.RawMessage(attributesJSON),
	}

	userID, err := testutils.CreateUser(user)
	ts.Require().NoError(err, "Failed to create test user")
	ts.T().Logf("Created test user with ID: %s", userID)

	return userID
}

func (ts *UserInfoTestSuite) createTestAuthenticationFlow() string {
	flow := testutils.Flow{
		Name:     "Test Auth Flow",
		FlowType: "AUTHENTICATION",
		Handle:   "test_auth_flow",
		Nodes: []map[string]interface{}{
			{
				"id":        "start",
				"type":      "START",
				"onSuccess": "prompt_credentials",
			},
			{
				"id":   "prompt_credentials",
				"type": "PROMPT",
				"prompts": []map[string]interface{}{
					{
						"inputs": []map[string]interface{}{
							{
								"ref":        "input_001",
								"identifier": "username",
								"type":       "TEXT_INPUT",
								"required":   true,
							},
							{
								"ref":        "input_002",
								"identifier": "password",
								"type":       "PASSWORD_INPUT",
								"required":   true,
							},
						},
						"action": map[string]interface{}{
							"ref":      "action_001",
							"nextNode": "basic_auth",
						},
					},
				},
			},
			{
				"id":   "basic_auth",
				"type": "TASK_EXECUTION",
				"executor": map[string]interface{}{
					"name": "BasicAuthExecutor",
					"inputs": []map[string]interface{}{
						{
							"ref":        "input_001",
							"identifier": "username",
							"type":       "TEXT_INPUT",
							"required":   true,
						},
						{
							"ref":        "input_002",
							"identifier": "password",
							"type":       "PASSWORD_INPUT",
							"required":   true,
						},
					},
				},
				"onSuccess": "auth_assert",
			},
			{
				"id":   "auth_assert",
				"type": "TASK_EXECUTION",
				"executor": map[string]interface{}{
					"name": "AuthAssertExecutor",
				},
				"onSuccess": "end",
			},
			{
				"id":   "end",
				"type": "END",
			},
		},
	}

	flowID, err := testutils.CreateFlow(flow)
	ts.Require().NoError(err, "Failed to create test authentication flow")
	ts.T().Logf("Created test authentication flow with ID: %s", flowID)

	return flowID
}

func (ts *UserInfoTestSuite) createTestApplication(authFlowID string) string {
	app := map[string]interface{}{
		"name":                         appName,
		"description":                  "Application for UserInfo integration tests",
		"auth_flow_id":                 authFlowID,
		"is_registration_flow_enabled": false,
		"allowed_user_types":           []string{"userinfo-person"},
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
						"urn:ietf:params:oauth:grant-type:token-exchange",
					},
					"response_types":             []string{"code"},
					"token_endpoint_auth_method": "client_secret_basic",
					"scopes":                     []string{"openid", "profile", "email"},
					"token": map[string]interface{}{
						"id_token": map[string]interface{}{
							"user_attributes": []string{"email", "firstName", "lastName", "name"},
						},
					},
					"scope_claims": map[string][]string{
						"profile": {"firstName", "lastName", "name"},
						"email":   {"email"},
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(app)
	ts.Require().NoError(err, "Failed to marshal application data")

	req, err := http.NewRequest("POST", testServerURL+"/applications", bytes.NewBuffer(jsonData))
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

func (ts *UserInfoTestSuite) deleteApplication(appID string) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/applications/%s", testServerURL, appID), nil)
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

// getClientCredentialsToken gets an access token using client_credentials grant
func (ts *UserInfoTestSuite) getClientCredentialsToken(scope string) (string, error) {
	reqBody := strings.NewReader(fmt.Sprintf("grant_type=client_credentials&scope=%s", scope))
	req, err := http.NewRequest("POST", testServerURL+"/oauth2/token", reqBody)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientID, clientSecret)

	resp, err := ts.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var tokenResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	accessToken, ok := tokenResp["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("access_token not found in response")
	}

	return accessToken, nil
}

// getAuthorizationCodeToken gets an access token using authorization_code grant
func (ts *UserInfoTestSuite) getAuthorizationCodeToken(scope string) (string, error) {
	// Step 1: Initiate authorization flow
	authzResp, err := testutils.InitiateAuthorizationFlow(clientID, redirectURI, "code", scope, "test_state")
	if err != nil {
		return "", fmt.Errorf("failed to initiate authorization: %w", err)
	}
	defer authzResp.Body.Close()

	location := authzResp.Header.Get("Location")
	if location == "" {
		return "", fmt.Errorf("no Location header in authorization response")
	}

	// Step 2: Extract auth ID and flow ID
	authId, flowID, err := testutils.ExtractAuthData(location)
	if err != nil {
		return "", fmt.Errorf("failed to extract auth ID: %w", err)
	}

	// Step 3: Initiate authentication flow
	_, err = testutils.ExecuteAuthenticationFlow(flowID, nil, "")
	if err != nil {
		return "", fmt.Errorf("failed to initiate authentication flow: %w", err)
	}

	// Step 4: Execute authentication flow
	authInputs := map[string]string{
		"username": "userinfo_test_user",
		"password": "SecurePass123!",
	}
	flowStep, err := testutils.ExecuteAuthenticationFlow(flowID, authInputs, "action_001")
	if err != nil {
		return "", fmt.Errorf("failed to execute authentication flow: %w", err)
	}

	if flowStep.Assertion == "" {
		return "", fmt.Errorf("assertion not found in flow step")
	}

	// Step 5: Complete authorization
	authzResponse, err := testutils.CompleteAuthorization(authId, flowStep.Assertion)
	if err != nil {
		return "", fmt.Errorf("failed to complete authorization: %w", err)
	}

	// Step 6: Extract authorization code
	code, err := testutils.ExtractAuthorizationCode(authzResponse.RedirectURI)
	if err != nil {
		return "", fmt.Errorf("failed to extract authorization code: %w", err)
	}

	// Step 7: Exchange code for token
	tokenResult, err := testutils.RequestToken(clientID, clientSecret, code, redirectURI, "authorization_code")
	if err != nil {
		return "", fmt.Errorf("failed to request token: %w", err)
	}

	if tokenResult.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed with status %d", tokenResult.StatusCode)
	}

	if tokenResult.Token == nil || tokenResult.Token.AccessToken == "" {
		return "", fmt.Errorf("access token not found in response")
	}

	return tokenResult.Token.AccessToken, nil
}

// getRefreshToken gets a refresh token and then uses it to get a new access token
func (ts *UserInfoTestSuite) getRefreshToken(scope string) (string, error) {
	// First get an access token with refresh token using authorization_code grant
	// Step 1: Initiate authorization flow
	authzResp, err := testutils.InitiateAuthorizationFlow(clientID, redirectURI, "code", scope, "test_state")
	if err != nil {
		return "", fmt.Errorf("failed to initiate authorization: %w", err)
	}
	defer authzResp.Body.Close()

	location := authzResp.Header.Get("Location")
	if location == "" {
		return "", fmt.Errorf("no Location header in authorization response")
	}

	// Step 2: Extract auth ID and flow ID
	authId, flowID, err := testutils.ExtractAuthData(location)
	if err != nil {
		return "", fmt.Errorf("failed to extract auth ID: %w", err)
	}

	// Step 3: Initiate authentication flow
	_, err = testutils.ExecuteAuthenticationFlow(flowID, map[string]string{}, "")
	if err != nil {
		return "", fmt.Errorf("failed to initiate authentication flow: %w", err)
	}

	// Step 4: Execute authentication flow
	authInputs := map[string]string{
		"username": "userinfo_test_user",
		"password": "SecurePass123!",
	}
	flowStep, err := testutils.ExecuteAuthenticationFlow(flowID, authInputs, "action_001")
	if err != nil {
		return "", fmt.Errorf("failed to execute authentication flow: %w", err)
	}

	if flowStep.Assertion == "" {
		return "", fmt.Errorf("assertion not found in flow step")
	}

	// Step 5: Complete authorization
	authzResponse, err := testutils.CompleteAuthorization(authId, flowStep.Assertion)
	if err != nil {
		return "", fmt.Errorf("failed to complete authorization: %w", err)
	}

	// Step 6: Extract authorization code
	code, err := testutils.ExtractAuthorizationCode(authzResponse.RedirectURI)
	if err != nil {
		return "", fmt.Errorf("failed to extract authorization code: %w", err)
	}

	// Step 7: Exchange code for token (this should include refresh_token)
	tokenResult, err := testutils.RequestToken(clientID, clientSecret, code, redirectURI, "authorization_code")
	if err != nil {
		return "", fmt.Errorf("failed to request token: %w", err)
	}

	if tokenResult.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed with status %d", tokenResult.StatusCode)
	}

	if tokenResult.Token == nil || tokenResult.Token.RefreshToken == "" {
		return "", fmt.Errorf("refresh_token not found in response")
	}

	refreshToken := tokenResult.Token.RefreshToken

	// Step 8: Use refresh token to get a new access token
	tokenData := url.Values{}
	tokenData.Set("grant_type", "refresh_token")
	tokenData.Set("refresh_token", refreshToken)

	req, err := http.NewRequest("POST", testServerURL+"/oauth2/token", bytes.NewBufferString(tokenData.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create refresh token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientID, clientSecret)

	resp, err := ts.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send refresh token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("refresh token request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var refreshTokenResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&refreshTokenResp); err != nil {
		return "", fmt.Errorf("failed to decode refresh token response: %w", err)
	}

	newAccessToken, ok := refreshTokenResp["access_token"].(string)
	if !ok || newAccessToken == "" {
		return "", fmt.Errorf("access_token not found in refresh token response")
	}

	return newAccessToken, nil
}

// getTokenExchangeToken gets an access token using token_exchange grant
func (ts *UserInfoTestSuite) getTokenExchangeToken(scope string) (string, error) {
	// First get a subject token using authorization_code
	subjectToken, err := ts.getAuthorizationCodeToken(scope)
	if err != nil {
		return "", err
	}

	// Exchange the subject token for a new token
	tokenData := url.Values{}
	tokenData.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
	tokenData.Set("subject_token", subjectToken)
	tokenData.Set("subject_token_type", "urn:ietf:params:oauth:token-type:access_token")
	tokenData.Set("scope", scope)

	req, err := http.NewRequest("POST", testServerURL+"/oauth2/token", bytes.NewBufferString(tokenData.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientID, clientSecret)

	resp, err := ts.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token exchange request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var tokenResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	accessToken, ok := tokenResp["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("access_token not found in token exchange response")
	}

	return accessToken, nil
}

// callUserInfo calls the UserInfo endpoint with the given access token
func (ts *UserInfoTestSuite) callUserInfo(accessToken string) (*http.Response, error) {
	req, err := http.NewRequest("GET", testServerURL+"/oauth2/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := ts.client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// TestUserInfo_ClientCredentialsGrant_Rejected tests that client_credentials grant tokens are rejected
func (ts *UserInfoTestSuite) TestUserInfo_ClientCredentialsGrant_Rejected() {
	// Get access token using client_credentials grant
	accessToken, err := ts.getClientCredentialsToken("read write")
	ts.Require().NoError(err, "Failed to get client_credentials token")
	ts.Require().NotEmpty(accessToken, "Access token should not be empty")

	// Call UserInfo endpoint
	resp, err := ts.callUserInfo(accessToken)
	ts.Require().NoError(err, "Failed to call UserInfo endpoint")
	defer resp.Body.Close()

	// Should return 401 Unauthorized
	assert.Equal(ts.T(), http.StatusUnauthorized, resp.StatusCode, "Should return 401 for client_credentials grant")

	// Parse error response
	var errorResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&errorResp)
	ts.Require().NoError(err, "Failed to parse error response")

	// Verify error details
	assert.Equal(ts.T(), "invalid_token", errorResp["error"], "Error should be invalid_token")
	assert.Contains(ts.T(), errorResp["error_description"].(string), "client_credentials", "Error description should mention client_credentials")
}

// TestUserInfo_AuthorizationCodeGrant_Allowed tests that authorization_code grant tokens are allowed
func (ts *UserInfoTestSuite) TestUserInfo_AuthorizationCodeGrant_Allowed() {
	// Get access token using authorization_code grant
	accessToken, err := ts.getAuthorizationCodeToken("openid profile email")
	ts.Require().NoError(err, "Failed to get authorization_code token")
	ts.Require().NotEmpty(accessToken, "Access token should not be empty")

	// Call UserInfo endpoint
	resp, err := ts.callUserInfo(accessToken)
	ts.Require().NoError(err, "Failed to call UserInfo endpoint")
	defer resp.Body.Close()

	// Should return 200 OK
	assert.Equal(ts.T(), http.StatusOK, resp.StatusCode, "Should return 200 for authorization_code grant")

	// Parse response
	var userInfo map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	ts.Require().NoError(err, "Failed to parse UserInfo response")

	// Verify response contains sub claim
	assert.Contains(ts.T(), userInfo, "sub", "Response should contain sub claim")
	assert.NotEmpty(ts.T(), userInfo["sub"], "Sub claim should not be empty")

	// Verify response contains user attributes based on scopes
	// Note: The actual attributes depend on the token configuration
	ts.T().Logf("UserInfo response: %+v", userInfo)
}

// TestUserInfo_RefreshTokenGrant_Allowed tests that refresh_token grant tokens are allowed
func (ts *UserInfoTestSuite) TestUserInfo_RefreshTokenGrant_Allowed() {
	// Get access token using refresh_token grant
	accessToken, err := ts.getRefreshToken("openid profile email")
	ts.Require().NoError(err, "Failed to get refresh_token token")
	ts.Require().NotEmpty(accessToken, "Access token should not be empty")

	// Call UserInfo endpoint
	resp, err := ts.callUserInfo(accessToken)
	ts.Require().NoError(err, "Failed to call UserInfo endpoint")
	defer resp.Body.Close()

	// Should return 200 OK
	assert.Equal(ts.T(), http.StatusOK, resp.StatusCode, "Should return 200 for refresh_token grant")

	// Parse response
	var userInfo map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	ts.Require().NoError(err, "Failed to parse UserInfo response")

	// Verify response contains sub claim
	assert.Contains(ts.T(), userInfo, "sub", "Response should contain sub claim")
	assert.NotEmpty(ts.T(), userInfo["sub"], "Sub claim should not be empty")

	ts.T().Logf("UserInfo response: %+v", userInfo)
}

// TestUserInfo_TokenExchangeGrant_Allowed tests that token_exchange grant tokens are allowed
func (ts *UserInfoTestSuite) TestUserInfo_TokenExchangeGrant_Allowed() {
	// Get access token using token_exchange grant
	accessToken, err := ts.getTokenExchangeToken("openid profile email")
	ts.Require().NoError(err, "Failed to get token_exchange token")
	ts.Require().NotEmpty(accessToken, "Access token should not be empty")

	// Call UserInfo endpoint
	resp, err := ts.callUserInfo(accessToken)
	ts.Require().NoError(err, "Failed to call UserInfo endpoint")
	defer resp.Body.Close()

	// Should return 200 OK
	assert.Equal(ts.T(), http.StatusOK, resp.StatusCode, "Should return 200 for token_exchange grant")

	// Parse response
	var userInfo map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	ts.Require().NoError(err, "Failed to parse UserInfo response")

	// Verify response contains sub claim
	assert.Contains(ts.T(), userInfo, "sub", "Response should contain sub claim")
	assert.NotEmpty(ts.T(), userInfo["sub"], "Sub claim should not be empty")

	ts.T().Logf("UserInfo response: %+v", userInfo)
}

// TestUserInfo_InvalidToken tests that invalid tokens are rejected
func (ts *UserInfoTestSuite) TestUserInfo_InvalidToken() {
	// Call UserInfo endpoint with invalid token
	resp, err := ts.callUserInfo("invalid_token")
	ts.Require().NoError(err, "Failed to call UserInfo endpoint")
	defer resp.Body.Close()

	// Should return 401 Unauthorized
	assert.Equal(ts.T(), http.StatusUnauthorized, resp.StatusCode, "Should return 401 for invalid token")

	// Parse error response
	var errorResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&errorResp)
	ts.Require().NoError(err, "Failed to parse error response")

	// Verify error details
	assert.Equal(ts.T(), "invalid_token", errorResp["error"], "Error should be invalid_token")
}

func (ts *UserInfoTestSuite) TestUserInfo_MissingToken() {
	// Call UserInfo endpoint without token
	req, err := http.NewRequest("GET", testServerURL+"/oauth2/userinfo", nil)
	ts.Require().NoError(err, "Failed to create request")

	resp, err := ts.client.Do(req)
	ts.Require().NoError(err, "Failed to call UserInfo endpoint")
	defer resp.Body.Close()

	// Should return 401 Unauthorized (OAuth2 spec: missing authentication returns 401)
	assert.Equal(ts.T(), http.StatusUnauthorized, resp.StatusCode, "Should return 401 for missing token")

	// Parse error response
	var errorResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&errorResp)
	ts.Require().NoError(err, "Failed to parse error response")

	// Verify error details
	assert.Equal(ts.T(), "invalid_request", errorResp["error"], "Error should be invalid_request")
}

func (ts *UserInfoTestSuite) TestUserInfo_SeparateAttributesConfiguration() {
	// Create a dedicated app for this test with specific UserInfo config
	// UserInfo config allows ONLY "email", while IDToken config allows "email", "firstName", "lastName"
	config := map[string]interface{}{
		"client_id":      "userinfo_config_test_client",
		"client_secret":  "userinfo_config_test_secret",
		"redirect_uris":  []string{redirectURI},
		"grant_types":    []string{"authorization_code"},
		"response_types": []string{"code"},
		"scopes":         []string{"openid", "profile", "email"},
		"token": map[string]interface{}{
			"id_token": map[string]interface{}{
				"user_attributes": []string{"email", "firstName", "lastName"},
			},
		},
		"user_info": map[string]interface{}{
			"user_attributes": []string{"email"}, // UserInfo strictly limited to email
		},
		"scope_claims": map[string][]string{
			"profile": {"firstName", "lastName", "name"},
			"email":   {"email"},
		},
	}

	appID := ts.createApplicationWithConfig("UserInfoConfigTestApp", config)
	defer ts.deleteApplication(appID)

	// Get access token using the new app's credentials
	accessToken, err := ts.getAuthorizationCodeTokenWithClient(
		"openid profile email", "userinfo_config_test_client", "userinfo_config_test_secret")
	ts.Require().NoError(err, "Failed to get authorization_code token")

	// Call UserInfo endpoint
	resp, err := ts.callUserInfo(accessToken)
	ts.Require().NoError(err, "Failed to call UserInfo endpoint")
	defer resp.Body.Close()

	assert.Equal(ts.T(), http.StatusOK, resp.StatusCode)

	var userInfo map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	ts.Require().NoError(err)

	// Verify sub claim is always present (required by OIDC spec)
	assert.Contains(ts.T(), userInfo, "sub")

	// Verify ONLY email is returned (plus sub), and NOT firstName/lastName
	// despite being in profile scope and IDToken config
	assert.Contains(ts.T(), userInfo, "email")
	assert.NotContains(ts.T(), userInfo, "firstName")
	assert.NotContains(ts.T(), userInfo, "lastName")
}

func (ts *UserInfoTestSuite) TestUserInfo_FallbackConfiguration() {
	// Create app with NO UserInfo config, but with IDToken config
	config := map[string]interface{}{
		"client_id":      "userinfo_fallback_test_client",
		"client_secret":  "userinfo_fallback_test_secret",
		"redirect_uris":  []string{redirectURI},
		"grant_types":    []string{"authorization_code"},
		"response_types": []string{"code"},
		"scopes":         []string{"openid", "profile", "email"},
		"token": map[string]interface{}{
			"id_token": map[string]interface{}{
				"user_attributes": []string{"email", "firstName"},
			},
		},
		// No user_info config
		"scope_claims": map[string][]string{
			"profile": {"firstName", "lastName", "name"},
			"email":   {"email"},
		},
	}

	appID := ts.createApplicationWithConfig("UserInfoFallbackTestApp", config)
	defer ts.deleteApplication(appID)

	accessToken, err := ts.getAuthorizationCodeTokenWithClient(
		"openid profile email", "userinfo_fallback_test_client", "userinfo_fallback_test_secret")
	ts.Require().NoError(err)

	resp, err := ts.callUserInfo(accessToken)
	ts.Require().NoError(err)
	defer resp.Body.Close()

	var userInfo map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	ts.Require().NoError(err)

	// Verify sub claim is always present (required by OIDC spec)
	assert.Contains(ts.T(), userInfo, "sub")

	// Should fallback to IDToken attributes: email AND firstName
	assert.Contains(ts.T(), userInfo, "email")
	assert.Contains(ts.T(), userInfo, "firstName")
	assert.Equal(ts.T(), "UserInfo", userInfo["firstName"])
}

func (ts *UserInfoTestSuite) TestUserInfo_DefaultClaims() {
	// Create app with UserInfo config that includes default claims
	config := map[string]interface{}{
		"client_id":      "userinfo_default_claims_client",
		"client_secret":  "userinfo_default_claims_secret",
		"redirect_uris":  []string{redirectURI},
		"grant_types":    []string{"authorization_code"},
		"response_types": []string{"code"},
		"scopes":         []string{"openid", "profile"},
		"token": map[string]interface{}{
			"id_token": map[string]interface{}{
				"user_attributes": []string{"email"},
			},
		},
		"user_info": map[string]interface{}{
			"user_attributes": []string{"userType", "ouId", "ouHandle", "ouName", "email"},
		},
		"scope_claims": map[string][]string{
			"profile": {"firstName", "lastName", "userType", "ouId", "ouHandle", "ouName"},
		},
	}

	appID := ts.createApplicationWithConfig("UserInfoDefaultClaimsApp", config)
	defer ts.deleteApplication(appID)

	// Get access token using the new app's credentials
	accessToken, err := ts.getAuthorizationCodeTokenWithClient(
		"openid profile", "userinfo_default_claims_client", "userinfo_default_claims_secret")
	ts.Require().NoError(err, "Failed to get authorization_code token")

	// Call UserInfo endpoint
	resp, err := ts.callUserInfo(accessToken)
	ts.Require().NoError(err, "Failed to call UserInfo endpoint")
	defer resp.Body.Close()

	assert.Equal(ts.T(), http.StatusOK, resp.StatusCode)

	var userInfo map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	ts.Require().NoError(err)

	// Verify sub claim is always present (required by OIDC spec)
	assert.Contains(ts.T(), userInfo, "sub")

	// Verify default claims are present
	assert.Contains(ts.T(), userInfo, "userType", "userType should be in response")
	assert.Contains(ts.T(), userInfo, "ouId", "ouId should be in response")
	assert.Contains(ts.T(), userInfo, "ouHandle", "ouHandle should be in response")
	assert.Contains(ts.T(), userInfo, "ouName", "ouName should be in response")

	// Verify the values match the test user's OU
	assert.Equal(ts.T(), "userinfo-person", userInfo["userType"])
	assert.Equal(ts.T(), ts.ouID, userInfo["ouId"])
	assert.Equal(ts.T(), "userinfo-test-ou", userInfo["ouHandle"])
	assert.Equal(ts.T(), "UserInfo Test OU", userInfo["ouName"])
}

// createApplicationWithConfig creates an OAuth application with the given config
func (ts *UserInfoTestSuite) createApplicationWithConfig(name string, oauthConfig map[string]interface{}) string {
	app := map[string]interface{}{
		"name":                         name,
		"description":                  "Application for UserInfo integration tests",
		"auth_flow_id":                 ts.flowID,
		"is_registration_flow_enabled": false,
		"allowed_user_types":           []string{"userinfo-person"},
		"inbound_auth_config": []map[string]interface{}{
			{
				"type":   "oauth2",
				"config": oauthConfig,
			},
		},
	}

	jsonData, err := json.Marshal(app)
	ts.Require().NoError(err)

	req, err := http.NewRequest("POST", testServerURL+"/applications", bytes.NewBuffer(jsonData))
	ts.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.client.Do(req)
	ts.Require().NoError(err)
	defer resp.Body.Close()

	ts.Require().Equal(http.StatusCreated, resp.StatusCode)

	var respData map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&respData)
	ts.Require().NoError(err)

	return respData["id"].(string)
}

// getAuthorizationCodeTokenWithClient is a helper that performs the full auth code flow
// using the specified client credentials
func (ts *UserInfoTestSuite) getAuthorizationCodeTokenWithClient(scope, cID, cSecret string) (string, error) {
	// 1. Initiate
	authzResp, err := testutils.InitiateAuthorizationFlow(cID, redirectURI, "code", scope, "test_state")
	if err != nil {
		return "", err
	}
	defer authzResp.Body.Close()
	location := authzResp.Header.Get("Location")

	// 2. Extract
	authId, flowID, err := testutils.ExtractAuthData(location)
	if err != nil {
		return "", err
	}

	// 3. Initiate Auth
	_, err = testutils.ExecuteAuthenticationFlow(flowID, nil, "")
	if err != nil {
		return "", err
	}

	// 4. Authenticate
	authInputs := map[string]string{
		"username": "userinfo_test_user",
		"password": "SecurePass123!",
	}
	flowStep, err := testutils.ExecuteAuthenticationFlow(flowID, authInputs, "action_001")
	if err != nil {
		return "", err
	}

	// 5. Complete
	if flowStep.Assertion == "" {
		return "", fmt.Errorf("Assertion missing")
	}
	authzResponse, err := testutils.CompleteAuthorization(authId, flowStep.Assertion)
	if err != nil {
		return "", err
	}

	// 6. Extract Code
	code, err := testutils.ExtractAuthorizationCode(authzResponse.RedirectURI)
	if err != nil {
		return "", err
	}

	// 7. Exchange
	tokenResult, err := testutils.RequestToken(cID, cSecret, code, redirectURI, "authorization_code")
	if err != nil {
		return "", err
	}
	if tokenResult.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d", tokenResult.StatusCode)
	}
	if tokenResult.Token == nil {
		return "", fmt.Errorf("token response is nil")
	}
	return tokenResult.Token.AccessToken, nil
}
