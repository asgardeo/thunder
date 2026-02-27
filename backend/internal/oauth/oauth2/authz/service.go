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

// Package authz implements the OAuth2 authorization functionality.
package authz

import (
	"errors"
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
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/jose/jwt"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
)

// AuthorizeServiceInterface defines the interface for authorization services.
type AuthorizeServiceInterface interface {
	GetAuthorizationCodeDetails(clientID string, code string) (*AuthorizationCode, error)
	HandleInitialAuthorizationRequest(msg *OAuthMessage) (*AuthorizationInitResult, *AuthorizationError)
	HandleAuthorizationCallback(authID string, assertion string) (string, *AuthorizationError)
}

// authorizeService implements the AuthorizeService for managing OAuth2 authorization flows.
type authorizeService struct {
	appService      application.ApplicationServiceInterface
	authZValidator  AuthorizationValidatorInterface
	authCodeStore   AuthorizationCodeStoreInterface
	authReqStore    authorizationRequestStoreInterface
	jwtService      jwt.JWTServiceInterface
	flowExecService flowexec.FlowExecServiceInterface
	logger          *log.Logger
}

// newAuthorizeService creates a new instance of authorizeService with injected dependencies.
func newAuthorizeService(
	appService application.ApplicationServiceInterface,
	jwtService jwt.JWTServiceInterface,
	flowExecService flowexec.FlowExecServiceInterface,
	authCodeStore AuthorizationCodeStoreInterface,
	authReqStore authorizationRequestStoreInterface,
) AuthorizeServiceInterface {
	return &authorizeService{
		appService:      appService,
		authZValidator:  newAuthorizationValidator(),
		authCodeStore:   authCodeStore,
		authReqStore:    authReqStore,
		jwtService:      jwtService,
		flowExecService: flowExecService,
		logger:          log.GetLogger().With(log.String(log.LoggerKeyComponentName, "AuthorizeService")),
	}
}

// GetAuthorizationCodeDetails retrieves and invalidates the authorization code.
func (as *authorizeService) GetAuthorizationCodeDetails(clientID string, code string) (*AuthorizationCode, error) {
	authCode, err := as.authCodeStore.GetAuthorizationCode(clientID, code)
	if err != nil {
		if errors.Is(err, ErrAuthorizationCodeNotFound) {
			return nil, errors.New("invalid authorization code")
		}
		as.logger.Error("error retrieving authorization code", log.Error(err))
		return nil, errors.New("failed to retrieve authorization code")
	}
	if authCode == nil || authCode.Code == "" {
		return nil, errors.New("invalid authorization code")
	}

	// Invalidate the authorization code after use.
	err = as.authCodeStore.DeactivateAuthorizationCode(*authCode)
	if err != nil {
		as.logger.Error("error invalidating authorization code", log.Error(err))
		return nil, errors.New("failed to invalidate authorization code")
	}
	return authCode, nil
}

