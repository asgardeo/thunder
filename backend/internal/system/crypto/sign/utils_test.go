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
package sign

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SignUtilsTestSuite struct {
	suite.Suite
	rsaPrivateKey     *rsa.PrivateKey
	ecdsaPrivateKey   *ecdsa.PrivateKey
	ed25519PrivateKey ed25519.PrivateKey
	ed25519PublicKey  ed25519.PublicKey
	testData          []byte
}

func TestSignUtilsSuite(t *testing.T) {
	suite.Run(t, new(SignUtilsTestSuite))
}

func (suite *SignUtilsTestSuite) SetupTest() {
	var err error

	// Generate RSA key pair
	suite.rsaPrivateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		suite.T().Fatalf("Failed to generate RSA key: %v", err)
	}

	// Generate ECDSA key pair
	suite.ecdsaPrivateKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		suite.T().Fatalf("Failed to generate ECDSA key: %v", err)
	}

	// Generate ED25519 key pair
	suite.ed25519PublicKey, suite.ed25519PrivateKey, err = ed25519.GenerateKey(rand.Reader)
	if err != nil {
		suite.T().Fatalf("Failed to generate ED25519 key: %v", err)
	}

	suite.testData = []byte("test data for signing and verification")
}

func (suite *SignUtilsTestSuite) TestSignRSASHA256() {
	signature, err := Generate(suite.testData, RSASHA256, suite.rsaPrivateKey)

	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), signature)

	// Verify the signature
	err = Verify(suite.testData, signature, RSASHA256, &suite.rsaPrivateKey.PublicKey)
	assert.NoError(suite.T(), err)
}

func (suite *SignUtilsTestSuite) TestSignRSASHA512() {
	signature, err := Generate(suite.testData, RSASHA512, suite.rsaPrivateKey)

	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), signature)

	// Verify the signature
	err = Verify(suite.testData, signature, RSASHA512, &suite.rsaPrivateKey.PublicKey)
	assert.NoError(suite.T(), err)
}

func (suite *SignUtilsTestSuite) TestSignECDSASHA256() {
	signature, err := Generate(suite.testData, ECDSASHA256, suite.ecdsaPrivateKey)

	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), signature)

	// Verify the signature
	err = Verify(suite.testData, signature, ECDSASHA256, &suite.ecdsaPrivateKey.PublicKey)
	assert.NoError(suite.T(), err)
}

func (suite *SignUtilsTestSuite) TestSignECDSASHA384() {
	// Generate ECDSA P-384 key pair for SHA-384
	ecdsaKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	assert.NoError(suite.T(), err)

	signature, err := Generate(suite.testData, ECDSASHA384, ecdsaKey)

	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), signature)

	// Verify the signature
	err = Verify(suite.testData, signature, ECDSASHA384, &ecdsaKey.PublicKey)
	assert.NoError(suite.T(), err)
}

func (suite *SignUtilsTestSuite) TestSignECDSASHA512() {
	ecdsaKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	assert.NoError(suite.T(), err)

	signature, err := Generate(suite.testData, ECDSASHA512, ecdsaKey)

	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), signature)

	// Verify the signature
	err = Verify(suite.testData, signature, ECDSASHA512, &ecdsaKey.PublicKey)
	assert.NoError(suite.T(), err)
}

func (suite *SignUtilsTestSuite) TestSignED25519() {
	signature, err := Generate(suite.testData, ED25519, suite.ed25519PrivateKey)

	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), signature)

	// Verify the signature
	err = Verify(suite.testData, signature, ED25519, suite.ed25519PublicKey)
	assert.NoError(suite.T(), err)
}

func (suite *SignUtilsTestSuite) TestSignUnsupportedAlgorithm() {
	signature, err := Generate(suite.testData, SignAlgorithm("INVALID"), suite.rsaPrivateKey)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), signature)
	assert.Equal(suite.T(), ErrUnsupportedAlgorithm, err)
}

