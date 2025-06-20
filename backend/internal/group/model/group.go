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

// Package model defines the data structures and interfaces for group management.
package model

import (
	"errors"
)

// ParentType represents the type of parent entity.
type ParentType string

const (
	// ParentTypeOrganizationUnit is the type for organization units.
	ParentTypeOrganizationUnit ParentType = "organizationUnit"
	// ParentTypeGroup is the type for groups.
	ParentTypeGroup ParentType = "group"
)

// Parent represents the parent of a group (either organization unit or another group).
type Parent struct {
	Type ParentType `json:"type"` // "organizationUnit" or "group"
	ID   string     `json:"id"`
}

// GroupBasic represents the basic information of a group.
type GroupBasic struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Parent      Parent `json:"parent"`
}

// Group represents a complete group with users and child groups.
type Group struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Parent      Parent   `json:"parent"`
	Users       []string `json:"users,omitempty"`
	Groups      []string `json:"groups"`
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

// Error variables
var (
	// ErrGroupNotFound is returned when the group is not found in the system.
	ErrGroupNotFound = errors.New("group not found")

	// ErrGroupNameConflict is returned when a group with the same name exists under the same parent.
	ErrGroupNameConflict = errors.New("a group with the same name exists under the same parent")

	// ErrCannotDeleteGroupWithChildren is returned when trying to delete a group that has child groups.
	ErrCannotDeleteGroupWithChildren = errors.New("cannot delete group with child groups")

	// ErrParentNotFound is returned when the parent group or organization unit is not found.
	ErrParentNotFound = errors.New("parent not found")

	// ErrInvalidRequest is returned when the request data is invalid.
	ErrInvalidRequest = errors.New("invalid request data")
)
