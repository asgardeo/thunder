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

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/sysauthz"
	"github.com/asgardeo/thunder/internal/userschema/model"
	"github.com/asgardeo/thunder/tests/mocks/consentmock"
	"github.com/asgardeo/thunder/tests/mocks/oumock"
	"github.com/asgardeo/thunder/tests/mocks/sysauthzmock"
)

const (
	testOUID1 = "00000000-0000-0000-0000-000000000001"
	testOUID2 = "00000000-0000-0000-0000-000000000002"
	testOUID3 = "00000000-0000-0000-0000-000000000003"
)

// newAllowAllAuthz returns a mock SystemAuthorizationServiceInterface that allows all actions.
func newAllowAllAuthz(t interface {
	mock.TestingT
	Cleanup(func())
}) *sysauthzmock.SystemAuthorizationServiceInterfaceMock {
	authzMock := sysauthzmock.NewSystemAuthorizationServiceInterfaceMock(t)
	authzMock.On("IsActionAllowed", mock.Anything, mock.Anything, mock.Anything).
		Return(true, nil).Maybe()
	authzMock.On("GetAccessibleResources", mock.Anything, mock.Anything, mock.Anything).
		Return(&sysauthz.AccessibleResources{AllAllowed: true}, nil).Maybe()
	return authzMock
}

// newConsentServiceMockEnabled creates a new consent service mock with IsEnabled returning true.
func newConsentServiceMockEnabled(t interface {
	mock.TestingT
	Cleanup(func())
}) *consentmock.ConsentServiceInterfaceMock {
	consentMock := consentmock.NewConsentServiceInterfaceMock(t)
	consentMock.On("IsEnabled").Return(true)
	return consentMock
}

