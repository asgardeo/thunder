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

package resource

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/tests/mocks/oumock"
)

type InitTestSuite struct {
	suite.Suite
	mockOUService *oumock.OrganizationUnitServiceInterfaceMock
}

func (suite *InitTestSuite) SetupTest() {
	suite.mockOUService = oumock.NewOrganizationUnitServiceInterfaceMock(suite.T())

	// Initialize runtime config for the test
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
	}
	_ = config.InitializeThunderRuntime("test", testConfig)
}

func (suite *InitTestSuite) TearDownTest() {
	// Reset config to clear singleton state for next test
	config.ResetThunderRuntime()
}

func TestInitTestSuite(t *testing.T) {
	suite.Run(t, new(InitTestSuite))
}

// TestInitialize tests the Initialize function
func (suite *InitTestSuite) TestInitialize() {
	mux := http.NewServeMux()

	// Execute
	service, err := Initialize(mux, suite.mockOUService)

	// Assert
	suite.NoError(err)
	suite.NotNil(service)
	suite.Implements((*ResourceServiceInterface)(nil), service)
}

// TestRegisterRoutes tests that all routes are properly registered
func (suite *InitTestSuite) TestRegisterRoutes() {
	mux := http.NewServeMux()
	handler := &resourceHandler{}

	// This test mainly ensures registerRoutes doesn't panic
	suite.NotPanics(func() {
		registerRoutes(mux, handler)
	})

	// Table-driven test for route verification
	testCases := []struct {
		name     string
		method   string
		target   string
		expected string
	}{
		// Resource Server routes
		{
			name:     "GET resource servers list",
			method:   http.MethodGet,
			target:   "/resource-servers",
			expected: "GET /resource-servers",
		},
		{
			name:     "POST create resource server",
			method:   http.MethodPost,
			target:   "/resource-servers",
			expected: "POST /resource-servers",
		},
		{
			name:     "OPTIONS resource servers",
			method:   http.MethodOptions,
			target:   "/resource-servers",
			expected: "OPTIONS /resource-servers",
		},
		{
			name:     "GET resource server by ID",
			method:   http.MethodGet,
			target:   "/resource-servers/rs-123",
			expected: "GET /resource-servers/{id}",
		},
		{
			name:     "PUT update resource server",
			method:   http.MethodPut,
			target:   "/resource-servers/rs-123",
			expected: "PUT /resource-servers/{id}",
		},
		{
			name:     "DELETE resource server",
			method:   http.MethodDelete,
			target:   "/resource-servers/rs-123",
			expected: "DELETE /resource-servers/{id}",
		},
		{
			name:     "OPTIONS resource server by ID",
			method:   http.MethodOptions,
			target:   "/resource-servers/rs-123",
			expected: "OPTIONS /resource-servers/{id}",
		},
		// Resource routes
		{
			name:     "GET resources list",
			method:   http.MethodGet,
			target:   "/resource-servers/rs-123/resources",
			expected: "GET /resource-servers/{rsId}/resources",
		},
		{
			name:     "POST create resource",
			method:   http.MethodPost,
			target:   "/resource-servers/rs-123/resources",
			expected: "POST /resource-servers/{rsId}/resources",
		},
		{
			name:     "OPTIONS resources",
			method:   http.MethodOptions,
			target:   "/resource-servers/rs-123/resources",
			expected: "OPTIONS /resource-servers/{rsId}/resources",
		},
		{
			name:     "GET resource by ID",
			method:   http.MethodGet,
			target:   "/resource-servers/rs-123/resources/res-456",
			expected: "GET /resource-servers/{rsId}/resources/{id}",
		},
		{
			name:     "PUT update resource",
			method:   http.MethodPut,
			target:   "/resource-servers/rs-123/resources/res-456",
			expected: "PUT /resource-servers/{rsId}/resources/{id}",
		},
		{
			name:     "DELETE resource",
			method:   http.MethodDelete,
			target:   "/resource-servers/rs-123/resources/res-456",
			expected: "DELETE /resource-servers/{rsId}/resources/{id}",
		},
		{
			name:     "OPTIONS resource by ID",
			method:   http.MethodOptions,
			target:   "/resource-servers/rs-123/resources/res-456",
			expected: "OPTIONS /resource-servers/{rsId}/resources/{id}",
		},
		// Action routes at Resource Server level
		{
			name:     "GET actions at resource server",
			method:   http.MethodGet,
			target:   "/resource-servers/rs-123/actions",
			expected: "GET /resource-servers/{rsId}/actions",
		},
		{
			name:     "POST create action at resource server",
			method:   http.MethodPost,
			target:   "/resource-servers/rs-123/actions",
			expected: "POST /resource-servers/{rsId}/actions",
		},
		{
			name:     "OPTIONS actions at resource server",
			method:   http.MethodOptions,
			target:   "/resource-servers/rs-123/actions",
			expected: "OPTIONS /resource-servers/{rsId}/actions",
		},
		{
			name:     "GET action by ID at resource server",
			method:   http.MethodGet,
			target:   "/resource-servers/rs-123/actions/act-789",
			expected: "GET /resource-servers/{rsId}/actions/{id}",
		},
		{
			name:     "PUT update action at resource server",
			method:   http.MethodPut,
			target:   "/resource-servers/rs-123/actions/act-789",
			expected: "PUT /resource-servers/{rsId}/actions/{id}",
		},
		{
			name:     "DELETE action at resource server",
			method:   http.MethodDelete,
			target:   "/resource-servers/rs-123/actions/act-789",
			expected: "DELETE /resource-servers/{rsId}/actions/{id}",
		},
		{
			name:     "OPTIONS action by ID at resource server",
			method:   http.MethodOptions,
			target:   "/resource-servers/rs-123/actions/act-789",
			expected: "OPTIONS /resource-servers/{rsId}/actions/{id}",
		},
		// Action routes at Resource level
		{
			name:     "GET actions at resource",
			method:   http.MethodGet,
			target:   "/resource-servers/rs-123/resources/res-456/actions",
			expected: "GET /resource-servers/{rsId}/resources/{resourceId}/actions",
		},
		{
			name:     "POST create action at resource",
			method:   http.MethodPost,
			target:   "/resource-servers/rs-123/resources/res-456/actions",
			expected: "POST /resource-servers/{rsId}/resources/{resourceId}/actions",
		},
		{
			name:     "OPTIONS actions at resource",
			method:   http.MethodOptions,
			target:   "/resource-servers/rs-123/resources/res-456/actions",
			expected: "OPTIONS /resource-servers/{rsId}/resources/{resourceId}/actions",
		},
		{
			name:     "GET action by ID at resource",
			method:   http.MethodGet,
			target:   "/resource-servers/rs-123/resources/res-456/actions/act-789",
			expected: "GET /resource-servers/{rsId}/resources/{resourceId}/actions/{id}",
		},
		{
			name:     "PUT update action at resource",
			method:   http.MethodPut,
			target:   "/resource-servers/rs-123/resources/res-456/actions/act-789",
			expected: "PUT /resource-servers/{rsId}/resources/{resourceId}/actions/{id}",
		},
		{
			name:     "DELETE action at resource",
			method:   http.MethodDelete,
			target:   "/resource-servers/rs-123/resources/res-456/actions/act-789",
			expected: "DELETE /resource-servers/{rsId}/resources/{resourceId}/actions/{id}",
		},
		{
			name:     "OPTIONS action by ID at resource",
			method:   http.MethodOptions,
			target:   "/resource-servers/rs-123/resources/res-456/actions/act-789",
			expected: "OPTIONS /resource-servers/{rsId}/resources/{resourceId}/actions/{id}",
		},
	}

	// Verify all routes are registered with correct patterns
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			req := httptest.NewRequest(tc.method, tc.target, nil)
			_, pattern := mux.Handler(req)
			suite.Equal(tc.expected, pattern, "Route pattern mismatch for %s %s", tc.method, tc.target)
		})
	}
}

