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
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/authn/credentials"
	"github.com/asgardeo/thunder/internal/authn/otp"
	"github.com/asgardeo/thunder/internal/idp"
	"github.com/asgardeo/thunder/internal/system/error/apierror"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
)

type AuthenticationHandlerTestSuite struct {
	suite.Suite
	mockService *AuthenticationServiceInterfaceMock
	handler     *AuthenticationHandler
}

func TestAuthenticationHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(AuthenticationHandlerTestSuite))
}

func (suite *AuthenticationHandlerTestSuite) SetupTest() {
	suite.mockService = NewAuthenticationServiceInterfaceMock(suite.T())
	suite.handler = &AuthenticationHandler{
		authService: suite.mockService,
	}
}

func (suite *AuthenticationHandlerTestSuite) testIDPAuthFinishSuccess(
	idpType idp.IDPType, endpoint string, handlerFunc func(http.ResponseWriter, *http.Request)) {
	authRequest := IDPAuthFinishRequestDTO{
		SessionToken:  testSessionTkn,
		SkipAssertion: false,
		Code:          "auth_code_123",
	}
	authResponse := &common.AuthenticationResponse{
		ID:               "user123",
		Type:             "person",
		OrganizationUnit: "test-ou",
		Assertion:        "jwt-token",
	}

	suite.mockService.On("FinishIDPAuthentication", idpType, authRequest.SessionToken,
		authRequest.SkipAssertion, authRequest.Assertion, authRequest.Code).Return(authResponse, nil)

	body, _ := json.Marshal(authRequest)
	req := httptest.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	w := httptest.NewRecorder()

	handlerFunc(w, req)

	suite.Equal(http.StatusOK, w.Code)
	var response AuthenticationResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(authResponse.ID, response.ID)
	suite.Equal(authResponse.Assertion, response.Assertion)
}

func (suite *AuthenticationHandlerTestSuite) TestHandleCredentialsAuthRequestSuccess() {
	authRequest := map[string]interface{}{
		"username": "testuser",
		"password": "testpass",
	}
	authResponse := &common.AuthenticationResponse{
		ID:               "user123",
		Type:             "person",
		OrganizationUnit: "test-ou",
		Assertion:        "jwt-token",
	}

	suite.mockService.On("AuthenticateWithCredentials", authRequest, false, "").Return(authResponse, nil)

	body, _ := json.Marshal(authRequest)
	req := httptest.NewRequest(http.MethodPost, "/authenticate/credentials", bytes.NewReader(body))
	w := httptest.NewRecorder()

	suite.handler.HandleCredentialsAuthRequest(w, req)

	suite.Equal(http.StatusOK, w.Code)
	var response AuthenticationResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(authResponse.ID, response.ID)
	suite.Equal(authResponse.Assertion, response.Assertion)
}

func (suite *AuthenticationHandlerTestSuite) TestHandleCredentialsAuthRequestWithSkipAssertion() {
	authRequest := map[string]interface{}{
		"username":       "testuser",
		"password":       "testpass",
		"skip_assertion": true,
	}
	expectedRequest := map[string]interface{}{
		"username": "testuser",
		"password": "testpass",
	}
	authResponse := &common.AuthenticationResponse{
		ID:               "user123",
		Type:             "person",
		OrganizationUnit: "test-ou",
	}

	suite.mockService.On("AuthenticateWithCredentials", expectedRequest, true, "").Return(authResponse, nil)

	body, _ := json.Marshal(authRequest)
	req := httptest.NewRequest(http.MethodPost, "/authenticate/credentials", bytes.NewReader(body))
	w := httptest.NewRecorder()

	suite.handler.HandleCredentialsAuthRequest(w, req)

	suite.Equal(http.StatusOK, w.Code)
	var response AuthenticationResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(authResponse.ID, response.ID)
	suite.Empty(response.Assertion)
}

