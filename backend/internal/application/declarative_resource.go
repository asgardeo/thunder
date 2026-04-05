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

package application

import (
	"context"
	"fmt"
	"testing"

	"encoding/json"

	"github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/entity"
	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"

	"gopkg.in/yaml.v3"
)

const (
	resourceTypeApplication = "application"
	paramTypApplication     = "Application"
)

// applicationExporter implements declarativeresource.ResourceExporter for applications.
type applicationExporter struct {
	service ApplicationServiceInterface
}

// newApplicationExporter creates a new application exporter.
func newApplicationExporter(service ApplicationServiceInterface) *applicationExporter {
	return &applicationExporter{service: service}
}

// NewApplicationExporterForTest creates a new application exporter for testing purposes.
func NewApplicationExporterForTest(service ApplicationServiceInterface) *applicationExporter {
	if !testing.Testing() {
		panic("only for tests!")
	}
	return newApplicationExporter(service)
}

// GetResourceType returns the resource type for applications.
func (e *applicationExporter) GetResourceType() string {
	return resourceTypeApplication
}

// GetParameterizerType returns the parameterizer type for applications.
func (e *applicationExporter) GetParameterizerType() string {
	return paramTypApplication
}

// GetAllResourceIDs retrieves all application IDs.
// In composite mode, this excludes declarative (YAML-based) applications.
func (e *applicationExporter) GetAllResourceIDs(ctx context.Context) ([]string, *serviceerror.ServiceError) {
	apps, err := e.service.GetApplicationList(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(apps.Applications))
	for _, app := range apps.Applications {
		// Only include mutable (database-backed) applications
		if !app.IsReadOnly {
			ids = append(ids, app.ID)
		}
	}
	return ids, nil
}

// GetResourceByID retrieves an application by its ID.
func (e *applicationExporter) GetResourceByID(ctx context.Context, id string) (
	interface{}, string, *serviceerror.ServiceError,
) {
	app, err := e.service.GetApplication(ctx, id)
	if err != nil {
		return nil, "", err
	}
	return app, app.Name, nil
}

// ValidateResource validates an application resource.
func (e *applicationExporter) ValidateResource(
	resource interface{}, id string, logger *log.Logger,
) (string, *declarativeresource.ExportError) {
	app, ok := resource.(*model.Application)
	if !ok {
		return "", declarativeresource.CreateTypeError(resourceTypeApplication, id)
	}

	if err := declarativeresource.ValidateResourceName(
		app.Name, resourceTypeApplication, id, "APP_VALIDATION_ERROR", logger); err != nil {
		return "", err
	}

	return app.Name, nil
}

// loadDeclarativeResources loads application resources from declarative files.
// Works in both declarative-only and composite modes:
// - In declarative mode: appStore is a fileBasedStore
// - In composite mode: appStore is a compositeApplicationStore (contains both file and DB stores)
func loadDeclarativeResources(appStore applicationStoreInterface, appService ApplicationServiceInterface) error {
	var fileStore applicationStoreInterface
	var dbStore applicationStoreInterface

	// Determine store type and extract appropriate stores
	switch store := appStore.(type) {
	case *compositeApplicationStore:
		// Composite mode: both file and DB stores available
		fileStore = store.fileStore
		dbStore = store.dbStore
	case *fileBasedStore:
		// Declarative-only mode: only file store available
		fileStore = store
		dbStore = nil
	default:
		return fmt.Errorf("invalid store type for loading declarative resources")
	}

	// Type assert to access Storer interface for resource loading
	fileBasedStoreImpl, ok := fileStore.(*fileBasedStore)
	if !ok {
		return fmt.Errorf("failed to assert fileStore to *fileBasedStore")
	}

	// Use a custom loader for applications due to transformation from DTO to ProcessedDTO
	resourceConfig := declarativeresource.ResourceConfig{
		ResourceType:  "Application",
		DirectoryName: "applications",
		Parser:        parseAndValidateApplicationWrapper(appService),
		Validator: func(data interface{}) error {
			return validateApplicationWrapper(data, fileStore, dbStore)
		},
		IDExtractor: func(data interface{}) string {
			return data.(*model.ApplicationProcessedDTO).ID
		},
	}

	loader := declarativeresource.NewResourceLoader(resourceConfig, fileBasedStoreImpl)
	if err := loader.LoadResources(); err != nil {
		return fmt.Errorf("failed to load application resources: %w", err)
	}

	return nil
}

