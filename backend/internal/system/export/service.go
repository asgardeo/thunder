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
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/userschema"
)

const (
	formatYAML = "yaml"
	formatJSON = "json"
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
	return &exportService{
		applicationService:        appService,
		idpService:                idpService,
		notificationSenderService: notificationSenderService,
		userSchemaService:         userSchemaService,
		parameterizer:             param,
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
				ResourceType: "application",
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
				ResourceType: "application",
				ResourceID:   appID,
				Error:        err.Error(),
				Code:         "TemplateGenerationError",
			})
			continue
		}
		content = templateContent

		// Determine file name and folder path based on options
		fileName = es.generateFileName(app.Name, "application", appID, options)
		folderPath := es.generateFolderPath("application", options)

		// Create export file
		exportFile := ExportFile{
			FileName:     fileName,
			Content:      content,
			FolderPath:   folderPath,
			ResourceType: "application",
			ResourceID:   appID,
		}
		exportFiles = append(exportFiles, exportFile)
	}

	return exportFiles, exportErrors
}

// exportIdentityProviders exports identity provider configurations as YAML files.
func (es *exportService) exportIdentityProviders(idpIDs []string, options *ExportOptions) (
	[]ExportFile, []ExportError) {
	logger := log.GetLogger().With(log.String("component", "ExportService"))
	exportFiles := make([]ExportFile, 0, len(idpIDs))
	exportErrors := make([]ExportError, 0, len(idpIDs))

	idpIDList := make([]string, 0)
	if len(idpIDs) == 1 && idpIDs[0] == "*" {
		// Export all identity providers
		idps, err := es.idpService.GetIdentityProviderList()
		if err != nil {
			logger.Warn("Failed to get all identity providers", log.Any("error", err))
			return nil, nil
		}
		for _, idp := range idps {
			idpIDList = append(idpIDList, idp.ID)
		}
	} else {
		idpIDList = idpIDs
	}

	for _, idpID := range idpIDList {
		// Get the identity provider
		idp, svcErr := es.idpService.GetIdentityProvider(idpID)
		if svcErr != nil {
			logger.Warn("Failed to get identity provider for export",
				log.String("idpID", idpID), log.String("error", svcErr.Error))
			exportErrors = append(exportErrors, ExportError{
				ResourceType: "identity_provider",
				ResourceID:   idpID,
				Error:        svcErr.Error,
				Code:         svcErr.Code,
			})
			continue // Skip IDPs that can't be found
		}

		// Validate IDP has required fields
		if idp.Name == "" {
			logger.Warn("Identity provider missing name, skipping export",
				log.String("idpID", idpID))
			exportErrors = append(exportErrors, ExportError{
				ResourceType: "identity_provider",
				ResourceID:   idpID,
				Error:        "Identity provider name is empty",
				Code:         "IDP_VALIDATION_ERROR",
			})
			continue
		}

		// Check if IDP has properties - warn if empty but continue
		if len(idp.Properties) == 0 {
			logger.Warn("Identity provider has no properties",
				log.String("idpID", idpID), log.String("name", idp.Name))
		}

		// Convert to export format based on options
		var content string
		var fileName string

		if options.Format == formatJSON {
			// Convert to JSON format (could be implemented later)
			logger.Warn("JSON format not yet implemented, falling back to YAML")
			options.Format = formatYAML
		}

		templateContent, err := es.generateTemplateFromStruct(idp, "IdentityProvider", idp.Name)
		if err != nil {
			logger.Warn("Failed to generate template from struct",
				log.String("idpID", idpID), log.String("error", err.Error()))
			exportErrors = append(exportErrors, ExportError{
				ResourceType: "identity_provider",
				ResourceID:   idpID,
				Error:        err.Error(),
				Code:         "TemplateGenerationError",
			})
			continue
		}
		content = templateContent

		// Determine file name and folder path based on options
		fileName = es.generateFileName(idp.Name, "identity_provider", idpID, options)
		folderPath := es.generateFolderPath("identity_provider", options)

		// Create export file
		exportFile := ExportFile{
			FileName:     fileName,
			Content:      content,
			FolderPath:   folderPath,
			ResourceType: "identity_provider",
			ResourceID:   idpID,
		}
		exportFiles = append(exportFiles, exportFile)
	}

	return exportFiles, exportErrors
}

