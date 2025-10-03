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
	"slices"
	"strings"
	"time"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/model"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/jwt"
	"github.com/asgardeo/thunder/internal/system/log"
	userservice "github.com/asgardeo/thunder/internal/user/service"
)

// refreshTokenGrantHandler handles the refresh token grant type.
type refreshTokenGrantHandler struct {
	JWTService  jwt.JWTServiceInterface
	UserService userservice.UserServiceInterface
}

// newRefreshTokenGrantHandler creates a new instance of RefreshTokenGrantHandler.
func newRefreshTokenGrantHandler() RefreshTokenGrantHandlerInterface {
	return &refreshTokenGrantHandler{
		JWTService:  jwt.GetJWTService(),
		UserService: userservice.GetUserService(),
	}
}

// ValidateGrant validates the refresh token grant request.
func (h *refreshTokenGrantHandler) ValidateGrant(tokenRequest *model.TokenRequest,
	oauthApp *appmodel.OAuthAppConfigProcessedDTO) *model.ErrorResponse {
	if constants.GrantType(tokenRequest.GrantType) != constants.GrantTypeRefreshToken {
		return &model.ErrorResponse{
			Error:            constants.ErrorUnsupportedGrantType,
			ErrorDescription: "Unsupported grant type",
		}
	}
	if tokenRequest.RefreshToken == "" {
		return &model.ErrorResponse{
			Error:            constants.ErrorInvalidRequest,
			ErrorDescription: "Refresh token is required",
		}
	}
	if tokenRequest.ClientID == "" {
		return &model.ErrorResponse{
			Error:            constants.ErrorInvalidRequest,
			ErrorDescription: "Client ID is required",
		}
	}

	return nil
}

// HandleGrant processes the refresh token grant request and generates a new token response.
func (h *refreshTokenGrantHandler) HandleGrant(tokenRequest *model.TokenRequest,
	oauthApp *appmodel.OAuthAppConfigProcessedDTO, ctx *model.TokenContext) (
	*model.TokenResponseDTO, *model.ErrorResponse) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "RefreshTokenGrantHandler"))

	if errResp := h.verifyRefreshToken(tokenRequest.RefreshToken, logger); errResp != nil {
		return nil, errResp
	}

	refreshTokenClaims, errResp := h.getValidatedClaims(tokenRequest.RefreshToken, tokenRequest.ClientID, logger)
	if errResp != nil {
		return nil, errResp
	}

	// Grant type of the access token
	tokenGrantType := refreshTokenClaims["grant_type"].(string)

	// Extract scopes for the tokens
	refreshTokenScopes, newTokenScopes, errResp := h.extractScopes(tokenRequest.Scope, refreshTokenClaims, logger)
	if errResp != nil {
		return nil, errResp
	}

	// Extract access token claims from the refresh token claims if available
	sub := ""
	aud := ""
	userAttributes := map[string]interface{}{}
	if val, ok := refreshTokenClaims["access_token_sub"]; ok && val != "" {
		if subVal, ok := val.(string); ok {
			sub = subVal
		}
	}
	if val, ok := refreshTokenClaims["access_token_aud"]; ok && val != "" {
		if audVal, ok := val.(string); ok {
			aud = audVal
		}
	}
	if val, ok := refreshTokenClaims["access_token_user_attributes"]; ok && val != "" {
		if attrs, ok := val.(map[string]interface{}); ok {
			userAttributes = attrs
		}
	}

	conf := config.GetThunderRuntime().Config

	// Get token configuration from OAuth app
	iss := ""
	validityPeriod := int64(0)
	if oauthApp.Token != nil && oauthApp.Token.AccessToken != nil {
		iss = oauthApp.Token.AccessToken.Issuer
		validityPeriod = oauthApp.Token.AccessToken.ValidityPeriod
	}
	if iss == "" {
		iss = conf.OAuth.JWT.Issuer
	}
	if validityPeriod == 0 {
		validityPeriod = conf.OAuth.JWT.ValidityPeriod
	}

	// Issue new access token
	jwtClaims := make(map[string]interface{})
	if len(newTokenScopes) > 0 {
		jwtClaims["scope"] = strings.Join(newTokenScopes, " ")
	}
	if len(userAttributes) > 0 {
		for k, v := range userAttributes {
			jwtClaims[k] = v
		}
	}

	accessToken, iat, err := h.JWTService.GenerateJWT(sub, aud, iss, validityPeriod, jwtClaims)
	if err != nil {
		return nil, &model.ErrorResponse{
			Error:            constants.ErrorServerError,
			ErrorDescription: "Failed to generate access token",
		}
	}

	// Prepare the token response
	tokenResponse := &model.TokenResponseDTO{
		AccessToken: model.TokenDTO{
			Token:     accessToken,
			TokenType: constants.TokenTypeBearer,
			IssuedAt:  iat,
			ExpiresIn: validityPeriod,
			Scopes:    newTokenScopes,
			ClientID:  tokenRequest.ClientID,
		},
	}

	// Issue a new refresh token if renew_on_grant is enabled.
	if conf.OAuth.RefreshToken.RenewOnGrant {
		refreshTokenCtx := &model.TokenContext{
			TokenAttributes: make(map[string]interface{}),
		}
		if sub != "" {
			refreshTokenCtx.TokenAttributes["sub"] = sub
		}
		if aud != "" {
			refreshTokenCtx.TokenAttributes["aud"] = aud
		}

		logger.Debug("Renewing refresh token", log.String("client_id", tokenRequest.ClientID))
		errResp := h.IssueRefreshToken(tokenResponse, oauthApp, refreshTokenCtx, tokenGrantType, refreshTokenScopes)
		if errResp != nil && errResp.Error != "" {
			errResp.ErrorDescription = "Error while issuing refresh token: " + errResp.ErrorDescription
			logger.Error("Failed to issue refresh token", log.String("error", errResp.Error))
			return nil, errResp
		}
	} else {
		tokenResponse.RefreshToken = model.TokenDTO{
			Token:     tokenRequest.RefreshToken,
			TokenType: constants.TokenTypeBearer,
			Scopes:    refreshTokenScopes,
			ClientID:  tokenRequest.ClientID,
		}

		// Resolve and add the issued at time for the refresh token
		switch issuedAt := refreshTokenClaims["iat"].(type) {
		case float64:
			tokenResponse.RefreshToken.IssuedAt = int64(issuedAt)
		case int64:
			tokenResponse.RefreshToken.IssuedAt = issuedAt
		}
	}

	return tokenResponse, nil
}

