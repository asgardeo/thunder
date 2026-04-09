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

package dcr

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/cert"
	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/tests/mocks/applicationmock"
	i18nmock "github.com/asgardeo/thunder/tests/mocks/i18n/mgtmock"
	"github.com/asgardeo/thunder/tests/mocks/oumock"
)

// DCRServiceTestSuite is the test suite for DCR service
type DCRServiceTestSuite struct {
	suite.Suite
	mockAppService *applicationmock.ApplicationServiceInterfaceMock
	mockOUService  *oumock.OrganizationUnitServiceInterfaceMock
	service        DCRServiceInterface
}

func TestDCRServiceTestSuite(t *testing.T) {
	suite.Run(t, new(DCRServiceTestSuite))
}

// MockTransactioner is a simple implementation of Transactioner for testing.
type MockTransactioner struct{}

func (m *MockTransactioner) Transact(ctx context.Context, txFunc func(context.Context) error) error {
	return txFunc(ctx)
}

func (s *DCRServiceTestSuite) SetupTest() {
	s.mockAppService = applicationmock.NewApplicationServiceInterfaceMock(s.T())
	s.mockOUService = oumock.NewOrganizationUnitServiceInterfaceMock(s.T())
	s.service = newDCRService(s.mockAppService, s.mockOUService, nil, &MockTransactioner{})
}

// TestNewDCRService tests the service constructor
func (s *DCRServiceTestSuite) TestNewDCRService() {
	service := newDCRService(s.mockAppService, s.mockOUService, nil, &MockTransactioner{})
	s.NotNil(service)
	s.Implements((*DCRServiceInterface)(nil), service)
}

// TestRegisterClient_NilRequest tests nil request handling
func (s *DCRServiceTestSuite) TestRegisterClient_NilRequest() {
	response, err := s.service.RegisterClient(context.Background(), nil)

	s.Nil(response)
	s.NotNil(err)
	s.Equal(ErrorInvalidRequestFormat.Code, err.Code)
}

// TestRegisterClient_JWKSConflict tests JWKS and JWKS_URI conflict
func (s *DCRServiceTestSuite) TestRegisterClient_JWKSConflict() {
	request := &DCRRegistrationRequest{
		RedirectURIs: []string{"https://client.example.com/callback"},
		GrantTypes:   []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
		JWKSUri:      "https://client.example.com/.well-known/jwks.json",
		JWKS:         map[string]interface{}{"keys": []interface{}{}},
	}

	response, err := s.service.RegisterClient(context.Background(), request)

	s.Nil(response)
	s.NotNil(err)
	s.Equal(ErrorJWKSConfigurationConflict.Code, err.Code)
}

// TestRegisterClient_ClientNameProvided tests registration with provided client name
func (s *DCRServiceTestSuite) TestRegisterClient_ClientNameProvided() {
	request := &DCRRegistrationRequest{
		OUID:         "test-ou-1",
		RedirectURIs: []string{"https://client.example.com/callback"},
		GrantTypes:   []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
		ClientName:   "Test Client",
	}

	appDTO := &model.ApplicationDTO{
		ID:   "app-id",
		Name: "Test Client",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					ClientID:     "client-id",
					ClientSecret: "client-secret",
					Scopes:       []string{},
				},
			},
		},
	}

	s.mockAppService.On(
		"CreateApplication", mock.Anything, mock.AnythingOfType("*model.ApplicationDTO"),
	).Return(appDTO, (*serviceerror.ServiceError)(nil))

	response, err := s.service.RegisterClient(context.Background(), request)

	s.NotNil(response)
	s.Nil(err)
	s.Equal("client-id", response.ClientID)
	s.Equal("Test Client", response.ClientName)
}

