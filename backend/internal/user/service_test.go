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
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	oupkg "github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/crypto/hash"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/userschema"
	"github.com/asgardeo/thunder/tests/mocks/crypto/hashmock"
	"github.com/asgardeo/thunder/tests/mocks/oumock"
	"github.com/asgardeo/thunder/tests/mocks/userschemamock"
)

const testUserType = "employee"

func TestOUStore_ValidateUserAndUniqueness(t *testing.T) {
	type testMocks struct {
		schemaService *userschemamock.UserSchemaServiceInterfaceMock
		userStore     *userStoreInterfaceMock
	}

	payloadWithEmail := []byte(`{"email":"employee@example.com"}`)
	emptyPayload := []byte(`{}`)

	testCases := []struct {
		name    string
		payload []byte
		setup   func(t *testing.T) (*userService, testMocks)
		assert  func(t *testing.T, err *serviceerror.ServiceError, mocks testMocks)
	}{
		{
			name:    "ReturnsInternalErrorWhenSchemaValidationFails",
			payload: payloadWithEmail,
			setup: func(t *testing.T) (*userService, testMocks) {
				schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				schemaMock.
					On("ValidateUser", testUserType, mock.Anything).
					Return(false, &serviceerror.ServiceError{
						Code:  "USRS-5000",
						Type:  serviceerror.ServerErrorType,
						Error: "schema validation failed",
					}).
					Once()

				return &userService{
					userSchemaService: schemaMock,
				}, testMocks{schemaService: schemaMock}
			},
			assert: func(t *testing.T, err *serviceerror.ServiceError, mocks testMocks) {
				require.NotNil(t, err)
				require.Equal(t, ErrorInternalServerError.Code, err.Code)
				mocks.schemaService.AssertNotCalled(t, "ValidateUserUniqueness",
					mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			name:    "ReturnsUserSchemaNotFoundWhenSchemaMissing",
			payload: emptyPayload,
			setup: func(t *testing.T) (*userService, testMocks) {
				schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				schemaMock.
					On("ValidateUser", testUserType, mock.Anything).
					Return(false, &userschema.ErrorUserSchemaNotFound).
					Once()

				return &userService{
					userSchemaService: schemaMock,
				}, testMocks{schemaService: schemaMock}
			},
			assert: func(t *testing.T, err *serviceerror.ServiceError, mocks testMocks) {
				require.NotNil(t, err)
				require.Equal(t, ErrorUserSchemaNotFound, *err)
				mocks.schemaService.AssertNotCalled(t, "ValidateUserUniqueness",
					mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			name:    "ReturnsInternalErrorWhenSchemaLookupFails",
			payload: emptyPayload,
			setup: func(t *testing.T) (*userService, testMocks) {
				schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				schemaMock.
					On("ValidateUser", testUserType, mock.Anything).
					Return(false, &serviceerror.ServiceError{
						Code:  "USRS-5000",
						Type:  serviceerror.ServerErrorType,
						Error: "unexpected error",
					}).
					Once()

				return &userService{
					userSchemaService: schemaMock,
				}, testMocks{schemaService: schemaMock}
			},
			assert: func(t *testing.T, err *serviceerror.ServiceError, mocks testMocks) {
				require.NotNil(t, err)
				require.Equal(t, ErrorInternalServerError, *err)
				mocks.schemaService.AssertNotCalled(t, "ValidateUserUniqueness",
					mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			name:    "ReturnsSchemaValidationFailed",
			payload: payloadWithEmail,
			setup: func(t *testing.T) (*userService, testMocks) {
				schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				schemaMock.
					On("ValidateUser", testUserType, mock.Anything).
					Return(false, nil).
					Once()

				return &userService{
					userSchemaService: schemaMock,
				}, testMocks{schemaService: schemaMock}
			},
			assert: func(t *testing.T, err *serviceerror.ServiceError, mocks testMocks) {
				require.NotNil(t, err)
				require.Equal(t, ErrorSchemaValidationFailed, *err)
				mocks.schemaService.AssertNotCalled(t, "ValidateUserUniqueness",
					mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			name:    "ReturnsInternalErrorWhenUniquenessValidationFails",
			payload: payloadWithEmail,
			setup: func(t *testing.T) (*userService, testMocks) {
				schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				schemaMock.
					On("ValidateUser", testUserType, mock.Anything).
					Return(true, nil).
					Once()
				schemaMock.
					On("ValidateUserUniqueness", testUserType, mock.Anything, mock.Anything).
					Return(false, &serviceerror.ServiceError{
						Code:  "USRS-5000",
						Type:  serviceerror.ServerErrorType,
						Error: "validation failed",
					}).
					Once()

				return &userService{
					userSchemaService: schemaMock,
				}, testMocks{schemaService: schemaMock}
			},
			assert: func(t *testing.T, err *serviceerror.ServiceError, mocks testMocks) {
				require.NotNil(t, err)
				require.Equal(t, ErrorInternalServerError, *err)
			},
		},
		{
			name:    "ReturnsUserSchemaNotFoundWhenUniquenessSchemaMissing",
			payload: payloadWithEmail,
			setup: func(t *testing.T) (*userService, testMocks) {
				schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				schemaMock.
					On("ValidateUser", testUserType, mock.Anything).
					Return(true, nil).
					Once()
				schemaMock.
					On("ValidateUserUniqueness", testUserType, mock.Anything, mock.Anything).
					Return(false, &userschema.ErrorUserSchemaNotFound).
					Once()

				return &userService{
					userSchemaService: schemaMock,
				}, testMocks{schemaService: schemaMock}
			},
			assert: func(t *testing.T, err *serviceerror.ServiceError, mocks testMocks) {
				require.NotNil(t, err)
				require.Equal(t, ErrorUserSchemaNotFound, *err)
			},
		},
		{
			name:    "ReturnsAttributeConflictWhenUniquenessCheckFails",
			payload: payloadWithEmail,
			setup: func(t *testing.T) (*userService, testMocks) {
				existingUserID := "user-123"
				schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				userStoreMock := newUserStoreInterfaceMock(t)
				userStoreMock.
					On("IdentifyUser", mock.Anything, mock.AnythingOfType("map[string]interface {}")).
					Return(&existingUserID, nil).
					Once()
				schemaMock.
					On("ValidateUser", testUserType, mock.Anything).
					Return(true, nil).
					Once()
				schemaMock.
					On("ValidateUserUniqueness", testUserType, mock.Anything, mock.Anything).
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
						userStore:         userStoreMock,
					}, testMocks{
						schemaService: schemaMock,
						userStore:     userStoreMock,
					}
			},
			assert: func(t *testing.T, err *serviceerror.ServiceError, mocks testMocks) {
				require.NotNil(t, err)
				require.Equal(t, ErrorAttributeConflict, *err)
			},
		},
		{
			name:    "ReturnsNilWhenValidationSucceeds",
			payload: payloadWithEmail,
			setup: func(t *testing.T) (*userService, testMocks) {
				schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				userStoreMock := newUserStoreInterfaceMock(t)
				userStoreMock.
					On("IdentifyUser", mock.Anything, mock.AnythingOfType("map[string]interface {}")).
					Return((*string)(nil), nil).
					Once()
				schemaMock.
					On("ValidateUser", testUserType, mock.Anything).
					Return(true, nil).
					Once()
				schemaMock.
					On("ValidateUserUniqueness", testUserType, mock.Anything, mock.Anything).
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
						userStore:         userStoreMock,
					}, testMocks{
						schemaService: schemaMock,
						userStore:     userStoreMock,
					}
			},
			assert: func(t *testing.T, err *serviceerror.ServiceError, mocks testMocks) {
				require.Nil(t, err)
			},
		},
		{
			name:    "ReturnsInternalErrorWhenIdentifyFails",
			payload: payloadWithEmail,
			setup: func(t *testing.T) (*userService, testMocks) {
				schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				userStoreMock := newUserStoreInterfaceMock(t)
				userStoreMock.
					On("IdentifyUser", mock.Anything, mock.AnythingOfType("map[string]interface {}")).
					Return((*string)(nil), errors.New("store failure")).
					Once()
				schemaMock.
					On("ValidateUser", testUserType, mock.Anything).
					Return(true, nil).
					Once()
				schemaMock.
					On("ValidateUserUniqueness", testUserType, mock.Anything, mock.Anything).
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
						userStore:         userStoreMock,
					}, testMocks{
						schemaService: schemaMock,
						userStore:     userStoreMock,
					}
			},
			assert: func(t *testing.T, err *serviceerror.ServiceError, mocks testMocks) {
				require.NotNil(t, err)
				require.Equal(t, ErrorInternalServerError, *err)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			service, mocks := tc.setup(t)
			logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserServiceTest"))

			err := service.validateUserAndUniqueness(testUserType, tc.payload, logger)
			tc.assert(t, err, mocks)
		})
	}
}

func TestOUStore_ValidateOrganizationUnitForUserType(t *testing.T) {
	type testMocks struct {
		ouService         *oumock.OrganizationUnitServiceInterfaceMock
		userSchemaService *userschemamock.UserSchemaServiceInterfaceMock
	}

	testCases := []struct {
		name        string
		userType    string
		ouID        string
		setup       func(t *testing.T) (*userService, testMocks)
		expectedErr *serviceerror.ServiceError
	}{
		{
			name:     "ReturnsErrorWhenIDEmpty",
			userType: testUserType,
			ouID:     "",
			setup: func(t *testing.T) (*userService, testMocks) {
				return &userService{}, testMocks{}
			},
			expectedErr: &ErrorInvalidOrganizationUnitID,
		},
		{
			name:     "ReturnsErrorWhenIDInvalid",
			userType: testUserType,
			ouID:     "invalid-id",
			setup: func(t *testing.T) (*userService, testMocks) {
				return &userService{}, testMocks{}
			},
			expectedErr: &ErrorInvalidOrganizationUnitID,
		},
		{
			name:     "ReturnsErrorWhenOrganizationUnitMissing",
			userType: testUserType,
			ouID:     "4d8b40d6-3a17-4c19-9a94-5866df9b6bf5",
			setup: func(t *testing.T) (*userService, testMocks) {
				ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
				ouServiceMock.On("IsOrganizationUnitExists",
					"4d8b40d6-3a17-4c19-9a94-5866df9b6bf5").Return(false, (*serviceerror.ServiceError)(nil)).Once()

				return &userService{
						ouService: ouServiceMock,
					}, testMocks{
						ouService: ouServiceMock,
					}
			},
			expectedErr: &ErrorOrganizationUnitNotFound,
		},
		{
			name:     "HandlesClientErrorWhenOrganizationUnitMissing",
			userType: testUserType,
			ouID:     "6c8f5afd-8884-4ea0-a317-3d8579346d86",
			setup: func(t *testing.T) (*userService, testMocks) {
				ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
				ouServiceMock.On("IsOrganizationUnitExists",
					"6c8f5afd-8884-4ea0-a317-3d8579346d86").Return(false, &serviceerror.ServiceError{
					Type: serviceerror.ClientErrorType,
					Code: oupkg.ErrorOrganizationUnitNotFound.Code,
				}).Once()

				return &userService{
						ouService: ouServiceMock,
					}, testMocks{
						ouService: ouServiceMock,
					}
			},
			expectedErr: &ErrorOrganizationUnitNotFound,
		},
		{
			name:     "HandlesClientErrorWhenOrganizationUnitIDInvalid",
			userType: testUserType,
			ouID:     "8d0c2f4e-8bb1-40bc-a0e1-ca5c4aacff63",
			setup: func(t *testing.T) (*userService, testMocks) {
				ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
				ouServiceMock.On("IsOrganizationUnitExists",
					"8d0c2f4e-8bb1-40bc-a0e1-ca5c4aacff63").Return(false, &serviceerror.ServiceError{
					Type: serviceerror.ClientErrorType,
					Code: oupkg.ErrorInvalidRequestFormat.Code,
				}).Once()

				return &userService{
						ouService: ouServiceMock,
					}, testMocks{
						ouService: ouServiceMock,
					}
			},
			expectedErr: &ErrorInvalidOrganizationUnitID,
		},
		{
			name:     "ReturnsMismatchWhenSchemaDoesNotMatchOU",
			userType: testUserType,
			ouID:     "f4e7c7b2-0b11-46a4-83be-4b43a7f69c7e",
			setup: func(t *testing.T) (*userService, testMocks) {
				parentOU := "a88cbecc-53a3-4c3e-958f-7ee4bf2d7a28"
				ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
				ouServiceMock.On("IsOrganizationUnitExists",
					"f4e7c7b2-0b11-46a4-83be-4b43a7f69c7e").Return(true, (*serviceerror.ServiceError)(nil)).Once()
				ouServiceMock.
					On("IsParent", parentOU, "f4e7c7b2-0b11-46a4-83be-4b43a7f69c7e").
					Return(false, (*serviceerror.ServiceError)(nil)).
					Once()

				userSchemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				userSchemaMock.
					On("GetUserSchemaByName", testUserType).
					Return(&userschema.UserSchema{
						OrganizationUnitID: parentOU,
					}, (*serviceerror.ServiceError)(nil)).
					Once()

				return &userService{
						ouService:         ouServiceMock,
						userSchemaService: userSchemaMock,
					}, testMocks{
						ouService:         ouServiceMock,
						userSchemaService: userSchemaMock,
					}
			},
			expectedErr: &ErrorOrganizationUnitMismatch,
		},
		{
			name:     "AllowsChildOrganizationUnit",
			userType: testUserType,
			ouID:     "1b5c7208-0d6f-4d5d-8fb9-6e8573549533",
			setup: func(t *testing.T) (*userService, testMocks) {
				parentOU := "c7e99c3b-e563-4c47-981f-1f7f755c8c68"
				ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
				ouServiceMock.On("IsOrganizationUnitExists",
					"1b5c7208-0d6f-4d5d-8fb9-6e8573549533").Return(true, (*serviceerror.ServiceError)(nil)).Once()
				ouServiceMock.On("IsParent", parentOU,
					"1b5c7208-0d6f-4d5d-8fb9-6e8573549533").Return(true, (*serviceerror.ServiceError)(nil)).Once()

				userSchemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				userSchemaMock.
					On("GetUserSchemaByName", testUserType).
					Return(&userschema.UserSchema{
						OrganizationUnitID: parentOU,
					}, (*serviceerror.ServiceError)(nil)).
					Once()

				return &userService{
						ouService:         ouServiceMock,
						userSchemaService: userSchemaMock,
					}, testMocks{
						ouService:         ouServiceMock,
						userSchemaService: userSchemaMock,
					}
			},
			expectedErr: nil,
		},
		{
			name:     "HandlesParentCheckErrorsOrganizationUnitNotFound",
			userType: testUserType,
			ouID:     "d9e12416-58d3-4c17-a4e4-cc4d96122598",
			setup: func(t *testing.T) (*userService, testMocks) {
				parentOU := "0a08d914-d223-48c2-8939-55d719739a17"
				ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
				ouServiceMock.On("IsOrganizationUnitExists",
					"d9e12416-58d3-4c17-a4e4-cc4d96122598").Return(true, (*serviceerror.ServiceError)(nil)).Once()
				ouServiceMock.On("IsParent", parentOU,
					"d9e12416-58d3-4c17-a4e4-cc4d96122598").Return(false, &serviceerror.ServiceError{
					Code: oupkg.ErrorOrganizationUnitNotFound.Code,
				}).Once()

				userSchemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				userSchemaMock.
					On("GetUserSchemaByName", testUserType).
					Return(&userschema.UserSchema{
						OrganizationUnitID: parentOU,
					}, (*serviceerror.ServiceError)(nil)).
					Once()

				return &userService{
						ouService:         ouServiceMock,
						userSchemaService: userSchemaMock,
					}, testMocks{
						ouService:         ouServiceMock,
						userSchemaService: userSchemaMock,
					}
			},
			expectedErr: &ErrorOrganizationUnitNotFound,
		},
		{
			name:     "HandlesParentCheckErrorsInternalServerError",
			userType: testUserType,
			ouID:     "d9e12416-58d3-4c17-a4e4-cc4d96122598",
			setup: func(t *testing.T) (*userService, testMocks) {
				parentOU := "0a08d914-d223-48c2-8939-55d719739a17"
				ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
				ouServiceMock.On("IsOrganizationUnitExists",
					"d9e12416-58d3-4c17-a4e4-cc4d96122598").Return(true, (*serviceerror.ServiceError)(nil)).Once()
				ouServiceMock.On("IsParent", parentOU,
					"d9e12416-58d3-4c17-a4e4-cc4d96122598").Return(false, &serviceerror.ServiceError{
					Code: oupkg.ErrorInternalServerError.Code,
				}).Once()

				userSchemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				userSchemaMock.
					On("GetUserSchemaByName", testUserType).
					Return(&userschema.UserSchema{
						OrganizationUnitID: parentOU,
					}, (*serviceerror.ServiceError)(nil)).
					Once()

				return &userService{
						ouService:         ouServiceMock,
						userSchemaService: userSchemaMock,
					}, testMocks{
						ouService:         ouServiceMock,
						userSchemaService: userSchemaMock,
					}
			},
			expectedErr: &ErrorInternalServerError,
		},
		{
			name:     "ReturnsNilWhenValid",
			userType: testUserType,
			ouID:     "e5c3aa8a-d7df-46f8-9f3f-bb3245c95d7c",
			setup: func(t *testing.T) (*userService, testMocks) {
				ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
				ouServiceMock.On("IsOrganizationUnitExists",
					"e5c3aa8a-d7df-46f8-9f3f-bb3245c95d7c").Return(true, (*serviceerror.ServiceError)(nil)).Once()

				userSchemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				userSchemaMock.
					On("GetUserSchemaByName", testUserType).
					Return(&userschema.UserSchema{
						OrganizationUnitID: "e5c3aa8a-d7df-46f8-9f3f-bb3245c95d7c",
					}, (*serviceerror.ServiceError)(nil)).
					Once()

				return &userService{
						ouService:         ouServiceMock,
						userSchemaService: userSchemaMock,
					}, testMocks{
						ouService:         ouServiceMock,
						userSchemaService: userSchemaMock,
					}
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			service, _ := tc.setup(t)
			logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserServiceTest"))

			err := service.validateOrganizationUnitForUserType(tc.userType, tc.ouID, logger)
			if tc.expectedErr == nil {
				require.Nil(t, err)
				return
			}

			require.NotNil(t, err)
			require.Equal(t, *tc.expectedErr, *err)
		})
	}
}

func TestUserService_GetUsersByPath_HandlesOUServiceErrors(t *testing.T) {
	testCases := []struct {
		name        string
		setup       func(t *testing.T) *userService
		expectedErr *serviceerror.ServiceError
	}{
		{
			name: "ReturnsInvalidHandlePathWhenResolverFails",
			setup: func(t *testing.T) *userService {
				ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
				ouServiceMock.
					On("GetOrganizationUnitByPath", "root").
					Return(oupkg.OrganizationUnit{}, &serviceerror.ServiceError{
						Type: serviceerror.ClientErrorType,
						Code: oupkg.ErrorInvalidHandlePath.Code,
					}).
					Once()

				return &userService{
					ouService: ouServiceMock,
				}
			},
			expectedErr: &ErrorInvalidHandlePath,
		},
		{
			name: "ReturnsInvalidLimitWhenListingUsersFails",
			setup: func(t *testing.T) *userService {
				ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
				ouServiceMock.
					On("GetOrganizationUnitByPath", "root").
					Return(oupkg.OrganizationUnit{ID: "ou-id"}, (*serviceerror.ServiceError)(nil)).
					Once()
				ouServiceMock.
					On("GetOrganizationUnitUsers", "ou-id", 10, 0).
					Return((*oupkg.UserListResponse)(nil), &serviceerror.ServiceError{
						Type: serviceerror.ClientErrorType,
						Code: oupkg.ErrorInvalidLimit.Code,
					}).
					Once()

				return &userService{
					ouService: ouServiceMock,
				}
			},
			expectedErr: &ErrorInvalidLimit,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			service := tc.setup(t)

			resp, err := service.GetUsersByPath(context.Background(), "root", 10, 0, nil)
			require.Nil(t, resp)
			require.NotNil(t, err)
			require.Equal(t, *tc.expectedErr, *err)
		})
	}
}

func TestUserService_CreateUserByPath_HandlesOUServiceErrors(t *testing.T) {
	ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
	ouServiceMock.
		On("GetOrganizationUnitByPath", "root/engineering").
		Return(oupkg.OrganizationUnit{}, &serviceerror.ServiceError{
			Type: serviceerror.ClientErrorType,
			Code: oupkg.ErrorInvalidHandlePath.Code,
		}).
		Once()

	service := &userService{
		ouService: ouServiceMock,
	}

	resp, err := service.CreateUserByPath(context.Background(), "root/engineering", CreateUserByPathRequest{
		Type: testUserType,
	})
	require.Nil(t, resp)
	require.NotNil(t, err)
	require.Equal(t, ErrorInvalidHandlePath, *err)
}

func TestUserService_ContainsCredentialFields(t *testing.T) {
	service := &userService{}

	t.Run("ReturnsFalseWhenAttributesEmpty", func(t *testing.T) {
		hasCreds, err := service.containsCredentialFields(json.RawMessage{})
		require.False(t, hasCreds)
		require.Nil(t, err)
	})

	t.Run("ReturnsErrorForInvalidJSON", func(t *testing.T) {
		hasCreds, err := service.containsCredentialFields(json.RawMessage(`{"password":`))
		require.False(t, hasCreds)
		require.NotNil(t, err)
		require.Equal(t, ErrorInvalidRequestFormat, *err)
	})

	t.Run("ReturnsFalseWhenNoCredentialFields", func(t *testing.T) {
		hasCreds, err := service.containsCredentialFields(json.RawMessage(`{"email":"a@b.com"}`))
		require.False(t, hasCreds)
		require.Nil(t, err)
	})

	t.Run("ReturnsTrueForSupportedCredentialFields", func(t *testing.T) {
		for _, field := range []string{"password", "pin", "secret"} {
			payload, marshalErr := json.Marshal(map[string]string{
				"email": "user@example.com",
				field:   "value",
			})
			require.NoError(t, marshalErr)

			hasCreds, err := service.containsCredentialFields(payload)
			require.Nil(t, err)
			require.True(t, hasCreds, "expected field %s to be detected", field)
		}
	})
}

func TestUserService_UpdateUserCredentials_Validation(t *testing.T) {
	t.Run("ReturnsAuthErrorWhenUserIDMissing", func(t *testing.T) {
		service := &userService{}

		err := service.UpdateUserCredentials(context.Background(), "", json.RawMessage(`{"password":"pw"}`))
		require.NotNil(t, err)
		require.Equal(t, ErrorAuthenticationFailed, *err)
	})

	t.Run("ReturnsMissingCredentialsWhenPayloadEmpty", func(t *testing.T) {
		service := &userService{}

		err := service.UpdateUserCredentials(context.Background(), "user-1", json.RawMessage{})
		require.NotNil(t, err)
		require.Equal(t, ErrorMissingCredentials, *err)
	})
}

func TestUserService_UpdateUserCredentials_UserNotFound(t *testing.T) {
	userStoreMock := newUserStoreInterfaceMock(t)
	userStoreMock.
		On("GetCredentials", mock.Anything, "user-1").
		Return(User{}, []Credential{}, ErrUserNotFound).
		Once()
	hashServiceMock := hashmock.NewHashServiceInterfaceMock(t)
	hashServiceMock.
		On("Generate", mock.Anything).
		Return(hash.Credential{
			Algorithm:  hash.PBKDF2,
			Hash:       "hashed-value",
			Parameters: hash.CredParameters{Salt: "random-salt"},
		}, nil).
		Once()

	service := &userService{
		userStore:   userStoreMock,
		hashService: hashServiceMock,
	}

	config.ResetThunderRuntime()
	initErr := config.InitializeThunderRuntime("", &config.Config{
		Crypto: config.CryptoConfig{
			PasswordHashing: config.PasswordHashingConfig{
				Algorithm: string(hash.PBKDF2),
			},
		},
	})
	require.NoError(t, initErr)
	t.Cleanup(config.ResetThunderRuntime)

	svcErr := service.UpdateUserCredentials(context.Background(), "user-1", json.RawMessage(`{"password":"pw"}`))
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorUserNotFound, *svcErr)
	userStoreMock.AssertNotCalled(t, "UpdateUserCredentials", mock.Anything, mock.Anything)
}

func TestUserService_UpdateUserCredentials_Succeeds(t *testing.T) {
	userStoreMock := newUserStoreInterfaceMock(t)
	existingCredentials := []Credential{
		{
			CredentialType: "password",
			StorageType:    "hash",
			StorageAlgo:    hash.SHA256,
			Value:          "old-hash",
			StorageAlgoParams: hash.CredParameters{
				Salt: "old-salt",
			},
		},
		{
			CredentialType: "pin",
			StorageType:    "hash",
			StorageAlgo:    hash.SHA256,
			Value:          "pin-hash",
			StorageAlgoParams: hash.CredParameters{
				Salt: "pin-salt",
			},
		},
	}
	userStoreMock.
		On("GetCredentials", mock.Anything, "user-1").
		Return(User{ID: "user-1"}, existingCredentials, nil).
		Once()
	var captured []Credential
	userStoreMock.
		On("UpdateUserCredentials", mock.Anything, "user-1", mock.Anything).
		Run(func(args mock.Arguments) {
			if creds, ok := args[2].([]Credential); ok {
				captured = creds
			}
		}).
		Return(nil).
		Once()

	hashServiceMock := hashmock.NewHashServiceInterfaceMock(t)
	hashServiceMock.
		On("Generate", mock.Anything).
		Return(hash.Credential{
			Algorithm:  hash.PBKDF2,
			Hash:       "hash",
			Parameters: hash.CredParameters{Salt: "random-salt"},
		}, nil).
		Once()

	service := &userService{
		userStore:   userStoreMock,
		hashService: hashServiceMock,
	}

	config.ResetThunderRuntime()
	initErr := config.InitializeThunderRuntime("", &config.Config{
		Crypto: config.CryptoConfig{
			PasswordHashing: config.PasswordHashingConfig{
				Algorithm: string(hash.PBKDF2),
			},
		},
	})
	require.NoError(t, initErr)
	t.Cleanup(config.ResetThunderRuntime)

	svcErr := service.UpdateUserCredentials(context.Background(), "user-1", json.RawMessage(`{"password":"newPass"}`))
	require.Nil(t, svcErr)
	require.Len(t, captured, 2)

	var passwordCred, pinCred *Credential
	for i := range captured {
		switch captured[i].CredentialType {
		case "password":
			passwordCred = &captured[i]
		case "pin":
			pinCred = &captured[i]
		}
	}

	require.NotNil(t, passwordCred)
	require.NotNil(t, pinCred)
	require.Equal(t, "hash", passwordCred.StorageType)
	require.NotEmpty(t, passwordCred.Value)
	require.NotEmpty(t, passwordCred.StorageAlgoParams.Salt)
	require.NotEqual(t, "old-hash", passwordCred.Value)
	require.Equal(t, "pin-hash", pinCred.Value)
	require.Equal(t, "pin-salt", pinCred.StorageAlgoParams.Salt)
}

func TestUserService_MergeCredentials(t *testing.T) {
	service := &userService{}

	type testCase struct {
		name     string
		existing []Credential
		provided []Credential
		expected []Credential
	}

	tests := []testCase{
		{
			name: "ReplacesMatchingAndPreservesExistingOrder",
			existing: []Credential{
				{
					CredentialType: "password", StorageType: "hash", Value: "old-pass",
					StorageAlgoParams: hash.CredParameters{Salt: "salt-1"},
				},
				{
					CredentialType: "pin", StorageType: "hash", Value: "old-pin",
					StorageAlgoParams: hash.CredParameters{Salt: "salt-2"},
				},
			},
			provided: []Credential{
				{
					CredentialType: "password", StorageType: "hash", Value: "new-pass",
					StorageAlgoParams: hash.CredParameters{Salt: "salt-3"},
				},
				{
					CredentialType: "secret", StorageType: "hash", Value: "secret",
					StorageAlgoParams: hash.CredParameters{Salt: "salt-4"},
				},
			},
			expected: []Credential{
				{
					CredentialType: "password", StorageType: "hash", Value: "new-pass",
					StorageAlgoParams: hash.CredParameters{Salt: "salt-3"},
				},
				{
					CredentialType: "pin", StorageType: "hash", Value: "old-pin",
					StorageAlgoParams: hash.CredParameters{Salt: "salt-2"},
				},
				{
					CredentialType: "secret", StorageType: "hash", Value: "secret",
					StorageAlgoParams: hash.CredParameters{Salt: "salt-4"},
				},
			},
		},
		{
			name: "ProvidedDuplicatesKeepLastValue",
			existing: []Credential{
				{
					CredentialType: "pin", StorageType: "hash", Value: "existing-pin",
					StorageAlgoParams: hash.CredParameters{Salt: "salt-1"},
				},
			},
			provided: []Credential{
				{
					CredentialType: "password", StorageType: "hash", Value: "first-pass",
					StorageAlgoParams: hash.CredParameters{Salt: "salt-a"},
				},
				{
					CredentialType: "password", StorageType: "hash", Value: "second-pass",
					StorageAlgoParams: hash.CredParameters{Salt: "salt-b"},
				},
			},
			expected: []Credential{
				{
					CredentialType: "pin", StorageType: "hash", Value: "existing-pin",
					StorageAlgoParams: hash.CredParameters{Salt: "salt-1"},
				},
				{
					CredentialType: "password", StorageType: "hash", Value: "second-pass",
					StorageAlgoParams: hash.CredParameters{Salt: "salt-b"},
				},
			},
		},
		{
			name: "NoProvidedCredentialsReturnsExisting",
			existing: []Credential{
				{
					CredentialType: "password", StorageType: "hash", Value: "only",
					StorageAlgoParams: hash.CredParameters{Salt: "salt-1"},
				},
			},
			provided: []Credential{},
			expected: []Credential{
				{
					CredentialType: "password", StorageType: "hash", Value: "only",
					StorageAlgoParams: hash.CredParameters{Salt: "salt-1"},
				},
			},
		},
		{
			name:     "NoExistingCredentialsReturnsProvidedInOrder",
			existing: []Credential{},
			provided: []Credential{
				{CredentialType: "password", StorageType: "hash", Value: "first",
					StorageAlgoParams: hash.CredParameters{Salt: "salt-1"}},
				{CredentialType: "password", StorageType: "hash", Value: "second",
					StorageAlgoParams: hash.CredParameters{Salt: "salt-2"}},
				{CredentialType: "pin", StorageType: "hash", Value: "pin-val",
					StorageAlgoParams: hash.CredParameters{Salt: "salt-3"}},
			},
			expected: []Credential{
				{CredentialType: "password", StorageType: "hash", Value: "second",
					StorageAlgoParams: hash.CredParameters{Salt: "salt-2"}},
				{CredentialType: "pin", StorageType: "hash", Value: "pin-val",
					StorageAlgoParams: hash.CredParameters{Salt: "salt-3"}},
			},
		},
		{
			name: "ReplacesMultipleExistingTypesAndAppendsNew",
			existing: []Credential{
				{CredentialType: "password", StorageType: "hash", Value: "old-pass",
					StorageAlgoParams: hash.CredParameters{Salt: "salt-1"}},
				{CredentialType: "pin", StorageType: "hash", Value: "old-pin",
					StorageAlgoParams: hash.CredParameters{Salt: "salt-2"}},
			},
			provided: []Credential{
				{CredentialType: "pin", StorageType: "hash", Value: "new-pin",
					StorageAlgoParams: hash.CredParameters{Salt: "salt-3"}},
				{CredentialType: "password", StorageType: "hash", Value: "new-pass",
					StorageAlgoParams: hash.CredParameters{Salt: "salt-4"}},
				{CredentialType: "secret", StorageType: "hash", Value: "new-secret",
					StorageAlgoParams: hash.CredParameters{Salt: "salt-5"}},
			},
			expected: []Credential{
				{CredentialType: "password", StorageType: "hash", Value: "new-pass",
					StorageAlgoParams: hash.CredParameters{Salt: "salt-4"}},
				{CredentialType: "pin", StorageType: "hash", Value: "new-pin",
					StorageAlgoParams: hash.CredParameters{Salt: "salt-3"}},
				{CredentialType: "secret", StorageType: "hash", Value: "new-secret",
					StorageAlgoParams: hash.CredParameters{Salt: "salt-5"}},
			},
		},
		{
			name:     "ReturnsEmptyWhenNoCredentials",
			existing: []Credential{},
			provided: []Credential{},
			expected: []Credential{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			merged := service.mergeCredentials(tc.existing, tc.provided)
			require.Equal(t, tc.expected, merged)
		})
	}
}

func TestUserService_UpdateUserAttributes_Validation(t *testing.T) {
	service := &userService{}

	resp, err := service.UpdateUserAttributes(context.Background(), "", json.RawMessage(`{"email":"a@b.com"}`))
	require.Nil(t, resp)
	require.NotNil(t, err)
	require.Equal(t, ErrorMissingUserID, *err)

	resp, err = service.UpdateUserAttributes(context.Background(), "user-1", json.RawMessage{})
	require.Nil(t, resp)
	require.NotNil(t, err)
	require.Equal(t, ErrorInvalidRequestFormat, *err)

	resp, err = service.UpdateUserAttributes(context.Background(), "user-1", json.RawMessage(`{"password":"Secret123"}`))
	require.Nil(t, resp)
	require.NotNil(t, err)
	require.Equal(t, ErrorInvalidRequestFormat, *err)
}

func TestUserService_UpdateUserAttributes_UserNotFound(t *testing.T) {
	storeMock := newUserStoreInterfaceMock(t)
	storeMock.On("GetUser", mock.Anything, "user-1").Return(User{}, ErrUserNotFound).Once()

	service := &userService{
		userStore: storeMock,
	}

	resp, err := service.UpdateUserAttributes(context.Background(), "user-1", json.RawMessage(`{"email":"a@b.com"}`))
	require.Nil(t, resp)
	require.NotNil(t, err)
	require.Equal(t, ErrorUserNotFound, *err)
	storeMock.AssertNotCalled(t, "UpdateUser", mock.Anything)
}

func TestUserService_UpdateUserAttributes_SchemaValidationFails(t *testing.T) {
	storeMock := newUserStoreInterfaceMock(t)
	storeMock.
		On("GetUser", mock.Anything, "user-1").
		Return(User{ID: "user-1", Type: testUserType, Attributes: json.RawMessage(`{"email":"old"}`)}, nil).
		Once()

	schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
	schemaMock.
		On("ValidateUser", testUserType, mock.Anything).
		Return(false, &userschema.ErrorUserSchemaNotFound).
		Once()

	service := &userService{
		userStore:         storeMock,
		userSchemaService: schemaMock,
	}

	resp, err := service.UpdateUserAttributes(context.Background(), "user-1", json.RawMessage(`{"email":"new@example.com"}`))
	require.Nil(t, resp)
	require.NotNil(t, err)
	require.Equal(t, ErrorUserSchemaNotFound, *err)
	storeMock.AssertNotCalled(t, "UpdateUser", mock.Anything)
}

func TestUserService_UpdateUserAttributes_Succeeds(t *testing.T) {
	storeMock := newUserStoreInterfaceMock(t)
	storeMock.
		On("GetUser", mock.Anything, "user-1").
		Return(User{ID: "user-1", Type: testUserType, Attributes: json.RawMessage(`{"email":"old@example.com"}`)}, nil).
		Once()

	schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
	schemaMock.
		On("ValidateUser", testUserType, mock.Anything).
		Return(true, (*serviceerror.ServiceError)(nil)).
		Once()
	schemaMock.
		On("ValidateUserUniqueness", testUserType, mock.Anything, mock.Anything).
		Return(true, (*serviceerror.ServiceError)(nil)).
		Once()

	var savedUser *User
	storeMock.
		On("UpdateUser", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			if u, ok := args[1].(*User); ok {
				savedUser = u
			}
		}).
		Return(nil).
		Once()

	service := &userService{
		userStore:         storeMock,
		userSchemaService: schemaMock,
	}

	newAttrs := json.RawMessage(`{"email":"new@example.com"}`)
	resp, err := service.UpdateUserAttributes(context.Background(), "user-1", newAttrs)
	require.Nil(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "user-1", resp.ID)
	require.JSONEq(t, string(newAttrs), string(resp.Attributes))

	require.NotNil(t, savedUser)
	require.Equal(t, "user-1", savedUser.ID)
	require.JSONEq(t, string(newAttrs), string(savedUser.Attributes))
}
