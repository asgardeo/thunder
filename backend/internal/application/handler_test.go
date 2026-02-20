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
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/cert"
	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/system/error/apierror"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
)

type HandlerTestSuite struct {
	suite.Suite
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}

func (suite *HandlerTestSuite) TestNewApplicationHandler() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	assert.NotNil(suite.T(), handler)
	assert.Equal(suite.T(), mockService, handler.service)
}

func (suite *HandlerTestSuite) TestHandleApplicationPostRequest_Success() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	appRequest := model.ApplicationRequest{
		Name:        "TestApp",
		Description: "Test Description",
	}

	expectedApp := &model.ApplicationDTO{
		ID:          "test-app-id",
		Name:        "TestApp",
		Description: "Test Description",
	}

	mockService.On("CreateApplication", mock.AnythingOfType("*model.ApplicationDTO")).Return(expectedApp, nil)

	body, _ := json.Marshal(appRequest)
	req := httptest.NewRequest(http.MethodPost, "/applications", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleApplicationPostRequest(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"))

	var response model.ApplicationCompleteResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test-app-id", response.ID)
	assert.Equal(suite.T(), "TestApp", response.Name)

	mockService.AssertExpectations(suite.T())
}

func (suite *HandlerTestSuite) TestHandleApplicationPostRequest_SuccessWithOAuth() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	appRequest := model.ApplicationRequest{
		Name:        "TestApp",
		Description: "Test Description",
		InboundAuthConfig: []model.InboundAuthConfigComplete{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigComplete{
					ClientID:                "test-client-id",
					ClientSecret:            "test-secret",
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
				},
			},
		},
	}

	expectedApp := &model.ApplicationDTO{
		ID:          "test-app-id",
		Name:        "TestApp",
		Description: "Test Description",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					ClientID:                "test-client-id",
					ClientSecret:            "test-secret",
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
				},
			},
		},
	}

	mockService.On("CreateApplication", mock.AnythingOfType("*model.ApplicationDTO")).Return(expectedApp, nil)

	body, _ := json.Marshal(appRequest)
	req := httptest.NewRequest(http.MethodPost, "/applications", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleApplicationPostRequest(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"))

	var response model.ApplicationCompleteResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test-app-id", response.ID)
	assert.Equal(suite.T(), "TestApp", response.Name)
	assert.Equal(suite.T(), "test-client-id", response.ClientID)

	mockService.AssertExpectations(suite.T())
}

func (suite *HandlerTestSuite) TestHandleApplicationPostRequest_TemplateScenarios() {
	testCases := []struct {
		name             string
		template         string
		expectedTemplate string
	}{
		{
			name:             "with template",
			template:         "spa",
			expectedTemplate: "spa",
		},
		{
			name:             "with empty template",
			template:         "",
			expectedTemplate: "",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			mockService := NewApplicationServiceInterfaceMock(suite.T())
			handler := newApplicationHandler(mockService)

			appRequest := model.ApplicationRequest{
				Name:        "TestApp",
				Description: "Test Description",
				Template:    tc.template,
			}

			expectedApp := &model.ApplicationDTO{
				ID:          "test-app-id",
				Name:        "TestApp",
				Description: "Test Description",
				Template:    tc.expectedTemplate,
			}

			mockService.On("CreateApplication", mock.AnythingOfType("*model.ApplicationDTO")).Return(expectedApp, nil)

			body, _ := json.Marshal(appRequest)
			req := httptest.NewRequest(http.MethodPost, "/applications", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.HandleApplicationPostRequest(w, req)

			assert.Equal(suite.T(), http.StatusCreated, w.Code)

			var response model.ApplicationCompleteResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(suite.T(), err)
			assert.Equal(suite.T(), "test-app-id", response.ID)
			assert.Equal(suite.T(), tc.expectedTemplate, response.Template)

			mockService.AssertExpectations(suite.T())
		})
	}
}

func (suite *HandlerTestSuite) TestHandleApplicationPostRequest_InvalidJSON() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/applications", bytes.NewBufferString("{invalid json}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleApplicationPostRequest(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"))

	var errResp apierror.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), ErrorInvalidRequestFormat.Code, errResp.Code)
}

func (suite *HandlerTestSuite) TestHandleApplicationPostRequest_ServiceError() {
	tests := []struct {
		name           string
		svcErr         *serviceerror.ServiceError
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "InvalidApplicationName",
			svcErr:         &ErrorInvalidApplicationName,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   ErrorInvalidApplicationName.Code,
		},
		{
			name:           "InternalServerError",
			svcErr:         &ErrorInternalServerError,
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   ErrorInternalServerError.Code,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			mockService := NewApplicationServiceInterfaceMock(suite.T())
			handler := newApplicationHandler(mockService)

			appRequest := model.ApplicationRequest{
				Name: "TestApp",
			}

			mockService.On("CreateApplication", mock.AnythingOfType("*model.ApplicationDTO")).Return(nil, tt.svcErr)

			body, _ := json.Marshal(appRequest)
			req := httptest.NewRequest(http.MethodPost, "/applications", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.HandleApplicationPostRequest(w, req)

			assert.Equal(suite.T(), tt.expectedStatus, w.Code)
			assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"))

			var errResp apierror.ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &errResp)
			assert.NoError(suite.T(), err)
			assert.Equal(suite.T(), tt.expectedCode, errResp.Code)

			mockService.AssertExpectations(suite.T())
		})
	}
}

func (suite *HandlerTestSuite) TestHandleApplicationPostRequest_ProcessInboundAuthConfigError() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	appRequest := model.ApplicationRequest{
		Name: "TestApp",
		InboundAuthConfig: []model.InboundAuthConfigComplete{
			{
				Type:           model.OAuthInboundAuthType,
				OAuthAppConfig: nil, // This will cause processInboundAuthConfig to return empty
			},
		},
	}

	// Create app with inbound auth config that has unsupported type
	expectedApp := &model.ApplicationDTO{
		ID:   "test-app-id",
		Name: "TestApp",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type:           "unsupported",
				OAuthAppConfig: nil,
			},
		},
	}

	mockService.On("CreateApplication", mock.AnythingOfType("*model.ApplicationDTO")).Return(expectedApp, nil)

	body, _ := json.Marshal(appRequest)
	req := httptest.NewRequest(http.MethodPost, "/applications", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleApplicationPostRequest(w, req)

	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)

	mockService.AssertExpectations(suite.T())
}

func (suite *HandlerTestSuite) TestHandleApplicationListRequest_Success() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	expectedList := &model.ApplicationListResponse{
		TotalResults: 2,
		Count:        2,
		Applications: []model.BasicApplicationResponse{
			{
				ID:          "app1",
				Name:        "App1",
				Description: "Description 1",
			},
			{
				ID:          "app2",
				Name:        "App2",
				Description: "Description 2",
			},
		},
	}

	mockService.On("GetApplicationList").Return(expectedList, nil)

	req := httptest.NewRequest(http.MethodGet, "/applications", nil)
	w := httptest.NewRecorder()

	handler.HandleApplicationListRequest(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"))

	var response model.ApplicationListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, response.TotalResults)
	assert.Equal(suite.T(), 2, response.Count)
	assert.Len(suite.T(), response.Applications, 2)

	mockService.AssertExpectations(suite.T())
}

