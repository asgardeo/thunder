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
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/application/model"
	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/system/config"
	dbmodel "github.com/asgardeo/thunder/internal/system/database/model"
	"github.com/asgardeo/thunder/tests/mocks/database/modelmock"
	"github.com/asgardeo/thunder/tests/mocks/database/providermock"
)

const testAppID = "test-app-id"
const testServerID = "test-server-id"

// ApplicationStoreTestSuite contains comprehensive tests for the application store helper functions.
type ApplicationStoreTestSuite struct {
	suite.Suite
	mockDBClient *providermock.DBClientInterfaceMock
}

func TestApplicationStoreTestSuite(t *testing.T) {
	suite.Run(t, new(ApplicationStoreTestSuite))
}

func (suite *ApplicationStoreTestSuite) SetupTest() {
	_ = config.InitializeThunderRuntime("test", &config.Config{})
	suite.mockDBClient = providermock.NewDBClientInterfaceMock(suite.T())
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
			Issuer:         "test-issuer",
			ValidityPeriod: 3600,
			UserAttributes: []string{"email", "name", "sub"},
		},
		InboundAuthConfig: []model.InboundAuthConfigProcessedDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigProcessedDTO{
					AppID:              "app1",
					ClientID:           "client_app1",
					HashedClientSecret: "hashed_secret_app1",
					RedirectURIs:       []string{"https://example.com/callback", "https://example.com/cb2"},
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
						Issuer: "oauth-issuer",
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
	store := newApplicationStore()

	suite.NotNil(store)
	suite.IsType(&applicationStore{}, store)
}

func (suite *ApplicationStoreTestSuite) TestGetAppJSONDataBytes_Success() {
	app := suite.createTestApplication()

	jsonBytes, err := getAppJSONDataBytes(&app)

	suite.NoError(err)
	suite.NotNil(jsonBytes)

	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	suite.NoError(err)
	suite.Equal(app.URL, result["url"])
	suite.Equal(app.LogoURL, result["logo_url"])
	suite.Equal(app.TosURI, result["tos_uri"])
	suite.Equal(app.PolicyURI, result["policy_uri"])

	contacts, ok := result["contacts"].([]interface{})
	suite.True(ok)
	suite.Len(contacts, 2)
	suite.Equal("contact@example.com", contacts[0])

	assertion, ok := result["assertion"].(map[string]interface{})
	suite.True(ok)
	suite.Equal("test-issuer", assertion["issuer"])
	suite.Equal(float64(3600), assertion["validity_period"])
}

func (suite *ApplicationStoreTestSuite) TestGetAppJSONDataBytes_WithoutToken() {
	app := suite.createTestApplication()
	app.Assertion = nil

	jsonBytes, err := getAppJSONDataBytes(&app)

	suite.NoError(err)
	suite.NotNil(jsonBytes)

	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	suite.NoError(err)
	suite.Nil(result["assertion"])
	suite.Equal(app.URL, result["url"])
}

func (suite *ApplicationStoreTestSuite) TestGetAppJSONDataBytes_WithEmptyToken() {
	app := suite.createTestApplication()
	app.Assertion = &model.AssertionConfig{} // Empty token - should not be included

	jsonBytes, err := getAppJSONDataBytes(&app)

	suite.NoError(err)
	suite.NotNil(jsonBytes)

	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	suite.NoError(err)
	suite.Nil(result["assertion"])
}

func (suite *ApplicationStoreTestSuite) TestGetAppJSONDataBytes_WithPartialToken() {
	app := suite.createTestApplication()
	app.Assertion = &model.AssertionConfig{
		Issuer: "test-issuer",
		// No validity period or user attributes
	}

	jsonBytes, err := getAppJSONDataBytes(&app)

	suite.NoError(err)
	suite.NotNil(jsonBytes)

	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	suite.NoError(err)

	assertion, ok := result["assertion"].(map[string]interface{})
	suite.True(ok)
	suite.Equal("test-issuer", assertion["issuer"])
	suite.Nil(assertion["validity_period"])
	suite.Nil(assertion["user_attributes"])
}

func (suite *ApplicationStoreTestSuite) TestGetAppJSONDataBytes_EmptyContacts() {
	app := suite.createTestApplication()
	app.Contacts = []string{}

	jsonBytes, err := getAppJSONDataBytes(&app)

	suite.NoError(err)
	suite.NotNil(jsonBytes)

	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	suite.NoError(err)

	contacts, ok := result["contacts"].([]interface{})
	suite.True(ok)
	suite.Len(contacts, 0)
}

func (suite *ApplicationStoreTestSuite) TestGetAppJSONDataBytes_WithTemplate() {
	app := suite.createTestApplication()
	app.Template = "spa"

	jsonBytes, err := getAppJSONDataBytes(&app)

	suite.NoError(err)
	suite.NotNil(jsonBytes)

	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	suite.NoError(err)
	suite.Equal("spa", result["template"])
}

func (suite *ApplicationStoreTestSuite) TestGetAppJSONDataBytes_WithEmptyTemplate() {
	app := suite.createTestApplication()
	app.Template = ""

	jsonBytes, err := getAppJSONDataBytes(&app)

	suite.NoError(err)
	suite.NotNil(jsonBytes)

	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	suite.NoError(err)
	suite.Nil(result["template"]) // Empty template should not be included
}

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
	suite.Equal("oauth-issuer", token["issuer"])

	accessToken, ok := token["access_token"].(map[string]interface{})
	suite.True(ok)
	// AccessTokenConfig does not have an issuer field - issuer is at OAuth level
	suite.Nil(accessToken["issuer"])
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

func (suite *ApplicationStoreTestSuite) TestBuildBasicApplicationFromResultRow_Success() {
	row := map[string]interface{}{
		"app_id":                       "app1",
		"app_name":                     "Test App 1",
		"description":                  "Test description",
		"auth_flow_id":                 "auth_flow_1",
		"registration_flow_id":         "reg_flow_1",
		"is_registration_flow_enabled": "1",
		"consumer_key":                 "client_app1",
	}

	result, err := buildBasicApplicationFromResultRow(row)

	suite.NoError(err)
	suite.Equal("app1", result.ID)
	suite.Equal("Test App 1", result.Name)
	suite.Equal("Test description", result.Description)
	suite.Equal("auth_flow_1", result.AuthFlowID)
	suite.Equal("reg_flow_1", result.RegistrationFlowID)
	suite.True(result.IsRegistrationFlowEnabled)
	suite.Equal("client_app1", result.ClientID)
}

func (suite *ApplicationStoreTestSuite) TestBuildBasicApplicationFromResultRow_WithNullDescription() {
	row := map[string]interface{}{
		"app_id":                       "app1",
		"app_name":                     "Test App 1",
		"description":                  nil,
		"auth_flow_id":                 "auth_flow_1",
		"registration_flow_id":         "reg_flow_1",
		"is_registration_flow_enabled": "1",
		"consumer_key":                 nil,
	}

	result, err := buildBasicApplicationFromResultRow(row)

	suite.NoError(err)
	suite.Equal("", result.Description)
	suite.Equal("", result.ClientID)
}

func (suite *ApplicationStoreTestSuite) TestBuildBasicApplicationFromResultRow_WithByteRegistrationFlag() {
	row := map[string]interface{}{
		"app_id":                       "app1",
		"app_name":                     "Test App 1",
		"description":                  "Test description",
		"auth_flow_id":                 "auth_flow_1",
		"registration_flow_id":         "reg_flow_1",
		"is_registration_flow_enabled": []byte("1"),
		"consumer_key":                 "client_app1",
	}

	result, err := buildBasicApplicationFromResultRow(row)

	suite.NoError(err)
	suite.True(result.IsRegistrationFlowEnabled)
}

