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
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
)

type UserInfoHandlerTestSuite struct {
	suite.Suite
	mockService *userInfoServiceInterfaceMock
	handler     *userInfoHandler
}

func TestUserInfoHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(UserInfoHandlerTestSuite))
}

func (s *UserInfoHandlerTestSuite) SetupTest() {
	s.mockService = new(userInfoServiceInterfaceMock)
	s.handler = newUserInfoHandler(s.mockService)
}

// TestHandleUserInfo_MissingAuthorizationHeader tests missing Authorization header
func (s *UserInfoHandlerTestSuite) TestHandleUserInfo_MissingAuthorizationHeader() {
	req := httptest.NewRequest(http.MethodGet, "/oauth2/userinfo", nil)
	rr := httptest.NewRecorder()

	s.handler.HandleUserInfo(rr, req)

	assert.Equal(s.T(), http.StatusUnauthorized, rr.Code)
	assert.Contains(s.T(), rr.Body.String(), constants.ErrorInvalidRequest)
	assert.Contains(s.T(), rr.Body.String(), "missing Authorization header")
}

// TestHandleUserInfo_InvalidAuthorizationHeaderFormat tests invalid Authorization header format
func (s *UserInfoHandlerTestSuite) TestHandleUserInfo_InvalidAuthorizationHeaderFormat() {
	req := httptest.NewRequest(http.MethodGet, "/oauth2/userinfo", nil)
	req.Header.Set("Authorization", "InvalidFormat token123")
	rr := httptest.NewRecorder()

	s.handler.HandleUserInfo(rr, req)

	assert.Equal(s.T(), http.StatusUnauthorized, rr.Code)
	assert.Contains(s.T(), rr.Body.String(), constants.ErrorInvalidRequest)
	assert.Contains(s.T(), rr.Body.String(), "invalid Authorization header format")
}

// TestHandleUserInfo_MissingBearerToken tests missing Bearer token
func (s *UserInfoHandlerTestSuite) TestHandleUserInfo_MissingBearerToken() {
	req := httptest.NewRequest(http.MethodGet, "/oauth2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer ")
	rr := httptest.NewRecorder()

	s.handler.HandleUserInfo(rr, req)

	assert.Equal(s.T(), http.StatusUnauthorized, rr.Code)
	assert.Contains(s.T(), rr.Body.String(), constants.ErrorInvalidRequest)
	assert.Contains(s.T(), rr.Body.String(), "missing access token")
}

