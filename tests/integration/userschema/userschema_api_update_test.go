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

	"github.com/stretchr/testify/suite"
)

type UpdateUserSchemaTestSuite struct {
	suite.Suite
	client          *http.Client
	testSchemaID    string
	anotherSchemaID string
}

func TestUpdateUserSchemaTestSuite(t *testing.T) {
	suite.Run(t, new(UpdateUserSchemaTestSuite))
}

func (ts *UpdateUserSchemaTestSuite) SetupSuite() {
	ts.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// Create test schemas for update tests
	schema1 := CreateUserSchemaRequest{
		Name: "update-test-schema-1",
		Schema: json.RawMessage(`{
			"originalField": {"type": "string"}
		}`),
	}

	schema2 := CreateUserSchemaRequest{
		Name: "update-test-schema-2",
		Schema: json.RawMessage(`{
			"anotherField": {"type": "string"}
		}`),
	}

	ts.testSchemaID = ts.createTestSchema(schema1)
	ts.anotherSchemaID = ts.createTestSchema(schema2)
}

func (ts *UpdateUserSchemaTestSuite) TearDownSuite() {
	// Clean up test schemas
	if ts.testSchemaID != "" {
		ts.deleteTestSchema(ts.testSchemaID)
	}
	if ts.anotherSchemaID != "" {
		ts.deleteTestSchema(ts.anotherSchemaID)
	}
}

