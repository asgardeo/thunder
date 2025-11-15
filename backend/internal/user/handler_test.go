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
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/config"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/error/apierror"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
)

const (
	defaultUserID   = "user-123"
	defaultUserPath = "root"
	userTypeBasic   = "basic"
)

type UserHandlerTestSuite struct {
	suite.Suite
}

func TestUserHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(UserHandlerTestSuite))
}

func (suite *UserHandlerTestSuite) SetupTest() {
	suite.ensureRuntime()
}

func (suite *UserHandlerTestSuite) TearDownTest() {
	config.ResetThunderRuntime()
}

func (suite *UserHandlerTestSuite) ensureRuntime() {
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("", &config.Config{})
	suite.Require().NoError(err)
}

type userHandlerTestCase struct {
	name          string
	method        string
	url           string
	body          string
	pathValues    map[string]string
	useFlaky      bool
	setJSONHeader bool
	setup         func(*UserServiceInterfaceMock)
	assert        func(*httptest.ResponseRecorder)
	assertService func(*UserServiceInterfaceMock)
}

type flakyResponseWriter struct {
	*httptest.ResponseRecorder
	failNext bool
}

func newFlakyResponseWriter() *flakyResponseWriter {
	return &flakyResponseWriter{
		ResponseRecorder: httptest.NewRecorder(),
		failNext:         true,
	}
}

func (w *flakyResponseWriter) Write(b []byte) (int, error) {
	if w.failNext {
		w.failNext = false
		return 0, errors.New("write failure")
	}
	return w.ResponseRecorder.Write(b)
}

func (suite *UserHandlerTestSuite) runHandlerTestCases(
	testCases []userHandlerTestCase,
	invoke func(*userHandler, http.ResponseWriter, *http.Request),
) {
	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			serviceMock := NewUserServiceInterfaceMock(suite.T())
			handler := newUserHandler(serviceMock)

			method := tc.method
			if method == "" {
				method = http.MethodGet
			}

			var bodyReader io.Reader
			if tc.body != "" {
				bodyReader = strings.NewReader(tc.body)
			}

			req := httptest.NewRequest(method, tc.url, bodyReader)
			if tc.setJSONHeader {
				req.Header.Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
			}

			for key, value := range tc.pathValues {
				req.SetPathValue(key, value)
			}

			var writer http.ResponseWriter
			var recorder *httptest.ResponseRecorder
			if tc.useFlaky {
				flaky := newFlakyResponseWriter()
				writer = flaky
				recorder = flaky.ResponseRecorder
			} else {
				recorder = httptest.NewRecorder()
				writer = recorder
			}

			if tc.setup != nil {
				tc.setup(serviceMock)
			}

			invoke(handler, writer, req)

			if tc.assert != nil {
				tc.assert(recorder)
			}

			if tc.assertService != nil {
				tc.assertService(serviceMock)
			} else {
				serviceMock.AssertExpectations(suite.T())
			}
		})
	}
}

func (suite *UserHandlerTestSuite) TestUserHandler_NewUserHandler() {
	handler := newUserHandler(nil)
	suite.NotNil(handler)
}

