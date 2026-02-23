/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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
	"context"
	"fmt"

	"gopkg.in/yaml.v3"

	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
)

const (
	resourceTypeRole = "role"
	paramTypeRole    = "Role"
)

// RoleExporter implements declarativeresource.ResourceExporter for roles.
type RoleExporter struct {
	service RoleServiceInterface
}

// newRoleExporter creates a new role exporter.
func newRoleExporter(service RoleServiceInterface) *RoleExporter {
	return &RoleExporter{service: service}
}

// GetResourceType returns the resource type for roles.
func (e *RoleExporter) GetResourceType() string {
	return resourceTypeRole
}

// GetParameterizerType returns the parameterizer type for roles.
func (e *RoleExporter) GetParameterizerType() string {
	return paramTypeRole
}

// GetAllResourceIDs retrieves all role IDs from the database store.
// In composite mode, this excludes declarative (YAML-based) roles.
func (e *RoleExporter) GetAllResourceIDs(ctx context.Context) ([]string, *serviceerror.ServiceError) {
	offset := 0
	limit := serverconst.MaxPageSize
	ids := []string{}

	for {
		roles, err := e.service.GetRoleList(ctx, limit, offset)
		if err != nil {
			return nil, err
		}

		for _, role := range roles.Roles {
			isDeclarative, svcErr := e.service.IsRoleDeclarative(ctx, role.ID)
			if svcErr != nil {
				return nil, svcErr
			}
			if !isDeclarative {
				ids = append(ids, role.ID)
			}
		}

		offset += len(roles.Roles)

		// Continue fetching while we get results; stop only on empty page
		if len(roles.Roles) == 0 {
			break
		}
	}

	return ids, nil
}

// GetResourceByID retrieves a role by its ID.
func (e *RoleExporter) GetResourceByID(
	ctx context.Context, id string) (interface{}, string, *serviceerror.ServiceError) {
	roleWithPermissions, err := e.service.GetRoleWithPermissions(ctx, id)
	if err != nil {
		return nil, "", err
	}

	assignments, err := e.getAllRoleAssignments(ctx, id)
	if err != nil {
		return nil, "", err
	}

	role := &RoleWithPermissionsAndAssignments{
		ID:                 roleWithPermissions.ID,
		Name:               roleWithPermissions.Name,
		Description:        roleWithPermissions.Description,
		OrganizationUnitID: roleWithPermissions.OrganizationUnitID,
		Permissions:        roleWithPermissions.Permissions,
		Assignments:        assignments,
	}

	return role, role.Name, nil
}

// ValidateResource validates a role resource.
func (e *RoleExporter) ValidateResource(
	resource interface{}, id string, logger *log.Logger,
) (string, *declarativeresource.ExportError) {
	role, ok := resource.(*RoleWithPermissionsAndAssignments)
	if !ok {
		return "", declarativeresource.CreateTypeError(resourceTypeRole, id)
	}

	if err := declarativeresource.ValidateResourceName(
		role.Name, resourceTypeRole, id, "ROLE_VALIDATION_ERROR", logger); err != nil {
		return "", err
	}

	return role.Name, nil
}

// GetResourceRules returns the parameterization rules for roles.
func (e *RoleExporter) GetResourceRules() *declarativeresource.ResourceRules {
	return &declarativeresource.ResourceRules{
		Variables:      []string{},
		ArrayVariables: []string{},
	}
}

// loadDeclarativeResources loads immutable role resources from files.
// The dbStore parameter is optional (can be nil) and is used for duplicate checking in composite mode.
func loadDeclarativeResources(fileStore *fileBasedStore, dbStore roleStoreInterface) error {
	resourceConfig := declarativeresource.ResourceConfig{
		ResourceType:  "Role",
		DirectoryName: "roles",
		Parser:        parseToRoleWrapper,
		Validator: func(data interface{}) error {
			return validateRoleWrapper(data, fileStore, dbStore)
		},
		IDExtractor: func(data interface{}) string {
			// Use safe type assertion to prevent panic
			if v, ok := data.(*RoleWithPermissionsAndAssignments); ok {
				return v.ID
			}
			// Log error and return empty string if type assertion fails
			log.GetLogger().Error("IDExtractor: type assertion failed for RoleWithPermissionsAndAssignments")
			return ""
		},
	}

	loader := declarativeresource.NewResourceLoader(resourceConfig, fileStore)
	if err := loader.LoadResources(); err != nil {
		return fmt.Errorf("failed to load role resources: %w", err)
	}

	return nil
}

