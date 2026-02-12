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

// Package granthandlers provides an interface and implementations for handling OAuth 2.0 grant types.
package granthandlers

import (
	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/model"
)

// GrantHandlerInterface defines the interface for handling OAuth 2.0 grants.
type GrantHandlerInterface interface {
	ValidateGrant(tokenRequest *model.TokenRequest, oauthApp *appmodel.OAuthAppConfigProcessedDTO) *model.ErrorResponse
	HandleGrant(tokenRequest *model.TokenRequest, oauthApp *appmodel.OAuthAppConfigProcessedDTO) (
		*model.TokenResponseDTO, *model.ErrorResponse)
}

// RefreshTokenGrantHandlerInterface defines the interface for handling refresh token grants.
type RefreshTokenGrantHandlerInterface interface {
	GrantHandlerInterface
	IssueRefreshToken(
		tokenResponse *model.TokenResponseDTO,
		oauthApp *appmodel.OAuthAppConfigProcessedDTO,
		subject, audience, grantType string,
		scopes []string,
	) *model.ErrorResponse
}
