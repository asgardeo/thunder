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
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/tests/mocks/crypto/pki/pkimock"
)

type JWKSServiceTestSuite struct {
	suite.Suite
	jwksService JWKSServiceInterface
	pkiMock     *pkimock.PKIServiceInterfaceMock
}

func TestJWKSServiceSuite(t *testing.T) {
	suite.Run(t, new(JWKSServiceTestSuite))
}

func (suite *JWKSServiceTestSuite) SetupTest() {
	// Reset runtime and set preferred key ID
	config.ResetThunderRuntime()
	testConfig := &config.Config{JWT: config.JWTConfig{PreferredKeyID: "kid-1"}}
	_ = config.InitializeThunderRuntime("", testConfig)

	// Create PKI mock and service under test
	suite.pkiMock = pkimock.NewPKIServiceInterfaceMock(suite.T())
	suite.jwksService = newJWKSService(suite.pkiMock)
}

func (suite *JWKSServiceTestSuite) TestGetJWKS_RSA_Success() {
	// Prepare RSA cert and mock
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	cert := &x509.Certificate{Raw: []byte("rsa-cert-raw"), PublicKey: &key.PublicKey}
	suite.pkiMock.EXPECT().GetX509Certificate("kid-1").Return(cert, nil)
	suite.pkiMock.EXPECT().GetCertThumbprint("kid-1").Return("kid-1")

	resp, svcErr := suite.jwksService.GetJWKS()
	assert.Nil(suite.T(), svcErr)
	assert.NotNil(suite.T(), resp)
	assert.Len(suite.T(), resp.Keys, 1)
	k := resp.Keys[0]
	assert.Equal(suite.T(), "RSA", k.Kty)
	assert.Equal(suite.T(), "RS256", k.Alg)
	assert.NotEmpty(suite.T(), k.N)
	assert.NotEmpty(suite.T(), k.E)
	assert.NotEmpty(suite.T(), k.X5c)
	assert.NotEmpty(suite.T(), k.X5t)
	assert.NotEmpty(suite.T(), k.X5tS256)
}

func (suite *JWKSServiceTestSuite) TestGetJWKS_ECDSA_P256_Success() {
	// Prepare ECDSA P-256 cert and mock
	ecdsaKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	cert := &x509.Certificate{Raw: []byte("ec-cert-raw"), PublicKey: &ecdsaKey.PublicKey}
	suite.pkiMock.EXPECT().GetX509Certificate("kid-1").Return(cert, nil)
	suite.pkiMock.EXPECT().GetCertThumbprint("kid-1").Return("kid-1")

	resp, svcErr := suite.jwksService.GetJWKS()
	assert.Nil(suite.T(), svcErr)
	assert.NotNil(suite.T(), resp)
	assert.Len(suite.T(), resp.Keys, 1)
	k := resp.Keys[0]
	assert.Equal(suite.T(), "EC", k.Kty)
	assert.Equal(suite.T(), "ES256", k.Alg)
	assert.Equal(suite.T(), "P-256", k.Crv)
	assert.NotEmpty(suite.T(), k.X)
	assert.NotEmpty(suite.T(), k.Y)
	assert.NotEmpty(suite.T(), k.X5c)
	assert.NotEmpty(suite.T(), k.X5t)
	assert.NotEmpty(suite.T(), k.X5tS256)
}

func (suite *JWKSServiceTestSuite) TestGetJWKS_CertParseError() {
	// Mock parse error
	parseErr := assert.AnError
	suite.pkiMock.EXPECT().GetX509Certificate("kid-1").Return(nil, parseErr)

	resp, svcErr := suite.jwksService.GetJWKS()
	assert.Nil(suite.T(), resp)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), ErrorWhileParsingCertificate.Code, svcErr.Code)
	assert.Equal(suite.T(), parseErr.Error(), svcErr.ErrorDescription)
}

func (suite *JWKSServiceTestSuite) TestGetJWKS_KidNotFoundError() {
	// Mock empty kid to trigger ErrorCertificateKidNotFound
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	cert := &x509.Certificate{Raw: []byte("rsa-cert-raw"), PublicKey: &key.PublicKey}
	suite.pkiMock.EXPECT().GetX509Certificate("kid-1").Return(cert, nil)
	suite.pkiMock.EXPECT().GetCertThumbprint("kid-1").Return("")

	resp, svcErr := suite.jwksService.GetJWKS()
	assert.Nil(suite.T(), resp)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), ErrorCertificateKidNotFound.Code, svcErr.Code)
}

func (suite *JWKSServiceTestSuite) TestGetJWKS_UnsupportedPublicKeyType() {
	// Provide unsupported ed25519 public key
	_, edPriv, _ := ed25519.GenerateKey(rand.Reader)
	cert := &x509.Certificate{Raw: []byte("ed-cert-raw"), PublicKey: edPriv.Public()}
	suite.pkiMock.EXPECT().GetX509Certificate("kid-1").Return(cert, nil)
	suite.pkiMock.EXPECT().GetCertThumbprint("kid-1").Return("kid-1")

	resp, svcErr := suite.jwksService.GetJWKS()
	assert.Nil(suite.T(), resp)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), ErrorUnsupportedPublicKeyType.Code, svcErr.Code)
}
