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
package model

import (
	"github.com/asgardeo/thunder/internal/cert"
)

// TokenConfig represents the token configuration structure.
type TokenConfig struct {
	Issuer         string   `json:"issuer"`
	ValidityPeriod int64    `json:"validity_period"`
	UserAttributes []string `json:"user_attributes"`
}

// IDTokenConfig represents the ID token configuration structure.
type IDTokenConfig struct {
	ValidityPeriod int64               `json:"validity_period"`
	UserAttributes []string            `json:"user_attributes"`
	ScopeClaims    map[string][]string `json:"scope_claims,omitempty"`
}

// OAuthTokenConfig represents the OAuth token configuration structure with access_token and id_token wrappers.
type OAuthTokenConfig struct {
	Issuer      string         `json:"issuer,omitempty"`
	AccessToken *TokenConfig   `json:"access_token,omitempty"`
	IDToken     *IDTokenConfig `json:"id_token,omitempty"`
}

// ApplicationDTO represents the data transfer object for application service operations.
type ApplicationDTO struct {
	ID                        string
	Name                      string
	Description               string
	AuthFlowGraphID           string
	RegistrationFlowGraphID   string
	IsRegistrationFlowEnabled bool

	URL       string
	LogoURL   string
	TosURI    string
	PolicyURI string
	Contacts  []string

	Token             *TokenConfig
	Certificate       *ApplicationCertificate
	InboundAuthConfig []InboundAuthConfigDTO
}

// BasicApplicationDTO represents a simplified data transfer object for application service operations.
type BasicApplicationDTO struct {
	ID                        string
	Name                      string
	Description               string
	AuthFlowGraphID           string
	RegistrationFlowGraphID   string
	IsRegistrationFlowEnabled bool
	ClientID                  string
}

// ApplicationProcessedDTO represents the processed data transfer object for application service operations.
type ApplicationProcessedDTO struct {
	ID                        string
	Name                      string
	Description               string
	AuthFlowGraphID           string
	RegistrationFlowGraphID   string
	IsRegistrationFlowEnabled bool

	URL       string
	LogoURL   string
	TosURI    string
	PolicyURI string
	Contacts  []string

	Token             *TokenConfig
	Certificate       *ApplicationCertificate
	InboundAuthConfig []InboundAuthConfigProcessedDTO
}

// InboundAuthConfigDTO represents the data transfer object for inbound authentication configuration.
// TODO: Need to refactor when supporting other/multiple inbound auth types.
type InboundAuthConfigDTO struct {
	Type           InboundAuthType    `json:"type"`
	OAuthAppConfig *OAuthAppConfigDTO `json:"oauth_app_config,omitempty"`
}

// InboundAuthConfigProcessedDTO represents the processed data transfer object for inbound authentication
// configuration.
type InboundAuthConfigProcessedDTO struct {
	Type           InboundAuthType             `json:"type"`
	OAuthAppConfig *OAuthAppConfigProcessedDTO `json:"oauth_app_config,omitempty"`
}

// ApplicationCertificate represents the certificate structure in the application request response.
type ApplicationCertificate struct {
	Type  cert.CertificateType `json:"type"`
	Value string               `json:"value"`
}

