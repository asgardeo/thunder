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
	githubAuthTestApp = testutils.Application{
		Name:                      "GitHub Auth Flow Test Application",
		Description:               "Application for testing GitHub authentication flows",
		IsRegistrationFlowEnabled: false,
		AuthFlowGraphID:           "auth_flow_config_github",
		RegistrationFlowGraphID:   "registration_flow_config_basic",
		ClientID:                  "github_auth_flow_test_client",
		ClientSecret:              "github_auth_flow_test_secret",
		RedirectURIs:              []string{"http://localhost:3000/callback"},
	}
)

var (
	githubAuthTestAppID string
	githubAuthTestOU = testutils.OrganizationUnit{
		Handle:      "github-auth-flow-test-ou",
		Name:        "GitHub Auth Flow Test OU",
		Description: "Organization unit for GitHub authentication flow tests",
	}
)



var githubUserSchema = testutils.UserSchema{
	Name: "github_flow_user",
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

type GithubAuthFlowTestSuite struct {
	suite.Suite
	mockGithubServer *testutils.MockGithubOAuthServer
	idpID            string
	userID           string
	userSchemaID     string
}

func TestGithubAuthFlowTestSuite(t *testing.T) {
	suite.Run(t, new(GithubAuthFlowTestSuite))
}

func (ts *GithubAuthFlowTestSuite) SetupSuite() {
	// Get shared GitHub OAuth server (started once for all test suites)
	var err error
	ts.mockGithubServer, err = testutils.GetSharedMockServers().GetGithubServer(
		"test_github_client", "test_github_secret")
	if err != nil {
		ts.T().Fatalf("Failed to get shared GitHub server: %v", err)
	}

	// Reset the server to clear any state from previous test suites
	ts.mockGithubServer.Reset()

	email := "testuser@github.com"
	ts.mockGithubServer.AddUser(&testutils.GithubUserInfo{
		Login:     "testuser",
		ID:        12345,
		NodeID:    "MDQ6VXNlcjEyMzQ1",
		Email:     &email,
		Name:      "Test User",
		AvatarURL: "https://avatars.githubusercontent.com/u/12345",
		Type:      "User",
		CreatedAt: "2020-01-01T00:00:00Z",
		UpdatedAt: "2024-01-01T00:00:00Z",
	}, []*testutils.GithubEmail{
		{
			Email:    email,
			Primary:  true,
			Verified: true,
		},
	})

	// Use the IDP created by database scripts
	ts.idpID = "test-github-idp-id"

	// Create test organization unit for GitHub auth tests
	ouID, err := testutils.CreateOrganizationUnit(githubAuthTestOU)
	if err != nil {
		ts.T().Fatalf("Failed to create test organization unit during setup: %v", err)
	}
	githubAuthTestOU.ID = ouID

	// Create user schema
	githubUserSchema.OrganizationUnitId = ouID
	schemaID, err := testutils.CreateUserType(githubUserSchema)
	ts.Require().NoError(err, "Failed to create GitHub user schema")
	ts.userSchemaID = schemaID

	// Create user
	userAttributes := map[string]interface{}{
		"username":   "githubflowuser",
		"password":   "Test@1234",
		"sub":        "12345",
		"email":      "testuser@github.com",
		"givenName":  "Test",
		"familyName": "User",
	}

	attributesJSON, err := json.Marshal(userAttributes)
	ts.Require().NoError(err)

	// Create user in the pre-configured OU from database scripts
	user := testutils.User{
		Type:             githubUserSchema.Name,
		OrganizationUnit: githubUserSchema.OrganizationUnitId,
		Attributes:       json.RawMessage(attributesJSON),
	}

	userID, err := testutils.CreateUser(user)
	ts.Require().NoError(err, "Failed to create test user")
	ts.userID = userID

	// Create test application for GitHub auth tests
	appID, err := testutils.CreateApplication(githubAuthTestApp)
	if err != nil {
		ts.T().Fatalf("Failed to create test application during setup: %v", err)
	}
	githubAuthTestAppID = appID
}

func (ts *GithubAuthFlowTestSuite) TearDownSuite() {
	// Delete test application
	if githubAuthTestAppID != "" {
		if err := testutils.DeleteApplication(githubAuthTestAppID); err != nil {
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

func (ts *GithubAuthFlowTestSuite) TestGithubAuthFlowInitiation() {
	// Initialize the flow by calling the flow execution API
	flowStep, err := initiateAuthFlow(githubAuthTestAppID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate GitHub authentication flow: %v", err)
	}

	// Verify flow status and type
	ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus, "Expected flow status to be INCOMPLETE")
	ts.Require().Equal("REDIRECTION", flowStep.Type, "Expected flow type to be REDIRECT")
	ts.Require().NotEmpty(flowStep.FlowID, "Flow ID should not be empty")

	// Validate redirect information
	ts.Require().NotEmpty(flowStep.Data, "Flow data should not be empty")
	ts.Require().NotEmpty(flowStep.Data.RedirectURL, "Redirect URL should not be empty")
	redirectURLStr := flowStep.Data.RedirectURL
	ts.Require().True(strings.HasPrefix(redirectURLStr, "http://localhost:8092/login/oauth/authorize"),
		"Redirect URL should point to mock GitHub server")

	// Parse and validate the redirect URL
	redirectURL, err := url.Parse(redirectURLStr)
	ts.Require().NoError(err, "Should be able to parse the redirect URL")

	// Check required query parameters in the redirect URL
	queryParams := redirectURL.Query()
	ts.Require().NotEmpty(queryParams.Get("client_id"), "client_id should be present in redirect URL")
	ts.Require().NotEmpty(queryParams.Get("redirect_uri"), "redirect_uri should be present in redirect URL")

	scope := queryParams.Get("scope")
	ts.Require().NotEmpty(scope, "scope should be present in redirect URL")

	scopesPresent := strings.Contains(scope, "read:user") &&
		strings.Contains(scope, "user:email")
	ts.Require().True(scopesPresent, "scope should include expected scopes")
}

func (ts *GithubAuthFlowTestSuite) TestGithubAuthFlowInvalidAppID() {
	errorResp, err := initiateAuthFlowWithError("invalid-github-app-id", nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate authentication flow with invalid app ID: %v", err)
	}

	ts.Require().Equal("FES-1003", errorResp.Code, "Expected error code for invalid app ID")
	ts.Require().Equal("Invalid request", errorResp.Message, "Expected error message for invalid request")
	ts.Require().Equal("Invalid app ID provided in the request", errorResp.Description,
		"Expected error description for invalid app ID")
}

func (ts *GithubAuthFlowTestSuite) TestGithubAuthFlowCompleteSuccess() {
	// Step 1: Initialize the flow by calling the flow execution API
	flowStep, err := initiateAuthFlow(githubAuthTestAppID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate GitHub authentication flow: %v", err)
	}

	// Verify flow status and type
	ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus, "Expected flow status to be INCOMPLETE")
	ts.Require().Equal("REDIRECTION", flowStep.Type, "Expected flow type to be REDIRECT")
	ts.Require().NotEmpty(flowStep.FlowID, "Flow ID should not be empty")

	flowID := flowStep.FlowID
	redirectURLStr := flowStep.Data.RedirectURL
	ts.Require().NotEmpty(redirectURLStr, "Redirect URL should not be empty")

	// Step 2: Simulate user authorization at GitHub (get authorization code)
	authCode, err := testutils.SimulateFederatedOAuthFlow(redirectURLStr)
	if err != nil {
		ts.T().Fatalf("Failed to simulate GitHub authorization: %v", err)
	}
	ts.Require().NotEmpty(authCode, "Authorization code should not be empty")

	// Step 3: Complete the flow with the authorization code
	inputs := map[string]string{
		"code": authCode,
	}

	completeFlowStep, err := completeAuthFlow(flowID, "", inputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete GitHub authentication flow: %v", err)
	}

	// Verify flow completion
	ts.Require().Equal("COMPLETE", completeFlowStep.FlowStatus, "Expected flow status to be COMPLETE")
	ts.Require().NotEmpty(completeFlowStep.Assertion, "Assertion token should be present")

	// Validate JWT assertion fields using common utility
	jwtClaims, err := testutils.ValidateJWTAssertionFields(
		completeFlowStep.Assertion,
		githubAuthTestAppID,
		githubUserSchema.Name,
		githubAuthTestOU.ID,
		githubAuthTestOU.Name,
		githubAuthTestOU.Handle,
	)
	ts.Require().NoError(err, "Failed to validate JWT assertion fields")
	ts.Require().NotNil(jwtClaims, "JWT claims should not be nil")
}

func (ts *GithubAuthFlowTestSuite) TestGithubAuthFlowCompleteWithInvalidCode() {
	// Step 1: Initialize the flow
	flowStep, err := initiateAuthFlow(githubAuthTestAppID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate GitHub authentication flow: %v", err)
	}

	flowID := flowStep.FlowID

	// Step 2: Try to complete with invalid authorization code
	inputs := map[string]string{
		"code": "invalid-auth-code-12345",
	}

	_, err = completeAuthFlow(flowID, "", inputs)
	ts.Require().Error(err, "Should fail with invalid authorization code")
}

func (ts *GithubAuthFlowTestSuite) TestGithubAuthFlowCompleteWithMissingCode() {
	// Step 1: Initialize the flow
	flowStep, err := initiateAuthFlow(githubAuthTestAppID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate GitHub authentication flow: %v", err)
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
