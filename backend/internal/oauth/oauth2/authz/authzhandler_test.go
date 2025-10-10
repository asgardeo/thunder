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

package authz

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	_ "modernc.org/sqlite"

	"github.com/asgardeo/thunder/internal/oauth/oauth2/authz/constants"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/authz/model"
	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	oauth2model "github.com/asgardeo/thunder/internal/oauth/oauth2/model"
	sessionmodel "github.com/asgardeo/thunder/internal/oauth/session/model"
	"github.com/asgardeo/thunder/internal/system/config"
)

type AuthorizeHandlerTestSuite struct {
	suite.Suite
	handler *AuthorizeHandler
	tempDir string
}

func TestAuthorizeHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(AuthorizeHandlerTestSuite))
}

func (suite *AuthorizeHandlerTestSuite) SetupTest() {
	// Create a temporary directory for the test database.
	tempDir, err := os.MkdirTemp("", "authzhandler-test-")
	if err != nil {
		suite.T().Fatalf("Failed to create temp dir: %v", err)
	}
	suite.tempDir = tempDir

	// Initialize Thunder Runtime config with a file-based database.
	testConfig := &config.Config{
		Database: config.DatabaseConfig{
			Identity: config.DataSource{
				Type: "sqlite",
				Path: filepath.Join(suite.tempDir, "identity.db"),
			},
			Runtime: config.DataSource{
				Type: "sqlite",
				Path: filepath.Join(suite.tempDir, "runtime.db"),
			},
		},
		GateClient: config.GateClientConfig{
			Scheme:    "https",
			Hostname:  "localhost",
			Port:      3000,
			LoginPath: "/login",
			ErrorPath: "/error",
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)

	suite.handler = NewAuthorizeHandler().(*AuthorizeHandler)
}

func (suite *AuthorizeHandlerTestSuite) TearDownTest() {
	config.ResetThunderRuntime()
	// Clean up the temporary directory.
	if err := os.RemoveAll(suite.tempDir); err != nil {
		suite.T().Logf("Failed to clean up temp dir: %v", err)
	}
}

func (suite *AuthorizeHandlerTestSuite) TestNewAuthorizeHandler() {
	handler := NewAuthorizeHandler()
	assert.NotNil(suite.T(), handler)
	assert.Implements(suite.T(), (*AuthorizeHandlerInterface)(nil), handler)
}

func (suite *AuthorizeHandlerTestSuite) TestGetOAuthMessageForGetRequest_Success() {
	req := httptest.NewRequest(http.MethodGet, "/auth?client_id=test-client&redirect_uri=https://example.com", nil)

	msg, err := suite.handler.getOAuthMessageForGetRequest(req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), msg)
	assert.Equal(suite.T(), oauth2const.TypeInitialAuthorizationRequest, msg.RequestType)
	assert.Equal(suite.T(), "test-client", msg.RequestQueryParams["client_id"])
	assert.Equal(suite.T(), "https://example.com", msg.RequestQueryParams["redirect_uri"])
	assert.Nil(suite.T(), msg.SessionData)
}

func (suite *AuthorizeHandlerTestSuite) TestGetOAuthMessageForGetRequest_ParseFormError() {
	// Create a malformed URL to trigger ParseForm error
	req := httptest.NewRequest(http.MethodGet, "/auth?client_id=%ZZ", nil)

	msg, err := suite.handler.getOAuthMessageForGetRequest(req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), msg)
	assert.Contains(suite.T(), err.Error(), "failed to parse form data")
}

func (suite *AuthorizeHandlerTestSuite) TestGetOAuthMessageForPostRequest_MissingSessionDataKey() {
	postData := model.AuthZPostRequest{
		SessionDataKey: "", // Missing session data key
		Assertion:      "test-assertion",
	}
	jsonData, _ := json.Marshal(postData)

	req := httptest.NewRequest(http.MethodPost, "/auth", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")

	msg, err := suite.handler.getOAuthMessageForPostRequest(req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), msg)
	assert.Contains(suite.T(), err.Error(), "sessionDataKey or assertion is missing")
}

func (suite *AuthorizeHandlerTestSuite) TestGetOAuthMessageForPostRequest_MissingAssertion() {
	postData := model.AuthZPostRequest{
		SessionDataKey: "test-session-key",
		Assertion:      "", // Missing assertion
	}
	jsonData, _ := json.Marshal(postData)

	req := httptest.NewRequest(http.MethodPost, "/auth", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")

	msg, err := suite.handler.getOAuthMessageForPostRequest(req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), msg)
	assert.Contains(suite.T(), err.Error(), "sessionDataKey or assertion is missing")
}

func (suite *AuthorizeHandlerTestSuite) TestGetOAuthMessage_UnsupportedMethod() {
	req := httptest.NewRequest(http.MethodPatch, "/auth", nil)
	rr := httptest.NewRecorder()

	msg := suite.handler.getOAuthMessage(req, rr)

	assert.Nil(suite.T(), msg)
	assert.Equal(suite.T(), http.StatusBadRequest, rr.Code)
}

func (suite *AuthorizeHandlerTestSuite) TestGetOAuthMessage_NilRequest() {
	rr := httptest.NewRecorder()

	msg := suite.handler.getOAuthMessage(nil, rr)

	assert.Nil(suite.T(), msg)
}

func (suite *AuthorizeHandlerTestSuite) TestGetOAuthMessage_NilResponseWriter() {
	req := httptest.NewRequest(http.MethodGet, "/auth", nil)

	msg := suite.handler.getOAuthMessage(req, nil)

	assert.Nil(suite.T(), msg)
}

