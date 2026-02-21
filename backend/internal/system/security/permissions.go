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

package security

import "strings"

// maxPublicPathLength defines the maximum allowed length for a public path.
// This prevents potential DoS attacks via excessively long paths (even with safe regex).
const maxPublicPathLength = 4096

// publicPaths defines the list of public paths using glob patterns.
// - "*": Matches a single path segment (e.g., /a/*/b).
// - "**": Matches zero or more path segments (subpaths) at the end of the path (e.g., /a/**).
// Not allowed in the middle of the path (e.g., /a/**/b is invalid).
var publicPaths = []string{
	"/health/**",
	"/auth/**",
	"/register/passkey/**",
	"/flow/execute/**",
	"/flow/meta",
	"/oauth2/**",
	"/.well-known/openid-configuration/**",
	"/.well-known/oauth-authorization-server/**",
	"/.well-known/oauth-protected-resource",
	"/gate/**",
	"/develop/**",
	"/error/**",
	"/design/resolve/**",
	"/i18n/languages",
	"/i18n/languages/*/translations/resolve",
	"/i18n/languages/*/translations/ns/*/keys/*/resolve",
	"/mcp/**", // MCP authorization is handled at MCP server handler.
}

// ---- Resource types ----

// ResourceType defines the category of system resource being acted upon.
type ResourceType string

// ResourceType defines the category of system resource being acted upon.
// ResourceTypeOU, ResourceTypeUser, and ResourceTypeGroup are the supported values.
const (
	// ResourceTypeOU identifies an organization unit resource.
	ResourceTypeOU ResourceType = "ou"
	// ResourceTypeUser identifies a user resource.
	ResourceTypeUser ResourceType = "user"
	// ResourceTypeGroup identifies a group resource.
	ResourceTypeGroup ResourceType = "group"
)

// ---- Actions ----

// Action defines a system operation that can be authorized.
type Action string

const (
	// ActionCreateOU creates a new organization unit.
	ActionCreateOU Action = "ou:create"
	// ActionReadOU reads an organization unit.
	ActionReadOU Action = "ou:read"
	// ActionUpdateOU updates an organization unit.
	ActionUpdateOU Action = "ou:update"
	// ActionDeleteOU deletes an organization unit.
	ActionDeleteOU Action = "ou:delete"
	// ActionListOUs lists organization units.
	ActionListOUs Action = "ou:list"

	// ActionCreateUser creates a new user.
	ActionCreateUser Action = "user:create"
	// ActionReadUser reads a user.
	ActionReadUser Action = "user:read"
	// ActionUpdateUser updates a user.
	ActionUpdateUser Action = "user:update"
	// ActionDeleteUser deletes a user.
	ActionDeleteUser Action = "user:delete"
	// ActionListUsers lists users.
	ActionListUsers Action = "user:list"

	// ActionCreateGroup creates a new group.
	ActionCreateGroup Action = "group:create"
	// ActionReadGroup reads a group.
	ActionReadGroup Action = "group:read"
	// ActionUpdateGroup updates a group.
	ActionUpdateGroup Action = "group:update"
	// ActionDeleteGroup deletes a group.
	ActionDeleteGroup Action = "group:delete"
	// ActionListGroups lists groups.
	ActionListGroups Action = "group:list"
)

// ---- Permissions ----

// SystemPermission is the root permission that implicitly covers all sub-permissions.
// A caller holding "system" is a full admin and bypasses all fine-grained checks.
const SystemPermission = "system"

// Fine-grained permissions. Each constant is a child scope of SystemPermission.
// Hierarchy uses ":" as delimiter: "system:ou" covers "system:ou:view".
const (
	PermissionOU        = "system:ou"
	PermissionOUView    = "system:ou:view"
	PermissionUser      = "system:user"
	PermissionUserView  = "system:user:view"
	PermissionGroup     = "system:group"
	PermissionGroupView = "system:group:view"
)

// ---- Action → Permission map ----

