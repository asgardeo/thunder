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

package layoutmgt

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/log"
)

const storeLoggerComponentName = "LayoutMgtStore"

var errLayoutNotFound = errors.New("layout not found")

// layoutMgtStoreInterface defines the interface for layout management store operations.
type layoutMgtStoreInterface interface {
	GetLayoutListCount() (int, error)
	GetLayoutList(limit, offset int) ([]Layout, error)
	CreateLayout(id string, layout CreateLayoutRequest) error
	GetLayout(id string) (Layout, error)
	IsLayoutExist(id string) (bool, error)
	UpdateLayout(id string, layout UpdateLayoutRequest) error
	DeleteLayout(id string) error
	GetApplicationsCountByLayoutID(id string) (int, error)
}

// layoutMgtStore is the default implementation of layoutMgtStoreInterface.
type layoutMgtStore struct {
	dbProvider   provider.DBProviderInterface
	deploymentID string
}

// newLayoutMgtStore creates a new instance of layoutMgtStore.
func newLayoutMgtStore() layoutMgtStoreInterface {
	return &layoutMgtStore{
		dbProvider:   provider.GetDBProvider(),
		deploymentID: config.GetThunderRuntime().Config.Server.Identifier,
	}
}

// GetLayoutListCount retrieves the total count of layout configurations.
func (s *layoutMgtStore) GetLayoutListCount() (int, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return 0, err
	}

	countResults, err := dbClient.Query(queryGetLayoutListCount, s.deploymentID)
	if err != nil {
		return 0, fmt.Errorf("failed to execute count query: %w", err)
	}

	return parseCountResult(countResults)
}

// GetLayoutList retrieves layout configurations with pagination.
func (s *layoutMgtStore) GetLayoutList(limit, offset int) ([]Layout, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return nil, err
	}

	results, err := dbClient.Query(queryGetLayoutList, limit, offset, s.deploymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute layout list query: %w", err)
	}

	layouts := make([]Layout, 0)
	for _, row := range results {
		layout, err := buildLayoutListItemFromResultRow(row)
		if err != nil {
			return nil, fmt.Errorf("failed to build layout from result row: %w", err)
		}
		layouts = append(layouts, layout)
	}

	return layouts, nil
}

// CreateLayout creates a new layout configuration in the database.
func (s *layoutMgtStore) CreateLayout(id string, layout CreateLayoutRequest) error {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return err
	}

	preferencesJSON, err := json.Marshal(layout.Preferences)
	if err != nil {
		return fmt.Errorf("failed to marshal preferences: %w", err)
	}

	_, err = dbClient.Execute(queryCreateLayout, id, layout.DisplayName, layout.Description, preferencesJSON, s.deploymentID)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}

// GetLayout retrieves a layout configuration by its id.
func (s *layoutMgtStore) GetLayout(id string) (Layout, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return Layout{}, err
	}

	results, err := dbClient.Query(queryGetLayoutByID, id, s.deploymentID)
	if err != nil {
		return Layout{}, fmt.Errorf("failed to execute query: %w", err)
	}

	if len(results) == 0 {
		return Layout{}, errLayoutNotFound
	}

	if len(results) != 1 {
		return Layout{}, fmt.Errorf("unexpected number of results: %d", len(results))
	}

	return buildLayoutFromResultRow(results[0])
}

// IsLayoutExist checks if a layout configuration exists by its ID.
func (s *layoutMgtStore) IsLayoutExist(id string) (bool, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return false, err
	}

	results, err := dbClient.Query(queryCheckLayoutExists, id, s.deploymentID)
	if err != nil {
		return false, fmt.Errorf("failed to check layout existence: %w", err)
	}

	if len(results) == 0 {
		return false, nil
	}

	count, err := parseCountResult(results)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// UpdateLayout updates a layout configuration.
