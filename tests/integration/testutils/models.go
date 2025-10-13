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

package testutils

import (
	"encoding/json"
)

// User represents a user in the system
type User struct {
	ID               string          `json:"id"`
	OrganizationUnit string          `json:"organizationUnit"`
	Type             string          `json:"type"`
	Attributes       json.RawMessage `json:"attributes"`
}

// Application represents an application in the system
type Application struct {
	ID                        string                   `json:"id,omitempty"`
	Name                      string                   `json:"name"`
	Description               string                   `json:"description"`
	IsRegistrationFlowEnabled bool                     `json:"is_registration_flow_enabled"`
	AuthFlowGraphID           string                   `json:"auth_flow_graph_id,omitempty"`
	RegistrationFlowGraphID   string                   `json:"registration_flow_graph_id,omitempty"`
	ClientID                  string                   `json:"client_id,omitempty"`
	ClientSecret              string                   `json:"client_secret,omitempty"`
	RedirectURIs              []string                 `json:"redirect_uris,omitempty"`
	Certificate               map[string]interface{}   `json:"certificate,omitempty"`
	InboundAuthConfig         []map[string]interface{} `json:"inbound_auth_config,omitempty"`
}

// OrganizationUnit represents an organization unit in the system
type OrganizationUnit struct {
	ID          string  `json:"id,omitempty"`
	Handle      string  `json:"handle"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Parent      *string `json:"parent,omitempty"`
}

// IDPProperty represents a property of an identity provider
type IDPProperty struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	IsSecret bool   `json:"is_secret"`
}

// IDP represents an identity provider in the system
type IDP struct {
	ID          string        `json:"id,omitempty"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Type        string        `json:"type"`
	Properties  []IDPProperty `json:"properties"`
}

// Link represents a pagination link.
type Link struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
}

// UserListResponse represents the paginated response for user listing
type UserListResponse struct {
	TotalResults int    `json:"totalResults"`
	StartIndex   int    `json:"startIndex"`
	Count        int    `json:"count"`
	Users        []User `json:"users"`
	Links        []Link `json:"links"`
}

// ErrorResponse represents an error response from the API
type ErrorResponse struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Description string `json:"description"`
}

// AuthenticationResponse represents the response from an authentication request
type AuthenticationResponse struct {
	ID               string `json:"id"`
	Type             string `json:"type"`
	OrganizationUnit string `json:"organization_unit"`
	Assertion        string `json:"assertion,omitempty"`
}
