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

	"github.com/asgardeo/thunder/internal/application/constants"
	"github.com/asgardeo/thunder/internal/application/model"
	appprovider "github.com/asgardeo/thunder/internal/application/provider"
	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
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
	ApplicationProvider appprovider.ApplicationProviderInterface
}

// NewApplicationHandler creates a new instance of ApplicationHandler.
func NewApplicationHandler() *ApplicationHandler {
	return &ApplicationHandler{
		ApplicationProvider: appprovider.NewApplicationProvider(),
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

	var appRequest model.ApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&appRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Need to refactor when supporting other/multiple inbound auth types.
	inboundAuthConfig := model.InboundAuthConfig{
		Type: constants.OAuthInboundAuthType,
		OAuthAppConfig: &model.OAuthAppConfig{
			ClientID:                appRequest.ClientID,
			ClientSecret:            appRequest.ClientSecret,
			RedirectURIs:            appRequest.RedirectURIs,
			GrantTypes:              appRequest.GrantTypes,
			ResponseTypes:           appRequest.ResponseTypes,
			TokenEndpointAuthMethod: appRequest.TokenEndpointAuthMethod,
		},
	}
	appDTO := model.ApplicationDTO{
		Name:                      appRequest.Name,
		Description:               appRequest.Description,
		AuthFlowGraphID:           appRequest.AuthFlowGraphID,
		RegistrationFlowGraphID:   appRequest.RegistrationFlowGraphID,
		IsRegistrationFlowEnabled: appRequest.IsRegistrationFlowEnabled,
		URL:                       appRequest.URL,
		LogoURL:                   appRequest.LogoURL,
		Certificate:               appRequest.Certificate,
		InboundAuthConfig:         []model.InboundAuthConfig{inboundAuthConfig},
	}

	// Create the app using the application service.
	appService := ah.ApplicationProvider.GetApplicationService()
	createdAppDTO, err := appService.CreateApplication(&appDTO)
	if err != nil {
		http.Error(w, "Failed to create application", http.StatusInternalServerError)
		return
	}

	// TODO: Need to refactor when supporting other/multiple inbound auth types.
	if len(createdAppDTO.InboundAuthConfig) == 0 ||
		createdAppDTO.InboundAuthConfig[0].Type != constants.OAuthInboundAuthType {
		logger.Error("Unsupported inbound authentication type returned",
			log.String("type", string(createdAppDTO.InboundAuthConfig[0].Type)))
		http.Error(w, "Unsupported inbound authentication type", http.StatusInternalServerError)
		return
	}
	returnInboundAuthConfig := createdAppDTO.InboundAuthConfig[0]
	if returnInboundAuthConfig.OAuthAppConfig == nil {
		logger.Error("OAuth application configuration is nil")
		http.Error(w, "Something went wrong while creating the application", http.StatusInternalServerError)
		return
	}

	redirectURIs := returnInboundAuthConfig.OAuthAppConfig.RedirectURIs
	if len(redirectURIs) == 0 {
		redirectURIs = []string{}
	}
	grantTypes := returnInboundAuthConfig.OAuthAppConfig.GrantTypes
	if len(grantTypes) == 0 {
		grantTypes = []oauth2const.GrantType{}
	}
	responseTypes := returnInboundAuthConfig.OAuthAppConfig.ResponseTypes
	if len(responseTypes) == 0 {
		responseTypes = []oauth2const.ResponseType{}
	}
	tokenAuthMethods := returnInboundAuthConfig.OAuthAppConfig.TokenEndpointAuthMethod
	if len(tokenAuthMethods) == 0 {
		tokenAuthMethods = []oauth2const.TokenEndpointAuthMethod{}
	}

	returnApp := model.ApplicationCompleteResponse{
		ID:                        createdAppDTO.ID,
		Name:                      createdAppDTO.Name,
		Description:               createdAppDTO.Description,
		ClientID:                  returnInboundAuthConfig.OAuthAppConfig.ClientID,
		ClientSecret:              returnInboundAuthConfig.OAuthAppConfig.ClientSecret,
		RedirectURIs:              redirectURIs,
		GrantTypes:                grantTypes,
		ResponseTypes:             responseTypes,
		TokenEndpointAuthMethod:   tokenAuthMethods,
		AuthFlowGraphID:           createdAppDTO.AuthFlowGraphID,
		RegistrationFlowGraphID:   createdAppDTO.RegistrationFlowGraphID,
		IsRegistrationFlowEnabled: createdAppDTO.IsRegistrationFlowEnabled,
		URL:                       createdAppDTO.URL,
		LogoURL:                   createdAppDTO.LogoURL,
		Certificate:               createdAppDTO.Certificate,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(returnApp)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Debug("Application POST response sent", log.String("appId", createdAppDTO.ID))
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

	appService := ah.ApplicationProvider.GetApplicationService()
	applications, err := appService.GetApplicationList()
	if err != nil {
		http.Error(w, "Failed get application list", http.StatusInternalServerError)
		return
	}

	returnAppList := make([]model.BasicApplicationResponse, len(applications))
	for i, app := range applications {
		returnAppList[i] = model.BasicApplicationResponse{
			ID:                        app.ID,
			Name:                      app.Name,
			Description:               app.Description,
			ClientID:                  app.ClientID,
			AuthFlowGraphID:           app.AuthFlowGraphID,
			RegistrationFlowGraphID:   app.RegistrationFlowGraphID,
			IsRegistrationFlowEnabled: app.IsRegistrationFlowEnabled,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(returnAppList)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

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

	appService := ah.ApplicationProvider.GetApplicationService()
	appDTO, err := appService.GetApplication(id)
	if err != nil {
		http.Error(w, "Failed get application", http.StatusInternalServerError)
		return
	}

	// TODO: Need to refactor when supporting other/multiple inbound auth types.
	if len(appDTO.InboundAuthConfig) == 0 || appDTO.InboundAuthConfig[0].Type != constants.OAuthInboundAuthType {
		logger.Error("Unsupported inbound authentication type returned",
			log.String("type", string(appDTO.InboundAuthConfig[0].Type)))
		http.Error(w, "Unsupported inbound authentication type", http.StatusInternalServerError)
		return
	}
	returnInboundAuthConfig := appDTO.InboundAuthConfig[0]
	if returnInboundAuthConfig.OAuthAppConfig == nil {
		logger.Error("OAuth application configuration is nil")
		http.Error(w, "Something went wrong while retrieving the application", http.StatusInternalServerError)
		return
	}

	redirectURIs := returnInboundAuthConfig.OAuthAppConfig.RedirectURIs
	if len(redirectURIs) == 0 {
		redirectURIs = []string{}
	}
	grantTypes := returnInboundAuthConfig.OAuthAppConfig.GrantTypes
	if len(grantTypes) == 0 {
		grantTypes = []oauth2const.GrantType{}
	}
	responseTypes := returnInboundAuthConfig.OAuthAppConfig.ResponseTypes
	if len(responseTypes) == 0 {
		responseTypes = []oauth2const.ResponseType{}
	}
	tokenAuthMethods := returnInboundAuthConfig.OAuthAppConfig.TokenEndpointAuthMethod
	if len(tokenAuthMethods) == 0 {
		tokenAuthMethods = []oauth2const.TokenEndpointAuthMethod{}
	}

	returnApp := model.ApplicationGetResponse{
		ID:                        appDTO.ID,
		Name:                      appDTO.Name,
		Description:               appDTO.Description,
		ClientID:                  returnInboundAuthConfig.OAuthAppConfig.ClientID,
		RedirectURIs:              redirectURIs,
		GrantTypes:                grantTypes,
		ResponseTypes:             responseTypes,
		TokenEndpointAuthMethod:   tokenAuthMethods,
		AuthFlowGraphID:           appDTO.AuthFlowGraphID,
		RegistrationFlowGraphID:   appDTO.RegistrationFlowGraphID,
		IsRegistrationFlowEnabled: appDTO.IsRegistrationFlowEnabled,
		URL:                       appDTO.URL,
		LogoURL:                   appDTO.LogoURL,
		Certificate:               appDTO.Certificate,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(returnApp)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Debug("Application GET response sent", log.String("appId", id))
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

	var appRequest model.ApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&appRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Need to refactor when supporting other/multiple inbound auth types.
	inboundAuthConfig := model.InboundAuthConfig{
		Type: constants.OAuthInboundAuthType,
		OAuthAppConfig: &model.OAuthAppConfig{
			ClientID:                appRequest.ClientID,
			ClientSecret:            appRequest.ClientSecret,
			RedirectURIs:            appRequest.RedirectURIs,
			GrantTypes:              appRequest.GrantTypes,
			ResponseTypes:           appRequest.ResponseTypes,
			TokenEndpointAuthMethod: appRequest.TokenEndpointAuthMethod,
		},
	}
	updateReqAppDTO := model.ApplicationDTO{
		ID:                        id,
		Name:                      appRequest.Name,
		Description:               appRequest.Description,
		AuthFlowGraphID:           appRequest.AuthFlowGraphID,
		RegistrationFlowGraphID:   appRequest.RegistrationFlowGraphID,
		IsRegistrationFlowEnabled: appRequest.IsRegistrationFlowEnabled,
		URL:                       appRequest.URL,
		LogoURL:                   appRequest.LogoURL,
		Certificate:               appRequest.Certificate,
		InboundAuthConfig:         []model.InboundAuthConfig{inboundAuthConfig},
	}

	// Update the application using the application service.
	appService := ah.ApplicationProvider.GetApplicationService()
	updatedAppDTO, err := appService.UpdateApplication(id, &updateReqAppDTO)
	if err != nil {
		logger.Error("Failed to update application", log.Error(err))
		http.Error(w, "Failed get application", http.StatusInternalServerError)
		return
	}

	// TODO: Need to refactor when supporting other/multiple inbound auth types.
	if len(updatedAppDTO.InboundAuthConfig) == 0 ||
		updatedAppDTO.InboundAuthConfig[0].Type != constants.OAuthInboundAuthType {
		logger.Error("Unsupported inbound authentication type returned",
			log.String("type", string(updatedAppDTO.InboundAuthConfig[0].Type)))
		http.Error(w, "Unsupported inbound authentication type", http.StatusInternalServerError)
		return
	}
	returnInboundAuthConfig := updatedAppDTO.InboundAuthConfig[0]
	if returnInboundAuthConfig.OAuthAppConfig == nil {
		logger.Error("OAuth application configuration is nil")
		http.Error(w, "Something went wrong while updating the application", http.StatusInternalServerError)
		return
	}

	redirectURIs := returnInboundAuthConfig.OAuthAppConfig.RedirectURIs
	if len(redirectURIs) == 0 {
		redirectURIs = []string{}
	}
	grantTypes := returnInboundAuthConfig.OAuthAppConfig.GrantTypes
	if len(grantTypes) == 0 {
		grantTypes = []oauth2const.GrantType{}
	}
	responseTypes := returnInboundAuthConfig.OAuthAppConfig.ResponseTypes
	if len(responseTypes) == 0 {
		responseTypes = []oauth2const.ResponseType{}
	}
	tokenAuthMethods := returnInboundAuthConfig.OAuthAppConfig.TokenEndpointAuthMethod
	if len(tokenAuthMethods) == 0 {
		tokenAuthMethods = []oauth2const.TokenEndpointAuthMethod{}
	}

	returnApp := model.ApplicationCompleteResponse{
		ID:                        updatedAppDTO.ID,
		Name:                      updatedAppDTO.Name,
		Description:               updatedAppDTO.Description,
		ClientID:                  returnInboundAuthConfig.OAuthAppConfig.ClientID,
		ClientSecret:              returnInboundAuthConfig.OAuthAppConfig.ClientSecret,
		RedirectURIs:              redirectURIs,
		GrantTypes:                grantTypes,
		ResponseTypes:             responseTypes,
		TokenEndpointAuthMethod:   tokenAuthMethods,
		AuthFlowGraphID:           updatedAppDTO.AuthFlowGraphID,
		RegistrationFlowGraphID:   updatedAppDTO.RegistrationFlowGraphID,
		IsRegistrationFlowEnabled: updatedAppDTO.IsRegistrationFlowEnabled,
		URL:                       updatedAppDTO.URL,
		LogoURL:                   updatedAppDTO.LogoURL,
		Certificate:               updatedAppDTO.Certificate,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(returnApp)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Debug("Application PUT response sent", log.String("appId", id))
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
	appService := ah.ApplicationProvider.GetApplicationService()
	err := appService.DeleteApplication(id)
	if err != nil {
		http.Error(w, "Failed delete application", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)

	// Log the application response.
	logger.Debug("Application DELETE response sent", log.String("appId", id))
}
