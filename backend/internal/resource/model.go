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

package resource

// HTTP Response Models

// ResourceServerResponse represents a resource server.
type ResourceServerResponse struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Description        string `json:"description,omitempty"`
	Identifier         string `json:"identifier,omitempty"`
	OrganizationUnitID string `json:"ouId"`
	Delimiter          string `json:"delimiter"`
}

// ResourceResponse represents a resource.
type ResourceResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Handle      string  `json:"handle"`
	Description string  `json:"description,omitempty"`
	Parent      *string `json:"parent,omitempty"`
	Permission  string  `json:"permission"`
}

// ActionResponse represents an action.
type ActionResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Handle      string `json:"handle"`
	Description string `json:"description,omitempty"`
	Permission  string `json:"permission"`
}

// LinkResponse represents a pagination link.
type LinkResponse struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
}

// ResourceServerListResponse represents the response for listing resource servers.
type ResourceServerListResponse struct {
	TotalResults    int                      `json:"totalResults"`
	StartIndex      int                      `json:"startIndex"`
	Count           int                      `json:"count"`
	ResourceServers []ResourceServerResponse `json:"resourceServers"`
	Links           []LinkResponse           `json:"links"`
}

// ResourceListResponse represents the response for listing resources.
type ResourceListResponse struct {
	TotalResults int                `json:"totalResults"`
	StartIndex   int                `json:"startIndex"`
	Count        int                `json:"count"`
	Resources    []ResourceResponse `json:"resources"`
	Links        []LinkResponse     `json:"links"`
}

// ActionListResponse represents the response for listing actions.
type ActionListResponse struct {
	TotalResults int              `json:"totalResults"`
	StartIndex   int              `json:"startIndex"`
	Count        int              `json:"count"`
	Actions      []ActionResponse `json:"actions"`
	Links        []LinkResponse   `json:"links"`
}

// CreateResourceServerRequest represents the request to create a resource server.
type CreateResourceServerRequest struct {
	Name               string `json:"name"`
	Description        string `json:"description,omitempty"`
	Identifier         string `json:"identifier,omitempty"`
	OrganizationUnitID string `json:"ouId"`
	Delimiter          string `json:"delimiter,omitempty"`
}

// UpdateResourceServerRequest represents the request to update a resource server.
type UpdateResourceServerRequest struct {
	Name               string `json:"name"`
	Description        string `json:"description,omitempty"`
	Identifier         string `json:"identifier,omitempty"`
	OrganizationUnitID string `json:"ouId"`
}

// CreateResourceRequest represents the request to create a resource.
type CreateResourceRequest struct {
	Name        string  `json:"name"`
	Handle      string  `json:"handle"`
	Description string  `json:"description,omitempty"`
	Parent      *string `json:"parent"`
}

// UpdateResourceRequest represents the request to update a resource.
type UpdateResourceRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// CreateActionRequest represents the request to create an action.
type CreateActionRequest struct {
	Name        string `json:"name"`
	Handle      string `json:"handle"`
	Description string `json:"description,omitempty"`
}

// UpdateActionRequest represents the request to update an action.
type UpdateActionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// Service layer structs

// ResourceServer represents a resource server in the service layer.
type ResourceServer struct {
	ID                 string
	Name               string
	Description        string
	Identifier         string
	OrganizationUnitID string
	Delimiter          string
}

// Resource represents a resource in the service layer.
type Resource struct {
	ID          string
	Name        string
	Handle      string
	Description string
	Parent      *string
	Permission  string
}

// Action represents an action in the service layer.
type Action struct {
	ID          string
	Name        string
	Handle      string
	Description string
	Permission  string
}

// Link represents a pagination link in the service layer.
type Link struct {
	Href string
	Rel  string
}

// ResourceServerList represents the result of listing resource servers.
type ResourceServerList struct {
	TotalResults    int
	StartIndex      int
	Count           int
	ResourceServers []ResourceServer
	Links           []Link
}

// ResourceList represents the result of listing resources.
type ResourceList struct {
	TotalResults int
	StartIndex   int
	Count        int
	Resources    []Resource
	Links        []Link
}

// ActionList represents the result of listing actions.
type ActionList struct {
	TotalResults int
	StartIndex   int
	Count        int
	Actions      []Action
	Links        []Link
}
