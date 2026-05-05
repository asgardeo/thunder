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

package application

import (
	"context"
	"encoding/json"

	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/cert"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	i18ncore "github.com/asgardeo/thunder/internal/system/i18n/core"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/pkg/application/model"
	"github.com/asgardeo/thunder/pkg/entityprovider"
	"github.com/asgardeo/thunder/pkg/inboundclient"
	inboundmodel "github.com/asgardeo/thunder/pkg/inboundclient/model"
	oauth2const "github.com/asgardeo/thunder/pkg/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/tests/mocks/entityprovidermock"
	"github.com/asgardeo/thunder/tests/mocks/inboundclientmock"
	"github.com/asgardeo/thunder/tests/mocks/oumock"
)

const testServiceAppID = "app123"
const testClientID = "test-client-id"
const testOUID = "default-ou"
const testConflictingAppID = "app456"

type ServiceTestSuite struct {
	suite.Suite
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}

func (suite *ServiceTestSuite) TestBuildBasicApplicationResponse() {
	cfg := inboundmodel.InboundClient{
		ID:                        "app-123",
		AuthFlowID:                "auth_flow_1",
		RegistrationFlowID:        "reg_flow_1",
		IsRegistrationFlowEnabled: true,
	}
	sysAttrs, _ := json.Marshal(map[string]interface{}{
		"name":        "Test App",
		"description": "Test Description",
		"clientId":    "client-123",
	})
	entity := &entityprovider.Entity{SystemAttributes: sysAttrs}

	result := buildBasicApplicationResponse(cfg, entity)

	assert.Equal(suite.T(), "app-123", result.ID)
	assert.Equal(suite.T(), "Test App", result.Name)
	assert.Equal(suite.T(), "Test Description", result.Description)
	assert.Equal(suite.T(), "auth_flow_1", result.AuthFlowID)
	assert.Equal(suite.T(), "reg_flow_1", result.RegistrationFlowID)
	assert.True(suite.T(), result.IsRegistrationFlowEnabled)
	assert.Equal(suite.T(), "client-123", result.ClientID)
}

func (suite *ServiceTestSuite) TestBuildBasicApplicationResponse_WithTemplate() {
	cfg := inboundmodel.InboundClient{
		ID:                        "app-123",
		AuthFlowID:                "auth_flow_1",
		RegistrationFlowID:        "reg_flow_1",
		IsRegistrationFlowEnabled: true,
		ThemeID:                   "theme-123",
		LayoutID:                  "layout-456",
		Properties: map[string]interface{}{
			"template": "spa",
			"logo_url": "https://example.com/logo.png",
		},
	}
	sysAttrs, _ := json.Marshal(map[string]interface{}{
		"name":     "Test App",
		"clientId": "client-123",
	})
	entity := &entityprovider.Entity{SystemAttributes: sysAttrs}

	result := buildBasicApplicationResponse(cfg, entity)

	assert.Equal(suite.T(), "app-123", result.ID)
	assert.Equal(suite.T(), "Test App", result.Name)
	assert.Equal(suite.T(), "theme-123", result.ThemeID)
	assert.Equal(suite.T(), "layout-456", result.LayoutID)
	assert.Equal(suite.T(), "spa", result.Template)
	assert.Equal(suite.T(), "client-123", result.ClientID)
	assert.Equal(suite.T(), "https://example.com/logo.png", result.LogoURL)
}

func (suite *ServiceTestSuite) TestBuildBasicApplicationResponse_WithEmptyTemplate() {
	cfg := inboundmodel.InboundClient{
		ID:                        "app-123",
		AuthFlowID:                "auth_flow_1",
		RegistrationFlowID:        "reg_flow_1",
		IsRegistrationFlowEnabled: true,
	}
	sysAttrs, _ := json.Marshal(map[string]interface{}{
		"name":     "Test App",
		"clientId": "client-123",
	})
	entity := &entityprovider.Entity{SystemAttributes: sysAttrs}

	result := buildBasicApplicationResponse(cfg, entity)

	assert.Equal(suite.T(), "app-123", result.ID)
	assert.Equal(suite.T(), "", result.Template)
}

