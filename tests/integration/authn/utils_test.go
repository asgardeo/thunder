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

package authn

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
)

// extractAssuranceLevelFromAssertion extracts AAL or IAL from the JWT assertion token
// This is a helper function used across authentication tests
func extractAssuranceLevelFromAssertion(assertion string, levelType string) string {
	// Split JWT token into its three parts
	parts := bytes.Split([]byte(assertion), []byte("."))
	if len(parts) < 2 {
		return ""
	}

	// Decode payload (second part) using standard base64 URL encoding
	decoded, err := base64.RawURLEncoding.DecodeString(string(parts[1]))
	if err != nil {
		return ""
	}

	// Unmarshal JWT claims
	var claims map[string]interface{}
	err = json.Unmarshal(decoded, &claims)
	if err != nil {
		return ""
	}

	// Look for assurance object
	if assurance, exists := claims["assurance"]; exists {
		if assuranceMap, ok := assurance.(map[string]interface{}); ok {
			if level, exists := assuranceMap[levelType]; exists {
				if levelStr, ok := level.(string); ok {
					return levelStr
				}
			}
		}
	}

	return ""
}
