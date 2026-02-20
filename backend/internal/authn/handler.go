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
	"net/http"

	"github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/authn/credentials"
	"github.com/asgardeo/thunder/internal/authn/otp"
	"github.com/asgardeo/thunder/internal/idp"
	notifcommon "github.com/asgardeo/thunder/internal/notification/common"
	"github.com/asgardeo/thunder/internal/system/error/apierror"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

// authenticationHandler defines the handler for managing authentication API requests.
type authenticationHandler struct {
	authService AuthenticationServiceInterface
}

// newAuthenticationHandler creates a new instance of AuthenticationHandler.
func newAuthenticationHandler(authnService AuthenticationServiceInterface) *authenticationHandler {
	return &authenticationHandler{
		authService: authnService,
	}
}

// HandleCredentialsAuthRequest handles the credentials authentication request.
func (ah *authenticationHandler) HandleCredentialsAuthRequest(w http.ResponseWriter, r *http.Request) {
	authRequestPtr, err := sysutils.DecodeJSONBody[AuthenticateWithCredentialsRequestDTO](r)
	if err != nil {
		sysutils.WriteErrorResponse(w, http.StatusBadRequest, common.APIErrorInvalidRequestFormat)
		return
	}
	authRequest := *authRequestPtr

	skipAssertion := false
	if authRequest.SkipAssertion != nil {
		skipAssertion = *authRequest.SkipAssertion
	}

	assertion := ""
	if authRequest.Assertion != nil {
		assertion = *authRequest.Assertion
	}

	authResponse, svcErr := ah.authService.AuthenticateWithCredentials(
		authRequest.Identifiers, authRequest.Credentials, skipAssertion, assertion)
	if svcErr != nil {
		ah.handleServiceError(w, svcErr)
		return
	}

	responseDTO := AuthenticationResponseDTO(*authResponse)
	sysutils.WriteSuccessResponse(w, http.StatusOK, responseDTO)
}

