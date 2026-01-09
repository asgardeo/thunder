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
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/config"
)

type JWKSServiceTestSuite struct {
	suite.Suite
	jwksService *jwksService
}

func TestJWKSServiceSuite(t *testing.T) {
	suite.Run(t, new(JWKSServiceTestSuite))
}

func (suite *JWKSServiceTestSuite) SetupTest() {
	suite.jwksService = &jwksService{}
}

func (suite *JWKSServiceTestSuite) setupRuntimeConfig(tlsConfig *tls.Config, certKid string) error {
	testConfig := &config.Config{
		TLS: config.TLSConfig{
			CertFile: "test-cert.pem",
			KeyFile:  "test-key.pem",
		},
	}
	err := config.InitializeThunderRuntime("/tmp", testConfig)
	if err != nil {
		return err
	}

	// Set the cert config in runtime
	runtime := config.GetThunderRuntime()
	runtime.CertConfig = config.CertConfig{
		TLSConfig: tlsConfig,
		CertKid:   certKid,
	}
	return nil
}

func (suite *JWKSServiceTestSuite) TestNewJWKSService() {
	service := newJWKSService()
	assert.NotNil(suite.T(), service)
	assert.Implements(suite.T(), (*JWKSServiceInterface)(nil), service)
}

func (suite *JWKSServiceTestSuite) TestGetJWKS_RSAKey_Success() {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(suite.T(), err)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: nil,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	assert.NoError(suite.T(), err)

	cert := tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  privateKey,
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	// Set up the runtime config with cert config
	err = suite.setupRuntimeConfig(tlsConfig, "test-kid")
	assert.NoError(suite.T(), err)

	result, svcErr := suite.jwksService.GetJWKS()
	assert.Nil(suite.T(), svcErr)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Keys, 1)

	jwk := result.Keys[0]
	assert.Equal(suite.T(), "test-kid", jwk.Kid)
	assert.Equal(suite.T(), "RSA", jwk.Kty)
	assert.Equal(suite.T(), "sig", jwk.Use)
	assert.Equal(suite.T(), "RS256", jwk.Alg)
	assert.NotEmpty(suite.T(), jwk.N)
	assert.NotEmpty(suite.T(), jwk.E)
	assert.NotEmpty(suite.T(), jwk.X5c)
	assert.NotEmpty(suite.T(), jwk.X5t)
	assert.NotEmpty(suite.T(), jwk.X5tS256)
}

func (suite *JWKSServiceTestSuite) TestGetJWKS_ECDSAKey_Success() {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(suite.T(), err)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: nil,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	assert.NoError(suite.T(), err)

	cert := tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  privateKey,
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	// Set up the runtime config with cert config
	err = suite.setupRuntimeConfig(tlsConfig, "test-kid")
	assert.NoError(suite.T(), err)

	result, svcErr := suite.jwksService.GetJWKS()
	assert.Nil(suite.T(), svcErr)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Keys, 1)

	jwk := result.Keys[0]
	assert.Equal(suite.T(), "test-kid", jwk.Kid)
	assert.Equal(suite.T(), "EC", jwk.Kty)
	assert.Equal(suite.T(), "sig", jwk.Use)
	assert.Equal(suite.T(), "ES256", jwk.Alg)
	assert.Equal(suite.T(), "P-256", jwk.Crv)
	assert.NotEmpty(suite.T(), jwk.X)
	assert.NotEmpty(suite.T(), jwk.Y)
	assert.NotEmpty(suite.T(), jwk.X5c)
	assert.NotEmpty(suite.T(), jwk.X5t)
	assert.NotEmpty(suite.T(), jwk.X5tS256)
}

func (suite *JWKSServiceTestSuite) TestGetJWKS_NoCertificatesInTLSConfig() {
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{},
		MinVersion:   tls.VersionTLS12,
	}

	// Set up the runtime config with cert config
	err := suite.setupRuntimeConfig(tlsConfig, "test-kid")
	assert.NoError(suite.T(), err)

	result, svcErr := suite.jwksService.GetJWKS()
	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), ErrorNoCertificateFound.Code, svcErr.Code)
}

func (suite *JWKSServiceTestSuite) TestGetJWKS_EmptyCertificateInTLSConfig() {
	cert := tls.Certificate{
		Certificate: [][]byte{},
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	// Set up the runtime config with cert config
	err := suite.setupRuntimeConfig(tlsConfig, "test-kid")
	assert.NoError(suite.T(), err)

	result, svcErr := suite.jwksService.GetJWKS()
	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), ErrorNoCertificateFound.Code, svcErr.Code)
}

func (suite *JWKSServiceTestSuite) TestGetJWKS_InvalidCertificateData() {
	cert := tls.Certificate{
		Certificate: [][]byte{[]byte("invalid certificate data")},
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	// Set up the runtime config with cert config
	err := suite.setupRuntimeConfig(tlsConfig, "test-kid")
	assert.NoError(suite.T(), err)

	result, svcErr := suite.jwksService.GetJWKS()
	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), ErrorWhileParsingCertificate.Code, svcErr.Code)
}

func (suite *JWKSServiceTestSuite) TestGetJWKS_CertKidNotFound() {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(suite.T(), err)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: nil,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	assert.NoError(suite.T(), err)

	cert := tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  privateKey,
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	// Set up the runtime config with cert config but empty CertKid
	err = suite.setupRuntimeConfig(tlsConfig, "")
	assert.NoError(suite.T(), err)

	result, svcErr := suite.jwksService.GetJWKS()
	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), ErrorCertificateKidNotFound.Code, svcErr.Code)
}

func (suite *JWKSServiceTestSuite) TestGetJWKS_TLSConfigNotFound() {
	// Set up the runtime config with cert config but nil TLSConfig
	err := suite.setupRuntimeConfig(nil, "test-kid")
	assert.NoError(suite.T(), err)

	result, svcErr := suite.jwksService.GetJWKS()
	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), ErrorTLSConfigNotFound.Code, svcErr.Code)
}

func (suite *JWKSServiceTestSuite) TestGetJWKS_ECDSAKey_P384() {
	privateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	assert.NoError(suite.T(), err)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: nil,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	assert.NoError(suite.T(), err)

	cert := tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  privateKey,
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	err = suite.setupRuntimeConfig(tlsConfig, "test-kid")
	assert.NoError(suite.T(), err)

	result, svcErr := suite.jwksService.GetJWKS()
	assert.Nil(suite.T(), svcErr)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Keys, 1)

	jwk := result.Keys[0]
	assert.Equal(suite.T(), "test-kid", jwk.Kid)
	assert.Equal(suite.T(), "EC", jwk.Kty)
	assert.Equal(suite.T(), "ES384", jwk.Alg)
	assert.Equal(suite.T(), "P-384", jwk.Crv)
}

func (suite *JWKSServiceTestSuite) TestGetJWKS_ECDSAKey_P521() {
	privateKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	assert.NoError(suite.T(), err)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: nil,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	assert.NoError(suite.T(), err)

	cert := tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  privateKey,
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	err = suite.setupRuntimeConfig(tlsConfig, "test-kid")
	assert.NoError(suite.T(), err)

	result, svcErr := suite.jwksService.GetJWKS()
	assert.Nil(suite.T(), svcErr)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Keys, 1)

	jwk := result.Keys[0]
	assert.Equal(suite.T(), "test-kid", jwk.Kid)
	assert.Equal(suite.T(), "EC", jwk.Kty)
	assert.Equal(suite.T(), "ES512", jwk.Alg)
	assert.Equal(suite.T(), "P-521", jwk.Crv)
}