func (suite *HandlerTestSuite) TestHandleApplicationListRequest_WithTemplate() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	expectedList := &model.ApplicationListResponse{
		TotalResults: 2,
		Count:        2,
		Applications: []model.BasicApplicationResponse{
			{
				ID:          "app1",
				Name:        "App1",
				Description: "Description 1",
				Template:    "spa",
			},
			{
				ID:          "app2",
				Name:        "App2",
				Description: "Description 2",
				Template:    "mobile",
			},
		},
	}

	mockService.On("GetApplicationList").Return(expectedList, nil)

	req := httptest.NewRequest(http.MethodGet, "/applications", nil)
	w := httptest.NewRecorder()

	handler.HandleApplicationListRequest(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response model.ApplicationListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, response.TotalResults)
	assert.Equal(suite.T(), "spa", response.Applications[0].Template)
	assert.Equal(suite.T(), "mobile", response.Applications[1].Template)

	mockService.AssertExpectations(suite.T())
}

func (suite *HandlerTestSuite) TestHandleApplicationListRequest_ServiceError() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	svcErr := &ErrorInternalServerError

	mockService.On("GetApplicationList").Return(nil, svcErr)

	req := httptest.NewRequest(http.MethodGet, "/applications", nil)
	w := httptest.NewRecorder()

	handler.HandleApplicationListRequest(w, req)

	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"))

	var errResp apierror.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), ErrorInternalServerError.Code, errResp.Code)

	mockService.AssertExpectations(suite.T())
}

func (suite *HandlerTestSuite) TestHandleApplicationGetRequest_Success() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	expectedApp := &model.Application{
		ID:          "test-app-id",
		Name:        "TestApp",
		Description: "Test Description",
	}

	mockService.On("GetApplication", "test-app-id").Return(expectedApp, nil)

	req := httptest.NewRequest(http.MethodGet, "/applications/test-app-id", nil)
	req.SetPathValue("id", "test-app-id")
	w := httptest.NewRecorder()

	handler.HandleApplicationGetRequest(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"))

	var response model.ApplicationGetResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test-app-id", response.ID)
	assert.Equal(suite.T(), "TestApp", response.Name)

	mockService.AssertExpectations(suite.T())
}

func (suite *HandlerTestSuite) TestHandleApplicationGetRequest_SuccessWithOAuth() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	expectedApp := &model.Application{
		ID:          "test-app-id",
		Name:        "TestApp",
		Description: "Test Description",
		InboundAuthConfig: []model.InboundAuthConfigComplete{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigComplete{
					ClientID:                "test-client-id",
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
				},
			},
		},
	}

	mockService.On("GetApplication", "test-app-id").Return(expectedApp, nil)

	req := httptest.NewRequest(http.MethodGet, "/applications/test-app-id", nil)
	req.SetPathValue("id", "test-app-id")
	w := httptest.NewRecorder()

	handler.HandleApplicationGetRequest(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"))

	var response model.ApplicationGetResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test-app-id", response.ID)
	assert.Equal(suite.T(), "TestApp", response.Name)
	assert.Equal(suite.T(), "test-client-id", response.ClientID)

	mockService.AssertExpectations(suite.T())
}

func (suite *HandlerTestSuite) TestHandleApplicationGetRequest_WithTemplate() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	expectedApp := &model.Application{
		ID:          "test-app-id",
		Name:        "TestApp",
		Description: "Test Description",
		ThemeID:     "theme-123",
		LayoutID:    "layout-456",
		Template:    "spa",
	}

	mockService.On("GetApplication", "test-app-id").Return(expectedApp, nil)

	req := httptest.NewRequest(http.MethodGet, "/applications/test-app-id", nil)
	req.SetPathValue("id", "test-app-id")
	w := httptest.NewRecorder()

	handler.HandleApplicationGetRequest(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response model.ApplicationGetResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test-app-id", response.ID)
	assert.Equal(suite.T(), "theme-123", response.ThemeID)
	assert.Equal(suite.T(), "layout-456", response.LayoutID)
	assert.Equal(suite.T(), "spa", response.Template)

	mockService.AssertExpectations(suite.T())
}

func (suite *HandlerTestSuite) TestHandleApplicationGetRequest_WithEmptyTemplate() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	expectedApp := &model.Application{
		ID:          "test-app-id",
		Name:        "TestApp",
		Description: "Test Description",
		Template:    "",
	}

	mockService.On("GetApplication", "test-app-id").Return(expectedApp, nil)

	req := httptest.NewRequest(http.MethodGet, "/applications/test-app-id", nil)
	req.SetPathValue("id", "test-app-id")
	w := httptest.NewRecorder()

	handler.HandleApplicationGetRequest(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response model.ApplicationGetResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test-app-id", response.ID)
	assert.Equal(suite.T(), "", response.Template)

	mockService.AssertExpectations(suite.T())
}

func (suite *HandlerTestSuite) TestHandleApplicationGetRequest_InvalidID() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/applications/", nil)
	req.SetPathValue("id", "")
	w := httptest.NewRecorder()

	handler.HandleApplicationGetRequest(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"))

	var errResp apierror.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), ErrorInvalidApplicationID.Code, errResp.Code)
}

//nolint:dupl // Testing different error scenarios
func (suite *HandlerTestSuite) TestHandleApplicationGetRequest_NotFound() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	svcErr := &ErrorApplicationNotFound

	mockService.On("GetApplication", "non-existent-id").Return(nil, svcErr)

	req := httptest.NewRequest(http.MethodGet, "/applications/non-existent-id", nil)
	req.SetPathValue("id", "non-existent-id")
	w := httptest.NewRecorder()

	handler.HandleApplicationGetRequest(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
	assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"))

	var errResp apierror.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), ErrorApplicationNotFound.Code, errResp.Code)

	mockService.AssertExpectations(suite.T())
}

//nolint:dupl // Testing different error scenarios
func (suite *HandlerTestSuite) TestHandleApplicationGetRequest_ServiceError() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	svcErr := &ErrorInternalServerError

	mockService.On("GetApplication", "test-app-id").Return(nil, svcErr)

	req := httptest.NewRequest(http.MethodGet, "/applications/test-app-id", nil)
	req.SetPathValue("id", "test-app-id")
	w := httptest.NewRecorder()

	handler.HandleApplicationGetRequest(w, req)

	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"))

	var errResp apierror.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), ErrorInternalServerError.Code, errResp.Code)

	mockService.AssertExpectations(suite.T())
}

func (suite *HandlerTestSuite) TestHandleApplicationGetRequest_UnsupportedInboundAuthType() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	expectedApp := &model.Application{
		ID:   "test-app-id",
		Name: "TestApp",
		InboundAuthConfig: []model.InboundAuthConfigComplete{
			{
				Type:           "unsupported",
				OAuthAppConfig: nil,
			},
		},
	}

	mockService.On("GetApplication", "test-app-id").Return(expectedApp, nil)

	req := httptest.NewRequest(http.MethodGet, "/applications/test-app-id", nil)
	req.SetPathValue("id", "test-app-id")
	w := httptest.NewRecorder()

	handler.HandleApplicationGetRequest(w, req)

	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)

	mockService.AssertExpectations(suite.T())
}

