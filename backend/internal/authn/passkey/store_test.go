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
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/protocol/webauthncose"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/tests/mocks/database/providermock"
)

type StoreUtilsTestSuite struct {
	suite.Suite
}

func TestStoreUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(StoreUtilsTestSuite))
}

func (suite *StoreUtilsTestSuite) TestGetMapKeys() {
	testMap := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	keys := getMapKeys(testMap)

	suite.NotNil(keys)
	suite.Len(keys, 3)
	suite.Contains(keys, "key1")
	suite.Contains(keys, "key2")
	suite.Contains(keys, "key3")
}

func (suite *StoreUtilsTestSuite) TestGetMapKeys_EmptyMap() {
	testMap := map[string]interface{}{}

	keys := getMapKeys(testMap)

	suite.NotNil(keys)
	suite.Len(keys, 0)
}

func (suite *StoreUtilsTestSuite) TestGetMapKeys_NilMap() {
	var testMap map[string]interface{}

	keys := getMapKeys(testMap)

	suite.NotNil(keys)
	suite.Len(keys, 0)
}

type SessionStoreTestSuite struct {
	suite.Suite
	mockDBProvider *providermock.DBProviderInterfaceMock
	mockDBClient   *providermock.DBClientInterfaceMock
	store          *sessionStore
}

func TestSessionStoreTestSuite(t *testing.T) {
	suite.Run(t, new(SessionStoreTestSuite))
}

func (suite *SessionStoreTestSuite) SetupSuite() {
	testConfig := &config.Config{
		Server: config.ServerConfig{
			Identifier: "test-deployment-id",
		},
	}
	err := config.InitializeThunderRuntime("", testConfig)
	if err != nil {
		suite.T().Fatalf("Failed to initialize ThunderRuntime: %v", err)
	}
}

func (suite *SessionStoreTestSuite) SetupTest() {
	suite.mockDBProvider = providermock.NewDBProviderInterfaceMock(suite.T())
	suite.mockDBClient = providermock.NewDBClientInterfaceMock(suite.T())

	suite.store = &sessionStore{
		dbProvider:   suite.mockDBProvider,
		deploymentID: "test-deployment-id",
		logger:       log.GetLogger().With(log.String(log.LoggerKeyComponentName, "WebAuthnSessionStoreTest")),
	}
}

func (suite *SessionStoreTestSuite) TestNewSessionStore() {
	// Initialize ThunderRuntime for the test
	testConfig := &config.Config{
		Server: config.ServerConfig{
			Identifier: "test-deployment-id",
		},
	}
	err := config.InitializeThunderRuntime("", testConfig)
	if err != nil {
		suite.T().Fatalf("Failed to initialize ThunderRuntime: %v", err)
	}

	store := newSessionStore()

	suite.NotNil(store)
	suite.IsType(&sessionStore{}, store)
}

func (suite *SessionStoreTestSuite) TestStoreSession_Success() {
	expiryTime := time.Now().Add(5 * time.Minute)

	sessionData := &SessionData{
		Challenge:        "challenge123",
		UserID:           []byte(testUserID),
		RelyingPartyID:   testRelyingPartyID,
		UserVerification: "preferred",
	}

	suite.mockDBProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil).Once()
	suite.mockDBClient.On("Execute", mock.AnythingOfType("model.DBQuery"),
		testSessionKey, testUserID, testRelyingPartyID, mock.Anything, expiryTime, "test-deployment-id").
		Return(int64(1), nil).Once()

	err := suite.store.storeSession(testSessionKey, testUserID, testRelyingPartyID, sessionData, expiryTime)

	suite.NoError(err)
	suite.mockDBProvider.AssertExpectations(suite.T())
	suite.mockDBClient.AssertExpectations(suite.T())
}

func (suite *SessionStoreTestSuite) TestStoreSession_DBClientError() {
	expiryTime := time.Now().Add(5 * time.Minute)

	sessionData := &SessionData{
		Challenge: "challenge123",
	}

	suite.mockDBProvider.On("GetRuntimeDBClient").Return(nil, assert.AnError).Once()

	err := suite.store.storeSession(testSessionKey, testUserID, testRelyingPartyID, sessionData, expiryTime)

	suite.Error(err)
	suite.mockDBProvider.AssertExpectations(suite.T())
}

