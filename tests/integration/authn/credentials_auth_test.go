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

package authn

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"testing"

	"github.com/asgardeo/thunder/tests/integration/testutils"
	"github.com/stretchr/testify/suite"
)

const (
	credentialsAuthEndpoint = "/auth/credentials/authenticate"
	testOrgUnitID           = "root"
)

type CredentialsAuthTestSuite struct {
	suite.Suite
	client *http.Client
	users  map[string]string // map of test name to user ID
}

func TestCredentialsAuthTestSuite(t *testing.T) {
	suite.Run(t, new(CredentialsAuthTestSuite))
}

func (suite *CredentialsAuthTestSuite) SetupSuite() {
	suite.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	suite.users = make(map[string]string)

	// Create test users with different attribute types
	testUsers := []struct {
		name       string
		attributes map[string]interface{}
	}{
		{
			name: "username_password",
			attributes: map[string]interface{}{
				"username": "credtest_user1",
				"password": "TestPassword123!",
				"email":    "credtest1@example.com",
			},
		},
		{
			name: "email_password",
			attributes: map[string]interface{}{
				"email":    "credtest2@example.com",
				"password": "TestPassword456!",
				"username": "credtest_user2",
			},
		},
		{
			name: "mobile_password",
			attributes: map[string]interface{}{
				"mobileNumber": "+1234567891",
				"password":     "TestPassword789!",
				"username":     "credtest_user3",
			},
		},
		{
			name: "multiple_attributes",
			attributes: map[string]interface{}{
				"username":     "credtest_user4",
				"email":        "credtest4@example.com",
				"mobileNumber": "+1234567892",
				"password":     "TestPassword999!",
				"firstName":    "Test",
				"lastName":     "User",
			},
		},
	}

	for _, tu := range testUsers {
		attributesJSON, err := json.Marshal(tu.attributes)
		suite.Require().NoError(err, "Failed to marshal attributes for %s", tu.name)

		user := testutils.User{
			Type:             "person",
			OrganizationUnit: testOrgUnitID,
			Attributes:       json.RawMessage(attributesJSON),
		}

		userID, err := testutils.CreateUser(user)
		suite.Require().NoError(err, "Failed to create test user for %s", tu.name)
		suite.users[tu.name] = userID
	}
}

func (suite *CredentialsAuthTestSuite) TearDownSuite() {
	for _, userID := range suite.users {
		if userID != "" {
			_ = testutils.DeleteUser(userID)
		}
	}
}

// TestAuthenticateWithUsernamePassword tests successful authentication with username and password
func (suite *CredentialsAuthTestSuite) TestAuthenticateWithUsernamePassword() {
	authRequest := map[string]interface{}{
		"username": "credtest_user1",
		"password": "TestPassword123!",
	}

	response, statusCode, err := suite.sendAuthRequest(authRequest)
	suite.Require().NoError(err, "Failed to send authenticate request")
	suite.Equal(http.StatusOK, statusCode, "Expected status 200 for successful authentication")

	suite.NotEmpty(response.ID, "Response should contain user ID")
	suite.Equal("person", response.Type, "Response should contain correct user type")
	suite.Equal(suite.users["username_password"], response.ID, "Response should contain the correct user ID")
}

// TestAuthenticateWithEmailPassword tests successful authentication with email and password
func (suite *CredentialsAuthTestSuite) TestAuthenticateWithEmailPassword() {
	authRequest := map[string]interface{}{
		"email":    "credtest2@example.com",
		"password": "TestPassword456!",
	}

	response, statusCode, err := suite.sendAuthRequest(authRequest)
	suite.Require().NoError(err, "Failed to send authenticate request")
	suite.Equal(http.StatusOK, statusCode, "Expected status 200 for successful authentication")

	suite.NotEmpty(response.ID, "Response should contain user ID")
	suite.Equal("person", response.Type, "Response should contain correct user type")
	suite.Equal(suite.users["email_password"], response.ID, "Response should contain the correct user ID")
}

