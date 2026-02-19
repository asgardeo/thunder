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

package usertype

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/tests/mocks/oumock"
)

const (
	testOUID1 = "00000000-0000-0000-0000-000000000001"
	testOUID2 = "00000000-0000-0000-0000-000000000002"
	testOUID3 = "00000000-0000-0000-0000-000000000003"
)

func TestCreateUserTypeReturnsErrorWhenOrganizationUnitMissing(t *testing.T) {
	// Initialize ThunderRuntime with default config
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	require.NoError(t, err)
	defer config.ResetThunderRuntime()

	storeMock := newUserTypeStoreInterfaceMock(t)
	ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)

	ouID := testOUID1
	ouServiceMock.On("IsOrganizationUnitExists", ouID).Return(false, (*serviceerror.ServiceError)(nil)).Once()

	service := &userTypeService{
		userTypeStore: storeMock,
		ouService:     ouServiceMock,
	}

	request := CreateUserTypeRequest{
		Name:               "test-schema",
		OrganizationUnitID: ouID,
		Schema:             json.RawMessage(`{"email":{"type":"string"}}`),
	}

	createdSchema, svcErr := service.CreateUserType(request)

	require.Nil(t, createdSchema)
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorInvalidUserTypeRequest.Code, svcErr.Code)
	require.Contains(t, svcErr.ErrorDescription, "organization unit id does not exist")
}

func TestCreateUserTypeReturnsInternalErrorWhenOUValidationFails(t *testing.T) {
	// Initialize ThunderRuntime with default config
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	require.NoError(t, err)
	defer config.ResetThunderRuntime()

	storeMock := newUserTypeStoreInterfaceMock(t)
	ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)

	ouID := testOUID2
	ouServiceMock.
		On("IsOrganizationUnitExists", ouID).
		Return(false, &serviceerror.ServiceError{Code: "OUS-5000"}).
		Once()

	service := &userTypeService{
		userTypeStore: storeMock,
		ouService:     ouServiceMock,
	}

	request := CreateUserTypeRequest{
		Name:               "test-schema",
		OrganizationUnitID: ouID,
		Schema:             json.RawMessage(`{"email":{"type":"string"}}`),
	}

	createdSchema, svcErr := service.CreateUserType(request)

	require.Nil(t, createdSchema)
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorInternalServerError, *svcErr)
}

func TestUpdateUserTypeReturnsErrorWhenOrganizationUnitMissing(t *testing.T) {
	// Initialize ThunderRuntime with default config
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	require.NoError(t, err)
	defer config.ResetThunderRuntime()

	storeMock := newUserTypeStoreInterfaceMock(t)
	ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)

	ouID := testOUID3
	ouServiceMock.On("IsOrganizationUnitExists", ouID).Return(false, (*serviceerror.ServiceError)(nil)).Once()

	service := &userTypeService{
		userTypeStore: storeMock,
		ouService:     ouServiceMock,
	}

	request := UpdateUserTypeRequest{
		Name:               "test-schema",
		OrganizationUnitID: ouID,
		Schema:             json.RawMessage(`{"email":{"type":"string"}}`),
	}

	updatedSchema, svcErr := service.UpdateUserType("schema-id", request)

	require.Nil(t, updatedSchema)
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorInvalidUserTypeRequest.Code, svcErr.Code)
}

func TestGetUserTypeByNameReturnsSchema(t *testing.T) {
	storeMock := newUserTypeStoreInterfaceMock(t)
	expectedSchema := UserType{
		ID:   "schema-id",
		Name: "employee",
	}
	storeMock.
		On("GetUserTypeByName", "employee").
		Return(expectedSchema, nil).
		Once()

	service := &userTypeService{
		userTypeStore: storeMock,
	}

	userType, svcErr := service.GetUserTypeByName("employee")

	require.Nil(t, svcErr)
	require.NotNil(t, userType)
	require.Equal(t, &expectedSchema, userType)
}

func TestGetUserTypeByNameReturnsNotFound(t *testing.T) {
	storeMock := newUserTypeStoreInterfaceMock(t)
	storeMock.
		On("GetUserTypeByName", "employee").
		Return(UserType{}, ErrUserTypeNotFound).
		Once()

	service := &userTypeService{
		userTypeStore: storeMock,
	}

	userType, svcErr := service.GetUserTypeByName("employee")

	require.Nil(t, userType)
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorUserTypeNotFound, *svcErr)
}

