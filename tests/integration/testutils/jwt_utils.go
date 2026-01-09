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
	OuName    string                 `json:"ouName,omitempty"`
	OuHandle  string                 `json:"ouHandle,omitempty"`
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

// ValidateJWTAssertionFields validates common JWT assertion fields including OU details.
// This is a helper function that can be used across integration tests to validate JWT assertions.
func ValidateJWTAssertionFields(assertion, expectedAudience, expectedUserType,
	expectedOuID, expectedOuName, expectedOuHandle string) (*JWTClaims, error) {
	// Decode JWT
	jwtClaims, err := DecodeJWT(assertion)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT assertion: %w", err)
	}

	if jwtClaims == nil {
		return nil, fmt.Errorf("JWT claims should not be nil")
	}

	// Validate audience
	if jwtClaims.Aud != expectedAudience {
		return nil, fmt.Errorf("expected aud to be '%s', got '%s'", expectedAudience, jwtClaims.Aud)
	}

	// Validate subject is not empty
	if jwtClaims.Sub == "" {
		return nil, fmt.Errorf("JWT subject should not be empty")
	}

	// Validate user type if provided
	if expectedUserType != "" && jwtClaims.UserType != expectedUserType {
		return nil, fmt.Errorf("expected userType to be '%s', got '%s'", expectedUserType, jwtClaims.UserType)
	}

	// Validate OU ID if provided
	if expectedOuID != "" && jwtClaims.OuID != expectedOuID {
		return nil, fmt.Errorf("expected ouId to be '%s', got '%s'", expectedOuID, jwtClaims.OuID)
	}

	// Validate OU name if provided
	if expectedOuName != "" && jwtClaims.OuName != expectedOuName {
		return nil, fmt.Errorf("expected ouName to be '%s', got '%s'", expectedOuName, jwtClaims.OuName)
	}

	// Validate OU handle if provided
	if expectedOuHandle != "" && jwtClaims.OuHandle != expectedOuHandle {
		return nil, fmt.Errorf("expected ouHandle to be '%s', got '%s'", expectedOuHandle, jwtClaims.OuHandle)
	}

	return jwtClaims, nil
}

// AssuranceExpectation defines expected assurance values for validation.
type AssuranceExpectation struct {
	AAL                    string
	IAL                    string
	ExpectedAuthenticators []string // Order matters
}

// ValidateAssurance validates the assurance block in JWT claims.
func ValidateAssurance(claims *JWTClaims, expected AssuranceExpectation) error {
	if claims.Assurance == nil {
		return fmt.Errorf("assurance block is missing from JWT claims")
	}

	// Validate AAL
	if aal, ok := claims.Assurance["aal"].(string); ok {
		if aal != expected.AAL {
			return fmt.Errorf("expected AAL '%s', got '%s'", expected.AAL, aal)
		}
	} else {
		return fmt.Errorf("AAL is missing or not a string")
	}

	// Validate IAL
	if ial, ok := claims.Assurance["ial"].(string); ok {
		if ial != expected.IAL {
			return fmt.Errorf("expected IAL '%s', got '%s'", expected.IAL, ial)
		}
	} else {
		return fmt.Errorf("IAL is missing or not a string")
	}

	// Validate authenticators
	authenticators, ok := claims.Assurance["authenticators"].([]interface{})
	if !ok {
		return fmt.Errorf("authenticators field is missing or not an array")
	}

	expectedCount := len(expected.ExpectedAuthenticators)
	if len(authenticators) != expectedCount {
		return fmt.Errorf("expected %d authenticators, got %d", expectedCount, len(authenticators))
	}

	// Validate each authenticator
	for i, auth := range authenticators {
		authMap, ok := auth.(map[string]interface{})
		if !ok {
			return fmt.Errorf("authenticator at index %d is not a valid object", i)
		}

		// Validate authenticator name
		authenticatorName, ok := authMap["authenticator"].(string)
		if !ok {
			return fmt.Errorf("authenticator name at index %d is missing or not a string", i)
		}
		if authenticatorName != expected.ExpectedAuthenticators[i] {
			return fmt.Errorf("expected authenticator at step %d to be '%s', got '%s'",
				i+1, expected.ExpectedAuthenticators[i], authenticatorName)
		}

		// Validate step number
		step, ok := authMap["step"].(float64)
		if !ok {
			return fmt.Errorf("step at index %d is missing or not a number", i)
		}
		if int(step) != i+1 {
			return fmt.Errorf("expected step %d, got %d", i+1, int(step))
		}

		// Validate timestamp exists and is positive
		timestamp, ok := authMap["timestamp"].(float64)
		if !ok {
			return fmt.Errorf("timestamp at index %d is missing or not a number", i)
		}
		if timestamp <= 0 {
			return fmt.Errorf("timestamp at index %d should be positive, got %f", i, timestamp)
		}
	}

	return nil
}
