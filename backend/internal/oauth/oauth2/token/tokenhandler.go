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

// Package token provides handler for managing OAuth 2.0 token requests.
package token

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/asgardeo/thunder/internal/application"
	"github.com/asgardeo/thunder/internal/system/hash"
	"github.com/asgardeo/thunder/internal/system/utils"

	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/granthandlers"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/model"
	"github.com/asgardeo/thunder/internal/oauth/scope"
	"github.com/asgardeo/thunder/internal/system/log"
)

// TokenHandlerInterface defines the interface for handling OAuth 2.0 token requests.
type TokenHandlerInterface interface {
	HandleTokenRequest(w http.ResponseWriter, r *http.Request)
}

// tokenHandler implements the TokenHandlerInterface.
type tokenHandler struct {
	appService           application.ApplicationServiceInterface
	grantHandlerProvider granthandlers.GrantHandlerProviderInterface
	scopeValidator       scope.ScopeValidatorInterface
}

// newTokenHandler creates a new instance of tokenHandler.
func newTokenHandler(
	appService application.ApplicationServiceInterface,
	grantHandlerProvider granthandlers.GrantHandlerProviderInterface,
	scopeValidator scope.ScopeValidatorInterface,
) TokenHandlerInterface {
	return &tokenHandler{
		appService:           appService,
		grantHandlerProvider: grantHandlerProvider,
		scopeValidator:       scopeValidator,
	}
}

