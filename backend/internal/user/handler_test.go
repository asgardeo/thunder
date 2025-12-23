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

package user

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/asgardeo/thunder/internal/system/crypto/hash"
	"github.com/asgardeo/thunder/internal/system/error/apierror"
	"github.com/asgardeo/thunder/internal/system/security"
	"github.com/asgardeo/thunder/tests/mocks/crypto/hashmock"
)

const testUserID = "test-user-123"

func TestHandleSelfUserGetRequest_Success(t *testing.T) {
	userID := testUserID
	authCtx := security.NewSecurityContextForTest(userID, "", "", "", nil)

	mockSvc := NewUserServiceInterfaceMock(t)
	hashMock := hashmock.NewHashServiceInterfaceMock(t)
	expectedUser := &User{
		ID:         userID,
		Attributes: json.RawMessage(`{"username":"alice"}`),
	}
	mockSvc.On("GetUser", userID).Return(expectedUser, nil)

	handler := newUserHandler(mockSvc, hashMock)
	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req = req.WithContext(security.WithSecurityContextTest(req.Context(), authCtx))
	rr := httptest.NewRecorder()

	handler.HandleSelfUserGetRequest(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Header().Get("Content-Type"), "application/json")

	var respUser User
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&respUser))
	require.Equal(t, expectedUser.ID, respUser.ID)
	require.JSONEq(t, string(expectedUser.Attributes), string(respUser.Attributes))
}

func TestHandleSelfUserGetRequest_Unauthorized(t *testing.T) {
	mockSvc := NewUserServiceInterfaceMock(t)
	hashMock := hashmock.NewHashServiceInterfaceMock(t)
	handler := newUserHandler(mockSvc, hashMock)
	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	rr := httptest.NewRecorder()

	handler.HandleSelfUserGetRequest(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)

	var errResp apierror.ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&errResp))
	require.Equal(t, ErrorAuthenticationFailed.Code, errResp.Code)
}

func TestHandleSelfUserPutRequest_Success(t *testing.T) {
	userID := testUserID
	authCtx := security.NewSecurityContextForTest(userID, "", "", "", nil)
	attributes := json.RawMessage(`{"email":"alice@example.com"}`)

	mockSvc := NewUserServiceInterfaceMock(t)
	hashMock := hashmock.NewHashServiceInterfaceMock(t)
	updatedUser := &User{
		ID:         userID,
		Type:       "employee",
		Attributes: attributes,
	}
	mockSvc.On("UpdateUserAttributes", userID, attributes).Return(updatedUser, nil)

	handler := newUserHandler(mockSvc, hashMock)
	body := bytes.NewBufferString(`{"attributes":{"email":"alice@example.com"}}`)
	req := httptest.NewRequest(http.MethodPut, "/users/me", body)
	req = req.WithContext(security.WithSecurityContextTest(req.Context(), authCtx))
	rr := httptest.NewRecorder()

	handler.HandleSelfUserPutRequest(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var respUser User
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&respUser))
	require.Equal(t, updatedUser.ID, respUser.ID)
	require.JSONEq(t, string(updatedUser.Attributes), string(respUser.Attributes))
}

func TestHandleSelfUserPutRequest_InvalidBody(t *testing.T) {
	userID := testUserID
	authCtx := security.NewSecurityContextForTest(userID, "", "", "", nil)

	mockSvc := NewUserServiceInterfaceMock(t)
	hashMock := hashmock.NewHashServiceInterfaceMock(t)
	handler := newUserHandler(mockSvc, hashMock)

	req := httptest.NewRequest(http.MethodPut, "/users/me", bytes.NewBufferString(`{"attributes":`))
	req = req.WithContext(security.WithSecurityContextTest(req.Context(), authCtx))
	rr := httptest.NewRecorder()

	handler.HandleSelfUserPutRequest(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)

	var errResp apierror.ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&errResp))
	require.Equal(t, ErrorInvalidRequestFormat.Code, errResp.Code)
}

