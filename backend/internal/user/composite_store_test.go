/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// CompositeStoreTestSuite tests the composite user store functionality.
type CompositeStoreTestSuite struct {
	suite.Suite
	mockDBStore    *userStoreInterfaceMock
	mockFileStore  *userStoreInterfaceMock
	compositeStore userStoreInterface
}

// SetupTest sets up the test environment.
func (suite *CompositeStoreTestSuite) SetupTest() {
	// Create mock stores
	suite.mockDBStore = newUserStoreInterfaceMock(suite.T())
	suite.mockFileStore = newUserStoreInterfaceMock(suite.T())

	// Create composite store
	suite.compositeStore = newCompositeUserStore(suite.mockFileStore, suite.mockDBStore)
}

// TestCompositeStore_GetUserListCount tests retrieving user count from composite store.
func (suite *CompositeStoreTestSuite) TestCompositeStore_GetUserListCount() {
	ctx := context.Background()

	dbUser := User{ID: "user-1"}
	fileUser := User{ID: "user-1"}

	suite.mockDBStore.On("GetUserListCount", ctx, mock.Anything).Return(1, nil)
	suite.mockFileStore.On("GetUserListCount", ctx, mock.Anything).Return(1, nil)
	suite.mockDBStore.On("GetUserList", ctx, 1, 0, mock.Anything).Return([]User{dbUser}, nil)
	suite.mockFileStore.On("GetUserList", ctx, 1, 0, mock.Anything).Return([]User{fileUser}, nil)

	count, err := suite.compositeStore.GetUserListCount(ctx, nil)
	suite.NoError(err)
	suite.Equal(1, count) // Deduplicated by ID
}

// TestCompositeStore_GetUserListCount_DBError tests error handling when DB fails.
func (suite *CompositeStoreTestSuite) TestCompositeStore_GetUserListCount_DBError() {
	ctx := context.Background()

	suite.mockDBStore.On("GetUserListCount", ctx, mock.Anything).
		Return(0, errors.New("db error"))

	count, err := suite.compositeStore.GetUserListCount(ctx, nil)
	suite.Error(err)
	suite.Equal(0, count)
}

// TestCompositeStore_GetUserList tests retrieving and merging user lists.
func (suite *CompositeStoreTestSuite) TestCompositeStore_GetUserList() {
	ctx := context.Background()

	dbUser := User{
		ID:               "user-db-1",
		OrganizationUnit: "ou-1",
		Type:             "default",
		Attributes:       json.RawMessage(`{"name":"DB User"}`),
	}

	fileUser := User{
		ID:               "user-file-1",
		OrganizationUnit: "ou-1",
		Type:             "default",
		Attributes:       json.RawMessage(`{"name":"File User"}`),
	}

	// DB store contains database users
	suite.mockDBStore.On("GetUserListCount", ctx, mock.Anything).Return(1, nil)
	suite.mockFileStore.On("GetUserListCount", ctx, mock.Anything).Return(1, nil)
	suite.mockDBStore.On("GetUserList", ctx, 1, 0, mock.Anything).
		Return([]User{dbUser}, nil)
	suite.mockFileStore.On("GetUserList", ctx, 1, 0, mock.Anything).
		Return([]User{fileUser}, nil)

	list, err := suite.compositeStore.GetUserList(ctx, 100, 0, nil)
	suite.NoError(err)
	suite.Len(list, 2)
}

// TestCompositeStore_GetUserListCountByOUIDs tests OU-scoped count merging.
func (suite *CompositeStoreTestSuite) TestCompositeStore_GetUserListCountByOUIDs() {
	ctx := context.Background()
	ouIDs := []string{"ou-1"}

	user1 := User{ID: "user-1", OrganizationUnit: "ou-1"}
	user2 := User{ID: "user-1", OrganizationUnit: "ou-1"}

	suite.mockDBStore.On("GetUserListCountByOUIDs", ctx, ouIDs, mock.Anything).Return(1, nil)
	suite.mockFileStore.On("GetUserListCountByOUIDs", ctx, ouIDs, mock.Anything).Return(1, nil)
	suite.mockDBStore.On("GetUserListByOUIDs", ctx, ouIDs, 1, 0, mock.Anything).
		Return([]User{user1}, nil)
	suite.mockFileStore.On("GetUserListByOUIDs", ctx, ouIDs, 1, 0, mock.Anything).
		Return([]User{user2}, nil)

	count, err := suite.compositeStore.GetUserListCountByOUIDs(ctx, ouIDs, nil)
	suite.NoError(err)
	suite.Equal(1, count)
}

