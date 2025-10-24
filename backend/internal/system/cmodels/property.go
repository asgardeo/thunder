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
	"strings"

	"github.com/asgardeo/thunder/internal/system/crypto"
)

// Property represents a generic property with name, value, and isSecret fields.
type Property struct {
	name     string
	value    string
	isSecret bool
}

// PropertyDTO represents a property for API communication.
type PropertyDTO struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	IsSecret bool   `json:"is_secret"`
}

// NewProperty creates a new Property instance with the given parameters.
// If isSecret is true, the value will be automatically encrypted.
func NewProperty(name, value string, isSecret bool) (*Property, error) {
	property := &Property{
		name:     name,
		value:    value,
		isSecret: isSecret,
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

// GetValue returns the decrypted value if it's a secret, otherwise returns the plain value
func (p *Property) GetValue() (string, error) {
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

// Encrypt encrypts the value if it's a secret property
func (p *Property) Encrypt() error {
	if !p.IsSecret() || p.value == "" {
		return nil
	}

	cryptoService := crypto.GetCryptoService()
	encryptedValue, err := cryptoService.EncryptString(p.value)
	if err != nil {
		return fmt.Errorf("failed to encrypt secret property %s: %w", p.GetName(), err)
	}

	p.value = encryptedValue
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
			Name:     property.GetName(),
			Value:    property.value,
			IsSecret: property.IsSecret(),
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
			name:     propertyDTO.Name,
			value:    propertyDTO.Value,
			isSecret: propertyDTO.IsSecret,
		}
		properties = append(properties, property)
	}

	return properties, nil
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
// PropertyValueDTO represents a property value that can be either a string or array of strings
type PropertyValueDTO struct {
	single   string
	multiple []string
	isArray  bool
}

// UnmarshalJSON implements json.Unmarshaler for PropertyValueDTO
// This handles both string values and array values from JSON
func (pv *PropertyValueDTO) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as array first
	var arr []string
	if err := json.Unmarshal(data, &arr); err == nil {
		pv.multiple = arr
		pv.isArray = true
		return nil
	}

	// If not array, treat as string
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	pv.single = str
	pv.isArray = false
	return nil
}

// MarshalJSON implements json.Marshaler for PropertyValueDTO
func (pv PropertyValueDTO) MarshalJSON() ([]byte, error) {
	if pv.isArray {
		return json.Marshal(pv.multiple)
	}
	return json.Marshal(pv.single)
}

// AsString returns the value as a string
// For arrays, returns comma-separated values
func (pv PropertyValueDTO) AsString() string {
	if pv.isArray {
		return strings.Join(pv.multiple, ",")
	}
	return pv.single
}

// AsArray returns the value as a string array
// Handles backward compatibility with comma-separated strings
func (pv PropertyValueDTO) AsArray() []string {
	if pv.isArray {
		return pv.multiple
	}

	// Backward compatibility: split comma-separated strings
	if pv.single == "" {
		return []string{}
	}

	parts := strings.Split(pv.single, ",")
	result := make([]string, len(parts))
	for i, part := range parts {
		result[i] = strings.TrimSpace(part)
	}
	return result
}

// IsArray checks if the property is multi-valued
func (pv PropertyValueDTO) IsArray() bool {
	return pv.isArray
}

// NewStringPropertyValue creates a single-valued property
func NewStringPropertyValue(value string) PropertyValueDTO {
	return PropertyValueDTO{
		single:  value,
		isArray: false,
	}
}

// NewArrayPropertyValue creates a multi-valued property
func NewArrayPropertyValue(values []string) PropertyValueDTO {
	return PropertyValueDTO{
		multiple: values,
		isArray:  true,
	}
}

// PropertiesMap represents a map of property names to values
type PropertiesMap map[string]PropertyValueDTO

// GetString retrieves a single-valued property
func (pm PropertiesMap) GetString(key string) string {
	if val, ok := pm[key]; ok {
		return val.AsString()
	}
	return ""
}

// GetArray retrieves a multi-valued property
func (pm PropertiesMap) GetArray(key string) []string {
	if val, ok := pm[key]; ok {
		return val.AsArray()
	}
	return []string{}
}

// SetString sets a single-valued property
func (pm PropertiesMap) SetString(key, value string) {
	pm[key] = NewStringPropertyValue(value)
}

// SetArray sets a multi-valued property
func (pm PropertiesMap) SetArray(key string, values []string) {
	pm[key] = NewArrayPropertyValue(values)
}
