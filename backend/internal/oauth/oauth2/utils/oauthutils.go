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
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	authzmodel "github.com/asgardeo/thunder/internal/oauth/oauth2/authz/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	sessionmodel "github.com/asgardeo/thunder/internal/oauth/session/model"
	sessionstore "github.com/asgardeo/thunder/internal/oauth/session/store"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/google/uuid"
)

func GetOAuthMessage(request *http.Request, responseWriter http.ResponseWriter) (*authzmodel.OAuthMessage, error) {

	if request == nil || responseWriter == nil {
		return nil, errors.New("request or response writer is nil")
	}

	logger := log.GetLogger()

	// Parse the query parameters.
	if err := request.ParseForm(); err != nil {
		return nil, errors.New("failed to parse form data: " + err.Error())
	}

	// Check if the session data is already stored with a session data key.
	sessionDataKey := request.FormValue(constants.SESSION_DATA_KEY)
	var sessionData sessionmodel.SessionData
	if sessionDataKey != "" {
		sessionDataStore := sessionstore.GetSessionDataStore()
		var ok bool
		ok, sessionData = sessionDataStore.GetSession(sessionDataKey)
		if !ok {
			logger.Debug("Session data not found for session data key",
				log.String("sessionDataKey", sessionDataKey))
		}
	}

	// Determine the request type.
	// TODO: Add other required request types.
	var requestType string
	if sessionDataKey != "" && request.FormValue(constants.SESSION_DATA_KEY_CONSENT) == "" {
		requestType = constants.TYPE_AUTHORIZATION_RESPONSE_FROM_FRAMEWORK
	} else if request.FormValue(constants.CLIENT_ID) != "" && sessionDataKey == "" &&
		request.FormValue(constants.SESSION_DATA_KEY_CONSENT) == "" {
		requestType = constants.TYPE_INITIAL_AUTHORIZATION_REQUEST
	} else {
		return nil, errors.New("invalid request type")
	}

	// Extract headers.
	headers := make(map[string][]string)
	for name, values := range request.Header {
		if len(values) > 0 {
			headers[name] = append([]string{}, values...)
		}
	}

	// Extract query parameters.
	queryParams := make(map[string]string)
	for key, values := range request.URL.Query() {
		if len(values) > 0 {
			queryParams[key] = values[0]
		}
	}

	// Extract form/body parameters.
	bodyParams := make(map[string]string)
	for key, values := range request.PostForm {
		if len(values) > 0 {
			bodyParams[key] = values[0]
		}
	}

	return &authzmodel.OAuthMessage{
		RequestType:        requestType,
		SessionData:        &sessionData,
		RequestHeaders:     headers,
		RequestQueryParams: queryParams,
		RequestBodyParams:  bodyParams,
	}, nil
}

// GetUriWithQueryParams constructs a URI with the given query parameters.
func GetUriWithQueryParams(uri string, queryParams map[string]string) (string, error) {

	// Parse the URI.
	parsedUrl, err := url.Parse(uri)
	if err != nil {
		return "", errors.New("failed to parse the return URI: " + err.Error())
	}

	// Return the URI if there are no query parameters.
	if len(queryParams) == 0 {
		return parsedUrl.String(), nil
	}

	// Add the query parameters to the URI.
	query := parsedUrl.Query()
	for key, value := range queryParams {
		query.Add(key, value)
	}
	parsedUrl.RawQuery = query.Encode()

	// Return the constructed URI.
	return parsedUrl.String(), nil
}

// GetLoginPageRedirectUri returns the login page URL with the given query parameters.
func GetLoginPageRedirectUri(queryParams map[string]string) (string, error) {

	serverConfig := config.GetThunderRuntime().Config.Server
	loginPageUrl := (&url.URL{
		Scheme: "https",
		Host:   fmt.Sprintf("%s:%d", serverConfig.Hostname, serverConfig.Port),
		Path:   "login",
	}).String()

	return GetUriWithQueryParams(loginPageUrl, queryParams)
}

// GetErrorPageURL returns the server error page URL.
func GetErrorPageURL(queryParams map[string]string) (string, error) {

	serverConfig := config.GetThunderRuntime().Config.Server
	errorPageUrl := (&url.URL{
		Scheme: "https",
		Host:   fmt.Sprintf("%s:%d", serverConfig.Hostname, serverConfig.Port),
		Path:   "oauth2_error",
	}).String()

	return GetUriWithQueryParams(errorPageUrl, queryParams)
}

// RedirectToErrorPage redirects the user to the error page with the given error details.
func RedirectToErrorPage(responseWriter http.ResponseWriter, request *http.Request, errorCode, errorMessage string) {

	if responseWriter == nil || request == nil {
		log.GetLogger().Error("Response writer or request is nil. Cannot redirect to error page.")
		return
	}

	queryParams := map[string]string{
		constants.OAUTH_ERROR_CODE:    errorCode,
		constants.OAUTH_ERROR_MESSAGE: errorMessage,
	}
	redirectURL, err := GetErrorPageURL(queryParams)
	if err != nil {
		log.GetLogger().Error("Failed to construct error page URL: " + err.Error())
		return
	}
	log.GetLogger().Info("Redirecting to error page: " + redirectURL)

	// Redirect with the request object.
	http.Redirect(responseWriter, request, redirectURL, http.StatusFound)
}

// GenerateNewSessionDataKey generates and returns a session data key.
func GenerateNewSessionDataKey() string {

	return uuid.New().String()
}

func GetAllowedOrigin(allowedOrigins []string, redirectUri string) string {

	if len(allowedOrigins) == 0 {
		return ""
	}

	for _, allowedOrigin := range allowedOrigins {
		if strings.Contains(redirectUri, allowedOrigin) {
			return allowedOrigin
		}
	}

	return ""
}
