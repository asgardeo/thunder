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

package introspect

import (
	"encoding/json"
	"net/http"

	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/model"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

// tokenIntrospectionHandler handles OAuth 2.0 token introspection requests.
type tokenIntrospectionHandler struct {
	service TokenIntrospectionServiceInterface
}

// newTokenIntrospectionHandler creates a new token introspection handler (internal use).
func newTokenIntrospectionHandler(introspectionService TokenIntrospectionServiceInterface) *tokenIntrospectionHandler {
	return &tokenIntrospectionHandler{
		service: introspectionService,
	}
}

// HandleIntrospect handles token introspection requests
func (h *tokenIntrospectionHandler) HandleIntrospect(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "TokenIntrospectionHandler"))

	if err := r.ParseForm(); err != nil {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)
		errResp := model.ErrorResponse{
			Error:            constants.ErrorInvalidRequest,
			ErrorDescription: "Failed to decode request body",
		}
		if err := json.NewEncoder(w).Encode(errResp); err != nil {
			logger.Error("Error encoding error response", log.Error(err))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	// Extract request parameters
	token := r.FormValue(constants.RequestParamToken)
	if token == "" {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)
		errResp := model.ErrorResponse{
			Error:            constants.ErrorInvalidRequest,
			ErrorDescription: "Token parameter is required",
		}
		if err := json.NewEncoder(w).Encode(errResp); err != nil {
			logger.Error("Error encoding error response", log.Error(err))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}
	// token_type_hint parameter is not supported due to non persistent tokens in the server
	tokenTypeHint := r.FormValue(constants.RequestParamTokenTypeHint)

	response, err := h.service.IntrospectToken(r.Context(), token, tokenTypeHint)
	if err != nil {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
		w.WriteHeader(http.StatusInternalServerError)
		errResp := model.ErrorResponse{
			Error:            constants.ErrorServerError,
			ErrorDescription: "Server error while introspecting token",
		}
		if err := json.NewEncoder(w).Encode(errResp); err != nil {
			logger.Error("Error encoding error response", log.Error(err))
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	sysutils.WriteSuccessResponse(w, http.StatusOK, response)
}
