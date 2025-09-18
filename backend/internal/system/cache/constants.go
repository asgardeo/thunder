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

// evictionPolicy defines the eviction policy for cache entries.
type evictionPolicy string

const (
	// evictionPolicyLRU represents the Least Recently Used eviction policy.
	evictionPolicyLRU evictionPolicy = "LRU"
	// evictionPolicyLFU represents the Least Frequently Used eviction policy.
	evictionPolicyLFU evictionPolicy = "LFU"
)

// cacheType defines the type of cache being used.
type cacheType string

const (
	// cacheTypeInMemory represents an in-memory cache type.
	cacheTypeInMemory cacheType = "inmemory"
)
