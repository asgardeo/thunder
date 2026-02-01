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
	callArgs := make([]interface{}, 0, 2+len(args))
	callArgs = append(callArgs, q)
	callArgs = append(callArgs, args...)
	ret := m.Called(callArgs...)

	var result []map[string]interface{}
	if r0 := ret.Get(0); r0 != nil {
		result = r0.([]map[string]interface{})
	}
	return result, ret.Error(1)
}

func (m *MockDBClient) QueryContext(ctx context.Context, q dbmodel.DBQuery,
	args ...interface{}) ([]map[string]interface{}, error) {
	callArgs := make([]interface{}, 0, 2+len(args))
	callArgs = append(callArgs, ctx, q)
	callArgs = append(callArgs, args...)
	ret := m.Called(callArgs...)

	var result []map[string]interface{}
	if r0 := ret.Get(0); r0 != nil {
		result = r0.([]map[string]interface{})
	}
	return result, ret.Error(1)
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

// MockDBProvider is a mock implementation of provider.DBProviderInterface
type MockDBProvider struct {
	mock.Mock
	client provider.DBClientInterface
}

func (m *MockDBProvider) GetConfigDBClient() (provider.DBClientInterface, error) {
	return m.client, nil
}
func (m *MockDBProvider) GetRuntimeDBClient() (provider.DBClientInterface, error) {
	return m.client, nil
}
func (m *MockDBProvider) GetUserDBClient() (provider.DBClientInterface, error) {
	return m.client, nil
}
func (m *MockDBProvider) GetConfigDBTransactioner() (transaction.Transactioner, error) {
	return nil, nil
}
func (m *MockDBProvider) GetUserDBTransactioner() (transaction.Transactioner, error) {
	return nil, nil
}
func (m *MockDBProvider) GetRuntimeDBTransactioner() (transaction.Transactioner, error) {
	return nil, nil
}

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
	mockProvider := &MockDBProvider{client: suite.mockDB}

	suite.store = &userStore{
		deploymentID: testDeploymentID,
		indexedAttributes: map[string]bool{
			"username":     true,
			"email":        true,
			"mobileNumber": true,
			"sub":          true,
		},
		dbProvider: mockProvider,
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

func (suite *UserStoreTestSuite) TestUpdateUserCredentials() {
	userID := svcTestUserID1
	credentials := Credentials{
		"password": []Credential{
			{
				StorageType: "hash",
				Value:       "hashed-pass",
			},
		},
	}
	credentialsJSON, _ := json.Marshal(credentials)

	suite.mockDB.On("ExecuteContext", mock.Anything, QueryUpdateUserCredentialsByUserID,
		userID, string(credentialsJSON), testDeploymentID).
		Return(int64(1), nil)

	err := suite.store.UpdateUserCredentials(context.Background(), userID, credentials)
	suite.NoError(err)
}

func (suite *UserStoreTestSuite) TestCreateUser() {
	attributesMap := map[string]interface{}{"username": "john.doe"}
	attributesBytes, _ := json.Marshal(attributesMap)

	user := User{
		ID:               svcTestUserID1,
		OrganizationUnit: "ou-1",
		Type:             "customer",
		Attributes:       json.RawMessage(attributesBytes),
	}
	credentials := Credentials{}

	// Expect insert query
	suite.mockDB.On("ExecuteContext", mock.Anything, QueryCreateUser,
		user.ID, user.OrganizationUnit, user.Type, string(attributesBytes), "{}", testDeploymentID).
		Return(int64(1), nil)

	// Expect indexed attributes sync
	suite.mockDB.On("ExecuteContext", mock.Anything, mock.MatchedBy(func(query dbmodel.DBQuery) bool {
		return strings.Contains(query.Query, "INSERT INTO USER_INDEXED_ATTRIBUTES")
	}), mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)

	err := suite.store.CreateUser(context.Background(), user, credentials)
	suite.NoError(err)
}

func (suite *UserStoreTestSuite) TestGetUser() {
	userID := svcTestUserID1
	attributesMap := map[string]interface{}{"username": "john.doe"}
	attributesBytes, _ := json.Marshal(attributesMap)

	expectedUser := User{
		ID:               userID,
		OrganizationUnit: "ou-1",
		Type:             "customer",
		Attributes:       json.RawMessage(attributesBytes),
	}

	row := map[string]interface{}{
		"user_id":    expectedUser.ID,
		"ou_id":      expectedUser.OrganizationUnit,
		"type":       expectedUser.Type,
		"attributes": string(attributesBytes),
	}

	suite.mockDB.On("QueryContext", mock.Anything, QueryGetUserByUserID, userID, testDeploymentID).
		Return([]map[string]interface{}{row}, nil)

	user, err := suite.store.GetUser(context.Background(), userID)
	suite.NoError(err)
	suite.Equal(expectedUser, user)
}

func (suite *UserStoreTestSuite) TestDeleteUser() {
	userID := svcTestUserID1

	suite.mockDB.On("ExecuteContext", mock.Anything, QueryDeleteUserByUserID, userID, testDeploymentID).
		Return(int64(1), nil)

	err := suite.store.DeleteUser(context.Background(), userID)
	suite.NoError(err)
}

func (suite *UserStoreTestSuite) TestUpdateUser() {
	attributesMap := map[string]interface{}{"username": "john.doe.updated"}
	attributesBytes, _ := json.Marshal(attributesMap)

	user := &User{
		ID:               svcTestUserID1,
		OrganizationUnit: "ou-1",
		Type:             "customer",
		Attributes:       json.RawMessage(attributesBytes),
	}

	// Expect update query
	suite.mockDB.On("ExecuteContext", mock.Anything, QueryUpdateUserByUserID,
		user.ID, user.OrganizationUnit, user.Type, string(attributesBytes), testDeploymentID).
		Return(int64(1), nil)

	// Expect delete indexed attributes query
	suite.mockDB.On("ExecuteContext", mock.Anything, QueryDeleteIndexedAttributesByUser,
		user.ID, testDeploymentID).
		Return(int64(1), nil)

	// Expect indexed attributes sync
	suite.mockDB.On("ExecuteContext", mock.Anything, mock.MatchedBy(func(query dbmodel.DBQuery) bool {
		return strings.Contains(query.Query, "INSERT INTO USER_INDEXED_ATTRIBUTES")
	}), mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)

	err := suite.store.UpdateUser(context.Background(), user)
	suite.NoError(err)
}

func (suite *UserStoreTestSuite) TestGetUserList() {
	limit := 10
	offset := 0

	row := map[string]interface{}{
		"user_id":    svcTestUserID1,
		"ou_id":      "ou-1",
		"type":       "customer",
		"attributes": `{"username": "john.doe"}`,
	}

	suite.mockDB.On("QueryContext", mock.Anything, QueryGetUserList, limit, offset, testDeploymentID).
		Return([]map[string]interface{}{row}, nil)

	users, err := suite.store.GetUserList(context.Background(), limit, offset, nil)
	suite.NoError(err)
	suite.Len(users, 1)
	suite.Equal(svcTestUserID1, users[0].ID)
}

func (suite *UserStoreTestSuite) TestGetUserListCount() {
	suite.mockDB.On("QueryContext", mock.Anything, QueryGetUserCount, testDeploymentID).
		Return([]map[string]interface{}{{"total": int64(5)}}, nil)

	count, err := suite.store.GetUserListCount(context.Background(), nil)
	suite.NoError(err)
	suite.Equal(5, count)
}

func (suite *UserStoreTestSuite) TestGetCredentials() {
	userID := svcTestUserID1
	attributesMap := map[string]interface{}{"username": "john.doe"}
	attributesBytes, _ := json.Marshal(attributesMap)
	credentials := Credentials{}
	credentialsJSON, _ := json.Marshal(credentials)

	row := map[string]interface{}{
		"user_id":     userID,
		"ou_id":       "ou-1",
		"type":        "customer",
		"attributes":  string(attributesBytes),
		"credentials": string(credentialsJSON),
	}

	suite.mockDB.On("QueryContext", mock.Anything, QueryValidateUserWithCredentials, userID, testDeploymentID).
		Return([]map[string]interface{}{row}, nil)

	_, creds, err := suite.store.GetCredentials(context.Background(), userID)
	suite.NoError(err)
	suite.Equal(credentials, creds)
}

func (suite *UserStoreTestSuite) TestGetGroupCountForUser() {
	userID := svcTestUserID1
	suite.mockDB.On("QueryContext", mock.Anything, QueryGetGroupCountForUser, userID, testDeploymentID).
		Return([]map[string]interface{}{{"total": int64(3)}}, nil)

	count, err := suite.store.GetGroupCountForUser(context.Background(), userID)
	suite.NoError(err)
	suite.Equal(3, count)
}

func (suite *UserStoreTestSuite) TestGetUserGroups() {
	userID := svcTestUserID1
	limit := 10
	offset := 0

	row := map[string]interface{}{
		"group_id": "group-1",
		"name":     "admin",
		"ou_id":    "ou-1",
	}

	suite.mockDB.On("QueryContext", mock.Anything, QueryGetGroupsForUser, userID, limit, offset, testDeploymentID).
		Return([]map[string]interface{}{row}, nil)

	groups, err := suite.store.GetUserGroups(context.Background(), userID, limit, offset)
	suite.NoError(err)
	suite.Len(groups, 1)
	suite.Equal("admin", groups[0].Name)
}

func (suite *UserStoreTestSuite) TestValidateUserIDs() {
	userIDs := []string{svcTestUserID1, "user-2"}

	// Mock bulk check query
	suite.mockDB.On("QueryContext", mock.Anything, mock.MatchedBy(func(query dbmodel.DBQuery) bool {
		// Just check ID or partial query since it's built dynamically
		return strings.Contains(query.Query, "SELECT USER_ID FROM \"USER\"") || query.ID == "ASQ-USER_MGT-09"
	}), mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return([]map[string]interface{}{
			{"user_id": svcTestUserID1},
			{"user_id": "user-2"},
		}, nil)

	invalid, err := suite.store.ValidateUserIDs(context.Background(), userIDs)
	suite.NoError(err)
	suite.Empty(invalid)
}

func (suite *UserStoreTestSuite) TestIdentifyUser_NoIndexedFilters() {
	filters := map[string]interface{}{"nonIndexed": "value"}

	suite.mockDB.On("QueryContext", mock.Anything, mock.MatchedBy(func(query dbmodel.DBQuery) bool {
		return query.ID == "ASQ-USER_MGT-08"
	}), mock.Anything, mock.Anything).
		Return([]map[string]interface{}{{"user_id": svcTestUserID1}}, nil)

	id, err := suite.store.IdentifyUser(context.Background(), filters)
	suite.NoError(err)
	suite.Equal(svcTestUserID1, *id)
}

func (suite *UserStoreTestSuite) TestIdentifyUser_AllIndexed() {
	filters := map[string]interface{}{"username": "john"}

	suite.mockDB.On("QueryContext", mock.Anything, mock.MatchedBy(func(query dbmodel.DBQuery) bool {
		return strings.Contains(query.Query, "USER_INDEXED_ATTRIBUTES")
	}), mock.Anything, mock.Anything, mock.Anything).
		Return([]map[string]interface{}{{"user_id": "user-indexed"}}, nil)

	id, err := suite.store.IdentifyUser(context.Background(), filters)
	suite.NoError(err)
	suite.Equal("user-indexed", *id)
}

func (suite *UserStoreTestSuite) TestIdentifyUser_Hybrid() {
	filters := map[string]interface{}{
		"username":   "john",
		"nonIndexed": "value",
	}

	suite.mockDB.On("QueryContext", mock.Anything, mock.MatchedBy(func(query dbmodel.DBQuery) bool {
		return strings.Contains(query.Query, "JOIN USER_INDEXED_ATTRIBUTES")
	}), mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return([]map[string]interface{}{{"user_id": "user-hybrid"}}, nil)

	id, err := suite.store.IdentifyUser(context.Background(), filters)
	suite.NoError(err)
	suite.Equal("user-hybrid", *id)
}

func (suite *UserStoreTestSuite) TestIdentifyUser_NotFound() {
	filters := map[string]interface{}{"username": "ghost"}

	suite.mockDB.On("QueryContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return([]map[string]interface{}{}, nil)

	id, err := suite.store.IdentifyUser(context.Background(), filters)
	suite.ErrorIs(err, ErrUserNotFound)
	suite.Nil(id)
}

func (suite *UserStoreTestSuite) TestCreateUser_DBError() {
	suite.mockDB.On("ExecuteContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("db error"))

	err := suite.store.CreateUser(context.Background(), User{ID: "u1"}, Credentials{})
	suite.Error(err)
}

func (suite *UserStoreTestSuite) TestUpdateUser_DBError() {
	suite.mockDB.On("ExecuteContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("db error"))

	err := suite.store.UpdateUser(context.Background(), &User{ID: "u1"})
	suite.Error(err)
}
