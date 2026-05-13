/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

package jwe

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"strings"
	"testing"

	"github.com/thunder-id/thunderid/internal/system/config"
	"github.com/thunder-id/thunderid/internal/system/cryptolab"
	"github.com/thunder-id/thunderid/internal/system/kmprovider"
	"github.com/thunder-id/thunderid/internal/system/log"
	"github.com/thunder-id/thunderid/tests/mocks/crypto/cryptomock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type JWEServiceTestSuite struct {
	suite.Suite
	jweService        *jweService
	testRSAPrivateKey *rsa.PrivateKey
	testECPrivateKey  *ecdsa.PrivateKey
}

func TestJWEServiceSuite(t *testing.T) {
	suite.Run(t, new(JWEServiceTestSuite))
}

func (suite *JWEServiceTestSuite) SetupTest() {
	config.ResetServerRuntime()

	rsaKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	suite.testRSAPrivateKey = rsaKey

	ecKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	suite.testECPrivateKey = ecKey

	testConfig := &config.Config{
		JWT: config.JWTConfig{
			PreferredKeyID: "test-kid",
		},
	}
	_ = config.InitializeServerRuntime("", testConfig)
}

func (suite *JWEServiceTestSuite) TestEncryptDecrypt_RSA() {
	encAlgs := []ContentEncAlgorithm{A128GCM, A192GCM, A256GCM}

	mockProvider := cryptomock.NewRuntimeCryptoProviderMock(suite.T())
	mockProvider.EXPECT().Decrypt(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(
			ctx context.Context, keyRef *kmprovider.KeyRef,
			params cryptolab.AlgorithmParams, content []byte,
		) ([]byte, error) {
			return cryptolab.Decrypt(suite.testRSAPrivateKey, params, content)
		}).Times(len(encAlgs))

	suite.jweService = &jweService{
		cryptoProvider: mockProvider,
		keyRef:         kmprovider.KeyRef{KeyID: "test-kid"},
		logger:         log.GetLogger(),
	}

	payload := []byte("Hello, RSA JWE!")
	recipientPublicKey := &suite.testRSAPrivateKey.PublicKey

	for _, enc := range encAlgs {
		jweToken, sErr := suite.jweService.Encrypt(payload, recipientPublicKey, RSAOAEP256, enc, "", "")
		assert.Nil(suite.T(), sErr)
		decrypted, sErr := suite.jweService.Decrypt(jweToken)
		assert.Nil(suite.T(), sErr)
		assert.Equal(suite.T(), payload, decrypted)
	}
}

func (suite *JWEServiceTestSuite) TestEncryptDecrypt_ECDH() {
	testCases := []struct {
		alg KeyEncAlgorithm
		enc ContentEncAlgorithm
	}{
		{ECDHES, A128GCM},
		{ECDHES, A192GCM},
		{ECDHES, A256GCM},
		{ECDHESA128KW, A128GCM},
		{ECDHESA256KW, A256GCM},
	}

	mockProvider := cryptomock.NewRuntimeCryptoProviderMock(suite.T())
	mockProvider.EXPECT().Decrypt(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(
			ctx context.Context, keyRef *kmprovider.KeyRef,
			params cryptolab.AlgorithmParams, content []byte,
		) ([]byte, error) {
			return cryptolab.Decrypt(suite.testECPrivateKey, params, content)
		}).Times(len(testCases))

	suite.jweService = &jweService{
		cryptoProvider: mockProvider,
		keyRef:         kmprovider.KeyRef{KeyID: "test-kid"},
		logger:         log.GetLogger(),
	}

	payload := []byte("Hello, ECDH JWE!")
	recipientPublicKey := &suite.testECPrivateKey.PublicKey

	for _, tc := range testCases {
		jweToken, sErr := suite.jweService.Encrypt(payload, recipientPublicKey, tc.alg, tc.enc, "", "")
		assert.Nil(suite.T(), sErr)
		decrypted, sErr := suite.jweService.Decrypt(jweToken)
		assert.Nil(suite.T(), sErr)
		assert.Equal(suite.T(), payload, decrypted)
	}
}

func (suite *JWEServiceTestSuite) TestEncrypt_Errors() {
	suite.jweService = &jweService{
		logger: log.GetLogger(),
	}

	// Unsupported Encryption algorithm
	_, sErr := suite.jweService.Encrypt([]byte("p"), &suite.testRSAPrivateKey.PublicKey, RSAOAEP256, "INVALID", "", "")
	assert.NotNil(suite.T(), sErr)
	assert.Equal(suite.T(), ErrorUnsupportedEncryptionAlgorithm, *sErr)

	// EncryptKey failure (RSA with EC key)
	_, sErr = suite.jweService.Encrypt([]byte("p"), &suite.testECPrivateKey.PublicKey, RSAOAEP256, A128GCM, "", "")
	assert.NotNil(suite.T(), sErr)
	assert.Equal(suite.T(), ErrorUnsupportedJWEAlgorithm, *sErr)
}

