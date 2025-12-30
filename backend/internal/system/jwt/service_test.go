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

package jwt

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/crypto/sign"
	"github.com/asgardeo/thunder/tests/mocks/crypto/pki/pkimock"
)

const (
	testAudience     = "test-audience"
	testIssuer       = "test-issuer"
	testAud          = "test-aud"
	testIss          = "test-iss"
	wrongAudience    = "wrong-audience"
	wrongIssuer      = "wrong-issuer"
	expectedAudience = "expected-audience"
	expectedIssuer   = "expected-issuer"
)

type JWTServiceTestSuite struct {
	suite.Suite
	jwtService     *JWTService
	testPrivateKey *rsa.PrivateKey
	testKeyPath    string
	tempFiles      []string
	pkiMock        *pkimock.PKIServiceInterfaceMock
}

func TestJWTServiceSuite(t *testing.T) {
	suite.Run(t, new(JWTServiceTestSuite))
}

func (suite *JWTServiceTestSuite) SetupSuite() {
	// Generate a test RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(suite.T(), err)
	suite.testPrivateKey = privateKey

	// Create a temporary private key file
	tempFile, err := os.CreateTemp("", "test_key_*.pem")
	assert.NoError(suite.T(), err)
	suite.testKeyPath = tempFile.Name()

	// Encode the private key to PEM
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// Write to file
	_, err = tempFile.Write(privateKeyPEM)
	assert.NoError(suite.T(), err)
	err = tempFile.Close()
	assert.NoError(suite.T(), err)
}

func (suite *JWTServiceTestSuite) TearDownSuite() {
	err := os.Remove(suite.testKeyPath)
	assert.NoError(suite.T(), err)
}

func (suite *JWTServiceTestSuite) AfterTest(_, _ string) {
	// Clean up any temporary files created during tests
	for _, file := range suite.tempFiles {
		err := os.Remove(file)
		if err != nil {
			suite.T().Logf("Failed to remove temp file %s: %v", file, err)
		}
	}
	suite.tempFiles = nil
}

func (suite *JWTServiceTestSuite) SetupTest() {
	// Reset ThunderRuntime before each test
	config.ResetThunderRuntime()

	// Create PKI mock
	suite.pkiMock = pkimock.NewPKIServiceInterfaceMock(suite.T())

	suite.jwtService = &JWTService{
		privateKey: suite.testPrivateKey,
		signAlg:    sign.RSASHA256,
		jwsAlg:     RS256,
		kid:        "test-kid",
	}

	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "https://test.thunder.io",
			ValidityPeriod: 3600, // Default validity period
			PreferredKeyID: "test-kid",
		},
		Crypto: config.CryptoConfig{
			Keys: []config.KeyConfig{
				{
					ID:       "test-kid",
					CertFile: suite.testKeyPath,
					KeyFile:  suite.testKeyPath,
				},
			},
		},
	}
	err := config.InitializeThunderRuntime("", testConfig)
	assert.NoError(suite.T(), err)
}

func (suite *JWTServiceTestSuite) TestNewJWTService() {
	// Set expectations for PKI interactions
	suite.pkiMock.EXPECT().GetPrivateKey(mock.Anything).Return(suite.testPrivateKey, nil)
	suite.pkiMock.EXPECT().GetCertThumbprint(mock.Anything).Return("test-kid")

	service, err := Initialize(suite.pkiMock)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), service)
	assert.Implements(suite.T(), (*JWTServiceInterface)(nil), service)
}

func (suite *JWTServiceTestSuite) TestInitScenarios() {
	testCases := []struct {
		name           string
		setupFunc      func() (string, *rsa.PrivateKey)
		expectSuccess  bool
		expectedErrMsg string
	}{
		{
			name: "Success",
			setupFunc: func() (string, *rsa.PrivateKey) {
				return suite.testKeyPath, suite.testPrivateKey // Use the existing valid key path
			},
			expectSuccess:  true,
			expectedErrMsg: "",
		},
		{
			name: "PKCS8Key",
			setupFunc: func() (string, *rsa.PrivateKey) {
				privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
				assert.NoError(suite.T(), err)

				pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
				assert.NoError(suite.T(), err)

				pkcs8KeyPEM := pem.EncodeToMemory(&pem.Block{
					Type:  "PRIVATE KEY", // This is the PKCS8 standard header
					Bytes: pkcs8Bytes,
				})

				tempFile, err := os.CreateTemp("", "pkcs8_key_*.pem")
				assert.NoError(suite.T(), err)
				suite.tempFiles = append(suite.tempFiles, tempFile.Name())

				_, err = tempFile.Write(pkcs8KeyPEM)
				assert.NoError(suite.T(), err)
				err = tempFile.Close()
				assert.NoError(suite.T(), err)

				return tempFile.Name(), privateKey
			},
			expectSuccess:  true,
			expectedErrMsg: "",
		},
		{
			name: "PrivateKeyRetrievalError",
			setupFunc: func() (string, *rsa.PrivateKey) {
				return suite.testKeyPath, suite.testPrivateKey
			},
			expectSuccess:  false,
			expectedErrMsg: "test error",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			_, privateKey := tc.setupFunc()

			// Create a new mock for each test case
			pkiMock := pkimock.NewPKIServiceInterfaceMock(t)

			if tc.name == "PrivateKeyRetrievalError" {
				pkiMock.EXPECT().GetPrivateKey(mock.Anything).Return(nil, fmt.Errorf("test error"))
			} else {
				pkiMock.EXPECT().GetPrivateKey(mock.Anything).Return(privateKey, nil)
				pkiMock.EXPECT().GetCertThumbprint(mock.Anything).Return("test-kid")
			}

			jwtService, err := Initialize(pkiMock)

			if tc.expectSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, jwtService)
			} else {
				assert.Error(t, err)
				if tc.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tc.expectedErrMsg)
				}
			}
		})
	}
}

func (suite *JWTServiceTestSuite) TestGetPublicKey() {
	testCases := []struct {
		name        string
		setupFunc   func() *JWTService
		expectValue bool
		expectedKey crypto.PublicKey
	}{
		{
			name: "WithValidKey",
			setupFunc: func() *JWTService {
				return suite.jwtService
			},
			expectValue: true,
			expectedKey: &suite.testPrivateKey.PublicKey,
		},
		{
			name: "WithNilKey",
			setupFunc: func() *JWTService {
				return &JWTService{
					privateKey: nil,
				}
			},
			expectValue: false,
			expectedKey: nil,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			jwtService := tc.setupFunc()
			publicKey := jwtService.GetPublicKey()

			if tc.expectValue {
				assert.NotNil(t, publicKey)
				if tc.expectedKey != nil {
					assert.Equal(t, tc.expectedKey, publicKey)
				}
			} else {
				assert.Nil(t, publicKey)
			}
		})
	}
}

