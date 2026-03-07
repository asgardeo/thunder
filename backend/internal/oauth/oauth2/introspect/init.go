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

package introspect

import (
	"net/http"

	"github.com/asgardeo/thunder/internal/application"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/clientauth"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/discovery"
	"github.com/asgardeo/thunder/internal/system/jose/jwt"
	"github.com/asgardeo/thunder/internal/system/middleware"
)

// Initialize initializes the token introspection handler and registers its routes.
func Initialize(
	mux *http.ServeMux,
	jwtService jwt.JWTServiceInterface,
	appService application.ApplicationServiceInterface,
	discoveryService discovery.DiscoveryServiceInterface,
) TokenIntrospectionServiceInterface {
	introspectionService := newTokenIntrospectionService(jwtService)
	introspectHandler := newTokenIntrospectionHandler(introspectionService)
	registerRoutes(mux, introspectHandler, appService, jwtService, discoveryService)
	return introspectionService
}

// registerRoutes registers the routes for the IntrospectionAPIService.
func registerRoutes(
	mux *http.ServeMux,
	introspectHandler *tokenIntrospectionHandler,
	appService application.ApplicationServiceInterface,
	jwtService jwt.JWTServiceInterface,
	discoveryService discovery.DiscoveryServiceInterface,
) {
	opts := middleware.CORSOptions{
		AllowedMethods:   "POST, OPTIONS",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}

	clientAuthMiddleware := clientauth.ClientAuthMiddleware(appService, jwtService, discoveryService)
	handler := clientAuthMiddleware(http.HandlerFunc(introspectHandler.HandleIntrospect))

	pattern, wrappedHandler := middleware.WithCORS(
		"POST /oauth2/introspect",
		handler.ServeHTTP,
		opts,
	)
	mux.HandleFunc(pattern, wrappedHandler)
	mux.HandleFunc(middleware.WithCORS("OPTIONS /oauth2/introspect",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, opts))
}
