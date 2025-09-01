/*
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
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

// Package hash provides generic hashing utilities for sensitive data.
package hash

import (
	"crypto/pbkdf2"
	"crypto/sha256"
	"encoding/hex"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/log"
)

// HashProviderInterface defines the interface for hashing and verifying data.
type HashProviderInterface interface {
	GetAlgorithm() string
	Hash(data []byte, salt []byte) string
	Verify(data []byte, salt []byte, expectedHash string) bool
}

// SHA256HashProvider implements the HashProviderInterface for SHA256 hashing.
type SHA256HashProvider struct{}

// PBKDF2HashProvider implements the HashProviderInterface for PBKDF2 hashing.
type PBKDF2HashProvider struct {
	Iterations int
	KeyLen     int
}

// NewHashProvider initializes and returns a HashProviderInterface based on configuration.
func NewHashProvider() HashProviderInterface {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "HashProvider"))
	// Try to get hash configs from the application configuration
	config := config.GetThunderRuntime().Config.Hash.Algorithm // Use the correct config getter

	switch config {
	case "SHA256":
		logger.Debug("Using SHA256 hash provider as per configuration.")
		return newSHA256HashProvider()
	default:
		logger.Debug("Using PBKDF2 hash provider as per configuration.")
		return newPBKDF2HashProvider(600000, 32)
	}
}

func newSHA256HashProvider() HashProviderInterface {
	return &SHA256HashProvider{}
}

// GetAlgorithm returns the name of the hashing algorithm.
func (hp *SHA256HashProvider) GetAlgorithm() string {
	return "SHA256"
}

// Hash generates a SHA256 hash of the input data combined with the salt.
func (hp *SHA256HashProvider) Hash(data []byte, salt []byte) string {
	data = append(data, salt...)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// Verify checks if the SHA256 hash of the input data and salt matches the expected hash.
func (hp *SHA256HashProvider) Verify(data []byte, salt []byte, expectedHash string) bool {
	hashedData := hp.Hash(data, salt)
	return hashedData == expectedHash
}

func newPBKDF2HashProvider(iterations int, keyLen int) HashProviderInterface {
	return &PBKDF2HashProvider{
		Iterations: iterations,
		KeyLen:     keyLen,
	}
}

// GetAlgorithm returns the name of the hashing algorithm.
func (hp *PBKDF2HashProvider) GetAlgorithm() string {
	return "PBKDF2"
}

// Hash generates a PBKDF2 hash of the input data using the provided salt.
func (hp *PBKDF2HashProvider) Hash(data []byte, salt []byte) string {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "PBKDF2HashProvider"))
	hash, err := pbkdf2.Key(sha256.New, string(data), salt, hp.Iterations, hp.KeyLen)
	if err != nil {
		logger.Error("Error hashing data with PBKDF2: %v", log.Error(err))
		return ""
	}
	return hex.EncodeToString(hash)
}

// Verify checks if the PBKDF2 hash of the input data and salt matches the expected hash.
func (hp *PBKDF2HashProvider) Verify(data []byte, salt []byte, expectedHash string) bool {
	hashedData := hp.Hash(data, salt)
	return hashedData == expectedHash
}
