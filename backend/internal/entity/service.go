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
	"fmt"
	"strings"

	"github.com/asgardeo/thunder/internal/system/crypto/hash"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/transaction"
	"github.com/asgardeo/thunder/internal/userschema"
)

// EntityServiceInterface is the interface for managing entities.
type EntityServiceInterface interface {
	// Core CRUD
	CreateEntity(ctx context.Context, entity *Entity,
		systemCredentials json.RawMessage) (*Entity, error)
	GetEntity(ctx context.Context, entityID string) (*Entity, error)
	GetEntityWithCredentials(ctx context.Context, entityID string) (
		*Entity, json.RawMessage, json.RawMessage, error)
	UpdateEntity(ctx context.Context, entityID string, entity *Entity) (*Entity, error)
	DeleteEntity(ctx context.Context, entityID string) error

	// Partial updates
	UpdateAttributes(ctx context.Context, entityID string, attributes json.RawMessage) error
	UpdateSystemAttributes(ctx context.Context, entityID string, attrs json.RawMessage) error
	UpdateSystemCredentials(ctx context.Context, entityID string,
		plaintextUpdates json.RawMessage) error

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

	// Authentication
	AuthenticateEntity(ctx context.Context, identifiers map[string]interface{},
		credentials map[string]interface{}) (*AuthenticateResult, error)

	// Declarative
	IsEntityDeclarative(ctx context.Context, entityID string) (bool, error)
	LoadDeclarativeResources(config DeclarativeLoaderConfig) error

	// Config
	LoadIndexedAttributes(attributes []string) error
}

// entityService is the default implementation of EntityServiceInterface.
type entityService struct {
	store             entityStoreInterface
	hashService       hash.HashServiceInterface
	userSchemaService userschema.UserSchemaServiceInterface
	transactioner     transaction.Transactioner
	logger            *log.Logger
}

// newEntityService creates a new entity service.
func newEntityService(
	store entityStoreInterface,
	hashService hash.HashServiceInterface,
	userSchemaService userschema.UserSchemaServiceInterface,
	transactioner transaction.Transactioner,
) EntityServiceInterface {
	return &entityService{
		store:             store,
		hashService:       hashService,
		userSchemaService: userSchemaService,
		transactioner:     transactioner,
		logger:            log.GetLogger().With(log.String(log.LoggerKeyComponentName, "EntityService")),
	}
}

