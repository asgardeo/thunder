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

package role

import (
	"errors"
	"fmt"

	"github.com/asgardeo/thunder/internal/system/config"
	dbmodel "github.com/asgardeo/thunder/internal/system/database/model"
	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/log"
)

const storeLoggerComponentName = "RoleStore"

// roleStoreInterface defines the interface for role store operations.
type roleStoreInterface interface {
	GetRoleListCount() (int, error)
	GetRoleList(limit, offset int) ([]Role, error)
	CreateRole(id string, role RoleCreationDetail) error
	GetRole(id string) (RoleWithPermissions, error)
	IsRoleExist(id string) (bool, error)
	GetRoleAssignments(id string, limit, offset int) ([]RoleAssignment, error)
	GetRoleAssignmentsCount(id string) (int, error)
	UpdateRole(id string, role RoleUpdateDetail) error
	DeleteRole(id string) error
	AddAssignments(id string, assignments []RoleAssignment) error
	RemoveAssignments(id string, assignments []RoleAssignment) error
	CheckRoleNameExists(ouID, name string) (bool, error)
	CheckRoleNameExistsExcludingID(ouID, name, excludeRoleID string) (bool, error)
	GetAuthorizedPermissions(userID string, groupIDs []string, requestedPermissions []string) ([]string, error)
}

// roleStore is the default implementation of roleStoreInterface.
type roleStore struct {
	dbProvider   provider.DBProviderInterface
	deploymentID string
}

// newRoleStore creates a new instance of roleStore.
func newRoleStore() roleStoreInterface {
	return &roleStore{
		dbProvider:   provider.GetDBProvider(),
		deploymentID: config.GetThunderRuntime().Config.Server.Identifier,
	}
}

// GetRoleListCount retrieves the total count of roles.
func (s *roleStore) GetRoleListCount() (int, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return 0, err
	}

	countResults, err := dbClient.Query(queryGetRoleListCount, s.deploymentID)
	if err != nil {
		return 0, fmt.Errorf("failed to execute count query: %w", err)
	}

	return parseCountResult(countResults)
}

// GetRoleList retrieves roles with pagination.
func (s *roleStore) GetRoleList(limit, offset int) ([]Role, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return nil, err
	}

	results, err := dbClient.Query(queryGetRoleList, limit, offset, s.deploymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute role list query: %w", err)
	}

	roles := make([]Role, 0)
	for _, row := range results {
		role, err := buildRoleBasicInfoFromResultRow(row)
		if err != nil {
			return nil, fmt.Errorf("failed to build role from result row: %w", err)
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// CreateRole creates a new role in the database.
func (s *roleStore) CreateRole(id string, role RoleCreationDetail) error {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return err
	}

	return s.executeInTransaction(dbClient, func(tx dbmodel.TxInterface) error {
		_, err := tx.Exec(
			queryCreateRole,
			id,
			role.OrganizationUnitID,
			role.Name,
			role.Description,
			s.deploymentID,
		)
		if err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}

		if err := addPermissionsToRole(tx, id, role.Permissions, s.deploymentID); err != nil {
			return err
		}

		if err := addAssignmentsToRole(tx, id, role.Assignments, s.deploymentID); err != nil {
			return err
		}

		return nil
	})
}

// GetRole retrieves a role by its id.
func (s *roleStore) GetRole(id string) (RoleWithPermissions, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return RoleWithPermissions{}, err
	}

	results, err := dbClient.Query(queryGetRoleByID, id, s.deploymentID)
	if err != nil {
		return RoleWithPermissions{}, fmt.Errorf("failed to execute query: %w", err)
	}

	if len(results) == 0 {
		return RoleWithPermissions{}, ErrRoleNotFound
	}

	if len(results) != 1 {
		return RoleWithPermissions{}, fmt.Errorf("unexpected number of results: %d", len(results))
	}

	row := results[0]
	roleBasicInfo, err := buildRoleBasicInfoFromResultRow(row)
	if err != nil {
		return RoleWithPermissions{}, err
	}

	permissions, err := s.getRolePermissions(dbClient, id)
	if err != nil {
		return RoleWithPermissions{}, fmt.Errorf("failed to get role permissions: %w", err)
	}

	return RoleWithPermissions{
		ID:                 roleBasicInfo.ID,
		Name:               roleBasicInfo.Name,
		Description:        roleBasicInfo.Description,
		OrganizationUnitID: roleBasicInfo.OrganizationUnitID,
		Permissions:        permissions,
	}, nil
}

// IsRoleExist checks if a role exists by its ID without fetching its details.
func (s *roleStore) IsRoleExist(id string) (bool, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return false, err
	}

	results, err := dbClient.Query(queryCheckRoleExists, id, s.deploymentID)
	if err != nil {
		return false, fmt.Errorf("failed to check role existence: %w", err)
	}

	return parseBoolFromCount(results)
}

