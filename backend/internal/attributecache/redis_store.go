/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

package attributecache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/asgardeo/thunder/internal/system/config"
	dbprovider "github.com/asgardeo/thunder/internal/system/database/provider"
)

// redisAttributeCacheStore is the Redis-backed implementation of attributeCacheStoreInterface.
type redisAttributeCacheStore struct {
	client       *redis.Client
	keyPrefix    string
	deploymentID string
}

// newRedisAttributeCacheStore creates a new Redis-backed attribute cache store.
func newRedisAttributeCacheStore(p dbprovider.RedisClientInterface) attributeCacheStoreInterface {
	return &redisAttributeCacheStore{
		client:       p.GetRedisClient(),
		keyPrefix:    p.GetKeyPrefix(),
		deploymentID: config.GetThunderRuntime().Config.Server.Identifier,
	}
}

// cacheKey builds the Redis key for an attribute cache entry.
func (s *redisAttributeCacheStore) cacheKey(id string) string {
	return fmt.Sprintf("%s:runtime:%s:attrcache:%s", s.keyPrefix, s.deploymentID, id)
}

// CreateAttributeCache stores the attribute cache entry in Redis with a TTL.
func (s *redisAttributeCacheStore) CreateAttributeCache(ctx context.Context, cache AttributeCache) error {
	data, err := json.Marshal(cache)
	if err != nil {
		return fmt.Errorf("failed to marshal attribute cache: %w", err)
	}

	ttl := time.Duration(cache.TTLSeconds) * time.Second
	if ttl <= 0 {
		ttl = time.Second
	}

	if err := s.client.Set(ctx, s.cacheKey(cache.ID), data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to store attribute cache in Redis: %w", err)
	}

	return nil
}

// GetAttributeCache retrieves an attribute cache entry by ID.
func (s *redisAttributeCacheStore) GetAttributeCache(ctx context.Context, id string) (AttributeCache, error) {
	key := s.cacheKey(id)

	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return AttributeCache{}, errAttributeCacheNotFound
		}
		return AttributeCache{}, fmt.Errorf("failed to get attribute cache from Redis: %w", err)
	}

	var cache AttributeCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return AttributeCache{}, fmt.Errorf("failed to unmarshal attribute cache: %w", err)
	}

	// Refresh TTLSeconds from the actual remaining Redis TTL.
	ttl, err := s.client.TTL(ctx, key).Result()
	if err == nil && ttl > 0 {
		cache.TTLSeconds = int(ttl.Seconds())
	}

	return cache, nil
}

// ExtendAttributeCacheTTL updates the TTL of an existing attribute cache entry.
func (s *redisAttributeCacheStore) ExtendAttributeCacheTTL(ctx context.Context, id string, ttlSeconds int) error {
	ok, err := s.client.Expire(ctx, s.cacheKey(id), time.Duration(ttlSeconds)*time.Second).Result()
	if err != nil {
		return fmt.Errorf("failed to extend attribute cache TTL: %w", err)
	}
	if !ok {
		return errAttributeCacheNotFound
	}

	return nil
}

// DeleteAttributeCache removes an attribute cache entry from Redis.
func (s *redisAttributeCacheStore) DeleteAttributeCache(ctx context.Context, id string) error {
	n, err := s.client.Del(ctx, s.cacheKey(id)).Result()
	if err != nil {
		return fmt.Errorf("failed to delete attribute cache from Redis: %w", err)
	}
	if n == 0 {
		return errAttributeCacheNotFound
	}

	return nil
}
