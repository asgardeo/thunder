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

package userschema

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/consent"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/tests/mocks/consentmock"
	"github.com/asgardeo/thunder/tests/mocks/oumock"
)

type UserSchemaServiceConsentTestSuite struct {
	suite.Suite
}

func TestUserSchemaServiceConsentTestSuite(t *testing.T) {
	suite.Run(t, new(UserSchemaServiceConsentTestSuite))
}

// newTestSchemaServiceWithConsent creates a userSchemaService with only the consentService field set.
func newTestSchemaServiceWithConsent(consentSvc consent.ConsentServiceInterface) *userSchemaService {
	return &userSchemaService{consentService: consentSvc}
}

// ----- extractAttributeNames -----

func (s *UserSchemaServiceConsentTestSuite) TestExtractAttributeNames_EmptySchema() {
	names, svcErr := extractAttributeNames(json.RawMessage{})

	s.Nil(svcErr)
	s.Nil(names)
}

func (s *UserSchemaServiceConsentTestSuite) TestExtractAttributeNames_ValidSchema() {
	schema := json.RawMessage(`{"email":{},"phone":{}}`)

	names, svcErr := extractAttributeNames(schema)

	s.Nil(svcErr)
	s.Len(names, 2)
	s.ElementsMatch([]string{"email", "phone"}, names)
}

func (s *UserSchemaServiceConsentTestSuite) TestExtractAttributeNames_InvalidJSON() {
	schema := json.RawMessage(`not-valid-json`)

	names, svcErr := extractAttributeNames(schema)

	s.Nil(names)
	s.NotNil(svcErr)
}

// ----- extractAttributeNamesAsMap -----

func (s *UserSchemaServiceConsentTestSuite) TestExtractAttributeNamesAsMap() {
	result, svcErr := extractAttributeNamesAsMap(json.RawMessage{})

	s.Nil(svcErr)
	s.Empty(result)
}

func (s *UserSchemaServiceConsentTestSuite) TestExtractAttributeNamesAsMap_ValidSchema() {
	schema := json.RawMessage(`{"email":{},"phone":{}}`)

	result, svcErr := extractAttributeNamesAsMap(schema)

	s.Nil(svcErr)
	s.Len(result, 2)
	s.True(result["email"])
	s.True(result["phone"])
}

func (s *UserSchemaServiceConsentTestSuite) TestExtractAttributeNamesAsMap_InvalidJSON() {
	result, svcErr := extractAttributeNamesAsMap(json.RawMessage(`{bad json}`))

	s.Nil(result)
	s.NotNil(svcErr)
}

// ----- wrapConsentServiceError (userschema) -----

func (s *UserSchemaServiceConsentTestSuite) TestWrapConsentServiceError_Nil() {
	result := wrapConsentServiceError(nil, nil)

	s.Nil(result)
}

func (s *UserSchemaServiceConsentTestSuite) TestWrapConsentServiceError_ClientError() {
	clientErr := &serviceerror.I18nServiceError{
		Type: serviceerror.ClientErrorType,
		Code: "CSE-1007",
	}

	result := wrapConsentServiceError(clientErr, log.GetLogger())

	s.NotNil(result)
	s.Equal(serviceerror.ClientErrorType, result.Type)
	s.Equal(ErrorConsentSyncFailed.Code, result.Code)
}

func (s *UserSchemaServiceConsentTestSuite) TestWrapConsentServiceError_ServerError() {
	serverErr := &serviceerror.I18nServiceError{
		Type: serviceerror.ServerErrorType,
		Code: "CSE-500",
	}

	result := wrapConsentServiceError(serverErr, log.GetLogger())

	s.NotNil(result)
	s.Equal(serviceerror.ServerErrorType, result.Type)
}

// ----- createMissingConsentElements -----

func (s *UserSchemaServiceConsentTestSuite) TestCreateMissingConsentElements_EmptyNames() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestSchemaServiceWithConsent(cMock)

	result := svc.createMissingConsentElements(context.Background(), "default", []string{}, log.GetLogger())

	s.Nil(result)
}

