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

package application

import (
	"context"

	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/application/model"
	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/tests/mocks/database/providermock"
)

const (
	testServerID = "test-server-id"
	testEntityID = "entity-1"
)

type mockTransactioner struct{}

func (m *mockTransactioner) Transact(ctx context.Context, operation func(txCtx context.Context) error) error {
	return operation(ctx)
}

// ApplicationStoreTestSuite contains comprehensive tests for the application store helper functions.
type ApplicationStoreTestSuite struct {
	suite.Suite
	mockDBProvider *providermock.DBProviderInterfaceMock
	mockDBClient   *providermock.DBClientInterfaceMock
	store          *applicationStore
}

func TestApplicationStoreTestSuite(t *testing.T) {
	suite.Run(t, new(ApplicationStoreTestSuite))
}

func (suite *ApplicationStoreTestSuite) SetupTest() {
	_ = config.InitializeThunderRuntime("test", &config.Config{})
	suite.mockDBProvider = providermock.NewDBProviderInterfaceMock(suite.T())
	suite.mockDBClient = providermock.NewDBClientInterfaceMock(suite.T())
	suite.store = &applicationStore{
		dbProvider:   suite.mockDBProvider,
		deploymentID: testServerID,
	}
}

func (suite *ApplicationStoreTestSuite) createTestApplication() model.ApplicationProcessedDTO {
	return model.ApplicationProcessedDTO{
		ID:                        "app1",
		Name:                      "Test App 1",
		Description:               "Test application description",
		AuthFlowID:                "auth_flow_1",
		RegistrationFlowID:        "reg_flow_1",
		IsRegistrationFlowEnabled: true,
		URL:                       "https://example.com",
		LogoURL:                   "https://example.com/logo.png",
		TosURI:                    "https://example.com/tos",
		PolicyURI:                 "https://example.com/policy",
		Contacts:                  []string{"contact@example.com", "support@example.com"},
		Assertion: &model.AssertionConfig{
			ValidityPeriod: 3600,
			UserAttributes: []string{"email", "name", "sub"},
		},
		InboundAuthConfig: []model.InboundAuthConfigProcessedDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigProcessedDTO{
					AppID:        "app1",
					ClientID:     "client_app1",
					RedirectURIs: []string{"https://example.com/callback", "https://example.com/cb2"},
					GrantTypes: []oauth2const.GrantType{
						oauth2const.GrantTypeAuthorizationCode,
						oauth2const.GrantTypeRefreshToken,
					},
					ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretPost,
					PKCERequired:            true,
					PublicClient:            false,
					Scopes:                  []string{"openid", "profile", "email"},
					Token: &model.OAuthTokenConfig{
						AccessToken: &model.AccessTokenConfig{
							ValidityPeriod: 7200,
							UserAttributes: []string{"sub", "email", "name"},
						},
						IDToken: &model.IDTokenConfig{
							ValidityPeriod: 3600,
							UserAttributes: []string{"sub", "email", "name", "given_name"},
						},
					},
					ScopeClaims: map[string][]string{
						"profile": {"name", "given_name", "family_name"},
						"email":   {"email", "email_verified"},
					},
				},
			},
		},
	}
}

func (suite *ApplicationStoreTestSuite) TestNewApplicationStore() {
	mockClient := providermock.NewDBClientInterfaceMock(suite.T())
	mockClient.On("QueryContext", mock.Anything, queryGetApplicationCount, mock.Anything).
		Return([]map[string]interface{}{{"total": int64(0)}}, nil)
	mockProvider := providermock.NewDBProviderInterfaceMock(suite.T())
	mockProvider.On("GetConfigDBClient").Return(mockClient, nil)
	mockProvider.On("GetConfigDBTransactioner").Return(&mockTransactioner{}, nil)
	originalGetDBProvider := getDBProvider
	getDBProvider = func() provider.DBProviderInterface { return mockProvider }
	defer func() { getDBProvider = originalGetDBProvider }()

	store, _, err := newApplicationStore()

	suite.NoError(err)
	suite.NotNil(store)
	suite.IsType(&applicationStore{}, store)
}

// --- Tests for getOAuthConfigJSONBytes ---

func (suite *ApplicationStoreTestSuite) TestGetOAuthConfigJSONBytes_Success() {
	app := suite.createTestApplication()
	inboundAuthConfig := app.InboundAuthConfig[0]

	jsonBytes, err := getOAuthConfigJSONBytes(inboundAuthConfig)

	suite.NoError(err)
	suite.NotNil(jsonBytes)

	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	suite.NoError(err)

	suite.NotNil(result["redirect_uris"])
	suite.NotNil(result["grant_types"])
	suite.NotNil(result["response_types"])
	suite.Equal("client_secret_post", result["token_endpoint_auth_method"])
	suite.Equal(true, result["pkce_required"])
	suite.Equal(false, result["public_client"])

	redirectURIs, ok := result["redirect_uris"].([]interface{})
	suite.True(ok)
	suite.Len(redirectURIs, 2)

	token, ok := result["token"].(map[string]interface{})
	suite.True(ok)
	suite.Nil(token["issuer"])

	accessToken, ok := token["access_token"].(map[string]interface{})
	suite.True(ok)
	suite.Equal(float64(7200), accessToken["validity_period"])

	idToken, ok := token["id_token"].(map[string]interface{})
	suite.True(ok)
	suite.Equal(float64(3600), idToken["validity_period"])
	suite.NotNil(result["scope_claims"])
}

