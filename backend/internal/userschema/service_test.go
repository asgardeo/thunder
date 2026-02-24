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

package userschema

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/tests/mocks/oumock"
)

const (
	testOUID1 = "00000000-0000-0000-0000-000000000001"
	testOUID2 = "00000000-0000-0000-0000-000000000002"
	testOUID3 = "00000000-0000-0000-0000-000000000003"
)

func TestCreateUserSchemaReturnsErrorWhenOrganizationUnitMissing(t *testing.T) {
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

	storeMock := newUserSchemaStoreInterfaceMock(t)
	ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)

	ouID := testOUID1
	ouServiceMock.On("IsOrganizationUnitExists", ouID).Return(false, (*serviceerror.ServiceError)(nil)).Once()

	service := &userSchemaService{
		userSchemaStore: storeMock,
		ouService:       ouServiceMock,
		transactioner:   &mockTransactioner{},
	}

	request := CreateUserSchemaRequest{
		Name:               "test-schema",
		OrganizationUnitID: ouID,
		Schema:             json.RawMessage(`{"email":{"type":"string"}}`),
	}

	createdSchema, svcErr := service.CreateUserSchema(context.Background(), request)

	require.Nil(t, createdSchema)
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorInvalidUserSchemaRequest.Code, svcErr.Code)
	require.Contains(t, svcErr.ErrorDescription, "organization unit id does not exist")
}

func TestCreateUserSchemaReturnsInternalErrorWhenOUValidationFails(t *testing.T) {
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

	storeMock := newUserSchemaStoreInterfaceMock(t)
	ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)

	ouID := testOUID2
	ouServiceMock.
		On("IsOrganizationUnitExists", ouID).
		Return(false, &serviceerror.ServiceError{Code: "OUS-5000"}).
		Once()

	service := &userSchemaService{
		userSchemaStore: storeMock,
		ouService:       ouServiceMock,
		transactioner:   &mockTransactioner{},
	}

	request := CreateUserSchemaRequest{
		Name:               "test-schema",
		OrganizationUnitID: ouID,
		Schema:             json.RawMessage(`{"email":{"type":"string"}}`),
	}

	createdSchema, svcErr := service.CreateUserSchema(context.Background(), request)

	require.Nil(t, createdSchema)
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorInternalServerError, *svcErr)
}

func TestUpdateUserSchemaReturnsErrorWhenOrganizationUnitMissing(t *testing.T) {
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

	storeMock := newUserSchemaStoreInterfaceMock(t)
	ouServiceMock := oumock.NewOrganizationUnitServiceInterfaceMock(t)

	ouID := testOUID3
	ouServiceMock.On("IsOrganizationUnitExists", ouID).Return(false, (*serviceerror.ServiceError)(nil)).Once()

	service := &userSchemaService{
		userSchemaStore: storeMock,
		ouService:       ouServiceMock,
		transactioner:   &mockTransactioner{},
	}

	request := UpdateUserSchemaRequest{
		Name:               "test-schema",
		OrganizationUnitID: ouID,
		Schema:             json.RawMessage(`{"email":{"type":"string"}}`),
	}

	updatedSchema, svcErr := service.UpdateUserSchema(context.Background(), "schema-id", request)

	require.Nil(t, updatedSchema)
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorInvalidUserSchemaRequest.Code, svcErr.Code)
}

func TestGetUserSchemaByNameReturnsSchema(t *testing.T) {
	storeMock := newUserSchemaStoreInterfaceMock(t)
	expectedSchema := UserSchema{
		ID:   "schema-id",
		Name: "employee",
	}
	storeMock.
		On("GetUserSchemaByName", context.Background(), "employee").
		Return(expectedSchema, nil).
		Once()

	service := &userSchemaService{
		userSchemaStore: storeMock,
		transactioner:   &mockTransactioner{},
	}

	userSchema, svcErr := service.GetUserSchemaByName(context.Background(), "employee")

	require.Nil(t, svcErr)
	require.NotNil(t, userSchema)
	require.Equal(t, &expectedSchema, userSchema)
}