func TestGetUserTypeByNameReturnsInternalErrorOnStoreFailure(t *testing.T) {
	storeMock := newUserTypeStoreInterfaceMock(t)
	storeMock.
		On("GetUserTypeByName", "employee").
		Return(UserType{}, errors.New("db failure")).
		Once()

	service := &userTypeService{
		userTypeStore: storeMock,
	}

	userType, svcErr := service.GetUserTypeByName("employee")

	require.Nil(t, userType)
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorInternalServerError, *svcErr)
}

func TestGetUserTypeByNameRequiresName(t *testing.T) {
	storeMock := newUserTypeStoreInterfaceMock(t)

	service := &userTypeService{
		userTypeStore: storeMock,
	}

	userType, svcErr := service.GetUserTypeByName("")

	require.Nil(t, userType)
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorInvalidUserTypeRequest.Code, svcErr.Code)
}

func TestValidateUserReturnsTrueWhenValidationPasses(t *testing.T) {
	storeMock := newUserTypeStoreInterfaceMock(t)
	storeMock.
		On("GetUserTypeByName", "employee").
		Return(UserType{
			Name:   "employee",
			Schema: json.RawMessage(`{"email":{"type":"string","required":true}}`),
		}, nil).
		Once()

	service := &userTypeService{
		userTypeStore: storeMock,
	}

	ok, svcErr := service.ValidateUser("employee", json.RawMessage(`{"email":"employee@example.com"}`))

	require.True(t, ok)
	require.Nil(t, svcErr)
}

func TestValidateUserReturnsInternalErrorWhenSchemaLoadFails(t *testing.T) {
	storeMock := newUserTypeStoreInterfaceMock(t)
	storeMock.
		On("GetUserTypeByName", "employee").
		Return(UserType{}, errors.New("db failure")).
		Once()

	service := &userTypeService{
		userTypeStore: storeMock,
	}

	ok, svcErr := service.ValidateUser("employee", json.RawMessage(`{}`))

	require.False(t, ok)
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorInternalServerError, *svcErr)
}

func TestValidateUserUniquenessReturnsTrueWhenNoConflicts(t *testing.T) {
	storeMock := newUserTypeStoreInterfaceMock(t)
	storeMock.
		On("GetUserTypeByName", "employee").
		Return(UserType{
			Name:   "employee",
			Schema: json.RawMessage(`{"email":{"type":"string","unique":true}}`),
		}, nil).
		Once()

	service := &userTypeService{
		userTypeStore: storeMock,
	}

	ok, svcErr := service.ValidateUserUniqueness(
		"employee",
		json.RawMessage(`{"email":"unique@example.com"}`),
		func(filters map[string]interface{}) (*string, error) {
			require.Equal(t, map[string]interface{}{"email": "unique@example.com"}, filters)
			return nil, nil
		},
	)

	require.True(t, ok)
	require.Nil(t, svcErr)
}

func TestValidateUserReturnsSchemaNotFoundWhenSchemaMissing(t *testing.T) {
	storeMock := newUserTypeStoreInterfaceMock(t)
	storeMock.
		On("GetUserTypeByName", "employee").
		Return(UserType{}, ErrUserTypeNotFound).
		Once()

	service := &userTypeService{
		userTypeStore: storeMock,
	}

	ok, svcErr := service.ValidateUser("employee", json.RawMessage(`{"email":"employee@example.com"}`))

	require.False(t, ok)
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorUserTypeNotFound, *svcErr)
}

func TestValidateUserUniquenessReturnsSchemaNotFoundWhenSchemaMissing(t *testing.T) {
	storeMock := newUserTypeStoreInterfaceMock(t)
	storeMock.
		On("GetUserTypeByName", "employee").
		Return(UserType{}, ErrUserTypeNotFound).
		Once()

	service := &userTypeService{
		userTypeStore: storeMock,
	}

	ok, svcErr := service.ValidateUserUniqueness(
		"employee",
		json.RawMessage(`{}`),
		func(map[string]interface{}) (*string, error) { return nil, nil },
	)

	require.False(t, ok)
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorUserTypeNotFound, *svcErr)
}