// exportNotificationSenders exports notification sender configurations as YAML files.
func (es *exportService) exportNotificationSenders(senderIDs []string, options *ExportOptions) (
	[]ExportFile, []ExportError) {
	logger := log.GetLogger().With(log.String("component", "ExportService"))
	exportFiles := make([]ExportFile, 0, len(senderIDs))
	exportErrors := make([]ExportError, 0, len(senderIDs))

	senderIDList := make([]string, 0)
	if len(senderIDs) == 1 && senderIDs[0] == "*" {
		// Export all notification senders
		senders, err := es.notificationSenderService.ListSenders()
		if err != nil {
			logger.Warn("Failed to get all notification senders", log.Any("error", err))
			return nil, nil
		}
		for _, sender := range senders {
			senderIDList = append(senderIDList, sender.ID)
		}
	} else {
		senderIDList = senderIDs
	}

	for _, senderID := range senderIDList {
		// Get the notification sender
		sender, svcErr := es.notificationSenderService.GetSender(senderID)
		if svcErr != nil {
			logger.Warn("Failed to get notification sender for export",
				log.String("senderID", senderID), log.String("error", svcErr.Error))
			exportErrors = append(exportErrors, ExportError{
				ResourceType: "notification_sender",
				ResourceID:   senderID,
				Error:        svcErr.Error,
				Code:         svcErr.Code,
			})
			continue // Skip senders that can't be found
		}

		// Validate sender has required fields
		if sender.Name == "" {
			logger.Warn("Notification sender missing name, skipping export",
				log.String("senderID", senderID))
			exportErrors = append(exportErrors, ExportError{
				ResourceType: "notification_sender",
				ResourceID:   senderID,
				Error:        "Notification sender name is empty",
				Code:         "SENDER_VALIDATION_ERROR",
			})
			continue
		}

		// Check if sender has properties - warn if empty but continue
		if len(sender.Properties) == 0 {
			logger.Warn("Notification sender has no properties",
				log.String("senderID", senderID), log.String("name", sender.Name))
		}

		// Convert to export format based on options
		var content string
		var fileName string

		if options.Format == formatJSON {
			// Convert to JSON format (could be implemented later)
			logger.Warn("JSON format not yet implemented, falling back to YAML")
			options.Format = formatYAML
		}

		templateContent, err := es.generateTemplateFromStruct(sender, "NotificationSender", sender.Name)
		if err != nil {
			logger.Warn("Failed to generate template from struct",
				log.String("senderID", senderID), log.String("error", err.Error()))
			exportErrors = append(exportErrors, ExportError{
				ResourceType: "notification_sender",
				ResourceID:   senderID,
				Error:        err.Error(),
				Code:         "TemplateGenerationError",
			})
			continue
		}
		content = templateContent

		// Determine file name and folder path based on options
		fileName = es.generateFileName(sender.Name, "notification_sender", senderID, options)
		folderPath := es.generateFolderPath("notification_sender", options)

		// Create export file
		exportFile := ExportFile{
			FileName:     fileName,
			Content:      content,
			FolderPath:   folderPath,
			ResourceType: "notification_sender",
			ResourceID:   senderID,
		}
		exportFiles = append(exportFiles, exportFile)
	}

	return exportFiles, exportErrors
}

// exportUserSchemas exports user schema configurations as YAML files.
func (es *exportService) exportUserSchemas(schemaIDs []string, options *ExportOptions) (
	[]ExportFile, []ExportError) {
	logger := log.GetLogger().With(log.String("component", "ExportService"))
	exportFiles := make([]ExportFile, 0, len(schemaIDs))
	exportErrors := make([]ExportError, 0, len(schemaIDs))

	schemaIDList := make([]string, 0)
	if len(schemaIDs) == 1 && schemaIDs[0] == "*" {
		// Export all user schemas
		schemas, err := es.userSchemaService.GetUserSchemaList(0, 1000)
		if err != nil {
			logger.Warn("Failed to get all user schemas", log.Any("error", err))
			return nil, nil
		}
		for _, schema := range schemas.Schemas {
			schemaIDList = append(schemaIDList, schema.ID)
		}
	} else {
		schemaIDList = schemaIDs
	}

	for _, schemaID := range schemaIDList {
		// Get the user schema
		schema, svcErr := es.userSchemaService.GetUserSchema(schemaID)
		if svcErr != nil {
			logger.Warn("Failed to get user schema for export",
				log.String("schemaID", schemaID), log.String("error", svcErr.Error))
			exportErrors = append(exportErrors, ExportError{
				ResourceType: "user_schema",
				ResourceID:   schemaID,
				Error:        svcErr.Error,
				Code:         svcErr.Code,
			})
			continue // Skip schemas that can't be found
		}

		// Validate schema has required fields
		if schema.Name == "" {
			logger.Warn("User schema missing name, skipping export",
				log.String("schemaID", schemaID))
			exportErrors = append(exportErrors, ExportError{
				ResourceType: "user_schema",
				ResourceID:   schemaID,
				Error:        "User schema name is empty",
				Code:         "SCHEMA_VALIDATION_ERROR",
			})
			continue
		}

		// Check if schema has a JSON schema - warn if empty but continue
		if len(schema.Schema) == 0 {
			logger.Warn("User schema has no schema definition",
				log.String("schemaID", schemaID), log.String("name", schema.Name))
		}

		// Convert to export format based on options
		var content string
		var fileName string

		if options.Format == formatJSON {
			// Convert to JSON format (could be implemented later)
			logger.Warn("JSON format not yet implemented, falling back to YAML")
			options.Format = formatYAML
		}

		templateContent, err := es.generateTemplateFromStruct(schema, "UserSchema", schema.Name)
		if err != nil {
			logger.Warn("Failed to generate template from struct",
				log.String("schemaID", schemaID), log.String("error", err.Error()))
			exportErrors = append(exportErrors, ExportError{
				ResourceType: "user_schema",
				ResourceID:   schemaID,
				Error:        err.Error(),
				Code:         "TemplateGenerationError",
			})
			continue
		}
		content = templateContent

		// Determine file name and folder path based on options
		fileName = es.generateFileName(schema.Name, "user_schema", schemaID, options)
		folderPath := es.generateFolderPath("user_schema", options)

		// Create export file
		exportFile := ExportFile{
			FileName:     fileName,
			Content:      content,
			FolderPath:   folderPath,
			ResourceType: "user_schema",
			ResourceID:   schemaID,
		}
		exportFiles = append(exportFiles, exportFile)
	}

	return exportFiles, exportErrors
}

func (es *exportService) generateTemplateFromStruct(
	data interface{}, resourceType string, resourceName string) (string, error) {
	template, err := es.parameterizer.ToParameterizedYAML(data, resourceType, resourceName)
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
