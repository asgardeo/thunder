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
	Hash(data []byte, salt []byte) string
	Verify(data []byte, salt []byte, expectedHash string) bool
}

type SHA256HashProvider struct{}

type PBKDF2HashProvider struct {
	Iterations int
	KeyLen     int
}

func NewHashProvider() HashProviderInterface {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "HashProvider"))
	// Try to get hash configs from the application configuration
	config := config.GetThunderRuntime().Config.Hash.Algorithm // Use the correct config getter

	switch config {
	case "SHA256":
		logger.Debug("Using SHA256 hash provider as per configuration.")
		return newSHA256HashProvider()
	case "PBKDF2":
		logger.Debug("Using PBKDF2 hash provider as per configuration.")
		return newPBKDF2HashProvider()
	default:
		logger.Debug("Unsupported hash algorithm in configuration. Defaulting to PBKDF2.")
		return newPBKDF2HashProvider()
	}
}

func newSHA256HashProvider() HashProviderInterface {
	return &SHA256HashProvider{}
}

func (hp *SHA256HashProvider) Hash(data []byte, salt []byte) string {
	combined := append(data, salt...)
	hash := sha256.Sum256(combined)
	return hex.EncodeToString(hash[:])
}

func (hp *SHA256HashProvider) Verify(data []byte, salt []byte, expectedHash string) bool {
	hashedData := hp.Hash(data, salt)
	return hashedData == expectedHash
}

func newPBKDF2HashProvider() HashProviderInterface {
	return &PBKDF2HashProvider{
		Iterations: 10000,
		KeyLen:     32,
	}
}

func (hp *PBKDF2HashProvider) Hash(data []byte, salt []byte) string {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "PBKDF2HashProvider"))
	hash, err := pbkdf2.Key(sha256.New, string(data), salt, hp.Iterations, hp.KeyLen)
	if err != nil {
		logger.Error("Error hashing data with PBKDF2: %v", log.Error(err))
		return ""
	}
	return hex.EncodeToString(hash)
}

func (hp *PBKDF2HashProvider) Verify(data []byte, salt []byte, expectedHash string) bool {
	hashedData := hp.Hash(data, salt)
	return hashedData == expectedHash
}