// CreateEntity creates a new entity.
// Uses a transaction to ensure the entity row and its indexed identifiers are created atomically.
func (s *entityService) CreateEntity(ctx context.Context, entity *Entity,
	systemCredentials json.RawMessage) (*Entity, error) {
	if entity == nil {
		return nil, ErrEntityNotFound
	}
	s.logger.Debug("Creating entity", log.String("id", entity.ID))

	// Validate entity attributes and uniqueness via schema.
	if err := s.validateEntitySchema(ctx, entity.Category, entity.Type, entity.Attributes, "", false); err != nil {
		return nil, err
	}

	// Extract schema-defined credential fields from Attributes.
	schemaCredsJSON, err := s.extractAndHashSchemaCredentials(ctx, entity)
	if err != nil {
		s.logger.Error("Failed to extract schema credentials", log.Error(err))
		return nil, fmt.Errorf("failed to extract schema credentials: %w", err)
	}

	// Hash plaintext system credentials.
	hashedSysCreds, err := s.hashPlaintextCredentials(systemCredentials)
	if err != nil {
		s.logger.Error("Failed to hash system credentials", log.Error(err))
		return nil, fmt.Errorf("failed to hash system credentials: %w", err)
	}

	var created Entity
	err = s.transactioner.Transact(ctx, func(txCtx context.Context) error {
		if err := s.store.CreateEntity(txCtx, *entity, schemaCredsJSON, hashedSysCreds); err != nil {
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

	// Validate entity attributes and uniqueness via schema (excludes self for uniqueness).
	if err := s.validateEntitySchema(ctx, entity.Category, entity.Type, entity.Attributes, entityID, true); err != nil {
		return nil, err
	}

	// Extract schema credentials from attributes.
	// These will be merged with existing credentials atomically.
	schemaCredsJSON, err := s.extractAndHashSchemaCredentials(ctx, entity)
	if err != nil {
		s.logger.Error("Failed to extract schema credentials for update", log.Error(err))
		return nil, fmt.Errorf("failed to extract schema credentials: %w", err)
	}

	var updated Entity
	err = s.transactioner.Transact(ctx, func(txCtx context.Context) error {
		entity.ID = entityID
		if err := s.store.UpdateEntity(txCtx, entity); err != nil {
			return err
		}

		// Merge extracted schema credentials with existing credentials.
		if len(schemaCredsJSON) > 0 {
			_, existingSchemaCredsJSON, _, getErr := s.store.GetEntityWithCredentials(txCtx, entityID)
			if getErr != nil {
				return getErr
			}

			mergedCreds := mergeCredentialJSON(existingSchemaCredsJSON, schemaCredsJSON)
			if err := s.store.UpdateCredentials(txCtx, entityID, mergedCreds); err != nil {
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
			s.logger.Error("Failed to update entity", log.String("id", entityID), log.Error(err))
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

// AuthenticateEntity authenticates an entity by combining identify and verify operations.
// Identifiers are used to find the entity, and credentials are verified against stored credentials.
func (s *entityService) AuthenticateEntity(
	ctx context.Context,
	identifiers map[string]interface{},
	credentials map[string]interface{},
) (*AuthenticateResult, error) {
	if len(identifiers) == 0 {
		return nil, ErrEntityNotFound
	}
	if len(credentials) == 0 {
		return nil, ErrAuthenticationFailed
	}

	entityID, err := s.IdentifyEntity(ctx, identifiers)
	if err != nil {
		return nil, err
	}

	e, schemaCredsJSON, systemCredsJSON, err := s.GetEntityWithCredentials(ctx, *entityID)
	if err != nil {
		return nil, err
	}

	if e.State != EntityStateActive {
		return nil, ErrEntityNotFound
	}

	if err := s.verifyCredentials(*entityID, credentials, schemaCredsJSON, systemCredsJSON); err != nil {
		return nil, err
	}

	return &AuthenticateResult{
		EntityID:           e.ID,
		EntityCategory:     e.Category,
		EntityType:         e.Type,
		OrganizationUnitID: e.OrganizationUnitID,
	}, nil
}

// verifyCredentials verifies provided credentials from both schema and system credentials.
func (s *entityService) verifyCredentials(entityID string, credentials map[string]interface{},
	schemaCredsJSON, systemCredsJSON json.RawMessage) error {
	// Merge both credential columns for verification.
	storedCreds := make(map[string][]storedCredential)
	if len(schemaCredsJSON) > 0 {
		var schemaCreds map[string][]storedCredential
		if err := json.Unmarshal(schemaCredsJSON, &schemaCreds); err != nil {
			s.logger.Error("Failed to unmarshal schema credentials", log.String("id", entityID), log.Error(err))
			return fmt.Errorf("failed to unmarshal schema credentials: %w", err)
		}
		for k, v := range schemaCreds {
			storedCreds[k] = v
		}
	}
	if len(systemCredsJSON) > 0 {
		var sysCreds map[string][]storedCredential
		if err := json.Unmarshal(systemCredsJSON, &sysCreds); err != nil {
			s.logger.Error("Failed to unmarshal system credentials", log.String("id", entityID), log.Error(err))
			return fmt.Errorf("failed to unmarshal system credentials: %w", err)
		}
		for k, v := range sysCreds {
			storedCreds[k] = v
		}
	}

	if len(storedCreds) == 0 {
		return ErrAuthenticationFailed
	}

	// Filter to credentials that have stored entries.
	credentialsToVerify := make(map[string]string)
	for credType, credValueInterface := range credentials {
		if _, exists := storedCreds[credType]; !exists {
			continue
		}
		credValue, ok := credValueInterface.(string)
		if !ok || credValue == "" {
			continue
		}
		credentialsToVerify[credType] = credValue
	}

	if len(credentialsToVerify) == 0 {
		return ErrAuthenticationFailed
	}

	// Verify each credential against stored values.
	for credType, credValue := range credentialsToVerify {
		credList := storedCreds[credType]
		verified := false
		for _, stored := range credList {
			ref := hash.Credential{
				Algorithm: stored.StorageAlgo,
				Hash:      stored.Value,
				Parameters: hash.CredParameters{
					Salt:       stored.StorageAlgoParams.Salt,
					Iterations: stored.StorageAlgoParams.Iterations,
					KeySize:    stored.StorageAlgoParams.KeySize,
				},
			}
			ok, verifyErr := s.hashService.Verify([]byte(credValue), ref)
			if verifyErr == nil && ok {
				verified = true
				break
			}
		}
		if !verified {
			return ErrAuthenticationFailed
		}
	}

	return nil
}

// UpdateSystemCredentials updates system credentials by hashing new plaintext values and
// merging with existing stored credentials. Existing credential types not in the update
// are preserved.
func (s *entityService) UpdateSystemCredentials(ctx context.Context, entityID string,
	plaintextUpdates json.RawMessage) error {
	if len(plaintextUpdates) == 0 {
		return nil
	}

	// Parse and validate new credential updates.
	var updates map[string]interface{}
	if err := json.Unmarshal(plaintextUpdates, &updates); err != nil {
		return fmt.Errorf("%w: failed to parse credentials", ErrInvalidCredential)
	}

	for credType, credValue := range updates {
		switch v := credValue.(type) {
		case string:
			if strings.TrimSpace(v) == "" {
				return fmt.Errorf("%w: empty value for credential type %q", ErrInvalidCredential, credType)
			}
		case []interface{}:
			// Structured credentials (e.g., passkey objects) — validate non-empty.
			if len(v) == 0 {
				return fmt.Errorf("%w: empty array for credential type %q", ErrInvalidCredential, credType)
			}
		case nil:
			return fmt.Errorf("%w: nil value for credential type %q", ErrInvalidCredential, credType)
		}
	}

	// Hash new plaintext values.
	hashedUpdates, err := s.hashPlaintextCredentials(plaintextUpdates)
	if err != nil {
		return fmt.Errorf("failed to hash credential updates: %w", err)
	}

	var hashedMap map[string]interface{}
	if err := json.Unmarshal(hashedUpdates, &hashedMap); err != nil {
		return fmt.Errorf("failed to unmarshal hashed updates: %w", err)
	}

	// Fetch existing, merge, and store.
	return s.transactioner.Transact(ctx, func(txCtx context.Context) error {
		_, _, existingCredsJSON, err := s.store.GetEntityWithCredentials(txCtx, entityID)
		if err != nil {
			return err
		}

		existingCreds := make(map[string]interface{})
		if len(existingCredsJSON) > 0 {
			if err := json.Unmarshal(existingCredsJSON, &existingCreds); err != nil {
				return fmt.Errorf("failed to unmarshal existing credentials: %w", err)
			}
		}

		// Merge: existing preserved, new/updated types replaced.
		for k, v := range hashedMap {
			existingCreds[k] = v
		}

		mergedJSON, err := json.Marshal(existingCreds)
		if err != nil {
			return fmt.Errorf("failed to marshal merged credentials: %w", err)
		}

		return s.store.UpdateSystemCredentials(txCtx, entityID, mergedJSON)
	})
}

// validateEntitySchema validates entity attributes and uniqueness against the entity schema.
// For user entities, it delegates to the user schema service for attribute validation and
// uniqueness checking. excludeEntityID is used to exclude the entity itself from uniqueness
// checks during updates (empty string for creates). skipCredentialRequired controls whether
// credential fields are required (false for creates, true for updates).
func (s *entityService) validateEntitySchema(
	ctx context.Context,
	category EntityCategory,
	entityType string,
	attributes json.RawMessage,
	excludeEntityID string,
	skipCredentialRequired bool,
) error {
	// Only user entities have schema validation for now.
	if category != EntityCategoryUser || s.userSchemaService == nil {
		return nil
	}

	// Validate attributes against schema (required fields, regex patterns, types).
	isValid, svcErr := s.userSchemaService.ValidateUser(ctx, entityType, attributes, skipCredentialRequired)
	if svcErr != nil {
		return fmt.Errorf("%w: %s", ErrSchemaValidationFailed, svcErr.ErrorDescription)
	}
	if !isValid {
		return ErrSchemaValidationFailed
	}

	// Validate attribute uniqueness
	isValid, svcErr = s.userSchemaService.ValidateUserUniqueness(ctx, entityType, attributes,
		func(filters map[string]interface{}) (*string, error) {
			id, err := s.IdentifyEntity(ctx, filters)
			if err != nil {
				if errors.Is(err, ErrEntityNotFound) {
					return nil, nil // Not found = unique
				}
				return nil, err
			}
			// Exclude self from uniqueness check during updates.
			if excludeEntityID != "" && id != nil && *id == excludeEntityID {
				return nil, nil
			}
			return id, nil
		})
	if svcErr != nil {
		return fmt.Errorf("%w: %s", ErrAttributeConflict, svcErr.ErrorDescription)
	}
	if !isValid {
		return ErrAttributeConflict
	}

	return nil
}

// mergeCredentialJSON merges new credential JSON into existing credential JSON.
// New credential types replace existing ones; types not in the update are preserved.
func mergeCredentialJSON(existing, updates json.RawMessage) json.RawMessage {
	if len(updates) == 0 {
		return existing
	}
	if len(existing) == 0 {
		return updates
	}

	var existingMap map[string]interface{}
	if err := json.Unmarshal(existing, &existingMap); err != nil {
		return updates
	}

	var updatesMap map[string]interface{}
	if err := json.Unmarshal(updates, &updatesMap); err != nil {
		return existing
	}

	for k, v := range updatesMap {
		existingMap[k] = v
	}

	merged, err := json.Marshal(existingMap)
	if err != nil {
		return updates
	}
	return merged
}

// extractAndHashSchemaCredentials extracts schema-defined credential fields from entity.Attributes,
// hashes them, and returns the hashed credentials.
func (s *entityService) extractAndHashSchemaCredentials(ctx context.Context, entity *Entity) (json.RawMessage, error) {
	// Only user entities have schema-defined credentials for now.
	if entity.Category != EntityCategoryUser {
		return nil, nil
	}

	if s.userSchemaService == nil || len(entity.Attributes) == 0 {
		return nil, nil
	}

	credentialFields, svcErr := s.userSchemaService.GetCredentialAttributes(ctx, entity.Type)
	if svcErr != nil {
		return nil, fmt.Errorf("failed to get credential attributes from schema: %s", svcErr.ErrorDescription)
	}

	if len(credentialFields) == 0 {
		return nil, nil
	}

	var attrsMap map[string]interface{}
	if err := json.Unmarshal(entity.Attributes, &attrsMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal entity attributes: %w", err)
	}

	plaintextCreds := make(map[string]string)
	for _, field := range credentialFields {
		if val, ok := attrsMap[field].(string); ok && val != "" {
			plaintextCreds[field] = val
			delete(attrsMap, field)
		}
	}

	if len(plaintextCreds) == 0 {
		return nil, nil
	}

	// Update entity.Attributes with credentials removed.
	cleanAttrs, err := json.Marshal(attrsMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal cleaned attributes: %w", err)
	}
	entity.Attributes = cleanAttrs

	// Hash and return as JSON.
	plaintextJSON, err := json.Marshal(plaintextCreds)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal plaintext credentials: %w", err)
	}

	return s.hashPlaintextCredentials(plaintextJSON)
}

// hashPlaintextCredentials processes system credentials JSON, hashing any plaintext values.
// Values that are already in the stored format (arrays of credential objects) are passed through as-is.
// This allows declarative resource loaders to pre-hash credentials.
func (s *entityService) hashPlaintextCredentials(creds json.RawMessage) (json.RawMessage, error) {
	if len(creds) == 0 {
		return creds, nil
	}

	var credsMap map[string]interface{}
	if err := json.Unmarshal(creds, &credsMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credentials: %w", err)
	}

	if len(credsMap) == 0 {
		return creds, nil
	}

	result := make(map[string]interface{}, len(credsMap))
	for credType, credValue := range credsMap {
		switch v := credValue.(type) {
		case string:
			// Plaintext string value — hash it.
			if v == "" {
				continue
			}
			credHash, err := s.hashService.Generate([]byte(v))
			if err != nil {
				return nil, fmt.Errorf("failed to hash credential %q: %w", credType, err)
			}
			result[credType] = []storedCredential{
				{
					StorageAlgo: credHash.Algorithm,
					StorageAlgoParams: hash.CredParameters{
						Salt:       credHash.Parameters.Salt,
						Iterations: credHash.Parameters.Iterations,
						KeySize:    credHash.Parameters.KeySize,
					},
					Value: credHash.Hash,
				},
			}
		default:
			// Already in stored format (array of credential objects) — pass through.
			result[credType] = credValue
		}
	}

	return json.Marshal(result)
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

// LoadIndexedAttributes loads attributes to be indexed for fast lookups.
// Consumers call this at startup to declare which of their attributes should be indexed.
func (s *entityService) LoadIndexedAttributes(attributes []string) error {
	if err := s.store.LoadIndexedAttributes(attributes); err != nil {
		s.logger.Error("Failed to load indexed attributes", log.Error(err))
		return err
	}
	return nil
}
