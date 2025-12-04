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
	"fmt"
	"net/http"
	"strings"

	oupkg "github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	filebasedruntime "github.com/asgardeo/thunder/internal/system/file_based_runtime"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/middleware"
	"github.com/asgardeo/thunder/internal/system/utils"
	"github.com/asgardeo/thunder/internal/userschema/model"

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
		configs, err := filebasedruntime.GetConfigs("user-schemas")
		if err != nil {
			logger.Fatal("Failed to read user schema configs from file-based runtime", log.Error(err))
		}
		for _, cfg := range configs {
			schemaDTO, err := parseToUserSchemaDTO(cfg)
			if err != nil {
				logger.Fatal("Error parsing user schema config", log.Error(err))
			}
			svcErr := validateUserSchemaForInit(schemaDTO, ouService)
			if svcErr != nil {
				logger.Fatal("Error validating user schema",
					log.String("schemaName", schemaDTO.Name), log.Any("serviceError", svcErr))
			}

			err = userSchemaStore.CreateUserSchema(*schemaDTO)
			if err != nil {
				logger.Fatal("Failed to store user schema in file-based store",
					log.String("schemaName", schemaDTO.Name), log.Error(err))
			}
		}
	}

	userSchemaHandler := newUserSchemaHandler(userSchemaService)
	registerRoutes(mux, userSchemaHandler)
	return userSchemaService
}

func parseToUserSchemaDTO(data []byte) (*UserSchema, error) {
	var schemaRequest UserSchemaRequestWithID
	err := yaml.Unmarshal(data, &schemaRequest)
	if err != nil {
		return nil, err
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

func validateUserSchemaForInit(schema *UserSchema, ouService oupkg.OrganizationUnitServiceInterface) *serviceerror.ServiceError {
	if schema == nil {
		return &ErrorInvalidUserSchemaRequest
	}
	if strings.TrimSpace(schema.Name) == "" {
		return &ErrorInvalidUserSchemaRequest
	}

	// Validate organization unit ID
	if strings.TrimSpace(schema.OrganizationUnitID) == "" {
		return &ErrorInvalidUserSchemaRequest
	}
	if !utils.IsValidUUID(schema.OrganizationUnitID) {
		return &ErrorInvalidUserSchemaRequest
	}

	// Ensure OU exists
	_, err := ouService.GetOrganizationUnit(schema.OrganizationUnitID)
	if err != nil {
		return &ErrorInvalidUserSchemaRequest
	}

	// Validate schema compilation
	if len(schema.Schema) == 0 {
		return &ErrorInvalidUserSchemaRequest
	}

	_, compileErr := model.CompileUserSchema(schema.Schema)
	if compileErr != nil {
		svcErr := ErrorInvalidUserSchemaRequest
		svcErr.ErrorDescription = fmt.Sprintf("schema compilation failed: %s", compileErr.Error())
		return &svcErr
	}

	return nil
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
