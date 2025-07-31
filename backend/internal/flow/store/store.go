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

// Package store provides the implementation for flow context persistence operations.
package store

import (
	"errors"
	"fmt"

	"github.com/asgardeo/thunder/internal/flow/model"
	dbmodel "github.com/asgardeo/thunder/internal/system/database/model"
	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/log"
)

const loggerComponentName = "FlowStore"

// FlowStoreInterface defines the methods for flow context storage operations.
type FlowStoreInterface interface {
	StoreFlowContext(ctx model.EngineContext) error
	GetFlowContext(flowID string) (*FlowContextWithUserDataDB, error)
	UpdateFlowContext(ctx model.EngineContext) error
	DeleteFlowContext(flowID string) error
}

// FlowStore implements the FlowStoreInterface for managing flow contexts.
type FlowStore struct {
	DBProvider provider.DBProviderInterface
}

// NewFlowStore creates a new instance of FlowStore.
func NewFlowStore() FlowStoreInterface {
	return &FlowStore{
		DBProvider: provider.NewDBProvider(),
	}
}

// StoreFlowContext stores the complete flow context in the database.
func (s *FlowStore) StoreFlowContext(ctx model.EngineContext) error {
	// Convert engine context to database model
	dbModel, err := FromEngineContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to convert engine context to database model: %w", err)
	}

	queries := []func(tx dbmodel.TxInterface) error{
		func(tx dbmodel.TxInterface) error {
			_, err := tx.Exec(QueryCreateFlowContext.Query, dbModel.FlowID, dbModel.AppID,
				dbModel.CurrentNodeID, dbModel.CurrentActionID, dbModel.GraphID, dbModel.RuntimeData)
			return err
		},
		func(tx dbmodel.TxInterface) error {
			_, err := tx.Exec(QueryCreateFlowUserData.Query, dbModel.FlowID, dbModel.IsAuthenticated, dbModel.UserID,
				dbModel.UserInputs, dbModel.UserAttributes)
			return err
		},
	}

	return s.executeTransaction(queries)
}

// GetFlowContext retrieves the flow context from the database.
func (s *FlowStore) GetFlowContext(flowID string) (*FlowContextWithUserDataDB, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	dbClient, err := s.DBProvider.GetDBClient("runtime")
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}
	defer func() {
		if closeErr := dbClient.Close(); closeErr != nil {
			logger.Error("Failed to close database client", log.Error(closeErr))
		}
	}()

	results, err := dbClient.Query(QueryGetFlowContextWithUserData, flowID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	if len(results) == 0 {
		return nil, nil
	}

	if len(results) != 1 {
		return nil, fmt.Errorf("unexpected number of results: %d", len(results))
	}

	row := results[0]
	return buildFlowContextFromResultRow(row)
}

// UpdateFlowContext updates the flow context in the database.
func (s *FlowStore) UpdateFlowContext(ctx model.EngineContext) error {
	// Convert engine context to database model
	dbModel, err := FromEngineContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to convert engine context to database model: %w", err)
	}

	queries := []func(tx dbmodel.TxInterface) error{
		func(tx dbmodel.TxInterface) error {
			_, err := tx.Exec(QueryUpdateFlowContext.Query, dbModel.FlowID,
				dbModel.CurrentNodeID, dbModel.CurrentActionID, dbModel.RuntimeData)
			return err
		},
		func(tx dbmodel.TxInterface) error {
			_, err := tx.Exec(QueryUpdateFlowUserData.Query, dbModel.FlowID, dbModel.IsAuthenticated, dbModel.UserID,
				dbModel.UserInputs, dbModel.UserAttributes)
			return err
		},
	}

	return s.executeTransaction(queries)
}

// DeleteFlowContext removes the flow context from the database.
func (s *FlowStore) DeleteFlowContext(flowID string) error {
	queries := []func(tx dbmodel.TxInterface) error{
		func(tx dbmodel.TxInterface) error {
			_, err := tx.Exec(QueryDeleteFlowUserData.Query, flowID)
			return err
		},
		func(tx dbmodel.TxInterface) error {
			_, err := tx.Exec(QueryDeleteFlowContext.Query, flowID)
			return err
		},
	}

	return s.executeTransaction(queries)
}

// executeTransaction is a helper function to handle database transactions.
func (s *FlowStore) executeTransaction(queries []func(tx dbmodel.TxInterface) error) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	dbClient, err := s.DBProvider.GetDBClient("runtime")
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}
	defer func() {
		if closeErr := dbClient.Close(); closeErr != nil {
			logger.Error("Failed to close database client", log.Error(closeErr))
		}
	}()

	tx, err := dbClient.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	for _, query := range queries {
		if err := query(tx); err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return fmt.Errorf("failed to rollback transaction: %w", rollbackErr)
			}
			return fmt.Errorf("transaction failed: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// buildFlowContextFromResultRow builds a FlowContextWithUserDataDB from a database result row.
func buildFlowContextFromResultRow(row map[string]interface{}) (*FlowContextWithUserDataDB, error) {
	// Parse required fields
	flowID, ok := row["flow_id"].(string)
	if !ok {
		return nil, errors.New("failed to parse flow_id as string")
	}

	appID, ok := row["app_id"].(string)
	if !ok {
		return nil, errors.New("failed to parse app_id as string")
	}

	graphID, ok := row["graph_id"].(string)
	if !ok {
		return nil, errors.New("failed to parse graph_id as string")
	}

	// Parse optional fields
	currentNodeID := parseOptionalString(row["current_node_id"])
	currentActionID := parseOptionalString(row["current_action_id"])
	userID := parseOptionalString(row["user_id"])
	userInputs := parseOptionalString(row["user_inputs"])
	runtimeData := parseOptionalString(row["runtime_data"])
	userAttributes := parseOptionalString(row["user_attributes"])

	// Parse boolean field with type conversion support
	isAuthenticated := parseBoolean(row["is_authenticated"])

	return &FlowContextWithUserDataDB{
		FlowID:          flowID,
		AppID:           appID,
		CurrentNodeID:   currentNodeID,
		CurrentActionID: currentActionID,
		GraphID:         graphID,
		RuntimeData:     runtimeData,
		IsAuthenticated: isAuthenticated,
		UserID:          userID,
		UserInputs:      userInputs,
		UserAttributes:  userAttributes,
	}, nil
}

// parseOptionalString safely parses an optional string field from the database row
func parseOptionalString(value interface{}) *string {
	if value == nil {
		return nil
	}
	if str, ok := value.(string); ok {
		return &str
	}
	return nil
}

// parseBoolean safely parses a boolean field from the database row with type conversion support
func parseBoolean(value interface{}) bool {
	if value == nil {
		return false
	}

	if boolVal, ok := value.(bool); ok {
		return boolVal
	}

	if intVal, ok := value.(int64); ok {
		return intVal != 0
	}

	return false
}