// newConsentServiceMockDisabled creates a new consent service mock with IsEnabled returning false.
func newConsentServiceMockDisabled(t interface {
	mock.TestingT
	Cleanup(func())
}) *consentmock.ConsentServiceInterfaceMock {
	consentMock := consentmock.NewConsentServiceInterfaceMock(t)
	consentMock.On("IsEnabled").Return(false)
	return consentMock
}

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
	ouServiceMock.On("IsOrganizationUnitExists", mock.Anything, ouID).
		Return(false, (*serviceerror.ServiceError)(nil)).
		Once()

	service := &userSchemaService{
		userSchemaStore: storeMock,
		ouService:       ouServiceMock,
		transactioner:   &mockTransactioner{},
	}

	request := CreateUserSchemaRequest{
		Name:   "test-schema",
		OUID:   ouID,
		Schema: json.RawMessage(`{"email":{"type":"string"}}`),
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
		On("IsOrganizationUnitExists", mock.Anything, ouID).
		Return(false, &serviceerror.ServiceError{Code: "OUS-5000"}).
		Once()

	service := &userSchemaService{
		userSchemaStore: storeMock,
		ouService:       ouServiceMock,
		transactioner:   &mockTransactioner{},
	}

	request := CreateUserSchemaRequest{
		Name:   "test-schema",
		OUID:   ouID,
		Schema: json.RawMessage(`{"email":{"type":"string"}}`),
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
	storeMock.On("IsUserSchemaDeclarative", "schema-id").Return(false).Once()
	ouServiceMock.On("IsOrganizationUnitExists", mock.Anything, ouID).
		Return(false, (*serviceerror.ServiceError)(nil)).
		Once()

	service := &userSchemaService{
		userSchemaStore: storeMock,
		ouService:       ouServiceMock,
		transactioner:   &mockTransactioner{},
	}

	request := UpdateUserSchemaRequest{
		Name:   "test-schema",
		OUID:   ouID,
		Schema: json.RawMessage(`{"email":{"type":"string"}}`),
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
		authzService:    newAllowAllAuthz(t),
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
		Name:   "test-schema",
		OUID:   validOUID,
		Schema: validSchema,
	}

	err := validateUserSchemaDefinition(schema)

	require.Nil(t, err)
}

func TestValidateUserSchemaDefinitionReturnsErrorWhenNameIsEmpty(t *testing.T) {
	validOUID := testOUID1
	validSchema := json.RawMessage(`{"email":{"type":"string"}}`)

	schema := UserSchema{
		Name:   "",
		OUID:   validOUID,
		Schema: validSchema,
	}

	err := validateUserSchemaDefinition(schema)

	require.NotNil(t, err)
	require.Equal(t, ErrorInvalidUserSchemaRequest.Code, err.Code)
	require.Contains(t, err.ErrorDescription, "user schema name must not be empty")
}

func TestValidateUserSchemaDefinitionReturnsErrorWhenOUIDIsEmpty(t *testing.T) {
	validSchema := json.RawMessage(`{"email":{"type":"string"}}`)

	schema := UserSchema{
		Name:   "test-schema",
		OUID:   "",
		Schema: validSchema,
	}

	err := validateUserSchemaDefinition(schema)

	require.NotNil(t, err)
	require.Equal(t, ErrorInvalidUserSchemaRequest.Code, err.Code)
	require.Contains(t, err.ErrorDescription, "organization unit id must not be empty")
}

func TestValidateUserSchemaDefinitionAllowsNonUUIDOUID(t *testing.T) {
	validSchema := json.RawMessage(`{"email":{"type":"string"}}`)

	schema := UserSchema{
		Name:   "test-schema",
		OUID:   "not-a-uuid",
		Schema: validSchema,
	}

	err := validateUserSchemaDefinition(schema)

	require.Nil(t, err)
}

func TestValidateUserSchemaDefinitionReturnsErrorWhenSchemaIsEmpty(t *testing.T) {
	validOUID := testOUID1

	schema := UserSchema{
		Name:   "test-schema",
		OUID:   validOUID,
		Schema: json.RawMessage{},
	}

	err := validateUserSchemaDefinition(schema)

	require.NotNil(t, err)
	require.Equal(t, ErrorInvalidUserSchemaRequest.Code, err.Code)
	require.Contains(t, err.ErrorDescription, "schema definition must not be empty")
}

func TestValidateUserSchemaDefinitionReturnsErrorWhenSchemaIsNil(t *testing.T) {
	validOUID := testOUID1

	schema := UserSchema{
		Name:   "test-schema",
		OUID:   validOUID,
		Schema: nil,
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
		Name:   "test-schema",
		OUID:   validOUID,
		Schema: invalidSchema,
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
		Name:   "test-schema",
		OUID:   validOUID,
		Schema: invalidSchema,
	}

	err := validateUserSchemaDefinition(schema)

	require.NotNil(t, err)
	require.Equal(t, ErrorInvalidUserSchemaRequest.Code, err.Code)
}

func TestValidateUserSchemaDefinitionReturnsErrorForEmptySchemaObject(t *testing.T) {
	validOUID := testOUID1
	emptySchema := json.RawMessage(`{}`)

	schema := UserSchema{
		Name:   "test-schema",
		OUID:   validOUID,
		Schema: emptySchema,
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
		Name:   "complex-schema",
		OUID:   validOUID,
		Schema: complexSchema,
	}

	err := validateUserSchemaDefinition(schema)

	require.Nil(t, err)
}

func TestValidateUserSchemaDefinitionReturnsErrorForMissingTypeField(t *testing.T) {
	validOUID := testOUID1
	schemaWithoutType := json.RawMessage(`{"email":{"required":true}}`)

	schema := UserSchema{
		Name:   "test-schema",
		OUID:   validOUID,
		Schema: schemaWithoutType,
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
		Name:   "test-schema",
		OUID:   validOUID,
		Schema: schemaWithInvalidType,
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
				Name:   "",
				OUID:   "",
				Schema: json.RawMessage(`{"email":{"type":"string"}}`),
			},
			expectedError: "user schema name must not be empty",
		},
		{
			name: "Non-UUID OU ID still validates schema payload",
			schema: UserSchema{
				Name:   "test",
				OUID:   "123",
				Schema: json.RawMessage{},
			},
			expectedError: "schema definition must not be empty",
		},
		{
			name: "Valid OU ID but empty schema",
			schema: UserSchema{
				Name:   "test",
				OUID:   testOUID1,
				Schema: json.RawMessage{},
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

func TestValidateUserSchemaDefinitionWithValidDisplayAttribute(t *testing.T) {
	schema := UserSchema{
		Name:             "test-schema",
		OUID:             testOUID1,
		SystemAttributes: &SystemAttributes{Display: "email"},
		Schema:           json.RawMessage(`{"email":{"type":"string"}}`),
	}

	err := validateUserSchemaDefinition(schema)

	require.Nil(t, err)
}

func TestValidateUserSchemaDefinitionRejectsNonExistentDisplayAttribute(t *testing.T) {
	schema := UserSchema{
		Name:             "test-schema",
		OUID:             testOUID1,
		SystemAttributes: &SystemAttributes{Display: "unknown"},
		Schema:           json.RawMessage(`{"email":{"type":"string"}}`),
	}

	err := validateUserSchemaDefinition(schema)

	require.NotNil(t, err)
	require.Equal(t, ErrorInvalidDisplayAttribute.Code, err.Code)
}

func TestValidateUserSchemaDefinitionRejectsNonDisplayableDisplayAttribute(t *testing.T) {
	schema := UserSchema{
		Name:             "test-schema",
		OUID:             testOUID1,
		SystemAttributes: &SystemAttributes{Display: "active"},
		Schema:           json.RawMessage(`{"active":{"type":"boolean"}}`),
	}

	err := validateUserSchemaDefinition(schema)

	require.NotNil(t, err)
	require.Equal(t, ErrorNonDisplayableAttribute.Code, err.Code)
}

func TestValidateUserSchemaDefinitionRejectsCredentialDisplayAttribute(t *testing.T) {
	schema := UserSchema{
		Name:             "test-schema",
		OUID:             testOUID1,
		SystemAttributes: &SystemAttributes{Display: "password"},
		Schema:           json.RawMessage(`{"password":{"type":"string","credential":true}}`),
	}

	err := validateUserSchemaDefinition(schema)

	require.NotNil(t, err)
	require.Equal(t, ErrorCredentialDisplayAttribute.Code, err.Code)
}

func TestValidateUserSchemaDefinitionWithNilSystemAttributes(t *testing.T) {
	schema := UserSchema{
		Name:   "test-schema",
		OUID:   testOUID1,
		Schema: json.RawMessage(`{"email":{"type":"string"}}`),
	}

	err := validateUserSchemaDefinition(schema)

	require.Nil(t, err)
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

type GetUniqueAttributesTestSuite struct {
	suite.Suite
}

func TestGetUniqueAttributesTestSuite(t *testing.T) {
	suite.Run(t, new(GetUniqueAttributesTestSuite))
}

func (s *GetUniqueAttributesTestSuite) TestReturnsUniqueFieldNames() {
	storeMock := newUserSchemaStoreInterfaceMock(s.T())
	storeMock.
		On("GetUserSchemaByName", context.Background(), "customer").
		Return(UserSchema{
			Schema: json.RawMessage(
				`{"email":{"type":"string","unique":true},` +
					`"username":{"type":"string","unique":true},` +
					`"given_name":{"type":"string"}}`,
			),
		}, nil).
		Once()

	service := &userSchemaService{
		userSchemaStore: storeMock,
		transactioner:   &mockTransactioner{},
	}

	fields, svcErr := service.GetUniqueAttributes(context.Background(), "customer")

	s.Require().Nil(svcErr)
	sort.Strings(fields)
	s.Require().Equal([]string{"email", "username"}, fields)
}

func (s *GetUniqueAttributesTestSuite) TestNoUniqueAttributes_ReturnsEmpty() {
	storeMock := newUserSchemaStoreInterfaceMock(s.T())
	storeMock.
		On("GetUserSchemaByName", context.Background(), "customer").
		Return(UserSchema{
			Schema: json.RawMessage(`{"given_name":{"type":"string"},"age":{"type":"number"}}`),
		}, nil).
		Once()

	service := &userSchemaService{
		userSchemaStore: storeMock,
		transactioner:   &mockTransactioner{},
	}

	fields, svcErr := service.GetUniqueAttributes(context.Background(), "customer")

	s.Require().Nil(svcErr)
	s.Require().Empty(fields)
}

func (s *GetUniqueAttributesTestSuite) TestSchemaNotFound_ReturnsError() {
	storeMock := newUserSchemaStoreInterfaceMock(s.T())
	storeMock.
		On("GetUserSchemaByName", context.Background(), "unknown").
		Return(UserSchema{}, ErrUserSchemaNotFound).
		Once()

	service := &userSchemaService{
		userSchemaStore: storeMock,
		transactioner:   &mockTransactioner{},
	}

	fields, svcErr := service.GetUniqueAttributes(context.Background(), "unknown")

	s.Require().Nil(fields)
	s.Require().NotNil(svcErr)
	s.Require().Equal(ErrorUserSchemaNotFound, *svcErr)
}

func (s *GetUniqueAttributesTestSuite) TestEmptyUserType_ReturnsError() {
	storeMock := newUserSchemaStoreInterfaceMock(s.T())

	service := &userSchemaService{
		userSchemaStore: storeMock,
		transactioner:   &mockTransactioner{},
	}

	fields, svcErr := service.GetUniqueAttributes(context.Background(), "")

	s.Require().Nil(fields)
	s.Require().NotNil(svcErr)
	s.Require().Equal(ErrorUserSchemaNotFound, *svcErr)
}

// ----- DeleteUserSchema Tests -----

func TestDeleteUserSchema(t *testing.T) {
	tests := []struct {
		name           string
		schemaID       string
		schema         json.RawMessage
		consentService *consentmock.ConsentServiceInterfaceMock
	}{
		{
			name:     "succeeds when attribute extraction fails but consent is enabled",
			schemaID: "schema-123",
			// Use invalid JSON to cause extractAttributeNames to fail
			schema:         json.RawMessage(`{invalid json}`),
			consentService: newConsentServiceMockEnabled(t),
		},
		{
			name:           "succeeds when consent is disabled",
			schemaID:       "schema-456",
			schema:         json.RawMessage(`{"email":{"type":"string"}}`),
			consentService: newConsentServiceMockDisabled(t),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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
			storeMock.On("GetUserSchemaByID", mock.Anything, tc.schemaID).Return(UserSchema{
				ID:     tc.schemaID,
				OUID:   testOUID1,
				Schema: tc.schema,
			}, nil).Once()
			storeMock.On("IsUserSchemaDeclarative", tc.schemaID).Return(false).Once()
			storeMock.On("DeleteUserSchemaByID", mock.Anything, tc.schemaID).Return(nil).Once()

			service := &userSchemaService{
				userSchemaStore: storeMock,
				transactioner:   &mockTransactioner{},
				consentService:  tc.consentService,
				authzService:    newAllowAllAuthz(t),
			}

			svcErr := service.DeleteUserSchema(context.Background(), tc.schemaID)

			require.Nil(t, svcErr)
			storeMock.AssertExpectations(t)
		})
	}
}

func TestValidateDisplayAttribute_NilSystemAttributes(t *testing.T) {
	compiled, err := model.CompileUserSchema(json.RawMessage(`{"email":{"type":"string"}}`))
	require.NoError(t, err)

	svcErr := validateDisplayAttribute(compiled, "")
	require.Nil(t, svcErr)
}

func TestValidateDisplayAttribute_EmptyDisplay(t *testing.T) {
	compiled, err := model.CompileUserSchema(json.RawMessage(`{"email":{"type":"string"}}`))
	require.NoError(t, err)

	svcErr := validateDisplayAttribute(compiled, "")
	require.Nil(t, svcErr)
}

func TestValidateDisplayAttribute_ValidStringAttribute(t *testing.T) {
	compiled, err := model.CompileUserSchema(json.RawMessage(`{
		"email":{"type":"string"},
		"password":{"type":"string","credential":true}
	}`))
	require.NoError(t, err)

	svcErr := validateDisplayAttribute(compiled, "email")
	require.Nil(t, svcErr)
}

func TestValidateDisplayAttribute_ValidNumberAttribute(t *testing.T) {
	compiled, err := model.CompileUserSchema(json.RawMessage(`{
		"email":{"type":"string"},
		"age":{"type":"number"}
	}`))
	require.NoError(t, err)

	svcErr := validateDisplayAttribute(compiled, "age")
	require.Nil(t, svcErr)
}

func TestValidateDisplayAttribute_BooleanAttributeRejected(t *testing.T) {
	compiled, err := model.CompileUserSchema(json.RawMessage(`{
		"email":{"type":"string"},
		"active":{"type":"boolean"}
	}`))
	require.NoError(t, err)

	svcErr := validateDisplayAttribute(compiled, "active")
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorNonDisplayableAttribute.Code, svcErr.Code)
}

func TestValidateDisplayAttribute_ObjectAttributeRejected(t *testing.T) {
	compiled, err := model.CompileUserSchema(json.RawMessage(`{
		"address":{"type":"object","properties":{"city":{"type":"string"}}}
	}`))
	require.NoError(t, err)

	svcErr := validateDisplayAttribute(compiled, "address")
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorNonDisplayableAttribute.Code, svcErr.Code)
}

func TestValidateDisplayAttribute_ArrayAttributeRejected(t *testing.T) {
	compiled, err := model.CompileUserSchema(json.RawMessage(`{
		"tags":{"type":"array","items":{"type":"string"}}
	}`))
	require.NoError(t, err)

	svcErr := validateDisplayAttribute(compiled, "tags")
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorNonDisplayableAttribute.Code, svcErr.Code)
}