// HandleSendSMSOTPRequest handles the send SMS OTP authentication request.
func (ah *authenticationHandler) HandleSendSMSOTPRequest(w http.ResponseWriter, r *http.Request) {
	otpRequest, err := sysutils.DecodeJSONBody[SendOTPAuthRequestDTO](r)
	if err != nil {
		sysutils.WriteErrorResponse(w, http.StatusBadRequest, common.APIErrorInvalidRequestFormat)
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
	sysutils.WriteSuccessResponse(w, http.StatusOK, response)
}

// HandleVerifySMSOTPRequest handles the verify SMS OTP authentication request.
func (ah *authenticationHandler) HandleVerifySMSOTPRequest(w http.ResponseWriter, r *http.Request) {
	otpRequest, err := sysutils.DecodeJSONBody[VerifyOTPAuthRequestDTO](r)
	if err != nil {
		sysutils.WriteErrorResponse(w, http.StatusBadRequest, common.APIErrorInvalidRequestFormat)
		return
	}

	authResponse, svcErr := ah.authService.VerifyOTP(otpRequest.SessionToken, otpRequest.SkipAssertion,
		otpRequest.Assertion, otpRequest.OTP)
	if svcErr != nil {
		ah.handleServiceError(w, svcErr)
		return
	}

	responseDTO := AuthenticationResponseDTO(*authResponse)
	sysutils.WriteSuccessResponse(w, http.StatusOK, responseDTO)
}

// HandleGoogleAuthStartRequest handles the Google OAuth start authentication request.
func (ah *authenticationHandler) HandleGoogleAuthStartRequest(w http.ResponseWriter, r *http.Request) {
	authRequest, err := sysutils.DecodeJSONBody[IDPAuthInitRequestDTO](r)
	if err != nil {
		sysutils.WriteErrorResponse(w, http.StatusBadRequest, common.APIErrorInvalidRequestFormat)
		return
	}

	authResponse, svcErr := ah.authService.StartIDPAuthentication(idp.IDPTypeGoogle, authRequest.IDPID)
	if svcErr != nil {
		ah.handleServiceError(w, svcErr)
		return
	}

	response := IDPAuthInitResponseDTO(*authResponse)
	sysutils.WriteSuccessResponse(w, http.StatusOK, response)
}

// HandleGoogleAuthFinishRequest handles the Google OAuth finish authentication request.
func (ah *authenticationHandler) HandleGoogleAuthFinishRequest(w http.ResponseWriter, r *http.Request) {
	authRequest, err := sysutils.DecodeJSONBody[IDPAuthFinishRequestDTO](r)
	if err != nil {
		sysutils.WriteErrorResponse(w, http.StatusBadRequest, common.APIErrorInvalidRequestFormat)
		return
	}

	authResponse, svcErr := ah.authService.FinishIDPAuthentication(idp.IDPTypeGoogle, authRequest.SessionToken,
		authRequest.SkipAssertion, authRequest.Assertion, authRequest.Code)
	if svcErr != nil {
		ah.handleServiceError(w, svcErr)
		return
	}

	responseDTO := AuthenticationResponseDTO(*authResponse)
	sysutils.WriteSuccessResponse(w, http.StatusOK, responseDTO)
}

// HandleGithubAuthStartRequest handles the GitHub OAuth start authentication request.
func (ah *authenticationHandler) HandleGithubAuthStartRequest(w http.ResponseWriter, r *http.Request) {
	authRequest, err := sysutils.DecodeJSONBody[IDPAuthInitRequestDTO](r)
	if err != nil {
		sysutils.WriteErrorResponse(w, http.StatusBadRequest, common.APIErrorInvalidRequestFormat)
		return
	}

	authResponse, svcErr := ah.authService.StartIDPAuthentication(idp.IDPTypeGitHub, authRequest.IDPID)
	if svcErr != nil {
		ah.handleServiceError(w, svcErr)
		return
	}

	responseDTO := IDPAuthInitResponseDTO(*authResponse)
	sysutils.WriteSuccessResponse(w, http.StatusOK, responseDTO)
}

// HandleGithubAuthFinishRequest handles the GitHub OAuth finish authentication request.
func (ah *authenticationHandler) HandleGithubAuthFinishRequest(w http.ResponseWriter, r *http.Request) {
	authRequest, err := sysutils.DecodeJSONBody[IDPAuthFinishRequestDTO](r)
	if err != nil {
		sysutils.WriteErrorResponse(w, http.StatusBadRequest, common.APIErrorInvalidRequestFormat)
		return
	}

	authResponse, svcErr := ah.authService.FinishIDPAuthentication(idp.IDPTypeGitHub, authRequest.SessionToken,
		authRequest.SkipAssertion, authRequest.Assertion, authRequest.Code)
	if svcErr != nil {
		ah.handleServiceError(w, svcErr)
		return
	}

	responseDTO := AuthenticationResponseDTO(*authResponse)
	sysutils.WriteSuccessResponse(w, http.StatusOK, responseDTO)
}

// HandleStandardOAuthStartRequest handles the standard OAuth start authentication request.
func (ah *authenticationHandler) HandleStandardOAuthStartRequest(w http.ResponseWriter, r *http.Request) {
	authRequest, err := sysutils.DecodeJSONBody[IDPAuthInitRequestDTO](r)
	if err != nil {
		sysutils.WriteErrorResponse(w, http.StatusBadRequest, common.APIErrorInvalidRequestFormat)
		return
	}

	authResponse, svcErr := ah.authService.StartIDPAuthentication(idp.IDPTypeOAuth, authRequest.IDPID)
	if svcErr != nil {
		ah.handleServiceError(w, svcErr)
		return
	}

	responseDTO := IDPAuthInitResponseDTO(*authResponse)
	sysutils.WriteSuccessResponse(w, http.StatusOK, responseDTO)
}

// HandleStandardOAuthFinishRequest handles the standard OAuth finish authentication request.
func (ah *authenticationHandler) HandleStandardOAuthFinishRequest(w http.ResponseWriter, r *http.Request) {
	authRequest, err := sysutils.DecodeJSONBody[IDPAuthFinishRequestDTO](r)
	if err != nil {
		sysutils.WriteErrorResponse(w, http.StatusBadRequest, common.APIErrorInvalidRequestFormat)
		return
	}

	authResponse, svcErr := ah.authService.FinishIDPAuthentication(idp.IDPTypeOAuth, authRequest.SessionToken,
		authRequest.SkipAssertion, authRequest.Assertion, authRequest.Code)
	if svcErr != nil {
		ah.handleServiceError(w, svcErr)
		return
	}

	responseDTO := AuthenticationResponseDTO(*authResponse)
	sysutils.WriteSuccessResponse(w, http.StatusOK, responseDTO)
}

// HandlePasskeyRegisterStartRequest handles the passkey start registration request.
func (ah *authenticationHandler) HandlePasskeyRegisterStartRequest(w http.ResponseWriter, r *http.Request) {
	regRequest, err := sysutils.DecodeJSONBody[PasskeyRegisterStartRequestDTO](r)
	if err != nil {
		sysutils.WriteErrorResponse(w, http.StatusBadRequest, common.APIErrorInvalidRequestFormat)
		return
	}

	regResponse, svcErr := ah.authService.StartPasskeyRegistration(
		regRequest.UserID,
		regRequest.RelyingPartyID,
		regRequest.RelyingPartyName,
		regRequest.AuthenticatorSelection,
		regRequest.Attestation,
	)
	if svcErr != nil {
		ah.handleServiceError(w, svcErr)
		return
	}

	sysutils.WriteSuccessResponse(w, http.StatusOK, regResponse)
}

// HandlePasskeyRegisterFinishRequest handles the passkey finish registration request.
func (ah *authenticationHandler) HandlePasskeyRegisterFinishRequest(w http.ResponseWriter, r *http.Request) {
	regRequest, err := sysutils.DecodeJSONBody[PasskeyRegisterFinishRequestDTO](r)
	if err != nil {
		sysutils.WriteErrorResponse(w, http.StatusBadRequest, common.APIErrorInvalidRequestFormat)
		return
	}

	regResponse, svcErr := ah.authService.FinishPasskeyRegistration(
		regRequest.PublicKeyCredential,
		regRequest.SessionToken,
		regRequest.CredentialName,
	)
	if svcErr != nil {
		ah.handleServiceError(w, svcErr)
		return
	}

	sysutils.WriteSuccessResponse(w, http.StatusOK, regResponse)
}

// HandlePasskeyStartRequest handles the passkey start authentication request.
func (ah *authenticationHandler) HandlePasskeyStartRequest(w http.ResponseWriter, r *http.Request) {
	authRequest, err := sysutils.DecodeJSONBody[PasskeyStartRequestDTO](r)
	if err != nil {
		sysutils.WriteErrorResponse(w, http.StatusBadRequest, common.APIErrorInvalidRequestFormat)
		return
	}

	authResponse, svcErr := ah.authService.StartPasskeyAuthentication(
		authRequest.UserID,
		authRequest.RelyingPartyID,
	)
	if svcErr != nil {
		ah.handleServiceError(w, svcErr)
		return
	}

	sysutils.WriteSuccessResponse(w, http.StatusOK, authResponse)
}

// HandlePasskeyFinishRequest handles the passkey finish authentication request.
func (ah *authenticationHandler) HandlePasskeyFinishRequest(w http.ResponseWriter, r *http.Request) {
	authRequest, err := sysutils.DecodeJSONBody[PasskeyFinishRequestDTO](r)
	if err != nil {
		sysutils.WriteErrorResponse(w, http.StatusBadRequest, common.APIErrorInvalidRequestFormat)
		return
	}

	authResponse, svcErr := ah.authService.FinishPasskeyAuthentication(
		authRequest.PublicKeyCredential.ID,
		authRequest.PublicKeyCredential.Type,
		authRequest.PublicKeyCredential.Response,
		authRequest.SessionToken,
		authRequest.SkipAssertion,
		authRequest.Assertion,
	)
	if svcErr != nil {
		ah.handleServiceError(w, svcErr)
		return
	}

	responseDTO := AuthenticationResponseDTO(*authResponse)
	sysutils.WriteSuccessResponse(w, http.StatusOK, responseDTO)
}

// handleServiceError converts service errors to appropriate HTTP responses.
func (ah *authenticationHandler) handleServiceError(w http.ResponseWriter, svcErr *serviceerror.ServiceError) {
	status := http.StatusInternalServerError
	if svcErr.Type == serviceerror.ClientErrorType {
		switch svcErr.Code {
		case credentials.ErrorInvalidCredentials.Code, otp.ErrorIncorrectOTP.Code:
			status = http.StatusUnauthorized
		case common.ErrorUserNotFound.Code:
			status = http.StatusNotFound
		default:
			status = http.StatusBadRequest
		}
	}

	errorResp := apierror.ErrorResponse{
		Code:        svcErr.Code,
		Message:     svcErr.Error,
		Description: svcErr.ErrorDescription,
	}
	sysutils.WriteErrorResponse(w, status, errorResp)
}
