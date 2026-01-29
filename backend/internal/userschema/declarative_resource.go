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

package userschema

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
	resourceTypeUserSchema = "user_schema"
	paramTypUserSchema     = "UserSchema"
)

// UserSchemaExporter implements declarative_resource.ResourceExporter for user schemas.
type UserSchemaExporter struct {
	service UserSchemaServiceInterface
}

// newUserSchemaExporter creates a new user schema exporter.
func newUserSchemaExporter(service UserSchemaServiceInterface) *UserSchemaExporter {
	return &UserSchemaExporter{service: service}
}

// NewUserSchemaExporterForTest creates a new user schema exporter for testing purposes.
func NewUserSchemaExporterForTest(service UserSchemaServiceInterface) *UserSchemaExporter {
	return newUserSchemaExporter(service)
}

// GetResourceType returns the resource type for user schemas.
func (e *UserSchemaExporter) GetResourceType() string {
	return resourceTypeUserSchema
}

// GetParameterizerType returns the parameterizer type for user schemas.
func (e *UserSchemaExporter) GetParameterizerType() string {
	return paramTypUserSchema
}

// GetAllResourceIDs retrieves all user schema IDs.
func (e *UserSchemaExporter) GetAllResourceIDs() ([]string, *serviceerror.ServiceError) {
	response, err := e.service.GetUserSchemaList(serverconst.MaxPageSize, 0)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(response.Schemas))
	for _, schema := range response.Schemas {
		ids = append(ids, schema.ID)
	}
	return ids, nil
}

// GetResourceByID retrieves a user schema by its ID.
func (e *UserSchemaExporter) GetResourceByID(id string) (interface{}, string, *serviceerror.ServiceError) {
	schema, err := e.service.GetUserSchema(id)
	if err != nil {
		return nil, "", err
	}
	return schema, schema.Name, nil
}

// ValidateResource validates a user schema resource.
func (e *UserSchemaExporter) ValidateResource(
	resource interface{}, id string, logger *log.Logger,
) (string, *declarativeresource.ExportError) {
	schema, ok := resource.(*UserSchema)
	if !ok {
		return "", declarativeresource.CreateTypeError(resourceTypeUserSchema, id)
	}

	err := declarativeresource.ValidateResourceName(
		schema.Name, resourceTypeUserSchema, id, "SCHEMA_VALIDATION_ERROR", logger,
	)
	if err != nil {
		return "", err
	}

	if len(schema.Schema) == 0 {
		logger.Warn("User schema has no schema definition",
			log.String("schemaID", id), log.String("name", schema.Name))
	}

	return schema.Name, nil
}

// GetResourceRules returns the parameterization rules for user schemas.
func (e *UserSchemaExporter) GetResourceRules() *declarativeresource.ResourceRules {
	return &declarativeresource.ResourceRules{}
}

// loadDeclarativeResources loads declarative user schema resources from files.
func loadDeclarativeResources(
	userSchemaStore userSchemaStoreInterface, ouService oupkg.OrganizationUnitServiceInterface) error {
	// Type assert to access Storer interface for resource loading
	fileBasedStore, ok := userSchemaStore.(*userSchemaFileBasedStore)
	if !ok {
		return fmt.Errorf("failed to assert userSchemaStore to *userSchemaFileBasedStore")
	}

	resourceConfig := declarativeresource.ResourceConfig{
		ResourceType:  "UserSchema",
		DirectoryName: "user_schemas",
		Parser:        parseToUserSchemaDTOWrapper,
		Validator:     validateUserSchemaWrapper(ouService),
		IDExtractor: func(data interface{}) string {
			return data.(*UserSchema).ID
		},
	}

	loader := declarativeresource.NewResourceLoader(resourceConfig, fileBasedStore)
	if err := loader.LoadResources(); err != nil {
		return fmt.Errorf("failed to load user schema resources: %w", err)
	}

	return nil
}

// parseToUserSchemaDTOWrapper wraps parseToUserSchemaDTO to match ResourceConfig.Parser signature.
func parseToUserSchemaDTOWrapper(data []byte) (interface{}, error) {
	return parseToUserSchemaDTO(data)
}

func parseToUserSchemaDTO(data []byte) (*UserSchema, error) {
	var schemaRequest UserSchemaRequestWithID
	err := yaml.Unmarshal(data, &schemaRequest)
	if err != nil {
		return nil, err
	}

	// Validate that schema is valid JSON
	schemaBytes := []byte(schemaRequest.Schema)
	if !json.Valid(schemaBytes) {
		return nil, fmt.Errorf("schema field contains invalid JSON")
	}

	schemaDTO := &UserSchema{
		ID:                    schemaRequest.ID,
		Name:                  schemaRequest.Name,
		OrganizationUnitID:    schemaRequest.OrganizationUnitID,
		AllowSelfRegistration: schemaRequest.AllowSelfRegistration,
		Schema:                []byte(schemaRequest.Schema),
	}

	return schemaDTO, nil
}

// validateUserSchemaWrapper wraps validateUserSchema to match ResourceConfig.Validator signature.
func validateUserSchemaWrapper(ouService oupkg.OrganizationUnitServiceInterface) func(interface{}) error {
	return func(dto interface{}) error {
		schemaDTO, ok := dto.(*UserSchema)
		if !ok {
			return fmt.Errorf("invalid type: expected *UserSchema")
		}
		return validateUserSchema(schemaDTO, ouService)
	}
}

func validateUserSchema(schemaDTO *UserSchema, ouService oupkg.OrganizationUnitServiceInterface) error {
	if strings.TrimSpace(schemaDTO.Name) == "" {
		return fmt.Errorf("user schema name is required")
	}

	if strings.TrimSpace(schemaDTO.ID) == "" {
		return fmt.Errorf("user schema ID is required")
	}

	if strings.TrimSpace(schemaDTO.OrganizationUnitID) == "" {
		return fmt.Errorf("organization unit ID is required for user schema '%s'", schemaDTO.Name)
	}

	// Validate organization unit exists
	_, err := ouService.GetOrganizationUnit(schemaDTO.OrganizationUnitID)
	if err != nil {
		return fmt.Errorf("organization unit '%s' not found for user schema '%s'",
			schemaDTO.OrganizationUnitID, schemaDTO.Name)
	}

	// Validate schema is valid JSON
	if len(schemaDTO.Schema) > 0 {
		var testSchema map[string]interface{}
		if err := json.Unmarshal(schemaDTO.Schema, &testSchema); err != nil {
			return fmt.Errorf("invalid schema JSON for user schema '%s': %w", schemaDTO.Name, err)
		}
	}

	return nil
}
