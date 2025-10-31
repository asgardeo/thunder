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
	"net/http"

	"github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/cert"
	"github.com/asgardeo/thunder/internal/system/config"
	filebasedruntime "github.com/asgardeo/thunder/internal/system/file_based_runtime"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/middleware"

	"gopkg.in/yaml.v3"
)

// Initialize initializes the application service and registers its routes.
func Initialize(mux *http.ServeMux, certService cert.CertificateServiceInterface) ApplicationServiceInterface {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationInit"))
	var appStore applicationStoreInterface
	if config.GetThunderRuntime().Config.ImmutableResources.Enabled {
		appStore = newFileBasedStore()
	} else {
		store := newApplicationStore()
		appStore = newCachedBackedApplicationStore(store)
	}

	appService := newApplicationService(appStore, certService)

	if config.GetThunderRuntime().Config.ImmutableResources.Enabled {
		configs, err := filebasedruntime.GetConfigs("applications")
		if err != nil {
			logger.Fatal("Failed to read application configs from file-based runtime", log.Error(err))
		}
		for _, cfg := range configs {
			appDTO, err := parseToApplicationDTO(cfg)
			if err != nil {
				logger.Fatal("Error parsing application config", log.Error(err))
			}
			validatedApp, _, svcErr := appService.ValidateApplication(appDTO)
			if svcErr != nil {
				logger.Fatal("Error validating application", log.String("applicationName", appDTO.Name))
			}

			err = appStore.CreateApplication(*validatedApp)
			if err != nil {
				logger.Fatal("Failed to store application in file-based store",
					log.String("applicationName", appDTO.Name), log.Error(err))
			}
		}
	}

	appHandler := newApplicationHandler(appService)
	registerRoutes(mux, appHandler)
	return appService
}

func parseToApplicationDTO(data []byte) (*model.ApplicationDTO, error) {
	var appRequest model.ApplicationRequest
	err := yaml.Unmarshal(data, &appRequest)
	if err != nil {
		return nil, err
	}

	appDTO := model.ApplicationDTO{
		Name:                      appRequest.Name,
		Description:               appRequest.Description,
		AuthFlowGraphID:           appRequest.AuthFlowGraphID,
		RegistrationFlowGraphID:   appRequest.RegistrationFlowGraphID,
		IsRegistrationFlowEnabled: appRequest.IsRegistrationFlowEnabled,
		URL:                       appRequest.URL,
		LogoURL:                   appRequest.LogoURL,
		Token:                     appRequest.Token,
		Certificate:               appRequest.Certificate,
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
