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
	mockHTTPServerPort = 9091
)

var (
	httpRequestTestApp = testutils.Application{
		Name:                      "HTTP Request Executor Test Application",
		Description:               "Application for testing HTTP request executor in authentication flows",
		AuthFlowGraphID:           "auth_flow_config_basic_http_request",
		RegistrationFlowGraphID:   "registration_flow_config_basic",
		IsRegistrationFlowEnabled: false,
		AllowedUserTypes:          []string{"http_request_test_person"},
	}

	httpRequestTestOU = testutils.OrganizationUnit{
		Name:        "HTTP Request Test OU",
		Handle:      "http-request-test-ou",
		Description: "OU for HTTP request executor authentication tests",
	}

	httpRequestTestUserSchema = testutils.UserSchema{
		Name: "http_request_test_person",
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

	httpRequestTestUser = testutils.User{
		Type: "http_request_test_person",
		Attributes: json.RawMessage(`{
			"username": "httprequestuser",
			"password": "SecurePass123!",
			"email": "httprequest@test.com",
			"firstName": "HTTP",
			"lastName": "User"
		}`),
	}
)

var (
	httpRequestTestAppID    string
	httpRequestTestOUID     string
	httpRequestUserSchemaID string
)

type HTTPRequestAuthFlowTestSuite struct {
	suite.Suite
	config     *TestSuiteConfig
	mockServer *testutils.MockHTTPServer
}

func TestHTTPRequestAuthFlowTestSuite(t *testing.T) {
	suite.Run(t, new(HTTPRequestAuthFlowTestSuite))
}

func (ts *HTTPRequestAuthFlowTestSuite) SetupSuite() {
	// Initialize config
	ts.config = &TestSuiteConfig{}

	// Create test organization unit
	ouID, err := testutils.CreateOrganizationUnit(httpRequestTestOU)
	if err != nil {
		ts.T().Fatalf("Failed to create test organization unit during setup: %v", err)
	}
	httpRequestTestOUID = ouID

	// Create test user schema within the OU
	httpRequestTestUserSchema.OrganizationUnitId = httpRequestTestOUID
	schemaID, err := testutils.CreateUserType(httpRequestTestUserSchema)
	if err != nil {
		ts.T().Fatalf("Failed to create test user schema during setup: %v", err)
	}
	httpRequestUserSchemaID = schemaID

	// Create test application
	appID, err := testutils.CreateApplication(httpRequestTestApp)
	if err != nil {
		ts.T().Fatalf("Failed to create test application during setup: %v", err)
	}
	httpRequestTestAppID = appID

	// Start mock HTTP server
	ts.mockServer = testutils.NewMockHTTPServer(mockHTTPServerPort)
	err = ts.mockServer.Start()
	if err != nil {
		ts.T().Fatalf("Failed to start mock HTTP server: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
	ts.T().Log("Mock HTTP server started successfully")

	// Create test user with the created OU
	testUser := httpRequestTestUser
	testUser.OrganizationUnit = httpRequestTestOUID
	userIDs, err := testutils.CreateMultipleUsers(testUser)
	if err != nil {
		ts.T().Fatalf("Failed to create test user during setup: %v", err)
	}
	ts.config.CreatedUserIDs = userIDs
	ts.T().Logf("Test user created with ID: %s", ts.config.CreatedUserIDs[0])
}

func (ts *HTTPRequestAuthFlowTestSuite) TearDownSuite() {
	// Stop the mock HTTP server
	if ts.mockServer != nil {
		err := ts.mockServer.Stop()
		if err != nil {
			ts.T().Logf("Failed to stop mock HTTP server during teardown: %v", err)
		}
	}

	// Delete all created users
	if err := testutils.CleanupUsers(ts.config.CreatedUserIDs); err != nil {
		ts.T().Logf("Failed to cleanup users during teardown: %v", err)
	}

	// Delete test application
	if httpRequestTestAppID != "" {
		if err := testutils.DeleteApplication(httpRequestTestAppID); err != nil {
			ts.T().Logf("Failed to delete test application during teardown: %v", err)
		}
	}

	// Delete test organization unit
	if httpRequestTestOUID != "" {
		if err := testutils.DeleteOrganizationUnit(httpRequestTestOUID); err != nil {
			ts.T().Logf("Failed to delete test organization unit during teardown: %v", err)
		}
	}

	// Delete test user schema
	if httpRequestUserSchemaID != "" {
		if err := testutils.DeleteUserType(httpRequestUserSchemaID); err != nil {
			ts.T().Logf("Failed to delete test user schema during teardown: %v", err)
		}
	}
}

func (ts *HTTPRequestAuthFlowTestSuite) SetupTest() {
	// Clear captured requests before each test
	if ts.mockServer != nil {
		ts.mockServer.ClearRequests()
	}
}

func (ts *HTTPRequestAuthFlowTestSuite) TestHTTPRequestAuthFlow_Success() {
	step1, err := initiateAuthFlow(httpRequestTestAppID, map[string]string{
		"username": "httprequestuser",
		"password": "SecurePass123!",
	})

	ts.NoError(err, "Authentication flow should complete without error")
	ts.NotNil(step1, "Flow response should not be nil")
	ts.Equal("COMPLETE", step1.FlowStatus, "Flow status should be COMPLETE")
	ts.Require().NotEmpty(step1.Assertion, "JWT assertion should be returned")
	ts.Require().Empty(step1.FailureReason, "Failure reason should be empty")

	time.Sleep(200 * time.Millisecond)

	requests := ts.mockServer.GetCapturedRequests()
	ts.NotEmpty(requests, "At least one HTTP request should be captured")

	var notificationRequest *testutils.HTTPRequest
	for _, req := range requests {
		if req.Path == "/api/notifications" {
			notificationRequest = &req
			break
		}
	}

	ts.NotNil(notificationRequest, "Notification request should be sent")
	ts.Equal("POST", notificationRequest.Method, "Request method should be POST")
	ts.NotNil(notificationRequest.Body, "Request body should not be nil")
	ts.NotEmpty(notificationRequest.Headers["Content-Type"], "Content-Type header should be present")
	ts.Contains(notificationRequest.Headers["Content-Type"], "application/json",
		"Content-Type should be application/json")

	ts.Equal("httprequestuser", notificationRequest.Body["username"], "Username should match")
	ts.Equal("user_authenticated", notificationRequest.Body["event"], "Event should be user_authenticated")
	ts.NotEmpty(notificationRequest.Body["userId"], "User ID should be present in payload")
	ts.Equal(ts.config.CreatedUserIDs[0], notificationRequest.Body["userId"],
		"User ID should match the authenticated user")
	ts.Equal("{{ context.unknownPlaceholder }}", notificationRequest.Body["unknownField"],
		"Unknown field should retain the placeholder value")
}

func (ts *HTTPRequestAuthFlowTestSuite) TestHTTPRequestAuthFlow_WithFailOnErrorFalse() {
	err := updateAppConfig(httpRequestTestAppID, "auth_flow_config_http_request_error_continue")
	ts.NoError(err, "App config update should succeed")

	defer func() {
		updateAppConfig(httpRequestTestAppID, "auth_flow_config_basic_http_request")
	}()

	step1, err := initiateAuthFlow(httpRequestTestAppID, map[string]string{
		"username": "httprequestuser",
		"password": "SecurePass123!",
	})

	ts.NoError(err, "Authentication flow should complete without error")
	ts.NotNil(step1, "Flow response should not be nil")
	ts.Equal("COMPLETE", step1.FlowStatus, "Flow status should be COMPLETE")
	ts.NotEmpty(step1.Assertion, "JWT assertion should be returned")
	ts.Empty(step1.FailureReason, "Failure reason should be empty")
}

func (ts *HTTPRequestAuthFlowTestSuite) TestHTTPRequestAuthFlow_WithFailOnErrorTrue() {
	err := updateAppConfig(httpRequestTestAppID, "auth_flow_config_http_request_error_fail")
	ts.NoError(err, "App config update should succeed")

	defer func() {
		updateAppConfig(httpRequestTestAppID, "auth_flow_config_basic_http_request")
	}()

	_, err = initiateAuthFlowWithError(httpRequestTestAppID, map[string]string{
		"username": "httprequestuser",
		"password": "SecurePass123!",
	})

	ts.Require().Error(err, "HTTP request failure should cause authentication flow to fail")
}