func (suite *ApplicationStoreTestSuite) TestGetOAuthConfigJSONBytes_WithoutToken() {
	app := suite.createTestApplication()
	inboundAuthConfig := app.InboundAuthConfig[0]
	inboundAuthConfig.OAuthAppConfig.Token = nil

	jsonBytes, err := getOAuthConfigJSONBytes(inboundAuthConfig)

	suite.NoError(err)
	suite.NotNil(jsonBytes)

	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	suite.NoError(err)
	suite.Nil(result["token"])
}

func (suite *ApplicationStoreTestSuite) TestGetOAuthConfigJSONBytes_WithoutAccessToken() {
	app := suite.createTestApplication()
	inboundAuthConfig := app.InboundAuthConfig[0]
	inboundAuthConfig.OAuthAppConfig.Token.AccessToken = nil

	jsonBytes, err := getOAuthConfigJSONBytes(inboundAuthConfig)

	suite.NoError(err)
	suite.NotNil(jsonBytes)

	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	suite.NoError(err)

	token, ok := result["token"].(map[string]interface{})
	suite.True(ok)
	suite.Nil(token["access_token"])
	suite.NotNil(token["id_token"])
}

func (suite *ApplicationStoreTestSuite) TestGetOAuthConfigJSONBytes_WithoutIDToken() {
	app := suite.createTestApplication()
	inboundAuthConfig := app.InboundAuthConfig[0]
	inboundAuthConfig.OAuthAppConfig.Token.IDToken = nil

	jsonBytes, err := getOAuthConfigJSONBytes(inboundAuthConfig)

	suite.NoError(err)
	suite.NotNil(jsonBytes)

	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	suite.NoError(err)

	token, ok := result["token"].(map[string]interface{})
	suite.True(ok)
	suite.NotNil(token["access_token"])
	suite.Nil(token["id_token"])
}

func (suite *ApplicationStoreTestSuite) TestGetOAuthConfigJSONBytes_EmptyScopes() {
	app := suite.createTestApplication()
	inboundAuthConfig := app.InboundAuthConfig[0]
	inboundAuthConfig.OAuthAppConfig.Scopes = []string{}

	jsonBytes, err := getOAuthConfigJSONBytes(inboundAuthConfig)

	suite.NoError(err)
	suite.NotNil(jsonBytes)

	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	suite.NoError(err)

	// Empty slice is marshaled as an empty array in JSON
	if scopes, ok := result["scopes"].([]interface{}); ok {
		suite.Len(scopes, 0)
	} else {
		// JSON unmarshaling might return nil for empty arrays
		suite.Nil(result["scopes"])
	}
}

func (suite *ApplicationStoreTestSuite) TestGetOAuthConfigJSONBytes_WithNilScopes() {
	app := suite.createTestApplication()
	inboundAuthConfig := app.InboundAuthConfig[0]
	inboundAuthConfig.OAuthAppConfig.Scopes = nil

	jsonBytes, err := getOAuthConfigJSONBytes(inboundAuthConfig)

	suite.NoError(err)
	suite.NotNil(jsonBytes)

	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	suite.NoError(err)
	// nil scopes should be marshaled as null or empty array
	if scopes, ok := result["scopes"].([]interface{}); ok {
		suite.Len(scopes, 0)
	} else {
		suite.Nil(result["scopes"])
	}
}

func (suite *ApplicationStoreTestSuite) TestGetOAuthConfigJSONBytes_WithAccessTokenNilUserAttributes() {
	app := suite.createTestApplication()
	inboundAuthConfig := app.InboundAuthConfig[0]
	inboundAuthConfig.OAuthAppConfig.Token.AccessToken.UserAttributes = nil

	jsonBytes, err := getOAuthConfigJSONBytes(inboundAuthConfig)

	suite.NoError(err)
	suite.NotNil(jsonBytes)

	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	suite.NoError(err)

	token, ok := result["token"].(map[string]interface{})
	suite.True(ok)
	accessToken, ok := token["access_token"].(map[string]interface{})
	suite.True(ok)
	suite.Equal(float64(7200), accessToken["validity_period"])
	// nil UserAttributes should be marshaled as null or empty array
	if userAttrs, ok := accessToken["user_attributes"].([]interface{}); ok {
		suite.Len(userAttrs, 0)
	} else {
		suite.Nil(accessToken["user_attributes"])
	}
}