// HandleTokenRequest handles the token request for OAuth 2.0.
// It validates the client credentials and delegates to the appropriate grant handler.
func (th *tokenHandler) HandleTokenRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "TokenHandler"))

	// Parse the form data from the request body.
	if err := r.ParseForm(); err != nil {
		utils.WriteJSONError(w, constants.ErrorInvalidRequest,
			"Failed to parse request body", http.StatusBadRequest, nil)
		return
	}

	// Validate the grant_type.
	grantTypeStr := r.FormValue(constants.RequestParamGrantType)
	if grantTypeStr == "" {
		utils.WriteJSONError(w, constants.ErrorInvalidRequest,
			"Missing grant_type parameter", http.StatusBadRequest, nil)
		return
	}
	grantType := constants.GrantType(grantTypeStr)
	if !grantType.IsValid() {
		utils.WriteJSONError(w, constants.ErrorUnsupportedGrantType,
			"Invalid grant_type parameter", http.StatusBadRequest, nil)
		return
	}

	grantHandler, handlerErr := th.grantHandlerProvider.GetGrantHandler(grantType)
	if handlerErr != nil {
		if errors.Is(handlerErr, constants.UnSupportedGrantTypeError) {
			utils.WriteJSONError(w, constants.ErrorUnsupportedGrantType, "Unsupported grant type",
				http.StatusBadRequest, nil)
			return
		}
		logger.Error("Failed to get grant handler", log.Error(handlerErr))
		utils.WriteJSONError(w, constants.ErrorServerError,
			"Failed to get grant handler", http.StatusInternalServerError, nil)
		return
	}

	clientID, clientSecret, tokenAuthMethod, ok := extractClientIDAndSecret(r, w)
	if !ok {
		return
	}

	// Retrieve the OAuth application based on the client id.
	oauthApp, err := th.appService.GetOAuthApplication(clientID)
	if err != nil || oauthApp == nil {
		utils.WriteJSONError(w, constants.ErrorInvalidClient,
			"Invalid client credentials", http.StatusUnauthorized, nil)
		return
	}

	// Validate the token endpoint authentication method.
	if !oauthApp.IsAllowedTokenEndpointAuthMethod(tokenAuthMethod) {
		if tokenAuthMethod == constants.TokenEndpointAuthMethodNone {
			utils.WriteJSONError(w, constants.ErrorInvalidClient,
				"Missing client_secret parameter",
				http.StatusUnauthorized, nil)
			return
		}
		utils.WriteJSONError(w, constants.ErrorUnauthorizedClient,
			"Client is not allowed to use the specified token endpoint authentication method",
			http.StatusUnauthorized, nil)
		return
	}

	// Validate the client credentials.
	hashedClientSecret := hash.GenerateThumbprintFromString(clientSecret)
	if tokenAuthMethod != constants.TokenEndpointAuthMethodNone {
		if clientID != oauthApp.ClientID || hashedClientSecret != oauthApp.HashedClientSecret {
			utils.WriteJSONError(w, constants.ErrorInvalidClient,
				"Invalid client credentials", http.StatusUnauthorized, nil)
			return
		}
	}

	// Validate grant type against the application.
	if !oauthApp.IsAllowedGrantType(grantType) {
		utils.WriteJSONError(w, constants.ErrorUnauthorizedClient,
			"The client is not authorized to use this grant type", http.StatusUnauthorized, nil)
		return
	}

	// Construct the token request.
	tokenRequest := &model.TokenRequest{
		GrantType:          grantTypeStr,
		ClientID:           clientID,
		ClientSecret:       clientSecret,
		Scope:              r.FormValue("scope"),
		Username:           r.FormValue("username"),
		Password:           r.FormValue("password"),
		RefreshToken:       r.FormValue("refresh_token"),
		CodeVerifier:       r.FormValue("code_verifier"),
		Code:               r.FormValue("code"),
		RedirectURI:        r.FormValue("redirect_uri"),
		Resource:           r.FormValue(constants.RequestParamResource),
		SubjectToken:       r.FormValue(constants.RequestParamSubjectToken),
		SubjectTokenType:   r.FormValue(constants.RequestParamSubjectTokenType),
		ActorToken:         r.FormValue(constants.RequestParamActorToken),
		ActorTokenType:     r.FormValue(constants.RequestParamActorTokenType),
		RequestedTokenType: r.FormValue(constants.RequestParamRequestedTokenType),
		Audience:           r.FormValue(constants.RequestParamAudience),
	}

	// Validate the token request.
	tokenError := grantHandler.ValidateGrant(tokenRequest, oauthApp)
	if tokenError != nil && tokenError.Error != "" {
		utils.WriteJSONError(w, tokenError.Error, tokenError.ErrorDescription, http.StatusBadRequest, nil)
		return
	}

	// Validate and filter scopes.
	validScopes, scopeError := th.scopeValidator.ValidateScopes(tokenRequest.Scope, oauthApp.ClientID)
	if scopeError != nil {
		utils.WriteJSONError(w, scopeError.Error, scopeError.ErrorDescription, http.StatusBadRequest, nil)
		return
	}
	tokenRequest.Scope = validScopes

	// Delegate to the grant handler.
	ctx := &model.TokenContext{
		TokenAttributes: make(map[string]interface{}),
	}
	tokenRespDTO, tokenError := grantHandler.HandleGrant(tokenRequest, oauthApp, ctx)
	if tokenError != nil && tokenError.Error != "" {
		utils.WriteJSONError(w, tokenError.Error, tokenError.ErrorDescription, http.StatusBadRequest, nil)
		return
	}

	// Generate and add refresh token if applicable.
	if grantType == constants.GrantTypeAuthorizationCode &&
		oauthApp.IsAllowedGrantType(constants.GrantTypeRefreshToken) {
		logger.Debug("Issuing refresh token for the token request", log.String("client_id", clientID),
			log.String("grant_type", grantTypeStr))

		refreshGrantHandler, handlerErr := th.grantHandlerProvider.GetGrantHandler(constants.GrantTypeRefreshToken)
		if handlerErr != nil {
			logger.Error("Failed to get refresh grant handler", log.Error(handlerErr))
			utils.WriteJSONError(w, constants.ErrorServerError,
				"Failed to get refresh grant handler", http.StatusInternalServerError, nil)
			return
		}
		refreshGrantHandlerTyped, ok := refreshGrantHandler.(granthandlers.RefreshTokenGrantHandlerInterface)
		if !ok {
			logger.Error("Failed to cast refresh grant handler", log.String("client_id", clientID),
				log.String("grant_type", grantTypeStr))
			utils.WriteJSONError(w, constants.ErrorServerError, "Something went wrong",
				http.StatusInternalServerError, nil)
			return
		}

		refreshTokenError := refreshGrantHandlerTyped.IssueRefreshToken(tokenRespDTO, oauthApp, ctx,
			grantTypeStr, tokenRespDTO.AccessToken.Scopes)
		if refreshTokenError != nil && refreshTokenError.Error != "" {
			utils.WriteJSONError(w, refreshTokenError.Error, refreshTokenError.ErrorDescription,
				http.StatusInternalServerError, nil)
			return
		}
	}

	scopes := strings.Join(tokenRespDTO.AccessToken.Scopes, " ")
	tokenResponse := &model.TokenResponse{
		AccessToken:  tokenRespDTO.AccessToken.Token,
		TokenType:    tokenRespDTO.AccessToken.TokenType,
		ExpiresIn:    tokenRespDTO.AccessToken.ExpiresIn,
		RefreshToken: tokenRespDTO.RefreshToken.Token,
		Scope:        scopes,
		IDToken:      tokenRespDTO.IDToken.Token,
	}

	// For token exchange, set the issued_token_type from context
	if grantType == constants.GrantTypeTokenExchange {
		if val, ok := ctx.TokenAttributes["issued_token_type"].(string); ok {
			tokenResponse.IssuedTokenType = val
		}
	}

	logger.Debug("Token generated successfully", log.String("client_id", clientID),
		log.String("grant_type", grantTypeStr))

	// Set the response headers.
	w.Header().Set("Content-Type", "application/json")
	// Must include the following headers when sensitive data is returned.
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")

	// Write the token response.
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tokenResponse); err != nil {
		logger.Error("Failed to write token response", log.Error(err))
		http.Error(w, "Failed to write token response", http.StatusInternalServerError)
		return
	}
	logger.Debug("Token response sent", log.String("client_id", clientID), log.String("grant_type", grantTypeStr))
}

