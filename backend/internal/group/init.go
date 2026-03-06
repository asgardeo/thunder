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
	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/middleware"
	"github.com/asgardeo/thunder/internal/user"
)

// Initialize initializes the group service and registers its routes.
func Initialize(
	mux *http.ServeMux,
	ouService oupkg.OrganizationUnitServiceInterface,
	userService user.UserServiceInterface,
) (GroupServiceInterface, error) {
	// Get transactioner from DB provider
	transactioner, err := provider.GetDBProvider().GetUserDBTransactioner()
	if err != nil {
		return nil, err
	}

	groupService := newGroupService(ouService, userService, transactioner)
	groupHandler := newGroupHandler(groupService)
	registerRoutes(mux, groupHandler)
	return groupService, nil
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
		AllowedMethods:   "GET, POST, PUT, DELETE",
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
	// Handle OPTIONS preflight for /groups/{id} and /groups/{id}/members using the same
	// catch-all pattern as the GET handler above, to avoid conflicts with /groups/tree/{path...}.
	mux.HandleFunc(middleware.WithCORS("OPTIONS /groups/",
		func(w http.ResponseWriter, r *http.Request) {
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

	// POST routes for /groups/{id}/members/add and /groups/{id}/members/remove.
	// These use a catch-all pattern to avoid route conflicts with /groups/tree/{path...}.
	opts4 := middleware.CORSOptions{
		AllowedMethods:   "POST",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}
	mux.HandleFunc(middleware.WithCORS("POST /groups/",
		func(w http.ResponseWriter, r *http.Request) {
			path := strings.TrimPrefix(r.URL.Path, "/groups/")
			segments := strings.Split(path, "/")

			// Match /groups/{id}/members/add and /groups/{id}/members/remove
			if len(segments) == 3 && segments[0] != "" && segments[1] == "members" {
				r.SetPathValue("id", segments[0])
				switch segments[2] {
				case "add":
					groupHandler.HandleGroupMembersAddRequest(w, r)
				case "remove":
					groupHandler.HandleGroupMembersRemoveRequest(w, r)
				default:
					http.NotFound(w, r)
				}
			} else {
				http.NotFound(w, r)
			}
		}, opts4))
}
