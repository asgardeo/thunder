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

package granthandlers

import (
	"encoding/json"
	"slices"
	"strings"
	"time"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/authz"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/pkce"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/jwt"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/user"
)

// TODO: Temporary constant, move to a common place/ use a different strategy.
const (
	// UserAttributeGroups is the constant for user's groups.
	UserAttributeGroups = "groups"
	// DefaultGroupListLimit is the default limit for group list retrieval.
	DefaultGroupListLimit = 20
)

// authorizationCodeGrantHandler handles the authorization code grant type.
type authorizationCodeGrantHandler struct {
	jwtService   jwt.JWTServiceInterface
	authzService authz.AuthorizeServiceInterface
	userService  user.UserServiceInterface
}

// newAuthorizationCodeGrantHandler creates a new instance of AuthorizationCodeGrantHandler.
func newAuthorizationCodeGrantHandler(
	jwtService jwt.JWTServiceInterface,
	userService user.UserServiceInterface,
	authzService authz.AuthorizeServiceInterface,
) GrantHandlerInterface {
	return &authorizationCodeGrantHandler{
		jwtService:   jwtService,
		authzService: authzService,
		userService:  userService,
	}
}

// ValidateGrant validates the authorization code grant request.
func (h *authorizationCodeGrantHandler) ValidateGrant(tokenRequest *model.TokenRequest,
	oauthApp *appmodel.OAuthAppConfigProcessedDTO) *model.ErrorResponse {
	if tokenRequest.GrantType == "" {
		return &model.ErrorResponse{
			Error:            constants.ErrorInvalidRequest,
			ErrorDescription: "Missing grant type",
		}
	}
	if constants.GrantType(tokenRequest.GrantType) != constants.GrantTypeAuthorizationCode {
		return &model.ErrorResponse{
			Error:            constants.ErrorUnsupportedGrantType,
			ErrorDescription: "Unsupported grant type",
		}
	}
	if tokenRequest.Code == "" {
		return &model.ErrorResponse{
			Error:            constants.ErrorInvalidGrant,
			ErrorDescription: "Authorization code is required",
		}
	}
	if tokenRequest.ClientID == "" {
		return &model.ErrorResponse{
			Error:            constants.ErrorInvalidClient,
			ErrorDescription: "Client Id is required",
		}
	}

	// TODO: Redirect uri is not mandatory when excluded in the authorize request and is valid scenario.
	//  This should be removed when supporting other means of authorization.
	if tokenRequest.RedirectURI == "" {
		return &model.ErrorResponse{
			Error:            constants.ErrorInvalidRequest,
			ErrorDescription: "Redirect URI is required",
		}
	}

	// Validate the authorization code.
	if tokenRequest.Code == "" {
		return &model.ErrorResponse{
			Error:            constants.ErrorInvalidRequest,
			ErrorDescription: "Authorization code is required",
		}
	}

	return nil
}

// HandleGrant processes the authorization code grant request and generates a token response.
func (h *authorizationCodeGrantHandler) HandleGrant(tokenRequest *model.TokenRequest,
	oauthApp *appmodel.OAuthAppConfigProcessedDTO, ctx *model.TokenContext) (
	*model.TokenResponseDTO, *model.ErrorResponse) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "AuthorizationCodeGrantHandler"))

	// Retrieve and validate authorization code
	authCode, errResponse := h.retrieveAndValidateAuthCode(tokenRequest, oauthApp, logger)
	if errResponse != nil {
		return nil, errResponse
	}

	// Parse authorized scopes
	authorizedScopes, authorizedScopesStr := parseAuthorizedScopes(authCode.Scopes)

	// Fetch user attributes and groups
	attrs, userGroups, errResponse := h.fetchUserAttributesAndGroups(authCode.AuthorizedUserID, oauthApp, logger)
	if errResponse != nil {
		return nil, errResponse
	}

	// Build access token claims and attributes
	jwtClaims, accessTokenAttributes := buildAccessTokenClaimsAndAttributes(
		authorizedScopesStr, attrs, userGroups, oauthApp, authCode.Resource)

	// Generate access token
	iss, validityPeriod := resolveTokenConfig(oauthApp)
	token, _, err := h.jwtService.GenerateJWT(authCode.AuthorizedUserID, authCode.ClientID,
		iss, validityPeriod, jwtClaims)
	if err != nil {
		return nil, &model.ErrorResponse{
			Error:            constants.ErrorServerError,
			ErrorDescription: "Failed to generate token",
		}
	}

	// Update context attributes
	updateContextAttributes(ctx, authCode)

	// Build token response
	tokenResponse := buildTokenResponse(token, authorizedScopes, tokenRequest.ClientID, accessTokenAttributes)

	// Generate ID token if 'openid' scope is present
	if slices.Contains(authorizedScopes, "openid") {
		idToken, errResponse := h.generateIDToken(authCode, tokenRequest, authorizedScopes, attrs, oauthApp)
		if errResponse != nil {
			return nil, errResponse
		}
		tokenResponse.IDToken = *idToken
	}

	return tokenResponse, nil
}

