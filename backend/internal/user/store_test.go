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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/asgardeo/thunder/internal/system/database/model"
	"github.com/asgardeo/thunder/tests/mocks/database/clientmock"
	"github.com/asgardeo/thunder/tests/mocks/database/providermock"
)

type userStoreTestContext struct {
	providerMock *providermock.DBProviderInterfaceMock
	dbClientMock *clientmock.DBClientInterfaceMock
	store        *userStore
}

func newUserStoreTestContext(t *testing.T) *userStoreTestContext {
	t.Helper()

	providerMock := providermock.NewDBProviderInterfaceMock(t)
	dbClientMock := clientmock.NewDBClientInterfaceMock(t)

	return &userStoreTestContext{
		providerMock: providerMock,
		dbClientMock: dbClientMock,
		store: &userStore{
			dbProvider:  providerMock,
			marshalJSON: json.Marshal,
		},
	}
}

func (ctx *userStoreTestContext) expectDBClient() {
	ctx.providerMock.On("GetDBClient", "identity").Return(ctx.dbClientMock, nil)
}

func matchQueryID(id string) func(interface{}) bool {
	return func(arg interface{}) bool {
		query, ok := arg.(model.DBQuery)
		if !ok {
			return false
		}
		return query.ID == id
	}
}

func sampleUserRow() map[string]interface{} {
	return map[string]interface{}{
		"user_id":    "user1",
		"ou_id":      "ou1",
		"type":       "basic",
		"attributes": []byte(`{"email":"user@example.com"}`),
	}
}

func sampleGroupRow() map[string]interface{} {
	return map[string]interface{}{
		"group_id": "group1",
		"name":     "Engineering",
		"ou_id":    "ou1",
	}
}

func TestUserStore_GetUserListCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		filters    map[string]interface{}
		wantCount  int
		wantErr    string
		setupMocks func(ctx *userStoreTestContext)
	}{
		{
			name:      "success",
			filters:   map[string]interface{}{"type": "basic"},
			wantCount: 4,
			setupMocks: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.dbClientMock.
					On("Query", mock.MatchedBy(matchQueryID("ASQ-USER_MGT-11")), "basic").
					Return([]map[string]interface{}{{"total": int64(4)}}, nil)
			},
		},
		{
			name:    "build query error",
			filters: map[string]interface{}{"invalid-key!": "value"},
			wantErr: "failed to build count query",
			setupMocks: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
			},
		},
		{
			name:    "db client error",
			filters: map[string]interface{}{},
			wantErr: "failed to get database client",
			setupMocks: func(ctx *userStoreTestContext) {
				ctx.providerMock.On("GetDBClient", "identity").Return(nil, errors.New("db error"))
			},
		},
		{
			name:    "query error",
			filters: map[string]interface{}{},
			wantErr: "failed to execute count query",
			setupMocks: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.dbClientMock.
					On("Query", mock.MatchedBy(matchQueryID(QueryGetUserCount.ID))).
					Return(nil, errors.New("query failure"))
			},
		},
		{
			name:    "unexpected total type",
			filters: map[string]interface{}{},
			wantErr: "unexpected type for total",
			setupMocks: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.dbClientMock.
					On("Query", mock.MatchedBy(matchQueryID(QueryGetUserCount.ID))).
					Return([]map[string]interface{}{{"total": "four"}}, nil)
			},
		},
		{
			name:    "empty result returns zero",
			filters: map[string]interface{}{},
			setupMocks: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.dbClientMock.
					On("Query", mock.MatchedBy(matchQueryID(QueryGetUserCount.ID))).
					Return([]map[string]interface{}{}, nil)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := newUserStoreTestContext(t)
			if tc.setupMocks != nil {
				tc.setupMocks(ctx)
			}

			count, err := ctx.store.GetUserListCount(tc.filters)

			if tc.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantCount, count)
		})
	}
}