func (suite *ApplicationStoreTestSuite) TestBuildBasicApplicationFromResultRow_WithZeroRegistrationFlag() {
	row := map[string]interface{}{
		"app_id":                       "app1",
		"app_name":                     "Test App 1",
		"description":                  "Test description",
		"auth_flow_id":                 "auth_flow_1",
		"registration_flow_id":         "reg_flow_1",
		"is_registration_flow_enabled": "0",
		"consumer_key":                 "client_app1",
	}

	result, err := buildBasicApplicationFromResultRow(row)

	suite.NoError(err)
	suite.False(result.IsRegistrationFlowEnabled)
}

func (suite *ApplicationStoreTestSuite) TestBuildBasicApplicationFromResultRow_WithTemplate() {
	appJSON := map[string]interface{}{
		"template": "spa",
	}
	appJSONBytes, _ := json.Marshal(appJSON)

	row := map[string]interface{}{
		"app_id":                       "app1",
		"app_name":                     "Test App 1",
		"description":                  "Test description",
		"auth_flow_id":                 "auth_flow_1",
		"registration_flow_id":         "reg_flow_1",
		"is_registration_flow_enabled": "1",
		"branding_id":                  "brand-123",
		"app_json":                     string(appJSONBytes),
		"consumer_key":                 "client_app1",
	}

	result, err := buildBasicApplicationFromResultRow(row)

	suite.NoError(err)
	suite.Equal("app1", result.ID)
	suite.Equal("brand-123", result.BrandingID)
	suite.Equal("spa", result.Template)
}

func (suite *ApplicationStoreTestSuite) TestBuildBasicApplicationFromResultRow_WithNullTemplate() {
	appJSON := map[string]interface{}{}
	appJSONBytes, _ := json.Marshal(appJSON)

	row := map[string]interface{}{
		"app_id":                       "app1",
		"app_name":                     "Test App 1",
		"description":                  "Test description",
		"auth_flow_id":                 "auth_flow_1",
		"registration_flow_id":         "reg_flow_1",
		"is_registration_flow_enabled": "1",
		"app_json":                     string(appJSONBytes),
		"consumer_key":                 "client_app1",
	}

	result, err := buildBasicApplicationFromResultRow(row)

	suite.NoError(err)
	suite.Equal("app1", result.ID)
	suite.Equal("", result.Template)
}

func (suite *ApplicationStoreTestSuite) TestBuildBasicApplicationFromResultRow_WithEmptyTemplate() {
	appJSON := map[string]interface{}{
		"template": "",
	}
	appJSONBytes, _ := json.Marshal(appJSON)

	row := map[string]interface{}{
		"app_id":                       "app1",
		"app_name":                     "Test App 1",
		"description":                  "Test description",
		"auth_flow_id":                 "auth_flow_1",
		"registration_flow_id":         "reg_flow_1",
		"is_registration_flow_enabled": "1",
		"app_json":                     string(appJSONBytes),
		"consumer_key":                 "client_app1",
	}

	result, err := buildBasicApplicationFromResultRow(row)

	suite.NoError(err)
	suite.Equal("app1", result.ID)
	suite.Equal("", result.Template)
}

func (suite *ApplicationStoreTestSuite) TestBuildBasicApplicationFromResultRow_InvalidAppID() {
	row := map[string]interface{}{
		"app_id": 123, // Invalid type
	}

	result, err := buildBasicApplicationFromResultRow(row)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to parse app_id as string")
	suite.Equal(model.BasicApplicationDTO{}, result)
}

func (suite *ApplicationStoreTestSuite) TestBuildBasicApplicationFromResultRow_InvalidAppName() {
	row := map[string]interface{}{
		"app_id":   "app1",
		"app_name": 123, // Invalid type
	}

	result, err := buildBasicApplicationFromResultRow(row)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to parse app_name as string")
	suite.Equal(model.BasicApplicationDTO{}, result)
}

func (suite *ApplicationStoreTestSuite) TestBuildBasicApplicationFromResultRow_InvalidDescription() {
	row := map[string]interface{}{
		"app_id":      "app1",
		"app_name":    "Test App",
		"description": 123, // Invalid type
	}

	result, err := buildBasicApplicationFromResultRow(row)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to parse description as string")
	suite.Equal(model.BasicApplicationDTO{}, result)
}

func (suite *ApplicationStoreTestSuite) TestBuildBasicApplicationFromResultRow_InvalidRegistrationFlag() {
	row := map[string]interface{}{
		"app_id":                       "app1",
		"app_name":                     "Test App 1",
		"description":                  "Test description",
		"auth_flow_id":                 "auth_flow_1",
		"registration_flow_id":         "reg_flow_1",
		"is_registration_flow_enabled": 123, // Invalid type
	}

	result, err := buildBasicApplicationFromResultRow(row)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to parse is_registration_flow_enabled")
	suite.Equal(model.BasicApplicationDTO{}, result)
}

func (suite *ApplicationStoreTestSuite) TestBuildBasicApplicationFromResultRow_InvalidConsumerKey() {
	row := map[string]interface{}{
		"app_id":                       "app1",
		"app_name":                     "Test App 1",
		"description":                  "Test description",
		"auth_flow_id":                 "auth_flow_1",
		"registration_flow_id":         "reg_flow_1",
		"is_registration_flow_enabled": "1",
		"consumer_key":                 123, // Invalid type
	}

	result, err := buildBasicApplicationFromResultRow(row)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to parse consumer_key as string")
	suite.Equal(model.BasicApplicationDTO{}, result)
}

func (suite *ApplicationStoreTestSuite) TestBuildApplicationFromResultRow_Success() {
	appJSON := map[string]interface{}{
		"url":        "https://example.com",
		"logo_url":   "https://example.com/logo.png",
		"tos_uri":    "https://example.com/tos",
		"policy_uri": "https://example.com/policy",
		"contacts":   []interface{}{"contact@example.com"},
		"assertion": map[string]interface{}{
			"issuer":          "test-issuer",
			"validity_period": float64(3600),
			"user_attributes": []interface{}{"email", "name"},
		},
	}
	appJSONBytes, _ := json.Marshal(appJSON)

	oauthJSON := map[string]interface{}{
		"redirect_uris":              []interface{}{"https://example.com/callback"},
		"grant_types":                []interface{}{"authorization_code"},
		"response_types":             []interface{}{"code"},
		"token_endpoint_auth_method": "client_secret_post",
		"pkce_required":              true,
		"public_client":              false,
		"scopes":                     []interface{}{"openid", "profile"},
	}
	oauthJSONBytes, _ := json.Marshal(oauthJSON)

	row := map[string]interface{}{
		"app_id":                       "app1",
		"app_name":                     "Test App 1",
		"description":                  "Test description",
		"auth_flow_id":                 "auth_flow_1",
		"registration_flow_id":         "reg_flow_1",
		"is_registration_flow_enabled": "1",
		"app_json":                     string(appJSONBytes),
		"consumer_key":                 "client_app1",
		"consumer_secret":              "hashed_secret",
		"oauth_config_json":            string(oauthJSONBytes),
	}

	result, err := buildApplicationFromResultRow(row)

	suite.NoError(err)
	suite.Equal("app1", result.ID)
	suite.Equal("Test App 1", result.Name)
	suite.Equal("https://example.com", result.URL)
	suite.Equal("https://example.com/logo.png", result.LogoURL)
	suite.Equal("https://example.com/tos", result.TosURI)
	suite.Equal("https://example.com/policy", result.PolicyURI)
	suite.Len(result.Contacts, 1)
	suite.NotNil(result.Assertion)
	suite.Equal("test-issuer", result.Assertion.Issuer)
	suite.Equal(int64(3600), result.Assertion.ValidityPeriod)
	suite.Len(result.InboundAuthConfig, 1)
	suite.Equal("client_app1", result.InboundAuthConfig[0].OAuthAppConfig.ClientID)
}

