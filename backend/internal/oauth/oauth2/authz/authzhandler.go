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

// Package authz provides handlers and utilities for managing OAuth2 authorization requests.
package authz

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	appprovider "github.com/asgardeo/thunder/internal/application/provider"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/authz/constants"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/authz/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/authz/store"
	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	oauth2model "github.com/asgardeo/thunder/internal/oauth/oauth2/model"
	oauth2utils "github.com/asgardeo/thunder/internal/oauth/oauth2/utils"
	sessionmodel "github.com/asgardeo/thunder/internal/oauth/session/model"
	sessionstore "github.com/asgardeo/thunder/internal/oauth/session/store"
	sessionutils "github.com/asgardeo/thunder/internal/oauth/session/utils"
	"github.com/asgardeo/thunder/internal/system/config"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/jwt"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
	systemutils "github.com/asgardeo/thunder/internal/system/utils"
)

const loggerComponentName = "AuthorizeHandler"

// AuthorizeHandlerInterface defines the interface for handling OAuth2 authorization requests.
type AuthorizeHandlerInterface interface {
	HandleAuthorizeGetRequest(w http.ResponseWriter, r *http.Request)
	HandleAuthorizePostRequest(w http.ResponseWriter, r *http.Request)
}

// AuthorizeHandler implements the AuthorizeHandlerInterface for handling OAuth2 authorization requests.
type AuthorizeHandler struct {
	AppProvider    appprovider.ApplicationProviderInterface
	AuthZValidator AuthorizationValidatorInterface
	AuthZStore     store.AuthorizationCodeStoreInterface
	SessionStore   sessionstore.SessionDataStoreInterface
	JWTService     jwt.JWTServiceInterface
}

// NewAuthorizeHandler creates a new instance of AuthorizeHandler.
func NewAuthorizeHandler() AuthorizeHandlerInterface {
	return &AuthorizeHandler{
		AppProvider:    appprovider.NewApplicationProvider(),
		AuthZValidator: NewAuthorizationValidator(),
		AuthZStore:     store.NewAuthorizationCodeStore(),
		SessionStore:   sessionstore.GetSessionDataStore(),
		JWTService:     jwt.GetJWTService(),
	}
}

// HandleAuthorizeGetRequest handles the GET request for OAuth2 authorization.
func (ah *AuthorizeHandler) HandleAuthorizeGetRequest(w http.ResponseWriter, r *http.Request) {
	oAuthMessage := ah.getOAuthMessage(r, w)
	if oAuthMessage == nil {
		return
	}
	ah.handleInitialAuthorizationRequest(oAuthMessage, w, r)
}

// HandleAuthorizePostRequest handles the POST request for OAuth2 authorization.
func (ah *AuthorizeHandler) HandleAuthorizePostRequest(w http.ResponseWriter, r *http.Request) {
	oAuthMessage := ah.getOAuthMessage(r, w)
	if oAuthMessage == nil {
		return
	}

	switch oAuthMessage.RequestType {
	case oauth2const.TypeAuthorizationResponseFromEngine:
		ah.handleAuthorizationResponseFromEngine(oAuthMessage, w)
	case oauth2const.TypeConsentResponseFromUser:
	// TODO: Handle the consent response from the user.
	//  Verify whether we need separate session data key for consent flow.
	//  Alternatively could add consent info also to the same session object.
	default:
		// Handle the case where the request is not recognized.
		utils.WriteJSONError(w, oauth2const.ErrorInvalidRequest, "Invalid authorization request",
			http.StatusBadRequest, nil)
	}
}

