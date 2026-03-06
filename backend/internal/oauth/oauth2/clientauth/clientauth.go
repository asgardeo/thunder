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

// Package clientauth provides shared client authentication logic for OAuth2 endpoints.
package clientauth

import (
	"encoding/base64"
	"net/http"
	"net/url"
	"strings"

	"github.com/asgardeo/thunder/internal/application"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/utils"
)

// authenticate authenticates the OAuth2 client from the request.
// It extracts credentials, validates them, and returns OAuthClientInfo on success.
// Returns an authError on failure.
func authenticate(
	r *http.Request,
	appService application.ApplicationServiceInterface,
) (*OAuthClientInfo, *authError) {
	var clientID string
	var clientSecret string
	var tokenAuthMethod constants.TokenEndpointAuthMethod

	// Check Authorization header (client_secret_basic)
	if r.Header.Get(serverconst.AuthorizationHeaderName) != "" {
		var err *authError
		clientID, clientSecret, err = extractBasicAuthCredentials(r)
		if err != nil {
			return nil, err
		}
	}

	// Check request body (client_secret_post)
	clientIDFromBody := r.FormValue(constants.RequestParamClientID)
	clientSecretFromBody := r.FormValue(constants.RequestParamClientSecret)

	// Error if both header and body have credentials
	if (clientID != "" || clientSecret != "") &&
		(clientIDFromBody != "" || clientSecretFromBody != "") {
		return nil, errBothHeaderAndBody
	}

	// Determine authentication method
	if clientID != "" && clientSecret != "" {
		tokenAuthMethod = constants.TokenEndpointAuthMethodClientSecretBasic
	}

	if clientIDFromBody != "" {
		clientID = clientIDFromBody
		if clientSecretFromBody != "" {
			clientSecret = clientSecretFromBody
			tokenAuthMethod = constants.TokenEndpointAuthMethodClientSecretPost
		} else {
			tokenAuthMethod = constants.TokenEndpointAuthMethodNone
		}
	}

	// Validate required fields
	if clientID == "" {
		return nil, errMissingClientID
	}

	if clientSecret == "" && tokenAuthMethod != constants.TokenEndpointAuthMethodNone {
		return nil, errMissingClientSecret
	}

	// Retrieve the OAuth application
	oauthApp, err := appService.GetOAuthApplication(clientID)
	if err != nil || oauthApp == nil {
		return nil, errInvalidClientCredentials
	}

	// Validate the token endpoint authentication method
	if !oauthApp.IsAllowedTokenEndpointAuthMethod(tokenAuthMethod) {
		if tokenAuthMethod == constants.TokenEndpointAuthMethodNone {
			return nil, errMissingClientSecret
		}
		return nil, errUnauthorizedAuthMethod
	}

	// Validate the client credentials
	if tokenAuthMethod != constants.TokenEndpointAuthMethodNone {
		if !oauthApp.ValidateCredentials(clientID, clientSecret) {
			return nil, errInvalidClientCredentials
		}
	}

	return &OAuthClientInfo{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		OAuthApp:     oauthApp,
	}, nil
}

// extractBasicAuthCredentials extracts the basic authentication credentials from the request header.
func extractBasicAuthCredentials(r *http.Request) (string, string, *authError) {
	authHeader := r.Header.Get(serverconst.AuthorizationHeaderName)
	if !utils.HasPrefixFold(authHeader, serverconst.AuthSchemeBasic) {
		return "", "", errInvalidAuthorizationHeader
	}

	encodedCredentials := utils.TrimPrefixFold(authHeader, serverconst.AuthSchemeBasic)
	decodedCredentials, err := base64.StdEncoding.DecodeString(encodedCredentials)
	if err != nil {
		return "", "", errInvalidAuthorizationHeader
	}

	credentials := strings.SplitN(string(decodedCredentials), ":", 2)
	if len(credentials) != 2 {
		return "", "", errInvalidAuthorizationHeader
	}

	// URL-decode client credentials.
	clientID, idErr := url.QueryUnescape(credentials[0])
	if idErr != nil {
		return "", "", errInvalidAuthorizationHeader
	}
	clientSecret, secretErr := url.QueryUnescape(credentials[1])
	if secretErr != nil {
		return "", "", errInvalidAuthorizationHeader
	}

	return clientID, clientSecret, nil
}
