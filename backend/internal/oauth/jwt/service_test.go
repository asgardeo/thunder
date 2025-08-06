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

package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"math/big"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/tests/mocks/certmock"
)

type JWTServiceTestSuite struct {
	suite.Suite
	mockCertService *certmock.MockSystemCertificateServiceInterface
	jwtService      *JWTService
	testKeyFile     string
	testCertFile    string
}

func TestJWTServiceSuite(t *testing.T) {
	suite.Run(t, new(JWTServiceTestSuite))
}

func (suite *JWTServiceTestSuite) SetupTest() {
	// Reset singleton
	instance = nil
	once = sync.Once{}

	// Create test key and cert files
	suite.createTestKeyFiles()

	// Initialize runtime configuration for testing
	testConfig := &config.Config{
		Security: config.SecurityConfig{
			CertFile: suite.testCertFile,
			KeyFile:  suite.testKeyFile,
		},
		OAuth: config.OAuthConfig{
			JWT: config.JWTConfig{
				Issuer:          "https://test-issuer.com",
				ValidityPeriod:  7200,
			},
		},
	}
	err := config.InitializeThunderRuntime("/tmp", testConfig)
	assert.NoError(suite.T(), err)

	// Setup mocks
	suite.mockCertService = &certmock.MockSystemCertificateServiceInterface{}
	
	suite.jwtService = &JWTService{
		SystemCertificateService: suite.mockCertService,
	}
}

func (suite *JWTServiceTestSuite) TearDownTest() {
	// Clean up test files
	if suite.testKeyFile != "" {
		os.Remove(suite.testKeyFile)
	}
	if suite.testCertFile != "" {
		os.Remove(suite.testCertFile)
	}
}

func (suite *JWTServiceTestSuite) createTestKeyFiles() {
	// Generate a test RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(suite.T(), err)

	// Create certificate
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	assert.NoError(suite.T(), err)

	// Save private key to file
	keyFile, err := os.CreateTemp("", "test-key-*.pem")
	assert.NoError(suite.T(), err)
	suite.testKeyFile = keyFile.Name()

	keyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	err = pem.Encode(keyFile, keyPEM)
	assert.NoError(suite.T(), err)
	keyFile.Close()

	// Save certificate to file
	certFile, err := os.CreateTemp("", "test-cert-*.pem")
	assert.NoError(suite.T(), err)
	suite.testCertFile = certFile.Name()

	certPEM := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	}
	err = pem.Encode(certFile, certPEM)
	assert.NoError(suite.T(), err)
	certFile.Close()
}

func (suite *JWTServiceTestSuite) TestGetJWTService() {
	service1 := GetJWTService()
	service2 := GetJWTService()
	
	assert.NotNil(suite.T(), service1)
	assert.Same(suite.T(), service1, service2) // Should be singleton
	assert.Implements(suite.T(), (*JWTServiceInterface)(nil), service1)
}

func (suite *JWTServiceTestSuite) TestInit_Success() {
	service := GetJWTService()
	err := service.Init()
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), service.GetPublicKey())
}

func (suite *JWTServiceTestSuite) TestInit_KeyFileNotFound() {
	// Update config to point to non-existent file
	testConfig := &config.Config{
		Security: config.SecurityConfig{
			KeyFile: "non-existent-key.pem",
		},
	}
	err := config.InitializeThunderRuntime("/tmp", testConfig)
	assert.NoError(suite.T(), err)

	service := GetJWTService()
	err = service.Init()
	
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "key file not found")
}

func (suite *JWTServiceTestSuite) TestInit_InvalidKeyFile() {
	// Create invalid key file
	invalidKeyFile, err := os.CreateTemp("", "invalid-key-*.pem")
	assert.NoError(suite.T(), err)
	defer os.Remove(invalidKeyFile.Name())

	_, err = invalidKeyFile.WriteString("invalid key data")
	assert.NoError(suite.T(), err)
	invalidKeyFile.Close()

	// Update config
	testConfig := &config.Config{
		Security: config.SecurityConfig{
			KeyFile: invalidKeyFile.Name(),
		},
	}
	err = config.InitializeThunderRuntime("/tmp", testConfig)
	assert.NoError(suite.T(), err)

	service := GetJWTService()
	err = service.Init()
	
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to decode PEM block")
}

