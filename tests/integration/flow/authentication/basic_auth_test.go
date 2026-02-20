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

package authentication

import (
	"encoding/json"
	"testing"

	"github.com/asgardeo/thunder/tests/integration/flow/common"
	"github.com/asgardeo/thunder/tests/integration/testutils"
	"github.com/stretchr/testify/suite"
)

var (
	testOU = testutils.OrganizationUnit{
		Handle:      "basicauth_flow_test_ou",
		Name:        "Test Organization Unit for BasicAuth Flow",
		Description: "Organization unit created for BasicAuth flow testing",
		Parent:      nil,
	}

	basicAuthTestFlow = testutils.Flow{
		Name:     "Basic Auth Test Auth Flow",
		FlowType: "AUTHENTICATION",
		Handle:   "auth_flow_basic_auth_test",
		Nodes: []map[string]interface{}{
			{
				"id":        "start",
				"type":      "START",
				"onSuccess": "prompt_credentials",
			},
			{
				"id":   "prompt_credentials",
				"type": "PROMPT",
				"prompts": []map[string]interface{}{
					{
						"inputs": []map[string]interface{}{
							{
								"ref":        "input_001",
								"identifier": "username",
								"type":       "TEXT_INPUT",
								"required":   true,
							},
							{
								"ref":        "input_002",
								"identifier": "password",
								"type":       "PASSWORD_INPUT",
								"required":   true,
							},
						},
						"action": map[string]interface{}{
							"ref":      "action_001",
							"nextNode": "basic_auth",
						},
					},
				},
			},
			{
				"id":   "basic_auth",
				"type": "TASK_EXECUTION",
				"executor": map[string]interface{}{
					"name": "BasicAuthExecutor",
					"inputs": []map[string]interface{}{
						{
							"ref":        "input_001",
							"identifier": "username",
							"type":       "TEXT_INPUT",
							"required":   true,
						},
						{
							"ref":        "input_002",
							"identifier": "password",
							"type":       "PASSWORD_INPUT",
							"required":   true,
						},
					},
				},
				"onSuccess": "auth_assert",
			},
			{
				"id":   "auth_assert",
				"type": "TASK_EXECUTION",
				"executor": map[string]interface{}{
					"name": "AuthAssertExecutor",
				},
				"onSuccess": "end",
			},
			{
				"id":   "end",
				"type": "END",
			},
		},
	}

	basicAuthWithoutPromptFlow = testutils.Flow{
		Name:     "Basic Auth Without Prompt Flow",
		FlowType: "AUTHENTICATION",
		Handle:   "auth_flow_basic_auth_without_prompt",
		Nodes: []map[string]interface{}{
			{
				"id":        "start",
				"type":      "START",
				"onSuccess": "basic_auth",
			},
			{
				"id":   "basic_auth",
				"type": "TASK_EXECUTION",
				"executor": map[string]interface{}{
					"name": "BasicAuthExecutor",
					"inputs": []map[string]interface{}{
						{
							"ref":        "input_001",
							"identifier": "username",
							"type":       "TEXT_INPUT",
							"required":   true,
						},
						{
							"ref":        "input_002",
							"identifier": "password",
							"type":       "PASSWORD_INPUT",
							"required":   true,
						},
					},
				},
				"onSuccess": "auth_assert",
			},
			{
				"id":   "auth_assert",
				"type": "TASK_EXECUTION",
				"executor": map[string]interface{}{
					"name": "AuthAssertExecutor",
				},
				"onSuccess": "end",
			},
			{
				"id":   "end",
				"type": "END",
			},
		},
	}

	testApp = testutils.Application{
		Name:                      "Flow Test Application",
		Description:               "Application for testing authentication flows",
		IsRegistrationFlowEnabled: false,
		ClientID:                  "flow_test_client",
		ClientSecret:              "flow_test_secret",
		RedirectURIs:              []string{"http://localhost:3000/callback"},
		AllowedUserTypes:          []string{"basic_auth_user"},
		AssertionConfig: map[string]interface{}{
			"user_attributes": []string{"userType", "ouId", "ouName", "ouHandle"},
		},
	}

	testUserSchema = testutils.UserSchema{
		Name: "basic_auth_user",
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
		},
	}

	testUser = testutils.User{
		Type: testUserSchema.Name,
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
	testAppID    string
	userSchemaID string
)

type BasicAuthFlowTestSuite struct {
	suite.Suite
	config *common.TestSuiteConfig
	ouID   string
}

