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

package authz

import (
	"net/http"
	"time"

	"github.com/asgardeo/thunder/internal/system/utils"

	appprovider "github.com/asgardeo/thunder/internal/application/provider"
	authzmodel "github.com/asgardeo/thunder/internal/oauth/oauth2/authz/model"
	authzutils "github.com/asgardeo/thunder/internal/oauth/oauth2/authz/utils"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/model"
	oauthutils "github.com/asgardeo/thunder/internal/oauth/oauth2/utils"
	sessionmodel "github.com/asgardeo/thunder/internal/oauth/session/model"
	sessionstore "github.com/asgardeo/thunder/internal/oauth/session/store"
	"github.com/asgardeo/thunder/internal/system/log"
)

// AuthorizeHandlerInterface defines the interface for handling OAuth2 authorization requests.
type AuthorizeHandlerInterface interface {
	HandleAuthorizeRequest(w http.ResponseWriter, r *http.Request)
}

// AuthorizeHandler implements the AuthorizeHandlerInterface for handling OAuth2 authorization requests.
type AuthorizeHandler struct {
	authValidator AuthorizationValidatorInterface
}

// NewAuthorizeHandler creates a new instance of AuthorizeHandler.
func NewAuthorizeHandler() AuthorizeHandlerInterface {
	return &AuthorizeHandler{
		authValidator: NewAuthorizationValidator(),
	}
}

// HandleAuthorizeRequest handles the OAuth2 authorization request.
func (ah *AuthorizeHandler) HandleAuthorizeRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger()

	// Construct the OAuthMessage.
	oAuthMessage, err := oauthutils.GetOAuthMessage(r, w)
	if err != nil {
		logger.Error("Failed to construct OAuthMessage", log.Error(err))
		utils.WriteJSONError(w, constants.ERROR_INVALID_REQUEST,
			"Invalid authorization request", http.StatusBadRequest, nil)
		return
	}
	if oAuthMessage == nil {
		logger.Error("OAuthMessage is nil")
		utils.WriteJSONError(w, constants.ERROR_INVALID_REQUEST,
			"Invalid authorization request", http.StatusBadRequest, nil)
		return
	}

	switch oAuthMessage.RequestType {
	case constants.TYPE_INITIAL_AUTHORIZATION_REQUEST:
		ah.handleInitialAuthorizationRequest(oAuthMessage, w, r)
	case constants.TYPE_AUTHORIZATION_RESPONSE_FROM_FRAMEWORK:
		ah.handleAuthenticationResponse(oAuthMessage, w, r)
	case constants.TYPE_CONSENT_RESPONSE_FROM_USER:
	// TODO: Handle the consent response from the user.
	//  Verify whether we need separate session data key for consent flow.
	//  Alternatively could add consent info also to the same session object.
	default:
		// Handle the case where the request is not recognized.
		utils.WriteJSONError(w, constants.ERROR_INVALID_REQUEST,
			"Invalid authorization request", http.StatusBadRequest, nil)
	}
}

