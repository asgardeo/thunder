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
	googleAuthTestApp = TestApplication{
		Name:                      "Google Auth Flow Test Application",
		Description:               "Application for testing Google authentication flows",
		IsRegistrationFlowEnabled: false,
		AuthFlowGraphID:           "auth_flow_config_google",
		RegistrationFlowGraphID:   "registration_flow_config_basic",
		ClientID:                  "google_auth_flow_test_client",
		ClientSecret:              "google_auth_flow_test_secret",
		RedirectURIs:              []string{"http://localhost:3000/callback"},
	}

	googleAuthTestOU = TestOrganizationUnit{
		Handle:      "google-auth-flow-test-ou",
		Name:        "Google Auth Flow Test Organization Unit",
		Description: "Organization unit for Google authentication flow testing",
		Parent:      nil,
	}
)

var (
	googleAuthTestAppID string
	googleAuthTestOUID  string
	googleAuthTestIDPID string
)

type GoogleAuthFlowTestSuite struct {
	suite.Suite
}

func TestGoogleAuthFlowTestSuite(t *testing.T) {
	suite.Run(t, new(GoogleAuthFlowTestSuite))
}

func (ts *GoogleAuthFlowTestSuite) SetupSuite() {
	// Create test organization unit for Google auth tests
	ouID, err := createOrganizationUnit(googleAuthTestOU)
	if err != nil {
		ts.T().Fatalf("Failed to create test organization unit during setup: %v", err)
	}
	googleAuthTestOUID = ouID

	// Create Google IDP for Google auth tests
	googleIDP := IDP{
		Name:        "Google",
		Description: "Google Identity Provider for authentication flow testing",
		Properties: []IDPProperty{
			{
				Name:     "client_id",
				Value:    "test_google_client",
				IsSecret: false,
			},
			{
				Name:     "client_secret",
				Value:    "test_google_secret",
				IsSecret: true,
			},
			{
				Name:     "redirect_uri",
				Value:    "https://localhost:3000/google/callback",
				IsSecret: false,
			},
			{
				Name:     "scopes",
				Value:    "openid,email,profile",
				IsSecret: false,
			},
		},
	}

	idpID, err := createIdp(googleIDP)
	if err != nil {
		ts.T().Fatalf("Failed to create Google IDP during setup: %v", err)
	}
	googleAuthTestIDPID = idpID

	// Create test application for Google auth tests
	appID, err := createApplication(googleAuthTestApp)
	if err != nil {
		ts.T().Fatalf("Failed to create test application during setup: %v", err)
	}
	googleAuthTestAppID = appID
}

func (ts *GoogleAuthFlowTestSuite) TearDownSuite() {
	// Delete test application
	if googleAuthTestAppID != "" {
		if err := deleteApplication(googleAuthTestAppID); err != nil {
			ts.T().Logf("Failed to delete test application during teardown: %v", err)
		}
	}

	// Delete Google IDP
	if googleAuthTestIDPID != "" {
		if err := deleteIdp(googleAuthTestIDPID); err != nil {
			ts.T().Logf("Failed to delete Google IDP during teardown: %v", err)
		}
	}

	// Delete test organization unit
	if googleAuthTestOUID != "" {
		if err := deleteOrganizationUnit(googleAuthTestOUID); err != nil {
			ts.T().Logf("Failed to delete test organization unit during teardown: %v", err)
		}
	}
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
	ts.Require().True(strings.HasPrefix(redirectURLStr, "https://accounts.google.com/o/oauth2/v2/auth"),
		"Redirect URL should point to Google authentication")

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