func (suite *SessionStoreTestSuite) TestStoreSession_ExecuteError() {
	expiryTime := time.Now().Add(5 * time.Minute)

	sessionData := &SessionData{
		Challenge: "challenge123",
	}

	suite.mockDBProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil).Once()
	suite.mockDBClient.On("Execute", mock.AnythingOfType("model.DBQuery"),
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(int64(0), assert.AnError).Once()

	err := suite.store.storeSession(testSessionKey, testUserID, testRelyingPartyID, sessionData, expiryTime)

	suite.Error(err)
}

func (suite *SessionStoreTestSuite) TestRetrieveSession_Success() {
	sessionDataJSON := `{
		"challenge": "challenge123",
		"rpId": "example.com",
		"userId": "` + base64.StdEncoding.EncodeToString([]byte(testUserID)) + `",
		"userVerification": "preferred"
	}`

	results := []map[string]interface{}{
		{
			dbColumnUserID:         testUserID,
			dbColumnRelyingPartyID: testRelyingPartyID,
			dbColumnSessionData:    sessionDataJSON,
		},
	}

	suite.mockDBProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil).Once()
	suite.mockDBClient.On("Query", mock.AnythingOfType("model.DBQuery"),
		testSessionKey, mock.AnythingOfType("time.Time"), "test-deployment-id").
		Return(results, nil).Once()

	sessionData, retrievedUserID, retrievedRPID, err := suite.store.retrieveSession(testSessionKey)

	suite.NoError(err)
	suite.NotNil(sessionData)
	suite.Equal(testUserID, retrievedUserID)
	suite.Equal(testRelyingPartyID, retrievedRPID)
	suite.Equal("challenge123", sessionData.Challenge)
}

func (suite *SessionStoreTestSuite) TestRetrieveSession_EmptyKey() {
	sessionData, userID, rpID, err := suite.store.retrieveSession("")

	suite.NoError(err)
	suite.Nil(sessionData)
	suite.Empty(userID)
	suite.Empty(rpID)
}

func (suite *SessionStoreTestSuite) TestRetrieveSession_NotFound() {
	suite.mockDBProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil).Once()
	suite.mockDBClient.On("Query", mock.AnythingOfType("model.DBQuery"),
		testSessionKey, mock.AnythingOfType("time.Time"), "test-deployment-id").
		Return([]map[string]interface{}{}, nil).Once()

	sessionData, userID, rpID, err := suite.store.retrieveSession(testSessionKey)

	suite.NoError(err)
	suite.Nil(sessionData)
	suite.Empty(userID)
	suite.Empty(rpID)
}

func (suite *SessionStoreTestSuite) TestRetrieveSession_DBClientError() {
	suite.mockDBProvider.On("GetRuntimeDBClient").Return(nil, assert.AnError).Once()

	sessionData, userID, rpID, err := suite.store.retrieveSession(testSessionKey)

	suite.Error(err)
	suite.Nil(sessionData)
	suite.Empty(userID)
	suite.Empty(rpID)
}

func (suite *SessionStoreTestSuite) TestRetrieveSession_QueryError() {
	suite.mockDBProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil).Once()
	suite.mockDBClient.On("Query", mock.AnythingOfType("model.DBQuery"),
		testSessionKey, mock.AnythingOfType("time.Time"), "test-deployment-id").
		Return(nil, assert.AnError).Once()

	sessionData, _, _, err := suite.store.retrieveSession(testSessionKey)

	suite.Error(err)
	suite.Nil(sessionData)
}

