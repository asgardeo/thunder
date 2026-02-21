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

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/security"
)

// TestMain enables debug-level logging for the entire package test binary so that
// every logger.IsDebugEnabled() branch in service.go is exercised.
func TestMain(m *testing.M) {
	_ = os.Setenv("LOG_LEVEL", "debug")
	os.Exit(m.Run())
}

// SystemAuthzTestSuite defines the test suite for the system authorization service.
type SystemAuthzTestSuite struct {
	suite.Suite
	service SystemAuthorizationServiceInterface
}

func (s *SystemAuthzTestSuite) SetupTest() {
	var err error
	s.service, err = Initialize()
	s.Require().NoError(err)
}

func TestSystemAuthzSuite(t *testing.T) {
	suite.Run(t, new(SystemAuthzTestSuite))
}

// ---------------------------------------------------------------------------
// Context builder helpers
// ---------------------------------------------------------------------------

// buildCtx creates an authenticated security context with the given space-separated permissions.
func buildCtx(permissions string) context.Context {
	return buildCtxWithOU(permissions, "")
}

// buildCtxWithOU creates an authenticated security context with given permissions and OU ID.
func buildCtxWithOU(permissions, ouID string) context.Context {
	var perms []string
	if permissions != "" {
		perms = strings.Fields(permissions)
	}
	authCtx := security.NewSecurityContextForTest("user123", ouID, "token", perms, nil)
	return security.WithSecurityContextTest(context.Background(), authCtx)
}

// buildSkipSecurityCtx returns a context with security enforcement skipped.
func buildSkipSecurityCtx() context.Context {
	return security.WithSkipSecurityTest(context.Background())
}

// buildRuntimeCtx returns a context marked as an internal runtime caller.
func buildRuntimeCtx() context.Context {
	return security.WithRuntimeContext(context.Background())
}

// ---------------------------------------------------------------------------
// IsActionAllowed
// ---------------------------------------------------------------------------

