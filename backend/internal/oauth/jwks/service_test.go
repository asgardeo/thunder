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
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/i18n/core"
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
	// Reset runtime
	config.ResetServerRuntime()
	testConfig := &config.Config{}
	_ = config.InitializeServerRuntime("", testConfig)

	// Create PKI mock and service under test
	suite.pkiMock = pkimock.NewPKIServiceInterfaceMock(suite.T())
	suite.jwksService = newJWKSService(suite.pkiMock)
}

func (suite *JWKSServiceTestSuite) TestGetJWKS_RSA_Success() {
	// Prepare RSA cert and mock
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	cert := &x509.Certificate{Raw: []byte("rsa-cert-raw"), PublicKey: &key.PublicKey}
	allCerts := map[string]*x509.Certificate{"kid-1": cert}
	suite.pkiMock.EXPECT().GetAllX509Certificates().Return(allCerts, nil)
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
	allCerts := map[string]*x509.Certificate{"kid-1": cert}
	suite.pkiMock.EXPECT().GetAllX509Certificates().Return(allCerts, nil)
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

func (suite *JWKSServiceTestSuite) TestGetJWKS_EdDSA_Success() {
	// Prepare EdDSA cert and mock
	_, edPriv, _ := ed25519.GenerateKey(rand.Reader)
	cert := &x509.Certificate{Raw: []byte("ed-cert-raw"), PublicKey: edPriv.Public()}
	allCerts := map[string]*x509.Certificate{"kid-1": cert}
	suite.pkiMock.EXPECT().GetAllX509Certificates().Return(allCerts, nil)
	suite.pkiMock.EXPECT().GetCertThumbprint("kid-1").Return("kid-1")

	resp, svcErr := suite.jwksService.GetJWKS()
	assert.Nil(suite.T(), svcErr)
	assert.NotNil(suite.T(), resp)
	assert.Len(suite.T(), resp.Keys, 1)
	k := resp.Keys[0]
	assert.Equal(suite.T(), "OKP", k.Kty)
	assert.Equal(suite.T(), "EdDSA", k.Alg)
	assert.Equal(suite.T(), "Ed25519", k.Crv)
	assert.NotEmpty(suite.T(), k.X)
	assert.NotEmpty(suite.T(), k.X5c)
	assert.NotEmpty(suite.T(), k.X5t)
	assert.NotEmpty(suite.T(), k.X5tS256)
}

func (suite *JWKSServiceTestSuite) TestGetJWKS_CertParseError() {
	// Mock parse error
	parseErr := serviceerror.CustomServiceError(serviceerror.InternalServerError, core.I18nMessage{
		Key:          "error.test.parse_error",
		DefaultValue: "parse error",
	})
	suite.pkiMock.EXPECT().GetAllX509Certificates().Return(nil, parseErr)

	resp, svcErr := suite.jwksService.GetJWKS()
	assert.Nil(suite.T(), resp)
	assert.NotNil(suite.T(), svcErr)
	// The error is passed through from PKI service, so we expect the PKI error code
	assert.Equal(suite.T(), parseErr.Code, svcErr.Code)
	assert.Equal(suite.T(), parseErr.ErrorDescription.DefaultValue, svcErr.ErrorDescription.DefaultValue)
}

func (suite *JWKSServiceTestSuite) TestGetJWKS_NoCertificatesFound() {
	// Mock empty certificates
	allCerts := map[string]*x509.Certificate{}
	suite.pkiMock.EXPECT().GetAllX509Certificates().Return(allCerts, nil)

	resp, svcErr := suite.jwksService.GetJWKS()
	assert.Nil(suite.T(), resp)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), serviceerror.InternalServerError.Code, svcErr.Code)
}

func (suite *JWKSServiceTestSuite) TestGetJWKS_UnsupportedPublicKeyType() {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	cert := &x509.Certificate{Raw: []byte("rsa-cert-raw"), PublicKey: &key.PublicKey}
	// Provide an unsupported public key type (using a string as an invalid example)
	errCert := &x509.Certificate{Raw: []byte("unsupported-cert-raw"), PublicKey: "unsupported-key-type"}
	allCerts := map[string]*x509.Certificate{"kid-1": errCert, "kid-2": cert}
	suite.pkiMock.EXPECT().GetAllX509Certificates().Return(allCerts, nil)
	suite.pkiMock.EXPECT().GetCertThumbprint("kid-1").Return("kid-1")
	suite.pkiMock.EXPECT().GetCertThumbprint("kid-2").Return("kid-2")

	resp, svcErr := suite.jwksService.GetJWKS()
	assert.Nil(suite.T(), svcErr)
	assert.NotNil(suite.T(), resp)
	// The unsupported key should be skipped, so only one valid key should be present
	assert.Len(suite.T(), resp.Keys, 1)
}

