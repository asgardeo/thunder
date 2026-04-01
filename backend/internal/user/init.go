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

package user

import (
	"net/http"
	"strings"

	"github.com/asgardeo/thunder/internal/entity"
	oupkg "github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/system/crypto/hash"
	declarativeresource "github.com/asgardeo/thunder/internal/system/declarative_resource"
	"github.com/asgardeo/thunder/internal/system/middleware"
	"github.com/asgardeo/thunder/internal/system/sysauthz"
	"github.com/asgardeo/thunder/internal/userschema"
)

// Initialize initializes the user service and registers its routes.
func Initialize(
	mux *http.ServeMux,
	entityService entity.EntityServiceInterface,
	ouService oupkg.OrganizationUnitServiceInterface,
	userSchemaService userschema.UserSchemaServiceInterface,
	hashService hash.HashServiceInterface,
	authzService sysauthz.SystemAuthorizationServiceInterface,
) (UserServiceInterface, oupkg.OUUserResolver, declarativeresource.ResourceExporter, error) {
	// Step 1: Create service with entity service
	userService := newUserService(authzService, entityService, ouService, userSchemaService, hashService)

	// Step 2: Load declarative resources via entity service
	if err := entityService.LoadDeclarativeResources(makeUserDeclarativeConfig()); err != nil {
		return nil, nil, nil, err
	}

	setUserService(userService) // Set the provider for backward compatibility

	userHandler := newUserHandler(userService)
	registerRoutes(mux, userHandler)

	// Create resolver for OU package to query user data without cross-DB access
	ouUserResolver := newOUUserResolver(entityService, userSchemaService)

	// Create and return exporter
	exporter := newUserExporter(userService)
	return userService, ouUserResolver, exporter, nil
}

// registerRoutes registers the routes for user management operations.
func registerRoutes(mux *http.ServeMux, userHandler *userHandler) {
	opts1 := middleware.CORSOptions{
		AllowedMethods:   "GET, POST",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("POST /users", userHandler.HandleUserPostRequest, opts1))
	mux.HandleFunc(middleware.WithCORS("GET /users", userHandler.HandleUserListRequest, opts1))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /users", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}, opts1))

	opts2 := middleware.CORSOptions{
		AllowedMethods:   "GET, PUT, DELETE",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("GET /users/",
		func(w http.ResponseWriter, r *http.Request) {
			path := strings.TrimPrefix(r.URL.Path, "/users/")
			segments := strings.Split(path, "/")
			r.SetPathValue("id", segments[0])

			if len(segments) == 1 {
				userHandler.HandleUserGetRequest(w, r)
			} else if len(segments) == 2 && segments[1] == "groups" {
				userHandler.HandleUserGroupsGetRequest(w, r)
			} else {
				http.NotFound(w, r)
			}
		}, opts2))
	mux.HandleFunc(middleware.WithCORS("PUT /users/", userHandler.HandleUserPutRequest, opts2))
	mux.HandleFunc(middleware.WithCORS("DELETE /users/", userHandler.HandleUserDeleteRequest, opts2))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /users/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}, opts2))

	optsSelf := middleware.CORSOptions{
		AllowedMethods:   "GET, PUT",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("GET /users/me", userHandler.HandleSelfUserGetRequest, optsSelf))
	mux.HandleFunc(middleware.WithCORS("PUT /users/me", userHandler.HandleSelfUserPutRequest, optsSelf))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /users/me", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}, optsSelf))

	optsSelfCredentials := middleware.CORSOptions{
		AllowedMethods:   "POST",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("POST /users/me/update-credentials",
		userHandler.HandleSelfUserCredentialUpdateRequest, optsSelfCredentials))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /users/me/update-credentials",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, optsSelfCredentials))

	opts3 := middleware.CORSOptions{
		AllowedMethods:   "GET, POST",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("GET /users/tree/{path...}",
		userHandler.HandleUserListByPathRequest, opts3))
	mux.HandleFunc(middleware.WithCORS("POST /users/tree/{path...}",
		userHandler.HandleUserPostByPathRequest, opts3))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /users/tree/{path...}",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, opts3))
}
