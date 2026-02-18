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
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/asgardeo/thunder/internal/application"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/clientauth"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/granthandlers"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/model"
	"github.com/asgardeo/thunder/internal/oauth/scope"
	sysContext "github.com/asgardeo/thunder/internal/system/context"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/observability"
	"github.com/asgardeo/thunder/internal/system/observability/event"
	"github.com/asgardeo/thunder/internal/system/utils"
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
	observabilitySvc     observability.ObservabilityServiceInterface
}

// newTokenHandler creates a new instance of tokenHandler.
func newTokenHandler(
	appService application.ApplicationServiceInterface,
	grantHandlerProvider granthandlers.GrantHandlerProviderInterface,
	scopeValidator scope.ScopeValidatorInterface,
	observabilitySvc observability.ObservabilityServiceInterface,
) TokenHandlerInterface {
	return &tokenHandler{
		appService:           appService,
		grantHandlerProvider: grantHandlerProvider,
		scopeValidator:       scopeValidator,
		observabilitySvc:     observabilitySvc,
	}
}

// HandleTokenRequest handles the token request for OAuth 2.0.
// It validates the client credentials and delegates to the appropriate grant handler.
func (th *tokenHandler) HandleTokenRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "TokenHandler"))

	startTime := time.Now().UnixMilli()

	// Parse the form data from the request body.
	if err := r.ParseForm(); err != nil {
		th.publishTokenIssuanceFailedEvent(r.Context(), "", "", "", http.StatusBadRequest, err.Error(), startTime)
		utils.WriteJSONError(w, constants.ErrorInvalidRequest,
			"Failed to parse request body", http.StatusBadRequest, nil)
		return
	}

	// Validate the grant_type.
	grantTypeStr := r.FormValue(constants.RequestParamGrantType)
	scopeStr := r.FormValue("scope")
	clientID := ""

	// Get authenticated client from context
	clientInfo := clientauth.GetOAuthClient(r.Context())
	if clientInfo != nil {
		clientID = clientInfo.ClientID
	}

	th.publishTokenIssuanceStartedEvent(r.Context(), clientID, grantTypeStr, scopeStr)

	if grantTypeStr == "" {
		th.publishTokenIssuanceFailedEvent(r.Context(), clientID, grantTypeStr, scopeStr, http.StatusBadRequest,
			"Missing grant_type parameter", startTime)
		utils.WriteJSONError(w, constants.ErrorInvalidRequest,
			"Missing grant_type parameter", http.StatusBadRequest, nil)
		return
	}
	grantType := constants.GrantType(grantTypeStr)
	if !grantType.IsValid() {
		th.publishTokenIssuanceFailedEvent(r.Context(), clientID, grantTypeStr, scopeStr, http.StatusBadRequest,
			"Invalid grant_type parameter", startTime)
		utils.WriteJSONError(w, constants.ErrorUnsupportedGrantType,
			"Invalid grant_type parameter", http.StatusBadRequest, nil)
		return
	}

	grantHandler, handlerErr := th.grantHandlerProvider.GetGrantHandler(grantType)
	if handlerErr != nil {
		if errors.Is(handlerErr, constants.UnSupportedGrantTypeError) {
			th.publishTokenIssuanceFailedEvent(r.Context(), clientID, grantTypeStr, scopeStr, http.StatusBadRequest,
				"Unsupported grant type", startTime)
			utils.WriteJSONError(w, constants.ErrorUnsupportedGrantType, "Unsupported grant type",
				http.StatusBadRequest, nil)
			return
		}
		logger.Error("Failed to get grant handler", log.Error(handlerErr))
		th.publishTokenIssuanceFailedEvent(r.Context(), clientID, grantTypeStr, scopeStr,
			http.StatusInternalServerError, "Failed to get grant handler", startTime)
		utils.WriteJSONError(w, constants.ErrorServerError,
			"Failed to get grant handler", http.StatusInternalServerError, nil)
		return
	}

	if clientInfo == nil {
		logger.Error("OAuth client not found in context - ClientAuthMiddleware must be applied")
		th.publishTokenIssuanceFailedEvent(r.Context(), clientID, grantTypeStr, scopeStr,
			http.StatusInternalServerError, "Authentication context not available", startTime)
		utils.WriteJSONError(w, constants.ErrorServerError,
			"Internal server error: authentication context not available", http.StatusInternalServerError, nil)
		return
	}

	oauthApp := clientInfo.OAuthApp
	clientSecret := clientInfo.ClientSecret

	// Validate grant type against the application.
	if !oauthApp.IsAllowedGrantType(grantType) {
		th.publishTokenIssuanceFailedEvent(r.Context(), clientID, grantTypeStr, scopeStr, http.StatusUnauthorized,
			"Client not authorized for grant type", startTime)
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
		th.publishTokenIssuanceFailedEvent(r.Context(), clientID, grantTypeStr, scopeStr, http.StatusBadRequest,
			tokenError.ErrorDescription, startTime)
		utils.WriteJSONError(w, tokenError.Error, tokenError.ErrorDescription, http.StatusBadRequest, nil)
		return
	}

	// Validate and filter scopes.
	validScopes, scopeError := th.scopeValidator.ValidateScopes(tokenRequest.Scope, oauthApp.ClientID)
	if scopeError != nil {
		th.publishTokenIssuanceFailedEvent(r.Context(), clientID, grantTypeStr, scopeStr, http.StatusBadRequest,
			scopeError.ErrorDescription, startTime)
		utils.WriteJSONError(w, scopeError.Error, scopeError.ErrorDescription, http.StatusBadRequest, nil)
		return
	}
	tokenRequest.Scope = validScopes

	// Delegate to the grant handler.
	tokenRespDTO, tokenError := grantHandler.HandleGrant(tokenRequest, oauthApp)
	if tokenError != nil && tokenError.Error != "" {
		code := http.StatusBadRequest
		if tokenError.Error == constants.ErrorServerError {
			code = http.StatusInternalServerError
		}
		th.publishTokenIssuanceFailedEvent(r.Context(), clientID, grantTypeStr, scopeStr, code,
			tokenError.ErrorDescription, startTime)
		utils.WriteJSONError(w, tokenError.Error, tokenError.ErrorDescription, code, nil)
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
			th.publishTokenIssuanceFailedEvent(r.Context(), clientID, grantTypeStr, scopeStr,
				http.StatusInternalServerError, "Failed to get refresh grant handler", startTime)
			utils.WriteJSONError(w, constants.ErrorServerError,
				"Failed to get refresh grant handler", http.StatusInternalServerError, nil)
			return
		}
		refreshGrantHandlerTyped, ok := refreshGrantHandler.(granthandlers.RefreshTokenGrantHandlerInterface)
		if !ok {
			logger.Error("Failed to cast refresh grant handler", log.String("client_id", clientID),
				log.String("grant_type", grantTypeStr))
			th.publishTokenIssuanceFailedEvent(r.Context(), clientID, grantTypeStr, scopeStr,
				http.StatusInternalServerError, "Internal Server Error", startTime)
			utils.WriteJSONError(w, constants.ErrorServerError, "Something went wrong",
				http.StatusInternalServerError, nil)
			return
		}

		refreshTokenError := refreshGrantHandlerTyped.IssueRefreshToken(tokenRespDTO, oauthApp,
			tokenRespDTO.AccessToken.Subject, tokenRespDTO.AccessToken.Audience,
			grantTypeStr, tokenRespDTO.AccessToken.Scopes, tokenRespDTO.AccessToken.ClaimsRequest,
			tokenRespDTO.AccessToken.ClaimsLocales)
		if refreshTokenError != nil && refreshTokenError.Error != "" {
			th.publishTokenIssuanceFailedEvent(r.Context(), clientID, grantTypeStr, scopeStr,
				http.StatusInternalServerError, refreshTokenError.ErrorDescription, startTime)
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

	// For token exchange, determine the issued_token_type from the request
	if grantType == constants.GrantTypeTokenExchange {
		requestedTokenType := tokenRequest.RequestedTokenType
		if requestedTokenType == "" || requestedTokenType == string(constants.TokenTypeIdentifierAccessToken) {
			tokenResponse.IssuedTokenType = string(constants.TokenTypeIdentifierAccessToken)
		} else {
			tokenResponse.IssuedTokenType = string(constants.TokenTypeIdentifierJWT)
		}
	}

	logger.Debug("Token generated successfully", log.String("client_id", clientID),
		log.String("grant_type", grantTypeStr))

	// Set the response headers.
	// Must include the following headers when sensitive data is returned.
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")

	// Write the token response.
	utils.WriteSuccessResponse(w, http.StatusOK, tokenResponse)
	logger.Debug("Token response sent", log.String("client_id", clientID), log.String("grant_type", grantTypeStr))

	th.publishTokenIssuedEvent(r.Context(), clientID, grantTypeStr, scopes, startTime)
}

