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
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/config"
)

type TestString string
type TestInt int

type CacheManagerTestSuite struct {
	suite.Suite
}

func TestCacheManagerSuite(t *testing.T) {
	suite.Run(t, new(CacheManagerTestSuite))
}

func (suite *CacheManagerTestSuite) SetupSuite() {
	mockConfig := &config.Config{
		Cache: config.CacheConfig{
			Disabled: true, // Disable cache globally for tests
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/test/thunder/home", mockConfig)
	if err != nil {
		suite.T().Fatal("Failed to initialize ThunderRuntime:", err)
	}
}

func (suite *CacheManagerTestSuite) TearDownSuite() {
	config.ResetThunderRuntime()
}

func (suite *CacheManagerTestSuite) SetupTest() {
	// Reset the singleton instance for each test
	instance = nil
	once = sync.Once{}
}

func (suite *CacheManagerTestSuite) TestGetCacheManager() {
	t := suite.T()

	// Get cache manager instance
	manager1 := GetCacheManager()
	assert.NotNil(t, manager1, "Cache manager should not be nil")
	assert.IsType(t, &CacheManager{}, manager1, "Cache manager should be of type *CacheManager")

	// Get another instance to verify singleton pattern
	manager2 := GetCacheManager()
	assert.Same(t, manager1, manager2, "GetCacheManager should return the same instance (singleton)")
}

func (suite *CacheManagerTestSuite) TestCacheManagerInit() {
	t := suite.T()

	// Test with cache disabled
	manager := &CacheManager{
		caches: make(map[string]interface{}),
	}
	manager.Init()
	assert.False(t, manager.IsEnabled(), "Cache should be disabled")

	// Test with cache enabled
	enabledConfig := &config.Config{
		Cache: config.CacheConfig{
			Disabled:        false,
			Size:            1000,
			TTL:             3600,
			EvictionPolicy:  "LRU",
			CleanupInterval: 60,
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/test/thunder/home", enabledConfig)
	assert.NoError(t, err)

	manager = &CacheManager{
		caches: make(map[string]interface{}),
	}
	manager.Init()
	assert.True(t, manager.IsEnabled(), "Cache should be enabled")
	assert.Equal(t, time.Duration(60)*time.Second, manager.cleanupInterval, "Cleanup interval should be set")
}

func (suite *CacheManagerTestSuite) TestGetMutex() {
	t := suite.T()

	manager := &CacheManager{
		caches: make(map[string]interface{}),
	}

	// Verify getMutex returns the expected mutex
	mu := manager.getMutex()
	assert.NotNil(t, mu, "getMutex should return a non-nil mutex")
	assert.IsType(t, &sync.RWMutex{}, mu, "getMutex should return a pointer to sync.RWMutex")
}

func (suite *CacheManagerTestSuite) TestAddAndGetCache() {
	t := suite.T()

	manager := &CacheManager{
		caches: make(map[string]interface{}),
	}

	// Test adding a cache
	mockCache := NewCacheInterfaceMock[any](t)
	cacheKey := "testCacheKey"

	manager.addCache(cacheKey, mockCache)

	// Verify it was added
	cacheInstance, exists := manager.getCache(cacheKey)
	assert.True(t, exists, "Cache should exist after adding")
	assert.Same(t, mockCache, cacheInstance, "Should return the same cache instance")

	// Add the same cache again (should be a no-op)
	manager.addCache(cacheKey, mockCache)

	// Test getting a non-existent cache
	_, exists = manager.getCache("nonExistentKey")
	assert.False(t, exists, "Non-existent cache should return false")
}

func (suite *CacheManagerTestSuite) TestGetCache() {
	t := suite.T()

	// Get a cache for string type
	cacheName := "testCache"
	cache1 := GetCache[string](cacheName)
	assert.NotNil(t, cache1, "Cache should not be nil")

	// Get the same cache again
	cache2 := GetCache[string](cacheName)
	assert.Same(t, cache1, cache2, "GetCache should return the same instance for the same type and name")

	// Test with a different cache name but same type
	differentCacheName := "anotherCache"
	cache3 := GetCache[string](differentCacheName)
	assert.NotNil(t, cache3, "Cache should not be nil")
	assert.NotSame(t, cache1, cache3, "Different cache names should create different caches")
}

func (suite *CacheManagerTestSuite) TestGetCacheMultipleTypes() {
	t := suite.T()

	cacheName := "testMultiTypeCache"

	// Get caches for different types
	cacheString := GetCache[string](cacheName)
	cacheInt := GetCache[int](cacheName)
	cacheTestString := GetCache[TestString](cacheName)
	cacheTestInt := GetCache[TestInt](cacheName)

	// Verify all caches are not nil
	assert.NotNil(t, cacheString, "String cache should not be nil")
	assert.NotNil(t, cacheInt, "Int cache should not be nil")
	assert.NotNil(t, cacheTestString, "TestString cache should not be nil")
	assert.NotNil(t, cacheTestInt, "TestInt cache should not be nil")

	// Verify different types create different instances even with the same cache name
	assert.NotSame(t, cacheString, cacheInt, "Different types should create different caches")
	assert.NotSame(t, cacheString, cacheTestString, "Different types should create different caches")
	assert.NotSame(t, cacheInt, cacheTestInt, "Different types should create different caches")
	assert.NotSame(t, cacheTestString, cacheTestInt, "Different types should create different caches")

	// Verify same type returns same instance
	cacheStringSame := GetCache[string](cacheName)
	assert.Same(t, cacheString, cacheStringSame, "Same type and name should return the same cache")
}

func (suite *CacheManagerTestSuite) TestResetCacheManager() {
	t := suite.T()

	// Create a cache
	cacheName := "testResetCache"
	cm := GetCache[string](cacheName)
	assert.NotNil(t, cm, "Cache should not be nil")

	// Verify cache manager has an entry
	manager := GetCacheManager().(*CacheManager)
	assert.NotEmpty(t, manager.caches, "Cache map should not be empty after creating a cache")

	// Reset the cache manager
	manager.reset()

	// Verify cache manager is now empty
	assert.Empty(t, manager.caches, "Cache map should be empty after reset")

	// Create a new cache and verify it's different
	cmNew := GetCache[string](cacheName)
	assert.NotNil(t, cmNew, "New cache should not be nil")
	assert.NotSame(t, cm, cmNew, "After reset, should get a new cache instance")
}

func (suite *CacheManagerTestSuite) TestCleanupAllCaches() {
	t := suite.T()

	// Create mock caches
	cacheName1 := "testCleanupCache1"
	cacheName2 := "testCleanupCache2"

	mockCache1 := NewCacheInterfaceMock[any](t)
	mockCache1.EXPECT().IsEnabled().Return(true)
	mockCache1.EXPECT().GetName().Return(cacheName1)
	mockCache1.EXPECT().CleanupExpired().Once()

	mockCache2 := NewCacheInterfaceMock[any](t)
	mockCache2.EXPECT().IsEnabled().Return(true)
	mockCache2.EXPECT().GetName().Return(cacheName2)
	mockCache2.EXPECT().CleanupExpired().Once()

	// Add mocks to the manager
	manager := &CacheManager{
		caches: make(map[string]interface{}),
	}
	manager.caches["testkey1"] = mockCache1
	manager.caches["testkey2"] = mockCache2

	// Call cleanupAllCaches
	manager.cleanupAllCaches()

	// Assertions are handled by the mock expectations
}

func (suite *CacheManagerTestSuite) TestConcurrentAccess() {
	t := suite.T()

	// Setup enabled config
	enabledConfig := &config.Config{
		Cache: config.CacheConfig{
			Disabled: false,
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/test/thunder/home", enabledConfig)
	assert.NoError(t, err)

	// Reset cache manager
	instance = nil
	once = sync.Once{}

	// Number of goroutines to use
	numGoroutines := 10
	done := make(chan bool, numGoroutines)
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Create multiple caches concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()
			// Use different cache names to avoid collisions
			cacheName := "concurrentCache" + string(rune('A'+index))
			cache := GetCache[string](cacheName)
			assert.NotNil(t, cache, "Cache should not be nil even with concurrent access")
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(done)

	// Count completed goroutines
	completedCount := 0
	for range done {
		completedCount++
	}

	// Verify all goroutines completed successfully
	assert.Equal(t, numGoroutines, completedCount, "All goroutines should complete successfully")

	// Verify manager has the expected number of entries
	manager := GetCacheManager().(*CacheManager)
	assert.Equal(t, numGoroutines, len(manager.caches), "Cache map should have an entry for each goroutine")
}

func (suite *CacheManagerTestSuite) TestTypeMismatch() {
	t := suite.T()

	cacheName := "typeMismatchCache"

	// Create a cache manager and manually inject a cache with mismatched type
	manager := GetCacheManager().(*CacheManager)

	var mockCache interface{} = &Cache[int]{} // Int type
	typeName := "string"
	cacheKey := cacheName + ":" + typeName

	// Directly inject the mismatched type
	manager.mu.Lock()
	manager.caches[cacheKey] = mockCache
	manager.mu.Unlock()

	// Try to get a string cache
	cache := GetCache[string](cacheName)
	assert.Nil(t, cache, "Should return nil when there's a type mismatch")
}

func (suite *CacheManagerTestSuite) TestNewCache() {
	t := suite.T()

	// Save and restore original config
	originalConfig := config.GetThunderRuntime().Config
	defer func() {
		// Reset config to original
		config.ResetThunderRuntime()
		err := config.InitializeThunderRuntime("/test/thunder/home", &originalConfig)
		assert.NoError(t, err)
	}()

	// Test 1: Test with cache globally disabled
	disabledConfig := config.Config{
		Cache: config.CacheConfig{
			Disabled: true,
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/test/thunder/home", &disabledConfig)
	assert.NoError(t, err)

	cache1 := newCache[string]("testDisabledCache")
	assert.NotNil(t, cache1)
	assert.False(t, cache1.IsEnabled())

	// Test 2: Test with specific cache disabled
	enabledConfig := config.Config{
		Cache: config.CacheConfig{
			Disabled: false,
			Properties: []config.CacheProperty{
				{
					Name:     "testSpecificDisabledCache",
					Disabled: true,
				},
			},
		},
	}
	config.ResetThunderRuntime()
	err = config.InitializeThunderRuntime("/test/thunder/home", &enabledConfig)
	assert.NoError(t, err)

	cache2 := newCache[string]("testSpecificDisabledCache")
	assert.NotNil(t, cache2)
	assert.False(t, cache2.IsEnabled())

	// Test 3: Test with in-memory cache type
	inMemConfig := config.Config{
		Cache: config.CacheConfig{
			Disabled: false,
			Type:     "inmemory",
			Properties: []config.CacheProperty{
				{
					Name: "testInMemCache",
					Size: 100,
					TTL:  300,
				},
			},
		},
	}
	config.ResetThunderRuntime()
	err = config.InitializeThunderRuntime("/test/thunder/home", &inMemConfig)
	assert.NoError(t, err)

	cache3 := newCache[string]("testInMemCache")
	assert.NotNil(t, cache3)
	assert.True(t, cache3.IsEnabled())

	// Test 4: Test with unknown cache type
	unknownTypeConfig := config.Config{
		Cache: config.CacheConfig{
			Disabled: false,
			Type:     "unknown-type",
		},
	}
	config.ResetThunderRuntime()
	err = config.InitializeThunderRuntime("/test/thunder/home", &unknownTypeConfig)
	assert.NoError(t, err)

	cache4 := newCache[string]("testUnknownTypeCache")
	assert.NotNil(t, cache4)
	assert.True(t, cache4.IsEnabled())
}

func (suite *CacheManagerTestSuite) TestGetCleanupInterval() {
	t := suite.T()
	config := config.CacheConfig{
		CleanupInterval: 120,
	}
	interval := getCleanupInterval(config)
	assert.Equal(t, time.Duration(120)*time.Second, interval, "Should use configured cleanup interval")
}

func (suite *CacheManagerTestSuite) TestStartCleanupRoutine() {
	t := suite.T()

	// Create a manager with short cleanup interval
	manager := &CacheManager{
		caches:          make(map[string]interface{}),
		enabled:         true,
		cleanupInterval: 50 * time.Millisecond,
	}

	// Create a mock cache
	mockCache := NewCacheInterfaceMock[any](t)
	mockCache.EXPECT().IsEnabled().Return(true).Maybe()
	mockCache.EXPECT().GetName().Return("testCache").Maybe()
	mockCache.EXPECT().CleanupExpired().Maybe()

	// Add mock to manager
	manager.caches["testkey"] = mockCache

	// Start cleanup routine
	manager.startCleanupRoutine()

	// Sleep to allow cleanup to run
	time.Sleep(100 * time.Millisecond)

	// No assertion needed as we're just testing that it doesn't crash
	// The mock is configured with Maybe() since we can't predict exactly how many times
	// the cleanup routine will execute in the short time window
}