func (suite *ApplicationStoreTestSuite) TestGetOAuthConfigJSONBytes_WithUserInfo() {
	app := suite.createTestApplication()
	inboundAuthConfig := app.InboundAuthConfig[0]
	inboundAuthConfig.OAuthAppConfig.UserInfo = &model.UserInfoConfig{
		ResponseType:   "jwt",
		UserAttributes: []string{"email", "name"},
	}

	jsonBytes, err := getOAuthConfigJSONBytes(inboundAuthConfig)

	suite.NoError(err)
	suite.NotNil(jsonBytes)

	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	suite.NoError(err)

	userInfo, ok := result["user_info"].(map[string]interface{})
	suite.True(ok)
	suite.Equal("jwt", userInfo["response_type"])

	userAttrs, ok := userInfo["user_attributes"].([]interface{})
	suite.True(ok)
	suite.Len(userAttrs, 2)
}

// --- Tests for buildAppConfigFromRow ---

func (suite *ApplicationStoreTestSuite) TestBuildAppConfigFromRow_Success() {
	assertionJSON, _ := json.Marshal(model.AssertionConfig{
		ValidityPeriod: 3600,
		UserAttributes: []string{"email", "name"},
	})
	loginConsentJSON, _ := json.Marshal(model.LoginConsentConfig{
		ValidityPeriod: 5400,
	})
	allowedTypesJSON, _ := json.Marshal([]string{"admin", "user"})
	propertiesJSON, _ := json.Marshal(map[string]interface{}{"template": "spa"})

	row := map[string]interface{}{
		"id":                           "app1",
		"auth_flow_id":                 "auth_flow_1",
		"registration_flow_id":         "reg_flow_1",
		"is_registration_flow_enabled": "1",
		"theme_id":                     "theme-123",
		"layout_id":                    "layout-456",
		"assertion":                    string(assertionJSON),
		"login_consent":                string(loginConsentJSON),
		"allowed_entity_types":         string(allowedTypesJSON),
		"properties":                   string(propertiesJSON),
	}

	result, err := buildAppConfigFromRow(row)

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal("app1", result.ID)
	suite.Equal("auth_flow_1", result.AuthFlowID)
	suite.Equal("reg_flow_1", result.RegistrationFlowID)
	suite.True(result.IsRegistrationFlowEnabled)
	suite.Equal("theme-123", result.ThemeID)
	suite.Equal("layout-456", result.LayoutID)
	suite.NotNil(result.Assertion)
	suite.Equal(int64(3600), result.Assertion.ValidityPeriod)
	suite.NotNil(result.LoginConsent)
	suite.Equal(int64(5400), result.LoginConsent.ValidityPeriod)
	suite.Equal([]string{"admin", "user"}, result.AllowedEntityTypes)
	suite.NotNil(result.Properties)
	suite.Equal("spa", result.Properties["template"])
}

func (suite *ApplicationStoreTestSuite) TestBuildAppConfigFromRow_InvalidID() {
	row := map[string]interface{}{
		"id": 123, // Invalid type
	}

	result, err := buildAppConfigFromRow(row)

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "failed to parse id as string")
}

func (suite *ApplicationStoreTestSuite) TestBuildAppConfigFromRow_MinimalRow() {
	row := map[string]interface{}{
		"id":                           "app1",
		"auth_flow_id":                 nil,
		"registration_flow_id":         nil,
		"is_registration_flow_enabled": nil,
		"theme_id":                     nil,
		"layout_id":                    nil,
		"assertion":                    nil,
		"login_consent":                nil,
		"allowed_entity_types":         nil,
		"properties":                   nil,
	}

	result, err := buildAppConfigFromRow(row)

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal("app1", result.ID)
	suite.Equal("", result.AuthFlowID)
	suite.Equal("", result.RegistrationFlowID)
	suite.False(result.IsRegistrationFlowEnabled)
	suite.Nil(result.Assertion)
	suite.Nil(result.LoginConsent)
	suite.Nil(result.AllowedEntityTypes)
	suite.Nil(result.Properties)
}

func (suite *ApplicationStoreTestSuite) TestBuildAppConfigFromRow_WithByteRegistrationFlag() {
	row := map[string]interface{}{
		"id":                           "app1",
		"auth_flow_id":                 "flow1",
		"registration_flow_id":         "reg1",
		"is_registration_flow_enabled": []byte("1"),
		"theme_id":                     nil,
		"layout_id":                    nil,
	}

	result, err := buildAppConfigFromRow(row)

	suite.NoError(err)
	suite.True(result.IsRegistrationFlowEnabled)
}

func (suite *ApplicationStoreTestSuite) TestBuildAppConfigFromRow_WithZeroRegistrationFlag() {
	row := map[string]interface{}{
		"id":                           "app1",
		"auth_flow_id":                 "flow1",
		"registration_flow_id":         "reg1",
		"is_registration_flow_enabled": "0",
		"theme_id":                     nil,
		"layout_id":                    nil,
	}

	result, err := buildAppConfigFromRow(row)

	suite.NoError(err)
	suite.False(result.IsRegistrationFlowEnabled)
}

