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

package entity

import (
	"fmt"
	"strings"

	"github.com/asgardeo/thunder/internal/system/config"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/log"
)

// getEntityStoreMode determines the store mode for the entity service.
//
// Resolution order:
//  1. If Entity.Store is explicitly configured, use it
//  2. Otherwise, fall back to User.Store for backward compatibility
//  3. Otherwise, fall back to global DeclarativeResources.Enabled:
//     - If enabled: return "declarative"
//     - If disabled: return "mutable"
func getEntityStoreMode() serverconst.StoreMode {
	cfg := config.GetThunderRuntime().Config

	// Check entity-level configuration first.
	if cfg.Entity.Store != "" {
		mode := serverconst.StoreMode(strings.ToLower(strings.TrimSpace(cfg.Entity.Store)))
		switch mode {
		case serverconst.StoreModeMutable, serverconst.StoreModeDeclarative, serverconst.StoreModeComposite:
			return mode
		default:
			msg := fmt.Sprintf(
				"Invalid entity store mode: %s, falling back to user store or global declarative resources setting",
				mode)
			log.GetLogger().Warn(msg)
		}
	}

	// TODO: Remove fallback to user-level configuration.
	if cfg.User.Store != "" {
		mode := serverconst.StoreMode(strings.ToLower(strings.TrimSpace(cfg.User.Store)))
		switch mode {
		case serverconst.StoreModeMutable, serverconst.StoreModeDeclarative, serverconst.StoreModeComposite:
			return mode
		default:
			msg := fmt.Sprintf(
				"Invalid user store mode: %s, falling back to global declarative resources setting", mode)
			log.GetLogger().Warn(msg)
		}
	}

	// Fall back to global declarative resources setting.
	if declarativeresource.IsDeclarativeModeEnabled() {
		return serverconst.StoreModeDeclarative
	}

	return serverconst.StoreModeMutable
}

// getIndexedAttributes reads indexed attributes from entity config,
// falling back to user config for temporary backward compatibility.
func getIndexedAttributes() []string {
	cfg := config.GetThunderRuntime().Config

	if len(cfg.Entity.IndexedAttributes) > 0 {
		return cfg.Entity.IndexedAttributes
	}

	return cfg.User.IndexedAttributes
}
