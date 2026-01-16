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
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
)

const (
	// sessionKeyLength is the length of the random session key in bytes.
	sessionKeyLength = 32
	// sessionTTLSeconds is the session time-to-live in seconds.
	sessionTTLSeconds = 120
	// cleanupIntervalMinutes is the interval for cleaning up expired sessions.
	cleanupIntervalMinutes = 5
)

// generateSessionKey generates a random base64-encoded session key.
func generateSessionKey() (string, error) {
	bytes := make([]byte, sessionKeyLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// startSessionCleanup starts a background routine to periodically clean up expired sessions.
func startSessionCleanup(store sessionStoreInterface) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "WebAuthnSessionCleanup"))
	logger.Debug("Starting session cleanup routine")

	go func() {
		ticker := time.NewTicker(cleanupIntervalMinutes * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			logger.Debug("Running expired session cleanup")
			if err := store.deleteExpiredSessions(); err != nil {
				logger.Error("Failed to cleanup expired sessions", log.Error(err))
			}
		}
	}()

	logger.Debug("Session cleanup routine started",
		log.Int("intervalMinutes", cleanupIntervalMinutes))
}

// storeSessionData stores session data in the database and returns a session key.
func (w *passkeyService) storeSessionData(
	userID, relyingPartyID string, sessionData *SessionData,
) (string, *serviceerror.ServiceError) {
	// Generate a random session key
	sessionKey, err := generateSessionKey()
	if err != nil {
		return "", &serviceerror.InternalServerError
	}

	expiresAt := time.Now().Add(time.Duration(sessionTTLSeconds) * time.Second)

	// Store session data in database
	err = w.sessionStore.storeSession(sessionKey, userID, relyingPartyID, sessionData, expiresAt)
	if err != nil {
		return "", &serviceerror.InternalServerError
	}

	return sessionKey, nil
}

// retrieveSessionData retrieves the session data from the database using the session key.
func (w *passkeyService) retrieveSessionData(
	sessionKey string,
) (*SessionData, string, string, *serviceerror.ServiceError) {
	// Retrieve session data from database
	sessionData, userID, relyingPartyID, err := w.sessionStore.retrieveSession(sessionKey)
	if err != nil {
		return nil, "", "", &ErrorSessionExpired
	}

	if sessionData == nil {
		return nil, "", "", &ErrorSessionExpired
	}

	return sessionData, userID, relyingPartyID, nil
}

// clearSessionData removes the session data from the database.
func (w *passkeyService) clearSessionData(sessionKey string) {
	// Remove session from database
	_ = w.sessionStore.deleteSession(sessionKey)
}