// retrieveAndValidateAuthCode retrieves and validates the authorization code, including PKCE validation.
func (h *authorizationCodeGrantHandler) retrieveAndValidateAuthCode(
	tokenRequest *model.TokenRequest,
	oauthApp *appmodel.OAuthAppConfigProcessedDTO,
	logger *log.Logger,
) (*authz.AuthorizationCode, *model.ErrorResponse) {
	authCode, codeErr := h.authzService.GetAuthorizationCodeDetails(tokenRequest.ClientID, tokenRequest.Code)
	if codeErr != nil {
		return nil, &model.ErrorResponse{
			Error:            constants.ErrorInvalidGrant,
			ErrorDescription: "Invalid authorization code",
		}
	}

	// Validate the retrieved authorization code
	errResponse := validateAuthorizationCode(tokenRequest, *authCode)
	if errResponse != nil && errResponse.Error != "" {
		return nil, errResponse
	}

	// Validate PKCE if required or if code challenge was provided during authorization
	if oauthApp.RequiresPKCE() || authCode.CodeChallenge != "" {
		if tokenRequest.CodeVerifier == "" {
			return nil, &model.ErrorResponse{
				Error:            constants.ErrorInvalidRequest,
				ErrorDescription: "code_verifier is required",
			}
		}

		// Validate PKCE
		if err := pkce.ValidatePKCE(authCode.CodeChallenge, authCode.CodeChallengeMethod,
			tokenRequest.CodeVerifier); err != nil {
			logger.Debug("PKCE validation failed", log.Error(err))
			return nil, &model.ErrorResponse{
				Error:            constants.ErrorInvalidGrant,
				ErrorDescription: "Invalid code verifier",
			}
		}
	}
	return authCode, nil
}

// parseAuthorizedScopes parses the authorized scopes from the authorization code.
func parseAuthorizedScopes(scopesStr string) ([]string, string) {
	authorizedScopesStr := strings.TrimSpace(scopesStr)
	authorizedScopes := []string{}
	if authorizedScopesStr != "" {
		authorizedScopes = strings.Split(authorizedScopesStr, " ")
	}
	return authorizedScopes, authorizedScopesStr
}

// fetchUserAttributesAndGroups fetches user attributes and groups if required.
func (h *authorizationCodeGrantHandler) fetchUserAttributesAndGroups(
	userID string,
	oauthApp *appmodel.OAuthAppConfigProcessedDTO,
	logger *log.Logger,
) (map[string]interface{}, []string, *model.ErrorResponse) {
	var attrs map[string]interface{}
	userGroups := make([]string, 0)

	// Check if user attributes or groups are required
	if len(oauthApp.Token.AccessToken.UserAttributes) == 0 &&
		(oauthApp.Token.IDToken == nil || len(oauthApp.Token.IDToken.UserAttributes) == 0) {
		return attrs, userGroups, nil
	}

	// Fetch user attributes
	user, svcErr := h.userService.GetUser(userID)
	if svcErr != nil {
		logger.Error("Failed to fetch user attributes", log.String("userID", userID), log.Any("error", svcErr))
		return nil, nil, &model.ErrorResponse{
			Error:            constants.ErrorServerError,
			ErrorDescription: "Something went wrong",
		}
	}

	if err := json.Unmarshal(user.Attributes, &attrs); err != nil {
		logger.Error("Failed to unmarshal user attributes", log.String("userID", userID), log.Error(err))
		return nil, nil, &model.ErrorResponse{
			Error:            constants.ErrorServerError,
			ErrorDescription: "Something went wrong",
		}
	}

	// Fetch user groups if required
	needsGroups := slices.Contains(oauthApp.Token.AccessToken.UserAttributes, UserAttributeGroups) ||
		(oauthApp.Token.IDToken != nil && slices.Contains(oauthApp.Token.IDToken.UserAttributes, UserAttributeGroups))

	if needsGroups {
		groups, svcErr := h.userService.GetUserGroups(userID, DefaultGroupListLimit, 0)
		if svcErr != nil {
			logger.Error("Failed to fetch user groups", log.String("userID", userID), log.Any("error", svcErr))
			return nil, nil, &model.ErrorResponse{
				Error:            constants.ErrorServerError,
				ErrorDescription: "Something went wrong",
			}
		}

		for _, group := range groups.Groups {
			userGroups = append(userGroups, group.Name)
		}
	}

	return attrs, userGroups, nil
}

