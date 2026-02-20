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
	"github.com/asgardeo/thunder/internal/cert"
)

// AssertionConfig represents the assertion configuration structure for application-level (root) assertion configs.
type AssertionConfig struct {
	ValidityPeriod int64    `json:"validity_period,omitempty" yaml:"validity_period,omitempty" jsonschema:"Assertion validity period in seconds."`
	UserAttributes []string `json:"user_attributes,omitempty" yaml:"user_attributes,omitempty" jsonschema:"User attributes to include in the assertion. List of user claim names to embed in the assertion (e.g., email, username, roles)."`
}

// ApplicationDTO represents the data transfer object for application service operations.
type ApplicationDTO struct {
	ID                        string `json:"id,omitempty" jsonschema:"Application ID. Auto-generated unique identifier."`
	Name                      string `json:"name" jsonschema:"Application name."`
	Description               string `json:"description,omitempty" jsonschema:"Optional description of the application's purpose or functionality."`
	AuthFlowID                string `json:"auth_flow_id,omitempty" jsonschema:"Authentication flow ID. Optional. Specifies which login flow to use (e.g., MFA, passwordless). Use list_flows to find available flows. If omitted, the default authentication flow is used."`
	RegistrationFlowID        string `json:"registration_flow_id,omitempty" jsonschema:"Registration flow ID. Optional. Specifies the user registration/signup flow. Use list_flows to find available flows."`
	IsRegistrationFlowEnabled bool   `json:"is_registration_flow_enabled,omitempty" jsonschema:"Enable self-service registration. Set to true to allow users to sign up themselves. Requires registration_flow_id to be set."`
	ThemeID                   string `json:"theme_id,omitempty" jsonschema:"Theme configuration ID. Optional. Customizes the visual styling (colors, typography) of login pages."`
	LayoutID                  string `json:"layout_id,omitempty" jsonschema:"Layout configuration ID. Optional. Customizes the screen structure and component positioning of login pages."`
	Template                  string `json:"template,omitempty" jsonschema:"Application template. Optional. Pre-configured application type template."`

	URL       string   `json:"url,omitempty" jsonschema:"Application home URL. Optional. The main URL where your application is hosted."`
	LogoURL   string   `json:"logo_url,omitempty" jsonschema:"Logo image URL. Optional. Displayed in login pages and application listings."`
	TosURI    string   `json:"tos_uri,omitempty" jsonschema:"Terms of Service URI. Optional. Link to your application's terms of service."`
	PolicyURI string   `json:"policy_uri,omitempty" jsonschema:"Privacy Policy URI. Optional. Link to your application's privacy policy."`
	Contacts  []string `json:"contacts,omitempty" jsonschema:"Contact email addresses. Optional. Administrative contact emails for this application."`

	Assertion         *AssertionConfig        `json:"assertion,omitempty" jsonschema:"Assertion configuration. Optional. Customize assertion validity periods and included user attributes."`
	Certificate       *ApplicationCertificate `json:"certificate,omitempty" jsonschema:"Application certificate. Optional. For certificate-based authentication or JWT validation."`
	InboundAuthConfig []InboundAuthConfigDTO  `json:"inbound_auth_config,omitempty" jsonschema:"OAuth/OIDC authentication configuration. Required for OAuth-enabled applications. Configure OAuth grant types, redirect URIs, and client authentication methods."`
	AllowedUserTypes  []string                `json:"allowed_user_types,omitempty" jsonschema:"Allowed user types. Optional. Restricts which types of users can register to this application."`
}

// BasicApplicationDTO represents a simplified data transfer object for application service operations.
type BasicApplicationDTO struct {
	ID                        string
	Name                      string
	Description               string
	AuthFlowID                string
	RegistrationFlowID        string
	IsRegistrationFlowEnabled bool
	ThemeID                   string
	LayoutID                  string
	Template                  string
	ClientID                  string
	LogoURL                   string
}

