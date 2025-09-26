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

// Package utils provides utility functions for OAuth2 operations.
package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"regexp"

	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/system/utils"
)

// GetURIWithQueryParams constructs a URI with the given query parameters.
// It validates the error code and error description according to the spec.
func GetURIWithQueryParams(uri string, queryParams map[string]string) (string, error) {
	// Validate the error params if present.
	if err := validateErrorParams(queryParams[constants.RequestParamError],
		queryParams[constants.RequestParamErrorDescription]); err != nil {
		return "", err
	}

	return utils.GetURIWithQueryParams(uri, queryParams)
}

// validateErrorParams validates the error code and error description parameters.
func validateErrorParams(err, desc string) error {
	// Define a regex pattern for the allowed character set: %x20-21 / %x23-5B / %x5D-7E
	allowedCharPattern := `^[\x20-\x21\x23-\x5B\x5D-\x7E]*$`
	allowedCharRegex := regexp.MustCompile(allowedCharPattern)

	// Validate the error code.
	if err != "" && !allowedCharRegex.MatchString(err) {
		return fmt.Errorf("invalid error code: %s", err)
	}

	// Validate the error description.
	if desc != "" && !allowedCharRegex.MatchString(desc) {
		return fmt.Errorf("invalid error description: %s", desc)
	}

	return nil
}

const (
	// OAuth2ClientIDLength specifies the byte length for OAuth client IDs (16 bytes = 128 bits)
	// This provides sufficient entropy while keeping the resulting base64 string reasonably short
	OAuth2ClientIDLength = 16

	// OAuth2ClientSecretLength specifies the byte length for OAuth client secrets (32 bytes = 256 bits)
	// This provides high entropy for cryptographic security as recommended by OAuth security best practices
	OAuth2ClientSecretLength = 32
)

// OAuth2CredentialType represents the type of OAuth 2.0 credential to generate
type OAuth2CredentialType string

const (
	// ClientIDCredential represents an OAuth 2.0 client identifier
	ClientIDCredential OAuth2CredentialType = "client ID"

	// ClientSecretCredential represents an OAuth 2.0 client secret
	ClientSecretCredential OAuth2CredentialType = "client secret"
)

// generateOAuth2Credential generates a base64url-encoded OAuth 2.0 credential.
// This private method contains the common logic for generating both client IDs and secrets.
// The length is automatically determined based on the credential type to ensure OAuth compliance.
func generateOAuth2Credential(credentialType OAuth2CredentialType) (string, error) {
	var length int

	switch credentialType {
	case ClientIDCredential:
		length = OAuth2ClientIDLength
	case ClientSecretCredential:
		length = OAuth2ClientSecretLength
	default:
		return "", fmt.Errorf("unsupported credential type: %s", credentialType)
	}

	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes for OAuth %s: %w", credentialType, err)
	}

	// Use base64 URL encoding without padding for web-friendly credentials
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// GenerateOAuth2ClientID generates a URL-safe OAuth 2.0 client identifier.
// Returns a base64url-encoded string (no padding) that is web-friendly and compliant
// with OAuth 2.1 specifications for client identifier format.
//
// The generated client ID:
// - Uses cryptographically secure random bytes
// - Is URL-safe (base64url encoding without padding)
// - Has sufficient entropy (128 bits) for uniqueness
// - Results in a ~22 character string (more compact than UUID)
func GenerateOAuth2ClientID() (string, error) {
	return generateOAuth2Credential(ClientIDCredential)
}

// GenerateOAuth2ClientSecret generates a cryptographically secure OAuth 2.0 client secret.
// Returns a base64url-encoded string with high entropy suitable for client authentication.
//
// The generated client secret:
// - Uses cryptographically secure random bytes
// - Has high entropy (256 bits) for security
// - Is base64url-encoded for safe transport/storage
// - Meets OAuth Security BCP (RFC 6819) recommendations
func GenerateOAuth2ClientSecret() (string, error) {
	return generateOAuth2Credential(ClientSecretCredential)
}
