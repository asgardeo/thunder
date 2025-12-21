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

package flowmgt

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/database/model"
	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/log"
)

// Database column names
const (
	colFlowID        = "flow_id"
	colHandle        = "handle"
	colName          = "name"
	colFlowType      = "flow_type"
	colActiveVersion = "active_version"
	colNodes         = "nodes"
	colCreatedAt     = "created_at"
	colUpdatedAt     = "updated_at"
	colVersion       = "version"
	colCount         = "count"
)

// flowStoreInterface defines the interface for flow store operations.
type flowStoreInterface interface {
	ListFlows(limit, offset int, flowType string) ([]BasicFlowDefinition, int, error)
	CreateFlow(flowID string, flow *FlowDefinition) (*CompleteFlowDefinition, error)
	GetFlowByID(flowID string) (*CompleteFlowDefinition, error)
	GetFlowByHandle(handle string, flowType common.FlowType) (*CompleteFlowDefinition, error)
	UpdateFlow(flowID string, flow *FlowDefinition) (*CompleteFlowDefinition, error)
	DeleteFlow(flowID string) error
	ListFlowVersions(flowID string) ([]BasicFlowVersion, error)
	GetFlowVersion(flowID string, version int) (*FlowVersion, error)
	RestoreFlowVersion(flowID string, version int) (*CompleteFlowDefinition, error)
	IsFlowExists(flowID string) (bool, error)
	IsFlowExistsByHandle(handle string, flowType common.FlowType) (bool, error)
}

// flowStore is the default implementation of flowStoreInterface.
type flowStore struct {
	dbProvider        provider.DBProviderInterface
	deploymentID      string
	maxVersionHistory int
	logger            *log.Logger
}

// newFlowStore creates a new instance of flowStore.
func newFlowStore() flowStoreInterface {
	return &flowStore{
		dbProvider:        provider.GetDBProvider(),
		deploymentID:      config.GetThunderRuntime().Config.Server.Identifier,
		maxVersionHistory: getMaxVersionHistory(),
		logger:            log.GetLogger().With(log.String(log.LoggerKeyComponentName, "FlowStore")),
	}
}

// ListFlows retrieves a paginated list of flow definitions with optional filtering by flow type.
func (s *flowStore) ListFlows(limit, offset int, flowType string) ([]BasicFlowDefinition, int, error) {
	var flows []BasicFlowDefinition
	var totalCount int

	err := s.withDBClient(func(dbClient provider.DBClientInterface) error {
		var countResults, results []map[string]interface{}
		var err error

		if flowType != "" {
			countResults, err = dbClient.Query(queryCountFlowsWithType, flowType, s.deploymentID)
			if err != nil {
				return fmt.Errorf("failed to count flows: %w", err)
			}

			results, err = dbClient.Query(queryListFlowsWithType, flowType, s.deploymentID, limit, offset)
			if err != nil {
				return fmt.Errorf("failed to list flows: %w", err)
			}
		} else {
			countResults, err = dbClient.Query(queryCountFlows, s.deploymentID)
			if err != nil {
				return fmt.Errorf("failed to count flows: %w", err)
			}

			results, err = dbClient.Query(queryListFlows, s.deploymentID, limit, offset)
			if err != nil {
				return fmt.Errorf("failed to list flows: %w", err)
			}
		}

		totalCount, err = s.parseCountResult(countResults)
		if err != nil {
			return err
		}

		flows = make([]BasicFlowDefinition, 0, len(results))
		for _, row := range results {
			flow, err := s.buildBasicFlowDefinitionFromRow(row)
			if err != nil {
				return fmt.Errorf("failed to build flow: %w", err)
			}
			flows = append(flows, flow)
		}

		return nil
	})

	if err != nil {
		return nil, 0, err
	}

	return flows, totalCount, nil
}