// handleInitialAuthorizationRequest handles the initial authorization request from the client.
func (ah *AuthorizeHandler) handleInitialAuthorizationRequest(msg *model.OAuthMessage,
	w http.ResponseWriter, r *http.Request) {
	// Extract required parameters.
	clientID := msg.RequestQueryParams[oauth2const.RequestParamClientID]
	redirectURI := msg.RequestQueryParams[oauth2const.RequestParamRedirectURI]
	scope := msg.RequestQueryParams[oauth2const.RequestParamScope]
	state := msg.RequestQueryParams[oauth2const.RequestParamState]
	responseType := msg.RequestQueryParams[oauth2const.RequestParamResponseType]

	if clientID == "" {
		ah.redirectToErrorPage(w, r, oauth2const.ErrorInvalidRequest, "Missing client_id parameter")
		return
	}

	// Retrieve the OAuth application based on the client Id.
	appService := ah.AppProvider.GetApplicationService()
	app, svcErr := appService.GetOAuthApplication(clientID)
	if svcErr != nil || app == nil {
		ah.redirectToErrorPage(w, r, oauth2const.ErrorInvalidClient, "Invalid client_id")
		return
	}

	// Validate the authorization request.
	sendErrorToApp, errorCode, errorMessage := ah.AuthZValidator.validateInitialAuthorizationRequest(msg, app)
	if errorCode != "" {
		if sendErrorToApp && redirectURI != "" {
			// Redirect to the redirect URI with an error.
			redirectURI, err := oauth2utils.GetURIWithQueryParams(redirectURI, map[string]string{
				oauth2const.RequestParamError:            errorCode,
				oauth2const.RequestParamErrorDescription: errorMessage,
			})
			if err != nil {
				ah.redirectToErrorPage(w, r, oauth2const.ErrorServerError, "Failed to redirect to login page")
				return
			}

			if state != "" {
				redirectURI += "&" + oauth2const.RequestParamState + "=" + state
			}
			http.Redirect(w, r, redirectURI, http.StatusFound)
			return
		} else {
			ah.redirectToErrorPage(w, r, errorCode, errorMessage)
			return
		}
	}

	// Construct session data.
	oauthParams := oauth2model.OAuthParameters{
		SessionDataKey: sessionutils.GenerateNewSessionDataKey(),
		State:          state,
		ClientID:       clientID,
		RedirectURI:    redirectURI,
		ResponseType:   responseType,
		Scopes:         scope,
	}

	// Set the redirect URI if not provided in the request. Invalid cases are already handled at this point.
	// TODO: This should be removed when supporting other means of authorization.
	if redirectURI == "" {
		oauthParams.RedirectURI = app.RedirectURIs[0]
	}

	sessionData := sessionmodel.SessionData{
		OAuthParameters: oauthParams,
		AuthTime:        time.Now(),
	}

	// Store session data in the session store.
	ah.SessionStore.AddSession(oauthParams.SessionDataKey, sessionData)

	// Add required query parameters.
	queryParams := make(map[string]string)
	queryParams[oauth2const.SessionDataKey] = oauthParams.SessionDataKey
	queryParams[oauth2const.AppID] = app.AppID

	// Add insecure warning if the redirect URI is not using TLS.
	// TODO: May require another redirection to a warn consent page when it directly goes to a federated IDP.
	parsedRedirectURI, err := utils.ParseURL(oauthParams.RedirectURI)
	if err != nil {
		ah.redirectToErrorPage(w, r, oauth2const.ErrorServerError, "Failed to redirect to login page")
		return
	}
	if parsedRedirectURI.Scheme == "http" {
		queryParams[oauth2const.ShowInsecureWarning] = "true"
	}

	ah.redirectToLoginPage(w, r, queryParams)
}