func (suite *JWTServiceTestSuite) TestGenerateJWTScenarios() {
	testCases := []struct {
		name               string
		sub                string
		aud                string
		iss                string
		validity           int64
		claims             map[string]interface{}
		setupMock          func() func() // Returns cleanup function
		setupService       func() *JWTService
		expectError        bool
		errorContains      string
		validateSuccess    func(t *testing.T, token string, iat int64)
		useDefaultValidity bool
	}{
		{
			name:     "Success",
			sub:      "test-subject",
			aud:      testAudience,
			iss:      testIssuer,
			validity: 3600,
			claims: map[string]interface{}{
				"name":  "John Doe",
				"email": "john@example.com",
			},
			setupMock: func() func() {
				return func() {}
			},
			setupService: func() *JWTService {
				return suite.jwtService
			},
			expectError: false,
			validateSuccess: func(t *testing.T, token string, iat int64) {
				parts := strings.Split(token, ".")
				assert.Len(t, parts, 3)

				headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
				assert.NoError(t, err)

				var header map[string]string
				err = json.Unmarshal(headerBytes, &header)
				assert.NoError(t, err)

				assert.Equal(t, "RS256", header["alg"])
				assert.Equal(t, "JWT", header["typ"])
				assert.Equal(t, "test-kid", header["kid"])

				payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
				assert.NoError(t, err)

				var payload map[string]interface{}
				err = json.Unmarshal(payloadBytes, &payload)
				assert.NoError(t, err)

				assert.Equal(t, "test-subject", payload["sub"])
				assert.Equal(t, testAudience, payload["aud"])
				assert.Equal(t, testIssuer, payload["iss"])
				assert.NotEmpty(t, payload["jti"])

				// Check claims
				assert.Equal(t, "John Doe", payload["name"])
				assert.Equal(t, "john@example.com", payload["email"])

				assert.True(t, payload["exp"].(float64) > float64(time.Now().Unix()))
				assert.True(t, payload["exp"].(float64) <= float64(time.Now().Unix()+3600+5))
			},
		},
		{
			name:     "DefaultValidity",
			sub:      "test-subject",
			aud:      testAudience,
			iss:      testIssuer,
			validity: 0, // Should use default
			claims:   map[string]interface{}{},
			setupMock: func() func() {
				return func() {}
			},
			setupService: func() *JWTService {
				return suite.jwtService
			},
			expectError:        false,
			useDefaultValidity: true,
		},
		{
			name:     "DefaultIssuer",
			sub:      "test-subject",
			aud:      testAudience,
			iss:      "", // Should use default
			validity: 3600,
			claims:   map[string]interface{}{},
			setupMock: func() func() {
				return func() {}
			},
			setupService: func() *JWTService {
				return suite.jwtService
			},
			expectError: false,
		},
		{
			name:      "NilPrivateKey",
			sub:       "sub",
			aud:       "aud",
			iss:       "iss",
			validity:  3600,
			claims:    nil,
			setupMock: func() func() { return func() {} },
			setupService: func() *JWTService {
				return &JWTService{
					privateKey: nil,
				}
			},
			expectError:   true,
			errorContains: "private key not loaded",
		},
		{
			name:     "WithEmptyClaims",
			sub:      "test-subject",
			aud:      testAudience,
			iss:      testIssuer,
			validity: 1800,
			claims:   nil,
			setupMock: func() func() {
				return func() {}
			},
			setupService: func() *JWTService {
				return suite.jwtService
			},
			expectError: false,
		},
		{
			name:     "SigningError",
			sub:      "sub",
			aud:      "aud",
			iss:      "iss",
			validity: 3600,
			claims:   nil,
			setupMock: func() func() {
				return func() {}
			},
			setupService: func() *JWTService {
				return &JWTService{
					privateKey: &rsa.PrivateKey{}, // Invalid private key
				}
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			cleanup := tc.setupMock()
			defer cleanup() // Ensure cleanup runs regardless of test outcome

			jwtService := tc.setupService()

			token, iat, err := jwtService.GenerateJWT(tc.sub, tc.aud, tc.iss, tc.validity, tc.claims)

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
				assert.Empty(t, token)
				assert.Equal(t, int64(0), iat)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, token)
			assert.True(t, iat > 0)

			parts := strings.Split(token, ".")
			assert.Len(t, parts, 3)

			if tc.validateSuccess != nil {
				tc.validateSuccess(t, token, iat)
			}

			if tc.useDefaultValidity {
				payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
				assert.NoError(t, err)

				var payload map[string]interface{}
				err = json.Unmarshal(payloadBytes, &payload)
				assert.NoError(t, err)

				now := time.Now().Unix()
				assert.True(t, payload["exp"].(float64) >= float64(now+3600-5))
				assert.True(t, payload["exp"].(float64) <= float64(now+3600+5))
			}
		})
	}
}