// CreateFlow creates a new flow definition with version 1.
func (s *flowStore) CreateFlow(flowID string, flow *FlowDefinition) (*CompleteFlowDefinition, error) {
	nodesJSON, err := json.Marshal(flow.Nodes)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal nodes: %w", err)
	}

	err = s.withTransaction(func(tx model.TxInterface) error {
		_, err := tx.Exec(queryCreateFlow, flowID, flow.Handle, flow.Name, flow.FlowType, int64(1), s.deploymentID)
		if err != nil {
			return fmt.Errorf("failed to create flow: %w", err)
		}

		internalID, err := s.getFlowInternalIDWithTx(tx, flowID)
		if err != nil {
			return err
		}

		_, err = tx.Exec(queryInsertFlowVersion, internalID, 1, string(nodesJSON), s.deploymentID)
		if err != nil {
			return fmt.Errorf("failed to create flow version: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.GetFlowByID(flowID)
}

// GetFlowByID retrieves the active version of a flow definition by its ID.
func (s *flowStore) GetFlowByID(flowID string) (*CompleteFlowDefinition, error) {
	var flow *CompleteFlowDefinition
	err := s.withDBClient(func(dbClient provider.DBClientInterface) error {
		results, err := dbClient.Query(queryGetFlow, flowID, s.deploymentID)
		if err != nil {
			return fmt.Errorf("failed to get flow: %w", err)
		}

		if len(results) == 0 {
			return errFlowNotFound
		}

		flow, err = s.buildCompleteFlowDefinitionFromRow(results[0])
		return err
	})

	return flow, err
}

// GetFlowByHandle retrieves a flow definition by handle and flow type.
func (s *flowStore) GetFlowByHandle(handle string, flowType common.FlowType) (*CompleteFlowDefinition, error) {
	var flow *CompleteFlowDefinition
	err := s.withDBClient(func(dbClient provider.DBClientInterface) error {
		results, err := dbClient.Query(queryGetFlowByHandle, handle, string(flowType), s.deploymentID)
		if err != nil {
			return fmt.Errorf("failed to get flow by handle: %w", err)
		}

		if len(results) == 0 {
			return errFlowNotFound
		}

		flow, err = s.buildCompleteFlowDefinitionFromRow(results[0])
		return err
	})

	return flow, err
}

// UpdateFlow updates a flow definition by creating a new version.
// Automatically deletes oldest versions if the count exceeds max_version_history.
func (s *flowStore) UpdateFlow(flowID string, flow *FlowDefinition) (*CompleteFlowDefinition, error) {
	nodesJSON, err := json.Marshal(flow.Nodes)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal nodes: %w", err)
	}

	err = s.withTransaction(func(tx model.TxInterface) error {
		flowResults, err := tx.Query(queryGetFlow, flowID, s.deploymentID)
		if err != nil {
			return fmt.Errorf("failed to get flow metadata: %w", err)
		}

		_, currentVersion, err := s.scanFlowMetadata(flowResults)
		if closeErr := flowResults.Close(); closeErr != nil {
			s.logger.Error("Failed to close flow results", log.Error(closeErr))
		}
		if err != nil {
			return errFlowNotFound
		}

		newVersion := int(currentVersion) + 1

		internalID, err := s.getFlowInternalIDWithTx(tx, flowID)
		if err != nil {
			return err
		}

		// Insert the new version first to ensure it succeeds before updating the flow
		if err := s.pushToVersionStack(tx, internalID, newVersion, string(nodesJSON)); err != nil {
			return err
		}

		_, err = tx.Exec(queryUpdateFlow, flowID, flow.Name, newVersion, s.deploymentID)
		if err != nil {
			return fmt.Errorf("failed to update flow: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.GetFlowByID(flowID)
}

// DeleteFlow deletes a flow definition and all its version history.
func (s *flowStore) DeleteFlow(flowID string) error {
	return s.withDBClient(func(dbClient provider.DBClientInterface) error {
		_, err := dbClient.Execute(queryDeleteFlow, flowID, s.deploymentID)
		if err != nil {
			return fmt.Errorf("failed to delete flow: %w", err)
		}
		return nil
	})
}

// IsFlowExists checks if a flow exists with a given flow ID.
func (s *flowStore) IsFlowExists(flowID string) (bool, error) {
	var exists bool
	err := s.withDBClient(func(dbClient provider.DBClientInterface) error {
		results, err := dbClient.Query(queryCheckFlowExistsByID, flowID, s.deploymentID)
		if err != nil {
			return fmt.Errorf("failed to check flow existence: %w", err)
		}

		exists = len(results) > 0
		return nil
	})

	return exists, err
}

// IsFlowExistsByHandle checks if a flow exists with the given handle and flow type.
func (s *flowStore) IsFlowExistsByHandle(handle string, flowType common.FlowType) (bool, error) {
	var exists bool
	err := s.withDBClient(func(dbClient provider.DBClientInterface) error {
		results, err := dbClient.Query(queryCheckFlowExistsByHandle, handle, string(flowType), s.deploymentID)
		if err != nil {
			return fmt.Errorf("failed to check flow existence by handle: %w", err)
		}

		exists = len(results) > 0
		return nil
	})

	return exists, err
}

// ListFlowVersions retrieves all versions of a flow definition.
func (s *flowStore) ListFlowVersions(flowID string) ([]BasicFlowVersion, error) {
	var versions []BasicFlowVersion

	err := s.withDBClient(func(dbClient provider.DBClientInterface) error {
		internalID, err := s.getFlowInternalID(dbClient, flowID)
		if err != nil {
			return err
		}

		results, err := dbClient.Query(queryListFlowVersions, internalID, s.deploymentID)
		if err != nil {
			return fmt.Errorf("failed to list flow versions: %w", err)
		}

		versions = make([]BasicFlowVersion, 0, len(results))
		for _, row := range results {
			version, err := s.buildBasicFlowVersionFromRow(row)
			if err != nil {
				return fmt.Errorf("failed to build flow version: %w", err)
			}
			versions = append(versions, version)
		}

		return nil
	})

	return versions, err
}

// GetFlowVersion retrieves a specific version of a flow definition.
func (s *flowStore) GetFlowVersion(flowID string, version int) (*FlowVersion, error) {
	var flowVersion *FlowVersion

	err := s.withDBClient(func(dbClient provider.DBClientInterface) error {
		results, err := dbClient.Query(queryGetFlowVersionWithMetadata, flowID, version, s.deploymentID)
		if err != nil {
			return fmt.Errorf("failed to get flow version: %w", err)
		}
		if len(results) == 0 {
			return errVersionNotFound
		}

		flowVersion, err = s.buildFlowVersionFromRow(results[0])
		return err
	})

	return flowVersion, err
}

// RestoreFlowVersion restores a specified version as the active version.
// This creates a new version by copying the configuration from the specified version.
// Automatically deletes oldest versions if the count exceeds max_version_history.
func (s *flowStore) RestoreFlowVersion(flowID string, version int) (*CompleteFlowDefinition, error) {
	err := s.withTransaction(func(tx model.TxInterface) error {
		flowResults, err := tx.Query(queryGetFlow, flowID, s.deploymentID)
		if err != nil {
			return fmt.Errorf("failed to get flow metadata: %w", err)
		}

		flowName, currentVersion, err := s.scanFlowMetadata(flowResults)
		if closeErr := flowResults.Close(); closeErr != nil {
			s.logger.Error("Failed to close flow results", log.Error(closeErr))
		}
		if err != nil {
			return errFlowNotFound
		}

		internalID, err := s.getFlowInternalIDWithTx(tx, flowID)
		if err != nil {
			return err
		}

		versionResults, err := tx.Query(queryGetFlowVersion, internalID, version, s.deploymentID)
		if err != nil {
			return fmt.Errorf("failed to get version to restore: %w", err)
		}

		nodesJSON, err := s.scanFlowVersion(versionResults)
		if closeErr := versionResults.Close(); closeErr != nil {
			s.logger.Error("Failed to close version results", log.Error(closeErr))
		}
		if err != nil {
			return errVersionNotFound
		}

		newVersion := int(currentVersion) + 1

		// Insert the new version first to ensure it succeeds before updating the flow
		if err := s.pushToVersionStack(tx, internalID, newVersion, nodesJSON); err != nil {
			return err
		}

		_, err = tx.Exec(queryUpdateFlow, flowID, flowName, newVersion, s.deploymentID)
		if err != nil {
			return fmt.Errorf("failed to update flow: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.GetFlowByID(flowID)
}

// pushToVersionStack adds a new version to the version history and removes the oldest version
// if the count exceeds max_version_history.
func (s *flowStore) pushToVersionStack(tx model.TxInterface,
	flowInternalID int64, version int, nodesJSON string) error {
	_, err := tx.Exec(queryInsertFlowVersion, flowInternalID, version, nodesJSON, s.deploymentID)
	if err != nil {
		return fmt.Errorf("failed to insert flow version: %w", err)
	}

	countResults, err := tx.Query(queryCountFlowVersions, flowInternalID, s.deploymentID)
	if err != nil {
		return fmt.Errorf("failed to count versions: %w", err)
	}

	versionCount, err := s.parseCountFromRows(countResults)
	if closeErr := countResults.Close(); closeErr != nil {
		s.logger.Error("Failed to close count results", log.Error(closeErr))
	}
	if err != nil {
		return err
	}

	if versionCount > s.maxVersionHistory {
		if _, err := tx.Exec(queryDeleteOldestVersion, flowInternalID, s.deploymentID); err != nil {
			return fmt.Errorf("failed to delete oldest version: %w", err)
		}
	}

	return nil
}

// getFlowInternalIDWithTx retrieves the internal ID of a flow by its flow ID within a transaction.
func (s *flowStore) getFlowInternalIDWithTx(tx model.TxInterface, flowID string) (int64, error) {
	results, err := tx.Query(queryGetFlowInternalID, flowID, s.deploymentID)
	if err != nil {
		return 0, fmt.Errorf("failed to get flow internal ID: %w", err)
	}

	if !results.Next() {
		_ = results.Close()
		return 0, errFlowNotFound
	}

	var internalID int64
	if err := results.Scan(&internalID); err != nil {
		_ = results.Close()
		return 0, fmt.Errorf("failed to scan internal ID: %w", err)
	}
	if closeErr := results.Close(); closeErr != nil {
		s.logger.Error("Failed to close internal ID results", log.Error(closeErr))
	}

	return internalID, nil
}

// getFlowInternalID retrieves the internal ID of a flow by its flow ID.
func (s *flowStore) getFlowInternalID(dbClient provider.DBClientInterface, flowID string) (int64, error) {
	results, err := dbClient.Query(queryGetFlowInternalID, flowID, s.deploymentID)
	if err != nil {
		return 0, fmt.Errorf("failed to get flow internal ID: %w", err)
	}

	if len(results) == 0 {
		return 0, errFlowNotFound
	}

	internalIDVal, ok := results[0]["id"]
	if !ok {
		return 0, fmt.Errorf("internal ID field not found in result")
	}

	internalID, ok := internalIDVal.(int64)
	if !ok {
		return 0, fmt.Errorf("unexpected internal ID type: %T", internalIDVal)
	}

	return internalID, nil
}

// getConfigDBClient retrieves the configuration database client.
func (s *flowStore) getConfigDBClient() (provider.DBClientInterface, error) {
	dbClient, err := s.dbProvider.GetConfigDBClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}
	return dbClient, nil
}

// withDBClient executes a function with a DB client, handling client retrieval errors.
func (s *flowStore) withDBClient(fn func(provider.DBClientInterface) error) error {
	dbClient, err := s.getConfigDBClient()
	if err != nil {
		return err
	}
	return fn(dbClient)
}

// withTransaction executes a function within a database transaction.
func (s *flowStore) withTransaction(fn func(model.TxInterface) error) error {
	return s.withDBClient(func(dbClient provider.DBClientInterface) error {
		tx, err := dbClient.BeginTx()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		if err := fn(tx); err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				err = errors.Join(err, fmt.Errorf("failed to rollback transaction: %w", rollbackErr))
			}
			return err
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		return nil
	})
}

// parseCountResult parses a count result from database query.
func (s *flowStore) parseCountResult(results []map[string]interface{}) (int, error) {
	if len(results) == 0 {
		return 0, nil
	}

	countVal, ok := results[0][colCount]
	if !ok {
		return 0, fmt.Errorf("count field not found in result")
	}

	switch v := countVal.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("unexpected count type: %T", countVal)
	}
}

// parseCountFromRows parses the count result from *sql.Rows.
func (s *flowStore) parseCountFromRows(rows *sql.Rows) (int, error) {
	if !rows.Next() {
		return 0, fmt.Errorf("no count result returned")
	}

	var count int64
	if err := rows.Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to scan count: %w", err)
	}

	return int(count), nil
}

// scanFlowMetadata scans a single row from FLOW table into individual fields.
func (s *flowStore) scanFlowMetadata(rows *sql.Rows) (flowName string, activeVersion int64, err error) {
	if !rows.Next() {
		return "", 0, fmt.Errorf("no flow found")
	}

	var flowID, handle, flowType, nodes, createdAt, updatedAt string
	err = rows.Scan(&flowID, &handle, &flowName, &flowType, &activeVersion, &nodes, &createdAt, &updatedAt)
	if err != nil {
		return "", 0, fmt.Errorf("failed to scan flow metadata: %w", err)
	}

	return flowName, activeVersion, nil
}

// scanFlowVersion scans a single row from FLOW_VERSION table into individual fields.
func (s *flowStore) scanFlowVersion(rows *sql.Rows) (nodes string, err error) {
	if !rows.Next() {
		return "", fmt.Errorf("no version found")
	}

	var version int64
	var createdAt string
	err = rows.Scan(&version, &nodes, &createdAt)
	if err != nil {
		return "", fmt.Errorf("failed to scan version data: %w", err)
	}

	return nodes, nil
}

// getString safely extracts a string value from a database row.
// Handles both string (SQLite) and []byte (PostgreSQL) types.
func (s *flowStore) getString(row map[string]interface{}, key string) (string, error) {
	val := row[key]
	switch v := val.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	default:
		return "", fmt.Errorf("%s field is missing or invalid", key)
	}
}

// getTimestamp safely extracts a timestamp value from a database row.
// Handles both string (SQLite) and time.Time (PostgreSQL) types.
func (s *flowStore) getTimestamp(row map[string]interface{}, key string) (string, error) {
	val := row[key]
	switch v := val.(type) {
	case string:
		return v, nil
	case time.Time:
		// Convert time.Time to RFC3339 format for consistency
		return v.Format(time.RFC3339), nil
	default:
		return "", fmt.Errorf("%s field is missing or invalid", key)
	}
}

// getInt64 safely extracts an int64 value from a database row.
func (s *flowStore) getInt64(row map[string]interface{}, key string) (int64, error) {
	if val, ok := row[key].(int64); ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s field is missing or invalid", key)
}

// buildBasicFlowDefinitionFromRow builds a BasicFlowDefinition from a database row.
func (s *flowStore) buildBasicFlowDefinitionFromRow(row map[string]interface{}) (
	BasicFlowDefinition, error) {
	flowID, err := s.getString(row, colFlowID)
	if err != nil {
		return BasicFlowDefinition{}, err
	}

	handle, err := s.getString(row, colHandle)
	if err != nil {
		return BasicFlowDefinition{}, err
	}

	name, err := s.getString(row, colName)
	if err != nil {
		return BasicFlowDefinition{}, err
	}

	flowTypeStr, err := s.getString(row, colFlowType)
	if err != nil {
		return BasicFlowDefinition{}, err
	}

	activeVersion, err := s.getInt64(row, colActiveVersion)
	if err != nil {
		return BasicFlowDefinition{}, err
	}

	createdAt, err := s.getTimestamp(row, colCreatedAt)
	if err != nil {
		return BasicFlowDefinition{}, err
	}

	updatedAt, err := s.getTimestamp(row, colUpdatedAt)
	if err != nil {
		return BasicFlowDefinition{}, err
	}

	return BasicFlowDefinition{
		ID:            flowID,
		Handle:        handle,
		Name:          name,
		FlowType:      common.FlowType(flowTypeStr),
		ActiveVersion: int(activeVersion),
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}, nil
}

// buildCompleteFlowDefinitionFromRow builds a CompleteFlowDefinition from a database row.
func (s *flowStore) buildCompleteFlowDefinitionFromRow(row map[string]interface{}) (
	*CompleteFlowDefinition, error) {
	flowID, err := s.getString(row, colFlowID)
	if err != nil {
		return nil, err
	}

	handle, err := s.getString(row, colHandle)
	if err != nil {
		return nil, err
	}

	name, err := s.getString(row, colName)
	if err != nil {
		return nil, err
	}

	flowTypeStr, err := s.getString(row, colFlowType)
	if err != nil {
		return nil, err
	}

	activeVersion, err := s.getInt64(row, colActiveVersion)
	if err != nil {
		return nil, err
	}

	createdAt, err := s.getTimestamp(row, colCreatedAt)
	if err != nil {
		return nil, err
	}

	updatedAt, err := s.getTimestamp(row, colUpdatedAt)
	if err != nil {
		return nil, err
	}

	nodesJSON, err := s.getString(row, colNodes)
	if err != nil {
		return nil, err
	}

	flow := &CompleteFlowDefinition{
		ID:            flowID,
		Handle:        handle,
		Name:          name,
		FlowType:      common.FlowType(flowTypeStr),
		ActiveVersion: int(activeVersion),
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}

	if err := json.Unmarshal([]byte(nodesJSON), &flow.Nodes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal nodes: %w", err)
	}

	return flow, nil
}

// buildBasicFlowVersionFromRow builds a BasicFlowVersion from a database row.
func (s *flowStore) buildBasicFlowVersionFromRow(row map[string]interface{}) (BasicFlowVersion, error) {
	version, err := s.getInt64(row, colVersion)
	if err != nil {
		return BasicFlowVersion{}, err
	}

	createdAt, err := s.getTimestamp(row, colCreatedAt)
	if err != nil {
		return BasicFlowVersion{}, err
	}

	activeVersion, err := s.getInt64(row, colActiveVersion)
	if err != nil {
		return BasicFlowVersion{}, err
	}

	return BasicFlowVersion{
		Version:   int(version),
		CreatedAt: createdAt,
		IsActive:  int(version) == int(activeVersion),
	}, nil
}

// buildFlowVersionFromRow builds a FlowVersion from a single joined database row.
func (s *flowStore) buildFlowVersionFromRow(row map[string]interface{}) (*FlowVersion, error) {
	flowID, err := s.getString(row, colFlowID)
	if err != nil {
		return nil, err
	}

	handle, err := s.getString(row, colHandle)
	if err != nil {
		return nil, err
	}

	name, err := s.getString(row, colName)
	if err != nil {
		return nil, err
	}

	flowTypeStr, err := s.getString(row, colFlowType)
	if err != nil {
		return nil, err
	}

	version, err := s.getInt64(row, colVersion)
	if err != nil {
		return nil, err
	}

	createdAt, err := s.getTimestamp(row, colCreatedAt)
	if err != nil {
		return nil, err
	}

	activeVersion, err := s.getInt64(row, colActiveVersion)
	if err != nil {
		return nil, err
	}

	nodesJSON, err := s.getString(row, colNodes)
	if err != nil {
		return nil, err
	}

	flowVersion := &FlowVersion{
		ID:        flowID,
		Handle:    handle,
		Name:      name,
		FlowType:  flowTypeStr,
		Version:   int(version),
		IsActive:  int(version) == int(activeVersion),
		CreatedAt: createdAt,
	}

	if err := json.Unmarshal([]byte(nodesJSON), &flowVersion.Nodes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal nodes: %w", err)
	}

	return flowVersion, nil
}

// getMaxVersionHistory retrieves the maximum version history size from configuration.
// If not set or invalid, returns the default value.
func getMaxVersionHistory() int {
	flowConfig := config.GetThunderRuntime().Config.Flow
	if flowConfig.MaxVersionHistory <= 0 {
		return defaultVersionHistory
	}
	if flowConfig.MaxVersionHistory > maxAllowedVersionHistory {
		return maxAllowedVersionHistory
	}

	return flowConfig.MaxVersionHistory
}
