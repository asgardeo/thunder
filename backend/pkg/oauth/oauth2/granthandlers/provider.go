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
	"github.com/asgardeo/thunder/internal/attributecache"
	rbacauthz "github.com/asgardeo/thunder/internal/authz"
	"github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/resource"
	"github.com/asgardeo/thunder/internal/system/jose/jwt"
	"github.com/asgardeo/thunder/pkg/entityprovider"
	"github.com/asgardeo/thunder/pkg/oauth/oauth2/authz"
	"github.com/asgardeo/thunder/pkg/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/pkg/oauth/oauth2/tokenservice"
)

// GrantHandlerProviderInterface defines the interface for the grant handler provider.
type GrantHandlerProviderInterface interface {
	GetGrantHandler(grantType constants.GrantType) (GrantHandlerInterface, error)
}

// GrantHandlerProvider implements the GrantHandlerProviderInterface.
type GrantHandlerProvider struct {
	clientCredentialsGrantHandler GrantHandlerInterface
	authorizationCodeGrantHandler GrantHandlerInterface
	refreshTokenGrantHandler      GrantHandlerInterface
	tokenExchangeGrantHandler     GrantHandlerInterface
}

// newGrantHandlerProvider creates a new instance of GrantHandlerProvider.
func newGrantHandlerProvider(
	jwtService jwt.JWTServiceInterface,
	authzService authz.AuthorizeServiceInterface,
	tokenBuilder tokenservice.TokenBuilderInterface,
	tokenValidator tokenservice.TokenValidatorInterface,
	attrCacheService attributecache.AttributeCacheServiceInterface,
	ouService ou.OrganizationUnitServiceInterface,
	rbacAuthzService rbacauthz.AuthorizationServiceInterface,
	entityProv entityprovider.EntityProviderInterface,
	resourceService resource.ResourceServiceInterface,
) GrantHandlerProviderInterface {
	return &GrantHandlerProvider{
		clientCredentialsGrantHandler: newClientCredentialsGrantHandler(
			tokenBuilder, ouService, rbacAuthzService, entityProv, resourceService),
		authorizationCodeGrantHandler: newAuthorizationCodeGrantHandler(
			authzService, tokenBuilder, attrCacheService, resourceService),
		refreshTokenGrantHandler: newRefreshTokenGrantHandler(
			jwtService, tokenBuilder, tokenValidator, attrCacheService, resourceService),
		tokenExchangeGrantHandler: newTokenExchangeGrantHandler(
			tokenBuilder, tokenValidator, resourceService),
	}
}

// GetGrantHandler returns the appropriate grant handler for the given grant type.
func (p *GrantHandlerProvider) GetGrantHandler(grantType constants.GrantType) (GrantHandlerInterface, error) {
	switch grantType {
	case constants.GrantTypeClientCredentials:
		return p.clientCredentialsGrantHandler, nil
	case constants.GrantTypeAuthorizationCode:
		return p.authorizationCodeGrantHandler, nil
	case constants.GrantTypeRefreshToken:
		return p.refreshTokenGrantHandler, nil
	case constants.GrantTypeTokenExchange:
		return p.tokenExchangeGrantHandler, nil
	default:
		return nil, constants.UnSupportedGrantTypeError
	}
}
