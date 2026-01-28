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

package engine

import "context"

// AuthorizationEngine is the internal interface for authorization engines.
// This interface is NOT exported and is used internally by the authorization service.
// Different engines can be plugged in (RBAC, ABAC, ReBAC, Custom) by implementing this interface.
type AuthorizationEngine interface {
	// GetAuthorizedPermissions returns the subset of requested permissions
	// that the user (directly or through groups) is authorized for.
	GetAuthorizedPermissions(
		ctx context.Context,
		userID string,
		groupIDs []string,
		requestedPermissions []string,
	) ([]string, error)
}
