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

package export

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/idp"
	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/system/cmodels"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/tests/mocks/applicationmock"
	"github.com/asgardeo/thunder/tests/mocks/idp/idpmock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	testAppID  = "test-app-id"
	testIDPID  = "test-idp-id"
	testApp1ID = "app1"
	testApp2ID = "app2"
	testApp3ID = "app3"
)

// ExportServiceTestSuite defines the test suite for the export service.
type ExportServiceTestSuite struct {
	suite.Suite
	appServiceMock *applicationmock.ApplicationServiceInterfaceMock
	idpServiceMock *idpmock.IDPServiceInterfaceMock
	exportService  ExportServiceInterface
}

// SetupTest sets up the test environment before each test.
func (suite *ExportServiceTestSuite) SetupTest() {
	// Create temporary directory and crypto key file
	tempDir := suite.T().TempDir()
	cryptoFile := filepath.Join(tempDir, "crypto.key")
	dummyCryptoKey := "0579f866ac7c9273580d0ff163fa01a7b2401a7ff3ddc3e3b14ae3136fa6025e"

	err := os.WriteFile(cryptoFile, []byte(dummyCryptoKey), 0600)
	if err != nil {
		suite.T().Fatalf("Failed to create crypto key file: %v", err)
	}

	// Initialize ThunderRuntime with immutable mode disabled
	// Use just the filename since InitializeThunderRuntime will prepend the base path
	config.ResetThunderRuntime()
	testConfig := &config.Config{
		Security: config.SecurityConfig{
			CryptoFile: "crypto.key",
		},
		ImmutableResources: config.ImmutableResources{
			Enabled: false,
		},
	}
	_ = config.InitializeThunderRuntime(tempDir, testConfig)

	suite.appServiceMock = applicationmock.NewApplicationServiceInterfaceMock(suite.T())
	suite.idpServiceMock = idpmock.NewIDPServiceInterfaceMock(suite.T())

	// Create parameterizer instance
	parameterizer := newParameterizer(rules)

	suite.exportService = newExportService(suite.appServiceMock, suite.idpServiceMock, parameterizer)
}

func (suite *ExportServiceTestSuite) TearDownTest() {
	config.ResetThunderRuntime()
}

// TestExportServiceTestSuite runs the test suite.
func TestExportServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ExportServiceTestSuite))
}

// TestExportResources_NilRequest tests ExportResources with nil request.
func (suite *ExportServiceTestSuite) TestExportResources_NilRequest() {
	result, err := suite.exportService.ExportResources(nil)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrorInvalidRequest.Code, err.Code)
	assert.Equal(suite.T(), "Invalid export request", err.Error)
}

// TestExportResources_EmptyRequest tests ExportResources with empty request.
func (suite *ExportServiceTestSuite) TestExportResources_EmptyRequest() {
	request := &ExportRequest{}

	result, err := suite.exportService.ExportResources(request)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrorNoResourcesFound.Code, err.Code)
	assert.Equal(suite.T(), "No resources found", err.Error)
}

// TestExportResources_DefaultOptions tests ExportResources with default options.
func (suite *ExportServiceTestSuite) TestExportResources_DefaultOptions() {
	appID := testAppID
	request := &ExportRequest{
		Applications: []string{appID},
	}

	mockApp := &appmodel.Application{
		ID:          appID,
		Name:        "Test App",
		Description: "Test Description",
	}

	suite.appServiceMock.EXPECT().GetApplication(appID).Return(mockApp, nil)

	result, err := suite.exportService.ExportResources(request)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Files, 1)
	assert.Equal(suite.T(), 1, result.Summary.TotalFiles)
	assert.Contains(suite.T(), result.Summary.ResourceTypes, "applications")
}

// TestExportResources_ApplicationNotFound tests ExportResources when application is not found.
func (suite *ExportServiceTestSuite) TestExportResources_ApplicationNotFound() {
	appID := "non-existent-app"
	request := &ExportRequest{
		Applications: []string{appID},
		Options: &ExportOptions{
			Format: "yaml",
		},
	}

	appError := &serviceerror.ServiceError{
		Code:  "APP_NOT_FOUND",
		Error: "Application not found",
	}

	suite.appServiceMock.EXPECT().GetApplication(appID).Return(nil, appError)

	result, err := suite.exportService.ExportResources(request)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrorNoResourcesFound.Code, err.Code)
}

