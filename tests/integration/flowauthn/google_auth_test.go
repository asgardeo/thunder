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

package flowauthn

import (
	"encoding/json"
	"net/url"
	"strings"
	"testing"

	"github.com/asgardeo/thunder/tests/integration/testutils"
	"github.com/stretchr/testify/suite"
)

var (
	googleAuthTestApp = testutils.Application{
		Name:                      "Google Auth Flow Test Application",
		Description:               "Application for testing Google authentication flows",
		IsRegistrationFlowEnabled: false,
		AuthFlowGraphID:           "auth_flow_config_google",
		RegistrationFlowGraphID:   "registration_flow_config_basic",
		ClientID:                  "google_auth_flow_test_client",
		ClientSecret:              "google_auth_flow_test_secret",
		RedirectURIs:              []string{"http://localhost:3000/callback"},
	}
)

var (
	googleAuthTestAppID string
	googleAuthTestOU = testutils.OrganizationUnit{
		Handle:      "google-auth-flow-test-ou",
		Name:        "Google Auth Flow Test OU",
		Description: "Organization unit for Google authentication flow tests",
	}
)



var googleUserSchema = testutils.UserSchema{
	Name: "google_flow_user",
	Schema: map[string]interface{}{
		"username": map[string]interface{}{
			"type": "string",
		},
		"password": map[string]interface{}{
			"type": "string",
		},
		"sub": map[string]interface{}{
			"type": "string",
		},
		"email": map[string]interface{}{
			"type": "string",
		},
		"givenName": map[string]interface{}{
			"type": "string",
		},
		"familyName": map[string]interface{}{
			"type": "string",
		},
	},
}

type GoogleAuthFlowTestSuite struct {
	suite.Suite
	mockGoogleServer *testutils.MockGoogleOIDCServer
	idpID            string
	userID           string
	userSchemaID     string
}

func TestGoogleAuthFlowTestSuite(t *testing.T) {
	suite.Run(t, new(GoogleAuthFlowTestSuite))
}

func (ts *GoogleAuthFlowTestSuite) SetupSuite() {
	// Get shared Google OIDC server (started once for all test suites)
	var err error
	ts.mockGoogleServer, err = testutils.GetSharedMockServers().GetGoogleServer(
		"test_google_client", "test_google_secret")
	if err != nil {
		ts.T().Fatalf("Failed to get shared Google server: %v", err)
	}

	// Reset the server to clear any state from previous test suites
	ts.mockGoogleServer.Reset()

	ts.mockGoogleServer.AddUser(&testutils.GoogleUserInfo{
		Sub:           "google-test-user-123",
		Email:         "testuser@gmail.com",
		EmailVerified: true,
		Name:          "Test User",
		GivenName:     "Test",
		FamilyName:    "User",
		Picture:       "https://example.com/picture.jpg",
		Locale:        "en",
	})

	// Use the IDP created by database scripts
	ts.idpID = "test-google-idp-id"

	// Create test organization unit for Google auth tests
	ouID, err := testutils.CreateOrganizationUnit(googleAuthTestOU)
	if err != nil {
		ts.T().Fatalf("Failed to create test organization unit during setup: %v", err)
	}
	googleAuthTestOU.ID = ouID

	// Create user schema
	googleUserSchema.OrganizationUnitId = ouID
	schemaID, err := testutils.CreateUserType(googleUserSchema)
	ts.Require().NoError(err, "Failed to create Google user schema")
	ts.userSchemaID = schemaID

	// Create user
	userAttributes := map[string]interface{}{
		"username":   "googleflowuser",
		"password":   "Test@1234",
		"sub":        "google-test-user-123",
		"email":      "testuser@gmail.com",
		"givenName":  "Test",
		"familyName": "User",
	}

	attributesJSON, err := json.Marshal(userAttributes)
	ts.Require().NoError(err)

	// Create user in the pre-configured OU from database scripts
	user := testutils.User{
		Type:             googleUserSchema.Name,
		OrganizationUnit: googleUserSchema.OrganizationUnitId,
		Attributes:       json.RawMessage(attributesJSON),
	}

	userID, err := testutils.CreateUser(user)
	ts.Require().NoError(err, "Failed to create test user")
	ts.userID = userID

	// Create test application for Google auth tests
	appID, err := testutils.CreateApplication(googleAuthTestApp)
	if err != nil {
		ts.T().Fatalf("Failed to create test application during setup: %v", err)
	}
	googleAuthTestAppID = appID
}

