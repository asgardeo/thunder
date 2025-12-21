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

package invitation

import (
	"net/http"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/middleware"
)

var (
	invitationService InvitationServiceInterface
)

// Initialize initializes the invitation package and registers its routes.
func Initialize(mux *http.ServeMux) error {
	store := newInvitationStore()
	runtime := config.GetThunderRuntime()

	invitationService = NewInvitationService(store, &runtime.Config.Server)
	handler := newInvitationHandler(invitationService)

	registerRoutes(mux, handler)

	return nil
}

// registerRoutes registers the routes for invitation operations.
func registerRoutes(mux *http.ServeMux, handler *invitationHandler) {
	// Invitation CRUD operations
	invitationOpts := middleware.CORSOptions{
		AllowedMethods:   "GET, POST, DELETE",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}

	// POST /invitations - Create a new invitation
	mux.HandleFunc(middleware.WithCORS("POST /invitations", handler.HandleCreateInvitation, invitationOpts))

	// GET /invitations/{id} or GET /invitations/{token}/validate
	mux.HandleFunc(middleware.WithCORS("GET /invitations/", handler.HandleGetInvitation, invitationOpts))

	// DELETE /invitations/{id}
	mux.HandleFunc(middleware.WithCORS("DELETE /invitations/", handler.HandleDeleteInvitation, invitationOpts))

	// OPTIONS /invitations
	mux.HandleFunc(middleware.WithCORS("OPTIONS /invitations", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}, invitationOpts))

	mux.HandleFunc(middleware.WithCORS("OPTIONS /invitations/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}, invitationOpts))

	// User-specific invitation endpoint
	userInviteOpts := middleware.CORSOptions{
		AllowedMethods:   "POST",
		AllowedHeaders:   "Content-Type, Authorization",
		AllowCredentials: true,
	}

	// POST /users/{id}/invite - Create invitation for a specific user
	mux.HandleFunc(middleware.WithCORS("POST /users/", func(w http.ResponseWriter, r *http.Request) {
		// Check if this is an invite request by checking the path
		if len(r.URL.Path) > 7 { // "/users/" = 7 chars
			path := r.URL.Path[7:] // Remove "/users/" prefix
			if len(path) > 0 && path[len(path)-6:] == "invite" {
				handler.HandleCreateInvitationForUser(w, r)
				return
			}
		}
		// If not an invite request, return 404
		http.NotFound(w, r)
	}, userInviteOpts))
}

// GetInvitationService returns the invitation service instance.
func GetInvitationService() InvitationServiceInterface {
	return invitationService
}
