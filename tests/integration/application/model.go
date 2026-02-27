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

package application

// Application represents the structure for application request and response in tests.
type Application struct {
	ID                        string              `json:"id,omitempty"`
	Name                      string              `json:"name"`
	Description               string              `json:"description,omitempty"`
	ClientID                  string              `json:"client_id,omitempty"`
	ClientSecret              string              `json:"client_secret,omitempty"`
	AuthFlowID                string              `json:"auth_flow_id,omitempty"`
	RegistrationFlowID        string              `json:"registration_flow_id,omitempty"`
	IsRegistrationFlowEnabled bool                `json:"is_registration_flow_enabled"`
	ThemeID                   string              `json:"theme_id,omitempty"`
	LayoutID                  string              `json:"layout_id,omitempty"`
	Template                  string              `json:"template,omitempty"`
	URL                       string              `json:"url,omitempty"`
	LogoURL                   string              `json:"logo_url,omitempty"`
	Certificate               *ApplicationCert    `json:"certificate,omitempty"`
	Assertion                 *AssertionConfig    `json:"assertion,omitempty"`
	TosURI                    string              `json:"tos_uri,omitempty"`
	PolicyURI                 string              `json:"policy_uri,omitempty"`
	Contacts                  []string            `json:"contacts,omitempty"`
	AllowedUserTypes          []string            `json:"allowed_user_types,omitempty"`
	InboundAuthConfig         []InboundAuthConfig `json:"inbound_auth_config,omitempty"`
}

// ApplicationCert represents the certificate structure in the application.
type ApplicationCert struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// InboundAuthConfig represents the inbound authentication configuration.
type InboundAuthConfig struct {
	Type           string          `json:"type"`
	OAuthAppConfig *OAuthAppConfig `json:"config,omitempty"`
}

// OAuthAppConfig represents the OAuth application configuration.
type OAuthAppConfig struct {
	ClientID                string              `json:"client_id"`
	ClientSecret            string              `json:"client_secret,omitempty"`
	RedirectURIs            []string            `json:"redirect_uris"`
	GrantTypes              []string            `json:"grant_types"`
	ResponseTypes           []string            `json:"response_types"`
	TokenEndpointAuthMethod string              `json:"token_endpoint_auth_method"`
	PKCERequired            bool                `json:"pkce_required"`
	PublicClient            bool                `json:"public_client"`
	Scopes                  []string            `json:"scopes,omitempty"`
	Token                   *OAuthTokenConfig   `json:"token,omitempty"`
	ScopeClaims             map[string][]string `json:"scope_claims,omitempty"`
	UserInfo                *UserInfoConfig     `json:"user_info,omitempty"`
}

// OAuthTokenConfig represents the OAuth token configuration.
type OAuthTokenConfig struct {
	AccessToken *AccessTokenConfig `json:"access_token,omitempty"`
	IDToken     *IDTokenConfig     `json:"id_token,omitempty"`
}

// UserInfoConfig represents the UserInfo endpoint configuration.
type UserInfoConfig struct {
	ResponseType   string   `json:"response_type,omitempty"`
	UserAttributes []string `json:"user_attributes,omitempty"`
}

// AssertionConfig represents the assertion configuration (used for application-level assertion config).
type AssertionConfig struct {
	ValidityPeriod int64    `json:"validity_period,omitempty"`
	UserAttributes []string `json:"user_attributes,omitempty"`
}

// AccessTokenConfig represents the access token configuration.
type AccessTokenConfig struct {
	ValidityPeriod int64    `json:"validity_period,omitempty"`
	UserAttributes []string `json:"user_attributes,omitempty"`
}

// IDTokenConfig represents the ID token configuration.
type IDTokenConfig struct {
	ValidityPeriod int64    `json:"validity_period,omitempty"`
	UserAttributes []string `json:"user_attributes,omitempty"`
}

// ApplicationList represents the response structure for listing applications.
type ApplicationList struct {
	TotalResults int           `json:"totalResults"`
	Count        int           `json:"count"`
	Applications []Application `json:"applications"`
}

func compareStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (app *Application) equals(expectedApp Application) bool {
	// Basic fields
	if app.ID != expectedApp.ID ||
		app.Name != expectedApp.Name ||
		app.Description != expectedApp.Description {
		return false
	}

	// For ClientID, we need to handle it being in both the root and OAuth config
	if app.ClientID != expectedApp.ClientID {
		return false
	}

	// Auth flow fields
	if app.AuthFlowID != expectedApp.AuthFlowID ||
		app.RegistrationFlowID != expectedApp.RegistrationFlowID ||
		app.IsRegistrationFlowEnabled != expectedApp.IsRegistrationFlowEnabled {
		return false
	}

	// Theme and Layout IDs
	if app.ThemeID != expectedApp.ThemeID || app.LayoutID != expectedApp.LayoutID {
		return false
	}

	// Template
	if app.Template != expectedApp.Template {
		return false
	}

	// URL fields
	if app.URL != expectedApp.URL ||
		app.LogoURL != expectedApp.LogoURL {
		return false
	}

	// Metadata fields
	if app.TosURI != expectedApp.TosURI ||
		app.PolicyURI != expectedApp.PolicyURI {
		return false
	}

	// Contacts
	if !compareStringSlices(app.Contacts, expectedApp.Contacts) {
		return false
	}

	// AllowedUserTypes
	if !compareStringSlices(app.AllowedUserTypes, expectedApp.AllowedUserTypes) {
		return false
	}

	// Assertion config
	if (app.Assertion != nil) && (expectedApp.Assertion != nil) {
		if app.Assertion.ValidityPeriod != expectedApp.Assertion.ValidityPeriod {
			return false
		}
		if !compareStringSlices(app.Assertion.UserAttributes, expectedApp.Assertion.UserAttributes) {
			return false
		}
	} else if (app.Assertion == nil && expectedApp.Assertion != nil) ||
		(app.Assertion != nil && expectedApp.Assertion == nil) {
		return false
	}

	// ClientSecret is only checked when both have it (create/update operations)
	// Don't check it for get operations where it shouldn't be returned
	if app.ClientSecret != "" && expectedApp.ClientSecret != "" &&
		app.ClientSecret != expectedApp.ClientSecret {
		return false
	}

	// Check certificate - allow nil in expected if actual has default empty certificate
	if (app.Certificate != nil) && (expectedApp.Certificate == nil) {
		// If expected has no certificate but actual does, check if it's the default empty one
		if app.Certificate.Type != "NONE" || app.Certificate.Value != "" {
			return false
		}
	} else if (app.Certificate == nil) && (expectedApp.Certificate != nil) {
		return false
	} else if app.Certificate != nil && expectedApp.Certificate != nil {
		if app.Certificate.Type != expectedApp.Certificate.Type ||
			app.Certificate.Value != expectedApp.Certificate.Value {
			return false
		}
	}

	// Check inbound auth config if present
	if len(app.InboundAuthConfig) != len(expectedApp.InboundAuthConfig) {
		return false
	}

	// Compare inbound auth config details
	if len(app.InboundAuthConfig) > 0 {
		for i, cfg := range app.InboundAuthConfig {
			expectedCfg := expectedApp.InboundAuthConfig[i]
			if cfg.Type != expectedCfg.Type {
				return false
			}

			// Compare OAuth configs if they exist
			if cfg.OAuthAppConfig != nil && expectedCfg.OAuthAppConfig != nil {
				oauth := cfg.OAuthAppConfig
				expectedOAuth := expectedCfg.OAuthAppConfig

				// Compare the fields
				if oauth.ClientID != expectedOAuth.ClientID {
					return false
				}

				if !compareStringSlices(oauth.RedirectURIs, expectedOAuth.RedirectURIs) {
					return false
				}

				if !compareStringSlices(oauth.GrantTypes, expectedOAuth.GrantTypes) {
					return false
				}

				if !compareStringSlices(oauth.ResponseTypes, expectedOAuth.ResponseTypes) {
					return false
				}

				if oauth.TokenEndpointAuthMethod != expectedOAuth.TokenEndpointAuthMethod {
					return false
				}

				if oauth.PKCERequired != expectedOAuth.PKCERequired {
					return false
				}

				if oauth.PublicClient != expectedOAuth.PublicClient {
					return false
				}

				// Compare ScopeClaims - lenient if expected is nil but actual is empty
				if expectedOAuth.ScopeClaims != nil {
					if !compareScopeClaimsMaps(oauth.ScopeClaims, expectedOAuth.ScopeClaims) {
						return false
					}
				}

				// Compare UserInfo config - lenient if expected is nil but actual is empty
				if expectedOAuth.UserInfo != nil {
					if oauth.UserInfo == nil {
						return false
					}
					if oauth.UserInfo.ResponseType != expectedOAuth.UserInfo.ResponseType {
						return false
					}
					if !compareStringSlices(oauth.UserInfo.UserAttributes, expectedOAuth.UserInfo.UserAttributes) {
						return false
					}
				}
				// If expected UserInfo is nil, we accept any value in actual (including empty object)
			} else if (cfg.OAuthAppConfig == nil && expectedCfg.OAuthAppConfig != nil) ||
				(cfg.OAuthAppConfig != nil && expectedCfg.OAuthAppConfig == nil) {
				return false
			}
		}
	}

	return true
}

// compareScopeClaimsMaps compares two scope claims maps for equality.
func compareScopeClaimsMaps(a, b map[string][]string) bool {
	if len(a) != len(b) {
		return false
	}
	for key, aVal := range a {
		bVal, exists := b[key]
		if !exists {
			return false
		}
		if !compareStringSlices(aVal, bVal) {
			return false
		}
	}
	return true
}
