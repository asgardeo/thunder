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
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type JEWUtilsTestSuite struct {
	suite.Suite
}

func TestJEWUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(JEWUtilsTestSuite))
}

func (s *JEWUtilsTestSuite) TestEncryptDecryptKey_RSA() {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	s.NoError(err)
	publicKey := &privateKey.PublicKey

	cek := []byte("this-is-a-32-byte-long-cek-key!!") // 32 bytes

	testCases := []struct {
		name string
		alg  KeyEncAlgorithm
	}{
		{"RSA_OAEP_256", RSAOAEP256},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			encryptedKey, headerExtras, err := encryptKey(cek, tc.alg, publicKey, A128GCM)
			s.NoError(err)
			s.NotNil(encryptedKey)
			s.Nil(headerExtras)

			decryptedKey, err := decryptKey(encryptedKey, tc.alg, privateKey, nil, A128GCM)
			s.NoError(err)
			s.Equal(cek, decryptedKey)
		})
	}
}

func (s *JEWUtilsTestSuite) TestEncryptDecryptKey_ECDH() {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	s.NoError(err)
	publicKey := &privateKey.PublicKey

	testCases := []struct {
		name string
		alg  KeyEncAlgorithm
		enc  ContentEncAlgorithm
	}{
		{"ECDH_ES_A128GCM", ECDHES, A128GCM},
		{"ECDH_ES_A256GCM", ECDHES, A256GCM},
		{"ECDH_ES_A128KW_A128GCM", ECDHESA128KW, A128GCM},
		{"ECDH_ES_A256KW_A128GCM", ECDHESA256KW, A128GCM},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cekSize := 16
			if tc.enc == A256GCM || tc.alg == ECDHESA256KW {
				cekSize = 32
			}
			cek := make([]byte, cekSize)
			if tc.alg != ECDHES {
				_, _ = rand.Read(cek)
			}

			encryptedKey, headerExtras, err := encryptKey(cek, tc.alg, publicKey, tc.enc)
			s.NoError(err)
			s.NotNil(headerExtras["epk"])

			decryptedKey, err := decryptKey(encryptedKey, tc.alg, privateKey, headerExtras, tc.enc)
			s.NoError(err)
			s.Equal(cek, decryptedKey)
		})
	}
}

func (s *JEWUtilsTestSuite) TestEncryptDecryptKey_Curves() {
	curves := []struct {
		name string
		c    elliptic.Curve
	}{
		{"P-256", elliptic.P256()},
		{"P-384", elliptic.P384()},
		{"P-521", elliptic.P521()},
	}

	for _, tc := range curves {
		s.Run(tc.name, func() {
			priv, _ := ecdsa.GenerateKey(tc.c, rand.Reader)
			cek := make([]byte, 16)
			encryptedKey, headerExtras, err := encryptKey(cek, ECDHES, &priv.PublicKey, A128GCM)
			s.NoError(err)
			s.NotNil(headerExtras["epk"])

			decryptedKey, err := decryptKey(encryptedKey, ECDHES, priv, headerExtras, A128GCM)
			s.NoError(err)
			s.Equal(cek, decryptedKey)
		})
	}
}

func (s *JEWUtilsTestSuite) TestEncryptKey_Errors() {
	cek := []byte("cek")

	// Unsupported key type for RSA
	ecKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	_, _, err := encryptKey(cek, RSAOAEP256, &ecKey.PublicKey, A128GCM)
	s.Error(err)

	// Unsupported algorithm
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	_, _, err = encryptKey(cek, "INVALID_ALG", &privateKey.PublicKey, A128GCM)
	s.Error(err)

	// ECDH-ES with unsupported enc
	_, _, err = encryptKey(cek, ECDHES, &ecKey.PublicKey, "INVALID_ENC")
	s.Error(err)

	// ECDH-ES with non-ECDSA key
	_, _, err = encryptKey(cek, ECDHES, &privateKey.PublicKey, A128GCM)
	s.Error(err)
}

