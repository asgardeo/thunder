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

// Package l2cache provides the implementation for L2 caching using external systems like Redis.
package l2cache

import (
	"github.com/asgardeo/thunder/internal/system/cache/model"
	"github.com/asgardeo/thunder/internal/system/log"
)

const loggerComponentName = "L2Cache"

// L2Cache implements the L2CacheInterface for distributed caching.
type L2Cache[T any] struct {
	enabled bool
}

// NewL2Cache creates a new instance of L2Cache.
func NewL2Cache[T any](enabled bool) model.CacheInterface[T] {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if !enabled {
		logger.Debug("L2 cache is disabled")
		return &L2Cache[T]{enabled: false}
	}

	logger.Warn("L2 cache is enabled but not yet implemented")
	// TODO: Initialize external cache connection
	return &L2Cache[T]{enabled: false}
}

// Set adds or updates an entry in the L2 cache.
func (l2 *L2Cache[T]) Set(key model.CacheKey, value T) error {
	if !l2.enabled {
		return nil
	}
	return nil
}

// Get retrieves a value from the L2 cache.
func (l2 *L2Cache[T]) Get(key model.CacheKey) (T, bool) {
	if !l2.enabled {
		var zero T
		return zero, false
	}
	var zero T
	return zero, false
}

// Delete removes an entry from the L2 cache.
func (l2 *L2Cache[T]) Delete(key model.CacheKey) error {
	if !l2.enabled {
		return nil
	}
	return nil
}

// Clear removes all entries from the L2 cache.
func (l2 *L2Cache[T]) Clear() error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if !l2.enabled {
		return nil
	}
	logger.Debug("L2 cache cleared")
	return nil
}

// IsEnabled returns whether the L2 cache is enabled.
func (l2 *L2Cache[T]) IsEnabled() bool {
	return l2.enabled
}

// GetStats returns L2 cache statistics.
func (l2 *L2Cache[T]) GetStats() model.CacheStat {
	if !l2.enabled {
		return model.CacheStat{Enabled: false}
	}
	// TODO: Implement external cache INFO operation to get stats
	return model.CacheStat{
		Enabled:    true,
		Size:       0,
		MaxSize:    -1,
		HitCount:   0,
		MissCount:  0,
		HitRate:    0,
		EvictCount: 0,
	}
}