func (suite *UserHandlerTestSuite) TestUserHandler_RegisterRoutes() {
	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		setJSON    bool
		setup      func(*UserServiceInterfaceMock)
		wantStatus int
		assertFunc func(*httptest.ResponseRecorder)
	}{
		{
			name:       "options root",
			method:     http.MethodOptions,
			path:       "/users",
			wantStatus: http.StatusNoContent,
		},
		{
			name:   "list dispatch",
			method: http.MethodGet,
			path:   "/users",
			setup: func(serviceMock *UserServiceInterfaceMock) {
				serviceMock.
					On("GetUserList", serverconst.DefaultPageSize, 0, mock.MatchedBy(func(filters map[string]interface{}) bool {
						return len(filters) == 0
					})).
					Return(&UserListResponse{TotalResults: 1, Count: 1, Users: []User{{ID: defaultUserID}}}, nil).
					Once()
			},
			wantStatus: http.StatusOK,
		},
		{
			name:    "post dispatch",
			method:  http.MethodPost,
			path:    "/users",
			body:    `{"type":"basic"}`,
			setJSON: true,
			setup: func(serviceMock *UserServiceInterfaceMock) {
				serviceMock.
					On("CreateUser", mock.MatchedBy(func(u *User) bool { return u.Type == userTypeBasic })).
					Return(&User{ID: defaultUserID}, nil).
					Once()
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:   "get dispatch",
			method: http.MethodGet,
			path:   "/users/" + defaultUserID,
			setup: func(serviceMock *UserServiceInterfaceMock) {
				serviceMock.
					On("GetUser", defaultUserID).
					Return(&User{ID: defaultUserID}, nil).
					Once()
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "user groups dispatch",
			method: http.MethodGet,
			path:   "/users/" + defaultUserID + "/groups",
			setup: func(serviceMock *UserServiceInterfaceMock) {
				serviceMock.
					On("GetUserGroups", defaultUserID, serverconst.DefaultPageSize, 0).
					Return(&UserGroupListResponse{TotalResults: 0}, nil).
					Once()
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "tree list dispatch",
			method: http.MethodGet,
			path:   "/users/tree/" + defaultUserPath,
			setup: func(serviceMock *UserServiceInterfaceMock) {
				serviceMock.
					On("GetUsersByPath", defaultUserPath, serverconst.DefaultPageSize, 0, mock.Anything).
					Return(&UserListResponse{TotalResults: 0}, nil).
					Once()
			},
			wantStatus: http.StatusOK,
		},
		{
			name:    "tree post dispatch",
			method:  http.MethodPost,
			path:    "/users/tree/" + defaultUserPath,
			body:    `{"type":"basic"}`,
			setJSON: true,
			setup: func(serviceMock *UserServiceInterfaceMock) {
				serviceMock.
					On("CreateUserByPath", defaultUserPath, mock.MatchedBy(func(req CreateUserByPathRequest) bool {
						return req.Type == userTypeBasic
					})).
					Return(&User{ID: defaultUserID}, nil).
					Once()
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "unknown subresource",
			method:     http.MethodGet,
			path:       "/users/" + defaultUserID + "/unknown",
			wantStatus: http.StatusNotFound,
			assertFunc: func(recorder *httptest.ResponseRecorder) {
				suite.Equal("404 page not found\n", recorder.Body.String())
			},
		},
		{
			name:       "too many segments",
			method:     http.MethodGet,
			path:       "/users/" + defaultUserID + "/foo/bar",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "options user route",
			method:     http.MethodOptions,
			path:       "/users/" + defaultUserID,
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "tree options route",
			method:     http.MethodOptions,
			path:       "/users/tree/" + defaultUserPath,
			wantStatus: http.StatusNoContent,
		},
	}

	for _, tc := range tests {
		tc := tc
		suite.Run(tc.name, func() {
			serviceMock := NewUserServiceInterfaceMock(suite.T())
			handler := newUserHandler(serviceMock)
			mux := http.NewServeMux()
			registerRoutes(mux, handler)

			if tc.setup != nil {
				tc.setup(serviceMock)
			}

			var bodyReader io.Reader
			if tc.body != "" {
				bodyReader = strings.NewReader(tc.body)
			}

			req := httptest.NewRequest(tc.method, tc.path, bodyReader)
			if tc.setJSON {
				req.Header.Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
			}

			resp := httptest.NewRecorder()
			mux.ServeHTTP(resp, req)

			suite.Equal(tc.wantStatus, resp.Code)
			if tc.assertFunc != nil {
				tc.assertFunc(resp)
			}

			serviceMock.AssertExpectations(suite.T())
		})
	}
}

func (suite *UserHandlerTestSuite) TestUserHandler_HandleUserListRequest() {
	suite.runHandlerTestCases(
		[]userHandlerTestCase{
			{
				name: "success",
				url:  "/users?limit=3&offset=2&filter=" + url.QueryEscape(`type eq "basic"`),
				setup: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.
						On("GetUserList", 3, 2, mock.MatchedBy(func(filters map[string]interface{}) bool {
							return len(filters) == 1 && filters["type"] == userTypeBasic
						})).
						Return(&UserListResponse{
							TotalResults: 2,
							StartIndex:   1,
							Count:        2,
							Users: []User{
								{ID: "user-1"},
								{ID: "user-2"},
							},
						}, nil).
						Once()
				},
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusOK, recorder.Code)
					suite.Equal(serverconst.ContentTypeJSON, recorder.Header().Get(serverconst.ContentTypeHeaderName))
					var resp UserListResponse
					suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &resp))
					suite.Equal(2, resp.TotalResults)
					suite.Len(resp.Users, 2)
				},
			},
			{
				name: "default limit applied",
				url:  "/users?offset=1",
				setup: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.
						On("GetUserList", serverconst.DefaultPageSize, 1, mock.Anything).
						Return(&UserListResponse{}, nil).
						Once()
				},
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusOK, recorder.Code)
				},
			},
			{
				name: "invalid limit",
				url:  "/users?limit=abc",
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusBadRequest, recorder.Code)
					var resp apierror.ErrorResponse
					suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &resp))
					suite.Equal(ErrorInvalidLimit.Code, resp.Code)
				},
				assertService: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.AssertNotCalled(suite.T(), "GetUserList", mock.Anything, mock.Anything, mock.Anything)
				},
			},
			{
				name: "invalid offset",
				url:  "/users?offset=abc",
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusBadRequest, recorder.Code)
					var resp apierror.ErrorResponse
					suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &resp))
					suite.Equal(ErrorInvalidOffset.Code, resp.Code)
				},
				assertService: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.AssertNotCalled(suite.T(), "GetUserList", mock.Anything, mock.Anything, mock.Anything)
				},
			},
			{
				name: "filter parse error",
				url:  "/users?filter=invalid",
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusBadRequest, recorder.Code)
					var resp apierror.ErrorResponse
					suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &resp))
					suite.Equal(ErrorInvalidFilter.Code, resp.Code)
				},
				assertService: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.AssertNotCalled(suite.T(), "GetUserList", mock.Anything, mock.Anything, mock.Anything)
				},
			},
			{
				name: "service error",
				url:  "/users",
				setup: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.
						On("GetUserList", serverconst.DefaultPageSize, 0, mock.Anything).
						Return((*UserListResponse)(nil), &ErrorInternalServerError).
						Once()
				},
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusInternalServerError, recorder.Code)
					var resp apierror.ErrorResponse
					suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &resp))
					suite.Equal(ErrorInternalServerError.Code, resp.Code)
				},
			},
			{
				name:     "encode error",
				url:      "/users",
				useFlaky: true,
				setup: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.
						On("GetUserList", serverconst.DefaultPageSize, 0, mock.Anything).
						Return(&UserListResponse{}, nil).
						Once()
				},
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusOK, recorder.Code)
					suite.Equal("Failed to encode response\n", recorder.Body.String())
				},
			},
		},
		func(handler *userHandler, w http.ResponseWriter, r *http.Request) {
			handler.HandleUserListRequest(w, r)
		},
	)
}