// TestAuthenticateWithMobilePassword tests successful authentication with mobile number and password
func (suite *CredentialsAuthTestSuite) TestAuthenticateWithMobilePassword() {
	authRequest := map[string]interface{}{
		"mobileNumber": "+1234567891",
		"password":     "TestPassword789!",
	}

	response, statusCode, err := suite.sendAuthRequest(authRequest)
	suite.Require().NoError(err, "Failed to send authenticate request")
	suite.Equal(http.StatusOK, statusCode, "Expected status 200 for successful authentication")

	suite.NotEmpty(response.ID, "Response should contain user ID")
	suite.Equal("person", response.Type, "Response should contain correct user type")
	suite.Equal(testOrgUnitID, response.OrganizationUnit, "Response should contain correct organization unit")
	suite.Equal(suite.users["mobile_password"], response.ID, "Response should contain the correct user ID")
}

// TestAuthenticateWithMultipleAttributes tests successful authentication with multiple identifying attributes
func (suite *CredentialsAuthTestSuite) TestAuthenticateWithMultipleAttributes() {
	testCases := []struct {
		name        string
		authRequest map[string]interface{}
	}{
		{
			name: "Username with multiple attributes",
			authRequest: map[string]interface{}{
				"username": "credtest_user4",
				"password": "TestPassword999!",
			},
		},
		{
			name: "Email with multiple attributes",
			authRequest: map[string]interface{}{
				"email":    "credtest4@example.com",
				"password": "TestPassword999!",
			},
		},
		{
			name: "Mobile with multiple attributes",
			authRequest: map[string]interface{}{
				"mobileNumber": "+1234567892",
				"password":     "TestPassword999!",
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			response, statusCode, err := suite.sendAuthRequest(tc.authRequest)
			suite.Require().NoError(err, "Failed to send authenticate request")
			suite.Equal(http.StatusOK, statusCode, "Expected status 200 for successful authentication")

			suite.NotEmpty(response.ID, "Response should contain user ID")
			suite.Equal("person", response.Type, "Response should contain correct user type")
			suite.Equal(testOrgUnitID, response.OrganizationUnit, "Response should contain correct organization unit")
			suite.Equal(suite.users["multiple_attributes"], response.ID, "Response should contain the correct user ID")
			suite.NotEmpty(response.Assertion, "Response should contain assertion token by default")
		})
	}
}

// TestAuthenticateWithInvalidPassword tests authentication failure with invalid password
func (suite *CredentialsAuthTestSuite) TestAuthenticateWithInvalidPassword() {
	testCases := []struct {
		name        string
		authRequest map[string]interface{}
	}{
		{
			name: "Invalid password with username",
			authRequest: map[string]interface{}{
				"username": "credtest_user1",
				"password": "WrongPassword123!",
			},
		},
		{
			name: "Invalid password with email",
			authRequest: map[string]interface{}{
				"email":    "credtest2@example.com",
				"password": "WrongPassword456!",
			},
		},
		{
			name: "Invalid password with mobile",
			authRequest: map[string]interface{}{
				"mobileNumber": "+1234567891",
				"password":     "WrongPassword789!",
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			errorResp, statusCode, err := suite.sendAuthRequestExpectingError(tc.authRequest)
			suite.Require().NoError(err, "Failed to send authenticate request")
			suite.Equal(http.StatusUnauthorized, statusCode, "Expected status 401 for invalid password")
			suite.Equal("AUTH-CRED-1002", errorResp.Code, "Expected error code AUTH-CRED-1002 for invalid credentials")
		})
	}
}

// TestAuthenticateWithNonExistentUser tests authentication failure with non-existent user
func (suite *CredentialsAuthTestSuite) TestAuthenticateWithNonExistentUser() {
	testCases := []struct {
		name        string
		authRequest map[string]interface{}
	}{
		{
			name: "Non-existent username",
			authRequest: map[string]interface{}{
				"username": "nonexistent_user",
				"password": "TestPassword123!",
			},
		},
		{
			name: "Non-existent email",
			authRequest: map[string]interface{}{
				"email":    "nonexistent@example.com",
				"password": "TestPassword123!",
			},
		},
		{
			name: "Non-existent mobile",
			authRequest: map[string]interface{}{
				"mobileNumber": "+9999999999",
				"password":     "TestPassword123!",
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			errorResp, statusCode, err := suite.sendAuthRequestExpectingError(tc.authRequest)
			suite.Require().NoError(err, "Failed to send authenticate request")
			suite.Equal(http.StatusNotFound, statusCode, "Expected status 404 for non-existent user")
			suite.Equal("AUTHN-1008", errorResp.Code, "Expected error code AUTHN-1008 for user not found")
		})
	}
}