func (suite *AuthenticationHandlerTestSuite) TestHandleCredentialsAuthRequestWithExistingAssertion() {
	existingAssertion := "existing.jwt.token"
	authRequest := map[string]interface{}{
		"username":  "testuser",
		"password":  "testpass",
		"assertion": existingAssertion,
	}
	expectedRequest := map[string]interface{}{
		"username": "testuser",
		"password": "testpass",
	}
	authResponse := &common.AuthenticationResponse{
		ID:               "user123",
		Type:             "person",
		OrganizationUnit: "test-ou",
		Assertion:        "updated.jwt.token",
	}

	suite.mockService.On("AuthenticateWithCredentials", expectedRequest, false, existingAssertion).
		Return(authResponse, nil)

	body, _ := json.Marshal(authRequest)
	req := httptest.NewRequest(http.MethodPost, "/authenticate/credentials", bytes.NewReader(body))
	w := httptest.NewRecorder()

	suite.handler.HandleCredentialsAuthRequest(w, req)

	suite.Equal(http.StatusOK, w.Code)
	var response AuthenticationResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(authResponse.ID, response.ID)
	suite.Equal(authResponse.Assertion, response.Assertion)
	suite.Equal("updated.jwt.token", response.Assertion)
}

func (suite *AuthenticationHandlerTestSuite) TestHandleCredentialsAuthRequestInvalidJSON() {
	req := httptest.NewRequest(http.MethodPost, "/authenticate/credentials", bytes.NewReader([]byte("invalid-json")))
	w := httptest.NewRecorder()

	suite.handler.HandleCredentialsAuthRequest(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
	var errResp apierror.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	suite.NoError(err)
	suite.Equal(common.APIErrorInvalidRequestFormat.Code, errResp.Code)
}

func (suite *AuthenticationHandlerTestSuite) TestHandleCredentialsAuthRequestServiceError() {
	cases := []struct {
		name               string
		authRequest        map[string]interface{}
		serviceError       *serviceerror.ServiceError
		expectedStatusCode int
		expectedErrorCode  string
	}{
		{
			name: "InvalidCredentials",
			authRequest: map[string]interface{}{
				"username": "testuser",
				"password": "wrongpass",
			},
			serviceError:       &credentials.ErrorInvalidCredentials,
			expectedStatusCode: http.StatusUnauthorized,
			expectedErrorCode:  credentials.ErrorInvalidCredentials.Code,
		},
		{
			name: "UserNotFound",
			authRequest: map[string]interface{}{
				"username": "nonexistent",
				"password": "testpass",
			},
			serviceError:       &common.ErrorUserNotFound,
			expectedStatusCode: http.StatusNotFound,
			expectedErrorCode:  common.ErrorUserNotFound.Code,
		},
		{
			name: "ClientError",
			authRequest: map[string]interface{}{
				"username": "testuser",
				"password": "testpass",
			},
			serviceError: &serviceerror.ServiceError{
				Type:             serviceerror.ClientErrorType,
				Code:             "CUSTOM_ERROR",
				Error:            "Custom error",
				ErrorDescription: "Custom error description",
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedErrorCode:  "CUSTOM_ERROR",
		},
		{
			name: "ServerError",
			authRequest: map[string]interface{}{
				"username": "testuser",
				"password": "testpass",
			},
			serviceError: &serviceerror.ServiceError{
				Type:             serviceerror.ServerErrorType,
				Code:             "INTERNAL_ERROR",
				Error:            "Internal error",
				ErrorDescription: "Internal error description",
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrorCode:  "INTERNAL_ERROR",
		},
	}

	for _, tc := range cases {
		suite.T().Run(tc.name, func(t *testing.T) {
			m := NewAuthenticationServiceInterfaceMock(t)
			m.On("AuthenticateWithCredentials", mock.Anything, mock.Anything, mock.Anything).Return(nil, tc.serviceError)
			h := &AuthenticationHandler{authService: m}

			body, _ := json.Marshal(tc.authRequest)
			req := httptest.NewRequest(http.MethodPost, "/authenticate/credentials", bytes.NewReader(body))
			w := httptest.NewRecorder()

			h.HandleCredentialsAuthRequest(w, req)

			suite.Equal(tc.expectedStatusCode, w.Code)
			var errResp apierror.ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &errResp)
			suite.NoError(err)
			suite.Equal(tc.expectedErrorCode, errResp.Code)
			m.AssertExpectations(t)
		})
	}
}

func (suite *AuthenticationHandlerTestSuite) TestHandleSendSMSOTPRequestSuccess() {
	otpRequest := SendOTPAuthRequestDTO{
		SenderID:  "sender123",
		Recipient: "+1234567890",
	}
	sessionToken := testSessionTkn

	suite.mockService.On("SendOTP", otpRequest.SenderID, mock.Anything, otpRequest.Recipient).
		Return(sessionToken, nil)

	body, _ := json.Marshal(otpRequest)
	req := httptest.NewRequest(http.MethodPost, "/authenticate/otp/send", bytes.NewReader(body))
	w := httptest.NewRecorder()

	suite.handler.HandleSendSMSOTPRequest(w, req)

	suite.Equal(http.StatusOK, w.Code)
	var response SendOTPAuthResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("SUCCESS", response.Status)
	suite.Equal(sessionToken, response.SessionToken)
}

func (suite *AuthenticationHandlerTestSuite) TestHandleSendSMSOTPRequestInvalidJSON() {
	req := httptest.NewRequest(http.MethodPost, "/authenticate/otp/send", bytes.NewReader([]byte("invalid-json")))
	w := httptest.NewRecorder()

	suite.handler.HandleSendSMSOTPRequest(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
	var errResp apierror.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	suite.NoError(err)
	suite.Equal(common.APIErrorInvalidRequestFormat.Code, errResp.Code)
}

func (suite *AuthenticationHandlerTestSuite) TestHandleSendSMSOTPRequestServiceError() {
	otpRequest := SendOTPAuthRequestDTO{
		SenderID:  "sender123",
		Recipient: "+1234567890",
	}
	serviceError := &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "OTP_ERROR",
		Error:            "OTP error",
		ErrorDescription: "Failed to send OTP",
	}

	suite.mockService.On("SendOTP", mock.Anything, mock.Anything, mock.Anything).Return("", serviceError)

	body, _ := json.Marshal(otpRequest)
	req := httptest.NewRequest(http.MethodPost, "/authenticate/otp/send", bytes.NewReader(body))
	w := httptest.NewRecorder()

	suite.handler.HandleSendSMSOTPRequest(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
	var errResp apierror.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	suite.NoError(err)
	suite.Equal("OTP_ERROR", errResp.Code)
}

func (suite *AuthenticationHandlerTestSuite) TestHandleVerifySMSOTPRequestSuccess() {
	otpRequest := VerifyOTPAuthRequestDTO{
		SessionToken:  testSessionTkn,
		SkipAssertion: false,
		OTP:           "123456",
	}
	authResponse := &common.AuthenticationResponse{
		ID:               "user123",
		Type:             "person",
		OrganizationUnit: "test-ou",
		Assertion:        "jwt-token",
	}

	suite.mockService.On("VerifyOTP", otpRequest.SessionToken, otpRequest.SkipAssertion, "", otpRequest.OTP).
		Return(authResponse, nil)

	body, _ := json.Marshal(otpRequest)
	req := httptest.NewRequest(http.MethodPost, "/authenticate/otp/verify", bytes.NewReader(body))
	w := httptest.NewRecorder()

	suite.handler.HandleVerifySMSOTPRequest(w, req)

	suite.Equal(http.StatusOK, w.Code)
	var response AuthenticationResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(authResponse.ID, response.ID)
	suite.Equal(authResponse.Assertion, response.Assertion)
}

func (suite *AuthenticationHandlerTestSuite) TestHandleVerifySMSOTPRequestInvalidJSON() {
	req := httptest.NewRequest(http.MethodPost, "/authenticate/otp/verify", bytes.NewReader([]byte("invalid-json")))
	w := httptest.NewRecorder()

	suite.handler.HandleVerifySMSOTPRequest(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
	var errResp apierror.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	suite.NoError(err)
	suite.Equal(common.APIErrorInvalidRequestFormat.Code, errResp.Code)
}

func (suite *AuthenticationHandlerTestSuite) TestHandleVerifySMSOTPRequestServiceError() {
	otpRequest := VerifyOTPAuthRequestDTO{
		SessionToken:  testSessionTkn,
		SkipAssertion: false,
		OTP:           "123456",
	}
	serviceError := &otp.ErrorIncorrectOTP

	suite.mockService.On("VerifyOTP", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, serviceError)

	body, _ := json.Marshal(otpRequest)
	req := httptest.NewRequest(http.MethodPost, "/authenticate/otp/verify", bytes.NewReader(body))
	w := httptest.NewRecorder()

	suite.handler.HandleVerifySMSOTPRequest(w, req)

	suite.Equal(http.StatusUnauthorized, w.Code)
	var errResp apierror.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	suite.NoError(err)
	suite.Equal(otp.ErrorIncorrectOTP.Code, errResp.Code)
}

func (suite *AuthenticationHandlerTestSuite) TestHandleGoogleAuthStartRequestSuccess() {
	authRequest := IDPAuthInitRequestDTO{
		IDPID: "google_idp_123",
	}
	authResponse := &IDPAuthInitData{
		RedirectURL:  "https://accounts.google.com/oauth/authorize",
		SessionToken: testSessionTkn,
	}

	suite.mockService.On("StartIDPAuthentication", idp.IDPTypeGoogle, authRequest.IDPID).
		Return(authResponse, nil)

	body, _ := json.Marshal(authRequest)
	req := httptest.NewRequest(http.MethodPost, "/authenticate/google/start", bytes.NewReader(body))
	w := httptest.NewRecorder()

	suite.handler.HandleGoogleAuthStartRequest(w, req)

	suite.Equal(http.StatusOK, w.Code)
	var response IDPAuthInitResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(authResponse.RedirectURL, response.RedirectURL)
	suite.Equal(authResponse.SessionToken, response.SessionToken)
}

func (suite *AuthenticationHandlerTestSuite) TestHandleGoogleAuthStartRequestInvalidJSON() {
	req := httptest.NewRequest(http.MethodPost, "/authenticate/google/start", bytes.NewReader([]byte("invalid-json")))
	w := httptest.NewRecorder()

	suite.handler.HandleGoogleAuthStartRequest(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
}

func (suite *AuthenticationHandlerTestSuite) TestHandleGoogleAuthStartRequestServiceError() {
	authRequest := IDPAuthInitRequestDTO{
		IDPID: "google_idp_123",
	}
	serviceError := &common.ErrorInvalidIDPID

	suite.mockService.On("StartIDPAuthentication", idp.IDPTypeGoogle, authRequest.IDPID).
		Return(nil, serviceError)

	body, _ := json.Marshal(authRequest)
	req := httptest.NewRequest(http.MethodPost, "/authenticate/google/start", bytes.NewReader(body))
	w := httptest.NewRecorder()

	suite.handler.HandleGoogleAuthStartRequest(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
	var errResp apierror.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	suite.NoError(err)
	suite.Equal(common.ErrorInvalidIDPID.Code, errResp.Code)
}

func (suite *AuthenticationHandlerTestSuite) TestHandleGoogleAuthFinishRequestSuccess() {
	suite.testIDPAuthFinishSuccess(idp.IDPTypeGoogle, "/authenticate/google/finish",
		suite.handler.HandleGoogleAuthFinishRequest)
}

func (suite *AuthenticationHandlerTestSuite) TestHandleGoogleAuthFinishRequestInvalidJSON() {
	req := httptest.NewRequest(http.MethodPost, "/authenticate/google/finish", bytes.NewReader([]byte("invalid-json")))
	w := httptest.NewRecorder()

	suite.handler.HandleGoogleAuthFinishRequest(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
}

func (suite *AuthenticationHandlerTestSuite) TestHandleGoogleAuthFinishRequestServiceError() {
	authRequest := IDPAuthFinishRequestDTO{
		SessionToken:  testSessionTkn,
		SkipAssertion: false,
		Code:          "auth_code_123",
	}
	serviceError := &common.ErrorInvalidSessionToken

	suite.mockService.On("FinishIDPAuthentication", idp.IDPTypeGoogle, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything).Return(nil, serviceError)

	body, _ := json.Marshal(authRequest)
	req := httptest.NewRequest(http.MethodPost, "/authenticate/google/finish", bytes.NewReader(body))
	w := httptest.NewRecorder()

	suite.handler.HandleGoogleAuthFinishRequest(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
	var errResp apierror.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	suite.NoError(err)
	suite.Equal(common.ErrorInvalidSessionToken.Code, errResp.Code)
}

func (suite *AuthenticationHandlerTestSuite) TestHandleGithubAuthStartRequestSuccess() {
	authRequest := IDPAuthInitRequestDTO{
		IDPID: "github_idp_123",
	}
	authResponse := &IDPAuthInitData{
		RedirectURL:  "https://github.com/login/oauth/authorize",
		SessionToken: testSessionTkn,
	}

	suite.mockService.On("StartIDPAuthentication", idp.IDPTypeGitHub, authRequest.IDPID).
		Return(authResponse, nil)

	body, _ := json.Marshal(authRequest)
	req := httptest.NewRequest(http.MethodPost, "/authenticate/github/start", bytes.NewReader(body))
	w := httptest.NewRecorder()

	suite.handler.HandleGithubAuthStartRequest(w, req)

	suite.Equal(http.StatusOK, w.Code)
	var response IDPAuthInitResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(authResponse.RedirectURL, response.RedirectURL)
}

func (suite *AuthenticationHandlerTestSuite) TestHandleGithubAuthStartRequestInvalidJSON() {
	req := httptest.NewRequest(http.MethodPost, "/authenticate/github/start", bytes.NewReader([]byte("invalid-json")))
	w := httptest.NewRecorder()

	suite.handler.HandleGithubAuthStartRequest(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
}

func (suite *AuthenticationHandlerTestSuite) TestHandleGithubAuthStartRequestServiceError() {
	authRequest := IDPAuthInitRequestDTO{
		IDPID: "github_idp_123",
	}
	serviceError := &common.ErrorClientErrorWhileRetrievingIDP

	suite.mockService.On("StartIDPAuthentication", idp.IDPTypeGitHub, authRequest.IDPID).
		Return(nil, serviceError)

	body, _ := json.Marshal(authRequest)
	req := httptest.NewRequest(http.MethodPost, "/authenticate/github/start", bytes.NewReader(body))
	w := httptest.NewRecorder()

	suite.handler.HandleGithubAuthStartRequest(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
}

func (suite *AuthenticationHandlerTestSuite) TestHandleGithubAuthFinishRequestSuccess() {
	authRequest := IDPAuthFinishRequestDTO{
		SessionToken:  testSessionTkn,
		SkipAssertion: true,
		Code:          "auth_code_123",
	}
	authResponse := &common.AuthenticationResponse{
		ID:               "user123",
		Type:             "person",
		OrganizationUnit: "test-ou",
	}

	suite.mockService.On("FinishIDPAuthentication", idp.IDPTypeGitHub, authRequest.SessionToken,
		authRequest.SkipAssertion, authRequest.Assertion, authRequest.Code).Return(authResponse, nil)

	body, _ := json.Marshal(authRequest)
	req := httptest.NewRequest(http.MethodPost, "/authenticate/github/finish", bytes.NewReader(body))
	w := httptest.NewRecorder()

	suite.handler.HandleGithubAuthFinishRequest(w, req)

	suite.Equal(http.StatusOK, w.Code)
	var response AuthenticationResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(authResponse.ID, response.ID)
	suite.Empty(response.Assertion)
}

func (suite *AuthenticationHandlerTestSuite) TestHandleGithubAuthFinishRequestInvalidJSON() {
	req := httptest.NewRequest(http.MethodPost, "/authenticate/github/finish", bytes.NewReader([]byte("invalid-json")))
	w := httptest.NewRecorder()

	suite.handler.HandleGithubAuthFinishRequest(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
}

func (suite *AuthenticationHandlerTestSuite) TestHandleGithubAuthFinishRequestServiceError() {
	authRequest := IDPAuthFinishRequestDTO{
		SessionToken:  testSessionTkn,
		SkipAssertion: false,
		Code:          "auth_code_123",
	}
	serviceError := &serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "INTERNAL_ERROR",
		Error:            "Internal error",
		ErrorDescription: "Internal error description",
	}

	suite.mockService.On("FinishIDPAuthentication", idp.IDPTypeGitHub, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything).Return(nil, serviceError)

	body, _ := json.Marshal(authRequest)
	req := httptest.NewRequest(http.MethodPost, "/authenticate/github/finish", bytes.NewReader(body))
	w := httptest.NewRecorder()

	suite.handler.HandleGithubAuthFinishRequest(w, req)

	suite.Equal(http.StatusInternalServerError, w.Code)
}

func (suite *AuthenticationHandlerTestSuite) TestHandleStandardOAuthStartRequestSuccess() {
	authRequest := IDPAuthInitRequestDTO{
		IDPID: "oauth_idp_123",
	}
	authResponse := &IDPAuthInitData{
		RedirectURL:  "https://oauth.provider.com/authorize",
		SessionToken: testSessionTkn,
	}

	suite.mockService.On("StartIDPAuthentication", idp.IDPTypeOAuth, authRequest.IDPID).
		Return(authResponse, nil)

	body, _ := json.Marshal(authRequest)
	req := httptest.NewRequest(http.MethodPost, "/authenticate/oauth/start", bytes.NewReader(body))
	w := httptest.NewRecorder()

	suite.handler.HandleStandardOAuthStartRequest(w, req)

	suite.Equal(http.StatusOK, w.Code)
	var response IDPAuthInitResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(authResponse.RedirectURL, response.RedirectURL)
}

func (suite *AuthenticationHandlerTestSuite) TestHandleStandardOAuthStartRequestInvalidJSON() {
	req := httptest.NewRequest(http.MethodPost, "/authenticate/oauth/start", bytes.NewReader([]byte("invalid-json")))
	w := httptest.NewRecorder()

	suite.handler.HandleStandardOAuthStartRequest(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
}

func (suite *AuthenticationHandlerTestSuite) TestHandleStandardOAuthStartRequestServiceError() {
	authRequest := IDPAuthInitRequestDTO{
		IDPID: "oauth_idp_123",
	}
	serviceError := &common.ErrorInvalidIDPType

	suite.mockService.On("StartIDPAuthentication", idp.IDPTypeOAuth, authRequest.IDPID).
		Return(nil, serviceError)

	body, _ := json.Marshal(authRequest)
	req := httptest.NewRequest(http.MethodPost, "/authenticate/oauth/start", bytes.NewReader(body))
	w := httptest.NewRecorder()

	suite.handler.HandleStandardOAuthStartRequest(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
}

func (suite *AuthenticationHandlerTestSuite) TestHandleStandardOAuthFinishRequestSuccess() {
	suite.testIDPAuthFinishSuccess(idp.IDPTypeOAuth, "/authenticate/oauth/finish",
		suite.handler.HandleStandardOAuthFinishRequest)
}

func (suite *AuthenticationHandlerTestSuite) TestHandleStandardOAuthFinishRequestInvalidJSON() {
	req := httptest.NewRequest(http.MethodPost, "/authenticate/oauth/finish", bytes.NewReader([]byte("invalid-json")))
	w := httptest.NewRecorder()

	suite.handler.HandleStandardOAuthFinishRequest(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
}

func (suite *AuthenticationHandlerTestSuite) TestHandleStandardOAuthFinishRequestServiceError() {
	authRequest := IDPAuthFinishRequestDTO{
		SessionToken:  testSessionTkn,
		SkipAssertion: false,
		Code:          "auth_code_123",
	}
	serviceError := &common.ErrorEmptyAuthCode

	suite.mockService.On("FinishIDPAuthentication", idp.IDPTypeOAuth, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything).Return(nil, serviceError)

	body, _ := json.Marshal(authRequest)
	req := httptest.NewRequest(http.MethodPost, "/authenticate/oauth/finish", bytes.NewReader(body))
	w := httptest.NewRecorder()

	suite.handler.HandleStandardOAuthFinishRequest(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
}