func TestGetUserSchemaByNameReturnsNotFound(t *testing.T) {
	storeMock := newUserSchemaStoreInterfaceMock(t)
	storeMock.
		On("GetUserSchemaByName", context.Background(), "employee").
		Return(UserSchema{}, ErrUserSchemaNotFound).
		Once()

	service := &userSchemaService{
		userSchemaStore: storeMock,
		transactioner:   &mockTransactioner{},
	}

	userSchema, svcErr := service.GetUserSchemaByName(context.Background(), "employee")

	require.Nil(t, userSchema)
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorUserSchemaNotFound, *svcErr)
}

func TestGetUserSchemaByNameReturnsInternalErrorOnStoreFailure(t *testing.T) {
	storeMock := newUserSchemaStoreInterfaceMock(t)
	storeMock.
		On("GetUserSchemaByName", context.Background(), "employee").
		Return(UserSchema{}, errors.New("db failure")).
		Once()

	service := &userSchemaService{
		userSchemaStore: storeMock,
		transactioner:   &mockTransactioner{},
	}

	userSchema, svcErr := service.GetUserSchemaByName(context.Background(), "employee")

	require.Nil(t, userSchema)
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorInternalServerError, *svcErr)
}

func TestGetUserSchemaByNameRequiresName(t *testing.T) {
	storeMock := newUserSchemaStoreInterfaceMock(t)

	service := &userSchemaService{
		userSchemaStore: storeMock,
		transactioner:   &mockTransactioner{},
	}

	userSchema, svcErr := service.GetUserSchemaByName(context.Background(), "")

	require.Nil(t, userSchema)
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorInvalidUserSchemaRequest.Code, svcErr.Code)
}

func TestValidateUserReturnsTrueWhenValidationPasses(t *testing.T) {
	storeMock := newUserSchemaStoreInterfaceMock(t)
	storeMock.
		On("GetUserSchemaByName", context.Background(), "employee").
		Return(UserSchema{
			Name:   "employee",
			Schema: json.RawMessage(`{"email":{"type":"string","required":true}}`),
		}, nil).
		Once()

	service := &userSchemaService{
		userSchemaStore: storeMock,
		transactioner:   &mockTransactioner{},
	}

	ok, svcErr := service.ValidateUser(
		context.Background(),
		"employee",
		json.RawMessage(`{"email":"employee@example.com"}`),
	)

	require.True(t, ok)
	require.Nil(t, svcErr)
}

func TestValidateUserReturnsInternalErrorWhenSchemaLoadFails(t *testing.T) {
	storeMock := newUserSchemaStoreInterfaceMock(t)
	storeMock.
		On("GetUserSchemaByName", context.Background(), "employee").
		Return(UserSchema{}, errors.New("db failure")).
		Once()

	service := &userSchemaService{
		userSchemaStore: storeMock,
		transactioner:   &mockTransactioner{},
	}

	ok, svcErr := service.ValidateUser(context.Background(), "employee", json.RawMessage(`{}`))

	require.False(t, ok)
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorInternalServerError, *svcErr)
}

func TestValidateUserUniquenessReturnsTrueWhenNoConflicts(t *testing.T) {
	storeMock := newUserSchemaStoreInterfaceMock(t)
	storeMock.
		On("GetUserSchemaByName", context.Background(), "employee").
		Return(UserSchema{
			Name:   "employee",
			Schema: json.RawMessage(`{"email":{"type":"string","unique":true}}`),
		}, nil).
		Once()

	service := &userSchemaService{
		userSchemaStore: storeMock,
		transactioner:   &mockTransactioner{},
	}

	ok, svcErr := service.ValidateUserUniqueness(
		context.Background(),
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
	storeMock := newUserSchemaStoreInterfaceMock(t)
	storeMock.
		On("GetUserSchemaByName", context.Background(), "employee").
		Return(UserSchema{}, ErrUserSchemaNotFound).
		Once()

	service := &userSchemaService{
		userSchemaStore: storeMock,
		transactioner:   &mockTransactioner{},
	}

	ok, svcErr := service.ValidateUser(
		context.Background(),
		"employee",
		json.RawMessage(`{"email":"employee@example.com"}`),
	)

	require.False(t, ok)
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorUserSchemaNotFound, *svcErr)
}

