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

// Package cmodels provides common data models used across Thunder modules.
package cmodels

import (
	"fmt"

	"github.com/asgardeo/thunder/internal/system/crypto"
)

// Property represents a generic property with Name, Value, IsSecret, and isEncrypted fields.
type Property struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	IsSecret    bool   `json:"is_secret"`
	isEncrypted bool
}

// IsEncrypted returns whether the property value is encrypted
func (p *Property) IsEncrypted() bool {
	return p.isEncrypted
}

// SetEncrypted sets the encryption state of the property
func (p *Property) SetEncrypted(encrypted bool) {
	p.isEncrypted = encrypted
}

// GetValue returns the decrypted value if it's a secret, otherwise returns the plain value
func (p *Property) GetValue() (string, error) {
	if !p.IsSecret || !p.IsEncrypted() {
		return p.Value, nil
	}

	cryptoService := crypto.GetCryptoService()
	decryptedValue, err := cryptoService.DecryptString(p.Value)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt secret property %s: %w", p.Name, err)
	}

	return decryptedValue, nil
}

// Encrypt encrypts the value if it's a secret property
func (p *Property) Encrypt() error {
	if !p.IsSecret || p.Value == "" || p.IsEncrypted() {
		return nil
	}

	cryptoService := crypto.GetCryptoService()
	encryptedValue, err := cryptoService.EncryptString(p.Value)
	if err != nil {
		return fmt.Errorf("failed to encrypt secret property %s: %w", p.Name, err)
	}

	p.Value = encryptedValue
	p.SetEncrypted(true)
	return nil
}