// handleInitialAuthorizationRequest handles the initial authorization request from the client.
func (ah *AuthorizeHandler) handleInitialAuthorizationRequest(msg *authzmodel.OAuthMessage,
	w http.ResponseWriter, r *http.Request) {
	// Extract required parameters.
	clientId := msg.RequestQueryParams[constants.CLIENT_ID]
	redirectUri := msg.RequestQueryParams[constants.REDIRECT_URI]
	scope := msg.RequestQueryParams[constants.SCOPE]
	state := msg.RequestQueryParams[constants.STATE]
	responseType := msg.RequestQueryParams[constants.RESPONSE_TYPE]

	if clientId == "" {
		oauthutils.RedirectToErrorPage(w, r, constants.ERROR_INVALID_REQUEST,
			"Missing client_id parameter")
		return
	}

	// Retrieve the OAuth application based on the client Id.
	appProvider := appprovider.NewApplicationProvider()
	appService := appProvider.GetApplicationService()

	app, err := appService.GetOAuthApplication(clientId)
	if err != nil || app == nil {
		oauthutils.RedirectToErrorPage(w, r, constants.ERROR_INVALID_CLIENT, "Invalid client_id")
		return
	}

	// Validate the authorization request.
	sendErrorToApp, errorCode, errorMessage := ah.authValidator.validateInitialAuthorizationRequest(msg, app)
	if errorCode != "" {
		if sendErrorToApp && redirectUri != "" {
			// Redirect to the redirect URI with an error.
			redirectUri, err := oauthutils.GetUriWithQueryParams(redirectUri, map[string]string{
				constants.ERROR:             errorCode,
				constants.ERROR_DESCRIPTION: errorMessage,
			})
			if err != nil {
				oauthutils.RedirectToErrorPage(w, r, constants.ERROR_SERVER_ERROR,
					"Failed to redirect to login page")
				return
			}

			if state != "" {
				redirectUri += "&" + constants.STATE + "=" + state
			}
			http.Redirect(w, r, redirectUri, http.StatusFound)
			return
		} else {
			oauthutils.RedirectToErrorPage(w, r, errorCode, errorMessage)
			return
		}
	}

	// Get query params sent in the request.
	queryParams := msg.RequestQueryParams

	// Construct session data.
	oauthParams := model.OAuthParameters{
		SessionDataKey: oauthutils.GenerateNewSessionDataKey(),
		State:          state,
		ClientId:       clientId,
		RedirectUri:    redirectUri,
		ResponseType:   responseType,
		Scopes:         scope,
	}

	// Set the redirect URI if not provided in the request. Invalid cases are already handled at this point.
	// TODO: This should be removed when supporting other means of authorization.
	if redirectUri == "" {
		oauthParams.RedirectUri = app.RedirectURIs[0]
	}

	sessionData := sessionmodel.SessionData{
		OAuthParameters: oauthParams,
		AuthTime:        time.Now(),
	}

	// Store session data in the session store.
	sessionDataStore := sessionstore.GetSessionDataStore()
	sessionDataStore.AddSession(oauthParams.SessionDataKey, sessionData)

	// Add other required query parameters.
	queryParams[constants.SESSION_DATA_KEY] = oauthParams.SessionDataKey

	// Append required query parameters to the redirect URI.
	loginPageUri, err := oauthutils.GetLoginPageRedirectUri(queryParams)
	if err != nil {
		oauthutils.RedirectToErrorPage(w, r, constants.ERROR_SERVER_ERROR,
			"Failed to redirect to login page")
	} else {
		// Redirect user-agent to the login page.
		http.Redirect(w, r, loginPageUri, http.StatusFound)
	}
}

func (ah *AuthorizeHandler) handleAuthenticationResponse(msg *authzmodel.OAuthMessage,
	w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger()

	// Validate the session data.
	sessionData := msg.SessionData
	if sessionData == nil {
		oauthutils.RedirectToErrorPage(w, r, constants.ERROR_INVALID_REQUEST,
			"Invalid authorization request")
		return
	}

	// If the user is not authenticated, redirect to the redirect URI with an error.
	authResult := sessionData.LoggedInUser
	if !authResult.IsAuthenticated {
		redirectUri := sessionData.OAuthParameters.RedirectUri
		if redirectUri == "" {
			logger.Error("Redirect URI is empty")
			oauthutils.RedirectToErrorPage(w, r, constants.ERROR_INVALID_REQUEST,
				"Invalid redirect URI")
			return
		}

		queryParams := map[string]string{
			constants.ERROR:             constants.ERROR_ACCESS_DENIED,
			constants.ERROR_DESCRIPTION: "User authentication failed",
		}
		if sessionData.OAuthParameters.State != "" {
			queryParams[constants.STATE] = sessionData.OAuthParameters.State
		}

		redirectUri, err := oauthutils.GetUriWithQueryParams(redirectUri, queryParams)
		if err != nil {
			logger.Error("Failed to construct redirect URI", log.Error(err))
			oauthutils.RedirectToErrorPage(w, r, constants.ERROR_SERVER_ERROR,
				"Failed to redirect to login page")
			return
		}

		http.Redirect(w, r, redirectUri, http.StatusFound)
	}

	// TODO: Do user authorization.
	//  Should validate for the scopes as well.

	// Generate the authorization code.
	authzCode, err := authzutils.GetAuthorizationCode(msg)
	if err != nil {
		logger.Error("Failed to generate authorization code", log.Error(err))
		oauthutils.RedirectToErrorPage(w, r, constants.ERROR_SERVER_ERROR,
			"Failed to generate authorization code")
		return
	}

	// Persist the authorization code.
	persistErr := InsertAuthorizationCode(authzCode)
	if persistErr != nil {
		logger.Error("Failed to persist authorization code", log.Error(persistErr))
		oauthutils.RedirectToErrorPage(w, r, constants.ERROR_SERVER_ERROR,
			"Failed to generate authorization code")
		return
	}

	// Redirect to the redirect URI with the authorization code.
	redirectUri := authzCode.RedirectUri + "?code=" + authzCode.Code
	if sessionData.OAuthParameters.State != "" {
		redirectUri += "&state=" + sessionData.OAuthParameters.State
	}
	http.Redirect(w, r, redirectUri, http.StatusFound)
}