func TestValidateUserUniquenessReturnsSchemaNotFoundWhenSchemaMissing(t *testing.T) {
	storeMock := newUserSchemaStoreInterfaceMock(t)
	storeMock.
		On("GetUserSchemaByName", context.Background(), "employee").
		Return(UserSchema{}, ErrUserSchemaNotFound).
		Once()

	service := &userSchemaService{
		userSchemaStore: storeMock,
		transactioner:   &mockTransactioner{},
	}

	ok, svcErr := service.ValidateUserUniqueness(
		context.Background(),
		"employee",
		json.RawMessage(`{}`),
		func(map[string]interface{}) (*string, error) { return nil, nil },
	)

	require.False(t, ok)
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorUserSchemaNotFound, *svcErr)
}

func TestValidateUserUniquenessReturnsInternalErrorWhenSchemaLoadFails(t *testing.T) {
	storeMock := newUserSchemaStoreInterfaceMock(t)
	storeMock.
		On("GetUserSchemaByName", context.Background(), "employee").
		Return(UserSchema{}, errors.New("db failure")).
		Once()

	service := &userSchemaService{
		userSchemaStore: storeMock,
		transactioner:   &mockTransactioner{},
	}

	ok, svcErr := service.ValidateUserUniqueness(
		context.Background(),
		"employee",
		json.RawMessage(`{}`),
		func(map[string]interface{}) (*string, error) { return nil, nil },
	)

	require.False(t, ok)
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorInternalServerError, *svcErr)
}

func TestValidateUserSchemaDefinitionSuccess(t *testing.T) {
	validOUID := testOUID1
	validSchema := json.RawMessage(`{"email":{"type":"string","required":true}}`)

	schema := UserSchema{
		Name:               "test-schema",
		OrganizationUnitID: validOUID,
		Schema:             validSchema,
	}

	err := validateUserSchemaDefinition(schema)

	require.Nil(t, err)
}

func TestValidateUserSchemaDefinitionReturnsErrorWhenNameIsEmpty(t *testing.T) {
	validOUID := testOUID1
	validSchema := json.RawMessage(`{"email":{"type":"string"}}`)

	schema := UserSchema{
		Name:               "",
		OrganizationUnitID: validOUID,
		Schema:             validSchema,
	}

	err := validateUserSchemaDefinition(schema)

	require.NotNil(t, err)
	require.Equal(t, ErrorInvalidUserSchemaRequest.Code, err.Code)
	require.Contains(t, err.ErrorDescription, "user schema name must not be empty")
}

func TestValidateUserSchemaDefinitionReturnsErrorWhenOrganizationUnitIDIsEmpty(t *testing.T) {
	validSchema := json.RawMessage(`{"email":{"type":"string"}}`)

	schema := UserSchema{
		Name:               "test-schema",
		OrganizationUnitID: "",
		Schema:             validSchema,
	}

	err := validateUserSchemaDefinition(schema)

	require.NotNil(t, err)
	require.Equal(t, ErrorInvalidUserSchemaRequest.Code, err.Code)
	require.Contains(t, err.ErrorDescription, "organization unit id must not be empty")
}

func TestValidateUserSchemaDefinitionReturnsErrorWhenOrganizationUnitIDIsNotUUID(t *testing.T) {
	validSchema := json.RawMessage(`{"email":{"type":"string"}}`)

	schema := UserSchema{
		Name:               "test-schema",
		OrganizationUnitID: "not-a-uuid",
		Schema:             validSchema,
	}

	err := validateUserSchemaDefinition(schema)

	require.NotNil(t, err)
	require.Equal(t, ErrorInvalidUserSchemaRequest.Code, err.Code)
	require.Contains(t, err.ErrorDescription, "organization unit id is not a valid UUID")
}

func TestValidateUserSchemaDefinitionReturnsErrorWhenSchemaIsEmpty(t *testing.T) {
	validOUID := testOUID1

	schema := UserSchema{
		Name:               "test-schema",
		OrganizationUnitID: validOUID,
		Schema:             json.RawMessage{},
	}

	err := validateUserSchemaDefinition(schema)

	require.NotNil(t, err)
	require.Equal(t, ErrorInvalidUserSchemaRequest.Code, err.Code)
	require.Contains(t, err.ErrorDescription, "schema definition must not be empty")
}

