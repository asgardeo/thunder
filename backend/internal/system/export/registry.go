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

package export

import (
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
)

// ResourceExporter defines the interface that each resource type must implement
// to be exportable. This makes it easy to add new resources to the export functionality.
type ResourceExporter interface {
	// GetResourceType returns the type identifier for this resource (e.g., "application", "identity_provider")
	GetResourceType() string

	// GetParameterizerType returns the type name used by the parameterizer (e.g., "Application", "IdentityProvider")
	GetParameterizerType() string

	// GetAllResourceIDs retrieves all resource IDs for wildcard export
	GetAllResourceIDs() ([]string, *serviceerror.ServiceError)

	// GetResourceByID retrieves a single resource by its ID
	// Returns: resource object, resource name, error
	GetResourceByID(id string) (interface{}, string, *serviceerror.ServiceError)

	// ValidateResource validates the resource and extracts its name
	// Returns: resource name, export error
	ValidateResource(resource interface{}, id string, logger *log.Logger) (string, *ExportError)
}

// ResourceExporterRegistry holds all registered resource exporters.
type ResourceExporterRegistry struct {
	exporters map[string]ResourceExporter
}

// NewResourceExporterRegistry creates a new registry for resource exporters.
func NewResourceExporterRegistry() *ResourceExporterRegistry {
	return &ResourceExporterRegistry{
		exporters: make(map[string]ResourceExporter),
	}
}

// Register adds a new resource exporter to the registry.
func (r *ResourceExporterRegistry) Register(exporter ResourceExporter) {
	r.exporters[exporter.GetResourceType()] = exporter
}

// Get retrieves a resource exporter by type.
func (r *ResourceExporterRegistry) Get(resourceType string) (ResourceExporter, bool) {
	exporter, exists := r.exporters[resourceType]
	return exporter, exists
}

// GetAll returns all registered exporters.
func (r *ResourceExporterRegistry) GetAll() map[string]ResourceExporter {
	return r.exporters
}
