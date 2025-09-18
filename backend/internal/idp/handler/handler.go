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

// Package handler provides the implementation for identity provider management operations.
package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/asgardeo/thunder/internal/idp/constants"
	"github.com/asgardeo/thunder/internal/idp/model"
	"github.com/asgardeo/thunder/internal/idp/service"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/error/apierror"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

// IDPHandler is the handler for identity provider management operations.
type IDPHandler struct {
	idpService service.IDPServiceInterface
}

// NewIDPHandler creates a new instance of IDPHandler.
func NewIDPHandler() *IDPHandler {
	return &IDPHandler{
		idpService: service.NewIDPService(),
	}
}

// HandleIDPPostRequest handles the create identity provider request.
func (ih *IDPHandler) HandleIDPPostRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "IDPHandler"))

	createRequest, err := sysutils.DecodeJSONBody[model.IdpRequest](r)
	if err != nil {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)

		errResp := apierror.ErrorResponse{
			Code:        constants.ErrorInvalidRequestFormat.Code,
			Message:     constants.ErrorInvalidRequestFormat.Error,
			Description: constants.ErrorInvalidRequestFormat.ErrorDescription,
		}
		if encodeErr := json.NewEncoder(w).Encode(errResp); encodeErr != nil {
			logger.Error("Error encoding error response", log.Error(encodeErr))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	idpDTO := &model.IdpDTO{
		Name:        sysutils.SanitizeString(createRequest.Name),
		Description: sysutils.SanitizeString(createRequest.Description),
		Properties:  getSanitizedProperties(createRequest.Properties),
	}
	createdIDP, svcErr := ih.idpService.CreateIdentityProvider(idpDTO)
	if svcErr != nil {
		writeServiceErrorResponse(w, svcErr, logger)
		return
	}

	idpResponse := getIDPResponse(*createdIDP)

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusCreated)

	if encodeErr := json.NewEncoder(w).Encode(idpResponse); encodeErr != nil {
		logger.Error("Error encoding response", log.Error(encodeErr))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// HandleIDPListRequest handles the list identity providers request.
func (ih *IDPHandler) HandleIDPListRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "IDPHandler"))

	idpList, svcErr := ih.idpService.GetIdentityProviderList()
	if svcErr != nil {
		writeServiceErrorResponse(w, svcErr, logger)
		return
	}

	idpListResponse := make([]model.BasicIdpResponse, 0, len(idpList))
	for _, idp := range idpList {
		idpListResponse = append(idpListResponse, model.BasicIdpResponse(idp))
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	if encodeErr := json.NewEncoder(w).Encode(idpListResponse); encodeErr != nil {
		logger.Error("Error encoding response", log.Error(encodeErr))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// HandleIDPGetRequest handles the get identity provider request.
func (ih *IDPHandler) HandleIDPGetRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "IDPHandler"))

	id := r.PathValue("id")
	if strings.TrimSpace(id) == "" {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)

		errResp := apierror.ErrorResponse{
			Code:        constants.ErrorInvalidIDPID.Code,
			Message:     constants.ErrorInvalidIDPID.Error,
			Description: constants.ErrorInvalidIDPID.ErrorDescription,
		}
		if encodeErr := json.NewEncoder(w).Encode(errResp); encodeErr != nil {
			logger.Error("Error encoding error response", log.Error(encodeErr))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	idp, svcErr := ih.idpService.GetIdentityProvider(id)
	if svcErr != nil {
		writeServiceErrorResponse(w, svcErr, logger)
		return
	}

	idpResponse := getIDPResponse(*idp)

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	if encodeErr := json.NewEncoder(w).Encode(idpResponse); encodeErr != nil {
		logger.Error("Error encoding response", log.Error(encodeErr))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// HandleIDPPutRequest handles the update identity provider request.
func (ih *IDPHandler) HandleIDPPutRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "IDPHandler"))

	id := r.PathValue("id")
	if strings.TrimSpace(id) == "" {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)

		errResp := apierror.ErrorResponse{
			Code:        constants.ErrorInvalidIDPID.Code,
			Message:     constants.ErrorInvalidIDPID.Error,
			Description: constants.ErrorInvalidIDPID.ErrorDescription,
		}
		if encodeErr := json.NewEncoder(w).Encode(errResp); encodeErr != nil {
			logger.Error("Error encoding error response", log.Error(encodeErr))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	updateRequest, err := sysutils.DecodeJSONBody[model.IdpRequest](r)
	if err != nil {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)

		errResp := apierror.ErrorResponse{
			Code:        constants.ErrorInvalidRequestFormat.Code,
			Message:     constants.ErrorInvalidRequestFormat.Error,
			Description: constants.ErrorInvalidRequestFormat.ErrorDescription,
		}
		if encodeErr := json.NewEncoder(w).Encode(errResp); encodeErr != nil {
			logger.Error("Error encoding error response", log.Error(encodeErr))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	idpDTO := &model.IdpDTO{
		Name:        sysutils.SanitizeString(updateRequest.Name),
		Description: sysutils.SanitizeString(updateRequest.Description),
		Properties:  getSanitizedProperties(updateRequest.Properties),
	}
	idpDTO.ID = id

	idp, svcErr := ih.idpService.UpdateIdentityProvider(id, idpDTO)
	if svcErr != nil {
		writeServiceErrorResponse(w, svcErr, logger)
		return
	}

	idpResponse := getIDPResponse(*idp)

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	if encodeErr := json.NewEncoder(w).Encode(idpResponse); encodeErr != nil {
		logger.Error("Error encoding response", log.Error(encodeErr))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// HandleIDPDeleteRequest handles the delete identity provider request.
func (ih *IDPHandler) HandleIDPDeleteRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "IDPHandler"))

	id := r.PathValue("id")
	if strings.TrimSpace(id) == "" {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)

		errResp := apierror.ErrorResponse{
			Code:        constants.ErrorInvalidIDPID.Code,
			Message:     constants.ErrorInvalidIDPID.Error,
			Description: constants.ErrorInvalidIDPID.ErrorDescription,
		}
		if encodeErr := json.NewEncoder(w).Encode(errResp); encodeErr != nil {
			logger.Error("Error encoding error response", log.Error(encodeErr))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	svcErr := ih.idpService.DeleteIdentityProvider(id)
	if svcErr != nil {
		writeServiceErrorResponse(w, svcErr, logger)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// writeServiceErrorResponse writes the appropriate HTTP error response based on the service error.
func writeServiceErrorResponse(w http.ResponseWriter, svcErr *serviceerror.ServiceError, logger *log.Logger) {
	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)

	var statusCode int
	if svcErr.Type == serviceerror.ClientErrorType {
		statusCode = getClientErrorStatusCode(svcErr.Code)
	} else {
		statusCode = http.StatusInternalServerError
	}
	w.WriteHeader(statusCode)

	errResp := apierror.ErrorResponse{
		Code:        svcErr.Code,
		Message:     svcErr.Error,
		Description: svcErr.ErrorDescription,
	}

	if encodeErr := json.NewEncoder(w).Encode(errResp); encodeErr != nil {
		logger.Error("Error encoding error response", log.Error(encodeErr))
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
	}
}

// getClientErrorStatusCode returns the appropriate HTTP status code for client errors.
func getClientErrorStatusCode(errorCode string) int {
	switch errorCode {
	case constants.ErrorIDPNotFound.Code:
		return http.StatusNotFound
	case constants.ErrorIDPAlreadyExists.Code:
		return http.StatusConflict
	default:
		return http.StatusBadRequest
	}
}

// getSanitizedProperties sanitizes the properties of an identity provider.
func getSanitizedProperties(properties []model.IdpProperty) []model.IdpProperty {
	sanitizedProperties := make([]model.IdpProperty, 0, len(properties))
	for _, property := range properties {
		sanitizedProperties = append(sanitizedProperties, model.IdpProperty{
			Name:     sysutils.SanitizeString(property.Name),
			Value:    sysutils.SanitizeString(property.Value),
			IsSecret: property.IsSecret,
		})
	}
	return sanitizedProperties
}

// getIDPResponse constructs the response for a identity provider.
func getIDPResponse(idp model.IdpDTO) model.IdpResponse {
	returnIDP := model.IdpResponse{
		ID:          idp.ID,
		Name:        idp.Name,
		Description: idp.Description,
	}

	// Mask secret properties in the response.
	idpProperties := make([]model.IdpProperty, 0, len(idp.Properties))
	for _, property := range idp.Properties {
		if property.IsSecret {
			idpProperties = append(idpProperties, model.IdpProperty{
				Name:     property.Name,
				Value:    "******",
				IsSecret: property.IsSecret,
			})
		} else {
			idpProperties = append(idpProperties, property)
		}
	}
	returnIDP.Properties = idpProperties

	return returnIDP
}
