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

// Package model defines OAuth-related types for inbound client configuration.
//
//nolint:lll
package model

import (
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/jose/jwe"
	"github.com/asgardeo/thunder/internal/system/jose/jws"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
	oauth2const "github.com/asgardeo/thunder/pkg/oauth/oauth2/constants"
)

// OAuthProfile is the OAuth inbound profile for a principal entity.
type OAuthProfile struct {
	AppID        string
	OAuthProfile *OAuthProfileData
}

// OAuthProfileData is the typed representation of the OAUTH_PROFILE JSONB column.
type OAuthProfileData struct {
	RedirectURIs                       []string            `json:"redirectUris"`
	GrantTypes                         []string            `json:"grantTypes"`
	ResponseTypes                      []string            `json:"responseTypes"`
	TokenEndpointAuthMethod            string              `json:"tokenEndpointAuthMethod"`
	PKCERequired                       bool                `json:"pkceRequired"`
	PublicClient                       bool                `json:"publicClient"`
	RequirePushedAuthorizationRequests bool                `json:"requirePushedAuthorizationRequests"`
	Token                              *OAuthTokenConfig   `json:"token,omitempty"`
	Scopes                             []string            `json:"scopes,omitempty"`
	UserInfo                           *UserInfoConfig     `json:"userInfo,omitempty"`
	ScopeClaims                        map[string][]string `json:"scopeClaims,omitempty"`
	Certificate                        *Certificate        `json:"certificate,omitempty"`
}

// OAuthTokenConfig wraps access and ID token configs.
type OAuthTokenConfig struct {
	AccessToken *AccessTokenConfig `json:"accessToken,omitempty" yaml:"access_token,omitempty" jsonschema:"Access token configuration."`
	IDToken     *IDTokenConfig     `json:"idToken,omitempty" yaml:"id_token,omitempty" jsonschema:"ID token configuration."`
}

// AccessTokenConfig is the access token configuration.
type AccessTokenConfig struct {
	ValidityPeriod int64    `json:"validityPeriod,omitempty" yaml:"validity_period,omitempty" jsonschema:"Access token validity period in seconds."`
	UserAttributes []string `json:"userAttributes,omitempty" yaml:"user_attributes,omitempty" jsonschema:"User attributes to embed in the access token."`
}

// IDTokenConfig is the ID token configuration.
type IDTokenConfig struct {
	ValidityPeriod int64    `json:"validityPeriod,omitempty" yaml:"validity_period,omitempty" jsonschema:"ID token validity period in seconds."`
	UserAttributes []string `json:"userAttributes,omitempty" yaml:"user_attributes,omitempty" jsonschema:"User attributes to embed in the ID token."`
}

// UserInfoConfig is the user info endpoint configuration.
type UserInfoConfig struct {
	ResponseType   UserInfoResponseType `json:"responseType,omitempty"   yaml:"response_type,omitempty"   jsonschema:"UserInfo response type (JSON, JWS, JWE, NESTED_JWT). Required algorithm fields must match the selected response type."`
	UserAttributes []string             `json:"userAttributes,omitempty" yaml:"user_attributes,omitempty" jsonschema:"User attributes to include in the userinfo response."`
	SigningAlg     string               `json:"signingAlg,omitempty"     yaml:"signing_alg,omitempty"     jsonschema:"JWS algorithm for signed userinfo responses (e.g. RS256)."`
	EncryptionAlg  string               `json:"encryptionAlg,omitempty"  yaml:"encryption_alg,omitempty"  jsonschema:"JWE key-management algorithm for encrypted userinfo responses (e.g. RSA-OAEP-256)."`
	EncryptionEnc  string               `json:"encryptionEnc,omitempty"  yaml:"encryption_enc,omitempty"  jsonschema:"JWE content-encryption algorithm (e.g. A256GCM). Required when encryptionAlg is set."`
}

// UserInfoResponseType is the response format of the UserInfo endpoint.
type UserInfoResponseType string

const (
	// UserInfoResponseTypeJSON is the JSON userinfo response type.
	UserInfoResponseTypeJSON UserInfoResponseType = "JSON"
	// UserInfoResponseTypeJWS is the signed JWT (JWS) userinfo response type.
	UserInfoResponseTypeJWS UserInfoResponseType = "JWS"
	// UserInfoResponseTypeJWE is the encrypted (JWE) userinfo response type.
	UserInfoResponseTypeJWE UserInfoResponseType = "JWE"
	// UserInfoResponseTypeNESTEDJWT is the signed-then-encrypted (Nested JWT) userinfo response type.
	UserInfoResponseTypeNESTEDJWT UserInfoResponseType = "NESTED_JWT"
)

// SupportedUserInfoSigningAlgs lists JWS algorithms supported for userinfo signing.
var SupportedUserInfoSigningAlgs = []string{
	string(jws.RS256), string(jws.RS512), string(jws.PS256),
	string(jws.ES256), string(jws.ES384), string(jws.ES512),
	string(jws.EdDSA),
}

// SupportedUserInfoEncryptionAlgs lists JWE key-management algorithms supported for userinfo encryption.
var SupportedUserInfoEncryptionAlgs = []string{string(jwe.RSAOAEP), string(jwe.RSAOAEP256)}

// SupportedUserInfoEncryptionEncs lists JWE content-encryption algorithms supported for userinfo encryption.
var SupportedUserInfoEncryptionEncs = []string{string(jwe.A128CBCHS256), string(jwe.A256GCM)}

