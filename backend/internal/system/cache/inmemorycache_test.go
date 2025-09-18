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
	"testing"
	"time"

	"github.com/asgardeo/thunder/internal/system/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type InMemoryCacheTestSuite struct {
	suite.Suite
}

func TestInMemoryCacheSuite(t *testing.T) {
	suite.Run(t, new(InMemoryCacheTestSuite))
}

func (suite *InMemoryCacheTestSuite) SetupSuite() {
	mockConfig := &config.Config{
		Cache: config.CacheConfig{
			Disabled:        false,
			Size:            1000,
			TTL:             3600,
			EvictionPolicy:  "LRU",
			CleanupInterval: 300,
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/test/thunder/home", mockConfig)
	if err != nil {
		suite.T().Fatal("Failed to initialize ThunderRuntime:", err)
	}
}

func (suite *InMemoryCacheTestSuite) TestNewInMemoryCache() {
	testCases := []struct {
		name           string
		enabled        bool
		size           int
		ttl            time.Duration
		evictionPolicy evictionPolicy
	}{
		{
			name:           "EnabledCache",
			enabled:        true,
			size:           100,
			ttl:            time.Second * 60,
			evictionPolicy: evictionPolicyLRU,
		},
		{
			name:           "DisabledCache",
			enabled:        false,
			size:           100,
			ttl:            time.Second * 60,
			evictionPolicy: evictionPolicyLRU,
		},
		{
			name:           "LFUEvictionPolicy",
			enabled:        true,
			size:           100,
			ttl:            time.Second * 60,
			evictionPolicy: evictionPolicyLFU,
		},
		{
			name:           "ZeroSize",
			enabled:        true,
			size:           0,
			ttl:            time.Second * 60,
			evictionPolicy: evictionPolicyLRU,
		},
		{
			name:           "ZeroTTL",
			enabled:        true,
			size:           100,
			ttl:            0,
			evictionPolicy: evictionPolicyLRU,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			cache := newInMemoryCache[string](tc.name, tc.enabled, tc.size, tc.ttl, tc.evictionPolicy)

			assert.NotNil(t, cache)
			assert.Equal(t, tc.enabled, cache.IsEnabled())
			assert.Equal(t, tc.name, cache.GetName())

			// Verify proper initialization by checking stats
			stats := cache.GetStats()
			assert.Equal(t, tc.enabled, stats.Enabled)

			if tc.enabled {
				assert.Equal(t, 0, stats.Size)
				assert.Equal(t, tc.size, stats.MaxSize)
			}
		})
	}
}

func (suite *InMemoryCacheTestSuite) TestSetAndGet() {
	testCases := []struct {
		name           string
		evictionPolicy evictionPolicy
	}{
		{
			name:           "LRUCache",
			evictionPolicy: evictionPolicyLRU,
		},
		{
			name:           "LFUCache",
			evictionPolicy: evictionPolicyLFU,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			cache := newInMemoryCache[string](tc.name, true, 100, time.Second*60, tc.evictionPolicy)

			// Test Set and Get operations
			key := CacheKey{Key: "testKey"}

			err := cache.Set(key, testValue)
			assert.NoError(t, err)

			retrievedValue, found := cache.Get(key)
			assert.True(t, found)
			assert.Equal(t, testValue, retrievedValue)

			// Check cache stats
			stats := cache.GetStats()
			assert.Equal(t, int64(1), stats.HitCount)
			assert.Equal(t, int64(0), stats.MissCount)
			assert.Equal(t, float64(1.0), stats.HitRate)
			assert.Equal(t, 1, stats.Size)

			// Test getting a non-existent key
			nonExistentKey := CacheKey{Key: "nonExistentKey"}
			_, found = cache.Get(nonExistentKey)
			assert.False(t, found)

			// Check updated stats
			stats = cache.GetStats()
			assert.Equal(t, int64(1), stats.HitCount)
			assert.Equal(t, int64(1), stats.MissCount)
			assert.Equal(t, 0.5, stats.HitRate)
		})
	}
}

func (suite *InMemoryCacheTestSuite) TestDelete() {
	cache := newInMemoryCache[string]("testCache", true, 100, time.Second*60, evictionPolicyLRU)

	// Add an entry and verify it exists
	key := CacheKey{Key: "testKey"}

	err := cache.Set(key, testValue)
	assert.NoError(suite.T(), err)

	_, found := cache.Get(key)
	assert.True(suite.T(), found)

	// Delete the entry and verify it no longer exists
	err = cache.Delete(key)
	assert.NoError(suite.T(), err)

	_, found = cache.Get(key)
	assert.False(suite.T(), found)

	// Verify cache size is now 0
	stats := cache.GetStats()
	assert.Equal(suite.T(), 0, stats.Size)
}