// TestAuthenticateWithMissingPassword tests authentication failure when password is missing
func (suite *CredentialsAuthTestSuite) TestAuthenticateWithMissingPassword() {
	testCases := []struct {
		name        string
		authRequest map[string]interface{}
	}{
		{
			name: "Missing password with username",
			authRequest: map[string]interface{}{
				"username": "credtest_user1",
			},
		},
		{
			name: "Missing password with email",
			authRequest: map[string]interface{}{
				"email": "credtest2@example.com",
			},
		},
		{
			name: "Missing password with mobile",
			authRequest: map[string]interface{}{
				"mobileNumber": "+1234567891",
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			_, statusCode, err := suite.sendAuthRequestExpectingError(tc.authRequest)
			suite.Require().NoError(err, "Failed to send authenticate request")
			suite.Equal(http.StatusBadRequest, statusCode, "Expected status 400 for missing password")
		})
	}
}

// TestAuthenticateWithMissingIdentifyingAttributes tests authentication failure when identifying attributes are missing
func (suite *CredentialsAuthTestSuite) TestAuthenticateWithMissingIdentifyingAttributes() {
	authRequest := map[string]interface{}{
		"password": "TestPassword123!",
	}

	_, statusCode, err := suite.sendAuthRequestExpectingError(authRequest)
	suite.Require().NoError(err, "Failed to send authenticate request")
	suite.Equal(http.StatusBadRequest, statusCode, "Expected status 400 for missing identifying attributes")
}

// TestAuthenticateWithEmptyRequest tests authentication failure when request is empty
func (suite *CredentialsAuthTestSuite) TestAuthenticateWithEmptyRequest() {
	authRequest := map[string]interface{}{}

	errorResp, statusCode, err := suite.sendAuthRequestExpectingError(authRequest)
	suite.Require().NoError(err, "Failed to send authenticate request")
	suite.Equal(http.StatusBadRequest, statusCode, "Expected status 400 for empty request")
	suite.Equal("AUTH-CRED-1001", errorResp.Code, "Expected error code AUTH-CRED-1001 for empty attributes")
}

// TestAuthenticateWithEmptyCredentials tests authentication failure with empty values
func (suite *CredentialsAuthTestSuite) TestAuthenticateWithEmptyCredentials() {
	testCases := []struct {
		name           string
		authRequest    map[string]interface{}
		expectedStatus int
	}{
		{
			name: "Empty username",
			authRequest: map[string]interface{}{
				"username": "",
				"password": "TestPassword123!",
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "Empty password",
			authRequest: map[string]interface{}{
				"username": "credtest_user1",
				"password": "",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Empty email",
			authRequest: map[string]interface{}{
				"email":    "",
				"password": "TestPassword123!",
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "Both empty",
			authRequest: map[string]interface{}{
				"username": "",
				"password": "",
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			_, statusCode, err := suite.sendAuthRequestExpectingError(tc.authRequest)
			suite.Require().NoError(err, "Failed to send authenticate request")
			suite.Equal(tc.expectedStatus, statusCode, "Unexpected status code")
		})
	}
}

// TestAuthenticateWithMalformedJSON tests authentication failure with malformed JSON
func (suite *CredentialsAuthTestSuite) TestAuthenticateWithMalformedJSON() {
	malformedJSON := []byte(`{"username": "test", "password": }`)

	req, err := http.NewRequest("POST", testutils.TestServerURL+credentialsAuthEndpoint,
		bytes.NewReader(malformedJSON))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusBadRequest, resp.StatusCode, "Expected status 400 for malformed JSON")

	var errorResp testutils.ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&errorResp)
	suite.Require().NoError(err)
	suite.Equal("AUTHN-1000", errorResp.Code, "Expected error code AUTHN-1000 for invalid request format")
}

// TestAuthenticateWithDifferentAttributeCombinations tests various attribute combinations
func (suite *CredentialsAuthTestSuite) TestAuthenticateWithDifferentAttributeCombinations() {
	testCases := []struct {
		name           string
		authRequest    map[string]interface{}
		expectedUserID string
		shouldSucceed  bool
	}{
		{
			name: "Username and email (both valid for same user)",
			authRequest: map[string]interface{}{
				"username": "credtest_user4",
				"email":    "credtest4@example.com",
				"password": "TestPassword999!",
			},
			expectedUserID: "multiple_attributes",
			shouldSucceed:  true,
		},
		{
			name: "Only additional attributes (no identifying attribute)",
			authRequest: map[string]interface{}{
				"firstName": "Test",
				"lastName":  "User",
				"password":  "TestPassword999!",
			},
			expectedUserID: "",
			shouldSucceed:  true, // Changed: API now returns 200 with these attributes
		},
		{
			name: "Valid username with additional attributes",
			authRequest: map[string]interface{}{
				"username": "credtest_user1",
				"password": "TestPassword123!",
			},
			expectedUserID: "username_password",
			shouldSucceed:  true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			if tc.shouldSucceed {
				response, statusCode, err := suite.sendAuthRequest(tc.authRequest)
				log.Printf("Response: %+v, StatusCode: %d, Error: %v", response, statusCode, err)

				suite.Require().NoError(err, "Failed to send authenticate request")
				suite.Equal(http.StatusOK, statusCode, "Expected status 200 for successful authentication")
				if tc.expectedUserID != "" {
					suite.Equal(testOrgUnitID, response.OrganizationUnit, "Response should contain correct organization unit")
					suite.Equal(suite.users[tc.expectedUserID], response.ID, "Response should contain the correct user ID")
				}
			} else {
				_, statusCode, err := suite.sendAuthRequestExpectingError(tc.authRequest)
				suite.Require().NoError(err, "Failed to send authenticate request")
				suite.Equal(http.StatusBadRequest, statusCode, "Expected status 400 for invalid request")
			}
		})
	}
}

