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

package flowexec

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/log"
)

// redisFlowStore is the Redis-backed implementation of flowStoreInterface.
type redisFlowStore struct {
	client       *redis.Client
	keyPrefix    string
	deploymentID string
}

// newRedisFlowStore creates a new Redis-backed flow store.
func newRedisFlowStore(p provider.RedisProviderInterface) flowStoreInterface {
	return &redisFlowStore{
		client:       p.GetRedisClient(),
		keyPrefix:    p.GetKeyPrefix(),
		deploymentID: config.GetThunderRuntime().Config.Server.Identifier,
	}
}

// flowKey builds the Redis key for a flow context.
func (s *redisFlowStore) flowKey(flowID string) string {
	return fmt.Sprintf("%s:runtime:%s:flow:%s", s.keyPrefix, s.deploymentID, flowID)
}

// StoreFlowContext serializes the engine context and stores it in Redis with a TTL.
func (s *redisFlowStore) StoreFlowContext(ctx context.Context, engineCtx EngineContext, expirySeconds int64) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "RedisFlowStore"))

	dbModel, err := FromEngineContext(engineCtx)
	if err != nil {
		return fmt.Errorf("failed to convert engine context to db model: %w", err)
	}

	data, err := json.Marshal(dbModel)
	if err != nil {
		return fmt.Errorf("failed to marshal flow context: %w", err)
	}

	ttl := time.Duration(expirySeconds) * time.Second
	if err := s.client.Set(ctx, s.flowKey(engineCtx.FlowID), data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to store flow context in Redis: %w", err)
	}

	logger.Debug("Stored flow context in Redis", log.String("flowID", engineCtx.FlowID))
	return nil
}

// GetFlowContext retrieves the flow context from Redis.
func (s *redisFlowStore) GetFlowContext(ctx context.Context, flowID string) (*FlowContextDB, error) {
	data, err := s.client.Get(ctx, s.flowKey(flowID)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get flow context from Redis: %w", err)
	}

	var result FlowContextDB
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal flow context: %w", err)
	}

	return &result, nil
}

// UpdateFlowContext updates the stored flow context, preserving the remaining TTL.
func (s *redisFlowStore) UpdateFlowContext(ctx context.Context, engineCtx EngineContext) error {
	key := s.flowKey(engineCtx.FlowID)

	// Preserve the remaining TTL.
	ttl, err := s.client.TTL(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to get TTL for flow context: %w", err)
	}
	if ttl <= 0 {
		ttl = 0
	}

	dbModel, err := FromEngineContext(engineCtx)
	if err != nil {
		return fmt.Errorf("failed to convert engine context to db model: %w", err)
	}

	data, err := json.Marshal(dbModel)
	if err != nil {
		return fmt.Errorf("failed to marshal flow context: %w", err)
	}

	if err := s.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to update flow context in Redis: %w", err)
	}

	return nil
}

// DeleteFlowContext removes the flow context from Redis.
func (s *redisFlowStore) DeleteFlowContext(ctx context.Context, flowID string) error {
	if err := s.client.Del(ctx, s.flowKey(flowID)).Err(); err != nil {
		return fmt.Errorf("failed to delete flow context from Redis: %w", err)
	}
	return nil
}
