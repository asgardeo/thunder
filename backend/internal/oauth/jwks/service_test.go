/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
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
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/oauth/jwks/constants"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/tests/mocks/certmock"
)

type JWKSServiceTestSuite struct {
	suite.Suite
	mockCertService *certmock.SystemCertificateServiceInterfaceMock
	jwksService     *JWKSService
}

func TestJWKSServiceSuite(t *testing.T) {
	suite.Run(t, new(JWKSServiceTestSuite))
}

func (suite *JWKSServiceTestSuite) SetupTest() {
	suite.mockCertService = &certmock.SystemCertificateServiceInterfaceMock{}
	suite.jwksService = &JWKSService{
		SystemCertService: suite.mockCertService,
	}

	// Initialize runtime configuration for testing
	testConfig := &config.Config{
		Security: config.SecurityConfig{
			CertFile: "test-cert.pem",
			KeyFile:  "test-key.pem",
		},
	}
	err := config.InitializeThunderRuntime("/tmp", testConfig)
	assert.NoError(suite.T(), err)
}

func (suite *JWKSServiceTestSuite) TestNewJWKSService() {
	service := NewJWKSService()
	assert.NotNil(suite.T(), service)
	assert.Implements(suite.T(), (*JWKSServiceInterface)(nil), service)
}

func (suite *JWKSServiceTestSuite) TestGetJWKS_RSAKey_Success() {
	// Generate a test RSA certificate
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

	// Set up mock expectations
	suite.mockCertService.On("GetCertificateKid").Return("test-kid", nil)
	suite.mockCertService.On("GetTLSConfig", mock.Anything, mock.Anything).Return(tlsConfig, nil)

	// Call the method
	result, svcErr := suite.jwksService.GetJWKS()

	// Verify results
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

	suite.mockCertService.AssertExpectations(suite.T())
}

func (suite *JWKSServiceTestSuite) TestGetJWKS_ECDSAKey_Success() {
	// Generate a test ECDSA certificate
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

	// Set up mock expectations
	suite.mockCertService.On("GetCertificateKid").Return("test-kid", nil)
	suite.mockCertService.On("GetTLSConfig", mock.Anything, mock.Anything).Return(tlsConfig, nil)

	// Call the method
	result, svcErr := suite.jwksService.GetJWKS()

	// Verify results
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

	suite.mockCertService.AssertExpectations(suite.T())
}

func (suite *JWKSServiceTestSuite) TestGetJWKS_ErrorRetrievingKid() {
	// Set up mock expectations
	suite.mockCertService.On("GetCertificateKid").Return("", errors.New("kid error"))

	// Call the method
	result, svcErr := suite.jwksService.GetJWKS()

	// Verify results
	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), constants.ErrorWhileRetrievingCertificateKid.Code, svcErr.Code)
	assert.Equal(suite.T(), "kid error", svcErr.ErrorDescription)

	suite.mockCertService.AssertExpectations(suite.T())
}

func (suite *JWKSServiceTestSuite) TestGetJWKS_ErrorRetrievingTLSConfig() {
	// Set up mock expectations
	suite.mockCertService.On("GetCertificateKid").Return("test-kid", nil)
	suite.mockCertService.On("GetTLSConfig", mock.Anything, mock.Anything).Return(nil, errors.New("tls error"))

	// Call the method
	result, svcErr := suite.jwksService.GetJWKS()

	// Verify results
	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), constants.ErrorWhileRetrievingTLSConfig.Code, svcErr.Code)
	assert.Equal(suite.T(), "tls error", svcErr.ErrorDescription)

	suite.mockCertService.AssertExpectations(suite.T())
}

func (suite *JWKSServiceTestSuite) TestGetJWKS_NoCertificatesInTLSConfig() {
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{},
		MinVersion:   tls.VersionTLS12,
	}

	// Set up mock expectations
	suite.mockCertService.On("GetCertificateKid").Return("test-kid", nil)
	suite.mockCertService.On("GetTLSConfig", mock.Anything, mock.Anything).Return(tlsConfig, nil)

	// Call the method
	result, svcErr := suite.jwksService.GetJWKS()

	// Verify results
	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), constants.ErrorNoCertificateFound.Code, svcErr.Code)

	suite.mockCertService.AssertExpectations(suite.T())
}

func (suite *JWKSServiceTestSuite) TestGetJWKS_EmptyCertificateInTLSConfig() {
	cert := tls.Certificate{
		Certificate: [][]byte{},
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	// Set up mock expectations
	suite.mockCertService.On("GetCertificateKid").Return("test-kid", nil)
	suite.mockCertService.On("GetTLSConfig", mock.Anything, mock.Anything).Return(tlsConfig, nil)

	// Call the method
	result, svcErr := suite.jwksService.GetJWKS()

	// Verify results
	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), constants.ErrorNoCertificateFound.Code, svcErr.Code)

	suite.mockCertService.AssertExpectations(suite.T())
}

func (suite *JWKSServiceTestSuite) TestGetJWKS_InvalidCertificateData() {
	cert := tls.Certificate{
		Certificate: [][]byte{[]byte("invalid certificate data")},
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	// Set up mock expectations
	suite.mockCertService.On("GetCertificateKid").Return("test-kid", nil)
	suite.mockCertService.On("GetTLSConfig", mock.Anything, mock.Anything).Return(tlsConfig, nil)

	// Call the method
	result, svcErr := suite.jwksService.GetJWKS()

	// Verify results
	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), svcErr)
	assert.Equal(suite.T(), constants.ErrorWhileParsingCertificate.Code, svcErr.Code)

	suite.mockCertService.AssertExpectations(suite.T())
}

func (suite *JWKSServiceTestSuite) TestJWKSServiceInterface() {
	// Test that JWKSService implements JWKSServiceInterface
	var _ JWKSServiceInterface = &JWKSService{}
}