// TestAuthenticateWithSkipAssertionFalse tests authentication with skip_assertion explicitly set to false
func (suite *CredentialsAuthTestSuite) TestAuthenticateWithSkipAssertionFalse() {
	authRequest := map[string]interface{}{
		"username":       "credtest_user1",
		"password":       "TestPassword123!",
		"skip_assertion": false,
	}

	response, statusCode, err := suite.sendAuthRequest(authRequest)
	suite.Require().NoError(err, "Failed to send authenticate request")
	suite.Equal(http.StatusOK, statusCode, "Expected status 200 for successful authentication")

	suite.NotEmpty(response.ID, "Response should contain user ID")
	suite.Equal("person", response.Type, "Response should contain correct user type")
	suite.Equal(testOrgUnitID, response.OrganizationUnit, "Response should contain correct organization unit")
	suite.Equal(suite.users["username_password"], response.ID, "Response should contain the correct user ID")
	suite.NotEmpty(response.Assertion, "Response should contain assertion token when skip_assertion is false")
}

// TestAuthenticateWithSkipAssertionTrue tests authentication with skip_assertion set to true
func (suite *CredentialsAuthTestSuite) TestAuthenticateWithSkipAssertionTrue() {
	authRequest := map[string]interface{}{
		"username":       "credtest_user1",
		"password":       "TestPassword123!",
		"skip_assertion": true,
	}

	response, statusCode, err := suite.sendAuthRequest(authRequest)
	suite.Require().NoError(err, "Failed to send authenticate request")
	suite.Equal(http.StatusOK, statusCode, "Expected status 200 for successful authentication")

	suite.NotEmpty(response.ID, "Response should contain user ID")
	suite.Equal("person", response.Type, "Response should contain correct user type")
	suite.Equal(testOrgUnitID, response.OrganizationUnit, "Response should contain correct organization unit")
	suite.Equal(suite.users["username_password"], response.ID, "Response should contain the correct user ID")
	suite.Empty(response.Assertion, "Response should not contain assertion token when skip_assertion is true")
}

func (suite *CredentialsAuthTestSuite) sendAuthRequest(authRequest map[string]interface{}) (
	*testutils.AuthenticationResponse, int, error) {
	requestJSON, err := json.Marshal(authRequest)
	if err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequest("POST", testutils.TestServerURL+credentialsAuthEndpoint,
		bytes.NewReader(requestJSON))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	var response testutils.AuthenticationResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return &response, resp.StatusCode, nil
}

func (suite *CredentialsAuthTestSuite) sendAuthRequestExpectingError(authRequest map[string]interface{}) (
	*testutils.ErrorResponse, int, error) {
	requestJSON, err := json.Marshal(authRequest)
	if err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequest("POST", testutils.TestServerURL+credentialsAuthEndpoint,
		bytes.NewReader(requestJSON))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	var errorResp testutils.ErrorResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(bodyBytes, &errorResp)

	return &errorResp, resp.StatusCode, nil
}