// TestRegisterClient_JWKSUriProvided tests registration with JWKS_URI
func (s *DCRServiceTestSuite) TestRegisterClient_JWKSUriProvided() {
	request := &DCRRegistrationRequest{
		OUID:         "test-ou-1",
		RedirectURIs: []string{"https://client.example.com/callback"},
		GrantTypes:   []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
		ClientName:   "Test Client",
		JWKSUri:      "https://client.example.com/.well-known/jwks.json",
	}

	appDTO := &model.ApplicationDTO{
		ID:   "app-id",
		Name: "Test Client",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					ClientID:     "client-id",
					ClientSecret: "client-secret",
					Scopes:       []string{},
				},
			},
		},
		Certificate: &model.ApplicationCertificate{
			Type:  cert.CertificateTypeJWKSURI,
			Value: "https://client.example.com/.well-known/jwks.json",
		},
	}

	s.mockAppService.On(
		"CreateApplication", mock.Anything, mock.AnythingOfType("*model.ApplicationDTO"),
	).Return(appDTO, (*serviceerror.ServiceError)(nil))

	response, err := s.service.RegisterClient(context.Background(), request)

	s.NotNil(response)
	s.Nil(err)
	s.Equal("https://client.example.com/.well-known/jwks.json", response.JWKSUri)
}

// TestRegisterClient_ApplicationServiceError tests application service error handling
func (s *DCRServiceTestSuite) TestRegisterClient_ApplicationServiceError() {
	request := &DCRRegistrationRequest{
		OUID:         "test-ou-1",
		RedirectURIs: []string{"not-a-valid-uri"},
		GrantTypes:   []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
	}

	appServiceErr := &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "APP-1014",
		Error:            "Invalid URI",
		ErrorDescription: "The redirect URI is invalid",
	}

	s.mockAppService.On("CreateApplication", mock.Anything, mock.AnythingOfType("*model.ApplicationDTO")).
		Return(nil, appServiceErr)

	response, err := s.service.RegisterClient(context.Background(), request)

	s.Nil(response)
	s.NotNil(err)
	s.Equal(ErrorInvalidRedirectURI.Code, err.Code)
}

// TestMapApplicationErrorToDCRError tests error mapping
func (s *DCRServiceTestSuite) TestMapApplicationErrorToDCRError() {
	testCases := []struct {
		name            string
		appErrCode      string
		expectedDCRCode string
	}{
		{
			name:            "Invalid Logo URL Error APP-1006",
			appErrCode:      "APP-1006",
			expectedDCRCode: ErrorInvalidClientMetadata.Code,
		},
		{
			name:            "Redirect URI Error APP-1014",
			appErrCode:      "APP-1014",
			expectedDCRCode: ErrorInvalidRedirectURI.Code,
		},
		{
			name:            "Redirect URI Error APP-1015",
			appErrCode:      "APP-1015",
			expectedDCRCode: ErrorInvalidRedirectURI.Code,
		},
		{
			name:            "Server Error APP-5001",
			appErrCode:      "APP-5001",
			expectedDCRCode: ErrorServerError.Code,
		},
		{
			name:            "Server Error APP-5002",
			appErrCode:      "APP-5002",
			expectedDCRCode: ErrorServerError.Code,
		},
		{
			name:            "Default Client Error",
			appErrCode:      "APP-9999",
			expectedDCRCode: ErrorInvalidClientMetadata.Code,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			appErr := &serviceerror.ServiceError{
				Code: tc.appErrCode,
			}

			service := s.service.(*dcrService)
			dcrErr := service.mapApplicationErrorToDCRError(appErr)

			s.Equal(tc.expectedDCRCode, dcrErr.Code)
		})
	}
}

func (s *DCRServiceTestSuite) TestRegisterClient_ConvertDCRToApplicationError() {
	request := &DCRRegistrationRequest{
		OUID:         "test-ou-1",
		RedirectURIs: []string{"https://client.example.com/callback"},
		GrantTypes:   []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
		JWKS:         map[string]interface{}{"keys": make(chan int)},
	}

	s.mockAppService.On(
		"CreateApplication", mock.Anything, mock.AnythingOfType("*model.ApplicationDTO"),
	).Return(nil, &serviceerror.ServiceError{
		Type: serviceerror.ServerErrorType,
		Code: "APP-5001",
	})

	response, err := s.service.RegisterClient(context.Background(), request)

	s.Nil(response)
	s.NotNil(err)
	s.Equal(ErrorServerError.Code, err.Code)
}

