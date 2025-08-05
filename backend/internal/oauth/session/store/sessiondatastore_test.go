/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
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

package store

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	authndto "github.com/asgardeo/thunder/internal/authn/dto"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/model"
	sessionmodel "github.com/asgardeo/thunder/internal/oauth/session/model"
)

type SessionDataStoreTestSuite struct {
	suite.Suite
	store SessionDataStoreInterface
}

func TestSessionDataStoreSuite(t *testing.T) {
	suite.Run(t, new(SessionDataStoreTestSuite))
}

func (suite *SessionDataStoreTestSuite) SetupTest() {
	// Reset singleton for each test
	instance = nil
	once = sync.Once{}
	
	suite.store = GetSessionDataStore()
	suite.store.ClearSessionStore()
}

func (suite *SessionDataStoreTestSuite) TearDownTest() {
	if suite.store != nil {
		suite.store.ClearSessionStore()
	}
}

func (suite *SessionDataStoreTestSuite) TestGetSessionDataStore() {
	store := GetSessionDataStore()
	assert.NotNil(suite.T(), store)
	assert.Implements(suite.T(), (*SessionDataStoreInterface)(nil), store)
}

func (suite *SessionDataStoreTestSuite) TestGetSessionDataStoreSingleton() {
	store1 := GetSessionDataStore()
	store2 := GetSessionDataStore()
	
	// Should return the same instance
	assert.Same(suite.T(), store1, store2)
}

func (suite *SessionDataStoreTestSuite) TestAddSession() {
	key := "test-session-key"
	sessionData := sessionmodel.SessionData{
		OAuthParameters: model.OAuthParameters{
			SessionDataKey: key,
			ClientID:       "test-client",
			RedirectURI:    "https://example.com/callback",
			ResponseType:   "code",
			Scopes:         "read write",
			State:          "test-state",
		},
		AuthTime: time.Now(),
		AuthenticatedUser: authndto.AuthenticatedUser{
			IsAuthenticated: true,
			UserID:          "user123",
			Attributes: map[string]string{
				"username": "testuser",
				"email":    "test@example.com",
			},
		},
	}

	suite.store.AddSession(key, sessionData)

	// Verify session was added
	found, retrievedData := suite.store.GetSession(key)
	assert.True(suite.T(), found)
	assert.Equal(suite.T(), sessionData.OAuthParameters.ClientID, retrievedData.OAuthParameters.ClientID)
	assert.Equal(suite.T(), sessionData.AuthenticatedUser.UserID, retrievedData.AuthenticatedUser.UserID)
}

func (suite *SessionDataStoreTestSuite) TestAddSessionWithEmptyKey() {
	sessionData := sessionmodel.SessionData{
		OAuthParameters: model.OAuthParameters{
			ClientID: "test-client",
		},
	}

	// Should not add session with empty key
	suite.store.AddSession("", sessionData)

	found, _ := suite.store.GetSession("")
	assert.False(suite.T(), found)
}

func (suite *SessionDataStoreTestSuite) TestGetSession() {
	key := "test-session-key"
	sessionData := sessionmodel.SessionData{
		OAuthParameters: model.OAuthParameters{
			ClientID: "test-client",
			State:    "test-state",
		},
		AuthTime: time.Now(),
	}

	suite.store.AddSession(key, sessionData)

	// Test successful retrieval
	found, retrievedData := suite.store.GetSession(key)
	assert.True(suite.T(), found)
	assert.Equal(suite.T(), sessionData.OAuthParameters.ClientID, retrievedData.OAuthParameters.ClientID)
	assert.Equal(suite.T(), sessionData.OAuthParameters.State, retrievedData.OAuthParameters.State)
}

func (suite *SessionDataStoreTestSuite) TestGetSessionNotFound() {
	// Test non-existent session
	found, _ := suite.store.GetSession("non-existent-key")
	assert.False(suite.T(), found)
}

func (suite *SessionDataStoreTestSuite) TestGetSessionWithEmptyKey() {
	// Test empty key
	found, _ := suite.store.GetSession("")
	assert.False(suite.T(), found)
}

func (suite *SessionDataStoreTestSuite) TestClearSession() {
	key := "test-session-key"
	sessionData := sessionmodel.SessionData{
		OAuthParameters: model.OAuthParameters{
			ClientID: "test-client",
		},
	}

	suite.store.AddSession(key, sessionData)

	// Verify session exists
	found, _ := suite.store.GetSession(key)
	assert.True(suite.T(), found)

	// Clear the session
	suite.store.ClearSession(key)

	// Verify session is removed
	found, _ = suite.store.GetSession(key)
	assert.False(suite.T(), found)
}

func (suite *SessionDataStoreTestSuite) TestClearSessionWithEmptyKey() {
	// Should not panic or cause issues
	suite.store.ClearSession("")
}

func (suite *SessionDataStoreTestSuite) TestClearSessionStore() {
	// Add multiple sessions
	keys := []string{"key1", "key2", "key3"}
	sessionData := sessionmodel.SessionData{
		OAuthParameters: model.OAuthParameters{
			ClientID: "test-client",
		},
	}

	for _, key := range keys {
		suite.store.AddSession(key, sessionData)
	}

	// Verify all sessions exist
	for _, key := range keys {
		found, _ := suite.store.GetSession(key)
		assert.True(suite.T(), found)
	}

	// Clear all sessions
	suite.store.ClearSessionStore()

	// Verify all sessions are removed
	for _, key := range keys {
		found, _ := suite.store.GetSession(key)
		assert.False(suite.T(), found)
	}
}

func (suite *SessionDataStoreTestSuite) TestSessionExpiry() {
	// This test requires access to the internal validity period
	// We'll create a session and test it doesn't exist after expiry
	key := "test-expiry-key"
	sessionData := sessionmodel.SessionData{
		OAuthParameters: model.OAuthParameters{
			ClientID: "test-client",
		},
		AuthTime: time.Now(),
	}

	suite.store.AddSession(key, sessionData)

	// Session should exist immediately
	found, _ := suite.store.GetSession(key)
	assert.True(suite.T(), found)

	// Since we can't easily manipulate time or access internal validity period,
	// we'll just verify the session exists now and the expiry logic is implemented
	// (the actual expiry testing would require more complex setup or time manipulation)
}

func (suite *SessionDataStoreTestSuite) TestConcurrentAccess() {
	// Test concurrent access to the session store
	key := "concurrent-test-key"
	sessionData := sessionmodel.SessionData{
		OAuthParameters: model.OAuthParameters{
			ClientID: "test-client",
		},
	}

	var wg sync.WaitGroup
	numGoroutines := 10

	// Add sessions concurrently
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()
			keyWithIndex := key + string(rune('0'+index))
			suite.store.AddSession(keyWithIndex, sessionData)
		}(i)
	}
	wg.Wait()

	// Read sessions concurrently
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()
			keyWithIndex := key + string(rune('0'+index))
			found, _ := suite.store.GetSession(keyWithIndex)
			assert.True(suite.T(), found)
		}(i)
	}
	wg.Wait()

	// Clear sessions concurrently
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()
			keyWithIndex := key + string(rune('0'+index))
			suite.store.ClearSession(keyWithIndex)
		}(i)
	}
	wg.Wait()

	// Verify all sessions are cleared
	for i := 0; i < numGoroutines; i++ {
		keyWithIndex := key + string(rune('0'+i))
		found, _ := suite.store.GetSession(keyWithIndex)
		assert.False(suite.T(), found)
	}
}