func (suite *InMemoryCacheTestSuite) TestClear() {
	cache := newInMemoryCache[string]("testCache", true, 100, time.Second*60, evictionPolicyLRU)

	// Add multiple entries
	for i := 0; i < 5; i++ {
		key := CacheKey{Key: "testKey" + string(rune('0'+i))}
		value := testValue + string(rune('0'+i))
		err := cache.Set(key, value)
		assert.NoError(suite.T(), err)
	}

	// Verify cache has entries
	stats := cache.GetStats()
	assert.Equal(suite.T(), 5, stats.Size)

	// Clear the cache
	err := cache.Clear()
	assert.NoError(suite.T(), err)

	// Verify cache is empty
	stats = cache.GetStats()
	assert.Equal(suite.T(), 0, stats.Size)
	assert.Equal(suite.T(), int64(0), stats.HitCount)
	assert.Equal(suite.T(), int64(0), stats.MissCount)
}

func (suite *InMemoryCacheTestSuite) TestExpiry() {
	// Create cache with very short TTL
	cache := newInMemoryCache[string]("testCache", true, 100, time.Millisecond*50, evictionPolicyLRU)

	// Add an entry
	key := CacheKey{Key: "testKey"}

	err := cache.Set(key, testValue)
	assert.NoError(suite.T(), err)

	// Verify it exists initially
	retrievedValue, found := cache.Get(key)
	assert.True(suite.T(), found)
	assert.Equal(suite.T(), testValue, retrievedValue)

	// Wait for expiration
	time.Sleep(time.Millisecond * 100)

	// Get should now return not found
	_, found = cache.Get(key)
	assert.False(suite.T(), found)
}

func (suite *InMemoryCacheTestSuite) TestCleanupExpired() {
	// Create cache with very short TTL
	cache := newInMemoryCache[string]("testCache", true, 100, time.Millisecond*50, evictionPolicyLRU)

	// Add multiple entries
	for i := 0; i < 5; i++ {
		key := CacheKey{Key: "testKey" + string(rune('0'+i))}
		value := testValue + string(rune('0'+i))
		err := cache.Set(key, value)
		assert.NoError(suite.T(), err)
	}

	// Verify cache has entries
	stats := cache.GetStats()
	assert.Equal(suite.T(), 5, stats.Size)

	// Wait for expiration
	time.Sleep(time.Millisecond * 100)

	// Manually trigger cleanup
	cache.CleanupExpired()

	// Verify cache is now empty
	stats = cache.GetStats()
	assert.Equal(suite.T(), 0, stats.Size)
}

func (suite *InMemoryCacheTestSuite) TestEviction() {
	testCases := []struct {
		name           string
		evictionPolicy evictionPolicy
	}{
		{
			name:           "LRUEviction",
			evictionPolicy: evictionPolicyLRU,
		},
		{
			name:           "LFUEviction",
			evictionPolicy: evictionPolicyLFU,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Create small cache (size 3)
			cache := newInMemoryCache[string](tc.name, true, 3, time.Second*60, tc.evictionPolicy)

			// Add 3 entries (fill the cache)
			for i := 0; i < 3; i++ {
				key := CacheKey{Key: "testKey" + string(rune('0'+i))}
				value := testValue + string(rune('0'+i))
				err := cache.Set(key, value)
				assert.NoError(t, err)
			}

			// If LFU, access key0 multiple times to increase its frequency
			if tc.evictionPolicy == evictionPolicyLFU {
				for i := 0; i < 3; i++ {
					_, _ = cache.Get(CacheKey{Key: "testKey0"})
				}
			}

			// For LRU, access key0 to make it most recently used
			if tc.evictionPolicy == evictionPolicyLRU {
				_, _ = cache.Get(CacheKey{Key: "testKey0"})
			}

			// Add a new entry to trigger eviction
			newKey := CacheKey{Key: "testKey3"}
			newValue := testValue + "3"
			err := cache.Set(newKey, newValue)
			assert.NoError(t, err)

			// Check eviction occurred
			stats := cache.GetStats()
			assert.Equal(t, 3, stats.Size)
			assert.Equal(t, int64(1), stats.EvictCount)

			// If LRU, key1 should be evicted as it was least recently used
			// If LFU, key1 or key2 should be evicted as they have lower frequency than key0
			if tc.evictionPolicy == evictionPolicyLRU {
				_, found := cache.Get(CacheKey{Key: "testKey1"})
				assert.False(t, found, "Expected key1 to be evicted in LRU cache")
			}

			// Verify new key exists
			retrievedValue, found := cache.Get(newKey)
			assert.True(t, found)
			assert.Equal(t, newValue, retrievedValue)

			// Verify key0 still exists (it had higher frequency or was more recently used)
			_, found = cache.Get(CacheKey{Key: "testKey0"})
			assert.True(t, found, "Expected key0 to remain in cache")
		})
	}
}

