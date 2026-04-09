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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/asgardeo/thunder/internal/cert"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

// AssertionConfig represents the assertion configuration structure for application-level (root) assertion configs.
type AssertionConfig struct {
	ValidityPeriod int64    `json:"validityPeriod,omitempty" yaml:"validity_period,omitempty" jsonschema:"Assertion validity period in seconds."`
	UserAttributes []string `json:"userAttributes,omitempty" yaml:"user_attributes,omitempty" jsonschema:"User attributes to include in the assertion. List of user claim names to embed in the assertion (e.g., email, username, roles)."`
}

// LoginConsentConfig represents the login consent configuration for an application.
type LoginConsentConfig struct {
	ValidityPeriod int64 `json:"validityPeriod" yaml:"validity_period" jsonschema:"Consent validity period in seconds. Default value 0 indicates consent is valid until revoked."`
}

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

	Assertion         *AssertionConfig        `json:"assertion,omitempty" jsonschema:"Assertion configuration. Optional. Customize assertion validity periods and included user attributes."`
	Certificate       *ApplicationCertificate `json:"certificate,omitempty" jsonschema:"Application certificate. Optional. For certificate-based authentication or JWT validation."`
	InboundAuthConfig []InboundAuthConfigDTO  `json:"inboundAuthConfig,omitempty" jsonschema:"OAuth/OIDC authentication configuration. Required for OAuth-enabled applications. Configure OAuth grant types, redirect URIs, and client authentication methods."`
	AllowedUserTypes  []string                `json:"allowedUserTypes,omitempty" jsonschema:"Allowed user types. Optional. Restricts which types of users can register to this application."`
	LoginConsent      *LoginConsentConfig     `json:"loginConsent,omitempty" jsonschema:"Login consent configuration settings."`
	Metadata          map[string]interface{}  `json:"metadata,omitempty" jsonschema:"Generic metadata. Optional arbitrary key-value pairs for consumer use."`

	// Localized variant maps keyed by normalised BCP 47 tag (e.g. "fr", "en-us").
	LocalisedClientName map[string]string `json:"client_name_localized,omitempty"`
	LocalisedLogoURL    map[string]string `json:"logo_uri_localized,omitempty"`
	LocalisedTosURI     map[string]string `json:"tos_uri_localized,omitempty"`
	LocalisedPolicyURI  map[string]string `json:"policy_uri_localized,omitempty"`
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

	Assertion         *AssertionConfig            `yaml:"assertion,omitempty" json:"assertion,omitempty" jsonschema:"Assertion configuration settings."`
	Certificate       *ApplicationCertificate     `yaml:"certificate,omitempty" json:"certificate,omitempty" jsonschema:"Application certificate settings."`
	InboundAuthConfig []InboundAuthConfigComplete `yaml:"inbound_auth_config,omitempty" json:"inboundAuthConfig,omitempty" jsonschema:"Inbound authentication configuration (OAuth2/OIDC settings)."`
	AllowedUserTypes  []string                    `yaml:"allowed_user_types,omitempty" json:"allowedUserTypes,omitempty" jsonschema:"Allowed user types for registration."`
	LoginConsent      *LoginConsentConfig         `yaml:"login_consent,omitempty" json:"loginConsent,omitempty" jsonschema:"Login consent configuration settings."`
	Metadata          map[string]interface{}      `yaml:"metadata,omitempty" json:"metadata,omitempty" jsonschema:"Generic metadata key-value pairs."`

	LocalisedClientName map[string]string `yaml:"client_name_localized,omitempty" json:"client_name_localized,omitempty"`
	LocalisedLogoURL    map[string]string `yaml:"logo_uri_localized,omitempty" json:"logo_uri_localized,omitempty"`
	LocalisedTosURI     map[string]string `yaml:"tos_uri_localized,omitempty" json:"tos_uri_localized,omitempty"`
	LocalisedPolicyURI  map[string]string `yaml:"policy_uri_localized,omitempty" json:"policy_uri_localized,omitempty"`
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

	Assertion         *AssertionConfig                `yaml:"assertion,omitempty"`
	Certificate       *ApplicationCertificate         `yaml:"certificate,omitempty"`
	InboundAuthConfig []InboundAuthConfigProcessedDTO `yaml:"inbound_auth_config,omitempty"`
	AllowedUserTypes  []string                        `yaml:"allowed_user_types,omitempty"`
	LoginConsent      *LoginConsentConfig             `yaml:"login_consent,omitempty"`
	Metadata          map[string]interface{}          `yaml:"metadata,omitempty"`

	LocalisedClientName map[string]string `yaml:"client_name_localized,omitempty"`
	LocalisedLogoURL    map[string]string `yaml:"logo_uri_localized,omitempty"`
	LocalisedTosURI     map[string]string `yaml:"tos_uri_localized,omitempty"`
	LocalisedPolicyURI  map[string]string `yaml:"policy_uri_localized,omitempty"`
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
	OUID                      string                      `json:"ouId,omitempty" yaml:"ou_id,omitempty"`
	Name                      string                      `json:"name" yaml:"name"`
	Description               string                      `json:"description" yaml:"description"`
	AuthFlowID                string                      `json:"authFlowId,omitempty" yaml:"auth_flow_id,omitempty"`
	RegistrationFlowID        string                      `json:"registrationFlowId,omitempty" yaml:"registration_flow_id,omitempty"`
	IsRegistrationFlowEnabled bool                        `json:"isRegistrationFlowEnabled" yaml:"is_registration_flow_enabled"`
	ThemeID                   string                      `json:"themeId,omitempty" yaml:"theme_id,omitempty"`
	LayoutID                  string                      `json:"layoutId,omitempty" yaml:"layout_id,omitempty"`
	Template                  string                      `json:"template,omitempty" yaml:"template,omitempty"`
	URL                       string                      `json:"url,omitempty" yaml:"url,omitempty"`
	LogoURL                   string                      `json:"logoUrl,omitempty" yaml:"logo_url,omitempty"`
	Assertion                 *AssertionConfig            `json:"assertion,omitempty" yaml:"assertion,omitempty"`
	Certificate               *ApplicationCertificate     `json:"certificate,omitempty" yaml:"certificate,omitempty"`
	TosURI                    string                      `json:"tosUri,omitempty" yaml:"tos_uri,omitempty"`
	PolicyURI                 string                      `json:"policyUri,omitempty" yaml:"policy_uri,omitempty"`
	Contacts                  []string                    `json:"contacts,omitempty" yaml:"contacts,omitempty"`
	InboundAuthConfig         []InboundAuthConfigComplete `json:"inboundAuthConfig,omitempty" yaml:"inbound_auth_config,omitempty"`
	AllowedUserTypes          []string                    `json:"allowedUserTypes,omitempty" yaml:"allowed_user_types,omitempty"`
	LoginConsent              *LoginConsentConfig         `json:"loginConsent,omitempty" yaml:"login_consent,omitempty"`
	Metadata                  map[string]interface{}      `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	// Localized variant maps — populated by UnmarshalJSON from field#tag keys; not in standard JSON output.
	LocalisedClientName map[string]string `json:"-"`
	LocalisedLogoURL    map[string]string `json:"-"`
	LocalisedTosURI     map[string]string `json:"-"`
	LocalisedPolicyURI  map[string]string `json:"-"`
}

// appRequestJSON is an alias used by UnmarshalJSON to decode fixed fields without recursion.
type appRequestJSON ApplicationRequest

// localisableAppFields is the set of application API field names that accept language-tagged variants.
var localisableAppFields = map[string]bool{
	"name": true, "logoUrl": true, "tosUri": true, "policyUri": true,
}

// UnmarshalJSON decodes an ApplicationRequest in a single JSON parse of the input bytes.
// It separates language-tagged keys (field#tag convention, OIDC DCR §2) into the localized
// variant maps, then decodes the remaining standard keys into the typed struct.
// Localisable fields with a non-string tagged value are rejected; tagged variants on
// non-localisable fields are silently ignored (AC-12).
func (r *ApplicationRequest) UnmarshalJSON(data []byte) error {
	// Single parse into a raw key-value map.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Separate tagged (#) keys from standard keys.
	clean := make(map[string]json.RawMessage, len(raw))
	for key, val := range raw {
		field, tag, ok := strings.Cut(key, "#")
		if !ok {
			clean[key] = val
			continue
		}
		// Multiple '#' in a key (e.g. "name#en#US") are invalid (AC-08).
		// Reject if the base field is localisable; silently skip otherwise.
		if strings.Contains(tag, "#") {
			if localisableAppFields[field] {
				return fmt.Errorf("invalid localized field key %q: tag must not contain '#'", key)
			}
			continue
		}
		var s string
		if err := json.Unmarshal(val, &s); err != nil {
			// Non-string value on a recognized localisable field is a client error.
			if localisableAppFields[field] {
				return err
			}
			continue
		}
		normTag := sysutils.NormaliseBCP47Tag(tag)
		switch field {
		case "name":
			if r.LocalisedClientName == nil {
				r.LocalisedClientName = make(map[string]string)
			}
			r.LocalisedClientName[normTag] = s
		case "logoUrl":
			if r.LocalisedLogoURL == nil {
				r.LocalisedLogoURL = make(map[string]string)
			}
			r.LocalisedLogoURL[normTag] = s
		case "tosUri":
			if r.LocalisedTosURI == nil {
				r.LocalisedTosURI = make(map[string]string)
			}
			r.LocalisedTosURI[normTag] = s
		case "policyUri":
			if r.LocalisedPolicyURI == nil {
				r.LocalisedPolicyURI = make(map[string]string)
			}
			r.LocalisedPolicyURI[normTag] = s
			// All other tagged fields are silently ignored (AC-12).
		}
	}

	// Decode standard (non-tagged) keys using the alias to avoid infinite recursion.
	cleanBytes, err := json.Marshal(clean)
	if err != nil {
		return err
	}
	return json.Unmarshal(cleanBytes, (*appRequestJSON)(r))
}

// ApplicationRequestWithID represents the request structure for importing an application using file based runtime.
//
//nolint:lll
type ApplicationRequestWithID struct {
	ID                        string                      `json:"id" yaml:"id"`
	OUID                      string                      `json:"ouId,omitempty" yaml:"ou_id,omitempty"`
	Name                      string                      `json:"name" yaml:"name"`
	Description               string                      `json:"description" yaml:"description"`
	AuthFlowID                string                      `json:"authFlowId,omitempty" yaml:"auth_flow_id,omitempty"`
	RegistrationFlowID        string                      `json:"registrationFlowId,omitempty" yaml:"registration_flow_id,omitempty"`
	IsRegistrationFlowEnabled bool                        `json:"isRegistrationFlowEnabled" yaml:"is_registration_flow_enabled"`
	ThemeID                   string                      `json:"themeId,omitempty" yaml:"theme_id,omitempty"`
	LayoutID                  string                      `json:"layoutId,omitempty" yaml:"layout_id,omitempty"`
	Template                  string                      `json:"template,omitempty" yaml:"template,omitempty"`
	URL                       string                      `json:"url,omitempty" yaml:"url,omitempty"`
	LogoURL                   string                      `json:"logoUrl,omitempty" yaml:"logo_url,omitempty"`
	Assertion                 *AssertionConfig            `json:"assertion,omitempty" yaml:"assertion,omitempty"`
	Certificate               *ApplicationCertificate     `json:"certificate,omitempty" yaml:"certificate,omitempty"`
	TosURI                    string                      `json:"tosUri,omitempty" yaml:"tos_uri,omitempty"`
	PolicyURI                 string                      `json:"policyUri,omitempty" yaml:"policy_uri,omitempty"`
	Contacts                  []string                    `json:"contacts,omitempty" yaml:"contacts,omitempty"`
	InboundAuthConfig         []InboundAuthConfigComplete `json:"inboundAuthConfig,omitempty" yaml:"inbound_auth_config,omitempty"`
	AllowedUserTypes          []string                    `json:"allowedUserTypes,omitempty" yaml:"allowed_user_types,omitempty"`
	Metadata                  map[string]interface{}      `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// ApplicationCompleteResponse represents the complete response structure for an application.