func TestBasicAuthFlowTestSuite(t *testing.T) {
	suite.Run(t, new(BasicAuthFlowTestSuite))
}

func (ts *BasicAuthFlowTestSuite) SetupSuite() {
	// Initialize config
	ts.config = &common.TestSuiteConfig{}

	// Create test organization unit
	ouID, err := testutils.CreateOrganizationUnit(testOU)
	ts.Require().NoError(err, "Failed to create test organization unit")
	ts.ouID = ouID

	// Create test user schema
	testUserSchema.OrganizationUnitId = ts.ouID
	schemaID, err := testutils.CreateUserType(testUserSchema)
	if err != nil {
		ts.T().Fatalf("Failed to create test user schema during setup: %v", err)
	}
	userSchemaID = schemaID

	// Create flows
	flowID, err := testutils.CreateFlow(basicAuthTestFlow)
	ts.Require().NoError(err, "Failed to create basic auth test flow")
	ts.config.CreatedFlowIDs = append(ts.config.CreatedFlowIDs, flowID)
	testApp.AuthFlowID = flowID

	withoutPromptFlow, err := testutils.CreateFlow(basicAuthWithoutPromptFlow)
	ts.Require().NoError(err, "Failed to create basic auth without prompt flow")
	ts.config.CreatedFlowIDs = append(ts.config.CreatedFlowIDs, withoutPromptFlow)

	// Create test application
	appID, err := testutils.CreateApplication(testApp)
	if err != nil {
		ts.T().Fatalf("Failed to create test application during setup: %v", err)
	}
	testAppID = appID

	// Create test user with the created OU
	testUser := testUser
	testUser.OrganizationUnit = ts.ouID
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

	// Delete test flows
	for _, flowID := range ts.config.CreatedFlowIDs {
		if err := testutils.DeleteFlow(flowID); err != nil {
			ts.T().Logf("Failed to delete test flow %s during teardown: %v", flowID, err)
		}
	}

	if userSchemaID != "" {
		if err := testutils.DeleteUserType(userSchemaID); err != nil {
			ts.T().Logf("Failed to delete test user schema during teardown: %v", err)
		}
	}

	// Delete the test organization unit
	if ts.ouID != "" {
		err := testutils.DeleteOrganizationUnit(ts.ouID)
		if err != nil {
			ts.T().Logf("Failed to delete test organization unit during teardown: %v", err)
		}
	}

}

func (ts *BasicAuthFlowTestSuite) TestBasicAuthFlowSuccess() {
	// Update application
	err := common.UpdateAppConfig(testAppID, ts.config.CreatedFlowIDs[0], "")
	ts.NoError(err, "App config update should succeed")

	// Step 1: Initialize the flow by calling the flow execution API
	flowStep, err := common.InitiateAuthenticationFlow(testAppID, false, nil, "")
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
	ts.Require().True(common.ValidateRequiredInputs(flowStep.Data.Inputs, []string{"username", "password"}),
		"Username and password inputs should be required")
	ts.Require().True(common.HasInput(flowStep.Data.Inputs, "username"), "Username input should be present")
	ts.Require().True(common.HasInput(flowStep.Data.Inputs, "password"), "Password input should be present")

	// Step 2: Continue the flow with valid credentials
	var userAttrs map[string]interface{}
	err = json.Unmarshal(testUser.Attributes, &userAttrs)
	ts.Require().NoError(err, "Failed to unmarshal user attributes")

	inputs := map[string]string{
		"username": userAttrs["username"].(string),
		"password": userAttrs["password"].(string),
	}

	completeFlowStep, err := common.CompleteFlow(flowStep.FlowID, inputs, "action_001")
	if err != nil {
		ts.T().Fatalf("Failed to complete authentication flow: %v", err)
	}

	// Verify successful authentication
	ts.Require().Equal("COMPLETE", completeFlowStep.FlowStatus, "Expected flow status to be COMPLETE")
	ts.Require().NotEmpty(completeFlowStep.Assertion,
		"JWT assertion should be returned after successful authentication")
	ts.Require().Empty(completeFlowStep.FailureReason, "Failure reason should be empty for successful authentication")

	// Validate JWT assertion fields using common utility
	jwtClaims, err := testutils.ValidateJWTAssertionFields(
		completeFlowStep.Assertion,
		testAppID,
		testUserSchema.Name,
		ts.ouID,
		testOU.Name,
		testOU.Handle,
	)
	ts.Require().NoError(err, "Failed to validate JWT assertion fields")
	ts.Require().NotNil(jwtClaims, "JWT claims should not be nil")
}

