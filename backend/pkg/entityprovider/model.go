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

// Package entityprovider defines the boundary contract between the gateway layer and the directory layer.
// Gateway-layer packages (application, flow executors, roles, etc.) use EntityProviderInterface
// to communicate with the entity service without depending on directory-layer internals.
package entityprovider

import "encoding/json"

// EntityCategory represents the category of an entity (e.g., user, application, agent).
type EntityCategory string

const (
	// EntityCategoryUser represents a user entity.
	EntityCategoryUser EntityCategory = "user"
	// EntityCategoryApp represents an application entity.
	EntityCategoryApp EntityCategory = "app"
	// EntityCategoryAgent represents an agent entity.
	EntityCategoryAgent EntityCategory = "agent"
)

// String returns the string representation of the entity category.
func (ec EntityCategory) String() string {
	return string(ec)
}

// EntityState represents the lifecycle state of an entity.
type EntityState string

const (
	// EntityStateActive represents an active entity.
	EntityStateActive EntityState = "ACTIVE"
)

// String returns the string representation of the entity state.
func (es EntityState) String() string {
	return string(es)
}

// Entity represents a unified identity principal returned by the entity provider.
type Entity struct {
	ID               string          `json:"id,omitempty"`
	Category         EntityCategory  `json:"category,omitempty"`
	Type             string          `json:"type,omitempty"`
	State            EntityState     `json:"state,omitempty"`
	OUID             string          `json:"ouId,omitempty"`
	OUHandle         string          `json:"ouHandle,omitempty"`
	Attributes       json.RawMessage `json:"attributes,omitempty"`
	SystemAttributes json.RawMessage `json:"systemAttributes,omitempty"`
}

// EntityGroup represents a group with basic information for entity group membership queries.
type EntityGroup struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	OUID string `json:"ouId"`
}