type ApplicationCompleteResponse struct {
	ID                        string                      `json:"id,omitempty"`
	OUID                      string                      `json:"ouId,omitempty"`
	Name                      string                      `json:"name"`
	Description               string                      `json:"description,omitempty"`
	ClientID                  string                      `json:"clientId,omitempty"`
	AuthFlowID                string                      `json:"authFlowId,omitempty"`
	RegistrationFlowID        string                      `json:"registrationFlowId,omitempty"`
	IsRegistrationFlowEnabled bool                        `json:"isRegistrationFlowEnabled"`
	ThemeID                   string                      `json:"themeId,omitempty"`
	LayoutID                  string                      `json:"layoutId,omitempty"`
	Template                  string                      `json:"template,omitempty"`
	URL                       string                      `json:"url,omitempty"`
	LogoURL                   string                      `json:"logoUrl,omitempty"`
	Assertion                 *AssertionConfig            `json:"assertion,omitempty"`
	Certificate               *ApplicationCertificate     `json:"certificate,omitempty"`
	TosURI                    string                      `json:"tosUri,omitempty"`
	PolicyURI                 string                      `json:"policyUri,omitempty"`
	Contacts                  []string                    `json:"contacts,omitempty"`
	InboundAuthConfig         []InboundAuthConfigComplete `json:"inboundAuthConfig,omitempty"`
	AllowedUserTypes          []string                    `json:"allowedUserTypes,omitempty"`
	LoginConsent              *LoginConsentConfig         `json:"loginConsent,omitempty"`
	Metadata                  map[string]interface{}      `json:"metadata,omitempty"`

	// Localized variant maps — emitted as field#tag keys by MarshalJSON; not decoded from standard JSON.
	LocalisedClientName map[string]string `json:"-"`
	LocalisedLogoURL    map[string]string `json:"-"`
	LocalisedTosURI     map[string]string `json:"-"`
	LocalisedPolicyURI  map[string]string `json:"-"`
}

