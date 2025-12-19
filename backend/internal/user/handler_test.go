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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/asgardeo/thunder/internal/system/error/apierror"
	"github.com/asgardeo/thunder/internal/system/security"
)

func TestHandleSelfUserGetRequest_Success(t *testing.T) {
	userID := "user-123"
	authCtx := security.NewSecurityContextForTest(userID, "", "", "", nil)

	mockSvc := NewUserServiceInterfaceMock(t)
	expectedUser := &User{
		ID:         userID,
		Attributes: json.RawMessage(`{"username":"alice"}`),
	}
	mockSvc.On("GetUser", mock.Anything, userID).Return(expectedUser, nil)

	handler := newUserHandler(mockSvc)
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
	handler := newUserHandler(mockSvc)
	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	rr := httptest.NewRecorder()

	handler.HandleSelfUserGetRequest(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)

	var errResp apierror.ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&errResp))
	require.Equal(t, ErrorAuthenticationFailed.Code, errResp.Code)
}

func TestHandleSelfUserPutRequest_Success(t *testing.T) {
	userID := "user-456"
	authCtx := security.NewSecurityContextForTest(userID, "", "", "", nil)
	attributes := json.RawMessage(`{"email":"alice@example.com"}`)

	mockSvc := NewUserServiceInterfaceMock(t)
	updatedUser := &User{
		ID:         userID,
		Type:       "employee",
		Attributes: attributes,
	}
	mockSvc.On("UpdateUserAttributes", mock.Anything, userID, attributes).Return(updatedUser, nil)

	handler := newUserHandler(mockSvc)
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
	userID := "user-456"
	authCtx := security.NewSecurityContextForTest(userID, "", "", "", nil)

	mockSvc := NewUserServiceInterfaceMock(t)
	handler := newUserHandler(mockSvc)

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
	userID := "user-789"
	authCtx := security.NewSecurityContextForTest(userID, "", "", "", nil)
	attrs := json.RawMessage(`{"password":"Secret123!"}`)

	mockSvc := NewUserServiceInterfaceMock(t)
	mockSvc.On("UpdateUserCredentials", mock.Anything, userID, attrs).Return(nil)

	handler := newUserHandler(mockSvc)
	req := httptest.NewRequest(http.MethodPost, "/users/me/update-credentials",
		bytes.NewBufferString(`{"attributes":{"password":"Secret123!"}}`))
	req = req.WithContext(security.WithSecurityContextTest(req.Context(), authCtx))
	rr := httptest.NewRecorder()

	handler.HandleSelfUserCredentialUpdateRequest(rr, req)

	require.Equal(t, http.StatusNoContent, rr.Code)
	require.Equal(t, 0, rr.Body.Len())
}

func TestHandleSelfUserCredentialUpdateRequest_MissingCredentials(t *testing.T) {
	userID := "user-789"
	authCtx := security.NewSecurityContextForTest(userID, "", "", "", nil)

	mockSvc := NewUserServiceInterfaceMock(t)
	mockSvc.On("UpdateUserCredentials", mock.Anything, userID, mock.Anything).Return(&ErrorMissingCredentials)
	handler := newUserHandler(mockSvc)

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