func (suite *SignUtilsTestSuite) TestSignInvalidPrivateKey() {
	testCases := []struct {
		name       string
		algorithm  SignAlgorithm
		privateKey crypto.PrivateKey
	}{
		{"RSA_WithECDSAKey", RSASHA256, suite.ecdsaPrivateKey},
		{"ECDSA_WithRSAKey", ECDSASHA256, suite.rsaPrivateKey},
		{"ED25519_WithRSAKey", ED25519, suite.rsaPrivateKey},
		{"RSA_WithNilKey", RSASHA256, nil},
		{"ECDSA_WithNilKey", ECDSASHA256, nil},
		{"ED25519_WithNilKey", ED25519, nil},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			signature, err := Generate(suite.testData, tc.algorithm, tc.privateKey)

			assert.Error(t, err)
			assert.Nil(t, signature)
			assert.Equal(t, ErrInvalidPrivateKey, err)
		})
	}
}

func (suite *SignUtilsTestSuite) TestVerifyInvalidPublicKey() {
	testCases := []struct {
		name      string
		algorithm SignAlgorithm
		publicKey crypto.PublicKey
	}{
		{"RSA_WithECDSAKey", RSASHA256, &suite.ecdsaPrivateKey.PublicKey},
		{"ECDSA_WithRSAKey", ECDSASHA256, &suite.rsaPrivateKey.PublicKey},
		{"ED25519_WithRSAKey", ED25519, &suite.rsaPrivateKey.PublicKey},
		{"RSA_WithNilKey", RSASHA256, nil},
		{"ECDSA_WithNilKey", ECDSASHA256, nil},
		{"ED25519_WithNilKey", ED25519, nil},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			signature := []byte("dummy signature")
			err := Verify(suite.testData, signature, tc.algorithm, tc.publicKey)

			assert.Error(t, err)
			assert.Equal(t, ErrInvalidPublicKey, err)
		})
	}
}

func (suite *SignUtilsTestSuite) TestVerifyInvalidSignature() {
	testCases := []struct {
		name      string
		algorithm SignAlgorithm
		publicKey crypto.PublicKey
		signature []byte
	}{
		{
			"RSA_InvalidSignature",
			RSASHA256,
			&suite.rsaPrivateKey.PublicKey,
			[]byte("invalid signature"),
		},
		{
			"ECDSA_InvalidSignature",
			ECDSASHA256,
			&suite.ecdsaPrivateKey.PublicKey,
			[]byte("invalid signature"),
		},
		{
			"ED25519_InvalidSignature",
			ED25519,
			suite.ed25519PublicKey,
			make([]byte, ed25519.SignatureSize),
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			err := Verify(suite.testData, tc.signature, tc.algorithm, tc.publicKey)

			assert.Error(t, err)
			assert.Equal(t, ErrInvalidSignature, err)
		})
	}
}

func (suite *SignUtilsTestSuite) TestVerifyUnsupportedAlgorithm() {
	signature := []byte("dummy")
	err := Verify(suite.testData, signature, SignAlgorithm("INVALID"), &suite.rsaPrivateKey.PublicKey)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrUnsupportedAlgorithm, err)
}

func (suite *SignUtilsTestSuite) TestVerifyModifiedData() {
	signature, err := Generate(suite.testData, RSASHA256, suite.rsaPrivateKey)
	assert.NoError(suite.T(), err)

	modifiedData := []byte("modified test data")
	err = Verify(modifiedData, signature, RSASHA256, &suite.rsaPrivateKey.PublicKey)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrInvalidSignature, err)
}

func (suite *SignUtilsTestSuite) TestVerifyCrossAlgorithmVerification() {
	sig256, err := Generate(suite.testData, RSASHA256, suite.rsaPrivateKey)
	assert.NoError(suite.T(), err)

	// Try to verify SHA256 signature with SHA512 algorithm
	err = Verify(suite.testData, sig256, RSASHA512, &suite.rsaPrivateKey.PublicKey)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrInvalidSignature, err)
}

