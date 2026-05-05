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
)

// ApplicationDTO represents the data transfer object for application service operations.
type ApplicationDTO struct {
	ID                        string `json:"id,omitempty" jsonschema:"Application ID. Auto-generated unique identifier."`
	OUID                      string `json:"ouId,omitempty" jsonschema:"Organization unit ID. The OU this application belongs to."`
	Name                      string `json:"name" jsonschema:"Application name."`
	Description               string `json:"description,omitempty" jsonschema:"Optional description of the application's purpose or functionality."`
	AuthFlowID                string `json:"authFlowId,omitempty" jsonschema:"Authentication flow ID. Optional. Specifies which login flow to use (e.g., MFA, passwordless). Use list_flows to find available flows. If omitted, the default authentication flow is used."`
	RegistrationFlowID        string `json:"registrationFlowId,omitempty" jsonschema:"Registration flow ID. Optional. Specifies the user registration/signup flow. Use list_flows to find available flows."`
	IsRegistrationFlowEnabled bool   `json:"isRegistrationFlowEnabled,omitempty" jsonschema:"Enable self-service registration. Set to true to allow users to sign up themselves. Requires registration_flow_id to be set."`
	ThemeID                   string `json:"themeId,omitempty" jsonschema:"Theme configuration ID. Optional. Customizes the visual styling (colors, typography) of login pages."`
	LayoutID                  string `json:"layoutId,omitempty" jsonschema:"Layout configuration ID. Optional. Customizes the screen structure and component positioning of login pages."`
	Template                  string `json:"template,omitempty" jsonschema:"Application template. Optional. Pre-configured application type template."`

	URL       string   `json:"url,omitempty" jsonschema:"Application home URL. Optional. The main URL where your application is hosted."`
	LogoURL   string   `json:"logoUrl,omitempty" jsonschema:"Logo image URL. Optional. Displayed in login pages and application listings."`
	TosURI    string   `json:"tosUri,omitempty" jsonschema:"Terms of Service URI. Optional. Link to your application's terms of service."`
	PolicyURI string   `json:"policyUri,omitempty" jsonschema:"Privacy Policy URI. Optional. Link to your application's privacy policy."`
	Contacts  []string `json:"contacts,omitempty" jsonschema:"Contact email addresses. Optional. Administrative contact emails for this application."`

	Assertion         *inboundmodel.AssertionConfig    `json:"assertion,omitempty" jsonschema:"Assertion configuration. Optional. Customize assertion validity periods and included user attributes."`
	Certificate       *inboundmodel.Certificate        `json:"certificate,omitempty" jsonschema:"Application certificate. Optional. For certificate-based authentication or JWT validation."`
	InboundAuthConfig []InboundAuthConfigDTO           `json:"inboundAuthConfig,omitempty" jsonschema:"OAuth/OIDC authentication configuration. Required for OAuth-enabled applications. Configure OAuth grant types, redirect URIs, and client authentication methods."`
	AllowedUserTypes  []string                         `json:"allowedUserTypes,omitempty" jsonschema:"Allowed user types. Optional. Restricts which types of users can register to this application."`
	LoginConsent      *inboundmodel.LoginConsentConfig `json:"loginConsent,omitempty" jsonschema:"Login consent configuration settings."`
	Metadata          map[string]interface{}           `json:"metadata,omitempty" jsonschema:"Generic metadata. Optional arbitrary key-value pairs for consumer use."`
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
	IsReadOnly                bool
}