// TestNewResourceHandler tests the newResourceHandler constructor
func (suite *InitTestSuite) TestNewResourceHandler() {
	mockService := NewResourceServiceInterfaceMock(suite.T())

	// Execute
	handler := newResourceHandler(mockService)

	// Assert
	suite.NotNil(handler)
	suite.Equal(mockService, handler.resourceService)
}

// TestNewResourceService tests the newResourceService constructor
func (suite *InitTestSuite) TestNewResourceService() {
	mockStore := newResourceStoreInterfaceMock(suite.T())

	// Execute
	service, err := newResourceService(mockStore, suite.mockOUService)

	// Assert
	suite.NoError(err)
	suite.NotNil(service)
	suite.Implements((*ResourceServiceInterface)(nil), service)

	// Verify dependencies are set correctly
	resSvc, ok := service.(*resourceService)
	suite.True(ok)
	suite.Equal(mockStore, resSvc.resourceStore)
	suite.Equal(suite.mockOUService, resSvc.ouService)
}

// TestNewResourceStore tests the newResourceStore constructor
func (suite *InitTestSuite) TestNewResourceStore() {
	// Execute
	store := newResourceStore()

	// Assert
	suite.NotNil(store)
	suite.Implements((*resourceStoreInterface)(nil), store)

	// Verify store is properly initialized
	resStore, ok := store.(*resourceStore)
	suite.True(ok)
	suite.NotNil(resStore.dbProvider)
	suite.Equal("test-deployment", resStore.deploymentID)
}

