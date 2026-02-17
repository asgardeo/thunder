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

package userinfo

import (
	"net/http"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
)

const handlerLoggerComponentName = "UserInfoHandler"

// userInfoHandler handles OIDC UserInfo requests.
type userInfoHandler struct {
	service userInfoServiceInterface
	logger  *log.Logger
}

// newUserInfoHandler creates a new userInfo handler.
func newUserInfoHandler(userInfoService userInfoServiceInterface) *userInfoHandler {
	return &userInfoHandler{
		service: userInfoService,
		logger:  log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName)),
	}
}

// HandleUserInfo handles UserInfo requests.
func (h *userInfoHandler) HandleUserInfo(w http.ResponseWriter, r *http.Request) {
	// Extract access token from Authorization header
	authHeader := r.Header.Get(serverconst.AuthorizationHeaderName)
	accessToken, err := utils.ExtractBearerToken(authHeader)
	if err != nil {
		utils.WriteJSONError(w, constants.ErrorInvalidRequest,
			err.Error(), http.StatusUnauthorized, nil)
		return
	}

	result, svcErr := h.service.GetUserInfo(accessToken)
	if svcErr != nil {
		h.writeServiceErrorResponse(w, svcErr)
		return
	}

	w.Header().Set(serverconst.CacheControlHeaderName, serverconst.CacheControlNoStore)
	w.Header().Set(serverconst.PragmaHeaderName, serverconst.PragmaNoCache)

	if result.Type == appmodel.UserInfoResponseTypeJWS {
		w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJWT)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(result.JWTBody))
	} else {
		utils.WriteSuccessResponse(w, http.StatusOK, result.JSONBody)
	}

	h.logger.Debug("UserInfo response sent successfully")
}

// writeServiceErrorResponse writes a service error response.
func (h *userInfoHandler) writeServiceErrorResponse(w http.ResponseWriter, svcErr *serviceerror.ServiceError) {
	var statusCode int

	switch svcErr.Type {
	case serviceerror.ClientErrorType:
		if svcErr == &errorInsufficientScope {
			statusCode = http.StatusForbidden
		} else {
			statusCode = http.StatusUnauthorized
		}
	case serviceerror.ServerErrorType:
		statusCode = http.StatusInternalServerError
	default:
		statusCode = http.StatusUnauthorized
	}

	utils.WriteJSONError(w, svcErr.Code, svcErr.ErrorDescription, statusCode, nil)
}
