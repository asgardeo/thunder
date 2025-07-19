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

// Package store provides the implementation for organization unit persistence operations.
package store

import (
	"fmt"

	"github.com/asgardeo/thunder/internal/ou/constants"
	"github.com/asgardeo/thunder/internal/ou/model"
	dbmodel "github.com/asgardeo/thunder/internal/system/database/model"
	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/log"
)

const loggerComponentName = "OrganizationUnitStore"

// GetOrganizationUnitListCount retrieves the total count of organization units.
func GetOrganizationUnitListCount() (int, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	dbClient, err := provider.NewDBProvider().GetDBClient("identity")
	if err != nil {
		return 0, fmt.Errorf("failed to get database client: %w", err)
	}
	defer func() {
		if closeErr := dbClient.Close(); closeErr != nil {
			logger.Error("Failed to close database client", log.Error(closeErr))
		}
	}()

	results, err := dbClient.Query(QueryGetOrganizationUnitListCount)
	if err != nil {
		return 0, fmt.Errorf("failed to execute count query: %w", err)
	}

	var total int
	if len(results) > 0 {
		if count, ok := results[0]["total"].(int64); ok {
			total = int(count)
		} else {
			return 0, fmt.Errorf("unexpected type for total: %T", results[0]["total"])
		}
	}

	return total, nil
}

// GetOrganizationUnitList retrieves organization units with pagination.
func GetOrganizationUnitList(limit, offset int) ([]model.OrganizationUnitBasic, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	dbClient, err := provider.NewDBProvider().GetDBClient("identity")
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}
	defer func() {
		if closeErr := dbClient.Close(); closeErr != nil {
			logger.Error("Failed to close database client", log.Error(closeErr))
		}
	}()

	results, err := dbClient.Query(QueryGetOrganizationUnitList, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	ous := make([]model.OrganizationUnitBasic, 0, len(results))
	for _, row := range results {
		ou, err := buildOrganizationUnitBasicFromResultRow(row)
		if err != nil {
			return nil, fmt.Errorf("failed to build organization unit basic: %w", err)
		}
		ous = append(ous, ou)
	}

	return ous, nil
}

// CreateOrganizationUnit creates a new organization unit in the database.
func CreateOrganizationUnit(ou model.OrganizationUnit) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	dbClient, err := provider.NewDBProvider().GetDBClient("identity")
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}
	defer func() {
		if closeErr := dbClient.Close(); closeErr != nil {
			logger.Error("Failed to close database client", log.Error(closeErr))
		}
	}()

	_, err = dbClient.Execute(
		QueryCreateOrganizationUnit,
		ou.ID,
		ou.Parent,
		ou.Handle,
		ou.Name,
		ou.Description,
	)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}

// GetOrganizationUnit retrieves an organization unit by its id.
func GetOrganizationUnit(id string) (model.OrganizationUnit, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	dbClient, err := provider.NewDBProvider().GetDBClient("identity")
	if err != nil {
		return model.OrganizationUnit{}, fmt.Errorf("failed to get database client: %w", err)
	}
	defer func() {
		if closeErr := dbClient.Close(); closeErr != nil {
			logger.Error("Failed to close database client", log.Error(closeErr))
		}
	}()

	results, err := dbClient.Query(QueryGetOrganizationUnitByID, id)
	if err != nil {
		return model.OrganizationUnit{}, fmt.Errorf("failed to execute query: %w", err)
	}

	if len(results) == 0 {
		return model.OrganizationUnit{}, constants.ErrOrganizationUnitNotFound
	}

	ou, err := buildOrganizationUnitFromResultRow(results[0])
	if err != nil {
		return model.OrganizationUnit{}, fmt.Errorf("failed to build organization unit: %w", err)
	}

	subOUs, err := GetSubOrganizationUnits(id)
	if err != nil {
		return model.OrganizationUnit{}, fmt.Errorf("failed to get sub organization units: %w", err)
	}
	ou.OrganizationUnits = *subOUs

	users, err := GetOrganizationUnitUsers(id)
	if err != nil {
		return model.OrganizationUnit{}, fmt.Errorf("failed to get organization unit users: %w", err)
	}
	ou.Users = *users

	groups, err := GetOrganizationUnitGroups(id)
	if err != nil {
		return model.OrganizationUnit{}, fmt.Errorf("failed to get organization unit groups: %w", err)
	}
	ou.Groups = *groups

	return ou, nil
}

