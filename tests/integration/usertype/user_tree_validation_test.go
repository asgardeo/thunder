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

package usertype

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/asgardeo/thunder/tests/integration/testutils"
	"github.com/stretchr/testify/suite"
)

type UserTreeValidationTestSuite struct {
	suite.Suite
	client       *http.Client
	createdTypes []string // Track types for cleanup
	createdUsers []string // Track users for cleanup
	createdOUs   []string // Track organization units for cleanup
	ou1ID        string   // ID of ou1
	ou2ID        string   // ID of ou2 (child of ou1)
}

var testUserTreeValidationOU = testutils.OrganizationUnit{
	Handle:      "test-user-tree-validation-ou",
	Name:        "Test Organization Unit for User Tree Validation",
	Description: "Organization unit created for user tree validation testing",
	Parent:      nil,
}

func TestUserTreeValidationTestSuite(t *testing.T) {
	suite.Run(t, new(UserTreeValidationTestSuite))
}

func (ts *UserTreeValidationTestSuite) SetupSuite() {
	ts.client = testutils.GetHTTPClient()
	ts.createdTypes = []string{}
	ts.createdUsers = []string{}
	ts.createdOUs = []string{}

	// Create organization units for testing
	ts.createOrganizationUnits()
}

func (ts *UserTreeValidationTestSuite) TearDownSuite() {
	// Clean up created users first
	for _, userID := range ts.createdUsers {
		ts.deleteUser(userID)
	}
	// Then clean up created types
	for _, typeID := range ts.createdTypes {
		ts.deleteType(typeID)
	}
	// Finally clean up created organization units in reverse order (children first)
	for i := len(ts.createdOUs) - 1; i >= 0; i-- {
		if err := testutils.DeleteOrganizationUnit(ts.createdOUs[i]); err != nil {
			ts.T().Logf("Failed to delete OU %s: %v", ts.createdOUs[i], err)
		}
	}
}

// TestCreateUserByPathWithValidUserType tests user creation via tree API with valid type
func (ts *UserTreeValidationTestSuite) TestCreateUserByPathWithValidUserType() {
	// First create a user type
	_, typeName := ts.createEmployeeType()

	// Create a user that conforms to the type via tree API
	createUserReq := CreateUserByPathRequest{
		Type: typeName,
		Attributes: json.RawMessage(`{
			"firstName": "Alice",
			"lastName": "Johnson",
			"email": "alice.johnson@company.com",
			"department": "Marketing",
			"isManager": true
		}`),
	}

	userID := ts.createUserByPathAndExpectSuccess("ou1/ou2", createUserReq)
	ts.createdUsers = append(ts.createdUsers, userID)
}

// TestCreateUserByPathWithInvalidUserType tests user creation via tree API with invalid type
func (ts *UserTreeValidationTestSuite) TestCreateUserByPathWithInvalidUserType() {
	// Create a user type
	_, typeName := ts.createEmployeeType()

	// Create a user with invalid data (wrong type for firstName)
	createUserReq := CreateUserByPathRequest{
		Type: typeName,
		Attributes: json.RawMessage(`{
			"firstName": 456,
			"lastName": "Smith",
			"email": "invalid@company.com",
			"department": "HR",
			"isManager": false
		}`),
	}

	ts.createUserByPathAndExpectError("ou1/ou2", createUserReq, "USR-1019")
}

// TestCreateUserByPathWithComplexUserType tests user creation with complex nested type
func (ts *UserTreeValidationTestSuite) TestCreateUserByPathWithComplexUserType() {
	// Create a complex type
	_, typeName := ts.createComplexType()

	// Test valid complex data
	createUserReq := CreateUserByPathRequest{
		Type: typeName,
		Attributes: json.RawMessage(`{
			"name": "Sarah Wilson",
			"profile": {
				"bio": "Experienced manager",
				"skills": ["leadership", "strategy", "communication"],
				"ratings": {
					"performance": 4.8,
					"teamwork": 4.9
				}
			},
			"teams": ["engineering", "product"]
		}`),
	}

	userID := ts.createUserByPathAndExpectSuccess("ou1/ou2", createUserReq)
	ts.createdUsers = append(ts.createdUsers, userID)

	// Test invalid complex data (wrong type in nested array)
	createUserReq2 := CreateUserByPathRequest{
		Type: typeName,
		Attributes: json.RawMessage(`{
			"name": "Bob Johnson",
			"profile": {
				"bio": "Another manager",
				"skills": ["leadership", 123, "communication"],
				"ratings": {
					"performance": 4.5,
					"teamwork": 4.7
				}
			},
			"teams": ["marketing", "sales"]
		}`),
	}

	ts.createUserByPathAndExpectError("ou1/ou2", createUserReq2, "USR-1019")
}

// Helper methods

func (ts *UserTreeValidationTestSuite) getUniqueName(baseName string) string {
	return fmt.Sprintf("%s_%d", baseName, time.Now().UnixNano())
}

func (ts *UserTreeValidationTestSuite) createEmployeeType() (string, string) {
	typeName := ts.getUniqueName("employee")
	userType := CreateUserTypeRequest{
		Name: typeName,
		Schema: json.RawMessage(`{
			"firstName": {"type": "string"},
			"lastName": {"type": "string"},
			"email": {"type": "string"},
			"department": {"type": "string"},
			"isManager": {"type": "boolean"}
		}`),
	}
	userType.OrganizationUnitID = ts.ou1ID

	typeID := ts.createUserType(userType)
	return typeID, typeName
}