func (s *DCRServiceTestSuite) TestRegisterClient_ConvertApplicationToDCRResponseError() {
	request := &DCRRegistrationRequest{
		OUID:         "test-ou-1",
		RedirectURIs: []string{"https://client.example.com/callback"},
		GrantTypes:   []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
		ClientName:   "Test Client",
	}

	appDTO := &model.ApplicationDTO{
		ID:   "app-id",
		Name: "Test Client",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					ClientID:     "client-id",
					ClientSecret: "client-secret",
					Scopes:       []string{},
				},
			},
		},
		Certificate: &model.ApplicationCertificate{
			Type:  cert.CertificateTypeJWKS,
			Value: "invalid json",
		},
	}

	s.mockAppService.On(
		"CreateApplication", mock.Anything, mock.AnythingOfType("*model.ApplicationDTO"),
	).Return(appDTO, (*serviceerror.ServiceError)(nil))

	response, err := s.service.RegisterClient(context.Background(), request)

	s.Nil(response)
	s.NotNil(err)
	s.Equal(ErrorServerError.Code, err.Code)
}

func (s *DCRServiceTestSuite) TestRegisterClient_WithJWKS() {
	request := &DCRRegistrationRequest{
		OUID:         "test-ou-1",
		RedirectURIs: []string{"https://client.example.com/callback"},
		GrantTypes:   []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
		ClientName:   "Test Client",
		JWKS:         map[string]interface{}{"keys": []interface{}{}},
	}

	appDTO := &model.ApplicationDTO{
		ID:   "app-id",
		Name: "Test Client",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					ClientID:     "client-id",
					ClientSecret: "client-secret",
					Scopes:       []string{},
				},
			},
		},
		Certificate: &model.ApplicationCertificate{
			Type:  cert.CertificateTypeJWKS,
			Value: `{"keys":[]}`,
		},
	}

	s.mockAppService.On(
		"CreateApplication", mock.Anything, mock.AnythingOfType("*model.ApplicationDTO"),
	).Return(appDTO, (*serviceerror.ServiceError)(nil))

	response, err := s.service.RegisterClient(context.Background(), request)

	s.NotNil(response)
	s.Nil(err)
	s.NotNil(response.JWKS)
}

func (s *DCRServiceTestSuite) TestRegisterClient_WithScope() {
	request := &DCRRegistrationRequest{
		OUID:         "test-ou-1",
		RedirectURIs: []string{"https://client.example.com/callback"},
		GrantTypes:   []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
		ClientName:   "Test Client",
		Scope:        "read write admin",
	}

	appDTO := &model.ApplicationDTO{
		ID:   "app-id",
		Name: "Test Client",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					ClientID:     "client-id",
					ClientSecret: "client-secret",
					Scopes:       []string{"read", "write", "admin"},
				},
			},
		},
	}

	s.mockAppService.On(
		"CreateApplication", mock.Anything, mock.AnythingOfType("*model.ApplicationDTO"),
	).Return(appDTO, (*serviceerror.ServiceError)(nil))

	response, err := s.service.RegisterClient(context.Background(), request)

	s.NotNil(response)
	s.Nil(err)
	s.Equal("read write admin", response.Scope)
}

func (s *DCRServiceTestSuite) TestRegisterClient_EmptyInboundAuthConfig() {
	request := &DCRRegistrationRequest{
		OUID:         "test-ou-1",
		RedirectURIs: []string{"https://client.example.com/callback"},
		GrantTypes:   []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
		ClientName:   "Test Client",
	}

	appDTO := &model.ApplicationDTO{
		ID:                "app-id",
		Name:              "Test Client",
		InboundAuthConfig: []model.InboundAuthConfigDTO{},
	}

	s.mockAppService.On(
		"CreateApplication", mock.Anything, mock.AnythingOfType("*model.ApplicationDTO"),
	).Return(appDTO, (*serviceerror.ServiceError)(nil))

	response, err := s.service.RegisterClient(context.Background(), request)

	s.NotNil(response)
	s.Nil(err)
	s.NotNil(response)
}

