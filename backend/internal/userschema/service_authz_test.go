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

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/security"
	"github.com/asgardeo/thunder/internal/system/sysauthz"
	"github.com/asgardeo/thunder/tests/mocks/oumock"
	"github.com/asgardeo/thunder/tests/mocks/sysauthzmock"
)

// ---------------------------------------------------------------------------
// Helper: create a deny-all authz mock
// ---------------------------------------------------------------------------

func newAuthzError(t interface {
	mock.TestingT
	Cleanup(func())
}) *sysauthzmock.SystemAuthorizationServiceInterfaceMock {
	svcErr := &serviceerror.ServiceError{
		Type:  serviceerror.ServerErrorType,
		Code:  "SSE-5000",
		Error: "authz failure",
	}
	m := sysauthzmock.NewSystemAuthorizationServiceInterfaceMock(t)
	m.On("IsActionAllowed", mock.Anything, mock.Anything, mock.Anything).
		Return(false, svcErr).Maybe()
	m.On("GetAccessibleResources", mock.Anything, mock.Anything, mock.Anything).
		Return((*sysauthz.AccessibleResources)(nil), svcErr).Maybe()
	return m
}

func initTestRuntime(t *testing.T) {
	t.Helper()
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{Enabled: false},
	}
	config.ResetThunderRuntime()
	require.NoError(t, config.InitializeThunderRuntime("/tmp/test", testConfig))
	t.Cleanup(config.ResetThunderRuntime)
}

// ---------------------------------------------------------------------------
// Suite for authorization tests
// ---------------------------------------------------------------------------

type AuthzTestSuite struct {
	suite.Suite
}

func TestAuthzTestSuite(t *testing.T) {
	suite.Run(t, new(AuthzTestSuite))
}

// ---- GetUserSchemaList ----

func (s *AuthzTestSuite) TestGetUserSchemaList_AllAllowed() {
	storeMock := newUserSchemaStoreInterfaceMock(s.T())
	storeMock.On("GetUserSchemaListCount", mock.Anything).Return(2, nil)
	storeMock.On("GetUserSchemaList", mock.Anything, 10, 0).Return([]UserSchemaListItem{
		{ID: "s1", Name: "schema1", OrganizationUnitID: testOUID1},
		{ID: "s2", Name: "schema2", OrganizationUnitID: testOUID2},
	}, nil)

	svc := &userSchemaService{
		userSchemaStore: storeMock,
		transactioner:   &mockTransactioner{},
		authzService:    newAllowAllAuthz(s.T()),
	}

	resp, svcErr := svc.GetUserSchemaList(context.Background(), 10, 0)
	s.Require().Nil(svcErr)
	s.Require().NotNil(resp)
	s.Equal(2, resp.TotalResults)
	s.Len(resp.Schemas, 2)
}

func (s *AuthzTestSuite) TestGetUserSchemaList_FilteredByOUIDs() {
	storeMock := newUserSchemaStoreInterfaceMock(s.T())
	storeMock.On("GetUserSchemaListCountByOUIDs", mock.Anything, []string{testOUID1}).Return(1, nil)
	storeMock.On("GetUserSchemaListByOUIDs", mock.Anything, []string{testOUID1}, 10, 0).
		Return([]UserSchemaListItem{
			{ID: "s1", Name: "schema1", OrganizationUnitID: testOUID1},
		}, nil)

	authzMock := sysauthzmock.NewSystemAuthorizationServiceInterfaceMock(s.T())
	authzMock.On("GetAccessibleResources", mock.Anything, security.ActionListUserSchemas,
		security.ResourceTypeUserSchema).
		Return(&sysauthz.AccessibleResources{AllAllowed: false, IDs: []string{testOUID1}}, nil)

	svc := &userSchemaService{
		userSchemaStore: storeMock,
		transactioner:   &mockTransactioner{},
		authzService:    authzMock,
	}

	resp, svcErr := svc.GetUserSchemaList(context.Background(), 10, 0)
	s.Require().Nil(svcErr)
	s.Require().NotNil(resp)
	s.Equal(1, resp.TotalResults)
	s.Len(resp.Schemas, 1)
	s.Equal("s1", resp.Schemas[0].ID)
}

