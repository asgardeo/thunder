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

// Package cache provides a centralized cache management system for different cache implementations.
package cache

import (
	"sync"
	"time"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/log"
)

// internalCacheInterface defines the common interface for internal cache implementations.
type internalCacheInterface[T any] interface {
	Set(key CacheKey, value T) error
	Get(key CacheKey) (T, bool)
	Delete(key CacheKey) error
	Clear() error
	IsEnabled() bool
	GetStats() CacheStat
	CleanupExpired()
	GetName() string
}

// CacheInterface defines the common interface for cache operations.
type CacheInterface[T any] interface {
	GetName() string
	Set(key CacheKey, value T) error
	Get(key CacheKey) (T, bool)
	Delete(key CacheKey) error
	Clear() error
	IsEnabled() bool
	CleanupExpired()
}

// Cache implements the CacheInterface for individual caches.
type Cache[T any] struct {
	enabled       bool
	cacheName     string
	InternalCache internalCacheInterface[T]
	mu            sync.RWMutex
}

// newCache creates a new cache instance.
func newCache[T any](cacheName string) CacheInterface[T] {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "Cache"),
		log.String("cacheName", cacheName))

	cacheConfig := config.GetThunderRuntime().Config.Cache
	if cacheConfig.Disabled {
		logger.Debug("Caching is disabled, returning empty")
		return &Cache[T]{
			enabled:       false,
			cacheName:     cacheName,
			InternalCache: nil,
		}
	}

	cacheProperty := getCacheProperty(cacheConfig, cacheName)

	if cacheProperty.Disabled {
		logger.Debug("Individual cache is disabled, returning empty")
		return &Cache[T]{
			enabled:       false,
			cacheName:     cacheName,
			InternalCache: nil,
		}
	}

	logger.Debug("Initializing the cache")

	cacheType := getCacheType(cacheConfig)
	evictionPolicy := getEvictionPolicy(cacheConfig, cacheProperty)
	size := getCacheSize(cacheConfig, cacheProperty)
	ttl := getCacheTTL(cacheConfig, cacheProperty)

	var internalCache internalCacheInterface[T]
	switch cacheType {
	case cacheTypeInMemory:
		internalCache = newInMemoryCache[T](
			cacheName,
			!cacheProperty.Disabled,
			size,
			time.Duration(ttl)*time.Second,
			evictionPolicy,
		)
	default:
		logger.Warn("Unknown cache type, defaulting to in-memory cache")
		internalCache = newInMemoryCache[T](
			cacheName,
			!cacheProperty.Disabled,
			size,
			time.Duration(ttl)*time.Second,
			evictionPolicyLRU,
		)
	}

	cache := &Cache[T]{
		enabled:       true,
		cacheName:     cacheName,
		InternalCache: internalCache,
	}

	return cache
}

// GetName returns the name of the cache.
func (c *Cache[T]) GetName() string {
	return c.cacheName
}

// Set stores a value in the cache.
func (c *Cache[T]) Set(key CacheKey, value T) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "Cache"),
		log.String("cacheName", c.cacheName))

	if c.IsEnabled() && c.InternalCache.IsEnabled() {
		c.mu.Lock()
		defer c.mu.Unlock()

		if err := c.InternalCache.Set(key, value); err != nil {
			logger.Warn("Failed to set value in the cache", log.String("key", key.ToString()), log.Error(err))
		}
	}

	return nil
}

// Get retrieves a value from the cache.
func (c *Cache[T]) Get(key CacheKey) (T, bool) {
	if c.IsEnabled() && c.InternalCache.IsEnabled() {
		c.mu.RLock()
		defer c.mu.RUnlock()

		if value, found := c.InternalCache.Get(key); found {
			return value, true
		}
	}

	var zero T
	return zero, false
}

// Delete removes a value from the cache.
func (c *Cache[T]) Delete(key CacheKey) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "Cache"),
		log.String("cacheName", c.cacheName))

	if c.IsEnabled() && c.InternalCache.IsEnabled() {
		c.mu.Lock()
		defer c.mu.Unlock()

		if err := c.InternalCache.Delete(key); err != nil {
			logger.Warn("Failed to delete value from the cache", log.String("key", key.ToString()), log.Error(err))
		}
	}

	return nil
}

// Clear removes all entries in the cache.
func (c *Cache[T]) Clear() error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "Cache"),
		log.String("cacheName", c.cacheName))

	if c.IsEnabled() && c.InternalCache.IsEnabled() {
		logger.Debug("Clearing all entries in the cache")

		c.mu.Lock()
		defer c.mu.Unlock()

		if err := c.InternalCache.Clear(); err != nil {
			logger.Warn("Failed to clear the cache", log.Error(err))
		}
	}

	return nil
}

// IsEnabled returns whether the cache is enabled.
func (c *Cache[T]) IsEnabled() bool {
	return c.enabled
}

// CleanupExpired cleans up expired entries in the cache.
func (c *Cache[T]) CleanupExpired() {
	if c.IsEnabled() && c.InternalCache.IsEnabled() {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.InternalCache.CleanupExpired()
	}
}

// getCacheType retrieves the cache type from the configuration.
//
//nolint:unparam // TODO: Ignoring linter check as we only support in-memory cache for now.
func getCacheType(cacheConfig config.CacheConfig) cacheType {
	if cacheConfig.Type == "" {
		return cacheTypeInMemory
	}
	switch cacheConfig.Type {
	case string(cacheTypeInMemory):
		return cacheTypeInMemory
	default:
		log.GetLogger().Warn("Unknown cache type, defaulting to in-memory cache")
		return cacheTypeInMemory
	}
}

// getCacheProperty retrieves the cache property for the specified cache name.
func getCacheProperty(cacheConfig config.CacheConfig, cacheName string) config.CacheProperty {
	for _, property := range cacheConfig.Properties {
		if property.Name == cacheName {
			return property
		}
	}
	return config.CacheProperty{}
}

// getEvictionPolicy retrieves the eviction policy from the cache configuration.
func getEvictionPolicy(cacheConfig config.CacheConfig, cacheProperty config.CacheProperty) evictionPolicy {
	evictionPolicy := cacheProperty.EvictionPolicy
	if evictionPolicy == "" {
		evictionPolicy = cacheConfig.EvictionPolicy
	}
	if evictionPolicy == "" {
		return evictionPolicyLRU
	}

	switch evictionPolicy {
	case string(evictionPolicyLRU):
		return evictionPolicyLRU
	case string(evictionPolicyLFU):
		return evictionPolicyLFU
	default:
		log.GetLogger().Warn("Unknown eviction policy, defaulting to LRU")
		return evictionPolicyLRU
	}
}

// getCacheSize retrieves the cache size from the cache configuration.
func getCacheSize(cacheConfig config.CacheConfig, cacheProperty config.CacheProperty) int {
	size := cacheProperty.Size
	if size <= 0 {
		size = cacheConfig.Size
	}
	return size
}

// getCacheTTL retrieves the cache TTL from the cache configuration.
func getCacheTTL(cacheConfig config.CacheConfig, cacheProperty config.CacheProperty) int {
	ttl := cacheProperty.TTL
	if ttl <= 0 {
		ttl = cacheConfig.TTL
	}
	return ttl
}
