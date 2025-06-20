/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
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

// ParentType represents the type of parent entity.
type ParentType string

const (
	ParentTypeOrganizationUnit ParentType = "organizationUnit"
	ParentTypeGroup            ParentType = "group"
)

// Parent represents the parent of a group (either organization unit or another group).
type Parent struct {
	Type ParentType `json:"type"` // "organizationUnit" or "group"
	Id   string     `json:"id"`
}

// GroupBasic represents the basic information of a group.
type GroupBasic struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Parent      Parent   `json:"parent"`
	Groups      []string `json:"groups"` // Child group Ids
}

// Group represents a complete group with users.
type Group struct {
	GroupBasic
	Users []string `json:"users,omitempty"` // User Ids
}

// CreateGroupRequest represents the request body for creating a group.
type CreateGroupRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Parent      Parent   `json:"parent"`
	Users       []string `json:"users,omitempty"`
}

// UpdateGroupRequest represents the request body for updating a group.
type UpdateGroupRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Parent      Parent   `json:"parent"`
	Users       []string `json:"users,omitempty"`
	Groups      []string `json:"groups,omitempty"`
}
