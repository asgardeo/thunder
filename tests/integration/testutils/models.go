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
	"time"
)

// UserSchema represents a user schema (user type) definition
type UserSchema struct {
	ID                    string                 `json:"id,omitempty"`
	Name                  string                 `json:"name"`
	OrganizationUnitId    string                 `json:"ouId"`
	AllowSelfRegistration bool                   `json:"allowSelfRegistration,omitempty"`
	Schema                map[string]interface{} `json:"schema"`
}

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
	AuthFlowID                string                   `json:"auth_flow_id,omitempty"`
	RegistrationFlowID        string                   `json:"registration_flow_id,omitempty"`
	ClientID                  string                   `json:"client_id,omitempty"`
	ClientSecret              string                   `json:"client_secret,omitempty"`
	RedirectURIs              []string                 `json:"redirect_uris,omitempty"`
	AllowedUserTypes          []string                 `json:"allowed_user_types,omitempty"`
	Certificate               map[string]interface{}   `json:"certificate,omitempty"`
	InboundAuthConfig         []map[string]interface{} `json:"inbound_auth_config,omitempty"`
	AssertionConfig           map[string]interface{}   `json:"assertion,omitempty"`
}

// OrganizationUnit represents an organization unit in the system
type OrganizationUnit struct {
	ID              string  `json:"id,omitempty"`
	Handle          string  `json:"handle"`
	Name            string  `json:"name"`
	Description     string  `json:"description,omitempty"`
	Parent          *string `json:"parent,omitempty"`
	LogoURL         string  `json:"logo_url,omitempty"`
	TosURI          string  `json:"tos_uri,omitempty"`
	PolicyURI       string  `json:"policy_uri,omitempty"`
	CookiePolicyURI string  `json:"cookie_policy_uri,omitempty"`
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

// GroupMember represents a member of a group
type GroupMember struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Display string `json:"display,omitempty"`
}

// Group represents a group in the system
type Group struct {
	ID                 string `json:"id,omitempty"`
	Name               string `json:"name"`
	Description        string `json:"description,omitempty"`
	OrganizationUnitId string `json:"organizationUnitId,omitempty"`
}

// Assignment represents a role assignment
type Assignment struct {
	ID      string `json:"id"`
	Type    string `json:"type"` // "user" or "group"
	Display string `json:"display,omitempty"`
}

// Role represents a role in the system
type Role struct {
	ID                 string                `json:"id,omitempty"`
	Name               string                `json:"name"`
	Description        string                `json:"description,omitempty"`
	OrganizationUnitID string                `json:"ouId"`
	Permissions        []ResourcePermissions `json:"permissions,omitempty"`
	Assignments        []Assignment          `json:"assignments,omitempty"`
}

// TokenResponse represents the response from token exchange
type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    float64   `json:"expires_in"`
	Scope        string    `json:"scope,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	IDToken      string    `json:"id_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"` // Absolute expiry time
}

// TokenHTTPResult captures raw HTTP response details from the token endpoint.
type TokenHTTPResult struct {
	StatusCode int
	Body       []byte
	Token      *TokenResponse
}

// AuthorizationResponse represents the response from authorization completion
type AuthorizationResponse struct {
	RedirectURI string `json:"redirect_uri"`
}

// ResourceServer represents a resource server in the system
type ResourceServer struct {
	ID                 string  `json:"id,omitempty"`
	Name               string  `json:"name"`
	Description        string  `json:"description,omitempty"`
	Identifier         string  `json:"identifier,omitempty"`
	OrganizationUnitID string  `json:"ouId"`
	Delimiter          *string `json:"delimiter,omitempty"`
}

// Action represents an action in the resource system
type Action struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Handle      string `json:"handle"`
	Description string `json:"description,omitempty"`
	Permission  string `json:"permission,omitempty"`
}

// ResourcePermissions represents permissions grouped by resource server
type ResourcePermissions struct {
	ResourceServerID string   `json:"resourceServerId"`
	Permissions      []string `json:"permissions"`
}

// FlowResponse represents the response from flow execution
type FlowResponse struct {
	FlowID        string    `json:"flowId"`
	FlowStatus    string    `json:"flowStatus"`
	Type          string    `json:"type"`
	Data          *FlowData `json:"data,omitempty"`
	Assertion     string    `json:"assertion,omitempty"`
	FailureReason string    `json:"failureReason,omitempty"`
}

// FlowData represents the data returned by flow execution
type FlowData struct {
	Inputs         []FlowInput            `json:"inputs,omitempty"`
	Actions        []FlowAction           `json:"actions,omitempty"`
	RedirectURL    string                 `json:"redirectUrl,omitempty"`
	Meta           map[string]interface{} `json:"meta,omitempty"`
	AdditionalData map[string]interface{} `json:"additionalData,omitempty"`
}

// FlowInput represents an input required by the flow
type FlowInput struct {
	Ref        string `json:"ref,omitempty"`
	Identifier string `json:"identifier"`
	Type       string `json:"type"`
	Required   bool   `json:"required"`
}

// FlowAction represents an action available in the flow
type FlowAction struct {
	Ref      string `json:"ref"`
	NextNode string `json:"nextNode"`
}

// FlowStep represents a single step in a flow execution
type FlowStep struct {
	FlowID        string    `json:"flowId"`
	FlowStatus    string    `json:"flowStatus"`
	Type          string    `json:"type"`
	Data          *FlowData `json:"data,omitempty"`
	Assertion     string    `json:"assertion,omitempty"`
	FailureReason string    `json:"failureReason,omitempty"`
}

// Flow represents a flow definition
type Flow struct {
	Name     string      `json:"name"`
	FlowType string      `json:"flowType"`
	Handle   string      `json:"handle"`
	Nodes    interface{} `json:"nodes"`
}

// NotificationSender represents a notification sender in the system
type NotificationSender struct {
	ID          string           `json:"id,omitempty"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Provider    string           `json:"provider"`
	Properties  []SenderProperty `json:"properties"`
}

// SenderProperty represents a property of a notification sender
type SenderProperty struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	IsSecret bool   `json:"is_secret"`
}