func (suite *InMemoryCacheTestSuite) TestDisabledCache() {
	cache := newInMemoryCache[string]("testCache", false, 100, time.Second*60, evictionPolicyLRU)

	// Verify cache is disabled
	assert.False(suite.T(), cache.IsEnabled())

	// Operations on disabled cache should be no-ops
	key := CacheKey{Key: "testKey"}

	// Set should not error but not actually store
	err := cache.Set(key, testValue)
	assert.NoError(suite.T(), err)

	// Get should return not found
	_, found := cache.Get(key)
	assert.False(suite.T(), found)

	// Delete should not error
	err = cache.Delete(key)
	assert.NoError(suite.T(), err)

	// Clear should not error
	err = cache.Clear()
	assert.NoError(suite.T(), err)

	// CleanupExpired should not error
	cache.CleanupExpired()

	// GetStats should return disabled status
	stats := cache.GetStats()
	assert.False(suite.T(), stats.Enabled)
}

func (suite *InMemoryCacheTestSuite) TestGetName() {
	// Test with various cache name configurations
	testCases := []struct {
		name      string
		cacheName string
		enabled   bool
	}{
		{
			name:      "EnabledCache",
			cacheName: "testEnabledCache",
			enabled:   true,
		},
		{
			name:      "DisabledCache",
			cacheName: "testDisabledCache",
			enabled:   false,
		},
		{
			name:      "EmptyNameCache",
			cacheName: "",
			enabled:   true,
		},
		{
			name:      "SpecialCharsNameCache",
			cacheName: "test-cache_123:special",
			enabled:   true,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			cache := newInMemoryCache[string](tc.cacheName, tc.enabled, 100, time.Second*60,
				evictionPolicyLRU)

			// Verify the cache name matches what was provided during creation
			assert.Equal(t, tc.cacheName, cache.GetName())
		})
	}
}

func (suite *InMemoryCacheTestSuite) TestGetStats() {
	// Test that GetStats returns the correct statistics for an enabled cache
	suite.T().Run("EnabledCacheStats", func(t *testing.T) {
		cache := newInMemoryCache[string]("statsTestCache", true, 100, time.Second*60,
			evictionPolicyLRU)

		// Initial stats should show an empty cache
		initialStats := cache.GetStats()
		assert.True(t, initialStats.Enabled)
		assert.Equal(t, 0, initialStats.Size)
		assert.Equal(t, 100, initialStats.MaxSize)
		assert.Equal(t, int64(0), initialStats.HitCount)
		assert.Equal(t, int64(0), initialStats.MissCount)
		assert.Equal(t, float64(0), initialStats.HitRate)
		assert.Equal(t, int64(0), initialStats.EvictCount)

		// Add entries and perform operations to change stats
		for i := 0; i < 5; i++ {
			key := CacheKey{Key: "key" + string(rune('0'+i))}
			value := "value" + string(rune('0'+i))
			err := cache.Set(key, value)
			assert.NoError(t, err)
		}

		// Get some existing entries (hits)
		_, _ = cache.Get(CacheKey{Key: "key0"})
		_, _ = cache.Get(CacheKey{Key: "key1"})
		_, _ = cache.Get(CacheKey{Key: "key2"})

		// Get some non-existing entries (misses)
		_, _ = cache.Get(CacheKey{Key: "nonexistent1"})
		_, _ = cache.Get(CacheKey{Key: "nonexistent2"})

		// Check updated stats
		updatedStats := cache.GetStats()
		assert.True(t, updatedStats.Enabled)
		assert.Equal(t, 5, updatedStats.Size)
		assert.Equal(t, 100, updatedStats.MaxSize)
		assert.Equal(t, int64(3), updatedStats.HitCount)
		assert.Equal(t, int64(2), updatedStats.MissCount)
		assert.Equal(t, float64(0.6), updatedStats.HitRate) // 3 hits / 5 total operations = 0.6
		assert.Equal(t, int64(0), updatedStats.EvictCount)

		// Fill the cache beyond capacity to trigger eviction
		for i := 0; i < 100; i++ {
			key := CacheKey{Key: "evictionKey" + string(rune('0'+i))}
			value := "evictionValue" + string(rune('0'+i))
			err := cache.Set(key, value)
			assert.NoError(t, err)
		}

		// Check that eviction count is now greater than 0
		evictionStats := cache.GetStats()
		assert.True(t, evictionStats.EvictCount > 0)
	})

	// Test GetStats with a disabled cache
	suite.T().Run("DisabledCacheStats", func(t *testing.T) {
		cache := newInMemoryCache[string]("disabledStatsCache", false, 100, time.Second*60,
			evictionPolicyLRU)

		stats := cache.GetStats()
		assert.False(t, stats.Enabled)
		// Other stats fields should be their zero values for a disabled cache
		assert.Equal(t, 0, stats.Size)
		assert.Equal(t, 0, stats.MaxSize)
		assert.Equal(t, int64(0), stats.HitCount)
		assert.Equal(t, int64(0), stats.MissCount)
		assert.Equal(t, float64(0), stats.HitRate)
		assert.Equal(t, int64(0), stats.EvictCount)
	})
}