// UpdateOrganizationUnit updates an existing organization unit.
func UpdateOrganizationUnit(ou model.OrganizationUnit) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	dbClient, err := provider.NewDBProvider().GetDBClient("identity")
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}
	defer func() {
		if closeErr := dbClient.Close(); closeErr != nil {
			logger.Error("Failed to close database client", log.Error(closeErr))
		}
	}()

	_, err = dbClient.Execute(
		QueryUpdateOrganizationUnit,
		ou.ID,
		ou.Parent,
		ou.Handle,
		ou.Name,
		ou.Description,
	)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}

// DeleteOrganizationUnit deletes an organization unit.
func DeleteOrganizationUnit(id string) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	dbClient, err := provider.NewDBProvider().GetDBClient("identity")
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}
	defer func() {
		if closeErr := dbClient.Close(); closeErr != nil {
			logger.Error("Failed to close database client", log.Error(closeErr))
		}
	}()

	_, err = dbClient.Execute(QueryDeleteOrganizationUnit, id)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}

// GetSubOrganizationUnitsByParentIDs retrieves sub organization units for multiple parent IDs.
func GetSubOrganizationUnitsByParentIDs(parentIDs []string) (map[string][]string, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	if len(parentIDs) == 0 {
		return make(map[string][]string), nil
	}

	dbClient, err := provider.NewDBProvider().GetDBClient("identity")
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}
	defer func() {
		if closeErr := dbClient.Close(); closeErr != nil {
			logger.Error("Failed to close database client", log.Error(closeErr))
		}
	}()

	query, args, err := buildSubOrganizationUnitsQuery(parentIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	results, err := dbClient.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	subOUsMap := make(map[string][]string)

	for _, parentID := range parentIDs {
		subOUsMap[parentID] = make([]string, 0)
	}

	for _, row := range results {
		ouID, ok := row["ou_id"].(string)
		if !ok {
			return nil, fmt.Errorf("ou_id is not a string")
		}

		parentID, ok := row["parent_id"].(string)
		if !ok {
			return nil, fmt.Errorf("parent_id is not a string")
		}

		subOUsMap[parentID] = append(subOUsMap[parentID], ouID)
	}

	return subOUsMap, nil
}

// executeQueryForStringArray executes a query and returns a slice of strings for a specified field name.
func executeQueryForStringArray(
	query dbmodel.DBQuery, fieldName string, params ...interface{},
) (*[]string, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	dbClient, err := provider.NewDBProvider().GetDBClient("identity")
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}
	defer func() {
		if closeErr := dbClient.Close(); closeErr != nil {
			logger.Error("Failed to close database client", log.Error(closeErr))
		}
	}()

	results, err := dbClient.Query(query, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	values := make([]string, 0)
	for _, row := range results {
		if val, ok := row[fieldName].(string); ok {
			values = append(values, val)
		} else {
			return nil, fmt.Errorf("expected %s to be a string", fieldName)
		}
	}

	return &values, nil
}

// GetSubOrganizationUnits retrieves the sub organization units of a given organization unit ID.
func GetSubOrganizationUnits(ouID string) (*[]string, error) {
	return executeQueryForStringArray(QueryGetSubOrganizationUnits, "ou_id", ouID)
}

// GetOrganizationUnitUsers retrieves the users of a given organization unit ID.
func GetOrganizationUnitUsers(ouID string) (*[]string, error) {
	return executeQueryForStringArray(QueryGetOrganizationUnitUsers, "user_id", ouID)
}

// GetOrganizationUnitGroups retrieves the groups of a given organization unit ID.
func GetOrganizationUnitGroups(ouID string) (*[]string, error) {
	return executeQueryForStringArray(QueryGetOrganizationUnitGroups, "group_id", ouID)
}

// CheckOrganizationUnitNameConflict checks if an organization unit name conflicts under the same parent.
func CheckOrganizationUnitNameConflict(name string, parentID *string) (bool, error) {
	return checkConflict(
		QueryCheckOrganizationUnitNameConflict,
		QueryCheckOrganizationUnitNameConflictRoot,
		name,
		parentID,
	)
}