// IssueRefreshToken generates a new refresh token for the given OAuth application and scopes.
func (h *refreshTokenGrantHandler) IssueRefreshToken(tokenResponse *model.TokenResponseDTO,
	oauthApp *appmodel.OAuthAppConfigProcessedDTO, ctx *model.TokenContext,
	grantType string, scopes []string) *model.ErrorResponse {
	// Extract sub and aud from the context attributes if available
	sub := ""
	aud := ""
	if len(ctx.TokenAttributes) > 0 {
		if val, ok := ctx.TokenAttributes["sub"]; ok && val != "" {
			sub = val.(string)
		}
		if val, ok := ctx.TokenAttributes["aud"]; ok && val != "" {
			aud = val.(string)
		}
	}

	// Get token configuration from OAuth app
	iss := ""
	validityPeriod := int64(0)
	if oauthApp.Token != nil && oauthApp.Token.AccessToken != nil {
		iss = oauthApp.Token.AccessToken.Issuer
		validityPeriod = oauthApp.Token.AccessToken.ValidityPeriod
	}
	if iss == "" {
		iss = config.GetThunderRuntime().Config.OAuth.JWT.Issuer
	}
	if validityPeriod == 0 {
		validityPeriod = config.GetThunderRuntime().Config.OAuth.JWT.ValidityPeriod
	}

	// Generate a JWT token for the refresh token.
	claims := map[string]interface{}{
		"client_id":  oauthApp.ClientID,
		"grant_type": grantType,
		"scopes":     strings.Join(scopes, " "),
	}
	if sub != "" {
		claims["access_token_sub"] = sub
	}
	if aud != "" {
		claims["access_token_aud"] = aud
	}
	if oauthApp.Token != nil && oauthApp.Token.AccessToken != nil &&
		len(oauthApp.Token.AccessToken.UserAttributes) > 0 && len(tokenResponse.AccessToken.UserAttributes) > 0 {
		claims["access_token_user_attributes"] = tokenResponse.AccessToken.UserAttributes
	}

	token, iat, err := h.JWTService.GenerateJWT(oauthApp.ClientID, oauthApp.ClientID, iss, validityPeriod, claims)
	if err != nil {
		return &model.ErrorResponse{
			Error:            constants.ErrorServerError,
			ErrorDescription: "Failed to generate refresh token",
		}
	}

	if tokenResponse == nil {
		tokenResponse = &model.TokenResponseDTO{}
	}
	tokenResponse.RefreshToken = model.TokenDTO{
		Token:     token,
		TokenType: constants.TokenTypeBearer,
		IssuedAt:  iat,
		ExpiresIn: validityPeriod,
		Scopes:    scopes,
		ClientID:  oauthApp.ClientID,
	}
	return nil
}

