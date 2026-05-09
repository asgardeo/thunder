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
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const (
	redisTestKeyPrefix    = "thunderid"
	redisTestDeploymentID = "test-deployment"
	redisTestCacheID      = "test-cache-id"
)

type RedisAttributeCacheStoreTestSuite struct {
	suite.Suite
	store      *redisAttributeCacheStore
	mockClient *redisClientMock
	ctx        context.Context
	testCache  AttributeCache
	cacheKey   string
}

func TestRedisAttributeCacheStoreSuite(t *testing.T) {
	suite.Run(t, new(RedisAttributeCacheStoreTestSuite))
}

func (suite *RedisAttributeCacheStoreTestSuite) SetupTest() {
	suite.mockClient = newRedisClientMock(suite.T())
	suite.ctx = context.Background()
	suite.testCache = AttributeCache{
		ID:         redisTestCacheID,
		Attributes: map[string]interface{}{"key": "value"},
		TTLSeconds: 3600,
	}
	suite.store = &redisAttributeCacheStore{
		client:       suite.mockClient,
		keyPrefix:    redisTestKeyPrefix,
		deploymentID: redisTestDeploymentID,
	}
	suite.cacheKey = fmt.Sprintf("%s:runtime:%s:attrcache:%s",
		redisTestKeyPrefix, redisTestDeploymentID, redisTestCacheID)
}

// Tests for cacheKey

func (suite *RedisAttributeCacheStoreTestSuite) TestCacheKey() {
	key := suite.store.cacheKey(redisTestCacheID)
	suite.Equal(suite.cacheKey, key)
}

// Tests for CreateAttributeCache

func (suite *RedisAttributeCacheStoreTestSuite) TestCreateAttributeCache_Success() {
	statusCmd := redis.NewStatusCmd(suite.ctx)
	suite.mockClient.On("Set", suite.ctx, suite.cacheKey, mock.Anything,
		time.Duration(suite.testCache.TTLSeconds)*time.Second).Return(statusCmd)

	err := suite.store.CreateAttributeCache(suite.ctx, suite.testCache)
	suite.NoError(err)
}

func (suite *RedisAttributeCacheStoreTestSuite) TestCreateAttributeCache_SetError() {
	statusCmd := redis.NewStatusCmd(suite.ctx)
	statusCmd.SetErr(errors.New("connection refused"))
	suite.mockClient.On("Set", suite.ctx, suite.cacheKey, mock.Anything,
		time.Duration(suite.testCache.TTLSeconds)*time.Second).Return(statusCmd)

	err := suite.store.CreateAttributeCache(suite.ctx, suite.testCache)
	suite.Error(err)
	suite.Contains(err.Error(), "failed to store attribute cache in Redis")
}

// Tests for GetAttributeCache

func (suite *RedisAttributeCacheStoreTestSuite) TestGetAttributeCache_Success() {
	data, _ := json.Marshal(suite.testCache)
	stringCmd := redis.NewStringCmd(suite.ctx)
	stringCmd.SetVal(string(data))
	suite.mockClient.On("Get", suite.ctx, suite.cacheKey).Return(stringCmd)

	durationCmd := redis.NewDurationCmd(suite.ctx, time.Second)
	durationCmd.SetVal(30 * time.Minute)
	suite.mockClient.On("TTL", suite.ctx, suite.cacheKey).Return(durationCmd)

	result, err := suite.store.GetAttributeCache(suite.ctx, redisTestCacheID)
	suite.NoError(err)
	suite.Equal(redisTestCacheID, result.ID)
	suite.Equal(1800, result.TTLSeconds) // Overridden by Redis TTL (30 min = 1800 s)
}

func (suite *RedisAttributeCacheStoreTestSuite) TestGetAttributeCache_NotFound() {
	stringCmd := redis.NewStringCmd(suite.ctx)
	stringCmd.SetErr(redis.Nil)
	suite.mockClient.On("Get", suite.ctx, suite.cacheKey).Return(stringCmd)

	result, err := suite.store.GetAttributeCache(suite.ctx, redisTestCacheID)
	suite.Error(err)
	suite.Equal(errAttributeCacheNotFound, err)
	suite.Equal(AttributeCache{}, result)
}

func (suite *RedisAttributeCacheStoreTestSuite) TestGetAttributeCache_GetError() {
	stringCmd := redis.NewStringCmd(suite.ctx)
	stringCmd.SetErr(errors.New("connection refused"))
	suite.mockClient.On("Get", suite.ctx, suite.cacheKey).Return(stringCmd)

	result, err := suite.store.GetAttributeCache(suite.ctx, redisTestCacheID)
	suite.Error(err)
	suite.Contains(err.Error(), "failed to get attribute cache from Redis")
	suite.Equal(AttributeCache{}, result)
}

