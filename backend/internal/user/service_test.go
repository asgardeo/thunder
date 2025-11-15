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

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	oupkg "github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	// "github.com/asgardeo/thunder/internal/system/hash"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/userschema"
	"github.com/asgardeo/thunder/tests/mocks/oumock"
	"github.com/asgardeo/thunder/tests/mocks/userschemamock"
)

const testUserID = "user1"

func TestUserService_ValidateUserAndUniqueness(t *testing.T) {
	const employeeType = "employee"
	payload := []byte(`{"email":"employee@example.com"}`)

	tests := []struct {
		name    string
		setup   func(t *testing.T) (*userService, *userschemamock.UserSchemaServiceInterfaceMock, *userStoreInterfaceMock)
		nilAttr bool
		wantErr *serviceerror.ServiceError
		assert  func(t *testing.T, schemaMock *userschemamock.UserSchemaServiceInterfaceMock,
			storeMock *userStoreInterfaceMock)
	}{
		{
			name: "Internal Error When Schema Validation Fails",
			setup: func(t *testing.T) (*userService, *userschemamock.UserSchemaServiceInterfaceMock, *userStoreInterfaceMock) {
				schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				schemaMock.
					On("ValidateUser", employeeType, mock.Anything).
					Return(false, &serviceerror.ServiceError{
						Code:  "USRS-5000",
						Type:  serviceerror.ServerErrorType,
						Error: "schema validation failed",
					}).
					Once()

				return &userService{
					userSchemaService: schemaMock,
				}, schemaMock, nil
			},
			wantErr: &ErrorInternalServerError,
			assert: func(t *testing.T, schemaMock *userschemamock.UserSchemaServiceInterfaceMock, _ *userStoreInterfaceMock) {
				schemaMock.AssertNotCalled(t, "ValidateUserUniqueness", mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			name: "User Schema Not Found When Schema Missing",
			setup: func(t *testing.T) (*userService, *userschemamock.UserSchemaServiceInterfaceMock, *userStoreInterfaceMock) {
				schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				schemaMock.
					On("ValidateUser", employeeType, mock.Anything).
					Return(false, &userschema.ErrorUserSchemaNotFound).
					Once()

				return &userService{userSchemaService: schemaMock}, schemaMock, nil
			},
			wantErr: &ErrorUserSchemaNotFound,
			assert: func(t *testing.T, schemaMock *userschemamock.UserSchemaServiceInterfaceMock, _ *userStoreInterfaceMock) {
				schemaMock.AssertNotCalled(t, "ValidateUserUniqueness", mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			name: "Internal Error When Schema Lookup Fails",
			setup: func(t *testing.T) (*userService, *userschemamock.UserSchemaServiceInterfaceMock, *userStoreInterfaceMock) {
				schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				schemaMock.
					On("ValidateUser", employeeType, mock.Anything).
					Return(false, &serviceerror.ServiceError{
						Code:  "USRS-5000",
						Type:  serviceerror.ServerErrorType,
						Error: "unexpected error",
					}).
					Once()

				return &userService{userSchemaService: schemaMock}, schemaMock, nil
			},
			wantErr: &ErrorInternalServerError,
			assert: func(t *testing.T, schemaMock *userschemamock.UserSchemaServiceInterfaceMock, _ *userStoreInterfaceMock) {
				schemaMock.AssertNotCalled(t, "ValidateUserUniqueness", mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			name: "Schema Validation Failed",
			setup: func(t *testing.T) (*userService, *userschemamock.UserSchemaServiceInterfaceMock, *userStoreInterfaceMock) {
				schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				schemaMock.
					On("ValidateUser", employeeType, mock.Anything).
					Return(false, nil).
					Once()

				return &userService{userSchemaService: schemaMock}, schemaMock, nil
			},
			wantErr: &ErrorSchemaValidationFailed,
			assert: func(t *testing.T, schemaMock *userschemamock.UserSchemaServiceInterfaceMock, _ *userStoreInterfaceMock) {
				schemaMock.AssertNotCalled(t, "ValidateUserUniqueness", mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			name: "Internal Error When Uniqueness Validation Fails",
			setup: func(t *testing.T) (*userService, *userschemamock.UserSchemaServiceInterfaceMock, *userStoreInterfaceMock) {
				schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				schemaMock.
					On("ValidateUser", employeeType, mock.Anything).
					Return(true, nil).
					Once()
				schemaMock.
					On("ValidateUserUniqueness", employeeType, mock.Anything, mock.Anything).
					Return(false, &serviceerror.ServiceError{
						Code:  "USRS-5000",
						Type:  serviceerror.ServerErrorType,
						Error: "validation failed",
					}).
					Once()

				return &userService{userSchemaService: schemaMock}, schemaMock, nil
			},
			wantErr: &ErrorInternalServerError,
		},
		{
			name: "User Schema Not Found When Uniqueness Schema Missing",
			setup: func(t *testing.T) (*userService, *userschemamock.UserSchemaServiceInterfaceMock, *userStoreInterfaceMock) {
				schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				schemaMock.
					On("ValidateUser", employeeType, mock.Anything).
					Return(true, nil).
					Once()
				schemaMock.
					On("ValidateUserUniqueness", employeeType, mock.Anything, mock.Anything).
					Return(false, &userschema.ErrorUserSchemaNotFound).
					Once()

				return &userService{userSchemaService: schemaMock}, schemaMock, nil
			},
			wantErr: &ErrorUserSchemaNotFound,
		},
		{
			name: "Attribute Conflict When Uniqueness Check Fails",
			setup: func(t *testing.T) (*userService, *userschemamock.UserSchemaServiceInterfaceMock, *userStoreInterfaceMock) {
				existingUserID := "user-123"
				schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				storeMock := newUserStoreInterfaceMock(t)
				storeMock.
					On("IdentifyUser", mock.AnythingOfType("map[string]interface {}")).
					Return(&existingUserID, nil).
					Once()
				schemaMock.
					On("ValidateUser", employeeType, mock.Anything).
					Return(true, nil).
					Once()
				schemaMock.
					On("ValidateUserUniqueness", employeeType, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						identify := args.Get(2).(func(map[string]interface{}) (*string, error))

						id, err := identify(map[string]interface{}{"email": "employee@example.com"})
						require.NoError(t, err)
						require.NotNil(t, id)
						require.Equal(t, existingUserID, *id)
					}).
					Return(false, nil).
					Once()

				return &userService{
					userSchemaService: schemaMock,
					userStore:         storeMock,
				}, schemaMock, storeMock
			},
			wantErr: &ErrorAttributeConflict,
		},
		{
			name: "Returns Nil When Validation Succeeds",
			setup: func(t *testing.T) (*userService, *userschemamock.UserSchemaServiceInterfaceMock, *userStoreInterfaceMock) {
				schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				storeMock := newUserStoreInterfaceMock(t)
				storeMock.
					On("IdentifyUser", mock.AnythingOfType("map[string]interface {}")).
					Return((*string)(nil), nil).
					Once()
				schemaMock.
					On("ValidateUser", employeeType, mock.Anything).
					Return(true, nil).
					Once()
				schemaMock.
					On("ValidateUserUniqueness", employeeType, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						identify := args.Get(2).(func(map[string]interface{}) (*string, error))

						id, err := identify(map[string]interface{}{"email": "employee@example.com"})
						require.NoError(t, err)
						require.Nil(t, id)
					}).
					Return(true, nil).
					Once()

				return &userService{
					userSchemaService: schemaMock,
					userStore:         storeMock,
				}, schemaMock, storeMock
			},
		},
		{
			name: "Internal Error When Identify Fails",
			setup: func(t *testing.T) (*userService, *userschemamock.UserSchemaServiceInterfaceMock, *userStoreInterfaceMock) {
				schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				storeMock := newUserStoreInterfaceMock(t)
				storeMock.
					On("IdentifyUser", mock.AnythingOfType("map[string]interface {}")).
					Return((*string)(nil), errors.New("store failure")).
					Once()
				schemaMock.
					On("ValidateUser", employeeType, mock.Anything).
					Return(true, nil).
					Once()
				schemaMock.
					On("ValidateUserUniqueness", employeeType, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						identify := args.Get(2).(func(map[string]interface{}) (*string, error))

						id, err := identify(map[string]interface{}{"email": "employee@example.com"})
						require.Error(t, err)
						require.Nil(t, id)
					}).
					Return(false, &serviceerror.ServiceError{
						Code:  "USRS-5000",
						Type:  serviceerror.ServerErrorType,
						Error: "validation failed",
					}).
					Once()

				return &userService{
					userSchemaService: schemaMock,
					userStore:         storeMock,
				}, schemaMock, storeMock
			},
			wantErr: &ErrorInternalServerError,
		},
		{
			name:    "Handles Nil Attributes",
			nilAttr: true,
			setup: func(t *testing.T) (*userService, *userschemamock.UserSchemaServiceInterfaceMock, *userStoreInterfaceMock) {
				schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				storeMock := newUserStoreInterfaceMock(t)
				storeMock.
					On("IdentifyUser", mock.AnythingOfType("map[string]interface {}")).
					Return((*string)(nil), nil).
					Once()
				schemaMock.
					On("ValidateUser", employeeType, mock.Anything).
					Run(func(args mock.Arguments) {
						attr, _ := args.Get(1).([]byte)
						require.Nil(t, attr)
					}).
					Return(true, nil).
					Once()
				schemaMock.
					On("ValidateUserUniqueness", employeeType, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						attr, _ := args.Get(1).([]byte)
						require.Nil(t, attr)
						identify := args.Get(2).(func(map[string]interface{}) (*string, error))

						id, err := identify(map[string]interface{}{"email": "employee@example.com"})
						require.NoError(t, err)
						require.Nil(t, id)
					}).
					Return(true, nil).
					Once()

				return &userService{
					userSchemaService: schemaMock,
					userStore:         storeMock,
				}, schemaMock, storeMock
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			service, schemaMock, storeMock := tc.setup(t)
			logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserServiceTest"))

			attributes := payload
			if tc.nilAttr {
				attributes = nil
			}

			err := service.validateUserAndUniqueness(employeeType, attributes, logger)

			if tc.wantErr != nil {
				require.NotNil(t, err)
				require.Equal(t, *tc.wantErr, *err)
			} else {
				require.Nil(t, err)
			}

			if tc.assert != nil {
				tc.assert(t, schemaMock, storeMock)
			}

			schemaMock.AssertExpectations(t)
			if storeMock != nil {
				storeMock.AssertExpectations(t)
			}
		})
	}
}

type userServiceFixture struct {
	service *userService
	store   *userStoreInterfaceMock
	ou      *oumock.OrganizationUnitServiceInterfaceMock
	schema  *userschemamock.UserSchemaServiceInterfaceMock
}

func newUserServiceFixture(t *testing.T) *userServiceFixture {
	storeMock := newUserStoreInterfaceMock(t)
	ouMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
	schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)

	return &userServiceFixture{
		service: &userService{
			userStore:         storeMock,
			ouService:         ouMock,
			userSchemaService: schemaMock,
		},
		store:  storeMock,
		ou:     ouMock,
		schema: schemaMock,
	}
}

func (f *userServiceFixture) expectValidationSuccess() {
	f.schema.
		On("ValidateUser", "basic", mock.Anything).
		Return(true, nil).
		Once()
	f.schema.
		On("ValidateUserUniqueness", "basic", mock.Anything, mock.Anything).
		Return(true, nil).
		Once()
}

func TestUserService_GetUserList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		limit   int
		offset  int
		filters map[string]interface{}
		setup   func(f *userServiceFixture)
		wantErr *serviceerror.ServiceError
	}{
		{name: "invalid pagination", limit: 0, offset: 0, wantErr: &ErrorInvalidLimit},
		{
			name:   "count error",
			limit:  5,
			offset: 0,
			setup: func(f *userServiceFixture) {
				f.store.
					On("GetUserListCount", mock.Anything).
					Return(0, errors.New("count err")).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
		{
			name:   "list error",
			limit:  5,
			offset: 0,
			setup: func(f *userServiceFixture) {
				f.store.
					On("GetUserListCount", mock.Anything).
					Return(2, nil).
					Once()
				f.store.
					On("GetUserList", 5, 0, mock.Anything).
					Return(nil, errors.New("list err")).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
		{
			name:    "success",
			limit:   2,
			offset:  1,
			filters: map[string]interface{}{"type": "basic"},
			setup: func(f *userServiceFixture) {
				f.store.
					On("GetUserListCount", mock.Anything).
					Return(2, nil).
					Once()
				f.store.
					On("GetUserList", 2, 1, mock.Anything).
					Return([]User{{ID: testUserID}, {ID: "user2"}}, nil).
					Once()
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			fixture := newUserServiceFixture(t)
			if tc.setup != nil {
				tc.setup(fixture)
			}

			resp, err := fixture.service.GetUserList(tc.limit, tc.offset, tc.filters)

			if tc.wantErr != nil {
				require.Nil(t, resp)
				require.Equal(t, tc.wantErr.Code, err.Code)
				return
			}

			require.Nil(t, err)
			require.Equal(t, tc.limit, resp.Count)
			require.Equal(t, tc.offset+1, resp.StartIndex)
			require.Len(t, resp.Users, 2)
		})
	}
}

func TestUserService_GetUsersByPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		path    string
		limit   int
		setup   func(f *userServiceFixture)
		wantErr *serviceerror.ServiceError
	}{
		{name: "invalid path", path: " ", limit: 5, wantErr: &ErrorInvalidHandlePath},
		{
			name:  "ou not found",
			path:  "root",
			limit: 5,
			setup: func(f *userServiceFixture) {
				f.ou.
					On("GetOrganizationUnitByPath", "root").
					Return(oupkg.OrganizationUnit{}, &oupkg.ErrorOrganizationUnitNotFound).
					Once()
			},
			wantErr: &ErrorOrganizationUnitNotFound,
		},
		{
			name:  "ou service error",
			path:  "root",
			limit: 5,
			setup: func(f *userServiceFixture) {
				f.ou.
					On("GetOrganizationUnitByPath", "root").
					Return(oupkg.OrganizationUnit{}, &serviceerror.ServiceError{Code: "OU-5000"}).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
		{
			name:  "invalid pagination",
			path:  "root",
			limit: 0,
			setup: func(f *userServiceFixture) {
				f.ou.
					On("GetOrganizationUnitByPath", "root").
					Return(oupkg.OrganizationUnit{ID: "ou1"}, nil).
					Once()
			},
			wantErr: &ErrorInvalidLimit,
		},
		{
			name:  "ou users error",
			path:  "root",
			limit: 5,
			setup: func(f *userServiceFixture) {
				f.ou.
					On("GetOrganizationUnitByPath", "root").
					Return(oupkg.OrganizationUnit{ID: "ou1"}, nil).
					Once()
				f.ou.
					On("GetOrganizationUnitUsers", "ou1", 5, 0).
					Return((*oupkg.UserListResponse)(nil), &serviceerror.ServiceError{Code: "USR-5000"}).
					Once()
			},
			wantErr: &serviceerror.ServiceError{Code: "USR-5000"},
		},
		{
			name:  "success",
			path:  "root",
			limit: 5,
			setup: func(f *userServiceFixture) {
				f.ou.
					On("GetOrganizationUnitByPath", "root").
					Return(oupkg.OrganizationUnit{ID: "ou1"}, nil).
					Once()
				f.ou.
					On("GetOrganizationUnitUsers", "ou1", 5, 0).
					Return(&oupkg.UserListResponse{
						TotalResults: 1,
						StartIndex:   1,
						Count:        1,
						Users:        []oupkg.User{{ID: testUserID}},
					}, nil).
					Once()
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			fixture := newUserServiceFixture(t)
			if tc.setup != nil {
				tc.setup(fixture)
			}

			resp, err := fixture.service.GetUsersByPath(tc.path, tc.limit, 0, map[string]interface{}{})

			if tc.wantErr != nil {
				require.Nil(t, resp)
				require.Equal(t, tc.wantErr.Code, err.Code)
				return
			}

			require.Nil(t, err)
			require.Len(t, resp.Users, 1)
		})
	}
}

func TestUserService_CreateUser(t *testing.T) {
	tests := []struct {
		name    string
		user    *User
		setup   func(f *userServiceFixture)
		wantErr *serviceerror.ServiceError
	}{
		{name: "nil user", user: nil, wantErr: &ErrorInvalidRequestFormat},
		{
			name: "validation error",
			user: &User{Type: "basic"},
			setup: func(f *userServiceFixture) {
				f.schema.
					On("ValidateUser", "basic", mock.Anything).
					Return(false, nil).
					Once()
			},
			wantErr: &ErrorSchemaValidationFailed,
		},
		{
			name: "extract credentials error",
			user: &User{Type: "basic", Attributes: json.RawMessage("{{")},
			setup: func(f *userServiceFixture) {
				f.expectValidationSuccess()
			},
			wantErr: &ErrorInternalServerError,
		},
		// {
		// 	name: "store error",
		// 	user: &User{Type: "basic", Attributes: json.RawMessage(`{"password":"secret"}`)},
		// 	setup: func(f *userServiceFixture) {
		// 		f.expectValidationSuccess()
		// 		f.store.
		// 			On("CreateUser", mock.AnythingOfType("user.User"), mock.Anything).
		// 			Return(errors.New("create err")).
		// 			Once()
		// 	},
		// 	wantErr: &ErrorInternalServerError,
		// },
		// {
		// 	name: "success",
		// 	user: &User{Type: "basic", Attributes: json.RawMessage(`{"password":"secret"}`)},
		// 	setup: func(f *userServiceFixture) {
		// 		f.expectValidationSuccess()
		// 		f.store.
		// 			On("CreateUser", mock.AnythingOfType("user.User"), mock.Anything).
		// 			Return(nil).
		// 			Once()
		// 	},
		// },
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			fixture := newUserServiceFixture(t)
			if tc.setup != nil {
				tc.setup(fixture)
			}

			result, err := fixture.service.CreateUser(tc.user)

			if tc.wantErr != nil {
				require.Nil(t, result)
				require.Equal(t, tc.wantErr.Code, err.Code)
				return
			}

			require.Nil(t, err)
			require.NotNil(t, result)
			require.NotEmpty(t, result.ID)
		})
	}
}

func TestUserService_CreateUserByPath(t *testing.T) {
	req := CreateUserByPathRequest{
		Type:       "basic",
		Attributes: json.RawMessage(`{"password":"secret"}`),
	}

	tests := []struct {
		name    string
		path    string
		setup   func(f *userServiceFixture)
		wantErr *serviceerror.ServiceError
	}{
		{name: "invalid path", path: " ", wantErr: &ErrorInvalidHandlePath},
		{
			name: "ou not found",
			path: "root",
			setup: func(f *userServiceFixture) {
				f.ou.
					On("GetOrganizationUnitByPath", "root").
					Return(oupkg.OrganizationUnit{}, &oupkg.ErrorOrganizationUnitNotFound).
					Once()
			},
			wantErr: &ErrorOrganizationUnitNotFound,
		},
		{
			name: "ou service error",
			path: "root",
			setup: func(f *userServiceFixture) {
				f.ou.
					On("GetOrganizationUnitByPath", "root").
					Return(oupkg.OrganizationUnit{}, &serviceerror.ServiceError{Code: "OU-5000"}).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
		// {
		// 	name: "success",
		// 	path: "root",
		// 	setup: func(f *userServiceFixture) {
		// 		f.ou.
		// 			On("GetOrganizationUnitByPath", "root").
		// 			Return(oupkg.OrganizationUnit{ID: "ou1"}, nil).
		// 			Once()
		// 		f.expectValidationSuccess()
		// 		f.store.
		// 			On("CreateUser", mock.AnythingOfType("user.User"), mock.Anything).
		// 			Return(nil).
		// 			Once()
		// 	},
		// },
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			fixture := newUserServiceFixture(t)
			if tc.setup != nil {
				tc.setup(fixture)
			}

			result, err := fixture.service.CreateUserByPath(tc.path, req)

			if tc.wantErr != nil {
				require.Nil(t, result)
				require.Equal(t, tc.wantErr.Code, err.Code)
				return
			}

			require.Nil(t, err)
		})
	}
}

func TestExtractCredentials(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		user := &User{Attributes: json.RawMessage("{{")}
		creds, err := extractCredentials(user)
		require.Error(t, err)
		require.Nil(t, creds)
	})

	// t.Run("removes credential fields", func(t *testing.T) {
	// 	user := &User{Attributes: json.RawMessage(`{"password":"secret","custom":"value"}`)}
	// 	creds, err := extractCredentials(user)

	// 	require.NoError(t, err)
	// 	require.Len(t, creds, 1)

	// 	var attrs map[string]interface{}
	// 	require.NoError(t, json.Unmarshal(user.Attributes, &attrs))
	// 	require.Equal(t, map[string]interface{}{"custom": "value"}, attrs)
	// })
}

func TestUserService_GetUserGroups(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		userID  string
		limit   int
		offset  int
		setup   func(f *userServiceFixture)
		wantErr *serviceerror.ServiceError
	}{
		{name: "missing id", userID: "", limit: 5, wantErr: &ErrorMissingUserID},
		{name: "invalid pagination", userID: testUserID, limit: 0, wantErr: &ErrorInvalidLimit},
		{
			name:   "validate ids error",
			userID: testUserID,
			limit:  5,
			setup: func(f *userServiceFixture) {
				f.store.
					On("ValidateUserIDs", []string{testUserID}).
					Return(nil, errors.New("boom")).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
		{
			name:   "user not found",
			userID: testUserID,
			limit:  5,
			setup: func(f *userServiceFixture) {
				f.store.
					On("ValidateUserIDs", []string{testUserID}).
					Return([]string{testUserID}, nil).
					Once()
			},
			wantErr: &ErrorUserNotFound,
		},
		{
			name:   "group count error",
			userID: testUserID,
			limit:  5,
			setup: func(f *userServiceFixture) {
				f.store.
					On("ValidateUserIDs", []string{testUserID}).
					Return([]string{}, nil).
					Once()
				f.store.
					On("GetGroupCountForUser", testUserID).
					Return(0, errors.New("count err")).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
		{
			name:   "group list error",
			userID: testUserID,
			limit:  5,
			setup: func(f *userServiceFixture) {
				f.store.
					On("ValidateUserIDs", []string{testUserID}).
					Return([]string{}, nil).
					Once()
				f.store.
					On("GetGroupCountForUser", testUserID).
					Return(1, nil).
					Once()
				f.store.
					On("GetUserGroups", testUserID, 5, 0).
					Return(nil, errors.New("list err")).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
		{
			name:   "success",
			userID: testUserID,
			limit:  5,
			offset: 3,
			setup: func(f *userServiceFixture) {
				f.store.
					On("ValidateUserIDs", []string{testUserID}).
					Return([]string{}, nil).
					Once()
				f.store.
					On("GetGroupCountForUser", testUserID).
					Return(2, nil).
					Once()
				f.store.
					On("GetUserGroups", testUserID, 5, 3).
					Return([]UserGroup{{ID: "grp1"}}, nil).
					Once()
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			fixture := newUserServiceFixture(t)
			if tc.setup != nil {
				tc.setup(fixture)
			}

			resp, err := fixture.service.GetUserGroups(tc.userID, tc.limit, tc.offset)

			if tc.wantErr != nil {
				require.Nil(t, resp)
				require.Equal(t, tc.wantErr.Code, err.Code)
				return
			}

			require.Nil(t, err)
			require.NotNil(t, resp)
			require.Equal(t, 2, resp.TotalResults)
			require.Len(t, resp.Groups, 1)
			require.NotEmpty(t, resp.Links)
		})
	}
}

func TestUserService_UpdateUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		userID  string
		user    *User
		setup   func(f *userServiceFixture)
		wantErr *serviceerror.ServiceError
	}{
		{name: "missing id", userID: "", user: &User{}, wantErr: &ErrorMissingUserID},
		{name: "nil user", userID: testUserID, user: nil, wantErr: &ErrorInvalidRequestFormat},
		{
			name:   "validation error",
			userID: testUserID,
			user:   &User{Type: "basic"},
			setup: func(f *userServiceFixture) {
				f.schema.
					On("ValidateUser", "basic", mock.Anything).
					Return(false, nil).
					Once()
			},
			wantErr: &ErrorSchemaValidationFailed,
		},
		{
			name:   "user not found",
			userID: testUserID,
			user:   &User{Type: "basic"},
			setup: func(f *userServiceFixture) {
				f.expectValidationSuccess()
				f.store.
					On("UpdateUser", mock.AnythingOfType("*user.User")).
					Return(ErrUserNotFound).
					Once()
			},
			wantErr: &ErrorUserNotFound,
		},
		{
			name:   "store error",
			userID: testUserID,
			user:   &User{Type: "basic"},
			setup: func(f *userServiceFixture) {
				f.expectValidationSuccess()
				f.store.
					On("UpdateUser", mock.AnythingOfType("*user.User")).
					Return(errors.New("update err")).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
		{
			name:   "success",
			userID: testUserID,
			user:   &User{Type: "basic"},
			setup: func(f *userServiceFixture) {
				f.expectValidationSuccess()
				f.store.
					On("UpdateUser", mock.AnythingOfType("*user.User")).
					Return(nil).
					Once()
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			fixture := newUserServiceFixture(t)
			if tc.setup != nil {
				tc.setup(fixture)
			}

			res, err := fixture.service.UpdateUser(tc.userID, tc.user)

			if tc.wantErr != nil {
				require.Nil(t, res)
				require.Equal(t, tc.wantErr.Code, err.Code)
				return
			}

			require.Nil(t, err)
			require.Equal(t, tc.user, res)
		})
	}
}

func TestUserService_DeleteUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		userID  string
		setup   func(f *userServiceFixture)
		wantErr *serviceerror.ServiceError
	}{
		{name: "missing id", userID: "", wantErr: &ErrorMissingUserID},
		{
			name:   "user not found",
			userID: testUserID,
			setup: func(f *userServiceFixture) {
				f.store.
					On("DeleteUser", testUserID).
					Return(ErrUserNotFound).
					Once()
			},
			wantErr: &ErrorUserNotFound,
		},
		{
			name:   "store error",
			userID: testUserID,
			setup: func(f *userServiceFixture) {
				f.store.
					On("DeleteUser", testUserID).
					Return(errors.New("delete err")).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
		{
			name:   "success",
			userID: testUserID,
			setup: func(f *userServiceFixture) {
				f.store.
					On("DeleteUser", testUserID).
					Return(nil).
					Once()
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			fixture := newUserServiceFixture(t)
			if tc.setup != nil {
				tc.setup(fixture)
			}

			err := fixture.service.DeleteUser(tc.userID)

			if tc.wantErr != nil {
				require.Equal(t, tc.wantErr.Code, err.Code)
				return
			}

			require.Nil(t, err)
		})
	}
}

func TestUserService_IdentifyUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		filters map[string]interface{}
		setup   func(f *userServiceFixture)
		wantErr *serviceerror.ServiceError
	}{
		{name: "missing filters", filters: map[string]interface{}{}, wantErr: &ErrorInvalidRequestFormat},
		{
			name:    "user not found",
			filters: map[string]interface{}{"email": "test"},
			setup: func(f *userServiceFixture) {
				f.store.
					On("IdentifyUser", mock.Anything).
					Return((*string)(nil), ErrUserNotFound).
					Once()
			},
			wantErr: &ErrorUserNotFound,
		},
		{
			name:    "store error",
			filters: map[string]interface{}{"email": "test"},
			setup: func(f *userServiceFixture) {
				f.store.
					On("IdentifyUser", mock.Anything).
					Return((*string)(nil), errors.New("boom")).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
		{
			name:    "success",
			filters: map[string]interface{}{"email": "test"},
			setup: func(f *userServiceFixture) {
				id := testUserID
				f.store.
					On("IdentifyUser", mock.Anything).
					Return(&id, nil).
					Once()
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			fixture := newUserServiceFixture(t)
			if tc.setup != nil {
				tc.setup(fixture)
			}

			id, err := fixture.service.IdentifyUser(tc.filters)

			if tc.wantErr != nil {
				require.Nil(t, id)
				require.Equal(t, tc.wantErr.Code, err.Code)
				return
			}

			require.Nil(t, err)
			require.Equal(t, testUserID, *id)
		})
	}
}

// func TestUserService_VerifyUser(t *testing.T) {
// 	validCred := hash.NewCredential([]byte("secret"))

// 	tests := []struct {
// 		name    string
// 		userID  string
// 		creds   map[string]interface{}
// 		setup   func(f *userServiceFixture)
// 		wantErr *serviceerror.ServiceError
// 	}{
// 		{name: "missing id", userID: "", creds: map[string]interface{}{"password": "secret"}, wantErr: &ErrorMissingUserID},
// 		{name: "missing creds", userID: testUserID, creds: map[string]interface{}{}, wantErr: &ErrorInvalidRequestFormat},
// 		{
// 			name:    "unsupported creds",
// 			userID:  testUserID,
// 			creds:   map[string]interface{}{"custom": "value"},
// 			wantErr: &ErrorAuthenticationFailed,
// 		},
// 		{
// 			name:   "store user not found",
// 			userID: testUserID,
// 			creds:  map[string]interface{}{"password": "secret"},
// 			setup: func(f *userServiceFixture) {
// 				f.store.
// 					On("VerifyUser", testUserID).
// 					Return(User{}, nil, ErrUserNotFound).
// 					Once()
// 			},
// 			wantErr: &ErrorUserNotFound,
// 		},
// 		{
// 			name:   "store error",
// 			userID: testUserID,
// 			creds:  map[string]interface{}{"password": "secret"},
// 			setup: func(f *userServiceFixture) {
// 				f.store.
// 					On("VerifyUser", testUserID).
// 					Return(User{}, nil, errors.New("boom")).
// 					Once()
// 			},
// 			wantErr: &ErrorInternalServerError,
// 		},
// 		{
// 			name:   "no stored credentials",
// 			userID: testUserID,
// 			creds:  map[string]interface{}{"password": "secret"},
// 			setup: func(f *userServiceFixture) {
// 				f.store.
// 					On("VerifyUser", testUserID).
// 					Return(User{}, []Credential{}, nil).
// 					Once()
// 			},
// 			wantErr: &ErrorAuthenticationFailed,
// 		},
// 		{
// 			name:   "missing credential type",
// 			userID: testUserID,
// 			creds:  map[string]interface{}{"password": "secret"},
// 			setup: func(f *userServiceFixture) {
// 				f.store.
// 					On("VerifyUser", testUserID).
// 					Return(User{}, []Credential{
// 						{CredentialType: "pin", StorageAlgo: validCred.Algorithm, Value: validCred.Hash, Salt: validCred.Salt},
// 					}, nil).
// 					Once()
// 			},
// 			wantErr: &ErrorAuthenticationFailed,
// 		},
// 		{
// 			name:   "success",
// 			userID: testUserID,
// 			creds:  map[string]interface{}{"password": "secret"},
// 			setup: func(f *userServiceFixture) {
// 				f.store.
// 					On("VerifyUser", testUserID).
// 					Return(User{ID: testUserID}, []Credential{
// 						{CredentialType: "password", StorageAlgo: validCred.Algorithm, Value: validCred.Hash, Salt: validCred.Salt},
// 					}, nil).
// 					Once()
// 			},
// 		},
// 	}

// 	for _, tc := range tests {
// 		tc := tc
// 		t.Run(tc.name, func(t *testing.T) {
// 			fixture := newUserServiceFixture(t)
// 			if tc.setup != nil {
// 				tc.setup(fixture)
// 			}

// 			user, err := fixture.service.VerifyUser(tc.userID, tc.creds)

// 			if tc.wantErr != nil {
// 				require.Nil(t, user)
// 				require.Equal(t, tc.wantErr.Code, err.Code)
// 				return
// 			}

// 			require.Nil(t, err)
// 			require.Equal(t, tc.userID, user.ID)
// 		})
// 	}
// }

// func TestUserService_AuthenticateUser(t *testing.T) {
// 	validCred := hash.NewCredential([]byte("secret"))

// 	tests := []struct {
// 		name    string
// 		request AuthenticateUserRequest
// 		setup   func(f *userServiceFixture)
// 		wantErr *serviceerror.ServiceError
// 	}{
// 		{name: "empty request", request: AuthenticateUserRequest{}, wantErr: &ErrorInvalidRequestFormat},
// 		{
// 			name:    "missing filters",
// 			request: AuthenticateUserRequest{"password": "secret"},
// 			wantErr: &ErrorMissingRequiredFields,
// 		},
// 		{
// 			name:    "missing credentials",
// 			request: AuthenticateUserRequest{"email": "user@example.com"},
// 			wantErr: &ErrorMissingCredentials,
// 		},
// 		{
// 			name:    "user not found",
// 			request: AuthenticateUserRequest{"email": "user@example.com", "password": "secret"},
// 			setup: func(f *userServiceFixture) {
// 				f.store.
// 					On("IdentifyUser", mock.Anything).
// 					Return((*string)(nil), ErrUserNotFound).
// 					Once()
// 			},
// 			wantErr: &ErrorUserNotFound,
// 		},
// 		{
// 			name:    "identify error",
// 			request: AuthenticateUserRequest{"email": "user@example.com", "password": "secret"},
// 			setup: func(f *userServiceFixture) {
// 				f.store.
// 					On("IdentifyUser", mock.Anything).
// 					Return((*string)(nil), errors.New("boom")).
// 					Once()
// 			},
// 			wantErr: &ErrorInternalServerError,
// 		},
// 		{
// 			name:    "verify error",
// 			request: AuthenticateUserRequest{"email": "user@example.com", "password": "secret"},
// 			setup: func(f *userServiceFixture) {
// 				id := testUserID
// 				f.store.
// 					On("IdentifyUser", mock.Anything).
// 					Return(&id, nil).
// 					Once()
// 				f.store.
// 					On("VerifyUser", testUserID).
// 					Return(User{}, []Credential{}, nil).
// 					Once()
// 			},
// 			wantErr: &ErrorAuthenticationFailed,
// 		},
// 		{
// 			name:    "success",
// 			request: AuthenticateUserRequest{"email": "user@example.com", "password": "secret"},
// 			setup: func(f *userServiceFixture) {
// 				id := testUserID
// 				f.store.
// 					On("IdentifyUser", mock.Anything).
// 					Return(&id, nil).
// 					Once()
// 				f.store.
// 					On("VerifyUser", testUserID).
// 					Return(User{ID: testUserID, Type: "basic", OrganizationUnit: "ou1"}, []Credential{
// 						{CredentialType: "password", StorageAlgo: validCred.Algorithm, Value: validCred.Hash, Salt: validCred.Salt},
// 					}, nil).
// 					Once()
// 			},
// 		},
// 	}

// 	for _, tc := range tests {
// 		tc := tc
// 		t.Run(tc.name, func(t *testing.T) {
// 			fixture := newUserServiceFixture(t)
// 			if tc.setup != nil {
// 				tc.setup(fixture)
// 			}

// 			resp, err := fixture.service.AuthenticateUser(tc.request)

// 			if tc.wantErr != nil {
// 				require.Nil(t, resp)
// 				require.Equal(t, tc.wantErr.Code, err.Code)
// 				return
// 			}

// 			require.Nil(t, err)
// 			require.Equal(t, testUserID, resp.ID)
// 		})
// 	}
// }

func TestUserService_ValidateUserIDs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		userIDs []string
		setup   func(f *userServiceFixture)
		wantErr *serviceerror.ServiceError
	}{
		{name: "empty", userIDs: []string{}},
		{
			name:    "store error",
			userIDs: []string{testUserID},
			setup: func(f *userServiceFixture) {
				f.store.
					On("ValidateUserIDs", []string{testUserID}).
					Return(nil, errors.New("boom")).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
		{
			name:    "success",
			userIDs: []string{testUserID},
			setup: func(f *userServiceFixture) {
				f.store.
					On("ValidateUserIDs", []string{testUserID}).
					Return([]string{"user2"}, nil).
					Once()
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			fixture := newUserServiceFixture(t)
			if tc.setup != nil {
				tc.setup(fixture)
			}

			result, err := fixture.service.ValidateUserIDs(tc.userIDs)

			if tc.wantErr != nil {
				require.Nil(t, result)
				require.Equal(t, tc.wantErr.Code, err.Code)
				return
			}

			require.Nil(t, err)
		})
	}
}

func TestUserService_GetUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		userID  string
		setup   func(f *userServiceFixture)
		wantErr *serviceerror.ServiceError
	}{
		{name: "missing id", userID: "", wantErr: &ErrorMissingUserID},
		{
			name:   "not found",
			userID: testUserID,
			setup: func(f *userServiceFixture) {
				f.store.
					On("GetUser", testUserID).
					Return(User{}, ErrUserNotFound).
					Once()
			},
			wantErr: &ErrorUserNotFound,
		},
		{
			name:   "store error",
			userID: testUserID,
			setup: func(f *userServiceFixture) {
				f.store.
					On("GetUser", testUserID).
					Return(User{}, errors.New("boom")).
					Once()
			},
			wantErr: &ErrorInternalServerError,
		},
		{
			name:   "success",
			userID: testUserID,
			setup: func(f *userServiceFixture) {
				f.store.
					On("GetUser", testUserID).
					Return(User{ID: testUserID}, nil).
					Once()
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			fixture := newUserServiceFixture(t)
			if tc.setup != nil {
				tc.setup(fixture)
			}

			user, err := fixture.service.GetUser(tc.userID)

			if tc.wantErr != nil {
				require.Nil(t, user)
				require.Equal(t, tc.wantErr.Code, err.Code)
				return
			}

			require.Nil(t, err)
			require.Equal(t, tc.userID, user.ID)
		})
	}
}