// setupTestService wires a service with permissive entity-provider / OU mocks and a
// no-op transactioner. Returns the service plus the inbound-client mock
// that tests typically need to extend.
func (suite *ServiceTestSuite) setupTestService() (
	*applicationService,
	*inboundclientmock.InboundClientServiceInterfaceMock,
) {
	mockStore := inboundclientmock.NewInboundClientServiceInterfaceMock(suite.T())
	mockEntityProvider := entityprovidermock.NewEntityProviderInterfaceMock(suite.T())
	epNotFound := entityprovider.NewEntityProviderError(
		entityprovider.ErrorCodeEntityNotFound, "not found", "")
	var noEPErr *entityprovider.EntityProviderError
	mockEntityProvider.On("IdentifyEntity", mock.Anything).Maybe().Return((*string)(nil), epNotFound)
	mockEntityProvider.On("GetEntity", mock.Anything).Maybe().Return((*entityprovider.Entity)(nil), epNotFound)
	mockEntityProvider.On("GetEntitiesByIDs", mock.Anything).Maybe().Return([]entityprovider.Entity{}, noEPErr)
	mockEntityProvider.On("CreateEntity", mock.Anything, mock.Anything).
		Maybe().Return(&entityprovider.Entity{}, noEPErr)
	mockEntityProvider.On("DeleteEntity", mock.Anything).Maybe().Return(noEPErr)
	mockEntityProvider.On("UpdateSystemAttributes", mock.Anything, mock.Anything).Maybe().Return(noEPErr)
	mockEntityProvider.On("UpdateSystemCredentials", mock.Anything, mock.Anything).Maybe().Return(noEPErr)
	mockStore.On("Validate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(nil)
	mockOUService := oumock.NewOrganizationUnitServiceInterfaceMock(suite.T())
	mockOUService.On("IsOrganizationUnitExists", mock.Anything, mock.Anything).Maybe().Return(true, nil)
	service := &applicationService{
		logger:               log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService")),
		inboundClientService: mockStore,
		entityProvider:       mockEntityProvider,
		ouService:            mockOUService,
	}
	return service, mockStore
}

// resetIdentifyEntity removes broad IdentifyEntity expectations from the entity provider mock
// so a test can register a specific expectation without conflict.
func resetIdentifyEntity(service *applicationService) *entityprovidermock.EntityProviderInterfaceMock {
	return resetEntityProviderMethod(service, "IdentifyEntity")
}

// resetEntityProviderMethod removes any broad expectation for the named method on the
// entity provider mock attached to the service.
func resetEntityProviderMethod(
	service *applicationService, method string,
) *entityprovidermock.EntityProviderInterfaceMock {
	ep := service.entityProvider.(*entityprovidermock.EntityProviderInterfaceMock)
	var kept []*mock.Call
	for _, c := range ep.ExpectedCalls {
		if c.Method != method {
			kept = append(kept, c)
		}
	}
	ep.ExpectedCalls = kept
	return ep
}

// mockLoadFullApplication sets up the inbound-client + entity-provider mocks so that
// applicationService.getApplication(ctx, dto.ID) returns a result equivalent to the given
// ApplicationProcessedDTO. Builds the InboundClient (with Properties), OAuthProfile, and
// entity system attributes via the same helpers production code uses.
func mockLoadFullApplication(
	mockStore *inboundclientmock.InboundClientServiceInterfaceMock,
	service *applicationService,
	dto *model.ApplicationProcessedDTO,
) {
	configDAO := toConfigDAO(dto)
	mockStore.On("GetInboundClientByEntityID", mock.Anything, dto.ID).Return(&configDAO, nil)

	var oauthDAO *inboundmodel.OAuthProfile
	if oauthProcessed := getOAuthInboundAuthConfigProcessedDTO(dto.InboundAuthConfig); oauthProcessed != nil {
		if oauthData := buildOAuthConfigData(*oauthProcessed); oauthData != nil {
			oauthDAO = &inboundmodel.OAuthProfile{AppID: dto.ID, OAuthProfile: oauthData}
		}
	}
	if oauthDAO != nil {
		mockStore.On("GetOAuthProfileByEntityID", mock.Anything, dto.ID).Return(oauthDAO, nil)
	} else {
		mockStore.On("GetOAuthProfileByEntityID", mock.Anything, dto.ID).
			Return((*inboundmodel.OAuthProfile)(nil), inboundclient.ErrInboundClientNotFound)
	}

	sysAttrs := map[string]interface{}{}
	if dto.Name != "" {
		sysAttrs["name"] = dto.Name
	}
	if dto.Description != "" {
		sysAttrs["description"] = dto.Description
	}
	if oauthProcessed := getOAuthInboundAuthConfigProcessedDTO(dto.InboundAuthConfig); oauthProcessed != nil &&
		oauthProcessed.OAuthAppConfig != nil && oauthProcessed.OAuthAppConfig.ClientID != "" {
		sysAttrs["clientId"] = oauthProcessed.OAuthAppConfig.ClientID
	}
	sysAttrsJSON, _ := json.Marshal(sysAttrs)
	ep := resetEntityProviderMethod(service, "GetEntity")
	ep.On("GetEntity", dto.ID).Return(
		&entityprovider.Entity{ID: dto.ID, OUID: dto.OUID, SystemAttributes: sysAttrsJSON},
		(*entityprovider.EntityProviderError)(nil),
	)
}

func (suite *ServiceTestSuite) TestGetOAuthApplication_EmptyClientID() {
	service, _ := suite.setupTestService()

	result, svcErr := service.GetOAuthApplication(context.Background(), "")

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestGetOAuthApplication_NotFound() {
	service, mockStore := suite.setupTestService()

	mockStore.EXPECT().GetOAuthClientByClientID(mock.Anything, "client123").Return(nil, nil)

	result, svcErr := service.GetOAuthApplication(context.Background(), "client123")

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestGetOAuthApplication_StoreError() {
	service, mockStore := suite.setupTestService()

	mockStore.EXPECT().GetOAuthClientByClientID(mock.Anything, "client123").
		Return(nil, errors.New("store error"))

	result, svcErr := service.GetOAuthApplication(context.Background(), "client123")

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestGetOAuthApplication_Success() {
	service, mockStore := suite.setupTestService()

	mockStore.EXPECT().GetOAuthClientByClientID(mock.Anything, "client123").
		Return(&inboundmodel.OAuthClient{ClientID: "client123"}, nil)

	result, svcErr := service.GetOAuthApplication(context.Background(), "client123")

	assert.NotNil(suite.T(), result)
	assert.Nil(suite.T(), svcErr)
	assert.Equal(suite.T(), "client123", result.ClientID)
}

func (suite *ServiceTestSuite) TestGetApplication_EmptyAppID() {
	service, _ := suite.setupTestService()

	result, svcErr := service.GetApplication(context.Background(), "")

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestGetApplication_NotFound() {
	service, mockStore := suite.setupTestService()

	mockStore.On("GetInboundClientByEntityID", mock.Anything, testServiceAppID).
		Return(nil, model.ApplicationNotFoundError)

	result, svcErr := service.GetApplication(context.Background(), testServiceAppID)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestGetApplication_StoreError() {
	service, mockStore := suite.setupTestService()

	mockStore.On("GetInboundClientByEntityID", mock.Anything, testServiceAppID).Return(nil, errors.New("store error"))

	result, svcErr := service.GetApplication(context.Background(), testServiceAppID)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestGetApplication_Success() {
	service, mockStore := suite.setupTestService()

	app := &model.ApplicationProcessedDTO{
		ID:       testServiceAppID,
		Name:     "Test App",
		Metadata: map[string]interface{}{"service_key": "service_val"},
	}

	mockLoadFullApplication(mockStore, service, app)
	mockStore.EXPECT().GetCertificate(mock.Anything,
		cert.CertificateReferenceTypeApplication, testServiceAppID).Return(nil, nil)

	result, svcErr := service.GetApplication(context.Background(), testServiceAppID)

	assert.NotNil(suite.T(), result)
	assert.Nil(suite.T(), svcErr)
	assert.Equal(suite.T(), testServiceAppID, result.ID)
	assert.Equal(suite.T(), map[string]interface{}{"service_key": "service_val"}, result.Metadata)
}

func (suite *ServiceTestSuite) TestGetApplication_WithInboundAuthConfig_Success() {
	service, mockStore := suite.setupTestService()

	app := &model.ApplicationProcessedDTO{
		ID:          testServiceAppID,
		Name:        "OAuth Test App",
		Description: "App with OAuth config",
		InboundAuthConfig: []model.InboundAuthConfigProcessedDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &inboundmodel.OAuthClient{
					ClientID:                "client-id-123",
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
					PKCERequired:            true,
					PublicClient:            false,
					Scopes:                  []string{"openid", "profile"},
				},
			},
		},
	}

	mockLoadFullApplication(mockStore, service, app)
	mockStore.EXPECT().GetCertificate(mock.Anything,
		cert.CertificateReferenceTypeApplication, testServiceAppID).Return(nil, nil)
	mockStore.EXPECT().GetCertificate(mock.Anything,
		cert.CertificateReferenceTypeOAuthApp, "client-id-123").Return(nil, nil)

	result, svcErr := service.GetApplication(context.Background(), testServiceAppID)

	assert.Nil(suite.T(), svcErr)
	require.NotNil(suite.T(), result)
	assert.Equal(suite.T(), testServiceAppID, result.ID)
	assert.Equal(suite.T(), "OAuth Test App", result.Name)

	require.Len(suite.T(), result.InboundAuthConfig, 1)
	inboundAuth := result.InboundAuthConfig[0]
	assert.Equal(suite.T(), model.OAuthInboundAuthType, inboundAuth.Type)
	require.NotNil(suite.T(), inboundAuth.OAuthAppConfig)
	assert.Equal(suite.T(), "client-id-123", inboundAuth.OAuthAppConfig.ClientID)
	assert.Equal(suite.T(), []string{"https://example.com/callback"}, inboundAuth.OAuthAppConfig.RedirectURIs)
	assert.Equal(suite.T(), []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
		inboundAuth.OAuthAppConfig.GrantTypes)
	assert.Equal(suite.T(), []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
		inboundAuth.OAuthAppConfig.ResponseTypes)
	assert.Equal(suite.T(), oauth2const.TokenEndpointAuthMethodClientSecretBasic,
		inboundAuth.OAuthAppConfig.TokenEndpointAuthMethod)
	assert.True(suite.T(), inboundAuth.OAuthAppConfig.PKCERequired)
	assert.False(suite.T(), inboundAuth.OAuthAppConfig.PublicClient)
	assert.Equal(suite.T(), []string{"openid", "profile"}, inboundAuth.OAuthAppConfig.Scopes)
	assert.Nil(suite.T(), inboundAuth.OAuthAppConfig.Certificate)
}

func (suite *ServiceTestSuite) TestGetApplicationList_Success() {
	service, mockStore := suite.setupTestService()

	sysAttrs1, _ := json.Marshal(map[string]interface{}{"name": "App 1"})
	sysAttrs2, _ := json.Marshal(map[string]interface{}{"name": "App 2"})
	entities := []entityprovider.Entity{
		{ID: "app1", Category: entityprovider.EntityCategoryApp, SystemAttributes: sysAttrs1},
		{ID: "app2", Category: entityprovider.EntityCategoryApp, SystemAttributes: sysAttrs2},
	}
	cfg1 := inboundmodel.InboundClient{ID: "app1"}
	cfg2 := inboundmodel.InboundClient{ID: "app2"}

	mockStore.On("GetInboundClientList", mock.Anything).
		Return([]inboundmodel.InboundClient{cfg1, cfg2}, nil)

	ep := resetEntityProviderMethod(service, "GetEntitiesByIDs")
	ep.On("GetEntitiesByIDs", mock.Anything).
		Return(entities, (*entityprovider.EntityProviderError)(nil))

	result, svcErr := service.GetApplicationList(context.Background())

	assert.NotNil(suite.T(), result)
	assert.Nil(suite.T(), svcErr)
	assert.Equal(suite.T(), 2, result.TotalResults)
	assert.Equal(suite.T(), 2, result.Count)
	assert.Len(suite.T(), result.Applications, 2)
}

func (suite *ServiceTestSuite) TestGetApplicationList_ListError() {
	service, mockStore := suite.setupTestService()

	mockStore.On("GetInboundClientList", mock.Anything).
		Return(([]inboundmodel.InboundClient)(nil), errors.New("db error"))

	result, svcErr := service.GetApplicationList(context.Background())

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestGetApplicationList_EntityFetchError() {
	service, mockStore := suite.setupTestService()

	cfg1 := inboundmodel.InboundClient{ID: "app1"}
	mockStore.On("GetInboundClientList", mock.Anything).
		Return([]inboundmodel.InboundClient{cfg1}, nil)

	ep := resetEntityProviderMethod(service, "GetEntitiesByIDs")
	epErr := &entityprovider.EntityProviderError{Code: "INTERNAL_ERROR"}
	ep.On("GetEntitiesByIDs", mock.Anything).
		Return(([]entityprovider.Entity)(nil), epErr)

	result, svcErr := service.GetApplicationList(context.Background())

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestValidateApplication_NilApp() {
	service, _ := suite.setupTestService()

	result, inboundAuth, svcErr := service.ValidateApplication(context.Background(), nil)

	assert.Nil(suite.T(), result)
	assert.Nil(suite.T(), inboundAuth)
	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestValidateApplication_EmptyName() {
	service, _ := suite.setupTestService()

	app := &model.ApplicationDTO{
		Name: "",
		OUID: testOUID,
	}

	result, inboundAuth, svcErr := service.ValidateApplication(context.Background(), app)

	assert.Nil(suite.T(), result)
	assert.Nil(suite.T(), inboundAuth)
	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestValidateApplication_ExistingName() {
	service, _ := suite.setupTestService()

	app := &model.ApplicationDTO{
		Name: "Existing App",
		OUID: testOUID,
	}

	mockEP := resetIdentifyEntity(service)
	existingID := "existing-id"
	mockEP.On("IdentifyEntity",
		map[string]interface{}{"name": "Existing App"}).
		Return(
			&existingID, (*entityprovider.EntityProviderError)(nil))

	result, inboundAuth, svcErr := service.ValidateApplication(context.Background(), app)

	assert.Nil(suite.T(), result)
	assert.Nil(suite.T(), inboundAuth)
	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestValidateApplicationForUpdate_EmptyAppID() {
	service, _ := suite.setupTestService()

	app := &model.ApplicationDTO{
		Name: "Test App",
		OUID: testOUID,
	}

	result, inboundAuth, svcErr := service.validateApplicationForUpdate(context.Background(), "", app)

	assert.Nil(suite.T(), result)
	assert.Nil(suite.T(), inboundAuth)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &ErrorInvalidApplicationID, svcErr)
}

func (suite *ServiceTestSuite) TestValidateApplicationForUpdate_NilApp() {
	service, _ := suite.setupTestService()

	result, inboundAuth, svcErr := service.validateApplicationForUpdate(context.Background(), testServiceAppID, nil)

	assert.Nil(suite.T(), result)
	assert.Nil(suite.T(), inboundAuth)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &ErrorApplicationNil, svcErr)
}

func (suite *ServiceTestSuite) TestValidateApplicationForUpdate_EmptyName() {
	service, _ := suite.setupTestService()

	app := &model.ApplicationDTO{
		Name: "",
		OUID: testOUID,
	}

	result, inboundAuth, svcErr := service.validateApplicationForUpdate(context.Background(), testServiceAppID, app)

	assert.Nil(suite.T(), result)
	assert.Nil(suite.T(), inboundAuth)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &ErrorInvalidApplicationName, svcErr)
}

func (suite *ServiceTestSuite) TestValidateApplicationForUpdate_ApplicationNotFound() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, mockStore := suite.setupTestService()

	app := &model.ApplicationDTO{
		Name: "Test App",
		OUID: testOUID,
	}

	mockStore.On("IsDeclarative", mock.Anything, testServiceAppID).Maybe().Return(false)
	mockStore.On("GetInboundClientByEntityID", mock.Anything, testServiceAppID).
		Return(nil, inboundclient.ErrInboundClientNotFound)

	result, inboundAuth, svcErr := service.validateApplicationForUpdate(context.Background(), testServiceAppID, app)

	assert.Nil(suite.T(), result)
	assert.Nil(suite.T(), inboundAuth)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &ErrorApplicationNotFound, svcErr)
}

func (suite *ServiceTestSuite) TestValidateApplicationForUpdate_ApplicationNilFromStore() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, mockStore := suite.setupTestService()

	app := &model.ApplicationDTO{
		Name: "Test App",
		OUID: testOUID,
	}

	mockStore.On("IsDeclarative", mock.Anything, testServiceAppID).Maybe().Return(false)
	mockStore.On("GetInboundClientByEntityID", mock.Anything, testServiceAppID).Return(nil, nil)

	result, inboundAuth, svcErr := service.validateApplicationForUpdate(context.Background(), testServiceAppID, app)

	assert.Nil(suite.T(), result)
	assert.Nil(suite.T(), inboundAuth)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &ErrorApplicationNotFound, svcErr)
}

func (suite *ServiceTestSuite) TestValidateApplicationForUpdate_StoreError() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, mockStore := suite.setupTestService()

	app := &model.ApplicationDTO{
		Name: "Test App",
		OUID: testOUID,
	}

	mockStore.On("IsDeclarative", mock.Anything, testServiceAppID).Maybe().Return(false)
	mockStore.On("GetInboundClientByEntityID", mock.Anything, testServiceAppID).
		Return(nil, errors.New("database error"))

	result, inboundAuth, svcErr := service.validateApplicationForUpdate(context.Background(), testServiceAppID, app)

	assert.Nil(suite.T(), result)
	assert.Nil(suite.T(), inboundAuth)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &serviceerror.InternalServerError, svcErr)
}

func (suite *ServiceTestSuite) TestValidateApplicationForUpdate_NameConflict() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, mockStore := suite.setupTestService()

	app := &model.ApplicationDTO{
		Name: "New Name",
		OUID: testOUID,
	}

	sysAttrs, _ := json.Marshal(map[string]interface{}{"name": "Old Name"})

	mockStore.On("IsDeclarative", mock.Anything, testServiceAppID).Maybe().Return(false)
	mockStore.On("GetInboundClientByEntityID", mock.Anything, testServiceAppID).
		Return(&inboundmodel.InboundClient{ID: testServiceAppID}, nil)
	mockStore.On("GetOAuthProfileByEntityID", mock.Anything, testServiceAppID).
		Return((*inboundmodel.OAuthProfile)(nil), nil)
	mockEP := resetIdentifyEntity(service)
	mockEP.On("GetEntity", testServiceAppID).Unset()
	mockEP.On("GetEntity", testServiceAppID).Return(
		&entityprovider.Entity{
			ID: testServiceAppID, SystemAttributes: sysAttrs,
		}, (*entityprovider.EntityProviderError)(nil))
	conflictingID := testConflictingAppID
	mockEP.On("IdentifyEntity",
		map[string]interface{}{"name": "New Name"}).
		Return(
			&conflictingID, (*entityprovider.EntityProviderError)(nil))

	result, inboundAuth, svcErr := service.validateApplicationForUpdate(context.Background(), testServiceAppID, app)

	assert.Nil(suite.T(), result)
	assert.Nil(suite.T(), inboundAuth)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &ErrorApplicationAlreadyExistsWithName, svcErr)
}