// OAuthClient is the resolved OAuth-client view used by the OAuth machinery (token
// issuance, grant handlers, userinfo, authz, dcr). Both application and agent services
// build OAuthClient from their consumer-specific DTOs so the OAuth machinery is
// consumer-agnostic.
type OAuthClient struct {
	AppID                              string                              `yaml:"app_id,omitempty"`
	OUID                               string                              `yaml:"ou_id,omitempty"`
	ClientID                           string                              `yaml:"client_id,omitempty"`
	RedirectURIs                       []string                            `yaml:"redirect_uris,omitempty"`
	GrantTypes                         []oauth2const.GrantType             `yaml:"grant_types,omitempty"`
	ResponseTypes                      []oauth2const.ResponseType          `yaml:"response_types,omitempty"`
	TokenEndpointAuthMethod            oauth2const.TokenEndpointAuthMethod `yaml:"token_endpoint_auth_method,omitempty"`
	PKCERequired                       bool                                `yaml:"pkce_required,omitempty"`
	PublicClient                       bool                                `yaml:"public_client,omitempty"`
	RequirePushedAuthorizationRequests bool                                `yaml:"require_pushed_authorization_requests,omitempty"`
	Token                              *OAuthTokenConfig                   `yaml:"token,omitempty"`
	Scopes                             []string                            `yaml:"scopes,omitempty"`
	UserInfo                           *UserInfoConfig                     `yaml:"user_info,omitempty"`
	ScopeClaims                        map[string][]string                 `yaml:"scope_claims,omitempty"`
	Certificate                        *Certificate                        `yaml:"certificate,omitempty"`
}

// IsAllowedGrantType reports whether the OAuth client allows the given grant type.
func (o *OAuthClient) IsAllowedGrantType(grantType oauth2const.GrantType) bool {
	return IsAllowedGrantType(o.GrantTypes, grantType)
}

// IsAllowedResponseType reports whether the OAuth client allows the given response type.
func (o *OAuthClient) IsAllowedResponseType(responseType string) bool {
	return IsAllowedResponseType(o.ResponseTypes, responseType)
}

// IsAllowedTokenEndpointAuthMethod reports whether the OAuth client uses the given auth method.
func (o *OAuthClient) IsAllowedTokenEndpointAuthMethod(method oauth2const.TokenEndpointAuthMethod) bool {
	return o.TokenEndpointAuthMethod == method
}

// ValidateRedirectURI validates the given redirect URI against the registered list.
func (o *OAuthClient) ValidateRedirectURI(redirectURI string) error {
	return ValidateRedirectURI(o.RedirectURIs, redirectURI)
}

// RequiresPKCE reports whether PKCE is required for this OAuth client.
func (o *OAuthClient) RequiresPKCE() bool {
	return o.PKCERequired || o.PublicClient
}

// RequiresPAR reports whether pushed authorization requests are required for this OAuth client.
func (o *OAuthClient) RequiresPAR() bool {
	return o.RequirePushedAuthorizationRequests || config.GetServerRuntime().Config.OAuth.PAR.RequirePAR
}

// IsAllowedGrantType reports whether the given grant type is in the allowed list.
// Returns false for an empty input grant type.
func IsAllowedGrantType(grantTypes []oauth2const.GrantType, grantType oauth2const.GrantType) bool {
	if grantType == "" {
		return false
	}
	return slices.Contains(grantTypes, grantType)
}

// IsAllowedResponseType reports whether the given response type is in the allowed list.
// Returns false for an empty input response type.
func IsAllowedResponseType(responseTypes []oauth2const.ResponseType, responseType string) bool {
	if responseType == "" {
		return false
	}
	return slices.Contains(responseTypes, oauth2const.ResponseType(responseType))
}

// ValidateRedirectURI validates the provided redirect URI against the registered redirect URIs.
func ValidateRedirectURI(redirectURIs []string, redirectURI string) error {
	logger := log.GetLogger()

	if redirectURI == "" {
		if len(redirectURIs) != 1 {
			return fmt.Errorf("redirect URI is required in the authorization request")
		}
		// AC-12: A wildcard pattern cannot serve as a concrete redirect target.
		if strings.Contains(redirectURIs[0], "*") {
			return fmt.Errorf("redirect URI is required in the authorization request")
		}
		parsed, err := url.Parse(redirectURIs[0])
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return fmt.Errorf("registered redirect URI is not fully qualified")
		}
		return nil
	}

	if !matchAnyRedirectURIPattern(redirectURIs, redirectURI) {
		return fmt.Errorf("your application's redirect URL does not match with the registered redirect URLs")
	}

	parsedRedirectURI, err := utils.ParseURL(redirectURI)
	if err != nil {
		logger.Error("Failed to parse redirect URI", log.Error(err))
		return fmt.Errorf("invalid redirect URI: %s", err.Error())
	}
	if parsedRedirectURI.Fragment != "" {
		return fmt.Errorf("redirect URI must not contain a fragment component")
	}

	return nil
}

// matchAnyRedirectURIPattern checks incoming against each registered URI or pattern.
// Exact URIs are compared directly; patterns containing * use wildcard path matching.
// First match wins (AC-11). Wildcard matching is skipped when the feature flag is off.
func matchAnyRedirectURIPattern(patterns []string, redirectURI string) bool {
	wildcardEnabled := config.GetServerRuntime().Config.OAuth.AllowWildcardRedirectURI
	for _, pattern := range patterns {
		if !wildcardEnabled || !strings.Contains(pattern, "*") {
			if pattern == redirectURI {
				return true
			}
			continue
		}
		matched, err := utils.MatchURIPattern(pattern, redirectURI)
		if err != nil {
			continue
		}
		if matched {
			return true
		}
	}
	return false
}
