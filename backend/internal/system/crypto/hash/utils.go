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
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

// GenerateSalt generates a random salt string.
func GenerateSalt() ([]byte, error) {
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
	return hex.EncodeToString(hash[:])
}

// GenerateThumbprintFromString generates a SHA-256 thumbprint for the given string data.
func GenerateThumbprintFromString(data string) string {
	return GenerateThumbprint([]byte(data))
}