// buildAccessTokenClaimsAndAttributes builds JWT claims and user attributes for the access token.
func buildAccessTokenClaimsAndAttributes(
	authorizedScopesStr string,
	attrs map[string]interface{},
	userGroups []string,
	oauthApp *appmodel.OAuthAppConfigProcessedDTO,
	resource string,
) (map[string]interface{}, map[string]interface{}) {
	jwtClaims := make(map[string]interface{})
	accessTokenAttributes := make(map[string]interface{})

	// Add scope to JWT claims
	if authorizedScopesStr != "" {
		jwtClaims["scope"] = authorizedScopesStr
	}

	// Add audience claim based on resource parameter (RFC 8707)
	if resource != "" {
		jwtClaims["aud"] = resource
	}

	// Add user attributes to claims and access token attributes
	if len(oauthApp.Token.AccessToken.UserAttributes) > 0 && attrs != nil {
		for _, attr := range oauthApp.Token.AccessToken.UserAttributes {
			if val, ok := attrs[attr]; ok {
				jwtClaims[attr] = val
				accessTokenAttributes[attr] = val
			}
		}
	}

	// Handle user groups
	if len(userGroups) > 0 {
		if slices.Contains(oauthApp.Token.AccessToken.UserAttributes, UserAttributeGroups) {
			jwtClaims[UserAttributeGroups] = userGroups
			accessTokenAttributes[UserAttributeGroups] = userGroups
		}
		if oauthApp.Token.IDToken != nil &&
			slices.Contains(oauthApp.Token.IDToken.UserAttributes, UserAttributeGroups) {
			attrs[UserAttributeGroups] = userGroups
		}
	}

	return jwtClaims, accessTokenAttributes
}

// resolveTokenConfig resolves the issuer and validity period for the access token.
func resolveTokenConfig(oauthApp *appmodel.OAuthAppConfigProcessedDTO) (string, int64) {
	iss := ""
	validityPeriod := int64(0)

	// Get issuer from token config
	if oauthApp.Token != nil && oauthApp.Token.Issuer != "" {
		iss = oauthApp.Token.Issuer
	} else {
		iss = config.GetThunderRuntime().Config.JWT.Issuer
	}

	// Get validity period from access token config
	if oauthApp.Token != nil && oauthApp.Token.AccessToken != nil {
		validityPeriod = oauthApp.Token.AccessToken.ValidityPeriod
	} else {
		validityPeriod = config.GetThunderRuntime().Config.JWT.ValidityPeriod
	}

	return iss, validityPeriod
}

// updateContextAttributes updates the token context with subject and audience.
func updateContextAttributes(ctx *model.TokenContext, authCode *authz.AuthorizationCode) {
	if ctx.TokenAttributes == nil {
		ctx.TokenAttributes = make(map[string]interface{})
	}
	ctx.TokenAttributes[constants.ClaimSub] = authCode.AuthorizedUserID
	if authCode.Resource != "" {
		ctx.TokenAttributes[constants.ClaimAud] = authCode.Resource
	} else {
		ctx.TokenAttributes[constants.ClaimAud] = authCode.ClientID
	}
}

// buildTokenResponse builds the token response with access token details.
func buildTokenResponse(
	token string,
	authorizedScopes []string,
	clientID string,
	accessTokenAttributes map[string]interface{},
) *model.TokenResponseDTO {
	accessToken := &model.TokenDTO{
		Token:          token,
		TokenType:      constants.TokenTypeBearer,
		IssuedAt:       time.Now().Unix(),
		ExpiresIn:      3600,
		Scopes:         authorizedScopes,
		ClientID:       clientID,
		UserAttributes: accessTokenAttributes,
	}

	return &model.TokenResponseDTO{
		AccessToken: *accessToken,
	}
}

