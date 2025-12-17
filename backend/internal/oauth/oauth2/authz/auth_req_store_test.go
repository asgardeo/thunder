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

package authz

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/log"

	"github.com/asgardeo/thunder/internal/oauth/oauth2/model"
	"github.com/asgardeo/thunder/internal/system/config"

	"github.com/asgardeo/thunder/tests/mocks/database/providermock"
)

type AuthorizationRequestStoreTestSuite struct {
	suite.Suite
	mockdbProvider         *providermock.DBProviderInterfaceMock
	mockDBClient           *providermock.DBClientInterfaceMock
	store                  *authorizationRequestStore
	testAuthRequestContext authRequestContext
}

func TestAuthorizationRequestStoreTestSuite(t *testing.T) {
	suite.Run(t, new(AuthorizationRequestStoreTestSuite))
}

func (suite *AuthorizationRequestStoreTestSuite) SetupTest() {
	testConfig := &config.Config{
		Database: config.DatabaseConfig{
			Identity: config.DataSource{
				Type: "sqlite",
				Path: ":memory:",
			},
			Runtime: config.DataSource{
				Type: "sqlite",
				Path: ":memory:",
			},
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)

	suite.mockdbProvider = &providermock.DBProviderInterfaceMock{}
	suite.mockDBClient = &providermock.DBClientInterfaceMock{}

	suite.store = &authorizationRequestStore{
		dbProvider:     suite.mockdbProvider,
		validityPeriod: 10 * time.Minute,
		deploymentID:   testDeploymentID,
		logger:         log.GetLogger().With(log.String(log.LoggerKeyComponentName, "AuthorizationRequestStore")),
	}

	suite.testAuthRequestContext = authRequestContext{
		OAuthParameters: model.OAuthParameters{
			State:               "test-state",
			ClientID:            "test-client-id",
			RedirectURI:         "https://client.example.com/callback",
			ResponseType:        "code",
			StandardScopes:      []string{"openid", "profile"},
			PermissionScopes:    []string{"read", "write"},
			CodeChallenge:       "test-challenge",
			CodeChallengeMethod: "S256",
			Resource:            "https://api.example.com/resource",
		},
	}
}

func (suite *AuthorizationRequestStoreTestSuite) TestNewAuthorizationRequestStore() {
	store := newAuthorizationRequestStore()
	assert.NotNil(suite.T(), store)
	assert.Implements(suite.T(), (*authorizationRequestStoreInterface)(nil), store)
}

func (suite *AuthorizationRequestStoreTestSuite) TestAddRequest_Success() {
	suite.mockdbProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil)

	suite.mockDBClient.On("Execute", queryInsertAuthRequest,
		mock.MatchedBy(func(key string) bool {
			return len(key) > 0 // UUID should be generated
		}),
		mock.Anything, // JSON data bytes
		mock.MatchedBy(func(expiryTime time.Time) bool {
			// Expiry is calculated from time.Now() when storing
			// Allow 1 second tolerance for timing
			now := time.Now()
			expectedExpiry := now.Add(10 * time.Minute)
			diff := expiryTime.Sub(expectedExpiry)
			return diff >= -time.Second && diff <= time.Second
		}),
		testDeploymentID).
		Return(int64(1), nil)

	identifier := suite.store.AddRequest(suite.testAuthRequestContext)
	assert.NotEmpty(suite.T(), identifier)

	suite.mockdbProvider.AssertExpectations(suite.T())
	suite.mockDBClient.AssertExpectations(suite.T())
}

func (suite *AuthorizationRequestStoreTestSuite) TestAddRequest_DBClientError() {
	suite.mockdbProvider.On("GetRuntimeDBClient").Return(nil, errors.New("db client error"))

	identifier := suite.store.AddRequest(suite.testAuthRequestContext)
	assert.Empty(suite.T(), identifier)

	suite.mockdbProvider.AssertExpectations(suite.T())
}

