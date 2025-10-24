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

package dcr

import (
	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
)

// Default values for DCR
const (
	ClientSecretExpiresAtNever = 0 // Never expires
)

// DCRRegistrationRequest represents the RFC 7591 Dynamic Client Registration request.
type DCRRegistrationRequest struct {
	RedirectURIs            []string                            `json:"redirect_uris"`
	GrantTypes              []oauth2const.GrantType             `json:"grant_types,omitempty"`
	ResponseTypes           []oauth2const.ResponseType          `json:"response_types,omitempty"`
	ClientName              string                              `json:"client_name,omitempty"`
	ClientURI               string                              `json:"client_uri,omitempty"`
	LogoURI                 string                              `json:"logo_uri,omitempty"`
	TokenEndpointAuthMethod oauth2const.TokenEndpointAuthMethod `json:"token_endpoint_auth_method,omitempty"`
	JWKSUri                 string                              `json:"jwks_uri,omitempty"`
	JWKS                    map[string]interface{}              `json:"jwks,omitempty"`
	Scope                   string                              `json:"scope,omitempty"`
	Contacts                []string                            `json:"contacts,omitempty"`
	TosURI                  string                              `json:"tos_uri,omitempty"`
	PolicyURI               string                              `json:"policy_uri,omitempty"`
}

// DCRRegistrationResponse represents the RFC 7591 Dynamic Client Registration response.
type DCRRegistrationResponse struct {
	ClientID                string                              `json:"client_id"`
	ClientSecret            string                              `json:"client_secret,omitempty"`
	ClientSecretExpiresAt   int64                               `json:"client_secret_expires_at"`
	RedirectURIs            []string                            `json:"redirect_uris,omitempty"`
	GrantTypes              []oauth2const.GrantType             `json:"grant_types,omitempty"`
	ResponseTypes           []oauth2const.ResponseType          `json:"response_types,omitempty"`
	ClientName              string                              `json:"client_name,omitempty"`
	ClientURI               string                              `json:"client_uri,omitempty"`
	LogoURI                 string                              `json:"logo_uri,omitempty"`
	TokenEndpointAuthMethod oauth2const.TokenEndpointAuthMethod `json:"token_endpoint_auth_method,omitempty"`
	JWKSUri                 string                              `json:"jwks_uri,omitempty"`
	JWKS                    map[string]interface{}              `json:"jwks,omitempty"`
	Scope                   string                              `json:"scope,omitempty"`
	Contacts                []string                            `json:"contacts,omitempty"`
	TosURI                  string                              `json:"tos_uri,omitempty"`
	PolicyURI               string                              `json:"policy_uri,omitempty"`
	AppID                   string                              `json:"app_id,omitempty"`
}

// DCRErrorResponse represents the RFC 7591 Dynamic Client Registration error response.
type DCRErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}
