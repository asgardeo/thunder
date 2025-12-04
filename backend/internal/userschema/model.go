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

package userschema

import (
	"encoding/json"
)

// Note: Complex JSON schema type definitions (array, boolean, number, object, schema, string)
// are kept in the model/ subdirectory to maintain clean separation and better organization.
// This file contains only the simple DTOs and API request/response structures.

// UserSchema represents a user type schema definition.
type UserSchema struct {
	ID                    string          `json:"id,omitempty" yaml:"id,omitempty"`
	Name                  string          `json:"name,omitempty" yaml:"name"`
	OrganizationUnitID    string          `json:"ouId" yaml:"organization_unit_id"`
	AllowSelfRegistration bool            `json:"allowSelfRegistration" yaml:"allow_self_registration,omitempty"`
	Schema                json.RawMessage `json:"schema,omitempty" yaml:"schema"`
}

// UserSchemaListItem represents a simplified user schema for listing operations.
type UserSchemaListItem struct {
	ID                    string `json:"id,omitempty"`
	Name                  string `json:"name,omitempty"`
	OrganizationUnitID    string `json:"ouId"`
	AllowSelfRegistration bool   `json:"allowSelfRegistration"`
}

// Link represents a hypermedia link in the API response.
type Link struct {
	Href string `json:"href,omitempty"`
	Rel  string `json:"rel,omitempty"`
}

// UserSchemaListResponse represents the response for listing user schemas with pagination.
type UserSchemaListResponse struct {
	TotalResults int                  `json:"totalResults"`
	StartIndex   int                  `json:"startIndex"`
	Count        int                  `json:"count"`
	Schemas      []UserSchemaListItem `json:"schemas"`
	Links        []Link               `json:"links"`
}

// CreateUserSchemaRequest represents the request body for creating a user schema.
type CreateUserSchemaRequest struct {
	Name                  string          `json:"name"`
	OrganizationUnitID    string          `json:"ouId"`
	AllowSelfRegistration bool            `json:"allowSelfRegistration,omitempty"`
	Schema                json.RawMessage `json:"schema"`
}

// UpdateUserSchemaRequest represents the request body for updating a user schema.
type UpdateUserSchemaRequest struct {
	Name                  string          `json:"name"`
	OrganizationUnitID    string          `json:"ouId"`
	AllowSelfRegistration bool            `json:"allowSelfRegistration,omitempty"`
	Schema                json.RawMessage `json:"schema"`
}

// UserSchemaRequestWithID represents the request structure for creating a user schema from file-based config.
type UserSchemaRequestWithID struct {
	ID                    string `yaml:"id"`
	Name                  string `yaml:"name"`
	OrganizationUnitID    string `yaml:"organization_unit_id"`
	AllowSelfRegistration bool   `yaml:"allow_self_registration,omitempty"`
	Schema                string `yaml:"schema"`
}