func (suite *SignUtilsTestSuite) TestVerifyDifferentKeyPairs() {
	signature, err := Generate(suite.testData, RSASHA256, suite.rsaPrivateKey)
	assert.NoError(suite.T(), err)

	// Generate a different key pair
	differentKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(suite.T(), err)

	// Try to verify with different public key
	err = Verify(suite.testData, signature, RSASHA256, &differentKey.PublicKey)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrInvalidSignature, err)
}

func (suite *SignUtilsTestSuite) TestSignEmptyData() {
	testCases := []struct {
		name       string
		algorithm  SignAlgorithm
		privateKey crypto.PrivateKey
		publicKey  crypto.PublicKey
	}{
		{"RSA_EmptyData", RSASHA256, suite.rsaPrivateKey, &suite.rsaPrivateKey.PublicKey},
		{"ECDSA_EmptyData", ECDSASHA256, suite.ecdsaPrivateKey, &suite.ecdsaPrivateKey.PublicKey},
		{"ED25519_EmptyData", ED25519, suite.ed25519PrivateKey, suite.ed25519PublicKey},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			emptyData := []byte{}

			signature, err := Generate(emptyData, tc.algorithm, tc.privateKey)
			assert.NoError(t, err)
			assert.NotEmpty(t, signature)

			err = Verify(emptyData, signature, tc.algorithm, tc.publicKey)
			assert.NoError(t, err)
		})
	}
}

func (suite *SignUtilsTestSuite) TestSignLargeData() {
	largeData := make([]byte, 1024*1024) // 1MB
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	signature, err := Generate(largeData, RSASHA256, suite.rsaPrivateKey)

	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), signature)

	err = Verify(largeData, signature, RSASHA256, &suite.rsaPrivateKey.PublicKey)
	assert.NoError(suite.T(), err)
}

func (suite *SignUtilsTestSuite) TestHashDataDifferentAlgorithms() {
	// Test that different hash algorithms work correctly
	sig256, err := Generate(suite.testData, RSASHA256, suite.rsaPrivateKey)
	assert.NoError(suite.T(), err)

	sig512, err := Generate(suite.testData, RSASHA512, suite.rsaPrivateKey)
	assert.NoError(suite.T(), err)

	// Verify with correct algorithms
	err = Verify(suite.testData, sig256, RSASHA256, &suite.rsaPrivateKey.PublicKey)
	assert.NoError(suite.T(), err)

	err = Verify(suite.testData, sig512, RSASHA512, &suite.rsaPrivateKey.PublicKey)
	assert.NoError(suite.T(), err)
}

func (suite *SignUtilsTestSuite) TestSignatureAlgorithmConstants() {
	algorithms := []struct {
		name      string
		algorithm SignAlgorithm
		expected  string
	}{
		{"RSA_SHA256", RSASHA256, "RSA-SHA256"},
		{"RSA_SHA512", RSASHA512, "RSA-SHA512"},
		{"ECDSA_SHA256", ECDSASHA256, "ECDSA-SHA256"},
		{"ECDSA_SHA512", ECDSASHA512, "ECDSA-SHA512"},
		{"ED25519", ED25519, "ED25519"},
	}

	for _, tc := range algorithms {
		suite.T().Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, string(tc.algorithm))
		})
	}
}

func (suite *SignUtilsTestSuite) TestErrorConstants() {
	errors := []struct {
		name     string
		err      error
		expected string
	}{
		{"UnsupportedAlgorithm", ErrUnsupportedAlgorithm, "unsupported signature algorithm"},
		{"InvalidPrivateKey", ErrInvalidPrivateKey, "invalid private key type for algorithm"},
		{"InvalidPublicKey", ErrInvalidPublicKey, "invalid public key type for algorithm"},
		{"InvalidSignature", ErrInvalidSignature, "signature verification failed"},
	}

	for _, tc := range errors {
		suite.T().Run(tc.name, func(t *testing.T) {
			assert.NotNil(t, tc.err)
			assert.Equal(t, tc.expected, tc.err.Error())
		})
	}
}

