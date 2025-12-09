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

package flowexec

import (
	"errors"
	"fmt"

	"github.com/asgardeo/thunder/internal/system/config"
	dbmodel "github.com/asgardeo/thunder/internal/system/database/model"
	"github.com/asgardeo/thunder/internal/system/database/provider"
)

// flowStoreInterface defines the methods for flow context storage operations.
type flowStoreInterface interface {
	StoreFlowContext(ctx EngineContext) error
	GetFlowContext(flowID string) (*FlowContextWithUserDataDB, error)
	UpdateFlowContext(ctx EngineContext) error
	DeleteFlowContext(flowID string) error
}

// flowStore implements the FlowStoreInterface for managing flow contexts.
type flowStore struct {
	dbProvider   provider.DBProviderInterface
	deploymentID string
}

// newFlowStore creates a new instance of FlowStore.
func newFlowStore(dbProvider provider.DBProviderInterface) flowStoreInterface {
	return &flowStore{
		dbProvider:   dbProvider,
		deploymentID: config.GetThunderRuntime().Config.Server.Identifier,
	}
}

// StoreFlowContext stores the complete flow context in the database.
func (s *flowStore) StoreFlowContext(ctx EngineContext) error {
	// Convert engine context to database model
	dbModel, err := FromEngineContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to convert engine context to database model: %w", err)
	}

	queries := []func(tx dbmodel.TxInterface) error{
		func(tx dbmodel.TxInterface) error {
			_, err := tx.Exec(QueryCreateFlowContext.Query,
				dbModel.FlowID, dbModel.AppID, dbModel.Verbose,
				dbModel.CurrentNodeID, dbModel.CurrentAction, dbModel.GraphID,
				dbModel.RuntimeData, dbModel.ExecutionHistory, s.deploymentID)
			return err
		},
		func(tx dbmodel.TxInterface) error {
			_, err := tx.Exec(QueryCreateFlowUserData.Query, dbModel.FlowID,
				dbModel.IsAuthenticated, dbModel.UserID, dbModel.OrganizationUnitID,
				dbModel.UserType, dbModel.UserInputs, dbModel.UserAttributes, s.deploymentID)
			return err
		},
	}

	return s.executeTransaction(queries)
}

// GetFlowContext retrieves the flow context from the database.
func (s *flowStore) GetFlowContext(flowID string) (*FlowContextWithUserDataDB, error) {
	dbClient, err := s.dbProvider.GetRuntimeDBClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(QueryGetFlowContextWithUserData, flowID, s.deploymentID)
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
	return s.buildFlowContextFromResultRow(row)
}

// UpdateFlowContext updates the flow context in the database.
func (s *flowStore) UpdateFlowContext(ctx EngineContext) error {
	// Convert engine context to database model
	dbModel, err := FromEngineContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to convert engine context to database model: %w", err)
	}

	queries := []func(tx dbmodel.TxInterface) error{
		func(tx dbmodel.TxInterface) error {
			_, err := tx.Exec(QueryUpdateFlowContext.Query, dbModel.FlowID,
				dbModel.CurrentNodeID, dbModel.CurrentAction, dbModel.RuntimeData, dbModel.ExecutionHistory,
				s.deploymentID)
			return err
		},
		func(tx dbmodel.TxInterface) error {
			_, err := tx.Exec(QueryUpdateFlowUserData.Query, dbModel.FlowID, dbModel.IsAuthenticated,
				dbModel.UserID, dbModel.OrganizationUnitID, dbModel.UserType,
				dbModel.UserInputs, dbModel.UserAttributes, s.deploymentID)
			return err
		},
	}

	return s.executeTransaction(queries)
}

// DeleteFlowContext removes the flow context from the database.
func (s *flowStore) DeleteFlowContext(flowID string) error {
	queries := []func(tx dbmodel.TxInterface) error{
		func(tx dbmodel.TxInterface) error {
			_, err := tx.Exec(QueryDeleteFlowUserData.Query, flowID, s.deploymentID)
			return err
		},
		func(tx dbmodel.TxInterface) error {
			_, err := tx.Exec(QueryDeleteFlowContext.Query, flowID, s.deploymentID)
			return err
		},
	}

	return s.executeTransaction(queries)
}

// executeTransaction is a helper function to handle database transactions.
func (s *flowStore) executeTransaction(queries []func(tx dbmodel.TxInterface) error) error {
	dbClient, err := s.dbProvider.GetRuntimeDBClient()
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
func (s *flowStore) buildFlowContextFromResultRow(row map[string]interface{}) (*FlowContextWithUserDataDB, error) {
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
	currentNodeID := s.parseOptionalString(row["current_node_id"])
	currentAction := s.parseOptionalString(row["current_action"])
	userID := s.parseOptionalString(row["user_id"])
	organizationUnitID := s.parseOptionalString(row["ou_id"])
	userType := s.parseOptionalString(row["user_type"])
	userInputs := s.parseOptionalString(row["user_inputs"])
	runtimeData := s.parseOptionalString(row["runtime_data"])
	userAttributes := s.parseOptionalString(row["user_attributes"])
	executionHistory := s.parseOptionalString(row["execution_history"])

	// Parse boolean fields with type conversion support
	isAuthenticated := s.parseBoolean(row["is_authenticated"])
	verbose := s.parseBoolean(row["verbose"])

	return &FlowContextWithUserDataDB{
		FlowID:             flowID,
		AppID:              appID,
		CurrentNodeID:      currentNodeID,
		CurrentAction:      currentAction,
		GraphID:            graphID,
		RuntimeData:        runtimeData,
		Verbose:            verbose,
		IsAuthenticated:    isAuthenticated,
		UserID:             userID,
		OrganizationUnitID: organizationUnitID,
		UserType:           userType,
		UserInputs:         userInputs,
		UserAttributes:     userAttributes,
		ExecutionHistory:   executionHistory,
	}, nil
}

// parseOptionalString safely parses an optional string field from the database row
func (s *flowStore) parseOptionalString(value interface{}) *string {
	if value == nil {
		return nil
	}
	if str, ok := value.(string); ok {
		return &str
	}
	// Handle []byte type (PostgreSQL may return TEXT/JSON as []byte)
	if bytes, ok := value.([]byte); ok {
		str := string(bytes)
		return &str
	}
	return nil
}

// parseBoolean safely parses a boolean field from the database row with type conversion support
func (s *flowStore) parseBoolean(value interface{}) bool {
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