func (suite *SessionStoreTestSuite) TestRetrieveSession_SessionDataAsBytes() {
	sessionDataJSON := []byte(`{
		"challenge": "challenge123",
		"rpId": "example.com",
		"userVerification": "preferred"
	}`)

	results := []map[string]interface{}{
		{
			dbColumnUserID:         testUserID,
			dbColumnRelyingPartyID: testRelyingPartyID,
			dbColumnSessionData:    sessionDataJSON, // As bytes
		},
	}

	suite.mockDBProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil).Once()
	suite.mockDBClient.On("Query", mock.AnythingOfType("model.DBQuery"),
		testSessionKey, mock.AnythingOfType("time.Time"), "test-deployment-id").
		Return(results, nil).Once()

	sessionData, retrievedUserID, retrievedRPID, err := suite.store.retrieveSession(testSessionKey)

	suite.NoError(err)
	suite.NotNil(sessionData)
	suite.Equal(testUserID, retrievedUserID)
	suite.Equal(testRelyingPartyID, retrievedRPID)
}

func (suite *SessionStoreTestSuite) TestDeleteSession_Success() {
	suite.mockDBProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil).Once()
	suite.mockDBClient.On("Execute", mock.AnythingOfType("model.DBQuery"),
		testSessionKey, "test-deployment-id").
		Return(int64(1), nil).Once()

	err := suite.store.deleteSession(testSessionKey)

	suite.NoError(err)
}

func (suite *SessionStoreTestSuite) TestDeleteSession_EmptyKey() {
	err := suite.store.deleteSession("")

	suite.NoError(err)
}

func (suite *SessionStoreTestSuite) TestDeleteSession_DBClientError() {
	suite.mockDBProvider.On("GetRuntimeDBClient").Return(nil, assert.AnError).Once()

	err := suite.store.deleteSession(testSessionKey)

	suite.Error(err)
}

func (suite *SessionStoreTestSuite) TestDeleteSession_ExecuteError() {
	suite.mockDBProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil).Once()
	suite.mockDBClient.On("Execute", mock.AnythingOfType("model.DBQuery"),
		testSessionKey, "test-deployment-id").
		Return(int64(0), assert.AnError).Once()

	err := suite.store.deleteSession(testSessionKey)

	suite.Error(err)
}

func (suite *SessionStoreTestSuite) TestDeleteExpiredSessions_Success() {
	suite.mockDBProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil).Once()
	suite.mockDBClient.On("Execute", mock.AnythingOfType("model.DBQuery"),
		mock.AnythingOfType("time.Time"), "test-deployment-id").
		Return(int64(5), nil).Once()

	err := suite.store.deleteExpiredSessions()

	suite.NoError(err)
}

func (suite *SessionStoreTestSuite) TestDeleteExpiredSessions_DBClientError() {
	suite.mockDBProvider.On("GetRuntimeDBClient").Return(nil, assert.AnError).Once()

	err := suite.store.deleteExpiredSessions()

	suite.Error(err)
}

func (suite *SessionStoreTestSuite) TestDeleteExpiredSessions_ExecuteError() {
	suite.mockDBProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil).Once()
	suite.mockDBClient.On("Execute", mock.AnythingOfType("model.DBQuery"),
		mock.AnythingOfType("time.Time"), "test-deployment-id").
		Return(int64(0), assert.AnError).Once()

	err := suite.store.deleteExpiredSessions()

	suite.Error(err)
}

func (suite *SessionStoreTestSuite) TestSerializeSessionData_MinimalData() {
	sessionData := &SessionData{
		Challenge:        "challenge123",
		UserVerification: "preferred",
	}

	jsonBytes, err := suite.store.serializeSessionData(sessionData)

	suite.NoError(err)
	suite.NotNil(jsonBytes)

	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	suite.NoError(err)
	suite.Equal("challenge123", result[jsonKeyChallenge])
	suite.Equal("preferred", result[jsonKeyUserVerification])
}

func (suite *SessionStoreTestSuite) TestSerializeSessionData_FullData() {
	expiryTime := time.Now().Add(5 * time.Minute)
	sessionData := &SessionData{
		Challenge:            "challenge123",
		UserID:               []byte("user123"),
		RelyingPartyID:       "example.com",
		UserVerification:     "required",
		Expires:              expiryTime,
		AllowedCredentialIDs: [][]byte{[]byte("cred1"), []byte("cred2")},
		Extensions:           map[string]interface{}{"ext1": "value1"},
		CredParams: []protocol.CredentialParameter{
			{Type: "public-key", Algorithm: webauthncose.AlgES256},
		},
		Mediation: "conditional",
	}

	jsonBytes, err := suite.store.serializeSessionData(sessionData)

	suite.NoError(err)
	suite.NotNil(jsonBytes)

	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	suite.NoError(err)
	suite.Equal("challenge123", result[jsonKeyChallenge])
	suite.Equal("example.com", result[jsonKeyRelyingPartyID])
	suite.NotNil(result[jsonKeyUserID])
	suite.NotNil(result[jsonKeyAllowedCredentials])
	suite.NotNil(result[jsonKeyExtensions])
	suite.NotNil(result[jsonKeyCredParams])
	suite.Equal("conditional", result[jsonKeyMediation])
}

