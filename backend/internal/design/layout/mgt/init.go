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

package layoutmgt

import (
	"net/http"

	"github.com/asgardeo/thunder/internal/system/middleware"
)

// Initialize initializes the layout management service and registers its routes.
func Initialize(mux *http.ServeMux) LayoutMgtServiceInterface {
	layoutMgtStore := newLayoutMgtStore()
	layoutMgtService := newLayoutMgtService(layoutMgtStore)
	layoutMgtHandler := newLayoutMgtHandler(layoutMgtService)
	registerRoutes(mux, layoutMgtHandler)
	return layoutMgtService
}

// registerRoutes registers the routes for layout management operations.
func registerRoutes(mux *http.ServeMux, layoutMgtHandler *layoutMgtHandler) {
	opts1 := middleware.CORSOptions{
		AllowedMethods:   "GET, POST",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("POST /design/layouts", layoutMgtHandler.HandleLayoutPostRequest, opts1))
	mux.HandleFunc(middleware.WithCORS("GET /design/layouts", layoutMgtHandler.HandleLayoutListRequest, opts1))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /design/layouts", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}, opts1))

	opts2 := middleware.CORSOptions{
		AllowedMethods:   "GET, PUT, DELETE",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("GET /design/layouts/{id}", layoutMgtHandler.HandleLayoutGetRequest, opts2))
	mux.HandleFunc(middleware.WithCORS("PUT /design/layouts/{id}", layoutMgtHandler.HandleLayoutPutRequest, opts2))
	mux.HandleFunc(middleware.WithCORS("DELETE /design/layouts/{id}", layoutMgtHandler.HandleLayoutDeleteRequest, opts2))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /design/layouts/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}, opts2))
}