func (ts *GoogleAuthFlowTestSuite) TearDownSuite() {
	// Delete test application
	if googleAuthTestAppID != "" {
		if err := testutils.DeleteApplication(googleAuthTestAppID); err != nil {
			ts.T().Logf("Failed to delete test application during teardown: %v", err)
		}
	}

	// Clean up user
	if ts.userID != "" {
		_ = testutils.DeleteUser(ts.userID)
	}

	if ts.userSchemaID != "" {
		_ = testutils.DeleteUserType(ts.userSchemaID)
	}

	// Note: We don't stop the mock server here because it's shared across test suites.
	// The shared server will be cleaned up when the test process exits.
}

func (ts *GoogleAuthFlowTestSuite) TestGoogleAuthFlowInitiation() {
	// Initialize the flow by calling the flow execution API
	flowStep, err := initiateAuthFlow(googleAuthTestAppID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate Google authentication flow: %v", err)
	}

	// Verify flow status and type
	ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus, "Expected flow status to be INCOMPLETE")
	ts.Require().Equal("REDIRECTION", flowStep.Type, "Expected flow type to be REDIRECT")
	ts.Require().NotEmpty(flowStep.FlowID, "Flow ID should not be empty")

	// Validate redirect information
	ts.Require().NotEmpty(flowStep.Data, "Flow data should not be empty")
	ts.Require().NotEmpty(flowStep.Data.RedirectURL, "Redirect URL should not be empty")
	redirectURLStr := flowStep.Data.RedirectURL
	ts.Require().True(strings.HasPrefix(redirectURLStr, "http://localhost:8093/o/oauth2/v2/auth"),
		"Redirect URL should point to mock Google server")

	// Parse and validate the redirect URL
	redirectURL, err := url.Parse(redirectURLStr)
	ts.Require().NoError(err, "Should be able to parse the redirect URL")

	// Check required query parameters in the redirect URL
	queryParams := redirectURL.Query()
	ts.Require().NotEmpty(queryParams.Get("client_id"), "client_id should be present in redirect URL")
	ts.Require().NotEmpty(queryParams.Get("redirect_uri"), "redirect_uri should be present in redirect URL")
	ts.Require().NotEmpty(queryParams.Get("response_type"), "response_type should be present in redirect URL")
	ts.Require().Equal("code", queryParams.Get("response_type"), "response_type should be 'code'")

	scope := queryParams.Get("scope")
	ts.Require().NotEmpty(scope, "scope should be present in redirect URL")

	scopesPresent := strings.Contains(scope, "openid") &&
		strings.Contains(scope, "email") &&
		strings.Contains(scope, "profile")
	ts.Require().True(scopesPresent, "scope should include expected scopes")
}

func (ts *GoogleAuthFlowTestSuite) TestGoogleAuthFlowInvalidAppID() {
	errorResp, err := initiateAuthFlowWithError("invalid-google-app-id", nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate authentication flow with invalid app ID: %v", err)
	}

	ts.Require().Equal("FES-1003", errorResp.Code, "Expected error code for invalid app ID")
	ts.Require().Equal("Invalid request", errorResp.Message, "Expected error message for invalid request")
	ts.Require().Equal("Invalid app ID provided in the request", errorResp.Description,
		"Expected error description for invalid app ID")
}

func (ts *GoogleAuthFlowTestSuite) TestGoogleAuthFlowCompleteSuccess() {
	// Step 1: Initialize the flow by calling the flow execution API
	flowStep, err := initiateAuthFlow(googleAuthTestAppID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate Google authentication flow: %v", err)
	}

	// Verify flow status and type
	ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus, "Expected flow status to be INCOMPLETE")
	ts.Require().Equal("REDIRECTION", flowStep.Type, "Expected flow type to be REDIRECT")
	ts.Require().NotEmpty(flowStep.FlowID, "Flow ID should not be empty")

	flowID := flowStep.FlowID
	redirectURLStr := flowStep.Data.RedirectURL
	ts.Require().NotEmpty(redirectURLStr, "Redirect URL should not be empty")

	// Step 2: Simulate user authorization at Google (get authorization code)
	authCode, err := testutils.SimulateFederatedOAuthFlow(redirectURLStr)
	if err != nil {
		ts.T().Fatalf("Failed to simulate Google authorization: %v", err)
	}
	ts.Require().NotEmpty(authCode, "Authorization code should not be empty")

	// Step 3: Complete the flow with the authorization code
	inputs := map[string]string{
		"code": authCode,
	}

	completeFlowStep, err := completeAuthFlow(flowID, "", inputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete Google authentication flow: %v", err)
	}

	// Verify flow completion
	ts.Require().Equal("COMPLETE", completeFlowStep.FlowStatus, "Expected flow status to be COMPLETE")
	ts.Require().NotEmpty(completeFlowStep.Assertion, "Assertion token should be present")

	// Validate JWT assertion fields using common utility
	jwtClaims, err := testutils.ValidateJWTAssertionFields(
		completeFlowStep.Assertion,
		googleAuthTestAppID,
		googleUserSchema.Name,
		googleAuthTestOU.ID,
		googleAuthTestOU.Name,
		googleAuthTestOU.Handle,
	)
	ts.Require().NoError(err, "Failed to validate JWT assertion fields")
	ts.Require().NotNil(jwtClaims, "JWT claims should not be nil")
}