// TestExportResources_CompleteOAuthApplication tests exporting an application with OAuth config.
func (suite *ExportServiceTestSuite) TestExportResources_CompleteOAuthApplication() {
	appID := "oauth-app-id"
	request := &ExportRequest{
		Applications: []string{appID},
		Options: &ExportOptions{
			Format: "yaml",
			FolderStructure: &FolderStructureOptions{
				GroupByType:       true,
				FileNamingPattern: "${name}_${id}",
			},
		},
	}

	mockOAuthConfig := &appmodel.OAuthAppConfigComplete{
		ClientID:                "client123",
		RedirectURIs:            []string{"http://localhost:3000/callback"},
		GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
		ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
		TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretPost,
		PKCERequired:            true,
		PublicClient:            false,
		Scopes:                  []string{"openid", "profile"},
		Token: &appmodel.OAuthTokenConfig{
			Issuer: "https://localhost:8090",
			AccessToken: &appmodel.AccessTokenConfig{
				ValidityPeriod: 3600,
				UserAttributes: []string{"email", "username"},
			},
			IDToken: &appmodel.IDTokenConfig{
				ValidityPeriod: 1800,
				UserAttributes: []string{"email"},
				ScopeClaims: map[string][]string{
					"profile": {"name", "picture"},
				},
			},
		},
	}

	mockApp := &appmodel.Application{
		ID:          appID,
		Name:        "OAuth Test App",
		Description: "OAuth Test Description",
		URL:         "https://example.com",
		InboundAuthConfig: []appmodel.InboundAuthConfigComplete{
			{
				Type:           appmodel.OAuthInboundAuthType,
				OAuthAppConfig: mockOAuthConfig,
			},
		},
		Token: &appmodel.TokenConfig{
			UserAttributes: []string{"email", "username"},
		},
	}

	suite.appServiceMock.EXPECT().GetApplication(appID).Return(mockApp, nil)

	result, err := suite.exportService.ExportResources(request)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Files, 1)

	file := result.Files[0]
	assert.Equal(suite.T(), "OAuth_Test_App_oauth-app-id.yaml", file.FileName)
	assert.Equal(suite.T(), "applications", file.FolderPath)
	assert.Contains(suite.T(), file.Content, "name: OAuth Test App")
	assert.Contains(suite.T(), file.Content, "client_id: {{.O_AUTH_TEST_APP_CLIENT_ID}}")
	assert.Contains(suite.T(), file.Content, "client_secret: {{.O_AUTH_TEST_APP_CLIENT_SECRET}}")
	assert.Contains(suite.T(), file.Content, "redirect_uris:")
	assert.Contains(suite.T(), file.Content, "{{- range .O_AUTH_TEST_APP_REDIRECT_URIS}}")

	assert.Equal(suite.T(), 1, result.Summary.ResourceTypes["applications"])
	assert.Equal(suite.T(), int64(len(file.Content)), file.Size)
}

// TestExportResources_MultipleApplications tests exporting multiple applications.
func (suite *ExportServiceTestSuite) TestExportResources_MultipleApplications() {
	request := &ExportRequest{
		Applications: []string{testApp1ID, testApp2ID},
		Options: &ExportOptions{
			Format: "yaml",
		},
	}

	mockApp1 := &appmodel.Application{
		ID:          testApp1ID,
		Name:        "App One",
		Description: "First App",
	}

	mockApp2 := &appmodel.Application{
		ID:          testApp2ID,
		Name:        "App Two",
		Description: "Second App",
	}

	suite.appServiceMock.EXPECT().GetApplication(testApp1ID).Return(mockApp1, nil)
	suite.appServiceMock.EXPECT().GetApplication(testApp2ID).Return(mockApp2, nil)

	result, err := suite.exportService.ExportResources(request)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Files, 2)
	assert.Equal(suite.T(), 2, result.Summary.TotalFiles)
	assert.Equal(suite.T(), 2, result.Summary.ResourceTypes["applications"])
}

// TestExportResources_PartialFailure tests exporting when some applications fail.
func (suite *ExportServiceTestSuite) TestExportResources_PartialFailure() {
	app1ID := "valid-app"
	app2ID := "invalid-app"
	request := &ExportRequest{
		Applications: []string{app1ID, app2ID},
		Options: &ExportOptions{
			Format: "yaml",
		},
	}

	mockApp1 := &appmodel.Application{
		ID:   app1ID,
		Name: "Valid App",
	}

	appError := &serviceerror.ServiceError{
		Code:  "APP_NOT_FOUND",
		Error: "Application not found",
	}

	suite.appServiceMock.EXPECT().GetApplication(app1ID).Return(mockApp1, nil)
	suite.appServiceMock.EXPECT().GetApplication(app2ID).Return(nil, appError)

	result, err := suite.exportService.ExportResources(request)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Files, 1) // Only one successful export
	assert.Equal(suite.T(), 1, result.Summary.TotalFiles)
	assert.Len(suite.T(), result.Summary.Errors, 1) // One error recorded
	assert.Equal(suite.T(), "application", result.Summary.Errors[0].ResourceType)
	assert.Equal(suite.T(), app2ID, result.Summary.Errors[0].ResourceID)
}

// TestExportResources_WildcardApplications tests exporting all applications using wildcard.
func (suite *ExportServiceTestSuite) TestExportResources_WildcardApplications() {
	request := &ExportRequest{
		Applications: []string{"*"},
		Options: &ExportOptions{
			Format: "yaml",
		},
	}

	// Mock GetApplicationList to return 3 applications
	mockAppList := &appmodel.ApplicationListResponse{
		TotalResults: 3,
		Count:        3,
		Applications: []appmodel.BasicApplicationResponse{
			{ID: testApp1ID, Name: "Application One"},
			{ID: testApp2ID, Name: "Application Two"},
			{ID: testApp3ID, Name: "Application Three"},
		},
	}

	mockApp1 := &appmodel.Application{
		ID:          testApp1ID,
		Name:        "Application One",
		Description: "First App",
	}

	mockApp2 := &appmodel.Application{
		ID:          testApp2ID,
		Name:        "Application Two",
		Description: "Second App",
	}

	mockApp3 := &appmodel.Application{
		ID:          testApp3ID,
		Name:        "Application Three",
		Description: "Third App",
	}

	suite.appServiceMock.EXPECT().GetApplicationList().Return(mockAppList, nil)
	suite.appServiceMock.EXPECT().GetApplication(testApp1ID).Return(mockApp1, nil)
	suite.appServiceMock.EXPECT().GetApplication(testApp2ID).Return(mockApp2, nil)
	suite.appServiceMock.EXPECT().GetApplication(testApp3ID).Return(mockApp3, nil)

	result, err := suite.exportService.ExportResources(request)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Files, 3) // All 3 applications exported
	assert.Equal(suite.T(), 3, result.Summary.TotalFiles)
	assert.Equal(suite.T(), 3, result.Summary.ResourceTypes["applications"])
	assert.Len(suite.T(), result.Summary.Errors, 0) // No errors
}

