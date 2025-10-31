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
	mockNotificationServerPort = 8098
)

var (
	smsAuthTestApp = testutils.Application{
		Name:                      "SMS Auth Flow Test Application",
		Description:               "Application for testing SMS authentication flows",
		IsRegistrationFlowEnabled: false,
		AuthFlowGraphID:           "auth_flow_config_basic", // Will be updated dynamically
		RegistrationFlowGraphID:   "registration_flow_config_basic",
		ClientID:                  "sms_auth_flow_test_client",
		ClientSecret:              "sms_auth_flow_test_secret",
		RedirectURIs:              []string{"http://localhost:3000/callback"},
	}

	smsAuthTestOU = testutils.OrganizationUnit{
		Handle:      "sms-auth-flow-test-ou",
		Name:        "SMS Auth Flow Test Organization Unit",
		Description: "Organization unit for SMS authentication flow testing",
		Parent:      nil,
	}

	testUserWithMobile = testutils.User{
		Type: "person",
		Attributes: json.RawMessage(`{
			"username": "smsuser",
			"password": "testpassword",
			"email": "smsuser@example.com",
			"firstName": "SMS",
			"lastName": "User",
			"mobileNumber": "+1234567890"
		}`),
	}
)

var (
	smsAuthTestAppID string
	smsAuthTestOUID  string
)

// NotificationSenderRequest represents the request to create a message notification sender.
type NotificationSenderRequest struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Provider    string           `json:"provider"`
	Properties  []SenderProperty `json:"properties"`
}

// NotificationSender represents a message notification sender.
type NotificationSender struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Provider    string           `json:"provider"`
	Properties  []SenderProperty `json:"properties"`
}

// SenderProperty represents a key-value property for a message notification sender.
type SenderProperty struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	IsSecret bool   `json:"is_secret"`
}

type SMSAuthFlowTestSuite struct {
	suite.Suite
	config     *TestSuiteConfig
	mockServer *testutils.MockNotificationServer
}

func TestSMSAuthFlowTestSuite(t *testing.T) {
	suite.Run(t, new(SMSAuthFlowTestSuite))
}