// TestCompositeStore_GetUserListByOUIDs tests OU-scoped list merging.
func (suite *CompositeStoreTestSuite) TestCompositeStore_GetUserListByOUIDs() {
	ctx := context.Background()
	ouIDs := []string{"ou-1"}

	user1 := User{ID: "user-db-1", OrganizationUnit: "ou-1", Type: "default"}
	user2 := User{ID: "user-file-1", OrganizationUnit: "ou-1", Type: "default"}

	suite.mockDBStore.On("GetUserListCountByOUIDs", ctx, ouIDs, mock.Anything).Return(1, nil)
	suite.mockFileStore.On("GetUserListCountByOUIDs", ctx, ouIDs, mock.Anything).Return(1, nil)
	suite.mockDBStore.On("GetUserListByOUIDs", ctx, ouIDs, 1, 0, mock.Anything).
		Return([]User{user1}, nil)
	suite.mockFileStore.On("GetUserListByOUIDs", ctx, ouIDs, 1, 0, mock.Anything).
		Return([]User{user2}, nil)

	list, err := suite.compositeStore.GetUserListByOUIDs(ctx, ouIDs, 10, 0, nil)
	suite.NoError(err)
	suite.Len(list, 2)
}

// TestCompositeStore_CreateUser tests that create operations go to DB store only.
func (suite *CompositeStoreTestSuite) TestCompositeStore_CreateUser() {
	ctx := context.Background()

	user := User{
		ID:               "new-user",
		OrganizationUnit: "ou-1",
		Type:             "default",
		Attributes:       json.RawMessage(`{"name":"New User"}`),
	}

	credentials := Credentials{
		"password": []Credential{
			{
				Value: "hashed-password",
			},
		},
	}

	// Create only goes to DB store
	suite.mockDBStore.On("CreateUser", ctx, user, credentials).Return(nil)

	err := suite.compositeStore.CreateUser(ctx, user, credentials)
	suite.NoError(err)
	suite.mockDBStore.AssertCalled(suite.T(), "CreateUser", ctx, user, credentials)
	suite.mockFileStore.AssertNotCalled(suite.T(), "CreateUser")
}

// TestCompositeStore_DeleteUser tests that delete operations go to DB store only.
func (suite *CompositeStoreTestSuite) TestCompositeStore_DeleteUser() {
	ctx := context.Background()

	suite.mockDBStore.On("DeleteUser", ctx, "user-id").Return(nil)

	err := suite.compositeStore.DeleteUser(ctx, "user-id")
	suite.NoError(err)
	suite.mockDBStore.AssertCalled(suite.T(), "DeleteUser", ctx, "user-id")
	suite.mockFileStore.AssertNotCalled(suite.T(), "DeleteUser")
}

// TestCompositeStore_IsUserDeclarative_FileStore tests identifying declarative users.
func (suite *CompositeStoreTestSuite) TestCompositeStore_IsUserDeclarative_FileStore() {
	ctx := context.Background()

	// File store has the user, DB store not called
	suite.mockFileStore.On("IsUserDeclarative", ctx, "file-user-1").
		Return(true, nil)

	isDeclarative, err := suite.compositeStore.IsUserDeclarative(ctx, "file-user-1")

	suite.NoError(err)
	suite.True(isDeclarative)
	suite.mockFileStore.AssertCalled(suite.T(), "IsUserDeclarative", ctx, "file-user-1")
	// DB store should not be called since file store found it
	suite.mockDBStore.AssertNotCalled(suite.T(), "IsUserDeclarative")
}

// TestCompositeStore_IsUserDeclarative_DBStore tests identifying mutable users.
func (suite *CompositeStoreTestSuite) TestCompositeStore_IsUserDeclarative_DBStore() {
	ctx := context.Background()

	// File store doesn't have it (not found), so DB store is called
	suite.mockFileStore.On("IsUserDeclarative", ctx, "db-user-1").
		Return(false, ErrUserNotFound)
	suite.mockDBStore.On("IsUserDeclarative", ctx, "db-user-1").
		Return(false, nil)

	isDeclarative, err := suite.compositeStore.IsUserDeclarative(ctx, "db-user-1")

	suite.NoError(err)
	suite.False(isDeclarative)
}