func TestUserStore_GetUserList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		limit      int
		offset     int
		filters    map[string]interface{}
		wantErr    string
		wantSize   int
		setupMocks func(ctx *userStoreTestContext)
	}{
		{
			name:     "success",
			limit:    10,
			offset:   0,
			filters:  map[string]interface{}{"type": "basic"},
			wantSize: 2,
			setupMocks: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				row1 := sampleUserRow()
				row2 := sampleUserRow()
				row2["user_id"] = "user2"
				row2["attributes"] = []byte(`{"email":"second@example.com"}`)
				ctx.dbClientMock.
					On("Query", mock.MatchedBy(matchQueryID("ASQ-USER_MGT-10")), "basic", 10, 0).
					Return([]map[string]interface{}{row1, row2}, nil)
			},
		},
		{
			name:    "build list query error",
			limit:   5,
			offset:  1,
			filters: map[string]interface{}{"invalid-key!": "value"},
			wantErr: "failed to build list query",
			setupMocks: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
			},
		},
		{
			name:    "db client error",
			wantErr: "failed to get database client",
			setupMocks: func(ctx *userStoreTestContext) {
				ctx.providerMock.On("GetDBClient", "identity").Return(nil, errors.New("db error"))
			},
		},
		{
			name:    "query error",
			wantErr: "failed to execute paginated query",
			setupMocks: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.dbClientMock.
					On("Query", mock.MatchedBy(matchQueryID(QueryGetUserList.ID)), 0, 0).
					Return(nil, errors.New("query failure"))
			},
		},
		{
			name:    "builder error",
			wantErr: "failed to build user from result row",
			setupMocks: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.dbClientMock.
					On("Query", mock.MatchedBy(matchQueryID(QueryGetUserList.ID)), 0, 0).
					Return([]map[string]interface{}{{"ou_id": "ou1"}}, nil)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := newUserStoreTestContext(t)
			if tc.setupMocks != nil {
				tc.setupMocks(ctx)
			}

			users, err := ctx.store.GetUserList(tc.limit, tc.offset, tc.filters)

			if tc.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErr)
				return
			}

			require.NoError(t, err)
			require.Len(t, users, tc.wantSize)
		})
	}
}

