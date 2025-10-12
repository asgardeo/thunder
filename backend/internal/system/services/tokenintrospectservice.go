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

package services

import (
	"net/http"

	"github.com/asgardeo/thunder/internal/oauth/oauth2/introspect"
	"github.com/asgardeo/thunder/internal/system/jwt"
	"github.com/asgardeo/thunder/internal/system/middleware"
)

// TODO: Introspection endpoint MUST require authentication and authorization.
//  Implement this when the support is added.

// TokenIntrospectionAPIService defines the API service for handling OAuth 2.0 token introspection requests.
type TokenIntrospectionAPIService struct {
	introspectHandler *introspect.TokenIntrospectionHandler
}

// NewIntrospectionAPIService creates a new instance of IntrospectionAPIService.
func NewIntrospectionAPIService(mux *http.ServeMux) ServiceInterface {
	jwtService := jwt.GetJWTService()
	introspectionService := introspect.NewTokenIntrospectionService(jwtService)

	instance := &TokenIntrospectionAPIService{
		introspectHandler: introspect.NewTokenIntrospectionHandler(introspectionService),
	}
	instance.RegisterRoutes(mux)

	return instance
}

// RegisterRoutes registers the routes for the IntrospectionAPIService.
func (s *TokenIntrospectionAPIService) RegisterRoutes(mux *http.ServeMux) {
	opts := middleware.CORSOptions{
		AllowedMethods:   "POST, OPTIONS",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}

	mux.HandleFunc(middleware.WithCORS("POST /oauth2/introspect",
		s.introspectHandler.HandleIntrospect, opts))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /oauth2/introspect",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, opts))
}
