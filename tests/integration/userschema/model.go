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

	"github.com/asgardeo/thunder/tests/integration/testutils"
)

const (
	testServerURL = "https://localhost:8095"
)

// UserSchema represents the user schema model for tests
type UserSchema struct {
	ID                    string          `json:"id,omitempty"`
	Name                  string          `json:"name"`
	OrganizationUnitID    string          `json:"ouId"`
	AllowSelfRegistration bool            `json:"allowSelfRegistration,omitempty"`
	DisplayAttribute      string          `json:"displayAttribute,omitempty"`
	Schema                json.RawMessage `json:"schema"`
}

// CreateUserSchemaRequest represents the request to create a user schema
type CreateUserSchemaRequest struct {
	Name                  string          `json:"name"`
	OrganizationUnitID    string          `json:"ouId"`
	AllowSelfRegistration bool            `json:"allowSelfRegistration,omitempty"`
	DisplayAttribute      string          `json:"displayAttribute,omitempty"`
	Schema                json.RawMessage `json:"schema"`
}

// UpdateUserSchemaRequest represents the request to update a user schema
type UpdateUserSchemaRequest struct {
	Name                  string          `json:"name"`
	OrganizationUnitID    string          `json:"ouId"`
	AllowSelfRegistration bool            `json:"allowSelfRegistration,omitempty"`
	DisplayAttribute      string          `json:"displayAttribute,omitempty"`
	Schema                json.RawMessage `json:"schema"`
}

// UserSchemaListItem represents a simplified user schema for listing operations in tests
type UserSchemaListItem struct {
	ID                    string `json:"id,omitempty"`
	Name                  string `json:"name,omitempty"`
	OrganizationUnitID    string `json:"ouId"`
	AllowSelfRegistration bool   `json:"allowSelfRegistration,omitempty"`
	DisplayAttribute      string `json:"displayAttribute,omitempty"`
}

// UserSchemaListResponse represents the response from listing user schemas
type UserSchemaListResponse struct {
	TotalResults int                  `json:"totalResults"`
	StartIndex   int                  `json:"startIndex"`
	Count        int                  `json:"count"`
	Schemas      []UserSchemaListItem `json:"schemas"`
	Links        []testutils.Link     `json:"links"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Description string `json:"description,omitempty"`
	TraceID     string `json:"traceId,omitempty"`
}

// OrganizationUnit represents an organization unit
type OrganizationUnit struct {
	ID          string  `json:"id"`
	Handle      string  `json:"handle"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Parent      *string `json:"parent,omitempty"`
}

// CreateUserRequest represents the request to create a user
type CreateUserRequest struct {
	OrganizationUnit string          `json:"organizationUnit"`
	Type             string          `json:"type"`
	Attributes       json.RawMessage `json:"attributes,omitempty"`
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	OrganizationUnit string          `json:"organizationUnit,omitempty"`
	Type             string          `json:"type,omitempty"`
	Attributes       json.RawMessage `json:"attributes,omitempty"`
}

// CreateUserByPathRequest represents the request to create a user under a handle path
type CreateUserByPathRequest struct {
	Type       string          `json:"type"`
	Attributes json.RawMessage `json:"attributes,omitempty"`
}

// CreateOURequest represents the request to create an organization unit
type CreateOURequest struct {
	Handle      string  `json:"handle"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Parent      *string `json:"parent,omitempty"`
}