func (suite *UserHandlerTestSuite) TestUserHandler_HandleUserPostRequest() {
	suite.runHandlerTestCases(
		[]userHandlerTestCase{
			{
				name:          "success",
				method:        http.MethodPost,
				url:           "/users",
				body:          `{"type":"basic"}`,
				setJSONHeader: true,
				setup: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.
						On("CreateUser", mock.MatchedBy(func(u *User) bool { return u.Type == userTypeBasic })).
						Return(&User{ID: defaultUserID}, nil).
						Once()
				},
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusCreated, recorder.Code)
					var resp User
					suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &resp))
					suite.Equal(defaultUserID, resp.ID)
				},
			},
			{
				name:          "invalid body",
				method:        http.MethodPost,
				url:           "/users",
				body:          "{invalid",
				setJSONHeader: true,
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusBadRequest, recorder.Code)
					suite.Contains(recorder.Body.String(), "Bad Request")
				},
				assertService: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.AssertNotCalled(suite.T(), "CreateUser", mock.Anything)
				},
			},
			{
				name:          "invalid body encode error",
				method:        http.MethodPost,
				url:           "/users",
				body:          "{invalid",
				setJSONHeader: true,
				useFlaky:      true,
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusBadRequest, recorder.Code)
					suite.Equal("", recorder.Body.String())
				},
				assertService: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.AssertNotCalled(suite.T(), "CreateUser", mock.Anything)
				},
			},
			{
				name:          "service conflict",
				method:        http.MethodPost,
				url:           "/users",
				body:          `{"type":"basic"}`,
				setJSONHeader: true,
				setup: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.
						On("CreateUser", mock.AnythingOfType("*user.User")).
						Return((*User)(nil), &ErrorAttributeConflict).
						Once()
				},
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusConflict, recorder.Code)
					var resp apierror.ErrorResponse
					suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &resp))
					suite.Equal(ErrorAttributeConflict.Code, resp.Code)
				},
			},
			{
				name:          "service error encode failure",
				method:        http.MethodPost,
				url:           "/users",
				body:          `{"type":"basic"}`,
				setJSONHeader: true,
				useFlaky:      true,
				setup: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.
						On("CreateUser", mock.AnythingOfType("*user.User")).
						Return((*User)(nil), &ErrorInternalServerError).
						Once()
				},
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusInternalServerError, recorder.Code)
					suite.Equal("Failed to encode error response\n", recorder.Body.String())
				},
			},
			{
				name:          "encode error",
				method:        http.MethodPost,
				url:           "/users",
				body:          `{"type":"basic"}`,
				setJSONHeader: true,
				useFlaky:      true,
				setup: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.
						On("CreateUser", mock.AnythingOfType("*user.User")).
						Return(&User{ID: defaultUserID}, nil).
						Once()
				},
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusCreated, recorder.Code)
					suite.Equal("Internal Server Error\n", recorder.Body.String())
				},
			},
		},
		func(handler *userHandler, w http.ResponseWriter, r *http.Request) {
			handler.HandleUserPostRequest(w, r)
		},
	)
}

