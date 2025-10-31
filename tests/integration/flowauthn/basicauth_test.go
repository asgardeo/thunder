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
	"testing"

	"github.com/asgardeo/thunder/tests/integration/testutils"
	"github.com/stretchr/testify/suite"
)

var (
	testApp = testutils.Application{
		Name:                      "Flow Test Application",
		Description:               "Application for testing authentication flows",
		IsRegistrationFlowEnabled: false,
		AuthFlowGraphID:           "auth_flow_config_basic",
		RegistrationFlowGraphID:   "registration_flow_config_basic",
		ClientID:                  "flow_test_client",
		ClientSecret:              "flow_test_secret",
		RedirectURIs:              []string{"http://localhost:3000/callback"},
	}

	testOU = testutils.OrganizationUnit{
		Handle:      "flow-test-ou",
		Name:        "Flow Test Organization Unit",
		Description: "Organization unit for flow testing",
		Parent:      nil,
	}

	testUser = testutils.User{
		Type: "person",
		Attributes: json.RawMessage(`{
			"username": "testuser",
			"password": "testpassword",
			"email": "test@example.com",
			"firstName": "Test",
			"lastName": "User"
		}`),
	}
)

var (
	testAppID string
	testOUID  string
)

type BasicAuthFlowTestSuite struct {
	suite.Suite
	config *TestSuiteConfig
}

func TestBasicAuthFlowTestSuite(t *testing.T) {
	suite.Run(t, new(BasicAuthFlowTestSuite))
}

func (ts *BasicAuthFlowTestSuite) SetupSuite() {
	// Initialize config
	ts.config = &TestSuiteConfig{}

	// Create test organization unit
	ouID, err := testutils.CreateOrganizationUnit(testOU)
	if err != nil {
		ts.T().Fatalf("Failed to create test organization unit during setup: %v", err)
	}
	testOUID = ouID

	// Create test application
	appID, err := testutils.CreateApplication(testApp)
	if err != nil {
		ts.T().Fatalf("Failed to create test application during setup: %v", err)
	}
	testAppID = appID

	// Create test user with the created OU
	testUser := testUser
	testUser.OrganizationUnit = testOUID
	userIDs, err := testutils.CreateMultipleUsers(testUser)
	if err != nil {
		ts.T().Fatalf("Failed to create test user during setup: %v", err)
	}
	ts.config.CreatedUserIDs = userIDs
}

func (ts *BasicAuthFlowTestSuite) TearDownSuite() {
	// Delete all created users
	if err := testutils.CleanupUsers(ts.config.CreatedUserIDs); err != nil {
		ts.T().Logf("Failed to cleanup users during teardown: %v", err)
	}

	// Delete test application
	if testAppID != "" {
		if err := testutils.DeleteApplication(testAppID); err != nil {
			ts.T().Logf("Failed to delete test application during teardown: %v", err)
		}
	}

	// Delete test organization unit
	if testOUID != "" {
		if err := testutils.DeleteOrganizationUnit(testOUID); err != nil {
			ts.T().Logf("Failed to delete test organization unit during teardown: %v", err)
		}
	}

}

func (ts *BasicAuthFlowTestSuite) TestBasicAuthFlowSuccess() {
	// Step 1: Initialize the flow by calling the flow execution API
	flowStep, err := initiateAuthFlow(testAppID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate authentication flow: %v", err)
	}

	ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus, "Expected flow status to be INCOMPLETE")
	ts.Require().Equal("VIEW", flowStep.Type, "Expected flow type to be VIEW")
	ts.Require().NotEmpty(flowStep.FlowID, "Flow ID should not be empty")

	// Validate that the required inputs are returned
	ts.Require().NotEmpty(flowStep.Data, "Flow data should not be empty")
	ts.Require().NotEmpty(flowStep.Data.Inputs, "Flow should require inputs")

	// Verify username and password are required inputs using utility function
	ts.Require().True(ValidateRequiredInputs(flowStep.Data.Inputs, []string{"username", "password"}),
		"Username and password inputs should be required")
	ts.Require().True(HasInput(flowStep.Data.Inputs, "username"), "Username input should be present")
	ts.Require().True(HasInput(flowStep.Data.Inputs, "password"), "Password input should be present")

	// Step 2: Continue the flow with valid credentials
	var userAttrs map[string]interface{}
	err = json.Unmarshal(testUser.Attributes, &userAttrs)
	ts.Require().NoError(err, "Failed to unmarshal user attributes")

	inputs := map[string]string{
		"username": userAttrs["username"].(string),
		"password": userAttrs["password"].(string),
	}

	completeFlowStep, err := completeAuthFlow(flowStep.FlowID, "", inputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete authentication flow: %v", err)
	}

	// Verify successful authentication
	ts.Require().Equal("COMPLETE", completeFlowStep.FlowStatus, "Expected flow status to be COMPLETE")
	ts.Require().NotEmpty(completeFlowStep.Assertion,
		"JWT assertion should be returned after successful authentication")
	ts.Require().Empty(completeFlowStep.FailureReason, "Failure reason should be empty for successful authentication")

	// Decode and validate JWT claims
	jwtClaims, err := testutils.DecodeJWT(completeFlowStep.Assertion)
	ts.Require().NoError(err, "Failed to decode JWT assertion")
	ts.Require().NotNil(jwtClaims, "JWT claims should not be nil")

	// Validate JWT contains expected user type and OU ID
	ts.Require().Equal("person", jwtClaims.UserType, "Expected userType to be 'person'")
	ts.Require().Equal(testOUID, jwtClaims.OuID, "Expected ouId to match the created organization unit")
	ts.Require().Equal(testAppID, jwtClaims.Aud, "Expected aud to match the application ID")
	ts.Require().NotEmpty(jwtClaims.Sub, "JWT subject should not be empty")
}

