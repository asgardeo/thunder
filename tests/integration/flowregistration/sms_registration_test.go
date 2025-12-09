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
	"fmt"
	"testing"
	"time"

	"github.com/asgardeo/thunder/tests/integration/testutils"
	"github.com/stretchr/testify/suite"
)

const (
	mockNotificationServerPort = 8098
)

var (
	smsRegTestOU = testutils.OrganizationUnit{
		Handle:      "sms-reg-flow-test-ou",
		Name:        "SMS Registration Flow Test Organization Unit",
		Description: "Organization unit for SMS registration flow testing",
		Parent:      nil,
	}

	smsRegTestUserSchema = testutils.UserSchema{
		Name: "sms-test-user-type",
		Schema: map[string]interface{}{
			"username": map[string]interface{}{
				"type": "string",
			},
			"password": map[string]interface{}{
				"type": "string",
			},
			"email": map[string]interface{}{
				"type": "string",
			},
			"firstName": map[string]interface{}{
				"type": "string",
			},
			"lastName": map[string]interface{}{
				"type": "string",
			},
			"mobileNumber": map[string]interface{}{
				"type": "string",
			},
		},
		AllowSelfRegistration: true,
	}
)

type SMSRegistrationFlowTestSuite struct {
	suite.Suite
	config       *TestSuiteConfig
	mockServer   *testutils.MockNotificationServer
	userSchemaID string
	testAppID    string
	testOUID     string
}

func TestSMSRegistrationFlowTestSuite(t *testing.T) {
	suite.Run(t, new(SMSRegistrationFlowTestSuite))
}