func (s *JEWUtilsTestSuite) TestDecryptKey_Errors() {
	encryptedKey := []byte("encrypted")

	// Unsupported key type for RSA
	ecKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	_, err := decryptKey(encryptedKey, RSAOAEP256, ecKey, nil, A128GCM)
	s.Error(err)

	// Unsupported algorithm
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	_, err = decryptKey(encryptedKey, "INVALID_ALG", privateKey, nil, A128GCM)
	s.Error(err)

	// ECDH-ES missing epk
	ecPriv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	_, err = decryptKey(encryptedKey, ECDHES, ecPriv, map[string]interface{}{}, A128GCM)
	s.Error(err)
	s.Contains(err.Error(), "missing epk")

	// ECDH-ES invalid epk type
	_, err = decryptKey(encryptedKey, ECDHES, ecPriv, map[string]interface{}{"epk": "not-a-map"}, A128GCM)
	s.Error(err)

	// ECDH-ES invalid epk content
	epkInvalid := map[string]interface{}{"epk": map[string]interface{}{"x": "invalid"}}
	_, err = decryptKey(encryptedKey, ECDHES, ecPriv, epkInvalid, A128GCM)
	s.Error(err)

	// ECDH-ES non-ECDSA private key
	epkEmpty := map[string]interface{}{"epk": map[string]interface{}{}}
	_, err = decryptKey(encryptedKey, ECDHES, privateKey, epkEmpty, A128GCM)
	s.Error(err)

	// ECDH-ES+KW missing epk
	_, err = decryptKey(encryptedKey, ECDHESA128KW, ecPriv, map[string]interface{}{}, A128GCM)
	s.Error(err)
}

func (s *JEWUtilsTestSuite) TestEncryptDecryptContent() {
	payload := []byte("Hello, JWE!")
	aad := []byte("additional-authenticated-data")

	testCases := []struct {
		name    string
		enc     ContentEncAlgorithm
		cekSize int
	}{
		{"A128GCM", A128GCM, 16},
		{"A192GCM", A192GCM, 24},
		{"A256GCM", A256GCM, 32},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cek := make([]byte, tc.cekSize)
			_, _ = rand.Read(cek)

			iv, ciphertext, tag, err := encryptContent(payload, cek, tc.enc, aad)
			s.NoError(err)

			decryptedPayload, err := decryptContent(ciphertext, iv, tag, cek, tc.enc, aad)
			s.NoError(err)
			s.Equal(payload, decryptedPayload)
		})
	}
}

func (s *JEWUtilsTestSuite) TestContentEncryption_Errors() {
	payload := []byte("payload")
	cek16 := make([]byte, 16)
	cekInvalid := []byte("too-short")

	// Encrypt errors
	_, _, _, err := encryptContent(payload, cekInvalid, A128GCM, nil)
	s.Error(err)
	_, _, _, err = encryptContent(payload, cek16, "INVALID", nil)
	s.Error(err)

	// Decrypt errors
	_, err = decryptContent(payload, cek16, cek16, cekInvalid, A128GCM, nil)
	s.Error(err)
	_, err = decryptContent(payload, cek16, cek16, cek16, "INVALID", nil)
	s.Error(err)
}

func (s *JEWUtilsTestSuite) TestDecodeJWE() {
	header := `{"alg":"RSA-OAEP-256","enc":"A128GCM"}`
	headerBase64 := base64.RawURLEncoding.EncodeToString([]byte(header))
	encryptedKey := base64.RawURLEncoding.EncodeToString([]byte("encrypted-key"))
	iv := base64.RawURLEncoding.EncodeToString([]byte("iv-iv-iv-iv-"))
	ciphertext := base64.RawURLEncoding.EncodeToString([]byte("ciphertext"))
	tag := base64.RawURLEncoding.EncodeToString([]byte("tag-tag-tag-tag-"))

	jweToken := fmt.Sprintf("%s.%s.%s.%s.%s", headerBase64, encryptedKey, iv, ciphertext, tag)

	decodedHeader, _, decodedEncryptedKey, _, _, _, err := DecodeJWE(jweToken)
	s.NoError(err)
	s.Equal("RSA-OAEP-256", decodedHeader["alg"])
	s.Equal([]byte("encrypted-key"), decodedEncryptedKey)
}

