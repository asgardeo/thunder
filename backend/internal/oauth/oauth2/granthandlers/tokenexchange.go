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
	"fmt"
	"net/url"
	"strings"
	"time"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/model"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/jwt"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
)

// tokenExchangeGrantHandler handles the token exchange grant type.
type tokenExchangeGrantHandler struct {
	jwtService jwt.JWTServiceInterface
}

// newTokenExchangeGrantHandler creates a new instance of tokenExchangeGrantHandler.
func newTokenExchangeGrantHandler(
	jwtService jwt.JWTServiceInterface,
) GrantHandlerInterface {
	return &tokenExchangeGrantHandler{
		jwtService: jwtService,
	}
}

// ValidateGrant validates the token exchange grant type request.
func (h *tokenExchangeGrantHandler) ValidateGrant(tokenRequest *model.TokenRequest,
	oauthApp *appmodel.OAuthAppConfigProcessedDTO) *model.ErrorResponse {
	if constants.GrantType(tokenRequest.GrantType) != constants.GrantTypeTokenExchange {
		return &model.ErrorResponse{
			Error:            constants.ErrorUnsupportedGrantType,
			ErrorDescription: "Unsupported grant type",
		}
	}

	if tokenRequest.SubjectToken == "" {
		return &model.ErrorResponse{
			Error:            constants.ErrorInvalidRequest,
			ErrorDescription: "Missing required parameter: subject_token",
		}
	}

	if tokenRequest.SubjectTokenType == "" {
		return &model.ErrorResponse{
			Error:            constants.ErrorInvalidRequest,
			ErrorDescription: "Missing required parameter: subject_token_type",
		}
	}

	if !constants.TokenTypeIdentifier(tokenRequest.SubjectTokenType).IsValid() {
		return &model.ErrorResponse{
			Error:            constants.ErrorInvalidRequest,
			ErrorDescription: fmt.Sprintf("Unsupported subject_token_type: %s", tokenRequest.SubjectTokenType),
		}
	}

	if tokenRequest.Resource != "" {
		if !utils.IsValidURI(tokenRequest.Resource) {
			return &model.ErrorResponse{
				Error:            constants.ErrorInvalidRequest,
				ErrorDescription: "Invalid resource parameter: must be an absolute URI",
			}
		}
		parsedURI, err := url.Parse(tokenRequest.Resource)
		if err != nil || parsedURI.Fragment != "" {
			return &model.ErrorResponse{
				Error:            constants.ErrorInvalidRequest,
				ErrorDescription: "Invalid resource parameter: must not contain a fragment component",
			}
		}
	}

	if tokenRequest.ActorToken != "" && tokenRequest.ActorTokenType == "" {
		return &model.ErrorResponse{
			Error:            constants.ErrorInvalidRequest,
			ErrorDescription: "actor_token_type is required when actor_token is provided",
		}
	}

	if tokenRequest.ActorTokenType != "" {
		if tokenRequest.ActorToken == "" {
			return &model.ErrorResponse{
				Error:            constants.ErrorInvalidRequest,
				ErrorDescription: "actor_token_type must not be provided without actor_token",
			}
		}
		if !constants.TokenTypeIdentifier(tokenRequest.ActorTokenType).IsValid() {
			return &model.ErrorResponse{
				Error:            constants.ErrorInvalidRequest,
				ErrorDescription: fmt.Sprintf("Unsupported actor_token_type: %s", tokenRequest.ActorTokenType),
			}
		}
	}

	if tokenRequest.RequestedTokenType != "" {
		requestedType := constants.TokenTypeIdentifier(tokenRequest.RequestedTokenType)

		if !requestedType.IsValid() {
			return &model.ErrorResponse{
				Error:            constants.ErrorInvalidRequest,
				ErrorDescription: fmt.Sprintf("Unsupported requested_token_type: %s", tokenRequest.RequestedTokenType),
			}
		}
		// TODO: Add support for other token types if needed
		if requestedType != constants.TokenTypeIdentifierAccessToken &&
			requestedType != constants.TokenTypeIdentifierJWT {
			return &model.ErrorResponse{
				Error: constants.ErrorInvalidTarget,
				ErrorDescription: fmt.Sprintf(
					"Requested token type '%s' is not supported. Only access tokens and JWT tokens are supported.",
					tokenRequest.RequestedTokenType,
				),
			}
		}
	}

	return nil
}

