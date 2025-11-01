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
}

// roleStore is the default implementation of roleStoreInterface.
type roleStore struct {
	dbProvider provider.DBProviderInterface
}

// newRoleStore creates a new instance of roleStore.
func newRoleStore() roleStoreInterface {
	return &roleStore{
		dbProvider: provider.GetDBProvider(),
	}
}

// GetRoleListCount retrieves the total count of roles.
func (s *roleStore) GetRoleListCount() (int, error) {
	dbClient, err := s.dbProvider.GetDBClient("identity")
	if err != nil {
		return 0, fmt.Errorf("failed to get database client: %w", err)
	}

	countResults, err := dbClient.Query(queryGetRoleListCount)
	if err != nil {
		return 0, fmt.Errorf("failed to execute count query: %w", err)
	}

	totalCount := 0
	if len(countResults) > 0 {
		if total, ok := countResults[0]["total"].(int64); ok {
			totalCount = int(total)
		} else {
			return 0, fmt.Errorf("failed to parse total count from query result")
		}
	}
	return totalCount, nil
}

// GetRoleList retrieves roles with pagination.
func (s *roleStore) GetRoleList(limit, offset int) ([]Role, error) {
	dbClient, err := s.dbProvider.GetDBClient("identity")
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(queryGetRoleList, limit, offset)
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
	dbClient, err := s.dbProvider.GetDBClient("identity")
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	tx, err := dbClient.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	_, err = tx.Exec(
		queryCreateRole.Query,
		id,
		role.OrganizationUnitID,
		role.Name,
		role.Description,
	)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			err = errors.Join(err, fmt.Errorf("failed to rollback transaction: %w", rollbackErr))
		}
		return fmt.Errorf("failed to execute query: %w", err)
	}

	err = addPermissionsToRole(tx, id, role.Permissions)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			err = errors.Join(err, fmt.Errorf("failed to rollback transaction: %w", rollbackErr))
		}
		return err
	}

	err = addAssignmentsToRole(tx, id, role.Assignments)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			err = errors.Join(err, fmt.Errorf("failed to rollback transaction: %w", rollbackErr))
		}
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetRole retrieves a role by its id.
func (s *roleStore) GetRole(id string) (RoleWithPermissions, error) {
	dbClient, err := s.dbProvider.GetDBClient("identity")
	if err != nil {
		return RoleWithPermissions{}, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(queryGetRoleByID, id)
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
	dbClient, err := s.dbProvider.GetDBClient("identity")
	if err != nil {
		return false, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(queryCheckRoleExists, id)
	if err != nil {
		return false, fmt.Errorf("failed to check role existence: %w", err)
	}

	if len(results) == 0 {
		return false, nil
	}

	count := int64(0)
	if countVal, ok := results[0]["count"].(int64); ok {
		count = countVal
	} else {
		return false, fmt.Errorf("failed to parse count from query result")
	}

	return count > 0, nil
}

// GetRoleAssignments retrieves assignments for a role with pagination.
func (s *roleStore) GetRoleAssignments(id string, limit, offset int) ([]RoleAssignment, error) {
	dbClient, err := s.dbProvider.GetDBClient("identity")
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(queryGetRoleAssignments, id, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get role assignments: %w", err)
	}

	assignments := make([]RoleAssignment, 0)
	for _, row := range results {
		assigneeID, ok := row["assignee_id"].(string)
		if !ok {
			return nil, fmt.Errorf("failed to parse assignee_id as string")
		}
		assigneeType, ok := row["assignee_type"].(string)
		if !ok {
			return nil, fmt.Errorf("failed to parse assignee_type as string")
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
	dbClient, err := s.dbProvider.GetDBClient("identity")
	if err != nil {
		return 0, fmt.Errorf("failed to get database client: %w", err)
	}

	countResults, err := dbClient.Query(queryGetRoleAssignmentsCount, id)
	if err != nil {
		return 0, fmt.Errorf("failed to get role assignments count: %w", err)
	}

	totalCount := 0
	if len(countResults) > 0 {
		if total, ok := countResults[0]["total"].(int64); ok {
			totalCount = int(total)
		} else {
			return 0, fmt.Errorf("failed to parse total count from query result")
		}
	}
	return totalCount, nil
}

// UpdateRole updates an existing role.
func (s *roleStore) UpdateRole(id string, role RoleUpdateDetail) error {
	dbClient, err := s.dbProvider.GetDBClient("identity")
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	tx, err := dbClient.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	result, err := tx.Exec(
		queryUpdateRole.Query,
		role.OrganizationUnitID,
		role.Name,
		role.Description,
		id,
	)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			err = errors.Join(err, fmt.Errorf("failed to rollback transaction: %w", rollbackErr))
		}
		return fmt.Errorf("failed to execute query: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			err = errors.Join(err, fmt.Errorf("failed to rollback transaction: %w", rollbackErr))
		}
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("failed to rollback transaction: %w", rollbackErr)
		}
		return ErrRoleNotFound
	}

	err = updateRolePermissions(tx, id, role.Permissions)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			err = errors.Join(err, fmt.Errorf("failed to rollback transaction: %w", rollbackErr))
		}
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteRole deletes a role.
func (s *roleStore) DeleteRole(id string) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, storeLoggerComponentName))

	dbClient, err := s.dbProvider.GetDBClient("identity")
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	rowsAffected, err := dbClient.Execute(queryDeleteRole, id)
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
	dbClient, err := s.dbProvider.GetDBClient("identity")
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	tx, err := dbClient.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	err = addAssignmentsToRole(tx, id, assignments)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			err = errors.Join(err, fmt.Errorf("failed to rollback transaction: %w", rollbackErr))
		}
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// RemoveAssignments removes assignments from a role.
func (s *roleStore) RemoveAssignments(id string, assignments []RoleAssignment) error {
	dbClient, err := s.dbProvider.GetDBClient("identity")
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	tx, err := dbClient.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	for _, assignment := range assignments {
		_, err := tx.Exec(queryDeleteRoleAssignmentsByIDs.Query, id, assignment.Type, assignment.ID)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				err = errors.Join(err, fmt.Errorf("failed to rollback transaction: %w", rollbackErr))
			}
			return fmt.Errorf("failed to remove assignment from role: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// getRolePermissions retrieves all permissions for a role.
func (s *roleStore) getRolePermissions(dbClient provider.DBClientInterface, id string) ([]string, error) {
	results, err := dbClient.Query(queryGetRolePermissions, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}

	permissions := make([]string, 0)
	for _, row := range results {
		permission, ok := row["permission"].(string)
		if !ok {
			return nil, fmt.Errorf("failed to parse permission as string")
		}
		permissions = append(permissions, permission)
	}

	return permissions, nil
}

// buildRoleSummaryFromResultRow constructs a Role from a database result row.
func buildRoleBasicInfoFromResultRow(row map[string]interface{}) (Role, error) {
	id, ok := row["role_id"].(string)
	if !ok {
		return Role{}, fmt.Errorf("failed to parse role_id as string")
	}

	name, ok := row["name"].(string)
	if !ok {
		return Role{}, fmt.Errorf("failed to parse name as string")
	}

	description, ok := row["description"].(string)
	if !ok {
		return Role{}, fmt.Errorf("failed to parse description as string")
	}

	ouID, ok := row["ou_id"].(string)
	if !ok {
		return Role{}, fmt.Errorf("failed to parse ou_id as string")
	}

	role := Role{
		ID:                 id,
		Name:               name,
		Description:        description,
		OrganizationUnitID: ouID,
	}

	return role, nil
}

// addPermissionsToRole adds a list of permissions to a role.
func addPermissionsToRole(
	tx dbmodel.TxInterface,
	id string,
	permissions []string,
) error {
	for _, permission := range permissions {
		_, err := tx.Exec(queryCreateRolePermission.Query, id, permission)
		if err != nil {
			return fmt.Errorf("failed to add permission to role: %w", err)
		}
	}
	return nil
}

// addAssignmentsToRole adds a list of assignments to a role.
func addAssignmentsToRole(
	tx dbmodel.TxInterface,
	id string,
	assignments []RoleAssignment,
) error {
	for _, assignment := range assignments {
		_, err := tx.Exec(queryCreateRoleAssignment.Query, id, assignment.Type, assignment.ID)
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
	permissions []string,
) error {
	_, err := tx.Exec(queryDeleteRolePermissions.Query, id)
	if err != nil {
		return fmt.Errorf("failed to delete existing role permissions: %w", err)
	}

	err = addPermissionsToRole(tx, id, permissions)
	if err != nil {
		return fmt.Errorf("failed to assign permissions to role: %w", err)
	}
	return nil
}

// CheckRoleNameExists checks if a role with the given name exists in the specified organization unit.
func (s *roleStore) CheckRoleNameExists(ouID, name string) (bool, error) {
	dbClient, err := s.dbProvider.GetDBClient("identity")
	if err != nil {
		return false, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(queryCheckRoleNameExists, ouID, name)
	if err != nil {
		return false, fmt.Errorf("failed to check role name existence: %w", err)
	}

	if len(results) == 0 {
		return false, nil
	}

	count := int64(0)
	if countVal, ok := results[0]["count"].(int64); ok {
		count = countVal
	} else {
		return false, fmt.Errorf("failed to parse count from query result")
	}

	return count > 0, nil
}

// CheckRoleNameExistsExcludingID checks if a role with the given name exists in the specified organization unit,
// excluding the role with the given ID.
func (s *roleStore) CheckRoleNameExistsExcludingID(ouID, name, excludeRoleID string) (bool, error) {
	dbClient, err := s.dbProvider.GetDBClient("identity")
	if err != nil {
		return false, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(queryCheckRoleNameExistsExcludingID, ouID, name, excludeRoleID)
	if err != nil {
		return false, fmt.Errorf("failed to check role name existence: %w", err)
	}

	if len(results) == 0 {
		return false, nil
	}

	count := int64(0)
	if countVal, ok := results[0]["count"].(int64); ok {
		count = countVal
	} else {
		return false, fmt.Errorf("failed to parse count from query result")
	}

	return count > 0, nil
}