func (s *AuthzTestSuite) TestGetUserSchemaList_EmptyAccessibleOUIDs() {
	authzMock := sysauthzmock.NewSystemAuthorizationServiceInterfaceMock(s.T())
	authzMock.On("GetAccessibleResources", mock.Anything, security.ActionListUserSchemas,
		security.ResourceTypeUserSchema).
		Return(&sysauthz.AccessibleResources{AllAllowed: false, IDs: []string{}}, nil)

	svc := &userSchemaService{
		userSchemaStore: newUserSchemaStoreInterfaceMock(s.T()),
		transactioner:   &mockTransactioner{},
		authzService:    authzMock,
	}

	resp, svcErr := svc.GetUserSchemaList(context.Background(), 10, 0)
	s.Require().Nil(svcErr)
	s.Require().NotNil(resp)
	s.Equal(0, resp.TotalResults)
	s.Empty(resp.Schemas)
}

func (s *AuthzTestSuite) TestGetUserSchemaList_AuthzServiceError() {
	svc := &userSchemaService{
		userSchemaStore: newUserSchemaStoreInterfaceMock(s.T()),
		transactioner:   &mockTransactioner{},
		authzService:    newAuthzError(s.T()),
	}

	resp, svcErr := svc.GetUserSchemaList(context.Background(), 10, 0)
	s.Nil(resp)
	s.Require().NotNil(svcErr)
	s.Equal(ErrorInternalServerError.Code, svcErr.Code)
}

func (s *AuthzTestSuite) TestGetUserSchemaList_NilAuthzService() {
	storeMock := newUserSchemaStoreInterfaceMock(s.T())
	storeMock.On("GetUserSchemaListCount", mock.Anything).Return(1, nil)
	storeMock.On("GetUserSchemaList", mock.Anything, 10, 0).Return([]UserSchemaListItem{
		{ID: "s1", Name: "schema1", OrganizationUnitID: testOUID1},
	}, nil)

	svc := &userSchemaService{
		userSchemaStore: storeMock,
		transactioner:   &mockTransactioner{},
		authzService:    nil,
	}

	resp, svcErr := svc.GetUserSchemaList(context.Background(), 10, 0)
	s.Require().Nil(svcErr)
	s.Require().NotNil(resp)
	s.Equal(1, resp.TotalResults)
}

// ---- CreateUserSchema ----

func (s *AuthzTestSuite) TestCreateUserSchema_Denied() {
	initTestRuntime(s.T())

	storeMock := newUserSchemaStoreInterfaceMock(s.T())
	ouMock := oumock.NewOrganizationUnitServiceInterfaceMock(s.T())
	ouMock.On("IsOrganizationUnitExists", mock.Anything, testOUID1).
		Return(true, (*serviceerror.ServiceError)(nil))

	authzMock := sysauthzmock.NewSystemAuthorizationServiceInterfaceMock(s.T())
	authzMock.On("IsActionAllowed", mock.Anything, security.ActionCreateUserSchema,
		&sysauthz.ActionContext{ResourceType: security.ResourceTypeUserSchema, OuID: testOUID1}).
		Return(false, nil)

	svc := &userSchemaService{
		userSchemaStore: storeMock,
		ouService:       ouMock,
		transactioner:   &mockTransactioner{},
		authzService:    authzMock,
	}

	result, svcErr := svc.CreateUserSchema(context.Background(), CreateUserSchemaRequest{
		Name:               "test-schema",
		OrganizationUnitID: testOUID1,
		Schema:             json.RawMessage(`{"email":{"type":"string"}}`),
	})
	s.Nil(result)
	s.Require().NotNil(svcErr)
	s.Equal(serviceerror.ErrorUnauthorized.Code, svcErr.Code)
}

func (s *AuthzTestSuite) TestCreateUserSchema_AuthzError() {
	initTestRuntime(s.T())

	storeMock := newUserSchemaStoreInterfaceMock(s.T())
	ouMock := oumock.NewOrganizationUnitServiceInterfaceMock(s.T())
	ouMock.On("IsOrganizationUnitExists", mock.Anything, testOUID1).
		Return(true, (*serviceerror.ServiceError)(nil))

	svc := &userSchemaService{
		userSchemaStore: storeMock,
		ouService:       ouMock,
		transactioner:   &mockTransactioner{},
		authzService:    newAuthzError(s.T()),
	}

	result, svcErr := svc.CreateUserSchema(context.Background(), CreateUserSchemaRequest{
		Name:               "test-schema",
		OrganizationUnitID: testOUID1,
		Schema:             json.RawMessage(`{"email":{"type":"string"}}`),
	})
	s.Nil(result)
	s.Require().NotNil(svcErr)
	s.Equal(ErrorInternalServerError.Code, svcErr.Code)
}

// ---- GetUserSchema ----

