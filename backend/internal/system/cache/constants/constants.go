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

// Package constants provides common constants used across caching.
package constants

// EvictionPolicy defines the eviction policy for cache entries.
// TODO: Multiple eviction policies are not yet implemented, and currently only support LRU.
type EvictionPolicy string

const (
	// EvictionPolicyLRU represents the Least Recently Used eviction policy.
	EvictionPolicyLRU EvictionPolicy = "LRU"
	// EvictionPolicyLFU represents the Least Frequently Used eviction policy.
	EvictionPolicyLFU EvictionPolicy = "LFU"
)

// CacheType represents the type of cache.
type CacheType string

const (
	// CacheTypeApplication represents the cache type for application data.
	CacheTypeApplication CacheType = "application"
)

const (
	// DefaultCleanupInterval represents the default interval for cleaning up caches.
	DefaultCleanupInterval = 300
	// L1DefaultTTL represents the default TTL for L1 cache entries.
	L1DefaultTTL = 3600
	// L1DefaultMaxSize represents the default maximum size for L1 cache.
	L1DefaultMaxSize = 1000
	// DefaultPromotionWorkerPoolSize represents the default number of workers for cache promotion.
	DefaultPromotionWorkerPoolSize = 5
	// DefaultPromotionChannelBuffer represents the default buffer size for promotion channel.
	DefaultPromotionChannelBuffer = 100
)
