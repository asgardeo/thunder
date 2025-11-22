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
	"slices"
	"strings"

	"github.com/asgardeo/thunder/internal/system/cmodels"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	filebasedruntime "github.com/asgardeo/thunder/internal/system/file_based_runtime"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/middleware"

	"gopkg.in/yaml.v3"
)

// Initialize initializes the IDP service and registers its routes.
func Initialize(mux *http.ServeMux) IDPServiceInterface {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "IDPInit"))
	var idpStore idpStoreInterface
	if config.GetThunderRuntime().Config.ImmutableResources.Enabled {
		idpStore = newIDPFileBasedStore()
	} else {
		idpStore = newIDPStore()
	}

	idpService := newIDPService(idpStore)

	if config.GetThunderRuntime().Config.ImmutableResources.Enabled {
		configs, err := filebasedruntime.GetConfigs("identity-providers")
		if err != nil {
			logger.Fatal("Failed to read identity provider configs from file-based runtime", log.Error(err))
		}
		for _, cfg := range configs {
			idpDTO, err := parseToIDPDTO(cfg)
			if err != nil {
				logger.Fatal("Error parsing identity provider config", log.Error(err))
			}
			svcErr := validateIDPForInit(idpDTO)
			if svcErr != nil {
				logger.Fatal("Error validating identity provider",
					log.String("idpName", idpDTO.Name), log.Any("serviceError", svcErr))
			}

			err = idpStore.CreateIdentityProvider(*idpDTO)
			if err != nil {
				logger.Fatal("Failed to store identity provider in file-based store",
					log.String("idpName", idpDTO.Name), log.Error(err))
			}
		}
	}

	idpHandler := newIDPHandler(idpService)
	registerRoutes(mux, idpHandler)
	return idpService
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

func validateIDPForInit(idp *IDPDTO) *serviceerror.ServiceError {
	if idp == nil {
		return &ErrorIDPNil
	}
	if strings.TrimSpace(idp.Name) == "" {
		return &ErrorInvalidIDPName
	}

	// Validate identity provider type
	if strings.TrimSpace(string(idp.Type)) == "" {
		return &ErrorInvalidIDPType
	}
	isValidType := slices.Contains(supportedIDPTypes, idp.Type)
	if !isValidType {
		return &ErrorInvalidIDPType
	}

	return validateIDPProperties(idp.Properties)
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