func (suite *AuthorizationRequestStoreTestSuite) TestAddRequest_ExecuteError() {
	suite.mockdbProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil)

	suite.mockDBClient.On("Execute", queryInsertAuthRequest,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(int64(0), errors.New("execute error"))

	identifier := suite.store.AddRequest(suite.testAuthRequestContext)
	assert.Empty(suite.T(), identifier)

	suite.mockdbProvider.AssertExpectations(suite.T())
	suite.mockDBClient.AssertExpectations(suite.T())
}

func (suite *AuthorizationRequestStoreTestSuite) TestAddRequest_JSONMarshalingError() {
	suite.mockdbProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil)

	suite.mockDBClient.On("Execute", queryInsertAuthRequest,
		mock.Anything,
		mock.MatchedBy(func(data []byte) bool {
			// Verify the JSON structure
			var jsonData map[string]interface{}
			err := json.Unmarshal(data, &jsonData)
			return err == nil &&
				jsonData["state"] == "test-state" &&
				jsonData["client_id"] == "test-client-id" &&
				jsonData["redirect_uri"] == "https://client.example.com/callback"
		}),
		mock.Anything,
		testDeploymentID).
		Return(int64(1), nil)

	identifier := suite.store.AddRequest(suite.testAuthRequestContext)
	assert.NotEmpty(suite.T(), identifier)

	suite.mockdbProvider.AssertExpectations(suite.T())
	suite.mockDBClient.AssertExpectations(suite.T())
}

func (suite *AuthorizationRequestStoreTestSuite) TestGetRequest_Success() {
	requestData := map[string]interface{}{
		"state":                 "test-state",
		"client_id":             "test-client-id",
		"redirect_uri":          "https://client.example.com/callback",
		"response_type":         "code",
		"standard_scopes":       []interface{}{"openid", "profile"},
		"permission_scopes":     []interface{}{"read", "write"},
		"code_challenge":        "test-challenge",
		"code_challenge_method": "S256",
		"resource":              "https://api.example.com/resource",
	}
	requestDataJSON, _ := json.Marshal(requestData)

	expiryTime := time.Now().Add(10 * time.Minute)

	suite.mockdbProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil)

	suite.mockDBClient.On("Query", queryGetAuthRequest, "test-request-id", mock.Anything, testDeploymentID).
		Return([]map[string]interface{}{
			{
				"auth_id":      "test-request-id",
				"request_data": string(requestDataJSON),
				"expiry_time":  expiryTime.Format("2006-01-02 15:04:05.999999999"),
			},
		}, nil)

	ok, result := suite.store.GetRequest("test-request-id")
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "test-state", result.OAuthParameters.State)
	assert.Equal(suite.T(), "test-client-id", result.OAuthParameters.ClientID)
	assert.Equal(suite.T(), "https://client.example.com/callback", result.OAuthParameters.RedirectURI)
	assert.Equal(suite.T(), "code", result.OAuthParameters.ResponseType)
	assert.Equal(suite.T(), []string{"openid", "profile"}, result.OAuthParameters.StandardScopes)
	assert.Equal(suite.T(), []string{"read", "write"}, result.OAuthParameters.PermissionScopes)
	assert.Equal(suite.T(), "test-challenge", result.OAuthParameters.CodeChallenge)
	assert.Equal(suite.T(), "S256", result.OAuthParameters.CodeChallengeMethod)
	assert.Equal(suite.T(), "https://api.example.com/resource", result.OAuthParameters.Resource)

	suite.mockdbProvider.AssertExpectations(suite.T())
	suite.mockDBClient.AssertExpectations(suite.T())
}

func (suite *AuthorizationRequestStoreTestSuite) TestGetRequest_EmptyKey() {
	ok, _ := suite.store.GetRequest("")
	assert.False(suite.T(), ok)
}

func (suite *AuthorizationRequestStoreTestSuite) TestGetRequest_DBClientError() {
	suite.mockdbProvider.On("GetRuntimeDBClient").Return(nil, errors.New("db client error"))

	ok, _ := suite.store.GetRequest("test-request-id")
	assert.False(suite.T(), ok)

	suite.mockdbProvider.AssertExpectations(suite.T())
}

