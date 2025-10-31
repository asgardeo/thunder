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
	"time"

	"github.com/asgardeo/thunder/tests/integration/testutils"
	"github.com/stretchr/testify/suite"
)

const (
	mockDecisionNotificationServerPort = 8098
)

var (
	decisionTestApp = testutils.Application{
		Name:                      "Decision Flow Test Application",
		Description:               "Application for testing decision flows",
		IsRegistrationFlowEnabled: false,
		AuthFlowGraphID:           "auth_flow_config_decision_and_mfa_test_1",
		RegistrationFlowGraphID:   "registration_flow_config_basic",
		ClientID:                  "decision_flow_test_client",
		ClientSecret:              "decision_flow_test_secret",
		RedirectURIs:              []string{"http://localhost:3000/callback"},
	}

	decisionTestOU = testutils.OrganizationUnit{
		Handle:      "decision-flow-test-ou",
		Name:        "Decision Flow Test Organization Unit",
		Description: "Organization unit for decision flow testing",
		Parent:      nil,
	}

	testUserWithMobileDecision = testutils.User{
		Type: "person",
		Attributes: json.RawMessage(`{
			"username": "decisionuser1",
			"password": "testpassword",
			"email": "decisionuser1@example.com",
			"firstName": "Decision",
			"lastName": "User1",
			"mobileNumber": "+1234567890"
		}`),
	}

	testUserWithoutMobileDecision = testutils.User{
		Type: "person",
		Attributes: json.RawMessage(`{
			"username": "decisionuser2",
			"password": "testpassword",
			"email": "decisionuser2@example.com",
			"firstName": "Decision",
			"lastName": "User2"
		}`),
	}
)

var (
	decisionTestAppID string
	decisionTestOUID  string
)

type DecisionAndMFAFlowTestSuite struct {
	suite.Suite
	config     *TestSuiteConfig
	mockServer *testutils.MockNotificationServer
}

func TestDecisionAndMFAFlowTestSuite(t *testing.T) {
	suite.Run(t, new(DecisionAndMFAFlowTestSuite))
}

