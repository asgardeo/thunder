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
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
)

type OAuth2UtilsTestSuite struct {
	suite.Suite
}

func TestOAuth2UtilsSuite(t *testing.T) {
	suite.Run(t, new(OAuth2UtilsTestSuite))
}

func (suite *OAuth2UtilsTestSuite) TestGetURIWithQueryParams_Success() {
	testCases := []struct {
		name        string
		uri         string
		queryParams map[string]string
		expectedURI string
	}{
		{
			name: "SimpleParams",
			uri:  "https://example.com/callback",
			queryParams: map[string]string{
				"code":  "test-code",
				"state": "test-state",
			},
			expectedURI: "https://example.com/callback?code=test-code&state=test-state",
		},
		{
			name: "EmptyParams",
			uri:  "https://example.com/callback",
			queryParams: map[string]string{},
			expectedURI: "https://example.com/callback",
		},
		{
			name:        "NilParams",
			uri:         "https://example.com/callback",
			queryParams: nil,
			expectedURI: "https://example.com/callback",
		},
		{
			name: "ValidErrorParams",
			uri:  "https://example.com/callback",
			queryParams: map[string]string{
				constants.RequestParamError:            "invalid_request",
				constants.RequestParamErrorDescription: "Missing client_id parameter",
			},
			expectedURI: "https://example.com/callback?error=invalid_request&error_description=Missing+client_id+parameter",
		},
		{
			name: "SpecialCharactersInParams",
			uri:  "https://example.com/callback",
			queryParams: map[string]string{
				"redirect_uri": "https://client.example.com/cb?param=value",
				"scope":        "read write admin",
			},
			expectedURI: "https://example.com/callback?redirect_uri=https%3A%2F%2Fclient.example.com%2Fcb%3Fparam%3Dvalue&scope=read+write+admin",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			result, err := GetURIWithQueryParams(tc.uri, tc.queryParams)
			assert.NoError(t, err)
			
			// Parse both URIs to compare them properly (query params can be in different order)
			expectedParsed, err := url.Parse(tc.expectedURI)
			assert.NoError(t, err)
			resultParsed, err := url.Parse(result)
			assert.NoError(t, err)
			
			assert.Equal(t, expectedParsed.Scheme, resultParsed.Scheme)
			assert.Equal(t, expectedParsed.Host, resultParsed.Host)
			assert.Equal(t, expectedParsed.Path, resultParsed.Path)
			
			// Compare query parameters
			expectedQuery := expectedParsed.Query()
			resultQuery := resultParsed.Query()
			assert.Equal(t, expectedQuery, resultQuery)
		})
	}
}

func (suite *OAuth2UtilsTestSuite) TestGetURIWithQueryParams_InvalidErrorCode() {
	testCases := []struct {
		name        string
		errorCode   string
		description string
	}{
		{
			name:      "InvalidCharacterInErrorCode",
			errorCode: "invalid\x22request", // Contains quote character which is not allowed
		},
		{
			name:      "ControlCharacterInErrorCode",
			errorCode: "invalid\x01request", // Contains control character
		},
		{
			name:      "InvalidCharacterAtStart",
			errorCode: "\x19invalid_request", // Control character at start
		},
		{
			name:      "InvalidCharacterAtEnd",
			errorCode: "invalid_request\x7F", // DEL character at end
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			queryParams := map[string]string{
				constants.RequestParamError: tc.errorCode,
			}
			if tc.description != "" {
				queryParams[constants.RequestParamErrorDescription] = tc.description
			}

			result, err := GetURIWithQueryParams("https://example.com/callback", queryParams)
			
			assert.Error(t, err)
			assert.Empty(t, result)
			assert.Contains(t, err.Error(), "invalid error code")
		})
	}
}

func (suite *OAuth2UtilsTestSuite) TestGetURIWithQueryParams_InvalidErrorDescription() {
	testCases := []struct {
		name        string
		description string
	}{
		{
			name:        "InvalidCharacterInDescription",
			description: "Missing \"client_id\" parameter", // Contains quote character
		},
		{
			name:        "ControlCharacterInDescription",
			description: "Missing\x01client_id parameter", // Contains control character
		},
		{
			name:        "BackslashInDescription",
			description: "Missing\\client_id parameter", // Contains backslash (\x5C)
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			queryParams := map[string]string{
				constants.RequestParamError:            "invalid_request",
				constants.RequestParamErrorDescription: tc.description,
			}

			result, err := GetURIWithQueryParams("https://example.com/callback", queryParams)
			
			assert.Error(t, err)
			assert.Empty(t, result)
			assert.Contains(t, err.Error(), "invalid error description")
		})
	}
}

func (suite *OAuth2UtilsTestSuite) TestGetURIWithQueryParams_ValidCharacterRange() {
	// Test with characters from the allowed range: %x20-21 / %x23-5B / %x5D-7E
	validChars := " !#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[]^_`abcdefghijklmnopqrstuvwxyz{|}~"
	
	queryParams := map[string]string{
		constants.RequestParamError:            "invalid_request",
		constants.RequestParamErrorDescription: validChars,
	}

	result, err := GetURIWithQueryParams("https://example.com/callback", queryParams)
	
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), result)
}