// TestExportResources_WildcardApplications_ListFailure tests wildcard export when GetApplicationList fails.
func (suite *ExportServiceTestSuite) TestExportResources_WildcardApplications_ListFailure() {
	request := &ExportRequest{
		Applications: []string{"*"},
		Options: &ExportOptions{
			Format: "yaml",
		},
	}

	listError := &serviceerror.ServiceError{
		Code:  "LIST_FAILED",
		Error: "Failed to list applications",
	}

	suite.appServiceMock.EXPECT().GetApplicationList().Return(nil, listError)

	result, err := suite.exportService.ExportResources(request)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrorNoResourcesFound.Code, err.Code)
}

// TestExportResources_WildcardApplications_EmptyList tests wildcard export with empty application list.
func (suite *ExportServiceTestSuite) TestExportResources_WildcardApplications_EmptyList() {
	request := &ExportRequest{
		Applications: []string{"*"},
		Options: &ExportOptions{
			Format: "yaml",
		},
	}

	// Mock GetApplicationList to return empty list
	mockAppList := &appmodel.ApplicationListResponse{
		TotalResults: 0,
		Count:        0,
		Applications: []appmodel.BasicApplicationResponse{},
	}

	suite.appServiceMock.EXPECT().GetApplicationList().Return(mockAppList, nil)

	result, err := suite.exportService.ExportResources(request)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrorNoResourcesFound.Code, err.Code)
}

// TestExportResources_WildcardApplications_PartialFailure tests wildcard export with partial failures.
func (suite *ExportServiceTestSuite) TestExportResources_WildcardApplications_PartialFailure() {
	request := &ExportRequest{
		Applications: []string{"*"},
		Options: &ExportOptions{
			Format: "yaml",
		},
	}

	// Mock GetApplicationList to return 3 applications
	mockAppList := &appmodel.ApplicationListResponse{
		TotalResults: 3,
		Count:        3,
		Applications: []appmodel.BasicApplicationResponse{
			{ID: testApp1ID, Name: "Application One"},
			{ID: testApp2ID, Name: "Application Two"},
			{ID: testApp3ID, Name: "Application Three"},
		},
	}

	mockApp1 := &appmodel.Application{
		ID:          testApp1ID,
		Name:        "Application One",
		Description: "First App",
	}

	mockApp3 := &appmodel.Application{
		ID:          testApp3ID,
		Name:        "Application Three",
		Description: "Third App",
	}

	appError := &serviceerror.ServiceError{
		Code:  "APP_NOT_FOUND",
		Error: "Application not found",
	}

	suite.appServiceMock.EXPECT().GetApplicationList().Return(mockAppList, nil)
	suite.appServiceMock.EXPECT().GetApplication(testApp1ID).Return(mockApp1, nil)
	suite.appServiceMock.EXPECT().GetApplication(testApp2ID).Return(nil, appError)
	suite.appServiceMock.EXPECT().GetApplication(testApp3ID).Return(mockApp3, nil)

	result, err := suite.exportService.ExportResources(request)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Files, 2) // 2 successful exports
	assert.Equal(suite.T(), 2, result.Summary.TotalFiles)
	assert.Equal(suite.T(), 2, result.Summary.ResourceTypes["applications"])
	assert.Len(suite.T(), result.Summary.Errors, 1) // One error recorded
	assert.Equal(suite.T(), "application", result.Summary.Errors[0].ResourceType)
	assert.Equal(suite.T(), testApp2ID, result.Summary.Errors[0].ResourceID)
}

// TestExportResources_IdentityProvider_Success tests exporting a single IDP successfully.
func (suite *ExportServiceTestSuite) TestExportResources_IdentityProvider_Success() {
	idpID := testIDPID
	request := &ExportRequest{
		IdentityProviders: []string{idpID},
		Options: &ExportOptions{
			Format: "yaml",
		},
	}

	mockProperty, _ := cmodels.NewProperty("client_id", "test-client-id", false)
	mockIDP := &idp.IDPDTO{
		ID:          idpID,
		Name:        "Test IDP",
		Description: "Test Identity Provider",
		Type:        idp.IDPTypeGoogle,
		Properties:  []cmodels.Property{*mockProperty},
	}

	suite.idpServiceMock.EXPECT().GetIdentityProvider(idpID).Return(mockIDP, nil)

	result, err := suite.exportService.ExportResources(request)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Files, 1)
	assert.Equal(suite.T(), 1, result.Summary.TotalFiles)
	assert.Contains(suite.T(), result.Summary.ResourceTypes, "identity_providers")
	assert.Equal(suite.T(), "Test_IDP.yaml", result.Files[0].FileName)
	assert.Equal(suite.T(), "identity_provider", result.Files[0].ResourceType)
	assert.Contains(suite.T(), result.Files[0].Content, "name: Test IDP")
}

