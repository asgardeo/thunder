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
	"encoding/json"
	"fmt"

	"github.com/asgardeo/thunder/internal/system/crypto"
)

// Property represents a generic property with name, value/values, and isSecret fields.
type Property struct {
	name          string
	value         string
	values        []string
	isSecret      bool
	isMultiValued bool
}

// PropertyDTO represents a property for API communication.
type PropertyDTO struct {
	Name          string   `json:"name"`
	Value         string   `json:"value,omitempty"`
	Values        []string `json:"values,omitempty"`
	IsSecret      bool     `json:"is_secret"`
	IsMultiValued bool     `json:"is_multi_valued"`
}

// NewProperty creates a new Property instance with the given parameters.
// If isSecret is true, the value will be automatically encrypted.
func NewProperty(name, value string, isSecret bool) (*Property, error) {
	property := &Property{
		name:          name,
		value:         value,
		isSecret:      isSecret,
		isMultiValued: false,
	}

	if isSecret && value != "" {
		if err := property.Encrypt(); err != nil {
			return nil, fmt.Errorf("failed to encrypt property %s: %w", name, err)
		}
	}

	return property, nil
}

// NewMultiValuedProperty creates a new multi-valued Property instance.
// If isSecret is true, all values will be automatically encrypted.
func NewMultiValuedProperty(name string, values []string, isSecret bool) (*Property, error) {
	property := &Property{
		name:          name,
		values:        values,
		isSecret:      isSecret,
		isMultiValued: true,
	}

	if isSecret && len(values) > 0 {
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

// IsMultiValued returns whether the property is multi-valued
func (p *Property) IsMultiValued() bool {
	return p.isMultiValued
}

// GetValue returns the decrypted value if it's a secret, otherwise returns the plain value.
// For multi-valued properties, use GetValues() instead.
func (p *Property) GetValue() (string, error) {
	if p.IsMultiValued() {
		return "", fmt.Errorf("property %s is multi-valued, use GetValues() instead", p.GetName())
	}

	if !p.IsSecret() {
		return p.value, nil
	}

	cryptoService := crypto.GetCryptoService()
	decryptedValue, err := cryptoService.DecryptString(p.value)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt secret property %s: %w", p.GetName(), err)
	}

	return decryptedValue, nil
}

// GetValues returns the decrypted values if it's a multi-valued secret, otherwise returns the plain values.
// Returns an error if called on a single-valued property.
func (p *Property) GetValues() ([]string, error) {
	if !p.IsMultiValued() {
		return nil, fmt.Errorf("property %s is not multi-valued", p.GetName())
	}

	if !p.IsSecret() {
		return p.values, nil
	}

	cryptoService := crypto.GetCryptoService()
	decryptedValues := make([]string, len(p.values))
	for i, encryptedValue := range p.values {
		decryptedValue, err := cryptoService.DecryptString(encryptedValue)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt secret property %s: %w", p.GetName(), err)
		}
		decryptedValues[i] = decryptedValue
	}

	return decryptedValues, nil
}

// Encrypt encrypts the value or values if it's a secret property
func (p *Property) Encrypt() error {
	if !p.IsSecret() {
		return nil
	}

	cryptoService := crypto.GetCryptoService()

	if p.IsMultiValued() {
		if len(p.values) == 0 {
			return nil
		}
		encryptedValues := make([]string, len(p.values))
		for i, val := range p.values {
			encryptedValue, err := cryptoService.EncryptString(val)
			if err != nil {
				return fmt.Errorf("failed to encrypt secret property %s: %w", p.GetName(), err)
			}
			encryptedValues[i] = encryptedValue
		}
		p.values = encryptedValues
	} else {
		if p.value == "" {
			return nil
		}
		encryptedValue, err := cryptoService.EncryptString(p.value)
		if err != nil {
			return fmt.Errorf("failed to encrypt secret property %s: %w", p.GetName(), err)
		}
		p.value = encryptedValue
	}

	return nil
}

// SerializePropertiesToJSONArray serializes an array of properties to a JSON array string
func SerializePropertiesToJSONArray(properties []Property) (string, error) {
	if len(properties) == 0 {
		return "", nil
	}

	propertyDTOs := make([]PropertyDTO, 0, len(properties))
	for _, property := range properties {
		propertyDTO := PropertyDTO{
			Name:          property.GetName(),
			IsSecret:      property.IsSecret(),
			IsMultiValued: property.IsMultiValued(),
		}

		if property.IsMultiValued() {
			propertyDTO.Values = property.values
		} else {
			propertyDTO.Value = property.value
		}

		propertyDTOs = append(propertyDTOs, propertyDTO)
	}

	jsonBytes, err := json.Marshal(propertyDTOs)
	if err != nil {
		return "", fmt.Errorf("failed to serialize properties to JSON: %w", err)
	}

	return string(jsonBytes), nil
}

// DeserializePropertiesFromJSON deserializes an array of properties from JSON string
func DeserializePropertiesFromJSON(propertiesJSON string) ([]Property, error) {
	if propertiesJSON == "" {
		return []Property{}, nil
	}

	var propertyDTOs []PropertyDTO
	if err := json.Unmarshal([]byte(propertiesJSON), &propertyDTOs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal properties JSON: %w", err)
	}

	properties := make([]Property, 0, len(propertyDTOs))
	for _, propertyDTO := range propertyDTOs {
		property := Property{
			name:          propertyDTO.Name,
			isSecret:      propertyDTO.IsSecret,
			isMultiValued: propertyDTO.IsMultiValued,
		}

		if propertyDTO.IsMultiValued {
			property.values = propertyDTO.Values
		} else {
			property.value = propertyDTO.Value
		}

		properties = append(properties, property)
	}

	return properties, nil
}

// ToProperty converts PropertyDTO to Property.
func (dto *PropertyDTO) ToProperty() (*Property, error) {
	if dto.IsMultiValued {
		return NewMultiValuedProperty(dto.Name, dto.Values, dto.IsSecret)
	}
	return NewProperty(dto.Name, dto.Value, dto.IsSecret)
}

// ToPropertyDTO converts Property to PropertyDTO.
func (p *Property) ToPropertyDTO() (*PropertyDTO, error) {
	dto := &PropertyDTO{
		Name:          p.GetName(),
		IsSecret:      p.IsSecret(),
		IsMultiValued: p.IsMultiValued(),
	}

	if p.IsMultiValued() {
		values, err := p.GetValues()
		if err != nil {
			return nil, fmt.Errorf("failed to get values for property %s: %w", p.GetName(), err)
		}
		dto.Values = values
	} else {
		value, err := p.GetValue()
		if err != nil {
			return nil, fmt.Errorf("failed to get value for property %s: %w", p.GetName(), err)
		}
		dto.Value = value
	}

	return dto, nil
}