func TestHandleSelfUserCredentialUpdateRequest_Success(t *testing.T) {
	userID := testUserID
	authCtx := security.NewSecurityContextForTest(userID, "", "", "", nil)

	mockSvc := NewUserServiceInterfaceMock(t)
	// Expect hashed credentials to be passed to service
	mockSvc.On("UpdateUserCredentials", userID, "password", mock.MatchedBy(func(creds []Credential) bool {
		return len(creds) == 1 &&
			creds[0].StorageType == "hash" &&
			creds[0].Value == "hashed-password" &&
			creds[0].StorageAlgo == hash.PBKDF2
	})).Return(nil)

	// Mock hash service
	hashMock := hashmock.NewHashServiceInterfaceMock(t)
	hashMock.On("Generate", []byte("Secret123!")).Return(hash.Credential{
		Algorithm: hash.PBKDF2,
		Hash:      "hashed-password",
		Parameters: hash.CredParameters{
			Salt:       "salt123",
			Iterations: 10000,
			KeySize:    32,
		},
	}, nil).Once()

	handler := newUserHandler(mockSvc, hashMock)
	req := httptest.NewRequest(http.MethodPost, "/users/me/update-credentials",
		bytes.NewBufferString(`{"attributes":{"password":[{"value":"Secret123!"}]}}`))
	req = req.WithContext(security.WithSecurityContextTest(req.Context(), authCtx))
	rr := httptest.NewRecorder()

	handler.HandleSelfUserCredentialUpdateRequest(rr, req)

	require.Equal(t, http.StatusNoContent, rr.Code)
	require.Equal(t, 0, rr.Body.Len())
}

func TestHandleSelfUserCredentialUpdateRequest_StringValue(t *testing.T) {
	userID := testUserID
	authCtx := security.NewSecurityContextForTest(userID, "", "", "", nil)

	mockSvc := NewUserServiceInterfaceMock(t)
	// Expect hashed credentials
	mockSvc.On("UpdateUserCredentials", userID, "password", mock.MatchedBy(func(creds []Credential) bool {
		return len(creds) == 1 &&
			creds[0].StorageType == "hash" &&
			creds[0].Value == "hashed-plaintext-pw"
	})).Return(nil)

	// Mock hash service
	hashMock := hashmock.NewHashServiceInterfaceMock(t)
	hashMock.On("Generate", []byte("plaintext-password")).Return(hash.Credential{
		Algorithm: hash.PBKDF2,
		Hash:      "hashed-plaintext-pw",
		Parameters: hash.CredParameters{
			Salt:       "salt456",
			Iterations: 10000,
			KeySize:    32,
		},
	}, nil).Once()

	handler := newUserHandler(mockSvc, hashMock)
	// Send string value (handler should hash it)
	req := httptest.NewRequest(http.MethodPost, "/users/me/update-credentials",
		bytes.NewBufferString(`{"attributes":{"password":"plaintext-password"}}`))
	req = req.WithContext(security.WithSecurityContextTest(req.Context(), authCtx))
	rr := httptest.NewRecorder()

	handler.HandleSelfUserCredentialUpdateRequest(rr, req)

	require.Equal(t, http.StatusNoContent, rr.Code)
	require.Equal(t, 0, rr.Body.Len())
}

func TestHandleSelfUserCredentialUpdateRequest_MissingCredentials(t *testing.T) {
	userID := testUserID
	authCtx := security.NewSecurityContextForTest(userID, "", "", "", nil)

	mockSvc := NewUserServiceInterfaceMock(t)
	hashMock := hashmock.NewHashServiceInterfaceMock(t)
	// No expectations because UpdateUserCredentials should not be called when attributes is empty
	handler := newUserHandler(mockSvc, hashMock)

	req := httptest.NewRequest(http.MethodPost, "/users/me/update-credentials",
		bytes.NewBufferString(`{"attributes":{}}`))
	req = req.WithContext(security.WithSecurityContextTest(req.Context(), authCtx))
	rr := httptest.NewRecorder()

	handler.HandleSelfUserCredentialUpdateRequest(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)

	var errResp apierror.ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&errResp))
	require.Equal(t, ErrorMissingCredentials.Code, errResp.Code)
}

func TestHandleSelfUserCredentialUpdateRequest_UnsupportedCredentialType(t *testing.T) {
	userID := testUserID
	authCtx := security.NewSecurityContextForTest(userID, "", "", "", nil)

	mockSvc := NewUserServiceInterfaceMock(t)
	hashMock := hashmock.NewHashServiceInterfaceMock(t)
	// No expectations - validation should fail before calling service or hash service
	handler := newUserHandler(mockSvc, hashMock)

	// Try to update with unsupported credential type "biometric"
	req := httptest.NewRequest(http.MethodPost, "/users/me/update-credentials",
		bytes.NewBufferString(`{"attributes":{"biometric":"fingerprint-data"}}`))
	req = req.WithContext(security.WithSecurityContextTest(req.Context(), authCtx))
	rr := httptest.NewRecorder()

	handler.HandleSelfUserCredentialUpdateRequest(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)

	var errResp apierror.ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&errResp))
	require.Equal(t, ErrorInvalidCredential.Code, errResp.Code)
	require.Contains(t, errResp.Description, "Invalid credential type: biometric")

	// Verify that neither the hash service nor user service were called
	hashMock.AssertNotCalled(t, "Generate", mock.Anything)
	mockSvc.AssertNotCalled(t, "UpdateUserCredentials", mock.Anything, mock.Anything, mock.Anything)
}