// TestHandleUserInfo_InvalidToken tests invalid token error
func (s *UserInfoHandlerTestSuite) TestHandleUserInfo_InvalidToken() {
	req := httptest.NewRequest(http.MethodGet, "/oauth2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rr := httptest.NewRecorder()

	s.mockService.On("GetUserInfo", "invalid-token").Return(nil, &errorInvalidAccessToken)

	s.handler.HandleUserInfo(rr, req)

	assert.Equal(s.T(), http.StatusUnauthorized, rr.Code)
	assert.Contains(s.T(), rr.Body.String(), errorInvalidAccessToken.Code)
	assert.Contains(s.T(), rr.Body.String(), errorInvalidAccessToken.ErrorDescription)
	s.mockService.AssertExpectations(s.T())
}

// TestHandleUserInfo_MissingSubClaim tests missing sub claim error
func (s *UserInfoHandlerTestSuite) TestHandleUserInfo_MissingSubClaim() {
	req := httptest.NewRequest(http.MethodGet, "/oauth2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer token123")
	rr := httptest.NewRecorder()

	s.mockService.On("GetUserInfo", "token123").Return(nil, &errorMissingSubClaim)

	s.handler.HandleUserInfo(rr, req)

	assert.Equal(s.T(), http.StatusUnauthorized, rr.Code)
	assert.Contains(s.T(), rr.Body.String(), errorMissingSubClaim.Code)
	assert.Contains(s.T(), rr.Body.String(), errorMissingSubClaim.ErrorDescription)
	s.mockService.AssertExpectations(s.T())
}

// TestHandleUserInfo_ServerError tests server error
func (s *UserInfoHandlerTestSuite) TestHandleUserInfo_ServerError() {
	req := httptest.NewRequest(http.MethodGet, "/oauth2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer token123")
	rr := httptest.NewRecorder()

	expectedError := serviceerror.CustomServiceError(serviceerror.InternalServerError,
		"An error occurred while fetching user attributes or groups")
	s.mockService.On("GetUserInfo", "token123").Return(nil, expectedError)

	s.handler.HandleUserInfo(rr, req)

	assert.Equal(s.T(), http.StatusInternalServerError, rr.Code)
	assert.Contains(s.T(), rr.Body.String(), expectedError.Code)
	assert.Contains(s.T(), rr.Body.String(), expectedError.ErrorDescription)
	s.mockService.AssertExpectations(s.T())
}

// TestHandleUserInfo_Success tests successful response
func (s *UserInfoHandlerTestSuite) TestHandleUserInfo_Success() {
	req := httptest.NewRequest(http.MethodGet, "/oauth2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()

	userInfo := map[string]interface{}{
		"sub":   "user123",
		"name":  "John Doe",
		"email": "john@example.com",
	}

	s.mockService.On("GetUserInfo", "valid-token").Return(jsonResponse(userInfo), nil)

	s.handler.HandleUserInfo(rr, req)

	assert.Equal(s.T(), http.StatusOK, rr.Code)
	assert.Equal(s.T(), "application/json", rr.Header().Get("Content-Type"))
	assert.Equal(s.T(), "no-store", rr.Header().Get("Cache-Control"))
	assert.Equal(s.T(), "no-cache", rr.Header().Get("Pragma"))
	assert.Contains(s.T(), rr.Body.String(), `"sub":"user123"`)
	assert.Contains(s.T(), rr.Body.String(), `"name":"John Doe"`)
	assert.Contains(s.T(), rr.Body.String(), `"email":"john@example.com"`)
	s.mockService.AssertExpectations(s.T())
}

// TestHandleUserInfo_Success_POST tests successful POST request
func (s *UserInfoHandlerTestSuite) TestHandleUserInfo_Success_POST() {
	req := httptest.NewRequest(http.MethodPost, "/oauth2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()

	userInfo := map[string]interface{}{
		"sub": "user123",
	}

	s.mockService.On("GetUserInfo", "valid-token").Return(jsonResponse(userInfo), nil)

	s.handler.HandleUserInfo(rr, req)

	assert.Equal(s.T(), http.StatusOK, rr.Code)
	assert.Contains(s.T(), rr.Body.String(), `"sub":"user123"`)
	s.mockService.AssertExpectations(s.T())
}

// TestHandleUserInfo_Success_WithGroups tests successful response with groups
func (s *UserInfoHandlerTestSuite) TestHandleUserInfo_Success_WithGroups() {
	req := httptest.NewRequest(http.MethodGet, "/oauth2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()

	userInfo := map[string]interface{}{
		"sub":    "user123",
		"name":   "John Doe",
		"groups": []interface{}{"admin", "users"},
	}

	s.mockService.On("GetUserInfo", "valid-token").Return(jsonResponse(userInfo), nil)

	s.handler.HandleUserInfo(rr, req)

	assert.Equal(s.T(), http.StatusOK, rr.Code)
	assert.Contains(s.T(), rr.Body.String(), `"sub":"user123"`)
	assert.Contains(s.T(), rr.Body.String(), `"name":"John Doe"`)
	assert.Contains(s.T(), rr.Body.String(), `"groups"`)
	s.mockService.AssertExpectations(s.T())
}

// TestHandleUserInfo_CaseInsensitiveBearer tests case-insensitive Bearer token
func (s *UserInfoHandlerTestSuite) TestHandleUserInfo_CaseInsensitiveBearer() {
	req := httptest.NewRequest(http.MethodGet, "/oauth2/userinfo", nil)
	req.Header.Set("Authorization", "bearer valid-token")
	rr := httptest.NewRecorder()

	userInfo := map[string]interface{}{
		"sub": "user123",
	}

	s.mockService.On("GetUserInfo", "valid-token").Return(jsonResponse(userInfo), nil)

	s.handler.HandleUserInfo(rr, req)

	assert.Equal(s.T(), http.StatusOK, rr.Code)
	s.mockService.AssertExpectations(s.T())
}

// TestHandleUserInfo_BEARERUpperCase tests BEARER in uppercase
func (s *UserInfoHandlerTestSuite) TestHandleUserInfo_BEARERUpperCase() {
	req := httptest.NewRequest(http.MethodGet, "/oauth2/userinfo", nil)
	req.Header.Set("Authorization", "BEARER valid-token")
	rr := httptest.NewRecorder()

	userInfo := map[string]interface{}{
		"sub": "user123",
	}

	s.mockService.On("GetUserInfo", "valid-token").Return(jsonResponse(userInfo), nil)

	s.handler.HandleUserInfo(rr, req)

	assert.Equal(s.T(), http.StatusOK, rr.Code)
	s.mockService.AssertExpectations(s.T())
}

// TestHandleUserInfo_EmptyResponse tests empty response
func (s *UserInfoHandlerTestSuite) TestHandleUserInfo_EmptyResponse() {
	req := httptest.NewRequest(http.MethodGet, "/oauth2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()

	userInfo := map[string]interface{}{
		"sub": "user123",
	}

	s.mockService.On("GetUserInfo", "valid-token").Return(jsonResponse(userInfo), nil)

	s.handler.HandleUserInfo(rr, req)

	assert.Equal(s.T(), http.StatusOK, rr.Code)
	assert.Contains(s.T(), rr.Body.String(), `"sub":"user123"`)
	s.mockService.AssertExpectations(s.T())
}

// TestHandleUserInfo_InvalidAuthorizationHeaderSinglePart tests invalid Authorization header with only one part
func (s *UserInfoHandlerTestSuite) TestHandleUserInfo_InvalidAuthorizationHeaderSinglePart() {
	req := httptest.NewRequest(http.MethodGet, "/oauth2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer")
	rr := httptest.NewRecorder()

	s.handler.HandleUserInfo(rr, req)

	assert.Equal(s.T(), http.StatusUnauthorized, rr.Code)
	assert.Contains(s.T(), rr.Body.String(), constants.ErrorInvalidRequest)
	assert.Contains(s.T(), rr.Body.String(), "invalid Authorization header format")
}

// TestHandleUserInfo_EncodingError tests encoding error handling
func (s *UserInfoHandlerTestSuite) TestHandleUserInfo_EncodingError() {
	req := httptest.NewRequest(http.MethodGet, "/oauth2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()

	// Use a function which cannot be JSON encoded - this will cause Encode to return an error
	userInfo := map[string]interface{}{
		"sub":  "user123",
		"name": "John Doe",
		"func": func() {}, // Function cannot be JSON encoded and will cause an error
	}

	s.mockService.On("GetUserInfo", "valid-token").Return(jsonResponse(userInfo), nil)

	s.handler.HandleUserInfo(rr, req)

	// With buffer approach, encoding fails BEFORE headers are sent, so we get HTTP 500
	assert.Equal(s.T(), http.StatusInternalServerError, rr.Code)
	// Verify that encoding error message is returned
	assert.Contains(s.T(), rr.Body.String(), serviceerror.ErrorEncodingError)
	s.mockService.AssertExpectations(s.T())
}

// TestWriteServiceErrorResponse_DefaultCase tests the default case in writeServiceErrorResponse
func (s *UserInfoHandlerTestSuite) TestWriteServiceErrorResponse_DefaultCase() {
	req := httptest.NewRequest(http.MethodGet, "/oauth2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer token123")
	rr := httptest.NewRecorder()

	// Create a service error with an unknown type (not ClientErrorType or ServerErrorType)
	unknownError := &serviceerror.ServiceError{
		Type:             "UnknownErrorType", // Unknown type
		Code:             "unknown_error",
		ErrorDescription: "An unknown error occurred",
	}
	s.mockService.On("GetUserInfo", "token123").Return(nil, unknownError)

	s.handler.HandleUserInfo(rr, req)

	// Default case should return StatusUnauthorized
	assert.Equal(s.T(), http.StatusUnauthorized, rr.Code)
	assert.Contains(s.T(), rr.Body.String(), "unknown_error")
	assert.Contains(s.T(), rr.Body.String(), "An unknown error occurred")
	s.mockService.AssertExpectations(s.T())
}

func jsonResponse(body map[string]interface{}) *UserInfoResponse {
	return &UserInfoResponse{
		JSONBody: body,
	}
}
