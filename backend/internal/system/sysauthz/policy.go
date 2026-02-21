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

	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/security"
)

// policyDecision is the outcome of a single policy evaluation.
type policyDecision int

const (
	// policyDecisionNotApplicable indicates the policy has no opinion on this
	// context (e.g., the action is not OU-scoped). The next policy in the chain
	// will be consulted. If all policies return NotApplicable, the action is allowed.
	policyDecisionNotApplicable policyDecision = iota
	// policyDecisionAllowed indicates the policy explicitly permits the action.
	policyDecisionAllowed
	// policyDecisionDenied indicates the policy explicitly denies the action.
	policyDecisionDenied
)

// authorizationPolicy defines an authorization rule evaluated after permission checks pass.
// It is the primary extension point for introducing fine-grained access control
// (e.g., attribute-based or relationship-based policies) without changing the
// SystemAuthorizationServiceInterface contract.
//
// A policy mirrors the two methods of SystemAuthorizationServiceInterface at the
// policy layer:
//   - isActionAllowed: called by IsActionAllowed for single-resource operations.
//   - getAccessibleResources: called by GetAccessibleResources for list operations.
type authorizationPolicy interface {
	// isActionAllowed returns the policy decision for the caller in the given context.
	// A non-nil ServiceError signals a policy evaluation failure, not a denial.
	isActionAllowed(ctx context.Context, actionCtx *ActionContext) (policyDecision, *serviceerror.ServiceError)

	// getAccessibleResources reports whether this policy is applicable for the
	// given action and resource type, and if so, the set of resources the caller
	// may access. When applicable is false the policy has no opinion for this
	// resource type and the result must be ignored by the caller.
	// A non-nil ServiceError signals an evaluation failure, not a denial.
	getAccessibleResources(ctx context.Context, action security.Action,
		resourceType security.ResourceType,
	) (applicable bool, result *AccessibleResources, err *serviceerror.ServiceError)
}

// ouMembershipPolicy enforces that the caller's organization unit matches the OU of the
// resource being acted upon. This prevents non-system callers from operating on
// resources that belong to a different OU.
//
// Future evolution: replace or augment with a ReBAC policy that queries a
// relationship graph (e.g., "is the caller a member of the resource's OU hierarchy?").
type ouMembershipPolicy struct{}

// isActionAllowed returns:
//   - PolicyDecisionNotApplicable when the action context carries no OuID.
//   - PolicyDecisionAllowed when the caller's OU matches the resource's OU.
//   - PolicyDecisionDenied when the caller's OU does not match.
func (p *ouMembershipPolicy) isActionAllowed(ctx context.Context,
	actionCtx *ActionContext) (policyDecision, *serviceerror.ServiceError) {
	if actionCtx == nil || actionCtx.OuID == "" {
		return policyDecisionNotApplicable, nil
	}
	if security.GetOUID(ctx) == actionCtx.OuID {
		return policyDecisionAllowed, nil
	}
	return policyDecisionDenied, nil
}

// getAccessibleResources constrains list operations by the caller's OU membership:
//   - For non-ResourceTypeOU resource types: not applicable — OU-based filtering
//     for users and groups is applied at the store layer.
//   - For ResourceTypeOU: the caller may only see their own OU.
func (p *ouMembershipPolicy) getAccessibleResources(ctx context.Context, action security.Action,
	resourceType security.ResourceType) (bool, *AccessibleResources, *serviceerror.ServiceError) {
	if resourceType != security.ResourceTypeOU {
		return false, nil, nil
	}
	ouID := security.GetOUID(ctx)
	if ouID == "" {
		return true, &AccessibleResources{AllAllowed: false, IDs: []string{}}, nil
	}
	return true, &AccessibleResources{AllAllowed: false, IDs: []string{ouID}}, nil
}

// isActionAllowedByPolicies runs all global policies against the given action context in order.
// - PolicyDecisionDenied from any policy stops the chain and denies the action.
// - PolicyDecisionNotApplicable skips to the next policy.
// - PolicyDecisionAllowed continues to the next policy.
// If all policies return NotApplicable, the action is allowed (permission check already passed).
func isActionAllowedByPolicies(ctx context.Context,
	actionCtx *ActionContext) (bool, *serviceerror.ServiceError) {
	for _, policy := range globalPolicies {
		decision, err := policy.isActionAllowed(ctx, actionCtx)
		if err != nil {
			return false, err
		}
		if decision == policyDecisionDenied {
			return false, nil
		}
	}
	return true, nil
}

// getAccessibleResourcesByPolicies iterates global policies to compute the accessible resource
// set for list operations. Policies that are not applicable for the given resource type are
// skipped. The result of the first applicable policy is returned immediately.
//
// NOTE: This is a first-applicable-wins strategy. If multiple policies need to be combined
// for the same resource type in the future, this function should be updated to intersect
// the results across all applicable policies.
func getAccessibleResourcesByPolicies(ctx context.Context, action security.Action,
	resourceType security.ResourceType) (*AccessibleResources, *serviceerror.ServiceError) {
	for _, policy := range globalPolicies {
		applicable, result, err := policy.getAccessibleResources(ctx, action, resourceType)
		if err != nil {
			return nil, err
		}
		if applicable {
			return result, nil
		}
	}
	return &AccessibleResources{AllAllowed: true}, nil
}

// globalPolicies is the ordered set of policies evaluated for every system action.
var globalPolicies = []authorizationPolicy{
	&ouMembershipPolicy{},
}
