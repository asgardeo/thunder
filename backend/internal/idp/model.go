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
	"strings"

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
// Returns an error if retrieving the property value fails.
func (dto *IDPDTO) GetPropertyValue(name string) (string, error) {
	for _, prop := range dto.Properties {
		if prop.GetName() == name {
			return prop.GetValue()
		}
	}
	return "", nil
}

func (dto *IDPDTO) GetPropertyAsArray(name string) ([]string, error) {
	value, err := dto.GetPropertyValue(name)
	if err != nil {
		return nil, err
	}
	
	if value == "" {
		return []string{}, nil
	}
	
	// Split comma-separated values first (preferred format)
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	
	// Backward compatibility: If only one item and it contains spaces but no commas,
	// treat it as space-separated (for legacy OAuth scopes like "openid profile email")
	if len(result) == 1 && strings.Contains(value, " ") && !strings.Contains(value, ",") {
		parts = strings.Split(value, " ")
		result = make([]string, 0, len(parts))
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
	}
	
	return result, nil
}

// GetScopes is a convenience method for getting OAuth scopes
func (dto *IDPDTO) GetScopes() ([]string, error) {
	return dto.GetPropertyAsArray("scopes")
}
