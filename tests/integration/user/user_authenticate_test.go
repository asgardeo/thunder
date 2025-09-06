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

package user

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

var (
	authTestOU = OUCreateRequest{
		Handle:      "auth-test-ou",
		Name:        "Authentication Test Organization Unit",
		Description: "Organization unit for user authentication testing",
		Parent:      nil,
	}

	testUserForAuth = User{
		Type:       "person",
		Attributes: json.RawMessage(`{"username": "testuser", "password": "testpass123", "email": "testuser@example.com", "firstName": "Test", "lastName": "User"}`),
	}
)

var (
	authTestOUID string
)

type UserAuthenticateAPITestSuite struct {
	suite.Suite
	createdUserID string
}

func TestUserAuthenticateAPITestSuite(t *testing.T) {
	suite.Run(t, new(UserAuthenticateAPITestSuite))
}

// SetupSuite creates a test organization unit and user for authentication tests
func (ts *UserAuthenticateAPITestSuite) SetupSuite() {
	// Create test organization unit
	ouID, err := ts.createAuthTestOrganizationUnit(authTestOU)
	if err != nil {
		ts.T().Fatalf("Failed to create test organization unit during setup: %v", err)
	}
	authTestOUID = ouID

	// Create test user with the created OU
	testUserForAuth := testUserForAuth
	testUserForAuth.OrganizationUnit = authTestOUID
	userID, err := ts.createTestUserWithOU(testUserForAuth)
	if err != nil {
		ts.T().Fatalf("Failed to create test user during setup: %v", err)
	}
	ts.createdUserID = userID
}

// TearDownSuite deletes the test user and organization unit
func (ts *UserAuthenticateAPITestSuite) TearDownSuite() {
	// Delete test user
	if ts.createdUserID != "" {
		err := ts.deleteTestUser(ts.createdUserID)
		if err != nil {
			ts.T().Logf("Failed to delete test user during teardown: %v", err)
		}
	}

	// Delete test organization unit
	if authTestOUID != "" {
		err := ts.deleteAuthTestOrganizationUnit(authTestOUID)
		if err != nil {
			ts.T().Logf("Failed to delete test organization unit during teardown: %v", err)
		}
	}
}

// TestAuthenticateUserWithUsernamePassword tests successful authentication with username and password
func (ts *UserAuthenticateAPITestSuite) TestAuthenticateUserWithUsernamePassword() {
	authRequest := map[string]interface{}{
		"username": "testuser",
		"password": "testpass123",
	}

	response, statusCode, err := ts.sendAuthenticateRequest(authRequest)
	ts.Require().NoError(err, "Failed to send authenticate request")
	ts.Require().Equal(http.StatusOK, statusCode, "Expected status 200 for successful authentication")

	// Validate response structure
	ts.Require().NotEmpty(response.ID, "Response should contain user ID")
	ts.Require().Equal("person", response.Type, "Response should contain correct user type")
	ts.Require().Equal(authTestOUID, response.OrganizationUnit, "Response should contain correct organization unit")
	ts.Require().Equal(ts.createdUserID, response.ID, "Response should contain the correct user ID")
}

// TestAuthenticateUserWithEmailPassword tests successful authentication with email and password
func (ts *UserAuthenticateAPITestSuite) TestAuthenticateUserWithEmailPassword() {
	authRequest := map[string]interface{}{
		"email":    "testuser@example.com",
		"password": "testpass123",
	}

	response, statusCode, err := ts.sendAuthenticateRequest(authRequest)
	ts.Require().NoError(err, "Failed to send authenticate request")
	ts.Require().Equal(http.StatusOK, statusCode, "Expected status 200 for successful authentication")

	// Validate response structure
	ts.Require().NotEmpty(response.ID, "Response should contain user ID")
	ts.Require().Equal("person", response.Type, "Response should contain correct user type")
	ts.Require().Equal(authTestOUID, response.OrganizationUnit, "Response should contain correct organization unit")
	ts.Require().Equal(ts.createdUserID, response.ID, "Response should contain the correct user ID")
}

// TestAuthenticateUserWithInvalidPassword tests authentication failure with invalid password
func (ts *UserAuthenticateAPITestSuite) TestAuthenticateUserWithInvalidPassword() {
	authRequest := map[string]interface{}{
		"username": "testuser",
		"password": "wrongpassword",
	}

	_, statusCode, err := ts.sendAuthenticateRequestExpectingError(authRequest)
	ts.Require().NoError(err, "Failed to send authenticate request")
	ts.Require().Equal(http.StatusUnauthorized, statusCode, "Expected status 401 for authentication failure")
}

// TestAuthenticateUserWithInvalidUsername tests authentication failure with invalid username
func (ts *UserAuthenticateAPITestSuite) TestAuthenticateUserWithInvalidUsername() {
	authRequest := map[string]interface{}{
		"username": "nonexistentuser",
		"password": "testpass123",
	}

	_, statusCode, err := ts.sendAuthenticateRequestExpectingError(authRequest)
	ts.Require().NoError(err, "Failed to send authenticate request")
	ts.Require().Equal(http.StatusNotFound, statusCode, "Expected status 404 for user not found")
}