// TestExportResources_IdentityProvider_Multiple tests exporting multiple IDPs.
func (suite *ExportServiceTestSuite) TestExportResources_IdentityProvider_Multiple() {
	request := &ExportRequest{
		IdentityProviders: []string{"idp1", "idp2"},
		Options: &ExportOptions{
			Format: "yaml",
		},
	}

	mockProperty1, _ := cmodels.NewProperty("client_id", "client1", false)
	mockIDP1 := &idp.IDPDTO{
		ID:         "idp1",
		Name:       "Google IDP",
		Type:       idp.IDPTypeGoogle,
		Properties: []cmodels.Property{*mockProperty1},
	}

	mockProperty2, _ := cmodels.NewProperty("client_id", "client2", false)
	mockIDP2 := &idp.IDPDTO{
		ID:         "idp2",
		Name:       "GitHub IDP",
		Type:       idp.IDPTypeGitHub,
		Properties: []cmodels.Property{*mockProperty2},
	}

	suite.idpServiceMock.EXPECT().GetIdentityProvider("idp1").Return(mockIDP1, nil)
	suite.idpServiceMock.EXPECT().GetIdentityProvider("idp2").Return(mockIDP2, nil)

	result, err := suite.exportService.ExportResources(request)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Files, 2)
	assert.Equal(suite.T(), 2, result.Summary.TotalFiles)
	assert.Equal(suite.T(), 2, result.Summary.ResourceTypes["identity_providers"])
}

// TestExportResources_IdentityProvider_Wildcard tests exporting all IDPs using wildcard.
func (suite *ExportServiceTestSuite) TestExportResources_IdentityProvider_Wildcard() {
	request := &ExportRequest{
		IdentityProviders: []string{"*"},
		Options: &ExportOptions{
			Format: "yaml",
		},
	}

	mockIDPList := []idp.BasicIDPDTO{
		{ID: "idp1", Name: "Google IDP"},
		{ID: "idp2", Name: "GitHub IDP"},
	}

	mockProperty1, _ := cmodels.NewProperty("client_id", "client1", false)
	mockIDP1 := &idp.IDPDTO{
		ID:         "idp1",
		Name:       "Google IDP",
		Type:       idp.IDPTypeGoogle,
		Properties: []cmodels.Property{*mockProperty1},
	}

	mockProperty2, _ := cmodels.NewProperty("client_id", "client2", false)
	mockIDP2 := &idp.IDPDTO{
		ID:         "idp2",
		Name:       "GitHub IDP",
		Type:       idp.IDPTypeGitHub,
		Properties: []cmodels.Property{*mockProperty2},
	}

	suite.idpServiceMock.EXPECT().GetIdentityProviderList().Return(mockIDPList, nil)
	suite.idpServiceMock.EXPECT().GetIdentityProvider("idp1").Return(mockIDP1, nil)
	suite.idpServiceMock.EXPECT().GetIdentityProvider("idp2").Return(mockIDP2, nil)

	result, err := suite.exportService.ExportResources(request)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Files, 2)
	assert.Equal(suite.T(), 2, result.Summary.TotalFiles)
}

// TestExportResources_Mixed_ApplicationsAndIDPs tests exporting both applications and IDPs.
func (suite *ExportServiceTestSuite) TestExportResources_Mixed_ApplicationsAndIDPs() {
	request := &ExportRequest{
		Applications:      []string{testAppID},
		IdentityProviders: []string{testIDPID},
		Options: &ExportOptions{
			Format: "yaml",
		},
	}

	mockApp := &appmodel.Application{
		ID:          testAppID,
		Name:        "Test App",
		Description: "Test Description",
	}

	mockProperty, _ := cmodels.NewProperty("client_id", "test-client-id", false)
	mockIDP := &idp.IDPDTO{
		ID:         testIDPID,
		Name:       "Test IDP",
		Type:       idp.IDPTypeGoogle,
		Properties: []cmodels.Property{*mockProperty},
	}

	suite.appServiceMock.EXPECT().GetApplication(testAppID).Return(mockApp, nil)
	suite.idpServiceMock.EXPECT().GetIdentityProvider(testIDPID).Return(mockIDP, nil)

	result, err := suite.exportService.ExportResources(request)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Files, 2) // 1 app + 1 IDP
	assert.Equal(suite.T(), 2, result.Summary.TotalFiles)
	assert.Contains(suite.T(), result.Summary.ResourceTypes, "applications")
	assert.Contains(suite.T(), result.Summary.ResourceTypes, "identity_providers")
	assert.Equal(suite.T(), 1, result.Summary.ResourceTypes["applications"])
	assert.Equal(suite.T(), 1, result.Summary.ResourceTypes["identity_providers"])
}

// TestExportResources_IdentityProvider_NotFound tests error handling when IDP not found.
func (suite *ExportServiceTestSuite) TestExportResources_IdentityProvider_NotFound() {
	request := &ExportRequest{
		IdentityProviders: []string{"non-existent-idp"},
		Options: &ExportOptions{
			Format: "yaml",
		},
	}

	idpError := &serviceerror.ServiceError{
		Code:  "IDP_NOT_FOUND",
		Error: "Identity provider not found",
	}

	suite.idpServiceMock.EXPECT().GetIdentityProvider("non-existent-idp").Return(nil, idpError)

	result, err := suite.exportService.ExportResources(request)

	// Should return error since no valid resources found
	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), ErrorNoResourcesFound.Code, err.Code)
}