func (suite *JWTServiceTestSuite) TestGetPublicKey() {
	service := GetJWTService()
	err := service.Init()
	assert.NoError(suite.T(), err)
	
	pubKey := service.GetPublicKey()
	assert.NotNil(suite.T(), pubKey)
	assert.IsType(suite.T(), &rsa.PublicKey{}, pubKey)
}

func (suite *JWTServiceTestSuite) TestGetPublicKey_NotInitialized() {
	service := &JWTService{}
	pubKey := service.GetPublicKey()
	assert.Nil(suite.T(), pubKey)
}

func (suite *JWTServiceTestSuite) TestGenerateJWT_Success() {
	// Setup mocks
	suite.mockCertService.On("GetCertificateKid").Return("test-kid", nil)
	
	service := &JWTService{
		SystemCertificateService: suite.mockCertService,
	}
	err := service.Init()
	assert.NoError(suite.T(), err)

	claims := map[string]string{
		"custom1": "value1",
		"custom2": "value2",
	}

	token, iat, err := service.GenerateJWT("test-subject", "test-audience", 3600, claims)
	
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), token)
	assert.Greater(suite.T(), iat, int64(0))

	// Verify token structure
	parts := strings.Split(token, ".")
	assert.Len(suite.T(), parts, 3)

	// Verify header
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	assert.NoError(suite.T(), err)
	
	var header map[string]interface{}
	err = json.Unmarshal(headerBytes, &header)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "RS256", header["alg"])
	assert.Equal(suite.T(), "JWT", header["typ"])
	assert.Equal(suite.T(), "test-kid", header["kid"])

	// Verify payload
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	assert.NoError(suite.T(), err)
	
	var payload map[string]interface{}
	err = json.Unmarshal(payloadBytes, &payload)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test-subject", payload["sub"])
	assert.Equal(suite.T(), "test-audience", payload["aud"])
	assert.Equal(suite.T(), "https://test-issuer.com", payload["iss"])
	assert.Equal(suite.T(), "value1", payload["custom1"])
	assert.Equal(suite.T(), "value2", payload["custom2"])
	assert.Contains(suite.T(), payload, "exp")
	assert.Contains(suite.T(), payload, "iat")
	assert.Contains(suite.T(), payload, "nbf")
	assert.Contains(suite.T(), payload, "jti")

	suite.mockCertService.AssertExpectations(suite.T())
}

func (suite *JWTServiceTestSuite) TestGenerateJWT_DefaultValidityPeriod() {
	// Setup mocks
	suite.mockCertService.On("GetCertificateKid").Return("test-kid", nil)
	
	service := &JWTService{
		SystemCertificateService: suite.mockCertService,
	}
	err := service.Init()
	assert.NoError(suite.T(), err)

	token, iat, err := service.GenerateJWT("test-subject", "test-audience", 0, nil)
	
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), token)
	assert.Greater(suite.T(), iat, int64(0))

	suite.mockCertService.AssertExpectations(suite.T())
}

func (suite *JWTServiceTestSuite) TestGenerateJWT_PrivateKeyNotLoaded() {
	service := &JWTService{
		SystemCertificateService: suite.mockCertService,
	}

	token, iat, err := service.GenerateJWT("test-subject", "test-audience", 3600, nil)
	
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "private key not loaded", err.Error())
	assert.Empty(suite.T(), token)
	assert.Equal(suite.T(), int64(0), iat)
}

func (suite *JWTServiceTestSuite) TestGenerateJWT_ErrorGettingKid() {
	// Setup mocks
	suite.mockCertService.On("GetCertificateKid").Return("", errors.New("kid error"))
	
	service := &JWTService{
		SystemCertificateService: suite.mockCertService,
	}
	err := service.Init()
	assert.NoError(suite.T(), err)

	token, iat, err := service.GenerateJWT("test-subject", "test-audience", 3600, nil)
	
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "kid error", err.Error())
	assert.Empty(suite.T(), token)
	assert.Equal(suite.T(), int64(0), iat)

	suite.mockCertService.AssertExpectations(suite.T())
}

