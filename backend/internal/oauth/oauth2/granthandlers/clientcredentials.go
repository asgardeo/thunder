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

package granthandlers

import (
	"strings"
	"time"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/oauth/jwt"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/model"
)

// ClientCredentialsGrantHandler handles the client credentials grant type.
type ClientCredentialsGrantHandler struct{}

var _ GrantHandler = (*ClientCredentialsGrantHandler)(nil)

// ValidateGrant validates the client credentials grant type.
func (h *ClientCredentialsGrantHandler) ValidateGrant(tokenRequest *model.TokenRequest,
	oauthApp *appmodel.OAuthAppConfigProcessed) *model.ErrorResponse {
	// Validate the grant type.
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
func (h *ClientCredentialsGrantHandler) HandleGrant(tokenRequest *model.TokenRequest,
	oauthApp *appmodel.OAuthAppConfigProcessed, ctx *model.TokenContext) (*model.TokenResponseDTO, *model.ErrorResponse) {
	scopeString := strings.TrimSpace(tokenRequest.Scope)
	scopes := []string{}
	if scopeString != "" {
		scopes = strings.Split(scopeString, " ")
	}

	// Generate a JWT token for the client.
	jwtClaims := make(map[string]string)
	if scopeString != "" {
		jwtClaims["scope"] = scopeString
	}
	token, _, err := jwt.GenerateJWT(tokenRequest.ClientID, tokenRequest.ClientID,
		jwt.GetJWTTokenValidityPeriod(), jwtClaims)
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
	ctx.TokenAttributes["sub"] = tokenRequest.ClientID
	ctx.TokenAttributes["aud"] = tokenRequest.ClientID

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
