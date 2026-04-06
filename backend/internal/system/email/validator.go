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

package email

import (
	"net/mail"
	"strings"
)

// IsValidEmail returns true for syntactically valid email addresses.
func IsValidEmail(emailAddr string) bool {
	// Reject CR/LF early (do not trim them away first) to prevent header injection.
	if strings.ContainsAny(emailAddr, "\r\n") {
		return false
	}

	emailAddr = strings.TrimSpace(emailAddr)
	if emailAddr == "" {
		return false
	}

	addr, err := mail.ParseAddress(emailAddr)
	if err != nil {
		return false
	}
	parsedEmail := addr.Address

	atIndex := strings.LastIndex(parsedEmail, "@")
	if atIndex == -1 || atIndex == 0 || atIndex == len(parsedEmail)-1 {
		return false
	}

	localPart := parsedEmail[:atIndex]
	domain := parsedEmail[atIndex+1:]

	if len(localPart) == 0 || strings.HasPrefix(localPart, ".") || strings.HasSuffix(localPart, ".") {
		return false
	}

	if !strings.Contains(domain, ".") ||
		strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") ||
		strings.HasPrefix(domain, "-") || strings.HasSuffix(domain, "-") {
		return false
	}

	return true
}
