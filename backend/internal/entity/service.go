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
	"context"
	"encoding/json"
	"errors"

	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/transaction"
)

// EntityServiceInterface is the interface for managing entities.
type EntityServiceInterface interface {
	// Core CRUD
	CreateEntity(ctx context.Context, entity *Entity,
		credentials json.RawMessage, systemCredentials json.RawMessage) (*Entity, error)
	GetEntity(ctx context.Context, entityID string) (*Entity, error)
	GetEntityWithCredentials(ctx context.Context, entityID string) (
		*Entity, json.RawMessage, json.RawMessage, error)
	UpdateEntity(ctx context.Context, entityID string, entity *Entity) (*Entity, error)
	UpdateEntityWithCredentials(ctx context.Context, entityID string,
		entity *Entity, systemCreds json.RawMessage) (*Entity, error)
	DeleteEntity(ctx context.Context, entityID string) error

	// Partial updates
	UpdateAttributes(ctx context.Context, entityID string, attributes json.RawMessage) error
	UpdateSystemAttributes(ctx context.Context, entityID string, attrs json.RawMessage) error
	UpdateCredentials(ctx context.Context, entityID string, creds json.RawMessage) error
	UpdateSystemCredentials(ctx context.Context, entityID string, creds json.RawMessage) error

	// Identification
	IdentifyEntity(ctx context.Context, filters map[string]interface{}) (*string, error)

	// Lists (category-scoped)
	GetEntityListCount(ctx context.Context, category EntityCategory,
		filters map[string]interface{}) (int, error)
	GetEntityList(ctx context.Context, category EntityCategory,
		limit, offset int, filters map[string]interface{}) ([]Entity, error)
	GetEntityListCountByOUIDs(ctx context.Context, category EntityCategory,
		ouIDs []string, filters map[string]interface{}) (int, error)
	GetEntityListByOUIDs(ctx context.Context, category EntityCategory,
		ouIDs []string, limit, offset int, filters map[string]interface{}) ([]Entity, error)

	// Bulk
	ValidateEntityIDs(ctx context.Context, entityIDs []string) ([]string, error)
	GetEntitiesByIDs(ctx context.Context, entityIDs []string) ([]Entity, error)
	ValidateEntityIDsInOUs(ctx context.Context, entityIDs []string, ouIDs []string) ([]string, error)

	// Groups
	GetGroupCountForEntity(ctx context.Context, entityID string) (int, error)
	GetEntityGroups(ctx context.Context, entityID string, limit, offset int) ([]EntityGroup, error)
	GetTransitiveEntityGroups(ctx context.Context, entityID string) ([]EntityGroup, error)

	// Declarative
	IsEntityDeclarative(ctx context.Context, entityID string) (bool, error)
	LoadDeclarativeResources(config DeclarativeLoaderConfig) error
}

// entityService is the default implementation of EntityServiceInterface.
type entityService struct {
	store         entityStoreInterface
	transactioner transaction.Transactioner
	logger        *log.Logger
}

// newEntityService creates a new entity service.
func newEntityService(
	store entityStoreInterface,
	transactioner transaction.Transactioner,
) EntityServiceInterface {
	return &entityService{
		store:         store,
		transactioner: transactioner,
		logger:        log.GetLogger().With(log.String(log.LoggerKeyComponentName, "EntityService")),
	}
}

// CreateEntity creates a new entity.
// Uses a transaction to ensure the entity row and its indexed identifiers are created atomically.
func (s *entityService) CreateEntity(ctx context.Context, entity *Entity,
	credentials json.RawMessage, systemCredentials json.RawMessage) (*Entity, error) {
	if entity == nil {
		return nil, ErrEntityNotFound
	}
	s.logger.Debug("Creating entity", log.String("id", entity.ID))

	var created Entity
	err := s.transactioner.Transact(ctx, func(txCtx context.Context) error {
		if err := s.store.CreateEntity(txCtx, *entity, credentials, systemCredentials); err != nil {
			return err
		}

		result, err := s.store.GetEntity(txCtx, entity.ID)
		if err != nil {
			return err
		}
		created = result
		return nil
	})
	if err != nil {
		s.logger.Error("Failed to create entity", log.String("id", entity.ID), log.Error(err))
		return nil, err
	}

	return &created, nil
}