func (ts *SMSRegistrationFlowTestSuite) SetupSuite() {
	// Initialize config
	ts.config = &TestSuiteConfig{}

	// Create test organization unit for SMS tests
	ouID, err := testutils.CreateOrganizationUnit(smsRegTestOU)
	if err != nil {
		ts.T().Fatalf("Failed to create test organization unit during setup: %v", err)
	}
	ts.testOUID = ouID

	// Create test user schema for SMS tests
	smsRegTestUserSchema.OrganizationUnitId = ts.testOUID
	schemaID, err := testutils.CreateUserType(smsRegTestUserSchema)
	if err != nil {
		ts.T().Fatalf("Failed to create test user schema during setup: %v", err)
	}
	ts.userSchemaID = schemaID

	// Create test application with allowed user types
	smsRegTestApp := testutils.Application{
		Name:                      "SMS Registration Flow Test Application",
		Description:               "Application for testing SMS registration flows",
		IsRegistrationFlowEnabled: true,
		AuthFlowGraphID:           "auth_flow_config_basic",
		RegistrationFlowGraphID:   "registration_flow_config_sms",
		ClientID:                  "sms_reg_flow_test_client",
		ClientSecret:              "sms_reg_flow_test_secret",
		RedirectURIs:              []string{"http://localhost:3000/callback"},
		AllowedUserTypes:          []string{smsRegTestUserSchema.Name},
	}

	appID, err := testutils.CreateApplication(smsRegTestApp)
	if err != nil {
		ts.T().Fatalf("Failed to create test application during setup: %v", err)
	}
	ts.testAppID = appID

	// Start mock notification server
	ts.mockServer = testutils.NewMockNotificationServer(mockNotificationServerPort)
	err = ts.mockServer.Start()
	if err != nil {
		ts.T().Fatalf("Failed to start mock notification server: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
	ts.T().Log("Mock notification server started successfully")

	// Store original app config (this will be the created app config)
	ts.config.OriginalAppConfig, err = getAppConfig(ts.testAppID)
	if err != nil {
		ts.T().Fatalf("Failed to get original app config during setup: %v", err)
	}
}

func (ts *SMSRegistrationFlowTestSuite) TearDownSuite() {
	// Delete test users created during SMS registration tests
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
	if ts.testAppID != "" {
		if err := testutils.DeleteApplication(ts.testAppID); err != nil {
			ts.T().Logf("Failed to delete test application during teardown: %v", err)
		}
	}

	// Delete test organization unit
	if ts.testOUID != "" {
		if err := testutils.DeleteOrganizationUnit(ts.testOUID); err != nil {
			ts.T().Logf("Failed to delete test organization unit during teardown: %v", err)
		}
	}

	if ts.userSchemaID != "" {
		if err := testutils.DeleteUserType(ts.userSchemaID); err != nil {
			ts.T().Logf("Failed to delete test user schema during teardown: %v", err)
		}
	}

}

func (ts *SMSRegistrationFlowTestSuite) TestSMSRegistrationFlowWithMobileNumber() {
	// Update app to use SMS flow
	err := updateAppConfig(ts.testAppID, "auth_flow_config_sms", "registration_flow_config_sms")
	if err != nil {
		ts.T().Fatalf("Failed to update app config for SMS flow: %v", err)
	}

	// Generate unique mobile number for registration
	mobileNumber := generateUniqueMobileNumber()

	// Step 1: Initialize the registration flow by calling the flow execution API
	flowStep, err := initiateRegistrationFlow(ts.testAppID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate registration flow: %v", err)
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
	inputs := map[string]string{
		"mobileNumber": mobileNumber,
	}

	otpFlowStep, err := completeRegistrationFlow(flowStep.FlowID, "action_001", inputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete registration flow with mobile number: %v", err)
	}

	// Verify OTP input is now required
	ts.Require().Equal("INCOMPLETE", otpFlowStep.FlowStatus, "Expected flow status to be INCOMPLETE")
	ts.Require().Equal("VIEW", otpFlowStep.Type, "Expected flow type to be VIEW")
	ts.Require().NotEmpty(otpFlowStep.Data, "Flow data should not be empty")
	ts.Require().NotEmpty(otpFlowStep.Data.Inputs, "Flow should require inputs")
	ts.Require().True(HasInput(otpFlowStep.Data.Inputs, "otp"), "OTP input should be required")

	// Wait for SMS to be sent
	time.Sleep(1000 * time.Millisecond)

	// Verify SMS was sent
	lastMessage := ts.mockServer.GetLastMessage()
	ts.Require().NotNil(lastMessage, "Last message should not be nil")
	ts.Require().NotEmpty(lastMessage.OTP, "OTP should be extracted from message")

	// Step 3: Complete registration with OTP
	otpInputs := map[string]string{
		"otp": lastMessage.OTP,
	}

	completeFlowStep, err := completeRegistrationFlow(flowStep.FlowID, "", otpInputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete registration flow with OTP: %v", err)
	}

	ts.Require().Equal("INCOMPLETE", completeFlowStep.FlowStatus,
		"Expected flow status to be INCOMPLETE after OTP input")
	ts.Require().Equal("VIEW", completeFlowStep.Type, "Expected flow type to be VIEW after OTP input")
	ts.Require().NotEmpty(completeFlowStep.Data, "Flow data should not be empty after OTP input")
	ts.Require().NotEmpty(completeFlowStep.Data.Inputs, "Flow should require inputs after OTP input")
	ts.Require().True(ValidateRequiredInputs(completeFlowStep.Data.Inputs, []string{"email"}),
		"Email should be a required inputs after first step")

	// Step 4: Provide additional attributes
	fillInputs := []Inputs{
		{
			Identifier: "firstName",
			Type:       "string",
			Required:   true,
		},
		{
			Identifier: "lastName",
			Type:       "string",
			Required:   true,
		},
	}
	fillInputs = append(fillInputs, completeFlowStep.Data.Inputs...)
	attrInputs := fillRequiredRegistrationAttributes(fillInputs, mobileNumber)
	completeFlowStep, err = completeRegistrationFlow(flowStep.FlowID, "", attrInputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete registration flow with attributes: %v", err)
	}

	// Verify successful registration
	ts.Require().Equal("COMPLETE", completeFlowStep.FlowStatus, "Expected flow status to be COMPLETE")
	ts.Require().NotEmpty(completeFlowStep.Assertion,
		"JWT assertion should be returned after successful registration")
	ts.Require().Empty(completeFlowStep.FailureReason,
		"Failure reason should be empty for successful registration")

	// Decode and validate JWT claims
	jwtClaims, err := testutils.DecodeJWT(completeFlowStep.Assertion)
	ts.Require().NoError(err, "Failed to decode JWT assertion")
	ts.Require().NotNil(jwtClaims, "JWT claims should not be nil")

	// Validate JWT contains expected user type and OU ID
	ts.Require().Equal(smsRegTestUserSchema.Name, jwtClaims.UserType, "Expected userType to match created schema")
	ts.Require().Equal(ts.testOUID, jwtClaims.OuID, "Expected ouId to match the created organization unit")
	ts.Require().Equal(ts.testAppID, jwtClaims.Aud, "Expected aud to match the application ID")
	ts.Require().NotEmpty(jwtClaims.Sub, "JWT subject should not be empty")

	// Step 5: Verify the user was created by searching via the user API
	user, err := testutils.FindUserByAttribute("mobileNumber", mobileNumber)
	if err != nil {
		ts.T().Fatalf("Failed to retrieve user by mobile number: %v", err)
	}
	ts.Require().NotNil(user, "User should be found in user list after registration")

	// Store the created user for cleanup
	if user != nil {
		ts.config.CreatedUserIDs = append(ts.config.CreatedUserIDs, user.ID)
	}
}

func (ts *SMSRegistrationFlowTestSuite) TestSMSRegistrationFlowWithUsername() {
	// Update app to use SMS flow with username
	err := updateAppConfig(ts.testAppID, "auth_flow_config_sms_with_username", "registration_flow_config_sms_with_username")
	if err != nil {
		ts.T().Fatalf("Failed to update app config for SMS flow with username: %v", err)
	}

	// Generate unique username
	username := generateUniqueUsername("smsreguser")

	// Step 1: Initialize the registration flow
	flowStep, err := initiateRegistrationFlow(ts.testAppID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate registration flow: %v", err)
	}

	// Validate that username input is required
	ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus, "Expected flow status to be INCOMPLETE")
	ts.Require().Equal("VIEW", flowStep.Type, "Expected flow type to be VIEW")
	ts.Require().NotEmpty(flowStep.FlowID, "Flow ID should not be empty")
	ts.Require().NotEmpty(flowStep.Data, "Flow data should not be empty")
	ts.Require().NotEmpty(flowStep.Data.Inputs, "Flow should require inputs")
	ts.Require().True(HasInput(flowStep.Data.Inputs, "username"), "Username input should be required")

	// Clear any previous messages
	ts.mockServer.ClearMessages()

	// Step 2: Continue the flow with username
	inputs := map[string]string{
		"username": username,
	}
	otpFlowStep, err := completeRegistrationFlow(flowStep.FlowID, "action_001", inputs)
	if err != nil {
		ts.T().Fatalf("Failed to continue registration flow with username: %v", err)
	}

	// Verify that mobile number input is now required
	ts.Require().Equal("INCOMPLETE", otpFlowStep.FlowStatus,
		"Expected flow status to be INCOMPLETE after username input")
	ts.Require().Equal("VIEW", otpFlowStep.Type, "Expected flow type to be VIEW after username input")
	ts.Require().NotEmpty(otpFlowStep.Data, "Flow data should not be empty after username input")
	ts.Require().NotEmpty(otpFlowStep.Data.Inputs, "Flow should require inputs after username input")
	ts.Require().True(HasInput(otpFlowStep.Data.Inputs, "mobileNumber"),
		"Mobile number input should be required after username input")

	// Step 3: Continue the flow with mobile number
	mobileNumber := generateUniqueMobileNumber()
	inputs = map[string]string{
		"mobileNumber": mobileNumber,
	}
	otpFlowStep, err = completeRegistrationFlow(otpFlowStep.FlowID, "", inputs)
	if err != nil {
		ts.T().Fatalf("Failed to continue registration flow with mobile number: %v", err)
	}

	// Verify OTP input is now required
	ts.Require().Equal("INCOMPLETE", otpFlowStep.FlowStatus,
		"Expected flow status to be INCOMPLETE after username input")
	ts.Require().Equal("VIEW", otpFlowStep.Type, "Expected flow type to be VIEW after username input")
	ts.Require().NotEmpty(otpFlowStep.Data, "Flow data should not be empty after username input")
	ts.Require().NotEmpty(otpFlowStep.Data.Inputs, "Flow should require inputs after username input")
	ts.Require().True(HasInput(otpFlowStep.Data.Inputs, "otp"), "OTP input should be required after username input")

	// Wait for SMS to be sent
	time.Sleep(500 * time.Millisecond)

	// Verify SMS was sent
	lastMessage := ts.mockServer.GetLastMessage()
	ts.Require().NotNil(lastMessage, "SMS should have been sent")
	ts.Require().NotEmpty(lastMessage.OTP, "OTP should be available")

	// Step 4: Complete registration with OTP
	otpInputs := map[string]string{
		"otp": lastMessage.OTP,
	}
	completeFlowStep, err := completeRegistrationFlow(flowStep.FlowID, "", otpInputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete registration flow with OTP: %v", err)
	}

	// Verify successful registration
	ts.Require().Equal("COMPLETE", completeFlowStep.FlowStatus, "Expected flow status to be COMPLETE")
	ts.Require().NotEmpty(completeFlowStep.Assertion,
		"JWT assertion should be returned after successful registration")
	ts.Require().Empty(completeFlowStep.FailureReason,
		"Failure reason should be empty for successful registration")

	// Decode and validate JWT claims
	jwtClaims, err := testutils.DecodeJWT(completeFlowStep.Assertion)
	ts.Require().NoError(err, "Failed to decode JWT assertion")
	ts.Require().NotNil(jwtClaims, "JWT claims should not be nil")

	// Validate JWT contains expected user type and OU ID
	ts.Require().Equal(smsRegTestUserSchema.Name, jwtClaims.UserType, "Expected userType to match created schema")
	ts.Require().Equal(ts.testOUID, jwtClaims.OuID, "Expected ouId to match the created organization unit")
	ts.Require().Equal(ts.testAppID, jwtClaims.Aud, "Expected aud to match the application ID")
	ts.Require().NotEmpty(jwtClaims.Sub, "JWT subject should not be empty")

	// Step 5: Verify the user was created by searching via the user API
	user, err := testutils.FindUserByAttribute("username", username)
	if err != nil {
		ts.T().Fatalf("Failed to retrieve user by username: %v", err)
	}
	ts.Require().NotNil(user, "User should be found in user list after registration")

	// Store the created user for cleanup
	if user != nil {
		ts.config.CreatedUserIDs = append(ts.config.CreatedUserIDs, user.ID)
	}
}

func (ts *SMSRegistrationFlowTestSuite) TestSMSRegistrationFlowInvalidOTP() {
	// Update app to use SMS flow
	err := updateAppConfig(ts.testAppID, "auth_flow_config_sms", "registration_flow_config_sms")
	if err != nil {
		ts.T().Fatalf("Failed to update app config for SMS flow: %v", err)
	}

	// Generate unique mobile number
	mobileNumber := generateUniqueMobileNumber()

	// Step 1: Initialize the registration flow and provide mobile number
	inputs := map[string]string{
		"mobileNumber": mobileNumber,
	}

	flowStep, err := initiateRegistrationFlow(ts.testAppID, inputs)
	if err != nil {
		ts.T().Fatalf("Failed to initiate registration flow: %v", err)
	}

	// Clear any previous messages
	ts.mockServer.ClearMessages()

	// Continue flow to trigger OTP sending
	otpFlowStep, err := completeRegistrationFlow(flowStep.FlowID, "action_001", inputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete registration flow with mobile number: %v", err)
	}

	ts.Require().Equal("INCOMPLETE", otpFlowStep.FlowStatus, "Expected flow status to be INCOMPLETE")

	// Wait for SMS to be sent
	time.Sleep(500 * time.Millisecond)

	// Step 2: Try with invalid OTP
	invalidOTPInputs := map[string]string{
		"otp": "000000",
	}

	completeFlowStep, err := completeRegistrationFlow(flowStep.FlowID, "", invalidOTPInputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete registration flow with invalid OTP: %v", err)
	}

	// Verify registration failure
	ts.Require().Equal("ERROR", completeFlowStep.FlowStatus, "Expected flow status to be ERROR")
	ts.Require().Empty(completeFlowStep.Assertion, "No JWT assertion should be returned for failed registration")
	ts.Require().NotEmpty(completeFlowStep.FailureReason, "Failure reason should be provided for invalid OTP")
	ts.Equal("invalid OTP provided", completeFlowStep.FailureReason,
		"Expected failure reason to indicate invalid OTP")
}

func (ts *SMSRegistrationFlowTestSuite) TestSMSRegistrationFlowSingleRequestWithMobileNumber() {
	// Update app to use SMS flow
	err := updateAppConfig(ts.testAppID, "auth_flow_config_sms", "registration_flow_config_sms")
	if err != nil {
		ts.T().Fatalf("Failed to update app config for SMS flow: %v", err)
	}

	// Clear any previous messages
	ts.mockServer.ClearMessages()

	// Generate unique mobile number
	mobileNumber := generateUniqueMobileNumber()

	// Step 1: Initialize the registration flow
	flowStep, err := initiateRegistrationFlow(ts.testAppID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate registration flow: %v", err)
	}

	// Step 2: Provide mobile number with action to trigger SMS
	ts.mockServer.ClearMessages()
	inputs := map[string]string{
		"mobileNumber": mobileNumber,
	}

	otpStep, err := completeRegistrationFlow(flowStep.FlowID, "action_001", inputs)
	if err != nil {
		ts.T().Fatalf("Failed to provide mobile number: %v", err)
	}

	// Should require OTP input now
	ts.Require().Equal("INCOMPLETE", otpStep.FlowStatus, "Expected flow status to be INCOMPLETE")
	ts.Require().Equal("VIEW", otpStep.Type, "Expected flow type to be VIEW")

	// Wait for SMS to be sent
	time.Sleep(500 * time.Millisecond)

	// Get the OTP from mock server
	lastMessage := ts.mockServer.GetLastMessage()
	ts.Require().NotNil(lastMessage, "SMS should have been sent")
	ts.Require().NotEmpty(lastMessage.OTP, "OTP should be available")

	// Step 3: Complete with OTP
	otpInputs := map[string]string{
		"otp": lastMessage.OTP,
	}

	provisionStep, err := completeRegistrationFlow(otpStep.FlowID, "", otpInputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete registration flow with OTP: %v", err)
	}

	// Should now require user attributes
	ts.Require().Equal("INCOMPLETE", provisionStep.FlowStatus, "Expected flow status to be INCOMPLETE for provisioning")

	// Step 4: Provide user attributes
	userInputs := map[string]string{
		"firstName":    "Test",
		"lastName":     "User",
		"email":        fmt.Sprintf("%s@example.com", mobileNumber),
		"mobileNumber": mobileNumber,
	}

	completeFlowStep, err := completeRegistrationFlow(provisionStep.FlowID, "", userInputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete registration with user attributes: %v", err)
	}

	// Verify successful registration
	ts.Require().Equal("COMPLETE", completeFlowStep.FlowStatus, "Expected flow status to be COMPLETE")
	ts.Require().NotEmpty(completeFlowStep.Assertion,
		"JWT assertion should be returned after successful registration")
	ts.Require().Empty(completeFlowStep.FailureReason,
		"Failure reason should be empty for successful registration")

	// Decode and validate JWT claims
	jwtClaims, err := testutils.DecodeJWT(completeFlowStep.Assertion)
	ts.Require().NoError(err, "Failed to decode JWT assertion")
	ts.Require().NotNil(jwtClaims, "JWT claims should not be nil")

	// Validate JWT contains expected user type and OU ID
	ts.Require().Equal(smsRegTestUserSchema.Name, jwtClaims.UserType, "Expected userType to match created schema")
	ts.Require().Equal(ts.testOUID, jwtClaims.OuID, "Expected ouId to match the created organization unit")
	ts.Require().Equal(ts.testAppID, jwtClaims.Aud, "Expected aud to match the application ID")
	ts.Require().NotEmpty(jwtClaims.Sub, "JWT subject should not be empty")

	// Step 3: Verify the user was created by searching via the user API
	user, err := testutils.FindUserByAttribute("mobileNumber", mobileNumber)
	if err != nil {
		ts.T().Fatalf("Failed to retrieve user by mobile number: %v", err)
	}
	ts.Require().NotNil(user, "User should be found in user list after registration")

	// Store the created user for cleanup
	if user != nil {
		ts.config.CreatedUserIDs = append(ts.config.CreatedUserIDs, user.ID)
	}
}

// Helper function to generate unique mobile numbers
func generateUniqueMobileNumber() string {
	return fmt.Sprintf("+1234567%d", time.Now().UnixNano()%10000)
}

// Helper to fill required attributes for registration
func fillRequiredRegistrationAttributes(inputs []Inputs, mobile string) map[string]string {
	attrInputs := map[string]string{}
	for _, input := range inputs {
		if input.Required {
			switch input.Identifier {
			case "firstName":
				attrInputs["firstName"] = "Test"
			case "lastName":
				attrInputs["lastName"] = "User"
			case "email":
				attrInputs["email"] = fmt.Sprintf("%s@example.com", mobile)
			default:
				attrInputs[input.Identifier] = "dummy"
			}
		}
	}
	return attrInputs
}
