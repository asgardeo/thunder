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

package thememgt

import "encoding/json"

// Theme represents a theme configuration.
type Theme struct {
	ID          string          `json:"id"`
	DisplayName string          `json:"displayName"`
	Description string          `json:"description"`
	Theme       json.RawMessage `json:"theme"`
}

// ThemeListItem represents a theme item in the list response.
type ThemeListItem struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
}

// CreateThemeRequest represents the request body for creating a theme configuration.
type CreateThemeRequest struct {
	DisplayName string          `json:"displayName"`
	Description string          `json:"description"`
	Theme       json.RawMessage `json:"theme"`
}

// UpdateThemeRequest represents the request body for updating a theme configuration.
type UpdateThemeRequest struct {
	DisplayName string          `json:"displayName"`
	Description string          `json:"description"`
	Theme       json.RawMessage `json:"theme"`
}

// ThemeListResponse represents the response for listing theme configurations with pagination.
type ThemeListResponse struct {
	TotalResults int             `json:"totalResults"`
	StartIndex   int             `json:"startIndex"`
	Count        int             `json:"count"`
	Themes       []ThemeListItem `json:"themes"`
	Links        []LinkResponse  `json:"links"`
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

// ThemeList represents the result of listing theme configurations.
type ThemeList struct {
	TotalResults int
	StartIndex   int
	Count        int
	Themes       []Theme
	Links        []Link
}
