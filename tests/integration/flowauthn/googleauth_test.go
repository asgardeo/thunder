/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
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

type GoogleAuthFlowTestSuite struct {
	suite.Suite
}

func TestGoogleAuthFlowTestSuite(t *testing.T) {
	suite.Run(t, new(GoogleAuthFlowTestSuite))
}

func (ts *GoogleAuthFlowTestSuite) SetupSuite() {
	err := updateAppConfig(appID, "auth_flow_config_google")
	if err != nil {
		ts.T().Fatalf("Failed to update system app for Google auth: %v", err)
	}
}

func (ts *GoogleAuthFlowTestSuite) TearDownSuite() {
	err := updateAppConfig(appID, "auth_flow_config_basic")
	if err != nil {
		ts.T().Fatalf("Failed to reset system app to basic auth: %v", err)
	}
}

func (ts *GoogleAuthFlowTestSuite) TestGoogleAuthFlowInitiation() {
	// Initialize the flow by calling the flow execution API
	flowStep, err := initiateAuthFlow(appID, nil)
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