// HandleGrant handles the token exchange grant type.
func (h *tokenExchangeGrantHandler) HandleGrant(tokenRequest *model.TokenRequest,
	oauthApp *appmodel.OAuthAppConfigProcessedDTO, ctx *model.TokenContext) (
	*model.TokenResponseDTO, *model.ErrorResponse) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "TokenExchangeGrantHandler"))

	subjectClaims, err := h.validateAndExtractClaims(tokenRequest.SubjectToken)
	if err != nil {
		logger.Error("Failed to validate subject token", log.Error(err))
		return nil, &model.ErrorResponse{
			Error:            constants.ErrorInvalidGrant,
			ErrorDescription: fmt.Sprintf("Invalid subject_token: %s", err.Error()),
		}
	}

	var actorClaims map[string]interface{}
	if tokenRequest.ActorToken != "" {
		actorClaims, err = h.validateAndExtractClaims(tokenRequest.ActorToken)
		if err != nil {
			logger.Error("Failed to validate actor token", log.Error(err))
			return nil, &model.ErrorResponse{
				Error:            constants.ErrorInvalidGrant,
				ErrorDescription: fmt.Sprintf("Invalid actor_token: %s", err.Error()),
			}
		}
	}

	finalScopes, errResp := h.getScopes(tokenRequest, subjectClaims)
	if errResp != nil {
		return nil, errResp
	}

	finalAudience := h.getAudience(tokenRequest, subjectClaims)

	userAttributes := h.extractUserAttributes(subjectClaims)
	subjectActor, _ := subjectClaims["act"].(map[string]interface{})
	newClaims := h.buildTokenClaims(
		tokenRequest,
		actorClaims,
		subjectActor,
		userAttributes,
		finalScopes,
	)

	iss := ""
	validityPeriod := int64(0)
	if oauthApp.Token != nil && oauthApp.Token.AccessToken != nil {
		iss = oauthApp.Token.AccessToken.Issuer
		validityPeriod = oauthApp.Token.AccessToken.ValidityPeriod
	}
	if iss == "" {
		iss = config.GetThunderRuntime().Config.JWT.Issuer
	}
	if validityPeriod == 0 {
		validityPeriod = config.GetThunderRuntime().Config.JWT.ValidityPeriod
	}

	sub, _ := subjectClaims["sub"].(string)
	token, _, err := h.jwtService.GenerateJWT(
		sub,
		finalAudience,
		iss,
		validityPeriod,
		newClaims,
	)
	if err != nil {
		logger.Error("Failed to generate token", log.Error(err))
		return nil, &model.ErrorResponse{
			Error:            constants.ErrorServerError,
			ErrorDescription: "Failed to generate token",
		}
	}

	if ctx.TokenAttributes == nil {
		ctx.TokenAttributes = make(map[string]interface{})
	}

	// Store the issued token type in context
	requestedTokenType := tokenRequest.RequestedTokenType
	tokenTypeID := constants.TokenTypeIdentifier(requestedTokenType)
	isAccessToken := requestedTokenType == "" || tokenTypeID == constants.TokenTypeIdentifierAccessToken
	if isAccessToken {
		ctx.TokenAttributes["issued_token_type"] = string(constants.TokenTypeIdentifierAccessToken)
	} else {
		ctx.TokenAttributes["issued_token_type"] = string(constants.TokenTypeIdentifierJWT)
	}

	accessToken := &model.TokenDTO{
		Token:          token,
		TokenType:      constants.TokenTypeBearer,
		IssuedAt:       time.Now().Unix(),
		ExpiresIn:      validityPeriod,
		Scopes:         finalScopes,
		ClientID:       tokenRequest.ClientID,
		UserAttributes: userAttributes,
	}

	return &model.TokenResponseDTO{
		AccessToken: *accessToken,
	}, nil
}

// getScopes validates and determines the scopes for the new token.
// TODO: Revise logic after scope implementation.
func (h *tokenExchangeGrantHandler) getScopes(
	tokenRequest *model.TokenRequest,
	subjectClaims map[string]interface{},
) ([]string, *model.ErrorResponse) {
	subjectScope, _ := subjectClaims["scope"].(string)
	subjectScopes := parseScopes(subjectScope)

	// If no scopes requested, return subject scopes
	if tokenRequest.Scope == "" {
		return subjectScopes, nil
	}

	requestedScopes := parseScopes(tokenRequest.Scope)

	if len(requestedScopes) == 0 {
		return []string{}, nil
	}

	// If subject token has no scopes, reject requests asking for scopes
	if len(subjectScopes) == 0 {
		return nil, &model.ErrorResponse{
			Error:            constants.ErrorInvalidScope,
			ErrorDescription: "Cannot request scopes when subject token has no scopes",
		}
	}

	// Validate requested scopes are subset of subject scopes
	subjectScopeSet := make(map[string]bool)
	for _, s := range subjectScopes {
		subjectScopeSet[s] = true
	}

	for _, requestedScope := range requestedScopes {
		if !subjectScopeSet[requestedScope] {
			return nil, &model.ErrorResponse{
				Error: constants.ErrorInvalidScope,
				ErrorDescription: fmt.Sprintf(
					"Requested scope '%s' not present in subject token",
					requestedScope,
				),
			}
		}
	}

	return requestedScopes, nil
}

