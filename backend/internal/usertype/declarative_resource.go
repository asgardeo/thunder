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

package usertype

import (
	"encoding/json"
	"fmt"
	"strings"

	oupkg "github.com/asgardeo/thunder/internal/ou"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"

	"gopkg.in/yaml.v3"
)

const (
	resourceTypeUserType = "user_type"
	paramTypeUserType    = "UserType"
)

// UserTypeExporter implements declarative_resource.ResourceExporter for user types.
type UserTypeExporter struct {
	service UserTypeServiceInterface
}

// newUserTypeExporter creates a new user type exporter.
func newUserTypeExporter(service UserTypeServiceInterface) *UserTypeExporter {
	return &UserTypeExporter{service: service}
}

// NewUserTypeExporterForTest creates a new user type exporter for testing purposes.
func NewUserTypeExporterForTest(service UserTypeServiceInterface) *UserTypeExporter {
	return newUserTypeExporter(service)
}

// GetResourceType returns the resource type for user types.
func (e *UserTypeExporter) GetResourceType() string {
	return resourceTypeUserType
}

// GetParameterizerType returns the parameterizer type for user types.
func (e *UserTypeExporter) GetParameterizerType() string {
	return paramTypeUserType
}

// GetAllResourceIDs retrieves all user type IDs.
func (e *UserTypeExporter) GetAllResourceIDs() ([]string, *serviceerror.ServiceError) {
	response, err := e.service.GetUserTypeList(serverconst.MaxPageSize, 0)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(response.Types))
	for _, schema := range response.Types {
		ids = append(ids, schema.ID)
	}
	return ids, nil
}

// GetResourceByID retrieves a user type by its ID.
func (e *UserTypeExporter) GetResourceByID(id string) (interface{}, string, *serviceerror.ServiceError) {
	schema, err := e.service.GetUserType(id)
	if err != nil {
		return nil, "", err
	}
	return schema, schema.Name, nil
}

// ValidateResource validates a user type resource.
func (e *UserTypeExporter) ValidateResource(
	resource interface{}, id string, logger *log.Logger,
) (string, *declarativeresource.ExportError) {
	schema, ok := resource.(*UserType)
	if !ok {
		return "", declarativeresource.CreateTypeError(resourceTypeUserType, id)
	}

	err := declarativeresource.ValidateResourceName(
		schema.Name, resourceTypeUserType, id, "SCHEMA_VALIDATION_ERROR", logger,
	)
	if err != nil {
		return "", err
	}

	if len(schema.Schema) == 0 {
		logger.Warn("User type has no schema definition",
			log.String("schemaID", id), log.String("name", schema.Name))
	}

	return schema.Name, nil
}

// GetResourceRules returns the parameterization rules for user types.
func (e *UserTypeExporter) GetResourceRules() *declarativeresource.ResourceRules {
	return &declarativeresource.ResourceRules{}
}

// loadDeclarativeResources loads declarative user type resources from files.
func loadDeclarativeResources(
	userTypeStore userTypeStoreInterface, ouService oupkg.OrganizationUnitServiceInterface) error {
	// Type assert to access Storer interface for resource loading
	fileBasedStore, ok := userTypeStore.(*userTypeFileBasedStore)
	if !ok {
		return fmt.Errorf("failed to assert userTypeStore to *userTypeFileBasedStore")
	}

	resourceConfig := declarativeresource.ResourceConfig{
		ResourceType:  "UserType",
		DirectoryName: "user_types",
		Parser:        parseToUserTypeDTOWrapper,
		Validator:     validateUserTypeWrapper(ouService),
		IDExtractor: func(data interface{}) string {
			return data.(*UserType).ID
		},
	}

	loader := declarativeresource.NewResourceLoader(resourceConfig, fileBasedStore)
	if err := loader.LoadResources(); err != nil {
		return fmt.Errorf("failed to load user type resources: %w", err)
	}

	return nil
}

// parseToUserTypeDTOWrapper wraps parseToUserTypeDTO to match ResourceConfig.Parser signature.
func parseToUserTypeDTOWrapper(data []byte) (interface{}, error) {
	return parseToUserTypeDTO(data)
}

func parseToUserTypeDTO(data []byte) (*UserType, error) {
	var schemaRequest UserTypeRequestWithID
	err := yaml.Unmarshal(data, &schemaRequest)
	if err != nil {
		return nil, err
	}

	// Validate that schema is valid JSON
	schemaBytes := []byte(schemaRequest.Schema)
	if !json.Valid(schemaBytes) {
		return nil, fmt.Errorf("schema field contains invalid JSON")
	}

	schemaDTO := &UserType{
		ID:                    schemaRequest.ID,
		Name:                  schemaRequest.Name,
		OrganizationUnitID:    schemaRequest.OrganizationUnitID,
		AllowSelfRegistration: schemaRequest.AllowSelfRegistration,
		Schema:                []byte(schemaRequest.Schema),
	}

	return schemaDTO, nil
}

// validateUserTypeWrapper wraps validateUserType to match ResourceConfig.Validator signature.
func validateUserTypeWrapper(ouService oupkg.OrganizationUnitServiceInterface) func(interface{}) error {
	return func(dto interface{}) error {
		schemaDTO, ok := dto.(*UserType)
		if !ok {
			return fmt.Errorf("invalid type: expected *UserType")
		}
		return validateUserType(schemaDTO, ouService)
	}
}

func validateUserType(schemaDTO *UserType, ouService oupkg.OrganizationUnitServiceInterface) error {
	if strings.TrimSpace(schemaDTO.Name) == "" {
		return fmt.Errorf("user type name is required")
	}

	if strings.TrimSpace(schemaDTO.ID) == "" {
		return fmt.Errorf("user type ID is required")
	}

	if strings.TrimSpace(schemaDTO.OrganizationUnitID) == "" {
		return fmt.Errorf("organization unit ID is required for user type '%s'", schemaDTO.Name)
	}

	// Validate organization unit exists
	_, err := ouService.GetOrganizationUnit(schemaDTO.OrganizationUnitID)
	if err != nil {
		return fmt.Errorf("organization unit '%s' not found for user type '%s'",
			schemaDTO.OrganizationUnitID, schemaDTO.Name)
	}

	// Validate schema is valid JSON
	if len(schemaDTO.Schema) > 0 {
		var testSchema map[string]interface{}
		if err := json.Unmarshal(schemaDTO.Schema, &testSchema); err != nil {
			return fmt.Errorf("invalid schema JSON for user type '%s': %w", schemaDTO.Name, err)
		}
	}

	return nil
}
