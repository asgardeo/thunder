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
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/asgardeo/thunder/tests/integration/testutils"
	"github.com/stretchr/testify/suite"
)

const (
	testServerURL    = testutils.TestServerURL
	googleAuthStart  = "/auth/oauth/google/start"
	googleAuthFinish = "/auth/oauth/google/finish"
	mockGooglePort   = 8090
)

type GoogleAuthTestSuite struct {
	suite.Suite
	mockGoogleServer *testutils.MockGoogleOIDCServer
	idpID            string
	userID           string
}

func TestGoogleAuthTestSuite(t *testing.T) {
	suite.Run(t, new(GoogleAuthTestSuite))
}

func (suite *GoogleAuthTestSuite) SetupSuite() {
	mockServer, err := testutils.NewMockGoogleOIDCServer(mockGooglePort,
		"test-google-client", "test-google-secret")
	suite.Require().NoError(err, "Failed to create mock Google server")
	suite.mockGoogleServer = mockServer

	suite.mockGoogleServer.AddUser(&testutils.GoogleUserInfo{
		Sub:           "google-test-user-123",
		Email:         "testuser@gmail.com",
		EmailVerified: true,
		Name:          "Test User",
		GivenName:     "Test",
		FamilyName:    "User",
		Picture:       "https://example.com/picture.jpg",
		Locale:        "en",
	})

	err = suite.mockGoogleServer.Start()
	suite.Require().NoError(err, "Failed to start mock Google server")

	userAttributes := map[string]interface{}{
		"username":   "googleuser",
		"password":   "Test@1234",
		"sub":        "google-test-user-123",
		"email":      "testuser@gmail.com",
		"givenName":  "Test",
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
		Name:        "Test Google IDP",
		Description: "Google Identity Provider for authentication testing",
		Type:        "GOOGLE",
		Properties: []testutils.IDPProperty{
			{
				Name:     "client_id",
				Value:    "test-google-client",
				IsSecret: false,
			},
			{
				Name:     "client_secret",
				Value:    "test-google-secret",
				IsSecret: true,
			},
			{
				Name:     "authorization_endpoint",
				Value:    suite.mockGoogleServer.GetURL() + "/o/oauth2/v2/auth",
				IsSecret: false,
			},
			{
				Name:     "token_endpoint",
				Value:    suite.mockGoogleServer.GetURL() + "/token",
				IsSecret: false,
			},
			{
				Name:     "userinfo_endpoint",
				Value:    suite.mockGoogleServer.GetURL() + "/v1/userinfo",
				IsSecret: false,
			},
			{
				Name:     "jwks_endpoint",
				Value:    suite.mockGoogleServer.GetURL() + "/oauth2/v3/certs",
				IsSecret: false,
			},
			{
				Name:     "scopes",
				Value:    "openid email profile",
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
	suite.Require().NoError(err, "Failed to create Google IDP")
	suite.idpID = idpID
}

func (suite *GoogleAuthTestSuite) TearDownSuite() {
	if suite.userID != "" {
		_ = testutils.DeleteUser(suite.userID)
	}

	if suite.idpID != "" {
		_ = testutils.DeleteIDP(suite.idpID)
	}

	if suite.mockGoogleServer != nil {
		_ = suite.mockGoogleServer.Stop()
	}
}

func (suite *GoogleAuthTestSuite) TestGoogleAuthStartSuccess() {
	startRequest := map[string]interface{}{
		"idp_id": suite.idpID,
	}
	startRequestJSON, err := json.Marshal(startRequest)
	suite.Require().NoError(err)

	req, err := http.NewRequest("POST", testServerURL+googleAuthStart, bytes.NewReader(startRequestJSON))
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

	suite.Contains(startResponse, "redirect_url")
	suite.Contains(startResponse, "session_token")

	redirectURL, ok := startResponse["redirect_url"].(string)
	suite.Require().True(ok)
	suite.Contains(redirectURL, suite.mockGoogleServer.GetURL())
	suite.Contains(redirectURL, "client_id=test-google-client")
	suite.Contains(redirectURL, "response_type=code")
}

func (suite *GoogleAuthTestSuite) TestGoogleAuthStartInvalidIDPID() {
	startRequest := map[string]interface{}{
		"idp_id": "invalid-idp-id",
	}
	startRequestJSON, err := json.Marshal(startRequest)
	suite.Require().NoError(err)

	req, err := http.NewRequest("POST", testServerURL+googleAuthStart, bytes.NewReader(startRequestJSON))
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

func (suite *GoogleAuthTestSuite) TestGoogleAuthStartMissingIDPID() {
	startRequest := map[string]interface{}{}
	startRequestJSON, err := json.Marshal(startRequest)
	suite.Require().NoError(err)

	req, err := http.NewRequest("POST", testServerURL+googleAuthStart, bytes.NewReader(startRequestJSON))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	client := getHTTPClient()
	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusBadRequest, resp.StatusCode)
}

func (suite *GoogleAuthTestSuite) TestGoogleAuthCompleteFlowSuccess() {
	startRequest := map[string]interface{}{
		"idp_id": suite.idpID,
	}
	startRequestJSON, err := json.Marshal(startRequest)
	suite.Require().NoError(err)

	req, err := http.NewRequest("POST", testServerURL+googleAuthStart, bytes.NewReader(startRequestJSON))
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

	authCode := suite.simulateGoogleAuthorization(redirectURL)
	suite.Require().NotEmpty(authCode)

	finishRequest := map[string]interface{}{
		"session_token": sessionToken,
		"code":          authCode,
	}
	finishRequestJSON, err := json.Marshal(finishRequest)
	suite.Require().NoError(err)

	req, err = http.NewRequest("POST", testServerURL+googleAuthFinish, bytes.NewReader(finishRequestJSON))
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

func (suite *GoogleAuthTestSuite) TestGoogleAuthFinishInvalidSessionToken() {
	finishRequest := map[string]interface{}{
		"session_token": "invalid-session-token",
		"code":          "some-auth-code",
	}
	finishRequestJSON, err := json.Marshal(finishRequest)
	suite.Require().NoError(err)

	req, err := http.NewRequest("POST", testServerURL+googleAuthFinish, bytes.NewReader(finishRequestJSON))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	client := getHTTPClient()
	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusBadRequest, resp.StatusCode)
}

func (suite *GoogleAuthTestSuite) TestGoogleAuthFinishMissingCode() {
	finishRequest := map[string]interface{}{
		"session_token": "some-session-token",
	}
	finishRequestJSON, err := json.Marshal(finishRequest)
	suite.Require().NoError(err)

	req, err := http.NewRequest("POST", testServerURL+googleAuthFinish, bytes.NewReader(finishRequestJSON))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	client := getHTTPClient()
	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusBadRequest, resp.StatusCode)
}

// TestGoogleAuthCompleteFlowWithSkipAssertionFalse tests complete Google auth flow with skip_assertion=false
func (suite *GoogleAuthTestSuite) TestGoogleAuthCompleteFlowWithSkipAssertionFalse() {
	startRequest := map[string]interface{}{
		"idp_id": suite.idpID,
	}
	startRequestJSON, err := json.Marshal(startRequest)
	suite.Require().NoError(err)

	req, err := http.NewRequest("POST", testServerURL+googleAuthStart, bytes.NewReader(startRequestJSON))
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

	authCode := suite.simulateGoogleAuthorization(redirectURL)
	suite.Require().NotEmpty(authCode)

	finishRequest := map[string]interface{}{
		"session_token":  sessionToken,
		"code":           authCode,
		"skip_assertion": false,
	}
	finishRequestJSON, err := json.Marshal(finishRequest)
	suite.Require().NoError(err)

	req, err = http.NewRequest("POST", testServerURL+googleAuthFinish, bytes.NewReader(finishRequestJSON))
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

// TestGoogleAuthCompleteFlowWithSkipAssertionTrue tests complete Google auth flow with skip_assertion=true
func (suite *GoogleAuthTestSuite) TestGoogleAuthCompleteFlowWithSkipAssertionTrue() {
	startRequest := map[string]interface{}{
		"idp_id": suite.idpID,
	}
	startRequestJSON, err := json.Marshal(startRequest)
	suite.Require().NoError(err)

	req, err := http.NewRequest("POST", testServerURL+googleAuthStart, bytes.NewReader(startRequestJSON))
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

	authCode := suite.simulateGoogleAuthorization(redirectURL)
	suite.Require().NotEmpty(authCode)

	finishRequest := map[string]interface{}{
		"session_token":  sessionToken,
		"code":           authCode,
		"skip_assertion": true,
	}
	finishRequestJSON, err := json.Marshal(finishRequest)
	suite.Require().NoError(err)

	req, err = http.NewRequest("POST", testServerURL+googleAuthFinish, bytes.NewReader(finishRequestJSON))
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

// simulateGoogleAuthorization simulates user authorization and returns authorization code
func (suite *GoogleAuthTestSuite) simulateGoogleAuthorization(redirectURL string) string {
	parsedURL, err := url.Parse(redirectURL)
	suite.Require().NoError(err)

	query := parsedURL.Query()
	clientID := query.Get("client_id")
	state := query.Get("state")
	redirectURI := query.Get("redirect_uri")
	scope := query.Get("scope")

	authURL := fmt.Sprintf("%s/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s",
		suite.mockGoogleServer.GetURL(),
		url.QueryEscape(clientID),
		url.QueryEscape(redirectURI),
		url.QueryEscape(scope),
		url.QueryEscape(state))

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Get(authURL)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	location := resp.Header.Get("Location")
	suite.Require().NotEmpty(location)

	locationURL, err := url.Parse(location)
	suite.Require().NoError(err)

	code := locationURL.Query().Get("code")
	return code
}

// TestGoogleAuthWithAssuranceLevelAAL1 tests that Google authentication generates AAL1 assurance level
func (suite *GoogleAuthTestSuite) TestGoogleAuthWithAssuranceLevelAAL1() {
	// Step 1: Start Google authentication
	startReq := map[string]interface{}{
		"idp_id": suite.idpID,
	}
	startReqJSON, err := json.Marshal(startReq)
	suite.Require().NoError(err)

	req, err := http.NewRequest("POST", testServerURL+googleAuthStart, bytes.NewReader(startReqJSON))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	client := getHTTPClient()
	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	var startResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&startResp)
	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)

	sessionToken := startResp["session_token"].(string)
	suite.NotEmpty(sessionToken)

	// Step 2: Get authorization code from mock server
	code := suite.simulateGoogleAuthorization(startResp["redirect_url"].(string))
	suite.NotEmpty(code)

	// Step 3: Finish Google authentication
	finishReq := map[string]interface{}{
		"session_token": sessionToken,
		"code":          code,
	}
	finishReqJSON, err := json.Marshal(finishReq)
	suite.Require().NoError(err)

	req, err = http.NewRequest("POST", testServerURL+googleAuthFinish, bytes.NewReader(finishReqJSON))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	var authResp testutils.AuthenticationResponse
	err = json.NewDecoder(resp.Body).Decode(&authResp)
	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)

	suite.NotEmpty(authResp.Assertion, "Response should contain assertion token by default")

	// Verify assertion contains AAL1 for single-factor Google authentication
	aal := extractAssuranceLevelFromAssertion(authResp.Assertion, "aal")
	suite.NotEmpty(aal, "Assertion should contain AAL information")
	suite.Equal("AAL1", aal, "Single-factor Google authentication should result in AAL1")

	// Verify IAL is present
	ial := extractAssuranceLevelFromAssertion(authResp.Assertion, "ial")
	suite.NotEmpty(ial, "Assertion should contain IAL information")
	suite.Equal("IAL1", ial, "Self-asserted identity should result in IAL1")
}

// TestGoogleAuthWithSkipAssertion tests Google authentication with skip_assertion=true
func (suite *GoogleAuthTestSuite) TestGoogleAuthWithSkipAssertion() {
	// Start authentication
	startReq := map[string]interface{}{
		"idp_id": suite.idpID,
	}
	startReqJSON, err := json.Marshal(startReq)
	suite.Require().NoError(err)

	req, err := http.NewRequest("POST", testServerURL+googleAuthStart, bytes.NewReader(startReqJSON))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	client := getHTTPClient()
	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	var startResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&startResp)
	suite.Require().NoError(err)

	sessionToken := startResp["session_token"].(string)
	code := suite.simulateGoogleAuthorization(startResp["redirect_url"].(string))

	// Finish with skip_assertion=true
	finishReq := map[string]interface{}{
		"session_token":  sessionToken,
		"code":           code,
		"skip_assertion": true,
	}
	finishReqJSON, err := json.Marshal(finishReq)
	suite.Require().NoError(err)

	req, err = http.NewRequest("POST", testServerURL+googleAuthFinish, bytes.NewReader(finishReqJSON))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	var authResp testutils.AuthenticationResponse
	err = json.NewDecoder(resp.Body).Decode(&authResp)
	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)

	suite.Empty(authResp.Assertion, "Response should not contain assertion when skip_assertion is true")
}

// getHTTPClient returns a configured HTTP client
func getHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}