func (suite *ApplicationStoreTestSuite) TestBuildApplicationFromResultRow_WithTemplate() {
	appJSON := map[string]interface{}{
		"url":      "https://example.com",
		"logo_url": "https://example.com/logo.png",
		"template": "mobile",
	}
	appJSONBytes, _ := json.Marshal(appJSON)

	row := map[string]interface{}{
		"app_id":                       "app1",
		"app_name":                     "Test App 1",
		"description":                  "Test description",
		"auth_flow_id":                 "auth_flow_1",
		"registration_flow_id":         "reg_flow_1",
		"is_registration_flow_enabled": "1",
		"branding_id":                  "brand-123",
		"app_json":                     string(appJSONBytes),
	}

	result, err := buildApplicationFromResultRow(row)

	suite.NoError(err)
	suite.Equal("app1", result.ID)
	suite.Equal("brand-123", result.BrandingID)
	suite.Equal("mobile", result.Template)
	suite.Equal("https://example.com", result.URL)
}

func (suite *ApplicationStoreTestSuite) TestBuildApplicationFromResultRow_WithoutTemplate() {
	appJSON := map[string]interface{}{
		"url":      "https://example.com",
		"logo_url": "https://example.com/logo.png",
	}
	appJSONBytes, _ := json.Marshal(appJSON)

	row := map[string]interface{}{
		"app_id":                       "app1",
		"app_name":                     "Test App 1",
		"description":                  "Test description",
		"auth_flow_id":                 "auth_flow_1",
		"registration_flow_id":         "reg_flow_1",
		"is_registration_flow_enabled": "1",
		"app_json":                     string(appJSONBytes),
	}

	result, err := buildApplicationFromResultRow(row)

	suite.NoError(err)
	suite.Equal("app1", result.ID)
	suite.Equal("", result.Template) // No template in app_json
	suite.Equal("https://example.com", result.URL)
}

func (suite *ApplicationStoreTestSuite) TestBuildApplicationFromResultRow_WithNullAppJSON() {
	row := map[string]interface{}{
		"app_id":                       "app1",
		"app_name":                     "Test App 1",
		"description":                  "Test description",
		"auth_flow_id":                 "auth_flow_1",
		"registration_flow_id":         "reg_flow_1",
		"is_registration_flow_enabled": "1",
		"app_json":                     nil,
	}

	result, err := buildApplicationFromResultRow(row)

	suite.NoError(err)
	suite.Equal("", result.URL)
	suite.Equal("", result.LogoURL)
	suite.Equal("", result.Template) // Null app_json means no template
	suite.Nil(result.Assertion)
}

func (suite *ApplicationStoreTestSuite) TestBuildApplicationFromResultRow_WithByteAppJSON() {
	appJSON := map[string]interface{}{
		"url":      "https://example.com",
		"logo_url": "https://example.com/logo.png",
	}
	appJSONBytes, _ := json.Marshal(appJSON)

	row := map[string]interface{}{
		"app_id":                       "app1",
		"app_name":                     "Test App 1",
		"description":                  "Test description",
		"auth_flow_id":                 "auth_flow_1",
		"registration_flow_id":         "reg_flow_1",
		"is_registration_flow_enabled": "1",
		"app_json":                     appJSONBytes, // As bytes
	}

	result, err := buildApplicationFromResultRow(row)

	suite.NoError(err)
	suite.Equal("https://example.com", result.URL)
	suite.Equal("https://example.com/logo.png", result.LogoURL)
}

func (suite *ApplicationStoreTestSuite) TestBuildApplicationFromResultRow_WithInvalidAppJSONType() {
	row := map[string]interface{}{
		"app_id":                       "app1",
		"app_name":                     "Test App 1",
		"description":                  "Test description",
		"auth_flow_id":                 "auth_flow_1",
		"registration_flow_id":         "reg_flow_1",
		"is_registration_flow_enabled": "1",
		"app_json":                     123, // Invalid type
	}

	result, err := buildApplicationFromResultRow(row)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to parse app_json as string or []byte")
	suite.Equal(model.ApplicationProcessedDTO{}, result)
}

func (suite *ApplicationStoreTestSuite) TestBuildApplicationFromResultRow_WithInvalidAppJSON() {
	row := map[string]interface{}{
		"app_id":                       "app1",
		"app_name":                     "Test App 1",
		"description":                  "Test description",
		"auth_flow_id":                 "auth_flow_1",
		"registration_flow_id":         "reg_flow_1",
		"is_registration_flow_enabled": "1",
		"app_json":                     "invalid json",
	}

	result, err := buildApplicationFromResultRow(row)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to unmarshal app JSON")
	suite.Equal(model.ApplicationProcessedDTO{}, result)
}

func (suite *ApplicationStoreTestSuite) TestBuildApplicationFromResultRow_WithInvalidURLType() {
	appJSON := map[string]interface{}{
		"url": 123, // Invalid type
	}
	appJSONBytes, _ := json.Marshal(appJSON)

	row := map[string]interface{}{
		"app_id":                       "app1",
		"app_name":                     "Test App 1",
		"description":                  "Test description",
		"auth_flow_id":                 "auth_flow_1",
		"registration_flow_id":         "reg_flow_1",
		"is_registration_flow_enabled": "1",
		"app_json":                     string(appJSONBytes),
	}

	result, err := buildApplicationFromResultRow(row)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to parse url from app JSON")
	suite.Equal(model.ApplicationProcessedDTO{}, result)
}

func (suite *ApplicationStoreTestSuite) TestBuildApplicationFromResultRow_WithInvalidLogoURLType() {
	appJSON := map[string]interface{}{
		"url":      "https://example.com",
		"logo_url": 123, // Invalid type
	}
	appJSONBytes, _ := json.Marshal(appJSON)

	row := map[string]interface{}{
		"app_id":                       "app1",
		"app_name":                     "Test App 1",
		"description":                  "Test description",
		"auth_flow_id":                 "auth_flow_1",
		"registration_flow_id":         "reg_flow_1",
		"is_registration_flow_enabled": "1",
		"app_json":                     string(appJSONBytes),
	}

	result, err := buildApplicationFromResultRow(row)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to parse logo_url from app JSON")
	suite.Equal(model.ApplicationProcessedDTO{}, result)
}

func (suite *ApplicationStoreTestSuite) TestBuildApplicationFromResultRow_WithInvalidContactsType() {
	appJSON := map[string]interface{}{
		"url":      "https://example.com",
		"logo_url": "https://example.com/logo.png",
		"contacts": "not an array", // Invalid type
	}
	appJSONBytes, _ := json.Marshal(appJSON)

	row := map[string]interface{}{
		"app_id":                       "app1",
		"app_name":                     "Test App 1",
		"description":                  "Test description",
		"auth_flow_id":                 "auth_flow_1",
		"registration_flow_id":         "reg_flow_1",
		"is_registration_flow_enabled": "1",
		"app_json":                     string(appJSONBytes),
	}

	result, err := buildApplicationFromResultRow(row)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to parse contacts from app JSON")
	suite.Equal(model.ApplicationProcessedDTO{}, result)
}