func (suite *HandlerTestSuite) TestHandleApplicationGetRequest_NilOAuthConfig() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	expectedApp := &model.Application{
		ID:   "test-app-id",
		Name: "TestApp",
		InboundAuthConfig: []model.InboundAuthConfigComplete{
			{
				Type:           model.OAuthInboundAuthType,
				OAuthAppConfig: nil,
			},
		},
	}

	mockService.On("GetApplication", "test-app-id").Return(expectedApp, nil)

	req := httptest.NewRequest(http.MethodGet, "/applications/test-app-id", nil)
	req.SetPathValue("id", "test-app-id")
	w := httptest.NewRecorder()

	handler.HandleApplicationGetRequest(w, req)

	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)

	mockService.AssertExpectations(suite.T())
}

func (suite *HandlerTestSuite) TestHandleApplicationPutRequest_Success() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	appRequest := model.ApplicationRequest{
		Name:        "UpdatedApp",
		Description: "Updated Description",
	}

	expectedApp := &model.ApplicationDTO{
		ID:          "test-app-id",
		Name:        "UpdatedApp",
		Description: "Updated Description",
	}

	mockService.On("UpdateApplication", "test-app-id", mock.AnythingOfType("*model.ApplicationDTO")).
		Return(expectedApp, nil)

	body, _ := json.Marshal(appRequest)
	req := httptest.NewRequest(http.MethodPut, "/applications/test-app-id", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "test-app-id")
	w := httptest.NewRecorder()

	handler.HandleApplicationPutRequest(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"))

	var response model.ApplicationCompleteResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test-app-id", response.ID)
	assert.Equal(suite.T(), "UpdatedApp", response.Name)

	mockService.AssertExpectations(suite.T())
}

func (suite *HandlerTestSuite) TestHandleApplicationPutRequest_WithTemplate() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	appRequest := model.ApplicationRequest{
		Name:        "UpdatedApp",
		Description: "Updated Description",
		Template:    "mobile",
	}

	expectedApp := &model.ApplicationDTO{
		ID:          "test-app-id",
		Name:        "UpdatedApp",
		Description: "Updated Description",
		Template:    "mobile",
	}

	mockService.On("UpdateApplication", "test-app-id", mock.AnythingOfType("*model.ApplicationDTO")).
		Return(expectedApp, nil)

	body, _ := json.Marshal(appRequest)
	req := httptest.NewRequest(http.MethodPut, "/applications/test-app-id", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "test-app-id")
	w := httptest.NewRecorder()

	handler.HandleApplicationPutRequest(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response model.ApplicationCompleteResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test-app-id", response.ID)
	assert.Equal(suite.T(), "mobile", response.Template)

	mockService.AssertExpectations(suite.T())
}

func (suite *HandlerTestSuite) TestHandleApplicationPutRequest_TemplateScenarios() {
	testCases := []struct {
		name             string
		template         string
		expectedTemplate string
	}{
		{
			name:             "update template",
			template:         "traditional_web_app",
			expectedTemplate: "traditional_web_app",
		},
		{
			name:             "clear template",
			template:         "",
			expectedTemplate: "",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			mockService := NewApplicationServiceInterfaceMock(suite.T())
			handler := newApplicationHandler(mockService)

			appRequest := model.ApplicationRequest{
				Name:        "UpdatedApp",
				Description: "Updated Description",
				Template:    tc.template,
			}

			expectedApp := &model.ApplicationDTO{
				ID:          "test-app-id",
				Name:        "UpdatedApp",
				Description: "Updated Description",
				Template:    tc.expectedTemplate,
			}

			mockService.On("UpdateApplication", "test-app-id", mock.AnythingOfType("*model.ApplicationDTO")).
				Return(expectedApp, nil)

			body, _ := json.Marshal(appRequest)
			req := httptest.NewRequest(http.MethodPut, "/applications/test-app-id", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.SetPathValue("id", "test-app-id")
			w := httptest.NewRecorder()

			handler.HandleApplicationPutRequest(w, req)

			assert.Equal(suite.T(), http.StatusOK, w.Code)

			var response model.ApplicationCompleteResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(suite.T(), err)
			assert.Equal(suite.T(), tc.expectedTemplate, response.Template)

			mockService.AssertExpectations(suite.T())
		})
	}
}

func (suite *HandlerTestSuite) TestHandleApplicationPutRequest_InvalidID() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	appRequest := model.ApplicationRequest{
		Name: "UpdatedApp",
	}

	body, _ := json.Marshal(appRequest)
	req := httptest.NewRequest(http.MethodPut, "/applications/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "")
	w := httptest.NewRecorder()

	handler.HandleApplicationPutRequest(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"))

	var errResp apierror.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), ErrorInvalidApplicationID.Code, errResp.Code)
}

func (suite *HandlerTestSuite) TestHandleApplicationPutRequest_InvalidJSON() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	req := httptest.NewRequest(http.MethodPut, "/applications/test-app-id", bytes.NewBufferString("{invalid json}"))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "test-app-id")
	w := httptest.NewRecorder()

	handler.HandleApplicationPutRequest(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"))

	var errResp apierror.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), ErrorInvalidRequestFormat.Code, errResp.Code)
}

func (suite *HandlerTestSuite) TestHandleApplicationPutRequest_ServiceError() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	appRequest := model.ApplicationRequest{
		Name: "UpdatedApp",
	}

	svcErr := &ErrorInvalidApplicationName

	mockService.On("UpdateApplication", "test-app-id", mock.AnythingOfType("*model.ApplicationDTO")).Return(nil, svcErr)

	body, _ := json.Marshal(appRequest)
	req := httptest.NewRequest(http.MethodPut, "/applications/test-app-id", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "test-app-id")
	w := httptest.NewRecorder()

	handler.HandleApplicationPutRequest(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"))

	var errResp apierror.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), ErrorInvalidApplicationName.Code, errResp.Code)

	mockService.AssertExpectations(suite.T())
}

func (suite *HandlerTestSuite) TestHandleApplicationPutRequest_NotFound() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	appRequest := model.ApplicationRequest{
		Name: "UpdatedApp",
	}

	svcErr := &ErrorApplicationNotFound

	mockService.On("UpdateApplication", "non-existent-id", mock.AnythingOfType("*model.ApplicationDTO")).
		Return(nil, svcErr)

	body, _ := json.Marshal(appRequest)
	req := httptest.NewRequest(http.MethodPut, "/applications/non-existent-id", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "non-existent-id")
	w := httptest.NewRecorder()

	handler.HandleApplicationPutRequest(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	mockService.AssertExpectations(suite.T())
}

func (suite *HandlerTestSuite) TestHandleApplicationDeleteRequest_Success() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	mockService.On("DeleteApplication", "test-app-id").Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/applications/test-app-id", nil)
	req.SetPathValue("id", "test-app-id")
	w := httptest.NewRecorder()

	handler.HandleApplicationDeleteRequest(w, req)

	assert.Equal(suite.T(), http.StatusNoContent, w.Code)

	mockService.AssertExpectations(suite.T())
}

func (suite *HandlerTestSuite) TestHandleApplicationDeleteRequest_InvalidID() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	req := httptest.NewRequest(http.MethodDelete, "/applications/", nil)
	req.SetPathValue("id", "")
	w := httptest.NewRecorder()

	handler.HandleApplicationDeleteRequest(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"))

	var errResp apierror.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), ErrorInvalidApplicationID.Code, errResp.Code)
}

