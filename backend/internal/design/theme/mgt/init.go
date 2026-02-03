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

package thememgt

import (
	"net/http"

	"github.com/asgardeo/thunder/internal/system/middleware"
)

// Initialize initializes the theme management service and registers its routes.
func Initialize(mux *http.ServeMux) ThemeMgtServiceInterface {
	themeMgtStore := newThemeMgtStore()
	themeMgtService := newThemeMgtService(themeMgtStore)
	themeMgtHandler := newThemeMgtHandler(themeMgtService)
	registerRoutes(mux, themeMgtHandler)
	return themeMgtService
}

// registerRoutes registers the routes for theme management operations.
func registerRoutes(mux *http.ServeMux, themeMgtHandler *themeMgtHandler) {
	opts1 := middleware.CORSOptions{
		AllowedMethods:   "GET, POST",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("POST /design/themes", themeMgtHandler.HandleThemePostRequest, opts1))
	mux.HandleFunc(middleware.WithCORS("GET /design/themes", themeMgtHandler.HandleThemeListRequest, opts1))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /design/themes", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}, opts1))

	opts2 := middleware.CORSOptions{
		AllowedMethods:   "GET, PUT, DELETE",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("GET /design/themes/{id}", themeMgtHandler.HandleThemeGetRequest, opts2))
	mux.HandleFunc(middleware.WithCORS("PUT /design/themes/{id}", themeMgtHandler.HandleThemePutRequest, opts2))
	mux.HandleFunc(middleware.WithCORS("DELETE /design/themes/{id}", themeMgtHandler.HandleThemeDeleteRequest, opts2))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /design/themes/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}, opts2))
}
