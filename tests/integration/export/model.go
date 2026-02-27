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

package export

// ExportRequest represents the request structure for exporting resources.
type ExportRequest struct {
	Applications      []string `json:"applications,omitempty"`
	IdentityProviders []string `json:"identity_providers,omitempty"`
}

// ExportResponse represents the response structure for exporting resources.
type ExportResponse struct {
	Files []ExportFile `json:"files"`
}

// ExportFile represents a single YAML file in the export response.
type ExportFile struct {
	FileName string `json:"file_name"`
	Content  string `json:"content"`
}

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
	URL                       string              `json:"url,omitempty"`
	LogoURL                   string              `json:"logo_url,omitempty"`
	Certificate               *ApplicationCert    `json:"certificate,omitempty"`
	Assertion                 *AssertionConfig    `json:"assertion,omitempty"`
	TosURI                    string              `json:"tos_uri,omitempty"`
	PolicyURI                 string              `json:"policy_uri,omitempty"`
	Contacts                  []string            `json:"contacts,omitempty"`
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
	ClientID                string            `json:"client_id"`
	ClientSecret            string            `json:"client_secret,omitempty"`
	RedirectURIs            []string          `json:"redirect_uris"`
	GrantTypes              []string          `json:"grant_types"`
	ResponseTypes           []string          `json:"response_types"`
	TokenEndpointAuthMethod string            `json:"token_endpoint_auth_method"`
	PKCERequired            bool              `json:"pkce_required"`
	PublicClient            bool              `json:"public_client"`
	Scopes                  []string          `json:"scopes,omitempty"`
	Token                   *OAuthTokenConfig `json:"token,omitempty"`
}

// OAuthTokenConfig represents the OAuth token configuration.
type OAuthTokenConfig struct {
	AccessToken *AccessTokenConfig `json:"access_token,omitempty"`
	IDToken     *IDTokenConfig     `json:"id_token,omitempty"`
}

// AssertionConfig represents the assertion configuration (for application-level).
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

// IDPProperty represents a property of an identity provider.
type IDPProperty struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	IsSecret bool   `json:"is_secret"`
}

// IDP represents an identity provider.
type IDP struct {
	ID          string        `json:"id,omitempty"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Type        string        `json:"type"`
	Properties  []IDPProperty `json:"properties"`
}