func TestUserStore_CreateUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		user        User
		credentials []Credential
		wantErr     string
		wantErrIs   error
		setupMocks  func(ctx *userStoreTestContext)
	}{
		{
			name: "success",
			user: User{
				ID:               "user1",
				OrganizationUnit: "ou1",
				Type:             "basic",
				Attributes:       json.RawMessage(`{"email":"user@example.com"}`),
			},
			setupMocks: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.dbClientMock.
					On("Execute",
						QueryCreateUser,
						"user1",
						"ou1",
						"basic",
						`{"email":"user@example.com"}`,
						"[]").
					Return(int64(1), nil)
			},
		},
		{
			name:    "db client error",
			user:    User{ID: "user1"},
			wantErr: "failed to get database client",
			setupMocks: func(ctx *userStoreTestContext) {
				ctx.providerMock.On("GetDBClient", "identity").Return(nil, errors.New("db error"))
			},
		},
		{
			name:      "attributes marshal error",
			user:      User{ID: "user1", Attributes: json.RawMessage(`{"email":"user@example.com"}`)},
			wantErrIs: ErrBadAttributesInRequest,
			setupMocks: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.store.marshalJSON = func(v interface{}) ([]byte, error) {
					if _, ok := v.(json.RawMessage); ok {
						return nil, errors.New("marshal error")
					}
					return json.Marshal(v)
				}
			},
		},
		{
			name:        "credentials marshal error",
			user:        User{ID: "user1", Attributes: json.RawMessage(`{"email":"user@example.com"}`)},
			credentials: []Credential{{CredentialType: "pwd"}},
			wantErrIs:   ErrBadAttributesInRequest,
			setupMocks: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.store.marshalJSON = func(v interface{}) ([]byte, error) {
					switch v.(type) {
					case []Credential:
						return nil, errors.New("marshal error")
					default:
						return json.Marshal(v)
					}
				}
			},
		},
		{
			name: "execute error",
			user: User{
				ID:               "user1",
				OrganizationUnit: "ou1",
				Type:             "basic",
				Attributes:       json.RawMessage(`{"email":"user@example.com"}`),
			},
			wantErr: "failed to execute query",
			setupMocks: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.dbClientMock.
					On("Execute", QueryCreateUser, "user1", "ou1", "basic", `{"email":"user@example.com"}`, "[]").
					Return(int64(0), errors.New("execute failure"))
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := newUserStoreTestContext(t)
			if tc.setupMocks != nil {
				tc.setupMocks(ctx)
			}

			err := ctx.store.CreateUser(tc.user, tc.credentials)

			if tc.wantErrIs != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.wantErrIs)
				return
			}

			if tc.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestUserStore_GetUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		userID    string
		wantErr   string
		wantErrIs error
		setup     func(ctx *userStoreTestContext)
	}{
		{
			name:   "success",
			userID: "user1",
			setup: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.dbClientMock.
					On("Query", QueryGetUserByUserID, "user1").
					Return([]map[string]interface{}{sampleUserRow()}, nil)
			},
		},
		{
			name:    "db client error",
			userID:  "user1",
			wantErr: "failed to get database client",
			setup: func(ctx *userStoreTestContext) {
				ctx.providerMock.On("GetDBClient", "identity").Return(nil, errors.New("db error"))
			},
		},
		{
			name:    "query error",
			userID:  "user1",
			wantErr: "failed to execute query",
			setup: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.dbClientMock.
					On("Query", QueryGetUserByUserID, "user1").
					Return(nil, errors.New("query failure"))
			},
		},
		{
			name:      "not found",
			userID:    "missing",
			wantErrIs: ErrUserNotFound,
			setup: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.dbClientMock.
					On("Query", QueryGetUserByUserID, "missing").
					Return([]map[string]interface{}{}, nil)
			},
		},
		{
			name:    "multiple results",
			userID:  "dup",
			wantErr: "unexpected number of results",
			setup: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.dbClientMock.
					On("Query", QueryGetUserByUserID, "dup").
					Return([]map[string]interface{}{sampleUserRow(), sampleUserRow()}, nil)
			},
		},
		{
			name:    "builder error",
			userID:  "user1",
			wantErr: "failed to build user from result row",
			setup: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.dbClientMock.
					On("Query", QueryGetUserByUserID, "user1").
					Return([]map[string]interface{}{{"ou_id": "ou1"}}, nil)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := newUserStoreTestContext(t)
			if tc.setup != nil {
				tc.setup(ctx)
			}

			user, err := ctx.store.GetUser(tc.userID)

			if tc.wantErrIs != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.wantErrIs)
				return
			}

			if tc.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.userID, user.ID)
			require.Equal(t, "ou1", user.OrganizationUnit)
		})
	}
}

func TestUserStore_GetGroupCountForUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		userID    string
		want      int
		wantErr   string
		wantErrIs error
		setup     func(ctx *userStoreTestContext)
	}{
		{
			name:   "success",
			userID: "user1",
			want:   3,
			setup: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.dbClientMock.
					On("Query", QueryGetGroupCountForUser, "user1").
					Return([]map[string]interface{}{{"total": int64(3)}}, nil)
			},
		},
		{
			name:    "db client error",
			userID:  "user1",
			wantErr: "failed to get database client",
			setup: func(ctx *userStoreTestContext) {
				ctx.providerMock.On("GetDBClient", "identity").Return(nil, errors.New("db error"))
			},
		},
		{
			name:    "query error",
			userID:  "user1",
			wantErr: "failed to get group count for user",
			setup: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.dbClientMock.
					On("Query", QueryGetGroupCountForUser, "user1").
					Return(nil, errors.New("query failure"))
			},
		},
		{
			name:   "empty results",
			userID: "user1",
			want:   0,
			setup: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.dbClientMock.
					On("Query", QueryGetGroupCountForUser, "user1").
					Return([]map[string]interface{}{}, nil)
			},
		},
		{
			name:    "invalid total type",
			userID:  "user1",
			wantErr: "unexpected type for total",
			setup: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.dbClientMock.
					On("Query", QueryGetGroupCountForUser, "user1").
					Return([]map[string]interface{}{{"total": "three"}}, nil)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := newUserStoreTestContext(t)
			if tc.setup != nil {
				tc.setup(ctx)
			}

			count, err := ctx.store.GetGroupCountForUser(tc.userID)

			if tc.wantErrIs != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.wantErrIs)
				return
			}

			if tc.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want, count)
		})
	}
}