func (s *JEWUtilsTestSuite) TestDecodeJWE_Errors() {
	// Wrong number of parts
	_, _, _, _, _, _, err := DecodeJWE("a.b.c.d")
	s.Error(err)

	// Invalid base64
	_, _, _, _, _, _, err = DecodeJWE("@@.b.c.d.e")
	s.Error(err)
	_, _, _, _, _, _, err = DecodeJWE("YQ.@@.c.d.e")
	s.Error(err)
	_, _, _, _, _, _, err = DecodeJWE("YQ.YQ.@@.d.e")
	s.Error(err)
	_, _, _, _, _, _, err = DecodeJWE("YQ.YQ.YQ.@@.e")
	s.Error(err)
	_, _, _, _, _, _, err = DecodeJWE("YQ.YQ.YQ.YQ.@@")
	s.Error(err)

	// Invalid JSON header
	_, _, _, _, _, _, err = DecodeJWE(base64.RawURLEncoding.EncodeToString([]byte("{invalid}")) + ".b.c.d.e")
	s.Error(err)
}

func (s *JEWUtilsTestSuite) TestAESKeyWrapUnwrap() {
	kek := []byte("this-is-a-16-bay")         // 16 bytes
	cek := []byte("this-is-a-24-byte-key!!!") // 24 bytes

	wrapped, err := aesKeyWrap(kek, cek)
	s.NoError(err)

	unwrapped, err := aesKeyUnwrap(kek, wrapped)
	s.NoError(err)
	s.Equal(cek, unwrapped)
}

func (s *JEWUtilsTestSuite) TestAESKeyWrap_Errors() {
	kek16 := make([]byte, 16)
	cek7 := make([]byte, 7)
	_, err := aesKeyWrap(kek16, cek7)
	s.Error(err)

	_, err = aesKeyWrap([]byte("invalid-kek"), make([]byte, 16))
	s.Error(err)
}

func (s *JEWUtilsTestSuite) TestAESKeyUnwrap_Errors() {
	kek16 := make([]byte, 16)

	_, err := aesKeyUnwrap(kek16, make([]byte, 7))
	s.Error(err)

	_, err = aesKeyUnwrap(kek16, make([]byte, 8))
	s.Error(err)

	_, err = aesKeyUnwrap([]byte("invalid-kek"), make([]byte, 16))
	s.Error(err)

	// IV mismatch
	wrapped := make([]byte, 16)
	_, err = aesKeyUnwrap(kek16, wrapped)
	s.Error(err)
	s.Contains(err.Error(), "IV mismatch")
}

func (s *JEWUtilsTestSuite) TestComputeSharedSecret_Errors() {
	// Non-ECDH private key
	_, err := computeSharedSecret("not-a-key", nil)
	s.Error(err)

	// Unsupported public key
	priv, _ := ecdh.P256().GenerateKey(rand.Reader)
	_, err = computeSharedSecret(priv, "not-a-key")
	s.Error(err)
}

func (s *JEWUtilsTestSuite) TestGenerateEphemeralKey_Errors() {
	// Non-ECDSA key
	_, _, err := generateEphemeralKey("not-a-key")
	s.Error(err)
}

func (s *JEWUtilsTestSuite) TestIsSupportedCurve() {
	tests := []struct {
		name     string
		curve    elliptic.Curve
		expected bool
	}{
		{"P-256", elliptic.P256(), true},
		{"P-384", elliptic.P384(), true},
		{"P-521", elliptic.P521(), true},
		{"P-224", elliptic.P224(), false}, // Not supported
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := isSupportedCurve(tt.curve)
			s.Equal(tt.expected, result)
		})
	}
}

