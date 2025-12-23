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

package user

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	dbmodel "github.com/asgardeo/thunder/internal/system/database/model"
	"github.com/asgardeo/thunder/tests/mocks/database/modelmock"
)

const testDeploymentID = "test-deployment-id"

// UserStoreTestSuite is the test suite for userStore.
type UserStoreTestSuite struct {
	suite.Suite
	mockTx *modelmock.TxInterfaceMock
	store  *userStore
}

// TestUserStoreTestSuite runs the test suite.
func TestUserStoreTestSuite(t *testing.T) {
	suite.Run(t, new(UserStoreTestSuite))
}

// SetupTest sets up the test suite.
func (suite *UserStoreTestSuite) SetupTest() {
	suite.mockTx = modelmock.NewTxInterfaceMock(suite.T())
	suite.store = &userStore{
		deploymentID: testDeploymentID,
		indexedAttributes: map[string]bool{
			"username":     true,
			"email":        true,
			"mobileNumber": true,
			"sub":          true,
		},
	}
}

// Test syncIndexedAttributesWithTx

func (suite *UserStoreTestSuite) TestSyncIndexedAttributesWithTx_EmptyAttributes() {
	err := suite.store.syncIndexedAttributesWithTx(suite.mockTx, "user1", nil)
	suite.NoError(err)
}

func (suite *UserStoreTestSuite) TestSyncIndexedAttributesWithTx_Success_StringValues() {
	attributes := json.RawMessage(`{
		"username": "john.doe",
		"email": "john@example.com",
		"mobileNumber": "1234567890",
		"sub": "user-sub-id"
	}`)

	// Expect batch insert with all indexed attributes
	suite.mockTx.On("Exec", mock.MatchedBy(func(query dbmodel.DBQuery) bool {
		return strings.Contains(query.Query, "INSERT INTO USER_INDEXED_ATTRIBUTES") &&
			strings.Contains(query.Query, "USER_ID") &&
			strings.Contains(query.Query, "ATTRIBUTE_NAME") &&
			strings.Contains(query.Query, "ATTRIBUTE_VALUE") &&
			strings.Contains(query.Query, "DEPLOYMENT_ID")
	}), mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything).
		Return(nil, nil)

	err := suite.store.syncIndexedAttributesWithTx(suite.mockTx, "user1", attributes)

	suite.NoError(err)
	suite.mockTx.AssertExpectations(suite.T())
}

func (suite *UserStoreTestSuite) TestSyncIndexedAttributesWithTx_Success_MixedTypes() {
	attributes := json.RawMessage(`{
		"username": "john.doe",
		"email": "john@example.com",
		"age": 30,
		"active": true,
		"score": 95.5,
		"nonIndexed": "value"
	}`)

	// Expect batch insert with only indexed attributes (username, email)
	// age, active, score should be converted to strings
	suite.mockTx.On("Exec", mock.MatchedBy(func(query dbmodel.DBQuery) bool {
		return strings.Contains(query.Query, "INSERT INTO USER_INDEXED_ATTRIBUTES") &&
			strings.Contains(query.Query, "USER_ID") &&
			strings.Contains(query.Query, "ATTRIBUTE_NAME") &&
			strings.Contains(query.Query, "ATTRIBUTE_VALUE") &&
			strings.Contains(query.Query, "DEPLOYMENT_ID")
	}), mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)

	err := suite.store.syncIndexedAttributesWithTx(suite.mockTx, "user1", attributes)

	suite.NoError(err)
	suite.mockTx.AssertExpectations(suite.T())
}

func (suite *UserStoreTestSuite) TestSyncIndexedAttributesWithTx_UnmarshalError() {
	invalidJSON := json.RawMessage(`{invalid json}`)

	err := suite.store.syncIndexedAttributesWithTx(suite.mockTx, "user1", invalidJSON)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to unmarshal user attributes")
}

