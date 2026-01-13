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

package mgt

import (
	"regexp"

	goi18n "golang.org/x/text/language"
)

// SystemLanguage is the default language code for the system.
const SystemLanguage = "en"

// SystemNamespace is the default namespace for system translations.
const SystemNamespace = "system"

// LanguagePreferenceOrder defines the priority of languages for fallback.
var LanguagePreferenceOrder = map[string]int{
	"en-US": 0,
	"en":    1,
}

// namespaceRegex defines the valid format for namespace strings.
// Namespaces can contain alphanumeric characters, underscores, and hyphens.
var namespaceRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// keyRegex defines the valid format for translation keys.
// Keys can contain alphanumeric characters, dots, underscores, and hyphens.
var keyRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// ValidateLanguage validates that a language tag is in the canonical form according to BCP 47 format.
func ValidateLanguage(language string) bool {
	tag, err := goi18n.BCP47.Parse(language)
	if err != nil {
		return false
	}
	return tag.String() == language
}

// ValidateNamespace validates that a namespace string matches the required format.
// Returns true if the namespace is non-empty and contains only alphanumeric characters, underscores, and hyphens.
func ValidateNamespace(namespace string) bool {
	if namespace == "" {
		return false
	}
	if len(namespace) > 64 {
		return false
	}
	return namespaceRegex.MatchString(namespace)
}

// ValidateKey validates that a key string matches the required format.
// Returns true if the key is non-empty and contains only alphanumeric characters, dots, underscores, and hyphens.
func ValidateKey(key string) bool {
	if key == "" {
		return false
	}
	if len(key) > 256 {
		return false
	}
	return keyRegex.MatchString(key)
}