// TestCompositeStore_IsUserDeclarative_NotFound tests error handling for non-existent users.
func (suite *CompositeStoreTestSuite) TestCompositeStore_IsUserDeclarative_NotFound() {
	ctx := context.Background()

	suite.mockFileStore.On("IsUserDeclarative", ctx, "non-existent").
		Return(false, ErrUserNotFound)
	suite.mockDBStore.On("IsUserDeclarative", ctx, "non-existent").
		Return(false, ErrUserNotFound)

	isDeclarative, err := suite.compositeStore.IsUserDeclarative(ctx, "non-existent")
	suite.Error(err)
	suite.False(isDeclarative)
}

// TestCompositeStore_GetUser_FromDB tests retrieving user from DB store.
func (suite *CompositeStoreTestSuite) TestCompositeStore_GetUser_FromDB() {
	ctx := context.Background()

	user := User{
		ID:               "user-db-1",
		OrganizationUnit: "ou-1",
		Type:             "default",
		Attributes:       json.RawMessage(`{"name":"DB User"}`),
	}

	// DB store returns it, file store is not called
	suite.mockDBStore.On("GetUser", ctx, "user-db-1").
		Return(user, nil)

	result, err := suite.compositeStore.GetUser(ctx, "user-db-1")
	suite.NoError(err)
	suite.Equal(user.ID, result.ID)
	suite.mockDBStore.AssertCalled(suite.T(), "GetUser", ctx, "user-db-1")
	suite.mockFileStore.AssertNotCalled(suite.T(), "GetUser")
}

// TestCompositeStore_UpdateUser tests that update operations go to DB store only.
func (suite *CompositeStoreTestSuite) TestCompositeStore_UpdateUser() {
	ctx := context.Background()

	user := &User{
		ID:               "user-1",
		OrganizationUnit: "ou-1",
		Type:             "default",
		Attributes:       json.RawMessage(`{"name":"Updated User"}`),
	}

	suite.mockDBStore.On("UpdateUser", ctx, user).Return(nil)

	err := suite.compositeStore.UpdateUser(ctx, user)
	suite.NoError(err)
	suite.mockDBStore.AssertCalled(suite.T(), "UpdateUser", ctx, user)
	suite.mockFileStore.AssertNotCalled(suite.T(), "UpdateUser")
}

// TestCompositeStore_UpdateUserCredentials tests update credentials route to DB store only.
func (suite *CompositeStoreTestSuite) TestCompositeStore_UpdateUserCredentials() {
	ctx := context.Background()
	creds := Credentials{"password": []Credential{{Value: "hashed"}}}

	suite.mockDBStore.On("UpdateUserCredentials", ctx, "user-1", creds).Return(nil)

	err := suite.compositeStore.UpdateUserCredentials(ctx, "user-1", creds)
	suite.NoError(err)
	suite.mockDBStore.AssertCalled(suite.T(), "UpdateUserCredentials", ctx, "user-1", creds)
	suite.mockFileStore.AssertNotCalled(suite.T(), "UpdateUserCredentials")
}

// TestCompositeStore_GetGroupCountForUser tests group count retrieval from DB store only.
func (suite *CompositeStoreTestSuite) TestCompositeStore_GetGroupCountForUser() {
	ctx := context.Background()

	suite.mockDBStore.On("GetGroupCountForUser", ctx, "user-1").Return(2, nil)

	count, err := suite.compositeStore.GetGroupCountForUser(ctx, "user-1")
	suite.NoError(err)
	suite.Equal(2, count)
}

// TestCompositeStore_GetUserGroups tests group list retrieval from DB store only.
func (suite *CompositeStoreTestSuite) TestCompositeStore_GetUserGroups() {
	ctx := context.Background()
	groups := []UserGroup{{ID: "group-1", Name: "Group 1"}}

	suite.mockDBStore.On("GetUserGroups", ctx, "user-1", 10, 0).Return(groups, nil)

	result, err := suite.compositeStore.GetUserGroups(ctx, "user-1", 10, 0)
	suite.NoError(err)
	suite.Equal(groups, result)
}