// --- Tests for buildOAuthConfigFromRow ---

func (suite *ApplicationStoreTestSuite) TestBuildOAuthConfigFromRow_Success() {
	oauthCfg := oAuthConfig{
		RedirectURIs:            []string{"https://example.com/callback"},
		GrantTypes:              []string{"authorization_code"},
		ResponseTypes:           []string{"code"},
		TokenEndpointAuthMethod: "client_secret_post",
		PKCERequired:            true,
		PublicClient:            false,
	}
	cfgBytes, _ := json.Marshal(oauthCfg)

	row := map[string]interface{}{
		"entity_id":    testEntityID,
		"oauth_config": string(cfgBytes),
	}

	result, err := buildOAuthConfigFromRow(row)

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(testEntityID, result.EntityID)
	suite.NotNil(result.OAuthConfig)
	suite.Equal([]string{"https://example.com/callback"}, result.OAuthConfig.RedirectURIs)
	suite.True(result.OAuthConfig.PKCERequired)
}

func (suite *ApplicationStoreTestSuite) TestBuildOAuthConfigFromRow_InvalidEntityID() {
	row := map[string]interface{}{
		"entity_id":    123, // Invalid type
		"oauth_config": "{}",
	}

	result, err := buildOAuthConfigFromRow(row)

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "failed to parse entity_id as string")
}

func (suite *ApplicationStoreTestSuite) TestBuildOAuthConfigFromRow_NilOAuthConfig() {
	row := map[string]interface{}{
		"entity_id":    testEntityID,
		"oauth_config": nil,
	}

	result, err := buildOAuthConfigFromRow(row)

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(testEntityID, result.EntityID)
	suite.Nil(result.OAuthConfig)
}

func (suite *ApplicationStoreTestSuite) TestBuildOAuthConfigFromRow_MalformedJSON() {
	row := map[string]interface{}{
		"entity_id":    testEntityID,
		"oauth_config": "{invalid json",
	}

	result, err := buildOAuthConfigFromRow(row)

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "failed to unmarshal OAuth config JSON")
}

// --- Tests for parseBoolFromCount ---

func (suite *ApplicationStoreTestSuite) TestParseBoolFromCount() {
	suite.Run("returns true when count is greater than zero", func() {
		results := []map[string]interface{}{
			{
				"count": int64(1),
			},
		}

		result, err := parseBoolFromCount(results)

		suite.NoError(err)
		suite.True(result)
	})

	suite.Run("returns false when count is zero", func() {
		results := []map[string]interface{}{
			{
				"count": int64(0),
			},
		}

		result, err := parseBoolFromCount(results)

		suite.NoError(err)
		suite.False(result)
	})

	suite.Run("returns false when results are empty", func() {
		results := []map[string]interface{}{}

		result, err := parseBoolFromCount(results)

		suite.NoError(err)
		suite.False(result)
	})

	suite.Run("returns error when count is invalid type", func() {
		results := []map[string]interface{}{
			{
				"count": "invalid",
			},
		}

		result, err := parseBoolFromCount(results)

		suite.Error(err)
		suite.False(result)
		suite.Contains(err.Error(), "failed to parse count from query result")
	})
}

// --- Tests for marshalNullableJSON ---

func (suite *ApplicationStoreTestSuite) TestMarshalNullableJSON() {
	suite.Run("returns nil for nil input", func() {
		result, err := marshalNullableJSON(nil)

		suite.NoError(err)
		suite.Nil(result)
	})

	suite.Run("marshals non-nil value", func() {
		result, err := marshalNullableJSON(map[string]string{"key": "value"})

		suite.NoError(err)
		suite.NotNil(result)
	})
}

// --- Tests for parseStringColumn ---

func (suite *ApplicationStoreTestSuite) TestParseStringColumn() {
	suite.Run("returns string value", func() {
		row := map[string]interface{}{"key": "value"}
		result := parseStringColumn(row, "key")
		suite.Equal("value", result)
	})

	suite.Run("returns empty string for nil", func() {
		row := map[string]interface{}{"key": nil}
		result := parseStringColumn(row, "key")
		suite.Equal("", result)
	})

	suite.Run("returns empty string for missing key", func() {
		row := map[string]interface{}{}
		result := parseStringColumn(row, "key")
		suite.Equal("", result)
	})

	suite.Run("returns empty string for non-string type", func() {
		row := map[string]interface{}{"key": 123}
		result := parseStringColumn(row, "key")
		suite.Equal("", result)
	})
}

// --- Tests for parseStringOrBytesColumn ---