func TestUserStore_GetUserGroups(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		userID     string
		limit      int
		offset     int
		wantErr    string
		wantLen    int
		setupMocks func(ctx *userStoreTestContext)
	}{
		{
			name:    "success",
			userID:  "user1",
			limit:   10,
			offset:  0,
			wantLen: 2,
			setupMocks: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				row1 := sampleGroupRow()
				row2 := sampleGroupRow()
				row2["group_id"] = "group2"
				row2["name"] = "QA"
				ctx.dbClientMock.
					On("Query", QueryGetGroupsForUser, "user1", 10, 0).
					Return([]map[string]interface{}{row1, row2}, nil)
			},
		},
		{
			name:    "db client error",
			userID:  "user1",
			wantErr: "failed to get database client",
			setupMocks: func(ctx *userStoreTestContext) {
				ctx.providerMock.On("GetDBClient", "identity").Return(nil, errors.New("db error"))
			},
		},
		{
			name:    "query error",
			userID:  "user1",
			wantErr: "failed to get groups for user",
			setupMocks: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.dbClientMock.
					On("Query", QueryGetGroupsForUser, "user1", 0, 0).
					Return(nil, errors.New("query failure"))
			},
		},
		{
			name:    "builder error",
			userID:  "user1",
			wantErr: "failed to build group from result row",
			setupMocks: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.dbClientMock.
					On("Query", QueryGetGroupsForUser, "user1", 0, 0).
					Return([]map[string]interface{}{{"name": "QA"}}, nil)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := newUserStoreTestContext(t)
			if tc.setupMocks != nil {
				tc.setupMocks(ctx)
			}

			groups, err := ctx.store.GetUserGroups(tc.userID, tc.limit, tc.offset)

			if tc.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErr)
				return
			}

			require.NoError(t, err)
			require.Len(t, groups, tc.wantLen)
		})
	}
}