func (s *UserSchemaServiceConsentTestSuite) TestCreateMissingConsentElements_AllExist() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestSchemaServiceWithConsent(cMock)

	names := []string{"email", "phone"}
	cMock.EXPECT().ValidateConsentElements(mock.Anything, "default", names).
		Return([]string{"email", "phone"}, nil)

	result := svc.createMissingConsentElements(context.Background(), "default", names, log.GetLogger())

	s.Nil(result)
}

func (s *UserSchemaServiceConsentTestSuite) TestCreateMissingConsentElements_SomeMissing() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestSchemaServiceWithConsent(cMock)

	names := []string{"email", "phone"}
	cMock.EXPECT().ValidateConsentElements(mock.Anything, "default", names).
		Return([]string{"email"}, nil)

	expectedInput := []consent.ConsentElementInput{
		{Name: "phone", Namespace: consent.NamespaceAttribute},
	}
	cMock.EXPECT().CreateConsentElements(mock.Anything, "default", expectedInput).
		Return([]consent.ConsentElement{{ID: "e1", Name: "phone"}}, nil)

	result := svc.createMissingConsentElements(context.Background(), "default", names, log.GetLogger())

	s.Nil(result)
}

func (s *UserSchemaServiceConsentTestSuite) TestCreateMissingConsentElements_ValidateError() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestSchemaServiceWithConsent(cMock)

	names := []string{"email"}
	cMock.EXPECT().ValidateConsentElements(mock.Anything, "default", names).
		Return(nil, &serviceerror.InternalServerErrorWithI18n)

	result := svc.createMissingConsentElements(context.Background(), "default", names, log.GetLogger())

	s.NotNil(result)
	s.Equal(serviceerror.ServerErrorType, result.Type)
}

func (s *UserSchemaServiceConsentTestSuite) TestCreateMissingConsentElements_CreateError() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestSchemaServiceWithConsent(cMock)

	names := []string{"email"}
	cMock.EXPECT().ValidateConsentElements(mock.Anything, "default", names).
		Return([]string{}, nil)
	cMock.EXPECT().CreateConsentElements(mock.Anything, "default", mock.Anything).
		Return(nil, &serviceerror.InternalServerErrorWithI18n)

	result := svc.createMissingConsentElements(context.Background(), "default", names, log.GetLogger())

	s.NotNil(result)
}

// ----- deleteConsentElements -----

func (s *UserSchemaServiceConsentTestSuite) TestDeleteConsentElements_EmptyList() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestSchemaServiceWithConsent(cMock)

	result := svc.deleteConsentElements(context.Background(), []string{}, log.GetLogger())

	s.Nil(result)
}

func (s *UserSchemaServiceConsentTestSuite) TestDeleteConsentElements() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestSchemaServiceWithConsent(cMock)

	cMock.EXPECT().ListConsentElements(mock.Anything, "default", consent.NamespaceAttribute, "email").
		Return([]consent.ConsentElement{{ID: "e1", Name: "email"}}, nil)
	cMock.EXPECT().DeleteConsentElement(mock.Anything, "default", "e1").
		Return((*serviceerror.I18nServiceError)(nil))

	result := svc.deleteConsentElements(context.Background(), []string{"email"}, log.GetLogger())

	s.Nil(result)
}

func (s *UserSchemaServiceConsentTestSuite) TestDeleteConsentElements_NoExistingElements() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestSchemaServiceWithConsent(cMock)

	cMock.EXPECT().ListConsentElements(mock.Anything, "default", consent.NamespaceAttribute, "email").
		Return([]consent.ConsentElement{}, nil)

	result := svc.deleteConsentElements(context.Background(), []string{"email"}, log.GetLogger())

	s.Nil(result)
}

