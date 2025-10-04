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

// Package constants defines constants used across the OAuth2 module.
package constants

import "errors"

// OAuth2 request parameters.
const (
	RequestParamGrantType           string = "grant_type"
	RequestParamClientID            string = "client_id"
	RequestParamClientSecret        string = "client_secret"
	RequestParamRedirectURI         string = "redirect_uri"
	RequestParamUsername            string = "username"
	RequestParamPassword            string = "password"
	RequestParamScope               string = "scope"
	RequestParamCode                string = "code"
	RequestParamCodeVerifier        string = "code_verifier"
	RequestParamCodeChallenge       string = "code_challenge"
	RequestParamCodeChallengeMethod string = "code_challenge_method"
	RequestParamRefreshToken        string = "refresh_token"
	RequestParamResponseType        string = "response_type"
	RequestParamState               string = "state"
	RequestParamError               string = "error"
	RequestParamErrorDescription    string = "error_description"
	RequestParamToken               string = "token"
	RequestParamTokenTypeHint       string = "token_type_hint"
)

// Server OAuth constants.
const (
	SessionDataKey        string = "sessionDataKey"
	SessionDataKeyConsent string = "sessionDataKeyConsent"
	ShowInsecureWarning   string = "showInsecureWarning"
	AppID                 string = "applicationId"
	Assertion             string = "assertion"
)

// Oauth message types.
const (
	TypeInitialAuthorizationRequest     string = "initialAuthorizationRequest"
	TypeAuthorizationResponseFromEngine string = "authorizationResponseFromEngine"
	TypeConsentResponseFromUser         string = "consentResponseFromUser"
)

// OAuth2 endpoints.
const (
	OAuth2TokenEndpoint         string = "/oauth2/token" // #nosec G101
	OAuth2AuthorizationEndpoint string = "/oauth2/authorize"
	OAuth2IntrospectionEndpoint string = "/oauth2/introspect"
	OAuth2RevokeEndpoint        string = "/oauth2/revoke"
	OAuth2UserInfoEndpoint      string = "/oauth2/userinfo"
	OAuth2JWKSEndpoint          string = "/oauth2/jwks"
	OAuth2LogoutEndpoint        string = "/oauth2/logout"
)

// GrantType defines a type for OAuth2 grant types.
type GrantType string

const (
	// GrantTypeAuthorizationCode represents the authorization code grant type.
	GrantTypeAuthorizationCode GrantType = "authorization_code"
	// GrantTypeClientCredentials represents the client credentials grant type.
	GrantTypeClientCredentials GrantType = "client_credentials"
	// GrantTypePassword represents the resource owner password credentials grant type.
	GrantTypePassword GrantType = "password"
	// GrantTypeImplicit represents the implicit grant type.
	GrantTypeImplicit GrantType = "implicit"
	// GrantTypeRefreshToken represents the refresh token grant type.
	GrantTypeRefreshToken GrantType = "refresh_token"
)

// IsValid checks if the GrantType is valid.
func (gt GrantType) IsValid() bool {
	switch gt {
	case GrantTypeAuthorizationCode,
		GrantTypeClientCredentials,
		GrantTypePassword,
		GrantTypeImplicit,
		GrantTypeRefreshToken:
		return true
	default:
		return false
	}
}

// ResponseType defines a type for OAuth2 response types.
type ResponseType string

const (
	// ResponseTypeCode represents the authorization code response type.
	ResponseTypeCode ResponseType = "code"
	// ResponseTypeToken represents the implicit token response type.
	ResponseTypeToken ResponseType = "token"
)

// IsValid checks if the ResponseType is valid.
func (rt ResponseType) IsValid() bool {
	switch rt {
	case ResponseTypeCode:
		return true
	default:
		return false
	}
}

// TokenEndpointAuthMethod defines a type for token endpoint authentication methods.
type TokenEndpointAuthMethod string

const (
	// TokenEndpointAuthMethodClientSecretBasic represents the client secret basic authentication method.
	TokenEndpointAuthMethodClientSecretBasic TokenEndpointAuthMethod = "client_secret_basic"
	// TokenEndpointAuthMethodClientSecretPost represents the client secret post authentication method.
	TokenEndpointAuthMethodClientSecretPost TokenEndpointAuthMethod = "client_secret_post"
	// TokenEndpointAuthMethodNone represents no authentication method.
	TokenEndpointAuthMethodNone TokenEndpointAuthMethod = "none"
)

// IsValid checks if the TokenEndpointAuthMethod is valid.
func (tam TokenEndpointAuthMethod) IsValid() bool {
	switch tam {
	case TokenEndpointAuthMethodClientSecretBasic,
		TokenEndpointAuthMethodClientSecretPost,
		TokenEndpointAuthMethodNone:
		return true
	default:
		return false
	}
}

// OAuth2 token types.
const (
	TokenTypeBearer = "Bearer"
)

// OAuth2 error codes.
const (
	ErrorInvalidRequest          string = "invalid_request"
	ErrorInvalidClient           string = "invalid_client"
	ErrorInvalidGrant            string = "invalid_grant"
	ErrorUnauthorizedClient      string = "unauthorized_client"
	ErrorUnsupportedGrantType    string = "unsupported_grant_type"
	ErrorInvalidScope            string = "invalid_scope"
	ErrorServerError             string = "server_error"
	ErrorUnsupportedResponseType string = "unsupported_response_type"
	ErrorAccessDenied            string = "access_denied"
)

// UnSupportedGrantTypeError is returned when an unsupported grant type is requested.
var UnSupportedGrantTypeError = errors.New("unsupported_grant_type")
