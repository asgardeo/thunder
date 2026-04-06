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

import "testing"

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		in  string
		out bool
	}{
		{"test@example.com", true},
		{"user.name@domain.org", true},
		{" invalid@example.com ", true},
		{"test@example.com\r\n", false},
		{"not-an-email", false},
		{"@example.com", false},
		{"", false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.in, func(t *testing.T) {
			if IsValidEmail(tc.in) != tc.out {
				t.Fatalf("IsValidEmail(%q)= %v, want %v", tc.in, IsValidEmail(tc.in), tc.out)
			}
		})
	}
}
