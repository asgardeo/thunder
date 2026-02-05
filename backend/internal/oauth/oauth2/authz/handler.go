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
	"slices"
	"strings"
	"time"

	"github.com/asgardeo/thunder/internal/application"
	appmodel "github.com/asgardeo/thunder/internal/application/model"
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

type AuthorizeHandlerInterface interface {
	HandleAuthorizeGetRequest(w http.ResponseWriter, r *http.Request)
	HandleAuthorizePostRequest(w http.ResponseWriter, r *http.Request)
}

type authorizeHandler struct {
	appService      application.ApplicationServiceInterface
	authZValidator  AuthorizationValidatorInterface
	authCodeStore   AuthorizationCodeStoreInterface
	authReqStore    authorizationRequestStoreInterface
	jwtService      jwt.JWTServiceInterface
	flowExecService flowexec.FlowExecServiceInterface
}

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

func (ah *authorizeHandler) HandleAuthorizeGetRequest(w http.ResponseWriter, r *http.Request) {
	oAuthMessage := ah.getOAuthMessage(r, w)
	if oAuthMessage == nil {
		return
	}
	ah.handleInitialAuthorizationRequest(oAuthMessage, w, r)
}

func (ah *authorizeHandler) HandleAuthorizePostRequest(w http.ResponseWriter, r *http.Request) {
	oAuthMessage := ah.getOAuthMessage(r, w)
	if oAuthMessage == nil {
		return
	}
	switch oAuthMessage.RequestType {
	case oauth2const.TypeAuthorizationResponseFromEngine:
		ah.handleAuthorizationResponseFromEngine(oAuthMessage, w)
	case oauth2const.TypeConsentResponseFromUser:
	default:
		utils.WriteJSONError(w, oauth2const.ErrorInvalidRequest, "Invalid authorization request", http.StatusBadRequest, nil)
	}
}