func (suite *ServiceTestSuite) TestValidateApplicationForUpdate_NameCheckStoreError() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, mockStore := suite.setupTestService()

	existingApp := &model.ApplicationProcessedDTO{
		ID:   testServiceAppID,
		Name: "Old Name",
	}

	app := &model.ApplicationDTO{
		Name: "New Name",
		OUID: testOUID,
	}

	mockStore.On("IsDeclarative", mock.Anything, testServiceAppID).Maybe().Return(false)
	mockLoadFullApplication(mockStore, service, existingApp)
	mockEP := resetIdentifyEntity(service)
	mockEP.On("IdentifyEntity",
		map[string]interface{}{"name": "New Name"}).
		Return((*string)(nil),
			entityprovider.NewEntityProviderError(
				entityprovider.ErrorCodeSystemError, "database error", ""))

	result, inboundAuth, svcErr := service.validateApplicationForUpdate(context.Background(), testServiceAppID, app)

	assert.Nil(suite.T(), result)
	assert.Nil(suite.T(), inboundAuth)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &serviceerror.InternalServerError, svcErr)
}

func (suite *ServiceTestSuite) TestValidateApplicationForUpdate_Success() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, mockStore := suite.setupTestService()

	existingApp := &model.ApplicationProcessedDTO{
		ID:   testServiceAppID,
		Name: "Test App",
	}

	app := &model.ApplicationDTO{
		Name:    "Test App",
		OUID:    testOUID,
		URL:     "https://example.com",
		LogoURL: "https://example.com/logo.png",
	}

	mockStore.On("IsDeclarative", mock.Anything, testServiceAppID).Maybe().Return(false)
	mockLoadFullApplication(mockStore, service, existingApp)

	result, inboundAuth, svcErr := service.validateApplicationForUpdate(context.Background(), testServiceAppID, app)

	assert.NotNil(suite.T(), result)
	assert.Nil(suite.T(), inboundAuth)
	assert.Nil(suite.T(), svcErr)
	assert.Equal(suite.T(), testServiceAppID, result.ID)
	assert.Equal(suite.T(), "Test App", result.Name)
}

func (suite *ServiceTestSuite) TestDeleteApplication_EmptyAppID() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, _ := suite.setupTestService()

	svcErr := service.DeleteApplication(context.Background(), "")

	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestDeleteApplication_NotFound() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, mockStore := suite.setupTestService()

	mockStore.On("DeleteInboundClient", mock.Anything, testServiceAppID).
		Return(inboundclient.ErrInboundClientNotFound)

	svcErr := service.DeleteApplication(context.Background(), testServiceAppID)

	// Should return nil (not error) when app not found
	assert.Nil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestDeleteApplication_StoreError() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, mockStore := suite.setupTestService()

	mockStore.On("DeleteInboundClient", mock.Anything, testServiceAppID).Return(errors.New("store error"))

	svcErr := service.DeleteApplication(context.Background(), testServiceAppID)

	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestDeleteApplication_Success() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, mockStore := suite.setupTestService()

	mockStore.On("DeleteInboundClient", mock.Anything, testServiceAppID).Return(nil)

	svcErr := service.DeleteApplication(context.Background(), testServiceAppID)

	assert.Nil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestDeleteApplication_CertError() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, mockStore := suite.setupTestService()

	mockStore.On("DeleteInboundClient", mock.Anything, testServiceAppID).
		Return(&inboundclient.CertOperationError{
			Operation:  inboundclient.CertOpDelete,
			RefType:    cert.CertificateReferenceTypeApplication,
			Underlying: &serviceerror.ServiceError{Type: serviceerror.ClientErrorType},
		})

	svcErr := service.DeleteApplication(context.Background(), testServiceAppID)

	assert.NotNil(suite.T(), svcErr)
}

// TestDeleteApplication_OAuthCertError verifies that when OAuth app certificate deletion fails,
// the error is properly propagated from DeleteApplication (covers deleteOAuthAppCertificate).
func (suite *ServiceTestSuite) TestDeleteApplication_OAuthCertError() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, mockStore := suite.setupTestService()

	mockStore.On("DeleteInboundClient", mock.Anything, testServiceAppID).
		Return(&inboundclient.CertOperationError{
			Operation:  inboundclient.CertOpDelete,
			RefType:    cert.CertificateReferenceTypeOAuthApp,
			Underlying: &serviceerror.ServiceError{Type: serviceerror.ServerErrorType, Code: "CERT-5001"},
		})

	svcErr := service.DeleteApplication(context.Background(), testServiceAppID)

	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &serviceerror.InternalServerError, svcErr)
}

