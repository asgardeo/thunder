/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
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

// Package handler provides the implementation for user management operations.
package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/user/model"
	userprovider "github.com/asgardeo/thunder/internal/user/provider"
)

// UserHandler is the handler for user management operations.
type UserHandler struct {
}

// HandleUserPostRequest handles the user request.
func (ah *UserHandler) HandleUserPostRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserHandler"))

	var userInCreationRequest model.User
	if err := json.NewDecoder(r.Body).Decode(&userInCreationRequest); err != nil {
		http.Error(w, "Bad Request: The request body is malformed or contains invalid data.", http.StatusBadRequest)
		return
	}

	// Create the user using the user service.
	userProvider := userprovider.NewUserProvider()
	userService := userProvider.GetUserService()
	createdUser, err := userService.CreateUser(&userInCreationRequest)
	if err != nil {
		if errors.Is(err, model.ErrBadAttributesInRequest) {
			http.Error(w, "Bad Request: The attributes element is malformed or contains invalid data.", http.StatusBadRequest)
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(createdUser)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Log the user creation response.
	logger.Debug("User POST response sent", log.String("user id", createdUser.ID))
}

// HandleUserListRequest handles the user request.
func (ah *UserHandler) HandleUserListRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserHandler"))

	// Get the user list using the user service.
	userProvider := userprovider.NewUserProvider()
	userService := userProvider.GetUserService()
	users, err := userService.GetUserList()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(users)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Log the user response.
	logger.Debug("User GET (list) response sent")
}

// HandleUserGetRequest handles the user request.
func (ah *UserHandler) HandleUserGetRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserHandler"))

	id := strings.TrimPrefix(r.URL.Path, "/users/")
	if id == "" {
		http.Error(w, "Bad Request: Missing user id.", http.StatusBadRequest)
		return
	}

	// Get the user using the user service.
	userProvider := userprovider.NewUserProvider()
	userService := userProvider.GetUserService()
	user, err := userService.GetUser(id)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			http.Error(w, "Not Found: The user with the specified id does not exist.", http.StatusNotFound)
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Log the user response.
	logger.Debug("User GET response sent", log.String("user id", id))
}

// HandleUserPutRequest handles the user request.
func (ah *UserHandler) HandleUserPutRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserHandler"))

	id := strings.TrimPrefix(r.URL.Path, "/users/")
	if id == "" {
		http.Error(w, "Bad Request: Missing user id.", http.StatusBadRequest)
		return
	}

	var updatedUser model.User
	if err := json.NewDecoder(r.Body).Decode(&updatedUser); err != nil {
		http.Error(w, "Bad Request: The request body is malformed or contains invalid data.", http.StatusBadRequest)
		return
	}
	updatedUser.ID = id

	// Update the user using the user service.
	userProvider := userprovider.NewUserProvider()
	userService := userProvider.GetUserService()
	user, err := userService.UpdateUser(id, &updatedUser)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			http.Error(w, "Not Found: The user with the specified id does not exist.", http.StatusNotFound)
		} else if errors.Is(err, model.ErrBadAttributesInRequest) {
			http.Error(w, "Bad Request: The attributes element is malformed or contains invalid data.", http.StatusBadRequest)
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Log the user response.
	logger.Debug("User PUT response sent", log.String("user id", id))
}

// HandleUserDeleteRequest handles the delete user request.
func (ah *UserHandler) HandleUserDeleteRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserHandler"))

	id := strings.TrimPrefix(r.URL.Path, "/users/")
	if id == "" {
		http.Error(w, "Bad Request: Missing user id.", http.StatusBadRequest)
		return
	}

	// Delete the user using the user service.
	userProvider := userprovider.NewUserProvider()
	userService := userProvider.GetUserService()
	err := userService.DeleteUser(id)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)

	// Log the user response.
	logger.Debug("User DELETE response sent", log.String("user id", id))
}
