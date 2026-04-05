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

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/asgardeo/thunder/internal/cert"
	"github.com/asgardeo/thunder/internal/consent"
	layoutmgt "github.com/asgardeo/thunder/internal/design/layout/mgt"
	thememgt "github.com/asgardeo/thunder/internal/design/theme/mgt"
	"github.com/asgardeo/thunder/internal/entity"
	"github.com/asgardeo/thunder/internal/entityprovider"
	flowmgt "github.com/asgardeo/thunder/internal/flow/mgt"
	oupkg "github.com/asgardeo/thunder/internal/ou"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/middleware"
	"github.com/asgardeo/thunder/internal/system/transaction"
	"github.com/asgardeo/thunder/internal/userschema"
)

// Initialize initializes the application service and registers its routes.
func Initialize(
	mux *http.ServeMux,
	mcpServer *mcp.Server,
	entityProvider entityprovider.EntityProviderInterface,
	entityService entity.EntityServiceInterface,
	ouService oupkg.OrganizationUnitServiceInterface,
	certService cert.CertificateServiceInterface,
	flowMgtService flowmgt.FlowMgtServiceInterface,
	themeMgtService thememgt.ThemeMgtServiceInterface,
	layoutMgtService layoutmgt.LayoutMgtServiceInterface,
	userSchemaService userschema.UserSchemaServiceInterface,
	consentService consent.ConsentServiceInterface,
) (ApplicationServiceInterface, declarativeresource.ResourceExporter, error) {
	// Step 1: Initialize store and transactioner based on store mode
	appStore, transactioner, err := initializeStore()
	if err != nil {
		return nil, nil, err
	}

	// Step 2: Create service with store and entity provider
	appService := newApplicationService(
		appStore, entityProvider, ouService, certService, flowMgtService,
		themeMgtService, layoutMgtService,
		userSchemaService, consentService,
		transactioner,
	)

	// Step 3: Register indexed attributes for automatic identifier syncing.
	if err := entityService.LoadIndexedAttributes(getAppIndexedAttributes()); err != nil {
		return nil, nil, err
	}

	// Step 4: Load declarative resources if store mode requires it.
	storeMode := getApplicationStoreMode()
	if storeMode == serverconst.StoreModeComposite || storeMode == serverconst.StoreModeDeclarative {
		if err := entityService.LoadDeclarativeResources(makeAppDeclarativeConfig()); err != nil {
			return nil, nil, err
		}
		if err := loadDeclarativeResources(appStore, appService); err != nil {
			return nil, nil, err
		}
	}

	appHandler := newApplicationHandler(appService)
	registerRoutes(mux, appHandler)

	// Register MCP tools
	if mcpServer != nil {
		registerMCPTools(mcpServer, appService)
	}

	// Create and return exporter
	exporter := newApplicationExporter(appService)
	return appService, exporter, nil
}

// Store Selection (based on application.store configuration):
//
// 1. MUTABLE mode (store: "mutable"):
//   - Uses database store only with cache (cachedBackedApplicationStore)
//   - Supports full CRUD operations (Create/Read/Update/Delete)
//   - All applications are mutable
//   - Export functionality exports DB-backed applications
//
// 2. IMMUTABLE mode (store: "declarative"):
//   - Uses file-based store only (from YAML resources)
//   - All applications are immutable (read-only)
//   - No create/update/delete operations allowed
//   - Export functionality not applicable
//
// 3. COMPOSITE mode (store: "composite" - hybrid):
//   - Uses both file-based store (immutable) + database store (mutable)
//   - YAML resources are loaded into file-based store (immutable, read-only)
//   - Database store handles runtime applications (mutable)
//   - Reads check both stores (merged results)
//   - Writes only go to database store
//   - Declarative applications cannot be updated or deleted
//   - Export only exports DB-backed applications (not YAML)
//
// Configuration Fallback:
// - If application.store is not specified, falls back to global declarative_resources.enabled:
//   - If declarative_resources.enabled = true: behaves as IMMUTABLE mode
//   - If declarative_resources.enabled = false: behaves as MUTABLE mode
func initializeStore() (applicationStoreInterface, transaction.Transactioner, error) {
	storeMode := getApplicationStoreMode()

	switch storeMode {
	case serverconst.StoreModeComposite:
		fileStore, _ := newFileBasedStore()
		dbStore, transactioner, err := newApplicationStore()
		if err != nil {
			return nil, nil, err
		}
		return newCompositeApplicationStore(fileStore, dbStore), transactioner, nil

	case serverconst.StoreModeDeclarative:
		fileStore, transactioner := newFileBasedStore()
		return fileStore, transactioner, nil

	default:
		dbStore, transactioner, err := newApplicationStore()
		if err != nil {
			return nil, nil, err
		}
		return newCachedBackedApplicationStore(dbStore), transactioner, nil
	}
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
