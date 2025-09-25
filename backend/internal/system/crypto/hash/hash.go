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
	"encoding/base64"
	"encoding/hex"

	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/log"
)

const (
	sha256Algorithm = "SHA256"
	pbkdf2Algorithm = "PBKDF2"

	defaultPBKDF2Iterations = 600000
	defaultPBKDF2KeyLength  = 32
)

// NewCredential creates a credential using the configured hash provider.
func NewCredential(credentialValue []byte) Credential {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "HashProvider"))
	algorithm := config.GetThunderRuntime().Config.Hash.Algorithm

	switch algorithm {
	case sha256Algorithm:
		logger.Debug("Using SHA256 hash provider as per configuration.")
		return newSHA256Credential(credentialValue)
	default:
		logger.Debug("Using PBKDF2 hash provider as per configuration.")
		return newPBKDF2Credential(credentialValue)
	}
}

// Verify verifies a credential by selecting the appropriate hash provider
// based on the reference credential's algorithm.
func Verify(credentialValueToVerify []byte, referenceCredential Credential) bool {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "HashProvider"))
	algorithm := referenceCredential.Algorithm

	switch algorithm {
	case sha256Algorithm:
		logger.Debug("Using SHA256 hash provider for verification as per credential algorithm.")
		return verifySHA256Credential(credentialValueToVerify, referenceCredential)
	case pbkdf2Algorithm:
		logger.Debug("Using PBKDF2 hash provider for verification as per credential algorithm.")
		return verifyPBKDF2Credential(credentialValueToVerify, referenceCredential)
	default:
		logger.Error("Unsupported hash algorithm in credential", log.String("algorithm", referenceCredential.Algorithm))
		return false
	}
}

// newSHA256Credential generates a SHA256 hash of the input data combined with salt.
func newSHA256Credential(credentialValue []byte) Credential {
	credSalt, _ := generateSalt()
	credentialValue = append(credentialValue, credSalt...)
	hash := sha256.Sum256(credentialValue)

	return Credential{
		Algorithm: sha256Algorithm,
		Hash:      hex.EncodeToString(hash[:]),
		Salt:      hex.EncodeToString(credSalt),
	}
}

// verifySHA256Credential checks if the SHA256 hash of the input data and salt matches the expected hash.
func verifySHA256Credential(credentialValueToVerify []byte, referenceCredential Credential) bool {
	saltBytes, err := hex.DecodeString(referenceCredential.Salt)
	if err != nil {
		return false
	}
	credentialValueToVerify = append(credentialValueToVerify, saltBytes...)
	hashedData := sha256.Sum256(credentialValueToVerify)
	return referenceCredential.Hash == hex.EncodeToString(hashedData[:])
}

// newPBKDF2Credential generates a PBKDF2 hash of the input data using the provided salt.
func newPBKDF2Credential(credentialValue []byte) Credential {
	credSalt, _ := generateSalt()
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "PBKDF2HashProvider"))
	hash, err := pbkdf2.Key(sha256.New, string(credentialValue), credSalt, defaultPBKDF2Iterations, defaultPBKDF2KeyLength)
	if err != nil {
		logger.Error("Error hashing data with PBKDF2: %v", log.Error(err))
		return Credential{}
	}
	return Credential{
		Algorithm: pbkdf2Algorithm,
		Hash:      hex.EncodeToString(hash),
		Salt:      hex.EncodeToString(credSalt),
	}
}

// verifyPBKDF2Credential checks if the PBKDF2 hash of the input data and salt matches the expected hash.
func verifyPBKDF2Credential(credentialValueToVerify []byte, referenceCredential Credential) bool {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "PBKDF2HashProvider"))
	saltBytes, err := hex.DecodeString(referenceCredential.Salt)
	if err != nil {
		logger.Error("Error decoding salt: %v", log.Error(err))
		return false
	}
	hash, err := pbkdf2.Key(sha256.New,
		string(credentialValueToVerify), saltBytes, defaultPBKDF2Iterations, defaultPBKDF2KeyLength)
	if err != nil {
		logger.Error("Error hashing data with PBKDF2: %v", log.Error(err))
		return false
	}
	return hex.EncodeToString(hash) == referenceCredential.Hash
}

// generateSalt generates a random salt string.
func generateSalt() ([]byte, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}

// GenerateThumbprint generates a SHA-256 thumbprint for the given data.
func GenerateThumbprint(data []byte) string {
	hash := sha256.Sum256(data)
	return base64.StdEncoding.EncodeToString(hash[:])
}

// GenerateThumbprintFromString generates a SHA-256 thumbprint for the given string data.
func GenerateThumbprintFromString(data string) string {
	return GenerateThumbprint([]byte(data))
}