// verifyRefreshToken verifies the refresh token using the server's public key.
func (h *refreshTokenGrantHandler) verifyRefreshToken(refreshToken string, logger *log.Logger) *model.ErrorResponse {
	if err := h.JWTService.VerifyJWT(refreshToken, "", ""); err != nil {
		logger.Error("Failed to verify refresh token signature", log.Error(err))
		return &model.ErrorResponse{
			Error:            constants.ErrorInvalidRequest,
			ErrorDescription: "Invalid refresh token",
		}
	}

	return nil
}

// getValidatedClaims validates the claims in the refresh token and returns them if valid.
func (h *refreshTokenGrantHandler) getValidatedClaims(refreshToken, clientID string,
	logger *log.Logger) (map[string]interface{}, *model.ErrorResponse) {
	claims, err := jwt.DecodeJWTPayload(refreshToken)
	if err != nil {
		logger.Error("Failed to parse refresh token claims", log.Error(err))
		return nil, &model.ErrorResponse{
			Error:            constants.ErrorInvalidRequest,
			ErrorDescription: "Invalid refresh token",
		}
	}

	if errResp := h.validateIssuedAt(claims, logger); errResp != nil {
		return nil, errResp
	}

	if errResp := h.validateExpiryTime(claims, logger); errResp != nil {
		return nil, errResp
	}

	if errResp := h.validateNBF(claims, logger); errResp != nil {
		return nil, errResp
	}

	// Validate grant_type in token
	grantTypeVal, ok := claims["grant_type"]
	if !ok || grantTypeVal == nil {
		logger.Debug("Refresh token does not contain grant_type")
		return nil, &model.ErrorResponse{
			Error:            constants.ErrorInvalidRequest,
			ErrorDescription: "Invalid refresh token",
		}
	}
	if _, ok := grantTypeVal.(string); !ok {
		logger.Debug("Refresh token grant_type does not match expected", log.Any("grantType", grantTypeVal))
		return nil, &model.ErrorResponse{
			Error:            constants.ErrorInvalidRequest,
			ErrorDescription: "Invalid refresh token",
		}
	}

	// Validate client_id
	if claims["client_id"] != clientID {
		logger.Debug("Refresh token client_id does not match request")
		return nil, &model.ErrorResponse{
			Error:            constants.ErrorInvalidRequest,
			ErrorDescription: "Invalid refresh token",
		}
	}

	return claims, nil
}

// validateIssuedAt validates the issued at time (iat) claim of the refresh token.
func (h *refreshTokenGrantHandler) validateIssuedAt(claims map[string]interface{},
	logger *log.Logger) *model.ErrorResponse {
	return h.validateTimeClaim(
		claims,
		"iat",
		func(now, claim int64) bool { return now < claim },
		"Refresh token not valid yet",
		"Refresh token not valid yet",
		logger,
	)
}

// validateExpiryTime validates the expiry time (exp) claim of the refresh token.
func (h *refreshTokenGrantHandler) validateExpiryTime(claims map[string]interface{},
	logger *log.Logger) *model.ErrorResponse {
	return h.validateTimeClaim(
		claims,
		"exp",
		func(now, claim int64) bool { return now > claim },
		"Refresh token has expired",
		"Expired refresh token",
		logger,
	)
}