func (suite *SessionStoreTestSuite) TestSerializeSessionData_WithEmptyFields() {
	sessionData := &SessionData{
		Challenge:        "challenge123",
		UserVerification: "preferred",
		RelyingPartyID:   "",       // Empty
		UserID:           []byte{}, // Empty
	}

	jsonBytes, err := suite.store.serializeSessionData(sessionData)

	suite.NoError(err)
	suite.NotNil(jsonBytes)

	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	suite.NoError(err)
	// Empty RelyingPartyID should not be in JSON
	_, hasRP := result[jsonKeyRelyingPartyID]
	suite.False(hasRP)
}

func (suite *SessionStoreTestSuite) TestBuildSessionDataFromResultRow_Success() {
	userID := "user123"
	relyingPartyID := "example.com"

	sessionDataJSON := map[string]interface{}{
		jsonKeyChallenge:        "challenge123",
		jsonKeyRelyingPartyID:   relyingPartyID,
		jsonKeyUserID:           base64.StdEncoding.EncodeToString([]byte(userID)),
		jsonKeyUserVerification: "preferred",
		jsonKeyExpires:          float64(time.Now().Add(5 * time.Minute).Unix()),
	}

	jsonBytes, _ := json.Marshal(sessionDataJSON)

	row := map[string]interface{}{
		dbColumnUserID:         userID,
		dbColumnRelyingPartyID: relyingPartyID,
		dbColumnSessionData:    string(jsonBytes),
	}

	sessionData, retrievedUserID, retrievedRPID, err := suite.store.buildSessionDataFromResultRow(row)

	suite.NoError(err)
	suite.NotNil(sessionData)
	suite.Equal(userID, retrievedUserID)
	suite.Equal(relyingPartyID, retrievedRPID)
	suite.Equal("challenge123", sessionData.Challenge)
	suite.Equal(relyingPartyID, sessionData.RelyingPartyID)
}

func (suite *SessionStoreTestSuite) TestBuildSessionDataFromResultRow_WithAllFields() {
	sessionDataJSON := map[string]interface{}{
		jsonKeyChallenge:        "challenge123",
		jsonKeyRelyingPartyID:   testRelyingPartyID,
		jsonKeyUserID:           base64.StdEncoding.EncodeToString([]byte(testUserID)),
		jsonKeyUserVerification: "required",
		jsonKeyExpires:          float64(time.Now().Add(5 * time.Minute).Unix()),
		jsonKeyAllowedCredentials: []interface{}{
			base64.StdEncoding.EncodeToString([]byte("cred1")),
			base64.StdEncoding.EncodeToString([]byte("cred2")),
		},
		jsonKeyExtensions: map[string]interface{}{"ext1": "value1"},
		jsonKeyCredParams: []interface{}{
			map[string]interface{}{
				"type": "public-key",
				"alg":  float64(-7),
			},
		},
		jsonKeyMediation: "conditional",
	}

	jsonBytes, _ := json.Marshal(sessionDataJSON)

	row := map[string]interface{}{
		dbColumnUserID:         testUserID,
		dbColumnRelyingPartyID: testRelyingPartyID,
		dbColumnSessionData:    string(jsonBytes),
	}

	sessionData, retrievedUserID, _, err := suite.store.buildSessionDataFromResultRow(row)

	suite.NoError(err)
	suite.NotNil(sessionData)
	suite.Equal(testUserID, retrievedUserID)
	suite.Equal("challenge123", sessionData.Challenge)
	suite.Len(sessionData.AllowedCredentialIDs, 2)
	suite.NotNil(sessionData.Extensions)
	suite.Len(sessionData.CredParams, 1)
	suite.Equal(protocol.CredentialMediationRequirement("conditional"), sessionData.Mediation)
}