// Application represents the structure for application which returns in GetApplicationById.
type Application struct {
	ID                        string `yaml:"id,omitempty" json:"id,omitempty" jsonschema:"Application ID. Auto-generated unique identifier."`
	OUID                      string `yaml:"ou_id,omitempty" json:"ouId,omitempty" jsonschema:"Organization unit ID. The OU this application belongs to."`
	Name                      string `yaml:"name,omitempty" json:"name,omitempty" jsonschema:"Application name."`
	Description               string `yaml:"description,omitempty" json:"description,omitempty" jsonschema:"Optional description of the application's purpose."`
	AuthFlowID                string `yaml:"auth_flow_id,omitempty" json:"authFlowId,omitempty" jsonschema:"Associated authentication flow ID."`
	RegistrationFlowID        string `yaml:"registration_flow_id,omitempty" json:"registrationFlowId,omitempty" jsonschema:"Associated registration flow ID."`
	IsRegistrationFlowEnabled bool   `yaml:"is_registration_flow_enabled,omitempty" json:"isRegistrationFlowEnabled,omitempty" jsonschema:"Indicates if self-service registration is enabled."`
	ThemeID                   string `yaml:"theme_id,omitempty" json:"themeId,omitempty" jsonschema:"Associated theme configuration ID."`
	LayoutID                  string `yaml:"layout_id,omitempty" json:"layoutId,omitempty" jsonschema:"Associated layout configuration ID."`
	Template                  string `yaml:"template,omitempty" json:"template,omitempty" jsonschema:"Template used to create the application."`

	URL       string   `yaml:"url,omitempty" json:"url,omitempty" jsonschema:"Application home URL."`
	LogoURL   string   `yaml:"logo_url,omitempty" json:"logoUrl,omitempty" jsonschema:"Application logo URL."`
	TosURI    string   `yaml:"tos_uri,omitempty" json:"tosUri,omitempty" jsonschema:"Terms of Service URI."`
	PolicyURI string   `yaml:"policy_uri,omitempty" json:"policyUri,omitempty" jsonschema:"Privacy Policy URI."`
	Contacts  []string `yaml:"contacts,omitempty" json:"contacts,omitempty"`

	Assertion         *inboundmodel.AssertionConfig    `yaml:"assertion,omitempty" json:"assertion,omitempty" jsonschema:"Assertion configuration settings."`
	Certificate       *inboundmodel.Certificate        `yaml:"certificate,omitempty" json:"certificate,omitempty" jsonschema:"Application certificate settings."`
	InboundAuthConfig []InboundAuthConfigComplete      `yaml:"inbound_auth_config,omitempty" json:"inboundAuthConfig,omitempty" jsonschema:"Inbound authentication configuration (OAuth2/OIDC settings)."`
	AllowedUserTypes  []string                         `yaml:"allowed_user_types,omitempty" json:"allowedUserTypes,omitempty" jsonschema:"Allowed user types for registration."`
	LoginConsent      *inboundmodel.LoginConsentConfig `yaml:"login_consent,omitempty" json:"loginConsent,omitempty" jsonschema:"Login consent configuration settings."`
	Metadata          map[string]interface{}           `yaml:"metadata,omitempty" json:"metadata,omitempty" jsonschema:"Generic metadata key-value pairs."`
}

// ApplicationProcessedDTO represents the processed data transfer object for application service operations.
type ApplicationProcessedDTO struct {
	ID                        string `yaml:"id,omitempty"`
	OUID                      string `yaml:"ou_id,omitempty"`
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

	Assertion         *inboundmodel.AssertionConfig    `yaml:"assertion,omitempty"`
	Certificate       *inboundmodel.Certificate        `yaml:"certificate,omitempty"`
	InboundAuthConfig []InboundAuthConfigProcessedDTO  `yaml:"inbound_auth_config,omitempty"`
	AllowedUserTypes  []string                         `yaml:"allowed_user_types,omitempty"`
	LoginConsent      *inboundmodel.LoginConsentConfig `yaml:"login_consent,omitempty"`
	Metadata          map[string]interface{}           `yaml:"metadata,omitempty"`
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
	Type           InboundAuthType           `json:"type" yaml:"type,omitempty"`
	OAuthAppConfig *inboundmodel.OAuthClient `json:"config,omitempty" yaml:"config,omitempty"`
}

// ApplicationCertificate is an alias for the canonical inboundclient type.
type ApplicationCertificate = inboundmodel.Certificate

