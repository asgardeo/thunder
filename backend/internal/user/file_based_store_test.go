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
	"testing"

	"github.com/stretchr/testify/suite"

	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/declarative_resource/entity"
)

// FileBasedStoreTestSuite tests the user file-based store behavior.
type FileBasedStoreTestSuite struct {
	suite.Suite
	store *userFileBasedStore
}

// TestFileBasedStoreTestSuite runs the test suite.
func TestFileBasedStoreTestSuite(t *testing.T) {
	suite.Run(t, new(FileBasedStoreTestSuite))
}

// SetupTest sets up a fresh store instance for each test.
func (suite *FileBasedStoreTestSuite) SetupTest() {
	suite.store = &userFileBasedStore{
		GenericFileBasedStore: declarativeresource.NewGenericFileBasedStoreForTest(entity.KeyTypeUser),
	}
}

// TestNewUserFileBasedStore verifies the constructor returns a usable store.
func (suite *FileBasedStoreTestSuite) TestNewUserFileBasedStore() {
	store, _ := newUserFileBasedStore()
	suite.NotNil(store)
	_, ok := store.(*userFileBasedStore)
	suite.True(ok)
}

func (suite *FileBasedStoreTestSuite) buildUser(id, ou string, attrs map[string]interface{}) User {
	payload, err := json.Marshal(attrs)
	suite.NoError(err)

	return User{
		ID:         id,
		OUID:       ou,
		Type:       "person",
		Attributes: payload,
	}
}

// TestCreate_InvalidType verifies Create rejects invalid resource types.
func (suite *FileBasedStoreTestSuite) TestCreate_InvalidType() {
	err := suite.store.Create("user-1", "invalid")
	suite.Error(err)
}

// TestCreateAndGetUser verifies CreateUser and GetUser round trip.
func (suite *FileBasedStoreTestSuite) TestCreateAndGetUser() {
	user := suite.buildUser("user-1", "ou-1", map[string]interface{}{
		"username": "alice",
		"email":    "alice@example.com",
	})
	credentials := Credentials{
		"password": []Credential{{Value: "hashed"}},
	}

	err := suite.store.CreateUser(context.Background(), user, credentials)
	suite.NoError(err)

	got, err := suite.store.GetUser(context.Background(), "user-1")
	suite.NoError(err)
	suite.Equal(user.ID, got.ID)
}

// TestGetUser_NotFound verifies GetUser returns ErrUserNotFound for missing users.
func (suite *FileBasedStoreTestSuite) TestGetUser_NotFound() {
	_, err := suite.store.GetUser(context.Background(), "missing")
	suite.Error(err)
	suite.Equal(ErrUserNotFound, err)
}

// TestGetUserListCountAndList_Filtering verifies filtering and pagination.
func (suite *FileBasedStoreTestSuite) TestGetUserListCountAndList_Filtering() {
	user1 := suite.buildUser("user-1", "ou-1", map[string]interface{}{
		"username": "alice",
		"age":      30,
		"active":   true,
		"profile": map[string]interface{}{
			"email": "alice@example.com",
		},
	})
	user2 := suite.buildUser("user-2", "ou-1", map[string]interface{}{
		"username": "bob",
		"age":      28,
	})

	suite.NoError(suite.store.CreateUser(context.Background(), user1, nil))
	suite.NoError(suite.store.CreateUser(context.Background(), user2, nil))

	filters := map[string]interface{}{
		"username":      "alice",
		"profile.email": "alice@example.com",
		"age":           30,
		"active":        true,
	}

	count, err := suite.store.GetUserListCount(context.Background(), filters)
	suite.NoError(err)
	suite.Equal(1, count)

	list, err := suite.store.GetUserList(context.Background(), 1, 0, filters)
	suite.NoError(err)
	suite.Len(list, 1)
	suite.Equal("user-1", list[0].ID)
}

// TestGetUserListCount_InvalidAttributes verifies invalid JSON attributes are filtered out.
func (suite *FileBasedStoreTestSuite) TestGetUserListCount_InvalidAttributes() {
	user := User{
		ID:         "user-1",
		OUID:       "ou-1",
		Type:       "person",
		Attributes: json.RawMessage("{invalid-json"),
	}

	suite.NoError(suite.store.CreateUser(context.Background(), user, nil))

	count, err := suite.store.GetUserListCount(context.Background(), map[string]interface{}{"username": "alice"})
	suite.NoError(err)
	suite.Equal(0, count)
}

