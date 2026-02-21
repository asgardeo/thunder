/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

package sysauthz

import "github.com/asgardeo/thunder/internal/system/security"

// ActionContext provides contextual information used to make an authorization decision.
// Not all fields are required for every action; populate only those relevant to the operation.
type ActionContext struct {
	// OuID is the organization unit ID scoping the action.
	// Leave empty if the action is not scoped to a specific OU.
	OuID string
	// ResourceType is the type of resource being acted upon.
	ResourceType security.ResourceType
	// ResourceID is the identifier of the specific resource being acted upon.
	// Leave empty for collection-level actions (e.g., list, create).
	ResourceID string
}

// AccessibleResources represents the set of resources a caller is permitted to access
// for a given action. It is used to pre-filter store queries before pagination is applied.
type AccessibleResources struct {
	// AllAllowed signals that the caller may access all resources of the requested type.
	// When true, the caller should apply no ID filter to the store query.
	// When false, only the IDs listed in IDs are accessible.
	AllAllowed bool
	// IDs is the explicit set of accessible resource IDs.
	// Only populated when AllAllowed is false.
	IDs []string
}
