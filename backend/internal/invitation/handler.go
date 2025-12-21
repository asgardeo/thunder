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
	"encoding/json"
	"net/http"
	"strings"

	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
)

const handlerComponentName = "InvitationHandler"

// invitationHandler handles HTTP requests for invitation operations.
type invitationHandler struct {
	service InvitationServiceInterface
}

// newInvitationHandler creates a new invitationHandler.
func newInvitationHandler(service InvitationServiceInterface) *invitationHandler {
	return &invitationHandler{
		service: service,
	}
}

// HandleCreateInvitation handles POST /invitations
func (h *invitationHandler) HandleCreateInvitation(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerComponentName))

	var request CreateInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		logger.Debug("Failed to decode request body", log.Error(err))
		handleError(w, &serviceerror.ServiceError{
			Code:             "INV-40000",
			Error:            "invalid_request",
			ErrorDescription: "Invalid request body",
			Type:             serviceerror.ClientErrorType,
		})
		return
	}

	response, svcErr := h.service.CreateInvitation(request)
	if svcErr != nil {
		handleError(w, svcErr)
		return
	}

	writeJSONResponse(w, http.StatusCreated, response)
}

// HandleCreateInvitationForUser handles POST /users/{id}/invite
func (h *invitationHandler) HandleCreateInvitationForUser(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerComponentName))

	// Extract user ID from path
	path := strings.TrimPrefix(r.URL.Path, "/users/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "invite" {
		handleError(w, &ErrorMissingUserID)
		return
	}
	userID := parts[0]

	if userID == "" {
		handleError(w, &ErrorMissingUserID)
		return
	}

	// Parse optional request body for application ID and expiration
	var request CreateInvitationRequest
	if r.Body != nil && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			logger.Debug("Failed to decode request body", log.Error(err))
			handleError(w, &serviceerror.ServiceError{
				Code:             "INV-40000",
				Error:            "invalid_request",
				ErrorDescription: "Invalid request body",
				Type:             serviceerror.ClientErrorType,
			})
			return
		}
	}

	request.UserID = userID

	response, svcErr := h.service.CreateInvitation(request)
	if svcErr != nil {
		handleError(w, svcErr)
		return
	}

	writeJSONResponse(w, http.StatusCreated, response)
}

// HandleValidateToken handles GET /invitations/{token}/validate
func (h *invitationHandler) HandleValidateToken(w http.ResponseWriter, r *http.Request) {
	// Extract token from path
	path := strings.TrimPrefix(r.URL.Path, "/invitations/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "validate" {
		handleError(w, &ErrorInvalidToken)
		return
	}
	token := parts[0]

	if token == "" {
		handleError(w, &ErrorInvalidToken)
		return
	}

	response, svcErr := h.service.ValidateToken(token)
	if svcErr != nil {
		// For validation errors, still return the response with valid=false
		if response != nil {
			writeJSONResponse(w, http.StatusOK, response)
			return
		}
		handleError(w, svcErr)
		return
	}

	writeJSONResponse(w, http.StatusOK, response)
}

// HandleDeleteInvitation handles DELETE /invitations/{invitationId}
func (h *invitationHandler) HandleDeleteInvitation(w http.ResponseWriter, r *http.Request) {
	// Extract invitation ID from path
	invitationID := strings.TrimPrefix(r.URL.Path, "/invitations/")
	if invitationID == "" {
		handleError(w, &ErrorInvitationNotFound)
		return
	}

	svcErr := h.service.DeleteInvitation(invitationID)
	if svcErr != nil {
		handleError(w, svcErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleGetInvitation handles GET /invitations/{invitationId}
func (h *invitationHandler) HandleGetInvitation(w http.ResponseWriter, r *http.Request) {
	// Extract invitation ID from path
	path := strings.TrimPrefix(r.URL.Path, "/invitations/")
	// Check if it's a validate request
	if strings.Contains(path, "/validate") {
		h.HandleValidateToken(w, r)
		return
	}

	invitationID := path
	if invitationID == "" {
		handleError(w, &ErrorInvitationNotFound)
		return
	}

	invitation, svcErr := h.service.GetInvitation(invitationID)
	if svcErr != nil {
		handleError(w, svcErr)
		return
	}

	writeJSONResponse(w, http.StatusOK, invitation)
}

// writeJSONResponse writes a JSON response with the given status code.
func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerComponentName))
		logger.Error("Failed to encode response", log.Error(err))
	}
}

// handleError writes an error response based on the service error.
func handleError(w http.ResponseWriter, svcErr *serviceerror.ServiceError) {
	statusCode := http.StatusInternalServerError
	if svcErr.Type == serviceerror.ClientErrorType {
		statusCode = http.StatusBadRequest
		// Check for specific error codes
		if svcErr.Code == ErrorInvitationNotFound.Code {
			statusCode = http.StatusNotFound
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]string{
		"code":        svcErr.Code,
		"message":     svcErr.Error,
		"description": svcErr.ErrorDescription,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerComponentName))
		logger.Error("Failed to encode error response", log.Error(err))
	}
}
