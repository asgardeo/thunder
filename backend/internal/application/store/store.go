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
	"errors"
	"fmt"
	"strings"

	"github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/system/database/client"
	dbmodel "github.com/asgardeo/thunder/internal/system/database/model"
	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
)

// CreateApplication creates a new application in the database.
func CreateApplication(app model.Application) error {
	queries := []func(tx dbmodel.TxInterface) error{
		func(tx dbmodel.TxInterface) error {
			_, err := tx.Exec(QueryCreateApplication.Query, app.ID, app.OrgID, app.Name, app.Description, app.AuthFlowGraphID,
				app.RegistrationFlowGraphID)
			return err
		},
		func(tx dbmodel.TxInterface) error {
			_, err := tx.Exec(QueryCreateOAuthApplication.Query, app.ID, app.ClientID, app.ClientSecret,
				strings.Join(app.CallbackURLs, ","), strings.Join(app.SupportedGrantTypes, ","))
			return err
		},
	}

	return executeTransaction(queries)
}

// GetApplicationList retrieves a list of applications from the database.
func GetApplicationList() ([]model.Application, error) {
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

	applications := make([]model.Application, 0)

	for _, row := range results {
		application, err := buildApplicationFromResultRow(row)
		if err != nil {
			logger.Error("failed to build application from result row", log.Error(err))
			return nil, fmt.Errorf("failed to build application from result row: %w", err)
		}
		applications = append(applications, application)
	}

	return applications, nil
}

// GetApplication retrieves a specific application by its ID from the database.
func GetApplication(id string) (model.Application, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationStore"))

	dbClient, err := provider.NewDBProvider().GetDBClient("identity")
	if err != nil {
		logger.Error("Failed to get database client", log.Error(err))
		return model.Application{}, fmt.Errorf("failed to get database client: %w", err)
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
		return model.Application{}, fmt.Errorf("failed to execute query: %w", err)
	}

	if len(results) == 0 {
		logger.Error("application not found")
		return model.Application{}, fmt.Errorf("application not found")
	}

	if len(results) != 1 {
		logger.Error("unexpected number of results")
		return model.Application{}, fmt.Errorf("unexpected number of results: %d", len(results))
	}

	row := results[0]

	application, err := buildApplicationFromResultRow(row)
	if err != nil {
		logger.Error("failed to build application from result row")
		return model.Application{}, fmt.Errorf("failed to build application from result row: %w", err)
	}
	return application, nil
}

// UpdateApplication updates an existing application in the database.
func UpdateApplication(app *model.Application) error {
	queries := []func(tx dbmodel.TxInterface) error{
		func(tx dbmodel.TxInterface) error {
			_, err := tx.Exec(QueryUpdateApplicationByAppID.Query, app.ID, app.OrgID, app.Name, app.Description,
				app.AuthFlowGraphID, app.RegistrationFlowGraphID)
			return err
		},
		func(tx dbmodel.TxInterface) error {
			_, err := tx.Exec(QueryUpdateOAuthApplicationByAppID.Query, app.ID, app.ClientID, app.ClientSecret,
				strings.Join(app.CallbackURLs, ","), strings.Join(app.SupportedGrantTypes, ","))
			return err
		},
	}

	return executeTransaction(queries)
}

// DeleteApplication deletes an application from the database by its ID.
func DeleteApplication(id string) error {
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

// buildApplicationFromResultRow constructs an Application object from a database result row.
func buildApplicationFromResultRow(row map[string]interface{}) (model.Application, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationStore"))

	appID, ok := row["app_id"].(string)
	if !ok {
		logger.Error("failed to parse app_id as string")
		return model.Application{}, fmt.Errorf("failed to parse app_id as string")
	}

	orgID, ok := row["org_id"].(string)
	if !ok {
		logger.Error("failed to parse org_id as string")
		return model.Application{}, fmt.Errorf("failed to parse org_id as string")
	}

	appName, ok := row["app_name"].(string)
	if !ok {
		logger.Error("failed to parse app_name as string")
		return model.Application{}, fmt.Errorf("failed to parse app_name as string")
	}

	description, ok := row["description"].(string)
	if !ok {
		logger.Error("failed to parse description as string")
		return model.Application{}, fmt.Errorf("failed to parse description as string")
	}

	authFlowGraphID, ok := row["auth_flow_graph_id"].(string)
	if !ok {
		logger.Error("failed to parse auth_flow_graph_id as string")
		return model.Application{}, fmt.Errorf("failed to parse auth_flow_graph_id as string")
	}

	regisFlowGraphID, ok := row["registration_flow_graph_id"].(string)
	if !ok {
		logger.Error("failed to parse registration_flow_graph_id as string")
		return model.Application{}, fmt.Errorf("failed to parse registration_flow_graph_id as string")
	}

	clientID, ok := row["consumer_key"].(string)
	if !ok {
		logger.Error("failed to parse consumer_key as string")
		return model.Application{}, fmt.Errorf("failed to parse consumer_key as string")
	}

	var redirectURIs []string
	if row["callback_uris"] != nil {
		if uris, ok := row["callback_uris"].(string); ok {
			redirectURIs = utils.ParseStringArray(uris)
		} else {
			logger.Error("failed to parse callback_uris as string")
			return model.Application{}, fmt.Errorf("failed to parse callback_uris as string")
		}
	}

	var allowedGrantTypes []string
	if row["grant_types"] != nil {
		if grants, ok := row["grant_types"].(string); ok {
			allowedGrantTypes = utils.ParseStringArray(grants)
		} else {
			logger.Error("failed to parse grant_types as string")
			return model.Application{}, fmt.Errorf("failed to parse grant_types as string")
		}
	}

	application := model.Application{
		ID:                      appID,
		OrgID:                   orgID,
		Name:                    appName,
		Description:             description,
		AuthFlowGraphID:         authFlowGraphID,
		RegistrationFlowGraphID: regisFlowGraphID,
		ClientID:                clientID,
		ClientSecret:            "***",
		CallbackURLs:            redirectURIs,
		SupportedGrantTypes:     allowedGrantTypes,
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
