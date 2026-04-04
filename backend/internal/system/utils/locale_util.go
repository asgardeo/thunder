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

package utils

import (
	"regexp"
	"strings"
)

// bcp47TagRegex matches valid BCP 47 language tags per RFC 5646.
// Primary language subtag: 1–8 letters only.
// Extension subtags: 1–8 alphanumeric characters that contain at least one letter
// (pure-digit subtags are not allowed except the 3-digit UN M.49 region codes).
var bcp47TagRegex = regexp.MustCompile(
	`^[a-zA-Z]{1,8}(-([a-zA-Z][a-zA-Z0-9]{0,7}|[0-9]{3}|[0-9][a-zA-Z][a-zA-Z0-9]{0,6}))*$`)

// MaxLocaleTagLength is the maximum allowed length for a BCP 47 language tag.
const MaxLocaleTagLength = 35

// IsValidBCP47Tag returns true if the tag is a syntactically valid BCP 47 language tag.
// Rejects empty strings, tags exceeding MaxLocaleTagLength, tags containing whitespace,
// and tags not matching the BCP 47 subtag structure.
func IsValidBCP47Tag(tag string) bool {
	if tag == "" || len(tag) > MaxLocaleTagLength {
		return false
	}
	return bcp47TagRegex.MatchString(tag)
}

// NormaliseBCP47Tag returns the lowercase-normalised form of a BCP 47 tag.
func NormaliseBCP47Tag(tag string) string {
	return strings.ToLower(tag)
}

// ResolveLocalisedValue returns the best matching value for the requested locale.
// Resolution order for each requested locale token: exact match → language-prefix match → base value.
// uiLocale may be a space-separated list per the OIDC spec; first-match-wins.
// If uiLocale is empty or no variant matches, base is returned.
func ResolveLocalisedValue(variants map[string]string, base, uiLocale string) string {
	if len(variants) == 0 || uiLocale == "" {
		return base
	}

	for _, locale := range strings.Fields(uiLocale) {
		normalised := NormaliseBCP47Tag(locale)

		// Exact match.
		if val, ok := variants[normalised]; ok {
			return val
		}

		// Language-prefix fallback: strip the last subtag and retry.
		// e.g. "fr-CA" → try "fr".
		tag := normalised
		for {
			idx := strings.LastIndex(tag, "-")
			if idx < 0 {
				break
			}
			tag = tag[:idx]
			if val, ok := variants[tag]; ok {
				return val
			}
		}
	}

	return base
}