func TestValidateDisplayAttribute_CredentialAttributeRejected(t *testing.T) {
	compiled, err := model.CompileUserSchema(json.RawMessage(`{
		"password":{"type":"string","credential":true}
	}`))
	require.NoError(t, err)

	svcErr := validateDisplayAttribute(compiled, "password")
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorCredentialDisplayAttribute.Code, svcErr.Code)
}

func TestValidateDisplayAttribute_NonExistentAttribute(t *testing.T) {
	compiled, err := model.CompileUserSchema(json.RawMessage(`{
		"email":{"type":"string"}
	}`))
	require.NoError(t, err)

	svcErr := validateDisplayAttribute(compiled, "username")
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorInvalidDisplayAttribute.Code, svcErr.Code)
}

func TestCreateUserSchemaReturnsErrorForInvalidDisplayAttribute(t *testing.T) {
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

	service := &userSchemaService{
		userSchemaStore: storeMock,
		transactioner:   &mockTransactioner{},
	}

	request := CreateUserSchemaRequest{
		Name:             "test-schema",
		OUID:             testOUID1,
		Schema:           json.RawMessage(`{"email":{"type":"string"}}`),
		SystemAttributes: &SystemAttributes{Display: "nonexistent"},
	}

	createdSchema, svcErr := service.CreateUserSchema(context.Background(), request)

	require.Nil(t, createdSchema)
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorInvalidDisplayAttribute.Code, svcErr.Code)
}