// TestExportResources_IdentityProvider_WildcardPartialFailure tests wildcard IDP export with partial failures.
func (suite *ExportServiceTestSuite) TestExportResources_IdentityProvider_WildcardPartialFailure() {
	request := &ExportRequest{
		IdentityProviders: []string{"*"},
		Options: &ExportOptions{
			Format: "yaml",
		},
	}

	mockIDPList := []idp.BasicIDPDTO{
		{ID: "idp1", Name: "Google IDP"},
		{ID: "idp2", Name: "GitHub IDP"},
		{ID: "idp3", Name: "OIDC IDP"},
	}

	mockProperty1, _ := cmodels.NewProperty("client_id", "client1", false)
	mockIDP1 := &idp.IDPDTO{
		ID:         "idp1",
		Name:       "Google IDP",
		Type:       idp.IDPTypeGoogle,
		Properties: []cmodels.Property{*mockProperty1},
	}

	mockProperty3, _ := cmodels.NewProperty("client_id", "client3", false)
	mockIDP3 := &idp.IDPDTO{
		ID:         "idp3",
		Name:       "OIDC IDP",
		Type:       idp.IDPTypeOIDC,
		Properties: []cmodels.Property{*mockProperty3},
	}

	idpError := &serviceerror.ServiceError{
		Code:  "IDP_NOT_FOUND",
		Error: "Identity provider not found",
	}

	suite.idpServiceMock.EXPECT().GetIdentityProviderList().Return(mockIDPList, nil)
	suite.idpServiceMock.EXPECT().GetIdentityProvider("idp1").Return(mockIDP1, nil)
	suite.idpServiceMock.EXPECT().GetIdentityProvider("idp2").Return(nil, idpError)
	suite.idpServiceMock.EXPECT().GetIdentityProvider("idp3").Return(mockIDP3, nil)

	result, err := suite.exportService.ExportResources(request)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Files, 2) // 2 successful exports
	assert.Equal(suite.T(), 2, result.Summary.TotalFiles)
	assert.Equal(suite.T(), 2, result.Summary.ResourceTypes["identity_providers"])
	assert.Len(suite.T(), result.Summary.Errors, 1) // One error recorded
	assert.Equal(suite.T(), "identity_provider", result.Summary.Errors[0].ResourceType)
	assert.Equal(suite.T(), "idp2", result.Summary.Errors[0].ResourceID)
}

// TestExportResources_IdentityProvider_NoProperties tests exporting IDP with no properties.
func (suite *ExportServiceTestSuite) TestExportResources_IdentityProvider_NoProperties() {
	request := &ExportRequest{
		IdentityProviders: []string{"idp-no-props"},
		Options: &ExportOptions{
			Format: "yaml",
		},
	}

	// IDP with no properties
	mockIDP := &idp.IDPDTO{
		ID:         "idp-no-props",
		Name:       "Empty IDP",
		Type:       idp.IDPTypeOIDC,
		Properties: []cmodels.Property{}, // Empty properties
	}

	suite.idpServiceMock.EXPECT().GetIdentityProvider("idp-no-props").Return(mockIDP, nil)

	result, err := suite.exportService.ExportResources(request)

	// Should succeed even with no properties
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Files, 1)
	assert.Equal(suite.T(), 1, result.Summary.TotalFiles)
	assert.Contains(suite.T(), result.Files[0].Content, "name: Empty IDP")
}

// TestExportResources_IdentityProvider_EmptyName tests validation for IDP with empty name.
func (suite *ExportServiceTestSuite) TestExportResources_IdentityProvider_EmptyName() {
	request := &ExportRequest{
		IdentityProviders: []string{"idp-no-name"},
		Options: &ExportOptions{
			Format: "yaml",
		},
	}

	mockProperty, _ := cmodels.NewProperty("key", "value", false)
	mockIDP := &idp.IDPDTO{
		ID:         "idp-no-name",
		Name:       "", // Empty name
		Type:       idp.IDPTypeOIDC,
		Properties: []cmodels.Property{*mockProperty},
	}

	suite.idpServiceMock.EXPECT().GetIdentityProvider("idp-no-name").Return(mockIDP, nil)

	result, err := suite.exportService.ExportResources(request)

	// Should return error since name is required
	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), ErrorNoResourcesFound.Code, err.Code)
}

// TestExportResources_IdentityProvider_PropertyParameterization verifies that IDP properties
// are correctly parameterized with context-aware variable names.
func (suite *ExportServiceTestSuite) TestExportResources_IdentityProvider_PropertyParameterization() {
	idpID := "test-parameterization-idp"
	request := &ExportRequest{
		IdentityProviders: []string{idpID},
		Options: &ExportOptions{
			Format: "yaml",
		},
	}

	// Create properties with various names
	clientIDProp, _ := cmodels.NewProperty("client_id", "test-client-123", true)
	clientSecretProp, _ := cmodels.NewProperty("client_secret", "super-secret", true)
	redirectURIProp, _ := cmodels.NewProperty("redirect_uri", "http://localhost:3000", false)

	mockIDP := &idp.IDPDTO{
		ID:          idpID,
		Name:        "Export Test IDP",
		Description: "Test IDP for parameterization",
		Type:        idp.IDPTypeGoogle,
		Properties: []cmodels.Property{
			*clientIDProp,
			*clientSecretProp,
			*redirectURIProp,
		},
	}

	suite.idpServiceMock.EXPECT().GetIdentityProvider(idpID).Return(mockIDP, nil)

	result, err := suite.exportService.ExportResources(request)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Files, 1)

	yamlContent := result.Files[0].Content

	// Verify the YAML contains parameterized property values with context-aware variable names
	// Variable names should be: IDP_NAME + PROPERTY_NAME in UPPER_SNAKE_CASE
	assert.Contains(suite.T(), yamlContent, "{{.EXPORT_TEST_IDP_CLIENT_ID}}")
	assert.Contains(suite.T(), yamlContent, "{{.EXPORT_TEST_IDP_CLIENT_SECRET}}")
	assert.Contains(suite.T(), yamlContent, "{{.EXPORT_TEST_IDP_REDIRECT_URI}}")

	// Verify property names are preserved
	assert.Contains(suite.T(), yamlContent, "name: client_id")
	assert.Contains(suite.T(), yamlContent, "name: client_secret")
	assert.Contains(suite.T(), yamlContent, "name: redirect_uri")

	// Verify secret flags are preserved (YAML uses 'is_secret' field name)
	assert.Contains(suite.T(), yamlContent, "is_secret: true")

	// Verify basic IDP fields
	assert.Contains(suite.T(), yamlContent, "name: Export Test IDP")
	assert.Contains(suite.T(), yamlContent, "type: GOOGLE")
}