// HandleInitialAuthorizationRequest processes an initial authorization request from the client.
// Returns the query params needed to redirect to the login page, or a structured authorization error.
func (as *authorizeService) HandleInitialAuthorizationRequest(msg *OAuthMessage) (
	*AuthorizationInitResult, *AuthorizationError) {
	// Extract required parameters.
	clientID := msg.RequestQueryParams[oauth2const.RequestParamClientID]
	redirectURI := msg.RequestQueryParams[oauth2const.RequestParamRedirectURI]
	scope := msg.RequestQueryParams[oauth2const.RequestParamScope]
	state := msg.RequestQueryParams[oauth2const.RequestParamState]
	responseType := msg.RequestQueryParams[oauth2const.RequestParamResponseType]

	// Extract PKCE parameters.
	codeChallenge := msg.RequestQueryParams[oauth2const.RequestParamCodeChallenge]
	codeChallengeMethod := msg.RequestQueryParams[oauth2const.RequestParamCodeChallengeMethod]

	// Extract resource parameter.
	resource := msg.RequestQueryParams[oauth2const.RequestParamResource]

	// Extract claims parameter.
	claimsParam := msg.RequestQueryParams[oauth2const.RequestParamClaims]

	// Extract claims_locales parameter.
	claimsLocales := msg.RequestQueryParams[oauth2const.RequestParamClaimsLocales]

	nonce := msg.RequestQueryParams[oauth2const.RequestParamNonce]

	if clientID == "" {
		return nil, &AuthorizationError{
			Code:    oauth2const.ErrorInvalidRequest,
			Message: "Missing client_id parameter",
		}
	}

	// Retrieve the OAuth application based on the client ID.
	app, svcErr := as.appService.GetOAuthApplication(clientID)
	if svcErr != nil {
		if svcErr.Type == serviceerror.ServerErrorType {
			as.logger.Error("Failed to retrieve OAuth application", log.String("error", svcErr.Error))
			return nil, &AuthorizationError{
				Code:    oauth2const.ErrorServerError,
				Message: "Failed to process authorization request",
			}
		}
		return nil, &AuthorizationError{
			Code:    oauth2const.ErrorInvalidClient,
			Message: "Invalid client_id",
		}
	}
	if app == nil {
		return nil, &AuthorizationError{
			Code:    oauth2const.ErrorInvalidClient,
			Message: "Invalid client_id",
		}
	}

	// Parse the claims parameter if present.
	var claimsRequest *oauth2model.ClaimsRequest
	if claimsParam != "" {
		var err error
		claimsRequest, err = oauth2utils.ParseClaimsRequest(claimsParam)
		if err != nil {
			as.logger.Debug("Failed to parse claims parameter", log.Error(err))
			return nil, &AuthorizationError{
				Code:    oauth2const.ErrorInvalidRequest,
				Message: "Invalid claims parameter",
			}
		}
	}

	// Validate the authorization request.
	sendErrorToApp, errorCode, errorMessage := as.authZValidator.validateInitialAuthorizationRequest(msg, app)
	if errorCode != "" {
		authErr := &AuthorizationError{
			Code:    errorCode,
			Message: errorMessage,
			State:   state,
		}
		if sendErrorToApp && redirectURI != "" {
			authErr.SendErrorToClient = true
			authErr.ClientRedirectURI = redirectURI
		}
		return nil, authErr
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
		ClaimsRequest:       claimsRequest,
		ClaimsLocales:       claimsLocales,
		Nonce:               nonce,
	}

	// Set the redirect URI if not provided in the request. Invalid cases are already handled at this point.
	// TODO: This should be removed when supporting other means of authorization.
	if redirectURI == "" {
		if len(app.RedirectURIs) == 0 {
			as.logger.Error("OAuth application has no registered redirect URIs",
				log.String("client_id", clientID))
			return nil, &AuthorizationError{
				Code:    oauth2const.ErrorServerError,
				Message: "Failed to process authorization request",
			}
		}
		oauthParams.RedirectURI = app.RedirectURIs[0]
	}

	// Compute required attributes from OIDC scopes, access token config, and claims parameter.
	requiredAttributes := getRequiredAttributes(oidcScopes, app, claimsRequest)

	// Initiate flow with OAuth context.
	runtimeData := map[string]string{
		flowcm.RuntimeKeyRequestedPermissions: utils.StringifyStringArray(nonOidcScopes, " "),
		flowcm.RuntimeKeyRequiredAttributes:   requiredAttributes,
		flowcm.RuntimeKeyRequiredLocales:      claimsLocales,
	}
	flowInitCtx := &flowexec.FlowInitContext{
		ApplicationID: app.AppID,
		FlowType:      string(flowcm.FlowTypeAuthentication),
		RuntimeData:   runtimeData,
	}

	flowID, flowErr := as.flowExecService.InitiateFlow(flowInitCtx)
	if flowErr != nil {
		as.logger.Error("Failed to initiate authentication flow", log.String("error", flowErr.Error))
		return nil, &AuthorizationError{
			Code:    oauth2const.ErrorServerError,
			Message: "Failed to initiate authentication flow",
		}
	}

	authRequestCtx := authRequestContext{
		OAuthParameters: oauthParams,
	}

	// Store authorization request context in the store.
	identifier := as.authReqStore.AddRequest(authRequestCtx)
	if identifier == "" {
		as.logger.Error("Failed to store authorization request context")
		return nil, &AuthorizationError{
			Code:    oauth2const.ErrorServerError,
			Message: "Failed to store authorization request",
		}
	}

	// Build query parameters for login page redirect.
	queryParams := make(map[string]string)
	queryParams[oauth2const.AuthID] = identifier
	queryParams[oauth2const.AppID] = app.AppID
	queryParams[oauth2const.FlowID] = flowID

	// Add insecure warning if the redirect URI is not using TLS.
	// TODO: May require another redirection to a warn consent page when it directly goes to a federated IDP.
	parsedRedirectURI, err := utils.ParseURL(oauthParams.RedirectURI)
	if err != nil {
		as.logger.Error("Failed to parse redirect URI", log.Error(err))
		return nil, &AuthorizationError{
			Code:    oauth2const.ErrorServerError,
			Message: "Failed to redirect to login page",
		}
	}
	if parsedRedirectURI.Scheme == "http" {
		queryParams[oauth2const.ShowInsecureWarning] = "true"
	}

	return &AuthorizationInitResult{QueryParams: queryParams}, nil
}