func (suite *JWTServiceTestSuite) TestVerifyJWT() {
	testCases := []struct {
		name          string
		setupFunc     func() (string, string, string)
		expectError   bool
		errorContains string
	}{
		{
			name: "ValidJWT",
			setupFunc: func() (string, string, string) {
				aud := testAudience
				iss := testIssuer
				token := suite.createBasicJWT(aud, iss,
					time.Now().Add(time.Hour).Unix(), time.Now().Unix())
				return token, aud, iss
			},
			expectError: false,
		},
		{
			name: "ValidJWTWithEmptyExpectedAudience",
			setupFunc: func() (string, string, string) {
				iss := testIssuer
				token := suite.createBasicJWT("any-audience", iss,
					time.Now().Add(time.Hour).Unix(), time.Now().Unix())
				return token, "", iss
			},
			expectError: false,
		},
		{
			name: "ValidJWTWithEmptyExpectedIssuer",
			setupFunc: func() (string, string, string) {
				aud := testAudience
				token := suite.createBasicJWT(aud, "any-issuer",
					time.Now().Add(time.Hour).Unix(), time.Now().Unix())
				return token, aud, ""
			},
			expectError: false,
		},
		{
			name: "InvalidJWTFormat",
			setupFunc: func() (string, string, string) {
				return suite.createMalformedJWT(), testAud, testIss
			},
			expectError:   true,
			errorContains: "invalid JWT token format",
		},
		{
			name: "InvalidSignature",
			setupFunc: func() (string, string, string) {
				token := suite.createBasicJWT(testAud, testIss, time.Now().Add(time.Hour).Unix(), time.Now().Unix())
				parts := strings.Split(token, ".")
				if len(parts) == 3 {
					token = parts[0] + "." + parts[1] + ".invalidSignature123"
				}
				return token, testAud, testIss
			},
			expectError:   true,
			errorContains: "invalid token signature",
		},
		{
			name: "ExpiredToken",
			setupFunc: func() (string, string, string) {
				aud := testAudience
				iss := testIssuer
				expiredTime := time.Now().Add(-time.Hour).Unix()
				token := suite.createBasicJWT(aud, iss,
					expiredTime, time.Now().Add(-2*time.Hour).Unix())
				return token, aud, iss
			},
			expectError:   true,
			errorContains: "token has expired",
		},
		{
			name: "TokenNotValidYet",
			setupFunc: func() (string, string, string) {
				aud := testAudience
				iss := testIssuer
				futureTime := time.Now().Add(time.Hour).Unix()
				token := suite.createBasicJWT(aud, iss,
					time.Now().Add(2*time.Hour).Unix(), futureTime)
				return token, aud, iss
			},
			expectError:   true,
			errorContains: "token not valid yet (nbf)",
		},
		{
			name: "InvalidAudience",
			setupFunc: func() (string, string, string) {
				aud := wrongAudience
				iss := testIssuer
				token := suite.createBasicJWT(aud, iss,
					time.Now().Add(time.Hour).Unix(), time.Now().Unix())
				return token, expectedAudience, iss
			},
			expectError:   true,
			errorContains: "invalid audience",
		},
		{
			name: "InvalidIssuer",
			setupFunc: func() (string, string, string) {
				aud := testAudience
				iss := wrongIssuer
				token := suite.createBasicJWT(aud, iss,
					time.Now().Add(time.Hour).Unix(), time.Now().Unix())
				return token, aud, expectedIssuer
			},
			expectError:   true,
			errorContains: "invalid issuer",
		},
		{
			name: "PublicKeyNotAvailable",
			setupFunc: func() (string, string, string) {
				token := suite.createBasicJWT(testAudience, testIssuer,
					time.Now().Add(time.Hour).Unix(), time.Now().Unix())
				return token, testAudience, testIssuer
			},
			expectError:   true,
			errorContains: "private key not loaded",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			token, expectedAud, expectedIss := tc.setupFunc()

			jwtService := suite.jwtService
			if tc.name == "PublicKeyNotAvailable" {
				jwtService = &JWTService{
					privateKey: nil,
				}
			}

			err := jwtService.VerifyJWT(token, expectedAud, expectedIss)

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (suite *JWTServiceTestSuite) TestVerifyJWTWithPublicKey() {
	testCases := []struct {
		name          string
		setupFunc     func() (string, crypto.PublicKey, string, string)
		expectError   bool
		errorContains string
	}{
		{
			name: "ValidJWT",
			setupFunc: func() (string, crypto.PublicKey, string, string) {
				aud := testAudience
				iss := testIssuer
				token := suite.createBasicJWT(aud, iss,
					time.Now().Add(time.Hour).Unix(), time.Now().Unix())
				return token, &suite.testPrivateKey.PublicKey, aud, iss
			},
			expectError: false,
		},
		{
			name: "ValidJWTWithEmptyExpectedAudience",
			setupFunc: func() (string, crypto.PublicKey, string, string) {
				iss := testIssuer
				token := suite.createBasicJWT("any-audience", iss,
					time.Now().Add(time.Hour).Unix(), time.Now().Unix())
				return token, &suite.testPrivateKey.PublicKey, "", iss
			},
			expectError: false,
		},
		{
			name: "ValidJWTWithEmptyExpectedIssuer",
			setupFunc: func() (string, crypto.PublicKey, string, string) {
				aud := testAudience
				token := suite.createBasicJWT(aud, "any-issuer",
					time.Now().Add(time.Hour).Unix(), time.Now().Unix())
				return token, &suite.testPrivateKey.PublicKey, aud, ""
			},
			expectError: false,
		},
		{
			name: "InvalidJWTFormat",
			setupFunc: func() (string, crypto.PublicKey, string, string) {
				return suite.createMalformedJWT(), &suite.testPrivateKey.PublicKey, testAud, testIss
			},
			expectError:   true,
			errorContains: "invalid JWT token format",
		},
		{
			name: "InvalidSignature",
			setupFunc: func() (string, crypto.PublicKey, string, string) {
				token := suite.createBasicJWT(testAud, testIss, time.Now().Add(time.Hour).Unix(), time.Now().Unix())
				parts := strings.Split(token, ".")
				if len(parts) == 3 {
					token = parts[0] + "." + parts[1] + ".invalidSignature123"
				}
				return token, &suite.testPrivateKey.PublicKey, testAud, testIss
			},
			expectError:   true,
			errorContains: "invalid token signature",
		},
		{
			name: "ExpiredToken",
			setupFunc: func() (string, crypto.PublicKey, string, string) {
				aud := testAudience
				iss := testIssuer
				expiredTime := time.Now().Add(-time.Hour).Unix()
				token := suite.createBasicJWT(aud, iss,
					expiredTime, time.Now().Add(-2*time.Hour).Unix())
				return token, &suite.testPrivateKey.PublicKey, aud, iss
			},
			expectError:   true,
			errorContains: "token has expired",
		},
		{
			name: "TokenNotValidYet",
			setupFunc: func() (string, crypto.PublicKey, string, string) {
				aud := testAudience
				iss := testIssuer
				futureTime := time.Now().Add(time.Hour).Unix()
				token := suite.createBasicJWT(aud, iss,
					time.Now().Add(2*time.Hour).Unix(), futureTime)
				return token, &suite.testPrivateKey.PublicKey, aud, iss
			},
			expectError:   true,
			errorContains: "token not valid yet (nbf)",
		},
		{
			name: "InvalidAudience",
			setupFunc: func() (string, crypto.PublicKey, string, string) {
				aud := "wrong-audience"
				iss := testIssuer
				token := suite.createBasicJWT(aud, iss,
					time.Now().Add(time.Hour).Unix(), time.Now().Unix())
				return token, &suite.testPrivateKey.PublicKey, "expected-audience", iss
			},
			expectError:   true,
			errorContains: "invalid audience",
		},
		{
			name: "InvalidIssuer",
			setupFunc: func() (string, crypto.PublicKey, string, string) {
				aud := testAudience
				iss := "wrong-issuer"
				token := suite.createBasicJWT(aud, iss,
					time.Now().Add(time.Hour).Unix(), time.Now().Unix())
				return token, &suite.testPrivateKey.PublicKey, aud, "expected-issuer"
			},
			expectError:   true,
			errorContains: "invalid issuer",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			token, pubKey, expectedAud, expectedIss := tc.setupFunc()

			err := suite.jwtService.VerifyJWTWithPublicKey(token, pubKey, expectedAud, expectedIss)

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (suite *JWTServiceTestSuite) TestVerifyJWTWithJWKS() {
	testCases := []struct {
		name          string
		setupFunc     func() (string, string, string, string)
		expectError   bool
		errorContains string
	}{
		{
			name: "ValidJWTWithJWKS",
			setupFunc: func() (string, string, string, string) {
				aud := testAudience
				iss := testIssuer
				token := suite.createBasicJWT(aud, iss,
					time.Now().Add(time.Hour).Unix(), time.Now().Unix())

				mockServer := suite.mockJWKSServer()
				suite.T().Cleanup(mockServer.Close)

				return token, mockServer.URL, aud, iss
			},
			expectError: false,
		},
		{
			name: "ValidJWTWithEmptyExpectedClaims",
			setupFunc: func() (string, string, string, string) {
				token := suite.createBasicJWT("any-aud", "any-iss",
					time.Now().Add(time.Hour).Unix(), time.Now().Unix())

				mockServer := suite.mockJWKSServer()
				suite.T().Cleanup(mockServer.Close)

				return token, mockServer.URL, "", "" // Empty expected aud and iss
			},
			expectError: false,
		},
		{
			name: "InvalidJWTFormat",
			setupFunc: func() (string, string, string, string) {
				mockServer := suite.mockJWKSServer()
				suite.T().Cleanup(mockServer.Close)

				return suite.createMalformedJWT(), mockServer.URL, testAud, testIss
			},
			expectError:   true,
			errorContains: "invalid JWT token format",
		},
		{
			name: "InvalidSignatureWithJWKS",
			setupFunc: func() (string, string, string, string) {
				// Create a valid token first, then invalidate the signature
				token := suite.createBasicJWT(testAud, testIss, time.Now().Add(time.Hour).Unix(), time.Now().Unix())

				// Replace signature to make it invalid
				parts := strings.Split(token, ".")
				if len(parts) == 3 {
					token = parts[0] + "." + parts[1] + ".invalidSignature123"
				}

				mockServer := suite.mockJWKSServer()
				suite.T().Cleanup(mockServer.Close)

				return token, mockServer.URL, testAud, testIss
			},
			expectError:   true,
			errorContains: "invalid token signature",
		},
		{
			name: "ExpiredTokenWithJWKS",
			setupFunc: func() (string, string, string, string) {
				aud := testAudience
				iss := testIssuer
				expiredTime := time.Now().Add(-time.Hour).Unix() // Expired 1 hour ago
				token := suite.createBasicJWT(aud, iss,
					expiredTime, time.Now().Add(-2*time.Hour).Unix())

				mockServer := suite.mockJWKSServer()
				suite.T().Cleanup(mockServer.Close)

				return token, mockServer.URL, aud, iss
			},
			expectError:   true,
			errorContains: "token has expired",
		},
		{
			name: "TokenNotValidYetWithJWKS",
			setupFunc: func() (string, string, string, string) {
				aud := testAudience
				iss := testIssuer
				futureTime := time.Now().Add(time.Hour).Unix() // Valid 1 hour from now
				token := suite.createBasicJWT(aud, iss,
					time.Now().Add(2*time.Hour).Unix(), futureTime)

				mockServer := suite.mockJWKSServer()
				suite.T().Cleanup(mockServer.Close)

				return token, mockServer.URL, aud, iss
			},
			expectError:   true,
			errorContains: "token not valid yet (nbf)",
		},
		{
			name: "InvalidAudienceWithJWKS",
			setupFunc: func() (string, string, string, string) {
				aud := "wrong-audience"
				iss := testIssuer
				token := suite.createBasicJWT(aud, iss,
					time.Now().Add(time.Hour).Unix(), time.Now().Unix())

				mockServer := suite.mockJWKSServer()
				suite.T().Cleanup(mockServer.Close)

				return token, mockServer.URL, "expected-audience", iss
			},
			expectError:   true,
			errorContains: "invalid audience",
		},
		{
			name: "InvalidIssuerWithJWKS",
			setupFunc: func() (string, string, string, string) {
				aud := testAudience
				iss := "wrong-issuer"
				token := suite.createBasicJWT(aud, iss,
					time.Now().Add(time.Hour).Unix(), time.Now().Unix())

				mockServer := suite.mockJWKSServer()
				suite.T().Cleanup(mockServer.Close)

				return token, mockServer.URL, aud, "expected-issuer"
			},
			expectError:   true,
			errorContains: "invalid issuer",
		},
		{
			name: "JWKSNetworkError",
			setupFunc: func() (string, string, string, string) {
				token := suite.createBasicJWT(testAud, testIss,
					time.Now().Add(time.Hour).Unix(), time.Now().Unix())

				return token, "http://localhost:99999/invalid", testAud, testIss
			},
			expectError:   true,
			errorContains: "invalid token signature",
		},
		{
			name: "JWKSHTTPError",
			setupFunc: func() (string, string, string, string) {
				token := suite.createBasicJWT(testAud, testIss,
					time.Now().Add(time.Hour).Unix(), time.Now().Unix())

				// Create a server that returns 404
				errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				}))
				suite.T().Cleanup(errorServer.Close)

				return token, errorServer.URL, testAud, testIss
			},
			expectError:   true,
			errorContains: "invalid token signature",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			token, jwksURL, expectedAud, expectedIss := tc.setupFunc()

			err := suite.jwtService.VerifyJWTWithJWKS(token, jwksURL, expectedAud, expectedIss)

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (suite *JWTServiceTestSuite) TestVerifyJWTClaimsEdgeCases() {
	testCases := []struct {
		name          string
		setupFunc     func() string
		expectedAud   string
		expectedIss   string
		expectError   bool
		errorContains string
	}{
		{
			name: "MissingExpClaim",
			setupFunc: func() string {
				payload := map[string]interface{}{
					"sub": "test-subject",
					"aud": testAudience,
					"iss": testIssuer,
					"iat": time.Now().Unix(),
					"nbf": time.Now().Unix(),
					// Missing exp claim
				}

				header := map[string]interface{}{
					"alg": "RS256",
					"typ": "JWT",
					"kid": "test-kid",
				}

				headerJSON, _ := json.Marshal(header)
				payloadJSON, _ := json.Marshal(payload)

				headerBase64 := base64.RawURLEncoding.EncodeToString(headerJSON)
				payloadBase64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

				signingInput := headerBase64 + "." + payloadBase64
				signature, _ := sign.Generate([]byte(signingInput), sign.RSASHA256, suite.testPrivateKey)
				signatureBase64 := base64.RawURLEncoding.EncodeToString(signature)

				return headerBase64 + "." + payloadBase64 + "." + signatureBase64
			},
			expectedAud:   testAudience,
			expectedIss:   testIssuer,
			expectError:   true,
			errorContains: "missing or invalid 'exp' claim",
		},
		{
			name: "MissingNbfClaim",
			setupFunc: func() string {
				payload := map[string]interface{}{
					"sub": "test-subject",
					"aud": testAudience,
					"iss": testIssuer,
					"exp": time.Now().Add(time.Hour).Unix(),
					"iat": time.Now().Unix(),
					// Missing nbf claim
				}

				header := map[string]interface{}{
					"alg": "RS256",
					"typ": "JWT",
					"kid": "test-kid",
				}

				headerJSON, _ := json.Marshal(header)
				payloadJSON, _ := json.Marshal(payload)

				headerBase64 := base64.RawURLEncoding.EncodeToString(headerJSON)
				payloadBase64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

				signingInput := headerBase64 + "." + payloadBase64
				signature, _ := sign.Generate([]byte(signingInput), sign.RSASHA256, suite.testPrivateKey)
				signatureBase64 := base64.RawURLEncoding.EncodeToString(signature)

				return headerBase64 + "." + payloadBase64 + "." + signatureBase64
			},
			expectedAud:   testAudience,
			expectedIss:   testIssuer,
			expectError:   true,
			errorContains: "missing or invalid 'nbf' claim",
		},
		{
			name: "MissingAudClaim",
			setupFunc: func() string {
				payload := map[string]interface{}{
					"sub": "test-subject",
					"iss": testIssuer,
					"exp": time.Now().Add(time.Hour).Unix(),
					"iat": time.Now().Unix(),
					"nbf": time.Now().Unix(),
					// Missing aud claim
				}
				return suite.createJWTWithCustomPayload(payload)
			},
			expectedAud:   testAudience,
			expectedIss:   testIssuer,
			expectError:   true,
			errorContains: "missing or invalid 'aud' claim",
		},
		{
			name: "MissingIssClaim",
			setupFunc: func() string {
				payload := map[string]interface{}{
					"sub": "test-subject",
					"aud": testAudience,
					"exp": time.Now().Add(time.Hour).Unix(),
					"iat": time.Now().Unix(),
					"nbf": time.Now().Unix(),
					// Missing iss claim
				}
				return suite.createJWTWithCustomPayload(payload)
			},
			expectedAud:   testAudience,
			expectedIss:   testIssuer,
			expectError:   true,
			errorContains: "missing or invalid 'iss' claim",
		},
		{
			name: "InvalidExpClaimType",
			setupFunc: func() string {
				payload := map[string]interface{}{
					"sub": "test-subject",
					"aud": testAudience,
					"iss": testIssuer,
					"exp": "invalid-exp-type", // Wrong type
					"iat": time.Now().Unix(),
					"nbf": time.Now().Unix(),
				}

				header := map[string]interface{}{
					"alg": "RS256",
					"typ": "JWT",
					"kid": "test-kid",
				}

				headerJSON, _ := json.Marshal(header)
				payloadJSON, _ := json.Marshal(payload)

				headerBase64 := base64.RawURLEncoding.EncodeToString(headerJSON)
				payloadBase64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

				signingInput := headerBase64 + "." + payloadBase64
				signature, _ := sign.Generate([]byte(signingInput), sign.RSASHA256, suite.testPrivateKey)
				signatureBase64 := base64.RawURLEncoding.EncodeToString(signature)

				return headerBase64 + "." + payloadBase64 + "." + signatureBase64
			},
			expectedAud:   testAudience,
			expectedIss:   testIssuer,
			expectError:   true,
			errorContains: "missing or invalid 'exp' claim",
		},
		{
			name: "InvalidNbfClaimType",
			setupFunc: func() string {
				payload := map[string]interface{}{
					"sub": "test-subject",
					"aud": testAudience,
					"iss": testIssuer,
					"exp": time.Now().Add(time.Hour).Unix(),
					"iat": time.Now().Unix(),
					"nbf": "invalid-nbf-type", // Wrong type
				}

				header := map[string]interface{}{
					"alg": "RS256",
					"typ": "JWT",
					"kid": "test-kid",
				}

				headerJSON, _ := json.Marshal(header)
				payloadJSON, _ := json.Marshal(payload)

				headerBase64 := base64.RawURLEncoding.EncodeToString(headerJSON)
				payloadBase64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

				signingInput := headerBase64 + "." + payloadBase64
				signature, _ := sign.Generate([]byte(signingInput), sign.RSASHA256, suite.testPrivateKey)
				signatureBase64 := base64.RawURLEncoding.EncodeToString(signature)

				return headerBase64 + "." + payloadBase64 + "." + signatureBase64
			},
			expectedAud:   testAudience,
			expectedIss:   testIssuer,
			expectError:   true,
			errorContains: "missing or invalid 'nbf' claim",
		},
		{
			name: "InvalidAudClaimType",
			setupFunc: func() string {
				payload := map[string]interface{}{
					"sub": "test-subject",
					"aud": 12345, // Wrong type
					"iss": testIssuer,
					"exp": time.Now().Add(time.Hour).Unix(),
					"iat": time.Now().Unix(),
					"nbf": time.Now().Unix(),
				}
				return suite.createJWTWithCustomPayload(payload)
			},
			expectedAud:   testAudience,
			expectedIss:   testIssuer,
			expectError:   true,
			errorContains: "missing or invalid 'aud' claim",
		},
		{
			name: "InvalidIssClaimType",
			setupFunc: func() string {
				payload := map[string]interface{}{
					"sub": "test-subject",
					"aud": testAudience,
					"iss": 12345, // Wrong type
					"exp": time.Now().Add(time.Hour).Unix(),
					"iat": time.Now().Unix(),
					"nbf": time.Now().Unix(),
				}
				return suite.createJWTWithCustomPayload(payload)
			},
			expectedAud:   testAudience,
			expectedIss:   testIssuer,
			expectError:   true,
			errorContains: "missing or invalid 'iss' claim",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			token := tc.setupFunc()
			publicKey := &suite.testPrivateKey.PublicKey

			err := suite.jwtService.VerifyJWTWithPublicKey(token, publicKey, tc.expectedAud, tc.expectedIss)

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (suite *JWTServiceTestSuite) TestVerifyJWTSignature() {
	testCases := []struct {
		name        string
		setupFunc   func() string
		expectError bool
	}{
		{
			name: "ValidToken",
			setupFunc: func() string {
				token, _, err := suite.jwtService.GenerateJWT("test-subject", testAudience, testIssuer, 3600, nil)
				assert.NoError(suite.T(), err)
				return token
			},
			expectError: false,
		},
		{
			name: "InvalidToken",
			setupFunc: func() string {
				return "invalid.token"
			},
			expectError: true,
		},
		{
			name: "TamperedToken",
			setupFunc: func() string {
				parts := []string{}
				for _, part := range []string{"header", "payload", "signature"} {
					jsonData, _ := json.Marshal(map[string]string{"tampered": part})
					parts = append(parts, base64.RawURLEncoding.EncodeToString(jsonData))
				}
				return parts[0] + "." + parts[1] + "." + parts[2]
			},
			expectError: true,
		},
		{
			name: "PublicKeyNotAvailable",
			setupFunc: func() string {
				token, _, err := suite.jwtService.GenerateJWT("test-subject", testAudience, testIssuer, 3600, nil)
				assert.NoError(suite.T(), err)
				return token
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			token := tc.setupFunc()

			jwtService := suite.jwtService
			if tc.name == "PublicKeyNotAvailable" {
				jwtService = &JWTService{
					privateKey: nil,
				}
			}

			err := jwtService.VerifyJWTSignature(token)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (suite *JWTServiceTestSuite) TestVerifyJWTSignatureWithPublicKey() {
	validToken, _, err := suite.jwtService.GenerateJWT("test-subject", testAudience, testIssuer, 3600, nil)
	assert.NoError(suite.T(), err)

	wrongKey, _ := rsa.GenerateKey(rand.Reader, 2048)

	parts := []string{}
	for _, part := range []string{"header", "payload", "signature"} {
		jsonData, _ := json.Marshal(map[string]string{"tampered": part})
		parts = append(parts, base64.RawURLEncoding.EncodeToString(jsonData))
	}
	tamperedToken := parts[0] + "." + parts[1] + "." + parts[2]

	testCases := []struct {
		name        string
		token       string
		publicKey   crypto.PublicKey
		expectError bool
	}{
		{"ValidToken", validToken, &suite.testPrivateKey.PublicKey, false},
		{"WrongKey", validToken, &wrongKey.PublicKey, true},
		{"InvalidToken", "invalid.token", &suite.testPrivateKey.PublicKey, true},
		{"TamperedToken", tamperedToken, &suite.testPrivateKey.PublicKey, true},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			err := suite.jwtService.VerifyJWTSignatureWithPublicKey(tc.token, tc.publicKey)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (suite *JWTServiceTestSuite) TestVerifyJWTSignatureWithJWKS() {
	token, _, err := suite.jwtService.GenerateJWT("test-subject", testAudience, testIssuer, 3600, nil)
	assert.NoError(suite.T(), err)

	testServer := suite.mockJWKSServer()
	defer testServer.Close()

	err = suite.jwtService.VerifyJWTSignatureWithJWKS(token, testServer.URL)
	assert.NoError(suite.T(), err)
}

func (suite *JWTServiceTestSuite) TestVerifyJWTSignatureWithJWKSInvalidToken() {
	testServer := suite.mockJWKSServer()
	defer testServer.Close()

	testCases := []struct {
		name  string
		token string
	}{
		{"EmptyToken", ""},
		{"MalformedToken", "not.valid.jwt"},
		{"InvalidFormat", "header.payload"},                 // Missing signature part
		{"CorruptedHeader", "aGVhZGVyCg.payload.signature"}, // Non-decodable header
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			err := suite.jwtService.VerifyJWTSignatureWithJWKS(tc.token, testServer.URL)
			assert.Error(t, err)
		})
	}
}

func (suite *JWTServiceTestSuite) TestVerifyJWTSignatureWithJWKSKeyIDNotFound() {
	testServer := suite.mockJWKSServer()
	defer testServer.Close()

	nonExistentKidJWT := suite.createJWTWithCustomHeader(map[string]interface{}{
		"alg": "RS256",
		"typ": "JWT",
		"kid": "non-existent-key-id",
	})

	err := suite.jwtService.VerifyJWTSignatureWithJWKS(nonExistentKidJWT, testServer.URL)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "no matching key found")
}

func (suite *JWTServiceTestSuite) TestVerifyJWTSignatureWithJWKSNoKeyID() {
	testServer := suite.mockJWKSServer()
	defer testServer.Close()

	noKidJWT := suite.createJWTWithCustomHeader(map[string]interface{}{
		"alg": "RS256",
		"typ": "JWT",
		// No kid field
	})

	err := suite.jwtService.VerifyJWTSignatureWithJWKS(noKidJWT, testServer.URL)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "missing 'kid' claim")
}

func (suite *JWTServiceTestSuite) TestVerifyJWTSignatureWithJWKSHTTPErrors() {
	testCases := []struct {
		name          string
		setupServer   func() *httptest.Server
		setupToken    func() string
		expectedError string
	}{
		{
			name: "HTTPError404",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				}))
			},
			setupToken: func() string {
				token, _, _ := suite.jwtService.GenerateJWT("test-subject", testAudience, testIssuer, 3600, nil)
				return token
			},
			expectedError: "failed to fetch JWKS, status code: 404",
		},
		{
			name: "InvalidJSONResponse",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					if _, err := w.Write([]byte("invalid json")); err != nil {
						suite.T().Errorf("Failed to write response: %v", err)
					}
				}))
			},
			setupToken: func() string {
				token, _, _ := suite.jwtService.GenerateJWT("test-subject", testAudience, testIssuer, 3600, nil)
				return token
			},
			expectedError: "failed to parse JWKS",
		},
		{
			name: "JWKSWithoutMatchingKid",
			setupServer: func() *httptest.Server {
				// Create JWKS with different kid
				jwks := map[string]interface{}{
					"keys": []interface{}{
						map[string]interface{}{
							"kty": "RSA",
							"kid": "different-kid",
							"n":   "some-n",
							"e":   "AQAB",
						},
					},
				}
				jwksData, _ := json.Marshal(jwks)
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					if _, err := w.Write(jwksData); err != nil {
						suite.T().Errorf("Failed to write response: %v", err)
					}
				}))
			},
			setupToken: func() string {
				token, _, _ := suite.jwtService.GenerateJWT("test-subject", testAudience, testIssuer, 3600, nil)
				return token
			},
			expectedError: "no matching key found",
		},
		{
			name: "InvalidJWKFormat",
			setupServer: func() *httptest.Server {
				// Create JWKS with invalid JWK (missing n and e)
				jwks := map[string]interface{}{
					"keys": []interface{}{
						map[string]interface{}{
							"kty": "RSA",
							"kid": "test-kid",
							// Missing n and e
						},
					},
				}
				jwksData, _ := json.Marshal(jwks)
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					if _, err := w.Write(jwksData); err != nil {
						suite.T().Errorf("Failed to write response: %v", err)
					}
				}))
			},
			setupToken: func() string {
				token, _, _ := suite.jwtService.GenerateJWT("test-subject", testAudience, testIssuer, 3600, nil)
				return token
			},
			expectedError: "failed to convert JWK to public key",
		},
		{
			name: "InvalidTokenSignature",
			setupServer: func() *httptest.Server {
				return suite.mockJWKSServer()
			},
			setupToken: func() string {
				// Create a token with wrong signature
				token := suite.createJWTWithCustomHeader(map[string]interface{}{
					"alg": "RS256",
					"typ": "JWT",
					"kid": "test-kid",
				})
				// Modify the last part (signature) to make it invalid
				parts := strings.Split(token, ".")
				parts[2] = "invalid-signature"
				return strings.Join(parts, ".")
			},
			expectedError: "invalid token signature",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			testServer := tc.setupServer()
			defer testServer.Close()

			token := tc.setupToken()

			err := suite.jwtService.VerifyJWTSignatureWithJWKS(token, testServer.URL)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

func (suite *JWTServiceTestSuite) TestVerifyJWTSignatureWithJWKSNetworkError() {
	// Test with invalid URL to trigger network error
	token, _, err := suite.jwtService.GenerateJWT("test-subject", testAudience, testIssuer, 3600, nil)
	assert.NoError(suite.T(), err)

	err = suite.jwtService.VerifyJWTSignatureWithJWKS(token, "http://localhost:99999/invalid")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to fetch JWKS")
}

// Helper method to create a JWT with a custom header
func (suite *JWTServiceTestSuite) createJWTWithCustomHeader(header map[string]interface{}) string {
	// Create payload
	payload := map[string]interface{}{
		"sub":  "1234567890",
		"name": "Test User",
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(time.Hour).Unix(),
	}

	headerJSON, _ := json.Marshal(header)
	payloadJSON, _ := json.Marshal(payload)

	// Encode header and payload
	headerBase64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadBase64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

	// Create signature input
	signingInput := headerBase64 + "." + payloadBase64

	// Sign
	signature, err := sign.Generate([]byte(signingInput), sign.RSASHA256, suite.testPrivateKey)
	if err != nil {
		suite.T().Fatalf("Failed to sign JWT: %v", err)
	}

	// Encode signature
	signatureBase64 := base64.RawURLEncoding.EncodeToString(signature)

	// Create full JWT
	return headerBase64 + "." + payloadBase64 + "." + signatureBase64
}

// Helper method to create mock JWKS data
func (suite *JWTServiceTestSuite) createMockJWKSData() string {
	n := base64.RawURLEncoding.EncodeToString(suite.testPrivateKey.PublicKey.N.Bytes())

	// Convert exponent to bytes
	eBytes := []byte{1, 0, 1} // 65537 in big-endian
	e := base64.RawURLEncoding.EncodeToString(eBytes)

	jwk := map[string]interface{}{
		"kty": "RSA",
		"n":   n,
		"e":   e,
		"kid": "test-kid",
		"use": "sig",
		"alg": "RS256",
	}

	jwks := map[string]interface{}{
		"keys": []interface{}{jwk},
	}

	jwksData, _ := json.Marshal(jwks)
	return string(jwksData)
}

// Helper method to mock a JWKS server
func (suite *JWTServiceTestSuite) mockJWKSServer() *httptest.Server {
	jwksData := suite.createMockJWKSData()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := fmt.Fprintln(w, jwksData); err != nil {
			suite.T().Errorf("Failed to write JWKS response: %v", err)
		}
	}))

	return server
}

