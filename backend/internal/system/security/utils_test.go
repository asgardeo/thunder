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

package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCompilePathPattern verifies that individual glob-style patterns are compiled
// to the correct anchored regular expression, and that invalid patterns are rejected.
func TestCompilePathPattern(t *testing.T) {
	tests := []struct {
		name           string
		pattern        string
		expectedRegex  string
		shouldMatch    []string
		shouldNotMatch []string
	}{
		{
			name:           "Exact path",
			pattern:        "/users/me",
			expectedRegex:  "^/users/me$",
			shouldMatch:    []string{"/users/me"},
			shouldNotMatch: []string{"/users/menu", "/users/me/profile", "/users"},
		},
		{
			name:           "Single wildcard segment",
			pattern:        "/api/*/users",
			expectedRegex:  "^/api/[^/]+/users$",
			shouldMatch:    []string{"/api/v1/users", "/api/test/users"},
			shouldNotMatch: []string{"/api/users", "/api/v1/v2/users"},
		},
		{
			name:           "Recursive wildcard suffix",
			pattern:        "/health/**",
			expectedRegex:  "^/health(?:/.*)?$",
			shouldMatch:    []string{"/health", "/health/", "/health/liveness", "/health/readiness/full"},
			shouldNotMatch: []string{"/healthz", "/other"},
		},
		{
			name:           "Multiple single wildcards",
			pattern:        "/i18n/languages/*/translations/ns/*/keys/*/resolve",
			expectedRegex:  "^/i18n/languages/[^/]+/translations/ns/[^/]+/keys/[^/]+/resolve$",
			shouldMatch:    []string{"/i18n/languages/en/translations/ns/common/keys/btn.submit/resolve"},
			shouldNotMatch: []string{"/i18n/languages/en/translations/ns/common/keys/btn.submit/extra"},
		},
		{
			name:           "Special characters escaped",
			pattern:        "/api/v1.0/user",
			expectedRegex:  "^/api/v1\\.0/user$",
			shouldMatch:    []string{"/api/v1.0/user"},
			shouldNotMatch: []string{"/api/v1a0/user"},
		},
		{
			name:           "Invalid: globstar in middle",
			pattern:        "/api/**/users",
			expectedRegex:  "",
			shouldMatch:    nil,
			shouldNotMatch: nil,
		},
		{
			name:           "Invalid: multiple globstars",
			pattern:        "/api/**/users/**",
			expectedRegex:  "",
			shouldMatch:    nil,
			shouldNotMatch: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re, err := compilePathPattern(tt.pattern)

			if tt.expectedRegex == "" {
				assert.Error(t, err)
				assert.Nil(t, re)
				assert.Contains(t, err.Error(), "invalid pattern")
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, re)
				assert.Equal(t, tt.expectedRegex, re.String())

				for _, matchPath := range tt.shouldMatch {
					assert.True(t, re.MatchString(matchPath), "Should match: %s", matchPath)
				}
				for _, mismatchPath := range tt.shouldNotMatch {
					assert.False(t, re.MatchString(mismatchPath), "Should not match: %s", mismatchPath)
				}
			}
		})
	}
}

// TestCompilePathPatterns verifies the batch wrapper: it returns the correct
// count of compiled patterns and stops at the first invalid entry.
func TestCompilePathPatterns(t *testing.T) {
	tests := []struct {
		name        string
		patterns    []string
		wantLen     int
		wantError   bool
		errContains string
	}{
		{
			name:     "Empty slice",
			patterns: []string{},
			wantLen:  0,
		},
		{
			name:     "All valid patterns",
			patterns: []string{"/health/**", "/api/*/resource", "/exact"},
			wantLen:  3,
		},
		{
			name:        "First pattern invalid",
			patterns:    []string{"/invalid/**/middle/**", "/valid/**"},
			wantError:   true,
			errContains: "invalid pattern",
		},
		{
			name:        "Last pattern invalid",
			patterns:    []string{"/valid/**", "/invalid/**/middle/**"},
			wantError:   true,
			errContains: "invalid pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiled, err := compilePathPatterns(tt.patterns)
			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, compiled)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
				assert.Len(t, compiled, tt.wantLen)
			}
		})
	}
}

// TestCompileAPIPermissions verifies that API permission entries are compiled
// to regex form correctly, and that invalid patterns are rejected.
func TestCompileAPIPermissions(t *testing.T) {
	tests := []struct {
		name        string
		entries     []apiPermissionEntry
		wantLen     int
		wantError   bool
		errContains string
	}{
		{
			name:    "Empty slice",
			entries: []apiPermissionEntry{},
			wantLen: 0,
		},
		{
			name: "Valid entries compiled",
			entries: []apiPermissionEntry{
				{"GET /users", PermissionUserView},
				{"GET /users/**", PermissionUserView},
				{"POST /users", PermissionUser},
			},
			wantLen: 3,
		},
		{
			name: "Single wildcard entry",
			entries: []apiPermissionEntry{
				{"GET /users/*/profile", PermissionUserView},
			},
			wantLen: 1,
		},
		{
			name: "Invalid pattern stops compilation",
			entries: []apiPermissionEntry{
				{"GET /valid/**", PermissionUserView},
				{"GET /invalid/**/middle/**", PermissionUser},
			},
			wantError:   true,
			errContains: "invalid pattern",
		},
		{
			name: "Invalid pattern as first entry",
			entries: []apiPermissionEntry{
				{"GET /invalid/**/middle/**", PermissionUser},
				{"GET /valid/**", PermissionUserView},
			},
			wantError:   true,
			errContains: "invalid pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiled, err := compileAPIPermissions(tt.entries)
			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, compiled)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
				assert.Len(t, compiled, tt.wantLen)
			}
		})
	}
}
