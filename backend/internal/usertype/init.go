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

package usertype

import (
	"net/http"

	oupkg "github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/system/config"
	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/middleware"
)

// Initialize initializes the user type service and registers its routes.
func Initialize(
	mux *http.ServeMux,
	ouService oupkg.OrganizationUnitServiceInterface,
) (UserTypeServiceInterface, declarativeresource.ResourceExporter, error) {
	var userTypeStore userTypeStoreInterface
	if config.GetThunderRuntime().Config.DeclarativeResources.Enabled {
		userTypeStore = newUserTypeFileBasedStore()
	} else {
		userTypeStore = newUserTypeStore()
	}

	userTypeService := newUserTypeService(ouService, userTypeStore)

	if config.GetThunderRuntime().Config.DeclarativeResources.Enabled {
		if err := loadDeclarativeResources(userTypeStore, ouService); err != nil {
			return nil, nil, err
		}
	}

	userTypeHandler := newUserTypeHandler(userTypeService)
	registerRoutes(mux, userTypeHandler)

	// Create and return exporter
	exporter := newUserTypeExporter(userTypeService)
	return userTypeService, exporter, nil
}

// registerRoutes registers the routes for user type management operations.
func registerRoutes(mux *http.ServeMux, userTypeHandler *userTypeHandler) {
	opts1 := middleware.CORSOptions{
		AllowedMethods:   "GET, POST",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("POST /user-types",
		userTypeHandler.HandleUserTypePostRequest, opts1))
	mux.HandleFunc(middleware.WithCORS("GET /user-types",
		userTypeHandler.HandleUserTypeListRequest, opts1))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /user-types",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, opts1))

	opts2 := middleware.CORSOptions{
		AllowedMethods:   "GET, PUT, DELETE",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("GET /user-types/{id}",
		userTypeHandler.HandleUserTypeGetRequest, opts2))
	mux.HandleFunc(middleware.WithCORS("PUT /user-types/{id}",
		userTypeHandler.HandleUserTypePutRequest, opts2))
	mux.HandleFunc(middleware.WithCORS("DELETE /user-types/{id}",
		userTypeHandler.HandleUserTypeDeleteRequest, opts2))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /user-types/{id}",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, opts2))
}