func (suite *HandlerTestSuite) TestHandleApplicationDeleteRequest_NotFound() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	svcErr := &ErrorApplicationNotFound

	mockService.On("DeleteApplication", "non-existent-id").Return(svcErr)

	req := httptest.NewRequest(http.MethodDelete, "/applications/non-existent-id", nil)
	req.SetPathValue("id", "non-existent-id")
	w := httptest.NewRecorder()

	handler.HandleApplicationDeleteRequest(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	mockService.AssertExpectations(suite.T())
}

func (suite *HandlerTestSuite) TestHandleApplicationDeleteRequest_ServiceError() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	svcErr := &ErrorInternalServerError

	mockService.On("DeleteApplication", "test-app-id").Return(svcErr)

	req := httptest.NewRequest(http.MethodDelete, "/applications/test-app-id", nil)
	req.SetPathValue("id", "test-app-id")
	w := httptest.NewRecorder()

	handler.HandleApplicationDeleteRequest(w, req)

	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)

	mockService.AssertExpectations(suite.T())
}

func (suite *HandlerTestSuite) TestProcessInboundAuthConfig_Success() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)
	logger := log.GetLogger()

	appDTO := &model.ApplicationDTO{
		ID:   "test-app-id",
		Name: "TestApp",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					ClientID:                "test-client-id",
					ClientSecret:            "test-secret",
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
				},
			},
		},
	}

	returnApp := &model.ApplicationCompleteResponse{
		ID:   "test-app-id",
		Name: "TestApp",
	}

	success := handler.processInboundAuthConfig(logger, appDTO, returnApp)

	assert.True(suite.T(), success)
	assert.NotNil(suite.T(), returnApp.InboundAuthConfig)
	assert.Len(suite.T(), returnApp.InboundAuthConfig, 1)
	assert.Equal(suite.T(), "test-client-id", returnApp.ClientID)
}

func (suite *HandlerTestSuite) TestProcessInboundAuthConfig_EmptyRedirectURIs() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)
	logger := log.GetLogger()

	appDTO := &model.ApplicationDTO{
		ID:   "test-app-id",
		Name: "TestApp",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					ClientID:                "test-client-id",
					ClientSecret:            "test-secret",
					RedirectURIs:            nil, // Empty redirect URIs
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
				},
			},
		},
	}

	returnApp := &model.ApplicationCompleteResponse{
		ID:   "test-app-id",
		Name: "TestApp",
	}

	success := handler.processInboundAuthConfig(logger, appDTO, returnApp)

	assert.True(suite.T(), success)
	assert.NotNil(suite.T(), returnApp.InboundAuthConfig)
	assert.Empty(suite.T(), returnApp.InboundAuthConfig[0].OAuthAppConfig.RedirectURIs)
}

func (suite *HandlerTestSuite) TestProcessInboundAuthConfig_EmptyGrantTypes() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)
	logger := log.GetLogger()

	appDTO := &model.ApplicationDTO{
		ID:   "test-app-id",
		Name: "TestApp",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					ClientID:                "test-client-id",
					ClientSecret:            "test-secret",
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              nil, // Empty grant types
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
				},
			},
		},
	}

	returnApp := &model.ApplicationCompleteResponse{
		ID:   "test-app-id",
		Name: "TestApp",
	}

	success := handler.processInboundAuthConfig(logger, appDTO, returnApp)

	assert.True(suite.T(), success)
	assert.NotNil(suite.T(), returnApp.InboundAuthConfig)
	assert.Empty(suite.T(), returnApp.InboundAuthConfig[0].OAuthAppConfig.GrantTypes)
}

func (suite *HandlerTestSuite) TestProcessInboundAuthConfig_UnsupportedType() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)
	logger := log.GetLogger()

	appDTO := &model.ApplicationDTO{
		ID:   "test-app-id",
		Name: "TestApp",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type:           "unsupported",
				OAuthAppConfig: nil,
			},
		},
	}

	returnApp := &model.ApplicationCompleteResponse{
		ID:   "test-app-id",
		Name: "TestApp",
	}

	success := handler.processInboundAuthConfig(logger, appDTO, returnApp)

	assert.False(suite.T(), success)
}

func (suite *HandlerTestSuite) TestProcessInboundAuthConfig_NilOAuthConfig() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)
	logger := log.GetLogger()

	appDTO := &model.ApplicationDTO{
		ID:   "test-app-id",
		Name: "TestApp",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type:           model.OAuthInboundAuthType,
				OAuthAppConfig: nil,
			},
		},
	}

	returnApp := &model.ApplicationCompleteResponse{
		ID:   "test-app-id",
		Name: "TestApp",
	}

	success := handler.processInboundAuthConfig(logger, appDTO, returnApp)

	assert.False(suite.T(), success)
}

func (suite *HandlerTestSuite) TestProcessInboundAuthConfig_EmptyInboundAuthConfig() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)
	logger := log.GetLogger()

	appDTO := &model.ApplicationDTO{
		ID:                "test-app-id",
		Name:              "TestApp",
		InboundAuthConfig: []model.InboundAuthConfigDTO{},
	}

	returnApp := &model.ApplicationCompleteResponse{
		ID:   "test-app-id",
		Name: "TestApp",
	}

	success := handler.processInboundAuthConfig(logger, appDTO, returnApp)

	assert.True(suite.T(), success)
}

func (suite *HandlerTestSuite) TestHandleError_ClientError() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	w := httptest.NewRecorder()

	svcErr := &ErrorInvalidApplicationName

	handler.handleError(w, svcErr)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"))

	var errResp apierror.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), ErrorInvalidApplicationName.Code, errResp.Code)
}

func (suite *HandlerTestSuite) TestHandleError_NotFoundError() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	w := httptest.NewRecorder()

	svcErr := &ErrorApplicationNotFound

	handler.handleError(w, svcErr)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
	assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"))

	var errResp apierror.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), ErrorApplicationNotFound.Code, errResp.Code)
}

func (suite *HandlerTestSuite) TestHandleError_ServerError() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	w := httptest.NewRecorder()

	svcErr := &ErrorInternalServerError

	handler.handleError(w, svcErr)

	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"))

	var errResp apierror.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), ErrorInternalServerError.Code, errResp.Code)
}

func (suite *HandlerTestSuite) TestProcessInboundAuthConfigFromRequest_Success() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	configs := []model.InboundAuthConfigComplete{
		{
			Type: model.OAuthInboundAuthType,
			OAuthAppConfig: &model.OAuthAppConfigComplete{
				ClientID:                "test-client-id",
				ClientSecret:            "test-secret",
				RedirectURIs:            []string{"https://example.com/callback"},
				GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
				ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
				TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
			},
		},
	}

	result := handler.processInboundAuthConfigFromRequest(configs)

	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result, 1)
	assert.Equal(suite.T(), model.OAuthInboundAuthType, result[0].Type)
	assert.Equal(suite.T(), "test-client-id", result[0].OAuthAppConfig.ClientID)
}

func (suite *HandlerTestSuite) TestProcessInboundAuthConfigFromRequest_EmptyConfigs() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	configs := []model.InboundAuthConfigComplete{}

	result := handler.processInboundAuthConfigFromRequest(configs)

	assert.Nil(suite.T(), result)
}

