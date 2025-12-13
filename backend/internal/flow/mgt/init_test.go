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

package flowmgt

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/config"
)

const (
	testFlowIDInit = "test-flow-id"
)

type InitTestSuite struct {
	suite.Suite
	mockService *FlowMgtServiceInterfaceMock
}

func (s *InitTestSuite) SetupTest() {
	s.mockService = NewFlowMgtServiceInterfaceMock(s.T())

	testConfig := &config.Config{
		Database: config.DatabaseConfig{
			Identity: config.DataSource{
				Type: "sqlite",
				Path: ":memory:",
			},
			Runtime: config.DataSource{
				Type: "sqlite",
				Path: ":memory:",
			},
		},
		Server: config.ServerConfig{
			Identifier: "test-deployment",
		},
		CORS: config.CORSConfig{
			AllowedOrigins: []string{"https://example.com", "https://localhost:3000"},
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)
}

func (s *InitTestSuite) TearDownTest() {
	config.ResetThunderRuntime()
}

func TestInitTestSuite(t *testing.T) {
	suite.Run(t, new(InitTestSuite))
}

func (s *InitTestSuite) TestRegisterRoutes_AllRoutesRegistered() {
	mux := http.NewServeMux()
	handler := newFlowMgtHandler(s.mockService)
	registerRoutes(mux, handler)

	// Test OPTIONS endpoints which don't require service calls
	testCases := []struct {
		name string
		path string
	}{
		{"OPTIONS /flows", "/flows"},
		{"OPTIONS /flows/{flowId}", "/flows/test-id"},
		{"OPTIONS /flows/{flowId}/versions", "/flows/test-id/versions"},
		{"OPTIONS /flows/{flowId}/versions/{version}", "/flows/test-id/versions/1"},
		{"OPTIONS /flows/{flowId}/restore", "/flows/test-id/restore"},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			req := httptest.NewRequest(http.MethodOptions, tc.path, nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)
			s.Equal(http.StatusNoContent, w.Code, "Route %s should be registered", tc.path)
		})
	}
}

func (s *InitTestSuite) TestRegisterRoutes_CORSHeadersConfigured() {
	mux := http.NewServeMux()
	handler := newFlowMgtHandler(s.mockService)

	registerRoutes(mux, handler)

	testCases := []struct {
		name                   string
		method                 string
		path                   string
		expectedAllowedMethods string
	}{
		{
			name:                   "CORS for /flows",
			method:                 http.MethodOptions,
			path:                   "/flows",
			expectedAllowedMethods: "GET, POST",
		},
		{
			name:                   "CORS for /flows/{flowId}",
			method:                 http.MethodOptions,
			path:                   "/flows/" + testFlowIDInit,
			expectedAllowedMethods: "GET, PUT, DELETE",
		},
		{
			name:                   "CORS for /flows/{flowId}/versions",
			method:                 http.MethodOptions,
			path:                   "/flows/" + testFlowIDInit + "/versions",
			expectedAllowedMethods: "GET",
		},
		{
			name:                   "CORS for /flows/{flowId}/versions/{version}",
			method:                 http.MethodOptions,
			path:                   "/flows/" + testFlowIDInit + "/versions/1",
			expectedAllowedMethods: "GET",
		},
		{
			name:                   "CORS for /flows/{flowId}/restore",
			method:                 http.MethodOptions,
			path:                   "/flows/" + testFlowIDInit + "/restore",
			expectedAllowedMethods: "POST",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			req.Header.Set("Origin", "https://example.com")
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			s.Equal(http.StatusNoContent, w.Code)
			s.Contains(w.Header().Get("Access-Control-Allow-Methods"), tc.expectedAllowedMethods)
			s.Contains(w.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
			s.Contains(w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
			s.Equal("true", w.Header().Get("Access-Control-Allow-Credentials"))
		})
	}
}

func (s *InitTestSuite) TestRegisterRoutes_OPTIONSHandlers() {
	mux := http.NewServeMux()
	handler := newFlowMgtHandler(s.mockService)

	registerRoutes(mux, handler)

	optionsPaths := []string{
		"/flows",
		"/flows/" + testFlowIDInit,
		"/flows/" + testFlowIDInit + "/versions",
		"/flows/" + testFlowIDInit + "/versions/1",
		"/flows/" + testFlowIDInit + "/restore",
	}

	for _, path := range optionsPaths {
		s.Run("OPTIONS "+path, func() {
			req := httptest.NewRequest(http.MethodOptions, path, nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			s.Equal(http.StatusNoContent, w.Code, "OPTIONS request should return 204")
			s.Empty(w.Body.String(), "OPTIONS response should have empty body")
		})
	}
}

func (s *InitTestSuite) TestRegisterRoutes_WithNilHandler() {
	mux := http.NewServeMux()

	// Routes can be registered with nil handler, but calling them would fail
	s.NotPanics(func() {
		registerRoutes(mux, nil)
	}, "Should not panic when handler is nil during registration")
}

func (s *InitTestSuite) TestRegisterRoutes_PreflightRequests() {
	mux := http.NewServeMux()
	handler := newFlowMgtHandler(s.mockService)

	registerRoutes(mux, handler)

	testCases := []struct {
		name           string
		path           string
		origin         string
		requestMethod  string
		requestHeaders string
	}{
		{
			name:           "Preflight for POST /flows",
			path:           "/flows",
			origin:         "https://example.com",
			requestMethod:  "POST",
			requestHeaders: "Content-Type",
		},
		{
			name:           "Preflight for PUT /flows/{flowId}",
			path:           "/flows/" + testFlowIDInit,
			origin:         "https://example.com",
			requestMethod:  "PUT",
			requestHeaders: "Authorization",
		},
		{
			name:           "Preflight for DELETE /flows/{flowId}",
			path:           "/flows/" + testFlowIDInit,
			origin:         "https://example.com",
			requestMethod:  "DELETE",
			requestHeaders: "Content-Type, Authorization",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			req := httptest.NewRequest(http.MethodOptions, tc.path, nil)
			req.Header.Set("Origin", tc.origin)
			req.Header.Set("Access-Control-Request-Method", tc.requestMethod)
			req.Header.Set("Access-Control-Request-Headers", tc.requestHeaders)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			s.Equal(http.StatusNoContent, w.Code)
			s.NotEmpty(w.Header().Get("Access-Control-Allow-Origin"))
			s.NotEmpty(w.Header().Get("Access-Control-Allow-Methods"))
			s.NotEmpty(w.Header().Get("Access-Control-Allow-Headers"))
		})
	}
}