// parseAndValidateApplicationWrapper combines parsing and validation for applications.
// This is needed because applications undergo transformation from ApplicationDTO to ApplicationProcessedDTO.
func parseAndValidateApplicationWrapper(appService ApplicationServiceInterface) func([]byte) (interface{}, error) {
	return func(data []byte) (interface{}, error) {
		appDTO, err := parseToApplicationDTO(data)
		if err != nil {
			return nil, err
		}

		// Validate and transform the application
		validatedApp, _, svcErr := appService.ValidateApplication(context.Background(), appDTO)
		if svcErr != nil {
			return nil, fmt.Errorf("error validating application '%s': %v", appDTO.Name, svcErr)
		}

		return validatedApp, nil
	}
}

func parseToApplicationDTO(data []byte) (*model.ApplicationDTO, error) {
	var appRequest model.ApplicationRequestWithID
	err := yaml.Unmarshal(data, &appRequest)
	if err != nil {
		return nil, err
	}

	appDTO := model.ApplicationDTO{
		ID:                        appRequest.ID,
		OUID:                      appRequest.OUID,
		Name:                      appRequest.Name,
		Description:               appRequest.Description,
		AuthFlowID:                appRequest.AuthFlowID,
		RegistrationFlowID:        appRequest.RegistrationFlowID,
		IsRegistrationFlowEnabled: appRequest.IsRegistrationFlowEnabled,
		ThemeID:                   appRequest.ThemeID,
		LayoutID:                  appRequest.LayoutID,
		Template:                  appRequest.Template,
		URL:                       appRequest.URL,
		LogoURL:                   appRequest.LogoURL,
		TosURI:                    appRequest.TosURI,
		PolicyURI:                 appRequest.PolicyURI,
		Contacts:                  appRequest.Contacts,
		Assertion:                 appRequest.Assertion,
		Certificate:               appRequest.Certificate,
		AllowedUserTypes:          appRequest.AllowedUserTypes,
		LoginConsent:              &model.LoginConsentConfig{ValidityPeriod: 0},
		Metadata:                  appRequest.Metadata,
	}
	if len(appRequest.InboundAuthConfig) > 0 {
		inboundAuthConfigDTOs := make([]model.InboundAuthConfigDTO, 0)
		for _, config := range appRequest.InboundAuthConfig {
			if config.Type != model.OAuthInboundAuthType || config.OAuthAppConfig == nil {
				continue
			}

			inboundAuthConfigDTO := model.InboundAuthConfigDTO{
				Type: config.Type,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					ClientID:                config.OAuthAppConfig.ClientID,
					ClientSecret:            config.OAuthAppConfig.ClientSecret,
					RedirectURIs:            config.OAuthAppConfig.RedirectURIs,
					GrantTypes:              config.OAuthAppConfig.GrantTypes,
					ResponseTypes:           config.OAuthAppConfig.ResponseTypes,
					TokenEndpointAuthMethod: config.OAuthAppConfig.TokenEndpointAuthMethod,
					PKCERequired:            config.OAuthAppConfig.PKCERequired,
					PublicClient:            config.OAuthAppConfig.PublicClient,
					Token:                   config.OAuthAppConfig.Token,
					Scopes:                  config.OAuthAppConfig.Scopes,
					UserInfo:                config.OAuthAppConfig.UserInfo,
					ScopeClaims:             config.OAuthAppConfig.ScopeClaims,
				},
			}
			inboundAuthConfigDTOs = append(inboundAuthConfigDTOs, inboundAuthConfigDTO)
		}
		appDTO.InboundAuthConfig = inboundAuthConfigDTOs
	}
	return &appDTO, nil
}

