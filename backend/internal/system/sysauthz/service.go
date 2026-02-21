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

// Package sysauthz provides system-level authorization services for protecting
// system operations management.
package sysauthz

import (
	"context"

	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/security"
)

// SystemAuthorizationServiceInterface defines the contract for system-level authorization.
type SystemAuthorizationServiceInterface interface {
	// IsActionAllowed checks whether the authenticated caller is permitted to perform
	// the given action. Returns true if allowed, false if denied. A non-nil ServiceError
	// indicates a processing failure, not an authorization denial.
	IsActionAllowed(ctx context.Context, action security.Action,
		actionCtx *ActionContext) (bool, *serviceerror.ServiceError)

	// GetAccessibleResources returns the set of resources the caller may access for the
	// given action and resource type. The result must be applied as a store-level filter
	// before pagination so that page sizes and total counts remain correct.
	//
	// When AllAllowed is true, no ID filter should be applied.
	// When AllAllowed is false, the store should restrict results to the returned IDs.
	GetAccessibleResources(ctx context.Context, action security.Action,
		resourceType security.ResourceType) (*AccessibleResources, *serviceerror.ServiceError)
}

// systemAuthorizationService is the default implementation of SystemAuthorizationServiceInterface.
type systemAuthorizationService struct {
	logger *log.Logger
}

// newSystemAuthorizationService returns a new systemAuthorizationService.
func newSystemAuthorizationService() SystemAuthorizationServiceInterface {
	return &systemAuthorizationService{
		logger: log.GetLogger().With(log.String("component", "SystemAuthorizationService")),
	}
}

// IsActionAllowed evaluates whether the authenticated caller may perform the given action.
//
// Decision logic (in order):
//  1. If the THUNDER_SKIP_SECURITY flag is enabled, access is unconditionally granted.
//  2. If the context is an internal runtime caller (WithRuntimeContext), access is granted.
//  3. The caller must be authenticated (non-empty subject in the security context).
//  4. If the caller holds the "system" permission, access is unconditionally granted (policies skipped).
//  5. The required permission for the action is resolved via security.ResolveActionPermission.
//     Unmapped actions fall back to requiring the "system" permission.
//     The caller must satisfy the required permission via hierarchical scope matching.
//  6. Global policies are evaluated for the given action context.
func (s *systemAuthorizationService) IsActionAllowed(ctx context.Context, action security.Action,
	actionCtx *ActionContext) (bool, *serviceerror.ServiceError) {
	logger := s.logger.WithContext(ctx)

	// Step 1: Check if SKIP_SECURITY flag is set.
	if security.IsSecuritySkipped(ctx) {
		logger.Debug("Authorization skipped: THUNDER_SKIP_SECURITY is enabled",
			log.String("action", string(action)))
		return true, nil
	}

	// Step 2: Check if this is an internal runtime caller.
	if security.IsRuntimeContext(ctx) {
		logger.Debug("Authorization granted: runtime context for the action",
			log.String("action", string(action)))
		return true, nil
	}

	// Step 3: Verify the caller is authenticated.
	subject := security.GetSubject(ctx)
	if subject == "" {
		logger.Debug("Authorization denied: unauthenticated caller",
			log.String("action", string(action)))
		return false, nil
	}

	permissions := security.GetPermissions(ctx)

	// Step 4: Short-circuit: the "system" permission grants access to all system operations.
	if security.HasSystemPermission(permissions) {
		return true, nil
	}

	// Step 5: Resolve required permission for the action and evaluate using hierarchical matching.
	requiredPermission := security.ResolveActionPermission(action)
	if !security.HasSufficientPermission(permissions, requiredPermission) {
		if logger.IsDebugEnabled() {
			logger.Debug("Authorization denied: insufficient permissions",
				log.String("action", string(action)),
				log.String("subject", log.MaskString(subject)))
		}
		return false, nil
	}

	// Step 6: Evaluate global policies (e.g., OU scope check).
	allowed, svcErr := isActionAllowedByPolicies(ctx, actionCtx)
	if svcErr != nil {
		return false, svcErr
	}
	if !allowed {
		if logger.IsDebugEnabled() {
			logger.Debug("Authorization denied: policy evaluation failed",
				log.String("action", string(action)),
				log.String("subject", log.MaskString(subject)))
		}
		return false, nil
	}

	if logger.IsDebugEnabled() {
		logger.Debug("Authorization granted",
			log.String("action", string(action)),
			log.String("subject", log.MaskString(subject)))
	}

	return true, nil
}