// getAudience determines the audience for the new token.
func (h *tokenExchangeGrantHandler) getAudience(
	tokenRequest *model.TokenRequest,
	subjectClaims map[string]interface{},
) string {
	if tokenRequest.Audience != "" {
		return tokenRequest.Audience
	}

	if tokenRequest.Resource != "" {
		return tokenRequest.Resource
	}

	if aud, ok := subjectClaims["aud"].(string); ok && aud != "" {
		return aud
	}

	return tokenRequest.ClientID
}

// buildTokenClaims constructs the claims for the new token.
func (h *tokenExchangeGrantHandler) buildTokenClaims(
	tokenRequest *model.TokenRequest,
	actorClaims map[string]interface{},
	subjectActor map[string]interface{},
	userAttributes map[string]interface{},
	scopes []string,
) map[string]interface{} {
	newClaims := make(map[string]interface{})

	if len(scopes) > 0 {
		newClaims["scope"] = strings.Join(scopes, " ")
	}

	newClaims["client_id"] = tokenRequest.ClientID

	if actorClaims != nil {
		actClaim := map[string]interface{}{}
		if sub, ok := actorClaims["sub"].(string); ok {
			actClaim["sub"] = sub
		}
		if iss, ok := actorClaims["iss"].(string); ok && iss != "" {
			actClaim["iss"] = iss
		}

		// If actor token has its own act claim, preserve it in the chain
		if actorAct, ok := actorClaims["act"].(map[string]interface{}); ok {
			if subjectActor != nil {
				nestedAct := map[string]interface{}{}
				for k, v := range actorAct {
					nestedAct[k] = v
				}
				nestedAct["act"] = subjectActor
				actClaim["act"] = nestedAct
			} else {
				actClaim["act"] = actorAct
			}
		} else if subjectActor != nil {
			actClaim["act"] = subjectActor
		}
		newClaims["act"] = actClaim
	} else if subjectActor != nil {
		newClaims["act"] = subjectActor
	}

	for key, value := range userAttributes {
		newClaims[key] = value
	}

	return newClaims
}

// validateAndExtractClaims validates a token and extracts its claims.
func (h *tokenExchangeGrantHandler) validateAndExtractClaims(token string) (map[string]interface{}, error) {
	// TODO: Validate for external token issuers
	if err := h.jwtService.VerifyJWTSignature(token); err != nil {
		return nil, fmt.Errorf("invalid token signature: %w", err)
	}

	claims, err := jwt.DecodeJWTPayload(token)
	if err != nil {
		return nil, fmt.Errorf("failed to decode token: %w", err)
	}

	if _, ok := claims["sub"].(string); !ok {
		return nil, fmt.Errorf("missing or invalid 'sub' claim")
	}

	now := time.Now().Unix()

	if exp, ok := claims["exp"].(float64); ok {
		if int64(exp) < now {
			return nil, fmt.Errorf("token has expired")
		}
	}

	if nbf, ok := claims["nbf"].(float64); ok {
		if int64(nbf) > now {
			return nil, fmt.Errorf("token not yet valid")
		}
	}

	return claims, nil
}

// extractUserAttributes extracts user attributes from claims, excluding standard JWT claims.
func (h *tokenExchangeGrantHandler) extractUserAttributes(claims map[string]interface{}) map[string]interface{} {
	standardClaims := map[string]bool{
		"sub": true, "iss": true, "aud": true, "exp": true,
		"nbf": true, "iat": true, "jti": true, "scope": true,
		"client_id": true, "act": true,
	}

	userAttributes := make(map[string]interface{})
	for key, value := range claims {
		if !standardClaims[key] {
			userAttributes[key] = value
		}
	}

	return userAttributes
}

// parseScopes parses a space-separated scope string into a slice of scopes.
func parseScopes(scopeStr string) []string {
	if scopeStr == "" {
		return []string{}
	}

	scopes := []string{}
	for _, s := range strings.Split(scopeStr, " ") {
		trimmed := strings.TrimSpace(s)
		if trimmed != "" {
			scopes = append(scopes, trimmed)
		}
	}
	return scopes
}
