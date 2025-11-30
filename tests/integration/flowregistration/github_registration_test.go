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

package flowregistration

import (
	"encoding/json"
	"net/url"
	"strings"
	"testing"

	"github.com/asgardeo/thunder/tests/integration/testutils"
	"github.com/stretchr/testify/suite"
)

var (
	githubRegTestOU = testutils.OrganizationUnit{
		Handle:      "github-reg-flow-test-ou",
		Name:        "GitHub Registration Flow Test Organization Unit",
		Description: "Organization unit for GitHub registration flow testing",
		Parent:      nil,
	}

	githubRegUserSchema = testutils.UserSchema{
		Name: "github_reg_flow_user",
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

	githubRegTestApp = testutils.Application{
		Name:                      "GitHub Registration Flow Test Application",
		Description:               "Application for testing GitHub registration flows",
		IsRegistrationFlowEnabled: true,
		AuthFlowGraphID:           "auth_flow_config_github",
		RegistrationFlowGraphID:   "registration_flow_config_github",
		ClientID:                  "github_reg_flow_test_client",
		ClientSecret:              "github_reg_flow_test_secret",
		RedirectURIs:              []string{"http://localhost:3000/callback"},
		AllowedUserTypes:          []string{githubRegUserSchema.Name},
	}
)

var (
	githubRegTestAppID string
	githubRegTestOUID  string
)



type GithubRegistrationFlowTestSuite struct {
	suite.Suite
	mockGithubServer *testutils.MockGithubOAuthServer
	idpID            string
	userSchemaID     string
	config           *TestSuiteConfig
}

func TestGithubRegistrationFlowTestSuite(t *testing.T) {
	suite.Run(t, new(GithubRegistrationFlowTestSuite))
}

func (ts *GithubRegistrationFlowTestSuite) SetupSuite() {
	ts.config = &TestSuiteConfig{}

	// Get shared GitHub OAuth server (started once for all test suites)
	var err error
	ts.mockGithubServer, err = testutils.GetSharedMockServers().GetGithubServer(
		"test_github_client", "test_github_secret")
	if err != nil {
		ts.T().Fatalf("Failed to get shared GitHub server: %v", err)
	}

	email := "reguser@github.com"
	ts.mockGithubServer.AddUser(&testutils.GithubUserInfo{
		Login:     "reguser",
		ID:        67890,
		NodeID:    "MDQ6VXNlcjY3ODkw",
		Email:     &email,
		Name:      "Registration User",
		AvatarURL: "https://avatars.githubusercontent.com/u/67890",
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

	// Create test organization unit for GitHub registration tests
	ouID, err := testutils.CreateOrganizationUnit(githubRegTestOU)
	if err != nil {
		ts.T().Fatalf("Failed to create test organization unit during setup: %v", err)
	}
	githubRegTestOUID = ouID

	// Create user schema
	githubRegUserSchema.OrganizationUnitId = githubRegTestOUID
	githubRegUserSchema.AllowSelfRegistration = true
	schemaID, err := testutils.CreateUserType(githubRegUserSchema)
	ts.Require().NoError(err, "Failed to create GitHub user schema")
	ts.userSchemaID = schemaID

	// Create test application
	appID, err := testutils.CreateApplication(githubRegTestApp)
	if err != nil {
		ts.T().Fatalf("Failed to create test application during setup: %v", err)
	}
	githubRegTestAppID = appID
}

func (ts *GithubRegistrationFlowTestSuite) TearDownTest() {
	// Clean up users created during each test
	if len(ts.config.CreatedUserIDs) > 0 {
		if err := testutils.CleanupUsers(ts.config.CreatedUserIDs); err != nil {
			ts.T().Logf("Failed to cleanup users after test: %v", err)
		}
		// Reset the list for the next test
		ts.config.CreatedUserIDs = []string{}
	}
}

func (ts *GithubRegistrationFlowTestSuite) TearDownSuite() {
	// Delete test application
	if githubRegTestAppID != "" {
		if err := testutils.DeleteApplication(githubRegTestAppID); err != nil {
			ts.T().Logf("Failed to delete test application during teardown: %v", err)
		}
	}

	// Delete test organization unit
	if githubRegTestOUID != "" {
		if err := testutils.DeleteOrganizationUnit(githubRegTestOUID); err != nil {
			ts.T().Logf("Failed to delete test organization unit during teardown: %v", err)
		}
	}

	// Clean up any remaining users
	if len(ts.config.CreatedUserIDs) > 0 {
		if err := testutils.CleanupUsers(ts.config.CreatedUserIDs); err != nil {
			ts.T().Logf("Failed to cleanup users during teardown: %v", err)
		}
	}

	if ts.userSchemaID != "" {
		_ = testutils.DeleteUserType(ts.userSchemaID)
	}

	// Note: We don't stop the mock server here because it's shared across test suites.
	// The shared server will be cleaned up when the test process exits.
}

func (ts *GithubRegistrationFlowTestSuite) TestGithubRegistrationFlowInitiation() {
	// Initialize the flow by calling the flow execution API
	flowStep, err := initiateRegistrationFlow(githubRegTestAppID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate GitHub registration flow: %v", err)
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

func (ts *GithubRegistrationFlowTestSuite) TestGithubRegistrationFlowCompleteSuccess() {
	// Step 1: Initialize the flow by calling the flow execution API
	flowStep, err := initiateRegistrationFlow(githubRegTestAppID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate GitHub registration flow: %v", err)
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

	completeFlowStep, err := completeRegistrationFlow(flowID, "", inputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete GitHub registration flow: %v", err)
	}

	// Verify flow completion
	ts.Require().Equal("COMPLETE", completeFlowStep.FlowStatus, "Expected flow status to be COMPLETE")
	ts.Require().NotEmpty(completeFlowStep.Assertion, "Assertion token should be present")

	// Verify the assertion token contains expected information
	ts.Require().Contains(completeFlowStep.Assertion, ".", "Assertion should be a JWT token")

	// Decode and validate JWT claims
	jwtClaims, err := testutils.DecodeJWT(completeFlowStep.Assertion)
	ts.Require().NoError(err, "Failed to decode JWT assertion")
	ts.Require().NotNil(jwtClaims, "JWT claims should not be nil")

	// Validate JWT contains expected user type and OU ID
	ts.Require().Equal(githubRegUserSchema.Name, jwtClaims.UserType, "Expected userType to match created schema")
	ts.Require().NotEmpty(jwtClaims.OuID, "Expected ouId to be present")
	ts.Require().Equal(githubRegTestAppID, jwtClaims.Aud, "Expected aud to match the application ID")
	ts.Require().NotEmpty(jwtClaims.Sub, "JWT subject should not be empty")

	// Verify the user was created by searching via the user API
	user, err := testutils.FindUserByAttribute("sub", "67890")
	if err != nil {
		ts.T().Fatalf("Failed to retrieve user by sub: %v", err)
	}
	ts.Require().NotNil(user, "User should be found in user list after registration")

	// Store the created user for cleanup
	if user != nil {
		ts.config.CreatedUserIDs = append(ts.config.CreatedUserIDs, user.ID)

		// Verify user attributes
		var attributes map[string]interface{}
		err = json.Unmarshal(user.Attributes, &attributes)
		ts.Require().NoError(err, "Should be able to unmarshal user attributes")
		ts.Require().Equal("67890", attributes["sub"], "User sub should match")
	}
}

func (ts *GithubRegistrationFlowTestSuite) TestGithubRegistrationFlowCompleteWithInvalidCode() {
	// Step 1: Initialize the flow
	flowStep, err := initiateRegistrationFlow(githubRegTestAppID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate GitHub registration flow: %v", err)
	}

	flowID := flowStep.FlowID

	// Step 2: Try to complete with invalid authorization code
	inputs := map[string]string{
		"code": "invalid-reg-auth-code-12345",
	}

	_, err = completeRegistrationFlow(flowID, "", inputs)
	ts.Require().Error(err, "Should fail with invalid authorization code")
}

func (ts *GithubRegistrationFlowTestSuite) TestGithubRegistrationFlowCompleteWithMissingCode() {
	// Step 1: Initialize the flow
	flowStep, err := initiateRegistrationFlow(githubRegTestAppID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate GitHub registration flow: %v", err)
	}

	flowID := flowStep.FlowID

	// Step 2: Try to complete without providing authorization code
	inputs := map[string]string{}

	// When required inputs are missing, the flow returns INCOMPLETE status (not an error)
	// and asks for the missing inputs again
	flowStep, err = completeRegistrationFlow(flowID, "", inputs)
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

func (ts *GithubRegistrationFlowTestSuite) TestGithubRegistrationFlowDuplicateUser() {
	// Step 1: First, create a user through registration
	flowStep, err := initiateRegistrationFlow(githubRegTestAppID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate first GitHub registration flow: %v", err)
	}

	redirectURLStr := flowStep.Data.RedirectURL
	authCode, err := testutils.SimulateFederatedOAuthFlow(redirectURLStr)
	if err != nil {
		ts.T().Fatalf("Failed to simulate first GitHub authorization: %v", err)
	}

	inputs := map[string]string{
		"code": authCode,
	}

	completeFlowStep, err := completeRegistrationFlow(flowStep.FlowID, "", inputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete first GitHub registration flow: %v", err)
	}

	ts.Require().Equal("COMPLETE", completeFlowStep.FlowStatus, "First registration should complete successfully")

	// Store created user for cleanup
	user, err := testutils.FindUserByAttribute("sub", "67890")
	if err == nil && user != nil {
		ts.config.CreatedUserIDs = append(ts.config.CreatedUserIDs, user.ID)
	}

	// Step 2: Try to register again with the same GitHub user
	flowStep2, err := initiateRegistrationFlow(githubRegTestAppID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate second GitHub registration flow: %v", err)
	}

	redirectURLStr2 := flowStep2.Data.RedirectURL
	authCode2, err := testutils.SimulateFederatedOAuthFlow(redirectURLStr2)
	if err != nil {
		ts.T().Fatalf("Failed to simulate second GitHub authorization: %v", err)
	}

	inputs2 := map[string]string{
		"code": authCode2,
	}

	completeFlowStep2, err := completeRegistrationFlow(flowStep2.FlowID, "", inputs2)
	if err != nil {
		ts.T().Fatalf("Failed to complete second GitHub registration flow: %v", err)
	}

	// Step 3: Verify registration failure due to duplicate user
	ts.Require().Equal("ERROR", completeFlowStep2.FlowStatus, "Expected flow status to be ERROR for duplicate user")
	ts.Require().Empty(completeFlowStep2.Assertion, "No JWT assertion should be returned for failed registration")
	ts.Require().NotEmpty(completeFlowStep2.FailureReason, "Failure reason should be provided for duplicate user")
}
