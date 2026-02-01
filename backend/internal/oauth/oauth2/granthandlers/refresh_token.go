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

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/tokenservice"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/jwt"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/user"
)

// refreshTokenGrantHandler handles the refresh token grant type.
type refreshTokenGrantHandler struct {
	jwtService     jwt.JWTServiceInterface
	userService    user.UserServiceInterface
	tokenBuilder   tokenservice.TokenBuilderInterface
	tokenValidator tokenservice.TokenValidatorInterface
}

// newRefreshTokenGrantHandler creates a new instance of RefreshTokenGrantHandler.
func newRefreshTokenGrantHandler(
	jwtService jwt.JWTServiceInterface,
	userService user.UserServiceInterface,
	tokenBuilder tokenservice.TokenBuilderInterface,
	tokenValidator tokenservice.TokenValidatorInterface,
) RefreshTokenGrantHandlerInterface {
	return &refreshTokenGrantHandler{
		jwtService:     jwtService,
		userService:    userService,
		tokenBuilder:   tokenBuilder,
		tokenValidator: tokenValidator,
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
	oauthApp *appmodel.OAuthAppConfigProcessedDTO) (
	*model.TokenResponseDTO, *model.ErrorResponse) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "RefreshTokenGrantHandler"))

	// Validate refresh token using token validator
	refreshTokenClaims, err := h.tokenValidator.ValidateRefreshToken(tokenRequest.RefreshToken, tokenRequest.ClientID)
	if err != nil {
		logger.Error("Failed to validate refresh token", log.Error(err))
		return nil, &model.ErrorResponse{
			Error:            constants.ErrorInvalidGrant,
			ErrorDescription: "Invalid refresh token",
		}
	}

	newTokenScopes := h.applyScopeDownscoping(tokenRequest.Scope, refreshTokenClaims.Scopes, logger)

	accessToken, err := h.tokenBuilder.BuildAccessToken(&tokenservice.AccessTokenBuildContext{
		Subject:        refreshTokenClaims.Sub,
		Audience:       refreshTokenClaims.Aud,
		ClientID:       tokenRequest.ClientID,
		Scopes:         newTokenScopes,
		UserAttributes: refreshTokenClaims.UserAttributes,
		GrantType:      refreshTokenClaims.GrantType,
		OAuthApp:       oauthApp,
	})
	if err != nil {
		logger.Error("Failed to generate access token", log.Error(err))
		return nil, &model.ErrorResponse{
			Error:            constants.ErrorServerError,
			ErrorDescription: "Failed to generate access token",
		}
	}

	// Prepare the token response
	tokenResponse := &model.TokenResponseDTO{
		AccessToken: *accessToken,
	}

	// Check configuration for refresh token renewal
	conf := config.GetThunderRuntime().Config
	renewRefreshToken := conf.OAuth.RefreshToken.RenewOnGrant

	// Issue a new refresh token if renew_on_grant is enabled
	if renewRefreshToken {
		logger.Debug("Renewing refresh token", log.String("client_id", tokenRequest.ClientID))
		errResp := h.IssueRefreshToken(tokenResponse, oauthApp, refreshTokenClaims.Sub, refreshTokenClaims.Aud,
			refreshTokenClaims.GrantType, newTokenScopes)
		if errResp != nil && errResp.Error != "" {
			errResp.ErrorDescription = "Error while issuing refresh token: " + errResp.ErrorDescription
			logger.Error("Failed to issue refresh token", log.String("error", errResp.Error))
			return nil, errResp
		}
	} else {
		// Return the existing refresh token
		tokenResponse.RefreshToken = model.TokenDTO{
			Token:    tokenRequest.RefreshToken,
			IssuedAt: refreshTokenClaims.Iat,
			Scopes:   refreshTokenClaims.Scopes,
			ClientID: tokenRequest.ClientID,
		}
	}

	return tokenResponse, nil
}

// IssueRefreshToken generates a new refresh token for the given OAuth application and scopes.
func (h *refreshTokenGrantHandler) IssueRefreshToken(
	tokenResponse *model.TokenResponseDTO,
	oauthApp *appmodel.OAuthAppConfigProcessedDTO,
	subject, audience, grantType string,
	scopes []string,
) *model.ErrorResponse {
	tokenCtx := &tokenservice.RefreshTokenBuildContext{
		ClientID:             oauthApp.ClientID,
		Scopes:               scopes,
		GrantType:            grantType,
		AccessTokenSubject:   subject,
		AccessTokenAudience:  audience,
		AccessTokenUserAttrs: tokenResponse.AccessToken.UserAttributes,
		OAuthApp:             oauthApp,
	}

	// Build refresh token using token builder
	refreshToken, err := h.tokenBuilder.BuildRefreshToken(tokenCtx)
	if err != nil {
		return &model.ErrorResponse{
			Error:            constants.ErrorServerError,
			ErrorDescription: "Failed to generate refresh token",
		}
	}

	if tokenResponse == nil {
		tokenResponse = &model.TokenResponseDTO{}
	}
	tokenResponse.RefreshToken = *refreshToken
	return nil
}

// applyScopeDownscoping applies OAuth2 scope downscoping logic.
// If no scopes are requested, all refresh token scopes are granted.
// If scopes are requested, only the intersection with refresh token scopes is granted.
func (h *refreshTokenGrantHandler) applyScopeDownscoping(requestedScopes string,
	refreshTokenScopes []string, logger *log.Logger) []string {
	trimmedRequestedScopes := tokenservice.ParseScopes(requestedScopes)

	if len(trimmedRequestedScopes) == 0 {
		logger.Debug("No scopes requested. Granting all scopes from refresh token",
			log.Any("scopes", refreshTokenScopes))
		return refreshTokenScopes
	}

	newTokenScopes := []string{}
	for _, requestedScope := range trimmedRequestedScopes {
		if slices.Contains(refreshTokenScopes, requestedScope) {
			newTokenScopes = append(newTokenScopes, requestedScope)
		} else {
			logger.Debug("Requested scope not found in refresh token, skipping",
				log.String("scope", requestedScope))
		}
	}

	logger.Debug("Applied scope downscoping", log.Any("grantedScopes", newTokenScopes))
	return newTokenScopes
}
