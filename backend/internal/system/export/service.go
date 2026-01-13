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
	"strings"
	"time"

	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	immutableresource "github.com/asgardeo/thunder/internal/system/immutable_resource"
	"github.com/asgardeo/thunder/internal/system/log"
)

const (
	formatYAML = "yaml"
	formatJSON = "json"

	resourceTypeApplication        = "application"
	resourceTypeIdentityProvider   = "identity_provider"
	resourceTypeNotificationSender = "notification_sender"
	resourceTypeUserSchema         = "user_schema"
	resourceTypeOU                 = "organization_unit"
	resourceTypeFlow               = "flow"
	resourceTypeTranslation        = "translation"
)

// parameterizerInterface defines the interface for template parameterization.
type parameterizerInterface interface {
	ToParameterizedYAML(obj interface{},
		resourceType string, resourceName string, rules *immutableresource.ResourceRules) (string, error)
}

// ExportServiceInterface defines the interface for the export service.
type ExportServiceInterface interface {
	ExportResources(request *ExportRequest) (*ExportResponse, *serviceerror.ServiceError)
}

// exportService implements the ExportServiceInterface.
type exportService struct {
	parameterizer parameterizerInterface
	registry      *ResourceExporterRegistry
}

// newExportService creates a new instance of exportService.
func newExportService(
	exporters []immutableresource.ResourceExporter, param parameterizerInterface,
) ExportServiceInterface {
	// Create registry and register all exporters
	registry := newResourceExporterRegistry()
	for _, exporter := range exporters {
		registry.Register(exporter)
	}

	return &exportService{
		parameterizer: param,
		registry:      registry,
	}
}

// ExportResources exports the specified resources as YAML files.
func (es *exportService) ExportResources(request *ExportRequest) (*ExportResponse, *serviceerror.ServiceError) {
	if request == nil {
		return nil, serviceerror.CustomServiceError(
			ErrorInvalidRequest,
			"Export request cannot be nil",
		)
	}

	// Set default options if not provided
	options := request.Options
	if options == nil {
		options = &ExportOptions{
			Format: formatYAML,
		}
	}
	if options.Format == "" {
		options.Format = formatYAML
	}

	var exportFiles []ExportFile
	var exportErrors []immutableresource.ExportError
	resourceCounts := make(map[string]int)

	// Map resource types to their IDs from the request
	resourceMap := map[string][]string{
		resourceTypeApplication:        request.Applications,
		resourceTypeIdentityProvider:   request.IdentityProviders,
		resourceTypeNotificationSender: request.NotificationSenders,
		resourceTypeUserSchema:         request.UserSchemas,
		resourceTypeOU:                 request.OrganizationUnits,
		resourceTypeFlow:               request.Flows,
		resourceTypeTranslation:        request.Translations,
	}

	// Export resources using the registry
	for resourceType, resourceIDs := range resourceMap {
		if len(resourceIDs) == 0 {
			continue
		}

		exporter, exists := es.registry.Get(resourceType)
		if !exists {
			log.GetLogger().Warn("No exporter registered for resource type",
				log.String("resourceType", resourceType))
			continue
		}

		files, errors := es.exportResourcesWithExporter(exporter, resourceIDs, options)
		exportFiles = append(exportFiles, files...)
		exportErrors = append(exportErrors, errors...)
		resourceCounts[resourceType] = len(files)
	}

	if len(exportFiles) == 0 {
		return nil, serviceerror.CustomServiceError(
			ErrorNoResourcesFound,
			"No valid resources found for export",
		)
	}

	// Calculate total size
	var totalSize int64
	for i := range exportFiles {
		exportFiles[i].Size = int64(len(exportFiles[i].Content))
		totalSize += exportFiles[i].Size
	}

	summary := &ExportSummary{
		TotalFiles:    len(exportFiles),
		TotalSize:     totalSize,
		ExportedAt:    time.Now().UTC().Format(time.RFC3339),
		ResourceTypes: resourceCounts,
		Errors:        exportErrors,
	}

	return &ExportResponse{
		Files:   exportFiles,
		Summary: summary,
	}, nil
}