// TestDeleteApplication_OAuthCertError_ClientError verifies that when OAuth app certificate deletion fails
// with a client error, the error is properly propagated from DeleteApplication.
func (suite *ServiceTestSuite) TestDeleteApplication_OAuthCertError_ClientError() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, mockStore := suite.setupTestService()

	mockStore.On("DeleteInboundClient", mock.Anything, testServiceAppID).
		Return(&inboundclient.CertOperationError{
			Operation: inboundclient.CertOpDelete,
			RefType:   cert.CertificateReferenceTypeOAuthApp,
			Underlying: &serviceerror.ServiceError{Type: serviceerror.ClientErrorType,
				Code: "CERT-1001", ErrorDescription: i18ncore.I18nMessage{DefaultValue: "Invalid client ID"}},
		})

	svcErr := service.DeleteApplication(context.Background(), testServiceAppID)

	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), ErrorCertificateClientError.Code, svcErr.Code)
	assert.Contains(suite.T(), svcErr.ErrorDescription.DefaultValue, "Failed to delete OAuth app certificate")
}

// TestDeleteApplication_WithOAuthCert_Success verifies successful deletion of an application with OAuth certificate.
// This test covers deleteOAuthAppCertificate's success path (return nil).
func (suite *ServiceTestSuite) TestDeleteApplication_WithOAuthCert_Success() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, mockStore := suite.setupTestService()

	mockStore.On("DeleteInboundClient", mock.Anything, testServiceAppID).Return(nil)

	svcErr := service.DeleteApplication(context.Background(), testServiceAppID)

	assert.Nil(suite.T(), svcErr)
}
func (suite *ServiceTestSuite) TestValidateOAuthParamsForCreateAndUpdate_EmptyInboundAuth() {
	app := &model.ApplicationDTO{
		Name: "Test App",
		OUID: testOUID,
	}

	result, svcErr := validateOAuthParamsForCreateAndUpdate(app)

	assert.Nil(suite.T(), result)
	assert.Nil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestValidateOAuthParamsForCreateAndUpdate_InvalidType() {
	app := &model.ApplicationDTO{
		Name: "Test App",
		OUID: testOUID,
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: "invalid_type",
			},
		},
	}

	result, svcErr := validateOAuthParamsForCreateAndUpdate(app)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestValidateOAuthParamsForCreateAndUpdate_NilOAuthConfig() {
	app := &model.ApplicationDTO{
		Name: "Test App",
		OUID: testOUID,
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type:           model.OAuthInboundAuthType,
				OAuthAppConfig: nil,
			},
		},
	}

	result, svcErr := validateOAuthParamsForCreateAndUpdate(app)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
}

func (suite *ServiceTestSuite) TestValidateOAuthParamsForCreateAndUpdate_WithDefaults() {
	app := &model.ApplicationDTO{
		Name: "Test App",
		OUID: testOUID,
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{},
					ResponseTypes:           []oauth2const.ResponseType{},
					TokenEndpointAuthMethod: "",
				},
			},
		},
	}

	result, svcErr := validateOAuthParamsForCreateAndUpdate(app)

	assert.NotNil(suite.T(), result)
	assert.Nil(suite.T(), svcErr)
	assert.Len(suite.T(), result.OAuthAppConfig.GrantTypes, 1)
	assert.Equal(suite.T(), oauth2const.GrantTypeAuthorizationCode, result.OAuthAppConfig.GrantTypes[0])
	assert.Equal(
		suite.T(),
		oauth2const.TokenEndpointAuthMethodClientSecretBasic,
		result.OAuthAppConfig.TokenEndpointAuthMethod,
	)
}

func (suite *ServiceTestSuite) TestValidateOAuthParamsForCreateAndUpdate_WithResponseTypeDefault() {
	app := &model.ApplicationDTO{
		Name: "Test App",
		OUID: testOUID,
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
					ResponseTypes:           []oauth2const.ResponseType{},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
				},
			},
		},
	}

	result, svcErr := validateOAuthParamsForCreateAndUpdate(app)

	assert.NotNil(suite.T(), result)
	assert.Nil(suite.T(), svcErr)
	assert.Len(suite.T(), result.OAuthAppConfig.ResponseTypes, 1)
	assert.Equal(suite.T(), oauth2const.ResponseTypeCode, result.OAuthAppConfig.ResponseTypes[0])
}

func (suite *ServiceTestSuite) TestValidateOAuthParamsForCreateAndUpdate_WithGrantTypeButNoResponseType() {
	app := &model.ApplicationDTO{
		Name: "Test App",
		OUID: testOUID,
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeClientCredentials},
					ResponseTypes:           []oauth2const.ResponseType{},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
				},
			},
		},
	}

	result, svcErr := validateOAuthParamsForCreateAndUpdate(app)

	assert.NotNil(suite.T(), result)
	assert.Nil(suite.T(), svcErr)
	assert.Len(suite.T(), result.OAuthAppConfig.ResponseTypes, 0)
}

func (suite *ServiceTestSuite) TestEnrichApplicationWithCertificate_Error() {
	service, mockStore := suite.setupTestService()

	app := &model.Application{
		ID:   testServiceAppID,
		Name: "Test App",
	}

	svcErr := &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		ErrorDescription: i18ncore.I18nMessage{DefaultValue: "Invalid certificate"},
	}

	mockStore.EXPECT().
		GetCertificate(mock.Anything, cert.CertificateReferenceTypeApplication, testServiceAppID).
		Return(nil, &inboundclient.CertOperationError{
			Operation:  inboundclient.CertOpRetrieve,
			RefType:    cert.CertificateReferenceTypeApplication,
			Underlying: svcErr,
		})

	result, err := service.enrichApplicationWithCertificate(context.Background(), app)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
}

func (suite *ServiceTestSuite) TestEnrichApplicationWithCertificate_Success() {
	service, mockStore := suite.setupTestService()

	app := &model.Application{
		ID:   testServiceAppID,
		Name: "Test App",
	}

	mockStore.EXPECT().
		GetCertificate(mock.Anything, cert.CertificateReferenceTypeApplication, testServiceAppID).
		Return(&inboundmodel.Certificate{
			Type:  cert.CertificateTypeJWKS,
			Value: `{"keys":[]}`,
		}, nil)

	result, err := service.enrichApplicationWithCertificate(context.Background(), app)

	assert.NotNil(suite.T(), result)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), cert.CertificateTypeJWKS, result.Certificate.Type)
}

func (suite *ServiceTestSuite) TestValidateOAuthParamsForCreateAndUpdate_PublicClientSuccess() {
	app := &model.ApplicationDTO{
		Name: "Test App",
		OUID: testOUID,
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodNone,
					PublicClient:            true,
					PKCERequired:            true,
				},
			},
		},
	}

	result, svcErr := validateOAuthParamsForCreateAndUpdate(app)

	assert.NotNil(suite.T(), result)
	assert.Nil(suite.T(), svcErr)
	assert.True(suite.T(), result.OAuthAppConfig.PublicClient)
}

func (suite *ServiceTestSuite) TestValidateApplication_StoreErrorNonNotFound() {
	service, _ := suite.setupTestService()

	app := &model.ApplicationDTO{
		Name: "Test App",
		OUID: testOUID,
	}

	// Return an entity provider error that's not EntityNotFound
	mockEP := resetIdentifyEntity(service)
	mockEP.On("IdentifyEntity",
		map[string]interface{}{"name": "Test App"}).
		Return((*string)(nil),
			entityprovider.NewEntityProviderError(
				entityprovider.ErrorCodeSystemError, "database connection error", ""))

	result, inboundAuth, svcErr := service.ValidateApplication(context.Background(), app)

	assert.Nil(suite.T(), result)
	assert.Nil(suite.T(), inboundAuth)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &serviceerror.InternalServerError, svcErr)
}

//nolint:dupl // Testing different URL validation scenarios
func (suite *ServiceTestSuite) TestValidateApplication_InvalidURL() {
	testConfig := &config.Config{}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, _ := suite.setupTestService()

	app := &model.ApplicationDTO{
		Name:       "Test App",
		OUID:       testOUID,
		URL:        "not-a-valid-uri",
		AuthFlowID: "edc013d0-e893-4dc0-990c-3e1d203e005b",
	}

	result, inboundAuth, svcErr := service.ValidateApplication(context.Background(), app)

	assert.Nil(suite.T(), result)
	assert.Nil(suite.T(), inboundAuth)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &ErrorInvalidApplicationURL, svcErr)
}

//nolint:dupl // Testing different URL validation scenarios
func (suite *ServiceTestSuite) TestValidateApplication_InvalidLogoURL() {
	testConfig := &config.Config{}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, _ := suite.setupTestService()

	app := &model.ApplicationDTO{
		Name:       "Test App",
		OUID:       testOUID,
		LogoURL:    "://invalid",
		AuthFlowID: "edc013d0-e893-4dc0-990c-3e1d203e005b",
	}

	result, inboundAuth, svcErr := service.ValidateApplication(context.Background(), app)

	assert.Nil(suite.T(), result)
	assert.Nil(suite.T(), inboundAuth)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &ErrorInvalidLogoURL, svcErr)
}

func (suite *ServiceTestSuite) TestCreateApplication_StoreErrorWithRollback() {
	suite.runCreateApplicationStoreErrorTest()
}