// parseToRoleWrapper wraps parseToRole to match the expected signature.
func parseToRoleWrapper(data []byte) (interface{}, error) {
	return parseToRole(data)
}

type roleDeclarativePermission ResourcePermissions

type roleDeclarativeResource struct {
	ID                 string                      `yaml:"id"`
	Name               string                      `yaml:"name"`
	Description        string                      `yaml:"description,omitempty"`
	OrganizationUnitID string                      `yaml:"ou_id"`
	Permissions        []roleDeclarativePermission `yaml:"permissions"`
	Assignments        []RoleAssignment            `yaml:"assignments,omitempty"`
}

// toResourcePermissions converts roleDeclarativePermission to ResourcePermissions.
func toResourcePermissions(perm roleDeclarativePermission) ResourcePermissions {
	return ResourcePermissions(perm)
}

// parseToRole parses YAML data to RoleWithPermissionsAndAssignments.
func parseToRole(data []byte) (*RoleWithPermissionsAndAssignments, error) {
	var roleResource roleDeclarativeResource
	if err := yaml.Unmarshal(data, &roleResource); err != nil {
		return nil, err
	}

	permissions := make([]ResourcePermissions, 0, len(roleResource.Permissions))
	for _, perm := range roleResource.Permissions {
		permissions = append(permissions, toResourcePermissions(perm))
	}

	role := &RoleWithPermissionsAndAssignments{
		ID:                 roleResource.ID,
		Name:               roleResource.Name,
		Description:        roleResource.Description,
		OrganizationUnitID: roleResource.OrganizationUnitID,
		Permissions:        permissions,
		Assignments:        roleResource.Assignments,
	}

	return role, nil
}

// validateRoleWrapper validates role declarative resources and checks for duplicates.
func validateRoleWrapper(data interface{}, fileStore *fileBasedStore, dbStore roleStoreInterface) error {
	role, ok := data.(*RoleWithPermissionsAndAssignments)
	if !ok {
		return fmt.Errorf("invalid type: expected *RoleWithPermissionsAndAssignments")
	}

	if role.ID == "" {
		return fmt.Errorf("role ID is required")
	}
	if role.Name == "" {
		return fmt.Errorf("role name is required")
	}
	if role.OrganizationUnitID == "" {
		return fmt.Errorf("organization unit ID is required")
	}

	for _, assignment := range role.Assignments {
		if assignment.ID == "" {
			return fmt.Errorf("assignment ID is required")
		}
		if assignment.Type != AssigneeTypeUser && assignment.Type != AssigneeTypeGroup {
			return fmt.Errorf("invalid assignment type '%s'", assignment.Type)
		}
	}

	for _, resourcePerms := range role.Permissions {
		if resourcePerms.ResourceServerID == "" {
			return fmt.Errorf("resource server ID is required")
		}
	}

	if fileStore != nil {
		if existingData, err := fileStore.GenericFileBasedStore.Get(role.ID); err == nil && existingData != nil {
			return fmt.Errorf("duplicate role ID '%s': role already exists in declarative resources", role.ID)
		}
	}

	if dbStore != nil {
		exists, err := dbStore.IsRoleExist(context.Background(), role.ID)
		if err != nil {
			// Fail loudly on DB errors during duplicate check
			return fmt.Errorf("checking role existence for '%s': %w", role.ID, err)
		}
		if exists {
			return fmt.Errorf("duplicate role ID '%s': role already exists in the database store", role.ID)
		}
	}

	return nil
}

func (e *RoleExporter) getAllRoleAssignments(
	ctx context.Context,
	roleID string,
) ([]RoleAssignment, *serviceerror.ServiceError) {
	offset := 0
	limit := serverconst.MaxPageSize
	assignments := []RoleAssignment{}

	for {
		list, err := e.service.GetRoleAssignments(ctx, roleID, limit, offset, false)
		if err != nil {
			return nil, err
		}

		for _, assignment := range list.Assignments {
			assignments = append(assignments, RoleAssignment{
				ID:   assignment.ID,
				Type: assignment.Type,
			})
		}

		offset += len(list.Assignments)

		// Continue fetching while we get results; stop only on empty page
		if len(list.Assignments) == 0 {
			break
		}
	}

	return assignments, nil
}
