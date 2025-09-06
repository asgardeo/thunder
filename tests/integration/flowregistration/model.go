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

package flowregistration

import "encoding/json"

type FlowStep struct {
	FlowID        string   `json:"flowId"`
	FlowStatus    string   `json:"flowStatus"`
	Type          string   `json:"type,omitempty"`
	Data          FlowData `json:"data,omitempty"`
	Assertion     string   `json:"assertion,omitempty"`
	FailureReason string   `json:"failureReason,omitempty"`
}

type FlowData struct {
	Inputs         []InputData       `json:"inputs,omitempty"`
	Actions        []FlowAction      `json:"actions,omitempty"`
	RedirectURL    string            `json:"redirectURL,omitempty"`
	AdditionalData map[string]string `json:"additionalData,omitempty"`
}

type InputData struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

type FlowAction struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type User struct {
	Id               string          `json:"id,omitempty"`
	OrganizationUnit string          `json:"organizationUnit"`
	Type             string          `json:"type"`
	Attributes       json.RawMessage `json:"attributes"`
}

// Link represents a pagination link
type Link struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}

// UserListResponse represents the paginated response for user listing
type UserListResponse struct {
	TotalResults int    `json:"totalResults"`
	StartIndex   int    `json:"startIndex"`
	Count        int    `json:"count"`
	Users        []User `json:"users"`
	Links        []Link `json:"links"`
}

type ErrorResponse struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Description string `json:"description"`
}

type TestSuiteConfig struct {
	CreatedUserIDs    []string
	CreatedSenderID   string
	OriginalAppConfig map[string]interface{}
}

type TestApplication struct {
	ID                        string   `json:"id,omitempty"`
	Name                      string   `json:"name"`
	Description               string   `json:"description"`
	IsRegistrationFlowEnabled bool     `json:"is_registration_flow_enabled"`
	AuthFlowGraphID           string   `json:"auth_flow_graph_id"`
	RegistrationFlowGraphID   string   `json:"registration_flow_graph_id"`
	ClientID                  string   `json:"client_id"`
	ClientSecret              string   `json:"client_secret"`
	RedirectURIs              []string `json:"redirect_uris"`
}

type TestOrganizationUnit struct {
	ID          string  `json:"id,omitempty"`
	Handle      string  `json:"handle"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Parent      *string `json:"parent,omitempty"`
}

// IDP structures for flow registration tests
type IDPProperty struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	IsSecret bool   `json:"is_secret"`
}

type IDP struct {
	ID          string        `json:"id,omitempty"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Properties  []IDPProperty `json:"properties"`
}