func TestUpdateUserSchemaReturnsErrorForInvalidDisplayAttribute(t *testing.T) {
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
	storeMock.On("IsUserSchemaDeclarative", "schema-id").Return(false).Once()

	service := &userSchemaService{
		userSchemaStore: storeMock,
		transactioner:   &mockTransactioner{},
	}

	request := UpdateUserSchemaRequest{
		Name:             "test-schema",
		OUID:             testOUID1,
		Schema:           json.RawMessage(`{"email":{"type":"string"}}`),
		SystemAttributes: &SystemAttributes{Display: "nonexistent"},
	}

	updatedSchema, svcErr := service.UpdateUserSchema(context.Background(), "schema-id", request)

	require.Nil(t, updatedSchema)
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorInvalidDisplayAttribute.Code, svcErr.Code)
}

func TestValidateDisplayAttribute_DottedPath_ValidNestedString(t *testing.T) {
	compiled, err := model.CompileUserSchema(json.RawMessage(`{
		"address":{"type":"object","properties":{"city":{"type":"string"}}}
	}`))
	require.NoError(t, err)

	svcErr := validateDisplayAttribute(compiled, "address.city")
	require.Nil(t, svcErr)
}

func TestValidateDisplayAttribute_DottedPath_NestedObjectRejected(t *testing.T) {
	compiled, err := model.CompileUserSchema(json.RawMessage(`{
		"profile":{"type":"object","properties":{
			"address":{"type":"object","properties":{"city":{"type":"string"}}}
		}}
	}`))
	require.NoError(t, err)

	svcErr := validateDisplayAttribute(compiled, "profile.address")
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorNonDisplayableAttribute.Code, svcErr.Code)
}

