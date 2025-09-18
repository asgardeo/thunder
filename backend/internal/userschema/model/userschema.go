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

// Package model defines the data structures for user schema management.
package model

import (
	"encoding/json"
)

// UserSchema represents a user type schema definition.
type UserSchema struct {
	ID     string          `json:"id,omitempty"`
	Name   string          `json:"name,omitempty"`
	Schema json.RawMessage `json:"schema,omitempty"`
}

// UserSchemaListItem represents a simplified user schema for listing operations.
type UserSchemaListItem struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
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
	Name   string          `json:"name"`
	Schema json.RawMessage `json:"schema"`
}

// UpdateUserSchemaRequest represents the request body for updating a user schema.
type UpdateUserSchemaRequest struct {
	Name   string          `json:"name"`
	Schema json.RawMessage `json:"schema"`
}

// JSON Schema type constants.
const (
	// TypeString represents the string type in JSON Schema.
	TypeString = "string"
	// TypeNumber represents the number type in JSON Schema.
	TypeNumber = "number"
	// TypeBoolean represents the boolean type in JSON Schema.
	TypeBoolean = "boolean"
	// TypeObject represents the object type in JSON Schema.
	TypeObject = "object"
	// TypeArray represents the array type in JSON Schema.
	TypeArray = "array"
)