func (ts *BasicAuthFlowTestSuite) TestBasicAuthFlowSuccessWithSingleRequest() {
	// Step 1: Initialize the flow by calling the flow execution API with user credentials
	var userAttrs map[string]interface{}
	err := json.Unmarshal(testUser.Attributes, &userAttrs)
	ts.Require().NoError(err, "Failed to unmarshal user attributes")

	inputs := map[string]string{
		"username": userAttrs["username"].(string),
		"password": userAttrs["password"].(string),
	}

	flowStep, err := initiateAuthFlow(testAppID, inputs)
	if err != nil {
		ts.T().Fatalf("Failed to initiate authentication flow: %v", err)
	}

	// Verify successful authentication
	ts.Require().Equal("COMPLETE", flowStep.FlowStatus, "Expected flow status to be COMPLETE")
	ts.Require().Empty(flowStep.Data, "Flow should not require additional data after successful authentication")
	ts.Require().NotEmpty(flowStep.Assertion,
		"JWT assertion should be returned after successful authentication")
	ts.Require().Empty(flowStep.FailureReason, "Failure reason should be empty for successful authentication")

	// Decode and validate JWT claims
	jwtClaims, err := testutils.DecodeJWT(flowStep.Assertion)
	ts.Require().NoError(err, "Failed to decode JWT assertion")
	ts.Require().NotNil(jwtClaims, "JWT claims should not be nil")

	// Validate JWT contains expected user type and OU ID
	ts.Require().Equal("person", jwtClaims.UserType, "Expected userType to be 'person'")
	ts.Require().Equal(testOUID, jwtClaims.OuID, "Expected ouId to match the created organization unit")
	ts.Require().Equal(testAppID, jwtClaims.Aud, "Expected aud to match the application ID")
	ts.Require().NotEmpty(jwtClaims.Sub, "JWT subject should not be empty")
}

