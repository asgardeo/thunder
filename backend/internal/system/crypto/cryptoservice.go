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

// Package crypto provides cryptographic functionality with algorithm agility.
package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"sync"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/crypto/hash"
	"github.com/asgardeo/thunder/internal/system/log"
)

// Algorithm represents supported encryption algorithms.
type Algorithm string

const (
	// AESGCM represents AES-GCM algorithm
	AESGCM Algorithm = "AES-GCM"
	// DefaultKeySize defines the default key size for AES-GCM
	DefaultKeySize = 32
)

// CryptoService provides cryptographic operations.
type CryptoService struct {
	Key []byte
	Kid string
	Alg CryptoAlgorithm
	Mu  sync.RWMutex // For thread-safe key updates
}

var (
	// instance is the singleton instance of CryptoService
	instance *CryptoService
	// once ensures the singleton is initialized only once
	once sync.Once
)

// GetCryptoService creates and returns a singleton instance of the CryptoService.
func GetCryptoService() *CryptoService {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "CryptoService"))
	once.Do(func() {
		var err error
		instance, err = initCryptoService()
		if err != nil {
			logger.Error("Failed to initialize CryptoService: %v", log.Error(err))
		}
	})
	return instance
}

// initCryptoService initializes the CryptoService from configuration sources.
func initCryptoService() (*CryptoService, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "CryptoService"))
	// Try to get key from the application configuration
	config := config.GetThunderRuntime().Config.Crypto.Key // Use the correct config getter

	// Check if crypto configuration exists
	if config != "" {
		key, err := hex.DecodeString(config)
		if err == nil {
			logger.Debug("Using crypto key from configuration")
			return NewCryptoService(key)
		}
		logger.Warn("Invalid crypto key in configuration, generating a new key")
	}

	// Generate new key as fallback for development
	logger.Warn("No valid crypto key found in configuration, generating a new one")

	key, err := GenerateRandomKey(DefaultKeySize)
	if err != nil {
		return nil, err
	}

	// Print the generated key for development purposes
	encodedKey := hex.EncodeToString(key)
	logger.Debug("Generated new crypto key (hex): %s", log.String("logKey", encodedKey))

	return NewCryptoService(key)
}

// NewCryptoService creates a new instance of CryptoService with the provided key.
func NewCryptoService(key []byte) (*CryptoService, error) {
	// Check key size for algorithm

	return &CryptoService{
		Key: key,
		Kid: getKeyID(key),      // Generate a unique key ID
		Alg: &AESGCMAlgorithm{}, // Default to AES-256-GCM
	}, nil
}

// GenerateRandomKey generates a random key of the specified size.
func GenerateRandomKey(keySize int) ([]byte, error) {
	key := make([]byte, keySize)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// Encrypt encrypts the given plaintext and returns a JSON string
// containing the encrypted data.
func (cs *CryptoService) Encrypt(plaintext []byte) (string, error) {
	cs.Mu.RLock()
	key := cs.Key
	cs.Mu.RUnlock()

	return cs.Alg.Encrypt(key, cs.Kid, plaintext)
}

// Decrypt decrypts the given JSON string produced by Encrypt
// and returns the original plaintext.
func (cs *CryptoService) Decrypt(encodedData string) ([]byte, error) {
	cs.Mu.RLock()
	key := cs.Key
	cs.Mu.RUnlock()

	return cs.Alg.Decrypt(key, encodedData)
}

// EncryptString encrypts the given plaintext string and returns a
// JSON string containing the encrypted data.
func (cs *CryptoService) EncryptString(plaintext string) (string, error) {
	return cs.Encrypt([]byte(plaintext))
}

// DecryptString decrypts the given JSON string produced by Encrypt
// and returns the original plaintext string.
func (cs *CryptoService) DecryptString(ciphertext string) (string, error) {
	plaintext, err := cs.Decrypt(ciphertext)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// getKeyID generates a unique identifier for the key.
func getKeyID(key []byte) string {
	return hash.GenerateThumbprint(key)
}