func (ts *BasicAuthFlowTestSuite) TestBasicAuthFlowSuccessWithSingleRequest() {
	// Update application
	err := common.UpdateAppConfig(testAppID, ts.config.CreatedFlowIDs[1], "")
	ts.NoError(err, "App config update should succeed")

	// Step 1: Initialize the flow by calling the flow execution API with user credentials
	var userAttrs map[string]interface{}
	err = json.Unmarshal(testUser.Attributes, &userAttrs)
	ts.Require().NoError(err, "Failed to unmarshal user attributes")

	inputs := map[string]string{
		"username": userAttrs["username"].(string),
		"password": userAttrs["password"].(string),
	}

	flowStep, err := common.InitiateAuthenticationFlow(testAppID, false, inputs, "")
	if err != nil {
		ts.T().Fatalf("Failed to initiate authentication flow: %v", err)
	}

	// Verify successful authentication
	ts.Require().Equal("COMPLETE", flowStep.FlowStatus, "Expected flow status to be COMPLETE")
	ts.Require().Empty(flowStep.Data, "Flow should not require additional data after successful authentication")
	ts.Require().NotEmpty(flowStep.Assertion,
		"JWT assertion should be returned after successful authentication")
	ts.Require().Empty(flowStep.FailureReason, "Failure reason should be empty for successful authentication")

	// Validate JWT assertion fields using common utility
	jwtClaims, err := testutils.ValidateJWTAssertionFields(
		flowStep.Assertion,
		testAppID,
		testUserSchema.Name,
		ts.ouID,
		testOU.Name,
		testOU.Handle,
	)
	ts.Require().NoError(err, "Failed to validate JWT assertion fields")
	ts.Require().NotNil(jwtClaims, "JWT claims should not be nil")
}

func (ts *BasicAuthFlowTestSuite) TestBasicAuthFlowWithTwoStepInput() {
	// Update application
	err := common.UpdateAppConfig(testAppID, ts.config.CreatedFlowIDs[0], "")
	ts.NoError(err, "App config update should succeed")

	// Step 1: Initialize the flow
	flowStep, err := common.InitiateAuthenticationFlow(testAppID, false, nil, "")
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

	intermediateFlowStep, err := common.CompleteFlow(flowStep.FlowID, inputs, "action_001")
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
	ts.Require().True(common.HasInput(flowStep.Data.Inputs, "password"), "Password input should be required")

	// Step 3: Continue the flow with the password
	inputs = map[string]string{
		"password": userAttrs["password"].(string),
	}

	completeFlowStep, err := common.CompleteFlow(flowStep.FlowID, inputs, "action_001")
	if err != nil {
		ts.T().Fatalf("Failed to complete authentication flow: %v", err)
	}

	// Verify successful authentication
	ts.Require().Equal("COMPLETE", completeFlowStep.FlowStatus, "Expected flow status to be COMPLETE")
	ts.Require().NotEmpty(completeFlowStep.Assertion,
		"JWT assertion should be returned after successful authentication")
	ts.Require().Empty(completeFlowStep.FailureReason, "Failure reason should be empty for successful authentication")

	// Validate JWT assertion fields using common utility
	jwtClaims, err := testutils.ValidateJWTAssertionFields(
		completeFlowStep.Assertion,
		testAppID,
		testUserSchema.Name,
		ts.ouID,
		testOU.Name,
		testOU.Handle,
	)
	ts.Require().NoError(err, "Failed to validate JWT assertion fields")
	ts.Require().NotNil(jwtClaims, "JWT claims should not be nil")
}

func (ts *BasicAuthFlowTestSuite) TestBasicAuthFlowInvalidCredentials() {
	// Update application
	err := common.UpdateAppConfig(testAppID, ts.config.CreatedFlowIDs[0], "")
	ts.NoError(err, "App config update should succeed")

	// Step 1: Initialize the flow
	flowStep, err := common.InitiateAuthenticationFlow(testAppID, false, nil, "")
	if err != nil {
		ts.T().Fatalf("Failed to initiate authentication flow: %v", err)
	}

	ts.Require().NotEmpty(flowStep.FlowID, "Flow ID should not be empty")

	// Step 2: Continue with invalid credentials
	inputs := map[string]string{
		"username": "invalid_user",
		"password": "wrong_password",
	}

	completeFlowStep, err := common.CompleteFlow(flowStep.FlowID, inputs, "action_001")
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
	errorResp, err := common.InitiateAuthFlowWithError("invalid-app-id", nil)
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
	flowStep, err := common.InitiateAuthenticationFlow(testAppID, false, nil, "")
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

	errorResp, err := common.CompleteAuthFlowWithError("invalid-flow-id", inputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete authentication flow: %v", err)
	}

	// Verify the error response
	ts.Require().Equal("FES-1004", errorResp.Code, "Expected error code for invalid flow ID")
	ts.Require().Equal("Invalid request", errorResp.Message, "Expected error message for invalid request")
	ts.Require().Equal("Invalid flow ID provided in the request", errorResp.Description,
		"Expected error description for invalid flow ID")
}