func (suite *JWEServiceTestSuite) TestDecrypt_Errors() {
	mockProvider := cryptomock.NewRuntimeCryptoProviderMock(suite.T())
	suite.jweService = &jweService{
		cryptoProvider: mockProvider,
		keyRef:         kmprovider.KeyRef{KeyID: "test-kid"},
		logger:         log.GetLogger(),
	}

	// Invalid JWE format — no provider call
	_, sErr := suite.jweService.Decrypt("invalid.jwe")
	assert.NotNil(suite.T(), sErr)
	assert.Equal(suite.T(), ErrorDecodingJWE, *sErr)

	// Encrypt a valid token (Encrypt does not use the provider)
	payload := []byte("data")
	jweToken, _ := suite.jweService.Encrypt(payload, &suite.testRSAPrivateKey.PublicKey, RSAOAEP256, A128GCM, "", "")

	// DecryptKey failure: provider returns an error
	mockProvider.EXPECT().Decrypt(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("key decryption error")).Once()
	_, sErr = suite.jweService.Decrypt(jweToken)
	assert.NotNil(suite.T(), sErr)
	assert.Equal(suite.T(), ErrorJWEDecryptionFailed, *sErr)

	// DecryptContent failure (tampered tag): provider returns correct CEK but tag is wrong
	mockProvider.EXPECT().Decrypt(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(
			ctx context.Context, keyRef *kmprovider.KeyRef,
			params cryptolab.AlgorithmParams, content []byte,
		) ([]byte, error) {
			return cryptolab.Decrypt(suite.testRSAPrivateKey, params, content)
		}).Once()
	parts := strings.Split(jweToken, ".")
	parts[4] = base64.RawURLEncoding.EncodeToString([]byte("wrong-tag"))
	tamperedToken := strings.Join(parts, ".")
	_, sErr = suite.jweService.Decrypt(tamperedToken)
	assert.NotNil(suite.T(), sErr)
	assert.Equal(suite.T(), ErrorJWEDecryptionFailed, *sErr)
}

func (suite *JWEServiceTestSuite) TestInitialize() {
	mockProvider := cryptomock.NewRuntimeCryptoProviderMock(suite.T())

	service, err := Initialize(mockProvider)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), service)
}

func (suite *JWEServiceTestSuite) TestEncrypt_ErrorCases() {
	suite.jweService = &jweService{
		logger: log.GetLogger(),
	}

	payload := []byte("test data")

	// Nil recipient key
	_, sErr := suite.jweService.Encrypt(payload, nil, RSAOAEP256, A128GCM, "", "")
	assert.NotNil(suite.T(), sErr)

	// Unsupported key type
	fakeKey := "not-a-real-key"
	_, sErr = suite.jweService.Encrypt(payload, fakeKey, RSAOAEP256, A128GCM, "", "")
	assert.NotNil(suite.T(), sErr)
}

func (suite *JWEServiceTestSuite) TestDecrypt_EdgeCases() {
	suite.jweService = &jweService{
		logger: log.GetLogger(),
	}

	// Test with malformed JWE (wrong number of parts)
	_, sErr := suite.jweService.Decrypt("malformed.jwe")
	assert.NotNil(suite.T(), sErr)
	assert.Equal(suite.T(), ErrorDecodingJWE, *sErr)

	// Test with invalid base64 in header
	_, sErr = suite.jweService.Decrypt("invalid-base64.key.iv.ciphertext.tag")
	assert.NotNil(suite.T(), sErr)
	assert.Equal(suite.T(), ErrorDecodingJWE, *sErr)

	// Test with invalid JSON in header
	invalidHeader := base64.RawURLEncoding.EncodeToString([]byte("{invalid json"))
	_, sErr = suite.jweService.Decrypt(invalidHeader + ".key.iv.ciphertext.tag")
	assert.NotNil(suite.T(), sErr)
	assert.Equal(suite.T(), ErrorDecodingJWE, *sErr)

	// Test with missing required header fields
	headerMissingAlg := base64.RawURLEncoding.EncodeToString([]byte(`{"enc":"A128GCM"}`))
	_, sErr = suite.jweService.Decrypt(headerMissingAlg + ".key.iv.ciphertext.tag")
	assert.NotNil(suite.T(), sErr)
	assert.Equal(suite.T(), ErrorUnsupportedJWEAlgorithm, *sErr)
}
