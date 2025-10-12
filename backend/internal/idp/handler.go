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

package idp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/asgardeo/thunder/internal/system/cmodels"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/error/apierror"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

// IDPHandler is the handler for identity provider management operations.
type IDPHandler struct {
	idpService IDPServiceInterface
}

// newIDPHandler creates a new instance of IDPHandler.
func newIDPHandler(idpService IDPServiceInterface) *IDPHandler {
	return &IDPHandler{
		idpService: idpService,
	}
}

// HandleIDPPostRequest handles the create identity provider request.
func (ih *IDPHandler) HandleIDPPostRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "IDPHandler"))

	createRequest, err := sysutils.DecodeJSONBody[idpRequest](r)
	if err != nil {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)

		errResp := apierror.ErrorResponse{
			Code:        ErrorInvalidRequestFormat.Code,
			Message:     ErrorInvalidRequestFormat.Error,
			Description: ErrorInvalidRequestFormat.ErrorDescription,
		}
		if encodeErr := json.NewEncoder(w).Encode(errResp); encodeErr != nil {
			logger.Error("Error encoding error response", log.Error(encodeErr))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	properties, err := getSanitizedProperties(createRequest.Properties)
	if err != nil {
		logger.Error("Failed to sanitize properties", log.Error(err))
		writeServiceErrorResponse(w, &ErrorInternalServerError, logger)
		return
	}

	idpDTO := &IDPDTO{
		Name:        sysutils.SanitizeString(createRequest.Name),
		Description: sysutils.SanitizeString(createRequest.Description),
		Type:        IDPType(sysutils.SanitizeString(createRequest.Type)),
		Properties:  properties,
	}
	createdIDP, svcErr := ih.idpService.CreateIdentityProvider(idpDTO)
	if svcErr != nil {
		writeServiceErrorResponse(w, svcErr, logger)
		return
	}

	idpResponse, err := getIDPResponse(*createdIDP)
	if err != nil {
		logger.Error("Failed to convert IDP to response", log.String("idp", createdIDP.Name), log.Error(err))
		writeServiceErrorResponse(w, &ErrorInternalServerError, logger)
		return
	}

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

	idpListResponse := make([]basicIDPResponse, 0, len(idpList))
	for _, idp := range idpList {
		idpListResponse = append(idpListResponse, basicIDPResponse{
			ID:          idp.ID,
			Name:        idp.Name,
			Description: idp.Description,
			Type:        string(idp.Type),
		})
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
			Code:        ErrorInvalidIDPID.Code,
			Message:     ErrorInvalidIDPID.Error,
			Description: ErrorInvalidIDPID.ErrorDescription,
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

	idpResponse, err := getIDPResponse(*idp)
	if err != nil {
		logger.Error("Failed to convert IDP to response", log.String("idp", idp.Name), log.Error(err))
		writeServiceErrorResponse(w, &ErrorInternalServerError, logger)
		return
	}

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
			Code:        ErrorInvalidIDPID.Code,
			Message:     ErrorInvalidIDPID.Error,
			Description: ErrorInvalidIDPID.ErrorDescription,
		}
		if encodeErr := json.NewEncoder(w).Encode(errResp); encodeErr != nil {
			logger.Error("Error encoding error response", log.Error(encodeErr))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	updateRequest, err := sysutils.DecodeJSONBody[idpRequest](r)
	if err != nil {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)

		errResp := apierror.ErrorResponse{
			Code:        ErrorInvalidRequestFormat.Code,
			Message:     ErrorInvalidRequestFormat.Error,
			Description: ErrorInvalidRequestFormat.ErrorDescription,
		}
		if encodeErr := json.NewEncoder(w).Encode(errResp); encodeErr != nil {
			logger.Error("Error encoding error response", log.Error(encodeErr))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	properties, err := getSanitizedProperties(updateRequest.Properties)
	if err != nil {
		logger.Error("Failed to sanitize properties", log.Error(err))
		writeServiceErrorResponse(w, &ErrorInternalServerError, logger)
		return
	}

	idpDTO := &IDPDTO{
		Name:        sysutils.SanitizeString(updateRequest.Name),
		Description: sysutils.SanitizeString(updateRequest.Description),
		Type:        IDPType(sysutils.SanitizeString(updateRequest.Type)),
		Properties:  properties,
	}
	idpDTO.ID = id

	idp, svcErr := ih.idpService.UpdateIdentityProvider(id, idpDTO)
	if svcErr != nil {
		writeServiceErrorResponse(w, svcErr, logger)
		return
	}

	idpResponse, err := getIDPResponse(*idp)
	if err != nil {
		logger.Error("Failed to convert IDP to response", log.String("idp", idp.Name), log.Error(err))
		writeServiceErrorResponse(w, &ErrorInternalServerError, logger)
		return
	}

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
			Code:        ErrorInvalidIDPID.Code,
			Message:     ErrorInvalidIDPID.Error,
			Description: ErrorInvalidIDPID.ErrorDescription,
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
	case ErrorIDPNotFound.Code:
		return http.StatusNotFound
	case ErrorIDPAlreadyExists.Code:
		return http.StatusConflict
	default:
		return http.StatusBadRequest
	}
}

// getSanitizedProperties sanitizes a slice of PropertyDTOs and converts them to Properties.
func getSanitizedProperties(propDTOs []cmodels.PropertyDTO) ([]cmodels.Property, error) {
	properties := make([]cmodels.Property, 0, len(propDTOs))
	for _, propDTO := range propDTOs {
		sanitizedDTO := cmodels.PropertyDTO{
			Name:     sysutils.SanitizeString(propDTO.Name),
			Value:    sysutils.SanitizeString(propDTO.Value),
			IsSecret: propDTO.IsSecret,
		}
		property, err := sanitizedDTO.ToProperty()
		if err != nil {
			return nil, err
		}
		properties = append(properties, *property)
	}
	return properties, nil
}

// getIDPResponse constructs the response for a identity provider.
func getIDPResponse(idp IDPDTO) (idpResponse, error) {
	returnIDP := idpResponse{
		ID:          idp.ID,
		Name:        idp.Name,
		Description: idp.Description,
		Type:        string(idp.Type),
	}

	// Convert Property to PropertyDTO and mask secret properties in the response.
	idpProperties := make([]cmodels.PropertyDTO, 0, len(idp.Properties))
	for _, property := range idp.Properties {
		if property.IsSecret() {
			maskedProperty := &cmodels.PropertyDTO{
				Name:     property.GetName(),
				Value:    "******",
				IsSecret: property.IsSecret(),
			}
			idpProperties = append(idpProperties, *maskedProperty)
		} else {
			propertyDTO, err := property.ToPropertyDTO()
			if err != nil {
				return idpResponse{}, fmt.Errorf("failed to convert property %s: %w", property.GetName(), err)
			}
			idpProperties = append(idpProperties, *propertyDTO)
		}
	}
	returnIDP.Properties = idpProperties

	return returnIDP, nil
}
