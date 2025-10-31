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

package testutils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// JWTClaims represents the decoded JWT claims.
type JWTClaims struct {
	Sub       string                 `json:"sub"`
	Aud       string                 `json:"aud"`
	Iss       string                 `json:"iss"`
	Exp       float64                `json:"exp"`
	Iat       float64                `json:"iat"`
	UserType  string                 `json:"userType,omitempty"`
	OuID      string                 `json:"ouId,omitempty"`
	Assurance map[string]interface{} `json:"assurance,omitempty"`
	// Additional claims can be accessed via the map
	Additional map[string]interface{} `json:"-"`
}

// DecodeJWT decodes a JWT assertion without verifying the signature.
// This is useful for integration tests to validate the JWT payload content.
func DecodeJWT(token string) (*JWTClaims, error) {
	// Split the JWT into its three parts
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format: expected 3 parts, got %d", len(parts))
	}

	// Decode the payload (second part)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT payload: %w", err)
	}

	// Unmarshal the payload into a map first to capture all claims
	var allClaims map[string]interface{}
	if err := json.Unmarshal(payload, &allClaims); err != nil {
		return nil, fmt.Errorf("failed to parse JWT payload: %w", err)
	}

	// Unmarshal into the structured JWTClaims
	var claims JWTClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse JWT claims: %w", err)
	}

	// Store all claims in Additional for flexible access
	claims.Additional = allClaims

	return &claims, nil
}

// ValidateJWTClaims validates that the JWT contains the expected claims.
func ValidateJWTClaims(claims *JWTClaims, expectedSub, expectedAud, expectedUserType, expectedOuID string) error {
	if claims.Sub != expectedSub {
		return fmt.Errorf("expected sub '%s', got '%s'", expectedSub, claims.Sub)
	}
	if claims.Aud != expectedAud {
		return fmt.Errorf("expected aud '%s', got '%s'", expectedAud, claims.Aud)
	}
	if expectedUserType != "" && claims.UserType != expectedUserType {
		return fmt.Errorf("expected userType '%s', got '%s'", expectedUserType, claims.UserType)
	}
	if expectedOuID != "" && claims.OuID != expectedOuID {
		return fmt.Errorf("expected ouId '%s', got '%s'", expectedOuID, claims.OuID)
	}
	return nil
}