func (ts *DecisionAndMFAFlowTestSuite) SetupSuite() {
	// Initialize config
	ts.config = &TestSuiteConfig{}

	// Create test organization unit for decision tests
	ouID, err := testutils.CreateOrganizationUnit(decisionTestOU)
	if err != nil {
		ts.T().Fatalf("Failed to create test organization unit during setup: %v", err)
	}
	decisionTestOUID = ouID

	// Create test application for decision tests
	appID, err := testutils.CreateApplication(decisionTestApp)
	if err != nil {
		ts.T().Fatalf("Failed to create test application during setup: %v", err)
	}
	decisionTestAppID = appID

	// Start mock notification server
	ts.mockServer = testutils.NewMockNotificationServer(mockDecisionNotificationServerPort)
	err = ts.mockServer.Start()
	if err != nil {
		ts.T().Fatalf("Failed to start mock notification server: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
	ts.T().Log("Mock notification server started successfully")

	// Create test users with the created OU
	userWithMobile := testUserWithMobileDecision
	userWithMobile.OrganizationUnit = decisionTestOUID
	userWithoutMobile := testUserWithoutMobileDecision
	userWithoutMobile.OrganizationUnit = decisionTestOUID

	userIDs, err := testutils.CreateMultipleUsers(userWithMobile, userWithoutMobile)
	if err != nil {
		ts.T().Fatalf("Failed to create test users during setup: %v", err)
	}
	ts.config.CreatedUserIDs = userIDs
	ts.T().Logf("Test users created with IDs: %v", ts.config.CreatedUserIDs)

	// Store original app config (this will be the created app config)
	ts.config.OriginalAppConfig, err = getAppConfig(decisionTestAppID)
	if err != nil {
		ts.T().Fatalf("Failed to get original app config during setup: %v", err)
	}

	// Update app to use decision flow template
	err = updateAppConfig(decisionTestAppID, "auth_flow_config_decision_and_mfa_test_1")
	if err != nil {
		ts.T().Fatalf("Failed to update app config for decision flow: %v", err)
	}
}

func (ts *DecisionAndMFAFlowTestSuite) TearDownSuite() {
	// Delete test users
	if err := testutils.CleanupUsers(ts.config.CreatedUserIDs); err != nil {
		ts.T().Logf("Failed to cleanup users during teardown: %v", err)
	}

	// Stop mock server
	if ts.mockServer != nil {
		err := ts.mockServer.Stop()
		if err != nil {
			ts.T().Logf("Failed to stop mock notification server during teardown: %v", err)
		}
	}

	// Delete test application
	if decisionTestAppID != "" {
		if err := testutils.DeleteApplication(decisionTestAppID); err != nil {
			ts.T().Logf("Failed to delete test application during teardown: %v", err)
		}
	}

	// Delete test organization unit
	if decisionTestOUID != "" {
		if err := testutils.DeleteOrganizationUnit(decisionTestOUID); err != nil {
			ts.T().Logf("Failed to delete test organization unit during teardown: %v", err)
		}
	}

}

func (ts *DecisionAndMFAFlowTestSuite) TestBasicAuthWithMobileUserSMSOTP() {
	// Step 1: Initialize the flow - should present decision choice
	flowStep, err := initiateAuthFlow(decisionTestAppID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate authentication flow: %v", err)
	}

	ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus, "Expected flow status to be INCOMPLETE")
	ts.Require().Equal("VIEW", flowStep.Type, "Expected flow type to be VIEW")
	ts.Require().NotEmpty(flowStep.FlowID, "Flow ID should not be empty")

	// Validate that decision input is required
	ts.Require().NotEmpty(flowStep.Data, "Flow data should not be empty")
	ts.Require().NotEmpty(flowStep.Data.Actions, "Flow should require actions")

	// Check if expected actions are present
	expectedActions := []string{"basic_auth", "prompt_mobile"}
	ts.Require().True(ValidateRequiredActions(flowStep.Data.Actions, expectedActions),
		"Expected actions basic_auth and prompt_mobile should be present")

	// Step 2: Choose basic auth
	basicAuthStep, err := completeAuthFlow(flowStep.FlowID, "basic_auth", map[string]string{})
	if err != nil {
		ts.T().Fatalf("Failed to complete authentication flow with decision: %v", err)
	}

	// Should now require username and password
	ts.Require().Equal("INCOMPLETE", basicAuthStep.FlowStatus, "Expected flow status to be INCOMPLETE")
	ts.Require().Equal("VIEW", basicAuthStep.Type, "Expected flow type to be VIEW")

	// Validate required inputs using utility function
	expectedInputs := []string{"username", "password"}
	ts.Require().True(ValidateRequiredInputs(basicAuthStep.Data.Inputs, expectedInputs),
		"Username and password inputs should be required")

	// Step 3: Provide username and password
	userAttrs, err := testutils.GetUserAttributes(testUserWithMobileDecision)
	ts.Require().NoError(err, "Failed to get user attributes")

	basicInputs := map[string]string{
		"username": userAttrs["username"].(string),
		"password": userAttrs["password"].(string),
	}

	// Clear any previous messages before SMS flow
	ts.mockServer.ClearMessages()

	otpFlowStep, err := completeAuthFlow(flowStep.FlowID, "", basicInputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete authentication flow with credentials: %v", err)
	}

	// Should now require OTP since user has mobile number
	ts.Require().Equal("INCOMPLETE", otpFlowStep.FlowStatus, "Expected flow status to be INCOMPLETE")
	ts.Require().Equal("VIEW", otpFlowStep.Type, "Expected flow type to be VIEW")

	var hasOTP bool
	for _, input := range otpFlowStep.Data.Inputs {
		if input.Name == "otp" {
			hasOTP = true
			break
		}
	}
	ts.Require().True(hasOTP, "OTP input should be required")

	// Wait for SMS to be sent
	time.Sleep(500 * time.Millisecond)

	// Verify SMS was sent
	lastMessage := ts.mockServer.GetLastMessage()
	ts.Require().NotNil(lastMessage, "Last message should not be nil")
	ts.Require().NotEmpty(lastMessage.OTP, "OTP should be extracted from message")

	// Step 4: Complete authentication with OTP
	otpInputs := map[string]string{
		"otp": lastMessage.OTP,
	}

	completeFlowStep, err := completeAuthFlow(flowStep.FlowID, "", otpInputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete authentication flow with OTP: %v", err)
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
	ts.Require().Equal(decisionTestOUID, jwtClaims.OuID, "Expected ouId to match the created organization unit")
	ts.Require().Equal(decisionTestAppID, jwtClaims.Aud, "Expected aud to match the application ID")
	ts.Require().NotEmpty(jwtClaims.Sub, "JWT subject should not be empty")
}

func (ts *DecisionAndMFAFlowTestSuite) TestBasicAuthWithoutMobileUserSMSOTP() {
	// Test case 1: Authentication with basic auth with user not having mobile, provide mobile, then SMS OTP
	ts.Run("TestBasicAuthWithoutMobileUserSMSOTP_ProvideMobile", func() {
		// Step 1: Initialize the flow - should present decision choice
		flowStep, err := initiateAuthFlow(decisionTestAppID, nil)
		if err != nil {
			ts.T().Fatalf("Failed to initiate authentication flow: %v", err)
		}

		// Validate that decision input is required
		ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus, "Expected flow status to be INCOMPLETE")
		ts.Require().NotEmpty(flowStep.Data, "Flow data should not be empty")
		ts.Require().NotEmpty(flowStep.Data.Actions, "Flow should require actions")

		// Check if expected actions are present
		for _, action := range flowStep.Data.Actions {
			if action.Type != "VIEW" {
				ts.T().Fatalf("Expected action type VIEW, but got %s", action.Type)
			}
			if action.ID != "basic_auth" && action.ID != "prompt_mobile" {
				ts.T().Fatalf("Expected action ID to be 'basic_auth' or 'prompt_mobile', but got %s", action.ID)
			}
		}

		// Step 2: Choose basic auth
		_, err = completeAuthFlow(flowStep.FlowID, "basic_auth", map[string]string{})
		if err != nil {
			ts.T().Fatalf("Failed to complete authentication flow with decision: %v", err)
		}

		// Step 3: Provide username and password
		var userAttrs map[string]interface{}
		err = json.Unmarshal(testUserWithoutMobileDecision.Attributes, &userAttrs)
		ts.Require().NoError(err, "Failed to unmarshal user attributes")

		basicInputs := map[string]string{
			"username": userAttrs["username"].(string),
			"password": userAttrs["password"].(string),
		}

		mobilePromptStep, err := completeAuthFlow(flowStep.FlowID, "", basicInputs)
		if err != nil {
			ts.T().Fatalf("Failed to complete authentication flow with credentials: %v", err)
		}

		// Should now ask for mobile number since user doesn't have one
		ts.Require().Equal("INCOMPLETE", mobilePromptStep.FlowStatus, "Expected flow status to be INCOMPLETE")
		ts.Require().Equal("VIEW", mobilePromptStep.Type, "Expected flow type to be VIEW")

		var hasMobileNumber bool
		for _, input := range mobilePromptStep.Data.Inputs {
			if input.Name == "mobileNumber" {
				hasMobileNumber = true
				break
			}
		}
		ts.Require().True(hasMobileNumber, "Mobile number input should be required")

		// Clear any previous messages before SMS flow
		ts.mockServer.ClearMessages()

		// Step 4: Provide mobile number
		mobileInputs := map[string]string{
			"mobileNumber": "+1987654321",
		}

		otpFlowStep, err := completeAuthFlow(flowStep.FlowID, "", mobileInputs)
		if err != nil {
			ts.T().Fatalf("Failed to complete authentication flow with mobile number: %v", err)
		}

		// Should now require OTP
		ts.Require().Equal("INCOMPLETE", otpFlowStep.FlowStatus, "Expected flow status to be INCOMPLETE")
		ts.Require().Equal("VIEW", otpFlowStep.Type, "Expected flow type to be VIEW")

		var hasOTP bool
		for _, input := range otpFlowStep.Data.Inputs {
			if input.Name == "otp" {
				hasOTP = true
				break
			}
		}
		ts.Require().True(hasOTP, "OTP input should be required")

		// Wait for SMS to be sent
		time.Sleep(500 * time.Millisecond)

		// Verify SMS was sent
		lastMessage := ts.mockServer.GetLastMessage()
		ts.Require().NotNil(lastMessage, "Last message should not be nil")
		ts.Require().NotEmpty(lastMessage.OTP, "OTP should be extracted from message")

		// Step 5: Complete authentication with OTP
		otpInputs := map[string]string{
			"otp": lastMessage.OTP,
		}

		completeFlowStep, err := completeAuthFlow(flowStep.FlowID, "", otpInputs)
		if err != nil {
			ts.T().Fatalf("Failed to complete authentication flow with OTP: %v", err)
		}

		// Verify successful authentication
		ts.Require().Equal("COMPLETE", completeFlowStep.FlowStatus, "Expected flow status to be COMPLETE")
		ts.Require().NotEmpty(completeFlowStep.Assertion,
			"JWT assertion should be returned after successful authentication")
		ts.Require().Empty(completeFlowStep.FailureReason,
			"Failure reason should be empty for successful authentication")

		// Decode and validate JWT claims
		jwtClaims, err := testutils.DecodeJWT(completeFlowStep.Assertion)
		ts.Require().NoError(err, "Failed to decode JWT assertion")
		ts.Require().NotNil(jwtClaims, "JWT claims should not be nil")

		// Validate JWT contains expected user type and OU ID
		ts.Require().Equal("person", jwtClaims.UserType, "Expected userType to be 'person'")
		ts.Require().Equal(decisionTestOUID, jwtClaims.OuID, "Expected ouId to match the created organization unit")
		ts.Require().Equal(decisionTestAppID, jwtClaims.Aud, "Expected aud to match the application ID")
		ts.Require().NotEmpty(jwtClaims.Sub, "JWT subject should not be empty")
	})

	// Test case 2: Retry auth flow for same user - should not prompt for mobile again
	ts.Run("TestBasicAuthWithoutMobileUserSMSOTP_RetryAuth", func() {
		// Step 1: Initialize the flow - should present decision choice
		flowStep, err := initiateAuthFlow(decisionTestAppID, nil)
		if err != nil {
			ts.T().Fatalf("Failed to initiate authentication flow: %v", err)
		}

		ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus, "Expected flow status to be INCOMPLETE")
		ts.Require().NotEmpty(flowStep.Data, "Flow data should not be empty")
		ts.Require().NotEmpty(flowStep.Data.Actions, "Flow should require actions")

		// Check if expected actions are present
		for _, action := range flowStep.Data.Actions {
			if action.Type != "VIEW" {
				ts.T().Fatalf("Expected action type VIEW, but got %s", action.Type)
			}
			if action.ID != "basic_auth" && action.ID != "prompt_mobile" {
				ts.T().Fatalf("Expected action ID to be 'basic_auth' or 'prompt_mobile', but got %s", action.ID)
			}
		}

		// Step 2: Choose basic auth
		basicAuthStep, err := completeAuthFlow(flowStep.FlowID, "basic_auth", map[string]string{})
		if err != nil {
			ts.T().Fatalf("Failed to complete authentication flow with decision: %v", err)
		}

		// Should now require username and password
		ts.Require().Equal("INCOMPLETE", basicAuthStep.FlowStatus, "Expected flow status to be INCOMPLETE")
		ts.Require().Equal("VIEW", basicAuthStep.Type, "Expected flow type to be VIEW")

		var hasUsername, hasPassword bool
		for _, input := range basicAuthStep.Data.Inputs {
			if input.Name == "username" {
				hasUsername = true
			}
			if input.Name == "password" {
				hasPassword = true
			}
		}
		ts.Require().True(hasUsername, "Username input should be required")
		ts.Require().True(hasPassword, "Password input should be required")

		// Step 3: Provide username and password
		var userAttrs map[string]interface{}
		err = json.Unmarshal(testUserWithoutMobileDecision.Attributes, &userAttrs)
		ts.Require().NoError(err, "Failed to unmarshal user attributes")

		basicInputs := map[string]string{
			"username": userAttrs["username"].(string),
			"password": userAttrs["password"].(string),
		}

		otpFlowStep, err := completeAuthFlow(flowStep.FlowID, "", basicInputs)
		if err != nil {
			ts.T().Fatalf("Failed to complete authentication flow with mobile number: %v", err)
		}

		// Should now require OTP
		ts.Require().Equal("INCOMPLETE", otpFlowStep.FlowStatus, "Expected flow status to be INCOMPLETE")
		ts.Require().Equal("VIEW", otpFlowStep.Type, "Expected flow type to be VIEW")

		var hasOTP bool
		for _, input := range otpFlowStep.Data.Inputs {
			if input.Name == "otp" {
				hasOTP = true
				break
			}
		}
		ts.Require().True(hasOTP, "OTP input should be required")

		// Wait for SMS to be sent
		time.Sleep(500 * time.Millisecond)

		// Verify SMS was sent
		lastMessage := ts.mockServer.GetLastMessage()
		ts.Require().NotNil(lastMessage, "Last message should not be nil")
		ts.Require().NotEmpty(lastMessage.OTP, "OTP should be extracted from message")

		// Step 5: Complete authentication with OTP
		otpInputs := map[string]string{
			"otp": lastMessage.OTP,
		}

		completeFlowStep, err := completeAuthFlow(flowStep.FlowID, "", otpInputs)
		if err != nil {
			ts.T().Fatalf("Failed to complete authentication flow with OTP: %v", err)
		}

		// Verify successful authentication
		ts.Require().Equal("COMPLETE", completeFlowStep.FlowStatus, "Expected flow status to be COMPLETE")
		ts.Require().NotEmpty(completeFlowStep.Assertion,
			"JWT assertion should be returned after successful authentication")
		ts.Require().Empty(completeFlowStep.FailureReason,
			"Failure reason should be empty for successful authentication")

		// Decode and validate JWT claims
		jwtClaims, err := testutils.DecodeJWT(completeFlowStep.Assertion)
		ts.Require().NoError(err, "Failed to decode JWT assertion")
		ts.Require().NotNil(jwtClaims, "JWT claims should not be nil")

		// Validate JWT contains expected user type and OU ID
		ts.Require().Equal("person", jwtClaims.UserType, "Expected userType to be 'person'")
		ts.Require().Equal(decisionTestOUID, jwtClaims.OuID, "Expected ouId to match the created organization unit")
		ts.Require().Equal(decisionTestAppID, jwtClaims.Aud, "Expected aud to match the application ID")
		ts.Require().NotEmpty(jwtClaims.Sub, "JWT subject should not be empty")
	})
}

