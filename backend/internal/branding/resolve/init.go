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

package brandingresolve

import (
	"net/http"

	"github.com/asgardeo/thunder/internal/application"
	brandingmgt "github.com/asgardeo/thunder/internal/branding/mgt"
	"github.com/asgardeo/thunder/internal/system/middleware"
)

// Initialize initializes the branding resolve service and registers its routes.
func Initialize(
	mux *http.ServeMux,
	brandingMgtService brandingmgt.BrandingMgtServiceInterface,
	applicationService application.ApplicationServiceInterface,
) BrandingResolveServiceInterface {
	brandingResolveService := newBrandingResolveService(brandingMgtService, applicationService)
	brandingResolveHandler := newBrandingResolveHandler(brandingResolveService)
	registerRoutes(mux, brandingResolveHandler)
	return brandingResolveService
}

// registerRoutes registers the routes for branding resolve operations.
func registerRoutes(mux *http.ServeMux, resolveHandler *brandingResolveHandler) {
	opts := middleware.CORSOptions{
		AllowedMethods:   "GET",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("GET /branding/resolve", resolveHandler.HandleResolveRequest, opts))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /branding/resolve", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}, opts))
}
