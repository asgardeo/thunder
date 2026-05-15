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
	cek := make([]byte, 32)
	_, _ = rand.Read(cek)

	s.Run("UnsupportedContentEncryptionAlgorithm", func() {
		_, _, _, err := encryptContent([]byte("test"), cek, ContentEncAlgorithm("INVALID-ENC"), nil)
		s.Error(err)
		s.Contains(err.Error(), "unsupported encryption algorithm")
	})
}

func (s *JEWUtilsTestSuite) TestInvalidDecryptionScenarios() {
	s.Run("InvalidContentDecryption", func() {
		cek := make([]byte, 16)
		iv := make([]byte, 12)
		tag := make([]byte, 16)
		ciphertext := []byte("fake-ciphertext")

		_, err := decryptContent(ciphertext, iv, tag, cek, ContentEncAlgorithm("INVALID-ENC"), nil)
		s.Error(err)
		s.Contains(err.Error(), "unsupported encryption algorithm")
	})
}

func (s *JEWUtilsTestSuite) TestEpkToMapEdgeCases() {
	// Test with P-384 curve
	privKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	s.NoError(err)

	// Convert ECDSA public key to ECDH public key
	ecdhPub, err := privKey.PublicKey.ECDH()
	s.NoError(err)

	epkMap, err := epkToMap(ecdhPub)
	s.NoError(err)
	s.NotNil(epkMap)
	s.Equal("P-384", epkMap["crv"])

	// Test with P-521 curve
	privKey521, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	s.NoError(err)

	ecdhPub521, err := privKey521.PublicKey.ECDH()
	s.NoError(err)

	epkMap521, err := epkToMap(ecdhPub521)
	s.NoError(err)
	s.NotNil(epkMap521)
	s.Equal("P-521", epkMap521["crv"])

	// Test with non-ECDH key
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	s.NoError(err)

	epkMap, err = epkToMap(&rsaKey.PublicKey)
	s.Error(err)
	s.Nil(epkMap)
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
