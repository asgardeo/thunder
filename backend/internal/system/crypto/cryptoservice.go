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
	"encoding/base64"
	"errors"
	"io"
	"log"
	"sync"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/utils"
)

// Algorithm represents supported encryption algorithms
type Algorithm string

const (
	// AES256GCM represents AES-GCM algorithm
	AES256GCM Algorithm = "AES-GCM"
)

// CryptoService provides cryptographic operations
type CryptoService struct {
	key   []byte
	keyID string
	algo  CryptoAlgorithm
	mu    sync.RWMutex // For thread-safe key updates
}

var (
	// instance is the singleton instance of CryptoService
	instance *CryptoService
	// once ensures the singleton is initialized only once
	once sync.Once
)

// GetCryptoService creates and returns a singleton instance of the CryptoService.
func GetCryptoService() *CryptoService {
	once.Do(func() {
		var err error
		instance, err = initCryptoService()
		if err != nil {
			log.Fatalf("Failed to initialize CryptoService: %v", err)
		}
	})
	return instance
}

// initCryptoService initializes the CryptoService from configuration sources
func initCryptoService() (*CryptoService, error) {
	// Try to get key from the application configuration
	config := config.GetThunderRuntime().Config.Crypto.Key // Use the correct config getter

	// Check if crypto configuration exists
	if config != "" {
		key, err := base64.StdEncoding.DecodeString(config)
		if err == nil && len(key) == 32 {
			log.Println("Using crypto key from configuration")
			return NewCryptoService(key)
		}
		log.Println("Warning: Invalid crypto key in configuration, generating a new key")
	}

	// Generate new key as fallback for development
	log.Println("Warning: No valid crypto key found in configuration, generating a new one")

	key, err := GenerateRandomKey()
	if err != nil {
		return nil, err
	}

	// Print the generated key for development purposes
	encodedKey := base64.StdEncoding.EncodeToString(key)
	log.Printf("Generated new crypto key (base64): %s", encodedKey)

	return NewCryptoService(key)
}

// NewCryptoService creates a new instance of CryptoService with the provided key
func NewCryptoService(key []byte) (*CryptoService, error) {
	if len(key) != 32 {
		return nil, errors.New("key must be exactly 32 bytes for AES-256")
	}

	return &CryptoService{
		key:   key,
		keyID: utils.GenerateUUID(), // Unique identifier for the key
		algo:  &AESGCMAlgorithm{},   // Default to AES-256-GCM
	}, nil
}

// GenerateRandomKey generates a random 32-byte key suitable for AES-256
func GenerateRandomKey() ([]byte, error) {
	key := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// Encrypt encrypts plaintext using AES-256-GCM and returns a serialized EncryptedData
func (cs *CryptoService) Encrypt(plaintext []byte) (string, error) {
	cs.mu.RLock()
	key := cs.key
	cs.mu.RUnlock()

	return cs.algo.Encrypt(key, plaintext)
}

// Decrypt decrypts the base64-encoded serialized EncryptedData
func (cs *CryptoService) Decrypt(encodedData string) ([]byte, error) {
	cs.mu.RLock()
	key := cs.key
	cs.mu.RUnlock()

	return cs.algo.Decrypt(key, encodedData)
}

// EncryptString is a convenience method to encrypt string data
func (cs *CryptoService) EncryptString(plaintext string) (string, error) {
	return cs.Encrypt([]byte(plaintext))
}

// DecryptString is a convenience method to decrypt to string data
func (cs *CryptoService) DecryptString(ciphertext string) (string, error) {
	plaintext, err := cs.Decrypt(ciphertext)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
