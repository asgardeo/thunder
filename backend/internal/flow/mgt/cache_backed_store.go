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

package flowmgt

import (
	"errors"

	"github.com/asgardeo/thunder/internal/system/cache"
	"github.com/asgardeo/thunder/internal/system/log"
)

const cacheBackedStoreLoggerComponentName = "CacheBackedFlowStore"

// cacheBackedFlowStore is the implementation of flowStoreInterface that uses caching.
type cacheBackedFlowStore struct {
	flowByIDCache cache.CacheInterface[*CompleteFlowDefinition]
	store         flowStoreInterface
	logger        *log.Logger
}

// newCacheBackedFlowStore creates a new instance of cacheBackedFlowStore.
func newCacheBackedFlowStore() flowStoreInterface {
	return &cacheBackedFlowStore{
		flowByIDCache: cache.GetCache[*CompleteFlowDefinition]("FlowByIDCache"),
		store:         newFlowStore(),
		logger: log.GetLogger().With(
			log.String(log.LoggerKeyComponentName, cacheBackedStoreLoggerComponentName)),
	}
}

// ListFlows retrieves a paginated list of flow definitions.
// Note: List operations are not cached as they can vary by parameters and change frequently.
func (s *cacheBackedFlowStore) ListFlows(limit, offset int, flowType string) (
	[]BasicFlowDefinition, int, error) {
	return s.store.ListFlows(limit, offset, flowType)
}

// CreateFlow creates a new flow definition and caches it.
func (s *cacheBackedFlowStore) CreateFlow(flowID string, flow *FlowDefinition) (
	*CompleteFlowDefinition, error) {
	createdFlow, err := s.store.CreateFlow(flowID, flow)
	if err != nil {
		return nil, err
	}
	s.cacheFlow(createdFlow)

	return createdFlow, nil
}

// GetFlowByID retrieves a flow definition by its ID, using cache if available.
func (s *cacheBackedFlowStore) GetFlowByID(flowID string) (*CompleteFlowDefinition, error) {
	cacheKey := cache.CacheKey{
		Key: flowID,
	}
	cachedFlow, ok := s.flowByIDCache.Get(cacheKey)
	if ok {
		return cachedFlow, nil
	}

	flow, err := s.store.GetFlowByID(flowID)
	if err != nil || flow == nil {
		return flow, err
	}
	s.cacheFlow(flow)

	return flow, nil
}

// UpdateFlow updates an existing flow definition and refreshes the cache.
func (s *cacheBackedFlowStore) UpdateFlow(flowID string, flow *FlowDefinition) (
	*CompleteFlowDefinition, error) {
	updatedFlow, err := s.store.UpdateFlow(flowID, flow)
	if err != nil {
		return nil, err
	}
	s.cacheFlow(updatedFlow)

	return updatedFlow, nil
}

// DeleteFlow deletes a flow definition by its ID and invalidates the cache.
func (s *cacheBackedFlowStore) DeleteFlow(flowID string) error {
	cacheKey := cache.CacheKey{
		Key: flowID,
	}
	existingFlow, ok := s.flowByIDCache.Get(cacheKey)
	if !ok {
		var err error
		existingFlow, err = s.store.GetFlowByID(flowID)
		if err != nil {
			if errors.Is(err, errFlowNotFound) {
				return nil
			}
			return err
		}
	}
	if existingFlow == nil {
		return nil
	}

	if err := s.store.DeleteFlow(flowID); err != nil {
		return err
	}
	s.invalidateFlowCache(flowID)

	return nil
}

// ListFlowVersions retrieves all versions of a flow.
// Note: Version operations are not cached as they are less frequently accessed.
func (s *cacheBackedFlowStore) ListFlowVersions(flowID string) ([]BasicFlowVersion, error) {
	return s.store.ListFlowVersions(flowID)
}

// GetFlowVersion retrieves a specific version of a flow.
// Note: Version operations are not cached as they are less frequently accessed.
func (s *cacheBackedFlowStore) GetFlowVersion(flowID string, version int) (*FlowVersion, error) {
	return s.store.GetFlowVersion(flowID, version)
}

// RestoreFlowVersion restores a flow to a specific version and invalidates the cache.
func (s *cacheBackedFlowStore) RestoreFlowVersion(flowID string, version int) (
	*CompleteFlowDefinition, error) {
	restoredFlow, err := s.store.RestoreFlowVersion(flowID, version)
	if err != nil {
		return nil, err
	}

	s.cacheFlow(restoredFlow)

	return restoredFlow, nil
}

// cacheFlow caches the flow definition by ID.
func (s *cacheBackedFlowStore) cacheFlow(flow *CompleteFlowDefinition) {
	if flow == nil {
		return
	}

	logger := s.logger.With(log.String("flowID", flow.ID))

	if flow.ID != "" {
		cacheKey := cache.CacheKey{
			Key: flow.ID,
		}
		if err := s.flowByIDCache.Set(cacheKey, flow); err != nil {
			logger.Error("Failed to cache flow by ID", log.Error(err))
		} else {
			logger.Debug("Flow cached by ID")
		}
	}
}

// invalidateFlowCache invalidates the flow cache for the given ID.
func (s *cacheBackedFlowStore) invalidateFlowCache(flowID string) {
	logger := s.logger.With(log.String("flowID", flowID))

	if flowID != "" {
		cacheKey := cache.CacheKey{
			Key: flowID,
		}
		if err := s.flowByIDCache.Delete(cacheKey); err != nil {
			logger.Error("Failed to invalidate flow cache by ID", log.Error(err))
		} else {
			logger.Debug("Flow cache invalidated by ID")
		}
	}
}