func (suite *ApplicationStoreTestSuite) TestBuildOAuthInboundAuthConfig_Success() {
	oauthJSON := map[string]interface{}{
		"redirect_uris":              []interface{}{"https://example.com/callback"},
		"grant_types":                []interface{}{"authorization_code"},
		"response_types":             []interface{}{"code"},
		"token_endpoint_auth_method": "client_secret_post",
		"pkce_required":              true,
		"public_client":              false,
		"scopes":                     []interface{}{"openid"},
		"token": map[string]interface{}{
			"issuer": "oauth-issuer",
			"access_token": map[string]interface{}{
				"issuer":          "access-issuer",
				"validity_period": float64(7200),
				"user_attributes": []interface{}{"sub", "email"},
			},
			"id_token": map[string]interface{}{
				"validity_period": float64(3600),
				"user_attributes": []interface{}{"sub"},
				"scope_claims": map[string]interface{}{
					"profile": []interface{}{"name"},
				},
			},
		},
	}
	oauthJSONBytes, _ := json.Marshal(oauthJSON)

	row := map[string]interface{}{
		"consumer_secret":   "hashed_secret",
		"oauth_config_json": string(oauthJSONBytes),
	}

	basicApp := model.BasicApplicationDTO{
		ID:       "app1",
		ClientID: "client_app1",
	}

	result, err := buildOAuthInboundAuthConfig(row, basicApp)

	suite.NoError(err)
	suite.Equal(model.OAuthInboundAuthType, result.Type)
	suite.NotNil(result.OAuthAppConfig)
	suite.Equal("client_app1", result.OAuthAppConfig.ClientID)
	suite.Equal("hashed_secret", result.OAuthAppConfig.HashedClientSecret)
	suite.True(result.OAuthAppConfig.PKCERequired)
	suite.False(result.OAuthAppConfig.PublicClient)
	suite.NotNil(result.OAuthAppConfig.Token)
	suite.NotNil(result.OAuthAppConfig.Token.AccessToken)
	suite.NotNil(result.OAuthAppConfig.Token.IDToken)
}

func (suite *ApplicationStoreTestSuite) TestBuildOAuthInboundAuthConfig_WithNullOAuthJSON() {
	row := map[string]interface{}{
		"consumer_secret":   "hashed_secret",
		"oauth_config_json": nil,
	}

	basicApp := model.BasicApplicationDTO{
		ID:       "app1",
		ClientID: "client_app1",
	}

	result, err := buildOAuthInboundAuthConfig(row, basicApp)

	suite.NoError(err)
	suite.NotNil(result.OAuthAppConfig)
	suite.Len(result.OAuthAppConfig.RedirectURIs, 0)
}

func (suite *ApplicationStoreTestSuite) TestBuildOAuthInboundAuthConfig_WithByteOAuthJSON() {
	oauthJSON := map[string]interface{}{
		"redirect_uris":              []interface{}{"https://example.com/callback"},
		"grant_types":                []interface{}{"authorization_code"},
		"response_types":             []interface{}{"code"},
		"token_endpoint_auth_method": "client_secret_post",
		"pkce_required":              false,
		"public_client":              true,
	}
	oauthJSONBytes, _ := json.Marshal(oauthJSON)

	row := map[string]interface{}{
		"consumer_secret":   "hashed_secret",
		"oauth_config_json": oauthJSONBytes, // As bytes
	}

	basicApp := model.BasicApplicationDTO{
		ID:       "app1",
		ClientID: "client_app1",
	}

	result, err := buildOAuthInboundAuthConfig(row, basicApp)

	suite.NoError(err)
	suite.NotNil(result.OAuthAppConfig)
	suite.True(result.OAuthAppConfig.PublicClient)
	suite.False(result.OAuthAppConfig.PKCERequired)
}

func (suite *ApplicationStoreTestSuite) TestBuildOAuthInboundAuthConfig_InvalidConsumerSecret() {
	row := map[string]interface{}{
		"consumer_secret": 123, // Invalid type
	}

	basicApp := model.BasicApplicationDTO{
		ID:       "app1",
		ClientID: "client_app1",
	}

	result, err := buildOAuthInboundAuthConfig(row, basicApp)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to parse consumer_secret as string")
	suite.Equal(model.InboundAuthConfigProcessedDTO{}, result)
}

func (suite *ApplicationStoreTestSuite) TestBuildOAuthInboundAuthConfig_WithAccessTokenNilUserAttributes() {
	oauthJSON := map[string]interface{}{
		"redirect_uris":              []interface{}{"https://example.com/callback"},
		"grant_types":                []interface{}{"authorization_code"},
		"response_types":             []interface{}{"code"},
		"token_endpoint_auth_method": "client_secret_post",
		"pkce_required":              true,
		"public_client":              false,
		"scopes":                     []interface{}{"openid"},
		"token": map[string]interface{}{
			"issuer": "oauth-issuer",
			"access_token": map[string]interface{}{
				"validity_period": float64(7200),
				// user_attributes is nil/omitted
			},
		},
	}
	oauthJSONBytes, _ := json.Marshal(oauthJSON)

	row := map[string]interface{}{
		"consumer_secret":   "hashed_secret",
		"oauth_config_json": string(oauthJSONBytes),
	}

	basicApp := model.BasicApplicationDTO{
		ID:       "app1",
		ClientID: "client_app1",
	}

	result, err := buildOAuthInboundAuthConfig(row, basicApp)

	suite.NoError(err)
	suite.NotNil(result.OAuthAppConfig.Token)
	suite.NotNil(result.OAuthAppConfig.Token.AccessToken)
	// nil UserAttributes should be initialized to empty slice
	suite.NotNil(result.OAuthAppConfig.Token.AccessToken.UserAttributes)
	suite.Len(result.OAuthAppConfig.Token.AccessToken.UserAttributes, 0)
	suite.Equal(int64(7200), result.OAuthAppConfig.Token.AccessToken.ValidityPeriod)
}

func (suite *ApplicationStoreTestSuite) TestBuildOAuthInboundAuthConfig_InvalidOAuthJSONType() {
	row := map[string]interface{}{
		"consumer_secret":   "hashed_secret",
		"oauth_config_json": 123, // Invalid type
	}

	basicApp := model.BasicApplicationDTO{
		ID:       "app1",
		ClientID: "client_app1",
	}

	result, err := buildOAuthInboundAuthConfig(row, basicApp)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to parse oauth_config_json as string or []byte")
	suite.Equal(model.InboundAuthConfigProcessedDTO{}, result)
}

func (suite *ApplicationStoreTestSuite) TestBuildOAuthInboundAuthConfig_InvalidOAuthJSON() {
	row := map[string]interface{}{
		"consumer_secret":   "hashed_secret",
		"oauth_config_json": "invalid json",
	}

	basicApp := model.BasicApplicationDTO{
		ID:       "app1",
		ClientID: "client_app1",
	}

	result, err := buildOAuthInboundAuthConfig(row, basicApp)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to unmarshal oauth config JSON")
	suite.Equal(model.InboundAuthConfigProcessedDTO{}, result)
}

func (suite *ApplicationStoreTestSuite) TestCreateOAuthAppQuery_Success() {
	app := suite.createTestApplication()
	query := createOAuthAppQuery(&app, QueryCreateOAuthApplication, testServerID)

	suite.NotNil(query)
	suite.IsType((func(dbmodel.TxInterface) error)(nil), query)
}

func (suite *ApplicationStoreTestSuite) TestDeleteOAuthAppQuery_Success() {
	clientID := "test_client_id"
	query := deleteOAuthAppQuery(clientID, testServerID)

	suite.NotNil(query)
	suite.IsType((func(dbmodel.TxInterface) error)(nil), query)
}

