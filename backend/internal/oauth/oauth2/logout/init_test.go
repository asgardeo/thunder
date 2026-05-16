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
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/thunder-id/thunderid/internal/oauth/oauth2/constants"
	"github.com/thunder-id/thunderid/tests/mocks/inboundclientmock"
	"github.com/thunder-id/thunderid/tests/mocks/jose/jwtmock"
)

type LogoutInitTestSuite struct {
	suite.Suite
	mockJWTService   *jwtmock.JWTServiceInterfaceMock
	mockInboundClient *inboundclientmock.InboundClientServiceInterfaceMock
}

func TestLogoutInitTestSuite(t *testing.T) {
	suite.Run(t, new(LogoutInitTestSuite))
}

func (s *LogoutInitTestSuite) SetupTest() {
	s.mockJWTService = jwtmock.NewJWTServiceInterfaceMock(s.T())
	s.mockInboundClient = inboundclientmock.NewInboundClientServiceInterfaceMock(s.T())
}

func (s *LogoutInitTestSuite) TestInitialize_RegistersRoutes() {
	mux := http.NewServeMux()

	Initialize(mux, s.mockJWTService, s.mockInboundClient)

	_, pattern := mux.Handler(&http.Request{
		Method: http.MethodGet,
		URL:    &url.URL{Path: constants.OAuth2LogoutEndpoint},
	})
	assert.Contains(s.T(), pattern, constants.OAuth2LogoutEndpoint)

	_, pattern = mux.Handler(&http.Request{
		Method: http.MethodOptions,
		URL:    &url.URL{Path: constants.OAuth2LogoutEndpoint},
	})
	assert.Contains(s.T(), pattern, constants.OAuth2LogoutEndpoint)
}