func TestHandleSelfUserGetRequest_ServiceError(t *testing.T) {
	userID := testUserID
	authCtx := security.NewSecurityContextForTest(userID, "", "", "", nil)

	mockSvc := NewUserServiceInterfaceMock(t)
	hashMock := hashmock.NewHashServiceInterfaceMock(t)
	mockSvc.On("GetUser", userID).Return(nil, &ErrorUserNotFound)

	handler := newUserHandler(mockSvc, hashMock)
	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req = req.WithContext(security.WithSecurityContextTest(req.Context(), authCtx))
	rr := httptest.NewRecorder()

	handler.HandleSelfUserGetRequest(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)

	var errResp apierror.ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&errResp))
	require.Equal(t, ErrorUserNotFound.Code, errResp.Code)
}

func TestHandleSelfUserPutRequest_Unauthorized(t *testing.T) {
	mockSvc := NewUserServiceInterfaceMock(t)
	hashMock := hashmock.NewHashServiceInterfaceMock(t)
	handler := newUserHandler(mockSvc, hashMock)

	body := bytes.NewBufferString(`{"attributes":{"email":"alice@example.com"}}`)
	req := httptest.NewRequest(http.MethodPut, "/users/me", body)
	rr := httptest.NewRecorder()

	handler.HandleSelfUserPutRequest(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)

	var errResp apierror.ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&errResp))
	require.Equal(t, ErrorAuthenticationFailed.Code, errResp.Code)
}

func TestHandleSelfUserPutRequest_EmptyAttributes(t *testing.T) {
	userID := testUserID
	authCtx := security.NewSecurityContextForTest(userID, "", "", "", nil)

	mockSvc := NewUserServiceInterfaceMock(t)
	hashMock := hashmock.NewHashServiceInterfaceMock(t)
	handler := newUserHandler(mockSvc, hashMock)

	// Send nil attributes (the handler checks len(updateRequest.Attributes) == 0)
	body := bytes.NewBufferString(`{}`)
	req := httptest.NewRequest(http.MethodPut, "/users/me", body)
	req = req.WithContext(security.WithSecurityContextTest(req.Context(), authCtx))
	rr := httptest.NewRecorder()

	handler.HandleSelfUserPutRequest(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)

	var errResp apierror.ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&errResp))
	require.Equal(t, ErrorInvalidRequestFormat.Code, errResp.Code)
}

func TestHandleSelfUserPutRequest_ServiceError(t *testing.T) {
	userID := testUserID
	authCtx := security.NewSecurityContextForTest(userID, "", "", "", nil)
	attributes := json.RawMessage(`{"email":"alice@example.com"}`)

	mockSvc := NewUserServiceInterfaceMock(t)
	hashMock := hashmock.NewHashServiceInterfaceMock(t)
	mockSvc.On("UpdateUserAttributes", userID, attributes).Return(nil, &ErrorInternalServerError)

	handler := newUserHandler(mockSvc, hashMock)
	body := bytes.NewBufferString(`{"attributes":{"email":"alice@example.com"}}`)
	req := httptest.NewRequest(http.MethodPut, "/users/me", body)
	req = req.WithContext(security.WithSecurityContextTest(req.Context(), authCtx))
	rr := httptest.NewRecorder()

	handler.HandleSelfUserPutRequest(rr, req)

	require.Equal(t, http.StatusInternalServerError, rr.Code)

	var errResp apierror.ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&errResp))
	require.Equal(t, ErrorInternalServerError.Code, errResp.Code)
}

func TestHandleSelfUserCredentialUpdateRequest_Unauthorized(t *testing.T) {
	mockSvc := NewUserServiceInterfaceMock(t)
	hashMock := hashmock.NewHashServiceInterfaceMock(t)
	handler := newUserHandler(mockSvc, hashMock)

	req := httptest.NewRequest(http.MethodPost, "/users/me/update-credentials",
		bytes.NewBufferString(`{"attributes":{"password":"Secret123!"}}`))
	rr := httptest.NewRecorder()

	handler.HandleSelfUserCredentialUpdateRequest(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)

	var errResp apierror.ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&errResp))
	require.Equal(t, ErrorAuthenticationFailed.Code, errResp.Code)
}

func TestHandleSelfUserCredentialUpdateRequest_InvalidJSON(t *testing.T) {
	userID := testUserID
	authCtx := security.NewSecurityContextForTest(userID, "", "", "", nil)

	mockSvc := NewUserServiceInterfaceMock(t)
	hashMock := hashmock.NewHashServiceInterfaceMock(t)
	handler := newUserHandler(mockSvc, hashMock)

	req := httptest.NewRequest(http.MethodPost, "/users/me/update-credentials",
		bytes.NewBufferString(`{invalid json`))
	req = req.WithContext(security.WithSecurityContextTest(req.Context(), authCtx))
	rr := httptest.NewRecorder()

	handler.HandleSelfUserCredentialUpdateRequest(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)

	var errResp apierror.ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&errResp))
	require.Equal(t, ErrorInvalidRequestFormat.Code, errResp.Code)
}

