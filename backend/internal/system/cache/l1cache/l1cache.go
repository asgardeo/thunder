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

// Package l1cache provides the implementation for L1 caching using in-memory storage.
package l1cache

import (
	"container/list"
	"sync"
	"time"

	"github.com/asgardeo/thunder/internal/system/cache/constants"
	"github.com/asgardeo/thunder/internal/system/cache/model"
	"github.com/asgardeo/thunder/internal/system/log"
)

const loggerComponentName = "L1Cache"

// L1CacheEntry represents an entry in the L1 cache with access tracking.
type L1CacheEntry[T any] struct {
	*model.CacheEntry[T]
	listElement *list.Element
	lastAccess  time.Time
}

// L1Cache implements the L1CacheInterface with in-memory caching and LRU eviction.
type L1Cache[T any] struct {
	enabled        bool
	cache          map[model.CacheKey]*L1CacheEntry[T]
	accessOrder    *list.List
	mu             sync.RWMutex
	maxSize        int
	ttl            time.Duration
	evictionPolicy string
	hitCount       int64
	missCount      int64
	evictCount     int64
}

// NewL1Cache creates a new instance of L1Cache.
func NewL1Cache[T any](enabled bool, maxSize int, ttl time.Duration, evictionPolicy string) model.CacheInterface[T] {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if !enabled {
		logger.Warn("L1 cache is disabled")
		return &L1Cache[T]{enabled: false}
	}

	cacheMaxSize := maxSize
	if cacheMaxSize <= 0 {
		cacheMaxSize = constants.L1DefaultMaxSize
	}

	cacheTTL := ttl
	if cacheTTL <= 0 {
		cacheTTL = constants.L1DefaultTTL * time.Second
	}

	logger.Debug("Initializing L1 cache", log.String("evictionPolicy", evictionPolicy),
		log.Int("maxSize", cacheMaxSize), log.Any("ttl", cacheTTL))

	return &L1Cache[T]{
		enabled:        true,
		cache:          make(map[model.CacheKey]*L1CacheEntry[T]),
		accessOrder:    list.New(),
		maxSize:        cacheMaxSize,
		ttl:            cacheTTL,
		evictionPolicy: evictionPolicy,
	}
}

// Set adds or updates an entry in the cache.
func (l1 *L1Cache[T]) Set(key model.CacheKey, value T) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if !l1.enabled {
		return nil
	}

	l1.mu.Lock()
	defer l1.mu.Unlock()

	now := time.Now()
	expiryTime := now.Add(l1.ttl)

	// Check if entry already exists
	if existingEntry, exists := l1.cache[key]; exists {
		// Update existing entry
		existingEntry.Value = value
		existingEntry.ExpiryTime = expiryTime
		existingEntry.lastAccess = now
		l1.accessOrder.MoveToFront(existingEntry.listElement)
		return nil
	}

	// Create new entry
	cacheEntry := &model.CacheEntry[T]{
		Value:      value,
		ExpiryTime: expiryTime,
	}

	listElement := l1.accessOrder.PushFront(key)
	l1CacheEntry := &L1CacheEntry[T]{
		CacheEntry:  cacheEntry,
		listElement: listElement,
		lastAccess:  now,
	}

	l1.cache[key] = l1CacheEntry

	// Check if we need to evict
	if len(l1.cache) > l1.maxSize {
		l1.evictOldest()
	}

	logger.Debug("L1 cache entry added", log.String("key", key.ToString()), log.Any("expiry", expiryTime))

	return nil
}

