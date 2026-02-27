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

	"golang.org/x/crypto/argon2"
)

const (
	maxUint8  = int(^uint8(0))
	maxUint32 = int(^uint32(0))
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

type argon2idHashProvider struct {
	SaltSize    int
	Memory      int
	Iterations  int
	Parallelism int
	KeySize     int
}

// newHashService initializes and returns the appropriate hash provider based on configuration
func newHashService() (HashServiceInterface, error) {
	cfg := config.GetThunderRuntime().Config.Crypto.PasswordHashing
	algorithm := CredAlgorithm(cfg.Algorithm)

	switch algorithm {
	case SHA256:
		logger.Debug("Using SHA256 hash algorithm for password hashing")
		if err := validateSHA256Config(cfg.SHA256); err != nil {
			return nil, err
		}
		return newSHA256Provider(cfg.SHA256.SaltSize), nil
	case PBKDF2:
		logger.Debug("Using PBKDF2 hash algorithm for password hashing")
		if err := validatePBKDF2Config(cfg.PBKDF2); err != nil {
			return nil, err
		}
		return newPBKDF2Provider(cfg.PBKDF2.SaltSize, cfg.PBKDF2.Iterations, cfg.PBKDF2.KeySize), nil
	case ARGON2ID:
		logger.Debug("Using Argon2id hash algorithm for password hashing")
		if err := validateArgon2idConfig(cfg.Argon2ID); err != nil {
			return nil, err
		}
		return newArgon2idProvider(
			cfg.Argon2ID.SaltSize,
			cfg.Argon2ID.Memory,
			cfg.Argon2ID.Iterations,
			cfg.Argon2ID.Parallelism,
			cfg.Argon2ID.KeySize,
		), nil
	default:
		return nil, fmt.Errorf("unsupported hash algorithm configured: %s", algorithm)
	}
}

// newSHA256Provider creates a new SHA256HashProvider instance
func newSHA256Provider(saltSize int) *sha256HashProvider {
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
	if err := validateCredentialAlgorithm(referenceCredential, SHA256); err != nil {
		return false, err
	}
	saltBytes, err := decodeSalt(referenceCredential.Parameters.Salt)
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
	if err := validateCredentialAlgorithm(referenceCredential, PBKDF2); err != nil {
		return false, err
	}
	iterations, err := requirePositiveInt(referenceCredential.Parameters.Iterations, "iterations")
	if err != nil {
		logger.Error("Invalid PBKDF2 iterations: %v", log.Error(err))
		return false, err
	}
	keySize, err := requirePositiveInt(referenceCredential.Parameters.KeySize, "key size")
	if err != nil {
		logger.Error("Invalid PBKDF2 key size: %v", log.Error(err))
		return false, err
	}
	saltBytes, err := decodeSalt(referenceCredential.Parameters.Salt)
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

func newArgon2idProvider(saltSize, memory, iterations, parallelism, keySize int) *argon2idHashProvider {
	return &argon2idHashProvider{
		SaltSize:    saltSize,
		Memory:      memory,
		Iterations:  iterations,
		Parallelism: parallelism,
		KeySize:     keySize,
	}
}

// Generate Argon2idCredential generates an Argon2id hash.
func (a *argon2idHashProvider) Generate(credentialValue []byte) (Credential, error) {
	credSalt, err := generateSalt(a.SaltSize)
	if err != nil {
		return Credential{}, err
	}

	//nolint:gosec // G115 - Conversion is safe
	hash := argon2.IDKey(
		credentialValue,
		credSalt,
		uint32(a.Iterations),
		uint32(a.Memory),
		uint8(a.Parallelism),
		uint32(a.KeySize),
	)

	return Credential{
		Algorithm: ARGON2ID,
		Hash:      hex.EncodeToString(hash),
		Parameters: CredParameters{
			Memory:      a.Memory,
			Iterations:  a.Iterations,
			Parallelism: a.Parallelism,
			KeySize:     a.KeySize,
			Salt:        hex.EncodeToString(credSalt),
		},
	}, nil
}

// Verify Argon2idCredential checks if the Argon2id hash of the input data and salt matches the expected hash.
func (a *argon2idHashProvider) Verify(credentialValueToVerify []byte, referenceCredential Credential) (bool, error) {
	if err := validateCredentialAlgorithm(referenceCredential, ARGON2ID); err != nil {
		return false, err
	}
	memory, err := requirePositiveIntWithMax(referenceCredential.Parameters.Memory, maxUint32, "memory")
	if err != nil {
		logger.Error("Invalid Argon2id memory: %v", log.Error(err))
		return false, err
	}
	iterations, err := requirePositiveIntWithMax(referenceCredential.Parameters.Iterations, maxUint32, "iterations")
	if err != nil {
		logger.Error("Invalid Argon2id iterations: %v", log.Error(err))
		return false, err
	}
	parallelism, err := requirePositiveIntWithMax(referenceCredential.Parameters.Parallelism, maxUint8, "parallelism")
	if err != nil {
		logger.Error("Invalid Argon2id parallelism: %v", log.Error(err))
		return false, err
	}
	keySize, err := requirePositiveIntWithMax(referenceCredential.Parameters.KeySize, maxUint32, "key size")
	if err != nil {
		logger.Error("Invalid Argon2id key size: %v", log.Error(err))
		return false, err
	}
	saltBytes, err := decodeSalt(referenceCredential.Parameters.Salt)
	if err != nil {
		logger.Error("Error decoding salt: %v", log.Error(err))
		return false, err
	}
	//nolint:gosec // G115 - Conversion is safe
	hash := argon2.IDKey(
		credentialValueToVerify,
		saltBytes,
		uint32(iterations),
		uint32(memory),
		uint8(parallelism),
		uint32(keySize),
	)
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

func decodeSalt(salt string) ([]byte, error) {
	if salt == "" {
		return nil, fmt.Errorf("salt must be provided")
	}
	saltBytes, err := hex.DecodeString(salt)
	if err != nil {
		return nil, err
	}
	return saltBytes, nil
}

func validateCredentialAlgorithm(referenceCredential Credential, expected CredAlgorithm) error {
	if referenceCredential.Algorithm != expected {
		return fmt.Errorf("credential algorithm mismatch: expected %s", expected)
	}
	return nil
}

func requirePositiveInt(value int, name string) (int, error) {
	if value <= 0 {
		return 0, fmt.Errorf("%s must be positive", name)
	}
	return value, nil
}

func requirePositiveIntWithMax(value, maxValue int, name string) (int, error) {
	normalized, err := requirePositiveInt(value, name)
	if err != nil {
		return 0, err
	}
	if normalized > maxValue {
		return 0, fmt.Errorf("%s exceeds maximum supported value", name)
	}
	return normalized, nil
}

func validateSHA256Config(cfg config.SHA256Config) error {
	if _, err := requirePositiveInt(cfg.SaltSize, "salt size"); err != nil {
		return err
	}
	return nil
}

func validatePBKDF2Config(cfg config.PBKDF2Config) error {
	if _, err := requirePositiveInt(cfg.SaltSize, "salt size"); err != nil {
		return err
	}
	if _, err := requirePositiveInt(cfg.Iterations, "iterations"); err != nil {
		return err
	}
	if _, err := requirePositiveInt(cfg.KeySize, "key size"); err != nil {
		return err
	}
	return nil
}

func validateArgon2idConfig(cfg config.Argon2IDConfig) error {
	if _, err := requirePositiveInt(cfg.SaltSize, "salt size"); err != nil {
		return err
	}
	if _, err := requirePositiveIntWithMax(cfg.Memory, maxUint32, "memory"); err != nil {
		return err
	}
	if _, err := requirePositiveIntWithMax(cfg.Iterations, maxUint32, "iterations"); err != nil {
		return err
	}
	if _, err := requirePositiveIntWithMax(cfg.Parallelism, maxUint8, "parallelism"); err != nil {
		return err
	}
	if _, err := requirePositiveIntWithMax(cfg.KeySize, maxUint32, "key size"); err != nil {
		return err
	}
	return nil
}