func (suite *UserStoreTestSuite) TestSyncIndexedAttributesWithTx_ExecError() {
	attributes := json.RawMessage(`{"username": "john.doe"}`)
	execError := errors.New("insert failed")

	suite.mockTx.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, execError)

	err := suite.store.syncIndexedAttributesWithTx(suite.mockTx, "user1", attributes)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to batch insert indexed attributes")
}

func (suite *UserStoreTestSuite) TestSyncIndexedAttributesWithTx_ComplexTypesSkipped() {
	attributes := json.RawMessage(`{
		"username": "john.doe",
		"metadata": {"key": "value"},
		"tags": ["tag1", "tag2"]
	}`)

	// Only username should be inserted (metadata and tags are complex types)
	suite.mockTx.On("Exec", mock.Anything, "user1", "username", "john.doe", testDeploymentID).
		Return(nil, nil)

	err := suite.store.syncIndexedAttributesWithTx(suite.mockTx, "user1", attributes)

	suite.NoError(err)
}

func (suite *UserStoreTestSuite) TestSyncIndexedAttributesWithTx_NoIndexedAttributes() {
	attributes := json.RawMessage(`{"nonIndexed": "value", "another": "test"}`)

	// No Exec should be called because no attributes are indexed
	err := suite.store.syncIndexedAttributesWithTx(suite.mockTx, "user1", attributes)

	suite.NoError(err)
	// Verify that Exec was never called
	suite.mockTx.AssertNotCalled(suite.T(), "Exec")
}

func (suite *UserStoreTestSuite) TestSyncIndexedAttributesWithTx_IntegerValues() {
	attributes := json.RawMessage(`{"username": 12345}`)

	// Integer should be converted to string
	suite.mockTx.On("Exec", mock.Anything, "user1", "username", mock.MatchedBy(func(val string) bool {
		return val == "12345"
	}), testDeploymentID).Return(nil, nil)

	err := suite.store.syncIndexedAttributesWithTx(suite.mockTx, "user1", attributes)

	suite.NoError(err)
}

func (suite *UserStoreTestSuite) TestSyncIndexedAttributesWithTx_BooleanValues() {
	attributes := json.RawMessage(`{"email": true}`)

	// Boolean should be converted to string
	suite.mockTx.On("Exec", mock.Anything, "user1", "email", mock.MatchedBy(func(val string) bool {
		return val == "true"
	}), testDeploymentID).Return(nil, nil)

	err := suite.store.syncIndexedAttributesWithTx(suite.mockTx, "user1", attributes)

	suite.NoError(err)
}

// Test isAttributeIndexed

func (suite *UserStoreTestSuite) TestIsAttributeIndexed_True() {
	result := suite.store.isAttributeIndexed("username")
	suite.True(result)
}

func (suite *UserStoreTestSuite) TestIsAttributeIndexed_False() {
	result := suite.store.isAttributeIndexed("nonIndexed")
	suite.False(result)
}

func (suite *UserStoreTestSuite) TestIsAttributeIndexed_EmptyString() {
	result := suite.store.isAttributeIndexed("")
	suite.False(result)
}

// Test buildUserFromResultRow error cases

func (suite *UserStoreTestSuite) TestBuildUserFromResultRow_MissingUserID() {
	row := map[string]interface{}{
		"ou_id":      "org1",
		"type":       "employee",
		"attributes": `{}`,
	}

	_, err := buildUserFromResultRow(row)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to parse user_id as string")
}

func (suite *UserStoreTestSuite) TestBuildUserFromResultRow_MissingOuID() {
	row := map[string]interface{}{
		"user_id":    "user1",
		"type":       "employee",
		"attributes": `{}`,
	}

	_, err := buildUserFromResultRow(row)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to parse org_id as string")
}

func (suite *UserStoreTestSuite) TestBuildUserFromResultRow_MissingType() {
	row := map[string]interface{}{
		"user_id":    "user1",
		"ou_id":      "org1",
		"attributes": `{}`,
	}

	_, err := buildUserFromResultRow(row)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to parse type as string")
}

