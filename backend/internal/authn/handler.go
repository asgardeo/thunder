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

package authn

import (
	"encoding/json"
	"net/http"

	"github.com/asgardeo/thunder/internal/idp"
	notifcommon "github.com/asgardeo/thunder/internal/notification/common"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/error/apierror"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

// AuthenticationHandler defines the handler for managing authentication API requests.
type AuthenticationHandler struct {
	authService AuthenticationServiceInterface
}

// NewAuthenticationHandler creates a new instance of AuthenticationHandler.
func NewAuthenticationHandler() *AuthenticationHandler {
	return &AuthenticationHandler{
		authService: NewAuthenticationService(),
	}
}

// HandleSendSMSOTPRequest handles the send SMS OTP authentication request.
func (ah *AuthenticationHandler) HandleSendSMSOTPRequest(w http.ResponseWriter, r *http.Request) {
	otpRequest, err := sysutils.DecodeJSONBody[SendOTPAuthRequestDTO](r)
	if err != nil {
		ah.writeErrorResponse(w, http.StatusBadRequest, APIErrorInvalidRequestFormat)
		return
	}

	sessionToken, svcErr := ah.authService.SendOTP(otpRequest.SenderID, notifcommon.ChannelTypeSMS,
		otpRequest.Recipient)
	if svcErr != nil {
		ah.handleServiceError(w, svcErr)
		return
	}

	response := SendOTPAuthResponseDTO{
		Status:       "SUCCESS",
		SessionToken: sessionToken,
	}
	ah.writeSuccessResponse(w, response)
}

// HandleVerifySMSOTPRequest handles the verify SMS OTP authentication request.
func (ah *AuthenticationHandler) HandleVerifySMSOTPRequest(w http.ResponseWriter, r *http.Request) {
	otpRequest, err := sysutils.DecodeJSONBody[VerifyOTPAuthRequestDTO](r)
	if err != nil {
		ah.writeErrorResponse(w, http.StatusBadRequest, APIErrorInvalidRequestFormat)
		return
	}

	authResponse, svcErr := ah.authService.VerifyOTP(otpRequest.SessionToken, otpRequest.OTP)
	if svcErr != nil {
		ah.handleServiceError(w, svcErr)
		return
	}

	responseDTO := AuthenticationResponseDTO(*authResponse)
	ah.writeSuccessResponse(w, responseDTO)
}

// HandleGoogleAuthStartRequest handles the Google OAuth start authentication request.
func (ah *AuthenticationHandler) HandleGoogleAuthStartRequest(w http.ResponseWriter, r *http.Request) {
	authRequest, err := sysutils.DecodeJSONBody[IDPAuthInitRequestDTO](r)
	if err != nil {
		ah.writeErrorResponse(w, http.StatusBadRequest, APIErrorInvalidRequestFormat)
		return
	}

	authResponse, svcErr := ah.authService.StartIDPAuthentication(idp.IDPTypeGoogle, authRequest.IDPID)
	if svcErr != nil {
		ah.handleServiceError(w, svcErr)
		return
	}

	response := IDPAuthInitResponseDTO(*authResponse)
	ah.writeSuccessResponse(w, response)
}

// HandleGoogleAuthFinishRequest handles the Google OAuth finish authentication request.
func (ah *AuthenticationHandler) HandleGoogleAuthFinishRequest(w http.ResponseWriter, r *http.Request) {
	authRequest, err := sysutils.DecodeJSONBody[IDPAuthFinishRequestDTO](r)
	if err != nil {
		ah.writeErrorResponse(w, http.StatusBadRequest, APIErrorInvalidRequestFormat)
		return
	}

	authResponse, svcErr := ah.authService.FinishIDPAuthentication(idp.IDPTypeGoogle,
		authRequest.SessionToken, authRequest.Code)
	if svcErr != nil {
		ah.handleServiceError(w, svcErr)
		return
	}

	responseDTO := AuthenticationResponseDTO(*authResponse)
	ah.writeSuccessResponse(w, responseDTO)
}

// HandleGithubAuthStartRequest handles the GitHub OAuth start authentication request.
func (ah *AuthenticationHandler) HandleGithubAuthStartRequest(w http.ResponseWriter, r *http.Request) {
	authRequest, err := sysutils.DecodeJSONBody[IDPAuthInitRequestDTO](r)
	if err != nil {
		ah.writeErrorResponse(w, http.StatusBadRequest, APIErrorInvalidRequestFormat)
		return
	}

	authResponse, svcErr := ah.authService.StartIDPAuthentication(idp.IDPTypeGitHub, authRequest.IDPID)
	if svcErr != nil {
		ah.handleServiceError(w, svcErr)
		return
	}

	responseDTO := IDPAuthInitResponseDTO(*authResponse)
	ah.writeSuccessResponse(w, responseDTO)
}

// HandleGithubAuthFinishRequest handles the GitHub OAuth finish authentication request.
func (ah *AuthenticationHandler) HandleGithubAuthFinishRequest(w http.ResponseWriter, r *http.Request) {
	authRequest, err := sysutils.DecodeJSONBody[IDPAuthFinishRequestDTO](r)
	if err != nil {
		ah.writeErrorResponse(w, http.StatusBadRequest, APIErrorInvalidRequestFormat)
		return
	}

	authResponse, svcErr := ah.authService.FinishIDPAuthentication(idp.IDPTypeGitHub,
		authRequest.SessionToken, authRequest.Code)
	if svcErr != nil {
		ah.handleServiceError(w, svcErr)
		return
	}

	responseDTO := AuthenticationResponseDTO(*authResponse)
	ah.writeSuccessResponse(w, responseDTO)
}

// HandleStandardOAuthStartRequest handles the standard OAuth start authentication request.
func (ah *AuthenticationHandler) HandleStandardOAuthStartRequest(w http.ResponseWriter, r *http.Request) {
	authRequest, err := sysutils.DecodeJSONBody[IDPAuthInitRequestDTO](r)
	if err != nil {
		ah.writeErrorResponse(w, http.StatusBadRequest, APIErrorInvalidRequestFormat)
		return
	}

	authResponse, svcErr := ah.authService.StartIDPAuthentication(idp.IDPTypeOAuth, authRequest.IDPID)
	if svcErr != nil {
		ah.handleServiceError(w, svcErr)
		return
	}

	responseDTO := IDPAuthInitResponseDTO(*authResponse)
	ah.writeSuccessResponse(w, responseDTO)
}

// HandleStandardOAuthFinishRequest handles the standard OAuth finish authentication request.
func (ah *AuthenticationHandler) HandleStandardOAuthFinishRequest(w http.ResponseWriter, r *http.Request) {
	authRequest, err := sysutils.DecodeJSONBody[IDPAuthFinishRequestDTO](r)
	if err != nil {
		ah.writeErrorResponse(w, http.StatusBadRequest, APIErrorInvalidRequestFormat)
		return
	}

	authResponse, svcErr := ah.authService.FinishIDPAuthentication(idp.IDPTypeOAuth,
		authRequest.SessionToken, authRequest.Code)
	if svcErr != nil {
		ah.handleServiceError(w, svcErr)
		return
	}

	responseDTO := AuthenticationResponseDTO(*authResponse)
	ah.writeSuccessResponse(w, responseDTO)
}

// handleServiceError converts service errors to appropriate HTTP responses.
func (ah *AuthenticationHandler) handleServiceError(w http.ResponseWriter, svcErr *serviceerror.ServiceError) {
	status := http.StatusBadRequest
	if svcErr.Type == serviceerror.ServerErrorType {
		status = http.StatusInternalServerError
	}

	errorResp := apierror.ErrorResponse{
		Code:        svcErr.Code,
		Message:     svcErr.Error,
		Description: svcErr.ErrorDescription,
	}
	ah.writeErrorResponse(w, status, errorResp)
}

// writeSuccessResponse writes a successful JSON response.
func (ah *AuthenticationHandler) writeSuccessResponse(w http.ResponseWriter, data interface{}) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "AuthenticationHandler"))

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error("Failed to encode response", log.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// writeErrorResponse writes an error response.
func (ah *AuthenticationHandler) writeErrorResponse(w http.ResponseWriter,
	statusCode int, errorResp apierror.ErrorResponse) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "AuthenticationHandler"))

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(errorResp); err != nil {
		logger.Error("Failed to encode error response", log.Error(err))
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
	}
}
