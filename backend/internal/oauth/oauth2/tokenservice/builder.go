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

package tokenservice

import (
	"fmt"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/model"
	"github.com/asgardeo/thunder/internal/system/jwt"
)

// TokenBuilderInterface defines the interface for building OAuth2 tokens.
type TokenBuilderInterface interface {
	BuildAccessToken(ctx *AccessTokenBuildContext) (*model.TokenDTO, error)
	BuildRefreshToken(ctx *RefreshTokenBuildContext) (*model.TokenDTO, error)
	BuildIDToken(ctx *IDTokenBuildContext) (*model.TokenDTO, error)
}

// TokenBuilder implements TokenBuilderInterface.
type tokenBuilder struct {
	jwtService jwt.JWTServiceInterface
}

// NewTokenBuilder creates a new TokenBuilder instance.
func newTokenBuilder(jwtService jwt.JWTServiceInterface) TokenBuilderInterface {
	return &tokenBuilder{
		jwtService: jwtService,
	}
}

// BuildAccessToken builds an access token with all necessary claims.
func (tb *tokenBuilder) BuildAccessToken(ctx *AccessTokenBuildContext) (*model.TokenDTO, error) {
	if ctx == nil {
		return nil, fmt.Errorf("build context cannot be nil")
	}

	tokenConfig := resolveTokenConfig(ctx.OAuthApp, TokenTypeAccess)

	userAttributes := tb.buildAccessTokenUserAttributes(ctx.UserAttributes, ctx.OAuthApp)
	jwtClaims := tb.buildAccessTokenClaims(ctx, userAttributes)

	tokenDTO := &model.TokenDTO{
		TokenType:      constants.TokenTypeBearer,
		ExpiresIn:      tokenConfig.ValidityPeriod,
		Scopes:         ctx.Scopes,
		ClientID:       ctx.ClientID,
		UserAttributes: userAttributes,
		Subject:        ctx.Subject,
		Audience:       ctx.Audience,
	}

	token, iat, err := tb.jwtService.GenerateJWT(
		ctx.Subject,
		ctx.Audience,
		tokenConfig.Issuer,
		tokenConfig.ValidityPeriod,
		jwtClaims,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %v", err)
	}

	// Assign generated token and issued at time
	tokenDTO.Token = token
	tokenDTO.IssuedAt = iat

	return tokenDTO, nil
}

// buildAccessTokenClaims builds the claims map for an access token.
func (tb *tokenBuilder) buildAccessTokenClaims(
	ctx *AccessTokenBuildContext,
	filteredAttributes map[string]interface{},
) map[string]interface{} {
	claims := make(map[string]interface{})

	if len(ctx.Scopes) > 0 {
		claims["scope"] = JoinScopes(ctx.Scopes)
	}

	if ctx.ClientID != "" {
		claims["client_id"] = ctx.ClientID
	}

	if ctx.GrantType != "" {
		claims["grant_type"] = ctx.GrantType
	}

	// Add filtered user attributes to claims
	for key, value := range filteredAttributes {
		claims[key] = value
	}

	if ctx.ActorClaims != nil {
		actClaim := tb.buildActorClaim(ctx.ActorClaims)
		claims["act"] = actClaim
	}

	return claims
}

// buildAccessTokenUserAttributes builds user attributes for the access token based on app configuration.
func (tb *tokenBuilder) buildAccessTokenUserAttributes(
	attrs map[string]interface{},
	oauthApp *appmodel.OAuthAppConfigProcessedDTO,
) map[string]interface{} {
	accessTokenAttributes := make(map[string]interface{})

	if attrs == nil {
		attrs = make(map[string]interface{})
	}

	// Get access token user attributes from config if available
	var accessTokenUserAttributes []string
	if oauthApp != nil && oauthApp.Token != nil && oauthApp.Token.AccessToken != nil {
		accessTokenUserAttributes = oauthApp.Token.AccessToken.UserAttributes
	}

	// If app config specifies which attributes to include, filter them
	if len(accessTokenUserAttributes) > 0 {
		for _, attr := range accessTokenUserAttributes {
			if val, ok := attrs[attr]; ok {
				accessTokenAttributes[attr] = val
			}
		}
	}
	// If no filtering configured, return empty attributes

	return accessTokenAttributes
}

