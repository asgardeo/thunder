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

package authz

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/asgardeo/thunder/internal/application"
	flowcm "github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/flowexec"
	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	oauth2model "github.com/asgardeo/thunder/internal/oauth/oauth2/model"
	oauth2utils "github.com/asgardeo/thunder/internal/oauth/oauth2/utils"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/jwt"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
)

const loggerComponentName = "AuthorizeHandler"

// AuthorizeHandlerInterface defines the interface for handling OAuth2 authorization requests.
type AuthorizeHandlerInterface interface {
	HandleAuthorizeGetRequest(w http.ResponseWriter, r *http.Request)
	HandleAuthorizePostRequest(w http.ResponseWriter, r *http.Request)
}

// authorizeHandler implements the AuthorizeHandlerInterface for handling OAuth2 authorization requests.
type authorizeHandler struct {
	appService      application.ApplicationServiceInterface
	authZValidator  AuthorizationValidatorInterface
	authCodeStore   AuthorizationCodeStoreInterface
	authReqStore    authorizationRequestStoreInterface
	jwtService      jwt.JWTServiceInterface
	flowExecService flowexec.FlowExecServiceInterface
}

// newAuthorizeHandler creates a new instance of authorizeHandler with injected dependencies.
func newAuthorizeHandler(
	appService application.ApplicationServiceInterface,
	jwtService jwt.JWTServiceInterface,
	authCodeStore AuthorizationCodeStoreInterface,
	authReqStore authorizationRequestStoreInterface,
	flowExecService flowexec.FlowExecServiceInterface,
) AuthorizeHandlerInterface {
	return &authorizeHandler{
		appService:      appService,
		authZValidator:  newAuthorizationValidator(),
		authCodeStore:   authCodeStore,
		authReqStore:    authReqStore,
		jwtService:      jwtService,
		flowExecService: flowExecService,
	}
}

// HandleAuthorizeGetRequest handles the GET request for OAuth2 authorization.
func (ah *authorizeHandler) HandleAuthorizeGetRequest(w http.ResponseWriter, r *http.Request) {
	oAuthMessage := ah.getOAuthMessage(r, w)
	if oAuthMessage == nil {
		return
	}
	ah.handleInitialAuthorizationRequest(oAuthMessage, w, r)
}