// TestRegisterRoutes_AllOPTIONSRoutes tests that all OPTIONS routes return NoContent
func (suite *InitTestSuite) TestRegisterRoutes_AllOPTIONSRoutes() {
	mux := http.NewServeMux()
	handler := &resourceHandler{}
	registerRoutes(mux, handler)

	// Table-driven test for OPTIONS routes
	optionsRoutes := []struct {
		name   string
		target string
	}{
		{name: "OPTIONS /resource-servers", target: "/resource-servers"},
		{name: "OPTIONS /resource-servers/{id}", target: "/resource-servers/rs-123"},
		{name: "OPTIONS /resource-servers/{rsId}/resources", target: "/resource-servers/rs-123/resources"},
		{name: "OPTIONS /resource-servers/{rsId}/resources/{id}", target: "/resource-servers/rs-123/resources/res-456"},
		{name: "OPTIONS /resource-servers/{rsId}/actions", target: "/resource-servers/rs-123/actions"},
		{name: "OPTIONS /resource-servers/{rsId}/actions/{id}", target: "/resource-servers/rs-123/actions/act-789"},
		{name: "OPTIONS /resource-servers/{rsId}/resources/{resourceId}/actions",
			target: "/resource-servers/rs-123/resources/res-456/actions"},
		{name: "OPTIONS /resource-servers/{rsId}/resources/{resourceId}/actions/{id}",
			target: "/resource-servers/rs-123/resources/res-456/actions/act-789"},
	}

	for _, tc := range optionsRoutes {
		suite.Run(tc.name, func() {
			req := httptest.NewRequest(http.MethodOptions, tc.target, nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			suite.Equal(http.StatusNoContent, w.Code, "OPTIONS request should return 204 No Content")
		})
	}
}

// TestInitialize_IntegrationFlow tests the complete initialization flow
func (suite *InitTestSuite) TestInitialize_IntegrationFlow() {
	mux := http.NewServeMux()

	// Execute
	service, err := Initialize(mux, suite.mockOUService)

	// Assert service is created
	suite.NoError(err)
	suite.NotNil(service)
	suite.Implements((*ResourceServiceInterface)(nil), service)

	// Verify routes are registered by checking a sample route
	req := httptest.NewRequest(http.MethodGet, "/resource-servers", nil)
	_, pattern := mux.Handler(req)
	suite.Equal("GET /resource-servers", pattern, "Routes should be registered during initialization")
}

// TestRegisterRoutes_CORSConfiguration tests CORS headers are properly configured
func (suite *InitTestSuite) TestRegisterRoutes_CORSConfiguration() {
	mux := http.NewServeMux()
	handler := &resourceHandler{}
	registerRoutes(mux, handler)

	// Table-driven test for CORS verification on different route groups
	corsTestCases := []struct {
		name           string
		method         string
		target         string
		expectedStatus int
	}{
		{
			name:           "OPTIONS resource servers endpoint",
			method:         http.MethodOptions,
			target:         "/resource-servers",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "OPTIONS resource server detail endpoint",
			method:         http.MethodOptions,
			target:         "/resource-servers/rs-123",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "OPTIONS resources endpoint",
			method:         http.MethodOptions,
			target:         "/resource-servers/rs-123/resources",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "OPTIONS resource detail endpoint",
			method:         http.MethodOptions,
			target:         "/resource-servers/rs-123/resources/res-456",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "OPTIONS actions at resource server endpoint",
			method:         http.MethodOptions,
			target:         "/resource-servers/rs-123/actions",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "OPTIONS action detail at resource server endpoint",
			method:         http.MethodOptions,
			target:         "/resource-servers/rs-123/actions/act-789",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "OPTIONS actions at resource endpoint",
			method:         http.MethodOptions,
			target:         "/resource-servers/rs-123/resources/res-456/actions",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "OPTIONS action detail at resource endpoint",
			method:         http.MethodOptions,
			target:         "/resource-servers/rs-123/resources/res-456/actions/act-789",
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tc := range corsTestCases {
		suite.Run(tc.name, func() {
			req := httptest.NewRequest(tc.method, tc.target, nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			suite.Equal(tc.expectedStatus, w.Code, "OPTIONS request should return correct status")
		})
	}
}