// TestUpdateUserSchema tests PUT /user-schemas/{id} with valid data
func (ts *UpdateUserSchemaTestSuite) TestUpdateUserSchema() {
	updateRequest := UpdateUserSchemaRequest{
		Name: "updated-schema-name",
		Schema: json.RawMessage(`{
			"updatedField": {"type": "string"},
			"newField": {"type": "number"},
			"complexField": {
				"type": "object",
				"properties": {
					"nestedField": {"type": "boolean"}
				}
			}
		}`),
	}

	jsonData, err := json.Marshal(updateRequest)
	if err != nil {
		ts.T().Fatalf("Failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("PUT", testServerURL+"/user-schemas/"+ts.testSchemaID, bytes.NewBuffer(jsonData))
	if err != nil {
		ts.T().Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.client.Do(req)
	if err != nil {
		ts.T().Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	ts.Assert().Equal(http.StatusOK, resp.StatusCode, "Should return 200 OK")

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		ts.T().Fatalf("Failed to read response body: %v", err)
	}

	var updatedSchema UserSchema
	err = json.Unmarshal(bodyBytes, &updatedSchema)
	if err != nil {
		ts.T().Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify updated schema according to API spec
	ts.Assert().Equal(ts.testSchemaID, updatedSchema.ID, "ID should remain the same")
	ts.Assert().Equal(updateRequest.Name, updatedSchema.Name, "Name should be updated")
	ts.Assert().JSONEq(string(updateRequest.Schema), string(updatedSchema.Schema), "Schema data should be updated")
}

// TestUpdateUserSchemaNotFound tests PUT /user-schemas/{id} with non-existent ID
func (ts *UpdateUserSchemaTestSuite) TestUpdateUserSchemaNotFound() {
	nonExistentID := "550e8400-e29b-41d4-a716-446655440000"

	updateRequest := UpdateUserSchemaRequest{
		Name:   "updated-name",
		Schema: json.RawMessage(`{"field": {"type": "string"}}`),
	}

	jsonData, err := json.Marshal(updateRequest)
	if err != nil {
		ts.T().Fatalf("Failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("PUT", testServerURL+"/user-schemas/"+nonExistentID, bytes.NewBuffer(jsonData))
	if err != nil {
		ts.T().Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.client.Do(req)
	if err != nil {
		ts.T().Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	ts.Assert().Equal(http.StatusNotFound, resp.StatusCode, "Should return 404 Not Found")

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		ts.T().Fatalf("Failed to read response body: %v", err)
	}

	var errorResp ErrorResponse
	err = json.Unmarshal(bodyBytes, &errorResp)
	if err != nil {
		ts.T().Fatalf("Failed to unmarshal error response: %v", err)
	}

	ts.Assert().NotEmpty(errorResp.Code, "Error should have code")
	ts.Assert().NotEmpty(errorResp.Message, "Error should have message")
}

// TestUpdateUserSchemaWithNameConflict tests PUT /user-schemas/{id} with conflicting name
func (ts *UpdateUserSchemaTestSuite) TestUpdateUserSchemaWithNameConflict() {
	// Try to update first schema with the name of the second schema
	updateRequest := UpdateUserSchemaRequest{
		Name:   "update-test-schema-2", // Name of another existing schema
		Schema: json.RawMessage(`{"conflictField": {"type": "string"}}`),
	}

	jsonData, err := json.Marshal(updateRequest)
	if err != nil {
		ts.T().Fatalf("Failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("PUT", testServerURL+"/user-schemas/"+ts.testSchemaID, bytes.NewBuffer(jsonData))
	if err != nil {
		ts.T().Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.client.Do(req)
	if err != nil {
		ts.T().Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	ts.Assert().Equal(http.StatusConflict, resp.StatusCode, "Should return 409 Conflict for name conflict")

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		ts.T().Fatalf("Failed to read response body: %v", err)
	}

	var errorResp ErrorResponse
	err = json.Unmarshal(bodyBytes, &errorResp)
	if err != nil {
		ts.T().Fatalf("Failed to unmarshal error response: %v", err)
	}

	ts.Assert().NotEmpty(errorResp.Code, "Error should have code")
	ts.Assert().NotEmpty(errorResp.Message, "Error should have message")
}

// TestUpdateUserSchemaWithInvalidData tests PUT /user-schemas/{id} with invalid request data
func (ts *UpdateUserSchemaTestSuite) TestUpdateUserSchemaWithInvalidData() {
	testCases := []struct {
		name        string
		requestBody string
	}{
		{
			name:        "empty name",
			requestBody: `{"name": "", "schema": {"field": {"type": "string"}}}`,
		},
		{
			name:        "missing name",
			requestBody: `{"schema": {"field": {"type": "string"}}}`,
		},
		{
			name:        "empty schema",
			requestBody: `{"name": "updated-name", "schema": {}}`,
		},
		{
			name:        "missing schema",
			requestBody: `{"name": "updated-name"}`,
		},
		{
			name:        "invalid JSON",
			requestBody: `{"name": "updated-name", "schema": invalid}`,
		},
		{
			name:        "malformed JSON",
			requestBody: `{"name": "updated-name"`,
		},
	}

	for _, tc := range testCases {
		ts.T().Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("PUT", testServerURL+"/user-schemas/"+ts.testSchemaID, bytes.NewBufferString(tc.requestBody))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := ts.client.Do(req)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()

			ts.Assert().Equal(http.StatusBadRequest, resp.StatusCode, "Should return 400 Bad Request for: %s", tc.name)

			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			var errorResp ErrorResponse
			err = json.Unmarshal(bodyBytes, &errorResp)
			if err != nil {
				t.Fatalf("Failed to unmarshal error response: %v", err)
			}

			ts.Assert().NotEmpty(errorResp.Code, "Error should have code")
			ts.Assert().NotEmpty(errorResp.Message, "Error should have message")
		})
	}
}

// TestUpdateUserSchemaWithComplexData tests PUT /user-schemas/{id} with complex schema
func (ts *UpdateUserSchemaTestSuite) TestUpdateUserSchemaWithComplexData() {
	updateRequest := UpdateUserSchemaRequest{
		Name: "complex-updated-schema",
		Schema: json.RawMessage(`{
			"user": {
				"type": "object",
				"properties": {
					"profile": {
						"type": "object",
						"properties": {
							"personalInfo": {
								"type": "object",
								"properties": {
									"firstName": {"type": "string"},
									"lastName": {"type": "string"},
									"dateOfBirth": {"type": "string", "regex": "^\\d{4}-\\d{2}-\\d{2}$"}
								}
							},
							"contactInfo": {
								"type": "object",
								"properties": {
									"email": {
										"type": "string",
										"regex": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
									},
									"phone": {"type": "string"}
								}
							}
						}
					},
					"preferences": {
						"type": "array",
						"items": {
							"type": "object",
							"properties": {
								"category": {"type": "string"},
								"enabled": {"type": "boolean"}
							}
						}
					}
				}
			}
		}`),
	}

	jsonData, err := json.Marshal(updateRequest)
	if err != nil {
		ts.T().Fatalf("Failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("PUT", testServerURL+"/user-schemas/"+ts.testSchemaID, bytes.NewBuffer(jsonData))
	if err != nil {
		ts.T().Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.client.Do(req)
	if err != nil {
		ts.T().Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	ts.Assert().Equal(http.StatusOK, resp.StatusCode, "Should return 200 OK")

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		ts.T().Fatalf("Failed to read response body: %v", err)
	}

	var updatedSchema UserSchema
	err = json.Unmarshal(bodyBytes, &updatedSchema)
	if err != nil {
		ts.T().Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify complex schema was updated correctly
	ts.Assert().Equal(ts.testSchemaID, updatedSchema.ID, "ID should remain the same")
	ts.Assert().Equal(updateRequest.Name, updatedSchema.Name, "Name should be updated")
	ts.Assert().JSONEq(string(updateRequest.Schema), string(updatedSchema.Schema), "Complex schema data should be updated")
}

// Helper function to create a test schema
func (ts *UpdateUserSchemaTestSuite) createTestSchema(schema CreateUserSchemaRequest) string {
	jsonData, err := json.Marshal(schema)
	if err != nil {
		ts.T().Fatalf("Failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", testServerURL+"/user-schemas", bytes.NewBuffer(jsonData))
	if err != nil {
		ts.T().Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.client.Do(req)
	if err != nil {
		ts.T().Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		ts.T().Fatalf("Failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		ts.T().Fatalf("Expected status 201, got %d. Response: %s", resp.StatusCode, string(bodyBytes))
	}

	var createdSchema UserSchema
	err = json.Unmarshal(bodyBytes, &createdSchema)
	if err != nil {
		ts.T().Fatalf("Failed to unmarshal response: %v", err)
	}

	return createdSchema.ID
}

// Helper function to delete a test schema
func (ts *UpdateUserSchemaTestSuite) deleteTestSchema(schemaID string) {
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
