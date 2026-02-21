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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/security"
)

// stubPolicy is a configurable authorizationPolicy for testing. It allows independent
// control of isActionAllowed and getAccessibleResources behavior.
type stubPolicy struct {
	// isActionAllowed response fields.
	decision  policyDecision
	actionErr *serviceerror.ServiceError

	// getAccessibleResources response fields.
	applicable  bool
	result      *AccessibleResources
	resourceErr *serviceerror.ServiceError
}

func (p *stubPolicy) isActionAllowed(_ context.Context,
	_ *ActionContext) (policyDecision, *serviceerror.ServiceError) {
	return p.decision, p.actionErr
}

func (p *stubPolicy) getAccessibleResources(_ context.Context, _ security.Action,
	_ security.ResourceType) (bool, *AccessibleResources, *serviceerror.ServiceError) {
	return p.applicable, p.result, p.resourceErr
}

// ---------------------------------------------------------------------------
// ouMembershipPolicy.isActionAllowed
// ---------------------------------------------------------------------------

func TestOuMembershipPolicy_IsActionAllowed(t *testing.T) {
	policy := &ouMembershipPolicy{}

	tests := []struct {
		name         string
		ctx          context.Context
		actionCtx    *ActionContext
		wantDecision policyDecision
	}{
		{
			name:         "NilActionCtx_NotApplicable",
			ctx:          context.Background(),
			actionCtx:    nil,
			wantDecision: policyDecisionNotApplicable,
		},
		{
			name:         "EmptyOuID_NotApplicable",
			ctx:          context.Background(),
			actionCtx:    &ActionContext{OuID: ""},
			wantDecision: policyDecisionNotApplicable,
		},
		{
			name:         "MatchingOU_Allowed",
			ctx:          buildCtxWithOU("", "ou1"),
			actionCtx:    &ActionContext{OuID: "ou1"},
			wantDecision: policyDecisionAllowed,
		},
		{
			name:         "MismatchedOU_Denied",
			ctx:          buildCtxWithOU("", "ou2"),
			actionCtx:    &ActionContext{OuID: "ou1"},
			wantDecision: policyDecisionDenied,
		},
		{
			name:         "NoOuInContext_Denied",
			ctx:          context.Background(),
			actionCtx:    &ActionContext{OuID: "ou1"},
			wantDecision: policyDecisionDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision, err := policy.isActionAllowed(tt.ctx, tt.actionCtx)
			assert.Nil(t, err)
			assert.Equal(t, tt.wantDecision, decision)
		})
	}
}

// ---------------------------------------------------------------------------
// ouMembershipPolicy.getAccessibleResources
// ---------------------------------------------------------------------------

