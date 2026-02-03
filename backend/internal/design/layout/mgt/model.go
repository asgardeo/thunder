package layoutmgt

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

import "encoding/json"

// Layout represents a layout configuration.
type Layout struct {
	ID          string          `json:"id"`
	DisplayName string          `json:"displayName"`
	Description string          `json:"description,omitempty"`
	Preferences json.RawMessage `json:"preferences"`
}

// LayoutListItem represents a layout item in the list response.
type LayoutListItem struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Description string `json:"description,omitempty"`
}

// CreateLayoutRequest represents the request body for creating a layout configuration.
type CreateLayoutRequest struct {
	DisplayName string          `json:"displayName"`
	Description string          `json:"description,omitempty"`
	Preferences json.RawMessage `json:"preferences"`
}

// UpdateLayoutRequest represents the request body for updating a layout configuration.
type UpdateLayoutRequest struct {
	DisplayName string          `json:"displayName"`
	Description string          `json:"description,omitempty"`
	Preferences json.RawMessage `json:"preferences"`
}

// LayoutListResponse represents the response for listing layout configurations with pagination.
type LayoutListResponse struct {
	TotalResults int              `json:"totalResults"`
	StartIndex   int              `json:"startIndex"`
	Count        int              `json:"count"`
	Layouts      []LayoutListItem `json:"layouts"`
	Links        []LinkResponse   `json:"links"`
}

// LinkResponse represents a pagination link.
type LinkResponse struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
}

// Link represents a pagination link.
type Link struct {
	Href string
	Rel  string
}

// LayoutList represents the result of listing layout configurations.
type LayoutList struct {
	TotalResults int
	StartIndex   int
	Count        int
	Layouts      []Layout
	Links        []Link
}
