/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

package flowmeta

import (
	"net/http"

	"github.com/asgardeo/thunder/internal/application"
	"github.com/asgardeo/thunder/internal/design/resolve"
	"github.com/asgardeo/thunder/internal/ou"
	i18nmgt "github.com/asgardeo/thunder/internal/system/i18n/mgt"
	"github.com/asgardeo/thunder/internal/system/middleware"
)

// Initialize creates and configures the flow metadata service components.
func Initialize(
	mux *http.ServeMux,
	applicationService application.ApplicationServiceInterface,
	ouService ou.OrganizationUnitServiceInterface,
	designResolve resolve.DesignResolveServiceInterface,
	i18nService i18nmgt.I18nServiceInterface,
) FlowMetaServiceInterface {
	// Create service instance
	flowMetaService := newFlowMetaService(applicationService, ouService, designResolve, i18nService)

	// Create handler and register routes
	handler := newFlowMetaHandler(flowMetaService)
	registerRoutes(mux, handler)

	return flowMetaService
}

func registerRoutes(mux *http.ServeMux, handler *flowMetaHandler) {
	// CORS options for flow metadata endpoint (follows the same security as flow/execute)
	opts := middleware.CORSOptions{
		AllowedMethods:   "GET, OPTIONS",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}

	// Register GET endpoint
	mux.HandleFunc(middleware.WithCORS("GET /flow/meta",
		middleware.CorrelationIDMiddleware(http.HandlerFunc(handler.HandleGetFlowMetadata)).ServeHTTP, opts))

	// Register OPTIONS endpoint for CORS preflight
	mux.HandleFunc(middleware.WithCORS("OPTIONS /flow/meta",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, opts))
}
