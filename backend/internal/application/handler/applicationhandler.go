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

// Package handler provides HTTP handlers for managing application-related API requests.
package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"github.com/asgardeo/thunder/internal/application/model"
	appprovider "github.com/asgardeo/thunder/internal/application/provider"
	"github.com/asgardeo/thunder/internal/system/log"
)

// ApplicationHandler defines the handler for managing application API requests.
//
// @title          Application Management API
// @version        1.0
// @description    This API is used to manage applications.
//
// @license.name   Apache 2.0
// @license.url    http://www.apache.org/licenses/LICENSE-2.0.html
//
// @host           localhost:8090
// @BasePath       /
type ApplicationHandler struct {
	store map[string]model.Application
	mu    *sync.RWMutex
}

// NewApplicationHandler creates a new instance of ApplicationHandler.
func NewApplicationHandler() *ApplicationHandler {
	return &ApplicationHandler{
		store: make(map[string]model.Application),
		mu:    &sync.RWMutex{},
	}
}

// HandleApplicationPostRequest handles the application request.
//
// @Summary      Create an application
// @Description  Creates a new application with the provided details.
// @Tags         applications
// @Accept       json
// @Produce      json
// @Param        application  body  model.Application  true  "Application data"
// @Success      201  {object}  model.Application
// @Failure      400  {string}  "Bad Request: The request body is malformed or contains invalid data."
// @Failure      500  {string}  "Internal Server Error: An unexpected error occurred while processing the request."
// @Router       /applications [post]
func (ah *ApplicationHandler) HandleApplicationPostRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationHandler"))

	var appInCreationRequest model.Application
	if err := json.NewDecoder(r.Body).Decode(&appInCreationRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Remove hardcoded org ID when multi-tenancy is implemented.
	appInCreationRequest.OrgID = "f5b738ee-8ee4-4kd0-b174-6f8d1eaf32ad"

	// Create the app using the application service.
	appProvider := appprovider.NewApplicationProvider()
	appService := appProvider.GetApplicationService()
	createdApplication, err := appService.CreateApplication(&appInCreationRequest)
	if err != nil {
		http.Error(w, "Failed to create application", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(createdApplication)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	// Log the application creation response.
	logger.Debug(
		"Application POST response sent",
		log.String("app id", createdApplication.ID),
		log.String("org_id", createdApplication.OrgID),
	)
}

// HandleApplicationListRequest handles the application request.
//
// @Summary      List applications
// @Description  Retrieve a list of all applications.
// @Tags         applications
// @Accept       json
// @Produce      json
// @Success      200  {array}   model.Application
// @Failure      400  {string}  "Bad Request: The request body is malformed or contains invalid data."
// @Failure      500  {string}  "Internal Server Error: An unexpected error occurred while processing the request."
// @Router       /applications [get]
func (ah *ApplicationHandler) HandleApplicationListRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationHandler"))

	// Get the application list using the application service.
	appProvider := appprovider.NewApplicationProvider()
	appService := appProvider.GetApplicationService()
	applications, err := appService.GetApplicationList()
	if err != nil {
		http.Error(w, "Failed get application list", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(applications)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	// Log the application response.
	logger.Debug("Application GET (list) response sent")
}

// HandleApplicationGetRequest handles the application request.
//
// @Summary      Get an application by ID
// @Description  Retrieve a specific application using its ID.
// @Tags         applications
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Application ID"
// @Success      200  {object}  model.Application
// @Failure      400  {string}  "Bad Request: The request body is malformed or contains invalid data."
// @Failure      404  {string}  "Not Found: The application with the specified ID does not exist."
// @Failure      500  {string}  "Internal Server Error: An unexpected error occurred while processing the request."
// @Router       /applications/{id} [get]
func (ah *ApplicationHandler) HandleApplicationGetRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationHandler"))

	id := strings.TrimPrefix(r.URL.Path, "/applications/")
	if id == "" {
		http.Error(w, "Missing application id", http.StatusBadRequest)
		return
	}

	// Get the application using the application service.
	appProvider := appprovider.NewApplicationProvider()
	appService := appProvider.GetApplicationService()
	application, err := appService.GetApplication(id)
	if err != nil {
		http.Error(w, "Failed get application", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(application)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	// Log the application response.
	logger.Debug(
		"Application GET response sent",
		log.String("app id", id),
		log.String("org_id", application.OrgID),
	)
}

// HandleApplicationPutRequest handles the application request.
//
// @Summary      Update an application
// @Description  Update the details of an existing application.
// @Tags         applications
// @Accept       json
// @Produce      json
// @Param        id           path   string            true  "Application ID"
// @Param        application  body   model.Application  true  "Updated application data"
// @Success      200  {object}  model.Application
// @Failure      400  {string}  "Bad Request: The request body is malformed or contains invalid data."
// @Failure      404  {string}  "Not Found: The application with the specified ID does not exist."
// @Failure      500  {string}  "Internal Server Error: An unexpected error occurred while processing the request."
// @Router       /applications/{id} [put]
func (ah *ApplicationHandler) HandleApplicationPutRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationHandler"))

	id := strings.TrimPrefix(r.URL.Path, "/applications/")
	if id == "" {
		http.Error(w, "Missing application id", http.StatusBadRequest)
		return
	}

	var updatedApp model.Application
	if err := json.NewDecoder(r.Body).Decode(&updatedApp); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	updatedApp.ID = id
	updatedApp.OrgID = "f5b738ee-8ee4-4kd0-b174-6f8d1eaf32ad"

	// Update the application using the application service.
	appProvider := appprovider.NewApplicationProvider()
	appService := appProvider.GetApplicationService()
	application, err := appService.UpdateApplication(id, &updatedApp)
	if err != nil {
		http.Error(w, "Failed get application", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(application)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	// Log the application response.
	logger.Debug(
		"Application PUT response sent",
		log.String("app id", id),
		log.String("org_id", application.OrgID),
	)
}

// HandleApplicationDeleteRequest handles the application request.
//
// @Summary      Delete an application
// @Description  Delete an application using its ID.
// @Tags         applications
// @Accept       json
// @Produce      json
// @Param        id   path   string  true  "Application ID"
// @Success      204
// @Failure      400  {string}  "Bad Request: The request body is malformed or contains invalid data."
// @Failure      404  {string}  "Not Found: The application with the specified ID does not exist."
// @Failure      500  {string}  "Internal Server Error: An unexpected error occurred while processing the request."
// @Router       /applications/{id} [delete]
func (ah *ApplicationHandler) HandleApplicationDeleteRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationHandler"))

	id := strings.TrimPrefix(r.URL.Path, "/applications/")
	if id == "" {
		http.Error(w, "Missing application id", http.StatusBadRequest)
		return
	}

	// Delete the application using the application service.
	appProvider := appprovider.NewApplicationProvider()
	appService := appProvider.GetApplicationService()

	// Retrieve the application to log the org ID before deletion.
	application, getErr := appService.GetApplication(id)
	if getErr != nil {
		logger.Warn("Failed to retrieve application before deletion", log.Error(getErr))
	}

	err := appService.DeleteApplication(id)
	if err != nil {
		http.Error(w, "Failed delete application", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)

	// Log the application response.
	if getErr == nil {
		logger.Debug(
			"Application DELETE response sent",
			log.String("app id", id),
			log.String("org_id", application.OrgID),
		)
	} else {
		logger.Debug("Application DELETE response sent", log.String("app id", id))
	}
}
