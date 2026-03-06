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

package user

import (
	"net/http"
	"strings"

	oupkg "github.com/asgardeo/thunder/internal/ou"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/crypto/hash"
	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/database/transaction"
	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/middleware"
	"github.com/asgardeo/thunder/internal/system/sysauthz"
	"github.com/asgardeo/thunder/internal/userschema"
)

// Initialize initializes the user service and registers its routes.
func Initialize(
	mux *http.ServeMux,
	ouService oupkg.OrganizationUnitServiceInterface,
	userSchemaService userschema.UserSchemaServiceInterface,
	hashService hash.HashServiceInterface,
	authzService sysauthz.SystemAuthorizationServiceInterface,
) (UserServiceInterface, declarativeresource.ResourceExporter, error) {
	// Step 1: Determine store mode and initialize store structure
	storeMode := getUserStoreMode()
	userStore, err := initializeStore(storeMode)
	if err != nil {
		return nil, nil, err
	}

	// Step 2: Get database transactioner (only needed for mutable modes)
	var transactioner transaction.Transactioner
	if storeMode == serverconst.StoreModeComposite || storeMode == serverconst.StoreModeMutable {
		dbProvider := provider.GetDBProvider()
		dbClient, err := dbProvider.GetUserDBClient()
		if err != nil {
			return nil, nil, err
		}
		transactioner, err = dbClient.GetTransactioner()
		if err != nil {
			return nil, nil, err
		}
	}

	// Step 3: Create service with store
	userService := newUserService(authzService, userStore, ouService, userSchemaService, hashService, transactioner)
	setUserService(userService) // Set the provider for backward compatibility

	// Step 4: Load declarative resources into store (if applicable)
	if storeMode == serverconst.StoreModeComposite || storeMode == serverconst.StoreModeDeclarative {
		if err := loadDeclarativeUserResources(userStore); err != nil {
			return nil, nil, err
		}
	}

	userHandler := newUserHandler(userService)
	registerRoutes(mux, userHandler)

	// Create and return exporter
	exporter := newUserExporter(userService)
	return userService, exporter, nil
}

// Store Selection (based on user.store configuration):
//
// 1. MUTABLE mode (store: "mutable"):
//   - Uses database store only
//   - Supports full CRUD operations (Create/Read/Update/Delete)
//   - All users are mutable
//   - Export functionality exports DB-backed users
//
// 2. IMMUTABLE mode (store: "declarative"):
//   - Uses file-based store only (from YAML resources)
//   - All users are immutable (read-only)
//   - No create/update/delete operations allowed
//   - Export functionality not applicable
//
// 3. COMPOSITE mode (store: "composite" - hybrid):
//   - Uses both file-based store (immutable) + database store (mutable)
//   - YAML resources are loaded into file-based store (immutable, read-only)
//   - Database store handles runtime users (mutable)
//   - Reads check both stores (merged results)
//   - Writes only go to database store
//   - Declarative users cannot be updated or deleted
//   - Export only exports DB-backed users (not YAML)
//
// Configuration Fallback:
// - If user.store is not specified, falls back to global declarative_resources.enabled:
//   - If declarative_resources.enabled = true: behaves as IMMUTABLE mode
//   - If declarative_resources.enabled = false: behaves as MUTABLE mode
func initializeStore(storeMode serverconst.StoreMode) (userStoreInterface, error) {
	var userStore userStoreInterface

	switch storeMode {
	case serverconst.StoreModeComposite:
		fileStore := newUserFileBasedStore()
		dbStore, err := newUserStore()
		if err != nil {
			return nil, err
		}
		userStore = newCompositeUserStore(fileStore.(*userFileBasedStore), dbStore)

	case serverconst.StoreModeDeclarative:
		fileStore := newUserFileBasedStore()
		userStore = fileStore

	default:
		dbStore, err := newUserStore()
		if err != nil {
			return nil, err
		}
		userStore = dbStore
	}

	return userStore, nil
}

// loadDeclarativeUserResources loads declarative user resources from files.
func loadDeclarativeUserResources(userStore userStoreInterface) error {
	var fileStore userStoreInterface
	var dbStore userStoreInterface

	// Determine store type and extract file store
	switch store := userStore.(type) {
	case *compositeUserStore:
		// Composite mode: extract file store and db store from composite
		fileStore = store.fileStore
		dbStore = store.dbStore
	case *userFileBasedStore:
		// Declarative-only mode: only file store available
		fileStore = store
		dbStore = nil
	default:
		return nil // Mutable mode: no declarative resources to load
	}

	// Type assert to access Storer interface for resource loading
	fileBasedStore, ok := fileStore.(*userFileBasedStore)
	if !ok {
		return nil // Not a file-based store
	}

	return loadDeclarativeResources(fileBasedStore, dbStore)
}

// registerRoutes registers the routes for user management operations.
func registerRoutes(mux *http.ServeMux, userHandler *userHandler) {
	opts1 := middleware.CORSOptions{
		AllowedMethods:   "GET, POST",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("POST /users", userHandler.HandleUserPostRequest, opts1))
	mux.HandleFunc(middleware.WithCORS("GET /users", userHandler.HandleUserListRequest, opts1))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /users", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}, opts1))

	opts2 := middleware.CORSOptions{
		AllowedMethods:   "GET, PUT, DELETE",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("GET /users/",
		func(w http.ResponseWriter, r *http.Request) {
			path := strings.TrimPrefix(r.URL.Path, "/users/")
			segments := strings.Split(path, "/")
			r.SetPathValue("id", segments[0])

			if len(segments) == 1 {
				userHandler.HandleUserGetRequest(w, r)
			} else if len(segments) == 2 && segments[1] == "groups" {
				userHandler.HandleUserGroupsGetRequest(w, r)
			} else {
				http.NotFound(w, r)
			}
		}, opts2))
	mux.HandleFunc(middleware.WithCORS("PUT /users/", userHandler.HandleUserPutRequest, opts2))
	mux.HandleFunc(middleware.WithCORS("DELETE /users/", userHandler.HandleUserDeleteRequest, opts2))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /users/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}, opts2))

	optsSelf := middleware.CORSOptions{
		AllowedMethods:   "GET, PUT",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("GET /users/me", userHandler.HandleSelfUserGetRequest, optsSelf))
	mux.HandleFunc(middleware.WithCORS("PUT /users/me", userHandler.HandleSelfUserPutRequest, optsSelf))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /users/me", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}, optsSelf))

	optsSelfCredentials := middleware.CORSOptions{
		AllowedMethods:   "POST",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("POST /users/me/update-credentials",
		userHandler.HandleSelfUserCredentialUpdateRequest, optsSelfCredentials))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /users/me/update-credentials",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, optsSelfCredentials))

	opts3 := middleware.CORSOptions{
		AllowedMethods:   "GET, POST",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("GET /users/tree/{path...}",
		userHandler.HandleUserListByPathRequest, opts3))
	mux.HandleFunc(middleware.WithCORS("POST /users/tree/{path...}",
		userHandler.HandleUserPostByPathRequest, opts3))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /users/tree/{path...}",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, opts3))
}
