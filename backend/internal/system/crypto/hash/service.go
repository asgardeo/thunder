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
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/log"
)

const (
	defaultSaltSize         = 16
	defaultPBKDF2Iterations = 600000
	defaultPBKDF2KeySize    = 32
)

var (
	logger = log.GetLogger().With(log.String(log.LoggerKeyComponentName, "HashService"))
)

// HashServiceInterface defines the interface for hashing services.
type HashServiceInterface interface {
	Generate(credentialValue []byte) (Credential, error)
	Verify(credentialValueToVerify []byte, referenceCredential Credential) (bool, error)
}

type sha256HashProvider struct {
	SaltSize int
}

type pbkdf2HashProvider struct {
	SaltSize   int
	Iterations int
	KeySize    int
}

// newHashService initializes and returns the appropriate hash provider based on configuration
func newHashService() HashServiceInterface {
	cfg := config.GetThunderRuntime().Config.Crypto.PasswordHashing
	algorithm := CredAlgorithm(cfg.Algorithm)
	params := cfg.Parameters

	switch algorithm {
	case SHA256:
		logger.Debug("Using SHA256 hash algorithm for password hashing")
		return newSHA256Provider(params.SaltSize)
	case PBKDF2:
		logger.Debug("Using PBKDF2 hash algorithm for password hashing")
		return newPBKDF2Provider(params.SaltSize, params.Iterations, params.KeySize)
	default:
		panic(fmt.Sprintf("unsupported hash algorithm configured: %s", algorithm))
	}
}

// newSHA256Provider creates a new SHA256HashProvider instance
func newSHA256Provider(saltSize int) *sha256HashProvider {
	if saltSize <= 0 {
		saltSize = defaultSaltSize
	}
	return &sha256HashProvider{
		SaltSize: saltSize,
	}
}

// Generate SHA256Credential generates a SHA256 hash
func (a *sha256HashProvider) Generate(credentialValue []byte) (Credential, error) {
	credSalt, err := generateSalt(a.SaltSize)
	if err != nil {
		return Credential{}, err
	}
	credentialWithSalt := append([]byte(nil), credentialValue...)
	credentialWithSalt = append(credentialWithSalt, credSalt...)
	hash := sha256.Sum256(credentialWithSalt)

	return Credential{
		Algorithm: SHA256,
		Hash:      hex.EncodeToString(hash[:]),
		Parameters: CredParameters{
			Salt: hex.EncodeToString(credSalt),
		},
	}, nil
}

// Verify SHA256Credential checks if the SHA256 hash of the input data and salt matches the expected hash.
func (a *sha256HashProvider) Verify(credentialValueToVerify []byte, referenceCredential Credential) (bool, error) {
	saltBytes, err := hex.DecodeString(referenceCredential.Parameters.Salt)
	if err != nil {
		return false, err
	}
	credentialWithSalt := append([]byte(nil), credentialValueToVerify...)
	credentialWithSalt = append(credentialWithSalt, saltBytes...)
	hashedData := sha256.Sum256(credentialWithSalt)
	return referenceCredential.Hash == hex.EncodeToString(hashedData[:]), nil
}

// newPBKDF2Provider creates a new PBKDF2HashProvider instance
func newPBKDF2Provider(saltSize, iterations, keySize int) *pbkdf2HashProvider {
	if saltSize <= 0 {
		saltSize = defaultSaltSize
	}
	if iterations <= 0 {
		iterations = defaultPBKDF2Iterations
	}
	if keySize <= 0 {
		keySize = defaultPBKDF2KeySize
	}
	return &pbkdf2HashProvider{
		SaltSize:   saltSize,
		Iterations: iterations,
		KeySize:    keySize,
	}
}

// Generate PBKDF2Credential generates a PBKDF2 hash of the input data using the provided salt.
func (a *pbkdf2HashProvider) Generate(credentialValue []byte) (Credential, error) {
	credSalt, err := generateSalt(a.SaltSize)
	if err != nil {
		return Credential{}, err
	}
	hash, err := pbkdf2.Key(sha256.New, string(credentialValue), credSalt, a.Iterations, a.KeySize)
	if err != nil {
		logger.Error("Error hashing data with PBKDF2: %v", log.Error(err))
		return Credential{}, err
	}
	return Credential{
		Algorithm: PBKDF2,
		Hash:      hex.EncodeToString(hash),
		Parameters: CredParameters{
			Iterations: a.Iterations,
			KeySize:    a.KeySize,
			Salt:       hex.EncodeToString(credSalt),
		},
	}, nil
}

// Verify PBKDF2Credential checks if the PBKDF2 hash of the input data and salt matches the expected hash.
func (a *pbkdf2HashProvider) Verify(credentialValueToVerify []byte, referenceCredential Credential) (bool, error) {
	iterations := referenceCredential.Parameters.Iterations
	if iterations <= 0 {
		iterations = defaultPBKDF2Iterations
	}
	keySize := referenceCredential.Parameters.KeySize
	if keySize <= 0 {
		keySize = defaultPBKDF2KeySize
	}
	saltBytes, err := hex.DecodeString(referenceCredential.Parameters.Salt)
	if err != nil {
		logger.Error("Error decoding salt: %v", log.Error(err))
		return false, err
	}
	hash, err := pbkdf2.Key(sha256.New,
		string(credentialValueToVerify), saltBytes, iterations, keySize)
	if err != nil {
		logger.Error("Error hashing data with PBKDF2: %v", log.Error(err))
		return false, err
	}
	return hex.EncodeToString(hash) == referenceCredential.Hash, nil
}

// generateSalt generates a random salt string.
func generateSalt(saltSize int) ([]byte, error) {
	salt := make([]byte, saltSize)
	_, err := rand.Read(salt)
	if err != nil {
		logger.Error("Error generating salt: %v", log.Error(err))
		return nil, err
	}
	return salt, nil
}
