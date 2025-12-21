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

	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/system/cache"
	"github.com/asgardeo/thunder/internal/system/log"
)

const cacheBackedStoreLoggerComponentName = "CacheBackedFlowStore"

// cacheBackedFlowStore is the implementation of flowStoreInterface that uses caching.
type cacheBackedFlowStore struct {
	flowByIDCache     cache.CacheInterface[*CompleteFlowDefinition]
	flowByHandleCache cache.CacheInterface[*CompleteFlowDefinition]
	store             flowStoreInterface
	logger            *log.Logger
}

// newCacheBackedFlowStore creates a new instance of cacheBackedFlowStore.
func newCacheBackedFlowStore() flowStoreInterface {
	return &cacheBackedFlowStore{
		flowByIDCache:     cache.GetCache[*CompleteFlowDefinition]("FlowByIDCache"),
		flowByHandleCache: cache.GetCache[*CompleteFlowDefinition]("FlowByHandleCache"),
		store:             newFlowStore(),
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

// GetFlowByHandle retrieves a flow definition by handle and flow type, using cache if available.
func (s *cacheBackedFlowStore) GetFlowByHandle(handle string, flowType common.FlowType) (
	*CompleteFlowDefinition, error) {
	cacheKey := getFlowByHandleCacheKey(handle, flowType)
	cachedFlow, ok := s.flowByHandleCache.Get(cacheKey)
	if ok {
		return cachedFlow, nil
	}

	flow, err := s.store.GetFlowByHandle(handle, flowType)
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
	s.invalidateFlowCacheByHandle(existingFlow.Handle, existingFlow.FlowType)

	return nil
}

// IsFlowExists checks if a flow exists with a given flow ID, using cache if available.
func (s *cacheBackedFlowStore) IsFlowExists(flowID string) (bool, error) {
	cacheKey := cache.CacheKey{
		Key: flowID,
	}
	cachedFlow, ok := s.flowByIDCache.Get(cacheKey)
	if ok && cachedFlow != nil {
		return true, nil
	}

	return s.store.IsFlowExists(flowID)
}

// IsFlowExistsByHandle checks if a flow exists with a given handle and flow type, using cache if available.
func (s *cacheBackedFlowStore) IsFlowExistsByHandle(handle string, flowType common.FlowType) (bool, error) {
	cacheKey := getFlowByHandleCacheKey(handle, flowType)
	cachedFlow, ok := s.flowByHandleCache.Get(cacheKey)
	if ok && cachedFlow != nil {
		return true, nil
	}

	return s.store.IsFlowExistsByHandle(handle, flowType)
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

// cacheFlow caches the flow definition by ID and by handle.
func (s *cacheBackedFlowStore) cacheFlow(flow *CompleteFlowDefinition) {
	if flow == nil {
		return
	}

	logger := s.logger.With(log.String("flowID", flow.ID))

	// Cache by ID
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

	// Cache by handle and flowType
	if flow.Handle != "" && flow.FlowType != "" {
		handleCacheKey := getFlowByHandleCacheKey(flow.Handle, flow.FlowType)
		if err := s.flowByHandleCache.Set(handleCacheKey, flow); err != nil {
			logger.Error("Failed to cache flow by handle",
				log.String("handle", flow.Handle), log.String("flowType", string(flow.FlowType)), log.Error(err))
		} else {
			logger.Debug("Flow cached by handle",
				log.String("handle", flow.Handle), log.String("flowType", string(flow.FlowType)))
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

// invalidateFlowCacheByHandle invalidates the flow cache for the given handle and type.
func (s *cacheBackedFlowStore) invalidateFlowCacheByHandle(handle string, flowType common.FlowType) {
	if handle == "" || flowType == "" {
		return
	}

	cacheKey := getFlowByHandleCacheKey(handle, flowType)
	if err := s.flowByHandleCache.Delete(cacheKey); err != nil {
		s.logger.Error("Failed to invalidate flow cache by handle",
			log.String("handle", handle), log.String("flowType", string(flowType)), log.Error(err))
	}
}

// getFlowByHandleCacheKey generates a cache key for flow lookup by handle and type.
func getFlowByHandleCacheKey(handle string, flowType common.FlowType) cache.CacheKey {
	return cache.CacheKey{
		Key: handle + ":" + string(flowType),
	}
}
