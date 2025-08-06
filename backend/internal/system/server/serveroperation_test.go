/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
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

package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/asgardeo/thunder/tests/mocks/database/clientmock"
	"github.com/asgardeo/thunder/tests/mocks/database/providermock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ServerOperationServiceTestSuite struct {
	suite.Suite
	service         ServerOperationServiceInterface
	handlerExecuted bool
}

func TestServerOperationServiceSuite(t *testing.T) {
	suite.Run(t, new(ServerOperationServiceTestSuite))
}

func (suite *ServerOperationServiceTestSuite) SetupTest() {
	svc := NewServerOperationService()
	suite.service = svc
}

func (suite *ServerOperationServiceTestSuite) BeforeTest(suiteName, testName string) {
	suite.handlerExecuted = false
}

func (suite *ServerOperationServiceTestSuite) TestWrapHandleFunction() {
	testCases := []struct {
		name            string
		requestOrigin   string
		allowedOrigins  string
		expectedOrigin  string
		expectedHandled bool
		expectedStatus  int
	}{
		{
			name:            "Valid origin",
			requestOrigin:   "http://example.com",
			allowedOrigins:  "http://example.com,https://wso2.com",
			expectedOrigin:  "http://example.com",
			expectedHandled: true,
			expectedStatus:  http.StatusOK,
		},
		{
			name:            "Invalid origin",
			requestOrigin:   "http://test.com",
			allowedOrigins:  "http://example.com,https://wso2.com",
			expectedOrigin:  "",
			expectedHandled: true,
			expectedStatus:  http.StatusOK,
		},
		{
			name:            "No origin header",
			requestOrigin:   "",
			allowedOrigins:  "http://example.com,https://wso2.com",
			expectedOrigin:  "",
			expectedHandled: true,
			expectedStatus:  http.StatusOK,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Reset the handler executed flag
			suite.handlerExecuted = false

			// Set up mocks
			dbClient := &clientmock.DBClientInterfaceMock{}
			dbClient.On("Close").Return(nil)
			dbClient.On("Query", mock.AnythingOfType("model.DBQuery")).Return([]map[string]interface{}{
				{"allowed_origins": tc.allowedOrigins},
			}, nil)

			dbProvider := &providermock.DBProviderInterfaceMock{}
			dbProvider.On("GetDBClient", "identity").Return(dbClient, nil)
			suite.service.(*ServerOperationService).DBProvider = dbProvider

			mux := http.NewServeMux()
			reqOps := &RequestWrapOptions{
				Cors: &Cors{
					AllowedMethods:   "GET, POST, OPTIONS",
					AllowedHeaders:   "Content-Type, Authorization",
					AllowCredentials: true,
				},
			}

			handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				suite.handlerExecuted = true
				w.WriteHeader(tc.expectedStatus)
			})

			suite.service.WrapHandleFunction(mux, "/test", reqOps, handlerFunc)

			// Serve a request to test the wrapped handler
			rec := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "http://localhost/test", nil)
			if tc.requestOrigin != "" {
				req.Header.Set("Origin", tc.requestOrigin)
			}
			mux.ServeHTTP(rec, req)

			// Assert handler was executed
			assert.Equal(t, tc.expectedHandled, suite.handlerExecuted, "Handler execution state mismatch")

			assert.Equal(t, tc.expectedOrigin, rec.Header().Get("Access-Control-Allow-Origin"),
				"Access-Control-Allow-Origin header mismatch")
			if tc.expectedOrigin != "" {
				assert.Equal(t, reqOps.Cors.AllowedMethods, rec.Header().Get("Access-Control-Allow-Methods"))
				assert.Equal(t, reqOps.Cors.AllowedHeaders, rec.Header().Get("Access-Control-Allow-Headers"))
				assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
			}
		})
	}
}

func (suite *ServerOperationServiceTestSuite) TestWrapHandleFunctionWithError() {
	dbProvider := &providermock.DBProviderInterfaceMock{}
	dbProvider.On("GetDBClient", "identity").Return(nil, errors.New("database connection error"))
	suite.service.(*ServerOperationService).DBProvider = dbProvider

	mux := http.NewServeMux()
	reqOps := &RequestWrapOptions{
		Cors: &Cors{
			AllowedMethods:   "GET, POST, OPTIONS",
			AllowedHeaders:   "Content-Type, Authorization",
			AllowCredentials: true,
		},
	}

	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suite.handlerExecuted = true
		w.WriteHeader(http.StatusOK)
	})

	suite.service.WrapHandleFunction(mux, "/test-error", reqOps, handlerFunc)

	// Serve a request to test error handling
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://localhost/test-error", nil)
	req.Header.Set("Origin", "http://example.com")
	mux.ServeHTTP(rec, req)

	// Even with an error in addAllowedOriginHeaders, the handler should still be executed
	assert.True(suite.T(), suite.handlerExecuted, "Handler should be executed even when addAllowedOriginHeaders fails")
	assert.Equal(suite.T(), http.StatusOK, rec.Code, "HTTP status should be OK")
}

func (suite *ServerOperationServiceTestSuite) TestGetAllowedOriginsDBClientError() {
	dbProvider := &providermock.DBProviderInterfaceMock{}
	dbProvider.On("GetDBClient", "identity").Return(nil, errors.New("database connection error"))
	suite.service.(*ServerOperationService).DBProvider = dbProvider

	// Cast to access the private method
	service := suite.service.(*ServerOperationService)
	origins, err := service.getAllowedOrigins()

	// Verify results
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), origins)
}