func TestValidateDisplayAttribute_DottedPath_NestedCredentialRejected(t *testing.T) {
	compiled, err := model.CompileUserSchema(json.RawMessage(`{
		"auth":{"type":"object","properties":{
			"password":{"type":"string","credential":true}
		}}
	}`))
	require.NoError(t, err)

	svcErr := validateDisplayAttribute(compiled, "auth.password")
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorCredentialDisplayAttribute.Code, svcErr.Code)
}

func TestValidateDisplayAttribute_DottedPath_NonExistentNested(t *testing.T) {
	compiled, err := model.CompileUserSchema(json.RawMessage(`{
		"address":{"type":"object","properties":{"city":{"type":"string"}}}
	}`))
	require.NoError(t, err)

	svcErr := validateDisplayAttribute(compiled, "address.zip")
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorInvalidDisplayAttribute.Code, svcErr.Code)
}

func TestValidateDisplayAttribute_DottedPath_TraverseIntoNonObject(t *testing.T) {
	compiled, err := model.CompileUserSchema(json.RawMessage(`{
		"email":{"type":"string"}
	}`))
	require.NoError(t, err)

	svcErr := validateDisplayAttribute(compiled, "email.domain")
	require.NotNil(t, svcErr)
	require.Equal(t, ErrorInvalidDisplayAttribute.Code, svcErr.Code)
}

