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

package userschema

import (
	"net/http"

	"github.com/asgardeo/thunder/internal/system/middleware"
)

// Initialize initializes the user schema service and registers its routes.
func Initialize(mux *http.ServeMux) UserSchemaServiceInterface {
	userSchemaService := newUserSchemaService()
	userSchemaHandler := newUserSchemaHandler(userSchemaService)
	registerRoutes(mux, userSchemaHandler)
	return userSchemaService
}

// registerRoutes registers the routes for user schema management operations.
func registerRoutes(mux *http.ServeMux, userSchemaHandler *userSchemaHandler) {
	opts1 := middleware.CORSOptions{
		AllowedMethods:   "GET, POST",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("POST /user-schemas",
		userSchemaHandler.HandleUserSchemaPostRequest, opts1))
	mux.HandleFunc(middleware.WithCORS("GET /user-schemas",
		userSchemaHandler.HandleUserSchemaListRequest, opts1))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /user-schemas",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, opts1))

	opts2 := middleware.CORSOptions{
		AllowedMethods:   "GET, PUT, DELETE",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("GET /user-schemas/{id}",
		userSchemaHandler.HandleUserSchemaGetRequest, opts2))
	mux.HandleFunc(middleware.WithCORS("PUT /user-schemas/{id}",
		userSchemaHandler.HandleUserSchemaPutRequest, opts2))
	mux.HandleFunc(middleware.WithCORS("DELETE /user-schemas/{id}",
		userSchemaHandler.HandleUserSchemaDeleteRequest, opts2))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /user-schemas/{id}",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, opts2))
}
