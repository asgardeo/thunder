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
	"container/heap"
	"container/list"
	"sync"
	"time"

	"github.com/asgardeo/thunder/internal/system/log"
)

// lfuHeapItem represents an item in the LFU heap.
type lfuHeapItem struct {
	key         CacheKey
	accessCount int64
	lastAccess  time.Time
	index       int // Index in the heap
}

// lfuHeap implements heap.Interface for LFU eviction.
type lfuHeap []*lfuHeapItem

func (h lfuHeap) Len() int { return len(h) }

func (h lfuHeap) Less(i, j int) bool {
	// Primary: fewer accesses come first
	if h[i].accessCount != h[j].accessCount {
		return h[i].accessCount < h[j].accessCount
	}
	// Tie-breaker: earlier access time comes first
	return h[i].lastAccess.Before(h[j].lastAccess)
}

func (h lfuHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *lfuHeap) Push(x any) {
	n := len(*h)
	item := x.(*lfuHeapItem)
	item.index = n
	*h = append(*h, item)
}

func (h *lfuHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*h = old[0 : n-1]
	return item
}

// inMemoryCacheEntry represents an entry in the in-memory cache with additional metadata.
type inMemoryCacheEntry[T any] struct {
	*CacheEntry[T]
	listElement *list.Element
	heapItem    *lfuHeapItem
	lastAccess  time.Time
	accessCount int64
}

// inMemoryCache implements the CacheInterface for an in-memory cache.
type inMemoryCache[T any] struct {
	enabled        bool
	name           string
	cache          map[CacheKey]*inMemoryCacheEntry[T]
	accessOrder    *list.List
	lfuHeap        *lfuHeap
	mu             sync.RWMutex
	size           int
	ttl            time.Duration
	evictionPolicy evictionPolicy
	hitCount       int64
	missCount      int64
	evictCount     int64
}

// newInMemoryCache creates a new instance of InMemoryCache.
func newInMemoryCache[T any](name string, enabled bool, size int, ttl time.Duration,
	evictionPolicy evictionPolicy) internalCacheInterface[T] {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "InMemoryCache"),
		log.String("name", name))

	if !enabled {
		logger.Warn("In-memory cache is disabled, returning empty cache")
		return &inMemoryCache[T]{
			name:    name,
			enabled: false,
		}
	}

	logger.Debug("Initializing In-memory cache", log.String("evictionPolicy", string(evictionPolicy)),
		log.Int("size", size), log.Any("ttl", ttl))

	lfuHeapInstance := &lfuHeap{}
	heap.Init(lfuHeapInstance)

	return &inMemoryCache[T]{
		enabled:        true,
		name:           name,
		cache:          make(map[CacheKey]*inMemoryCacheEntry[T]),
		accessOrder:    list.New(),
		lfuHeap:        lfuHeapInstance,
		size:           size,
		ttl:            ttl,
		evictionPolicy: evictionPolicy,
	}
}

// Set adds or updates an entry in the cache.
func (c *inMemoryCache[T]) Set(key CacheKey, value T) error {
	if !c.enabled {
		return nil
	}

	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "InMemoryCache"),
		log.String("name", c.GetName()))

	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	expiryTime := now.Add(c.ttl)

	// Update existing entry if an entry exists
	if existingEntry, exists := c.cache[key]; exists {
		existingEntry.Value = value
		existingEntry.ExpiryTime = expiryTime
		existingEntry.lastAccess = now
		existingEntry.accessCount++
		c.accessOrder.MoveToFront(existingEntry.listElement)

		// Update the heap item for LFU eviction
		if c.evictionPolicy == evictionPolicyLFU && existingEntry.heapItem != nil {
			existingEntry.heapItem.accessCount = existingEntry.accessCount
			existingEntry.heapItem.lastAccess = existingEntry.lastAccess
			heap.Fix(c.lfuHeap, existingEntry.heapItem.index)
		}
		return nil
	}

	// Create new entry
	cacheEntry := &CacheEntry[T]{
		Value:      value,
		ExpiryTime: expiryTime,
	}

	listElement := c.accessOrder.PushFront(key)

	// Create heap item for LFU eviction
	var heapItem *lfuHeapItem
	if c.evictionPolicy == evictionPolicyLFU {
		heapItem = &lfuHeapItem{
			key:         key,
			accessCount: 1,
			lastAccess:  now,
		}
		heap.Push(c.lfuHeap, heapItem)
	}

	inMemoryCacheEntry := &inMemoryCacheEntry[T]{
		CacheEntry:  cacheEntry,
		listElement: listElement,
		heapItem:    heapItem,
		lastAccess:  now,
		accessCount: 1,
	}
	c.cache[key] = inMemoryCacheEntry

	// Check if there's a requirement to evict an entry
	if len(c.cache) > c.size {
		logger.Debug("Cache size exceeded, evicting an entry")
		c.evict()
	}

	logger.Debug("Cache entry set", log.String("key", key.ToString()))
	return nil
}

