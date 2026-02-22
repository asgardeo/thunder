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

	"github.com/asgardeo/thunder/internal/application/model"
	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/system/config"
	dbmodel "github.com/asgardeo/thunder/internal/system/database/model"
	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
)

// oAuthConfig is the structure for unmarshaling OAuth configuration JSON.
type oAuthConfig struct {
	RedirectURIs            []string            `json:"redirect_uris"`
	GrantTypes              []string            `json:"grant_types"`
	ResponseTypes           []string            `json:"response_types"`
	TokenEndpointAuthMethod string              `json:"token_endpoint_auth_method"`
	PKCERequired            bool                `json:"pkce_required"`
	PublicClient            bool                `json:"public_client"`
	Token                   *oAuthTokenConfig   `json:"token,omitempty"`
	Scopes                  []string            `json:"scopes,omitempty"`
	UserInfo                *userInfoConfig     `json:"user_info,omitempty"`
	ScopeClaims             map[string][]string `json:"scope_claims,omitempty"`
}

// oAuthTokenConfig represents the OAuth token configuration structure for JSON marshaling/unmarshaling.
type oAuthTokenConfig struct {
	AccessToken *accessTokenConfig `json:"access_token,omitempty"`
	IDToken     *idTokenConfig     `json:"id_token,omitempty"`
}

// accessTokenConfig represents the access token configuration structure for JSON marshaling/unmarshaling.
type accessTokenConfig struct {
	ValidityPeriod int64    `json:"validity_period,omitempty"`
	UserAttributes []string `json:"user_attributes,omitempty"`
}

// idTokenConfig represents the ID token configuration structure for JSON marshaling/unmarshaling.
type idTokenConfig struct {
	ValidityPeriod int64    `json:"validity_period,omitempty"`
	UserAttributes []string `json:"user_attributes,omitempty"`
}

// userInfoConfig represents the user info endpoint configuration structure for JSON marshaling/unmarshaling.
type userInfoConfig struct {
	ResponseType   model.UserInfoResponseType `json:"response_type,omitempty"`
	UserAttributes []string                   `json:"user_attributes,omitempty"`
}

// ApplicationStoreInterface defines the interface for application data persistence operations.
type applicationStoreInterface interface {
	CreateApplication(app model.ApplicationProcessedDTO) error
	GetTotalApplicationCount() (int, error)
	GetApplicationList() ([]model.BasicApplicationDTO, error)
	GetOAuthApplication(clientID string) (*model.OAuthAppConfigProcessedDTO, error)
	GetApplicationByID(id string) (*model.ApplicationProcessedDTO, error)
	GetApplicationByName(name string) (*model.ApplicationProcessedDTO, error)
	UpdateApplication(existingApp, updatedApp *model.ApplicationProcessedDTO) error
	DeleteApplication(id string) error
	IsApplicationExists(id string) (bool, error)
	IsApplicationExistsByName(name string) (bool, error)
	IsApplicationDeclarative(id string) bool
}

// applicationStore implements the applicationStoreInterface for handling application data persistence.
type applicationStore struct {
	dbProvider   provider.DBProviderInterface
	deploymentID string
}

// NewApplicationStore creates a new instance of applicationStore.
func newApplicationStore() applicationStoreInterface {
	return &applicationStore{
		dbProvider:   provider.GetDBProvider(),
		deploymentID: config.GetThunderRuntime().Config.Server.Identifier,
	}
}

// CreateApplication creates a new application in the database.
func (st *applicationStore) CreateApplication(app model.ApplicationProcessedDTO) error {
	jsonDataBytes, err := getAppJSONDataBytes(&app)
	if err != nil {
		return err
	}

	queries := []func(tx dbmodel.TxInterface) error{
		func(tx dbmodel.TxInterface) error {
			isRegistrationEnabledStr := utils.BoolToNumString(app.IsRegistrationFlowEnabled)
			var themeID interface{}
			if app.ThemeID != "" {
				themeID = app.ThemeID
			} else {
				themeID = nil
			}
			var layoutID interface{}
			if app.LayoutID != "" {
				layoutID = app.LayoutID
			} else {
				layoutID = nil
			}
			_, err := tx.Exec(queryCreateApplication, app.ID, app.Name, app.Description,
				app.AuthFlowID, app.RegistrationFlowID, isRegistrationEnabledStr, themeID, layoutID,
				jsonDataBytes, st.deploymentID)
			return err
		},
	}
	// TODO: Need to refactor when supporting other/multiple inbound auth types.
	if len(app.InboundAuthConfig) > 0 {
		queries = append(queries, createOAuthAppQuery(&app, queryCreateOAuthApplication, st.deploymentID))
	}

	return executeTransaction(st.dbProvider, queries)
}

