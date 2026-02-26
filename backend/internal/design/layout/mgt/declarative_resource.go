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

package layoutmgt

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"

	"gopkg.in/yaml.v3"
)

const (
	resourceTypeLayout = "layout"
	paramTypeLayout    = "Layout"
)

// layoutExporter implements declarativeresource.ResourceExporter for layouts.
type layoutExporter struct {
	service LayoutMgtServiceInterface
}

// newLayoutExporter creates a new layout exporter.
func newLayoutExporter(service LayoutMgtServiceInterface) *layoutExporter {
	return &layoutExporter{service: service}
}

// GetResourceType returns the resource type for layouts.
func (e *layoutExporter) GetResourceType() string {
	return resourceTypeLayout
}

// GetParameterizerType returns the parameterizer type for layouts.
func (e *layoutExporter) GetParameterizerType() string {
	return paramTypeLayout
}

// GetAllResourceIDs retrieves all layout IDs.
func (e *layoutExporter) GetAllResourceIDs(ctx context.Context) ([]string, *serviceerror.ServiceError) {
	layoutList, err := e.service.GetLayoutList(100, 0) // Get a large number to fetch all
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(layoutList.Layouts))
	for _, layout := range layoutList.Layouts {
		ids = append(ids, layout.ID)
	}
	return ids, nil
}

// GetResourceByID retrieves a layout by its ID.
func (e *layoutExporter) GetResourceByID(ctx context.Context, id string) (
	interface{}, string, *serviceerror.ServiceError,
) {
	layout, err := e.service.GetLayout(id)
	if err != nil {
		return nil, "", err
	}
	return layout, layout.DisplayName, nil
}

// ValidateResource validates a layout resource.
func (e *layoutExporter) ValidateResource(
	resource interface{}, id string, logger *log.Logger,
) (string, *declarativeresource.ExportError) {
	layout, ok := resource.(*Layout)
	if !ok {
		return "", declarativeresource.CreateTypeError(resourceTypeLayout, id)
	}

	err := declarativeresource.ValidateResourceName(
		layout.DisplayName, resourceTypeLayout, id, "LAYOUT_VALIDATION_ERROR", logger,
	)
	if err != nil {
		return "", err
	}

	if len(layout.Layout) == 0 {
		logger.Warn("Layout has no layout configuration",
			log.String("layoutID", id), log.String("displayName", layout.DisplayName))
	}

	return layout.DisplayName, nil
}

// GetResourceRules returns the parameterization rules for layouts.
func (e *layoutExporter) GetResourceRules() *declarativeresource.ResourceRules {
	return &declarativeresource.ResourceRules{}
}

// loadDeclarativeResources loads declarative layout resources from files.
func loadDeclarativeResources(layoutStore layoutMgtStoreInterface) error {
	// Type assert to access Storer interface for resource loading
	fileBasedStore, ok := layoutStore.(*layoutFileBasedStore)
	if !ok {
		return fmt.Errorf("failed to assert layoutStore to *layoutFileBasedStore")
	}

	resourceConfig := declarativeresource.ResourceConfig{
		ResourceType:  "Layout",
		DirectoryName: "layouts",
		Parser:        parseToLayoutWrapper,
		Validator:     validateLayoutWrapper,
		IDExtractor: func(data interface{}) string {
			if layout, ok := data.(*Layout); ok {
				return layout.ID
			}
			return ""
		},
	}

	loader := declarativeresource.NewResourceLoader(resourceConfig, fileBasedStore)
	if err := loader.LoadResources(); err != nil {
		return fmt.Errorf("failed to load layout resources: %w", err)
	}

	return nil
}

// parseToLayoutWrapper wraps parseToLayout to match ResourceConfig.Parser signature.
func parseToLayoutWrapper(data []byte) (interface{}, error) {
	return parseToLayout(data)
}

// parseToLayout converts YAML data into a Layout object.
func parseToLayout(data []byte) (*Layout, error) {
	var layoutRequest layoutRequestWithID

	err := yaml.Unmarshal(data, &layoutRequest)
	if err != nil {
		return nil, err
	}

	// Convert layout to JSON bytes
	var layoutJSON json.RawMessage
	if layoutRequest.Layout != nil {
		// Handle both map structure and string format
		switch v := layoutRequest.Layout.(type) {
		case string:
			// JSON string format
			layoutJSON = []byte(v)
		default:
			// Map structure - marshal to JSON
			layoutBytes, err := json.Marshal(layoutRequest.Layout)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal layout to JSON: %w", err)
			}
			layoutJSON = layoutBytes
		}
	}

	layout := &Layout{
		ID:          layoutRequest.ID,
		DisplayName: layoutRequest.DisplayName,
		Description: layoutRequest.Description,
		Layout:      layoutJSON,
		CreatedAt:   "",
		UpdatedAt:   "",
	}

	return layout, nil
}

// validateLayoutWrapper wraps validateLayoutForDeclarativeResource to match ResourceConfig.Validator signature.
func validateLayoutWrapper(dto interface{}) error {
	layout, ok := dto.(*Layout)
	if !ok {
		return fmt.Errorf("invalid type: expected *Layout")
	}
	return validateLayoutForDeclarativeResource(layout)
}

// validateLayoutForDeclarativeResource validates a layout for declarative resource loading.
func validateLayoutForDeclarativeResource(layout *Layout) error {
	if strings.TrimSpace(layout.DisplayName) == "" {
		return fmt.Errorf("layout display name is required")
	}

	if strings.TrimSpace(layout.ID) == "" {
		return fmt.Errorf("layout ID is required")
	}

	if len(layout.Layout) == 0 {
		return fmt.Errorf("layout configuration is required for '%s'", layout.DisplayName)
	}

	// Validate that layout is valid JSON
	var layoutConfig interface{}
	if err := json.Unmarshal(layout.Layout, &layoutConfig); err != nil {
		return fmt.Errorf("invalid layout JSON for '%s': %w", layout.DisplayName, err)
	}

	return nil
}
