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
	immutableresource "github.com/asgardeo/thunder/internal/system/immutable_resource"
	"github.com/asgardeo/thunder/internal/system/middleware"
)

// Initialize initializes the organization unit service and registers its routes.
func Initialize(mux *http.ServeMux) (OrganizationUnitServiceInterface, immutableresource.ResourceExporter, error) {
	ouStore, err := initializeStore()
	if err != nil {
		return nil, nil, err
	}

	ouService := newOrganizationUnitService(ouStore)

	ouHandler := newOrganizationUnitHandler(ouService)
	registerRoutes(mux, ouHandler)

	// Create and return exporter
	exporter := newOUExporter(ouService)
	return ouService, exporter, nil
}

// Store Selection (based on organization_unit.store configuration):
//
// 1. MUTABLE mode (store: "mutable"):
//   - Uses database store only (organizationUnitStore)
//   - Supports full CRUD operations (Create/Read/Update/Delete)
//   - All OUs are mutable
//   - Export functionality exports DB-backed OUs
//
// 2. IMMUTABLE mode (store: "immutable"):
//   - Uses file-based store only (from YAML resources)
//   - All OUs are immutable (read-only)
//   - No create/update/delete operations allowed
//   - Export functionality not applicable
//
// 3. COMPOSITE mode (store: "composite" - hybrid):
//   - Uses both file-based store (immutable) + database store (mutable)
//   - YAML resources are loaded into file-based store (immutable, read-only)
//   - Database store handles runtime OUs (mutable)
//   - Reads check both stores (merged results)
//   - Writes only go to database store
//   - Immutable OUs cannot be updated or deleted
//   - Export only exports DB-backed OUs (not YAML)
//
// Configuration Fallback:
// - If organization_unit.store is not specified, falls back to global immutable_resources.enabled:
//   - If immutable_resources.enabled = true: behaves as IMMUTABLE mode
//   - If immutable_resources.enabled = false: behaves as MUTABLE mode
func initializeStore() (organizationUnitStoreInterface, error) {
	var ouStore organizationUnitStoreInterface

	storeMode := getOrganizationUnitStoreMode()

	switch storeMode {
	case config.StoreModeComposite:
		fileStore := newFileBasedStore()
		dbStore := newOrganizationUnitStore()
		ouStore = newCompositeOUStore(fileStore, dbStore)
		if err := loadImmutableResources(fileStore, dbStore); err != nil {
			return nil, err
		}

	case config.StoreModeImmutable:
		fileStore := newFileBasedStore()
		ouStore = fileStore

		if err := loadImmutableResources(fileStore, nil); err != nil {
			return nil, err
		}

	default:
		ouStore = newOrganizationUnitStore()
	}

	return ouStore, nil
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
	mux.HandleFunc(middleware.WithCORS("OPTIONS /organization-units/",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
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