func TestValidateUserSchemaDefinitionReturnsErrorWhenSchemaIsNil(t *testing.T) {
	validOUID := testOUID1

	schema := UserSchema{
		Name:               "test-schema",
		OrganizationUnitID: validOUID,
		Schema:             nil,
	}

	err := validateUserSchemaDefinition(schema)

	require.NotNil(t, err)
	require.Equal(t, ErrorInvalidUserSchemaRequest.Code, err.Code)
	require.Contains(t, err.ErrorDescription, "schema definition must not be empty")
}

func TestValidateUserSchemaDefinitionReturnsErrorWhenSchemaCompilationFails(t *testing.T) {
	validOUID := testOUID1
	invalidSchema := json.RawMessage(`{"email":"invalid"}`)

	schema := UserSchema{
		Name:               "test-schema",
		OrganizationUnitID: validOUID,
		Schema:             invalidSchema,
	}

	err := validateUserSchemaDefinition(schema)

	require.NotNil(t, err)
	require.Equal(t, ErrorInvalidUserSchemaRequest.Code, err.Code)
	require.Contains(t, err.ErrorDescription, "property definition must be an object")
}

func TestValidateUserSchemaDefinitionReturnsErrorForInvalidJSON(t *testing.T) {
	validOUID := testOUID1
	invalidSchema := json.RawMessage(`{invalid json}`)

	schema := UserSchema{
		Name:               "test-schema",
		OrganizationUnitID: validOUID,
		Schema:             invalidSchema,
	}

	err := validateUserSchemaDefinition(schema)

	require.NotNil(t, err)
	require.Equal(t, ErrorInvalidUserSchemaRequest.Code, err.Code)
}

func TestValidateUserSchemaDefinitionReturnsErrorForEmptySchemaObject(t *testing.T) {
	validOUID := testOUID1
	emptySchema := json.RawMessage(`{}`)

	schema := UserSchema{
		Name:               "test-schema",
		OrganizationUnitID: validOUID,
		Schema:             emptySchema,
	}

	err := validateUserSchemaDefinition(schema)

	require.NotNil(t, err)
	require.Equal(t, ErrorInvalidUserSchemaRequest.Code, err.Code)
	require.Contains(t, err.ErrorDescription, "schema cannot be empty")
}

func TestValidateUserSchemaDefinitionWithComplexSchema(t *testing.T) {
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

	schema := UserSchema{
		Name:               "complex-schema",
		OrganizationUnitID: validOUID,
		Schema:             complexSchema,
	}

	err := validateUserSchemaDefinition(schema)

	require.Nil(t, err)
}

func TestValidateUserSchemaDefinitionReturnsErrorForMissingTypeField(t *testing.T) {
	validOUID := testOUID1
	schemaWithoutType := json.RawMessage(`{"email":{"required":true}}`)

	schema := UserSchema{
		Name:               "test-schema",
		OrganizationUnitID: validOUID,
		Schema:             schemaWithoutType,
	}

	err := validateUserSchemaDefinition(schema)

	require.NotNil(t, err)
	require.Equal(t, ErrorInvalidUserSchemaRequest.Code, err.Code)
	require.Contains(t, err.ErrorDescription, "missing required 'type' field")
}

func TestValidateUserSchemaDefinitionReturnsErrorForInvalidType(t *testing.T) {
	validOUID := testOUID1
	schemaWithInvalidType := json.RawMessage(`{"email":{"type":"invalid-type"}}`)

	schema := UserSchema{
		Name:               "test-schema",
		OrganizationUnitID: validOUID,
		Schema:             schemaWithInvalidType,
	}

	err := validateUserSchemaDefinition(schema)

	require.NotNil(t, err)
	require.Equal(t, ErrorInvalidUserSchemaRequest.Code, err.Code)
}

func TestValidateUserSchemaDefinitionWithMultipleValidationErrors(t *testing.T) {
	testCases := []struct {
		name          string
		schema        UserSchema
		expectedError string
	}{
		{
			name: "Empty name and empty OU ID",
			schema: UserSchema{
				Name:               "",
				OrganizationUnitID: "",
				Schema:             json.RawMessage(`{"email":{"type":"string"}}`),
			},
			expectedError: "user schema name must not be empty",
		},
		{
			name: "Valid name but invalid OU ID format",
			schema: UserSchema{
				Name:               "test",
				OrganizationUnitID: "123",
				Schema:             json.RawMessage(`{"email":{"type":"string"}}`),
			},
			expectedError: "organization unit id is not a valid UUID",
		},
		{
			name: "Valid OU ID but empty schema",
			schema: UserSchema{
				Name:               "test",
				OrganizationUnitID: testOUID1,
				Schema:             json.RawMessage{},
			},
			expectedError: "schema definition must not be empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateUserSchemaDefinition(tc.schema)

			require.NotNil(t, err)
			require.Equal(t, ErrorInvalidUserSchemaRequest.Code, err.Code)
			require.Contains(t, err.ErrorDescription, tc.expectedError)
		})
	}
}

