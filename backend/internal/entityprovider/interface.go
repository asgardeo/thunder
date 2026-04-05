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

package entityprovider

import (
	"encoding/json"
)

// EntityProviderInterface defines the boundary contract between the gateway layer and the
// directory layer for entity operations. Gateway-layer packages (application, flow executors,
// roles, groups, OAuth) use this interface to interact with entities without depending on
// directory-layer internals.
type EntityProviderInterface interface {
	// IdentifyEntity resolves an entity ID from indexed attribute filters (e.g., email, clientId).
	IdentifyEntity(filters map[string]interface{}) (*string, *EntityProviderError)

	// GetEntity retrieves an entity by ID. Credentials are never returned.
	GetEntity(entityID string) (*Entity, *EntityProviderError)

	// CreateEntity creates a new entity. Entity Attributes may include credential fields
	// which are extracted and stored separately. SystemCredentials (e.g., clientSecret) are
	// hashed and stored in the SYSTEM_CREDENTIALS column.
	CreateEntity(entity *Entity,
		systemCredentials json.RawMessage) (*Entity, *EntityProviderError)

	// UpdateEntity updates an existing entity's core fields.
	UpdateEntity(entityID string, entity *Entity) (*Entity, *EntityProviderError)

	// DeleteEntity deletes an entity by ID. Cascades to identifiers.
	DeleteEntity(entityID string) *EntityProviderError

	// UpdateSystemAttributes updates system-managed attributes for an entity.
	UpdateSystemAttributes(entityID string,
		attributes json.RawMessage) *EntityProviderError

	// UpdateSystemCredentials updates system-managed credentials for an entity.
	UpdateSystemCredentials(entityID string,
		credentials json.RawMessage) *EntityProviderError

	// GetEntityGroups retrieves the groups an entity belongs to with pagination.
	GetEntityGroups(entityID string,
		limit, offset int) (*EntityGroupListResponse, *EntityProviderError)

	// GetTransitiveEntityGroups retrieves all groups an entity belongs to, including inherited groups.
	GetTransitiveEntityGroups(entityID string) ([]EntityGroup, *EntityProviderError)

	// ValidateEntityIDs validates that the given entity IDs exist. Returns IDs that are invalid.
	ValidateEntityIDs(entityIDs []string) ([]string, *EntityProviderError)

	// GetEntitiesByIDs retrieves multiple entities by their IDs.
	GetEntitiesByIDs(entityIDs []string) ([]Entity, *EntityProviderError)
}