func (ah *authorizeHandler) handleInitialAuthorizationRequest(msg *OAuthMessage, w http.ResponseWriter, r *http.Request) {
	clientID := msg.RequestQueryParams[oauth2const.RequestParamClientID]
	redirectURI := msg.RequestQueryParams[oauth2const.RequestParamRedirectURI]
	scope := msg.RequestQueryParams[oauth2const.RequestParamScope]
	state := msg.RequestQueryParams[oauth2const.RequestParamState]
	responseType := msg.RequestQueryParams[oauth2const.RequestParamResponseType]
	codeChallenge := msg.RequestQueryParams[oauth2const.RequestParamCodeChallenge]
	codeChallengeMethod := msg.RequestQueryParams[oauth2const.RequestParamCodeChallengeMethod]
	resource := msg.RequestQueryParams[oauth2const.RequestParamResource]
	claimsParam := msg.RequestQueryParams[oauth2const.RequestParamClaims]

	if clientID == "" {
		ah.redirectToErrorPage(w, r, oauth2const.ErrorInvalidRequest, "Missing client_id parameter")
		return
	}
	app, svcErr := ah.appService.GetOAuthApplication(clientID)
	if svcErr != nil || app == nil {
		ah.redirectToErrorPage(w, r, oauth2const.ErrorInvalidClient, "Invalid client_id")
		return
	}
	var claimsRequest *oauth2model.ClaimsRequest
	if claimsParam != "" {
		var err error
		claimsRequest, err = oauth2utils.ParseClaimsRequest(claimsParam)
		if err != nil {
			logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
			logger.Debug("Failed to parse claims parameter", log.Error(err))
			ah.redirectToErrorPage(w, r, oauth2const.ErrorInvalidRequest, "Invalid claims parameter")
			return
		}
	}
	sendErrorToApp, errorCode, errorMessage := ah.authZValidator.validateInitialAuthorizationRequest(msg, app)
	if errorCode != "" {
		if sendErrorToApp && redirectURI != "" {
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
		}
		ah.redirectToErrorPage(w, r, errorCode, errorMessage)
		return
	}
	oidcScopes, nonOidcScopes := oauth2utils.SeparateOIDCAndNonOIDCScopes(scope)
	oauthParams := oauth2model.OAuthParameters{
		State: state, ClientID: clientID, RedirectURI: redirectURI, ResponseType: responseType,
		StandardScopes: oidcScopes, PermissionScopes: nonOidcScopes,
		CodeChallenge: codeChallenge, CodeChallengeMethod: codeChallengeMethod,
		Resource: resource, ClaimsRequest: claimsRequest,
	}
	if redirectURI == "" {
		oauthParams.RedirectURI = app.RedirectURIs[0]
	}
	requiredAttributes := getRequiredAttributes(oidcScopes, app, claimsRequest)
	runtimeData := map[string]string{
		"requested_permissions": utils.StringifyStringArray(nonOidcScopes, " "),
		"required_attributes":   requiredAttributes,
	}
	flowInitCtx := &flowexec.FlowInitContext{
		ApplicationID: app.AppID, FlowType: string(flowcm.FlowTypeAuthentication), RuntimeData: runtimeData,
	}
	flowID, flowErr := ah.flowExecService.InitiateFlow(flowInitCtx)
	if flowErr != nil {
		ah.redirectToErrorPage(w, r, oauth2const.ErrorServerError, "Failed to initiate authentication flow")
		return
	}
	authRequestCtx := authRequestContext{OAuthParameters: oauthParams}
	identifier := ah.authReqStore.AddRequest(authRequestCtx)
	if identifier == "" {
		ah.redirectToErrorPage(w, r, oauth2const.ErrorServerError, "Failed to store authorization request")
		return
	}
	queryParams := make(map[string]string)
	queryParams[oauth2const.AuthID] = identifier
	queryParams[oauth2const.AppID] = app.AppID
	queryParams[oauth2const.FlowID] = flowID
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

func (ah *authorizeHandler) handleAuthorizationResponseFromEngine(msg *OAuthMessage, w http.ResponseWriter) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	authRequestCtx, err := ah.loadAuthRequestContext(msg.AuthID)
	if err != nil {
		ah.writeAuthZResponseToErrorPage(w, oauth2const.ErrorInvalidRequest, "Invalid authorization request", nil)
		return
	}
	assertion, ok := msg.RequestBodyParams[oauth2const.Assertion]
	if !ok || assertion == "" {
		ah.writeAuthZResponseToErrorPage(w, oauth2const.ErrorInvalidRequest, "Invalid authorization request", authRequestCtx)
		return
	}
	err = ah.verifyAssertion(assertion, logger)
	if err != nil {
		ah.writeAuthZResponseToErrorPage(w, oauth2const.ErrorInvalidRequest, err.Error(), authRequestCtx)
		return
	}
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
	hasOpenIDScope := slices.Contains(authRequestCtx.OAuthParameters.StandardScopes, "openid")
	if hasOpenIDScope {
		if err := validateSubClaimConstraint(authRequestCtx.OAuthParameters.ClaimsRequest, assertionClaims.userID); err != nil {
			logger.Debug("Sub claim validation failed", log.Error(err))
			ah.writeAuthZResponseToErrorPage(w, oauth2const.ErrorAccessDenied, "Subject identifier mismatch", authRequestCtx)
			return
		}
	}
	if assertionClaims.authorizedPermissions != "" {
		authRequestCtx.OAuthParameters.PermissionScopes = utils.ParseStringArray(assertionClaims.authorizedPermissions, " ")
	} else {
		authRequestCtx.OAuthParameters.PermissionScopes = []string{}
	}
	authzCode, err := createAuthorizationCode(authRequestCtx, &assertionClaims, authTime)
	if err != nil {
		logger.Error("Failed to generate authorization code", log.Error(err))
		ah.writeAuthZResponseToErrorPage(w, oauth2const.ErrorServerError, "Failed to generate authorization code", authRequestCtx)
		return
	}
	persistErr := ah.authCodeStore.InsertAuthorizationCode(authzCode)
	if persistErr != nil {
		logger.Error("Failed to persist authorization code", log.Error(persistErr))
		ah.writeAuthZResponseToErrorPage(w, oauth2const.ErrorServerError, "Failed to persist authorization code", authRequestCtx)
		return
	}
	queryParams := map[string]string{"code": authzCode.Code}
	if authRequestCtx.OAuthParameters.State != "" {
		queryParams[oauth2const.RequestParamState] = authRequestCtx.OAuthParameters.State
	}
	redirectURI, err := oauth2utils.GetURIWithQueryParams(authzCode.RedirectURI, queryParams)
	if err != nil {
		logger.Error("Failed to construct redirect URI: " + err.Error())
		ah.writeAuthZResponseToErrorPage(w, oauth2const.ErrorServerError, "Failed to redirect to client", authRequestCtx)
		return
	}
	ah.writeAuthZResponse(w, redirectURI)
}

func (ah *authorizeHandler) loadAuthRequestContext(authID string) (*authRequestContext, error) {
	ok, authRequestCtx := ah.authReqStore.GetRequest(authID)
	if !ok {
		return nil, fmt.Errorf("authorization request context not found for auth ID: %s", authID)
	}
	ah.authReqStore.ClearRequest(authID)
	return &authRequestCtx, nil
}

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
		utils.WriteJSONError(w, oauth2const.ErrorInvalidRequest, "Invalid authorization request", http.StatusBadRequest, nil)
	}
	return msg
}

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
	return &OAuthMessage{RequestType: oauth2const.TypeInitialAuthorizationRequest, RequestQueryParams: queryParams}, nil
}