func (suite *ServiceTestSuite) TestCreateApplication_StoreErrorWithRollbackFailure() {
	// Currently identical to success case as rollback behavior is internal
	suite.runCreateApplicationStoreErrorTest()
}

func (suite *ServiceTestSuite) runCreateApplicationStoreErrorTest() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, mockStore := suite.setupTestService()

	app := &model.ApplicationDTO{
		Name:               "Test App",
		OUID:               testOUID,
		AuthFlowID:         "edc013d0-e893-4dc0-990c-3e1d203e005b",
		RegistrationFlowID: "80024fb3-29ed-4c33-aa48-8aee5e96d522",
		Certificate: &inboundmodel.Certificate{
			Type:  "JWKS",
			Value: `{"keys":[]}`,
		},
	}

	mockStore.On("CreateInboundClient",
		mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything).Return(errors.New("store error"))

	result, svcErr := service.CreateApplication(context.Background(), app)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &serviceerror.InternalServerError, svcErr)
}

func (suite *ServiceTestSuite) TestUpdateApplication_StoreErrorNonNotFound() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, mockStore := suite.setupTestService()

	app := &model.ApplicationDTO{
		Name: "Updated App",
		OUID: testOUID,
	}

	mockStore.On("IsDeclarative", mock.Anything, testServiceAppID).Maybe().Return(false)
	// Return an error that's not ApplicationNotFoundError
	mockStore.On("GetInboundClientByEntityID", mock.Anything, testServiceAppID).
		Return(nil, errors.New("database connection error"))

	result, svcErr := service.UpdateApplication(context.Background(), testServiceAppID, app)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &serviceerror.InternalServerError, svcErr)
}

func (suite *ServiceTestSuite) TestUpdateApplication_StoreErrorWhenCheckingName() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, mockStore := suite.setupTestService()

	existingApp := &model.ApplicationProcessedDTO{
		ID:   testServiceAppID,
		Name: "Old App",
	}

	app := &model.ApplicationDTO{
		Name: "New App",
		OUID: testOUID,
	}

	mockStore.On("IsDeclarative", mock.Anything, testServiceAppID).Maybe().Return(false)
	mockLoadFullApplication(mockStore, service, existingApp)
	// Return an entity provider error when checking name uniqueness
	mockEP := resetIdentifyEntity(service)
	mockEP.On("IdentifyEntity",
		map[string]interface{}{"name": "New App"}).
		Return((*string)(nil),
			entityprovider.NewEntityProviderError(
				entityprovider.ErrorCodeSystemError, "database connection error", ""))

	result, svcErr := service.UpdateApplication(context.Background(), testServiceAppID, app)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &serviceerror.InternalServerError, svcErr)
}

func (suite *ServiceTestSuite) TestUpdateApplication_StoreErrorWhenCheckingClientID() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, mockStore := suite.setupTestService()

	existingApp := &model.ApplicationProcessedDTO{
		ID:   testServiceAppID,
		Name: "Test App",
		InboundAuthConfig: []model.InboundAuthConfigProcessedDTO{
			{
				OAuthAppConfig: &inboundmodel.OAuthClient{
					ClientID: "old-client-id",
				},
			},
		},
	}

	app := &model.ApplicationDTO{
		Name: "Test App",
		OUID: testOUID,
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					ClientID:                "new-client-id",
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
				},
			},
		},
	}

	mockStore.On("IsDeclarative", mock.Anything, testServiceAppID).Maybe().Return(false)
	mockLoadFullApplication(mockStore, service, existingApp)
	// Return an entity provider error when checking client ID uniqueness
	mockEP := resetIdentifyEntity(service)
	mockEP.On("IdentifyEntity",
		map[string]interface{}{"clientId": "new-client-id"}).
		Return((*string)(nil),
			entityprovider.NewEntityProviderError(
				entityprovider.ErrorCodeSystemError, "database connection error", ""))

	result, svcErr := service.UpdateApplication(context.Background(), testServiceAppID, app)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &serviceerror.InternalServerError, svcErr)
}

func (suite *ServiceTestSuite) TestUpdateApplication_StoreErrorWithRollback() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, mockStore := suite.setupTestService()

	existingApp := &model.ApplicationProcessedDTO{
		ID:   testServiceAppID,
		Name: "Test App",
	}

	app := &model.ApplicationDTO{
		ID:                 testServiceAppID,
		Name:               "Test App",
		OUID:               testOUID,
		AuthFlowID:         "edc013d0-e893-4dc0-990c-3e1d203e005b",
		RegistrationFlowID: "80024fb3-29ed-4c33-aa48-8aee5e96d522",
		Certificate: &inboundmodel.Certificate{
			Type:  "JWKS",
			Value: `{"keys":[]}`,
		},
	}

	mockStore.On("IsDeclarative", mock.Anything, testServiceAppID).Maybe().Return(false)
	mockLoadFullApplication(mockStore, service, existingApp)
	mockStore.On("UpdateInboundClient",
		mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("store error"))

	result, svcErr := service.UpdateApplication(context.Background(), testServiceAppID, app)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &serviceerror.InternalServerError, svcErr)
}

func (suite *ServiceTestSuite) TestCreateApplication_ValidateApplicationError() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, _ := suite.setupTestService()

	app := &model.ApplicationDTO{
		Name: "", // Invalid name to trigger ValidateApplication error
		OUID: testOUID,
	}

	result, svcErr := service.CreateApplication(context.Background(), app)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &ErrorInvalidApplicationName, svcErr)
}

func (suite *ServiceTestSuite) TestCreateApplication_CertificateCreationError() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, mockStore := suite.setupTestService()

	app := &model.ApplicationDTO{
		Name: "Test App",
		OUID: testOUID,
		Certificate: &inboundmodel.Certificate{
			Type:  "JWKS",
			Value: `{"keys":[]}`,
		},
		AuthFlowID:         "auth-flow-id",
		RegistrationFlowID: "reg-flow-id",
	}

	svcErrExpected := &serviceerror.ServiceError{Type: serviceerror.ServerErrorType}
	mockStore.On("CreateInboundClient",
		mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything).
		Return(&inboundclient.CertOperationError{
			Operation:  inboundclient.CertOpCreate,
			RefType:    cert.CertificateReferenceTypeApplication,
			Underlying: svcErrExpected,
		})

	result, svcErr := service.CreateApplication(context.Background(), app)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &serviceerror.InternalServerError, svcErr)
}

func (suite *ServiceTestSuite) TestCreateApplication_WithOAuthCertificate_Success() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, mockStore := suite.setupTestService()

	app := &model.ApplicationDTO{
		Name:               "Test OAuth Cert App",
		OUID:               testOUID,
		AuthFlowID:         "auth-flow-id",
		RegistrationFlowID: "reg-flow-id",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					ClientID:                testClientID,
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodPrivateKeyJWT,
					Certificate: &inboundmodel.Certificate{
						Type:  "JWKS",
						Value: `{"keys":[]}`,
					},
				},
			},
		},
	}

	mockStore.On("CreateInboundClient",
		mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything).Return(nil)

	result, svcErr := service.CreateApplication(context.Background(), app)

	assert.NotNil(suite.T(), result)
	assert.Nil(suite.T(), svcErr)
	assert.Equal(suite.T(), "Test OAuth Cert App", result.Name)
	require.Len(suite.T(), result.InboundAuthConfig, 1)
	assert.Equal(suite.T(), model.OAuthInboundAuthType, result.InboundAuthConfig[0].Type)
	require.NotNil(suite.T(), result.InboundAuthConfig[0].OAuthAppConfig)
	require.NotNil(suite.T(), result.InboundAuthConfig[0].OAuthAppConfig.Certificate)
	assert.Equal(suite.T(), cert.CertificateType("JWKS"), result.InboundAuthConfig[0].OAuthAppConfig.Certificate.Type)
	assert.Equal(suite.T(), `{"keys":[]}`, result.InboundAuthConfig[0].OAuthAppConfig.Certificate.Value)
}

func (suite *ServiceTestSuite) TestCreateApplication_StoreErrorWithOAuthCertRollback() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, mockStore := suite.setupTestService()

	app := &model.ApplicationDTO{
		Name:               "Test OAuth Cert App",
		OUID:               testOUID,
		AuthFlowID:         "auth-flow-id",
		RegistrationFlowID: "reg-flow-id",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					ClientID:                testClientID,
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodPrivateKeyJWT,
					Certificate: &inboundmodel.Certificate{
						Type:  "JWKS",
						Value: `{"keys":[]}`,
					},
				},
			},
		},
	}

	// Store creation fails
	mockStore.On("CreateInboundClient",
		mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything).Return(errors.New("store error"))

	result, svcErr := service.CreateApplication(context.Background(), app)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &serviceerror.InternalServerError, svcErr)
}

func (suite *ServiceTestSuite) TestCreateApplication_StoreErrorWithBothAppAndOAuthCertRollback() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, mockStore := suite.setupTestService()

	app := &model.ApplicationDTO{
		Name:               "Test App With Both Certs",
		OUID:               testOUID,
		AuthFlowID:         "auth-flow-id",
		RegistrationFlowID: "reg-flow-id",
		Certificate: &inboundmodel.Certificate{
			Type:  "JWKS",
			Value: `{"keys":[{"app":"cert"}]}`,
		},
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					ClientID:                testClientID,
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodPrivateKeyJWT,
					Certificate: &inboundmodel.Certificate{
						Type:  "JWKS",
						Value: `{"keys":[{"oauth":"cert"}]}`,
					},
				},
			},
		},
	}

	// Store creation fails
	mockStore.On("CreateInboundClient",
		mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything).Return(errors.New("store error"))

	result, svcErr := service.CreateApplication(context.Background(), app)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &serviceerror.InternalServerError, svcErr)
}