func (ts *SMSAuthFlowTestSuite) SetupSuite() {
	// Initialize config
	ts.config = &TestSuiteConfig{}

	// Create test organization unit for SMS auth tests
	ouID, err := testutils.CreateOrganizationUnit(smsAuthTestOU)
	if err != nil {
		ts.T().Fatalf("Failed to create test organization unit during setup: %v", err)
	}
	smsAuthTestOUID = ouID

	// Create test application for SMS auth tests
	appID, err := testutils.CreateApplication(smsAuthTestApp)
	if err != nil {
		ts.T().Fatalf("Failed to create test application during setup: %v", err)
	}
	smsAuthTestAppID = appID

	// Start mock notification server
	ts.mockServer = testutils.NewMockNotificationServer(mockNotificationServerPort)
	err = ts.mockServer.Start()
	if err != nil {
		ts.T().Fatalf("Failed to start mock notification server: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
	ts.T().Log("Mock notification server started successfully")

	// Create test user with mobile number using the created OU
	testUserWithMobile := testUserWithMobile
	testUserWithMobile.OrganizationUnit = smsAuthTestOUID
	userIDs, err := testutils.CreateMultipleUsers(testUserWithMobile)
	if err != nil {
		ts.T().Fatalf("Failed to create test user during setup: %v", err)
	}
	ts.config.CreatedUserIDs = userIDs
	ts.T().Logf("Test user created with ID: %s", ts.config.CreatedUserIDs[0])

	// Store original app config (this will be the created app config)
	ts.config.OriginalAppConfig, err = getAppConfig(smsAuthTestAppID)
	if err != nil {
		ts.T().Fatalf("Failed to get original app config during setup: %v", err)
	}
}

func (ts *SMSAuthFlowTestSuite) TearDownSuite() {
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
	if smsAuthTestAppID != "" {
		if err := testutils.DeleteApplication(smsAuthTestAppID); err != nil {
			ts.T().Logf("Failed to delete test application during teardown: %v", err)
		}
	}

	// Delete test organization unit
	if smsAuthTestOUID != "" {
		if err := testutils.DeleteOrganizationUnit(smsAuthTestOUID); err != nil {
			ts.T().Logf("Failed to delete test organization unit during teardown: %v", err)
		}
	}

}

func (ts *SMSAuthFlowTestSuite) TestSMSAuthFlowWithMobileNumber() {
	// Update app to use SMS flow
	err := updateAppConfig(smsAuthTestAppID, "auth_flow_config_sms")
	if err != nil {
		ts.T().Fatalf("Failed to update app config for SMS flow: %v", err)
	}

	// Step 1: Initialize the flow by calling the flow execution API
	flowStep, err := initiateAuthFlow(smsAuthTestAppID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate authentication flow: %v", err)
	}

	ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus, "Expected flow status to be INCOMPLETE")
	ts.Require().Equal("VIEW", flowStep.Type, "Expected flow type to be VIEW")
	ts.Require().NotEmpty(flowStep.FlowID, "Flow ID should not be empty")

	// Validate that mobile number input is required
	ts.Require().NotEmpty(flowStep.Data, "Flow data should not be empty")
	ts.Require().NotEmpty(flowStep.Data.Inputs, "Flow should require inputs")

	ts.Require().True(HasInput(flowStep.Data.Inputs, "mobileNumber"),
		"Mobile number input should be required")

	// Clear any previous messages
	ts.mockServer.ClearMessages()

	// Step 2: Continue the flow with mobile number
	userAttrs, err := testutils.GetUserAttributes(testUserWithMobile)
	ts.Require().NoError(err, "Failed to get user attributes")

	inputs := map[string]string{
		"mobileNumber": userAttrs["mobileNumber"].(string),
	}

	otpFlowStep, err := completeAuthFlow(flowStep.FlowID, "", inputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete authentication flow with mobile number: %v", err)
	}

	// Verify OTP input is now required
	ts.Require().Equal("INCOMPLETE", otpFlowStep.FlowStatus, "Expected flow status to be INCOMPLETE")
	ts.Require().Equal("VIEW", otpFlowStep.Type, "Expected flow type to be VIEW")
	ts.Require().NotEmpty(otpFlowStep.Data, "Flow data should not be empty")
	ts.Require().NotEmpty(otpFlowStep.Data.Inputs, "Flow should require inputs")

	ts.Require().True(HasInput(otpFlowStep.Data.Inputs, "otp"),
		"OTP input should be required")

	// Wait for SMS to be sent
	time.Sleep(500 * time.Millisecond)

	// Verify SMS was sent
	lastMessage := ts.mockServer.GetLastMessage()
	ts.Require().NotNil(lastMessage, "Last message should not be nil")
	ts.Require().NotEmpty(lastMessage.OTP, "OTP should be extracted from message")

	// Step 3: Complete authentication with OTP
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
	ts.Require().Equal(smsAuthTestOUID, jwtClaims.OuID, "Expected ouId to match the created organization unit")
	ts.Require().Equal(smsAuthTestAppID, jwtClaims.Aud, "Expected aud to match the application ID")
	ts.Require().NotEmpty(jwtClaims.Sub, "JWT subject should not be empty")
}

func (ts *SMSAuthFlowTestSuite) TestSMSAuthFlowWithUsername() {
	// Update app to use SMS flow with username
	err := updateAppConfig(smsAuthTestAppID, "auth_flow_config_sms_with_username")
	if err != nil {
		ts.T().Fatalf("Failed to update app config for SMS flow with username: %v", err)
	}

	// Step 1: Initialize the flow
	flowStep, err := initiateAuthFlow(smsAuthTestAppID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate authentication flow: %v", err)
	}

	ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus, "Expected flow status to be INCOMPLETE")
	ts.Require().Equal("VIEW", flowStep.Type, "Expected flow type to be VIEW")
	ts.Require().NotEmpty(flowStep.FlowID, "Flow ID should not be empty")

	// Validate that username input is required
	ts.Require().NotEmpty(flowStep.Data, "Flow data should not be empty")
	ts.Require().NotEmpty(flowStep.Data.Inputs, "Flow should require inputs")

	var hasUsername bool
	for _, input := range flowStep.Data.Inputs {
		if input.Name == "username" {
			hasUsername = true
		}
	}
	ts.Require().True(hasUsername, "Username input should be required")

	// Clear any previous messages
	ts.mockServer.ClearMessages()

	// Step 2: Continue the flow with username
	var userAttrs map[string]interface{}
	err = json.Unmarshal(testUserWithMobile.Attributes, &userAttrs)
	ts.Require().NoError(err, "Failed to unmarshal user attributes")

	inputs := map[string]string{
		"username": userAttrs["username"].(string),
	}

	otpFlowStep, err := completeAuthFlow(flowStep.FlowID, "", inputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete authentication flow with username: %v", err)
	}

	// Verify OTP input is now required
	ts.Require().Equal("INCOMPLETE", otpFlowStep.FlowStatus, "Expected flow status to be INCOMPLETE")
	ts.Require().Equal("VIEW", otpFlowStep.Type, "Expected flow type to be VIEW")
	ts.Require().NotEmpty(otpFlowStep.Data, "Flow data should not be empty")
	ts.Require().NotEmpty(otpFlowStep.Data.Inputs, "Flow should require inputs")

	var hasOTP bool
	for _, input := range otpFlowStep.Data.Inputs {
		if input.Name == "otp" {
			hasOTP = true
		}
	}
	ts.Require().True(hasOTP, "OTP input should be required")

	// Wait for SMS to be sent
	time.Sleep(500 * time.Millisecond)

	// Verify SMS was sent
	lastMessage := ts.mockServer.GetLastMessage()
	ts.Require().NotNil(lastMessage, "Last message should not be nil")
	ts.Require().NotEmpty(lastMessage.OTP, "OTP should be extracted from message")

	// Step 3: Complete authentication with OTP
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
	ts.Require().Equal(smsAuthTestOUID, jwtClaims.OuID, "Expected ouId to match the created organization unit")
	ts.Require().Equal(smsAuthTestAppID, jwtClaims.Aud, "Expected aud to match the application ID")
	ts.Require().NotEmpty(jwtClaims.Sub, "JWT subject should not be empty")
}