func (ts *DecisionAndMFAFlowTestSuite) TestSMSOTPAuthWithValidMobile() {
	// Step 1: Initialize the flow - should present decision choice
	flowStep, err := initiateAuthFlow(decisionTestAppID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate authentication flow: %v", err)
	}

	ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus, "Expected flow status to be INCOMPLETE")
	ts.Require().NotEmpty(flowStep.Data, "Flow data should not be empty")
	ts.Require().NotEmpty(flowStep.Data.Actions, "Flow should require actions")

	// Check if expected actions are present
	for _, action := range flowStep.Data.Actions {
		if action.Type != "VIEW" {
			ts.T().Fatalf("Expected action type VIEW, but got %s", action.Type)
		}
		if action.ID != "basic_auth" && action.ID != "prompt_mobile" {
			ts.T().Fatalf("Expected action ID to be 'basic_auth' or 'prompt_mobile', but got %s", action.ID)
		}
	}

	// Step 2: Choose sms OTP auth
	smsAuthStep, err := completeAuthFlow(flowStep.FlowID, "prompt_mobile", map[string]string{})
	if err != nil {
		ts.T().Fatalf("Failed to complete authentication flow with decision: %v", err)
	}

	// Should ask for mobile number
	ts.Require().Equal("INCOMPLETE", smsAuthStep.FlowStatus, "Expected flow status to be INCOMPLETE")
	ts.Require().Equal("VIEW", smsAuthStep.Type, "Expected flow type to be VIEW")

	var hasMobileNumber bool
	for _, input := range smsAuthStep.Data.Inputs {
		if input.Name == "mobileNumber" {
			hasMobileNumber = true
			break
		}
	}
	ts.Require().True(hasMobileNumber, "Mobile number input should be required")

	// Clear any previous messages before SMS flow
	ts.mockServer.ClearMessages()

	// Step 3: Provide valid mobile number from user profile
	var userAttrs map[string]interface{}
	err = json.Unmarshal(testUserWithMobileDecision.Attributes, &userAttrs)
	ts.Require().NoError(err, "Failed to unmarshal user attributes")

	mobileInputs := map[string]string{
		"mobileNumber": userAttrs["mobileNumber"].(string),
	}

	otpFlowStep, err := completeAuthFlow(flowStep.FlowID, "", mobileInputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete authentication flow with mobile number: %v", err)
	}

	// Should now require OTP
	ts.Require().Equal("INCOMPLETE", otpFlowStep.FlowStatus, "Expected flow status to be INCOMPLETE")
	ts.Require().Equal("VIEW", otpFlowStep.Type, "Expected flow type to be VIEW")

	var hasOTP bool
	for _, input := range otpFlowStep.Data.Inputs {
		if input.Name == "otp" {
			hasOTP = true
			break
		}
	}
	ts.Require().True(hasOTP, "OTP input should be required")

	// Wait for SMS to be sent
	time.Sleep(500 * time.Millisecond)

	// Verify SMS was sent
	lastMessage := ts.mockServer.GetLastMessage()
	ts.Require().NotNil(lastMessage, "Last message should not be nil")
	ts.Require().NotEmpty(lastMessage.OTP, "OTP should be extracted from message")

	// Step 4: Complete authentication with OTP
	otpInputs := map[string]string{
		"otp": lastMessage.OTP,
	}

	completeFlowStep, err := completeAuthFlow(flowStep.FlowID, "", otpInputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete authentication flow with OTP: %v", err)
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
	ts.Require().Equal(decisionTestOUID, jwtClaims.OuID, "Expected ouId to match the created organization unit")
	ts.Require().Equal(decisionTestAppID, jwtClaims.Aud, "Expected aud to match the application ID")
	ts.Require().NotEmpty(jwtClaims.Sub, "JWT subject should not be empty")
}

