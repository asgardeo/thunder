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
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
)

type SessionUtilsTestSuite struct {
	suite.Suite
}

func TestSessionUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(SessionUtilsTestSuite))
}

func (suite *SessionUtilsTestSuite) TestGenerateSessionKey() {
	key1, err1 := generateSessionKey()
	key2, err2 := generateSessionKey()

	suite.Nil(err1)
	suite.Nil(err2)
	suite.NotEmpty(key1)
	suite.NotEmpty(key2)
	suite.NotEqual(key1, key2, "Session keys should be unique")
	suite.Greater(len(key1), 32, "Session key should be base64 encoded and longer than 32 chars")
}

func (suite *SessionUtilsTestSuite) TestGenerateSessionKey_Success() {
	// Generate multiple keys to verify uniqueness
	keys := make(map[string]bool)
	for i := 0; i < 100; i++ {
		key, err := generateSessionKey()
		suite.NoError(err)
		suite.NotEmpty(key)

		// Verify uniqueness
		suite.False(keys[key], "Generated duplicate key")
		keys[key] = true

		// Verify base64 encoding (should be 44 chars for 32 bytes)
		suite.Equal(44, len(key), "Base64 encoded 32 bytes should be 44 chars")
	}
}

type SessionServiceTestSuite struct {
	suite.Suite
	mockSessionStore *sessionStoreInterfaceMock
	service          *passkeyService
}

func TestSessionServiceTestSuite(t *testing.T) {
	suite.Run(t, new(SessionServiceTestSuite))
}

func (suite *SessionServiceTestSuite) SetupTest() {
	suite.mockSessionStore = &sessionStoreInterfaceMock{}
	suite.service = &passkeyService{
		sessionStore: suite.mockSessionStore,
	}
}

func (suite *SessionServiceTestSuite) TestStoreSessionData_Success() {
	// Tests the happy path of storeSessionData
	sessionData := &SessionData{
		Challenge: "dGVzdC1jaGFsbGVuZ2U", // base64 encoded "test-challenge"
	}

	// Mock successful session storage
	suite.mockSessionStore.On("storeSession",
		mock.AnythingOfType("string"), // sessionKey (random)
		testUserID,
		testRelyingPartyID,
		sessionData,
		mock.AnythingOfType("time.Time"), // expiresAt
	).Return(nil).Once()

	sessionKey, err := suite.service.storeSessionData(testUserID, testRelyingPartyID, sessionData)

	suite.Nil(err)
	suite.NotEmpty(sessionKey)
	suite.Greater(len(sessionKey), 32)
	suite.mockSessionStore.AssertExpectations(suite.T())
}

func (suite *SessionServiceTestSuite) TestStoreSessionData_StoreSessionError() {
	sessionData := &SessionData{
		Challenge: "dGVzdC1jaGFsbGVuZ2U",
	}

	// Mock session storage failure
	storeError := errors.New("database error")
	suite.mockSessionStore.On("storeSession",
		mock.AnythingOfType("string"),
		testUserID,
		testRelyingPartyID,
		sessionData,
		mock.AnythingOfType("time.Time"),
	).Return(storeError).Once()

	sessionKey, err := suite.service.storeSessionData(testUserID, testRelyingPartyID, sessionData)

	suite.NotNil(err)
	suite.Equal(&serviceerror.InternalServerError, err)
	suite.Empty(sessionKey)
	suite.mockSessionStore.AssertExpectations(suite.T())
}

func (suite *SessionServiceTestSuite) TestStoreSessionData_ExpiryCalculation() {
	// Verify that expiry time is calculated correctly
	sessionData := &SessionData{
		Challenge: "dGVzdC1jaGFsbGVuZ2U",
	}

	startTime := time.Now()

	suite.mockSessionStore.On("storeSession",
		mock.AnythingOfType("string"),
		testUserID,
		testRelyingPartyID,
		sessionData,
		mock.MatchedBy(func(expiresAt time.Time) bool {
			// Verify expiry is approximately sessionTTLSeconds in the future
			expectedExpiry := startTime.Add(time.Duration(sessionTTLSeconds) * time.Second)
			diff := expiresAt.Sub(expectedExpiry).Abs()
			return diff < 1*time.Second // Allow 1 second tolerance
		}),
	).Return(nil).Once()

	_, err := suite.service.storeSessionData(testUserID, testRelyingPartyID, sessionData)

	suite.Nil(err)
	suite.mockSessionStore.AssertExpectations(suite.T())
}

