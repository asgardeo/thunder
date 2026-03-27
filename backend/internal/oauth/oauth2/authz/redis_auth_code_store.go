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
)

// consumeScript atomically checks that the stored code has the expected state,
// then updates it to the new state while preserving the remaining TTL.
var consumeScript = redis.NewScript(`
local val = redis.call('GET', KEYS[1])
if not val then return 0 end
local ok, data = pcall(cjson.decode, val)
if not ok then return 0 end
if data['State'] ~= ARGV[1] then return 0 end
data['State'] = ARGV[2]
local ttl = redis.call('TTL', KEYS[1])
if ttl < 0 then ttl = 0 end
redis.call('SET', KEYS[1], cjson.encode(data), 'EX', ttl)
return 1
`)

// redisAuthorizationCodeStore is the Redis-backed implementation of AuthorizationCodeStoreInterface.
type redisAuthorizationCodeStore struct {
	client       *redis.Client
	keyPrefix    string
	deploymentID string
}

// newRedisAuthorizationCodeStore creates a new Redis-backed authorization code store.
func newRedisAuthorizationCodeStore(p provider.RedisClientInterface) AuthorizationCodeStoreInterface {
	return &redisAuthorizationCodeStore{
		client:       p.GetRedisClient(),
		keyPrefix:    p.GetKeyPrefix(),
		deploymentID: config.GetThunderRuntime().Config.Server.Identifier,
	}
}

// codeKey builds the Redis key for an authorization code.
func (s *redisAuthorizationCodeStore) codeKey(clientID, code string) string {
	return fmt.Sprintf("%s:runtime:%s:authcode:%s:%s", s.keyPrefix, s.deploymentID, clientID, code)
}

// InsertAuthorizationCode stores a new authorization code in Redis with a TTL derived from ExpiryTime.
func (s *redisAuthorizationCodeStore) InsertAuthorizationCode(
	ctx context.Context, authzCode AuthorizationCode,
) error {
	data, err := json.Marshal(authzCode)
	if err != nil {
		return fmt.Errorf("failed to marshal authorization code: %w", err)
	}

	ttl := time.Until(authzCode.ExpiryTime)
	if ttl <= 0 {
		ttl = time.Second
	}

	if err := s.client.Set(ctx, s.codeKey(authzCode.ClientID, authzCode.Code), data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to insert authorization code into Redis: %w", err)
	}

	return nil
}

// ConsumeAuthorizationCode atomically transitions a code from ACTIVE to INACTIVE via a Lua script.
// Returns true if this call performed the transition; false if the code was already consumed or not found.
func (s *redisAuthorizationCodeStore) ConsumeAuthorizationCode(
	ctx context.Context, clientID, authCode string,
) (bool, error) {
	key := s.codeKey(clientID, authCode)
	result, err := consumeScript.Run(ctx, s.client, []string{key},
		AuthCodeStateActive, AuthCodeStateInactive).Int()
	if err != nil && !errors.Is(err, redis.Nil) {
		return false, fmt.Errorf("failed to consume authorization code: %w", err)
	}

	return result == 1, nil
}

// GetAuthorizationCode retrieves an authorization code by clientID and code value.
func (s *redisAuthorizationCodeStore) GetAuthorizationCode(
	ctx context.Context, clientID, authCode string,
) (*AuthorizationCode, error) {
	data, err := s.client.Get(ctx, s.codeKey(clientID, authCode)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errAuthorizationCodeNotFound
		}
		return nil, fmt.Errorf("failed to get authorization code from Redis: %w", err)
	}

	var code AuthorizationCode
	if err := json.Unmarshal(data, &code); err != nil {
		return nil, fmt.Errorf("failed to unmarshal authorization code: %w", err)
	}

	return &code, nil
}
