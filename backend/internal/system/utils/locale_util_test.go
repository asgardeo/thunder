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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidBCP47Tag(t *testing.T) {
	valid := []string{
		"en",
		"fr",
		"zh",
		"en-US",
		"en-GB",
		"fr-CA",
		"zh-Hans",
		"zh-Hans-CN",
		"sr-Latn-RS",
		"en-US-x-twain",
		"sl-Latn-IT-nedis",
		"en-001", // primary + UN M.49 three-digit region
	}
	for _, tag := range valid {
		assert.True(t, IsValidBCP47Tag(tag), "expected valid: %q", tag)
	}

	invalid := []string{
		"",                      // empty
		strings.Repeat("a", 36), // exceeds MaxLocaleTagLength
		"en-",                   // trailing dash
		"-en",                   // leading dash
		"en--US",                // double dash
		"en-1234",               // 4-digit all-numeric subtag
		"en-12",                 // 2-digit all-numeric subtag
		"123",                   // all-numeric primary subtag
		"en US",                 // space instead of dash
		"en_US",                 // underscore instead of dash
	}
	for _, tag := range invalid {
		assert.False(t, IsValidBCP47Tag(tag), "expected invalid: %q", tag)
	}
}

func TestNormaliseBCP47Tag(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"en", "en"},
		{"EN", "en"},
		{"fr-CA", "fr-ca"},
		{"zh-Hans-CN", "zh-hans-cn"},
		{"en-US", "en-us"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, NormaliseBCP47Tag(tc.input))
	}
}

func TestResolveLocalisedValue(t *testing.T) {
	variants := map[string]string{
		"fr":    "Bonjour",
		"fr-ca": "Bonjour Canada",
		"de":    "Hallo",
	}

	t.Run("exact match", func(t *testing.T) {
		assert.Equal(t, "Bonjour", ResolveLocalisedValue(variants, "Hello", "fr"))
	})

	t.Run("exact match case-insensitive", func(t *testing.T) {
		assert.Equal(t, "Bonjour", ResolveLocalisedValue(variants, "Hello", "FR"))
	})

	t.Run("subtag exact match", func(t *testing.T) {
		assert.Equal(t, "Bonjour Canada", ResolveLocalisedValue(variants, "Hello", "fr-CA"))
	})

	t.Run("subtag fallback to parent", func(t *testing.T) {
		// fr-BE has no variant, falls back to fr
		assert.Equal(t, "Bonjour", ResolveLocalisedValue(variants, "Hello", "fr-BE"))
	})

	t.Run("deep subtag fallback", func(t *testing.T) {
		// fr-CA-x-custom → fr-ca (exact) wins
		assert.Equal(t, "Bonjour Canada", ResolveLocalisedValue(variants, "Hello", "fr-CA-x-custom"))
	})

	t.Run("no match falls back to base", func(t *testing.T) {
		assert.Equal(t, "Hello", ResolveLocalisedValue(variants, "Hello", "ja"))
	})

	t.Run("empty ui_locale returns base", func(t *testing.T) {
		assert.Equal(t, "Hello", ResolveLocalisedValue(variants, "Hello", ""))
	})

	t.Run("nil variants returns base", func(t *testing.T) {
		assert.Equal(t, "Hello", ResolveLocalisedValue(nil, "Hello", "fr"))
	})

	t.Run("empty variants returns base", func(t *testing.T) {
		assert.Equal(t, "Hello", ResolveLocalisedValue(map[string]string{}, "Hello", "fr"))
	})

	t.Run("space-separated locales first match wins", func(t *testing.T) {
		// de matches before fr in preference list
		assert.Equal(t, "Hallo", ResolveLocalisedValue(variants, "Hello", "de fr"))
	})

	t.Run("space-separated locales falls through to second", func(t *testing.T) {
		// ja has no match, fr does
		assert.Equal(t, "Bonjour", ResolveLocalisedValue(variants, "Hello", "ja fr"))
	})

	t.Run("empty base value preserved", func(t *testing.T) {
		assert.Equal(t, "", ResolveLocalisedValue(nil, "", "fr"))
	})
}
