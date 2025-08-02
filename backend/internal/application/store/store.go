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

// Package store provides functionality for handling application data persistence.
package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/asgardeo/thunder/internal/application/constants"
	"github.com/asgardeo/thunder/internal/application/model"
	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/system/database/client"
	dbmodel "github.com/asgardeo/thunder/internal/system/database/model"
	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

// ApplicationStoreInterface defines the interface for application data persistence operations.
type ApplicationStoreInterface interface {
	CreateApplication(app model.ApplicationProcessedDTO) error
	GetApplicationList() ([]model.BasicApplicationDTO, error)
	GetApplication(id string) (model.ApplicationProcessedDTO, error)
	GetOAuthApplication(clientID string) (*model.OAuthAppConfigProcessed, error)
	UpdateApplication(app *model.ApplicationProcessedDTO) error
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
	return createOrUpdateApplication(&app, QueryCreateApplication, QueryCreateOAuthApplication)
}

// GetApplicationList retrieves a list of applications from the database.
func (st *ApplicationStore) GetApplicationList() ([]model.BasicApplicationDTO, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationPersistence"))

	dbClient, err := provider.NewDBProvider().GetDBClient("identity")
	if err != nil {
		logger.Error("Failed to get database client", log.Error(err))
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}
	defer func(dbc client.DBClientInterface) {
		err := dbc.Close()
		if err != nil {
			logger.Error("Failed to close database client", log.Error(err))
		}
	}(dbClient)

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

// GetApplication retrieves a specific application by its ID from the database.
func (st *ApplicationStore) GetApplication(id string) (model.ApplicationProcessedDTO, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationStore"))

	dbClient, err := provider.NewDBProvider().GetDBClient("identity")
	if err != nil {
		logger.Error("Failed to get database client", log.Error(err))
		return model.ApplicationProcessedDTO{}, fmt.Errorf("failed to get database client: %w", err)
	}
	defer func() {
		if closeErr := dbClient.Close(); closeErr != nil {
			logger.Error("Failed to close database client", log.Error(closeErr))
			err = fmt.Errorf("failed to close database client: %w", closeErr)
		}
	}()

	results, err := dbClient.Query(QueryGetApplicationByAppID, id)
	if err != nil {
		logger.Error("Failed to execute query", log.Error(err))
		return model.ApplicationProcessedDTO{}, fmt.Errorf("failed to execute query: %w", err)
	}

	if len(results) == 0 {
		logger.Error("application not found")
		return model.ApplicationProcessedDTO{}, fmt.Errorf("application not found")
	}
	if len(results) != 1 {
		logger.Error("unexpected number of results")
		return model.ApplicationProcessedDTO{}, fmt.Errorf("unexpected number of results: %d", len(results))
	}

	row := results[0]
	application, err := buildApplicationFromResultRow(row)
	if err != nil {
		logger.Error("failed to build application from result row", log.Error(err))
		return model.ApplicationProcessedDTO{}, fmt.Errorf("failed to build application from result row: %w", err)
	}

	return application, nil
}

// GetOAuthApplication retrieves an OAuth application by its client ID.
func (st *ApplicationStore) GetOAuthApplication(clientID string) (*model.OAuthAppConfigProcessed, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationStore"))

	dbClient, err := provider.NewDBProvider().GetDBClient("identity")
	if err != nil {
		logger.Error("Failed to get database client", log.Error(err))
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}
	defer func() {
		if closeErr := dbClient.Close(); closeErr != nil {
			logger.Error("Failed to close database client", log.Error(closeErr))
			err = fmt.Errorf("failed to close database client: %w", closeErr)
		}
	}()

	results, err := dbClient.Query(QueryGetOAuthApplicationByClientID, clientID)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errors.New("OAuth application not found")
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

	var redirectURIs []string
	if row["callback_uris"] != nil {
		if uris, ok := row["callback_uris"].(string); ok {
			redirectURIs = utils.ParseStringArray(uris, ",")
		} else {
			return nil, errors.New("failed to parse callback_uris as string")
		}
	}

	var grantTypes []oauth2const.GrantType
	if row["grant_types"] != nil {
		if grants, ok := row["grant_types"].(string); ok {
			grantTypes = utils.ParseTypedStringArray[oauth2const.GrantType](grants, ",")
		} else {
			return nil, errors.New("failed to parse grant_types as string")
		}
	}

	var responseTypes []oauth2const.ResponseType
	if row["response_types"] != nil {
		if responses, ok := row["response_types"].(string); ok {
			responseTypes = utils.ParseTypedStringArray[oauth2const.ResponseType](responses, ",")
		} else {
			return nil, errors.New("failed to parse response_types as string")
		}
	}

	var tokenEndpointAuthMethods []oauth2const.TokenEndpointAuthMethod
	if row["token_endpoint_auth_methods"] != nil {
		if methods, ok := row["token_endpoint_auth_methods"].(string); ok {
			tokenEndpointAuthMethods = utils.ParseTypedStringArray[oauth2const.TokenEndpointAuthMethod](methods, ",")
		} else {
			return nil, errors.New("failed to parse token_endpoint_auth_methods as string")
		}
	}

	return &model.OAuthAppConfigProcessed{
		AppID:                   appID,
		ClientID:                clientID,
		HashedClientSecret:      hashedClientSecret,
		RedirectURIs:            redirectURIs,
		GrantTypes:              grantTypes,
		ResponseTypes:           responseTypes,
		TokenEndpointAuthMethod: tokenEndpointAuthMethods,
	}, nil
}

// UpdateApplication updates an existing application in the database.
func (st *ApplicationStore) UpdateApplication(app *model.ApplicationProcessedDTO) error {
	return createOrUpdateApplication(app, QueryUpdateApplicationByAppID, QueryUpdateOAuthApplicationByAppID)
}

// DeleteApplication deletes an application from the database by its ID.
func (st *ApplicationStore) DeleteApplication(id string) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationStore"))

	dbClient, err := provider.NewDBProvider().GetDBClient("identity")
	if err != nil {
		logger.Error("Failed to get database client", log.Error(err))
		return fmt.Errorf("failed to get database client: %w", err)
	}
	defer func() {
		if closeErr := dbClient.Close(); closeErr != nil {
			logger.Error("Failed to close database client", log.Error(closeErr))
			err = fmt.Errorf("failed to close database client: %w", closeErr)
		}
	}()

	_, err = dbClient.Execute(QueryDeleteApplicationByAppID, id)
	if err != nil {
		logger.Error("Failed to execute query", log.Error(err))
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}

// getProcessedOAuthParams extracts oauth configuration strings from the inbound auth configs.
func getProcessedOAuthParams(inboundAuthConfig model.InboundAuthConfigProcessed) (string, string, string, string) {
	callBackURIs := ""
	if len(inboundAuthConfig.OAuthAppConfig.RedirectURIs) > 0 {
		callBackURIs = strings.Join(inboundAuthConfig.OAuthAppConfig.RedirectURIs, ",")
	}

	grantTypes := ""
	if len(inboundAuthConfig.OAuthAppConfig.GrantTypes) > 0 {
		strs := make([]string, len(inboundAuthConfig.OAuthAppConfig.GrantTypes))
		for i, g := range inboundAuthConfig.OAuthAppConfig.GrantTypes {
			strs[i] = string(g)
		}
		grantTypes = strings.Join(strs, ",")
	}

	responseTypes := ""
	if len(inboundAuthConfig.OAuthAppConfig.ResponseTypes) > 0 {
		strs := make([]string, len(inboundAuthConfig.OAuthAppConfig.ResponseTypes))
		for i, r := range inboundAuthConfig.OAuthAppConfig.ResponseTypes {
			strs[i] = string(r)
		}
		responseTypes = strings.Join(strs, ",")
	}

	tokenAuthMethods := ""
	if len(inboundAuthConfig.OAuthAppConfig.TokenEndpointAuthMethod) > 0 {
		strs := make([]string, len(inboundAuthConfig.OAuthAppConfig.TokenEndpointAuthMethod))
		for i, m := range inboundAuthConfig.OAuthAppConfig.TokenEndpointAuthMethod {
			strs[i] = string(m)
		}
		tokenAuthMethods = strings.Join(strs, ",")
	}

	return callBackURIs, grantTypes, responseTypes, tokenAuthMethods
}

// createOrUpdateApplication creates or updates an application in the database.
func createOrUpdateApplication(app *model.ApplicationProcessedDTO,
	appMgtQuery dbmodel.DBQuery, oauthAppMgtQuery dbmodel.DBQuery) error {
	// TODO: Need to refactor when supporting other/multiple inbound auth types.
	inboundAuthConfig := app.InboundAuthConfig[0]
	callBackURIs, grantTypes, responseTypes, tokenAuthMethods := getProcessedOAuthParams(inboundAuthConfig)

	// Construct the app JSON
	jsonData := map[string]interface{}{
		"url":      app.URL,
		"logo_url": app.LogoURL,
	}
	jsonDataBytes, err := json.Marshal(jsonData)
	if err != nil {
		return fmt.Errorf("failed to marshal application JSON: %w", err)
	}

	queries := []func(tx dbmodel.TxInterface) error{
		func(tx dbmodel.TxInterface) error {
			_, err := tx.Exec(appMgtQuery.Query, app.ID, app.Name, app.Description,
				app.AuthFlowGraphID, app.RegistrationFlowGraphID, app.IsRegistrationFlowEnabled, jsonDataBytes)
			return err
		},
		func(tx dbmodel.TxInterface) error {
			_, err := tx.Exec(oauthAppMgtQuery.Query, app.ID,
				inboundAuthConfig.OAuthAppConfig.ClientID, inboundAuthConfig.OAuthAppConfig.HashedClientSecret,
				callBackURIs, grantTypes, responseTypes, tokenAuthMethods)
			return err
		},
	}

	return executeTransaction(queries)
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

	isRegistrationFlowEnabledStr, ok := row["is_registration_flow_enabled"].(string)
	if !ok {
		return model.BasicApplicationDTO{}, fmt.Errorf("failed to parse is_registration_flow_enabled as string")
	}
	isRegistrationFlowEnabled := sysutils.NumStringToBool(isRegistrationFlowEnabledStr)

	clientID, ok := row["consumer_key"].(string)
	if !ok {
		return model.BasicApplicationDTO{}, fmt.Errorf("failed to parse consumer_key as string")
	}

	application := model.BasicApplicationDTO{
		ID:                        appID,
		Name:                      appName,
		Description:               description,
		AuthFlowGraphID:           authFlowGraphID,
		RegistrationFlowGraphID:   regisFlowGraphID,
		IsRegistrationFlowEnabled: isRegistrationFlowEnabled,
		ClientID:                  clientID,
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

	hashedClientSecret, ok := row["consumer_secret"].(string)
	if !ok {
		return model.ApplicationProcessedDTO{}, fmt.Errorf("failed to parse consumer_secret as string")
	}

	var redirectURIs []string
	if row["callback_uris"] != nil {
		if uris, ok := row["callback_uris"].(string); ok {
			redirectURIs = utils.ParseStringArray(uris, ",")
		} else {
			return model.ApplicationProcessedDTO{}, fmt.Errorf("failed to parse callback_uris as string")
		}
	}

	var grantTypes []oauth2const.GrantType
	if row["grant_types"] != nil {
		if grants, ok := row["grant_types"].(string); ok {
			grantTypes = utils.ParseTypedStringArray[oauth2const.GrantType](grants, ",")
		} else {
			return model.ApplicationProcessedDTO{}, fmt.Errorf("failed to parse grant_types as string")
		}
	}

	var responseTypes []oauth2const.ResponseType
	if row["response_types"] != nil {
		if responses, ok := row["response_types"].(string); ok {
			responseTypes = utils.ParseTypedStringArray[oauth2const.ResponseType](responses, ",")
		} else {
			return model.ApplicationProcessedDTO{}, fmt.Errorf("failed to parse response_types as string")
		}
	}

	var tokenEndpointAuthMethods []oauth2const.TokenEndpointAuthMethod
	if row["token_endpoint_auth_methods"] != nil {
		if methods, ok := row["token_endpoint_auth_methods"].(string); ok {
			tokenEndpointAuthMethods = utils.ParseTypedStringArray[oauth2const.TokenEndpointAuthMethod](methods, ",")
		} else {
			return model.ApplicationProcessedDTO{}, fmt.Errorf("failed to parse token_endpoint_auth_methods as string")
		}
	}

	// TODO: Need to refactor when supporting other/multiple inbound auth types.
	inboundAuthConfig := model.InboundAuthConfigProcessed{
		Type: constants.OAuthInboundAuthType,
		OAuthAppConfig: &model.OAuthAppConfigProcessed{
			AppID:                   basicApp.ID,
			ClientID:                basicApp.ClientID,
			HashedClientSecret:      hashedClientSecret,
			RedirectURIs:            redirectURIs,
			GrantTypes:              grantTypes,
			ResponseTypes:           responseTypes,
			TokenEndpointAuthMethod: tokenEndpointAuthMethods,
		},
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
		InboundAuthConfig:         []model.InboundAuthConfigProcessed{inboundAuthConfig},
	}

	return application, nil
}

// executeTransaction is a helper function to handle database transactions.
func executeTransaction(queries []func(tx dbmodel.TxInterface) error) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationStore"))

	dbClient, err := provider.NewDBProvider().GetDBClient("identity")
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}
	defer func() {
		if closeErr := dbClient.Close(); closeErr != nil {
			logger.Error("Failed to close database client", log.Error(closeErr))
			err = fmt.Errorf("failed to close database client: %w", closeErr)
		}
	}()

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
