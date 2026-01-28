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
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	dbmodel "github.com/asgardeo/thunder/internal/system/database/model"
	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/database/transaction"
)

const testDeploymentID = "test-deployment-id"

// MockDBClient is a mock implementation of provider.DBClientInterface
type MockDBClient struct {
	mock.Mock
}

func (m *MockDBClient) Query(q dbmodel.DBQuery, args ...interface{}) ([]map[string]interface{}, error) {
	return nil, nil
}

func (m *MockDBClient) QueryContext(ctx context.Context, q dbmodel.DBQuery,
	args ...interface{}) ([]map[string]interface{}, error) {
	return nil, nil
}

func (m *MockDBClient) Execute(q dbmodel.DBQuery, args ...interface{}) (int64, error) {
	return 0, nil
}

func (m *MockDBClient) ExecuteContext(ctx context.Context, q dbmodel.DBQuery, args ...interface{}) (int64, error) {
	callArgs := make([]interface{}, 0, 2+len(args))
	callArgs = append(callArgs, ctx, q)
	callArgs = append(callArgs, args...)
	ret := m.Called(callArgs...)

	// Handle return values. If first return is int, use it. Usually rows affected.
	var rows int64
	if r0 := ret.Get(0); r0 != nil {
		if v, ok := r0.(int64); ok {
			rows = v
		} else if v, ok := r0.(int); ok {
			rows = int64(v)
		}
	}
	return rows, ret.Error(1)
}

func (m *MockDBClient) BeginTx() (dbmodel.TxInterface, error) {
	return nil, nil
}

func (m *MockDBClient) GetTransactioner() (transaction.Transactioner, error) {
	return nil, nil
}

var _ provider.DBClientInterface = (*MockDBClient)(nil)

// UserStoreTestSuite is the test suite for userStore.
type UserStoreTestSuite struct {
	suite.Suite
	mockDB *MockDBClient
	store  *userStore
}

// TestUserStoreTestSuite runs the test suite.
func TestUserStoreTestSuite(t *testing.T) {
	suite.Run(t, new(UserStoreTestSuite))
}

// SetupTest sets up the test suite.
func (suite *UserStoreTestSuite) SetupTest() {
	suite.mockDB = new(MockDBClient)
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

// Test syncIndexedAttributes

func (suite *UserStoreTestSuite) TestSyncIndexedAttributes_EmptyAttributes() {
	err := suite.store.syncIndexedAttributes(context.Background(), suite.mockDB, "user1", nil)
	suite.NoError(err)
}

func (suite *UserStoreTestSuite) TestSyncIndexedAttributes_Success_StringValues() {
	attributes := json.RawMessage(`{
		"username": "john.doe",
		"email": "john@example.com",
		"mobileNumber": "1234567890",
		"sub": "user-sub-id"
	}`)

	// Expect batch insert with all indexed attributes
	suite.mockDB.On("ExecuteContext", mock.Anything, mock.MatchedBy(func(query dbmodel.DBQuery) bool {
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

	err := suite.store.syncIndexedAttributes(context.Background(), suite.mockDB, "user1", attributes)

	suite.NoError(err)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *UserStoreTestSuite) TestSyncIndexedAttributes_Success_MixedTypes() {
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
	suite.mockDB.On("ExecuteContext", mock.Anything, mock.MatchedBy(func(query dbmodel.DBQuery) bool {
		return strings.Contains(query.Query, "INSERT INTO USER_INDEXED_ATTRIBUTES") &&
			strings.Contains(query.Query, "USER_ID") &&
			strings.Contains(query.Query, "ATTRIBUTE_NAME") &&
			strings.Contains(query.Query, "ATTRIBUTE_VALUE") &&
			strings.Contains(query.Query, "DEPLOYMENT_ID")
	}), mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)

	err := suite.store.syncIndexedAttributes(context.Background(), suite.mockDB, "user1", attributes)

	suite.NoError(err)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *UserStoreTestSuite) TestSyncIndexedAttributes_UnmarshalError() {
	invalidJSON := json.RawMessage(`{invalid json}`)

	err := suite.store.syncIndexedAttributes(context.Background(), suite.mockDB, "user1", invalidJSON)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to unmarshal user attributes")
}

func (suite *UserStoreTestSuite) TestSyncIndexedAttributes_ExecError() {
	attributes := json.RawMessage(`{"username": "john.doe"}`)
	execError := errors.New("insert failed")

	suite.mockDB.On("ExecuteContext", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything).
		Return(nil, execError)

	err := suite.store.syncIndexedAttributes(context.Background(), suite.mockDB, "user1", attributes)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to batch insert indexed attributes")
}

func (suite *UserStoreTestSuite) TestSyncIndexedAttributes_ComplexTypesSkipped() {
	attributes := json.RawMessage(`{
		"username": "john.doe",
		"metadata": {"key": "value"},
		"tags": ["tag1", "tag2"]
	}`)

	// Only username should be inserted (metadata and tags are complex types)
	suite.mockDB.On("ExecuteContext", mock.Anything, mock.Anything, "user1", "username", "john.doe", testDeploymentID).
		Return(nil, nil)

	err := suite.store.syncIndexedAttributes(context.Background(), suite.mockDB, "user1", attributes)

	suite.NoError(err)
}

func (suite *UserStoreTestSuite) TestSyncIndexedAttributes_NoIndexedAttributes() {
	attributes := json.RawMessage(`{"nonIndexed": "value", "another": "test"}`)

	// No Exec should be called because no attributes are indexed
	err := suite.store.syncIndexedAttributes(context.Background(), suite.mockDB, "user1", attributes)

	suite.NoError(err)
	// Verify that Exec was never called
	suite.mockDB.AssertNotCalled(suite.T(), "ExecuteContext")
}

func (suite *UserStoreTestSuite) TestSyncIndexedAttributes_IntegerValues() {
	attributes := json.RawMessage(`{"username": 12345}`)

	// Integer should be converted to string
	suite.mockDB.On("ExecuteContext", mock.Anything, mock.Anything, "user1", "username",
		mock.MatchedBy(func(val string) bool {
			return val == "12345"
		}), testDeploymentID).Return(nil, nil)

	err := suite.store.syncIndexedAttributes(context.Background(), suite.mockDB, "user1", attributes)

	suite.NoError(err)
}

func (suite *UserStoreTestSuite) TestSyncIndexedAttributes_BooleanValues() {
	attributes := json.RawMessage(`{"email": true}`)

	// Boolean should be converted to string
	suite.mockDB.On("ExecuteContext", mock.Anything, mock.Anything, "user1", "email",
		mock.MatchedBy(func(val string) bool {
			return val == "true"
		}), testDeploymentID).Return(nil, nil)

	err := suite.store.syncIndexedAttributes(context.Background(), suite.mockDB, "user1", attributes)

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