// Helper method to create a JWT with custom claims and validity
func (suite *JWTServiceTestSuite) createJWTWithClaims(sub, aud, iss string, exp int64, nbf int64,
	customClaims map[string]interface{}) string {
	// Create payload
	payload := map[string]interface{}{
		"sub": sub,
		"aud": aud,
		"iss": iss,
		"exp": exp,
		"iat": time.Now().Unix(),
		"nbf": nbf,
		"jti": "test-jti-" + fmt.Sprintf("%d", time.Now().UnixNano()),
	}

	// Add custom claims if provided
	for k, v := range customClaims {
		payload[k] = v
	}

	// Create header
	header := map[string]interface{}{
		"alg": "RS256",
		"typ": "JWT",
		"kid": "test-kid",
	}

	headerJSON, _ := json.Marshal(header)
	payloadJSON, _ := json.Marshal(payload)

	// Encode header and payload
	headerBase64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadBase64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

	// Create signature input
	signingInput := headerBase64 + "." + payloadBase64

	// Sign
	signature, err := sign.Generate([]byte(signingInput), sign.RSASHA256, suite.testPrivateKey)
	if err != nil {
		suite.T().Fatalf("Failed to sign JWT: %v", err)
	}

	// Encode signature
	signatureBase64 := base64.RawURLEncoding.EncodeToString(signature)

	// Create full JWT
	return headerBase64 + "." + payloadBase64 + "." + signatureBase64
}

