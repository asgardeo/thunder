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

package role

// AssigneeType represents the type of assignee entity.
type AssigneeType string

const (
	// AssigneeTypeUser is the type for users.
	AssigneeTypeUser AssigneeType = "user"
	// AssigneeTypeGroup is the type for groups.
	AssigneeTypeGroup AssigneeType = "group"
)

// Assignment represents an assignment of a role to a user or group.
type Assignment struct {
	ID   string       `json:"id"`
	Type AssigneeType `json:"type"`
}

// RoleSummary represents the basic information of a role.
type RoleSummary struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Description        string `json:"description,omitempty"`
	OrganizationUnitID string `json:"organizationUnitId"`
}

// Role represents a complete role with permissions and assignments.
type Role struct {
	ID                 string       `json:"id"`
	Name               string       `json:"name"`
	Description        string       `json:"description,omitempty"`
	OrganizationUnitID string       `json:"organizationUnitId"`
	Permissions        []string     `json:"permissions"`
	Assignments        []Assignment `json:"assignments,omitempty"`
}

// CreateRoleRequest represents the request body for creating a role.
type CreateRoleRequest struct {
	Name               string       `json:"name"`
	Description        string       `json:"description,omitempty"`
	OrganizationUnitID string       `json:"organizationUnitId"`
	Permissions        []string     `json:"permissions"`
	Assignments        []Assignment `json:"assignments,omitempty"`
}

// UpdateRoleRequest represents the request body for updating a role.
type UpdateRoleRequest struct {
	Name               string       `json:"name"`
	Description        string       `json:"description,omitempty"`
	OrganizationUnitID string       `json:"organizationUnitId"`
	Permissions        []string     `json:"permissions"`
	Assignments        []Assignment `json:"assignments,omitempty"`
}

// AssignmentsRequest represents the request body for adding or removing assignments.
type AssignmentsRequest struct {
	Assignments []Assignment `json:"assignments"`
}

// Link represents a pagination link.
type Link struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
}

// RoleListResponse represents the response for listing roles with pagination.
type RoleListResponse struct {
	TotalResults int           `json:"totalResults"`
	StartIndex   int           `json:"startIndex"`
	Count        int           `json:"count"`
	Roles        []RoleSummary `json:"roles"`
	Links        []Link        `json:"links"`
}

// AssignmentListResponse represents the response for listing role assignments with pagination.
type AssignmentListResponse struct {
	TotalResults int          `json:"totalResults"`
	StartIndex   int          `json:"startIndex"`
	Count        int          `json:"count"`
	Assignments  []Assignment `json:"assignments"`
	Links        []Link       `json:"links"`
}
