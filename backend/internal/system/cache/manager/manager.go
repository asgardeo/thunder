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

// Package manager provides the cache manager that orchestrates multiple cache levels.
package manager

import (
	"context"
	"sync"
	"time"

	"github.com/asgardeo/thunder/internal/system/cache/constants"
	"github.com/asgardeo/thunder/internal/system/cache/l1cache"
	"github.com/asgardeo/thunder/internal/system/cache/l2cache"
	"github.com/asgardeo/thunder/internal/system/cache/model"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/log"
)

const loggerComponentName = "CacheManager"

// CacheManagerInterface defines the interface for cache manager.
type CacheManagerInterface interface {
	Set(key model.CacheKey, value interface{}) error
	Get(key model.CacheKey) (interface{}, bool)
	Delete(key model.CacheKey) error
	Clear() error
	IsEnabled() bool
	Shutdown()
}

// CacheManager implements the CacheManagerInterface for managing caches.
type CacheManager struct {
	enabled          bool
	l1Cache          model.CacheInterface
	l2Cache          model.CacheInterface
	mu               sync.RWMutex
	promotionChannel chan model.PromotionTask
	promotionWg      sync.WaitGroup
	promotionCtx     context.Context
	promotionCancel  context.CancelFunc
}

// NewCacheManager creates a new cache manager instance.
func NewCacheManager() CacheManagerInterface {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	// TODO: Add check for validating L2 cache configuration when implemented
	cacheConfig := config.GetThunderRuntime().Config.Cache
	if !cacheConfig.L1.Enabled {
		logger.Debug("Cache system is disabled")
		return &CacheManager{
			enabled: false,
			l1Cache: l1cache.NewL1Cache(false, 0, 0, ""),
			l2Cache: l2cache.NewL2Cache(false),
		}
	}

	logger.Debug("Initializing cache manager")

	maxSize := cacheConfig.L1.MaxSize
	if maxSize <= 0 {
		maxSize = constants.L1DefaultMaxSize
	}

	defaultTTL := cacheConfig.L1.DefaultTTL
	if defaultTTL <= 0 {
		defaultTTL = constants.L1DefaultTTL
	}

	l1Cache := l1cache.NewL1Cache(
		cacheConfig.L1.Enabled,
		maxSize,
		time.Duration(defaultTTL)*time.Second,
		cacheConfig.L1.EvictionPolicy,
	)
	l2Cache := l2cache.NewL2Cache(false)

	// Initialize worker pool for cache promotion
	promotionCtx, promotionCancel := context.WithCancel(context.Background())
	promotionChannel := make(chan model.PromotionTask, constants.DefaultPromotionChannelBuffer)

	cm := &CacheManager{
		enabled:          true,
		l1Cache:          l1Cache,
		l2Cache:          l2Cache,
		promotionChannel: promotionChannel,
		promotionCtx:     promotionCtx,
		promotionCancel:  promotionCancel,
	}

	// Start cleanup routine
	cm.startCleanupRoutine()

	// Start worker pool for cache promotion
	cm.startPromotionWorkers()

	return cm
}

// Set stores a value in the cache with default TTL.
func (cm *CacheManager) Set(key model.CacheKey, value interface{}) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if !cm.enabled {
		return nil
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Write to L1 cache
	if cm.l1Cache.IsEnabled() {
		if err := cm.l1Cache.Set(key, value); err != nil {
			logger.Warn("Failed to set value in L1 cache", log.String("key", key.ToString()), log.Error(err))
		}
	}

	// Write to L2 cache
	if cm.l2Cache.IsEnabled() {
		if err := cm.l2Cache.Set(key, value); err != nil {
			logger.Warn("Failed to set value in L2 cache", log.String("key", key.ToString()), log.Error(err))
		}
	}

	return nil
}

// Get retrieves a value from the cache.
func (cm *CacheManager) Get(key model.CacheKey) (interface{}, bool) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if !cm.enabled {
		return nil, false
	}

	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Try L1 cache first
	if cm.l1Cache.IsEnabled() {
		if value, found := cm.l1Cache.Get(key); found {
			return value, true
		}
	}

	// Try L2 cache if L1 miss
	if cm.l2Cache.IsEnabled() {
		if value, found := cm.l2Cache.Get(key); found {
			// Populate L1 cache with the value from L2 (cache promotion)
			if cm.l1Cache.IsEnabled() && config.GetThunderRuntime().Config.Cache.L1.EnablePromotion {
				select {
				case cm.promotionChannel <- model.PromotionTask{Key: key, Value: value}:
				default:
					logger.Debug("Promotion channel full, skipping cache promotion",
						log.String("key", key.ToString()))
				}
			}
			return value, true
		}
	}

	return nil, false
}

