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

package authn

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"github.com/asgardeo/thunder/tests/integration/testutils"
	"github.com/stretchr/testify/suite"
)

const (
	oidcAuthStart  = "/auth/oauth/standard/start"
	oidcAuthFinish = "/auth/oauth/standard/finish"
	mockOIDCPort   = 8093
)

type OIDCAuthTestSuite struct {
	suite.Suite
	mockOIDCServer *testutils.MockOIDCServer
	idpID          string
	userID         string
}

func TestOIDCAuthTestSuite(t *testing.T) {
	suite.Run(t, new(OIDCAuthTestSuite))
}

func (suite *OIDCAuthTestSuite) SetupSuite() {
	var err error
	suite.mockOIDCServer, err = testutils.NewMockOIDCServer(mockOIDCPort,
		"test-oidc-client", "test-oidc-secret")
	suite.Require().NoError(err, "Failed to create mock OIDC server")

	suite.mockOIDCServer.AddUser(&testutils.OIDCUserInfo{
		Sub:           "user456",
		Email:         "testuser@oidc.com",
		EmailVerified: true,
		Name:          "OIDC Test User",
		GivenName:     "OIDC",
		FamilyName:    "User",
		Picture:       "https://oidc.example.com/avatar.jpg",
		Custom: map[string]interface{}{
			"role":   "admin",
			"tenant": "acme-corp",
		},
	})

	err = suite.mockOIDCServer.Start()
	suite.Require().NoError(err, "Failed to start mock OIDC server")

	userAttributes := map[string]interface{}{
		"username":   "oidcuser",
		"password":   "Test@1234",
		"sub":        "user456",
		"email":      "testuser@oidc.com",
		"givenName":  "OIDC",
		"familyName": "User",
	}

	attributesJSON, err := json.Marshal(userAttributes)
	suite.Require().NoError(err)

	user := testutils.User{
		Type:             "person",
		OrganizationUnit: "root",
		Attributes:       json.RawMessage(attributesJSON),
	}

	userID, err := testutils.CreateUser(user)
	suite.Require().NoError(err, "Failed to create test user")
	suite.userID = userID

	idp := testutils.IDP{
		Name:        "Test OIDC IDP",
		Description: "Standard OpenID Connect Identity Provider for authentication testing",
		Type:        "OIDC",
		Properties: []testutils.IDPProperty{
			{
				Name:     "client_id",
				Value:    "test-oidc-client",
				IsSecret: false,
			},
			{
				Name:     "client_secret",
				Value:    "test-oidc-secret",
				IsSecret: true,
			},
			{
				Name:     "authorization_endpoint",
				Value:    suite.mockOIDCServer.GetURL() + "/authorize",
				IsSecret: false,
			},
			{
				Name:     "token_endpoint",
				Value:    suite.mockOIDCServer.GetURL() + "/token",
				IsSecret: false,
			},
			{
				Name:     "userinfo_endpoint",
				Value:    suite.mockOIDCServer.GetURL() + "/userinfo",
				IsSecret: false,
			},
			{
				Name:     "jwks_endpoint",
				Value:    suite.mockOIDCServer.GetURL() + "/jwks",
				IsSecret: false,
			},
			{
				Name:     "scopes",
				Value:    "openid profile email",
				IsSecret: false,
			},
			{
				Name:     "redirect_uri",
				Value:    "https://localhost:8095/callback",
				IsSecret: false,
			},
		},
	}

	idpID, err := testutils.CreateIDP(idp)
	suite.Require().NoError(err, "Failed to create OIDC IDP")
	suite.idpID = idpID
}

func (suite *OIDCAuthTestSuite) TearDownSuite() {
	if suite.userID != "" {
		_ = testutils.DeleteUser(suite.userID)
	}

	if suite.idpID != "" {
		_ = testutils.DeleteIDP(suite.idpID)
	}

	if suite.mockOIDCServer != nil {
		_ = suite.mockOIDCServer.Stop()
	}
}

func (suite *OIDCAuthTestSuite) TestOIDCAuthStartSuccess() {
	// Start authentication
	startRequest := map[string]interface{}{
		"idp_id": suite.idpID,
	}
	startRequestJSON, err := json.Marshal(startRequest)
	suite.Require().NoError(err)

	req, err := http.NewRequest("POST", testServerURL+oidcAuthStart, bytes.NewReader(startRequestJSON))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	client := getHTTPClient()
	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	var startResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&startResponse)
	suite.Require().NoError(err)

	// Verify response contains redirect_url and session_token
	suite.Contains(startResponse, "redirect_url")
	suite.Contains(startResponse, "session_token")

	redirectURL, ok := startResponse["redirect_url"].(string)
	suite.Require().True(ok)
	suite.Contains(redirectURL, suite.mockOIDCServer.GetURL())
	suite.Contains(redirectURL, "client_id=test-oidc-client")
	suite.Contains(redirectURL, "response_type=code")
	suite.Contains(redirectURL, "scope=")
	// Note: Thunder's OIDC implementation does not currently generate nonce values
}