func (suite *JWKSServiceTestSuite) TestGetJWKS_MultipleCertificates() {
	// Prepare multiple certs (RSA and ECDSA)
	rsaKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	rsaCert := &x509.Certificate{Raw: []byte("rsa-cert-raw"), PublicKey: &rsaKey.PublicKey}

	ecdsaKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	ecCert := &x509.Certificate{Raw: []byte("ec-cert-raw"), PublicKey: &ecdsaKey.PublicKey}

	allCerts := map[string]*x509.Certificate{
		"rsa-kid": rsaCert,
		"ec-kid":  ecCert,
	}
	suite.pkiMock.EXPECT().GetAllX509Certificates().Return(allCerts, nil)
	suite.pkiMock.EXPECT().GetCertThumbprint("rsa-kid").Return("rsa-kid")
	suite.pkiMock.EXPECT().GetCertThumbprint("ec-kid").Return("ec-kid")

	resp, svcErr := suite.jwksService.GetJWKS()
	assert.Nil(suite.T(), svcErr)
	assert.NotNil(suite.T(), resp)
	assert.Len(suite.T(), resp.Keys, 2)

	// Check that both key types are present
	rsaFound := false
	ecFound := false
	for _, k := range resp.Keys {
		if k.Kty == "RSA" {
			rsaFound = true
			assert.Equal(suite.T(), "RS256", k.Alg)
		}
		if k.Kty == "EC" {
			ecFound = true
			assert.Equal(suite.T(), "ES256", k.Alg)
		}
	}
	assert.True(suite.T(), rsaFound, "RSA key not found in JWKS")
	assert.True(suite.T(), ecFound, "EC key not found in JWKS")
}

func (suite *JWKSServiceTestSuite) TestGetJWKS_ECDSA_AdditionalCurves() {
	tests := []struct {
		name  string
		curve elliptic.Curve
		alg   string
		crv   string
	}{
		{
			name:  "P-384",
			curve: elliptic.P384(),
			alg:   "ES384",
			crv:   "P-384",
		},
		{
			name:  "P-521",
			curve: elliptic.P521(),
			alg:   "ES512",
			crv:   "P-521",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ecdsaKey, _ := ecdsa.GenerateKey(tt.curve, rand.Reader)
			cert := &x509.Certificate{Raw: []byte("ec-cert-raw-" + tt.name), PublicKey: &ecdsaKey.PublicKey}
			kid := "kid-" + tt.name
			allCerts := map[string]*x509.Certificate{kid: cert}

			suite.pkiMock.EXPECT().GetAllX509Certificates().Return(allCerts, nil).Once()
			suite.pkiMock.EXPECT().GetCertThumbprint(kid).Return(kid).Once()

			resp, svcErr := suite.jwksService.GetJWKS()
			assert.Nil(suite.T(), svcErr)
			assert.NotNil(suite.T(), resp)
			assert.Len(suite.T(), resp.Keys, 1)
			k := resp.Keys[0]
			assert.Equal(suite.T(), "EC", k.Kty)
			assert.Equal(suite.T(), tt.alg, k.Alg)
			assert.Equal(suite.T(), tt.crv, k.Crv)
			assert.NotEmpty(suite.T(), k.X)
			assert.NotEmpty(suite.T(), k.Y)
		})
	}
}

func (suite *JWKSServiceTestSuite) TestGetJWKS_OnlyUnsupportedKeys() {
	// Provide only an unsupported public key type
	errCert := &x509.Certificate{Raw: []byte("unsupported-cert-raw"), PublicKey: "unsupported-key-type"}
	allCerts := map[string]*x509.Certificate{"kid-1": errCert}
	suite.pkiMock.EXPECT().GetAllX509Certificates().Return(allCerts, nil)
	suite.pkiMock.EXPECT().GetCertThumbprint("kid-1").Return("kid-1")

	resp, svcErr := suite.jwksService.GetJWKS()
	assert.Nil(suite.T(), resp)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), serviceerror.InternalServerError.Code, svcErr.Code)
}

func (suite *JWKSServiceTestSuite) TestGetJWKS_RSA_ZeroExponent() {
	// Prepare RSA cert with E=0 and mock
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	key.PublicKey.E = 0
	cert := &x509.Certificate{Raw: []byte("rsa-cert-raw-zero"), PublicKey: &key.PublicKey}
	allCerts := map[string]*x509.Certificate{"kid-zero": cert}
	suite.pkiMock.EXPECT().GetAllX509Certificates().Return(allCerts, nil)
	suite.pkiMock.EXPECT().GetCertThumbprint("kid-zero").Return("kid-zero")

	resp, svcErr := suite.jwksService.GetJWKS()
	assert.Nil(suite.T(), svcErr)
	assert.NotNil(suite.T(), resp)
	assert.Len(suite.T(), resp.Keys, 1)
	k := resp.Keys[0]
	assert.Equal(suite.T(), "RSA", k.Kty)
	// Base64Url(0) -> "AA" (if []byte{0}) or ""?
	// The code does:
	// if len(eBytes) == 0 { eBytes = []byte{0} }
	// encodeBase64URL([]byte{0}) -> "AA" (because base64 of 0 is AA==, trimmed =) -> "AA"
	assert.Equal(suite.T(), "AA", k.E)
}