// CheckOrganizationUnitHandleConflict checks if an organization unit handle conflicts under the same parent.
func CheckOrganizationUnitHandleConflict(handle string, parentID *string) (bool, error) {
	return checkConflict(
		QueryCheckOrganizationUnitHandleConflict,
		QueryCheckOrganizationUnitHandleConflictRoot,
		handle,
		parentID,
	)
}

// CheckOrganizationUnitHasChildResources checks if an organization unit has users groups or sub-ous.
func CheckOrganizationUnitHasChildResources(ouID string) (bool, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	dbClient, err := provider.NewDBProvider().GetDBClient("identity")
	if err != nil {
		return false, fmt.Errorf("failed to get database client: %w", err)
	}
	defer func() {
		if closeErr := dbClient.Close(); closeErr != nil {
			logger.Error("Failed to close database client", log.Error(closeErr))
		}
	}()

	results, err := dbClient.Query(QueryCheckOrganizationUnitHasUsersOrGroups, ouID)
	if err != nil {
		return false, fmt.Errorf("failed to execute query: %w", err)
	}

	if len(results) > 0 {
		if count, ok := results[0]["count"].(int64); ok && count > 0 {
			return true, nil
		}
	}

	return false, nil
}

// buildOrganizationUnitBasicFromResultRow constructs a model.OrganizationUnitBasic from a database result row.
func buildOrganizationUnitBasicFromResultRow(
	row map[string]interface{},
) (model.OrganizationUnitBasic, error) {
	ouID, ok := row["ou_id"].(string)
	if !ok {
		return model.OrganizationUnitBasic{}, fmt.Errorf("ou_id is not a string")
	}

	name, ok := row["name"].(string)
	if !ok {
		return model.OrganizationUnitBasic{}, fmt.Errorf("name is not a string")
	}

	handle, ok := row["handle"].(string)
	if !ok {
		return model.OrganizationUnitBasic{}, fmt.Errorf("handle is not a string")
	}

	description := ""
	if desc, ok := row["description"]; ok && desc != nil {
		if descStr, ok := desc.(string); ok {
			description = descStr
		}
	}

	var parentID *string
	if parent, ok := row["parent_id"]; ok && parent != nil {
		if parentStr, ok := parent.(string); ok {
			parentID = &parentStr
		}
	}

	return model.OrganizationUnitBasic{
		ID:                ouID,
		Handle:            handle,
		Name:              name,
		Description:       description,
		Parent:            parentID,
		OrganizationUnits: make([]string, 0), // Will be populated by caller
	}, nil
}

// buildOrganizationUnitFromResultRow constructs a model.OrganizationUnit from a database result row.
func buildOrganizationUnitFromResultRow(
	row map[string]interface{},
) (model.OrganizationUnit, error) {
	ou, err := buildOrganizationUnitBasicFromResultRow(row)
	if err != nil {
		return model.OrganizationUnit{}, fmt.Errorf("failed to build organization unit: %w", err)
	}

	return model.OrganizationUnit{
		ID:                ou.ID,
		Handle:            ou.Handle,
		Name:              ou.Name,
		Description:       ou.Description,
		Parent:            ou.Parent,
		Users:             []string{},
		Groups:            []string{},
		OrganizationUnits: []string{},
	}, nil
}

// checkConflict is a helper function to check for conflicts in organization unit attributes.
func checkConflict(
	queryWithParent, queryWithoutParent dbmodel.DBQuery,
	value string,
	parentID *string,
	extraArgs ...interface{},
) (bool, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	dbClient, err := provider.NewDBProvider().GetDBClient("identity")
	if err != nil {
		return false, fmt.Errorf("failed to get database client: %w", err)
	}
	defer func() {
		if closeErr := dbClient.Close(); closeErr != nil {
			logger.Error("Failed to close database client", log.Error(closeErr))
		}
	}()

	var results []map[string]interface{}

	if parentID != nil {
		args := append([]interface{}{value, *parentID}, extraArgs...)
		results, err = dbClient.Query(queryWithParent, args...)
	} else {
		args := append([]interface{}{value}, extraArgs...)
		results, err = dbClient.Query(queryWithoutParent, args...)
	}

	if err != nil {
		return false, fmt.Errorf("failed to execute query: %w", err)
	}

	if len(results) > 0 {
		if count, ok := results[0]["count"].(int64); ok && count > 0 {
			return true, nil
		}
	}

	return false, nil
}