func (ts *UserTreeValidationTestSuite) createComplexType() (string, string) {
	typeName := ts.getUniqueName("manager")
	userType := CreateUserTypeRequest{
		Name: typeName,
		Schema: json.RawMessage(`{
			"name": {"type": "string"},
			"profile": {
				"type": "object",
				"properties": {
					"bio": {"type": "string"},
					"skills": {
						"type": "array",
						"items": {"type": "string"}
					},
					"ratings": {
						"type": "object",
						"properties": {
							"performance": {"type": "number"},
							"teamwork": {"type": "number"}
						}
					}
				}
			},
			"teams": {
				"type": "array",
				"items": {"type": "string"}
			}
		}`),
	}
	userType.OrganizationUnitID = ts.ou1ID

	typeID := ts.createUserType(userType)
	return typeID, typeName
}

func (ts *UserTreeValidationTestSuite) createUserType(userType CreateUserTypeRequest) string {
	jsonData, err := json.Marshal(userType)
	ts.Require().NoError(err, "Failed to marshal type request")

	req, err := http.NewRequest("POST", testServerURL+"/user-types", bytes.NewBuffer(jsonData))
	ts.Require().NoError(err, "Failed to create type request")
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.client.Do(req)
	ts.Require().NoError(err, "Failed to send type request")
	defer resp.Body.Close()

	ts.Require().Equal(201, resp.StatusCode, "Type creation should succeed")

	body, err := io.ReadAll(resp.Body)
	ts.Require().NoError(err, "Failed to read type response body")

	var createdType UserType
	err = json.Unmarshal(body, &createdType)
	ts.Require().NoError(err, "Failed to unmarshal type response")

	ts.createdTypes = append(ts.createdTypes, createdType.ID)
	return createdType.ID
}

func (ts *UserTreeValidationTestSuite) createUserByPathAndExpectSuccess(path string, createUserReq CreateUserByPathRequest) string {
	jsonData, err := json.Marshal(createUserReq)
	ts.Require().NoError(err, "Failed to marshal user request")

	req, err := http.NewRequest("POST", testServerURL+"/users/tree/"+path, bytes.NewBuffer(jsonData))
	ts.Require().NoError(err, "Failed to create user request")
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.client.Do(req)
	ts.Require().NoError(err, "Failed to send user request")
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	ts.Require().NoError(err, "Failed to read user response body")

	if resp.StatusCode != 201 {
		ts.T().Logf("User creation failed with status %d: %s", resp.StatusCode, string(body))
	}
	ts.Require().Equal(201, resp.StatusCode, "User creation should succeed")

	var createdUser testutils.User
	err = json.Unmarshal(body, &createdUser)
	ts.Require().NoError(err, "Failed to unmarshal user response")

	return createdUser.ID
}

func (ts *UserTreeValidationTestSuite) createUserByPathAndExpectError(path string, createUserReq CreateUserByPathRequest, expectedErrorCode string) {
	jsonData, err := json.Marshal(createUserReq)
	ts.Require().NoError(err, "Failed to marshal user request")

	req, err := http.NewRequest("POST", testServerURL+"/users/tree/"+path, bytes.NewBuffer(jsonData))
	ts.Require().NoError(err, "Failed to create user request")
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.client.Do(req)
	ts.Require().NoError(err, "Failed to send user request")
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	ts.Require().NoError(err, "Failed to read user response body")

	ts.Require().Equal(400, resp.StatusCode, "User creation should fail with validation error")

	var errorResp ErrorResponse
	err = json.Unmarshal(body, &errorResp)
	ts.Require().NoError(err, "Failed to unmarshal error response")
	ts.Require().Equal(expectedErrorCode, errorResp.Code, "Error code should match expected")
}

func (ts *UserTreeValidationTestSuite) deleteUser(userID string) {
	req, err := http.NewRequest("DELETE", testServerURL+"/users/"+userID, nil)
	if err != nil {
		ts.T().Logf("Failed to create delete user request: %v", err)
		return
	}

	resp, err := ts.client.Do(req)
	if err != nil {
		ts.T().Logf("Failed to send delete user request: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 && resp.StatusCode != 404 {
		body, _ := io.ReadAll(resp.Body)
		ts.T().Logf("Failed to delete user %s: status %d, body: %s", userID, resp.StatusCode, string(body))
	}
}

func (ts *UserTreeValidationTestSuite) deleteType(typeID string) {
	req, err := http.NewRequest("DELETE", testServerURL+"/user-types/"+typeID, nil)
	if err != nil {
		ts.T().Logf("Failed to create delete type request: %v", err)
		return
	}

	resp, err := ts.client.Do(req)
	if err != nil {
		ts.T().Logf("Failed to send delete type request: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 && resp.StatusCode != 404 {
		body, _ := io.ReadAll(resp.Body)
		ts.T().Logf("Failed to delete type %s: status %d, body: %s", typeID, resp.StatusCode, string(body))
	}
}

func (ts *UserTreeValidationTestSuite) createOrganizationUnits() {
	parentOU := testutils.OrganizationUnit{
		Handle:      "ou1",
		Name:        "Organization Unit 1",
		Description: "Test OU 1 for type validation",
	}

	ou1ID, err := testutils.CreateOrganizationUnit(parentOU)
	ts.Require().NoError(err, "Failed to create parent organization unit")
	ts.ou1ID = ou1ID
	ts.createdOUs = append(ts.createdOUs, ou1ID)

	childOU := testutils.OrganizationUnit{
		Handle:      "ou2",
		Name:        "Organization Unit 2",
		Description: "Test OU 2 for type validation",
		Parent:      &ts.ou1ID,
	}

	ou2ID, err := testutils.CreateOrganizationUnit(childOU)
	ts.Require().NoError(err, "Failed to create child organization unit")
	ts.ou2ID = ou2ID
	ts.createdOUs = append(ts.createdOUs, ou2ID)
}
