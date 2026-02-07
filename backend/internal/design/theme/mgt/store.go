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

package thememgt

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/log"
)

const storeLoggerComponentName = "ThemeMgtStore"

var errThemeNotFound = errors.New("theme not found")

// themeMgtStoreInterface defines the interface for theme management store operations.
type themeMgtStoreInterface interface {
	GetThemeListCount() (int, error)
	GetThemeList(limit, offset int) ([]Theme, error)
	CreateTheme(id string, theme CreateThemeRequest) error
	GetTheme(id string) (Theme, error)
	IsThemeExist(id string) (bool, error)
	UpdateTheme(id string, theme UpdateThemeRequest) error
	DeleteTheme(id string) error
	GetApplicationsCountByThemeID(id string) (int, error)
}

// themeMgtStore is the default implementation of themeMgtStoreInterface.
type themeMgtStore struct {
	dbProvider   provider.DBProviderInterface
	deploymentID string
}

// newThemeMgtStore creates a new instance of themeMgtStore.
func newThemeMgtStore() themeMgtStoreInterface {
	return &themeMgtStore{
		dbProvider:   provider.GetDBProvider(),
		deploymentID: config.GetThunderRuntime().Config.Server.Identifier,
	}
}

// GetThemeListCount retrieves the total count of theme configurations.
func (s *themeMgtStore) GetThemeListCount() (int, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return 0, err
	}

	countResults, err := dbClient.Query(queryGetThemeListCount, s.deploymentID)
	if err != nil {
		return 0, fmt.Errorf("failed to execute count query: %w", err)
	}

	return parseCountResult(countResults)
}

// GetThemeList retrieves theme configurations with pagination.
func (s *themeMgtStore) GetThemeList(limit, offset int) ([]Theme, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return nil, err
	}

	results, err := dbClient.Query(queryGetThemeList, limit, offset, s.deploymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute theme list query: %w", err)
	}

	themes := make([]Theme, 0)
	for _, row := range results {
		theme, err := buildThemeListItemFromResultRow(row)
		if err != nil {
			return nil, fmt.Errorf("failed to build theme from result row: %w", err)
		}
		themes = append(themes, theme)
	}

	return themes, nil
}

// CreateTheme creates a new theme configuration in the database.
func (s *themeMgtStore) CreateTheme(id string, theme CreateThemeRequest) error {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return err
	}

	themeJSON, err := json.Marshal(theme.Theme)
	if err != nil {
		return fmt.Errorf("failed to marshal theme: %w", err)
	}

	_, err = dbClient.Execute(queryCreateTheme, id, theme.DisplayName, theme.Description, themeJSON, s.deploymentID)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}

// GetTheme retrieves a theme configuration by its id.
func (s *themeMgtStore) GetTheme(id string) (Theme, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return Theme{}, err
	}

	results, err := dbClient.Query(queryGetThemeByID, id, s.deploymentID)
	if err != nil {
		return Theme{}, fmt.Errorf("failed to execute query: %w", err)
	}

	if len(results) == 0 {
		return Theme{}, errThemeNotFound
	}

	if len(results) != 1 {
		return Theme{}, fmt.Errorf("unexpected number of results: %d", len(results))
	}

	return buildThemeFromResultRow(results[0])
}

// IsThemeExist checks if a theme configuration exists by its ID.
func (s *themeMgtStore) IsThemeExist(id string) (bool, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return false, err
	}

	results, err := dbClient.Query(queryCheckThemeExists, id, s.deploymentID)
	if err != nil {
		return false, fmt.Errorf("failed to check theme existence: %w", err)
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

// UpdateTheme updates a theme configuration.
func (s *themeMgtStore) UpdateTheme(id string, theme UpdateThemeRequest) error {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return err
	}

	themeJSON, err := json.Marshal(theme.Theme)
	if err != nil {
		return fmt.Errorf("failed to marshal theme: %w", err)
	}

	_, err = dbClient.Execute(queryUpdateTheme, theme.DisplayName, theme.Description, themeJSON, id, s.deploymentID)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}

// DeleteTheme deletes a theme configuration.
func (s *themeMgtStore) DeleteTheme(id string) error {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return err
	}

	_, err = dbClient.Execute(queryDeleteTheme, id, s.deploymentID)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}

// GetApplicationsCountByThemeID returns the count of applications using a specific theme.
func (s *themeMgtStore) GetApplicationsCountByThemeID(id string) (int, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return 0, err
	}

	results, err := dbClient.Query(queryGetApplicationsCountByThemeID, id, s.deploymentID)
	if err != nil {
		return 0, fmt.Errorf("failed to get applications count: %w", err)
	}

	return parseCountResult(results)
}

// getIdentityDBClient retrieves the identity database client.
func (s *themeMgtStore) getIdentityDBClient() (provider.DBClientInterface, error) {
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

// buildThemeListItemFromResultRow builds a Theme from a database result row (list view).
func buildThemeListItemFromResultRow(row map[string]interface{}) (Theme, error) {
	id, ok := row["theme_id"].(string)
	if !ok {
		return Theme{}, fmt.Errorf("theme_id not found or invalid type")
	}

	displayName, ok := row["display_name"].(string)
	if !ok {
		return Theme{}, fmt.Errorf("display_name not found or invalid type")
	}

	return Theme{
		ID:          id,
		DisplayName: displayName,
	}, nil
}

// buildThemeFromResultRow builds a Theme from a database result row (detail view).
func buildThemeFromResultRow(row map[string]interface{}) (Theme, error) {
	id, ok := row["theme_id"].(string)
	if !ok {
		return Theme{}, fmt.Errorf("theme_id not found or invalid type")
	}

	displayName, ok := row["display_name"].(string)
	if !ok {
		return Theme{}, fmt.Errorf("display_name not found or invalid type")
	}

	description := ""
	if desc, ok := row["description"].(string); ok {
		description = desc
	}

	themeInterface, ok := row["theme"]
	if !ok {
		return Theme{}, fmt.Errorf("theme not found")
	}

	var theme json.RawMessage
	switch v := themeInterface.(type) {
	case string:
		theme = json.RawMessage(v)
	case []byte:
		theme = json.RawMessage(v)
	default:
		return Theme{}, fmt.Errorf("unexpected type for theme: %T", themeInterface)
	}

	return Theme{
		ID:          id,
		DisplayName: displayName,
		Description: description,
		Theme:       theme,
	}, nil
}
