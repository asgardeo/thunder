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

// Package encrypt provides cryptographic functionality with algorithm agility.
package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/crypto/hash"
	"github.com/asgardeo/thunder/internal/system/log"
)

// EncryptionService provides cryptographic operations.
type EncryptionService struct {
	Key []byte
	Kid string
}

var (
	// instance is the singleton instance of EncryptionService
	instance *EncryptionService
	// once ensures the singleton is initialized only once
	once sync.Once
)

// GetEncryptionService creates and returns a singleton instance of the EncryptionService.
func GetEncryptionService() *EncryptionService {
	once.Do(func() {
		var err error
		instance, err = initEncryptionService()
		if err != nil {
			panic(fmt.Sprintf("failed to initialize EncryptionService: %v", err))
		}
	})
	return instance
}

// initEncryptionService initializes the EncryptionService from configuration sources.
func initEncryptionService() (*EncryptionService, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "EncryptionService"))
	// Try to get key from the application configuration
	config := config.GetThunderRuntime().Config.Crypto.Encrypt.Key

	// Check if crypto configuration exists
	if config != "" {
		key, err := hex.DecodeString(config)
		if err == nil {
			return NewEncryptionService(key), nil
		} else {
			logger.Error("failed to decode encryption key from config", log.Error(err))
			return nil, err
		}
	} else {
		return nil, errors.New("encryption key not found")
	}
}

// NewEncryptionService creates a new instance of EncryptionService with the provided key.
func NewEncryptionService(key []byte) *EncryptionService {
	return &EncryptionService{
		Key: key,
		Kid: getKeyID(key), // Generate a unique key ID
	}
}

// Encrypt encrypts the given plaintext and returns a JSON string
// containing the encrypted data.
func (cs *EncryptionService) Encrypt(plaintext []byte) (string, error) {
	// Create AES cipher
	block, err := aes.NewCipher(cs.Key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM mode: %w", err)
	}

	// Create a nonce
	nonce := make([]byte, aesgcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt and authenticate plaintext, prepend nonce
	ciphertext := aesgcm.Seal(nonce, nonce, plaintext, nil)

	// Create metadata structure
	encData := EncryptedData{
		Algorithm:  AESGCM,
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
		KeyID:      cs.Kid, // Unique identifier for the key
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(encData)
	if err != nil {
		return "", fmt.Errorf("failed to serialize encrypted data: %w", err)
	}

	return string(jsonData), nil
}

// Decrypt decrypts the given JSON string produced by Encrypt
// and returns the original plaintext.
func (cs *EncryptionService) Decrypt(encodedData string) ([]byte, error) {
	// Deserialize JSON
	var encData EncryptedData
	if err := json.Unmarshal([]byte(encodedData), &encData); err != nil {
		return nil, fmt.Errorf("invalid data format: %w", err)
	}

	// Verify algorithm
	if encData.Algorithm != AESGCM {
		return nil, fmt.Errorf("unsupported algorithm: %s", encData.Algorithm)
	}

	// Decode the payload
	ciphertext, err := base64.StdEncoding.DecodeString(encData.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("invalid payload encoding: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(cs.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM mode: %w", err)
	}

	// Verify ciphertext length
	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	// Extract nonce and decrypt
	nonce, encryptedData := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// EncryptString encrypts the given plaintext string and returns a
// JSON string containing the encrypted data.
func (cs *EncryptionService) EncryptString(plaintext string) (string, error) {
	return cs.Encrypt([]byte(plaintext))
}

// DecryptString decrypts the given JSON string produced by Encrypt
// and returns the original plaintext string.
func (cs *EncryptionService) DecryptString(ciphertext string) (string, error) {
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
