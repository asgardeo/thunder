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
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/cert"
	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/tests/mocks/applicationmock"
)

// DCRServiceTestSuite is the test suite for DCR service
type DCRServiceTestSuite struct {
	suite.Suite
	mockAppService *applicationmock.ApplicationServiceInterfaceMock
	service        DCRServiceInterface
}

func TestDCRServiceTestSuite(t *testing.T) {
	suite.Run(t, new(DCRServiceTestSuite))
}

func (s *DCRServiceTestSuite) SetupTest() {
	s.mockAppService = applicationmock.NewApplicationServiceInterfaceMock(s.T())
	s.service = newDCRService(s.mockAppService)
}

// TestNewDCRService tests the service constructor
func (s *DCRServiceTestSuite) TestNewDCRService() {
	service := newDCRService(s.mockAppService)
	s.NotNil(service)
	s.Implements((*DCRServiceInterface)(nil), service)
}

// TestRegisterClient_NilRequest tests nil request handling
func (s *DCRServiceTestSuite) TestRegisterClient_NilRequest() {
	response, err := s.service.RegisterClient(nil)

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

	response, err := s.service.RegisterClient(request)

	s.Nil(response)
	s.NotNil(err)
	s.Equal(ErrorJWKSConfigurationConflict.Code, err.Code)
}

// TestRegisterClient_ClientNameProvided tests registration with provided client name
func (s *DCRServiceTestSuite) TestRegisterClient_ClientNameProvided() {
	request := &DCRRegistrationRequest{
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
		"CreateApplication", mock.AnythingOfType("*model.ApplicationDTO"),
	).Return(appDTO, (*serviceerror.ServiceError)(nil))

	response, err := s.service.RegisterClient(request)

	s.NotNil(response)
	s.Nil(err)
	s.Equal("client-id", response.ClientID)
	s.Equal("Test Client", response.ClientName)
}

// TestRegisterClient_JWKSUriProvided tests registration with JWKS_URI
func (s *DCRServiceTestSuite) TestRegisterClient_JWKSUriProvided() {
	request := &DCRRegistrationRequest{
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
		"CreateApplication", mock.AnythingOfType("*model.ApplicationDTO"),
	).Return(appDTO, (*serviceerror.ServiceError)(nil))

	response, err := s.service.RegisterClient(request)

	s.NotNil(response)
	s.Nil(err)
	s.Equal("https://client.example.com/.well-known/jwks.json", response.JWKSUri)
}

// TestRegisterClient_ApplicationServiceError tests application service error handling
func (s *DCRServiceTestSuite) TestRegisterClient_ApplicationServiceError() {
	request := &DCRRegistrationRequest{
		RedirectURIs: []string{"not-a-valid-uri"},
		GrantTypes:   []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
	}

	appServiceErr := &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "APP-1014",
		Error:            "Invalid URI",
		ErrorDescription: "The redirect URI is invalid",
	}

	s.mockAppService.On("CreateApplication", mock.AnythingOfType("*model.ApplicationDTO")).Return(nil, appServiceErr)

	response, err := s.service.RegisterClient(request)

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
			name:            "Redirect URI Error APP-1006",
			appErrCode:      "APP-1006",
			expectedDCRCode: ErrorInvalidRedirectURI.Code,
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
