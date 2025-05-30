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

package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// CryptoAlgorithm defines the interface that all encryption algorithms must implement
type CryptoAlgorithm interface {
	Encrypt(key []byte, plaintext []byte) (string, error)
	Decrypt(key []byte, encodedData string) ([]byte, error)
	Name() Algorithm
	KeySize() int
}

// AESGCMAlgorithm implements the CryptoAlgorithm interface for AES-256-GCM
type AESGCMAlgorithm struct{}

// Encrypt encrypts plaintext using AES-256-GCM and returns a serialized EncryptedData
func (a *AESGCMAlgorithm) Encrypt(key []byte, plaintext []byte) (string, error) {
	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Create GCM mode
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Create a nonce
	nonce := make([]byte, aesgcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt and authenticate plaintext, prepend nonce
	ciphertext := aesgcm.Seal(nonce, nonce, plaintext, nil)

	// Create metadata structure
	encData := EncryptedData{
		Algorithm:  AES256GCM,
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(encData)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(jsonData), nil
}

// Decrypt decrypts the base64-encoded serialized EncryptedData
func (a *AESGCMAlgorithm) Decrypt(key []byte, encodedData string) ([]byte, error) {
	// Decode base64
	jsonData, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 encoding: %w", err)
	}

	// Deserialize JSON
	var encData EncryptedData
	if err := json.Unmarshal(jsonData, &encData); err != nil {
		return nil, fmt.Errorf("invalid data format: %w", err)
	}

	// Verify algorithm
	if encData.Algorithm != AES256GCM {
		return nil, fmt.Errorf("unsupported algorithm: %s", encData.Algorithm)
	}

	// Decode the payload
	ciphertext, err := base64.StdEncoding.DecodeString(encData.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("invalid payload encoding: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create GCM mode
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
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

// Name returns the algorithm name
func (a *AESGCMAlgorithm) Name() Algorithm {
	return AES256GCM
}

// KeySize returns the required key size in bytes
func (a *AESGCMAlgorithm) KeySize() int {
	return 32 // 256 bits
}
