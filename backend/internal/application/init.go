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

// Package application provides functionality for managing applications.
package application

import (
	"fmt"
	"net/http"

	"github.com/asgardeo/thunder/internal/application/model"
	brandingmgt "github.com/asgardeo/thunder/internal/branding/mgt"
	"github.com/asgardeo/thunder/internal/cert"
	"github.com/asgardeo/thunder/internal/flow/flowmgt"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/immutableresource"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/middleware"
	"github.com/asgardeo/thunder/internal/userschema"

	"gopkg.in/yaml.v3"
)

// Initialize initializes the application service and registers its routes.
func Initialize(
	mux *http.ServeMux,
	certService cert.CertificateServiceInterface,
	flowMgtService flowmgt.FlowMgtServiceInterface,
	brandingService brandingmgt.BrandingMgtServiceInterface,
	userSchemaService userschema.UserSchemaServiceInterface,
) ApplicationServiceInterface {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationInit"))
	var appStore applicationStoreInterface
	if config.GetThunderRuntime().Config.ImmutableResources.Enabled {
		appStore = newFileBasedStore()
	} else {
		store := newApplicationStore()
		appStore = newCachedBackedApplicationStore(store)
	}

	appService := newApplicationService(appStore, certService, flowMgtService, brandingService, userSchemaService)

	if config.GetThunderRuntime().Config.ImmutableResources.Enabled {
		// Type assert to access Storer interface for resource loading
		fileBasedStore, ok := appStore.(*fileBasedStore)
		if !ok {
			logger.Fatal("Failed to assert appStore to *fileBasedStore")
		}

		// Use a custom loader for applications due to transformation from DTO to ProcessedDTO
		resourceConfig := immutableresource.ResourceConfig{
			ResourceType:  "Application",
			DirectoryName: "applications",
			Parser:        parseAndValidateApplicationWrapper(appService),
			Validator:     nil, // Validation is done in the parser for applications
			IDExtractor: func(data interface{}) string {
				return data.(*model.ApplicationProcessedDTO).ID
			},
		}

		loader := immutableresource.NewResourceLoader(resourceConfig, fileBasedStore)
		if err := loader.LoadResources(); err != nil {
			logger.Fatal("Failed to load application resources", log.Error(err))
		}
	}

	appHandler := newApplicationHandler(appService)
	registerRoutes(mux, appHandler)
	return appService
}

// parseAndValidateApplicationWrapper combines parsing and validation for applications
// This is needed because applications undergo transformation from ApplicationDTO to ApplicationProcessedDTO
func parseAndValidateApplicationWrapper(appService ApplicationServiceInterface) func([]byte) (interface{}, error) {
	return func(data []byte) (interface{}, error) {
		appDTO, err := parseToApplicationDTO(data)
		if err != nil {
			return nil, err
		}

		// Validate and transform the application
		validatedApp, _, svcErr := appService.ValidateApplication(appDTO)
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
		Name:                      appRequest.Name,
		Description:               appRequest.Description,
		AuthFlowGraphID:           appRequest.AuthFlowGraphID,
		RegistrationFlowGraphID:   appRequest.RegistrationFlowGraphID,
		IsRegistrationFlowEnabled: appRequest.IsRegistrationFlowEnabled,
		URL:                       appRequest.URL,
		LogoURL:                   appRequest.LogoURL,
		Token:                     appRequest.Token,
		Certificate:               appRequest.Certificate,
		AllowedUserTypes:          appRequest.AllowedUserTypes,
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
				},
			}
			inboundAuthConfigDTOs = append(inboundAuthConfigDTOs, inboundAuthConfigDTO)
		}
		appDTO.InboundAuthConfig = inboundAuthConfigDTOs
	}
	return &appDTO, nil
}

func registerRoutes(mux *http.ServeMux, appHandler *applicationHandler) {
	opts1 := middleware.CORSOptions{
		AllowedMethods:   "GET, POST",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("POST /applications",
		appHandler.HandleApplicationPostRequest, opts1))
	mux.HandleFunc(middleware.WithCORS("GET /applications",
		appHandler.HandleApplicationListRequest, opts1))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /applications",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, opts1))

	opts2 := middleware.CORSOptions{
		AllowedMethods:   "GET, PUT, DELETE",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("GET /applications/{id}",
		appHandler.HandleApplicationGetRequest, opts2))
	mux.HandleFunc(middleware.WithCORS("PUT /applications/{id}",
		appHandler.HandleApplicationPutRequest, opts2))
	mux.HandleFunc(middleware.WithCORS("DELETE /applications/{id}",
		appHandler.HandleApplicationDeleteRequest, opts2))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /applications/",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, opts2))
}
