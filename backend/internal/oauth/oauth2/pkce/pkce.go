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

// Package pkce provides PKCE (Proof Key for Code Exchange) validation utilities
package pkce

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
)

// PKCE Code Challenge Methods.
const (
	CodeChallengeMethodS256 = "S256"
)

// PKCE validation errors
var (
	ErrInvalidCodeVerifier    = errors.New("invalid code verifier")
	ErrInvalidCodeChallenge   = errors.New("invalid code challenge")
	ErrInvalidChallengeMethod = errors.New("invalid code challenge method")
	ErrPKCEValidationFailed   = errors.New("PKCE validation failed")
)

// isValidASCIIUnreserved validates that a character is in the unreserved set.
func isValidASCIIUnreserved(c rune) bool {
	return (c >= 'A' && c <= 'Z') ||
		(c >= 'a' && c <= 'z') ||
		(c >= '0' && c <= '9') ||
		c == '-' || c == '_' || c == '.' || c == '~'
}

// isValidBase64URLChar validates that a character is in the base64url alphabet.
func isValidBase64URLChar(c rune) bool {
	return (c >= 'A' && c <= 'Z') ||
		(c >= 'a' && c <= 'z') ||
		(c >= '0' && c <= '9') ||
		c == '-' || c == '_'
}

// ValidatePKCE validates the PKCE code verifier against the stored code challenge.
// Only S256 code challenge method is supported as per OAuth 2.0 Security Best Current Practice.
func ValidatePKCE(codeChallenge, codeChallengeMethod, codeVerifier string) error {
	if codeChallengeMethod != "" && codeChallengeMethod != CodeChallengeMethodS256 {
		return ErrInvalidChallengeMethod
	}

	if err := validateCodeVerifier(codeVerifier); err != nil {
		return err
	}

	if codeChallenge == "" {
		return ErrInvalidCodeChallenge
	}

	return validateS256Challenge(codeChallenge, codeVerifier)
}

// validateCodeVerifier validates the format of a code verifier according to RFC 7636.
func validateCodeVerifier(codeVerifier string) error {
	if codeVerifier == "" {
		return ErrInvalidCodeVerifier
	}
	if len(codeVerifier) < 43 || len(codeVerifier) > 128 {
		return ErrInvalidCodeVerifier
	}
	for _, c := range codeVerifier {
		if !isValidASCIIUnreserved(c) {
			return ErrInvalidCodeVerifier
		}
	}
	return nil
}

// validateS256Challenge validates an S256 code challenge.
func validateS256Challenge(codeChallenge, codeVerifier string) error {
	hash := sha256.Sum256([]byte(codeVerifier))

	expectedChallenge := base64.RawURLEncoding.EncodeToString(hash[:])
	if codeChallenge != expectedChallenge {
		return ErrPKCEValidationFailed
	}
	return nil
}

// GenerateCodeChallenge generates a code challenge from a code verifier using the specified method.
// Only S256 code challenge method is supported as per OAuth 2.0 Security Best Current Practice.
func GenerateCodeChallenge(codeVerifier, method string) (string, error) {
	if err := validateCodeVerifier(codeVerifier); err != nil {
		return "", err
	}

	if method != "" && method != CodeChallengeMethodS256 {
		return "", ErrInvalidChallengeMethod
	}

	hash := sha256.Sum256([]byte(codeVerifier))
	return base64.RawURLEncoding.EncodeToString(hash[:]), nil
}

// ValidateCodeChallenge validates the format of a code challenge according to RFC 7636.
// Only S256 code challenge method is supported as per OAuth 2.0 Security Best Current Practice.
func ValidateCodeChallenge(codeChallenge, codeChallengeMethod string) error {
	if codeChallengeMethod != "" && codeChallengeMethod != CodeChallengeMethodS256 {
		return ErrInvalidChallengeMethod
	}

	if len(codeChallenge) != 43 {
		return ErrInvalidCodeChallenge
	}

	for _, c := range codeChallenge {
		if !isValidBase64URLChar(c) {
			return ErrInvalidCodeChallenge
		}
	}
	return nil
}

// GetSupportedCodeChallengeMethods returns all supported PKCE code challenge methods.
func GetSupportedCodeChallengeMethods() []string {
	return []string{
		CodeChallengeMethodS256,
	}
}
