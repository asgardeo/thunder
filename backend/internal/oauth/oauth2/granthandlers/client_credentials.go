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
	"context"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/tokenservice"
)

// clientCredentialsGrantHandler handles the client credentials grant type.
type clientCredentialsGrantHandler struct {
	tokenBuilder tokenservice.TokenBuilderInterface
}

// newClientCredentialsGrantHandler creates a new instance of ClientCredentialsGrantHandler.
func newClientCredentialsGrantHandler(
	tokenBuilder tokenservice.TokenBuilderInterface,
) GrantHandlerInterface {
	return &clientCredentialsGrantHandler{
		tokenBuilder: tokenBuilder,
	}
}

// ValidateGrant validates the client credentials grant type.
func (h *clientCredentialsGrantHandler) ValidateGrant(ctx context.Context, tokenRequest *model.TokenRequest,
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
func (h *clientCredentialsGrantHandler) HandleGrant(ctx context.Context, tokenRequest *model.TokenRequest,
	oauthApp *appmodel.OAuthAppConfigProcessedDTO) (
	*model.TokenResponseDTO, *model.ErrorResponse) {
	scopes := tokenservice.ParseScopes(tokenRequest.Scope)

	finalAudience := tokenservice.DetermineAudience("", tokenRequest.Resource, "", tokenRequest.ClientID)

	accessToken, err := h.tokenBuilder.BuildAccessToken(ctx, &tokenservice.AccessTokenBuildContext{
		Subject:        tokenRequest.ClientID,
		Audience:       finalAudience,
		ClientID:       tokenRequest.ClientID,
		Scopes:         scopes,
		UserAttributes: make(map[string]interface{}),
		GrantType:      string(constants.GrantTypeClientCredentials),
		OAuthApp:       oauthApp,
	})
	if err != nil {
		return nil, &model.ErrorResponse{
			Error:            constants.ErrorServerError,
			ErrorDescription: "Failed to generate token",
		}
	}

	return &model.TokenResponseDTO{
		AccessToken: *accessToken,
	}, nil
}