// TestCompositeStore_IdentifyUser tests identify fallback logic.
func (suite *CompositeStoreTestSuite) TestCompositeStore_IdentifyUser() {
	ctx := context.Background()
	filters := map[string]interface{}{"username": "alice"}
	userID := "user-1"

	suite.mockDBStore.On("IdentifyUser", ctx, filters).Return(&userID, nil)

	result, err := suite.compositeStore.IdentifyUser(ctx, filters)
	suite.NoError(err)
	suite.Equal(userID, *result)
}

// TestCompositeStore_IdentifyUser_Fallback tests identify fallback to file store.
func (suite *CompositeStoreTestSuite) TestCompositeStore_IdentifyUser_Fallback() {
	ctx := context.Background()
	filters := map[string]interface{}{"username": "alice"}
	userID := "user-2"

	suite.mockDBStore.On("IdentifyUser", ctx, filters).Return((*string)(nil), ErrUserNotFound)
	suite.mockFileStore.On("IdentifyUser", ctx, filters).Return(&userID, nil)

	result, err := suite.compositeStore.IdentifyUser(ctx, filters)
	suite.NoError(err)
	suite.Equal(userID, *result)
}

// TestCompositeStore_GetCredentials tests credentials fallback logic.
func (suite *CompositeStoreTestSuite) TestCompositeStore_GetCredentials() {
	ctx := context.Background()
	user := User{ID: "user-1"}
	creds := Credentials{"password": []Credential{{Value: "hashed"}}}

	suite.mockDBStore.On("GetCredentials", ctx, "user-1").Return(user, creds, nil)

	resultUser, resultCreds, err := suite.compositeStore.GetCredentials(ctx, "user-1")
	suite.NoError(err)
	suite.Equal(user.ID, resultUser.ID)
	suite.Equal(creds["password"][0].Value, resultCreds["password"][0].Value)
}

// TestCompositeStore_GetCredentials_Fallback tests credential fallback to file store.
func (suite *CompositeStoreTestSuite) TestCompositeStore_GetCredentials_Fallback() {
	ctx := context.Background()
	user := User{ID: "user-2"}
	creds := Credentials{"password": []Credential{{Value: "hashed"}}}

	suite.mockDBStore.On("GetCredentials", ctx, "user-2").Return(User{}, nil, ErrUserNotFound)
	suite.mockFileStore.On("GetCredentials", ctx, "user-2").Return(user, creds, nil)

	resultUser, resultCreds, err := suite.compositeStore.GetCredentials(ctx, "user-2")
	suite.NoError(err)
	suite.Equal(user.ID, resultUser.ID)
	suite.Equal(creds["password"][0].Value, resultCreds["password"][0].Value)
}

// TestCompositeStore_ValidateUserIDs tests validation of multiple user IDs.
func (suite *CompositeStoreTestSuite) TestCompositeStore_ValidateUserIDs() {
	ctx := context.Background()

	userIDs := []string{"user-1", "user-2"}

	// ValidateUserIDs calls GetUser for each ID
	// GetUser tries dbStore first
	user1 := User{ID: "user-1"}
	user2 := User{ID: "user-2"}

	suite.mockDBStore.On("GetUser", ctx, "user-1").
		Return(user1, nil)
	suite.mockDBStore.On("GetUser", ctx, "user-2").
		Return(user2, nil)

	result, err := suite.compositeStore.ValidateUserIDs(ctx, userIDs)
	suite.NoError(err)
	suite.Len(result, 0)                                      // All IDs are valid
	suite.mockFileStore.AssertNotCalled(suite.T(), "GetUser") // Not called since dbStore found both
}

// TestCompositeStore_ValidateUserIDs_StoreError tests error propagation from GetUser.
func (suite *CompositeStoreTestSuite) TestCompositeStore_ValidateUserIDs_StoreError() {
	ctx := context.Background()
	userIDs := []string{"user-1"}

	suite.mockDBStore.On("GetUser", ctx, "user-1").Return(User{}, errors.New("db error"))

	result, err := suite.compositeStore.ValidateUserIDs(ctx, userIDs)
	suite.Error(err)
	suite.Nil(result)
}

// TestCompositeStoreTestSuite runs the test suite.
func TestCompositeStoreTestSuite(t *testing.T) {
	suite.Run(t, new(CompositeStoreTestSuite))
}
