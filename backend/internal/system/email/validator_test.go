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
	"testing"

	"github.com/stretchr/testify/suite"
)

type ValidatorTestSuite struct {
	suite.Suite
}

func TestValidatorTestSuite(t *testing.T) {
	suite.Run(t, new(ValidatorTestSuite))
}

func (s *ValidatorTestSuite) TestIsValidEmail() {
	tests := []struct {
		in  string
		out bool
	}{
		{"test@example.com", true},
		{"user.name@domain.org", true},
		{" invalid@example.com ", true},
		{"Test User <test@example.com>", false},
		{"<test@example.com>", false},
		{"test@example.com\r\n", false},
		{"not-an-email", false},
		{"@example.com", false},
		{"", false},
	}

	for _, tc := range tests {
		s.Run(tc.in, func() {
			s.Equal(tc.out, IsValidEmail(tc.in), "IsValidEmail(%q)", tc.in)
		})
	}
}