// TestRegisterClient_WithLocalizedVariants tests that localized fields are persisted and echoed.
func (s *DCRServiceTestSuite) TestRegisterClient_WithLocalizedVariants() {
	mockI18n := i18nmock.NewI18nServiceInterfaceMock(s.T())
	svc := newDCRService(s.mockAppService, s.mockOUService, mockI18n, &MockTransactioner{})

	request := &DCRRegistrationRequest{
		OUID:                "test-ou-1",
		ClientName:          "Test Client",
		LocalizedClientName: map[string]string{"fr": "Client FR", "de": "Client DE"},
		LocalizedLogoURI:    map[string]string{"fr": "https://example.fr/logo.png"},
	}

	appDTO := &model.ApplicationDTO{
		ID:   "app-id",
		Name: "Test Client",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					ClientID:     "client-id",
					ClientSecret: "client-secret",
					Scopes:       []string{},
				},
			},
		},
	}

	s.mockAppService.On(
		"CreateApplication", mock.Anything, mock.AnythingOfType("*model.ApplicationDTO"),
	).Return(appDTO, (*serviceerror.ServiceError)(nil))

	// Expect one SetTranslationOverrideForKey call per localized entry.
	mockI18n.On("SetTranslationOverrideForKey", mock.Anything, "app.app-id", mock.Anything, mock.Anything).
		Return(nil, (*serviceerror.I18nServiceError)(nil))

	response, err := svc.RegisterClient(context.Background(), request)

	s.NotNil(response)
	s.Nil(err)
	s.Equal(map[string]string{"fr": "Client FR", "de": "Client DE"}, response.LocalizedClientName)
	s.Equal(map[string]string{"fr": "https://example.fr/logo.png"}, response.LocalizedLogoURI)
}

// TestRegisterClient_LocalizedVariantsWriteFailure tests that a failed i18n write triggers
// partial-row cleanup and app compensation delete.
func (s *DCRServiceTestSuite) TestRegisterClient_LocalizedVariantsWriteFailure() {
	mockI18n := i18nmock.NewI18nServiceInterfaceMock(s.T())
	svc := newDCRService(s.mockAppService, s.mockOUService, mockI18n, &MockTransactioner{})

	request := &DCRRegistrationRequest{
		OUID:                "test-ou-1",
		ClientName:          "Test Client",
		LocalizedClientName: map[string]string{"fr": "Client FR"},
	}

	appDTO := &model.ApplicationDTO{
		ID:   "app-id",
		Name: "Test Client",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					ClientID: "client-id",
					Scopes:   []string{},
				},
			},
		},
	}

	i18nErr := &serviceerror.I18nServiceError{Code: "I18N-500"}

	s.mockAppService.On(
		"CreateApplication", mock.Anything, mock.AnythingOfType("*model.ApplicationDTO"),
	).Return(appDTO, (*serviceerror.ServiceError)(nil))
	mockI18n.On("SetTranslationOverrideForKey", mock.Anything, "app.app-id", mock.Anything, mock.Anything).
		Return(nil, i18nErr)
	mockI18n.On("DeleteTranslationsByNamespace", "app.app-id").
		Return((*serviceerror.I18nServiceError)(nil))
	s.mockAppService.On("DeleteApplication", mock.Anything, "app-id").
		Return((*serviceerror.ServiceError)(nil))

	response, err := svc.RegisterClient(context.Background(), request)

	s.Nil(response)
	s.NotNil(err)
	s.Equal(ErrorServerError.Code, err.Code)
	mockI18n.AssertCalled(s.T(), "DeleteTranslationsByNamespace", "app.app-id")
	s.mockAppService.AssertCalled(s.T(), "DeleteApplication", mock.Anything, "app-id")
}