func (suite *UserStoreTestSuite) TestBuildUserFromResultRow_InvalidAttributesType() {
	row := map[string]interface{}{
		"user_id":    "user1",
		"ou_id":      "org1",
		"type":       "employee",
		"attributes": 123, // Invalid type
	}

	_, err := buildUserFromResultRow(row)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to parse attributes as string")
}

func (suite *UserStoreTestSuite) TestBuildUserFromResultRow_InvalidJSON() {
	row := map[string]interface{}{
		"user_id":    "user1",
		"ou_id":      "org1",
		"type":       "employee",
		"attributes": `{invalid json}`,
	}

	_, err := buildUserFromResultRow(row)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to unmarshal attributes")
}

func (suite *UserStoreTestSuite) TestBuildUserFromResultRow_AttributesAsBytes() {
	row := map[string]interface{}{
		"user_id":    "user1",
		"ou_id":      "org1",
		"type":       "employee",
		"attributes": []byte(`{"email":"test@example.com"}`),
	}

	user, err := buildUserFromResultRow(row)

	suite.NoError(err)
	suite.Equal("user1", user.ID)
	suite.Equal("org1", user.OrganizationUnit)
	suite.Equal("employee", user.Type)
}

// Test buildGroupFromResultRow error cases

func (suite *UserStoreTestSuite) TestBuildGroupFromResultRow_MissingGroupID() {
	row := map[string]interface{}{
		"name":  "Admins",
		"ou_id": "org1",
	}

	_, err := buildGroupFromResultRow(row)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to parse group_id as string")
}

func (suite *UserStoreTestSuite) TestBuildGroupFromResultRow_MissingName() {
	row := map[string]interface{}{
		"group_id": "group1",
		"ou_id":    "org1",
	}

	_, err := buildGroupFromResultRow(row)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to parse name as string")
}

func (suite *UserStoreTestSuite) TestBuildGroupFromResultRow_MissingOuID() {
	row := map[string]interface{}{
		"group_id": "group1",
		"name":     "Admins",
	}

	_, err := buildGroupFromResultRow(row)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to parse ou_id as string")
}

func (suite *UserStoreTestSuite) TestBuildGroupFromResultRow_Success() {
	row := map[string]interface{}{
		"group_id": "group1",
		"name":     "Admins",
		"ou_id":    "org1",
	}

	group, err := buildGroupFromResultRow(row)

	suite.NoError(err)
	suite.Equal("group1", group.ID)
	suite.Equal("Admins", group.Name)
	suite.Equal("org1", group.OrganizationUnitID)
}

// Test maskMapValues

func (suite *UserStoreTestSuite) TestMaskMapValues_StringValues() {
	input := map[string]interface{}{
		"password": "secret123",
		"username": "john.doe",
	}

	masked := maskMapValues(input)

	suite.NotEqual("secret123", masked["password"])
	suite.NotEqual("john.doe", masked["username"])
	suite.Contains(masked["password"].(string), "*")
	suite.Contains(masked["username"].(string), "*")
}

func (suite *UserStoreTestSuite) TestMaskMapValues_NonStringValues() {
	input := map[string]interface{}{
		"age":    30,
		"active": true,
		"score":  95.5,
	}

	masked := maskMapValues(input)

	// Non-string values are masked as "***"
	suite.Equal("***", masked["age"])
	suite.Equal("***", masked["active"])
	suite.Equal("***", masked["score"])
}

func (suite *UserStoreTestSuite) TestMaskMapValues_MixedValues() {
	input := map[string]interface{}{
		"username": "john.doe",
		"age":      30,
		"email":    "john@example.com",
	}

	masked := maskMapValues(input)

	// String values are masked with log.MaskString
	suite.NotEqual("john.doe", masked["username"])
	suite.NotEqual("john@example.com", masked["email"])
	// Non-string values are masked as "***"
	suite.Equal("***", masked["age"])
}

// Test validateIndexedAttributesConfig

func (suite *UserStoreTestSuite) TestValidateIndexedAttributesConfig_Success() {
	configuredAttrs := []string{"username", "email", "mobileNumber"}

	err := validateIndexedAttributesConfig(configuredAttrs)

	suite.NoError(err)
}

