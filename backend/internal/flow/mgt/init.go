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

package flowmgt

import (
	"net/http"

	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/flow/executor"
	"github.com/asgardeo/thunder/internal/system/config"
	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/middleware"
)

// Initialize initializes the flow management service and registers HTTP routes.
func Initialize(
	mux *http.ServeMux,
	flowFactory core.FlowFactoryInterface,
	executorRegistry executor.ExecutorRegistryInterface,
	graphCache core.GraphCacheInterface,
) (FlowMgtServiceInterface, declarativeresource.ResourceExporter, error) {
	var store flowStoreInterface
	if config.GetThunderRuntime().Config.DeclarativeResources.Enabled {
		store = newFileBasedStore()
	} else {
		store = newCacheBackedFlowStore()
	}

	inferenceService := newFlowInferenceService()
	graphBuilder := newGraphBuilder(flowFactory, executorRegistry, graphCache)
	service := newFlowMgtService(store, inferenceService, graphBuilder, executorRegistry)

	if config.GetThunderRuntime().Config.DeclarativeResources.Enabled {
		if err := loadDeclarativeResources(store); err != nil {
			return nil, nil, err
		}
	}

	handler := newFlowMgtHandler(service)
	registerRoutes(mux, handler)

	// Create and return exporter
	exporter := newFlowGraphExporter(service)
	return service, exporter, nil
}

// registerRoutes registers the HTTP routes for flow management.
func registerRoutes(mux *http.ServeMux, handler *flowMgtHandler) {
	opts1 := middleware.CORSOptions{
		AllowedMethods:   "GET, POST",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("GET /flows", handler.listFlows, opts1))
	mux.HandleFunc(middleware.WithCORS("POST /flows", handler.createFlow, opts1))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /flows", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}, opts1))

	opts2 := middleware.CORSOptions{
		AllowedMethods:   "GET, PUT, DELETE",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("GET /flows/{flowId}", handler.getFlow, opts2))
	mux.HandleFunc(middleware.WithCORS("PUT /flows/{flowId}", handler.updateFlow, opts2))
	mux.HandleFunc(middleware.WithCORS("DELETE /flows/{flowId}", handler.deleteFlow, opts2))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /flows/{flowId}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}, opts2))

	opts3 := middleware.CORSOptions{
		AllowedMethods:   "GET",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("GET /flows/{flowId}/versions", handler.listFlowVersions, opts3))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /flows/{flowId}/versions",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, opts3),
	)
	mux.HandleFunc(middleware.WithCORS("GET /flows/{flowId}/versions/{version}", handler.getFlowVersion, opts3))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /flows/{flowId}/versions/{version}",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, opts3),
	)

	opts4 := middleware.CORSOptions{
		AllowedMethods:   "POST",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("POST /flows/{flowId}/restore", handler.restoreFlowVersion, opts4))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /flows/{flowId}/restore",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, opts4),
	)
}