func (s *JEWUtilsTestSuite) TestInvalidAlgorithmCombinations() {
	// Test what happens when invalid algorithm combinations are used
	// during actual encryption/decryption since validation was removed

	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	s.NoError(err)

	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	s.NoError(err)

	cek := make([]byte, 32)
	_, _ = rand.Read(cek)

	s.Run("UnsupportedKeyEncryptionAlgorithm", func() {
		// Test with completely invalid algorithm
		_, _, err := encryptKey(cek, KeyEncAlgorithm("INVALID-ALG"), &rsaKey.PublicKey, A128GCM)
		s.Error(err)
		s.Contains(err.Error(), "unsupported JWE algorithm")
	})

	s.Run("UnsupportedContentEncryptionAlgorithm", func() {
		// Test with invalid content encryption algorithm
		_, _, err := encryptKey(cek, RSAOAEP256, &rsaKey.PublicKey, ContentEncAlgorithm("INVALID-ENC"))
		// This should succeed in encryptKey but fail later in content encryption
		s.NoError(err)

		// Test content encryption with invalid algorithm
		_, _, _, err = encryptContent([]byte("test"), cek, ContentEncAlgorithm("INVALID-ENC"), nil)
		s.Error(err)
		s.Contains(err.Error(), "unsupported encryption algorithm")
	})

	s.Run("WrongKeyTypeForAlgorithm", func() {
		// Test RSA algorithm with EC key - should fail during encryption
		_, _, err := encryptKey(cek, RSAOAEP256, &ecKey.PublicKey, A128GCM)
		s.Error(err)
		s.Contains(err.Error(), "unsupported public key type")

		// Test ECDH algorithm with RSA key - should fail during ephemeral key generation
		_, _, err = encryptKey(cek, ECDHES, &rsaKey.PublicKey, A128GCM)
		s.Error(err)
		s.Contains(err.Error(), "not an ECDSA key")
	})
}

func (s *JEWUtilsTestSuite) TestInvalidDecryptionScenarios() {
	// Test what happens during decryption with invalid combinations

	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	s.NoError(err)

	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	s.NoError(err)

	s.Run("UnsupportedAlgorithmInDecryption", func() {
		// Test decryption with unsupported algorithm
		_, err := decryptKey([]byte("fake-encrypted-key"), KeyEncAlgorithm("INVALID-ALG"), rsaKey, nil, A128GCM)
		s.Error(err)
		s.Contains(err.Error(), "unsupported JWE algorithm")
	})

	s.Run("WrongKeyTypeForDecryption", func() {
		// Test RSA decryption with EC private key
		_, err := decryptKey([]byte("fake-encrypted-key"), RSAOAEP256, ecKey, nil, A128GCM)
		s.Error(err)
		s.Contains(err.Error(), "unsupported private key type")

		// Test ECDH decryption with RSA private key
		header := map[string]interface{}{
			"epk": map[string]interface{}{
				"kty": "EC",
				"crv": "P-256",
				"x":   "WKn-ZIGevcwGIyyrzFoZNBdaq9_TsqzGHwHitJBcBmQ",
				"y":   "y77As5vbZdIgh9BzxPztXDBhKwuDiAv6rU9xDPVv3rI",
			},
		}
		_, err = decryptKey([]byte("fake-encrypted-key"), ECDHES, rsaKey, header, A128GCM)
		s.Error(err)
		// The actual error message may vary, just check that it fails
		s.NotEmpty(err.Error())
	})

	s.Run("InvalidContentDecryption", func() {
		// Test content decryption with invalid algorithm
		cek := make([]byte, 16)
		iv := make([]byte, 12)
		tag := make([]byte, 16)
		ciphertext := []byte("fake-ciphertext")

		_, err := decryptContent(ciphertext, iv, tag, cek, ContentEncAlgorithm("INVALID-ENC"), nil)
		s.Error(err)
		s.Contains(err.Error(), "unsupported encryption algorithm")
	})
}

func (s *JEWUtilsTestSuite) TestComputeSharedSecretForRecipient() {
	// Generate ECDSA keys for testing
	recipientPrivKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	s.NoError(err)

	ephemeralPrivKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	s.NoError(err)

	// Test successful shared secret computation
	sharedSecret, err := computeSharedSecretForRecipient(recipientPrivKey, &ephemeralPrivKey.PublicKey)
	s.NoError(err)
	s.NotEmpty(sharedSecret)

	// Test with non-ECDSA private key
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	s.NoError(err)
	_, err = computeSharedSecretForRecipient(rsaKey, &ephemeralPrivKey.PublicKey)
	s.Error(err)
	s.Contains(err.Error(), "private key is not an ECDSA key")
}

