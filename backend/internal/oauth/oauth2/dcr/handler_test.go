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
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
)

// mockDCRService is a mock implementation of DCRServiceInterface
type mockDCRService struct {
	mock.Mock
}

func (m *mockDCRService) RegisterClient(
	request *DCRRegistrationRequest,
) (*DCRRegistrationResponse, *serviceerror.ServiceError) {
	args := m.Called(request)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*serviceerror.ServiceError)
	}
	return args.Get(0).(*DCRRegistrationResponse), args.Get(1).(*serviceerror.ServiceError)
}

// DCRHandlerTestSuite is the test suite for DCR handler
type DCRHandlerTestSuite struct {
	suite.Suite
	mockService *mockDCRService
	handler     *dcrHandler
}

func TestDCRHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(DCRHandlerTestSuite))
}

func (s *DCRHandlerTestSuite) SetupTest() {
	s.mockService = new(mockDCRService)
	s.handler = newDCRHandler(s.mockService)
}

// TestHandleDCRRegistration_InvalidRequestFormat tests handling of invalid JSON in request body
func (s *DCRHandlerTestSuite) TestHandleDCRRegistration_InvalidRequestFormat() {
	// Create a request with invalid JSON
	invalidJSON := `{"invalid": json}`
	req := httptest.NewRequest(http.MethodPost, "/oauth2/dcr", bytes.NewReader([]byte(invalidJSON)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handler.HandleDCRRegistration(rr, req)

	assert.Equal(s.T(), http.StatusBadRequest, rr.Code)
	var errorResponse map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
	s.NoError(err)
	assert.Contains(s.T(), errorResponse, "error")
}

// TestHandleDCRRegistration_ServiceError tests handling of service errors
func (s *DCRHandlerTestSuite) TestHandleDCRRegistration_ServiceError() {
	request := &DCRRegistrationRequest{
		RedirectURIs: []string{"https://client.example.com/callback"},
		GrantTypes:   []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
	}

	serviceErr := &ErrorInvalidRedirectURI
	s.mockService.On("RegisterClient", request).Return(nil, serviceErr)

	requestJSON, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/oauth2/dcr", bytes.NewReader(requestJSON))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handler.HandleDCRRegistration(rr, req)

	assert.Equal(s.T(), http.StatusBadRequest, rr.Code)
	s.mockService.AssertExpectations(s.T())
}

// TestHandleDCRRegistration_ClientError tests handling of client errors
func (s *DCRHandlerTestSuite) TestHandleDCRRegistration_ClientError() {
	request := &DCRRegistrationRequest{
		RedirectURIs: []string{"not-a-valid-uri"},
		GrantTypes:   []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
	}

	serviceErr := &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "invalid_client_metadata",
		Error:            "Invalid client metadata",
		ErrorDescription: "Invalid grant type",
	}
	s.mockService.On("RegisterClient", request).Return(nil, serviceErr)

	requestJSON, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/oauth2/dcr", bytes.NewReader(requestJSON))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handler.HandleDCRRegistration(rr, req)

	assert.Equal(s.T(), http.StatusBadRequest, rr.Code)
	var errorResponse map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
	s.NoError(err)
	assert.Equal(s.T(), "invalid_client_metadata", errorResponse["error"])
	s.mockService.AssertExpectations(s.T())
}

// TestHandleDCRRegistration_ServerError tests handling of server errors
func (s *DCRHandlerTestSuite) TestHandleDCRRegistration_ServerError() {
	request := &DCRRegistrationRequest{
		RedirectURIs: []string{"https://client.example.com/callback"},
		GrantTypes:   []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
	}

	serviceErr := &ErrorServerError
	s.mockService.On("RegisterClient", request).Return(nil, serviceErr)

	requestJSON, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/oauth2/dcr", bytes.NewReader(requestJSON))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handler.HandleDCRRegistration(rr, req)

	assert.Equal(s.T(), http.StatusInternalServerError, rr.Code)
	var errorResponse map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
	s.NoError(err)
	assert.Equal(s.T(), "server_error", errorResponse["error"])
	s.mockService.AssertExpectations(s.T())
}

// TestHandleDCRRegistration_UnknownErrorType tests handling of unknown error types (defaults to BadRequest)
func (s *DCRHandlerTestSuite) TestHandleDCRRegistration_UnknownErrorType() {
	request := &DCRRegistrationRequest{
		RedirectURIs: []string{"https://client.example.com/callback"},
		GrantTypes:   []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
	}

	serviceErr := &serviceerror.ServiceError{
		Type:             "UnknownErrorType",
		Code:             "unknown_error",
		Error:            "Unknown error",
		ErrorDescription: "An unknown error occurred",
	}
	s.mockService.On("RegisterClient", request).Return(nil, serviceErr)

	requestJSON, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/oauth2/dcr", bytes.NewReader(requestJSON))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handler.HandleDCRRegistration(rr, req)

	// Unknown error type should default to BadRequest
	assert.Equal(s.T(), http.StatusBadRequest, rr.Code)
	s.mockService.AssertExpectations(s.T())
}

// TestHandleDCRRegistration_Success tests successful registration
func (s *DCRHandlerTestSuite) TestHandleDCRRegistration_Success() {
	request := &DCRRegistrationRequest{
		RedirectURIs: []string{"https://client.example.com/callback"},
		GrantTypes:   []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
		ClientName:   "Test Client",
	}

	response := &DCRRegistrationResponse{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		ClientName:   "Test Client",
		RedirectURIs: []string{"https://client.example.com/callback"},
		GrantTypes:   []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
	}

	s.mockService.On("RegisterClient", request).Return(response, (*serviceerror.ServiceError)(nil))

	requestJSON, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/oauth2/dcr", bytes.NewReader(requestJSON))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handler.HandleDCRRegistration(rr, req)

	assert.Equal(s.T(), http.StatusCreated, rr.Code)
	var responseBody DCRRegistrationResponse
	err := json.Unmarshal(rr.Body.Bytes(), &responseBody)
	s.NoError(err)
	assert.Equal(s.T(), "test-client-id", responseBody.ClientID)
	assert.Equal(s.T(), "test-client-secret", responseBody.ClientSecret)
	assert.Equal(s.T(), "Test Client", responseBody.ClientName)
	s.mockService.AssertExpectations(s.T())
}

// TestHandleDCRRegistration_EmptyBody tests handling of empty request body
func (s *DCRHandlerTestSuite) TestHandleDCRRegistration_EmptyBody() {
	req := httptest.NewRequest(http.MethodPost, "/oauth2/dcr", bytes.NewReader([]byte("")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handler.HandleDCRRegistration(rr, req)

	assert.Equal(s.T(), http.StatusBadRequest, rr.Code)
	var errorResponse map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
	s.NoError(err)
	assert.Contains(s.T(), errorResponse, "error")
}

// TestNewDCRHandler tests the handler constructor
func TestNewDCRHandler(t *testing.T) {
	mockService := new(mockDCRService)
	handler := newDCRHandler(mockService)

	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.dcrService)
}

// TestWriteServiceErrorResponse_DirectCall tests the writeServiceErrorResponse function directly
func TestWriteServiceErrorResponse_DirectCall(t *testing.T) {
	mockService := new(mockDCRService)
	handler := newDCRHandler(mockService)

	testCases := []struct {
		name           string
		serviceError   *serviceerror.ServiceError
		expectedStatus int
	}{
		{
			name: "Client Error",
			serviceError: &serviceerror.ServiceError{
				Type:             serviceerror.ClientErrorType,
				Code:             "test_code",
				Error:            "Test error",
				ErrorDescription: "Test description",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Server Error",
			serviceError: &serviceerror.ServiceError{
				Type:             serviceerror.ServerErrorType,
				Code:             "test_code",
				Error:            "Test error",
				ErrorDescription: "Test description",
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Unknown Error Type",
			serviceError: &serviceerror.ServiceError{
				Type:             "UnknownType",
				Code:             "test_code",
				Error:            "Test error",
				ErrorDescription: "Test description",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			handler.writeServiceErrorResponse(rr, tc.serviceError)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			var errorResponse map[string]interface{}
			err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
			assert.NoError(t, err)
			assert.Equal(t, tc.serviceError.Code, errorResponse["error"])
		})
	}
}
