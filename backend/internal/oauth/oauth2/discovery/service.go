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

package discovery

import (
	"fmt"

	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/pkce"
	"github.com/asgardeo/thunder/internal/system/config"
)

// DiscoveryServiceInterface defines the interface for discovery services
type DiscoveryServiceInterface interface {
	GetOAuth2AuthorizationServerMetadata() *OAuth2AuthorizationServerMetadata
	GetOIDCMetadata() *OIDCProviderMetadata
}

// discoveryService implements DiscoveryServiceInterface
type discoveryService struct {
	baseURL string
}

// NewDiscoveryService creates a new discovery service instance
func newDiscoveryService() DiscoveryServiceInterface {
	ds := &discoveryService{}
	ds.baseURL = getBaseURL()
	return ds
}

// GetOAuth2AuthorizationServerMetadata returns OAuth 2.0 Authorization Server Metadata
func (ds *discoveryService) GetOAuth2AuthorizationServerMetadata() *OAuth2AuthorizationServerMetadata {
	return &OAuth2AuthorizationServerMetadata{
		Issuer:                            ds.getIssuer(),
		AuthorizationEndpoint:             ds.getAuthorizationEndpoint(),
		TokenEndpoint:                     ds.getTokenEndpoint(),
		JWKSUri:                           ds.getJWKSUri(),
		RegistrationEndpoint:              ds.getRegistrationEndpoint(),
		IntrospectionEndpoint:             ds.getIntrospectionEndpoint(),
		ScopesSupported:                   ds.getSupportedScopes(),
		ResponseTypesSupported:            ds.getSupportedResponseTypes(),
		GrantTypesSupported:               ds.getSupportedGrantTypes(),
		TokenEndpointAuthMethodsSupported: ds.getSupportedTokenEndpointAuthMethods(),
		CodeChallengeMethodsSupported:     ds.getSupportedCodeChallengeMethods(),
	}
}

// GetOIDCMetadata returns OpenID Connect Provider Metadata
func (ds *discoveryService) GetOIDCMetadata() *OIDCProviderMetadata {
	oauth2Meta := ds.GetOAuth2AuthorizationServerMetadata()

	return &OIDCProviderMetadata{
		OAuth2AuthorizationServerMetadata: *oauth2Meta,
		SubjectTypesSupported:             ds.getSupportedSubjectTypes(),
		IDTokenSigningAlgValuesSupported:  ds.getSupportedIDTokenSigningAlgorithms(),
		ClaimsSupported:                   ds.getSupportedClaims(),
	}
}

// Helper methods for building URLs and discovering capabilities
func getBaseURL() string {
	runtime := config.GetThunderRuntime()
	if runtime.Config.Server.PublicHostname != "" {
		return runtime.Config.Server.PublicHostname
	}
	scheme := "https"
	if runtime.Config.Server.HTTPOnly {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s:%d",
		scheme,
		runtime.Config.Server.Hostname,
		runtime.Config.Server.Port)
}

func (ds *discoveryService) getIssuer() string {
	return config.GetThunderRuntime().Config.JWT.Issuer
}

func (ds *discoveryService) getAuthorizationEndpoint() string {
	return ds.baseURL + constants.OAuth2AuthorizationEndpoint
}

func (ds *discoveryService) getTokenEndpoint() string {
	return ds.baseURL + constants.OAuth2TokenEndpoint
}

func (ds *discoveryService) getJWKSUri() string {
	return ds.baseURL + constants.OAuth2JWKSEndpoint
}

func (ds *discoveryService) getIntrospectionEndpoint() string {
	return ds.baseURL + constants.OAuth2IntrospectionEndpoint
}

func (ds *discoveryService) getRegistrationEndpoint() string {
	return ds.baseURL + constants.OAuth2DCREndpoint
}

func (ds *discoveryService) getSupportedScopes() []string {
	scopes := make([]string, 0, len(constants.StandardOIDCScopes))
	for scope := range constants.StandardOIDCScopes {
		scopes = append(scopes, scope)
	}
	return scopes
}

func (ds *discoveryService) getSupportedResponseTypes() []string {
	return constants.GetSupportedResponseTypes()
}

func (ds *discoveryService) getSupportedGrantTypes() []string {
	return constants.GetSupportedGrantTypes()
}

func (ds *discoveryService) getSupportedTokenEndpointAuthMethods() []string {
	return constants.GetSupportedTokenEndpointAuthMethods()
}

func (ds *discoveryService) getSupportedCodeChallengeMethods() []string {
	return pkce.GetSupportedCodeChallengeMethods()
}

func (ds *discoveryService) getSupportedSubjectTypes() []string {
	return constants.GetSupportedSubjectTypes()
}

func (ds *discoveryService) getSupportedIDTokenSigningAlgorithms() []string {
	return constants.GetSupportedIDTokenSigningAlgorithms()
}

func (ds *discoveryService) getSupportedClaims() []string {
	// Extract claims from OIDC scopes
	var claims []string
	claims = append(claims, constants.GetStandardClaims()...)

	for _, scope := range constants.StandardOIDCScopes {
		claims = append(claims, scope.Claims...)
	}

	// Remove duplicates
	claimMap := make(map[string]bool)
	var uniqueClaims []string
	for _, claim := range claims {
		if !claimMap[claim] {
			claimMap[claim] = true
			uniqueClaims = append(uniqueClaims, claim)
		}
	}

	return uniqueClaims
}
