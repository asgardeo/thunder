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

package ou

import (
	"strings"

	"github.com/asgardeo/thunder/internal/system/config"
	immutableresource "github.com/asgardeo/thunder/internal/system/immutable_resource"
)

// Store mode constants for organization unit service.

// getOrganizationUnitStoreMode determines the store mode for organization units.
//
// Resolution order:
//  1. If OrganizationUnit.Store is explicitly configured, use it
//  2. Otherwise, fall back to global ImmutableResources.Enabled:
//     - If enabled: return "immutable"
//     - If disabled: return "mutable"
//
// Returns normalized store mode: "mutable", "immutable", or "composite"
func getOrganizationUnitStoreMode() string {
	cfg := config.GetThunderRuntime().Config
	// Check if service-level configuration is explicitly set
	if cfg.OrganizationUnit.Store != "" {
		mode := strings.ToLower(strings.TrimSpace(cfg.OrganizationUnit.Store))
		// Validate and normalize
		switch mode {
		case config.StoreModeMutable, config.StoreModeImmutable, config.StoreModeComposite:
			return mode
		}
	}

	// Fall back to global immutable resources setting
	if immutableresource.IsImmutableModeEnabled() {
		return config.StoreModeImmutable
	}

	return config.StoreModeMutable
}

// isCompositeModeEnabled checks if composite store mode is enabled for organization units.
func isCompositeModeEnabled() bool {
	return getOrganizationUnitStoreMode() == config.StoreModeComposite
}

// isMutableModeEnabled checks if mutable-only store mode is enabled for organization units.
func isMutableModeEnabled() bool {
	return getOrganizationUnitStoreMode() == config.StoreModeMutable
}

// isImmutableModeEnabled checks if immutable-only store mode is enabled for organization units.
func isImmutableModeEnabled() bool {
	return getOrganizationUnitStoreMode() == config.StoreModeImmutable
}
