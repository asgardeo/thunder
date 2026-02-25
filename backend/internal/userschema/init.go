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
	"net/http"

	oupkg "github.com/asgardeo/thunder/internal/ou"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/database/transaction"
	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/middleware"
	"github.com/asgardeo/thunder/internal/system/sysauthz"
)

// Initialize initializes the user schema service and registers its routes.
func Initialize(
	mux *http.ServeMux,
	ouService oupkg.OrganizationUnitServiceInterface,
	authzService sysauthz.SystemAuthorizationServiceInterface,
) (UserSchemaServiceInterface, declarativeresource.ResourceExporter, error) {
	// Step 1: Determine store mode and initialize store structure
	storeMode := getUserSchemaStoreMode()
	userSchemaStore := initializeStore(storeMode)

	// Step 2: Get database transactioner (only needed for mutable modes)
	var transactioner transaction.Transactioner
	var err error
	if storeMode == serverconst.StoreModeComposite || storeMode == serverconst.StoreModeMutable {
		dbProvider := provider.GetDBProvider()
		transactioner, err = dbProvider.GetConfigDBTransactioner()
		if err != nil {
			return nil, nil, err
		}
	}

	// Step 3: Create service with store
	userSchemaService := newUserSchemaService(ouService, userSchemaStore, transactioner, authzService)

	// Step 4: Load declarative resources into store (if applicable)
	if storeMode == serverconst.StoreModeComposite || storeMode == serverconst.StoreModeDeclarative {
		if err := loadDeclarativeResources(userSchemaStore, ouService); err != nil {
			return nil, nil, err
		}
	}

	userSchemaHandler := newUserSchemaHandler(userSchemaService)
	registerRoutes(mux, userSchemaHandler)

	// Create and return exporter
	exporter := newUserSchemaExporter(userSchemaService)
	return userSchemaService, exporter, nil
}

// Store Selection (based on user_schema.store configuration):
//
// 1. MUTABLE mode (store: "mutable"):
//   - Uses database store only
//   - Supports full CRUD operations (Create/Read/Update/Delete)
//   - All user schemas are mutable
//   - Export functionality exports DB-backed user schemas
//
// 2. IMMUTABLE mode (store: "declarative"):
//   - Uses file-based store only (from YAML resources)
//   - All user schemas are immutable (read-only)
//   - No create/update/delete operations allowed
//   - Export functionality not applicable
//
// 3. COMPOSITE mode (store: "composite" - hybrid):
//   - Uses both file-based store (immutable) + database store (mutable)
//   - YAML resources are loaded into file-based store (immutable, read-only)
//   - Database store handles runtime user schemas (mutable)
//   - Reads check both stores (merged results)
//   - Writes only go to database store
//   - Declarative user schemas cannot be updated or deleted
//   - Export only exports DB-backed user schemas (not YAML)
//
// Configuration Fallback:
// - If user_schema.store is not specified, falls back to global declarative_resources.enabled:
//   - If declarative_resources.enabled = true: behaves as IMMUTABLE mode
//   - If declarative_resources.enabled = false: behaves as MUTABLE mode
func initializeStore(storeMode serverconst.StoreMode) userSchemaStoreInterface {
	var userSchemaStore userSchemaStoreInterface

	switch storeMode {
	case serverconst.StoreModeComposite:
		fileStore := newUserSchemaFileBasedStore()
		dbStore := newUserSchemaStore()
		userSchemaStore = newCompositeUserSchemaStore(fileStore, dbStore)

	case serverconst.StoreModeDeclarative:
		fileStore := newUserSchemaFileBasedStore()
		userSchemaStore = fileStore

	default:
		dbStore := newUserSchemaStore()
		userSchemaStore = dbStore
	}

	return userSchemaStore
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