func TestOuMembershipPolicy_GetAccessibleResources(t *testing.T) {
	policy := &ouMembershipPolicy{}

	tests := []struct {
		name           string
		ctx            context.Context
		resourceType   security.ResourceType
		wantApplicable bool
		wantAllAllowed bool
		wantIDs        []string
	}{
		{
			name:           "UserResource_NotApplicable",
			ctx:            context.Background(),
			resourceType:   security.ResourceTypeUser,
			wantApplicable: false,
		},
		{
			name:           "GroupResource_NotApplicable",
			ctx:            context.Background(),
			resourceType:   security.ResourceTypeGroup,
			wantApplicable: false,
		},
		{
			name:           "OUResource_EmptyOuIDInContext_RestrictedEmpty",
			ctx:            context.Background(),
			resourceType:   security.ResourceTypeOU,
			wantApplicable: true,
			wantAllAllowed: false,
			wantIDs:        []string{},
		},
		{
			name:           "OUResource_NonEmptyOuID_RestrictedToOU",
			ctx:            buildCtxWithOU("", "ou1"),
			resourceType:   security.ResourceTypeOU,
			wantApplicable: true,
			wantAllAllowed: false,
			wantIDs:        []string{"ou1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			applicable, result, err := policy.getAccessibleResources(tt.ctx, security.ActionListOUs, tt.resourceType)
			assert.Nil(t, err)
			assert.Equal(t, tt.wantApplicable, applicable)
			if tt.wantApplicable {
				assert.NotNil(t, result)
				assert.Equal(t, tt.wantAllAllowed, result.AllAllowed)
				assert.ElementsMatch(t, tt.wantIDs, result.IDs)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// isActionAllowedByPolicies
// ---------------------------------------------------------------------------

func TestIsActionAllowedByPolicies(t *testing.T) {
	errSvc := &serviceerror.ServiceError{Code: "ERR-100", Error: "policy evaluation error"}

	tests := []struct {
		name        string
		policies    []authorizationPolicy
		wantAllowed bool
		wantErr     bool
	}{
		{
			name:        "EmptyPolicies_DefaultAllowed",
			policies:    []authorizationPolicy{},
			wantAllowed: true,
		},
		{
			name: "AllNotApplicable_DefaultAllowed",
			policies: []authorizationPolicy{
				&stubPolicy{decision: policyDecisionNotApplicable},
				&stubPolicy{decision: policyDecisionNotApplicable},
			},
			wantAllowed: true,
		},
		{
			name: "PolicyDenied_ReturnsFalse",
			policies: []authorizationPolicy{
				&stubPolicy{decision: policyDecisionDenied},
			},
			wantAllowed: false,
		},
		{
			name: "AllowedThenDenied_ReturnsFalse",
			policies: []authorizationPolicy{
				&stubPolicy{decision: policyDecisionAllowed},
				&stubPolicy{decision: policyDecisionDenied},
			},
			wantAllowed: false,
		},
		{
			name: "PolicyError_ReturnsFalseAndError",
			policies: []authorizationPolicy{
				&stubPolicy{actionErr: errSvc},
			},
			wantAllowed: false,
			wantErr:     true,
		},
		{
			name: "AllAllowed_DefaultAllowed",
			policies: []authorizationPolicy{
				&stubPolicy{decision: policyDecisionAllowed},
				&stubPolicy{decision: policyDecisionAllowed},
			},
			wantAllowed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := globalPolicies
			globalPolicies = tt.policies
			defer func() { globalPolicies = original }()

			allowed, err := isActionAllowedByPolicies(context.Background(), nil)
			assert.Equal(t, tt.wantAllowed, allowed)
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// getAccessibleResourcesByPolicies
// ---------------------------------------------------------------------------

func TestGetAccessibleResourcesByPolicies(t *testing.T) {
	errSvc := &serviceerror.ServiceError{Code: "ERR-200", Error: "resource policy error"}

	tests := []struct {
		name           string
		policies       []authorizationPolicy
		wantAllAllowed bool
		wantIDs        []string
		wantErr        bool
	}{
		{
			name: "NoApplicablePolicies_AllAllowed",
			policies: []authorizationPolicy{
				&stubPolicy{applicable: false},
				&stubPolicy{applicable: false},
			},
			wantAllAllowed: true,
		},
		{
			name: "FirstApplicableResultReturned",
			policies: []authorizationPolicy{
				&stubPolicy{applicable: false},
				&stubPolicy{
					applicable: true,
					result:     &AccessibleResources{AllAllowed: false, IDs: []string{"ou1", "ou2"}},
				},
			},
			wantAllAllowed: false,
			wantIDs:        []string{"ou1", "ou2"},
		},
		{
			name: "SubsequentPoliciesSkippedAfterFirstApplicable",
			policies: []authorizationPolicy{
				&stubPolicy{applicable: true, result: &AccessibleResources{AllAllowed: false, IDs: []string{"ou1"}}},
				&stubPolicy{applicable: true, result: &AccessibleResources{AllAllowed: false, IDs: []string{"ou2"}}},
			},
			wantAllAllowed: false,
			wantIDs:        []string{"ou1"},
		},
		{
			name: "PolicyError_ReturnsNilAndError",
			policies: []authorizationPolicy{
				&stubPolicy{applicable: true, resourceErr: errSvc},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := globalPolicies
			globalPolicies = tt.policies
			defer func() { globalPolicies = original }()

			result, err := getAccessibleResourcesByPolicies(
				context.Background(), security.ActionListOUs, security.ResourceTypeOU)
			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Nil(t, result)
				return
			}
			assert.Nil(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.wantAllAllowed, result.AllAllowed)
			if tt.wantIDs != nil {
				assert.ElementsMatch(t, tt.wantIDs, result.IDs)
			}
		})
	}
}
