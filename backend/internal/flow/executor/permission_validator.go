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

package executor

import (
	"context"
	"strings"

	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/security"
)

const (
	permissionValidatorLoggerComponentName = "PermissionValidator"
	defaultRequiredScope                   = "system"
)

// permissionValidator validates that the request has the required permission/scope to access the next node.
type permissionValidator struct {
	core.ExecutorInterface
	logger *log.Logger
}

// newPermissionValidator creates a new permission validator executor.
func newPermissionValidator(flowFactory core.FlowFactoryInterface) *permissionValidator {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, permissionValidatorLoggerComponentName))
	base := flowFactory.CreateExecutor(
		ExecutorNamePermissionValidator,
		common.ExecutorTypeUtility,
		[]common.Input{},
		[]common.Input{},
	)
	return &permissionValidator{
		ExecutorInterface: base,
		logger:            logger,
	}
}

// Execute validates that the request has the required permission/scope to access the next node.
func (e *permissionValidator) Execute(ctx *core.NodeContext) (*common.ExecutorResponse, error) {
	logger := e.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))

	execResp := &common.ExecutorResponse{
		Status: common.ExecComplete,
	}

	// Get required scope from node properties, default to "system"
	requiredScope := defaultRequiredScope
	if ctx.NodeProperties != nil {
		if scope, ok := ctx.NodeProperties["requiredScope"].(string); ok && scope != "" {
			requiredScope = scope
		}
	}

	logger.Debug("Checking scope protection", log.String("requiredScope", requiredScope))

	// Check if HTTP context exists
	if ctx.HTTPContext == nil {
		logger.Warn("No HTTP context available - blocking access")
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "Insufficient permissions"
		return execResp, nil
	}

	// Extract scopes from HTTP request context
	scopes := extractScopesFromHTTPContext(ctx.HTTPContext)
	logger.Debug("Extracted scopes from HTTP context",
		log.Int("scopeCount", len(scopes)),
		log.String("scopes", strings.Join(scopes, ", ")))

	// Check if required scope is present
	hasRequiredScope := false
	for _, scope := range scopes {
		if scope == requiredScope {
			hasRequiredScope = true
			break
		}
	}

	if !hasRequiredScope {
		logger.Warn("Request lacks required scope",
			log.String("requiredScope", requiredScope))
		execResp.Status = common.ExecFailure
		execResp.FailureReason = "Insufficient permissions"
		return execResp, nil
	}

	logger.Debug("Scope protection passed", log.String("requiredScope", requiredScope))

	return execResp, nil
}

// extractScopesFromHTTPContext extracts scopes/permissions from the HTTP request context.
// It checks for scope, scopes and authorized_permissions claims.
func extractScopesFromHTTPContext(httpCtx context.Context) []string {
	if scopeAttr := security.GetAttribute(httpCtx, "scope"); scopeAttr != nil {
		if scopeStr, ok := scopeAttr.(string); ok && scopeStr != "" {
			return strings.Fields(scopeStr)
		}
	}

	if scopesAttr := security.GetAttribute(httpCtx, "scopes"); scopesAttr != nil {
		if scopes, ok := scopesAttr.([]string); ok {
			return scopes
		}

		if scopesInterface, ok := scopesAttr.([]interface{}); ok {
			result := make([]string, 0, len(scopesInterface))
			for _, s := range scopesInterface {
				if str, ok := s.(string); ok {
					result = append(result, str)
				}
			}
			return result
		}
	}

	if permsAttr := security.GetAttribute(httpCtx, "authorized_permissions"); permsAttr != nil {
		if permsStr, ok := permsAttr.(string); ok && permsStr != "" {
			return strings.Fields(permsStr)
		}
	}

	return []string{}
}
