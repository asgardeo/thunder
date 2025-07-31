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

package utils

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
)

const testRedirectURI = "http://localhost:8080/callback"

type OAuthUtilsTestSuite struct {
	suite.Suite
}

func TestOAuthUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(OAuthUtilsTestSuite))
}

func (suite *OAuthUtilsTestSuite) TestGetOAuthMessage_ValidInitialAuthorizationRequest() {
	// Create a valid initial authorization request
	url := "/oauth2/authorize?client_id=test_client&response_type=code&redirect_uri=" + testRedirectURI
	req := httptest.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()

	// Parse form to populate PostForm
	err := req.ParseForm()
	assert.NoError(suite.T(), err)

	message, err := GetOAuthMessage(req, w)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), message)
	assert.Equal(suite.T(), constants.TypeInitialAuthorizationRequest, message.RequestType)
	assert.Equal(suite.T(), "test_client", message.RequestQueryParams[constants.ClientID])
	assert.Equal(suite.T(), "code", message.RequestQueryParams[constants.ResponseType])
	assert.Equal(suite.T(), testRedirectURI, message.RequestQueryParams[constants.RedirectURI])
}

func (suite *OAuthUtilsTestSuite) TestGetOAuthMessage_ValidAuthorizationResponseFromFramework() {
	// Create a request with session data key
	req := httptest.NewRequest("POST", "/oauth2/authorize", strings.NewReader("sessionDataKey=test_session_key"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	// Parse form to populate PostForm
	err := req.ParseForm()
	assert.NoError(suite.T(), err)

	message, err := GetOAuthMessage(req, w)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), message)
	assert.Equal(suite.T(), constants.TypeAuthorizationResponseFromFramework, message.RequestType)
	assert.Equal(suite.T(), "test_session_key", message.RequestBodyParams[constants.SessionDataKey])
}

func (suite *OAuthUtilsTestSuite) TestGetOAuthMessage_NilRequest() {
	w := httptest.NewRecorder()

	message, err := GetOAuthMessage(nil, w)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), message)
	assert.Equal(suite.T(), "request or response writer is nil", err.Error())
}

func (suite *OAuthUtilsTestSuite) TestGetOAuthMessage_NilResponseWriter() {
	req := httptest.NewRequest("GET", "/oauth2/authorize", nil)

	message, err := GetOAuthMessage(req, nil)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), message)
	assert.Equal(suite.T(), "request or response writer is nil", err.Error())
}

func (suite *OAuthUtilsTestSuite) TestGetOAuthMessage_InvalidRequestType() {
	// Create a request that doesn't match any request type pattern
	req := httptest.NewRequest("GET", "/oauth2/authorize", nil)
	w := httptest.NewRecorder()

	// Parse form to populate PostForm
	err := req.ParseForm()
	assert.NoError(suite.T(), err)

	message, err := GetOAuthMessage(req, w)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), message)
	assert.Equal(suite.T(), "invalid request type", err.Error())
}

func (suite *OAuthUtilsTestSuite) TestGetOAuthMessage_FormParsingError() {
	// Create a request with malformed form data
	req := httptest.NewRequest("POST", "/oauth2/authorize", strings.NewReader("invalid%form%data"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	// Don't parse form beforehand to trigger parsing error
	message, err := GetOAuthMessage(req, w)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), message)
	assert.Contains(suite.T(), err.Error(), "failed to parse form data")
}

func (suite *OAuthUtilsTestSuite) TestGetURIWithQueryParams_ValidParams() {
	uri := testRedirectURI
	params := map[string]string{
		"code":  "authorization_code",
		"state": "test_state",
	}

	result, err := GetURIWithQueryParams(uri, params)

	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), result, "code=authorization_code")
	assert.Contains(suite.T(), result, "state=test_state")
}

func (suite *OAuthUtilsTestSuite) TestGetURIWithQueryParams_WithError() {
	uri := testRedirectURI
	params := map[string]string{
		constants.Error:            "invalid_request",
		constants.ErrorDescription: "Invalid client ID",
	}

	result, err := GetURIWithQueryParams(uri, params)

	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), result, "error=invalid_request")
	assert.Contains(suite.T(), result, url.QueryEscape("Invalid client ID"))
}

func (suite *OAuthUtilsTestSuite) TestGetURIWithQueryParams_InvalidError() {
	uri := testRedirectURI
	params := map[string]string{
		constants.Error: "invalid\"error", // Contains invalid character
	}

	result, err := GetURIWithQueryParams(uri, params)

	assert.Error(suite.T(), err)
	assert.Empty(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "invalid error code")
}

func (suite *OAuthUtilsTestSuite) TestValidateErrorParams_ValidErrorCode() {
	err := validateErrorParams("invalid_request", "")
	assert.NoError(suite.T(), err)
}

func (suite *OAuthUtilsTestSuite) TestValidateErrorParams_ValidErrorDescription() {
	err := validateErrorParams("", "Client authentication failed")
	assert.NoError(suite.T(), err)
}

func (suite *OAuthUtilsTestSuite) TestValidateErrorParams_ValidBoth() {
	err := validateErrorParams("invalid_client", "Client authentication failed")
	assert.NoError(suite.T(), err)
}

func (suite *OAuthUtilsTestSuite) TestValidateErrorParams_EmptyBoth() {
	err := validateErrorParams("", "")
	assert.NoError(suite.T(), err)
}

func (suite *OAuthUtilsTestSuite) TestValidateErrorParams_InvalidErrorCode() {
	// Test with invalid character (quotation mark)
	err := validateErrorParams("invalid\"error", "")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "invalid error code")
}

func (suite *OAuthUtilsTestSuite) TestValidateErrorParams_InvalidErrorDescription() {
	// Test with invalid character (quotation mark)
	err := validateErrorParams("", "Invalid\"description")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "invalid error description")
}

func (suite *OAuthUtilsTestSuite) TestValidateErrorParams_AllowedSpecialCharacters() {
	// Test with characters at boundary of allowed range
	err := validateErrorParams("error_with-special.chars", "Description with spaces and symbols!@#$%^&*()_+-=[]{}|;:,.<>?")
	assert.NoError(suite.T(), err)
}
