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

	// Create OAuth application
	ts.applicationID = ts.createTestApplication()
}

func (ts *UserInfoTestSuite) TearDownSuite() {
	// Clean up application
	if ts.applicationID != "" {
		ts.deleteApplication(ts.applicationID)
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

func (ts *UserInfoTestSuite) createTestApplication() string {
	app := map[string]interface{}{
		"name":                         appName,
		"description":                  "Application for UserInfo integration tests",
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
				},
			},
		},
		"token_config": map[string]interface{}{
			"id_token": map[string]interface{}{
				"user_attributes": []string{"email", "firstName", "lastName", "name"},
				"scope_claims": map[string][]string{
					"profile": {"firstName", "lastName", "name"},
					"email":   {"email"},
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

	// Step 3: Execute authentication flow
	authInputs := map[string]string{
		"username": "userinfo_test_user",
		"password": "SecurePass123!",
	}
	flowStep, err := testutils.ExecuteAuthenticationFlow(flowID, authInputs, "")
	if err != nil {
		return "", fmt.Errorf("failed to execute authentication flow: %w", err)
	}

	if flowStep.Assertion == "" {
		return "", fmt.Errorf("assertion not found in flow step")
	}

	// Step 4: Complete authorization
	authzResponse, err := testutils.CompleteAuthorization(authId, flowStep.Assertion)
	if err != nil {
		return "", fmt.Errorf("failed to complete authorization: %w", err)
	}

	// Step 5: Extract authorization code
	code, err := testutils.ExtractAuthorizationCode(authzResponse.RedirectURI)
	if err != nil {
		return "", fmt.Errorf("failed to extract authorization code: %w", err)
	}

	// Step 6: Exchange code for token
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

	// Step 3: Execute authentication flow
	authInputs := map[string]string{
		"username": "userinfo_test_user",
		"password": "SecurePass123!",
	}
	flowStep, err := testutils.ExecuteAuthenticationFlow(flowID, authInputs, "")
	if err != nil {
		return "", fmt.Errorf("failed to execute authentication flow: %w", err)
	}

	if flowStep.Assertion == "" {
		return "", fmt.Errorf("assertion not found in flow step")
	}

	// Step 4: Complete authorization
	authzResponse, err := testutils.CompleteAuthorization(authId, flowStep.Assertion)
	if err != nil {
		return "", fmt.Errorf("failed to complete authorization: %w", err)
	}

	// Step 5: Extract authorization code
	code, err := testutils.ExtractAuthorizationCode(authzResponse.RedirectURI)
	if err != nil {
		return "", fmt.Errorf("failed to extract authorization code: %w", err)
	}

	// Step 6: Exchange code for token (this should include refresh_token)
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

	// Step 7: Use refresh token to get a new access token
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

// TestUserInfo_MissingToken tests that missing tokens are rejected
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