func (s *UserSchemaServiceConsentTestSuite) TestDeleteConsentElements_AssociatedPurposeError() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestSchemaServiceWithConsent(cMock)

	// "email" can't be deleted due to associated purpose — should warn and continue
	cMock.EXPECT().ListConsentElements(mock.Anything, "default", consent.NamespaceAttribute, "email").
		Return([]consent.ConsentElement{{ID: "e1", Name: "email"}}, nil)
	cMock.EXPECT().DeleteConsentElement(mock.Anything, "default", "e1").
		Return(&consent.ErrorDeletingConsentElementWithAssociatedPurpose)

	result := svc.deleteConsentElements(context.Background(), []string{"email"}, log.GetLogger())

	s.Nil(result)
}

func (s *UserSchemaServiceConsentTestSuite) TestDeleteConsentElements_OtherDeleteError() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestSchemaServiceWithConsent(cMock)

	cMock.EXPECT().ListConsentElements(mock.Anything, "default", consent.NamespaceAttribute, "email").
		Return([]consent.ConsentElement{{ID: "e1", Name: "email"}}, nil)
	cMock.EXPECT().DeleteConsentElement(mock.Anything, "default", "e1").
		Return(&serviceerror.InternalServerErrorWithI18n)

	result := svc.deleteConsentElements(context.Background(), []string{"email"}, log.GetLogger())

	s.NotNil(result)
}

func (s *UserSchemaServiceConsentTestSuite) TestDeleteConsentElements_ListError() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestSchemaServiceWithConsent(cMock)

	cMock.EXPECT().ListConsentElements(mock.Anything, "default", consent.NamespaceAttribute, "email").
		Return(nil, &serviceerror.InternalServerErrorWithI18n)

	result := svc.deleteConsentElements(context.Background(), []string{"email"}, log.GetLogger())

	s.NotNil(result)
}

func (s *UserSchemaServiceConsentTestSuite) TestDeleteConsentElements_MultipleAttrs() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestSchemaServiceWithConsent(cMock)

	cMock.EXPECT().ListConsentElements(mock.Anything, "default", consent.NamespaceAttribute, "email").
		Return([]consent.ConsentElement{{ID: "e1", Name: "email"}}, nil)
	cMock.EXPECT().DeleteConsentElement(mock.Anything, "default", "e1").
		Return((*serviceerror.I18nServiceError)(nil))

	cMock.EXPECT().ListConsentElements(mock.Anything, "default", consent.NamespaceAttribute, "phone").
		Return([]consent.ConsentElement{{ID: "e2", Name: "phone"}}, nil)
	cMock.EXPECT().DeleteConsentElement(mock.Anything, "default", "e2").
		Return((*serviceerror.I18nServiceError)(nil))

	result := svc.deleteConsentElements(context.Background(), []string{"email", "phone"}, log.GetLogger())

	s.Nil(result)
}

// ----- syncConsentElementsOnCreate -----

func (s *UserSchemaServiceConsentTestSuite) TestSyncConsentElementsOnCreate_EmptySchema() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestSchemaServiceWithConsent(cMock)

	result := svc.syncConsentElementsOnCreate(context.Background(), json.RawMessage{}, log.GetLogger())

	s.Nil(result)
}

func (s *UserSchemaServiceConsentTestSuite) TestSyncConsentElementsOnCreate_WithAttributes() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestSchemaServiceWithConsent(cMock)

	schema := json.RawMessage(`{"email":{},"phone":{}}`)
	cMock.EXPECT().ValidateConsentElements(mock.Anything, "default", mock.Anything).
		Return([]string{}, nil)
	cMock.EXPECT().CreateConsentElements(mock.Anything, "default", mock.Anything).
		Return([]consent.ConsentElement{{ID: "e1"}, {ID: "e2"}}, nil)

	result := svc.syncConsentElementsOnCreate(context.Background(), schema, log.GetLogger())

	s.Nil(result)
}

func (s *UserSchemaServiceConsentTestSuite) TestSyncConsentElementsOnCreate_InvalidSchema() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestSchemaServiceWithConsent(cMock)

	result := svc.syncConsentElementsOnCreate(context.Background(),
		json.RawMessage(`{bad`), log.GetLogger())

	s.NotNil(result)
}