func (ah *authorizeHandler) getOAuthMessageForPostRequest(r *http.Request) (*OAuthMessage, error) {
	authZReq, err := utils.DecodeJSONBody[AuthZPostRequest](r)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JSON body: %w", err)
	}
	if authZReq.AuthID == "" || authZReq.Assertion == "" {
		return nil, errors.New("authId or assertion is missing")
	}
	requestType := oauth2const.TypeAuthorizationResponseFromEngine
	bodyParams := map[string]string{oauth2const.Assertion: authZReq.Assertion}
	return &OAuthMessage{RequestType: requestType, AuthID: authZReq.AuthID, RequestBodyParams: bodyParams}, nil
}

func getLoginPageRedirectURI(queryParams map[string]string) (string, error) {
	gateClientConfig := config.GetThunderRuntime().Config.GateClient
	loginPageURL := (&url.URL{
		Scheme: gateClientConfig.Scheme,
		Host:   fmt.Sprintf("%s:%d", gateClientConfig.Hostname, gateClientConfig.Port),
		Path:   gateClientConfig.LoginPath,
	}).String()
	return oauth2utils.GetURIWithQueryParams(loginPageURL, queryParams)
}

func (ah *authorizeHandler) redirectToLoginPage(w http.ResponseWriter, r *http.Request, queryParams map[string]string) {
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

func getErrorPageRedirectURL(code, msg string) (string, error) {
	gateClientConfig := config.GetThunderRuntime().Config.GateClient
	errorPageURL := (&url.URL{
		Scheme: gateClientConfig.Scheme,
		Host:   fmt.Sprintf("%s:%d", gateClientConfig.Hostname, gateClientConfig.Port),
		Path:   gateClientConfig.ErrorPath,
	}).String()
	queryParams := map[string]string{"errorCode": code, "errorMessage": msg}
	return oauth2utils.GetURIWithQueryParams(errorPageURL, queryParams)
}

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

func (ah *authorizeHandler) writeAuthZResponse(w http.ResponseWriter, redirectURI string) {
	authZResp := AuthZPostResponse{RedirectURI: redirectURI}
	utils.WriteSuccessResponse(w, http.StatusOK, authZResp)
}

func (ah *authorizeHandler) writeAuthZResponseToErrorPage(w http.ResponseWriter, code, msg string, authRequestCtx *authRequestContext) {
	redirectURI, err := getErrorPageRedirectURL(code, msg)
	if err != nil {
		http.Error(w, "Failed to redirect to error page", http.StatusInternalServerError)
		return
	}
	if authRequestCtx != nil && authRequestCtx.OAuthParameters.State != "" {
		queryParams := map[string]string{oauth2const.RequestParamState: authRequestCtx.OAuthParameters.State}
		redirectURI, err = oauth2utils.GetURIWithQueryParams(redirectURI, queryParams)
		if err != nil {
			http.Error(w, "Failed to redirect to error page", http.StatusInternalServerError)
			return
		}
	}
	ah.writeAuthZResponse(w, redirectURI)
}

func createAuthorizationCode(authRequestCtx *authRequestContext, assertionClaims *assertionClaims, authTime time.Time) (AuthorizationCode, error) {
	clientID := authRequestCtx.OAuthParameters.ClientID
	redirectURI := authRequestCtx.OAuthParameters.RedirectURI
	if clientID == "" || redirectURI == "" {
		return AuthorizationCode{}, errors.New("client_id or redirect_uri is missing")
	}
	if assertionClaims.userID == "" {
		return AuthorizationCode{}, errors.New("authenticated user not found")
	}
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
	codeID, err := utils.GenerateUUIDv7()
	if err != nil {
		return AuthorizationCode{}, errors.New("Failed to generate UUID")
	}
	code, err := utils.GenerateUUIDv7()
	if err != nil {
		return AuthorizationCode{}, errors.New("Failed to generate UUID")
	}
	return AuthorizationCode{
		CodeID: codeID, Code: code, ClientID: clientID, RedirectURI: redirectURI,
		AuthorizedUserID: assertionClaims.userID, UserAttributes: assertionClaims.userAttributes,
		TimeCreated: authTime, ExpiryTime: expiryTime, Scopes: utils.StringifyStringArray(allScopes, " "),
		State: AuthCodeStateActive, CodeChallenge: authRequestCtx.OAuthParameters.CodeChallenge,
		CodeChallengeMethod: authRequestCtx.OAuthParameters.CodeChallengeMethod, Resource: resource,
		ClaimsRequest: authRequestCtx.OAuthParameters.ClaimsRequest,
	}, nil
}

func (ah *authorizeHandler) verifyAssertion(assertion string, logger *log.Logger) error {
	if err := ah.jwtService.VerifyJWT(assertion, "", ""); err != nil {
		logger.Debug("Invalid assertion signature", log.String("error", err.Error))
		return errors.New("invalid assertion signature")
	}
	return nil
}

func decodeAttributesFromAssertion(assertion string) (assertionClaims, time.Time, error) {
	assertionClaims := assertionClaims{userAttributes: make(map[string]interface{})}
	_, jwtPayload, err := jwt.DecodeJWT(assertion)
	if err != nil {
		return assertionClaims, time.Time{}, errors.New("Failed to decode the JWT token: " + err.Error())
	}
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
	standardClaims := map[string]bool{
		"iss": true, "sub": true, "aud": true, "exp": true, "nbf": true, "iat": true, "jti": true,
		"assurance": true, "authorized_permissions": true,
	}
	userAttributes := make(map[string]interface{})
	for key, value := range jwtPayload {
		if key == oauth2const.ClaimSub {
			if strValue, ok := value.(string); ok {
				assertionClaims.userID = strValue
			} else {
				return assertionClaims, time.Time{}, errors.New("JWT 'sub' claim is not a string")
			}
			continue
		}
		if key == "authorized_permissions" {
			if strValue, ok := value.(string); ok {
				assertionClaims.authorizedPermissions = strValue
			}
			continue
		}
		if standardClaims[key] {
			continue
		}
		userAttributes[key] = value
	}
	assertionClaims.userAttributes = userAttributes
	return assertionClaims, authTime, nil
}

func getRequiredAttributes(oidcScopes []string, app *appmodel.OAuthAppConfigProcessedDTO, claimsRequest *oauth2model.ClaimsRequest) string {
	requiredAttrsSet := make(map[string]bool)
	if app == nil || app.Token == nil {
		return ""
	}
	hasOpenIDScope := slices.Contains(oidcScopes, "openid")
	if hasOpenIDScope {
		var idTokenAllowedSet map[string]bool
		scopeClaimsMapping := app.ScopeClaims
		if app.Token.IDToken != nil {
			if len(app.Token.IDToken.UserAttributes) > 0 {
				idTokenAllowedSet = make(map[string]bool, len(app.Token.IDToken.UserAttributes))
				for _, attr := range app.Token.IDToken.UserAttributes {
					idTokenAllowedSet[attr] = true
				}
			}
		}
		if claimsRequest != nil && claimsRequest.IDToken != nil && idTokenAllowedSet != nil {
			for claimName := range claimsRequest.IDToken {
				if idTokenAllowedSet[claimName] {
					requiredAttrsSet[claimName] = true
				}
			}
		}
		for _, scope := range oidcScopes {
			var scopeClaims []string
			if scopeClaimsMapping != nil {
				if appClaims, exists := scopeClaimsMapping[scope]; exists {
					scopeClaims = appClaims
				}
			}
			if scopeClaims == nil {
				if standardScope, exists := oauth2const.StandardOIDCScopes[scope]; exists {
					scopeClaims = standardScope.Claims
				}
			}
			for _, claim := range scopeClaims {
				if idTokenAllowedSet != nil && idTokenAllowedSet[claim] {
					requiredAttrsSet[claim] = true
				}
			}
		}
	}
	if app.Token.AccessToken != nil && len(app.Token.AccessToken.UserAttributes) > 0 {
		for _, attr := range app.Token.AccessToken.UserAttributes {
			requiredAttrsSet[attr] = true
		}
	}
	return mapKeysToSpaceSeparatedString(requiredAttrsSet)
}

func mapKeysToSpaceSeparatedString(m map[string]bool) string {
	if len(m) == 0 {
		return ""
	}
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return strings.Join(keys, " ")
}

func validateSubClaimConstraint(claimsRequest *oauth2model.ClaimsRequest, actualSubject string) error {
	if claimsRequest == nil {
		return nil
	}
	if claimsRequest.IDToken != nil {
		if subReq, exists := claimsRequest.IDToken["sub"]; exists && subReq != nil {
			if !subReq.MatchesValue(actualSubject) {
				return errors.New("sub claim in id_token does not match requested value")
			}
		}
	}
	if claimsRequest.UserInfo != nil {
		if subReq, exists := claimsRequest.UserInfo["sub"]; exists && subReq != nil {
			if !subReq.MatchesValue(actualSubject) {
				return errors.New("sub claim in userinfo does not match requested value")
			}
		}
	}
	return nil
}
