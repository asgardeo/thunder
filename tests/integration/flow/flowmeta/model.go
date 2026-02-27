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

package flowmeta

import "encoding/json"

// FlowMetadataResponse represents the aggregated metadata response from /flow/meta.
type FlowMetadataResponse struct {
	IsRegistrationFlowEnabled bool                 `json:"is_registration_flow_enabled"`
	Application               *ApplicationMetadata `json:"application,omitempty"`
	OU                        *OUMetadata          `json:"ou,omitempty"`
	Design                    DesignMetadata       `json:"design"`
	I18n                      I18nMetadata         `json:"i18n"`
}

// ApplicationMetadata represents application-specific metadata.
type ApplicationMetadata struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	LogoURL   string `json:"logo_url,omitempty"`
	URL       string `json:"url,omitempty"`
	TosURI    string `json:"tos_uri,omitempty"`
	PolicyURI string `json:"policy_uri,omitempty"`
}

// OUMetadata represents organization unit metadata.
type OUMetadata struct {
	ID              string `json:"id,omitempty"`
	Handle          string `json:"handle,omitempty"`
	Name            string `json:"name,omitempty"`
	Description     string `json:"description,omitempty"`
	LogoURL         string `json:"logo_url,omitempty"`
	TosURI          string `json:"tos_uri,omitempty"`
	PolicyURI       string `json:"policy_uri,omitempty"`
	CookiePolicyURI string `json:"cookie_policy_uri,omitempty"`
}

// DesignMetadata represents theme and layout configuration.
type DesignMetadata struct {
	Theme  json.RawMessage `json:"theme"`
	Layout json.RawMessage `json:"layout"`
}

// I18nMetadata represents internationalization data.
type I18nMetadata struct {
	Languages    []string                     `json:"languages"`
	Language     string                       `json:"language"`
	TotalResults int                          `json:"totalResults"`
	Translations map[string]map[string]string `json:"translations"`
}

// ErrorResponse represents an error response from the API.
type ErrorResponse struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Description string `json:"description,omitempty"`
}