// Helper method to create an invalid JWT (malformed)
func (suite *JWTServiceTestSuite) createMalformedJWT() string {
	return "invalid.jwt"
}

// Helper method to create a JWT with custom payload for testing edge cases
func (suite *JWTServiceTestSuite) createJWTWithCustomPayload(payload map[string]interface{}) string {
	header := map[string]interface{}{
		"alg": "RS256",
		"typ": "JWT",
		"kid": "test-kid",
	}

	headerJSON, _ := json.Marshal(header)
	payloadJSON, _ := json.Marshal(payload)

	headerBase64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadBase64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

	signingInput := headerBase64 + "." + payloadBase64
	signature, _ := sign.Generate([]byte(signingInput), sign.RSASHA256, suite.testPrivateKey)
	signatureBase64 := base64.RawURLEncoding.EncodeToString(signature)

	return headerBase64 + "." + payloadBase64 + "." + signatureBase64
}

// Helper method to create a JWT with basic claims for testing
func (suite *JWTServiceTestSuite) createBasicJWT(aud, iss string, exp int64, nbf int64) string {
	return suite.createJWTWithClaims("test-subject", aud, iss, exp, nbf, nil)
}

func (suite *JWTServiceTestSuite) TestInitWithECDSAKeys() {
	testCases := []struct {
		name            string
		curve           elliptic.Curve
		expectedAlg     JWSAlgorithm
		expectedSignAlg sign.SignAlgorithm
	}{
		{
			name:            "P256Key",
			curve:           elliptic.P256(),
			expectedAlg:     ES256,
			expectedSignAlg: sign.ECDSASHA256,
		},
		{
			name:            "P384Key",
			curve:           elliptic.P384(),
			expectedAlg:     ES384,
			expectedSignAlg: sign.ECDSASHA384,
		},
		{
			name:            "P521Key",
			curve:           elliptic.P521(),
			expectedAlg:     ES512,
			expectedSignAlg: sign.ECDSASHA512,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Generate ECDSA key
			ecKey, err := ecdsa.GenerateKey(tc.curve, rand.Reader)
			assert.NoError(t, err)

			// Marshal to PKCS8
			pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(ecKey)
			assert.NoError(t, err)

			keyPEM := pem.EncodeToMemory(&pem.Block{
				Type:  "PRIVATE KEY",
				Bytes: pkcs8Bytes,
			})

			// Write to temp file
			tempFile, err := os.CreateTemp("", "ec_key_*.pem")
			assert.NoError(t, err)
			defer func() {
				if err := os.Remove(tempFile.Name()); err != nil {
					t.Logf("Failed to remove temp file: %v", err)
				}
			}()

			_, err = tempFile.Write(keyPEM)
			assert.NoError(t, err)
			err = tempFile.Close()
			assert.NoError(t, err)

			// Initialize JWT service with mock
			pkiMock := pkimock.NewPKIServiceInterfaceMock(t)
			pkiMock.EXPECT().GetPrivateKey(mock.Anything).Return(ecKey, nil)
			pkiMock.EXPECT().GetCertThumbprint(mock.Anything).Return("test-kid")

			jwtService, err := Initialize(pkiMock)

			assert.NoError(t, err)
			assert.NotNil(t, jwtService)

			// Cast to access internal fields for testing
			jwtSvc, ok := jwtService.(*JWTService)
			assert.True(t, ok)
			assert.NotNil(t, jwtSvc.privateKey)
			assert.Equal(t, tc.expectedSignAlg, jwtSvc.signAlg)
			assert.Equal(t, tc.expectedAlg, jwtSvc.jwsAlg)

			// Test JWT generation with ECDSA key
			token, _, err := jwtService.GenerateJWT("test-subject", "test-aud", "test-iss", 3600, nil)
			assert.NoError(t, err)
			assert.NotEmpty(t, token)

			// Verify token header has correct alg
			header, err := DecodeJWTHeader(token)
			assert.NoError(t, err)
			assert.Equal(t, string(tc.expectedAlg), header["alg"])

			// Verify signature
			err = jwtService.VerifyJWTSignature(token)
			assert.NoError(t, err)
		})
	}
}

