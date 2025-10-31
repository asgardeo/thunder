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
	"strings"
	"time"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/model"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/jwt"
)

// clientCredentialsGrantHandler handles the client credentials grant type.
type clientCredentialsGrantHandler struct {
	jwtService jwt.JWTServiceInterface
}

// newClientCredentialsGrantHandler creates a new instance of ClientCredentialsGrantHandler.
func newClientCredentialsGrantHandler(jwtService jwt.JWTServiceInterface) GrantHandlerInterface {
	return &clientCredentialsGrantHandler{
		jwtService: jwtService,
	}
}

// ValidateGrant validates the client credentials grant type.
func (h *clientCredentialsGrantHandler) ValidateGrant(tokenRequest *model.TokenRequest,
	oauthApp *appmodel.OAuthAppConfigProcessedDTO) *model.ErrorResponse {
	if constants.GrantType(tokenRequest.GrantType) != constants.GrantTypeClientCredentials {
		return &model.ErrorResponse{
			Error:            constants.ErrorUnsupportedGrantType,
			ErrorDescription: "Unsupported grant type",
		}
	}

	// Validate the client ID and secret.
	if tokenRequest.ClientID == "" || tokenRequest.ClientSecret == "" {
		return &model.ErrorResponse{
			Error:            constants.ErrorInvalidRequest,
			ErrorDescription: "Client Id and secret are required",
		}
	}

	return nil
}

// HandleGrant handles the client credentials grant type.
func (h *clientCredentialsGrantHandler) HandleGrant(tokenRequest *model.TokenRequest,
	oauthApp *appmodel.OAuthAppConfigProcessedDTO, ctx *model.TokenContext) (
	*model.TokenResponseDTO, *model.ErrorResponse) {
	scopeString := strings.TrimSpace(tokenRequest.Scope)
	scopes := []string{}
	if scopeString != "" {
		scopes = strings.Split(scopeString, " ")
	}

	// Generate a JWT token for the client.
	jwtClaims := make(map[string]interface{})
	if scopeString != "" {
		jwtClaims["scope"] = scopeString
	}

	// Add audience claim based on resource parameter (RFC 8707)
	if tokenRequest.Resource != "" {
		jwtClaims["aud"] = tokenRequest.Resource
	}

	// Get token configuration from OAuth app
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

	token, _, err := h.jwtService.GenerateJWT(tokenRequest.ClientID, tokenRequest.ClientID, iss,
		validityPeriod, jwtClaims)
	if err != nil {
		return nil, &model.ErrorResponse{
			Error:            constants.ErrorServerError,
			ErrorDescription: "Failed to generate token",
		}
	}

	// Add context attributes.
	if ctx.TokenAttributes == nil {
		ctx.TokenAttributes = make(map[string]interface{})
	}
	ctx.TokenAttributes[constants.ClaimSub] = tokenRequest.ClientID
	// Set audience to resource if provided, otherwise use client ID
	if tokenRequest.Resource != "" {
		ctx.TokenAttributes[constants.ClaimAud] = tokenRequest.Resource
	} else {
		ctx.TokenAttributes[constants.ClaimAud] = tokenRequest.ClientID
	}

	// Prepare the token response.
	accessToken := &model.TokenDTO{
		Token:     token,
		TokenType: constants.TokenTypeBearer,
		IssuedAt:  time.Now().Unix(),
		ExpiresIn: 3600,
		Scopes:    scopes,
		ClientID:  tokenRequest.ClientID,
	}

	return &model.TokenResponseDTO{
		AccessToken: *accessToken,
	}, nil
}