// Application represents the structure for application which returns in GetApplicationById.
type Application struct {
	ID                        string `yaml:"id,omitempty" json:"id,omitempty" jsonschema:"Application ID. Auto-generated unique identifier."`
	Name                      string `yaml:"name,omitempty" json:"name,omitempty" jsonschema:"Application name."`
	Description               string `yaml:"description,omitempty" json:"description,omitempty" jsonschema:"Optional description of the application's purpose."`
	AuthFlowID                string `yaml:"auth_flow_id,omitempty" json:"auth_flow_id,omitempty" jsonschema:"Associated authentication flow ID."`
	RegistrationFlowID        string `yaml:"registration_flow_id,omitempty" json:"registration_flow_id,omitempty" jsonschema:"Associated registration flow ID."`
	IsRegistrationFlowEnabled bool   `yaml:"is_registration_flow_enabled,omitempty" json:"is_registration_flow_enabled,omitempty" jsonschema:"Indicates if self-service registration is enabled."`
	ThemeID                   string `yaml:"theme_id,omitempty" json:"theme_id,omitempty" jsonschema:"Associated theme configuration ID."`
	LayoutID                  string `yaml:"layout_id,omitempty" json:"layout_id,omitempty" jsonschema:"Associated layout configuration ID."`
	Template                  string `yaml:"template,omitempty" json:"template,omitempty" jsonschema:"Template used to create the application."`

	URL       string   `yaml:"url,omitempty" json:"url,omitempty" jsonschema:"Application home URL."`
	LogoURL   string   `yaml:"logo_url,omitempty" json:"logo_url,omitempty" jsonschema:"Application logo URL."`
	TosURI    string   `yaml:"tos_uri,omitempty" json:"tos_uri,omitempty" jsonschema:"Terms of Service URI."`
	PolicyURI string   `yaml:"policy_uri,omitempty" json:"policy_uri,omitempty" jsonschema:"Privacy Policy URI."`
	Contacts  []string `yaml:"contacts,omitempty" json:"contacts,omitempty"`

	Assertion         *AssertionConfig            `yaml:"assertion,omitempty" json:"assertion,omitempty" jsonschema:"Assertion configuration settings."`
	Certificate       *ApplicationCertificate     `yaml:"certificate,omitempty" json:"certificate,omitempty" jsonschema:"Application certificate settings."`
	InboundAuthConfig []InboundAuthConfigComplete `yaml:"inbound_auth_config,omitempty" json:"inbound_auth_config,omitempty" jsonschema:"Inbound authentication configuration (OAuth2/OIDC settings)."`
	AllowedUserTypes  []string                    `yaml:"allowed_user_types,omitempty" json:"allowed_user_types,omitempty" jsonschema:"Allowed user types for registration."`
}

// ApplicationProcessedDTO represents the processed data transfer object for application service operations.
type ApplicationProcessedDTO struct {
	ID                        string `yaml:"id,omitempty"`
	Name                      string `yaml:"name,omitempty"`
	Description               string `yaml:"description,omitempty"`
	AuthFlowID                string `yaml:"auth_flow_id,omitempty"`
	RegistrationFlowID        string `yaml:"registration_flow_id,omitempty"`
	IsRegistrationFlowEnabled bool   `yaml:"is_registration_flow_enabled,omitempty"`
	ThemeID                   string `yaml:"theme_id,omitempty"`
	LayoutID                  string `yaml:"layout_id,omitempty"`
	Template                  string `yaml:"template,omitempty"`

	URL       string `yaml:"url,omitempty"`
	LogoURL   string `yaml:"logo_url,omitempty"`
	TosURI    string `yaml:"tos_uri,omitempty"`
	PolicyURI string `yaml:"policy_uri,omitempty"`
	Contacts  []string

	Assertion         *AssertionConfig                `yaml:"assertion,omitempty"`
	Certificate       *ApplicationCertificate         `yaml:"certificate,omitempty"`
	InboundAuthConfig []InboundAuthConfigProcessedDTO `yaml:"inbound_auth_config,omitempty"`
	AllowedUserTypes  []string                        `yaml:"allowed_user_types,omitempty"`
}

// InboundAuthConfigDTO represents the data transfer object for inbound authentication configuration.
// TODO: Need to refactor when supporting other/multiple inbound auth types.
type InboundAuthConfigDTO struct {
	Type           InboundAuthType    `json:"type" jsonschema:"Inbound authentication type. Use 'oauth2' for OAuth/OIDC applications."`
	OAuthAppConfig *OAuthAppConfigDTO `json:"config,omitempty" jsonschema:"OAuth/OIDC configuration. Required when type is 'oauth2'. Defines OAuth grant types, redirect URIs, client authentication, and PKCE settings."`
}

// InboundAuthConfigProcessedDTO represents the processed data transfer object for inbound authentication
// configuration.
type InboundAuthConfigProcessedDTO struct {
	Type           InboundAuthType             `json:"type" yaml:"type,omitempty"`
	OAuthAppConfig *OAuthAppConfigProcessedDTO `json:"config,omitempty" yaml:"config,omitempty"`
}