// validateAuthorizationCode validates the authorization code against the token request.
func validateAuthorizationCode(tokenRequest *model.TokenRequest,
	code authz.AuthorizationCode) *model.ErrorResponse {
	if tokenRequest.ClientID != code.ClientID {
		return &model.ErrorResponse{
			Error:            constants.ErrorInvalidClient,
			ErrorDescription: "Invalid client Id",
		}
	}

	// redirect_uri is not mandatory in certain scenarios. Should match if provided with the authorization.
	if code.RedirectURI != "" && tokenRequest.RedirectURI != code.RedirectURI {
		return &model.ErrorResponse{
			Error:            constants.ErrorInvalidGrant,
			ErrorDescription: "Invalid redirect URI",
		}
	}

	// Validate resource parameter consistency with authorization code
	if code.Resource != "" && code.Resource != tokenRequest.Resource {
		return &model.ErrorResponse{
			Error:            constants.ErrorInvalidTarget,
			ErrorDescription: "Resource parameter mismatch",
		}
	}

	if code.State == authz.AuthCodeStateInactive {
		// TODO: Revoke all the tokens issued for this authorization code.

		return &model.ErrorResponse{
			Error:            constants.ErrorInvalidGrant,
			ErrorDescription: "Inactive authorization code",
		}
	} else if code.State != authz.AuthCodeStateActive {
		return &model.ErrorResponse{
			Error:            constants.ErrorInvalidGrant,
			ErrorDescription: "Inactive authorization code",
		}
	}

	if code.ExpiryTime.Before(time.Now()) {
		return &model.ErrorResponse{
			Error:            constants.ErrorInvalidGrant,
			ErrorDescription: "Expired authorization code",
		}
	}

	return nil
}

// getIDTokenClaims generates ID token claims based on scopes and application configuration
func getIDTokenClaims(scopes []string, userAttributes map[string]interface{},
	oauthApp *appmodel.OAuthAppConfigProcessedDTO) map[string]interface{} {
	claims := make(map[string]interface{})
	now := time.Now().Unix()
	claims[constants.ClaimAuthTime] = now

	var idTokenUserAttributes []string
	if oauthApp.Token != nil && oauthApp.Token.IDToken != nil {
		idTokenUserAttributes = oauthApp.Token.IDToken.UserAttributes
	}

	if len(idTokenUserAttributes) == 0 {
		return claims
	}

	for _, scope := range scopes {
		var scopeClaims []string

		if oauthApp.Token != nil && oauthApp.Token.IDToken != nil &&
			oauthApp.Token.IDToken.ScopeClaims != nil {
			if appClaims, exists := oauthApp.Token.IDToken.ScopeClaims[scope]; exists {
				scopeClaims = appClaims
			}
		}

		if scopeClaims == nil {
			if scope, exists := constants.StandardOIDCScopes[scope]; exists {
				scopeClaims = scope.Claims
			}
		}

		for _, claim := range scopeClaims {
			if slices.Contains(idTokenUserAttributes, claim) && userAttributes[claim] != nil {
				claims[claim] = userAttributes[claim]
			}
		}
	}

	return claims
}

// generateIDToken generates an ID token for the given authorization code and scopes
func (h *authorizationCodeGrantHandler) generateIDToken(authCode *authz.AuthorizationCode,
	tokenRequest *model.TokenRequest, authorizedScopes []string, attrs map[string]interface{},
	oauthApp *appmodel.OAuthAppConfigProcessedDTO) (*model.TokenDTO, *model.ErrorResponse) {
	idTokenClaims := getIDTokenClaims(authorizedScopes, attrs, oauthApp)

	// Resolve ID token issuer and validity period
	idTokenIss := ""
	idTokenValidityPeriod := int64(0)

	// Get issuer from token config
	if oauthApp.Token != nil && oauthApp.Token.Issuer != "" {
		idTokenIss = oauthApp.Token.Issuer
	} else {
		idTokenIss = config.GetThunderRuntime().Config.JWT.Issuer
	}

	// Get validity period from ID token config
	if oauthApp.Token != nil && oauthApp.Token.IDToken != nil {
		idTokenValidityPeriod = oauthApp.Token.IDToken.ValidityPeriod
	}
	if idTokenValidityPeriod == 0 {
		idTokenValidityPeriod = config.GetThunderRuntime().Config.JWT.ValidityPeriod
	}

	// Generate ID token JWT
	idToken, _, err := h.jwtService.GenerateJWT(authCode.AuthorizedUserID, authCode.ClientID,
		idTokenIss, idTokenValidityPeriod, idTokenClaims)
	if err != nil {
		return nil, &model.ErrorResponse{
			Error:            constants.ErrorServerError,
			ErrorDescription: "Failed to generate ID token",
		}
	}

	// Create ID token DTO
	idTokenDTO := &model.TokenDTO{
		Token:          idToken,
		TokenType:      constants.TokenTypeJWT,
		IssuedAt:       time.Now().Unix(),
		ExpiresIn:      idTokenValidityPeriod,
		Scopes:         authorizedScopes,
		ClientID:       tokenRequest.ClientID,
		UserAttributes: idTokenClaims,
	}

	return idTokenDTO, nil
}
