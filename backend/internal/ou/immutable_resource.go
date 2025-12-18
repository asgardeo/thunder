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
	"testing"

	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	immutableresource "github.com/asgardeo/thunder/internal/system/immutable_resource"
	"github.com/asgardeo/thunder/internal/system/log"

	"gopkg.in/yaml.v3"
)

const (
	resourceTypeOU = "organization_unit"
	paramTypeOU    = "OrganizationUnit"
)

// OUExporter implements immutableresource.ResourceExporter for organization units.
type OUExporter struct {
	service OrganizationUnitServiceInterface
}

// newOUExporter creates a new OU exporter.
func newOUExporter(service OrganizationUnitServiceInterface) *OUExporter {
	return &OUExporter{service: service}
}

// NewOUExporterForTest creates a new OU exporter for testing purposes.
func NewOUExporterForTest(service OrganizationUnitServiceInterface) *OUExporter {
	if !testing.Testing() {
		panic("only for tests!")
	}
	return newOUExporter(service)
}

// GetResourceType returns the resource type for organization units.
func (e *OUExporter) GetResourceType() string {
	return resourceTypeOU
}

// GetParameterizerType returns the parameterizer type for organization units.
func (e *OUExporter) GetParameterizerType() string {
	return paramTypeOU
}

// GetAllResourceIDs retrieves all organization unit IDs.
func (e *OUExporter) GetAllResourceIDs() ([]string, *serviceerror.ServiceError) {
	// Get all OUs by requesting a large limit
	ous, err := e.service.GetOrganizationUnitList(1000, 0)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(ous.OrganizationUnits))
	for _, ou := range ous.OrganizationUnits {
		ids = append(ids, ou.ID)
	}

	// Also get all child OUs recursively
	allIDs := make(map[string]bool)
	for _, id := range ids {
		allIDs[id] = true
		childIDs, err := e.getAllChildIDs(id)
		if err != nil {
			return nil, err
		}
		for _, childID := range childIDs {
			allIDs[childID] = true
		}
	}

	result := make([]string, 0, len(allIDs))
	for id := range allIDs {
		result = append(result, id)
	}

	return result, nil
}

// getAllChildIDs recursively retrieves all child OU IDs.
func (e *OUExporter) getAllChildIDs(parentID string) ([]string, *serviceerror.ServiceError) {
	children, err := e.service.GetOrganizationUnitChildren(parentID, 1000, 0)
	if err != nil {
		return nil, err
	}

	allIDs := []string{}
	for _, child := range children.OrganizationUnits {
		allIDs = append(allIDs, child.ID)
		grandchildIDs, err := e.getAllChildIDs(child.ID)
		if err != nil {
			return nil, err
		}
		allIDs = append(allIDs, grandchildIDs...)
	}

	return allIDs, nil
}

// GetResourceByID retrieves an organization unit by its ID.
func (e *OUExporter) GetResourceByID(id string) (interface{}, string, *serviceerror.ServiceError) {
	ou, err := e.service.GetOrganizationUnit(id)
	if err != nil {
		return nil, "", err
	}
	return &ou, ou.Name, nil
}

// ValidateResource validates an organization unit resource.
func (e *OUExporter) ValidateResource(
	resource interface{}, id string, logger *log.Logger,
) (string, *immutableresource.ExportError) {
	ou, ok := resource.(*OrganizationUnit)
	if !ok {
		return "", immutableresource.CreateTypeError(resourceTypeOU, id)
	}

	if err := immutableresource.ValidateResourceName(
		ou.Name, resourceTypeOU, id, "OU_VALIDATION_ERROR", logger); err != nil {
		return "", err
	}

	return ou.Name, nil
}

// GetResourceRules returns the parameterization rules for organization units.
func (e *OUExporter) GetResourceRules() *immutableresource.ResourceRules {
	// OUs typically don't have parameterizable fields
	return &immutableresource.ResourceRules{
		Variables:      []string{},
		ArrayVariables: []string{},
	}
}

// loadImmutableResources loads immutable organization unit resources from files.
func loadImmutableResources(ouStore organizationUnitStoreInterface) error {
	// Type assert to access Storer interface for resource loading
	fileBasedStore, ok := ouStore.(*fileBasedStore)
	if !ok {
		return fmt.Errorf("failed to assert ouStore to *fileBasedStore")
	}

	resourceConfig := immutableresource.ResourceConfig{
		ResourceType:  "OrganizationUnit",
		DirectoryName: "organizational_units",
		Parser:        parseToOUWrapper,
		Validator:     validateOUWrapper,
		IDExtractor: func(data interface{}) string {
			return data.(*OrganizationUnit).ID
		},
	}

	loader := immutableresource.NewResourceLoader(resourceConfig, fileBasedStore)
	if err := loader.LoadResources(); err != nil {
		return fmt.Errorf("failed to load organization unit resources: %w", err)
	}

	return nil
}

// parseToOUWrapper wraps parseToOU to match the expected signature.
func parseToOUWrapper(data []byte) (interface{}, error) {
	return parseToOU(data)
}

// parseToOU parses YAML data to OrganizationUnit.
func parseToOU(data []byte) (*OrganizationUnit, error) {
	var ouRequest struct {
		ID          string  `yaml:"id"`
		Handle      string  `yaml:"handle"`
		Name        string  `yaml:"name"`
		Description string  `yaml:"description,omitempty"`
		Parent      *string `yaml:"parent,omitempty"`
	}

	err := yaml.Unmarshal(data, &ouRequest)
	if err != nil {
		return nil, err
	}

	ou := &OrganizationUnit{
		ID:          ouRequest.ID,
		Handle:      ouRequest.Handle,
		Name:        ouRequest.Name,
		Description: ouRequest.Description,
		Parent:      ouRequest.Parent,
	}

	return ou, nil
}

// validateOUWrapper wraps validateOU to match ResourceConfig.Validator signature.
func validateOUWrapper(data interface{}) error {
	ou, ok := data.(*OrganizationUnit)
	if !ok {
		return fmt.Errorf("invalid type: expected *OrganizationUnit")
	}

	if ou.ID == "" {
		return fmt.Errorf("organization unit ID is required")
	}

	if ou.Name == "" {
		return fmt.Errorf("organization unit name is required")
	}

	if ou.Handle == "" {
		return fmt.Errorf("organization unit handle is required")
	}

	return nil
}