// HandleAuthorizationCallback processes the callback assertion from the flow engine.
// Returns the client redirect URI (with authorization code) on success, or a structured error.
func (as *authorizeService) HandleAuthorizationCallback(authID string, assertion string) (
	string, *AuthorizationError) {
	// Load the authorization request context.
	authRequestCtx, err := as.loadAuthRequestContext(authID)
	if err != nil {
		return "", &AuthorizationError{
			Code:    oauth2const.ErrorInvalidRequest,
			Message: "Invalid authorization request",
		}
	}

	if assertion == "" {
		return "", &AuthorizationError{
			Code:    oauth2const.ErrorInvalidRequest,
			Message: "Invalid authorization request",
			State:   authRequestCtx.OAuthParameters.State,
		}
	}

	// Verify the assertion.
	if err := as.verifyAssertion(assertion); err != nil {
		return "", &AuthorizationError{
			Code:    oauth2const.ErrorInvalidRequest,
			Message: err.Error(),
			State:   authRequestCtx.OAuthParameters.State,
		}
	}

	// Decode user attributes from the assertion.
	claims, authTime, err := decodeAttributesFromAssertion(assertion)
	if err != nil {
		as.logger.Error("Failed to decode user attributes from assertion", log.Error(err))
		return "", &AuthorizationError{
			Code:    oauth2const.ErrorInvalidRequest,
			Message: "Something went wrong",
			State:   authRequestCtx.OAuthParameters.State,
		}
	}

	if claims.userID == "" {
		as.logger.Error("User ID is empty after decoding assertion")
		return "", &AuthorizationError{
			Code:    oauth2const.ErrorInvalidRequest,
			Message: "Invalid user ID",
			State:   authRequestCtx.OAuthParameters.State,
		}
	}

	// Validate sub claim constraint if specified in claims parameter.
	// If sub claim is requested with a value constraint and doesn't match, authentication must fail.
	hasOpenIDScope := slices.Contains(authRequestCtx.OAuthParameters.StandardScopes, oauth2const.ScopeOpenID)
	if hasOpenIDScope {
		if err := validateSubClaimConstraint(authRequestCtx.OAuthParameters.ClaimsRequest, claims.userID); err != nil {
			as.logger.Debug("Sub claim validation failed", log.Error(err))
			return "", &AuthorizationError{
				Code:    oauth2const.ErrorAccessDenied,
				Message: "Subject identifier mismatch",
				State:   authRequestCtx.OAuthParameters.State,
			}
		}
	}

	// Extract authorized permissions for permission scopes.
	// Overwrite the non-OIDC scopes in auth request context with the authorized scopes from the assertion.
	if claims.authorizedPermissions != "" {
		authRequestCtx.OAuthParameters.PermissionScopes = utils.ParseStringArray(
			claims.authorizedPermissions, " ")
	} else {
		// Clear permission scopes if no authorized permissions in assertion.
		authRequestCtx.OAuthParameters.PermissionScopes = []string{}
	}

	// Generate the authorization code.
	authzCode, err := createAuthorizationCode(authRequestCtx, &claims, authTime)
	if err != nil {
		as.logger.Error("Failed to generate authorization code", log.Error(err))
		return "", &AuthorizationError{
			Code:    oauth2const.ErrorServerError,
			Message: "Failed to generate authorization code",
			State:   authRequestCtx.OAuthParameters.State,
		}
	}

	// Persist the authorization code.
	if persistErr := as.authCodeStore.InsertAuthorizationCode(authzCode); persistErr != nil {
		as.logger.Error("Failed to persist authorization code", log.Error(persistErr))
		return "", &AuthorizationError{
			Code:    oauth2const.ErrorServerError,
			Message: "Failed to persist authorization code",
			State:   authRequestCtx.OAuthParameters.State,
		}
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
		as.logger.Error("Failed to construct redirect URI", log.Error(err))
		return "", &AuthorizationError{
			Code:    oauth2const.ErrorServerError,
			Message: "Failed to redirect to client",
			State:   authRequestCtx.OAuthParameters.State,
		}
	}

	return redirectURI, nil
}

