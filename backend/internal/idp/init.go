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

// Package idp handles the identity provider management operations.
package idp

import (
	"net/http"

	"github.com/asgardeo/thunder/internal/system/middleware"
)

// Initialize initializes the IDP service and registers its routes.
func Initialize(mux *http.ServeMux) IDPServiceInterface {
	idpService := newIDPService()
	idpHandler := newIDPHandler(idpService)
	registerRoutes(mux, idpHandler)
	return idpService
}

// RegisterRoutes registers the routes for identity provider operations.
func registerRoutes(mux *http.ServeMux, idpHandler *idpHandler) {
	opts1 := middleware.CORSOptions{
		AllowedMethods:   "GET, POST",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("POST /identity-providers", idpHandler.HandleIDPPostRequest, opts1))
	mux.HandleFunc(middleware.WithCORS("GET /identity-providers", idpHandler.HandleIDPListRequest, opts1))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /identity-providers",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, opts1))

	opts2 := middleware.CORSOptions{
		AllowedMethods:   "GET, PUT, DELETE",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("GET /identity-providers/{id}",
		idpHandler.HandleIDPGetRequest, opts2))
	mux.HandleFunc(middleware.WithCORS("PUT /identity-providers/{id}",
		idpHandler.HandleIDPPutRequest, opts2))
	mux.HandleFunc(middleware.WithCORS("DELETE /identity-providers/{id}",
		idpHandler.HandleIDPDeleteRequest, opts2))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /identity-providers/{id}",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, opts2))
}
