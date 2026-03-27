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

package passkey

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

// redisSessionEnvelope wraps the session data with associated user and relying party identifiers.
type redisSessionEnvelope struct {
	UserID         string       `json:"user_id"`
	RelyingPartyID string       `json:"relying_party_id"`
	Session        *sessionData `json:"session"`
}

// redisSessionStore is the Redis-backed implementation of sessionStoreInterface.
type redisSessionStore struct {
	client       *redis.Client
	keyPrefix    string
	deploymentID string
	logger       *log.Logger
}

// newRedisSessionStore creates a new Redis-backed passkey session store.
func newRedisSessionStore(p provider.RedisClientInterface) sessionStoreInterface {
	return &redisSessionStore{
		client:       p.GetRedisClient(),
		keyPrefix:    p.GetKeyPrefix(),
		deploymentID: config.GetThunderRuntime().Config.Server.Identifier,
		logger:       log.GetLogger().With(log.String(log.LoggerKeyComponentName, "WebAuthnRedisSessionStore")),
	}
}

// sessionKey builds the Redis key for a passkey session.
func (s *redisSessionStore) sessionKey(key string) string {
	return fmt.Sprintf("%s:runtime:%s:passkey:%s", s.keyPrefix, s.deploymentID, key)
}

// storeSession serializes the session envelope and stores
func (s *redisSessionStore) storeSession(
	sessionKey, userID, relyingPartyID string,
	session *sessionData,
	expiryTime time.Time,
) error {
	envelope := redisSessionEnvelope{
		UserID:         userID,
		RelyingPartyID: relyingPartyID,
		Session:        session,
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		s.logger.Error("Failed to marshal passkey session", log.Error(err))
		return err
	}

	ttl := time.Until(expiryTime)
	if ttl <= 0 {
		ttl = time.Second
	}

	if err := s.client.Set(context.Background(), s.sessionKey(sessionKey), data, ttl).Err(); err != nil {
		s.logger.Error("Failed to store passkey session in Redis", log.Error(err))
		return err
	}

	return nil
}

// retrieveSession loads and deserializes the passkey session envelope from Redis.
func (s *redisSessionStore) retrieveSession(sessionKey string) (*sessionData, string, string, error) {
	if sessionKey == "" {
		return nil, "", "", nil
	}

	data, err := s.client.Get(context.Background(), s.sessionKey(sessionKey)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, "", "", nil
		}
		s.logger.Error("Failed to retrieve passkey session from Redis", log.Error(err))
		return nil, "", "", err
	}

	var envelope redisSessionEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		s.logger.Error("Failed to unmarshal passkey session", log.Error(err))
		return nil, "", "", err
	}

	return envelope.Session, envelope.UserID, envelope.RelyingPartyID, nil
}

// deleteSession removes a passkey session from Redis.
func (s *redisSessionStore) deleteSession(sessionKey string) error {
	if sessionKey == "" {
		return nil
	}

	if err := s.client.Del(context.Background(), s.sessionKey(sessionKey)).Err(); err != nil {
		s.logger.Error("Failed to delete passkey session from Redis", log.Error(err))
		return err
	}

	return nil
}