func (suite *ServerOperationServiceTestSuite) TestGetAllowedOriginsQueryError() {
	dbClient := &clientmock.DBClientInterfaceMock{}
	dbClient.On("Close").Return(nil)
	dbClient.On("Query", mock.AnythingOfType("model.DBQuery")).Return(nil, errors.New("query execution error"))

	dbProvider := &providermock.DBProviderInterfaceMock{}
	dbProvider.On("GetDBClient", "identity").Return(dbClient, nil)
	suite.service.(*ServerOperationService).DBProvider = dbProvider

	// Cast to access the private method
	service := suite.service.(*ServerOperationService)
	origins, err := service.getAllowedOrigins()

	// Verify results
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), origins)
}

func (suite *ServerOperationServiceTestSuite) TestGetAllowedOriginsEmptyResults() {
	dbClient := &clientmock.DBClientInterfaceMock{}
	dbClient.On("Close").Return(nil)
	dbClient.On("Query", mock.AnythingOfType("model.DBQuery")).Return([]map[string]interface{}{}, nil)

	dbProvider := &providermock.DBProviderInterfaceMock{}
	dbProvider.On("GetDBClient", "identity").Return(dbClient, nil)
	suite.service.(*ServerOperationService).DBProvider = dbProvider

	// Cast to access the private method
	service := suite.service.(*ServerOperationService)
	origins, err := service.getAllowedOrigins()

	// Verify results
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), origins)
}

func (suite *ServerOperationServiceTestSuite) TestNewServerOperationService() {
	svc := NewServerOperationService()
	assert.NotNil(suite.T(), svc)
	assert.IsType(suite.T(), &ServerOperationService{}, svc)

	// Verify that DBProvider is properly initialized
	serverOpSvc, ok := svc.(*ServerOperationService)
	assert.True(suite.T(), ok)
	assert.NotNil(suite.T(), serverOpSvc.DBProvider)
}

func (suite *ServerOperationServiceTestSuite) TestAddAllowedOriginHeaders() {
	testCases := []struct {
		name           string
		requestOrigin  string
		allowedOrigins []string
		corsOptions    *Cors
		dbError        bool
		expectedOrigin string
		expectedError  bool
	}{
		{
			name:           "Valid origin with all CORS options",
			requestOrigin:  "http://example.com",
			allowedOrigins: []string{"http://example.com", "https://wso2.com"},
			corsOptions: &Cors{
				AllowedMethods:   "GET, POST, OPTIONS",
				AllowedHeaders:   "Content-Type, Authorization",
				AllowCredentials: true,
			},
			dbError:        false,
			expectedOrigin: "http://example.com",
			expectedError:  false,
		},
		{
			name:           "Valid origin with minimal CORS options",
			requestOrigin:  "http://example.com",
			allowedOrigins: []string{"http://example.com"},
			corsOptions: &Cors{
				AllowCredentials: false,
			},
			dbError:        false,
			expectedOrigin: "http://example.com",
			expectedError:  false,
		},
		{
			name:           "No request origin",
			requestOrigin:  "",
			allowedOrigins: []string{"http://example.com"},
			corsOptions:    &Cors{},
			dbError:        false,
			expectedOrigin: "",
			expectedError:  false,
		},
		{
			name:           "DB error",
			requestOrigin:  "http://example.com",
			allowedOrigins: nil,
			corsOptions:    &Cors{},
			dbError:        true,
			expectedOrigin: "",
			expectedError:  true,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Set up mocks
			dbClient := &clientmock.DBClientInterfaceMock{}
			dbClient.On("Close").Return(nil)

			dbProvider := &providermock.DBProviderInterfaceMock{}

			service := suite.service.(*ServerOperationService)
			service.DBProvider = dbProvider

			if tc.dbError {
				dbProvider.On("GetDBClient", "identity").Return(nil, errors.New("database connection error"))
			} else {
				dbClient.On("Query", mock.AnythingOfType("model.DBQuery")).Return([]map[string]interface{}{
					{"allowed_origins": "http://example.com,https://wso2.com"},
				}, nil)
				dbProvider.On("GetDBClient", "identity").Return(dbClient, nil)
			}

			// Create request and response for testing
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "http://localhost/test", nil)
			if tc.requestOrigin != "" {
				r.Header.Set("Origin", tc.requestOrigin)
			}

			// Test addAllowedOriginHeaders
			err := service.addAllowedOriginHeaders(w, r, &RequestWrapOptions{Cors: tc.corsOptions})

			// Verify results
			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Check if headers were set correctly
			assert.Equal(t, tc.expectedOrigin, w.Header().Get("Access-Control-Allow-Origin"))

			if tc.expectedOrigin != "" && tc.corsOptions != nil {
				if tc.corsOptions.AllowedMethods != "" {
					assert.Equal(t, tc.corsOptions.AllowedMethods, w.Header().Get("Access-Control-Allow-Methods"))
				}
				if tc.corsOptions.AllowedHeaders != "" {
					assert.Equal(t, tc.corsOptions.AllowedHeaders, w.Header().Get("Access-Control-Allow-Headers"))
				}
				if tc.corsOptions.AllowCredentials {
					assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
				} else {
					assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
				}
			}
		})
	}
}
