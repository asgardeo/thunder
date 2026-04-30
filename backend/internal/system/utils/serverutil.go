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

// Package utils provides utility functions for server wide operations.
package utils

import "strings"

// GetAllowedOrigin checks if the redirect URI is allowed and returns the allowed origin.
func GetAllowedOrigin(allowedOrigins []string, redirectURI string) string {
	if len(allowedOrigins) == 0 {
		return ""
	}

	for _, allowedOrigin := range allowedOrigins {
		if strings.Contains(redirectURI, allowedOrigin) {
			return allowedOrigin
		}
	}

	return ""
}
