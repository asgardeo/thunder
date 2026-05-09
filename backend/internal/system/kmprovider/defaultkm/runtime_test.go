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

package defaultkm

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/asgardeo/thunder/internal/system/cryptolab"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/i18n/core"
	"github.com/asgardeo/thunder/internal/system/kmprovider"
	"github.com/asgardeo/thunder/tests/mocks/crypto/pki/pkimock"
)

const testKeyID = "test-key-id"

func newTestSvcErr() *serviceerror.ServiceError {
	return &serviceerror.ServiceError{
		Code:  "TEST-001",
		Error: core.I18nMessage{DefaultValue: "key not found"},
	}
}

// TestEncrypt_RSAOAEP256_Success covers line 68: the success path for RSA-OAEP-256
// when the PKI service returns a valid RSA certificate.
func TestEncrypt_RSAOAEP256_Success(t *testing.T) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	cert := &x509.Certificate{PublicKey: &rsaKey.PublicKey}

	pkiMock := pkimock.NewPKIServiceInterfaceMock(t)
	pkiMock.EXPECT().
		GetX509Certificate(testKeyID).
		Return(cert, nil)

	svc := NewRuntimeCryptoService(pkiMock, nil)

	params := cryptolab.AlgorithmParams{
		Algorithm: cryptolab.AlgorithmRSAOAEP256,
		RSAOAEP256: cryptolab.RSAOAEP256Params{
			ContentEncryptionAlgorithm: "A256GCM",
		},
	}
	wrappedCEK, details, err := svc.Encrypt(
		context.Background(), kmprovider.KeyRef{KeyID: testKeyID}, params, nil,
	)
	require.NoError(t, err)
	assert.NotEmpty(t, wrappedCEK)
	assert.NotNil(t, details)
}

// TestEncrypt_RSAOAEP256_GetPublicKeyError covers the error return path inside the
// RSA-OAEP-256 case (lines 65-66) when the PKI service cannot find the key.
func TestEncrypt_RSAOAEP256_GetPublicKeyError(t *testing.T) {
	pkiMock := pkimock.NewPKIServiceInterfaceMock(t)
	pkiMock.EXPECT().
		GetX509Certificate(testKeyID).
		Return(nil, newTestSvcErr())

	svc := NewRuntimeCryptoService(pkiMock, nil)

	params := cryptolab.AlgorithmParams{
		Algorithm: cryptolab.AlgorithmRSAOAEP256,
		RSAOAEP256: cryptolab.RSAOAEP256Params{
			ContentEncryptionAlgorithm: "A256GCM",
		},
	}
	_, _, err := svc.Encrypt(context.Background(), kmprovider.KeyRef{KeyID: testKeyID}, params, []byte("data"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), testKeyID)
}

// TestEncrypt_ECDHES_Success covers the ECDH-ES algorithm case (lines 69-70, 74) on the
// happy path where the PKI service returns a valid EC certificate.
func TestEncrypt_ECDHES_Success(t *testing.T) {
	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	cert := &x509.Certificate{PublicKey: &ecKey.PublicKey}

	pkiMock := pkimock.NewPKIServiceInterfaceMock(t)
	pkiMock.EXPECT().
		GetX509Certificate(testKeyID).
		Return(cert, nil)

	svc := NewRuntimeCryptoService(pkiMock, nil)

	params := cryptolab.AlgorithmParams{
		Algorithm: cryptolab.AlgorithmECDHES,
		ECDHES: cryptolab.ECDHESParams{
			ContentEncryptionAlgorithm: "A128GCM",
		},
	}
	_, details, err := svc.Encrypt(context.Background(), kmprovider.KeyRef{KeyID: testKeyID}, params, nil)
	require.NoError(t, err)
	assert.NotNil(t, details)
}

// TestEncrypt_ECDHES_GetPublicKeyError covers the error return path inside the
// ECDH-ES case (lines 71-72) when the PKI service cannot find the key.
func TestEncrypt_ECDHES_GetPublicKeyError(t *testing.T) {
	pkiMock := pkimock.NewPKIServiceInterfaceMock(t)
	pkiMock.EXPECT().
		GetX509Certificate(testKeyID).
		Return(nil, newTestSvcErr())

	svc := NewRuntimeCryptoService(pkiMock, nil)

	params := cryptolab.AlgorithmParams{
		Algorithm: cryptolab.AlgorithmECDHES,
		ECDHES: cryptolab.ECDHESParams{
			ContentEncryptionAlgorithm: "A128GCM",
		},
	}
	_, _, err := svc.Encrypt(context.Background(), kmprovider.KeyRef{KeyID: testKeyID}, params, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), testKeyID)
}

// TestEncrypt_UnsupportedAlgorithm covers the default branch (lines 75-76) for an
// algorithm that is not handled by the switch.
func TestEncrypt_UnsupportedAlgorithm(t *testing.T) {
	svc := NewRuntimeCryptoService(nil, nil)

	params := cryptolab.AlgorithmParams{Algorithm: "UNSUPPORTED"}
	_, _, err := svc.Encrypt(context.Background(), kmprovider.KeyRef{KeyID: testKeyID}, params, []byte("data"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported algorithm")
}
