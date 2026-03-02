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

package clientauth

import (
	"net/http"

	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
)

// authError represents an authentication error.
type authError struct {
	ErrorCode        string
	ErrorDescription string
	StatusCode       int
	ResponseHeaders  map[string]string
}

// newAuthError creates a new authentication error.
func newAuthError(errorCode, errorDescription string, statusCode int) *authError {
	return &authError{
		ErrorCode:        errorCode,
		ErrorDescription: errorDescription,
		StatusCode:       statusCode,
	}
}

// newAuthErrorWithHeaders creates a new authentication error with response headers.
func newAuthErrorWithHeaders(errorCode, errorDescription string, statusCode int, headers map[string]string) *authError {
	return &authError{
		ErrorCode:        errorCode,
		ErrorDescription: errorDescription,
		StatusCode:       statusCode,
		ResponseHeaders:  headers,
	}
}

// Common authentication errors
var (
	errInvalidAuthorizationHeader = newAuthErrorWithHeaders(
		constants.ErrorInvalidClient,
		"Invalid client credentials",
		http.StatusUnauthorized,
		map[string]string{"WWW-Authenticate": "Basic"},
	)
	errInvalidClientCredentials = newAuthError(
		constants.ErrorInvalidClient,
		"Invalid client credentials",
		http.StatusUnauthorized,
	)
	errMultipleAuthMethods = newAuthError(
		constants.ErrorInvalidRequest,
		"Multiple client authentication methods were provided",
		http.StatusBadRequest,
	)
	errMissingClientID = newAuthError(
		constants.ErrorInvalidClient,
		"Missing client_id parameter",
		http.StatusUnauthorized,
	)
	errUnauthorizedAuthMethod = newAuthError(
		constants.ErrorUnauthorizedClient,
		"Client is not allowed to use the specified token endpoint authentication method",
		http.StatusUnauthorized,
	)
	errClientIDMismatch = newAuthError(
		constants.ErrorInvalidRequest,
		"client_id in request body does not match client_id from authentication credentials",
		http.StatusBadRequest,
	)
	errInvalidClientAssertion = newAuthError(
		constants.ErrorInvalidClient,
		"Invalid client assertion",
		http.StatusUnauthorized,
	)
)
