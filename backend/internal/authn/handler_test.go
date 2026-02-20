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
	handler     *authenticationHandler
}

func TestAuthenticationHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(AuthenticationHandlerTestSuite))
}

func (suite *AuthenticationHandlerTestSuite) SetupTest() {
	suite.mockService = NewAuthenticationServiceInterfaceMock(suite.T())
	suite.handler = &authenticationHandler{
		authService: suite.mockService,
	}
}

func (suite *AuthenticationHandlerTestSuite) TestNewAuthenticationHandler() {
	mockService := NewAuthenticationServiceInterfaceMock(suite.T())

	handler := newAuthenticationHandler(mockService)

	suite.NotNil(handler)
	suite.Equal(mockService, handler.authService)
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
	identifiers := map[string]interface{}{
		"username": "testuser",
	}
	credentials := map[string]interface{}{
		"password": "testpass",
	}
	authRequest := map[string]interface{}{
		"identifiers": identifiers,
		"credentials": credentials,
	}
	authResponse := &common.AuthenticationResponse{
		ID:               "user123",
		Type:             "person",
		OrganizationUnit: "test-ou",
		Assertion:        "jwt-token",
	}

	suite.mockService.On("AuthenticateWithCredentials", identifiers, credentials, false, "").Return(authResponse, nil)

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
	identifiers := map[string]interface{}{
		"username": "testuser",
	}
	credentials := map[string]interface{}{
		"password": "testpass",
	}
	authRequest := map[string]interface{}{
		"identifiers":    identifiers,
		"credentials":    credentials,
		"skip_assertion": true,
	}
	authResponse := &common.AuthenticationResponse{
		ID:               "user123",
		Type:             "person",
		OrganizationUnit: "test-ou",
	}

	suite.mockService.On("AuthenticateWithCredentials", identifiers, credentials, true, "").Return(authResponse, nil)

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
	identifiers := map[string]interface{}{
		"username": "testuser",
	}
	credentials := map[string]interface{}{
		"password": "testpass",
	}
	authRequest := map[string]interface{}{
		"identifiers": identifiers,
		"credentials": credentials,
		"assertion":   existingAssertion,
	}
	authResponse := &common.AuthenticationResponse{
		ID:               "user123",
		Type:             "person",
		OrganizationUnit: "test-ou",
		Assertion:        "updated.jwt.token",
	}

	suite.mockService.On("AuthenticateWithCredentials", identifiers, credentials, false, existingAssertion).
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
				"identifiers": map[string]interface{}{
					"username": "testuser",
				},
				"credentials": map[string]interface{}{
					"password": "wrongpass",
				},
			},
			serviceError:       &credentials.ErrorInvalidCredentials,
			expectedStatusCode: http.StatusUnauthorized,
			expectedErrorCode:  credentials.ErrorInvalidCredentials.Code,
		},
		{
			name: "UserNotFound",
			authRequest: map[string]interface{}{
				"identifiers": map[string]interface{}{
					"username": "nonexistent",
				},
				"credentials": map[string]interface{}{
					"password": "testpass",
				},
			},
			serviceError:       &common.ErrorUserNotFound,
			expectedStatusCode: http.StatusNotFound,
			expectedErrorCode:  common.ErrorUserNotFound.Code,
		},
		{
			name: "ClientError",
			authRequest: map[string]interface{}{
				"identifiers": map[string]interface{}{
					"username": "testuser",
				},
				"credentials": map[string]interface{}{
					"password": "testpass",
				},
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
				"identifiers": map[string]interface{}{
					"username": "testuser",
				},
				"credentials": map[string]interface{}{
					"password": "testpass",
				},
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
			m.On("AuthenticateWithCredentials", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(nil, tc.serviceError)
			h := &authenticationHandler{authService: m}

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

	suite.mockService.On("VerifyOTP", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, serviceError)

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

func (suite *AuthenticationHandlerTestSuite) TestHandlePasskeyRegisterStartRequestSuccess() {
	regRequest := PasskeyRegisterStartRequestDTO{
		UserID:           "user123",
		RelyingPartyID:   "example.com",
		RelyingPartyName: "Example Corp",
		Attestation:      "direct",
	}
	regResponse := map[string]interface{}{
		"publicKeyCredentialCreationOptions": map[string]interface{}{
			"challenge": "base64-challenge",
			"rp": map[string]interface{}{
				"name": "Example Corp",
				"id":   "example.com",
			},
			"user": map[string]interface{}{
				"id":          "user123",
				"name":        "testuser",
				"displayName": "Test User",
			},
		},
		"sessionToken": testSessionTkn,
	}

	suite.mockService.On("StartPasskeyRegistration",
		regRequest.UserID,
		regRequest.RelyingPartyID,
		regRequest.RelyingPartyName,
		regRequest.AuthenticatorSelection,
		regRequest.Attestation).Return(regResponse, nil)

	body, _ := json.Marshal(regRequest)
	req := httptest.NewRequest(http.MethodPost, "/authenticate/passkey/register/start", bytes.NewReader(body))
	w := httptest.NewRecorder()

	suite.handler.HandlePasskeyRegisterStartRequest(w, req)

	suite.Equal(http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(testSessionTkn, response["sessionToken"])
	suite.NotNil(response["publicKeyCredentialCreationOptions"])
}

func (suite *AuthenticationHandlerTestSuite) TestHandlePasskeyRegisterStartRequestInvalidJSON() {
	req := httptest.NewRequest(http.MethodPost, "/authenticate/passkey/register/start",
		bytes.NewReader([]byte("invalid-json")))
	w := httptest.NewRecorder()

	suite.handler.HandlePasskeyRegisterStartRequest(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
	var errResp apierror.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	suite.NoError(err)
	suite.Equal(common.APIErrorInvalidRequestFormat.Code, errResp.Code)
}

func (suite *AuthenticationHandlerTestSuite) TestHandlePasskeyRegisterStartRequestServiceError() {
	regRequest := PasskeyRegisterStartRequestDTO{
		UserID:         "user123",
		RelyingPartyID: "example.com",
	}
	serviceError := &common.ErrorUserNotFound

	suite.mockService.On("StartPasskeyRegistration",
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, serviceError)

	body, _ := json.Marshal(regRequest)
	req := httptest.NewRequest(http.MethodPost, "/authenticate/passkey/register/start", bytes.NewReader(body))
	w := httptest.NewRecorder()

	suite.handler.HandlePasskeyRegisterStartRequest(w, req)

	suite.Equal(http.StatusNotFound, w.Code)
	var errResp apierror.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	suite.NoError(err)
	suite.Equal(common.ErrorUserNotFound.Code, errResp.Code)
}

func (suite *AuthenticationHandlerTestSuite) TestHandlePasskeyRegisterFinishRequestSuccess() {
	regRequest := PasskeyRegisterFinishRequestDTO{
		PublicKeyCredential: PasskeyPublicKeyCredentialDTO{
			ID:   "credential-id-123",
			Type: "public-key",
			Response: PasskeyCredentialResponseDTO{
				ClientDataJSON:    "base64-client-data",
				AttestationObject: "base64-attestation",
			},
		},
		SessionToken:   testSessionTkn,
		CredentialName: "My Passkey",
	}
	regResponse := map[string]interface{}{
		"credentialId":   "credential-id-123",
		"credentialName": "My Passkey",
		"createdAt":      "2025-01-01T00:00:00Z",
	}

	suite.mockService.On("FinishPasskeyRegistration",
		regRequest.PublicKeyCredential,
		testSessionTkn,
		"My Passkey").Return(regResponse, nil)

	body, _ := json.Marshal(regRequest)
	req := httptest.NewRequest(http.MethodPost, "/authenticate/passkey/register/finish", bytes.NewReader(body))
	w := httptest.NewRecorder()

	suite.handler.HandlePasskeyRegisterFinishRequest(w, req)

	suite.Equal(http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("credential-id-123", response["credentialId"])
	suite.Equal("My Passkey", response["credentialName"])
}

func (suite *AuthenticationHandlerTestSuite) TestHandlePasskeyRegisterFinishRequestInvalidJSON() {
	req := httptest.NewRequest(http.MethodPost, "/authenticate/passkey/register/finish",
		bytes.NewReader([]byte("invalid-json")))
	w := httptest.NewRecorder()

	suite.handler.HandlePasskeyRegisterFinishRequest(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
	var errResp apierror.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	suite.NoError(err)
	suite.Equal(common.APIErrorInvalidRequestFormat.Code, errResp.Code)
}

func (suite *AuthenticationHandlerTestSuite) TestHandlePasskeyRegisterFinishRequestServiceError() {
	regRequest := PasskeyRegisterFinishRequestDTO{
		PublicKeyCredential: PasskeyPublicKeyCredentialDTO{
			ID:   "credential-id-123",
			Type: "public-key",
			Response: PasskeyCredentialResponseDTO{
				ClientDataJSON:    "base64-client-data",
				AttestationObject: "base64-attestation",
			},
		},
		SessionToken: testSessionTkn,
	}
	serviceError := &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "INVALID_ATTESTATION",
		Error:            "Invalid attestation",
		ErrorDescription: "Failed to verify attestation",
	}

	suite.mockService.On("FinishPasskeyRegistration",
		mock.Anything, mock.Anything, mock.Anything).Return(nil, serviceError)

	body, _ := json.Marshal(regRequest)
	req := httptest.NewRequest(http.MethodPost, "/authenticate/passkey/register/finish", bytes.NewReader(body))
	w := httptest.NewRecorder()

	suite.handler.HandlePasskeyRegisterFinishRequest(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
	var errResp apierror.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	suite.NoError(err)
	suite.Equal("INVALID_ATTESTATION", errResp.Code)
}

func (suite *AuthenticationHandlerTestSuite) TestHandlePasskeyStartRequestSuccess() {
	authRequest := PasskeyStartRequestDTO{
		UserID:         "user123",
		RelyingPartyID: "example.com",
	}
	authResponse := map[string]interface{}{
		"publicKeyCredentialRequestOptions": map[string]interface{}{
			"challenge": "base64-challenge",
			"rpId":      "example.com",
			"allowCredentials": []map[string]interface{}{
				{
					"type": "public-key",
					"id":   "credential-id-123",
				},
			},
		},
		"sessionToken": testSessionTkn,
	}

	suite.mockService.On("StartPasskeyAuthentication",
		authRequest.UserID,
		authRequest.RelyingPartyID).Return(authResponse, nil)

	body, _ := json.Marshal(authRequest)
	req := httptest.NewRequest(http.MethodPost, "/authenticate/passkey/start", bytes.NewReader(body))
	w := httptest.NewRecorder()

	suite.handler.HandlePasskeyStartRequest(w, req)

	suite.Equal(http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(testSessionTkn, response["sessionToken"])
	suite.NotNil(response["publicKeyCredentialRequestOptions"])
}

func (suite *AuthenticationHandlerTestSuite) TestHandlePasskeyStartRequestInvalidJSON() {
	req := httptest.NewRequest(http.MethodPost, "/authenticate/passkey/start",
		bytes.NewReader([]byte("invalid-json")))
	w := httptest.NewRecorder()

	suite.handler.HandlePasskeyStartRequest(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
	var errResp apierror.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	suite.NoError(err)
	suite.Equal(common.APIErrorInvalidRequestFormat.Code, errResp.Code)
}

func (suite *AuthenticationHandlerTestSuite) TestHandlePasskeyStartRequestServiceError() {
	authRequest := PasskeyStartRequestDTO{
		UserID:         "nonexistent",
		RelyingPartyID: "example.com",
	}
	serviceError := &common.ErrorUserNotFound

	suite.mockService.On("StartPasskeyAuthentication",
		mock.Anything, mock.Anything).Return(nil, serviceError)

	body, _ := json.Marshal(authRequest)
	req := httptest.NewRequest(http.MethodPost, "/authenticate/passkey/start", bytes.NewReader(body))
	w := httptest.NewRecorder()

	suite.handler.HandlePasskeyStartRequest(w, req)

	suite.Equal(http.StatusNotFound, w.Code)
	var errResp apierror.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	suite.NoError(err)
	suite.Equal(common.ErrorUserNotFound.Code, errResp.Code)
}

func (suite *AuthenticationHandlerTestSuite) TestHandlePasskeyFinishRequestSuccess() {
	authRequest := PasskeyFinishRequestDTO{
		PublicKeyCredential: PasskeyPublicKeyCredentialDTO{
			ID:   "credential-id-123",
			Type: "public-key",
			Response: PasskeyCredentialResponseDTO{
				ClientDataJSON:    "base64-client-data",
				AuthenticatorData: "base64-auth-data",
				Signature:         "base64-signature",
			},
		},
		SessionToken:  testSessionTkn,
		SkipAssertion: false,
		Assertion:     "",
	}
	authResponse := &common.AuthenticationResponse{
		ID:               "user123",
		Type:             "person",
		OrganizationUnit: "test-ou",
		Assertion:        "jwt-token",
	}

	suite.mockService.On("FinishPasskeyAuthentication",
		authRequest.PublicKeyCredential.ID,
		authRequest.PublicKeyCredential.Type,
		authRequest.PublicKeyCredential.Response,
		testSessionTkn,
		false,
		"").Return(authResponse, nil)

	body, _ := json.Marshal(authRequest)
	req := httptest.NewRequest(http.MethodPost, "/authenticate/passkey/finish", bytes.NewReader(body))
	w := httptest.NewRecorder()

	suite.handler.HandlePasskeyFinishRequest(w, req)

	suite.Equal(http.StatusOK, w.Code)
	var response AuthenticationResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(authResponse.ID, response.ID)
	suite.Equal(authResponse.Assertion, response.Assertion)
}

func (suite *AuthenticationHandlerTestSuite) TestHandlePasskeyFinishRequestInvalidJSON() {
	req := httptest.NewRequest(http.MethodPost, "/authenticate/passkey/finish",
		bytes.NewReader([]byte("invalid-json")))
	w := httptest.NewRecorder()

	suite.handler.HandlePasskeyFinishRequest(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
	var errResp apierror.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	suite.NoError(err)
	suite.Equal(common.APIErrorInvalidRequestFormat.Code, errResp.Code)
}

func (suite *AuthenticationHandlerTestSuite) TestHandlePasskeyFinishRequestServiceError() {
	authRequest := PasskeyFinishRequestDTO{
		PublicKeyCredential: PasskeyPublicKeyCredentialDTO{
			ID:   "credential-id-123",
			Type: "public-key",
			Response: PasskeyCredentialResponseDTO{
				ClientDataJSON: "base64-client-data",
			},
		},
		SessionToken: testSessionTkn,
	}
	serviceError := &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "INVALID_SIGNATURE",
		Error:            "Invalid signature",
		ErrorDescription: "Failed to verify signature",
	}

	suite.mockService.On("FinishPasskeyAuthentication",
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, serviceError)

	body, _ := json.Marshal(authRequest)
	req := httptest.NewRequest(http.MethodPost, "/authenticate/passkey/finish", bytes.NewReader(body))
	w := httptest.NewRecorder()

	suite.handler.HandlePasskeyFinishRequest(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
	var errResp apierror.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	suite.NoError(err)
	suite.Equal("INVALID_SIGNATURE", errResp.Code)
}
