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

package branding

import (
	"encoding/json"
	"fmt"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/log"
)

const storeLoggerComponentName = "BrandingStore"

// brandingStoreInterface defines the interface for branding store operations.
type brandingStoreInterface interface {
	GetBrandingListCount() (int, error)
	GetBrandingList(limit, offset int) ([]Branding, error)
	CreateBranding(id string, branding CreateBrandingRequest) error
	GetBranding(id string) (Branding, error)
	IsBrandingExist(id string) (bool, error)
	UpdateBranding(id string, branding UpdateBrandingRequest) error
	DeleteBranding(id string) error
	GetApplicationsCountByBrandingID(id string) (int, error)
}

// brandingStore is the default implementation of brandingStoreInterface.
type brandingStore struct {
	dbProvider   provider.DBProviderInterface
	deploymentID string
}

// newBrandingStore creates a new instance of brandingStore.
func newBrandingStore() brandingStoreInterface {
	return &brandingStore{
		dbProvider:   provider.GetDBProvider(),
		deploymentID: config.GetThunderRuntime().Config.Server.Identifier,
	}
}

// GetBrandingListCount retrieves the total count of branding configurations.
func (s *brandingStore) GetBrandingListCount() (int, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return 0, err
	}

	countResults, err := dbClient.Query(queryGetBrandingListCount, s.deploymentID)
	if err != nil {
		return 0, fmt.Errorf("failed to execute count query: %w", err)
	}

	return parseCountResult(countResults)
}

// GetBrandingList retrieves branding configurations with pagination.
func (s *brandingStore) GetBrandingList(limit, offset int) ([]Branding, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return nil, err
	}

	results, err := dbClient.Query(queryGetBrandingList, limit, offset, s.deploymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute branding list query: %w", err)
	}

	brandings := make([]Branding, 0)
	for _, row := range results {
		branding, err := buildBrandingListItemFromResultRow(row)
		if err != nil {
			return nil, fmt.Errorf("failed to build branding from result row: %w", err)
		}
		brandings = append(brandings, branding)
	}

	return brandings, nil
}

// CreateBranding creates a new branding configuration in the database.
func (s *brandingStore) CreateBranding(id string, branding CreateBrandingRequest) error {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return err
	}

	preferencesJSON, err := json.Marshal(branding.Preferences)
	if err != nil {
		return fmt.Errorf("failed to marshal preferences: %w", err)
	}

	_, err = dbClient.Execute(queryCreateBranding, id, branding.DisplayName, preferencesJSON, s.deploymentID)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}

// GetBranding retrieves a branding configuration by its id.
func (s *brandingStore) GetBranding(id string) (Branding, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return Branding{}, err
	}

	results, err := dbClient.Query(queryGetBrandingByID, id, s.deploymentID)
	if err != nil {
		return Branding{}, fmt.Errorf("failed to execute query: %w", err)
	}

	if len(results) == 0 {
		return Branding{}, ErrBrandingNotFound
	}

	if len(results) != 1 {
		return Branding{}, fmt.Errorf("unexpected number of results: %d", len(results))
	}

	return buildBrandingFromResultRow(results[0])
}

// IsBrandingExist checks if a branding configuration exists by its ID without fetching its details.
func (s *brandingStore) IsBrandingExist(id string) (bool, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return false, err
	}

	results, err := dbClient.Query(queryCheckBrandingExists, id, s.deploymentID)
	if err != nil {
		return false, fmt.Errorf("failed to check branding existence: %w", err)
	}

	return parseBoolFromCount(results)
}

// UpdateBranding updates an existing branding configuration.
func (s *brandingStore) UpdateBranding(id string, branding UpdateBrandingRequest) error {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return err
	}

	preferencesJSON, err := json.Marshal(branding.Preferences)
	if err != nil {
		return fmt.Errorf("failed to marshal preferences: %w", err)
	}

	rowsAffected, err := dbClient.Execute(queryUpdateBranding, branding.DisplayName, preferencesJSON, id, s.deploymentID)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	if rowsAffected == 0 {
		return ErrBrandingNotFound
	}

	return nil
}

// DeleteBranding deletes a branding configuration.
func (s *brandingStore) DeleteBranding(id string) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, storeLoggerComponentName))

	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	rowsAffected, err := dbClient.Execute(queryDeleteBranding, id, s.deploymentID)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	if rowsAffected == 0 {
		logger.Debug("Branding not found with id: " + id)
	}

	return nil
}

// GetApplicationsCountByBrandingID retrieves the count of applications using a branding configuration.
func (s *brandingStore) GetApplicationsCountByBrandingID(id string) (int, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return 0, err
	}

	results, err := dbClient.Query(queryGetApplicationsCountByBrandingID, id, s.deploymentID)
	if err != nil {
		return 0, fmt.Errorf("failed to get applications count: %w", err)
	}

	return parseCountResult(results)
}

// getIdentityDBClient is a helper method to get the database client for the identity database.
func (s *brandingStore) getIdentityDBClient() (provider.DBClientInterface, error) {
	dbClient, err := s.dbProvider.GetConfigDBClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}
	return dbClient, nil
}

// buildBrandingFromResultRow builds a Branding struct from a database result row.
func buildBrandingFromResultRow(row map[string]interface{}) (Branding, error) {
	id, err := parseStringField(row, "branding_id")
	if err != nil {
		return Branding{}, err
	}

	displayName, err := parseStringField(row, "display_name")
	if err != nil {
		return Branding{}, err
	}

	var preferences json.RawMessage
	preferencesVal := row["preferences"]
	if preferencesVal != nil {
		switch v := preferencesVal.(type) {
		case string:
			preferences = json.RawMessage(v)
		case []byte:
			preferences = json.RawMessage(v)
		default:
			prefBytes, err := json.Marshal(v)
			if err != nil {
				return Branding{}, fmt.Errorf("failed to marshal preferences: %w", err)
			}
			preferences = json.RawMessage(prefBytes)
		}
	}

	return Branding{
		ID:          id,
		DisplayName: displayName,
		Preferences: preferences,
	}, nil
}

// buildBrandingListItemFromResultRow builds a Branding struct from a database result row.
func buildBrandingListItemFromResultRow(row map[string]interface{}) (Branding, error) {
	id, err := parseStringField(row, "branding_id")
	if err != nil {
		return Branding{}, err
	}

	displayName, err := parseStringField(row, "display_name")
	if err != nil {
		return Branding{}, err
	}

	return Branding{
		ID:          id,
		DisplayName: displayName,
	}, nil
}

// parseCountResult parses a count result from a database query result.
func parseCountResult(results []map[string]interface{}) (int, error) {
	if len(results) == 0 {
		return 0, nil
	}

	if countVal, ok := results[0]["total"].(int64); ok {
		return int(countVal), nil
	}
	if countVal, ok := results[0]["count"].(int64); ok {
		return int(countVal), nil
	}
	return 0, fmt.Errorf("failed to parse total/count from query result")
}

// parseBoolFromCount parses a count result and returns true if count > 0.
func parseBoolFromCount(results []map[string]interface{}) (bool, error) {
	if len(results) == 0 {
		return false, nil
	}

	if countVal, ok := results[0]["count"].(int64); ok {
		return countVal > 0, nil
	}
	return false, fmt.Errorf("failed to parse count from query result")
}

// parseStringField extracts a string field from a database result row.
func parseStringField(row map[string]interface{}, fieldName string) (string, error) {
	value, ok := row[fieldName].(string)
	if !ok {
		return "", fmt.Errorf("failed to parse %s as string", fieldName)
	}
	return value, nil
}