func (suite *JWTServiceTestSuite) TestGetJWTTokenValidityPeriod() {
	validityPeriod := GetJWTTokenValidityPeriod()
	assert.Equal(suite.T(), int64(7200), validityPeriod)
}

func (suite *JWTServiceTestSuite) TestGetJWTTokenValidityPeriod_DefaultValue() {
	// Test with config having 0 validity period
	testConfig := &config.Config{
		OAuth: config.OAuthConfig{
			JWT: config.JWTConfig{
				ValidityPeriod: 0,
			},
		},
	}
	err := config.InitializeThunderRuntime("/tmp", testConfig)
	assert.NoError(suite.T(), err)

	validityPeriod := GetJWTTokenValidityPeriod()
	assert.Equal(suite.T(), int64(defaultTokenValidity), validityPeriod)
}

func (suite *JWTServiceTestSuite) TestDecodeJWT_Success() {
	// Create a test JWT manually
	header := map[string]interface{}{
		"alg": "RS256",
		"typ": "JWT",
	}
	payload := map[string]interface{}{
		"sub": "test-subject",
		"aud": "test-audience",
		"exp": time.Now().Add(time.Hour).Unix(),
	}

	headerBytes, _ := json.Marshal(header)
	payloadBytes, _ := json.Marshal(payload)

	headerB64 := base64.RawURLEncoding.EncodeToString(headerBytes)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadBytes)
	signature := "fake-signature"

	token := headerB64 + "." + payloadB64 + "." + signature

	decodedHeader, decodedPayload, err := DecodeJWT(token)
	
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "RS256", decodedHeader["alg"])
	assert.Equal(suite.T(), "JWT", decodedHeader["typ"])
	assert.Equal(suite.T(), "test-subject", decodedPayload["sub"])
	assert.Equal(suite.T(), "test-audience", decodedPayload["aud"])
}

func (suite *JWTServiceTestSuite) TestDecodeJWT_InvalidFormat() {
	testCases := []struct {
		name  string
		token string
	}{
		{
			name:  "TwoParts",
			token: "header.payload",
		},
		{
			name:  "FourParts",
			token: "header.payload.signature.extra",
		},
		{
			name:  "Empty",
			token: "",
		},
		{
			name:  "SinglePart",
			token: "header",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			_, _, err := DecodeJWT(tc.token)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid JWT format")
		})
	}
}

func (suite *JWTServiceTestSuite) TestDecodeJWT_InvalidBase64Header() {
	token := "invalid-base64.payload.signature"
	_, _, err := DecodeJWT(token)
	
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to decode JWT header")
}

func (suite *JWTServiceTestSuite) TestDecodeJWT_InvalidBase64Payload() {
	headerBytes, _ := json.Marshal(map[string]interface{}{"alg": "RS256"})
	headerB64 := base64.RawURLEncoding.EncodeToString(headerBytes)
	
	token := headerB64 + ".invalid-base64.signature"
	_, _, err := DecodeJWT(token)
	
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to decode JWT payload")
}

func (suite *JWTServiceTestSuite) TestDecodeJWT_InvalidJSONHeader() {
	headerB64 := base64.RawURLEncoding.EncodeToString([]byte("invalid json"))
	payloadBytes, _ := json.Marshal(map[string]interface{}{"sub": "test"})
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadBytes)
	
	token := headerB64 + "." + payloadB64 + ".signature"
	_, _, err := DecodeJWT(token)
	
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to unmarshal JWT header")
}

func (suite *JWTServiceTestSuite) TestDecodeJWT_InvalidJSONPayload() {
	headerBytes, _ := json.Marshal(map[string]interface{}{"alg": "RS256"})
	headerB64 := base64.RawURLEncoding.EncodeToString(headerBytes)
	payloadB64 := base64.RawURLEncoding.EncodeToString([]byte("invalid json"))
	
	token := headerB64 + "." + payloadB64 + ".signature"
	_, _, err := DecodeJWT(token)
	
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to unmarshal JWT payload")
}

func (suite *JWTServiceTestSuite) TestJWTServiceInterface() {
	// Test that JWTService implements JWTServiceInterface
	var _ JWTServiceInterface = &JWTService{}
}