// exportResourcesWithExporter exports resources using a registered exporter.
func (es *exportService) exportResourcesWithExporter(
	exporter immutableresource.ResourceExporter,
	resourceIDs []string,
	options *ExportOptions,
) ([]ExportFile, []immutableresource.ExportError) {
	logger := log.GetLogger().With(log.String("component", "ExportService"))
	resourceType := exporter.GetResourceType()
	exportFiles := make([]ExportFile, 0, len(resourceIDs))
	exportErrors := make([]immutableresource.ExportError, 0, len(resourceIDs))
	var resourceIDList []string
	if len(resourceIDs) == 1 && resourceIDs[0] == "*" {
		// Export all resources
		ids, err := exporter.GetAllResourceIDs()
		if err != nil {
			logger.Warn("Failed to get all resources",
				log.String("resourceType", resourceType), log.Any("error", err))
			return []ExportFile{}, []immutableresource.ExportError{}
		}
		resourceIDList = ids
	} else {
		resourceIDList = resourceIDs
	}

	for _, resourceID := range resourceIDList {
		// Get the resource
		resource, _, svcErr := exporter.GetResourceByID(resourceID)
		if svcErr != nil {
			logger.Warn("Failed to get resource for export",
				log.String("resourceType", resourceType),
				log.String("resourceID", resourceID),
				log.String("error", svcErr.Error))
			exportErrors = append(exportErrors, immutableresource.ExportError{
				ResourceType: resourceType,
				ResourceID:   resourceID,
				Error:        svcErr.Error,
				Code:         svcErr.Code,
			})
			continue
		}

		// Validate resource
		validatedName, exportErr := exporter.ValidateResource(resource, resourceID, logger)
		if exportErr != nil {
			exportErrors = append(exportErrors, *exportErr)
			continue
		}

		// Convert to export format based on options
		var content string
		var fileName string

		if options.Format == formatJSON {
			// Convert to JSON format (could be implemented later)
			logger.Warn("JSON format not yet implemented, falling back to YAML")
			options.Format = formatYAML
		}

		templateContent, err := es.generateTemplateFromStruct(
			resource, exporter.GetParameterizerType(), validatedName, exporter)
		if err != nil {
			logger.Warn("Failed to generate template from struct",
				log.String("resourceType", resourceType),
				log.String("resourceID", resourceID),
				log.String("error", err.Error()))
			exportErrors = append(exportErrors, immutableresource.ExportError{
				ResourceType: resourceType,
				ResourceID:   resourceID,
				Error:        err.Error(),
				Code:         "TemplateGenerationError",
			})
			continue
		}
		content = templateContent

		// Determine file name and folder path based on options
		fileName = es.generateFileName(validatedName, resourceType, resourceID, options)
		folderPath := es.generateFolderPath(resourceType, options)

		// Create export file
		exportFile := ExportFile{
			FileName:     fileName,
			Content:      content,
			FolderPath:   folderPath,
			ResourceType: resourceType,
			ResourceID:   resourceID,
		}
		exportFiles = append(exportFiles, exportFile)
	}

	return exportFiles, exportErrors
}

func (es *exportService) generateTemplateFromStruct(data interface{},
	paramResourceType string, resourceName string, exporter immutableresource.ResourceExporter) (string, error) {
	template, err := es.parameterizer.ToParameterizedYAML(
		data, paramResourceType, resourceName, exporter.GetResourceRules())
	if err != nil {
		return "", err
	}
	return template, nil
}

// sanitizeFileName sanitizes a filename by removing invalid characters.
func sanitizeFileName(name string) string {
	// Replace spaces with underscores and remove special characters
	sanitized := strings.ReplaceAll(name, " ", "_")
	// Remove any characters that are not alphanumeric, hyphens, or underscores
	var result strings.Builder
	for _, char := range sanitized {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '-' || char == '_' {
			result.WriteRune(char)
		}
	}
	sanitizedName := result.String()
	if sanitizedName == "" {
		sanitizedName = "resource"
	}
	return sanitizedName
}

// generateFileName generates a file name based on naming pattern and options.
// nolint:unparam
func (es *exportService) generateFileName(
	resourceName, resourceType, resourceID string, options *ExportOptions) string {
	// Get file extension based on format
	ext := ".yaml"
	if options.Format == "json" {
		ext = ".json"
	}

	// Use custom naming pattern if provided
	if options.FolderStructure != nil && options.FolderStructure.FileNamingPattern != "" {
		pattern := options.FolderStructure.FileNamingPattern
		pattern = strings.ReplaceAll(pattern, "${name}", sanitizeFileName(resourceName))
		pattern = strings.ReplaceAll(pattern, "${type}", resourceType)
		pattern = strings.ReplaceAll(pattern, "${id}", resourceID)
		return pattern + ext
	}

	// Default naming: sanitized resource name
	return sanitizeFileName(resourceName) + ext
}

// generateFolderPath generates the folder path for a resource based on options.
// nolint:unparam
func (es *exportService) generateFolderPath(resourceType string, options *ExportOptions) string {
	if options.FolderStructure == nil {
		return "" // No folder structure
	}

	// Check for custom structure first
	if options.FolderStructure.CustomStructure != nil {
		if customPath, exists := options.FolderStructure.CustomStructure[resourceType]; exists {
			return customPath
		}
	}

	// Group by type if enabled
	if options.FolderStructure.GroupByType {
		return resourceType + "s" // applications, groups, users, etc.
	}

	return ""
}