// GetAccessibleResources returns the set of resources the caller can access for the given
// action and resource type.
//
// Decision logic (in order):
//  1. If the THUNDER_SKIP_SECURITY flag is enabled, all resources are accessible.
//  2. If the context is an internal runtime caller (WithRuntimeContext), all resources are accessible.
//  3. The caller must be authenticated (non-empty subject in the security context).
//  4. If the caller holds the "system" permission, all resources are accessible.
//  5. The required permissions for the action are resolved via security.ResolveActionPermission.
//     Unmapped actions fall back to requiring the "system" permission.
//     The caller must hold at least one of the required permissions.
//  6. Global policies are evaluated to determine the accessible resource set for the caller.
func (s *systemAuthorizationService) GetAccessibleResources(ctx context.Context, action security.Action,
	resourceType security.ResourceType) (*AccessibleResources, *serviceerror.ServiceError) {
	logger := s.logger.WithContext(ctx)

	// Step 1: Check if SKIP_SECURITY flag is set.
	if security.IsSecuritySkipped(ctx) {
		logger.Debug("GetAccessibleResources skipped: THUNDER_SKIP_SECURITY is enabled",
			log.String("action", string(action)),
			log.String("resourceType", string(resourceType)))
		return &AccessibleResources{AllAllowed: true}, nil
	}

	// Step 2: Check if this is an internal runtime caller — return all resources.
	if security.IsRuntimeContext(ctx) {
		logger.Debug("GetAccessibleResources: runtime context, returning all resources",
			log.String("action", string(action)),
			log.String("resourceType", string(resourceType)))
		return &AccessibleResources{AllAllowed: true}, nil
	}

	// Step 3: Verify the caller is authenticated.
	subject := security.GetSubject(ctx)
	if subject == "" {
		logger.Debug("GetAccessibleResources denied: unauthenticated caller",
			log.String("action", string(action)),
			log.String("resourceType", string(resourceType)))
		return &AccessibleResources{AllAllowed: false, IDs: []string{}}, nil
	}

	permissions := security.GetPermissions(ctx)

	// Step 4: Short-circuit: the "system" permission grants access to all resources.
	if security.HasSystemPermission(permissions) {
		return &AccessibleResources{AllAllowed: true}, nil
	}

	// Step 5: Verify the caller holds an adequate permission for the action using hierarchical matching.
	requiredPermission := security.ResolveActionPermission(action)
	if !security.HasSufficientPermission(permissions, requiredPermission) {
		if logger.IsDebugEnabled() {
			logger.Debug("GetAccessibleResources denied: insufficient permissions",
				log.String("action", string(action)),
				log.String("resourceType", string(resourceType)),
				log.String("subject", log.MaskString(subject)))
		}
		return &AccessibleResources{AllAllowed: false, IDs: []string{}}, nil
	}

	// Step 6: Delegate to the policy chain to determine the accessible resource set.
	result, svcErr := getAccessibleResourcesByPolicies(ctx, action, resourceType)
	if svcErr != nil {
		return nil, svcErr
	}
	if logger.IsDebugEnabled() && !result.AllAllowed {
		logger.Debug("GetAccessibleResources: restricted by policy",
			log.String("action", string(action)),
			log.String("resourceType", string(resourceType)),
			log.String("subject", log.MaskString(subject)),
			log.Int("accessibleCount", len(result.IDs)))
	}
	return result, nil
}

// Permission resolution and matching logic lives in the security package:
// - security.ResolveActionPermission(action) — maps Action → required permission
// - security.HasSystemPermission(permissions) — fast-path check for root scope
// - security.HasSufficientPermission(permissions, required) — hierarchical scope match