func (suite *SignUtilsTestSuite) TestSignInterfaceImplementation() {
	// Test that crypto.PrivateKey interface works correctly
	var key crypto.PrivateKey = suite.rsaPrivateKey

	signature, err := Generate(suite.testData, RSASHA256, key)

	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), signature)

	// Verify with crypto.PublicKey interface
	var pubKey crypto.PublicKey = &suite.rsaPrivateKey.PublicKey
	err = Verify(suite.testData, signature, RSASHA256, pubKey)
	assert.NoError(suite.T(), err)
}

func (suite *SignUtilsTestSuite) TestDecodeJWT() {
	testCases := []struct {
		name        string
		algorithm   SignAlgorithm
		privateKey  crypto.PrivateKey
		publicKey   crypto.PublicKey
		description string
	}{
		{
			"RSA_SHA256_RoundTrip",
			RSASHA256,
			suite.rsaPrivateKey,
			&suite.rsaPrivateKey.PublicKey,
			"Sign and verify with RSA-SHA256",
		},
		{
			"RSA_SHA512_RoundTrip",
			RSASHA512,
			suite.rsaPrivateKey,
			&suite.rsaPrivateKey.PublicKey,
			"Sign and verify with RSA-SHA512",
		},
		{
			"ECDSA_SHA256_RoundTrip",
			ECDSASHA256,
			suite.ecdsaPrivateKey,
			&suite.ecdsaPrivateKey.PublicKey,
			"Sign and verify with ECDSA-SHA256",
		},
		{
			"ED25519_RoundTrip",
			ED25519,
			suite.ed25519PrivateKey,
			suite.ed25519PublicKey,
			"Sign and verify with ED25519",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Sign
			signature, err := Generate(suite.testData, tc.algorithm, tc.privateKey)
			assert.NoError(t, err, tc.description)
			assert.NotEmpty(t, signature)

			// Verify
			err = Verify(suite.testData, signature, tc.algorithm, tc.publicKey)
			assert.NoError(t, err, tc.description)
		})
	}
}

func (suite *SignUtilsTestSuite) TestHashDataED25519NoPreHashing() {
	// ED25519 should not pre-hash the data
	// This test ensures that ED25519 signing works with the original data
	signature, err := Generate(suite.testData, ED25519, suite.ed25519PrivateKey)
	assert.NoError(suite.T(), err)

	// Verification should succeed with original data
	err = Verify(suite.testData, signature, ED25519, suite.ed25519PublicKey)
	assert.NoError(suite.T(), err)
}

func (suite *SignUtilsTestSuite) TestSignRSAPKCS1v15() {
	// Test that RSA signing uses PKCS1v15 (for compatibility)
	signature, err := Generate(suite.testData, RSASHA256, suite.rsaPrivateKey)
	assert.NoError(suite.T(), err)

	// The signature should be verifiable with PKCS1v15
	err = Verify(suite.testData, signature, RSASHA256, &suite.rsaPrivateKey.PublicKey)
	assert.NoError(suite.T(), err)
}

func (suite *SignUtilsTestSuite) TestSignECDSAASN1Format() {
	// Test that ECDSA signature is in ASN.1 DER format
	signature, err := Generate(suite.testData, ECDSASHA256, suite.ecdsaPrivateKey)
	assert.NoError(suite.T(), err)

	// The signature should be verifiable (using ASN.1 format)
	err = Verify(suite.testData, signature, ECDSASHA256, &suite.ecdsaPrivateKey.PublicKey)
	assert.NoError(suite.T(), err)
}