// TestGetUserListByOUIDs verifies OU scoping for list and count methods.
func (suite *FileBasedStoreTestSuite) TestGetUserListByOUIDs() {
	user1 := suite.buildUser("user-1", "ou-1", map[string]interface{}{"username": "alice"})
	user2 := suite.buildUser("user-2", "ou-2", map[string]interface{}{"username": "bob"})

	suite.NoError(suite.store.CreateUser(context.Background(), user1, nil))
	suite.NoError(suite.store.CreateUser(context.Background(), user2, nil))

	count, err := suite.store.GetUserListCountByOUIDs(context.Background(), []string{"ou-1"}, nil)
	suite.NoError(err)
	suite.Equal(1, count)

	list, err := suite.store.GetUserListByOUIDs(context.Background(), []string{"ou-1"}, 10, 0, nil)
	suite.NoError(err)
	suite.Len(list, 1)
	suite.Equal("user-1", list[0].ID)
}

// TestIdentifyUser verifies IdentifyUser behavior for zero, one, and multiple matches.
func (suite *FileBasedStoreTestSuite) TestIdentifyUser() {
	user1 := suite.buildUser("user-1", "ou-1", map[string]interface{}{
		"username": "alice",
		"email":    "dup@example.com",
	})
	user2 := suite.buildUser("user-2", "ou-1", map[string]interface{}{
		"username": "bob",
		"email":    "dup@example.com",
	})

	suite.NoError(suite.store.CreateUser(context.Background(), user1, nil))
	suite.NoError(suite.store.CreateUser(context.Background(), user2, nil))

	id, err := suite.store.IdentifyUser(context.Background(), map[string]interface{}{"username": "alice"})
	suite.NoError(err)
	suite.Equal("user-1", *id)

	_, err = suite.store.IdentifyUser(context.Background(), map[string]interface{}{"email": "dup@example.com"})
	suite.Error(err)
	suite.ErrorIs(err, ErrAmbiguousUser)

	_, err = suite.store.IdentifyUser(context.Background(), map[string]interface{}{"email": "missing@example.com"})
	suite.Error(err)
	suite.Equal(ErrUserNotFound, err)
}

// TestSearchUsers verifies SearchUsers behavior for zero, one, and multiple matches.
func (suite *FileBasedStoreTestSuite) TestSearchUsers() {
	user1 := suite.buildUser("user-1", "ou-1", map[string]interface{}{
		"username": "alice",
		"email":    "dup@example.com",
	})
	user2 := suite.buildUser("user-2", "ou-1", map[string]interface{}{
		"username": "bob",
		"email":    "dup@example.com",
	})

	suite.NoError(suite.store.CreateUser(context.Background(), user1, nil))
	suite.NoError(suite.store.CreateUser(context.Background(), user2, nil))

	// Single match
	users, err := suite.store.SearchUsers(context.Background(), map[string]interface{}{"username": "alice"})
	suite.NoError(err)
	suite.Len(users, 1)
	suite.Equal("user-1", users[0].ID)

	// Multiple matches
	users, err = suite.store.SearchUsers(context.Background(), map[string]interface{}{"email": "dup@example.com"})
	suite.NoError(err)
	suite.Len(users, 2)

	// No matches
	_, err = suite.store.SearchUsers(context.Background(), map[string]interface{}{"email": "missing@example.com"})
	suite.Error(err)
	suite.ErrorIs(err, ErrUserNotFound)
}

// TestGetCredentials verifies credentials can be retrieved for a stored user.
func (suite *FileBasedStoreTestSuite) TestGetCredentials() {
	user := suite.buildUser("user-1", "ou-1", map[string]interface{}{"username": "alice"})
	credentials := Credentials{
		"password": []Credential{{Value: "hashed"}},
	}

	suite.NoError(suite.store.CreateUser(context.Background(), user, credentials))

	gotUser, gotCreds, err := suite.store.GetCredentials(context.Background(), "user-1")
	suite.NoError(err)
	suite.Equal(user.ID, gotUser.ID)
	suite.Equal(credentials["password"][0].Value, gotCreds["password"][0].Value)
}

// TestGetCredentials_NotFound verifies missing users return ErrUserNotFound.
func (suite *FileBasedStoreTestSuite) TestGetCredentials_NotFound() {
	_, _, err := suite.store.GetCredentials(context.Background(), "missing")
	suite.Error(err)
	suite.Equal(ErrUserNotFound, err)
}

// TestGetCredentials_Corrupted verifies corrupted data returns an error.
func (suite *FileBasedStoreTestSuite) TestGetCredentials_Corrupted() {
	suite.NoError(suite.store.GenericFileBasedStore.Create("corrupt", "invalid"))
	_, _, err := suite.store.GetCredentials(context.Background(), "corrupt")
	suite.Error(err)
}

// TestValidateUserIDs verifies detection of missing IDs.
func (suite *FileBasedStoreTestSuite) TestValidateUserIDs() {
	user := suite.buildUser("user-1", "ou-1", map[string]interface{}{"username": "alice"})
	suite.NoError(suite.store.CreateUser(context.Background(), user, nil))

	invalid, err := suite.store.ValidateUserIDs(context.Background(), []string{"user-1", "missing"})
	suite.NoError(err)
	suite.Equal([]string{"missing"}, invalid)
}

