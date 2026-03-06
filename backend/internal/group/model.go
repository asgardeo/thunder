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

package group

// MemberType represents the type of member entity.
type MemberType string

const (
	// MemberTypeUser is the type for users.
	MemberTypeUser MemberType = "user"
	// MemberTypeGroup is the type for groups.
	MemberTypeGroup MemberType = "group"
)

// Member represents a member of a group (either user or another group).
type Member struct {
	ID   string     `json:"id"`
	Type MemberType `json:"type"`
}

// GroupBasic represents the basic information of a group.
type GroupBasic struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Description        string `json:"description,omitempty"`
	OrganizationUnitID string `json:"organizationUnitId"`
}

// GroupBasicDAO represents a data access object for basic group information,
type GroupBasicDAO struct {
	ID                 string
	Name               string
	Description        string
	OrganizationUnitID string
}

// Group represents a complete group with members.
type Group struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	Description        string   `json:"description,omitempty"`
	OrganizationUnitID string   `json:"organizationUnitId"`
	Members            []Member `json:"members,omitempty"`
}

// GroupDAO represents a data access object for a group, used for database operations.
type GroupDAO struct {
	ID                 string
	Name               string
	Description        string
	OrganizationUnitID string
	Members            []Member
}

// MembersRequest represents the request body for adding or removing members from a group.
type MembersRequest struct {
	Members []Member `json:"members"`
}

// CreateGroupRequest represents the request body for creating a group.
type CreateGroupRequest struct {
	Name               string   `json:"name"`
	Description        string   `json:"description,omitempty"`
	OrganizationUnitID string   `json:"organizationUnitId"`
	Members            []Member `json:"members,omitempty"`
}

// UpdateGroupRequest represents the request body for updating a group.
type UpdateGroupRequest struct {
	Name               string `json:"name"`
	Description        string `json:"description,omitempty"`
	OrganizationUnitID string `json:"organizationUnitId"`
}

// Link represents a pagination link.
type Link struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
}

// GroupListResponse represents the response for listing groups with pagination.
type GroupListResponse struct {
	TotalResults int          `json:"totalResults"`
	StartIndex   int          `json:"startIndex"`
	Count        int          `json:"count"`
	Groups       []GroupBasic `json:"groups"`
	Links        []Link       `json:"links"`
}

// MemberListResponse represents the response for listing group members with pagination.
type MemberListResponse struct {
	TotalResults int      `json:"totalResults"`
	StartIndex   int      `json:"startIndex"`
	Count        int      `json:"count"`
	Members      []Member `json:"members"`
	Links        []Link   `json:"links"`
}

// CreateGroupByPathRequest represents the request body for creating a group under a specific OU path.
type CreateGroupByPathRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Members     []Member `json:"members,omitempty"`
}