// actionPermissionMap maps each system action to the minimum permission required to perform it.
// Actions not present in this map default to requiring SystemPermission.
var actionPermissionMap = map[Action]string{
	// Organization unit actions.
	ActionCreateOU: PermissionOU,
	ActionReadOU:   PermissionOUView,
	ActionUpdateOU: PermissionOU,
	ActionDeleteOU: PermissionOU,
	ActionListOUs:  PermissionOUView,

	// User actions.
	ActionCreateUser: PermissionUser,
	ActionReadUser:   PermissionUserView,
	ActionUpdateUser: PermissionUser,
	ActionDeleteUser: PermissionUser,
	ActionListUsers:  PermissionUserView,

	// Group actions.
	ActionCreateGroup: PermissionGroup,
	ActionReadGroup:   PermissionGroupView,
	ActionUpdateGroup: PermissionGroup,
	ActionDeleteGroup: PermissionGroup,
	ActionListGroups:  PermissionGroupView,
}

// ---- API → Permission map ----

// apiPermissionEntry pairs a "METHOD glob-path" pattern with the minimum permission
// required for matching requests.
type apiPermissionEntry struct {
	pattern    string
	permission string
}

// apiPermissionEntries defines the ordered set of API permission rules.
// Evaluation is first-match-wins, so more specific patterns must appear before
// broader wildcard patterns. Pattern syntax (applied to the full "METHOD /path" string)
// follows the same glob rules used by publicPaths:
//   - "*"  matches exactly one path segment (e.g., a resource ID).
//   - "**" matches zero or more path segments; only valid as the final component
//     after "/" (e.g., "GET /users/me/**" covers all sub-paths of /users/me).
var apiPermissionEntries = []apiPermissionEntry{
	// Self-service paths — accessible to any authenticated user (empty permission).
	// Listed before their parent wildcards so they always win on first-match.
	{"GET /users/me", ""},
	{"PUT /users/me", ""},
	{"GET /users/me/**", ""},
	{"PUT /users/me/**", ""},
	{"POST /users/me/update-credentials", ""},
	{"GET /register/passkey/**", ""},
	{"POST /register/passkey/**", ""},

	// Organization unit APIs — exact named paths before wildcards.
	{"GET /organization-units/tree", PermissionOUView},
	{"PUT /organization-units/tree", PermissionOU},
	{"DELETE /organization-units/tree", PermissionOU},
	{"GET /organization-units", PermissionOUView},
	{"POST /organization-units", PermissionOU},
	{"GET /organization-units/**", PermissionOUView},
	{"PUT /organization-units/**", PermissionOU},
	{"DELETE /organization-units/**", PermissionOU},

	// User APIs.
	{"GET /users", PermissionUserView},
	{"POST /users", PermissionUser},
	{"GET /users/**", PermissionUserView},
	{"PUT /users/**", PermissionUser},
	{"DELETE /users/**", PermissionUser},

	// Group APIs.
	{"GET /groups", PermissionGroupView},
	{"POST /groups", PermissionGroup},
	{"GET /groups/**", PermissionGroupView},
	{"PUT /groups/**", PermissionGroup},
	{"DELETE /groups/**", PermissionGroup},
}

// ---- Helper functions ----

// HasSystemPermission returns true if the caller holds the root "system" permission.
// This is a fast-path check used to grant unconditional access (skipping policy evaluation).
func HasSystemPermission(permissions []string) bool {
	for _, p := range permissions {
		if p == SystemPermission {
			return true
		}
	}
	return false
}

// HasSufficientPermission returns true if any permission in userPermissions satisfies
// the required permission using hierarchical scope matching.
//
// Matching rules:
//   - Empty required: always satisfied (self-service paths with no specific permission requirement)
//   - Exact match: "system:ou:view" satisfies "system:ou:view"
//   - Parent scope: "system:ou" satisfies "system:ou:view" (parent covers all children)
//   - Root scope: "system" satisfies any "system:*" permission
func HasSufficientPermission(userPermissions []string, required string) bool {
	if required == "" {
		return true
	}
	for _, p := range userPermissions {
		if p == required || strings.HasPrefix(required, p+":") {
			return true
		}
	}
	return false
}

// ResolveActionPermission returns the minimum permission required to perform the given
// action. Falls back to SystemPermission for actions not listed in the action permission map.
func ResolveActionPermission(action Action) string {
	if perm, ok := actionPermissionMap[action]; ok {
		return perm
	}
	return SystemPermission
}