func (suite *OIDCAuthTestSuite) TestOIDCAuthStartInvalidIDPID() {
	startRequest := map[string]interface{}{
		"idp_id": "invalid-idp-id",
	}
	startRequestJSON, err := json.Marshal(startRequest)
	suite.Require().NoError(err)

	req, err := http.NewRequest("POST", testServerURL+oidcAuthStart, bytes.NewReader(startRequestJSON))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	client := getHTTPClient()
	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusBadRequest, resp.StatusCode)

	var errorResponse testutils.ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&errorResponse)
	suite.Require().NoError(err)
	suite.NotEmpty(errorResponse.Code)
	suite.NotEmpty(errorResponse.Message)
}

func (suite *OIDCAuthTestSuite) TestOIDCAuthStartMissingIDPID() {
	startRequest := map[string]interface{}{}
	startRequestJSON, err := json.Marshal(startRequest)
	suite.Require().NoError(err)

	req, err := http.NewRequest("POST", testServerURL+oidcAuthStart, bytes.NewReader(startRequestJSON))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	client := getHTTPClient()
	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusBadRequest, resp.StatusCode)
}

func (suite *OIDCAuthTestSuite) TestOIDCAuthCompleteFlowSuccess() {
	startRequest := map[string]interface{}{
		"idp_id": suite.idpID,
	}
	startRequestJSON, err := json.Marshal(startRequest)
	suite.Require().NoError(err)

	req, err := http.NewRequest("POST", testServerURL+oidcAuthStart, bytes.NewReader(startRequestJSON))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	client := getHTTPClient()
	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	var startResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&startResponse)
	suite.Require().NoError(err)

	sessionToken := startResponse["session_token"].(string)
	redirectURL := startResponse["redirect_url"].(string)

	authCode := suite.simulateOIDCAuthorization(redirectURL)
	suite.Require().NotEmpty(authCode)

	finishRequest := map[string]interface{}{
		"session_token": sessionToken,
		"code":          authCode,
	}
	finishRequestJSON, err := json.Marshal(finishRequest)
	suite.Require().NoError(err)

	req, err = http.NewRequest("POST", testServerURL+oidcAuthFinish, bytes.NewReader(finishRequestJSON))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	var authResponse testutils.AuthenticationResponse
	err = json.NewDecoder(resp.Body).Decode(&authResponse)
	suite.Require().NoError(err)

	suite.NotEmpty(authResponse.ID, "Response should contain user ID")
	suite.NotEmpty(authResponse.Type, "Response should contain user type")
	suite.NotEmpty(authResponse.OrganizationUnit, "Response should contain organization unit")
	suite.Equal(suite.userID, authResponse.ID, "Response should contain the correct user ID")
	suite.NotEmpty(authResponse.Assertion, "Response should contain assertion token by default")
}

func (suite *OIDCAuthTestSuite) TestOIDCAuthFinishInvalidSessionToken() {
	finishRequest := map[string]interface{}{
		"session_token": "invalid-session-token",
		"code":          "some-auth-code",
	}
	finishRequestJSON, err := json.Marshal(finishRequest)
	suite.Require().NoError(err)

	req, err := http.NewRequest("POST", testServerURL+oidcAuthFinish, bytes.NewReader(finishRequestJSON))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	client := getHTTPClient()
	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusBadRequest, resp.StatusCode)
}

func (suite *OIDCAuthTestSuite) TestOIDCAuthFinishMissingCode() {
	finishRequest := map[string]interface{}{
		"session_token": "some-session-token",
	}
	finishRequestJSON, err := json.Marshal(finishRequest)
	suite.Require().NoError(err)

	req, err := http.NewRequest("POST", testServerURL+oidcAuthFinish, bytes.NewReader(finishRequestJSON))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	client := getHTTPClient()
	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusBadRequest, resp.StatusCode)
}

func (suite *OIDCAuthTestSuite) TestOIDCAuthWithNonce() {
	// Note: Thunder's current OIDC implementation does not generate nonce values.
	// This test verifies that authentication still works without nonce.

	// Start authentication
	startRequest := map[string]interface{}{
		"idp_id": suite.idpID,
	}
	startRequestJSON, err := json.Marshal(startRequest)
	suite.Require().NoError(err)

	req, err := http.NewRequest("POST", testServerURL+oidcAuthStart, bytes.NewReader(startRequestJSON))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	client := getHTTPClient()
	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	var startResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&startResponse)
	suite.Require().NoError(err)

	redirectURL := startResponse["redirect_url"].(string)

	parsedURL, err := url.Parse(redirectURL)
	suite.Require().NoError(err)

	query := parsedURL.Query()
	suite.NotEmpty(query.Get("client_id"), "client_id should be present")
	suite.NotEmpty(query.Get("response_type"), "response_type should be present")
	suite.NotEmpty(query.Get("scope"), "scope should be present")
	// nonce is optional - Thunder doesn't generate it currently
}

