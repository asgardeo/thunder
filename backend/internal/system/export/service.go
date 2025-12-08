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

	"github.com/asgardeo/thunder/internal/application"
	"github.com/asgardeo/thunder/internal/idp"
	"github.com/asgardeo/thunder/internal/notification"
	"github.com/asgardeo/thunder/internal/notification/common"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/userschema"
)

const (
	formatYAML = "yaml"
	formatJSON = "json"

	resourceTypeApplication        = "application"
	resourceTypeIdentityProvider   = "identity_provider"
	resourceTypeNotificationSender = "notification_sender"
	resourceTypeUserSchema         = "user_schema"
)

// ParameterizerInterface defines the interface for template parameterization.
type ParameterizerInterface interface {
	ToParameterizedYAML(obj interface{}, resourceType string, resourceName string) (string, error)
}

// ExportServiceInterface defines the interface for the export service.
type ExportServiceInterface interface {
	ExportResources(request *ExportRequest) (*ExportResponse, *serviceerror.ServiceError)
}

// exportService implements the ExportServiceInterface.
type exportService struct {
	applicationService        application.ApplicationServiceInterface
	idpService                idp.IDPServiceInterface
	notificationSenderService notification.NotificationSenderMgtSvcInterface
	userSchemaService         userschema.UserSchemaServiceInterface
	parameterizer             ParameterizerInterface
	registry                  *ResourceExporterRegistry
	// Future: Add other service dependencies
	// groupService group.GroupServiceInterface
	// userService  user.UserServiceInterface
}

