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

package logout

import (
	"net/http"

	"github.com/thunder-id/thunderid/internal/oauth/oauth2/constants"
	serverconst "github.com/thunder-id/thunderid/internal/system/constants"
	"github.com/thunder-id/thunderid/internal/system/log"
	"github.com/thunder-id/thunderid/internal/system/utils"
)

const handlerLoggerComponentName = "LogoutHandler"

type logoutHandler struct {
	service logoutServiceInterface
	logger  *log.Logger
}

func newLogoutHandler(svc logoutServiceInterface) *logoutHandler {
	return &logoutHandler{
		service: svc,
		logger:  log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName)),
	}
}

// HandleLogout processes RP-Initiated Logout (OIDC Core 1.0 §5 / OIDC Session 1.0 §5).
func (h *logoutHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	idTokenHint := q.Get("id_token_hint")
	postLogoutRedirectURI := q.Get("post_logout_redirect_uri")
	state := q.Get("state")

	if idTokenHint == "" {
		utils.WriteJSONError(w, constants.ErrorInvalidRequest,
			"id_token_hint is required", http.StatusBadRequest, nil)
		return
	}

	redirectURI, svcErr := h.service.Logout(r.Context(), idTokenHint, postLogoutRedirectURI)
	if svcErr != nil {
		h.logger.Error("Logout failed", log.String("error", svcErr.Error()))
		utils.WriteJSONError(w, constants.ErrorInvalidRequest, svcErr.Error(), http.StatusBadRequest, nil)
		return
	}

	if redirectURI == "" {
		w.Header().Set(serverconst.CacheControlHeaderName, serverconst.CacheControlNoStore)
		w.WriteHeader(http.StatusOK)
		return
	}

	target := redirectURI
	if state != "" {
		target += "?state=" + state
	}

	http.Redirect(w, r, target, http.StatusFound)
}
