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

// Package hash provides stateful credential hashing services and generic hashing utilities.
// Initialize must be called with a fully populated HashConfig; no config system is read here.
package hash

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/argon2"
)

const (
	maxUint8  = int(^uint8(0))
	maxUint32 = int(^uint32(0))
)

// HashServiceInterface defines the interface for credential hashing services.
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

func newHashService(cfg HashConfig) (HashServiceInterface, error) {
	switch cfg.Algorithm {
	case SHA256:
		if err := validatePositiveInt(cfg.SaltSize, "salt size"); err != nil {
			return nil, err
		}
		return newSHA256Provider(cfg.SaltSize), nil
	case PBKDF2:
		if err := validatePositiveInt(cfg.SaltSize, "salt size"); err != nil {
			return nil, err
		}
		if err := validatePositiveInt(cfg.Iterations, "iterations"); err != nil {
			return nil, err
		}
		if err := validatePositiveInt(cfg.KeySize, "key size"); err != nil {
			return nil, err
		}
		return newPBKDF2Provider(cfg.SaltSize, cfg.Iterations, cfg.KeySize), nil
	case ARGON2ID:
		if err := validatePositiveInt(cfg.SaltSize, "salt size"); err != nil {
			return nil, err
		}
		if err := validatePositiveIntWithMax(cfg.Memory, maxUint32, "memory"); err != nil {
			return nil, err
		}
		if err := validatePositiveIntWithMax(cfg.Iterations, maxUint32, "iterations"); err != nil {
			return nil, err
		}
		if err := validatePositiveIntWithMax(cfg.Parallelism, maxUint8, "parallelism"); err != nil {
			return nil, err
		}
		if err := validatePositiveIntWithMax(cfg.KeySize, maxUint32, "key size"); err != nil {
			return nil, err
		}
		return newArgon2idProvider(cfg.SaltSize, cfg.Memory, cfg.Iterations, cfg.Parallelism, cfg.KeySize), nil
	default:
		return nil, fmt.Errorf("unsupported hash algorithm: %s", cfg.Algorithm)
	}
}

func newSHA256Provider(saltSize int) *sha256HashProvider {
	return &sha256HashProvider{SaltSize: saltSize}
}

func (a *sha256HashProvider) Generate(credentialValue []byte) (Credential, error) {
	credSalt, err := generateSalt(a.SaltSize)
	if err != nil {
		return Credential{}, err
	}
	credentialWithSalt := append([]byte(nil), credentialValue...)
	credentialWithSalt = append(credentialWithSalt, credSalt...)
	h := sha256.Sum256(credentialWithSalt)
	return Credential{
		Algorithm: SHA256,
		Hash:      hex.EncodeToString(h[:]),
		Parameters: CredParameters{
			Salt: hex.EncodeToString(credSalt),
		},
	}, nil
}

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
	referenceHash, err := hex.DecodeString(referenceCredential.Hash)
	if err != nil {
		return false, err
	}
	return subtle.ConstantTimeCompare(hashedData[:], referenceHash) == 1, nil
}

func newPBKDF2Provider(saltSize, iterations, keySize int) *pbkdf2HashProvider {
	return &pbkdf2HashProvider{
		SaltSize:   saltSize,
		Iterations: iterations,
		KeySize:    keySize,
	}
}

func (a *pbkdf2HashProvider) Generate(credentialValue []byte) (Credential, error) {
	credSalt, err := generateSalt(a.SaltSize)
	if err != nil {
		return Credential{}, err
	}
	h, err := pbkdf2.Key(sha256.New, string(credentialValue), credSalt, a.Iterations, a.KeySize)
	if err != nil {
		return Credential{}, err
	}
	return Credential{
		Algorithm: PBKDF2,
		Hash:      hex.EncodeToString(h),
		Parameters: CredParameters{
			Iterations: a.Iterations,
			KeySize:    a.KeySize,
			Salt:       hex.EncodeToString(credSalt),
		},
	}, nil
}