func (suite *ApplicationStoreTestSuite) TestParseStringOrBytesColumn() {
	suite.Run("returns string value", func() {
		row := map[string]interface{}{"key": "value"}
		result := parseStringOrBytesColumn(row, "key")
		suite.Equal("value", result)
	})

	suite.Run("returns string from bytes", func() {
		row := map[string]interface{}{"key": []byte("value")}
		result := parseStringOrBytesColumn(row, "key")
		suite.Equal("value", result)
	})

	suite.Run("returns empty string for nil", func() {
		row := map[string]interface{}{"key": nil}
		result := parseStringOrBytesColumn(row, "key")
		suite.Equal("", result)
	})

	suite.Run("returns empty string for other types", func() {
		row := map[string]interface{}{"key": 123}
		result := parseStringOrBytesColumn(row, "key")
		suite.Equal("", result)
	})
}

// --- Tests for parseJSONColumnString ---

func (suite *ApplicationStoreTestSuite) TestParseJSONColumnString() {
	suite.Run("returns string value", func() {
		row := map[string]interface{}{"col": `{"key":"value"}`}
		result := parseJSONColumnString(row, "col")
		suite.Equal(`{"key":"value"}`, result)
	})

	suite.Run("returns string from bytes", func() {
		row := map[string]interface{}{"col": []byte(`{"key":"value"}`)}
		result := parseJSONColumnString(row, "col")
		suite.Equal(`{"key":"value"}`, result)
	})

	suite.Run("returns empty string for nil", func() {
		row := map[string]interface{}{"col": nil}
		result := parseJSONColumnString(row, "col")
		suite.Equal("", result)
	})

	suite.Run("returns empty string for missing key", func() {
		row := map[string]interface{}{}
		result := parseJSONColumnString(row, "col")
		suite.Equal("", result)
	})
}

// --- Tests for IsApplicationDeclarative ---

func (suite *ApplicationStoreTestSuite) TestApplicationStore_IsApplicationDeclarative() {
	suite.Run("returns false for database application", func() {
		result := suite.store.IsApplicationDeclarative(context.Background(), "any-app-id")
		suite.False(result)
	})
}

// --- Tests for IsApplicationExists ---

func (suite *ApplicationStoreTestSuite) TestIsApplicationExists() {
	suite.Run("returns true when application exists", func() {
		suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil).Once()
		suite.mockDBClient.On("QueryContext", mock.Anything, queryCheckApplicationExistsByID,
			"existing-app", testServerID).
			Return([]map[string]interface{}{
				{
					"count": int64(1),
				},
			}, nil).Once()

		exists, err := suite.store.IsApplicationExists(context.Background(), "existing-app")

		suite.NoError(err)
		suite.True(exists)
	})

	suite.Run("returns false when application not found", func() {
		suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil).Once()
		suite.mockDBClient.On("QueryContext", mock.Anything, queryCheckApplicationExistsByID,
			"non-existent-app", testServerID).
			Return([]map[string]interface{}{
				{
					"count": int64(0),
				},
			}, nil).Once()

		exists, err := suite.store.IsApplicationExists(context.Background(), "non-existent-app")

		suite.NoError(err)
		suite.False(exists)
	})

	suite.Run("returns error when database query fails", func() {
		suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil).Once()
		suite.mockDBClient.On("QueryContext", mock.Anything, queryCheckApplicationExistsByID,
			"test-app", testServerID).
			Return(nil, errors.New("database connection error")).Once()

		exists, err := suite.store.IsApplicationExists(context.Background(), "test-app")

		suite.False(exists)
		suite.Error(err)
		suite.Contains(err.Error(), "database connection error")
	})

	suite.Run("returns error when db provider fails", func() {
		suite.mockDBProvider.On("GetConfigDBClient").
			Return(nil, errors.New("db provider unavailable")).Once()

		exists, err := suite.store.IsApplicationExists(context.Background(), "test-app")

		suite.False(exists)
		suite.Error(err)
		suite.Contains(err.Error(), "failed to get database client")
	})
}

// --- Tests for DeleteApplication ---

func (suite *ApplicationStoreTestSuite) runDeleteExecTest(
	fn func(ctx context.Context, id string) error,
	query any, id, execErrMsg string,
) {
	suite.Run("successfully deletes", func() {
		suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil).Once()
		suite.mockDBClient.On("ExecuteContext", mock.Anything, query,
			id, testServerID).Return(int64(1), nil).Once()

		err := fn(context.Background(), id)

		suite.NoError(err)
	})

	suite.Run("returns error when database client fails", func() {
		suite.mockDBProvider.On("GetConfigDBClient").
			Return(nil, errors.New("db provider unavailable")).Once()

		err := fn(context.Background(), id)

		suite.Error(err)
		suite.Contains(err.Error(), "failed to get database client")
	})

	suite.Run("returns error when execute query fails", func() {
		suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil).Once()
		suite.mockDBClient.On("ExecuteContext", mock.Anything, query,
			id, testServerID).Return(int64(0), errors.New("database delete error")).Once()

		err := fn(context.Background(), id)

		suite.Error(err)
		suite.Contains(err.Error(), execErrMsg)
	})
}

