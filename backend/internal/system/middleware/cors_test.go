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

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/config"
)

type CORSMiddlewareTestSuite struct {
	suite.Suite
}

func TestCORSMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(CORSMiddlewareTestSuite))
}

func (suite *CORSMiddlewareTestSuite) SetupTest() {
	// Initialize runtime config
	cfg := &config.Config{
		CORS: config.CORSConfig{
			AllowedOrigins: []string{"https://example.com", "https://test.com"},
		},
	}
	_ = config.InitializeThunderRuntime("/tmp", cfg)
}

func (suite *CORSMiddlewareTestSuite) TestWithCORS_ValidOrigin() {
	opts := CORSOptions{
		AllowedMethods:   "GET, POST",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}

	pattern, wrappedHandler := WithCORS("GET /test", handler, opts)

	assert.Equal(suite.T(), "GET /test", pattern)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	wrappedHandler(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Equal(suite.T(), "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(suite.T(), "GET, POST", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(suite.T(), "Content-Type, Authorization", w.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(suite.T(), "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(suite.T(), "OK", w.Body.String())
}

func (suite *CORSMiddlewareTestSuite) TestWithCORS_InvalidOrigin() {
	opts := CORSOptions{
		AllowedMethods:   "GET, POST",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}

	_, wrappedHandler := WithCORS("GET /test", handler, opts)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://malicious.com")
	w := httptest.NewRecorder()

	wrappedHandler(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Empty(suite.T(), w.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(suite.T(), w.Header().Get("Access-Control-Allow-Methods"))
	assert.Empty(suite.T(), w.Header().Get("Access-Control-Allow-Headers"))
	assert.Empty(suite.T(), w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(suite.T(), "OK", w.Body.String())
}

func (suite *CORSMiddlewareTestSuite) TestWithCORS_NoOriginHeader() {
	opts := CORSOptions{
		AllowedMethods:   "GET, POST",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}

	_, wrappedHandler := WithCORS("GET /test", handler, opts)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	wrappedHandler(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Empty(suite.T(), w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(suite.T(), "OK", w.Body.String())
}

func (suite *CORSMiddlewareTestSuite) TestWithCORS_WithoutCredentials() {
	opts := CORSOptions{
		AllowedMethods:   "GET",
		AllowedHeaders:   "Content-Type",
		AllowCredentials: false,
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	_, wrappedHandler := WithCORS("GET /test", handler, opts)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	wrappedHandler(w, req)

	assert.Equal(suite.T(), "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(suite.T(), "GET", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(suite.T(), "Content-Type", w.Header().Get("Access-Control-Allow-Headers"))
	assert.Empty(suite.T(), w.Header().Get("Access-Control-Allow-Credentials"))
}

func (suite *CORSMiddlewareTestSuite) TestGetAllowedOrigins() {
	// Call should return configured origins
	origins := getAllowedOrigins()
	assert.Len(suite.T(), origins, 2)
	assert.Contains(suite.T(), origins, "https://example.com")
	assert.Contains(suite.T(), origins, "https://test.com")
}

func (suite *CORSMiddlewareTestSuite) TestApplyCORSHeaders_EmptyOptions() {
	opts := CORSOptions{}

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	_, wrappedHandler := WithCORS("GET /test", handler, opts)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	wrappedHandler(w, req)

	assert.Equal(suite.T(), "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(suite.T(), w.Header().Get("Access-Control-Allow-Methods"))
	assert.Empty(suite.T(), w.Header().Get("Access-Control-Allow-Headers"))
	assert.Empty(suite.T(), w.Header().Get("Access-Control-Allow-Credentials"))
}

func (suite *CORSMiddlewareTestSuite) TestWithCORS_MultipleAllowedOrigins() {
	opts := CORSOptions{
		AllowedMethods:   "GET",
		AllowedHeaders:   "Content-Type",
		AllowCredentials: true,
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	_, wrappedHandler := WithCORS("GET /test", handler, opts)

	// Test first allowed origin
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.Header.Set("Origin", "https://example.com")
	w1 := httptest.NewRecorder()
	wrappedHandler(w1, req1)
	assert.Equal(suite.T(), "https://example.com", w1.Header().Get("Access-Control-Allow-Origin"))

	// Test second allowed origin
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("Origin", "https://test.com")
	w2 := httptest.NewRecorder()
	wrappedHandler(w2, req2)
	assert.Equal(suite.T(), "https://test.com", w2.Header().Get("Access-Control-Allow-Origin"))
}
