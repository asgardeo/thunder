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

package registration

import (
	"encoding/json"
	"testing"

	"github.com/asgardeo/thunder/tests/integration/flow/common"
	"github.com/asgardeo/thunder/tests/integration/testutils"
	"github.com/stretchr/testify/suite"
)

var (
	testOU = testutils.OrganizationUnit{
		Handle:      "reg-flow-test-ou",
		Name:        "Registration Flow Test Organization Unit",
		Description: "Organization unit for registration flow testing",
		Parent:      nil,
	}

	testUserSchema = testutils.UserSchema{
		Name: "test-user-type",
		Schema: map[string]interface{}{
			"username": map[string]interface{}{
				"type": "string",
			},
			"password": map[string]interface{}{
				"type": "string",
				"credential": true,
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

type BasicRegistrationFlowTestSuite struct {
	suite.Suite
	config           *common.TestSuiteConfig
	userSchemaID     string
	testAppID        string
	testOUID         string
	testUserTypeName string
}

func TestBasicRegistrationFlowTestSuite(t *testing.T) {
	suite.Run(t, new(BasicRegistrationFlowTestSuite))
}

func (ts *BasicRegistrationFlowTestSuite) SetupSuite() {
	// Initialize config
	ts.config = &common.TestSuiteConfig{}

	// Create test organization unit
	ouID, err := testutils.CreateOrganizationUnit(testOU)
	if err != nil {
		ts.T().Fatalf("Failed to create test organization unit during setup: %v", err)
	}
	ts.testOUID = ouID

	// Create test user schema
	testUserSchema.OrganizationUnitId = ts.testOUID
	schemaID, err := testutils.CreateUserType(testUserSchema)
	if err != nil {
		ts.T().Fatalf("Failed to create test user schema during setup: %v", err)
	}
	ts.userSchemaID = schemaID
	ts.testUserTypeName = testUserSchema.Name

	// Create test application with allowed user types
	// Application relies on the default flow set by the server. Hence no need
	// to set the flow IDs here.
	testApp := testutils.Application{
		Name:                      "Registration Flow Test Application",
		Description:               "Application for testing registration flows",
		IsRegistrationFlowEnabled: true,
		ClientID:                  "reg_flow_test_client",
		ClientSecret:              "reg_flow_test_secret",
		RedirectURIs:              []string{"http://localhost:3000/callback"},
		AllowedUserTypes:          []string{testUserSchema.Name},
		AssertionConfig: map[string]interface{}{
			"user_attributes": []string{"userType", "ouId", "ouName", "ouHandle"},
		},
	}

	appID, err := testutils.CreateApplication(testApp)
	if err != nil {
		ts.T().Fatalf("Failed to create test application during setup: %v", err)
	}
	ts.testAppID = appID
}

func (ts *BasicRegistrationFlowTestSuite) TearDownSuite() {
	// Clean up users created during registration tests
	if err := testutils.CleanupUsers(ts.config.CreatedUserIDs); err != nil {
		ts.T().Logf("Failed to cleanup users during teardown: %v", err)
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

func (ts *BasicRegistrationFlowTestSuite) TestBasicRegistrationFlowSuccess() {
	// Generate unique username for this test
	username := common.GenerateUniqueUsername("reguser")

	// Step 1: Initialize the registration flow
	flowStep, err := common.InitiateRegistrationFlow(ts.testAppID, false, nil, "")
	if err != nil {
		ts.T().Fatalf("Failed to initiate registration flow: %v", err)
	}

	ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus, "Expected flow status to be INCOMPLETE")
	ts.Require().Equal("VIEW", flowStep.Type, "Expected flow type to be VIEW")
	ts.Require().NotEmpty(flowStep.FlowID, "Flow ID should not be empty")
	ts.Require().NotEmpty(flowStep.Data, "Flow data should not be empty")
	ts.Require().NotEmpty(flowStep.Data.Inputs, "Flow should require inputs")

	// Verify username and password are required inputs
	ts.Require().True(common.ValidateRequiredInputs(flowStep.Data.Inputs, []string{"username", "password"}),
		"Username and password inputs should be required")
	ts.Require().True(common.HasInput(flowStep.Data.Inputs, "username"), "Username input should be present")
	ts.Require().True(common.HasInput(flowStep.Data.Inputs, "password"), "Password input should be present")

	// Step 2: Continue the flow with registration credentials
	inputs := map[string]string{
		"username": username,
		"password": "testpassword123",
	}

	completeFlowStep, err := common.CompleteFlow(flowStep.FlowID, inputs, "action_credentials")
	if err != nil {
		ts.T().Fatalf("Failed to complete registration flow: %v", err)
	}

	// Step 3: Continue the flow with additional attributes
	ts.Require().Equal("INCOMPLETE", completeFlowStep.FlowStatus,
		"Expected flow status to be INCOMPLETE after first step")
	ts.Require().NotEmpty(completeFlowStep.Data, "Flow data should not be empty after first step")
	ts.Require().NotEmpty(completeFlowStep.Data.Inputs, "Flow should require additional inputs after first step")
	ts.Require().True(common.ValidateRequiredInputs(completeFlowStep.Data.Inputs,
		[]string{"email", "firstName", "lastName"}),
		"Email, first name, and last name should be required inputs after first step")

	inputs = map[string]string{
		"email":     username + "@example.com",
		"firstName": "Test",
		"lastName":  "User",
	}
	completeFlowStep, err = common.CompleteFlow(completeFlowStep.FlowID, inputs, "action_user_info")
	if err != nil {
		ts.T().Fatalf("Failed to complete registration flow with additional attributes: %v", err)
	}

	// Step 4: Verify successful registration
	ts.Require().Equal("COMPLETE", completeFlowStep.FlowStatus, "Expected flow status to be COMPLETE")
	ts.Require().NotEmpty(completeFlowStep.Assertion,
		"JWT assertion should be returned after successful registration")
	ts.Require().Empty(completeFlowStep.FailureReason, "Failure reason should be empty for successful registration")

	// Decode and validate JWT claims
	jwtClaims, err := testutils.DecodeJWT(completeFlowStep.Assertion)
	ts.Require().NoError(err, "Failed to decode JWT assertion")
	ts.Require().NotNil(jwtClaims, "JWT claims should not be nil")

	// Validate JWT contains expected user type and OU ID
	ts.Require().Equal(testUserSchema.Name, jwtClaims.UserType, "Expected userType to match created schema")
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

func (ts *BasicRegistrationFlowTestSuite) TestBasicRegistrationFlowDuplicateUser() {
	// Create a test user first
	testUser := testutils.User{
		OrganizationUnit: ts.testOUID,
		Type:             testUserSchema.Name,
		Attributes: json.RawMessage(`{
			"username": "duplicateuser",
			"password": "testpassword",
			"email": "duplicate@example.com",
			"firstName": "Duplicate",
			"lastName": "User"
		}`),
	}

	userIDs, err := testutils.CreateMultipleUsers(testUser)
	if err != nil {
		ts.T().Fatalf("Failed to create test user for duplicate test: %v", err)
	}
	ts.config.CreatedUserIDs = append(ts.config.CreatedUserIDs, userIDs...)

	// Step 1: Initialize the registration flow
	flowStep, err := common.InitiateRegistrationFlow(ts.testAppID, false, nil, "")
	if err != nil {
		ts.T().Fatalf("Failed to initiate registration flow: %v", err)
	}

	// Step 2: Try to register with existing username
	inputs := map[string]string{
		"username": "duplicateuser",
		"password": "newpassword123",
	}

	completeFlowStep, err := common.CompleteFlow(flowStep.FlowID, inputs, "action_credentials")
	if err != nil {
		ts.T().Fatalf("Failed to complete registration flow: %v", err)
	}

	// Step 3: Verify registration failure due to duplicate username
	ts.Require().Equal("ERROR", completeFlowStep.FlowStatus, "Expected flow status to be ERROR")
	ts.Require().Empty(completeFlowStep.Assertion, "No JWT assertion should be returned for failed registration")
	ts.Require().NotEmpty(completeFlowStep.FailureReason, "Failure reason should be provided for duplicate user")
	ts.Equal("User already exists with the provided attributes.", completeFlowStep.FailureReason,
		"Failure reason should indicate duplicate username")
}

func (ts *BasicRegistrationFlowTestSuite) TestBasicRegistrationFlowInitialInvalidInput() {
	// Step 1: Initialize the registration flow
	flowStep, err := common.InitiateRegistrationFlow(ts.testAppID, false, nil, "")
	if err != nil {
		ts.T().Fatalf("Failed to initiate registration flow: %v", err)
	}

	// Step 2: Try to register with only the username
	username := common.GenerateUniqueUsername("newuser")
	inputs := map[string]string{
		"username": username,
	}
	completeFlowStep, err := common.CompleteFlow(flowStep.FlowID, inputs, "")
	if err != nil {
		ts.T().Fatalf("Failed to complete registration flow: %v", err)
	}

	// Step 3: Verify flow prompt for username again
	ts.Require().Equal("INCOMPLETE", completeFlowStep.FlowStatus, "Expected flow status to be INCOMPLETE")
	ts.Require().Empty(completeFlowStep.Assertion, "No JWT assertion should be returned for incomplete registration")
	ts.Require().Empty(completeFlowStep.FailureReason, "Failure reason should be empty for incomplete registration")
	ts.Require().True(common.HasInput(completeFlowStep.Data.Inputs, "password"),
		"Flow should prompt for password after invalid input")

	// Step 4: Continue with the password input
	inputs = map[string]string{
		"password": "testpassword123",
	}
	completeFlowStep, err = common.CompleteFlow(completeFlowStep.FlowID, inputs, "action_credentials")
	if err != nil {
		ts.T().Fatalf("Failed to complete registration flow with username input: %v", err)
	}

	// Step 5: Continue the flow with additional attributes
	ts.Require().Equal("INCOMPLETE", completeFlowStep.FlowStatus,
		"Expected flow status to be INCOMPLETE after first step")
	ts.Require().NotEmpty(completeFlowStep.Data, "Flow data should not be empty after first step")
	ts.Require().NotEmpty(completeFlowStep.Data.Inputs, "Flow should require additional inputs after first step")
	ts.Require().True(common.ValidateRequiredInputs(completeFlowStep.Data.Inputs,
		[]string{"email", "firstName", "lastName"}),
		"Email, first name, and last name should be required inputs after first step")

	inputs = map[string]string{
		"email":     username + "@example.com",
		"firstName": "Test",
		"lastName":  "User",
	}
	completeFlowStep, err = common.CompleteFlow(completeFlowStep.FlowID, inputs, "action_user_info")
	if err != nil {
		ts.T().Fatalf("Failed to complete registration flow with additional attributes: %v", err)
	}

	// Step 6: Verify successful registration
	ts.Require().Equal("COMPLETE", completeFlowStep.FlowStatus, "Expected flow status to be COMPLETE")
	ts.Require().NotEmpty(completeFlowStep.Assertion,
		"JWT assertion should be returned after successful registration")
	ts.Require().Empty(completeFlowStep.FailureReason, "Failure reason should be empty for successful registration")

	// Decode and validate JWT claims
	jwtClaims, err := testutils.DecodeJWT(completeFlowStep.Assertion)
	ts.Require().NoError(err, "Failed to decode JWT assertion")
	ts.Require().NotNil(jwtClaims, "JWT claims should not be nil")

	// Validate JWT contains expected user type and OU ID
	ts.Require().Equal(testUserSchema.Name, jwtClaims.UserType, "Expected userType to match created schema")
	ts.Require().Equal(ts.testOUID, jwtClaims.OuID, "Expected ouId to match the created organization unit")
	ts.Require().Equal(ts.testAppID, jwtClaims.Aud, "Expected aud to match the application ID")
	ts.Require().NotEmpty(jwtClaims.Sub, "JWT subject should not be empty")

	// Step 7: Verify the user was created by searching via the user API
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

func (ts *BasicRegistrationFlowTestSuite) TestBasicRegistrationFlowSingleRequest() {
	// Generate unique username for this test
	username := common.GenerateUniqueUsername("singlereguser")

	// Step 1: Initialize the registration flow with credentials in one request
	inputs := map[string]string{
		"username":  username,
		"password":  "testpassword123",
		"email":     username + "@example.com",
		"firstName": "Single",
		"lastName":  "Request",
	}

	flowStep, err := common.InitiateRegistrationFlow(ts.testAppID, false, inputs, "")
	if err != nil {
		ts.T().Fatalf("Failed to initiate registration flow with inputs: %v", err)
	}

	// Step 2: Verify successful registration in a single request
	ts.Require().Equal("COMPLETE", flowStep.FlowStatus, "Expected flow status to be COMPLETE")
	ts.Require().NotEmpty(flowStep.Assertion,
		"JWT assertion should be returned after successful registration")
	ts.Require().Empty(flowStep.FailureReason, "Failure reason should be empty for successful registration")

	// Decode and validate JWT claims
	jwtClaims, err := testutils.DecodeJWT(flowStep.Assertion)
	ts.Require().NoError(err, "Failed to decode JWT assertion")
	ts.Require().NotNil(jwtClaims, "JWT claims should not be nil")

	// Validate JWT contains expected user type and OU ID
	ts.Require().Equal(testUserSchema.Name, jwtClaims.UserType, "Expected userType to match created schema")
	ts.Require().Equal(ts.testOUID, jwtClaims.OuID, "Expected ouId to match the created organization unit")
	ts.Require().Equal(ts.testAppID, jwtClaims.Aud, "Expected aud to match the application ID")
	ts.Require().NotEmpty(jwtClaims.Sub, "JWT subject should not be empty")

	// Step 3: Verify the user was created by searching via the user API
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

// TestBasicRegistrationFlow_WithoutTokenConfig tests that userType and OU attributes are NOT included
// in JWT assertion when TokenConfig is not specified.
func (ts *BasicRegistrationFlowTestSuite) TestBasicRegistrationFlow_WithoutTokenConfig() {
	// Create a new application without TokenConfig
	appWithoutTokenConfig := testutils.Application{
		Name:                      "Registration Flow Test Application Without Token Config",
		Description:               "Application for testing default behavior without token config",
		IsRegistrationFlowEnabled: true,
		ClientID:                  "reg_flow_test_client_no_token_config",
		ClientSecret:              "reg_flow_test_secret_no_token_config",
		RedirectURIs:              []string{"http://localhost:3000/callback"},
		AllowedUserTypes:          []string{testUserSchema.Name},
		// TokenConfig is nil - not specified
	}

	appID, err := testutils.CreateApplication(appWithoutTokenConfig)
	ts.Require().NoError(err, "Failed to create application without token config")
	defer func() {
		if err := testutils.DeleteApplication(appID); err != nil {
			ts.T().Logf("Failed to delete test application: %v", err)
		}
	}()

	// Generate unique username for this test
	username := common.GenerateUniqueUsername("reguser")

	// Execute registration flow
	flowStep, err := common.InitiateRegistrationFlow(appID, false, nil, "")
	ts.Require().NoError(err, "Failed to initiate registration flow")

	inputs := map[string]string{
		"username": username,
		"password": "testpassword123",
	}
	completeFlowStep, err := common.CompleteFlow(flowStep.FlowID, inputs, "action_credentials")
	ts.Require().NoError(err, "Failed to complete registration flow")

	inputs = map[string]string{
		"email":     username + "@example.com",
		"firstName": "Test",
		"lastName":  "User",
	}
	completeFlowStep, err = common.CompleteFlow(completeFlowStep.FlowID, inputs, "action_user_info")
	ts.Require().NoError(err, "Failed to complete registration flow with additional attributes")

	ts.Require().Equal("COMPLETE", completeFlowStep.FlowStatus, "Expected flow status to be COMPLETE")
	ts.Require().NotEmpty(completeFlowStep.Assertion, "JWT assertion should be returned")

	// Decode and validate JWT claims
	jwtClaims, err := testutils.DecodeJWT(completeFlowStep.Assertion)
	ts.Require().NoError(err, "Failed to decode JWT assertion")
	ts.Require().NotNil(jwtClaims, "JWT claims should not be nil")

	// Verify standard claims are present
	ts.Require().Equal(appID, jwtClaims.Aud, "JWT aud should match app ID")
	ts.Require().NotEmpty(jwtClaims.Sub, "JWT subject should not be empty")

	// Verify userType and OU attributes are NOT present (since TokenConfig is not specified)
	ts.Require().Empty(jwtClaims.UserType, "userType should NOT be present when TokenConfig is not specified")
	ts.Require().Empty(jwtClaims.OuID, "ouId should NOT be present when TokenConfig is not specified")
	ts.Require().Empty(jwtClaims.OuName, "ouName should NOT be present when TokenConfig is not specified")
	ts.Require().Empty(jwtClaims.OuHandle, "ouHandle should NOT be present when TokenConfig is not specified")
}

// TestBasicRegistrationFlow_WithEmptyUserAttributes tests that userType and OU attributes are NOT included
// in JWT assertion when user_attributes is an empty array.
func (ts *BasicRegistrationFlowTestSuite) TestBasicRegistrationFlow_WithEmptyUserAttributes() {
	// Create a new application with empty user_attributes
	appWithEmptyAttrs := testutils.Application{
		Name:                      "Registration Flow Test Application With Empty User Attributes",
		Description:               "Application for testing behavior with empty user_attributes",
		IsRegistrationFlowEnabled: true,
		ClientID:                  "reg_flow_test_client_empty_attrs",
		ClientSecret:              "reg_flow_test_secret_empty_attrs",
		RedirectURIs:              []string{"http://localhost:3000/callback"},
		AllowedUserTypes:          []string{testUserSchema.Name},
		AssertionConfig: map[string]interface{}{
			"user_attributes": []string{}, // Empty array
		},
	}

	appID, err := testutils.CreateApplication(appWithEmptyAttrs)
	ts.Require().NoError(err, "Failed to create application with empty user attributes")
	defer func() {
		if err := testutils.DeleteApplication(appID); err != nil {
			ts.T().Logf("Failed to delete test application: %v", err)
		}
	}()

	// Generate unique username for this test
	username := common.GenerateUniqueUsername("reguser")

	// Execute registration flow
	flowStep, err := common.InitiateRegistrationFlow(appID, false, nil, "")
	ts.Require().NoError(err, "Failed to initiate registration flow")

	inputs := map[string]string{
		"username": username,
		"password": "testpassword123",
	}
	completeFlowStep, err := common.CompleteFlow(flowStep.FlowID, inputs, "action_credentials")
	ts.Require().NoError(err, "Failed to complete registration flow")

	inputs = map[string]string{
		"email":     username + "@example.com",
		"firstName": "Test",
		"lastName":  "User",
	}
	completeFlowStep, err = common.CompleteFlow(completeFlowStep.FlowID, inputs, "action_user_info")
	ts.Require().NoError(err, "Failed to complete registration flow with additional attributes")

	ts.Require().Equal("COMPLETE", completeFlowStep.FlowStatus, "Expected flow status to be COMPLETE")
	ts.Require().NotEmpty(completeFlowStep.Assertion, "JWT assertion should be returned")

	// Decode and validate JWT claims
	jwtClaims, err := testutils.DecodeJWT(completeFlowStep.Assertion)
	ts.Require().NoError(err, "Failed to decode JWT assertion")
	ts.Require().NotNil(jwtClaims, "JWT claims should not be nil")

	// Verify standard claims are present
	ts.Require().Equal(appID, jwtClaims.Aud, "JWT aud should match app ID")
	ts.Require().NotEmpty(jwtClaims.Sub, "JWT subject should not be empty")

	// Verify userType and OU attributes are NOT present (since user_attributes is empty)
	ts.Require().Empty(jwtClaims.UserType, "userType should NOT be present when user_attributes is empty")
	ts.Require().Empty(jwtClaims.OuID, "ouId should NOT be present when user_attributes is empty")
	ts.Require().Empty(jwtClaims.OuName, "ouName should NOT be present when user_attributes is empty")
	ts.Require().Empty(jwtClaims.OuHandle, "ouHandle should NOT be present when user_attributes is empty")
}
