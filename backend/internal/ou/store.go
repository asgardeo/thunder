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

package ou

import (
	"fmt"

	"github.com/asgardeo/thunder/internal/system/config"
	dbmodel "github.com/asgardeo/thunder/internal/system/database/model"
	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/log"
)

const storeLoggerComponentName = "OrganizationUnitStore"

// organizationUnitStoreInterface defines the interface for organization unit store operations.
type organizationUnitStoreInterface interface {
	GetOrganizationUnitListCount() (int, error)
	GetOrganizationUnitList(limit, offset int) ([]OrganizationUnitBasic, error)
	CreateOrganizationUnit(ou OrganizationUnit) error
	GetOrganizationUnit(id string) (OrganizationUnit, error)
	GetOrganizationUnitByPath(handles []string) (OrganizationUnit, error)
	IsOrganizationUnitExists(id string) (bool, error)
	IsOrganizationUnitDeclarative(id string) bool
	CheckOrganizationUnitNameConflict(name string, parent *string) (bool, error)
	CheckOrganizationUnitHandleConflict(handle string, parent *string) (bool, error)
	UpdateOrganizationUnit(ou OrganizationUnit) error
	DeleteOrganizationUnit(id string) error
	CheckOrganizationUnitHasChildResources(id string) (bool, error)
	GetOrganizationUnitChildrenCount(id string) (int, error)
	GetOrganizationUnitChildrenList(id string, limit, offset int) ([]OrganizationUnitBasic, error)
	GetOrganizationUnitUsersCount(id string) (int, error)
	GetOrganizationUnitUsersList(id string, limit, offset int) ([]User, error)
	GetOrganizationUnitGroupsCount(id string) (int, error)
	GetOrganizationUnitGroupsList(id string, limit, offset int) ([]Group, error)
}

// organizationUnitStore is the default implementation of organizationUnitStoreInterface.
type organizationUnitStore struct {
	dbProvider   provider.DBProviderInterface
	deploymentID string
}

// newOrganizationUnitStore creates a new instance of organizationUnitStore.
func newOrganizationUnitStore() organizationUnitStoreInterface {
	return &organizationUnitStore{
		dbProvider:   provider.GetDBProvider(),
		deploymentID: config.GetThunderRuntime().Config.Server.Identifier,
	}
}

