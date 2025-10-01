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

import "github.com/asgardeo/thunder/internal/system/cmodels"

// IDPDTO represents the data transfer object for an identity provider.
type IDPDTO struct {
	ID          string
	Name        string
	Description string
	Type        IDPType
	IsEnabled   bool
	Properties  []cmodels.Property
}

// BasicIDPDTO represents a basic data transfer object for an identity provider.
type BasicIDPDTO struct {
	ID          string
	Name        string
	Description string
	Type        IDPType
	IsEnabled   bool
}

// idpRequest represents the request payload for creating or updating an identity provider.
type idpRequest struct {
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Type        IDPType            `json:"type"`
	IsEnabled   bool               `json:"is_enabled"`
	Properties  []cmodels.Property `json:"properties,omitempty"`
}

// idpResponse represents the response payload for an identity provider.
type idpResponse struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Type        IDPType            `json:"type"`
	IsEnabled   bool               `json:"is_enabled"`
	Properties  []cmodels.Property `json:"properties,omitempty"`
}

// basicIDPResponse represents a basic response payload for an identity provider.
type basicIDPResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Type        IDPType `json:"type"`
	IsEnabled   bool    `json:"is_enabled"`
}