func (suite *JWTServiceTestSuite) TestInitWithEd25519Key() {
	// Generate Ed25519 key
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	assert.NoError(suite.T(), err)
	_ = pub // silence unused

	// Marshal to PKCS8
	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(priv)
	assert.NoError(suite.T(), err)

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	})

	// Write to temp file
	tempFile, err := os.CreateTemp("", "ed25519_key_*.pem")
	assert.NoError(suite.T(), err)
	defer func() {
		if err := os.Remove(tempFile.Name()); err != nil {
			suite.T().Logf("Failed to remove temp file: %v", err)
		}
	}()

	_, err = tempFile.Write(keyPEM)
	assert.NoError(suite.T(), err)
	err = tempFile.Close()
	assert.NoError(suite.T(), err)

	pkiMock := pkimock.NewPKIServiceInterfaceMock(suite.T())
	pkiMock.EXPECT().GetPrivateKey(mock.Anything).Return(priv, nil)
	pkiMock.EXPECT().GetCertThumbprint(mock.Anything).Return("test-kid")

	jwtService, err := Initialize(pkiMock)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), jwtService)

	// Cast to access internal fields for testing
	jwtSvc, ok := jwtService.(*JWTService)
	assert.True(suite.T(), ok)
	assert.NotNil(suite.T(), jwtSvc.privateKey)
	assert.Equal(suite.T(), sign.ED25519, jwtSvc.signAlg)
	assert.Equal(suite.T(), EdDSA, jwtSvc.jwsAlg)

	// Test JWT generation with Ed25519 key
	token, _, err := jwtService.GenerateJWT("test-subject", "test-aud", "test-iss", 3600, nil)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), token)

	// Verify token header has correct alg
	header, err := DecodeJWTHeader(token)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "EdDSA", header["alg"])

	// Verify signature
	err = jwtService.VerifyJWTSignature(token)
	assert.NoError(suite.T(), err)
}