// ApplicationRequest represents the request structure for creating or updating an application.
//
//nolint:lll
type ApplicationRequest struct {
	OUID                      string                           `json:"ouId,omitempty" yaml:"ou_id,omitempty"`
	Name                      string                           `json:"name" yaml:"name"`
	Description               string                           `json:"description" yaml:"description"`
	AuthFlowID                string                           `json:"authFlowId,omitempty" yaml:"auth_flow_id,omitempty"`
	RegistrationFlowID        string                           `json:"registrationFlowId,omitempty" yaml:"registration_flow_id,omitempty"`
	IsRegistrationFlowEnabled bool                             `json:"isRegistrationFlowEnabled" yaml:"is_registration_flow_enabled"`
	ThemeID                   string                           `json:"themeId,omitempty" yaml:"theme_id,omitempty"`
	LayoutID                  string                           `json:"layoutId,omitempty" yaml:"layout_id,omitempty"`
	Template                  string                           `json:"template,omitempty" yaml:"template,omitempty"`
	URL                       string                           `json:"url,omitempty" yaml:"url,omitempty"`
	LogoURL                   string                           `json:"logoUrl,omitempty" yaml:"logo_url,omitempty"`
	Assertion                 *inboundmodel.AssertionConfig    `json:"assertion,omitempty" yaml:"assertion,omitempty"`
	Certificate               *inboundmodel.Certificate        `json:"certificate,omitempty" yaml:"certificate,omitempty"`
	TosURI                    string                           `json:"tosUri,omitempty" yaml:"tos_uri,omitempty"`
	PolicyURI                 string                           `json:"policyUri,omitempty" yaml:"policy_uri,omitempty"`
	Contacts                  []string                         `json:"contacts,omitempty" yaml:"contacts,omitempty"`
	InboundAuthConfig         []InboundAuthConfigComplete      `json:"inboundAuthConfig,omitempty" yaml:"inbound_auth_config,omitempty"`
	AllowedUserTypes          []string                         `json:"allowedUserTypes,omitempty" yaml:"allowed_user_types,omitempty"`
	LoginConsent              *inboundmodel.LoginConsentConfig `json:"loginConsent,omitempty" yaml:"login_consent,omitempty"`
	Metadata                  map[string]interface{}           `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// ApplicationRequestWithID represents the request structure for importing an application using file based runtime.
//
//nolint:lll
type ApplicationRequestWithID struct {
	ID                        string                        `json:"id" yaml:"id"`
	OUID                      string                        `json:"ouId,omitempty" yaml:"ou_id,omitempty"`
	Name                      string                        `json:"name" yaml:"name"`
	Description               string                        `json:"description" yaml:"description"`
	AuthFlowID                string                        `json:"authFlowId,omitempty" yaml:"auth_flow_id,omitempty"`
	RegistrationFlowID        string                        `json:"registrationFlowId,omitempty" yaml:"registration_flow_id,omitempty"`
	IsRegistrationFlowEnabled bool                          `json:"isRegistrationFlowEnabled" yaml:"is_registration_flow_enabled"`
	ThemeID                   string                        `json:"themeId,omitempty" yaml:"theme_id,omitempty"`
	LayoutID                  string                        `json:"layoutId,omitempty" yaml:"layout_id,omitempty"`
	Template                  string                        `json:"template,omitempty" yaml:"template,omitempty"`
	URL                       string                        `json:"url,omitempty" yaml:"url,omitempty"`
	LogoURL                   string                        `json:"logoUrl,omitempty" yaml:"logo_url,omitempty"`
	Assertion                 *inboundmodel.AssertionConfig `json:"assertion,omitempty" yaml:"assertion,omitempty"`
	Certificate               *inboundmodel.Certificate     `json:"certificate,omitempty" yaml:"certificate,omitempty"`
	TosURI                    string                        `json:"tosUri,omitempty" yaml:"tos_uri,omitempty"`
	PolicyURI                 string                        `json:"policyUri,omitempty" yaml:"policy_uri,omitempty"`
	Contacts                  []string                      `json:"contacts,omitempty" yaml:"contacts,omitempty"`
	InboundAuthConfig         []InboundAuthConfigComplete   `json:"inboundAuthConfig,omitempty" yaml:"inbound_auth_config,omitempty"`
	AllowedUserTypes          []string                      `json:"allowedUserTypes,omitempty" yaml:"allowed_user_types,omitempty"`
	Metadata                  map[string]interface{}        `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// ApplicationCompleteResponse represents the complete response structure for an application.
type ApplicationCompleteResponse struct {
	ID                        string                           `json:"id,omitempty"`
	OUID                      string                           `json:"ouId,omitempty"`
	Name                      string                           `json:"name"`
	Description               string                           `json:"description,omitempty"`
	ClientID                  string                           `json:"clientId,omitempty"`
	AuthFlowID                string                           `json:"authFlowId,omitempty"`
	RegistrationFlowID        string                           `json:"registrationFlowId,omitempty"`
	IsRegistrationFlowEnabled bool                             `json:"isRegistrationFlowEnabled"`
	ThemeID                   string                           `json:"themeId,omitempty"`
	LayoutID                  string                           `json:"layoutId,omitempty"`
	Template                  string                           `json:"template,omitempty"`
	URL                       string                           `json:"url,omitempty"`
	LogoURL                   string                           `json:"logoUrl,omitempty"`
	Assertion                 *inboundmodel.AssertionConfig    `json:"assertion,omitempty"`
	Certificate               *inboundmodel.Certificate        `json:"certificate,omitempty"`
	TosURI                    string                           `json:"tosUri,omitempty"`
	PolicyURI                 string                           `json:"policyUri,omitempty"`
	Contacts                  []string                         `json:"contacts,omitempty"`
	InboundAuthConfig         []InboundAuthConfigComplete      `json:"inboundAuthConfig,omitempty"`
	AllowedUserTypes          []string                         `json:"allowedUserTypes,omitempty"`
	LoginConsent              *inboundmodel.LoginConsentConfig `json:"loginConsent,omitempty"`
	Metadata                  map[string]interface{}           `json:"metadata,omitempty"`
}

// ApplicationGetResponse represents the response structure for getting an application.
type ApplicationGetResponse struct {
	ID                        string                           `json:"id,omitempty"`
	OUID                      string                           `json:"ouId,omitempty"`
	Name                      string                           `json:"name"`
	Description               string                           `json:"description,omitempty"`
	ClientID                  string                           `json:"clientId,omitempty"`
	AuthFlowID                string                           `json:"authFlowId,omitempty"`
	RegistrationFlowID        string                           `json:"registrationFlowId,omitempty"`
	IsRegistrationFlowEnabled bool                             `json:"isRegistrationFlowEnabled"`
	ThemeID                   string                           `json:"themeId,omitempty"`
	LayoutID                  string                           `json:"layoutId,omitempty"`
	Template                  string                           `json:"template,omitempty"`
	URL                       string                           `json:"url,omitempty"`
	LogoURL                   string                           `json:"logoUrl,omitempty"`
	Assertion                 *inboundmodel.AssertionConfig    `json:"assertion,omitempty"`
	Certificate               *inboundmodel.Certificate        `json:"certificate,omitempty"`
	TosURI                    string                           `json:"tosUri,omitempty"`
	PolicyURI                 string                           `json:"policyUri,omitempty"`
	Contacts                  []string                         `json:"contacts,omitempty"`
	InboundAuthConfig         []InboundAuthConfig              `json:"inboundAuthConfig,omitempty"`
	AllowedUserTypes          []string                         `json:"allowedUserTypes,omitempty"`
	LoginConsent              *inboundmodel.LoginConsentConfig `json:"loginConsent,omitempty"`
	Metadata                  map[string]interface{}           `json:"metadata,omitempty"`
}

// BasicApplicationResponse represents a simplified response structure for an application.
type BasicApplicationResponse struct {
	ID                        string `json:"id,omitempty" jsonschema:"Application ID."`
	Name                      string `json:"name" jsonschema:"Application name."`
	Description               string `json:"description,omitempty" jsonschema:"Application description."`
	ClientID                  string `json:"clientId,omitempty" jsonschema:"OAuth Client ID."`
	LogoURL                   string `json:"logoUrl,omitempty" jsonschema:"Logo URL."`
	AuthFlowID                string `json:"authFlowId,omitempty" jsonschema:"Authentication Flow ID."`
	RegistrationFlowID        string `json:"registrationFlowId,omitempty" jsonschema:"Registration Flow ID."`
	IsRegistrationFlowEnabled bool   `json:"isRegistrationFlowEnabled" jsonschema:"Registration enabled status."`
	ThemeID                   string `json:"themeId,omitempty" jsonschema:"Theme ID."`
	LayoutID                  string `json:"layoutId,omitempty" jsonschema:"Layout ID."`
	Template                  string `json:"template,omitempty" jsonschema:"Application Template."`
	IsReadOnly                bool   `json:"isReadOnly" jsonschema:"Indicates if the application is read-only (declarative/immutable)."`
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