func (s *AuthzTestSuite) TestGetUserSchema_Denied() {
	storeMock := newUserSchemaStoreInterfaceMock(s.T())
	storeMock.On("GetUserSchemaByID", mock.Anything, "schema-1").
		Return(UserSchema{ID: "schema-1", OrganizationUnitID: testOUID1}, nil)

	authzMock := sysauthzmock.NewSystemAuthorizationServiceInterfaceMock(s.T())
	authzMock.On("IsActionAllowed", mock.Anything, security.ActionReadUserSchema,
		&sysauthz.ActionContext{ResourceType: security.ResourceTypeUserSchema, OuID: testOUID1}).
		Return(false, nil)

	svc := &userSchemaService{
		userSchemaStore: storeMock,
		transactioner:   &mockTransactioner{},
		authzService:    authzMock,
	}

	result, svcErr := svc.GetUserSchema(context.Background(), "schema-1")
	s.Nil(result)
	s.Require().NotNil(svcErr)
	s.Equal(serviceerror.ErrorUnauthorized.Code, svcErr.Code)
}

func (s *AuthzTestSuite) TestGetUserSchema_AuthzError() {
	storeMock := newUserSchemaStoreInterfaceMock(s.T())
	storeMock.On("GetUserSchemaByID", mock.Anything, "schema-1").
		Return(UserSchema{ID: "schema-1", OrganizationUnitID: testOUID1}, nil)

	svc := &userSchemaService{
		userSchemaStore: storeMock,
		transactioner:   &mockTransactioner{},
		authzService:    newAuthzError(s.T()),
	}

	result, svcErr := svc.GetUserSchema(context.Background(), "schema-1")
	s.Nil(result)
	s.Require().NotNil(svcErr)
	s.Equal(ErrorInternalServerError.Code, svcErr.Code)
}

// ---- GetUserSchemaByName ----

func (s *AuthzTestSuite) TestGetUserSchemaByName_Denied() {
	storeMock := newUserSchemaStoreInterfaceMock(s.T())
	storeMock.On("GetUserSchemaByName", mock.Anything, "employee").
		Return(UserSchema{ID: "schema-1", Name: "employee", OrganizationUnitID: testOUID2}, nil)

	authzMock := sysauthzmock.NewSystemAuthorizationServiceInterfaceMock(s.T())
	authzMock.On("IsActionAllowed", mock.Anything, security.ActionReadUserSchema,
		&sysauthz.ActionContext{ResourceType: security.ResourceTypeUserSchema, OuID: testOUID2}).
		Return(false, nil)

	svc := &userSchemaService{
		userSchemaStore: storeMock,
		transactioner:   &mockTransactioner{},
		authzService:    authzMock,
	}

	result, svcErr := svc.GetUserSchemaByName(context.Background(), "employee")
	s.Nil(result)
	s.Require().NotNil(svcErr)
	s.Equal(serviceerror.ErrorUnauthorized.Code, svcErr.Code)
}

func (s *AuthzTestSuite) TestGetUserSchemaByName_AuthzError() {
	storeMock := newUserSchemaStoreInterfaceMock(s.T())
	storeMock.On("GetUserSchemaByName", mock.Anything, "employee").
		Return(UserSchema{ID: "schema-1", Name: "employee", OrganizationUnitID: testOUID2}, nil)

	svc := &userSchemaService{
		userSchemaStore: storeMock,
		transactioner:   &mockTransactioner{},
		authzService:    newAuthzError(s.T()),
	}

	result, svcErr := svc.GetUserSchemaByName(context.Background(), "employee")
	s.Nil(result)
	s.Require().NotNil(svcErr)
	s.Equal(ErrorInternalServerError.Code, svcErr.Code)
}

// ---- UpdateUserSchema ----

func (s *AuthzTestSuite) TestUpdateUserSchema_Denied() {
	initTestRuntime(s.T())

	storeMock := newUserSchemaStoreInterfaceMock(s.T())
	storeMock.On("IsUserSchemaDeclarative", mock.Anything).Return(false).Maybe()
	storeMock.On("GetUserSchemaByID", mock.Anything, "schema-1").
		Return(UserSchema{
			ID:                 "schema-1",
			Name:               "employee",
			OrganizationUnitID: testOUID1,
			Schema:             json.RawMessage(`{"email":{"type":"string"}}`),
		}, nil)

	ouMock := oumock.NewOrganizationUnitServiceInterfaceMock(s.T())
	ouMock.On("IsOrganizationUnitExists", mock.Anything, testOUID1).
		Return(true, (*serviceerror.ServiceError)(nil))

	authzMock := sysauthzmock.NewSystemAuthorizationServiceInterfaceMock(s.T())
	authzMock.On("IsActionAllowed", mock.Anything, security.ActionUpdateUserSchema,
		&sysauthz.ActionContext{ResourceType: security.ResourceTypeUserSchema, OuID: testOUID1}).
		Return(false, nil)

	svc := &userSchemaService{
		userSchemaStore: storeMock,
		ouService:       ouMock,
		transactioner:   &mockTransactioner{},
		authzService:    authzMock,
	}

	result, svcErr := svc.UpdateUserSchema(context.Background(), "schema-1", UpdateUserSchemaRequest{
		Name:               "employee",
		OrganizationUnitID: testOUID1,
		Schema:             json.RawMessage(`{"email":{"type":"string"}}`),
	})
	s.Nil(result)
	s.Require().NotNil(svcErr)
	s.Equal(serviceerror.ErrorUnauthorized.Code, svcErr.Code)
}