// GetRoleAssignments retrieves assignments for a role with pagination.
func (s *roleStore) GetRoleAssignments(id string, limit, offset int) ([]RoleAssignment, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return nil, err
	}

	results, err := dbClient.Query(queryGetRoleAssignments, id, limit, offset, s.deploymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role assignments: %w", err)
	}

	assignments := make([]RoleAssignment, 0)
	for _, row := range results {
		assigneeID, err := parseStringField(row, "assignee_id")
		if err != nil {
			return nil, err
		}
		assigneeType, err := parseStringField(row, "assignee_type")
		if err != nil {
			return nil, err
		}
		assignments = append(assignments, RoleAssignment{
			ID:   assigneeID,
			Type: AssigneeType(assigneeType),
		})
	}

	return assignments, nil
}

// GetRoleAssignmentsCount retrieves the total count of assignments for a role.
func (s *roleStore) GetRoleAssignmentsCount(id string) (int, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return 0, err
	}

	countResults, err := dbClient.Query(queryGetRoleAssignmentsCount, id, s.deploymentID)
	if err != nil {
		return 0, fmt.Errorf("failed to get role assignments count: %w", err)
	}

	return parseCountResult(countResults)
}

// UpdateRole updates an existing role.
func (s *roleStore) UpdateRole(id string, role RoleUpdateDetail) error {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return err
	}

	return s.executeInTransaction(dbClient, func(tx dbmodel.TxInterface) error {
		result, err := tx.Exec(
			queryUpdateRole,
			role.OrganizationUnitID,
			role.Name,
			role.Description,
			id,
			s.deploymentID,
		)
		if err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return ErrRoleNotFound
		}

		if err := updateRolePermissions(tx, id, role.Permissions, s.deploymentID); err != nil {
			return err
		}

		return nil
	})
}

// DeleteRole deletes a role.
func (s *roleStore) DeleteRole(id string) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, storeLoggerComponentName))

	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return err
	}

	rowsAffected, err := dbClient.Execute(queryDeleteRole, id, s.deploymentID)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	if rowsAffected == 0 {
		logger.Debug("Role not found with id: " + id)
	}

	return nil
}

// AddAssignments adds assignments to a role.
func (s *roleStore) AddAssignments(id string, assignments []RoleAssignment) error {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return err
	}

	return s.executeInTransaction(dbClient, func(tx dbmodel.TxInterface) error {
		return addAssignmentsToRole(tx, id, assignments, s.deploymentID)
	})
}

// RemoveAssignments removes assignments from a role.
func (s *roleStore) RemoveAssignments(id string, assignments []RoleAssignment) error {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return err
	}

	return s.executeInTransaction(dbClient, func(tx dbmodel.TxInterface) error {
		for _, assignment := range assignments {
			_, err := tx.Exec(queryDeleteRoleAssignmentsByIDs, id, assignment.Type, assignment.ID, s.deploymentID)
			if err != nil {
				return fmt.Errorf("failed to remove assignment from role: %w", err)
			}
		}
		return nil
	})
}

// getRolePermissions retrieves all permissions for a role.
func (s *roleStore) getRolePermissions(dbClient provider.DBClientInterface, id string) ([]ResourcePermissions, error) {
	results, err := dbClient.Query(queryGetRolePermissions, id, s.deploymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}

	// Group permissions by resource server
	permMap := make(map[string][]string)
	var resourceServerOrder []string

	for _, row := range results {
		permission, ok := row["permission"].(string)
		if !ok {
			return nil, fmt.Errorf("failed to parse permission as string")
		}

		resourceServerID, ok := row["resource_server_id"].(string)
		if !ok {
			return nil, fmt.Errorf("failed to parse resource_server_id as string")
		}

		// Track order of resource servers as they appear
		if _, exists := permMap[resourceServerID]; !exists {
			resourceServerOrder = append(resourceServerOrder, resourceServerID)
		}

		permMap[resourceServerID] = append(permMap[resourceServerID], permission)
	}

	// Convert map to array of ResourcePermissions
	permissions := make([]ResourcePermissions, 0, len(permMap))
	for _, rsID := range resourceServerOrder {
		permissions = append(permissions, ResourcePermissions{
			ResourceServerID: rsID,
			Permissions:      permMap[rsID],
		})
	}

	return permissions, nil
}

// buildRoleSummaryFromResultRow constructs a Role from a database result row.
func buildRoleBasicInfoFromResultRow(row map[string]interface{}) (Role, error) {
	fields, err := parseStringFields(row, "role_id", "name", "description", "ou_id")
	if err != nil {
		return Role{}, err
	}

	return Role{
		ID:                 fields[0],
		Name:               fields[1],
		Description:        fields[2],
		OrganizationUnitID: fields[3],
	}, nil
}

