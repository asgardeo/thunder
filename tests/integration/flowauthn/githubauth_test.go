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
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

var (
	githubAuthTestApp = TestApplication{
		Name:                      "GitHub Auth Flow Test Application",
		Description:               "Application for testing GitHub authentication flows",
		IsRegistrationFlowEnabled: false,
		AuthFlowGraphID:           "auth_flow_config_github",
		RegistrationFlowGraphID:   "registration_flow_config_basic",
		ClientID:                  "github_auth_flow_test_client",
		ClientSecret:              "github_auth_flow_test_secret",
		RedirectURIs:              []string{"http://localhost:3000/callback"},
	}

	githubAuthTestOU = TestOrganizationUnit{
		Handle:      "github-auth-flow-test-ou",
		Name:        "GitHub Auth Flow Test Organization Unit",
		Description: "Organization unit for GitHub authentication flow testing",
		Parent:      nil,
	}
)

var (
	githubAuthTestAppID string
	githubAuthTestOUID  string
	githubAuthTestIDPID string
)

type GithubAuthFlowTestSuite struct {
	suite.Suite
}

func TestGithubAuthFlowTestSuite(t *testing.T) {
	suite.Run(t, new(GithubAuthFlowTestSuite))
}

func (ts *GithubAuthFlowTestSuite) SetupSuite() {
	// Create test organization unit for GitHub auth tests
	ouID, err := createOrganizationUnit(githubAuthTestOU)
	if err != nil {
		ts.T().Fatalf("Failed to create test organization unit during setup: %v", err)
	}
	githubAuthTestOUID = ouID

	// Create GitHub IDP for GitHub auth tests
	githubIDP := IDP{
		Name:        "Github",
		Description: "GitHub Identity Provider for authentication flow testing",
		Properties: []IDPProperty{
			{
				Name:     "client_id",
				Value:    "test_github_client",
				IsSecret: false,
			},
			{
				Name:     "client_secret",
				Value:    "test_github_secret",
				IsSecret: true,
			},
			{
				Name:     "redirect_uri",
				Value:    "https://localhost:3000/github/callback",
				IsSecret: false,
			},
			{
				Name:     "scopes",
				Value:    "user:email,read:user",
				IsSecret: false,
			},
		},
	}

	idpID, err := createIdp(githubIDP)
	if err != nil {
		ts.T().Fatalf("Failed to create GitHub IDP during setup: %v", err)
	}
	githubAuthTestIDPID = idpID

	// Create test application for GitHub auth tests
	appID, err := createApplication(githubAuthTestApp)
	if err != nil {
		ts.T().Fatalf("Failed to create test application during setup: %v", err)
	}
	githubAuthTestAppID = appID
}

func (ts *GithubAuthFlowTestSuite) TearDownSuite() {
	// Delete test application
	if githubAuthTestAppID != "" {
		if err := deleteApplication(githubAuthTestAppID); err != nil {
			ts.T().Logf("Failed to delete test application during teardown: %v", err)
		}
	}

	// Delete GitHub IDP
	if githubAuthTestIDPID != "" {
		if err := deleteIdp(githubAuthTestIDPID); err != nil {
			ts.T().Logf("Failed to delete GitHub IDP during teardown: %v", err)
		}
	}

	// Delete test organization unit
	if githubAuthTestOUID != "" {
		if err := deleteOrganizationUnit(githubAuthTestOUID); err != nil {
			ts.T().Logf("Failed to delete test organization unit during teardown: %v", err)
		}
	}
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
	ts.Require().True(strings.HasPrefix(redirectURLStr, "https://github.com/login/oauth/authorize"),
		"Redirect URL should point to GitHub authentication")

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