func (suite *UserHandlerTestSuite) TestUserHandler_HandleUserGetRequest() {
	suite.runHandlerTestCases(
		[]userHandlerTestCase{
			{
				name: "missing id",
				url:  "/users/" + defaultUserID,
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusBadRequest, recorder.Code)
					suite.Contains(recorder.Body.String(), "Missing user id")
				},
				assertService: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.AssertNotCalled(suite.T(), "GetUser", mock.Anything)
				},
			},
			{
				name: "not found",
				url:  "/users/" + defaultUserID,
				pathValues: map[string]string{
					"id": defaultUserID,
				},
				setup: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.
						On("GetUser", defaultUserID).
						Return((*User)(nil), &ErrorUserNotFound).
						Once()
				},
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusNotFound, recorder.Code)
					var resp apierror.ErrorResponse
					suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &resp))
					suite.Equal(ErrorUserNotFound.Code, resp.Code)
				},
			},
			{
				name: "service error",
				url:  "/users/" + defaultUserID,
				pathValues: map[string]string{
					"id": defaultUserID,
				},
				setup: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.
						On("GetUser", defaultUserID).
						Return((*User)(nil), &ErrorInternalServerError).
						Once()
				},
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusInternalServerError, recorder.Code)
					var resp apierror.ErrorResponse
					suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &resp))
					suite.Equal(ErrorInternalServerError.Code, resp.Code)
				},
			},
			{
				name: "encode error",
				url:  "/users/" + defaultUserID,
				pathValues: map[string]string{
					"id": defaultUserID,
				},
				useFlaky: true,
				setup: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.
						On("GetUser", defaultUserID).
						Return(&User{ID: defaultUserID}, nil).
						Once()
				},
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusInternalServerError, recorder.Code)
					suite.Equal("Internal Server Error\n", recorder.Body.String())
				},
			},
		},
		func(handler *userHandler, w http.ResponseWriter, r *http.Request) {
			handler.HandleUserGetRequest(w, r)
		},
	)
}

func (suite *UserHandlerTestSuite) TestUserHandler_HandleUserGroupsGetRequest() {
	suite.runHandlerTestCases(
		[]userHandlerTestCase{
			{
				name: "success",
				url:  "/users/" + defaultUserID + "/groups?limit=2&offset=1",
				pathValues: map[string]string{
					"id": defaultUserID,
				},
				setup: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.
						On("GetUserGroups", defaultUserID, 2, 1).
						Return(&UserGroupListResponse{
							TotalResults: 1,
							Groups: []UserGroup{
								{ID: "group-1"},
							},
						}, nil).
						Once()
				},
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusOK, recorder.Code)
					var resp UserGroupListResponse
					suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &resp))
					suite.Equal(1, resp.TotalResults)
					suite.Len(resp.Groups, 1)
				},
			},
			{
				name: "missing id",
				url:  "/users/" + defaultUserID + "/groups",
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusNotFound, recorder.Code)
					var resp apierror.ErrorResponse
					suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &resp))
					suite.Equal(ErrorMissingUserID.Code, resp.Code)
				},
				assertService: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.AssertNotCalled(suite.T(), "GetUserGroups", mock.Anything, mock.Anything, mock.Anything)
				},
			},
			{
				name: "invalid pagination",
				url:  "/users/" + defaultUserID + "/groups?limit=abc",
				pathValues: map[string]string{
					"id": defaultUserID,
				},
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusBadRequest, recorder.Code)
					var resp apierror.ErrorResponse
					suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &resp))
					suite.Equal(ErrorInvalidLimit.Code, resp.Code)
				},
				assertService: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.AssertNotCalled(suite.T(), "GetUserGroups", mock.Anything, mock.Anything, mock.Anything)
				},
			},
			{
				name: "service error",
				url:  "/users/" + defaultUserID + "/groups",
				pathValues: map[string]string{
					"id": defaultUserID,
				},
				setup: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.
						On("GetUserGroups", defaultUserID, serverconst.DefaultPageSize, 0).
						Return((*UserGroupListResponse)(nil), &ErrorInternalServerError).
						Once()
				},
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusInternalServerError, recorder.Code)
					var resp apierror.ErrorResponse
					suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &resp))
					suite.Equal(ErrorInternalServerError.Code, resp.Code)
				},
			},
			{
				name: "encode error",
				url:  "/users/" + defaultUserID + "/groups",
				pathValues: map[string]string{
					"id": defaultUserID,
				},
				useFlaky: true,
				setup: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.
						On("GetUserGroups", defaultUserID, serverconst.DefaultPageSize, 0).
						Return(&UserGroupListResponse{}, nil).
						Once()
				},
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusOK, recorder.Code)
					suite.Equal("Failed to encode response\n", recorder.Body.String())
				},
			},
		},
		func(handler *userHandler, w http.ResponseWriter, r *http.Request) {
			handler.HandleUserGroupsGetRequest(w, r)
		},
	)
}