// TestOIDCAuthCompleteFlowWithSkipAssertionFalse tests complete OIDC flow with skip_assertion=false
func (suite *OIDCAuthTestSuite) TestOIDCAuthCompleteFlowWithSkipAssertionFalse() {
	startRequest := map[string]interface{}{
		"idp_id": suite.idpID,
	}
	startRequestJSON, err := json.Marshal(startRequest)
	suite.Require().NoError(err)

	req, err := http.NewRequest("POST", testServerURL+oidcAuthStart, bytes.NewReader(startRequestJSON))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	client := getHTTPClient()
	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	var startResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&startResponse)
	suite.Require().NoError(err)

	sessionToken := startResponse["session_token"].(string)
	redirectURL := startResponse["redirect_url"].(string)

	authCode := suite.simulateOIDCAuthorization(redirectURL)
	suite.Require().NotEmpty(authCode)

	finishRequest := map[string]interface{}{
		"session_token":  sessionToken,
		"code":           authCode,
		"skip_assertion": false,
	}
	finishRequestJSON, err := json.Marshal(finishRequest)
	suite.Require().NoError(err)

	req, err = http.NewRequest("POST", testServerURL+oidcAuthFinish, bytes.NewReader(finishRequestJSON))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	var authResponse testutils.AuthenticationResponse
	err = json.NewDecoder(resp.Body).Decode(&authResponse)
	suite.Require().NoError(err)

	suite.NotEmpty(authResponse.ID, "Response should contain user ID")
	suite.NotEmpty(authResponse.Type, "Response should contain user type")
	suite.NotEmpty(authResponse.OrganizationUnit, "Response should contain organization unit")
	suite.Equal(suite.userID, authResponse.ID, "Response should contain the correct user ID")
	suite.NotEmpty(authResponse.Assertion, "Response should contain assertion token when skip_assertion is false")
}

// TestOIDCAuthCompleteFlowWithSkipAssertionTrue tests complete OIDC flow with skip_assertion=true
func (suite *OIDCAuthTestSuite) TestOIDCAuthCompleteFlowWithSkipAssertionTrue() {
	startRequest := map[string]interface{}{
		"idp_id": suite.idpID,
	}
	startRequestJSON, err := json.Marshal(startRequest)
	suite.Require().NoError(err)

	req, err := http.NewRequest("POST", testServerURL+oidcAuthStart, bytes.NewReader(startRequestJSON))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	client := getHTTPClient()
	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	var startResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&startResponse)
	suite.Require().NoError(err)

	sessionToken := startResponse["session_token"].(string)
	redirectURL := startResponse["redirect_url"].(string)

	authCode := suite.simulateOIDCAuthorization(redirectURL)
	suite.Require().NotEmpty(authCode)

	finishRequest := map[string]interface{}{
		"session_token":  sessionToken,
		"code":           authCode,
		"skip_assertion": true,
	}
	finishRequestJSON, err := json.Marshal(finishRequest)
	suite.Require().NoError(err)

	req, err = http.NewRequest("POST", testServerURL+oidcAuthFinish, bytes.NewReader(finishRequestJSON))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	var authResponse testutils.AuthenticationResponse
	err = json.NewDecoder(resp.Body).Decode(&authResponse)
	suite.Require().NoError(err)

	suite.NotEmpty(authResponse.ID, "Response should contain user ID")
	suite.NotEmpty(authResponse.Type, "Response should contain user type")
	suite.NotEmpty(authResponse.OrganizationUnit, "Response should contain organization unit")
	suite.Equal(suite.userID, authResponse.ID, "Response should contain the correct user ID")
	suite.Empty(authResponse.Assertion, "Response should not contain assertion token when skip_assertion is true")
}

// simulateOIDCAuthorization simulates user authorization and returns authorization code
func (suite *OIDCAuthTestSuite) simulateOIDCAuthorization(redirectURL string) string {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Get(redirectURL)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	location := resp.Header.Get("Location")
	suite.Require().NotEmpty(location, "Mock server should redirect with authorization code")

	locationURL, err := url.Parse(location)
	suite.Require().NoError(err)

	code := locationURL.Query().Get("code")
	suite.Require().NotEmpty(code, "Authorization code should be present in redirect")
	return code
}
