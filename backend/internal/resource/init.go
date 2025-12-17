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

package resource

import (
	"net/http"

	oupkg "github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/system/middleware"
)

// Initialize initializes the resource service and registers its routes.
func Initialize(
	mux *http.ServeMux,
	ouService oupkg.OrganizationUnitServiceInterface,
) (ResourceServiceInterface, error) {
	resourceStore := newResourceStore()
	resourceService, err := newResourceService(resourceStore, ouService)
	if err != nil {
		return nil, err
	}
	resourceHandler := newResourceHandler(resourceService)
	registerRoutes(mux, resourceHandler)
	return resourceService, nil
}

// registerRoutes registers all routes for the resource management API.
func registerRoutes(mux *http.ServeMux, handler *resourceHandler) {
	// Resource Server routes
	resourceServerOpts := middleware.CORSOptions{
		AllowedMethods:   "GET, POST",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}

	mux.HandleFunc(middleware.WithCORS("GET /resource-servers",
		handler.HandleResourceServerListRequest, resourceServerOpts))
	mux.HandleFunc(middleware.WithCORS("POST /resource-servers",
		handler.HandleResourceServerPostRequest, resourceServerOpts))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /resource-servers",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, resourceServerOpts))

	resourceServerDetailOpts := middleware.CORSOptions{
		AllowedMethods:   "GET, PUT, DELETE",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}

	mux.HandleFunc(middleware.WithCORS("GET /resource-servers/{id}",
		handler.HandleResourceServerGetRequest, resourceServerDetailOpts))
	mux.HandleFunc(middleware.WithCORS("PUT /resource-servers/{id}",
		handler.HandleResourceServerPutRequest, resourceServerDetailOpts))
	mux.HandleFunc(middleware.WithCORS("DELETE /resource-servers/{id}",
		handler.HandleResourceServerDeleteRequest, resourceServerDetailOpts))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /resource-servers/{id}",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, resourceServerDetailOpts))

	// Resource routes
	resourceOpts := middleware.CORSOptions{
		AllowedMethods:   "GET, POST",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}

	mux.HandleFunc(middleware.WithCORS("GET /resource-servers/{rsId}/resources",
		handler.HandleResourceListRequest, resourceOpts))
	mux.HandleFunc(middleware.WithCORS("POST /resource-servers/{rsId}/resources",
		handler.HandleResourcePostRequest, resourceOpts))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /resource-servers/{rsId}/resources",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, resourceOpts))

	resourceDetailOpts := middleware.CORSOptions{
		AllowedMethods:   "GET, PUT, DELETE",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}

	mux.HandleFunc(middleware.WithCORS("GET /resource-servers/{rsId}/resources/{id}",
		handler.HandleResourceGetRequest, resourceDetailOpts))
	mux.HandleFunc(middleware.WithCORS("PUT /resource-servers/{rsId}/resources/{id}",
		handler.HandleResourcePutRequest, resourceDetailOpts))
	mux.HandleFunc(middleware.WithCORS("DELETE /resource-servers/{rsId}/resources/{id}",
		handler.HandleResourceDeleteRequest, resourceDetailOpts))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /resource-servers/{rsId}/resources/{id}",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, resourceDetailOpts))

	// Action routes (Resource Server level)
	actionRSOpts := middleware.CORSOptions{
		AllowedMethods:   "GET, POST",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}

	mux.HandleFunc(middleware.WithCORS("GET /resource-servers/{rsId}/actions",
		handler.HandleActionListAtResourceServerRequest, actionRSOpts))
	mux.HandleFunc(middleware.WithCORS("POST /resource-servers/{rsId}/actions",
		handler.HandleActionPostAtResourceServerRequest, actionRSOpts))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /resource-servers/{rsId}/actions",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, actionRSOpts))

	actionRSDetailOpts := middleware.CORSOptions{
		AllowedMethods:   "GET, PUT, DELETE",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}

	mux.HandleFunc(middleware.WithCORS("GET /resource-servers/{rsId}/actions/{id}",
		handler.HandleActionGetAtResourceServerRequest, actionRSDetailOpts))
	mux.HandleFunc(middleware.WithCORS("PUT /resource-servers/{rsId}/actions/{id}",
		handler.HandleActionPutAtResourceServerRequest, actionRSDetailOpts))
	mux.HandleFunc(middleware.WithCORS("DELETE /resource-servers/{rsId}/actions/{id}",
		handler.HandleActionDeleteAtResourceServerRequest, actionRSDetailOpts))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /resource-servers/{rsId}/actions/{id}",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, actionRSDetailOpts))

	// Action routes (Resource level)
	actionResourceOpts := middleware.CORSOptions{
		AllowedMethods:   "GET, POST",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}

	mux.HandleFunc(middleware.WithCORS("GET /resource-servers/{rsId}/resources/{resourceId}/actions",
		handler.HandleActionListAtResourceRequest, actionResourceOpts))
	mux.HandleFunc(middleware.WithCORS("POST /resource-servers/{rsId}/resources/{resourceId}/actions",
		handler.HandleActionPostAtResourceRequest, actionResourceOpts))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /resource-servers/{rsId}/resources/{resourceId}/actions",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, actionResourceOpts))

	actionResourceDetailOpts := middleware.CORSOptions{
		AllowedMethods:   "GET, PUT, DELETE",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}

	mux.HandleFunc(middleware.WithCORS("GET /resource-servers/{rsId}/resources/{resourceId}/actions/{id}",
		handler.HandleActionGetAtResourceRequest, actionResourceDetailOpts))
	mux.HandleFunc(middleware.WithCORS("PUT /resource-servers/{rsId}/resources/{resourceId}/actions/{id}",
		handler.HandleActionPutAtResourceRequest, actionResourceDetailOpts))
	mux.HandleFunc(middleware.WithCORS("DELETE /resource-servers/{rsId}/resources/{resourceId}/actions/{id}",
		handler.HandleActionDeleteAtResourceRequest, actionResourceDetailOpts))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /resource-servers/{rsId}/resources/{resourceId}/actions/{id}",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, actionResourceDetailOpts))
}
