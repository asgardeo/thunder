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

package provider

import (
	"context"
	"sync"

	"github.com/redis/go-redis/v9"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/log"
)

// RedisClientInterface defines the provider for the Redis client.
type RedisClientInterface interface {
	GetRedisClient() *redis.Client
	GetKeyPrefix() string
}

// RedisClientCloser provides a method to close the Redis client. Only the lifecycle manager should use this interface.
type RedisClientCloser interface {
	Close() error
}

type redisClientProvider struct {
	client    *redis.Client
	keyPrefix string
	mu        sync.RWMutex
}

var (
	redisInstance *redisClientProvider
	redisOnce     sync.Once
)

func initRedisClient() {
	redisOnce.Do(func() {
		cfg := config.GetThunderRuntime().Config.Database.Runtime
		if cfg.Type != dataSourceTypeRedis {
			return
		}

		logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "RedisClient"))

		client := redis.NewClient(&redis.Options{
			Addr:     cfg.Address,
			Password: cfg.Password,
			DB:       cfg.DB,
		})

		if err := client.Ping(context.Background()).Err(); err != nil {
			logger.Error("Failed to connect to Redis runtime store", log.Error(err))
			if closeErr := client.Close(); closeErr != nil {
				logger.Warn("Failed to close Redis client after ping failure", log.Error(closeErr))
			}
			return
		}

		logger.Info("Connected to Redis store", log.String("address", cfg.Address))
		redisInstance = &redisClientProvider{
			client:    client,
			keyPrefix: cfg.KeyPrefix,
		}
	})
}

// GetRedisClientProvider returns the Redis client provider.
func GetRedisClientProvider() RedisClientInterface {
	initRedisClient()
	return redisInstance
}

// GetRedisClientCloser returns the Redis client closer interface.
func GetRedisClientCloser() RedisClientCloser {
	if redisInstance == nil {
		return nil
	}
	return redisInstance
}

func (r *redisClientProvider) GetRedisClient() *redis.Client {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.client
}

func (r *redisClientProvider) GetKeyPrefix() string {
	return r.keyPrefix
}

func (r *redisClientProvider) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.client != nil {
		if err := r.client.Close(); err != nil {
			return err
		}
		r.client = nil
	}
	return nil
}
