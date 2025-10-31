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

package idp

import (
	"fmt"

	"github.com/asgardeo/thunder/internal/system/cmodels"
)

// IDPDTO represents the data transfer object for an identity provider.
type IDPDTO struct {
	ID          string
	Name        string
	Description string
	Type        IDPType
	Properties  []cmodels.Property
}

// BasicIDPDTO represents a basic data transfer object for an identity provider.
type BasicIDPDTO struct {
	ID          string
	Name        string
	Description string
	Type        IDPType
}

// idpRequest represents the request payload for creating or updating an identity provider.
type idpRequest struct {
	Name        string                `json:"name"`
	Description string                `json:"description,omitempty"`
	Type        string                `json:"type"`
	Properties  []cmodels.PropertyDTO `json:"properties,omitempty"`
}

// idpResponse represents the response payload for an identity provider.
type idpResponse struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description,omitempty"`
	Type        string                `json:"type"`
	Properties  []cmodels.PropertyDTO `json:"properties,omitempty"`
}

// basicIDPResponse represents a basic response payload for an identity provider.
type basicIDPResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type"`
}

// GetPropertyValue returns the property value with the given name.
// Returns an empty string if the property is not found.
// Returns an error if retrieving the property value fails or if the property is multi-valued.
func (dto *IDPDTO) GetPropertyValue(name string) (string, error) {
	for _, prop := range dto.Properties {
		if prop.GetName() == name {
			if prop.IsMultiValued() {
				return "", fmt.Errorf("property %s is multi-valued, use GetPropertyValues() instead", name)
			}
			return prop.GetValue()
		}
	}
	return "", nil
}

// GetPropertyValues returns the property values with the given name for multi-valued properties.
// Returns an empty slice if the property is not found.
// Returns an error if retrieving the property values fails.
// For backward compatibility, single-valued properties are returned as a single-element array.
func (dto *IDPDTO) GetPropertyValues(name string) ([]string, error) {
	for _, prop := range dto.Properties {
		if prop.GetName() == name {
			if !prop.IsMultiValued() {
				// Backward compatibility: return single value as array
				value, err := prop.GetValue()
				if err != nil {
					return nil, err
				}
				if value == "" {
					return []string{}, nil
				}
				return []string{value}, nil
			}
			return prop.GetValues()
		}
	}
	return []string{}, nil
}

// GetScopes is a convenience method for getting OAuth scopes.
// Returns the scopes as an array, whether stored as single or multi-valued property.
func (dto *IDPDTO) GetScopes() ([]string, error) {
	return dto.GetPropertyValues("scopes")
}