// appCompleteResponseJSON is an alias used by MarshalJSON to avoid infinite recursion.
type appCompleteResponseJSON ApplicationCompleteResponse

// MarshalJSON serializes ApplicationCompleteResponse, injecting language-tagged variant keys
// (e.g. "name#fr", "logoUrl#fr") as top-level JSON properties.
func (r ApplicationCompleteResponse) MarshalJSON() ([]byte, error) {
	base, err := json.Marshal(appCompleteResponseJSON(r))
	if err != nil {
		return nil, err
	}
	return injectLocalisedFields(base, r.LocalisedClientName, r.LocalisedLogoURL, r.LocalisedTosURI, r.LocalisedPolicyURI)
}

// ApplicationGetResponse represents the response structure for getting an application.
type ApplicationGetResponse struct {
	ID                        string                  `json:"id,omitempty"`
	OUID                      string                  `json:"ouId,omitempty"`
	Name                      string                  `json:"name"`
	Description               string                  `json:"description,omitempty"`
	ClientID                  string                  `json:"clientId,omitempty"`
	AuthFlowID                string                  `json:"authFlowId,omitempty"`
	RegistrationFlowID        string                  `json:"registrationFlowId,omitempty"`
	IsRegistrationFlowEnabled bool                    `json:"isRegistrationFlowEnabled"`
	ThemeID                   string                  `json:"themeId,omitempty"`
	LayoutID                  string                  `json:"layoutId,omitempty"`
	Template                  string                  `json:"template,omitempty"`
	URL                       string                  `json:"url,omitempty"`
	LogoURL                   string                  `json:"logoUrl,omitempty"`
	Assertion                 *AssertionConfig        `json:"assertion,omitempty"`
	Certificate               *ApplicationCertificate `json:"certificate,omitempty"`
	TosURI                    string                  `json:"tosUri,omitempty"`
	PolicyURI                 string                  `json:"policyUri,omitempty"`
	Contacts                  []string                `json:"contacts,omitempty"`
	InboundAuthConfig         []InboundAuthConfig     `json:"inboundAuthConfig,omitempty"`
	AllowedUserTypes          []string                `json:"allowedUserTypes,omitempty"`
	LoginConsent              *LoginConsentConfig     `json:"loginConsent,omitempty"`
	Metadata                  map[string]interface{}  `json:"metadata,omitempty"`

	// Localized variant maps — emitted as field#tag keys by MarshalJSON; not decoded from standard JSON.
	LocalisedClientName map[string]string `json:"-"`
	LocalisedLogoURL    map[string]string `json:"-"`
	LocalisedTosURI     map[string]string `json:"-"`
	LocalisedPolicyURI  map[string]string `json:"-"`
}

