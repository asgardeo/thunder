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

package ou

import (
	"net/http"
	"strings"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	filebasedruntime "github.com/asgardeo/thunder/internal/system/file_based_runtime"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/middleware"

	"gopkg.in/yaml.v3"
)

// Initialize initializes the organization unit service and registers its routes.
func Initialize(mux *http.ServeMux) OrganizationUnitServiceInterface {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "OUInit"))
	var ouStore organizationUnitStoreInterface

	if config.GetThunderRuntime().Config.ImmutableResources.Enabled {
		// Create both file-based and database stores
		fileStore := newOUFileBasedStore()
		dbStore := newOrganizationUnitStore()

		// Load file-based OUs
		configs, err := filebasedruntime.GetConfigs("organization-units")
		if err != nil {
			logger.Fatal("Failed to read organization unit configs from file-based runtime", log.Error(err))
		}

		// Parse all configs first
		var ous []*OrganizationUnit
		for _, cfg := range configs {
			ouDTO, err := parseToOUDTO(cfg)
			if err != nil {
				logger.Fatal("Error parsing organization unit config", log.Error(err))
			}
			ous = append(ous, ouDTO)
		}

		// Sort OUs: root OUs first (parent == nil), then by parent dependency
		sortedOUs := topologicalSortOUs(ous)

		// Validate and store in order
		for _, ou := range sortedOUs {
			svcErr := validateOUForInit(ou, fileStore)
			if svcErr != nil {
				logger.Fatal("Error validating organization unit",
					log.String("ouName", ou.Name), log.Any("serviceError", svcErr))
			}

			err = fileStore.CreateOrganizationUnit(*ou)
			if err != nil {
				logger.Fatal("Failed to store organization unit in file-based store",
					log.String("ouName", ou.Name), log.Error(err))
			}
		}

		// Create hybrid store that combines both
		ouStore = newOUHybridStore(fileStore, dbStore)
	} else {
		ouStore = newOrganizationUnitStore()
	}

	ouService := newOrganizationUnitService(ouStore)
	ouHandler := newOrganizationUnitHandler(ouService)
	registerRoutes(mux, ouHandler)
	return ouService
}

// parseToOUDTO parses YAML data to OrganizationUnit DTO.
func parseToOUDTO(data []byte) (*OrganizationUnit, error) {
	var ouRequest OrganizationUnitWithID
	err := yaml.Unmarshal(data, &ouRequest)
	if err != nil {
		return nil, err
	}

	return &OrganizationUnit{
		ID:          ouRequest.ID,
		Handle:      ouRequest.Handle,
		Name:        ouRequest.Name,
		Description: ouRequest.Description,
		Parent:      ouRequest.Parent,
	}, nil
}

// validateOUForInit validates an organization unit during initialization.
func validateOUForInit(ou *OrganizationUnit, store organizationUnitStoreInterface) *serviceerror.ServiceError {
	if ou == nil {
		return &ErrorInvalidRequestFormat
	}
	if strings.TrimSpace(ou.Name) == "" {
		return &ErrorInvalidRequestFormat
	}
	if strings.TrimSpace(ou.Handle) == "" {
		return &ErrorInvalidRequestFormat
	}
	if strings.Contains(ou.Handle, "/") {
		return &ErrorInvalidRequestFormat
	}

	// Validate parent exists if specified
	if ou.Parent != nil {
		exists, err := store.IsOrganizationUnitExists(*ou.Parent)
		if err != nil {
			return &ErrorInternalServerError
		}
		if !exists {
			return &ErrorParentOrganizationUnitNotFound
		}
	}

	return nil
}

// topologicalSortOUs sorts organization units so parents come before children.
func topologicalSortOUs(ous []*OrganizationUnit) []*OrganizationUnit {
	// Create a map for quick lookup
	ouMap := make(map[string]*OrganizationUnit)
	for _, ou := range ous {
		ouMap[ou.ID] = ou
	}

	// Build dependency graph
	visited := make(map[string]bool)
	var sorted []*OrganizationUnit

	var visit func(ou *OrganizationUnit)
	visit = func(ou *OrganizationUnit) {
		if visited[ou.ID] {
			return
		}

		// Visit parent first if it exists
		if ou.Parent != nil {
			if parent, exists := ouMap[*ou.Parent]; exists {
				visit(parent)
			}
		}

		visited[ou.ID] = true
		sorted = append(sorted, ou)
	}

	// Visit all OUs
	for _, ou := range ous {
		visit(ou)
	}

	return sorted
}

// registerRoutes registers the routes for organization unit management operations.
func registerRoutes(mux *http.ServeMux, ouHandler *organizationUnitHandler) {
	corsOptions1 := middleware.CORSOptions{
		AllowedMethods:   "GET, POST",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("POST /organization-units",
		ouHandler.HandleOUPostRequest, corsOptions1))
	mux.HandleFunc(middleware.WithCORS("GET /organization-units",
		ouHandler.HandleOUListRequest, corsOptions1))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /organization-units",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, corsOptions1))

	corsOptions2 := middleware.CORSOptions{
		AllowedMethods:   "GET, PUT, DELETE",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("GET /organization-units/",
		func(w http.ResponseWriter, r *http.Request) {
			path := strings.TrimPrefix(r.URL.Path, "/organization-units/")
			segments := strings.Split(path, "/")
			r.SetPathValue("id", segments[0])

			if len(segments) == 1 {
				ouHandler.HandleOUGetRequest(w, r)
			} else if len(segments) == 2 {
				switch segments[1] {
				case "ous":
					ouHandler.HandleOUChildrenListRequest(w, r)
				case "users":
					ouHandler.HandleOUUsersListRequest(w, r)
				case "groups":
					ouHandler.HandleOUGroupsListRequest(w, r)
				default:
					http.NotFound(w, r)
				}
			} else {
				http.NotFound(w, r)
			}
		}, corsOptions2))
	mux.HandleFunc(middleware.WithCORS("PUT /organization-units/{id}",
		ouHandler.HandleOUPutRequest, corsOptions2))
	mux.HandleFunc(middleware.WithCORS("DELETE /organization-units/{id}",
		ouHandler.HandleOUDeleteRequest, corsOptions2))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /organization-units/{id}",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, corsOptions2))

	mux.HandleFunc(middleware.WithCORS("GET /organization-units/tree/{path...}",
		func(w http.ResponseWriter, r *http.Request) {
			pathValue := r.PathValue("path")
			handlers := map[string]func(http.ResponseWriter, *http.Request){
				"/ous":    ouHandler.HandleOUChildrenListByPathRequest,
				"/users":  ouHandler.HandleOUUsersListByPathRequest,
				"/groups": ouHandler.HandleOUGroupsListByPathRequest,
			}

			for suffix, handlerFunc := range handlers {
				if strings.HasSuffix(pathValue, suffix) {
					newPath := strings.TrimSuffix(pathValue, suffix)
					r.SetPathValue("path", newPath)
					handlerFunc(w, r)
					return
				}
			}

			newPath := "/organization-units/tree/" + pathValue
			r.URL.Path = newPath
			ouHandler.HandleOUGetByPathRequest(w, r)
		}, corsOptions2))
	mux.HandleFunc(middleware.WithCORS("PUT /organization-units/tree/{path...}",
		ouHandler.HandleOUPutByPathRequest, corsOptions2))
	mux.HandleFunc(middleware.WithCORS("DELETE /organization-units/tree/{path...}",
		ouHandler.HandleOUDeleteByPathRequest, corsOptions2))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /organization-units/tree/{path...}",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, corsOptions2))
}
