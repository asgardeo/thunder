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

// Package store provides functionality for handling application data persistence.
package store

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/asgardeo/thunder/internal/application/constants"
	"github.com/asgardeo/thunder/internal/application/model"
	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	dbmodel "github.com/asgardeo/thunder/internal/system/database/model"
	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

// oAuthConfig is the structure for unmarshaling OAuth configuration JSON.
type oAuthConfig struct {
	RedirectURIs            []string          `json:"redirect_uris"`
	GrantTypes              []string          `json:"grant_types"`
	ResponseTypes           []string          `json:"response_types"`
	TokenEndpointAuthMethod []string          `json:"token_endpoint_auth_methods"`
	PKCERequired            bool              `json:"pkce_required"`
	Token                   *oAuthTokenConfig `json:"token,omitempty"`
}

// oAuthTokenConfig represents the OAuth token configuration structure for JSON marshaling/unmarshaling.
type oAuthTokenConfig struct {
	AccessToken *tokenConfig `json:"access_token,omitempty"`
}

// tokenConfig represents the token configuration structure for JSON marshaling/unmarshaling.
type tokenConfig struct {
	Issuer         string   `json:"issuer,omitempty"`
	ValidityPeriod int64    `json:"validity_period,omitempty"`
	UserAttributes []string `json:"user_attributes,omitempty"`
}

// ApplicationStoreInterface defines the interface for application data persistence operations.
type ApplicationStoreInterface interface {
	CreateApplication(app model.ApplicationProcessedDTO) error
	GetTotalApplicationCount() (int, error)
	GetApplicationList() ([]model.BasicApplicationDTO, error)
	GetOAuthApplication(clientID string) (*model.OAuthAppConfigProcessedDTO, error)
	GetApplicationByID(id string) (*model.ApplicationProcessedDTO, error)
	GetApplicationByName(name string) (*model.ApplicationProcessedDTO, error)
	UpdateApplication(existingApp, updatedApp *model.ApplicationProcessedDTO) error
	DeleteApplication(id string) error
}

// ApplicationStore implements the ApplicationStoreInterface for handling application data persistence.
type ApplicationStore struct{}

// NewApplicationStore creates a new instance of ApplicationStore.
func NewApplicationStore() ApplicationStoreInterface {
	return &ApplicationStore{}
}

// CreateApplication creates a new application in the database.
func (st *ApplicationStore) CreateApplication(app model.ApplicationProcessedDTO) error {
	jsonDataBytes, err := getAppJSONDataBytes(&app)
	if err != nil {
		return err
	}

	queries := []func(tx dbmodel.TxInterface) error{
		func(tx dbmodel.TxInterface) error {
			isRegistrationEnabledStr := utils.BoolToNumString(app.IsRegistrationFlowEnabled)
			_, err := tx.Exec(QueryCreateApplication.Query, app.ID, app.Name, app.Description,
				app.AuthFlowGraphID, app.RegistrationFlowGraphID, isRegistrationEnabledStr, jsonDataBytes)
			return err
		},
	}
	// TODO: Need to refactor when supporting other/multiple inbound auth types.
	if len(app.InboundAuthConfig) > 0 {
		queries = append(queries, createOAuthAppQuery(&app, QueryCreateOAuthApplication))
	}

	return executeTransaction(queries)
}