func (suite *RedisAttributeCacheStoreTestSuite) TestGetAttributeCache_UnmarshalError() {
	stringCmd := redis.NewStringCmd(suite.ctx)
	stringCmd.SetVal("not valid json{{{")
	suite.mockClient.On("Get", suite.ctx, suite.cacheKey).Return(stringCmd)

	result, err := suite.store.GetAttributeCache(suite.ctx, redisTestCacheID)
	suite.Error(err)
	suite.Contains(err.Error(), "failed to unmarshal attribute cache")
	suite.Equal(AttributeCache{}, result)
}

func (suite *RedisAttributeCacheStoreTestSuite) TestGetAttributeCache_TTLError_KeepsStoredTTL() {
	// When TTL call fails, the stored TTLSeconds from JSON is kept.
	data, _ := json.Marshal(suite.testCache)
	stringCmd := redis.NewStringCmd(suite.ctx)
	stringCmd.SetVal(string(data))
	suite.mockClient.On("Get", suite.ctx, suite.cacheKey).Return(stringCmd)

	durationCmd := redis.NewDurationCmd(suite.ctx, time.Second)
	durationCmd.SetErr(errors.New("redis error"))
	suite.mockClient.On("TTL", suite.ctx, suite.cacheKey).Return(durationCmd)

	result, err := suite.store.GetAttributeCache(suite.ctx, redisTestCacheID)
	suite.NoError(err)
	suite.Equal(suite.testCache.TTLSeconds, result.TTLSeconds) // Falls back to stored value
}

// Tests for ExtendAttributeCacheTTL

func (suite *RedisAttributeCacheStoreTestSuite) TestExtendAttributeCacheTTL_Success() {
	newTTL := 7200
	boolCmd := redis.NewBoolCmd(suite.ctx)
	boolCmd.SetVal(true)
	suite.mockClient.On("Expire", suite.ctx, suite.cacheKey,
		time.Duration(newTTL)*time.Second).Return(boolCmd)

	err := suite.store.ExtendAttributeCacheTTL(suite.ctx, redisTestCacheID, newTTL)
	suite.NoError(err)
}

func (suite *RedisAttributeCacheStoreTestSuite) TestExtendAttributeCacheTTL_NotFound() {
	newTTL := 7200
	boolCmd := redis.NewBoolCmd(suite.ctx)
	boolCmd.SetVal(false) // Key not found
	suite.mockClient.On("Expire", suite.ctx, suite.cacheKey,
		time.Duration(newTTL)*time.Second).Return(boolCmd)

	err := suite.store.ExtendAttributeCacheTTL(suite.ctx, redisTestCacheID, newTTL)
	suite.Error(err)
	suite.Equal(errAttributeCacheNotFound, err)
}

func (suite *RedisAttributeCacheStoreTestSuite) TestExtendAttributeCacheTTL_ExpireError() {
	newTTL := 7200
	boolCmd := redis.NewBoolCmd(suite.ctx)
	boolCmd.SetErr(errors.New("connection refused"))
	suite.mockClient.On("Expire", suite.ctx, suite.cacheKey,
		time.Duration(newTTL)*time.Second).Return(boolCmd)

	err := suite.store.ExtendAttributeCacheTTL(suite.ctx, redisTestCacheID, newTTL)
	suite.Error(err)
	suite.Contains(err.Error(), "failed to extend attribute cache TTL in Redis")
}

// Tests for DeleteAttributeCache

func (suite *RedisAttributeCacheStoreTestSuite) TestDeleteAttributeCache_Success() {
	intCmd := redis.NewIntCmd(suite.ctx)
	intCmd.SetVal(1)
	suite.mockClient.On("Del", suite.ctx, suite.cacheKey).Return(intCmd)

	err := suite.store.DeleteAttributeCache(suite.ctx, redisTestCacheID)
	suite.NoError(err)
}

func (suite *RedisAttributeCacheStoreTestSuite) TestDeleteAttributeCache_NotFound() {
	intCmd := redis.NewIntCmd(suite.ctx)
	intCmd.SetVal(0) // Key not found
	suite.mockClient.On("Del", suite.ctx, suite.cacheKey).Return(intCmd)

	err := suite.store.DeleteAttributeCache(suite.ctx, redisTestCacheID)
	suite.Error(err)
	suite.Equal(errAttributeCacheNotFound, err)
}

func (suite *RedisAttributeCacheStoreTestSuite) TestDeleteAttributeCache_DelError() {
	intCmd := redis.NewIntCmd(suite.ctx)
	intCmd.SetErr(errors.New("connection refused"))
	suite.mockClient.On("Del", suite.ctx, suite.cacheKey).Return(intCmd)

	err := suite.store.DeleteAttributeCache(suite.ctx, redisTestCacheID)
	suite.Error(err)
	suite.Contains(err.Error(), "failed to delete attribute cache from Redis")
}