func (suite *ApplicationStoreTestSuite) TestDeleteApplication() {
	suite.runDeleteExecTest(
		suite.store.DeleteApplication, queryDeleteApplicationByID,
		"app-to-delete", "failed to delete application",
	)
}

// --- Tests for DeleteOAuthConfig ---

func (suite *ApplicationStoreTestSuite) TestDeleteOAuthConfig() {
	suite.runDeleteExecTest(
		suite.store.DeleteOAuthConfig, queryDeleteOAuthConfigByEntityID,
		testEntityID, "failed to delete OAuth config",
	)
}

// --- Tests for CreateApplication ---

func (suite *ApplicationStoreTestSuite) runAppConfigExecTest(
	fn func(ctx context.Context, app applicationConfigDAO) error,
	query any, execErrMsg string,
) {
	suite.Run("successfully executes", func() {
		app := applicationConfigDAO{
			ID:                        "app1",
			AuthFlowID:                "flow_1",
			RegistrationFlowID:        "reg_flow_1",
			IsRegistrationFlowEnabled: true,
		}
		suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil).Once()
		suite.mockDBClient.On("ExecuteContext", mock.Anything, query,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything).
			Return(int64(1), nil).Once()

		err := fn(context.Background(), app)

		suite.NoError(err)
	})

	suite.Run("returns error when database client fails", func() {
		app := applicationConfigDAO{ID: "app1"}
		suite.mockDBProvider.On("GetConfigDBClient").
			Return(nil, errors.New("db provider unavailable")).Once()

		err := fn(context.Background(), app)

		suite.Error(err)
		suite.Contains(err.Error(), "failed to get database client")
	})

	suite.Run("returns error when execute query fails", func() {
		app := applicationConfigDAO{ID: "app1"}
		suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil).Once()
		suite.mockDBClient.On("ExecuteContext", mock.Anything, query,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything).
			Return(int64(0), errors.New("database error")).Once()

		err := fn(context.Background(), app)

		suite.Error(err)
		suite.Contains(err.Error(), execErrMsg)
	})
}

func (suite *ApplicationStoreTestSuite) runOAuthConfigExecTest(
	fn func(ctx context.Context, entityID string, config json.RawMessage) error,
	query any, execErrMsg string,
) {
	entityID := testEntityID
	config := json.RawMessage(`{}`)

	suite.Run("successfully executes", func() {
		suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil).Once()
		suite.mockDBClient.On("ExecuteContext", mock.Anything, query,
			entityID, mock.Anything, testServerID).Return(int64(1), nil).Once()

		err := fn(context.Background(), entityID, config)

		suite.NoError(err)
	})

	suite.Run("returns error when database client fails", func() {
		suite.mockDBProvider.On("GetConfigDBClient").
			Return(nil, errors.New("db provider unavailable")).Once()

		err := fn(context.Background(), entityID, config)

		suite.Error(err)
		suite.Contains(err.Error(), "failed to get database client")
	})

	suite.Run("returns error when execute query fails", func() {
		suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil).Once()
		suite.mockDBClient.On("ExecuteContext", mock.Anything, query,
			entityID, mock.Anything, testServerID).Return(int64(0), errors.New("query failed")).Once()

		err := fn(context.Background(), entityID, config)

		suite.Error(err)
		suite.Contains(err.Error(), execErrMsg)
	})
}

func (suite *ApplicationStoreTestSuite) TestCreateApplication() {
	suite.runAppConfigExecTest(
		suite.store.CreateApplication, queryCreateApplication, "failed to insert application",
	)
}

// --- Tests for CreateOAuthConfig ---

func (suite *ApplicationStoreTestSuite) TestCreateOAuthConfig() {
	suite.runOAuthConfigExecTest(
		suite.store.CreateOAuthConfig, queryCreateOAuthApplication, "failed to insert OAuth config",
	)
}

// --- Tests for GetTotalApplicationCount ---