func (suite *UserHandlerTestSuite) TestUserHandler_HandleUserPutRequest() {
	suite.runHandlerTestCases(
		[]userHandlerTestCase{
			{
				name:          "success",
				method:        http.MethodPut,
				url:           "/users/" + defaultUserID,
				body:          `{"type":"basic"}`,
				setJSONHeader: true,
				setup: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.
						On("UpdateUser", defaultUserID, mock.MatchedBy(func(u *User) bool {
							return u.ID == defaultUserID && u.Type == userTypeBasic
						})).
						Return(&User{ID: defaultUserID, Type: userTypeBasic}, nil).
						Once()
				},
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusOK, recorder.Code)
					var resp User
					suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &resp))
					suite.Equal(defaultUserID, resp.ID)
				},
			},
			{
				name:          "missing id",
				method:        http.MethodPut,
				url:           "/users/",
				body:          `{"type":"basic"}`,
				setJSONHeader: true,
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusBadRequest, recorder.Code)
					suite.Contains(recorder.Body.String(), "Missing user id")
				},
				assertService: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.AssertNotCalled(suite.T(), "UpdateUser", mock.Anything, mock.Anything)
				},
			},
			{
				name:          "invalid body",
				method:        http.MethodPut,
				url:           "/users/" + defaultUserID,
				body:          "{invalid",
				setJSONHeader: true,
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusBadRequest, recorder.Code)
					suite.Contains(recorder.Body.String(), "Bad Request")
				},
				assertService: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.AssertNotCalled(suite.T(), "UpdateUser", mock.Anything, mock.Anything)
				},
			},
			{
				name:          "service not found",
				method:        http.MethodPut,
				url:           "/users/" + defaultUserID,
				body:          `{"type":"basic"}`,
				setJSONHeader: true,
				setup: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.
						On("UpdateUser", defaultUserID, mock.AnythingOfType("*user.User")).
						Return((*User)(nil), &ErrorUserNotFound).
						Once()
				},
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusNotFound, recorder.Code)
					var resp apierror.ErrorResponse
					suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &resp))
					suite.Equal(ErrorUserNotFound.Code, resp.Code)
				},
			},
			{
				name:          "encode error",
				method:        http.MethodPut,
				url:           "/users/" + defaultUserID,
				body:          `{"type":"basic"}`,
				setJSONHeader: true,
				useFlaky:      true,
				setup: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.
						On("UpdateUser", defaultUserID, mock.AnythingOfType("*user.User")).
						Return(&User{ID: defaultUserID}, nil).
						Once()
				},
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusInternalServerError, recorder.Code)
					suite.Equal("Internal Server Error\n", recorder.Body.String())
				},
			},
		},
		func(handler *userHandler, w http.ResponseWriter, r *http.Request) {
			handler.HandleUserPutRequest(w, r)
		},
	)
}

func (suite *UserHandlerTestSuite) TestUserHandler_HandleUserDeleteRequest() {
	suite.runHandlerTestCases(
		[]userHandlerTestCase{
			{
				name:   "success",
				method: http.MethodDelete,
				url:    "/users/" + defaultUserID,
				setup: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.
						On("DeleteUser", defaultUserID).
						Return((*serviceerror.ServiceError)(nil)).
						Once()
				},
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusNoContent, recorder.Code)
				},
			},
			{
				name:   "missing id",
				method: http.MethodDelete,
				url:    "/users/",
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusBadRequest, recorder.Code)
					suite.Contains(recorder.Body.String(), "Missing user id")
				},
				assertService: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.AssertNotCalled(suite.T(), "DeleteUser", mock.Anything)
				},
			},
			{
				name:   "service error",
				method: http.MethodDelete,
				url:    "/users/" + defaultUserID,
				setup: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.
						On("DeleteUser", defaultUserID).
						Return(&ErrorUserNotFound).
						Once()
				},
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusNotFound, recorder.Code)
					var resp apierror.ErrorResponse
					suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &resp))
					suite.Equal(ErrorUserNotFound.Code, resp.Code)
				},
			},
		},
		func(handler *userHandler, w http.ResponseWriter, r *http.Request) {
			handler.HandleUserDeleteRequest(w, r)
		},
	)
}