// HandleAuthorizePostRequest handles the POST request for OAuth2 authorization.
func (ah *authorizeHandler) HandleAuthorizePostRequest(w http.ResponseWriter, r *http.Request) {
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
func (ah *authorizeHandler) handleInitialAuthorizationRequest(msg *OAuthMessage,
	w http.ResponseWriter, r *http.Request) {
	// Extract required parameters.
	clientID := msg.RequestQueryParams[oauth2const.RequestParamClientID]
	redirectURI := msg.RequestQueryParams[oauth2const.RequestParamRedirectURI]
	scope := msg.RequestQueryParams[oauth2const.RequestParamScope]
	state := msg.RequestQueryParams[oauth2const.RequestParamState]
	responseType := msg.RequestQueryParams[oauth2const.RequestParamResponseType]

	// Extract PKCE parameters
	codeChallenge := msg.RequestQueryParams[oauth2const.RequestParamCodeChallenge]
	codeChallengeMethod := msg.RequestQueryParams[oauth2const.RequestParamCodeChallengeMethod]

	// Extract resource parameter
	resource := msg.RequestQueryParams[oauth2const.RequestParamResource]

	if clientID == "" {
		ah.redirectToErrorPage(w, r, oauth2const.ErrorInvalidRequest, "Missing client_id parameter")
		return
	}

	// Retrieve the OAuth application based on the client Id.
	app, svcErr := ah.appService.GetOAuthApplication(clientID)
	if svcErr != nil || app == nil {
		ah.redirectToErrorPage(w, r, oauth2const.ErrorInvalidClient, "Invalid client_id")
		return
	}

	// Validate the authorization request.
	sendErrorToApp, errorCode, errorMessage := ah.authZValidator.validateInitialAuthorizationRequest(msg, app)
	if errorCode != "" {
		if sendErrorToApp && redirectURI != "" {
			// Redirect to the redirect URI with an error.
			queryParams := map[string]string{
				oauth2const.RequestParamError:            errorCode,
				oauth2const.RequestParamErrorDescription: errorMessage,
			}
			if state != "" {
				queryParams[oauth2const.RequestParamState] = state
			}
			redirectURI, err := oauth2utils.GetURIWithQueryParams(redirectURI, queryParams)
			if err != nil {
				ah.redirectToErrorPage(w, r, oauth2const.ErrorServerError, "Failed to redirect to login page")
				return
			}
			http.Redirect(w, r, redirectURI, http.StatusFound)
			return
		} else {
			ah.redirectToErrorPage(w, r, errorCode, errorMessage)
			return
		}
	}

	oidcScopes, nonOidcScopes := oauth2utils.SeparateOIDCAndNonOIDCScopes(scope)

	// Construct authorization request context.
	oauthParams := oauth2model.OAuthParameters{
		State:               state,
		ClientID:            clientID,
		RedirectURI:         redirectURI,
		ResponseType:        responseType,
		StandardScopes:      oidcScopes,
		PermissionScopes:    nonOidcScopes,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		Resource:            resource,
	}

	// Set the redirect URI if not provided in the request. Invalid cases are already handled at this point.
	// TODO: This should be removed when supporting other means of authorization.
	if redirectURI == "" {
		oauthParams.RedirectURI = app.RedirectURIs[0]
	}

	// Initiate flow with OAuth context
	flowInitCtx := &flowexec.FlowInitContext{
		ApplicationID: app.AppID,
		FlowType:      string(flowcm.FlowTypeAuthentication),
		RuntimeData: map[string]string{
			"requested_permissions": utils.StringifyStringArray(nonOidcScopes, " "),
		},
	}

	flowID, flowErr := ah.flowExecService.InitiateFlow(flowInitCtx)
	if flowErr != nil {
		ah.redirectToErrorPage(w, r, oauth2const.ErrorServerError, "Failed to initiate authentication flow")
		return
	}

	authRequestCtx := authRequestContext{
		OAuthParameters: oauthParams,
	}

	// Store authorization request context in the store.
	identifier := ah.authReqStore.AddRequest(authRequestCtx)
	if identifier == "" {
		ah.redirectToErrorPage(w, r, oauth2const.ErrorServerError, "Failed to store authorization request")
		return
	}

	// Add required query parameters.
	queryParams := make(map[string]string)
	queryParams[oauth2const.AuthID] = identifier
	queryParams[oauth2const.AppID] = app.AppID
	queryParams[oauth2const.FlowID] = flowID

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
func (ah *authorizeHandler) handleAuthorizationResponseFromEngine(msg *OAuthMessage,
	w http.ResponseWriter) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	// Validate the authorization request context.
	authRequestCtx, err := ah.loadAuthRequestContext(msg.AuthID)
	if err != nil {
		ah.writeAuthZResponseToErrorPage(w, oauth2const.ErrorInvalidRequest, "Invalid authorization request", nil)
		return
	}

	// Read the assertion from the request body.
	assertion, ok := msg.RequestBodyParams[oauth2const.Assertion]
	if !ok || assertion == "" {
		ah.writeAuthZResponseToErrorPage(w, oauth2const.ErrorInvalidRequest, "Invalid authorization request",
			authRequestCtx)
		return
	}

	// Verify the assertion.
	err = ah.verifyAssertion(assertion, logger)
	if err != nil {
		ah.writeAuthZResponseToErrorPage(w, oauth2const.ErrorInvalidRequest, err.Error(), authRequestCtx)
		return
	}

	// Decode user attributes from the assertion.
	assertionClaims, authTime, err := decodeAttributesFromAssertion(assertion)
	if err != nil {
		logger.Error("Failed to decode user attributes from assertion", log.Error(err))
		ah.writeAuthZResponseToErrorPage(w, oauth2const.ErrorInvalidRequest, "Something went wrong", authRequestCtx)
		return
	}

	if assertionClaims.userID == "" {
		logger.Error("User ID is empty after decoding assertion")
		ah.writeAuthZResponseToErrorPage(w, oauth2const.ErrorInvalidRequest, "Invalid user ID", authRequestCtx)
		return
	}

	// Extract authorized permissions for permission scopes
	// Overwrite the non oidc scopes in auth request context with the authorized scopes from the assertion.
	if assertionClaims.authorizedPermissions != "" {
		authRequestCtx.OAuthParameters.PermissionScopes = utils.ParseStringArray(
			assertionClaims.authorizedPermissions, " ")
	} else {
		// Clear permission scopes if no authorized permissions in assertion
		authRequestCtx.OAuthParameters.PermissionScopes = []string{}
	}

	// Generate the authorization code.
	authzCode, err := createAuthorizationCode(authRequestCtx, &assertionClaims, authTime)
	if err != nil {
		logger.Error("Failed to generate authorization code", log.Error(err))
		ah.writeAuthZResponseToErrorPage(w, oauth2const.ErrorServerError, "Failed to generate authorization code",
			authRequestCtx)
		return
	}

	// Persist the authorization code.
	persistErr := ah.authCodeStore.InsertAuthorizationCode(authzCode)
	if persistErr != nil {
		logger.Error("Failed to persist authorization code", log.Error(persistErr))
		ah.writeAuthZResponseToErrorPage(w, oauth2const.ErrorServerError, "Failed to persist authorization code",
			authRequestCtx)
		return
	}

	// Construct the redirect URI with the authorization code.
	queryParams := map[string]string{
		"code": authzCode.Code,
	}
	if authRequestCtx.OAuthParameters.State != "" {
		queryParams[oauth2const.RequestParamState] = authRequestCtx.OAuthParameters.State
	}
	redirectURI, err := oauth2utils.GetURIWithQueryParams(authzCode.RedirectURI, queryParams)
	if err != nil {
		logger.Error("Failed to construct redirect URI: " + err.Error())
		ah.writeAuthZResponseToErrorPage(w, oauth2const.ErrorServerError, "Failed to redirect to client",
			authRequestCtx)
		return
	}

	ah.writeAuthZResponse(w, redirectURI)
}

func (ah *authorizeHandler) loadAuthRequestContext(authID string) (*authRequestContext, error) {
	ok, authRequestCtx := ah.authReqStore.GetRequest(authID)
	if !ok {
		return nil, fmt.Errorf("authorization request context not found for auth ID: %s", authID)
	}

	// Remove the authorization request context after retrieval.
	ah.authReqStore.ClearRequest(authID)
	return &authRequestCtx, nil
}

// getOAuthMessage extracts the OAuth message from the request and response writer.
func (ah *authorizeHandler) getOAuthMessage(r *http.Request, w http.ResponseWriter) *OAuthMessage {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if r == nil || w == nil {
		logger.Error("Request or response writer is nil")
		return nil
	}

	var msg *OAuthMessage
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
func (ah *authorizeHandler) getOAuthMessageForGetRequest(r *http.Request) (*OAuthMessage, error) {
	if err := r.ParseForm(); err != nil {
		return nil, errors.New("failed to parse form data: " + err.Error())
	}

	queryParams := make(map[string]string)
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			queryParams[key] = values[0]
		}
	}

	return &OAuthMessage{
		RequestType:        oauth2const.TypeInitialAuthorizationRequest,
		RequestQueryParams: queryParams,
	}, nil
}