func (s *JEWUtilsTestSuite) TestEpkToMapEdgeCases() {
	// Test with P-384 curve
	privKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	s.NoError(err)

	// Convert ECDSA public key to ECDH public key
	ecdhPub, err := privKey.PublicKey.ECDH()
	s.NoError(err)

	epkMap := epkToMap(ecdhPub)
	s.NotNil(epkMap)
	s.Equal("P-384", epkMap["crv"])

	// Test with P-521 curve
	privKey521, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	s.NoError(err)

	ecdhPub521, err := privKey521.PublicKey.ECDH()
	s.NoError(err)

	epkMap521 := epkToMap(ecdhPub521)
	s.NotNil(epkMap521)
	s.Equal("P-521", epkMap521["crv"])

	// Test with non-ECDH key
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	s.NoError(err)

	epkMap = epkToMap(&rsaKey.PublicKey)
	s.Nil(epkMap)
}

func (s *JEWUtilsTestSuite) TestGenerateEphemeralKeyDifferentCurves() {
	curves := []struct {
		name  string
		curve elliptic.Curve
	}{
		{"P-256", elliptic.P256()},
		{"P-384", elliptic.P384()},
		{"P-521", elliptic.P521()},
	}

	for _, tc := range curves {
		s.Run(tc.name, func() {
			recipientKey, err := ecdsa.GenerateKey(tc.curve, rand.Reader)
			s.NoError(err)

			ephemeralPrivKey, ephemeralPubKey, err := generateEphemeralKey(&recipientKey.PublicKey)
			s.NoError(err)
			s.NotNil(ephemeralPrivKey)
			s.NotNil(ephemeralPubKey)

			// Generate the EPK map for verification
			epkMap := epkToMap(ephemeralPubKey)
			s.NotNil(epkMap)

			// Verify the curve matches
			expectedCrv := ""
			switch tc.curve {
			case elliptic.P256():
				expectedCrv = "P-256"
			case elliptic.P384():
				expectedCrv = "P-384"
			case elliptic.P521():
				expectedCrv = "P-521"
			}
			s.Equal(expectedCrv, epkMap["crv"])
		})
	}
}

func (s *JEWUtilsTestSuite) TestDecodeJWEWithDifferentHeaders() {
	// Test with ECDH-ES header containing epk
	epkHeader := map[string]interface{}{
		"alg": "ECDH-ES",
		"enc": "A256GCM",
		"epk": map[string]interface{}{
			"kty": "EC",
			"crv": "P-256",
			"x":   "WKn-ZIGevcwGIyyrzFoZNBdaq9_TsqzGHwHitJBcBmQ",
			"y":   "y77As5vbZdIgh9BzxPztXDBhKwuDiAv6rU9xDPVv3rI",
		},
	}

	headerJSON, _ := json.Marshal(epkHeader)
	headerBase64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	encryptedKey := base64.RawURLEncoding.EncodeToString([]byte{})
	iv := base64.RawURLEncoding.EncodeToString([]byte("123456789012"))
	ciphertext := base64.RawURLEncoding.EncodeToString([]byte("encrypted-data"))
	tag := base64.RawURLEncoding.EncodeToString([]byte("auth-tag-here!!!"))

	jweToken := fmt.Sprintf("%s.%s.%s.%s.%s", headerBase64, encryptedKey, iv, ciphertext, tag)

	decodedHeader, headerExtras, _, _, _, _, err := DecodeJWE(jweToken)
	s.NoError(err)
	s.Equal("ECDH-ES", decodedHeader["alg"])
	s.NotNil(headerExtras)

	// Test header missing mandatory fields
	incompleteHeader := map[string]interface{}{
		"alg": "RSA-OAEP-256",
		// missing "enc"
	}
	headerJSON, _ = json.Marshal(incompleteHeader)
	headerBase64 = base64.RawURLEncoding.EncodeToString(headerJSON)
	incompleteJWE := fmt.Sprintf("%s.%s.%s.%s.%s", headerBase64, encryptedKey, iv, ciphertext, tag)

	// DecodeJWE should succeed even with missing fields - validation happens later
	_, _, _, _, _, _, err = DecodeJWE(incompleteJWE)
	s.NoError(err)
}

// isSupportedCurve checks if the elliptic curve is supported.
func isSupportedCurve(curve elliptic.Curve) bool {
	switch curve {
	case elliptic.P256(), elliptic.P384(), elliptic.P521():
		return true
	default:
		return false
	}
}