// TestExportResources_IdentityProvider_PropertyStructure verifies that IDP properties
// are exported with correct YAML structure including name, value, and is_secret fields.
func (suite *ExportServiceTestSuite) TestExportResources_IdentityProvider_PropertyStructure() {
	idpID := "test-property-structure"
	request := &ExportRequest{
		IdentityProviders: []string{idpID},
		Options: &ExportOptions{
			Format: "yaml",
		},
	}

	// Create properties with various combinations - some secret, some not
	clientIDProp, _ := cmodels.NewProperty("client_id", "test-client-123", false)
	clientSecretProp, _ := cmodels.NewProperty("client_secret", "super-secret-value", true)
	apiKeyProp, _ := cmodels.NewProperty("api_key", "api-key-xyz", true)
	callbackURLProp, _ := cmodels.NewProperty("callback_url", "https://example.com/callback", false)

	mockIDP := &idp.IDPDTO{
		ID:          idpID,
		Name:        "Property Structure Test",
		Description: "Test IDP for property YAML structure validation",
		Type:        idp.IDPTypeOIDC,
		Properties: []cmodels.Property{
			*clientIDProp,
			*clientSecretProp,
			*apiKeyProp,
			*callbackURLProp,
		},
	}

	suite.idpServiceMock.EXPECT().GetIdentityProvider(idpID).Return(mockIDP, nil)

	result, err := suite.exportService.ExportResources(request)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Files, 1)

	yamlContent := result.Files[0].Content

	// Verify property names are preserved in the YAML
	assert.Contains(suite.T(), yamlContent, "name: client_id")
	assert.Contains(suite.T(), yamlContent, "name: client_secret")
	assert.Contains(suite.T(), yamlContent, "name: api_key")
	assert.Contains(suite.T(), yamlContent, "name: callback_url")

	// Verify all properties have value fields (template variables due to DynamicPropertyFields)
	assert.Contains(suite.T(), yamlContent, "value:")

	// Verify secret flags are preserved for secret properties
	// Count occurrences of "is_secret: true" - should be 2 (client_secret and api_key)
	secretCount := strings.Count(yamlContent, "is_secret: true")
	assert.Equal(suite.T(), 2, secretCount, "Should have exactly 2 secret properties")

	// Verify the properties section exists and has proper structure
	assert.Contains(suite.T(), yamlContent, "properties:")

	// Verify basic IDP fields
	assert.Contains(suite.T(), yamlContent, "name: Property Structure Test")
	assert.Contains(suite.T(), yamlContent, "description: Test IDP for property YAML structure validation")
	assert.Contains(suite.T(), yamlContent, "type: OIDC")

	// Verify proper indentation and YAML list structure for properties
	assert.Contains(suite.T(), yamlContent, "properties:\n  - name:")
}

// TestExportResources_PartialFailure_DetailedErrorValidation enhances the existing partial failure test
// with detailed error field validation.
func (suite *ExportServiceTestSuite) TestExportResources_PartialFailure_DetailedErrorValidation() {
	app1ID := "app1"
	app2ID := "app2-not-found"

	request := &ExportRequest{
		Applications: []string{app1ID, app2ID},
		Options: &ExportOptions{
			Format: "yaml",
		},
	}

	mockApp1 := &appmodel.Application{
		ID:   app1ID,
		Name: "Valid App",
	}

	appError := &serviceerror.ServiceError{
		Code:  "APP_NOT_FOUND",
		Error: "Application not found",
	}

	suite.appServiceMock.EXPECT().GetApplication(app1ID).Return(mockApp1, nil)
	suite.appServiceMock.EXPECT().GetApplication(app2ID).Return(nil, appError)

	result, err := suite.exportService.ExportResources(request)

	// Verify successful export
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Files, 1)
	assert.Equal(suite.T(), 1, result.Summary.TotalFiles)

	// Verify error details
	assert.Len(suite.T(), result.Summary.Errors, 1)
	exportError := result.Summary.Errors[0]
	assert.Equal(suite.T(), "application", exportError.ResourceType)
	assert.Equal(suite.T(), app2ID, exportError.ResourceID)
	assert.Equal(suite.T(), "APP_NOT_FOUND", exportError.Code)
	assert.Equal(suite.T(), "Application not found", exportError.Error)

	// Verify file size calculation
	assert.Equal(suite.T(), int64(len(result.Files[0].Content)), result.Files[0].Size)
	assert.Greater(suite.T(), result.Summary.TotalSize, int64(0))
}

