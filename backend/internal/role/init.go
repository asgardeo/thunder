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

package role

import (
	"net/http"
	"strings"

	"github.com/asgardeo/thunder/internal/group"
	oupkg "github.com/asgardeo/thunder/internal/ou"
	resourcepkg "github.com/asgardeo/thunder/internal/resource"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/database/provider"
	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/middleware"
	"github.com/asgardeo/thunder/internal/user"
)

// Initialize initializes the role service and registers its routes.
func Initialize(
	mux *http.ServeMux,
	userService user.UserServiceInterface,
	groupService group.GroupServiceInterface,
	ouService oupkg.OrganizationUnitServiceInterface,
	resourceService resourcepkg.ResourceServiceInterface,
) (RoleServiceInterface, declarativeresource.ResourceExporter, error) {
	// Step 1: Initialize store based on configuration
	roleStore, err := initializeStore()
	if err != nil {
		return nil, nil, err
	}

	// Get transactioner from DB provider
	dbProvider := provider.GetDBProvider()
	dbClient, err := dbProvider.GetConfigDBClient()
	if err != nil {
		log.GetLogger().Fatal("Failed to get DB client", log.Error(err))
	}
	transactioner, err := dbClient.GetTransactioner()
	if err != nil {
		log.GetLogger().Fatal("Failed to get transactioner", log.Error(err))
	}

	// Step 2: Create service with store
	roleService := newRoleService(roleStore, userService, groupService, ouService, resourceService, transactioner)
	roleHandler := newRoleHandler(roleService)
	registerRoutes(mux, roleHandler)
	exporter := newRoleExporter(roleService)
	return roleService, exporter, nil
}

// Store Selection (based on role.store configuration):
//
// 1. MUTABLE mode (store: "mutable"):
//   - Uses database store only
//   - Supports full CRUD operations (Create/Read/Update/Delete)
//   - All roles are mutable
//
// 2. IMMUTABLE mode (store: "declarative"):
//   - Uses file-based store only (from YAML resources)
//   - All roles are immutable (read-only)
//   - No create/update/delete operations allowed
//
// 3. COMPOSITE mode (store: "composite" - hybrid):
//   - Uses both file-based store (immutable) + database store (mutable)
//   - YAML resources are loaded into file-based store (immutable, read-only)
//   - Database store handles runtime roles (mutable)
//   - Reads check both stores (merged results)
//   - Writes only go to database store
//   - Declarative roles cannot be updated or deleted
//
// Configuration Fallback:
// - If role.store is not specified, falls back to global declarative_resources.enabled:
//   - If declarative_resources.enabled = true: behaves as IMMUTABLE mode
//   - If declarative_resources.enabled = false: behaves as MUTABLE mode
func initializeStore() (roleStoreInterface, error) {
	var roleStore roleStoreInterface

	storeMode := getRoleStoreMode()

	switch storeMode {
	case serverconst.StoreModeComposite:
		fileStoreInterface := newFileBasedStore()
		fileStore := fileStoreInterface.(*fileBasedStore)
		dbStore := newRoleStore()
		roleStore = newCompositeRoleStore(fileStoreInterface, dbStore)
		if err := loadDeclarativeResources(fileStore, dbStore); err != nil {
			return nil, err
		}

	case serverconst.StoreModeDeclarative:
		fileStoreInterface := newFileBasedStore()
		fileStore := fileStoreInterface.(*fileBasedStore)
		roleStore = fileStoreInterface
		if err := loadDeclarativeResources(fileStore, nil); err != nil {
			return nil, err
		}

	default:
		dbStore := newRoleStore()
		roleStore = dbStore
	}

	return roleStore, nil
}

// registerRoutes registers the routes for role management operations.
func registerRoutes(mux *http.ServeMux, roleHandler *roleHandler) {
	opts1 := middleware.CORSOptions{
		AllowedMethods:   "GET, POST",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("POST /roles", roleHandler.HandleRolePostRequest, opts1))
	mux.HandleFunc(middleware.WithCORS("GET /roles", roleHandler.HandleRoleListRequest, opts1))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /roles", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}, opts1))

	opts2 := middleware.CORSOptions{
		AllowedMethods:   "GET, PUT, DELETE",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	// Special handling for /roles/{id} and /roles/{id}/assignments
	mux.HandleFunc(middleware.WithCORS("GET /roles/",
		func(w http.ResponseWriter, r *http.Request) {
			path := strings.TrimPrefix(r.URL.Path, "/roles/")
			segments := strings.Split(path, "/")
			r.SetPathValue("id", segments[0])

			if len(segments) == 1 {
				roleHandler.HandleRoleGetRequest(w, r)
			} else if len(segments) == 2 && segments[1] == "assignments" {
				roleHandler.HandleRoleAssignmentsGetRequest(w, r)
			} else {
				http.NotFound(w, r)
			}
		}, opts2))
	mux.HandleFunc(middleware.WithCORS("PUT /roles/{id}", roleHandler.HandleRolePutRequest, opts2))
	mux.HandleFunc(middleware.WithCORS("DELETE /roles/{id}", roleHandler.HandleRoleDeleteRequest, opts2))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /roles/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}, opts2))

	opts3 := middleware.CORSOptions{
		AllowedMethods:   "POST",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("POST /roles/{id}/assignments/add",
		roleHandler.HandleRoleAddAssignmentsRequest, opts3))
	mux.HandleFunc(middleware.WithCORS("POST /roles/{id}/assignments/remove",
		roleHandler.HandleRoleRemoveAssignmentsRequest, opts3))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /roles/{id}/assignments/add",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, opts3))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /roles/{id}/assignments/remove",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, opts3))
}