func (suite *JWTServiceTestSuite) TestGetPublicKeyForAllKeyTypes() {
	testCases := []struct {
		name        string
		setupKey    func() (crypto.PrivateKey, sign.SignAlgorithm, JWSAlgorithm)
		validatePub func(t *testing.T, pub crypto.PublicKey)
	}{
		{
			name: "RSAKey",
			setupKey: func() (crypto.PrivateKey, sign.SignAlgorithm, JWSAlgorithm) {
				key, _ := rsa.GenerateKey(rand.Reader, 2048)
				return key, sign.RSASHA256, RS256
			},
			validatePub: func(t *testing.T, pub crypto.PublicKey) {
				_, ok := pub.(*rsa.PublicKey)
				assert.True(t, ok, "Expected RSA public key")
			},
		},
		{
			name: "ECDSAKey",
			setupKey: func() (crypto.PrivateKey, sign.SignAlgorithm, JWSAlgorithm) {
				key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
				return key, sign.ECDSASHA256, ES256
			},
			validatePub: func(t *testing.T, pub crypto.PublicKey) {
				_, ok := pub.(*ecdsa.PublicKey)
				assert.True(t, ok, "Expected ECDSA public key")
			},
		},
		{
			name: "Ed25519Key",
			setupKey: func() (crypto.PrivateKey, sign.SignAlgorithm, JWSAlgorithm) {
				_, priv, _ := ed25519.GenerateKey(rand.Reader)
				return priv, sign.ED25519, EdDSA
			},
			validatePub: func(t *testing.T, pub crypto.PublicKey) {
				_, ok := pub.(ed25519.PublicKey)
				assert.True(t, ok, "Expected Ed25519 public key")
			},
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			priv, signAlg, jwsAlg := tc.setupKey()
			jwtService := &JWTService{
				privateKey: priv,
				signAlg:    signAlg,
				jwsAlg:     jwsAlg,
			}

			pub := jwtService.GetPublicKey()
			assert.NotNil(t, pub)
			tc.validatePub(t, pub)
		})
	}
}

func (suite *JWTServiceTestSuite) TestInitWithECPrivateKeyFormat() {
	// Test EC PRIVATE KEY format (not PKCS8)
	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(suite.T(), err)

	// Marshal as EC PRIVATE KEY (not PKCS8)
	ecBytes, err := x509.MarshalECPrivateKey(ecKey)
	assert.NoError(suite.T(), err)

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: ecBytes,
	})

	tempFile, err := os.CreateTemp("", "ec_priv_key_*.pem")
	assert.NoError(suite.T(), err)
	defer func() {
		if err := os.Remove(tempFile.Name()); err != nil {
			suite.T().Logf("Failed to remove temp file: %v", err)
		}
	}()

	_, err = tempFile.Write(keyPEM)
	assert.NoError(suite.T(), err)
	err = tempFile.Close()
	assert.NoError(suite.T(), err)

	pkiMock := pkimock.NewPKIServiceInterfaceMock(suite.T())
	pkiMock.EXPECT().GetPrivateKey(mock.Anything).Return(ecKey, nil)
	pkiMock.EXPECT().GetCertThumbprint(mock.Anything).Return("test-kid")

	jwtService, err := Initialize(pkiMock)

	assert.NoError(suite.T(), err)

	// Cast to access internal fields for testing
	jwtSvc, ok := jwtService.(*JWTService)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), sign.ECDSASHA256, jwtSvc.signAlg)
	assert.Equal(suite.T(), ES256, jwtSvc.jwsAlg)
}

func (suite *JWTServiceTestSuite) TestInitWithUnsupportedECCurve() {
	// Generate P-224 key (unsupported curve)
	ecKey, err := ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	assert.NoError(suite.T(), err)

	ecBytes, err := x509.MarshalECPrivateKey(ecKey)
	assert.NoError(suite.T(), err)

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: ecBytes,
	})

	tempFile, err := os.CreateTemp("", "ec_unsupported_*.pem")
	assert.NoError(suite.T(), err)
	defer func() {
		if err := os.Remove(tempFile.Name()); err != nil {
			suite.T().Logf("Failed to remove temp file: %v", err)
		}
	}()

	_, err = tempFile.Write(keyPEM)
	assert.NoError(suite.T(), err)
	pkiMock := pkimock.NewPKIServiceInterfaceMock(suite.T())
	pkiMock.EXPECT().GetPrivateKey(mock.Anything).Return(nil, fmt.Errorf("unsupported EC curve"))
	_, err = Initialize(pkiMock)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "unsupported EC curve")
}