func (s *UserSchemaServiceConsentTestSuite) TestSyncConsentElementsOnCreate_CreateError() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestSchemaServiceWithConsent(cMock)

	schema := json.RawMessage(`{"email":{}}`)
	cMock.EXPECT().ValidateConsentElements(mock.Anything, "default", mock.Anything).
		Return(nil, &serviceerror.InternalServerErrorWithI18n)

	result := svc.syncConsentElementsOnCreate(context.Background(), schema, log.GetLogger())

	s.NotNil(result)
}

// ----- syncConsentElementsOnUpdate -----

func (s *UserSchemaServiceConsentTestSuite) TestSyncConsentElementsOnUpdate_NoChanges() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestSchemaServiceWithConsent(cMock)

	schema := json.RawMessage(`{"email":{}}`)
	cMock.EXPECT().ValidateConsentElements(mock.Anything, "default", mock.Anything).
		Return([]string{"email"}, nil)

	result := svc.syncConsentElementsOnUpdate(context.Background(), schema, schema, log.GetLogger())

	s.Nil(result)
}

func (s *UserSchemaServiceConsentTestSuite) TestSyncConsentElementsOnUpdate_NewAttrAdded() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestSchemaServiceWithConsent(cMock)

	oldSchema := json.RawMessage(`{"email":{}}`)
	newSchema := json.RawMessage(`{"email":{},"phone":{}}`)

	cMock.EXPECT().ValidateConsentElements(mock.Anything, "default", mock.Anything).
		Return([]string{"email"}, nil)
	cMock.EXPECT().CreateConsentElements(mock.Anything, "default",
		[]consent.ConsentElementInput{{Name: "phone", Namespace: consent.NamespaceAttribute}}).
		Return([]consent.ConsentElement{{ID: "e2", Name: "phone"}}, nil)

	result := svc.syncConsentElementsOnUpdate(context.Background(), oldSchema, newSchema, log.GetLogger())

	s.Nil(result)
}

func (s *UserSchemaServiceConsentTestSuite) TestSyncConsentElementsOnUpdate_AttrRemoved() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestSchemaServiceWithConsent(cMock)

	oldSchema := json.RawMessage(`{"email":{},"phone":{}}`)
	newSchema := json.RawMessage(`{"email":{}}`)

	// Validate "email" exists, no new elements to create
	cMock.EXPECT().ValidateConsentElements(mock.Anything, "default", mock.Anything).
		Return([]string{"email"}, nil)

	// Delete "phone" which was removed
	cMock.EXPECT().ListConsentElements(mock.Anything, "default", consent.NamespaceAttribute, "phone").
		Return([]consent.ConsentElement{{ID: "e2", Name: "phone"}}, nil)
	cMock.EXPECT().DeleteConsentElement(mock.Anything, "default", "e2").
		Return((*serviceerror.I18nServiceError)(nil))

	result := svc.syncConsentElementsOnUpdate(context.Background(), oldSchema, newSchema, log.GetLogger())

	s.Nil(result)
}

func (s *UserSchemaServiceConsentTestSuite) TestSyncConsentElementsOnUpdate_InvalidOldSchema() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestSchemaServiceWithConsent(cMock)

	newSchema := json.RawMessage(`{"email":{}}`)

	result := svc.syncConsentElementsOnUpdate(context.Background(),
		json.RawMessage(`{bad`), newSchema, log.GetLogger())

	s.NotNil(result)
}

func (s *UserSchemaServiceConsentTestSuite) TestSyncConsentElementsOnUpdate_InvalidNewSchema() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestSchemaServiceWithConsent(cMock)

	oldSchema := json.RawMessage(`{"email":{}}`)

	result := svc.syncConsentElementsOnUpdate(context.Background(),
		oldSchema, json.RawMessage(`{bad`), log.GetLogger())

	s.NotNil(result)
}

