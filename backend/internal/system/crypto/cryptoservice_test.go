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
	"testing"
)

// Mock config for testing
type MockThunderRuntime struct {
	Config struct {
		Crypto struct {
			Key string
		}
	}
}

func TestCryptoService(t *testing.T) {
	// Generate a random key
	key, err := GenerateRandomKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	t.Logf("Generated random key: %x", key)

	// Create crypto service
	service, err := NewCryptoService(key)
	if err != nil {
		t.Fatalf("Failed to create crypto service: %v", err)
	}

	// Test data
	original := "This is a secret message that needs encryption!"

	// Encrypt
	encrypted, err := service.EncryptString(original)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	t.Logf("Encrypted data: %x", encrypted)

	// Decrypt
	decrypted, err := service.DecryptString(encrypted)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	t.Logf("Decrypted data: %s", decrypted)

	// Verify
	if decrypted != original {
		t.Errorf("Decryption result doesn't match original. Got %q, want %q", decrypted, original)
	}
}

func TestInvalidKey(t *testing.T) {
	// Test with invalid key size
	_, err := NewCryptoService([]byte("too-short-key"))
	if err == nil {
		t.Error("Expected error for invalid key size, but got none")
	}
}

func TestTampering(t *testing.T) {
	// Generate a random key
	key, _ := GenerateRandomKey()
	service, _ := NewCryptoService(key)

	// Encrypt some data
	original := "Protected data"
	encrypted, _ := service.EncryptString(original)

	// Tamper with the encrypted data
	tamperedBytes := []byte(encrypted)
	if len(tamperedBytes) > 10 {
		tamperedBytes[len(tamperedBytes)-5] ^= 0x01 // Flip a bit
	}
	tampered := string(tamperedBytes)

	// Attempt to decrypt tampered data
	_, err := service.DecryptString(tampered)
	if err == nil {
		t.Error("Expected decryption of tampered data to fail, but it succeeded")
	}
}

func TestEncryptDecryptCycle(t *testing.T) {
	// Generate a key
	key, _ := GenerateRandomKey()
	service, _ := NewCryptoService(key)

	// Test various data types
	testCases := []string{
		"",                               // Empty string
		"Hello World",                    // Simple text
		"特殊文字列",                          // Unicode characters
		"123456789012345678901234567890", // Long string
		`{"name":"John","age":30}`,       // JSON string
	}

	for _, tc := range testCases {
		encrypted, err := service.EncryptString(tc)
		if err != nil {
			t.Errorf("Failed to encrypt %q: %v", tc, err)
			continue
		}

		decrypted, err := service.DecryptString(encrypted)
		if err != nil {
			t.Errorf("Failed to decrypt %q: %v", tc, err)
			continue
		}

		if decrypted != tc {
			t.Errorf("Decryption result doesn't match original. Got %q, want %q", decrypted, tc)
		}
		t.Logf("Decryption successful. Decrypted data: %q", decrypted)
	}
}

func TestDifferentKeysEncryption(t *testing.T) {
	// Generate two different keys
	key1, _ := GenerateRandomKey()
	key2, _ := GenerateRandomKey()

	service1, _ := NewCryptoService(key1)
	service2, _ := NewCryptoService(key2)

	// Encrypt with first service
	original := "Secret message"
	encrypted, err := service1.EncryptString(original)
	if err != nil {
		t.Fatalf("Encryption with first key failed: %v", err)
	}

	// Try to decrypt with second service (should fail)
	_, err = service2.DecryptString(encrypted)
	if err == nil {
		t.Error("Expected decryption with different key to fail, but it succeeded")
	}
}