// GetTotalApplicationCount retrieves the total count of applications from the database.
func (st *applicationStore) GetTotalApplicationCount() (int, error) {
	dbClient, err := st.dbProvider.GetConfigDBClient()
	if err != nil {
		return 0, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(queryGetApplicationCount, st.deploymentID)
	if err != nil {
		return 0, fmt.Errorf("failed to execute query: %w", err)
	}

	totalCount := 0
	if len(results) > 0 {
		if total, ok := results[0]["total"].(int64); ok {
			totalCount = int(total)
		} else {
			return 0, fmt.Errorf("failed to parse total count from query result")
		}
	}

	return totalCount, nil
}

// GetApplicationList retrieves a list of applications from the database.
func (st *applicationStore) GetApplicationList() ([]model.BasicApplicationDTO, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationPersistence"))

	dbClient, err := st.dbProvider.GetConfigDBClient()
	if err != nil {
		logger.Error("Failed to get database client", log.Error(err))
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(queryGetApplicationList, st.deploymentID)
	if err != nil {
		logger.Error("Failed to execute query", log.Error(err))
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	applications := make([]model.BasicApplicationDTO, 0)

	for _, row := range results {
		application, err := buildBasicApplicationFromResultRow(row)
		if err != nil {
			logger.Error("failed to build application from result row", log.Error(err))
			return nil, fmt.Errorf("failed to build application from result row: %w", err)
		}
		applications = append(applications, application)
	}

	return applications, nil
}

// GetOAuthApplication retrieves an OAuth application by its client ID.
func (st *applicationStore) GetOAuthApplication(clientID string) (*model.OAuthAppConfigProcessedDTO, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationStore"))

	dbClient, err := st.dbProvider.GetConfigDBClient()
	if err != nil {
		logger.Error("Failed to get database client", log.Error(err))
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(queryGetOAuthApplicationByClientID, clientID, st.deploymentID)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, model.ApplicationNotFoundError
	}

	row := results[0]

	appID, ok := row["app_id"].(string)
	if !ok {
		return nil, errors.New("failed to parse app_id as string")
	}

	hashedClientSecret, ok := row["client_secret"].(string)
	if !ok {
		return nil, errors.New("failed to parse client_secret as string")
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
		return nil, fmt.Errorf("failed to parse oauth_config_json as string or []byte")
	}

	var oAuthConfig oAuthConfig
	if err := json.Unmarshal([]byte(oauthConfigJSON), &oAuthConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal oauth config JSON: %w", err)
	}

	// Convert the typed arrays to the required types
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
		oauthTokenConfig = &model.OAuthTokenConfig{}
		if oAuthConfig.Token.AccessToken != nil {
			userAttributes := oAuthConfig.Token.AccessToken.UserAttributes
			if userAttributes == nil {
				userAttributes = make([]string, 0)
			}
			oauthTokenConfig.AccessToken = &model.AccessTokenConfig{
				ValidityPeriod: oAuthConfig.Token.AccessToken.ValidityPeriod,
				UserAttributes: userAttributes,
			}
		}
		if oAuthConfig.Token.IDToken != nil {
			userAttributes := oAuthConfig.Token.IDToken.UserAttributes
			if userAttributes == nil {
				userAttributes = make([]string, 0)
			}
			oauthTokenConfig.IDToken = &model.IDTokenConfig{
				ValidityPeriod: oAuthConfig.Token.IDToken.ValidityPeriod,
				UserAttributes: userAttributes,
			}
		}
	}

	// Handle UserInfo config
	var userInfoConfig *model.UserInfoConfig
	if oAuthConfig.UserInfo != nil {
		userAttributes := oAuthConfig.UserInfo.UserAttributes
		if userAttributes == nil {
			userAttributes = make([]string, 0)
		}
		userInfoConfig = &model.UserInfoConfig{
			ResponseType:   oAuthConfig.UserInfo.ResponseType,
			UserAttributes: userAttributes,
		}
	}

	// Handle ScopeClaims config
	scopeClaims := oAuthConfig.ScopeClaims
	if scopeClaims == nil {
		scopeClaims = make(map[string][]string)
	}

	return &model.OAuthAppConfigProcessedDTO{
		AppID:                   appID,
		ClientID:                clientID,
		HashedClientSecret:      hashedClientSecret,
		RedirectURIs:            oAuthConfig.RedirectURIs,
		GrantTypes:              grantTypes,
		ResponseTypes:           responseTypes,
		TokenEndpointAuthMethod: tokenEndpointAuthMethod,
		PKCERequired:            oAuthConfig.PKCERequired,
		PublicClient:            oAuthConfig.PublicClient,
		Token:                   oauthTokenConfig,
		Scopes:                  oAuthConfig.Scopes,
		UserInfo:                userInfoConfig,
		ScopeClaims:             scopeClaims,
	}, nil
}

// GetApplicationByID retrieves a specific application by its ID from the database.
func (st *applicationStore) GetApplicationByID(id string) (*model.ApplicationProcessedDTO, error) {
	return st.getApplicationByQuery(queryGetApplicationByAppID, id, st.deploymentID)
}

// GetApplicationByName retrieves a specific application by its name from the database.
func (st *applicationStore) GetApplicationByName(name string) (*model.ApplicationProcessedDTO, error) {
	return st.getApplicationByQuery(queryGetApplicationByName, name, st.deploymentID)
}

// getApplicationByQuery retrieves a specific application from the database using the provided query and parameter.
func (st *applicationStore) getApplicationByQuery(query dbmodel.DBQuery, params ...interface{}) (
	*model.ApplicationProcessedDTO, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationStore"))

	dbClient, err := st.dbProvider.GetConfigDBClient()
	if err != nil {
		logger.Error("Failed to get database client", log.Error(err))
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(query, params...)
	if err != nil {
		logger.Error("Failed to execute query", log.Error(err))
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	if len(results) == 0 {
		return nil, model.ApplicationNotFoundError
	}
	if len(results) != 1 {
		logger.Error("unexpected number of results")
		return nil, fmt.Errorf("unexpected number of results: %d", len(results))
	}

	row := results[0]
	application, err := buildApplicationFromResultRow(row)
	if err != nil {
		logger.Error("failed to build application from result row", log.Error(err))
		return nil, fmt.Errorf("failed to build application from result row: %w", err)
	}

	return &application, nil
}

// UpdateApplication updates an existing application in the database.
func (st *applicationStore) UpdateApplication(existingApp, updatedApp *model.ApplicationProcessedDTO) error {
	jsonDataBytes, err := getAppJSONDataBytes(updatedApp)
	if err != nil {
		return err
	}

	queries := []func(tx dbmodel.TxInterface) error{
		func(tx dbmodel.TxInterface) error {
			isRegistrationEnabledStr := utils.BoolToNumString(updatedApp.IsRegistrationFlowEnabled)
			var themeID interface{}
			if updatedApp.ThemeID != "" {
				themeID = updatedApp.ThemeID
			} else {
				themeID = nil
			}
			var layoutID interface{}
			if updatedApp.LayoutID != "" {
				layoutID = updatedApp.LayoutID
			} else {
				layoutID = nil
			}
			_, err := tx.Exec(queryUpdateApplicationByAppID, updatedApp.ID, updatedApp.Name,
				updatedApp.Description, updatedApp.AuthFlowID, updatedApp.RegistrationFlowID,
				isRegistrationEnabledStr, themeID, layoutID, jsonDataBytes, st.deploymentID)
			return err
		},
	}
	// TODO: Need to refactor when supporting other/multiple inbound auth types.
	if len(updatedApp.InboundAuthConfig) > 0 && len(existingApp.InboundAuthConfig) > 0 {
		queries = append(queries, createOAuthAppQuery(updatedApp, queryUpdateOAuthApplicationByAppID, st.deploymentID))
	} else if len(existingApp.InboundAuthConfig) > 0 {
		clientID := ""
		if len(existingApp.InboundAuthConfig) > 0 && existingApp.InboundAuthConfig[0].OAuthAppConfig != nil {
			clientID = existingApp.InboundAuthConfig[0].OAuthAppConfig.ClientID
		}
		queries = append(queries, deleteOAuthAppQuery(clientID, st.deploymentID))
	} else if len(updatedApp.InboundAuthConfig) > 0 {
		queries = append(queries, createOAuthAppQuery(updatedApp, queryCreateOAuthApplication, st.deploymentID))
	}

	return executeTransaction(st.dbProvider, queries)
}

// DeleteApplication deletes an application from the database by its ID.
func (st *applicationStore) DeleteApplication(id string) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationStore"))

	dbClient, err := st.dbProvider.GetConfigDBClient()
	if err != nil {
		logger.Error("Failed to get database client", log.Error(err))
		return fmt.Errorf("failed to get database client: %w", err)
	}

	_, err = dbClient.Execute(queryDeleteApplicationByAppID, id, st.deploymentID)
	if err != nil {
		logger.Error("Failed to execute query", log.Error(err))
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}

// IsApplicationExists checks if an application exists in the database by ID.
func (st *applicationStore) IsApplicationExists(id string) (bool, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationStore"))

	dbClient, err := st.dbProvider.GetConfigDBClient()
	if err != nil {
		logger.Error("Failed to get database client", log.Error(err))
		return false, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(queryCheckApplicationExistsByID, id, st.deploymentID)
	if err != nil {
		logger.Error("Failed to execute existence check query", log.Error(err))
		return false, fmt.Errorf("failed to execute existence check query: %w", err)
	}

	return parseBoolFromCount(results)
}

// IsApplicationExistsByName checks if an application exists in the database by name.
func (st *applicationStore) IsApplicationExistsByName(name string) (bool, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationStore"))

	dbClient, err := st.dbProvider.GetConfigDBClient()
	if err != nil {
		logger.Error("Failed to get database client", log.Error(err))
		return false, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(queryCheckApplicationExistsByName, name, st.deploymentID)
	if err != nil {
		logger.Error("Failed to execute existence check query", log.Error(err))
		return false, fmt.Errorf("failed to execute existence check query: %w", err)
	}

	return parseBoolFromCount(results)
}

// IsApplicationDeclarative checks if an application is immutable.
// For database store, all applications are mutable (not declarative).
func (st *applicationStore) IsApplicationDeclarative(id string) bool {
	return false
}

// getAppJSONDataBytes constructs the JSON data bytes for the application.
func getAppJSONDataBytes(app *model.ApplicationProcessedDTO) ([]byte, error) {
	jsonData := map[string]interface{}{
		"url":        app.URL,
		"logo_url":   app.LogoURL,
		"tos_uri":    app.TosURI,
		"policy_uri": app.PolicyURI,
		"contacts":   app.Contacts,
	}

	// Include template if present
	if app.Template != "" {
		jsonData["template"] = app.Template
	}

	// Include allowed_user_types if present (include even if empty to preserve the field)
	if app.AllowedUserTypes != nil {
		jsonData["allowed_user_types"] = app.AllowedUserTypes
	}

	// Include assertion config if present
	if app.Assertion != nil {
		assertionData := map[string]interface{}{}
		if app.Assertion.ValidityPeriod != 0 {
			assertionData["validity_period"] = app.Assertion.ValidityPeriod
		}
		if len(app.Assertion.UserAttributes) > 0 {
			assertionData["user_attributes"] = app.Assertion.UserAttributes
		}
		if len(assertionData) > 0 {
			jsonData["assertion"] = assertionData
		}
	}

	jsonDataBytes, err := json.Marshal(jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal application JSON: %w", err)
	}
	return jsonDataBytes, nil
}

// getOAuthConfigJSONBytes constructs the OAuth configuration JSON data bytes.
func getOAuthConfigJSONBytes(inboundAuthConfig model.InboundAuthConfigProcessedDTO) ([]byte, error) {
	oauthConfig := oAuthConfig{
		RedirectURIs:            inboundAuthConfig.OAuthAppConfig.RedirectURIs,
		GrantTypes:              utils.ConvertToStringSlice(inboundAuthConfig.OAuthAppConfig.GrantTypes),
		ResponseTypes:           utils.ConvertToStringSlice(inboundAuthConfig.OAuthAppConfig.ResponseTypes),
		TokenEndpointAuthMethod: string(inboundAuthConfig.OAuthAppConfig.TokenEndpointAuthMethod),
		PKCERequired:            inboundAuthConfig.OAuthAppConfig.PKCERequired,
		PublicClient:            inboundAuthConfig.OAuthAppConfig.PublicClient,
		Scopes:                  inboundAuthConfig.OAuthAppConfig.Scopes,
	}

	// Include token config if present
	if inboundAuthConfig.OAuthAppConfig.Token != nil {
		oauthConfig.Token = &oAuthTokenConfig{}
		if inboundAuthConfig.OAuthAppConfig.Token.AccessToken != nil {
			oauthConfig.Token.AccessToken = &accessTokenConfig{
				ValidityPeriod: inboundAuthConfig.OAuthAppConfig.Token.AccessToken.ValidityPeriod,
				UserAttributes: inboundAuthConfig.OAuthAppConfig.Token.AccessToken.UserAttributes,
			}
		}
		if inboundAuthConfig.OAuthAppConfig.Token.IDToken != nil {
			oauthConfig.Token.IDToken = &idTokenConfig{
				ValidityPeriod: inboundAuthConfig.OAuthAppConfig.Token.IDToken.ValidityPeriod,
				UserAttributes: inboundAuthConfig.OAuthAppConfig.Token.IDToken.UserAttributes,
			}
		}
	}

	// Handle UserInfo config
	if inboundAuthConfig.OAuthAppConfig.UserInfo != nil {
		oauthConfig.UserInfo = &userInfoConfig{
			ResponseType:   inboundAuthConfig.OAuthAppConfig.UserInfo.ResponseType,
			UserAttributes: inboundAuthConfig.OAuthAppConfig.UserInfo.UserAttributes,
		}
	}

	// Handle ScopeClaims config
	oauthConfig.ScopeClaims = inboundAuthConfig.OAuthAppConfig.ScopeClaims

	oauthConfigJSONBytes, err := json.Marshal(oauthConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal OAuth configuration JSON: %w", err)
	}
	return oauthConfigJSONBytes, nil
}

// createOAuthAppQuery creates a query function for creating or updating an OAuth application.
func createOAuthAppQuery(app *model.ApplicationProcessedDTO,
	oauthAppMgtQuery dbmodel.DBQuery, deploymentID string) func(tx dbmodel.TxInterface) error {
	inboundAuthConfig := app.InboundAuthConfig[0]
	clientID := inboundAuthConfig.OAuthAppConfig.ClientID
	clientSecret := inboundAuthConfig.OAuthAppConfig.HashedClientSecret

	// Generate the OAuth config JSON
	oauthConfigJSON, err := getOAuthConfigJSONBytes(inboundAuthConfig)
	if err != nil {
		return func(tx dbmodel.TxInterface) error {
			return err
		}
	}

	return func(tx dbmodel.TxInterface) error {
		_, err := tx.Exec(oauthAppMgtQuery, app.ID, clientID, clientSecret, oauthConfigJSON, deploymentID)
		return err
	}
}

// deleteOAuthAppQuery creates a query function for deleting an OAuth application by client ID.
func deleteOAuthAppQuery(clientID string, deploymentID string) func(tx dbmodel.TxInterface) error {
	return func(tx dbmodel.TxInterface) error {
		_, err := tx.Exec(queryDeleteOAuthApplicationByClientID, clientID, deploymentID)
		return err
	}
}

// buildBasicApplicationFromResultRow constructs a BasicApplicationDTO from a database result row.
func buildBasicApplicationFromResultRow(row map[string]interface{}) (model.BasicApplicationDTO, error) {
	appID, ok := row["app_id"].(string)
	if !ok {
		return model.BasicApplicationDTO{}, fmt.Errorf("failed to parse app_id as string")
	}

	appName, ok := row["app_name"].(string)
	if !ok {
		return model.BasicApplicationDTO{}, fmt.Errorf("failed to parse app_name as string")
	}

	var description string
	if row["description"] == nil {
		description = ""
	} else if desc, ok := row["description"].(string); ok {
		description = desc
	} else {
		return model.BasicApplicationDTO{}, fmt.Errorf("failed to parse description as string")
	}

	authFlowID, ok := row["auth_flow_id"].(string)
	if !ok {
		return model.BasicApplicationDTO{}, fmt.Errorf("failed to parse auth_flow_id as string")
	}

	regisFlowID, ok := row["registration_flow_id"].(string)
	if !ok {
		return model.BasicApplicationDTO{}, fmt.Errorf("failed to parse registration_flow_id as string")
	}

	var isRegistrationFlowEnabledStr string
	switch v := row["is_registration_flow_enabled"].(type) {
	case string:
		isRegistrationFlowEnabledStr = v
	case []byte:
		isRegistrationFlowEnabledStr = string(v)
	default:
		logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationStore"))
		logger.Debug("Failed to parse is_registration_flow_enabled",
			log.String("type", fmt.Sprintf("%T", row["is_registration_flow_enabled"])),
			log.String("value", fmt.Sprintf("%v", row["is_registration_flow_enabled"])))
		return model.BasicApplicationDTO{},
			fmt.Errorf("failed to parse is_registration_flow_enabled as string or []byte")
	}
	isRegistrationFlowEnabled := utils.NumStringToBool(isRegistrationFlowEnabledStr)

	var themeID string
	if row["theme_id"] != nil {
		if tid, ok := row["theme_id"].(string); ok {
			themeID = tid
		}
	}

	var layoutID string
	if row["layout_id"] != nil {
		if lid, ok := row["layout_id"].(string); ok {
			layoutID = lid
		}
	}

	application := model.BasicApplicationDTO{
		ID:                        appID,
		Name:                      appName,
		Description:               description,
		AuthFlowID:                authFlowID,
		RegistrationFlowID:        regisFlowID,
		IsRegistrationFlowEnabled: isRegistrationFlowEnabled,
		ThemeID:                   themeID,
		LayoutID:                  layoutID,
	}

	if row["client_id"] != nil {
		clientID, ok := row["client_id"].(string)
		if !ok {
			return model.BasicApplicationDTO{}, fmt.Errorf("failed to parse client_id as string")
		}
		application.ClientID = clientID
	}

	// Extract logo_url and template from app_json if present.
	if row["app_json"] != nil {
		var appJSON string
		if v, ok := row["app_json"].(string); ok {
			appJSON = v
		} else if v, ok := row["app_json"].([]byte); ok {
			appJSON = string(v)
		}

		if appJSON != "" && appJSON != "{}" {
			var appJSONData map[string]interface{}
			if err := json.Unmarshal([]byte(appJSON), &appJSONData); err != nil {
				return model.BasicApplicationDTO{}, fmt.Errorf("failed to unmarshal app JSON: %w", err)
			}

			logoURL, err := extractStringFromJSON(appJSONData, "logo_url")
			if err != nil {
				return model.BasicApplicationDTO{}, err
			}
			application.LogoURL = logoURL

			template, err := extractStringFromJSON(appJSONData, "template")
			if err != nil {
				return model.BasicApplicationDTO{}, err
			}
			application.Template = template
		}
	}

	return application, nil
}

// extractStringFromJSON extracts a string value from JSON data, returns empty string if not found or invalid.
func extractStringFromJSON(data map[string]interface{}, key string) (string, error) {
	if data[key] == nil {
		return "", nil
	}
	if str, ok := data[key].(string); ok {
		return str, nil
	}
	return "", fmt.Errorf("failed to parse %s from app JSON", key)
}

// extractStringArrayFromJSON extracts a string array from JSON data.
func extractStringArrayFromJSON(data map[string]interface{}, key string) ([]string, error) {
	if data[key] == nil {
		return []string{}, nil
	}
	if arr, ok := data[key].([]interface{}); ok {
		result := make([]string, 0, len(arr))
		for i, item := range arr {
			if str, ok := item.(string); ok {
				result = append(result, str)
			} else {
				return nil, fmt.Errorf(
					"failed to parse %s from app JSON: item at index %d is not a string (type: %T, value: %v)",
					key, i, item, item)
			}
		}
		return result, nil
	}
	return nil, fmt.Errorf("failed to parse %s from app JSON", key)
}

// extractAssertionConfigFromJSON extracts assertion configuration from JSON data.
func extractAssertionConfigFromJSON(data map[string]interface{}) *model.AssertionConfig {
	assertionData, exists := data["assertion"]
	if !exists || assertionData == nil {
		return nil
	}
	assertionMap, ok := assertionData.(map[string]interface{})
	if !ok {
		return nil
	}

	config := &model.AssertionConfig{}
	if validityPeriod, ok := assertionMap["validity_period"].(float64); ok {
		config.ValidityPeriod = int64(validityPeriod)
	}
	if userAttrs, ok := assertionMap["user_attributes"].([]interface{}); ok {
		for _, attr := range userAttrs {
			if attrStr, ok := attr.(string); ok {
				config.UserAttributes = append(config.UserAttributes, attrStr)
			}
		}
	}
	return config
}

// buildApplicationFromResultRow constructs an Application object from a database result row.
func buildApplicationFromResultRow(row map[string]interface{}) (model.ApplicationProcessedDTO, error) {
	basicApp, err := buildBasicApplicationFromResultRow(row)
	if err != nil {
		return model.ApplicationProcessedDTO{}, err
	}

	// Extract JSON data from the row.
	var appJSON string
	if row["app_json"] == nil {
		appJSON = "{}"
	} else if v, ok := row["app_json"].(string); ok {
		appJSON = v
	} else if v, ok := row["app_json"].([]byte); ok {
		appJSON = string(v)
	} else {
		return model.ApplicationProcessedDTO{}, fmt.Errorf("failed to parse app_json as string or []byte")
	}

	var appJSONData map[string]interface{}
	if err := json.Unmarshal([]byte(appJSON), &appJSONData); err != nil {
		return model.ApplicationProcessedDTO{}, fmt.Errorf("failed to unmarshal app JSON: %w", err)
	}

	// Extract fields from JSON data using helper functions.
	url, err := extractStringFromJSON(appJSONData, "url")
	if err != nil {
		return model.ApplicationProcessedDTO{}, err
	}

	logoURL, err := extractStringFromJSON(appJSONData, "logo_url")
	if err != nil {
		return model.ApplicationProcessedDTO{}, err
	}

	tosURI, err := extractStringFromJSON(appJSONData, "tos_uri")
	if err != nil {
		return model.ApplicationProcessedDTO{}, err
	}

	policyURI, err := extractStringFromJSON(appJSONData, "policy_uri")
	if err != nil {
		return model.ApplicationProcessedDTO{}, err
	}

	contacts, err := extractStringArrayFromJSON(appJSONData, "contacts")
	if err != nil {
		return model.ApplicationProcessedDTO{}, err
	}

	// Extract allowed_user_types from app JSON if present
	allowedUserTypes, err := extractStringArrayFromJSON(appJSONData, "allowed_user_types")
	if err != nil {
		return model.ApplicationProcessedDTO{}, err
	}

	assertionConfig := extractAssertionConfigFromJSON(appJSONData)

	// Extract template from app JSON if present
	template, err := extractStringFromJSON(appJSONData, "template")
	if err != nil {
		return model.ApplicationProcessedDTO{}, err
	}

	application := model.ApplicationProcessedDTO{
		ID:                        basicApp.ID,
		Name:                      basicApp.Name,
		Description:               basicApp.Description,
		AuthFlowID:                basicApp.AuthFlowID,
		RegistrationFlowID:        basicApp.RegistrationFlowID,
		IsRegistrationFlowEnabled: basicApp.IsRegistrationFlowEnabled,
		ThemeID:                   basicApp.ThemeID,
		LayoutID:                  basicApp.LayoutID,
		Template:                  template,
		URL:                       url,
		LogoURL:                   logoURL,
		Assertion:                 assertionConfig,
		TosURI:                    tosURI,
		PolicyURI:                 policyURI,
		Contacts:                  contacts,
		AllowedUserTypes:          allowedUserTypes,
	}

	if basicApp.ClientID != "" {
		inboundAuthConfig, err := buildOAuthInboundAuthConfig(row, basicApp)
		if err != nil {
			return model.ApplicationProcessedDTO{}, err
		}
		application.InboundAuthConfig = []model.InboundAuthConfigProcessedDTO{inboundAuthConfig}
	}

	return application, nil
}

// buildOAuthInboundAuthConfig builds OAuth inbound auth configuration from database row and basic app data.
func buildOAuthInboundAuthConfig(row map[string]interface{}, basicApp model.BasicApplicationDTO) (
	model.InboundAuthConfigProcessedDTO, error) {
	hashedClientSecret, ok := row["client_secret"].(string)
	if !ok {
		return model.InboundAuthConfigProcessedDTO{}, fmt.Errorf("failed to parse client_secret as string")
	}

	// Extract OAuth JSON data from the row.
	var oauthConfigJSON string
	if row["oauth_config_json"] == nil {
		oauthConfigJSON = "{}"
	} else if v, ok := row["oauth_config_json"].(string); ok {
		oauthConfigJSON = v
	} else if v, ok := row["oauth_config_json"].([]byte); ok {
		oauthConfigJSON = string(v)
	} else {
		return model.InboundAuthConfigProcessedDTO{},
			fmt.Errorf("failed to parse oauth_config_json as string or []byte")
	}

	var oauthConfig oAuthConfig
	if err := json.Unmarshal([]byte(oauthConfigJSON), &oauthConfig); err != nil {
		return model.InboundAuthConfigProcessedDTO{}, fmt.Errorf("failed to unmarshal oauth config JSON: %w", err)
	}

	// Convert the typed arrays to the required types
	grantTypes := make([]oauth2const.GrantType, 0, len(oauthConfig.GrantTypes))
	for _, gt := range oauthConfig.GrantTypes {
		grantTypes = append(grantTypes, oauth2const.GrantType(gt))
	}

	responseTypes := make([]oauth2const.ResponseType, 0, len(oauthConfig.ResponseTypes))
	for _, rt := range oauthConfig.ResponseTypes {
		responseTypes = append(responseTypes, oauth2const.ResponseType(rt))
	}

	tokenEndpointAuthMethod := oauth2const.TokenEndpointAuthMethod(oauthConfig.TokenEndpointAuthMethod)

	// Extract token config from OAuth config if present
	var oauthTokenConfig *model.OAuthTokenConfig
	if oauthConfig.Token != nil {
		oauthTokenConfig = &model.OAuthTokenConfig{}
		if oauthConfig.Token.AccessToken != nil {
			userAttributes := oauthConfig.Token.AccessToken.UserAttributes
			if userAttributes == nil {
				userAttributes = make([]string, 0)
			}
			oauthTokenConfig.AccessToken = &model.AccessTokenConfig{
				ValidityPeriod: oauthConfig.Token.AccessToken.ValidityPeriod,
				UserAttributes: userAttributes,
			}
		}
		if oauthConfig.Token.IDToken != nil {
			userAttributes := oauthConfig.Token.IDToken.UserAttributes
			if userAttributes == nil {
				userAttributes = make([]string, 0)
			}
			oauthTokenConfig.IDToken = &model.IDTokenConfig{
				ValidityPeriod: oauthConfig.Token.IDToken.ValidityPeriod,
				UserAttributes: userAttributes,
			}
		}
	}

	// Handle UserInfo config
	var userInfoConfig *model.UserInfoConfig
	if oauthConfig.UserInfo != nil {
		userAttributes := oauthConfig.UserInfo.UserAttributes
		if userAttributes == nil {
			userAttributes = make([]string, 0)
		}
		userInfoConfig = &model.UserInfoConfig{
			ResponseType:   oauthConfig.UserInfo.ResponseType,
			UserAttributes: userAttributes,
		}
	}

	// Handle ScopeClaims config
	scopeClaims := oauthConfig.ScopeClaims
	if scopeClaims == nil {
		scopeClaims = make(map[string][]string)
	}

	// TODO: Need to refactor when supporting other/multiple inbound auth types.
	inboundAuthConfig := model.InboundAuthConfigProcessedDTO{
		Type: model.OAuthInboundAuthType,
		OAuthAppConfig: &model.OAuthAppConfigProcessedDTO{
			AppID:                   basicApp.ID,
			ClientID:                basicApp.ClientID,
			HashedClientSecret:      hashedClientSecret,
			RedirectURIs:            oauthConfig.RedirectURIs,
			GrantTypes:              grantTypes,
			ResponseTypes:           responseTypes,
			TokenEndpointAuthMethod: tokenEndpointAuthMethod,
			PKCERequired:            oauthConfig.PKCERequired,
			PublicClient:            oauthConfig.PublicClient,
			Token:                   oauthTokenConfig,
			Scopes:                  oauthConfig.Scopes,
			UserInfo:                userInfoConfig,
			ScopeClaims:             scopeClaims,
		},
	}
	return inboundAuthConfig, nil
}

// parseBoolFromCount parses the count result from an existence check query.
func parseBoolFromCount(results []map[string]interface{}) (bool, error) {
	if len(results) == 0 {
		return false, nil
	}
	if countVal, ok := results[0]["count"].(int64); ok {
		return countVal > 0, nil
	}
	return false, fmt.Errorf("failed to parse count from query result")
}

// executeTransaction is a helper function to handle database transactions.
func executeTransaction(dbProvider provider.DBProviderInterface,
	queries []func(tx dbmodel.TxInterface) error) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationStore"))

	dbClient, err := dbProvider.GetConfigDBClient()
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	tx, err := dbClient.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	for _, query := range queries {
		if err := query(tx); err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				logger.Error("Failed to rollback transaction", log.Error(rollbackErr))
				err = errors.Join(err, errors.New("failed to rollback transaction: "+rollbackErr.Error()))
			}
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
