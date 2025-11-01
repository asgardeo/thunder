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

package role

import (
	"net/http"
	"strings"

	"github.com/asgardeo/thunder/internal/group"
	oupkg "github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/system/middleware"
	"github.com/asgardeo/thunder/internal/user"
)

// Initialize initializes the role service and registers its routes.
func Initialize(
	mux *http.ServeMux,
	userService user.UserServiceInterface,
	groupService group.GroupServiceInterface,
	ouService oupkg.OrganizationUnitServiceInterface,
) RoleServiceInterface {
	roleStore := newRoleStore()
	roleService := newRoleService(roleStore, userService, groupService, ouService)
	roleHandler := newRoleHandler(roleService)
	registerRoutes(mux, roleHandler)
	return roleService
}

// registerRoutes registers the routes for role management operations.
func registerRoutes(mux *http.ServeMux, roleHandler *roleHandler) {
	opts1 := middleware.CORSOptions{
		AllowedMethods:   "GET, POST",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("POST /roles", roleHandler.HandleRolePostRequest, opts1))
	mux.HandleFunc(middleware.WithCORS("GET /roles", roleHandler.HandleRoleListRequest, opts1))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /roles", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}, opts1))

	opts2 := middleware.CORSOptions{
		AllowedMethods:   "GET, PUT, DELETE",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	// Special handling for /roles/{id} and /roles/{id}/assignments
	mux.HandleFunc(middleware.WithCORS("GET /roles/",
		func(w http.ResponseWriter, r *http.Request) {
			path := strings.TrimPrefix(r.URL.Path, "/roles/")
			segments := strings.Split(path, "/")
			r.SetPathValue("id", segments[0])

			if len(segments) == 1 {
				roleHandler.HandleRoleGetRequest(w, r)
			} else if len(segments) == 2 && segments[1] == "assignments" {
				roleHandler.HandleRoleAssignmentsGetRequest(w, r)
			} else {
				http.NotFound(w, r)
			}
		}, opts2))
	mux.HandleFunc(middleware.WithCORS("PUT /roles/{id}", roleHandler.HandleRolePutRequest, opts2))
	mux.HandleFunc(middleware.WithCORS("DELETE /roles/{id}", roleHandler.HandleRoleDeleteRequest, opts2))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /roles/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}, opts2))

	opts3 := middleware.CORSOptions{
		AllowedMethods:   "POST",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("POST /roles/{id}/assignments/add",
		roleHandler.HandleRoleAddAssignmentsRequest, opts3))
	mux.HandleFunc(middleware.WithCORS("POST /roles/{id}/assignments/remove",
		roleHandler.HandleRoleRemoveAssignmentsRequest, opts3))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /roles/{id}/assignments/add",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, opts3))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /roles/{id}/assignments/remove",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, opts3))
}