func (suite *AuthorizationRequestStoreTestSuite) TestGetRequest_QueryError() {
	suite.mockdbProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil)

	suite.mockDBClient.On("Query", queryGetAuthRequest, "test-request-id", mock.Anything, testDeploymentID).
		Return(nil, errors.New("query error"))

	ok, _ := suite.store.GetRequest("test-request-id")
	assert.False(suite.T(), ok)

	suite.mockdbProvider.AssertExpectations(suite.T())
	suite.mockDBClient.AssertExpectations(suite.T())
}

func (suite *AuthorizationRequestStoreTestSuite) TestGetRequest_NoResults() {
	suite.mockdbProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil)

	suite.mockDBClient.On("Query", queryGetAuthRequest, "test-request-id", mock.Anything, testDeploymentID).
		Return([]map[string]interface{}{}, nil)

	ok, _ := suite.store.GetRequest("test-request-id")
	assert.False(suite.T(), ok)

	suite.mockdbProvider.AssertExpectations(suite.T())
	suite.mockDBClient.AssertExpectations(suite.T())
}

func (suite *AuthorizationRequestStoreTestSuite) TestGetRequest_Expired() {
	// Query with expiry check should return no results if expired
	suite.mockdbProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil)

	suite.mockDBClient.On("Query", queryGetAuthRequest, "test-request-id", mock.Anything, testDeploymentID).
		Return([]map[string]interface{}{}, nil)

	ok, _ := suite.store.GetRequest("test-request-id")
	assert.False(suite.T(), ok)

	suite.mockdbProvider.AssertExpectations(suite.T())
	suite.mockDBClient.AssertExpectations(suite.T())
}

func (suite *AuthorizationRequestStoreTestSuite) TestGetRequest_MissingRequestData() {
	suite.mockdbProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil)

	suite.mockDBClient.On("Query", queryGetAuthRequest, "test-request-id", mock.Anything, testDeploymentID).
		Return([]map[string]interface{}{
			{
				"auth_id":     "test-request-id",
				"expiry_time": time.Now().Add(10 * time.Minute).Format("2006-01-02 15:04:05.999999999"),
				// Missing request_data
			},
		}, nil)

	ok, _ := suite.store.GetRequest("test-request-id")
	assert.False(suite.T(), ok)

	suite.mockdbProvider.AssertExpectations(suite.T())
	suite.mockDBClient.AssertExpectations(suite.T())
}

func (suite *AuthorizationRequestStoreTestSuite) TestGetRequest_EmptyRequestDataString() {
	suite.mockdbProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil)

	suite.mockDBClient.On("Query", queryGetAuthRequest, "test-request-id", mock.Anything, testDeploymentID).
		Return([]map[string]interface{}{
			{
				"auth_id":      "test-request-id",
				"request_data": "", // Empty string
				"expiry_time":  time.Now().Add(10 * time.Minute).Format("2006-01-02 15:04:05.999999999"),
			},
		}, nil)

	ok, _ := suite.store.GetRequest("test-request-id")
	assert.False(suite.T(), ok)

	suite.mockdbProvider.AssertExpectations(suite.T())
	suite.mockDBClient.AssertExpectations(suite.T())
}

func (suite *AuthorizationRequestStoreTestSuite) TestGetRequest_RequestDataAsBytes() {
	requestData := map[string]interface{}{
		"state":        "test-state",
		"client_id":    "test-client-id",
		"redirect_uri": "https://client.example.com/callback",
	}
	requestDataJSON, _ := json.Marshal(requestData)

	expiryTime := time.Now().Add(10 * time.Minute)

	suite.mockdbProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil)

	suite.mockDBClient.On("Query", queryGetAuthRequest, "test-request-id", mock.Anything, testDeploymentID).
		Return([]map[string]interface{}{
			{
				"auth_id":      "test-request-id",
				"request_data": requestDataJSON, // Byte array
				"expiry_time":  expiryTime.Format("2006-01-02 15:04:05.999999999"),
			},
		}, nil)

	ok, result := suite.store.GetRequest("test-request-id")
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "test-state", result.OAuthParameters.State)
	assert.Equal(suite.T(), "test-client-id", result.OAuthParameters.ClientID)

	suite.mockdbProvider.AssertExpectations(suite.T())
	suite.mockDBClient.AssertExpectations(suite.T())
}