func (suite *ServiceTestSuite) TestUpdateApplication_NotFound() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, mockStore := suite.setupTestService()

	app := &model.ApplicationDTO{
		Name: "New Name",
		OUID: testOUID,
	}

	mockStore.On("IsDeclarative", mock.Anything, testServiceAppID).Maybe().Return(false)
	mockStore.On("GetInboundClientByEntityID", mock.Anything, testServiceAppID).
		Return(nil, inboundclient.ErrInboundClientNotFound)

	result, svcErr := service.UpdateApplication(context.Background(), testServiceAppID, app)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &ErrorApplicationNotFound, svcErr)
}

func (suite *ServiceTestSuite) TestUpdateApplication_NameConflict() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, mockStore := suite.setupTestService()

	existingApp := &model.ApplicationProcessedDTO{
		ID:   testServiceAppID,
		Name: "Old Name",
	}

	app := &model.ApplicationDTO{
		Name: "New Name",
		OUID: testOUID,
	}

	mockStore.On("IsDeclarative", mock.Anything, testServiceAppID).Maybe().Return(false)
	mockLoadFullApplication(mockStore, service, existingApp)
	mockEP := resetIdentifyEntity(service)
	conflictingID := testConflictingAppID
	mockEP.On("IdentifyEntity",
		map[string]interface{}{"name": "New Name"}).
		Return(
			&conflictingID, (*entityprovider.EntityProviderError)(nil))

	result, svcErr := service.UpdateApplication(context.Background(), testServiceAppID, app)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &ErrorApplicationAlreadyExistsWithName, svcErr)
}

func (suite *ServiceTestSuite) TestUpdateApplication_MetadataUpdate() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, mockStore := suite.setupTestService()

	existingApp := &model.ApplicationProcessedDTO{
		ID:                 testServiceAppID,
		Name:               "Test App",
		AuthFlowID:         "default-auth-flow",
		RegistrationFlowID: "default-reg-flow",
		Metadata: map[string]interface{}{
			"old_key": "old_value",
		},
	}

	updatedApp := &model.ApplicationDTO{
		Name:               "Test App",
		OUID:               testOUID,
		AuthFlowID:         "default-auth-flow",
		RegistrationFlowID: "default-reg-flow",
		Metadata: map[string]interface{}{
			"new_key":     "new_value",
			"another_key": "another_value",
		},
	}

	mockStore.On("IsDeclarative", mock.Anything, testServiceAppID).Maybe().Return(false)
	mockLoadFullApplication(mockStore, service, existingApp)
	mockStore.On("UpdateInboundClient",
		mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	result, svcErr := service.UpdateApplication(context.Background(), testServiceAppID, updatedApp)

	assert.NotNil(suite.T(), result)
	assert.Nil(suite.T(), svcErr)
	assert.Equal(suite.T(), "new_value", result.Metadata["new_key"])
	assert.Equal(suite.T(), "another_value", result.Metadata["another_key"])
	mockStore.AssertExpectations(suite.T())
}

// TestUpdateApplication_AppCertificateUpdateError verifies that when the app certificate update fails
// inside the transaction, UpdateApplication returns the certificate error.
func (suite *ServiceTestSuite) TestUpdateApplication_AppCertificateUpdateError() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{Enabled: false},
		JWT:                  config.JWTConfig{ValidityPeriod: 3600},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, mockStore := suite.setupTestService()

	existingApp := &model.ApplicationProcessedDTO{
		ID:   testServiceAppID,
		Name: "Test App",
	}
	app := &model.ApplicationDTO{
		Name:               "Test App",
		OUID:               testOUID,
		AuthFlowID:         "edc013d0-e893-4dc0-990c-3e1d203e005b",
		RegistrationFlowID: "80024fb3-29ed-4c33-aa48-8aee5e96d522",
	}

	certServerError := &serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "CERT-5001",
		Error:            i18ncore.I18nMessage{DefaultValue: "Database error"},
		ErrorDescription: i18ncore.I18nMessage{DefaultValue: "Failed to retrieve certificate from database"},
	}

	mockStore.On("IsDeclarative", mock.Anything, testServiceAppID).Maybe().Return(false)
	mockLoadFullApplication(mockStore, service, existingApp)
	mockStore.On("UpdateInboundClient",
		mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&inboundclient.CertOperationError{
			Operation:  inboundclient.CertOpRetrieve,
			RefType:    cert.CertificateReferenceTypeApplication,
			Underlying: certServerError,
		})

	result, svcErr := service.UpdateApplication(context.Background(), testServiceAppID, app)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &serviceerror.InternalServerError, svcErr)
}

// TestResolveClientSecret_PublicClient tests that no secret is generated for public clients.
func TestResolveClientSecret_PublicClient(t *testing.T) {
	inboundAuthConfig := &model.InboundAuthConfigDTO{
		OAuthAppConfig: &model.OAuthAppConfigDTO{
			TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodNone,
			ClientSecret:            "",
			PublicClient:            true,
		},
	}

	err := resolveClientSecret(inboundAuthConfig, nil)

	assert.Nil(t, err)
	assert.Equal(t, "", inboundAuthConfig.OAuthAppConfig.ClientSecret)
}

// TestResolveClientSecret_SecretAlreadyProvided tests that existing secrets are not overwritten.
func TestResolveClientSecret_SecretAlreadyProvided(t *testing.T) {
	providedSecret := "user-provided-secret"
	inboundAuthConfig := &model.InboundAuthConfigDTO{
		OAuthAppConfig: &model.OAuthAppConfigDTO{
			TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
			ClientSecret:            providedSecret,
			PublicClient:            false,
		},
	}

	err := resolveClientSecret(inboundAuthConfig, nil)

	assert.Nil(t, err)
	assert.Equal(t, providedSecret, inboundAuthConfig.OAuthAppConfig.ClientSecret)
}

// TestResolveClientSecret_GenerateForNewConfidentialClient tests secret generation for new clients.
func TestResolveClientSecret_GenerateForNewConfidentialClient(t *testing.T) {
	inboundAuthConfig := &model.InboundAuthConfigDTO{
		OAuthAppConfig: &model.OAuthAppConfigDTO{
			TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
			ClientSecret:            "",
			PublicClient:            false,
		},
	}

	err := resolveClientSecret(inboundAuthConfig, nil)

	assert.Nil(t, err)
	assert.NotEmpty(t, inboundAuthConfig.OAuthAppConfig.ClientSecret)
	// Verify it's a valid OAuth2 secret (should be non-empty and have sufficient length)
	assert.Greater(t, len(inboundAuthConfig.OAuthAppConfig.ClientSecret), 20)
}

// TestResolveClientSecret_PreserveExistingSecret tests that existing secrets are preserved during updates.
func TestResolveClientSecret_PreserveExistingSecret(t *testing.T) {
	existingApp := &model.ApplicationProcessedDTO{
		InboundAuthConfig: []model.InboundAuthConfigProcessedDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &inboundmodel.OAuthClient{
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
					PublicClient:            false,
				},
			},
		},
	}

	inboundAuthConfig := &model.InboundAuthConfigDTO{
		OAuthAppConfig: &model.OAuthAppConfigDTO{
			TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
			ClientSecret:            "",
			PublicClient:            false,
		},
	}

	err := resolveClientSecret(inboundAuthConfig, existingApp)

	assert.Nil(t, err)
	// Secret should remain empty (not generated) because existing app has a secret
	assert.Equal(t, "", inboundAuthConfig.OAuthAppConfig.ClientSecret)
}

// TestResolveClientSecret_NoExistingApp tests secret generation when no existing app.
func TestResolveClientSecret_NoExistingApp(t *testing.T) {
	inboundAuthConfig := &model.InboundAuthConfigDTO{
		OAuthAppConfig: &model.OAuthAppConfigDTO{
			TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
			ClientSecret:            "",
			PublicClient:            false,
		},
	}

	err := resolveClientSecret(inboundAuthConfig, nil)

	assert.Nil(t, err)
	assert.NotEmpty(t, inboundAuthConfig.OAuthAppConfig.ClientSecret)
}

// TestResolveClientSecret_ExistingAppWithoutSecret tests secret generation when existing app has no secret.
func TestResolveClientSecret_ExistingAppWithoutSecret(t *testing.T) {
	existingApp := &model.ApplicationProcessedDTO{
		InboundAuthConfig: []model.InboundAuthConfigProcessedDTO{
			{
				OAuthAppConfig: &inboundmodel.OAuthClient{
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodNone,
					PublicClient:            false,
				},
			},
		},
	}

	inboundAuthConfig := &model.InboundAuthConfigDTO{
		OAuthAppConfig: &model.OAuthAppConfigDTO{
			TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
			ClientSecret:            "",
			PublicClient:            false,
		},
	}

	err := resolveClientSecret(inboundAuthConfig, existingApp)

	assert.Nil(t, err)
	// Should generate a new secret since existing app doesn't have one
	assert.NotEmpty(t, inboundAuthConfig.OAuthAppConfig.ClientSecret)
}