func TestUserStore_UpdateUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		user      *User
		wantErr   string
		wantErrIs error
		setup     func(ctx *userStoreTestContext)
	}{
		{
			name: "success",
			user: &User{
				ID:               "user1",
				OrganizationUnit: "ou1",
				Type:             "basic",
				Attributes:       json.RawMessage(`{"email":"user@example.com"}`),
			},
			setup: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.dbClientMock.
					On("Execute", QueryUpdateUserByUserID, "user1", "ou1", "basic", `{"email":"user@example.com"}`).
					Return(int64(1), nil)
			},
		},
		{
			name: "marshal error",
			user: &User{
				ID:         "user1",
				Attributes: json.RawMessage(`{"email":"user@example.com"}`),
			},
			wantErrIs: ErrBadAttributesInRequest,
			setup: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.store.marshalJSON = func(v interface{}) ([]byte, error) {
					return nil, errors.New("marshal error")
				}
			},
		},
		{
			name:    "db client error",
			user:    &User{ID: "user1"},
			wantErr: "failed to get database client",
			setup: func(ctx *userStoreTestContext) {
				ctx.providerMock.On("GetDBClient", "identity").Return(nil, errors.New("db error"))
			},
		},
		{
			name:    "execute error",
			user:    &User{ID: "user1"},
			wantErr: "failed to execute query",
			setup: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.dbClientMock.
					On("Execute", QueryUpdateUserByUserID, "user1", "", "", "null").
					Return(int64(0), errors.New("execute failure"))
			},
		},
		{
			name:      "not found",
			user:      &User{ID: "user1"},
			wantErrIs: ErrUserNotFound,
			setup: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.dbClientMock.
					On("Execute", QueryUpdateUserByUserID, "user1", "", "", "null").
					Return(int64(0), nil)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := newUserStoreTestContext(t)
			if tc.setup != nil {
				tc.setup(ctx)
			}

			err := ctx.store.UpdateUser(tc.user)

			if tc.wantErrIs != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.wantErrIs)
				return
			}

			if tc.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestUserStore_DeleteUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		userID  string
		wantErr string
		setup   func(ctx *userStoreTestContext)
	}{
		{
			name:   "success",
			userID: "user1",
			setup: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.dbClientMock.
					On("Execute", QueryDeleteUserByUserID, "user1").
					Return(int64(1), nil)
			},
		},
		{
			name:    "db client error",
			userID:  "user1",
			wantErr: "failed to get database client",
			setup: func(ctx *userStoreTestContext) {
				ctx.providerMock.On("GetDBClient", "identity").Return(nil, errors.New("db error"))
			},
		},
		{
			name:    "execute error",
			userID:  "user1",
			wantErr: "failed to execute query",
			setup: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.dbClientMock.
					On("Execute", QueryDeleteUserByUserID, "user1").
					Return(int64(0), errors.New("execute failure"))
			},
		},
		{
			name:   "rows affected zero still ok",
			userID: "missing",
			setup: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.dbClientMock.
					On("Execute", QueryDeleteUserByUserID, "missing").
					Return(int64(0), nil)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := newUserStoreTestContext(t)
			if tc.setup != nil {
				tc.setup(ctx)
			}

			err := ctx.store.DeleteUser(tc.userID)

			if tc.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestUserStore_IdentifyUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		filters   map[string]interface{}
		wantErr   string
		wantErrIs error
		setup     func(ctx *userStoreTestContext)
	}{
		{
			name:    "success",
			filters: map[string]interface{}{"email": "user@example.com"},
			setup: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.dbClientMock.
					On("Query", mock.MatchedBy(matchQueryID("ASQ-USER_MGT-08")), "user@example.com").
					Return([]map[string]interface{}{{"user_id": "user1"}}, nil)
			},
		},
		{
			name:    "build query error",
			filters: map[string]interface{}{"bad-key!": "value"},
			wantErr: "failed to build identify query",
			setup: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
			},
		},
		{
			name:    "db client error",
			filters: map[string]interface{}{"id": "user1"},
			wantErr: "failed to get database client",
			setup: func(ctx *userStoreTestContext) {
				ctx.providerMock.On("GetDBClient", "identity").Return(nil, errors.New("db error"))
			},
		},
		{
			name:    "query error",
			filters: map[string]interface{}{"id": "user1"},
			wantErr: "failed to execute query",
			setup: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.dbClientMock.
					On("Query", mock.Anything, "user1").
					Return(nil, errors.New("query failure"))
			},
		},
		{
			name:      "not found",
			filters:   map[string]interface{}{"id": "missing"},
			wantErrIs: ErrUserNotFound,
			setup: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.dbClientMock.
					On("Query", mock.Anything, "missing").
					Return([]map[string]interface{}{}, nil)
			},
		},
		{
			name:    "multiple results",
			filters: map[string]interface{}{"email": "dup@example.com"},
			wantErr: "unexpected number of results",
			setup: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.dbClientMock.
					On("Query", mock.Anything, "dup@example.com").
					Return([]map[string]interface{}{{"user_id": "user1"}, {"user_id": "user2"}}, nil)
			},
		},
		{
			name:    "invalid user_id type",
			filters: map[string]interface{}{"email": "user@example.com"},
			wantErr: "failed to parse user_id as string",
			setup: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.dbClientMock.
					On("Query", mock.Anything, "user@example.com").
					Return([]map[string]interface{}{{"user_id": 123}}, nil)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := newUserStoreTestContext(t)
			if tc.setup != nil {
				tc.setup(ctx)
			}

			userID, err := ctx.store.IdentifyUser(tc.filters)

			if tc.wantErrIs != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.wantErrIs)
				return
			}

			if tc.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErr)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, userID)
			require.Equal(t, "user1", *userID)
		})
	}
}