func (suite *HandlerTestSuite) TestProcessInboundAuthConfigFromRequest_NilConfigs() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	result := handler.processInboundAuthConfigFromRequest(nil)

	assert.Nil(suite.T(), result)
}

func (suite *HandlerTestSuite) TestProcessInboundAuthConfigFromRequest_UnsupportedType() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	configs := []model.InboundAuthConfigComplete{
		{
			Type:           "unsupported",
			OAuthAppConfig: nil,
		},
	}

	result := handler.processInboundAuthConfigFromRequest(configs)

	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result, 0) // Should skip unsupported types
}

func (suite *HandlerTestSuite) TestProcessInboundAuthConfigFromRequest_NilOAuthConfig() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	configs := []model.InboundAuthConfigComplete{
		{
			Type:           model.OAuthInboundAuthType,
			OAuthAppConfig: nil,
		},
	}

	result := handler.processInboundAuthConfigFromRequest(configs)

	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result, 0) // Should skip configs with nil OAuth config
}

func (suite *HandlerTestSuite) TestProcessInboundAuthConfigFromRequest_MultipleConfigs() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	configs := []model.InboundAuthConfigComplete{
		{
			Type: model.OAuthInboundAuthType,
			OAuthAppConfig: &model.OAuthAppConfigComplete{
				ClientID:     "client-1",
				ClientSecret: "secret-1",
			},
		},
		{
			Type: model.OAuthInboundAuthType,
			OAuthAppConfig: &model.OAuthAppConfigComplete{
				ClientID:     "client-2",
				ClientSecret: "secret-2",
			},
		},
	}

	result := handler.processInboundAuthConfigFromRequest(configs)

	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result, 2)
	assert.Equal(suite.T(), "client-1", result[0].OAuthAppConfig.ClientID)
	assert.Equal(suite.T(), "client-2", result[1].OAuthAppConfig.ClientID)
}

func (suite *HandlerTestSuite) TestProcessInboundAuthConfigFromRequest_WithTokenConfig() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	configs := []model.InboundAuthConfigComplete{
		{
			Type: model.OAuthInboundAuthType,
			OAuthAppConfig: &model.OAuthAppConfigComplete{
				ClientID:     "test-client-id",
				ClientSecret: "test-secret",
				Token: &model.OAuthTokenConfig{
					AccessToken: &model.AccessTokenConfig{
						ValidityPeriod: 3600,
						UserAttributes: []string{"email", "name"},
					},
					IDToken: &model.IDTokenConfig{
						ValidityPeriod: 3600,
						UserAttributes: []string{"email"},
					},
				},
			},
		},
	}

	result := handler.processInboundAuthConfigFromRequest(configs)

	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result, 1)
	assert.NotNil(suite.T(), result[0].OAuthAppConfig.Token)
}

func (suite *HandlerTestSuite) TestProcessInboundAuthConfigFromRequest_WithScopes() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	configs := []model.InboundAuthConfigComplete{
		{
			Type: model.OAuthInboundAuthType,
			OAuthAppConfig: &model.OAuthAppConfigComplete{
				ClientID:     "test-client-id",
				ClientSecret: "test-secret",
				Scopes:       []string{"openid", "profile", "email"},
			},
		},
	}

	result := handler.processInboundAuthConfigFromRequest(configs)

	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result, 1)
	assert.Equal(suite.T(), []string{"openid", "profile", "email"}, result[0].OAuthAppConfig.Scopes)
}

func (suite *HandlerTestSuite) TestProcessInboundAuthConfigFromRequest_PublicClient() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	configs := []model.InboundAuthConfigComplete{
		{
			Type: model.OAuthInboundAuthType,
			OAuthAppConfig: &model.OAuthAppConfigComplete{
				ClientID:                "test-client-id",
				PublicClient:            true,
				PKCERequired:            true,
				TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodNone,
			},
		},
	}

	result := handler.processInboundAuthConfigFromRequest(configs)

	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result, 1)
	assert.True(suite.T(), result[0].OAuthAppConfig.PublicClient)
	assert.True(suite.T(), result[0].OAuthAppConfig.PKCERequired)
}

func (suite *HandlerTestSuite) TestHandleApplicationPostRequest_WithCertificate() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	appRequest := model.ApplicationRequest{
		Name:        "TestApp",
		Description: "Test Description",
		Certificate: &model.ApplicationCertificate{
			Type:  cert.CertificateTypeJWKS,
			Value: `{"keys":[{"kty":"RSA","kid":"test"}]}`,
		},
	}

	expectedApp := &model.ApplicationDTO{
		ID:          "test-app-id",
		Name:        "TestApp",
		Description: "Test Description",
		Certificate: &model.ApplicationCertificate{
			Type:  cert.CertificateTypeJWKS,
			Value: `{"keys":[{"kty":"RSA","kid":"test"}]}`,
		},
	}

	mockService.On("CreateApplication", mock.AnythingOfType("*model.ApplicationDTO")).Return(expectedApp, nil)

	body, _ := json.Marshal(appRequest)
	req := httptest.NewRequest(http.MethodPost, "/applications", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleApplicationPostRequest(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	mockService.AssertExpectations(suite.T())
}

func (suite *HandlerTestSuite) TestHandleApplicationGetRequest_WithEmptyArrays() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	expectedApp := &model.Application{
		ID:          "test-app-id",
		Name:        "TestApp",
		Description: "Test Description",
		InboundAuthConfig: []model.InboundAuthConfigComplete{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigComplete{
					ClientID:                "test-client-id",
					RedirectURIs:            []string{}, // Empty array
					GrantTypes:              []oauth2const.GrantType{},
					ResponseTypes:           []oauth2const.ResponseType{},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
				},
			},
		},
	}

	mockService.On("GetApplication", "test-app-id").Return(expectedApp, nil)

	req := httptest.NewRequest(http.MethodGet, "/applications/test-app-id", nil)
	req.SetPathValue("id", "test-app-id")
	w := httptest.NewRecorder()

	handler.HandleApplicationGetRequest(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response model.ApplicationGetResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response.InboundAuthConfig)
	assert.Empty(suite.T(), response.InboundAuthConfig[0].OAuthAppConfig.RedirectURIs)
	assert.Empty(suite.T(), response.InboundAuthConfig[0].OAuthAppConfig.GrantTypes)
	assert.Empty(suite.T(), response.InboundAuthConfig[0].OAuthAppConfig.ResponseTypes)

	mockService.AssertExpectations(suite.T())
}

func (suite *HandlerTestSuite) TestHandleApplicationPutRequest_WithOAuth() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	appRequest := model.ApplicationRequest{
		Name:        "UpdatedApp",
		Description: "Updated Description",
		InboundAuthConfig: []model.InboundAuthConfigComplete{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigComplete{
					ClientID:                "updated-client-id",
					ClientSecret:            "updated-secret",
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
				},
			},
		},
	}

	expectedApp := &model.ApplicationDTO{
		ID:          "test-app-id",
		Name:        "UpdatedApp",
		Description: "Updated Description",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					ClientID:                "updated-client-id",
					ClientSecret:            "updated-secret",
					RedirectURIs:            []string{"https://example.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
				},
			},
		},
	}

	mockService.On("UpdateApplication", "test-app-id", mock.AnythingOfType("*model.ApplicationDTO")).
		Return(expectedApp, nil)

	body, _ := json.Marshal(appRequest)
	req := httptest.NewRequest(http.MethodPut, "/applications/test-app-id", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "test-app-id")
	w := httptest.NewRecorder()

	handler.HandleApplicationPutRequest(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response model.ApplicationCompleteResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "updated-client-id", response.ClientID)

	mockService.AssertExpectations(suite.T())
}

