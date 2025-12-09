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
	"testing"
	"time"

	"github.com/asgardeo/thunder/tests/integration/testutils"
	"github.com/stretchr/testify/suite"
)

const (
	mockHTTPServerPortReg = 9091
)

var (
	httpRequestRegTestOU = testutils.OrganizationUnit{
		Name:        "HTTP Request Registration Test OU",
		Handle:      "http-request-reg-test-ou",
		Description: "OU for HTTP request executor registration tests",
	}

	httpRequestRegTestUserSchema = testutils.UserSchema{
		Name: "http_request_reg_test_person",
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

	httpRequestRegTestApp = testutils.Application{
		Name:                      "HTTP Request Executor Registration Test Application",
		Description:               "Application for testing HTTP request executor in registration flows",
		AuthFlowGraphID:           "auth_flow_config_basic",
		RegistrationFlowGraphID:   "registration_flow_config_basic_http_request",
		IsRegistrationFlowEnabled: true,
		AllowedUserTypes:          []string{httpRequestRegTestUserSchema.Name},
	}
)

var (
	httpRequestRegTestAppID    string
	httpRequestRegTestOUID     string
	httpRequestRegUserSchemaID string
)

type HTTPRequestRegistrationFlowTestSuite struct {
	suite.Suite
	config     *TestSuiteConfig
	mockServer *testutils.MockHTTPServer
}

func TestHTTPRequestRegistrationFlowTestSuite(t *testing.T) {
	suite.Run(t, new(HTTPRequestRegistrationFlowTestSuite))
}

func (ts *HTTPRequestRegistrationFlowTestSuite) SetupSuite() {
	// Initialize config
	ts.config = &TestSuiteConfig{}

	// Create test organization unit
	ouID, err := testutils.CreateOrganizationUnit(httpRequestRegTestOU)
	if err != nil {
		ts.T().Fatalf("Failed to create test organization unit during setup: %v", err)
	}
	httpRequestRegTestOUID = ouID

	// Create test user schema within the test OU
	httpRequestRegTestUserSchema.OrganizationUnitId = httpRequestRegTestOUID
	httpRequestRegTestUserSchema.AllowSelfRegistration = true
	schemaID, err := testutils.CreateUserType(httpRequestRegTestUserSchema)
	if err != nil {
		ts.T().Fatalf("Failed to create test user schema during setup: %v", err)
	}
	httpRequestRegUserSchemaID = schemaID

	// Create test application
	appID, err := testutils.CreateApplication(httpRequestRegTestApp)
	if err != nil {
		ts.T().Fatalf("Failed to create test application during setup: %v", err)
	}
	httpRequestRegTestAppID = appID

	// Start mock HTTP server
	ts.mockServer = testutils.NewMockHTTPServer(mockHTTPServerPortReg)
	err = ts.mockServer.Start()
	if err != nil {
		ts.T().Fatalf("Failed to start mock HTTP server: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
	ts.T().Log("Mock HTTP server started successfully")
}

func (ts *HTTPRequestRegistrationFlowTestSuite) TearDownSuite() {
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
	if httpRequestRegTestAppID != "" {
		if err := testutils.DeleteApplication(httpRequestRegTestAppID); err != nil {
			ts.T().Logf("Failed to delete test application during teardown: %v", err)
		}
	}

	// Delete test organization unit
	if httpRequestRegTestOUID != "" {
		if err := testutils.DeleteOrganizationUnit(httpRequestRegTestOUID); err != nil {
			ts.T().Logf("Failed to delete test organization unit during teardown: %v", err)
		}
	}

	// Delete test user schema
	if httpRequestRegUserSchemaID != "" {
		if err := testutils.DeleteUserType(httpRequestRegUserSchemaID); err != nil {
			ts.T().Logf("Failed to delete test user schema during teardown: %v", err)
		}
	}
}

func (ts *HTTPRequestRegistrationFlowTestSuite) SetupTest() {
	// Clear captured requests before each test
	if ts.mockServer != nil {
		ts.mockServer.ClearRequests()
	}
}

func (ts *HTTPRequestRegistrationFlowTestSuite) TestHTTPRequestRegistrationFlow_Success() {
	step1, err := initiateRegistrationFlow(httpRequestRegTestAppID, false, map[string]string{
		"username":  "newuser123",
		"password":  "NewUserPass123!",
		"email":     "newuser@test.com",
		"firstName": "New",
		"lastName":  "User",
	}, "")

	ts.NoError(err, "Registration flow should complete without error")
	ts.NotNil(step1, "Flow response should not be nil")
	ts.Equal("COMPLETE", step1.FlowStatus, "Flow status should be COMPLETE")

	time.Sleep(200 * time.Millisecond)

	requests := ts.mockServer.GetCapturedRequests()
	ts.NotEmpty(requests, "At least one HTTP request should be captured")

	var userCreationRequest *testutils.HTTPRequest
	for _, req := range requests {
		if req.Path == "/api/users" {
			userCreationRequest = &req
			break
		}
	}

	ts.NotNil(userCreationRequest, "User creation request should be sent")
	ts.Equal("POST", userCreationRequest.Method, "Request method should be POST")
	ts.NotNil(userCreationRequest.Body, "Request body should not be nil")
	ts.NotEmpty(userCreationRequest.Headers["Content-Type"], "Content-Type header should be present")
	ts.Contains(userCreationRequest.Headers["Content-Type"], "application/json",
		"Content-Type should be application/json")
	ts.NotEmpty(userCreationRequest.Headers["Authorization"], "Authorization header should be present")
	ts.Contains(userCreationRequest.Headers["Authorization"], "Bearer test-token",
		"Authorization header should contain bearer token")

	ts.Equal("newuser123", userCreationRequest.Body["username"], "Username should match")
	ts.Equal("newuser@test.com", userCreationRequest.Body["email"], "Email should match")
	ts.Equal("New", userCreationRequest.Body["firstName"], "First name should match")
	ts.Equal("User", userCreationRequest.Body["lastName"], "Last name should match")
	ts.NotEmpty(userCreationRequest.Body["externalId"], "External ID should be present in payload")
	ts.Equal("{{ context.unknownPlaceholder }}", userCreationRequest.Body["unknownField"],
		"Unknown field should retain the placeholder value")
}