// GetEntity retrieves an entity by ID.
func (s *entityService) GetEntity(ctx context.Context, entityID string) (*Entity, error) {
	entity, err := s.store.GetEntity(ctx, entityID)
	if err != nil {
		if !errors.Is(err, ErrEntityNotFound) {
			s.logger.Error("Failed to get entity", log.String("id", entityID), log.Error(err))
		}
		return nil, err
	}
	return &entity, nil
}

// GetEntityWithCredentials retrieves an entity with all credential columns.
func (s *entityService) GetEntityWithCredentials(ctx context.Context, entityID string) (
	*Entity, json.RawMessage, json.RawMessage, error) {
	entity, creds, sysCreds, err := s.store.GetEntityWithCredentials(ctx, entityID)
	if err != nil {
		if !errors.Is(err, ErrEntityNotFound) {
			s.logger.Error("Failed to get entity with credentials", log.String("id", entityID), log.Error(err))
		}
		return nil, nil, nil, err
	}
	return &entity, creds, sysCreds, nil
}

// UpdateEntity updates an entity.
// Uses a transaction to ensure the entity update and identifier re-sync are atomic.
func (s *entityService) UpdateEntity(ctx context.Context, entityID string, entity *Entity) (*Entity, error) {
	if entity == nil {
		return nil, ErrEntityNotFound
	}
	s.logger.Debug("Updating entity", log.String("id", entityID))

	var updated Entity
	err := s.transactioner.Transact(ctx, func(txCtx context.Context) error {
		entity.ID = entityID
		if err := s.store.UpdateEntity(txCtx, entity); err != nil {
			return err
		}

		result, err := s.store.GetEntity(txCtx, entityID)
		if err != nil {
			return err
		}
		updated = result
		return nil
	})
	if err != nil {
		if !errors.Is(err, ErrEntityNotFound) {
			s.logger.Error("Failed to update entity", log.String("id", entityID), log.Error(err))
		}
		return nil, err
	}

	return &updated, nil
}

// UpdateEntityWithCredentials atomically updates an entity and its system credentials.
// If systemCreds is nil, only the entity is updated.
func (s *entityService) UpdateEntityWithCredentials(ctx context.Context, entityID string,
	entity *Entity, systemCreds json.RawMessage) (*Entity, error) {
	if entity == nil {
		return nil, ErrEntityNotFound
	}
	s.logger.Debug("Updating entity with credentials", log.String("id", entityID))

	var updated Entity
	err := s.transactioner.Transact(ctx, func(txCtx context.Context) error {
		entity.ID = entityID
		if err := s.store.UpdateEntity(txCtx, entity); err != nil {
			return err
		}

		if systemCreds != nil {
			if err := s.store.UpdateSystemCredentials(txCtx, entityID, systemCreds); err != nil {
				return err
			}
		}

		result, err := s.store.GetEntity(txCtx, entityID)
		if err != nil {
			return err
		}
		updated = result
		return nil
	})
	if err != nil {
		if !errors.Is(err, ErrEntityNotFound) {
			s.logger.Error("Failed to update entity with credentials", log.String("id", entityID), log.Error(err))
		}
		return nil, err
	}
	return &updated, nil
}

// DeleteEntity deletes an entity.
// Uses a transaction to ensure the entity row and its indexed identifiers are deleted atomically.
func (s *entityService) DeleteEntity(ctx context.Context, entityID string) error {
	s.logger.Debug("Deleting entity", log.String("id", entityID))
	err := s.transactioner.Transact(ctx, func(txCtx context.Context) error {
		return s.store.DeleteEntity(txCtx, entityID)
	})
	if err != nil {
		if !errors.Is(err, ErrEntityNotFound) {
			s.logger.Error("Failed to delete entity", log.String("id", entityID), log.Error(err))
		}
		return err
	}
	return nil
}