func (suite *UserHandlerTestSuite) TestUserHandler_HandleUserListByPathRequest() {
	suite.runHandlerTestCases(
		[]userHandlerTestCase{
			{
				name: "success",
				url:  "/users/tree/" + defaultUserPath + "?limit=2&offset=1",
				pathValues: map[string]string{
					"path": defaultUserPath,
				},
				setup: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.
						On("GetUsersByPath", defaultUserPath, 2, 1, mock.MatchedBy(func(filters map[string]interface{}) bool {
							return len(filters) == 0
						})).
						Return(&UserListResponse{
							TotalResults: 1,
							Users:        []User{{ID: defaultUserID}},
						}, nil).
						Once()
				},
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusOK, recorder.Code)
					var resp UserListResponse
					suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &resp))
					suite.Equal(1, resp.TotalResults)
				},
			},
			{
				name: "missing path",
				url:  "/users/tree/",
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusBadRequest, recorder.Code)
					var resp apierror.ErrorResponse
					suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &resp))
					suite.Equal(ErrorHandlePathRequired.Code, resp.Code)
				},
			},
			{
				name: "invalid pagination",
				url:  "/users/tree/" + defaultUserPath + "?limit=abc",
				pathValues: map[string]string{
					"path": defaultUserPath,
				},
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusBadRequest, recorder.Code)
					var resp apierror.ErrorResponse
					suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &resp))
					suite.Equal(ErrorInvalidLimit.Code, resp.Code)
				},
				assertService: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.AssertNotCalled(
						suite.T(),
						"GetUsersByPath",
						mock.Anything,
						mock.Anything,
						mock.Anything,
						mock.Anything,
					)
				},
			},
			{
				name: "filter parse error",
				url:  "/users/tree/" + defaultUserPath + "?filter=invalid",
				pathValues: map[string]string{
					"path": defaultUserPath,
				},
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusBadRequest, recorder.Code)
					var resp apierror.ErrorResponse
					suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &resp))
					suite.Equal(ErrorInvalidFilter.Code, resp.Code)
				},
				assertService: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.AssertNotCalled(
						suite.T(),
						"GetUsersByPath",
						mock.Anything,
						mock.Anything,
						mock.Anything,
						mock.Anything,
					)
				},
			},
			{
				name: "service error",
				url:  "/users/tree/" + defaultUserPath,
				pathValues: map[string]string{
					"path": defaultUserPath,
				},
				setup: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.
						On("GetUsersByPath", defaultUserPath, serverconst.DefaultPageSize, 0, mock.Anything).
						Return((*UserListResponse)(nil), &ErrorInternalServerError).
						Once()
				},
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusInternalServerError, recorder.Code)
					var resp apierror.ErrorResponse
					suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &resp))
					suite.Equal(ErrorInternalServerError.Code, resp.Code)
				},
			},
			{
				name: "encode error",
				url:  "/users/tree/" + defaultUserPath,
				pathValues: map[string]string{
					"path": defaultUserPath,
				},
				useFlaky: true,
				setup: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.
						On("GetUsersByPath", defaultUserPath, serverconst.DefaultPageSize, 0, mock.Anything).
						Return(&UserListResponse{}, nil).
						Once()
				},
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusOK, recorder.Code)
					suite.Equal("Failed to encode response\n", recorder.Body.String())
				},
			},
		},
		func(handler *userHandler, w http.ResponseWriter, r *http.Request) {
			handler.HandleUserListByPathRequest(w, r)
		},
	)
}

func (suite *UserHandlerTestSuite) TestUserHandler_HandleUserPostByPathRequest() {
	suite.runHandlerTestCases(
		[]userHandlerTestCase{
			{
				name:          "success",
				method:        http.MethodPost,
				url:           "/users/tree/" + defaultUserPath,
				body:          `{"type":"basic"}`,
				setJSONHeader: true,
				pathValues: map[string]string{
					"path": defaultUserPath,
				},
				setup: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.
						On("CreateUserByPath", defaultUserPath, mock.MatchedBy(func(req CreateUserByPathRequest) bool {
							return req.Type == userTypeBasic
						})).
						Return(&User{ID: defaultUserID}, nil).
						Once()
				},
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusCreated, recorder.Code)
					var resp User
					suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &resp))
					suite.Equal(defaultUserID, resp.ID)
				},
			},
			{
				name:          "missing path",
				method:        http.MethodPost,
				url:           "/users/tree/",
				body:          `{"type":"basic"}`,
				setJSONHeader: true,
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusBadRequest, recorder.Code)
					var resp apierror.ErrorResponse
					suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &resp))
					suite.Equal(ErrorHandlePathRequired.Code, resp.Code)
				},
				assertService: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.AssertNotCalled(suite.T(), "CreateUserByPath", mock.Anything, mock.Anything)
				},
			},
			{
				name:          "bad body",
				method:        http.MethodPost,
				url:           "/users/tree/" + defaultUserPath,
				body:          "{invalid",
				setJSONHeader: true,
				pathValues: map[string]string{
					"path": defaultUserPath,
				},
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusBadRequest, recorder.Code)
					var resp apierror.ErrorResponse
					suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &resp))
					suite.Equal(ErrorInvalidRequestFormat.Code, resp.Code)
				},
				assertService: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.AssertNotCalled(suite.T(), "CreateUserByPath", mock.Anything, mock.Anything)
				},
			},
			{
				name:          "bad body encode failure",
				method:        http.MethodPost,
				url:           "/users/tree/" + defaultUserPath,
				body:          "{invalid",
				setJSONHeader: true,
				useFlaky:      true,
				pathValues: map[string]string{
					"path": defaultUserPath,
				},
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusBadRequest, recorder.Code)
					suite.Equal("Failed to encode error response\n", recorder.Body.String())
				},
			},
			{
				name:          "service error",
				method:        http.MethodPost,
				url:           "/users/tree/" + defaultUserPath,
				body:          `{"type":"basic"}`,
				setJSONHeader: true,
				pathValues: map[string]string{
					"path": defaultUserPath,
				},
				setup: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.
						On("CreateUserByPath", defaultUserPath, mock.AnythingOfType("user.CreateUserByPathRequest")).
						Return((*User)(nil), &ErrorInternalServerError).
						Once()
				},
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusInternalServerError, recorder.Code)
					var resp apierror.ErrorResponse
					suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &resp))
					suite.Equal(ErrorInternalServerError.Code, resp.Code)
				},
			},
			{
				name:          "service error encode failure",
				method:        http.MethodPost,
				url:           "/users/tree/" + defaultUserPath,
				body:          `{"type":"basic"}`,
				setJSONHeader: true,
				useFlaky:      true,
				pathValues: map[string]string{
					"path": defaultUserPath,
				},
				setup: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.
						On("CreateUserByPath", defaultUserPath, mock.AnythingOfType("user.CreateUserByPathRequest")).
						Return((*User)(nil), &ErrorInternalServerError).
						Once()
				},
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusInternalServerError, recorder.Code)
					suite.Equal("Failed to encode error response\n", recorder.Body.String())
				},
			},
			{
				name:          "encode error",
				method:        http.MethodPost,
				url:           "/users/tree/" + defaultUserPath,
				body:          `{"type":"basic"}`,
				setJSONHeader: true,
				useFlaky:      true,
				pathValues: map[string]string{
					"path": defaultUserPath,
				},
				setup: func(serviceMock *UserServiceInterfaceMock) {
					serviceMock.
						On("CreateUserByPath", defaultUserPath, mock.AnythingOfType("user.CreateUserByPathRequest")).
						Return(&User{ID: defaultUserID}, nil).
						Once()
				},
				assert: func(recorder *httptest.ResponseRecorder) {
					suite.Equal(http.StatusCreated, recorder.Code)
					suite.Equal("Failed to encode response\n", recorder.Body.String())
				},
			},
		},
		func(handler *userHandler, w http.ResponseWriter, r *http.Request) {
			handler.HandleUserPostByPathRequest(w, r)
		},
	)
}

