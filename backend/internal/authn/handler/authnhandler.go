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

// Package handler provides the HTTP handler for authentication requests.
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/asgardeo/thunder/internal/authn"
	"github.com/asgardeo/thunder/internal/authn/constants"
	"github.com/asgardeo/thunder/internal/authn/model"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/error/apierror"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	systemutils "github.com/asgardeo/thunder/internal/system/utils"
)

// AuthenticationHandlerInterface defines the interface for handling authentication requests.
type AuthenticationHandlerInterface interface {
	HandleAuthenticationRequest(w http.ResponseWriter, r *http.Request)
}

// AuthenticationHandler implements the AuthenticationHandlerInterface to handle authentication requests.
type AuthenticationHandler struct {
	AuthNService authn.AuthenticationServiceInterface
}

// NewAuthenticationHandler creates a new instance of AuthenticationHandler.
func NewAuthenticationHandler() AuthenticationHandlerInterface {
	return &AuthenticationHandler{
		AuthNService: authn.NewAuthenticationService(),
	}
}

// HandleAuthenticationRequest handles the authentication request received.
func (ah *AuthenticationHandler) HandleAuthenticationRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "AuthenticationHandler"))
	logger.Debug("Handling authentication request")

	authR, err := systemutils.DecodeJSONBody[model.AuthNRequest](r)
	if err != nil {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(constants.APIErrorJSONDecodeError); err != nil {
			logger.Error("Error encoding error response", log.Error(err))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
			return
		}
	}

	// Sanitize the inputs
	sessionDataKey := systemutils.SanitizeString(authR.SessionDataKey)
	flowID := systemutils.SanitizeString(authR.FlowID)
	actionID := systemutils.SanitizeString(authR.ActionID)
	inputs := systemutils.SanitizeStringMap(authR.Inputs)

	authResp, svcErr := ah.AuthNService.Execute(sessionDataKey, flowID, actionID, inputs)
	if svcErr != nil {
		handleServiceError(w, logger, svcErr)
		return
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(authResp)
	if err != nil {
		logger.Error("Error encoding response", log.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// handleServiceError handles service errors by writing an appropriate error response to the HTTP response writer.
func handleServiceError(w http.ResponseWriter, logger *log.Logger, svcErr *serviceerror.ServiceError) {
	errResp := apierror.ErrorResponse{
		Code:        svcErr.Code,
		Message:     svcErr.Error,
		Description: svcErr.ErrorDescription,
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	if svcErr.Type == serviceerror.ClientErrorType {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		errResp.Message = "Internal Server Error"
		errResp.Description = "An unexpected error occurred while executing the authentication flow"
		w.WriteHeader(http.StatusInternalServerError)
	}

	if err := json.NewEncoder(w).Encode(errResp); err != nil {
		logger.Error("Error encoding error response", log.Error(err))
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
	}
}