func (ts *GoogleAuthFlowTestSuite) TestGoogleAuthFlowCompleteWithInvalidCode() {
	// Step 1: Initialize the flow
	flowStep, err := initiateAuthFlow(googleAuthTestAppID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate Google authentication flow: %v", err)
	}

	flowID := flowStep.FlowID

	// Step 2: Try to complete with invalid authorization code
	inputs := map[string]string{
		"code": "invalid-auth-code-12345",
	}

	_, err = completeAuthFlow(flowID, "", inputs)
	ts.Require().Error(err, "Should fail with invalid authorization code")
}

func (ts *GoogleAuthFlowTestSuite) TestGoogleAuthFlowCompleteWithMissingCode() {
	// Step 1: Initialize the flow
	flowStep, err := initiateAuthFlow(googleAuthTestAppID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate Google authentication flow: %v", err)
	}

	flowID := flowStep.FlowID

	// Step 2: Try to complete without providing authorization code
	inputs := map[string]string{}

	// When required inputs are missing, the flow returns INCOMPLETE status (not an error)
	// and asks for the missing inputs again
	flowStep, err = completeAuthFlow(flowID, "", inputs)
	ts.Require().NoError(err, "Should not return error when inputs are missing")
	ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus,
		"Flow should remain INCOMPLETE when required inputs are missing")
	ts.Require().Equal("REDIRECTION", flowStep.Type, "Flow should still be REDIRECTION type")

	// Verify that code input is still required
	ts.Require().NotEmpty(flowStep.Data.Inputs, "Should still require inputs")
	hasCodeInput := false
	for _, input := range flowStep.Data.Inputs {
		if input.Name == "code" && input.Required {
			hasCodeInput = true
			break
		}
	}
	ts.Require().True(hasCodeInput, "Code input should still be required")
}

func (ts *GoogleAuthFlowTestSuite) TestGoogleAuthFlowMultipleUsersSuccess() {
	// This test verifies that the flow works correctly when the IDP is configured
	// with multiple users, and one of them authenticates successfully

	// Step 1: Initialize the flow
	flowStep, err := initiateAuthFlow(googleAuthTestAppID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate Google authentication flow: %v", err)
	}

	flowID := flowStep.FlowID
	redirectURLStr := flowStep.Data.RedirectURL

	// Step 2: Simulate user authorization at Google
	authCode, err := testutils.SimulateFederatedOAuthFlow(redirectURLStr)
	if err != nil {
		ts.T().Fatalf("Failed to simulate Google authorization: %v", err)
	}

	// Step 3: Complete the flow with the authorization code
	inputs := map[string]string{
		"code": authCode,
	}

	completeFlowStep, err := completeAuthFlow(flowID, "", inputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete Google authentication flow: %v", err)
	}

	// Verify flow completion
	ts.Require().Equal("COMPLETE", completeFlowStep.FlowStatus, "Expected flow status to be COMPLETE")
	ts.Require().NotEmpty(completeFlowStep.Assertion, "Assertion token should be present")

	// Validate JWT assertion fields using common utility
	jwtClaims, err := testutils.ValidateJWTAssertionFields(
		completeFlowStep.Assertion,
		googleAuthTestAppID,
		googleUserSchema.Name,
		googleAuthTestOU.ID,
		googleAuthTestOU.Name,
		googleAuthTestOU.Handle,
	)
	ts.Require().NoError(err, "Failed to validate JWT assertion fields")
	ts.Require().NotNil(jwtClaims, "JWT claims should not be nil")
}