// failingResponseWriter is a mock http.ResponseWriter that fails on Write
type failingResponseWriter struct {
	header        http.Header
	statusCode    int
	failOnce      bool
	writeCount    int
	headerWritten bool
}

func (f *failingResponseWriter) Header() http.Header {
	if f.header == nil {
		f.header = make(http.Header)
	}
	return f.header
}

func (f *failingResponseWriter) Write(b []byte) (int, error) {
	// Auto-set status code if not set
	if !f.headerWritten {
		f.statusCode = http.StatusOK
		f.headerWritten = true
	}

	f.writeCount++
	if f.failOnce || f.writeCount > 1 {
		return 0, assert.AnError
	}
	return len(b), nil
}

func (f *failingResponseWriter) WriteHeader(statusCode int) {
	// Only allow setting status code once
	if !f.headerWritten {
		f.statusCode = statusCode
		f.headerWritten = true
	}
}

func (suite *HandlerTestSuite) TestHandleApplicationListRequest_EncodeResponseError() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	listResponse := &model.ApplicationListResponse{
		Applications: []model.BasicApplicationResponse{
			{
				ID:          "test-app-1",
				Name:        "Test App 1",
				Description: "Test Description 1",
			},
		},
		TotalResults: 1,
		Count:        1,
	}

	mockService.On("GetApplicationList").Return(listResponse, nil)

	req := httptest.NewRequest(http.MethodGet, "/applications", nil)
	w := &failingResponseWriter{failOnce: true}

	handler.HandleApplicationListRequest(w, req)

	// Should attempt to write but fail - verify service was called
	mockService.AssertExpectations(suite.T())
}

func (suite *HandlerTestSuite) TestHandleApplicationGetRequest_EncodeErrorResponseFails() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	// Test with empty id - should try to encode error response
	req := httptest.NewRequest(http.MethodGet, "/applications/", nil)
	req.SetPathValue("id", "")
	w := &failingResponseWriter{failOnce: true}

	handler.HandleApplicationGetRequest(w, req)

	// Error response encoding should fail
	assert.Equal(suite.T(), http.StatusBadRequest, w.statusCode)
}

func (suite *HandlerTestSuite) TestHandleApplicationPutRequest_EncodeErrorResponseFailsOnEmptyID() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	req := httptest.NewRequest(http.MethodPut, "/applications/", nil)
	req.SetPathValue("id", "")
	w := &failingResponseWriter{failOnce: true}

	handler.HandleApplicationPutRequest(w, req)

	// Error response encoding should fail
	assert.Equal(suite.T(), http.StatusBadRequest, w.statusCode)
}

func (suite *HandlerTestSuite) TestHandleApplicationPutRequest_EncodeErrorResponseFailsOnInvalidJSON() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	// Invalid JSON body
	req := httptest.NewRequest(http.MethodPut, "/applications/test-id", bytes.NewBufferString("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "test-id")
	w := &failingResponseWriter{failOnce: true}

	handler.HandleApplicationPutRequest(w, req)

	// Error response encoding should fail
	assert.Equal(suite.T(), http.StatusBadRequest, w.statusCode)
}

func (suite *HandlerTestSuite) TestHandleApplicationPostRequest_EncodeResponseError() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	appRequest := model.ApplicationRequest{
		Name:        "Test App",
		Description: "Test Description",
	}

	createdApp := &model.ApplicationDTO{
		ID:          "test-app-id",
		Name:        "Test App",
		Description: "Test Description",
	}

	mockService.On("CreateApplication", mock.AnythingOfType("*model.ApplicationDTO")).Return(createdApp, nil)

	body, _ := json.Marshal(appRequest)
	req := httptest.NewRequest(http.MethodPost, "/applications", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := &failingResponseWriter{failOnce: true}

	handler.HandleApplicationPostRequest(w, req)

	// Should attempt to write response but fail
	mockService.AssertExpectations(suite.T())
	assert.Equal(suite.T(), http.StatusCreated, w.statusCode)
}

func (suite *HandlerTestSuite) TestHandleApplicationPutRequest_EncodeResponseError() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	appRequest := model.ApplicationRequest{
		Name:        "Updated App",
		Description: "Updated Description",
	}

	updatedApp := &model.ApplicationDTO{
		ID:          "test-app-id",
		Name:        "Updated App",
		Description: "Updated Description",
	}

	mockService.On("UpdateApplication", "test-app-id", mock.AnythingOfType("*model.ApplicationDTO")).
		Return(updatedApp, nil)

	body, _ := json.Marshal(appRequest)
	req := httptest.NewRequest(http.MethodPut, "/applications/test-app-id", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "test-app-id")
	w := &failingResponseWriter{failOnce: true}

	handler.HandleApplicationPutRequest(w, req)

	// Should attempt to write response but fail
	mockService.AssertExpectations(suite.T())
	assert.Equal(suite.T(), http.StatusOK, w.statusCode)
}

func (suite *HandlerTestSuite) TestHandleApplicationGetRequest_EncodeResponseError() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	expectedApp := &model.Application{
		ID:          "test-app-id",
		Name:        "Test App",
		Description: "Test Description",
	}

	mockService.On("GetApplication", "test-app-id").Return(expectedApp, nil)

	req := httptest.NewRequest(http.MethodGet, "/applications/test-app-id", nil)
	req.SetPathValue("id", "test-app-id")
	w := &failingResponseWriter{failOnce: true}

	handler.HandleApplicationGetRequest(w, req)

	// Should attempt to write response but fail
	mockService.AssertExpectations(suite.T())
	assert.Equal(suite.T(), http.StatusOK, w.statusCode)
}

func (suite *HandlerTestSuite) TestHandleApplicationPostRequest_MultipleInboundAuthConfigs() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	// Test with multiple OAuth configs (edge case - should only process first one properly)
	appRequest := model.ApplicationRequest{
		Name:        "Multi Config App",
		Description: "App with multiple inbound auth configs",
		InboundAuthConfig: []model.InboundAuthConfigComplete{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigComplete{
					ClientID:                "client-1",
					ClientSecret:            "secret-1",
					RedirectURIs:            []string{"https://example1.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
				},
			},
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigComplete{
					ClientID:                "client-2",
					ClientSecret:            "secret-2",
					RedirectURIs:            []string{"https://example2.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeClientCredentials},
					ResponseTypes:           []oauth2const.ResponseType{},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretPost,
				},
			},
		},
	}

	createdApp := &model.ApplicationDTO{
		ID:          "multi-config-app-id",
		Name:        "Multi Config App",
		Description: "App with multiple inbound auth configs",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					ClientID:                "client-1",
					ClientSecret:            "secret-1",
					RedirectURIs:            []string{"https://example1.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
				},
			},
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					ClientID:                "client-2",
					ClientSecret:            "secret-2",
					RedirectURIs:            []string{"https://example2.com/callback"},
					GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeClientCredentials},
					ResponseTypes:           []oauth2const.ResponseType{},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretPost,
				},
			},
		},
	}

	mockService.On("CreateApplication", mock.AnythingOfType("*model.ApplicationDTO")).Return(createdApp, nil)

	body, _ := json.Marshal(appRequest)
	req := httptest.NewRequest(http.MethodPost, "/applications", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleApplicationPostRequest(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response model.ApplicationCompleteResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "client-1", response.ClientID)
	// Should have both inbound auth configs in response
	assert.Len(suite.T(), response.InboundAuthConfig, 2)

	mockService.AssertExpectations(suite.T())
}