// GetTotalApplicationCount retrieves the total count of applications from the database.
func (st *ApplicationStore) GetTotalApplicationCount() (int, error) {
	dbClient, err := provider.GetDBProvider().GetDBClient("identity")
	if err != nil {
		return 0, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(QueryGetApplicationCount)
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
func (st *ApplicationStore) GetApplicationList() ([]model.BasicApplicationDTO, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationPersistence"))

	dbClient, err := provider.GetDBProvider().GetDBClient("identity")
	if err != nil {
		logger.Error("Failed to get database client", log.Error(err))
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(QueryGetApplicationList)
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
func (st *ApplicationStore) GetOAuthApplication(clientID string) (*model.OAuthAppConfigProcessedDTO, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationStore"))

	dbClient, err := provider.GetDBProvider().GetDBClient("identity")
	if err != nil {
		logger.Error("Failed to get database client", log.Error(err))
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(QueryGetOAuthApplicationByClientID, clientID)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, constants.ApplicationNotFoundError
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

	tokenEndpointAuthMethods := make([]oauth2const.TokenEndpointAuthMethod, 0)
	for _, am := range oAuthConfig.TokenEndpointAuthMethod {
		tokenEndpointAuthMethods = append(tokenEndpointAuthMethods, oauth2const.TokenEndpointAuthMethod(am))
	}

	// Convert token config if present
	var oauthTokenConfig *model.OAuthTokenConfig
	if oAuthConfig.Token != nil && oAuthConfig.Token.AccessToken != nil {
		oauthTokenConfig = &model.OAuthTokenConfig{
			AccessToken: &model.TokenConfig{
				Issuer:         oAuthConfig.Token.AccessToken.Issuer,
				ValidityPeriod: oAuthConfig.Token.AccessToken.ValidityPeriod,
				UserAttributes: oAuthConfig.Token.AccessToken.UserAttributes,
			},
		}
	}

	return &model.OAuthAppConfigProcessedDTO{
		AppID:                   appID,
		ClientID:                clientID,
		HashedClientSecret:      hashedClientSecret,
		RedirectURIs:            oAuthConfig.RedirectURIs,
		GrantTypes:              grantTypes,
		ResponseTypes:           responseTypes,
		TokenEndpointAuthMethod: tokenEndpointAuthMethods,
		PKCERequired:            oAuthConfig.PKCERequired,
		Token:                   oauthTokenConfig,
	}, nil
}

// GetApplicationByID retrieves a specific application by its ID from the database.
func (st *ApplicationStore) GetApplicationByID(id string) (*model.ApplicationProcessedDTO, error) {
	return st.getApplicationByQuery(QueryGetApplicationByAppID, id)
}

// GetApplicationByName retrieves a specific application by its name from the database.
func (st *ApplicationStore) GetApplicationByName(name string) (*model.ApplicationProcessedDTO, error) {
	return st.getApplicationByQuery(QueryGetApplicationByName, name)
}

// getApplicationByQuery retrieves a specific application from the database using the provided query and parameter.
func (st *ApplicationStore) getApplicationByQuery(query dbmodel.DBQuery, param string) (
	*model.ApplicationProcessedDTO, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationStore"))

	dbClient, err := provider.GetDBProvider().GetDBClient("identity")
	if err != nil {
		logger.Error("Failed to get database client", log.Error(err))
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(query, param)
	if err != nil {
		logger.Error("Failed to execute query", log.Error(err))
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	if len(results) == 0 {
		return nil, constants.ApplicationNotFoundError
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
func (st *ApplicationStore) UpdateApplication(existingApp, updatedApp *model.ApplicationProcessedDTO) error {
	jsonDataBytes, err := getAppJSONDataBytes(updatedApp)
	if err != nil {
		return err
	}

	queries := []func(tx dbmodel.TxInterface) error{
		func(tx dbmodel.TxInterface) error {
			isRegistrationEnabledStr := utils.BoolToNumString(updatedApp.IsRegistrationFlowEnabled)
			_, err := tx.Exec(QueryUpdateApplicationByAppID.Query, updatedApp.ID, updatedApp.Name,
				updatedApp.Description, updatedApp.AuthFlowGraphID, updatedApp.RegistrationFlowGraphID,
				isRegistrationEnabledStr, jsonDataBytes)
			return err
		},
	}
	// TODO: Need to refactor when supporting other/multiple inbound auth types.
	if len(updatedApp.InboundAuthConfig) > 0 && len(existingApp.InboundAuthConfig) > 0 {
		queries = append(queries, createOAuthAppQuery(updatedApp, QueryUpdateOAuthApplicationByAppID))
	} else if len(existingApp.InboundAuthConfig) > 0 {
		clientID := ""
		if len(existingApp.InboundAuthConfig) > 0 && existingApp.InboundAuthConfig[0].OAuthAppConfig != nil {
			clientID = existingApp.InboundAuthConfig[0].OAuthAppConfig.ClientID
		}
		queries = append(queries, deleteOAuthAppQuery(clientID))
	} else if len(updatedApp.InboundAuthConfig) > 0 {
		queries = append(queries, createOAuthAppQuery(updatedApp, QueryCreateOAuthApplication))
	}

	return executeTransaction(queries)
}

// DeleteApplication deletes an application from the database by its ID.
func (st *ApplicationStore) DeleteApplication(id string) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationStore"))

	dbClient, err := provider.GetDBProvider().GetDBClient("identity")
	if err != nil {
		logger.Error("Failed to get database client", log.Error(err))
		return fmt.Errorf("failed to get database client: %w", err)
	}

	_, err = dbClient.Execute(QueryDeleteApplicationByAppID, id)
	if err != nil {
		logger.Error("Failed to execute query", log.Error(err))
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}

// getAppJSONDataBytes constructs the JSON data bytes for the application.
func getAppJSONDataBytes(app *model.ApplicationProcessedDTO) ([]byte, error) {
	jsonData := map[string]interface{}{
		"url":      app.URL,
		"logo_url": app.LogoURL,
	}

	// Include token config if present
	if app.Token != nil {
		tokenData := map[string]interface{}{}
		if app.Token.Issuer != "" {
			tokenData["issuer"] = app.Token.Issuer
		}
		if app.Token.ValidityPeriod != 0 {
			tokenData["validity_period"] = app.Token.ValidityPeriod
		}
		if len(app.Token.UserAttributes) > 0 {
			tokenData["user_attributes"] = app.Token.UserAttributes
		}
		if len(tokenData) > 0 {
			jsonData["token"] = tokenData
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
		RedirectURIs:  inboundAuthConfig.OAuthAppConfig.RedirectURIs,
		GrantTypes:    sysutils.ConvertToStringSlice(inboundAuthConfig.OAuthAppConfig.GrantTypes),
		ResponseTypes: sysutils.ConvertToStringSlice(inboundAuthConfig.OAuthAppConfig.ResponseTypes),
		TokenEndpointAuthMethod: sysutils.ConvertToStringSlice(
			inboundAuthConfig.OAuthAppConfig.TokenEndpointAuthMethod),
		PKCERequired: inboundAuthConfig.OAuthAppConfig.PKCERequired,
	}

	// Include token config if present
	if inboundAuthConfig.OAuthAppConfig.Token != nil && inboundAuthConfig.OAuthAppConfig.Token.AccessToken != nil {
		oauthConfig.Token = &oAuthTokenConfig{
			AccessToken: &tokenConfig{
				Issuer:         inboundAuthConfig.OAuthAppConfig.Token.AccessToken.Issuer,
				ValidityPeriod: inboundAuthConfig.OAuthAppConfig.Token.AccessToken.ValidityPeriod,
				UserAttributes: inboundAuthConfig.OAuthAppConfig.Token.AccessToken.UserAttributes,
			},
		}
	}

	oauthConfigJSONBytes, err := json.Marshal(oauthConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal OAuth configuration JSON: %w", err)
	}
	return oauthConfigJSONBytes, nil
}

// createOAuthAppQuery creates a query function for creating or updating an OAuth application.
func createOAuthAppQuery(app *model.ApplicationProcessedDTO,
	oauthAppMgtQuery dbmodel.DBQuery) func(tx dbmodel.TxInterface) error {
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
		_, err := tx.Exec(oauthAppMgtQuery.Query, app.ID, clientID, clientSecret, oauthConfigJSON)
		return err
	}
}

// deleteOAuthAppQuery creates a query function for deleting an OAuth application by client ID.
func deleteOAuthAppQuery(clientID string) func(tx dbmodel.TxInterface) error {
	return func(tx dbmodel.TxInterface) error {
		_, err := tx.Exec(QueryDeleteOAuthApplicationByClientID.Query, clientID)
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

	authFlowGraphID, ok := row["auth_flow_graph_id"].(string)
	if !ok {
		return model.BasicApplicationDTO{}, fmt.Errorf("failed to parse auth_flow_graph_id as string")
	}

	regisFlowGraphID, ok := row["registration_flow_graph_id"].(string)
	if !ok {
		return model.BasicApplicationDTO{}, fmt.Errorf("failed to parse registration_flow_graph_id as string")
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
	isRegistrationFlowEnabled := sysutils.NumStringToBool(isRegistrationFlowEnabledStr)

	application := model.BasicApplicationDTO{
		ID:                        appID,
		Name:                      appName,
		Description:               description,
		AuthFlowGraphID:           authFlowGraphID,
		RegistrationFlowGraphID:   regisFlowGraphID,
		IsRegistrationFlowEnabled: isRegistrationFlowEnabled,
	}

	if row["consumer_key"] != nil {
		clientID, ok := row["consumer_key"].(string)
		if !ok {
			return model.BasicApplicationDTO{}, fmt.Errorf("failed to parse consumer_key as string")
		}
		application.ClientID = clientID
	}

	return application, nil
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

	// Extract URL and LogoURL from the app JSON data.
	var url string
	if appJSONData["url"] == nil {
		url = ""
	} else if u, ok := appJSONData["url"].(string); ok {
		url = u
	} else {
		return model.ApplicationProcessedDTO{}, fmt.Errorf("failed to parse url from app JSON")
	}

	var logoURL string
	if appJSONData["logo_url"] == nil {
		logoURL = ""
	} else if lu, ok := appJSONData["logo_url"].(string); ok {
		logoURL = lu
	} else {
		return model.ApplicationProcessedDTO{}, fmt.Errorf("failed to parse logo_url from app JSON")
	}

	// Extract token config from app JSON if present
	var rootTokenConfig *model.TokenConfig
	if tokenData, exists := appJSONData["token"]; exists && tokenData != nil {
		if tokenMap, ok := tokenData.(map[string]interface{}); ok {
			rootTokenConfig = &model.TokenConfig{}
			if issuer, ok := tokenMap["issuer"].(string); ok {
				rootTokenConfig.Issuer = issuer
			}
			if validityPeriod, ok := tokenMap["validity_period"].(float64); ok {
				vp := int64(validityPeriod)
				rootTokenConfig.ValidityPeriod = vp
			}
			if userAttrs, ok := tokenMap["user_attributes"].([]interface{}); ok {
				for _, attr := range userAttrs {
					if attrStr, ok := attr.(string); ok {
						rootTokenConfig.UserAttributes = append(rootTokenConfig.UserAttributes, attrStr)
					}
				}
			}
		}
	}

	application := model.ApplicationProcessedDTO{
		ID:                        basicApp.ID,
		Name:                      basicApp.Name,
		Description:               basicApp.Description,
		AuthFlowGraphID:           basicApp.AuthFlowGraphID,
		RegistrationFlowGraphID:   basicApp.RegistrationFlowGraphID,
		IsRegistrationFlowEnabled: basicApp.IsRegistrationFlowEnabled,
		URL:                       url,
		LogoURL:                   logoURL,
		Token:                     rootTokenConfig,
	}

	if basicApp.ClientID != "" {
		hashedClientSecret, ok := row["consumer_secret"].(string)
		if !ok {
			return model.ApplicationProcessedDTO{}, fmt.Errorf("failed to parse consumer_secret as string")
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
			return model.ApplicationProcessedDTO{}, fmt.Errorf("failed to parse oauth_config_json as string or []byte")
		}

		var oauthConfig oAuthConfig
		if err := json.Unmarshal([]byte(oauthConfigJSON), &oauthConfig); err != nil {
			return model.ApplicationProcessedDTO{}, fmt.Errorf("failed to unmarshal oauth config JSON: %w", err)
		}

		// Convert the typed arrays to the required types
		var grantTypes []oauth2const.GrantType
		for _, gt := range oauthConfig.GrantTypes {
			grantTypes = append(grantTypes, oauth2const.GrantType(gt))
		}

		var responseTypes []oauth2const.ResponseType
		for _, rt := range oauthConfig.ResponseTypes {
			responseTypes = append(responseTypes, oauth2const.ResponseType(rt))
		}

		var tokenEndpointAuthMethods []oauth2const.TokenEndpointAuthMethod
		for _, am := range oauthConfig.TokenEndpointAuthMethod {
			tokenEndpointAuthMethods = append(tokenEndpointAuthMethods, oauth2const.TokenEndpointAuthMethod(am))
		}

		// Extract token config from OAuth config if present
		var oauthTokenConfig *model.OAuthTokenConfig
		if oauthConfig.Token != nil && oauthConfig.Token.AccessToken != nil {
			oauthTokenConfig = &model.OAuthTokenConfig{
				AccessToken: &model.TokenConfig{
					Issuer:         oauthConfig.Token.AccessToken.Issuer,
					ValidityPeriod: oauthConfig.Token.AccessToken.ValidityPeriod,
					UserAttributes: oauthConfig.Token.AccessToken.UserAttributes,
				},
			}
		}

		// TODO: Need to refactor when supporting other/multiple inbound auth types.
		inboundAuthConfig := model.InboundAuthConfigProcessedDTO{
			Type: constants.OAuthInboundAuthType,
			OAuthAppConfig: &model.OAuthAppConfigProcessedDTO{
				AppID:                   basicApp.ID,
				ClientID:                basicApp.ClientID,
				HashedClientSecret:      hashedClientSecret,
				RedirectURIs:            oauthConfig.RedirectURIs,
				GrantTypes:              grantTypes,
				ResponseTypes:           responseTypes,
				TokenEndpointAuthMethod: tokenEndpointAuthMethods,
				PKCERequired:            oauthConfig.PKCERequired,
				Token:                   oauthTokenConfig,
			},
		}
		application.InboundAuthConfig = []model.InboundAuthConfigProcessedDTO{inboundAuthConfig}
	}

	return application, nil
}

// executeTransaction is a helper function to handle database transactions.
func executeTransaction(queries []func(tx dbmodel.TxInterface) error) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationStore"))

	dbClient, err := provider.GetDBProvider().GetDBClient("identity")
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