// ApplicationCertificate represents the certificate structure in the application request response.
type ApplicationCertificate struct {
	Type  cert.CertificateType `json:"type,omitempty" yaml:"type,omitempty" jsonschema:"Certificate type. Specifies the certificate format (e.g., PEM, JWK). Used for certificate-based client authentication or JWT signature validation."`
	Value string               `json:"value,omitempty" yaml:"value,omitempty" jsonschema:"Certificate value. The actual certificate content in the format specified by type. For PEM: base64-encoded certificate. For JWK: JSON Web Key."`
}

// ApplicationRequest represents the request structure for creating or updating an application.
//
//nolint:lll
type ApplicationRequest struct {
	Name                      string                      `json:"name" yaml:"name"`
	Description               string                      `json:"description" yaml:"description"`
	AuthFlowID                string                      `json:"auth_flow_id,omitempty" yaml:"auth_flow_id,omitempty"`
	RegistrationFlowID        string                      `json:"registration_flow_id,omitempty" yaml:"registration_flow_id,omitempty"`
	IsRegistrationFlowEnabled bool                        `json:"is_registration_flow_enabled" yaml:"is_registration_flow_enabled"`
	ThemeID                   string                      `json:"theme_id,omitempty" yaml:"theme_id,omitempty"`
	LayoutID                  string                      `json:"layout_id,omitempty" yaml:"layout_id,omitempty"`
	Template                  string                      `json:"template,omitempty" yaml:"template,omitempty"`
	URL                       string                      `json:"url,omitempty" yaml:"url,omitempty"`
	LogoURL                   string                      `json:"logo_url,omitempty" yaml:"logo_url,omitempty"`
	Assertion                 *AssertionConfig            `json:"assertion,omitempty" yaml:"assertion,omitempty"`
	Certificate               *ApplicationCertificate     `json:"certificate,omitempty" yaml:"certificate,omitempty"`
	TosURI                    string                      `json:"tos_uri,omitempty" yaml:"tos_uri,omitempty"`
	PolicyURI                 string                      `json:"policy_uri,omitempty" yaml:"policy_uri,omitempty"`
	Contacts                  []string                    `json:"contacts,omitempty" yaml:"contacts,omitempty"`
	InboundAuthConfig         []InboundAuthConfigComplete `json:"inbound_auth_config,omitempty" yaml:"inbound_auth_config,omitempty"`
	AllowedUserTypes          []string                    `json:"allowed_user_types,omitempty" yaml:"allowed_user_types,omitempty"`
}

// ApplicationRequestWithID represents the request structure for importing an application using file based runtime.
//
//nolint:lll
type ApplicationRequestWithID struct {
	ID                        string                      `json:"id" yaml:"id"`
	Name                      string                      `json:"name" yaml:"name"`
	Description               string                      `json:"description" yaml:"description"`
	AuthFlowID                string                      `json:"auth_flow_id,omitempty" yaml:"auth_flow_id,omitempty"`
	RegistrationFlowID        string                      `json:"registration_flow_id,omitempty" yaml:"registration_flow_id,omitempty"`
	IsRegistrationFlowEnabled bool                        `json:"is_registration_flow_enabled" yaml:"is_registration_flow_enabled"`
	ThemeID                   string                      `json:"theme_id,omitempty" yaml:"theme_id,omitempty"`
	LayoutID                  string                      `json:"layout_id,omitempty" yaml:"layout_id,omitempty"`
	Template                  string                      `json:"template,omitempty" yaml:"template,omitempty"`
	URL                       string                      `json:"url,omitempty" yaml:"url,omitempty"`
	LogoURL                   string                      `json:"logo_url,omitempty" yaml:"logo_url,omitempty"`
	Assertion                 *AssertionConfig            `json:"assertion,omitempty" yaml:"assertion,omitempty"`
	Certificate               *ApplicationCertificate     `json:"certificate,omitempty" yaml:"certificate,omitempty"`
	TosURI                    string                      `json:"tos_uri,omitempty" yaml:"tos_uri,omitempty"`
	PolicyURI                 string                      `json:"policy_uri,omitempty" yaml:"policy_uri,omitempty"`
	Contacts                  []string                    `json:"contacts,omitempty" yaml:"contacts,omitempty"`
	InboundAuthConfig         []InboundAuthConfigComplete `json:"inbound_auth_config,omitempty" yaml:"inbound_auth_config,omitempty"`
	AllowedUserTypes          []string                    `json:"allowed_user_types,omitempty" yaml:"allowed_user_types,omitempty"`
}

