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

package token

import (
	"encoding/json"
	"net/http"

	"github.com/asgardeo/thunder/internal/system/server"
	"github.com/asgardeo/thunder/internal/system/utils"

	appprovider "github.com/asgardeo/thunder/internal/application/provider"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/granthandlers"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/model"
	oauthutils "github.com/asgardeo/thunder/internal/oauth/oauth2/utils"
	scopeprovider "github.com/asgardeo/thunder/internal/oauth/scope/provider"
	"github.com/asgardeo/thunder/internal/system/log"
)

type TokenHandler struct{}

// HandleTokenRequest handles the token request for OAuth 2.0.
// It validates the client credentials and delegates to the appropriate grant handler.
func (th *TokenHandler) HandleTokenRequest(w http.ResponseWriter, r *http.Request) {

	logger := log.GetLogger().With(log.String(log.LOGGER_KEY_COMPONENT_NAME, "TokenHandler"))

	// Parse the form data from the request body.
	if err := r.ParseForm(); err != nil {
		utils.WriteJSONError(w, constants.ERROR_INVALID_REQUEST,
			"Failed to parse request body", http.StatusBadRequest, nil)
		return
	}

	// Validate the grant_type.
	grantType := r.FormValue(constants.GRANT_TYPE)
	if grantType == "" {
		utils.WriteJSONError(w, constants.ERROR_INVALID_REQUEST,
			"Missing grant_type parameter", http.StatusBadRequest, nil)
		return
	}

	var grantHandler granthandlers.GrantHandler
	switch grantType {
	case constants.GRANT_TYPE_CLIENT_CREDENTIALS:
		grantHandler = &granthandlers.ClientCredentialsGrantHandler{}
	case constants.GRANT_TYPE_AUTHORIZATION_CODE:
		grantHandler = &granthandlers.AuthorizationCodeGrantHandler{}
	default:
		utils.WriteJSONError(w, constants.ERROR_UNSUPPORTED_GRANT_TYPE,
			"Unsupported grant type", http.StatusBadRequest, nil)
		return
	}

	// Extract client credentials from the request.
	clientId := ""
	clientSecret := ""
	if r.Header.Get("Authorization") != "" {
		var err error
		clientId, clientSecret, err = utils.ExtractBasicAuthCredentials(r)
		if err != nil {
			if err.Error() == "invalid authorization header" {
				responseHeaders := []map[string]string{
					{"WWW-Authenticate": "Basic"},
				}
				utils.WriteJSONError(w, constants.ERROR_INVALID_CLIENT,
					"Invalid client credentials", http.StatusUnauthorized, responseHeaders)
				return
			}
			utils.WriteJSONError(w, constants.ERROR_INVALID_CLIENT,
				"Invalid client credentials", http.StatusUnauthorized, nil)
			return
		}
	}

	// Check for client credentials in the request body.
	clientIdFromBody := r.FormValue(constants.CLIENT_ID)
	clientSecretFromBody := r.FormValue(constants.CLIENT_SECRET)

	if clientIdFromBody != "" && clientSecretFromBody != "" {
		if clientId != "" && clientSecret != "" {
			utils.WriteJSONError(w, constants.ERROR_INVALID_REQUEST,
				"Authorization information is provided in both header and body", http.StatusBadRequest, nil)
			return
		}

		clientId = clientIdFromBody
		clientSecret = clientSecretFromBody
	} else {
		if clientId == "" {
			clientId = clientIdFromBody
		}
		if clientSecret == "" {
			clientSecret = clientSecretFromBody
		}
	}

	// Construct the token request.
	tokenRequest := &model.TokenRequest{
		GrantType:    grantType,
		ClientId:     clientId,
		ClientSecret: clientSecret,
		Scope:        r.FormValue("scope"),
		Username:     r.FormValue("username"),
		Password:     r.FormValue("password"),
		RefreshToken: r.FormValue("refresh_token"),
		CodeVerifier: r.FormValue("code_verifier"),
		Code:         r.FormValue("code"),
		RedirectUri:  r.FormValue("redirect_uri"),
	}

	// Validate the token request.
	tokenError := grantHandler.ValidateGrant(tokenRequest)
	if tokenError != nil && tokenError.Error != "" {
		utils.WriteJSONError(w, tokenError.Error, tokenError.ErrorDescription, http.StatusBadRequest, nil)
		return
	}

	// Retrieve the OAuth application based on the client id.
	appProvider := appprovider.NewApplicationProvider()
	appService := appProvider.GetApplicationService()

	oauthApp, err := appService.GetOAuthApplication(clientId)
	if err != nil || oauthApp == nil {
		utils.WriteJSONError(w, constants.ERROR_INVALID_CLIENT,
			"Invalid client credentials", http.StatusUnauthorized, nil)
		return
	}

	// Validate grant type against the application.
	if !oauthApp.IsAllowedGrantType(tokenRequest.GrantType) {
		utils.WriteJSONError(w, constants.ERROR_UNAUTHORIZED_CLIENT,
			"The authenticated client is not authorized to use this grant type", http.StatusUnauthorized, nil)
		return
	}

	// Validate and filter scopes.
	scopeValidatorProvider := scopeprovider.NewScopeValidatorProvider()
	scopeValidator := scopeValidatorProvider.GetScopeValidator()

	validScopes, scopeError := scopeValidator.ValidateScopes(tokenRequest.Scope, oauthApp.ClientId)
	if scopeError != nil {
		utils.WriteJSONError(w, scopeError.Error, scopeError.ErrorDescription, http.StatusBadRequest, nil)
		return
	}
	tokenRequest.Scope = validScopes

	// Delegate to the grant handler.
	tokenResponse, tokenError := grantHandler.HandleGrant(tokenRequest, oauthApp)
	if tokenError != nil && tokenError.Error != "" {
		utils.WriteJSONError(w, tokenError.Error, tokenError.ErrorDescription, http.StatusBadRequest, nil)
		return
	}

	// Log successful token generation.
	logger.Info("Token generated successfully", log.String("client_id", clientId))

	// Set the response headers.
	w.Header().Set("Content-Type", "application/json")
	// Must include the following headers when sensitive data is returned.
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")

	// Retrieve allowed origins from the server configuration.
	allowedOrigins, err := server.GetAllowedOrigins()
	if err != nil {
		logger.Error("Failed to get allowed origins", log.Error(err))
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
	}

	// Set the CORS headers if allowed origins are configured.
	allowedOrigin := oauthutils.GetAllowedOrigin(allowedOrigins, tokenRequest.RedirectUri)
	if allowedOrigin != "" {
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	}

	// Write the token response.
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tokenResponse); err != nil {
		logger.Error("Failed to write token response", log.Error(err))
		http.Error(w, "Failed to write token response", http.StatusInternalServerError)
		return
	}
	// Log the token response.
	logger.Info("Token response sent", log.String("client_id", clientId))
}