func (suite *UserHandlerTestSuite) TestUserHandler_ParsePaginationParams() {
	tests := []struct {
		name       string
		values     url.Values
		wantLimit  int
		wantOffset int
		wantErr    *serviceerror.ServiceError
	}{
		{
			name:       "defaults",
			values:     url.Values{},
			wantLimit:  0,
			wantOffset: 0,
		},
		{
			name: "valid params",
			values: url.Values{
				"limit":  []string{"5"},
				"offset": []string{"2"},
			},
			wantLimit:  5,
			wantOffset: 2,
		},
		{
			name: "invalid limit",
			values: url.Values{
				"limit": []string{"abc"},
			},
			wantErr: &ErrorInvalidLimit,
		},
		{
			name: "zero limit",
			values: url.Values{
				"limit": []string{"0"},
			},
			wantErr: &ErrorInvalidLimit,
		},
		{
			name: "invalid offset",
			values: url.Values{
				"offset": []string{"-1"},
			},
			wantErr: &ErrorInvalidOffset,
		},
	}

	for _, tc := range tests {
		tc := tc
		suite.Run(tc.name, func() {
			limit, offset, err := parsePaginationParams(tc.values)
			suite.Equal(tc.wantLimit, limit)
			suite.Equal(tc.wantOffset, offset)
			if tc.wantErr != nil {
				suite.Require().NotNil(err)
				suite.Equal(tc.wantErr.Code, err.Code)
				return
			}
			suite.Nil(err)
		})
	}
}

func (suite *UserHandlerTestSuite) TestUserHandler_HandleError() {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	tests := []struct {
		name       string
		err        *serviceerror.ServiceError
		useFlaky   bool
		wantStatus int
		wantCode   string
		wantBody   string
	}{
		{
			name:       "not found",
			err:        &ErrorUserNotFound,
			wantStatus: http.StatusNotFound,
			wantCode:   ErrorUserNotFound.Code,
		},
		{
			name:       "conflict",
			err:        &ErrorAttributeConflict,
			wantStatus: http.StatusConflict,
			wantCode:   ErrorAttributeConflict.Code,
		},
		{
			name:       "unauthorized",
			err:        &ErrorAuthenticationFailed,
			wantStatus: http.StatusUnauthorized,
			wantCode:   ErrorAuthenticationFailed.Code,
		},
		{
			name: "default bad request",
			err: &serviceerror.ServiceError{
				Type:  serviceerror.ClientErrorType,
				Code:  "USR-1999",
				Error: "bad request",
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "USR-1999",
		},
		{
			name:       "server error",
			err:        &ErrorInternalServerError,
			wantStatus: http.StatusInternalServerError,
			wantCode:   ErrorInternalServerError.Code,
		},
		{
			name:       "encode error",
			err:        &ErrorInternalServerError,
			useFlaky:   true,
			wantStatus: http.StatusInternalServerError,
			wantBody:   "Failed to encode error response\n",
		},
	}

	for _, tc := range tests {
		tc := tc
		suite.Run(tc.name, func() {
			var writer http.ResponseWriter
			var recorder *httptest.ResponseRecorder
			if tc.useFlaky {
				flaky := newFlakyResponseWriter()
				writer = flaky
				recorder = flaky.ResponseRecorder
			} else {
				recorder = httptest.NewRecorder()
				writer = recorder
			}

			handleError(writer, logger, tc.err)

			suite.Equal(tc.wantStatus, recorder.Code)
			if tc.wantBody != "" {
				suite.Equal(tc.wantBody, recorder.Body.String())
				return
			}

			var resp apierror.ErrorResponse
			suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &resp))
			suite.Equal(tc.wantCode, resp.Code)
		})
	}
}