func (suite *HandlerTestSuite) TestHandleError_EncodeErrorResponseFails() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	mockService.On("GetApplication", "test-id").Return(nil, &ErrorApplicationNotFound)

	req := httptest.NewRequest(http.MethodGet, "/applications/test-id", nil)
	req.SetPathValue("id", "test-id")
	w := &failingResponseWriter{failOnce: true}

	handler.HandleApplicationGetRequest(w, req)

	// Should try to encode error response but fail
	assert.Equal(suite.T(), http.StatusNotFound, w.statusCode)
	mockService.AssertExpectations(suite.T())
}

func (suite *HandlerTestSuite) TestHandleApplicationDeleteRequest_EncodeErrorResponseFails() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	// Test with empty id to trigger error encoding
	req := httptest.NewRequest(http.MethodDelete, "/applications/", nil)
	req.SetPathValue("id", "")
	w := &failingResponseWriter{failOnce: true}

	handler.HandleApplicationDeleteRequest(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.statusCode)
}

func (suite *HandlerTestSuite) TestHandleApplicationPostRequest_UnsupportedInboundAuthType() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	appRequest := model.ApplicationRequest{
		Name:        "TestApp",
		Description: "Test Description",
	}

	// Service returns app with unsupported auth type
	createdApp := &model.ApplicationDTO{
		ID:          "test-app-id",
		Name:        "TestApp",
		Description: "Test Description",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: "UNSUPPORTED_TYPE", // Not OAuthInboundAuthType
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					ClientID: "test-client-id",
				},
			},
		},
	}

	mockService.On("CreateApplication", mock.AnythingOfType("*model.ApplicationDTO")).Return(createdApp, nil)

	body, _ := json.Marshal(appRequest)
	req := httptest.NewRequest(http.MethodPost, "/applications", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleApplicationPostRequest(w, req)

	// Should return 500 because processInboundAuthConfig returns false
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"))

	var errResp apierror.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), ErrorInternalServerError.Code, errResp.Code)

	mockService.AssertExpectations(suite.T())
}

func (suite *HandlerTestSuite) TestHandleApplicationPostRequest_NilOAuthConfig() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	appRequest := model.ApplicationRequest{
		Name:        "TestApp",
		Description: "Test Description",
	}

	// Service returns app with OAuth auth type but nil config
	createdApp := &model.ApplicationDTO{
		ID:          "test-app-id",
		Name:        "TestApp",
		Description: "Test Description",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type:           model.OAuthInboundAuthType,
				OAuthAppConfig: nil, // Nil OAuth config
			},
		},
	}

	mockService.On("CreateApplication", mock.AnythingOfType("*model.ApplicationDTO")).Return(createdApp, nil)

	body, _ := json.Marshal(appRequest)
	req := httptest.NewRequest(http.MethodPost, "/applications", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleApplicationPostRequest(w, req)

	// Should return 500 because processInboundAuthConfig returns false
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"))

	var errResp apierror.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), ErrorInternalServerError.Code, errResp.Code)

	mockService.AssertExpectations(suite.T())
}

func (suite *HandlerTestSuite) TestHandleApplicationPostRequest_ProcessInboundAuthConfigErrorEncodingFails() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	appRequest := model.ApplicationRequest{
		Name:        "TestApp",
		Description: "Test Description",
	}

	// Service returns app with unsupported auth type
	createdApp := &model.ApplicationDTO{
		ID:          "test-app-id",
		Name:        "TestApp",
		Description: "Test Description",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type:           "UNSUPPORTED_TYPE",
				OAuthAppConfig: &model.OAuthAppConfigDTO{ClientID: "test"},
			},
		},
	}

	mockService.On("CreateApplication", mock.AnythingOfType("*model.ApplicationDTO")).Return(createdApp, nil)

	body, _ := json.Marshal(appRequest)
	req := httptest.NewRequest(http.MethodPost, "/applications", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := &failingResponseWriter{failOnce: true}

	handler.HandleApplicationPostRequest(w, req)

	// Should have set status to 500
	assert.Equal(suite.T(), http.StatusInternalServerError, w.statusCode)
	mockService.AssertExpectations(suite.T())
}

func (suite *HandlerTestSuite) TestHandleApplicationPostRequest_SuccessResponseEncodingFails() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	appRequest := model.ApplicationRequest{
		Name:        "TestApp",
		Description: "Test Description",
	}

	expectedApp := &model.ApplicationDTO{
		ID:          "test-app-id",
		Name:        "TestApp",
		Description: "Test Description",
	}

	mockService.On("CreateApplication", mock.AnythingOfType("*model.ApplicationDTO")).Return(expectedApp, nil)

	body, _ := json.Marshal(appRequest)
	req := httptest.NewRequest(http.MethodPost, "/applications", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := &failingResponseWriter{failOnce: true}

	handler.HandleApplicationPostRequest(w, req)

	// Should have set status to 201 before encoding fails
	assert.Equal(suite.T(), http.StatusCreated, w.statusCode)
	mockService.AssertExpectations(suite.T())
}

func (suite *HandlerTestSuite) TestHandleApplicationPutRequest_UnsupportedInboundAuthType() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	appRequest := model.ApplicationRequest{
		Name:        "Updated App",
		Description: "Updated Description",
	}

	// Service returns app with unsupported auth type
	updatedApp := &model.ApplicationDTO{
		ID:          "test-app-id",
		Name:        "Updated App",
		Description: "Updated Description",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: "UNSUPPORTED_TYPE",
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					ClientID: "test-client-id",
				},
			},
		},
	}

	mockService.On("UpdateApplication", "test-app-id", mock.AnythingOfType("*model.ApplicationDTO")).
		Return(updatedApp, nil)

	body, _ := json.Marshal(appRequest)
	req := httptest.NewRequest(http.MethodPut, "/applications/test-app-id", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "test-app-id")
	w := httptest.NewRecorder()

	handler.HandleApplicationPutRequest(w, req)

	// Should return 500 because processInboundAuthConfig returns false
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"))

	var errResp apierror.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), ErrorInternalServerError.Code, errResp.Code)

	mockService.AssertExpectations(suite.T())
}