func (s *layoutMgtStore) UpdateLayout(id string, layout UpdateLayoutRequest) error {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return err
	}

	preferencesJSON, err := json.Marshal(layout.Preferences)
	if err != nil {
		return fmt.Errorf("failed to marshal preferences: %w", err)
	}

	_, err = dbClient.Execute(queryUpdateLayout, layout.DisplayName, layout.Description, preferencesJSON, id, s.deploymentID)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}

// DeleteLayout deletes a layout configuration.
func (s *layoutMgtStore) DeleteLayout(id string) error {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return err
	}

	_, err = dbClient.Execute(queryDeleteLayout, id, s.deploymentID)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}

// GetApplicationsCountByLayoutID returns the count of applications using a specific layout.
func (s *layoutMgtStore) GetApplicationsCountByLayoutID(id string) (int, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return 0, err
	}

	results, err := dbClient.Query(queryGetApplicationsCountByLayoutID, id, s.deploymentID)
	if err != nil {
		return 0, fmt.Errorf("failed to get applications count: %w", err)
	}

	return parseCountResult(results)
}

// getIdentityDBClient retrieves the identity database client.
func (s *layoutMgtStore) getIdentityDBClient() (provider.DBClientInterface, error) {
	dbClient, err := s.dbProvider.GetConfigDBClient()
	if err != nil {
		logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, storeLoggerComponentName))
		logger.Error("Failed to get identity database client", log.Error(err))
		return nil, fmt.Errorf("failed to get identity database client: %w", err)
	}
	return dbClient, nil
}

// parseCountResult parses count query results.
func parseCountResult(results []map[string]interface{}) (int, error) {
	if len(results) == 0 {
		return 0, fmt.Errorf("no results returned from count query")
	}

	totalInterface, exists := results[0]["total"]
	if !exists {
		return 0, fmt.Errorf("total field not found in result")
	}

	var total int
	switch v := totalInterface.(type) {
	case int64:
		total = int(v)
	case int:
		total = v
	default:
		return 0, fmt.Errorf("unexpected type for total: %T", totalInterface)
	}

	return total, nil
}

// buildLayoutListItemFromResultRow builds a Layout from a database result row (list view).
func buildLayoutListItemFromResultRow(row map[string]interface{}) (Layout, error) {
	id, ok := row["layout_id"].(string)
	if !ok {
		return Layout{}, fmt.Errorf("layout_id not found or invalid type")
	}

	displayName, ok := row["display_name"].(string)
	if !ok {
		return Layout{}, fmt.Errorf("display_name not found or invalid type")
	}

	description := ""
	if descInterface, ok := row["description"]; ok && descInterface != nil {
		description, _ = descInterface.(string)
	}

	return Layout{
		ID:          id,
		DisplayName: displayName,
		Description: description,
	}, nil
}

// buildLayoutFromResultRow builds a Layout from a database result row (detail view).
func buildLayoutFromResultRow(row map[string]interface{}) (Layout, error) {
	id, ok := row["layout_id"].(string)
	if !ok {
		return Layout{}, fmt.Errorf("layout_id not found or invalid type")
	}

	displayName, ok := row["display_name"].(string)
	if !ok {
		return Layout{}, fmt.Errorf("display_name not found or invalid type")
	}

	description := ""
	if descInterface, ok := row["description"]; ok && descInterface != nil {
		description, _ = descInterface.(string)
	}

	preferencesInterface, ok := row["preferences"]
	if !ok {
		return Layout{}, fmt.Errorf("preferences not found")
	}

	var preferences json.RawMessage
	switch v := preferencesInterface.(type) {
	case string:
		preferences = json.RawMessage(v)
	case []byte:
		preferences = json.RawMessage(v)
	default:
		return Layout{}, fmt.Errorf("unexpected type for preferences: %T", preferencesInterface)
	}

	return Layout{
		ID:          id,
		DisplayName: displayName,
		Description: description,
		Preferences: preferences,
	}, nil
}