// Delete removes a value from the cache.
func (cm *CacheManager) Delete(key model.CacheKey) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if !cm.enabled {
		return nil
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Delete from L1 cache
	if cm.l1Cache.IsEnabled() {
		if err := cm.l1Cache.Delete(key); err != nil {
			logger.Warn("Failed to delete value from L1 cache", log.String("key", key.ToString()), log.Error(err))
		}
	}

	// Delete from L2 cache
	if cm.l2Cache.IsEnabled() {
		if err := cm.l2Cache.Delete(key); err != nil {
			logger.Warn("Failed to delete value from L2 cache", log.String("key", key.ToString()), log.Error(err))
		}
	}

	return nil
}

// Clear removes all entries in the cache.
func (cm *CacheManager) Clear() error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if !cm.enabled {
		return nil
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Clear L1 cache
	if cm.l1Cache.IsEnabled() {
		if err := cm.l1Cache.Clear(); err != nil {
			logger.Warn("Failed to clear L1 cache", log.Error(err))
		}
	}

	// Clear L2 cache
	if cm.l2Cache.IsEnabled() {
		if err := cm.l2Cache.Clear(); err != nil {
			logger.Warn("Failed to clear L2 cache", log.Error(err))
		}
	}

	logger.Debug("Cache is cleared")
	return nil
}

// IsEnabled returns whether the cache manager is enabled.
func (cm *CacheManager) IsEnabled() bool {
	return cm.enabled
}

// startCleanupRoutine starts a background routine to clean up expired entries.
func (cm *CacheManager) startCleanupRoutine() {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if !cm.enabled {
		return
	}

	cleanupInterval := config.GetThunderRuntime().Config.Cache.CleanupInterval
	if cleanupInterval == -1 {
		logger.Warn("Cache cleanup routine is disabled")
		return
	} else if cleanupInterval <= 0 {
		cleanupInterval = constants.DefaultCleanupInterval
	}

	go func() {
		ticker := time.NewTicker(time.Duration(cleanupInterval) * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			// Clean up L1 cache expired entries
			if l1Cache, ok := cm.l1Cache.(*l1cache.L1Cache); ok && l1Cache.IsEnabled() {
				l1Cache.CleanupExpired()
			}
		}
	}()

	logger.Debug("Cache cleanup routine started", log.Any("interval", cleanupInterval))
}

// startPromotionWorkers starts the worker pool for cache promotion tasks.
func (cm *CacheManager) startPromotionWorkers() {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if !cm.enabled || !config.GetThunderRuntime().Config.Cache.L1.EnablePromotion {
		logger.Debug("Cache promotion is disabled or cache manager is not enabled, skipping worker pool startup")
		return
	}

	workerCount := constants.DefaultPromotionWorkerPoolSize
	logger.Debug("Starting cache promotion worker pool", log.Any("workers", workerCount))

	for i := 0; i < workerCount; i++ {
		cm.promotionWg.Add(1)
		go cm.promotionWorker(i)
	}
}

// promotionWorker processes cache promotion tasks from the channel.
func (cm *CacheManager) promotionWorker(workerID int) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.Int("workerId", workerID))

	defer cm.promotionWg.Done()

	for {
		select {
		case <-cm.promotionCtx.Done():
			logger.Debug("Promotion worker stopping due to context cancellation")
			return
		case task, ok := <-cm.promotionChannel:
			if !ok {
				logger.Debug("Promotion channel closed, worker stopping")
				return
			}

			if cm.l1Cache.IsEnabled() {
				if err := cm.l1Cache.Set(task.Key, task.Value); err != nil {
					logger.Debug("Failed to promote value from L2 to L1",
						log.String("key", task.Key.ToString()), log.Error(err))
				}
			}
		}
	}
}

// Shutdown gracefully shuts down the cache manager and its worker pool.
func (cm *CacheManager) Shutdown() {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if !cm.enabled {
		logger.Debug("Cache manager is disabled, nothing to shutdown")
		return
	}

	logger.Debug("Shutting down cache manager")

	// Cancel context to signal workers to stop
	if cm.promotionCancel != nil {
		cm.promotionCancel()
	}

	// Close the promotion channel to signal no more tasks
	if cm.promotionChannel != nil {
		close(cm.promotionChannel)
	}

	// Wait for all workers to finish
	cm.promotionWg.Wait()

	logger.Debug("Cache manager shutdown complete")
}
