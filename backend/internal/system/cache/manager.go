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

package cache

import (
	"reflect"
	"sync"
	"time"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/log"
)

// CacheManagerInterface defines the interface for managing caches.
type CacheManagerInterface interface {
	Init()
	IsEnabled() bool
	getMutex() *sync.RWMutex
	getCache(cacheKey string) (interface{}, bool)
	addCache(cacheKey string, cacheInstance interface{})
	startCleanupRoutine()
	cleanupAllCaches()
	reset()
}

// CacheManager implements the CacheManagerInterface for managing multiple caches.
type CacheManager struct {
	caches          map[string]interface{}
	mu              sync.RWMutex
	enabled         bool
	cleanupInterval time.Duration
}

var (
	instance CacheManagerInterface
	once     sync.Once
)

// GetCacheManager returns a singleton instance of CacheManager.
func GetCacheManager() CacheManagerInterface {
	once.Do(func() {
		instance = &CacheManager{
			caches: make(map[string]interface{}),
		}
	})
	return instance
}

// Init initializes the CacheManager, setting up caches and starting cleanup routines.
func (cm *CacheManager) Init() {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "CacheManager"))
	logger.Debug("Initializing Cache Manager")

	cacheConfig := config.GetThunderRuntime().Config.Cache
	if cacheConfig.Disabled {
		cm.enabled = false
		logger.Debug("Caching is disabled. Skipping initialization")
		return
	}

	cm.enabled = true
	cm.cleanupInterval = getCleanupInterval(cacheConfig)
	cm.startCleanupRoutine()

	logger.Debug("Cache Manager initialized", log.Bool("enabled", cm.enabled),
		log.Any("cleanupInterval", cm.cleanupInterval))
}

// IsEnabled checks if the CacheManager is enabled.
func (cm *CacheManager) IsEnabled() bool {
	return cm.enabled
}

// getMutex returns the mutex for synchronizing access to the caches.
func (cm *CacheManager) getMutex() *sync.RWMutex {
	return &cm.mu
}

// getCache retrieves a cache instance by its key.
func (cm *CacheManager) getCache(cacheKey string) (interface{}, bool) {
	cacheInstance, exists := cm.caches[cacheKey]
	return cacheInstance, exists
}

// addCache adds a new cache instance to the manager.
func (cm *CacheManager) addCache(cacheKey string, cacheInstance interface{}) {
	if _, exists := cm.caches[cacheKey]; !exists {
		cm.caches[cacheKey] = cacheInstance
		log.GetLogger().Debug("Cache added", log.String("cacheKey", cacheKey))
	}
}

// startCleanupRoutine starts a background routine to clean up expired caches at regular intervals.
func (cm *CacheManager) startCleanupRoutine() {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "CacheManager"))
	logger.Debug("Starting cleanup routine for caches")

	go func() {
		ticker := time.NewTicker(cm.cleanupInterval)
		defer ticker.Stop()

		for range ticker.C {
			cm.cleanupAllCaches()
		}
	}()

	logger.Debug("Cleanup routine started", log.Any("interval", cm.cleanupInterval))
}

// cleanupAllCaches cleans up expired entries in all caches.
func (cm *CacheManager) cleanupAllCaches() {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "CacheManager"))
	logger.Debug("Cleaning up expired caches")

	for _, cacheEntry := range cm.caches {
		// Use type switch to handle different cache types
		switch cache := cacheEntry.(type) {
		case interface {
			IsEnabled() bool
			GetName() string
			CleanupExpired()
		}:
			if cache.IsEnabled() {
				logger.Debug("Cleaning up cache", log.String("cacheName", cache.GetName()))
				cache.CleanupExpired()
			}
		default:
			logger.Warn("Unknown cache type encountered", log.Any("type", reflect.TypeOf(cacheEntry)))
		}
	}
}

// reset resets the CacheManager, clearing all caches.
func (cm *CacheManager) reset() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.caches = make(map[string]interface{})
}

// GetCache returns a singleton cache instance for the given type and cache name.
func GetCache[T any](cacheName string) CacheInterface[T] {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "CacheManager"))

	cm := GetCacheManager()

	// Create unique key for the cache
	var t T
	typeName := reflect.TypeOf(t).String()
	cacheKey := cacheName + ":" + typeName

	// First try to get from the map
	cm.getMutex().RLock()
	if cache, exists := cm.getCache(cacheKey); exists {
		cm.getMutex().RUnlock()
		if retCache, ok := cache.(CacheInterface[T]); ok {
			return retCache
		}
		logger.Warn("Type mismatch for cache", log.String("cacheName", cacheName),
			log.String("expectedType", typeName), log.String("actualType", reflect.TypeOf(cache).String()))

		return nil
	}
	cm.getMutex().RUnlock()

	// Acquire write lock to create a new cache
	cm.getMutex().Lock()
	defer cm.getMutex().Unlock()

	if cache, exists := cm.getCache(cacheKey); exists {
		if retCache, ok := cache.(CacheInterface[T]); ok {
			return retCache
		}
		logger.Warn("Type mismatch for cache", log.String("cacheName", cacheName),
			log.String("expectedType", typeName), log.String("actualType", reflect.TypeOf(cache).String()))

		return nil
	}

	// Create a new cache
	logger.Debug("Creating new cache", log.String("cacheName", cacheName), log.String("type", typeName))
	newCache := newCache[T](cacheName)
	cm.addCache(cacheKey, newCache)

	return newCache
}

// getCleanupInterval retrieves the cleanup interval from the cache configuration.
func getCleanupInterval(cacheConfig config.CacheConfig) time.Duration {
	cleanupIntervalInt := cacheConfig.CleanupInterval
	return time.Duration(cleanupIntervalInt) * time.Second
}
