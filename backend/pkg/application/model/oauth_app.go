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

// Package model defines the data structures for the application module.
//
//nolint:lll
package model

import (
	inboundmodel "github.com/asgardeo/thunder/pkg/inboundclient/model"
	oauth2const "github.com/asgardeo/thunder/pkg/oauth/oauth2/constants"
)

// OAuthAppConfig represents the structure for OAuth application configuration.
type OAuthAppConfig struct {
	ClientID                           string                              `json:"clientId"`
	RedirectURIs                       []string                            `json:"redirectUris"`
	GrantTypes                         []oauth2const.GrantType             `json:"grantTypes"`
	ResponseTypes                      []oauth2const.ResponseType          `json:"responseTypes"`
	TokenEndpointAuthMethod            oauth2const.TokenEndpointAuthMethod `json:"tokenEndpointAuthMethod"`
	PKCERequired                       bool                                `json:"pkceRequired"`
	PublicClient                       bool                                `json:"publicClient"`
	RequirePushedAuthorizationRequests bool                                `json:"requirePushedAuthorizationRequests"`
	Token                              *inboundmodel.OAuthTokenConfig      `json:"token,omitempty"`
	Scopes                             []string                            `json:"scopes,omitempty"`
	UserInfo                           *inboundmodel.UserInfoConfig        `json:"userInfo,omitempty"`
	ScopeClaims                        map[string][]string                 `json:"scopeClaims,omitempty"`
	Certificate                        *inboundmodel.Certificate           `json:"certificate,omitempty"`
}

// OAuthAppConfigComplete represents the complete structure for OAuth application configuration.
type OAuthAppConfigComplete struct {
	ClientID                           string                              `json:"clientId" yaml:"client_id"`
	ClientSecret                       string                              `json:"clientSecret,omitempty" yaml:"client_secret,omitempty"`
	RedirectURIs                       []string                            `json:"redirectUris" yaml:"redirect_uris"`
	GrantTypes                         []oauth2const.GrantType             `json:"grantTypes" yaml:"grant_types"`
	ResponseTypes                      []oauth2const.ResponseType          `json:"responseTypes" yaml:"response_types"`
	TokenEndpointAuthMethod            oauth2const.TokenEndpointAuthMethod `json:"tokenEndpointAuthMethod" yaml:"token_endpoint_auth_method"`
	PKCERequired                       bool                                `json:"pkceRequired" yaml:"pkce_required"`
	PublicClient                       bool                                `json:"publicClient" yaml:"public_client"`
	RequirePushedAuthorizationRequests bool                                `json:"requirePushedAuthorizationRequests" yaml:"require_pushed_authorization_requests"`
	Token                              *inboundmodel.OAuthTokenConfig      `json:"token,omitempty" yaml:"token,omitempty"`
	Scopes                             []string                            `json:"scopes,omitempty" yaml:"scopes,omitempty"`
	UserInfo                           *inboundmodel.UserInfoConfig        `json:"userInfo,omitempty" yaml:"user_info,omitempty"`
	ScopeClaims                        map[string][]string                 `json:"scopeClaims,omitempty" yaml:"scope_claims,omitempty"`
	Certificate                        *inboundmodel.Certificate           `json:"certificate,omitempty" jsonschema:"Application certificate. Optional. For certificate-based authentication or JWT validation."`
}

// OAuthAppConfigDTO represents the data transfer object for OAuth application configuration.
type OAuthAppConfigDTO struct {
	AppID                              string                              `json:"appId,omitempty" jsonschema:"The unique identifier of the OAuth application"`
	ClientID                           string                              `json:"clientId,omitempty" jsonschema:"OAuth client ID (auto-generated if not provided)"`
	ClientSecret                       string                              `json:"clientSecret,omitempty" jsonschema:"OAuth client secret (auto-generated if not provided)"`
	RedirectURIs                       []string                            `json:"redirectUris,omitempty" jsonschema:"Allowed redirect URIs. Required for Public (SPA/Mobile) and Confidential (Server) clients. Omit for M2M."`
	GrantTypes                         []oauth2const.GrantType             `json:"grantTypes,omitempty" jsonschema:"OAuth grant types. Common: [authorization_code, refresh_token] for user apps, [client_credentials] for M2M."`
	ResponseTypes                      []oauth2const.ResponseType          `json:"responseTypes,omitempty" jsonschema:"OAuth response types. Common: [code] for user apps. Omit for M2M."`
	TokenEndpointAuthMethod            oauth2const.TokenEndpointAuthMethod `json:"tokenEndpointAuthMethod,omitempty" jsonschema:"Client authentication method. Use 'none' for Public clients, 'client_secret_basic' for Confidential/M2M."`
	PKCERequired                       bool                                `json:"pkceRequired,omitempty" jsonschema:"Require PKCE for security. Recommended for all user-interactive flows."`
	PublicClient                       bool                                `json:"publicClient,omitempty" jsonschema:"Identify if client is public (cannot store secrets). Set true for SPA/Mobile."`
	RequirePushedAuthorizationRequests bool                                `json:"requirePushedAuthorizationRequests,omitempty" jsonschema:"Require Pushed Authorization Requests (PAR) per RFC 9126."`
	Token                              *inboundmodel.OAuthTokenConfig      `json:"token,omitempty" jsonschema:"Token configuration for access tokens and ID tokens"`
	Scopes                             []string                            `json:"scopes,omitempty" jsonschema:"Allowed OAuth scopes. Add custom scopes as needed for your application."`
	UserInfo                           *inboundmodel.UserInfoConfig        `json:"userInfo,omitempty" jsonschema:"UserInfo endpoint configuration. Configure user attributes returned from the OIDC userinfo endpoint."`
	ScopeClaims                        map[string][]string                 `json:"scopeClaims,omitempty" jsonschema:"Scope-to-claims mapping. Maps OAuth scopes to user claims for both ID token and userinfo."`
	Certificate                        *inboundmodel.Certificate           `json:"certificate,omitempty" jsonschema:"Application certificate. Optional. For certificate-based authentication or JWT validation."`
}

// IsAllowedGrantType reports whether the input DTO allows the given grant type.
func (o *OAuthAppConfigDTO) IsAllowedGrantType(grantType oauth2const.GrantType) bool {
	return inboundmodel.IsAllowedGrantType(o.GrantTypes, grantType)
}

// IsAllowedResponseType reports whether the input DTO allows the given response type.
func (o *OAuthAppConfigDTO) IsAllowedResponseType(responseType string) bool {
	return inboundmodel.IsAllowedResponseType(o.ResponseTypes, responseType)
}

// IsAllowedTokenEndpointAuthMethod reports whether the input DTO uses the given auth method.
func (o *OAuthAppConfigDTO) IsAllowedTokenEndpointAuthMethod(method oauth2const.TokenEndpointAuthMethod) bool {
	return o.TokenEndpointAuthMethod == method
}

// ValidateRedirectURI validates the input DTO redirect URI against the registered list.
func (o *OAuthAppConfigDTO) ValidateRedirectURI(redirectURI string) error {
	return inboundmodel.ValidateRedirectURI(o.RedirectURIs, redirectURI)
}