func (s *AuthzTestSuite) TestUpdateUserSchema_AuthzError() {
	initTestRuntime(s.T())

	storeMock := newUserSchemaStoreInterfaceMock(s.T())
	storeMock.On("IsUserSchemaDeclarative", mock.Anything).Return(false).Maybe()
	storeMock.On("GetUserSchemaByID", mock.Anything, "schema-1").
		Return(UserSchema{
			ID:                 "schema-1",
			Name:               "employee",
			OrganizationUnitID: testOUID1,
			Schema:             json.RawMessage(`{"email":{"type":"string"}}`),
		}, nil)

	ouMock := oumock.NewOrganizationUnitServiceInterfaceMock(s.T())
	ouMock.On("IsOrganizationUnitExists", mock.Anything, testOUID1).
		Return(true, (*serviceerror.ServiceError)(nil))

	svc := &userSchemaService{
		userSchemaStore: storeMock,
		ouService:       ouMock,
		transactioner:   &mockTransactioner{},
		authzService:    newAuthzError(s.T()),
	}

	result, svcErr := svc.UpdateUserSchema(context.Background(), "schema-1", UpdateUserSchemaRequest{
		Name:               "employee",
		OrganizationUnitID: testOUID1,
		Schema:             json.RawMessage(`{"email":{"type":"string"}}`),
	})
	s.Nil(result)
	s.Require().NotNil(svcErr)
	s.Equal(ErrorInternalServerError.Code, svcErr.Code)
}

// ---- DeleteUserSchema ----

func (s *AuthzTestSuite) TestDeleteUserSchema_Denied() {
	initTestRuntime(s.T())

	storeMock := newUserSchemaStoreInterfaceMock(s.T())
	storeMock.On("IsUserSchemaDeclarative", mock.Anything).Return(false).Maybe()
	storeMock.On("GetUserSchemaByID", mock.Anything, "schema-1").
		Return(UserSchema{
			ID:                 "schema-1",
			OrganizationUnitID: testOUID1,
		}, nil)

	authzMock := sysauthzmock.NewSystemAuthorizationServiceInterfaceMock(s.T())
	authzMock.On("IsActionAllowed", mock.Anything, security.ActionDeleteUserSchema,
		&sysauthz.ActionContext{ResourceType: security.ResourceTypeUserSchema, OuID: testOUID1}).
		Return(false, nil)

	svc := &userSchemaService{
		userSchemaStore: storeMock,
		transactioner:   &mockTransactioner{},
		authzService:    authzMock,
	}

	svcErr := svc.DeleteUserSchema(context.Background(), "schema-1")
	s.Require().NotNil(svcErr)
	s.Equal(serviceerror.ErrorUnauthorized.Code, svcErr.Code)
}

func (s *AuthzTestSuite) TestDeleteUserSchema_AuthzError() {
	initTestRuntime(s.T())

	storeMock := newUserSchemaStoreInterfaceMock(s.T())
	storeMock.On("IsUserSchemaDeclarative", mock.Anything).Return(false).Maybe()
	storeMock.On("GetUserSchemaByID", mock.Anything, "schema-1").
		Return(UserSchema{
			ID:                 "schema-1",
			OrganizationUnitID: testOUID1,
		}, nil)

	svc := &userSchemaService{
		userSchemaStore: storeMock,
		transactioner:   &mockTransactioner{},
		authzService:    newAuthzError(s.T()),
	}

	svcErr := svc.DeleteUserSchema(context.Background(), "schema-1")
	s.Require().NotNil(svcErr)
	s.Equal(ErrorInternalServerError.Code, svcErr.Code)
}