func (s *UserSchemaServiceConsentTestSuite) TestSyncConsentElementsOnUpdate_CreateError() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestSchemaServiceWithConsent(cMock)

	oldSchema := json.RawMessage(`{"email":{}}`)
	newSchema := json.RawMessage(`{"email":{},"phone":{}}`)

	cMock.EXPECT().ValidateConsentElements(mock.Anything, "default", mock.Anything).
		Return(nil, &serviceerror.InternalServerErrorWithI18n)

	result := svc.syncConsentElementsOnUpdate(context.Background(), oldSchema, newSchema, log.GetLogger())

	s.NotNil(result)
}

func (s *UserSchemaServiceConsentTestSuite) TestSyncConsentElementsOnUpdate_DeleteError() {
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())
	svc := newTestSchemaServiceWithConsent(cMock)

	oldSchema := json.RawMessage(`{"email":{},"phone":{}}`)
	newSchema := json.RawMessage(`{"email":{}}`)

	cMock.EXPECT().ValidateConsentElements(mock.Anything, "default", mock.Anything).
		Return([]string{"email"}, nil)
	cMock.EXPECT().ListConsentElements(mock.Anything, "default", consent.NamespaceAttribute, "phone").
		Return(nil, &serviceerror.InternalServerErrorWithI18n)

	result := svc.syncConsentElementsOnUpdate(context.Background(), oldSchema, newSchema, log.GetLogger())

	s.NotNil(result)
}

// ----- CreateUserSchema compensation tests -----

// TestCreateUserSchema_ConsentSyncFails_CompensatesWithSchemaDeletion verifies that when
// consent element sync fails after schema creation, the schema is deleted as compensation.
func (s *UserSchemaServiceConsentTestSuite) TestCreateUserSchema_ConsentSyncFails_CompensatesWithSchemaDeletion() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{Enabled: false},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	require.NoError(s.T(), err)
	defer config.ResetThunderRuntime()

	storeMock := newUserSchemaStoreInterfaceMock(s.T())
	ouMock := oumock.NewOrganizationUnitServiceInterfaceMock(s.T())
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())

	svc := &userSchemaService{
		userSchemaStore: storeMock,
		ouService:       ouMock,
		transactioner:   &mockTransactioner{},
		consentService:  cMock,
	}

	ouMock.On("IsOrganizationUnitExists", mock.Anything, testOUID1).Return(true, (*serviceerror.ServiceError)(nil))
	storeMock.On("GetUserSchemaByName", mock.Anything, "test-schema").Return(UserSchema{}, ErrUserSchemaNotFound)
	storeMock.On("CreateUserSchema", mock.Anything, mock.Anything).Return(nil)
	// Consent sync fails.
	cMock.On("IsEnabled").Return(true)
	cMock.On("ValidateConsentElements", mock.Anything, "default", mock.Anything).
		Return(nil, &serviceerror.InternalServerErrorWithI18n)
	// Compensation: schema must be deleted.
	storeMock.On("DeleteUserSchemaByID", mock.Anything, mock.Anything).Return(nil).Maybe()

	request := CreateUserSchemaRequest{
		Name:               "test-schema",
		OrganizationUnitID: testOUID1,
		Schema:             json.RawMessage(`{"email":{"type":"string"}}`),
	}

	result, svcErr := svc.CreateUserSchema(context.Background(), request)

	s.Nil(result)
	s.NotNil(svcErr)
	storeMock.AssertCalled(s.T(), "DeleteUserSchemaByID", mock.Anything, mock.Anything)
}

// ----- UpdateUserSchema compensation tests -----