func TestValidateUserUniquenessReturnsInternalErrorWhenSchemaLoadFails(t *testing.T) {
	storeMock := newUserTypeStoreInterfaceMock(t)
	storeMock.
		On("GetUserTypeByName", "employee").
		Return(UserType{}, errors.New("db failure")).
		Once()

	service := &userTypeService{
		userTypeStore: storeMock,
	}

	ok, svcErr := service.ValidateUserUniqueness(
		"employee",
		json.RawMessage(`{}`),
		func(map[string]interface{}) (*string, error) { return nil, nil },
	)

	require.False(t, ok)
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorInternalServerError, *svcErr)
}

func TestValidateUserTypeDefinitionSuccess(t *testing.T) {
	validOUID := testOUID1
	validSchema := json.RawMessage(`{"email":{"type":"string","required":true}}`)

	schema := UserType{
		Name:               "test-schema",
		OrganizationUnitID: validOUID,
		Schema:             validSchema,
	}

	err := validateUserTypeDefinition(schema)

	require.Nil(t, err)
}

func TestValidateUserTypeDefinitionReturnsErrorWhenNameIsEmpty(t *testing.T) {
	validOUID := testOUID1
	validSchema := json.RawMessage(`{"email":{"type":"string"}}`)

	schema := UserType{
		Name:               "",
		OrganizationUnitID: validOUID,
		Schema:             validSchema,
	}

	err := validateUserTypeDefinition(schema)

	require.NotNil(t, err)
	require.Equal(t, ErrorInvalidUserTypeRequest.Code, err.Code)
	require.Contains(t, err.ErrorDescription, "user type name must not be empty")
}

func TestValidateUserTypeDefinitionReturnsErrorWhenOrganizationUnitIDIsEmpty(t *testing.T) {
	validSchema := json.RawMessage(`{"email":{"type":"string"}}`)

	schema := UserType{
		Name:               "test-schema",
		OrganizationUnitID: "",
		Schema:             validSchema,
	}

	err := validateUserTypeDefinition(schema)

	require.NotNil(t, err)
	require.Equal(t, ErrorInvalidUserTypeRequest.Code, err.Code)
	require.Contains(t, err.ErrorDescription, "organization unit id must not be empty")
}

func TestValidateUserTypeDefinitionReturnsErrorWhenOrganizationUnitIDIsNotUUID(t *testing.T) {
	validSchema := json.RawMessage(`{"email":{"type":"string"}}`)

	schema := UserType{
		Name:               "test-schema",
		OrganizationUnitID: "not-a-uuid",
		Schema:             validSchema,
	}

	err := validateUserTypeDefinition(schema)

	require.NotNil(t, err)
	require.Equal(t, ErrorInvalidUserTypeRequest.Code, err.Code)
	require.Contains(t, err.ErrorDescription, "organization unit id is not a valid UUID")
}

func TestValidateUserTypeDefinitionReturnsErrorWhenSchemaIsEmpty(t *testing.T) {
	validOUID := testOUID1

	schema := UserType{
		Name:               "test-schema",
		OrganizationUnitID: validOUID,
		Schema:             json.RawMessage{},
	}

	err := validateUserTypeDefinition(schema)

	require.NotNil(t, err)
	require.Equal(t, ErrorInvalidUserTypeRequest.Code, err.Code)
	require.Contains(t, err.ErrorDescription, "schema definition must not be empty")
}

func TestValidateUserTypeDefinitionReturnsErrorWhenSchemaIsNil(t *testing.T) {
	validOUID := testOUID1

	schema := UserType{
		Name:               "test-schema",
		OrganizationUnitID: validOUID,
		Schema:             nil,
	}

	err := validateUserTypeDefinition(schema)

	require.NotNil(t, err)
	require.Equal(t, ErrorInvalidUserTypeRequest.Code, err.Code)
	require.Contains(t, err.ErrorDescription, "schema definition must not be empty")
}

func TestValidateUserTypeDefinitionReturnsErrorWhenSchemaCompilationFails(t *testing.T) {
	validOUID := testOUID1
	invalidSchema := json.RawMessage(`{"email":"invalid"}`)

	schema := UserType{
		Name:               "test-schema",
		OrganizationUnitID: validOUID,
		Schema:             invalidSchema,
	}

	err := validateUserTypeDefinition(schema)

	require.NotNil(t, err)
	require.Equal(t, ErrorInvalidUserTypeRequest.Code, err.Code)
	require.Contains(t, err.ErrorDescription, "property definition must be an object")
}