// handleAuthorizationResponseFromEngine handles the authorization response from the engine.
func (ah *AuthorizeHandler) handleAuthorizationResponseFromEngine(msg *model.OAuthMessage,
	w http.ResponseWriter) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	// Validate the session data.
	sessionData := msg.SessionData
	if sessionData == nil {
		ah.writeAuthZResponseToErrorPage(w, oauth2const.ErrorInvalidRequest, "Invalid authorization request", nil)
		return
	}

	// Read the assertion from the request body.
	assertion, ok := msg.RequestBodyParams[oauth2const.Assertion]
	if !ok || assertion == "" {
		ah.writeAuthZResponseToErrorPage(w, oauth2const.ErrorInvalidRequest, "Invalid authorization request",
			sessionData)
		return
	}

	// Verify the assertion.
	err := ah.verifyAssertion(assertion, logger)
	if err != nil {
		ah.writeAuthZResponseToErrorPage(w, oauth2const.ErrorInvalidRequest, err.Error(), sessionData)
		return
	}

	// Decode user attributes from the assertion.
	userID, _, err := decodeAttributesFromAssertion(assertion)
	if err != nil {
		logger.Error("Failed to decode user attributes from assertion", log.Error(err))
		ah.writeAuthZResponseToErrorPage(w, oauth2const.ErrorInvalidRequest, "Something went wrong", sessionData)
		return
	}

	if userID == "" {
		logger.Error("User ID is empty after decoding assertion")
		ah.writeAuthZResponseToErrorPage(w, oauth2const.ErrorInvalidRequest, "Invalid user ID", sessionData)
		return
	}

	// TODO: Do user authorization.
	//  Should validate for the scopes as well.

	// Generate the authorization code.
	authzCode, err := getAuthorizationCode(msg, userID)
	if err != nil {
		logger.Error("Failed to generate authorization code", log.Error(err))
		ah.writeAuthZResponseToErrorPage(w, oauth2const.ErrorServerError, "Failed to generate authorization code",
			sessionData)
		return
	}

	// Persist the authorization code.
	persistErr := ah.AuthZStore.InsertAuthorizationCode(authzCode)
	if persistErr != nil {
		logger.Error("Failed to persist authorization code", log.Error(persistErr))
		ah.writeAuthZResponseToErrorPage(w, oauth2const.ErrorServerError, "Failed to persist authorization code",
			sessionData)
		return
	}

	// Construct the redirect URI with the authorization code.
	redirectURI := authzCode.RedirectURI + "?code=" + authzCode.Code
	if sessionData.OAuthParameters.State != "" {
		redirectURI += "&state=" + sessionData.OAuthParameters.State
	}

	ah.writeAuthZResponse(w, redirectURI)
}

// getOAuthMessage extracts the OAuth message from the request and response writer.
func (ah *AuthorizeHandler) getOAuthMessage(r *http.Request, w http.ResponseWriter) *model.OAuthMessage {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if r == nil || w == nil {
		logger.Error("Request or response writer is nil")
		return nil
	}

	var msg *model.OAuthMessage
	var err error

	switch r.Method {
	case http.MethodGet:
		msg, err = ah.getOAuthMessageForGetRequest(r)
	case http.MethodPost:
		msg, err = ah.getOAuthMessageForPostRequest(r)
	default:
		err = errors.New("unsupported request method: " + r.Method)
	}

	if err != nil {
		utils.WriteJSONError(w, oauth2const.ErrorInvalidRequest, "Invalid authorization request",
			http.StatusBadRequest, nil)
	}

	return msg
}

// getOAuthMessageForGetRequest extracts the OAuth message from a authorization GET request.
func (ah *AuthorizeHandler) getOAuthMessageForGetRequest(r *http.Request) (*model.OAuthMessage, error) {
	if err := r.ParseForm(); err != nil {
		return nil, errors.New("failed to parse form data: " + err.Error())
	}

	queryParams := make(map[string]string)
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			queryParams[key] = values[0]
		}
	}

	return &model.OAuthMessage{
		RequestType:        oauth2const.TypeInitialAuthorizationRequest,
		SessionData:        nil,
		RequestQueryParams: queryParams,
	}, nil
}