func (a *pbkdf2HashProvider) Verify(credentialValueToVerify []byte, referenceCredential Credential) (bool, error) {
	if err := validateCredentialAlgorithm(referenceCredential, PBKDF2); err != nil {
		return false, err
	}
	iterations, err := requirePositiveInt(referenceCredential.Parameters.Iterations, "iterations")
	if err != nil {
		return false, err
	}
	keySize, err := requirePositiveInt(referenceCredential.Parameters.KeySize, "key size")
	if err != nil {
		return false, err
	}
	saltBytes, err := decodeSalt(referenceCredential.Parameters.Salt)
	if err != nil {
		return false, err
	}
	h, err := pbkdf2.Key(sha256.New, string(credentialValueToVerify), saltBytes, iterations, keySize)
	if err != nil {
		return false, err
	}
	referenceHash, err := hex.DecodeString(referenceCredential.Hash)
	if err != nil {
		return false, err
	}
	return subtle.ConstantTimeCompare(h, referenceHash) == 1, nil
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

func (a *argon2idHashProvider) Generate(credentialValue []byte) (Credential, error) {
	credSalt, err := generateSalt(a.SaltSize)
	if err != nil {
		return Credential{}, err
	}
	//nolint:gosec // G115 - Conversion is safe
	h := argon2.IDKey(
		credentialValue,
		credSalt,
		uint32(a.Iterations),
		uint32(a.Memory),
		uint8(a.Parallelism),
		uint32(a.KeySize),
	)
	return Credential{
		Algorithm: ARGON2ID,
		Hash:      hex.EncodeToString(h),
		Parameters: CredParameters{
			Memory:      a.Memory,
			Iterations:  a.Iterations,
			Parallelism: a.Parallelism,
			KeySize:     a.KeySize,
			Salt:        hex.EncodeToString(credSalt),
		},
	}, nil
}

func (a *argon2idHashProvider) Verify(credentialValueToVerify []byte, referenceCredential Credential) (bool, error) {
	if err := validateCredentialAlgorithm(referenceCredential, ARGON2ID); err != nil {
		return false, err
	}
	memory, err := requirePositiveIntWithMax(referenceCredential.Parameters.Memory, maxUint32, "memory")
	if err != nil {
		return false, err
	}
	iterations, err := requirePositiveIntWithMax(referenceCredential.Parameters.Iterations, maxUint32, "iterations")
	if err != nil {
		return false, err
	}
	parallelism, err := requirePositiveIntWithMax(referenceCredential.Parameters.Parallelism, maxUint8, "parallelism")
	if err != nil {
		return false, err
	}
	keySize, err := requirePositiveIntWithMax(referenceCredential.Parameters.KeySize, maxUint32, "key size")
	if err != nil {
		return false, err
	}
	saltBytes, err := decodeSalt(referenceCredential.Parameters.Salt)
	if err != nil {
		return false, err
	}
	//nolint:gosec // G115 - Conversion is safe
	h := argon2.IDKey(
		credentialValueToVerify,
		saltBytes,
		uint32(iterations),
		uint32(memory),
		uint8(parallelism),
		uint32(keySize),
	)
	referenceHash, err := hex.DecodeString(referenceCredential.Hash)
	if err != nil {
		return false, err
	}
	return subtle.ConstantTimeCompare(h, referenceHash) == 1, nil
}

func generateSalt(saltSize int) ([]byte, error) {
	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	return salt, nil
}

func decodeSalt(salt string) ([]byte, error) {
	if salt == "" {
		return nil, fmt.Errorf("salt must be provided")
	}
	return hex.DecodeString(salt)
}

func validateCredentialAlgorithm(referenceCredential Credential, expected CredAlgorithm) error {
	if referenceCredential.Algorithm != expected {
		return fmt.Errorf("credential algorithm mismatch: expected %s", expected)
	}
	return nil
}

func validatePositiveInt(value int, name string) error {
	_, err := requirePositiveInt(value, name)
	return err
}

func validatePositiveIntWithMax(value, maxValue int, name string) error {
	_, err := requirePositiveIntWithMax(value, maxValue, name)
	return err
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