// TestGetClient_AppNotFound tests that a missing application returns ErrorClientNotFound.
func (s *DCRServiceTestSuite) TestGetClient_AppNotFound() {
	s.mockAppService.On("GetApplication", mock.Anything, "app-id").
		Return((*model.Application)(nil), &serviceerror.ServiceError{Code: "APP-1001"})

	response, err := s.service.GetClient(context.Background(), "app-id")

	s.Nil(response)
	s.NotNil(err)
	s.Equal(ErrorClientNotFound.Code, err.Code)
}

// TestGetClient_NoOAuthConfig tests that an app without OAuth config returns ErrorClientNotFound.
func (s *DCRServiceTestSuite) TestGetClient_NoOAuthConfig() {
	app := &model.Application{
		ID:                "app-id",
		Name:              "Test Client",
		InboundAuthConfig: []model.InboundAuthConfigComplete{},
	}

	s.mockAppService.On("GetApplication", mock.Anything, "app-id").
		Return(app, (*serviceerror.ServiceError)(nil))

	response, err := s.service.GetClient(context.Background(), "app-id")

	s.Nil(response)
	s.NotNil(err)
	s.Equal(ErrorClientNotFound.Code, err.Code)
}

// TestGetClient_Success tests happy-path retrieval without localized variants.
func (s *DCRServiceTestSuite) TestGetClient_Success() {
	app := &model.Application{
		ID:   "app-id",
		Name: "Test Client",
		InboundAuthConfig: []model.InboundAuthConfigComplete{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigComplete{
					ClientID:     "client-id",
					ClientSecret: "client-secret",
					Scopes:       []string{"read", "write"},
					RedirectURIs: []string{"https://example.com/cb"},
				},
			},
		},
		URL:    "https://example.com",
		TosURI: "https://example.com/tos",
	}

	s.mockAppService.On("GetApplication", mock.Anything, "app-id").
		Return(app, (*serviceerror.ServiceError)(nil))

	response, err := s.service.GetClient(context.Background(), "app-id")

	s.NotNil(response)
	s.Nil(err)
	s.Equal("client-id", response.ClientID)
	s.Equal("client-secret", response.ClientSecret)
	s.Equal("Test Client", response.ClientName)
	s.Equal("read write", response.Scope)
	s.Equal("https://example.com/tos", response.TosURI)
	s.Equal("app-id", response.AppID)
	s.Nil(response.LocalizedClientName)
}

// TestGetClient_WithLocalizedVariants tests that localized fields are loaded from the i18n store.
func (s *DCRServiceTestSuite) TestGetClient_WithLocalizedVariants() {
	mockI18n := i18nmock.NewI18nServiceInterfaceMock(s.T())
	svc := newDCRService(s.mockAppService, s.mockOUService, mockI18n, &MockTransactioner{})

	app := &model.Application{
		ID:   "app-id",
		Name: "Test Client",
		InboundAuthConfig: []model.InboundAuthConfigComplete{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigComplete{
					ClientID: "client-id",
					Scopes:   []string{},
				},
			},
		},
	}

	localized := map[string]map[string]string{
		"name":    {"fr": "Client FR", "de": "Client DE"},
		"tos_uri": {"fr": "https://example.fr/tos"},
	}

	s.mockAppService.On("GetApplication", mock.Anything, "app-id").
		Return(app, (*serviceerror.ServiceError)(nil))
	mockI18n.On("GetTranslationsByNamespace", "app.app-id").
		Return(localized, (*serviceerror.I18nServiceError)(nil))

	response, err := svc.GetClient(context.Background(), "app-id")

	s.NotNil(response)
	s.Nil(err)
	s.Equal(map[string]string{"fr": "Client FR", "de": "Client DE"}, response.LocalizedClientName)
	s.Equal(map[string]string{"fr": "https://example.fr/tos"}, response.LocalizedTosURI)
	s.Nil(response.LocalizedLogoURI)
	s.Nil(response.LocalizedPolicyURI)
}