// newExportService creates a new instance of exportService.
func newExportService(appService application.ApplicationServiceInterface,
	idpService idp.IDPServiceInterface,
	notificationSenderService notification.NotificationSenderMgtSvcInterface,
	userSchemaService userschema.UserSchemaServiceInterface,
	param ParameterizerInterface) ExportServiceInterface {
	// Create registry and register all exporters
	registry := NewResourceExporterRegistry()
	registry.Register(NewApplicationExporter(appService))
	registry.Register(NewIDPExporter(idpService))
	registry.Register(NewNotificationSenderExporter(notificationSenderService))
	registry.Register(NewUserSchemaExporter(userSchemaService))

	return &exportService{
		applicationService:        appService,
		idpService:                idpService,
		notificationSenderService: notificationSenderService,
		userSchemaService:         userSchemaService,
		parameterizer:             param,
		registry:                  registry,
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
	var exportErrors []ExportError
	resourceCounts := make(map[string]int)

	// Export applications if requested
	if len(request.Applications) > 0 {
		appFiles, appErrors := es.exportApplications(request.Applications, options)
		exportFiles = append(exportFiles, appFiles...)
		exportErrors = append(exportErrors, appErrors...)
		resourceCounts["applications"] = len(appFiles)
	}

	// Export identity providers if requested
	if len(request.IdentityProviders) > 0 {
		idpFiles, idpErrors := es.exportIdentityProviders(request.IdentityProviders, options)
		exportFiles = append(exportFiles, idpFiles...)
		exportErrors = append(exportErrors, idpErrors...)
		resourceCounts["identity_providers"] = len(idpFiles)
	}

	// Export notification senders if requested
	if len(request.NotificationSenders) > 0 {
		senderFiles, senderErrors := es.exportNotificationSenders(request.NotificationSenders, options)
		exportFiles = append(exportFiles, senderFiles...)
		exportErrors = append(exportErrors, senderErrors...)
		resourceCounts["notification_senders"] = len(senderFiles)
	}

	// Export user schemas if requested
	if len(request.UserSchemas) > 0 {
		schemaFiles, schemaErrors := es.exportUserSchemas(request.UserSchemas, options)
		exportFiles = append(exportFiles, schemaFiles...)
		exportErrors = append(exportErrors, schemaErrors...)
		resourceCounts["user_schemas"] = len(schemaFiles)
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

// createTypeError creates a standardized type assertion error.
func createTypeError(resourceType, resourceID string) *ExportError {
	return &ExportError{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Error:        "Invalid resource type",
		Code:         "INVALID_TYPE",
	}
}

// validateResourceName validates that a resource name is not empty and returns an error if it is.
func validateResourceName(name, resourceType, resourceID, errorCode string, logger *log.Logger) *ExportError {
	if name == "" {
		logger.Warn(resourceType+" missing name, skipping export",
			log.String("resourceID", resourceID))
		return &ExportError{
			ResourceType: resourceType,
			ResourceID:   resourceID,
			Error:        resourceType + " name is empty",
			Code:         errorCode,
		}
	}
	return nil
}

// validateResourceGeneric provides generic validation for any resource type.
func validateResourceGeneric(
	resource interface{},
	id string,
	resourceType string,
	validationCode string,
	logger *log.Logger,
	castResource func(interface{}) (interface{}, bool),
	extractName func(interface{}) string,
	validateExtra func(interface{}, string, string, *log.Logger),
) (string, *ExportError) {
	// Type assertion
	castedResource, ok := castResource(resource)
	if !ok {
		return "", createTypeError(resourceType, id)
	}

	// Extract and validate name
	name := extractName(castedResource)
	if err := validateResourceName(name, resourceType, id, validationCode, logger); err != nil {
		return "", err
	}

	// Additional validation (resource-specific)
	if validateExtra != nil {
		validateExtra(castedResource, id, name, logger)
	}

	return name, nil
}

// getAllResourceIDsGeneric provides a generic way to get all resource IDs from a service.
// It accepts a callback function that retrieves the list and extracts IDs.
func getAllResourceIDsGeneric(
	getList func() (interface{}, *serviceerror.ServiceError),
	extractIDs func(interface{}) []string,
) ([]string, *serviceerror.ServiceError) {
	result, err := getList()
	if err != nil {
		return nil, err
	}
	return extractIDs(result), nil
}

// exportApplications exports application configurations as YAML files.
func (es *exportService) exportApplications(applicationIDs []string, options *ExportOptions) (
	[]ExportFile, []ExportError) {
	logger := log.GetLogger().With(log.String("component", "ExportService"))
	exportFiles := make([]ExportFile, 0, len(applicationIDs))
	exportErrors := make([]ExportError, 0, len(applicationIDs))

	applicationIDList := make([]string, 0)
	if len(applicationIDs) == 1 && applicationIDs[0] == "*" {
		// Support pagination once applicationList supports it.
		apps, err := es.applicationService.GetApplicationList()
		if err != nil {
			logger.Warn("Failed to get all applications", log.Any("error", err))
			return nil, nil
		}
		for _, app := range apps.Applications {
			applicationIDList = append(applicationIDList, app.ID)
		}
	} else {
		applicationIDList = applicationIDs
	}

	for _, appID := range applicationIDList {
		// Get the application
		app, svcErr := es.applicationService.GetApplication(appID)
		if svcErr != nil {
			logger.Warn("Failed to get application for export",
				log.String("appID", appID), log.String("error", svcErr.Error))
			exportErrors = append(exportErrors, ExportError{
				ResourceType: resourceTypeApplication,
				ResourceID:   appID,
				Error:        svcErr.Error,
				Code:         svcErr.Code,
			})
			continue // Skip applications that can't be found
		}

		// Convert to export format based on options
		var content string
		var fileName string

		if options.Format == formatJSON {
			// Convert to JSON format (could be implemented later)
			logger.Warn("JSON format not yet implemented, falling back to YAML")
			options.Format = formatYAML
		}

		templateContent, err := es.generateTemplateFromStruct(app, "Application", app.Name)
		if err != nil {
			logger.Warn("Failed to generate template from struct",
				log.String("appID", appID), log.String("error", err.Error()))
			exportErrors = append(exportErrors, ExportError{
				ResourceType: resourceTypeApplication,
				ResourceID:   appID,
				Error:        err.Error(),
				Code:         "TemplateGenerationError",
			})
			continue
		}
		content = templateContent

		// Determine file name and folder path based on options
		fileName = es.generateFileName(app.Name, resourceTypeApplication, appID, options)
		folderPath := es.generateFolderPath(resourceTypeApplication, options)

		// Create export file
		exportFile := ExportFile{
			FileName:     fileName,
			Content:      content,
			FolderPath:   folderPath,
			ResourceType: resourceTypeApplication,
			ResourceID:   appID,
		}
		exportFiles = append(exportFiles, exportFile)
	}

	return exportFiles, exportErrors
}

// exportIdentityProviders exports identity provider configurations as YAML files.
// nolint:dupl // Intentional duplication - follows same pattern as other export functions
func (es *exportService) exportIdentityProviders(idpIDs []string, options *ExportOptions) (
	[]ExportFile, []ExportError) {
	logger := log.GetLogger().With(log.String("component", "ExportService"))

	getAllResources := func() ([]string, *serviceerror.ServiceError) {
		return getAllResourceIDsGeneric(
			func() (interface{}, *serviceerror.ServiceError) {
				return es.idpService.GetIdentityProviderList()
			},
			func(result interface{}) []string {
				idps := result.([]idp.BasicIDPDTO)
				ids := make([]string, 0, len(idps))
				for _, idp := range idps {
					ids = append(ids, idp.ID)
				}
				return ids
			},
		)
	}

	getResource := func(id string) (interface{}, string, *serviceerror.ServiceError) {
		idpDTO, svcErr := es.idpService.GetIdentityProvider(id)
		if svcErr != nil {
			return nil, "", svcErr
		}
		return idpDTO, idpDTO.Name, nil
	}

	validateResource := func(resource interface{}, id string) (string, *ExportError) {
		return validateResourceGeneric(
			resource, id, resourceTypeIdentityProvider, "IDP_VALIDATION_ERROR", logger,
			func(r interface{}) (interface{}, bool) {
				idpDTO, ok := r.(*idp.IDPDTO)
				return idpDTO, ok
			},
			func(r interface{}) string {
				return r.(*idp.IDPDTO).Name
			},
			func(r interface{}, id, name string, logger *log.Logger) {
				idpDTO := r.(*idp.IDPDTO)
				if len(idpDTO.Properties) == 0 {
					logger.Warn("Identity provider has no properties",
						log.String("idpID", id), log.String("name", name))
				}
			},
		)
	}

	return es.exportResourceGeneric(
		idpIDs, resourceTypeIdentityProvider, getAllResources, getResource, validateResource, options)
}

// exportNotificationSenders exports notification sender configurations as YAML files.
// nolint:dupl // Intentional duplication - follows same pattern as other export functions
func (es *exportService) exportNotificationSenders(senderIDs []string, options *ExportOptions) (
	[]ExportFile, []ExportError) {
	logger := log.GetLogger().With(log.String("component", "ExportService"))

	getAllResources := func() ([]string, *serviceerror.ServiceError) {
		return getAllResourceIDsGeneric(
			func() (interface{}, *serviceerror.ServiceError) {
				return es.notificationSenderService.ListSenders()
			},
			func(result interface{}) []string {
				senders := result.([]common.NotificationSenderDTO)
				ids := make([]string, 0, len(senders))
				for _, sender := range senders {
					ids = append(ids, sender.ID)
				}
				return ids
			},
		)
	}

	getResource := func(id string) (interface{}, string, *serviceerror.ServiceError) {
		senderDTO, svcErr := es.notificationSenderService.GetSender(id)
		if svcErr != nil {
			return nil, "", svcErr
		}
		return senderDTO, senderDTO.Name, nil
	}

	validateResource := func(resource interface{}, id string) (string, *ExportError) {
		return validateResourceGeneric(
			resource, id, resourceTypeNotificationSender, "SENDER_VALIDATION_ERROR", logger,
			func(r interface{}) (interface{}, bool) {
				senderDTO, ok := r.(*common.NotificationSenderDTO)
				return senderDTO, ok
			},
			func(r interface{}) string {
				return r.(*common.NotificationSenderDTO).Name
			},
			func(r interface{}, id, name string, logger *log.Logger) {
				senderDTO := r.(*common.NotificationSenderDTO)
				if len(senderDTO.Properties) == 0 {
					logger.Warn("Notification sender has no properties",
						log.String("senderID", id), log.String("name", name))
				}
			},
		)
	}

	return es.exportResourceGeneric(
		senderIDs, resourceTypeNotificationSender, getAllResources, getResource, validateResource, options)
}

// exportUserSchemas exports user schema configurations as YAML files.
func (es *exportService) exportUserSchemas(schemaIDs []string, options *ExportOptions) (
	[]ExportFile, []ExportError) {
	logger := log.GetLogger().With(log.String("component", "ExportService"))

	getAllResources := func() ([]string, *serviceerror.ServiceError) {
		return getAllResourceIDsGeneric(
			func() (interface{}, *serviceerror.ServiceError) {
				return es.userSchemaService.GetUserSchemaList(0, 1000)
			},
			func(result interface{}) []string {
				response := result.(*userschema.UserSchemaListResponse)
				ids := make([]string, 0, len(response.Schemas))
				for _, schema := range response.Schemas {
					ids = append(ids, schema.ID)
				}
				return ids
			},
		)
	}

	getResource := func(id string) (interface{}, string, *serviceerror.ServiceError) {
		schemaDTO, svcErr := es.userSchemaService.GetUserSchema(id)
		if svcErr != nil {
			return nil, "", svcErr
		}
		return schemaDTO, schemaDTO.Name, nil
	}

	validateResource := func(resource interface{}, id string) (string, *ExportError) {
		return validateResourceGeneric(
			resource, id, resourceTypeUserSchema, "SCHEMA_VALIDATION_ERROR", logger,
			func(r interface{}) (interface{}, bool) {
				schemaDTO, ok := r.(*userschema.UserSchema)
				return schemaDTO, ok
			},
			func(r interface{}) string {
				return r.(*userschema.UserSchema).Name
			},
			func(r interface{}, id, name string, logger *log.Logger) {
				schemaDTO := r.(*userschema.UserSchema)
				if len(schemaDTO.Schema) == 0 {
					logger.Warn("User schema has no schema definition",
						log.String("schemaID", id), log.String("name", name))
				}
			},
		)
	}

	return es.exportResourceGeneric(
		schemaIDs, resourceTypeUserSchema, getAllResources, getResource, validateResource, options)
}

// exportResourceGeneric is a generic helper function to export resources with common logic.
func (es *exportService) exportResourceGeneric(
	resourceIDs []string,
	resourceType string,
	getAllResources func() ([]string, *serviceerror.ServiceError),
	getResource func(id string) (interface{}, string, *serviceerror.ServiceError),
	validateResource func(resource interface{}, id string) (string, *ExportError),
	options *ExportOptions,
) ([]ExportFile, []ExportError) {
	logger := log.GetLogger().With(log.String("component", "ExportService"))
	exportFiles := make([]ExportFile, 0, len(resourceIDs))
	exportErrors := make([]ExportError, 0, len(resourceIDs))

	var resourceIDList []string
	if len(resourceIDs) == 1 && resourceIDs[0] == "*" {
		// Export all resources
		ids, err := getAllResources()
		if err != nil {
			logger.Warn("Failed to get all resources", log.String("resourceType", resourceType), log.Any("error", err))
			return nil, nil
		}
		resourceIDList = ids
	} else {
		resourceIDList = resourceIDs
	}

	for _, resourceID := range resourceIDList {
		// Get the resource
		resource, _, svcErr := getResource(resourceID)
		if svcErr != nil {
			logger.Warn("Failed to get resource for export",
				log.String("resourceType", resourceType),
				log.String("resourceID", resourceID),
				log.String("error", svcErr.Error))
			exportErrors = append(exportErrors, ExportError{
				ResourceType: resourceType,
				ResourceID:   resourceID,
				Error:        svcErr.Error,
				Code:         svcErr.Code,
			})
			continue
		}

		// Validate resource
		validatedName, exportErr := validateResource(resource, resourceID)
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

		templateContent, err := es.generateTemplateFromStruct(resource, resourceType, validatedName)
		if err != nil {
			logger.Warn("Failed to generate template from struct",
				log.String("resourceType", resourceType),
				log.String("resourceID", resourceID),
				log.String("error", err.Error()))
			exportErrors = append(exportErrors, ExportError{
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

// toParameterizerResourceType maps external resource type names to parameterizer resource type names.
func toParameterizerResourceType(resourceType string) string {
	switch resourceType {
	case resourceTypeApplication:
		return "Application"
	case resourceTypeIdentityProvider:
		return "IdentityProvider"
	case resourceTypeNotificationSender:
		return "NotificationSender"
	case resourceTypeUserSchema:
		return "UserSchema"
	default:
		return resourceType
	}
}

func (es *exportService) generateTemplateFromStruct(
	data interface{}, resourceType string, resourceName string) (string, error) {
	paramResourceType := toParameterizerResourceType(resourceType)
	template, err := es.parameterizer.ToParameterizedYAML(data, paramResourceType, resourceName)
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