func (suite *UserHandlerTestSuite) TestUserHandler_ExtractAndValidatePath() {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))

	suite.Run("success", func() {
		req := httptest.NewRequest(http.MethodGet, "/users/tree/"+defaultUserPath, nil)
		req.SetPathValue("path", defaultUserPath)
		recorder := httptest.NewRecorder()

		path, failed := extractAndValidatePath(recorder, req, logger)

		suite.False(failed)
		suite.Equal(defaultUserPath, path)
	})

	suite.Run("missing path", func() {
		req := httptest.NewRequest(http.MethodGet, "/users/tree/", nil)
		recorder := httptest.NewRecorder()

		_, failed := extractAndValidatePath(recorder, req, logger)

		suite.True(failed)
		suite.Equal(http.StatusBadRequest, recorder.Code)
		var resp apierror.ErrorResponse
		suite.NoError(json.Unmarshal(recorder.Body.Bytes(), &resp))
		suite.Equal(ErrorHandlePathRequired.Code, resp.Code)
	})

	suite.Run("encode error", func() {
		req := httptest.NewRequest(http.MethodGet, "/users/tree/", nil)
		flaky := newFlakyResponseWriter()

		_, failed := extractAndValidatePath(flaky, req, logger)

		suite.True(failed)
		suite.Equal(http.StatusBadRequest, flaky.ResponseRecorder.Code)
		suite.Equal("Failed to encode error response\n", flaky.Body.String())
	})
}

func (suite *UserHandlerTestSuite) TestUserHandler_ParseFilterParams() {
	tests := []struct {
		name    string
		values  url.Values
		want    map[string]interface{}
		wantErr *serviceerror.ServiceError
	}{
		{
			name:   "no filter",
			values: url.Values{},
			want:   map[string]interface{}{},
		},
		{
			name: "string filter",
			values: url.Values{
				"filter": []string{`type eq "basic"`},
			},
			want: map[string]interface{}{"type": userTypeBasic},
		},
		{
			name: "numeric filter",
			values: url.Values{
				"filter": []string{"age eq 10"},
			},
			want: map[string]interface{}{"age": int64(10)},
		},
		{
			name: "float filter",
			values: url.Values{
				"filter": []string{"weight eq 12.5"},
			},
			want: map[string]interface{}{"weight": 12.5},
		},
		{
			name: "invalid filter",
			values: url.Values{
				"filter": []string{"invalid"},
			},
			wantErr: &ErrorInvalidFilter,
		},
	}

	for _, tc := range tests {
		tc := tc
		suite.Run(tc.name, func() {
			result, err := parseFilterParams(tc.values)
			if tc.wantErr != nil {
				suite.Require().NotNil(err)
				suite.Equal(tc.wantErr.Code, err.Code)
				return
			}

			suite.Nil(err)
			suite.Equal(tc.want, result)
		})
	}
}

func (suite *UserHandlerTestSuite) TestUserHandler_ParseFilterExpression() {
	tests := []struct {
		name    string
		input   string
		want    map[string]interface{}
		wantErr string
	}{
		{
			name:  "string value",
			input: `type eq "basic"`,
			want:  map[string]interface{}{"type": userTypeBasic},
		},
		{
			name:  "numeric value",
			input: "age eq 42",
			want:  map[string]interface{}{"age": int64(42)},
		},
		{
			name:  "boolean value",
			input: "active eq true",
			want:  map[string]interface{}{"active": true},
		},
		{
			name:  "float value",
			input: "weight eq 12.5",
			want:  map[string]interface{}{"weight": 12.5},
		},
		{
			name:    "invalid operator",
			input:   `type ne "basic"`,
			wantErr: "unsupported operator",
		},
		{
			name:    "invalid format",
			input:   "bad format",
			wantErr: "invalid filter format",
		},
	}

	for _, tc := range tests {
		tc := tc
		suite.Run(tc.name, func() {
			result, err := parseFilterExpression(tc.input)
			if tc.wantErr != "" {
				suite.Require().Error(err)
				suite.Contains(err.Error(), tc.wantErr)
				return
			}

			suite.NoError(err)
			suite.Equal(tc.want, result)
		})
	}
}

func (suite *UserHandlerTestSuite) TestUserHandler_SanitizeFilter() {
	tests := []struct {
		name string
		in   map[string]interface{}
		want map[string]interface{}
	}{
		{
			name: "sanitizes strings",
			in:   map[string]interface{}{"type": "  basic <script> "},
			want: map[string]interface{}{"type": "basic &lt;script&gt;"},
		},
		{
			name: "preserves numeric",
			in:   map[string]interface{}{"age": int64(10)},
			want: map[string]interface{}{"age": int64(10)},
		},
		{
			name: "preserves bool",
			in:   map[string]interface{}{"active": true},
			want: map[string]interface{}{"active": true},
		},
	}

	for _, tc := range tests {
		tc := tc
		suite.Run(tc.name, func() {
			result := sanitizeFilter(tc.in)
			suite.Equal(tc.want, result)
		})
	}
}
