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

package group

import (
	"net/http"
	"strings"

	oupkg "github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/system/middleware"
	"github.com/asgardeo/thunder/internal/user"
)

// Initialize initializes the group service and registers its routes.
func Initialize(
	mux *http.ServeMux,
	ouService oupkg.OrganizationUnitServiceInterface,
	userService user.UserServiceInterface,
) GroupServiceInterface {
	groupService := newGroupService(ouService, userService)
	groupHandler := newGroupHandler(groupService)
	registerRoutes(mux, groupHandler)
	return groupService
}

// registerRoutes registers the routes for group management operations.
func registerRoutes(mux *http.ServeMux, groupHandler *groupHandler) {
	opts1 := middleware.CORSOptions{
		AllowedMethods:   "GET, POST",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("POST /groups", groupHandler.HandleGroupPostRequest, opts1))
	mux.HandleFunc(middleware.WithCORS("GET /groups", groupHandler.HandleGroupListRequest, opts1))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /groups", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}, opts1))

	opts2 := middleware.CORSOptions{
		AllowedMethods:   "GET, PUT, DELETE",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	// Special handling for /groups/{id} and /groups/{id}/members
	mux.HandleFunc(middleware.WithCORS("GET /groups/",
		func(w http.ResponseWriter, r *http.Request) {
			path := strings.TrimPrefix(r.URL.Path, "/groups/")
			segments := strings.Split(path, "/")
			r.SetPathValue("id", segments[0])

			if len(segments) == 1 {
				groupHandler.HandleGroupGetRequest(w, r)
			} else if len(segments) == 2 && segments[1] == "members" {
				groupHandler.HandleGroupMembersGetRequest(w, r)
			} else {
				http.NotFound(w, r)
			}
		}, opts2))
	mux.HandleFunc(middleware.WithCORS("PUT /groups/{id}", groupHandler.HandleGroupPutRequest, opts2))
	mux.HandleFunc(middleware.WithCORS("DELETE /groups/{id}", groupHandler.HandleGroupDeleteRequest, opts2))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /groups/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}, opts2))

	opts3 := middleware.CORSOptions{
		AllowedMethods:   "GET, POST",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("GET /groups/tree/{path...}",
		groupHandler.HandleGroupListByPathRequest, opts3))
	mux.HandleFunc(middleware.WithCORS("POST /groups/tree/{path...}",
		groupHandler.HandleGroupPostByPathRequest, opts3))
	mux.HandleFunc(middleware.WithCORS("OPTIONS /groups/tree/{path...}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}, opts3))
}
