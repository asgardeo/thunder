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

// Package idp handles the identity provider management operations.
package idp

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/asgardeo/thunder/internal/system/cmodels"
	"github.com/asgardeo/thunder/internal/system/immutableresource"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/middleware"

	"gopkg.in/yaml.v3"
)

// Initialize initializes the IDP service and registers its routes.
func Initialize(mux *http.ServeMux) IDPServiceInterface {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "IDPInit"))

	// Create store based on configuration
	var idpStore idpStoreInterface
	if immutableresource.IsImmutableModeEnabled() {
		idpStore = newIDPFileBasedStore()
	} else {
		idpStore = newIDPStore()
	}

	idpService := newIDPService(idpStore)

	// Load immutable resources if enabled
	if immutableresource.IsImmutableModeEnabled() {
		// Create a storer wrapper since idpStore interface doesn't expose Create directly
		var storer immutableresource.Storer
		if fileBasedStore, ok := idpStore.(*idpFileBasedStore); ok {
			storer = fileBasedStore
		} else {
			logger.Fatal("Invalid store type for immutable resources")
		}

		resourceConfig := immutableresource.ResourceConfig{
			ResourceType:  "IdentityProvider",
			DirectoryName: "identity_providers",
			Parser:        parseToIDPDTOWrapper,
			Validator:     validateIDPWrapper,
			IDExtractor: func(dto interface{}) string {
				return dto.(*IDPDTO).ID
			},
		}

		loader := immutableresource.NewResourceLoader(resourceConfig, storer)
		if err := loader.LoadResources(); err != nil {
			logger.Fatal("Failed to load identity providers", log.Error(err))
		}
	}

	idpHandler := newIDPHandler(idpService)
	registerRoutes(mux, idpHandler)
	return idpService
}

// parseToIDPDTOWrapper wraps parseToIDPDTO to match the expected signature
func parseToIDPDTOWrapper(data []byte) (interface{}, error) {
	return parseToIDPDTO(data)
}

// validateIDPWrapper wraps validateIDP to match the expected signature
func validateIDPWrapper(dto interface{}) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "IDPInit"))
	idpDTO := dto.(*IDPDTO)
	svcErr := validateIDP(idpDTO, logger)
	if svcErr != nil {
		return fmt.Errorf("%s: %s", svcErr.Error, svcErr.ErrorDescription)
	}
	return nil
}

func parseToIDPDTO(data []byte) (*IDPDTO, error) {
	var idpRequest idpRequestWithID
	err := yaml.Unmarshal(data, &idpRequest)
	if err != nil {
		return nil, err
	}

	idpDTO := &IDPDTO{
		ID:          idpRequest.ID,
		Name:        idpRequest.Name,
		Description: idpRequest.Description,
	}

	// Parse IDP type
	idpType, err := parseIDPType(idpRequest.Type)
	if err != nil {
		return nil, err
	}
	idpDTO.Type = idpType

	// Convert PropertyDTO to Property
	if len(idpRequest.Properties) > 0 {
		properties := make([]cmodels.Property, 0, len(idpRequest.Properties))
		for _, propDTO := range idpRequest.Properties {
			prop, err := cmodels.NewProperty(propDTO.Name, propDTO.Value, propDTO.IsSecret)
			if err != nil {
				return nil, err
			}
			properties = append(properties, *prop)
		}
		idpDTO.Properties = properties
	}

	return idpDTO, nil
}

func parseIDPType(typeStr string) (IDPType, error) {
	// Convert string to uppercase for case-insensitive matching
	typeStrUpper := IDPType(strings.ToUpper(typeStr))

	// Check if it's a valid type
	for _, supportedType := range supportedIDPTypes {
		if supportedType == typeStrUpper {
			return supportedType, nil
		}
	}

	return "", fmt.Errorf("unsupported IDP type: %s", typeStr)
}

// RegisterRoutes registers the routes for identity provider operations.
func registerRoutes(mux *http.ServeMux, idpHandler *idpHandler) {
	opts1 := middleware.CORSOptions{
		AllowedMethods:   "GET, POST",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("POST /identity-providers", idpHandler.HandleIDPPostRequest, opts1))
	mux.HandleFunc(middleware.WithCORS("GET /identity-providers", idpHandler.HandleIDPListRequest, opts1))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /identity-providers",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, opts1))

	opts2 := middleware.CORSOptions{
		AllowedMethods:   "GET, PUT, DELETE",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("GET /identity-providers/{id}",
		idpHandler.HandleIDPGetRequest, opts2))
	mux.HandleFunc(middleware.WithCORS("PUT /identity-providers/{id}",
		idpHandler.HandleIDPPutRequest, opts2))
	mux.HandleFunc(middleware.WithCORS("DELETE /identity-providers/{id}",
		idpHandler.HandleIDPDeleteRequest, opts2))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /identity-providers/{id}",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, opts2))
}