func (ts *SMSAuthFlowTestSuite) TestSMSAuthFlowInvalidOTP() {
	// Update app to use SMS flow
	err := updateAppConfig(smsAuthTestAppID, "auth_flow_config_sms")
	if err != nil {
		ts.T().Fatalf("Failed to update app config for SMS flow: %v", err)
	}

	// Step 1: Initialize the flow and provide mobile number
	var userAttrs map[string]interface{}
	err = json.Unmarshal(testUserWithMobile.Attributes, &userAttrs)
	ts.Require().NoError(err, "Failed to unmarshal user attributes")

	inputs := map[string]string{
		"mobileNumber": userAttrs["mobileNumber"].(string),
	}

	flowStep, err := initiateAuthFlow(smsAuthTestAppID, inputs)
	if err != nil {
		ts.T().Fatalf("Failed to initiate authentication flow: %v", err)
	}

	// Clear any previous messages
	ts.mockServer.ClearMessages()

	// Continue flow to trigger OTP sending
	otpFlowStep, err := completeAuthFlow(flowStep.FlowID, "", inputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete authentication flow with mobile number: %v", err)
	}

	ts.Require().Equal("INCOMPLETE", otpFlowStep.FlowStatus, "Expected flow status to be INCOMPLETE")

	// Wait for SMS to be sent
	time.Sleep(500 * time.Millisecond)

	// Step 2: Try with invalid OTP
	invalidOTPInputs := map[string]string{
		"otp": "000000", // Invalid OTP
	}

	completeFlowStep, err := completeAuthFlow(flowStep.FlowID, "", invalidOTPInputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete authentication flow with invalid OTP: %v", err)
	}

	// Verify authentication failure
	ts.Require().Equal("ERROR", completeFlowStep.FlowStatus, "Expected flow status to be ERROR")
	ts.Require().Empty(completeFlowStep.Assertion, "No JWT assertion should be returned for failed authentication")
	ts.Require().NotEmpty(completeFlowStep.FailureReason, "Failure reason should be provided for invalid OTP")
}

func (ts *SMSAuthFlowTestSuite) TestSMSAuthFlowSingleRequestWithMobileNumber() {
	// Update app to use SMS flow
	err := updateAppConfig(smsAuthTestAppID, "auth_flow_config_sms")
	if err != nil {
		ts.T().Fatalf("Failed to update app config for SMS flow: %v", err)
	}

	// Clear any previous messages
	ts.mockServer.ClearMessages()

	// Get user attributes
	var userAttrs map[string]interface{}
	err = json.Unmarshal(testUserWithMobile.Attributes, &userAttrs)
	ts.Require().NoError(err, "Failed to unmarshal user attributes")

	// Step 1: Initialize the flow with mobile number
	inputs := map[string]string{
		"mobileNumber": userAttrs["mobileNumber"].(string),
	}

	flowStep, err := initiateAuthFlow(smsAuthTestAppID, inputs)
	if err != nil {
		ts.T().Fatalf("Failed to initiate authentication flow: %v", err)
	}

	// Should require OTP input now
	ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus, "Expected flow status to be INCOMPLETE")
	ts.Require().Equal("VIEW", flowStep.Type, "Expected flow type to be VIEW")

	// Wait for SMS to be sent
	time.Sleep(500 * time.Millisecond)

	// Get the OTP from mock server
	lastMessage := ts.mockServer.GetLastMessage()
	ts.Require().NotNil(lastMessage, "SMS should have been sent")
	ts.Require().NotEmpty(lastMessage.OTP, "OTP should be available")

	// Step 2: Complete with OTP
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
	ts.Require().Equal(smsAuthTestOUID, jwtClaims.OuID, "Expected ouId to match the created organization unit")
	ts.Require().Equal(smsAuthTestAppID, jwtClaims.Aud, "Expected aud to match the application ID")
	ts.Require().NotEmpty(jwtClaims.Sub, "JWT subject should not be empty")
}