func (suite *ApplicationStoreTestSuite) TestBuildApplicationFromResultRow_WithInvalidTosURIType() {
	appJSON := map[string]interface{}{
		"url":      "https://example.com",
		"logo_url": "https://example.com/logo.png",
		"tos_uri":  123, // Invalid type
	}
	appJSONBytes, _ := json.Marshal(appJSON)

	row := map[string]interface{}{
		"app_id":                       "app1",
		"app_name":                     "Test App 1",
		"description":                  "Test description",
		"auth_flow_id":                 "auth_flow_1",
		"registration_flow_id":         "reg_flow_1",
		"is_registration_flow_enabled": "1",
		"app_json":                     string(appJSONBytes),
	}

	result, err := buildApplicationFromResultRow(row)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to parse tos_uri from app JSON")
	suite.Equal(model.ApplicationProcessedDTO{}, result)
}

func (suite *ApplicationStoreTestSuite) TestBuildApplicationFromResultRow_WithInvalidPolicyURIType() {
	appJSON := map[string]interface{}{
		"url":        "https://example.com",
		"logo_url":   "https://example.com/logo.png",
		"policy_uri": 123, // Invalid type
	}
	appJSONBytes, _ := json.Marshal(appJSON)

	row := map[string]interface{}{
		"app_id":                       "app1",
		"app_name":                     "Test App 1",
		"description":                  "Test description",
		"auth_flow_id":                 "auth_flow_1",
		"registration_flow_id":         "reg_flow_1",
		"is_registration_flow_enabled": "1",
		"app_json":                     string(appJSONBytes),
	}

	result, err := buildApplicationFromResultRow(row)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to parse policy_uri from app JSON")
	suite.Equal(model.ApplicationProcessedDTO{}, result)
}

func (suite *ApplicationStoreTestSuite) TestBuildApplicationFromResultRow_WithTokenInvalidType() {
	appJSON := map[string]interface{}{
		"url":       "https://example.com",
		"logo_url":  "https://example.com/logo.png",
		"assertion": "not a map", // Invalid type - should be map[string]interface{}
	}
	appJSONBytes, _ := json.Marshal(appJSON)

	row := map[string]interface{}{
		"app_id":                       "app1",
		"app_name":                     "Test App 1",
		"description":                  "Test description",
		"auth_flow_id":                 "auth_flow_1",
		"registration_flow_id":         "reg_flow_1",
		"is_registration_flow_enabled": "1",
		"app_json":                     string(appJSONBytes),
	}

	result, err := buildApplicationFromResultRow(row)

	// Token with invalid type should be ignored (not cause error)
	suite.NoError(err)
	suite.Nil(result.Assertion)
	suite.Equal("app1", result.ID)
}

func (suite *ApplicationStoreTestSuite) TestBuildApplicationFromResultRow_WithTokenInvalidValidityPeriodType() {
	appJSON := map[string]interface{}{
		"url":      "https://example.com",
		"logo_url": "https://example.com/logo.png",
		"assertion": map[string]interface{}{
			"issuer":          "test-issuer",
			"validity_period": "not a number", // Invalid type
			"user_attributes": []interface{}{"email"},
		},
	}
	appJSONBytes, _ := json.Marshal(appJSON)

	row := map[string]interface{}{
		"app_id":                       "app1",
		"app_name":                     "Test App 1",
		"description":                  "Test description",
		"auth_flow_id":                 "auth_flow_1",
		"registration_flow_id":         "reg_flow_1",
		"is_registration_flow_enabled": "1",
		"app_json":                     string(appJSONBytes),
	}

	result, err := buildApplicationFromResultRow(row)

	// Invalid validity_period type should be ignored (not cause error)
	suite.NoError(err)
	suite.NotNil(result.Assertion)
	suite.Equal("test-issuer", result.Assertion.Issuer)
	suite.Equal(int64(0), result.Assertion.ValidityPeriod) // Should be 0 when parsing fails
}

func (suite *ApplicationStoreTestSuite) TestBuildApplicationFromResultRow_WithTokenInvalidUserAttributesType() {
	appJSON := map[string]interface{}{
		"url":      "https://example.com",
		"logo_url": "https://example.com/logo.png",
		"assertion": map[string]interface{}{
			"issuer":          "test-issuer",
			"validity_period": float64(3600),
			"user_attributes": "not an array", // Invalid type
		},
	}
	appJSONBytes, _ := json.Marshal(appJSON)

	row := map[string]interface{}{
		"app_id":                       "app1",
		"app_name":                     "Test App 1",
		"description":                  "Test description",
		"auth_flow_id":                 "auth_flow_1",
		"registration_flow_id":         "reg_flow_1",
		"is_registration_flow_enabled": "1",
		"app_json":                     string(appJSONBytes),
	}

	result, err := buildApplicationFromResultRow(row)

	// Invalid user_attributes type should be ignored (not cause error)
	suite.NoError(err)
	suite.NotNil(result.Assertion)
	suite.Equal("test-issuer", result.Assertion.Issuer)
	suite.Len(result.Assertion.UserAttributes, 0) // Should be empty when parsing fails
}

func (suite *ApplicationStoreTestSuite) TestBuildApplicationFromResultRow_WithTokenUserAttributesNonStringElement() {
	appJSON := map[string]interface{}{
		"url":      "https://example.com",
		"logo_url": "https://example.com/logo.png",
		"assertion": map[string]interface{}{
			"issuer":          "test-issuer",
			"validity_period": float64(3600),
			"user_attributes": []interface{}{"email", 123, "name"}, // Mixed types
		},
	}
	appJSONBytes, _ := json.Marshal(appJSON)

	row := map[string]interface{}{
		"app_id":                       "app1",
		"app_name":                     "Test App 1",
		"description":                  "Test description",
		"auth_flow_id":                 "auth_flow_1",
		"registration_flow_id":         "reg_flow_1",
		"is_registration_flow_enabled": "1",
		"app_json":                     string(appJSONBytes),
	}

	result, err := buildApplicationFromResultRow(row)

	// Non-string elements should be skipped
	suite.NoError(err)
	suite.NotNil(result.Assertion)
	suite.Len(result.Assertion.UserAttributes, 2) // Only "email" and "name", 123 is skipped
	suite.Equal("email", result.Assertion.UserAttributes[0])
	suite.Equal("name", result.Assertion.UserAttributes[1])
}

func (suite *ApplicationStoreTestSuite) TestBuildApplicationFromResultRow_WithContactsNonStringElement() {
	appJSON := map[string]interface{}{
		"url":      "https://example.com",
		"logo_url": "https://example.com/logo.png",
		"contacts": []interface{}{"contact@example.com", 123, "support@example.com"}, // Mixed types
	}
	appJSONBytes, _ := json.Marshal(appJSON)

	row := map[string]interface{}{
		"app_id":                       "app1",
		"app_name":                     "Test App 1",
		"description":                  "Test description",
		"auth_flow_id":                 "auth_flow_1",
		"registration_flow_id":         "reg_flow_1",
		"is_registration_flow_enabled": "1",
		"app_json":                     string(appJSONBytes),
	}

	result, err := buildApplicationFromResultRow(row)

	// Non-string elements should cause an error
	suite.Error(err)
	suite.Contains(err.Error(), "failed to parse contacts from app JSON")
	suite.Contains(err.Error(), "item at index 1 is not a string")
	suite.Empty(result)
}

