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

package idp

import (
	"fmt"

	"github.com/asgardeo/thunder/internal/system/cmodels"
	dbmodel "github.com/asgardeo/thunder/internal/system/database/model"
	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/log"
)

// idpStoreInterface defines the interface for identity provider store operations.
type idpStoreInterface interface {
	CreateIdentityProvider(idp IDPDTO) error
	GetIdentityProviderList() ([]BasicIDPDTO, error)
	GetIdentityProvider(idpID string) (*IDPDTO, error)
	GetIdentityProviderByName(idpName string) (*IDPDTO, error)
	UpdateIdentityProvider(idp *IDPDTO) error
	DeleteIdentityProvider(idpID string) error
}

// idpStore is the default implementation of IDPStoreInterface.
type idpStore struct {
	dbProvider provider.DBProviderInterface
}

// newIDPStore creates a new instance of IDPStore.
func newIDPStore() idpStoreInterface {
	return &idpStore{
		dbProvider: provider.GetDBProvider(),
	}
}

// CreateIdentityProvider handles the IdP creation in the database.
func (s *idpStore) CreateIdentityProvider(idp IDPDTO) error {
	dbClient, err := s.dbProvider.GetDBClient("identity")
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	var propertiesJSON string
	if len(idp.Properties) > 0 {
		propertiesJSON, err = cmodels.SerializePropertiesToJSONArray(idp.Properties)
		if err != nil {
			return fmt.Errorf("failed to serialize properties to JSON: %w", err)
		}
	}

	_, err = dbClient.Execute(queryCreateIdentityProvider, idp.ID, idp.Name, idp.Description, idp.Type, propertiesJSON)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}

// GetIdentityProviderList retrieves a list of IdPs from the database.
func (s *idpStore) GetIdentityProviderList() ([]BasicIDPDTO, error) {
	dbClient, err := s.dbProvider.GetDBClient("identity")
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(queryGetIdentityProviderList)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	idpList := make([]BasicIDPDTO, 0)
	for _, row := range results {
		idp, err := buildIDPFromResultRow(row)
		if err != nil {
			return nil, fmt.Errorf("failed to build idp from result row: %w", err)
		}
		idpList = append(idpList, *idp)
	}

	return idpList, nil
}

// GetIdentityProvider retrieves a specific idp by its ID from the database.
func (s *idpStore) GetIdentityProvider(id string) (*IDPDTO, error) {
	return s.getIDP(queryGetIdentityProviderByID, id)
}

// GetIdentityProviderByName retrieves a specific idp by its name from the database.
func (s *idpStore) GetIdentityProviderByName(name string) (*IDPDTO, error) {
	return s.getIDP(queryGetIdentityProviderByName, name)
}

// getIDP retrieves an IDP based on the provided query and identifier.
func (s *idpStore) getIDP(query dbmodel.DBQuery, identifier string) (*IDPDTO, error) {
	dbClient, err := s.dbProvider.GetDBClient("identity")
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(query, identifier)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	if len(results) == 0 {
		return nil, ErrIDPNotFound
	}
	if len(results) != 1 {
		return nil, fmt.Errorf("unexpected number of results: %d", len(results))
	}

	row := results[0]

	basicIDP, err := buildIDPFromResultRow(row)
	if err != nil {
		return nil, fmt.Errorf("failed to build idp from result row: %w", err)
	}

	var properties []cmodels.Property
	var propertiesJSON string

	// Handle both string and []byte types for properties
	switch v := row["properties"].(type) {
	case string:
		propertiesJSON = v
	case []byte:
		propertiesJSON = string(v)
	}

	if propertiesJSON != "" {
		var err error
		properties, err = cmodels.DeserializePropertiesFromJSON(propertiesJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize properties from JSON: %w", err)
		}
	}

	idp := &IDPDTO{
		ID:          basicIDP.ID,
		Name:        basicIDP.Name,
		Description: basicIDP.Description,
		Type:        basicIDP.Type,
		Properties:  properties,
	}

	return idp, nil
}

// UpdateIdentityProvider updates the idp in the database.
func (s *idpStore) UpdateIdentityProvider(idp *IDPDTO) error {
	dbClient, err := s.dbProvider.GetDBClient("identity")
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	var propertiesJSON string
	if len(idp.Properties) > 0 {
		propertiesJSON, err = cmodels.SerializePropertiesToJSONArray(idp.Properties)
		if err != nil {
			return fmt.Errorf("failed to serialize properties to JSON: %w", err)
		}
	}

	// Update the IDP in the database
	_, err = dbClient.Execute(queryUpdateIdentityProviderByID, idp.ID, idp.Name,
		idp.Description, idp.Type, propertiesJSON)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}

// DeleteIdentityProvider deletes the idp from the database.
func (s *idpStore) DeleteIdentityProvider(id string) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "IdPStore"))

	dbClient, err := s.dbProvider.GetDBClient("identity")
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	rowsAffected, err := dbClient.Execute(queryDeleteIdentityProviderByID, id)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	if rowsAffected == 0 {
		logger.Debug("idp not found with id: " + id)
	}

	return nil
}

func buildIDPFromResultRow(row map[string]interface{}) (*BasicIDPDTO, error) {
	idpID, ok := row["idp_id"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to parse idp_id as string")
	}

	idpName, ok := row["name"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to parse name as string")
	}

	idpDescription, ok := row["description"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to parse description as string")
	}

	idpType, ok := row["type"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to parse type as string")
	}

	idp := BasicIDPDTO{
		ID:          idpID,
		Name:        idpName,
		Description: idpDescription,
		Type:        IDPType(idpType),
	}

	return &idp, nil
}
