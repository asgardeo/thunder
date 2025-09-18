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

package userschema

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/asgardeo/thunder/tests/integration/testutils"
	"github.com/stretchr/testify/suite"
)

// UserSchemaRequiredAPITestSuite contains API tests for validating the required attribute behavior.
type UserSchemaRequiredAPITestSuite struct {
	suite.Suite
	client         *http.Client
	createdSchemas []string
	createdUsers   []string
}

func TestUserSchemaRequiredAPITestSuite(t *testing.T) {
	suite.Run(t, new(UserSchemaRequiredAPITestSuite))
}

func (ts *UserSchemaRequiredAPITestSuite) SetupSuite() {
	ts.client = &http.Client{
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	ts.createdSchemas = []string{}
	ts.createdUsers = []string{}
}

func (ts *UserSchemaRequiredAPITestSuite) TearDownSuite() {
	for _, userID := range ts.createdUsers {
		ts.deleteUser(userID)
	}
	for _, schemaID := range ts.createdSchemas {
		ts.deleteSchema(schemaID)
	}
}

// Test top-level required string attribute.
func (ts *UserSchemaRequiredAPITestSuite) TestRequiredTopLevelString() {
	schema := CreateUserSchemaRequest{
		Name: "req-top-level-string",
		Schema: json.RawMessage(`{
            "email": {"type": "string", "required": true},
            "nickname": {"type": "string"}
        }`),
	}
	schemaID := ts.createSchemaHelper(schema)
	ts.createdSchemas = append(ts.createdSchemas, schemaID)

	// Missing required email -> expect validation error USR-1019
	reqMissing := CreateUserRequest{
		OrganizationUnit: "456e8400-e29b-41d4-a716-446655440001",
		Type:             schema.Name,
		Attributes:       json.RawMessage(`{"nickname": "neo"}`),
	}
	ts.createUserAndExpectError(reqMissing, "USR-1019")

	// Provide required email -> success
	reqPresent := CreateUserRequest{
		OrganizationUnit: "456e8400-e29b-41d4-a716-446655440001",
		Type:             schema.Name,
		Attributes:       json.RawMessage(`{"email": "a@b.com"}`),
	}
	userID := ts.createUserAndExpectSuccess(reqPresent)
	ts.createdUsers = append(ts.createdUsers, userID)
}

// Test required object attribute with nested required property.
func (ts *UserSchemaRequiredAPITestSuite) TestRequiredObjectAndNested() {
	schema := CreateUserSchemaRequest{
		Name: "req-object-nested",
		Schema: json.RawMessage(`{
            "address": {
                "type": "object",
                "required": true,
                "properties": {
                    "city": {"type": "string", "required": true},
                    "zip": {"type": "string"}
                }
            }
        }`),
	}
	schemaID := ts.createSchemaHelper(schema)
	ts.createdSchemas = append(ts.createdSchemas, schemaID)

	// Missing required object -> fail
	reqMissingObj := CreateUserRequest{
		OrganizationUnit: "456e8400-e29b-41d4-a716-446655440001",
		Type:             schema.Name,
		Attributes:       json.RawMessage(`{}`),
	}
	ts.createUserAndExpectError(reqMissingObj, "USR-1019")

	// Object present, missing required nested city -> fail
	reqMissingNested := CreateUserRequest{
		OrganizationUnit: "456e8400-e29b-41d4-a716-446655440001",
		Type:             schema.Name,
		Attributes:       json.RawMessage(`{"address": {"zip": "94040"}}`),
	}
	ts.createUserAndExpectError(reqMissingNested, "USR-1019")

	// Provide required nested city -> success
	reqOK := CreateUserRequest{
		OrganizationUnit: "456e8400-e29b-41d4-a716-446655440001",
		Type:             schema.Name,
		Attributes:       json.RawMessage(`{"address": {"city": "Colombo"}}`),
	}
	userID := ts.createUserAndExpectSuccess(reqOK)
	ts.createdUsers = append(ts.createdUsers, userID)
}

// Test required array attribute at top level.
func (ts *UserSchemaRequiredAPITestSuite) TestRequiredArrayTopLevel() {
	schema := CreateUserSchemaRequest{
		Name: "req-array-top-level",
		Schema: json.RawMessage(`{
            "tags": {"type": "array", "required": true, "items": {"type": "string"}}
        }`),
	}
	schemaID := ts.createSchemaHelper(schema)
	ts.createdSchemas = append(ts.createdSchemas, schemaID)

	// Missing required array -> fail
	reqMissing := CreateUserRequest{
		OrganizationUnit: "456e8400-e29b-41d4-a716-446655440001",
		Type:             schema.Name,
		Attributes:       json.RawMessage(`{}`),
	}
	ts.createUserAndExpectError(reqMissing, "USR-1019")

	// Present empty array -> fail
	reqEmpty := CreateUserRequest{
		OrganizationUnit: "456e8400-e29b-41d4-a716-446655440001",
		Type:             schema.Name,
		Attributes:       json.RawMessage(`{"tags": []}`),
	}
	ts.createUserAndExpectError(reqEmpty, "USR-1019")

	// Present array with items -> success
	reqWithItems := CreateUserRequest{
		OrganizationUnit: "456e8400-e29b-41d4-a716-446655440001",
		Type:             schema.Name,
		Attributes:       json.RawMessage(`{"tags": ["tag1", "tag2"]}`),
	}
	userID := ts.createUserAndExpectSuccess(reqWithItems)
	ts.createdUsers = append(ts.createdUsers, userID)
}

func (ts *UserSchemaRequiredAPITestSuite) createSchemaHelper(schema CreateUserSchemaRequest) string {
	jsonData, err := json.Marshal(schema)
	ts.Require().NoError(err, "Failed to marshal request")

	req, err := http.NewRequest("POST", testServerURL+"/user-schemas", bytes.NewBuffer(jsonData))
	ts.Require().NoError(err, "Failed to create request")
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.client.Do(req)
	ts.Require().NoError(err, "Failed to send request")
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	ts.Require().NoError(err, "Failed to read response body")

	if resp.StatusCode != http.StatusCreated {
		ts.T().Logf("Create schema failed: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}
	ts.Require().Equal(http.StatusCreated, resp.StatusCode, "Should return 201 Created")

	var createdSchema UserSchema
	ts.Require().NoError(json.Unmarshal(bodyBytes, &createdSchema))
	return createdSchema.ID
}

func (ts *UserSchemaRequiredAPITestSuite) createUserAndExpectSuccess(reqBody CreateUserRequest) string {
	jsonData, err := json.Marshal(reqBody)
	ts.Require().NoError(err, "Failed to marshal user request")

	req, err := http.NewRequest("POST", testServerURL+"/users", bytes.NewBuffer(jsonData))
	ts.Require().NoError(err, "Failed to create user request")
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.client.Do(req)
	ts.Require().NoError(err, "Failed to send user request")
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	ts.Require().NoError(err, "Failed to read response body")

	if resp.StatusCode != http.StatusCreated {
		ts.T().Logf("User creation failed: status %d, body: %s", resp.StatusCode, string(body))
	}
	ts.Require().Equal(http.StatusCreated, resp.StatusCode, "User creation should succeed")

	var createdUser testutils.User
	ts.Require().NoError(json.Unmarshal(body, &createdUser))
	return createdUser.ID
}

func (ts *UserSchemaRequiredAPITestSuite) createUserAndExpectError(reqBody CreateUserRequest, expectedErrorCode string) {
	jsonData, err := json.Marshal(reqBody)
	ts.Require().NoError(err, "Failed to marshal user request")

	req, err := http.NewRequest("POST", testServerURL+"/users", bytes.NewBuffer(jsonData))
	ts.Require().NoError(err, "Failed to create user request")
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.client.Do(req)
	ts.Require().NoError(err, "Failed to send user request")
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	ts.Require().NoError(err, "Failed to read response body")
	ts.Require().Equal(http.StatusBadRequest, resp.StatusCode, "User creation should fail with validation error")

	var errorResp ErrorResponse
	ts.Require().NoError(json.Unmarshal(body, &errorResp))
	ts.Require().Equal(expectedErrorCode, errorResp.Code, "Error code should match expected")
}

func (ts *UserSchemaRequiredAPITestSuite) deleteUser(userID string) {
	req, err := http.NewRequest("DELETE", testServerURL+"/users/"+userID, nil)
	if err != nil {
		return
	}
	resp, err := ts.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}

func (ts *UserSchemaRequiredAPITestSuite) deleteSchema(schemaID string) {
	req, err := http.NewRequest("DELETE", testServerURL+"/user-schemas/"+schemaID, nil)
	if err != nil {
		return
	}
	resp, err := ts.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}
