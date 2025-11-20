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

package notification

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/tests/mocks/jwtmock"
)

type InitTestSuite struct {
	suite.Suite
	mockJWTService *jwtmock.JWTServiceInterfaceMock
	mux            *http.ServeMux
}

func TestInitTestSuite(t *testing.T) {
	suite.Run(t, new(InitTestSuite))
}

func (suite *InitTestSuite) SetupSuite() {
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "test-issuer",
			ValidityPeriod: 3600,
		},
		Crypto: config.CryptoConfig{
			Encrypt: config.EncryptConfig{
				Key: "b735c757d03c6496575f4c1eb5ba708ad3d06a635566d0fe0440802cf750b3f8",
			},
		},
	}
	err := config.InitializeThunderRuntime("", testConfig)
	if err != nil {
		suite.T().Fatalf("Failed to initialize ThunderRuntime: %v", err)
	}
}

func (suite *InitTestSuite) SetupTest() {
	suite.mockJWTService = jwtmock.NewJWTServiceInterfaceMock(suite.T())
	suite.mux = http.NewServeMux()
}

func (suite *InitTestSuite) TestInitialize() {
	mgtService, otpService := Initialize(suite.mux, suite.mockJWTService)

	suite.NotNil(mgtService)
	suite.NotNil(otpService)
	suite.Implements((*NotificationSenderMgtSvcInterface)(nil), mgtService)
	suite.Implements((*OTPServiceInterface)(nil), otpService)
}

func (suite *InitTestSuite) TestRegisterRoutes_ListEndpoint() {
	Initialize(suite.mux, suite.mockJWTService)

	req := httptest.NewRequest(http.MethodGet, "/notification-senders/message", nil)
	w := httptest.NewRecorder()

	suite.mux.ServeHTTP(w, req)

	suite.NotEqual(http.StatusNotFound, w.Code)
}

func (suite *InitTestSuite) TestRegisterRoutes_CreateEndpoint() {
	Initialize(suite.mux, suite.mockJWTService)

	req := httptest.NewRequest(http.MethodPost, "/notification-senders/message", nil)
	w := httptest.NewRecorder()

	suite.mux.ServeHTTP(w, req)

	suite.NotEqual(http.StatusNotFound, w.Code)
}

func (suite *InitTestSuite) TestRegisterRoutes_GetByIDEndpoint() {
	Initialize(suite.mux, suite.mockJWTService)

	req := httptest.NewRequest(http.MethodGet, "/notification-senders/message/test-id", nil)
	w := httptest.NewRecorder()

	suite.mux.ServeHTTP(w, req)

	suite.NotEqual(http.StatusNotFound, w.Code)
}

func (suite *InitTestSuite) TestRegisterRoutes_UpdateEndpoint() {
	Initialize(suite.mux, suite.mockJWTService)

	req := httptest.NewRequest(http.MethodPut, "/notification-senders/message/test-id", nil)
	w := httptest.NewRecorder()

	suite.mux.ServeHTTP(w, req)

	suite.NotEqual(http.StatusNotFound, w.Code)
}

func (suite *InitTestSuite) TestRegisterRoutes_DeleteEndpoint() {
	Initialize(suite.mux, suite.mockJWTService)

	req := httptest.NewRequest(http.MethodDelete, "/notification-senders/message/test-id", nil)
	w := httptest.NewRecorder()

	suite.mux.ServeHTTP(w, req)

	suite.NotEqual(http.StatusNotFound, w.Code)
}

func (suite *InitTestSuite) TestRegisterRoutes_SendOTPEndpoint() {
	Initialize(suite.mux, suite.mockJWTService)

	req := httptest.NewRequest(http.MethodPost, "/notification-senders/otp/send", nil)
	w := httptest.NewRecorder()

	suite.mux.ServeHTTP(w, req)

	suite.NotEqual(http.StatusNotFound, w.Code)
}

func (suite *InitTestSuite) TestRegisterRoutes_VerifyOTPEndpoint() {
	Initialize(suite.mux, suite.mockJWTService)

	req := httptest.NewRequest(http.MethodPost, "/notification-senders/otp/verify", nil)
	w := httptest.NewRecorder()

	suite.mux.ServeHTTP(w, req)

	suite.NotEqual(http.StatusNotFound, w.Code)
}

func (suite *InitTestSuite) TestRegisterRoutes_CORSPreflight() {
	Initialize(suite.mux, suite.mockJWTService)

	req := httptest.NewRequest(http.MethodOptions, "/notification-senders/message", nil)
	w := httptest.NewRecorder()

	suite.mux.ServeHTTP(w, req)

	suite.NotEqual(http.StatusNotFound, w.Code)
}
