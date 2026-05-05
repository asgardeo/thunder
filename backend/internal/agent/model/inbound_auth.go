/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

package model

import (
	inboundmodel "github.com/asgardeo/thunder/pkg/inboundclient/model"
	oauth2const "github.com/asgardeo/thunder/pkg/oauth/oauth2/constants"
)

// InboundAuthType represents the type of inbound authentication for an agent.
type InboundAuthType string

const (
	// OAuthInboundAuthType is the OAuth 2.0 inbound authentication type.
	OAuthInboundAuthType InboundAuthType = "oauth2"
)

// InboundAuthConfig wraps the OAuth client configuration for an agent. The clientSecret
// field is populated only on create / update responses.
type InboundAuthConfig struct {
	Type   InboundAuthType   `json:"type"`
	Config *OAuthAgentConfig `json:"config,omitempty"`
}

// OAuthAgentConfig is the agent-facing OAuth client configuration. The structure mirrors
// the OAuth profile data persisted by the inbound client subsystem.
//
// TODO: Refactor and get common fields from inboundclient without duplicating
//
//nolint:lll
type OAuthAgentConfig struct {
	ClientID                           string                              `json:"clientId,omitempty"`
	ClientSecret                       string                              `json:"clientSecret,omitempty"`
	RedirectURIs                       []string                            `json:"redirectUris,omitempty"`
	GrantTypes                         []oauth2const.GrantType             `json:"grantTypes,omitempty"`
	ResponseTypes                      []oauth2const.ResponseType          `json:"responseTypes,omitempty"`
	TokenEndpointAuthMethod            oauth2const.TokenEndpointAuthMethod `json:"tokenEndpointAuthMethod,omitempty"`
	PKCERequired                       bool                                `json:"pkceRequired,omitempty"`
	PublicClient                       bool                                `json:"publicClient,omitempty"`
	RequirePushedAuthorizationRequests bool                                `json:"requirePushedAuthorizationRequests,omitempty"`
	Certificate                        *inboundmodel.Certificate           `json:"certificate,omitempty"`
	Token                              *inboundmodel.OAuthTokenConfig      `json:"token,omitempty"`
	Scopes                             []string                            `json:"scopes,omitempty"`
	UserInfo                           *inboundmodel.UserInfoConfig        `json:"userInfo,omitempty"`
	ScopeClaims                        map[string][]string                 `json:"scopeClaims,omitempty"`
}