// getOAuthMessageForPostRequest extracts the OAuth message from a authorization POST request.
func (ah *AuthorizeHandler) getOAuthMessageForPostRequest(r *http.Request) (*model.OAuthMessage, error) {
	authZReq, err := systemutils.DecodeJSONBody[model.AuthZPostRequest](r)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JSON body: %w", err)
	}

	if authZReq.SessionDataKey == "" || authZReq.Assertion == "" {
		return nil, errors.New("sessionDataKey or assertion is missing")
	}

	// Determine the request type.
	// TODO: Require to handle other types such as user consent, etc.
	requestType := oauth2const.TypeAuthorizationResponseFromEngine

	sessionDataKey := authZReq.SessionDataKey
	ok, sessionData := ah.SessionStore.GetSession(sessionDataKey)
	if !ok {
		return nil, fmt.Errorf("session data not found for session data key: %s", sessionDataKey)
	}

	// Remove the session data after retrieval.
	ah.SessionStore.ClearSession(sessionDataKey)

	bodyParams := map[string]string{
		oauth2const.Assertion: authZReq.Assertion,
	}

	return &model.OAuthMessage{
		RequestType:       requestType,
		SessionData:       &sessionData,
		RequestBodyParams: bodyParams,
	}, nil
}

// getLoginPageRedirectURI constructs the login page URL with the provided query parameters.
func getLoginPageRedirectURI(queryParams map[string]string) (string, error) {
	gateClientConfig := config.GetThunderRuntime().Config.GateClient
	loginPageURL := (&url.URL{
		Scheme: gateClientConfig.Scheme,
		Host:   fmt.Sprintf("%s:%d", gateClientConfig.Hostname, gateClientConfig.Port),
		Path:   gateClientConfig.LoginPath,
	}).String()

	return utils.GetURIWithQueryParams(loginPageURL, queryParams)
}

// redirectToLoginPage constructs the login page URL and redirects the user to it.
func (ah *AuthorizeHandler) redirectToLoginPage(w http.ResponseWriter, r *http.Request,
	queryParams map[string]string) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if w == nil || r == nil {
		logger.Error("Response writer or request is nil. Cannot redirect to login page.")
		return
	}

	redirectURI, err := getLoginPageRedirectURI(queryParams)
	if err != nil {
		logger.Error("Failed to construct login page URL: " + err.Error())
		return
	}
	logger.Debug("Redirecting to login page: " + redirectURI)

	http.Redirect(w, r, redirectURI, http.StatusFound)
}

// getErrorPageRedirectURL constructs the error page URL with the provided error code and message.
func getErrorPageRedirectURL(code, msg string) (string, error) {
	gateClientConfig := config.GetThunderRuntime().Config.GateClient
	errorPageURL := (&url.URL{
		Scheme: gateClientConfig.Scheme,
		Host:   fmt.Sprintf("%s:%d", gateClientConfig.Hostname, gateClientConfig.Port),
		Path:   gateClientConfig.ErrorPath,
	}).String()

	queryParams := map[string]string{
		"errorCode":    code,
		"errorMessage": msg,
	}

	return utils.GetURIWithQueryParams(errorPageURL, queryParams)
}