// UpdateAttributes updates only the schema attributes of an entity.
func (s *entityService) UpdateAttributes(ctx context.Context, entityID string, attributes json.RawMessage) error {
	s.logger.Debug("Updating entity attributes", log.String("id", entityID))
	err := s.transactioner.Transact(ctx, func(txCtx context.Context) error {
		return s.store.UpdateAttributes(txCtx, entityID, attributes)
	})
	if err != nil {
		if !errors.Is(err, ErrEntityNotFound) {
			s.logger.Error("Failed to update entity attributes", log.String("id", entityID), log.Error(err))
		}
		return err
	}
	return nil
}

// UpdateSystemAttributes updates the system-managed attributes of an entity.
func (s *entityService) UpdateSystemAttributes(ctx context.Context, entityID string,
	attrs json.RawMessage) error {
	s.logger.Debug("Updating entity system attributes", log.String("id", entityID))
	err := s.transactioner.Transact(ctx, func(txCtx context.Context) error {
		return s.store.UpdateSystemAttributes(txCtx, entityID, attrs)
	})
	if err != nil {
		if !errors.Is(err, ErrEntityNotFound) {
			s.logger.Error("Failed to update entity system attributes", log.String("id", entityID), log.Error(err))
		}
		return err
	}
	return nil
}

// UpdateCredentials updates the schema-defined credentials of an entity.
func (s *entityService) UpdateCredentials(ctx context.Context, entityID string,
	creds json.RawMessage) error {
	err := s.store.UpdateCredentials(ctx, entityID, creds)
	if err != nil {
		if !errors.Is(err, ErrEntityNotFound) {
			s.logger.Error("Failed to update entity credentials", log.String("id", entityID), log.Error(err))
		}
		return err
	}
	return nil
}

// UpdateSystemCredentials updates the system credentials of an entity.
func (s *entityService) UpdateSystemCredentials(ctx context.Context, entityID string,
	creds json.RawMessage) error {
	err := s.store.UpdateSystemCredentials(ctx, entityID, creds)
	if err != nil {
		if !errors.Is(err, ErrEntityNotFound) {
			s.logger.Error("Failed to update entity system credentials", log.String("id", entityID), log.Error(err))
		}
		return err
	}
	return nil
}

// IdentifyEntity identifies an entity using the given filters.
func (s *entityService) IdentifyEntity(ctx context.Context,
	filters map[string]interface{}) (*string, error) {
	id, err := s.store.IdentifyEntity(ctx, filters)
	if err != nil {
		if !errors.Is(err, ErrEntityNotFound) {
			s.logger.Error("Failed to identify entity", log.Error(err))
		}
		return nil, err
	}
	return id, nil
}

// GetEntityListCount retrieves the total count of entities by category.
func (s *entityService) GetEntityListCount(ctx context.Context, category EntityCategory,
	filters map[string]interface{}) (int, error) {
	count, err := s.store.GetEntityListCount(ctx, string(category), filters)
	if err != nil {
		s.logger.Error("Failed to get entity list count", log.Error(err))
		return 0, err
	}
	return count, nil
}

// GetEntityList retrieves a list of entities by category.
func (s *entityService) GetEntityList(ctx context.Context, category EntityCategory,
	limit, offset int, filters map[string]interface{}) ([]Entity, error) {
	entities, err := s.store.GetEntityList(ctx, string(category), limit, offset, filters)
	if err != nil {
		s.logger.Error("Failed to get entity list", log.Error(err))
		return nil, err
	}
	return entities, nil
}

// GetEntityListCountByOUIDs retrieves the total count of entities scoped to OU IDs.
func (s *entityService) GetEntityListCountByOUIDs(ctx context.Context, category EntityCategory,
	ouIDs []string, filters map[string]interface{}) (int, error) {
	count, err := s.store.GetEntityListCountByOUIDs(ctx, string(category), ouIDs, filters)
	if err != nil {
		s.logger.Error("Failed to get entity list count by OU IDs", log.Error(err))
		return 0, err
	}
	return count, nil
}