func (suite *AuthorizationRequestStoreTestSuite) TestGetRequest_InvalidRequestDataJSON() {
	suite.mockdbProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil)

	suite.mockDBClient.On("Query", queryGetAuthRequest, "test-request-id", mock.Anything, testDeploymentID).
		Return([]map[string]interface{}{
			{
				"auth_id":      "test-request-id",
				"request_data": "{invalid json", // Invalid JSON
				"expiry_time":  time.Now().Add(10 * time.Minute).Format("2006-01-02 15:04:05.999999999"),
			},
		}, nil)

	ok, _ := suite.store.GetRequest("test-request-id")
	assert.False(suite.T(), ok)

	suite.mockdbProvider.AssertExpectations(suite.T())
	suite.mockDBClient.AssertExpectations(suite.T())
}

func (suite *AuthorizationRequestStoreTestSuite) TestGetRequest_EmptyScopes() {
	requestData := map[string]interface{}{
		"state":        "test-state",
		"client_id":    "test-client-id",
		"redirect_uri": "https://client.example.com/callback",
		// No scopes - should default to empty slices
	}
	requestDataJSON, _ := json.Marshal(requestData)

	expiryTime := time.Now().Add(10 * time.Minute)

	suite.mockdbProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil)

	suite.mockDBClient.On("Query", queryGetAuthRequest, "test-request-id", mock.Anything, testDeploymentID).
		Return([]map[string]interface{}{
			{
				"auth_id":      "test-request-id",
				"request_data": string(requestDataJSON),
				"expiry_time":  expiryTime.Format("2006-01-02 15:04:05.999999999"),
			},
		}, nil)

	ok, result := suite.store.GetRequest("test-request-id")
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), []string{}, result.OAuthParameters.StandardScopes)
	assert.Equal(suite.T(), []string{}, result.OAuthParameters.PermissionScopes)

	suite.mockdbProvider.AssertExpectations(suite.T())
	suite.mockDBClient.AssertExpectations(suite.T())
}

func (suite *AuthorizationRequestStoreTestSuite) TestGetRequest_NilScopes() {
	requestData := map[string]interface{}{
		"state":             "test-state",
		"client_id":         "test-client-id",
		"redirect_uri":      "https://client.example.com/callback",
		"standard_scopes":   nil,
		"permission_scopes": nil,
	}
	requestDataJSON, _ := json.Marshal(requestData)

	expiryTime := time.Now().Add(10 * time.Minute)

	suite.mockdbProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil)

	suite.mockDBClient.On("Query", queryGetAuthRequest, "test-request-id", mock.Anything, testDeploymentID).
		Return([]map[string]interface{}{
			{
				"auth_id":      "test-request-id",
				"request_data": string(requestDataJSON),
				"expiry_time":  expiryTime.Format("2006-01-02 15:04:05.999999999"),
			},
		}, nil)

	ok, result := suite.store.GetRequest("test-request-id")
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), []string{}, result.OAuthParameters.StandardScopes)
	assert.Equal(suite.T(), []string{}, result.OAuthParameters.PermissionScopes)

	suite.mockdbProvider.AssertExpectations(suite.T())
	suite.mockDBClient.AssertExpectations(suite.T())
}

func (suite *AuthorizationRequestStoreTestSuite) TestGetRequest_StringScopes() {
	requestData := map[string]interface{}{
		"state":             "test-state",
		"client_id":         "test-client-id",
		"standard_scopes":   []string{"openid", "profile"},
		"permission_scopes": []string{"read", "write"},
	}
	requestDataJSON, _ := json.Marshal(requestData)

	expiryTime := time.Now().Add(10 * time.Minute)

	suite.mockdbProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil)

	suite.mockDBClient.On("Query", queryGetAuthRequest, "test-request-id", mock.Anything, testDeploymentID).
		Return([]map[string]interface{}{
			{
				"auth_id":      "test-request-id",
				"request_data": string(requestDataJSON),
				"expiry_time":  expiryTime.Format("2006-01-02 15:04:05.999999999"),
			},
		}, nil)

	ok, result := suite.store.GetRequest("test-request-id")
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), []string{"openid", "profile"}, result.OAuthParameters.StandardScopes)
	assert.Equal(suite.T(), []string{"read", "write"}, result.OAuthParameters.PermissionScopes)

	suite.mockdbProvider.AssertExpectations(suite.T())
	suite.mockDBClient.AssertExpectations(suite.T())
}

