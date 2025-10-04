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

package token

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/config"
)

type TokenHandlerTestSuite struct {
	suite.Suite
}

func TestTokenHandlerSuite(t *testing.T) {
	suite.Run(t, new(TokenHandlerTestSuite))
}

func (suite *TokenHandlerTestSuite) SetupTest() {
	// Initialize Thunder Runtime config with basic test config
	testConfig := &config.Config{
		OAuth: config.OAuthConfig{
			JWT: config.JWTConfig{
				ValidityPeriod: 3600,
			},
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)
}

func (suite *TokenHandlerTestSuite) TestNewTokenHandler() {
	handler := NewTokenHandler()
	assert.NotNil(suite.T(), handler)
	assert.Implements(suite.T(), (*TokenHandlerInterface)(nil), handler)
}

func (suite *TokenHandlerTestSuite) TestHandleTokenRequest_InvalidFormData() {
	handler := NewTokenHandler()
	req, _ := http.NewRequest("POST", "/token", strings.NewReader("invalid-form-data%"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	handler.HandleTokenRequest(rr, req)

	assert.Equal(suite.T(), http.StatusBadRequest, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "invalid_request", response["error"])
	assert.Equal(suite.T(), "Failed to parse request body", response["error_description"])
}

func (suite *TokenHandlerTestSuite) TestHandleTokenRequest_MissingGrantType() {
	handler := NewTokenHandler()
	formData := url.Values{}
	formData.Set("client_id", "test-client-id")
	formData.Set("client_secret", "test-secret")

	req, _ := http.NewRequest("POST", "/token", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	handler.HandleTokenRequest(rr, req)

	assert.Equal(suite.T(), http.StatusBadRequest, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "invalid_request", response["error"])
	assert.Equal(suite.T(), "Missing grant_type parameter", response["error_description"])
}

// Helper function to test token request error scenarios
func (suite *TokenHandlerTestSuite) testTokenRequestError(formData url.Values,
	expectedStatusCode int, expectedError, expectedErrorDescription string) {
	handler := NewTokenHandler()

	req, _ := http.NewRequest("POST", "/token", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	handler.HandleTokenRequest(rr, req)

	assert.Equal(suite.T(), expectedStatusCode, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedError, response["error"])
	assert.Equal(suite.T(), expectedErrorDescription, response["error_description"])
}

func (suite *TokenHandlerTestSuite) TestHandleTokenRequest_InvalidGrantType() {
	formData := url.Values{}
	formData.Set("grant_type", "invalid_grant")
	formData.Set("client_id", "test-client-id")
	formData.Set("client_secret", "test-secret")

	suite.testTokenRequestError(formData, http.StatusBadRequest, "unsupported_grant_type",
		"Invalid grant_type parameter")
}

func (suite *TokenHandlerTestSuite) TestHandleTokenRequest_MissingClientID() {
	handler := NewTokenHandler()
	formData := url.Values{}
	formData.Set("grant_type", "authorization_code")
	formData.Set("client_secret", "test-secret")

	req, _ := http.NewRequest("POST", "/token", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	handler.HandleTokenRequest(rr, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "invalid_client", response["error"])
	assert.Equal(suite.T(), "Missing client_id parameter", response["error_description"])
}

func (suite *TokenHandlerTestSuite) TestHandleTokenRequest_MissingClientSecret() {
	handler := NewTokenHandler()
	formData := url.Values{}
	formData.Set("grant_type", "authorization_code")
	formData.Set("client_id", "test-client-id")

	req, _ := http.NewRequest("POST", "/token", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	handler.HandleTokenRequest(rr, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "invalid_client", response["error"])
	assert.Equal(suite.T(), "Invalid client credentials", response["error_description"])
}

func (suite *TokenHandlerTestSuite) TestHandleTokenRequest_InvalidClient() {
	formData := url.Values{}
	formData.Set("grant_type", "authorization_code")
	formData.Set("client_id", "invalid-client")
	formData.Set("client_secret", "test-secret")

	suite.testTokenRequestError(formData, http.StatusUnauthorized, "invalid_client", "Invalid client credentials")
}

// Helper functions tested independently
func (suite *TokenHandlerTestSuite) TestExtractClientIDAndSecret_Success() {
	formData := url.Values{}
	formData.Set("client_id", "test-client")
	formData.Set("client_secret", "test-secret")

	req, _ := http.NewRequest("POST", "/token", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()

	rr := httptest.NewRecorder()

	clientID, clientSecret, authMethod, ok := extractClientIDAndSecret(req, rr)

	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "test-client", clientID)
	assert.Equal(suite.T(), "test-secret", clientSecret)
	assert.Equal(suite.T(), "client_secret_post", string(authMethod))
}

func (suite *TokenHandlerTestSuite) TestExtractClientIDAndSecret_NoClientID() {
	req, _ := http.NewRequest("POST", "/token", nil)
	_ = req.ParseForm()

	rr := httptest.NewRecorder()

	clientID, clientSecret, _, ok := extractClientIDAndSecret(req, rr)

	assert.False(suite.T(), ok)
	assert.Equal(suite.T(), "", clientID)
	assert.Equal(suite.T(), "", clientSecret)
	// Don't test auth method when extraction fails
}