func (suite *ApplicationStoreTestSuite) TestGetTotalApplicationCount() {
	suite.Run("returns correct count when applications exist", func() {
		suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil).Once()
		suite.mockDBClient.On("QueryContext", mock.Anything, queryGetApplicationCount, testServerID).
			Return([]map[string]interface{}{
				{
					"total": int64(5),
				},
			}, nil).Once()

		count, err := suite.store.GetTotalApplicationCount(context.Background())

		suite.NoError(err)
		suite.Equal(5, count)
	})

	suite.Run("returns zero when no applications exist", func() {
		suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil).Once()
		suite.mockDBClient.On("QueryContext", mock.Anything, queryGetApplicationCount, testServerID).
			Return([]map[string]interface{}{
				{
					"total": int64(0),
				},
			}, nil).Once()

		count, err := suite.store.GetTotalApplicationCount(context.Background())

		suite.NoError(err)
		suite.Equal(0, count)
	})

	suite.Run("returns zero when query returns empty result", func() {
		suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil).Once()
		suite.mockDBClient.On("QueryContext", mock.Anything, queryGetApplicationCount, testServerID).
			Return([]map[string]interface{}{}, nil).Once()

		count, err := suite.store.GetTotalApplicationCount(context.Background())

		suite.NoError(err)
		suite.Equal(0, count)
	})

	suite.Run("returns error when database client fails", func() {
		suite.mockDBProvider.On("GetConfigDBClient").
			Return(nil, errors.New("db provider unavailable")).Once()

		count, err := suite.store.GetTotalApplicationCount(context.Background())

		suite.Error(err)
		suite.Equal(0, count)
		suite.Contains(err.Error(), "failed to get database client")
	})

	suite.Run("returns error when query fails", func() {
		suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil).Once()
		suite.mockDBClient.On("QueryContext", mock.Anything, queryGetApplicationCount, testServerID).
			Return(nil, errors.New("query execution failed")).Once()

		count, err := suite.store.GetTotalApplicationCount(context.Background())

		suite.Error(err)
		suite.Equal(0, count)
		suite.Contains(err.Error(), "failed to execute query")
	})

	suite.Run("returns error when total count is invalid type", func() {
		suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil).Once()
		suite.mockDBClient.On("QueryContext", mock.Anything, queryGetApplicationCount, testServerID).
			Return([]map[string]interface{}{
				{
					"total": "invalid", // Invalid type
				},
			}, nil).Once()

		count, err := suite.store.GetTotalApplicationCount(context.Background())

		suite.Error(err)
		suite.Equal(0, count)
		suite.Contains(err.Error(), "failed to parse total count")
	})
}

// --- Tests for GetApplicationList ---

func (suite *ApplicationStoreTestSuite) TestGetApplicationList() {
	suite.Run("returns list of applications", func() {
		mockRows := []map[string]interface{}{
			{
				"id":                           "app1",
				"auth_flow_id":                 "auth_flow_1",
				"registration_flow_id":         "reg_flow_1",
				"is_registration_flow_enabled": "1",
				"theme_id":                     nil,
				"layout_id":                    nil,
				"assertion":                    nil,
				"login_consent":                nil,
				"allowed_entity_types":         nil,
				"properties":                   nil,
			},
			{
				"id":                           "app2",
				"auth_flow_id":                 "auth_flow_1",
				"registration_flow_id":         "reg_flow_1",
				"is_registration_flow_enabled": "0",
				"theme_id":                     nil,
				"layout_id":                    nil,
				"assertion":                    nil,
				"login_consent":                nil,
				"allowed_entity_types":         nil,
				"properties":                   nil,
			},
		}

		suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil).Once()
		suite.mockDBClient.On("QueryContext", mock.Anything, queryGetApplicationList, testServerID).
			Return(mockRows, nil).Once()

		apps, err := suite.store.GetApplicationList(context.Background())

		suite.NoError(err)
		suite.Len(apps, 2)
		suite.Equal("app1", apps[0].ID)
		suite.True(apps[0].IsRegistrationFlowEnabled)
		suite.Equal("app2", apps[1].ID)
		suite.False(apps[1].IsRegistrationFlowEnabled)
	})

	suite.Run("returns empty list when no applications exist", func() {
		suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil).Once()
		suite.mockDBClient.On("QueryContext", mock.Anything, queryGetApplicationList, testServerID).
			Return([]map[string]interface{}{}, nil).Once()

		apps, err := suite.store.GetApplicationList(context.Background())

		suite.NoError(err)
		suite.Len(apps, 0)
	})

	suite.Run("returns error when database client fails", func() {
		suite.mockDBProvider.On("GetConfigDBClient").
			Return(nil, errors.New("db provider unavailable")).Once()

		apps, err := suite.store.GetApplicationList(context.Background())

		suite.Error(err)
		suite.Nil(apps)
		suite.Contains(err.Error(), "failed to get database client")
	})

	suite.Run("returns error when query fails", func() {
		suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil).Once()
		suite.mockDBClient.On("QueryContext", mock.Anything, queryGetApplicationList, testServerID).
			Return(nil, errors.New("query execution failed")).Once()

		apps, err := suite.store.GetApplicationList(context.Background())

		suite.Error(err)
		suite.Nil(apps)
		suite.Contains(err.Error(), "failed to execute query")
	})

	suite.Run("returns error when row parsing fails", func() {
		mockRows := []map[string]interface{}{
			{
				"id": 123, // Invalid type - should be string
			},
		}

		suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil).Once()
		suite.mockDBClient.On("QueryContext", mock.Anything, queryGetApplicationList, testServerID).
			Return(mockRows, nil).Once()

		apps, err := suite.store.GetApplicationList(context.Background())

		suite.Error(err)
		suite.Nil(apps)
		suite.Contains(err.Error(), "failed to build application from result row")
	})
}

