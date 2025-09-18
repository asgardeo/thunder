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

// Package model defines the data structures and interfaces for IdP management.
package model

// IdpDTO represents the data transfer object for an identity provider.
type IdpDTO struct {
	ID          string
	Name        string
	Description string
	Properties  []IdpProperty
}

// BasicIdpDTO represents a basic data transfer object for an identity provider.
type BasicIdpDTO struct {
	ID          string
	Name        string
	Description string
}

// IdpProperty represents a property of an identity provider.
type IdpProperty struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	IsSecret bool   `json:"is_secret"`
}

// IdpRequest represents the request payload for creating or updating an identity provider.
type IdpRequest struct {
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	Properties  []IdpProperty `json:"properties,omitempty"`
}

// IdpResponse represents the response payload for an identity provider.
type IdpResponse struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	Properties  []IdpProperty `json:"properties,omitempty"`
}

// BasicIdpResponse represents a basic response payload for an identity provider.
type BasicIdpResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}