func TestHandleSelfUserCredentialUpdateRequest_InvalidAttributesJSON(t *testing.T) {
	userID := testUserID
	authCtx := security.NewSecurityContextForTest(userID, "", "", "", nil)

	mockSvc := NewUserServiceInterfaceMock(t)
	hashMock := hashmock.NewHashServiceInterfaceMock(t)
	handler := newUserHandler(mockSvc, hashMock)

	// Send invalid JSON for attributes that will fail unmarshaling
	req := httptest.NewRequest(http.MethodPost, "/users/me/update-credentials",
		bytes.NewBufferString(`{"attributes":"invalid"}`))
	req = req.WithContext(security.WithSecurityContextTest(req.Context(), authCtx))
	rr := httptest.NewRecorder()

	handler.HandleSelfUserCredentialUpdateRequest(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)

	var errResp apierror.ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&errResp))
	require.Equal(t, ErrorInvalidRequestFormat.Code, errResp.Code)
}

func TestHandleSelfUserCredentialUpdateRequest_InvalidCredentialValue(t *testing.T) {
	userID := testUserID
	authCtx := security.NewSecurityContextForTest(userID, "", "", "", nil)

	mockSvc := NewUserServiceInterfaceMock(t)
	hashMock := hashmock.NewHashServiceInterfaceMock(t)
	handler := newUserHandler(mockSvc, hashMock)

	// Send credential value that is neither array nor string (e.g., object)
	req := httptest.NewRequest(http.MethodPost, "/users/me/update-credentials",
		bytes.NewBufferString(`{"attributes":{"password":{"invalid":"object"}}}`))
	req = req.WithContext(security.WithSecurityContextTest(req.Context(), authCtx))
	rr := httptest.NewRecorder()

	handler.HandleSelfUserCredentialUpdateRequest(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)

	var errResp apierror.ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&errResp))
	require.Equal(t, ErrorInvalidRequestFormat.Code, errResp.Code)
}

func TestHandleSelfUserCredentialUpdateRequest_HashingError(t *testing.T) {
	userID := testUserID
	authCtx := security.NewSecurityContextForTest(userID, "", "", "", nil)

	mockSvc := NewUserServiceInterfaceMock(t)
	hashMock := hashmock.NewHashServiceInterfaceMock(t)

	// Mock hash service to return an error
	hashMock.On("Generate", []byte("Secret123!")).Return(hash.Credential{},
		errors.New("hashing failed"))

	handler := newUserHandler(mockSvc, hashMock)
	req := httptest.NewRequest(http.MethodPost, "/users/me/update-credentials",
		bytes.NewBufferString(`{"attributes":{"password":"Secret123!"}}`))
	req = req.WithContext(security.WithSecurityContextTest(req.Context(), authCtx))
	rr := httptest.NewRecorder()

	handler.HandleSelfUserCredentialUpdateRequest(rr, req)

	require.Equal(t, http.StatusInternalServerError, rr.Code)

	var errResp apierror.ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&errResp))
	require.Equal(t, ErrorInternalServerError.Code, errResp.Code)
}

func TestHandleSelfUserCredentialUpdateRequest_UpdateServiceError(t *testing.T) {
	userID := testUserID
	authCtx := security.NewSecurityContextForTest(userID, "", "", "", nil)

	mockSvc := NewUserServiceInterfaceMock(t)
	mockSvc.On("UpdateUserCredentials", userID, "password", mock.Anything).
		Return(&ErrorUserNotFound)

	hashMock := hashmock.NewHashServiceInterfaceMock(t)
	hashMock.On("Generate", []byte("Secret123!")).Return(hash.Credential{
		Algorithm: hash.PBKDF2,
		Hash:      "hashed-password",
		Parameters: hash.CredParameters{
			Salt:       "salt123",
			Iterations: 10000,
			KeySize:    32,
		},
	}, nil).Once()

	handler := newUserHandler(mockSvc, hashMock)
	req := httptest.NewRequest(http.MethodPost, "/users/me/update-credentials",
		bytes.NewBufferString(`{"attributes":{"password":"Secret123!"}}`))
	req = req.WithContext(security.WithSecurityContextTest(req.Context(), authCtx))
	rr := httptest.NewRecorder()

	handler.HandleSelfUserCredentialUpdateRequest(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)

	var errResp apierror.ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&errResp))
	require.Equal(t, ErrorUserNotFound.Code, errResp.Code)
}