// TestIsUserDeclarative verifies declarative checks for existing and corrupted users.
func (suite *FileBasedStoreTestSuite) TestIsUserDeclarative() {
	user := suite.buildUser("user-1", "ou-1", map[string]interface{}{"username": "alice"})
	suite.NoError(suite.store.CreateUser(context.Background(), user, nil))

	ok, err := suite.store.IsUserDeclarative(context.Background(), "user-1")
	suite.NoError(err)
	suite.True(ok)

	missing, err := suite.store.IsUserDeclarative(context.Background(), "missing")
	suite.NoError(err)
	suite.False(missing)

	suite.NoError(suite.store.GenericFileBasedStore.Create("corrupt", "invalid"))
	corrupt, err := suite.store.IsUserDeclarative(context.Background(), "corrupt")
	suite.Error(err)
	suite.False(corrupt)
}

// TestUnsupportedOperations verifies unsupported operations return errors.
func (suite *FileBasedStoreTestSuite) TestUnsupportedOperations() {
	user := &User{ID: "user-1"}
	creds := Credentials{"password": []Credential{{Value: "hashed"}}}

	err := suite.store.UpdateUser(context.Background(), user)
	suite.Error(err)

	err = suite.store.UpdateUserCredentials(context.Background(), "user-1", creds)
	suite.Error(err)

	err = suite.store.DeleteUser(context.Background(), "user-1")
	suite.Error(err)
}

// TestGroupMethods verifies group-related methods return empty results.
func (suite *FileBasedStoreTestSuite) TestGroupMethods() {
	count, err := suite.store.GetGroupCountForUser(context.Background(), "user-1")
	suite.NoError(err)
	suite.Equal(0, count)

	groups, err := suite.store.GetUserGroups(context.Background(), "user-1", 10, 0)
	suite.NoError(err)
	suite.Len(groups, 0)
}

// TestValidateUserIDsInOUs_EmptyUserIDs verifies short-circuit when no user IDs provided.
func (suite *FileBasedStoreTestSuite) TestValidateUserIDsInOUs_EmptyUserIDs() {
	out, err := suite.store.ValidateUserIDsInOUs(context.Background(), []string{}, []string{"ou-1"})
	suite.NoError(err)
	suite.Empty(out)
}

// TestValidateUserIDsInOUs_EmptyOUIDs verifies that all user IDs are out-of-scope when no OUs given.
func (suite *FileBasedStoreTestSuite) TestValidateUserIDsInOUs_EmptyOUIDs() {
	userIDs := []string{"usr-001", "usr-002"}
	out, err := suite.store.ValidateUserIDsInOUs(context.Background(), userIDs, []string{})
	suite.NoError(err)
	suite.Equal(userIDs, out)
}

// TestValidateUserIDsInOUs covers the core OU-scope logic using real in-memory users.
func (suite *FileBasedStoreTestSuite) TestValidateUserIDsInOUs() {
	ctx := context.Background()

	// Seed two users in different OUs.
	user1 := suite.buildUser("usr-001", "ou-1", map[string]interface{}{"username": "alice"})
	user2 := suite.buildUser("usr-002", "ou-2", map[string]interface{}{"username": "bob"})
	suite.NoError(suite.store.CreateUser(ctx, user1, nil))
	suite.NoError(suite.store.CreateUser(ctx, user2, nil))

	testCases := []struct {
		name           string
		userIDs        []string
		ouIDs          []string
		wantOutOfScope []string
	}{
		{
			name:           "user in scope",
			userIDs:        []string{"usr-001"},
			ouIDs:          []string{"ou-1"},
			wantOutOfScope: []string{},
		},
		{
			name:           "user out of scope",
			userIDs:        []string{"usr-001"},
			ouIDs:          []string{"ou-2"},
			wantOutOfScope: []string{"usr-001"},
		},
		{
			name:           "user not found treated as out of scope",
			userIDs:        []string{"usr-missing"},
			ouIDs:          []string{"ou-1"},
			wantOutOfScope: []string{"usr-missing"},
		},
		{
			name:           "mixed in scope and out of scope",
			userIDs:        []string{"usr-001", "usr-002"},
			ouIDs:          []string{"ou-1"},
			wantOutOfScope: []string{"usr-002"},
		},
		{
			name:           "all users in scope across multiple OUs",
			userIDs:        []string{"usr-001", "usr-002"},
			ouIDs:          []string{"ou-1", "ou-2"},
			wantOutOfScope: []string{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			out, err := suite.store.ValidateUserIDsInOUs(ctx, tc.userIDs, tc.ouIDs)
			suite.NoError(err)
			suite.Equal(tc.wantOutOfScope, out)
		})
	}
}