// TestExportResources_IdentityProvider_PartialFailure_DetailedErrorValidation tests IDP partial failure.
func (suite *ExportServiceTestSuite) TestExportResources_IdentityProvider_PartialFailure_DetailedErrorValidation() {
	request := &ExportRequest{
		IdentityProviders: []string{"idp1", "idp2-not-found", "idp3"},
		Options: &ExportOptions{
			Format: "yaml",
		},
	}

	mockProperty1, _ := cmodels.NewProperty("client_id", "client1", false)
	mockIDP1 := &idp.IDPDTO{
		ID:         "idp1",
		Name:       "Google IDP",
		Type:       idp.IDPTypeGoogle,
		Properties: []cmodels.Property{*mockProperty1},
	}

	mockProperty3, _ := cmodels.NewProperty("client_id", "client3", false)
	mockIDP3 := &idp.IDPDTO{
		ID:         "idp3",
		Name:       "GitHub IDP",
		Type:       idp.IDPTypeGitHub,
		Properties: []cmodels.Property{*mockProperty3},
	}

	idpError := &serviceerror.ServiceError{
		Code:  "IDP_NOT_FOUND",
		Error: "Identity provider not found",
	}

	suite.idpServiceMock.EXPECT().GetIdentityProvider("idp1").Return(mockIDP1, nil)
	suite.idpServiceMock.EXPECT().GetIdentityProvider("idp2-not-found").Return(nil, idpError)
	suite.idpServiceMock.EXPECT().GetIdentityProvider("idp3").Return(mockIDP3, nil)

	result, err := suite.exportService.ExportResources(request)

	// Verify partial success
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Files, 2) // Two successful exports
	assert.Equal(suite.T(), 2, result.Summary.TotalFiles)

	// Verify error details
	assert.Len(suite.T(), result.Summary.Errors, 1)
	exportError := result.Summary.Errors[0]
	assert.Equal(suite.T(), "identity_provider", exportError.ResourceType)
	assert.Equal(suite.T(), "idp2-not-found", exportError.ResourceID)
	assert.Equal(suite.T(), "IDP_NOT_FOUND", exportError.Code)
	assert.Equal(suite.T(), "Identity provider not found", exportError.Error)

	// Verify file sizes
	for _, file := range result.Files {
		assert.Equal(suite.T(), int64(len(file.Content)), file.Size)
	}
	assert.Greater(suite.T(), result.Summary.TotalSize, int64(0))
}

// TestExportResources_MixedResources_WithErrors tests exporting both apps and IDPs with some failures.
func (suite *ExportServiceTestSuite) TestExportResources_MixedResources_WithErrors() {
	request := &ExportRequest{
		Applications:      []string{"app1", "app2-not-found"},
		IdentityProviders: []string{"idp1", "idp2-not-found"},
		Options: &ExportOptions{
			Format: "yaml",
		},
	}

	// Setup successful app
	mockApp1 := &appmodel.Application{
		ID:   "app1",
		Name: "Valid App",
	}

	// Setup app error
	appError := &serviceerror.ServiceError{
		Code:  "APP_NOT_FOUND",
		Error: "Application not found",
	}

	// Setup successful IDP
	mockProperty1, _ := cmodels.NewProperty("client_id", "client1", false)
	mockIDP1 := &idp.IDPDTO{
		ID:         "idp1",
		Name:       "Google IDP",
		Type:       idp.IDPTypeGoogle,
		Properties: []cmodels.Property{*mockProperty1},
	}

	// Setup IDP error
	idpError := &serviceerror.ServiceError{
		Code:  "IDP_NOT_FOUND",
		Error: "Identity provider not found",
	}

	suite.appServiceMock.EXPECT().GetApplication("app1").Return(mockApp1, nil)
	suite.appServiceMock.EXPECT().GetApplication("app2-not-found").Return(nil, appError)
	suite.idpServiceMock.EXPECT().GetIdentityProvider("idp1").Return(mockIDP1, nil)
	suite.idpServiceMock.EXPECT().GetIdentityProvider("idp2-not-found").Return(nil, idpError)

	result, err := suite.exportService.ExportResources(request)

	// Verify partial success
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Files, 2) // One app + one IDP
	assert.Equal(suite.T(), 2, result.Summary.TotalFiles)

	// Verify resource type counts
	assert.Equal(suite.T(), 1, result.Summary.ResourceTypes["applications"])
	assert.Equal(suite.T(), 1, result.Summary.ResourceTypes["identity_providers"])

	// Verify errors - should have 2 errors (1 app, 1 IDP)
	assert.Len(suite.T(), result.Summary.Errors, 2)

	// Verify app error
	var appErrorFound bool
	var idpErrorFound bool
	for _, e := range result.Summary.Errors {
		if e.ResourceType == "application" {
			appErrorFound = true
			assert.Equal(suite.T(), "app2-not-found", e.ResourceID)
			assert.Equal(suite.T(), "APP_NOT_FOUND", e.Code)
		}
		if e.ResourceType == "identity_provider" {
			idpErrorFound = true
			assert.Equal(suite.T(), "idp2-not-found", e.ResourceID)
			assert.Equal(suite.T(), "IDP_NOT_FOUND", e.Code)
		}
	}
	assert.True(suite.T(), appErrorFound, "Application error not found in Summary.Errors")
	assert.True(suite.T(), idpErrorFound, "IDP error not found in Summary.Errors")
}

