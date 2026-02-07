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

package design

import "encoding/json"

// CreateThemeRequest represents the request payload for creating a theme.
type CreateThemeRequest struct {
	DisplayName string          `json:"displayName"`
	Description string          `json:"description,omitempty"`
	Theme       json.RawMessage `json:"theme"`
}

// UpdateThemeRequest represents the request payload for updating a theme.
type UpdateThemeRequest struct {
	DisplayName string          `json:"displayName"`
	Description string          `json:"description,omitempty"`
	Theme       json.RawMessage `json:"theme"`
}

// ThemeResponse represents a theme response.
type ThemeResponse struct {
	ID          string          `json:"id"`
	DisplayName string          `json:"displayName"`
	Description string          `json:"description,omitempty"`
	Theme       json.RawMessage `json:"theme"`
}

// ThemeListItem represents a theme in the list response.
type ThemeListItem struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
}

// ThemeListResponse represents the response for listing themes.
type ThemeListResponse struct {
	TotalResults int             `json:"totalResults"`
	StartIndex   int             `json:"startIndex"`
	Count        int             `json:"count"`
	Themes       []ThemeListItem `json:"themes"`
	Links        []Link          `json:"links,omitempty"`
}

// CreateLayoutRequest represents the request payload for creating a layout.
type CreateLayoutRequest struct {
	DisplayName string          `json:"displayName"`
	Description string          `json:"description,omitempty"`
	Layout      json.RawMessage `json:"layout"`
}

// UpdateLayoutRequest represents the request payload for updating a layout.
type UpdateLayoutRequest struct {
	DisplayName string          `json:"displayName"`
	Description string          `json:"description,omitempty"`
	Layout      json.RawMessage `json:"layout"`
}

// LayoutResponse represents a layout response.
type LayoutResponse struct {
	ID          string          `json:"id"`
	DisplayName string          `json:"displayName"`
	Description string          `json:"description,omitempty"`
	Layout      json.RawMessage `json:"layout"`
}

// LayoutListItem represents a layout in the list response.
type LayoutListItem struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Description string `json:"description,omitempty"`
}

// LayoutListResponse represents the response for listing layouts.
type LayoutListResponse struct {
	TotalResults int              `json:"totalResults"`
	StartIndex   int              `json:"startIndex"`
	Count        int              `json:"count"`
	Layouts      []LayoutListItem `json:"layouts"`
	Links        []Link           `json:"links,omitempty"`
}

// DesignResolveResponse represents the response for design resolve endpoint.
type DesignResolveResponse struct {
	Theme  json.RawMessage `json:"theme,omitempty"`
	Layout json.RawMessage `json:"layout,omitempty"`
}

// Link represents a pagination link.
type Link struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Description string `json:"description,omitempty"`
}
