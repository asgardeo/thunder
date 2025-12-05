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
	suite.mockTx.On("Exec", mock.MatchedBy(func(query string) bool {
		return strings.Contains(query, "INSERT INTO USER_INDEXED_ATTRIBUTES") &&
			strings.Contains(query, "USER_ID") &&
			strings.Contains(query, "ATTRIBUTE_NAME") &&
			strings.Contains(query, "ATTRIBUTE_VALUE") &&
			strings.Contains(query, "DEPLOYMENT_ID")
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
	suite.mockTx.On("Exec", mock.MatchedBy(func(query string) bool {
		return strings.Contains(query, "INSERT INTO USER_INDEXED_ATTRIBUTES") &&
			strings.Contains(query, "USER_ID") &&
			strings.Contains(query, "ATTRIBUTE_NAME") &&
			strings.Contains(query, "ATTRIBUTE_VALUE") &&
			strings.Contains(query, "DEPLOYMENT_ID")
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