// extractClientIDAndSecret extracts the client ID and secret from the request.
// It returns the client ID, client secret, token authentication method, and a boolean indicating success.
func extractClientIDAndSecret(r *http.Request, w http.ResponseWriter) (
	string, string, constants.TokenEndpointAuthMethod, bool) {
	var clientID string
	var clientSecret string
	var tokenAuthMethod constants.TokenEndpointAuthMethod

	if r.Header.Get("Authorization") != "" {
		var err error
		clientID, clientSecret, err = utils.ExtractBasicAuthCredentials(r)
		if err != nil {
			if err.Error() == "invalid authorization header" {
				responseHeaders := []map[string]string{
					{"WWW-Authenticate": "Basic"},
				}
				utils.WriteJSONError(w, constants.ErrorInvalidClient,
					"Invalid client credentials", http.StatusUnauthorized, responseHeaders)
				return "", "", "", false
			}
			utils.WriteJSONError(w, constants.ErrorInvalidClient,
				"Invalid client credentials", http.StatusUnauthorized, nil)
			return "", "", "", false
		}
	}

	// Check for client credentials in the request body.
	clientIDFromBody := r.FormValue(constants.RequestParamClientID)
	clientSecretFromBody := r.FormValue(constants.RequestParamClientSecret)

	if (clientID != "" || clientSecret != "") && (clientIDFromBody != "" || clientSecretFromBody != "") {
		utils.WriteJSONError(w, constants.ErrorInvalidRequest,
			"Authorization information is provided in both header and body", http.StatusBadRequest, nil)
		return "", "", "", false
	}

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

	if clientID == "" {
		utils.WriteJSONError(w, constants.ErrorInvalidClient, "Missing client_id parameter",
			http.StatusUnauthorized, nil)
		return "", "", "", false
	}

	if clientSecret == "" && tokenAuthMethod != constants.TokenEndpointAuthMethodNone {
		utils.WriteJSONError(w, constants.ErrorInvalidClient, "Missing client_secret parameter",
			http.StatusUnauthorized, nil)
		return "", "", "", false
	}

	return clientID, clientSecret, tokenAuthMethod, true
}