func (suite *JWTServiceTestSuite) TestMapJWSAlgToSignAlg() {
	testCases := []struct {
		name        string
		jwsAlg      string
		expectedAlg sign.SignAlgorithm
		expectError bool
	}{
		{"RS256", "RS256", sign.RSASHA256, false},
		{"RS512", "RS512", sign.RSASHA512, false},
		{"ES256", "ES256", sign.ECDSASHA256, false},
		{"ES384", "ES384", sign.ECDSASHA384, false},
		{"ES512", "ES512", sign.ECDSASHA512, false},
		{"EdDSA", "EdDSA", sign.ED25519, false},
		{"Unsupported", "HS256", "", true},
		{"Empty", "", "", true},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			alg, err := mapJWSAlgToSignAlg(JWSAlgorithm(tc.jwsAlg))
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedAlg, alg)
			}
		})
	}
}

func (suite *JWTServiceTestSuite) TestJWKToPublicKeyErrorCases() {
	testCases := []struct {
		name          string
		jwk           map[string]interface{}
		errorContains string
	}{
		{
			name:          "MissingKty",
			jwk:           map[string]interface{}{},
			errorContains: "JWK missing kty",
		},
		{
			name:          "InvalidKty",
			jwk:           map[string]interface{}{"kty": 123},
			errorContains: "JWK missing kty",
		},
		{
			name:          "UnsupportedKty",
			jwk:           map[string]interface{}{"kty": "oct"},
			errorContains: "unsupported JWK kty",
		},
		{
			name:          "RSA_MissingModulus",
			jwk:           map[string]interface{}{"kty": "RSA", "e": "AQAB"},
			errorContains: "JWK missing RSA modulus or exponent",
		},
		{
			name:          "RSA_MissingExponent",
			jwk:           map[string]interface{}{"kty": "RSA", "n": "test"},
			errorContains: "JWK missing RSA modulus or exponent",
		},
		{
			name:          "RSA_InvalidModulus",
			jwk:           map[string]interface{}{"kty": "RSA", "n": "invalid!base64", "e": "AQAB"},
			errorContains: "failed to decode RSA modulus",
		},
		{
			name:          "RSA_InvalidExponent",
			jwk:           map[string]interface{}{"kty": "RSA", "n": "AQAB", "e": "invalid!base64"},
			errorContains: "failed to decode RSA exponent",
		},
		{
			name:          "EC_MissingCurve",
			jwk:           map[string]interface{}{"kty": "EC", "x": "test", "y": "test"},
			errorContains: "JWK missing EC parameters",
		},
		{
			name:          "EC_MissingX",
			jwk:           map[string]interface{}{"kty": "EC", "crv": "P-256", "y": "test"},
			errorContains: "JWK missing EC parameters",
		},
		{
			name:          "EC_MissingY",
			jwk:           map[string]interface{}{"kty": "EC", "crv": "P-256", "x": "test"},
			errorContains: "JWK missing EC parameters",
		},
		{
			name:          "EC_UnsupportedCurve",
			jwk:           map[string]interface{}{"kty": "EC", "crv": "P-224", "x": "test", "y": "test"},
			errorContains: "unsupported EC curve",
		},
		{
			name:          "EC_InvalidX",
			jwk:           map[string]interface{}{"kty": "EC", "crv": "P-256", "x": "invalid!base64", "y": "AQAB"},
			errorContains: "failed to decode EC x",
		},
		{
			name:          "EC_InvalidY",
			jwk:           map[string]interface{}{"kty": "EC", "crv": "P-256", "x": "AQAB", "y": "invalid!base64"},
			errorContains: "failed to decode EC y",
		},
		{
			name: "EC_InvalidXLength",
			jwk: map[string]interface{}{
				"kty": "EC", "crv": "P-256",
				"x": base64.RawURLEncoding.EncodeToString([]byte{1}),        // 1 byte
				"y": base64.RawURLEncoding.EncodeToString(make([]byte, 32)), // 32 bytes
			},
			errorContains: "invalid EC coordinate length",
		},
		{
			name: "EC_InvalidYLength",
			jwk: map[string]interface{}{
				"kty": "EC", "crv": "P-256",
				"x": base64.RawURLEncoding.EncodeToString(make([]byte, 32)), // 32 bytes
				"y": base64.RawURLEncoding.EncodeToString([]byte{1}),        // 1 byte
			},
			errorContains: "invalid EC coordinate length",
		},
		{
			name: "EC_PointNotOnCurve",
			jwk: map[string]interface{}{
				"kty": "EC", "crv": "P-256",
				"x": base64.RawURLEncoding.EncodeToString(make([]byte, 32)), // 32 zero bytes
				"y": base64.RawURLEncoding.EncodeToString(make([]byte, 32)), // 32 zero bytes
			},
			errorContains: "EC point not on curve",
		},
		{
			name:          "OKP_MissingCurve",
			jwk:           map[string]interface{}{"kty": "OKP", "x": "test"},
			errorContains: "JWK missing OKP parameters",
		},
		{
			name:          "OKP_MissingX",
			jwk:           map[string]interface{}{"kty": "OKP", "crv": "Ed25519"},
			errorContains: "JWK missing OKP parameters",
		},
		{
			name:          "OKP_UnsupportedCurve",
			jwk:           map[string]interface{}{"kty": "OKP", "crv": "Ed448", "x": "test"},
			errorContains: "unsupported OKP curve",
		},
		{
			name:          "OKP_InvalidX",
			jwk:           map[string]interface{}{"kty": "OKP", "crv": "Ed25519", "x": "invalid!base64"},
			errorContains: "failed to decode Ed25519 x",
		},
		{
			name: "OKP_InvalidKeyLength",
			jwk: map[string]interface{}{
				"kty": "OKP", "crv": "Ed25519", "x": base64.RawURLEncoding.EncodeToString([]byte("short")),
			},
			errorContains: "invalid Ed25519 public key length",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			_, err := jwkToPublicKey(tc.jwk)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.errorContains)
		})
	}
}

func (suite *JWTServiceTestSuite) TestVerifyJWTSignatureWithPublicKeyAlgorithmDetection() {
	// Test that VerifyJWTSignatureWithPublicKey correctly detects algorithm from header
	testCases := []struct {
		name        string
		setupKey    func() (crypto.PrivateKey, sign.SignAlgorithm, JWSAlgorithm)
		expectError bool
	}{
		{
			name: "RS256Token",
			setupKey: func() (crypto.PrivateKey, sign.SignAlgorithm, JWSAlgorithm) {
				key, _ := rsa.GenerateKey(rand.Reader, 2048)
				return key, sign.RSASHA256, RS256
			},
			expectError: false,
		},
		{
			name: "ES256Token",
			setupKey: func() (crypto.PrivateKey, sign.SignAlgorithm, JWSAlgorithm) {
				key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
				return key, sign.ECDSASHA256, ES256
			},
			expectError: false,
		},
		{
			name: "EdDSAToken",
			setupKey: func() (crypto.PrivateKey, sign.SignAlgorithm, JWSAlgorithm) {
				_, priv, _ := ed25519.GenerateKey(rand.Reader)
				return priv, sign.ED25519, EdDSA
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			priv, signAlg, jwsAlg := tc.setupKey()
			jwtService := &JWTService{
				privateKey: priv,
				signAlg:    signAlg,
				jwsAlg:     jwsAlg,
			}

			// Generate token
			token, _, err := jwtService.GenerateJWT("test-sub", "test-aud", "test-iss", 3600, nil)
			assert.NoError(t, err)

			// Get public key
			pubKey := jwtService.GetPublicKey()
			assert.NotNil(t, pubKey)

			// Verify with public key (should detect algorithm from header)
			err = jwtService.VerifyJWTSignatureWithPublicKey(token, pubKey)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