// ApplicationCompleteResponse represents the complete response structure for an application.
type ApplicationCompleteResponse struct {
	ID                        string                      `json:"id,omitempty"`
	Name                      string                      `json:"name"`
	Description               string                      `json:"description,omitempty"`
	ClientID                  string                      `json:"client_id,omitempty"`
	AuthFlowID                string                      `json:"auth_flow_id,omitempty"`
	RegistrationFlowID        string                      `json:"registration_flow_id,omitempty"`
	IsRegistrationFlowEnabled bool                        `json:"is_registration_flow_enabled"`
	ThemeID                   string                      `json:"theme_id,omitempty"`
	LayoutID                  string                      `json:"layout_id,omitempty"`
	Template                  string                      `json:"template,omitempty"`
	URL                       string                      `json:"url,omitempty"`
	LogoURL                   string                      `json:"logo_url,omitempty"`
	Assertion                 *AssertionConfig            `json:"assertion,omitempty"`
	Certificate               *ApplicationCertificate     `json:"certificate,omitempty"`
	TosURI                    string                      `json:"tos_uri,omitempty"`
	PolicyURI                 string                      `json:"policy_uri,omitempty"`
	Contacts                  []string                    `json:"contacts,omitempty"`
	InboundAuthConfig         []InboundAuthConfigComplete `json:"inbound_auth_config,omitempty"`
	AllowedUserTypes          []string                    `json:"allowed_user_types,omitempty"`
}

// ApplicationGetResponse represents the response structure for getting an application.
type ApplicationGetResponse struct {
	ID                        string                  `json:"id,omitempty"`
	Name                      string                  `json:"name"`
	Description               string                  `json:"description,omitempty"`
	ClientID                  string                  `json:"client_id,omitempty"`
	AuthFlowID                string                  `json:"auth_flow_id,omitempty"`
	RegistrationFlowID        string                  `json:"registration_flow_id,omitempty"`
	IsRegistrationFlowEnabled bool                    `json:"is_registration_flow_enabled"`
	ThemeID                   string                  `json:"theme_id,omitempty"`
	LayoutID                  string                  `json:"layout_id,omitempty"`
	Template                  string                  `json:"template,omitempty"`
	URL                       string                  `json:"url,omitempty"`
	LogoURL                   string                  `json:"logo_url,omitempty"`
	Assertion                 *AssertionConfig        `json:"assertion,omitempty"`
	Certificate               *ApplicationCertificate `json:"certificate,omitempty"`
	TosURI                    string                  `json:"tos_uri,omitempty"`
	PolicyURI                 string                  `json:"policy_uri,omitempty"`
	Contacts                  []string                `json:"contacts,omitempty"`
	InboundAuthConfig         []InboundAuthConfig     `json:"inbound_auth_config,omitempty"`
	AllowedUserTypes          []string                `json:"allowed_user_types,omitempty"`
}

// BasicApplicationResponse represents a simplified response structure for an application.
type BasicApplicationResponse struct {
	ID                        string `json:"id,omitempty" jsonschema:"Application ID."`
	Name                      string `json:"name" jsonschema:"Application name."`
	Description               string `json:"description,omitempty" jsonschema:"Application description."`
	ClientID                  string `json:"client_id,omitempty" jsonschema:"OAuth Client ID."`
	LogoURL                   string `json:"logo_url,omitempty" jsonschema:"Logo URL."`
	AuthFlowID                string `json:"auth_flow_id,omitempty" jsonschema:"Authentication Flow ID."`
	RegistrationFlowID        string `json:"registration_flow_id,omitempty" jsonschema:"Registration Flow ID."`
	IsRegistrationFlowEnabled bool   `json:"is_registration_flow_enabled" jsonschema:"Registration enabled status."`
	ThemeID                   string `json:"theme_id,omitempty" jsonschema:"Theme ID."`
	LayoutID                  string `json:"layout_id,omitempty" jsonschema:"Layout ID."`
	Template                  string `json:"template,omitempty" jsonschema:"Application Template."`
}

// ApplicationListResponse represents the response structure for listing applications.
type ApplicationListResponse struct {
	TotalResults int                        `json:"totalResults"`
	Count        int                        `json:"count"`
	Applications []BasicApplicationResponse `json:"applications"`
}

// InboundAuthConfig represents the structure for inbound authentication configuration.
type InboundAuthConfig struct {
	Type           InboundAuthType `json:"type"`
	OAuthAppConfig *OAuthAppConfig `json:"config,omitempty"`
}

// InboundAuthConfigComplete represents the complete structure for inbound authentication configuration.
type InboundAuthConfigComplete struct {
	Type           InboundAuthType         `json:"type" yaml:"type"`
	OAuthAppConfig *OAuthAppConfigComplete `json:"config,omitempty" yaml:"config,omitempty"`
}
