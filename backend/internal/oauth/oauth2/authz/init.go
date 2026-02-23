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

package authz

import (
	"net/http"

	"github.com/asgardeo/thunder/internal/application"
	"github.com/asgardeo/thunder/internal/flow/flowexec"
	"github.com/asgardeo/thunder/internal/system/jose/jwt"
	"github.com/asgardeo/thunder/internal/system/middleware"
)

// Initialize initializes the authorization handler and registers its routes.
func Initialize(
	mux *http.ServeMux,
	applicationService application.ApplicationServiceInterface,
	jwtService jwt.JWTServiceInterface,
	flowExecService flowexec.FlowExecServiceInterface,
) AuthorizeServiceInterface {
	authzCodeStore := newAuthorizationCodeStore()
	authzReqStore := newAuthorizationRequestStore()
	authzService := newAuthorizeService(applicationService, jwtService, flowExecService, authzCodeStore, authzReqStore)
	authzHandler := newAuthorizeHandler(authzService)
	registerRoutes(mux, authzHandler)
	return authzService
}

// registerRoutes registers the routes for OAuth2 authorization operations.
func registerRoutes(mux *http.ServeMux, authzHandler AuthorizeHandlerInterface) {
	authorizeOpts := middleware.CORSOptions{
		AllowedMethods:   "GET",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}

	callbackOpts := middleware.CORSOptions{
		AllowedMethods:   "POST",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}

	mux.HandleFunc(middleware.WithCORS("GET /oauth2/authorize",
		authzHandler.HandleAuthorizeGetRequest, authorizeOpts))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /oauth2/authorize",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, authorizeOpts))

	mux.HandleFunc(middleware.WithCORS("POST /oauth2/auth/callback",
		authzHandler.HandleAuthCallbackPostRequest, callbackOpts))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /oauth2/auth/callback",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, callbackOpts))
}