// redirectToErrorPage constructs the error page URL and redirects the user to it.
func (ah *AuthorizeHandler) redirectToErrorPage(w http.ResponseWriter, r *http.Request, code, msg string) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if w == nil || r == nil {
		logger.Error("Response writer or request is nil. Cannot redirect to error page.")
		return
	}

	redirectURL, err := getErrorPageRedirectURL(code, msg)
	if err != nil {
		logger.Error("Failed to construct error page URL: " + err.Error())
		http.Error(w, "Failed to redirect to error page", http.StatusInternalServerError)
		return
	}
	logger.Debug("Redirecting to error page: " + redirectURL)

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// writeAuthZResponse writes the authorization response to the HTTP response writer.
func (ah *AuthorizeHandler) writeAuthZResponse(w http.ResponseWriter, redirectURI string) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	authZResp := model.AuthZPostResponse{
		RedirectURI: redirectURI,
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	err := json.NewEncoder(w).Encode(authZResp)
	if err != nil {
		logger.Error("Error encoding response", log.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// writeAuthZResponseToErrorPage writes the authorization response to the error page.
func (ah *AuthorizeHandler) writeAuthZResponseToErrorPage(w http.ResponseWriter, code, msg string,
	sessionData *sessionmodel.SessionData) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	redirectURI, err := getErrorPageRedirectURL(code, msg)
	if err != nil {
		logger.Error("Failed to construct error page URL: " + err.Error())
		http.Error(w, "Failed to redirect to error page", http.StatusInternalServerError)
		return
	}

	if sessionData != nil && sessionData.OAuthParameters.State != "" {
		redirectURI += "&state=" + sessionData.OAuthParameters.State
	}

	ah.writeAuthZResponse(w, redirectURI)
}

// getAuthorizationCode generates an authorization code based on the provided OAuth message.
func getAuthorizationCode(oAuthMessage *model.OAuthMessage, authUserID string) (
	model.AuthorizationCode, error) {
	sessionData := oAuthMessage.SessionData
	clientID := sessionData.OAuthParameters.ClientID
	if clientID == "" {
		clientID = oAuthMessage.RequestQueryParams["client_id"]
	}
	redirectURI := sessionData.OAuthParameters.RedirectURI
	if redirectURI == "" {
		redirectURI = oAuthMessage.RequestQueryParams["redirect_uri"]
	}

	if clientID == "" || redirectURI == "" {
		return model.AuthorizationCode{}, errors.New("client_id or redirect_uri is missing")
	}

	if authUserID == "" {
		return model.AuthorizationCode{}, errors.New("authenticated user not found")
	}

	authTime := sessionData.AuthTime
	if authTime.IsZero() {
		return model.AuthorizationCode{}, errors.New("authentication time is not set")
	}

	scope := sessionData.OAuthParameters.Scopes
	if scope == "" {
		scope = oAuthMessage.RequestQueryParams["scope"]
	}

	// TODO: Add expiry time logic.
	expiryTime := authTime.Add(10 * time.Minute)

	return model.AuthorizationCode{
		CodeID:           utils.GenerateUUID(),
		Code:             utils.GenerateUUID(),
		ClientID:         clientID,
		RedirectURI:      redirectURI,
		AuthorizedUserID: authUserID,
		TimeCreated:      authTime,
		ExpiryTime:       expiryTime,
		Scopes:           scope,
		State:            constants.AuthCodeStateActive,
	}, nil
}

// verifyAssertion verifies the JWT assertion.
func (ah *AuthorizeHandler) verifyAssertion(assertion string, logger *log.Logger) error {
	if err := ah.JWTService.VerifyJWT(assertion, "", ""); err != nil {
		logger.Debug("Invalid assertion signature", log.Error(err))
		return errors.New("invalid assertion signature")
	}

	return nil
}

// decodeAttributesFromAssertion decodes user attributes from the flow assertion JWT.
// It returns the user ID, a map of user attributes, and an error if any.
func decodeAttributesFromAssertion(assertion string) (string, map[string]string, error) {
	_, jwtPayload, err := jwt.DecodeJWT(assertion)
	if err != nil {
		return "", nil, errors.New("Failed to decode the JWT token: " + err.Error())
	}

	userAttributes := make(map[string]string)
	userID := ""
	for key, value := range jwtPayload {
		switch key {
		case "sub":
			if strValue, ok := value.(string); ok {
				userID = strValue
			} else {
				return "", nil, errors.New("JWT 'sub' claim is not a string")
			}
		case "username":
			if strValue, ok := value.(string); ok {
				userAttributes["username"] = strValue
			} else {
				return "", nil, errors.New("JWT 'username' claim is not a string")
			}
		case "email":
			if strValue, ok := value.(string); ok {
				userAttributes["email"] = strValue
			} else {
				return "", nil, errors.New("JWT 'email' claim is not a string")
			}
		case "firstName":
			if strValue, ok := value.(string); ok {
				userAttributes["firstName"] = strValue
			} else {
				return "", nil, errors.New("JWT 'firstName' claim is not a string")
			}
		case "lastName":
			if strValue, ok := value.(string); ok {
				userAttributes["lastName"] = strValue
			} else {
				return "", nil, errors.New("JWT 'lastName' claim is not a string")
			}
		}
	}

	return userID, userAttributes, nil
}