// Get retrieves a value from the cache.
func (c *inMemoryCache[T]) Get(key CacheKey) (T, bool) {
	if !c.enabled {
		var zero T
		return zero, false
	}

	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "InMemoryCache"),
		log.String("name", c.GetName()))

	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.cache[key]
	if !exists {
		c.missCount++
		var zero T
		return zero, false
	}

	// Check if the entry has expired
	if time.Now().After(entry.ExpiryTime) {
		c.deleteEntry(key, entry)
		c.missCount++
		var zero T
		return zero, false
	}

	// Update access order for LRU/LFU
	entry.lastAccess = time.Now()
	entry.accessCount++
	c.accessOrder.MoveToFront(entry.listElement)
	c.hitCount++

	// Update the heap item for LFU eviction
	if c.evictionPolicy == evictionPolicyLFU && entry.heapItem != nil {
		entry.heapItem.accessCount = entry.accessCount
		entry.heapItem.lastAccess = entry.lastAccess
		heap.Fix(c.lfuHeap, entry.heapItem.index)
	}

	logger.Debug("Cache hit", log.String("key", key.ToString()))
	return entry.Value, true
}

// Delete removes an entry from the cache.
func (c *inMemoryCache[T]) Delete(key CacheKey) error {
	if !c.enabled {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, exists := c.cache[key]; exists {
		c.deleteEntry(key, entry)
	}

	return nil
}

// Clear removes all entries from the cache.
func (c *inMemoryCache[T]) Clear() error {
	if !c.enabled {
		return nil
	}

	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "InMemoryCache"),
		log.String("name", c.GetName()))

	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[CacheKey]*inMemoryCacheEntry[T])
	c.accessOrder.Init()
	c.lfuHeap = &lfuHeap{}
	heap.Init(c.lfuHeap)
	c.hitCount = 0
	c.missCount = 0
	c.evictCount = 0

	logger.Debug("Cleared all entries in the cache")
	return nil
}

// IsEnabled returns whether the cache is enabled.
func (c *inMemoryCache[T]) IsEnabled() bool {
	return c.enabled
}

// GetName returns the name of the cache.
func (c *inMemoryCache[T]) GetName() string {
	return c.name
}

// GetStats returns cache statistics.
func (c *inMemoryCache[T]) GetStats() CacheStat {
	if !c.enabled {
		return CacheStat{Enabled: false}
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	size := len(c.cache)
	totalOps := c.hitCount + c.missCount
	var hitRate float64
	if totalOps > 0 {
		hitRate = float64(c.hitCount) / float64(totalOps)
	}

	return CacheStat{
		Enabled:    true,
		Size:       size,
		MaxSize:    c.size,
		HitCount:   c.hitCount,
		MissCount:  c.missCount,
		HitRate:    hitRate,
		EvictCount: c.evictCount,
	}
}

// evict removes an entry based on the eviction policy.
func (c *inMemoryCache[T]) evict() {
	if c.evictionPolicy == evictionPolicyLFU {
		c.evictLeastFrequent()
	} else {
		c.evictOldest()
	}
}

// evictOldest removes the oldest entry from the cache (LRU eviction).
func (c *inMemoryCache[T]) evictOldest() {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "InMemoryCache"),
		log.String("name", c.GetName()))

	if c.accessOrder.Len() == 0 {
		return
	}

	// Get the least recently used item
	oldest := c.accessOrder.Back()
	if oldest != nil {
		key := oldest.Value.(CacheKey)
		if entry, exists := c.cache[key]; exists {
			c.deleteEntry(key, entry)
			c.evictCount++
			logger.Debug("Cache entry evicted", log.String("key", key.ToString()))
		}
	}
}

// evictLeastFrequent removes the least frequently used entry from the cache (LFU eviction).
func (c *inMemoryCache[T]) evictLeastFrequent() {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "InMemoryCache"),
		log.String("name", c.GetName()))

	if c.lfuHeap.Len() == 0 {
		return
	}

	// Get the least frequently used item from the heap
	leastFrequentItem := heap.Pop(c.lfuHeap).(*lfuHeapItem)

	if entry, exists := c.cache[leastFrequentItem.key]; exists {
		c.deleteEntry(leastFrequentItem.key, entry)
		c.evictCount++
		logger.Debug("Cache entry evicted (LFU)", log.String("key", leastFrequentItem.key.ToString()),
			log.Any("accessCount", leastFrequentItem.accessCount))
	}
}

// deleteEntry removes an entry from both the map and the access order list.
func (c *inMemoryCache[T]) deleteEntry(key CacheKey, entry *inMemoryCacheEntry[T]) {
	delete(c.cache, key)
	c.accessOrder.Remove(entry.listElement)

	// Remove from heap if using LFU eviction
	if c.evictionPolicy == evictionPolicyLFU && entry.heapItem != nil && entry.heapItem.index >= 0 {
		heap.Remove(c.lfuHeap, entry.heapItem.index)
	}
}

// CleanupExpired removes all expired entries from the cache.
func (c *inMemoryCache[T]) CleanupExpired() {
	if !c.enabled {
		return
	}

	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "InMemoryCache"),
		log.String("name", c.GetName()))
	logger.Debug("Cleaning up expired entries from the cache")

	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	cleaned := 0
	for key, entry := range c.cache {
		if now.After(entry.ExpiryTime) {
			c.deleteEntry(key, entry)
			cleaned++
		}
	}

	if logger.IsDebugEnabled() {
		if cleaned > 0 {
			logger.Debug("Expired cache entries cleaned", log.Int("count", cleaned))
		} else {
			logger.Debug("No expired entries found in the cache")
		}
	}
}