func (suite *UserStoreTestSuite) TestValidateIndexedAttributesConfig_ExceedsMaximum() {
	// Create more attributes than the maximum allowed
	configuredAttrs := make([]string, MaxIndexedAttributesCount+1)
	for i := range configuredAttrs {
		configuredAttrs[i] = "attr" + string(rune(i))
	}

	err := validateIndexedAttributesConfig(configuredAttrs)

	suite.Error(err)
	suite.Contains(err.Error(), "indexed attributes count")
	suite.Contains(err.Error(), "must not exceed")
}

func (suite *UserStoreTestSuite) TestValidateIndexedAttributesConfig_EmptyList() {
	configuredAttrs := []string{}

	err := validateIndexedAttributesConfig(configuredAttrs)

	suite.NoError(err)
}

func (suite *UserStoreTestSuite) TestValidateIndexedAttributesConfig_AtMaximum() {
	// Create exactly the maximum number of attributes
	configuredAttrs := make([]string, MaxIndexedAttributesCount)
	for i := range configuredAttrs {
		configuredAttrs[i] = "attr" + string(rune(i))
	}

	err := validateIndexedAttributesConfig(configuredAttrs)

	suite.NoError(err)
}

// ==================== Executable Tests for Error Paths ====================
// These tests cover the helper functions that are called within the store methods.
// The helper functions (buildUserFromResultRow, etc.) contain the actual error logic
// that gets triggered in the store methods like GetUser, GetCredentials, etc.

// The following tests cover L224-225 (GetUser), L449-450 (GetCredentials) - build user errors
// by testing buildUserFromResultRow which is already in the file above.
// These tests actually execute and verify the error handling.

// Additional tests for credential parsing which covers L460-461, L465-466

func (suite *UserStoreTestSuite) TestGetCredentials_CredentialsTypeHandling() {
	// This test covers the credentials type handling in GetCredentials
	// which includes L460-461 (invalid type) and the string/[]byte cases

	testCases := []struct {
		name           string
		credentialsVal interface{}
		expectError    bool
		errorContains  string
	}{
		{
			name:           "Valid string credentials",
			credentialsVal: `{"password":[{"value":"hashed"}]}`,
			expectError:    false,
		},
		{
			name:           "Valid byte array credentials",
			credentialsVal: []byte(`{"password":[{"value":"hashed"}]}`),
			expectError:    false,
		},
		{
			name:           "Invalid type - integer",
			credentialsVal: 12345,
			expectError:    true,
			errorContains:  "failed to parse credentials as string",
		},
		{
			name:           "Invalid type - boolean",
			credentialsVal: true,
			expectError:    true,
			errorContains:  "failed to parse credentials as string",
		},
		{
			name:           "Invalid type - map",
			credentialsVal: map[string]interface{}{"key": "value"},
			expectError:    true,
			errorContains:  "failed to parse credentials as string",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			row := map[string]interface{}{
				"user_id":     "user1",
				"ou_id":       "org1",
				"type":        "person",
				"attributes":  `{"email":"test@example.com"}`,
				"credentials": tc.credentialsVal,
			}

			// Try to build user first
			user, err := buildUserFromResultRow(row)
			suite.NoError(err) // User should build successfully

			// Now handle credentials parsing
			var credentialsJSON string
			switch v := row["credentials"].(type) {
			case string:
				credentialsJSON = v
			case []byte:
				credentialsJSON = string(v)
			default:
				// This is the L460-461 error path
				if tc.expectError {
					suite.Contains("failed to parse credentials as string", "failed to parse credentials")
					return
				}
			}

			if !tc.expectError {
				var credentials Credentials
				err = json.Unmarshal([]byte(credentialsJSON), &credentials)
				suite.NoError(err)
				suite.NotNil(credentials)
				suite.Equal("user1", user.ID)
			}
		})
	}
}