func TestValidateUserTypeDefinitionReturnsErrorForInvalidJSON(t *testing.T) {
	validOUID := testOUID1
	invalidSchema := json.RawMessage(`{invalid json}`)

	schema := UserType{
		Name:               "test-schema",
		OrganizationUnitID: validOUID,
		Schema:             invalidSchema,
	}

	err := validateUserTypeDefinition(schema)

	require.NotNil(t, err)
	require.Equal(t, ErrorInvalidUserTypeRequest.Code, err.Code)
}

func TestValidateUserTypeDefinitionReturnsErrorForEmptySchemaObject(t *testing.T) {
	validOUID := testOUID1
	emptySchema := json.RawMessage(`{}`)

	schema := UserType{
		Name:               "test-schema",
		OrganizationUnitID: validOUID,
		Schema:             emptySchema,
	}

	err := validateUserTypeDefinition(schema)

	require.NotNil(t, err)
	require.Equal(t, ErrorInvalidUserTypeRequest.Code, err.Code)
	require.Contains(t, err.ErrorDescription, "schema cannot be empty")
}

func TestValidateUserTypeDefinitionWithComplexSchema(t *testing.T) {
	validOUID := testOUID1
	complexSchema := json.RawMessage(`{
		"email": {
			"type": "string",
			"required": true,
			"unique": true,
			"pattern": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
		},
		"age": {
			"type": "number",
			"required": false
		},
		"isActive": {
			"type": "boolean",
			"required": true
		},
		"address": {
			"type": "object",
			"properties": {
				"street": {"type": "string"},
				"city": {"type": "string"}
			}
		},
		"tags": {
			"type": "array",
			"items": {"type": "string"}
		}
	}`)

	schema := UserType{
		Name:               "complex-schema",
		OrganizationUnitID: validOUID,
		Schema:             complexSchema,
	}

	err := validateUserTypeDefinition(schema)

	require.Nil(t, err)
}

func TestValidateUserTypeDefinitionReturnsErrorForMissingTypeField(t *testing.T) {
	validOUID := testOUID1
	schemaWithoutType := json.RawMessage(`{"email":{"required":true}}`)

	schema := UserType{
		Name:               "test-schema",
		OrganizationUnitID: validOUID,
		Schema:             schemaWithoutType,
	}

	err := validateUserTypeDefinition(schema)

	require.NotNil(t, err)
	require.Equal(t, ErrorInvalidUserTypeRequest.Code, err.Code)
	require.Contains(t, err.ErrorDescription, "missing required 'type' field")
}

func TestValidateUserTypeDefinitionReturnsErrorForInvalidType(t *testing.T) {
	validOUID := testOUID1
	schemaWithInvalidType := json.RawMessage(`{"email":{"type":"invalid-type"}}`)

	schema := UserType{
		Name:               "test-schema",
		OrganizationUnitID: validOUID,
		Schema:             schemaWithInvalidType,
	}

	err := validateUserTypeDefinition(schema)

	require.NotNil(t, err)
	require.Equal(t, ErrorInvalidUserTypeRequest.Code, err.Code)
}

func TestValidateUserTypeDefinitionWithMultipleValidationErrors(t *testing.T) {
	testCases := []struct {
		name          string
		schema        UserType
		expectedError string
	}{
		{
			name: "Empty name and empty OU ID",
			schema: UserType{
				Name:               "",
				OrganizationUnitID: "",
				Schema:             json.RawMessage(`{"email":{"type":"string"}}`),
			},
			expectedError: "user type name must not be empty",
		},
		{
			name: "Valid name but invalid OU ID format",
			schema: UserType{
				Name:               "test",
				OrganizationUnitID: "123",
				Schema:             json.RawMessage(`{"email":{"type":"string"}}`),
			},
			expectedError: "organization unit id is not a valid UUID",
		},
		{
			name: "Valid OU ID but empty schema",
			schema: UserType{
				Name:               "test",
				OrganizationUnitID: testOUID1,
				Schema:             json.RawMessage{},
			},
			expectedError: "schema definition must not be empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateUserTypeDefinition(tc.schema)

			require.NotNil(t, err)
			require.Equal(t, ErrorInvalidUserTypeRequest.Code, err.Code)
			require.Contains(t, err.ErrorDescription, tc.expectedError)
		})
	}
}