// buildActorClaim builds the actor claim for token exchange.
func (tb *tokenBuilder) buildActorClaim(actorClaims *SubjectTokenClaims) map[string]interface{} {
	actClaim := map[string]interface{}{
		"sub": actorClaims.Sub,
	}

	if actorClaims.Iss != "" {
		actClaim["iss"] = actorClaims.Iss
	}

	if len(actorClaims.NestedAct) > 0 {
		actClaim["act"] = actorClaims.NestedAct
	}

	return actClaim
}

// BuildRefreshToken builds a refresh token with all necessary claims.
func (tb *tokenBuilder) BuildRefreshToken(ctx *RefreshTokenBuildContext) (*model.TokenDTO, error) {
	if ctx == nil {
		return nil, fmt.Errorf("build context cannot be nil")
	}

	tokenConfig := resolveTokenConfig(ctx.OAuthApp, TokenTypeRefresh)

	claims := tb.buildRefreshTokenClaims(ctx)

	tokenDTO := &model.TokenDTO{
		ExpiresIn: tokenConfig.ValidityPeriod,
		Scopes:    ctx.Scopes,
		ClientID:  ctx.ClientID,
		Subject:   ctx.AccessTokenSubject,
		Audience:  tokenConfig.Issuer,
	}

	token, iat, err := tb.jwtService.GenerateJWT(
		ctx.ClientID,
		tokenConfig.Issuer,
		tokenConfig.Issuer,
		tokenConfig.ValidityPeriod,
		claims,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %v", err)
	}

	// Assign generated token and issued at time
	tokenDTO.Token = token
	tokenDTO.IssuedAt = iat

	return tokenDTO, nil
}

// buildRefreshTokenClaims builds the claims map for a refresh token.
func (tb *tokenBuilder) buildRefreshTokenClaims(ctx *RefreshTokenBuildContext) map[string]interface{} {
	claims := make(map[string]interface{})

	if len(ctx.Scopes) > 0 {
		claims["scope"] = JoinScopes(ctx.Scopes)
	}

	claims["access_token_sub"] = ctx.AccessTokenSubject
	claims["access_token_aud"] = ctx.AccessTokenAudience
	claims["grant_type"] = ctx.GrantType

	if ctx.OAuthApp != nil &&
		ctx.OAuthApp.Token != nil &&
		ctx.OAuthApp.Token.AccessToken != nil &&
		len(ctx.OAuthApp.Token.AccessToken.UserAttributes) > 0 &&
		len(ctx.AccessTokenUserAttrs) > 0 {
		claims["access_token_user_attributes"] = ctx.AccessTokenUserAttrs
	}

	return claims
}

// BuildIDToken builds an OIDC ID token with all necessary claims.
func (tb *tokenBuilder) BuildIDToken(ctx *IDTokenBuildContext) (*model.TokenDTO, error) {
	if ctx == nil {
		return nil, fmt.Errorf("build context cannot be nil")
	}

	tokenConfig := resolveTokenConfig(ctx.OAuthApp, TokenTypeID)

	jwtClaims := tb.buildIDTokenClaims(ctx)

	tokenDTO := &model.TokenDTO{
		ExpiresIn: tokenConfig.ValidityPeriod,
		Scopes:    ctx.Scopes,
		ClientID:  ctx.Audience,
		Subject:   ctx.Subject,
		Audience:  ctx.Audience,
	}

	token, iat, err := tb.jwtService.GenerateJWT(
		ctx.Subject,
		ctx.Audience,
		tokenConfig.Issuer,
		tokenConfig.ValidityPeriod,
		jwtClaims,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ID token: %v", err)
	}

	// Assign generated token and issued at time
	tokenDTO.Token = token
	tokenDTO.IssuedAt = iat

	return tokenDTO, nil
}

// buildIDTokenClaims builds the claims map for an ID token (OIDC).
func (tb *tokenBuilder) buildIDTokenClaims(ctx *IDTokenBuildContext) map[string]interface{} {
	claims := make(map[string]interface{})

	if ctx.AuthTime > 0 {
		claims["auth_time"] = ctx.AuthTime
	}

	userAttributes := ctx.UserAttributes
	if userAttributes == nil {
		userAttributes = make(map[string]interface{})
	}

	scopeClaims := BuildOIDCClaimsFromScopes(
		ctx.Scopes,
		userAttributes,
		ctx.OAuthApp,
	)

	for key, value := range scopeClaims {
		claims[key] = value
	}

	return claims
}
