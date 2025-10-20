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
	GetRoleList(limit, offset int) ([]RoleSummary, error)
	CreateRole(role Role) error
	GetRole(id string) (Role, error)
	GetRoleAssignments(roleID string, limit, offset int) ([]Assignment, error)
	GetRoleAssignmentsCount(roleID string) (int, error)
	UpdateRole(role Role) error
	DeleteRole(id string) error
	AddAssignments(roleID string, assignments []Assignment) error
	RemoveAssignments(roleID string, assignments []Assignment) error
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
func (s *roleStore) GetRoleList(limit, offset int) ([]RoleSummary, error) {
	dbClient, err := s.dbProvider.GetDBClient("identity")
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(queryGetRoleList, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to execute role list query: %w", err)
	}

	roles := make([]RoleSummary, 0)
	for _, row := range results {
		role, err := buildRoleSummaryFromResultRow(row)
		if err != nil {
			return nil, fmt.Errorf("failed to build role from result row: %w", err)
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// CreateRole creates a new role in the database.
func (s *roleStore) CreateRole(role Role) error {
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
		role.ID,
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

	err = addPermissionsToRole(tx, role.ID, role.Permissions)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			err = errors.Join(err, fmt.Errorf("failed to rollback transaction: %w", rollbackErr))
		}
		return err
	}

	err = addAssignmentsToRole(tx, role.ID, role.Assignments)
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
func (s *roleStore) GetRole(id string) (Role, error) {
	dbClient, err := s.dbProvider.GetDBClient("identity")
	if err != nil {
		return Role{}, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(queryGetRoleByID, id)
	if err != nil {
		return Role{}, fmt.Errorf("failed to execute query: %w", err)
	}

	if len(results) == 0 {
		return Role{}, ErrRoleNotFound
	}

	if len(results) != 1 {
		return Role{}, fmt.Errorf("unexpected number of results: %d", len(results))
	}

	row := results[0]
	role, err := buildRoleFromResultRow(row)
	if err != nil {
		return Role{}, err
	}

	permissions, err := s.getRolePermissions(dbClient, id)
	if err != nil {
		return Role{}, fmt.Errorf("failed to get role permissions: %w", err)
	}
	role.Permissions = permissions

	assignments, err := s.getAllRoleAssignments(dbClient, id)
	if err != nil {
		return Role{}, fmt.Errorf("failed to get role assignments: %w", err)
	}
	role.Assignments = assignments

	return role, nil
}

// GetRoleAssignments retrieves assignments for a role with pagination.
func (s *roleStore) GetRoleAssignments(roleID string, limit, offset int) ([]Assignment, error) {
	dbClient, err := s.dbProvider.GetDBClient("identity")
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}

	results, err := dbClient.Query(queryGetRoleAssignments, roleID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get role assignments: %w", err)
	}

	assignments := make([]Assignment, 0)
	for _, row := range results {
		if assigneeID, ok := row["assignee_id"].(string); ok {
			if assigneeType, ok := row["assignee_type"].(string); ok {
				assignments = append(assignments, Assignment{
					ID:   assigneeID,
					Type: AssigneeType(assigneeType),
				})
			}
		}
	}

	return assignments, nil
}

// GetRoleAssignmentsCount retrieves the total count of assignments for a role.
func (s *roleStore) GetRoleAssignmentsCount(roleID string) (int, error) {
	dbClient, err := s.dbProvider.GetDBClient("identity")
	if err != nil {
		return 0, fmt.Errorf("failed to get database client: %w", err)
	}

	countResults, err := dbClient.Query(queryGetRoleAssignmentsCount, roleID)
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
func (s *roleStore) UpdateRole(role Role) error {
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
		role.ID,
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

	err = updateRolePermissions(tx, role.ID, role.Permissions)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			err = errors.Join(err, fmt.Errorf("failed to rollback transaction: %w", rollbackErr))
		}
		return err
	}

	err = updateRoleAssignments(tx, role.ID, role.Assignments)
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
func (s *roleStore) AddAssignments(roleID string, assignments []Assignment) error {
	dbClient, err := s.dbProvider.GetDBClient("identity")
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	tx, err := dbClient.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	err = addAssignmentsToRole(tx, roleID, assignments)
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
func (s *roleStore) RemoveAssignments(roleID string, assignments []Assignment) error {
	dbClient, err := s.dbProvider.GetDBClient("identity")
	if err != nil {
		return fmt.Errorf("failed to get database client: %w", err)
	}

	tx, err := dbClient.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	for _, assignment := range assignments {
		_, err := tx.Exec(queryDeleteRoleAssignmentsByIDs.Query, roleID, assignment.Type, assignment.ID)
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
func (s *roleStore) getRolePermissions(dbClient provider.DBClientInterface, roleID string) ([]string, error) {
	results, err := dbClient.Query(queryGetRolePermissions, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}

	permissions := make([]string, 0)
	for _, row := range results {
		if permission, ok := row["permission"].(string); ok {
			permissions = append(permissions, permission)
		}
	}

	return permissions, nil
}

// getAllRoleAssignments retrieves all assignments for a role (without pagination).
func (s *roleStore) getAllRoleAssignments(dbClient provider.DBClientInterface, roleID string) ([]Assignment, error) {
	results, err := dbClient.Query(queryGetAllRoleAssignments, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role assignments: %w", err)
	}

	assignments := make([]Assignment, 0)
	for _, row := range results {
		if assigneeID, ok := row["assignee_id"].(string); ok {
			if assigneeType, ok := row["assignee_type"].(string); ok {
				assignments = append(assignments, Assignment{
					ID:   assigneeID,
					Type: AssigneeType(assigneeType),
				})
			}
		}
	}

	return assignments, nil
}

// buildRoleSummaryFromResultRow constructs a RoleSummary from a database result row.
func buildRoleSummaryFromResultRow(row map[string]interface{}) (RoleSummary, error) {
	roleID, ok := row["role_id"].(string)
	if !ok {
		return RoleSummary{}, fmt.Errorf("failed to parse role_id as string")
	}

	name, ok := row["name"].(string)
	if !ok {
		return RoleSummary{}, fmt.Errorf("failed to parse name as string")
	}

	description := ""
	if desc, ok := row["description"].(string); ok {
		description = desc
	}

	ouID, ok := row["ou_id"].(string)
	if !ok {
		return RoleSummary{}, fmt.Errorf("failed to parse ou_id as string")
	}

	role := RoleSummary{
		ID:                 roleID,
		Name:               name,
		Description:        description,
		OrganizationUnitID: ouID,
	}

	return role, nil
}

// buildRoleFromResultRow constructs a Role from a database result row.
func buildRoleFromResultRow(row map[string]interface{}) (Role, error) {
	roleID, ok := row["role_id"].(string)
	if !ok {
		return Role{}, fmt.Errorf("failed to parse role_id as string")
	}

	name, ok := row["name"].(string)
	if !ok {
		return Role{}, fmt.Errorf("failed to parse name as string")
	}

	description := ""
	if desc, ok := row["description"].(string); ok {
		description = desc
	}

	ouID, ok := row["ou_id"].(string)
	if !ok {
		return Role{}, fmt.Errorf("failed to parse ou_id as string")
	}

	role := Role{
		ID:                 roleID,
		Name:               name,
		Description:        description,
		OrganizationUnitID: ouID,
	}

	return role, nil
}

// addPermissionsToRole adds a list of permissions to a role.
func addPermissionsToRole(
	tx dbmodel.TxInterface,
	roleID string,
	permissions []string,
) error {
	for _, permission := range permissions {
		_, err := tx.Exec(queryCreateRolePermission.Query, roleID, permission)
		if err != nil {
			return fmt.Errorf("failed to add permission to role: %w", err)
		}
	}
	return nil
}

// addAssignmentsToRole adds a list of assignments to a role.
func addAssignmentsToRole(
	tx dbmodel.TxInterface,
	roleID string,
	assignments []Assignment,
) error {
	for _, assignment := range assignments {
		_, err := tx.Exec(queryCreateRoleAssignment.Query, roleID, assignment.Type, assignment.ID)
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
	roleID string,
	permissions []string,
) error {
	_, err := tx.Exec(queryDeleteRolePermissions.Query, roleID)
	if err != nil {
		return fmt.Errorf("failed to delete existing role permissions: %w", err)
	}

	err = addPermissionsToRole(tx, roleID, permissions)
	if err != nil {
		return fmt.Errorf("failed to assign permissions to role: %w", err)
	}
	return nil
}

// updateRoleAssignments updates the assignments assigned to the role by first deleting existing assignments and
// then adding new ones.
func updateRoleAssignments(
	tx dbmodel.TxInterface,
	roleID string,
	assignments []Assignment,
) error {
	_, err := tx.Exec(queryDeleteRoleAssignments.Query, roleID)
	if err != nil {
		return fmt.Errorf("failed to delete existing role assignments: %w", err)
	}

	err = addAssignmentsToRole(tx, roleID, assignments)
	if err != nil {
		return fmt.Errorf("failed to assign assignments to role: %w", err)
	}
	return nil
}
