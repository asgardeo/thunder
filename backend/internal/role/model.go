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

// AssignmentResponse represents an assignment of a role to a user or group.
type AssignmentResponse struct {
	ID      string       `json:"id"`
	Type    AssigneeType `json:"type"`
	Display string       `json:"display,omitempty"`
}

// AssignmentRequest represents an assignment of a role to a user or group.
type AssignmentRequest struct {
	ID   string       `json:"id"`
	Type AssigneeType `json:"type"`
}

// RoleSummaryResponse represents the basic information of a role.
type RoleSummaryResponse struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Description        string `json:"description,omitempty"`
	OrganizationUnitID string `json:"ouId"`
}

// RoleResponse represents a complete role with permissions.
type RoleResponse struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	Description        string   `json:"description,omitempty"`
	OrganizationUnitID string   `json:"ouId"`
	Permissions        []string `json:"permissions"`
}

// CreateRoleRequest represents the request body for creating a role.
type CreateRoleRequest struct {
	Name               string              `json:"name"`
	Description        string              `json:"description,omitempty"`
	OrganizationUnitID string              `json:"ouId"`
	Permissions        []string            `json:"permissions"`
	Assignments        []AssignmentRequest `json:"assignments,omitempty"`
}

// CreateRoleResponse represents the response body for creating a role.
type CreateRoleResponse struct {
	ID                 string               `json:"id"`
	Name               string               `json:"name"`
	Description        string               `json:"description,omitempty"`
	OrganizationUnitID string               `json:"ouId"`
	Permissions        []string             `json:"permissions"`
	Assignments        []AssignmentResponse `json:"assignments,omitempty"`
}

// UpdateRoleRequest represents the request body for updating a role.
type UpdateRoleRequest struct {
	Name               string   `json:"name"`
	Description        string   `json:"description,omitempty"`
	OrganizationUnitID string   `json:"ouId"`
	Permissions        []string `json:"permissions"`
}

// AssignmentsRequest represents the request body for adding or removing assignments.
type AssignmentsRequest struct {
	Assignments []AssignmentRequest `json:"assignments"`
}

// LinkResponse represents a pagination link.
type LinkResponse struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
}

// RoleListResponse represents the response for listing roles with pagination.
type RoleListResponse struct {
	TotalResults int                   `json:"totalResults"`
	StartIndex   int                   `json:"startIndex"`
	Count        int                   `json:"count"`
	Roles        []RoleSummaryResponse `json:"roles"`
	Links        []LinkResponse        `json:"links"`
}

// AssignmentListResponse represents the response for listing role assignments with pagination.
type AssignmentListResponse struct {
	TotalResults int                  `json:"totalResults"`
	StartIndex   int                  `json:"startIndex"`
	Count        int                  `json:"count"`
	Assignments  []AssignmentResponse `json:"assignments"`
	Links        []LinkResponse       `json:"links"`
}

// Internal service layer structs - used for business logic processing

// RoleCreationDetail represents the parameters for creating a role.
type RoleCreationDetail struct {
	Name               string
	Description        string
	OrganizationUnitID string
	Permissions        []string
	Assignments        []RoleAssignment
}

// RoleWithPermissionsAndAssignments represents the parameters for creating a role.
type RoleWithPermissionsAndAssignments struct {
	ID                 string
	Name               string
	Description        string
	OrganizationUnitID string
	Permissions        []string
	Assignments        []RoleAssignment
}

// RoleAssignment represents an assignment used internally by the service layer.
type RoleAssignment struct {
	ID   string
	Type AssigneeType
}

// RoleAssignmentWithDisplay represents an assignment used internally by the service layer.
type RoleAssignmentWithDisplay struct {
	ID      string
	Type    AssigneeType
	Display string
}

// Role represents basic role information used internally by the service layer.
type Role struct {
	ID                 string
	Name               string
	Description        string
	OrganizationUnitID string
}

// RoleWithPermissions represents complete role details used internally by the service layer.
type RoleWithPermissions struct {
	ID                 string
	Name               string
	Description        string
	OrganizationUnitID string
	Permissions        []string
}

// RoleUpdateDetail represents the parameters for creating a role.
type RoleUpdateDetail struct {
	Name               string
	Description        string
	OrganizationUnitID string
	Permissions        []string
}

// Link represents a pagination link.
type Link struct {
	Href string
	Rel  string
}

// RoleList represents the result of listing roles.
type RoleList struct {
	TotalResults int
	StartIndex   int
	Count        int
	Roles        []Role
	Links        []Link
}

// AssignmentList represents the result of listing role assignments.
type AssignmentList struct {
	TotalResults int
	StartIndex   int
	Count        int
	Assignments  []RoleAssignmentWithDisplay
	Links        []Link
}