// appGetResponseJSON is an alias used by MarshalJSON to avoid infinite recursion.
type appGetResponseJSON ApplicationGetResponse

// MarshalJSON serializes ApplicationGetResponse, injecting language-tagged variant keys
// (e.g. "name#fr", "logoUrl#fr") as top-level JSON properties.
func (r ApplicationGetResponse) MarshalJSON() ([]byte, error) {
	base, err := json.Marshal(appGetResponseJSON(r))
	if err != nil {
		return nil, err
	}
	return injectLocalisedFields(base, r.LocalisedClientName, r.LocalisedLogoURL, r.LocalisedTosURI, r.LocalisedPolicyURI)
}

// injectLocalisedFields merges language-tagged variant entries into a serialized JSON object.
// The field names use the application endpoint JSON naming convention (name, logoUrl, tosUri, policyUri).
func injectLocalisedFields(base []byte, clientName, logoURL, tosURI, policyURI map[string]string) ([]byte, error) {
	if len(clientName) == 0 && len(logoURL) == 0 && len(tosURI) == 0 && len(policyURI) == 0 {
		return base, nil
	}

	var m map[string]interface{}
	if err := json.Unmarshal(base, &m); err != nil {
		return nil, err
	}

	for tag, val := range clientName {
		m["name#"+tag] = val
	}
	for tag, val := range logoURL {
		m["logoUrl#"+tag] = val
	}
	for tag, val := range tosURI {
		m["tosUri#"+tag] = val
	}
	for tag, val := range policyURI {
		m["policyUri#"+tag] = val
	}

	return json.Marshal(m)
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
