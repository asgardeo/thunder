/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
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

// Package cache provides the caching implementation managing multiple caches.
package cache

import (
	"slices"
	"sync"

	"github.com/asgardeo/thunder/internal/system/cache/constants"
	"github.com/asgardeo/thunder/internal/system/cache/manager"
	"github.com/asgardeo/thunder/internal/system/log"
)

const loggerComponentName = "CacheProvider"

// validCacheTypes defines the list of supported cache types.
var validCacheTypes = []constants.CacheType{
	constants.CacheTypeApplication,
}

// CacheProviderInterface defines the interface for cache provider.
type CacheProviderInterface interface {
	GetCacheManager(cacheType constants.CacheType) manager.CacheManagerInterface
}

// CacheProvider implements the CacheProviderInterface for managing cache instances.
type CacheProvider struct {
	cacheManagers map[constants.CacheType]manager.CacheManagerInterface
	mu            sync.RWMutex
}

// NewCacheProvider creates a new cache provider.
func NewCacheProvider() CacheProviderInterface {
	return &CacheProvider{
		cacheManagers: make(map[constants.CacheType]manager.CacheManagerInterface),
	}
}

// GetCacheManager returns the cache manager for the specified cache type.
func (cp *CacheProvider) GetCacheManager(cacheType constants.CacheType) manager.CacheManagerInterface {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if !isValidCacheType(cacheType) {
		logger.Error("Invalid cache type", log.String("cacheType", string(cacheType)))
		return nil
	}

	cp.mu.RLock()
	if manager, exists := cp.cacheManagers[cacheType]; exists {
		cp.mu.RUnlock()
		return manager
	}
	cp.mu.RUnlock()

	// Create a new cache manager if it does not exist
	cp.mu.Lock()
	defer cp.mu.Unlock()

	if manager, exists := cp.cacheManagers[cacheType]; exists {
		return manager
	}

	cp.cacheManagers[cacheType] = manager.NewCacheManager()
	logger.Info("Cache manager created for type", log.String("cacheType", string(cacheType)))

	// Start cleanup routine if cache is enabled
	if cp.cacheManagers[cacheType].IsEnabled() {
		if manager, ok := cp.cacheManagers[cacheType].(*manager.CacheManager); ok {
			manager.StartCleanupRoutine()
			logger.Info("Started cleanup routine for cache manager", log.String("cacheType", string(cacheType)))
		}
	}

	return cp.cacheManagers[cacheType]
}

// isValidCacheType checks if the provided cache type is valid.
func isValidCacheType(cacheType constants.CacheType) bool {
	return slices.Contains(validCacheTypes, cacheType)
}
