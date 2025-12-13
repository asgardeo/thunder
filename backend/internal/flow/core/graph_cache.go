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

package core

import (
	"errors"

	"github.com/asgardeo/thunder/internal/system/cache"
)

// GraphCacheInterface defines operations for caching graphs.
type GraphCacheInterface interface {
	Get(flowID string) (GraphInterface, bool)
	Set(flowID string, graph GraphInterface) error
	Invalidate(flowID string) error
}

// graphCache implements GraphCacheInterface.
type graphCache struct {
	cache cache.CacheInterface[*graph]
}

// newGraphCache creates a new instance of graphCache.
func newGraphCache() GraphCacheInterface {
	return &graphCache{
		cache: cache.GetCache[*graph]("FlowGraphCache"),
	}
}

// Get retrieves a graph from cache by flow ID.
func (gc *graphCache) Get(flowID string) (GraphInterface, bool) {
	if flowID == "" {
		return nil, false
	}
	cacheKey := cache.CacheKey{Key: flowID}
	g, ok := gc.cache.Get(cacheKey)
	if ok && g != nil {
		return g, true
	}
	return nil, false
}

// Set stores a graph in cache.
func (gc *graphCache) Set(flowID string, g GraphInterface) error {
	if flowID == "" || g == nil {
		return errors.New("flowID and graph cannot be empty")
	}

	// Cast to concrete type for caching
	concreteGraph, ok := g.(*graph)
	if !ok {
		return errors.New("graph must be of concrete type *graph")
	}

	cacheKey := cache.CacheKey{Key: flowID}
	return gc.cache.Set(cacheKey, concreteGraph)
}

// Invalidate removes a graph from cache.
func (gc *graphCache) Invalidate(flowID string) error {
	if flowID == "" {
		return nil
	}
	cacheKey := cache.CacheKey{Key: flowID}
	return gc.cache.Delete(cacheKey)
}