// TestUpdateApplication_StoreFails_RollbackCertFails verifies that when the store update fails
// and rolling back the certificate also fails, the rollback error is returned.
func (suite *ServiceTestSuite) TestUpdateApplication_StoreFails_RollbackCertFails() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{Enabled: false},
		JWT:                  config.JWTConfig{ValidityPeriod: 3600},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, mockStore := suite.setupTestService()
	existingApp := &model.ApplicationProcessedDTO{
		ID:   "app123",
		Name: "Test App",
	}
	app := &model.ApplicationDTO{
		ID:                 "app123",
		Name:               "Test App",
		OUID:               testOUID,
		AuthFlowID:         "edc013d0-e893-4dc0-990c-3e1d203e005b",
		RegistrationFlowID: "80024fb3-29ed-4c33-aa48-8aee5e96d522",
	}

	mockStore.On("IsDeclarative", mock.Anything, "app123").Maybe().Return(false)
	mockLoadFullApplication(mockStore, service, existingApp)
	mockStore.On("UpdateInboundClient",
		mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("store error"))

	result, svcErr := service.UpdateApplication(context.Background(), "app123", app)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &serviceerror.InternalServerError, svcErr)
}

// TestUpdateApplication_WithOAuthConfig_Success tests successful update of an application with OAuth configuration.
func (suite *ServiceTestSuite) TestUpdateApplication_WithOAuthConfig_Success() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
		JWT: config.JWTConfig{
			ValidityPeriod: 3600,
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, mockStore := suite.setupTestService()

	existingApp := &model.ApplicationProcessedDTO{
		ID:                 testServiceAppID,
		Name:               "Test App",
		AuthFlowID:         "auth-flow-id",
		RegistrationFlowID: "reg-flow-id",
		InboundAuthConfig: []model.InboundAuthConfigProcessedDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &inboundmodel.OAuthClient{
					ClientID:                testClientID,
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
				},
			},
		},
	}

	updatedApp := &model.ApplicationDTO{
		ID:                 testServiceAppID,
		Name:               "Test App Updated",
		OUID:               testOUID,
		AuthFlowID:         "auth-flow-id",
		RegistrationFlowID: "reg-flow-id",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					ClientID: testClientID,
					RedirectURIs: []string{"https://example.com/callback",
						"https://example.com/callback2"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
				},
			},
		},
	}

	mockStore.On("IsDeclarative", mock.Anything, testServiceAppID).Maybe().Return(false)
	mockLoadFullApplication(mockStore, service, existingApp)

	mockStore.On("UpdateInboundClient",
		mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	result, svcErr := service.UpdateApplication(context.Background(), testServiceAppID, updatedApp)

	assert.NotNil(suite.T(), result)
	assert.Nil(suite.T(), svcErr)
	assert.Equal(suite.T(), "Test App Updated", result.Name)
	require.Len(suite.T(), result.InboundAuthConfig, 1)
	assert.Equal(suite.T(), testClientID, result.InboundAuthConfig[0].OAuthAppConfig.ClientID)
	assert.Len(suite.T(), result.InboundAuthConfig[0].OAuthAppConfig.RedirectURIs, 2)
	mockStore.AssertExpectations(suite.T())
}

// TestUpdateApplication_AddOAuthConfig_Success tests adding OAuth configuration to an app that didn't have it.
func (suite *ServiceTestSuite) TestUpdateApplication_AddOAuthConfig_Success() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
		JWT: config.JWTConfig{
			ValidityPeriod: 3600,
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, mockStore := suite.setupTestService()

	existingApp := &model.ApplicationProcessedDTO{
		ID:                 testServiceAppID,
		Name:               "Test App",
		AuthFlowID:         "auth-flow-id",
		RegistrationFlowID: "reg-flow-id",
		InboundAuthConfig:  []model.InboundAuthConfigProcessedDTO{}, // No OAuth config initially
	}

	updatedApp := &model.ApplicationDTO{
		ID:                 testServiceAppID,
		Name:               "Test App",
		OUID:               testOUID,
		AuthFlowID:         "auth-flow-id",
		RegistrationFlowID: "reg-flow-id",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					ClientID:                "new-client-id",
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
				},
			},
		},
	}

	mockStore.On("IsDeclarative", mock.Anything, testServiceAppID).Maybe().Return(false)
	mockLoadFullApplication(mockStore, service, existingApp)

	mockStore.On("UpdateInboundClient",
		mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	result, svcErr := service.UpdateApplication(context.Background(), testServiceAppID, updatedApp)

	assert.NotNil(suite.T(), result)
	assert.Nil(suite.T(), svcErr)
	require.Len(suite.T(), result.InboundAuthConfig, 1)
	assert.Equal(suite.T(), "new-client-id", result.InboundAuthConfig[0].OAuthAppConfig.ClientID)
	mockStore.AssertExpectations(suite.T())
}

// TestUpdateApplication_UpdateOAuthClientID_Success tests changing the OAuth client ID.
func (suite *ServiceTestSuite) TestUpdateApplication_UpdateOAuthClientID_Success() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
		JWT: config.JWTConfig{
			ValidityPeriod: 3600,
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, mockStore := suite.setupTestService()

	existingApp := &model.ApplicationProcessedDTO{
		ID:                 testServiceAppID,
		Name:               "Test App",
		AuthFlowID:         "auth-flow-id",
		RegistrationFlowID: "reg-flow-id",
		InboundAuthConfig: []model.InboundAuthConfigProcessedDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &inboundmodel.OAuthClient{
					ClientID:                "old-client-id",
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
				},
			},
		},
	}

	updatedApp := &model.ApplicationDTO{
		ID:                 testServiceAppID,
		Name:               "Test App",
		OUID:               testOUID,
		AuthFlowID:         "auth-flow-id",
		RegistrationFlowID: "reg-flow-id",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					ClientID:                "new-client-id",
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
				},
			},
		},
	}

	mockStore.On("IsDeclarative", mock.Anything, testServiceAppID).Maybe().Return(false)
	mockLoadFullApplication(mockStore, service, existingApp)

	mockStore.On("UpdateInboundClient",
		mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	result, svcErr := service.UpdateApplication(context.Background(), testServiceAppID, updatedApp)

	assert.NotNil(suite.T(), result)
	assert.Nil(suite.T(), svcErr)
	require.Len(suite.T(), result.InboundAuthConfig, 1)
	assert.Equal(suite.T(), "new-client-id", result.InboundAuthConfig[0].OAuthAppConfig.ClientID)
	mockStore.AssertExpectations(suite.T())
}

func (suite *ServiceTestSuite) runUpdateApplicationWithJWKSCert(jwksValue string) {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
		JWT: config.JWTConfig{
			ValidityPeriod: 3600,
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, mockStore := suite.setupTestService()

	existingApp := &model.ApplicationProcessedDTO{
		ID:                 testServiceAppID,
		Name:               "Test App",
		AuthFlowID:         "auth-flow-id",
		RegistrationFlowID: "reg-flow-id",
		InboundAuthConfig: []model.InboundAuthConfigProcessedDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &inboundmodel.OAuthClient{
					ClientID:                testClientID,
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodPrivateKeyJWT,
				},
			},
		},
	}

	updatedApp := &model.ApplicationDTO{
		ID:                 testServiceAppID,
		Name:               "Test App",
		OUID:               testOUID,
		AuthFlowID:         "auth-flow-id",
		RegistrationFlowID: "reg-flow-id",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					ClientID:                testClientID,
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodPrivateKeyJWT,
					Certificate: &inboundmodel.Certificate{
						Type:  cert.CertificateTypeJWKS,
						Value: jwksValue,
					},
				},
			},
		},
	}

	mockStore.On("IsDeclarative", mock.Anything, testServiceAppID).Maybe().Return(false)
	mockLoadFullApplication(mockStore, service, existingApp)

	mockStore.On("UpdateInboundClient",
		mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	result, svcErr := service.UpdateApplication(context.Background(), testServiceAppID, updatedApp)

	assert.NotNil(suite.T(), result)
	assert.Nil(suite.T(), svcErr)
	require.Len(suite.T(), result.InboundAuthConfig, 1)
	assert.NotNil(suite.T(), result.InboundAuthConfig[0].OAuthAppConfig.Certificate)
	assert.Equal(suite.T(), cert.CertificateTypeJWKS, result.InboundAuthConfig[0].OAuthAppConfig.Certificate.Type)
	mockStore.AssertExpectations(suite.T())
}

// TestUpdateApplication_WithOAuthCertificate_Success tests updating an application with a new OAuth certificate.
func (suite *ServiceTestSuite) TestUpdateApplication_WithOAuthCertificate_Success() {
	suite.runUpdateApplicationWithJWKSCert(`{"keys":[{"kty":"RSA"}]}`)
}

// TestUpdateApplication_UpdateOAuthCertificate_Success tests updating an application with a replaced OAuth certificate.
func (suite *ServiceTestSuite) TestUpdateApplication_UpdateOAuthCertificate_Success() {
	suite.runUpdateApplicationWithJWKSCert(`{"keys":[{"kty":"RSA","n":"new-value"}]}`)
}

