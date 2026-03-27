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

package authz

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/utils"
)

// redisAuthorizationRequestStore is the Redis-backed implementation of authorizationRequestStoreInterface.
type redisAuthorizationRequestStore struct {
	client         *redis.Client
	keyPrefix      string
	deploymentID   string
	validityPeriod time.Duration
}

// newRedisAuthorizationRequestStore creates a new Redis-backed authorization request store.
func newRedisAuthorizationRequestStore(p provider.RedisClientInterface) authorizationRequestStoreInterface {
	return &redisAuthorizationRequestStore{
		client:         p.GetRedisClient(),
		keyPrefix:      p.GetKeyPrefix(),
		deploymentID:   config.GetThunderRuntime().Config.Server.Identifier,
		validityPeriod: 10 * time.Minute,
	}
}

// reqKey builds the Redis key for an authorization request.
func (s *redisAuthorizationRequestStore) reqKey(key string) string {
	return fmt.Sprintf("%s:runtime:%s:authreq:%s", s.keyPrefix, s.deploymentID, key)
}

// AddRequest stores an authorization request context in Redis and returns the generated key.
func (s *redisAuthorizationRequestStore) AddRequest(
	ctx context.Context, value authRequestContext,
) (string, error) {
	key, err := utils.GenerateUUIDv7()
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID: %w", err)
	}

	data, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("failed to marshal authorization request: %w", err)
	}

	if err := s.client.Set(ctx, s.reqKey(key), data, s.validityPeriod).Err(); err != nil {
		return "", fmt.Errorf("failed to store authorization request in Redis: %w", err)
	}

	return key, nil
}

// GetRequest retrieves an authorization request context by key.
func (s *redisAuthorizationRequestStore) GetRequest(
	ctx context.Context, key string,
) (bool, authRequestContext, error) {
	if key == "" {
		return false, authRequestContext{}, nil
	}

	data, err := s.client.Get(ctx, s.reqKey(key)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, authRequestContext{}, nil
		}
		return false, authRequestContext{}, fmt.Errorf("failed to get authorization request from Redis: %w", err)
	}

	var reqCtx authRequestContext
	if err := json.Unmarshal(data, &reqCtx); err != nil {
		return false, authRequestContext{}, fmt.Errorf("failed to unmarshal authorization request: %w", err)
	}

	return true, reqCtx, nil
}

// ClearRequest removes an authorization request from Redis.
func (s *redisAuthorizationRequestStore) ClearRequest(ctx context.Context, key string) error {
	if key == "" {
		return nil
	}

	if err := s.client.Del(ctx, s.reqKey(key)).Err(); err != nil {
		return fmt.Errorf("failed to delete authorization request from Redis: %w", err)
	}

	return nil
}
