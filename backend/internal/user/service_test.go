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
	dbmodel "github.com/asgardeo/thunder/internal/system/database/model"
	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/database/transaction"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/userschema"
	"github.com/asgardeo/thunder/tests/mocks/crypto/hashmock"
	"github.com/asgardeo/thunder/tests/mocks/oumock"
	"github.com/asgardeo/thunder/tests/mocks/userschemamock"
)

const (
	svcTestUserID1   = "user-1"
	svcTestUserID123 = "user-123"
	testUserType     = "employee"
)
const testOrgID = "11111111-1111-1111-1111-111111111111"

func TestOUStore_ValidateUserAndUniqueness(t *testing.T) {
	type testMocks struct {
		schemaService *userschemamock.UserSchemaServiceInterfaceMock
		userStore     *userStoreInterfaceMock
	}

	payloadWithEmail := []byte(`{"email":"employee@example.com"}`)
	emptyPayload := []byte(`{}`)

	testCases := []struct {
		name          string
		payload       []byte
		excludeUserID string
		setup         func(t *testing.T) (*userService, testMocks)
		assert        func(t *testing.T, err *serviceerror.ServiceError, mocks testMocks)
	}{
		{
			name:    "ReturnsInternalErrorWhenSchemaValidationFails",
			payload: payloadWithEmail,
			setup: func(t *testing.T) (*userService, testMocks) {
				schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				schemaMock.
					On("ValidateUser", mock.Anything, testUserType, mock.Anything).
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
					On("ValidateUser", mock.Anything, testUserType, mock.Anything).
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
					On("ValidateUser", mock.Anything, testUserType, mock.Anything).
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
					On("ValidateUser", mock.Anything, testUserType, mock.Anything).
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
					On("ValidateUser", mock.Anything, testUserType, mock.Anything).
					Return(true, nil).
					Once()
				schemaMock.
					On("ValidateUserUniqueness", mock.Anything, testUserType, mock.Anything, mock.Anything).
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
					On("ValidateUser", mock.Anything, testUserType, mock.Anything).
					Return(true, nil).
					Once()
				schemaMock.
					On("ValidateUserUniqueness", mock.Anything, testUserType, mock.Anything, mock.Anything).
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
				existingUserID := svcTestUserID123
				schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				userStoreMock := newUserStoreInterfaceMock(t)
				userStoreMock.
					On("IdentifyUser", mock.Anything, mock.AnythingOfType("map[string]interface {}")).
					Return(&existingUserID, nil).
					Once()
				schemaMock.
					On("ValidateUser", mock.Anything, testUserType, mock.Anything).
					Return(true, nil).
					Once()
				schemaMock.
					On("ValidateUserUniqueness", mock.Anything, testUserType, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						identify := args.Get(3).(func(map[string]interface{}) (*string, error))

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
					On("ValidateUser", mock.Anything, testUserType, mock.Anything).
					Return(true, nil).
					Once()
				schemaMock.
					On("ValidateUserUniqueness", mock.Anything, testUserType, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						identify := args.Get(3).(func(map[string]interface{}) (*string, error))

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
					On("ValidateUser", mock.Anything, testUserType, mock.Anything).
					Return(true, nil).
					Once()
				schemaMock.
					On("ValidateUserUniqueness", mock.Anything, testUserType, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						identify := args.Get(3).(func(map[string]interface{}) (*string, error))

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
		{
			name:          "ReturnsNilWhenConflictIsWithSameUser",
			payload:       payloadWithEmail,
			excludeUserID: svcTestUserID123,
			setup: func(t *testing.T) (*userService, testMocks) {
				existingUserID := svcTestUserID123
				schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
				userStoreMock := newUserStoreInterfaceMock(t)
				userStoreMock.
					On("IdentifyUser", mock.Anything, mock.AnythingOfType("map[string]interface {}")).
					Return(&existingUserID, nil).
					Once()
				schemaMock.
					On("ValidateUser", mock.Anything, testUserType, mock.Anything).
					Return(true, nil).
					Once()
				schemaMock.
					On("ValidateUserUniqueness", mock.Anything, testUserType, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						identify := args.Get(3).(func(map[string]interface{}) (*string, error))

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
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			service, mocks := tc.setup(t)
			logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserServiceTest"))

			err := service.validateUserAndUniqueness(context.Background(), testUserType, tc.payload, logger,
				tc.excludeUserID)
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
					On("GetUserSchemaByName", mock.Anything, testUserType).
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
					On("GetUserSchemaByName", mock.Anything, testUserType).
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
					On("GetUserSchemaByName", mock.Anything, testUserType).
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
					On("GetUserSchemaByName", mock.Anything, testUserType).
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
					On("GetUserSchemaByName", mock.Anything, testUserType).
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

			err := service.validateOrganizationUnitForUserType(context.TODO(), tc.userType, tc.ouID, logger)
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

func TestUserService_CreateUser_UsesTransactionAndStore(t *testing.T) {
	ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
	ouServiceMock.On("IsOrganizationUnitExists", testOrgID).Return(true, (*serviceerror.ServiceError)(nil)).Once()

	userSchemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
	userSchemaMock.On("GetUserSchemaByName", mock.Anything, testUserType).
		Return(&userschema.UserSchema{OrganizationUnitID: testOrgID}, (*serviceerror.ServiceError)(nil)).
		Once()
	userSchemaMock.On("ValidateUser", mock.Anything, testUserType, mock.Anything).
		Return(true, (*serviceerror.ServiceError)(nil)).
		Once()
	userSchemaMock.On("ValidateUserUniqueness", mock.Anything, testUserType, mock.Anything, mock.Anything).
		Return(true, (*serviceerror.ServiceError)(nil)).
		Once()
	userSchemaMock.On("GetCredentialAttributes", mock.Anything, testUserType).
		Return([]string{"password"}, (*serviceerror.ServiceError)(nil)).
		Once()

	storeMock := newUserStoreInterfaceMock(t)
	var capturedCtx context.Context
	storeMock.
		On("CreateUser", mock.Anything, mock.MatchedBy(func(u User) bool {
			return u.OrganizationUnit == testOrgID && u.Type == testUserType && u.ID != ""
		}), mock.Anything).
		Run(func(args mock.Arguments) {
			capturedCtx = args[0].(context.Context)
		}).
		Return(nil).
		Once()

	txMock := &fakeTransactioner{}

	service := &userService{
		userStore:         storeMock,
		ouService:         ouServiceMock,
		userSchemaService: userSchemaMock,
		hashService:       hashmock.NewHashServiceInterfaceMock(t),
		transactioner:     txMock,
	}

	user := &User{
		Type:             testUserType,
		OrganizationUnit: testOrgID,
		Attributes:       json.RawMessage(`{}`),
	}

	created, err := service.CreateUser(context.Background(), user)
	require.Nil(t, err)
	require.NotNil(t, created)
	require.Equal(t, testOrgID, created.OrganizationUnit)
	require.NotEmpty(t, created.ID)
	require.Equal(t, 1, txMock.transactCalls)
	require.NotNil(t, capturedCtx)
}

func TestUserService_CreateUser_PropagatesStoreError(t *testing.T) {
	storeErr := errors.New("store failure")

	ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
	ouServiceMock.On("IsOrganizationUnitExists", testOrgID).Return(true, (*serviceerror.ServiceError)(nil)).Once()

	userSchemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
	userSchemaMock.On("GetUserSchemaByName", mock.Anything, testUserType).
		Return(&userschema.UserSchema{OrganizationUnitID: testOrgID}, (*serviceerror.ServiceError)(nil)).
		Once()
	userSchemaMock.On("ValidateUser", mock.Anything, testUserType, mock.Anything).
		Return(true, (*serviceerror.ServiceError)(nil)).
		Once()
	userSchemaMock.On("ValidateUserUniqueness", mock.Anything, testUserType, mock.Anything, mock.Anything).
		Return(true, (*serviceerror.ServiceError)(nil)).
		Once()
	userSchemaMock.On("GetCredentialAttributes", mock.Anything, testUserType).
		Return([]string{"password"}, (*serviceerror.ServiceError)(nil)).
		Once()

	storeMock := newUserStoreInterfaceMock(t)
	storeMock.
		On("CreateUser", mock.Anything, mock.Anything, mock.Anything).
		Return(storeErr).
		Once()

	txMock := &fakeTransactioner{}

	service := &userService{
		userStore:         storeMock,
		ouService:         ouServiceMock,
		userSchemaService: userSchemaMock,
		hashService:       hashmock.NewHashServiceInterfaceMock(t),
		transactioner:     txMock,
	}

	user := &User{
		Type:             testUserType,
		OrganizationUnit: testOrgID,
		Attributes:       json.RawMessage(`{}`),
	}

	created, svcErr := service.CreateUser(context.Background(), user)
	require.Nil(t, created)
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorInternalServerError, *svcErr)
	require.Equal(t, 1, txMock.transactCalls)
}

func TestUserService_CreateUser_TransactionerError(t *testing.T) {
	ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
	ouServiceMock.On("IsOrganizationUnitExists", testOrgID).Return(true, (*serviceerror.ServiceError)(nil)).Once()

	userSchemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
	userSchemaMock.On("GetUserSchemaByName", mock.Anything, testUserType).
		Return(&userschema.UserSchema{OrganizationUnitID: testOrgID}, (*serviceerror.ServiceError)(nil)).
		Once()
	userSchemaMock.On("ValidateUser", mock.Anything, testUserType, mock.Anything).
		Return(true, (*serviceerror.ServiceError)(nil)).
		Once()
	userSchemaMock.On("ValidateUserUniqueness", mock.Anything, testUserType, mock.Anything, mock.Anything).
		Return(true, (*serviceerror.ServiceError)(nil)).
		Once()
	userSchemaMock.On("GetCredentialAttributes", mock.Anything, testUserType).
		Return([]string{"password"}, (*serviceerror.ServiceError)(nil)).
		Once()

	storeMock := newUserStoreInterfaceMock(t)
	storeMock.AssertNotCalled(t, "CreateUser", mock.Anything, mock.Anything, mock.Anything)

	txMock := &fakeTransactioner{err: errors.New("tx failed")}

	service := &userService{
		userStore:         storeMock,
		ouService:         ouServiceMock,
		userSchemaService: userSchemaMock,
		hashService:       hashmock.NewHashServiceInterfaceMock(t),
		transactioner:     txMock,
	}

	user := &User{
		Type:             testUserType,
		OrganizationUnit: testOrgID,
		Attributes:       json.RawMessage(`{}`),
	}

	created, svcErr := service.CreateUser(context.Background(), user)
	require.Nil(t, created)
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorInternalServerError, *svcErr)
	require.Equal(t, 1, txMock.transactCalls)
	storeMock.AssertNotCalled(t, "CreateUser", mock.Anything, mock.Anything, mock.Anything)
}

func TestUserService_ContainsCredentialAttributes(t *testing.T) {
	service := &userService{}
	schemaCredFields := []string{"password", "pin", "secret"}

	t.Run("ReturnsFalseWhenAttributesEmpty", func(t *testing.T) {
		hasCreds, err := service.containsCredentialAttributes(json.RawMessage{}, schemaCredFields)
		require.False(t, hasCreds)
		require.Nil(t, err)
	})

	t.Run("ReturnsErrorForInvalidJSON", func(t *testing.T) {
		hasCreds, err := service.containsCredentialAttributes(json.RawMessage(`{"password":`), schemaCredFields)
		require.False(t, hasCreds)
		require.NotNil(t, err)
		require.Equal(t, ErrorInvalidRequestFormat, *err)
	})

	t.Run("ReturnsFalseWhenNoCredentialAttributes", func(t *testing.T) {
		hasCreds, err := service.containsCredentialAttributes(json.RawMessage(`{"email":"a@b.com"}`), schemaCredFields)
		require.False(t, hasCreds)
		require.Nil(t, err)
	})

	t.Run("ReturnsTrueForSchemaCredentialAttributes", func(t *testing.T) {
		for _, field := range []string{"password", "pin", "secret"} {
			payload, marshalErr := json.Marshal(map[string]string{
				"email": "user@example.com",
				field:   "value",
			})
			require.NoError(t, marshalErr)

			hasCreds, err := service.containsCredentialAttributes(payload, schemaCredFields)
			require.Nil(t, err)
			require.True(t, hasCreds, "expected field %s to be detected", field)
		}
	})

	t.Run("ReturnsTrueForSystemManagedCredentialAttributes", func(t *testing.T) {
		payload, marshalErr := json.Marshal(map[string]string{
			"email":   "user@example.com",
			"passkey": "value",
		})
		require.NoError(t, marshalErr)

		hasCreds, err := service.containsCredentialAttributes(payload, nil)
		require.Nil(t, err)
		require.True(t, hasCreds, "expected passkey to be detected as system-managed credential")
	})
}

func TestUserStore_SyncIndexedAttributes_BuildsBatchInsert(t *testing.T) {
	client := &fakeDBClient{}
	us := &userStore{
		deploymentID:      "dep",
		indexedAttributes: map[string]bool{"email": true, "nickname": false, "profile": true},
	}

	attrs := json.RawMessage(`{"email":"a@b.com","nickname":"nick","profile":{"city":"ny"}}`)
	err := us.syncIndexedAttributes(context.Background(), client, svcTestUserID1, attrs)
	require.NoError(t, err)
	require.True(t, client.called)
	require.Equal(t, QueryBatchInsertIndexedAttributes.ID, client.query.ID)
	require.Equal(t, 4, len(client.args))
	require.Equal(t, svcTestUserID1, client.args[0])
	require.Equal(t, "email", client.args[1])
	require.Equal(t, "a@b.com", client.args[2])
	require.Equal(t, "dep", client.args[3])
}

func TestUserStore_SyncIndexedAttributes_NoIndexedAttributes(t *testing.T) {
	client := &fakeDBClient{}
	us := &userStore{
		deploymentID:      "dep",
		indexedAttributes: map[string]bool{},
	}

	attrs := json.RawMessage(`{"nickname":"nick"}`)
	err := us.syncIndexedAttributes(context.Background(), client, svcTestUserID1, attrs)
	require.NoError(t, err)
	require.False(t, client.called)
}

func TestUserStore_SyncIndexedAttributes_ExecuteError(t *testing.T) {
	client := &fakeDBClient{retErr: errors.New("db error")}
	us := &userStore{
		deploymentID:      "dep",
		indexedAttributes: map[string]bool{"email": true},
	}

	attrs := json.RawMessage(`{"email":"a@b.com"}`)
	err := us.syncIndexedAttributes(context.Background(), client, svcTestUserID1, attrs)
	require.Error(t, err)
	require.Contains(t, err.Error(), QueryBatchInsertIndexedAttributes.ID)
	require.True(t, client.called)
}

// fakeTransactioner is a light-weight test double to capture transaction usage without sql mock plumbing.
type fakeTransactioner struct {
	transactCalls int
	err           error
}

func (f *fakeTransactioner) Transact(ctx context.Context, txFunc func(context.Context) error) error {
	f.transactCalls++
	if f.err != nil {
		return f.err
	}
	return txFunc(ctx)
}

// fakeDBClient captures ExecuteContext calls for syncIndexedAttributes.
type fakeDBClient struct {
	called bool
	query  dbmodel.DBQuery
	args   []interface{}
	retErr error
}

func (f *fakeDBClient) Query(dbmodel.DBQuery, ...interface{}) ([]map[string]interface{}, error) {
	return nil, nil
}

func (f *fakeDBClient) QueryContext(
	context.Context, dbmodel.DBQuery, ...interface{},
) ([]map[string]interface{}, error) {
	return nil, nil
}

func (f *fakeDBClient) Execute(dbmodel.DBQuery, ...interface{}) (int64, error) {
	return 0, nil
}

func (f *fakeDBClient) ExecuteContext(_ context.Context, q dbmodel.DBQuery, args ...interface{}) (int64, error) {
	f.called = true
	f.query = q
	f.args = args
	return 1, f.retErr
}

func (f *fakeDBClient) BeginTx() (dbmodel.TxInterface, error) {
	return nil, nil
}

func (f *fakeDBClient) GetTransactioner() (transaction.Transactioner, error) {
	return nil, nil
}

var _ provider.DBClientInterface = (*fakeDBClient)(nil)

func TestUserService_UpdateUserCredentials_Validation(t *testing.T) {
	t.Run("ReturnsAuthErrorWhenUserIDMissing", func(t *testing.T) {
		service := &userService{
			transactioner: &fakeTransactioner{},
		}

		err := service.UpdateUserCredentials(context.Background(), "", json.RawMessage(`{"password":"newpass"}`))
		require.NotNil(t, err)
		require.Equal(t, ErrorAuthenticationFailed, *err)
	})

	t.Run("ReturnsMissingCredentialsWhenPayloadEmpty", func(t *testing.T) {
		service := &userService{
			transactioner: &fakeTransactioner{},
		}

		err := service.UpdateUserCredentials(context.Background(), svcTestUserID1, json.RawMessage(``))
		require.NotNil(t, err)
		require.Equal(t, ErrorMissingCredentials, *err)
	})

	t.Run("ReturnsInvalidRequestFormatWhenInvalidJSON", func(t *testing.T) {
		service := &userService{
			transactioner: &fakeTransactioner{},
		}

		err := service.UpdateUserCredentials(context.Background(), svcTestUserID1, json.RawMessage(`invalid json`))
		require.NotNil(t, err)
		require.Equal(t, ErrorInvalidRequestFormat, *err)
	})

	t.Run("ReturnsInvalidCredentialForUnsupportedType", func(t *testing.T) {
		userStoreMock := newUserStoreInterfaceMock(t)
		userStoreMock.
			On("GetCredentials", mock.Anything, svcTestUserID1).
			Return(User{ID: svcTestUserID1, Type: "Person"}, Credentials{}, nil).
			Once()

		schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
		schemaMock.On("GetCredentialAttributes", mock.Anything, "Person").
			Return([]string{"password"}, (*serviceerror.ServiceError)(nil)).
			Once()

		service := &userService{
			userStore:         userStoreMock,
			userSchemaService: schemaMock,
			transactioner:     &fakeTransactioner{},
		}

		err := service.UpdateUserCredentials(context.Background(), svcTestUserID1,
			json.RawMessage(`{"invalidtype":"value"}`))
		require.NotNil(t, err)
		require.Equal(t, ErrorInvalidCredential.Code, err.Code)
	})

	t.Run("ReturnsMissingCredentialsWhenMapEmpty", func(t *testing.T) {
		service := &userService{
			transactioner: &fakeTransactioner{},
		}

		err := service.UpdateUserCredentials(context.Background(), svcTestUserID1, json.RawMessage(`{}`))
		require.NotNil(t, err)
		require.Equal(t, ErrorMissingCredentials, *err)
	})
}

func TestUserService_UpdateUserCredentials_UserNotFound(t *testing.T) {
	userStoreMock := newUserStoreInterfaceMock(t)
	userStoreMock.
		On("GetCredentials", mock.Anything, svcTestUserID1).
		Return(User{}, Credentials{}, ErrUserNotFound).
		Once()

	service := &userService{
		userStore:     userStoreMock,
		transactioner: &fakeTransactioner{},
	}

	credentialsJSON := json.RawMessage(`{"password":"newpassword"}`)
	svcErr := service.UpdateUserCredentials(context.Background(), svcTestUserID1, credentialsJSON)
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorUserNotFound, *svcErr)
	userStoreMock.AssertNotCalled(t, "UpdateUserCredentials", mock.Anything, mock.Anything, mock.Anything)
}

func TestUserService_UpdateUserCredentials_Succeeds(t *testing.T) {
	userStoreMock := newUserStoreInterfaceMock(t)
	existingCredentials := Credentials{
		CredentialType("password"): {
			{
				StorageType: "hash",
				StorageAlgo: hash.SHA256,
				Value:       "old-hash",
				StorageAlgoParams: hash.CredParameters{
					Salt: "old-salt",
				},
			},
		},
		CredentialType("pin"): {
			{
				StorageType: "hash",
				StorageAlgo: hash.SHA256,
				Value:       "pin-hash",
				StorageAlgoParams: hash.CredParameters{
					Salt: "pin-salt",
				},
			},
		},
	}
	userStoreMock.
		On("GetCredentials", mock.Anything, svcTestUserID1).
		Return(User{ID: svcTestUserID1, Type: "Person"}, existingCredentials, nil).
		Once()
	var captured Credentials
	userStoreMock.
		On("UpdateUserCredentials", mock.Anything, svcTestUserID1, mock.Anything).
		Run(func(args mock.Arguments) {
			if creds, ok := args[2].(Credentials); ok {
				captured = creds
			}
		}).
		Return(nil).
		Once()

	hashServiceMock := hashmock.NewHashServiceInterfaceMock(t)
	hashServiceMock.
		On("Generate", []byte("newpassword")).
		Return(hash.Credential{
			Algorithm: hash.PBKDF2,
			Hash:      "hashed-newpassword",
			Parameters: hash.CredParameters{
				Salt:       "salt123",
				Iterations: 10000,
				KeySize:    32,
			},
		}, nil).
		Once()

	schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
	schemaMock.On("GetCredentialAttributes", mock.Anything, "Person").
		Return([]string{"password", "pin"}, (*serviceerror.ServiceError)(nil)).
		Once()

	service := &userService{
		userStore:         userStoreMock,
		userSchemaService: schemaMock,
		hashService:       hashServiceMock,
		transactioner:     &fakeTransactioner{},
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

	// Send plain text password - service will hash it
	credentialsJSON := json.RawMessage(`{"password":"newpassword"}`)
	svcErr := service.UpdateUserCredentials(context.Background(), svcTestUserID1, credentialsJSON)
	require.Nil(t, svcErr)

	// Verify password credential was hashed and stored
	passwordCreds, exists := captured[CredentialType("password")]
	require.True(t, exists)
	require.Len(t, passwordCreds, 1)
	require.Equal(t, "hash", passwordCreds[0].StorageType)
	require.Equal(t, "hashed-newpassword", passwordCreds[0].Value)
	require.Equal(t, hash.PBKDF2, passwordCreds[0].StorageAlgo)
	require.Equal(t, "salt123", passwordCreds[0].StorageAlgoParams.Salt)
	require.Equal(t, 10000, passwordCreds[0].StorageAlgoParams.Iterations)
	require.Equal(t, 32, passwordCreds[0].StorageAlgoParams.KeySize)

	// Verify PIN credential was preserved
	pinCreds, exists := captured[CredentialType("pin")]
	require.True(t, exists)
	require.Len(t, pinCreds, 1)
	require.Equal(t, "pin-hash", pinCreds[0].Value)
	require.Equal(t, "pin-salt", pinCreds[0].StorageAlgoParams.Salt)
}

func TestUserService_UpdateUserCredentials_MultiplePasskeys(t *testing.T) {
	userStoreMock := newUserStoreInterfaceMock(t)
	existingCredentials := Credentials{}

	userStoreMock.
		On("GetCredentials", mock.Anything, svcTestUserID1).
		Return(User{ID: svcTestUserID1, Type: "Person"}, existingCredentials, nil).
		Once()

	var captured Credentials
	userStoreMock.
		On("UpdateUserCredentials", mock.Anything, svcTestUserID1, mock.Anything).
		Run(func(args mock.Arguments) {
			if creds, ok := args[2].(Credentials); ok {
				captured = creds
			}
		}).
		Return(nil).
		Once()

	hashServiceMock := hashmock.NewHashServiceInterfaceMock(t)

	schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
	schemaMock.On("GetCredentialAttributes", mock.Anything, "Person").
		Return([]string{}, (*serviceerror.ServiceError)(nil)).
		Once()

	service := &userService{
		userStore:         userStoreMock,
		hashService:       hashServiceMock,
		userSchemaService: schemaMock,
		transactioner:     &fakeTransactioner{},
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

	// Send multiple passkeys as an array - passkey supports multiple credentials
	credentialsJSON := json.RawMessage(
		`{"passkey":[{"value":"passkey-credential-1"}, {"value":"passkey-credential-2"}]}`)
	svcErr := service.UpdateUserCredentials(context.Background(), svcTestUserID1, credentialsJSON)
	require.Nil(t, svcErr)

	// Verify both passkeys were stored (not hashed)
	passkeyCreds, exists := captured[CredentialTypePasskey]
	require.True(t, exists)
	require.Len(t, passkeyCreds, 2)

	// First passkey
	require.Equal(t, "passkey-credential-1", passkeyCreds[0].Value)

	// Second passkey
	require.Equal(t, "passkey-credential-2", passkeyCreds[1].Value)
}

func TestUserService_UpdateUserCredentials_RejectsMultiplePasswords(t *testing.T) {
	userStoreMock := newUserStoreInterfaceMock(t)
	existingCredentials := Credentials{}

	userStoreMock.
		On("GetCredentials", mock.Anything, svcTestUserID1).
		Return(User{ID: svcTestUserID1, Type: "Person"}, existingCredentials, nil).
		Once()

	hashServiceMock := hashmock.NewHashServiceInterfaceMock(t)

	schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
	schemaMock.On("GetCredentialAttributes", mock.Anything, "Person").
		Return([]string{"password"}, (*serviceerror.ServiceError)(nil)).
		Once()

	service := &userService{
		userStore:         userStoreMock,
		hashService:       hashServiceMock,
		userSchemaService: schemaMock,
		transactioner:     &fakeTransactioner{},
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

	// Attempt to send multiple passwords - should be rejected
	credentialsJSON := json.RawMessage(`{"password":[{"value":"password1"}, {"value":"password2"}]}`)
	svcErr := service.UpdateUserCredentials(context.Background(), svcTestUserID1, credentialsJSON)

	// Should return error
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorInvalidCredential.Code, svcErr.Code)
	require.Contains(t, svcErr.ErrorDescription, "does not support multiple credentials")

	// Store should not be called
	userStoreMock.AssertNotCalled(t, "UpdateUserCredentials", mock.Anything, mock.Anything, mock.Anything)
}

func TestUserService_GetUserCredentialsByType_Validation(t *testing.T) {
	service := &userService{}

	// Test missing user ID
	creds, err := service.GetUserCredentialsByType(context.Background(), "", "password")
	require.Nil(t, creds)
	require.NotNil(t, err)
	require.Equal(t, ErrorMissingUserID, *err)

	// Test missing credential type
	creds, err = service.GetUserCredentialsByType(context.Background(), svcTestUserID1, "")
	require.Nil(t, creds)
	require.NotNil(t, err)
	require.Equal(t, ErrorInvalidRequestFormat, *err)
}

func TestUserService_GetUserCredentialsByType_UserNotFound(t *testing.T) {
	userStoreMock := newUserStoreInterfaceMock(t)
	userStoreMock.
		On("GetCredentials", mock.Anything, svcTestUserID1).
		Return(User{}, Credentials{}, ErrUserNotFound).
		Once()

	service := &userService{
		userStore: userStoreMock,
	}

	creds, err := service.GetUserCredentialsByType(context.Background(), svcTestUserID1, "password")
	require.Nil(t, creds)
	require.NotNil(t, err)
	require.Equal(t, ErrorUserNotFound, *err)
}

func TestUserService_GetUserCredentialsByType_StoreError(t *testing.T) {
	userStoreMock := newUserStoreInterfaceMock(t)
	userStoreMock.
		On("GetCredentials", mock.Anything, svcTestUserID1).
		Return(User{}, Credentials{}, errors.New("database error")).
		Once()

	service := &userService{
		userStore: userStoreMock,
	}

	creds, err := service.GetUserCredentialsByType(context.Background(), svcTestUserID1, "password")
	require.Nil(t, creds)
	require.NotNil(t, err)
	require.Equal(t, ErrorInternalServerError.Code, err.Code)
}

func TestUserService_GetUserCredentialsByType_CredentialTypeNotFound(t *testing.T) {
	userStoreMock := newUserStoreInterfaceMock(t)
	existingCredentials := Credentials{
		"pin": {
			{
				StorageType: "hash",
				Value:       "pin-hash",
			},
		},
	}
	userStoreMock.
		On("GetCredentials", mock.Anything, svcTestUserID1).
		Return(User{ID: svcTestUserID1}, existingCredentials, nil).
		Once()

	service := &userService{
		userStore: userStoreMock,
	}

	// Request password credentials when only pin exists
	creds, err := service.GetUserCredentialsByType(context.Background(), svcTestUserID1, "password")
	require.Nil(t, err)
	require.NotNil(t, creds)
	require.Empty(t, creds) // Should return empty array, not nil
}

func TestUserService_GetUserCredentialsByType_EmptyCredentialArray(t *testing.T) {
	userStoreMock := newUserStoreInterfaceMock(t)
	existingCredentials := Credentials{
		"password": {}, // Empty array
		"pin": {
			{
				StorageType: "hash",
				Value:       "pin-hash",
			},
		},
	}
	userStoreMock.
		On("GetCredentials", mock.Anything, svcTestUserID1).
		Return(User{ID: svcTestUserID1}, existingCredentials, nil).
		Once()

	service := &userService{
		userStore: userStoreMock,
	}

	// Request password credentials when array is empty
	creds, err := service.GetUserCredentialsByType(context.Background(), svcTestUserID1, "password")
	require.Nil(t, err)
	require.NotNil(t, creds)
	require.Empty(t, creds) // Should return empty array
}

func TestUserService_GetUserCredentialsByType_Succeeds(t *testing.T) {
	userStoreMock := newUserStoreInterfaceMock(t)
	existingCredentials := Credentials{
		"password": {
			{
				StorageType: "hash",
				StorageAlgo: hash.PBKDF2,
				Value:       "hashed-password",
				StorageAlgoParams: hash.CredParameters{
					Salt:       "salt123",
					Iterations: 10000,
					KeySize:    32,
				},
			},
		},
		"passkey": {
			{
				Value: "public-key-1",
			},
			{
				Value: "public-key-2",
			},
		},
	}
	userStoreMock.
		On("GetCredentials", mock.Anything, svcTestUserID1).
		Return(User{ID: svcTestUserID1}, existingCredentials, nil).
		Once()

	service := &userService{
		userStore: userStoreMock,
	}

	// Get password credentials
	creds, err := service.GetUserCredentialsByType(context.Background(), svcTestUserID1, "password")
	require.Nil(t, err)
	require.NotNil(t, creds)
	require.Len(t, creds, 1)
	require.Equal(t, "hash", creds[0].StorageType)
	require.Equal(t, hash.PBKDF2, creds[0].StorageAlgo)
	require.Equal(t, "hashed-password", creds[0].Value)
	require.Equal(t, "salt123", creds[0].StorageAlgoParams.Salt)
	require.Equal(t, 10000, creds[0].StorageAlgoParams.Iterations)
	require.Equal(t, 32, creds[0].StorageAlgoParams.KeySize)
}

func TestUserService_GetUserCredentialsByType_MultipleCredentials(t *testing.T) {
	userStoreMock := newUserStoreInterfaceMock(t)
	existingCredentials := Credentials{
		"passkey": {
			{Value: "public-key-1"},
			{Value: "public-key-2"},
			{Value: "public-key-3"},
		},
	}
	userStoreMock.
		On("GetCredentials", mock.Anything, svcTestUserID1).
		Return(User{ID: svcTestUserID1}, existingCredentials, nil).
		Once()

	service := &userService{
		userStore: userStoreMock,
	}

	// Get passkey credentials
	creds, err := service.GetUserCredentialsByType(context.Background(), svcTestUserID1, "passkey")
	require.Nil(t, err)
	require.NotNil(t, creds)
	require.Len(t, creds, 3)
	require.Equal(t, "public-key-1", creds[0].Value)
	require.Equal(t, "public-key-2", creds[1].Value)
	require.Equal(t, "public-key-3", creds[2].Value)
}

func TestUserService_UpdateUserAttributes_Validation(t *testing.T) {
	service := &userService{}

	resp, err := service.UpdateUserAttributes(context.Background(), "", json.RawMessage(`{"email":"a@b.com"}`))
	require.Nil(t, resp)
	require.NotNil(t, err)
	require.Equal(t, ErrorMissingUserID, *err)

	resp, err = service.UpdateUserAttributes(context.Background(), svcTestUserID1, json.RawMessage{})
	require.Nil(t, resp)
	require.NotNil(t, err)
	require.Equal(t, ErrorInvalidRequestFormat, *err)
}

func TestUserService_UpdateUserAttributes_RejectsCredentialAttributes(t *testing.T) {
	storeMock := newUserStoreInterfaceMock(t)
	storeMock.On("GetUser", mock.Anything, svcTestUserID1).
		Return(User{ID: svcTestUserID1, Type: "Person"}, nil).
		Once()

	schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
	schemaMock.On("GetCredentialAttributes", mock.Anything, "Person").
		Return([]string{"password"}, (*serviceerror.ServiceError)(nil)).
		Once()

	service := &userService{
		userStore:         storeMock,
		userSchemaService: schemaMock,
	}

	resp, err := service.UpdateUserAttributes(context.Background(), svcTestUserID1,
		json.RawMessage(`{"password":"Secret123"}`))
	require.Nil(t, resp)
	require.NotNil(t, err)
	require.Equal(t, ErrorInvalidRequestFormat, *err)
}

func TestUserService_UpdateUserAttributes_UserNotFound(t *testing.T) {
	storeMock := newUserStoreInterfaceMock(t)
	storeMock.On("GetUser", mock.Anything, svcTestUserID1).Return(User{}, ErrUserNotFound).Once()

	service := &userService{
		userStore: storeMock,
	}

	resp, err := service.UpdateUserAttributes(context.Background(), svcTestUserID1,
		json.RawMessage(`{"email":"a@b.com"}`))
	require.Nil(t, resp)
	require.NotNil(t, err)
	require.Equal(t, ErrorUserNotFound, *err)
	storeMock.AssertNotCalled(t, "UpdateUser", mock.Anything)
}

func TestUserService_UpdateUserAttributes_SchemaValidationFails(t *testing.T) {
	storeMock := newUserStoreInterfaceMock(t)
	storeMock.
		On("GetUser", mock.Anything, svcTestUserID1).
		Return(User{ID: svcTestUserID1, Type: testUserType, Attributes: json.RawMessage(`{"email":"old"}`)}, nil)

	schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
	schemaMock.On("GetCredentialAttributes", mock.Anything, testUserType).
		Return([]string{"password"}, (*serviceerror.ServiceError)(nil)).
		Once()
	schemaMock.
		On("ValidateUser", mock.Anything, testUserType, mock.Anything).
		Return(false, &userschema.ErrorUserSchemaNotFound).
		Once()

	service := &userService{
		userStore:         storeMock,
		userSchemaService: schemaMock,
		transactioner:     &fakeTransactioner{},
	}

	resp, err := service.UpdateUserAttributes(context.Background(), svcTestUserID1,
		json.RawMessage(`{"email":"new@example.com"}`))
	require.Nil(t, resp)
	require.NotNil(t, err)
	require.Equal(t, ErrorUserSchemaNotFound, *err)
	storeMock.AssertNotCalled(t, "UpdateUser", mock.Anything)
}

func TestUserService_UpdateUserAttributes_Succeeds(t *testing.T) {
	storeMock := newUserStoreInterfaceMock(t)
	storeMock.
		On("GetUser", mock.Anything, svcTestUserID1).
		Return(User{ID: svcTestUserID1, Type: testUserType,
			Attributes: json.RawMessage(`{"email":"old@example.com"}`)}, nil)

	schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
	schemaMock.On("GetCredentialAttributes", mock.Anything, testUserType).
		Return([]string{"password"}, (*serviceerror.ServiceError)(nil)).
		Once()
	schemaMock.
		On("ValidateUser", mock.Anything, testUserType, mock.Anything).
		Return(true, (*serviceerror.ServiceError)(nil)).
		Once()
	schemaMock.
		On("ValidateUserUniqueness", mock.Anything, testUserType, mock.Anything, mock.Anything).
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
		transactioner:     &fakeTransactioner{},
	}

	newAttrs := json.RawMessage(`{"email":"new@example.com"}`)
	resp, err := service.UpdateUserAttributes(context.Background(), svcTestUserID1, newAttrs)
	require.Nil(t, err)
	require.NotNil(t, resp)
	require.Equal(t, svcTestUserID1, resp.ID)
	require.JSONEq(t, string(newAttrs), string(resp.Attributes))

	require.NotNil(t, savedUser)
	require.Equal(t, svcTestUserID1, savedUser.ID)
	require.JSONEq(t, string(newAttrs), string(savedUser.Attributes))
}

func TestUserService_GetUser_ReturnsUser(t *testing.T) {
	userID := svcTestUserID1
	expectedUser := User{ID: userID, OrganizationUnit: testOrgID}

	storeMock := newUserStoreInterfaceMock(t)
	storeMock.On("GetUser", mock.Anything, userID).Return(expectedUser, nil).Once()

	service := &userService{
		userStore: storeMock,
	}

	user, err := service.GetUser(context.Background(), userID)
	require.Nil(t, err)
	require.Equal(t, expectedUser, *user)
}

func TestUserService_DeleteUser(t *testing.T) {
	userID := svcTestUserID1

	storeMock := newUserStoreInterfaceMock(t)
	storeMock.On("DeleteUser", mock.Anything, userID).Return(nil).Once()

	txMock := &fakeTransactioner{}

	service := &userService{
		userStore:     storeMock,
		transactioner: txMock,
	}

	err := service.DeleteUser(context.Background(), userID)
	require.Nil(t, err)
	require.Equal(t, 1, txMock.transactCalls)
}

func TestUserService_UpdateUser(t *testing.T) {
	userID := svcTestUserID1
	updatedUser := User{ID: userID, OrganizationUnit: testOrgID, Type: testUserType,
		Attributes: json.RawMessage(`{"updated":"true"}`)}

	storeMock := newUserStoreInterfaceMock(t)

	// Mock UpdateUser call
	storeMock.On("UpdateUser", mock.Anything, mock.MatchedBy(func(u *User) bool {
		return u.ID == userID
	})).Return(nil).Once()

	ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
	ouServiceMock.On("IsOrganizationUnitExists", testOrgID).Return(true, (*serviceerror.ServiceError)(nil)).Once()

	userSchemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
	userSchemaMock.On("GetCredentialAttributes", mock.Anything, testUserType).
		Return([]string{"password"}, (*serviceerror.ServiceError)(nil)).Once()
	userSchemaMock.On("GetUserSchemaByName", mock.Anything, testUserType).
		Return(&userschema.UserSchema{OrganizationUnitID: testOrgID}, (*serviceerror.ServiceError)(nil)).
		Once()
	userSchemaMock.On("ValidateUser", mock.Anything, testUserType, mock.Anything).
		Return(true, (*serviceerror.ServiceError)(nil)).Once()
	userSchemaMock.On("ValidateUserUniqueness", mock.Anything, testUserType, mock.Anything, mock.Anything).
		Return(true, (*serviceerror.ServiceError)(nil)).Once()

	txMock := &fakeTransactioner{}

	service := &userService{
		userStore:         storeMock,
		ouService:         ouServiceMock,
		userSchemaService: userSchemaMock,
		transactioner:     txMock,
	}

	resp, err := service.UpdateUser(context.Background(), userID, &updatedUser)
	_ = resp // Check resp fields if needed, or ignore if only testing error/transaction
	require.Nil(t, err)
	require.Equal(t, 1, txMock.transactCalls)
}

func TestUserService_UpdateUser_WithCredentials(t *testing.T) {
	userID := svcTestUserID1

	// Test the new credential extraction, merging, and update logic
	updatedUser := User{
		ID:               userID,
		OrganizationUnit: testOrgID,
		Type:             testUserType,
		Attributes:       json.RawMessage(`{"email":"test@example.com","password":"newPassword123"}`),
	}

	storeMock := newUserStoreInterfaceMock(t)
	hashMock := hashmock.NewHashServiceInterfaceMock(t)
	ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
	userSchemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
	txMock := &fakeTransactioner{}

	// Mock hash generation for password
	hashMock.On("Generate", mock.Anything).Return(hash.Credential{
		Algorithm: "pbkdf2",
		Hash:      "hashedPassword",
		Parameters: hash.CredParameters{
			Salt:       "somesalt",
			Iterations: 10000,
			KeySize:    32,
		},
	}, nil).Once()

	// Mock validation calls
	ouServiceMock.On("IsOrganizationUnitExists", testOrgID).Return(true, (*serviceerror.ServiceError)(nil)).Once()
	userSchemaMock.On("GetCredentialAttributes", mock.Anything, testUserType).
		Return([]string{"password"}, (*serviceerror.ServiceError)(nil)).Once()
	userSchemaMock.On("GetUserSchemaByName", mock.Anything, testUserType).
		Return(&userschema.UserSchema{OrganizationUnitID: testOrgID}, (*serviceerror.ServiceError)(nil)).Once()
	userSchemaMock.On("ValidateUser", mock.Anything, testUserType, mock.Anything).
		Return(true, (*serviceerror.ServiceError)(nil)).Once()
	userSchemaMock.On("ValidateUserUniqueness", mock.Anything, testUserType, mock.Anything, mock.Anything).
		Return(true, (*serviceerror.ServiceError)(nil)).Once()

	// Mock UpdateUser - should receive user WITHOUT password in attributes
	storeMock.On("UpdateUser", mock.Anything, mock.MatchedBy(func(u *User) bool {
		var attrs map[string]interface{}
		if err := json.Unmarshal(u.Attributes, &attrs); err != nil {
			return false
		}
		_, hasPassword := attrs["password"]
		return u.ID == userID && !hasPassword // Password should be removed from attributes
	})).Return(nil).Once()

	// Mock GetCredentials - return existing credentials (e.g., passkey)
	existingCreds := Credentials{
		CredentialTypePasskey: []Credential{
			{Value: "existingPasskey", StorageType: "passkey"},
		},
	}
	storeMock.On("GetCredentials", mock.Anything, userID).
		Return(User{ID: userID, Type: testUserType}, existingCreds, nil).Once()

	// Mock UpdateUserCredentials - should receive merged credentials (password + existing passkey)
	storeMock.On("UpdateUserCredentials", mock.Anything, userID, mock.MatchedBy(func(creds Credentials) bool {
		passwordCreds, hasPassword := creds[CredentialType("password")]
		passkeyCreds, hasPasskey := creds[CredentialTypePasskey]
		// Verify password was added and passkey was preserved
		return hasPassword && len(passwordCreds) == 1 && hasPasskey && len(passkeyCreds) == 1
	})).Return(nil).Once()

	service := &userService{
		userStore:         storeMock,
		ouService:         ouServiceMock,
		userSchemaService: userSchemaMock,
		hashService:       hashMock,
		transactioner:     txMock,
	}

	resp, err := service.UpdateUser(context.Background(), userID, &updatedUser)

	// Assertions
	require.Nil(t, err)
	require.NotNil(t, resp)
	require.Equal(t, userID, resp.ID)
	require.Equal(t, 1, txMock.transactCalls)

	// Verify all expected calls were made
	storeMock.AssertExpectations(t)
	hashMock.AssertExpectations(t)
	ouServiceMock.AssertExpectations(t)
	userSchemaMock.AssertExpectations(t)
}

func TestUserService_UpdateUser_ErrorPaths(t *testing.T) {
	userID := svcTestUserID1
	ctx := context.Background()

	t.Run("UserNotFound_InCredentialUpdate", func(t *testing.T) {
		updatedUser := User{
			ID:               userID,
			OrganizationUnit: testOrgID,
			Type:             testUserType,
			Attributes:       json.RawMessage(`{"email":"test@example.com","password":"newPassword"}`),
		}

		storeMock := newUserStoreInterfaceMock(t)
		hashMock := hashmock.NewHashServiceInterfaceMock(t)
		ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
		userSchemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
		txMock := &fakeTransactioner{}

		hashMock.On("Generate", mock.Anything).Return(hash.Credential{
			Algorithm: "pbkdf2", Hash: "hashed", Parameters: hash.CredParameters{Salt: "s", Iterations: 1, KeySize: 32},
		}, nil).Once()

		userSchemaMock.On("GetCredentialAttributes", mock.Anything, testUserType).
			Return([]string{"password"}, (*serviceerror.ServiceError)(nil)).Maybe()
		ouServiceMock.On("IsOrganizationUnitExists", testOrgID).Return(true, (*serviceerror.ServiceError)(nil)).Maybe()
		userSchemaMock.On("GetUserSchemaByName", mock.Anything, testUserType).
			Return(&userschema.UserSchema{OrganizationUnitID: testOrgID}, (*serviceerror.ServiceError)(nil)).Maybe()
		userSchemaMock.On("ValidateUser", mock.Anything, testUserType, mock.Anything).
			Return(true, (*serviceerror.ServiceError)(nil)).Maybe()
		userSchemaMock.On("ValidateUserUniqueness", mock.Anything, testUserType, mock.Anything, mock.Anything).
			Return(true, (*serviceerror.ServiceError)(nil)).Maybe()

		storeMock.On("UpdateUser", mock.Anything, mock.Anything).Return(nil).Once()
		storeMock.On("GetCredentials", mock.Anything, userID).Return(User{}, Credentials{}, ErrUserNotFound).Once()

		service := &userService{
			userStore: storeMock, ouService: ouServiceMock, userSchemaService: userSchemaMock,
			hashService: hashMock, transactioner: txMock,
		}

		resp, err := service.UpdateUser(ctx, userID, &updatedUser)
		require.NotNil(t, err)
		require.Nil(t, resp)
		require.Equal(t, ErrorUserNotFound.Code, err.Code)
	})

	t.Run("GenericError_InCredentialUpdate", func(t *testing.T) {
		updatedUser := User{
			ID:               userID,
			OrganizationUnit: testOrgID,
			Type:             testUserType,
			Attributes:       json.RawMessage(`{"email":"test@example.com","password":"newPass"}`),
		}

		storeMock := newUserStoreInterfaceMock(t)
		hashMock := hashmock.NewHashServiceInterfaceMock(t)
		ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
		userSchemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
		txMock := &fakeTransactioner{}

		hashMock.On("Generate", mock.Anything).Return(hash.Credential{
			Algorithm: "pbkdf2", Hash: "hashed", Parameters: hash.CredParameters{Salt: "s", Iterations: 1, KeySize: 32},
		}, nil).Once()

		userSchemaMock.On("GetCredentialAttributes", mock.Anything, testUserType).
			Return([]string{"password"}, (*serviceerror.ServiceError)(nil)).Maybe()
		ouServiceMock.On("IsOrganizationUnitExists", testOrgID).Return(true, (*serviceerror.ServiceError)(nil)).Maybe()
		userSchemaMock.On("GetUserSchemaByName", mock.Anything, testUserType).
			Return(&userschema.UserSchema{OrganizationUnitID: testOrgID}, (*serviceerror.ServiceError)(nil)).Maybe()
		userSchemaMock.On("ValidateUser", mock.Anything, testUserType, mock.Anything).
			Return(true, (*serviceerror.ServiceError)(nil)).Maybe()
		userSchemaMock.On("ValidateUserUniqueness", mock.Anything, testUserType, mock.Anything, mock.Anything).
			Return(true, (*serviceerror.ServiceError)(nil)).Maybe()

		storeMock.On("UpdateUser", mock.Anything, mock.Anything).Return(errors.New("db connection lost")).Once()

		service := &userService{
			userStore: storeMock, ouService: ouServiceMock, userSchemaService: userSchemaMock,
			hashService: hashMock, transactioner: txMock,
		}

		resp, err := service.UpdateUser(ctx, userID, &updatedUser)
		require.NotNil(t, err)
		require.Nil(t, resp)
		require.Equal(t, ErrorInternalServerError.Code, err.Code)
	})

	t.Run("UpdateUser_WithoutCredentials_Success", func(t *testing.T) {
		updatedUser := User{
			ID:               userID,
			OrganizationUnit: testOrgID,
			Type:             testUserType,
			Attributes:       json.RawMessage(`{"email":"updated@example.com"}`),
		}

		storeMock := newUserStoreInterfaceMock(t)
		ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
		userSchemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
		txMock := &fakeTransactioner{}

		userSchemaMock.On("GetCredentialAttributes", mock.Anything, testUserType).
			Return([]string{"password"}, (*serviceerror.ServiceError)(nil)).Once()
		ouServiceMock.On("IsOrganizationUnitExists", testOrgID).Return(true, (*serviceerror.ServiceError)(nil)).Once()
		userSchemaMock.On("GetUserSchemaByName", mock.Anything, testUserType).
			Return(&userschema.UserSchema{OrganizationUnitID: testOrgID}, (*serviceerror.ServiceError)(nil)).Once()
		userSchemaMock.On("ValidateUser", mock.Anything, testUserType, mock.Anything).
			Return(true, (*serviceerror.ServiceError)(nil)).Once()
		userSchemaMock.On("ValidateUserUniqueness", mock.Anything, testUserType, mock.Anything, mock.Anything).
			Return(true, (*serviceerror.ServiceError)(nil)).Once()

		storeMock.On("UpdateUser", mock.Anything, mock.Anything).Return(nil).Once()

		service := &userService{
			userStore: storeMock, ouService: ouServiceMock, userSchemaService: userSchemaMock, transactioner: txMock,
		}

		resp, err := service.UpdateUser(ctx, userID, &updatedUser)
		require.Nil(t, err)
		require.NotNil(t, resp)
		require.Equal(t, userID, resp.ID)
		storeMock.AssertExpectations(t)
	})

	t.Run("ValidationError_InsideTransaction", func(t *testing.T) {
		updatedUser := User{
			ID:               userID,
			OrganizationUnit: testOrgID,
			Type:             testUserType,
			Attributes:       json.RawMessage(`{"email":"test@example.com"}`),
		}

		storeMock := newUserStoreInterfaceMock(t)
		ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
		userSchemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
		txMock := &fakeTransactioner{}

		userSchemaMock.On("GetCredentialAttributes", mock.Anything, testUserType).
			Return([]string{"password"}, (*serviceerror.ServiceError)(nil)).Maybe()
		ouServiceMock.On("IsOrganizationUnitExists", testOrgID).Return(true, (*serviceerror.ServiceError)(nil)).Maybe()
		userSchemaMock.On("GetUserSchemaByName", mock.Anything, testUserType).
			Return(&userschema.UserSchema{OrganizationUnitID: testOrgID}, (*serviceerror.ServiceError)(nil)).Maybe()
		// Validation fails - schema validation error
		userSchemaMock.On("ValidateUser", mock.Anything, testUserType, mock.Anything).
			Return(false, (*serviceerror.ServiceError)(nil)).Once()

		service := &userService{
			userStore: storeMock, ouService: ouServiceMock, userSchemaService: userSchemaMock, transactioner: txMock,
		}

		resp, err := service.UpdateUser(ctx, userID, &updatedUser)
		require.NotNil(t, err)
		require.Nil(t, resp)
		// When ValidateUser returns false with nil error, it becomes SchemaValidationFailed
		require.Equal(t, ErrorSchemaValidationFailed.Code, err.Code)
	})

	t.Run("ExtractCredentials_HashError", func(t *testing.T) {
		updatedUser := User{
			ID:               userID,
			OrganizationUnit: testOrgID,
			Type:             testUserType,
			Attributes:       json.RawMessage(`{"email":"test@example.com","password":"somePassword"}`),
		}

		storeMock := newUserStoreInterfaceMock(t)
		hashMock := hashmock.NewHashServiceInterfaceMock(t)
		ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
		userSchemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
		txMock := &fakeTransactioner{}

		userSchemaMock.On("GetCredentialAttributes", mock.Anything, testUserType).
			Return([]string{"password"}, (*serviceerror.ServiceError)(nil)).Maybe()
		ouServiceMock.On("IsOrganizationUnitExists", testOrgID).Return(true, (*serviceerror.ServiceError)(nil)).Maybe()
		userSchemaMock.On("GetUserSchemaByName", mock.Anything, testUserType).
			Return(&userschema.UserSchema{OrganizationUnitID: testOrgID}, (*serviceerror.ServiceError)(nil)).Maybe()
		userSchemaMock.On("ValidateUser", mock.Anything, testUserType, mock.Anything).
			Return(true, (*serviceerror.ServiceError)(nil)).Maybe()
		userSchemaMock.On("ValidateUserUniqueness", mock.Anything, testUserType, mock.Anything, mock.Anything).
			Return(true, (*serviceerror.ServiceError)(nil)).Maybe()

		// Hash generation fails before the transaction is opened; the DB connection is never acquired.
		hashMock.On("Generate", mock.Anything).
			Return(hash.Credential{}, errors.New("hash generation failed")).Once()

		service := &userService{
			userStore: storeMock, ouService: ouServiceMock, userSchemaService: userSchemaMock,
			hashService: hashMock, transactioner: txMock,
		}

		resp, err := service.UpdateUser(ctx, userID, &updatedUser)
		require.NotNil(t, err)
		require.Nil(t, resp)
		require.Equal(t, ErrorInternalServerError.Code, err.Code)
		// Transaction must NOT have been entered - the connection pool was not touched.
		require.Equal(t, 0, txMock.transactCalls)
	})

	t.Run("UpdateCredentials_StoreError", func(t *testing.T) {
		updatedUser := User{
			ID:               userID,
			OrganizationUnit: testOrgID,
			Type:             testUserType,
			Attributes:       json.RawMessage(`{"email":"test@example.com","password":"somePassword"}`),
		}

		storeMock := newUserStoreInterfaceMock(t)
		hashMock := hashmock.NewHashServiceInterfaceMock(t)
		ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
		userSchemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
		txMock := &fakeTransactioner{}

		userSchemaMock.On("GetCredentialAttributes", mock.Anything, testUserType).
			Return([]string{"password"}, (*serviceerror.ServiceError)(nil)).Maybe()
		ouServiceMock.On("IsOrganizationUnitExists", testOrgID).Return(true, (*serviceerror.ServiceError)(nil)).Maybe()
		userSchemaMock.On("GetUserSchemaByName", mock.Anything, testUserType).
			Return(&userschema.UserSchema{OrganizationUnitID: testOrgID}, (*serviceerror.ServiceError)(nil)).Maybe()
		userSchemaMock.On("ValidateUser", mock.Anything, testUserType, mock.Anything).
			Return(true, (*serviceerror.ServiceError)(nil)).Maybe()
		userSchemaMock.On("ValidateUserUniqueness", mock.Anything, testUserType, mock.Anything, mock.Anything).
			Return(true, (*serviceerror.ServiceError)(nil)).Maybe()

		hashMock.On("Generate", mock.Anything).Return(hash.Credential{
			Algorithm: "pbkdf2", Hash: "hashed", Parameters: hash.CredParameters{Salt: "s", Iterations: 1, KeySize: 32},
		}, nil).Once()

		storeMock.On("UpdateUser", mock.Anything, mock.Anything).Return(nil).Once()
		storeMock.On("GetCredentials", mock.Anything, userID).Return(User{ID: userID}, Credentials{}, nil).Once()
		// UpdateUserCredentials fails - triggers return err inside if len(credentials) > 0
		storeMock.On("UpdateUserCredentials", mock.Anything, userID, mock.Anything).
			Return(errors.New("failed to update credentials")).Once()

		service := &userService{
			userStore: storeMock, ouService: ouServiceMock, userSchemaService: userSchemaMock,
			hashService: hashMock, transactioner: txMock,
		}

		resp, err := service.UpdateUser(ctx, userID, &updatedUser)
		require.NotNil(t, err)
		require.Nil(t, resp)
		require.Equal(t, ErrorInternalServerError.Code, err.Code)
	})
}

func TestUserService_UpdateUser_PreservesMultipleCredentials(t *testing.T) {
	ctx := context.Background()
	userID := svcTestUserID123
	testOU := testOrgID

	// User update with ONLY password (should preserve PIN)
	updatedUser := User{
		ID:               userID,
		Type:             testUserType,
		OrganizationUnit: testOU,
		Attributes: json.RawMessage(`{
			"username": "john.doe",
			"email": "john.updated@example.com",
			"firstName": "John",
			"lastName": "Doe",
			"password": "NewPassword456!"
		}`),
	}

	// Existing credentials in database: password + PIN
	existingCredentials := Credentials{
		CredentialType("password"): []Credential{
			{
				StorageType: "hash",
				StorageAlgo: hash.PBKDF2,
				Value:       "old_hashed_password",
				StorageAlgoParams: hash.CredParameters{
					Salt:       "old_salt",
					Iterations: 10000,
					KeySize:    32,
				},
			},
		},
		CredentialType("pin"): []Credential{
			{
				StorageType: "hash",
				StorageAlgo: hash.PBKDF2,
				Value:       "hashed_pin_123456",
				StorageAlgoParams: hash.CredParameters{
					Salt:       "pin_salt",
					Iterations: 10000,
					KeySize:    32,
				},
			},
		},
	}

	// Setup mocks
	storeMock := newUserStoreInterfaceMock(t)
	ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
	userSchemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
	hashMock := hashmock.NewHashServiceInterfaceMock(t)
	txMock := &fakeTransactioner{}

	// Mock OU validation
	ouServiceMock.On("IsOrganizationUnitExists", testOU).
		Return(true, (*serviceerror.ServiceError)(nil)).Once()
	ouServiceMock.On("IsParent", mock.Anything, testOU).
		Return(true, (*serviceerror.ServiceError)(nil)).Maybe()

	// Mock schema validation
	userSchemaMock.On("GetCredentialAttributes", mock.Anything, testUserType).
		Return([]string{"password", "pin"}, (*serviceerror.ServiceError)(nil)).Once()
	userSchemaMock.On("GetUserSchemaByName", mock.Anything, testUserType).
		Return(&userschema.UserSchema{
			Name:               testUserType,
			OrganizationUnitID: testOU,
		}, (*serviceerror.ServiceError)(nil)).Once()
	userSchemaMock.On("ValidateUser", mock.Anything, testUserType, mock.Anything).
		Return(true, (*serviceerror.ServiceError)(nil)).Once()
	userSchemaMock.On("ValidateUserUniqueness", mock.Anything, testUserType, mock.Anything, mock.Anything).
		Return(true, (*serviceerror.ServiceError)(nil)).Once()

	// Mock hash generation for NEW password
	hashMock.On("Generate", []byte("NewPassword456!")).
		Return(hash.Credential{
			Algorithm: hash.PBKDF2,
			Hash:      "new_hashed_password",
			Parameters: hash.CredParameters{
				Salt:       "new_salt",
				Iterations: 10000,
				KeySize:    32,
			},
		}, nil).Once()

	// Mock UpdateUser - should receive user WITHOUT password in attributes
	storeMock.On("UpdateUser", mock.Anything, mock.MatchedBy(func(u *User) bool {
		var attrs map[string]interface{}
		if err := json.Unmarshal(u.Attributes, &attrs); err != nil {
			return false
		}
		_, hasPassword := attrs["password"]
		_, hasPin := attrs["pin"]
		return u.ID == userID && !hasPassword && !hasPin
	})).Return(nil).Once()

	// Mock GetCredentials - return existing credentials (password + PIN)
	storeMock.On("GetCredentials", mock.Anything, userID).
		Return(User{ID: userID}, existingCredentials, nil).Once()

	// Capture merged credentials passed to UpdateUserCredentials
	var capturedCredentials Credentials
	storeMock.On("UpdateUserCredentials", mock.Anything, userID, mock.Anything).
		Run(func(args mock.Arguments) {
			capturedCredentials = args.Get(2).(Credentials)
		}).
		Return(nil).Once()

	// Create service
	service := &userService{
		userStore:         storeMock,
		ouService:         ouServiceMock,
		userSchemaService: userSchemaMock,
		hashService:       hashMock,
		transactioner:     txMock,
	}

	// Execute UpdateUser
	result, svcErr := service.UpdateUser(ctx, userID, &updatedUser)

	// Assertions
	require.Nil(t, svcErr)
	require.NotNil(t, result)
	require.Equal(t, userID, result.ID)

	// Verify transaction was used
	require.Equal(t, 1, txMock.transactCalls)

	// Verify merged credentials
	require.NotNil(t, capturedCredentials)

	// Verify password was UPDATED (new hash)
	require.Len(t, capturedCredentials[CredentialType("password")], 1,
		"Password should be updated")
	require.Equal(t, "new_hashed_password",
		capturedCredentials[CredentialType("password")][0].Value,
		"Password should have new hashed value")
	require.Equal(t, "new_salt",
		capturedCredentials[CredentialType("password")][0].StorageAlgoParams.Salt,
		"Password should have new salt")

	// Verify PIN was PRESERVED (original hash)
	require.Len(t, capturedCredentials[CredentialType("pin")], 1,
		"PIN should be preserved during password update")
	require.Equal(t, "hashed_pin_123456",
		capturedCredentials[CredentialType("pin")][0].Value,
		"PIN should retain original hashed value")
	require.Equal(t, "pin_salt",
		capturedCredentials[CredentialType("pin")][0].StorageAlgoParams.Salt,
		"PIN should retain original salt")

	// Verify response attributes don't contain credentials
	var attrs map[string]interface{}
	err := json.Unmarshal(result.Attributes, &attrs)
	require.NoError(t, err)
	_, hasPassword := attrs["password"]
	_, hasPin := attrs["pin"]
	require.False(t, hasPassword, "Password should not be in response attributes")
	require.False(t, hasPin, "PIN should not be in response attributes")

	// Verify all mocks were called
	storeMock.AssertExpectations(t)
	ouServiceMock.AssertExpectations(t)
	userSchemaMock.AssertExpectations(t)
	hashMock.AssertExpectations(t)
}

func TestUserService_GetUserList(t *testing.T) {
	limit := 10
	offset := 0
	filters := map[string]interface{}{}

	storeMock := newUserStoreInterfaceMock(t)
	storeMock.On("GetUserListCount", mock.Anything, filters).Return(5, nil).Once()
	storeMock.On("GetUserList", mock.Anything, limit, offset, filters).
		Return([]User{{ID: svcTestUserID1}}, nil).
		Once()

	service := &userService{
		userStore: storeMock,
	}

	resp, err := service.GetUserList(context.Background(), limit, offset, filters)
	require.Nil(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 5, resp.TotalResults)
	require.Len(t, resp.Users, 1)
}

func TestUserService_GetUserGroups(t *testing.T) {
	mockStore := newUserStoreInterfaceMock(t)
	userID := svcTestUserID123
	limit, offset := 10, 0

	mockStore.On("ValidateUserIDs", mock.Anything, []string{userID}).Return([]string{}, nil)
	mockStore.On("GetGroupCountForUser", mock.Anything, userID).Return(5, nil)
	mockStore.On("GetUserGroups", mock.Anything, userID, limit, offset).
		Return([]UserGroup{{ID: "g1", Name: "Group 1"}}, nil)

	service := &userService{userStore: mockStore}
	resp, err := service.GetUserGroups(context.Background(), userID, limit, offset)

	require.Nil(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 5, resp.TotalResults)
	require.Len(t, resp.Groups, 1)
}

func TestUserService_VerifyUser(t *testing.T) {
	mockStore := newUserStoreInterfaceMock(t)
	mockHash := hashmock.NewHashServiceInterfaceMock(t)
	userID := svcTestUserID123
	creds := map[string]interface{}{"password": "password123"}

	storedUser := User{ID: userID}
	storedCreds := Credentials{
		"password": []Credential{
			{
				Value:             "hashed_password",
				StorageAlgo:       "argon2id",
				StorageAlgoParams: hash.CredParameters{Salt: "salt"},
			},
		},
	}

	mockStore.On("GetCredentials", mock.Anything, userID).Return(storedUser, storedCreds, nil)
	mockHash.On("Verify", []byte("password123"), mock.Anything).Return(true, nil)

	service := &userService{
		userStore:   mockStore,
		hashService: mockHash,
	}

	user, err := service.VerifyUser(context.Background(), userID, creds)

	require.Nil(t, err)
	require.NotNil(t, user)
	require.Equal(t, userID, user.ID)
}

func TestUserService_AuthenticateUser(t *testing.T) {
	mockStore := newUserStoreInterfaceMock(t)
	mockHash := hashmock.NewHashServiceInterfaceMock(t)

	identifiers := map[string]interface{}{
		"username": "alice",
	}
	credentials := map[string]interface{}{
		"password": "password123",
	}

	userID := svcTestUserID123
	mockStore.On("IdentifyUser", mock.Anything, mock.Anything).Return(&userID, nil)

	storedUser := User{ID: userID, Type: "employee", OrganizationUnit: "ou-1"}
	storedCreds := Credentials{
		"password": []Credential{
			{
				Value:             "hashed_password",
				StorageAlgo:       "argon2id",
				StorageAlgoParams: hash.CredParameters{Salt: "salt"},
			},
		},
	}
	mockStore.On("GetCredentials", mock.Anything, userID).Return(storedUser, storedCreds, nil)
	mockHash.On("Verify", []byte("password123"), mock.Anything).Return(true, nil)

	service := &userService{
		userStore:   mockStore,
		hashService: mockHash,
	}

	resp, err := service.AuthenticateUser(context.Background(), identifiers, credentials)

	require.Nil(t, err)
	require.NotNil(t, resp)
	require.Equal(t, userID, resp.ID)
}

func TestUserService_ValidateUserIDs(t *testing.T) {
	mockStore := newUserStoreInterfaceMock(t)
	userIDs := []string{"u1", "u2"}

	mockStore.On("ValidateUserIDs", mock.Anything, userIDs).Return([]string{}, nil)

	service := &userService{userStore: mockStore}
	invalidIDs, err := service.ValidateUserIDs(context.Background(), userIDs)

	require.Nil(t, err)
	require.Len(t, invalidIDs, 0)
}

func TestUserService_GetUserGroups_ErrorCases(t *testing.T) {
	mockStore := newUserStoreInterfaceMock(t)
	service := &userService{userStore: mockStore}
	ctx := context.Background()

	t.Run("MissingUserID", func(t *testing.T) {
		_, err := service.GetUserGroups(ctx, "", 10, 0)
		require.NotNil(t, err)
		require.Equal(t, ErrorMissingUserID.Code, err.Code)
	})

	t.Run("InvalidPagination", func(t *testing.T) {
		_, err := service.GetUserGroups(ctx, "u1", -1, 0)
		require.NotNil(t, err)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		mockStore.On("ValidateUserIDs", mock.Anything, []string{"u1"}).Return([]string{"u1"}, nil).Once()
		_, err := service.GetUserGroups(ctx, "u1", 10, 0)
		require.NotNil(t, err)
		require.Equal(t, ErrorUserNotFound.Code, err.Code)
	})

	t.Run("StoreErrorOnValidate", func(t *testing.T) {
		mockStore.On("ValidateUserIDs", mock.Anything, []string{"u1"}).Return(nil, errors.New("db error")).Once()
		_, err := service.GetUserGroups(ctx, "u1", 10, 0)
		require.NotNil(t, err)
		require.Equal(t, ErrorInternalServerError.Code, err.Code)
	})

	t.Run("StoreErrorOnCount", func(t *testing.T) {
		mockStore.On("ValidateUserIDs", mock.Anything, []string{"u1"}).Return([]string{}, nil).Once()
		mockStore.On("GetGroupCountForUser", mock.Anything, "u1").
			Return(0, errors.New("db error")).Once()
		_, err := service.GetUserGroups(ctx, "u1", 10, 0)
		require.NotNil(t, err)
		require.Equal(t, ErrorInternalServerError.Code, err.Code)
	})
}

func TestUserService_VerifyUser_ErrorCases(t *testing.T) {
	mockStore := newUserStoreInterfaceMock(t)
	mockHash := hashmock.NewHashServiceInterfaceMock(t)
	service := &userService{userStore: mockStore, hashService: mockHash}
	ctx := context.Background()

	t.Run("MissingUserID", func(t *testing.T) {
		_, err := service.VerifyUser(ctx, "", nil)
		require.NotNil(t, err)
	})

	t.Run("NoCredentials", func(t *testing.T) {
		_, err := service.VerifyUser(ctx, "u1", nil)
		require.NotNil(t, err)
	})

	t.Run("NoValidCredentials", func(t *testing.T) {
		mockStore.On("GetCredentials", mock.Anything, "u1").
			Return(User{ID: "u1"}, Credentials{CredentialType("password"): []Credential{{Value: "h"}}}, nil).Once()
		_, err := service.VerifyUser(ctx, "u1", map[string]interface{}{"invalid": "val"})
		require.NotNil(t, err)
		require.Equal(t, ErrorAuthenticationFailed.Code, err.Code)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		mockStore.On("GetCredentials", mock.Anything, "u1").Return(User{}, nil, ErrUserNotFound).Once()
		_, err := service.VerifyUser(ctx, "u1", map[string]interface{}{"password": "p"})
		require.NotNil(t, err)
		require.Equal(t, ErrorUserNotFound.Code, err.Code)
	})

	t.Run("NoStoredCredentials", func(t *testing.T) {
		mockStore.On("GetCredentials", mock.Anything, "u1").Return(User{ID: "u1"}, Credentials{}, nil).Once()
		_, err := service.VerifyUser(ctx, "u1", map[string]interface{}{"password": "p"})
		require.NotNil(t, err)
		require.Equal(t, ErrorAuthenticationFailed.Code, err.Code)
	})

	t.Run("CredentialTypeMismatch", func(t *testing.T) {
		mockStore.On("GetCredentials", mock.Anything, "u1").
			Return(User{ID: "u1"}, Credentials{"pin": []Credential{{}}}, nil).Once()
		_, err := service.VerifyUser(ctx, "u1", map[string]interface{}{"password": "p"})
		require.NotNil(t, err)
		require.Equal(t, ErrorAuthenticationFailed.Code, err.Code)
	})

	t.Run("HashVerifyFalse", func(t *testing.T) {
		storedCreds := Credentials{"password": []Credential{{Value: "h", StorageAlgo: "a"}}}
		mockStore.On("GetCredentials", mock.Anything, "u1").Return(User{ID: "u1"}, storedCreds, nil).Once()
		mockHash.On("Verify", mock.Anything, mock.Anything).Return(false, nil).Once()
		_, err := service.VerifyUser(ctx, "u1", map[string]interface{}{"password": "p"})
		require.NotNil(t, err)
		require.Equal(t, ErrorAuthenticationFailed.Code, err.Code)
	})
}

func TestUserService_AuthenticateUser_ErrorCases(t *testing.T) {
	mockStore := newUserStoreInterfaceMock(t)
	service := &userService{userStore: mockStore}
	ctx := context.Background()

	t.Run("EmptyIdentifiers", func(t *testing.T) {
		_, err := service.AuthenticateUser(ctx, nil, map[string]interface{}{"password": "p"})
		require.NotNil(t, err)
		require.Equal(t, ErrorMissingRequiredFields.Code, err.Code)
	})

	t.Run("EmptyCredentials", func(t *testing.T) {
		_, err := service.AuthenticateUser(ctx, map[string]interface{}{"username": "u"}, nil)
		require.NotNil(t, err)
		require.Equal(t, ErrorMissingCredentials.Code, err.Code)
	})

	t.Run("IdentifyUserNotFound", func(t *testing.T) {
		mockStore.On("IdentifyUser", mock.Anything, mock.Anything).Return((*string)(nil), ErrUserNotFound).Once()
		_, err := service.AuthenticateUser(ctx,
			map[string]interface{}{"username": "u"},
			map[string]interface{}{"password": "p"})
		require.NotNil(t, err)
		require.Equal(t, ErrorUserNotFound.Code, err.Code)
	})
}

func TestBuildPaginationLinks(t *testing.T) {
	links := buildPaginationLinks("/users", 10, 20, 55)
	// totalResults 55, limit 10
	// 0-9, 10-19, 20-29, 30-39, 40-49, 50-54
	// offset 20 (3rd page)
	// first: 0
	// prev: 10
	// next: 30
	// last: 50
	require.Len(t, links, 4)

	relMap := make(map[string]string)
	for _, l := range links {
		relMap[l.Rel] = l.Href
	}

	require.Equal(t, "/users?offset=0&limit=10", relMap["first"])
	require.Equal(t, "/users?offset=30&limit=10", relMap["next"])
	require.Equal(t, "/users?offset=50&limit=10", relMap["last"])
}

func TestUserService_CRUD_ErrorCases(t *testing.T) {
	mockStore := newUserStoreInterfaceMock(t)
	service := &userService{
		userStore:     mockStore,
		transactioner: &fakeTransactioner{},
	}
	ctx := context.Background()

	t.Run("GetUser_MissingID", func(t *testing.T) {
		_, err := service.GetUser(ctx, "")
		require.NotNil(t, err)
		require.Equal(t, ErrorMissingUserID.Code, err.Code)
	})

	t.Run("GetUser_NotFound", func(t *testing.T) {
		mockStore.On("GetUser", mock.Anything, "u1").Return(User{}, ErrUserNotFound).Once()
		_, err := service.GetUser(ctx, "u1")
		require.NotNil(t, err)
		require.Equal(t, ErrorUserNotFound.Code, err.Code)
	})

	t.Run("DeleteUser_MissingID", func(t *testing.T) {
		err := service.DeleteUser(ctx, "")
		require.NotNil(t, err)
		require.Equal(t, ErrorMissingUserID.Code, err.Code)
	})

	t.Run("DeleteUser_NotFound", func(t *testing.T) {
		mockStore.On("DeleteUser", mock.Anything, "u1").Return(ErrUserNotFound).Once()
		err := service.DeleteUser(ctx, "u1")
		require.NotNil(t, err)
		require.Equal(t, ErrorUserNotFound.Code, err.Code)
	})

	t.Run("CreateUser_MissingType", func(t *testing.T) {
		_, err := service.CreateUser(ctx, &User{ID: "u1"})
		require.NotNil(t, err)
		require.Equal(t, ErrorUserSchemaNotFound.Code, err.Code)
	})

	t.Run("UpdateUser_MissingID", func(t *testing.T) {
		_, err := service.UpdateUser(ctx, "", &User{})
		require.NotNil(t, err)
		require.Equal(t, ErrorMissingUserID.Code, err.Code)
	})
}

func TestUserService_ExtractCredentials_EdgeCases(t *testing.T) {
	mockHash := hashmock.NewHashServiceInterfaceMock(t)
	service := &userService{hashService: mockHash}
	schemaCredFields := []string{"password"}

	t.Run("NilAttributes", func(t *testing.T) {
		creds, err := service.extractCredentials(&User{Attributes: nil}, schemaCredFields)
		require.NoError(t, err)
		require.Empty(t, creds)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		_, err := service.extractCredentials(&User{Attributes: json.RawMessage("invalid")}, schemaCredFields)
		require.Error(t, err)
	})

	t.Run("HashError", func(t *testing.T) {
		mockHash.On("Generate", mock.Anything).Return(hash.Credential{}, errors.New("hash error")).Once()
		attributes := json.RawMessage(`{"password": "pass"}`)
		_, err := service.extractCredentials(&User{Attributes: attributes}, schemaCredFields)
		require.Error(t, err)
	})

	t.Run("NonStringCredential", func(t *testing.T) {
		attributes := json.RawMessage(`{"password": 123}`)
		creds, err := service.extractCredentials(&User{Attributes: attributes}, schemaCredFields)
		require.NoError(t, err)
		require.Empty(t, creds)
	})
}

func TestUserService_GetUsersByPath(t *testing.T) {
	mockOU := oumock.NewOrganizationUnitServiceInterfaceMock(t)
	service := &userService{ouService: mockOU}
	ctx := context.Background()

	mockOU.On("GetOrganizationUnitByPath", "root").Return(oupkg.OrganizationUnit{ID: "ou-1"}, nil).Once()
	mockOU.On("GetOrganizationUnitUsers", "ou-1", 10, 0).Return(&oupkg.UserListResponse{
		TotalResults: 20,
		Users:        []oupkg.User{{ID: "u1"}},
	}, nil).Once()

	resp, err := service.GetUsersByPath(ctx, "root", 10, 0, nil)
	require.Nil(t, err)
	require.Equal(t, 20, resp.TotalResults)
	require.NotEmpty(t, resp.Links)
}

func TestProvider(t *testing.T) {
	svc := &userService{}
	setUserService(svc)
	require.Equal(t, svc, GetUserService())
}

func TestNewFunctions(t *testing.T) {
	svc := newUserService(nil, nil, nil, nil, nil)
	require.NotNil(t, svc)

	handler := newUserHandler(svc)
	require.NotNil(t, handler)
}

func TestUserService_Validation_EdgeCases(t *testing.T) {
	service := &userService{}

	t.Run("ValidateOU_InvalidUUID", func(t *testing.T) {
		err := service.validateOrganizationUnitForUserType(context.TODO(), "customer", "invalid-uuid", nil)
		require.NotNil(t, err)
		require.Equal(t, ErrorInvalidOrganizationUnitID.Code, err.Code)
	})

	t.Run("ValidateOU_EmptyOU", func(t *testing.T) {
		err := service.validateOrganizationUnitForUserType(context.TODO(), "customer", "", nil)
		require.NotNil(t, err)
		require.Equal(t, ErrorInvalidOrganizationUnitID.Code, err.Code)
	})

	t.Run("ValidateUserIDs_Empty", func(t *testing.T) {
		invalid, err := service.ValidateUserIDs(context.Background(), []string{})
		require.Nil(t, err)
		require.Empty(t, invalid)
	})
}

func TestUserService_CredentialValidation_EdgeCases(t *testing.T) {
	service := &userService{}

	t.Run("ValidateCredential_Nil", func(t *testing.T) {
		err := service.validateCredential(nil)
		require.Error(t, err)
	})

	t.Run("ValidateCredential_EmptyValue", func(t *testing.T) {
		err := service.validateCredential(&Credential{Value: ""})
		require.Error(t, err)
	})
}

func TestUserService_HashCredentials_ErrorCase(t *testing.T) {
	mockHash := hashmock.NewHashServiceInterfaceMock(t)
	service := &userService{hashService: mockHash}
	logger := log.GetLogger()

	t.Run("GenerateError", func(t *testing.T) {
		mockHash.On("Generate", mock.Anything).Return(hash.Credential{}, errors.New("hash error")).Once()
		creds := []Credential{{Value: "pass"}}
		_, err := service.hashCredentials(creds, CredentialType("password"), logger)
		require.NotNil(t, err)
		require.Equal(t, ErrorInternalServerError.Code, err.Code)
	})
}

func TestUserService_IdentifyVerify_EdgeCases(t *testing.T) {
	ctx := context.Background()

	t.Run("IdentifyUser_EmptyFilters", func(t *testing.T) {
		service := &userService{}
		_, err := service.IdentifyUser(ctx, nil)
		require.NotNil(t, err)
		require.Equal(t, ErrorInvalidRequestFormat.Code, err.Code)
	})

	t.Run("VerifyUser_MissingID", func(t *testing.T) {
		service := &userService{}
		_, err := service.VerifyUser(ctx, "", map[string]interface{}{"password": "p"})
		require.NotNil(t, err)
		require.Equal(t, ErrorMissingUserID.Code, err.Code)
	})

	t.Run("VerifyUser_NoCredentials", func(t *testing.T) {
		service := &userService{}
		_, err := service.VerifyUser(ctx, "u1", map[string]interface{}{})
		require.NotNil(t, err)
		require.Equal(t, ErrorInvalidRequestFormat.Code, err.Code)
	})

	t.Run("VerifyUser_InvalidCredentialType", func(t *testing.T) {
		mockStore := newUserStoreInterfaceMock(t)
		mockStore.On("GetCredentials", mock.Anything, "u1").
			Return(User{ID: "u1"}, Credentials{CredentialType("password"): []Credential{{Value: "h"}}}, nil).Once()
		service := &userService{userStore: mockStore}
		_, err := service.VerifyUser(ctx, "u1", map[string]interface{}{"invalid": "v"})
		require.NotNil(t, err)
		require.Equal(t, ErrorAuthenticationFailed.Code, err.Code)
	})
}

func TestUserService_MoreErrorCases(t *testing.T) {
	storeMock := &userStoreInterfaceMock{}
	ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
	userSchemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
	txMock := &fakeTransactioner{}
	service := &userService{
		userStore:         storeMock,
		ouService:         ouServiceMock,
		userSchemaService: userSchemaMock,
		transactioner:     txMock,
	}
	ctx := context.Background()

	t.Run("UpdateUser_StoreError", func(t *testing.T) {
		userIn := &User{Type: "customer", OrganizationUnit: testOrgID}
		storeMock.On("UpdateUser", mock.Anything, mock.Anything).Return(errors.New("db error")).Once()

		// Mock all validation steps in the transaction with broad matches to ensure they hit
		userSchemaMock.On("GetCredentialAttributes", mock.Anything, mock.Anything).
			Return([]string{}, (*serviceerror.ServiceError)(nil)).Maybe()
		ouServiceMock.On("IsOrganizationUnitExists", mock.Anything).Return(true, nil).Maybe()
		ouServiceMock.On("IsParent", mock.Anything, mock.Anything).Return(true, nil).Maybe()
		userSchemaMock.On("GetUserSchemaByName", mock.Anything, mock.Anything).
			Return(&userschema.UserSchema{}, nil).Maybe()
		userSchemaMock.On(
			"ValidateUser", mock.Anything, mock.Anything, mock.Anything,
		).Return(true, nil).Maybe()
		userSchemaMock.On(
			"ValidateUserUniqueness", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		).
			Return(true, nil).Maybe()
		storeMock.On("IdentifyUser", mock.Anything, mock.Anything).Return(nil, ErrUserNotFound).Maybe()

		_, err := service.UpdateUser(ctx, "u1", userIn)
		require.NotNil(t, err)
		require.Equal(t, ErrorInternalServerError.Code, err.Code)
	})

	t.Run("DeleteUser_StoreError", func(t *testing.T) {
		storeMock.On("DeleteUser", mock.Anything, "u1").Return(errors.New("db error")).Once()
		err := service.DeleteUser(ctx, "u1")
		require.NotNil(t, err)
		require.Equal(t, ErrorInternalServerError.Code, err.Code)
	})

	t.Run("CreateUserByPath_MissingPath", func(t *testing.T) {
		_, err := service.CreateUserByPath(ctx, "", CreateUserByPathRequest{})
		require.NotNil(t, err)
		require.Equal(t, ErrorInvalidHandlePath.Code, err.Code)
	})
}

func TestUserService_ProcessCredentialType(t *testing.T) {
	t.Run("StringValue_SchemaCredential_HashesAndReturns", func(t *testing.T) {
		mockHash := hashmock.NewHashServiceInterfaceMock(t)
		service := &userService{hashService: mockHash}
		logger := log.GetLogger()

		mockHash.On("Generate", []byte("secret123")).Return(hash.Credential{
			Algorithm: "PBKDF2WithHmacSHA256",
			Hash:      "hashed-value",
			Parameters: hash.CredParameters{
				Iterations: 600000,
				KeySize:    256,
				Salt:       "test-salt",
			},
		}, nil).Once()

		result, svcErr := service.processCredentialType(
			CredentialType("password"),
			json.RawMessage(`"secret123"`),
			logger,
		)

		require.Nil(t, svcErr)
		require.Len(t, result, 1)
		require.Equal(t, "hash", result[0].StorageType)
		require.Equal(t, "hashed-value", result[0].Value)
		require.Equal(t, hash.CredAlgorithm("PBKDF2WithHmacSHA256"), result[0].StorageAlgo)
	})

	t.Run("ArrayValue_SystemManaged_ReturnsRaw", func(t *testing.T) {
		service := &userService{}
		logger := log.GetLogger()

		credJSON := json.RawMessage(
			`[{"value":"passkey-data-1"},{"value":"passkey-data-2"}]`,
		)
		result, svcErr := service.processCredentialType(
			CredentialTypePasskey,
			credJSON,
			logger,
		)

		require.Nil(t, svcErr)
		require.Len(t, result, 2)
		require.Equal(t, "passkey-data-1", result[0].Value)
		require.Equal(t, "passkey-data-2", result[1].Value)
	})

	t.Run("MultipleCredentials_NonSystemManaged_ReturnsError", func(t *testing.T) {
		service := &userService{}
		logger := log.GetLogger()

		credJSON := json.RawMessage(
			`[{"value":"pass1"},{"value":"pass2"}]`,
		)
		result, svcErr := service.processCredentialType(
			CredentialType("password"),
			credJSON,
			logger,
		)

		require.Nil(t, result)
		require.NotNil(t, svcErr)
		require.Equal(t, ErrorInvalidCredential.Code, svcErr.Code)
	})

	t.Run("InvalidJSON_ReturnsError", func(t *testing.T) {
		service := &userService{}
		logger := log.GetLogger()

		result, svcErr := service.processCredentialType(
			CredentialType("password"),
			json.RawMessage(`{invalid`),
			logger,
		)

		require.Nil(t, result)
		require.NotNil(t, svcErr)
		require.Equal(t, ErrorInvalidRequestFormat.Code, svcErr.Code)
	})

	t.Run("EmptyCredentialValue_ReturnsValidationError", func(t *testing.T) {
		service := &userService{}
		logger := log.GetLogger()

		result, svcErr := service.processCredentialType(
			CredentialType("password"),
			json.RawMessage(`""`),
			logger,
		)

		require.Nil(t, result)
		require.NotNil(t, svcErr)
		require.Equal(t, ErrorInvalidCredential.Code, svcErr.Code)
	})

	t.Run("HashError_ReturnsError", func(t *testing.T) {
		mockHash := hashmock.NewHashServiceInterfaceMock(t)
		service := &userService{hashService: mockHash}
		logger := log.GetLogger()

		mockHash.On("Generate", mock.Anything).
			Return(hash.Credential{}, errors.New("hash failure")).Once()

		result, svcErr := service.processCredentialType(
			CredentialType("password"),
			json.RawMessage(`"secret123"`),
			logger,
		)

		require.Nil(t, result)
		require.NotNil(t, svcErr)
		require.Equal(t, ErrorInternalServerError.Code, svcErr.Code)
	})
}

func TestUserService_HashCredentials_Success(t *testing.T) {
	mockHash := hashmock.NewHashServiceInterfaceMock(t)
	service := &userService{hashService: mockHash}
	logger := log.GetLogger()

	t.Run("SingleCredential", func(t *testing.T) {
		mockHash.On("Generate", []byte("mypassword")).Return(hash.Credential{
			Algorithm: "PBKDF2WithHmacSHA256",
			Hash:      "hashed-password",
			Parameters: hash.CredParameters{
				Iterations: 600000,
				KeySize:    256,
				Salt:       "salt1",
			},
		}, nil).Once()

		creds := []Credential{{Value: "mypassword"}}
		result, svcErr := service.hashCredentials(
			creds, CredentialType("password"), logger,
		)

		require.Nil(t, svcErr)
		require.Len(t, result, 1)
		require.Equal(t, "hash", result[0].StorageType)
		require.Equal(t, "hashed-password", result[0].Value)
		require.Equal(t,
			hash.CredAlgorithm("PBKDF2WithHmacSHA256"), result[0].StorageAlgo)
		require.Equal(t, 600000, result[0].StorageAlgoParams.Iterations)
		require.Equal(t, 256, result[0].StorageAlgoParams.KeySize)
		require.Equal(t, "salt1", result[0].StorageAlgoParams.Salt)
	})

	t.Run("MultipleCredentials", func(t *testing.T) {
		mockHash.On("Generate", []byte("cred1")).Return(hash.Credential{
			Algorithm:  "PBKDF2WithHmacSHA256",
			Hash:       "hash1",
			Parameters: hash.CredParameters{Salt: "s1"},
		}, nil).Once()
		mockHash.On("Generate", []byte("cred2")).Return(hash.Credential{
			Algorithm:  "PBKDF2WithHmacSHA256",
			Hash:       "hash2",
			Parameters: hash.CredParameters{Salt: "s2"},
		}, nil).Once()

		creds := []Credential{{Value: "cred1"}, {Value: "cred2"}}
		result, svcErr := service.hashCredentials(
			creds, CredentialType("password"), logger,
		)

		require.Nil(t, svcErr)
		require.Len(t, result, 2)
		require.Equal(t, "hash1", result[0].Value)
		require.Equal(t, "hash2", result[1].Value)
	})
}

func TestUserService_ExtractCredentials_HappyPath(t *testing.T) {
	t.Run("SchemaCredential_ExtractedAndHashed", func(t *testing.T) {
		mockHash := hashmock.NewHashServiceInterfaceMock(t)
		service := &userService{hashService: mockHash}

		mockHash.On("Generate", []byte("secret")).Return(hash.Credential{
			Algorithm:  "PBKDF2WithHmacSHA256",
			Hash:       "hashed-secret",
			Parameters: hash.CredParameters{Salt: "s"},
		}, nil).Once()

		usr := &User{
			Attributes: json.RawMessage(
				`{"email":"a@b.com","password":"secret"}`,
			),
		}
		creds, err := service.extractCredentials(
			usr, []string{"password"},
		)

		require.NoError(t, err)
		require.Len(t, creds, 1)
		require.Contains(t, creds, CredentialType("password"))
		require.Equal(t, "hashed-secret", creds[CredentialType("password")][0].Value)

		// Verify password removed from attributes.
		var attrs map[string]interface{}
		require.NoError(t, json.Unmarshal(usr.Attributes, &attrs))
		require.NotContains(t, attrs, "password")
		require.Contains(t, attrs, "email")
	})

	t.Run("MultipleSchemaCredentials", func(t *testing.T) {
		mockHash := hashmock.NewHashServiceInterfaceMock(t)
		service := &userService{hashService: mockHash}

		mockHash.On("Generate", []byte("pass")).Return(hash.Credential{
			Algorithm:  "PBKDF2WithHmacSHA256",
			Hash:       "h-pass",
			Parameters: hash.CredParameters{Salt: "s1"},
		}, nil).Once()
		mockHash.On("Generate", []byte("1234")).Return(hash.Credential{
			Algorithm:  "PBKDF2WithHmacSHA256",
			Hash:       "h-pin",
			Parameters: hash.CredParameters{Salt: "s2"},
		}, nil).Once()

		usr := &User{
			Attributes: json.RawMessage(
				`{"email":"a@b.com","password":"pass","pin":"1234"}`,
			),
		}
		creds, err := service.extractCredentials(
			usr, []string{"password", "pin"},
		)

		require.NoError(t, err)
		require.Len(t, creds, 2)
		require.Equal(t, "h-pass", creds[CredentialType("password")][0].Value)
		require.Equal(t, "h-pin", creds[CredentialType("pin")][0].Value)

		var attrs map[string]interface{}
		require.NoError(t, json.Unmarshal(usr.Attributes, &attrs))
		require.NotContains(t, attrs, "password")
		require.NotContains(t, attrs, "pin")
		require.Contains(t, attrs, "email")
	})

	t.Run("EmptyCredentialValue_Skipped", func(t *testing.T) {
		mockHash := hashmock.NewHashServiceInterfaceMock(t)
		service := &userService{hashService: mockHash}

		usr := &User{
			Attributes: json.RawMessage(
				`{"email":"a@b.com","password":""}`,
			),
		}
		creds, err := service.extractCredentials(
			usr, []string{"password"},
		)

		require.NoError(t, err)
		require.Empty(t, creds)
	})

	t.Run("SystemManagedCredential_ExtractedRaw", func(t *testing.T) {
		service := &userService{}

		usr := &User{
			Attributes: json.RawMessage(
				`{"email":"a@b.com","passkey":"pk-data"}`,
			),
		}
		creds, err := service.extractCredentials(
			usr, []string{},
		)

		require.NoError(t, err)
		require.Len(t, creds, 1)
		require.Contains(t, creds, CredentialTypePasskey)
		require.Equal(t, "pk-data", creds[CredentialTypePasskey][0].Value)

		var attrs map[string]interface{}
		require.NoError(t, json.Unmarshal(usr.Attributes, &attrs))
		require.NotContains(t, attrs, "passkey")
	})

	t.Run("NoCredentialAttributes_ReturnsEmpty", func(t *testing.T) {
		service := &userService{}

		usr := &User{
			Attributes: json.RawMessage(`{"email":"a@b.com"}`),
		}
		creds, err := service.extractCredentials(usr, []string{})

		require.NoError(t, err)
		require.Empty(t, creds)
	})
}

func TestCredentialType_Methods(t *testing.T) {
	t.Run("IsSystemManaged", func(t *testing.T) {
		require.True(t, CredentialTypePasskey.IsSystemManaged())
		require.False(t, CredentialType("password").IsSystemManaged())
		require.False(t, CredentialType("pin").IsSystemManaged())
		require.False(t, CredentialType("invalid").IsSystemManaged())
	})

	t.Run("String", func(t *testing.T) {
		require.Equal(t, "password", CredentialType("password").String())
		require.Equal(t, "passkey", CredentialTypePasskey.String())
	})
}

func TestUserService_CreateUser_SchemaNotFound(t *testing.T) {
	ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
	ouServiceMock.On("IsOrganizationUnitExists", testOrgID).Return(true, (*serviceerror.ServiceError)(nil)).Once()

	userSchemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
	userSchemaMock.On("GetUserSchemaByName", mock.Anything, testUserType).
		Return(&userschema.UserSchema{OrganizationUnitID: testOrgID}, (*serviceerror.ServiceError)(nil)).Once()
	userSchemaMock.On("ValidateUser", mock.Anything, testUserType, mock.Anything).
		Return(true, (*serviceerror.ServiceError)(nil)).Once()
	userSchemaMock.On("ValidateUserUniqueness", mock.Anything, testUserType, mock.Anything, mock.Anything).
		Return(true, (*serviceerror.ServiceError)(nil)).Once()
	userSchemaMock.On("GetCredentialAttributes", mock.Anything, testUserType).
		Return(nil, &userschema.ErrorUserSchemaNotFound).Once()

	service := &userService{
		userStore:         newUserStoreInterfaceMock(t),
		ouService:         ouServiceMock,
		userSchemaService: userSchemaMock,
		transactioner:     &fakeTransactioner{},
	}

	user := &User{
		Type:             testUserType,
		OrganizationUnit: testOrgID,
		Attributes:       json.RawMessage(`{}`),
	}

	created, svcErr := service.CreateUser(context.Background(), user)
	require.Nil(t, created)
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorUserSchemaNotFound, *svcErr)
}

func TestUserService_CreateUser_GetCredentialAttributesInternalError(t *testing.T) {
	ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)
	ouServiceMock.On("IsOrganizationUnitExists", testOrgID).Return(true, (*serviceerror.ServiceError)(nil)).Once()

	userSchemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
	userSchemaMock.On("GetUserSchemaByName", mock.Anything, testUserType).
		Return(&userschema.UserSchema{OrganizationUnitID: testOrgID}, (*serviceerror.ServiceError)(nil)).Once()
	userSchemaMock.On("ValidateUser", mock.Anything, testUserType, mock.Anything).
		Return(true, (*serviceerror.ServiceError)(nil)).Once()
	userSchemaMock.On("ValidateUserUniqueness", mock.Anything, testUserType, mock.Anything, mock.Anything).
		Return(true, (*serviceerror.ServiceError)(nil)).Once()
	userSchemaMock.On("GetCredentialAttributes", mock.Anything, testUserType).
		Return(nil, &serviceerror.InternalServerError).Once()

	service := &userService{
		userStore:         newUserStoreInterfaceMock(t),
		ouService:         ouServiceMock,
		userSchemaService: userSchemaMock,
		transactioner:     &fakeTransactioner{},
	}

	user := &User{
		Type:             testUserType,
		OrganizationUnit: testOrgID,
		Attributes:       json.RawMessage(`{}`),
	}

	created, svcErr := service.CreateUser(context.Background(), user)
	require.Nil(t, created)
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorInternalServerError.Code, svcErr.Code)
}

func TestUserService_UpdateUser_NilSchemaService(t *testing.T) {
	service := &userService{
		userStore:     newUserStoreInterfaceMock(t),
		transactioner: &fakeTransactioner{},
	}

	user := &User{
		ID:               svcTestUserID1,
		Type:             testUserType,
		OrganizationUnit: testOrgID,
		Attributes:       json.RawMessage(`{"email":"test@example.com"}`),
	}

	resp, svcErr := service.UpdateUser(context.Background(), svcTestUserID1, user)
	require.Nil(t, resp)
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorInternalServerError, *svcErr)
}

func TestUserService_UpdateUser_SchemaNotFound(t *testing.T) {
	userSchemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
	userSchemaMock.On("GetCredentialAttributes", mock.Anything, testUserType).
		Return(nil, &userschema.ErrorUserSchemaNotFound).Once()

	service := &userService{
		userStore:         newUserStoreInterfaceMock(t),
		userSchemaService: userSchemaMock,
		transactioner:     &fakeTransactioner{},
	}

	user := &User{
		ID:               svcTestUserID1,
		Type:             testUserType,
		OrganizationUnit: testOrgID,
		Attributes:       json.RawMessage(`{"email":"test@example.com"}`),
	}

	resp, svcErr := service.UpdateUser(context.Background(), svcTestUserID1, user)
	require.Nil(t, resp)
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorUserSchemaNotFound, *svcErr)
}

func TestUserService_UpdateUserAttributes_NilSchemaService(t *testing.T) {
	storeMock := newUserStoreInterfaceMock(t)
	storeMock.On("GetUser", mock.Anything, svcTestUserID1).
		Return(User{ID: svcTestUserID1, Type: testUserType}, nil).Once()

	service := &userService{
		userStore:     storeMock,
		transactioner: &fakeTransactioner{},
	}

	resp, svcErr := service.UpdateUserAttributes(context.Background(), svcTestUserID1,
		json.RawMessage(`{"email":"a@b.com"}`))
	require.Nil(t, resp)
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorInternalServerError, *svcErr)
}

func TestUserService_UpdateUserAttributes_SchemaNotFound(t *testing.T) {
	storeMock := newUserStoreInterfaceMock(t)
	storeMock.On("GetUser", mock.Anything, svcTestUserID1).
		Return(User{ID: svcTestUserID1, Type: testUserType}, nil).Once()

	schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
	schemaMock.On("GetCredentialAttributes", mock.Anything, testUserType).
		Return(nil, &userschema.ErrorUserSchemaNotFound).Once()

	service := &userService{
		userStore:         storeMock,
		userSchemaService: schemaMock,
		transactioner:     &fakeTransactioner{},
	}

	resp, svcErr := service.UpdateUserAttributes(context.Background(), svcTestUserID1,
		json.RawMessage(`{"email":"a@b.com"}`))
	require.Nil(t, resp)
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorUserSchemaNotFound, *svcErr)
}

func TestUserService_UpdateUserCredentials_NilSchemaService(t *testing.T) {
	userStoreMock := newUserStoreInterfaceMock(t)
	userStoreMock.On("GetCredentials", mock.Anything, svcTestUserID1).
		Return(User{ID: svcTestUserID1, Type: testUserType}, Credentials{}, nil).Once()

	service := &userService{
		userStore:     userStoreMock,
		transactioner: &fakeTransactioner{},
	}

	svcErr := service.UpdateUserCredentials(context.Background(), svcTestUserID1,
		json.RawMessage(`{"password":"newpassword"}`))
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorInternalServerError, *svcErr)
}

func TestUserService_UpdateUserCredentials_SchemaNotFound(t *testing.T) {
	userStoreMock := newUserStoreInterfaceMock(t)
	userStoreMock.On("GetCredentials", mock.Anything, svcTestUserID1).
		Return(User{ID: svcTestUserID1, Type: testUserType}, Credentials{}, nil).Once()

	schemaMock := userschemamock.NewUserSchemaServiceInterfaceMock(t)
	schemaMock.On("GetCredentialAttributes", mock.Anything, testUserType).
		Return(nil, &userschema.ErrorUserSchemaNotFound).Once()

	service := &userService{
		userStore:         userStoreMock,
		userSchemaService: schemaMock,
		transactioner:     &fakeTransactioner{},
	}

	svcErr := service.UpdateUserCredentials(context.Background(), svcTestUserID1,
		json.RawMessage(`{"password":"newpassword"}`))
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorUserSchemaNotFound, *svcErr)
}