// TestUpdateApplication_OAuthClientIDConflict tests when the new client ID already exists.
func (suite *ServiceTestSuite) TestUpdateApplication_OAuthClientIDConflict() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, mockStore := suite.setupTestService()

	existingApp := &model.ApplicationProcessedDTO{
		ID:                 testServiceAppID,
		Name:               "Test App",
		AuthFlowID:         "auth-flow-id",
		RegistrationFlowID: "reg-flow-id",
		InboundAuthConfig: []model.InboundAuthConfigProcessedDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &inboundmodel.OAuthClient{
					ClientID:                "old-client-id",
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
				},
			},
		},
	}

	updatedApp := &model.ApplicationDTO{
		ID:                 testServiceAppID,
		Name:               "Test App",
		OUID:               testOUID,
		AuthFlowID:         "auth-flow-id",
		RegistrationFlowID: "reg-flow-id",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					ClientID:                "existing-client-id",
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
				},
			},
		},
	}

	mockStore.On("IsDeclarative", mock.Anything, testServiceAppID).Maybe().Return(false)
	mockLoadFullApplication(mockStore, service, existingApp)

	// Mock that another app already has this client ID via entity provider.
	mockEP := resetIdentifyEntity(service)
	conflictingEntityID := testConflictingAppID
	mockEP.On("IdentifyEntity",
		map[string]interface{}{"clientId": "existing-client-id"}).
		Return(
			&conflictingEntityID, (*entityprovider.EntityProviderError)(nil))

	result, svcErr := service.UpdateApplication(context.Background(), testServiceAppID, updatedApp)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &ErrorApplicationAlreadyExistsWithClientID, svcErr)
}

// TestUpdateApplication_OAuthInvalidRedirectURI tests updating with an invalid redirect URI.

// TestUpdateApplication_OAuthCertUpdateError tests when certificate update fails.
func (suite *ServiceTestSuite) TestUpdateApplication_OAuthCertUpdateError() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
		JWT: config.JWTConfig{
			ValidityPeriod: 3600,
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, mockStore := suite.setupTestService()

	existingApp := &model.ApplicationProcessedDTO{
		ID:                 testServiceAppID,
		Name:               "Test App",
		AuthFlowID:         "auth-flow-id",
		RegistrationFlowID: "reg-flow-id",
		InboundAuthConfig: []model.InboundAuthConfigProcessedDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &inboundmodel.OAuthClient{
					ClientID:                testClientID,
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodPrivateKeyJWT,
				},
			},
		},
	}

	updatedApp := &model.ApplicationDTO{
		ID:                 testServiceAppID,
		Name:               "Test App",
		OUID:               testOUID,
		AuthFlowID:         "auth-flow-id",
		RegistrationFlowID: "reg-flow-id",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					ClientID:                testClientID,
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodPrivateKeyJWT,
					Certificate: &inboundmodel.Certificate{
						Type:  cert.CertificateTypeJWKS,
						Value: `{"keys":[{"kty":"RSA"}]}`,
					},
				},
			},
		},
	}

	certError := &serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "CERT-500",
		Error:            i18ncore.I18nMessage{DefaultValue: "Internal certificate error"},
		ErrorDescription: i18ncore.I18nMessage{DefaultValue: "Failed to retrieve certificate"},
	}

	mockStore.On("IsDeclarative", mock.Anything, testServiceAppID).Maybe().Return(false)
	mockLoadFullApplication(mockStore, service, existingApp)
	mockStore.On("UpdateInboundClient",
		mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&inboundclient.CertOperationError{
			Operation:  inboundclient.CertOpRetrieve,
			RefType:    cert.CertificateReferenceTypeOAuthApp,
			Underlying: certError,
		})

	result, svcErr := service.UpdateApplication(context.Background(), testServiceAppID, updatedApp)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &serviceerror.InternalServerError, svcErr)
}

// TestUpdateApplication_OAuthStoreErrorWithRollback tests when store update fails with OAuth cert rollback.
func (suite *ServiceTestSuite) TestUpdateApplication_OAuthStoreErrorWithRollback() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
		JWT: config.JWTConfig{
			ValidityPeriod: 3600,
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, mockStore := suite.setupTestService()

	existingApp := &model.ApplicationProcessedDTO{
		ID:                 testServiceAppID,
		Name:               "Test App",
		AuthFlowID:         "auth-flow-id",
		RegistrationFlowID: "reg-flow-id",
		InboundAuthConfig: []model.InboundAuthConfigProcessedDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &inboundmodel.OAuthClient{
					ClientID:                testClientID,
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodPrivateKeyJWT,
				},
			},
		},
	}

	updatedApp := &model.ApplicationDTO{
		ID:                 testServiceAppID,
		Name:               "Test App",
		OUID:               testOUID,
		AuthFlowID:         "auth-flow-id",
		RegistrationFlowID: "reg-flow-id",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					ClientID:                testClientID,
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodPrivateKeyJWT,
					Certificate: &inboundmodel.Certificate{
						Type:  cert.CertificateTypeJWKS,
						Value: `{"keys":[{"kty":"RSA"}]}`,
					},
				},
			},
		},
	}

	mockStore.On("IsDeclarative", mock.Anything, testServiceAppID).Maybe().Return(false)
	mockLoadFullApplication(mockStore, service, existingApp)

	// Mock store update failure
	mockStore.On("UpdateInboundClient",
		mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("store error"))

	result, svcErr := service.UpdateApplication(context.Background(), testServiceAppID, updatedApp)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &serviceerror.InternalServerError, svcErr)
}

// TestUpdateApplication_OAuthTokenConfigUpdate tests updating OAuth token configuration.
func (suite *ServiceTestSuite) TestUpdateApplication_OAuthTokenConfigUpdate() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
		JWT: config.JWTConfig{
			ValidityPeriod: 3600,
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, mockStore := suite.setupTestService()

	existingApp := &model.ApplicationProcessedDTO{
		ID:                 testServiceAppID,
		Name:               "Test App",
		AuthFlowID:         "auth-flow-id",
		RegistrationFlowID: "reg-flow-id",
		InboundAuthConfig: []model.InboundAuthConfigProcessedDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &inboundmodel.OAuthClient{
					ClientID:                testClientID,
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
				},
			},
		},
	}

	updatedApp := &model.ApplicationDTO{
		ID:                 testServiceAppID,
		Name:               "Test App",
		OUID:               testOUID,
		AuthFlowID:         "auth-flow-id",
		RegistrationFlowID: "reg-flow-id",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					ClientID:                testClientID,
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
					Token: &inboundmodel.OAuthTokenConfig{
						AccessToken: &inboundmodel.AccessTokenConfig{
							ValidityPeriod: 7200,
							UserAttributes: []string{"email", "name"},
						},
						IDToken: &inboundmodel.IDTokenConfig{
							ValidityPeriod: 3600,
							UserAttributes: []string{"sub", "email"},
						},
					},
				},
			},
		},
	}

	mockStore.On("IsDeclarative", mock.Anything, testServiceAppID).Maybe().Return(false)
	mockLoadFullApplication(mockStore, service, existingApp)

	mockStore.On("UpdateInboundClient",
		mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	result, svcErr := service.UpdateApplication(context.Background(), testServiceAppID, updatedApp)

	assert.NotNil(suite.T(), result)
	assert.Nil(suite.T(), svcErr)
	require.Len(suite.T(), result.InboundAuthConfig, 1)
	assert.NotNil(suite.T(), result.InboundAuthConfig[0].OAuthAppConfig.Token)
	assert.Equal(suite.T(), int64(7200), result.InboundAuthConfig[0].OAuthAppConfig.Token.AccessToken.ValidityPeriod)
	assert.Equal(suite.T(), int64(3600), result.InboundAuthConfig[0].OAuthAppConfig.Token.IDToken.ValidityPeriod)
	mockStore.AssertExpectations(suite.T())
}

func (suite *ServiceTestSuite) TestCreateApplication_NilApplication() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: false,
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, _ := suite.setupTestService()

	result, svcErr := service.CreateApplication(context.Background(), nil)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &ErrorApplicationNil, svcErr)
}

func (suite *ServiceTestSuite) TestCreateApplication_DeclarativeMode() {
	testConfig := &config.Config{
		DeclarativeResources: config.DeclarativeResources{
			Enabled: true,
		},
	}
	config.ResetServerRuntime()
	err := config.InitializeServerRuntime("/tmp/test", testConfig)
	require.NoError(suite.T(), err)
	defer config.ResetServerRuntime()

	service, _ := suite.setupTestService()

	app := &model.ApplicationDTO{
		Name: "Test App",
		OUID: testOUID,
	}

	result, svcErr := service.CreateApplication(context.Background(), app)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &ErrorCannotModifyDeclarativeResource, svcErr)
}

// TestValidateApplication_ErrorFromProcessInboundAuthConfig tests error from
// processInboundAuthConfig when invalid inbound auth config is provided.
func (suite *ServiceTestSuite) TestValidateApplication_ErrorFromProcessInboundAuthConfig() {
	service, _ := suite.setupTestService()

	app := &model.ApplicationDTO{
		Name: "Test App",
		OUID: testOUID,
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: "InvalidType", // Invalid type, not OAuth
			},
		},
	}

	result, inboundAuth, svcErr := service.ValidateApplication(context.Background(), app)

	assert.Nil(suite.T(), result)
	assert.Nil(suite.T(), inboundAuth)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), &ErrorInvalidInboundAuthConfig, svcErr)
}

// TestValidateApplication_ErrorFromValidateAuthFlowID tests error from validateAuthFlowID
// when an invalid auth flow ID is provided.