// TestAuthenticateUserWithMissingPassword tests validation error when password is missing
func (ts *UserAuthenticateAPITestSuite) TestAuthenticateUserWithMissingPassword() {
	authRequest := map[string]interface{}{
		"username": "testuser",
	}

	errorResp, statusCode, err := ts.sendAuthenticateRequestExpectingError(authRequest)
	ts.Require().NoError(err, "Failed to send authenticate request")
	ts.Require().Equal(http.StatusBadRequest, statusCode, "Expected status 400 for missing credentials")
	ts.Require().Equal("USR-1017", errorResp.Code, "Expected error code USR-1017 for missing credentials")
}

// TestAuthenticateUserWithMissingIdentifyingAttributes tests validation error when identifying attributes are missing
func (ts *UserAuthenticateAPITestSuite) TestAuthenticateUserWithMissingIdentifyingAttributes() {
	authRequest := map[string]interface{}{
		"password": "testpass123",
	}

	errorResp, statusCode, err := ts.sendAuthenticateRequestExpectingError(authRequest)
	ts.Require().NoError(err, "Failed to send authenticate request")
	ts.Require().Equal(http.StatusBadRequest, statusCode, "Expected status 400 for missing required fields")
	ts.Require().Equal("USR-1016", errorResp.Code, "Expected error code USR-1016 for missing required fields")
}

// TestAuthenticateUserWithEmptyRequest tests validation error when request body is empty
func (ts *UserAuthenticateAPITestSuite) TestAuthenticateUserWithEmptyRequest() {
	authRequest := map[string]interface{}{}

	errorResp, statusCode, err := ts.sendAuthenticateRequestExpectingError(authRequest)
	ts.Require().NoError(err, "Failed to send authenticate request")
	ts.Require().Equal(http.StatusBadRequest, statusCode, "Expected status 400 for invalid request format")
	ts.Require().Equal("USR-1001", errorResp.Code, "Expected error code USR-1001 for invalid request format")
}

// TestAuthenticateUserWithMalformedJSON tests validation error when request body is malformed
func (ts *UserAuthenticateAPITestSuite) TestAuthenticateUserWithMalformedJSON() {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	malformedJSON := `{"username": "testuser", "password": "testpass123"`
	req, err := http.NewRequest("POST", testServerURL+"/users/authenticate", bytes.NewReader([]byte(malformedJSON)))
	ts.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	ts.Require().NoError(err)
	defer resp.Body.Close()

	ts.Require().Equal(http.StatusBadRequest, resp.StatusCode, "Expected status 400 for malformed JSON")

	var errorResp ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&errorResp)
	ts.Require().NoError(err)
	ts.Require().Equal("USR-1001", errorResp.Code, "Expected error code USR-1001 for invalid request format")
}

// Helper method to create a test organization unit
func (ts *UserAuthenticateAPITestSuite) createAuthTestOrganizationUnit(ouRequest OUCreateRequest) (string, error) {
	ouJSON, err := json.Marshal(ouRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal OU request: %w", err)
	}

	reqBody := bytes.NewReader(ouJSON)
	req, err := http.NewRequest("POST", testServerURL+"/organization-units", reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to create OU request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send OU request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("expected status 201, got %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var ouResponse OUResponse
	if err := json.NewDecoder(resp.Body).Decode(&ouResponse); err != nil {
		return "", fmt.Errorf("failed to decode OU response: %w", err)
	}

	return ouResponse.ID, nil
}

// Helper method to delete a test organization unit
func (ts *UserAuthenticateAPITestSuite) deleteAuthTestOrganizationUnit(ouID string) error {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("DELETE", testServerURL+"/organization-units/"+ouID, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send delete request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("expected status 204, got %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// Helper method to create a test user with organization unit
func (ts *UserAuthenticateAPITestSuite) createTestUserWithOU(user User) (string, error) {
	userJSON, err := json.Marshal(user)
	if err != nil {
		return "", fmt.Errorf("failed to marshal test user: %w", err)
	}

	reqBody := bytes.NewReader(userJSON)
	req, err := http.NewRequest("POST", testServerURL+"/users", reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("expected status 201, got %d. Response body: %s", resp.StatusCode, string(body))
	}

	var respBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return "", fmt.Errorf("failed to parse response body: %w", err)
	}

	id, ok := respBody["id"].(string)
	if !ok {
		return "", fmt.Errorf("response does not contain id")
	}

	return id, nil
}

// Helper method to delete a test user
func (ts *UserAuthenticateAPITestSuite) deleteTestUser(userID string) error {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("DELETE", testServerURL+"/users/"+userID, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send delete request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("expected status 204, got %d. Response body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Helper method to send authenticate request expecting success
func (ts *UserAuthenticateAPITestSuite) sendAuthenticateRequest(authRequest map[string]interface{}) (*AuthenticateUserResponse, int, error) {
	reqJSON, err := json.Marshal(authRequest)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("POST", testServerURL+"/users/authenticate", bytes.NewReader(reqJSON))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var response AuthenticateUserResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, resp.StatusCode, nil
}

// Helper method to send authenticate request expecting error
func (ts *UserAuthenticateAPITestSuite) sendAuthenticateRequestExpectingError(authRequest map[string]interface{}) (*ErrorResponse, int, error) {
	reqJSON, err := json.Marshal(authRequest)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("POST", testServerURL+"/users/authenticate", bytes.NewReader(reqJSON))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var errorResp ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&errorResp)
	if err != nil {
		// For some error cases, the response might not be JSON
		body, _ := io.ReadAll(resp.Body)
		return nil, resp.StatusCode, fmt.Errorf("failed to decode error response: %w. Body: %s", err, string(body))
	}

	return &errorResp, resp.StatusCode, nil
}