func TestUserStore_VerifyUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		userID    string
		wantErr   string
		wantErrIs error
		setup     func(ctx *userStoreTestContext)
	}{
		{
			name:   "success",
			userID: "user1",
			setup: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				row := sampleUserRow()
				row["credentials"] = `[{"credentialType":"pwd"}]`
				ctx.dbClientMock.
					On("Query", QueryValidateUserWithCredentials, "user1").
					Return([]map[string]interface{}{row}, nil)
			},
		},
		{
			name:    "db client error",
			userID:  "user1",
			wantErr: "failed to get database client",
			setup: func(ctx *userStoreTestContext) {
				ctx.providerMock.On("GetDBClient", "identity").Return(nil, errors.New("db error"))
			},
		},
		{
			name:    "query error",
			userID:  "user1",
			wantErr: "failed to execute query",
			setup: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.dbClientMock.
					On("Query", QueryValidateUserWithCredentials, "user1").
					Return(nil, errors.New("query failure"))
			},
		},
		{
			name:      "not found",
			userID:    "missing",
			wantErrIs: ErrUserNotFound,
			setup: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.dbClientMock.
					On("Query", QueryValidateUserWithCredentials, "missing").
					Return([]map[string]interface{}{}, nil)
			},
		},
		{
			name:    "multiple results",
			userID:  "dup",
			wantErr: "unexpected number of results",
			setup: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				row := sampleUserRow()
				row["credentials"] = "[]"
				ctx.dbClientMock.
					On("Query", QueryValidateUserWithCredentials, "dup").
					Return([]map[string]interface{}{row, row}, nil)
			},
		},
		{
			name:    "builder error",
			userID:  "user1",
			wantErr: "failed to build user from result row",
			setup: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.dbClientMock.
					On("Query", QueryValidateUserWithCredentials, "user1").
					Return([]map[string]interface{}{{"credentials": "[]"}}, nil)
			},
		},
		{
			name:    "invalid credentials type",
			userID:  "user1",
			wantErr: "failed to parse credentials as string",
			setup: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				row := sampleUserRow()
				row["credentials"] = 5
				ctx.dbClientMock.
					On("Query", QueryValidateUserWithCredentials, "user1").
					Return([]map[string]interface{}{row}, nil)
			},
		},
		{
			name:    "unmarshal credentials error",
			userID:  "user1",
			wantErr: "failed to unmarshal credentials",
			setup: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				row := sampleUserRow()
				row["credentials"] = "invalid json"
				ctx.dbClientMock.
					On("Query", QueryValidateUserWithCredentials, "user1").
					Return([]map[string]interface{}{row}, nil)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := newUserStoreTestContext(t)
			if tc.setup != nil {
				tc.setup(ctx)
			}

			user, credentials, err := ctx.store.VerifyUser(tc.userID)

			if tc.wantErrIs != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.wantErrIs)
				return
			}

			if tc.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.userID, user.ID)
			require.NotEmpty(t, credentials)
		})
	}
}

func TestUserStore_ValidateUserIDs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		userIDs         []string
		wantErr         string
		wantResult      []string
		setup           func(ctx *userStoreTestContext)
		overrideBuilder func([]string) (model.DBQuery, []interface{}, error)
	}{
		{
			name:       "empty input",
			userIDs:    []string{},
			wantResult: []string{},
		},
		{
			name:    "db client error",
			userIDs: []string{"user1"},
			wantErr: "failed to get database client",
			setup: func(ctx *userStoreTestContext) {
				ctx.providerMock.On("GetDBClient", "identity").Return(nil, errors.New("db error"))
			},
		},
		{
			name:    "build query error",
			userIDs: []string{"user1"},
			wantErr: "failed to build bulk user exists query",
			setup: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
			},
			overrideBuilder: func([]string) (model.DBQuery, []interface{}, error) {
				return model.DBQuery{}, nil, errors.New("builder error")
			},
		},
		{
			name:    "query error",
			userIDs: []string{"user1"},
			wantErr: "failed to execute query",
			setup: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.dbClientMock.
					On("Query", mock.MatchedBy(matchQueryID("ASQ-USER_MGT-09")), "user1").
					Return(nil, errors.New("query failure"))
			},
		},
		{
			name:       "returns invalid IDs",
			userIDs:    []string{"user1", "user2"},
			wantResult: []string{"user2"},
			setup: func(ctx *userStoreTestContext) {
				ctx.expectDBClient()
				ctx.dbClientMock.
					On("Query", mock.MatchedBy(matchQueryID("ASQ-USER_MGT-09")), "user1", "user2").
					Return([]map[string]interface{}{{"user_id": "user1"}}, nil)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := newUserStoreTestContext(t)
			if tc.overrideBuilder != nil {
				originalBuilder := bulkUserExistsQueryBuilder
				bulkUserExistsQueryBuilder = tc.overrideBuilder
				t.Cleanup(func() {
					bulkUserExistsQueryBuilder = originalBuilder
				})
			}
			if tc.setup != nil {
				tc.setup(ctx)
			}

			result, err := ctx.store.ValidateUserIDs(tc.userIDs)

			if tc.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantResult, result)
		})
	}
}

