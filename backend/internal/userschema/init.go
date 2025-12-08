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

package userschema

import (
	"encoding/json"
	"fmt"
	"net/http"

	oupkg "github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/immutableresource"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/middleware"

	"gopkg.in/yaml.v3"
)

// Initialize initializes the user schema service and registers its routes.
func Initialize(
	mux *http.ServeMux,
	ouService oupkg.OrganizationUnitServiceInterface,
) UserSchemaServiceInterface {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserSchemaInit"))
	var userSchemaStore userSchemaStoreInterface
	if config.GetThunderRuntime().Config.ImmutableResources.Enabled {
		userSchemaStore = newUserSchemaFileBasedStore()
	} else {
		userSchemaStore = newUserSchemaStore()
	}

	userSchemaService := newUserSchemaService(ouService, userSchemaStore)

	if config.GetThunderRuntime().Config.ImmutableResources.Enabled {
		// Type assert to access Storer interface for resource loading
		fileBasedStore, ok := userSchemaStore.(*userSchemaFileBasedStore)
		if !ok {
			logger.Fatal("Failed to assert userSchemaStore to *userSchemaFileBasedStore")
		}

		resourceConfig := immutableresource.ResourceConfig{
			ResourceType:  "UserSchema",
			DirectoryName: "user_schemas",
			Parser:        parseToUserSchemaDTOWrapper,
			Validator:     validateUserSchemaWrapper(ouService),
			IDExtractor: func(data interface{}) string {
				return data.(*UserSchema).ID
			},
		}

		loader := immutableresource.NewResourceLoader(resourceConfig, fileBasedStore)
		if err := loader.LoadResources(); err != nil {
			logger.Fatal("Failed to load user schema resources", log.Error(err))
		}
	}

	userSchemaHandler := newUserSchemaHandler(userSchemaService)
	registerRoutes(mux, userSchemaHandler)
	return userSchemaService
}

// parseToUserSchemaDTOWrapper wraps parseToUserSchemaDTO to match ResourceConfig.Parser signature
func parseToUserSchemaDTOWrapper(data []byte) (interface{}, error) {
	return parseToUserSchemaDTO(data)
}

// validateUserSchemaWrapper wraps validation logic to match ResourceConfig.Validator signature
func validateUserSchemaWrapper(ouService oupkg.OrganizationUnitServiceInterface) func(interface{}) error {
	return func(data interface{}) error {
		schema := data.(*UserSchema)

		// Validate user schema definition
		if validationErr := validateUserSchemaDefinition(*schema); validationErr != nil {
			return fmt.Errorf("invalid user schema configuration for '%s': %s - %s",
				schema.Name, validationErr.Error, validationErr.ErrorDescription)
		}

		// Validate organization unit reference
		_, svcErr := ouService.GetOrganizationUnit(schema.OrganizationUnitID)
		if svcErr != nil {
			return fmt.Errorf("failed to fetch referred organization unit for user schema '%s' with ouID '%s': %v",
				schema.Name, schema.OrganizationUnitID, svcErr)
		}

		return nil
	}
}

func parseToUserSchemaDTO(data []byte) (*UserSchema, error) {
	var schemaRequest UserSchemaRequestWithID
	err := yaml.Unmarshal(data, &schemaRequest)
	if err != nil {
		return nil, err
	}

	// Validate that schema is valid JSON
	schemaBytes := []byte(schemaRequest.Schema)
	if !json.Valid(schemaBytes) {
		return nil, fmt.Errorf("schema field contains invalid JSON")
	}

	schemaDTO := &UserSchema{
		ID:                    schemaRequest.ID,
		Name:                  schemaRequest.Name,
		OrganizationUnitID:    schemaRequest.OrganizationUnitID,
		AllowSelfRegistration: schemaRequest.AllowSelfRegistration,
		Schema:                []byte(schemaRequest.Schema),
	}

	return schemaDTO, nil
}

// registerRoutes registers the routes for user schema management operations.
func registerRoutes(mux *http.ServeMux, userSchemaHandler *userSchemaHandler) {
	opts1 := middleware.CORSOptions{
		AllowedMethods:   "GET, POST",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("POST /user-schemas",
		userSchemaHandler.HandleUserSchemaPostRequest, opts1))
	mux.HandleFunc(middleware.WithCORS("GET /user-schemas",
		userSchemaHandler.HandleUserSchemaListRequest, opts1))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /user-schemas",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, opts1))

	opts2 := middleware.CORSOptions{
		AllowedMethods:   "GET, PUT, DELETE",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("GET /user-schemas/{id}",
		userSchemaHandler.HandleUserSchemaGetRequest, opts2))
	mux.HandleFunc(middleware.WithCORS("PUT /user-schemas/{id}",
		userSchemaHandler.HandleUserSchemaPutRequest, opts2))
	mux.HandleFunc(middleware.WithCORS("DELETE /user-schemas/{id}",
		userSchemaHandler.HandleUserSchemaDeleteRequest, opts2))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /user-schemas/{id}",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, opts2))
}