// ApplicationRequest represents the request structure for creating or updating an application.
//
//nolint:lll
type ApplicationRequest struct {
	Name                      string                      `json:"name" yaml:"name"`
	Description               string                      `json:"description" yaml:"description"`
	AuthFlowGraphID           string                      `json:"auth_flow_graph_id,omitempty" yaml:"auth_flow_graph_id,omitempty"`
	RegistrationFlowGraphID   string                      `json:"registration_flow_graph_id,omitempty" yaml:"registration_flow_graph_id,omitempty"`
	IsRegistrationFlowEnabled bool                        `json:"is_registration_flow_enabled" yaml:"is_registration_flow_enabled"`
	URL                       string                      `json:"url,omitempty" yaml:"url,omitempty"`
	LogoURL                   string                      `json:"logo_url,omitempty" yaml:"logo_url,omitempty"`
	Token                     *TokenConfig                `json:"token,omitempty" yaml:"token,omitempty"`
	Certificate               *ApplicationCertificate     `json:"certificate,omitempty" yaml:"certificate,omitempty"`
	TosURI                    string                      `json:"tos_uri,omitempty" yaml:"tos_uri,omitempty"`
	PolicyURI                 string                      `json:"policy_uri,omitempty" yaml:"policy_uri,omitempty"`
	Contacts                  []string                    `json:"contacts,omitempty" yaml:"contacts,omitempty"`
	InboundAuthConfig         []InboundAuthConfigComplete `json:"inbound_auth_config,omitempty" yaml:"inbound_auth_config,omitempty"`
}

// ApplicationCompleteResponse represents the complete response structure for an application.
type ApplicationCompleteResponse struct {
	ID                        string                      `json:"id,omitempty"`
	Name                      string                      `json:"name"`
	Description               string                      `json:"description,omitempty"`
	ClientID                  string                      `json:"client_id,omitempty"`
	AuthFlowGraphID           string                      `json:"auth_flow_graph_id,omitempty"`
	RegistrationFlowGraphID   string                      `json:"registration_flow_graph_id,omitempty"`
	IsRegistrationFlowEnabled bool                        `json:"is_registration_flow_enabled"`
	URL                       string                      `json:"url,omitempty"`
	LogoURL                   string                      `json:"logo_url,omitempty"`
	Token                     *TokenConfig                `json:"token,omitempty"`
	Certificate               *ApplicationCertificate     `json:"certificate,omitempty"`
	TosURI                    string                      `json:"tos_uri,omitempty"`
	PolicyURI                 string                      `json:"policy_uri,omitempty"`
	Contacts                  []string                    `json:"contacts,omitempty"`
	InboundAuthConfig         []InboundAuthConfigComplete `json:"inbound_auth_config,omitempty"`
}

// ApplicationGetResponse represents the response structure for getting an application.
type ApplicationGetResponse struct {
	ID                        string                  `json:"id,omitempty"`
	Name                      string                  `json:"name"`
	Description               string                  `json:"description,omitempty"`
	ClientID                  string                  `json:"client_id,omitempty"`
	AuthFlowGraphID           string                  `json:"auth_flow_graph_id,omitempty"`
	RegistrationFlowGraphID   string                  `json:"registration_flow_graph_id,omitempty"`
	IsRegistrationFlowEnabled bool                    `json:"is_registration_flow_enabled"`
	URL                       string                  `json:"url,omitempty"`
	LogoURL                   string                  `json:"logo_url,omitempty"`
	Token                     *TokenConfig            `json:"token,omitempty"`
	Certificate               *ApplicationCertificate `json:"certificate,omitempty"`
	TosURI                    string                  `json:"tos_uri,omitempty"`
	PolicyURI                 string                  `json:"policy_uri,omitempty"`
	Contacts                  []string                `json:"contacts,omitempty"`
	InboundAuthConfig         []InboundAuthConfig     `json:"inbound_auth_config,omitempty"`
}

// BasicApplicationResponse represents a simplified response structure for an application.
type BasicApplicationResponse struct {
	ID                        string `json:"id,omitempty"`
	Name                      string `json:"name"`
	Description               string `json:"description,omitempty"`
	ClientID                  string `json:"client_id,omitempty"`
	AuthFlowGraphID           string `json:"auth_flow_graph_id,omitempty"`
	RegistrationFlowGraphID   string `json:"registration_flow_graph_id,omitempty"`
	IsRegistrationFlowEnabled bool   `json:"is_registration_flow_enabled"`
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
	Type           InboundAuthType         `json:"type"`
	OAuthAppConfig *OAuthAppConfigComplete `json:"config,omitempty" yaml:"config,omitempty"`
}