// addPermissionsToRole adds a list of permissions to a role.
func addPermissionsToRole(
	tx dbmodel.TxInterface,
	id string,
	permissions []ResourcePermissions,
	deploymentID string,
) error {
	for _, resPerm := range permissions {
		for _, permission := range resPerm.Permissions {
			_, err := tx.Exec(queryCreateRolePermission, id, resPerm.ResourceServerID, permission, deploymentID)
			if err != nil {
				return fmt.Errorf("failed to add permission to role: %w", err)
			}
		}
	}
	return nil
}

// addAssignmentsToRole adds a list of assignments to a role.
func addAssignmentsToRole(
	tx dbmodel.TxInterface,
	id string,
	assignments []RoleAssignment,
	deploymentID string,
) error {
	for _, assignment := range assignments {
		_, err := tx.Exec(queryCreateRoleAssignment, id, assignment.Type, assignment.ID, deploymentID)
		if err != nil {
			return fmt.Errorf("failed to add assignment to role: %w", err)
		}
	}
	return nil
}

// updateRolePermissions updates the permissions assigned to the role by first deleting existing permissions and
// then adding new ones.
func updateRolePermissions(
	tx dbmodel.TxInterface,
	id string,
	permissions []ResourcePermissions,
	deploymentID string,
) error {
	_, err := tx.Exec(queryDeleteRolePermissions, id, deploymentID)
	if err != nil {
		return fmt.Errorf("failed to delete existing role permissions: %w", err)
	}

	err = addPermissionsToRole(tx, id, permissions, deploymentID)
	if err != nil {
		return fmt.Errorf("failed to assign permissions to role: %w", err)
	}
	return nil
}

// CheckRoleNameExists checks if a role with the given name exists in the specified organization unit.
func (s *roleStore) CheckRoleNameExists(ouID, name string) (bool, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return false, err
	}

	results, err := dbClient.Query(queryCheckRoleNameExists, ouID, name, s.deploymentID)
	if err != nil {
		return false, fmt.Errorf("failed to check role name existence: %w", err)
	}

	return parseBoolFromCount(results)
}

// CheckRoleNameExistsExcludingID checks if a role with the given name exists in the specified organization unit,
// excluding the role with the given ID.
func (s *roleStore) CheckRoleNameExistsExcludingID(ouID, name, excludeRoleID string) (bool, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return false, err
	}

	results, err := dbClient.Query(queryCheckRoleNameExistsExcludingID, ouID, name, excludeRoleID, s.deploymentID)
	if err != nil {
		return false, fmt.Errorf("failed to check role name existence: %w", err)
	}

	return parseBoolFromCount(results)
}

// GetAuthorizedPermissions retrieves the permissions that a user is authorized for based on their
// direct role assignments and group memberships.
func (s *roleStore) GetAuthorizedPermissions(
	userID string,
	groupIDs []string,
	requestedPermissions []string,
) ([]string, error) {
	dbClient, err := s.getIdentityDBClient()
	if err != nil {
		return nil, err
	}

	// Handle nil groupIDs slice
	if groupIDs == nil {
		groupIDs = []string{}
	}

	// Build dynamic query based on provided parameters
	query, args := buildAuthorizedPermissionsQuery(userID, groupIDs, requestedPermissions, s.deploymentID)

	results, err := dbClient.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get authorized permissions: %w", err)
	}

	permissions := make([]string, 0)
	for _, row := range results {
		if permission, ok := row["permission"].(string); ok {
			permissions = append(permissions, permission)
		}
	}

	return permissions, nil
}

// getIdentityDBClient is a helper method to get the database client for the identity database.
func (s *roleStore) getIdentityDBClient() (provider.DBClientInterface, error) {
	dbClient, err := s.dbProvider.GetConfigDBClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}
	return dbClient, nil
}

// parseCountResult parses a count result from a database query result.
func parseCountResult(results []map[string]interface{}) (int, error) {
	if len(results) == 0 {
		return 0, nil
	}

	if countVal, ok := results[0]["total"].(int64); ok {
		return int(countVal), nil
	}
	return 0, fmt.Errorf("failed to parse total from query result")
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

// parseStringFields extracts multiple string fields from a database result row.
func parseStringFields(row map[string]interface{}, fieldNames ...string) ([]string, error) {
	result := make([]string, len(fieldNames))
	for i, fieldName := range fieldNames {
		value, err := parseStringField(row, fieldName)
		if err != nil {
			return nil, err
		}
		result[i] = value
	}
	return result, nil
}

// executeInTransaction executes a function within a database transaction.
// It automatically handles transaction begin, commit, and rollback on error.
func (s *roleStore) executeInTransaction(
	dbClient provider.DBClientInterface,
	fn func(tx dbmodel.TxInterface) error,
) error {
	tx, err := dbClient.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	err = fn(tx)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Join(err, fmt.Errorf("failed to rollback transaction: %w", rollbackErr))
		}
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
