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

package model

import (
	"fmt"
	"net/url"
	"slices"

	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
)

// OAuthAppConfig represents the structure for OAuth application configuration.
type OAuthAppConfig struct {
	ClientID                string                                `json:"client_id"`
	RedirectURIs            []string                              `json:"redirect_uris"`
	GrantTypes              []oauth2const.GrantType               `json:"grant_types"`
	ResponseTypes           []oauth2const.ResponseType            `json:"response_types"`
	TokenEndpointAuthMethod []oauth2const.TokenEndpointAuthMethod `json:"token_endpoint_auth_methods"`
	PKCERequired            bool                                  `json:"pkce_required"`
	Token                   *OAuthTokenConfig                     `json:"token,omitempty"`
}

// OAuthAppConfigComplete represents the complete structure for OAuth application configuration.
type OAuthAppConfigComplete struct {
	ClientID                string                                `json:"client_id"`
	ClientSecret            string                                `json:"client_secret"`
	RedirectURIs            []string                              `json:"redirect_uris"`
	GrantTypes              []oauth2const.GrantType               `json:"grant_types"`
	ResponseTypes           []oauth2const.ResponseType            `json:"response_types"`
	TokenEndpointAuthMethod []oauth2const.TokenEndpointAuthMethod `json:"token_endpoint_auth_methods"`
	PKCERequired            bool                                  `json:"pkce_required"`
	Token                   *OAuthTokenConfig                     `json:"token,omitempty"`
}

// OAuthAppConfigDTO represents the data transfer object for OAuth application configuration.
type OAuthAppConfigDTO struct {
	AppID                   string
	ClientID                string
	ClientSecret            string
	RedirectURIs            []string
	GrantTypes              []oauth2const.GrantType
	ResponseTypes           []oauth2const.ResponseType
	TokenEndpointAuthMethod []oauth2const.TokenEndpointAuthMethod
	PKCERequired            bool
	Token                   *OAuthTokenConfig
}

// IsAllowedGrantType checks if the provided grant type is allowed.
func (o *OAuthAppConfigDTO) IsAllowedGrantType(grantType oauth2const.GrantType) bool {
	return isAllowedGrantType(o.GrantTypes, grantType)
}

// IsAllowedResponseType checks if the provided response type is allowed.
func (o *OAuthAppConfigDTO) IsAllowedResponseType(responseType string) bool {
	return isAllowedResponseType(o.ResponseTypes, responseType)
}

// IsAllowedTokenEndpointAuthMethod checks if the provided token endpoint authentication method is allowed.
func (o *OAuthAppConfigDTO) IsAllowedTokenEndpointAuthMethod(method oauth2const.TokenEndpointAuthMethod) bool {
	return isAllowedTokenEndpointAuthMethod(o.TokenEndpointAuthMethod, method)
}

// ValidateRedirectURI validates the provided redirect URI against the registered redirect URIs.
func (o *OAuthAppConfigDTO) ValidateRedirectURI(redirectURI string) error {
	return validateRedirectURI(o.RedirectURIs, redirectURI)
}

// OAuthAppConfigProcessedDTO represents the processed data transfer object for OAuth application configuration.
type OAuthAppConfigProcessedDTO struct {
	AppID                   string
	ClientID                string
	HashedClientSecret      string
	RedirectURIs            []string
	GrantTypes              []oauth2const.GrantType
	ResponseTypes           []oauth2const.ResponseType
	TokenEndpointAuthMethod []oauth2const.TokenEndpointAuthMethod
	PKCERequired            bool
	Token                   *OAuthTokenConfig
}

// IsAllowedGrantType checks if the provided grant type is allowed.
func (o *OAuthAppConfigProcessedDTO) IsAllowedGrantType(grantType oauth2const.GrantType) bool {
	return isAllowedGrantType(o.GrantTypes, grantType)
}

// IsAllowedResponseType checks if the provided response type is allowed.
func (o *OAuthAppConfigProcessedDTO) IsAllowedResponseType(responseType string) bool {
	return isAllowedResponseType(o.ResponseTypes, responseType)
}

// IsAllowedTokenEndpointAuthMethod checks if the provided token endpoint authentication method is allowed.
func (o *OAuthAppConfigProcessedDTO) IsAllowedTokenEndpointAuthMethod(method oauth2const.TokenEndpointAuthMethod) bool {
	return isAllowedTokenEndpointAuthMethod(o.TokenEndpointAuthMethod, method)
}

// ValidateRedirectURI validates the provided redirect URI against the registered redirect URIs.
func (o *OAuthAppConfigProcessedDTO) ValidateRedirectURI(redirectURI string) error {
	return validateRedirectURI(o.RedirectURIs, redirectURI)
}

// RequiresPKCE checks if PKCE is required for this application.
func (o *OAuthAppConfigProcessedDTO) RequiresPKCE() bool {
	return o.PKCERequired || o.IsPublicClient()
}

// IsPublicClient checks if this is a public client.
func (o *OAuthAppConfigProcessedDTO) IsPublicClient() bool {
	return len(o.TokenEndpointAuthMethod) == 1 &&
		slices.Contains(o.TokenEndpointAuthMethod, oauth2const.TokenEndpointAuthMethodNone)
}

// isAllowedGrantType checks if the provided grant type is in the allowed list.
func isAllowedGrantType(grantTypes []oauth2const.GrantType, grantType oauth2const.GrantType) bool {
	if grantType == "" {
		return false
	}
	return slices.Contains(grantTypes, grantType)
}

// isAllowedResponseType checks if the provided response type is in the allowed list.
func isAllowedResponseType(responseTypes []oauth2const.ResponseType, responseType string) bool {
	if responseType == "" {
		return false
	}
	return slices.Contains(responseTypes, oauth2const.ResponseType(responseType))
}

// isAllowedTokenEndpointAuthMethod checks if the provided token authentication method is in the allowed list.
func isAllowedTokenEndpointAuthMethod(methods []oauth2const.TokenEndpointAuthMethod,
	method oauth2const.TokenEndpointAuthMethod) bool {
	if method == "" {
		return false
	}
	return slices.Contains(methods, method)
}

// validateRedirectURI checks if the provided redirect URI is valid against the registered redirect URIs.
func validateRedirectURI(redirectURIs []string, redirectURI string) error {
	logger := log.GetLogger()

	// Check if the redirect URI is empty.
	if redirectURI == "" {
		// Check if multiple redirect URIs are registered.
		if len(redirectURIs) != 1 {
			return fmt.Errorf("redirect URI is required in the authorization request")
		}
		// Check if only a part of the redirect uri is registered.
		parsed, err := url.Parse(redirectURIs[0])
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return fmt.Errorf("registered redirect URI is not fully qualified")
		}

		// Valid scenario.
		return nil
	}

	// Check if the redirect URI is registered.
	if !slices.Contains(redirectURIs, redirectURI) {
		return fmt.Errorf("your application's redirect URL does not match with the registered redirect URLs")
	}

	// Parse the redirect URI.
	parsedRedirectURI, err := utils.ParseURL(redirectURI)
	if err != nil {
		logger.Error("Failed to parse redirect URI", log.Error(err))
		return fmt.Errorf("invalid redirect URI: %s", err.Error())
	}
	// Check if it is a fragment URI.
	if parsedRedirectURI.Fragment != "" {
		return fmt.Errorf("redirect URI must not contain a fragment component")
	}

	return nil
}
