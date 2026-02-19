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
	layoutmgt "github.com/asgardeo/thunder/internal/design/layout/mgt"
	thememgt "github.com/asgardeo/thunder/internal/design/theme/mgt"
	flowmgt "github.com/asgardeo/thunder/internal/flow/mgt"
	"github.com/asgardeo/thunder/internal/system/config"
	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/middleware"
	"github.com/asgardeo/thunder/internal/usertype"
)

// Initialize initializes the application service and registers its routes.
func Initialize(
	mux *http.ServeMux,
	mcpServer *mcp.Server,
	certService cert.CertificateServiceInterface,
	flowMgtService flowmgt.FlowMgtServiceInterface,
	themeMgtService thememgt.ThemeMgtServiceInterface,
	layoutMgtService layoutmgt.LayoutMgtServiceInterface,
	userTypeService usertype.UserTypeServiceInterface,
) (ApplicationServiceInterface, declarativeresource.ResourceExporter, error) {
	var appStore applicationStoreInterface
	if config.GetThunderRuntime().Config.DeclarativeResources.Enabled {
		appStore = newFileBasedStore()
	} else {
		store := newApplicationStore()
		appStore = newCachedBackedApplicationStore(store)
	}

	appService := newApplicationService(
		appStore, certService, flowMgtService, themeMgtService, layoutMgtService, userTypeService)

	if config.GetThunderRuntime().Config.DeclarativeResources.Enabled {
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