// loadAuthRequestContext loads the authorization request context from the store using the auth ID.
func (as *authorizeService) loadAuthRequestContext(authID string) (*authRequestContext, error) {
	ok, authRequestCtx := as.authReqStore.GetRequest(authID)
	if !ok {
		return nil, errors.New("authorization request context not found for auth ID: " + authID)
	}

	// Remove the authorization request context after retrieval.
	as.authReqStore.ClearRequest(authID)
	return &authRequestCtx, nil
}

// verifyAssertion verifies the JWT assertion.
func (as *authorizeService) verifyAssertion(assertion string) error {
	if err := as.jwtService.VerifyJWT(assertion, "", ""); err != nil {
		as.logger.Debug("Invalid assertion signature", log.String("error", err.Error))
		return errors.New("invalid assertion signature")
	}
	return nil
}

// decodeAttributesFromAssertion decodes user attributes from the flow assertion JWT.
func decodeAttributesFromAssertion(assertion string) (assertionClaims, time.Time, error) {
	claims := assertionClaims{
		userAttributes: make(map[string]interface{}),
	}

	_, jwtPayload, err := jwt.DecodeJWT(assertion)
	if err != nil {
		return claims, time.Time{}, errors.New("Failed to decode the JWT token: " + err.Error())
	}

	// Extract authentication time from iat claim.
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
			return claims, time.Time{}, errors.New("JWT 'iat' claim has unexpected type")
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
		// Extract sub claim.
		if key == oauth2const.ClaimSub {
			if strValue, ok := value.(string); ok {
				claims.userID = strValue
			} else {
				return claims, time.Time{}, errors.New("JWT 'sub' claim is not a string")
			}
			continue
		}

		// Extract authorized_permissions claim.
		if key == "authorized_permissions" {
			if strValue, ok := value.(string); ok {
				claims.authorizedPermissions = strValue
			}
			continue
		}

		// Skip standard JWT claims.
		if standardClaims[key] {
			continue
		}

		// All other claims are treated as user attributes.
		userAttributes[key] = value
	}
	claims.userAttributes = userAttributes

	return claims, authTime, nil
}

// createAuthorizationCode generates an authorization code based on the provided
// authorization request context and authenticated user.
func createAuthorizationCode(
	authRequestCtx *authRequestContext,
	claims *assertionClaims,
	authTime time.Time,
) (AuthorizationCode, error) {
	clientID := authRequestCtx.OAuthParameters.ClientID
	redirectURI := authRequestCtx.OAuthParameters.RedirectURI

	if clientID == "" || redirectURI == "" {
		return AuthorizationCode{}, errors.New("client_id or redirect_uri is missing")
	}

	if claims.userID == "" {
		return AuthorizationCode{}, errors.New("authenticated user not found")
	}

	// Use provided authTime, or fallback to current time if zero (iat claim was not available).
	if authTime.IsZero() {
		authTime = time.Now()
	}

	standardScopes := authRequestCtx.OAuthParameters.StandardScopes
	permissionScopes := authRequestCtx.OAuthParameters.PermissionScopes
	allScopes := append(append([]string{}, standardScopes...), permissionScopes...)
	resource := authRequestCtx.OAuthParameters.Resource

	oauthConfig := config.GetThunderRuntime().Config.OAuth
	validityPeriod := oauthConfig.AuthorizationCode.ValidityPeriod
	expiryTime := authTime.Add(time.Duration(validityPeriod) * time.Second)

	codeID, err := utils.GenerateUUIDv7()
	if err != nil {
		return AuthorizationCode{}, errors.New("failed to generate UUID")
	}

	code, err := utils.GenerateUUIDv7()
	if err != nil {
		return AuthorizationCode{}, errors.New("failed to generate UUID")
	}

	return AuthorizationCode{
		CodeID:              codeID,
		Code:                code,
		ClientID:            clientID,
		RedirectURI:         redirectURI,
		AuthorizedUserID:    claims.userID,
		UserAttributes:      claims.userAttributes,
		TimeCreated:         authTime,
		ExpiryTime:          expiryTime,
		Scopes:              utils.StringifyStringArray(allScopes, " "),
		State:               AuthCodeStateActive,
		CodeChallenge:       authRequestCtx.OAuthParameters.CodeChallenge,
		CodeChallengeMethod: authRequestCtx.OAuthParameters.CodeChallengeMethod,
		Resource:            resource,
		ClaimsRequest:       authRequestCtx.OAuthParameters.ClaimsRequest,
		ClaimsLocales:       authRequestCtx.OAuthParameters.ClaimsLocales,
		Nonce:               authRequestCtx.OAuthParameters.Nonce,
	}, nil
}