func (suite *AuthorizeHandlerTestSuite) TestGetAuthorizationCode_Success() {
	// Create a valid OAuth message with session data
	sessionData := &sessionmodel.SessionData{
		OAuthParameters: oauth2model.OAuthParameters{
			ClientID:    "test-client",
			RedirectURI: "https://client.example.com/callback",
			Scopes:      "read write",
		},
		AuthTime: time.Now(),
	}

	oAuthMessage := &model.OAuthMessage{
		SessionData: sessionData,
	}

	result, err := getAuthorizationCode(oAuthMessage, "test-user")

	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), result.CodeID)
	assert.NotEmpty(suite.T(), result.Code)
	assert.Equal(suite.T(), "test-client", result.ClientID)
	assert.Equal(suite.T(), "https://client.example.com/callback", result.RedirectURI)
	assert.Equal(suite.T(), "test-user", result.AuthorizedUserID)
	assert.Equal(suite.T(), "read write", result.Scopes)
	assert.Equal(suite.T(), constants.AuthCodeStateActive, result.State)
	assert.NotZero(suite.T(), result.TimeCreated)
	assert.True(suite.T(), result.ExpiryTime.After(result.TimeCreated))
}

func (suite *AuthorizeHandlerTestSuite) TestGetAuthorizationCode_MissingClientID() {
	sessionData := &sessionmodel.SessionData{
		OAuthParameters: oauth2model.OAuthParameters{
			ClientID:    "", // Missing client ID
			RedirectURI: "https://client.example.com/callback",
		},
		AuthTime: time.Now(),
	}

	oAuthMessage := &model.OAuthMessage{
		SessionData: sessionData,
		RequestQueryParams: map[string]string{
			"redirect_uri": "https://client.example.com/callback",
		},
	}

	result, err := getAuthorizationCode(oAuthMessage, "test-user")

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "client_id or redirect_uri is missing")
	assert.Equal(suite.T(), model.AuthorizationCode{}, result)
}

func (suite *AuthorizeHandlerTestSuite) TestGetAuthorizationCode_MissingRedirectURI() {
	sessionData := &sessionmodel.SessionData{
		OAuthParameters: oauth2model.OAuthParameters{
			ClientID:    "test-client",
			RedirectURI: "", // Missing redirect URI
		},
		AuthTime: time.Now(),
	}

	oAuthMessage := &model.OAuthMessage{
		SessionData: sessionData,
		RequestQueryParams: map[string]string{
			"client_id": "test-client",
		},
	}

	result, err := getAuthorizationCode(oAuthMessage, "test-user")

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "client_id or redirect_uri is missing")
	assert.Equal(suite.T(), model.AuthorizationCode{}, result)
}

func (suite *AuthorizeHandlerTestSuite) TestGetAuthorizationCode_EmptyUserID() {
	sessionData := &sessionmodel.SessionData{
		OAuthParameters: oauth2model.OAuthParameters{
			ClientID:    "test-client",
			RedirectURI: "https://client.example.com/callback",
		},
		AuthTime: time.Now(),
	}

	oAuthMessage := &model.OAuthMessage{
		SessionData: sessionData,
	}

	result, err := getAuthorizationCode(oAuthMessage, "") // Empty user ID

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "authenticated user not found")
	assert.Equal(suite.T(), model.AuthorizationCode{}, result)
}

func (suite *AuthorizeHandlerTestSuite) TestGetAuthorizationCode_ZeroAuthTime() {
	sessionData := &sessionmodel.SessionData{
		OAuthParameters: oauth2model.OAuthParameters{
			ClientID:    "test-client",
			RedirectURI: "https://client.example.com/callback",
		},
		AuthTime: time.Time{}, // Zero time
	}

	oAuthMessage := &model.OAuthMessage{
		SessionData: sessionData,
	}

	result, err := getAuthorizationCode(oAuthMessage, "test-user")

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "authentication time is not set")
	assert.Equal(suite.T(), model.AuthorizationCode{}, result)
}

func (suite *AuthorizeHandlerTestSuite) TestGetAuthorizationCode_FallbackToQueryParams() {
	// Test fallback to query params when session data is missing values
	sessionData := &sessionmodel.SessionData{
		OAuthParameters: oauth2model.OAuthParameters{
			ClientID:    "", // Missing in session, should fallback to query params
			RedirectURI: "", // Missing in session, should fallback to query params
			Scopes:      "", // Missing in session, should fallback to query params
		},
		AuthTime: time.Now(),
	}

	oAuthMessage := &model.OAuthMessage{
		SessionData: sessionData,
		RequestQueryParams: map[string]string{
			"client_id":    "fallback-client",
			"redirect_uri": "https://fallback.example.com/callback",
			"scope":        "fallback-scope",
		},
	}

	result, err := getAuthorizationCode(oAuthMessage, "test-user")

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "fallback-client", result.ClientID)
	assert.Equal(suite.T(), "https://fallback.example.com/callback", result.RedirectURI)
	assert.Equal(suite.T(), "fallback-scope", result.Scopes)
}

func (suite *AuthorizeHandlerTestSuite) TestGetLoginPageRedirectURI_Success() {
	queryParams := map[string]string{
		"sessionDataKey": "test-key",
		"appId":          "test-app",
	}

	redirectURI, err := getLoginPageRedirectURI(queryParams)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), redirectURI, "sessionDataKey=test-key")
	assert.Contains(suite.T(), redirectURI, "appId=test-app")
}

func (suite *AuthorizeHandlerTestSuite) TestGetErrorPageRedirectURL_Success() {
	redirectURI, err := getErrorPageRedirectURL("invalid_request", "Missing parameter")
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), redirectURI, "errorCode=invalid_request")
	assert.Contains(suite.T(), redirectURI, "errorMessage=Missing+parameter")
}