func (ts *BasicAuthFlowTestSuite) TestBasicAuthFlowWithTwoStepInput() {
	// Step 1: Initialize the flow
	flowStep, err := initiateAuthFlow(testAppID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate authentication flow: %v", err)
	}

	ts.Require().NotEmpty(flowStep.FlowID, "Flow ID should not be empty")

	var userAttrs map[string]interface{}
	err = json.Unmarshal(testUser.Attributes, &userAttrs)
	ts.Require().NoError(err, "Failed to unmarshal user attributes")

	// Step 2: Continue with missing password
	inputs := map[string]string{
		"username": userAttrs["username"].(string),
	}

	intermediateFlowStep, err := completeAuthFlow(flowStep.FlowID, "", inputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete authentication flow with missing credentials: %v", err)
	}

	ts.Require().Equal("INCOMPLETE", intermediateFlowStep.FlowStatus, "Expected flow status to be INCOMPLETE")
	ts.Require().Equal("VIEW", intermediateFlowStep.Type, "Expected flow type to be VIEW")
	ts.Require().NotEmpty(intermediateFlowStep.FlowID, "Flow ID should not be empty")

	// Validate that the required inputs are returned
	ts.Require().NotEmpty(intermediateFlowStep.Data, "Flow data should not be empty")
	ts.Require().NotEmpty(intermediateFlowStep.Data.Inputs, "Flow should require inputs")

	// Verify password is required input using utility function
	ts.Require().True(HasInput(flowStep.Data.Inputs, "password"), "Password input should be required")

	// Step 3: Continue the flow with the password
	inputs = map[string]string{
		"password": userAttrs["password"].(string),
	}

	completeFlowStep, err := completeAuthFlow(flowStep.FlowID, "", inputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete authentication flow: %v", err)
	}

	// Verify successful authentication
	ts.Require().Equal("COMPLETE", completeFlowStep.FlowStatus, "Expected flow status to be COMPLETE")
	ts.Require().NotEmpty(completeFlowStep.Assertion,
		"JWT assertion should be returned after successful authentication")
	ts.Require().Empty(completeFlowStep.FailureReason, "Failure reason should be empty for successful authentication")

	// Decode and validate JWT claims
	jwtClaims, err := testutils.DecodeJWT(completeFlowStep.Assertion)
	ts.Require().NoError(err, "Failed to decode JWT assertion")
	ts.Require().NotNil(jwtClaims, "JWT claims should not be nil")

	// Validate JWT contains expected user type and OU ID
	ts.Require().Equal("person", jwtClaims.UserType, "Expected userType to be 'person'")
	ts.Require().Equal(testOUID, jwtClaims.OuID, "Expected ouId to match the created organization unit")
	ts.Require().Equal(testAppID, jwtClaims.Aud, "Expected aud to match the application ID")
	ts.Require().NotEmpty(jwtClaims.Sub, "JWT subject should not be empty")
}

func (ts *BasicAuthFlowTestSuite) TestBasicAuthFlowInvalidCredentials() {
	// Step 1: Initialize the flow
	flowStep, err := initiateAuthFlow(testAppID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate authentication flow: %v", err)
	}

	ts.Require().NotEmpty(flowStep.FlowID, "Flow ID should not be empty")

	// Step 2: Continue with invalid credentials
	inputs := map[string]string{
		"username": "invalid_user",
		"password": "wrong_password",
	}

	completeFlowStep, err := completeAuthFlow(flowStep.FlowID, "", inputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete authentication flow with invalid credentials: %v", err)
	}

	// Verify authentication failure
	ts.Require().Equal("ERROR", completeFlowStep.FlowStatus, "Expected flow status to be ERROR")
	ts.Require().Empty(completeFlowStep.Assertion, "No JWT assertion should be returned for failed authentication")
	ts.Require().NotEmpty(completeFlowStep.FailureReason, "Failure reason should be provided for invalid credentials")
}

func (ts *BasicAuthFlowTestSuite) TestBasicAuthFlowInvalidAppID() {
	// Try to initialize the flow with an invalid app ID
	errorResp, err := initiateAuthFlowWithError("invalid-app-id", nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate authentication flow with invalid app ID: %v", err)
	}

	// Verify the error response
	ts.Require().Equal("FES-1003", errorResp.Code, "Expected error code for invalid app ID")
	ts.Require().Equal("Invalid request", errorResp.Message, "Expected error message for invalid request")
	ts.Require().Equal("Invalid app ID provided in the request", errorResp.Description,
		"Expected error description for invalid app ID")
}

func (ts *BasicAuthFlowTestSuite) TestBasicAuthFlowInvalidFlowID() {
	// Step 1: Initialize the flow by calling the flow execution API
	flowStep, err := initiateAuthFlow(testAppID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate authentication flow: %v", err)
	}

	ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus, "Expected flow status to be INCOMPLETE")
	ts.Require().Equal("VIEW", flowStep.Type, "Expected flow type to be VIEW")
	ts.Require().NotEmpty(flowStep.FlowID, "Flow ID should not be empty")
	ts.Require().NotEmpty(flowStep.Data, "Flow data should not be empty")
	ts.Require().NotEmpty(flowStep.Data.Inputs, "Flow should require inputs")

	// Step 2: Attempt to complete a flow with an invalid flow ID
	inputs := map[string]string{
		"username": "someuser",
		"password": "somepassword",
	}

	errorResp, err := completeAuthFlowWithError("invalid-flow-id", inputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete authentication flow: %v", err)
	}

	// Verify the error response
	ts.Require().Equal("FES-1004", errorResp.Code, "Expected error code for invalid flow ID")
	ts.Require().Equal("Invalid request", errorResp.Message, "Expected error message for invalid request")
	ts.Require().Equal("Invalid flow ID provided in the request", errorResp.Description,
		"Expected error description for invalid flow ID")
}
