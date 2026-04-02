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

// consumeAuthCodeScript atomically transitions an authorization code from ACTIVE to INACTIVE.
// Returns the updated JSON on success, or nil if not found / already consumed.
var consumeAuthCodeScript = redis.NewScript(`
local val = redis.call('GET', KEYS[1])
if not val then return nil end
local data = cjson.decode(val)
if data['State'] ~= ARGV[1] then return nil end
data['State'] = ARGV[2]
local newVal = cjson.encode(data)
redis.call('SET', KEYS[1], newVal, 'KEEPTTL')
return newVal
`)

// redisAuthorizationCodeStore is the Redis-backed implementation of AuthorizationCodeStoreInterface.
type redisAuthorizationCodeStore struct {
	client       *redis.Client
	keyPrefix    string
	deploymentID string
}

// newRedisAuthorizationCodeStore creates a new Redis-backed authorization code store.
func newRedisAuthorizationCodeStore(p provider.RedisProviderInterface) AuthorizationCodeStoreInterface {
	return &redisAuthorizationCodeStore{
		client:       p.GetRedisClient(),
		keyPrefix:    p.GetKeyPrefix(),
		deploymentID: config.GetThunderRuntime().Config.Server.Identifier,
	}
}

// authCodeKey builds the Redis key for an authorization code.
func (s *redisAuthorizationCodeStore) authCodeKey(clientID, code string) string {
	return fmt.Sprintf("%s:runtime:%s:authcode:%s:%s", s.keyPrefix, s.deploymentID, clientID, code)
}

// InsertAuthorizationCode serializes the authorization code and stores it in Redis with a TTL.
func (s *redisAuthorizationCodeStore) InsertAuthorizationCode(
	ctx context.Context, authzCode AuthorizationCode,
) error {
	data, err := json.Marshal(authzCode)
	if err != nil {
		return fmt.Errorf("failed to marshal authorization code: %w", err)
	}

	ttl := time.Until(authzCode.ExpiryTime)
	if err := s.client.Set(ctx, s.authCodeKey(authzCode.ClientID, authzCode.Code), data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to store authorization code in Redis: %w", err)
	}

	return nil
}

// ConsumeAuthorizationCode atomically transitions an ACTIVE code to INACTIVE and returns it.
// Returns errAuthorizationCodeNotFound if the code does not exist or is not ACTIVE.
func (s *redisAuthorizationCodeStore) ConsumeAuthorizationCode(
	ctx context.Context, clientID, authCode string,
) (*AuthorizationCode, error) {
	key := s.authCodeKey(clientID, authCode)

	result, err := consumeAuthCodeScript.Run(ctx, s.client, []string{key},
		AuthCodeStateActive, AuthCodeStateInactive).Text()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errAuthorizationCodeNotFound
		}
		return nil, fmt.Errorf("failed to consume authorization code: %w", err)
	}

	var consumed AuthorizationCode
	if err := json.Unmarshal([]byte(result), &consumed); err != nil {
		return nil, fmt.Errorf("failed to unmarshal authorization code: %w", err)
	}

	return &consumed, nil
}

// GetAuthorizationCode retrieves an authorization code by client ID and code.
func (s *redisAuthorizationCodeStore) GetAuthorizationCode(
	ctx context.Context, clientID, authCode string,
) (*AuthorizationCode, error) {
	data, err := s.client.Get(ctx, s.authCodeKey(clientID, authCode)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errAuthorizationCodeNotFound
		}
		return nil, fmt.Errorf("failed to get authorization code from Redis: %w", err)
	}

	var result AuthorizationCode
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal authorization code: %w", err)
	}

	return &result, nil
}
