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
// GetValues returns the multi-valued property values.
func (p *Property) GetValues() []string {
	return p.values
}
	name        string
	value       string
	values      []string
	isSecret    bool
	isEncrypted bool
}

// PropertyDTO represents a property for API communication.
type PropertyDTO struct {
// GetValues returns the multi-valued property values.
func (p *PropertyDTO) GetValues() []string {
	return p.Values
}
	Name     string   `json:"name"`
	Value    string   `json:"value,omitempty"`
	Values   []string `json:"values,omitempty"`
	IsSecret bool     `json:"is_secret"`
}

// NewRawProperty creates a new Property instance with the given parameters.
func NewRawProperty(name, value string, isSecret, isEncrypted bool) *Property {
	return &Property{
		name:        name,
		value:       value,
		isSecret:    isSecret,
		isEncrypted: isEncrypted,
	}
}

// NewProperty creates a new Property instance with the given parameters.
// If isSecret is true, the value will be automatically encrypted.
func NewProperty(name, value string, isSecret bool) (*Property, error) {
	property := &Property{
		name:        name,
		value:       value,
		isSecret:    isSecret,
		isEncrypted: false,
	}

	if isSecret && value != "" {
		if err := property.Encrypt(); err != nil {
			return nil, fmt.Errorf("failed to encrypt property %s: %w", name, err)
		}
	}

	return property, nil
}

// GetName returns the name of the property
func (p *Property) GetName() string {
	return p.name
}

// IsSecret returns whether the property is a secret
func (p *Property) IsSecret() bool {
	return p.isSecret
}

// IsEncrypted returns whether the property value is encrypted
func (p *Property) IsEncrypted() bool {
	return p.isEncrypted
}

// SetEncrypted sets the encryption state of the property
func (p *Property) SetEncrypted(encrypted bool) {
	p.isEncrypted = encrypted
}

// GetStorageValue returns the value as is
func (p *Property) GetStorageValue() string {
	return p.value
}

// GetValue returns the decrypted value if it's a secret, otherwise returns the plain value
func (p *Property) GetValue() (string, error) {
	if !p.IsSecret() || !p.IsEncrypted() {
		return p.value, nil
	}

	cryptoService := crypto.GetCryptoService()
	decryptedValue, err := cryptoService.DecryptString(p.value)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt secret property %s: %w", p.GetName(), err)
	}

	return decryptedValue, nil
}

// Encrypt encrypts the value if it's a secret property
func (p *Property) Encrypt() error {
	if !p.IsSecret() || p.value == "" || p.IsEncrypted() {
		return nil
	}

	cryptoService := crypto.GetCryptoService()
	encryptedValue, err := cryptoService.EncryptString(p.value)
	if err != nil {
		return fmt.Errorf("failed to encrypt secret property %s: %w", p.GetName(), err)
	}

	p.value = encryptedValue
	p.SetEncrypted(true)
	return nil
}

// ToProperty converts PropertyDTO to Property.
func (dto *PropertyDTO) ToProperty() (*Property, error) {
	return NewProperty(dto.Name, dto.Value, dto.IsSecret)
}

// ToPropertyDTO converts Property to PropertyDTO.
func (p *Property) ToPropertyDTO() (*PropertyDTO, error) {
	value, err := p.GetValue()
	if err != nil {
		return nil, fmt.Errorf("failed to get value for property %s: %w", p.GetName(), err)
	}

	return &PropertyDTO{
		Name:     p.GetName(),
		Value:    value,
		IsSecret: p.IsSecret(),
	}, nil
}
