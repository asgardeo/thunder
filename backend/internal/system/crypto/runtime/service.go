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

// Package runtime provides the RuntimeCryptoProvider implementation backed by PKI key material.
package runtime

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"errors"
	"fmt"
	"sync"

	"github.com/asgardeo/thunder/internal/system/crypto"
	"github.com/asgardeo/thunder/internal/system/crypto/config"
	"github.com/asgardeo/thunder/internal/system/crypto/pki"
	"github.com/asgardeo/thunder/internal/system/log"
)

// runtimeCryptoService implements RuntimeCryptoProvider backed by PKI key material.
type runtimeCryptoService struct {
	pkiService pki.PKIServiceInterface
	logger     *log.Logger
}

var (
	runtimeInstance *runtimeCryptoService
	runtimeOnce     sync.Once
)

// GetRuntimeCryptoService returns the singleton RuntimeCryptoProvider instance.
func GetRuntimeCryptoService() crypto.RuntimeCryptoProvider {
	runtimeOnce.Do(func() {
		logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "RuntimeCryptoService"))
		pkiSvc, err := pki.Initialize()
		if err != nil {
			logger.Warn("PKI service unavailable; RSA operations will fail", log.String("reason", err.Error()))
		}
		runtimeInstance = &runtimeCryptoService{
			pkiService: pkiSvc,
			logger:     logger,
		}
	})
	return runtimeInstance
}

// Encrypt encrypts content using the algorithm specified. AlgorithmAESGCM delegates to the config
// encryption service; AlgorithmRSAOAEP256 uses the PKI public key identified by keyRef.
func (s *runtimeCryptoService) Encrypt(
	ctx context.Context, keyRef crypto.KeyRef, algorithm crypto.Algorithm, content []byte,
) ([]byte, error) {
	switch algorithm {
	case crypto.AlgorithmAESGCM:
		return config.GetEncryptionService().Encrypt(ctx, content)
	case crypto.AlgorithmRSAOAEP256:
		if s.pkiService == nil {
			return nil, errors.New("PKI service not initialized")
		}
		cert, svcErr := s.pkiService.GetX509Certificate(keyRef.KeyID)
		if svcErr != nil {
			return nil, fmt.Errorf("key not found for id %s: [%s] %s",
				keyRef.KeyID, svcErr.Code, svcErr.Error.DefaultValue)
		}
		rsaPub, ok := cert.PublicKey.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("key is not an RSA public key")
		}
		return rsa.EncryptOAEP(sha256.New(), rand.Reader, rsaPub, content, nil)
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
}

// Decrypt decrypts content using the algorithm specified. AlgorithmAESGCM delegates to the config
// encryption service; AlgorithmRSAOAEP256 uses the PKI private key identified by keyRef.
func (s *runtimeCryptoService) Decrypt(
	ctx context.Context, keyRef crypto.KeyRef, algorithm crypto.Algorithm, content []byte,
) ([]byte, error) {
	switch algorithm {
	case crypto.AlgorithmAESGCM:
		return config.GetEncryptionService().Decrypt(ctx, content)
	case crypto.AlgorithmRSAOAEP256:
		if s.pkiService == nil {
			return nil, errors.New("PKI service not initialized")
		}
		privKey, svcErr := s.pkiService.GetPrivateKey(keyRef.KeyID)
		if svcErr != nil {
			return nil, fmt.Errorf("key not found for id %s: [%s] %s",
				keyRef.KeyID, svcErr.Code, svcErr.Error.DefaultValue)
		}
		rsaPriv, ok := privKey.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("key is not an RSA private key")
		}
		// #nosec G776 -- rand.Reader is used as the random source; label parameter is nil per RFC 8017
		return rsa.DecryptOAEP(sha256.New(), rand.Reader, rsaPriv, content, nil)
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
}

// Sign is not yet implemented.
func (s *runtimeCryptoService) Sign(_ context.Context, _ crypto.KeyRef, _ crypto.Algorithm, _ []byte) ([]byte, error) {
	return nil, errors.New("not implemented")
}

// GetPublicKeys is not yet implemented.
func (s *runtimeCryptoService) GetPublicKeys(
	_ context.Context, _ crypto.PublicKeyFilter,
) ([]crypto.PublicKeyInfo, error) {
	return nil, errors.New("not implemented")
}

// GetTLSMaterial is not yet implemented.
func (s *runtimeCryptoService) GetTLSMaterial(_ context.Context, _ crypto.KeyRef) (*crypto.TLSMaterial, error) {
	return nil, errors.New("not implemented")
}