func (suite *AuthorizationRequestStoreTestSuite) TestClearRequest_Success() {
	suite.mockdbProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil)

	suite.mockDBClient.On("Execute", queryDeleteAuthRequest, "test-request-id", testDeploymentID).
		Return(int64(1), nil)

	suite.store.ClearRequest("test-request-id")

	suite.mockdbProvider.AssertExpectations(suite.T())
	suite.mockDBClient.AssertExpectations(suite.T())
}

func (suite *AuthorizationRequestStoreTestSuite) TestClearRequest_EmptyKey() {
	// Should return early without calling DB
	suite.store.ClearRequest("")

	// No expectations set, so this should pass
}

func (suite *AuthorizationRequestStoreTestSuite) TestClearRequest_DBClientError() {
	suite.mockdbProvider.On("GetRuntimeDBClient").Return(nil, errors.New("db client error"))

	suite.store.ClearRequest("test-request-id")

	suite.mockdbProvider.AssertExpectations(suite.T())
}

func (suite *AuthorizationRequestStoreTestSuite) TestClearRequest_ExecuteError() {
	suite.mockdbProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil)

	suite.mockDBClient.On("Execute", queryDeleteAuthRequest, "test-request-id", testDeploymentID).
		Return(int64(0), errors.New("execute error"))

	suite.store.ClearRequest("test-request-id")

	suite.mockdbProvider.AssertExpectations(suite.T())
	suite.mockDBClient.AssertExpectations(suite.T())
}

func (suite *AuthorizationRequestStoreTestSuite) TestParseTimeFieldForAuthRequest_StringInput() {
	testTime := "2023-12-01 10:30:45.123456789"
	expectedTime, _ := time.Parse("2006-01-02 15:04:05.999999999", testTime)

	result, err := parseTimeField(testTime, "test_field")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedTime, result)
}

func (suite *AuthorizationRequestStoreTestSuite) TestParseTimeFieldForAuthRequest_StringWithExtraContent() {
	testTime := "2023-12-01 10:30:45.123456789 extra content"
	expectedTime, _ := time.Parse("2006-01-02 15:04:05.999999999", "2023-12-01 10:30:45.123456789")

	result, err := parseTimeField(testTime, "test_field")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedTime, result)
}

func (suite *AuthorizationRequestStoreTestSuite) TestParseTimeFieldForAuthRequest_TimeInput() {
	testTime := time.Now()

	result, err := parseTimeField(testTime, "test_field")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), testTime, result)
}

func (suite *AuthorizationRequestStoreTestSuite) TestParseTimeFieldForAuthRequest_InvalidStringFormat() {
	testTime := "invalid-time-format"

	result, err := parseTimeField(testTime, "test_field")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "error parsing test_field")
	assert.True(suite.T(), result.IsZero())
}

func (suite *AuthorizationRequestStoreTestSuite) TestParseTimeFieldForAuthRequest_InvalidType() {
	result, err := parseTimeField(12345, "test_field")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "unexpected type for test_field")
	assert.True(suite.T(), result.IsZero())
}

func (suite *AuthorizationRequestStoreTestSuite) TestTrimTimeStringForAuthRequest() {
	input := "2023-12-01 10:30:45.123456789 extra content here"
	expected := "2023-12-01 10:30:45.123456789"

	result := trimTimeString(input)
	assert.Equal(suite.T(), expected, result)
}

func (suite *AuthorizationRequestStoreTestSuite) TestTrimTimeStringForAuthRequest_ShortInput() {
	input := "2023-12-01"

	result := trimTimeString(input)
	assert.Equal(suite.T(), input, result)
}