func validateApplicationWrapper(
	data interface{},
	fileStore applicationStoreInterface,
	dbStore applicationStoreInterface,
) error {
	app, ok := data.(*model.ApplicationProcessedDTO)
	if !ok {
		return fmt.Errorf("invalid type: expected *ApplicationProcessedDTO")
	}

	if app.Name == "" {
		return fmt.Errorf("application name cannot be empty")
	}

	// Check for duplicate ID in the file store
	exists, err := fileStore.IsApplicationExists(context.Background(), app.ID)
	if err != nil {
		return fmt.Errorf("failed to check application existence: %w", err)
	}
	if exists {
		return fmt.Errorf("duplicate application ID '%s': "+
			"an application with this ID already exists in declarative resources", app.ID)
	}

	// COMPOSITE MODE: Check for duplicate ID in the database store
	if dbStore != nil {
		exists, err := dbStore.IsApplicationExists(context.Background(), app.ID)
		if err != nil {
			return fmt.Errorf("failed to check application existence: %w", err)
		}
		if exists {
			return fmt.Errorf("duplicate application ID '%s': "+
				"an application with this ID already exists in the database store", app.ID)
		}
	}

	// TODO: Add more validation as needed

	return nil
}

// GetResourceRules returns the parameterization rules for applications.
func (e *applicationExporter) GetResourceRules() *declarativeresource.ResourceRules {
	return &declarativeresource.ResourceRules{
		Variables: []string{
			"InboundAuthConfig[].OAuthAppConfig.ClientID",
			"InboundAuthConfig[].OAuthAppConfig.ClientSecret",
		},
		ArrayVariables: []string{
			"InboundAuthConfig[].OAuthAppConfig.RedirectURIs",
		},
	}
}

// makeAppDeclarativeConfig creates the declarative loader config for loading application
// identity data into the entity file store.
func makeAppDeclarativeConfig() entity.DeclarativeLoaderConfig {
	return entity.DeclarativeLoaderConfig{
		Directory: "applications",
		Category:  entity.EntityCategoryApp,
		Parser:    makeAppEntityParser(),
	}
}

// makeAppEntityParser creates a parser that converts application YAML into an entity.
func makeAppEntityParser() func(data []byte) (*entity.Entity, json.RawMessage, json.RawMessage, error) {
	return func(data []byte) (*entity.Entity, json.RawMessage, json.RawMessage, error) {
		var appRequest model.ApplicationRequestWithID
		if err := yaml.Unmarshal(data, &appRequest); err != nil {
			return nil, nil, nil, fmt.Errorf("failed to unmarshal application YAML: %w", err)
		}

		// Build system attributes from app identity fields.
		sysAttrs := map[string]interface{}{
			fieldName: appRequest.Name,
		}
		if appRequest.Description != "" {
			sysAttrs[fieldDescription] = appRequest.Description
		}

		sysAttrsJSON, err := json.Marshal(sysAttrs)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to marshal system attributes: %w", err)
		}

		// Build system credentials if OAuth config with client secret exists.
		var sysCredsJSON json.RawMessage
		if len(appRequest.InboundAuthConfig) > 0 && appRequest.InboundAuthConfig[0].OAuthAppConfig != nil {
			clientSecret := appRequest.InboundAuthConfig[0].OAuthAppConfig.ClientSecret
			if clientSecret != "" {
				sysAttrs[fieldClientID] = appRequest.InboundAuthConfig[0].OAuthAppConfig.ClientID
				// Re-marshal with clientId added.
				sysAttrsJSON, _ = json.Marshal(sysAttrs)

				creds := map[string]interface{}{
					fieldClientSecret: clientSecret,
				}
				sysCredsJSON, _ = json.Marshal(creds)
			}
		}

		e := &entity.Entity{
			ID:                 appRequest.ID,
			Category:           entity.EntityCategoryApp,
			Type:               "application",
			State:              entity.EntityStateActive,
			OrganizationUnitID: appRequest.OUID,
			SystemAttributes:   sysAttrsJSON,
		}

		return e, nil, sysCredsJSON, nil
	}
}