// getRequiredAttributes computes the required attributes based on OIDC scopes, access token config,
// and claims parameter.
func getRequiredAttributes(oidcScopes []string, app *appmodel.OAuthAppConfigProcessedDTO,
	claimsRequest *oauth2model.ClaimsRequest) string {
	requiredAttrsSet := make(map[string]bool)

	// Early return if no app config.
	if app == nil || app.Token == nil {
		return ""
	}

	// Check if openid scope is present.
	hasOpenIDScope := slices.Contains(oidcScopes, oauth2const.ScopeOpenID)

	// Process OIDC-related claims only if openid scope is present.
	if hasOpenIDScope {
		// Build allowed attributes set and scope claims mapping from ID token config.
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

		// Add claims from claims parameter for ID token.
		if claimsRequest != nil && claimsRequest.IDToken != nil && idTokenAllowedSet != nil {
			for claimName := range claimsRequest.IDToken {
				if idTokenAllowedSet[claimName] {
					requiredAttrsSet[claimName] = true
				}
			}
		}

		// Add claims from OIDC scopes.
		for _, scope := range oidcScopes {
			var scopeClaims []string

			// Check app-specific scope claims first.
			if scopeClaimsMapping != nil {
				if appClaims, exists := scopeClaimsMapping[scope]; exists {
					scopeClaims = appClaims
				}
			}

			// Fall back to standard OIDC scopes if no app-specific mapping.
			if scopeClaims == nil {
				if standardScope, exists := oauth2const.StandardOIDCScopes[scope]; exists {
					scopeClaims = standardScope.Claims
				}
			}

			// Add claims to the set if they are allowed in ID token according to app config.
			for _, claim := range scopeClaims {
				if idTokenAllowedSet != nil && idTokenAllowedSet[claim] {
					requiredAttrsSet[claim] = true
				}
			}
		}
	}

	// Add access token attributes from app config.
	if app.Token.AccessToken != nil && len(app.Token.AccessToken.UserAttributes) > 0 {
		for _, attr := range app.Token.AccessToken.UserAttributes {
			requiredAttrsSet[attr] = true
		}
	}

	return mapKeysToSpaceSeparatedString(requiredAttrsSet)
}

// mapKeysToSpaceSeparatedString converts map keys to a space-separated string.
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

// validateSubClaimConstraint validates the sub claim constraint if specified in the claims parameter.
func validateSubClaimConstraint(claimsRequest *oauth2model.ClaimsRequest, actualSubject string) error {
	if claimsRequest == nil {
		return nil
	}

	// Check id_token sub claim constraint.
	if claimsRequest.IDToken != nil {
		if subReq, exists := claimsRequest.IDToken["sub"]; exists && subReq != nil {
			if !subReq.MatchesValue(actualSubject) {
				return errors.New("sub claim in id_token does not match requested value")
			}
		}
	}

	// Check userinfo sub claim constraint.
	if claimsRequest.UserInfo != nil {
		if subReq, exists := claimsRequest.UserInfo["sub"]; exists && subReq != nil {
			if !subReq.MatchesValue(actualSubject) {
				return errors.New("sub claim in userinfo does not match requested value")
			}
		}
	}

	return nil
}