func (s *SystemAuthzTestSuite) TestIsActionAllowed() {
	tests := []struct {
		name             string
		ctx              context.Context
		action           security.Action
		actionCtx        *ActionContext
		wantAllowed      bool
		wantErr          bool
		overridePolicies []authorizationPolicy
	}{
		{
			// Step 1: THUNDER_SKIP_SECURITY flag bypasses all checks.
			name:        "SecuritySkipped_GrantsAccess",
			ctx:         buildSkipSecurityCtx(),
			action:      security.ActionReadUser,
			wantAllowed: true,
		},
		{
			// Step 2: Internal runtime caller is granted access without a subject.
			name:        "RuntimeContext_GrantsAccess",
			ctx:         buildRuntimeCtx(),
			action:      security.ActionCreateOU,
			wantAllowed: true,
		},
		{
			// Step 3: No security context → empty subject → denied.
			name:        "UnauthenticatedCaller_Denied",
			ctx:         context.Background(),
			action:      security.ActionReadUser,
			wantAllowed: false,
		},
		{
			// Step 4: The "system" permission short-circuits to allowed.
			name:        "SystemScope_GrantsAllActions",
			ctx:         buildCtx("system"),
			action:      security.ActionDeleteUser,
			wantAllowed: true,
		},
		{
			// Step 4: "system" among other permissions still grants access.
			name:        "SystemScopeAmongOtherScopes_GrantsAccess",
			ctx:         buildCtx("users:read system groups:write"),
			action:      security.ActionDeleteUser,
			wantAllowed: true,
		},
		{
			// Step 4 + actionCtx: system scope ignores actionCtx and grants access.
			name:   "SystemScope_WithActionContext_GrantsAccess",
			ctx:    buildCtx("system"),
			action: security.ActionReadUser,
			actionCtx: &ActionContext{
				OuID: "ou-abc", ResourceType: security.ResourceTypeUser, ResourceID: "user-xyz",
			},
			wantAllowed: true,
		},
		{
			// Step 5: Insufficient permissions → denied (also exercises IsDebugEnabled branch).
			name:        "InsufficientScopes_Denied",
			ctx:         buildCtx("users:read groups:manage"),
			action:      security.ActionDeleteUser,
			wantAllowed: false,
		},
		{
			// Step 5: Empty permission set → denied.
			name:        "EmptyScopes_Denied",
			ctx:         buildCtx(""),
			action:      security.ActionReadUser,
			wantAllowed: false,
		},
		{
			// Step 5: Unmapped action without system permission falls back to "system" requirement.
			name:        "UnmappedAction_InsufficientScope_Denied",
			ctx:         buildCtx("users:read"),
			action:      security.Action("custom:action"),
			wantAllowed: false,
		},
		{
			// Step 4: Unmapped action with system permission is still allowed.
			name:        "UnmappedAction_SystemScope_Allowed",
			ctx:         buildCtx("system"),
			action:      security.Action("custom:action"),
			wantAllowed: true,
		},
		{
			// Step 6: Has required permission, nil actionCtx → policy NotApplicable → allowed.
			// Also exercises the final IsDebugEnabled("Authorization granted") branch.
			name:        "RequiredPermission_NilActionCtx_PolicyNotApplicable_Allowed",
			ctx:         buildCtx("system:user"),
			action:      security.ActionCreateUser,
			actionCtx:   nil,
			wantAllowed: true,
		},
		{
			// Step 6: Has required permission, actionCtx OU matches context OU → policy Allowed.
			name:        "RequiredPermission_MatchingOU_PolicyAllowed",
			ctx:         buildCtxWithOU("system:ou", "ou1"),
			action:      security.ActionCreateOU,
			actionCtx:   &ActionContext{OuID: "ou1"},
			wantAllowed: true,
		},
		{
			// Step 6: Has required permission, actionCtx OU differs from context OU → policy Denied.
			// Also exercises the IsDebugEnabled("Authorization denied: policy evaluation failed") branch.
			name:        "RequiredPermission_MismatchedOU_PolicyDenied",
			ctx:         buildCtxWithOU("system:ou", "ou1"),
			action:      security.ActionCreateOU,
			actionCtx:   &ActionContext{OuID: "ou2"},
			wantAllowed: false,
		},
		{
			// Step 6: Policy returns a ServiceError → propagated to caller.
			name:        "PolicyError_PropagatedToCallerAsServiceError",
			ctx:         buildCtx("system:user"),
			action:      security.ActionCreateUser,
			wantAllowed: false,
			wantErr:     true,
			overridePolicies: []authorizationPolicy{
				&stubPolicy{actionErr: &serviceerror.ServiceError{Code: "ERR-001", Error: "policy failure"}},
			},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			if tt.overridePolicies != nil {
				original := globalPolicies
				globalPolicies = tt.overridePolicies
				defer func() { globalPolicies = original }()
			}
			allowed, svcErr := s.service.IsActionAllowed(tt.ctx, tt.action, tt.actionCtx)
			assert.Equal(t, tt.wantAllowed, allowed)
			if tt.wantErr {
				assert.NotNil(t, svcErr)
			} else {
				assert.Nil(t, svcErr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GetAccessibleResources
// ---------------------------------------------------------------------------

func (s *SystemAuthzTestSuite) TestGetAccessibleResources() {
	tests := []struct {
		name             string
		ctx              context.Context
		action           security.Action
		resourceType     security.ResourceType
		wantAllAllowed   bool
		wantIDs          []string
		wantErr          bool
		overridePolicies []authorizationPolicy
	}{
		{
			// Step 1: THUNDER_SKIP_SECURITY flag → all resources accessible.
			name:           "SecuritySkipped_AllAllowed",
			ctx:            buildSkipSecurityCtx(),
			action:         security.ActionListUsers,
			resourceType:   security.ResourceTypeUser,
			wantAllAllowed: true,
		},
		{
			// Step 2: Internal runtime caller → all resources accessible.
			name:           "RuntimeContext_AllAllowed",
			ctx:            buildRuntimeCtx(),
			action:         security.ActionListUsers,
			resourceType:   security.ResourceTypeUser,
			wantAllAllowed: true,
		},
		{
			// Step 3: No security context → empty subject → no resources.
			name:           "UnauthenticatedCaller_Denied",
			ctx:            context.Background(),
			action:         security.ActionListUsers,
			resourceType:   security.ResourceTypeUser,
			wantAllAllowed: false,
			wantIDs:        []string{},
		},
		{
			// Step 4: "system" permission → all resources accessible.
			name:           "SystemScope_AllAllowed_OUs",
			ctx:            buildCtx("system"),
			action:         security.ActionListOUs,
			resourceType:   security.ResourceTypeOU,
			wantAllAllowed: true,
		},
		{
			// Step 4: "system" permission for all resource types.
			name:           "SystemScope_AllAllowed_Groups",
			ctx:            buildCtx("system"),
			action:         security.ActionListGroups,
			resourceType:   security.ResourceTypeGroup,
			wantAllAllowed: true,
		},
		{
			// Step 5: Insufficient permissions → no resources (also exercises IsDebugEnabled branch).
			name:           "InsufficientScopes_Denied",
			ctx:            buildCtx("users:read"),
			action:         security.ActionListUsers,
			resourceType:   security.ResourceTypeUser,
			wantAllAllowed: false,
			wantIDs:        []string{},
		},
		{
			// Step 5: Empty permission set → no resources.
			name:           "EmptyScopes_Denied",
			ctx:            buildCtx(""),
			action:         security.ActionListGroups,
			resourceType:   security.ResourceTypeGroup,
			wantAllAllowed: false,
			wantIDs:        []string{},
		},
		{
			// Step 6: Has required permission, non-OU resource → policy not applicable → all allowed.
			name:           "RequiredPermission_NonOUResource_PolicyNotApplicable_AllAllowed",
			ctx:            buildCtxWithOU("system:user:view", "ou1"),
			action:         security.ActionListUsers,
			resourceType:   security.ResourceTypeUser,
			wantAllAllowed: true,
		},
		{
			// Step 6: Has required permission, OU resource with a non-empty ouID in context
			// → ouMembershipPolicy restricts to that OU.
			// Also exercises the IsDebugEnabled("restricted by policy") branch.
			name:           "RequiredPermission_OUResource_PolicyRestricted",
			ctx:            buildCtxWithOU("system:ou:view", "ou1"),
			action:         security.ActionListOUs,
			resourceType:   security.ResourceTypeOU,
			wantAllAllowed: false,
			wantIDs:        []string{"ou1"},
		},
		{
			// Step 6: Policy returns a ServiceError → propagated to caller.
			name:         "PolicyError_PropagatedToCallerAsServiceError",
			ctx:          buildCtx("system:user:view"),
			action:       security.ActionListUsers,
			resourceType: security.ResourceTypeUser,
			wantErr:      true,
			overridePolicies: []authorizationPolicy{
				&stubPolicy{
					applicable:  true,
					resourceErr: &serviceerror.ServiceError{Code: "ERR-002", Error: "resource policy error"},
				},
			},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			if tt.overridePolicies != nil {
				original := globalPolicies
				globalPolicies = tt.overridePolicies
				defer func() { globalPolicies = original }()
			}
			result, svcErr := s.service.GetAccessibleResources(tt.ctx, tt.action, tt.resourceType)
			if tt.wantErr {
				assert.NotNil(t, svcErr)
				assert.Nil(t, result)
				return
			}
			assert.Nil(t, svcErr)
			assert.NotNil(t, result)
			assert.Equal(t, tt.wantAllAllowed, result.AllAllowed)
			if tt.wantIDs != nil {
				assert.ElementsMatch(t, tt.wantIDs, result.IDs)
			}
		})
	}
}