func (s *AuthzTestSuite) TestDeleteUserSchema_NotFound_StillChecksAuthz() {
	initTestRuntime(s.T())

	storeMock := newUserSchemaStoreInterfaceMock(s.T())
	storeMock.On("IsUserSchemaDeclarative", mock.Anything).Return(false).Maybe()
	storeMock.On("GetUserSchemaByID", mock.Anything, "nonexistent").
		Return(UserSchema{}, ErrUserSchemaNotFound)

	authzMock := sysauthzmock.NewSystemAuthorizationServiceInterfaceMock(s.T())
	// Expect delete authz check with empty OU (schema doesn't exist, so no OU to check against).
	authzMock.On("IsActionAllowed", mock.Anything, security.ActionDeleteUserSchema,
		&sysauthz.ActionContext{ResourceType: security.ResourceTypeUserSchema, OuID: ""}).
		Return(false, nil)

	svc := &userSchemaService{
		userSchemaStore: storeMock,
		transactioner:   &mockTransactioner{},
		authzService:    authzMock,
	}

	svcErr := svc.DeleteUserSchema(context.Background(), "nonexistent")
	s.Require().NotNil(svcErr)
	s.Equal(serviceerror.ErrorUnauthorized.Code, svcErr.Code,
		"delete of nonexistent schema should still return unauthorized for denied callers")
}

func (s *AuthzTestSuite) TestDeleteUserSchema_NotFound_Authorized_ReturnsNil() {
	initTestRuntime(s.T())

	storeMock := newUserSchemaStoreInterfaceMock(s.T())
	storeMock.On("IsUserSchemaDeclarative", mock.Anything).Return(false).Maybe()
	storeMock.On("GetUserSchemaByID", mock.Anything, "nonexistent").
		Return(UserSchema{}, ErrUserSchemaNotFound)

	authzMock := sysauthzmock.NewSystemAuthorizationServiceInterfaceMock(s.T())
	authzMock.On("IsActionAllowed", mock.Anything, security.ActionDeleteUserSchema,
		&sysauthz.ActionContext{ResourceType: security.ResourceTypeUserSchema, OuID: ""}).
		Return(true, nil)

	svc := &userSchemaService{
		userSchemaStore: storeMock,
		transactioner:   &mockTransactioner{},
		authzService:    authzMock,
	}

	svcErr := svc.DeleteUserSchema(context.Background(), "nonexistent")
	s.Nil(svcErr, "authorized caller deleting nonexistent schema should get nil (idempotent)")
}

// ---- Nil authzService (backward compatibility) ----

func (s *AuthzTestSuite) TestGetUserSchema_NilAuthz_NoError() {
	storeMock := newUserSchemaStoreInterfaceMock(s.T())
	storeMock.On("GetUserSchemaByID", mock.Anything, "schema-1").
		Return(UserSchema{ID: "schema-1", Name: "test", OrganizationUnitID: testOUID1}, nil)

	svc := &userSchemaService{
		userSchemaStore: storeMock,
		transactioner:   &mockTransactioner{},
		authzService:    nil,
	}

	result, svcErr := svc.GetUserSchema(context.Background(), "schema-1")
	s.Require().Nil(svcErr)
	s.Require().NotNil(result)
	s.Equal("schema-1", result.ID)
}

func (s *AuthzTestSuite) TestGetUserSchemaByName_NilAuthz_NoError() {
	storeMock := newUserSchemaStoreInterfaceMock(s.T())
	storeMock.On("GetUserSchemaByName", mock.Anything, "employee").
		Return(UserSchema{ID: "schema-1", Name: "employee"}, nil)

	svc := &userSchemaService{
		userSchemaStore: storeMock,
		transactioner:   &mockTransactioner{},
		authzService:    nil,
	}

	result, svcErr := svc.GetUserSchemaByName(context.Background(), "employee")
	s.Require().Nil(svcErr)
	s.Require().NotNil(result)
}

func (s *AuthzTestSuite) TestDeleteUserSchema_NilAuthz_NoError() {
	initTestRuntime(s.T())

	storeMock := newUserSchemaStoreInterfaceMock(s.T())
	storeMock.On("IsUserSchemaDeclarative", mock.Anything).Return(false).Maybe()
	storeMock.On("GetUserSchemaByID", mock.Anything, "schema-1").
		Return(UserSchema{ID: "schema-1", OrganizationUnitID: testOUID1}, nil)
	storeMock.On("DeleteUserSchemaByID", mock.Anything, "schema-1").Return(nil)

	svc := &userSchemaService{
		userSchemaStore: storeMock,
		transactioner:   &mockTransactioner{},
		authzService:    nil,
	}

	svcErr := svc.DeleteUserSchema(context.Background(), "schema-1")
	s.Nil(svcErr)
}