func TestValidateDisplayAttribute_DottedPath_DeeplyNestedValid(t *testing.T) {
	compiled, err := model.CompileUserSchema(json.RawMessage(`{
		"profile":{"type":"object","properties":{
			"name":{"type":"object","properties":{
				"first":{"type":"string"}
			}}
		}}
	}`))
	require.NoError(t, err)

	svcErr := validateDisplayAttribute(compiled, "profile.name.first")
	require.Nil(t, svcErr)
}

// GetDisplayAttributesByNames tests

type GetDisplayAttributesByNamesTestSuite struct {
	suite.Suite
}

func TestGetDisplayAttributesByNamesTestSuite(t *testing.T) {
	suite.Run(t, new(GetDisplayAttributesByNamesTestSuite))
}

func (s *GetDisplayAttributesByNamesTestSuite) TestReturnsDisplayAttributes() {
	storeMock := newUserSchemaStoreInterfaceMock(s.T())
	expected := map[string]string{"SchemaA": "email", "SchemaB": "given_name"}
	storeMock.
		On("GetDisplayAttributesByNames", mock.Anything, []string{"SchemaA", "SchemaB"}).
		Return(expected, nil).
		Once()

	service := &userSchemaService{userSchemaStore: storeMock}

	result, svcErr := service.GetDisplayAttributesByNames(
		context.Background(), []string{"SchemaA", "SchemaB"})

	s.Require().Nil(svcErr)
	s.Require().Equal(expected, result)
}

func (s *GetDisplayAttributesByNamesTestSuite) TestEmptyInput_ReturnsEmptyMap() {
	service := &userSchemaService{}

	result, svcErr := service.GetDisplayAttributesByNames(context.Background(), []string{})

	s.Require().Nil(svcErr)
	s.Require().Empty(result)
}

func (s *GetDisplayAttributesByNamesTestSuite) TestStoreError_ReturnsServerError() {
	storeMock := newUserSchemaStoreInterfaceMock(s.T())
	storeMock.
		On("GetDisplayAttributesByNames", mock.Anything, []string{"SchemaA"}).
		Return(map[string]string(nil), errors.New("db error")).
		Once()

	service := &userSchemaService{userSchemaStore: storeMock}

	_, svcErr := service.GetDisplayAttributesByNames(context.Background(), []string{"SchemaA"})

	s.Require().NotNil(svcErr)
	s.Require().Equal(ErrorInternalServerError, *svcErr)
}