// validateTimeClaim validates a given time-based claim in the refresh token.
func (h *refreshTokenGrantHandler) validateTimeClaim(claims map[string]interface{}, claimKey string,
	cmp func(now, claim int64) bool, errMsg, errDesc string, logger *log.Logger) *model.ErrorResponse {
	val, ok := claims[claimKey]
	if !ok || val == nil {
		logger.Debug("Refresh token does not contain " + claimKey + " time")
		return &model.ErrorResponse{
			Error:            constants.ErrorInvalidRequest,
			ErrorDescription: "Invalid refresh token",
		}
	}

	now := time.Now().Unix()
	switch t := val.(type) {
	case float64:
		if cmp(now, int64(t)) {
			logger.Debug(errMsg, log.Any(claimKey+"Time", time.Unix(int64(t), 0)))
			return &model.ErrorResponse{
				Error:            constants.ErrorInvalidRequest,
				ErrorDescription: errDesc,
			}
		}
	case int64:
		if cmp(now, t) {
			logger.Debug(errMsg, log.Any(claimKey+"Time", time.Unix(t, 0)))
			return &model.ErrorResponse{
				Error:            constants.ErrorInvalidRequest,
				ErrorDescription: errDesc,
			}
		}
	default:
		logger.Debug("Refresh token "+claimKey+" time is not a valid type", log.Any(claimKey+"Time", t))
		return &model.ErrorResponse{
			Error:            constants.ErrorInvalidRequest,
			ErrorDescription: "Invalid refresh token",
		}
	}
	return nil
}

// validateNBF validates the not before time (nbf) claim of the refresh token if present.
func (h *refreshTokenGrantHandler) validateNBF(claims map[string]interface{},
	logger *log.Logger) *model.ErrorResponse {
	nbfVal, ok := claims["nbf"]
	if ok && nbfVal != nil {
		switch nbf := nbfVal.(type) {
		case float64:
			if time.Now().Unix() < int64(nbf) {
				logger.Debug("Refresh token not valid yet", log.Any("notBeforeTime", time.Unix(int64(nbf), 0)))
				return &model.ErrorResponse{
					Error:            constants.ErrorInvalidRequest,
					ErrorDescription: "Refresh token not valid yet",
				}
			}
		case int64:
			if time.Now().Unix() < nbf {
				logger.Debug("Refresh token not valid yet", log.Any("notBeforeTime", time.Unix(nbf, 0)))
				return &model.ErrorResponse{
					Error:            constants.ErrorInvalidRequest,
					ErrorDescription: "Refresh token not valid yet",
				}
			}
		default:
			logger.Debug("Refresh token not before time is not a valid type", log.Any("notBeforeTime", nbf))
			return &model.ErrorResponse{
				Error:            constants.ErrorInvalidRequest,
				ErrorDescription: "Invalid refresh token",
			}
		}
	}

	return nil
}

// extractScopes extracts and validates the scopes from the refresh token claims. It returns scopes for the
// refresh token and the new access token.
func (h *refreshTokenGrantHandler) extractScopes(requestedScopes string, refreshTokenClaims map[string]interface{},
	logger *log.Logger) ([]string, []string, *model.ErrorResponse) {
	// Extract scopes from the refresh token claims
	refreshTokenScopes := []string{}
	if s, ok := refreshTokenClaims["scopes"]; ok && s != "" {
		if scopeStr, ok := s.(string); ok {
			trimmedScopeStr := strings.TrimSpace(scopeStr)
			if trimmedScopeStr != "" {
				refreshTokenScopes = strings.Split(trimmedScopeStr, " ")
			}
		} else {
			logger.Debug("Scopes in refresh token are not a valid string", log.Any("scopes", s))
			return nil, nil, &model.ErrorResponse{
				Error:            constants.ErrorInvalidRequest,
				ErrorDescription: "Invalid refresh token",
			}
		}
	}

	// Validate and filter new token scopes
	newTokenScopes := []string{}
	if len(refreshTokenScopes) == 0 {
		logger.Debug("Scopes not found in the refresh token. Skipping granting any scopes")
	} else {
		trimmedScopes := strings.TrimSpace(requestedScopes)
		if trimmedScopes != "" {
			logger.Debug("Requested scopes found in the token request", log.Any("requestedScopes", trimmedScopes))
			for _, scope := range strings.Split(trimmedScopes, " ") {
				if scope == "" {
					continue
				}
				if slices.Contains(refreshTokenScopes, scope) {
					newTokenScopes = append(newTokenScopes, scope)
				} else {
					logger.Debug("Requested scope not found in refresh token. skipping", log.String("scope", scope))
				}
			}
		} else {
			logger.Debug("No scopes requested in the token request. Granting all scopes from refresh token",
				log.Any("scopes", refreshTokenScopes))
			newTokenScopes = refreshTokenScopes
		}
	}

	return refreshTokenScopes, newTokenScopes, nil
}