func (th *tokenHandler) publishTokenIssuanceStartedEvent(ctx context.Context, clientID, grantType, scope string) {
	if th.observabilitySvc == nil || !th.observabilitySvc.IsEnabled() {
		return
	}

	evt := event.NewEvent(
		sysContext.GetTraceID(ctx),
		string(event.EventTypeTokenIssuanceStarted),
		event.ComponentAuthHandler,
	).
		WithStatus(event.StatusInProgress).
		WithData(event.DataKey.ClientID, clientID).
		WithData(event.DataKey.GrantType, grantType).
		WithData(event.DataKey.Scope, scope)

	th.observabilitySvc.PublishEvent(evt)
}

func (th *tokenHandler) publishTokenIssuedEvent(
	ctx context.Context, clientID, grantType, scope string, startTime int64,
) {
	if th.observabilitySvc == nil || !th.observabilitySvc.IsEnabled() {
		return
	}

	duration := time.Now().UnixMilli() - startTime

	evt := event.NewEvent(
		sysContext.GetTraceID(ctx),
		string(event.EventTypeTokenIssued),
		event.ComponentAuthHandler,
	).
		WithStatus(event.StatusSuccess).
		WithData(event.DataKey.ClientID, clientID).
		WithData(event.DataKey.GrantType, grantType).
		WithData(event.DataKey.Scope, scope).
		WithData(event.DataKey.DurationMs, fmt.Sprintf("%d", duration))

	th.observabilitySvc.PublishEvent(evt)
}

func (th *tokenHandler) publishTokenIssuanceFailedEvent(
	ctx context.Context, clientID, grantType, scope string, statusCode int, message string, startTime int64,
) {
	if th.observabilitySvc == nil || !th.observabilitySvc.IsEnabled() {
		return
	}

	duration := time.Now().UnixMilli() - startTime

	// Classify error type based on status code
	errorType := "client_error"
	if statusCode >= 500 {
		errorType = "server_error"
	}

	evt := event.NewEvent(
		sysContext.GetTraceID(ctx),
		string(event.EventTypeTokenIssuanceFailed),
		event.ComponentAuthHandler,
	).
		WithStatus(event.StatusFailure).
		WithData(event.DataKey.ClientID, clientID).
		WithData(event.DataKey.GrantType, grantType).
		WithData(event.DataKey.Scope, scope).
		WithData(event.DataKey.Error, message).
		WithData(event.DataKey.ErrorCode, fmt.Sprintf("%d", statusCode)).
		WithData(event.DataKey.ErrorType, errorType).
		WithData(event.DataKey.DurationMs, fmt.Sprintf("%d", duration))

	th.observabilitySvc.PublishEvent(evt)
}