func TestUserStore_BuildUserFromResultRow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		row     map[string]interface{}
		wantErr string
	}{
		{
			name: "success",
			row: map[string]interface{}{
				"user_id":    "user1",
				"ou_id":      "ou1",
				"type":       "basic",
				"attributes": []byte(`{"key":"value"}`),
			},
		},
		{
			name:    "missing user_id",
			row:     map[string]interface{}{"ou_id": "ou1", "type": "basic"},
			wantErr: "failed to parse user_id as string",
		},
		{
			name:    "missing org id",
			row:     map[string]interface{}{"user_id": "user1", "type": "basic"},
			wantErr: "failed to parse org_id as string",
		},
		{
			name:    "missing type",
			row:     map[string]interface{}{"user_id": "user1", "ou_id": "ou1"},
			wantErr: "failed to parse type as string",
		},
		{
			name: "invalid attributes type",
			row: map[string]interface{}{
				"user_id":    "user1",
				"ou_id":      "ou1",
				"type":       "basic",
				"attributes": 123,
			},
			wantErr: "failed to parse attributes as string",
		},
		{
			name: "invalid attributes json",
			row: map[string]interface{}{
				"user_id":    "user1",
				"ou_id":      "ou1",
				"type":       "basic",
				"attributes": []byte(`invalid`),
			},
			wantErr: "failed to unmarshal attributes",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			user, err := buildUserFromResultRow(tc.row)

			if tc.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.row["user_id"], user.ID)
			require.Equal(t, tc.row["ou_id"], user.OrganizationUnit)
		})
	}
}

func TestUserStore_BuildGroupFromResultRow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		row     map[string]interface{}
		wantErr string
	}{
		{
			name: "success",
			row: map[string]interface{}{
				"group_id": "group1",
				"name":     "Engineering",
				"ou_id":    "ou1",
			},
		},
		{
			name:    "missing group_id",
			row:     map[string]interface{}{"name": "Engineering", "ou_id": "ou1"},
			wantErr: "failed to parse group_id as string",
		},
		{
			name:    "missing name",
			row:     map[string]interface{}{"group_id": "group1", "ou_id": "ou1"},
			wantErr: "failed to parse name as string",
		},
		{
			name:    "missing ou_id",
			row:     map[string]interface{}{"group_id": "group1", "name": "Engineering"},
			wantErr: "failed to parse ou_id as string",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			group, err := buildGroupFromResultRow(tc.row)

			if tc.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.row["group_id"], group.ID)
			require.Equal(t, tc.row["name"], group.Name)
		})
	}
}

func TestUserStore_MaskMapValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   map[string]interface{}
		out  map[string]interface{}
	}{
		{
			name: "masks string values",
			in:   map[string]interface{}{"password": "secret", "username": "alice"},
			out:  map[string]interface{}{"password": "s****t", "username": "a***e"},
		},
		{
			name: "masks non-string values",
			in:   map[string]interface{}{"attempts": 3, "enabled": true},
			out:  map[string]interface{}{"attempts": "***", "enabled": "***"},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := maskMapValues(tc.in)
			assert.Equal(t, tc.out, result)
		})
	}
}