func (suite *SessionStoreTestSuite) TestBuildSessionDataFromResultRow_MissingSessionData() {
	row := map[string]interface{}{
		dbColumnUserID:         "user123",
		dbColumnRelyingPartyID: "example.com",
		// Missing dbColumnSessionData
	}

	sessionData, userID, rpID, err := suite.store.buildSessionDataFromResultRow(row)

	suite.Error(err)
	suite.Nil(sessionData)
	suite.Empty(userID)
	suite.Empty(rpID)
	suite.Contains(err.Error(), "SESSION_DATA is missing or invalid")
}

func (suite *SessionStoreTestSuite) TestBuildSessionDataFromResultRow_InvalidJSON() {
	row := map[string]interface{}{
		dbColumnUserID:         "user123",
		dbColumnRelyingPartyID: "example.com",
		dbColumnSessionData:    "invalid json",
	}

	sessionData, _, _, err := suite.store.buildSessionDataFromResultRow(row)

	suite.Error(err)
	suite.Nil(sessionData)
}

func (suite *SessionStoreTestSuite) TestRetrieveSession_BuildSessionDataError_InvalidJSON() {
	// Create invalid JSON that will fail to unmarshal
	invalidJSON := "not-valid-json{{"

	row := map[string]interface{}{
		dbColumnUserID:         "user123",
		dbColumnRelyingPartyID: "example.com",
		dbColumnSessionData:    invalidJSON,
	}

	suite.mockDBProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil).Once()
	suite.mockDBClient.On("Query", mock.AnythingOfType("model.DBQuery"),
		testSessionKey, mock.AnythingOfType("time.Time"), "test-deployment-id").
		Return([]map[string]interface{}{row}, nil).Once()

	sessionData, userID, rpID, err := suite.store.retrieveSession(testSessionKey)

	suite.Error(err)
	suite.Nil(sessionData)
	suite.Equal("", userID)
	suite.Equal("", rpID)
	suite.Contains(err.Error(), "invalid character")
}

func (suite *SessionStoreTestSuite) TestRetrieveSession_BuildSessionDataError_MissingSessionData() {
	row := map[string]interface{}{
		dbColumnUserID:         "user123",
		dbColumnRelyingPartyID: "example.com",
	}

	suite.mockDBProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil).Once()
	suite.mockDBClient.On("Query", mock.AnythingOfType("model.DBQuery"),
		testSessionKey, mock.AnythingOfType("time.Time"), "test-deployment-id").
		Return([]map[string]interface{}{row}, nil).Once()

	sessionData, userID, rpID, err := suite.store.retrieveSession(testSessionKey)

	suite.Error(err)
	suite.Nil(sessionData)
	suite.Equal("", userID)
	suite.Equal("", rpID)
	suite.Contains(err.Error(), "SESSION_DATA is missing or invalid")
}

func (suite *SessionStoreTestSuite) TestRetrieveSession_BuildSessionDataError_InvalidUserIDBase64() {
	sessionDataJSON := map[string]interface{}{
		jsonKeyChallenge:      "challenge123",
		jsonKeyUserID:         "invalid-base64!!!",
		jsonKeyRelyingPartyID: "example.com",
	}
	jsonBytes, _ := json.Marshal(sessionDataJSON)

	row := map[string]interface{}{
		dbColumnUserID:         "user123",
		dbColumnRelyingPartyID: "example.com",
		dbColumnSessionData:    string(jsonBytes),
	}

	suite.mockDBProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil).Once()
	suite.mockDBClient.On("Query", mock.AnythingOfType("model.DBQuery"),
		testSessionKey, mock.AnythingOfType("time.Time"), "test-deployment-id").
		Return([]map[string]interface{}{row}, nil).Once()

	sessionData, userID, rpID, err := suite.store.retrieveSession(testSessionKey)

	suite.Error(err)
	suite.Nil(sessionData)
	suite.Equal("", userID)
	suite.Equal("", rpID)
	suite.Contains(err.Error(), "illegal base64 data")
}

