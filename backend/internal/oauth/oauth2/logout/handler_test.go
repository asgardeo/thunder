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
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// mockLogoutService is an inline test double for logoutServiceInterface.
type mockLogoutService struct {
	logoutFn func(ctx context.Context, idTokenHint, postLogoutRedirectURI string) (string, error)
}

func (m *mockLogoutService) Logout(ctx context.Context, idTokenHint, postLogoutRedirectURI string) (string, error) {
	return m.logoutFn(ctx, idTokenHint, postLogoutRedirectURI)
}

type LogoutHandlerTestSuite struct {
	suite.Suite
}

func TestLogoutHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(LogoutHandlerTestSuite))
}

func (s *LogoutHandlerTestSuite) TestHandleLogout_MissingIDTokenHint() {
	h := newLogoutHandler(&mockLogoutService{})
	req := httptest.NewRequest(http.MethodGet, "/oauth2/logout", nil)
	rr := httptest.NewRecorder()

	h.HandleLogout(rr, req)

	assert.Equal(s.T(), http.StatusBadRequest, rr.Code)
	assert.Contains(s.T(), rr.Body.String(), "id_token_hint is required")
}

func (s *LogoutHandlerTestSuite) TestHandleLogout_ServiceError() {
	svc := &mockLogoutService{
		logoutFn: func(_ context.Context, _, _ string) (string, error) {
			return "", errors.New("invalid id_token_hint")
		},
	}
	h := newLogoutHandler(svc)

	q := url.Values{}
	q.Set("id_token_hint", "bad.token.here")
	req := httptest.NewRequest(http.MethodGet, "/oauth2/logout?"+q.Encode(), nil)
	rr := httptest.NewRecorder()

	h.HandleLogout(rr, req)

	assert.Equal(s.T(), http.StatusBadRequest, rr.Code)
}

func (s *LogoutHandlerTestSuite) TestHandleLogout_NoRedirectURI_Returns200() {
	svc := &mockLogoutService{
		logoutFn: func(_ context.Context, _, _ string) (string, error) {
			return "", nil
		},
	}
	h := newLogoutHandler(svc)

	q := url.Values{}
	q.Set("id_token_hint", "valid.token.here")
	req := httptest.NewRequest(http.MethodGet, "/oauth2/logout?"+q.Encode(), nil)
	rr := httptest.NewRecorder()

	h.HandleLogout(rr, req)

	assert.Equal(s.T(), http.StatusOK, rr.Code)
}

func (s *LogoutHandlerTestSuite) TestHandleLogout_WithRedirectURI_Returns302() {
	redirectURI := "http://localhost:3000"
	svc := &mockLogoutService{
		logoutFn: func(_ context.Context, _, _ string) (string, error) {
			return redirectURI, nil
		},
	}
	h := newLogoutHandler(svc)

	q := url.Values{}
	q.Set("id_token_hint", "valid.token.here")
	q.Set("post_logout_redirect_uri", redirectURI)
	req := httptest.NewRequest(http.MethodGet, "/oauth2/logout?"+q.Encode(), nil)
	rr := httptest.NewRecorder()

	h.HandleLogout(rr, req)

	assert.Equal(s.T(), http.StatusFound, rr.Code)
	assert.Equal(s.T(), redirectURI, rr.Header().Get("Location"))
}

func (s *LogoutHandlerTestSuite) TestHandleLogout_WithRedirectURIAndState_AppendsState() {
	redirectURI := "http://localhost:3000"
	state := "sign_out_success"
	svc := &mockLogoutService{
		logoutFn: func(_ context.Context, _, _ string) (string, error) {
			return redirectURI, nil
		},
	}
	h := newLogoutHandler(svc)

	q := url.Values{}
	q.Set("id_token_hint", "valid.token.here")
	q.Set("post_logout_redirect_uri", redirectURI)
	q.Set("state", state)
	req := httptest.NewRequest(http.MethodGet, "/oauth2/logout?"+q.Encode(), nil)
	rr := httptest.NewRecorder()

	h.HandleLogout(rr, req)

	assert.Equal(s.T(), http.StatusFound, rr.Code)
	assert.Equal(s.T(), redirectURI+"?state="+state, rr.Header().Get("Location"))
}