// TestExportResources_FileSizeCalculation tests that file sizes are calculated correctly.
func (suite *ExportServiceTestSuite) TestExportResources_FileSizeCalculation() {
	request := &ExportRequest{
		Applications: []string{testApp1ID, testApp2ID},
		Options: &ExportOptions{
			Format: "yaml",
		},
	}

	mockApp1 := &appmodel.Application{
		ID:          testApp1ID,
		Name:        "Application One",
		Description: "First application",
	}

	mockApp2 := &appmodel.Application{
		ID:          testApp2ID,
		Name:        "Application Two",
		Description: "Second application with longer description",
	}

	suite.appServiceMock.EXPECT().GetApplication(testApp1ID).Return(mockApp1, nil)
	suite.appServiceMock.EXPECT().GetApplication(testApp2ID).Return(mockApp2, nil)

	result, err := suite.exportService.ExportResources(request)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Files, 2)

	// Verify each file's size matches its content length
	var totalCalculatedSize int64
	for _, file := range result.Files {
		expectedSize := int64(len(file.Content))
		assert.Equal(suite.T(), expectedSize, file.Size, "File size mismatch for %s", file.FileName)
		assert.Greater(suite.T(), file.Size, int64(0), "File size should be greater than 0")
		totalCalculatedSize += file.Size
	}

	// Verify total size matches sum of individual file sizes
	assert.Equal(suite.T(), totalCalculatedSize, result.Summary.TotalSize)
	assert.Greater(suite.T(), result.Summary.TotalSize, int64(0))
}

// MockParameterizer is a mock implementation of ParameterizerInterface for testing.
type MockParameterizer struct {
	shouldFail bool
	errorMsg   string
}

func (m *MockParameterizer) ToParameterizedYAML(
	obj interface{}, resourceType string, resourceName string) (string, error) {
	if m.shouldFail {
		return "", fmt.Errorf("%s", m.errorMsg)
	}
	// Return minimal valid YAML
	return "id: test\nname: test\n", nil
}

// TestExportResources_TemplateGenerationError tests the error path in generateTemplateFromStruct.
func (suite *ExportServiceTestSuite) TestExportResources_TemplateGenerationError() {
	request := &ExportRequest{
		Applications: []string{testApp1ID, testApp2ID},
		Options: &ExportOptions{
			Format: "yaml",
		},
	}

	mockApp1 := &appmodel.Application{
		ID:   testApp1ID,
		Name: "Valid App",
	}

	mockApp2 := &appmodel.Application{
		ID:   testApp2ID,
		Name: "App That Fails Template Generation",
	}

	suite.appServiceMock.EXPECT().GetApplication(testApp1ID).Return(mockApp1, nil)
	suite.appServiceMock.EXPECT().GetApplication(testApp2ID).Return(mockApp2, nil)

	// Create a mock parameterizer that returns errors
	mockParameterizer := &MockParameterizer{
		shouldFail: true,
		errorMsg:   "template generation failed: unknown resource type",
	}

	// Create a new export service with the mock parameterizer
	exportServiceWithMock := newExportService(suite.appServiceMock, suite.idpServiceMock, mockParameterizer)

	result, err := exportServiceWithMock.ExportResources(request)

	// When all resources fail template generation, service returns error
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrorNoResourcesFound.Code, err.Code)

	// Result should be nil when error is returned
	assert.Nil(suite.T(), result)
}

// TestExportResources_WithCustomFolderStructure tests the CustomStructure path in generateFolderPath.
func (suite *ExportServiceTestSuite) TestExportResources_WithCustomFolderStructure() {
	request := &ExportRequest{
		Applications: []string{testApp1ID},
		Options: &ExportOptions{
			Format: "yaml",
			FolderStructure: &FolderStructureOptions{
				CustomStructure: map[string]string{
					"application": "custom/apps/folder",
				},
			},
		},
	}

	mockApp := &appmodel.Application{
		ID:   testApp1ID,
		Name: "Test Application",
	}

	suite.appServiceMock.EXPECT().GetApplication(testApp1ID).Return(mockApp, nil)

	result, err := suite.exportService.ExportResources(request)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Files, 1)

	// Verify the file has custom folder path
	assert.Equal(suite.T(), "custom/apps/folder", result.Files[0].FolderPath)
}

// TestExportResources_WithGroupByTypeStructure tests the GroupByType path in generateFolderPath.
func (suite *ExportServiceTestSuite) TestExportResources_WithGroupByTypeStructure() {
	request := &ExportRequest{
		Applications:      []string{testApp1ID, testApp2ID},
		IdentityProviders: []string{"idp1"},
		Options: &ExportOptions{
			Format: "yaml",
			FolderStructure: &FolderStructureOptions{
				GroupByType: true,
			},
		},
	}

	mockApp1 := &appmodel.Application{
		ID:   testApp1ID,
		Name: "Application One",
	}

	mockApp2 := &appmodel.Application{
		ID:   testApp2ID,
		Name: "Application Two",
	}

	mockProperty, _ := cmodels.NewProperty("client_id", "test-client", false)
	mockIDP := &idp.IDPDTO{
		ID:         "idp1",
		Name:       "Test IDP",
		Type:       idp.IDPTypeGoogle,
		Properties: []cmodels.Property{*mockProperty},
	}

	suite.appServiceMock.EXPECT().GetApplication(testApp1ID).Return(mockApp1, nil)
	suite.appServiceMock.EXPECT().GetApplication(testApp2ID).Return(mockApp2, nil)
	suite.idpServiceMock.EXPECT().GetIdentityProvider("idp1").Return(mockIDP, nil)

	result, err := suite.exportService.ExportResources(request)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Files, 3) // 2 apps + 1 IDP

	// Verify applications are in "applications" folder
	appFiles := 0
	idpFiles := 0
	for _, file := range result.Files {
		if file.ResourceType == "application" {
			assert.Equal(suite.T(), "applications", file.FolderPath)
			appFiles++
		} else if file.ResourceType == "identity_provider" {
			assert.Equal(suite.T(), "identity_providers", file.FolderPath)
			idpFiles++
		}
	}

	assert.Equal(suite.T(), 2, appFiles, "Should have 2 application files")
	assert.Equal(suite.T(), 1, idpFiles, "Should have 1 IDP file")
}