// TestBasicAuthFlow_WithoutTokenConfig tests that userType and OU attributes are NOT included
// in JWT assertion when TokenConfig is not specified.
func (ts *BasicAuthFlowTestSuite) TestBasicAuthFlow_WithoutTokenConfig() {
	// Create a new application without TokenConfig
	appWithoutTokenConfig := testutils.Application{
		Name:                      "Flow Test Application Without Token Config",
		Description:               "Application for testing default behavior without token config",
		IsRegistrationFlowEnabled: false,
		ClientID:                  "flow_test_client_no_token_config",
		ClientSecret:              "flow_test_secret_no_token_config",
		RedirectURIs:              []string{"http://localhost:3000/callback"},
		AllowedUserTypes:          []string{"basic_auth_user"},
		// TokenConfig is nil - not specified
	}

	appID, err := testutils.CreateApplication(appWithoutTokenConfig)
	ts.Require().NoError(err, "Failed to create application without token config")
	defer func() {
		if err := testutils.DeleteApplication(appID); err != nil {
			ts.T().Logf("Failed to delete test application: %v", err)
		}
	}()

	// Update application with flow
	err = common.UpdateAppConfig(appID, ts.config.CreatedFlowIDs[0], "")
	ts.Require().NoError(err, "App config update should succeed")

	// Execute authentication flow
	var userAttrs map[string]interface{}
	err = json.Unmarshal(testUser.Attributes, &userAttrs)
	ts.Require().NoError(err, "Failed to unmarshal user attributes")

	inputs := map[string]string{
		"username": userAttrs["username"].(string),
		"password": userAttrs["password"].(string),
	}

	flowStep, err := common.InitiateAuthenticationFlow(appID, false, inputs, "")
	ts.Require().NoError(err, "Failed to initiate authentication flow")

	ts.Require().Equal("COMPLETE", flowStep.FlowStatus, "Expected flow status to be COMPLETE")
	ts.Require().NotEmpty(flowStep.Assertion, "JWT assertion should be returned")

	// Decode and validate JWT claims
	jwtClaims, err := testutils.DecodeJWT(flowStep.Assertion)
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

// TestBasicAuthFlow_WithEmptyUserAttributes tests that userType and OU attributes are NOT included
// in JWT assertion when user_attributes is an empty array.
func (ts *BasicAuthFlowTestSuite) TestBasicAuthFlow_WithEmptyUserAttributes() {
	// Create a new application with empty user_attributes
	appWithEmptyAttrs := testutils.Application{
		Name:                      "Flow Test Application With Empty User Attributes",
		Description:               "Application for testing behavior with empty user_attributes",
		IsRegistrationFlowEnabled: false,
		ClientID:                  "flow_test_client_empty_attrs",
		ClientSecret:              "flow_test_secret_empty_attrs",
		RedirectURIs:              []string{"http://localhost:3000/callback"},
		AllowedUserTypes:          []string{"basic_auth_user"},
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

	// Update application with flow
	err = common.UpdateAppConfig(appID, ts.config.CreatedFlowIDs[0], "")
	ts.Require().NoError(err, "App config update should succeed")

	// Execute authentication flow
	var userAttrs map[string]interface{}
	err = json.Unmarshal(testUser.Attributes, &userAttrs)
	ts.Require().NoError(err, "Failed to unmarshal user attributes")

	inputs := map[string]string{
		"username": userAttrs["username"].(string),
		"password": userAttrs["password"].(string),
	}

	flowStep, err := common.InitiateAuthenticationFlow(appID, false, inputs, "")
	ts.Require().NoError(err, "Failed to initiate authentication flow")

	ts.Require().Equal("COMPLETE", flowStep.FlowStatus, "Expected flow status to be COMPLETE")
	ts.Require().NotEmpty(flowStep.Assertion, "JWT assertion should be returned")

	// Decode and validate JWT claims
	jwtClaims, err := testutils.DecodeJWT(flowStep.Assertion)
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