// Get retrieves a value from the cache.
func (l1 *L1Cache[T]) Get(key model.CacheKey) (T, bool) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if !l1.enabled {
		var zero T
		return zero, false
	}

	l1.mu.Lock()
	defer l1.mu.Unlock()

	entry, exists := l1.cache[key]
	if !exists {
		l1.missCount++
		var zero T
		return zero, false
	}

	// Check if entry has expired
	if time.Now().After(entry.ExpiryTime) {
		l1.deleteEntry(key, entry)
		l1.missCount++
		var zero T
		return zero, false
	}

	// Update access order for LRU
	entry.lastAccess = time.Now()
	l1.accessOrder.MoveToFront(entry.listElement)
	l1.hitCount++

	logger.Debug("L1 cache hit", log.String("key", key.ToString()))

	return entry.Value, true
}

// Delete removes an entry from the cache.
func (l1 *L1Cache[T]) Delete(key model.CacheKey) error {
	if !l1.enabled {
		return nil
	}

	l1.mu.Lock()
	defer l1.mu.Unlock()

	if entry, exists := l1.cache[key]; exists {
		l1.deleteEntry(key, entry)
	}

	return nil
}

// Clear removes all entries from the cache.
func (l1 *L1Cache[T]) Clear() error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if !l1.enabled {
		return nil
	}

	l1.mu.Lock()
	defer l1.mu.Unlock()

	l1.cache = make(map[model.CacheKey]*L1CacheEntry[T])
	l1.accessOrder.Init()
	l1.hitCount = 0
	l1.missCount = 0
	l1.evictCount = 0

	logger.Debug("L1 cache cleared")
	return nil
}

// IsEnabled returns whether the cache is enabled.
func (l1 *L1Cache[T]) IsEnabled() bool {
	return l1.enabled
}

// GetStats returns cache statistics.
func (l1 *L1Cache[T]) GetStats() model.CacheStat {
	if !l1.enabled {
		return model.CacheStat{Enabled: false}
	}

	l1.mu.RLock()
	defer l1.mu.RUnlock()

	size := len(l1.cache)
	totalOps := l1.hitCount + l1.missCount
	var hitRate float64
	if totalOps > 0 {
		hitRate = float64(l1.hitCount) / float64(totalOps)
	}

	return model.CacheStat{
		Enabled:    true,
		Size:       size,
		MaxSize:    l1.maxSize,
		HitCount:   l1.hitCount,
		MissCount:  l1.missCount,
		HitRate:    hitRate,
		EvictCount: l1.evictCount,
	}
}

// evictOldest removes the oldest entry from the cache (LRU eviction).
func (l1 *L1Cache[T]) evictOldest() {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if l1.accessOrder.Len() == 0 {
		return
	}

	// Get the least recently used item
	oldest := l1.accessOrder.Back()
	if oldest != nil {
		key := oldest.Value.(model.CacheKey)
		if entry, exists := l1.cache[key]; exists {
			l1.deleteEntry(key, entry)
			l1.evictCount++
			logger.Debug("L1 cache entry evicted", log.String("key", key.ToString()))
		}
	}
}

// deleteEntry removes an entry from both the map and the access order list.
func (l1 *L1Cache[T]) deleteEntry(key model.CacheKey, entry *L1CacheEntry[T]) {
	delete(l1.cache, key)
	l1.accessOrder.Remove(entry.listElement)
}

// CleanupExpired removes all expired entries from the cache.
func (l1 *L1Cache[T]) CleanupExpired() {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Cleaning up expired entries from L1 cache")

	if !l1.enabled {
		return
	}

	l1.mu.Lock()
	defer l1.mu.Unlock()

	now := time.Now()
	var expiredKeys []model.CacheKey

	for key, entry := range l1.cache {
		if now.After(entry.ExpiryTime) {
			expiredKeys = append(expiredKeys, key)
		}
	}

	for _, key := range expiredKeys {
		if entry, exists := l1.cache[key]; exists {
			l1.deleteEntry(key, entry)
		}
	}

	if logger.IsDebugEnabled() {
		if len(expiredKeys) > 0 {
			logger.Debug("L1 cache expired entries cleaned", log.Int("count", len(expiredKeys)))
		} else {
			logger.Debug("No expired entries found in L1 cache")
		}
	}
}
