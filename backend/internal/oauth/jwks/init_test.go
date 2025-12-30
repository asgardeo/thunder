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

package jwks

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/tests/mocks/crypto/pki/pkimock"
)

type InitTestSuite struct {
	suite.Suite
}

func TestInitTestSuite(t *testing.T) {
	suite.Run(t, new(InitTestSuite))
}

func (suite *InitTestSuite) SetupTest() {
	// Initialize Thunder Runtime config for CORS middleware
	testConfig := &config.Config{}
	_ = config.InitializeThunderRuntime("test", testConfig)
}

func (suite *InitTestSuite) TearDownTest() {
	config.ResetThunderRuntime()
}

func (suite *InitTestSuite) TestInitialize() {
	mux := http.NewServeMux()

	// Prepare PKI mock with minimal expectations
	pkiMock := pkimock.NewPKIServiceInterfaceMock(suite.T())

	service := Initialize(mux, pkiMock)

	assert.NotNil(suite.T(), service)
	assert.Implements(suite.T(), (*JWKSServiceInterface)(nil), service)
}

func (suite *InitTestSuite) TestInitialize_RegistersRoutes() {
	mux := http.NewServeMux()
	// Prepare PKI mock with minimal expectations for handler invocation
	pkiMock := pkimock.NewPKIServiceInterfaceMock(suite.T())
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	cert := &x509.Certificate{Raw: []byte("raw-cert"), PublicKey: &key.PublicKey}
	pkiMock.EXPECT().GetX509Certificate(mock.Anything).Return(cert, nil)
	pkiMock.EXPECT().GetCertThumbprint(mock.Anything).Return("test-kid")

	_ = Initialize(mux, pkiMock)

	// Test that routes are registered by making requests
	req := httptest.NewRequest("GET", "/oauth2/jwks", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.NotEqual(suite.T(), http.StatusNotFound, w.Code)

	// Test OPTIONS request
	req = httptest.NewRequest("OPTIONS", "/oauth2/jwks", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusNoContent, w.Code)
}