func (suite *SessionServiceTestSuite) TestRetrieveSessionData_Success() {
	expectedSessionData := &SessionData{
		Challenge: "dGVzdC1jaGFsbGVuZ2U",
	}

	suite.mockSessionStore.On("retrieveSession", testSessionToken).
		Return(expectedSessionData, testUserID, testRelyingPartyID, nil).Once()

	sessionData, userID, relyingPartyID, err := suite.service.retrieveSessionData(testSessionToken)

	suite.Nil(err)
	suite.Equal(expectedSessionData, sessionData)
	suite.Equal(testUserID, userID)
	suite.Equal(testRelyingPartyID, relyingPartyID)
	suite.mockSessionStore.AssertExpectations(suite.T())
}

func (suite *SessionServiceTestSuite) TestRetrieveSessionData_RetrieveError() {
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
	// Mock returns nil sessionData even though no error
	suite.mockSessionStore.On("retrieveSession", testSessionToken).
		Return(nil, testUserID, testRelyingPartyID, nil).Once()

	sessionData, userID, relyingPartyID, err := suite.service.retrieveSessionData(testSessionToken)

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

func (suite *SessionServiceTestSuite) TestClearSessionData_Success() {
	suite.mockSessionStore.On("deleteSession", testSessionToken).
		Return(nil).Once()

	// Should not panic or return error
	suite.service.clearSessionData(testSessionToken)

	suite.mockSessionStore.AssertExpectations(suite.T())
}

func (suite *SessionServiceTestSuite) TestClearSessionData_DeleteError() {
	// Tests that errors from deleteSession are ignored (L111: _ = ...)
	deleteError := errors.New("delete failed")

	suite.mockSessionStore.On("deleteSession", testSessionToken).
		Return(deleteError).Once()

	// Should not panic even if delete fails
	suite.service.clearSessionData(testSessionToken)

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

func (suite *SessionServiceTestSuite) TestSessionRoundTrip() {
	sessionData := &SessionData{
		Challenge:      "dGVzdC1jaGFsbGVuZ2U",
		RelyingPartyID: testRelyingPartyID,
		UserID:         []byte(testUserID),
	}

	var capturedSessionKey string

	// Mock store
	suite.mockSessionStore.On("storeSession",
		mock.AnythingOfType("string"),
		testUserID,
		testRelyingPartyID,
		sessionData,
		mock.AnythingOfType("time.Time"),
	).Run(func(args mock.Arguments) {
		capturedSessionKey = args.Get(0).(string)
	}).Return(nil).Once()

	// Mock retrieve with captured key
	suite.mockSessionStore.On("retrieveSession", mock.MatchedBy(func(key string) bool {
		return key == capturedSessionKey
	})).Return(sessionData, testUserID, testRelyingPartyID, nil).Once()

	// Mock delete
	suite.mockSessionStore.On("deleteSession", mock.MatchedBy(func(key string) bool {
		return key == capturedSessionKey
	})).Return(nil).Once()

	// Store
	sessionKey, err := suite.service.storeSessionData(testUserID, testRelyingPartyID, sessionData)
	suite.Nil(err)
	suite.NotEmpty(sessionKey)

	// Retrieve
	retrievedData, retrievedUserID, retrievedRPID, err := suite.service.retrieveSessionData(sessionKey)
	suite.Nil(err)
	suite.Equal(sessionData, retrievedData)
	suite.Equal(testUserID, retrievedUserID)
	suite.Equal(testRelyingPartyID, retrievedRPID)

	// Clear
	suite.service.clearSessionData(sessionKey)

	suite.mockSessionStore.AssertExpectations(suite.T())
}

func (suite *SessionUtilsTestSuite) TestSessionConstants() {
	// Verify session constants are reasonable
	suite.Equal(32, sessionKeyLength, "Session key should be 32 bytes")
	suite.Equal(120, sessionTTLSeconds, "Session TTL should be 120 seconds (2 minutes)")
	suite.Equal(5, cleanupIntervalMinutes, "Cleanup interval should be 5 minutes")

	// Verify cleanup interval is longer than session TTL
	cleanupSeconds := cleanupIntervalMinutes * 60
	suite.Greater(cleanupSeconds, sessionTTLSeconds,
		"Cleanup interval should be longer than session TTL")
}