func (suite *SessionStoreTestSuite) TestRetrieveSession_BuildSessionDataError_InvalidCredentialBase64() {
	sessionDataJSON := map[string]interface{}{
		jsonKeyChallenge:          "challenge123",
		jsonKeyUserID:             base64.StdEncoding.EncodeToString([]byte("user123")),
		jsonKeyRelyingPartyID:     "example.com",
		jsonKeyAllowedCredentials: []interface{}{"invalid-base64!!!", "another-invalid!!!"},
	}
	jsonBytes, _ := json.Marshal(sessionDataJSON)

	row := map[string]interface{}{
		dbColumnUserID:         "user123",
		dbColumnRelyingPartyID: "example.com",
		dbColumnSessionData:    string(jsonBytes),
	}

	suite.mockDBProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil).Once()
	suite.mockDBClient.On("Query", mock.AnythingOfType("model.DBQuery"),
		testSessionKey, mock.AnythingOfType("time.Time"), "test-deployment-id").
		Return([]map[string]interface{}{row}, nil).Once()

	sessionData, userID, rpID, err := suite.store.retrieveSession(testSessionKey)

	suite.Error(err)
	suite.Nil(sessionData)
	suite.Equal("", userID)
	suite.Equal("", rpID)
	suite.Contains(err.Error(), "illegal base64 data")
}

func (suite *SessionStoreTestSuite) TestRetrieveSession_BuildSessionDataError_WrongSessionDataType() {
	row := map[string]interface{}{
		dbColumnUserID:         "user123",
		dbColumnRelyingPartyID: "example.com",
		dbColumnSessionData:    12345, // Invalid type: int instead of string or []byte
	}

	suite.mockDBProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil).Once()
	suite.mockDBClient.On("Query", mock.AnythingOfType("model.DBQuery"),
		testSessionKey, mock.AnythingOfType("time.Time"), "test-deployment-id").
		Return([]map[string]interface{}{row}, nil).Once()

	sessionData, userID, rpID, err := suite.store.retrieveSession(testSessionKey)

	suite.Error(err)
	suite.Nil(sessionData)
	suite.Equal("", userID)
	suite.Equal("", rpID)
	suite.Contains(err.Error(), "SESSION_DATA is missing or invalid")
}

func (suite *SessionStoreTestSuite) TestRetrieveSession_BuildSessionDataError_EmptySessionData() {
	row := map[string]interface{}{
		dbColumnUserID:         "user123",
		dbColumnRelyingPartyID: "example.com",
		dbColumnSessionData:    "", // Empty string
	}

	suite.mockDBProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil).Once()
	suite.mockDBClient.On("Query", mock.AnythingOfType("model.DBQuery"),
		testSessionKey, mock.AnythingOfType("time.Time"), "test-deployment-id").
		Return([]map[string]interface{}{row}, nil).Once()

	sessionData, userID, rpID, err := suite.store.retrieveSession(testSessionKey)

	suite.Error(err)
	suite.Nil(sessionData)
	suite.Equal("", userID)
	suite.Equal("", rpID)
	suite.Contains(err.Error(), "SESSION_DATA is missing or invalid")
}

func (suite *SessionStoreTestSuite) TestRetrieveSession_BuildSessionDataError_EmptyByteArray() {
	row := map[string]interface{}{
		dbColumnUserID:         "user123",
		dbColumnRelyingPartyID: "example.com",
		dbColumnSessionData:    []byte{}, // Empty byte array
	}

	suite.mockDBProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil).Once()
	suite.mockDBClient.On("Query", mock.AnythingOfType("model.DBQuery"),
		testSessionKey, mock.AnythingOfType("time.Time"), "test-deployment-id").
		Return([]map[string]interface{}{row}, nil).Once()

	sessionData, userID, rpID, err := suite.store.retrieveSession(testSessionKey)

	suite.Error(err)
	suite.Nil(sessionData)
	suite.Equal("", userID)
	suite.Equal("", rpID)
	suite.Contains(err.Error(), "SESSION_DATA is missing or invalid")
}