// TestUpdateUserSchema_ConsentSyncFails_CompensatesWithSchemaRevert verifies that when
// consent element sync fails after schema update, the schema is reverted as compensation.
func (s *UserSchemaServiceConsentTestSuite) TestUpdateUserSchema_ConsentSyncFails_CompensatesWithSchemaRevert() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{Enabled: false},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	require.NoError(s.T(), err)
	defer config.ResetThunderRuntime()

	storeMock := newUserSchemaStoreInterfaceMock(s.T())
	ouMock := oumock.NewOrganizationUnitServiceInterfaceMock(s.T())
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())

	svc := &userSchemaService{
		userSchemaStore: storeMock,
		ouService:       ouMock,
		transactioner:   &mockTransactioner{},
		consentService:  cMock,
	}

	existingSchema := UserSchema{
		ID:                 "schema-id",
		Name:               "test-schema",
		OrganizationUnitID: testOUID1,
		Schema:             json.RawMessage(`{"email":{"type":"string"}}`),
	}

	storeMock.On("IsUserSchemaDeclarative", "schema-id").Return(false)
	ouMock.On("IsOrganizationUnitExists", mock.Anything, testOUID1).Return(true, (*serviceerror.ServiceError)(nil))
	storeMock.On("GetUserSchemaByID", mock.Anything, "schema-id").Return(existingSchema, nil)
	// Both the actual update (in tx) and the compensation revert share the same mock.
	storeMock.On("UpdateUserSchemaByID", mock.Anything, "schema-id", mock.Anything).Return(nil)
	// Consent sync fails: ValidateConsentElements returns an I18n error.
	cMock.On("IsEnabled").Return(true)
	cMock.On("ValidateConsentElements", mock.Anything, "default", mock.Anything).
		Return(nil, &serviceerror.InternalServerErrorWithI18n)

	request := UpdateUserSchemaRequest{
		Name:               "test-schema",
		OrganizationUnitID: testOUID1,
		Schema:             json.RawMessage(`{"email":{"type":"string"}}`),
	}

	result, svcErr := svc.UpdateUserSchema(context.Background(), "schema-id", request)

	s.Nil(result)
	s.NotNil(svcErr)
	// Verify compensation was called: UpdateUserSchemaByID twice (update + revert).
	storeMock.AssertNumberOfCalls(s.T(), "UpdateUserSchemaByID", 2)
}

// ----- DeleteUserSchema consent tests -----

// TestDeleteUserSchema_ConsentEnabled_DeletesConsentElementsAfterSchemaDeletion verifies
// that when consent is enabled, consent elements are cleaned up after the schema is deleted.
func (s *UserSchemaServiceConsentTestSuite) TestDeleteUserSchema_ConsentEnabled_DeletesConsentElements() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{Enabled: false},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/tmp/test", testConfig)
	require.NoError(s.T(), err)
	defer config.ResetThunderRuntime()

	storeMock := newUserSchemaStoreInterfaceMock(s.T())
	cMock := consentmock.NewConsentServiceInterfaceMock(s.T())

	svc := &userSchemaService{
		userSchemaStore: storeMock,
		transactioner:   &mockTransactioner{},
		consentService:  cMock,
	}

	existingSchema := UserSchema{
		ID:                 "schema-id",
		Name:               "test-schema",
		OrganizationUnitID: testOUID1,
		Schema:             json.RawMessage(`{"email":{"type":"string"}}`),
	}

	storeMock.On("GetUserSchemaByID", mock.Anything, "schema-id").Return(existingSchema, nil)
	storeMock.On("IsUserSchemaDeclarative", "schema-id").Return(false)
	cMock.On("IsEnabled").Return(true)
	storeMock.On("DeleteUserSchemaByID", mock.Anything, "schema-id").Return(nil)
	// Consent element cleanup: ListConsentElements → found → DeleteConsentElement
	cMock.On("ListConsentElements", mock.Anything, "default", consent.NamespaceAttribute, "email").
		Return([]consent.ConsentElement{{ID: "elem-1", Name: "email"}}, (*serviceerror.I18nServiceError)(nil))
	cMock.On("DeleteConsentElement", mock.Anything, "default", "elem-1").
		Return((*serviceerror.I18nServiceError)(nil))

	svcErr := svc.DeleteUserSchema(context.Background(), "schema-id")

	s.Nil(svcErr)
	storeMock.AssertCalled(s.T(), "DeleteUserSchemaByID", mock.Anything, "schema-id")
	cMock.AssertCalled(s.T(), "ListConsentElements", mock.Anything, "default", consent.NamespaceAttribute, "email")
}
