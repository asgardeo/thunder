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

package authnprovider

import "encoding/json"

// AuthnMetadata contains metadata for authentication.
type AuthnMetadata struct {
	// TODO: Application should have a extension metadata field
	// Those values should be fetched from there and passed to the authn provider
	AppMetadata map[string]interface{} `json:"appMetadata,omitempty"`
}

// AvailableAttribute represents an attribute available from the identity provider.
type AvailableAttribute struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Verified    bool   `json:"verified"`
}

// AuthnResult represents the result of an authentication attempt.
type AuthnResult struct {
	UserID              string               `json:"userId"`
	UserType            string               `json:"userType"`
	OrganizationUnitID  string               `json:"ouId"`
	Token               string               `json:"token"`
	AvailableAttributes []AvailableAttribute `json:"availableAttributes"`
}

// GetAttributesMetadata contains metadata for fetching attributes.
type GetAttributesMetadata struct {
	AppMetadata map[string]interface{} `json:"appMetadata,omitempty"`
	Locale      string                 `json:"locale"`
}

// GetAttributesResult represents the result of fetching attributes.
type GetAttributesResult struct {
	UserID             string          `json:"userId"`
	UserType           string          `json:"userType"`
	OrganizationUnitID string          `json:"ouId"`
	Attributes         json.RawMessage `json:"attributes,omitempty"`
}
