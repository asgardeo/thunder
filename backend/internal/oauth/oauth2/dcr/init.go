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

// Package dcr provides Dynamic Client Registration (DCR) implementation.
package dcr

import (
	"net/http"

	"github.com/asgardeo/thunder/internal/application"
	"github.com/asgardeo/thunder/internal/system/middleware"
)

// Initialize initializes the DCR service and registers its routes.
func Initialize(mux *http.ServeMux, appService application.ApplicationServiceInterface) DCRServiceInterface {
	dcrService := newDCRService(appService)
	dcrHandler := newDCRHandler(dcrService)
	registerRoutes(mux, dcrHandler)
	return dcrService
}

// registerRoutes registers the routes for DCR operations.
func registerRoutes(mux *http.ServeMux, dcrHandler *dcrHandler) {
	opts := middleware.CORSOptions{
		AllowedMethods:   "POST, OPTIONS",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}

	mux.HandleFunc(middleware.WithCORS("POST /oauth2/dcr/register",
		dcrHandler.HandleDCRRegistration, opts))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /oauth2/dcr/register",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, opts))
}
