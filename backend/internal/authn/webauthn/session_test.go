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

package webauthn

import (
	"errors"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// ==================== Test generateSessionKey ====================

func TestGenerateSessionKey(t *testing.T) {
	key1, err1 := generateSessionKey()
	key2, err2 := generateSessionKey()

	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.NotEmpty(t, key1)
	assert.NotEmpty(t, key2)
	assert.NotEqual(t, key1, key2, "Session keys should be unique")
	assert.Greater(t, len(key1), 32, "Session key should be base64 encoded and longer than 32 chars")
}

func TestGenerateSessionKey_Success(t *testing.T) {
	// Generate multiple keys to verify uniqueness
	keys := make(map[string]bool)
	for i := 0; i < 100; i++ {
		key, err := generateSessionKey()
		assert.NoError(t, err)
		assert.NotEmpty(t, key)

		// Verify uniqueness
		assert.False(t, keys[key], "Generated duplicate key")
		keys[key] = true

		// Verify base64 encoding (should be 44 chars for 32 bytes)
		assert.Equal(t, 44, len(key), "Base64 encoded 32 bytes should be 44 chars")
	}
}

// ==================== Test Session Service Methods ====================

// sessionStoreInterfaceMock is already defined in service_test.go
// We'll reuse it here for consistency

type SessionServiceTestSuite struct {
	suite.Suite
	mockSessionStore *sessionStoreInterfaceMock
	service          *webAuthnAuthnService
}

func TestSessionServiceTestSuite(t *testing.T) {
	suite.Run(t, new(SessionServiceTestSuite))
}

func (suite *SessionServiceTestSuite) SetupTest() {
	suite.mockSessionStore = &sessionStoreInterfaceMock{}
	suite.service = &webAuthnAuthnService{
		sessionStore: suite.mockSessionStore,
	}
}

// ==================== Tests for storeSessionData ====================

// Test L78-80: generateSessionKey error handling
// Note: This is difficult to test directly as crypto/rand.Read rarely fails
// The error handling exists for defensive programming

func (suite *SessionServiceTestSuite) TestStoreSessionData_Success() {
	// Tests the happy path of storeSessionData
	userID := "user123"
	relyingPartyID := "example.com"
	sessionData := &webauthn.SessionData{
		Challenge: "dGVzdC1jaGFsbGVuZ2U", // base64 encoded "test-challenge"
	}

	// Mock successful session storage
	suite.mockSessionStore.On("storeSession",
		mock.AnythingOfType("string"), // sessionKey (random)
		userID,
		relyingPartyID,
		sessionData,
		mock.AnythingOfType("time.Time"), // expiresAt
	).Return(nil).Once()

	sessionKey, err := suite.service.storeSessionData(userID, relyingPartyID, sessionData)

	suite.Nil(err)
	suite.NotEmpty(sessionKey)
	suite.Greater(len(sessionKey), 32)
	suite.mockSessionStore.AssertExpectations(suite.T())
}

func (suite *SessionServiceTestSuite) TestStoreSessionData_StoreSessionError() {
	// Tests L84-86: storeSession error handling
	userID := "user123"
	relyingPartyID := "example.com"
	sessionData := &webauthn.SessionData{
		Challenge: "dGVzdC1jaGFsbGVuZ2U",
	}

	// Mock session storage failure
	storeError := errors.New("database error")
	suite.mockSessionStore.On("storeSession",
		mock.AnythingOfType("string"),
		userID,
		relyingPartyID,
		sessionData,
		mock.AnythingOfType("time.Time"),
	).Return(storeError).Once()

	sessionKey, err := suite.service.storeSessionData(userID, relyingPartyID, sessionData)

	suite.NotNil(err)
	suite.Equal(&ErrorInternalServerError, err)
	suite.Empty(sessionKey)
	suite.mockSessionStore.AssertExpectations(suite.T())
}

func (suite *SessionServiceTestSuite) TestStoreSessionData_ExpiryCalculation() {
	// Verify that expiry time is calculated correctly
	userID := "user123"
	relyingPartyID := "example.com"
	sessionData := &webauthn.SessionData{
		Challenge: "dGVzdC1jaGFsbGVuZ2U",
	}

	startTime := time.Now()

	suite.mockSessionStore.On("storeSession",
		mock.AnythingOfType("string"),
		userID,
		relyingPartyID,
		sessionData,
		mock.MatchedBy(func(expiresAt time.Time) bool {
			// Verify expiry is approximately sessionTTLSeconds in the future
			expectedExpiry := startTime.Add(time.Duration(sessionTTLSeconds) * time.Second)
			diff := expiresAt.Sub(expectedExpiry).Abs()
			return diff < 1*time.Second // Allow 1 second tolerance
		}),
	).Return(nil).Once()

	_, err := suite.service.storeSessionData(userID, relyingPartyID, sessionData)

	suite.Nil(err)
	suite.mockSessionStore.AssertExpectations(suite.T())
}

// ==================== Tests for retrieveSessionData ====================

func (suite *SessionServiceTestSuite) TestRetrieveSessionData_Success() {
	// Tests the happy path of retrieveSessionData
	sessionKey := "test-session-key"
	expectedUserID := "user123"
	expectedRelyingPartyID := "example.com"
	expectedSessionData := &webauthn.SessionData{
		Challenge: "dGVzdC1jaGFsbGVuZ2U",
	}

	suite.mockSessionStore.On("retrieveSession", sessionKey).
		Return(expectedSessionData, expectedUserID, expectedRelyingPartyID, nil).Once()

	sessionData, userID, relyingPartyID, err := suite.service.retrieveSessionData(sessionKey)

	suite.Nil(err)
	suite.Equal(expectedSessionData, sessionData)
	suite.Equal(expectedUserID, userID)
	suite.Equal(expectedRelyingPartyID, relyingPartyID)
	suite.mockSessionStore.AssertExpectations(suite.T())
}

func (suite *SessionServiceTestSuite) TestRetrieveSessionData_RetrieveError() {
	// Tests L98-100: retrieveSession error handling
	sessionKey := "invalid-session-key"
	retrieveError := errors.New("session not found")

	suite.mockSessionStore.On("retrieveSession", sessionKey).
		Return(nil, "", "", retrieveError).Once()

	sessionData, userID, relyingPartyID, err := suite.service.retrieveSessionData(sessionKey)

	suite.NotNil(err)
	suite.Equal(&ErrorSessionExpired, err)
	suite.Nil(sessionData)
	suite.Empty(userID)
	suite.Empty(relyingPartyID)
	suite.mockSessionStore.AssertExpectations(suite.T())
}

func (suite *SessionServiceTestSuite) TestRetrieveSessionData_NilSessionData() {
	// Tests L101-103: sessionData == nil check
	sessionKey := "test-session-key"

	// Mock returns nil sessionData even though no error
	suite.mockSessionStore.On("retrieveSession", sessionKey).
		Return(nil, "user123", "example.com", nil).Once()

	sessionData, userID, relyingPartyID, err := suite.service.retrieveSessionData(sessionKey)

	suite.NotNil(err)
	suite.Equal(&ErrorSessionExpired, err)
	suite.Nil(sessionData)
	suite.Empty(userID)
	suite.Empty(relyingPartyID)
	suite.mockSessionStore.AssertExpectations(suite.T())
}

func (suite *SessionServiceTestSuite) TestRetrieveSessionData_ExpiredSession() {
	// Tests the scenario where session is expired (common case for L98)
	sessionKey := "expired-session-key"
	expiredError := errors.New("session expired")

	suite.mockSessionStore.On("retrieveSession", sessionKey).
		Return(nil, "", "", expiredError).Once()

	sessionData, userID, relyingPartyID, err := suite.service.retrieveSessionData(sessionKey)

	suite.NotNil(err)
	suite.Equal(&ErrorSessionExpired, err)
	suite.Nil(sessionData)
	suite.Empty(userID)
	suite.Empty(relyingPartyID)
	suite.mockSessionStore.AssertExpectations(suite.T())
}

// ==================== Tests for clearSessionData ====================

func (suite *SessionServiceTestSuite) TestClearSessionData_Success() {
	sessionKey := "test-session-key"

	suite.mockSessionStore.On("deleteSession", sessionKey).
		Return(nil).Once()

	// Should not panic or return error
	suite.service.clearSessionData(sessionKey)

	suite.mockSessionStore.AssertExpectations(suite.T())
}

func (suite *SessionServiceTestSuite) TestClearSessionData_DeleteError() {
	// Tests that errors from deleteSession are ignored (L111: _ = ...)
	sessionKey := "test-session-key"
	deleteError := errors.New("delete failed")

	suite.mockSessionStore.On("deleteSession", sessionKey).
		Return(deleteError).Once()

	// Should not panic even if delete fails
	suite.service.clearSessionData(sessionKey)

	suite.mockSessionStore.AssertExpectations(suite.T())
}

func (suite *SessionServiceTestSuite) TestClearSessionData_EmptyKey() {
	sessionKey := ""

	suite.mockSessionStore.On("deleteSession", sessionKey).
		Return(nil).Once()

	// Should handle empty key gracefully
	suite.service.clearSessionData(sessionKey)

	suite.mockSessionStore.AssertExpectations(suite.T())
}

// ==================== Integration-Style Tests ====================

func (suite *SessionServiceTestSuite) TestSessionRoundTrip() {
	// Tests complete flow: store -> retrieve -> clear
	userID := "user123"
	relyingPartyID := "example.com"
	sessionData := &webauthn.SessionData{
		Challenge:      "dGVzdC1jaGFsbGVuZ2U",
		RelyingPartyID: relyingPartyID,
		UserID:         []byte(userID),
	}

	var capturedSessionKey string

	// Mock store
	suite.mockSessionStore.On("storeSession",
		mock.AnythingOfType("string"),
		userID,
		relyingPartyID,
		sessionData,
		mock.AnythingOfType("time.Time"),
	).Run(func(args mock.Arguments) {
		capturedSessionKey = args.Get(0).(string)
	}).Return(nil).Once()

	// Mock retrieve with captured key
	suite.mockSessionStore.On("retrieveSession", mock.MatchedBy(func(key string) bool {
		return key == capturedSessionKey
	})).Return(sessionData, userID, relyingPartyID, nil).Once()

	// Mock delete
	suite.mockSessionStore.On("deleteSession", mock.MatchedBy(func(key string) bool {
		return key == capturedSessionKey
	})).Return(nil).Once()

	// Store
	sessionKey, err := suite.service.storeSessionData(userID, relyingPartyID, sessionData)
	suite.Nil(err)
	suite.NotEmpty(sessionKey)

	// Retrieve
	retrievedData, retrievedUserID, retrievedRPID, err := suite.service.retrieveSessionData(sessionKey)
	suite.Nil(err)
	suite.Equal(sessionData, retrievedData)
	suite.Equal(userID, retrievedUserID)
	suite.Equal(relyingPartyID, retrievedRPID)

	// Clear
	suite.service.clearSessionData(sessionKey)

	suite.mockSessionStore.AssertExpectations(suite.T())
}

// ==================== Constants Validation Tests ====================

func TestSessionConstants(t *testing.T) {
	// Verify session constants are reasonable
	assert.Equal(t, 32, sessionKeyLength, "Session key should be 32 bytes")
	assert.Equal(t, 120, sessionTTLSeconds, "Session TTL should be 120 seconds (2 minutes)")
	assert.Equal(t, 5, cleanupIntervalMinutes, "Cleanup interval should be 5 minutes")

	// Verify cleanup interval is longer than session TTL
	cleanupSeconds := cleanupIntervalMinutes * 60
	assert.Greater(t, cleanupSeconds, sessionTTLSeconds,
		"Cleanup interval should be longer than session TTL")
}