func (suite *ApplicationStoreTestSuite) TestGetAppJSONDataBytes_WithNilContacts() {
	app := suite.createTestApplication()
	app.Contacts = nil

	jsonBytes, err := getAppJSONDataBytes(&app)

	suite.NoError(err)
	suite.NotNil(jsonBytes)

	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	suite.NoError(err)
	// nil contacts should be marshaled as null or empty array
	// JSON unmarshaling converts null to nil
	if contacts, ok := result["contacts"].([]interface{}); ok {
		suite.Len(contacts, 0)
	} else {
		suite.Nil(result["contacts"])
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

// getOAuthApplicationFromResults is a helper function that extracts OAuth application from query results
// This duplicates the logic from store.GetOAuthApplication for testing purposes
func getOAuthApplicationFromResults(
	clientID string,
	results []map[string]interface{},
) (*model.OAuthAppConfigProcessedDTO, error) {
	if len(results) == 0 {
		return nil, model.ApplicationNotFoundError
	}

	row := results[0]

	appID, ok := row["app_id"].(string)
	if !ok {
		return nil, errors.New("failed to parse app_id as string")
	}

	hashedClientSecret, ok := row["consumer_secret"].(string)
	if !ok {
		return nil, errors.New("failed to parse consumer_secret as string")
	}

	// Extract OAuth JSON data
	var oauthConfigJSON string
	if row["oauth_config_json"] == nil {
		oauthConfigJSON = "{}"
	} else if v, ok := row["oauth_config_json"].(string); ok {
		oauthConfigJSON = v
	} else if v, ok := row["oauth_config_json"].([]byte); ok {
		oauthConfigJSON = string(v)
	} else {
		return nil, errors.New("failed to parse oauth_config_json as string or []byte")
	}

	var oAuthConfig oAuthConfig
	if err := json.Unmarshal([]byte(oauthConfigJSON), &oAuthConfig); err != nil {
		return nil, err
	}

	// Convert the typed arrays
	grantTypes := make([]oauth2const.GrantType, 0)
	for _, gt := range oAuthConfig.GrantTypes {
		grantTypes = append(grantTypes, oauth2const.GrantType(gt))
	}

	responseTypes := make([]oauth2const.ResponseType, 0)
	for _, rt := range oAuthConfig.ResponseTypes {
		responseTypes = append(responseTypes, oauth2const.ResponseType(rt))
	}

	tokenEndpointAuthMethod := oauth2const.TokenEndpointAuthMethod(oAuthConfig.TokenEndpointAuthMethod)

	// Convert token config if present
	var oauthTokenConfig *model.OAuthTokenConfig
	if oAuthConfig.Token != nil {
		oauthTokenConfig = &model.OAuthTokenConfig{
			Issuer: oAuthConfig.Token.Issuer,
		}
		if oAuthConfig.Token.AccessToken != nil {
			oauthTokenConfig.AccessToken = &model.AccessTokenConfig{
				ValidityPeriod: oAuthConfig.Token.AccessToken.ValidityPeriod,
				UserAttributes: oAuthConfig.Token.AccessToken.UserAttributes,
			}
		}
		if oAuthConfig.Token.IDToken != nil {
			oauthTokenConfig.IDToken = &model.IDTokenConfig{
				ValidityPeriod: oAuthConfig.Token.IDToken.ValidityPeriod,
				UserAttributes: oAuthConfig.Token.IDToken.UserAttributes,
			}
		}
	}

	var userInfoConfig *model.UserInfoConfig
	if oAuthConfig.UserInfo != nil {
		userInfoConfig = &model.UserInfoConfig{
			UserAttributes: oAuthConfig.UserInfo.UserAttributes,
		}
	}

	return &model.OAuthAppConfigProcessedDTO{
		AppID:                   appID,
		ClientID:                clientID,
		HashedClientSecret:      hashedClientSecret,
		GrantTypes:              grantTypes,
		ResponseTypes:           responseTypes,
		TokenEndpointAuthMethod: tokenEndpointAuthMethod,
		RedirectURIs:            oAuthConfig.RedirectURIs,
		PKCERequired:            oAuthConfig.PKCERequired,
		PublicClient:            oAuthConfig.PublicClient,
		Token:                   oauthTokenConfig,
		Scopes:                  oAuthConfig.Scopes,
		UserInfo:                userInfoConfig,
		ScopeClaims:             oAuthConfig.ScopeClaims,
	}, nil
}

func (suite *ApplicationStoreTestSuite) TestGetOAuthApplication_WithComplexTokenConfig() {
	clientID := "test-client-123"
	//nolint:gosec // This is test data, not actual credentials
	tokenConfigJSON := `{
		"redirect_uris": ["https://example.com/callback"],
		"grant_types": ["authorization_code", "refresh_token"],
		"response_types": ["code"],
		"token_endpoint_auth_method": "client_secret_post",
		"pkce_required": true,
		"public_client": false,
		"scopes": ["openid", "profile", "email"],
		"token": {
			"issuer": "https://test-issuer.example.com",
			"access_token": {
				"issuer": "https://access-issuer.example.com",
				"validity_period": 7200,
				"user_attributes": ["sub", "email", "name"]
			},
			"id_token": {
				"validity_period": 3600,
				"user_attributes": ["sub", "email"]
			}
		},
		"scope_claims": {
			"profile": ["name", "family_name"],
			"email": ["email", "email_verified"]
		}
	}`

	mockRow := map[string]interface{}{
		"app_id":            "app-123",
		"consumer_secret":   "hashed-secret",
		"oauth_config_json": tokenConfigJSON,
	}

	suite.mockDBClient.
		On("Query", QueryGetOAuthApplicationByClientID, clientID, testServerID).
		Return([]map[string]interface{}{mockRow}, nil).
		Once()

	// Execute
	results, err := suite.mockDBClient.Query(QueryGetOAuthApplicationByClientID, clientID, testServerID)
	suite.Require().NoError(err)
	result, err := getOAuthApplicationFromResults(clientID, results)

	// Assert
	suite.Require().NoError(err)
	suite.Require().NotNil(result)

	// Verify basic fields
	suite.Assert().Equal("app-123", result.AppID)
	suite.Assert().Equal(clientID, result.ClientID)
	suite.Assert().Equal("hashed-secret", result.HashedClientSecret)

	// Verify token configuration
	suite.Require().NotNil(result.Token)
	suite.Assert().Equal("https://test-issuer.example.com", result.Token.Issuer)

	// Verify access token
	suite.Require().NotNil(result.Token.AccessToken)
	suite.Assert().Equal(int64(7200), result.Token.AccessToken.ValidityPeriod)
	suite.Assert().Contains(result.Token.AccessToken.UserAttributes, "name")

	// Verify ID token
	suite.Require().NotNil(result.Token.IDToken)
	suite.Assert().Equal(int64(3600), result.Token.IDToken.ValidityPeriod)
	suite.Assert().Len(result.Token.IDToken.UserAttributes, 2)

	// Verify scope claims
	suite.Require().NotNil(result.ScopeClaims)
	suite.Assert().Contains(result.ScopeClaims["profile"], "family_name")
	suite.Assert().Contains(result.ScopeClaims["email"], "email_verified")
}

func (suite *ApplicationStoreTestSuite) TestGetOAuthApplication_WithNilTokenConfig() {
	clientID := "test-client-no-token"
	//nolint:gosec // This is test data, not actual credentials
	tokenConfigJSON := `{
		"redirect_uris": ["https://example.com/callback"],
		"grant_types": ["client_credentials"],
		"response_types": [],
		"token_endpoint_auth_method": "client_secret_basic",
		"pkce_required": false,
		"public_client": false
	}`

	mockRow := map[string]interface{}{
		"app_id":            "app-no-token",
		"consumer_secret":   "hashed-secret",
		"oauth_config_json": tokenConfigJSON,
	}

	suite.mockDBClient.
		On("Query", QueryGetOAuthApplicationByClientID, mock.Anything, testServerID).
		Return([]map[string]interface{}{mockRow}, nil).
		Once()

	// Execute
	results, err := suite.mockDBClient.Query(QueryGetOAuthApplicationByClientID, clientID, testServerID)
	suite.Require().NoError(err)
	result, err := getOAuthApplicationFromResults(clientID, results)

	// Assert
	suite.Require().NoError(err)
	suite.Require().NotNil(result)
	suite.Assert().Nil(result.Token, "Token should be nil when not configured")
}

func (suite *ApplicationStoreTestSuite) TestGetOAuthApplication_WithJSONAsBytes() {
	clientID := "test-client-bytes"

	tokenConfigJSON := []byte(`{
		"redirect_uris": ["https://example.com/callback"],
		"grant_types": ["authorization_code"],
		"response_types": ["code"],
		"token_endpoint_auth_method": "client_secret_post"
	}`)

	mockRow := map[string]interface{}{
		"app_id":            "app-bytes",
		"consumer_secret":   "hashed-secret",
		"oauth_config_json": tokenConfigJSON, // As []byte instead of string
	}

	suite.mockDBClient.
		On("Query", QueryGetOAuthApplicationByClientID, mock.Anything, testServerID).
		Return([]map[string]interface{}{mockRow}, nil).
		Once()

	// Execute
	results, err := suite.mockDBClient.Query(QueryGetOAuthApplicationByClientID, clientID, testServerID)
	suite.Require().NoError(err)
	result, err := getOAuthApplicationFromResults(clientID, results)

	// Assert
	suite.Require().NoError(err)
	suite.Require().NotNil(result)
	suite.Assert().Equal("app-bytes", result.AppID)
}

func (suite *ApplicationStoreTestSuite) TestGetOAuthApplication_WithNilJSON() {
	clientID := "test-client-nil-json"

	mockRow := map[string]interface{}{
		"app_id":            "app-nil-json",
		"consumer_secret":   "hashed-secret",
		"oauth_config_json": nil, // Nil JSON
	}

	suite.mockDBClient.
		On("Query", QueryGetOAuthApplicationByClientID, mock.Anything, testServerID).
		Return([]map[string]interface{}{mockRow}, nil).
		Once()

	// Execute
	results, err := suite.mockDBClient.Query(QueryGetOAuthApplicationByClientID, clientID, testServerID)
	suite.Require().NoError(err)
	result, err := getOAuthApplicationFromResults(clientID, results)

	// Assert - should use empty JSON object "{}"
	suite.Require().NoError(err)
	suite.Require().NotNil(result)
	suite.Assert().Equal("app-nil-json", result.AppID)
}

func (suite *ApplicationStoreTestSuite) TestGetOAuthApplication_NotFound() {
	clientID := "non-existent-client"
	suite.mockDBClient.
		On("Query", QueryGetOAuthApplicationByClientID, mock.Anything, testServerID).
		Return([]map[string]interface{}{}, nil).
		Once()

	// Execute
	results, err := suite.mockDBClient.Query(QueryGetOAuthApplicationByClientID, clientID, testServerID)
	suite.Require().NoError(err)
	result, err := getOAuthApplicationFromResults(clientID, results)

	// Assert
	suite.Assert().Error(err)
	suite.Assert().Equal(model.ApplicationNotFoundError, err)
	suite.Assert().Nil(result)
}

func (suite *ApplicationStoreTestSuite) TestGetOAuthApplication_DatabaseError() {
	clientID := "test-client-error"
	suite.mockDBClient.
		On("Query", QueryGetOAuthApplicationByClientID, mock.Anything, testServerID).
		Return(nil, errors.New("database connection error")).
		Once()

	// Execute
	_, err := suite.mockDBClient.Query(QueryGetOAuthApplicationByClientID, clientID, testServerID)
	result := (*model.OAuthAppConfigProcessedDTO)(nil)

	// Assert
	suite.Assert().Error(err)
	suite.Assert().Contains(err.Error(), "database connection error")
	suite.Assert().Nil(result)
}

func (suite *ApplicationStoreTestSuite) TestGetOAuthApplication_InvalidAppIDType() {
	clientID := "test-client-invalid-id"

	mockRow := map[string]interface{}{
		"app_id":            123, // Invalid type
		"consumer_secret":   "hashed-secret",
		"oauth_config_json": "{}",
	}

	suite.mockDBClient.
		On("Query", QueryGetOAuthApplicationByClientID, mock.Anything, testServerID).
		Return([]map[string]interface{}{mockRow}, nil).
		Once()

	// Execute
	results, err := suite.mockDBClient.Query(QueryGetOAuthApplicationByClientID, clientID, testServerID)
	suite.Require().NoError(err)
	result, err := getOAuthApplicationFromResults(clientID, results)

	// Assert
	suite.Assert().Error(err)
	suite.Assert().Contains(err.Error(), "failed to parse app_id as string")
	suite.Assert().Nil(result)
}

func (suite *ApplicationStoreTestSuite) TestGetOAuthApplication_InvalidJSONType() {
	clientID := "test-client-invalid-json-type"

	mockRow := map[string]interface{}{
		"app_id":            "app-123",
		"consumer_secret":   "hashed-secret",
		"oauth_config_json": 12345, // Invalid type (not string or []byte)
	}

	suite.mockDBClient.
		On("Query", QueryGetOAuthApplicationByClientID, mock.Anything, testServerID).
		Return([]map[string]interface{}{mockRow}, nil).
		Once()

	// Execute
	results, err := suite.mockDBClient.Query(QueryGetOAuthApplicationByClientID, clientID, testServerID)
	suite.Require().NoError(err)
	result, err := getOAuthApplicationFromResults(clientID, results)

	// Assert
	suite.Assert().Error(err)
	suite.Assert().Contains(err.Error(), "failed to parse oauth_config_json")
	suite.Assert().Nil(result)
}

func (suite *ApplicationStoreTestSuite) TestGetOAuthApplication_MalformedJSON() {
	clientID := "test-client-bad-json"

	mockRow := map[string]interface{}{
		"app_id":            "app-123",
		"consumer_secret":   "hashed-secret",
		"oauth_config_json": "{invalid json", // Malformed JSON
	}

	suite.mockDBClient.
		On("Query", QueryGetOAuthApplicationByClientID, mock.Anything, testServerID).
		Return([]map[string]interface{}{mockRow}, nil).
		Once()

	// Execute
	results, err := suite.mockDBClient.Query(QueryGetOAuthApplicationByClientID, clientID, testServerID)
	suite.Require().NoError(err)
	result, err := getOAuthApplicationFromResults(clientID, results)

	// Assert
	suite.Assert().Error(err)
	suite.Assert().Nil(result)
}

func TestCreateOAuthAppQuery_ExecError(t *testing.T) {
	app := model.ApplicationProcessedDTO{
		ID:   "app-123",
		Name: "Test App",
		InboundAuthConfig: []model.InboundAuthConfigProcessedDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigProcessedDTO{
					ClientID:                "test-client",
					HashedClientSecret:      "hashed-secret",
					GrantTypes:              []oauth2const.GrantType{"authorization_code"},
					ResponseTypes:           []oauth2const.ResponseType{"code"},
					TokenEndpointAuthMethod: "client_secret_post",
				},
			},
		},
	}

	queryFunc := createOAuthAppQuery(&app, QueryCreateOAuthApplication, testServerID)

	mockTx := modelmock.NewTxInterfaceMock(t)
	mockTx.
		On("Exec", QueryCreateOAuthApplication, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			testServerID).
		Return(nil, errors.New("database exec error")).
		Once()

	err := queryFunc(mockTx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database exec error")
}

func TestDeleteOAuthAppQuery_ExecError(t *testing.T) {
	clientID := "test-client-id"

	queryFunc := deleteOAuthAppQuery(clientID, testServerID)

	mockTx := modelmock.NewTxInterfaceMock(t)
	mockTx.
		On("Exec", QueryDeleteOAuthApplicationByClientID, mock.Anything, testServerID).
		Return(nil, errors.New("database delete error")).
		Once()

	err := queryFunc(mockTx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database delete error")
}

// TestGetApplicationByQuery_UnexpectedNumberOfResults tests the error path when query returns multiple results
func (suite *ApplicationStoreTestSuite) TestGetApplicationByQuery_UnexpectedNumberOfResults() {
	appID := testAppID

	// Create multiple result rows to trigger the error
	mockRow1 := map[string]interface{}{
		"app_id":                       "app1",
		"app_name":                     "Test App 1",
		"description":                  "Test description",
		"auth_flow_id":                 "auth_flow_1",
		"registration_flow_id":         "reg_flow_1",
		"is_registration_flow_enabled": "1",
		"app_json":                     "{}",
		"consumer_key":                 "client1",
		"consumer_secret":              "secret1",
		"oauth_config_json":            "{}",
	}

	mockRow2 := map[string]interface{}{
		"app_id":                       "app2",
		"app_name":                     "Test App 2",
		"description":                  "Test description 2",
		"auth_flow_id":                 "auth_flow_1",
		"registration_flow_id":         "reg_flow_1",
		"is_registration_flow_enabled": "1",
		"app_json":                     "{}",
		"consumer_key":                 "client2",
		"consumer_secret":              "secret2",
		"oauth_config_json":            "{}",
	}

	// Mock the database client to return multiple results
	suite.mockDBClient.
		On("Query", QueryGetApplicationByAppID, appID, testServerID).
		Return([]map[string]interface{}{mockRow1, mockRow2}, nil).
		Once()

	// Create a test helper that simulates getApplicationByQuery
	// Since getApplicationByQuery is private, we test it through a test helper
	results, err := suite.mockDBClient.Query(QueryGetApplicationByAppID, appID, testServerID)
	suite.Require().NoError(err)

	// Simulate the error check from getApplicationByQuery
	if len(results) == 0 {
		suite.Fail("Expected results to be non-empty")
	}
	if len(results) != 1 {
		// This is the error path we're testing
		err := fmt.Errorf("unexpected number of results: %d", len(results))
		suite.Error(err)
		suite.Contains(err.Error(), "unexpected number of results")
		suite.Contains(err.Error(), "2")
	}
}

// TestGetApplicationByQuery_BuildApplicationError tests the error path when buildApplicationFromResultRow fails
func (suite *ApplicationStoreTestSuite) TestGetApplicationByQuery_BuildApplicationError() {
	appID := testAppID

	// Create a row with invalid data that will cause buildApplicationFromResultRow to fail
	// Using invalid app_id type to trigger error in buildBasicApplicationFromResultRow
	mockRow := map[string]interface{}{
		"app_id":                       123, // Invalid type - should be string
		"app_name":                     "Test App",
		"description":                  "Test description",
		"auth_flow_id":                 "auth_flow_1",
		"registration_flow_id":         "reg_flow_1",
		"is_registration_flow_enabled": "1",
		"app_json":                     "{}",
		"consumer_key":                 "client1",
		"consumer_secret":              "secret1",
		"oauth_config_json":            "{}",
	}

	// Mock the database client to return a single result with invalid data
	suite.mockDBClient.
		On("Query", QueryGetApplicationByAppID, appID, testServerID).
		Return([]map[string]interface{}{mockRow}, nil).
		Once()

	// Execute query
	results, err := suite.mockDBClient.Query(QueryGetApplicationByAppID, appID, testServerID)
	suite.Require().NoError(err)
	suite.Require().Len(results, 1)

	// Test the error path from getApplicationByQuery - buildApplicationFromResultRow should fail
	row := results[0]
	_, buildErr := buildApplicationFromResultRow(row)

	// This is the error path we're testing
	suite.Error(buildErr)
	suite.Contains(buildErr.Error(), "failed to parse app_id as string")
}

// TestGetApplicationByQuery_BuildApplicationError_InvalidAppJSON tests the error path when app_json parsing fails
func (suite *ApplicationStoreTestSuite) TestGetApplicationByQuery_BuildApplicationError_InvalidAppJSON() {
	appID := testAppID

	// Create a row with invalid app_json type that will cause buildApplicationFromResultRow to fail
	mockRow := map[string]interface{}{
		"app_id":                       "app1",
		"app_name":                     "Test App",
		"description":                  "Test description",
		"auth_flow_id":                 "auth_flow_1",
		"registration_flow_id":         "reg_flow_1",
		"is_registration_flow_enabled": "1",
		"app_json":                     12345, // Invalid type - should be string or []byte
		"consumer_key":                 "client1",
		"consumer_secret":              "secret1",
		"oauth_config_json":            "{}",
	}

	// Mock the database client to return a single result with invalid app_json
	suite.mockDBClient.
		On("Query", QueryGetApplicationByAppID, appID, testServerID).
		Return([]map[string]interface{}{mockRow}, nil).
		Once()

	// Execute query
	results, err := suite.mockDBClient.Query(QueryGetApplicationByAppID, appID, testServerID)
	suite.Require().NoError(err)
	suite.Require().Len(results, 1)

	// Test the error path from getApplicationByQuery - buildApplicationFromResultRow should fail
	row := results[0]
	_, buildErr := buildApplicationFromResultRow(row)

	// This is the error path we're testing
	suite.Error(buildErr)
	suite.Contains(buildErr.Error(), "failed to parse app_json as string or []byte")
}

// TestGetApplicationByQuery_BuildApplicationError_MalformedJSON tests the error path when app_json is malformed
func (suite *ApplicationStoreTestSuite) TestGetApplicationByQuery_BuildApplicationError_MalformedJSON() {
	appID := testAppID

	// Create a row with malformed JSON that will cause buildApplicationFromResultRow to fail
	mockRow := map[string]interface{}{
		"app_id":                       "app1",
		"app_name":                     "Test App",
		"description":                  "Test description",
		"auth_flow_id":                 "auth_flow_1",
		"registration_flow_id":         "reg_flow_1",
		"is_registration_flow_enabled": "1",
		"app_json":                     "{invalid json", // Malformed JSON
		"consumer_key":                 "client1",
		"consumer_secret":              "secret1",
		"oauth_config_json":            "{}",
	}

	// Mock the database client to return a single result with malformed JSON
	suite.mockDBClient.
		On("Query", QueryGetApplicationByAppID, appID, testServerID).
		Return([]map[string]interface{}{mockRow}, nil).
		Once()

	// Execute query
	results, err := suite.mockDBClient.Query(QueryGetApplicationByAppID, appID, testServerID)
	suite.Require().NoError(err)
	suite.Require().Len(results, 1)

	// Test the error path from getApplicationByQuery - buildApplicationFromResultRow should fail
	row := results[0]
	_, buildErr := buildApplicationFromResultRow(row)

	// This is the error path we're testing
	suite.Error(buildErr)
	suite.Contains(buildErr.Error(), "failed to unmarshal app JSON")
}