func (suite *OAuth2UtilsTestSuite) TestGetURIWithQueryParams_EmptyErrorParams() {
	// Test with empty error parameters (should be valid)
	queryParams := map[string]string{
		constants.RequestParamError:            "",
		constants.RequestParamErrorDescription: "",
		"other_param":                          "value",
	}

	result, err := GetURIWithQueryParams("https://example.com/callback", queryParams)
	
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), result, "other_param=value")
}

func (suite *OAuth2UtilsTestSuite) TestGetURIWithQueryParams_OnlyErrorCode() {
	// Test with only error code, no description
	queryParams := map[string]string{
		constants.RequestParamError: "invalid_client",
	}

	result, err := GetURIWithQueryParams("https://example.com/callback", queryParams)
	
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), result, "error=invalid_client")
	assert.NotContains(suite.T(), result, "error_description")
}

func (suite *OAuth2UtilsTestSuite) TestGetURIWithQueryParams_OnlyErrorDescription() {
	// Test with only error description, no error code
	queryParams := map[string]string{
		constants.RequestParamErrorDescription: "Something went wrong",
	}

	result, err := GetURIWithQueryParams("https://example.com/callback", queryParams)
	
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), result, "error_description=Something+went+wrong")
	assert.NotContains(suite.T(), result, "error=")
}

func (suite *OAuth2UtilsTestSuite) TestValidateErrorParams_DirectCall() {
	// Test the validateErrorParams function directly (even though it's not exported)
	// We test it through the public function
	
	testCases := []struct {
		name        string
		errorCode   string
		description string
		expectError bool
	}{
		{
			name:        "ValidParams",
			errorCode:   "invalid_request",
			description: "Missing required parameter",
			expectError: false,
		},
		{
			name:        "EmptyParams",
			errorCode:   "",
			description: "",
			expectError: false,
		},
		{
			name:        "InvalidErrorCode",
			errorCode:   "invalid\"request",
			description: "Valid description",
			expectError: true,
		},
		{
			name:        "InvalidDescription",
			errorCode:   "invalid_request",
			description: "Invalid\"description",
			expectError: true,
		},
		{
			name:        "BothInvalid",
			errorCode:   "invalid\"request",
			description: "Invalid\"description",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			queryParams := map[string]string{
				constants.RequestParamError:            tc.errorCode,
				constants.RequestParamErrorDescription: tc.description,
			}

			_, err := GetURIWithQueryParams("https://example.com/callback", queryParams)
			
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (suite *OAuth2UtilsTestSuite) TestGetURIWithQueryParams_MalformedBaseURI() {
	// Test that the function still validates error params even with malformed base URI
	queryParams := map[string]string{
		constants.RequestParamError: "invalid\"request", // Invalid character
	}

	result, err := GetURIWithQueryParams("not-a-valid-uri", queryParams)
	
	// Should fail on error validation before URI processing
	assert.Error(suite.T(), err)
	assert.Empty(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "invalid error code")
}

func (suite *OAuth2UtilsTestSuite) TestGetURIWithQueryParams_SpecialErrorCodes() {
	// Test with standard OAuth2 error codes
	standardErrorCodes := []string{
		"invalid_request",
		"invalid_client",
		"invalid_grant",
		"unauthorized_client",
		"unsupported_grant_type",
		"invalid_scope",
		"server_error",
		"temporarily_unavailable",
		"unsupported_response_type",
		"access_denied",
	}

	for _, errorCode := range standardErrorCodes {
		suite.T().Run("ErrorCode_"+errorCode, func(t *testing.T) {
			queryParams := map[string]string{
				constants.RequestParamError:            errorCode,
				constants.RequestParamErrorDescription: "Standard OAuth2 error",
			}

			result, err := GetURIWithQueryParams("https://example.com/callback", queryParams)
			
			assert.NoError(t, err)
			assert.Contains(t, result, "error="+errorCode)
		})
	}
}

func (suite *OAuth2UtilsTestSuite) TestGetURIWithQueryParams_BoundaryCharacters() {
	// Test characters at the boundaries of allowed ranges
	testCases := []struct {
		name      string
		char      string
		shouldPass bool
	}{
		{"Space_0x20", "\x20", true},      // First allowed character
		{"Exclamation_0x21", "\x21", true}, // Last of first range
		{"Quote_0x22", "\x22", false},     // Not allowed (between ranges)
		{"Hash_0x23", "\x23", true},       // First of second range
		{"LeftBracket_0x5B", "\x5B", true}, // Last of second range
		{"Backslash_0x5C", "\x5C", false}, // Not allowed (between ranges)
		{"RightBracket_0x5D", "\x5D", true}, // First of third range
		{"Tilde_0x7E", "\x7E", true},      // Last allowed character
		{"DEL_0x7F", "\x7F", false},       // Not allowed (after range)
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			queryParams := map[string]string{
				constants.RequestParamError: "test_error" + tc.char,
			}

			_, err := GetURIWithQueryParams("https://example.com/callback", queryParams)
			
			if tc.shouldPass {
				assert.NoError(t, err, "Character %s should be allowed", tc.char)
			} else {
				assert.Error(t, err, "Character %s should not be allowed", tc.char)
			}
		})
	}
}