// GetEntityListByOUIDs retrieves a list of entities scoped to OU IDs.
func (s *entityService) GetEntityListByOUIDs(ctx context.Context, category EntityCategory,
	ouIDs []string, limit, offset int, filters map[string]interface{}) ([]Entity, error) {
	entities, err := s.store.GetEntityListByOUIDs(ctx, string(category), ouIDs, limit, offset, filters)
	if err != nil {
		s.logger.Error("Failed to get entity list by OU IDs", log.Error(err))
		return nil, err
	}
	return entities, nil
}

// ValidateEntityIDs checks if all provided entity IDs exist.
func (s *entityService) ValidateEntityIDs(ctx context.Context, entityIDs []string) ([]string, error) {
	invalid, err := s.store.ValidateEntityIDs(ctx, entityIDs)
	if err != nil {
		s.logger.Error("Failed to validate entity IDs", log.Error(err))
		return nil, err
	}
	return invalid, nil
}

// GetEntitiesByIDs retrieves entities by a list of IDs.
func (s *entityService) GetEntitiesByIDs(ctx context.Context, entityIDs []string) ([]Entity, error) {
	entities, err := s.store.GetEntitiesByIDs(ctx, entityIDs)
	if err != nil {
		s.logger.Error("Failed to get entities by IDs", log.Error(err))
		return nil, err
	}
	return entities, nil
}

// ValidateEntityIDsInOUs checks which of the provided entity IDs belong to the given OU scope.
func (s *entityService) ValidateEntityIDsInOUs(ctx context.Context,
	entityIDs []string, ouIDs []string) ([]string, error) {
	invalid, err := s.store.ValidateEntityIDsInOUs(ctx, entityIDs, ouIDs)
	if err != nil {
		s.logger.Error("Failed to validate entity IDs in OUs", log.Error(err))
		return nil, err
	}
	return invalid, nil
}

// GetGroupCountForEntity retrieves the total count of groups an entity belongs to.
func (s *entityService) GetGroupCountForEntity(ctx context.Context, entityID string) (int, error) {
	count, err := s.store.GetGroupCountForEntity(ctx, entityID)
	if err != nil {
		s.logger.Error("Failed to get group count for entity", log.String("id", entityID), log.Error(err))
		return 0, err
	}
	return count, nil
}

// GetEntityGroups retrieves groups that an entity belongs to with pagination.
func (s *entityService) GetEntityGroups(ctx context.Context, entityID string,
	limit, offset int) ([]EntityGroup, error) {
	groups, err := s.store.GetEntityGroups(ctx, entityID, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get entity groups", log.String("id", entityID), log.Error(err))
		return nil, err
	}
	return groups, nil
}

// GetTransitiveEntityGroups retrieves all groups an entity belongs to, including nested group membership.
func (s *entityService) GetTransitiveEntityGroups(ctx context.Context, entityID string) ([]EntityGroup, error) {
	groups, err := s.store.GetTransitiveEntityGroups(ctx, entityID)
	if err != nil {
		s.logger.Error("Failed to get transitive entity groups", log.String("id", entityID), log.Error(err))
		return nil, err
	}
	return groups, nil
}

// IsEntityDeclarative checks if an entity is declarative (immutable).
func (s *entityService) IsEntityDeclarative(ctx context.Context, entityID string) (bool, error) {
	ok, err := s.store.IsEntityDeclarative(ctx, entityID)
	if err != nil {
		s.logger.Error("Failed to check if entity is declarative", log.String("id", entityID), log.Error(err))
		return false, err
	}
	return ok, nil
}

// LoadDeclarativeResources loads declarative resources for a given entity category.
// Consumer packages provide parser/validator callbacks for type-specific YAML processing.
func (s *entityService) LoadDeclarativeResources(config DeclarativeLoaderConfig) error {
	if err := loadDeclarativeResources(s.store, s, config); err != nil {
		s.logger.Error("Failed to load declarative resources", log.Error(err))
		return err
	}
	return nil
}