func (suite *AuthorizationRequestStoreTestSuite) TestConvertToStringArray() {
	input := []interface{}{"one", "two", "three"}
	expected := []string{"one", "two", "three"}

	result := convertToStringArray(input)
	assert.Equal(suite.T(), expected, result)
}

func (suite *AuthorizationRequestStoreTestSuite) TestConvertToStringArray_WithNonStringValues() {
	input := []interface{}{"one", 123, "three", true}
	expected := []string{"one", "three"} // Only strings are included

	result := convertToStringArray(input)
	assert.Equal(suite.T(), expected, result)
}

func (suite *AuthorizationRequestStoreTestSuite) TestConvertToStringArray_Empty() {
	input := []interface{}{}
	expected := []string{}

	result := convertToStringArray(input)
	assert.Equal(suite.T(), expected, result)
}

func (suite *AuthorizationRequestStoreTestSuite) TestGetRequest_EmptyByteArray() {
	suite.mockdbProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil)

	suite.mockDBClient.On("Query", queryGetAuthRequest, "test-request-id", mock.Anything, testDeploymentID).
		Return([]map[string]interface{}{
			{
				"auth_id":      "test-request-id",
				"request_data": []byte{}, // Empty byte array
				"expiry_time":  time.Now().Add(10 * time.Minute).Format("2006-01-02 15:04:05.999999999"),
			},
		}, nil)

	ok, _ := suite.store.GetRequest("test-request-id")
	assert.False(suite.T(), ok)

	suite.mockdbProvider.AssertExpectations(suite.T())
	suite.mockDBClient.AssertExpectations(suite.T())
}

func (suite *AuthorizationRequestStoreTestSuite) TestGetRequest_RequestDataAsUnexpectedType() {
	suite.mockdbProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil)

	suite.mockDBClient.On("Query", queryGetAuthRequest, "test-request-id", mock.Anything, testDeploymentID).
		Return([]map[string]interface{}{
			{
				"auth_id":      "test-request-id",
				"request_data": 12345, // Unexpected type (int)
				"expiry_time":  time.Now().Add(10 * time.Minute).Format("2006-01-02 15:04:05.999999999"),
			},
		}, nil)

	ok, _ := suite.store.GetRequest("test-request-id")
	assert.False(suite.T(), ok)

	suite.mockdbProvider.AssertExpectations(suite.T())
	suite.mockDBClient.AssertExpectations(suite.T())
}

func (suite *AuthorizationRequestStoreTestSuite) TestParseTimeFieldForAuthRequest_AlternativeFormat() {
	// Test ISO 8601 format when custom format fails
	testTime := "2023-12-01T10:30:45Z"
	expectedTime, _ := time.Parse("2006-01-02T15:04:05Z07:00", testTime)

	result, err := parseTimeField(testTime, "test_field")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedTime, result)
}

func (suite *AuthorizationRequestStoreTestSuite) TestParseTimeFieldForAuthRequest_AlternativeFormatWithTimezone() {
	// Test ISO 8601 format with timezone
	testTime := "2023-12-01T10:30:45+05:30"
	expectedTime, _ := time.Parse("2006-01-02T15:04:05Z07:00", testTime)

	result, err := parseTimeField(testTime, "test_field")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedTime, result)
}

func (suite *AuthorizationRequestStoreTestSuite) TestGetRequest_AllOptionalFieldsMissing() {
	// Test when optional fields are missing from JSON
	requestData := map[string]interface{}{
		"state": "test-state",
		// Missing client_id, redirect_uri, etc.
	}
	requestDataJSON, _ := json.Marshal(requestData)

	expiryTime := time.Now().Add(10 * time.Minute)

	suite.mockdbProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil)

	suite.mockDBClient.On("Query", queryGetAuthRequest, "test-request-id", mock.Anything, testDeploymentID).
		Return([]map[string]interface{}{
			{
				"auth_id":      "test-request-id",
				"request_data": string(requestDataJSON),
				"expiry_time":  expiryTime.Format("2006-01-02 15:04:05.999999999"),
			},
		}, nil)

	ok, result := suite.store.GetRequest("test-request-id")
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "test-state", result.OAuthParameters.State)
	assert.Empty(suite.T(), result.OAuthParameters.ClientID)
	assert.Empty(suite.T(), result.OAuthParameters.RedirectURI)
	assert.Equal(suite.T(), []string{}, result.OAuthParameters.StandardScopes)
	assert.Equal(suite.T(), []string{}, result.OAuthParameters.PermissionScopes)

	suite.mockdbProvider.AssertExpectations(suite.T())
	suite.mockDBClient.AssertExpectations(suite.T())
}