// GetOrganizationUnitListCount retrieves the total count of organization units.
func (s *organizationUnitStore) GetOrganizationUnitListCount() (int, error) {
	dbClient, err := s.dbProvider.GetUserDBClient()
	if err != nil {
		return 0, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(queryGetRootOrganizationUnitListCount, s.deploymentID)
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
func (s *organizationUnitStore) GetOrganizationUnitList(limit, offset int) ([]OrganizationUnitBasic, error) {
	dbClient, err := s.dbProvider.GetUserDBClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(queryGetRootOrganizationUnitList, limit, offset, s.deploymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	ous := make([]OrganizationUnitBasic, 0, len(results))
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
func (s *organizationUnitStore) CreateOrganizationUnit(ou OrganizationUnit) error {
	dbClient, err := s.dbProvider.GetUserDBClient()
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	_, err = dbClient.Execute(
		queryCreateOrganizationUnit,
		ou.ID,
		ou.Parent,
		ou.Handle,
		ou.Name,
		ou.Description,
		s.deploymentID,
	)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}

// GetOrganizationUnit retrieves an organization unit by its id.
func (s *organizationUnitStore) GetOrganizationUnit(id string) (OrganizationUnit, error) {
	dbClient, err := s.dbProvider.GetUserDBClient()
	if err != nil {
		return OrganizationUnit{}, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(queryGetOrganizationUnitByID, id, s.deploymentID)
	if err != nil {
		return OrganizationUnit{}, fmt.Errorf("failed to execute query: %w", err)
	}

	if len(results) == 0 {
		return OrganizationUnit{}, ErrOrganizationUnitNotFound
	}

	ou, err := buildOrganizationUnitFromResultRow(results[0])
	if err != nil {
		return OrganizationUnit{}, fmt.Errorf("failed to build organization unit: %w", err)
	}

	return ou, nil
}

// GetOrganizationUnitByPath retrieves an organization unit by its hierarchical handle path.
func (s *organizationUnitStore) GetOrganizationUnitByPath(handlePath []string) (OrganizationUnit, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, storeLoggerComponentName))

	if len(handlePath) == 0 {
		return OrganizationUnit{}, ErrOrganizationUnitNotFound
	}

	dbClient, err := s.dbProvider.GetUserDBClient()
	if err != nil {
		return OrganizationUnit{}, fmt.Errorf("failed to get database client: %w", err)
	}

	var currentOU OrganizationUnit
	var parentID *string
	var fullPath string

	for i, handle := range handlePath {
		fullPath = fullPath + "/" + handle
		var results []map[string]interface{}

		if parentID == nil {
			results, err = dbClient.Query(queryGetRootOrganizationUnitByHandle, handle, s.deploymentID)
		} else {
			results, err = dbClient.Query(queryGetOrganizationUnitByHandle, handle, *parentID, s.deploymentID)
		}

		if err != nil {
			return OrganizationUnit{}, fmt.Errorf("failed to execute query for handle %s: %w", handle, err)
		}

		if len(results) == 0 {
			logger.Debug("Organization unit not found in path",
				log.String("handle", handle),
				log.Int("pathIndex", i),
				log.String("fullPath", fullPath))
			return OrganizationUnit{}, ErrOrganizationUnitNotFound
		}

		currentOU, err = buildOrganizationUnitFromResultRow(results[0])
		if err != nil {
			return OrganizationUnit{}, fmt.Errorf("failed to build organization unit for handle %s: %w", handle, err)
		}

		parentID = &currentOU.ID
	}

	return currentOU, nil
}

// IsOrganizationUnitExists checks if an organization unit exists by ID.
func (s *organizationUnitStore) IsOrganizationUnitExists(id string) (bool, error) {
	dbClient, err := s.dbProvider.GetUserDBClient()
	if err != nil {
		return false, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(queryCheckOrganizationUnitExists, id, s.deploymentID)
	if err != nil {
		return false, fmt.Errorf("failed to execute existence check query: %w", err)
	}

	if len(results) == 0 {
		return false, nil
	}

	if countInterface, exists := results[0]["count"]; exists {
		if count, ok := countInterface.(int64); ok {
			return count > 0, nil
		}
	}

	return false, fmt.Errorf("failed to parse existence check result")
}

// IsOrganizationUnitDeclarative checks if an organization unit is immutable.
// Database store resources are always mutable, so this always returns false.
func (s *organizationUnitStore) IsOrganizationUnitDeclarative(id string) bool {
	return false
}

// UpdateOrganizationUnit updates an existing organization unit.
func (s *organizationUnitStore) UpdateOrganizationUnit(ou OrganizationUnit) error {
	dbClient, err := s.dbProvider.GetUserDBClient()
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	_, err = dbClient.Execute(
		queryUpdateOrganizationUnit,
		ou.ID,
		ou.Parent,
		ou.Handle,
		ou.Name,
		ou.Description,
		s.deploymentID,
	)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}

// DeleteOrganizationUnit deletes an organization unit.
func (s *organizationUnitStore) DeleteOrganizationUnit(id string) error {
	dbClient, err := s.dbProvider.GetUserDBClient()
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	_, err = dbClient.Execute(queryDeleteOrganizationUnit, id, s.deploymentID)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}

// GetOrganizationUnitChildrenCount retrieves the total count of child organization units for a given parent ID.
func (s *organizationUnitStore) GetOrganizationUnitChildrenCount(parentID string) (int, error) {
	dbClient, err := s.dbProvider.GetUserDBClient()
	if err != nil {
		return 0, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(queryGetOrganizationUnitChildrenCount, parentID, s.deploymentID)
	if err != nil {
		return 0, fmt.Errorf("failed to execute count query: %w", err)
	}

	if len(results) == 0 {
		return 0, nil
	}

	if totalInterface, exists := results[0]["total"]; exists {
		if total, ok := totalInterface.(int64); ok {
			return int(total), nil
		}
	}

	return 0, fmt.Errorf("failed to parse count result")
}

// GetOrganizationUnitChildrenList retrieves a paginated list of child organization units for a given parent ID.
func (s *organizationUnitStore) GetOrganizationUnitChildrenList(
	parentID string, limit, offset int,
) ([]OrganizationUnitBasic, error) {
	dbClient, err := s.dbProvider.GetUserDBClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(queryGetOrganizationUnitChildrenList, parentID, limit, offset, s.deploymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	childOUs := make([]OrganizationUnitBasic, 0, len(results))
	for _, row := range results {
		childOU, err := buildOrganizationUnitBasicFromResultRow(row)
		if err != nil {
			return nil, fmt.Errorf("failed to build organization unit basic: %w", err)
		}
		childOUs = append(childOUs, childOU)
	}

	return childOUs, nil
}

// GetOrganizationUnitUsersCount retrieves the total count of users in a given organization unit.
func (s *organizationUnitStore) GetOrganizationUnitUsersCount(ouID string) (int, error) {
	dbClient, err := s.dbProvider.GetUserDBClient()
	if err != nil {
		return 0, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(queryGetOrganizationUnitUsersCount, ouID, s.deploymentID)
	if err != nil {
		return 0, fmt.Errorf("failed to execute count query: %w", err)
	}

	if len(results) == 0 {
		return 0, nil
	}

	if totalInterface, exists := results[0]["total"]; exists {
		if total, ok := totalInterface.(int64); ok {
			return int(total), nil
		}
	}

	return 0, fmt.Errorf("failed to parse count result")
}

// GetOrganizationUnitUsersList retrieves a paginated list of users in a given organization unit.
func (s *organizationUnitStore) GetOrganizationUnitUsersList(ouID string, limit, offset int) ([]User, error) {
	dbClient, err := s.dbProvider.GetUserDBClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(queryGetOrganizationUnitUsersList, ouID, limit, offset, s.deploymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	users := make([]User, 0, len(results))
	for _, row := range results {
		if userIDInterface, exists := row["user_id"]; exists {
			if userID, ok := userIDInterface.(string); ok {
				users = append(users, User{ID: userID})
			} else {
				return nil, fmt.Errorf("expected user_id to be a string")
			}
		}
	}

	return users, nil
}

// GetOrganizationUnitGroupsCount retrieves the total count of groups in a given organization unit.
func (s *organizationUnitStore) GetOrganizationUnitGroupsCount(ouID string) (int, error) {
	dbClient, err := s.dbProvider.GetUserDBClient()
	if err != nil {
		return 0, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(queryGetOrganizationUnitGroupsCount, ouID, s.deploymentID)
	if err != nil {
		return 0, fmt.Errorf("failed to execute count query: %w", err)
	}

	if len(results) == 0 {
		return 0, nil
	}

	if totalInterface, exists := results[0]["total"]; exists {
		if total, ok := totalInterface.(int64); ok {
			return int(total), nil
		}
	}

	return 0, fmt.Errorf("failed to parse count result")
}

// GetOrganizationUnitGroupsList retrieves a paginated list of groups in a given organization unit.
func (s *organizationUnitStore) GetOrganizationUnitGroupsList(ouID string, limit, offset int) ([]Group, error) {
	dbClient, err := s.dbProvider.GetUserDBClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(queryGetOrganizationUnitGroupsList, ouID, limit, offset, s.deploymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	groups := make([]Group, 0, len(results))
	for _, row := range results {
		var group Group

		if groupIDInterface, exists := row["group_id"]; exists {
			if groupID, ok := groupIDInterface.(string); ok {
				group.ID = groupID
			} else {
				return nil, fmt.Errorf("expected group_id to be a string")
			}
		}

		if nameInterface, exists := row["name"]; exists {
			if name, ok := nameInterface.(string); ok {
				group.Name = name
			} else {
				return nil, fmt.Errorf("expected name to be a string")
			}
		}

		groups = append(groups, group)
	}

	return groups, nil
}

// CheckOrganizationUnitNameConflict checks if an organization unit name conflicts under the same parent.
func (s *organizationUnitStore) CheckOrganizationUnitNameConflict(name string, parentID *string) (bool, error) {
	return s.checkConflict(
		queryCheckOrganizationUnitNameConflict,
		queryCheckOrganizationUnitNameConflictRoot,
		name,
		parentID,
		s.deploymentID,
	)
}

// CheckOrganizationUnitHandleConflict checks if an organization unit handle conflicts under the same parent.
func (s *organizationUnitStore) CheckOrganizationUnitHandleConflict(handle string, parentID *string) (bool, error) {
	return s.checkConflict(
		queryCheckOrganizationUnitHandleConflict,
		queryCheckOrganizationUnitHandleConflictRoot,
		handle,
		parentID,
		s.deploymentID,
	)
}

// CheckOrganizationUnitHasChildResources checks if an organization unit has users groups or sub-ous.
func (s *organizationUnitStore) CheckOrganizationUnitHasChildResources(ouID string) (bool, error) {
	dbClient, err := s.dbProvider.GetUserDBClient()
	if err != nil {
		return false, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(queryCheckOrganizationUnitHasUsersOrGroups, ouID, s.deploymentID)
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

// buildOrganizationUnitBasicFromResultRow constructs a OrganizationUnitBasic from a database result row.
func buildOrganizationUnitBasicFromResultRow(
	row map[string]interface{},
) (OrganizationUnitBasic, error) {
	ouID, ok := row["ou_id"].(string)
	if !ok {
		return OrganizationUnitBasic{}, fmt.Errorf("ou_id is not a string")
	}

	name, ok := row["name"].(string)
	if !ok {
		return OrganizationUnitBasic{}, fmt.Errorf("name is not a string")
	}

	handle, ok := row["handle"].(string)
	if !ok {
		return OrganizationUnitBasic{}, fmt.Errorf("handle is not a string")
	}

	description := ""
	if desc, ok := row["description"]; ok && desc != nil {
		if descStr, ok := desc.(string); ok {
			description = descStr
		}
	}

	return OrganizationUnitBasic{
		ID:          ouID,
		Handle:      handle,
		Name:        name,
		Description: description,
	}, nil
}

// buildOrganizationUnitFromResultRow constructs a OrganizationUnit from a database result row.
func buildOrganizationUnitFromResultRow(
	row map[string]interface{},
) (OrganizationUnit, error) {
	ou, err := buildOrganizationUnitBasicFromResultRow(row)
	if err != nil {
		return OrganizationUnit{}, fmt.Errorf("failed to build organization unit: %w", err)
	}

	var parentID *string
	if parent, ok := row["parent_id"]; ok && parent != nil {
		if parentStr, ok := parent.(string); ok {
			parentID = &parentStr
		}
	}

	return OrganizationUnit{
		ID:          ou.ID,
		Handle:      ou.Handle,
		Name:        ou.Name,
		Description: ou.Description,
		Parent:      parentID,
	}, nil
}

// checkConflict is a helper function to check for conflicts in organization unit attributes.
func (s *organizationUnitStore) checkConflict(
	queryWithParent, queryWithoutParent dbmodel.DBQuery,
	value string,
	parentID *string,
	extraArgs ...interface{},
) (bool, error) {
	dbClient, err := s.dbProvider.GetUserDBClient()
	if err != nil {
		return false, fmt.Errorf("failed to get database client: %w", err)
	}

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