type GetCredentialAttributesTestSuite struct {
	suite.Suite
}

func TestGetCredentialAttributesTestSuite(t *testing.T) {
	suite.Run(t, new(GetCredentialAttributesTestSuite))
}

func (s *GetCredentialAttributesTestSuite) TestReturnsCredentialFieldNames() {
	storeMock := newUserSchemaStoreInterfaceMock(s.T())
	storeMock.
		On("GetUserSchemaByName", context.Background(), "customer").
		Return(UserSchema{
			Schema: json.RawMessage(
				`{"password":{"type":"string","credential":true},` +
					`"apiKey":{"type":"string","credential":true},` +
					`"email":{"type":"string","unique":true}}`,
			),
		}, nil).
		Once()

	service := &userSchemaService{
		userSchemaStore: storeMock,
		transactioner:   &mockTransactioner{},
	}

	fields, svcErr := service.GetCredentialAttributes(
		context.Background(), "customer",
	)

	s.Require().Nil(svcErr)
	sort.Strings(fields)
	s.Require().Equal([]string{"apiKey", "password"}, fields)
}

func (s *GetCredentialAttributesTestSuite) TestNoCredentials_ReturnsEmpty() {
	storeMock := newUserSchemaStoreInterfaceMock(s.T())
	storeMock.
		On("GetUserSchemaByName", context.Background(), "customer").
		Return(UserSchema{
			Schema: json.RawMessage(
				`{"email":{"type":"string"},"age":{"type":"number"}}`,
			),
		}, nil).
		Once()

	service := &userSchemaService{
		userSchemaStore: storeMock,
		transactioner:   &mockTransactioner{},
	}

	fields, svcErr := service.GetCredentialAttributes(
		context.Background(), "customer",
	)

	s.Require().Nil(svcErr)
	s.Require().Empty(fields)
}

func (s *GetCredentialAttributesTestSuite) TestSchemaNotFound_ReturnsError() {
	storeMock := newUserSchemaStoreInterfaceMock(s.T())
	storeMock.
		On("GetUserSchemaByName", context.Background(), "unknown").
		Return(UserSchema{}, ErrUserSchemaNotFound).
		Once()

	service := &userSchemaService{
		userSchemaStore: storeMock,
		transactioner:   &mockTransactioner{},
	}

	fields, svcErr := service.GetCredentialAttributes(
		context.Background(), "unknown",
	)

	s.Require().Nil(fields)
	s.Require().NotNil(svcErr)
	s.Require().Equal(ErrorUserSchemaNotFound, *svcErr)
}

func (s *GetCredentialAttributesTestSuite) TestEmptyUserType_ReturnsError() {
	storeMock := newUserSchemaStoreInterfaceMock(s.T())

	service := &userSchemaService{
		userSchemaStore: storeMock,
		transactioner:   &mockTransactioner{},
	}

	fields, svcErr := service.GetCredentialAttributes(
		context.Background(), "",
	)

	s.Require().Nil(fields)
	s.Require().NotNil(svcErr)
	s.Require().Equal(ErrorUserSchemaNotFound, *svcErr)
}

func (s *GetCredentialAttributesTestSuite) TestStoreError_ReturnsInternalError() {
	storeMock := newUserSchemaStoreInterfaceMock(s.T())
	storeMock.
		On("GetUserSchemaByName", context.Background(), "customer").
		Return(UserSchema{}, errors.New("db failure")).
		Once()

	service := &userSchemaService{
		userSchemaStore: storeMock,
		transactioner:   &mockTransactioner{},
	}

	fields, svcErr := service.GetCredentialAttributes(
		context.Background(), "customer",
	)

	s.Require().Nil(fields)
	s.Require().NotNil(svcErr)
	s.Require().Equal(ErrorInternalServerError, *svcErr)
}