func (suite *AuthorizationRequestStoreTestSuite) TestGetRequest_NonStringScopes() {
	// Test when scopes are not string arrays but other types
	requestData := map[string]interface{}{
		"state":             "test-state",
		"client_id":         "test-client-id",
		"standard_scopes":   "not-an-array", // Wrong type
		"permission_scopes": 12345,          // Wrong type
	}
	requestDataJSON, _ := json.Marshal(requestData)

	expiryTime := time.Now().Add(10 * time.Minute)

	suite.mockdbProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil)

	suite.mockDBClient.On("Query", queryGetAuthRequest, "test-request-id", mock.Anything, testDeploymentID).
		Return([]map[string]interface{}{
			{
				"auth_id":      "test-request-id",
				"request_data": string(requestDataJSON),
				"expiry_time":  expiryTime.Format("2006-01-02 15:04:05.999999999"),
			},
		}, nil)

	ok, result := suite.store.GetRequest("test-request-id")
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "test-state", result.OAuthParameters.State)
	assert.Equal(suite.T(), []string{}, result.OAuthParameters.StandardScopes)   // Should default to empty
	assert.Equal(suite.T(), []string{}, result.OAuthParameters.PermissionScopes) // Should default to empty

	suite.mockdbProvider.AssertExpectations(suite.T())
	suite.mockDBClient.AssertExpectations(suite.T())
}

func (suite *AuthorizationRequestStoreTestSuite) TestGetRequest_StandardScopesOtherType() {
	// Test when standard_scopes is neither []interface{}, []string, nor nil
	requestData := map[string]interface{}{
		"state":           "test-state",
		"standard_scopes": map[string]string{"key": "value"}, // Wrong type - map instead of array
	}
	suite.testGetRequestWithInvalidScopesType(requestData, func(result authRequestContext) {
		assert.Equal(suite.T(), []string{}, result.OAuthParameters.StandardScopes) // Should default to empty
	})
}

func (suite *AuthorizationRequestStoreTestSuite) TestGetRequest_PermissionScopesOtherType() {
	// Test when permission_scopes is neither []interface{}, []string, nor nil
	requestData := map[string]interface{}{
		"state":             "test-state",
		"permission_scopes": map[string]string{"key": "value"}, // Wrong type - map instead of array
	}
	suite.testGetRequestWithInvalidScopesType(requestData, func(result authRequestContext) {
		assert.Equal(suite.T(), []string{}, result.OAuthParameters.PermissionScopes) // Should default to empty
	})
}

// testGetRequestWithInvalidScopesType is a helper function to test GetRequest with invalid scope types
func (suite *AuthorizationRequestStoreTestSuite) testGetRequestWithInvalidScopesType(
	requestData map[string]interface{},
	assertFn func(authRequestContext),
) {
	requestDataJSON, _ := json.Marshal(requestData)

	expiryTime := time.Now().Add(10 * time.Minute)

	suite.mockdbProvider.On("GetRuntimeDBClient").Return(suite.mockDBClient, nil)

	suite.mockDBClient.On("Query", queryGetAuthRequest, "test-request-id", mock.Anything, testDeploymentID).
		Return([]map[string]interface{}{
			{
				"auth_id":      "test-request-id",
				"request_data": string(requestDataJSON),
				"expiry_time":  expiryTime.Format("2006-01-02 15:04:05.999999999"),
			},
		}, nil)

	ok, result := suite.store.GetRequest("test-request-id")
	assert.True(suite.T(), ok)
	assertFn(result)

	suite.mockdbProvider.AssertExpectations(suite.T())
	suite.mockDBClient.AssertExpectations(suite.T())
}
