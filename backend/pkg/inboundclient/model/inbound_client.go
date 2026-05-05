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

// Package model holds the public data types for the inbound client subsystem. Kept as a leaf
// package (no service / store / flow dependencies) so any consumer — including
// application/model — can import it without forming a dependency cycle through the
// inboundclient parent package.
//
//nolint:lll
package model

import (
	"github.com/asgardeo/thunder/internal/cert"
)

// InboundClient is the protocol-agnostic inbound client record for a principal entity.
// Identity data (name, description, clientId, credentials) lives in the entity layer.
type InboundClient struct {
	ID                        string
	AuthFlowID                string
	RegistrationFlowID        string
	IsRegistrationFlowEnabled bool
	ThemeID                   string
	LayoutID                  string
	Assertion                 *AssertionConfig
	LoginConsent              *LoginConsentConfig
	AllowedEntityTypes        []string
	Properties                map[string]interface{}
	IsReadOnly                bool
}

// AssertionConfig is the entity-level (root) assertion config that token configs fall back to.
type AssertionConfig struct {
	ValidityPeriod int64    `json:"validityPeriod,omitempty" yaml:"validity_period,omitempty" jsonschema:"Assertion validity period in seconds."`
	UserAttributes []string `json:"userAttributes,omitempty" yaml:"user_attributes,omitempty" jsonschema:"User attributes to include in the assertion."`
}

// LoginConsentConfig is the login consent configuration for an inbound client.
type LoginConsentConfig struct {
	ValidityPeriod int64 `json:"validityPeriod" yaml:"validity_period" jsonschema:"Consent validity period in seconds. 0 means never expire."`
}

// Certificate is a user-supplied certificate input for an inbound client.
type Certificate struct {
	Type  cert.CertificateType `json:"type,omitempty" yaml:"type,omitempty" jsonschema:"Certificate type (PEM, JWK, etc.)."`
	Value string               `json:"value,omitempty" yaml:"value,omitempty" jsonschema:"Certificate value in the format specified by type."`
}

// DeclarativeLoaderConfig describes how to load inbound clients from a caller-owned YAML
// resource directory.
type DeclarativeLoaderConfig struct {
	ResourceType  string
	DirectoryName string
	Parser        func(data []byte) (*InboundClient, error)
	Validator     func(*InboundClient) error
}