func (suite *HandlerTestSuite) TestHandleApplicationPutRequest_NilOAuthConfig() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	appRequest := model.ApplicationRequest{
		Name:        "Updated App",
		Description: "Updated Description",
	}

	// Service returns app with OAuth auth type but nil config
	updatedApp := &model.ApplicationDTO{
		ID:          "test-app-id",
		Name:        "Updated App",
		Description: "Updated Description",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type:           model.OAuthInboundAuthType,
				OAuthAppConfig: nil,
			},
		},
	}

	mockService.On("UpdateApplication", "test-app-id", mock.AnythingOfType("*model.ApplicationDTO")).
		Return(updatedApp, nil)

	body, _ := json.Marshal(appRequest)
	req := httptest.NewRequest(http.MethodPut, "/applications/test-app-id", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "test-app-id")
	w := httptest.NewRecorder()

	handler.HandleApplicationPutRequest(w, req)

	// Should return 500 because processInboundAuthConfig returns false
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"))

	var errResp apierror.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), ErrorInternalServerError.Code, errResp.Code)

	mockService.AssertExpectations(suite.T())
}

func (suite *HandlerTestSuite) TestHandleApplicationPutRequest_ProcessInboundAuthConfigErrorEncodingFails() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	appRequest := model.ApplicationRequest{
		Name:        "Updated App",
		Description: "Updated Description",
	}

	updatedApp := &model.ApplicationDTO{
		ID:          "test-app-id",
		Name:        "Updated App",
		Description: "Updated Description",
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type:           "UNSUPPORTED_TYPE",
				OAuthAppConfig: &model.OAuthAppConfigDTO{ClientID: "test"},
			},
		},
	}

	mockService.On("UpdateApplication", "test-app-id", mock.AnythingOfType("*model.ApplicationDTO")).
		Return(updatedApp, nil)

	body, _ := json.Marshal(appRequest)
	req := httptest.NewRequest(http.MethodPut, "/applications/test-app-id", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "test-app-id")
	w := &failingResponseWriter{failOnce: true}

	handler.HandleApplicationPutRequest(w, req)

	// Should have set status to 500
	assert.Equal(suite.T(), http.StatusInternalServerError, w.statusCode)
	mockService.AssertExpectations(suite.T())
}

func (suite *HandlerTestSuite) TestHandleApplicationPutRequest_SuccessResponseEncodingFails() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	appRequest := model.ApplicationRequest{
		Name:        "Updated App",
		Description: "Updated Description",
	}

	updatedApp := &model.ApplicationDTO{
		ID:          "test-app-id",
		Name:        "Updated App",
		Description: "Updated Description",
	}

	mockService.On("UpdateApplication", "test-app-id", mock.AnythingOfType("*model.ApplicationDTO")).
		Return(updatedApp, nil)

	body, _ := json.Marshal(appRequest)
	req := httptest.NewRequest(http.MethodPut, "/applications/test-app-id", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "test-app-id")
	w := &failingResponseWriter{failOnce: true}

	handler.HandleApplicationPutRequest(w, req)

	// Should have set status to 200 before encoding fails
	assert.Equal(suite.T(), http.StatusOK, w.statusCode)
	mockService.AssertExpectations(suite.T())
}

func (suite *HandlerTestSuite) TestHandleApplicationPutRequest_InvalidJSONErrorEncodingFails() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	req := httptest.NewRequest(http.MethodPut, "/applications/test-id", bytes.NewBufferString("{invalid json}"))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "test-id")
	w := &failingResponseWriter{failOnce: true}

	handler.HandleApplicationPutRequest(w, req)

	// Should have set status to 400
	assert.Equal(suite.T(), http.StatusBadRequest, w.statusCode)
}

func (suite *HandlerTestSuite) TestHandleApplicationPutRequest_EmptyIDErrorEncodingFails() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	appRequest := model.ApplicationRequest{
		Name: "Test",
	}

	body, _ := json.Marshal(appRequest)
	req := httptest.NewRequest(http.MethodPut, "/applications/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "")
	w := &failingResponseWriter{failOnce: true}

	handler.HandleApplicationPutRequest(w, req)

	// Should have set status to 400
	assert.Equal(suite.T(), http.StatusBadRequest, w.statusCode)
}

func (suite *HandlerTestSuite) TestHandleApplicationPostRequest_InvalidJSONErrorEncodingFails() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/applications", bytes.NewBufferString("{invalid json}"))
	req.Header.Set("Content-Type", "application/json")
	w := &failingResponseWriter{failOnce: true}

	handler.HandleApplicationPostRequest(w, req)

	// Should have set status to 400 before encoding fails
	assert.Equal(suite.T(), http.StatusBadRequest, w.statusCode)
}

func (suite *HandlerTestSuite) TestHandleApplicationGetRequest_UnsupportedAuthTypeErrorEncodingFails() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	// Return app with unsupported auth type
	app := &model.Application{
		ID:   "test-app-id",
		Name: "TestApp",
		InboundAuthConfig: []model.InboundAuthConfigComplete{
			{
				Type:           "UNSUPPORTED_TYPE", // Not OAuthInboundAuthType
				OAuthAppConfig: &model.OAuthAppConfigComplete{ClientID: "test-client-id"},
			},
		},
	}
	mockService.On("GetApplication", "test-app-id").Return(app, nil)

	req := httptest.NewRequest(http.MethodGet, "/applications/test-app-id", nil)
	req.SetPathValue("id", "test-app-id")
	w := &failingResponseWriter{failOnce: true}

	handler.HandleApplicationGetRequest(w, req)

	// Should have set status to 500 before encoding fails
	assert.Equal(suite.T(), http.StatusInternalServerError, w.statusCode)
	mockService.AssertExpectations(suite.T())
}

func (suite *HandlerTestSuite) TestHandleApplicationGetRequest_NilOAuthConfigErrorEncodingFails() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	// Return app with nil OAuth config
	app := &model.Application{
		ID:   "test-app-id",
		Name: "TestApp",
		InboundAuthConfig: []model.InboundAuthConfigComplete{
			{
				Type:           model.OAuthInboundAuthType,
				OAuthAppConfig: nil, // Nil OAuth config
			},
		},
	}
	mockService.On("GetApplication", "test-app-id").Return(app, nil)

	req := httptest.NewRequest(http.MethodGet, "/applications/test-app-id", nil)
	req.SetPathValue("id", "test-app-id")
	w := &failingResponseWriter{failOnce: true}

	handler.HandleApplicationGetRequest(w, req)

	// Should have set status to 500 before encoding fails
	assert.Equal(suite.T(), http.StatusInternalServerError, w.statusCode)
	mockService.AssertExpectations(suite.T())
}

func (suite *HandlerTestSuite) TestHandleApplicationGetRequest_EmptyResponseTypes() {
	mockService := NewApplicationServiceInterfaceMock(suite.T())
	handler := newApplicationHandler(mockService)

	// Return app with empty response types and grant types
	app := &model.Application{
		ID:   "test-app-id",
		Name: "TestApp",
		InboundAuthConfig: []model.InboundAuthConfigComplete{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigComplete{
					ClientID:      "test-client-id",
					GrantTypes:    []oauth2const.GrantType{},    // Empty grant types
					ResponseTypes: []oauth2const.ResponseType{}, // Empty response types
				},
			},
		},
	}
	mockService.On("GetApplication", "test-app-id").Return(app, nil)

	req := httptest.NewRequest(http.MethodGet, "/applications/test-app-id", nil)
	req.SetPathValue("id", "test-app-id")
	w := httptest.NewRecorder()

	handler.HandleApplicationGetRequest(w, req)

	// Should return 200 OK - handler properly handles empty arrays
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Verify we got a valid JSON response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response)

	mockService.AssertExpectations(suite.T())
}
