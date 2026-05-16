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
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/thunder-id/thunderid/internal/inboundclient/model"
	sysconfig "github.com/thunder-id/thunderid/internal/system/config"
	"github.com/thunder-id/thunderid/internal/system/error/serviceerror"
	"github.com/thunder-id/thunderid/internal/system/i18n/core"
	"github.com/thunder-id/thunderid/tests/mocks/inboundclientmock"
	"github.com/thunder-id/thunderid/tests/mocks/jose/jwtmock"
)

// buildUnsignedJWT builds a minimal JWT with the given payload claims (no signature verification).
func buildUnsignedJWT(claims map[string]interface{}) string {
	header := map[string]interface{}{"alg": "RS256", "typ": "JWT"}
	headerBytes, _ := json.Marshal(header)
	claimsBytes, _ := json.Marshal(claims)
	h := base64.RawURLEncoding.EncodeToString(headerBytes)
	c := base64.RawURLEncoding.EncodeToString(claimsBytes)
	return h + "." + c + ".fakesig"
}

type LogoutServiceTestSuite struct {
	suite.Suite
	jwtMock         *jwtmock.JWTServiceInterfaceMock
	inboundMock     *inboundclientmock.InboundClientServiceInterfaceMock
	service         logoutServiceInterface
}

func TestLogoutServiceTestSuite(t *testing.T) {
	suite.Run(t, new(LogoutServiceTestSuite))
}

func (s *LogoutServiceTestSuite) SetupTest() {
	s.Require().NoError(sysconfig.InitializeServerRuntime("/tmp/test", &sysconfig.Config{}))
	s.jwtMock = jwtmock.NewJWTServiceInterfaceMock(s.T())
	s.inboundMock = inboundclientmock.NewInboundClientServiceInterfaceMock(s.T())
	s.service = newLogoutService(s.jwtMock, s.inboundMock)
}

func (s *LogoutServiceTestSuite) TestLogout_InvalidSignature() {
	token := buildUnsignedJWT(map[string]interface{}{"aud": "client1"})
	s.jwtMock.On("VerifyJWTSignature", token).Return(&serviceerror.ServiceError{
		ErrorDescription: core.I18nMessage{DefaultValue: "signature verification failed"},
	})

	_, err := s.service.Logout(context.Background(), token, "")
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "invalid id_token_hint")
	s.jwtMock.AssertExpectations(s.T())
}

func (s *LogoutServiceTestSuite) TestLogout_MalformedPayload() {
	// A token with a non-base64 payload segment
	badToken := "aGVhZGVy.!!!.c2ln"
	s.jwtMock.On("VerifyJWTSignature", badToken).Return(nil)

	_, err := s.service.Logout(context.Background(), badToken, "")
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "malformed id_token_hint")
	s.jwtMock.AssertExpectations(s.T())
}

func (s *LogoutServiceTestSuite) TestLogout_MissingAudClaim() {
	token := buildUnsignedJWT(map[string]interface{}{"sub": "user1"})
	s.jwtMock.On("VerifyJWTSignature", token).Return(nil)

	_, err := s.service.Logout(context.Background(), token, "")
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "missing aud claim")
	s.jwtMock.AssertExpectations(s.T())
}

func (s *LogoutServiceTestSuite) TestLogout_NoPostLogoutURI_ReturnsEmpty() {
	token := buildUnsignedJWT(map[string]interface{}{"aud": "client1"})
	s.jwtMock.On("VerifyJWTSignature", token).Return(nil)

	redirectURI, err := s.service.Logout(context.Background(), token, "")
	assert.NoError(s.T(), err)
	assert.Empty(s.T(), redirectURI)
	s.jwtMock.AssertExpectations(s.T())
}

func (s *LogoutServiceTestSuite) TestLogout_ClientNotFound() {
	token := buildUnsignedJWT(map[string]interface{}{"aud": "unknown-client"})
	s.jwtMock.On("VerifyJWTSignature", token).Return(nil)
	s.inboundMock.On("GetOAuthClientByClientID", mock.Anything, "unknown-client").
		Return(nil, assert.AnError)

	_, err := s.service.Logout(context.Background(), token, "http://localhost:3000")
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "client not found")
	s.jwtMock.AssertExpectations(s.T())
	s.inboundMock.AssertExpectations(s.T())
}

func (s *LogoutServiceTestSuite) TestLogout_InvalidRedirectURI() {
	clientID := "client1"
	token := buildUnsignedJWT(map[string]interface{}{"aud": clientID})
	s.jwtMock.On("VerifyJWTSignature", token).Return(nil)
	s.inboundMock.On("GetOAuthClientByClientID", mock.Anything, clientID).
		Return(&model.OAuthClient{
			RedirectURIs: []string{"http://localhost:3000"},
		}, nil)

	_, err := s.service.Logout(context.Background(), token, "http://evil.com")
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "post_logout_redirect_uri not registered")
	s.jwtMock.AssertExpectations(s.T())
	s.inboundMock.AssertExpectations(s.T())
}

func (s *LogoutServiceTestSuite) TestLogout_ValidRedirectURI() {
	clientID := "client1"
	redirectURI := "http://localhost:3000"
	token := buildUnsignedJWT(map[string]interface{}{"aud": clientID})
	s.jwtMock.On("VerifyJWTSignature", token).Return(nil)
	s.inboundMock.On("GetOAuthClientByClientID", mock.Anything, clientID).
		Return(&model.OAuthClient{
			RedirectURIs: []string{redirectURI},
		}, nil)

	result, err := s.service.Logout(context.Background(), token, redirectURI)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), redirectURI, result)
	s.jwtMock.AssertExpectations(s.T())
	s.inboundMock.AssertExpectations(s.T())
}

func (s *LogoutServiceTestSuite) TestLogout_AudAsSlice() {
	clientID := "client1"
	redirectURI := "http://localhost:3000"
	token := buildUnsignedJWT(map[string]interface{}{"aud": []interface{}{clientID}})
	s.jwtMock.On("VerifyJWTSignature", token).Return(nil)
	s.inboundMock.On("GetOAuthClientByClientID", mock.Anything, clientID).
		Return(&model.OAuthClient{
			RedirectURIs: []string{redirectURI},
		}, nil)

	result, err := s.service.Logout(context.Background(), token, redirectURI)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), redirectURI, result)
}