// getOAuthMessageForPostRequest extracts the OAuth message from a authorization POST request.
func (ah *authorizeHandler) getOAuthMessageForPostRequest(r *http.Request) (*OAuthMessage, error) {
	authZReq, err := utils.DecodeJSONBody[AuthZPostRequest](r)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JSON body: %w", err)
	}

	if authZReq.AuthID == "" || authZReq.Assertion == "" {
		return nil, errors.New("authId or assertion is missing")
	}

	// Determine the request type.
	// TODO: Require to handle other types such as user consent, etc.
	requestType := oauth2const.TypeAuthorizationResponseFromEngine

	bodyParams := map[string]string{
		oauth2const.Assertion: authZReq.Assertion,
	}

	return &OAuthMessage{
		RequestType:       requestType,
		AuthID:            authZReq.AuthID,
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

	return oauth2utils.GetURIWithQueryParams(loginPageURL, queryParams)
}

// redirectToLoginPage constructs the login page URL and redirects the user to it.
func (ah *authorizeHandler) redirectToLoginPage(w http.ResponseWriter, r *http.Request,
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

	return oauth2utils.GetURIWithQueryParams(errorPageURL, queryParams)
}

// redirectToErrorPage constructs the error page URL and redirects the user to it.
func (ah *authorizeHandler) redirectToErrorPage(w http.ResponseWriter, r *http.Request, code, msg string) {
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
func (ah *authorizeHandler) writeAuthZResponse(w http.ResponseWriter, redirectURI string) {
	authZResp := AuthZPostResponse{
		RedirectURI: redirectURI,
	}

	utils.WriteSuccessResponse(w, http.StatusOK, authZResp)
}

// writeAuthZResponseToErrorPage writes the authorization response to the error page.
func (ah *authorizeHandler) writeAuthZResponseToErrorPage(w http.ResponseWriter, code, msg string,
	authRequestCtx *authRequestContext) {
	redirectURI, err := getErrorPageRedirectURL(code, msg)
	if err != nil {
		http.Error(w, "Failed to redirect to error page", http.StatusInternalServerError)
		return
	}

	if authRequestCtx != nil && authRequestCtx.OAuthParameters.State != "" {
		queryParams := map[string]string{
			oauth2const.RequestParamState: authRequestCtx.OAuthParameters.State,
		}
		redirectURI, err = oauth2utils.GetURIWithQueryParams(redirectURI, queryParams)
		if err != nil {
			http.Error(w, "Failed to redirect to error page", http.StatusInternalServerError)
			return
		}
	}

	ah.writeAuthZResponse(w, redirectURI)
}

// createAuthorizationCode generates an authorization code based on the provided
// authorization request context and authenticated user.
func createAuthorizationCode(
	authRequestCtx *authRequestContext,
	assertionClaims *assertionClaims,
	authTime time.Time,
) (AuthorizationCode, error) {
	clientID := authRequestCtx.OAuthParameters.ClientID
	redirectURI := authRequestCtx.OAuthParameters.RedirectURI

	if clientID == "" || redirectURI == "" {
		return AuthorizationCode{}, errors.New("client_id or redirect_uri is missing")
	}

	if assertionClaims.userID == "" {
		return AuthorizationCode{}, errors.New("authenticated user not found")
	}

	// Use provided authTime, or fallback to current time if zero (iat claim was not available)
	if authTime.IsZero() {
		authTime = time.Now()
	}

	StandardScopes := authRequestCtx.OAuthParameters.StandardScopes
	permissionScopes := authRequestCtx.OAuthParameters.PermissionScopes
	allScopes := append(append([]string{}, StandardScopes...), permissionScopes...)
	resource := authRequestCtx.OAuthParameters.Resource

	oauthConfig := config.GetThunderRuntime().Config.OAuth
	validityPeriod := oauthConfig.AuthorizationCode.ValidityPeriod
	expiryTime := authTime.Add(time.Duration(validityPeriod) * time.Second)

	return AuthorizationCode{
		CodeID:              utils.GenerateUUID(),
		Code:                utils.GenerateUUID(),
		ClientID:            clientID,
		RedirectURI:         redirectURI,
		AuthorizedUserID:    assertionClaims.userID,
		UserAttributes:      assertionClaims.userAttributes,
		TimeCreated:         authTime,
		ExpiryTime:          expiryTime,
		Scopes:              utils.StringifyStringArray(allScopes, " "),
		State:               AuthCodeStateActive,
		CodeChallenge:       authRequestCtx.OAuthParameters.CodeChallenge,
		CodeChallengeMethod: authRequestCtx.OAuthParameters.CodeChallengeMethod,
		Resource:            resource,
	}, nil
}

// verifyAssertion verifies the JWT assertion.
func (ah *authorizeHandler) verifyAssertion(assertion string, logger *log.Logger) error {
	if err := ah.jwtService.VerifyJWT(assertion, "", ""); err != nil {
		logger.Debug("Invalid assertion signature", log.String("error", err.Error))
		return errors.New("invalid assertion signature")
	}

	return nil
}

// decodeAttributesFromAssertion decodes user attributes from the flow assertion JWT.
func decodeAttributesFromAssertion(assertion string) (assertionClaims, time.Time, error) {
	assertionClaims := assertionClaims{
		userAttributes: make(map[string]interface{}),
	}

	_, jwtPayload, err := jwt.DecodeJWT(assertion)
	if err != nil {
		return assertionClaims, time.Time{}, errors.New("Failed to decode the JWT token: " + err.Error())
	}

	// Extract authentication time from iat claim
	authTime := time.Time{}
	if iatValue, ok := jwtPayload["iat"]; ok {
		switch v := iatValue.(type) {
		case float64:
			authTime = time.Unix(int64(v), 0)
		case int64:
			authTime = time.Unix(v, 0)
		case int:
			authTime = time.Unix(int64(v), 0)
		default:
			return assertionClaims, time.Time{}, errors.New("JWT 'iat' claim has unexpected type")
		}
	}

	// Standard JWT claims that should not be treated as user attributes.
	standardClaims := map[string]bool{
		"iss": true, "sub": true, "aud": true, "exp": true, "nbf": true, "iat": true, "jti": true,
		"assurance":              true,
		"authorized_permissions": true,
	}

	userAttributes := make(map[string]interface{})
	for key, value := range jwtPayload {
		// Extract sub claim
		if key == oauth2const.ClaimSub {
			if strValue, ok := value.(string); ok {
				assertionClaims.userID = strValue
			} else {
				return assertionClaims, time.Time{}, errors.New("JWT 'sub' claim is not a string")
			}
			continue
		}

		// Extract authorized_permissions claim
		if key == "authorized_permissions" {
			if strValue, ok := value.(string); ok {
				assertionClaims.authorizedPermissions = strValue
			}
			continue
		}

		// Skip standard JWT claims
		if standardClaims[key] {
			continue
		}

		// All other claims are treated as user attributes
		userAttributes[key] = value
	}
	assertionClaims.userAttributes = userAttributes

	return assertionClaims, authTime, nil
}