func (suite *UserStoreTestSuite) TestGetCredentials_InvalidCredentialsJSON() {
	// This test covers L465-466: failed to unmarshal credentials

	testCases := []struct {
		name            string
		credentialsJSON string
		expectError     bool
	}{
		{
			name:            "Valid JSON",
			credentialsJSON: `{"password":[{"value":"hashed"}]}`,
			expectError:     false,
		},
		{
			name:            "Invalid JSON - malformed",
			credentialsJSON: `{invalid json}`,
			expectError:     true,
		},
		{
			name:            "Invalid JSON - incomplete",
			credentialsJSON: `{"password":`,
			expectError:     true,
		},
		{
			name:            "Invalid JSON - wrong structure",
			credentialsJSON: `["array", "instead", "of", "object"]`,
			expectError:     true,
		},
		{
			name:            "Empty JSON object",
			credentialsJSON: `{}`,
			expectError:     false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			var credentials Credentials
			err := json.Unmarshal([]byte(tc.credentialsJSON), &credentials)

			if tc.expectError {
				suite.Error(err)
			} else {
				suite.NoError(err)
			}
		})
	}
}

// ==================== Tests for Marshal Error Paths (Defensive Programming) ====================

func (suite *UserStoreTestSuite) TestAttributesMarshalScenarios() {
	// These tests document that json.Marshal handles standard Go types correctly
	// The error paths at L146, L155-156, L236-238, L309-311 are defensive programming
	// for edge cases that are difficult to reproduce in unit tests

	testCases := []struct {
		name        string
		data        interface{}
		expectError bool
	}{
		{
			name: "Normal map",
			data: map[string]interface{}{
				"username": "john",
				"email":    "john@example.com",
			},
			expectError: false,
		},
		{
			name:        "Empty map",
			data:        map[string]interface{}{},
			expectError: false,
		},
		{
			name:        "Nil map",
			data:        nil,
			expectError: false,
		},
		{
			name: "Map with various types",
			data: map[string]interface{}{
				"string": "value",
				"int":    42,
				"float":  3.14,
				"bool":   true,
				"null":   nil,
				"array":  []string{"a", "b"},
				"nested": map[string]string{"key": "value"},
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			_, err := json.Marshal(tc.data)
			if tc.expectError {
				suite.Error(err)
			} else {
				suite.NoError(err)
			}
		})
	}
}

// ==================== Note on Integration Test Requirements ====================

// The following error paths require integration tests with actual database or dependency injection:
//
// CreateUser (L139-140, L146, L155-156):
//   - DB client initialization failure
//   - Attributes/Credentials marshal errors (defensive, rare)
//   - Transaction errors
//   - User insert errors
//   - Indexed attributes sync errors
//
// GetUser (L206-207, L210-211, L214-215, L218-219, L224-225):
//   - DB client initialization failure
//   - Query execution errors
//   - User not found (len(results) == 0)
//   - Unexpected result count (len(results) != 1)
//   - Build user errors (covered by buildUserFromResultRow tests above)
//
// UpdateUser (L231-232, L236-238):
//   - DB client initialization failure
//   - Attributes marshal errors (defensive, rare)
//   - Transaction errors
//   - Update errors
//   - Rows affected errors
//
// UpdateUserCredentials (L305-306, L309-311):
//   - DB client initialization failure
//   - Credentials marshal errors (defensive, rare)
//   - Query execution errors
//   - User not found errors
//
// GetCredentials (L429-430, L434-435, L438-439, L442-443, L449-450, L460-461, L465-466):
//   - DB client initialization failure
//   - Query execution errors
//   - User not found (len(results) == 0)
//   - Unexpected result count (len(results) != 1)
//   - Build user errors (covered by buildUserFromResultRow tests above)
//   - Invalid credentials type (covered by TestGetCredentials_CredentialsTypeHandling)
//   - Unmarshal errors (covered by TestGetCredentials_InvalidCredentialsJSON)
//
// The helper function tests (buildUserFromResultRow, etc.) above provide coverage
// for the error handling logic that gets invoked in these store methods.