func (ts *DecisionAndMFAFlowTestSuite) TestSMSOTPAuthWithInvalidMobile() {
	ts.T().Log("Test Case 5: Authentication with SMS OTP decision - invalid mobile should fail")

	// Step 1: Initialize the flow - should present decision choice
	flowStep, err := initiateAuthFlow(decisionTestAppID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate authentication flow: %v", err)
	}

	ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus, "Expected flow status to be INCOMPLETE")
	ts.Require().NotEmpty(flowStep.Data, "Flow data should not be empty")
	ts.Require().NotEmpty(flowStep.Data.Actions, "Flow should require actions")

	// Check if expected actions are present
	for _, action := range flowStep.Data.Actions {
		if action.Type != "VIEW" {
			ts.T().Fatalf("Expected action type VIEW, but got %s", action.Type)
		}
		if action.ID != "basic_auth" && action.ID != "prompt_mobile" {
			ts.T().Fatalf("Expected action ID to be 'basic_auth' or 'prompt_mobile', but got %s", action.ID)
		}
	}

	// Step 2: Choose sms OTP auth
	smsAuthStep, err := completeAuthFlow(flowStep.FlowID, "prompt_mobile", map[string]string{})
	if err != nil {
		ts.T().Fatalf("Failed to complete authentication flow with decision: %v", err)
	}

	// Should ask for mobile number
	ts.Require().Equal("INCOMPLETE", smsAuthStep.FlowStatus, "Expected flow status to be INCOMPLETE")

	// Step 3: Provide invalid mobile number (not in any user profile)
	mobileInputs := map[string]string{
		"mobileNumber": "+9999999999", // Invalid mobile not associated with any user
	}

	// This should result in failure or error
	errorResp, err := completeAuthFlowWithError(flowStep.FlowID, mobileInputs)
	if err != nil {
		// If the API returned an error response, that's expected
		ts.T().Logf("Expected error occurred: %v", err)
		return
	}

	if errorResp != nil {
		// If we get an error response back, that's expected
		ts.Require().NotEmpty(errorResp.Message, "Error message should be provided")
		ts.T().Logf("Authentication failed as expected: %s", errorResp.Message)
	} else {
		ts.T().Fatalf("Expected authentication to fail with invalid mobile number")
	}
}