// --- Tests for GetApplicationByID ---

func (suite *ApplicationStoreTestSuite) TestGetApplicationByID() {
	suite.Run("returns application when found", func() {
		mockRow := map[string]interface{}{
			"id":                           "app1",
			"auth_flow_id":                 "auth_flow_1",
			"registration_flow_id":         "reg_flow_1",
			"is_registration_flow_enabled": "1",
			"theme_id":                     nil,
			"layout_id":                    nil,
			"assertion":                    nil,
			"login_consent":                nil,
			"allowed_entity_types":         nil,
			"properties":                   nil,
		}

		suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil).Once()
		suite.mockDBClient.On("QueryContext", mock.Anything, queryGetApplicationByID, "app1", testServerID).
			Return([]map[string]interface{}{mockRow}, nil).Once()

		app, err := suite.store.GetApplicationByID(context.Background(), "app1")

		suite.NoError(err)
		suite.NotNil(app)
		suite.Equal("app1", app.ID)
	})

	suite.Run("returns error when application not found", func() {
		suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil).Once()
		suite.mockDBClient.On("QueryContext", mock.Anything, queryGetApplicationByID,
			"non-existent", testServerID).Return([]map[string]interface{}{}, nil).Once()

		app, err := suite.store.GetApplicationByID(context.Background(), "non-existent")

		suite.Error(err)
		suite.Nil(app)
		suite.Equal(model.ApplicationNotFoundError, err)
	})

	suite.Run("returns error when database client fails", func() {
		suite.mockDBProvider.On("GetConfigDBClient").
			Return(nil, errors.New("db provider unavailable")).Once()

		app, err := suite.store.GetApplicationByID(context.Background(), "app1")

		suite.Error(err)
		suite.Nil(app)
		suite.Contains(err.Error(), "failed to get database client")
	})

	suite.Run("returns error when query fails", func() {
		suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil).Once()
		suite.mockDBClient.On("QueryContext", mock.Anything, queryGetApplicationByID, "app1", testServerID).
			Return(nil, errors.New("query execution failed")).Once()

		app, err := suite.store.GetApplicationByID(context.Background(), "app1")

		suite.Error(err)
		suite.Nil(app)
		suite.Contains(err.Error(), "failed to execute query")
	})
}

// --- Tests for GetOAuthConfigByEntityID ---

func (suite *ApplicationStoreTestSuite) TestGetOAuthConfigByEntityID() {
	suite.Run("returns OAuth config when found", func() {
		oauthCfg := oAuthConfig{
			RedirectURIs: []string{"https://example.com/callback"},
			GrantTypes:   []string{"authorization_code"},
			PKCERequired: true,
		}
		cfgBytes, _ := json.Marshal(oauthCfg)

		mockRow := map[string]interface{}{
			"entity_id":    testEntityID,
			"oauth_config": string(cfgBytes),
		}

		suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil).Once()
		suite.mockDBClient.On("QueryContext", mock.Anything, queryGetOAuthConfigByEntityID,
			testEntityID, testServerID).
			Return([]map[string]interface{}{mockRow}, nil).Once()

		result, err := suite.store.GetOAuthConfigByEntityID(context.Background(), testEntityID)

		suite.NoError(err)
		suite.NotNil(result)
		suite.Equal(testEntityID, result.EntityID)
		suite.NotNil(result.OAuthConfig)
		suite.True(result.OAuthConfig.PKCERequired)
	})

	suite.Run("returns error when not found", func() {
		suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil).Once()
		suite.mockDBClient.On("QueryContext", mock.Anything, queryGetOAuthConfigByEntityID,
			"non-existent", testServerID).
			Return([]map[string]interface{}{}, nil).Once()

		result, err := suite.store.GetOAuthConfigByEntityID(context.Background(), "non-existent")

		suite.Error(err)
		suite.Nil(result)
		suite.Equal(model.ApplicationNotFoundError, err)
	})

	suite.Run("returns error when database client fails", func() {
		suite.mockDBProvider.On("GetConfigDBClient").
			Return(nil, errors.New("db provider unavailable")).Once()

		result, err := suite.store.GetOAuthConfigByEntityID(context.Background(), testEntityID)

		suite.Error(err)
		suite.Nil(result)
		suite.Contains(err.Error(), "failed to get database client")
	})
}

// --- Tests for UpdateApplication ---

func (suite *ApplicationStoreTestSuite) TestUpdateApplication() {
	suite.runAppConfigExecTest(
		suite.store.UpdateApplication, queryUpdateApplicationByID, "failed to update application",
	)
}

// --- Tests for UpdateOAuthConfig ---

func (suite *ApplicationStoreTestSuite) TestUpdateOAuthConfig() {
	suite.runOAuthConfigExecTest(
		suite.store.UpdateOAuthConfig, queryUpdateOAuthConfigByEntityID, "failed to update OAuth config",
	)
}
